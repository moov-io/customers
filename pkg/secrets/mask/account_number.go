// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package mask

import (
	"strings"
	"unicode/utf8"
)

func AccountNumber(s string) string {
	length := utf8.RuneCountInString(s)
	if length < 5 {
		return "****" // too short, we can't keep anything
	}
	return strings.Repeat("*", length-4) + s[length-4:]
}
