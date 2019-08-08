// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package database

import (
	"errors"
	"strings"
	"testing"

	"github.com/go-kit/kit/log"
)

func TestSQLite__basic(t *testing.T) {
	db := CreateTestSqliteDB(t)
	defer db.Close()

	if err := db.DB.Ping(); err != nil {
		t.Fatal(err)
	}

	// error case
	s := sqliteConnection(log.NewNopLogger(), "/tmp/path/doesnt/exist")

	conn, err := s.Connect()

	defer conn.Close()

	if err := conn.Ping(); err == nil {
		t.Fatal("expected error")
	}
	if err == nil {
		t.Fatalf("conn=%#v expected error", conn)
	}
}

func TestSqlite__getSqlitePath(t *testing.T) {
	if v := getSqlitePath(); v != "customers.db" {
		t.Errorf("got %s", v)
	}
}

func TestSqliteUniqueViolation(t *testing.T) {
	err := errors.New(`problem upserting customer="7d676c65eccd48090ff238a0d5e35eb6126c23f2", first_name="John", middle_name="B", last_name="Doe": : UNIQUE constraint failed: customer.customer_id`)
	if !UniqueViolation(err) {
		t.Error("should have matched unique violation")
	}
}

// TestSQLite__SERR validates sqliteConnection doesn't contain an error
func TestSQLite__SERR(t *testing.T) {
	// error case
	s := sqliteConnection(log.NewNopLogger(), "/tmp/path/doesnt/exist")

	s.err = errors.New(": invalid")
	_, err := s.Connect()

	if err != nil {
		if !strings.Contains(err.Error(), ": invalid") {
			t.Fatal("Expected invalid")
		}
	}
}
