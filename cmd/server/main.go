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
	"strings"
	"syscall"
	"time"

	"github.com/moov-io/base/admin"
	moovhttp "github.com/moov-io/base/http"
	"github.com/moov-io/base/http/bind"

	mainPkg "github.com/moov-io/customers"

	"github.com/moov-io/base/database"
	"github.com/moov-io/base/log"

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
)

func main() {
	flag.Parse()

	var logger log.Logger
	if strings.ToLower(*flagLogFormat) == "json" {
		logger = log.NewJSONLogger()
	} else {
		logger = log.NewDefaultLogger()
	}

	logger = logger.WithKeyValue("app", "customers")
	logger.WithKeyValue("phase", "startup").Log(fmt.Sprintf("Starting moov-io/customers server version %s", mainPkg.Version))

	// Channel for errors
	errs := make(chan error)

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	// Setup SQLite database
	if sqliteVersion, _, _ := sqlite3.Version(); sqliteVersion != "" {
		logger.Log(fmt.Sprintf("sqlite version %s", sqliteVersion))
	}

	// Setup database connection
	conf := config.New()
	err := conf.Load()
	if err != nil {
		logger.LogError("failed to load application config", err)
		os.Exit(1)
	}

	ctx := context.TODO()
	db, closeDB, err := database.NewAndMigrate(*conf.Database, logger, ctx)
	if err != nil {
		logger.LogError("failed to connect to database", err)
		os.Exit(1)
	}
	defer closeDB()

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
		logger.WithKeyValue("admin", fmt.Sprintf("listening on %s", adminServer.BindAddr()))
		if err := adminServer.Listen(); err != nil {
			err = fmt.Errorf("problem starting admin http: %v", err)
			logger.WithKeyValue("admin", err.Error())
			errs <- err
		}
	}()
	defer adminServer.Shutdown()

	// Setup our cloud Storage object
	bucketName := os.Getenv("BUCKET_NAME")
	if bucketName == "" {
		bucketName = "./storage"
	}
	cloudProvider := strings.ToLower(os.Getenv("SSN_SECRET_PROVIDER"))
	if cloudProvider == "" {
		cloudProvider = "file"
	}
	signer := setupSigner(logger, bucketName, cloudProvider)

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

	// Setup Customer SSN storage wrapper
	keeper, err := secrets.OpenSecretKeeper(context.Background(), "customer-ssn", os.Getenv("SSN_SECRET_PROVIDER"), os.Getenv("SSN_SECRET_KEY"))
	if err != nil {
		panic(err)
	}
	stringKeeper := secrets.NewStringKeeper(keeper, 10*time.Second)

	customerSSNStorage := customers.NewSSNStorage(stringKeeper, customerSSNRepo)

	// read transit keeper
	transitKeeper, err := secrets.OpenLocal(os.Getenv("TRANSIT_LOCAL_BASE64_KEY"))
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

	// Setup business HTTP routes
	router := mux.NewRouter()
	moovhttp.AddCORSHandler(router)
	addPingRoute(router)
	accounts.RegisterRoutes(logger, router, accountsRepo, validationsRepo, fedClient, stringKeeper, transitStringKeeper, validationStrategies, &accountOfacSeacher)
	customers.AddCustomerRoutes(logger, router, customerRepo, customerSSNStorage, ofac)
	customers.AddCustomerAddressRoutes(logger, router, customerRepo)
	documents.AddDisclaimerRoutes(logger, router, disclaimerRepo)

	docsStorageProvider := util.Or(os.Getenv("DOCUMENTS_STORAGE_PROVIDER"), "file")
	docsBucketName := util.Or(os.Getenv("DOCUMENTS_BUCKET_NAME"), "./storage")
	bucket := storage.GetBucket(logger, docsBucketName, docsStorageProvider, signer)

	docsSecretProvider := util.Or(os.Getenv("DOCUMENTS_SECRET_PROVIDER"), "local")
	docsKeeper, err := secrets.OpenSecretKeeper(context.Background(), "customer-documents", docsSecretProvider, os.Getenv("DOCUMENTS_SECRET_KEY"))
	if err != nil {
		panic(err)
	}
	defer docsKeeper.Close()

	documents.AddDocumentRoutes(logger, router, documentRepo, docsKeeper, bucket)
	customers.AddOFACRoutes(logger, router, customerRepo, ofac)
	reports.AddRoutes(logger, router, customerRepo, accountsRepo)

	// Add Configuration routes
	configRepo := configuration.NewRepository(db)
	configuration.RegisterRoutes(logger, router, configRepo, bucket)

	// Optionally serve /files/ as our fileblob routes
	// Note: FILEBLOB_BASE_URL needs to match something that's routed to /files/...
	if cloudProvider == "file" {
		storage.AddFileblobRoutes(logger, router, signer, bucket)
	}

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
			logger.WithKeyValue("phase", "shutdown").LogError("failed to shutdown server", err)
		}
	}

	// Start business logic HTTP server
	go func() {
		if certFile, keyFile := os.Getenv("HTTPS_CERT_FILE"), os.Getenv("HTTPS_KEY_FILE"); certFile != "" && keyFile != "" {
			logger.WithKeyValue("phase", "startup").Log(fmt.Sprintf("binding to %s for secure HTTP server", *httpAddr))
			if err := serve.ListenAndServeTLS(certFile, keyFile); err != nil {
				logger.LogError("failed to start TLS server", err)
			}
		} else {
			logger.WithKeyValue("phase", "startup").Log(fmt.Sprintf("binding to %s for HTTP server", *httpAddr))
			if err := serve.ListenAndServe(); err != nil {
				logger.LogError("failed to start server", err)
			}
		}
	}()

	// Block/Wait for an error
	if err := <-errs; err != nil {
		shutdownServer()
		logger.LogError("service error", err)
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

func setupSigner(logger log.Logger, bucketName, cloudProvider string) *fileblob.URLSignerHMAC {
	if cloudProvider == "file" || cloudProvider == "" {
		baseURL, secret := os.Getenv("FILEBLOB_BASE_URL"), os.Getenv("FILEBLOB_HMAC_SECRET")
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
