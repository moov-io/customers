// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package customers

import (
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/moov-io/base/strx"
	client "github.com/moov-io/customers/pkg/client"
	tmw "github.com/moov-io/tumbler/pkg/middleware"
)

func (c *customersController) searchCustomersHandler(w http.ResponseWriter, r *http.Request) {
	tmw.WithClaimsFromRequest(w, r, func(claims tmw.TumblerClaims) {
		params := readSearchParams(r.URL)
		customers, err := c.repo.searchCustomers(params)
		if err != nil {
			errorResponse(w, err)
			return
		}
		c.logger.Log(fmt.Sprintf("found %d customers in search", len(customers)))
		jsonResponse(w, customers)
	})
}

type searchParams struct {
	Query string
	Email string
	Limit int64
}

func readSearchParams(u *url.URL) searchParams {
	params := searchParams{
		Query: strings.ToLower(strings.TrimSpace(u.Query().Get("query"))),
		Email: strings.ToLower(strings.TrimSpace(u.Query().Get("email"))),
	}
	if limit, err := strconv.ParseInt(strx.Or(u.Query().Get("limit"), "20"), 10, 32); err == nil {
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
	return query + " order by created_at asc limit ?;", append(args, fmt.Sprintf("%d", params.Limit))
}
