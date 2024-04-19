package api

import (
	"aws-llama/config"
	"aws-llama/credentials"
	"aws-llama/log"
	"aws-llama/saml"
	"encoding/base64"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"

	"github.com/gin-gonic/gin"
)

type SAMLResponseBody struct {
	SAMLResponse string `json:"SAMLResponse"`
	RelayState   string `json:"RelayState"`
}

type CredentialSummary struct {
	AccountId  string
	Expiration *time.Time
}

func routeIndex(c *gin.Context) {
	var summaries []CredentialSummary
	for _, entry := range credentials.CredentialStore.Entries {
		summary := CredentialSummary{
			AccountId:  entry.AccountId,
			Expiration: &entry.Expiration,
		}
		summaries = append(summaries, summary)
	}
	c.JSON(200, gin.H{"credentials": summaries})
}

func routeLogin(c *gin.Context) {
	metadataURLRaw := c.Query("metadata_url")
	if metadataURLRaw == "" {
		metadataURLRaw = credentials.NextMetadataURLForRefresh()
	}

	// Nothing to do if nothing needs refreshing.
	if metadataURLRaw == "" {
		c.Redirect(http.StatusFound, "/")
		return
	}

	middleware, err := saml.MiddlewareForURL(metadataURLRaw)
	if err != nil {
		c.JSON(400, gin.H{"error": fmt.Sprintf("Failed to retrieve middleware for url: %s. %s", metadataURLRaw, err.Error())})
		return
	}
	redirectURL, err := saml.MakeRedirectUrl(middleware, metadataURLRaw)
	if err != nil {
		c.JSON(400, gin.H{"error": "Failed to build a SAML instance. " + err.Error()})
		return
	}

	c.Redirect(http.StatusFound, redirectURL.String())
}

func routeSAML(c *gin.Context) {
	samlResponse := SAMLResponseBody{}
	err := c.Bind(&samlResponse)
	if err != nil {
		c.JSON(400, gin.H{"error": "Failed to bind body. " + err.Error()})
		return
	}
	// log.Logger.Debug("Response Body after login", "samlResponse", samlResponse)

	rawResponseBuf, err := base64.StdEncoding.DecodeString(samlResponse.SAMLResponse)
	if err != nil {
		c.JSON(400, gin.H{"error": "Failed to decode SAMLResponse from POST body. " + err.Error()})
		return
	}

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

	for _, pair := range pairs {
		fmt.Printf("Processing pair from response: %+v\n", pair)
		credsResponse, err := saml.AssumeRoleWithSAML(pair.ProviderARN, pair.RoleARN, samlResponse.SAMLResponse)
		if err != nil {
			c.JSON(500, gin.H{"error": fmt.Sprintf("Failed to assume role: %s. %s", pair.RoleARN, err.Error())})
			return
		}
		log.Logger.Debug("Got credentials after saml response", credsResponse)

		credentialEntry, err := credentials.AWSCredentialEntryFromOutput(credsResponse, samlResponse.RelayState)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}
		credentials.CredentialStore.UpsertEntry(*credentialEntry)
	}

	credentials.StoreCredentials(credentials.CredentialStore.Entries)

	// Check to see if there's any other credentials that need to be fetched and do so.
	metadataURL := credentials.NextMetadataURLForRefresh()
	if metadataURL != "" {
		c.Redirect(302, "/login?metadata_url="+url.QueryEscape(metadataURL))
		return
	}
	c.Redirect(302, "/")
}

func CreateGinWebserver() *gin.Engine {
	r := gin.Default()
	r.SetTrustedProxies(nil)
	r.GET("/", routeIndex)
	r.GET("/login", routeLogin)
	r.POST("/sso/saml", routeSAML)
	return r
}

func RunWebserver(r *gin.Engine) {
	bind := fmt.Sprintf("127.0.0.1:%d", config.CurrentConfig.ListenPort)
	r.Run(bind)
}

func IsWebserverRunning() bool {
	timeout := time.Second
	address := fmt.Sprintf("127.0.0.1:%d", config.CurrentConfig.ListenPort)
	conn, err := net.DialTimeout("tcp", address, timeout)
	if err != nil {
		return false
	}
	if conn != nil {
		conn.Close()
		return true
	}

	panic("This should never happen (got both error and a connection).")
}
