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
