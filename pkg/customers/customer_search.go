// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package customers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	moovhttp "github.com/moov-io/base/http"

	"github.com/moov-io/base/log"

	"github.com/moov-io/customers/pkg/client"
	"github.com/moov-io/customers/pkg/model"
	"github.com/moov-io/customers/pkg/route"
)

func searchCustomers(logger log.Logger, repo CustomerRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)

		organization := route.GetOrganization(w, r)
		if organization == "" {
			return
		}

		params, err := parseSearchParams(r)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}
		params.Organization = organization

		customers, err := repo.searchCustomers(params)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}

		logger.Logf("found %d customers in search", len(customers))

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(customers)
	}
}

type SearchParams struct {
	Organization string
	Query        string
	Email        string
	Status       string
	Type         string
	Skip         int64
	Count        int64
	CustomerIDs  []string
}

func parseSearchParams(r *http.Request) (SearchParams, error) {
	queryParams := r.URL.Query()
	getQueryParam := func(key string) string {
		return strings.ToLower(strings.TrimSpace(queryParams.Get(key)))
	}
	params := SearchParams{
		Query:  getQueryParam("query"),
		Email:  getQueryParam("email"),
		Status: getQueryParam("status"),
		Type:   getQueryParam("type"),
	}
	customerIDsInput := getQueryParam("customerIDs")
	if customerIDsInput != "" {
		params.CustomerIDs = strings.Split(customerIDsInput, ",")
	}

	skip, count, exists, err := moovhttp.GetSkipAndCount(r)
	if exists && err != nil {
		return params, err
	}

	params.Skip = int64(skip)
	params.Count = int64(count)

	return params, nil
}

func (r *sqlCustomerRepository) searchCustomers(params SearchParams) ([]*client.Customer, error) {
	query, args := buildSearchQuery(params)

	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	customers := make([]*client.Customer, 0)
	rows, err := stmt.Query(args...)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var c client.Customer
		var birthDate *time.Time
		err := rows.Scan(
			&c.CustomerID,
			&c.FirstName,
			&c.MiddleName,
			&c.LastName,
			&c.NickName,
			&c.Suffix,
			&c.Type,
			&c.BusinessName,
			&c.DoingBusinessAs,
			&c.BusinessType,
			&c.EIN,
			&c.DUNS,
			&c.SICCode,
			&c.NAICSCode,
			&birthDate,
			&c.Status,
			&c.Email,
			&c.Website,
			&c.DateBusinessEstablished,
			&c.CreatedAt,
			&c.LastModified,
		)
		if err != nil {
			return nil, err
		}
		if birthDate != nil {
			c.BirthDate = birthDate.Format(model.YYYYMMDD_Format)
		}
		customers = append(customers, &c)
	}

	if len(customers) == 0 {
		return customers, nil
	}

	var customerIDs []string
	for _, c := range customers {
		customerIDs = append(customerIDs, c.CustomerID)
	}

	phonesByCustomerID, err := r.GetPhones(customerIDs, client.OWNERTYPE_CUSTOMER)
	if err != nil {
		return nil, fmt.Errorf("fetching customer phones: %v", err)
	}
	addressesByCustomerID, err := r.GetAddresses(customerIDs, client.OWNERTYPE_CUSTOMER)
	if err != nil {
		return nil, fmt.Errorf("fetching customer addresses: %v", err)
	}
	representativesByCustomerID, err := r.getRepresentatives(customerIDs)
	if err != nil {
		return nil, fmt.Errorf("fetching customer representatives: %v", err)
	}
	metadataByCustomerID, err := r.getMetadata(customerIDs)
	if err != nil {
		return nil, fmt.Errorf("fetching customer metadata: %v", err)
	}

	for _, c := range customers {
		c.Phones = phonesByCustomerID[c.CustomerID]
		c.Addresses = addressesByCustomerID[c.CustomerID]
		c.Representatives = representativesByCustomerID[c.CustomerID]
		c.Metadata = metadataByCustomerID[c.CustomerID].Metadata
	}

	return customers, nil
}

func buildSearchQuery(params SearchParams) (string, []interface{}) {
	var args []interface{}
	query := `select customer_id, first_name, middle_name, last_name, nick_name, suffix, type, business_name, doing_business_as, business_type, ein, duns, sic_code, naics_code, birth_date, status, email, website, date_business_established, created_at, last_modified
from customers where deleted_at is null`

	if params.Organization != "" {
		query += " and organization = ?"
		args = append(args, params.Organization)
	}

	if params.Query != "" {
		if params.Type == string(client.CUSTOMERTYPE_BUSINESS) {
			query += " and lower(business_name) LIKE ?"
		} else {
			// warning: this will ONLY work for MySQL
			query += " and lower(concat(first_name, ' ', last_name)) LIKE ?"
		}
		args = append(args, fmt.Sprintf("%%%s%%", params.Query))
	}

	if params.Email != "" {
		query += " and lower(email) like ?"
		args = append(args, "%"+params.Email)
	}

	if params.Status != "" {
		query += " and status like ?"
		args = append(args, "%"+params.Status)
	}

	if params.Type != "" {
		query += " and type like ?"
		args = append(args, "%"+params.Type)
	}

	if len(params.CustomerIDs) > 0 {
		query += fmt.Sprintf(" and customer_id in (?%s)", strings.Repeat(",?", len(params.CustomerIDs)-1))
		for _, id := range params.CustomerIDs {
			args = append(args, id)
		}
	}

	query += " order by created_at desc limit ?"
	args = append(args, fmt.Sprintf("%d", params.Count))

	if params.Skip > 0 {
		query += " offset ?"
		args = append(args, fmt.Sprintf("%d", params.Skip))
	}
	query += ";"
	return query, args
}

func (r *sqlCustomerRepository) GetPhones(ownerIDs []string, ownerType client.OwnerType) (map[string][]client.Phone, error) {
	query := fmt.Sprintf(
		"select owner_id, owner_type, number, valid, type from phones where owner_id in (?%s) and owner_type = ?",
		strings.Repeat(",?", len(ownerIDs)-1),
	)

	rows, err := r.queryRowsByCustomerIDsAndOwnerType(query, ownerIDs, ownerType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ret := make(map[string][]client.Phone)
	for rows.Next() {
		var p client.Phone
		var ownerID string
		err := rows.Scan(
			&ownerID,
			&p.OwnerType,
			&p.Number,
			&p.Valid,
			&p.Type,
		)
		if err != nil {
			return nil, fmt.Errorf("scanning row: %v", err)
		}
		ret[ownerID] = append(ret[ownerID], p)
	}

	return ret, nil
}

func (r *sqlCustomerRepository) GetAddresses(customerIDs []string, ownerType client.OwnerType) (map[string][]client.Address, error) {
	query := fmt.Sprintf(
		"select owner_id, owner_type, address_id, type, address1, address2, city, state, postal_code, country, validated from addresses where owner_id in (?%s) and owner_type = ? and deleted_at is null;",
		strings.Repeat(",?", len(customerIDs)-1),
	)
	rows, err := r.queryRowsByCustomerIDsAndOwnerType(query, customerIDs, ownerType)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	ret := make(map[string][]client.Address)
	for rows.Next() {
		var a client.Address
		var ownerID string
		if err := rows.Scan(
			&ownerID,
			&a.OwnerType,
			&a.AddressID,
			&a.Type,
			&a.Address1,
			&a.Address2,
			&a.City,
			&a.State,
			&a.PostalCode,
			&a.Country,
			&a.Validated,
		); err != nil {
			return nil, fmt.Errorf("scanning row: %v", err)
		}
		ret[ownerID] = append(ret[ownerID], a)
	}

	return ret, nil
}

func (r *sqlCustomerRepository) getMetadata(customerIDs []string) (map[string]client.CustomerMetadata, error) {
	query := fmt.Sprintf(
		"select customer_id, meta_key, meta_value from customer_metadata where customer_id in (?%s);",
		strings.Repeat(",?", len(customerIDs)-1),
	)
	rows, err := r.queryRowsByCustomerIDs(query, customerIDs)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]client.CustomerMetadata)
	m := client.CustomerMetadata{
		Metadata: make(map[string]string),
	}

	for rows.Next() {
		var customerID string
		var k, v string
		if err := rows.Scan(&customerID, &k, &v); err != nil {
			return nil, fmt.Errorf("scanning row: %v", err)
		}
		m.Metadata[k] = v
		result[customerID] = m
	}

	return result, nil
}

func (r *sqlCustomerRepository) queryRowsByCustomerIDsAndOwnerType(query string, customerIDs []string, ownerType client.OwnerType) (*sql.Rows, error) {
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("preparing query: %v", err)
	}
	defer stmt.Close()

	var args []interface{}
	for _, id := range customerIDs {
		args = append(args, id)
	}

	args = append(args, string(ownerType))

	rows, err := stmt.Query(args...)
	if err != nil {
		return nil, fmt.Errorf("executing query: %v", err)
	}

	return rows, nil
}

func (r *sqlCustomerRepository) queryRowsByCustomerIDs(query string, customerIDs []string) (*sql.Rows, error) {
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("preparing query: %v", err)
	}
	defer stmt.Close()

	var args []interface{}
	for _, id := range customerIDs {
		args = append(args, id)
	}

	rows, err := stmt.Query(args...)
	if err != nil {
		return nil, fmt.Errorf("executing query: %v", err)
	}

	return rows, nil
}
