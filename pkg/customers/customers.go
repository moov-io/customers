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
	logger = logger.Set("package", "customers")

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
		if organization == "" {
			return
		}

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
		logger.LogErrorf("getCustomer: lookup: %v", err)
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
	CustomerID              string                `json:"-"`
	FirstName               string                `json:"firstName"`
	MiddleName              string                `json:"middleName"`
	LastName                string                `json:"lastName"`
	NickName                string                `json:"nickName"`
	Suffix                  string                `json:"suffix"`
	Type                    client.CustomerType   `json:"type"`
	BusinessName            string                `json:"businessName"`
	DoingBusinessAs         string                `json:"doingBusinessAs"`
	BusinessType            client.BusinessType   `json:"businessType"`
	EIN                     string                `json:"EIN"`
	DUNS                    string                `json:"DUNS"`
	SICCode                 client.SicCode        `json:"sicCode"`
	NAICSCode               client.NaicsCode      `json:"naicsCode"`
	BirthDate               model.YYYYMMDD        `json:"birthDate"`
	Status                  client.CustomerStatus `json:"-"`
	Email                   string                `json:"email"`
	Website                 string                `json:"website"`
	DateBusinessEstablished string                `json:"dateBusinessEstablished"`
	SSN                     string                `json:"SSN"`
	Phones                  []phone               `json:"phones"`
	Addresses               []address             `json:"addresses"`
	Representatives         []representative      `json:"representatives"`
	Metadata                map[string]string     `json:"metadata"`
}

type phone struct {
	Number    string           `json:"number"`
	Type      client.PhoneType `json:"type"`
	OwnerType client.OwnerType `json:"ownerType"`
}

func (p *phone) validate() error {
	p.Type = client.PhoneType(strings.ToLower(string(p.Type)))
	p.OwnerType = client.OwnerType(strings.ToLower(string(p.OwnerType)))

	switch p.Type {
	case client.PHONETYPE_HOME, client.PHONETYPE_MOBILE, client.PHONETYPE_WORK:
	default:
		return fmt.Errorf("unknown type: %s", p.Type)
	}

	return nil
}

type address struct {
	Type       client.AddressType `json:"type"`
	OwnerType  client.OwnerType   `json:"ownerType"`
	Address1   string             `json:"address1"`
	Address2   string             `json:"address2,omitempty"`
	City       string             `json:"city"`
	State      string             `json:"state"`
	PostalCode string             `json:"postalCode"` // TODO(adam): validate against US postal codes
	Country    string             `json:"country"`
}

func (add *address) validate() error {
	add.Type = client.AddressType(strings.ToLower(string(add.Type)))
	add.OwnerType = client.OwnerType(strings.ToLower(string(add.OwnerType)))

	switch add.Type {
	case client.ADDRESSTYPE_PRIMARY, client.ADDRESSTYPE_SECONDARY:
	default:
		return fmt.Errorf("unknown type: %s", add.Type)
	}

	if !usstates.Valid(add.State) {
		return fmt.Errorf("create customer: invalid state=%s", add.State)
	}
	return nil
}

type representative struct {
	FirstName string    `json:"firstName"`
	LastName  string    `json:"lastName"`
	JobTitle  string    `json:"jobTitle,omitempty"`
	BirthDate string    `json:"birthDate,omitempty"`
	Addresses []address `json:"addresses,omitempty"`
	Phones    []phone   `json:"phones,omitempty"`
}

func (rep *representative) validate() error {
	if rep.FirstName == "" || rep.LastName == "" {
		return errors.New("invalid customer representative fields: empty name field(s)")
	}
	return nil
}

func (req customerRequest) validate() error {
	isIndividualOrSoleProprietor := req.Type == client.CUSTOMERTYPE_INDIVIDUAL || req.BusinessType == client.BUSINESSTYPE_SOLE_PROPRIETOR
	if isIndividualOrSoleProprietor && (req.FirstName == "" || req.LastName == "") {
		return errors.New("invalid customer fields: empty name field(s)")
	} else if req.Type == client.CUSTOMERTYPE_BUSINESS && req.BusinessName == "" {
		return errors.New("invalid customer fields: empty business name")
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
	if err := validatePhones(req.Phones); err != nil {
		return fmt.Errorf("invalid customer phone: %v", err)
	}
	if err := validateRepresentatives(req.Representatives); err != nil {
		return fmt.Errorf("invalid customer representative: %v", err)
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
		if hasPrimaryAddr && addr.Type == client.ADDRESSTYPE_PRIMARY {
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

func validatePhones(phones []phone) error {
	for _, p := range phones {
		if err := p.validate(); err != nil {
			return err
		}
	}

	return nil
}

func validateRepresentatives(representatives []representative) error {
	for _, r := range representatives {
		if err := r.validate(); err != nil {
			return err
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
		CustomerID:              req.CustomerID,
		FirstName:               req.FirstName,
		MiddleName:              req.MiddleName,
		LastName:                req.LastName,
		NickName:                req.NickName,
		Suffix:                  req.Suffix,
		Type:                    req.Type,
		BusinessName:            req.BusinessName,
		DoingBusinessAs:         req.DoingBusinessAs,
		BusinessType:            req.BusinessType,
		EIN:                     req.EIN,
		DUNS:                    req.DUNS,
		SICCode:                 req.SICCode,
		NAICSCode:               req.NAICSCode,
		BirthDate:               string(req.BirthDate),
		Email:                   req.Email,
		Website:                 req.Website,
		DateBusinessEstablished: req.DateBusinessEstablished,
		Status:                  req.Status,
		Metadata:                req.Metadata,
	}

	for i := range req.Phones {
		customer.Phones = append(customer.Phones, client.Phone{
			Number:    req.Phones[i].Number,
			Type:      req.Phones[i].Type,
			OwnerType: client.OWNERTYPE_CUSTOMER,
		})
	}
	for i := range req.Addresses {
		customer.Addresses = append(customer.Addresses, client.Address{
			AddressID:  base.ID(),
			Address1:   req.Addresses[i].Address1,
			Address2:   req.Addresses[i].Address2,
			City:       req.Addresses[i].City,
			State:      req.Addresses[i].State,
			PostalCode: req.Addresses[i].PostalCode,
			Country:    req.Addresses[i].Country,
			Type:       req.Addresses[i].Type,
			OwnerType:  client.OWNERTYPE_CUSTOMER,
		})
	}
	for i := range req.Representatives {
		custRep := client.Representative{
			RepresentativeID: base.ID(),
			FirstName:        req.Representatives[i].FirstName,
			LastName:         req.Representatives[i].LastName,
			JobTitle:         req.Representatives[i].JobTitle,
			BirthDate:        req.Representatives[i].BirthDate,
		}

		for j := range req.Representatives[i].Addresses {
			custRep.Addresses = append(custRep.Addresses, client.Address{
				AddressID:  base.ID(),
				Address1:   req.Representatives[i].Addresses[j].Address1,
				Address2:   req.Representatives[i].Addresses[j].Address2,
				City:       req.Representatives[i].Addresses[j].City,
				State:      req.Representatives[i].Addresses[j].State,
				PostalCode: req.Representatives[i].Addresses[j].PostalCode,
				Country:    req.Representatives[i].Addresses[j].Country,
				Type:       req.Representatives[i].Addresses[j].Type,
				OwnerType:  client.OWNERTYPE_REPRESENTATIVE,
			})
		}
		for j := range req.Representatives[i].Phones {
			custRep.Phones = append(custRep.Phones, client.Phone{
				Number:    req.Representatives[i].Phones[j].Number,
				Type:      req.Representatives[i].Phones[j].Type,
				OwnerType: client.OWNERTYPE_REPRESENTATIVE,
			})
		}
		customer.Representatives = append(customer.Representatives, custRep)
	}
	if req.SSN != "" {
		ssn, err := storage.encryptRaw(customer.CustomerID, client.OWNERTYPE_CUSTOMER, req.SSN)
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
			logger.LogErrorf("error validating new customer: %v", err)
			moovhttp.Problem(w, err)
			return
		}

		cust, ssn, err := req.asCustomer(customerSSNStorage)
		if err != nil {
			logger.LogErrorf("problem transforming request into Customer=%s: %v", cust.CustomerID, err)
			moovhttp.Problem(w, err)
			return
		}
		if ssn != nil {
			err := customerSSNStorage.repo.saveSSN(ssn)
			if err != nil {
				logger.LogErrorf("problem saving SSN for Customer=%s: %v", cust.CustomerID, err)
				moovhttp.Problem(w, fmt.Errorf("saveCustomerSSN: %v", err))
				return
			}
		}
		if err := repo.CreateCustomer(cust, organization); err != nil {
			logger.LogErrorf("createCustomer: %v", err)
			moovhttp.Problem(w, err)
			return
		}
		if err := repo.replaceCustomerMetadata(cust.CustomerID, cust.Metadata); err != nil {
			logger.LogErrorf("updating metadata for customer=%s failed: %v", cust.CustomerID, err)
			moovhttp.Problem(w, err)
			return
		}

		// Perform an OFAC search with the Customer information
		if err := ofac.storeCustomerOFACSearch(cust, requestID); err != nil {
			logger.LogErrorf("error with OFAC search for customer=%s: %v", cust.CustomerID, err)
		}

		logger.Logf("created customer=%s", cust.CustomerID)

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
			logger.LogErrorf("error validating customer payload: %v", err)
			moovhttp.Problem(w, err)
			return
		}

		cust, ssn, err := req.asCustomer(customerSSNStorage)
		if err != nil {
			logger.LogErrorf("transforming request into Customer=%s: %v", cust.CustomerID, err)
			moovhttp.Problem(w, err)
			return
		}
		if ssn != nil {
			err := customerSSNStorage.repo.saveSSN(ssn)
			if err != nil {
				logger.LogErrorf("error saving SSN for Customer=%s: %v", cust.CustomerID, err)
				moovhttp.Problem(w, fmt.Errorf("saving customer's SSN: %v", err))
				return
			}
		}
		if err := repo.updateCustomer(cust, organization); err != nil {
			logger.LogErrorf("error updating customer: %v", err)
			moovhttp.Problem(w, fmt.Errorf("updating customer: %v", err))
			return
		}

		if err := repo.replaceCustomerMetadata(cust.CustomerID, cust.Metadata); err != nil {
			logger.LogErrorf("error updating metadata for customer=%s: %v", cust.CustomerID, err)
			moovhttp.Problem(w, err)
			return
		}

		logger.Logf("updated customer=%s", cust.CustomerID)
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
		if organization == "" {
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

	GetRepresentative(representativeID string) (*client.Representative, error)
	CreateRepresentative(c *client.Representative, customerID string) error
	updateRepresentative(c *client.Representative, customerID string) error
	deleteRepresentative(representativeID string) error

	addAddress(ownerID string, ownerType client.OwnerType, address address) error
	updateAddress(ownerID, addressID string, ownerType client.OwnerType, req updateAddressRequest) error
	deleteAddress(ownerID string, ownerType client.OwnerType, addressID string) error

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
	query := `insert into customers (customer_id, first_name, middle_name, last_name, nick_name, suffix, type, business_name, doing_business_as, business_type, ein, duns, sic_code, naics_code, birth_date, status, email, website, date_business_established, created_at, last_modified, organization)
values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`
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
	customerType := c.Type
	if customerType == "" {
		customerType = client.CUSTOMERTYPE_INDIVIDUAL
	}
	_, err = stmt.Exec(c.CustomerID, c.FirstName, c.MiddleName, c.LastName, c.NickName, c.Suffix, customerType, c.BusinessName, c.DoingBusinessAs, c.BusinessType, c.EIN, c.DUNS, c.SICCode, c.NAICSCode, birthDate, client.CUSTOMERSTATUS_UNKNOWN, c.Email, c.Website, c.DateBusinessEstablished, now, now, organization)
	if err != nil {
		return fmt.Errorf("CreateCustomer: insert into customers err=%v | rollback=%v", err, tx.Rollback())
	}

	err = r.updatePhonesByOwnerID(tx, c.CustomerID, client.OWNERTYPE_CUSTOMER, c.Phones)
	if err != nil {
		return fmt.Errorf("updating customer's phones: %v", err)
	}

	err = r.updateAddressesByOwnerID(tx, c.CustomerID, client.OWNERTYPE_CUSTOMER, c.Addresses)
	if err != nil {
		return fmt.Errorf("updating customer's addresses: %v", err)
	}

	err = r.updateRepresentativesByCustomerID(tx, c.CustomerID, c.Representatives)
	if err != nil {
		return fmt.Errorf("updating customer's representatives: %v", err)
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

	query := `update customers set first_name = ?, middle_name = ?, last_name = ?, nick_name = ?, suffix = ?, type = ?, business_name = ?, doing_business_as = ?, business_type = ?, ein = ?, duns = ?, sic_code = ?, naics_code = ?, birth_date = ?, status = ?, email =?,
	website = ?, date_business_established = ?, last_modified = ?,
	organization = ? where customer_id = ? and deleted_at is null;`
	stmt, err := tx.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	now := time.Now()
	res, err := stmt.Exec(c.FirstName, c.MiddleName, c.LastName, c.NickName, c.Suffix, c.Type, c.BusinessName, c.DoingBusinessAs, c.BusinessType, c.EIN, c.DUNS, c.SICCode, c.NAICSCode, c.BirthDate, c.Status, c.Email, c.Website, c.DateBusinessEstablished, now, organization, c.CustomerID)
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

	err = r.updatePhonesByOwnerID(tx, c.CustomerID, client.OWNERTYPE_CUSTOMER, c.Phones)
	if err != nil {
		return fmt.Errorf("updating customer's phones: %v", err)
	}

	err = r.updateAddressesByOwnerID(tx, c.CustomerID, client.OWNERTYPE_CUSTOMER, c.Addresses)
	if err != nil {
		return fmt.Errorf("updating customer's addresses: %v", err)
	}

	err = r.updateRepresentativesByCustomerID(tx, c.CustomerID, c.Representatives)
	if err != nil {
		return fmt.Errorf("updating customer's representatives: %v", err)
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("CreateCustomer: tx.Commit: %v", err)
	}
	return nil
}

func (r *sqlCustomerRepository) updatePhonesByOwnerID(tx *sql.Tx, ownerID string, ownerType client.OwnerType, phones []client.Phone) error {
	deleteQuery := `delete from phones where owner_id = ? and owner_type = ?`
	var args []interface{}
	args = append(args, ownerID, ownerType)
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

	replaceQuery := `replace into phones (owner_id, owner_type, number, valid, type) values (?, ?, ?, ?, ?);`
	stmt, err = tx.Prepare(replaceQuery)
	if err != nil {
		return fmt.Errorf("preparing query: %v", err)
	}
	defer stmt.Close()

	for _, phone := range phones {
		_, err := stmt.Exec(ownerID, string(ownerType), phone.Number, phone.Valid, phone.Type)
		if err != nil {
			return fmt.Errorf("executing update on customer's phone: %v", err)
		}
	}

	return nil
}

func (r *sqlCustomerRepository) updateAddressesByOwnerID(tx *sql.Tx, ownerID string, ownerType client.OwnerType, addresses []client.Address) error {
	deleteQuery := `delete from addresses where owner_id = ? and owner_type = ?`
	var args []interface{}
	args = append(args, ownerID, ownerType)
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

	replaceQuery := `replace into addresses(address_id, owner_id, owner_type, type, address1, address2, city, state, postal_code, country, validated) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`
	stmt, err = tx.Prepare(replaceQuery)
	if err != nil {
		return fmt.Errorf("preparing query: %v", err)
	}
	defer stmt.Close()

	for _, addr := range addresses {
		_, err := stmt.Exec(addr.AddressID, ownerID, string(ownerType), addr.Type, addr.Address1, addr.Address2, addr.City, addr.State, addr.PostalCode, addr.Country, addr.Validated)
		if err != nil {
			return fmt.Errorf("executing query: %v", err)
		}
	}

	return nil
}

func (r *sqlCustomerRepository) updateRepresentativesByCustomerID(tx *sql.Tx, customerID string, representatives []client.Representative) error {
	deleteQuery := `delete from representatives where customer_id = ?;`

	stmt, err := tx.Prepare(deleteQuery)
	if err != nil {
		return fmt.Errorf("preparing query: %v", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(customerID)
	if err != nil {
		panic(err)
	}

	replaceQuery := `replace into representatives(representative_id, customer_id, first_name, last_name, job_title, birth_date) values (?, ?, ?, ?, ?, ?);`
	stmt, err = tx.Prepare(replaceQuery)
	if err != nil {
		return fmt.Errorf("preparing query: %v", err)
	}
	defer stmt.Close()

	for _, rep := range representatives {
		_, err := stmt.Exec(rep.RepresentativeID, customerID, rep.FirstName, rep.LastName, rep.JobTitle, rep.BirthDate)
		if err != nil {
			return fmt.Errorf("executing query: %v", err)
		}

		err = r.updatePhonesByOwnerID(tx, rep.RepresentativeID, client.OWNERTYPE_REPRESENTATIVE, rep.Phones)
		if err != nil {
			return fmt.Errorf("updating customer representative's phones: %v", err)
		}

		err = r.updateAddressesByOwnerID(tx, rep.RepresentativeID, client.OWNERTYPE_REPRESENTATIVE, rep.Addresses)
		if err != nil {
			return fmt.Errorf("updating customer representative's addresses: %v", err)
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

func (r *sqlCustomerRepository) addAddress(ownerID string, ownerType client.OwnerType, req address) error {
	query := `insert into addresses (address_id, owner_id, owner_type, type, address1, address2, city, state, postal_code, country, validated) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return fmt.Errorf("addAddress: prepare: %v", err)
	}
	defer stmt.Close()

	if _, err := stmt.Exec(base.ID(), ownerID, string(ownerType), req.Type, req.Address1, req.Address2, req.City, req.State, req.PostalCode, req.Country, false); err != nil {
		return fmt.Errorf("addAddress: exec: %v", err)
	}
	return nil
}

func (r *sqlCustomerRepository) updateAddress(ownerID, addressID string, ownerType client.OwnerType, req updateAddressRequest) error {
	query := `update addresses set type = ?, address1 = ?, address2 = ?, city = ?, state = ?, postal_code = ?, country = ?,
	validated = ? where owner_id = ? and owner_type = ? and address_id = ? and deleted_at is null;`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return fmt.Errorf("updateAddress: prepare: %v", err)
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
		ownerID,
		string(ownerType),
		addressID)
	if err != nil {
		return fmt.Errorf("updateAddress: exec: %v", err)
	}
	return nil
}

func (r *sqlCustomerRepository) deleteAddress(ownerID string, ownerType client.OwnerType, addressID string) error {
	query := `update addresses set deleted_at = ? where owner_id = ? and owner_type = ? and address_id = ? and deleted_at is null;`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(time.Now(), ownerID, string(ownerType), addressID)
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
