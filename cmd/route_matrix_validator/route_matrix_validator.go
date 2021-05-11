package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"runtime"
	"time"

	"github.com/go-kit/kit/log/level"

	"github.com/networknext/backend/modules/backend"
	"github.com/networknext/backend/modules/common/helpers"
	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/routing"
)

func main() {

	// WIP load testing validator.
	serviceName := "route_matrix_validator"

	ctx := context.Background()

	gcpProjectID := backend.GetGCPProjectID()

	logger, err := backend.GetLogger(ctx, gcpProjectID, serviceName)
	if err != nil {
		level.Error(logger).Log("err", err)
		return
	}

	env, err := backend.GetEnv()
	if err != nil {
		level.Error(logger).Log("err", err)
		return
	}

	metricsHandler, err := backend.GetMetricsHandler(ctx, logger, gcpProjectID)
	if err != nil {
		level.Error(logger).Log("err", err)
		return
	}

	// Create server backend metrics
	backendMetrics, err := metrics.NewServerBackendMetrics(ctx, metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create server_backend metrics", "err", err)
		return
	}

	if gcpProjectID != "" {
		if err := backend.InitStackDriverProfiler(gcpProjectID, serviceName, env); err != nil {
			level.Error(logger).Log("msg", "failed to initialze StackDriver profiler", "err", err)
			return
		}
	}

	// Create a goroutine to update metrics
	go func() {
		memoryUsed := func() float64 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			return float64(m.Alloc) / (1000.0 * 1000.0)
		}

		for {
			backendMetrics.ServiceMetrics.Goroutines.Set(float64(runtime.NumGoroutine()))
			backendMetrics.ServiceMetrics.MemoryAllocated.Set(memoryUsed())

			time.Sleep(time.Second * 10)
		}
	}()

	uri := envvar.Get("RELAY_FRONTEND_URI", "")
	if uri != "" {
		level.Error(logger).Log("err", fmt.Errorf("no matrix uri specified"))
		return
	}

	syncInterval, err := envvar.GetDuration("ROUTE_MATRIX_SYNC_INTERVAL", time.Second)
	if err != nil {
		level.Error(logger).Log("err", err)
		return
	}

	go func() {
		httpClient := &http.Client{
			Timeout: time.Second * 2,
		}

		valveBackend, err := envvar.GetBool("VALVE_SERVER_BACKEND", false)
		if err != nil {
			level.Error(logger).Log("err", err)
		}

		if valveBackend {
			uri = fmt.Sprintf("%s_%s", uri, "valve")
		}

		syncTimer := helpers.NewSyncTimer(syncInterval)
		for {
			syncTimer.Run()

			var buffer []byte
			start := time.Now()

			var routeEntriesReader io.ReadCloser

			// Default to reading route matrix from file
			if f, err := os.Open(uri); err == nil {
				routeEntriesReader = f
			}

			// Prefer to get it remotely if possible
			if r, err := httpClient.Get(uri); err == nil {
				routeEntriesReader = r.Body
			}

			if routeEntriesReader == nil {
				continue
			}

			buffer, err = ioutil.ReadAll(routeEntriesReader)

			if routeEntriesReader != nil {
				routeEntriesReader.Close()
			}

			if err != nil {
				level.Error(logger).Log("envvar", "ROUTE_MATRIX_URI", "value", uri, "msg", "could not read route matrix", "err", err)
				continue // Don't swap route matrix if we fail to read
			}

			var newRouteMatrix routing.RouteMatrix
			if len(buffer) > 0 {
				rs := encoding.CreateReadStream(buffer)
				if err := newRouteMatrix.Serialize(rs); err != nil {
					level.Error(logger).Log("msg", "could not serialize route matrix", "err", err)
					continue // Don't swap route matrix if we fail to serialize
				}
			}

			routeEntriesTime := time.Since(start)

			duration := float64(routeEntriesTime.Milliseconds())
			backendMetrics.RouteMatrixUpdateDuration.Set(duration)

			if duration > 100 {
				backendMetrics.RouteMatrixUpdateLongDuration.Add(1)
			}

			numRoutes := int32(0)
			for i := range newRouteMatrix.RouteEntries {
				numRoutes += newRouteMatrix.RouteEntries[i].NumRoutes
			}
			backendMetrics.RouteMatrixNumRoutes.Set(float64(numRoutes))
			backendMetrics.RouteMatrixBytes.Set(float64(len(buffer)))

			// todo need a way to check time created and version
		}
	}()

}
