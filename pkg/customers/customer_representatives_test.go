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
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/moov-io/base/database"

	"github.com/moov-io/customers/pkg/client"

	"github.com/gorilla/mux"
	"github.com/moov-io/base/log"
)

func setupMockOrganizationCustomerAndRepresentative(t *testing.T, repo *sqlCustomerRepository) (string, *client.Customer, *client.Representative) {
	organization := "best-business"
	cust, _, _ := (customerRequest{
		FirstName:    "Jane",
		LastName:     "Doe",
		Email:        "jane@example.com",
		BusinessName: "Best Business",
		BusinessType: client.BUSINESSTYPE_CORPORATION,
		EIN:          "123-456-789",
	}).asCustomer(testCustomerSSNStorage(t))
	if err := repo.CreateCustomer(cust, organization); err != nil {
		t.Fatal(err)
	}

	rep, _, _ := (customerRepresentativeRequest{
		CustomerID: cust.CustomerID,
		FirstName:  "Jane",
		LastName:   "Doe",
		JobTitle:   "CEO",
	}).asRepresentative(testCustomerSSNStorage(t))
	if err := repo.CreateRepresentative(rep, cust.CustomerID); err != nil {
		t.Fatal(err)
	}

	return organization, cust, rep
}

func setupMockCustomer(t *testing.T, repo *sqlCustomerRepository) (*client.Customer, string) {
	organization := "best-business"
	cust, _, _ := (customerRequest{
		FirstName:    "Jane",
		LastName:     "Doe",
		Email:        "jane@example.com",
		BusinessName: "Best Business",
		BusinessType: client.BUSINESSTYPE_CORPORATION,
		EIN:          "123-456-789",
	}).asCustomer(testCustomerSSNStorage(t))
	if err := repo.CreateCustomer(cust, organization); err != nil {
		t.Fatal(err)
	}

	return cust, organization
}

func TestCustomers__DeleteRepresentative(t *testing.T) {
	repo := createTestCustomerRepository(t)
	defer repo.close()

	_, cust, rep := setupMockOrganizationCustomerAndRepresentative(t, repo)

	got, err := repo.GetRepresentative(rep.RepresentativeID)
	require.NoError(t, err)
	require.NotNil(t, got)

	router := mux.NewRouter()
	w := httptest.NewRecorder()
	req := httptest.NewRequest("DELETE", fmt.Sprintf("/customers/%s/representatives/%s", cust.CustomerID, rep.RepresentativeID), nil)

	AddRepresentativeRoutes(log.NewNopLogger(), router, repo, testCustomerSSNStorage(t))
	router.ServeHTTP(w, req)
	w.Flush()

	require.Equal(t, http.StatusNoContent, w.Code)
	require.Empty(t, w.Body)
	got, err = repo.GetRepresentative(rep.RepresentativeID)
	require.Error(t, err)
	require.Nil(t, got)
}

func TestCustomerRepository__CreateRepresentative(t *testing.T) {
	check := func(t *testing.T, repo *sqlCustomerRepository) {
		organization, cust, rep := setupMockOrganizationCustomerAndRepresentative(t, repo)

		cust, err := repo.GetCustomer(cust.CustomerID, organization)
		if err != nil {
			t.Fatal(err)
		}
		if cust == nil {
			t.Error("got nil Customer")
		}

		rep, err = repo.GetRepresentative(rep.RepresentativeID)
		if err != nil {
			t.Fatal(err)
		}
		if rep == nil {
			t.Error("got nil Representative")
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

func TestCustomers__customerRepresentativeRequest(t *testing.T) {
	req := &customerRepresentativeRequest{JobTitle: "awesome job title"}
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

	req.Phones = append(req.Phones, phone{
		Number: "123.456.7890",
		Type:   "mobile",
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
	representative, _, _ := req.asRepresentative(testCustomerSSNStorage(t))
	if representative.RepresentativeID == "" {
		t.Errorf("empty Customer Representative: %#v", representative)
	}
	if len(representative.Phones) != 1 {
		t.Errorf("representative.Phones: %#v", representative.Phones)
	}
	if len(representative.Addresses) != 1 {
		t.Errorf("representative.Addresses: %#v", representative.Addresses)
	}
}

func TestCustomers__CreateRepresentative(t *testing.T) {
	repo := createTestCustomerRepository(t)
	defer repo.close()

	newCust, organization := setupMockCustomer(t, repo)
	if newCust.CustomerID == "" {
		t.Error("empty Customer.CustomerID")
	}

	w := httptest.NewRecorder()
	phone := `{"number": "555.555.5555", "type": "mobile", "ownerType": "representative"}`
	address := `{"type": "primary", "ownerType": "representative", "address1": "123 1st St", "city": "Denver", "state": "CO", "postalCode": "12345", "country": "USA"}`
	body := fmt.Sprintf(`{"firstName": "jane", "lastName": "doe", "birthDate": "1991-04-01", "ssn": "987654321", "phones": [%s], "addresses": [%s]}`, phone, address)
	req := httptest.NewRequest("POST", fmt.Sprintf("/customers/%s/representatives", newCust.CustomerID), strings.NewReader(body))
	req.Header.Set("x-organization", organization)
	req.Header.Set("x-request-id", "test")

	customerSSNStorage := testCustomerSSNStorage(t)

	router := mux.NewRouter()
	AddRepresentativeRoutes(log.NewNopLogger(), router, repo, customerSSNStorage)
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("bogus status code: %d: %v", w.Code, w.Body.String())
	}

	var cust *client.Customer
	if err := json.NewDecoder(w.Body).Decode(&cust); err != nil {
		t.Fatal(err)
	}
	if cust.CustomerID == "" {
		t.Error("empty Customer.CustomerID")
	}

	// sad path
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", fmt.Sprintf("/customers/%s/representatives", newCust.CustomerID), strings.NewReader("null"))
	req.Header.Set("x-organization", "test")
	req.Header.Set("x-request-id", "test")
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusBadRequest {
		t.Fatalf("bogus HTTP status code: %d", w.Code)
	}

	// customerSSNStorage sad path
	w = httptest.NewRecorder()
	req = httptest.NewRequest("POST", fmt.Sprintf("/customers/%s/representatives", newCust.CustomerID), strings.NewReader(body))
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
	if s := w.Body.String(); !strings.Contains(s, "saveSSN: ") {
		t.Errorf("unexpected error: %v", s)
	}
}

func TestCustomers__updateRepresentative(t *testing.T) {
	repo := createTestCustomerRepository(t)
	defer repo.close()

	organization := "best-business"
	createCustomerReq := &customerRequest{
		FirstName:    "Jane",
		LastName:     "Doe",
		Email:        "jane@example.com",
		BusinessName: "Best Business",
		BusinessType: client.BUSINESSTYPE_CORPORATION,
		EIN:          "123-456-789",
	}
	cust, _, _ := createCustomerReq.asCustomer(testCustomerSSNStorage(t))
	if err := repo.CreateCustomer(cust, organization); err != nil {
		t.Fatal(err)
	}

	createRepresentativeReq := &customerRepresentativeRequest{
		CustomerID: cust.CustomerID,
		FirstName:  "Jane",
		LastName:   "Doe",
		JobTitle:   "CEO",
		Phones: []phone{
			{
				Number: "123.456.7890",
				Type:   "mobile",
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
	rep, _, _ := createRepresentativeReq.asRepresentative(testCustomerSSNStorage(t))
	if err := repo.CreateRepresentative(rep, cust.CustomerID); err != nil {
		t.Fatal(err)
	}

	_, err := repo.GetRepresentative(rep.RepresentativeID)
	require.NoError(t, err)

	router := mux.NewRouter()
	w := httptest.NewRecorder()

	updateReq := *createRepresentativeReq
	updateReq.RepresentativeID = rep.RepresentativeID
	updateReq.FirstName = "Jim"
	updateReq.LastName = "Smith"
	updateReq.JobTitle = "CEO"
	updateReq.BirthDate = "2020-01-01"
	updateReq.Phones = []phone{
		{
			Number:    "555.555.5555",
			Type:      "mobile",
			OwnerType: "representative",
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
			OwnerType:  "representative",
		},
	}
	payload, err := json.Marshal(&updateReq)
	require.NoError(t, err)

	req := httptest.NewRequest("PUT", fmt.Sprintf("/customers/%s/representatives/%s", cust.CustomerID, rep.RepresentativeID), bytes.NewReader(payload))
	req.Header.Set("x-organization", organization)
	req.Header.Set("x-request-id", "test")
	AddRepresentativeRoutes(log.NewNopLogger(), router, repo, testCustomerSSNStorage(t))
	router.ServeHTTP(w, req)
	w.Flush()
	require.Equal(t, http.StatusOK, w.Code)

	var got *client.Representative
	require.NoError(t, json.NewDecoder(w.Body).Decode(&got))
	want, _, _ := updateReq.asRepresentative(testCustomerSSNStorage(t))
	require.NoError(t, err)
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

func TestCustomerRepository__updateRepresentative(t *testing.T) {
	repo := createTestCustomerRepository(t)
	defer repo.close()

	organization := "best-business"
	createReq := &customerRequest{
		FirstName:    "Jane",
		LastName:     "Doe",
		Email:        "jane@example.com",
		BusinessName: "Best Business",
		BusinessType: client.BUSINESSTYPE_CORPORATION,
		EIN:          "123-456-789",
	}
	cust, _, _ := createReq.asCustomer(testCustomerSSNStorage(t))
	if err := repo.CreateCustomer(cust, organization); err != nil {
		t.Fatal(err)
	}

	createRepresentativeReq := &customerRepresentativeRequest{
		CustomerID: cust.CustomerID,
		FirstName:  "Jane",
		LastName:   "Doe",
		JobTitle:   "CEO",
		Phones: []phone{
			{
				Number: "123.456.7890",
				Type:   "mobile",
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
	rep, _, _ := createRepresentativeReq.asRepresentative(testCustomerSSNStorage(t))
	err := repo.CreateRepresentative(rep, cust.CustomerID)
	require.NoError(t, err)

	updateReq := customerRepresentativeRequest{
		RepresentativeID: rep.RepresentativeID,
		CustomerID:       cust.CustomerID,
		FirstName:        "Jim",
		LastName:         "Smith",
		Phones: []phone{
			{
				Number: "555.555.5555",
				Type:   "mobile",
			},
		},
		Addresses: []address{
			{
				Address1: "555 5th st",
				City:     "real city",
				Type:     "primary",
			},
		},
	}

	updatedRep, _, _ := updateReq.asRepresentative(testCustomerSSNStorage(t))
	err = repo.updateRepresentative(updatedRep, cust.CustomerID)
	require.NoError(t, err)

	require.Equal(t, updateReq.FirstName, updatedRep.FirstName)
	require.Equal(t, updateReq.LastName, updatedRep.LastName)
	require.Equal(t, updateReq.Phones[0].Number, updatedRep.Phones[0].Number)
	require.Equal(t, updateReq.Addresses[0].Address1, updatedRep.Addresses[0].Address1)
	require.Equal(t, updateReq.Addresses[0].City, updatedRep.Addresses[0].City)
}

func TestCustomersRepository__addRepresentativeAddress(t *testing.T) {
	repo := createTestCustomerRepository(t)
	defer repo.close()

	organization := "best-business"
	createReq := &customerRequest{
		FirstName:    "Jane",
		LastName:     "Doe",
		Email:        "jane@example.com",
		BusinessName: "Best Business",
		BusinessType: client.BUSINESSTYPE_CORPORATION,
		EIN:          "123-456-789",
	}
	cust, _, _ := createReq.asCustomer(testCustomerSSNStorage(t))
	if err := repo.CreateCustomer(cust, organization); err != nil {
		t.Fatal(err)
	}

	createRepresentativeReq := &customerRepresentativeRequest{
		CustomerID: cust.CustomerID,
		FirstName:  "Jane",
		LastName:   "Doe",
		JobTitle:   "CEO",
		Phones: []phone{
			{
				Number: "123.456.7890",
				Type:   "mobile",
			},
		},
	}
	rep, _, _ := createRepresentativeReq.asRepresentative(testCustomerSSNStorage(t))
	if err := repo.CreateRepresentative(rep, cust.CustomerID); err != nil {
		t.Fatal(err)
	}

	// add an address
	if err := repo.addAddress(rep.RepresentativeID, client.OWNERTYPE_REPRESENTATIVE, address{
		Address1:   "123 1st st",
		City:       "fake city",
		State:      "CA",
		PostalCode: "90210",
		Country:    "US",
		Type:       "primary",
	},
	); err != nil {
		t.Fatal(err)
	}

	// re-read
	cust, err := repo.GetCustomer(cust.CustomerID, organization)
	if err != nil {
		t.Fatal(err)
	}
	if len(cust.Representatives[0].Addresses) != 1 {
		t.Errorf("got %d Addresses", len(cust.Representatives[0].Addresses))
	}
	if cust.Representatives[0].Addresses[0].Address1 != "123 1st st" {
		t.Errorf("rep.Addresses[0].Address1=%s", rep.Addresses[0].Address1)
	}
}

func TestRepresentatives__minimumFields(t *testing.T) {
	repo := createTestCustomerRepository(t)
	defer repo.close()

	organization, cust, _ := setupMockOrganizationCustomerAndRepresentative(t, repo)

	w := httptest.NewRecorder()
	body := `{"firstName": "jane", "lastName": "doe"}`
	req := httptest.NewRequest("POST", fmt.Sprintf("/customers/%s/representatives", cust.CustomerID), strings.NewReader(body))
	req.Header.Set("x-organization", organization)
	req.Header.Set("x-request-id", "test")

	router := mux.NewRouter()
	AddRepresentativeRoutes(log.NewNopLogger(), router, repo, testCustomerSSNStorage(t))
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("bogus status code: %d: %v", w.Code, w.Body.String())
	}
}

func TestRepresentatives__BadReq(t *testing.T) {
	repo := createTestCustomerRepository(t)
	defer repo.close()

	organization, cust, _ := setupMockOrganizationCustomerAndRepresentative(t, repo)

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", fmt.Sprintf("/customers/%s/representatives", cust.CustomerID), strings.NewReader("®"))
	req.Header.Set("x-organization", organization)
	req.Header.Set("x-request-id", "test")

	router := mux.NewRouter()
	AddRepresentativeRoutes(log.NewNopLogger(), router, repo, testCustomerSSNStorage(t))
	router.ServeHTTP(w, req)
	w.Flush()

	if !strings.Contains(w.Body.String(), "invalid character") {
		t.Errorf("Expected SSN error received %s", w.Body.String())
	}
}
