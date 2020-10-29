// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package customers

import (
	"database/sql"
	"fmt"
	"github.com/moov-io/customers/pkg/client"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/moov-io/customers/pkg/secrets"

	"github.com/moov-io/base/log"
)

type SSN struct {
	ownerID   string
	ownerType client.OwnerType
	encrypted string
	masked    string
}

func (s *SSN) String() string {
	return fmt.Sprintf("SSN: ownerId=%s ownerType=%s masked=%s", s.ownerID, s.ownerType, s.masked)
}

type ssnStorage struct {
	keeper *secrets.StringKeeper
	repo   SSNRepository
}

func NewSSNStorage(keeper *secrets.StringKeeper, repo SSNRepository) *ssnStorage {
	return &ssnStorage{
		keeper: keeper,
		repo:   repo,
	}
}

func (s *ssnStorage) encryptRaw(ownerID string, ownerType client.OwnerType, raw string) (*SSN, error) {
	defer func() {
		raw = ""
	}()
	if ownerID == "" || raw == "" {
		return nil, fmt.Errorf("missing parent=%s and/or SSN", ownerID)
	}
	encrypted, err := s.keeper.EncryptString(raw)
	if err != nil {
		return nil, fmt.Errorf("ssnStorage: encrypt owner=%s: %v", ownerID, err)
	}
	return &SSN{
		ownerID:   ownerID,
		ownerType: ownerType,
		encrypted: encrypted,
		masked:    maskSSN(raw),
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

type SSNRepository interface {
	saveSSN(*SSN) error
	getSSN(ownerID string, ownerType client.OwnerType) (*SSN, error)
}

func NewCustomerSSNRepository(logger log.Logger, db *sql.DB) SSNRepository {
	return &sqlSSNRepository{
		db:     db,
		logger: logger,
	}
}

type sqlSSNRepository struct {
	db     *sql.DB
	logger log.Logger
}

//

func (r *sqlSSNRepository) saveSSN(ssn *SSN) error {
	query := `replace into ssn (owner_id, owner_type, ssn, ssn_masked, created_at) values (?, ?, ?, ?, ?);`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return fmt.Errorf("sqlSSNRepository: saveSSN prepare: %v", err)
	}
	defer stmt.Close()

	if _, err := stmt.Exec(ssn.ownerID, string(ssn.ownerType), ssn.encrypted, ssn.masked, time.Now()); err != nil {
		return fmt.Errorf("sqlSSNRepository: saveSSN: exec: %v", err)
	}
	return nil
}

func (r *sqlSSNRepository) getSSN(ownerID string, ownerType client.OwnerType) (*SSN, error) {
	query := `select ssn, ssn_masked from ssn where owner_id = ? and owner_type = ? limit 1;`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("sqlSSNRepository: getSSN prepare: %v", err)
	}
	defer stmt.Close()

	row := stmt.QueryRow(ownerID, string(ownerType))

	ssn := SSN{ownerID: ownerID, ownerType: ownerType}
	if err := row.Scan(&ssn.encrypted, &ssn.masked); err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // not found
		}
		return nil, fmt.Errorf("sqlSSNRepository: getSSN scan: %v", err)
	}
	return &ssn, nil
}
