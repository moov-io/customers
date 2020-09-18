// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package configuration

import (
	"testing"

	"github.com/moov-io/base"
	"github.com/moov-io/customers/internal/database"
	"github.com/moov-io/customers/pkg/client"
)

func TestRepository(t *testing.T) {
	t.Parallel()

	check := func(t *testing.T, repo Repository) {
		namespace := base.ID()

		cfg, err := repo.Get(namespace)
		if err != nil {
			t.Fatal(err)
		}
		if cfg.LegalEntity != "" || cfg.PrimaryAccount != "" {
			t.Errorf("unexpected legal entity: %q", cfg.LegalEntity)
			t.Errorf("unexpected primary account: %q", cfg.PrimaryAccount)
		}

		// TODO(adam): write customer and account

		// write config
		cfg = &client.NamespaceConfiguration{
			LegalEntity:    base.ID(),
			PrimaryAccount: base.ID(),
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
	check(t, mysqlRepo(t))
}

func sqliteRepo(t *testing.T) Repository {
	db := database.CreateTestSqliteDB(t)
	repo := NewRepository(db.DB)
	t.Cleanup(func() {
		db.Close()
	})
	return repo
}

func mysqlRepo(t *testing.T) Repository {
	db := database.CreateTestMySQLDB(t)
	repo := NewRepository(db.DB)
	t.Cleanup(func() {
		db.Close()
	})
	return repo
}
