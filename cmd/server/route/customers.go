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
	ErrNoCustomerID = errors.New("no Customer ID found")
	ErrNoAccountID  = errors.New("no Account ID found")
)

func GetCustomerID(w http.ResponseWriter, r *http.Request) string {
	v, ok := mux.Vars(r)["customerID"]
	if !ok || v == "" {
		moovhttp.Problem(w, ErrNoCustomerID)
		return ""
	}
	return v
}

func GetAccountID(w http.ResponseWriter, r *http.Request) string {
	v, ok := mux.Vars(r)["accountID"]
	if !ok || v == "" {
		moovhttp.Problem(w, ErrNoAccountID)
		return ""
	}
	return v
}
