package credentials

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/service/sts"
	"gopkg.in/ini.v1"
)

const PROFILE_PREFIX = "llama"
const CREDENTIALS_HEADER = "; This file is managed by aws-llama"

type AWSCredential struct {
	AccessKeyId     string
	SecretAccessKey string
	SessionToken    string
	SecurityToken   string
}

type AWSCredentialEntry struct {
	AccountId   string
	Credential  AWSCredential
	MetadataURL string

	// Time when the current credentials expire.
	Expiration time.Time
}

func AWSCredentialEntryFromOutput(output *sts.AssumeRoleWithSAMLOutput) (*AWSCredentialEntry, error) {
	accountId, err := ExtractAccountIdFromARN(*output.AssumedRoleUser.Arn)
	if err != nil {
		return nil, err
	}
	credentialEntry := AWSCredentialEntry{
		AccountId: accountId,
		Credential: AWSCredential{
			AccessKeyId:     *output.Credentials.AccessKeyId,
			SecretAccessKey: *output.Credentials.SecretAccessKey,
			SessionToken:    *output.Credentials.SessionToken,
		},
	}
	return &credentialEntry, nil
}

func ExtractAccountIdFromARN(arn string) (string, error) {
	// Sample Input: arn:aws:sts::050283019178:assumed-role/developer/Val_Komarov@rapid7.com
	// Sample Output: 050283019178
	splitStr := strings.Split(arn, ":")
	if len(splitStr) < 6 {
		return "", fmt.Errorf("unable to extract account from malformed ARN: %s", arn)
	}

	account := splitStr[4]
	return account, nil
}

func getCredentialsPath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	return filepath.Join(homeDir, ".aws/credentials"), nil
}

func maybeBackupExistingCredentials() error {
	credentialsPath, err := getCredentialsPath()
	if err != nil {
		return err
	}

	contents, err := os.ReadFile(credentialsPath)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, no need to backup.
			return nil
		}
		return err
	}

	contentStr := strings.Trim(string(contents), " \n")
	if strings.HasPrefix(contentStr, CREDENTIALS_HEADER) {
		// File already is being managed. Nothing to do here.
		return nil
	}

	// We know we need to backup credentials here.
	backupCredentialsPath := credentialsPath + ".bak"
	err = os.Remove(backupCredentialsPath)
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}
	}

	return os.Rename(credentialsPath, backupCredentialsPath)
}

func writeIniToDisk(iniFile *ini.File) error {
	err := maybeBackupExistingCredentials()
	if err != nil {
		return err
	}

	credentialsPath, err := getCredentialsPath()
	if err != nil {
		return nil
	}
	err = os.MkdirAll(filepath.Base(credentialsPath), 0755)
	if err != nil {
		return err
	}

	f, err := os.Create(credentialsPath)
	if err != nil {
		return err
	}
	defer f.Close()

	_, err = f.WriteString(CREDENTIALS_HEADER + "\n")
	if err != nil {
		return err
	}

	_, err = iniFile.WriteTo(f)
	if err != nil {
		return err
	}
	return nil
}

func (a *AWSCredentialEntry) writeToIni(iniFile *ini.File) {
	sectionName := fmt.Sprintf("%s-%s", PROFILE_PREFIX, a.AccountId)
	section := iniFile.Section(sectionName)
	section.Key("aws_access_key_id").SetValue(a.Credential.AccessKeyId)
	section.Key("aws_secret_access_key").SetValue(a.Credential.SecretAccessKey)
	if a.Credential.SessionToken != "" {
		section.Key("aws_session_token").SetValue(a.Credential.SessionToken)
	}
	if a.Credential.SecurityToken != "" {
		section.Key("aws_security_token").SetValue(a.Credential.SecurityToken)
	}
}

func StoreCredentials(credentials []AWSCredentialEntry) error {
	iniFile := ini.Empty()
	for _, credential := range credentials {
		credential.writeToIni(iniFile)
	}
	return writeIniToDisk(iniFile)
}
