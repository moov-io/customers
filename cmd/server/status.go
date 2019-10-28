// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"encoding/json"
	"fmt"
	"strings"
)

type CustomerStatus int

const (
	CustomerStatusDeceased CustomerStatus = iota
	CustomerStatusRejected
	CustomerStatusNone
	CustomerStatusReviewRequired
	CustomerStatusKYC
	CustomerStatusOFAC
	CustomerStatusCIP
)

var (
	customerStatusStrings = []string{"deceased", "rejected", "none", "reviewrequired", "kyc", "ofac", "cip"}
)

func (cs CustomerStatus) validate() error {
	switch cs {
	case CustomerStatusDeceased, CustomerStatusRejected:
		return nil
	case CustomerStatusReviewRequired, CustomerStatusNone:
		return nil
	case CustomerStatusKYC, CustomerStatusOFAC, CustomerStatusCIP:
		return nil
	default:
		return fmt.Errorf("CustomerStatus(%v) is invalid", cs)
	}
}

func (cs CustomerStatus) String() string {
	if cs < CustomerStatusDeceased || cs > CustomerStatusCIP {
		return "unknown"
	}
	return customerStatusStrings[int(cs)]
}

func (cs *CustomerStatus) UnmarshalJSON(b []byte) error {
	var s string
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	cs.fromString(s)
	if err := cs.validate(); err != nil {
		return err
	}
	return nil
}

func (cs *CustomerStatus) fromString(s string) {
	for i := range customerStatusStrings {
		if strings.EqualFold(s, customerStatusStrings[i]) {
			*cs = CustomerStatus(i)
			return
		}
	}
	*cs = CustomerStatus(-1)
}

// LiftStatus will attempt to return an enum value of CustomerStatus after reading
// the string value.
func LiftStatus(str string) (*CustomerStatus, error) {
	var cs CustomerStatus
	cs.fromString(str)
	if err := cs.validate(); err != nil {
		return nil, err
	}
	return &cs, nil
}

// ApprovedAt returns true only if the customerStatus is higher than ReviewRequired
// and is at least the minimum status. It's used to ensure a specific customer is at least
// KYC, OFAC, or CIP in applications.
func (cs CustomerStatus) ApprovedAt(minimum CustomerStatus) bool {
	return ApprovedAt(cs, minimum)
}

// ApprovedAt returns true only if the customerStatus is higher than ReviewRequired
// and is at least the minimum status. It's used to ensure a specific customer is at least
// KYC, OFAC, or CIP in applications.
func ApprovedAt(customerStatus CustomerStatus, minimum CustomerStatus) bool {
	if customerStatus <= CustomerStatusReviewRequired {
		return false // any status below ReveiewRequired is never approved
	}
	return customerStatus >= minimum
}
