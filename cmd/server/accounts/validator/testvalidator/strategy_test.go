// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package testvalidator

import (
	"testing"

	"github.com/moov-io/customers/client"
	"github.com/moov-io/customers/cmd/server/accounts/validator"
	"github.com/stretchr/testify/require"
)

func TestStrategy(t *testing.T) {
	strategy := NewStrategy()
	initResponse, err := strategy.InitAccountValidation("userID", "accountID", "customerID")
	require.NoError(t, err)
	require.Equal(t, "initiated", (*initResponse)["result"])

	// test successful completion
	request := &validator.VendorRequest{
		"result": "success",
	}

	account := &client.Account{
		AccountID:     "xxx",
		RoutingNumber: "xxx",
	}

	response, err := strategy.CompleteAccountValidation("userID", "customerID", account, "accountNumber", request)
	require.NoError(t, err)
	require.Equal(t, "validated", (*response)["result"])

	// test error
	request = &validator.VendorRequest{
		"result": "error",
	}

	response, err = strategy.CompleteAccountValidation("userID", "customerID", account, "accountNumber", request)
	require.Error(t, err)
}
