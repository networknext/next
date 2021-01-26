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
	"github.com/networknext/backend/modules/encoding"
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

	// Create a local beaconer
	var beaconer beacon.Beaconer = &beacon.LocalBeaconer{
		Logger:  logger,
		Metrics: &beaconServiceMetrics.BeaconMetrics,
	}

	pubsubEmulatorOK := envvar.Exists("PUBSUB_EMULATOR_HOST")
	if gcpOK || pubsubEmulatorOK {

		pubsubCtx := ctx
		if pubsubEmulatorOK {
			gcpProjectID = "local"

			var cancelFunc context.CancelFunc
			pubsubCtx, cancelFunc = context.WithDeadline(ctx, time.Now().Add(10*time.Second))
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

			pubsub, err := beacon.NewGooglePubSubBeaconer(pubsubCtx, &beaconServiceMetrics.BeaconMetrics, logger, gcpProjectID, "beacon", clientCount, countThreshold, byteThreshold, &settings)
			if err != nil {
				level.Error(logger).Log("msg", "could not create pubsub beaconer", "err", err)
				return 1
			}

			beaconer = pubsub
		}
	}

	if gcpOK {
		// Stackdriver Profiler
		if err := backend.InitStackDriverProfiler(gcpProjectID, serviceName, env); err != nil {
			level.Error(logger).Log("msg", "failed to initialze StackDriver profiler", "err", err)
			return 1
		}

		// // Google Bigquery
		// {
		// 	beaconDataset := envvar.Get("GOOGLE_BIGQUERY_DATASET_BEACON", "")
		// 	if beaconDataset != "" {
		// 		batchSize := beacon.DefaultBigQueryBatchSize

		// 		batchSize, err := envvar.GetInt("GOOGLE_BIGQUERY_BATCH_SIZE", beacon.DefaultBigQueryBatchSize)
		// 		if err != nil {
		// 			level.Error(logger).Log("err", err)
		// 			return 1
		// 		}

		// 		// Create biquery client and start write loop
		// 		bqClient, err := bigquery.NewClient(ctx, gcpProjectID)
		// 		if err != nil {
		// 			level.Error(logger).Log("err", err)
		// 			return 1
		// 		}

		// 		beaconTable := envvar.Get("GOOGLE_BIGQUERY_TABLE_BEACON", "")

		// 		b := beacon.GoogleBigQueryClient{
		// 			Metrics:       &beaconServiceMetrics.BeaconMetrics,
		// 			Logger:        logger,
		// 			TableInserter: bqClient.Dataset(beaconDataset).Table(beaconTable).Inserter(),
		// 			BatchSize:     batchSize,
		// 		}

		// 		// Set the Beaconer to BigQuery
		// 		beaconer = &b

		// 		// Start the background WriteLoop to batch write to BigQuery
		// 		go func() {
		// 			b.WriteLoop(ctx)
		// 		}()
		// 	}
		// }
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
				default:
				}
			}
		}()
	}

	// TODO: setup stackdriver metrics

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
				fmt.Printf("%d beacon entries received\n", int(beaconServiceMetrics.BeaconMetrics.EntriesReceived.Value()))
				fmt.Printf("%d beacon entries sent\n", int(beaconServiceMetrics.BeaconMetrics.EntriesSent.Value()))
				fmt.Printf("%d beacon entries submitted\n", int(beaconServiceMetrics.BeaconMetrics.EntriesSubmitted.Value()))
				fmt.Printf("%d beacon entries queued\n", int(beaconServiceMetrics.BeaconMetrics.EntriesQueued.Value()))
				fmt.Printf("%d beacon entries flushed\n", int(beaconServiceMetrics.BeaconMetrics.EntriesFlushed.Value()))
				fmt.Printf("%d beacon entry send failures\n", int(beaconServiceMetrics.BeaconMetrics.ErrorMetrics.BeaconSendFailure.Value()))
				fmt.Printf("%d beacon entry channel full\n", int(beaconServiceMetrics.BeaconMetrics.ErrorMetrics.BeaconChannelFull.Value()))
				fmt.Printf("%d beacon entry write failure\n", int(beaconServiceMetrics.BeaconMetrics.ErrorMetrics.BeaconWriteFailure.Value()))
				fmt.Printf("-----------------------------\n")

				time.Sleep(time.Second * 10)
			}
		}()
	}

	// Start HTTP server
	{
		router := mux.NewRouter()
		router.HandleFunc("/health", transport.HealthHandlerFunc())
		router.HandleFunc("/version", transport.VersionHandlerFunc(buildtime, sha, tag, commitMessage, false, []string{}))
		router.Handle("/debug/vars", expvar.Handler())

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

	fmt.Printf("\nstarted beacon on port %d\n\n", port)

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

			beaconPacket := &transport.NextBeaconPacket{}

			for {
				data := dataArray[:]
				size, fromAddr, err := conn.ReadFromUDP(data)
				if err != nil {
					level.Error(logger).Log("msg", "failed to read UDP packet", "err", err)
					break
				}

				if size <= 1 {
					continue
				}

				data = data[:size]

				if data[0] != transport.PacketTypeBeacon {
					continue
				}

				readStream := encoding.CreateReadStream(data[1:])
				err = beaconPacket.Serialize(readStream)
				if err != nil {
					fmt.Printf("error reading beacon packet: %v\n", err)
					continue
				}

				// Set the timestamp of the packet
				beaconPacket.Timestamp = uint64(time.Now().Unix())

				fmt.Printf("beacon packet: %d, %v, %x, %x, %x, %x, %x, %d, %d, %v, %v, %v, %v\n",
					beaconPacket.Version,
					beaconPacket.Timestamp,
					beaconPacket.CustomerID,
					beaconPacket.DatacenterID,
					beaconPacket.UserHash,
					beaconPacket.AddressHash,
					beaconPacket.SessionID,
					beaconPacket.PlatformID,
					beaconPacket.ConnectionType,
					beaconPacket.Enabled,
					beaconPacket.Upgraded,
					beaconPacket.Next,
					beaconPacket.FallbackToDirect,
				)

				// Not sure if we need fromAddr for anything else
				_ = fromAddr

				// Insert packet into internal channel for local or bigquery
				select {
				case beaconPacketChan <- beaconPacket:
					beaconServiceMetrics.BeaconMetrics.EntriesReceived.Add(1)
				default:
					beaconServiceMetrics.BeaconMetrics.ErrorMetrics.BeaconChannelFull.Add(1)
					continue
				}

				// TODO: write metrics to stackdriver / local metrics
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
