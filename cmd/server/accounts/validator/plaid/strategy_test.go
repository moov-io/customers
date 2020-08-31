// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package plaid

import (
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/moov-io/base"
	"github.com/moov-io/customers/client"
	"github.com/moov-io/customers/cmd/server/accounts/validator"
	"github.com/plaid/plaid-go/plaid"
	"github.com/stretchr/testify/require"
)

func TestStrategy(t *testing.T) {
	if os.Getenv("PLAID_CLIENT_ID") == "" {
		t.Skip("No configuration found for Plaid")
	}

	options := StrategyOptions{
		os.Getenv("PLAID_CLIENT_ID"),
		os.Getenv("PLAID_SECRET"),
		"sandbox",
		"Moov Test",
	}

	strategy, err := NewStrategy(options)
	require.NoError(t, err)

	customerID, userID, accountID := base.ID(), base.ID(), base.ID()

	t.Run("Test InitAccountValidation", func(t *testing.T) {
		initResponse, err := strategy.InitAccountValidation(userID, accountID, customerID)
		require.NoError(t, err)
		fmt.Println(initResponse)

		require.Contains(t, (*initResponse)["link_token"], "link-sandbox-")
	})

	t.Run("Test CompleteAccountValidation", func(t *testing.T) {
		// In order to test CompleteAccountValidation we need to obtain
		// public token. In production we receive it from Plaid Link (widget)
		// for test purposes we will use Plaid /sandbox/public_token/create
		// to create it.
		// More information here: https://plaid.com/docs/api/2017-03-08/#sandbox-institutions
		var (
			sandboxInstitution = "ins_109508"
			testProducts       = []string{"auth"}
		)

		var testClient, _ = plaid.NewClient(plaid.ClientOptions{
			ClientID:    os.Getenv("PLAID_CLIENT_ID"),
			Secret:      os.Getenv("PLAID_SECRET"),
			Environment: plaid.Sandbox,
			HTTPClient:  &http.Client{},
		})

		sandboxResp, err := testClient.CreateSandboxPublicToken(sandboxInstitution, testProducts)
		require.NoError(t, err)

		request := &validator.VendorRequest{
			"public_token": sandboxResp.PublicToken,
		}

		// create account that corresponds to Plaid sandbox account
		account := &client.Account{
			AccountID:     "xxx",
			RoutingNumber: "011401533",
		}
		accountNumber := "1111222233330000"

		response, err := strategy.CompleteAccountValidation(userID, customerID, account, accountNumber, request)
		require.NoError(t, err)
		fmt.Println(response)

		require.Contains(t, (*response)["result"], "validated")
	})
}
