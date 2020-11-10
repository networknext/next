package main

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
	"github.com/gorilla/rpc/v2"
	"github.com/gorilla/rpc/v2/json2"
	"gopkg.in/auth0.v4/management"

	gcplogging "cloud.google.com/go/logging"
	"cloud.google.com/go/profiler"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/networknext/backend/backend"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/logging"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport"
	"github.com/networknext/backend/transport/jsonrpc"
	"github.com/networknext/backend/transport/middleware"
)

var (
	buildtime     string
	commitMessage string
	sha           string
	tag           string
)

const (
	MAILCHIMP_SERVER_PREFIX = "us20"
	MAILCHIMP_LIST_ID       = "553903bc6f"
)

func main() {
	fmt.Printf("portal: Git Hash: %s - Commit: %s\n", sha, commitMessage)

	ctx := context.Background()

	// Configure logging
	logger := log.NewLogfmtLogger(os.Stdout)

	// Stackdriver Logging
	{
		var enableSDLogging bool
		enableSDLoggingString, ok := os.LookupEnv("ENABLE_STACKDRIVER_LOGGING")
		if ok {
			var err error
			enableSDLogging, err = strconv.ParseBool(enableSDLoggingString)
			if err != nil {
				level.Error(logger).Log("envvar", "ENABLE_STACKDRIVER_LOGGING", "msg", "could not parse", "err", err)
				os.Exit(1)
			}
		}

		if enableSDLogging {
			if projectID, ok := os.LookupEnv("GOOGLE_PROJECT_ID"); ok {
				loggingClient, err := gcplogging.NewClient(ctx, projectID)
				if err != nil {
					level.Error(logger).Log("err", err)
					os.Exit(1)
				}

				logger = logging.NewStackdriverLogger(loggingClient, "portal")
			}
		}
	}

	{
		switch os.Getenv("BACKEND_LOG_LEVEL") {
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

	// Get env
	env, ok := os.LookupEnv("ENV")
	if !ok {
		level.Error(logger).Log("err", "ENV not set")
		os.Exit(1)
	}

	var relayPublicKey []byte
	var customerID uint64
	var customerPublicKey []byte
	{
		if key := os.Getenv("RELAY_PUBLIC_KEY"); len(key) != 0 {
			relayPublicKey, _ = base64.StdEncoding.DecodeString(key)
		}

		if key := os.Getenv("NEXT_CUSTOMER_PUBLIC_KEY"); len(key) != 0 {
			customerPublicKey, _ = base64.StdEncoding.DecodeString(key)
			customerID = binary.LittleEndian.Uint64(customerPublicKey[:8])
			customerPublicKey = customerPublicKey[8:]
		}
	}

	redisPoolTopSessions := storage.NewRedisPool(os.Getenv("REDIS_HOST_TOP_SESSIONS"), 5, 64)
	if err := storage.ValidateRedisPool(redisPoolTopSessions); err != nil {
		level.Error(logger).Log("envvar", "REDIS_HOST_TOP_SESSIONS", "err", err)
		os.Exit(1)
	}

	redisPoolSessionMap := storage.NewRedisPool(os.Getenv("REDIS_HOST_SESSION_MAP"), 5, 64)
	if err := storage.ValidateRedisPool(redisPoolSessionMap); err != nil {
		level.Error(logger).Log("envvar", "REDIS_HOST_SESSION_MAP", "err", err)
		os.Exit(1)
	}

	redisPoolSessionMeta := storage.NewRedisPool(os.Getenv("REDIS_HOST_SESSION_META"), 5, 64)
	if err := storage.ValidateRedisPool(redisPoolSessionMeta); err != nil {
		level.Error(logger).Log("envvar", "REDIS_HOST_SESSION_META", "err", err)
		os.Exit(1)
	}

	redisPoolSessionSlices := storage.NewRedisPool(os.Getenv("REDIS_HOST_SESSION_SLICES"), 5, 64)
	if err := storage.ValidateRedisPool(redisPoolSessionSlices); err != nil {
		level.Error(logger).Log("envvar", "REDIS_HOST_SESSION_SLICES", "err", err)
		os.Exit(1)
	}

	var db storage.Storer = &storage.InMemory{
		LocalMode: true,
	}

	manager, err := management.New(
		os.Getenv("AUTH_DOMAIN"),
		os.Getenv("AUTH_CLIENTID"),
		os.Getenv("AUTH_CLIENTSECRET"),
	)
	if err != nil {
		level.Error(logger).Log("err", err)
		os.Exit(1)
	}

	var userManager storage.UserManager = manager.User
	var jobManager storage.JobManager = manager.Job

	gcpProjectID, gcpOK := os.LookupEnv("GOOGLE_PROJECT_ID")
	_, emulatorOK := os.LookupEnv("FIRESTORE_EMULATOR_HOST")
	if emulatorOK {
		gcpProjectID = "local"

		level.Info(logger).Log("msg", "Detected firestore emulator")
	}

	if gcpOK || emulatorOK {
		// Firestore
		{
			// Create a Firestore Storer
			fs, err := storage.NewFirestore(ctx, gcpProjectID, logger)
			if err != nil {
				level.Error(logger).Log("err", err)
				os.Exit(1)
			}

			fssyncinterval := os.Getenv("GOOGLE_FIRESTORE_SYNC_INTERVAL")
			syncInterval, err := time.ParseDuration(fssyncinterval)
			if err != nil {
				level.Error(logger).Log("envvar", "GOOGLE_FIRESTORE_SYNC_INTERVAL", "value", fssyncinterval, "err", err)
				os.Exit(1)
			}
			// Start a goroutine to sync from Firestore
			go func() {
				ticker := time.NewTicker(syncInterval)
				fs.SyncLoop(ctx, ticker.C)
			}()

			// Set the Firestore Storer to give to handlers
			db = fs
		}
	}

	// Setup Bigtable

	btEmulatorOK := envvar.Exists("BIGTABLE_EMULATOR_HOST")
	if btEmulatorOK {
		// Emulator is used for local testing
		// Requires that emulator has been started in another terminal to work as intended
		gcpProjectID = "local"
		level.Info(logger).Log("msg", "Detected bigtable emulator")
	}

	useBigtable, err := envvar.GetBool("FEATURE_BIGTABLE", false)
	if err != nil {
		level.Error(logger).Log("err", err)
		os.Exit(1)
	}

	var btClient *storage.BigTable

	if useBigtable && (gcpOK || btEmulatorOK) {
		// Get Bigtable instance ID
		btInstanceID := envvar.Get("BIGTABLE_INSTANCE_ID", "")

		// Get the table name
		btTableName := envvar.Get("BIGTABLE_TABLE_NAME", "")

		// Create a bigtable admin for setup
		btAdmin, err := storage.NewBigTableAdmin(ctx, gcpProjectID, btInstanceID, logger)
		if err != nil {
			level.Error(logger).Log("err", err)
			os.Exit(1)
		}

		// Check if the table exists in the instance
		tableExists, err := btAdmin.VerifyTableExists(ctx, btTableName)
		if err != nil {
			level.Error(logger).Log("err", err)
			os.Exit(1)
		}

		if !tableExists {
			level.Error(logger).Log("envvar", "BIGTABLE_TABLE_NAME", "msg", "Table does not exist in Bigtable instance")
			os.Exit(1)
		}

		// Close the admin client
		if err = btAdmin.Close(); err != nil {
			level.Error(logger).Log("err", err)
			os.Exit(1)
		}

		// Create a standard client for writing to the table
		btClient, err = storage.NewBigTable(ctx, gcpProjectID, btInstanceID, logger)
		if err != nil {
			level.Error(logger).Log("err", err)
			os.Exit(1)
		}
	}

	metricsHandler, err := backend.GetMetricsHandler(ctx, logger, gcpProjectID)
	if err != nil {
		level.Error(logger).Log("err", err)
		os.Exit(1)
	}

	btMetrics, err := metrics.NewBigTableMetrics(ctx, metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create bigtable metrics", "err", err)
		os.Exit(1)
	}

	// Configure all GCP related services if the GOOGLE_PROJECT_ID is set
	// GCP VMs actually get populated with the GOOGLE_APPLICATION_CREDENTIALS
	// on creation so we can use that for the default then
	if gcpOK {
		// Stackdriver Profiler
		{
			var enableSDProfiler bool
			enableSDProfilerString, ok := os.LookupEnv("ENABLE_STACKDRIVER_PROFILER")
			if ok {
				enableSDProfiler, err = strconv.ParseBool(enableSDProfilerString)
				if err != nil {
					level.Error(logger).Log("envvar", "ENABLE_STACKDRIVER_PROFILER", "msg", "could not parse", "err", err)
					os.Exit(1)
				}
			}

			if enableSDProfiler {
				// Set up StackDriver profiler
				if err := profiler.Start(profiler.Config{
					Service:        "portal",
					ServiceVersion: env,
					ProjectID:      gcpProjectID,
					MutexProfiling: true,
				}); err != nil {
					level.Error(logger).Log("msg", "Failed to initialze StackDriver profiler", "err", err)
					os.Exit(1)
				}
			}
		}
	}

	if env == "local" {
		storage.SeedStorage(logger, ctx, db, relayPublicKey, customerID, customerPublicKey)
	}
	// We're not using the route matrix in the portal anymore, because RouteSelection()
	// is commented out in the ops service.

	// routeMatrix := &routing.RouteMatrix{}
	// var routeMatrixMutex sync.RWMutex

	// {
	// 	if uri, ok := os.LookupEnv("ROUTE_MATRIX_URI"); ok {
	// 		rmsyncinterval := os.Getenv("ROUTE_MATRIX_SYNC_INTERVAL")
	// 		syncInterval, err := time.ParseDuration(rmsyncinterval)
	// 		if err != nil {
	// 			level.Error(logger).Log("envvar", "ROUTE_MATRIX_SYNC_INTERVAL", "value", rmsyncinterval, "err", err)
	// 			os.Exit(1)
	// 		}

	// 		go func() {
	// 			for {
	// 				newRouteMatrix := &routing.RouteMatrix{}
	// 				var matrixReader io.Reader

	// 				// Default to reading route matrix from file
	// 				if f, err := os.Open(uri); err == nil {
	// 					matrixReader = f
	// 				}

	// 				// Prefer to get it remotely if possible
	// 				if r, err := http.Get(uri); err == nil {
	// 					matrixReader = r.Body
	// 				}

	// 				// Don't swap route matrix if we fail to read
	// 				_, err := newRouteMatrix.ReadFrom(matrixReader)
	// 				if err != nil {
	// 					level.Warn(logger).Log("matrix", "route", "op", "read", "envvar", "ROUTE_MATRIX_URI", "value", uri, "err", err, "msg", "forcing empty route matrix to avoid stale routes")
	// 					time.Sleep(syncInterval)
	// 					continue
	// 				}

	// 				// Swap the route matrix pointer to the new one
	// 				// This double buffered route matrix approach makes the route matrix lockless
	// 				routeMatrixMutex.Lock()
	// 				routeMatrix = newRouteMatrix
	// 				routeMatrixMutex.Unlock()

	// 				time.Sleep(syncInterval)
	// 			}
	// 		}()
	// 	}
	// }

	// Generate Sessions Map Points periodically
	buyerService := jsonrpc.BuyersService{
		BigTable:               btClient,
		BigTableMetrics:        btMetrics,
		Logger:                 logger,
		RedisPoolTopSessions:   redisPoolTopSessions,
		RedisPoolSessionMeta:   redisPoolSessionMeta,
		RedisPoolSessionSlices: redisPoolSessionSlices,
		RedisPoolSessionMap:    redisPoolSessionMap,
		Storage:                db,
	}

	go func() {
		genmapinterval := os.Getenv("SESSION_MAP_INTERVAL")
		syncInterval, err := time.ParseDuration(genmapinterval)
		if err != nil {
			level.Error(logger).Log("envvar", "SESSION_MAP_INTERVAL", "value", genmapinterval, "err", err)
			os.Exit(1)
		}

		for {
			if err := buyerService.GenerateMapPointsPerBuyer(); err != nil {
				level.Error(logger).Log("msg", "error generating sessions map points", "err", err)
				os.Exit(1)
			}
			time.Sleep(syncInterval)
		}
	}()

	uiDir := os.Getenv("UI_DIR")
	if uiDir == "" {
		level.Error(logger).Log("err", "env var UI_DIR must be set")
		os.Exit(1)
	}

	relayMap := jsonrpc.NewRelayStatsMap()

	go func() {
		relayStatsURL := os.Getenv("RELAY_STATS_URI")

		sleepInterval := time.Second
		if siStr, ok := os.LookupEnv("RELAY_STATS_SYNC_SLEEP_INTERVAL"); ok {
			if si, err := time.ParseDuration(siStr); err == nil {
				sleepInterval = si
			} else {
				level.Error(logger).Log("msg", "could not parse stats sync sleep interval", "err", err)
			}
		}

		for {
			time.Sleep(sleepInterval)

			res, err := http.Get(relayStatsURL)
			if err != nil {
				level.Error(logger).Log("msg", "unable to get relay stats", "err", err)
				continue
			}

			if res.StatusCode != http.StatusOK {
				level.Error(logger).Log("msg", "bad relay_stats request")
				continue
			}

			if res.ContentLength == -1 {
				level.Error(logger).Log("msg", fmt.Sprintf("relay_stats content length invalid: %d\n", res.ContentLength))
				res.Body.Close()
				continue
			}

			data, err := ioutil.ReadAll(res.Body)
			if err != nil {
				level.Error(logger).Log("msg", "unable to read response body", "err", err)
				res.Body.Close()
				continue
			}
			res.Body.Close()

			if err := relayMap.ReadAndSwap(data); err != nil {
				level.Error(logger).Log("msg", "unable to read relay stats map", "err", err)
			}
		}
	}()

	go func() {
		port, ok := os.LookupEnv("PORT")
		if !ok {
			level.Error(logger).Log("err", "env var PORT must be set")
			os.Exit(1)
		}

		level.Info(logger).Log("msg", fmt.Sprintf("Starting portal on port %s", port))

		s := rpc.NewServer()
		s.RegisterInterceptFunc(func(i *rpc.RequestInfo) *http.Request {
			user := i.Request.Context().Value(jsonrpc.Keys.UserKey)
			if user != nil {
				claims := user.(*jwt.Token).Claims.(jwt.MapClaims)

				if requestData, ok := claims["https://networknext.com/userData"]; ok {
					var userRoles []string
					if roles, ok := requestData.(map[string]interface{})["roles"]; ok {
						rolesInterface := roles.([]interface{})
						userRoles = make([]string, len(rolesInterface))
						for i, v := range rolesInterface {
							userRoles[i] = v.(string)
						}
					}
					var companyCode string
					if companyCodeInterface, ok := requestData.(map[string]interface{})["company_code"]; ok {
						companyCode = companyCodeInterface.(string)
					}
					var newsletterConsent bool
					if consent, ok := requestData.(map[string]interface{})["newsletter"]; ok {
						newsletterConsent = consent.(bool)
					}
					return jsonrpc.AddTokenContext(i.Request, userRoles, companyCode, newsletterConsent)
				}
			}
			return jsonrpc.SetIsAnonymous(i.Request, i.Request.Header.Get("Authorization") == "")
		})
		s.RegisterCodec(json2.NewCodec(), "application/json")
		s.RegisterService(&jsonrpc.OpsService{
			Logger:    logger,
			Release:   tag,
			BuildTime: buildtime,
			Storage:   db,
			RelayMap:  &relayMap,
			// RouteMatrix: &routeMatrix,
		}, "")
		s.RegisterService(&buyerService, "")
		s.RegisterService(&jsonrpc.AuthService{
			MailChimpManager: transport.MailChimpHandler{
				HTTPHandler: *http.DefaultClient,
				MembersURI:  fmt.Sprintf("https://%s.api.mailchimp.com/3.0/lists/%s/members", MAILCHIMP_SERVER_PREFIX, MAILCHIMP_LIST_ID),
			},
			Logger:      logger,
			UserManager: userManager,
			JobManager:  jobManager,
			Storage:     db,
		}, "")

		allowCORSStr := os.Getenv("CORS")
		allowCORS := true
		if ok, err := strconv.ParseBool(allowCORSStr); err == nil {
			allowCORS = ok
		}

		r := mux.NewRouter()

		r.Handle("/rpc", jsonrpc.AuthMiddleware(os.Getenv("JWT_AUDIENCE"), handlers.CompressHandler(s), allowCORS))
		r.HandleFunc("/health", transport.HealthHandlerFunc())
		r.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage))

		spa := spaHandler{staticPath: uiDir, indexPath: "index.html"}

		r.PathPrefix("/").Handler(middleware.CacheControl(os.Getenv("HTTP_CACHE_CONTROL"), handlers.CompressHandler(spa)))

		level.Info(logger).Log("addr", ":"+port)

		// If the port is set to 443 then build the certificates and run a TLS-enabled HTTP server
		if port == "443" {
			cert, err := tls.X509KeyPair(transport.TLSCertificate, transport.TLSPrivateKey)
			if err != nil {
				level.Error(logger).Log("err", err)
				os.Exit(1)
			}

			server := &http.Server{
				Addr:      ":" + port,
				TLSConfig: &tls.Config{Certificates: []tls.Certificate{cert}},
				Handler:   r,
			}

			err = server.ListenAndServeTLS("", "")
			if err != nil {
				level.Error(logger).Log("err", err)
				os.Exit(1)
			}
		}

		// Fall through to running on any other port defined with TLS disabled
		err := http.ListenAndServe(":"+port, r)
		if err != nil {
			level.Error(logger).Log("err", err)
			os.Exit(1)
		}
	}()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint
}

// spaHandler implements the http.Handler interface, so we can use it
// to respond to HTTP requests. The path to the static directory and
// path to the index file within that static directory are used to
// serve the SPA in the given static directory.
type spaHandler struct {
	staticPath string
	indexPath  string
}

// ServeHTTP inspects the URL path to locate a file within the static dir
// on the SPA handler. If a file is found, it will be served. If not, the
// file located at the index path on the SPA handler will be served. This
// is suitable behavior for serving an SPA (single page application).
func (h spaHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// get the absolute path to prevent directory traversal
	path, err := filepath.Abs(r.URL.Path)
	if err != nil {
		// if we failed to get the absolute path respond with a 400 bad request
		// and stop
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// prepend the path with the path to the static directory
	path = filepath.Join(h.staticPath, path)

	// check whether a file exists at the given path
	_, err = os.Stat(path)
	if os.IsNotExist(err) {
		// file does not exist, serve index.html
		http.ServeFile(w, r, filepath.Join(h.staticPath, h.indexPath))
		return
	} else if err != nil {
		// if we got an error (that wasn't that the file doesn't exist) stating the
		// file, return a 500 internal server error and stop
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// otherwise, use http.FileServer to serve the static dir
	http.FileServer(http.Dir(h.staticPath)).ServeHTTP(w, r)
}
