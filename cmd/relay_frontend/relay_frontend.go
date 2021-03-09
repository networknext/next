package main

import (
	"context"
	"expvar"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/signal"
	"time"

	"github.com/gorilla/mux"
	"github.com/networknext/backend/modules/common/helpers"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/transport"

	rm "github.com/networknext/backend/modules/relay_frontend"
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

	serviceName := "relay_frontend"
	fmt.Printf("%s: Git Hash: %s - Commit: %s\n", serviceName, sha, commitMessage)

	ctx := context.Background()
	gcpProjectID := backend.GetGCPProjectID()
	logger, err := backend.GetLogger(ctx, gcpProjectID, serviceName)
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}

	cfg, err := rm.GetConfig()
	if err != nil {
		_ = level.Error(logger).Log("err", err)
	}

	store, err := storage.NewRedisMatrixStore(cfg.MatrixStoreAddress, cfg.MSReadTimeout, cfg.MSWriteTimeout, cfg.MSMatrixTimeout)
	if err != nil {
		_ = level.Error(logger).Log("err", err)
		os.Exit(1)
	}
	svc, err := rm.New(store, cfg)
	if err != nil {
		_ = level.Error(logger).Log("err", err)
		os.Exit(1)
	}

	shutdown := false

	//core loop
	go func() {
		syncTimer := helpers.NewSyncTimer(1000 * time.Millisecond)
		for {
			syncTimer.Run()
			if shutdown {
				return
			}

			err := svc.UpdateRelayBackendMaster()
			if err != nil {
				_ = level.Error(logger).Log("error", err)
			}

			err = svc.CacheMatrix(rm.MatrixTypeCost)
			if err != nil {
				_ = level.Error(logger).Log("msg", "error getting cost matrix", "error", err)
			}

			err = svc.CacheMatrix(rm.MatrixTypeNormal)
			if err != nil {
				_ = level.Error(logger).Log("msg", "error getting normal matrix", "error", err)
			}

			err = svc.CacheMatrix(rm.MatrixTypeValve)
			if err != nil {
				_ = level.Error(logger).Log("msg", "error getting valve matrix", "error", err)
			}
		}
	}()

	fmt.Printf("starting http server\n")

	router := mux.NewRouter()
	router.HandleFunc("/health", transport.HealthHandlerFunc())
	router.HandleFunc("/cost_matrix", svc.GetCostMatrix()).Methods("GET")
	router.HandleFunc("/route_matrix", svc.GetRouteMatrix()).Methods("GET")
	router.HandleFunc("/route_matrix_valve", svc.GetRouteMatrixValve()).Methods("GET")
	router.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage, []string{}))
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
}
