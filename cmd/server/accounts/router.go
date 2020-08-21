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
	"github.com/moov-io/customers/cmd/server/fed"
	"github.com/moov-io/customers/cmd/server/paygate"
	"github.com/moov-io/customers/cmd/server/route"
	"github.com/moov-io/customers/pkg/secrets"
	"github.com/moov-io/customers/pkg/secrets/hash"
	"github.com/moov-io/customers/pkg/secrets/mask"

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

		customerID := route.GetCustomerID(w, r)
		accounts, err := repo.getCustomerAccounts(customerID)
		if err != nil {
			moovhttp.Problem(w, err)
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

type createAccountRequest struct {
	HolderName    string             `json:"holderName"`
	AccountNumber string             `json:"accountNumber"`
	RoutingNumber string             `json:"routingNumber"`
	Type          client.AccountType `json:"type"`

	// fields we compute from the inbound AccountNumber
	encryptedAccountNumber string
	hashedAccountNumber    string
	maskedAccountNumber    string
}

func (req *createAccountRequest) validate() error {
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
	if !at(req.Type, client.CHECKING) && !at(req.Type, client.SAVINGS) {
		return fmt.Errorf("invalid account type: %s", req.Type)
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

func createCustomerAccount(logger log.Logger, repo Repository, fedClient fed.Client, keeper *secrets.StringKeeper) http.HandlerFunc {
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

		if _, err := fedClient.LookupInstitution(request.RoutingNumber); err != nil {
			logger.Log("accounts", fmt.Sprintf("problem looking up routing number=%q: %v", request.RoutingNumber, err), "requestID", moovhttp.GetRequestID(r))
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
