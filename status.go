// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package customers

import (
	"encoding/json"
	"fmt"
	"strings"
)

type Status int

const (
	Deceased Status = iota
	Rejected
	None
	ReviewRequired
	KYC
	OFAC
	CIP
)

var (
	customerStatusStrings = []string{"deceased", "rejected", "none", "reviewrequired", "kyc", "ofac", "cip"}
)

func (cs Status) validate() error {
	switch cs {
	case Deceased, Rejected:
		return nil
	case ReviewRequired, None:
		return nil
	case KYC, OFAC, CIP:
		return nil
	default:
		return fmt.Errorf("status '%v' is invalid", cs)
	}
}

func (cs Status) String() string {
	if cs < Deceased || cs > CIP {
		return "unknown"
	}
	return customerStatusStrings[int(cs)]
}

func (cs *Status) UnmarshalJSON(b []byte) error {
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

func (cs *Status) fromString(s string) {
	for i := range customerStatusStrings {
		if strings.EqualFold(s, customerStatusStrings[i]) {
			*cs = Status(i)
			return
		}
	}
	*cs = Status(-1)
}

// LiftStatus will attempt to return an enum value of CustomerStatus after reading
// the string value.
func LiftStatus(str string) (*Status, error) {
	var cs Status
	cs.fromString(str)
	if err := cs.validate(); err != nil {
		return nil, err
	}
	return &cs, nil
}

// ApprovedAt returns true only if the customerStatus is higher than ReviewRequired
// and is at least the minimum status. It's used to ensure a specific customer is at least
// KYC, OFAC, or CIP in applications.
func (cs Status) ApprovedAt(minimum Status) bool {
	return ApprovedAt(cs, minimum)
}

// ApprovedAt returns true only if the customerStatus is higher than ReviewRequired
// and is at least the minimum status. It's used to ensure a specific customer is at least
// KYC, OFAC, or CIP in applications.
func ApprovedAt(customerStatus Status, minimum Status) bool {
	if customerStatus <= ReviewRequired {
		return false // any status below ReveiewRequired is never approved
	}
	return customerStatus >= minimum
}
