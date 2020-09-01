package service_test

import (
	"os"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/moov-io/customers/pkg/service"
	"github.com/moov-io/identity/pkg/logging"
	"github.com/stretchr/testify/assert"
)

func Test_Environment_Startup(t *testing.T) {
	a := assert.New(t)

	env := &service.Environment{
		Logger: logging.NewLogger(log.NewLogfmtLogger(log.NewSyncWriter(os.Stderr))),
	}

	env, err := service.NewEnvironment(env)
	a.Nil(err)

	shutdown := env.RunServers(false)
	t.Cleanup(shutdown)
}
