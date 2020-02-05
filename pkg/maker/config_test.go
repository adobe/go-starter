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

import (
	"gopkg.in/yaml.v2"
	"reflect"
	"testing"
)

func TestStringOrSlice_UnmarshalYAML(t *testing.T) {
	tests := []struct {
		input string
		slice StringOrSlice
	}{
		{input: `"cmd"`, slice: StringOrSlice{"cmd"}},
		{input: `"cmd arg1 arg2"`, slice: StringOrSlice{"cmd", "arg1", "arg2"}},
		{input: `["cmd", "arg1", "arg2"]`, slice: StringOrSlice{"cmd", "arg1", "arg2"}},
		{input: `["cmd", "arg 1", "arg 2"]`, slice: StringOrSlice{"cmd", "arg 1", "arg 2"}},
		{input: `~`, slice: nil},
		{input: `""`, slice: StringOrSlice{""}},
		{input: `[]`, slice: StringOrSlice{}},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			got := StringOrSlice{}

			if err := yaml.Unmarshal([]byte(test.input), &got); err != nil {
				t.Fatalf("YAML %#v is incorrect: %v", test.input, err)
			}

			if want := test.slice; !reflect.DeepEqual(got, want) {
				t.Fatalf("Parsed value does not match: got %#v, want %#v", got, want)
			}
		})
	}
}
