package main

import (
	"fmt"
	"math/rand"
	"time"

	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/portal"
	"github.com/networknext/backend/modules/common"

	"github.com/gomodule/redigo/redis"
)

func RunSessionCrunchThreads(pool *redis.Pool, threadCount int) {

	fmt.Printf("RunSessionCrunchThreads\n")

	for k := 0; k < threadCount; k++ {

		go func(thread int) {

			time.Sleep(time.Duration(rand.Intn(10000)) * time.Millisecond)

			iteration := uint64(0)

			near_relay_max := uint64(0)

			for {

				redisClient := pool.Get()

				start := time.Now()
				secs := start.Unix()
				minutes := secs / 60

				all_sessions := redis.Args{}.Add(fmt.Sprintf("s-%d", minutes))
				next_sessions := redis.Args{}.Add(fmt.Sprintf("n-%d", minutes))

				for j := 0; j < 1000; j++ {

					sessionId := uint64(thread*1000000) + uint64(j) + iteration

					score := rand.Intn(10000)

					all_sessions = all_sessions.Add(score)
					all_sessions = all_sessions.Add(fmt.Sprintf("%016x", sessionId))

					next := ((uint64(j) + iteration) % 10) == 0
					if next {
						next_sessions = next_sessions.Add(score)
						next_sessions = next_sessions.Add(fmt.Sprintf("%016x", sessionId))
					}

					sessionData := portal.GenerateRandomSessionData()
					redisClient.Send("SET", fmt.Sprintf("sd-%016x", sessionId), sessionData.Value())
					redisClient.Send("EXPIRE", fmt.Sprintf("sd-%016x 30", sessionId))

					mapData := portal.MapData{}
					mapData.Latitude = float32(common.RandomInt(-90000, +90000)) / 1000.0
					mapData.Longitude = float32(common.RandomInt(-18000, +18000)) / 1000.0
					mapData.Next = next
					redisClient.Send("SET", fmt.Sprintf("m-%016x", sessionId), mapData.Value())
					redisClient.Send("EXPIRE", fmt.Sprintf("m-%016x 30", sessionId))

					sliceData := portal.GenerateRandomSliceData()
					redisClient.Send("RPUSH", fmt.Sprintf("sl-%016x", sessionId), sliceData.Value())
					redisClient.Send("EXPIRE", fmt.Sprintf("sl-%016x 30", sessionId))

					if sessionId > near_relay_max {
						nearRelayData := portal.GenerateRandomNearRelayData()
						redisClient.Send("RPUSH", fmt.Sprintf("nr-%016x"), nearRelayData.Value())
						redisClient.Send("EXPIRE", fmt.Sprintf("nr-%016x 30", sessionId))
						near_relay_max = sessionId
					}
				}

				if len(all_sessions) > 1 {
					redisClient.Send("ZADD", all_sessions...)
					redisClient.Send("EXPIRE", fmt.Sprintf("s-%d", minutes), 30)
				}

				if len(next_sessions) > 1 {
					redisClient.Send("ZADD", next_sessions...)
					redisClient.Send("EXPIRE", fmt.Sprintf("n-%d", minutes), 30)
				}

				redisClient.Flush()

				redisClient.Close()

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
				} else {
					// todo
					fmt.Printf("nil session data?!\n")
				}
			}

			start = time.Now()

			mapData, err := portal.GetMapData(pool, minutes)
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

	redisPool := portal.CreateRedisPool(redisHostname)

	threadCount := envvar.GetInt("REDIS_THREAD_COUNT", 100)

	RunSessionCrunchThreads(redisPool, threadCount)

	RunPollThread(redisPool)

	time.Sleep(time.Minute)
}
