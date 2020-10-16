package customers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/moov-io/base/log"
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
	organization := "organization"
	err := repo.CreateCustomer(cust, organization)
	require.NoError(t, err)

	addrPayload := address{
		Address1:   "123 1st st",
		City:       "Denver",
		State:      "CO",
		PostalCode: "12345",
		Country:    "USA",
		Type:       "primary",
	}
	payload, err := json.Marshal(addrPayload)
	require.NoError(t, err)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", fmt.Sprintf("/customers/%s/addresses", cust.CustomerID), bytes.NewReader(payload))
	req.Header.Set("x-organization", organization)
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
		Type:       addrPayload.Type,
		Address1:   addrPayload.Address1,
		Address2:   addrPayload.Address2,
		City:       addrPayload.City,
		State:      addrPayload.State,
		PostalCode: addrPayload.PostalCode,
		Country:    addrPayload.Country,
	}

	require.Equal(t, want, got)

	/* Error on duplicate type primary */
	w = httptest.NewRecorder()
	payload, err = json.Marshal(addrPayload)
	require.NoError(t, err)
	req = httptest.NewRequest("POST", fmt.Sprintf("/customers/%s/addresses", cust.CustomerID), bytes.NewReader(payload))
	req.Header.Set("x-organization", organization)
	router.ServeHTTP(w, req)
	require.Equal(t, http.StatusBadRequest, w.Code)
	var errResp struct {
		ErrorMsg string `json:"error"`
	}
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &errResp))
	require.Contains(t, errResp.ErrorMsg, ErrAddressTypeDuplicate.Error())
}

func TestCustomers__updateCustomerAddress(t *testing.T) {
	db := createTestCustomerRepository(t)
	repo := NewCustomerRepo(log.NewNopLogger(), db.db)

	customerRequest := customerRequest{
		FirstName: "Jane",
		LastName:  "Doe",
	}
	cust, _, _ := customerRequest.asCustomer(testCustomerSSNStorage(t))
	organization := "organization"
	err := repo.CreateCustomer(cust, organization)
	require.NoError(t, err)

	addrRequests := []address{
		{
			Address1:   "111 1st st",
			City:       "Denver",
			State:      "CO",
			PostalCode: "12345",
			Country:    "USA",
			Type:       "primary",
		},
		{
			Address1:   "222 2nd st",
			City:       "Denver",
			State:      "CO",
			PostalCode: "12345",
			Country:    "USA",
			Type:       "secondary",
		},
	}
	for _, req := range addrRequests {
		require.NoError(t, repo.addCustomerAddress(cust.CustomerID, req))
		cust, err = repo.GetCustomer(cust.CustomerID, organization) // refresh customer object after updating address
		require.NoError(t, err)
	}

	// find address with primaryid
	var primaryAddressID string
	var secondaryAddressID string
	for _, addr := range cust.Addresses {
		if addr.Type == "primary" {
			primaryAddressID = addr.AddressID
		} else {
			secondaryAddressID = addr.AddressID
		}
	}

	updateReq := updateCustomerAddressRequest{
		address: address{
			Type:       "primary",
			Address1:   "555 5th st",
			City:       "Denver",
			State:      "CO",
			PostalCode: "12345",
			Country:    "USA",
		},
		Validated: true,
	}
	payload, err := json.Marshal(updateReq)
	require.NoError(t, err)

	router := mux.NewRouter()
	AddCustomerAddressRoutes(log.NewNopLogger(), router, repo)

	url := fmt.Sprintf("/customers/%s/addresses/%s", cust.CustomerID, primaryAddressID)
	req, err := http.NewRequest("PUT", url, bytes.NewReader(payload))
	require.NoError(t, err)

	req.Header.Set("x-organization", organization)
	req.Header.Set("x-request-id", "test")

	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	require.Equalf(t, http.StatusOK, res.Code, "response body: %s", res.Body.String())

	var customerResp *client.Customer
	require.NoError(t, json.Unmarshal(res.Body.Bytes(), &customerResp))

	var got client.CustomerAddress
	for _, a := range customerResp.Addresses {
		if a.AddressID == primaryAddressID {
			got = a
		}
	}
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

	/* Error when trying to update a secondary address to primary when one already exists */
	res = httptest.NewRecorder()
	payload, err = json.Marshal(updateReq)
	require.NoError(t, err)
	url = fmt.Sprintf("/customers/%s/addresses/%s", cust.CustomerID, secondaryAddressID)

	req, err = http.NewRequest("PUT", url, bytes.NewReader(payload))
	req.Header.Set("x-organization", organization)
	require.NoError(t, err)
	router.ServeHTTP(res, req)
	require.Equal(t, http.StatusBadRequest, res.Code)
	var errResp struct {
		ErrorMsg string `json:"error"`
	}
	require.NoError(t, json.Unmarshal(res.Body.Bytes(), &errResp))
	require.Contains(t, errResp.ErrorMsg, ErrAddressTypeDuplicate.Error())
}

func TestCustomers__deleteCustomerAddress(t *testing.T) {
	db := createTestCustomerRepository(t)
	repo := NewCustomerRepo(log.NewNopLogger(), db.db)

	customerRequest := customerRequest{
		FirstName: "Jane",
		LastName:  "Doe",
	}
	cust, _, _ := customerRequest.asCustomer(testCustomerSSNStorage(t))
	organization := "organization"
	err := repo.CreateCustomer(cust, organization)
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

	cust, err = repo.GetCustomer(cust.CustomerID, organization)
	require.NoError(t, err)
	addressID := cust.Addresses[0].AddressID

	router := mux.NewRouter()
	AddCustomerAddressRoutes(log.NewNopLogger(), router, repo)

	url := fmt.Sprintf("/customers/%s/addresses/%s", cust.CustomerID, addressID)
	req, err := http.NewRequest("DELETE", url, nil)
	require.NoError(t, err)
	req.Header.Set("x-organization", organization)
	req.Header.Set("x-request-id", "test")

	res := httptest.NewRecorder()
	router.ServeHTTP(res, req)
	require.Equal(t, http.StatusNoContent, res.Code)

	cust, err = repo.GetCustomer(cust.CustomerID, organization)
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
	organization := "organization"
	err := repo.CreateCustomer(cust, organization)
	require.NoError(t, err)

	addrRequest := address{
		Address1:   "123 1st st",
		City:       "Denver",
		State:      "CO",
		PostalCode: "12345",
		Country:    "USA",
		Type:       "primary",
	}
	require.NoError(t, repo.addCustomerAddress(cust.CustomerID, addrRequest))

	cust, err = repo.GetCustomer(cust.CustomerID, organization)
	require.NoError(t, err)

	addressID := cust.Addresses[0].AddressID
	updateReq := updateCustomerAddressRequest{
		address: address{
			Type:       "primary",
			Address1:   "555 5th st",
			City:       "Denver",
			State:      "CO",
			PostalCode: "12345",
			Country:    "USA",
		},
		Validated: true,
	}
	err = repo.updateCustomerAddress(cust.CustomerID, addressID, updateReq)
	require.NoError(t, err)

	cust, err = repo.GetCustomer(cust.CustomerID, organization)
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
	organization := "organization"
	err := repo.CreateCustomer(cust, organization)
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

	cust, err = repo.GetCustomer(cust.CustomerID, organization)
	require.NoError(t, err)

	addressID := cust.Addresses[0].AddressID
	err = repo.deleteCustomerAddress(cust.CustomerID, addressID)
	require.NoError(t, err)

	cust, err = repo.GetCustomer(cust.CustomerID, organization)
	require.NoError(t, err)

	require.Len(t, cust.Addresses, 0)
}
