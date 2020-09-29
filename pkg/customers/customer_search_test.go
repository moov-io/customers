// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package customers

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"

	"github.com/moov-io/customers/internal/database"
	"github.com/moov-io/customers/pkg/client"
)

func TestCustomersSearchRouter(t *testing.T) {
	repo := createTestCustomerRepository(t)
	defer repo.close()

	router := mux.NewRouter()
	AddCustomerRoutes(log.NewNopLogger(), router, repo, nil, nil)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/customers?query=jane+doe", nil)
	req.Header.Set("X-Organization", "organization")
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
	if err := repo.createCustomer(cust, "organization"); err != nil {
		t.Error(err)
	}

	// find a customer from their partial name
	w = httptest.NewRecorder()
	req = httptest.NewRequest("GET", "/customers?query=jane", nil)
	req.Header.Set("X-Organization", "organization")
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
	req.Header.Set("X-Organization", "organization")
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
		Namespace: "foo",
		Query:     "jane doe",
		Count:     100,
	})
	if query != "select customer_id from customers where deleted_at is null and organization = ? and lower(first_name) || \" \" || lower(last_name) LIKE ? order by created_at desc limit ?;" {
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

	// Eamil search
	query, args = buildSearchQuery(searchParams{
		Namespace: "foo",
		Email:     "jane.doe@moov.io",
	})
	if query != "select customer_id from customers where deleted_at is null and organization = ? and lower(email) like ? order by created_at desc limit ?;" {
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

	// Query and Eamil saerch
	query, args = buildSearchQuery(searchParams{
		Namespace: "foo",
		Query:     "jane doe",
		Email:     "jane.doe@moov.io",
		Count:     25,
	})
	if query != "select customer_id from customers where deleted_at is null and organization = ? and lower(first_name) || \" \" || lower(last_name) LIKE ? and lower(email) like ? order by created_at desc limit ?;" {
		t.Errorf("unexpected query: %q", query)
	}
	if err := prepare(sqliteDB.DB, query); err != nil {
		t.Errorf("sqlite: %v", err)
	}
	if err := prepare(mysqlDB.DB, query); err != nil {
		t.Errorf("mysql: %v", err)
	}
	if len(args) != 4 {
		t.Errorf("unexpected args: %#v", args)
	}
}

func TestCustomerSearchEmpty(t *testing.T) {
	scope := Setup(t)
	customers, _ := scope.GetCustomers("")
	if customers == nil {
		t.Fatalf("expected allocated array:\n  %T %#v", customers, customers)
	}
}

func TestGet20MostRecentlyCreatedCustomersByDefault(t *testing.T) {
	scope := Setup(t)
	scope.CreateCustomers(100, client.INDIVIDUAL)
	customers, _ := scope.GetCustomers("")
	scope.assert.Equal(20, len(customers))
}

func TestGet10MostRecentlyCreatedCustomersByDefault(t *testing.T) {
	scope := Setup(t)
	scope.CreateCustomers(10, client.INDIVIDUAL)
	customers, _ := scope.GetCustomers("")
	scope.assert.Equal(10, len(customers))
}

func TestGet50MostRecentlyCreatedCustomersWhenSpecifyingLimit(t *testing.T) {
	scope := Setup(t)
	scope.CreateCustomers(100, client.INDIVIDUAL)
	customers, _ := scope.GetCustomers("?count=50")
	scope.assert.Equal(50, len(customers))
}

func TestGet100MostRecentlyCreatedCustomersWhenSpecifyingMoreThanAvailable(t *testing.T) {
	scope := Setup(t)
	scope.CreateCustomers(100, client.INDIVIDUAL)
	customers, _ := scope.GetCustomers("?count=120")
	scope.assert.Equal(100, len(customers))
}

func TestGetCustomersWithVerifiedStatus(t *testing.T) {
	// Create two customers. 1 with Unknown STATUS and 1 with Verified
	scope := Setup(t)
	customer := scope.CreateCustomer("John", "Doe", "john.doe@email.com", client.INDIVIDUAL)
	if err := scope.customerRepo.updateCustomerStatus(customer.CustomerID, client.VERIFIED, "test comment"); err != nil {
		print(err)
	}
	scope.CreateCustomer("Jane", "Doe", "jane.doe@email.com", client.INDIVIDUAL)

	// Should have 1 Verified Status
	verifiedCustomers, _ := scope.GetCustomers("?status=Verified&count=20")
	scope.assert.Equal(1, len(verifiedCustomers))
	for i := 0; i < len(verifiedCustomers); i++ {
		scope.assert.Equal("Verified", string(verifiedCustomers[i].Status))
	}

	// Should have 1 Unknown Status
	unknownStatusCustomers, _ := scope.GetCustomers("?status=Unknown&count=20")
	scope.assert.Equal(1, len(unknownStatusCustomers))
	for i := 0; i < len(unknownStatusCustomers); i++ {
		scope.assert.Equal("Unknown", string(unknownStatusCustomers[i].Status))
	}
}

func TestGetCustomersByName(t *testing.T) {
	scope := Setup(t)
	_ = scope.CreateCustomer("Jane", "Doe", "jane.doe@gmail.com", client.INDIVIDUAL)
	_ = scope.CreateCustomer("John", "Doe", "jane.doe@gmail.com", client.INDIVIDUAL)

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
	_ = scope.CreateCustomer("Jane", "Doe", "jane.doe@gmail.com", client.INDIVIDUAL)
	_ = scope.CreateCustomer("John", "Doe", "john.doe@gmail.com", client.INDIVIDUAL)

	customers, _ := scope.GetCustomers("?email=jane.doe@gmail.com")
	scope.assert.Equal(1, len(customers))

	customers, _ = scope.GetCustomers("?email=john.doe@gmail.com")
	scope.assert.Equal(1, len(customers))

	customers, _ = scope.GetCustomers("?email=jim.doe@gmail.com")
	scope.assert.Equal(0, len(customers))
}

func TestGetCustomersByNameAndEmail(t *testing.T) {
	scope := Setup(t)
	_ = scope.CreateCustomer("Jane", "Doe", "jane.doe@gmail.com", client.INDIVIDUAL)
	_ = scope.CreateCustomer("John", "Doe", "john.doe@gmail.com", client.INDIVIDUAL)

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
	_ = scope.CreateCustomers(30, client.INDIVIDUAL)

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

func TestGetCustomersUsingPagingFailure(t *testing.T) {
	scope := Setup(t)
	_ = scope.CreateCustomers(30, client.INDIVIDUAL)

	customers, _ := scope.GetCustomers("?skip=123abc")
	scope.assert.Equal(0, len(customers))

	customers, _ = scope.GetCustomers("?count=123abc")
	scope.assert.Equal(0, len(customers))

	customers, _ = scope.GetCustomers("?skip=123abc&count=123abc")
	scope.assert.Equal(0, len(customers))
}

func TestGetCustomersByType(t *testing.T) {
	scope := Setup(t)
	_ = scope.CreateCustomers(10, client.INDIVIDUAL)
	_ = scope.CreateCustomers(20, client.BUSINESS)

	individualCustomers, _ := scope.GetCustomers("?type=individual")
	scope.assert.Equal(10, len(individualCustomers))

	businessCustomers, _ := scope.GetCustomers("?type=business")
	scope.assert.Equal(20, len(businessCustomers))
}
