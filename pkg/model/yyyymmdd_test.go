// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package model

import (
	"encoding/json"
	"strings"
	"testing"
)

type response struct {
	BirthDate YYYYMMDD `json:"birthDate"`
}

func TestYYYYMMDD(t *testing.T) {
	in := strings.NewReader(`{"birthDate": "1989-11-09"}`)

	var resp response
	if err := json.NewDecoder(in).Decode(&resp); err != nil {
		t.Fatal(err)
	}

	if v := string(resp.BirthDate); v != "1989-11-09" {
		t.Errorf("got %q", v)
	}
}

func TestYYYYMMDD_Error(t *testing.T) {
	in := strings.NewReader(`{"birthDate": "1989-11-INVALID"}`)

	var resp response
	if err := json.NewDecoder(in).Decode(&resp); err == nil {
		t.Fatalf("expcted error, got %#v", resp)
	}
}
