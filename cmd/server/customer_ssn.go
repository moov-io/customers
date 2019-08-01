// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"github.com/go-kit/kit/log"
	"strings"
	"time"
	"unicode/utf8"
)

type SSN struct {
	customerId string

	encrypted []byte
	masked    string
}

func (s *SSN) String() string {
	return fmt.Sprintf("SSN: customerId=%s masked=%s", s.customerId, s.masked)
}

type ssnStorage struct {
	keeperFactory secretFunc
	repo          customerSSNRepository
}

func (s *ssnStorage) encryptRaw(customerId, raw string) (*SSN, error) {
	defer func() {
		raw = ""
	}()
	if customerId == "" || raw == "" {
		return nil, fmt.Errorf("missing customer=%s and/or SSN", customerId)
	}
	keeper, err := s.keeperFactory(fmt.Sprintf("customer-%s-ssn", customerId))
	if err != nil {
		return nil, fmt.Errorf("ssnStorage: keeper init customer=%s: %v", customerId, err)
	}
	encrypted, err := keeper.Encrypt(context.Background(), []byte(raw))
	if err != nil {
		return nil, fmt.Errorf("ssnStorage: encrypt customer=%s: %v", customerId, err)
	}
	return &SSN{
		customerId: customerId,
		encrypted:  encrypted,
		masked:     maskSSN(raw),
	}, nil
}

func maskSSN(s string) string {
	s = strings.NewReplacer("-", "", ".", "").Replace(strings.TrimSpace(s))
	if utf8.RuneCountInString(s) < 3 {
		return "##" // too short, we can't mask anything
	} else {
		// turn '123456789' into '1******9'
		first, last := s[0:1], s[len(s)-1:]
		return fmt.Sprintf("%s%s%s", first, strings.Repeat("#", len(s)-2), last)
	}
}

type customerSSNRepository interface {
	saveCustomerSSN(*SSN) error
	getCustomerSSN(customerId string) (*SSN, error)
}

type sqlCustomerSSNRepository struct {
	db     *sql.DB
	logger log.Logger
}

func (r *sqlCustomerSSNRepository) close() error {
	return r.db.Close()
}

//

func (r *sqlCustomerSSNRepository) saveCustomerSSN(ssn *SSN) error {
	query := `replace into customer_ssn (customer_id, ssn, ssn_masked, created_at) values (?, ?, ?, ?);`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return fmt.Errorf("sqlCustomerSSNRepository: saveCustomerSSN prepare: %v", err)
	}
	defer stmt.Close()

	encoded := base64.StdEncoding.EncodeToString(ssn.encrypted)
	if _, err := stmt.Exec(ssn.customerId, encoded, ssn.masked, time.Now()); err != nil {
		return fmt.Errorf("sqlCustomerSSNRepository: saveCustomerSSN: exec: %v", err)
	}
	return nil
}

func (r *sqlCustomerSSNRepository) getCustomerSSN(customerId string) (*SSN, error) {
	query := `select ssn, ssn_masked from customer_ssn where customer_id = ? limit 1;`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("sqlCustomerSSNRepository: getCustomerSSN prepare: %v", err)
	}
	defer stmt.Close()

	row := stmt.QueryRow(customerId)

	var encoded string
	ssn := SSN{
		customerId: customerId,
	}
	if err := row.Scan(&encoded, &ssn.masked); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // not found
		}
		return nil, fmt.Errorf("sqlCustomerSSNRepository: getCustomerSSN scan: %v", err)
	}

	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return nil, fmt.Errorf("sqlCustomerSSNRepository: getCustomerSSN decode: %v", err)
	}
	ssn.encrypted = decoded
	return &ssn, nil
}
