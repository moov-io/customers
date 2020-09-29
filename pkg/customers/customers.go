// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package customers

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

	"github.com/moov-io/customers/internal/usstates"
	"github.com/moov-io/customers/pkg/client"
	"github.com/moov-io/customers/pkg/model"
	"github.com/moov-io/customers/pkg/route"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

func AddCustomerRoutes(logger log.Logger, r *mux.Router, repo CustomerRepository, customerSSNStorage *ssnStorage, ofac *OFACSearcher) {
	r.Methods("GET").Path("/customers").HandlerFunc(searchCustomers(logger, repo))
	r.Methods("GET").Path("/customers/{customerID}").HandlerFunc(getCustomer(logger, repo))
	r.Methods("PUT").Path("/customers/{customerID}").HandlerFunc(updateCustomer(logger, repo, customerSSNStorage))
	r.Methods("DELETE").Path("/customers/{customerID}").HandlerFunc(deleteCustomer(logger, repo))
	r.Methods("POST").Path("/customers").HandlerFunc(createCustomer(logger, repo, customerSSNStorage, ofac))
	r.Methods("PUT").Path("/customers/{customerID}/metadata").HandlerFunc(replaceCustomerMetadata(logger, repo))
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

func getCustomer(logger log.Logger, repo CustomerRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)

		customerID, requestID := route.GetCustomerID(w, r), moovhttp.GetRequestID(r)
		if customerID == "" {
			return
		}

		respondWithCustomer(logger, w, customerID, requestID, repo)
	}
}

func deleteCustomer(logger log.Logger, repo CustomerRepository) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)

		customerID := route.GetCustomerID(w, r)
		if customerID == "" {
			return
		}

		err := repo.deleteCustomer(customerID)
		if err != nil {
			moovhttp.Problem(w, fmt.Errorf("deleting customer: %v", err))
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func respondWithCustomer(logger log.Logger, w http.ResponseWriter, customerID string, requestID string, repo CustomerRepository) {
	cust, err := repo.getCustomer(customerID)
	if err != nil {
		logger.Log("customers", fmt.Sprintf("getCustomer: lookup: %v", err), "requestID", requestID)
		moovhttp.Problem(w, err)
		return
	}
	if cust == nil {
		w.WriteHeader(http.StatusNotFound)
	} else {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(cust)
	}
}

// customerRequest holds the information for creating a Customer from the HTTP api
//
// TODO(adam): What GDPR implications does this information have? IIRC if any EU citizen uses
// this software we have to fully comply.
type customerRequest struct {
	CustomerID string                `json:"-"`
	FirstName  string                `json:"firstName"`
	MiddleName string                `json:"middleName"`
	LastName   string                `json:"lastName"`
	NickName   string                `json:"nickName"`
	Suffix     string                `json:"suffix"`
	Type       client.CustomerType   `json:"type"`
	BirthDate  model.YYYYMMDD        `json:"birthDate"`
	Status     client.CustomerStatus `json:"-"`
	Email      string                `json:"email"`
	SSN        string                `json:"SSN"`
	Phones     []phone               `json:"phones"`
	Addresses  []address             `json:"addresses"`
	Metadata   map[string]string     `json:"metadata"`
}

type phone struct {
	Number string `json:"number"`
	Type   string `json:"type"`
}

type address struct {
	Type       string `json:"type"`
	Address1   string `json:"address1"`
	Address2   string `json:"address2,omitempty"`
	City       string `json:"city"`
	State      string `json:"state"`
	PostalCode string `json:"postalCode"` // TODO(adam): validate against US postal codes
	Country    string `json:"country"`
}

func (add address) validate() error {
	if !usstates.Valid(add.State) {
		return fmt.Errorf("create customer: invalid state=%s", add.State)
	}
	return nil
}

func (req customerRequest) validate() error {
	if req.FirstName == "" || req.LastName == "" {
		return errors.New("create customer: empty name field(s)")
	}
	if err := validateCustomerType(req.Type); err != nil {
		return fmt.Errorf("create customer: %v", err)
	}
	if err := validateMetadata(req.Metadata); err != nil {
		return fmt.Errorf("create customer: %v", err)
	}
	for i := range req.Addresses {
		if err := req.Addresses[i].validate(); err != nil {
			return fmt.Errorf("address=%v validation failed: %v", req.Addresses[i], err)
		}
	}
	return nil
}

func validateCustomerType(t client.CustomerType) error {
	norm := func(t client.CustomerType) string {
		return strings.ToLower(string(t))
	}
	switch norm(t) {
	case norm(client.INDIVIDUAL), norm(client.BUSINESS):
		return nil
	}
	return fmt.Errorf("unknown type: %s", t)
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
	if req.CustomerID == "" {
		req.CustomerID = base.ID()
	}

	if req.Status == "" {
		req.Status = client.UNKNOWN
	}

	customer := &client.Customer{
		CustomerID: req.CustomerID,
		FirstName:  req.FirstName,
		MiddleName: req.MiddleName,
		LastName:   req.LastName,
		NickName:   req.NickName,
		Suffix:     req.Suffix,
		Type:       req.Type,
		BirthDate:  string(req.BirthDate),
		Email:      req.Email,
		Status:     req.Status,
		Metadata:   req.Metadata,
	}

	for i := range req.Phones {
		customer.Phones = append(customer.Phones, client.Phone{
			Number: req.Phones[i].Number,
			Type:   req.Phones[i].Type,
		})
	}
	for i := range req.Addresses {
		customer.Addresses = append(customer.Addresses, client.CustomerAddress{
			AddressID:  base.ID(),
			Address1:   req.Addresses[i].Address1,
			Address2:   req.Addresses[i].Address2,
			City:       req.Addresses[i].City,
			State:      req.Addresses[i].State,
			PostalCode: req.Addresses[i].PostalCode,
			Country:    req.Addresses[i].Country,
		})
	}
	if req.SSN != "" {
		ssn, err := storage.encryptRaw(customer.CustomerID, req.SSN)
		return customer, ssn, err
	}
	return customer, nil, nil
}

func createCustomer(logger log.Logger, repo CustomerRepository, customerSSNStorage *ssnStorage, ofac *OFACSearcher) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		requestID, organization := moovhttp.GetRequestID(r), route.GetOrganization(w, r)
		if organization == "" {
			return
		}

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
			logger.Log("customers", fmt.Sprintf("problem transforming request into Customer=%s: %v", cust.CustomerID, err), "requestID", requestID)
			moovhttp.Problem(w, err)
			return
		}
		if ssn != nil {
			err := customerSSNStorage.repo.saveCustomerSSN(ssn)
			if err != nil {
				logger.Log("customers", fmt.Sprintf("problem saving SSN for Customer=%s: %v", cust.CustomerID, err), "requestID", requestID)
				moovhttp.Problem(w, fmt.Errorf("saveCustomerSSN: %v", err))
				return
			}
		}
		if err := repo.createCustomer(cust, organization); err != nil {
			logger.Log("customers", fmt.Sprintf("createCustomer: %v", err), "requestID", requestID)
			moovhttp.Problem(w, err)
			return
		}
		if err := repo.replaceCustomerMetadata(cust.CustomerID, cust.Metadata); err != nil {
			logger.Log("customers", fmt.Sprintf("updating metadata for customer=%s failed: %v", cust.CustomerID, err), "requestID", requestID)
			moovhttp.Problem(w, err)
			return
		}

		// Perform an OFAC search with the Customer information
		if err := ofac.storeCustomerOFACSearch(cust, requestID); err != nil {
			logger.Log("customers", fmt.Sprintf("error with OFAC search for customer=%s: %v", cust.CustomerID, err), "requestID", requestID)
		}

		logger.Log("customers", fmt.Sprintf("created customer=%s", cust.CustomerID), "requestID", requestID)

		cust, err = repo.getCustomer(cust.CustomerID)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(cust)
	}
}

func updateCustomer(logger log.Logger, repo CustomerRepository, customerSSNStorage *ssnStorage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)
		w.Header().Set("Content-Type", "application/json; charset=utf-8")

		requestID, organization := moovhttp.GetRequestID(r), route.GetOrganization(w, r)
		if organization == "" {
			return
		}

		var req customerRequest
		req.CustomerID = route.GetCustomerID(w, r)
		if req.CustomerID == "" {
			return
		}

		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			moovhttp.Problem(w, err)
			return
		}
		if err := req.validate(); err != nil {
			logger.Log("customers", "error validating customer payload", "error", err, "requestID", requestID)
			moovhttp.Problem(w, err)
			return
		}

		cust, ssn, err := req.asCustomer(customerSSNStorage)
		if err != nil {
			logger.Log("customers", fmt.Sprintf("transforming request into Customer=%s: %v", cust.CustomerID, err), "requestID", requestID)
			moovhttp.Problem(w, err)
			return
		}
		if ssn != nil {
			err := customerSSNStorage.repo.saveCustomerSSN(ssn)
			if err != nil {
				logger.Log("customers", fmt.Sprintf("error saving SSN for Customer=%s: %v", cust.CustomerID, err), "requestID", requestID)
				moovhttp.Problem(w, fmt.Errorf("saving customer's SSN: %v", err))
				return
			}
		}
		if err := repo.updateCustomer(cust, organization); err != nil {
			logger.Log("customers", fmt.Sprintf("error updating customer: %v", err), "requestID", requestID)
			moovhttp.Problem(w, fmt.Errorf("updating customer: %v", err))
			return
		}

		if err := repo.replaceCustomerMetadata(cust.CustomerID, cust.Metadata); err != nil {
			logger.Log("customers", fmt.Sprintf("error updating metadata for customer=%s: %v", cust.CustomerID, err), "requestID", requestID)
			moovhttp.Problem(w, err)
			return
		}

		logger.Log("customers", fmt.Sprintf("updated customer=%s", cust.CustomerID), "requestID", requestID)
		cust, err = repo.getCustomer(cust.CustomerID)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(cust)
	}
}

type replaceMetadataRequest struct {
	Metadata map[string]string `json:"metadata"`
}

func replaceCustomerMetadata(logger log.Logger, repo CustomerRepository) http.HandlerFunc {
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
		customerID, requestID := route.GetCustomerID(w, r), moovhttp.GetRequestID(r)
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

type CustomerRepository interface {
	getCustomer(customerID string) (*client.Customer, error)
	createCustomer(c *client.Customer, organization string) error
	updateCustomer(c *client.Customer, organization string) error
	updateCustomerStatus(customerID string, status client.CustomerStatus, comment string) error
	deleteCustomer(customerID string) error

	searchCustomers(params searchParams) ([]*client.Customer, error)

	getCustomerMetadata(customerID string) (map[string]string, error)
	replaceCustomerMetadata(customerID string, metadata map[string]string) error

	addCustomerAddress(customerID string, address address) error
	updateCustomerAddress(customerID, addressID string, req updateCustomerAddressRequest) error
	deleteCustomerAddress(customerID string, addressID string) error

	getLatestCustomerOFACSearch(customerID string) (*ofacSearchResult, error)
	saveCustomerOFACSearch(customerID string, result ofacSearchResult) error
}

func NewCustomerRepo(logger log.Logger, db *sql.DB) CustomerRepository {
	return &sqlCustomerRepository{
		db:     db,
		logger: logger,
	}
}

type sqlCustomerRepository struct {
	db     *sql.DB
	logger log.Logger
}

func (r *sqlCustomerRepository) close() error {
	return r.db.Close()
}

func (r *sqlCustomerRepository) deleteCustomer(customerID string) error {
	query := `update customers set deleted_at = ? where customer_id = ? and deleted_at is null;`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(time.Now(), customerID)
	return err
}

func (r *sqlCustomerRepository) createCustomer(c *client.Customer, organization string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	// Insert customer record
	query := `insert into customers (customer_id, first_name, middle_name, last_name, nick_name, suffix, type, birth_date, status, email, created_at, last_modified, 
	organization)
values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`
	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	var birthDate *string
	if c.BirthDate != "" {
		birthDate = &c.BirthDate
	}

	now := time.Now()
	_, err = stmt.Exec(c.CustomerID, c.FirstName, c.MiddleName, c.LastName, c.NickName, c.Suffix, c.Type, birthDate, client.UNKNOWN, c.Email, now, now, organization)
	if err != nil {
		return fmt.Errorf("createCustomer: insert into customers err=%v | rollback=%v", err, tx.Rollback())
	}

	err = r.updatePhonesByCustomerID(tx, c.CustomerID, c.Phones)
	if err != nil {
		return fmt.Errorf("updating customer's phones: %v", err)
	}

	err = r.updateAddressesByCustomerID(tx, c.CustomerID, c.Addresses)
	if err != nil {
		return fmt.Errorf("updating customer's addresses: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("createCustomer: tx.Commit: %v", err)
	}
	return nil
}

func (r *sqlCustomerRepository) updateCustomer(c *client.Customer, organization string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	query := `update customers set first_name = ?, middle_name = ?, last_name = ?, nick_name = ?, suffix = ?, type = ?, birth_date = ?, status = ?, email =?, 
	last_modified = ?,
	organization = ? where customer_id = ? and deleted_at is null;`
	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now()
	res, err := stmt.Exec(c.FirstName, c.MiddleName, c.LastName, c.NickName, c.Suffix, c.Type, c.BirthDate, c.Status, c.Email, now, organization, c.CustomerID)
	if err != nil {
		return fmt.Errorf("updating customer: %v", err)
	}

	numRows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("getting rows affected: %v", err)
	}

	if numRows == 0 {
		return fmt.Errorf("no records to update with customer id=%s", c.CustomerID)
	}

	err = r.updatePhonesByCustomerID(tx, c.CustomerID, c.Phones)
	if err != nil {
		return fmt.Errorf("updating customer's phones: %v", err)
	}

	err = r.updateAddressesByCustomerID(tx, c.CustomerID, c.Addresses)
	if err != nil {
		return fmt.Errorf("updating customer's addresses: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("createCustomer: tx.Commit: %v", err)
	}
	return nil
}

func (r *sqlCustomerRepository) updatePhonesByCustomerID(tx *sql.Tx, customerID string, phones []client.Phone) error {
	query := `replace into customers_phones (customer_id, number, valid, type) values (?, ?, ?, ?);`
	stmt, err := tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("preparing tx update on customers_phones err=%v | rollback=%v", err, tx.Rollback())
	}
	defer stmt.Close()

	for _, phone := range phones {
		_, err := stmt.Exec(customerID, phone.Number, phone.Valid, phone.Type)
		if err != nil {
			return fmt.Errorf("executing update on customers_phones err=%v | rollback=%v", err, tx.Rollback())
		}
	}

	return nil
}

func (r *sqlCustomerRepository) updateAddressesByCustomerID(tx *sql.Tx, customerID string, addresses []client.CustomerAddress) error {
	query := `replace into customers_addresses(address_id, customer_id, type, address1, address2, city, state, postal_code, country, validated) values (?, ?, ?, ?, ?, ?, ?, ?, 
	?, ?);`
	stmt, err := tx.Prepare(query)
	if err != nil {
		return fmt.Errorf("preparing tx on customers_addresses err=%v | rollback=%v", err, tx.Rollback())
	}
	defer stmt.Close()

	for _, addr := range addresses {
		_, err := stmt.Exec(addr.AddressID, customerID, addr.Type, addr.Address1, addr.Address2, addr.City, addr.State, addr.PostalCode, addr.Country, addr.Validated)
		if err != nil {
			return fmt.Errorf("executing update on customers_addresses err=%v | rollback=%v", err, tx.Rollback())
		}
	}
	return nil
}

func (r *sqlCustomerRepository) getCustomer(customerID string) (*client.Customer, error) {
	query := `select first_name, middle_name, last_name, nick_name, suffix, type, birth_date, status, email, created_at, last_modified from customers where customer_id = ? and deleted_at is null limit 1;`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, err
	}

	row := stmt.QueryRow(customerID)

	var birthDate *string
	var cust client.Customer
	cust.CustomerID = customerID
	err = row.Scan(&cust.FirstName, &cust.MiddleName, &cust.LastName, &cust.NickName, &cust.Suffix, &cust.Type, &birthDate, &cust.Status, &cust.Email, &cust.CreatedAt,
		&cust.LastModified)
	stmt.Close()
	if err != nil && !strings.Contains(err.Error(), "no rows in result set") {
		return nil, fmt.Errorf("getCustomer: %v", err)
	}
	if cust.FirstName == "" {
		return nil, nil // not found
	}
	if birthDate != nil {
		cust.BirthDate = *birthDate
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

func (r *sqlCustomerRepository) readAddresses(customerID string) ([]client.CustomerAddress, error) {
	query := `select address_id, type, address1, address2, city, state, postal_code, country, validated from customers_addresses where customer_id = ? and deleted_at is null;`
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

	var adds []client.CustomerAddress
	for rows.Next() {
		var a client.CustomerAddress
		if err := rows.Scan(&a.AddressID, &a.Type, &a.Address1, &a.Address2, &a.City, &a.State, &a.PostalCode, &a.Country, &a.Validated); err != nil {
			return nil, fmt.Errorf("readAddresses: scan customers_addresses: err=%v", err)
		}
		adds = append(adds, a)
	}
	return adds, rows.Err()
}

func (r *sqlCustomerRepository) updateCustomerStatus(customerID string, status client.CustomerStatus, comment string) error {
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
	if _, err := stmt.Exec(status, customerID); err != nil {
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
	if _, err := stmt.Exec(customerID, status, comment, time.Now()); err != nil {
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
	query := `insert into customers_addresses (address_id, customer_id, type, address1, address2, city, state, postal_code, country, validated) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return fmt.Errorf("createCustomerAddress: prepare: %v", err)
	}
	defer stmt.Close()

	if _, err := stmt.Exec(base.ID(), customerID, req.Type, req.Address1, req.Address2, req.City, req.State, req.PostalCode, req.Country, false); err != nil {
		return fmt.Errorf("createCustomerAddress: exec: %v", err)
	}
	return nil
}

func (r *sqlCustomerRepository) updateCustomerAddress(customerID, addressID string, req updateCustomerAddressRequest) error {
	query := `update customers_addresses set type = ?, address1 = ?, address2 = ?, city = ?, state = ?, postal_code = ?, country = ?, 
	validated = ? where customer_id = ? and address_id = ? and deleted_at is null;`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return fmt.Errorf("updateCustomerAddress: prepare: %v", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		req.Type,
		req.Address1,
		req.Address2,
		req.City,
		req.State,
		req.PostalCode,
		req.Country,
		req.Validated,
		customerID,
		addressID)
	if err != nil {
		return fmt.Errorf("updateCustomerAddress: exec: %v", err)
	}
	return nil
}

func (r *sqlCustomerRepository) deleteCustomerAddress(customerID string, addressID string) error {
	query := `update customers_addresses set deleted_at = ? where customer_id = ? and address_id = ? and deleted_at is null;`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(time.Now(), customerID, addressID)
	return err
}

func (r *sqlCustomerRepository) getLatestCustomerOFACSearch(customerID string) (*ofacSearchResult, error) {
	query := `select entity_id, sdn_name, sdn_type, percentage_match, created_at from customer_ofac_searches where customer_id = ? order by created_at desc limit 1;`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("getLatestCustomerOFACSearch: prepare: %v", err)
	}
	defer stmt.Close()

	row := stmt.QueryRow(customerID)
	var res ofacSearchResult
	if err := row.Scan(&res.EntityID, &res.SDNName, &res.SDNType, &res.Match, &res.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // nothing found
		}
		return nil, fmt.Errorf("getLatestCustomerOFACSearch: scan: %v", err)
	}
	return &res, nil
}

func (r *sqlCustomerRepository) saveCustomerOFACSearch(customerID string, result ofacSearchResult) error {
	query := `insert into customer_ofac_searches (customer_id, entity_id, sdn_name, sdn_type, percentage_match, created_at) values (?, ?, ?, ?, ?, ?);`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return fmt.Errorf("saveCustomerOFACSearch: prepare: %v", err)
	}
	defer stmt.Close()

	if result.CreatedAt.IsZero() {
		result.CreatedAt = time.Now()
	}

	if _, err := stmt.Exec(customerID, result.EntityID, result.SDNName, result.SDNType, result.Match, result.CreatedAt); err != nil {
		return fmt.Errorf("saveCustomerOFACSearch: exec: %v", err)
	}
	return nil
}
