package main

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/handlers"
	"github.com/gorilla/rpc/v2"
	"github.com/gorilla/rpc/v2/json2"
	"gopkg.in/auth0.v4/management"

	gcplogging "cloud.google.com/go/logging"
	"cloud.google.com/go/profiler"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/logging"
	"github.com/networknext/backend/routing"
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

	redisPortalHosts := os.Getenv("REDIS_HOST_PORTAL")
	splitPortalHosts := strings.Split(redisPortalHosts, ",")
	redisClientPortal := storage.NewRedisClient(splitPortalHosts...)
	if err := redisClientPortal.Ping().Err(); err != nil {
		level.Error(logger).Log("envvar", "REDIS_HOST_PORTAL", "value", redisPortalHosts, "err", err)
		os.Exit(1)
	}

	redisRelayHost := os.Getenv("REDIS_HOST_RELAYS")
	redisClientRelays := storage.NewRedisClient(redisRelayHost)
	if err := redisClientRelays.Ping().Err(); err != nil {
		level.Error(logger).Log("envvar", "REDIS_HOST_RELAYS", "value", redisRelayHost, "err", err)
		os.Exit(1)
	}

	var db storage.Storer = &storage.InMemory{
		LocalMode: true,
	}

	{
		if err := db.AddBuyer(ctx, routing.Buyer{
			ID:                   customerID,
			Name:                 "local",
			PublicKey:            customerPublicKey,
			RoutingRulesSettings: routing.LocalRoutingRulesSettings,
		}); err != nil {
			level.Error(logger).Log("msg", "could not add buyer to storage", "err", err)
			os.Exit(1)
		}

		seller := routing.Seller{
			ID:                        "sellerID",
			Name:                      "local",
			IngressPriceNibblinsPerGB: 0.1 * 1e9,
			EgressPriceNibblinsPerGB:  0.2 * 1e9,
		}

		datacenter := routing.Datacenter{
			ID:           crypto.HashID("local"),
			Name:         "local",
			SupplierName: "usw2-az4",
		}

		if err := db.AddSeller(ctx, seller); err != nil {
			level.Error(logger).Log("msg", "could not add seller to storage", "err", err)
			os.Exit(1)
		}

		if err := db.AddDatacenter(ctx, datacenter); err != nil {
			level.Error(logger).Log("msg", "could not add datacenter to storage", "err", err)
			os.Exit(1)
		}

		addr1 := net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 10000}
		if err := db.AddRelay(ctx, routing.Relay{
			Name:           "local.test_relay.a",
			ID:             crypto.HashID(addr1.String()),
			Addr:           addr1,
			PublicKey:      relayPublicKey,
			Seller:         seller,
			Datacenter:     datacenter,
			ManagementAddr: "127.0.0.1",
			SSHUser:        "root",
			SSHPort:        22,
			MRC:            19700000000000,
			Overage:        26000000000000,
			BWRule:         routing.BWRuleBurst,
			ContractTerm:   12,
			StartDate:      time.Now(),
			EndDate:        time.Now(),
			Type:           routing.BareMetal,
		}); err != nil {
			level.Error(logger).Log("msg", "could not add relay to storage", "err", err)
			os.Exit(1)
		}

		addr2 := net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 10001}
		if err := db.AddRelay(ctx, routing.Relay{
			Name:           "local.test_relay.b",
			ID:             crypto.HashID(addr2.String()),
			Addr:           addr2,
			PublicKey:      relayPublicKey,
			Seller:         seller,
			Datacenter:     datacenter,
			ManagementAddr: "127.0.0.1",
			SSHUser:        "root",
			SSHPort:        22,
		}); err != nil {
			level.Error(logger).Log("msg", "could not add relay to storage", "err", err)
			os.Exit(1)
		}

		addr3 := net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 10002}
		if err := db.AddRelay(ctx, routing.Relay{
			Name:           "abc.xyz",
			ID:             crypto.HashID(addr3.String()),
			Addr:           addr3,
			PublicKey:      relayPublicKey,
			Seller:         seller,
			Datacenter:     datacenter,
			ManagementAddr: "127.0.0.1",
			SSHUser:        "root",
			SSHPort:        22,
		}); err != nil {
			level.Error(logger).Log("msg", "could not add relay to storage", "err", err)
			os.Exit(1)
		}
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

	auth0Client := storage.Auth0{
		Manager: manager,
		Logger:  logger,
	}

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
		Logger:      logger,
		RedisClient: redisClientPortal,
		Storage:     db,
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

	go func() {
		port, ok := os.LookupEnv("PORT")
		if !ok {
			level.Error(logger).Log("err", "env var PORT must be set")
			os.Exit(1)
		}

		level.Info(logger).Log("msg", fmt.Sprintf("Starting portal on port %s", port))

		s := rpc.NewServer()
		s.RegisterInterceptFunc(func(i *rpc.RequestInfo) *http.Request {
			user := i.Request.Context().Value("user")
			if user != nil {
				claims := user.(*jwt.Token).Claims.(jwt.MapClaims)

				if requestRoles, ok := claims["https://networknext.com/userRoles"]; ok {
					if roles, ok := requestRoles.(map[string]interface{})["roles"]; ok {
						rolesInterface := roles.([]interface{})
						userRoles := make([]string, len(rolesInterface))
						for i, v := range rolesInterface {
							userRoles[i] = v.(string)
						}
						if len(userRoles) > 0 {
							return jsonrpc.SetRoles(i.Request, userRoles)
						}
					}
				}
			}
			return jsonrpc.SetIsAnonymous(i.Request, i.Request.Header.Get("Authorization") == "")
		})
		s.RegisterCodec(json2.NewCodec(), "application/json")
		s.RegisterService(&jsonrpc.OpsService{
			Logger:      logger,
			Release:     tag,
			BuildTime:   buildtime,
			RedisClient: redisClientRelays,
			Storage:     db,
			// RouteMatrix: &routeMatrix,
		}, "")
		s.RegisterService(&buyerService, "")
		s.RegisterService(&jsonrpc.AuthService{
			Logger:  logger,
			Auth0:   auth0Client,
			Storage: db,
		}, "")

		allowCORSStr := os.Getenv("CORS")
		allowCORS := true
		if ok, err := strconv.ParseBool(allowCORSStr); err != nil {
			allowCORS = ok
		}

		http.Handle("/rpc", jsonrpc.AuthMiddleware(os.Getenv("JWT_AUDIENCE"), handlers.CompressHandler(s), allowCORS))

		if allowCORS {
			http.Handle("/", middleware.CacheControl(os.Getenv("HTTP_CACHE_CONTROL"), http.FileServer(http.Dir(uiDir))))
		}

		http.HandleFunc("/health", transport.HealthHandlerFunc())
		http.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage))

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
			}

			err = server.ListenAndServeTLS("", "")
			if err != nil {
				level.Error(logger).Log("err", err)
				os.Exit(1)
			}
		}

		// Fall through to running on any other port defined with TLS disabled
		err := http.ListenAndServe(":"+port, nil)
		if err != nil {
			level.Error(logger).Log("err", err)
			os.Exit(1)
		}
	}()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint
}
