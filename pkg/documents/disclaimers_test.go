// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package documents

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/moov-io/base"
	"github.com/moov-io/base/admin"
	"github.com/moov-io/customers/internal/database"
	"github.com/moov-io/customers/pkg/client"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

type testDisclaimerRepository struct {
	disclaimers []*client.Disclaimer
	err         error
}

func (r *testDisclaimerRepository) getCustomerDisclaimer(customerID, documentID string) (*client.Disclaimer, error) {
	if r.err != nil {
		return nil, r.err
	}
	if len(r.disclaimers) > 0 {
		return r.disclaimers[0], nil
	}
	return nil, nil
}

func (r *testDisclaimerRepository) getCustomerDisclaimers(customerID string) ([]*client.Disclaimer, error) {
	if r.err != nil {
		return nil, r.err
	}
	return r.disclaimers, nil
}

func (r *testDisclaimerRepository) acceptDisclaimer(customerID, disclaimerID string) error {
	return r.err
}

func (r *testDisclaimerRepository) insertDisclaimer(text, documentID string) (*client.Disclaimer, error) {
	if r.err != nil {
		return nil, r.err
	}
	if len(r.disclaimers) == 0 {
		r.disclaimers = append(r.disclaimers, &client.Disclaimer{
			DisclaimerID: base.ID(),
			Text:         text,
			DocumentID:   documentID,
		})
		return r.disclaimers[0], nil
	}
	return nil, nil
}

func TestDisclaimers__getDisclaimerID(t *testing.T) {
	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/ping", nil)
	req.Header.Set("x-request-id", "test")

	if id := getDisclaimerID(w, req); id != "" {
		t.Errorf("unexpected ID: %s", id)
	}
}

func TestDisclaimers__getCustomerDisclaimer(t *testing.T) {
	check := func(t *testing.T, repo *sqlDisclaimerRepository) {
		defer repo.close()

		customerID, disclaimerID := base.ID(), base.ID()

		disclaimer, err := repo.getCustomerDisclaimer(customerID, disclaimerID)
		if err != nil {
			t.Fatal(err)
		}
		if disclaimer != nil {
			t.Errorf("expected no disclaimer")
		}
	}

	// SQLite tests
	sqliteDB := database.CreateTestSqliteDB(t)
	defer sqliteDB.Close()
	check(t, &sqlDisclaimerRepository{sqliteDB.DB, log.NewNopLogger()})

	// MySQL tests
	mysqlDB := database.CreateTestMySQLDB(t)
	defer mysqlDB.Close()
	check(t, &sqlDisclaimerRepository{mysqlDB.DB, log.NewNopLogger()})
}

func TestDisclaimers__getCustomerDisclaimers(t *testing.T) {
	repo := &testDisclaimerRepository{
		disclaimers: []*client.Disclaimer{
			{
				DisclaimerID: base.ID(),
				Text:         "terms and conditions",
			},
		},
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/customers/adam/disclaimers", nil)
	req.Header.Set("x-request-id", "test")

	router := mux.NewRouter()
	AddDisclaimerRoutes(log.NewNopLogger(), router, repo)
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("bogus HTTP status: %d", w.Code)
	}

	var resp []*client.Disclaimer
	if err := json.NewDecoder(w.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if len(resp) != 1 {
		t.Errorf("disclaimers: %#v", resp)
	}

	// set an error and verify
	repo.err = errors.New("bad error")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusBadRequest {
		t.Errorf("bogus HTTP status: %d", w.Code)
	}
}

func TestDisclaimers__acceptDisclaimer(t *testing.T) {
	disclaimerID := base.ID()
	repo := &testDisclaimerRepository{
		disclaimers: []*client.Disclaimer{
			{
				DisclaimerID: disclaimerID,
				Text:         "terms and conditions",
			},
		},
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest("POST", fmt.Sprintf("/customers/adam/disclaimers/%s", disclaimerID), nil)
	req.Header.Set("x-request-id", "test")

	router := mux.NewRouter()
	AddDisclaimerRoutes(log.NewNopLogger(), router, repo)
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusOK {
		t.Errorf("bogus HTTP status: %d", w.Code)
	}

	// set an error and verify
	repo.err = errors.New("bad error")

	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	w.Flush()

	if w.Code != http.StatusBadRequest {
		t.Errorf("bogus HTTP status: %d", w.Code)
	}
}

func TestDisclaimersRepository(t *testing.T) {
	customerID := base.ID()

	check := func(t *testing.T, repo *sqlDisclaimerRepository) {
		defer repo.close()

		disclaimers, err := repo.getCustomerDisclaimers(customerID)
		if err != nil {
			t.Fatal(err)
		}
		if len(disclaimers) != 0 {
			t.Errorf("unexpected disclaimers: %#v", disclaimers)
		}

		// write a Disclaimer and verify
		documentID := ""
		disc, err := repo.insertDisclaimer("terms and conditions", documentID)
		if err != nil {
			t.Fatal(err)
		}
		disclaimers, err = repo.getCustomerDisclaimers(customerID)
		if err != nil {
			t.Fatal(err)
		}
		if len(disclaimers) != 1 {
			t.Errorf("unexpected disclaimers: %#v", disclaimers)
		}

		// Accept the disclaimer
		if err := repo.acceptDisclaimer(customerID, disc.DisclaimerID); err != nil {
			t.Fatal(err)
		}

		// Verify a different disclaimer ID is rejected
		if err := repo.acceptDisclaimer(customerID, base.ID()); err == nil {
			t.Error("expected error")
		}
	}

	// SQLite tests
	sqliteDB := database.CreateTestSqliteDB(t)
	defer sqliteDB.Close()
	check(t, &sqlDisclaimerRepository{sqliteDB.DB, log.NewNopLogger()})

	// MySQL tests
	mysqlDB := database.CreateTestMySQLDB(t)
	defer mysqlDB.Close()
	check(t, &sqlDisclaimerRepository{mysqlDB.DB, log.NewNopLogger()})
}

func TestDisclaimersAdmin__create(t *testing.T) {
	disclaimerRepo := &testDisclaimerRepository{}

	documentID := base.ID()
	docRepo := &testDocumentRepository{
		documents: []*client.Document{
			{
				DocumentID: documentID,
			},
		},
	}

	svc := admin.NewServer(":0")
	defer svc.Shutdown()
	AddDisclaimerAdminRoutes(log.NewNopLogger(), svc, disclaimerRepo, docRepo)
	go svc.Listen()

	body := strings.NewReader(fmt.Sprintf(`{"text": "terms and conditions", "documentId": "%s"}`, documentID))
	req, err := http.NewRequest("POST", "http://"+svc.BindAddr()+"/customers/adam/disclaimers", body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("x-organization", "test")
	req.Header.Set("x-request-id", "test")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	respBody, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		t.Errorf("bogus HTTP status: %d: %v", resp.StatusCode, string(respBody))
	}

	var disclaimer client.Disclaimer
	if err := json.NewDecoder(bytes.NewReader(respBody)).Decode(&disclaimer); err != nil {
		t.Fatal(err)
	}

	if disclaimer.Text != "terms and conditions" {
		t.Errorf("disclaimer.Text=%s", disclaimer.Text)
	}
}

func TestDisclaimersAdmin__createErr(t *testing.T) {
	disclaimerRepo := &testDisclaimerRepository{}
	docRepo := &testDocumentRepository{}

	svc := admin.NewServer(":0")
	defer svc.Shutdown()
	AddDisclaimerAdminRoutes(log.NewNopLogger(), svc, disclaimerRepo, docRepo)
	go svc.Listen()

	body := strings.NewReader(`{"text": "", "documentId": ""}`)
	req, err := http.NewRequest("POST", "http://"+svc.BindAddr()+"/customers/adam/disclaimers", body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("x-organization", "test")
	req.Header.Set("x-request-id", "test")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	respBody, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("bogus HTTP status: %d: %v", resp.StatusCode, string(respBody))
	}
}

func TestDisclaimersAdmin__createMethodErr(t *testing.T) {
	disclaimerRepo := &testDisclaimerRepository{}
	docRepo := &testDocumentRepository{}

	svc := admin.NewServer(":0")
	defer svc.Shutdown()
	AddDisclaimerAdminRoutes(log.NewNopLogger(), svc, disclaimerRepo, docRepo)
	go svc.Listen()

	req, err := http.NewRequest("PUT", "http://"+svc.BindAddr()+"/customers/adam/disclaimers", nil)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("x-organization", "test")
	req.Header.Set("x-request-id", "test")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	respBody, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("bogus HTTP status: %d: %v", resp.StatusCode, string(respBody))
	}
}

func TestDisclaimersAdmin__createJSONErr(t *testing.T) {
	disclaimerRepo := &testDisclaimerRepository{}
	docRepo := &testDocumentRepository{}

	svc := admin.NewServer(":0")
	defer svc.Shutdown()
	AddDisclaimerAdminRoutes(log.NewNopLogger(), svc, disclaimerRepo, docRepo)
	go svc.Listen()

	body := strings.NewReader(`not-valid-json`)
	req, err := http.NewRequest("POST", "http://"+svc.BindAddr()+"/customers/adam/disclaimers", body)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("x-organization", "test")
	req.Header.Set("x-request-id", "test")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	respBody, _ := ioutil.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("bogus HTTP status: %d: %v", resp.StatusCode, string(respBody))
	}
}
