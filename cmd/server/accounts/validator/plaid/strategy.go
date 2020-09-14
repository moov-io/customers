// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package plaid

import (
	"fmt"
	"net/http"

	"github.com/mitchellh/mapstructure"
	"github.com/moov-io/customers/cmd/server/accounts/validator"
	"github.com/moov-io/customers/pkg/client"
	"github.com/plaid/plaid-go/plaid"
)

type plaidStrategy struct {
	client         *plaid.Client
	linkClientName string
}

type StrategyOptions struct {
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

func NewStrategy(options StrategyOptions) (validator.Strategy, error) {
	env, found := environments[options.Environment]
	if !found {
		return nil, fmt.Errorf("unsupported environment %q is specified for Plaid", options.Environment)
	}

	plaidOptions := plaid.ClientOptions{
		ClientID:    options.ClientID,
		Secret:      options.Secret,
		Environment: env,
		HTTPClient:  &http.Client{},
	}

	client, err := plaid.NewClient(plaidOptions)
	if err != nil {
		return nil, err
	}

	return &plaidStrategy{
		client:         client,
		linkClientName: options.ClientName,
	}, nil
}

func (s *plaidStrategy) InitAccountValidation(userID, accountID, customerID string) (*validator.VendorResponse, error) {
	tokenResponse, err := s.client.CreateLinkToken(plaid.LinkTokenConfigs{
		User: &plaid.LinkTokenUser{
			ClientUserID: customerID,
		},
		ClientName:   s.linkClientName,
		Products:     []string{"auth"},
		CountryCodes: []string{"US"},
		Language:     "en",
	})
	if err != nil {
		return nil, err
	}

	return &validator.VendorResponse{
		"link_token": tokenResponse.LinkToken,
		"expiration": tokenResponse.Expiration,
	}, nil
}

type completeAccountValidationRequest struct {
	PublicToken string `json:"public_token" mapstructure:"public_token"`
}

func (s *plaidStrategy) CompleteAccountValidation(userID, customerID string, account *client.Account, accountNumber string, request *validator.VendorRequest) (*validator.VendorResponse, error) {
	input := &completeAccountValidationRequest{}
	if err := mapstructure.Decode(request, input); err != nil {
		return nil, fmt.Errorf("unable to parse request params: %v", err)
	}

	tokenResponse, err := s.client.ExchangePublicToken(input.PublicToken)
	if err != nil {
		return nil, err
	}

	authResponse, err := s.client.GetAuth(tokenResponse.AccessToken)
	if err != nil {
		return nil, err
	}

	// look for account number and routing number in Plaid results
	for _, acc := range authResponse.Numbers.ACH {
		if acc.Account == accountNumber && acc.Routing == account.RoutingNumber {
			return &validator.VendorResponse{
				"result": "validated",
			}, nil
		}
	}

	return nil, fmt.Errorf("failed to validate account=%s", account.AccountID)
}
