// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package model

import (
	"fmt"
	"strconv"
	"time"
)

var (
	YYYYMMDD_Format = "2006-01-02"
)

type YYYYMMDD string

func (d *YYYYMMDD) UnmarshalJSON(data []byte) error {
	// Ignore null, like in the main JSON package.
	if string(data) == "null" {
		return nil
	}

	s, err := strconv.Unquote(string(data))
	if err != nil {
		return fmt.Errorf("YYMMDD: %v", err)
	}
	t, err := time.Parse(YYYYMMDD_Format, s)
	if err != nil {
		return fmt.Errorf("YYMMDD: %v", err)
	}
	*d = YYYYMMDD(t.Format(YYYYMMDD_Format))
	return nil
}
