// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package database

import (
	"errors"
	"testing"

	"github.com/moov-io/base/log"
)

func TestMySQL__basic(t *testing.T) {
	db := CreateTestMySQLDB(t)
	defer db.Close()

	if err := db.DB.Ping(); err != nil {
		t.Fatal(err)
	}

	// create a phony MySQL
	m := mysqlConnection(log.NewNopLogger(), "user", "pass", "127.0.0.1:3006", "db")

	conn, err := m.Connect()
	defer func() {
		if conn != nil {
			conn.Close()
		}
	}()
	if conn != nil || err == nil {
		t.Fatalf("conn=%#v expected error", conn)
	}
}

func TestMySQLUniqueViolation(t *testing.T) {
	err := errors.New(`problem upserting customer="7d676c65eccd48090ff238a0d5e35eb6126c23f2", first_name="John", middle_name="B", last_name="Doe": Error 1062: Duplicate entry '282f6ffcd9ba5b029afbf2b739ee826e22d9df3b' for key 'PRIMARY'`)
	if !UniqueViolation(err) {
		t.Error("should have matched unique violation")
	}
}
