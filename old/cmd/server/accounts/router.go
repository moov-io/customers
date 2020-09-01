// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package accounts

import (
	"encoding/json"
	"net/http"

	moovhttp "github.com/moov-io/base/http"
	"github.com/moov-io/customers/client"
	"github.com/moov-io/customers/cmd/server/fed"
	"github.com/moov-io/customers/cmd/server/paygate"
	"github.com/moov-io/customers/cmd/server/route"
	"github.com/moov-io/customers/pkg/secrets"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

func RegisterRoutes(logger log.Logger, r *mux.Router, repo Repository, fedClient fed.Client, paygateClient paygate.Client, keeper, transitKeeper *secrets.StringKeeper) {
	r.Methods("GET").Path("/customers/{customerID}/accounts").HandlerFunc(getCustomerAccounts(logger, repo, fedClient))
	r.Methods("POST").Path("/customers/{customerID}/accounts").HandlerFunc(createCustomerAccount(logger, repo, fedClient, keeper))
	r.Methods("POST").Path("/customers/{customerID}/accounts/{accountID}/decrypt").HandlerFunc(decryptAccountNumber(logger, repo, keeper, transitKeeper))
	r.Methods("DELETE").Path("/customers/{customerID}/accounts/{accountID}").HandlerFunc(removeCustomerAccount(logger, repo))
	r.Methods("PUT").Path("/customers/{customerID}/accounts/{accountID}/validate").HandlerFunc(validateAccount(logger, repo, paygateClient))
}

func getCustomerAccounts(logger log.Logger, repo Repository, fedClient fed.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(accounts)
	}
}

func createCustomerAccount(logger log.Logger, repo Repository, fedClient fed.Client, keeper *secrets.StringKeeper) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(account)
	}
}

func decryptAccountNumber(logger log.Logger, repo Repository, keeper *secrets.StringKeeper, transitKeeper *secrets.StringKeeper) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)

		customerID, accountID := route.GetCustomerID(w, r), getAccountID(w, r)
		if customerID == "" || accountID == "" {
			return
		}

		// grab encrypted value
		encrypted, err := repo.getEncryptedAccountNumber(customerID, accountID)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}
		// decrypt from database
		decrypted, err := keeper.DecryptString(encrypted)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}
		// encrypt for transit response
		encrypted, err = transitKeeper.EncryptString(decrypted)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(client.TransitAccountNumber{
			AccountNumber: encrypted,
		})
	}
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
