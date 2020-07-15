// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"os"
	"testing"
)

func TestMain(t *testing.M) {
	main()
	os.Exit(t.Run())
}
