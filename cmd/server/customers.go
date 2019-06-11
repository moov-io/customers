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
	client "github.com/moov-io/customers/client"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

var (
	errNoCustomerId = errors.New("no Customer ID found")
)

type CustomerStatus string

const (
	CustomerStatusDeceased       = "deceased"
	CustomerStatusRejected       = "rejected"
	CustomerStatusNone           = "none"
	CustomerStatusReviewRequired = "reviewrequired"
	CustomerStatusKYC            = "kyc"
	CustomerStatusOFAC           = "ofac"
	CustomerStatusCIP            = "cip"
)

func (cs CustomerStatus) validate() error {
	switch cs {
	case CustomerStatusDeceased, CustomerStatusRejected:
		return nil
	case CustomerStatusReviewRequired, CustomerStatusNone:
		return nil
	case CustomerStatusKYC, CustomerStatusOFAC, CustomerStatusCIP:
		return nil
	default:
		return fmt.Errorf("CustomerStatus(%s) is invalid", cs)
	}
}

func (cs *CustomerStatus) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	*cs = CustomerStatus(strings.TrimSpace(strings.ToLower(s)))
	if err := cs.validate(); err != nil {
		return err
	}
	return nil
}

func addCustomerRoutes(logger log.Logger, r *mux.Router, repo customerRepository, ofac *ofacSearcher) {
	r.Methods("GET").Path("/customers/{customerId}").HandlerFunc(getCustomer(logger, repo))
	r.Methods("POST").Path("/customers").HandlerFunc(createCustomer(logger, repo, ofac))
	r.Methods("PUT").Path("/customers/{customerId}/metadata").HandlerFunc(replaceCustomerMetadata(logger, repo))
	r.Methods("POST").Path("/customers/{customerId}/address").HandlerFunc(addCustomerAddress(logger, repo))
}

func getCustomerId(w http.ResponseWriter, r *http.Request) string {
	v, ok := mux.Vars(r)["customerId"]
	if !ok || v == "" {
		moovhttp.Problem(w, errNoCustomerId)
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

		customerId, requestId := getCustomerId(w, r), moovhttp.GetRequestId(r)
		if customerId == "" {
			return
		}

		respondWithCustomer(logger, w, customerId, requestId, repo)
	}
}

func respondWithCustomer(logger log.Logger, w http.ResponseWriter, customerId string, requestId string, repo customerRepository) {
	cust, err := repo.getCustomer(customerId)
	if err != nil {
		logger.Log("customers", fmt.Sprintf("getCustomer: lookup: %v", err), "requestId", requestId)
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
	if req.Email == "" {
		return errors.New("create customer: empty email")
	}
	if len(req.Phones) == 0 {
		return errors.New("create customer: phone array is required")
	}
	if len(req.Addresses) == 0 {
		return errors.New("create customer: address array is required")
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

func (req customerRequest) asCustomer() client.Customer {
	// TODO(adam): How do we store off SSN (and wipe from models)
	customer := client.Customer{
		Id:         base.ID(),
		FirstName:  req.FirstName,
		MiddleName: req.MiddleName,
		LastName:   req.LastName,
		NickName:   req.NickName,
		Suffix:     req.Suffix,
		BirthDate:  req.BirthDate,
		Email:      req.Email,
		Status:     CustomerStatusNone,
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
			Id:         base.ID(),
			Address1:   req.Addresses[i].Address1,
			Address2:   req.Addresses[i].Address2,
			City:       req.Addresses[i].City,
			State:      req.Addresses[i].State,
			PostalCode: req.Addresses[i].PostalCode,
			Country:    req.Addresses[i].Country,
			Active:     true,
		})
	}
	return customer
}

func createCustomer(logger log.Logger, repo customerRepository, ofac *ofacSearcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = wrapResponseWriter(logger, w, r)

		var req customerRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			moovhttp.Problem(w, err)
			return
		}
		if err := req.validate(); err != nil {
			moovhttp.Problem(w, err)
			return
		}
		requestId := moovhttp.GetRequestId(r)

		cust, err := repo.createCustomer(req)
		if err != nil {
			if requestId != "" {
				logger.Log("customers", fmt.Sprintf("createCustomer: %v", err), "requestId", requestId)
			}
			moovhttp.Problem(w, err)
			return
		}
		if err := repo.replaceCustomerMetadata(cust.Id, cust.Metadata); err != nil {
			logger.Log("customers", fmt.Sprintf("updating metadata for customer=%s failed: %v", cust.Id, err), "requestId", requestId)
			moovhttp.Problem(w, err)
			return
		}

		// Try an OFAC search with the Customer information
		go func(logger log.Logger, cust *client.Customer, requestId string) {
			if err := ofac.storeCustomerOFACSearch(cust, requestId); err != nil {
				logger.Log("customers", fmt.Sprintf("error with OFAC search for customer=%s: %v", cust.Id, err), "requestId", requestId)
			}
		}(logger, cust, requestId)

		logger.Log("customers", fmt.Sprintf("created customer=%s", cust.Id), "requestId", requestId)

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(cust)
	}
}

type repalceMetadataRequest struct {
	Metadata map[string]string `json:"metadata"`
}

func replaceCustomerMetadata(logger log.Logger, repo customerRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req repalceMetadataRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			moovhttp.Problem(w, err)
			return
		}
		if err := validateMetadata(req.Metadata); err != nil {
			moovhttp.Problem(w, err)
			return
		}
		customerId, requestId := getCustomerId(w, r), moovhttp.GetRequestId(r)
		if customerId == "" {
			return
		}
		if err := repo.replaceCustomerMetadata(customerId, req.Metadata); err != nil {
			moovhttp.Problem(w, err)
			return
		}

		respondWithCustomer(logger, w, customerId, requestId, repo)
	}
}

func addCustomerAddress(logger log.Logger, repo customerRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		customerId, requestId := getCustomerId(w, r), moovhttp.GetRequestId(r)
		if customerId == "" {
			return
		}

		var req address
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			moovhttp.Problem(w, err)
			return
		}

		if err := repo.addCustomerAddress(customerId, req); err != nil {
			moovhttp.Problem(w, err)
			return
		}

		logger.Log("customers", fmt.Sprintf("added address for customer=%s", customerId), "requestId", requestId)

		respondWithCustomer(logger, w, customerId, requestId, repo)
	}
}

type customerRepository interface {
	getCustomer(customerId string) (*client.Customer, error)
	createCustomer(req customerRequest) (*client.Customer, error)
	updateCustomerStatus(customerId string, status CustomerStatus, comment string) error

	getCustomerMetadata(customerId string) (map[string]string, error)
	replaceCustomerMetadata(customerId string, metadata map[string]string) error

	addCustomerAddress(customerId string, address address) error
	updateCustomerAddress(customerId, addressId string, _type string, validated bool) error

	saveCustomerOFACSearch(customerId string, result ofacSearchResult) error
}

type sqliteCustomerRepository struct {
	db *sql.DB
}

func (r *sqliteCustomerRepository) close() error {
	return r.db.Close()
}

func (r *sqliteCustomerRepository) createCustomer(req customerRequest) (*client.Customer, error) {
	c := req.asCustomer()

	tx, err := r.db.Begin()
	if err != nil {
		return nil, err
	}

	// Insert customer record
	query := `insert into customers (customer_id, first_name, middle_name, last_name, nick_name, suffix, birthdate, status, email, created_at, last_modified)
values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`
	stmt, err := tx.Prepare(query)
	if err != nil {
		return nil, err
	}
	now := time.Now()
	_, err = stmt.Exec(c.Id, c.FirstName, c.MiddleName, c.LastName, c.NickName, c.Suffix, c.BirthDate, c.Status, c.Email, now, now)
	if err != nil {
		return nil, fmt.Errorf("createCustomer: insert into customers err=%v | rollback=%v", err, tx.Rollback())
	}
	stmt.Close()

	// Insert customer phone numbers
	query = `insert or replace into customers_phones (customer_id, number, valid, type) values (?, ?, ?, ?);`
	stmt, err = tx.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("createCustomer: insert into customers_phones err=%v | rollback=%v", err, tx.Rollback())
	}
	for i := range c.Phones {
		_, err := stmt.Exec(c.Id, c.Phones[i].Number, c.Phones[i].Valid, c.Phones[i].Type)
		if err != nil {
			stmt.Close()
			return nil, fmt.Errorf("createCustomer: customers_phones exec err=%v | rollback=%v", err, tx.Rollback())
		}
	}
	stmt.Close()

	// Insert customer addresses
	query = `insert or replace into customers_addresses(address_id, customer_id, type, address1, address2, city, state, postal_code, country, validated, active) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`
	stmt, err = tx.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("createCustomer: insert into customers_addresses err=%v | rollback=%v", err, tx.Rollback())
	}
	for i := range c.Addresses {
		_, err := stmt.Exec(c.Addresses[i].Id, c.Id, c.Addresses[i].Type, c.Addresses[i].Address1, c.Addresses[i].Address2, c.Addresses[i].City, c.Addresses[i].State, c.Addresses[i].PostalCode, c.Addresses[i].Country, c.Addresses[i].Validated, c.Addresses[i].Active)
		if err != nil {
			stmt.Close()
			return nil, fmt.Errorf("createCustomer: customers_addresses exec err=%v | rollback=%v", err, tx.Rollback())
		}
	}
	stmt.Close()

	if err := tx.Commit(); err != nil {
		return nil, fmt.Errorf("createCustomer: tx.Commit: %v", err)
	}
	return &c, nil
}

func (r *sqliteCustomerRepository) getCustomer(customerId string) (*client.Customer, error) {
	query := `select first_name, middle_name, last_name, nick_name, suffix, birthdate, status, email, created_at, last_modified from customers where customer_id = ? and deleted_at is null limit 1;`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, err
	}

	row := stmt.QueryRow(customerId)

	var cust client.Customer
	cust.Id = customerId
	err = row.Scan(&cust.FirstName, &cust.MiddleName, &cust.LastName, &cust.NickName, &cust.Suffix, &cust.BirthDate, &cust.Status, &cust.Email, &cust.CreatedAt, &cust.LastModified)
	stmt.Close()
	if err != nil && !strings.Contains(err.Error(), "no rows in result set") {
		return nil, fmt.Errorf("getCustomer: %v", err)
	}
	if cust.FirstName == "" {
		return nil, nil // not found
	}

	phones, err := r.readPhones(customerId)
	if err != nil {
		return nil, fmt.Errorf("getCustomer: readPhones: %v", err)
	}
	cust.Phones = phones

	addresses, err := r.readAddresses(customerId)
	if err != nil {
		return nil, fmt.Errorf("getCustomer: readAddresses: %v", err)
	}
	cust.Addresses = addresses

	metadata, err := r.getCustomerMetadata(customerId)
	if err != nil {
		return nil, fmt.Errorf("getCustomer: getCustomerMetadata: %v", err)
	}
	cust.Metadata = metadata

	return &cust, nil
}

func (r *sqliteCustomerRepository) readPhones(customerId string) ([]client.Phone, error) {
	query := `select number, valid, type from customers_phones where customer_id = ?;`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("getCustomer: prepare customers_phones: err=%v", err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(customerId)
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

func (r *sqliteCustomerRepository) readAddresses(customerId string) ([]client.Address, error) {
	query := `select address_id, type, address1, address2, city, state, postal_code, country, validated, active from customers_addresses where customer_id = ?;`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("readAddresses: prepare customers_addresses: err=%v", err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(customerId)
	if err != nil {
		return nil, fmt.Errorf("readAddresses: query customers_addresses: err=%v", err)
	}
	defer rows.Close()

	var adds []client.Address
	for rows.Next() {
		var a client.Address
		if err := rows.Scan(&a.Id, &a.Type, &a.Address1, &a.Address2, &a.City, &a.State, &a.PostalCode, &a.Country, &a.Validated, &a.Active); err != nil {
			return nil, fmt.Errorf("readAddresses: scan customers_addresses: err=%v", err)
		}
		adds = append(adds, a)
	}
	return adds, rows.Err()
}

func (r *sqliteCustomerRepository) updateCustomerStatus(customerId string, status CustomerStatus, comment string) error {
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
	if _, err := stmt.Exec(status, customerId); err != nil {
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
	if _, err := stmt.Exec(customerId, status, comment, time.Now()); err != nil {
		return fmt.Errorf("updateCustomerStatus: insert status exec: %v", err)
	}
	return tx.Commit()
}

func (r *sqliteCustomerRepository) getCustomerMetadata(customerId string) (map[string]string, error) {
	out := make(map[string]string)

	query := `select key, value from customer_metadata where customer_id = ?;`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return out, fmt.Errorf("getCustomerMetadata: prepare: %v", err)
	}
	defer stmt.Close()

	rows, err := stmt.Query(customerId)
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

func (r *sqliteCustomerRepository) replaceCustomerMetadata(customerId string, metadata map[string]string) error {
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
	if _, err := stmt.Exec(customerId); err != nil {
		stmt.Close()
		return fmt.Errorf("replaceCustomerMetadata: delete exec: %v", err)
	}
	stmt.Close()

	// Insert each k/v pair
	query = `insert into customer_metadata (customer_id, key, value) values (?, ?, ?);`
	stmt, err = tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("replaceCustomerMetadata: insert prepare: %v", err)
	}
	defer stmt.Close()
	for k, v := range metadata {
		if _, err := stmt.Exec(customerId, k, v); err != nil {
			return fmt.Errorf("replaceCustomerMetadata: insert %s: %v", k, err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("replaceCustomerMetadata: commit: %v", err)
	}
	return nil
}

func (r *sqliteCustomerRepository) addCustomerAddress(customerId string, req address) error {
	query := `insert into customers_addresses (address_id, customer_id, type, address1, address2, city, state, postal_code, country, validated, active) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return fmt.Errorf("addCustomerAddress: prepare: %v", err)
	}
	defer stmt.Close()

	if _, err := stmt.Exec(base.ID(), customerId, "Secondary", req.Address1, req.Address2, req.City, req.State, req.PostalCode, req.Country, false, true); err != nil {
		return fmt.Errorf("addCustomerAddress: exec: %v", err)
	}
	return nil
}

func (r *sqliteCustomerRepository) updateCustomerAddress(customerId, addressId string, _type string, validated bool) error {
	query := `update customers_addresses set type = ?, validated = ? where customer_id = ? and address_id = ?;`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return fmt.Errorf("updateCustomerAddress: prepare: %v", err)
	}
	defer stmt.Close()

	if _, err := stmt.Exec(_type, validated, customerId, addressId); err != nil {
		return fmt.Errorf("updateCustomerAddress: exec: %v", err)
	}
	return nil
}

func (r *sqliteCustomerRepository) saveCustomerOFACSearch(customerId string, result ofacSearchResult) error {
	query := `insert into customer_ofac_searches (customer_id, entity_id, sdn_name, sdn_type, match, created_at) values (?, ?, ?, ?, ?, ?);`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return fmt.Errorf("saveCustomerOFACSearch: prepare: %v", err)
	}
	defer stmt.Close()

	if _, err := stmt.Exec(customerId, result.entityId, result.sdnName, result.sdnType, result.match, time.Now()); err != nil {
		return fmt.Errorf("saveCustomerOFACSearch: exec: %v", err)
	}
	return nil
}
