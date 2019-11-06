// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/moov-io/base"
	moovhttp "github.com/moov-io/base/http"
	"github.com/moov-io/customers"
	client "github.com/moov-io/customers/client"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

var (
	errNoCustomerID = errors.New("no Customer ID found")
)

func addCustomerRoutes(logger log.Logger, r *mux.Router, repo customerRepository, customerSSNStorage *ssnStorage, ofac *ofacSearcher) {
	r.Methods("GET").Path("/customers/{customerID}").HandlerFunc(getCustomer(logger, repo))
	r.Methods("POST").Path("/customers").HandlerFunc(createCustomer(logger, repo, customerSSNStorage, ofac))
	r.Methods("PUT").Path("/customers/{customerID}/metadata").HandlerFunc(replaceCustomerMetadata(logger, repo))
	r.Methods("POST").Path("/customers/{customerID}/address").HandlerFunc(addCustomerAddress(logger, repo))
}

func getCustomerID(w http.ResponseWriter, r *http.Request) string {
	v, ok := mux.Vars(r)["customerID"]
	if !ok || v == "" {
		moovhttp.Problem(w, errNoCustomerID)
		return ""
	}
	return v
}

// formatCustomerName returns a Customer's name joined as one string. It accounts for
// first, middle, last and suffix. Each field is whitespace trimmed.
func formatCustomerName(c *client.Customer) string {
	if c == nil {
		return ""
	}
	out := strings.TrimSpace(c.FirstName)
	if c.MiddleName != "" {
		out += " " + strings.TrimSpace(c.MiddleName)
	}
	out = strings.TrimSpace(out + " " + strings.TrimSpace(c.LastName))
	if c.Suffix != "" {
		out += " " + c.Suffix
	}
	return strings.TrimSpace(out)
}

func getCustomer(logger log.Logger, repo customerRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = wrapResponseWriter(logger, w, r)

		customerID, requestID := getCustomerID(w, r), moovhttp.GetRequestID(r)
		if customerID == "" {
			return
		}

		respondWithCustomer(logger, w, customerID, requestID, repo)
	}
}

func respondWithCustomer(logger log.Logger, w http.ResponseWriter, customerID string, requestID string, repo customerRepository) {
	cust, err := repo.getCustomer(customerID)
	if err != nil {
		logger.Log("customers", fmt.Sprintf("getCustomer: lookup: %v", err), "requestID", requestID)
		moovhttp.Problem(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(cust)
}

// customerRequest holds the information for creating a Customer from the HTTP api
//
// TODO(adam): What GDPR implications does this information have? IIRC if any EU citizen uses
// this software we have to fully comply.
type customerRequest struct {
	FirstName  string            `json:"firstName"`
	MiddleName string            `json:"middleName"`
	LastName   string            `json:"lastName"`
	NickName   string            `json:"nickName"`
	Suffix     string            `json:"suffix"`
	BirthDate  time.Time         `json:"birthDate"`
	Email      string            `json:"email"`
	SSN        string            `json:"SSN"`
	Phones     []phone           `json:"phones"`
	Addresses  []address         `json:"addresses"`
	Metadata   map[string]string `json:"metadata"`
}

type phone struct {
	Number string `json:"number"`
	Type   string `json:"type"`
}

type address struct {
	Address1   string `json:"address1"`
	Address2   string `json:"address2,omitempty"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postalCode"`
	Country    string `json:"country"`
}

func (req customerRequest) validate() error {
	if req.FirstName == "" || req.LastName == "" {
		return errors.New("create customer: empty name field(s)")
	}
	if err := validateMetadata(req.Metadata); err != nil {
		return fmt.Errorf("create customer: %v", err)
	}
	return nil
}

func validateMetadata(meta map[string]string) error {
	// both are arbitrary limits, open an issue if this needs bumped
	if len(meta) > 100 {
		return errors.New("metadata is limited to 100 entries")
	}
	for k, v := range meta {
		if utf8.RuneCountInString(v) > 1000 {
			return fmt.Errorf("metadata key %s value is too long", k)
		}
	}
	return nil
}

func (req customerRequest) asCustomer(storage *ssnStorage) (*client.Customer, *SSN, error) {
	customer := &client.Customer{
		ID:         base.ID(),
		FirstName:  req.FirstName,
		MiddleName: req.MiddleName,
		LastName:   req.LastName,
		NickName:   req.NickName,
		Suffix:     req.Suffix,
		BirthDate:  req.BirthDate,
		Email:      req.Email,
		Status:     customers.None.String(),
		Metadata:   req.Metadata,
	}
	for i := range req.Phones {
		customer.Phones = append(customer.Phones, client.Phone{
			Number: req.Phones[i].Number,
			Type:   req.Phones[i].Type,
		})
	}
	for i := range req.Addresses {
		customer.Addresses = append(customer.Addresses, client.Address{
			ID:         base.ID(),
			Address1:   req.Addresses[i].Address1,
			Address2:   req.Addresses[i].Address2,
			City:       req.Addresses[i].City,
			State:      req.Addresses[i].State,
			PostalCode: req.Addresses[i].PostalCode,
			Country:    req.Addresses[i].Country,
			Active:     true,
		})
	}
	if req.SSN != "" {
		ssn, err := storage.encryptRaw(customer.ID, req.SSN)
		return customer, ssn, err
	}
	return customer, nil, nil
}

func createCustomer(logger log.Logger, repo customerRepository, customerSSNStorage *ssnStorage, ofac *ofacSearcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = wrapResponseWriter(logger, w, r)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		requestID := moovhttp.GetRequestID(r)

		var req customerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			moovhttp.Problem(w, err)
			return
		}
		if err := req.validate(); err != nil {
			logger.Log("customers", "error validating new customer", "error", err, "requestID", requestID)
			moovhttp.Problem(w, err)
			return
		}

		cust, ssn, err := req.asCustomer(customerSSNStorage)
		if err != nil {
			logger.Log("customers", fmt.Sprintf("problem transforming request into Customer=%s: %v", cust.ID, err), "requestID", requestID)
			moovhttp.Problem(w, err)
			return
		}
		if ssn != nil {
			err := customerSSNStorage.repo.saveCustomerSSN(ssn)
			if err != nil {
				logger.Log("customers", fmt.Sprintf("problem saving SSN for Customer=%s: %v", cust.ID, err), "requestID", requestID)
				moovhttp.Problem(w, fmt.Errorf("saveCustomerSSN: %v", err))
				return
			}
		}
		if err := repo.createCustomer(cust); err != nil {
			if requestID != "" {
				logger.Log("customers", fmt.Sprintf("createCustomer: %v", err), "requestID", requestID)
			}
			moovhttp.Problem(w, err)
			return
		}
		if err := repo.replaceCustomerMetadata(cust.ID, cust.Metadata); err != nil {
			logger.Log("customers", fmt.Sprintf("updating metadata for customer=%s failed: %v", cust.ID, err), "requestID", requestID)
			moovhttp.Problem(w, err)
			return
		}

		// Try an OFAC search with the Customer information
		go func(logger log.Logger, cust *client.Customer, requestID string) {
			if err := ofac.storeCustomerOFACSearch(cust, requestID); err != nil {
				logger.Log("customers", fmt.Sprintf("error with OFAC search for customer=%s: %v", cust.ID, err), "requestID", requestID)
			}
		}(logger, cust, requestID)

		logger.Log("customers", fmt.Sprintf("created customer=%s", cust.ID), "requestID", requestID)

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(cust)
	}
}

type replaceMetadataRequest struct {
	Metadata map[string]string `json:"metadata"`
}

func replaceCustomerMetadata(logger log.Logger, repo customerRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req replaceMetadataRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			moovhttp.Problem(w, err)
			return
		}
		if err := validateMetadata(req.Metadata); err != nil {
			moovhttp.Problem(w, err)
			return
		}
		customerID, requestID := getCustomerID(w, r), moovhttp.GetRequestID(r)
		if customerID == "" {
			return
		}
		if err := repo.replaceCustomerMetadata(customerID, req.Metadata); err != nil {
			moovhttp.Problem(w, err)
			return
		}

		respondWithCustomer(logger, w, customerID, requestID, repo)
	}
}

func addCustomerAddress(logger log.Logger, repo customerRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		customerID, requestID := getCustomerID(w, r), moovhttp.GetRequestID(r)
		if customerID == "" {
			return
		}

		var req address
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			moovhttp.Problem(w, err)
			return
		}

		if err := repo.addCustomerAddress(customerID, req); err != nil {
			moovhttp.Problem(w, err)
			return
		}

		logger.Log("customers", fmt.Sprintf("added address for customer=%s", customerID), "requestID", requestID)

		respondWithCustomer(logger, w, customerID, requestID, repo)
	}
}

type customerRepository interface {
	getCustomer(customerID string) (*client.Customer, error)
	createCustomer(c *client.Customer) error
	updateCustomerStatus(customerID string, status customers.Status, comment string) error

	getCustomerMetadata(customerID string) (map[string]string, error)
	replaceCustomerMetadata(customerID string, metadata map[string]string) error

	addCustomerAddress(customerID string, address address) error
	updateCustomerAddress(customerID, addressID string, _type string, validated bool) error

	getLatestCustomerOFACSearch(customerID string) (*ofacSearchResult, error)
	saveCustomerOFACSearch(customerID string, result ofacSearchResult) error
}

type sqlCustomerRepository struct {
	db     *sql.DB
	logger log.Logger
}

func (r *sqlCustomerRepository) close() error {
	return r.db.Close()
}

func (r *sqlCustomerRepository) createCustomer(c *client.Customer) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	// Insert customer record
	query := `insert into customers (customer_id, first_name, middle_name, last_name, nick_name, suffix, birth_date, status, email, created_at, last_modified)
values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`
	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	now := time.Now()
	_, err = stmt.Exec(c.ID, c.FirstName, c.MiddleName, c.LastName, c.NickName, c.Suffix, c.BirthDate, c.Status, c.Email, now, now)
	if err != nil {
		return fmt.Errorf("createCustomer: insert into customers err=%v | rollback=%v", err, tx.Rollback())
	}
	stmt.Close()

	// Insert customer phone numbers
	query = `insert or replace into customers_phones (customer_id, number, valid, type) values (?, ?, ?, ?);`
	stmt, err = tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("createCustomer: insert into customers_phones err=%v | rollback=%v", err, tx.Rollback())
	}
	for i := range c.Phones {
		_, err := stmt.Exec(c.ID, c.Phones[i].Number, c.Phones[i].Valid, c.Phones[i].Type)
		if err != nil {
			stmt.Close()
			return fmt.Errorf("createCustomer: customers_phones exec err=%v | rollback=%v", err, tx.Rollback())
		}
	}
	stmt.Close()

	// Insert customer addresses
	query = `insert or replace into customers_addresses(address_id, customer_id, type, address1, address2, city, state, postal_code, country, validated, active) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`
	stmt, err = tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("createCustomer: insert into customers_addresses err=%v | rollback=%v", err, tx.Rollback())
	}
	for i := range c.Addresses {
		_, err := stmt.Exec(c.Addresses[i].ID, c.ID, c.Addresses[i].Type, c.Addresses[i].Address1, c.Addresses[i].Address2, c.Addresses[i].City, c.Addresses[i].State, c.Addresses[i].PostalCode, c.Addresses[i].Country, c.Addresses[i].Validated, c.Addresses[i].Active)
		if err != nil {
			stmt.Close()
			return fmt.Errorf("createCustomer: customers_addresses exec err=%v | rollback=%v", err, tx.Rollback())
		}
	}
	stmt.Close()

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("createCustomer: tx.Commit: %v", err)
	}
	return nil
}

func (r *sqlCustomerRepository) getCustomer(customerID string) (*client.Customer, error) {
	query := `select first_name, middle_name, last_name, nick_name, suffix, birth_date, status, email, created_at, last_modified from customers where customer_id = ? and deleted_at is null limit 1;`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, err
	}

	row := stmt.QueryRow(customerID)

	var cust client.Customer
	cust.ID = customerID
	err = row.Scan(&cust.FirstName, &cust.MiddleName, &cust.LastName, &cust.NickName, &cust.Suffix, &cust.BirthDate, &cust.Status, &cust.Email, &cust.CreatedAt, &cust.LastModified)
	stmt.Close()
	if err != nil && !strings.Contains(err.Error(), "no rows in result set") {
		return nil, fmt.Errorf("getCustomer: %v", err)
	}
	if cust.FirstName == "" {
		return nil, nil // not found
	}

	phones, err := r.readPhones(customerID)
	if err != nil {
		return nil, fmt.Errorf("getCustomer: readPhones: %v", err)
	}
	cust.Phones = phones

	addresses, err := r.readAddresses(customerID)
	if err != nil {
		return nil, fmt.Errorf("getCustomer: readAddresses: %v", err)
	}
	cust.Addresses = addresses

	metadata, err := r.getCustomerMetadata(customerID)
	if err != nil {
		return nil, fmt.Errorf("getCustomer: getCustomerMetadata: %v", err)
	}
	cust.Metadata = metadata

	return &cust, nil
}

func (r *sqlCustomerRepository) readPhones(customerID string) ([]client.Phone, error) {
	query := `select number, valid, type from customers_phones where customer_id = ?;`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("getCustomer: prepare customers_phones: err=%v", err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(customerID)
	if err != nil {
		return nil, fmt.Errorf("getCustomer: query customers_phones: err=%v", err)
	}
	defer rows.Close()

	var phones []client.Phone
	for rows.Next() {
		var p client.Phone
		if err := rows.Scan(&p.Number, &p.Valid, &p.Type); err != nil {
			return nil, fmt.Errorf("getCustomer: scan customers_phones: err=%v", err)
		}
		phones = append(phones, p)
	}
	return phones, rows.Err()
}

func (r *sqlCustomerRepository) readAddresses(customerID string) ([]client.Address, error) {
	query := `select address_id, type, address1, address2, city, state, postal_code, country, validated, active from customers_addresses where customer_id = ?;`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("readAddresses: prepare customers_addresses: err=%v", err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(customerID)
	if err != nil {
		return nil, fmt.Errorf("readAddresses: query customers_addresses: err=%v", err)
	}
	defer rows.Close()

	var adds []client.Address
	for rows.Next() {
		var a client.Address
		if err := rows.Scan(&a.ID, &a.Type, &a.Address1, &a.Address2, &a.City, &a.State, &a.PostalCode, &a.Country, &a.Validated, &a.Active); err != nil {
			return nil, fmt.Errorf("readAddresses: scan customers_addresses: err=%v", err)
		}
		adds = append(adds, a)
	}
	return adds, rows.Err()
}

func (r *sqlCustomerRepository) updateCustomerStatus(customerID string, status customers.Status, comment string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("updateCustomerStatus: tx begin: %v", err)
	}

	// update 'customers' table
	query := `update customers set status = ? where customer_id = ?;`
	stmt, err := tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("updateCustomerStatus: update customers prepare: %v", err)
	}
	if _, err := stmt.Exec(status.String(), customerID); err != nil {
		stmt.Close()
		return fmt.Errorf("updateCustomerStatus: update customers exec: %v", err)
	}
	stmt.Close()

	// update 'customer_status_updates' table
	query = `insert into customer_status_updates (customer_id, future_status, comment, changed_at) values (?, ?, ?, ?);`
	stmt, err = tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("updateCustomerStatus: insert status prepare: %v", err)
	}
	defer stmt.Close()
	if _, err := stmt.Exec(customerID, status.String(), comment, time.Now()); err != nil {
		return fmt.Errorf("updateCustomerStatus: insert status exec: %v", err)
	}
	return tx.Commit()
}

func (r *sqlCustomerRepository) getCustomerMetadata(customerID string) (map[string]string, error) {
	out := make(map[string]string)

	query := `select meta_key, meta_value from customer_metadata where customer_id = ?;`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return out, fmt.Errorf("getCustomerMetadata: prepare: %v", err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(customerID)
	if err != nil {
		return out, fmt.Errorf("getCustomerMetadata: query: %v", err)
	}
	defer rows.Close()

	for rows.Next() {
		key, value := "", ""
		if err := rows.Scan(&key, &value); err != nil {
			return out, fmt.Errorf("getCustomerMetadata: scan: %v", err)
		}
		out[key] = value
	}
	return out, nil
}

func (r *sqlCustomerRepository) replaceCustomerMetadata(customerID string, metadata map[string]string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return fmt.Errorf("replaceCustomerMetadata: tx begin: %v", err)
	}

	// Delete each existing k/v pair
	query := `delete from customer_metadata where customer_id = ?;`
	stmt, err := tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("replaceCustomerMetadata: delete prepare: %v", err)
	}
	if _, err := stmt.Exec(customerID); err != nil {
		stmt.Close()
		return fmt.Errorf("replaceCustomerMetadata: delete exec: %v", err)
	}
	stmt.Close()

	// Insert each k/v pair
	query = `insert into customer_metadata (customer_id, meta_key, meta_value) values (?, ?, ?);`
	stmt, err = tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("replaceCustomerMetadata: insert prepare: %v", err)
	}
	defer stmt.Close()
	for k, v := range metadata {
		if _, err := stmt.Exec(customerID, k, v); err != nil {
			return fmt.Errorf("replaceCustomerMetadata: insert %s: %v", k, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("replaceCustomerMetadata: commit: %v", err)
	}
	return nil
}

func (r *sqlCustomerRepository) addCustomerAddress(customerID string, req address) error {
	query := `insert into customers_addresses (address_id, customer_id, type, address1, address2, city, state, postal_code, country, validated, active) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return fmt.Errorf("addCustomerAddress: prepare: %v", err)
	}
	defer stmt.Close()

	if _, err := stmt.Exec(base.ID(), customerID, "Secondary", req.Address1, req.Address2, req.City, req.State, req.PostalCode, req.Country, false, true); err != nil {
		return fmt.Errorf("addCustomerAddress: exec: %v", err)
	}
	return nil
}

func (r *sqlCustomerRepository) updateCustomerAddress(customerID, addressID string, _type string, validated bool) error {
	query := `update customers_addresses set type = ?, validated = ? where customer_id = ? and address_id = ?;`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return fmt.Errorf("updateCustomerAddress: prepare: %v", err)
	}
	defer stmt.Close()

	if _, err := stmt.Exec(_type, validated, customerID, addressID); err != nil {
		return fmt.Errorf("updateCustomerAddress: exec: %v", err)
	}
	return nil
}

func (r *sqlCustomerRepository) getLatestCustomerOFACSearch(customerID string) (*ofacSearchResult, error) {
	query := `select entity_id, sdn_name, sdn_type, match from customer_ofac_searches where customer_id = ? order by created_at desc limit 1;`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("getLatestCustomerOFACSearch: prepare: %v", err)
	}
	defer stmt.Close()

	row := stmt.QueryRow(customerID)
	var res ofacSearchResult
	if err := row.Scan(&res.EntityId, &res.SDNName, &res.SDNType, &res.Match); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // nothing found
		}
		return nil, fmt.Errorf("getLatestCustomerOFACSearch: scan: %v", err)
	}
	return &res, nil
}

func (r *sqlCustomerRepository) saveCustomerOFACSearch(customerID string, result ofacSearchResult) error {
	query := `insert into customer_ofac_searches (customer_id, entity_id, sdn_name, sdn_type, match, created_at) values (?, ?, ?, ?, ?, ?);`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return fmt.Errorf("saveCustomerOFACSearch: prepare: %v", err)
	}
	defer stmt.Close()

	if _, err := stmt.Exec(customerID, result.EntityId, result.SDNName, result.SDNType, result.Match, time.Now()); err != nil {
		return fmt.Errorf("saveCustomerOFACSearch: exec: %v", err)
	}
	return nil
}
