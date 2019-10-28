// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/moov-io/base"
)

func TestCustomerStatus__json(t *testing.T) {
	cs := CustomerStatus(10)
	valid := map[string]CustomerStatus{
		"deCEAsed":       CustomerStatusDeceased,
		"Rejected":       CustomerStatusRejected,
		"ReviewRequired": CustomerStatusReviewRequired,
		"NONE":           CustomerStatusNone,
		"KYC":            CustomerStatusKYC,
		"ofaC":           CustomerStatusOFAC,
		"cip":            CustomerStatusCIP,
	}
	for k, v := range valid {
		in := []byte(fmt.Sprintf(`"%s"`, k))
		if err := json.Unmarshal(in, &cs); err != nil {
			t.Error(err.Error())
		}
		if cs != v {
			t.Errorf("got cs=%#v, v=%#v", cs, v)
		}
	}

	// make sure other values fail
	in := []byte(fmt.Sprintf(`"%v"`, base.ID()))
	if err := json.Unmarshal(in, &cs); err == nil {
		t.Error("expected error")
	}
}

func TestCustomerStatus__string(t *testing.T) {
	if v := CustomerStatusOFAC.String(); v != "ofac" {
		t.Errorf("got %s", v)
	}
	if v := CustomerStatusDeceased.String(); v != "deceased" {
		t.Errorf("got %s", v)
	}
}

func TestCustomerStatus__liftStatus(t *testing.T) {
	if cs, err := LiftStatus("kyc"); *cs != CustomerStatusKYC || err != nil {
		t.Errorf("got %s error=%v", cs, err)
	}
	if cs, err := LiftStatus("none"); *cs != CustomerStatusNone || err != nil {
		t.Errorf("got %s error=%v", cs, err)
	}
	if cs, err := LiftStatus("cip"); *cs != CustomerStatusCIP || err != nil {
		t.Errorf("got %s error=%v", cs, err)
	}
}

func TestCustomerStatus__approvedAt(t *testing.T) {
	// authorized
	if !ApprovedAt(CustomerStatusOFAC, CustomerStatusOFAC) {
		t.Errorf("expected ApprovedAt")
	}
	if !ApprovedAt(CustomerStatusOFAC, CustomerStatusKYC) {
		t.Errorf("expected ApprovedAt")
	}
	if !ApprovedAt(CustomerStatusCIP, CustomerStatusKYC) {
		t.Errorf("expected ApprovedAt")
	}

	// not authorized
	if ApprovedAt(CustomerStatusReviewRequired, CustomerStatusReviewRequired) {
		t.Errorf("expected not ApprovedAt")
	}
	if ApprovedAt(CustomerStatusNone, CustomerStatusOFAC) {
		t.Errorf("expected not ApprovedAt")
	}
	if ApprovedAt(CustomerStatusOFAC, CustomerStatusCIP) {
		t.Errorf("expected not ApprovedAt")
	}
	if ApprovedAt(CustomerStatusRejected, CustomerStatusOFAC) {
		t.Errorf("expected not ApprovedAt")
	}
}
