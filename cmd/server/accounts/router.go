// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package accounts

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/moov-io/ach"
	moovhttp "github.com/moov-io/base/http"
	"github.com/moov-io/customers/client"
	"github.com/moov-io/customers/cmd/server/route"
	"github.com/moov-io/customers/internal/secrets"
	"github.com/moov-io/customers/internal/secrets/hash"
	"github.com/moov-io/customers/internal/secrets/mask"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

func RegisterRoutes(logger log.Logger, r *mux.Router, repo Repository, keeper *secrets.StringKeeper) {
	r.Methods("GET").Path("/customers/{customerID}/accounts").HandlerFunc(getCustomerAccounts(logger, repo))
	r.Methods("POST").Path("/customers/{customerID}/accounts").HandlerFunc(createCustomerAccount(logger, repo, keeper))
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
	AccountNumber string             `json:"accountNumber"`
	RoutingNumber string             `json:"routingNumber"`
	Type          client.AccountType `json:"type"`
	HolderType    client.HolderType  `json:"holderType"`

	// fields we compute from the inbound AccountNumber
	encryptedAccountNumber string
	hashedAccountNumber    string
	maskedAccountNumber    string
}

func (req *createAccountRequest) validate() error {
	if req.AccountNumber == "" {
		return errors.New("missing AccountNumber")
	}
	if err := ach.CheckRoutingNumber(req.RoutingNumber); err != nil {
		return err
	}

	at := func(t1, t2 client.AccountType) bool {
		return strings.EqualFold(string(t1), string(t2))
	}
	if !at(req.Type, client.CHECKING) && !at(req.Type, client.SAVINGS) {
		return fmt.Errorf("invalid account type: %s", req.Type)
	}

	ht := func(t1, t2 client.HolderType) bool {
		return strings.EqualFold(string(t1), string(t2))
	}
	if !ht(req.HolderType, client.INDIVIDUAL) && !ht(req.HolderType, client.BUSINESS) {
		return fmt.Errorf("invalid holder type: %s", req.HolderType)
	}
	return nil
}

func (req *createAccountRequest) disfigure(keeper *secrets.StringKeeper) error {
	if enc, err := keeper.EncryptString(req.AccountNumber); err != nil {
		return fmt.Errorf("problem encrypting account number: %v", err)
	} else {
		req.encryptedAccountNumber = enc
	}
	if v, err := hash.AccountNumber(req.AccountNumber); err != nil {
		return fmt.Errorf("problem hashing account number: %v", err)
	} else {
		req.hashedAccountNumber = v
	}
	req.maskedAccountNumber = mask.AccountNumber(req.AccountNumber)
	return nil
}

func createCustomerAccount(logger log.Logger, repo Repository, keeper *secrets.StringKeeper) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)

		var request createAccountRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			moovhttp.Problem(w, err)
			return
		}
		if err := request.validate(); err != nil {
			logger.Log("accounts", fmt.Sprintf("problem validating account: %v", err), "requestID", moovhttp.GetRequestID(r))
			moovhttp.Problem(w, err)
			return
		}
		if err := request.disfigure(keeper); err != nil {
			logger.Log("accounts", fmt.Sprintf("problem disfiguring account: %v", err), "requestID", moovhttp.GetRequestID(r))
			moovhttp.Problem(w, err)
			return
		}

		customerID, userID := route.GetCustomerID(w, r), moovhttp.GetUserID(r)
		account, err := repo.createCustomerAccount(customerID, userID, &request)
		if err != nil {
			logger.Log("accounts", fmt.Sprintf("problem saving account: %v", err), "requestID", moovhttp.GetRequestID(r))
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
