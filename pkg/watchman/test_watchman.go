package watchman

import (
	"context"

	watchman "github.com/moov-io/watchman/client"
)

type TestWatchmanClient struct {
	sdn *watchman.OfacSdn

	// error to be returned instead of field from above
	err error
}

func NewTestWatchmanClient(sdn *watchman.OfacSdn, err error) *TestWatchmanClient {
	return &TestWatchmanClient{
		sdn: sdn,
		err: err,
	}
}

func (c *TestWatchmanClient) Ping() error {
	return c.err
}

func (c *TestWatchmanClient) Search(_ context.Context, name string, _ string) (*watchman.OfacSdn, error) {
	if c.err != nil {
		return nil, c.err
	}

	return c.sdn, nil
}
