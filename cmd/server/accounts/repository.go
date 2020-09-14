// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package accounts

import (
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/moov-io/base"
	"github.com/moov-io/customers/pkg/admin"
	"github.com/moov-io/customers/pkg/client"

	"github.com/go-kit/kit/log"
)

type Repository interface {
	getCustomerAccount(customerID, accountID string) (*client.Account, error)
	getCustomerAccounts(customerID string) ([]*client.Account, error)

	createCustomerAccount(customerID, userID string, req *createAccountRequest) (*client.Account, error)
	deactivateCustomerAccount(accountID string) error

	updateAccountStatus(accountID string, status admin.AccountStatus) error

	getEncryptedAccountNumber(customerID, accountID string) (string, error)
}

func NewRepo(logger log.Logger, db *sql.DB) *sqlAccountRepository {
	return &sqlAccountRepository{logger: logger, db: db}
}

type sqlAccountRepository struct {
	db     *sql.DB
	logger log.Logger
}

func (r *sqlAccountRepository) getCustomerAccount(customerID, accountID string) (*client.Account, error) {
	query := `select account_id, holder_name, masked_account_number, routing_number, status, type from accounts where customer_id = ? and account_id = ? and deleted_at is null limit 1;`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	var a client.Account
	row := stmt.QueryRow(customerID, accountID)
	if err := row.Scan(&a.AccountID, &a.HolderName, &a.MaskedAccountNumber, &a.RoutingNumber, &a.Status, &a.Type); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fmt.Errorf("account: %s for customer: %s was not found", accountID, customerID)
		}
		return nil, err
	}
	return &a, nil
}

func (r *sqlAccountRepository) getCustomerAccounts(customerID string) ([]*client.Account, error) {
	query := `select account_id from accounts where customer_id = ? and deleted_at is null;`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	rows, err := stmt.Query(customerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := make([]*client.Account, 0) // allocate array so JSON marshal's [] instead of null
	for rows.Next() {
		var accountID string
		if err := rows.Scan(&accountID); err != nil {
			return nil, err
		}
		acct, err := r.getCustomerAccount(customerID, accountID)
		if err != nil {
			return nil, fmt.Errorf("problem reading accountID=%s error=%v", accountID, err)
		}
		out = append(out, acct)
	}
	return out, nil
}

func (r *sqlAccountRepository) createCustomerAccount(customerID, userID string, req *createAccountRequest) (*client.Account, error) {
	account := &client.Account{
		AccountID:           base.ID(),
		MaskedAccountNumber: req.maskedAccountNumber,
		RoutingNumber:       req.RoutingNumber,
		Status:              client.NONE,
		Type:                req.Type,
	}
	query := `insert into accounts (
  account_id, customer_id, user_id, holder_name,
  encrypted_account_number, hashed_account_number, masked_account_number,
  routing_number, status, type, created_at
) values (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?);`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, err
	}
	defer stmt.Close()

	_, err = stmt.Exec(
		account.AccountID, customerID, userID, req.HolderName,
		req.encryptedAccountNumber, req.hashedAccountNumber, req.maskedAccountNumber,
		account.RoutingNumber, account.Status, account.Type, time.Now(),
	)
	if err != nil {
		return nil, fmt.Errorf("problem creating account=%s: %v", account.AccountID, err)
	}
	return account, nil
}

func (r *sqlAccountRepository) deactivateCustomerAccount(accountID string) error {
	query := `update accounts set deleted_at = ? where account_id = ? and deleted_at is null;`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(time.Now(), accountID)
	if err == sql.ErrNoRows {
		return nil
	}
	return err
}

func (r *sqlAccountRepository) updateAccountStatus(accountID string, status admin.AccountStatus) error {
	query := `update accounts set status = ? where account_id = ? and deleted_at is null;`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(status, accountID)
	return err
}

func (r *sqlAccountRepository) getEncryptedAccountNumber(customerID, accountID string) (string, error) {
	query := `select encrypted_account_number from accounts where customer_id = ? and account_id = ? and deleted_at is null limit 1;`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return "", err
	}
	defer stmt.Close()

	var encrypted string
	if err := stmt.QueryRow(customerID, accountID).Scan(&encrypted); err != nil {
		if err == sql.ErrNoRows {
			return "", nil
		}
		return "", err
	}
	return encrypted, nil
}
