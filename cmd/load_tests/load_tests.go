/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"fmt"
	"context"
	"time"
	"sync"
	"math/rand"

	"github.com/networknext/backend/transport"
)

func load_test() {

	fmt.Printf("load_test\n")

	numServers := 250000
	numSessions := 500000

	vetoMap := transport.NewVetoMap()
	serverMap := transport.NewServerMap()
	sessionMap := transport.NewSessionMap()

	ctx := context.Background()
	{
		go func() {
			timeout := int64(60*5)
			frequency := time.Millisecond * 100
			ticker := time.NewTicker(frequency)
			vetoMap.TimeoutLoop(ctx, timeout, ticker.C)
		}()

		go func() {
			timeout := int64(30)
			frequency := time.Millisecond * 100
			ticker := time.NewTicker(frequency)
			serverMap.TimeoutLoop(ctx, timeout, ticker.C)
		}()

		go func() {
			timeout := int64(30)
			frequency := time.Millisecond * 100
			ticker := time.NewTicker(frequency)
			sessionMap.TimeoutLoop(ctx, timeout, ticker.C)
		}()
	}

	maxServerDuration := 0.0
	averageServerDuration := 0.0
	var serverDurationMutex sync.Mutex
	for i := 0; i < numServers; i++ {
		go func(serverId int) {
			time.Sleep(time.Duration(float64(time.Second)*10.0*rand.Float64()))
			serverAddress := fmt.Sprintf("%x", serverId)
			buyerId := uint64(0)
			for {
				start := time.Now()
				serverMap.Lock(buyerId, serverAddress)
				serverDataReadOnly := serverMap.GetServerData(buyerId, serverAddress)
				if serverDataReadOnly == nil {
					serverDataReadOnly = &transport.ServerData{}
					// fmt.Printf("new server %05x (%d/%d)\n", serverId, serverMap.GetServerCount()+1, numServers)
				}
				serverCopy := *serverDataReadOnly
				serverCopy.Timestamp = time.Now().Unix()
				serverMap.UpdateServerData(buyerId, serverAddress, &serverCopy)
				serverMap.Unlock(buyerId, serverAddress)
				duration := time.Since(start).Seconds()
				serverDurationMutex.Lock()
				averageServerDuration += ( duration - averageServerDuration ) * 0.01
				if duration > maxServerDuration {
					maxServerDuration = duration
				}
				serverDurationMutex.Unlock()
				time.Sleep(time.Second*10)
			}
		}(i)
	}

	maxSessionDuration := 0.0
	averageSessionDuration := 0.0
	var sessionDurationMutex sync.Mutex
	for i := 0; i < numSessions; i++ {
		go func(sessionId uint64) {
			time.Sleep(time.Duration(float64(time.Second)*10.0*rand.Float64()))
			for {
				start := time.Now()
				vetoMap.Lock(sessionId)
				vetoReason := vetoMap.GetVeto(sessionId)
				sessionMap.Lock(sessionId)
				sessionDataReadOnly := sessionMap.GetSessionData(sessionId)
				if sessionDataReadOnly == nil {
					sessionDataReadOnly = transport.NewSessionData()
					// fmt.Printf("new session %05x (%d/%d)\n", sessionId, sessionMap.GetSessionCount()+1, numSessions)
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
				vetoMap.SetVeto(sessionId, vetoReason)
				sessionMap.Unlock(sessionId)
				vetoMap.Unlock(sessionId)
				duration := time.Since(start).Seconds()
				sessionDurationMutex.Lock()
				averageSessionDuration += ( duration - averageSessionDuration ) * 0.01
				if duration > maxSessionDuration {
					maxSessionDuration = duration
				}
				sessionDurationMutex.Unlock()
				time.Sleep(time.Second*10)
			}
		}(uint64(i))
	}

	for {
		fmt.Printf("========================================================\n")
		serverDurationMutex.Lock()
		serverDuration_max := maxServerDuration
		serverDuration_avg := averageServerDuration
		serverDurationMutex.Unlock()
		sessionDurationMutex.Lock()
		sessionDuration_max := maxSessionDuration
		sessionDuration_avg := averageSessionDuration
		sessionDurationMutex.Unlock()
		fmt.Printf("avg server duration = %f seconds\n", serverDuration_avg)
		fmt.Printf("max server duration = %f seconds\n", serverDuration_max)
		fmt.Printf("avg session duration = %f seconds\n", sessionDuration_avg)
		fmt.Printf("max session duration = %f seconds\n", sessionDuration_max)
		fmt.Printf("========================================================\n")
		time.Sleep(time.Second)
	}

	// todo: need to count number of timeouts on maps etc. if timeouts occur then the load test fails
}

func main() {
	load_test()
}
