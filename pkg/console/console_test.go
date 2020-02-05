/*
Copyright 2019 Adobe
All Rights Reserved.

NOTICE: Adobe permits you to use, modify, and distribute this file in
accordance with the terms of the Adobe license agreement accompanying
it. If you have received this file from a source other than Adobe,
then your use, modification, or distribution of it requires the prior
written permission of Adobe.
*/

package console

import (
	"io/ioutil"
	"strings"
	"testing"
)

func TestConsole_ReadString(t *testing.T) {
	want := "hello world"

	c := New(strings.NewReader(want+"\ndiscarded\n"), ioutil.Discard)

	got := c.ReadString("")

	if want != got {
		t.Errorf("Input does not match: got %#v, want %#v", got, want)
	}
}
