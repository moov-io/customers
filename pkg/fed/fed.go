// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package fed

import (
	"os"
	"strconv"

	"github.com/moov-io/customers/internal/util"

	"github.com/go-kit/kit/log"
)

func Cache(logger log.Logger, endpoint string, debug bool) Client {
	client := NewClient(logger, endpoint, debug)

	data := util.Or(os.Getenv("FED_CACHE_SIZE"), "1024")
	maxSize, _ := strconv.ParseInt(data, 10, 32)

	return NewCacheClient(client, int(maxSize))
}
