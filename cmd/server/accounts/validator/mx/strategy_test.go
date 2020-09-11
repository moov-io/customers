// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package mx

import (
	"os"
	"testing"

	"github.com/moov-io/base"
	"github.com/moov-io/customers/cmd/server/accounts/validator"
	"github.com/moov-io/customers/pkg/client"
	"github.com/stretchr/testify/require"
)

func TestStrategy(t *testing.T) {
	if os.Getenv("ATRIUM_API_KEY") == "" {
		t.Skip("No configuration found for MX")
	}

	options := StrategyOptions{
		ClientID: os.Getenv("ATRIUM_CLIENT_ID"),
		APIKey:   os.Getenv("ATRIUM_API_KEY"),
	}

	strategy := NewStrategy(options)

	customerID, userID, accountID := base.ID(), base.ID(), base.ID()

	t.Run("Test InitAccountValidation", func(t *testing.T) {
		initResponse, err := strategy.InitAccountValidation(userID, accountID, customerID)
		require.NoError(t, err)
		require.Contains(t, (*initResponse)["connect_widget_url"], "https://int-widgets.moneydesktop.com")
	})

	t.Run("Test CompleteAccountValidation", func(t *testing.T) {
		// To test out setup without incurring the cost of calling the
		// verify endpoint we should use user GUID "test_atrium" and
		// the member GUID "test_atrium_member". More information here:
		// https://atrium.mx.com/docs/getting_started/verification#testing-verification
		var (
			userGUID   = "test_atrium"
			memberGUID = "test_atrium_member"
		)

		request := &validator.VendorRequest{
			"user_guid":   userGUID,
			"member_guid": memberGUID,
		}

		// create account that corresponds to account of MX test user
		account := &client.Account{
			AccountID:     "test_acc",
			RoutingNumber: "68899990000000",
		}
		accountNumber := "10001"

		response, err := strategy.CompleteAccountValidation(userID, customerID, account, accountNumber, request)
		require.NoError(t, err)
		require.Contains(t, (*response)["result"], "validated")
	})
}
