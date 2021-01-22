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
	"github.com/networknext/backend/modules/transport"
	"golang.org/x/sys/unix"
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
	}

	if gcpOK {
		// Stackdriver Profiler
		if err := backend.InitStackDriverProfiler(gcpProjectID, serviceName, env); err != nil {
			level.Error(logger).Log("msg", "failed to initialze StackDriver profiler", "err", err)
			return 1
		}

		// Google Bigquery
		{
			beaconDataset := envvar.Get("GOOGLE_BIGQUERY_DATASET_BEACON", "")
			if beaconDataset != "" {
				batchSize := beacon.DefaultBigQueryBatchSize

				batchSize, err := envvar.GetInt("GOOGLE_BIGQUERY_BATCH_SIZE", beacon.DefaultBigQueryBatchSize)
				if err != nil {
					level.Error(logger).Log("err", err)
					return 1
				}

				// Create biquery client and start write loop
				bqClient, err := bigquery.NewClient(ctx, gcpProjectID)
				if err != nil {
					level.Error(logger).Log("err", err)
					return 1
				}

				beaconTable := envvar.Get("GOOGLE_BIGQUERY_TABLE_BEACON", "")

				b := beacon.GoogleBigQueryClient{
					Metrics:       &beaconServiceMetrics.BeaconMetrics,
					Logger:        logger,
					TableInserter: bqClient.Dataset(beaconDataset).Table(beaconTable).Inserter(),
					BatchSize:     batchSize,
				}

				// Start the background WriteLoop to batch write to BigQuery
				go func() {
					b.WriteLoop(ctx)
				}()
			}
		}
	}

	// TODO: setup stackdriver metrics
	// TODO: googlepubsub?

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
				fmt.Printf("%d beacon entries submitted\n", int(beaconServiceMetrics.BeaconMetrics.EntriesSubmitted.Value()))
				fmt.Printf("%d beacon entries queued\n", int(beaconServiceMetrics.BeaconMetrics.EntriesQueued.Value()))
				fmt.Printf("%d beacon entries flushed\n", int(beaconServiceMetrics.BeaconMetrics.EntriesFlushed.Value()))
				fmt.Printf("-----------------------------\n")

				time.Sleep(time.Second * 10)
			}
		}()
	}

	// Start HTTP server
	{
		router := mux.NewRouter()
		router.HandleFunc("/health", transport.HealthHandlerFunc())
		router.Handle("/debug/vars", expvar.Handler())

		go func() {
			httpPort := envvar.Get("HTTP_PORT", "40001")

			err := http.ListenAndServe(":"+httpPort, router)
			if err != nil {
				level.Error(logger).Log("err", err)
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

	var wg sync.WaitGroup

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

			packet := beacon.NextBeaconPacket{}

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

				if data[0] != 118 {
					continue
				}

				readStream := encoding.CreateReadStream(data[1:])
				err = packet.Serialize(readStream)
				if err != nil {
					fmt.Printf("error reading beacon packet: %v\n", err)
					continue
				}

				fmt.Printf("beacon packet: %x, %x, %x, %x, %x, %d, %d, %v, %v, %v, %v\n",
					packet.CustomerID,
					packet.DatacenterID,
					packet.UserHash,
					packet.AddressHash,
					packet.SessionID,
					packet.PlatformID,
					packet.ConnectionType,
					packet.Enabled,
					packet.Upgraded,
					packet.Next,
					packet.FallbackToDirect,
				)

				// todo
				_ = fromAddr

				// TODO: insert into bigquery if gcpOK
				// TODO: write metrics to stackdriver / local metrics 

			}

			wg.Done()
		}(i)
	}

	level.Info(logger).Log("msg", "waiting for incoming connections")

	// Wait for interrupt signal
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint

	return 0
}
