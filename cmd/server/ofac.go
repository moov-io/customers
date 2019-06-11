// Copyright 2019 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/antihax/optional"
	"github.com/moov-io/base/http/bind"
	"github.com/moov-io/base/k8s"
	ofac "github.com/moov-io/ofac/client"

	"github.com/go-kit/kit/log"
)

type OFACClient interface {
	Ping() error

	Search(ctx context.Context, name string, requestId string) (*ofac.Sdn, error)
}

type moovOFACClient struct {
	underlying *ofac.APIClient
	logger     log.Logger
}

func (c *moovOFACClient) Ping() error {
	// create a context just for this so ping requests don't require the setup of one
	ctx, cancelFn := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancelFn()

	resp, err := c.underlying.OFACApi.Ping(ctx)
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}
	if resp == nil {
		return fmt.Errorf("ofac.Ping: failed: %v", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("ofac.Ping: got status: %s", resp.Status)
	}
	return err
}

// Search returns the top OFAC match given the provided options across SDN names and AltNames
func (c *moovOFACClient) Search(ctx context.Context, name string, requestId string) (*ofac.Sdn, error) {
	search, resp, err := c.underlying.OFACApi.Search(ctx, &ofac.SearchOpts{
		Q:          optional.NewString(name),
		Limit:      optional.NewInt32(1),
		XRequestId: optional.NewString(requestId),
	})
	if err != nil {
		return nil, fmt.Errorf("ofac.Search: %v", err)
	}
	resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("ofac.Search: customer=%q (status code: %d): %v", name, resp.StatusCode, err)
	}
	// We prefer to return the SDN, but if there's an AltName with a higher match return that instead.
	if (len(search.SDNs) > 0 && len(search.AltNames) > 0) && ((search.AltNames[0].Match > 0.1) && (search.AltNames[0].Match > search.SDNs[0].Match)) {
		alt := search.AltNames[0]

		// AltName matched higher than SDN names, so return the SDN of the matched AltName
		sdn, resp, err := c.underlying.OFACApi.GetSDN(ctx, alt.EntityID, &ofac.GetSDNOpts{
			XRequestId: optional.NewString(requestId),
		})
		resp.Body.Close()
		if err != nil {
			return nil, fmt.Errorf("ofac.Search: found alt name: %v", err)
		}
		sdn.Match = alt.Match // copy match from original search (GetSDN doesn't do string matching)
		c.logger.Log("ofac", fmt.Sprintf("AltName=%s,SDN=%s had higher match than SDN=%s", alt.AlternateID, alt.EntityID, search.SDNs[0].EntityID), "requestId", requestId)
		return &sdn, nil
	} else {
		if len(search.SDNs) > 0 {
			return &search.SDNs[0], nil // return the SDN which had a higher match than any AltNames
		}
	}
	return nil, nil // no OFAC results found, so cust not blocked
}

// newOFACClient returns an OFACClient instance and will default to using the OFAC address in
// moov's standard Kubernetes setup.
//
// endpoint is a DNS record responsible for routing us to an OFAC instance.
// Example: http://ofac.apps.svc.cluster.local:8080
func newOFACClient(logger log.Logger, endpoint string) OFACClient {
	conf := ofac.NewConfiguration()
	conf.BasePath = "http://localhost" + bind.HTTP("ofac")

	if k8s.Inside() {
		conf.BasePath = "http://ofac.apps.svc.cluster.local:8080"
	}
	if endpoint != "" {
		conf.BasePath = endpoint // override from provided OFAC_ENDPOINT env variable
	}

	logger.Log("ofac", fmt.Sprintf("using %s for OFAC address", conf.BasePath))

	return &moovOFACClient{
		underlying: ofac.NewAPIClient(conf),
		logger:     logger,
	}
}
