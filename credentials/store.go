package credentials

import (
	"aws-llama/config"
	"time"
)

var CredentialStore AWSCredentialStore = AWSCredentialStore{Entries: make([]AWSCredentialEntry, 0)}

type AWSCredentialStore struct {
	Entries []AWSCredentialEntry
}

func (a *AWSCredentialStore) UpsertEntry(entry AWSCredentialEntry) {
	a.RemoveEntryForAccountId(entry.AccountId)
	a.Entries = append(a.Entries, entry)
}

func (a *AWSCredentialStore) RemoveEntryForAccountId(accountId string) {
	idx := a.indexForAccount(accountId)
	if idx != -1 {
		a.removeByIndex(idx)
	}
}

func (a *AWSCredentialStore) ExpiringEntries(withinSeconds float64) []AWSCredentialEntry {
	currentTime := time.Now().UTC()
	expiringEntries := make([]AWSCredentialEntry, 0)

	for _, entry := range a.Entries {
		delta := entry.Expiration.Sub(currentTime)
		if delta.Seconds() < withinSeconds {
			expiringEntries = append(expiringEntries, entry)
		}
	}

	return expiringEntries
}

func (a *AWSCredentialStore) NextExpiringEntry(withinSeconds float64) *AWSCredentialEntry {
	expiringEntries := a.ExpiringEntries(withinSeconds)
	if len(expiringEntries) > 0 {
		return &expiringEntries[0]
	}

	return nil
}

func (a *AWSCredentialStore) ContainsMetadataURL(metadataURL string) bool {
	for _, entry := range a.Entries {
		if entry.MetadataURL == metadataURL {
			return true
		}
	}
	return false
}

func NextMetadataURLForRefresh() string {
	// First return any configured accounts for which we don't have credentials yet.
	for _, account := range config.CurrentConfig.Accounts {
		if !CredentialStore.ContainsMetadataURL(account.MetadataURL) {
			return account.MetadataURL
		}
	}

	// Then return the next expiring one (if there is one)
	entry := CredentialStore.NextExpiringEntry(config.CurrentConfig.RenewWithinSeconds)
	if entry != nil {
		return entry.MetadataURL
	}

	// We've got nothing to refresh right now.
	return ""
}

func (a *AWSCredentialStore) indexForAccount(accountId string) int {
	for idx, entry := range a.Entries {
		if entry.AccountId == accountId {
			return idx
		}
	}
	return -1
}

func (a *AWSCredentialStore) removeByIndex(idx int) {
	a.Entries = append(a.Entries[:idx], a.Entries[idx+1:]...)
}
