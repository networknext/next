package portal

import (
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/core"

	"github.com/gomodule/redigo/redis"
)

// ------------------------------------------------------------------------------------------------------------

func GetSessionCounts(pool *redis.Pool, minutes int64) (int, int) {

	redisClient := pool.Get()

	redisClient.Send("ZCARD", fmt.Sprintf("s-%d", minutes-1))
	redisClient.Send("ZCARD", fmt.Sprintf("s-%d", minutes))
	redisClient.Send("ZCARD", fmt.Sprintf("n-%d", minutes-1))
	redisClient.Send("ZCARD", fmt.Sprintf("n-%d", minutes))

	redisClient.Flush()

	totalSessionCount_a, err := redis.Int(redisClient.Receive())
	if err != nil {
		core.Error("redis get total sessions count a failed: %v", err)
		return 0, 0
	}

	totalSessionCount_b, err := redis.Int(redisClient.Receive())
	if err != nil {
		core.Error("redis get total sessions count b failed: %v", err)
		return 0, 0
	}

	nextSessionCount_a, err := redis.Int(redisClient.Receive())
	if err != nil {
		core.Error("redis get next sessions count a failed: %v", err)
		return 0, 0
	}

	nextSessionCount_b, err := redis.Int(redisClient.Receive())
	if err != nil {
		core.Error("redis get next sessions count b failed: %v", err)
		return 0, 0
	}

	redisClient.Close()

	totalSessionCount := totalSessionCount_a
	if totalSessionCount_b > totalSessionCount {
		totalSessionCount = totalSessionCount_b
	}

	nextSessionCount := nextSessionCount_a
	if nextSessionCount_b > nextSessionCount {
		nextSessionCount = nextSessionCount_b
	}

	return totalSessionCount, nextSessionCount
}

type SessionEntry struct {
	SessionId uint64 `json:"session_id"`
	Score     uint32 `json:"score"`
}

func GetSessions(pool *redis.Pool, minutes int64, begin int, end int) []SessionEntry {

	if begin < 0 {
		core.Error("invalid begin passed to get sessions: %d", begin)
		return nil
	}

	if end < 0 {
		core.Error("invalid end passed to get sessions: %d", end)
		return nil
	}

	if end <= begin {
		core.Error("invalid begin passed to get sessions: %d", begin)
		return nil
	}

	redisClient := pool.Get()

	redisClient.Send("ZREVRANGE", fmt.Sprintf("s-%d", minutes-1), begin, end-1, "WITHSCORES")
	redisClient.Send("ZREVRANGE", fmt.Sprintf("s-%d", minutes), begin, end-1, "WITHSCORES")

	redisClient.Flush()

	sessions_a, err := redis.Strings(redisClient.Receive())
	if err != nil {
		core.Error("redis get sessions a failed: %v", err)
		return nil
	}

	sessions_b, err := redis.Strings(redisClient.Receive())
	if err != nil {
		core.Error("redis get sessions b failed: %v", err)
		return nil
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

	return sessions
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

// ------------------------------------------------------------------------------------------------------------

func GetServerCount(pool *redis.Pool, minutes int64) int {

	redisClient := pool.Get()

	redisClient.Send("ZCARD", fmt.Sprintf("sv-%d", minutes-1))
	redisClient.Send("ZCARD", fmt.Sprintf("sv-%d", minutes))

	redisClient.Flush()

	serverCount_a, err := redis.Int(redisClient.Receive())
	if err != nil {
		core.Error("redis get server count a failed: %v", err)
		return 0
	}

	serverCount_b, err := redis.Int(redisClient.Receive())
	if err != nil {
		core.Error("redis get server count b failed: %v", err)
		return 0
	}

	redisClient.Close()

	serverCount := serverCount_a
	if serverCount_b > serverCount {
		serverCount = serverCount_b
	}

	return serverCount
}

type ServerEntry struct {
	Address string `json:"address"`
	Score   uint64 `json:"score"`
}

func GetServers(pool *redis.Pool, minutes int64, begin int, end int) []ServerEntry {

	if begin < 0 {
		core.Error("invalid begin passed to get servers: %d", begin)
		return nil
	}

	if end < 0 {
		core.Error("invalid end passed to get servers: %d", end)
		return nil
	}

	if end <= begin {
		core.Error("end must be greater than begin")
		return nil
	}

	redisClient := pool.Get()

	redisClient.Send("ZREVRANGE", fmt.Sprintf("sv-%d", minutes-1), begin, end-1, "WITHSCORES")
	redisClient.Send("ZREVRANGE", fmt.Sprintf("sv-%d", minutes), begin, end-1, "WITHSCORES")

	redisClient.Flush()

	servers_a, err := redis.Strings(redisClient.Receive())
	if err != nil {
		core.Error("redis get servers a failed: %v", err)
		return nil
	}

	servers_b, err := redis.Strings(redisClient.Receive())
	if err != nil {
		core.Error("redis get servers b failed: %v", err)
		return nil
	}

	redisClient.Close()

	serverMap := make(map[string]ServerEntry)

	for i := 0; i < len(servers_b); i += 2 {
		address := servers_b[i]
		score, _ := strconv.ParseUint(servers_b[i+1], 10, 64)
		serverMap[servers_b[i]] = ServerEntry{
			Address: address,
			Score:   uint64(score),
		}
	}

	for i := 0; i < len(servers_a); i += 2 {
		address := servers_a[i]
		score, _ := strconv.ParseUint(servers_a[i+1], 10, 64)
		serverMap[servers_a[i]] = ServerEntry{
			Address: address,
			Score:   uint64(score),
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

	return servers
}

func GetServerData(pool *redis.Pool, serverAddress string) *ServerData {

	redisClient := pool.Get()

	redisClient.Send("GET", fmt.Sprintf("svd-%016x", serverAddress))

	redisClient.Flush()

	redis_server_data, err := redis.String(redisClient.Receive())
	if err != nil {
		return nil
	}

	redisClient.Close()

	serverData := ServerData{}
	serverData.Parse(redis_server_data)

	return &serverData
}

// ------------------------------------------------------------------------------------------------------

// todo: GetRelayCount

type RelayEntry struct {
	Address string `json:"address"`
	Score   uint32 `json:"score"`
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
		address := relays_b[i]
		score, _ := strconv.ParseUint(relays_b[i+1], 10, 32)
		relayMap[address] = RelayEntry{
			Address: address,
			Score:   uint32(score),
		}
	}

	for i := 0; i < len(relays_a); i += 2 {
		address := relays_a[i]
		score, _ := strconv.ParseUint(relays_a[i+1], 10, 32)
		relayMap[address] = RelayEntry{
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

// ----------------------------------------------------------------------------------------------------

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
		inserter.redisClient.Do("")
		inserter.redisClient.Close()
		inserter.numPending = 0
		inserter.lastFlushTime = time.Now()
		inserter.redisClient = inserter.redisPool.Get()
		inserter.allSessions = redis.Args{}
		inserter.nextSessions = redis.Args{}
	}
}

// ----------------------------------------------------------------------------------

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
		inserter.redisClient.Do("")
		inserter.redisClient.Close()
		inserter.numPending = 0
		inserter.lastFlushTime = time.Now()
		inserter.redisClient = inserter.redisPool.Get()
	}
}

// ----------------------------------------------------------------------------------

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

func (inserter *ServerInserter) Insert(serverData *ServerData) {

	currentTime := time.Now()

	minutes := currentTime.Unix() / 60

	if len(inserter.servers) == 0 {
		inserter.servers = redis.Args{}.Add(fmt.Sprintf("sv-%d", minutes))
	}

	score := common.HashString(serverData.ServerAddress)

	inserter.servers = inserter.servers.Add(score)
	inserter.servers = inserter.servers.Add(serverData.ServerAddress)

	inserter.redisClient.Send("SET", fmt.Sprintf("svd-%016x", serverData.ServerAddress), serverData.Value())
	inserter.redisClient.Send("EXPIRE", fmt.Sprintf("svd-%016x", serverData.ServerAddress), 30)

	inserter.numPending++

	if inserter.numPending > inserter.batchSize || currentTime.Sub(inserter.lastFlushTime) >= time.Second {
		if len(inserter.servers) > 1 {
			inserter.redisClient.Send("ZADD", inserter.servers...)
			inserter.redisClient.Send("EXPIRE", fmt.Sprintf("sv-%d", minutes), 30)
		}
		inserter.redisClient.Do("")
		inserter.redisClient.Close()
		inserter.numPending = 0
		inserter.lastFlushTime = time.Now()
		inserter.redisClient = inserter.redisPool.Get()
		inserter.servers = redis.Args{}
	}
}

// ----------------------------------------------------------------------------------

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

func (inserter *RelayInserter) Insert(relayData *RelayData) {

	currentTime := time.Now()

	minutes := currentTime.Unix() / 60

	if len(inserter.relays) == 0 {
		inserter.relays = redis.Args{}.Add(fmt.Sprintf("r-%d", minutes))
	}

	score := relayData.RelayId

	inserter.relays = inserter.relays.Add(score)
	inserter.relays = inserter.relays.Add(relayData.RelayAddress)

	inserter.numPending++

	if inserter.numPending > inserter.batchSize || currentTime.Sub(inserter.lastFlushTime) >= time.Second {
		if len(inserter.relays) > 1 {
			inserter.redisClient.Send("ZADD", inserter.relays...)
			inserter.redisClient.Send("EXPIRE", fmt.Sprintf("r-%d", minutes), 30)
		}
		inserter.redisClient.Do("")
		inserter.redisClient.Close()
		inserter.numPending = 0
		inserter.lastFlushTime = time.Now()
		inserter.redisClient = inserter.redisPool.Get()
		inserter.relays = redis.Args{}
	}
}

// ----------------------------------------------------------------------------------
