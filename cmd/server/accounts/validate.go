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
	"github.com/moov-io/customers/client"
	"github.com/moov-io/customers/cmd/server/accounts/validator"
	"github.com/moov-io/customers/cmd/server/paygate"
	"github.com/moov-io/customers/cmd/server/route"

	"github.com/go-kit/kit/log"
)

func initAccountValidation(logger log.Logger, repo Repository, strategies map[validator.StrategyKey]validator.Strategy) http.HandlerFunc {
	type request struct {
		Strategy string `json:"strategy"`
		Vendor   string `json:"vendor"`
	}

	type response struct {
		ValidationID   string                    `json:"validationID"`
		Status         string                    `json:"status"`
		Strategy       string                    `json:"strategy"`
		Vendor         string                    `json:"vendor"`
		VendorResponse *validator.VendorResponse `json:"vendor_response"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		// check required parameters
		w = route.Responder(logger, w, r)

		// TODO discuss
		// following methods Get...ID have side effect inside: moovhttp.Problem(w, ErrNoCustomerID)
		// customerID, accountID := route.GetCustomerID(w, r), getAccountID(w, r)
		vars := mux.Vars(r)
		customerID, accountID := vars["customerID"], vars["accountID"]

		if customerID == "" || accountID == "" {
			moovhttp.Problem(w, fmt.Errorf("missing customerID: %s and/or accountID: %s", customerID, accountID))
			return
		}

		// check if account is not validated yet
		// ...

		// decode request params
		req := &request{}
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

		// within transaction create validation
		// ...

		// execute strategy and get vendor response
		vendorResponse, err := strategy.InitAccountValidation("", "", "")
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}

		// render validation with vendor response
		res := &response{
			ValidationID:   "1234",
			Status:         "pending",
			Strategy:       "test",
			Vendor:         "moov",
			VendorResponse: vendorResponse,
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
	}
}

func validateAccount(logger log.Logger, repo Repository, paygateClient paygate.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)

		customerID, accountID := route.GetCustomerID(w, r), getAccountID(w, r)
		if customerID == "" || accountID == "" {
			return
		}

		// Lookup the account and verify it needs to be validated
		account, err := repo.getCustomerAccount(customerID, accountID)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}
		if !strings.EqualFold(string(account.Status), string(client.NONE)) {
			moovhttp.Problem(w, fmt.Errorf("unexpected accountID=%s status=%s", accountID, account.Status))
			return
		}

		var req client.UpdateValidation
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			moovhttp.Problem(w, fmt.Errorf("unable to read UpdateValidation: %v", err))
			return
		}

		switch req.Strategy {
		case "micro-deposits":
			userID := moovhttp.GetUserID(r)
			if err := handleMicroDepositValidation(repo, paygateClient, accountID, customerID, userID, req.MicroDeposits); err != nil {
				moovhttp.Problem(w, err)
				return
			}

		default:
			moovhttp.Problem(w, fmt.Errorf("unknown strategy %s", req.Strategy))
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
