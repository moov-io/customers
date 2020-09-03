// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package microdeposits

import (
	"testing"

	"github.com/moov-io/base"
	"github.com/moov-io/customers/client"
	"github.com/moov-io/customers/cmd/server/accounts/validator"
	"github.com/moov-io/customers/cmd/server/paygate"
	"github.com/stretchr/testify/require"

	payclient "github.com/moov-io/paygate/pkg/client"
)

func TestInitAccountValidation(t *testing.T) {
	paygateClient := &paygate.MockClient{}
	strategy := NewStrategy(paygateClient)

	response, err := strategy.InitAccountValidation("userID", "accountID", "customerID")
	require.NoError(t, err)
	require.Equal(t, &validator.VendorResponse{"result": "initiated"}, response)
}

func TestCompleteAccountValidation(t *testing.T) {
	paygateClient := &paygate.MockClient{
		Micro: &payclient.MicroDeposits{
			MicroDepositID: base.ID(),
			Amounts:        []string{"USD 0.03", "USD 0.07"},
			Status:         payclient.PROCESSED,
		},
	}
	strategy := NewStrategy(paygateClient)

	// test successful completion
	request := &validator.VendorRequest{
		"micro-deposits": []string{"USD 0.03", "USD 0.07"},
	}

	account := &client.Account{
		AccountID:     "xxx",
		RoutingNumber: "xxx",
	}
	accountNumber := "xxx"

	response, err := strategy.CompleteAccountValidation("userID", "customerID", account, accountNumber, request)
	require.NoError(t, err)
	require.Equal(t, &validator.VendorResponse{"result": "validated"}, response)
}
