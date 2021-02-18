package main

import (
	"context"
	"expvar"
	"fmt"
	"net/http"
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
	//update matrix service
	go func() {
		errorCount := 0
		syncTimer := helpers.NewSyncTimer(250 * time.Millisecond)
		for {
			syncTimer.Run()
			if shutdown {
				return
			}
			err := svc.UpdateSvcDB()
			if err != nil {
				_ = level.Error(logger).Log("err", err)
				errorCount++
				if errorCount >= 3 {
					_ = level.Error(logger).Log("msg", "updating svc failed multiple times in a row")
					os.Exit(1)
				}
				continue
			}
			errorCount = 0
		}
	}()

	//core loop
	go func() {
		syncTimer := helpers.NewSyncTimer(1000 * time.Millisecond)
		for {
			syncTimer.Run()
			if shutdown {
				return
			}

			err := svc.DetermineMaster()
			if err != nil {
				_ = level.Error(logger).Log("error", err)
				continue
			}

			if svc.AmMaster() {
				if cfg.RB15Enabled {
					err := svc.UpdateRelayBackendMaster()
					if err != nil {
						_ = level.Error(logger).Log("error", err)
					}
					currentMBAddr := svc.GetRelayBackendMasterAddress()
					err = svc.UpdateLiveRouteMatrixBackend(fmt.Sprintf("http://%s/route_matrix", currentMBAddr), storage.MatrixTypeNormal)
					if err != nil {
						_ = level.Error(logger).Log("error", err)
					}
					err = svc.UpdateLiveRouteMatrixBackend(fmt.Sprintf("http://%s/route_matrix_valve", currentMBAddr), storage.MatrixTypeValve)
					if err != nil {
						_ = level.Error(logger).Log("error", err)
					}
				}

				if cfg.RB20Enabled {
					err = svc.UpdateLiveRouteMatrixOptimizer()
					if err != nil {
						_ = level.Error(logger).Log("error", err)
					}
				}
				err = svc.CleanUpDB()
				if err != nil {
					_ = level.Error(logger).Log("error", err)
				}
			}

			err = svc.CacheMatrix(storage.MatrixTypeNormal)
			if err != nil {
				_ = level.Error(logger).Log("msg", "error getting normal matrix", "error", err)
			}

			err = svc.CacheMatrix(storage.MatrixTypeValve)
			if err != nil {
				_ = level.Error(logger).Log("msg", "error getting valve matrix", "error", err)
			}
		}
	}()

	fmt.Printf("starting http server\n")

	router := mux.NewRouter()
	router.HandleFunc("/health", transport.HealthHandlerFunc())
	router.HandleFunc("/route_matrix", svc.GetMatrix()).Methods("GET")
	router.HandleFunc("/route_matrix_valve", svc.GetMatrixValve()).Methods("GET")
	router.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage, false, []string{}))
	router.Handle("/debug/vars", expvar.Handler())

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
