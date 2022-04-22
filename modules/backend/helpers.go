package backend

import (
	"context"
	"encoding/binary"
	"encoding/gob"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"os"
	"sort"
	"strconv"
	"time"

	"cloud.google.com/go/compute/metadata"
	gcplogging "cloud.google.com/go/logging"
	"cloud.google.com/go/profiler"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/logging"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/storage"
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

func GetInstanceID(env string) (string, error) {
	if env != "local" {
		return metadata.InstanceID()
	}

	return "local", nil
}

// GetLogger returns a logger for the backend service to use.
// The logger supports four levels: "debug", "info", "warn", and "error". If the level is "none", no logs are outputted.
// It is guaranteed to always return at least a local logger, even if an error is returned.
// If a gcp project ID is specified, it will return a StackDriver logger.
func GetLogger(ctx context.Context, gcpProjectID string, serviceName string) (log.Logger, error) {
	logger := log.NewLogfmtLogger(os.Stdout)

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

// GetTSMetricsHandler returns a time series metrics handler for the backend service to use for data related to the Buyer ID.
// It is guaranteed to always return at least a local metrics handler, even if an error is returned.
// If a gcp project ID is specified, it will return a StackDriver metrics handler.
// It differs from GetMetricsHandler() based on how frequently it writes to StackDriver
func GetTSMetricsHandler(ctx context.Context, logger log.Logger, gcpProjectID string) (metrics.Handler, error) {
	var tsMetricsHandler metrics.Handler = &metrics.LocalHandler{}

	if gcpProjectID != "" {
		enableSDMetrics, err := envvar.GetBool("ENABLE_STACKDRIVER_METRICS", false)
		if err != nil {
			return tsMetricsHandler, err
		}

		if enableSDMetrics {
			sd := metrics.StackDriverHandler{
				ProjectID:          gcpProjectID,
				OverwriteFrequency: time.Second,
				OverwriteTimeout:   10 * time.Second,
			}

			if err := sd.Open(ctx); err != nil {
				return tsMetricsHandler, fmt.Errorf("failed to create Time Series StackDriver metrics client: %v", err)
			}

			sdWriteInterval, err := envvar.GetDuration("FEATURE_VANITY_METRIC_WRITE_INTERVAL", time.Second*5)
			if err != nil {
				return tsMetricsHandler, err
			}

			go func() {
				tsMetricsHandler.WriteLoop(ctx, logger, sdWriteInterval, 200)
			}()

			tsMetricsHandler = &sd
		}
	}

	return tsMetricsHandler, nil
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

		pgsql, err := envvar.GetBool("FEATURE_POSTGRESQL", false)
		if err != nil {
			return nil, fmt.Errorf("could not parse FEATURE_POSTGRESQL boolean: %v", err)
		}

		if pgsql {
			pgsqlHostIP := envvar.Get("POSTGRESQL_HOST_IP", "")
			if pgsqlHostIP == "" {
				return nil, fmt.Errorf("could not parse FEATURE_POSTGRESQL string: %v", err)
			}
			pgsqlUserName := envvar.Get("POSTGRESQL_USER_NAME", "")
			if pgsqlUserName == "" {
				return nil, fmt.Errorf("could not parse POSTGRESQL_USER_NAME string: %v", err)
			}
			pgsqlPassword := envvar.Get("POSTGRESQL_PASSWORD", "")
			if pgsqlPassword == "" {
				return nil, fmt.Errorf("could not parse POSTGRESQL_PASSWORD string: %v", err)
			}

			db, err = storage.NewPostgreSQL(ctx, logger, pgsqlHostIP, pgsqlUserName, pgsqlPassword)
			if err != nil {
				err := fmt.Errorf("NewPostgreSQL() error loading PostgreSQL: %w", err)
				return nil, err
			}
		} else if env == "staging" {
			db, err = storage.NewSQLite3Staging(ctx, logger)
			if err != nil {
				err := fmt.Errorf("NewSQLLite3Staging() error loading sqlite3: %w", err)
				return nil, err
			}
		} else {
			db, err = storage.NewSQLite3(ctx, logger)
			if err != nil {
				err := fmt.Errorf("NewSQLite3() error loading sqlite3: %w", err)
				return nil, err
			}
		}

		if env == "local" {
			customerPublicKey, err := envvar.GetBase64("NEXT_CUSTOMER_PUBLIC_KEY", nil)
			if err != nil {
				level.Error(logger).Log("err", err)
				return nil, err
			}
			customerID := binary.LittleEndian.Uint64(customerPublicKey[:8])
			customerPublicKey = customerPublicKey[8:]

			if !envvar.Exists("RELAY_PUBLIC_KEY") {
				return nil, errors.New("RELAY_PUBLIC_KEY not set")
			}

			relayPublicKey, err := envvar.GetBase64("RELAY_PUBLIC_KEY", nil)
			if err != nil {
				return nil, err
			}

			storage.SeedSQLStorage(ctx, db, relayPublicKey, customerID, customerPublicKey)
		}

		if env == "staging" && !pgsql {
			filePath := envvar.Get("BIN_PATH", "./database.bin")
			file, err := os.Open(filePath)
			if err != nil {
				return nil, fmt.Errorf("could not open database file: %v", err)
			}
			defer file.Close()

			database := routing.CreateEmptyDatabaseBinWrapper()

			if err = DecodeBinWrapper(file, database); err != nil {
				return nil, fmt.Errorf("failed to read database: %v", err)
			}

			// Seed staging storer
			if err = storage.SeedSQLStorageStaging(ctx, db, database); err != nil {
				return nil, fmt.Errorf("failed to seed sql storage for staging: %v", err)
			}
		}

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
		customerPublicKey = customerPublicKey[8:]

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

// Parses a string for a UDP address
func ParseAddress(input string) *net.UDPAddr {
	address := &net.UDPAddr{}
	ip_string, port_string, err := net.SplitHostPort(input)
	if err != nil {
		address.IP = net.ParseIP(input)
		address.Port = 0
		return address
	}
	address.IP = net.ParseIP(ip_string)
	address.Port, _ = strconv.Atoi(port_string)
	return address
}

// Decodes a Overlay Bin Wrapper from GOB
func DecodeOverlayWrapper(file *os.File, overlayWrapper *routing.OverlayBinWrapper) error {
	decoder := gob.NewDecoder(file)
	err := decoder.Decode(overlayWrapper)
	return err
}

// Decodes a Database Bin Wrapper from GOB
func DecodeBinWrapper(file *os.File, binWrapper *routing.DatabaseBinWrapper) error {
	decoder := gob.NewDecoder(file)
	err := decoder.Decode(binWrapper)
	return err
}

// Sorts a relay array and hash via Name
func SortAndHashRelayArray(relayArray []routing.Relay, relayHash map[uint64]routing.Relay) {
	sort.SliceStable(relayArray, func(i, j int) bool {
		return relayArray[i].Name < relayArray[j].Name
	})

	for i := range relayArray {
		relayHash[relayArray[i].ID] = relayArray[i]
	}
}

// Prints a list of relays in the relay array to console
func DisplayLoadedRelays(relayArray []routing.Relay) {
	fmt.Printf("\n=======================================\n")
	fmt.Printf("\nLoaded %d relays:\n\n", len(relayArray))
	for i := range relayArray {
		fmt.Printf("\t%s - %s [%x]\n", relayArray[i].Name, relayArray[i].Addr.String(), relayArray[i].ID)
	}
	fmt.Printf("\n=======================================\n")
}

// Helper function to create a random string of a specified length
// Useful for testing constant string lengths
// Adapted from: https://stackoverflow.com/a/22892986
func GenerateRandomStringSequence(length int) string {
	letters := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ")

	// Seed randomness
	rand.Seed(time.Now().UnixNano())

	b := make([]rune, length)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return string(b)
}
