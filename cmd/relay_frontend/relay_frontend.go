package main

import (
	"context"
	"expvar"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"sync"
	"time"

	"github.com/networknext/backend/modules/metrics"

	"github.com/gorilla/mux"
	"github.com/networknext/backend/modules/common/helpers"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/transport"

	rf "github.com/networknext/backend/modules/relay_frontend"
	"github.com/networknext/backend/modules/storage"

	//logging
	"github.com/go-kit/kit/log/level"
	"github.com/networknext/backend/modules/backend" // todo: not a good name for a module
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

	ctx := context.Background()
	gcpProjectID := backend.GetGCPProjectID()
	logger, err := backend.GetLogger(ctx, gcpProjectID, serviceName)
	if err != nil {
		fmt.Println(err.Error())
		return 1
	}

	cfg, err := rf.GetConfig()
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
		if err := backend.InitStackDriverProfiler(gcpProjectID, serviceName, cfg.ENV); err != nil {
			level.Error(logger).Log("msg", "failed to initialze StackDriver profiler", "err", err)
			return 1
		}
	}

	frontendMetrics, err := metrics.NewRelayFrontendMetrics(ctx, metricsHandler)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	store, err := storage.NewRedisMatrixStore(cfg.MatrixStoreAddress, cfg.MSReadTimeout, cfg.MSWriteTimeout, cfg.MSMatrixTimeout)
	if err != nil {
		_ = level.Error(logger).Log("err", err)
		return 1
	}

	svc, err := rf.NewRelayFrontend(store, cfg)
	if err != nil {
		_ = level.Error(logger).Log("err", err)
		return 1
	}

	shutdown := false

	// core loop
	go func() {
		syncTimer := helpers.NewSyncTimer(1000 * time.Millisecond)
		for {
			syncTimer.Run()
			if shutdown {
				return
			}

			frontendMetrics.MasterSelect.Add(1)
			err := svc.UpdateRelayBackendMaster()
			if err != nil {
				frontendMetrics.MasterSelectError.Add(1)
				_ = level.Error(logger).Log("error", err)
				continue
			}
			wg := sync.WaitGroup{}

			wg.Add(1)
			go func() {
				frontendMetrics.CostMatrix.Invocations.Add(1)
				err = svc.CacheMatrix(rf.MatrixTypeCost)
				if err != nil {
					frontendMetrics.CostMatrix.Error.Add(1)
					_ = level.Error(logger).Log("msg", "error getting cost matrix", "error", err)
				}
				wg.Done()
			}()

			wg.Add(1)
			go func() {
				frontendMetrics.RouteMatrix.Invocations.Add(1)
				err = svc.CacheMatrix(rf.MatrixTypeNormal)
				if err != nil {
					frontendMetrics.RouteMatrix.Error.Add(1)
					_ = level.Error(logger).Log("msg", "error getting normal matrix", "error", err)
				}
				wg.Done()
			}()

			if cfg.ValveMatrix {
				wg.Add(1)
				go func() {
					frontendMetrics.ValveMatrix.Invocations.Add(1)
					err = svc.CacheMatrix(rf.MatrixTypeValve)
					if err != nil {
						frontendMetrics.ValveMatrix.Error.Add(1)
						_ = level.Error(logger).Log("msg", "error getting valve matrix", "error", err)
					}
					wg.Done()
				}()
			}
			wg.Wait()
		}
	}()

	fmt.Printf("starting http server\n")

	router := mux.NewRouter()
	router.HandleFunc("/health", transport.HealthHandlerFunc())
	router.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage, []string{}))
	router.HandleFunc("/cost_matrix", svc.GetCostMatrix()).Methods("GET")
	router.HandleFunc("/route_matrix", svc.GetRouteMatrix()).Methods("GET")
	router.Handle("/debug/vars", expvar.Handler())

	if cfg.ValveMatrix {
		router.HandleFunc("/route_matrix_valve", svc.GetRouteMatrixValve()).Methods("GET")
	}

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
			os.Exit(1) // todo: don't os.Exit() here, but find a way to exit
		}
	}()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)

	select {
	case <-sigint:
		shutdown = true
		time.Sleep(5 * time.Second)
	}
	return 0
}
