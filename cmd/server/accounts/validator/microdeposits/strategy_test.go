package microdeposits

import (
	"testing"

	"github.com/moov-io/base"
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
	require.Equal(t, &validator.VendorResponse{}, response)
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

	response, err := strategy.CompleteAccountValidation("userID", "accountID", "customerID", request)
	require.NoError(t, err)
	require.Equal(t, &validator.VendorResponse{}, response)
}
