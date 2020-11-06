

package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"time"

	rm "github.com/networknext/backend/route_matrix"
	"github.com/networknext/backend/storage"

	//logging
	gcplogging "cloud.google.com/go/logging"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/networknext/backend/logging"
)


func main() {

	// Configure logging
	logger := log.NewLogfmtLogger(os.Stdout)
	ctx := context.Background()

	var enableSDLogging bool
	enableSDLoggingString, ok := os.LookupEnv("ENABLE_STACKDRIVER_LOGGING")
	if ok {
		var err error
		enableSDLogging, err = strconv.ParseBool(enableSDLoggingString)
		if err != nil {
			_ = level.Error(logger).Log("envvar", "ENABLE_STACKDRIVER_LOGGING", "msg", "could not parse", "err", err)
			os.Exit(1)
		}
	}

	if enableSDLogging {
		if projectID, ok := os.LookupEnv("GOOGLE_PROJECT_ID"); ok {
			loggingClient, err := gcplogging.NewClient(ctx, projectID)
			if err != nil {
				_ = level.Error(logger).Log("err", err)
				os.Exit(1)
			}

			logger = logging.NewStackdriverLogger(loggingClient, "route_matrix")
		}
	}

	{
		switch os.Getenv("LOG_LEVEL") {
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
		_ = level.Error(logger).Log("err", "ENV not set")
		os.Exit(1)
	}

	fmt.Printf("env is %s\n", env)

	shutdown := make(chan bool)
	var err error
	var matrixSvcTimeVariance int64 = 2000
	if timeStr, ok := os.LookupEnv("MATRIX_SVC_TIME_VARIANCE"); ok{
		matrixSvcTimeVariance, err = strconv.ParseInt(timeStr, 10, 64)
		if err != nil {
			_ = level.Error(logger).Log("msg", "unable to parse Matrix time variance", "err", err)
		}
	}

	var optimizerTimeVariance int64 = 5000
	if timeStr, ok := os.LookupEnv("OPTIMIZER_TIME_VARIANCE"); ok{
		optimizerTimeVariance, err = strconv.ParseInt(timeStr, 10, 64)
		if err != nil {
			_ = level.Error(logger).Log("msg", "unable to parse optimizer time variance", "err", err)
		}
	}

	store := storage.NewRedisMatrixStore()
	svc, err := rm.New(store, matrixSvcTimeVariance, optimizerTimeVariance)
	if err != nil {
		_ = level.Error(logger).Log("err", err)
		os.Exit(1)
	}

	//update matrix service
	go func() {
		errorCount := 0
		for {
			ticker := time.NewTicker(250 * time.Millisecond)
			select {
			case <-shutdown:
				return
			case <-ticker.C:
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
		}
	}()

	//core loop
	go func() {
		ticker := time.NewTicker(1000 * time.Millisecond)
		for {
			select {
			case <-shutdown:
				return
			case <-ticker.C:
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
		}
	}()

	// Wait for interrupt signal
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)

	select{
		case <-sigint:
			shutdown <- true
			time.Sleep(5 * time.Second)
	}

}
