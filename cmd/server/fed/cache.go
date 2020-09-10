// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package fed

import (
	"errors"

	"github.com/moov-io/customers/pkg/client"

	hashlru "github.com/hashicorp/golang-lru"
)

type cacheClient struct {
	underlying Client
	cache      *hashlru.Cache
}

func NewCacheClient(client Client, maxSize int) Client {
	cache, _ := hashlru.New(maxSize)
	return &cacheClient{
		underlying: client,
		cache:      cache,
	}
}

func (c *cacheClient) Ping() error {
	if c == nil || c.underlying == nil {
		return nil
	}
	return c.underlying.Ping()
}

func (c *cacheClient) LookupInstitution(routingNumber string) (*client.InstitutionDetails, error) {
	if c == nil || c.underlying == nil {
		return nil, errors.New("nil Client")
	}

	val, exists := c.cache.Get(routingNumber)
	if exists {
		if details, ok := val.(*client.InstitutionDetails); ok {
			return details, nil
		}
	}

	details, err := c.underlying.LookupInstitution(routingNumber)
	if err != nil {
		return nil, err
	}
	if details != nil {
		c.cache.Add(routingNumber, details)
	}
	return details, nil
}
