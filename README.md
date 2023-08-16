# AWS Llama

AWS Llama is an SSO authenticator for AWS using SAML-based role assumption.

AWS Llama uses the existing Chrome installation to authenticate with an IDP like Okta to obtain AWS credentials for one or more accounts.

## Configuration

Create a configuration file and fill in the metadata urls you'd like to keep available for use:

```
cat > ~/.aws-llama.json << EOF
{
    "accounts": [
        {
            "metadata_url": "https://company.okta.com/app/some_id/sso/saml/metadata",
            "nickname": "Account #1"
        },
        {
            "metadata_url": "https://company.okta.com/app/some_id/sso/saml/metadata",
            "nickname": "Account #2"
        }
    ]
}
EOF
```

## Usage

Download AWS-Llama from [the releases page](https://github.com/vkomarov-r7/aws-llama/releases). Make sure it's on the PATH
somewhere (like `/usr/local/bin`). Start it in a terminal like so:

```
aws-llama serve &
```

## Developing

1. Install the latest version of golang
2. Download deps: `go mod download`
3. Start the app using: `go run . serve`
