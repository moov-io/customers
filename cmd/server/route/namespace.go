// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package route

import (
	"errors"
	"net/http"

	moovhttp "github.com/moov-io/base/http"
)

var (
	ErrNoNamespace = errors.New("no Namespace found")
)

func GetNamespace(w http.ResponseWriter, r *http.Request) string {
	if ns := r.Header.Get("X-Namespace"); ns == "" {
		moovhttp.Problem(w, ErrNoNamespace)
		return ""
	} else {
		return ns
	}
}
