/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"bytes"
	"context"
	"expvar"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"time"

	"cloud.google.com/go/pubsub"
	"github.com/gorilla/mux"
	"github.com/networknext/backend/backend"
	"github.com/networknext/backend/modules/analytics"
	"github.com/networknext/backend/modules/common/helpers"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/transport"



	"github.com/go-kit/kit/log/level"

	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/routing"
)

var (
	buildtime     string
	commitMessage string
	sha           string
	tag           string
)

// Allows us to return an exit code and allows log flushes and deferred functions
// to finish before exiting.
func main() {
	os.Exit(mainReturnWithCode())
}

func mainReturnWithCode() int {
	serviceName := "Optimizer"
	fmt.Printf("%s: Git Hash: %s - Commit: %s\n", serviceName, sha, commitMessage)

	ctx := context.Background()
	gcpProjectID := backend.GetGCPProjectID()
	logger, err := backend.GetLogger(ctx, gcpProjectID, serviceName)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	env, err := backend.GetEnv()
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	if gcpProjectID != "" {
		if err := backend.InitStackDriverProfiler(gcpProjectID, serviceName, env); err != nil {
			level.Error(logger).Log("msg", "failed to initialze StackDriver profiler", "err", err)
			return 1
		}
	}

	storer, err := backend.GetStorer(ctx, logger, gcpProjectID, env)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	metricsHandler, err := backend.GetMetricsHandler(ctx, logger, gcpProjectID)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	cfg, err := GetConfig()
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	metrics, err, msg := NewMetrics(ctx, metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", msg, "err", err)
	}

	statsDB := routing.NewStatsDatabase()

	// Create the relay map
	cleanupCallback := func(relayData *routing.RelayData) error {
		// Remove relay entry from statsDB (which in turn means it won't appear in cost matrix)
		statsDB.DeleteEntry(relayData.ID)
		level.Warn(logger).Log("msg", "relay timed out", "relay ID", relayData.ID, "relay addr", relayData.Addr.String(), "relay name", relayData.Name)
		return nil
	}

	relayMap := routing.NewRelayMap(cleanupCallback)
	go func() {
		timeout := int64(routing.RelayTimeout.Seconds())
		frequency := time.Second
		ticker := time.NewTicker(frequency)
		relayMap.TimeoutLoop(ctx, timeout, ticker.C)
	}()

	optimizer := New(cfg, metrics, relayMap, storer, statsDB, logger)

	// ping stats
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

				pubsub, err := analytics.NewGooglePubSubPingStatsPublisher(pubsubCtx, &relayBackendMetrics.PingStatsMetrics, logger, gcpProjectID, "ping_stats", settings)
				if err != nil {
					level.Error(logger).Log("msg", "could not create analytics pubsub publisher", "err", err)
					return 1
				}

				pingStatsPublisher = pubsub
			}
		}

		go func() {
			publishInterval, err := envvar.GetDuration("PING_STATS_PUBLISH_INTERVAL", time.Minute)
			if err != nil {
				level.Error(logger).Log("err", err)
				os.Exit(1) // todo: don't os.Exit() here, but find a way to exit
			}

			syncTimer := helpers.NewSyncTimer(publishInterval)
			for {
				syncTimer.Run()

				cpy := statsDB.MakeCopy()
				entries := analytics.ExtractPingStats(cpy)
				if err := pingStatsPublisher.Publish(ctx, entries); err != nil {
					level.Error(logger).Log("err", err)
					os.Exit(1) // todo: don't os.Exit() here, but find a way to exit
				}
			}
		}()
	}

	// relay stats
	var relayStatsPublisher analytics.RelayStatsPublisher = &analytics.NoOpRelayStatsPublisher{}
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

				pubsub, err := analytics.NewGooglePubSubRelayStatsPublisher(pubsubCtx, &relayBackendMetrics.RelayStatsMetrics, logger, gcpProjectID, "relay_stats", settings)
				if err != nil {
					level.Error(logger).Log("msg", "could not create analytics pubsub publisher", "err", err)
					return 1
				}

				relayStatsPublisher = pubsub
			}
		}

		go func() {
			publishInterval, err := envvar.GetDuration("RELAY_STATS_PUBLISH_INTERVAL", time.Second*10)
			if err != nil {
				level.Error(logger).Log("err", err)
				os.Exit(1) // todo: don't os.Exit() here, but find a way to exit
			}

			syncTimer := NewSyncTimer(publishInterval)
			for {
				syncTimer.Run()
				allRelayData := relayMap.GetAllRelayData()
				entries := make([]analytics.RelayStatsEntry, len(allRelayData))

				count := 0
				for _, relay := range allRelayData {
					// convert peak to mbps

					var traffic routing.TrafficStats

					relay.TrafficMu.Lock()
					for i := range relay.TrafficStatsBuff {
						stats := &relay.TrafficStatsBuff[i]
						traffic = traffic.Add(stats)
					}
					relay.TrafficStatsBuff = relay.TrafficStatsBuff[:0]
					numSessions := relay.PeakTrafficStats.SessionCount
					envUp := relay.PeakTrafficStats.EnvelopeUpKbps
					envDown := relay.PeakTrafficStats.EnvelopeDownKbps
					relay.PeakTrafficStats.SessionCount = 0
					relay.PeakTrafficStats.EnvelopeUpKbps = 0
					relay.PeakTrafficStats.EnvelopeDownKbps = 0
					relay.TrafficMu.Unlock()

					elapsed := time.Since(relay.LastStatsPublishTime)
					relay.LastStatsPublishTime = time.Now()

					fsrelay, err := storer.Relay(relay.ID)
					if err != nil {
						continue
					}

					// use the sum of all the stats since the last publish here and convert to mbps
					bwSentMbps := float32(float64(traffic.AllTx()) * 8.0 / 1000000.0 / elapsed.Seconds())
					bwRecvMbps := float32(float64(traffic.AllRx()) * 8.0 / 1000000.0 / elapsed.Seconds())

					// use the peak envelope values here and convert, it's already per second so no need for time adjustment
					envSentMbps := float32(float64(envUp) / 1000.0)
					envRecvMbps := float32(float64(envDown) / 1000.0)

					var numRouteable uint32 = 0
					for _, otherRelay := range allRelayData {
						if relay.ID == otherRelay.ID {
							continue
						}

						rtt, jitter, pl := statsDB.GetSample(relay.ID, otherRelay.ID)
						if rtt != routing.InvalidRouteValue && jitter != routing.InvalidRouteValue && pl != routing.InvalidRouteValue {
							numRouteable++
						}
					}

					var bwSentPercent float32
					var bwRecvPercent float32
					var envSentPercent float32
					var envRecvPercent float32
					if fsrelay.NICSpeedMbps != 0 {
						bwSentPercent = bwSentMbps / float32(fsrelay.NICSpeedMbps) * 100.0
						bwRecvPercent = bwRecvMbps / float32(fsrelay.NICSpeedMbps) * 100.0
						envSentPercent = envSentMbps / float32(fsrelay.NICSpeedMbps) * 100.0
						envRecvPercent = envRecvMbps / float32(fsrelay.NICSpeedMbps) * 100.0
					}

					entries[count] = analytics.RelayStatsEntry{
						ID:                       relay.ID,
						CPUUsage:                 relay.CPUUsage,
						MemUsage:                 relay.MemUsage,
						BandwidthSentPercent:     bwSentPercent,
						BandwidthReceivedPercent: bwRecvPercent,
						EnvelopeSentPercent:      envSentPercent,
						EnvelopeReceivedPercent:  envRecvPercent,
						BandwidthSentMbps:        bwSentMbps,
						BandwidthReceivedMbps:    bwRecvMbps,
						EnvelopeSentMbps:         envSentMbps,
						EnvelopeReceivedMbps:     envRecvMbps,
						NumSessions:              uint32(numSessions),
						MaxSessions:              relay.MaxSessions,
						NumRoutable:              numRouteable,
						NumUnroutable:            uint32(len(allRelayData)) - 1 - numRouteable,
					}

					count++
				}

				if err := relayStatsPublisher.Publish(ctx, entries[:count]); err != nil {
					level.Error(logger).Log("err", err)
					os.Exit(1) // todo: don't os.Exit() here, but find a way to exit
				}
			}
		}()

	}

	syncInterval, err := envvar.GetDuration("COST_MATRIX_INTERVAL", time.Second)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	costMatrixData := new(helpers.MatrixData)
	routeMatrixData := new(helpers.MatrixData)
	valveRouteMatrixData := new(helpers.MatrixData)

	// Generate the route matrix
	go func() {
		syncTimer := helpers.NewSyncTimer(syncInterval)
		for {

			syncTimer.Run()

			costMatrix, routeMatrix := optimizer.GetRouteMatrix()
			if costMatrix == nil || routeMatrix == nil {
				continue
			}

			costMatrixData.SetMatrix(costMatrix.GetResponseData())
			routeMatrixData.SetMatrix(routeMatrix.GetResponseData())

			optimizer.MetricsOutput()

		}
	}()

	// Generate the route matrix specifically for valve
	go func() {
		syncTimer := helpers.NewSyncTimer(syncInterval)
		for {
			syncTimer.Run()

			_, routeMatrix := optimizer.GetValveRouteMatrix()
			if routeMatrix == nil {
				continue
			}

			valveRouteMatrixData.SetMatrix(routeMatrix.GetResponseData())
		}
	}()

	serveRouteMatrixFunc := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")

		buffer := bytes.NewBuffer(routeMatrixData.GetMatrix())
		_, err := buffer.WriteTo(w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	serveValveRouteMatrixFunc := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")

		buffer := bytes.NewBuffer(valveRouteMatrixData.GetMatrix())
		_, err := buffer.WriteTo(w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	serveCostMatrixFunc := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")

		buffer := bytes.NewBuffer(costMatrixData.GetMatrix())
		_, err := buffer.WriteTo(w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}

	fmt.Printf("starting http server\n")

	router := mux.NewRouter()
	router.HandleFunc("/health", transport.HealthHandlerFunc())
	router.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage))
	router.Handle("/debug/vars", expvar.Handler())

	router.HandleFunc("/cost_matrix", serveCostMatrixFunc).Methods("GET")
	router.HandleFunc("/route_matrix", serveRouteMatrixFunc).Methods("GET")
	router.HandleFunc("/route_matrix_valve", serveValveRouteMatrixFunc).Methods("GET")

	router.HandleFunc("/relay_dashboard", transport.RelayDashboardHandlerFunc(relayMap, getRouteMatrixFunc, statsdb, "local", "local", maxJitter))
	router.HandleFunc("/relay_stats", transport.RelayStatsFunc(logger, relayMap))

	go func() {
		port := envvar.Get("PORT", "30000")

		level.Info(logger).Log("addr", ":"+port)

		err := http.ListenAndServe(":"+port, router)
		if err != nil {
			level.Error(logger).Log("err", err)
			os.Exit(1) // todo: don't os.Exit() here, but find a way to exit
		}
	}()

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint

	return 0
}

