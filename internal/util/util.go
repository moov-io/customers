// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package util

import (
	"strings"
)

// Or returns the first non-empty string
func Or(options ...string) string {
	for i := range options {
		if v := strings.TrimSpace(options[i]); v != "" {
			return v
		}
	}
	return ""
}
