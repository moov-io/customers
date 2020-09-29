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
	"github.com/moov-io/customers/internal/database"
	"github.com/moov-io/customers/internal/util"
	"github.com/moov-io/customers/pkg/accounts"
	"github.com/moov-io/customers/pkg/configuration"
	"github.com/moov-io/customers/pkg/customers"
	"github.com/moov-io/customers/pkg/documents"
	"github.com/moov-io/customers/pkg/documents/storage"
	"github.com/moov-io/customers/pkg/fed"
	"github.com/moov-io/customers/pkg/paygate"
	"github.com/moov-io/customers/pkg/secrets"
	"github.com/moov-io/customers/pkg/validator"
	"github.com/moov-io/customers/pkg/validator/microdeposits"
	"github.com/moov-io/customers/pkg/validator/mx"
	"github.com/moov-io/customers/pkg/validator/plaid"
	"github.com/moov-io/customers/pkg/watchman"

	"github.com/go-kit/kit/log"
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
		logger = log.NewJSONLogger(os.Stderr)
	} else {
		logger = log.NewLogfmtLogger(os.Stderr)
	}
	logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	logger = log.With(logger, "caller", log.DefaultCaller)

	logger.Log("startup", fmt.Sprintf("Starting moov-io/customers server version %s", mainPkg.Version))

	// Channel for errors
	errs := make(chan error)

	go func() {
		c := make(chan os.Signal, 1)
		signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)
		errs <- fmt.Errorf("%s", <-c)
	}()

	// Setup SQLite database
	if sqliteVersion, _, _ := sqlite3.Version(); sqliteVersion != "" {
		logger.Log("main", fmt.Sprintf("sqlite version %s", sqliteVersion))
	}

	// Setup database connection
	db, err := database.New(logger, os.Getenv("DATABASE_TYPE"))
	if err != nil {
		logger.Log("main", err)
		os.Exit(1)
	}
	defer func() {
		if err := db.Close(); err != nil {
			logger.Log("main", err)
		}
	}()

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
		logger.Log("admin", fmt.Sprintf("listening on %s", adminServer.BindAddr()))
		if err := adminServer.Listen(); err != nil {
			err = fmt.Errorf("problem starting admin http: %v", err)
			logger.Log("admin", err)
			errs <- err
		}
	}()
	defer adminServer.Shutdown()

	// Setup our cloud Storage object
	bucketName := os.Getenv("BUCKET_NAME")
	if bucketName == "" {
		bucketName = "./storage"
	}
	cloudProvider := strings.ToLower(os.Getenv("CLOUD_PROVIDER"))
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
	customers.AddApprovalRoutes(logger, adminServer, customerRepo, customerSSNRepo, ofac)
	documents.AddDisclaimerAdminRoutes(logger, adminServer, disclaimerRepo, documentRepo)

	// Setup Customer SSN storage wrapper
	ctx := context.Background()
	keeper, err := secrets.OpenSecretKeeper(ctx, "customer-ssn", os.Getenv("CLOUD_PROVIDER"))
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

	bucket := storage.GetBucket(logger, os.Getenv("DOCUMENTS_BUCKET"), util.Or(os.Getenv("DOCUMENTS_PROVIDER"), "file"), signer)

	// Setup business HTTP routes
	router := mux.NewRouter()
	moovhttp.AddCORSHandler(router)
	addPingRoute(router)
	accounts.RegisterRoutes(logger, router, accountsRepo, validationsRepo, fedClient, stringKeeper, transitStringKeeper, validationStrategies, &accountOfacSeacher)
	customers.AddCustomerRoutes(logger, router, customerRepo, customerSSNStorage, ofac)
	customers.AddCustomerAddressRoutes(logger, router, customerRepo)
	documents.AddDisclaimerRoutes(logger, router, disclaimerRepo)
	documents.AddDocumentRoutes(logger, router, documentRepo, bucket)
	customers.AddOFACRoutes(logger, router, customerRepo, ofac)

	// Add Configuration routes
	configRepo := configuration.NewRepository(db)
	configuration.RegisterRoutes(logger, router, configRepo)

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
			logger.Log("shutdown", err)
		}
	}

	// Start business logic HTTP server
	go func() {
		if certFile, keyFile := os.Getenv("HTTPS_CERT_FILE"), os.Getenv("HTTPS_KEY_FILE"); certFile != "" && keyFile != "" {
			logger.Log("startup", fmt.Sprintf("binding to %s for secure HTTP server", *httpAddr))
			if err := serve.ListenAndServeTLS(certFile, keyFile); err != nil {
				logger.Log("exit", err)
			}
		} else {
			logger.Log("startup", fmt.Sprintf("binding to %s for HTTP server", *httpAddr))
			if err := serve.ListenAndServe(); err != nil {
				logger.Log("exit", err)
			}
		}
	}()

	// Block/Wait for an error
	if err := <-errs; err != nil {
		shutdownServer()
		logger.Log("exit", err)
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
			logger.Log("main", "WARNING!!!! USING INSECURE DEFAULT FILE STORAGE, set FILEBLOB_HMAC_SECRET for ANY production usage")
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
