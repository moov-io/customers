package service_test

import (
	"testing"

	"github.com/moov-io/customers/pkg/service"
	"github.com/moov-io/identity/pkg/config"
	"github.com/moov-io/identity/pkg/logging"
	"github.com/stretchr/testify/require"
)

func Test_ConfigLoading(t *testing.T) {
	logger := logging.NewNopLogger()

	ConfigService := config.NewConfigService(logger)

	gc := &service.GlobalConfig{}
	err := ConfigService.Load(gc)
	require.Nil(t, err)
}
