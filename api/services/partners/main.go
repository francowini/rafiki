package main

import (
	"context"
	"encoding/hex"
	"errors"
	"expvar"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/ardanlabs/conf/v3"

	"github.com/francowini/rafiki/api/services/partners/all"
	"github.com/francowini/rafiki/app/sdk/auth"
	"github.com/francowini/rafiki/app/sdk/debug"
	"github.com/francowini/rafiki/app/sdk/mux"
	"github.com/francowini/rafiki/business/domain/momentbus"
	"github.com/francowini/rafiki/business/domain/momentbus/stores/momentdb"
	"github.com/francowini/rafiki/business/domain/thinkbus"
	"github.com/francowini/rafiki/business/domain/thinkbus/stores/thinkdb"
	"github.com/francowini/rafiki/business/domain/userbus"
	"github.com/francowini/rafiki/business/domain/userbus/stores/userdb"
	"github.com/francowini/rafiki/business/domain/valuebus"
	"github.com/francowini/rafiki/business/domain/valuebus/stores/valuedb"
	"github.com/francowini/rafiki/business/sdk/delegate"
	"github.com/francowini/rafiki/business/sdk/encrypt"
	"github.com/francowini/rafiki/business/sdk/migrate"
	"github.com/francowini/rafiki/business/sdk/sqldb"
	"github.com/francowini/rafiki/foundation/keystore"
	"github.com/francowini/rafiki/foundation/logger"
	"github.com/francowini/rafiki/foundation/otel"
)

var (
	build  = "develop"
	routes = "all" // go build -ldflags "-X main.routes=crud"
)

func main() {
	var log *logger.Logger

	events := logger.Events{
		Error: func(ctx context.Context, r logger.Record) {
			log.Info(ctx, "******* SEND ALERT *******")
		},
	}

	traceIDFn := func(ctx context.Context) string {
		return otel.GetTraceID(ctx)
	}

	log = logger.NewWithEvents(os.Stdout, logger.LevelInfo, "RAFIKI", traceIDFn, events)

	// -------------------------------------------------------------------------

	ctx := context.Background()

	if err := run(ctx, log); err != nil {
		log.Error(ctx, "startup", "err", err)
		os.Exit(1)
	}
}

func run(ctx context.Context, log *logger.Logger) error {
	// -------------------------------------------------------------------------
	// GOMAXPROCS

	log.Info(ctx, "startup", "GOMAXPROCS", runtime.GOMAXPROCS(0))

	// -------------------------------------------------------------------------
	// Configuration

	cfg := struct {
		conf.Version
		Web struct {
			ReadTimeout        time.Duration `conf:"default:5s"`
			WriteTimeout       time.Duration `conf:"default:10s"`
			IdleTimeout        time.Duration `conf:"default:120s"`
			ShutdownTimeout    time.Duration `conf:"default:20s"`
			APIHost            string        `conf:"default:0.0.0.0:3000"`
			DebugHost          string        `conf:"default:0.0.0.0:3010"`
			CORSAllowedOrigins []string      `conf:"default:*"`
		}
		DB struct {
			User         string `conf:"default:postgres"`
			Password     string `conf:"default:postgres,mask"`
			Host         string `conf:"default:database-service"`
			Port         string `conf:"default:5432"`
			Name         string `conf:"default:postgres"`
			MaxIdleConns int    `conf:"default:0"`
			MaxOpenConns int    `conf:"default:0"`
			DisableTLS   bool   `conf:"default:false"`
			SSLMode      string `conf:"default:require,env:DB_SSLMODE"`
		}
		Encryption struct {
			Key string `conf:"required,mask"`
		}
		Tempo struct {
			Host        string  `conf:"default:tempo:4317"`
			ServiceName string  `conf:"default:rafiki-service"`
			Probability float64 `conf:"default:1.0"`
			// Shouldn't use a high Probability value in non-developer systems.
			// 0.05 should be enough for most systems. Some might want to have
			// this even lower.
		}
	}{
		Version: conf.Version{
			Build: build,
			Desc:  "Partner",
		},
	}

	const prefix = "Partner"
	help, err := conf.Parse(prefix, &cfg)
	if err != nil {
		if errors.Is(err, conf.ErrHelpWanted) {
			fmt.Println(help)
			return nil
		}
		return fmt.Errorf("parsing config: %w", err)
	}

	// -------------------------------------------------------------------------
	// App Starting

	log.Info(ctx, "starting service", "version", cfg.Build)
	defer log.Info(ctx, "shutdown complete")

	out, err := conf.String(&cfg)
	if err != nil {
		return fmt.Errorf("generating config for output: %w", err)
	}
	log.Info(ctx, "startup", "config", out)

	log.BuildInfo(ctx)

	expvar.NewString("build").Set(cfg.Build)

	// -------------------------------------------------------------------------
	// Database Support

	log.Info(ctx, "startup", "status", "initializing database support", "hostport", cfg.DB.Host)

	db, err := sqldb.Open(sqldb.Config{
		User:         cfg.DB.User,
		Password:     cfg.DB.Password,
		Host:         cfg.DB.Host,
		Port:         cfg.DB.Port,
		Name:         cfg.DB.Name,
		MaxIdleConns: cfg.DB.MaxIdleConns,
		MaxOpenConns: cfg.DB.MaxOpenConns,
		DisableTLS:   cfg.DB.DisableTLS,
		SSLMode:      cfg.DB.SSLMode,
	})
	if err != nil {
		return fmt.Errorf("connecting to db: %w", err)
	}

	defer func() {
		if err := db.Close(); err != nil {
			log.Error(ctx, "shutdown", "status", "closing database", "err", err)
		}
	}()

	// -------------------------------------------------------------------------
	// Run Database Migrations

	log.Info(ctx, "startup", "status", "running database migrations")

	if err := migrate.Migrate(ctx, db); err != nil {
		return fmt.Errorf("running migrations: %w", err)
	}

	log.Info(ctx, "startup", "status", "database migrations completed")

	// -------------------------------------------------------------------------
	// Initialize Encryption Support

	log.Info(ctx, "startup", "status", "initializing encryption")

	if len(cfg.Encryption.Key) != 64 {
		return fmt.Errorf("encryption key must be 64 hex characters (32 bytes), got %d characters", len(cfg.Encryption.Key))
	}

	encryptionKey, err := hex.DecodeString(cfg.Encryption.Key)
	if err != nil {
		return fmt.Errorf("decode encryption key (must be valid hex): %w", err)
	}

	encryptor, err := encrypt.NewAESEncryptor(encryptionKey)
	if err != nil {
		return fmt.Errorf("create encryptor: %w", err)
	}

	// Clear key from memory after encryptor is created
	for i := range encryptionKey {
		encryptionKey[i] = 0
	}

	log.Info(ctx, "startup", "status", "encryption initialized", "algorithm", "AES-256-GCM")

	// -------------------------------------------------------------------------
	// Create Delegate for Cross-Domain Events

	dlg := delegate.New(log)

	// -------------------------------------------------------------------------
	// Create Business Packages

	// Independent domains (no parent needed)
	thinkStore := thinkdb.NewStore(log, db, encryptor)
	thinkBus := thinkbus.NewBusiness(log, thinkStore)

	momentStore := momentdb.NewStore(log, db, encryptor)
	momentBus := momentbus.NewBusiness(log, momentStore)

	// Create userbus first (root domain) - needs delegate for publishing events
	userStore := userdb.NewStore(log, db)
	userBus := userbus.NewBusiness(log, dlg, userStore)

	// Create child domains with parent injection
	valueStore := valuedb.NewStore(log, db, encryptor)
	valueBus := valuebus.NewBusiness(log, userBus, dlg, valueStore)

	// -------------------------------------------------------------------------
	// Initialize authentication support

	log.Info(ctx, "startup", "status", "initializing authentication support")

	// Load RSA keys for JWT signing/verification
	ks := keystore.New()

	// Load from filesystem (development and production)
	keysLoaded := 0

	// Try loading from zarf/keys directory
	keysFS := os.DirFS("./zarf/keys")
	keysLoaded, err = ks.LoadByFileSystem(keysFS)
	if err != nil {
		return fmt.Errorf("loading auth keys from filesystem: %w", err)
	}

	if keysLoaded == 0 {
		return fmt.Errorf("no authentication keys loaded - cannot start service")
	}

	log.Info(ctx, "startup", "keys_loaded", keysLoaded)

	// Initialize Auth (userBus already created above)
	authInstance := auth.New(auth.Config{
		Log:       log,
		UserBus:   userBus,
		KeyLookup: ks,
		Issuer:    "rafiki-service",
	})

	log.Info(ctx, "startup", "status", "authentication support enabled")

	// -------------------------------------------------------------------------
	// Start Tracing Support

	log.Info(ctx, "startup", "status", "initializing tracing support")

	traceProvider, teardown, err := otel.InitTracing(log, otel.Config{
		ServiceName: cfg.Tempo.ServiceName,
		Host:        cfg.Tempo.Host,
		ExcludedRoutes: map[string]struct{}{
			"/v1/readiness": {},
		},
		Probability: cfg.Tempo.Probability,
	})
	if err != nil {
		return fmt.Errorf("starting tracing: %w", err)
	}

	defer teardown(context.Background())

	tracer := traceProvider.Tracer(cfg.Tempo.ServiceName)

	// -------------------------------------------------------------------------
	// Start Debug Service

	go func() {
		log.Info(ctx, "startup", "status", "debug v1 router started", "host", cfg.Web.DebugHost)

		if err := http.ListenAndServe(cfg.Web.DebugHost, debug.Mux()); err != nil {
			log.Error(ctx, "shutdown", "status", "debug v1 router closed", "host", cfg.Web.DebugHost, "msg", err)
		}
	}()

	// -------------------------------------------------------------------------
	// Start API Service

	log.Info(ctx, "startup", "status", "initializing V1 API support")

	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, syscall.SIGINT, syscall.SIGTERM)

	cfgMux := mux.Config{
		Build:  build,
		Log:    log,
		DB:     db,
		Tracer: tracer,
		BusConfig: mux.BusConfig{
			ThinkBus:  thinkBus,
			MomentBus: momentBus,
			ValueBus:  valueBus,
			UserBus:   userBus,
			Auth:      authInstance,
		},
	}

	webAPI := mux.WebAPI(cfgMux,
		buildRoutes(),
		mux.WithCORS(cfg.Web.CORSAllowedOrigins),
	)

	api := http.Server{
		Addr:         cfg.Web.APIHost,
		Handler:      webAPI,
		ReadTimeout:  cfg.Web.ReadTimeout,
		WriteTimeout: cfg.Web.WriteTimeout,
		IdleTimeout:  cfg.Web.IdleTimeout,
		ErrorLog:     logger.NewStdLogger(log, logger.LevelError),
	}

	serverErrors := make(chan error, 1)

	go func() {
		log.Info(ctx, "startup", "status", "api router started", "host", api.Addr)

		serverErrors <- api.ListenAndServe()
	}()

	// -------------------------------------------------------------------------
	// Shutdown

	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		log.Info(ctx, "shutdown", "status", "shutdown started", "signal", sig)
		defer log.Info(ctx, "shutdown", "status", "shutdown complete", "signal", sig)

		ctx, cancel := context.WithTimeout(ctx, cfg.Web.ShutdownTimeout)
		defer cancel()

		if err := api.Shutdown(ctx); err != nil {
			if closeErr := api.Close(); closeErr != nil {
				log.Error(ctx, "shutdown", "status", "force closing server", "err", closeErr)
			}
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
	}

	return nil
}

func buildRoutes() mux.RouteAdder {
	// The idea here is that we can build different versions of the binary
	// with different sets of exposed web APIs. By default we build a single
	// instance with all the web APIs.
	//
	// Here is the scenario. It would be nice to build two binaries, one for the
	// transactional APIs (CRUD) and one for the reporting APIs. This would allow
	// the system to run two instances of the database. One instance tuned for the
	// transactional database calls and the other tuned for the reporting calls.
	// Tuning meaning indexing and memory requirements. The two databases can be
	// kept in sync with replication.

	if routes != "all" {
		panic("unexpected route")
	}

	return all.Routes()
}
