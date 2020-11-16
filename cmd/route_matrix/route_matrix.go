

package main

import (
	"context"
	"fmt"
	"github.com/networknext/backend/modules/common/helpers"
	"os"
	"os/signal"
	"time"

	rm "github.com/networknext/backend/route_matrix"
	"github.com/networknext/backend/storage"

	//logging
	"github.com/go-kit/kit/log/level"
	"github.com/networknext/backend/backend"
)

var (
	buildtime     string
	commitMessage string
	sha           string
	tag           string
)

func main() {

	serviceName := "Optimizer"
	fmt.Printf("%s: Git Hash: %s - Commit: %s\n", serviceName, sha, commitMessage)

	ctx := context.Background()
	gcpProjectID := backend.GetGCPProjectID()
	logger, err := backend.GetLogger(ctx, gcpProjectID, serviceName)
	if err != nil {
		level.Error(logger).Log("err", err)
		os.Exit(2)
	}

	cfg,err := rm.GetConfig()
	if err != nil {
		level.Error(logger).Log("err", err)
	}

	store, err := storage.NewRedisMatrixStore(cfg.MatrixStoreAddress, cfg.MSReadTimeout, cfg.MSWriteTimeout, cfg.MSMatrixTimeout)
	if err != nil {
		_ = level.Error(logger).Log("err", err)
		os.Exit(1)
	}
	svc, err := rm.New(store, cfg.MatrixSvcTimeVariance, cfg.OptimizerTimeVariance)
	if err != nil {
		_ = level.Error(logger).Log("err", err)
		os.Exit(1)
	}

	shutdown := false
	//update matrix service
	go func() {
		errorCount := 0
		syncTimer :=helpers.NewSyncTimer(250 *time.Millisecond)
		for {
			syncTimer.Run()
			if shutdown {
				return
			}
			err := svc.UpdateSvcDB()
			if err != nil {
				_ = level.Error(logger).Log("err", err)
				errorCount++
				if errorCount >= 3{
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
		syncTimer := helpers.NewSyncTimer(1000 *time.Millisecond)
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

			if !svc.AmMaster(){
				continue
			}

			err = svc.UpdateLiveRouteMatrix()
			if err != nil {
				_ = level.Error(logger).Log("error", err)
			}

			err = svc.CleanUpDB()
			if err != nil {
				_ = level.Error(logger).Log("error", err)
			}

		}
	}()

	// Wait for interrupt signal
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)

	select{
		case <-sigint:
			shutdown = true
			time.Sleep(5 * time.Second)
	}
}
