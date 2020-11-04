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

	"github.com/moov-io/customers/pkg/client"
	"github.com/moov-io/customers/pkg/route"
	"github.com/moov-io/customers/pkg/secrets"
	"github.com/moov-io/customers/pkg/validator"

	"github.com/moov-io/base/log"
)

func getAccountValidation(logger log.Logger, accounts Repository, validations validator.Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)

		vars := mux.Vars(r)

		customerID := vars["customerID"]
		accountID := vars["accountID"]
		validationID := vars["validationID"]

		account, err := accounts.getCustomerAccount(customerID, accountID)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}

		validation, err := validations.GetValidation(account.AccountID, validationID)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}

		res := client.AccountValidationResponse{
			ValidationID: validation.ValidationID,
			Strategy:     validation.Strategy,
			Vendor:       validation.Vendor,
			Status:       validation.Status,
			CreatedAt:    validation.CreatedAt,
			UpdatedAt:    validation.UpdatedAt,
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
	}
}

func initAccountValidation(logger log.Logger, accounts Repository, validations validator.Repository, strategies map[validator.StrategyKey]validator.Strategy) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)

		vars := mux.Vars(r)
		customerID := vars["customerID"]
		accountID := vars["accountID"]

		organization := route.GetOrganization(w, r)
		if organization == "" {
			return
		}

		account, err := accounts.getCustomerAccount(customerID, accountID)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}
		if !strings.EqualFold(string(account.Status), string(client.ACCOUNTSTATUS_NONE)) {
			moovhttp.Problem(w, fmt.Errorf("expected accountID=%s status to be '%s', but it is '%s'", accountID, client.ACCOUNTSTATUS_NONE, account.Status))
			return
		}

		// decode request params
		req := &client.InitAccountValidationRequest{}
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

		// TODO I would wrap it into transaction, so if strategy fails
		// then no validation record is created
		validation := &validator.Validation{
			AccountID: accountID,
			Strategy:  req.Strategy,
			Vendor:    req.Vendor,
			Status:    validator.StatusInit,
		}
		err = validations.CreateValidation(validation)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}

		// execute strategy and get vendor response
		vendorResponse, err := strategy.InitAccountValidation(organization, accountID, customerID)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}

		// render validation with vendor response
		res := client.InitAccountValidationResponse{
			ValidationID:   validation.ValidationID,
			Strategy:       validation.Strategy,
			Vendor:         validation.Vendor,
			Status:         validation.Status,
			CreatedAt:      validation.CreatedAt,
			UpdatedAt:      validation.UpdatedAt,
			VendorResponse: *vendorResponse,
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
	}
}

func completeAccountValidation(logger log.Logger, repo Repository, validations validator.Repository, keeper *secrets.StringKeeper, strategies map[validator.StrategyKey]validator.Strategy) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)

		vars := mux.Vars(r)
		customerID := vars["customerID"]
		accountID := vars["accountID"]
		validationID := vars["validationID"]

		organization := route.GetOrganization(w, r)
		if organization == "" {
			return
		}

		account, err := repo.getCustomerAccount(customerID, accountID)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}
		if !strings.EqualFold(string(account.Status), string(client.ACCOUNTSTATUS_NONE)) {
			moovhttp.Problem(w, fmt.Errorf("expected accountID=%s status to be '%s', but it is '%s'", accountID, client.ACCOUNTSTATUS_NONE, account.Status))
			return
		}

		// decode request params
		req := &client.CompleteAccountValidationRequest{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			moovhttp.Problem(w, fmt.Errorf("unable to read request: %v", err))
			return
		}

		validation, err := validations.GetValidation(account.AccountID, validationID)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}

		if validation.Status != validator.StatusInit {
			moovhttp.Problem(w, fmt.Errorf("expected validation: %s status to be '%s', but it is '%s'", validationID, validator.StatusInit, validation.Status))
			return
		}

		// find requested strategy
		strategyKey := validator.StrategyKey{
			Strategy: validation.Strategy,
			Vendor:   validation.Vendor,
		}

		strategy, found := strategies[strategyKey]
		if !found {
			moovhttp.Problem(w, fmt.Errorf("strategy %s for vendor %s was not found", strategyKey.Strategy, strategyKey.Vendor))
			return
		}

		// grab encrypted account number
		encrypted, err := repo.getEncryptedAccountNumber(organization, customerID, accountID)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}

		// decrypt from database
		accountNumber, err := keeper.DecryptString(encrypted)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}

		vendorRequest := validator.VendorRequest(req.VendorRequest)
		vendorResponse, err := strategy.CompleteAccountValidation(organization, customerID, account, accountNumber, &vendorRequest)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}

		validation.Status = validator.StatusCompleted
		err = validations.UpdateValidation(validation)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}

		err = repo.updateAccountStatus(accountID, client.ACCOUNTSTATUS_VALIDATED)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}

		// render validation with vendor response
		res := &client.CompleteAccountValidationResponse{
			ValidationID:   validation.ValidationID,
			Strategy:       validation.Strategy,
			Vendor:         validation.Vendor,
			Status:         validation.Status,
			CreatedAt:      validation.CreatedAt,
			UpdatedAt:      validation.UpdatedAt,
			VendorResponse: *vendorResponse,
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(res)
	}
}
