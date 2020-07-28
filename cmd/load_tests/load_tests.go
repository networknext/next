/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"fmt"
	"os"
	"runtime"
	"reflect"
	"context"
	"time"
	"math/rand"

	"github.com/networknext/backend/transport"
)

func test_simulator() {

	fmt.Printf("test_simulator\n")

	numServers := 100000
	numSessions := 500000

	vetoMap := transport.NewVetoMap()
	serverMap := transport.NewServerMap()
	sessionMap := transport.NewSessionMap()

	ctx := context.Background()
	{
		go func() {
			timeout := int64(10)
			frequency := time.Millisecond * 100
			ticker := time.NewTicker(frequency)
			vetoMap.TimeoutLoop(ctx, timeout, ticker.C)
		}()

		go func() {
			timeout := int64(10)
			frequency := time.Millisecond * 100
			ticker := time.NewTicker(frequency)
			serverMap.TimeoutLoop(ctx, timeout, ticker.C)
		}()

		go func() {
			timeout := int64(10)
			frequency := time.Millisecond * 100
			ticker := time.NewTicker(frequency)
			sessionMap.TimeoutLoop(ctx, timeout, ticker.C)
		}()
	}

	for i := 0; i < numServers; i++ {
		go func(serverId int) {
			time.Sleep(time.Duration(float64(time.Second)*10.0*rand.Float64()))
			serverAddress := fmt.Sprintf("%x", serverId)
			buyerId := uint64(0)
			for {
				serverMap.Lock(buyerId, serverAddress)
				serverDataReadOnly := serverMap.GetServerData(buyerId, serverAddress)
				if serverDataReadOnly == nil {
					serverDataReadOnly = &transport.ServerData{}
					fmt.Printf("new server %05x (%d/%d)\n", serverId, serverMap.GetServerCount()+1, numServers)
				}
				serverCopy := *serverDataReadOnly
				serverCopy.Timestamp = time.Now().Unix()
				serverMap.UpdateServerData(buyerId, serverAddress, &serverCopy)
				serverMap.Unlock(buyerId, serverAddress)
				time.Sleep(time.Second*10)
			}
		}(i)
	}

	for i := 0; i < numSessions; i++ {
		go func(sessionId uint64) {
			time.Sleep(time.Duration(float64(time.Second)*10.0*rand.Float64()))
			for {
				sessionMap.Lock(sessionId)
				sessionDataReadOnly := sessionMap.GetSessionData(sessionId)
				if sessionDataReadOnly == nil {
					sessionDataReadOnly = transport.NewSessionData()
					fmt.Printf("new session %05x (%d/%d)\n", sessionId, sessionMap.GetSessionCount()+1, numSessions)
				}
				session := transport.SessionData{
					Timestamp:            time.Now().Unix(),
					Location:             sessionDataReadOnly.Location,
					Sequence:             sessionDataReadOnly.Sequence + 1,
					NearRelays:           sessionDataReadOnly.NearRelays,
					RouteHash:            0,
					Initial:              sessionDataReadOnly.Initial,
					RouteDecision:        sessionDataReadOnly.RouteDecision,
					NextSliceCounter:     sessionDataReadOnly.NextSliceCounter,
					CommittedData:        sessionDataReadOnly.CommittedData,
					RouteExpireTimestamp: sessionDataReadOnly.RouteExpireTimestamp,
					TokenVersion:         sessionDataReadOnly.TokenVersion,
					CachedResponse:       nil,
					SliceMutexes:         sessionDataReadOnly.SliceMutexes,
				}
				sessionMap.UpdateSessionData(sessionId, &session)
				sessionMap.Unlock(sessionId)
				time.Sleep(time.Second*10)
			}
		}(uint64(i))
	}

	for {
		time.Sleep(time.Second)
	}
}

func test_server_map() {

	fmt.Printf("test_server_map\n")

	numThreads := 100000

	serverMap := transport.NewServerMap()

	ctx := context.Background()
	{
		go func() {
			timeout := int64(10)
			frequency := time.Millisecond * 100
			ticker := time.NewTicker(frequency)
			serverMap.TimeoutLoop(ctx, timeout, ticker.C)
		}()
	}

	buyerId := uint64(0)

	for i := 0; i < numThreads; i++ {

		go func() {

			serverId := uint64(1000*i)

			for {

				serverId ++
				serverId := ( serverId % 100001 )

				serverAddress := fmt.Sprintf("%x", serverId)

				serverMap.Lock(buyerId, serverAddress)

				serverDataReadOnly := serverMap.GetServerData(buyerId, serverAddress)
				if serverDataReadOnly == nil {
					serverDataReadOnly = &transport.ServerData{}
					fmt.Printf("new server %x (%d/100000)\n", serverId, serverMap.GetServerCount())
				}

				serverCopy := *serverDataReadOnly
				serverCopy.Timestamp = time.Now().Unix()

				serverMap.UpdateServerData(buyerId, serverAddress, &serverCopy)

				serverMap.Unlock(buyerId, serverAddress)
			}

		}()

	}

	for {
		time.Sleep(time.Second)
	}
}

func test_session_map() {

	fmt.Printf("test_session_map\n")

	numThreads := 100000

	sessionMap := transport.NewSessionMap()

	ctx := context.Background()
	{
		go func() {
			timeout := int64(10)
			frequency := time.Millisecond * 100
			ticker := time.NewTicker(frequency)
			sessionMap.TimeoutLoop(ctx, timeout, ticker.C)
		}()
	}

	for i := 0; i < numThreads; i++ {

		go func() {

			sessionId := uint64(1000*i)

			for {

				sessionId ++
				sessionId := ( sessionId % 100001 )

				sessionMap.Lock(sessionId)

				sessionDataReadOnly := sessionMap.GetSessionData(sessionId)
				if sessionDataReadOnly == nil {
					sessionDataReadOnly = transport.NewSessionData()
					fmt.Printf("new session %x (%d/100000)\n", sessionId, sessionMap.GetSessionCount())
				}

				session := transport.SessionData{
					Timestamp:            time.Now().Unix(),
					Location:             sessionDataReadOnly.Location,
					Sequence:             sessionDataReadOnly.Sequence + 1,
					NearRelays:           sessionDataReadOnly.NearRelays,
					RouteHash:            0,
					Initial:              sessionDataReadOnly.Initial,
					RouteDecision:        sessionDataReadOnly.RouteDecision,
					NextSliceCounter:     sessionDataReadOnly.NextSliceCounter,
					CommittedData:        sessionDataReadOnly.CommittedData,
					RouteExpireTimestamp: sessionDataReadOnly.RouteExpireTimestamp,
					TokenVersion:         sessionDataReadOnly.TokenVersion,
					CachedResponse:       nil,
					SliceMutexes:         sessionDataReadOnly.SliceMutexes,
				}

				// ...

				sessionMap.UpdateSessionData(sessionId, &session)

				sessionMap.Unlock(sessionId)
			}

		}()

	}

	for {
		time.Sleep(time.Second)
	}
}

func test_veto_map() {

	fmt.Printf("test_veto_map\n")

	numThreads := 100000

	vetoMap := transport.NewVetoMap()

	ctx := context.Background()
	{
		go func() {
			timeout := int64(10)
			frequency := time.Millisecond * 100
			ticker := time.NewTicker(frequency)
			vetoMap.TimeoutLoop(ctx, timeout, ticker.C)
		}()
	}

	for i := 0; i < numThreads; i++ {

		go func() {

			sessionId := uint64(1000*i)

			for {

				sessionId ++
				sessionId := ( sessionId % 100001 )

				vetoMap.Lock(sessionId)

				vetoReason := vetoMap.GetVeto(sessionId)

				if vetoReason == 0 {
					fmt.Printf("new veto %x (%d/100000)\n", sessionId, vetoMap.GetVetoCount())
					vetoReason = 1
				}

				vetoMap.SetVeto(sessionId, vetoReason)
				
				vetoMap.Unlock(sessionId)
			}

		}()

	}

	for {
		time.Sleep(time.Second)
	}
}

type test_function func()

func main() {
	allTests := []test_function{
		test_simulator,
		// test_server_map,
		// test_session_map,
		// test_veto_map,
	}

	// If there are command line arguments, use reflection to see what tests to run
	var tests []test_function
	prefix := "main."
	if len(os.Args) > 1 {
		for _, funcName := range os.Args[1:] {
			for _, test := range allTests {
				name := runtime.FuncForPC(reflect.ValueOf(test).Pointer()).Name()
				name = name[len(prefix):]
				if funcName == name {
					tests = append(tests, test)
				}
			}
		}
	} else {
		tests = allTests // No command line args, run all tests
	}

	for i := range tests {
		tests[i]()
	}
}
