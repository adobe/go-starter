package keychainx

import (
	"errors"
	"github.com/keybase/go-keychain"
)

var ErrNotFound = errors.New("item not found")

func Ask(label, help string) (string, string, error) {
	return "", "", nil
}

func Save(label, user, password string) error {
	item := keychain.NewItem()
	item.SetSecClass(keychain.SecClassInternetPassword)
	item.SetLabel(label)
	item.SetAccount(user)
	item.SetData([]byte(password))

	return keychain.AddItem(item)
}

// Load credentials with a given label
func Load(label string) (string, string, error) {
	query := keychain.NewItem()
	query.SetSecClass(keychain.SecClassInternetPassword)
	query.SetLabel(label)
	query.SetMatchLimit(keychain.MatchLimitOne)
	query.SetReturnAttributes(true)
	query.SetReturnData(true)

	results, err := keychain.QueryItem(query)
	if err != nil {
		return "", "", err
	}

	for _, r := range results {
		return string(r.Account), string(r.Data), nil
	}

	return "", "", ErrNotFound
}
