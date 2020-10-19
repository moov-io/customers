// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package customers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/moov-io/base/log"
	"github.com/stretchr/testify/require"

	"github.com/moov-io/base/database"
	"github.com/moov-io/customers/pkg/client"
)

func TestCustomersSearchRouter(t *testing.T) {
	db := database.CreateTestMySQLDB(t)
	repo := NewCustomerRepo(log.NewNopLogger(), db.DB)
	defer db.Close()

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
	if err := repo.CreateCustomer(cust, "organization"); err != nil {
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

func TestRepository__searchCustomers(t *testing.T) {
	logger := log.NewNopLogger()
	org := "test"
	createCustomer := func(repo CustomerRepository, i int) *client.Customer {
		var req = &customerRequest{
			FirstName: fmt.Sprintf("jane-%d", i),
			LastName:  fmt.Sprintf("doe-%d", i),
			Email:     fmt.Sprintf("jane-%d@moov.com", i),
			Phones: []phone{
				{
					Number: "555-555-5555",
					Type:   "primary",
				},
			},
			Addresses: []address{
				{
					Type:       "primary",
					Address1:   "123 Cool St.",
					City:       "San Francisco",
					State:      "CA",
					PostalCode: "94030",
					Country:    "US",
				},
			},
			Metadata: map[string]string{"key": "val"},
		}
		cust, _, _ := req.asCustomer(testCustomerSSNStorage(t))
		require.NoError(t, repo.CreateCustomer(cust, org))
		return cust
	}

	tests := []struct {
		desc string
		db   *sql.DB
	}{
		{
			desc: "sqlite",
			db:   database.CreateTestSqliteDB(t).DB,
		},
		{
			desc: "mysql",
			db:   database.CreateTestMySQLDB(t).DB,
		},
	}

	for _, tc := range tests {
		t.Run(tc.desc, func(t *testing.T) {
			defer tc.db.Close()
			repo := NewCustomerRepo(logger, tc.db)

			n := 20 // seed database with customers
			var customers []*client.Customer
			for i := 0; i < n; i++ {
				customers = append(customers, createCustomer(repo, i))
			}

			/* Search by Count */
			params := SearchParams{
				Organization: org,
				Count:        5,
			}
			got, err := repo.searchCustomers(params)
			require.NoError(t, err)
			require.Len(t, got, int(params.Count))

			/* Search by email */
			params = SearchParams{
				Organization: org,
				Count:        1,
				Email:        "JaNe-3@moov.com",
			}
			got, err = repo.searchCustomers(params)
			require.NoError(t, err)
			require.Equal(t, strings.ToLower(params.Email), got[0].Email)

			if tc.desc == "mysql" {
				/* Search by query */
				params = SearchParams{
					Organization: org,
					Count:        100,
					Query:        "jane-10",
				}
				got, err = repo.searchCustomers(params)
				require.NoError(t, err)
				require.Len(t, got, 1)
			}

			/* Search by customerIDs */
			wantIDs := make([]string, 3)
			for i := 0; i < len(wantIDs); i++ {
				wantIDs[i] = customers[i].CustomerID
			}

			got, err = repo.searchCustomers(SearchParams{
				Organization: org,
				CustomerIDs:  wantIDs,
				Count:        10,
			})
			require.NoError(t, err)
			require.Len(t, got, len(wantIDs))

			var gotIDs []string
			for _, c := range got {
				gotIDs = append(gotIDs, c.CustomerID)
			}

			require.ElementsMatch(t, wantIDs, gotIDs)
		})
	}
}

func TestCustomerSearchEmpty(t *testing.T) {
	scope := Setup(t)
	organization := "organization"
	customers, _ := scope.GetCustomers("", organization)
	if customers == nil {
		t.Fatalf("expected allocated array:\n  %T %#v", customers, customers)
	}
}

func TestGet20MostRecentlyCreatedCustomersByDefault(t *testing.T) {
	scope := Setup(t)
	organization := "organization"
	scope.CreateCustomers(100, client.CUSTOMERTYPE_INDIVIDUAL, organization)
	customers, _ := scope.GetCustomers("", organization)
	scope.assert.Equal(20, len(customers))
}

func TestGet10MostRecentlyCreatedCustomersByDefault(t *testing.T) {
	scope := Setup(t)
	organization := "organization"
	scope.CreateCustomers(10, client.CUSTOMERTYPE_INDIVIDUAL, organization)
	customers, _ := scope.GetCustomers("", organization)
	scope.assert.Equal(10, len(customers))
}

func TestGet50MostRecentlyCreatedCustomersWhenSpecifyingLimit(t *testing.T) {
	scope := Setup(t)
	organization := "organization"
	scope.CreateCustomers(100, client.CUSTOMERTYPE_INDIVIDUAL, organization)
	customers, _ := scope.GetCustomers("?count=50", organization)
	scope.assert.Equal(50, len(customers))
}

func TestGet100MostRecentlyCreatedCustomersWhenSpecifyingMoreThanAvailable(t *testing.T) {
	scope := Setup(t)
	organization := "organization"
	scope.CreateCustomers(100, client.CUSTOMERTYPE_INDIVIDUAL, organization)
	customers, _ := scope.GetCustomers("?count=120", organization)
	scope.assert.Equal(100, len(customers))
}

func TestSearchCustomersWithVerifiedStatus(t *testing.T) {
	// Create two customers. 1 with Unknown STATUS and 1 with Verified
	scope := Setup(t)
	organization := "organization"
	customer := scope.CreateCustomer("John", "Doe", organization, "john.doe@email.com", client.CUSTOMERTYPE_INDIVIDUAL)
	if err := scope.customerRepo.updateCustomerStatus(customer.CustomerID, client.CUSTOMERSTATUS_VERIFIED, "test comment"); err != nil {
		print(err)
	}
	scope.CreateCustomer("Jane", "Doe", organization, "jane.doe@email.com", client.CUSTOMERTYPE_INDIVIDUAL)

	// Should have 1 Verified Status
	verifiedCustomers, _ := scope.GetCustomers("?status=Verified&count=20", organization)
	scope.assert.Equal(1, len(verifiedCustomers))
	for i := 0; i < len(verifiedCustomers); i++ {
		scope.assert.Equal("Verified", string(verifiedCustomers[i].Status))
	}

	// Should have 1 Unknown Status
	unknownStatusCustomers, _ := scope.GetCustomers("?status=Unknown&count=20", organization)
	scope.assert.Equal(1, len(unknownStatusCustomers))
	for i := 0; i < len(unknownStatusCustomers); i++ {
		scope.assert.Equal("Unknown", string(unknownStatusCustomers[i].Status))
	}
}

func TestSearchCustomersByName(t *testing.T) {
	scope := Setup(t)
	scope.customerRepo = &sqlCustomerRepository{
		logger: log.NewNopLogger(),
		db:     database.CreateTestMySQLDB(t).DB,
	}
	organization := "organization"
	_ = scope.CreateCustomer("Jane", "Doe", organization, "jane.doe@gmail.com", client.CUSTOMERTYPE_INDIVIDUAL)
	_ = scope.CreateCustomer("John", "Doe", organization, "jane.doe@gmail.com", client.CUSTOMERTYPE_BUSINESS)

	customers, _ := scope.GetCustomers("?query=jane", organization)
	scope.assert.Equal(1, len(customers))

	customers, _ = scope.GetCustomers("?query=jane+doe", organization)
	scope.assert.Equal(1, len(customers))

	customers, _ = scope.GetCustomers("?query=john", organization)
	scope.assert.Equal(1, len(customers))

	customers, _ = scope.GetCustomers("?query=john+doe", organization)
	scope.assert.Equal(1, len(customers))

	customers, _ = scope.GetCustomers("?query=doe", organization)
	scope.assert.Equal(2, len(customers))

	customers, _ = scope.GetCustomers("?query=jim+doe", organization)
	scope.assert.Equal(0, len(customers))
}

func TestSearchCustomersByEmail(t *testing.T) {
	scope := Setup(t)
	organization := "organization"
	_ = scope.CreateCustomer("Jane", "Doe", organization, "jane.doe@gmail.com", client.CUSTOMERTYPE_INDIVIDUAL)
	_ = scope.CreateCustomer("John", "Doe", organization, "john.doe@gmail.com", client.CUSTOMERTYPE_BUSINESS)

	customers, _ := scope.GetCustomers("?email=jane.doe@gmail.com", organization)
	scope.assert.Equal(1, len(customers))

	customers, _ = scope.GetCustomers("?email=john.doe@gmail.com", organization)
	scope.assert.Equal(1, len(customers))

	customers, _ = scope.GetCustomers("?email=jim.doe@gmail.com", organization)
	scope.assert.Equal(0, len(customers))
}

func TestSearchCustomersByNameAndEmail(t *testing.T) {
	scope := Setup(t)
	scope.customerRepo = &sqlCustomerRepository{
		logger: log.NewNopLogger(),
		db:     database.CreateTestMySQLDB(t).DB,
	}
	organization := "organization"
	_ = scope.CreateCustomer("Jane", "Doe", organization, "jane.doe@gmail.com", client.CUSTOMERTYPE_INDIVIDUAL)
	_ = scope.CreateCustomer("John", "Doe", organization, "john.doe@gmail.com", client.CUSTOMERTYPE_BUSINESS)

	customers, _ := scope.GetCustomers("?query=jane+doe&email=jane.doe@gmail.com", organization)
	scope.assert.Equal(1, len(customers))

	customers, _ = scope.GetCustomers("?query=jane&email=jane.doe@gmail.com", organization)
	scope.assert.Equal(1, len(customers))

	customers, _ = scope.GetCustomers("?query=john+doe&email=john.doe@gmail.com", organization)
	scope.assert.Equal(1, len(customers))

	customers, _ = scope.GetCustomers("?query=john&email=john.doe@gmail.com", organization)
	scope.assert.Equal(1, len(customers))

	customers, _ = scope.GetCustomers("?query=jane+doe&email=jim.doe@gmail.com", organization)
	scope.assert.Equal(0, len(customers))

	customers, _ = scope.GetCustomers("?query=jane&email=jim.doe@gmail.com", organization)
	scope.assert.Equal(0, len(customers))

	customers, _ = scope.GetCustomers("?query=jim&email=jane.doe@gmail.com", organization)
	scope.assert.Equal(0, len(customers))
}

func TestSearchCustomersUsingPaging(t *testing.T) {
	scope := Setup(t)
	organization := "organization"
	_ = scope.CreateCustomers(30, client.CUSTOMERTYPE_INDIVIDUAL, organization)

	// Get first page of 10
	customers, _ := scope.GetCustomers("?skip=0&count=10", organization)
	scope.assert.Equal(10, len(customers))

	// Get second page of 10
	customers, _ = scope.GetCustomers("?skip=10&count=10", organization)
	scope.assert.Equal(10, len(customers))

	// Get third page of 10
	customers, _ = scope.GetCustomers("?skip=20&count=10", organization)
	scope.assert.Equal(10, len(customers))

	// Should be no 4th page of 10
	customers, _ = scope.GetCustomers("?skip=30&count=10", organization)
	scope.assert.Equal(0, len(customers))
}

func TestSearchCustomersUsingPagingFailure(t *testing.T) {
	scope := Setup(t)
	organization := "organization"
	_ = scope.CreateCustomers(30, client.CUSTOMERTYPE_INDIVIDUAL, organization)

	customers, _ := scope.GetCustomers("?skip=123abc", organization)
	scope.assert.Equal(0, len(customers))

	customers, _ = scope.GetCustomers("?count=123abc", organization)
	scope.assert.Equal(0, len(customers))

	customers, _ = scope.GetCustomers("?skip=123abc&count=123abc", organization)
	scope.assert.Equal(0, len(customers))
}

func TestSearchCustomersByType(t *testing.T) {
	scope := Setup(t)
	organization := "organization"
	_ = scope.CreateCustomers(10, client.CUSTOMERTYPE_INDIVIDUAL, organization)
	_ = scope.CreateCustomers(20, client.CUSTOMERTYPE_BUSINESS, organization)

	individualCustomers, _ := scope.GetCustomers("?type=individual", organization)
	scope.assert.Equal(10, len(individualCustomers))

	businessCustomers, _ := scope.GetCustomers("?type=business", organization)
	scope.assert.Equal(20, len(businessCustomers))
}
