/*
Copyright 2019 Adobe
All Rights Reserved.

NOTICE: Adobe permits you to use, modify, and distribute this file in
accordance with the terms of the Adobe license agreement accompanying
it. If you have received this file from a source other than Adobe,
then your use, modification, or distribution of it requires the prior
written permission of Adobe.
*/

package maker

import "testing"

func TestVars_Set(t *testing.T) {
	tests := []struct {
		input string
		key   string
		value string
	}{
		{input: "key=value", key: "key", value: "value"},
		{input: "key=", key: "key", value: ""},
		{input: "key", key: "key", value: "1"},
		{input: "key=value=value", key: "key", value: "value=value"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			vars := Vars{}

			if err := vars.Set(test.input); err != nil {
				t.Fatalf("Vars.Set should not return an error, but it returned %v", err)
			}

			if len(vars) != 1 {
				t.Fatalf("Vars should have 1 element, but it has %v instead", len(vars))
			}

			val, ok := vars[test.key]
			if !ok {
				t.Errorf("Vars should contain key %#v, got %#v instaed", test.key, vars)
			}

			if val != test.value {
				t.Errorf("Value does not match, want %#v, got %#v", test.value, val)
			}
		})
	}
}
