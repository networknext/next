/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"context"
	"expvar"
	"fmt"
	"net"
	"net/http"
	_ "net/http/pprof"
	"runtime"
	"strconv"
	"sync"
	"syscall"
	"time"

	"os"
	"os/signal"

	"github.com/go-kit/kit/log/level"
	"github.com/gorilla/mux"

	"github.com/networknext/backend/modules/backend"
	"github.com/networknext/backend/modules/beacon"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/transport"
	"golang.org/x/sys/unix"

	googlepubsub "cloud.google.com/go/pubsub"
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
	fmt.Printf("beacon: Git Hash: %s - Commit: %s\n", sha, commitMessage)

	serviceName := "beacon"

	ctx := context.Background()

	gcpProjectID := backend.GetGCPProjectID()
	gcpOK := gcpProjectID != ""

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

	// Get metrics handler
	metricsHandler, err := backend.GetMetricsHandler(ctx, logger, gcpProjectID)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	// Create beacon metrics
	beaconServiceMetrics, err := metrics.NewBeaconServiceMetrics(ctx, metricsHandler)
	if err != nil {
		level.Error(logger).Log("msg", "failed to create beacon service metrics", "err", err)
		return 1
	}

	if gcpOK {
		// Stackdriver Profiler
		if err := backend.InitStackDriverProfiler(gcpProjectID, serviceName, env); err != nil {
			level.Error(logger).Log("msg", "failed to initialze StackDriver profiler", "err", err)
			return 1
		}
	}

	// Create a local beaconer
	var beaconer beacon.Beaconer = &beacon.LocalBeaconer{
		Logger:  logger,
		Metrics: beaconServiceMetrics.BeaconMetrics,
	}

	pubsubEmulatorOK := envvar.Exists("PUBSUB_EMULATOR_HOST")
	if gcpOK || pubsubEmulatorOK {

		pubsubCtx := ctx
		if pubsubEmulatorOK {
			gcpProjectID = "local"

			var cancelFunc context.CancelFunc
			pubsubCtx, cancelFunc = context.WithDeadline(ctx, time.Now().Add(5*time.Second))
			defer cancelFunc()

			level.Info(logger).Log("msg", "Detected pubsub emulator")
		}

		// Google Pubsub
		{
			clientCount, err := envvar.GetInt("BEACON_CLIENT_COUNT", 1)
			if err != nil {
				level.Error(logger).Log("err", err)
				return 1
			}

			countThreshold, err := envvar.GetInt("BEACON_BATCHED_MESSAGE_COUNT", 10)
			if err != nil {
				level.Error(logger).Log("err", err)
				return 1
			}

			byteThreshold, err := envvar.GetInt("BEACON_BATCHED_MESSAGE_MIN_BYTES", 512)
			if err != nil {
				level.Error(logger).Log("err", err)
				return 1
			}

			// We do our own batching so don't stack the library's batching on top of ours
			// Specifically, don't stack the message count thresholds
			settings := googlepubsub.DefaultPublishSettings
			settings.CountThreshold = 1
			settings.ByteThreshold = byteThreshold
			settings.NumGoroutines = runtime.GOMAXPROCS(0)

			pubsub, err := beacon.NewGooglePubSubBeaconer(pubsubCtx, beaconServiceMetrics.BeaconMetrics, logger, gcpProjectID, "beacon", clientCount, countThreshold, byteThreshold, &settings)
			if err != nil {
				level.Error(logger).Log("msg", "could not create pubsub beaconer", "err", err)
				return 1
			}

			beaconer = pubsub
		}
	}

	channelBufferSize, err := envvar.GetInt("CHANNEL_BUFFER_SIZE", 100000)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}
	numGoroutines, err := envvar.GetInt("NUM_GOROUTINES", 1)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	var wg sync.WaitGroup
	// Create error channel to error out from any goroutines
	errChan := make(chan error, 1)

	// Create an internal channel to receive beacon packets and submit them
	beaconPacketChan := make(chan *transport.NextBeaconPacket, channelBufferSize)
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			for {
				select {
				case beaconPacket := <-beaconPacketChan:
					// Record beacon packet stats
					if beaconPacket.Next {
						beaconServiceMetrics.BeaconMetrics.NextEntries.Add(1)
					} else {
						beaconServiceMetrics.BeaconMetrics.DirectEntries.Add(1)
					}
					if beaconPacket.Upgraded {
						beaconServiceMetrics.BeaconMetrics.UpgradedEntries.Add(1)
					} else {
						beaconServiceMetrics.BeaconMetrics.NotUpgradedEntries.Add(1)
					}
					if beaconPacket.Enabled {
						beaconServiceMetrics.BeaconMetrics.EnabledEntries.Add(1)
					} else {
						beaconServiceMetrics.BeaconMetrics.NotEnabledEntries.Add(1)
					}
					if beaconPacket.FallbackToDirect {
						beaconServiceMetrics.BeaconMetrics.FallbackToDirect.Add(1)
					}

					// Submit beacon packet
					err := beaconer.Submit(ctx, beaconPacket)
					if err != nil {
						level.Error(logger).Log("msg", "Could not send beacon packet to Google Pubsub", "err", err)
						beaconServiceMetrics.BeaconMetrics.ErrorMetrics.BeaconSendFailure.Add(1)
						errChan <- err
						return
					}

					beaconServiceMetrics.BeaconMetrics.EntriesSent.Add(1)
				case <-ctx.Done():
					level.Error(logger).Log("err", ctx.Err())
					errChan <- ctx.Err()
					return
				}
			}
		}()
	}

	// Setup the stats print routine
	{
		memoryUsed := func() float64 {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			return float64(m.Alloc) / (1000.0 * 1000.0)
		}

		go func() {
			for {
				beaconServiceMetrics.ServiceMetrics.Goroutines.Set(float64(runtime.NumGoroutine()))
				beaconServiceMetrics.ServiceMetrics.MemoryAllocated.Set(memoryUsed())

				fmt.Printf("-----------------------------\n")
				fmt.Printf("%d goroutines\n", int(beaconServiceMetrics.ServiceMetrics.Goroutines.Value()))
				fmt.Printf("%.2f mb allocated\n", beaconServiceMetrics.ServiceMetrics.MemoryAllocated.Value())
				fmt.Printf("%d invocations\n", int(beaconServiceMetrics.HandlerMetrics.Invocations.Value()))
				fmt.Printf("%d beacon entries received\n", int(beaconServiceMetrics.BeaconMetrics.EntriesReceived.Value()))
				fmt.Printf("%d beacon entries sent\n", int(beaconServiceMetrics.BeaconMetrics.EntriesSent.Value()))
				fmt.Printf("%d beacon entries submitted\n", int(beaconServiceMetrics.BeaconMetrics.EntriesSubmitted.Value()))
				fmt.Printf("%d beacon entries flushed\n", int(beaconServiceMetrics.BeaconMetrics.EntriesFlushed.Value()))
				fmt.Printf("%d beacon entries on next\n", int(beaconServiceMetrics.BeaconMetrics.NextEntries.Value()))
				fmt.Printf("%d beacon entries on direct\n", int(beaconServiceMetrics.BeaconMetrics.DirectEntries.Value()))
				fmt.Printf("%d beacon entries upgraded\n", int(beaconServiceMetrics.BeaconMetrics.UpgradedEntries.Value()))
				fmt.Printf("%d beacon entries not upgraded\n", int(beaconServiceMetrics.BeaconMetrics.NotUpgradedEntries.Value()))
				fmt.Printf("%d beacon entries enabled\n", int(beaconServiceMetrics.BeaconMetrics.EnabledEntries.Value()))
				fmt.Printf("%d beacon entries not enabled\n", int(beaconServiceMetrics.BeaconMetrics.NotEnabledEntries.Value()))
				fmt.Printf("%d beacon entries fallen back to direct\n", int(beaconServiceMetrics.BeaconMetrics.FallbackToDirect.Value()))
				fmt.Printf("%d beacon entry send failures\n", int(beaconServiceMetrics.BeaconMetrics.ErrorMetrics.BeaconSendFailure.Value()))
				fmt.Printf("%d beacon entry channel full\n", int(beaconServiceMetrics.BeaconMetrics.ErrorMetrics.BeaconChannelFull.Value()))
				fmt.Printf("%d beacon entry publish failure\n", int(beaconServiceMetrics.BeaconMetrics.ErrorMetrics.BeaconPublishFailure.Value()))
				fmt.Printf("-----------------------------\n")

				time.Sleep(time.Second * 10)
			}
		}()
	}

	// Start HTTP server
	{
		router := mux.NewRouter()
		router.HandleFunc("/health", transport.HealthHandlerFunc())
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
			httpPort := envvar.Get("HTTP_PORT", "40001")

			err := http.ListenAndServe(":"+httpPort, router)
			if err != nil {
				level.Error(logger).Log("err", err)
				errChan <- err
				return
			}
		}()
	}

	numThreads, err := envvar.GetInt("NUM_THREADS", 1)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	readBuffer, err := envvar.GetInt("READ_BUFFER", 100000)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	writeBuffer, err := envvar.GetInt("WRITE_BUFFER", 100000)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	udpPort := envvar.Get("UDP_PORT", "30000")

	wg.Add(numThreads)

	lc := net.ListenConfig{
		Control: func(network string, address string, c syscall.RawConn) error {
			err := c.Control(func(fileDescriptor uintptr) {
				err := unix.SetsockoptInt(int(fileDescriptor), unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
				if err != nil {
					panic(fmt.Sprintf("failed to set reuse address socket option: %v", err))
				}

				err = unix.SetsockoptInt(int(fileDescriptor), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
				if err != nil {
					panic(fmt.Sprintf("failed to set reuse port socket option: %v", err))
				}
			})

			return err
		},
	}

	port, _ := strconv.Atoi(udpPort)

	level.Info(logger).Log("msg", "Started beacon on port", "port", port)

	for i := 0; i < numThreads; i++ {
		go func(thread int) {
			lp, err := lc.ListenPacket(ctx, "udp", "0.0.0.0:"+udpPort)
			if err != nil {
				panic(fmt.Sprintf("could not bind socket: %v", err))
			}

			conn := lp.(*net.UDPConn)
			defer conn.Close()

			if err := conn.SetReadBuffer(readBuffer); err != nil {
				panic(fmt.Sprintf("could not set connection read buffer size: %v", err))
			}

			if err := conn.SetWriteBuffer(writeBuffer); err != nil {
				panic(fmt.Sprintf("could not set connection write buffer size: %v", err))
			}

			dataArray := [transport.DefaultMaxPacketSize]byte{}

			var beaconPacket *transport.NextBeaconPacket

			for {
				beaconPacket = &transport.NextBeaconPacket{}
				data := dataArray[:]
				size, _, err := conn.ReadFromUDP(data)
				if err != nil {
					level.Error(logger).Log("msg", "failed to read UDP packet", "err", err)
					beaconServiceMetrics.BeaconMetrics.ErrorMetrics.BeaconReadPacketFailure.Add(1)
					break
				}

				if size <= 1 {
					continue
				}

				data = data[:size]

				// Check if we received a non-beacon packet
				if data[0] != transport.PacketTypeBeacon {
					level.Error(logger).Log("err", "unknown packet type", "packet_type", data[0])
					beaconServiceMetrics.BeaconMetrics.NonBeaconPacketsReceived.Add(1)
					continue
				}

				// Start timer for packet processing
				timeStart := time.Now()

				err = transport.ReadBeaconEntry(beaconPacket, data[1:])
				if err != nil {
					level.Error(logger).Log("msg", "failed to serialize beacon packet", "err", err)
					beaconServiceMetrics.BeaconMetrics.ErrorMetrics.BeaconSerializePacketFailure.Add(1)
					continue
				}

				// Finish timing packet processing
				milliseconds := float64(time.Since(timeStart).Milliseconds())
				beaconServiceMetrics.HandlerMetrics.Duration.Set(milliseconds)

				if milliseconds > 100 {
					beaconServiceMetrics.HandlerMetrics.LongDuration.Add(1)
				}

				// Insert packet into internal channel for local or bigquery
				select {
				case beaconPacketChan <- beaconPacket:
					beaconServiceMetrics.BeaconMetrics.EntriesReceived.Add(1)
				default:
					level.Error(logger).Log("err", "Beacon channel full")
					beaconServiceMetrics.BeaconMetrics.ErrorMetrics.BeaconChannelFull.Add(1)
					continue
				}
			}

			wg.Done()
		}(i)
	}

	level.Info(logger).Log("msg", "waiting for incoming connections")

	// Wait for interrupt signal
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)

	select {
	case <-sigint:
		return 0
	case <-ctx.Done():
		// Let the goroutines finish up
		wg.Wait()
		return 1
	case <-errChan: // Exit with an error code of 1 if we receive any errors from goroutines
		return 1
	}
}
