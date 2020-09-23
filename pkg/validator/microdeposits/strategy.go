// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package microdeposits

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/mitchellh/mapstructure"
	customersclient "github.com/moov-io/customers/pkg/client"
	"github.com/moov-io/customers/pkg/paygate"
	"github.com/moov-io/customers/pkg/validator"
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
	micro, err := t.client.GetMicroDeposits(accountID, userID)
	if err != nil {
		return nil, fmt.Errorf("problem reading micro-deposits for accountID=%s: %v", accountID, err)
	}

	if micro != nil && micro.MicroDepositID != "" {
		return nil, fmt.Errorf("micro-deposits were already created for accountID=%s", accountID)
	}

	err = t.client.InitiateMicroDeposits(userID, client.Destination{
		CustomerID: customerID,
		AccountID:  accountID,
	})
	if err != nil {
		return nil, fmt.Errorf("problem initiating micro-deposits for accountID=%s: %v", accountID, err)
	}

	return &validator.VendorResponse{
		"result": "initiated",
	}, nil
}

type completeAccountValidationRequest struct {
	MicroDeposits []string `json:"micro-deposits,omitempty" mapstructure:"micro-deposits"`
}

func (t *microdepositsStrategy) CompleteAccountValidation(userID, customerID string, account *customersclient.Account, accountID string, request *validator.VendorRequest) (*validator.VendorResponse, error) {
	micro, err := t.client.GetMicroDeposits(accountID, userID)
	if err != nil {
		return nil, fmt.Errorf("problem reading micro-deposits for accountID=%s: %v", accountID, err)
	}

	// If no micro-deposit was found then no init call was made?
	if micro == nil || micro.MicroDepositID == "" {
		return nil, fmt.Errorf("no micro-deposits record was found")
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

		return &validator.VendorResponse{
			"result": "validated",
		}, nil
	}

	return nil, fmt.Errorf("microDepositID=%s is in status: %s but expected to be in %s", micro.MicroDepositID, micro.Status, client.PROCESSED)
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
