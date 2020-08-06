/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/transport"
	"github.com/networknext/backend/transport/pubsub"
	"github.com/pebbe/zmq4"
)

// Shared
const (
	LoadTestDuration = time.Minute * 5 // How long to run the load test
)

// in memory map load test
const (
	SessionNextSwitchFrequency = time.Second * 30 // How often a session will randomly switch between direct and next
	SessionNextChance          = 0.25             // How likely the session is to pick next
	SessionLengthMin           = time.Minute      // The minimum playtime for the session
	SessionLengthMax           = time.Minute * 5  // The maximum playtime for the session
	NumServers                 = 250000
	NumSessions                = 500000
)

// zeromq load test
const (
	ZeroMQPublishDelay = 15000 // How long to wait before sending another message (in loop cycles). This number is CPU dependent.
)

func in_memory_map_load_test() {

	fmt.Printf("in_memory_map_load_test\n")

	runTime := time.Now()

	vetoMap := transport.NewVetoMap()
	serverMap := transport.NewServerMap()
	sessionMap := transport.NewSessionMap()

	ctx := context.Background()
	{
		go func() {
			timeout := int64(60 * 5)
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
	for i := 0; i < NumServers; i++ {
		go func(serverId int) {
			time.Sleep(time.Duration(float64(time.Second) * 10.0 * rand.Float64()))
			serverAddress := fmt.Sprintf("%x", serverId)
			buyerId := uint64(0)
			for {
				if time.Since(runTime) >= LoadTestDuration {
					break
				}

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
				averageServerDuration += (duration - averageServerDuration) * 0.01
				if duration > maxServerDuration {
					maxServerDuration = duration
				}
				serverDurationMutex.Unlock()
				time.Sleep(time.Second * 10)
			}
		}(i)
	}

	maxSessionDuration := 0.0
	averageSessionDuration := 0.0
	var sessionDurationMutex sync.Mutex
	for i := 0; i < NumSessions; i++ {
		go func(sessionId uint64) {
			time.Sleep(time.Duration(float64(time.Second) * 10.0 * rand.Float64()))

			sessionStartTime := time.Now()

			sessionLength := rand.Float64()

			sessionTimeout := time.Duration(float64(SessionLengthMin) + sessionLength*float64(SessionLengthMax-SessionLengthMin))

			buyerID := rand.Uint64()

			var nextSwitchCount uint64
			var nextSliceCounter uint64
			for {
				if time.Since(sessionStartTime) >= sessionTimeout || time.Since(runTime) >= LoadTestDuration {
					break
				}

				start := time.Now()
				vetoMap.Lock(sessionId)
				vetoReason := vetoMap.GetVeto(sessionId)
				sessionMap.Lock(sessionId)
				sessionDataReadOnly := sessionMap.GetSessionData(sessionId)
				if sessionDataReadOnly == nil {
					sessionDataReadOnly = transport.NewSessionData()
					// fmt.Printf("new session %05x (%d/%d)\n", sessionId, sessionMap.GetSessionCount()+1, numSessions)
				}

				if time.Since(sessionStartTime) >= SessionNextSwitchFrequency*time.Duration(nextSwitchCount) {
					if rand.Float32() < SessionNextChance {
						nextSliceCounter = 1
					} else {
						nextSliceCounter = 0
					}
					nextSwitchCount++
				}

				session := transport.SessionData{
					Timestamp:            time.Now().Unix(),
					BuyerID:              buyerID,
					Location:             sessionDataReadOnly.Location,
					Sequence:             sessionDataReadOnly.Sequence + 1,
					NearRelays:           sessionDataReadOnly.NearRelays,
					RouteHash:            0,
					Initial:              sessionDataReadOnly.Initial,
					RouteDecision:        sessionDataReadOnly.RouteDecision,
					NextSliceCounter:     nextSliceCounter,
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
				averageSessionDuration += (duration - averageSessionDuration) * 0.01
				if duration > maxSessionDuration {
					maxSessionDuration = duration
				}
				sessionDurationMutex.Unlock()
				time.Sleep(time.Second * 10)
			}
		}(uint64(i))
	}

	for {
		if time.Since(runTime) >= LoadTestDuration {
			break
		}

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
		fmt.Printf("total session count = %d sessions\n", sessionMap.GetSessionCount())
		fmt.Printf("direct session count = %d sessions\n", sessionMap.GetDirectSessionCount())
		fmt.Printf("next session count = %d sessions\n", sessionMap.GetNextSessionCount())
		fmt.Printf("========================================================\n")
		time.Sleep(time.Second)
	}

	// todo: need to count number of timeouts on maps etc. if timeouts occur then the load test fails
}

func zeromq_load_test() {
	fmt.Printf("zeromq_load_test\n")

	runTime := time.Now()

	recvstderr := make([]string, 0)

	subscriber, err := pubsub.NewPortalCruncherSubscriber("40000", 1000000)
	if err != nil {
		fmt.Printf("couldn't connect subscriber over zeromq socket: %v\n", err)
		return
	}

	publisher, err := pubsub.NewPortalCruncherPublisher("tcp://127.0.0.1:40000", 1000000)
	if err != nil {
		fmt.Printf("couldn't connect publisher over zeromq socket: %v\n", err)
		return
	}

	mockSessionCountData := transport.SessionCountData{
		InstanceID:             0,
		TotalNumDirectSessions: 50000,
		TotalNumNextSessions:   1000,
		NumDirectSessionsPerBuyer: map[uint64]uint64{
			0: 10000,
			1: 10000,
			2: 10000,
			3: 10000,
			4: 10000,
		},
		NumNextSessionsPerBuyer: map[uint64]uint64{
			0: 200,
			1: 200,
			2: 200,
			3: 200,
			4: 200,
		},
	}

	nearRelays := make([]transport.NearRelayPortalData, 0)
	for i := 0; i < transport.MaxNearRelays; i++ {
		nearRelays = append(nearRelays, transport.NearRelayPortalData{
			ID:   uint64(i),
			Name: "relay" + fmt.Sprintf("%d", i),
		})
	}

	mockSessionData := transport.SessionPortalData{
		Meta: transport.SessionMeta{
			ID:              0,
			UserHash:        0,
			DatacenterName:  "local",
			DatacenterAlias: "local",
			OnNetworkNext:   true,
			NextRTT:         30,
			DirectRTT:       50,
			DeltaRTT:        20,
			Location:        routing.LocationNullIsland,
			ClientAddr:      "127.0.0.1:40000",
			ServerAddr:      "127.0.0.1:40000",
			Hops: []transport.RelayHop{
				{
					ID:   1,
					Name: "relay1",
				},
				{
					ID:   2,
					Name: "relay2",
				},
				{
					ID:   3,
					Name: "relay3",
				},
				{
					ID:   4,
					Name: "relay4",
				},
			},
			SDK:          "3.3.3",
			Connection:   0,
			NearbyRelays: nearRelays,
			Platform:     0,
			BuyerID:      0,
		},
		Slice: transport.SessionSlice{
			Timestamp: time.Now(),
			Next:      routing.Stats{},
			Direct:    routing.Stats{},
			Envelope: routing.Envelope{
				Up:   int64(0),
				Down: int64(0),
			},
			IsMultiPath:       true,
			IsTryBeforeYouBuy: true,
			OnNetworkNext:     true,
		},
		Point: transport.SessionMapPoint{
			Latitude:      0,
			Longitude:     0,
			OnNetworkNext: true,
		},
	}

	subscriber.Subscribe(pubsub.TopicPortalCruncherSessionData)
	if err != nil {
		fmt.Printf("couldn't subscribe to session data topic: %v\n", err)
		return
	}

	subscriber.Subscribe(pubsub.TopicPortalCruncherSessionCounts)
	if err != nil {
		fmt.Printf("couldn't subscribe to session counts topic: %v\n", err)
		return
	}

	go func() {
		var index uint64
		for {
			topic, message, err := subscriber.ReceiveMessage()
			if err != nil {
				recvstderr = append(recvstderr, fmt.Sprintf("error receiving message: %v\n", err))
			}

			switch topic {
			case pubsub.TopicPortalCruncherSessionData:
				var sessionData transport.SessionPortalData
				if err := sessionData.UnmarshalBinary(message); err != nil {
					fmt.Printf("error unmarshaling portal data: %v\n", err)
					os.Exit(1)
				}

				if sessionData.Meta.ID != index {
					recvstderr = append(recvstderr, fmt.Sprintf("portal data received out of order or missing, messageID=%d index=%d\n", sessionData.Meta.ID, index))
				}
			case pubsub.TopicPortalCruncherSessionCounts:
				var sessionCounts transport.SessionCountData
				if err := sessionCounts.UnmarshalBinary(message); err != nil {
					fmt.Printf("error unmarshaling count data: %v\n", err)
					os.Exit(1)
				}

				if sessionCounts.InstanceID != index {
					recvstderr = append(recvstderr, fmt.Sprintf("count data received out of order or missing, messageID=%d index=%d\n", sessionCounts.InstanceID, index))
				}
			}

			index++

			if len(recvstderr) > 200 {
				for _, line := range recvstderr {
					fmt.Printf(line)
				}
				os.Exit(1)
			}
		}
	}()

	// Wait a small amount of time before publishing data so that we know
	// the subscriber is ready
	time.Sleep(time.Second * 2)

	fmt.Println("Starting publish routine")

	var publishIndex uint64
	go func() {
		for {
			mockSessionData.Meta.ID = publishIndex
			sessionBytes, err := mockSessionData.MarshalBinary()
			if err != nil {
				fmt.Printf("couldn't marshal session data: %v\n", err)
				return
			}

			retry := true
			errorTime := time.Since(runTime)
			for retry {
				_, err = publisher.Publish(pubsub.TopicPortalCruncherSessionData, sessionBytes)
				if err != nil {
					fmt.Printf("error publishing session data: %v\n", err)

					errno := zmq4.AsErrno(err)
					switch errno {
					case zmq4.AsErrno(syscall.EAGAIN):
						fmt.Printf("retrying index %d\n", publishIndex)
						sendRate := float64(publishIndex) / errorTime.Seconds()
						fmt.Printf("upper bound average send rate: %.0f msg/sec\n", sendRate)
						time.Sleep(time.Millisecond * 100)
					default:
						fmt.Println(err)
						os.Exit(1)
					}
				} else {
					retry = false
				}
			}

			publishIndex++

			mockSessionCountData.InstanceID = publishIndex
			countBytes, err := mockSessionCountData.MarshalBinary()
			if err != nil {
				fmt.Printf("couldn't marshal session counts: %v\n", err)
				return
			}

			retry = true
			errorTime = time.Since(runTime)
			for retry {
				_, err = publisher.Publish(pubsub.TopicPortalCruncherSessionCounts, countBytes)
				if err != nil {
					fmt.Printf("error publishing session counts: %v\n", err)
					fmt.Println(publishIndex)

					errno := zmq4.AsErrno(err)
					switch errno {
					case zmq4.AsErrno(syscall.EAGAIN):
						fmt.Printf("retrying index %d\n", publishIndex)
						sendRate := float64(publishIndex) / errorTime.Seconds()
						fmt.Printf("upper bound average send rate: %.0f msg/sec\n", sendRate)
						time.Sleep(time.Millisecond * 100)
					default:
						fmt.Println(err)
						os.Exit(1)
					}
				} else {
					retry = false
				}
			}

			publishIndex++

			// We can't avoid doing some sort of waiting here, otherwise we'll flood ZeroMQ's internal send buffer.
			// However using time.Sleep actually lowers the message send rate significantly, so just use a useless loop.
			// This will prove that we won't flood between session updates (because session updates aren't instantaneous).
			var unused int
			for i := 0; i < ZeroMQPublishDelay; i++ {
				unused++
			}
		}
	}()

	go func() {
		for {
			if time.Since(runTime) >= LoadTestDuration {
				doneTime := time.Since(runTime)
				sendRate := float64(publishIndex) / doneTime.Seconds()
				fmt.Printf("average send rate: %.0f msg/sec\n", sendRate)
				os.Exit(0)
			}

			time.Sleep(time.Second)
		}
	}()

	// Wait for interrupt signal
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint

	doneTime := time.Since(runTime)
	sendRate := float64(publishIndex) / doneTime.Seconds()
	fmt.Printf("\naverage send rate: %.0f msg/sec\n", sendRate)
}

func main() {
	// in_memory_map_load_test()
	zeromq_load_test()
}
