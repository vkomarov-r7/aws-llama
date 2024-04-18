package config

import (
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"path/filepath"
)

type Account struct {
	MetadataURL string `json:"metadata_url"`
	Nickname    string `json:"nickname"`
}

type Config struct {
	Accounts           []Account `json:"accounts"`
	RenewWithinSeconds float64
	RootUrl            *url.URL
	ChromeUserDataDir  string
	ListenPort         int
	Username           string `json:"username"`
	Password           string `json:"password"`
	StorageStatePath   string
}

func (c *Config) HasLogin() bool {
	return c.Username != "" && c.Password != ""
}

var CurrentConfig *Config

func InitConfig() {
	config, err := loadConfigFromJSON()
	if err != nil {
		panic(err)
	}

	CurrentConfig = config
}

func loadConfigFromJSON() (*Config, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}
	configFp := filepath.Join(homeDir, ".aws-llama.json")

	bytes, err := os.ReadFile(configFp)
	if err != nil {
		// It's ok if it doesn't exist.
		if !os.IsNotExist(err) {
			return nil, err
		}
	}

	rootUrl, err := url.Parse("http://localhost:2600")
	if err != nil {
		return nil, err
	}

	userDataDir, err := getChromeUserDataDir()
	if err != nil {
		return nil, err
	}
	storageStatePath, err := getStorageStatePath()
	if err != nil {
		return nil, err
	}

	config := Config{
		RootUrl:            rootUrl,
		RenewWithinSeconds: 15 * 60, // 15 mins.
		ChromeUserDataDir:  userDataDir,
		ListenPort:         2600,
		StorageStatePath:   storageStatePath,
	}
	if bytes != nil {
		err = json.Unmarshal(bytes, &config)
		if err != nil {
			return nil, err
		}
	}

	return &config, nil
}

func getChromeUserDataDir() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	userdir := filepath.Join(homeDir, "Library/Caches/awsllama/chrome")
	err = os.MkdirAll(userdir, 0700)
	if err != nil {
		return "", err
	}
	return userdir, nil
}

func getStorageStatePath() (string, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	storagePath := filepath.Join(homeDir, ".aws-llama-storage")
	if _, err := os.Stat(storagePath); errors.Is(err, os.ErrNotExist) {
		err = os.WriteFile(storagePath, []byte{'{', '}'}, 0600)
		if err != nil {
			return "", err
		}
	}
	return storagePath, nil
}
