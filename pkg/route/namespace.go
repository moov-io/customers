// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package route

import (
	"fmt"
	"net/http"
	"os"

	moovhttp "github.com/moov-io/base/http"
	"github.com/moov-io/customers/internal/util"
)

var (
	organizationHeaderKey = util.Or(os.Getenv("ORGANIZATION_HEADER"), "X-Organization")
)

func GetNamespace(w http.ResponseWriter, r *http.Request) string {
	if ns := r.Header.Get(organizationHeaderKey); ns == "" {
		moovhttp.Problem(w, fmt.Errorf("missing %s header", organizationHeaderKey))
		return ""
	} else {
		return ns
	}
}
