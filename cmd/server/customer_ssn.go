// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
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

type sqliteCustomerSSNRepository struct {
	db *sql.DB
}

func (r *sqliteCustomerSSNRepository) close() error {
	return r.db.Close()
}

func (r *sqliteCustomerSSNRepository) saveCustomerSSN(*SSN) error {
	return nil
}

func (r *sqliteCustomerSSNRepository) getCustomerSSN(customerId string) (*SSN, error) {
	return nil, nil
}
