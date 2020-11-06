package backend

import (
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/logging"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/storage"

	gcplogging "cloud.google.com/go/logging"
	"cloud.google.com/go/profiler"
)

func GetEnv() (string, error) {
	if !envvar.Exists("ENV") {
		return "", errors.New("ENV not set")
	}

	return envvar.Get("ENV", ""), nil
}

func GetGCPProjectID() string {
	return envvar.Get("GOOGLE_PROJECT_ID", "")
}

// GetLogger returns a logger for the backend service to use.
// The logger supports four levels: "debug", "info", "warn", and "error". If the level is "none", no logs are outputted.
// It is guaranteed to always return at least a local logger, even if an error is returned.
// If a gcp project ID is specified, it will return a StackDriver logger.
func GetLogger(ctx context.Context, gcpProjectID string, serviceName string) (log.Logger, error) {
	logger := log.NewLogfmtLogger(os.Stdout)
	{
		backendLogLevel := envvar.Get("BACKEND_LOG_LEVEL", "none")
		switch backendLogLevel {
		case "none":
			logger = level.NewFilter(logger, level.AllowNone())
		case level.ErrorValue().String():
			logger = level.NewFilter(logger, level.AllowError())
		case level.WarnValue().String():
			logger = level.NewFilter(logger, level.AllowWarn())
		case level.InfoValue().String():
			logger = level.NewFilter(logger, level.AllowInfo())
		case level.DebugValue().String():
			logger = level.NewFilter(logger, level.AllowDebug())
		default:
			logger = level.NewFilter(logger, level.AllowWarn())
		}

		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	}

	if gcpProjectID != "" {
		enableSDLogging, err := envvar.GetBool("ENABLE_STACKDRIVER_LOGGING", false)
		if err != nil {
			return logger, err
		}

		if enableSDLogging {
			loggingClient, err := gcplogging.NewClient(ctx, gcpProjectID)
			if err != nil {
				return logger, fmt.Errorf("failed to create GCP logging client: %v", err)
			}

			logger = logging.NewStackdriverLogger(loggingClient, serviceName)
		}
	}

	return logger, nil
}

// GetMetricsHandler returns a metrics handler for the backend service to use.
// It is guaranteed to always return at least a local metrics handler, even if an error is returned.
// If a gcp project ID is specified, it will return a StackDriver metrics handler.
func GetMetricsHandler(ctx context.Context, logger log.Logger, gcpProjectID string) (metrics.Handler, error) {
	var metricsHandler metrics.Handler = &metrics.LocalHandler{}

	if gcpProjectID != "" {
		enableSDMetrics, err := envvar.GetBool("ENABLE_STACKDRIVER_METRICS", false)
		if err != nil {
			return metricsHandler, err
		}

		if enableSDMetrics {
			sd := metrics.StackDriverHandler{
				ProjectID:          gcpProjectID,
				OverwriteFrequency: time.Second,
				OverwriteTimeout:   10 * time.Second,
			}

			if err := sd.Open(ctx); err != nil {
				return metricsHandler, fmt.Errorf("failed to create StackDriver metrics client: %v", err)
			}

			sdWriteInterval, err := envvar.GetDuration("GOOGLE_STACKDRIVER_METRICS_WRITE_INTERVAL", time.Minute)
			if err != nil {
				return metricsHandler, err
			}

			go func() {
				metricsHandler.WriteLoop(ctx, logger, sdWriteInterval, 200)
			}()

			metricsHandler = &sd
		}
	}

	return metricsHandler, nil
}

func InitStackDriverProfiler(gcpProjectID string, serviceName string, env string) error {
	enableSDProfiler, err := envvar.GetBool("ENABLE_STACKDRIVER_PROFILER", false)
	if err != nil {
		return err
	}

	if enableSDProfiler {
		// Set up StackDriver profiler
		if err := profiler.Start(profiler.Config{
			Service:        serviceName,
			ServiceVersion: env,
			ProjectID:      gcpProjectID,
			MutexProfiling: true,
		}); err != nil {
			return err
		}
	}

	return nil
}

// GetStorer returns a storer for the backend service to use.
// It is guaranteed to always return at least a local storer, even if an error is returned.
// If a gcp project ID is specified, it will return a Firestore storer.
func GetStorer(ctx context.Context, logger log.Logger, gcpProjectID string, env string) (storage.Storer, error) {

	postgresOK := envvar.Exists("FEATURE_POSTGRESQL")
	if postgresOK {
		// returns a pointer to the storage solution specified for the provide environment. The
		// database targets are currently gcp/Firestore but will move to gcp/PostgreSQL
		//
		//	 prod: production database
		//    dev: development database
		//  local: local PostgreSQL install
		//  local: local SQLite file
		//
		// dev and prod connections to Cloud SQL will be made by proxy and are project-specific:
		// https://cloud.google.com/sql/docs/postgres/connect-compute-engine#gce-connect-proxy
		//
		// Environment Variables:
		// 	ENV=local|dev|prod|staging
		//		required
		//	GOOGLE_PROJECT_ID=network-next-v3-dev
		//		required for dev and prod services
		//	PGSQL=true|false
		//		for local usage
		//			true: overrides the default (SQLite)
		//         false: local testing/happy path will use SQLite
		// 	CGO_ENABLED=1
		//		required for the cgo go-sqlite3 package, but not currently used???
		//	DB_SYNC_INTERVAL
		var db storage.Storer

		pgsqlStr := envvar.Get("FEATURE_POSTGRESQL", "false")
		pgsql, err := strconv.ParseBool(pgsqlStr)
		if err != nil {
			return nil, fmt.Errorf("could not parse FEATURE_POSTGRESQL boolean: %v", err)
		}

		if pgsql {
			// TODO: the cloud version of PostgreSQL will require gcpProjectID, etc.
			level.Info(logger).Log("msg", "Setting up PostgreSQL storage")
			db, err = storage.NewPostgreSQL(ctx, logger)
			if err != nil {
				err := fmt.Errorf("NewPostgreSQL() error loading PostgreSQL: %w", err)
				return nil, err
			}
		} else {
			level.Info(logger).Log("msg", "Setting up SQLite storage")
			db, err = storage.NewSQLite3(ctx, logger)
			if err != nil {
				err := fmt.Errorf("NewSQLite3() error loading sqlite3: %w", err)
				return nil, err
			}
		}

		dbSyncInterval, err := envvar.GetDuration("GOOGLE_FIRESTORE_SYNC_INTERVAL", time.Second*10)
		if err != nil {
			return nil, err
		}

		// Start a goroutine to sync from Firestore
		go func() {
			ticker := time.NewTicker(dbSyncInterval)
			db.SyncLoop(ctx, ticker.C)
		}()
		return db, nil
	}

	// No SQL yet so use original Firestore / InMemory rig
	var storer storage.Storer = &storage.InMemory{
		LocalMode: true,
	}

	// Check for the firestore emulator
	firestoreEmulatorOK := envvar.Exists("FIRESTORE_EMULATOR_HOST")
	if firestoreEmulatorOK {
		gcpProjectID = "local"
		level.Info(logger).Log("msg", "Detected firestore emulator")
	}

	if gcpProjectID != "" || firestoreEmulatorOK {
		fs, err := storage.NewFirestore(ctx, gcpProjectID, logger)
		if err != nil {
			return storer, fmt.Errorf("could not create firestore: %v", err)
		}

		fsSyncInterval, err := envvar.GetDuration("GOOGLE_FIRESTORE_SYNC_INTERVAL", time.Second*10)
		if err != nil {
			return storer, err
		}

		// Start a goroutine to sync from Firestore
		go func() {
			ticker := time.NewTicker(fsSyncInterval)
			fs.SyncLoop(ctx, ticker.C)
		}()

		storer = fs
	}

	// Create dummy entries in storer for local testing
	if env == "local" {
		if !envvar.Exists("NEXT_CUSTOMER_PUBLIC_KEY") {
			return storer, errors.New("NEXT_CUSTOMER_PUBLIC_KEY not set")
		}

		customerPublicKey, err := envvar.GetBase64("NEXT_CUSTOMER_PUBLIC_KEY", nil)
		if err != nil {
			level.Error(logger).Log("err", err)
			return storer, err
		}
		customerID := binary.LittleEndian.Uint64(customerPublicKey[:8])

		if !envvar.Exists("RELAY_PUBLIC_KEY") {
			return storer, errors.New("RELAY_PUBLIC_KEY not set")
		}

		relayPublicKey, err := envvar.GetBase64("RELAY_PUBLIC_KEY", nil)
		if err != nil {
			return storer, err
		}

		// Create dummy buyer and datacenter for local testing
		storage.SeedStorage(ctx, storer, relayPublicKey, customerID, customerPublicKey)
	}

	return storer, nil
}
