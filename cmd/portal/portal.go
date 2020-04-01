package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/gorilla/rpc/v2"
	"github.com/gorilla/rpc/v2/json2"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport/jsonrpc"
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
			logger = level.NewFilter(logger, level.AllowInfo())
		}

		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	}

	var relayPublicKey []byte
	{
		if key := os.Getenv("RELAY_PUBLIC_KEY"); len(key) != 0 {
			relayPublicKey, _ = base64.StdEncoding.DecodeString(key)
		}
	}

	redisRelayHost := os.Getenv("REDIS_HOST_RELAYS")
	redisClientRelays := storage.NewRedisClient(redisRelayHost)
	if err := redisClientRelays.Ping().Err(); err != nil {
		level.Error(logger).Log("envvar", "REDIS_HOST_RELAYS", "value", redisRelayHost, "err", err)
		os.Exit(1)
	}

	addr := net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 40000}
	var db storage.Storer = &storage.InMemory{
		LocalRelays: []routing.Relay{
			routing.Relay{
				ID:        crypto.HashID(addr.String()),
				Addr:      addr,
				PublicKey: relayPublicKey,
				Datacenter: routing.Datacenter{
					ID:   crypto.HashID("local"),
					Name: "local",
				},
				Seller: routing.Seller{
					Name:              "local",
					IngressPriceCents: 10,
					EgressPriceCents:  20,
				},
			},
		},
	}

	// Configure all GCP related services if the GOOGLE_PROJECT_ID is set
	// GCP VMs actually get populated with the GOOGLE_APPLICATION_CREDENTIALS
	// on creation so we can use that for the default then
	if gcpProjectID, ok := os.LookupEnv("GOOGLE_PROJECT_ID"); ok {
		firestoreClient, err := firestore.NewClient(ctx, gcpProjectID)
		if err != nil {
			level.Error(logger).Log("err", err)
			os.Exit(1)
		}

		// Create a Firestore Storer
		fs := storage.Firestore{
			Client: firestoreClient,
			Logger: logger,
		}

		// Start a goroutine to sync from Firestore
		go func() {
			ticker := time.NewTicker(10 * time.Second)
			fs.SyncLoop(ctx, ticker.C)
		}()

		// Set the Firestore Storer to give to handlers
		db = &fs
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

	go func() {
		port, ok := os.LookupEnv("PORT")
		if !ok {
			level.Error(logger).Log("err", "env var PORT must be set")
			os.Exit(1)
		}

		level.Info(logger).Log("msg", fmt.Sprintf("Starting portal on port %s", port))

		s := rpc.NewServer()
		s.RegisterCodec(json2.NewCodec(), "application/json")
		s.RegisterService(&jsonrpc.OpsService{
			RedisClient: redisClientRelays,
			Storage:     db,
		}, "")
		http.Handle("/rpc", s)

		http.Handle("/", http.FileServer(http.Dir("./cmd/portal/public")))

		// http.HandleFunc("/", transport.PortalHandlerFunc(redisClientRelays, &routeMatrix, os.Getenv("BASIC_AUTH_USERNAME"), os.Getenv("BASIC_AUTH_PASSWORD")))
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
