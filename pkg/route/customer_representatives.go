// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package route

import (
	"errors"
	"net/http"

	moovhttp "github.com/moov-io/base/http"

	"github.com/gorilla/mux"
)

var (
	ErrNoRepresentativeID = errors.New("no Representative ID found")
)

func GetRepresentativeID(w http.ResponseWriter, r *http.Request) string {
	v, ok := mux.Vars(r)["representativeID"]
	if !ok || v == "" {
		moovhttp.Problem(w, ErrNoRepresentativeID)
		return ""
	}
	return v
}
