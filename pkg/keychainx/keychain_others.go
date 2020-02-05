// +build !darwin

/*
Copyright 2019 Adobe
All Rights Reserved.

NOTICE: Adobe permits you to use, modify, and distribute this file in
accordance with the terms of the Adobe license agreement accompanying
it. If you have received this file from a source other than Adobe,
then your use, modification, or distribution of it requires the prior
written permission of Adobe.
*/

package keychainx

func Save(label, user, password string) error {
	return nil
}

// Load credentials with a given label
func Load(label string) (string, string, error) {
	return "", "", ErrNotFound
}
