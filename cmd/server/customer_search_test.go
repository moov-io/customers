// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	client "github.com/moov-io/customers/client"
	"github.com/moov-io/customers/internal/database"
)

func TestCustomersSearchRouter(t *testing.T) {
	repo := createTestCustomerRepository(t)
	defer repo.close()

	router := mux.NewRouter()
	addCustomerRoutes(log.NewNopLogger(), router, repo, nil, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/customers?query=jane+doe", nil)
	router.ServeHTTP(w, req)

	// verify with zero results we don't return null
	if body := w.Body.String(); body != "[]\n" {
		t.Errorf("got %q", body)
	}
	if w.Code != http.StatusOK {
		t.Errorf("bogus HTTP status: %v", w.Code)
	}

	// write a customer we can search for
	cust, _, _ := (customerRequest{
		FirstName: "Jane",
		LastName:  "Doe",
		Email:     "jane@example.com",
	}).asCustomer(testCustomerSSNStorage(t))
	if err := repo.createCustomer(cust); err != nil {
		t.Error(err)
	}

	// find a customer from their partial name
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/customers?query=jane", nil)
	router.ServeHTTP(w, req)

	var resp []*client.Customer
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if len(resp) != 1 || resp[0].CustomerID != cust.CustomerID {
		t.Errorf("unexpected customers: %#v", resp)
	}

	// find a customer from full name
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/customers?query=jane+doe", nil)
	router.ServeHTTP(w, req)

	var resp2 []*client.Customer
	if err := json.NewDecoder(w.Body).Decode(&resp2); err != nil {
		t.Fatal(err)
	}
	if len(resp2) != 1 || resp2[0].CustomerID != cust.CustomerID {
		t.Errorf("unexpected customers: %#v", resp2)
	}
}

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
	if query != "select customer_id from customers where deleted_at is null and lower(first_name) || \" \" || lower(last_name) LIKE ? order by created_at asc limit ?;" {
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
	if query != "select customer_id from customers where deleted_at is null and lower(first_name) || \" \" || lower(last_name) LIKE ? and lower(email) like ? order by created_at asc limit ?;" {
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
