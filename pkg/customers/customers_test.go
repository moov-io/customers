// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package customers

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/moov-io/customers/internal/database"

	"github.com/moov-io/base"

	"github.com/moov-io/customers/pkg/client"

	"github.com/gorilla/mux"
	"github.com/moov-io/base/log"
)

var _ CustomerRepository = (*testCustomerRepository)(nil)

type testCustomerRepository struct {
	err          error
	customer     *client.Customer
	searchResult *client.OfacSearch

	createdCustomer   *client.Customer
	updatedStatus     client.CustomerStatus
	savedSearchResult *client.OfacSearch
}

func (r *testCustomerRepository) GetCustomer(customerID string) (*client.Customer, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.customer, nil
}

func (r *testCustomerRepository) CreateCustomer(c *client.Customer, organization string) error {
	r.createdCustomer = c
	return r.err
}

func (r *testCustomerRepository) deleteCustomer(customerID string) error {
	r.customer = nil
	return r.err
}

func (r *testCustomerRepository) updateCustomer(c *client.Customer, organization string) error {
	r.customer = c
	return r.err
}

func (r *testCustomerRepository) updateCustomerStatus(customerID string, status client.CustomerStatus, comment string) error {
	r.updatedStatus = status
	return r.err
}

func (r *testCustomerRepository) searchCustomers(params SearchParams) ([]*client.Customer, error) {
	if r.err != nil {
		return nil, r.err
	}
	if r.customer != nil {
		return []*client.Customer{r.customer}, nil
	}
	return nil, nil
}

func (r *testCustomerRepository) replaceCustomerMetadata(customerID string, metadata map[string]string) error {
	return r.err
}

func (r *testCustomerRepository) addCustomerAddress(customerID string, address address) error {
	return r.err
}

func (r *testCustomerRepository) updateCustomerAddress(customerID, addressID string, req updateCustomerAddressRequest) error {
	return r.err
}

func (r *testCustomerRepository) deleteCustomerAddress(customerID string, addressID string) error {
	return r.err
}

func (r *testCustomerRepository) getLatestCustomerOFACSearch(customerID string) (*client.OfacSearch, error) {
	if r.err != nil {
		return nil, r.err
	}
	if r.savedSearchResult != nil {
		return r.savedSearchResult, nil
	}
	return r.searchResult, nil
}

func (r *testCustomerRepository) saveCustomerOFACSearch(customerID string, result client.OfacSearch) error {
	r.savedSearchResult = &result
	return r.err
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
	}).asCustomer(testCustomerSSNStorage(t))
	if err := repo.CreateCustomer(cust, "organization"); err != nil {
		t.Fatal(err)
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", fmt.Sprintf("/customers/%s", cust.CustomerID), nil)
	req.Header.Set("x-organization", "test")
	req.Header.Set("x-request-id", "test")

	router := mux.NewRouter()
	AddCustomerRoutes(log.NewNopLogger(), router, repo, testCustomerSSNStorage(t), createTestOFACSearcher(nil, nil))
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("bogus status code: %d", w.Code)
	}

	var customer client.Customer
	if err := json.NewDecoder(w.Body).Decode(&customer); err != nil {
		t.Fatal(err)
	}
	if customer.CustomerID == "" {
		t.Error("empty Customer.ID")
	}
}

func TestCustomers__GetCustomerEmpty(t *testing.T) {
	repo := &testCustomerRepository{}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", fmt.Sprintf("/customers/%s", base.ID()), nil)
	req.Header.Set("x-organization", "test")
	req.Header.Set("x-request-id", "test")

	router := mux.NewRouter()
	AddCustomerRoutes(log.NewNopLogger(), router, repo, testCustomerSSNStorage(t), createTestOFACSearcher(nil, nil))
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusNotFound {
		t.Errorf("bogus status code: %d", w.Code)
	}
}

func TestCustomers__DeleteCustomer(t *testing.T) {
	repo := createTestCustomerRepository(t)
	defer repo.close()

	cr := &customerRequest{
		FirstName: "Jane",
		LastName:  "Doe",
		Email:     "jane@example.com",
	}
	customer, _, _ := cr.asCustomer(testCustomerSSNStorage(t))
	require.NoError(t,
		repo.CreateCustomer(customer, "organization"),
	)

	got, err := repo.GetCustomer(customer.CustomerID)
	require.NoError(t, err)
	require.NotNil(t, got)

	router := mux.NewRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", fmt.Sprintf("/customers/%s", customer.CustomerID), nil)

	AddCustomerRoutes(log.NewNopLogger(), router, repo, testCustomerSSNStorage(t), createTestOFACSearcher(nil, nil))
	router.ServeHTTP(w, req)
	w.Flush()

	require.Equal(t, http.StatusNoContent, w.Code)
	require.Empty(t, w.Body)
	got, err = repo.GetCustomer(customer.CustomerID)
	require.Error(t, err)
	require.Nil(t, got)
}

func TestCustomerRepository__CreateCustomer(t *testing.T) {
	check := func(t *testing.T, repo *sqlCustomerRepository) {
		cust, _, _ := (customerRequest{
			FirstName: "Jane",
			LastName:  "Doe",
			Email:     "jane@example.com",
		}).asCustomer(testCustomerSSNStorage(t))
		if err := repo.CreateCustomer(cust, "organization"); err != nil {
			t.Fatal(err)
		}

		cust, err := repo.GetCustomer(cust.CustomerID)
		if err != nil {
			t.Fatal(err)
		}
		if cust == nil {
			t.Error("got nil Customer")
		}
	}

	// SQLite tests
	sqliteDB := database.CreateTestSqliteDB(t)
	defer sqliteDB.Close()
	check(t, &sqlCustomerRepository{sqliteDB.DB, log.NewNopLogger()})

	// MySQL tests
	mysqlDB := database.CreateTestMySQLDB(t)
	defer mysqlDB.Close()
	check(t, &sqlCustomerRepository{mysqlDB.DB, log.NewNopLogger()})
}

func TestCustomers__searchCustomersError(t *testing.T) {
	repo := &testCustomerRepository{err: errors.New("bad error")}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/customers/foo", nil)
	req.Header.Set("x-organization", "test")
	req.Header.Set("x-request-id", "test")

	router := mux.NewRouter()
	AddCustomerRoutes(log.NewNopLogger(), router, repo, testCustomerSSNStorage(t), createTestOFACSearcher(nil, nil))
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusBadRequest {
		t.Errorf("bogus status code: %d", w.Code)
	}
}

func TestCustomers__customerRequest(t *testing.T) {
	req := &customerRequest{Type: client.CUSTOMERTYPE_INDIVIDUAL}
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
		Type:       "primary",
	})
	if err := req.validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// asCustomer
	cust, _, _ := req.asCustomer(testCustomerSSNStorage(t))
	if cust.CustomerID == "" {
		t.Errorf("empty Customer: %#v", cust)
	}
	if len(cust.Phones) != 1 {
		t.Errorf("cust.Phones: %#v", cust.Phones)
	}
	if len(cust.Addresses) != 1 {
		t.Errorf("cust.Addresses: %#v", cust.Addresses)
	}
}

func TestCustomers__addressValidate(t *testing.T) {
	add := address{Type: "primary"}

	add.State = "IA"
	if err := add.validate(); err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	add.State = "ZZ"
	if err := add.validate(); err == nil {
		t.Error("expected error")
	}
}

func TestCustomers__CreateCustomer(t *testing.T) {
	w := httptest.NewRecorder()
	phone := `{"number": "555.555.5555", "type": "mobile"}`
	address := `{"type": "primary", "address1": "123 1st St", "city": "Denver", "state": "CO", "postalCode": "12345", "country": "USA"}`
	body := fmt.Sprintf(`{"firstName": "jane", "lastName": "doe", "email": "jane@example.com", "birthDate": "1991-04-01", "ssn": "123456789", "type": "individual", "phones": [%s], "addresses": [%s]}`, phone, address)
	req := httptest.NewRequest("POST", "/customers", strings.NewReader(body))
	req.Header.Set("x-organization", "test")
	req.Header.Set("x-request-id", "test")

	repo := createTestCustomerRepository(t)
	defer repo.close()

	customerSSNStorage := testCustomerSSNStorage(t)

	router := mux.NewRouter()
	AddCustomerRoutes(log.NewNopLogger(), router, repo, customerSSNStorage, createTestOFACSearcher(nil, nil))
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("bogus status code: %d: %v", w.Code, w.Body.String())
	}

	var cust client.Customer
	if err := json.NewDecoder(w.Body).Decode(&cust); err != nil {
		t.Fatal(err)
	}
	if cust.CustomerID == "" {
		t.Error("empty Customer.CustomerID")
	}

	// sad path
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/customers", strings.NewReader("null"))
	req.Header.Set("x-organization", "test")
	req.Header.Set("x-request-id", "test")
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("bogus HTTP status code: %d", w.Code)
	}

	// customerSSNStorage sad path
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", "/customers", strings.NewReader(body))
	req.Header.Set("x-organization", "test")

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

func TestCustomers__updateCustomer(t *testing.T) {
	repo := createTestCustomerRepository(t)
	defer repo.close()

	createReq := &customerRequest{
		FirstName: "Jane",
		LastName:  "Doe",
		Type:      "individual",
		BirthDate: "1999-01-01",
		Email:     "jane@example.com",
		SSN:       "123456789",
		Phones: []phone{
			{
				Number: "123.456.7890",
				Type:   "cell",
			},
		},
		Addresses: []address{
			{
				Address1:   "123 1st st",
				City:       "fake city",
				State:      "CA",
				PostalCode: "90210",
				Country:    "US",
				Type:       "primary",
			},
		},
	}
	customer, _, _ := createReq.asCustomer(testCustomerSSNStorage(t))
	require.NoError(t,
		repo.CreateCustomer(customer, "organization"),
	)

	_, err := repo.GetCustomer(customer.CustomerID)
	require.NoError(t, err)

	router := mux.NewRouter()
	w := httptest.NewRecorder()

	updateReq := *createReq
	updateReq.FirstName = "Jim"
	updateReq.LastName = "Smith"
	updateReq.Email = "jim@google.com"
	updateReq.Phones = []phone{
		{
			Number: "555.555.5555",
			Type:   "cell",
		},
	}
	updateReq.Addresses = []address{
		{
			Address1:   "555 5th st",
			City:       "real city",
			State:      "CA",
			PostalCode: "90210",
			Country:    "US",
			Type:       "primary",
		},
	}
	payload, err := json.Marshal(&updateReq)
	require.NoError(t, err)

	req := httptest.NewRequest("PUT", fmt.Sprintf("/customers/%s", customer.CustomerID), bytes.NewReader(payload))
	req.Header.Set("x-organization", "test")
	req.Header.Set("x-request-id", "test")
	AddCustomerRoutes(log.NewNopLogger(), router, repo, testCustomerSSNStorage(t), createTestOFACSearcher(nil, nil))
	router.ServeHTTP(w, req)
	w.Flush()
	require.Equal(t, http.StatusOK, w.Code)

	var got *client.Customer
	require.NoError(t, json.NewDecoder(w.Body).Decode(&got))
	want, _, _ := updateReq.asCustomer(testCustomerSSNStorage(t))
	require.NoError(t, err)

	want.CustomerID = got.CustomerID
	want.Addresses[0].AddressID = got.Addresses[0].AddressID
	want.BirthDate = got.BirthDate
	want.Status = got.Status
	want.CreatedAt = got.CreatedAt
	want.LastModified = got.LastModified
	require.Equal(t, want, got)

	/* Error when settings two addresses as primary */
	updateReq.Addresses = []address{
		{
			Address1:   "555 5th st",
			City:       "real city",
			State:      "CA",
			PostalCode: "90210",
			Country:    "US",
			Type:       "primary",
		},
		{
			Address1:   "444 4th st",
			City:       "real city",
			State:      "CA",
			PostalCode: "90210",
			Country:    "US",
			Type:       "primary",
		},
	}
	payload, err = json.Marshal(&updateReq)
	require.NoError(t, err)
	req = req.Clone(context.Background())
	req.Body = ioutil.NopCloser(bytes.NewReader(payload))
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)
	var errResp struct {
		ErrorMsg string `json:"error"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &errResp))
	require.Contains(t, errResp.ErrorMsg, ErrAddressTypeDuplicate.Error())
}

func createTestCustomerRepository(t *testing.T) *sqlCustomerRepository {
	t.Helper()

	db := database.CreateTestSqliteDB(t)
	return &sqlCustomerRepository{db.DB, log.NewNopLogger()}
}

func TestCustomers__repository(t *testing.T) {
	repo := createTestCustomerRepository(t)
	defer repo.close()

	cust, err := repo.GetCustomer(base.ID())
	if err == nil {
		t.Errorf("expected not found error: %v", err)
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
	}).asCustomer(testCustomerSSNStorage(t))
	if err := repo.CreateCustomer(cust, "organization"); err != nil {
		t.Error(err)
	}
	if cust == nil {
		t.Fatal("nil Customer")
	}
	if len(cust.Phones) != 1 || len(cust.Addresses) != 1 {
		t.Errorf("len(cust.Phones)=%d and len(cust.Addresses)=%d", len(cust.Phones), len(cust.Addresses))
	}

	// read
	cust, err = repo.GetCustomer(cust.CustomerID)
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

func TestCustomerRepository__delete(t *testing.T) {
	repo := createTestCustomerRepository(t)
	defer repo.close()
	type customer struct {
		*client.Customer
		deleted bool
	}
	customers := make([]*customer, 10)

	for i := 0; i < len(customers); i++ {
		cr := &customerRequest{
			FirstName: "Jane",
			LastName:  "Doe",
			Email:     "jane@example.com",
		}
		cust, _, _ := cr.asCustomer(testCustomerSSNStorage(t))
		require.NoError(t,
			repo.CreateCustomer(cust, "organization"),
		)

		customers[i] = &customer{
			Customer: cust,
		}
	}

	// mark customers to be deleted
	indexesToDelete := []int{1, 2, 5, 8}
	for _, idx := range indexesToDelete {
		require.Less(t, idx, len(customers))
		customers[idx].deleted = true
		require.NoError(t,
			repo.deleteCustomer(customers[idx].CustomerID),
		)
	}

	deletedCustomerIds := make(map[string]bool)
	// query all customers that have been marked as deleted
	query := `select customer_id from customers where deleted_at is not null;`
	stmt, err := repo.db.Prepare(query)
	require.NoError(t, err)

	rows, err := stmt.Query()
	require.NoError(t, err)

	for rows.Next() {
		var customerID *string
		require.NoError(t, rows.Scan(&customerID))
		deletedCustomerIds[*customerID] = true
	}

	for _, cust := range customers {
		_, ok := deletedCustomerIds[cust.CustomerID]
		require.Equal(t, cust.deleted, ok)
	}
}

func TestCustomerRepository__updateCustomer(t *testing.T) {
	repo := createTestCustomerRepository(t)
	defer repo.close()

	createReq := customerRequest{
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
	}
	newCust, _, _ := createReq.asCustomer(testCustomerSSNStorage(t))
	err := repo.CreateCustomer(newCust, "organization")
	require.NoError(t, err)

	updateReq := customerRequest{
		CustomerID: newCust.CustomerID,
		FirstName:  "Jim",
		LastName:   "Smith",
		Email:      "jim@google.com",
		Phones: []phone{
			{
				Number: "555.555.5555",
			},
		},
		Addresses: []address{
			{
				Address1: "555 5th st",
				City:     "real city",
			},
		},
	}

	updatedCust, _, _ := updateReq.asCustomer(testCustomerSSNStorage(t))
	err = repo.updateCustomer(updatedCust, "organization")
	require.NoError(t, err)

	require.Equal(t, newCust.CustomerID, updatedCust.CustomerID)
	require.Equal(t, updateReq.FirstName, updatedCust.FirstName)
	require.Equal(t, updateReq.LastName, updatedCust.LastName)
	require.Equal(t, updateReq.Email, updatedCust.Email)
	require.Equal(t, updateReq.Phones[0].Number, updatedCust.Phones[0].Number)
	require.Equal(t, updateReq.Addresses[0].Address1, updatedCust.Addresses[0].Address1)
	require.Equal(t, updateReq.Addresses[0].City, updatedCust.Addresses[0].City)
}

func TestCustomerRepository__updateCustomerStatus(t *testing.T) {
	repo := createTestCustomerRepository(t)
	defer repo.close()

	cust, _, _ := (customerRequest{
		FirstName: "Jane",
		LastName:  "Doe",
		Email:     "jane@example.com",
	}).asCustomer(testCustomerSSNStorage(t))
	if err := repo.CreateCustomer(cust, "organization"); err != nil {
		t.Fatal(err)
	}

	// update status
	if err := repo.updateCustomerStatus(cust.CustomerID, client.CUSTOMERSTATUS_VERIFIED, "test comment"); err != nil {
		t.Fatal(err)
	}

	// read the status back
	customer, err := repo.GetCustomer(cust.CustomerID)
	if err != nil {
		t.Fatal(err)
	}
	if customer.Status != client.CUSTOMERSTATUS_VERIFIED {
		t.Errorf("unexpected status: %s", customer.Status)
	}
}

func TestCustomersRepository__addCustomerAddress(t *testing.T) {
	repo := createTestCustomerRepository(t)
	defer repo.close()

	cust, _, _ := (customerRequest{
		FirstName: "Jane",
		LastName:  "Doe",
		Email:     "jane@example.com",
	}).asCustomer(testCustomerSSNStorage(t))
	if err := repo.CreateCustomer(cust, "organization"); err != nil {
		t.Fatal(err)
	}

	// add an address
	if err := repo.addCustomerAddress(cust.CustomerID, address{
		Address1:   "123 1st st",
		City:       "fake city",
		State:      "CA",
		PostalCode: "90210",
		Country:    "US",
	},
	); err != nil {
		t.Fatal(err)
	}

	// re-read
	cust, err := repo.GetCustomer(cust.CustomerID)
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

func TestCustomers__replaceCustomerMetadata(t *testing.T) {
	repo := createTestCustomerRepository(t)
	defer repo.close()

	cust, _, _ := (customerRequest{
		FirstName: "Jane",
		LastName:  "Doe",
		Email:     "jane@example.com",
	}).asCustomer(testCustomerSSNStorage(t))
	if err := repo.CreateCustomer(cust, "organization"); err != nil {
		t.Fatal(err)
	}

	body := strings.NewReader(`{ "metadata": { "key": "bar"} }`)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("PUT", fmt.Sprintf("/customers/%s/metadata", cust.CustomerID), body)
	req.Header.Set("x-organization", "test")
	req.Header.Set("x-request-id", "test")

	router := mux.NewRouter()
	AddCustomerRoutes(log.NewNopLogger(), router, repo, testCustomerSSNStorage(t), createTestOFACSearcher(nil, nil))
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
	req.Header.Set("x-organization", "test")
	req.Header.Set("x-request-id", "test")

	router2 := mux.NewRouter()
	AddCustomerRoutes(log.NewNopLogger(), router2, repo2, testCustomerSSNStorage(t), createTestOFACSearcher(nil, nil))
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
	}).asCustomer(testCustomerSSNStorage(t))
	if err := repo.CreateCustomer(cust, "organization"); err != nil {
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
	req := httptest.NewRequest("PUT", fmt.Sprintf("/customers/%s/metadata", cust.CustomerID), &buf)
	req.Header.Set("x-organization", "test")
	req.Header.Set("x-request-id", "test")

	router := mux.NewRouter()
	AddCustomerRoutes(log.NewNopLogger(), router, repo, testCustomerSSNStorage(t), createTestOFACSearcher(nil, nil))
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusBadRequest {
		t.Errorf("bogus status code: %d", w.Code)
	}

	// invalid JSON
	w = httptest.NewRecorder()
	req = httptest.NewRequest("PUT", fmt.Sprintf("/customers/%s/metadata", cust.CustomerID), strings.NewReader("{invalid-json"))
	req.Header.Set("x-organization", "test")
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
	req.Header.Set("x-organization", "test")
	req.Header.Set("x-request-id", "test")

	router := mux.NewRouter()
	AddCustomerRoutes(log.NewNopLogger(), router, repo, testCustomerSSNStorage(t), createTestOFACSearcher(nil, nil))
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusBadRequest {
		t.Errorf("bogus status code: %d", w.Code)
	}
}

func TestCustomerRepository__metadata(t *testing.T) {
	repo := createTestCustomerRepository(t)
	defer repo.close()

	getMetadata := func(customerID string) map[string]string {
		meta, err := repo.getMetadata([]string{customerID})
		if err != nil {
			t.Fatal(err)
		}
		return meta[customerID].Metadata
	}
	customerID := base.ID()

	meta := getMetadata(customerID)
	if len(meta) != 0 {
		t.Errorf("unknown metadata: %#v", meta)
	}

	// replace
	if err := repo.replaceCustomerMetadata(customerID, map[string]string{"key": "bar"}); err != nil {
		t.Fatal(err)
	}

	meta = getMetadata(customerID)
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

func TestCustomerRepository__OFAC(t *testing.T) {
	repo := createTestCustomerRepository(t)
	defer repo.close()

	customerID := base.ID()

	res, err := repo.getLatestCustomerOFACSearch(customerID)
	if err != nil {
		t.Fatal(err)
	}
	if res != nil {
		t.Errorf("unexpected ofac search: %#v", res)
	}

	// save a record and read it back
	if err := repo.saveCustomerOFACSearch(customerID, client.OfacSearch{EntityID: "14141"}); err != nil {
		t.Fatal(err)
	}
	res, err = repo.getLatestCustomerOFACSearch(customerID)
	if err != nil {
		t.Fatal(err)
	}
	if res == nil || res.EntityID != "14141" {
		t.Errorf("ofac search: %#v", res)
	}

	// save another and get it back
	if err := repo.saveCustomerOFACSearch(customerID, client.OfacSearch{EntityID: "777121"}); err != nil {
		t.Fatal(err)
	}
	res, err = repo.getLatestCustomerOFACSearch(customerID)
	if err != nil {
		t.Fatal(err)
	}
	if res == nil || res.EntityID != "777121" {
		t.Errorf("ofac search: %#v", res)
	}
}

func mockCustomerRequest() customerRequest {
	c := customerRequest{}
	c.FirstName = "John"
	c.LastName = "Doe"
	c.Email = "johndoe@example.net"
	c.SSN = "123456789"
	c.Type = "individual"

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
		if !strings.Contains(err.Error(), ": metadata") {
			t.Fatal("Expected metadata error")
		}
	}
}

func TestCustomers__minimumFields(t *testing.T) {
	w := httptest.NewRecorder()
	body := `{"firstName": "jane", "lastName": "doe", "type": "individual"}`
	req := httptest.NewRequest("POST", "/customers", strings.NewReader(body))
	req.Header.Set("x-organization", "test")
	req.Header.Set("x-request-id", "test")

	repo := createTestCustomerRepository(t)
	defer repo.close()

	customerSSNStorage := testCustomerSSNStorage(t)

	router := mux.NewRouter()
	AddCustomerRoutes(log.NewNopLogger(), router, repo, customerSSNStorage, createTestOFACSearcher(nil, nil))
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("bogus status code: %d: %v", w.Code, w.Body.String())
	}
}

func TestCustomers__BadReq(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", "/customers", strings.NewReader("®"))
	req.Header.Set("x-organization", "test")
	req.Header.Set("x-request-id", "test")

	repo := createTestCustomerRepository(t)
	defer repo.close()

	customerSSNStorage := testCustomerSSNStorage(t)

	router := mux.NewRouter()
	AddCustomerRoutes(log.NewNopLogger(), router, nil, customerSSNStorage, createTestOFACSearcher(nil, nil))
	router.ServeHTTP(w, req)
	w.Flush()

	if !strings.Contains(w.Body.String(), "invalid character") {
		t.Errorf("Expected SSN error received %s", w.Body.String())
	}
}
