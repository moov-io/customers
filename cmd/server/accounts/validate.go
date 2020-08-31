// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package accounts

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/gorilla/mux"
	moovhttp "github.com/moov-io/base/http"
	"github.com/moov-io/customers/admin"
	"github.com/moov-io/customers/client"
	"github.com/moov-io/customers/cmd/server/accounts/validator"
	"github.com/moov-io/customers/cmd/server/route"
	"github.com/moov-io/customers/pkg/secrets"

	"github.com/go-kit/kit/log"
)

type initAccountValidationRequest struct {
	Strategy string `json:"strategy"`
	Vendor   string `json:"vendor"`
}

type initAccountValidationResponse struct {
	// TODO: do we want to create DB records for validation?
	// following fields may be stored in DB
	// ValidationID   string                    `json:"validationID"`
	// Status         string                    `json:"status"`
	// Strategy       string                    `json:"strategy"`
	// Vendor         string                    `json:"vendor"`
	VendorResponse *validator.VendorResponse `json:"vendor_response"`
}

func initAccountValidation(logger log.Logger, repo Repository, strategies map[validator.StrategyKey]validator.Strategy) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)

		// TODO: discuss
		// following methods Get...ID have side effect inside: moovhttp.Problem(w, ErrNoCustomerID)
		// customerID, accountID := route.GetCustomerID(w, r), getAccountID(w, r)
		vars := mux.Vars(r)
		userID, customerID, accountID := moovhttp.GetUserID(r), vars["customerID"], vars["accountID"]

		if customerID == "" || accountID == "" {
			moovhttp.Problem(w, fmt.Errorf("missing customerID: %s and/or accountID: %s", customerID, accountID))
			return
		}

		// Lookup the account and verify it needs to be validated
		account, err := repo.getCustomerAccount(customerID, accountID)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}
		if !strings.EqualFold(string(account.Status), string(client.NONE)) {
			moovhttp.Problem(w, fmt.Errorf("expected accountID=%s status to be '%s', but it is '%s'", accountID, client.NONE, account.Status))
			return
		}

		// decode request params
		req := &initAccountValidationRequest{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			moovhttp.Problem(w, fmt.Errorf("unable to read request: %v", err))
			return
		}

		// set default vendor if not specified
		if req.Vendor == "" {
			req.Vendor = "moov"
		}

		// find requested strategy
		strategyKey := validator.StrategyKey{
			Strategy: req.Strategy,
			Vendor:   req.Vendor,
		}

		strategy, found := strategies[strategyKey]
		if !found {
			moovhttp.Problem(w, fmt.Errorf("strategy %s for vendor %s was not found", req.Strategy, req.Vendor))
			return
		}

		// execute strategy and get vendor response
		vendorResponse, err := strategy.InitAccountValidation(userID, accountID, customerID)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}

		// render validation with vendor response
		res := &initAccountValidationResponse{
			VendorResponse: vendorResponse,
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
	}
}

type completeAccountValidationRequest struct {
	Strategy      string                   `json:"strategy"`
	Vendor        string                   `json:"vendor"`
	VendorRequest *validator.VendorRequest `json:"vendor_request"`
}

type completeAccountValidationResponse struct {
	VendorResponse *validator.VendorResponse `json:"vendor_response"`
}

func completeAccountValidation(logger log.Logger, repo Repository, keeper *secrets.StringKeeper, strategies map[validator.StrategyKey]validator.Strategy) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)

		vars := mux.Vars(r)
		userID, customerID, accountID := moovhttp.GetUserID(r), vars["customerID"], vars["accountID"]

		if customerID == "" || accountID == "" {
			moovhttp.Problem(w, fmt.Errorf("missing customerID: %s and/or accountID: %s", customerID, accountID))
			return
		}

		// // check if account is not validated yet
		account, err := repo.getCustomerAccount(customerID, accountID)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}
		if !strings.EqualFold(string(account.Status), string(client.NONE)) {
			moovhttp.Problem(w, fmt.Errorf("expected accountID=%s status to be '%s', but it is '%s'", accountID, client.NONE, account.Status))
			return
		}

		// decode request params
		req := &completeAccountValidationRequest{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			moovhttp.Problem(w, fmt.Errorf("unable to read request: %v", err))
			return
		}

		// set default vendor if not specified
		// if we have Validation record we can get strategy and vendor from it
		// for now let's keep it as is
		if req.Vendor == "" {
			req.Vendor = "moov"
		}

		// find requested strategy
		strategyKey := validator.StrategyKey{
			Strategy: req.Strategy,
			Vendor:   req.Vendor,
		}

		strategy, found := strategies[strategyKey]
		if !found {
			moovhttp.Problem(w, fmt.Errorf("strategy %s for vendor %s was not found", req.Strategy, req.Vendor))
			return
		}

		// grab encrypted account number
		encrypted, err := repo.getEncryptedAccountNumber(customerID, accountID)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}

		fmt.Println("Encrypted", encrypted)

		// decrypt from database
		accountNumber, err := keeper.DecryptString(encrypted)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}

		// execute strategy and get vendor response
		vendorResponse, err := strategy.CompleteAccountValidation(userID, customerID, account, accountNumber, req.VendorRequest)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}

		// TODO: why admin.VALIDATED is used here?
		// account.Status is client.AccountStatus (with client.VALIDATED)
		// should I change it into client.VALIDATED?
		err = repo.updateAccountStatus(accountID, admin.VALIDATED)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}

		// render validation with vendor response
		res := &completeAccountValidationResponse{
			VendorResponse: vendorResponse,
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
	}
}
