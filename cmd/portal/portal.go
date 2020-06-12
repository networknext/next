package main

import (
	"context"
	"crypto/tls"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gorilla/rpc/v2"
	"github.com/gorilla/rpc/v2/json2"
	"gopkg.in/auth0.v4/management"

	gcplogging "cloud.google.com/go/logging"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/logging"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport"
	"github.com/networknext/backend/transport/jsonrpc"
)

var (
	release   string
	buildtime string
)

func main() {
	ctx := context.Background()

	// Configure logging
	logger := log.NewLogfmtLogger(os.Stdout)
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
	if projectID, ok := os.LookupEnv("GOOGLE_PROJECT_ID"); ok {
		loggingClient, err := gcplogging.NewClient(ctx, projectID)
		if err != nil {
			level.Error(logger).Log("err", err)
			os.Exit(1)
		}

		logger = logging.NewStackdriverLogger(loggingClient, "relay-backend")
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
			ID:                "sellerID",
			Name:              "local",
			IngressPriceCents: 10,
			EgressPriceCents:  20,
		}

		datacenter := routing.Datacenter{
			ID:   crypto.HashID("local"),
			Name: "local",
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

	// Configure all GCP related services if the GOOGLE_PROJECT_ID is set
	// GCP VMs actually get populated with the GOOGLE_APPLICATION_CREDENTIALS
	// on creation so we can use that for the default then
	if gcpProjectID, ok := os.LookupEnv("GOOGLE_PROJECT_ID"); ok {
		// Create a Firestore Storer
		fs, err := storage.NewFirestore(ctx, gcpProjectID, logger)
		if err != nil {
			level.Error(logger).Log("err", err)
			os.Exit(1)
		}

		// Start a goroutine to sync from Firestore
		go func() {
			ticker := time.NewTicker(10 * time.Second)
			fs.SyncLoop(ctx, ticker.C)
		}()

		// Set the Firestore Storer to give to handlers
		db = fs
	}

	var routeMatrix routing.RouteMatrix
	{
		if uri, ok := os.LookupEnv("ROUTE_MATRIX_URI"); ok {
			go func() {
				for {
					var matrixReader io.Reader

					// Default to reading route matrix from file
					if f, err := os.Open(uri); err == nil {
						matrixReader = f
					}

					// Prefer to get it remotely if possible
					if r, err := http.Get(uri); err == nil {
						matrixReader = r.Body
					}

					// Attempt to read, and intentionally force to empty route matrix if any errors are encountered to avoid stale routes
					_, err := routeMatrix.ReadFrom(matrixReader)
					if err != nil {
						routeMatrix = routing.RouteMatrix{}
						level.Warn(logger).Log("matrix", "route", "op", "read", "envvar", "ROUTE_MATRIX_URI", "value", uri, "err", err, "msg", "forcing empty route matrix to avoid stale routes")
					}

					level.Info(logger).Log("matrix", "route", "entries", len(routeMatrix.Entries))

					time.Sleep(10 * time.Second)
				}
			}()
		}
	}

	// Remove sessions from top sessions when session meta keys expire
	redisClientPortal.ConfigSet("notify-keyspace-events", "Ex")
	go func() {
		ps := redisClientPortal.Subscribe("__keyevent@0__:expired")
		for {
			msg, err := ps.ReceiveMessage()
			if err != nil {
				level.Error(logger).Log("msg", "error receiving expired message from pubsub", "err", err)
				os.Exit(1)
			}

			if strings.HasSuffix(msg.Payload, "-meta") {
				keyparts := strings.Split(msg.Payload, "-")
				sessionID := keyparts[1]

				topscaniter := redisClientPortal.Scan(0, "top-buyer-*", 1000).Iterator()
				mapscaniter := redisClientPortal.Scan(0, "map-points-buyer-*", 1000).Iterator()
				userscaniter := redisClientPortal.Scan(0, "user-*", 1000).Iterator()

				tx := redisClientPortal.TxPipeline()

				tx.ZRem("top-global", sessionID)
				for topscaniter.Next() {
					tx.ZRem(topscaniter.Val(), sessionID)
				}

				tx.SRem("map-points-global", sessionID)
				for mapscaniter.Next() {
					tx.SRem(mapscaniter.Val(), sessionID)
				}

				for userscaniter.Next() {
					tx.SRem(userscaniter.Val(), sessionID)
				}

				if _, err := tx.Exec(); err != nil {
					level.Error(logger).Log("msg", "error cleaning up top sessions and map sessions", "err", err)
					os.Exit(1)
				}

				level.Info(logger).Log("msg", "removed session from top sessions and map sessions", "session_id", sessionID)
			}
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
			var userRoles = management.RoleList{}
			user := i.Request.Context().Value("user")
			if user != nil {
				claims := user.(*jwt.Token).Claims.(jwt.MapClaims)

				if requestID, ok := claims["sub"]; ok {
					roles, err := auth0Client.Manager.User.Roles(requestID.(string))

					if err == nil {
						userRoles = *roles
					}
				}
				return jsonrpc.SetRoles(i.Request, userRoles)
			}
			return jsonrpc.SetIsAnonymous(i.Request, i.Request.Header.Get("Authorization") == "")
		})
		s.RegisterCodec(json2.NewCodec(), "application/json")
		s.RegisterService(&jsonrpc.OpsService{
			Logger:      logger,
			Release:     release,
			BuildTime:   buildtime,
			RedisClient: redisClientRelays,
			Storage:     db,
			RouteMatrix: &routeMatrix,
		}, "")
		s.RegisterService(&jsonrpc.BuyersService{
			Logger:      logger,
			RedisClient: redisClientPortal,
			Storage:     db,
		}, "")
		s.RegisterService(&jsonrpc.AuthService{
			Logger:  logger,
			Auth0:   auth0Client,
			Storage: db,
		}, "")

		http.Handle("/rpc", jsonrpc.AuthMiddleware(os.Getenv("JWT_AUDIENCE"), s))

		http.Handle("/", http.FileServer(http.Dir(uiDir)))

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
