package saml

import (
	"aws-llama/config"
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/sts"
	"github.com/crewjam/saml"
	"github.com/crewjam/saml/samlsp"
)

type RolePair struct {
	RoleARN     string
	ProviderARN string
}

var middlewareCache map[string]*samlsp.Middleware = make(map[string]*samlsp.Middleware)

func MakeRedirectUrl(m *samlsp.Middleware, relayState string) (*url.URL, error) {
	// Adapted from Middleware.HandleStartAuthFlow.
	var binding, bindingLocation string
	if m.Binding != "" {
		binding = m.Binding
		bindingLocation = m.ServiceProvider.GetSSOBindingLocation(binding)
	} else {
		binding = saml.HTTPRedirectBinding
		bindingLocation = m.ServiceProvider.GetSSOBindingLocation(binding)
		if bindingLocation == "" {
			binding = saml.HTTPPostBinding
			bindingLocation = m.ServiceProvider.GetSSOBindingLocation(binding)
		}
	}

	authReq, err := m.ServiceProvider.MakeAuthenticationRequest(bindingLocation, binding, m.ResponseBinding)
	if err != nil {
		return nil, err
	}

	if binding != saml.HTTPRedirectBinding {
		return nil, fmt.Errorf("unsupported binding type: %s", binding)
	}

	redirectURL, err := authReq.Redirect(relayState, &m.ServiceProvider)
	return redirectURL, err
}

func ExtractPairsFromAssertion(assertion *saml.Assertion) ([]*RolePair, error) {
	rolePairs := make([]*RolePair, 0)

	for _, statement := range assertion.AttributeStatements {
		for _, attribute := range statement.Attributes {
			if attribute.Name == "https://aws.amazon.com/SAML/Attributes/Role" {
				for _, value := range attribute.Values {
					pair, err := ExtractPairFromString(value.Value)
					if err != nil {
						return nil, err
					}
					rolePairs = append(rolePairs, pair)
				}
			}
		}
	}

	return rolePairs, nil
}

func ExtractPairFromString(s string) (*RolePair, error) {
	strs := strings.Split(s, ",")
	if len(strs) != 2 {
		return nil, fmt.Errorf("failed to parse pair from string: %s", s)
	}

	pair := RolePair{
		RoleARN:     strs[0],
		ProviderARN: strs[1],
	}
	return &pair, nil
}

func AssumeRoleWithSAML(principalArn string, roleArn string, samlAssertion string) (*sts.AssumeRoleWithSAMLOutput, error) {
	stsSession := session.Must(session.NewSession())
	client := sts.New(stsSession)
	assumeRoleInput := sts.AssumeRoleWithSAMLInput{
		PrincipalArn:  &principalArn,
		RoleArn:       &roleArn,
		SAMLAssertion: &samlAssertion,
	}
	output, err := client.AssumeRoleWithSAML(&assumeRoleInput)

	return output, err
}

func FetchSAMLMetadata(metadataURL string) (*saml.EntityDescriptor, error) {
	url, err := url.Parse(metadataURL)
	if err != nil {
		return nil, err
	}

	metadata, err := samlsp.FetchMetadata(context.Background(), http.DefaultClient, *url)
	return metadata, err
}

func MiddlewareForURL(metadataURL string) (*samlsp.Middleware, error) {
	middleware, ok := middlewareCache[metadataURL]
	if ok {
		return middleware, nil
	}

	metadata, err := FetchSAMLMetadata(metadataURL)
	if err != nil {
		return nil, err
	}

	rootURL, err := url.Parse(config.ROOT_URL_RAW)
	if err != nil {
		return nil, err
	}

	middleware, err = samlsp.New(samlsp.Options{
		URL:               *rootURL,
		IDPMetadata:       metadata,
		AllowIDPInitiated: true, // Prevent request tracking validation.
	})
	// Reconfigure AcsURL to pass validation checks.
	middleware.ServiceProvider.AcsURL = *middleware.ServiceProvider.AcsURL.ResolveReference(&url.URL{Path: "/sso/saml"})
	middleware.ServiceProvider.MetadataURL = *middleware.ServiceProvider.MetadataURL.ResolveReference(&url.URL{Path: "/sso/saml"})
	if err != nil {
		return nil, err
	}

	middlewareCache[metadataURL] = middleware
	return middleware, nil
}
