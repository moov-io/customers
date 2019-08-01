// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"github.com/moov-io/customers/internal/database"
	"strings"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/moov-io/base"

	"gocloud.dev/secrets"
)

var (
	testCustomerSSNStorage = &ssnStorage{
		keeperFactory: testSecretKeeper(testSecretKey),
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

func TestSSN(t *testing.T) {
	customerId := base.ID()
	ssn := &SSN{customerId: customerId, masked: "1###5"}
	if v := ssn.String(); v != fmt.Sprintf("SSN: customerId=%s masked=1###5", customerId) {
		t.Errorf("got %s", v)
	}

	// ssn storage error case
	storage := &ssnStorage{
		keeperFactory: func(path string) (*secrets.Keeper, error) {
			return nil, errors.New("bad error")
		},
	}
	if _, err := storage.encryptRaw(customerId, "1###5"); err == nil {
		t.Error("expected error")
	} else {
		if !strings.Contains(err.Error(), "ssnStorage: keeper init") {
			t.Errorf("unexpected error: %v", err)
		}
	}
}

func TestCustomerSSNStorage(t *testing.T) {
	storage := &ssnStorage{
		keeperFactory: testSecretKeeper(testSecretKey),
		repo:          &testCustomerSSNRepository{},
	}
	if _, err := storage.encryptRaw("", ""); err == nil {
		t.Errorf("expected error")
	}
	if _, err := storage.encryptRaw(base.ID(), ""); err == nil {
		t.Errorf("expected error")
	}

	// encrypt SSN
	customerId := base.ID()
	ssn, err := storage.encryptRaw(customerId, "123456789")
	if err != nil {
		t.Error(err)
	}
	if ssn.customerId != customerId {
		t.Errorf("ssn.customerId=%s", ssn.customerId)
	}
	if ssn.masked != "1#######9" {
		t.Errorf("ssn.masked=%s", ssn.masked)
	}

	keeper, err := storage.keeperFactory(fmt.Sprintf("customer-%s-ssn", customerId))
	if err != nil {
		t.Fatal(err)
	}
	decrypted, err := keeper.Decrypt(context.Background(), ssn.encrypted)
	if err != nil {
		t.Fatal(err)
	}
	if v := string(decrypted); v != "123456789" {
		t.Errorf("decrypted SSN=%s", v)
	}
}

func TestCustomerSSNRepository(t *testing.T) {
	customerId := base.ID()
	check := func(t *testing.T, customerSSNRepo *sqliteCustomerSSNRepository) {

		if ssn, err := customerSSNRepo.getCustomerSSN(customerId); ssn != nil || err != nil {
			t.Fatalf("ssn=%v error=%v", ssn, err)
		}

		// write
		bs := base64.StdEncoding.EncodeToString([]byte("123456789"))
		ssn := &SSN{customerId: customerId, encrypted: []byte(bs), masked: "1#######9"}
		if err := customerSSNRepo.saveCustomerSSN(ssn); err != nil {
			t.Fatal(err)
		}

		// read again
		ssn, err := customerSSNRepo.getCustomerSSN(customerId)
		if ssn == nil || err != nil {
			t.Fatalf("ssn=%v error=%v", ssn, err)
		}
		out, err := base64.StdEncoding.DecodeString(string(ssn.encrypted))
		if err != nil {
			t.Fatal(err)
		}
		if v := string(out); v != "123456789" {
			t.Errorf("ssn.encrypte=%s", v)
		}
		if ssn.masked != "1#######9" {
			t.Errorf("ssn.masked=%s", ssn.masked)
		}
	}

	// SQLite tests
	sqliteDB := database.CreateTestSqliteDB(t)
	defer sqliteDB.Close()
	check(t, &sqliteCustomerSSNRepository{sqliteDB.DB, log.NewNopLogger()})

	// MySQL tests
	mysqlDB := database.CreateTestMySQLDB(t)
	defer mysqlDB.Close()
	check(t, &sqliteCustomerSSNRepository{mysqlDB.DB, log.NewNopLogger()})
}
