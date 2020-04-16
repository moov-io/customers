// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package accounts

import (
	"encoding/json"
	"errors"
	"net/http"

	moovhttp "github.com/moov-io/base/http"
	"github.com/moov-io/customers/cmd/server/route"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

func RegisterRoutes(logger log.Logger, r *mux.Router, repo Repository) {
	r.Methods("GET").Path("/customers/{customerID}/accounts").HandlerFunc(getCustomerAccounts(logger, repo))
	r.Methods("POST").Path("/customers/{customerID}/accounts").HandlerFunc(createCustomerAccount(logger, repo))
	r.Methods("DELETE").Path("/customers/{customerID}/accounts/{accountID}").HandlerFunc(removeCustomerAccount(logger, repo))
}

func getCustomerAccounts(logger log.Logger, repo Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)

		customerID := route.GetCustomerID(w, r)
		accounts, err := repo.getCustomerAccounts(customerID)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(accounts)
	}
}

type createAccountRequest struct {
	AccountNumber string `json:"AccountNumber,omitempty"`
	RoutingNumber string `json:"routingNumber,omitempty"`
	Type          string `json:"type,omitempty"`
	HolderType    string `json:"holderType,omitempty"`
}

func createCustomerAccount(logger log.Logger, repo Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)

		var request createAccountRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			moovhttp.Problem(w, err)
			return
		}

		customerID, userID := route.GetCustomerID(w, r), moovhttp.GetUserID(r)
		account, err := repo.createCustomerAccount(customerID, userID, &request)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(account)
	}
}

func getAccountID(w http.ResponseWriter, r *http.Request) string {
	v, ok := mux.Vars(r)["accountID"]
	if !ok || v == "" {
		moovhttp.Problem(w, errors.New("no accountID"))
		return ""
	}
	return v
}

func removeCustomerAccount(logger log.Logger, repo Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)

		accountID := getAccountID(w, r)
		if err := repo.deactivateCustomerAccount(accountID); err != nil {
			moovhttp.Problem(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
	}
}
