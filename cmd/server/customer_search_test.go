// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"database/sql"
	"testing"

	"github.com/moov-io/customers/internal/database"
)

func TestCustomerSearch__query(t *testing.T) {
	sqliteDB := database.CreateTestSqliteDB(t)
	defer sqliteDB.Close()

	mysqlDB := database.CreateTestMySQLDB(t)
	defer mysqlDB.Close()

	prepare := func(db *sql.DB, query string) error {
		stmt, err := db.Prepare(query)
		if err != nil {
			return err
		}
		return stmt.Close()
	}

	// Query search
	query, args := buildSearchQuery(searchParams{
		Query: "jane doe",
		Limit: 100,
	})
	if query != "select customer_id from customers where deleted_at is null and lower(first_name) || lower(last_name) LIKE ? order by created_at asc limit ?;" {
		t.Errorf("unexpected query: %q", query)
	}
	if err := prepare(sqliteDB.DB, query); err != nil {
		t.Errorf("sqlite: %v", err)
	}
	if err := prepare(mysqlDB.DB, query); err != nil {
		t.Errorf("mysql: %v", err)
	}
	if len(args) != 2 {
		t.Errorf("unexpected args: %#v", args)
	}

	// Eamil search
	query, args = buildSearchQuery(searchParams{
		Email: "jane.doe@moov.io",
	})
	if query != "select customer_id from customers where deleted_at is null and lower(email) like ? order by created_at asc limit ?;" {
		t.Errorf("unexpected query: %q", query)
	}
	if err := prepare(sqliteDB.DB, query); err != nil {
		t.Errorf("sqlite: %v", err)
	}
	if err := prepare(mysqlDB.DB, query); err != nil {
		t.Errorf("mysql: %v", err)
	}
	if len(args) != 2 {
		t.Errorf("unexpected args: %#v", args)
	}

	// Query and Eamil saerch
	query, args = buildSearchQuery(searchParams{
		Query: "jane doe",
		Email: "jane.doe@moov.io",
		Limit: 25,
	})
	if query != "select customer_id from customers where deleted_at is null and lower(first_name) || lower(last_name) LIKE ? and lower(email) like ? order by created_at asc limit ?;" {
		t.Errorf("unexpected query: %q", query)
	}
	if err := prepare(sqliteDB.DB, query); err != nil {
		t.Errorf("sqlite: %v", err)
	}
	if err := prepare(mysqlDB.DB, query); err != nil {
		t.Errorf("mysql: %v", err)
	}
	if len(args) != 3 {
		t.Errorf("unexpected args: %#v", args)
	}
}
