package main

import (
	"context"
	"encoding/json"
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

	"github.com/networknext/backend/modules/backend"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/transport/middleware"

	frontend "github.com/networknext/backend/modules/relay_frontend"
	"github.com/networknext/backend/modules/storage"
	"github.com/networknext/backend/modules/transport"

	"github.com/gorilla/mux"
)

var (
	buildtime     string
	commitMessage string
	sha           string
	tag           string
	keys          middleware.JWKS
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
		core.Error("failed to get logger: %v", err)
		return 1
	}

	cfg, err := GetRelayFrontendConfig()
	if err != nil {
		core.Error("failed to get relay frontend config: %v", err)
		return 1
	}

	metricsHandler, err := backend.GetMetricsHandler(ctx, logger, gcpProjectID)
	if err != nil {
		core.Error("failed to get metrics handler: %v", err)
		return 1
	}

	if gcpProjectID != "" {
		if err := backend.InitStackDriverProfiler(gcpProjectID, serviceName, cfg.Env); err != nil {
			core.Error("failed to initialize StackDriver profiler: %v", err)
			return 1
		}
	}

	frontendMetrics, err := metrics.NewRelayFrontendMetrics(ctx, metricsHandler)
	if err != nil {
		core.Error("failed to get relay frontend metrics: %v", err)
		return 1
	}

	// Get the redis matrix store
	store, err := storage.NewRedisMatrixStore(cfg.MatrixStoreAddress, cfg.MatrixStorePassword, cfg.MSMaxIdleConnections, cfg.MSMaxActiveConnections, cfg.MSReadTimeout, cfg.MSWriteTimeout, cfg.MSMatrixExpireTimeout)
	if err != nil {
		core.Error("failed to get redis matrix store: %v", err)
		return 1
	}

	// Get the relay frontend
	frontendClient, err := frontend.NewRelayFrontend(store, cfg)
	if err != nil {
		core.Error("failed to get relay frontend client: %v", err)
		return 1
	}

	// Create an error chan for exiting from goroutines
	errChan := make(chan error, 1)

	auth0Domain := os.Getenv("AUTH0_DOMAIN")
	newKeys, err := middleware.FetchAuth0Cert(auth0Domain)
	if err != nil {
		core.Error("failed to fetch auth0 cert: %v", err)
		return 1
	}
	keys = newKeys

	fetchAuthCertInterval, err := envvar.GetDuration("AUTH0_CERT_INTERVAL", time.Minute*10)
	if err != nil {
		core.Error("invalid AUTH0_CERT_INTERVAL: %v", err)
		return 1
	}

	go func() {
		ticker := time.NewTicker(fetchAuthCertInterval)
		for {
			select {
			case <-ticker.C:
				newKeys, err := middleware.FetchAuth0Cert(auth0Domain)
				if err != nil {
					continue
				}
				keys = newKeys
			case <-ctx.Done():
				return
			}
		}
	}()

	matrixSyncInterval, err := envvar.GetDuration("MATRIX_SYNC_INTERVAL", time.Second*1)
	if err != nil {
		core.Error("invalid MATRIX_SYNC_INTERVAL: %v", err)
		return 1
	}

	// Start a goroutine for updating the master relay backend
	go func() {
		ticker := time.NewTicker(matrixSyncInterval)
		for {
			select {
			case <-ctx.Done():
				// Shutdown signal received
				return
			case <-ticker.C:
				if frontendClient.ReachedRetryLimit() {
					// Couldn't get the master relay backend for UPDATE_RETRY_COUNT attempts
					// Reset the cost and route matrix cache
					frontendClient.ResetCachedMatrix(frontend.MatrixTypeCost)
					frontendClient.ResetCachedMatrix(frontend.MatrixTypeNormal)
					core.Error("reached retry limit, reset cost and route matrix cache")
				}

				// Get the oldest relay backend
				frontendMetrics.MasterSelect.Add(1)

				err := frontendClient.UpdateRelayBackendMaster()
				if err != nil {
					frontendClient.RetryCount++
					frontendMetrics.ErrorMetrics.MasterSelectError.Add(1)
					core.Error("failed ot update master relay backend (retry counter %d): %v", frontendClient.RetryCount, err)
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
						core.Error("error getting cost matrix: %v", err)
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
						core.Error("error getting route matrix: %v", err)
						return
					}

					frontendMetrics.RouteMatrix.Success.Add(1)
				}()

				wg.Wait()
			}
		}
	}()

	// Setup the status handler info

	statusData := &metrics.RelayFrontendStatus{}
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

				newStatusData := &metrics.RelayFrontendStatus{}

				// Service Information
				newStatusData.ServiceName = serviceName
				newStatusData.GitHash = sha
				newStatusData.Started = startTime.Format("Mon, 02 Jan 2006 15:04:05 EST")
				newStatusData.Uptime = time.Since(startTime).String()

				// Invocations
				newStatusData.MasterSelectInvocations = int(frontendMetrics.MasterSelect.Value())
				newStatusData.CostMatrixInvocations = int(frontendMetrics.CostMatrix.Invocations.Value())
				newStatusData.RouteMatrixInvocations = int(frontendMetrics.RouteMatrix.Invocations.Value())

				// Success
				newStatusData.MasterSelectSuccessCount = int(frontendMetrics.MasterSelectSuccess.Value())
				newStatusData.CostMatrixSuccessCount = int(frontendMetrics.CostMatrix.Success.Value())
				newStatusData.RouteMatrixSuccessCount = int(frontendMetrics.RouteMatrix.Success.Value())

				// Error
				newStatusData.MasterSelectError = int(frontendMetrics.ErrorMetrics.MasterSelectError.Value())
				newStatusData.CostMatrixError = int(frontendMetrics.CostMatrix.Error.Value())
				newStatusData.RouteMatrixError = int(frontendMetrics.RouteMatrix.Error.Value())

				statusMutex.Lock()
				statusData = newStatusData
				statusMutex.Unlock()

				time.Sleep(time.Second * 10)
			}
		}()
	}

	serveStatusFunc := func(w http.ResponseWriter, r *http.Request) {
		statusMutex.RLock()
		data := statusData
		statusMutex.RUnlock()

		w.Header().Set("Content-Type", "application/json")
		if err := json.NewEncoder(w).Encode(data); err != nil {
			core.Error("could not write status data to json: %v\n%+v", err, data)
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	// Start HTTP Server
	{
		allowedOrigins := envvar.Get("ALLOWED_ORIGINS", "")
		if allowedOrigins == "" {
			core.Debug("unable to parse ALLOWED_ORIGINS environment variable")
		}

		auth0Issuer := envvar.Get("AUTH0_ISSUER", "")
		if auth0Issuer == "" {
			core.Debug("unable to parse AUTH0_ISSUER environment variable")
		}

		port := envvar.Get("PORT", "30005")

		router := mux.NewRouter()
		router.HandleFunc("/health", transport.HealthHandlerFunc())
		router.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage, []string{}))
		router.HandleFunc("/status", serveStatusFunc).Methods("GET")
		router.HandleFunc("/route_matrix", frontendClient.GetRouteMatrixHandlerFunc()).Methods("GET")
		router.HandleFunc("/database_version", frontendClient.GetRelayBackendHandlerFunc("/database_version")).Methods("GET")
		router.HandleFunc("/dest_relays", frontendClient.GetRelayBackendHandlerFunc("/dest_relays")).Methods("GET")
		router.HandleFunc("/master_status", frontendClient.GetRelayBackendHandlerFunc("/status")).Methods("GET")
		router.HandleFunc("/master", frontendClient.GetRelayBackendMasterHandlerFunc()).Methods("GET")
		router.Handle("/debug/vars", expvar.Handler())

		// Wrap the following endpoints in auth and CORS middleware
		// NOTE: the next tool is unaware of CORS and its requests simply pass through
		costMatrixHandler := http.HandlerFunc(frontendClient.GetCostMatrixHandlerFunc())
		router.Handle("/cost_matrix", middleware.PlainHttpAuthMiddleware(keys, envvar.GetList("JWT_AUDIENCES", []string{}), costMatrixHandler, strings.Split(allowedOrigins, ","), auth0Issuer))

		relaysCsvHandler := http.HandlerFunc(frontendClient.GetRelayBackendHandlerFunc("/relays"))
		router.Handle("/relays", middleware.PlainHttpAuthMiddleware(keys, envvar.GetList("JWT_AUDIENCES", []string{}), relaysCsvHandler, strings.Split(allowedOrigins, ","), auth0Issuer))

		jsonDashboardHandler := http.HandlerFunc(frontendClient.GetRelayDashboardDataHandlerFunc())
		router.Handle("/relay_dashboard_data", middleware.PlainHttpAuthMiddleware(keys, envvar.GetList("JWT_AUDIENCES", []string{}), jsonDashboardHandler, strings.Split(allowedOrigins, ","), auth0Issuer))

		jsonDashboardAnalysisHandler := http.HandlerFunc(frontendClient.GetRelayDashboardAnalysisHandlerFunc())
		router.Handle("/relay_dashboard_analysis", middleware.PlainHttpAuthMiddleware(keys, envvar.GetList("JWT_AUDIENCES", []string{}), jsonDashboardAnalysisHandler, strings.Split(allowedOrigins, ","), auth0Issuer))

		enablePProf, err := envvar.GetBool("FEATURE_ENABLE_PPROF", false)
		if err != nil {
			core.Error("could not parse FEATURE_ENABLE_PPROF: %v", err)
		}
		if enablePProf {
			router.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
		}

		go func() {
			fmt.Printf("starting http server on port %s\n", port)
			err := http.ListenAndServe(":"+port, router)
			if err != nil {
				core.Error("could not start http server: %v", err)
				errChan <- err
			}
		}()
	}

	// Wait for shutdown signal
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-termChan: // Exit with an error code of 0 if we receive SIGINT or SIGTERM
		fmt.Println("Received shutdown signal.")

		cancel()

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
