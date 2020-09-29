// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package configuration

import (
	"database/sql"
	"testing"

	"github.com/moov-io/base"
	"github.com/moov-io/customers/internal/database"
	"github.com/moov-io/customers/pkg/client"
)

func TestRepository(t *testing.T) {
	t.Parallel()

	check := func(t *testing.T, repo *sqlRepo) {
		namespace := base.ID()

		cfg, err := repo.Get(namespace)
		if err != nil {
			t.Fatal(err)
		}
		if cfg.LegalEntity != "" || cfg.PrimaryAccount != "" {
			t.Errorf("unexpected legal entity: %q", cfg.LegalEntity)
			t.Errorf("unexpected primary account: %q", cfg.PrimaryAccount)
		}

		customerID, accountID := base.ID(), base.ID()
		writeCustomerAndAccount(t, repo.db, namespace, customerID, accountID)

		// write config
		cfg = &client.OrganizationConfiguration{
			LegalEntity:    customerID,
			PrimaryAccount: accountID,
		}
		if _, err := repo.Update(namespace, cfg); err != nil {
			t.Fatal(err)
		}

		// verify
		cfg, err = repo.Get(namespace)
		if err != nil {
			t.Fatal(err)
		}
		if cfg.LegalEntity == "" || cfg.PrimaryAccount == "" {
			t.Errorf("expected legal entity: %q", cfg.LegalEntity)
			t.Errorf("expected primary account: %q", cfg.PrimaryAccount)
		}
	}

	check(t, sqliteRepo(t))
}

func sqliteRepo(t *testing.T) *sqlRepo {
	db := database.CreateTestSqliteDB(t)
	t.Cleanup(func() {
		db.Close()
	})
	return &sqlRepo{db: db.DB}
}

func writeCustomerAndAccount(t *testing.T, db *sql.DB, namespace string, customerID, accountID string) {
	// TODO(adam): replace after customers/acconts Repository are moved to ./pkg/
	query := `insert into customers (customer_id, namespace, first_name, last_name) values (?, ?, ?, ?);`
	stmt, err := db.Prepare(query)
	if err != nil {
		t.Fatal(err)
	}
	defer stmt.Close()
	if _, err := stmt.Exec(customerID, namespace, "jane", "doe"); err != nil {
		t.Fatal(err)
	}

	// insert account
	query = `insert into accounts (account_id, customer_id, masked_account_number) values (?, ?, ?);`
	stmt, err = db.Prepare(query)
	if err != nil {
		t.Fatal(err)
	}
	defer stmt.Close()
	if _, err := stmt.Exec(accountID, customerID, "XXX456"); err != nil {
		t.Fatal(err)
	}
}
