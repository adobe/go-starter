package maker

import "testing"

func TestResolveTemplateURL(t *testing.T) {
	tests := []struct {
		input string
		url   string
	}{
		{input: "go-scaffolding", url: "https://github.com/magento-mcom/go-scaffolding"},
		{input: "mom/go-scaffolding", url: "https://github.com/mom/go-scaffolding"},
		{input: "//github.adobe.com/mom/go-scaffolding", url: "https://github.adobe.com/mom/go-scaffolding"},
		{input: "git://github.adobe.com/mom/go-scaffolding", url: "git://github.adobe.com/mom/go-scaffolding"},
		{input: "git@github.adobe.com:mom/go-scaffolding", url: "git@github.adobe.com:mom/go-scaffolding"},
	}
	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			got := ResolveTemplateURL(test.input)

			if want := test.url; got != want {
				t.Errorf("Resolved URL does not match, got %#v, want %#v", got, want)
			}
		})
	}
}
