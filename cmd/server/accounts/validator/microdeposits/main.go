package microdeposits

import (
	"fmt"

	"github.com/moov-io/customers/cmd/server/accounts/validator"
	"github.com/moov-io/customers/cmd/server/paygate"
	"github.com/moov-io/paygate/pkg/client"
)

type microdepositsStrategy struct {
	client paygate.Client
}

func NewStrategy(paygateClient paygate.Client) validator.Strategy {
	return &microdepositsStrategy{
		client: paygateClient,
	}
}

func (t *microdepositsStrategy) InitAccountValidation(userID, accountID, customerID string) (*validator.VendorResponse, error) {
	// TODO let's disucss this:
	// we should expect that if we are here it means that no micro-depists was created before
	// or we can create two more new micro deposits (not sure if it makes sense)
	// if we have to support previous version where there is no Validation record in DB
	// we will pull info about micro-deposits from PayGate
	micro, err := t.client.GetMicroDeposits(accountID, userID)
	if err != nil {
		return nil, fmt.Errorf("problem reading micro-deposits for accountID=%s: %v", accountID, err)
	}

	// If no micro-deposit was found then initiate them.
	if micro == nil || micro.MicroDepositID == "" {
		err = t.client.InitiateMicroDeposits(userID, client.Destination{
			CustomerID: customerID,
			AccountID:  accountID,
		})
		if err != nil {
			return nil, fmt.Errorf("problem initiating micro-deposits for acocuntID=%s: %v", accountID, err)
		}
	}

	return &validator.VendorResponse{}, nil
}
