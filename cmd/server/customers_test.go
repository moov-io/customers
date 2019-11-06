// Copyright 2018 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/moov-io/customers/internal/database"

	"github.com/moov-io/base"
	"github.com/moov-io/customers"
	client "github.com/moov-io/customers/client"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

type testCustomerRepository struct {
	err              error
	customer         *client.Customer
	ofacSearchResult *ofacSearchResult

	createdCustomer       *client.Customer
	updatedStatus         customers.Status
	savedOFACSearchResult *ofacSearchResult
}

func (r *testCustomerRepository) getCustomer(customerID string) (*client.Customer, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.customer, nil
}

func (r *testCustomerRepository) createCustomer(c *client.Customer) error {
	r.createdCustomer = c
	return r.err
}

func (r *testCustomerRepository) updateCustomerStatus(customerID string, status customers.Status, comment string) error {
	r.updatedStatus = status
	return r.err
}

func (r *testCustomerRepository) getCustomerMetadata(customerID string) (map[string]string, error) {
	out := make(map[string]string)
	return out, r.err
}

func (r *testCustomerRepository) replaceCustomerMetadata(customerID string, metadata map[string]string) error {
	return r.err
}

func (r *testCustomerRepository) addCustomerAddress(customerID string, address address) error {
	return r.err
}

func (r *testCustomerRepository) updateCustomerAddress(customerID, addressID string, _type string, validated bool) error {
	return r.err
}

func (r *testCustomerRepository) getLatestCustomerOFACSearch(customerID string) (*ofacSearchResult, error) {
	if r.err != nil {
		return nil, r.err
	}
	if r.savedOFACSearchResult != nil {
		return r.savedOFACSearchResult, nil
	}
	return r.ofacSearchResult, nil
}

func (r *testCustomerRepository) saveCustomerOFACSearch(customerID string, result ofacSearchResult) error {
	r.savedOFACSearchResult = &result
	return r.err
}

func TestCustomers__getCustomerID(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/ping", nil)

	if id := getCustomerID(w, req); id != "" {
		t.Errorf("unexpected id: %v", id)
	}
}

func TestCustomers__formatCustomerName(t *testing.T) {
	if out := formatCustomerName(nil); out != "" {
		t.Errorf("got %q", out)
	}

	cases := []struct {
		output, expected string
	}{
		{formatCustomerName(&client.Customer{FirstName: "Jane"}), "Jane"},
		{formatCustomerName(&client.Customer{FirstName: "Jane", LastName: "Doe"}), "Jane Doe"},
		{formatCustomerName(&client.Customer{FirstName: "Jane", MiddleName: " B ", LastName: "Doe"}), "Jane B Doe"},
		{formatCustomerName(&client.Customer{FirstName: " John", MiddleName: "M", LastName: "Doe", Suffix: "Jr"}), "John M Doe Jr"},
		{formatCustomerName(&client.Customer{FirstName: "John ", MiddleName: "M", LastName: " Doe ", Suffix: "Jr "}), "John M Doe Jr"},
		{formatCustomerName(&client.Customer{FirstName: "John ", MiddleName: "M", Suffix: "Jr "}), "John M Jr"},
		{formatCustomerName(&client.Customer{FirstName: "John ", Suffix: "Jr "}), "John Jr"},
		{formatCustomerName(&client.Customer{MiddleName: "M", LastName: " Doe ", Suffix: "Jr "}), "M Doe Jr"},
		{formatCustomerName(&client.Customer{MiddleName: "M", LastName: " Doe "}), "M Doe"},
		{formatCustomerName(&client.Customer{LastName: " Doe "}), "Doe"},
	}
	for i := range cases {
		if cases[i].output != cases[i].expected {
			t.Errorf("got %q expected %q", cases[i].output, cases[i].expected)
		}
	}
}

func TestCustomers__GetCustomer(t *testing.T) {
	repo := createTestCustomerRepository(t)
	defer repo.close()

	cust, _, _ := (customerRequest{
		FirstName: "Jane",
		LastName:  "Doe",
		Email:     "jane@example.com",
	}).asCustomer(testCustomerSSNStorage)
	if err := repo.createCustomer(cust); err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", fmt.Sprintf("/customers/%s", cust.ID), nil)
	req.Header.Set("x-user-id", "test")
	req.Header.Set("x-request-id", "test")

	router := mux.NewRouter()
	addCustomerRoutes(log.NewNopLogger(), router, repo, testCustomerSSNStorage, createTestOFACSearcher(nil, nil))
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("bogus status code: %d", w.Code)
	}

	var customer client.Customer
	if err := json.NewDecoder(w.Body).Decode(&customer); err != nil {
		t.Fatal(err)
	}
	if customer.ID == "" {
		t.Error("empty Customer.ID")
	}
}

func TestCustomers__GetCustomersError(t *testing.T) {
	repo := &testCustomerRepository{err: errors.New("bad error")}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/customers/foo", nil)
	req.Header.Set("x-user-id", "test")
	req.Header.Set("x-request-id", "test")

	router := mux.NewRouter()
	addCustomerRoutes(log.NewNopLogger(), router, repo, testCustomerSSNStorage, createTestOFACSearcher(nil, nil))
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusBadRequest {
		t.Errorf("bogus status code: %d", w.Code)
	}
}

func TestCustomers__customerRequest(t *testing.T) {
	req := &customerRequest{}
	if err := req.validate(); err == nil {
		t.Error("expected error")
	}
	req.FirstName = "jane"
	if err := req.validate(); err == nil {
		t.Error("expected error")
	}

	req.LastName = "doe"
	if err := req.validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	req.Email = "jane.doe@example.com"
	if err := req.validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	req.Phones = append(req.Phones, phone{
		Number: "123.456.7890",
		Type:   "Checking",
	})
	if err := req.validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	req.Addresses = append(req.Addresses, address{
		Address1:   "123 1st st",
		City:       "fake city",
		State:      "CA",
		PostalCode: "90210",
		Country:    "US",
	})
	if err := req.validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// asCustomer
	cust, _, _ := req.asCustomer(testCustomerSSNStorage)
	if cust.ID == "" {
		t.Errorf("empty Customer: %#v", cust)
	}
	if len(cust.Phones) != 1 {
		t.Errorf("cust.Phones: %#v", cust.Phones)
	}
	if len(cust.Addresses) != 1 {
		t.Errorf("cust.Addresses: %#v", cust.Addresses)
	}
}

func TestCustomers__createCustomer(t *testing.T) {
	w := httptest.NewRecorder()
	phone := `{"number": "555.555.5555", "type": "mobile"}`
	address := `{"type": "home", "address1": "123 1st St", "city": "Denver", "state": "CO", "postalCode": "12345", "country": "USA"}`
	body := fmt.Sprintf(`{"firstName": "jane", "lastName": "doe", "email": "jane@example.com", "ssn": "123456789", "phones": [%s], "addresses": [%s]}`, phone, address)
	req := httptest.NewRequest("POST", "/customers", strings.NewReader(body))
	req.Header.Set("x-user-id", "test")
	req.Header.Set("x-request-id", "test")

	repo := createTestCustomerRepository(t)
	defer repo.close()

	customerSSNStorage := testCustomerSSNStorage

	router := mux.NewRouter()
	addCustomerRoutes(log.NewNopLogger(), router, repo, customerSSNStorage, createTestOFACSearcher(nil, nil))
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("bogus status code: %d: %v", w.Code, w.Body.String())
	}

	var cust client.Customer
	if err := json.NewDecoder(w.Body).Decode(&cust); err != nil {
		t.Fatal(err)
	}
	if cust.ID == "" {
		t.Error("empty Customer.ID")
	}

	// sad path
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/customers", strings.NewReader("null"))
	req.Header.Set("x-user-id", "test")
	req.Header.Set("x-request-id", "test")
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("bogus HTTP status code: %d", w.Code)
	}

	// customerSSNStorage sad path
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/customers", strings.NewReader(body))
	req.Header.Set("x-user-id", "test")

	if r, ok := customerSSNStorage.repo.(*testCustomerSSNRepository); !ok {
		t.Fatalf("got %T", customerSSNStorage.repo)
	} else {
		r.err = errors.New("bad error")
	}
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusBadRequest {
		t.Errorf("bogus HTTP status code: %d: %v", w.Code, w.Body.String())
	}
	if s := w.Body.String(); !strings.Contains(s, "saveCustomerSSN: ") {
		t.Errorf("unexpected error: %v", s)
	}
}

func createTestCustomerRepository(t *testing.T) *sqlCustomerRepository {
	t.Helper()

	db := database.CreateTestSqliteDB(t)
	return &sqlCustomerRepository{db.DB, log.NewNopLogger()}
}

func TestCustomers__repository(t *testing.T) {
	repo := createTestCustomerRepository(t)
	defer repo.close()

	cust, err := repo.getCustomer(base.ID())
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if cust != nil {
		t.Error("expected no Customer")
	}

	// write
	cust, _, _ = (customerRequest{
		FirstName: "Jane",
		LastName:  "Doe",
		Email:     "jane@example.com",
		Phones: []phone{
			{
				Number: "123.456.7890",
				Type:   "Checking",
			},
		},
		Addresses: []address{
			{
				Address1:   "123 1st st",
				City:       "fake city",
				State:      "CA",
				PostalCode: "90210",
				Country:    "US",
			},
		},
	}).asCustomer(testCustomerSSNStorage)
	if err := repo.createCustomer(cust); err != nil {
		t.Error(err)
	}
	if cust == nil {
		t.Fatal("nil Customer")
	}
	if len(cust.Phones) != 1 || len(cust.Addresses) != 1 {
		t.Errorf("len(cust.Phones)=%d and len(cust.Addresses)=%d", len(cust.Phones), len(cust.Addresses))
	}

	// read
	cust, err = repo.getCustomer(cust.ID)
	if err != nil {
		t.Error(err)
	}
	if cust == nil {
		t.Fatal("nil Customer")
	}
	if len(cust.Phones) != 1 || len(cust.Addresses) != 1 {
		t.Errorf("len(cust.Phones)=%d and len(cust.Addresses)=%d", len(cust.Phones), len(cust.Addresses))
	}
}

func TestCustomerRepository__updateCustomerStatus(t *testing.T) {
	repo := createTestCustomerRepository(t)
	defer repo.close()

	cust, _, _ := (customerRequest{
		FirstName: "Jane",
		LastName:  "Doe",
		Email:     "jane@example.com",
	}).asCustomer(testCustomerSSNStorage)
	if err := repo.createCustomer(cust); err != nil {
		t.Fatal(err)
	}

	// update status
	if err := repo.updateCustomerStatus(cust.ID, customers.KYC, "test comment"); err != nil {
		t.Fatal(err)
	}

	// read the status back
	customer, err := repo.getCustomer(cust.ID)
	if err != nil {
		t.Fatal(err)
	}
	if customer.Status != customers.KYC.String() {
		t.Errorf("unexpected status: %s", customer.Status)
	}
}

func TestCustomers__replaceCustomerMetadata(t *testing.T) {
	repo := createTestCustomerRepository(t)
	defer repo.close()

	cust, _, _ := (customerRequest{
		FirstName: "Jane",
		LastName:  "Doe",
		Email:     "jane@example.com",
	}).asCustomer(testCustomerSSNStorage)
	if err := repo.createCustomer(cust); err != nil {
		t.Fatal(err)
	}

	body := strings.NewReader(`{ "metadata": { "key": "bar"} }`)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", fmt.Sprintf("/customers/%s/metadata", cust.ID), body)
	req.Header.Set("x-user-id", "test")
	req.Header.Set("x-request-id", "test")

	router := mux.NewRouter()
	addCustomerRoutes(log.NewNopLogger(), router, repo, testCustomerSSNStorage, createTestOFACSearcher(nil, nil))
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("bogus status code: %d", w.Code)
	}

	var customer client.Customer
	if err := json.NewDecoder(w.Body).Decode(&customer); err != nil {
		t.Fatal(err)
	}
	if customer.Metadata["key"] != "bar" {
		t.Errorf("unknown Customer metadata: %#v", customer.Metadata)
	}

	// sad path
	repo2 := &testCustomerRepository{err: errors.New("bad error")}

	w = httptest.NewRecorder()
	req = httptest.NewRequest("PUT", "/customers/foo/metadata", nil)
	req.Header.Set("x-user-id", "test")
	req.Header.Set("x-request-id", "test")

	router2 := mux.NewRouter()
	addCustomerRoutes(log.NewNopLogger(), router2, repo2, testCustomerSSNStorage, createTestOFACSearcher(nil, nil))
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusBadRequest {
		t.Errorf("bogus status code: %d", w.Code)
	}
}

func TestCustomers__replaceCustomerMetadataInvalid(t *testing.T) {
	repo := createTestCustomerRepository(t)
	defer repo.close()

	cust, _, _ := (customerRequest{
		FirstName: "Jane",
		LastName:  "Doe",
		Email:     "jane@example.com",
	}).asCustomer(testCustomerSSNStorage)
	if err := repo.createCustomer(cust); err != nil {
		t.Fatal(err)
	}

	r := replaceMetadataRequest{
		Metadata: map[string]string{"key": strings.Repeat("a", 100000)},
	}
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(r); err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", fmt.Sprintf("/customers/%s/metadata", cust.ID), &buf)
	req.Header.Set("x-user-id", "test")
	req.Header.Set("x-request-id", "test")

	router := mux.NewRouter()
	addCustomerRoutes(log.NewNopLogger(), router, repo, testCustomerSSNStorage, createTestOFACSearcher(nil, nil))
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusBadRequest {
		t.Errorf("bogus status code: %d", w.Code)
	}

	// invalid JSON
	w = httptest.NewRecorder()
	req = httptest.NewRequest("PUT", fmt.Sprintf("/customers/%s/metadata", cust.ID), strings.NewReader("{invalid-json"))
	req.Header.Set("x-user-id", "test")
	req.Header.Set("x-request-id", "test")

	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusBadRequest {
		t.Errorf("bogus status code: %d", w.Code)
	}
}

func TestCustomers__replaceCustomerMetadataError(t *testing.T) {
	repo := &testCustomerRepository{err: errors.New("bad error")}

	body := strings.NewReader(`{ "metadata": { "key": "bar"} }`)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", "/customers/foo/metadata", body)
	req.Header.Set("x-user-id", "test")
	req.Header.Set("x-request-id", "test")

	router := mux.NewRouter()
	addCustomerRoutes(log.NewNopLogger(), router, repo, testCustomerSSNStorage, createTestOFACSearcher(nil, nil))
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusBadRequest {
		t.Errorf("bogus status code: %d", w.Code)
	}
}

func TestCustomerRepository__metadata(t *testing.T) {
	repo := createTestCustomerRepository(t)
	defer repo.close()

	customerID := base.ID()

	meta, err := repo.getCustomerMetadata(customerID)
	if err != nil {
		t.Fatal(err)
	}
	if len(meta) != 0 {
		t.Errorf("unknown metadata: %#v", meta)
	}

	// replace
	if err := repo.replaceCustomerMetadata(customerID, map[string]string{"key": "bar"}); err != nil {
		t.Fatal(err)
	}
	meta, err = repo.getCustomerMetadata(customerID)
	if err != nil {
		t.Fatal(err)
	}
	if len(meta) != 1 || meta["key"] != "bar" {
		t.Errorf("unknown metadata: %#v", meta)
	}
}

func TestCustomers__validateMetadata(t *testing.T) {
	meta := make(map[string]string)
	if err := validateMetadata(meta); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	meta["key"] = "foo"
	if err := validateMetadata(meta); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	meta["bar"] = strings.Repeat("b", 100000) // invalid
	if err := validateMetadata(meta); err == nil {
		t.Error("expected error")
	}

	meta["bar"] = "baz"         // valid again
	for i := 0; i < 1000; i++ { // add too many keys
		meta[fmt.Sprintf("key-%d", i)] = fmt.Sprintf("%d", i)
	}
	if err := validateMetadata(meta); err == nil {
		t.Error("expected error")
	}
}

func TestCustomers__addCustomerAddress(t *testing.T) {
	repo := &testCustomerRepository{
		customer: &client.Customer{
			ID: base.ID(),
		},
		err: errors.New("bad error"),
	}

	address := `{"type": "home", "address1": "123 1st St", "city": "Denver", "state": "CO", "postalCode": "12345", "country": "USA"}`

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/customers/foo/address", strings.NewReader(address))
	req.Header.Set("x-user-id", "test")
	req.Header.Set("x-request-id", "test")

	router := mux.NewRouter()
	addCustomerRoutes(log.NewNopLogger(), router, repo, testCustomerSSNStorage, createTestOFACSearcher(nil, nil))
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusBadRequest {
		t.Errorf("bogus HTTP status: %d", w.Code)
	}

	// remove error and retry
	repo.err = nil

	req = httptest.NewRequest("POST", "/customers/foo/address", strings.NewReader(address))
	req.Header.Set("x-user-id", "test")
	req.Header.Set("x-request-id", "test")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("bogus HTTP status: %d: %v", w.Code, w.Body.String())
	}
}

func TestCustomersRepository__addCustomerAddress(t *testing.T) {
	repo := createTestCustomerRepository(t)
	defer repo.close()

	cust, _, _ := (customerRequest{
		FirstName: "Jane",
		LastName:  "Doe",
		Email:     "jane@example.com",
	}).asCustomer(testCustomerSSNStorage)
	if err := repo.createCustomer(cust); err != nil {
		t.Fatal(err)
	}

	// add an Address
	if err := repo.addCustomerAddress(cust.ID, address{
		Address1:   "123 1st st",
		City:       "fake city",
		State:      "CA",
		PostalCode: "90210",
		Country:    "US",
	}); err != nil {
		t.Fatal(err)
	}

	// re-read
	cust, err := repo.getCustomer(cust.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(cust.Addresses) != 1 {
		t.Errorf("got %d Addresses", len(cust.Addresses))
	}
	if cust.Addresses[0].Address1 != "123 1st st" {
		t.Errorf("cust.Addresses[0].Address1=%s", cust.Addresses[0].Address1)
	}
}

func TestCustomerRepository__OFAC(t *testing.T) {
	repo := createTestCustomerRepository(t)
	defer repo.close()

	customerID := base.ID()

	res, err := repo.getLatestCustomerOFACSearch(customerID)
	if err != nil {
		t.Fatal(err)
	}
	if res != nil {
		t.Errorf("unexpected ofacSearchResult: %#v", res)
	}

	// save a record and read it back
	if err := repo.saveCustomerOFACSearch(customerID, ofacSearchResult{EntityId: "14141"}); err != nil {
		t.Fatal(err)
	}
	res, err = repo.getLatestCustomerOFACSearch(customerID)
	if err != nil {
		t.Fatal(err)
	}
	if res == nil || res.EntityId != "14141" {
		t.Errorf("ofacSearchResult: %#v", res)
	}

	// save another and get it back
	if err := repo.saveCustomerOFACSearch(customerID, ofacSearchResult{EntityId: "777121"}); err != nil {
		t.Fatal(err)
	}
	res, err = repo.getLatestCustomerOFACSearch(customerID)
	if err != nil {
		t.Fatal(err)
	}
	if res == nil || res.EntityId != "777121" {
		t.Errorf("ofacSearchResult: %#v", res)
	}
}

func mockCustomerRequest() customerRequest {
	c := customerRequest{}
	c.FirstName = "John"
	c.LastName = "Doe"
	c.Email = "johndoe@example.net"
	c.SSN = "123456789"

	p := phone{}
	p.Number = "123-456-7892"
	p.Type = "cell"
	c.Phones = append(c.Phones, p)

	a := address{}
	a.Address1 = "Any Street"
	a.Address2 = ""
	a.City = "Any City"
	a.Country = "USA"
	a.PostalCode = "19456"
	a.State = "MA"
	c.Addresses = append(c.Addresses, a)
	return c
}

// TestCustomers__MetaDataValidate validates customer meta data
func TestCustomers__MetaDataValidate(t *testing.T) {
	c := mockCustomerRequest()
	m := make(map[string]string)
	for i := 0; i < 101; i++ {
		s := strconv.Itoa(i)

		m[s] = s
	}
	c.Metadata = m

	if err := c.validate(); err != nil {
		if err != nil {
			if !strings.Contains(err.Error(), ": metadata") {
				t.Fatal("Expected metadata error")
			}
		}
	}
}

func TestCustomers__minimumFields(t *testing.T) {
	w := httptest.NewRecorder()
	body := `{"firstName": "jane", "lastName": "doe"}`
	req := httptest.NewRequest("POST", "/customers", strings.NewReader(body))
	req.Header.Set("x-user-id", "test")
	req.Header.Set("x-request-id", "test")

	repo := createTestCustomerRepository(t)
	defer repo.close()

	customerSSNStorage := testCustomerSSNStorage

	router := mux.NewRouter()
	addCustomerRoutes(log.NewNopLogger(), router, repo, customerSSNStorage, createTestOFACSearcher(nil, nil))
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("bogus status code: %d: %v", w.Code, w.Body.String())
	}
}

func TestCustomers__BadReq(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/customers", strings.NewReader("®"))
	req.Header.Set("x-user-id", "test")
	req.Header.Set("x-request-id", "test")

	repo := createTestCustomerRepository(t)
	defer repo.close()

	customerSSNStorage := testCustomerSSNStorage

	router := mux.NewRouter()
	addCustomerRoutes(log.NewNopLogger(), router, nil, customerSSNStorage, createTestOFACSearcher(nil, nil))
	router.ServeHTTP(w, req)
	w.Flush()

	if !strings.Contains(w.Body.String(), "invalid character") {
		t.Errorf("Expected SSN error received %s", w.Body.String())
	}
}
