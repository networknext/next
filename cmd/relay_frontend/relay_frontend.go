package main

import (
	"bytes"
	"context"
	"expvar"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"sync"
	"syscall"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/networknext/backend/modules/analytics"
	"github.com/networknext/backend/modules/backend"
	"github.com/networknext/backend/modules/common/helpers"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/transport/middleware"

	frontend "github.com/networknext/backend/modules/relay_frontend"
	"github.com/networknext/backend/modules/storage"
	"github.com/networknext/backend/modules/transport"

	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"
)

var (
	buildtime     string
	commitMessage string
	sha           string
	tag           string
)

func main() {
	os.Exit(mainReturnWithCode())
}

func mainReturnWithCode() int {

	serviceName := "relay_frontend"
	fmt.Printf("%s: Git Hash: %s - Commit: %s\n", serviceName, sha, commitMessage)

	est, _ := time.LoadLocation("EST")
	startTime := time.Now().In(est)

	// Setup the service
	ctx, cancel := context.WithCancel(context.Background())
	gcpProjectID := backend.GetGCPProjectID()
	logger, err := backend.GetLogger(ctx, gcpProjectID, serviceName)
	if err != nil {
		fmt.Println(err.Error())
		return 1
	}

	cfg, err := GetRelayFrontendConfig()
	if err != nil {
		_ = level.Error(logger).Log("err", err)
		return 1
	}

	metricsHandler, err := backend.GetMetricsHandler(ctx, logger, gcpProjectID)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	if gcpProjectID != "" {
		if err := backend.InitStackDriverProfiler(gcpProjectID, serviceName, cfg.Env); err != nil {
			level.Error(logger).Log("msg", "failed to initialze StackDriver profiler", "err", err)
			return 1
		}
	}

	frontendMetrics, err := metrics.NewRelayFrontendMetrics(ctx, metricsHandler)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	// Get the redis matrix store
	store, err := storage.NewRedisMatrixStore(cfg.MatrixStoreAddress, cfg.MatrixStorePassword, cfg.MSMaxIdleConnections, cfg.MSMaxActiveConnections, cfg.MSReadTimeout, cfg.MSWriteTimeout, cfg.MSMatrixExpireTimeout)
	if err != nil {
		_ = level.Error(logger).Log("err", err)
		return 1
	}

	// Get the relay frontend
	frontendClient, err := frontend.NewRelayFrontend(store, cfg)
	if err != nil {
		_ = level.Error(logger).Log("err", err)
		return 1
	}

	// Start a goroutine for updating the master relay backend
	go func() {
		syncTimer := helpers.NewSyncTimer(1000 * time.Millisecond)
		for {
			syncTimer.Run()
			select {
			case <-ctx.Done():
				// Shutdown signal received
				return
			default:
				if frontendClient.ReachedRetryLimit() {
					// Couldn't get the master relay backend for UPDATE_RETRY_COUNT attempts
					// Reset the cost and route matrix cache
					frontendClient.ResetCachedMatrix(frontend.MatrixTypeCost)
					frontendClient.ResetCachedMatrix(frontend.MatrixTypeNormal)
					level.Debug(logger).Log("msg", "reached retry limit, reset cost and route matrix cache")
				}

				// Get the oldest relay backend
				frontendMetrics.MasterSelect.Add(1)

				err := frontendClient.UpdateRelayBackendMaster()
				if err != nil {
					frontendClient.RetryCount++
					frontendMetrics.ErrorMetrics.MasterSelectError.Add(1)
					_ = level.Error(logger).Log("error", err)
					continue
				}
				frontendClient.RetryCount = 0
				frontendMetrics.MasterSelectSuccess.Add(1)

				// Create waitgroup for worker goroutines
				wg := sync.WaitGroup{}

				// Cache the cost matrix
				wg.Add(1)
				go func() {
					defer wg.Done()

					frontendMetrics.CostMatrix.Invocations.Add(1)

					err = frontendClient.CacheMatrix(frontend.MatrixTypeCost)
					if err != nil {
						frontendMetrics.CostMatrix.Error.Add(1)
						_ = level.Error(logger).Log("msg", "error getting cost matrix", "error", err)
						return
					}

					frontendMetrics.CostMatrix.Success.Add(1)
				}()

				// Cache the route matrix
				wg.Add(1)
				go func() {
					defer wg.Done()

					frontendMetrics.RouteMatrix.Invocations.Add(1)

					err = frontendClient.CacheMatrix(frontend.MatrixTypeNormal)
					if err != nil {
						frontendMetrics.RouteMatrix.Error.Add(1)
						_ = level.Error(logger).Log("msg", "error getting normal matrix", "error", err)
						return
					}

					frontendMetrics.RouteMatrix.Success.Add(1)
				}()

				wg.Wait()
			}
		}
	}()

	errChan := make(chan error, 1)

	// relay stats

	var relayStatsPublisher analytics.RelayStatsPublisher = &analytics.NoOpRelayStatsPublisher{}
	var pingStatsPublisher analytics.PingStatsPublisher = &analytics.NoOpPingStatsPublisher{}
	{
		emulatorOK := envvar.Exists("PUBSUB_EMULATOR_HOST")
		if gcpProjectID != "" || emulatorOK {

			pubsubCtx := ctx
			if emulatorOK {
				gcpProjectID = "local"

				var cancelFunc context.CancelFunc
				pubsubCtx, cancelFunc = context.WithDeadline(ctx, time.Now().Add(60*time.Minute))
				defer cancelFunc()

				level.Info(logger).Log("msg", "Detected pubsub emulator")
			}

			// Google Pubsub
			{
				settings := pubsub.PublishSettings{
					DelayThreshold: time.Second,
					CountThreshold: 1,
					ByteThreshold:  1 << 14,
					NumGoroutines:  runtime.GOMAXPROCS(0),
					Timeout:        time.Minute,
				}

				pingPubsub, err := analytics.NewGooglePubSubPingStatsPublisher(pubsubCtx, &frontendMetrics.PingStatsMetrics, logger, gcpProjectID, "ping_stats", settings)
				if err != nil {
					level.Error(logger).Log("msg", "could not create ping stats analytics pubsub publisher", "err", err)
					return 1
				}

				pingStatsPublisher = pingPubsub

				relayPubsub, err := analytics.NewGooglePubSubRelayStatsPublisher(pubsubCtx, &frontendMetrics.RelayStatsMetrics, logger, gcpProjectID, "relay_stats", settings)
				if err != nil {
					level.Error(logger).Log("msg", "could not create relay stats analytics pubsub publisher", "err", err)
					return 1
				}

				relayStatsPublisher = relayPubsub
			}
		}

		go func() {
			publishInterval, err := envvar.GetDuration("PING_STATS_PUBLISH_INTERVAL", time.Minute)
			if err != nil {
				level.Error(logger).Log("err", err)
				errChan <- err
			}

			syncTimer := helpers.NewSyncTimer(publishInterval)
			for {
				syncTimer.Run()
				routeMatrixBuffer := frontendClient.GetRouteMatrix()

				if len(routeMatrixBuffer) > 0 {
					var routeMatrix routing.RouteMatrix
					readStream := encoding.CreateReadStream(routeMatrixBuffer)
					if err := routeMatrix.Serialize(readStream); err != nil {
						level.Error(logger).Log("err", err)
						continue
					}

					numPingStats := len(routeMatrix.PingStats)

					core.Debug("Number of ping stats to be published: %d", numPingStats)
					if numPingStats > 0 {
						if err := pingStatsPublisher.Publish(ctx, routeMatrix.PingStats); err != nil {
							level.Error(logger).Log("err", err)
							errChan <- err
						}
					}
				}
			}
		}()

		go func() {
			publishInterval, err := envvar.GetDuration("RELAY_STATS_PUBLISH_INTERVAL", time.Second*10)
			if err != nil {
				level.Error(logger).Log("err", err)
				errChan <- err
			}

			syncTimer := helpers.NewSyncTimer(publishInterval)
			for {
				syncTimer.Run()

				routeMatrixBuffer := frontendClient.GetRouteMatrix()

				if len(routeMatrixBuffer) > 0 {
					var routeMatrix routing.RouteMatrix
					readStream := encoding.CreateReadStream(routeMatrixBuffer)
					if err := routeMatrix.Serialize(readStream); err != nil {
						level.Error(logger).Log("err", err)
						continue
					}

					numRelayStats := len(routeMatrix.RelayStats)

					core.Debug("Number of relay stats to be published: %d", len(routeMatrix.RelayStats))
					if numRelayStats > 0 {
						if err := relayStatsPublisher.Publish(ctx, routeMatrix.RelayStats); err != nil {
							level.Error(logger).Log("err", err)
						}
					}
				}
			}
		}()
	}

	// Setup the status handler info
	var statusData []byte
	var statusMutex sync.RWMutex

	{
		memoryUsed := func() float64 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			return float64(m.Alloc) / (1000.0 * 1000.0)
		}

		go func() {
			for {
				frontendMetrics.FrontendServiceMetrics.Goroutines.Set(float64(runtime.NumGoroutine()))
				frontendMetrics.FrontendServiceMetrics.MemoryAllocated.Set(memoryUsed())

				statusDataString := fmt.Sprintf("%s\n", serviceName)
				statusDataString += fmt.Sprintf("git hash %s\n", sha)
				statusDataString += fmt.Sprintf("started %s\n", startTime.Format("Mon, 02 Jan 2006 15:04:05 EST"))
				statusDataString += fmt.Sprintf("uptime %s\n", time.Since(startTime))
				statusDataString += fmt.Sprintf("%d goroutines\n", int(frontendMetrics.FrontendServiceMetrics.Goroutines.Value()))
				statusDataString += fmt.Sprintf("%.2f mb allocated\n", frontendMetrics.FrontendServiceMetrics.MemoryAllocated.Value())
				statusDataString += fmt.Sprintf("%d master select invocations\n", int(frontendMetrics.MasterSelect.Value()))
				statusDataString += fmt.Sprintf("%d cost matrix invocations\n", int(frontendMetrics.CostMatrix.Invocations.Value()))
				statusDataString += fmt.Sprintf("%d route matrix invocations\n", int(frontendMetrics.RouteMatrix.Invocations.Value()))
				statusDataString += fmt.Sprintf("%d master select success count\n", int(frontendMetrics.MasterSelectSuccess.Value()))
				statusDataString += fmt.Sprintf("%d cost matrix success count\n", int(frontendMetrics.CostMatrix.Success.Value()))
				statusDataString += fmt.Sprintf("%d route matrix success count\n", int(frontendMetrics.RouteMatrix.Success.Value()))
				statusDataString += fmt.Sprintf("%d master select errors\n", int(frontendMetrics.ErrorMetrics.MasterSelectError.Value()))
				statusDataString += fmt.Sprintf("%d cost matrix errors\n", int(frontendMetrics.CostMatrix.Error.Value()))
				statusDataString += fmt.Sprintf("%d route matrix errors\n", int(frontendMetrics.RouteMatrix.Error.Value()))
				statusDataString += fmt.Sprintf("%d ping stats entries submitted\n", int(frontendMetrics.PingStatsMetrics.EntriesSubmitted.Value()))
				statusDataString += fmt.Sprintf("%d ping stats entries queued\n", int(frontendMetrics.PingStatsMetrics.EntriesQueued.Value()))
				statusDataString += fmt.Sprintf("%d ping stats entries flushed\n", int(frontendMetrics.PingStatsMetrics.EntriesFlushed.Value()))
				statusDataString += fmt.Sprintf("%d relay stats entries submitted\n", int(frontendMetrics.RelayStatsMetrics.EntriesSubmitted.Value()))
				statusDataString += fmt.Sprintf("%d relay stats entries queued\n", int(frontendMetrics.RelayStatsMetrics.EntriesQueued.Value()))
				statusDataString += fmt.Sprintf("%d relay stats entries flushed\n", int(frontendMetrics.RelayStatsMetrics.EntriesFlushed.Value()))

				statusMutex.Lock()
				statusData = []byte(statusDataString)
				statusMutex.Unlock()

				time.Sleep(time.Second * 10)
			}
		}()
	}

	serveStatusFunc := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		statusMutex.RLock()
		data := statusData
		statusMutex.RUnlock()
		buffer := bytes.NewBuffer(data)
		_, err := buffer.WriteTo(w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	allowedOrigins, found := os.LookupEnv("ALLOWED_ORIGINS")
	if !found {
		level.Error(logger).Log("msg", "unable to parse ALLOWED_ORIGINS environment variable")
	}

	audience, found := os.LookupEnv("JWT_AUDIENCE")
	if !found {
		level.Error(logger).Log("msg", "unable to parse JWT_AUDIENCE environment variable")
	}

	port := envvar.Get("PORT", "30005")
	fmt.Printf("starting http server on port %s\n", port)

	router := mux.NewRouter()
	router.HandleFunc("/health", transport.HealthHandlerFunc())
	router.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage, []string{}))
	router.HandleFunc("/status", serveStatusFunc).Methods("GET")
	router.HandleFunc("/route_matrix", frontendClient.GetRouteMatrixHandlerFunc()).Methods("GET")
	router.HandleFunc("/database_version", frontendClient.GetRelayBackendHandlerFunc("/database_version")).Methods("GET")
	router.HandleFunc("/relay_dashboard", frontendClient.GetRelayDashboardHandlerFunc("local", "local")).Methods("GET")
	router.HandleFunc("/dest_relays", frontendClient.GetRelayBackendHandlerFunc("/dest_relays")).Methods("GET")
	router.HandleFunc("/master_status", frontendClient.GetRelayBackendHandlerFunc("/status")).Methods("GET")
	router.HandleFunc("/master", frontendClient.GetRelayBackendMasterHandlerFunc()).Methods("GET")
	router.Handle("/debug/vars", expvar.Handler())

	// Wrap the following endpoints in auth and CORS middleware
	// NOTE: the next tool is unaware of CORS and its requests simply pass through

	// this call will not work via auth, fails within the auth0 stack
	// router.HandleFunc("/cost_matrix", frontendClient.GetCostMatrixHandlerFunc()).Methods("GET")
	costMatrixHandler := http.HandlerFunc(frontendClient.GetCostMatrixHandlerFunc())
	router.Handle("/cost_matrix", middleware.PlainHttpAuthMiddleware(audience, costMatrixHandler, strings.Split(allowedOrigins, ",")))

	relaysCsvHandler := http.HandlerFunc(frontendClient.GetRelayBackendHandlerFunc("/relays"))
	router.Handle("/relays", middleware.PlainHttpAuthMiddleware(audience, relaysCsvHandler, strings.Split(allowedOrigins, ",")))

	jsonDashboardHandler := http.HandlerFunc(frontendClient.GetRelayDashboardDataHandlerFunc())
	router.Handle("/relay_dashboard_data", middleware.PlainHttpAuthMiddleware(audience, jsonDashboardHandler, strings.Split(allowedOrigins, ",")))

	jsonDashboardAnalysisHandler := http.HandlerFunc(frontendClient.GetRelayDashboardAnalysisHandlerFunc())
	router.Handle("/relay_dashboard_analysis", middleware.PlainHttpAuthMiddleware(audience, jsonDashboardAnalysisHandler, strings.Split(allowedOrigins, ",")))

	enablePProf, err := envvar.GetBool("FEATURE_ENABLE_PPROF", false)
	if err != nil {
		level.Error(logger).Log("err", err)
	}
	if enablePProf {
		router.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
	}

	go func() {

		_ = level.Info(logger).Log("addr", ":"+port)

		err := http.ListenAndServe(":"+port, router)
		if err != nil {
			_ = level.Error(logger).Log("err", err)
			errChan <- err
		}
	}()

	// Wait for shutdown signal
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-termChan: // Exit with an error code of 0 if we receive SIGINT or SIGTERM
		level.Debug(logger).Log("msg", "Received shutdown signal")
		fmt.Println("Received shutdown signal.")

		cancel()

		level.Debug(logger).Log("msg", "Successfully shutdown.")
		fmt.Println("Successfully shutdown.")
		return 0
	case <-errChan: // Exit with an error code of 1 if we receive any errors from goroutines
		cancel()
		return 1
	}

	return 0
}

func GetRelayFrontendConfig() (*frontend.RelayFrontendConfig, error) {
	cfg := new(frontend.RelayFrontendConfig)
	var err error

	cfg.Env = envvar.Get("ENV", "local")

	cfg.MasterTimeVariance, err = envvar.GetDuration("MASTER_TIME_VARIANCE", 5*time.Second)
	if err != nil {
		return nil, err
	}

	cfg.UpdateRetryCount, err = envvar.GetInt("UPDATE_RETRY_COUNT", 5)
	if err != nil {
		return nil, err
	}

	cfg.MatrixStoreAddress = envvar.Get("MATRIX_STORE_ADDRESS", "")
	if cfg.MatrixStoreAddress == "" {
		return nil, fmt.Errorf("MATRIX_STORE_ADDRESS not set")
	}

	cfg.MatrixStorePassword = envvar.Get("MATRIX_STORE_PASSWORD", "")

	maxIdleConnections, err := envvar.GetInt("MATRIX_STORE_MAX_IDLE_CONNS", 5)
	if err != nil {
		return nil, err
	}
	cfg.MSMaxIdleConnections = maxIdleConnections

	maxActiveConnections, err := envvar.GetInt("MATRIX_STORE_MAX_ACTIVE_CONNS", 5)
	if err != nil {
		return nil, err
	}
	cfg.MSMaxActiveConnections = maxActiveConnections

	cfg.MSReadTimeout, err = envvar.GetDuration("MATRIX_STORE_READ_TIMEOUT", 250*time.Millisecond)
	if err != nil {
		return nil, err
	}

	cfg.MSWriteTimeout, err = envvar.GetDuration("MATRIX_STORE_WRITE_TIMEOUT", 250*time.Millisecond)
	if err != nil {
		return nil, err
	}

	cfg.MSMatrixExpireTimeout, err = envvar.GetDuration("MATRIX_STORE_EXPIRE_TIMEOUT", 5*time.Second)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
