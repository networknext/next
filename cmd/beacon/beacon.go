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
	"strconv"
	"sync"
	"syscall"

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

const (
	NEXT_CONNECTION_TYPE_UNKNOWN  = 0
	NEXT_CONNECTION_TYPE_WIRED    = 1
	NEXT_CONNECTION_TYPE_WIFI     = 2
	NEXT_CONNECTION_TYPE_CELLULAR = 3
	NEXT_CONNECTION_TYPE_MAX      = 3
)

const (
	NEXT_PLATFORM_UNKNOWN       = 0
	NEXT_PLATFORM_WINDOWS       = 1
	NEXT_PLATFORM_MAC           = 2
	NEXT_PLATFORM_UNIX          = 3
	NEXT_PLATFORM_SWITCH        = 4
	NEXT_PLATFORM_PS4           = 5
	NEXT_PLATFORM_IOS           = 6
	NEXT_PLATFORM_XBOX_ONE      = 7
	NEXT_PLATFORM_XBOX_SERIES_X = 8
	NEXT_PLATFORM_PS5           = 9
	NEXT_PLATFORM_MAX           = 9
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

	// TODO: Create bigquery client and start write loop
	// TODO: create metrics handler

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
					packet.CustomerId,
					packet.DatacenterId,
					packet.UserHash,
					packet.AddressHash,
					packet.SessionId,
					packet.PlatformId,
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
