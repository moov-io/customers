// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package customers

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/moov-io/base"
)

func TestStatus__json(t *testing.T) {
	cs := Status(10)
	valid := map[string]Status{
		"deCEAsed":       Deceased,
		"Rejected":       Rejected,
		"ReviewRequired": ReviewRequired,
		"NONE":           None,
		"KYC":            KYC,
		"ofaC":           OFAC,
		"cip":            CIP,
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

func TestStatus__string(t *testing.T) {
	if v := OFAC.String(); v != "ofac" {
		t.Errorf("got %s", v)
	}
	if v := Deceased.String(); v != "deceased" {
		t.Errorf("got %s", v)
	}
}

func TestStatus__liftStatus(t *testing.T) {
	if cs, err := LiftStatus("kyc"); *cs != KYC || err != nil {
		t.Errorf("got %s error=%v", cs, err)
	}
	if cs, err := LiftStatus("none"); *cs != None || err != nil {
		t.Errorf("got %s error=%v", cs, err)
	}
	if cs, err := LiftStatus("cip"); *cs != CIP || err != nil {
		t.Errorf("got %s error=%v", cs, err)
	}
}

func TestStatus__approvedAt(t *testing.T) {
	// authorized
	if !ApprovedAt(OFAC, OFAC) {
		t.Errorf("expected ApprovedAt")
	}
	if !ApprovedAt(OFAC, KYC) {
		t.Errorf("expected ApprovedAt")
	}
	if !ApprovedAt(CIP, KYC) {
		t.Errorf("expected ApprovedAt")
	}

	// not authorized
	if ApprovedAt(ReviewRequired, ReviewRequired) {
		t.Errorf("expected not ApprovedAt")
	}
	if ApprovedAt(None, OFAC) {
		t.Errorf("expected not ApprovedAt")
	}
	if ApprovedAt(OFAC, CIP) {
		t.Errorf("expected not ApprovedAt")
	}
	if ApprovedAt(Rejected, OFAC) {
		t.Errorf("expected not ApprovedAt")
	}
}
