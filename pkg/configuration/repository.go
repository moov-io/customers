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
	Get(organization string) (*client.OrganizationConfiguration, error)
	Update(organization string, cfg *client.OrganizationConfiguration) (*client.OrganizationConfiguration, error)
}

func NewRepository(db *sql.DB) Repository {
	return &sqlRepo{db: db}
}

type sqlRepo struct {
	db *sql.DB
}

func (r *sqlRepo) Get(organization string) (*client.OrganizationConfiguration, error) {
	query := `select legal_entity, primary_account from organization_configuration
where organization = ? limit 1;`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("organization get: %v", err)
	}
	defer stmt.Close()

	var cfg client.OrganizationConfiguration
	if err := stmt.QueryRow(organization).Scan(&cfg.LegalEntity, &cfg.PrimaryAccount); err != nil {
		if err == sql.ErrNoRows {
			return &cfg, nil // nothing found, return an empty config
		}
		return nil, fmt.Errorf("organization scan: %v", err)
	}
	return &cfg, nil
}

func (r *sqlRepo) Update(organization string, cfg *client.OrganizationConfiguration) (*client.OrganizationConfiguration, error) {
	if err := r.verifyCustomerInfo(organization, cfg); err != nil {
		return nil, errors.New("organization: customerID or accountID does not belong")
	}

	query := `replace into organization_configuration (organization, legal_entity, primary_account) values (?, ?, ?);`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return nil, fmt.Errorf("organization update: %v", err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(organization, cfg.LegalEntity, cfg.PrimaryAccount)
	return cfg, err
}

func (r *sqlRepo) verifyCustomerInfo(organization string, cfg *client.OrganizationConfiguration) error {
	query := `select c.organization from customers as c
inner join accounts as a
on c.customer_id = a.customer_id
where c.organization = ? and (c.customer_id = ? and c.deleted_at is null) and (a.account_id = ? and a.deleted_at is null)
limit 1;`
	stmt, err := r.db.Prepare(query)
	if err != nil {
		return err
	}
	defer stmt.Close()

	var ns string
	if err := stmt.QueryRow(organization, cfg.LegalEntity, cfg.PrimaryAccount).Scan(&ns); err != nil {
		return err
	}
	if ns != organization {
		return errors.New("organization mis-match")
	}
	return nil
}
