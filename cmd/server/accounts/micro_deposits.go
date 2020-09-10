// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package accounts

import (
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/moov-io/customers/cmd/server/paygate"
	"github.com/moov-io/customers/pkg/admin"
	"github.com/moov-io/paygate/pkg/client"
)

func handleMicroDepositValidation(repo Repository, paygateClient paygate.Client, accountID, customerID, userID string, microDeposits []string) error {
	micro, err := paygateClient.GetMicroDeposits(accountID, userID)
	if err != nil {
		return fmt.Errorf("problem reading micro-deposits for accountID=%s: %v", accountID, err)
	}

	// If no micro-deposit was found then initiate them.
	if micro == nil || micro.MicroDepositID == "" {
		err = paygateClient.InitiateMicroDeposits(userID, client.Destination{
			CustomerID: customerID,
			AccountID:  accountID,
		})
		if err != nil {
			return fmt.Errorf("problem initiating micro-deposits for acocuntID=%s: %v", accountID, err)
		}
		return nil
	}

	// If the micro-deposits have been processed then require amounts as we will only call
	// handleMicroDepositValidation when the account needs to be VALIDATED still.
	if strings.EqualFold(string(micro.Status), string(client.PROCESSED)) {
		// Check the amounts in the request against what PayGate created
		if err := validateAmounts(micro, microDeposits); err != nil {
			return err
		}
		// Amounts validated, so mark the Account has approved
		return repo.updateAccountStatus(accountID, admin.VALIDATED)
	}

	return fmt.Errorf("microDepositID=%s is in status: %s", micro.MicroDepositID, micro.Status)
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
