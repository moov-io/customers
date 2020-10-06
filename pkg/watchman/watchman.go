// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package watchman

import (
	"context"
	"fmt"
	"time"

	"github.com/antihax/optional"
	"github.com/moov-io/base/http/bind"
	"github.com/moov-io/base/k8s"
	watchman "github.com/moov-io/watchman/client"

	"github.com/moov-io/base/log"
)

type Client interface {
	Ping() error

	Search(ctx context.Context, name string, requestID string) (*watchman.OfacSdn, error)
}

type moovWatchmanClient struct {
	underlying *watchman.APIClient
	logger     log.Logger
}

func (c *moovWatchmanClient) Ping() error {
	// create a context just for this so ping requests don't require the setup of one
	ctx, cancelFn := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancelFn()

	resp, err := c.underlying.WatchmanApi.Ping(ctx)
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}
	if resp == nil {
		return fmt.Errorf("watchman.Ping: failed: %v", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return fmt.Errorf("watchman.Ping: got status: %s", resp.Status)
	}
	return err
}

func (c *moovWatchmanClient) ofacSearch(ctx context.Context, name string, sdnType string, requestID string) (*watchman.Search, error) {
	search, resp, err := c.underlying.WatchmanApi.Search(ctx, &watchman.SearchOpts{
		Q:          optional.NewString(name),
		SdnType:    optional.NewString(sdnType),
		Limit:      optional.NewInt32(10), // Ask for multiple results as Watchman takes N and then filters, probably a bug in that code.
		XRequestID: optional.NewString(requestID),
	})
	if err != nil {
		return nil, fmt.Errorf("watchman.Search: sdnType=%s: %v", sdnType, err)
	}
	resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, fmt.Errorf("watchman.Search: customer=%q sdnType=%s (status code: %d): %v", name, sdnType, resp.StatusCode, err)
	}
	return &search, err
}

func highestOfacSearchMatch(results ...*watchman.Search) *watchman.Search {
	var max *watchman.Search
	for i := range results {
		if max == nil {
			max = results[i]
			continue
		}

		// Save the result if it's SDN matches higher than what we've saved
		if len(results[i].SDNs) > 0 && len(max.SDNs) > 0 {
			if results[i].SDNs[0].Match > max.SDNs[0].Match {
				// Make sure incoming SDN match is higher than max.alt
				if len(max.AltNames) > 0 {
					if results[i].SDNs[0].Match > max.AltNames[0].Match {
						max = results[i]
						continue
					}
				} else {
					max = results[i]
				}
			}
		}

		// Check if first alt name match is higher than max sdn or alts
		if alts := results[i].AltNames; len(alts) > 0 {
			if len(max.SDNs) > 0 && alts[0].Match > max.SDNs[0].Match {
				max = results[i]
				continue
			}
			if len(max.AltNames) > 0 && alts[0].Match > max.AltNames[0].Match {
				max = results[i]
				continue
			}
		}
	}
	if max == nil {
		return &watchman.Search{}
	}
	return max
}

func (c *moovWatchmanClient) altToSDN(ctx context.Context, alt watchman.OfacAlt, requestID string) (*watchman.OfacSdn, error) {
	sdn, resp, err := c.underlying.WatchmanApi.GetSDN(ctx, alt.EntityID, &watchman.GetSDNOpts{
		XRequestID: optional.NewString(requestID),
	})
	resp.Body.Close()
	if err != nil {
		return nil, fmt.Errorf("watchman.Search: found alt name: %v", err)
	}
	sdn.Match = alt.Match // copy match from original search (GetSDN doesn't do string matching)
	return &sdn, nil
}

// Search returns the top Watchman match given the provided options across SDN names and AltNames
func (c *moovWatchmanClient) Search(ctx context.Context, name string, requestID string) (*watchman.OfacSdn, error) {
	individualSearch, err := c.ofacSearch(ctx, name, "individual", requestID)
	if err != nil {
		return nil, err
	}
	entitySearch, err := c.ofacSearch(ctx, name, "entity", requestID)
	if err != nil {
		return nil, err
	}

	search := highestOfacSearchMatch(individualSearch, entitySearch)

	if search == nil || (len(search.SDNs) == 0 && len(search.AltNames) == 0) {
		return nil, nil // Nothing found
	}

	// We prefer to return the SDN, but if there's an AltName with a higher match return that instead.
	if len(search.SDNs) > 0 && len(search.AltNames) == 0 {
		return &search.SDNs[0], nil // return SDN as it was all we got
	}
	// Take an Alt and find the SDN for it if that was the highest match
	if len(search.SDNs) == 0 && len(search.AltNames) > 0 {
		alt := search.AltNames[0]
		c.logger.Log(fmt.Sprintf("Found AltName=%s,SDN=%s with no higher matched SDNs", alt.AlternateID, alt.EntityID))
		return c.altToSDN(ctx, search.AltNames[0], requestID)
	}
	// AltName matched higher than SDN names, so return the SDN of the matched AltName
	if len(search.SDNs) > 0 && len(search.AltNames) > 0 && (search.AltNames[0].Match > 0.1) && search.AltNames[0].Match > search.SDNs[0].Match {
		alt := search.AltNames[0]
		c.logger.Log(fmt.Sprintf("AltName=%s,SDN=%s had higher match than SDN=%s", alt.AlternateID, alt.EntityID, search.SDNs[0].EntityID))
		return c.altToSDN(ctx, alt, requestID)
	}
	// Return the SDN as Alts matched lower
	if len(search.SDNs) > 0 {
		return &search.SDNs[0], nil
	}

	return nil, nil // Nothing found
}

// NewClient returns an WatchmanClient instance and will default to using the Watchman address in
// moov's standard Kubernetes setup.
//
// endpoint is a DNS record responsible for routing us to an Watchman instance.
// Example: http://watchman.apps.svc.cluster.local:8080
func NewClient(logger log.Logger, endpoint string, debug bool) Client {
	conf := watchman.NewConfiguration()
	conf.BasePath = "http://localhost" + bind.HTTP("watchman")
	conf.Debug = debug

	if k8s.Inside() {
		conf.BasePath = "http://watchman.apps.svc.cluster.local:8080"
	}
	if endpoint != "" {
		conf.BasePath = endpoint // override from provided Watchman_ENDPOINT env variable
	}

	logger = logger.WithKeyValue("package", "watchman")
	logger.Log(fmt.Sprintf("using %s for Watchman address", conf.BasePath))

	return &moovWatchmanClient{
		underlying: watchman.NewAPIClient(conf),
		logger:     logger,
	}
}
