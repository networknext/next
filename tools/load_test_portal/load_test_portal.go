package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/envvar"
	"github.com/networknext/next/modules/portal"
)

var RedisNodes = []string{"127.0.0.1:10000", "127.0.0.1:10001", "127.0.0.1:10002", "127.0.0.1:10003", "127.0.0.1:10004", "127.0.0.1:10005"}

var SessionCruncherURL = "http://127.0.0.1:40200/session_batch"

func RunSessionInsertThreads(ctx context.Context, threadCount int) {

	for k := 0; k < threadCount; k++ {

		go func(thread int) {

			redisClient := common.CreateRedisClusterClient(RedisNodes)

			sessionInserter := portal.CreateSessionInserter(ctx, redisClient, SessionCruncherURL, 1000)

			iteration := uint64(0)

			nearRelayInserter := portal.CreateNearRelayInserter(redisClient, 1000)

			near_relay_max := uint64(0)

			time.Sleep(time.Duration(rand.Intn(10000)) * time.Millisecond)

			for {

				for j := 0; j < 10000; j++ {

					sessionId := uint64(thread*1000000) + uint64(j) + iteration

					userHash := uint64(j) + iteration

					sessionData := portal.GenerateRandomSessionData()

					sessionData.SessionId = sessionId
					sessionData.UserHash = userHash

					sliceData := portal.GenerateRandomSliceData()

					next := common.RandomBool()

					currentScore := uint32(common.RandomInt(0,999))
					previousScore := uint32(common.RandomInt(0,999))

					sessionInserter.Insert(ctx, sessionId, userHash, next, currentScore, previousScore, sessionData, sliceData)

					if sessionId > near_relay_max {
						nearRelayData := portal.GenerateRandomNearRelayData()
						nearRelayInserter.Insert(ctx, sessionId, nearRelayData)
						near_relay_max = sessionId
					}
				}

				time.Sleep(10 * time.Second)

				iteration++
			}
		}(k)
	}
}

func RunServerInsertThreads(ctx context.Context, threadCount int) {

	for k := 0; k < threadCount; k++ {

		go func(thread int) {

			redisClient := common.CreateRedisClusterClient(RedisNodes)

			serverInserter := portal.CreateServerInserter(redisClient, 1000)

			iteration := uint64(0)

			time.Sleep(time.Duration(rand.Intn(10000)) * time.Millisecond)

			for {

				for j := 0; j < 1000; j++ {

					serverData := portal.GenerateRandomServerData()

					id := uint32(iteration + uint64(j))

					serverData.ServerAddress = fmt.Sprintf("%d.%d.%d.%d:%d", id&0xFF, (id>>8)&0xFF, (id>>16)&0xFF, (id>>24)&0xFF, uint64(thread))

					serverInserter.Insert(ctx, serverData)
				}

				time.Sleep(10 * time.Second)

				iteration++
			}
		}(k)
	}
}

func RunRelayInsertThreads(ctx context.Context, threadCount int) {

	for k := 0; k < threadCount; k++ {

		go func(thread int) {

			redisClient := common.CreateRedisClusterClient(RedisNodes)

			relayInserter := portal.CreateRelayInserter(redisClient, 1000)

			iteration := uint64(0)

			time.Sleep(time.Duration(rand.Intn(10000)) * time.Millisecond)

			for {

				for j := 0; j < 10; j++ {

					relayData := portal.GenerateRandomRelayData()

					id := uint32(iteration + uint64(j))

					relayData.RelayAddress = fmt.Sprintf("%d.%d.%d.%d:%d", id&0xFF, (id>>8)&0xFF, (id>>16)&0xFF, (id>>24)&0xFF, uint64(thread))

					relayInserter.Insert(ctx, relayData)
				}

				time.Sleep(10 * time.Second)

				iteration++
			}
		}(k)
	}
}

func RunPollThread(ctx context.Context) {

	iteration := uint64(0)

	go func() {

		redisClient := common.CreateRedisClusterClient(RedisNodes)

		for {

			// ------------------------------------------------------------------------------------------

			fmt.Printf("------------------------------------------------------------------------------------------------\n")

			start := time.Now()

			sessionIds := make([]uint64, 1000)
			for i := 0; i < len(sessionIds); i++ {
				sessionIds[i] = uint64(1000000) + uint64(i)
			}

			sessionList := portal.GetSessionList(ctx, redisClient, sessionIds)
			if sessionList != nil {
				fmt.Printf("session list %d (%.3fms)\n", len(sessionList), float64(time.Since(start).Milliseconds()))
			}

			// ------------------------------------------------------------------------------------------

			start = time.Now()

			minutes := start.Unix() / 60

			userHash := uint64(0x131e)

			userSessionList := portal.GetUserSessionList(ctx, redisClient, userHash, minutes, 0, 100)
			if userSessionList != nil {
				fmt.Printf("user session list %d (%.3fms)\n", len(userSessionList), float64(time.Since(start).Milliseconds()))
			}

			// ------------------------------------------------------------------------------------------

			start = time.Now()

			sessionId := uint64(1000000) + iteration

			sessionData, sliceData, nearRelayData := portal.GetSessionData(ctx, redisClient, sessionId)
			if sessionData != nil {
				fmt.Printf("session data %x -> %d slices, %d near relay data (%.3fms)\n", sessionData.SessionId, len(sliceData), len(nearRelayData), float64(time.Since(start).Milliseconds()))
			}

			// ------------------------------------------------------------------------------------------

			start = time.Now()

			serverCount := portal.GetServerCount(ctx, redisClient, minutes)

			fmt.Printf("servers: %d (%.3fms)\n", serverCount, float64(time.Since(start).Milliseconds()))

			// ------------------------------------------------------------------------------------------

			start = time.Now()

			serverAddresses := portal.GetServerAddresses(ctx, redisClient, minutes, 0, 100)

			fmt.Printf("server addresses -> %d server addresses (%.3fms)\n", len(serverAddresses), float64(time.Since(start).Milliseconds()))

			// ------------------------------------------------------------------------------------------

			if len(serverAddresses) > 0 {

				start := time.Now()

				serverAddress := serverAddresses[0]

				serverData, serverSessions := portal.GetServerData(ctx, redisClient, serverAddress, minutes)

				if serverData != nil {
					fmt.Printf("server data %s -> %d sessions (%.3fms)\n", serverData.ServerAddress, len(serverSessions), float64(time.Since(start).Milliseconds()))
				}				
			}

			// ------------------------------------------------------------------------------------------

			start = time.Now()

			serverList := portal.GetServerList(ctx, redisClient, serverAddresses)
			if serverList != nil {
				fmt.Printf("server list %d (%.3fms)\n", len(serverList), float64(time.Since(start).Milliseconds()))
			}

			// ------------------------------------------------------------------------------------------

			start = time.Now()

			relayCount := portal.GetRelayCount(ctx, redisClient, minutes)

			fmt.Printf("relays: %d (%.3fms)\n", relayCount, float64(time.Since(start).Milliseconds()))

			// ------------------------------------------------------------------------------------------

			start = time.Now()

			relayAddresses := portal.GetRelayAddresses(ctx, redisClient, minutes, 0, 100)

			fmt.Printf("relay addresses -> %d relay addresses (%.3fms)\n", len(relayAddresses), float64(time.Since(start).Milliseconds()))

			// ------------------------------------------------------------------------------------------

			if len(relayAddresses) > 0 {

				start := time.Now()

				relayAddress := relayAddresses[0]

				relayData := portal.GetRelayData(ctx, redisClient, relayAddress)

				if relayData != nil {
					fmt.Printf("relay data %s (%.3fms)\n", relayData.RelayAddress, float64(time.Since(start).Milliseconds()))
				}				
			}

			// ------------------------------------------------------------------------------------------

			start = time.Now()

			relayList := portal.GetRelayList(ctx, redisClient, relayAddresses)
			if serverList != nil {
				fmt.Printf("relay list %d (%.3fms)\n", len(relayList), float64(time.Since(start).Milliseconds()))
			}

			// ------------------------------------------------------------------------------------------

			time.Sleep(time.Second)

			iteration++
		}
	}()
}

func main() {

	threadCount := envvar.GetInt("REDIS_THREAD_COUNT", 100)

	ctx := context.Background()

	RunSessionInsertThreads(ctx, threadCount)

	RunServerInsertThreads(ctx, threadCount)

	RunRelayInsertThreads(ctx, threadCount)

	RunPollThread(ctx)

	time.Sleep(time.Minute)
}
