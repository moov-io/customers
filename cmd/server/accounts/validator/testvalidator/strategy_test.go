package testvalidator

import (
	"testing"

	"github.com/moov-io/customers/cmd/server/accounts/validator"
	"github.com/stretchr/testify/require"
)

func TestStrategy(t *testing.T) {
	strategy := NewStrategy()
	initResponse, err := strategy.InitAccountValidation("userID", "accountID", "customerID")
	require.NoError(t, err)
	require.Equal(t, "success", (*initResponse)["result"])

	// test successful completion
	request := &validator.VendorRequest{
		"result": "success",
	}

	response, err := strategy.CompleteAccountValidation("userID", "accountID", "customerID", request)
	require.NoError(t, err)
	require.Equal(t, "success", (*response)["result"])

	// test error
	request = &validator.VendorRequest{
		"result": "error",
	}

	response, err = strategy.CompleteAccountValidation("userID", "accountID", "customerID", request)
	require.Error(t, err)
}
