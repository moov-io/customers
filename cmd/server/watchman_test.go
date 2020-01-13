// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"errors"
	"fmt"
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
		Tag:        "v0.13.0",
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
