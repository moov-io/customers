// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package paygate

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/moov-io/base/http/bind"
	"github.com/moov-io/base/k8s"
	"github.com/moov-io/paygate/pkg/client"

	"github.com/moov-io/base/log"
)

type Client interface {
	Ping() error

	GetMicroDeposits(accountID, userID string) (*client.MicroDeposits, error)
	InitiateMicroDeposits(userID string, destination client.Destination) error
}

type moovClient struct {
	logger     log.Logger
	underlying *client.APIClient
}

func (c *moovClient) Ping() error {
	ctx, cancelFn := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancelFn()

	resp, err := c.underlying.MonitorApi.Ping(ctx)
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}
	if resp == nil {
		return fmt.Errorf("PayGate ping failed: %v", err)
	}
	if resp.StatusCode < http.StatusOK || resp.StatusCode >= http.StatusMultipleChoices {
		return fmt.Errorf("PayGate ping got status: %s", resp.Status)
	}
	return err
}

func sqlNoRows(err error) error {
	// Dig into the OpenAPI error and try to find sql.ErrNoRows
	if e, ok := err.(client.GenericOpenAPIError); ok {
		if e, ok := e.Model().(client.Error); ok {
			if strings.EqualFold(e.Error, sql.ErrNoRows.Error()) {
				return nil
			}
		}
	}
	return err
}

func (c *moovClient) GetMicroDeposits(accountID, userID string) (*client.MicroDeposits, error) {
	ctx, cancelFn := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancelFn()

	micro, resp, err := c.underlying.ValidationApi.GetAccountMicroDeposits(ctx, accountID, userID)
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}
	if err != nil {
		// Check if the error is sql.ErrNoRows and if so return nil
		if err := sqlNoRows(err); err == nil {
			return nil, nil
		}
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected HTTP status: %s", resp.Status)
	}
	return &micro, nil
}

func (c *moovClient) InitiateMicroDeposits(userID string, destination client.Destination) error {
	ctx, cancelFn := context.WithTimeout(context.TODO(), 10*time.Second)
	defer cancelFn()

	micro, resp, err := c.underlying.ValidationApi.InitiateMicroDeposits(ctx, userID, client.CreateMicroDeposits{
		Destination: destination,
	})
	if resp != nil && resp.Body != nil {
		resp.Body.Close()
	}
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected HTTP status: %s", resp.Status)
	}

	c.logger.Logf("created microDepositID=%s for accountID=%s", micro.MicroDepositID, destination.AccountID)

	return nil
}

var (
	httpClient = &http.Client{
		Timeout: 10 * time.Second,
	}
)

func NewClient(logger log.Logger, endpoint string, debug bool) Client {
	conf := client.NewConfiguration()
	conf.HTTPClient = httpClient
	conf.Debug = debug

	if endpoint != "" {
		conf.BasePath = endpoint
	} else {
		if k8s.Inside() {
			conf.BasePath = "http://paygate.apps.svc.cluster.local:8080"
		} else {
			conf.BasePath = "http://localhost" + bind.HTTP("paygate")
		}
	}

	logger = logger.Set("package", "paygate")
	logger.Logf("using %s for PayGate address", conf.BasePath)

	return &moovClient{
		logger:     logger,
		underlying: client.NewAPIClient(conf),
	}
}
