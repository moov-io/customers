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

	"github.com/moov-io/base"
	moovhttp "github.com/moov-io/base/http"
	client "github.com/moov-io/customers/client"

	"github.com/go-kit/kit/log"
	"github.com/gorilla/mux"
)

var (
	errNoCustomerId = errors.New("no Customer ID found")
)

func addCustomerRoutes(logger log.Logger, r *mux.Router, repo customerRepository) {
	r.Methods("GET").Path("/customers/{customerId}").HandlerFunc(getCustomer(logger, repo))
	r.Methods("POST").Path("/customers").HandlerFunc(createCustomer(logger, repo))
}

func getCustomerId(w http.ResponseWriter, r *http.Request) string {
	v, ok := mux.Vars(r)["customerId"]
	if !ok || v == "" {
		moovhttp.Problem(w, errNoCustomerId)
		return ""
	}
	return v
}

func getCustomer(logger log.Logger, repo customerRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = wrapResponseWriter(logger, w, r)

		cust, err := repo.getCustomer(getCustomerId(w, r))
		if err != nil {
			if requestId := moovhttp.GetRequestId(r); requestId != "" {
				logger.Log("accounts", fmt.Sprintf("%v", err), "requestId", requestId)
			}
			moovhttp.Problem(w, err)
			return
		}

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(cust)
	}
}

type customerRequest struct {
	FirstName  string    `json:"firstName"`
	MiddleName string    `json:"middleName"`
	LastName   string    `json:"lastName"`
	NickName   string    `json:"nickName"`
	Suffix     string    `json:"suffix"`
	BirthDate  time.Time `json:"birthDate"`
	Email      string    `json:"email"`
	SSN        string    `json:"SSN"`
	Phones     []phone   `json:"phones"`
	Addresses  []address `json:"addresses"`
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
		Status:     "Applied",
	}
	for i := range req.Phones {
		customer.Phones = append(customer.Phones, client.Phone{
			Number: req.Phones[i].Number,
			Type:   req.Phones[i].Type,
		})
	}
	for i := range req.Addresses {
		customer.Addresses = append(customer.Addresses, client.Address{
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

func createCustomer(logger log.Logger, repo customerRepository) http.HandlerFunc {
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
				logger.Log("accounts", fmt.Sprintf("%v", err), "requestId", requestId)
			}
			moovhttp.Problem(w, err)
			return
		}

		logger.Log("customers", fmt.Sprintf("created customer=%s", cust.Id), "requestId", requestId)

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(cust)
	}
}

type customerRepository interface {
	createCustomer(req customerRequest) (*client.Customer, error)
	getCustomer(customerId string) (*client.Customer, error)
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
	query = `insert or replace into customers_addresses(customer_id, type, address1, address2, city, state, postal_code, country, validated, active) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`
	stmt, err = tx.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("createCustomer: insert into customers_addresses err=%v | rollback=%v", err, tx.Rollback())
	}
	for i := range c.Addresses {
		_, err := stmt.Exec(c.Id, c.Addresses[i].Type, c.Addresses[i].Address1, c.Addresses[i].Address2, c.Addresses[i].City, c.Addresses[i].State, c.Addresses[i].PostalCode, c.Addresses[i].Country, c.Addresses[i].Validated, c.Addresses[i].Active)
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
	// TODO(adam): read all DB fields once we handle all in the request
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
	query := `select type, address1, address2, city, state, postal_code, country, validated, active from customers_addresses where customer_id = ?;`
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
		if err := rows.Scan(&a.Type, &a.Address1, &a.Address2, &a.City, &a.State, &a.PostalCode, &a.Country, &a.Validated, &a.Active); err != nil {
			return nil, fmt.Errorf("readAddresses: scan customers_addresses: err=%v", err)
		}
		adds = append(adds, a)
	}
	return adds, rows.Err()
}
