// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package customers

import (
	"encoding/base64"
	"github.com/moov-io/customers/pkg/client"
	"testing"

	"github.com/moov-io/base/database"

	"github.com/moov-io/base"
	"github.com/moov-io/base/log"
)

func TestRepresentativeSSNStorage(t *testing.T) {
	storage := testCustomerSSNStorage(t)

	if _, err := storage.encryptRaw("", client.OWNERTYPE_REPRESENTATIVE, ""); err == nil {
		t.Errorf("expected error")
	}
	if _, err := storage.encryptRaw(base.ID(), client.OWNERTYPE_REPRESENTATIVE, ""); err == nil {
		t.Errorf("expected error")
	}

	// encrypt SSN
	representativeID := base.ID()
	ssn, err := storage.encryptRaw(representativeID, client.OWNERTYPE_REPRESENTATIVE, "987654321")
	if err != nil {
		t.Error(err)
	}
	if ssn.ownerID != representativeID {
		t.Errorf("ssn.ownerID=%s", ssn.ownerID)
	}
	if ssn.masked != "9#######1" {
		t.Errorf("ssn.masked=%s", ssn.masked)
	}

	decrypted, err := storage.keeper.DecryptString(ssn.encrypted)
	if err != nil {
		t.Fatal(err)
	}
	if decrypted != "987654321" {
		t.Errorf("decrypted SSN=%s", decrypted)
	}
}

func TestRepresentativeSSNRepository(t *testing.T) {
	representativeID := base.ID()
	ownerType := client.OWNERTYPE_REPRESENTATIVE
	check := func(t *testing.T, customerSSNRepo *sqlSSNRepository) {

		if ssn, err := customerSSNRepo.getSSN(representativeID, ownerType); ssn != nil || err != nil {
			t.Fatalf("ssn=%v error=%v", ssn, err)
		}

		// write
		bs := base64.StdEncoding.EncodeToString([]byte("987654321"))
		ssn := &SSN{ownerID: representativeID, ownerType: ownerType, encrypted: bs, masked: "9#######1"}
		if err := customerSSNRepo.saveSSN(ssn); err != nil {
			t.Fatal(err)
		}

		// read again
		ssn, err := customerSSNRepo.getSSN(representativeID, ownerType)
		if ssn == nil || err != nil {
			t.Fatalf("ssn=%v error=%v", ssn, err)
		}
		out, err := base64.StdEncoding.DecodeString(string(ssn.encrypted))
		if err != nil {
			t.Fatal(err)
		}
		if v := string(out); v != "987654321" {
			t.Errorf("ssn.encrypte=%s", v)
		}
		if ssn.masked != "9#######1" {
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
