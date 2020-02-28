// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/moov-io/base"
	"github.com/moov-io/base/admin"
	moovhttp "github.com/moov-io/base/http"
	client "github.com/moov-io/customers/client"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

var (
	errNoDisclaimerID = errors.New("no Disclaimer ID found")
)

func addDisclaimerRoutes(logger log.Logger, r *mux.Router, repo disclaimerRepository) {
	r.Methods("GET").Path("/customers/{customerID}/disclaimers").HandlerFunc(getCustomerDisclaimers(logger, repo))
	r.Methods("POST").Path("/customers/{customerID}/disclaimers/{disclaimerID}").HandlerFunc(acceptDisclaimer(logger, repo))
}

func addDisclaimerAdminRoutes(logger log.Logger, svc *admin.Server, disclaimRepo disclaimerRepository, docRepo documentRepository) {
	svc.AddHandler("/customers/{customerID}/disclaimers", createDisclaimer(logger, disclaimRepo, docRepo))
}

func getDisclaimerID(w http.ResponseWriter, r *http.Request) string {
	v, ok := mux.Vars(r)["disclaimerID"]
	if !ok || v == "" {
		moovhttp.Problem(w, errNoDisclaimerID)
		return ""
	}
	return v
}

func getCustomerDisclaimers(logger log.Logger, repo disclaimerRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = wrapResponseWriter(logger, w, r)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		customerID := getCustomerID(w, r)
		if customerID == "" {
			return
		}

		disclaimers, err := repo.getCustomerDisclaimers(customerID)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(disclaimers)
	}
}

func acceptDisclaimer(logger log.Logger, repo disclaimerRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = wrapResponseWriter(logger, w, r)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		customerID, disclaimerID := getCustomerID(w, r), getDisclaimerID(w, r)
		if customerID == "" || disclaimerID == "" {
			return
		}

		if err := repo.acceptDisclaimer(customerID, disclaimerID); err != nil {
			moovhttp.Problem(w, err)
			return
		}

		disclaimer, err := repo.getCustomerDisclaimer(customerID, disclaimerID)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(disclaimer)
	}
}

type createDisclaimerRequest struct {
	DocumentID string `json:"documentId,omitempty"`
	Text       string `json:"text"`
}

func createDisclaimer(logger log.Logger, disclaimRepo disclaimerRepository, docRepo documentRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = wrapResponseWriter(logger, w, r)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		if r.Method != "POST" {
			moovhttp.Problem(w, fmt.Errorf("unsupported HTTP verb %s", r.Method))
			return
		}

		customerID := getCustomerID(w, r)
		if customerID == "" {
			return
		}

		var req createDisclaimerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			moovhttp.Problem(w, err)
			return
		}

		if err := documentExistsForCustomer(customerID, req, docRepo); err != nil {
			moovhttp.Problem(w, err)
			return
		}

		if req.Text == "" {
			moovhttp.Problem(w, errors.New("empty disclaimer text"))
			return
		}

		disclaimer, err := disclaimRepo.insertDisclaimer(req.Text, req.DocumentID)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(disclaimer)
	}
}

func documentExistsForCustomer(customerID string, req createDisclaimerRequest, docRepo documentRepository) error {
	if req.DocumentID != "" {
		docs, err := docRepo.getCustomerDocuments(customerID)
		if err != nil {
			return err
		}
		for i := range docs {
			if docs[i].ID == req.DocumentID {
				return nil
			}
		}
		return errors.New("document not found")
	}
	return nil
}

type disclaimerRepository interface {
	getCustomerDisclaimer(customerID, disclaimerID string) (*client.Disclaimer, error)
	getCustomerDisclaimers(customerID string) ([]*client.Disclaimer, error)
	acceptDisclaimer(customerID, disclaimerID string) error
	insertDisclaimer(text, documentID string) (*client.Disclaimer, error)
}

type sqlDisclaimerRepository struct {
	db     *sql.DB
	logger log.Logger
}

func (r *sqlDisclaimerRepository) close() error {
	return r.db.Close()
}

func (r *sqlDisclaimerRepository) getCustomerDisclaimer(customerID, disclaimerID string) (*client.Disclaimer, error) {
	query := `select d.disclaimer_id, d.text, d.document_id, da.accepted_at from disclaimers as d
left outer join disclaimer_acceptances as da on d.disclaimer_id = da.disclaimer_id
where d.deleted_at is null and d.disclaimer_id = ? limit 1;`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var acceptedAt *time.Time
	var d client.Disclaimer

	if err := stmt.QueryRow(disclaimerID).Scan(&d.ID, &d.Text, &d.DocumentID, &acceptedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	if acceptedAt != nil && !acceptedAt.IsZero() {
		d.AcceptedAt = *acceptedAt
	}

	return &d, nil
}

func (r *sqlDisclaimerRepository) getCustomerDisclaimers(customerID string) ([]*client.Disclaimer, error) {
	query := `select disclaimer_id from disclaimers where deleted_at is null;`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, err
	}

	var out []*client.Disclaimer
	rows, err := stmt.Query()
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var disclaimerID string
		if err := rows.Scan(&disclaimerID); err != nil {
			return nil, err
		}
		disc, err := r.getCustomerDisclaimer(customerID, disclaimerID)
		if err != nil {
			return nil, err
		}
		out = append(out, disc)
	}
	return out, rows.Err()
}

func (r *sqlDisclaimerRepository) acceptDisclaimer(customerID, disclaimerID string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	query := `select disclaimer_id from disclaimers where disclaimer_id = ? and deleted_at is null limit 1;`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		tx.Rollback()
		return err
	}

	row := stmt.QueryRow(disclaimerID)
	var discID string
	if err := row.Scan(&discID); discID != disclaimerID || err != nil {
		stmt.Close()
		return fmt.Errorf("acceptDisclaimer: missing disclaimer: %v rollback=%v", err, tx.Rollback())
	}
	stmt.Close()

	// write the acceptance row now
	query = `insert into disclaimer_acceptances (disclaimer_id, customer_id, accepted_at) values (?, ?, ?);`
	stmt, err = tx.Prepare(query)
	if err != nil {
		tx.Rollback()
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(disclaimerID, customerID, time.Now())
	if err != nil {
		tx.Rollback()
		return err
	}
	return tx.Commit()
}

func (r *sqlDisclaimerRepository) insertDisclaimer(text, documentID string) (*client.Disclaimer, error) {
	query := `insert into disclaimers (disclaimer_id, text, document_id, created_at) values (?, ?, ?, ?);`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	disc := &client.Disclaimer{
		ID:         base.ID(),
		Text:       text,
		DocumentID: documentID,
	}
	_, err = stmt.Exec(disc.ID, disc.Text, disc.DocumentID, time.Now())
	return disc, err
}
