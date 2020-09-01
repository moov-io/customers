package main

import (
	"os"

	"github.com/moov-io/customers/pkg/service"
	"github.com/moov-io/identity/pkg/logging"
)

func main() {
	env := &service.Environment{
		Logger: logging.NewDefaultLogger().WithKeyValue("app", "customers"),
	}

	env, err := service.NewEnvironment(env)
	if err != nil {
		env.Logger.Fatal().LogError("Error loading up environment.", err)
		os.Exit(1)
	}
	defer env.Shutdown()

	env.Logger.Info().Log("Starting services")
	shutdown := env.RunServers(true)
	defer shutdown()
}
