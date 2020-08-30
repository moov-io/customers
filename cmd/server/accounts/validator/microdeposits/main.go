package microdeposits

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/mitchellh/mapstructure"
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
	// or we can create two more new micro deposits (not sure if it makes sense)?
	micro, err := t.client.GetMicroDeposits(accountID, userID)
	if err != nil {
		return nil, fmt.Errorf("problem reading micro-deposits for accountID=%s: %v", accountID, err)
	}

	// TODO
	// if micro-deposits was found then return error as
	// you can't run microdeposits twice?

	// If no micro-deposit was found then initiate them.
	// TODO why do we check both micro and MicroDepositID?
	// is it possible to get micro with empty MicroDepositID?
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

// here we can use struct generated from open-api
// but not sure how it will work with mapstructure.Decode :)
type completeAccountValidationRequest struct {
	MicroDeposits []string `json:"micro-deposits,omitempty" mapstructure:"micro-deposits"`
}

func (t *microdepositsStrategy) CompleteAccountValidation(userID, accountID, customerID string, request *validator.VendorRequest) (*validator.VendorResponse, error) {

	micro, err := t.client.GetMicroDeposits(accountID, userID)
	if err != nil {
		return nil, fmt.Errorf("problem reading micro-deposits for accountID=%s: %v", accountID, err)
	}

	// If no micro-deposit was found then no init call was made?
	if micro == nil || micro.MicroDepositID == "" {
		return nil, fmt.Errorf("no micro-deposits was found")
	}

	// If the micro-deposits have been processed then require amounts as we will only call
	// handleMicroDepositValidation when the account needs to be VALIDATED still.
	if strings.EqualFold(string(micro.Status), string(client.PROCESSED)) {
		// Check the amounts in the request against what PayGate created
		input := &completeAccountValidationRequest{}
		if err := mapstructure.Decode(request, input); err != nil {
			return nil, fmt.Errorf("unable to parse request params: %v", err)
		}

		if err := validateAmounts(micro, input.MicroDeposits); err != nil {
			return nil, err
		}
		// Amounts validated, so mark the Account has approved
		return &validator.VendorResponse{}, nil
	}

	return nil, fmt.Errorf("microDepositID=%s is in status: %s", micro.MicroDepositID, micro.Status)
}

func validateAmounts(micro *client.MicroDeposits, requestAmounts []string) error {
	if len(requestAmounts) == 0 {
		return errors.New("missing micro-deposits for validation")
	}
	sort.Strings(requestAmounts)

	requiredAmounts := micro.Amounts
	sort.Strings(requiredAmounts)

	if len(requestAmounts) != len(requiredAmounts) {
		return fmt.Errorf("invalid number of micro-deposits, got %d", len(requestAmounts))
	}
	for i := range requestAmounts {
		if requestAmounts[i] != requiredAmounts[i] {
			return errors.New("incorrect micro-deposit")
		}
	}

	return nil
}
