// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"testing"

	"github.com/moov-io/base"
	"github.com/moov-io/base/docker"
	watchman "github.com/moov-io/watchman/client"

	"github.com/go-kit/kit/log"
	"github.com/ory/dockertest/v3"
)

type testWatchmanClient struct {
	sdn *watchman.OfacSdn

	// error to be returned instead of field from above
	err error
}

func (c *testWatchmanClient) Ping() error {
	return c.err
}

func (c *testWatchmanClient) Search(_ context.Context, name string, _ string) (*watchman.OfacSdn, error) {
	if c.err != nil {
		return nil, c.err
	}
	return c.sdn, nil
}

type watchmanDeployment struct {
	res    *dockertest.Resource
	client WatchmanClient
}

func (d *watchmanDeployment) close(t *testing.T) {
	if err := d.res.Close(); err != nil {
		t.Error(err)
	}
}

func spawnWatchman(t *testing.T) *watchmanDeployment {
	// no t.Helper() call so we know where it failed

	if testing.Short() {
		t.Skip("-short flag enabled")
	}
	if !docker.Enabled() {
		t.Skip("Docker not enabled")
	}

	// Spawn Watchman docker image
	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatal(err)
	}
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "moov/watchman",
		Tag:        "static",
		Cmd:        []string{"-http.addr=:8080"},
	})
	if err != nil {
		t.Fatal(err)
	}

	client := newWatchmanClient(log.NewNopLogger(), fmt.Sprintf("http://localhost:%s", resource.GetPort("8080/tcp")))
	err = pool.Retry(func() error {
		return client.Ping()
	})
	if err != nil {
		t.Fatal(err)
	}
	return &watchmanDeployment{resource, client}
}

func TestWatchman__client(t *testing.T) {
	endpoint := ""
	if client := newWatchmanClient(log.NewNopLogger(), endpoint); client == nil {
		t.Fatal("expected non-nil client")
	}

	// Spawn an Watchman Docker image and ping against it
	deployment := spawnWatchman(t)
	if err := deployment.client.Ping(); err != nil {
		t.Fatal(err)
	}
	deployment.close(t) // close only if successful
}

func TestWatchman__search(t *testing.T) {
	ctx := context.TODO()

	deployment := spawnWatchman(t)

	// Search query that matches an SDN higher than an AltName
	sdn, err := deployment.client.Search(ctx, "Nicolas Maduro", base.ID())
	if err != nil || sdn == nil {
		t.Fatalf("sdn=%v err=%v", sdn, err)
	}
	if sdn.EntityID != "22790" {
		t.Errorf("SDN=%s %#v", sdn.EntityID, sdn)
	}

	// Search query that matches an AltName higher than SDN
	sdn, err = deployment.client.Search(ctx, "Osama BIN LADIN", base.ID())
	if err != nil || sdn == nil {
		t.Fatalf("sdn=%v err=%v", sdn, err)
	}
	if sdn.EntityID != "6365" {
		t.Errorf("SDN=%s %#v", sdn.EntityID, sdn)
	}

	// Lookup an SDN from an alt
	if c, ok := deployment.client.(*moovWatchmanClient); ok {
		search, err := c.ofacSearch(ctx, "Osama Bin Ladin", "individual", base.ID())
		if err != nil {
			t.Fatal(err)
		}
		if len(search.SDNs) == 0 {
			t.Fatalf("no Search results: %#v", search)
		}
		if sdn := search.SDNs[0]; sdn.EntityID != "6365" {
			t.Errorf("SDN=%s %#v", sdn.EntityID, sdn)
		}
	}

	deployment.close(t) // close only if successful
}

func TestWatchman_ping(t *testing.T) {
	client := &testWatchmanClient{}

	// Ping tests
	if err := client.Ping(); err != nil {
		t.Error("expected no error")
	}

	// set error and verify we get it
	client.err = errors.New("ping error")
	if err := client.Ping(); err == nil {
		t.Error("expected error")
	} else {
		if !strings.Contains(err.Error(), "ping error") {
			t.Errorf("unknown error: %v", err)
		}
	}
}

func sdn(match float32) watchman.OfacSdn {
	return watchman.OfacSdn{
		EntityID: base.ID(),
		Match:    match,
	}
}

func alt(match float32) watchman.OfacAlt {
	return watchman.OfacAlt{
		EntityID: base.ID(),
		Match:    match,
	}
}

func TestWatchman__highestOfacSearchMatch(t *testing.T) {
	search := highestOfacSearchMatch(nil)
	if len(search.SDNs) != 0 || len(search.AltNames) != 0 {
		t.Errorf("expected nil ofac Search, got %#v", search)
	}

	// SDN with no alts
	search = highestOfacSearchMatch(
		&watchman.Search{
			SDNs: []watchman.OfacSdn{sdn(0.80)},
		},
	)
	if len(search.SDNs) != 1 || len(search.AltNames) != 0 {
		t.Errorf("unexpected search: %#v", search)
	}

	// Alt with no SDNs
	search = highestOfacSearchMatch(
		&watchman.Search{
			AltNames: []watchman.OfacAlt{alt(0.70)},
		},
	)
	if len(search.SDNs) != 0 || len(search.AltNames) != 1 {
		t.Errorf("unexpected search: %#v", search)
	}

	// Two SDN's, first is higher
	search = highestOfacSearchMatch(
		&watchman.Search{
			SDNs: []watchman.OfacSdn{sdn(0.80)},
		},
		&watchman.Search{
			SDNs: []watchman.OfacSdn{sdn(0.75)},
		},
	)
	if len(search.SDNs) != 1 || math.Abs(float64(0.80-search.SDNs[0].Match)) > 0.01 {
		t.Errorf("unexpected search: %#v", search)
	}

	// Two SDN's, second is higher
	search = highestOfacSearchMatch(
		&watchman.Search{
			SDNs: []watchman.OfacSdn{sdn(0.77)},
		},
		&watchman.Search{
			SDNs: []watchman.OfacSdn{sdn(0.82)},
		},
	)
	if len(search.SDNs) != 1 || math.Abs(float64(0.82-search.SDNs[0].Match)) > 0.01 {
		t.Errorf("unexpected search: %#v", search)
	}

	// Two Alts, first is higher
	search = highestOfacSearchMatch(
		&watchman.Search{
			AltNames: []watchman.OfacAlt{alt(0.90)},
		},
		&watchman.Search{
			AltNames: []watchman.OfacAlt{alt(0.87)},
		},
	)
	if len(search.AltNames) != 1 || math.Abs(float64(0.90-search.AltNames[0].Match)) > 0.01 {
		t.Errorf("unexpected search: %#v", search)
	}

	// Two Alts, second is higher
	search = highestOfacSearchMatch(
		&watchman.Search{
			AltNames: []watchman.OfacAlt{alt(0.87)},
		},
		&watchman.Search{
			AltNames: []watchman.OfacAlt{alt(0.90)},
		},
	)
	if len(search.AltNames) != 1 || math.Abs(float64(0.90-search.AltNames[0].Match)) > 0.01 {
		t.Errorf("unexpected search: %#v", search)
	}

	// SDN first, but other alt is higher
	search = highestOfacSearchMatch(
		&watchman.Search{
			SDNs: []watchman.OfacSdn{sdn(0.80)},
		},
		&watchman.Search{
			AltNames: []watchman.OfacAlt{alt(0.90)},
		},
	)
	if len(search.SDNs) != 0 || len(search.AltNames) != 1 {
		t.Errorf("unexpected search: %#v", search)
	}

	// SDN first, but other alt is higher
	search = highestOfacSearchMatch(
		&watchman.Search{
			SDNs: []watchman.OfacSdn{sdn(0.80)},
		},
		&watchman.Search{
			SDNs:     []watchman.OfacSdn{sdn(0.70)},
			AltNames: []watchman.OfacAlt{alt(0.90)},
		},
	)
	if len(search.SDNs) != 1 || len(search.AltNames) != 1 {
		t.Errorf("unexpected search: %#v", search)
	}

	// SDN first, but other sdn is higher
	search = highestOfacSearchMatch(
		&watchman.Search{
			SDNs: []watchman.OfacSdn{sdn(0.80)},
		},
		&watchman.Search{
			SDNs:     []watchman.OfacSdn{sdn(0.99)},
			AltNames: []watchman.OfacAlt{alt(0.90)},
		},
	)
	if len(search.SDNs) != 1 || len(search.AltNames) != 1 {
		t.Errorf("unexpected search: %#v", search)
	}

	// SDN second and alt is higher
	search = highestOfacSearchMatch(
		&watchman.Search{
			SDNs:     []watchman.OfacSdn{sdn(0.70)},
			AltNames: []watchman.OfacAlt{alt(0.90)},
		},
		&watchman.Search{
			SDNs: []watchman.OfacSdn{sdn(0.80)},
		},
	)
	if len(search.SDNs) != 1 || len(search.AltNames) != 1 {
		t.Errorf("unexpected search: %#v", search)
	}

	// Check
	search = highestOfacSearchMatch(
		&watchman.Search{
			SDNs:     []watchman.OfacSdn{sdn(0.70)},
			AltNames: []watchman.OfacAlt{alt(0.72)},
		},
		&watchman.Search{
			SDNs:     []watchman.OfacSdn{sdn(0.75)},
			AltNames: []watchman.OfacAlt{alt(0.50)},
		},
	)
	if len(search.SDNs) != 1 || len(search.AltNames) != 1 {
		t.Fatalf("unexpected search: %#v", search)
	}
	if math.Abs(float64(0.75-search.SDNs[0].Match)) > 0.01 {
		t.Errorf("unexpected search: %#v", search)
	}
	if math.Abs(float64(0.50-search.AltNames[0].Match)) > 0.01 {
		t.Errorf("unexpected search: %#v", search)
	}
}
