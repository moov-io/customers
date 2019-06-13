// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	"github.com/go-kit/kit/log"
	_ "github.com/mattn/go-sqlite3"
)

var (
	// migrations holds all our SQL migrations to be done (in order)
	migrations = []string{
		// Customer tables
		`create table if not exists customers(customer_id primary key, first_name, middle_name, last_name, nick_name, suffix, birthdate datetime, status, email, created_at datetime, last_modified datetime, deleted_at datetime);`,
		`create table if not exists customers_phones(customer_id, number, valid, type, unique (customer_id, number) on conflict abort);`,
		`create table if not exists customers_addresses(address_id primary key, customer_id, type, address1, address2, city, state, postal_code, country, validated, active, unique (customer_id, address1) on conflict abort);`,
		`create table if not exists customer_metadata(customer_id, key, value, unique(key, value));`,
		`create table if not exists customer_status_updates(customer_id, future_status, comment, changed_at datetime);`,
		`create table if not exists customer_ofac_searches(customer_id, entity_id, sdn_name, sdn_type, match, created_at datetime);`,

		// Document tables
		`create table if not exists documents(document_id primary key, customer_id, type, content_type, uploaded_at datetime);`,
	}
)

func getSqlitePath() string {
	path := os.Getenv("SQLITE_DB_PATH")
	if path == "" || strings.Contains(path, "..") {
		// set default if empty or trying to escape
		// don't filepath.ABS to avoid full-fs reads
		path = "customers.db"
	}
	return path
}

func createSqliteConnection(logger log.Logger, path string) (*sql.DB, error) {
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		err = fmt.Errorf("problem opening sqlite3 file %s: %v", path, err)
		if logger != nil {
			logger.Log("sqlite", err)
		}
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("problem with Ping against *sql.DB %s: %v", path, err)
	}
	return db, nil
}

// migrate runs our database migrations (defined at the top of this file)
// over a sqlite database it creates first.
// To configure where on disk the sqlite db is set SQLITE_DB_PATH.
//
// You use db like any other database/sql driver.
//
// https://github.com/mattn/go-sqlite3/blob/master/_example/simple/simple.go
// https://astaxie.gitbooks.io/build-web-application-with-golang/en/05.3.html
func migrate(logger log.Logger, db *sql.DB) error {
	if logger != nil {
		logger.Log("sqlite", "starting database migrations")
	}
	for i := range migrations {
		row := migrations[i]
		res, err := db.Exec(row)
		if err != nil {
			return fmt.Errorf("migration #%d [%s...] had problem: %v", i, row[:40], err)
		}
		n, err := res.RowsAffected()
		if err == nil && logger != nil {
			logger.Log("sqlite", fmt.Sprintf("migration #%d [%s...] changed %d rows", i, row[:40], n))
		}
	}
	if logger != nil {
		logger.Log("sqlite", "finished migrations")
	}
	return nil
}
