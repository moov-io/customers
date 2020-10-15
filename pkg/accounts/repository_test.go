// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package accounts

import (
	"testing"

	"github.com/moov-io/base"
	"github.com/stretchr/testify/require"

	"github.com/moov-io/customers/internal/database"
	"github.com/moov-io/customers/pkg/client"
	"github.com/moov-io/customers/pkg/customers"
	"github.com/moov-io/customers/pkg/secrets"

	"github.com/moov-io/base/log"
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
	})

	return &testAccountRepository{Repository: repo, db: db}
}

func TestRepository(t *testing.T) {
	userID := base.ID()
	logger := log.NewNopLogger()
	db := database.CreateTestSqliteDB(t).DB
	customerRepo := customers.NewCustomerRepo(logger, db)
	accountRepo := NewRepo(logger, db)
	organization := "test"

	customer := &client.Customer{
		CustomerID: base.ID(),
		FirstName:  "jane",
		LastName:   "doe",
		Type:       client.CUSTOMERTYPE_INDIVIDUAL,
	}
	err := customerRepo.CreateCustomer(customer, organization)
	require.NoError(t, err)

	// look for account that does not exist
	_, err = accountRepo.getCustomerAccount(customer.CustomerID, "xxx")
	require.Error(t, err)

	// initial read, find no accounts
	accounts, err := accountRepo.getAccounts(customer.CustomerID, organization)
	if len(accounts) != 0 || err != nil {
		t.Fatalf("got accounts=%#v error=%v", accounts, err)
	}

	// create account
	acct, err := accountRepo.CreateCustomerAccount(customer.CustomerID, userID, &CreateAccountRequest{
		AccountNumber: "123",
		RoutingNumber: "987654320",
		Type:          client.ACCOUNTTYPE_CHECKING,
	})
	if err != nil {
		t.Fatal(err)
	}

	// read after creating
	accounts, err = accountRepo.getAccounts(customer.CustomerID, organization)
	if len(accounts) != 1 || err != nil {
		t.Fatalf("got accounts=%#v error=%v", accounts, err)
	}
	if accounts[0].AccountID != acct.AccountID {
		t.Errorf("accounts[0].AccountID=%s acct.AccountID=%s", accounts[0].AccountID, acct.AccountID)
	}

	// delete, expect no accounts
	if err := accountRepo.deactivateCustomerAccount(acct.AccountID); err != nil {
		t.Fatal(err)
	}
	accounts, err = accountRepo.getAccounts(customer.CustomerID, organization)
	if len(accounts) != 0 || err != nil {
		t.Fatalf("got accounts=%#v error=%v", accounts, err)
	}
}

func TestRepository__getEncryptedAccountNumber(t *testing.T) {
	customerID, userID := base.ID(), base.ID()
	organization := "test-org"
	repo := setupTestAccountRepository(t)

	keeper := secrets.TestStringKeeper(t)

	// create account
	req := &CreateAccountRequest{
		AccountNumber: "123",
		RoutingNumber: "987654320",
		Type:          client.ACCOUNTTYPE_CHECKING,
	}
	if err := req.disfigure(keeper); err != nil {
		t.Fatal(err)
	}
	acct, err := repo.CreateCustomerAccount(customerID, userID, req)
	if err != nil {
		t.Fatal(err)
	}

	// create customer
	customerRepo := customers.NewCustomerRepo(log.NewNopLogger(), repo.db.DB)
	cust := &client.Customer{
		CustomerID: customerID,
		FirstName:  "jane",
		LastName:   "doe",
		Type:       client.CUSTOMERTYPE_INDIVIDUAL,
	}
	custErr := customerRepo.CreateCustomer(cust, organization)
	if custErr != nil {
		t.Fatal(custErr)
	}

	// read encrypted account number
	encrypted, err := repo.getEncryptedAccountNumber(organization, customerID, acct.AccountID)
	if err != nil {
		t.Fatal(err)
	}
	if encrypted == "" {
		t.Error("missing encrypted account number")
	}
}

func TestRepository__updateAccountStatus(t *testing.T) {
	customerID, userID := base.ID(), base.ID()
	repo := setupTestAccountRepository(t)

	keeper := secrets.TestStringKeeper(t)

	// create account
	req := &CreateAccountRequest{
		AccountNumber: "123",
		RoutingNumber: "987654320",
		Type:          client.ACCOUNTTYPE_CHECKING,
	}
	if err := req.disfigure(keeper); err != nil {
		t.Fatal(err)
	}
	acct, err := repo.CreateCustomerAccount(customerID, userID, req)
	if err != nil {
		t.Fatal(err)
	}

	// update status
	if err := repo.updateAccountStatus(acct.AccountID, client.ACCOUNTSTATUS_VALIDATED); err != nil {
		t.Fatal(err)
	}

	// check status after update
	acct, err = repo.getCustomerAccount(customerID, acct.AccountID)
	if err != nil {
		t.Fatal(err)
	}
	if acct.Status != client.ACCOUNTSTATUS_VALIDATED {
		t.Errorf("unexpected status: %s", acct.Status)
	}
}

func TestRepositoryUnique(t *testing.T) {
	keeper := secrets.TestStringKeeper(t)

	check := func(t *testing.T, repo *sqlAccountRepository) {
		customerID, userID := base.ID(), base.ID()
		req := &CreateAccountRequest{
			AccountNumber: "156421",
			RoutingNumber: "123456780",
			Type:          client.ACCOUNTTYPE_SAVINGS,
		}
		if err := req.disfigure(keeper); err != nil {
			t.Fatal(err)
		}

		// first write should pass
		if _, err := repo.CreateCustomerAccount(customerID, userID, req); err != nil {
			t.Fatal(err)
		}
		// second write should fail
		if _, err := repo.CreateCustomerAccount(customerID, userID, req); err != nil {
			if !database.UniqueViolation(err) {
				t.Fatalf("unexpected error: %v", err)
			}
		}
	}

	// SQLite tests
	sqliteDB := database.CreateTestSqliteDB(t)
	defer sqliteDB.Close()
	check(t, NewRepo(log.NewNopLogger(), sqliteDB.DB))

	// MySQL tests
	mysqlDB := database.CreateTestMySQLDB(t)
	defer mysqlDB.Close()
	check(t, NewRepo(log.NewNopLogger(), mysqlDB.DB))
}
