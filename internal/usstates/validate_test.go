// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package usstates

import (
	"testing"
)

func TestValidate(t *testing.T) {
	cases := []string{"OR", "or", "Tx"}
	for i := range cases {
		if !Valid(cases[i]) {
			t.Errorf("expected %q to be a state", cases[i])
		}
	}

	cases = []string{"", "XX"}
	for i := range cases {
		if Valid(cases[i]) {
			t.Errorf("expected %q to be invalid", cases[i])
		}
	}
}
