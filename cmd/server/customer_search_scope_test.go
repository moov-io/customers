package main

import (
	"encoding/json"
	"github.com/go-kit/kit/log"
	fuzz "github.com/google/gofuzz"
	"github.com/gorilla/mux"
	"github.com/moov-io/customers/client"
	"github.com/stretchr/testify/require"
	"net/http/httptest"
	"testing"
)

type Scope struct {
	assert       *require.Assertions
	customerRepo *sqlCustomerRepository
	fuzzer       *fuzz.Fuzzer
	t            *testing.T
}

func Setup(t *testing.T) Scope {
	return Scope{
		assert:       require.New(t),
		customerRepo: createTestCustomerRepository(t),
		fuzzer:       fuzz.New(),
		t:            t,
	}
}

func (scope *Scope) GetCustomers(query string) ([]*client.Customer, error) {
	router := mux.NewRouter()
	addCustomerRoutes(log.NewNopLogger(), router, scope.customerRepo, nil, nil)
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/customers"+query, nil)
	router.ServeHTTP(w, req)

	var customers []*client.Customer
	if err := json.NewDecoder(w.Body).Decode(&customers); err != nil {
		return nil, err
	}
	return customers, nil
}

func (scope *Scope) CreateCustomers(count int) []client.Customer {
	var customers []client.Customer
	for i := 0; i < count; i++ {
		var firstName string
		var lastName string
		var email string
		scope.fuzzer.Fuzz(&firstName)
		scope.fuzzer.Fuzz(&lastName)
		scope.fuzzer.Fuzz(&email)
		customer := scope.CreateCustomer(firstName, lastName, email)
		customers = append(customers, customer)
	}
	return customers
}

func (scope *Scope) CreateCustomer(firstName, lastName, email string) client.Customer {
	cust, _, _ := (customerRequest{
		FirstName: firstName,
		LastName:  lastName,
		Email:     email,
	}).asCustomer(testCustomerSSNStorage(scope.t))
	if err := scope.customerRepo.createCustomer(cust); err != nil {
		scope.t.Error(err)
	}
	return *cust
}
