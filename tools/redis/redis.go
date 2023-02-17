package main

import (
	"bufio"
	"fmt"
	"math/rand"
	"net"
	"os"
	"sort"
	// "strings"
	"strconv"
	"time"

	"github.com/gomodule/redigo/redis"
)

func getEnvInt(name string, defaultValue int) int {
	string, ok := os.LookupEnv(name)
	if !ok {
		return defaultValue
	}
	value, err := strconv.ParseInt(string, 10, 64)
	if err != nil {
		panic(fmt.Sprintf("env string is not an integer: %s\n", name))
	}
	return int(value)
}

func createRedisPool(hostNameEnv string) *redis.Pool {
	hostname := os.Getenv(hostNameEnv)
	if hostname == "" {
		hostname = "127.0.0.1:6379"
	}
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

func createRedisClient(hostNameEnv string) net.Conn {
	// todo: this is crazy. needs better error handling, and reconnect support...
	hostname := os.Getenv(hostNameEnv)
	if hostname == "" {
		hostname = "127.0.0.1:6379"
	}
	client, err := net.Dial("tcp", hostname)
	if err != nil {
		panic(err)
	}
	go func() {
		reader := bufio.NewReader(client)
		for {
			message, _ := reader.ReadString('\n')
			_ = message
		}
	}()
	return client
}

type TopSessionEntry struct {
	sessionId string
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
	redisClient.Send("KEYS", "m-*")

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

	mapKeys, err := redis.Strings(redisClient.Receive())
	if err != nil {
		panic(err)
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
			panic(err)
		}

		if len(mapValues) != len(mapKeys) {
			panic("could not get map values")
		}
	}

	redisClient.Close()

	topSessionsMap := make(map[string]TopSessionEntry)

	for i := 0; i < len(topSessions_a); i += 2 {
		sessionId := topSessions_a[i]
		score, _ := strconv.ParseUint(topSessions_a[i+1], 10, 32)
		topSessionsMap[sessionId] = TopSessionEntry{
			sessionId: sessionId,
			score:     uint32(score),
		}
	}

	for i := 0; i < len(topSessions_b); i += 2 {
		sessionId := topSessions_b[i]
		score, _ := strconv.ParseUint(topSessions_b[i+1], 10, 32)
		topSessionsMap[sessionId] = TopSessionEntry{
			sessionId: sessionId,
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

	// todo: actually convert the map values to binary data
	/*
		fmt.Printf("---------------------------------------\n")
		for i := 0; i < len(mapValues); i++ {
			fmt.Printf("%s -> %s\n", mapKeys[i], mapValues[i])
		}
		fmt.Printf("---------------------------------------\n")
	*/

	return topSessions, totalSessionCount, nextSessionCount
}

type MapData struct {
	sessionId uint64
	latitude  int16
	longitude int16
	next      bool
}

func getMapData(pool *redis.Pool, minutes int64) []MapData {

	redisClient := pool.Get()

	redisClient.Send("KEYS", "m-*")

	redisClient.Flush()

	mapKeys, err := redis.Strings(redisClient.Receive())
	if err != nil {
		panic(err)
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
			panic(err)
		}

		if len(mapValues) != len(mapKeys) {
			panic("could not get map values")
		}
	}

	redisClient.Close()

	mapData := make([]MapData, len(mapKeys))

	// todo
	/*
	for i := range mapKeys {
		mapData[i].sessionId, _ = strconv.ParseUint(mapKeys[i][2:], 16, 64)
		values := strings.Split(mapValues[i], "|")
		latitude, _ := strconv.ParseInt(values[0], 10, 16)
		longitude, _ := strconv.ParseInt(values[1], 10, 16)
		mapData[i].latitude = int16(latitude)
		mapData[i].longitude = int16(longitude)
		mapData[i].next = values[2] == "1"
	}
	*/

	return mapData
}

func main() {

	pool := createRedisPool("REDIS_HOST")

	threadCount := getEnvInt("REDIS_THREAD_COUNT", 100)

	for k := 0; k < threadCount; k++ {

		go func(thread int) {

			time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

			redisClient := createRedisClient("REDIS_HOST")

			iteration := uint64(0)

			for {

				start := time.Now()
				secs := start.Unix()
				minutes := secs / 60

				all_sessions := ""
				next_sessions := ""
				session_data := ""
				map_data := ""
				slice_data := []string{}

				for j := 0; j < 100; j++ {

					sessionId := uint64(thread*1000000) + uint64(j) + iteration

					score := sessionId % 1000

					zadd_data := fmt.Sprintf(" %d %016x", score, sessionId)

					all_sessions += zadd_data

					next := (sessionId % 10) == 0
					if next {
						next_sessions += zadd_data
					}

					session_data += fmt.Sprintf("SET ss-%016x \"Comcast ISP Name, LLC|1|2|latitude|longitude|a45c351912345781|a45c351912345781|12345781a45c3519|127.0.0.1:50000|MatchId\"\r\nEXPIRE ss-%016x 30\r\n", sessionId, sessionId)

					map_data += fmt.Sprintf("SET m-%016x \"-150,200,1\"\r\nEXPIRE m-%016x 30\r\n", sessionId, sessionId)

					slice_data = append(slice_data, fmt.Sprintf("RPUSH sl-%016x \"SliceNumber|Timestamp|DirectRTT|NextRTT|PredictedRTT|DirectJitter|NextJitter|RealJitter|DirectPacketLoss|NextPacketLoss|RealPacketLoss|RealOutOfOrder|INTERNALEVENTS|SESSIONEVENTS\"\r\nEXPIRE sl-%016x\r\n", sessionId, sessionId))
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

				time.Sleep(time.Second)

				iteration++
			}
		}(k)

	}

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

			mapData := getMapData(pool, minutes)

			fmt.Printf("map data: %d points (%.1fms)\n", len(mapData), float64(time.Since(start).Milliseconds()))

			fmt.Printf("-------------------------------------------------\n")

			time.Sleep(time.Second)
		}
	}()

	duration := getEnvInt("DURATION", 60)

	if duration < 0 {
		for {
			time.Sleep(time.Minute)
		}
	}

	time.Sleep(time.Second * time.Duration(duration))
}
