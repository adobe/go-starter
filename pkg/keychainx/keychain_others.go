// +build !darwin

package keychainx

func Save(label, user, password string) error {
	return nil
}

// Load credentials with a given label
func Load(label string) (string, string, error) {
	return "", "", ErrNotFound
}
