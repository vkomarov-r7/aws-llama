package api

import (
	"aws-llama/config"
	"aws-llama/saml"
	"encoding/base64"
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

type SAMLResponseBody struct {
	SAMLResponse string `json:"SAMLResponse"`
	RelayState   string `json:"RelayState"`
}

func routeIndex(c *gin.Context) {
	c.JSON(200, gin.H{"hello": "world"})
}

func routeLogin(c *gin.Context) {
	metadataURLRaw := c.Query("metadata_url")
	if metadataURLRaw == "" {
		metadataURLRaw = config.GetMetadataUrls()[0]
		// c.JSON(400, gin.H{"error": "Must specify the 'metadata_url' as a query string for this path."})
		// return
	}

	middleware, err := saml.MiddlewareForURL(metadataURLRaw)
	if err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Failed to retrieve middleware for url: %s. %s", metadataURLRaw, err.Error())})

	}
	redirectURL, err := saml.MakeRedirectUrl(middleware, metadataURLRaw)
	if err != nil {
		c.JSON(400, gin.H{"error": "Failed to build a SAML instance. " + err.Error()})
		return
	}

	fmt.Printf("Redirect URL: %s", redirectURL)

	c.Redirect(http.StatusFound, redirectURL.String())
}

func routeSAML(c *gin.Context) {
	samlResponse := SAMLResponseBody{}
	err := c.Bind(&samlResponse)
	if err != nil {
		c.JSON(400, gin.H{"error": "Failed to bind body. " + err.Error()})
		return
	}

	fmt.Printf("Response Body after login: %+v", samlResponse)

	rawResponseBuf, err := base64.StdEncoding.DecodeString(samlResponse.SAMLResponse)
	if err != nil {
		c.JSON(400, gin.H{"error": "Failed to decode SAMLResponse from POST body. " + err.Error()})
		return
	}
	fmt.Printf("Decoded response body after SAML Assertion: %s", rawResponseBuf)

	middleware, err := saml.MiddlewareForURL(samlResponse.RelayState)
	if err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Failed to resolve middlware for origin url: %s. %s", samlResponse.RelayState, err.Error())})
		return
	}

	assertion, err := middleware.ServiceProvider.ParseXMLResponse(rawResponseBuf, make([]string, 0))
	if err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Failed to parse SAML Response for URL: %s. %s", samlResponse.RelayState, err.Error())})
		return
	}

	pairs, err := saml.ExtractPairsFromAssertion(assertion)
	if err != nil {
		c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to extract pairs SAML: %s", err.Error())})
		return
	}

	fmt.Printf("Got pairs from saml callback: %+v", pairs)
	for _, pair := range pairs {
		credentials, err := saml.AssumeRoleWithSAML(pair.ProviderARN, pair.RoleARN, samlResponse.SAMLResponse)
		if err != nil {
			c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to assume role: %s. %s", pair.RoleARN, err.Error())})
			return
		}
		fmt.Printf("Got credentials after saml response: %+v", credentials)
	}
	c.JSON(200, gin.H{"msg": "Thanks!"})
}

func CreateGinWebserver() *gin.Engine {
	r := gin.Default()
	r.GET("/", routeIndex)
	r.GET("/login", routeLogin)
	r.POST("/sso/saml", routeSAML)
	return r
}
