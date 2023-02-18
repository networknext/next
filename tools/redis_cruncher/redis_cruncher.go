package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"sort"
	"strconv"
	"time"

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/portal"
	"github.com/networknext/backend/modules/envvar"

	"github.com/gomodule/redigo/redis"
)

func createRedisPool(hostname string) *redis.Pool {
	pool := redis.Pool{
		MaxIdle:     1000,
		MaxActive:   64,
		IdleTimeout: 60 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", hostname)
		},
	}
	redisClient := pool.Get()
	redisClient.Send("PING")
	redisClient.Send("FLUSHDB")
	redisClient.Flush()
	pong, err := redisClient.Receive()
	if err != nil || pong != "PONG" {
		panic(err)
	}
	redisClient.Close()
	return &pool
}

func createRedisClient(hostname string) net.Conn {
	client, err := net.Dial("tcp", hostname)
	if err != nil {
		panic(err)
	}
	go func() {
		// todo: this is for insert only clients but we need to detect a closed connection and recreate it
		// todo: we should also ping here to make sure we're OK
		reader := bufio.NewReader(client)
		for {
			message, _ := reader.ReadString('\n')
			_ = message
		}
	}()
	return client
}

type TopSessionEntry struct {
	sessionId uint64
	score     uint32
}

func getTopSessions(pool *redis.Pool, minutes int64) ([]TopSessionEntry, int, int) {

	redisClient := pool.Get()

	redisClient.Send("ZREVRANGE", fmt.Sprintf("s-%d", minutes-1), 0, 999, "WITHSCORES")
	redisClient.Send("ZREVRANGE", fmt.Sprintf("s-%d", minutes), 0, 999, "WITHSCORES")
	redisClient.Send("ZCARD", fmt.Sprintf("s-%d", minutes-1))
	redisClient.Send("ZCARD", fmt.Sprintf("s-%d", minutes))
	redisClient.Send("ZCARD", fmt.Sprintf("n-%d", minutes-1))
	redisClient.Send("ZCARD", fmt.Sprintf("n-%d", minutes))

	redisClient.Flush()

	topSessions_a, err := redis.Strings(redisClient.Receive())
	if err != nil {
		panic(err)
	}

	topSessions_b, err := redis.Strings(redisClient.Receive())
	if err != nil {
		panic(err)
	}

	totalSessionCount_a, err := redis.Int(redisClient.Receive())
	if err != nil {
		panic(err)
	}

	totalSessionCount_b, err := redis.Int(redisClient.Receive())
	if err != nil {
		panic(err)
	}

	nextSessionCount_a, err := redis.Int(redisClient.Receive())
	if err != nil {
		panic(err)
	}

	nextSessionCount_b, err := redis.Int(redisClient.Receive())
	if err != nil {
		panic(err)
	}

	redisClient.Close()

	topSessionsMap := make(map[uint64]TopSessionEntry)

	for i := 0; i < len(topSessions_a); i += 2 {
		sessionId, _ := strconv.ParseUint(topSessions_a[i], 16, 64)
		score, _ := strconv.ParseUint(topSessions_a[i+1], 10, 32)
		topSessionsMap[sessionId] = TopSessionEntry{
			sessionId: uint64(sessionId),
			score:     uint32(score),
		}
	}

	for i := 0; i < len(topSessions_b); i += 2 {
		sessionId, _ := strconv.ParseUint(topSessions_b[i], 16, 64)
		score, _ := strconv.ParseUint(topSessions_b[i+1], 10, 32)
		topSessionsMap[sessionId] = TopSessionEntry{
			sessionId: uint64(sessionId),
			score:     uint32(score),
		}
	}

	topSessions := make([]TopSessionEntry, len(topSessionsMap))
	topSessions = topSessions[:0]
	for _, v := range topSessionsMap {
		topSessions = append(topSessions, v)
	}

	sort.SliceStable(topSessions, func(i, j int) bool { return topSessions[i].score > topSessions[j].score })

	if len(topSessions) > 1000 {
		topSessions = topSessions[:1000]
	}

	totalSessionCount := totalSessionCount_a
	if totalSessionCount_b > totalSessionCount {
		totalSessionCount = totalSessionCount_b
	}

	nextSessionCount := nextSessionCount_a
	if nextSessionCount_b > nextSessionCount {
		nextSessionCount = nextSessionCount_b
	}

	return topSessions, totalSessionCount, nextSessionCount
}

func getMapData(pool *redis.Pool, minutes int64) ([]portal.MapData, error) {

	redisClient := pool.Get()

	redisClient.Send("KEYS", "m-*")

	redisClient.Flush()

	mapKeys, err := redis.Strings(redisClient.Receive())
	if err != nil {
		return nil, err
	}

	var mapValues []string

	if len(mapKeys) > 0 {

		var keys []interface{}
		for i := range mapKeys {
			keys = append(keys, mapKeys[i])
		}

		redisClient.Send("MGET", keys...)

		redisClient.Flush()

		mapValues, err = redis.Strings(redisClient.Receive())
		if err != nil {
			return nil, err
		}

		if len(mapValues) != len(mapKeys) {
			return nil, fmt.Errorf("number of map values and map keys don't match")
		}
	}

	redisClient.Close()

	mapData := make([]portal.MapData, len(mapKeys))
	for i := range mapKeys {
		mapData[i].Parse(mapKeys[i], mapValues[i])
	}

	return mapData, nil
}

func getSessionData(pool *redis.Pool, sessionId uint64) *portal.SessionData {

	redisClient := pool.Get()

	// redisClient.Send("KEYS", "m-*")

	redisClient.Flush()

	/*
		mapKeys, err := redis.Strings(redisClient.Receive())
		if err != nil {
			panic(err)
		}
	*/

	redisClient.Close()

	// todo: handle if we can't find the session

	sessionData := portal.SessionData{}

	// todo

	return &sessionData
}

func RunCrunchThreads(redisHostname string, threadCount int ) {

	for k := 0; k < threadCount; k++ {

		go func(thread int) {

			time.Sleep(time.Duration(rand.Intn(10000)) * time.Millisecond)

			redisClient := createRedisClient(redisHostname)

			iteration := uint64(0)

			near_relay_max := uint64(0)

			for {

				start := time.Now()
				secs := start.Unix()
				minutes := secs / 60

				all_sessions := ""
				next_sessions := ""
				session_data := ""
				map_data := ""
				slice_data := []string{}
				near_relay_data := []string{}

				for j := 0; j < 1000; j++ {

					sessionId := uint64(thread*1000000) + uint64(j) + iteration

					score := rand.Intn(10000)

					zadd_data := fmt.Sprintf(" %d %016x", score, sessionId)

					all_sessions += zadd_data

					next := ((uint64(j) + iteration) % 10) == 0
					if next {
						next_sessions += zadd_data
					}

					sessionData := portal.GenerateRandomSessionData()
					session_data += fmt.Sprintf("SET ss-%016x \"%s\"\r\nEXPIRE ss-%016x 30\r\n", sessionId, sessionData.Value(), sessionId)

					mapData := portal.MapData{}
					mapData.Latitude = float32(common.RandomInt(-90000, +90000)) / 1000.0
					mapData.Longitude = float32(common.RandomInt(-18000, +18000)) / 1000.0
					mapData.Next = next

					map_data += fmt.Sprintf("SET m-%016x \"%s\"\r\nEXPIRE m-%016x 30\r\n", sessionId, mapData.Value(), sessionId)

					sliceData := portal.GenerateRandomSessionData()
					slice_data = append(slice_data, fmt.Sprintf("RPUSH sl-%016x \"%s\"\r\nEXPIRE sl-%016x\r\n", sessionId, sliceData.Value(), sessionId))

					if sessionId > near_relay_max {
						nearRelayData := portal.GenerateRandomNearRelayData()
						near_relay_data = append(slice_data, fmt.Sprintf("RPUSH nr-%016x \"%s\"\r\nEXPIRE sl-%016x\r\n", sessionId, nearRelayData.Value(), sessionId))
						near_relay_max = sessionId
					}
				}

				commands := ""

				if len(all_sessions) > 0 {
					commands += fmt.Sprintf("ZADD s-%d %s\r\n", minutes, all_sessions)
					commands += fmt.Sprintf("EXPIRE s-%d 30\r\n", minutes)
				}

				if len(next_sessions) > 0 {
					commands += fmt.Sprintf("ZADD n-%d %s\r\n", minutes, next_sessions)
					commands += fmt.Sprintf("EXPIRE n-%d 30\r\n", minutes)
				}

				if len(session_data) > 0 {
					commands += session_data
				}

				if len(map_data) > 0 {
					commands += map_data
				}

				redisClient.Write([]byte(commands))

				commands = ""
				for i := range slice_data {
					commands += slice_data[i]
					if len(commands) >= 512*1024 {
						redisClient.Write([]byte(commands))
						commands = ""
					}
				}

				if len(commands) > 0 {
					redisClient.Write([]byte(commands))
				}

				commands = ""
				for i := range near_relay_data {
					commands += near_relay_data[i]
					if len(commands) >= 512*1024 {
						redisClient.Write([]byte(commands))
						commands = ""
					}
				}

				if len(commands) > 0 {
					redisClient.Write([]byte(commands))
				}

				time.Sleep(10 * time.Second)

				iteration++
			}
		}(k)

	}
}

func RunPollThread(redisHostname string) {

	pool := createRedisPool(redisHostname)

	go func() {

		fmt.Printf("\n")

		for {

			fmt.Printf("-------------------------------------------------\n")

			start := time.Now()
			secs := start.Unix()
			minutes := secs / 60

			topSessions, totalSessionCount, nextSessionCount := getTopSessions(pool, minutes)

			fmt.Printf("top sessions: %d of %d/%d (%.1fms)\n", len(topSessions), nextSessionCount, totalSessionCount, float64(time.Since(start).Milliseconds()))

			start = time.Now()

			mapData, err := getMapData(pool, minutes)
			if err != nil {
				panic(fmt.Sprintf("failed to get map data: %v", err))
			}

			fmt.Printf("map data: %d points (%.1fms)\n", len(mapData), float64(time.Since(start).Milliseconds()))

			// todo: bring back, but return each thing separately
			/*
			if len(topSessions) > 0 {
				start = time.Now()
				sessionData := getSessionData(pool, topSessions[0].sessionId)
				fmt.Printf("session data: %d slices, %d near relay data (%.1fms)\n", len(sessionData.sliceData), len(sessionData.NearRelayData), float64(time.Since(start).Milliseconds()))
			}
			*/

			fmt.Printf("-------------------------------------------------\n")

			time.Sleep(time.Second)
		}
	}()
}

func main() {

	redisHostname := envvar.GetString("REDIS_HOSTNAME", "127.0.0.1:6379")

	threadCount := envvar.GetInt("REDIS_THREAD_COUNT", 100)

	RunCrunchThreads(redisHostname, threadCount)

	RunPollThread(redisHostname)

	time.Sleep(time.Minute)
}
