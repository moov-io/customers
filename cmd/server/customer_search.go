// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	moovhttp "github.com/moov-io/base/http"
	client "github.com/moov-io/customers/client"
	"github.com/moov-io/customers/cmd/server/route"
	"github.com/moov-io/customers/internal/util"

	"github.com/go-kit/kit/log"
)

func searchCustomers(logger log.Logger, repo customerRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)

		params := readSearchParams(r.URL)
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
	Query string
	Email string
	Status string
	Limit int64
}

func readSearchParams(u *url.URL) searchParams {
	params := searchParams{
		Query: strings.ToLower(strings.TrimSpace(u.Query().Get("query"))),
		Email: strings.ToLower(strings.TrimSpace(u.Query().Get("email"))),
		Status: strings.ToLower(strings.TrimSpace(u.Query().Get("status"))),
	}
	if limit, err := strconv.ParseInt(util.Or(u.Query().Get("limit"), "20"), 10, 32); err == nil {
		params.Limit = limit
	}
	return params
}

func (r *sqlCustomerRepository) searchCustomers(params searchParams) ([]*client.Customer, error) {
	query, args := buildSearchQuery(params)

	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var customerIDs []string
	rows, err := stmt.Query(args...)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var customerID string
		if err := rows.Scan(&customerID); err == nil {
			customerIDs = append(customerIDs, customerID)
		} else {
			return nil, err
		}
	}

	customers := make([]*client.Customer, 0)
	for i := range customerIDs {
		cust, err := r.getCustomer(customerIDs[i])
		if err != nil {
			return nil, fmt.Errorf("search: customerID=%s error=%v", customerIDs[i], err)
		}
		customers = append(customers, cust)
	}
	return customers, nil
}

func buildSearchQuery(params searchParams) (string, []interface{}) {
	var args []interface{}
	query := `select customer_id from customers where deleted_at is null`
	if params.Query != "" {
		query += " and lower(first_name) || \" \" || lower(last_name) LIKE ?"
		args = append(args, "%"+strings.ToLower(params.Query)+"%")
	}
	if params.Email != "" {
		query += " and lower(email) like ?"
		args = append(args, "%"+strings.ToLower(params.Email))
	}
	if params.Status != "" {
		query += " and status like ?"
		args = append(args, "%"+strings.ToLower(params.Status))
	}
	return query + " order by created_at asc limit ?;", append(args, fmt.Sprintf("%d", params.Limit))
}
