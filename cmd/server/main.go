// Copyright 2020 The Moov Authors
// Use of this source code is governed by an Apache License
// license that can be found in the LICENSE file.

package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/moov-io/base/admin"
	moovhttp "github.com/moov-io/base/http"
	"github.com/moov-io/base/http/bind"

	"github.com/moov-io/base/database"
	mainPkg "github.com/moov-io/customers"

	"github.com/moov-io/base/log"

	"github.com/moov-io/customers/internal"
	"github.com/moov-io/customers/internal/util"
	"github.com/moov-io/customers/pkg/accounts"
	"github.com/moov-io/customers/pkg/config"
	"github.com/moov-io/customers/pkg/configuration"
	"github.com/moov-io/customers/pkg/customers"
	"github.com/moov-io/customers/pkg/documents"
	"github.com/moov-io/customers/pkg/documents/storage"
	"github.com/moov-io/customers/pkg/fed"
	"github.com/moov-io/customers/pkg/paygate"
	"github.com/moov-io/customers/pkg/reports"
	"github.com/moov-io/customers/pkg/secrets"
	"github.com/moov-io/customers/pkg/validator"
	"github.com/moov-io/customers/pkg/validator/microdeposits"
	"github.com/moov-io/customers/pkg/validator/mx"
	"github.com/moov-io/customers/pkg/validator/plaid"
	"github.com/moov-io/customers/pkg/watchman"

	"github.com/gorilla/mux"
	"github.com/mattn/go-sqlite3"
	"gocloud.dev/blob/fileblob"
)

var (
	httpAddr  = flag.String("http.addr", bind.HTTP("customers"), "HTTP listen address")
	adminAddr = flag.String("admin.addr", bind.Admin("customers"), "Admin HTTP listen address")

	flagLogFormat = flag.String("log.format", "", "Format for log lines (Options: json, plain")

	preventInsecureStartup = func() bool {
		prevent, err := strconv.ParseBool(os.Getenv("PREVENT_INSECURE_STARTUP"))
		if err != nil {
			return false
		}
		return prevent
	}()
)

type securityConfiguration struct {
	appSalt            string
	ssnSecretsProvider string
	ssnLocalKey        string
	docSecretsProvider string
	docLocalKey        string
	docStorageProvider string
	docBucketName      string
	fileblobURLSecret  string
	transitLocalKey    string
}

func main() {
	flag.Parse()

	var logger log.Logger
	if strings.ToLower(*flagLogFormat) == "json" {
		logger = log.NewJSONLogger()
	} else {
		logger = log.NewDefaultLogger()
	}

	logger = logger.Set("app", log.String("customers"))
	logger.Set("phase", log.String("startup")).Logf("Starting moov-io/customers server version %s", mainPkg.Version)

	// Channel for errors
	errs := make(chan error)

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	// Setup SQLite database
	if sqliteVersion, _, _ := sqlite3.Version(); sqliteVersion != "" {
		logger.Logf("sqlite version %s", sqliteVersion)
	}

	dbConf := config.New()
	if err := dbConf.Load(); err != nil {
		logger.LogErrorf("failed to load application config: %v", err)
		os.Exit(1)
	}

	ctx := context.TODO()
	db, err := database.NewAndMigrate(ctx, logger, *dbConf.Database)
	if err != nil {
		logger.LogErrorf("failed to connect to database: %v", err)
		os.Exit(1)
	}
	defer db.Close()

	accountsRepo := accounts.NewRepo(logger, db)
	customerRepo := customers.NewCustomerRepo(logger, db)
	customerSSNRepo := customers.NewCustomerSSNRepository(logger, db)
	disclaimerRepo := documents.NewDisclaimerRepo(logger, db)
	documentRepo := documents.NewDocumentRepo(logger, db)
	validationsRepo := validator.NewRepo(db)

	// Start Admin server (with Prometheus metrics)
	adminServer := admin.NewServer(*adminAddr)
	adminServer.AddVersionHandler(mainPkg.Version) // Setup 'GET /version'
	go func() {
		logger.Logf("admin listening on %s", adminServer.BindAddr())
		if err := adminServer.Listen(); err != nil {
			logger.LogErrorf("problem starting admin http: %v", err)
			errs <- err
		}
	}()
	defer adminServer.Shutdown()

	// Create our Watchman client
	debugWatchmanCalls := util.Or(os.Getenv("WATCHMAN_DEBUG_CALLS"), "false")
	watchmanEndpoint := util.Or(os.Getenv("WATCHMAN_ENDPOINT"), os.Getenv("OFAC_ENDPOINT"))
	watchmanClient := watchman.NewClient(logger, watchmanEndpoint, util.Yes(debugWatchmanCalls))
	if watchmanClient == nil {
		panic("No Watchman client created, see WATCHMAN_ENDPOINT")
	}
	adminServer.AddLivenessCheck("watchman", watchmanClient.Ping)
	ofac := customers.NewOFACSearcher(customerRepo, watchmanClient)

	// Register our admin routes
	documents.AddDisclaimerAdminRoutes(logger, adminServer, disclaimerRepo, documentRepo)

	securityCfg := loadSecurityConfig()
	missingOpts := checkMissingSecurityOptions(securityCfg)
	if len(missingOpts) > 0 {
		if preventInsecureStartup {
			logger.Fatal().Log(fmt.Sprintf("prevented insecure startup - missing: %s", strings.Join(missingOpts, ", ")))
			os.Exit(0)
		}
		logger.Warn().Log(fmt.Sprintf("running with insecure configuration - missing: %s", strings.Join(missingOpts, ", ")))
	}

	// Setup Customer SSN storage wrapper
	keeper, err := secrets.OpenSecretKeeper(context.Background(), "customer-ssn", securityCfg.ssnSecretsProvider, securityCfg.ssnLocalKey)
	if err != nil {
		panic(err)
	}
	stringKeeper := secrets.NewStringKeeper(keeper, 10*time.Second)

	customerSSNStorage := customers.NewSSNStorage(stringKeeper, customerSSNRepo)

	// read transit keeper
	transitKeeper, err := secrets.OpenLocal(securityCfg.transitLocalKey)
	if err != nil {
		panic(err)
	}
	transitStringKeeper := secrets.NewStringKeeper(transitKeeper, 10*time.Second)

	debugFedCalls := util.Or(os.Getenv("FED_DEBUG_CALLS"), "false")
	fedClient := fed.Cache(logger, os.Getenv("FED_ENDPOINT"), util.Yes(debugFedCalls))
	adminServer.AddLivenessCheck("fed", fedClient.Ping)

	validationStrategies, err := setupValidationStrategies(logger, adminServer)
	if err != nil {
		panic(err)
	}

	accountOfacSeacher := accounts.AccountOfacSearcher{
		Repo: accountsRepo, WatchmanClient: watchmanClient,
	}

	if util.Yes(os.Getenv("REHASH_ACCOUNTS")) {
		if count, err := internal.RehashStoredAccountNumber(logger, db, securityCfg.appSalt, stringKeeper); err != nil {
			panic(logger.LogErrorf("Failed to re-hash account numbers: %v", err))
		} else {
			logger.Logf("Re-hashed %d account numbers", count)
		}
	}

	// Setup business HTTP routes
	router := mux.NewRouter()
	moovhttp.AddCORSHandler(router)
	addPingRoute(router)
	accounts.RegisterRoutes(logger, router, accountsRepo, validationsRepo, fedClient, stringKeeper, transitStringKeeper, validationStrategies, &accountOfacSeacher, securityCfg.appSalt)
	customers.AddCustomerRoutes(logger, router, customerRepo, customerSSNStorage, ofac)
	customers.AddCustomerAddressRoutes(logger, router, customerRepo)
	customers.AddRepresentativeRoutes(logger, router, customerRepo, customerSSNStorage)
	documents.AddDisclaimerRoutes(logger, router, disclaimerRepo)

	signer := setupSigner(logger, securityCfg.docStorageProvider, securityCfg.fileblobURLSecret)
	bucket := storage.GetBucket(logger, securityCfg.docBucketName, securityCfg.docStorageProvider, signer)
	docsKeeper, err := secrets.OpenSecretKeeper(context.Background(), "customer-documents", securityCfg.docSecretsProvider, securityCfg.docLocalKey)
	if err != nil {
		panic(err)
	}
	defer docsKeeper.Close()

	documents.AddDocumentRoutes(logger, router, documentRepo, docsKeeper, bucket)

	// Optionally serve /files/ as our fileblob routes
	// Note: FILEBLOB_BASE_URL needs to match something that's routed to /files/...
	if securityCfg.docStorageProvider == "file" {
		storage.AddFileblobRoutes(logger, router, signer, bucket)
	}

	customers.AddOFACRoutes(logger, router, customerRepo, ofac)
	reports.AddRoutes(logger, router, customerRepo, accountsRepo)

	// Add Configuration routes
	configRepo := configuration.NewRepository(db)
	configuration.RegisterRoutes(logger, router, configRepo, bucket)

	// Start business HTTP server
	readTimeout, _ := time.ParseDuration("30s")
	writTimeout, _ := time.ParseDuration("30s")
	idleTimeout, _ := time.ParseDuration("60s")

	serve := &http.Server{
		Addr:    *httpAddr,
		Handler: router,
		TLSConfig: &tls.Config{
			InsecureSkipVerify:       false,
			PreferServerCipherSuites: true,
			MinVersion:               tls.VersionTLS12,
		},
		ReadTimeout:  readTimeout,
		WriteTimeout: writTimeout,
		IdleTimeout:  idleTimeout,
	}
	shutdownServer := func() {
		if err := serve.Shutdown(context.TODO()); err != nil {
			logger.Set("phase", log.String("shutdown")).LogErrorf("failed to shutdown server: %v", err)
		}
	}

	// Start business logic HTTP server
	go func() {
		if certFile, keyFile := os.Getenv("HTTPS_CERT_FILE"), os.Getenv("HTTPS_KEY_FILE"); certFile != "" && keyFile != "" {
			logger.Set("phase", log.String("startup")).Logf("binding to %s for secure HTTP server", *httpAddr)
			if err := serve.ListenAndServeTLS(certFile, keyFile); err != nil {
				logger.LogErrorf("failed to start TLS server: %v", err)
			}
		} else {
			logger.Set("phase", log.String("startup")).Logf("binding to %s for HTTP server", *httpAddr)
			if err := serve.ListenAndServe(); err != nil {
				logger.LogErrorf("failed to start server: %v", err)
			}
		}
	}()

	// Block/Wait for an error
	if err := <-errs; err != nil {
		shutdownServer()
		logger.LogErrorf("service error: %v", err)
	}
}

func addPingRoute(r *mux.Router) {
	r.Methods("GET").Path("/ping").HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		moovhttp.SetAccessControlAllowHeaders(w, r.Header.Get("Origin"))
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("PONG"))
	})
}

func loadSecurityConfig() *securityConfiguration {
	cfg := &securityConfiguration{
		appSalt:            os.Getenv("APP_SALT"),
		ssnSecretsProvider: util.Or(os.Getenv("SSN_SECRET_PROVIDER"), "local"),
		ssnLocalKey:        os.Getenv("SSN_SECRET_KEY"),
		docSecretsProvider: util.Or(os.Getenv("DOCUMENTS_SECRET_PROVIDER"), "local"),
		docLocalKey:        os.Getenv("DOCUMENTS_SECRET_KEY"),
		docStorageProvider: util.Or(os.Getenv("DOCUMENTS_STORAGE_PROVIDER"), "file"),
		docBucketName:      util.Or(os.Getenv("DOCUMENTS_BUCKET_NAME"), "./storage"),
		fileblobURLSecret:  os.Getenv("FILEBLOB_HMAC_SECRET"),
		transitLocalKey:    os.Getenv("TRANSIT_LOCAL_BASE64_KEY"),
	}

	return cfg
}

func checkMissingSecurityOptions(cfg *securityConfiguration) []string {
	var missingOpts []string
	if cfg.appSalt == "" {
		missingOpts = append(missingOpts, "APP_SALT")
	}
	if cfg.ssnSecretsProvider == "local" && cfg.ssnLocalKey == "" {
		missingOpts = append(missingOpts, "SSN_SECRET_KEY")
	}
	if cfg.docSecretsProvider == "local" && cfg.docLocalKey == "" {
		missingOpts = append(missingOpts, "DOCUMENTS_SECRET_KEY")
	}
	if cfg.docStorageProvider == "file" && cfg.fileblobURLSecret == "" {
		missingOpts = append(missingOpts, "FILEBLOB_HMAC_SECRET")
	}
	if cfg.transitLocalKey == "" {
		missingOpts = append(missingOpts, "TRANSIT_LOCAL_BASE64_KEY")
	}

	return missingOpts
}
func setupSigner(logger log.Logger, cloudProvider, secret string) *fileblob.URLSignerHMAC {
	if cloudProvider == "file" || cloudProvider == "" {
		baseURL := os.Getenv("FILEBLOB_BASE_URL")
		if baseURL == "" {
			baseURL = fmt.Sprintf("http://localhost%s/files", bind.HTTP("customers"))
		}
		if secret == "" {
			secret = "secret"
			logger.Log("WARNING!!!! USING INSECURE DEFAULT FILE STORAGE, set FILEBLOB_HMAC_SECRET for ANY production usage")
		}
		signer, err := storage.FileblobSigner(baseURL, secret)
		if err != nil {
			panic(fmt.Sprintf("fileBucket: %v", err))
		}
		return signer
	}
	return nil
}

func setupValidationStrategies(logger log.Logger, adminServer *admin.Server) (map[validator.StrategyKey]validator.Strategy, error) {
	strategies := map[validator.StrategyKey]validator.Strategy{}

	// setup microdeposits strategy with moov/paygate
	// we can make paygate optional for customers, but we need a flag for this
	debugPayGateCalls := util.Or(os.Getenv("PAYGATE_DEBUG_CALLS"), "false")
	paygateClient := paygate.NewClient(logger, os.Getenv("PAYGATE_ENDPOINT"), util.Yes(debugPayGateCalls))
	adminServer.AddLivenessCheck("paygate", paygateClient.Ping)
	strategies[validator.StrategyKey{Strategy: "micro-deposits", Vendor: "moov"}] = microdeposits.NewStrategy(paygateClient)

	// setup Plaid instant account verification
	if os.Getenv("PLAID_CLIENT_ID") != "" {
		options := plaid.StrategyOptions{
			ClientID:    os.Getenv("PLAID_CLIENT_ID"),
			Secret:      os.Getenv("PLAID_SECRET"),
			Environment: util.Or(os.Getenv("PLAID_ENVIRONMENT"), "sandbox"),
			ClientName:  os.Getenv("PLAID_CLIENT_NAME"),
		}

		strategy, err := plaid.NewStrategy(options)
		if err != nil {
			return nil, err
		}

		strategies[validator.StrategyKey{Strategy: "instant", Vendor: "plaid"}] = strategy
	}

	if os.Getenv("ATRIUM_CLIENT_ID") != "" {
		options := mx.StrategyOptions{
			ClientID: os.Getenv("ATRIUM_CLIENT_ID"),
			APIKey:   os.Getenv("ATRIUM_API_KEY"),
		}
		strategy := mx.NewStrategy(options)
		strategies[validator.StrategyKey{Strategy: "instant", Vendor: "mx"}] = strategy
	}

	return strategies, nil
}
