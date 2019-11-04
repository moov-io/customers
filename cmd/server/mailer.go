// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"database/sql"
	"time"

	"github.com/go-kit/kit/log"
)

// TODO(adam): use something like the following, cloud providers usually block raw SMTP:25 access
//
// https://github.com/sendgrid/sendgrid-go
// https://godoc.org/github.com/jordan-wright/email
// https://godoc.org/gopkg.in/gomail.v2

type email struct {
	id         string
	customerID string
	to         string
	body       string
	createdAt  time.Time
	sentAt     time.Time
}

type mailRepository interface {
	enqueueEmail(*email) error

	getCursor() *mailCursor
}

type sqlMailRepository struct {
	logger log.Logger
	db     *sql.DB
}

// TODO(adam): db schemas
//
// email_activation_codes(email_id varchar(40) primary key, hash varchar(128), created_at datetime, clicked_at datetime, deleted_at datetime)
//
// outbound_emails(email_id varchar(40) primary key, customer_id varchar(40), to_address varchar(128), body text, created_at datetime, sent_at datetime, deleted_at datetime)

func (r *sqlMailRepository) close() {
	r.db.Close()
}

func (r *sqlMailRepository) enqueueEmail(em *email) error {
	return nil
}

func (r *sqlMailRepository) getCursor() *mailCursor {
	return setupMailCursor(r.logger, r.db)
}

type mailCursor struct {
	db     *sql.DB
	logger log.Logger

	newerThan time.Time
}

func setupMailCursor(logger log.Logger, db *sql.DB) *mailCursor {
	return &mailCursor{db: db, logger: logger}
}

func (cur *mailCursor) Next() ([]*email, error) {
	return nil, nil
}
