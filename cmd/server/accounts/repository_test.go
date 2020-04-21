// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package accounts

import (
	"testing"

	"github.com/moov-io/base"
	"github.com/moov-io/customers/client"
	"github.com/moov-io/customers/internal/database"

	"github.com/go-kit/kit/log"
)

type testAccountRepository struct {
	Repository

	db *database.TestSQLiteDB
}

func setupTestAccountRepository(t *testing.T) *testAccountRepository {
	db := database.CreateTestSqliteDB(t)
	repo := NewRepo(log.NewNopLogger(), db.DB)

	t.Cleanup(func() {
		db.Close()
		repo.Close()
	})

	return &testAccountRepository{Repository: repo, db: db}
}

func TestRepository(t *testing.T) {
	customerID, userID := base.ID(), base.ID()
	repo := setupTestAccountRepository(t)

	// initial read, find no accounts
	accounts, err := repo.getCustomerAccounts(customerID)
	if len(accounts) != 0 || err != nil {
		t.Fatalf("got accounts=%#v error=%v", accounts, err)
	}

	// create account
	acct, err := repo.createCustomerAccount(customerID, userID, &createAccountRequest{
		AccountNumber: "123",
		RoutingNumber: "987654320",
		Type:          client.CHECKING,
		HolderType:    client.INDIVIDUAL,
	})
	if err != nil {
		t.Fatal(err)
	}

	// read after creating
	accounts, err = repo.getCustomerAccounts(customerID)
	if len(accounts) != 1 || err != nil {
		t.Fatalf("got accounts=%#v error=%v", accounts, err)
	}
	if accounts[0].AccountID != acct.AccountID {
		t.Errorf("accounts[0].AccountID=%s acct.AccountID=%s", accounts[0].AccountID, acct.AccountID)
	}

	// delete, expect no accounts
	if err := repo.deactivateCustomerAccount(acct.AccountID); err != nil {
		t.Fatal(err)
	}
	accounts, err = repo.getCustomerAccounts(customerID)
	if len(accounts) != 0 || err != nil {
		t.Fatalf("got accounts=%#v error=%v", accounts, err)
	}
}
