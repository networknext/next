package portal

import (
	"fmt"
	"sort"
	"strconv"
	"time"
	"net"

	"github.com/networknext/backend/modules/core"

	"github.com/gomodule/redigo/redis"
)

func CreateRedisPool(hostname string) *redis.Pool {
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

type SessionEntry struct {
	SessionId uint64
	Score     uint32
}

func GetSessions(pool *redis.Pool, minutes int64, begin int, end int) ([]SessionEntry, int, int) {

	if begin < 0 {
		core.Error("invalid begin passed to get sessions: %d", begin)
		return nil, 0, 0
	}

	if end < 0 {
		core.Error("invalid end passed to get sessions: %d", end)
		return nil, 0, 0
	}

	if end <= begin {
		core.Error("invalid begin passed to get sessions: %d", begin)
		return nil, 0, 0
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
		core.Error("redis get sessions a failed: %v", err)
		return nil, 0, 0
	}

	sessions_b, err := redis.Strings(redisClient.Receive())
	if err != nil {
		core.Error("redis get sessions b failed: %v", err)
		return nil, 0, 0
	}

	totalSessionCount_a, err := redis.Int(redisClient.Receive())
	if err != nil {
		core.Error("redis get total sessions count a failed: %v", err)
		return nil, 0, 0
	}

	totalSessionCount_b, err := redis.Int(redisClient.Receive())
	if err != nil {
		core.Error("redis get total sessions count b failed: %v", err)
		return nil, 0, 0
	}

	nextSessionCount_a, err := redis.Int(redisClient.Receive())
	if err != nil {
		core.Error("redis get next sessions count a failed: %v", err)
		return nil, 0, 0
	}

	nextSessionCount_b, err := redis.Int(redisClient.Receive())
	if err != nil {
		core.Error("redis get next sessions count b failed: %v", err)
		return nil, 0, 0
	}

	redisClient.Close()

	sessionsMap := make(map[uint64]SessionEntry)

	for i := 0; i < len(sessions_a); i += 2 {
		sessionId, _ := strconv.ParseUint(sessions_a[i], 16, 64)
		score, _ := strconv.ParseUint(sessions_a[i+1], 10, 32)
		sessionsMap[sessionId] = SessionEntry{
			SessionId: uint64(sessionId),
			Score:     uint32(score),
		}
	}

	for i := 0; i < len(sessions_b); i += 2 {
		sessionId, _ := strconv.ParseUint(sessions_b[i], 16, 64)
		score, _ := strconv.ParseUint(sessions_b[i+1], 10, 32)
		sessionsMap[sessionId] = SessionEntry{
			SessionId: uint64(sessionId),
			Score:     uint32(score),
		}
	}

	sessions := make([]SessionEntry, len(sessionsMap))
	index := 0
	for _, v := range sessionsMap {
		sessions[index] = v
		index++
	}

	sort.SliceStable(sessions, func(i, j int) bool { return sessions[i].Score > sessions[j].Score })

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

type ServerEntry struct {
	Address net.UDPAddr
	Score   uint32
}

func GetServers(pool *redis.Pool, minutes int64, begin int, end int) ([]ServerEntry, int) {

	if begin < 0 {
		core.Error("invalid begin passed to get servers: %d", begin)
		return nil, 0
	}

	if end < 0 {
		core.Error("invalid end passed to get servers: %d", end)
		return nil, 0
	}

	if end <= begin {
		core.Error("end must be greater than begin")
		return nil, 0
	}

	redisClient := pool.Get()

	redisClient.Send("ZREVRANGE", fmt.Sprintf("sv-%d", minutes-1), begin, end-1, "WITHSCORES")
	redisClient.Send("ZREVRANGE", fmt.Sprintf("sv-%d", minutes), begin, end-1, "WITHSCORES")
	redisClient.Send("ZCARD", fmt.Sprintf("sv-%d", minutes-1))
	redisClient.Send("ZCARD", fmt.Sprintf("sv-%d", minutes))

	redisClient.Flush()

	servers_a, err := redis.Strings(redisClient.Receive())
	if err != nil {
		core.Error("redis get servers a failed: %v", err)
		return nil, 0
	}

	servers_b, err := redis.Strings(redisClient.Receive())
	if err != nil {
		core.Error("redis get servers b failed: %v", err)
		return nil, 0
	}

	totalServerCount_a, err := redis.Int(redisClient.Receive())
	if err != nil {
		core.Error("redis get server count a failed: %v", err)
		return nil, 0
	}

	totalServerCount_b, err := redis.Int(redisClient.Receive())
	if err != nil {
		core.Error("redis get server count b failed: %v", err)
		return nil, 0
	}

	redisClient.Close()

	serverMap := make(map[string]ServerEntry)

	for i := 0; i < len(servers_a); i += 2 {
		address := core.ParseAddress(servers_a[i])
		score, _ := strconv.ParseUint(servers_a[i+1], 10, 32)
		serverMap[servers_a[i]] = ServerEntry{
			Address: address,
			Score:   uint32(score),
		}
	}

	for i := 0; i < len(servers_b); i += 2 {
		address := core.ParseAddress(servers_b[i])
		score, _ := strconv.ParseUint(servers_b[i+1], 10, 32)
		serverMap[servers_b[i]] = ServerEntry{
			Address: address,
			Score:   uint32(score),
		}
	}

	servers := make([]ServerEntry, len(serverMap))
	index := 0
	for _, v := range serverMap {
		servers[index] = v
		index++
	}

	sort.SliceStable(servers, func(i, j int) bool { return servers[i].Score > servers[j].Score })

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

type RelayEntry struct {
	Address net.UDPAddr
	Score        uint32
}

func GetRelays(pool *redis.Pool, minutes int64, begin int, end int) ([]RelayEntry, int) {

	if begin < 0 {
		core.Error("invalid begin passed to get relays: %d", begin)
		return nil, 0
	}

	if end < 0 {
		core.Error("invalid end passed to get servers: %d", end)
		return nil, 0
	}

	if end <= begin {
		core.Error("end must be greater than begin")
		return nil, 0
	}

	redisClient := pool.Get()

	redisClient.Send("ZREVRANGE", fmt.Sprintf("r-%d", minutes-1), begin, end-1, "WITHSCORES")
	redisClient.Send("ZREVRANGE", fmt.Sprintf("r-%d", minutes), begin, end-1, "WITHSCORES")
	redisClient.Send("ZCARD", fmt.Sprintf("r-%d", minutes-1))
	redisClient.Send("ZCARD", fmt.Sprintf("r-%d", minutes))

	redisClient.Flush()

	relays_a, err := redis.Strings(redisClient.Receive())
	if err != nil {
		core.Error("redis get relays a failed: %v", err)
		return nil, 0
	}

	relays_b, err := redis.Strings(redisClient.Receive())
	if err != nil {
		core.Error("redis get relays b failed: %v", err)
		return nil, 0
	}

	totalRelayCount_a, err := redis.Int(redisClient.Receive())
	if err != nil {
		core.Error("redis get relays count a failed: %v", err)
		return nil, 0
	}

	totalRelayCount_b, err := redis.Int(redisClient.Receive())
	if err != nil {
		core.Error("redis get relays count b failed: %v", err)
		return nil, 0
	}

	redisClient.Close()

	relayMap := make(map[string]RelayEntry)

	for i := 0; i < len(relays_a); i += 2 {
		address := core.ParseAddress(relays_a[i])
		score, _ := strconv.ParseUint(relays_a[i+1], 10, 32)
		relayMap[address.String()] = RelayEntry{
			Address: address,
			Score:   uint32(score),
		}
	}

	for i := 0; i < len(relays_b); i += 2 {
		address := core.ParseAddress(relays_b[i])
		score, _ := strconv.ParseUint(relays_b[i+1], 10, 32)
		relayMap[address.String()] = RelayEntry{
			Address: address,
			Score:   uint32(score),
		}
	}
	
	relays := make([]RelayEntry, len(relayMap))
	index := 0
	for _, v := range relayMap {
		relays[index] = v
		index++
	}

	sort.SliceStable(relays, func(i, j int) bool { return relays[i].Score > relays[j].Score })

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

func GetMapData(pool *redis.Pool, minutes int64) ([]MapData, error) {

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

	mapData := make([]MapData, len(mapKeys))
	for i := range mapKeys {
		mapData[i].Parse(mapKeys[i], mapValues[i])
	}

	return mapData, nil
}

func GetSessionData(pool *redis.Pool, sessionId uint64) (*SessionData, []SliceData, []NearRelayData) {

	redisClient := pool.Get()

	redisClient.Send("GET", fmt.Sprintf("sd-%016x", sessionId))
	redisClient.Send("LRANGE", fmt.Sprintf("sl-%016x", sessionId), 0, -1)
	redisClient.Send("LRANGE", fmt.Sprintf("nr-%016x", sessionId), 0, -1)

	redisClient.Flush()

	redis_session_data, err := redis.String(redisClient.Receive())
	if err != nil {
		return nil, nil, nil
	}

	redis_slice_data, err := redis.Strings(redisClient.Receive())
	if err != nil {
		return nil, nil, nil
	}

	redis_near_relay_data, err := redis.Strings(redisClient.Receive())
	if err != nil {
		return nil, nil, nil
	}

	redisClient.Close()

	sessionData := SessionData{}
	sessionData.Parse(redis_session_data)

	sliceData := make([]SliceData, len(redis_slice_data))
	for i := 0; i < len(redis_slice_data); i++ {
		sliceData[i].Parse(redis_slice_data[i])
	}

	nearRelayData := make([]NearRelayData, len(redis_near_relay_data))
	for i := 0; i < len(redis_near_relay_data); i++ {
		sliceData[i].Parse(redis_near_relay_data[i])
	}

	return &sessionData, sliceData, nearRelayData
}
