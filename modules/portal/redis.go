package portal

import (
	"fmt"
	"net"
	"sort"
	"strconv"
	"time"

	"github.com/networknext/backend/modules/core"

	"github.com/gomodule/redigo/redis"
)

func CreateRedisPool(hostname string, size int) *redis.Pool {
	pool := redis.Pool{
		MaxIdle:     size * 10,
		MaxActive:   size,
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

	for i := 0; i < len(sessions_b); i += 2 {
		sessionId, _ := strconv.ParseUint(sessions_b[i], 16, 64)
		score, _ := strconv.ParseUint(sessions_b[i+1], 10, 32)
		sessionsMap[sessionId] = SessionEntry{
			SessionId: uint64(sessionId),
			Score:     uint32(score),
		}
	}

	for i := 0; i < len(sessions_a); i += 2 {
		sessionId, _ := strconv.ParseUint(sessions_a[i], 16, 64)
		score, _ := strconv.ParseUint(sessions_a[i+1], 10, 32)
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

	for i := 0; i < len(servers_b); i += 2 {
		address := core.ParseAddress(servers_b[i])
		score, _ := strconv.ParseUint(servers_b[i+1], 10, 32)
		serverMap[servers_b[i]] = ServerEntry{
			Address: address,
			Score:   uint32(score),
		}
	}

	for i := 0; i < len(servers_a); i += 2 {
		address := core.ParseAddress(servers_a[i])
		score, _ := strconv.ParseUint(servers_a[i+1], 10, 32)
		serverMap[servers_a[i]] = ServerEntry{
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
	Score   uint32
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

	for i := 0; i < len(relays_b); i += 2 {
		address := core.ParseAddress(relays_b[i])
		score, _ := strconv.ParseUint(relays_b[i+1], 10, 32)
		relayMap[address.String()] = RelayEntry{
			Address: address,
			Score:   uint32(score),
		}
	}

	for i := 0; i < len(relays_a); i += 2 {
		address := core.ParseAddress(relays_a[i])
		score, _ := strconv.ParseUint(relays_a[i+1], 10, 32)
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

func GetMapData(pool *redis.Pool, currentTime time.Time) ([]MapData, error) {

	seconds := uint64(currentTime.Unix())
	minutes := seconds / 60

	redisClient := pool.Get()

	redisClient.Send("HGETALL", fmt.Sprintf("m-%d", minutes))
	redisClient.Send("HGETALL", fmt.Sprintf("m-%d", minutes-1))

	redisClient.Flush()

	keys_and_values_a, _ := redis.Strings(redisClient.Receive())
	keys_and_values_b, _ := redis.Strings(redisClient.Receive())

	mapHash := make(map[string]string)

	for i := 0; i < len(keys_and_values_a); i += 2 {
		mapHash[keys_and_values_a[i]] = keys_and_values_a[i+1]
	}

	for i := 0; i < len(keys_and_values_b); i += 2 {
		mapHash[keys_and_values_b[i]] = keys_and_values_b[i+1]
	}

	mapData := make([]MapData, len(mapHash))
	index := 0
	for k, v := range mapHash {
		mapData[index].Parse(k, v)
		if seconds-mapData[index].LastUpdateTime > 30 {
			continue
		}
		index++
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

type SessionInserter struct {
	redisPool     *redis.Pool
	redisClient   redis.Conn
	lastFlushTime time.Time
	batchSize     int
	numPending    int
	allSessions   redis.Args
	nextSessions  redis.Args
}

func CreateSessionInserter(pool *redis.Pool, batchSize int) *SessionInserter {
	inserter := SessionInserter{}
	inserter.redisPool = pool
	inserter.redisClient = pool.Get()
	inserter.lastFlushTime = time.Now()
	inserter.batchSize = batchSize
	return &inserter
}

func (inserter *SessionInserter) Insert(sessionId uint64, score uint32, next bool, sessionData *SessionData, sliceData *SliceData) {

	currentTime := time.Now()

	minutes := currentTime.Unix() / 60

	if len(inserter.allSessions) == 0 {
		inserter.allSessions = redis.Args{}.Add(fmt.Sprintf("s-%d", minutes))
		inserter.nextSessions = redis.Args{}.Add(fmt.Sprintf("n-%d", minutes))
	}

	inserter.allSessions = inserter.allSessions.Add(score)
	inserter.allSessions = inserter.allSessions.Add(fmt.Sprintf("%016x", sessionId))

	if next {
		inserter.nextSessions = inserter.nextSessions.Add(score)
		inserter.nextSessions = inserter.nextSessions.Add(fmt.Sprintf("%016x", sessionId))
	}

	inserter.redisClient.Send("SET", fmt.Sprintf("sd-%016x", sessionId), sessionData.Value())
	inserter.redisClient.Send("EXPIRE", fmt.Sprintf("sd-%016x", sessionId), 30)

	inserter.redisClient.Send("RPUSH", fmt.Sprintf("sl-%016x", sessionId), sliceData.Value())
	inserter.redisClient.Send("EXPIRE", fmt.Sprintf("sl-%016x", sessionId), 30)

	mapData := MapData{}
	mapData.Latitude = sessionData.Latitude
	mapData.Longitude = sessionData.Longitude
	mapData.Next = next
	mapData.LastUpdateTime = uint64(currentTime.Unix())
	inserter.redisClient.Send("HSET", fmt.Sprintf("m-%d", minutes), fmt.Sprintf("%016x", sessionId), mapData.Value())
	inserter.redisClient.Send("EXPIRE", fmt.Sprintf("m-%d", minutes), 30)

	inserter.numPending++

	inserter.CheckForFlush(currentTime)
}

func (inserter *SessionInserter) CheckForFlush(currentTime time.Time) {
	if inserter.numPending > inserter.batchSize || currentTime.Sub(inserter.lastFlushTime) >= time.Second {
		minutes := currentTime.Unix() / 60
		if len(inserter.allSessions) > 1 {
			inserter.redisClient.Send("ZADD", inserter.allSessions...)
			inserter.redisClient.Send("EXPIRE", fmt.Sprintf("s-%d", minutes), 30)
		}
		if len(inserter.nextSessions) > 1 {
			inserter.redisClient.Send("ZADD", inserter.nextSessions...)
			inserter.redisClient.Send("EXPIRE", fmt.Sprintf("n-%d", minutes), 30)
		}
		inserter.redisClient.Flush()
		inserter.redisClient.Close()
		inserter.numPending = 0
		inserter.lastFlushTime = time.Now()
		inserter.redisClient = inserter.redisPool.Get()
		inserter.allSessions = redis.Args{}
		inserter.nextSessions = redis.Args{}
	}
}

type NearRelayInserter struct {
	redisPool     *redis.Pool
	redisClient   redis.Conn
	lastFlushTime time.Time
	batchSize     int
	numPending    int
}

func CreateNearRelayInserter(pool *redis.Pool, batchSize int) *NearRelayInserter {
	inserter := NearRelayInserter{}
	inserter.redisPool = pool
	inserter.redisClient = pool.Get()
	inserter.lastFlushTime = time.Now()
	inserter.batchSize = batchSize
	return &inserter
}

func (inserter *NearRelayInserter) Insert(sessionId uint64, nearRelayData *NearRelayData) {

	currentTime := time.Now()

	inserter.redisClient.Send("RPUSH", fmt.Sprintf("nr-%016x", sessionId), nearRelayData.Value())
	inserter.redisClient.Send("EXPIRE", fmt.Sprintf("nr-%016x", sessionId), 30)

	inserter.numPending++

	if inserter.numPending > inserter.batchSize || currentTime.Sub(inserter.lastFlushTime) >= time.Second {
		inserter.redisClient.Flush()
		inserter.redisClient.Close()
		inserter.numPending = 0
		inserter.lastFlushTime = time.Now()
		inserter.redisClient = inserter.redisPool.Get()
	}
}

type ServerInserter struct {
	redisPool     *redis.Pool
	redisClient   redis.Conn
	lastFlushTime time.Time
	batchSize     int
	numPending    int
	servers       redis.Args
}

func CreateServerInserter(pool *redis.Pool, batchSize int) *ServerInserter {
	inserter := ServerInserter{}
	inserter.redisPool = pool
	inserter.redisClient = pool.Get()
	inserter.lastFlushTime = time.Now()
	inserter.batchSize = batchSize
	return &inserter
}

func (inserter *ServerInserter) Insert(score uint32, serverData *ServerData) {

	currentTime := time.Now()

	minutes := currentTime.Unix() / 60

	if len(inserter.servers) == 0 {
		inserter.servers = redis.Args{}.Add(fmt.Sprintf("sv-%d", minutes))
	}

	inserter.servers = inserter.servers.Add(score)
	inserter.servers = inserter.servers.Add(serverData.ServerAddress.String())

	inserter.numPending++

	if inserter.numPending > inserter.batchSize || currentTime.Sub(inserter.lastFlushTime) >= time.Second {
		if len(inserter.servers) > 1 {
			inserter.redisClient.Send("ZADD", inserter.servers...)
			inserter.redisClient.Send("EXPIRE", fmt.Sprintf("sv-%d", minutes), 30)
		}
		inserter.redisClient.Flush()
		inserter.redisClient.Close()
		inserter.numPending = 0
		inserter.lastFlushTime = time.Now()
		inserter.redisClient = inserter.redisPool.Get()
		inserter.servers = redis.Args{}
	}
}

type RelayInserter struct {
	redisPool     *redis.Pool
	redisClient   redis.Conn
	lastFlushTime time.Time
	batchSize     int
	numPending    int
	relays        redis.Args
}

func CreateRelayInserter(pool *redis.Pool, batchSize int) *RelayInserter {
	inserter := RelayInserter{}
	inserter.redisPool = pool
	inserter.redisClient = pool.Get()
	inserter.lastFlushTime = time.Now()
	inserter.batchSize = batchSize
	return &inserter
}

func (inserter *RelayInserter) Insert(score uint32, relayData *RelayData) {

	currentTime := time.Now()

	minutes := currentTime.Unix() / 60

	if len(inserter.relays) == 0 {
		inserter.relays = redis.Args{}.Add(fmt.Sprintf("r-%d", minutes))
	}

	inserter.relays = inserter.relays.Add(score)
	inserter.relays = inserter.relays.Add(relayData.RelayAddress.String())

	inserter.numPending++

	if inserter.numPending > inserter.batchSize || currentTime.Sub(inserter.lastFlushTime) >= time.Second {
		if len(inserter.relays) > 1 {
			inserter.redisClient.Send("ZADD", inserter.relays...)
			inserter.redisClient.Send("EXPIRE", fmt.Sprintf("r-%d", minutes), 30)
		}
		inserter.redisClient.Flush()
		inserter.redisClient.Close()
		inserter.numPending = 0
		inserter.lastFlushTime = time.Now()
		inserter.redisClient = inserter.redisPool.Get()
		inserter.relays = redis.Args{}
	}
}
