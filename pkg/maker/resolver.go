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
	"net/url"
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

	return u.String()
}
