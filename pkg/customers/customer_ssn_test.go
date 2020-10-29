// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package customers

import (
	"encoding/base64"
	"github.com/moov-io/customers/pkg/client"
	"testing"

	"github.com/moov-io/base/database"
	"github.com/moov-io/customers/pkg/secrets"

	"github.com/moov-io/base"
	"github.com/moov-io/base/log"
)

var (
	testCustomerSSNStorage = func(t *testing.T) *ssnStorage {
		return &ssnStorage{
			keeper: secrets.TestStringKeeper(t),
			repo:   &testCustomerSSNRepository{},
		}
	}
)

type testCustomerSSNRepository struct {
	err error
	ssn *SSN
}

func (r *testCustomerSSNRepository) saveSSN(*SSN) error {
	return r.err
}

func (r *testCustomerSSNRepository) getSSN(ownerID string, ownerType client.OwnerType) (*SSN, error) {
	if r.ssn != nil {
		return r.ssn, nil
	}
	return nil, r.err
}

func TestCustomerSSNStorage(t *testing.T) {
	storage := testCustomerSSNStorage(t)

	if _, err := storage.encryptRaw("", client.OWNERTYPE_CUSTOMER, ""); err == nil {
		t.Errorf("expected error")
	}
	if _, err := storage.encryptRaw(base.ID(), client.OWNERTYPE_CUSTOMER, ""); err == nil {
		t.Errorf("expected error")
	}

	// encrypt SSN
	customerID := base.ID()
	ssn, err := storage.encryptRaw(customerID, client.OWNERTYPE_CUSTOMER, "123456789")
	if err != nil {
		t.Error(err)
	}
	if ssn.ownerID != customerID {
		t.Errorf("ssn.ownerID=%s", ssn.ownerID)
	}
	if ssn.masked != "1#######9" {
		t.Errorf("ssn.masked=%s", ssn.masked)
	}

	decrypted, err := storage.keeper.DecryptString(ssn.encrypted)
	if err != nil {
		t.Fatal(err)
	}
	if decrypted != "123456789" {
		t.Errorf("decrypted SSN=%s", decrypted)
	}
}

func TestCustomerSSNRepository(t *testing.T) {
	customerID := base.ID()
	ownerType := client.OWNERTYPE_CUSTOMER
	check := func(t *testing.T, customerSSNRepo *sqlSSNRepository) {

		if ssn, err := customerSSNRepo.getSSN(customerID, ownerType); ssn != nil || err != nil {
			t.Fatalf("ssn=%v error=%v", ssn, err)
		}

		// write
		bs := base64.StdEncoding.EncodeToString([]byte("123456789"))
		ssn := &SSN{ownerID: customerID, ownerType: ownerType, encrypted: bs, masked: "1#######9"}
		if err := customerSSNRepo.saveSSN(ssn); err != nil {
			t.Fatal(err)
		}

		// read again
		ssn, err := customerSSNRepo.getSSN(customerID, ownerType)
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
	check(t, &sqlSSNRepository{sqliteDB.DB, log.NewNopLogger()})

	// MySQL tests
	mysqlDB := database.CreateTestMySQLDB(t)
	defer mysqlDB.Close()
	check(t, &sqlSSNRepository{mysqlDB.DB, log.NewNopLogger()})
}
