// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package database

import (
	"testing"

	"github.com/go-kit/kit/log"
)

func TestNewDataBase__MYSQL(t *testing.T) {
	_, err := New(log.NewNopLogger(), "Mysql")
	if err == nil {
		t.Error("Error: expected access denied for use")
	}
}

func TestNewDataBase__sqlite(t *testing.T) {
	db, err := New(log.NewNopLogger(), "sqlite")
	if err != nil {
		t.Error("Error: expected access denied for use")
	}
	if db == nil {
		t.Error("Error: expected a database connection")
	}
}

func TestNewDataBase__sqliteDefault(t *testing.T) {
	db, err := New(log.NewNopLogger(), "")
	if err != nil {
		t.Error("Error: expected access denied for use")
	}
	if db == nil {
		t.Error("Error: expected a database connection")
	}
}

func TestNewDataBase__invalid_type(t *testing.T) {
	_, err := New(log.NewNopLogger(), "db2")
	if err == nil {
		t.Errorf("Error: expected unknown database type %s", "db2")
	}
}
