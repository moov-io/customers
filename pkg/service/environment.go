package service

import (
	"context"
	"database/sql"

	"github.com/gorilla/mux"
	"github.com/moov-io/identity/pkg/config"
	"github.com/moov-io/identity/pkg/database"
	"github.com/moov-io/identity/pkg/logging"
	"github.com/moov-io/identity/pkg/stime"
	tmw "github.com/moov-io/tumbler/pkg/middleware"
	"github.com/moov-io/tumbler/pkg/webkeys"
)

// Environment - Contains everything thats been instantiated for this service.
type Environment struct {
	Logger       logging.Logger
	Config       *Config
	TimeService  stime.TimeService
	GatewayKeys  webkeys.WebKeysService
	PublicRouter *mux.Router
	Shutdown     func()
}

// NewEnvironment - Generates a new default environment. Overrides can be specified via configs.
func NewEnvironment(env *Environment) (*Environment, error) {
	if env == nil {
		env = &Environment{}
	}

	if env.Logger == nil {
		env.Logger = logging.NewDefaultLogger()
	}

	if env.Config == nil {
		ConfigService := config.NewConfigService(env.Logger)

		global := &GlobalConfig{}
		if err := ConfigService.Load(global); err != nil {
			return nil, err
		}

		env.Config = &global.Customers
	}

	//db setup
	db, close, err := initializeDatabase(env.Logger, env.Config.Database)
	if err != nil {
		close()
		return nil, err
	}
	_ = db // delete once used.

	if env.TimeService == nil {
		env.TimeService = stime.NewSystemTimeService()
	}

	// router
	if env.PublicRouter == nil {
		env.PublicRouter = mux.NewRouter()
	}

	// auth middleware for the tokens coming from the gateway
	GatewayMiddleware, err := tmw.NewServerFromConfig(env.Logger, env.TimeService, env.Config.Gateway)
	if err != nil {
		return nil, env.Logger.Fatal().LogErrorF("Can't startup the Gateway middleware - %w", err)
	}

	env.PublicRouter.Use(GatewayMiddleware.Handler)

	env.Shutdown = func() {
		close()
	}

	return env, nil
}

func initializeDatabase(logger logging.Logger, config database.DatabaseConfig) (*sql.DB, func(), error) {
	ctx, cancelFunc := context.WithCancel(context.Background())

	// migrate database
	db, err := database.New(ctx, logger, config)
	if err != nil {
		return nil, cancelFunc, logger.Fatal().LogError("Error creating database", err)
	}

	shutdown := func() {
		logger.Info().Log("Shutting down the db")
		cancelFunc()
		if err := db.Close(); err != nil {
			logger.Fatal().LogError("Error closing DB", err)
		}
	}

	if err := database.RunMigrations(logger, db, config); err != nil {
		return nil, shutdown, logger.Fatal().LogError("Error running migrations", err)
	}

	logger.Info().Log("finished initializing db")

	return db, shutdown, err
}
