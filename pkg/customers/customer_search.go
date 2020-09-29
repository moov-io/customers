// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package customers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/go-kit/kit/log"
	moovhttp "github.com/moov-io/base/http"
	"github.com/moov-io/customers/pkg/client"
	"github.com/moov-io/customers/pkg/route"
)

func searchCustomers(logger log.Logger, repo CustomerRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)

		organization := route.GetOrganization(w, r)
		if organization == "" {
			return
		}

		params, err := readSearchParams(r)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}
		params.Namespace = organization

		customers, err := repo.searchCustomers(params)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}

		logger.Log("customers", fmt.Sprintf("found %d customers in search", len(customers)))

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(customers)
	}
}

type searchParams struct {
	Namespace string
	Query     string
	Email     string
	Status    string
	Type      string
	Skip      int64
	Count     int64
}

func readSearchParams(r *http.Request) (searchParams, error) {
	params := searchParams{
		Query:  strings.ToLower(strings.TrimSpace(r.URL.Query().Get("query"))),
		Email:  strings.ToLower(strings.TrimSpace(r.URL.Query().Get("email"))),
		Status: strings.ToLower(strings.TrimSpace(r.URL.Query().Get("status"))),
		Type:   strings.ToLower(strings.TrimSpace(r.URL.Query().Get("type"))),
	}
	skip, count, exists, err := moovhttp.GetSkipAndCount(r)
	if exists && err != nil {
		return params, err
	}

	params.Skip = int64(skip)
	params.Count = int64(count)

	return params, nil
}

func (r *sqlCustomerRepository) searchCustomers(params searchParams) ([]*client.Customer, error) {
	query, args := buildSearchQuery(params)

	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var customerIDs []string
	customers := make([]*client.Customer, 0)

	// grab customerIDs
	rows, err := stmt.Query(args...)
	if err != nil {
		return customers, err
	}
	for rows.Next() {
		var customerID string
		if err := rows.Scan(&customerID); err == nil {
			customerIDs = append(customerIDs, customerID)
		} else {
			return customers, err
		}
	}

	// Read each Customer
	for i := range customerIDs {
		cust, err := r.getCustomer(customerIDs[i])
		if err != nil {
			return customers, fmt.Errorf("search: customerID=%s error=%v", customerIDs[i], err)
		}
		customers = append(customers, cust)
	}
	return customers, nil
}

func buildSearchQuery(params searchParams) (string, []interface{}) {
	var args []interface{}
	query := `select customer_id from customers where deleted_at is null and organization = ?`
	args = append(args, params.Namespace)
	if params.Query != "" {
		query += " and lower(first_name) || \" \" || lower(last_name) LIKE ?"
		args = append(args, "%"+params.Query+"%")
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
	query += " order by created_at desc limit ?"
	args = append(args, fmt.Sprintf("%d", params.Count))

	if params.Skip > 0 {
		query += " offset ?"
		args = append(args, fmt.Sprintf("%d", params.Skip))
	}
	query += ";"
	return query, args
}
