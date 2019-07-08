package maker

import (
	"net/url"
	"strings"
)

// ResolveTemplateURL URL
func ResolveTemplateURL(template string) string {
	u, err := url.Parse(template)
	if err != nil {
		return template
	}

	if u.Scheme == "" {
		u.Scheme = "https"
	}

	if u.Host == "" {
		u.Host = "github.com"
	}

	if !strings.ContainsRune(u.Path, '/') {
		u.Path = "magento-mcom/" + u.Path
	}

	return u.String()
}
