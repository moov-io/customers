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
	namespaceHeaderKey = util.Or(os.Getenv("NAMESPACE_HEADER"), "X-Namespace")
)

func GetNamespace(w http.ResponseWriter, r *http.Request) string {
	if ns := r.Header.Get(namespaceHeaderKey); ns == "" {
		moovhttp.Problem(w, fmt.Errorf("missing %s header", namespaceHeaderKey))
		return ""
	} else {
		return ns
	}
}
