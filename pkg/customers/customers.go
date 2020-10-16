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

	"github.com/gorilla/mux"
	"github.com/moov-io/base/log"
)

func AddCustomerRoutes(logger log.Logger, r *mux.Router, repo CustomerRepository, customerSSNStorage *ssnStorage, ofac *OFACSearcher) {
	logger = logger.WithKeyValue("package", "customers")

	r.Methods("GET").Path("/customers").HandlerFunc(searchCustomers(logger, repo))
	r.Methods("GET").Path("/customers/{customerID}").HandlerFunc(getCustomer(logger, repo))
	r.Methods("PUT").Path("/customers/{customerID}").HandlerFunc(updateCustomer(logger, repo, customerSSNStorage))
	r.Methods("DELETE").Path("/customers/{customerID}").HandlerFunc(deleteCustomer(logger, repo))
	r.Methods("POST").Path("/customers").HandlerFunc(createCustomer(logger, repo, customerSSNStorage, ofac))
	r.Methods("PUT").Path("/customers/{customerID}/metadata").HandlerFunc(replaceCustomerMetadata(logger, repo))
	r.Methods("PUT").Path("/customers/{customerID}/status").HandlerFunc(updateCustomerStatus(logger, repo))
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

		organization := route.GetOrganization(w, r)

		respondWithCustomer(logger, w, customerID, organization, requestID, repo)
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

func respondWithCustomer(logger log.Logger, w http.ResponseWriter, customerID, organization string, requestID string, repo CustomerRepository) {
	cust, err := repo.GetCustomer(customerID, organization)
	if err != nil {
		logger.LogErrorF("getCustomer: lookup: %v", err)
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
	switch t := strings.ToLower(add.Type); t {
	case "primary", "secondary":
	default:
		return fmt.Errorf("unknown type: %s", t)
	}

	if !usstates.Valid(add.State) {
		return fmt.Errorf("create customer: invalid state=%s", add.State)
	}
	return nil
}

func (req customerRequest) validate() error {
	if req.FirstName == "" || req.LastName == "" {
		return errors.New("invalid customer fields: empty name field(s)")
	}
	if err := validateCustomerType(req.Type); err != nil {
		return fmt.Errorf("invalid customer type: %v", err)
	}
	if err := validateMetadata(req.Metadata); err != nil {
		return fmt.Errorf("invalid customer metadata: %v", err)
	}
	if err := validateAddresses(req.Addresses); err != nil {
		return fmt.Errorf("invalid customer addresses: %v", err)
	}

	return nil
}

func validateCustomerType(t client.CustomerType) error {
	norm := func(t client.CustomerType) string {
		return strings.ToLower(string(t))
	}
	switch norm(t) {
	case norm(client.CUSTOMERTYPE_INDIVIDUAL), norm(client.CUSTOMERTYPE_BUSINESS):
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

func validateAddresses(addrs []address) error {
	hasPrimaryAddr := false
	for _, addr := range addrs {
		if hasPrimaryAddr {
			return ErrAddressTypeDuplicate
		}

		if err := addr.validate(); err != nil {
			return fmt.Errorf("validating address: %v", err)
		}

		if addr.Type == "primary" {
			hasPrimaryAddr = true
		}
	}

	return nil
}

func (req customerRequest) asCustomer(storage *ssnStorage) (*client.Customer, *SSN, error) {
	if req.CustomerID == "" {
		req.CustomerID = base.ID()
	}

	if req.Status == "" {
		req.Status = client.CUSTOMERSTATUS_UNKNOWN
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
			logger.LogError("error validating new customer", err)
			moovhttp.Problem(w, err)
			return
		}

		cust, ssn, err := req.asCustomer(customerSSNStorage)
		if err != nil {
			logger.LogErrorF("problem transforming request into Customer=%s: %v", cust.CustomerID, err)
			moovhttp.Problem(w, err)
			return
		}
		if ssn != nil {
			err := customerSSNStorage.repo.saveCustomerSSN(ssn)
			if err != nil {
				logger.LogErrorF("problem saving SSN for Customer=%s: %v", cust.CustomerID, err)
				moovhttp.Problem(w, fmt.Errorf("saveCustomerSSN: %v", err))
				return
			}
		}
		if err := repo.CreateCustomer(cust, organization); err != nil {
			logger.LogErrorF("createCustomer: %v", err)
			moovhttp.Problem(w, err)
			return
		}
		if err := repo.replaceCustomerMetadata(cust.CustomerID, cust.Metadata); err != nil {
			logger.LogErrorF("updating metadata for customer=%s failed: %v", cust.CustomerID, err)
			moovhttp.Problem(w, err)
			return
		}

		// Perform an OFAC search with the Customer information
		if err := ofac.storeCustomerOFACSearch(cust, requestID); err != nil {
			logger.LogErrorF("error with OFAC search for customer=%s: %v", cust.CustomerID, err)
		}

		logger.Log(fmt.Sprintf("created customer=%s", cust.CustomerID))

		cust, err = repo.GetCustomer(cust.CustomerID, organization)
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

		organization := route.GetOrganization(w, r)
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
			logger.LogError("error validating customer payload", err)
			moovhttp.Problem(w, err)
			return
		}

		cust, ssn, err := req.asCustomer(customerSSNStorage)
		if err != nil {
			logger.LogErrorF("transforming request into Customer=%s: %v", cust.CustomerID, err)
			moovhttp.Problem(w, err)
			return
		}
		if ssn != nil {
			err := customerSSNStorage.repo.saveCustomerSSN(ssn)
			if err != nil {
				logger.LogErrorF("error saving SSN for Customer=%s: %v", cust.CustomerID, err)
				moovhttp.Problem(w, fmt.Errorf("saving customer's SSN: %v", err))
				return
			}
		}
		if err := repo.updateCustomer(cust, organization); err != nil {
			logger.LogErrorF("error updating customer: %v", err)
			moovhttp.Problem(w, fmt.Errorf("updating customer: %v", err))
			return
		}

		if err := repo.replaceCustomerMetadata(cust.CustomerID, cust.Metadata); err != nil {
			logger.LogErrorF("error updating metadata for customer=%s: %v", cust.CustomerID, err)
			moovhttp.Problem(w, err)
			return
		}

		logger.Log(fmt.Sprintf("updated customer=%s", cust.CustomerID))
		cust, err = repo.GetCustomer(cust.CustomerID, organization)
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
		organization := route.GetOrganization(w, r)
		customerID, requestID := route.GetCustomerID(w, r), moovhttp.GetRequestID(r)
		if customerID == "" {
			return
		}
		if err := repo.replaceCustomerMetadata(customerID, req.Metadata); err != nil {
			moovhttp.Problem(w, err)
			return
		}

		respondWithCustomer(logger, w, customerID, organization, requestID, repo)
	}
}

type CustomerRepository interface {
	GetCustomer(customerID, organization string) (*client.Customer, error)
	CreateCustomer(c *client.Customer, organization string) error
	updateCustomer(c *client.Customer, organization string) error
	updateCustomerStatus(customerID string, status client.CustomerStatus, comment string) error
	deleteCustomer(customerID string) error

	searchCustomers(params SearchParams) ([]*client.Customer, error)

	replaceCustomerMetadata(customerID string, metadata map[string]string) error

	addCustomerAddress(customerID string, address address) error
	updateCustomerAddress(customerID, addressID string, req updateCustomerAddressRequest) error
	deleteCustomerAddress(customerID string, addressID string) error

	getLatestCustomerOFACSearch(customerID, organization string) (*client.OfacSearch, error)
	saveCustomerOFACSearch(customerID string, result client.OfacSearch) error
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

func (r *sqlCustomerRepository) CreateCustomer(c *client.Customer, organization string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}

	// Insert customer record
	query := `insert into customers (customer_id, first_name, middle_name, last_name, nick_name, suffix, type, birth_date, status, email, created_at, last_modified, organization)
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
	_, err = stmt.Exec(c.CustomerID, c.FirstName, c.MiddleName, c.LastName, c.NickName, c.Suffix, c.Type, birthDate, client.CUSTOMERSTATUS_UNKNOWN, c.Email, now, now, organization)
	if err != nil {
		return fmt.Errorf("CreateCustomer: insert into customers err=%v | rollback=%v", err, tx.Rollback())
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
		return fmt.Errorf("CreateCustomer: tx.Commit: %v", err)
	}
	return nil
}

func (r *sqlCustomerRepository) updateCustomer(c *client.Customer, organization string) error {
	tx, err := r.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

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
		return fmt.Errorf("CreateCustomer: tx.Commit: %v", err)
	}
	return nil
}

func (r *sqlCustomerRepository) updatePhonesByCustomerID(tx *sql.Tx, customerID string, phones []client.Phone) error {
	deleteQuery := `delete from customers_phones where customer_id = ?`
	var args []interface{}
	args = append(args, customerID)
	if len(phones) > 0 {
		deleteQuery = fmt.Sprintf("%s and number not in (?%s)", deleteQuery, strings.Repeat(",?", len(phones)-1))
		for _, p := range phones {
			args = append(args, p.Number)
		}
	}
	deleteQuery = fmt.Sprintf("%s;", deleteQuery)

	stmt, err := tx.Prepare(deleteQuery)
	if err != nil {
		return fmt.Errorf("preparing query: %v", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(args...)
	if err != nil {
		return fmt.Errorf("executing query: %v", err)
	}

	replaceQuery := `replace into customers_phones (customer_id, number, valid, type) values (?, ?, ?, ?);`
	stmt, err = tx.Prepare(replaceQuery)
	if err != nil {
		return fmt.Errorf("preparing query: %v", err)
	}
	defer stmt.Close()

	for _, phone := range phones {
		_, err := stmt.Exec(customerID, phone.Number, phone.Valid, phone.Type)
		if err != nil {
			return fmt.Errorf("executing update on customer's phone: %v", err)
		}
	}

	return nil
}

func (r *sqlCustomerRepository) updateAddressesByCustomerID(tx *sql.Tx, customerID string, addresses []client.CustomerAddress) error {
	deleteQuery := `delete from customers_addresses where customer_id = ?`
	var args []interface{}
	args = append(args, customerID)
	if len(addresses) > 0 {
		deleteQuery = fmt.Sprintf("%s and address1 not in (?%s)", deleteQuery, strings.Repeat(",?", len(addresses)-1))
		for _, a := range addresses {
			args = append(args, a.Address1)
		}
	}
	deleteQuery = fmt.Sprintf("%s;", deleteQuery)

	stmt, err := tx.Prepare(deleteQuery)
	if err != nil {
		return fmt.Errorf("preparing query: %v", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(args...)
	if err != nil {
		panic(err)
	}

	replaceQuery := `replace into customers_addresses(address_id, customer_id, type, address1, address2, city, state, postal_code, country, validated) values (?, ?, ?, ?, ?, ?, ?, ?,
	?, ?);`
	stmt, err = tx.Prepare(replaceQuery)
	if err != nil {
		return fmt.Errorf("preparing query: %v", err)
	}
	defer stmt.Close()

	for _, addr := range addresses {
		_, err := stmt.Exec(addr.AddressID, customerID, addr.Type, addr.Address1, addr.Address2, addr.City, addr.State, addr.PostalCode, addr.Country, addr.Validated)
		if err != nil {
			return fmt.Errorf("executing query: %v", err)
		}
	}

	return nil
}

func (r *sqlCustomerRepository) GetCustomer(customerID, organization string) (*client.Customer, error) {
	custs, err := r.searchCustomers(SearchParams{
		Count:        1,
		CustomerIDs:  []string{customerID},
		Organization: organization,
	})
	if err != nil {
		return nil, fmt.Errorf("getting customer: %v", err)
	}

	if len(custs) == 0 {
		return nil, errors.New("customer not found")
	}

	return custs[0], nil
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

func (r *sqlCustomerRepository) getLatestCustomerOFACSearch(customerID, organization string) (*client.OfacSearch, error) {
	query := `select entity_id, blocked, sdn_name, sdn_type, percentage_match, cos.created_at 
from customer_ofac_searches as cos
inner join customers as c on c.customer_id = cos.customer_id
where cos.customer_id = ? and c.organization = ? order by cos.created_at desc limit 1;`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("getLatestCustomerOFACSearch: prepare: %v", err)
	}
	defer stmt.Close()

	row := stmt.QueryRow(customerID, organization)
	var res client.OfacSearch
	if err := row.Scan(&res.EntityID, &res.Blocked, &res.SdnName, &res.SdnType, &res.Match, &res.CreatedAt); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // nothing found
		}
		return nil, fmt.Errorf("getLatestCustomerOFACSearch: scan: %v", err)
	}
	return &res, nil
}

func (r *sqlCustomerRepository) saveCustomerOFACSearch(customerID string, result client.OfacSearch) error {
	query := `insert into customer_ofac_searches (customer_id, blocked, entity_id, sdn_name, sdn_type, percentage_match, created_at) values (?, ?, ?, ?, ?, ?, ?);`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return fmt.Errorf("saveCustomerOFACSearch: prepare: %v", err)
	}
	defer stmt.Close()

	if result.CreatedAt.IsZero() {
		result.CreatedAt = time.Now()
	}

	if _, err := stmt.Exec(customerID, result.Blocked, result.EntityID, result.SdnName, result.SdnType, result.Match, result.CreatedAt); err != nil {
		return fmt.Errorf("saveCustomerOFACSearch: exec: %v", err)
	}
	return nil
}
