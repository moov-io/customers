// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package configuration

import (
	"database/sql"
	"errors"
	"fmt"

	"github.com/moov-io/customers/pkg/client"
)

type Repository interface {
	Get(namespace string) (*client.NamespaceConfiguration, error)
	Update(namespace string, cfg *client.NamespaceConfiguration) (*client.NamespaceConfiguration, error)
}

func NewRepository(db *sql.DB) Repository {
	return &sqlRepo{db: db}
}

type sqlRepo struct {
	db *sql.DB
}

func (r *sqlRepo) Get(namespace string) (*client.NamespaceConfiguration, error) {
	query := `select legal_entity, primary_account from namespace_configuration
where namespace = ? limit 1;`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("namespace get: %v", err)
	}
	defer stmt.Close()

	var cfg client.NamespaceConfiguration
	if err := stmt.QueryRow(namespace).Scan(&cfg.LegalEntity, &cfg.PrimaryAccount); err != nil {
		if err == sql.ErrNoRows {
			return &cfg, nil // nothing found, return an empty config
		}
		return nil, fmt.Errorf("namespace scan: %v", err)
	}
	return &cfg, nil
}

func (r *sqlRepo) Update(namespace string, cfg *client.NamespaceConfiguration) (*client.NamespaceConfiguration, error) {
	// TODO(adam): need to break out repositories to properly test and include this check
	// if err := r.verifyCustomerInfo(namespace, cfg); err != nil {
	// 	return nil, errors.New("namespace: customerID or accountID does not belong")
	// }

	query := `replace into namespace_configuration (namespace, legal_entity, primary_account) values (?, ?, ?);`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("namespace update: %v", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(namespace, cfg.LegalEntity, cfg.PrimaryAccount)
	return cfg, err
}

func (r *sqlRepo) verifyCustomerInfo(namespace string, cfg *client.NamespaceConfiguration) error {
	query := `select c.namespace from customers as c
inner join accounts as a
on c.customer_id = a.customer_id
where c.namespace = ? and (c.customer_id = ? and c.deleted_at is null) and (a.account_id = ? and a.deleted_at is null)
limit 1;`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	var ns string
	if err := stmt.QueryRow(namespace, cfg.LegalEntity, cfg.PrimaryAccount).Scan(&ns); err != nil {
		return err
	}
	if ns != namespace {
		return errors.New("namespace mis-match")
	}
	return nil
}
