package plaid

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	moovhttp "github.com/moov-io/base/http"
	"github.com/moov-io/customers/cmd/server/accounts"
	"github.com/moov-io/customers/cmd/server/route"
	"github.com/moov-io/customers/cmd/server/verification"
	"github.com/moov-io/customers/pkg/secrets"
	"github.com/plaid/plaid-go/plaid"
)

var _ verification.AccountVerifier = (*verifier)(nil)

type verifier struct {
	client         *plaid.Client
	linkClientName string
}

type VerifierOptions struct {
	ClientID    string
	Secret      string
	Environment string
	ClientName  string
}

var environments = map[string]plaid.Environment{
	"sandbox":     plaid.Sandbox,
	"development": plaid.Development,
	"production":  plaid.Production,
}

func Factory() (verification.AccountVerifier, error) {
	options := VerifierOptions{
		os.Getenv("PLAID_CLIENT_ID"),
		os.Getenv("PLAID_SECRET"),
		os.Getenv("PLAID_ENVIRONMENT"),
		os.Getenv("PLAID_CLIENT_NAME"),
	}

	return New(options)
}

func New(options VerifierOptions) (*verifier, error) {
	plaidOptions := plaid.ClientOptions{
		options.ClientID,
		options.Secret,
		environments[options.Environment],
		&http.Client{},
	}

	client, err := plaid.NewClient(plaidOptions)
	if err != nil {
		return nil, err
	}

	return &verifier{
		client:         client,
		linkClientName: options.ClientName,
	}, nil
}

func (v *verifier) InitiateAccountVerification() http.HandlerFunc {
	type response struct {
		LinkToken  string    `json:"link_token"`
		Expiration time.Time `json:"expiration"`
	}

	return func(w http.ResponseWriter, r *http.Request) {
		customerID := route.GetCustomerID(w, r)

		tokenResponse, err := v.client.CreateLinkToken(plaid.LinkTokenConfigs{
			User: &plaid.LinkTokenUser{
				ClientUserID: customerID,
			},
			ClientName:   v.linkClientName,
			Products:     []string{"auth"},
			CountryCodes: []string{"US"},
			Language:     "en",
		})
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}

		resp := &response{
			LinkToken:  tokenResponse.LinkToken,
			Expiration: tokenResponse.Expiration,
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}
}

func (v *verifier) CompleteAccountVerification(repo accounts.Repository, keeper *secrets.StringKeeper) http.HandlerFunc {
	type request struct {
		PublicToken string
	}

	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		customerID, accountID := vars["customerID"], vars["accountID"]

		if customerID == "" || accountID == "" {
			moovhttp.Problem(w, fmt.Errorf("missing customer=%s and/or account=%s", customerID, accountID))
			return
		}

		req := &request{}
		if err := json.NewDecoder(r.Body).Decode(req); err != nil {
			moovhttp.Problem(w, err)
			return
		}

		tokenResponse, err := v.client.ExchangePublicToken(req.PublicToken)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}

		authResponse, err := v.client.GetAuth(tokenResponse.AccessToken)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}

		// Get customer account

		account := repo.getCustomerAccount(customerID, accountID)
		// ASK: how can I get decrypted account number?
		// getEncryptedAccountNumber is private method (as all methods of the repository)
		encrypted, err := repo.getEncryptedAccountNumber(customerID, accountID)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}
		// decrypt from database
		decryptedAccountNumber, err := keeper.DecryptString(encrypted)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}

		// compare what do we get from Plaid
		// for _, acc := range authResponse.Numbers.ACH {
		// 	if acc.Account == "" && acc.Routing = "" {
		// 		// set account as verified
		// 	}
		// }

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(authResponse)
	}
}
