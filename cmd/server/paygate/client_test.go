// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package paygate

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/moov-io/base/docker"
	"github.com/moov-io/paygate/pkg/config"

	"github.com/go-kit/kit/log"
	"github.com/ory/dockertest/v3"
	"gopkg.in/yaml.v2"
)

// GetMicroDeposits(accountID, userID string) (*client.MicroDeposits, error)
// InitiateMicroDeposits(userID string, destination client.Destination) error

type deployment struct {
	res    *dockertest.Resource
	client Client
}

func (d *deployment) close(t *testing.T) {
	if d.res == nil {
		return
	}

	if err := d.res.Close(); err != nil {
		t.Error(err)
	}
}

func spawnPayGate(t *testing.T) *deployment {
	// no t.Helper() call so we know where it failed

	if testing.Short() {
		t.Skip("-short flag enabled")
	}

	if os.Getenv("PAYGATE_ENDPOINT") != "" {
		client := NewClient(log.NewNopLogger(), os.Getenv("PAYGATE_ENDPOINT"), false)
		return &deployment{client: client}
	}

	if !docker.Enabled() {
		t.Skip("Docker not enabled")
	}

	// Create a temp dir for our config file
	wd, _ := os.Getwd()
	dir, err := ioutil.TempDir(wd, "paygate")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { os.RemoveAll(dir) })

	fmt.Println("one")
	paygateTag := "v0.8.0-dev"
	writePayGateConfig(t, dir, paygateTag)

	// Spawn Watchman docker image
	pool, err := dockertest.NewPool("")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println("two")
	resource, err := pool.RunWithOptions(&dockertest.RunOptions{
		Repository: "moov/paygate",
		Tag:        paygateTag,
		Cmd:        []string{"-config=/conf/config.yaml"},
		Mounts:     []string{fmt.Sprintf("%s:/conf/", dir)},
	})
	if err != nil {
		t.Fatal(err)
	}

	client := NewClient(log.NewNopLogger(), fmt.Sprintf("http://localhost:%s", resource.GetPort("8080/tcp")), false)

	fmt.Println("three")
	err = pool.Retry(func() error {
		return client.Ping()
	})
	if err != nil {
		t.Fatal(err)
	}

	fmt.Println("four")
	return &deployment{resource, client}
}

// writePayGateConfig pulls down the example config from PayGate's repository for the given
// docker tag we're running. This is easier than trying to stay updated with that project.
func writePayGateConfig(t *testing.T, dir string, tag string) {
	if strings.EqualFold(tag, "latest") || strings.HasSuffix(tag, "-dev") {
		tag = "master"
	}
	url := fmt.Sprintf("https://raw.githubusercontent.com/moov-io/paygate/%s/examples/config.yaml", tag)
	t.Logf("reading paygate config from: %v", url)
	resp, err := http.DefaultClient.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Read(bs)
	if err != nil {
		t.Fatal(err)
	}
	cfg.Http.BindAddress = ":8080"

	fd, err := os.Create(filepath.Join(dir, "config.yaml"))
	if err != nil {
		t.Fatal(err)
	}
	defer fd.Close()

	if err := yaml.NewEncoder(fd).Encode(cfg); err != nil {
		t.Fatal(err)
	}
}

func TestPayGate__client(t *testing.T) {
	if client := NewClient(log.NewNopLogger(), "", false); client == nil {
		t.Fatal("expected non-nil client")
	}

	// Spawn an PayGate Docker image and ping against it
	deployment := spawnPayGate(t)
	if err := deployment.client.Ping(); err != nil {
		t.Fatal(err)
	}
	deployment.close(t) // close only if successful
}
