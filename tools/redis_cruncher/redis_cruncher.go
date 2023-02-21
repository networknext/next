package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/portal"

	"github.com/gomodule/redigo/redis"
)

func RunSessionInsertThreads(pool *redis.Pool, threadCount int) {

	for k := 0; k < threadCount; k++ {

		go func(thread int) {

			sessionInserter := portal.CreateSessionInserter(pool, 1000)

			nearRelayInserter := portal.CreateNearRelayInserter(pool, 1000)

			iteration := uint64(0)

			time.Sleep(time.Duration(rand.Intn(10000)) * time.Millisecond)

			near_relay_max := uint64(0)

			for {

				for j := 0; j < 1000; j++ {

					sessionId := uint64(thread*1000000) + uint64(j) + iteration
					score := uint32(rand.Intn(10000))
					next := ((uint64(j) + iteration) % 10) == 0

					sessionData := portal.GenerateRandomSessionData()

					sliceData := portal.GenerateRandomSliceData()

					sessionInserter.Insert(sessionId, score, next, sessionData, sliceData)

					if sessionId > near_relay_max {
						nearRelayData := portal.GenerateRandomNearRelayData()
						nearRelayInserter.Insert(sessionId, nearRelayData)
						near_relay_max = sessionId
					}
				}

				time.Sleep(10 * time.Second)

				iteration++
			}
		}(k)
	}
}

func RunServerInsertThreads(pool *redis.Pool, threadCount int) {

	for k := 0; k < threadCount; k++ {

		go func(thread int) {

			serverInserter := portal.CreateServerInserter(pool, 1000)

			iteration := uint64(0)

			time.Sleep(time.Duration(rand.Intn(10000)) * time.Millisecond)

			for {

				for j := 0; j < 1000; j++ {

					serverData := portal.GenerateRandomServerData()

					id := uint32(iteration + uint64(j))

					serverData.ServerAddress = core.ParseAddress(fmt.Sprintf("%d.%d.%d.%d:%d", id&0xFF, (id>>8)&0xFF, (id>>16)&0xFF, (id>>24)&0xFF, uint64(thread)))

					score := uint32(rand.Intn(10000))

					serverInserter.Insert(score, serverData)
				}

				time.Sleep(10 * time.Second)

				iteration++
			}
		}(k)
	}
}

func RunRelayInsertThreads(pool *redis.Pool, threadCount int) {

	for k := 0; k < threadCount; k++ {

		go func(thread int) {

			relayInserter := portal.CreateRelayInserter(pool, 1000)

			iteration := uint64(0)

			time.Sleep(time.Duration(rand.Intn(10000)) * time.Millisecond)

			for {

				for j := 0; j < 1000; j++ {

					relayData := portal.GenerateRandomRelayData()

					id := uint32(iteration + uint64(j))

					relayData.RelayAddress = core.ParseAddress(fmt.Sprintf("%d.%d.%d.%d:%d", id&0xFF, (id>>8)&0xFF, (id>>16)&0xFF, (id>>24)&0xFF, uint64(thread)))

					score := uint32(rand.Intn(10000))

					relayInserter.Insert(score, relayData)
				}

				time.Sleep(10 * time.Second)

				iteration++
			}
		}(k)
	}
}

func RunPollThread(pool *redis.Pool) {

	go func() {

		fmt.Printf("\n")

		for {

			fmt.Printf("-------------------------------------------------\n")

			start := time.Now()
			secs := start.Unix()
			minutes := secs / 60

			begin := 0
			end := 1000

			sessions, totalSessionCount, nextSessionCount := portal.GetSessions(pool, minutes, begin, end)

			fmt.Printf("sessions: %d of %d/%d (%.1fms)\n", len(sessions), nextSessionCount, totalSessionCount, float64(time.Since(start).Milliseconds()))

			start = time.Now()

			if len(sessions) > 0 {
				start = time.Now()
				sessionData, sliceData, nearRelayData := portal.GetSessionData(pool, sessions[0].SessionId)
				if sessionData != nil {
					fmt.Printf("session data: %x, %d slices, %d near relay data (%.1fms)\n", sessionData.SessionId, len(sliceData), len(nearRelayData), float64(time.Since(start).Milliseconds()))
				}
			}

			start = time.Now()

			mapData, err := portal.GetMapData(pool, start)
			if err != nil {
				panic(fmt.Sprintf("failed to get map data: %v", err))
			}

			fmt.Printf("map data: %d points (%.1fms)\n", len(mapData), float64(time.Since(start).Milliseconds()))

			start = time.Now()

			servers, totalServerCount := portal.GetServers(pool, minutes, begin, end)

			fmt.Printf("servers: %d of %d (%.1fms)\n", len(servers), totalServerCount, float64(time.Since(start).Milliseconds()))

			start = time.Now()

			relays, totalRelayCount := portal.GetRelays(pool, minutes, begin, end)

			fmt.Printf("relays: %d of %d (%.1fms)\n", len(relays), totalRelayCount, float64(time.Since(start).Milliseconds()))

			fmt.Printf("-------------------------------------------------\n")

			time.Sleep(time.Second)
		}
	}()
}

func main() {

	redisHostname := envvar.GetString("REDIS_HOSTNAME", "127.0.0.1:6379")

	redisPool := portal.CreateRedisPool(redisHostname, 1000)

	threadCount := envvar.GetInt("REDIS_THREAD_COUNT", 100)

	RunSessionInsertThreads(redisPool, threadCount)
	RunServerInsertThreads(redisPool, threadCount)
	RunRelayInsertThreads(redisPool, threadCount)

	RunPollThread(redisPool)

	time.Sleep(time.Hour)
}
