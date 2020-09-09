// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"database/sql"
	"encoding/json"
	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	client "github.com/moov-io/customers/client"
	"github.com/moov-io/customers/internal/database"
	"net/http"
	"net/http/httptest"
	"testing"
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
		Count: 100,
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
		Count: 25,
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

func TestGet20MostRecentlyCreatedCustomersByDefault(t *testing.T) {
	scope := Setup(t)
	scope.CreateCustomers(100)
	customers, _ := scope.GetCustomers("")
	scope.assert.Equal(20, len(customers))
}

func TestGet10MostRecentlyCreatedCustomersByDefault(t *testing.T) {
	scope := Setup(t)
	scope.CreateCustomers(10)
	customers, _ := scope.GetCustomers("")
	scope.assert.Equal(10, len(customers))
}

func TestGet50MostRecentlyCreatedCustomersWhenSpecifyingLimit(t *testing.T) {
	scope := Setup(t)
	scope.CreateCustomers(100)
	customers, _ := scope.GetCustomers("?count=50")
	scope.assert.Equal(50, len(customers))
}

func TestGet100MostRecentlyCreatedCustomersWhenSpecifyingMoreThanAvailable(t *testing.T) {
	scope := Setup(t)
	scope.CreateCustomers(100)
	customers, _ := scope.GetCustomers("?count=120")
	scope.assert.Equal(100, len(customers))
}

func TestGetCustomersWithVerifiedStatus(t *testing.T) {
	// Create two customers. 1 with Unknown STATUS and 1 with Verified
	scope := Setup(t)
	scope.CreateCustomers(2)
	customers, _ := scope.GetCustomers("?count=120")
	scope.assert.Equal(2, len(customers))
	for i := 0; i < len(customers); i++ {
		if i % 2 == 0 {
			// update status
			if err := scope.customerRepo.updateCustomerStatus(customers[i].CustomerID, client.VERIFIED, "test comment"); err != nil {
				print(err)
			}
		}
	}

	// Should have 1 Verified Status
	customers, _ = scope.GetCustomers("?status=Verified&count=20")
	scope.assert.Equal(1, len(customers))
	for i := 0; i < len(customers); i++ {
		scope.assert.Equal("Verified", string(customers[i].Status))
	}

	// Should have 1 Unknown Status
	customers, _ = scope.GetCustomers("?status=Unknown&count=20")
	scope.assert.Equal(1, len(customers))
	for i := 0; i < len(customers); i++ {
		scope.assert.Equal("Unknown", string(customers[i].Status))
	}
}

func TestGetCustomersByName(t *testing.T) {
	scope := Setup(t)
	_ = scope.CreateCustomer("Jane", "Doe", "jane.doe@gmail.com")
	_ = scope.CreateCustomer("John", "Doe", "jane.doe@gmail.com")

	customers, _ := scope.GetCustomers("?query=jane")
	scope.assert.Equal(1, len(customers))

	customers, _ = scope.GetCustomers("?query=jane+doe")
	scope.assert.Equal(1, len(customers))

	customers, _ = scope.GetCustomers("?query=john")
	scope.assert.Equal(1, len(customers))

	customers, _ = scope.GetCustomers("?query=john+doe")
	scope.assert.Equal(1, len(customers))

	customers, _ = scope.GetCustomers("?query=doe")
	scope.assert.Equal(2, len(customers))

	customers, _ = scope.GetCustomers("?query=jim+doe")
	scope.assert.Equal(0, len(customers))
}

func TestGetCustomersByEmail(t *testing.T) {
	scope := Setup(t)
	_ = scope.CreateCustomer("Jane", "Doe", "jane.doe@gmail.com")
	_ = scope.CreateCustomer("John", "Doe", "john.doe@gmail.com")

	customers, _ := scope.GetCustomers("?email=jane.doe@gmail.com")
	scope.assert.Equal(1, len(customers))

	customers, _ = scope.GetCustomers("?email=john.doe@gmail.com")
	scope.assert.Equal(1, len(customers))

	customers, _ = scope.GetCustomers("?email=jim.doe@gmail.com")
	scope.assert.Equal(0, len(customers))
}

func TestGetCustomersByNameAndEmail(t *testing.T) {
	scope := Setup(t)
	_ = scope.CreateCustomer("Jane", "Doe", "jane.doe@gmail.com")
	_ = scope.CreateCustomer("John", "Doe", "john.doe@gmail.com")

	customers, _ := scope.GetCustomers("?query=jane+doe&email=jane.doe@gmail.com")
	scope.assert.Equal(1, len(customers))

	customers, _ = scope.GetCustomers("?query=jane&email=jane.doe@gmail.com")
	scope.assert.Equal(1, len(customers))

	customers, _ = scope.GetCustomers("?query=john+doe&email=john.doe@gmail.com")
	scope.assert.Equal(1, len(customers))

	customers, _ = scope.GetCustomers("?query=john&email=john.doe@gmail.com")
	scope.assert.Equal(1, len(customers))

	customers, _ = scope.GetCustomers("?query=jane+doe&email=jim.doe@gmail.com")
	scope.assert.Equal(0, len(customers))

	customers, _ = scope.GetCustomers("?query=jane&email=jim.doe@gmail.com")
	scope.assert.Equal(0, len(customers))

	customers, _ = scope.GetCustomers("?query=jim&email=jane.doe@gmail.com")
	scope.assert.Equal(0, len(customers))
}

func TestGetCustomersByType(t *testing.T) {
	scope := Setup(t)
	_ = scope.CreateCustomer("Jane", "Doe", "jane.doe@gmail.com")
	_ = scope.CreateCustomer("John", "Doe", "john.doe@gmail.com")

	customers, _ := scope.GetCustomers("?query=jane+doe&email=jane.doe@gmail.com")
	scope.assert.Equal(1, len(customers))

	customers, _ = scope.GetCustomers("?query=jane&email=jane.doe@gmail.com")
	scope.assert.Equal(1, len(customers))

	customers, _ = scope.GetCustomers("?query=john+doe&email=john.doe@gmail.com")
	scope.assert.Equal(1, len(customers))

	customers, _ = scope.GetCustomers("?query=john&email=john.doe@gmail.com")
	scope.assert.Equal(1, len(customers))

	customers, _ = scope.GetCustomers("?query=jane+doe&email=jim.doe@gmail.com")
	scope.assert.Equal(0, len(customers))

	customers, _ = scope.GetCustomers("?query=jane&email=jim.doe@gmail.com")
	scope.assert.Equal(0, len(customers))

	customers, _ = scope.GetCustomers("?query=jim&email=jane.doe@gmail.com")
	scope.assert.Equal(0, len(customers))
}

func TestGetCustomersUsingPaging(t *testing.T) {
	scope := Setup(t)
	_ = scope.CreateCustomers(30)

	// Get first page of 10
	customers, _ := scope.GetCustomers("?skip=0&count=10")
	scope.assert.Equal(10, len(customers))

	// Get second page of 10
	customers, _ = scope.GetCustomers("?skip=10&count=10")
	scope.assert.Equal(10, len(customers))

	// Get third page of 10
	customers, _ = scope.GetCustomers("?skip=20&count=10")
	scope.assert.Equal(10, len(customers))

	// Should be no 4th page of 10
	customers, _ = scope.GetCustomers("?skip=30&count=10")
	scope.assert.Equal(0, len(customers))
}