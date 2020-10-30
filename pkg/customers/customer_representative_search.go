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

func searchCustomerRepresentatives(logger log.Logger, repo CustomerRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w = route.Responder(logger, w, r)

		organization := route.GetOrganization(w, r)
		if organization == "" {
			return
		}

		params, err := parseRepresentativeSearchParams(r)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}

		customerReps, err := repo.searchCustomerRepresentatives(params)
		if err != nil {
			moovhttp.Problem(w, err)
			return
		}

		logger.Logf("found %d customer representatives in search", len(customerReps))

		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(customerReps)
	}
}

type RepresentativeSearchParams struct {
	Query             string
	Skip              int64
	Count             int64
	RepresentativeIDs []string
}

func parseRepresentativeSearchParams(r *http.Request) (RepresentativeSearchParams, error) {
	queryParams := r.URL.Query()
	getQueryParam := func(key string) string {
		return strings.ToLower(strings.TrimSpace(queryParams.Get(key)))
	}
	params := RepresentativeSearchParams{
		Query: getQueryParam("query"),
	}
	representativeIDsInput := getQueryParam("representativeIDs")
	if representativeIDsInput != "" {
		params.RepresentativeIDs = strings.Split(representativeIDsInput, ",")
	}

	skip, count, exists, err := moovhttp.GetSkipAndCount(r)
	if exists && err != nil {
		return params, err
	}

	params.Skip = int64(skip)
	params.Count = int64(count)

	return params, nil
}

func (r *sqlCustomerRepository) searchCustomerRepresentatives(params RepresentativeSearchParams) ([]*client.CustomerRepresentative, error) {
	query, args := buildRepresentativeSearchQuery(params)

	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	customerRepresentatives := make([]*client.CustomerRepresentative, 0)
	rows, err := stmt.Query(args...)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var r client.CustomerRepresentative
		var birthDate *time.Time
		err := rows.Scan(
			&r.RepresentativeID,
			&r.CustomerID,
			&r.FirstName,
			&r.LastName,
			&birthDate,
			&r.CreatedAt,
			&r.LastModified,
		)
		if err != nil {
			return nil, err
		}
		if birthDate != nil {
			r.BirthDate = birthDate.Format(model.YYYYMMDD_Format)
		}
		customerRepresentatives = append(customerRepresentatives, &r)
	}

	if len(customerRepresentatives) == 0 {
		return customerRepresentatives, nil
	}

	var representativeIDs []string
	for _, r := range customerRepresentatives {
		representativeIDs = append(representativeIDs, r.RepresentativeID)
	}

	phonesByCustomerRepresentativeID, err := r.getPhones(representativeIDs, client.OWNERTYPE_REPRESENTATIVE)
	if err != nil {
		return nil, fmt.Errorf("fetching customer representative phones: %v", err)
	}
	addressesByCustomerRepresentativeID, err := r.getAddresses(representativeIDs, client.OWNERTYPE_REPRESENTATIVE)
	if err != nil {
		return nil, fmt.Errorf("fetching customer representative addresses: %v", err)
	}

	for _, r := range customerRepresentatives {
		r.Phones = phonesByCustomerRepresentativeID[r.RepresentativeID]
		r.Addresses = addressesByCustomerRepresentativeID[r.RepresentativeID]
	}

	return customerRepresentatives, nil
}

func buildRepresentativeSearchQuery(params RepresentativeSearchParams) (string, []interface{}) {
	var args []interface{}
	query := `select representative_id, customer_id, first_name, last_name, birth_date, created_at, last_modified
from customer_representatives where deleted_at is null`

	if params.Query != "" {
		// warning: this will ONLY work for MySQL
		query += " and lower(concat(first_name,' ', last_name)) LIKE ?"
		args = append(args, fmt.Sprintf("%%%s%%", params.Query))
	}

	if len(params.RepresentativeIDs) > 0 {
		query += fmt.Sprintf(" and representative_id in (?%s)", strings.Repeat(",?", len(params.RepresentativeIDs)-1))
		for _, id := range params.RepresentativeIDs {
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

func (r *sqlCustomerRepository) queryRowsByCustomerRepresentativeIDs(query string, representativeIDs []string) (*sql.Rows, error) {
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("preparing query: %v", err)
	}
	defer stmt.Close()

	var args []interface{}
	for _, id := range representativeIDs {
		args = append(args, id)
	}

	rows, err := stmt.Query(args...)
	if err != nil {
		return nil, fmt.Errorf("executing query: %v", err)
	}

	return rows, nil
}
