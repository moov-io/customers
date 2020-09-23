// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package microdeposits

import (
	"testing"

	"github.com/moov-io/base"
	"github.com/moov-io/customers/pkg/client"
	"github.com/moov-io/customers/pkg/paygate"
	"github.com/moov-io/customers/pkg/validator"
	"github.com/stretchr/testify/require"

	payclient "github.com/moov-io/paygate/pkg/client"
)

func TestInitAccountValidation(t *testing.T) {
	t.Run("When no micro-desposits was created before", func(t *testing.T) {
		paygateClient := &paygate.MockClient{}
		strategy := NewStrategy(paygateClient)

		response, err := strategy.InitAccountValidation("userID", "accountID", "customerID")
		require.NoError(t, err)
		require.Equal(t, &validator.VendorResponse{"result": "initiated"}, response)
	})

	t.Run("When micro-desposits was created before", func(t *testing.T) {
		paygateClient := &paygate.MockClient{
			Micro: &payclient.MicroDeposits{
				MicroDepositID: base.ID(),
				Amounts:        []string{"USD 0.03", "USD 0.07"},
				Status:         payclient.PROCESSED,
			},
		}
		strategy := NewStrategy(paygateClient)
		_, err := strategy.InitAccountValidation("userID", "accountID", "customerID")
		require.Error(t, err, "micro-deposits were already created for accountID=accountID")
	})
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

	t.Run("Test when micro-deposits were processed", func(t *testing.T) {
		paygateClient.Micro.Status = payclient.PROCESSED

		response, err := strategy.CompleteAccountValidation("userID", "customerID", account, accountNumber, request)
		require.NoError(t, err)
		require.Equal(t, &validator.VendorResponse{"result": "validated"}, response)
	})

	t.Run("Test when micro-deposits status in not processed", func(t *testing.T) {
		paygateClient.Micro.Status = payclient.PENDING

		_, err := strategy.CompleteAccountValidation("userID", "customerID", account, accountNumber, request)
		require.Error(t, err)
		require.Contains(t, err.Error(), "is in status: pending but expected to be in processed")
	})
}
