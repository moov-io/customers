// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"testing"

	"github.com/moov-io/base"
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
	db, err := createTestSqliteDB()
	if err != nil {
		t.Fatal(err)
	}
	defer db.close()

	customerId := base.ID()
	repo := &sqliteCustomerSSNRepository{db.db}

	if ssn, err := repo.getCustomerSSN(customerId); ssn != nil || err != nil {
		t.Fatalf("ssn=%v error=%v", ssn, err)
	}

	// write
	bs := base64.StdEncoding.EncodeToString([]byte("123456789"))
	ssn := &SSN{customerId: customerId, encrypted: []byte(bs), masked: "1########9"}
	if err := repo.saveCustomerSSN(ssn); err != nil {
		t.Fatal(err)
	}

	// read again
	ssn, err = repo.getCustomerSSN(customerId)
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
	if ssn.masked != "1########9" {
		t.Errorf("ssn.masked=%s", ssn.masked)
	}
}
