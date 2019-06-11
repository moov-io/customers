// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

var (
	testOFACSearcher = &ofacSearcher{
		repo:       &testCustomerRepository{},
		ofacClient: &testOFACClient{},
	}
)
