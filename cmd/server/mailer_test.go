// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

type testMailerRepository struct {
	enqueuedEmails []*email
}

func (r *testMailerRepository) enqueueEmail(em *email) error {
	r.enqueuedEmails = append(r.enqueuedEmails, em)
}

func (r *testMailerRepository) getCursor() *mailCursor {
	return nil // TODO(adam): impl around an r.db ??
}
