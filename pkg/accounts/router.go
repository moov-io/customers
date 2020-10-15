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

	"github.com/moov-io/customers/pkg/client"
	"github.com/moov-io/customers/pkg/fed"
	"github.com/moov-io/customers/pkg/route"
	"github.com/moov-io/customers/pkg/secrets"
	"github.com/moov-io/customers/pkg/secrets/hash"
	"github.com/moov-io/customers/pkg/secrets/mask"
	"github.com/moov-io/customers/pkg/validator"

	"github.com/gorilla/mux"
	"github.com/moov-io/base/log"
)

func RegisterRoutes(logger log.Logger, r *mux.Router, accounts Repository, validations validator.Repository, fedClient fed.Client, keeper, transitKeeper *secrets.StringKeeper, validationStrategies map[validator.StrategyKey]validator.Strategy, ofac *AccountOfacSearcher) {
	logger = logger.WithKeyValue("package", "accounts")

	r.Methods("GET").Path("/customers/{customerID}/accounts").HandlerFunc(getCustomerAccounts(logger, accounts, fedClient))
	r.Methods("POST").Path("/customers/{customerID}/accounts").HandlerFunc(createCustomerAccount(logger, accounts, fedClient, keeper, ofac))
	r.Methods("POST").Path("/customers/{customerID}/accounts/{accountID}/decrypt").HandlerFunc(decryptAccountNumber(logger, accounts, keeper, transitKeeper))
	r.Methods("DELETE").Path("/customers/{customerID}/accounts/{accountID}").HandlerFunc(removeCustomerAccount(logger, accounts))

	r.Methods("GET").Path("/customers/{customerID}/accounts/{accountID}").HandlerFunc(getCustomerAccountByID(logger, accounts, fedClient))
	r.Methods("GET").Path("/customers/{customerID}/accounts/{accountID}/ofac").HandlerFunc(getAccountOfacSearch(logger, accounts))
	r.Methods("PUT").Path("/customers/{customerID}/accounts/{accountID}/refresh/ofac").HandlerFunc(refreshAccountOfac(logger, accounts, ofac))

	r.Methods("PUT").Path("/customers/{customerID}/accounts/{accountID}/status").HandlerFunc(updateAccountStatus(logger, accounts))

	r.Methods("POST").Path("/customers/{customerID}/accounts/{accountID}/validations").HandlerFunc(initAccountValidation(logger, accounts, validations, validationStrategies))
	r.Methods("GET").Path("/customers/{customerID}/accounts/{accountID}/validations/{validationID}").HandlerFunc(getAccountValidation(logger, accounts, validations))
	r.Methods("PUT").Path("/customers/{customerID}/accounts/{accountID}/validations/{validationID}").HandlerFunc(completeAccountValidation(logger, accounts, validations, keeper, validationStrategies))
}

func getCustomerAccounts(logger log.Logger, repo Repository, fedClient fed.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)

		customerID := route.GetCustomerID(w, r)
		if customerID == "" {
			return
		}

		organization := route.GetOrganization(w, r)
		if organization == "" {
			return
		}

		accounts, err := repo.getAccounts(customerID, organization)
		if err != nil {
			moovhttp.Problem(w, fmt.Errorf("getting accounts: %v", err))
			return
		}
		accounts = decorateInstitutionDetails(accounts, fedClient)

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(accounts)
	}
}

func decorateInstitutionDetails(accounts []*client.Account, client fed.Client) []*client.Account {
	for i := range accounts {
		if details, _ := client.LookupInstitution(accounts[i].RoutingNumber); details != nil {
			accounts[i].Institution = *details
		}
	}
	return accounts
}

type CreateAccountRequest struct {
	HolderName    string             `json:"holderName"`
	AccountNumber string             `json:"accountNumber"`
	RoutingNumber string             `json:"routingNumber"`
	Type          client.AccountType `json:"type"`

	// fields we compute from the inbound AccountNumber
	encryptedAccountNumber string
	hashedAccountNumber    string
	maskedAccountNumber    string
}

func (req *CreateAccountRequest) validate() error {
	if req.HolderName == "" {
		return errors.New("missing HolderName")
	}
	if req.AccountNumber == "" {
		return errors.New("missing AccountNumber")
	}
	if err := ach.CheckRoutingNumber(req.RoutingNumber); err != nil {
		return err
	}

	at := func(t1, t2 client.AccountType) bool {
		return strings.EqualFold(string(t1), string(t2))
	}
	if !at(req.Type, client.ACCOUNTTYPE_CHECKING) && !at(req.Type, client.ACCOUNTTYPE_SAVINGS) {
		return fmt.Errorf("invalid account type: %s", req.Type)
	}

	return nil
}

func (req *CreateAccountRequest) disfigure(keeper *secrets.StringKeeper) error {
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

func createCustomerAccount(logger log.Logger, repo Repository, fedClient fed.Client, keeper *secrets.StringKeeper, ofac *AccountOfacSearcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)

		requestID := moovhttp.GetRequestID(r)

		var request CreateAccountRequest
		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			moovhttp.Problem(w, err)
			return
		}
		if err := request.validate(); err != nil {
			logger.LogErrorF("problem validating account: %v", err)
			moovhttp.Problem(w, err)
			return
		}
		if err := request.disfigure(keeper); err != nil {
			logger.LogErrorF("problem disfiguring account: %v", err)
			moovhttp.Problem(w, err)
			return
		}

		if _, err := fedClient.LookupInstitution(request.RoutingNumber); err != nil {
			logger.LogErrorF("problem looking up routing number=%q: %v", request.RoutingNumber, err)
			moovhttp.Problem(w, err)
			return
		}

		customerID, userID := route.GetCustomerID(w, r), moovhttp.GetUserID(r)
		account, err := repo.CreateCustomerAccount(customerID, userID, &request)
		if err != nil {
			logger.LogErrorF("problem saving account: %v", err)
			moovhttp.Problem(w, err)
			return
		}

		// Perform an OFAC search with the Customer information
		if err := ofac.StoreAccountOFACSearch(account, requestID); err != nil {
			logger.LogErrorF("error with OFAC search for account=%s: %v", account.AccountID, err)
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(account)
	}
}

func getCustomerAccountByID(logger log.Logger, repo Repository, fedClient fed.Client) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)
		customerID := route.GetCustomerID(w, r)
		accountID := getAccountID(w, r)

		account, err := repo.getCustomerAccount(customerID, accountID)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}

		if account == nil {
			moovhttp.Problem(w, fmt.Errorf("account with customerID=%s and accountID=%s not found", customerID, accountID))
			return
		}

		details, err := fedClient.LookupInstitution(account.RoutingNumber)
		if err != nil {
			moovhttp.Problem(w, fmt.Errorf("looking up institution details: %v", err))
			return
		}

		if details != nil {
			account.Institution = *details
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

func decryptAccountNumber(logger log.Logger, repo Repository, keeper *secrets.StringKeeper, transitKeeper *secrets.StringKeeper) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)

		customerID, accountID := route.GetCustomerID(w, r), getAccountID(w, r)
		if customerID == "" || accountID == "" {
			return
		}

		// get organization
		organization := route.GetOrganization(w, r)

		// grab encrypted value
		encrypted, err := repo.getEncryptedAccountNumber(organization, customerID, accountID)
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

func getAccountOfacSearch(logger log.Logger, repo Repository) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		accountID := getAccountID(w, r)
		if accountID == "" {
			return
		}

		result, err := repo.getLatestAccountOFACSearch(accountID)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(result)
	}
}

func refreshAccountOfac(logger log.Logger, repo Repository, ofac *AccountOfacSearcher) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		requestID := moovhttp.GetRequestID(r)
		customerID, accountID := route.GetCustomerID(w, r), getAccountID(w, r)
		if customerID == "" || accountID == "" {
			moovhttp.Problem(w, errors.New("customerID and accountID required"))
			return
		}
		account, err := repo.getCustomerAccount(customerID, accountID)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}
		err = ofac.StoreAccountOFACSearch(account, requestID)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}
		result, err := repo.getLatestAccountOFACSearch(accountID)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(result)
	}
}

func updateAccountStatus(logger log.Logger, repo Repository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		customerID, accountID := route.GetCustomerID(w, r), getAccountID(w, r)
		if customerID == "" || accountID == "" {
			moovhttp.Problem(w, errors.New("customerID and accountID required"))
			return
		}

		var req client.UpdateAccountStatus
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			moovhttp.Problem(w, err)
			return
		}
		switch req.Status {
		case client.ACCOUNTSTATUS_NONE, client.ACCOUNTSTATUS_VALIDATED:
			// do nothing
		default:
			moovhttp.Problem(w, fmt.Errorf("invalid status: %s", req.Status))
			return
		}

		if err := repo.updateAccountStatus(accountID, req.Status); err != nil {
			moovhttp.Problem(w, err)
			return
		}

		account, err := repo.getCustomerAccount(customerID, accountID)
		if err != nil {
			moovhttp.Problem(w, errors.New("there was an error getting the account"))
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(account)
	}
}
