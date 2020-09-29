// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package route

import (
	"fmt"
	"net/http"

	moovhttp "github.com/moov-io/base/http"
)

const organizationHeaderKey = "X-Organization"

// GetOrganization returns the value from the X-Organization header and writes an error to w if it's missing
func GetOrganization(w http.ResponseWriter, r *http.Request) string {
	if ns := r.Header.Get(organizationHeaderKey); ns == "" {
		moovhttp.Problem(w, fmt.Errorf("missing %s header", organizationHeaderKey))
		return ""
	} else {
		return ns
	}
}
