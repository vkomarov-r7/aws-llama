package config

const ROOT_URL_RAW = "http://localhost:2600"

func GetMetadataUrls() []string {
	return []string{
		"https://rapid7.okta.com/app/exk1j8yygsyMLF6kI0h8/sso/saml/metadata", // Divvy QA
		"https://rapid7.okta.com/app/exk1j7498mkWenZXO0h8/sso/saml/metadata", // Divvy Hosted Shared
		"https://rapid7.okta.com/app/exk1j8yfmjvUWyyVK0h8/sso/saml/metadata", // Divvy Sales Demo
		"https://rapid7.okta.com/app/exk1iwckan4QzfBbt0h8/sso/saml/metadata", // Divvy Dev
		"https://rapid7.okta.com/app/exk1jvlnht9A0gytE0h8/sso/saml/metadata", // Divvy Platform Prod
	}
}
