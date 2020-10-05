// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package accounts

import (
	"github.com/moov-io/customers/pkg/client"
)

var _ Repository = (*mockRepository)(nil)

type mockRepository struct {
	Accounts      []*client.Account
	AccountNumber string
	Err           error
}

func (r *mockRepository) getCustomerAccount(customerID, accountID string) (*client.Account, error) {
	if r.Err != nil {
		return nil, r.Err
	}
	if len(r.Accounts) > 0 {
		return r.Accounts[0], nil
	}
	return nil, nil
}

func (r *mockRepository) GetCustomerAccountsByIDs(accountIDs []string) ([]*client.Account, error) {
	if r.Err != nil {
		return nil, r.Err
	}
	return r.Accounts, nil
}

func (r *mockRepository) getAccountsByCustomerID(customerID string) ([]*client.Account, error) {
	if r.Err != nil {
		return nil, r.Err
	}
	return r.Accounts, nil
}

func (r *mockRepository) CreateCustomerAccount(customerID, userID string, req *CreateAccountRequest) (*client.Account, error) {
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

func (r *mockRepository) updateAccountStatus(accountID string, status client.AccountStatus) error {
	return r.Err
}

func (r *mockRepository) getEncryptedAccountNumber(customerID, accountID string) (string, error) {
	if r.Err != nil {
		return "", r.Err
	}
	return r.AccountNumber, nil
}

func (r *mockRepository) getLatestAccountOFACSearch(accountID string) (*client.OfacSearch, error) {
	panic("implement me")
}

func (r *mockRepository) saveAccountOFACSearch(id string, result *client.OfacSearch) error {
	return nil
}
