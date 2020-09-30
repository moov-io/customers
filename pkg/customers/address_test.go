package customers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"

	"github.com/moov-io/customers/pkg/client"
)

func TestCustomers__addCustomerAddress(t *testing.T) {
	db := createTestCustomerRepository(t)
	repo := NewCustomerRepo(log.NewNopLogger(), db.db)

	customerRequest := customerRequest{
		FirstName: "Jane",
		LastName:  "Doe",
	}
	cust, _, _ := customerRequest.asCustomer(testCustomerSSNStorage(t))
	err := repo.createCustomer(cust, "organization")
	require.NoError(t, err)

	address := address{
		Address1:   "123 1st st",
		City:       "Denver",
		State:      "CO",
		PostalCode: "12345",
		Country:    "USA",
		Type:       "primary",
	}
	payload, err := json.Marshal(address)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", fmt.Sprintf("/customers/%s/addresses", cust.CustomerID), bytes.NewReader(payload))
	req.Header.Set("x-organization", "test")
	req.Header.Set("x-request-id", "test")

	router := mux.NewRouter()
	AddCustomerAddressRoutes(log.NewNopLogger(), router, repo)
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusOK, w.Code)

	var customerResp *client.Customer
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &customerResp))
	got := customerResp.Addresses[0]
	want := client.CustomerAddress{
		AddressID:  got.AddressID,
		Type:       address.Type,
		Address1:   address.Address1,
		Address2:   address.Address2,
		City:       address.City,
		State:      address.State,
		PostalCode: address.PostalCode,
		Country:    address.Country,
	}

	require.Equal(t, want, got)
}

func TestCustomers__updateCustomerAddress(t *testing.T) {
	db := createTestCustomerRepository(t)
	repo := NewCustomerRepo(log.NewNopLogger(), db.db)

	customerRequest := customerRequest{
		FirstName: "Jane",
		LastName:  "Doe",
	}
	cust, _, _ := customerRequest.asCustomer(testCustomerSSNStorage(t))
	err := repo.createCustomer(cust, "organization")
	require.NoError(t, err)

	address := address{
		Address1:   "123 1st st",
		City:       "Denver",
		State:      "CO",
		PostalCode: "12345",
		Country:    "USA",
		Type:       "primary",
	}
	require.NoError(t, repo.addCustomerAddress(cust.CustomerID, address))
	cust, err = repo.getCustomer(cust.CustomerID) // refresh customer object after updating address
	require.NoError(t, err)

	updateReq := updateCustomerAddressRequest{
		Type:       "primary",
		Address1:   "555 5th st",
		City:       "Denver",
		State:      "CO",
		PostalCode: "12345",
		Country:    "USA",
		Validated:  true,
	}
	payload, err := json.Marshal(updateReq)
	require.NoError(t, err)

	addressID := cust.Addresses[0].AddressID

	router := mux.NewRouter()
	AddCustomerAddressRoutes(log.NewNopLogger(), router, repo)

	url := fmt.Sprintf("/customers/%s/addresses/%s", cust.CustomerID, addressID)
	req, err := http.NewRequest("PUT", url, bytes.NewReader(payload))
	require.NoError(t, err)

	req.Header.Set("x-organization", "test")
	req.Header.Set("x-request-id", "test")

	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	require.Equal(t, http.StatusOK, res.Code)

	var customerResp *client.Customer
	require.NoError(t, json.Unmarshal(res.Body.Bytes(), &customerResp))

	got := customerResp.Addresses[0]
	want := client.CustomerAddress{
		AddressID:  got.AddressID,
		Type:       updateReq.Type,
		Address1:   updateReq.Address1,
		Address2:   updateReq.Address2,
		City:       updateReq.City,
		State:      updateReq.State,
		PostalCode: updateReq.PostalCode,
		Country:    updateReq.Country,
		Validated:  updateReq.Validated,
	}
	require.Equal(t, want, got)
}

func TestCustomers__deleteCustomerAddress(t *testing.T) {
	db := createTestCustomerRepository(t)
	repo := NewCustomerRepo(log.NewNopLogger(), db.db)

	customerRequest := customerRequest{
		FirstName: "Jane",
		LastName:  "Doe",
	}
	cust, _, _ := customerRequest.asCustomer(testCustomerSSNStorage(t))
	err := repo.createCustomer(cust, "organization")
	require.NoError(t, err)

	address := address{
		Address1:   "123 1st st",
		City:       "Denver",
		State:      "CO",
		PostalCode: "12345",
		Country:    "USA",
		Type:       "primary",
	}
	require.NoError(t, repo.addCustomerAddress(cust.CustomerID, address))

	cust, err = repo.getCustomer(cust.CustomerID)
	require.NoError(t, err)
	addressID := cust.Addresses[0].AddressID

	router := mux.NewRouter()
	AddCustomerAddressRoutes(log.NewNopLogger(), router, repo)

	url := fmt.Sprintf("/customers/%s/addresses/%s", cust.CustomerID, addressID)
	req, err := http.NewRequest("DELETE", url, nil)
	require.NoError(t, err)
	req.Header.Set("x-organization", "test")
	req.Header.Set("x-request-id", "test")

	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	require.Equal(t, http.StatusNoContent, res.Code)

	cust, err = repo.getCustomer(cust.CustomerID)
	require.NoError(t, err)
	require.Empty(t, cust.Addresses)
}

func TestCustomers__updateCustomerAddressFailure(t *testing.T) {
	repo := &testCustomerRepository{}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/customers/foo/address/bar", nil)
	updateCustomerAddress(log.NewNopLogger(), repo)(w, req)
	w.Flush()

	if w.Code != http.StatusBadRequest {
		t.Errorf("bogus HTTP status: %d", w.Code)
	}

	// try the proper HTTP verb
	w = httptest.NewRecorder()
	req = httptest.NewRequest("PUT", "/customers/foo/address/bar", nil)
	updateCustomerAddress(log.NewNopLogger(), repo)(w, req)
	w.Flush()

	if w.Code != http.StatusBadRequest {
		t.Errorf("bogus HTTP status: %d", w.Code)
	}
}

func TestCustomerRepository__updateCustomerAddress(t *testing.T) {
	db := createTestCustomerRepository(t)
	repo := NewCustomerRepo(log.NewNopLogger(), db.db)

	customerRequest := customerRequest{
		FirstName: "Jane",
		LastName:  "Doe",
	}
	cust, _, _ := customerRequest.asCustomer(testCustomerSSNStorage(t))
	err := repo.createCustomer(cust, "organization")
	require.NoError(t, err)

	address := address{
		Address1:   "123 1st st",
		City:       "Denver",
		State:      "CO",
		PostalCode: "12345",
		Country:    "USA",
		Type:       "primary",
	}
	require.NoError(t, repo.addCustomerAddress(cust.CustomerID, address))

	cust, err = repo.getCustomer(cust.CustomerID)
	require.NoError(t, err)

	addressID := cust.Addresses[0].AddressID
	updateReq := updateCustomerAddressRequest{
		Type:       "primary",
		Address1:   "555 5th st",
		City:       "Denver",
		State:      "CO",
		PostalCode: "12345",
		Country:    "USA",
		Validated:  true,
	}
	err = repo.updateCustomerAddress(cust.CustomerID, addressID, updateReq)
	require.NoError(t, err)

	cust, err = repo.getCustomer(cust.CustomerID)
	require.NoError(t, err)

	require.Len(t, cust.Addresses, 1)
	want := client.CustomerAddress{
		AddressID:  addressID,
		Type:       updateReq.Type,
		Address1:   updateReq.Address1,
		Address2:   updateReq.Address2,
		City:       updateReq.City,
		State:      updateReq.State,
		PostalCode: updateReq.PostalCode,
		Country:    updateReq.Country,
		Validated:  updateReq.Validated,
	}
	got := cust.Addresses[0]
	require.Equal(t, want, got)
}

func TestCustomerRepository__deleteCustomerAddress(t *testing.T) {
	db := createTestCustomerRepository(t)
	repo := NewCustomerRepo(log.NewNopLogger(), db.db)

	customerRequest := customerRequest{
		FirstName: "Jane",
		LastName:  "Doe",
	}
	cust, _, _ := customerRequest.asCustomer(testCustomerSSNStorage(t))
	err := repo.createCustomer(cust, "organization")
	require.NoError(t, err)

	address := address{
		Address1:   "123 1st st",
		City:       "Denver",
		State:      "CO",
		PostalCode: "12345",
		Country:    "USA",
		Type:       "primary",
	}
	require.NoError(t, repo.addCustomerAddress(cust.CustomerID, address))

	cust, err = repo.getCustomer(cust.CustomerID)
	require.NoError(t, err)

	addressID := cust.Addresses[0].AddressID
	err = repo.deleteCustomerAddress(cust.CustomerID, addressID)
	require.NoError(t, err)

	cust, err = repo.getCustomer(cust.CustomerID)
	require.NoError(t, err)

	require.Len(t, cust.Addresses, 0)
}
