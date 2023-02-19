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

type ServerEntry struct {
	serverAddress net.UDPAddr
	score         uint32
}

func getServers(pool *redis.Pool, minutes int64, begin int, end int) ([]ServerEntry, int) {

	if begin < 0 {
		panic(fmt.Sprintf("invalid begin passed to get servers: %d", begin))
	}

	if end < 0 {
		panic(fmt.Sprintf("invalid end passed to get servers: %d", end))
	}

	if end <= begin {
		panic("end must be greater than begin")
	}

	redisClient := pool.Get()

	redisClient.Send("ZREVRANGE", fmt.Sprintf("sv-%d", minutes-1), begin, end-1, "WITHSCORES")
	redisClient.Send("ZREVRANGE", fmt.Sprintf("sv-%d", minutes), begin, end-1, "WITHSCORES")
	redisClient.Send("ZCARD", fmt.Sprintf("sv-%d", minutes-1))
	redisClient.Send("ZCARD", fmt.Sprintf("sv-%d", minutes))

	redisClient.Flush()

	servers_a, err := redis.Strings(redisClient.Receive())
	if err != nil {
		panic(err)
	}

	servers_b, err := redis.Strings(redisClient.Receive())
	if err != nil {
		panic(err)
	}

	totalServerCount_a, err := redis.Int(redisClient.Receive())
	if err != nil {
		panic(err)
	}

	totalServerCount_b, err := redis.Int(redisClient.Receive())
	if err != nil {
		panic(err)
	}

	redisClient.Close()

	serverMap := make(map[uint64]ServerEntry)

	// todo
	_ = servers_a
	_ = servers_b

	/*
	for i := 0; i < len(servers_a); i += 2 {
		sessionId, _ := strconv.ParseUint(sessions_a[i], 16, 64)
		score, _ := strconv.ParseUint(sessions_a[i+1], 10, 32)
		sessionsMap[sessionId] = SessionEntry{
			sessionId: uint64(sessionId),
			score:     uint32(score),
		}
	}

	for i := 0; i < len(sessions_b); i += 2 {
		sessionId, _ := strconv.ParseUint(sessions_b[i], 16, 64)
		score, _ := strconv.ParseUint(sessions_b[i+1], 10, 32)
		sessionsMap[sessionId] = SessionEntry{
			sessionId: uint64(sessionId),
			score:     uint32(score),
		}
	}
	*/

	servers := make([]ServerEntry, len(serverMap))
	servers = servers[:0] // todo: wut
	for _, v := range serverMap {
		servers = append(servers, v)
	}

	sort.SliceStable(servers, func(i, j int) bool { return servers[i].score > servers[j].score })

	maxSize := end - begin
	if len(servers) > maxSize {
		servers = servers[:maxSize]
	}

	totalServerCount := totalServerCount_a
	if totalServerCount_b > totalServerCount {
		totalServerCount = totalServerCount_b
	}

	return servers, totalServerCount
}

type SessionEntry struct {
	sessionId uint64
	score     uint32
}

func getSessions(pool *redis.Pool, minutes int64, begin int, end int) ([]SessionEntry, int, int) {

	if begin < 0 {
		panic(fmt.Sprintf("invalid begin passed to get sessions: %d", begin))
	}

	if end < 0 {
		panic(fmt.Sprintf("invalid end passed to get sessions: %d", end))
	}

	if end <= begin {
		panic("end must be greater than begin")
	}

	redisClient := pool.Get()

	redisClient.Send("ZREVRANGE", fmt.Sprintf("s-%d", minutes-1), begin, end-1, "WITHSCORES")
	redisClient.Send("ZREVRANGE", fmt.Sprintf("s-%d", minutes), begin, end-1, "WITHSCORES")
	redisClient.Send("ZCARD", fmt.Sprintf("s-%d", minutes-1))
	redisClient.Send("ZCARD", fmt.Sprintf("s-%d", minutes))
	redisClient.Send("ZCARD", fmt.Sprintf("n-%d", minutes-1))
	redisClient.Send("ZCARD", fmt.Sprintf("n-%d", minutes))

	redisClient.Flush()

	sessions_a, err := redis.Strings(redisClient.Receive())
	if err != nil {
		panic(err)
	}

	sessions_b, err := redis.Strings(redisClient.Receive())
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

	sessionsMap := make(map[uint64]SessionEntry)

	for i := 0; i < len(sessions_a); i += 2 {
		sessionId, _ := strconv.ParseUint(sessions_a[i], 16, 64)
		score, _ := strconv.ParseUint(sessions_a[i+1], 10, 32)
		sessionsMap[sessionId] = SessionEntry{
			sessionId: uint64(sessionId),
			score:     uint32(score),
		}
	}

	for i := 0; i < len(sessions_b); i += 2 {
		sessionId, _ := strconv.ParseUint(sessions_b[i], 16, 64)
		score, _ := strconv.ParseUint(sessions_b[i+1], 10, 32)
		sessionsMap[sessionId] = SessionEntry{
			sessionId: uint64(sessionId),
			score:     uint32(score),
		}
	}

	sessions := make([]SessionEntry, len(sessionsMap))
	sessions = sessions[:0]
	for _, v := range sessionsMap {
		sessions = append(sessions, v)
	}

	sort.SliceStable(sessions, func(i, j int) bool { return sessions[i].score > sessions[j].score })

	maxSize := end - begin
	if len(sessions) > maxSize {
		sessions = sessions[:maxSize]
	}

	totalSessionCount := totalSessionCount_a
	if totalSessionCount_b > totalSessionCount {
		totalSessionCount = totalSessionCount_b
	}

	nextSessionCount := nextSessionCount_a
	if nextSessionCount_b > nextSessionCount {
		nextSessionCount = nextSessionCount_b
	}

	return sessions, totalSessionCount, nextSessionCount
}

type RelayEntry struct {
	relayAddress net.UDPAddr
	score         uint32
}

func getRelays(pool *redis.Pool, minutes int64, begin int, end int) ([]RelayEntry, int) {

	if begin < 0 {
		panic(fmt.Sprintf("invalid begin passed to get relays: %d", begin))
	}

	if end < 0 {
		panic(fmt.Sprintf("invalid end passed to get relays: %d", end))
	}

	if end <= begin {
		panic("end must be greater than begin")
	}

	redisClient := pool.Get()

	redisClient.Send("ZREVRANGE", fmt.Sprintf("r-%d", minutes-1), begin, end-1, "WITHSCORES")
	redisClient.Send("ZREVRANGE", fmt.Sprintf("r-%d", minutes), begin, end-1, "WITHSCORES")
	redisClient.Send("ZCARD", fmt.Sprintf("r-%d", minutes-1))
	redisClient.Send("ZCARD", fmt.Sprintf("r-%d", minutes))

	redisClient.Flush()

	relays_a, err := redis.Strings(redisClient.Receive())
	if err != nil {
		panic(err)
	}

	relays_b, err := redis.Strings(redisClient.Receive())
	if err != nil {
		panic(err)
	}

	totalRelayCount_a, err := redis.Int(redisClient.Receive())
	if err != nil {
		panic(err)
	}

	totalRelayCount_b, err := redis.Int(redisClient.Receive())
	if err != nil {
		panic(err)
	}

	redisClient.Close()

	serverMap := make(map[uint64]RelayEntry)

	// todo
	_ = relays_a
	_ = relays_b

	/*
	for i := 0; i < len(relays_a); i += 2 {
		sessionId, _ := strconv.ParseUint(sessions_a[i], 16, 64)
		score, _ := strconv.ParseUint(sessions_a[i+1], 10, 32)
		sessionsMap[sessionId] = SessionEntry{
			sessionId: uint64(sessionId),
			score:     uint32(score),
		}
	}

	for i := 0; i < len(sessions_b); i += 2 {
		sessionId, _ := strconv.ParseUint(sessions_b[i], 16, 64)
		score, _ := strconv.ParseUint(sessions_b[i+1], 10, 32)
		sessionsMap[sessionId] = SessionEntry{
			sessionId: uint64(sessionId),
			score:     uint32(score),
		}
	}
	*/

	relays := make([]RelayEntry, len(serverMap))
	relays = relays[:0] // todo: wut
	for _, v := range serverMap {
		relays = append(relays, v)
	}

	sort.SliceStable(relays, func(i, j int) bool { return relays[i].score > relays[j].score })

	maxSize := end - begin
	if len(relays) > maxSize {
		relays = relays[:maxSize]
	}

	totalRelayCount := totalRelayCount_a
	if totalRelayCount_b > totalRelayCount {
		totalRelayCount = totalRelayCount_b
	}

	return relays, totalRelayCount
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

func getSessionData(pool *redis.Pool, sessionId uint64) (*portal.SessionData, []portal.SliceData, []portal.NearRelayData) {

	redisClient := pool.Get()

	redisClient.Send("GET", fmt.Sprintf("sd-%016x", sessionId))
	redisClient.Send("LRANGE", fmt.Sprintf("sl-%016x", sessionId), 0, -1)
	redisClient.Send("LRANGE", fmt.Sprintf("nr-%016x", sessionId), 0, -1)

	redisClient.Flush()

	redis_session_data, err := redis.String(redisClient.Receive())
	if err != nil {
		panic(err)
	}

	redis_slice_data, err := redis.Strings(redisClient.Receive())
	if err != nil {
		panic(err)
	}

	redis_near_relay_data, err := redis.Strings(redisClient.Receive())
	if err != nil {
		panic(err)
	}

	redisClient.Close()

	sessionData := portal.SessionData{}
	sessionData.Parse(redis_session_data);

	sliceData := make([]portal.SliceData, len(redis_slice_data))
	for i := 0; i < len(redis_slice_data); i++ {
		sliceData[i].Parse(redis_slice_data[i])
	}

	nearRelayData := make([]portal.NearRelayData, len(redis_near_relay_data))
	for i := 0; i < len(redis_near_relay_data); i++ {
		sliceData[i].Parse(redis_near_relay_data[i])
	}

	return &sessionData, sliceData, nearRelayData
}

func RunSessionCrunchThreads(redisHostname string, threadCount int ) {

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
					session_data += fmt.Sprintf("SET sd-%016x \"%s\"\r\nEXPIRE sd-%016x 30\r\n", sessionId, sessionData.Value(), sessionId)

					mapData := portal.MapData{}
					mapData.Latitude = float32(common.RandomInt(-90000, +90000)) / 1000.0
					mapData.Longitude = float32(common.RandomInt(-18000, +18000)) / 1000.0
					mapData.Next = next

					map_data += fmt.Sprintf("SET m-%016x \"%s\"\r\nEXPIRE m-%016x 30\r\n", sessionId, mapData.Value(), sessionId)

					sliceData := portal.GenerateRandomSessionData()
					slice_data = append(slice_data, fmt.Sprintf("RPUSH sl-%016x \"%s\"\r\nEXPIRE sl-%016x 30\r\n", sessionId, sliceData.Value(), sessionId))

					if sessionId > near_relay_max {
						nearRelayData := portal.GenerateRandomNearRelayData()
						near_relay_data = append(near_relay_data, fmt.Sprintf("RPUSH nr-%016x \"%s\"\r\nEXPIRE nr-%016x 3600\r\n", sessionId, nearRelayData.Value(), sessionId))
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

func RunServerCrunchThreads(redisHostname string, threadCount int ) {

	for k := 0; k < threadCount; k++ {

		go func(thread int) {

			time.Sleep(time.Duration(rand.Intn(10000)) * time.Millisecond)

			redisClient := createRedisClient(redisHostname)

			iteration := uint64(0)

			for {

				start := time.Now()
				secs := start.Unix()
				minutes := secs / 60

				servers := ""

				for j := 0; j < 1000; j++ {

					serverAddress := fmt.Sprintf("127.0.0.1:%d", uint16(iteration+uint64(j)))

					score := rand.Intn(10000)

					servers += fmt.Sprintf(" %d %s", score, serverAddress)

					/*
					serverData := portal.GenerateRandomSessionData()
					server_data += fmt.Sprintf("SET sd-%016x \"%s\"\r\nEXPIRE sd-%016x 30\r\n", sessionId, sessionData.Value(), sessionId)
					*/
				}

				commands := ""

				if len(servers) > 0 {
					commands += fmt.Sprintf("ZADD sv-%d %s\r\n", minutes, servers)
					commands += fmt.Sprintf("EXPIRE sv-%d 30\r\n", minutes)
				}

				redisClient.Write([]byte(commands))

				time.Sleep(10 * time.Second)

				iteration++
			}
		}(k)

	}
}

func RunRelayCrunchThreads(redisHostname string, threadCount int ) {

	for k := 0; k < threadCount; k++ {

		go func(thread int) {

			time.Sleep(time.Duration(rand.Intn(10000)) * time.Millisecond)

			redisClient := createRedisClient(redisHostname)

			iteration := uint64(0)

			for {

				start := time.Now()
				secs := start.Unix()
				minutes := secs / 60

				relays := ""

				for j := 0; j < 1000; j++ {

					relayAddress := fmt.Sprintf("127.0.0.1:%d", uint16(iteration+uint64(j)))

					score := rand.Intn(10000)

					relays += fmt.Sprintf(" %d %s", score, relayAddress)

					/*
					serverData := portal.GenerateRandomSessionData()
					server_data += fmt.Sprintf("SET sd-%016x \"%s\"\r\nEXPIRE sd-%016x 30\r\n", sessionId, sessionData.Value(), sessionId)
					*/
				}

				commands := ""

				if len(relays) > 0 {
					commands += fmt.Sprintf("ZADD r-%d %s\r\n", minutes, relays)
					commands += fmt.Sprintf("EXPIRE r-%d 30\r\n", minutes)
				}

				redisClient.Write([]byte(commands))

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

			begin := 0
			end := 1000

			sessions, totalSessionCount, nextSessionCount := getSessions(pool, minutes, begin, end)

			fmt.Printf("sessions: %d of %d/%d (%.1fms)\n", len(sessions), nextSessionCount, totalSessionCount, float64(time.Since(start).Milliseconds()))

			start  = time.Now()

			if len(sessions) > 0 {
				start = time.Now()
				sessionData, sliceData, nearRelayData := getSessionData(pool, sessions[0].sessionId)
				fmt.Printf("session data: %x session id, %d slices, %d near relay data (%.1fms)\n", sessionData.SessionId, len(sliceData), len(nearRelayData), float64(time.Since(start).Milliseconds()))
			}

			start = time.Now()

			mapData, err := getMapData(pool, minutes)
			if err != nil {
				panic(fmt.Sprintf("failed to get map data: %v", err))
			}

			fmt.Printf("map data: %d points (%.1fms)\n", len(mapData), float64(time.Since(start).Milliseconds()))

			start = time.Now()

			servers, totalServerCount := getServers(pool, minutes, begin, end)

			fmt.Printf("servers: %d of %d (%.1fms)\n", len(servers), totalServerCount, float64(time.Since(start).Milliseconds()))

			start = time.Now()

			relays, totalRelayCount := getServers(pool, minutes, begin, end)

			fmt.Printf("relays: %d of %d (%.1fms)\n", len(relays), totalRelayCount, float64(time.Since(start).Milliseconds()))

			fmt.Printf("-------------------------------------------------\n")

			time.Sleep(time.Second)
		}
	}()
}

func main() {

	redisHostname := envvar.GetString("REDIS_HOSTNAME", "127.0.0.1:6379")

	threadCount := envvar.GetInt("REDIS_THREAD_COUNT", 100)

	RunServerCrunchThreads(redisHostname, threadCount)
	RunSessionCrunchThreads(redisHostname, threadCount)
	RunRelayCrunchThreads(redisHostname, threadCount)

	RunPollThread(redisHostname)

	time.Sleep(time.Minute)
}
