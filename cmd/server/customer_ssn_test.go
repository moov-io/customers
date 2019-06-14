// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

var (
	testCustomerSSNStorage = &ssnStorage{
		keeperFactory: testSecretKeeper,
		repo:          &testCustomerSSNRepository{},
	}
)

type testCustomerSSNRepository struct {
	err error
	ssn *SSN
}

func (r *testCustomerSSNRepository) saveCustomerSSN(*SSN) error {
	return r.err
}

func (r *testCustomerSSNRepository) getCustomerSSN(customerId string) (*SSN, error) {
	if r.ssn != nil {
		return r.ssn, nil
	}
	return nil, r.err
}
