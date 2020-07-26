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

	"github.com/networknext/backend/transport"
)

func test_server_map() {

	fmt.Printf("test_server_map\n")

	numThreads := 100000

	serverMap := transport.NewServerMap()

	ctx := context.Background()
	{
		go func() {
			timeout := int64(60)
			frequency := time.Millisecond * 100
			ticker := time.NewTicker(frequency)
			serverMap.TimeoutLoop(ctx, timeout, ticker.C)
		}()
	}

	// todo
	/*
	for i := 0; i < numThreads; i++ {

		go func() {

			sessionId := uint64(1000*i)

			for {

				sessionId ++
				sessionId := ( sessionId % 100000 )

				sessionMap.Lock(sessionId)

				sessionDataReadOnly := sessionMap.GetSessionData(sessionId)
				if sessionDataReadOnly == nil {
					sessionDataReadOnly = transport.NewSessionData()
					fmt.Printf("new session %x\n", sessionId)
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
	*/

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
			timeout := int64(60)
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
				sessionId := ( sessionId % 100000 )

				sessionMap.Lock(sessionId)

				sessionDataReadOnly := sessionMap.GetSessionData(sessionId)
				if sessionDataReadOnly == nil {
					sessionDataReadOnly = transport.NewSessionData()
					fmt.Printf("new session %x\n", sessionId)
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

type test_function func()

func main() {
	allTests := []test_function{
		test_server_map,
		// test_session_map,
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
