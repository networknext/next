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
	"sync"
	"time"

	"github.com/networknext/backend/modules/backend"
	"github.com/networknext/backend/modules/common/helpers"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/metrics"

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
	ctx, cancelFunc := context.WithCancel(context.Background())
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
	store, err := storage.NewRedisMatrixStore(cfg.MatrixStoreAddress, cfg.MSMaxIdleConnections, cfg.MSMaxActiveConnections, cfg.MSReadTimeout, cfg.MSWriteTimeout, cfg.MSMatrixExpireTimeout)
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
				// Get the oldest relay backend
				frontendMetrics.MasterSelect.Add(1)

				err := frontendClient.UpdateRelayBackendMaster()
				if err != nil {
					frontendMetrics.ErrorMetrics.MasterSelectError.Add(1)
					_ = level.Error(logger).Log("error", err)
					continue
				}

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

	fmt.Printf("starting http server\n")

	router := mux.NewRouter()
	router.HandleFunc("/health", transport.HealthHandlerFunc())
	router.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage, []string{}))
	router.HandleFunc("/status", serveStatusFunc).Methods("GET")
	router.HandleFunc("/cost_matrix", frontendClient.GetCostMatrix()).Methods("GET")
	router.HandleFunc("/route_matrix", frontendClient.GetRouteMatrix()).Methods("GET")
	router.HandleFunc("/relay_stats", frontendClient.GetRelayStats())
	router.Handle("/debug/vars", expvar.Handler())

	enablePProf, err := envvar.GetBool("FEATURE_ENABLE_PPROF", false)
	if err != nil {
		level.Error(logger).Log("err", err)
	}
	if enablePProf {
		router.PathPrefix("/debug/pprof/").Handler(http.DefaultServeMux)
	}

	go func() {
		port := envvar.Get("PORT", "30005")

		_ = level.Info(logger).Log("addr", ":"+port)

		err := http.ListenAndServe(":"+port, router)
		if err != nil {
			_ = level.Error(logger).Log("err", err)
			os.Exit(1) // TODO: don't os.Exit() here, but find a way to exit
		}
	}()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)

	select {
	case <-sigint:
		cancelFunc()
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

	cfg.MatrixStoreAddress = envvar.Get("MATRIX_STORE_ADDRESS", "")
	if cfg.MatrixStoreAddress == "" {
		return nil, fmt.Errorf("MATRIX_STORE_ADDRESS not set")
	}

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
