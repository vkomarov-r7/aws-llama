package credentials

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

// func init() {
// 	credentialStore = AWSCredentialStore{Entries: make([]AWSCredentialEntry, 0)}
// }

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
