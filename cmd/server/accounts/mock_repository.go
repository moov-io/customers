// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package accounts

import (
	"github.com/moov-io/customers/admin"
	"github.com/moov-io/customers/client"
)

type mockRepository struct {
	Accounts      []*client.Account
	AccountNumber string
	Err           error
}

func (r *mockRepository) getCustomerAccounts(customerID string) ([]*client.Account, error) {
	if r.Err != nil {
		return nil, r.Err
	}
	return r.Accounts, nil
}

func (r *mockRepository) createCustomerAccount(customerID, userID string, req *createAccountRequest) (*client.Account, error) {
	if r.Err != nil {
		return nil, r.Err
	}
	if len(r.Accounts) > 0 {
		return r.Accounts[0], nil
	}
	return nil, nil
}

func (r *mockRepository) deactivateCustomerAccount(accountID string) error {
	return r.Err
}

func (r *mockRepository) updateAccountStatus(accountID string, status admin.AccountStatus) error {
	return r.Err
}

func (r *mockRepository) getEncryptedAccountNumber(customerID, accountID string) (string, error) {
	if r.Err != nil {
		return "", r.Err
	}
	return r.AccountNumber, nil
}
