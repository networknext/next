package main

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"

	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/constants"
	"github.com/networknext/next/modules/core"
	"github.com/networknext/next/modules/envvar"

	"github.com/redis/go-redis/v9"
)

var redisNodes []string

func CreateRedisClusterClient() *redis.ClusterClient {
	clusterOptions := redis.ClusterOptions{Addrs: redisNodes}
	redisClient := redis.NewClusterClient(&clusterOptions)
	return redisClient
}

func RunSessionInsertThreads(ctx context.Context, threadCount int) {

	for k := 0; k < threadCount; k++ {

		go func(thread int) {

			redisClient := CreateRedisClusterClient()

			sessionInserter := CreateSessionInserter(redisClient, 1000)

			iteration := uint64(0)

			time.Sleep(time.Duration(rand.Intn(10000)) * time.Millisecond)

			for {

				for j := 0; j < 10000; j++ {

					sessionId := uint64(thread*1000000) + uint64(j) + iteration
					userHash := uint64(j) + iteration
					score := uint32(rand.Intn(10000))
					next := ((uint64(j) + iteration) % 10) == 0

					sessionData := GenerateRandomSessionData()

					sliceData := GenerateRandomSliceData()

					sessionInserter.Insert(ctx, sessionId, userHash, score, next, sessionData, sliceData)
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

			redisClient := CreateRedisClusterClient()

			serverInserter := CreateServerInserter(redisClient, 1000)

			iteration := uint64(0)

			time.Sleep(time.Duration(rand.Intn(10000)) * time.Millisecond)

			for {

				for j := 0; j < 1000; j++ {

					serverData := GenerateRandomServerData()

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

			redisClient := CreateRedisClusterClient()

			relayInserter := CreateRelayInserter(redisClient, 1000)

			iteration := uint64(0)

			time.Sleep(time.Duration(rand.Intn(10000)) * time.Millisecond)

			for {

				for j := 0; j < 10; j++ {

					relayData := GenerateRandomRelayData()

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

	go func() {

		redisClient := CreateRedisClusterClient()

		for {

			fmt.Printf("-------------------------------------------------\n")

			start := time.Now()

			secs := start.Unix()

			minutes := secs / 60

			totalSessionCount, nextSessionCount := GetSessionCounts(ctx, redisClient, minutes)

			fmt.Printf("sessions: %d/%d (%.1fms)\n", nextSessionCount, totalSessionCount, float64(time.Since(start).Milliseconds()))

			// ------------------------------------------------------------------------------------------

			start = time.Now()

			serverCount := GetServerCount(ctx, redisClient, minutes)

			fmt.Printf("servers: %d (%.1fms)\n", serverCount, float64(time.Since(start).Milliseconds()))

			// ------------------------------------------------------------------------------------------

			start = time.Now()

			relayCount := GetRelayCount(ctx, redisClient, minutes)

			fmt.Printf("relays: %d (%.1fms)\n", relayCount, float64(time.Since(start).Milliseconds()))

			// ------------------------------------------------------------------------------------------

			begin := 0
			end := 100

			start = time.Now()

			sessions := GetSessions(ctx, redisClient, minutes, begin, end)

			fmt.Printf("session list: %d (%.1fms)\n", len(sessions), float64(time.Since(start).Milliseconds()))

			// ------------------------------------------------------------------------------------------

			/*
				if len(sessions) > 0 {
					start = time.Now()
					sessionData, sliceData, nearRelayData := GetSessionData(ctx, redisClient, sessions[0].SessionId)
					if sessionData != nil {
						fmt.Printf("session %x: %d slices (%.1fms)\n", sessionData.SessionId, len(sliceData), float64(time.Since(start).Milliseconds()))
					}
				}
			*/

			// ------------------------------------------------------------------------------------------

			time.Sleep(time.Second)
		}
	}()
}

func main() {

	redisNodes = []string{"127.0.0.1:10000", "127.0.0.1:10001", "127.0.0.1:10002", "127.0.0.1:10003", "127.0.0.1:10004", "127.0.0.1:10005"}

	threadCount := envvar.GetInt("REDIS_THREAD_COUNT", 100)

	ctx := context.Background()

	RunSessionInsertThreads(ctx, threadCount)

	RunServerInsertThreads(ctx, threadCount)

	RunRelayInsertThreads(ctx, threadCount)

	RunPollThread(ctx)

	time.Sleep(time.Minute)
}

// --------------------------------------------------------------------------------------------------

type SessionData struct {
	SessionId      uint64  `json:"session_id,string"`
	UserHash       uint64  `json:"user_hash,string"`
	StartTime      uint64  `json:"start_time,string"`
	ISP            string  `json:"isp"`
	ConnectionType uint8   `json:"connection_type"`
	PlatformType   uint8   `json:"platform_type"`
	Latitude       float32 `json:"latitude"`
	Longitude      float32 `json:"longitude"`
	DirectRTT      uint32  `json:"direct_rtt"`
	NextRTT        uint32  `json:"next_rtt"`
	BuyerId        uint64  `json:"buyer_id,string"`
	DatacenterId   uint64  `json:"datacenter_id,string"`
	ServerAddress  string  `json:"server_address"`
}

func (data *SessionData) Value() string {
	return fmt.Sprintf("%x|%x|%x|%s|%d|%d|%.2f|%.2f|%d|%d|%x|%x|%s",
		data.SessionId,
		data.UserHash,
		data.StartTime,
		data.ISP,
		data.ConnectionType,
		data.PlatformType,
		data.Latitude,
		data.Longitude,
		data.DirectRTT,
		data.NextRTT,
		data.BuyerId,
		data.DatacenterId,
		data.ServerAddress,
	)
}

func (data *SessionData) Parse(value string) {
	values := strings.Split(value, "|")
	if len(values) != 13 {
		return
	}
	sessionId, err := strconv.ParseUint(values[0], 16, 64)
	if err != nil {
		return
	}
	userHash, err := strconv.ParseUint(values[1], 16, 64)
	if err != nil {
		return
	}
	startTime, err := strconv.ParseUint(values[2], 16, 64)
	if err != nil {
		return
	}
	isp := values[3]
	connectionType, err := strconv.ParseUint(values[4], 10, 32)
	if err != nil {
		return
	}
	platformType, err := strconv.ParseUint(values[5], 10, 32)
	if err != nil {
		return
	}
	latitude, err := strconv.ParseFloat(values[6], 32)
	if err != nil {
		return
	}
	longitude, err := strconv.ParseFloat(values[7], 32)
	if err != nil {
		return
	}
	directRTT, err := strconv.ParseUint(values[8], 10, 32)
	if err != nil {
		return
	}
	nextRTT, err := strconv.ParseUint(values[9], 10, 32)
	if err != nil {
		return
	}
	buyerId, err := strconv.ParseUint(values[10], 16, 64)
	if err != nil {
		return
	}
	datacenterId, err := strconv.ParseUint(values[11], 16, 64)
	if err != nil {
		return
	}
	serverAddress := values[12]

	data.SessionId = sessionId
	data.UserHash = userHash
	data.StartTime = startTime
	data.ISP = isp
	data.ConnectionType = uint8(connectionType)
	data.PlatformType = uint8(platformType)
	data.Latitude = float32(latitude)
	data.Longitude = float32(longitude)
	data.DirectRTT = uint32(directRTT)
	data.NextRTT = uint32(nextRTT)
	data.BuyerId = buyerId
	data.DatacenterId = datacenterId
	data.ServerAddress = serverAddress
}

func GenerateRandomSessionData() *SessionData {
	data := SessionData{}
	data.SessionId = rand.Uint64()
	data.UserHash = rand.Uint64()
	data.ISP = "Comcast Internet Company, LLC"
	data.ConnectionType = uint8(common.RandomInt(0, constants.MaxConnectionType))
	data.PlatformType = uint8(common.RandomInt(0, constants.MaxPlatformType))
	data.Latitude = float32(common.RandomInt(-9000, +9000)) / 100.0
	data.Longitude = float32(common.RandomInt(-18000, +18000)) / 100.0
	data.DirectRTT = rand.Uint32()
	data.NextRTT = rand.Uint32()
	data.BuyerId = rand.Uint64()
	data.DatacenterId = rand.Uint64()
	data.ServerAddress = fmt.Sprintf("127.0.0.1:%d", common.RandomInt(1000, 65535))
	return &data
}

// --------------------------------------------------------------------------------------------------

type SliceData struct {
	Timestamp        uint64
	SliceNumber      uint32
	DirectRTT        uint32
	NextRTT          uint32
	PredictedRTT     uint32
	DirectJitter     uint32
	NextJitter       uint32
	RealJitter       uint32
	DirectPacketLoss float32
	NextPacketLoss   float32
	RealPacketLoss   float32
	RealOutOfOrder   float32
	InternalEvents   uint64
	SessionEvents    uint64
	DirectKbpsUp     uint32
	DirectKbpsDown   uint32
	NextKbpsUp       uint32
	NextKbpsDown     uint32
	Next             bool
}

func (data *SliceData) Value() string {
	return fmt.Sprintf("%x|%d|%d|%d|%d|%d|%d|%d|%.2f|%.2f|%.2f|%.2f|%x|%x|%d|%d|%d|%d|%v",
		data.Timestamp,
		data.SliceNumber,
		data.DirectRTT,
		data.NextRTT,
		data.PredictedRTT,
		data.DirectJitter,
		data.NextJitter,
		data.RealJitter,
		data.DirectPacketLoss,
		data.NextPacketLoss,
		data.RealPacketLoss,
		data.RealOutOfOrder,
		data.InternalEvents,
		data.SessionEvents,
		data.DirectKbpsUp,
		data.DirectKbpsDown,
		data.NextKbpsUp,
		data.NextKbpsDown,
		data.Next,
	)
}

func (data *SliceData) Parse(value string) {
	values := strings.Split(value, "|")
	if len(values) != 19 {
		return
	}
	timestamp, err := strconv.ParseUint(values[0], 16, 64)
	if err != nil {
		return
	}
	sliceNumber, err := strconv.ParseUint(values[1], 10, 32)
	if err != nil {
		return
	}
	directRTT, err := strconv.ParseUint(values[2], 10, 32)
	if err != nil {
		return
	}
	nextRTT, err := strconv.ParseUint(values[3], 10, 32)
	if err != nil {
		return
	}
	predictedRTT, err := strconv.ParseUint(values[4], 10, 32)
	if err != nil {
		return
	}
	directJitter, err := strconv.ParseUint(values[5], 10, 32)
	if err != nil {
		return
	}
	nextJitter, err := strconv.ParseUint(values[6], 10, 32)
	if err != nil {
		return
	}
	realJitter, err := strconv.ParseUint(values[7], 10, 32)
	if err != nil {
		return
	}
	directPacketLoss, err := strconv.ParseFloat(values[8], 32)
	if err != nil {
		return
	}
	nextPacketLoss, err := strconv.ParseFloat(values[9], 32)
	if err != nil {
		return
	}
	realPacketLoss, err := strconv.ParseFloat(values[10], 32)
	if err != nil {
		return
	}
	realOutOfOrder, err := strconv.ParseFloat(values[11], 32)
	if err != nil {
		return
	}
	internalEvents, err := strconv.ParseUint(values[12], 16, 64)
	if err != nil {
		return
	}
	sessionEvents, err := strconv.ParseUint(values[13], 16, 64)
	if err != nil {
		return
	}
	directKbpsUp, err := strconv.ParseUint(values[14], 10, 32)
	if err != nil {
		return
	}
	directKbpsDown, err := strconv.ParseUint(values[15], 10, 32)
	if err != nil {
		return
	}
	nextKbpsUp, err := strconv.ParseUint(values[16], 10, 32)
	if err != nil {
		return
	}
	nextKbpsDown, err := strconv.ParseUint(values[17], 10, 32)
	if err != nil {
		return
	}
	next := values[18] == "true"
	data.Timestamp = timestamp
	data.SliceNumber = uint32(sliceNumber)
	data.DirectRTT = uint32(directRTT)
	data.NextRTT = uint32(nextRTT)
	data.PredictedRTT = uint32(predictedRTT)
	data.DirectJitter = uint32(directJitter)
	data.NextJitter = uint32(nextJitter)
	data.RealJitter = uint32(realJitter)
	data.DirectPacketLoss = float32(directPacketLoss)
	data.NextPacketLoss = float32(nextPacketLoss)
	data.RealPacketLoss = float32(realPacketLoss)
	data.RealOutOfOrder = float32(realOutOfOrder)
	data.InternalEvents = internalEvents
	data.SessionEvents = sessionEvents
	data.DirectKbpsUp = uint32(directKbpsUp)
	data.DirectKbpsDown = uint32(directKbpsDown)
	data.NextKbpsUp = uint32(nextKbpsUp)
	data.NextKbpsDown = uint32(nextKbpsDown)
	data.Next = next
}

func GenerateRandomSliceData() *SliceData {
	data := SliceData{}
	data.Timestamp = rand.Uint64()
	data.SliceNumber = rand.Uint32()
	data.DirectRTT = rand.Uint32()
	data.NextRTT = rand.Uint32()
	data.PredictedRTT = rand.Uint32()
	data.DirectJitter = rand.Uint32()
	data.NextJitter = rand.Uint32()
	data.RealJitter = rand.Uint32()
	data.DirectPacketLoss = float32(common.RandomInt(0, 100000)) / 100.0
	data.NextPacketLoss = float32(common.RandomInt(0, 100000)) / 100.0
	data.RealPacketLoss = float32(common.RandomInt(0, 100000)) / 100.0
	data.RealOutOfOrder = float32(common.RandomInt(0, 100000)) / 100.0
	data.InternalEvents = rand.Uint64()
	data.SessionEvents = rand.Uint64()
	data.DirectKbpsUp = rand.Uint32()
	data.DirectKbpsDown = rand.Uint32()
	data.NextKbpsUp = rand.Uint32()
	data.NextKbpsDown = rand.Uint32()
	data.Next = common.RandomBool()
	return &data
}

// --------------------------------------------------------------------------------------------------

type ServerData struct {
	ServerAddress    string
	SDKVersion_Major uint8
	SDKVersion_Minor uint8
	SDKVersion_Patch uint8
	BuyerId          uint64
	DatacenterId     uint64
	NumSessions      uint32
	StartTime        uint64
}

func (data *ServerData) Value() string {
	return fmt.Sprintf("%s|%d|%d|%d|%x|%x|%d|%x",
		data.ServerAddress,
		data.SDKVersion_Major,
		data.SDKVersion_Minor,
		data.SDKVersion_Patch,
		data.BuyerId,
		data.DatacenterId,
		data.NumSessions,
		data.StartTime,
	)
}

func (data *ServerData) Parse(value string) {
	values := strings.Split(value, "|")
	if len(values) != 8 {
		return
	}
	serverAddress := values[0]
	sdkVersionMajor, err := strconv.ParseUint(values[1], 10, 8)
	if err != nil {
		return
	}
	sdkVersionMinor, err := strconv.ParseUint(values[2], 10, 8)
	if err != nil {
		return
	}
	sdkVersionPatch, err := strconv.ParseUint(values[3], 10, 8)
	if err != nil {
		return
	}
	buyerId, err := strconv.ParseUint(values[4], 16, 64)
	if err != nil {
		return
	}
	datacenterId, err := strconv.ParseUint(values[5], 16, 64)
	if err != nil {
		return
	}
	numSessions, err := strconv.ParseUint(values[6], 10, 32)
	if err != nil {
		return
	}
	startTime, err := strconv.ParseUint(values[7], 16, 64)
	if err != nil {
		return
	}
	data.ServerAddress = serverAddress
	data.SDKVersion_Major = uint8(sdkVersionMajor)
	data.SDKVersion_Minor = uint8(sdkVersionMinor)
	data.SDKVersion_Patch = uint8(sdkVersionPatch)
	data.BuyerId = buyerId
	data.DatacenterId = datacenterId
	data.NumSessions = uint32(numSessions)
	data.StartTime = startTime
}

func GenerateRandomServerData() *ServerData {
	data := ServerData{}
	data.ServerAddress = fmt.Sprintf("127.0.0.1:%d", common.RandomInt(1000, 65535))
	data.SDKVersion_Major = uint8(common.RandomInt(0, 255))
	data.SDKVersion_Minor = uint8(common.RandomInt(0, 255))
	data.SDKVersion_Patch = uint8(common.RandomInt(0, 255))
	data.BuyerId = rand.Uint64()
	data.DatacenterId = rand.Uint64()
	data.NumSessions = rand.Uint32()
	data.StartTime = rand.Uint64()
	return &data
}

// --------------------------------------------------------------------------------------------------

type RelayData struct {
	RelayName    string `json:"relay_name"`
	RelayId      uint64 `json:"relay_id,string"`
	RelayAddress string `json:"relay_address"`
	NumSessions  uint32 `json:"num_sessions"`
	MaxSessions  uint32 `json:"max_sessions"`
	StartTime    uint64 `json:"start_time,string"`
	RelayFlags   uint64 `json:"relay_flags,string"`
	RelayVersion string `json:"relay_version"`
}

func (data *RelayData) Value() string {
	return fmt.Sprintf("%s|%x|%s|%d|%d|%x|%x|%s",
		data.RelayName,
		data.RelayId,
		data.RelayAddress,
		data.NumSessions,
		data.MaxSessions,
		data.StartTime,
		data.RelayFlags,
		data.RelayVersion,
	)
}

func (data *RelayData) Parse(value string) {

	values := strings.Split(value, "|")
	if len(values) != 8 {
		return
	}
	relayName := values[0]
	relayId, err := strconv.ParseUint(values[1], 16, 64)
	if err != nil {
		return
	}
	relayAddress := values[2]
	numSessions, err := strconv.ParseUint(values[3], 10, 32)
	if err != nil {
		return
	}
	maxSessions, err := strconv.ParseUint(values[4], 10, 32)
	if err != nil {
		return
	}
	startTime, err := strconv.ParseUint(values[5], 16, 64)
	if err != nil {
		return
	}
	relayFlags, err := strconv.ParseUint(values[6], 16, 64)
	if err != nil {
		return
	}
	relayVersion := values[7]
	data.RelayName = relayName
	data.RelayId = relayId
	data.RelayAddress = relayAddress
	data.NumSessions = uint32(numSessions)
	data.MaxSessions = uint32(maxSessions)
	data.StartTime = startTime
	data.RelayFlags = relayFlags
	data.RelayVersion = relayVersion
}

func GenerateRandomRelayData() *RelayData {
	data := RelayData{}
	data.RelayName = common.RandomString(32)
	data.RelayId = rand.Uint64()
	data.RelayAddress = fmt.Sprintf("127.0.0.1:%d", common.RandomInt(1000, 65535))
	data.NumSessions = rand.Uint32()
	data.MaxSessions = rand.Uint32()
	data.StartTime = rand.Uint64()
	data.RelayFlags = rand.Uint64()
	data.RelayVersion = common.RandomString(constants.MaxRelayVersionLength)
	return &data
}

// ------------------------------------------------------------------------------------------------------------

type SessionInserter struct {
	redisClient   *redis.ClusterClient
	lastFlushTime time.Time
	batchSize     int
	numPending    int
	pipeline      redis.Pipeliner
}

func CreateSessionInserter(redisClient *redis.ClusterClient, batchSize int) *SessionInserter {
	inserter := SessionInserter{}
	inserter.redisClient = redisClient
	inserter.lastFlushTime = time.Now()
	inserter.batchSize = batchSize
	inserter.pipeline = redisClient.Pipeline()
	return &inserter
}

func (inserter *SessionInserter) Insert(ctx context.Context, sessionId uint64, userHash uint64, score uint32, next bool, sessionData *SessionData, sliceData *SliceData) {

	currentTime := time.Now()

	minutes := currentTime.Unix() / 60

	sessionIdString := fmt.Sprintf("%016x", sessionId)

	key := fmt.Sprintf("s-%d", minutes)
	inserter.pipeline.ZAdd(ctx, key, redis.Z{Score: float64(score), Member: sessionIdString})

	if next {
		key = fmt.Sprintf("n-%d", minutes)
		inserter.pipeline.ZAdd(ctx, key, redis.Z{Score: float64(score), Member: sessionIdString})
	}

	key = fmt.Sprintf("u-%016x", userHash)
	inserter.pipeline.ZAdd(ctx, key, redis.Z{Score: float64(sessionData.StartTime), Member: sessionIdString})

	key = fmt.Sprintf("sd-%s", sessionIdString)
	inserter.pipeline.Set(ctx, key, sessionData.Value(), 0)

	key = fmt.Sprintf("sl-%s", sessionIdString)
	inserter.pipeline.RPush(ctx, key, sliceData.Value())

	key = fmt.Sprintf("svs-%s-%d", sessionData.ServerAddress, minutes)
	inserter.pipeline.HSet(ctx, key, sessionIdString, currentTime.Unix())

	inserter.numPending++

	inserter.CheckForFlush(ctx, currentTime)
}

func (inserter *SessionInserter) CheckForFlush(ctx context.Context, currentTime time.Time) {
	if inserter.numPending > inserter.batchSize || currentTime.Sub(inserter.lastFlushTime) >= time.Second {
		inserter.Flush(ctx)
	}
}

func (inserter *SessionInserter) Flush(ctx context.Context) {
	_, err := inserter.pipeline.Exec(ctx)
	if err != nil {
		core.Error("session insert error: %v", err)
	}
	inserter.numPending = 0
	inserter.lastFlushTime = time.Now()
	inserter.pipeline = inserter.redisClient.Pipeline()
}

// --------------------------------------------------------------------------------------------------

func GetSessionCounts(ctx context.Context, redisClient *redis.ClusterClient, minutes int64) (int, int) {

	pipeline := redisClient.Pipeline()

	pipeline.ZCard(ctx, fmt.Sprintf("s-%03x-%d", minutes-1))
	pipeline.ZCard(ctx, fmt.Sprintf("s-%03x-%d", minutes))
	pipeline.ZCard(ctx, fmt.Sprintf("n-%03x-%d", minutes-1))
	pipeline.ZCard(ctx, fmt.Sprintf("n-%03x-%d", minutes))

	cmds, err := pipeline.Exec(ctx)
	if err != nil {
		core.Error("failed to get session counts: %v", err)
		return 0, 0
	}

	var totalSessionCount_a int
	var totalSessionCount_b int
	var nextSessionCount_a int
	var nextSessionCount_b int

	total_a := int(cmds[0].(*redis.IntCmd).Val())
	total_b := int(cmds[1].(*redis.IntCmd).Val())
	next_a := int(cmds[2].(*redis.IntCmd).Val())
	next_b := int(cmds[3].(*redis.IntCmd).Val())

	totalSessionCount_a += total_a
	totalSessionCount_b += total_b
	nextSessionCount_a += next_a
	nextSessionCount_b += next_b

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
	SessionId uint64
	Score     uint32
}

func GetSessions(ctx context.Context, redisClient *redis.ClusterClient, minutes int64, begin int, end int) []SessionData {

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

	// get session ids in order in the range [begin,end]

	pipeline := redisClient.Pipeline()

	pipeline.ZRevRangeWithScores(ctx, fmt.Sprintf("s-%d", minutes-1), int64(begin), int64(end-1))
	pipeline.ZRevRangeWithScores(ctx, fmt.Sprintf("s-%d", minutes), int64(begin), int64(end-1))

	cmds, err := pipeline.Exec(ctx)
	if err != nil {
		core.Error("failed to get sessions: %v", err)
		return nil
	}

	_ = cmds

	/*
		sessions_a := cmds[0].(*redis.ZSliceCmd).Val()
		if err != nil {
			core.Error("redis get sessions a failed: %v", err)
			return nil
		}

		sessions_b := cmds[1].(*redis.ZSliceCmd).Val()
		if err != nil {
			core.Error("redis get sessions b failed: %v", err)
			return nil
		}

		fmt.Printf("got %d sessions\n", len(sessions_a))
	*/

	/*
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

		sessionEntries := make([]SessionEntry, len(sessionsMap))
		index := 0
		for _, v := range sessionsMap {
			sessionEntries[index] = v
			index++
		}

		sort.SliceStable(sessionEntries, func(i, j int) bool { return sessionEntries[i].SessionId < sessionEntries[j].SessionId })
		sort.SliceStable(sessionEntries, func(i, j int) bool { return sessionEntries[i].Score > sessionEntries[j].Score })

		maxSize := end - begin
		if len(sessionEntries) > maxSize {
			sessionEntries = sessionEntries[:maxSize]
		}

		// now get session data for the set of session ids in [begin, end]

		if len(sessionEntries) == 0 {
			return nil
		}

		redisClient = pool.Get()

		args := redis.Args{}
		for i := range sessionEntries {
			args = args.Add(fmt.Sprintf("sd-%016x", sessionEntries[i].SessionId))
		}

		redisClient.Send("MGET", args...)

		redisClient.Flush()

		redis_session_data, err := redis.Strings(redisClient.Receive())
		if err != nil {
			core.Error("redis mget get session data failed: %v", err)
			return nil
		}

		redisClient.Close()

		sessions := make([]SessionData, len(redis_session_data))

		for i := range sessions {
			sessions[i].Parse(redis_session_data[i])
			sessions[i].SessionId = sessionEntries[i].SessionId
		}
	*/

	// todo
	sessions := make([]SessionData, 0)

	return sessions
}

/*
func GetUserSessions(pool *redis.Pool, userHash uint64, minutes int64, begin int, end int) []SessionData {

	if begin < 0 {
		core.Error("invalid begin passed to get user sessions: %d", begin)
		return nil
	}

	if end < 0 {
		core.Error("invalid end passed to get user sessions: %d", end)
		return nil
	}

	if end <= begin {
		core.Error("invalid begin passed to get user sessions: %d", begin)
		return nil
	}

	// get session ids in order in the range [begin,end]

	redisClient := pool.Get()

	redisClient.Send("ZREVRANGE", fmt.Sprintf("u-%016x-%d", userHash, minutes-1), begin, end-1, "WITHSCORES")
	redisClient.Send("ZREVRANGE", fmt.Sprintf("u-%016x-%d", userHash, minutes), begin, end-1, "WITHSCORES")

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

	sessionEntries := make([]SessionEntry, len(sessionsMap))
	index := 0
	for _, v := range sessionsMap {
		sessionEntries[index] = v
		index++
	}

	sort.SliceStable(sessionEntries, func(i, j int) bool { return sessionEntries[i].SessionId < sessionEntries[j].SessionId })
	sort.SliceStable(sessionEntries, func(i, j int) bool { return sessionEntries[i].Score > sessionEntries[j].Score })

	maxSize := end - begin
	if len(sessionEntries) > maxSize {
		sessionEntries = sessionEntries[:maxSize]
	}

	// now get session data for the set of session ids in [begin, end]

	if len(sessionEntries) == 0 {
		return nil
	}

	redisClient = pool.Get()

	args := redis.Args{}
	for i := range sessionEntries {
		args = args.Add(fmt.Sprintf("sd-%016x", sessionEntries[i].SessionId))
	}

	redisClient.Send("MGET", args...)

	redisClient.Flush()

	redis_session_data, err := redis.Strings(redisClient.Receive())
	if err != nil {
		core.Error("redis mget get session data failed: %v", err)
		return nil
	}

	redisClient.Close()

	sessions := make([]SessionData, len(redis_session_data))

	for i := range sessions {
		sessions[i].Parse(redis_session_data[i])
		sessions[i].SessionId = sessionEntries[i].SessionId
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
		nearRelayData[i].Parse(redis_near_relay_data[i])
	}

	return &sessionData, sliceData, nearRelayData
}
*/

// ------------------------------------------------------------------------------------------------------------

type ServerInserter struct {
	redisClient   *redis.ClusterClient
	lastFlushTime time.Time
	batchSize     int
	numPending    int
	pipeline      redis.Pipeliner
}

func CreateServerInserter(redisClient *redis.ClusterClient, batchSize int) *ServerInserter {
	inserter := ServerInserter{}
	inserter.redisClient = redisClient
	inserter.lastFlushTime = time.Now()
	inserter.batchSize = batchSize
	inserter.pipeline = redisClient.Pipeline()
	return &inserter
}

func (inserter *ServerInserter) Insert(ctx context.Context, serverData *ServerData) {

	currentTime := time.Now()

	minutes := currentTime.Unix() / 60

	serverId := common.HashString(serverData.ServerAddress)

	score := uint32(serverId) ^ uint32(serverId>>32)

	key := fmt.Sprintf("sv-%d", minutes)
	inserter.pipeline.ZAdd(ctx, key, redis.Z{Score: float64(score), Member: serverData.ServerAddress})

	inserter.pipeline.Set(ctx, fmt.Sprintf("svd-%s", serverData.ServerAddress), serverData.Value(), 0)

	inserter.numPending++

	inserter.CheckForFlush(ctx, currentTime)
}

func (inserter *ServerInserter) CheckForFlush(ctx context.Context, currentTime time.Time) {
	if inserter.numPending > inserter.batchSize || currentTime.Sub(inserter.lastFlushTime) >= time.Second {
		inserter.Flush(ctx)
	}
}

func (inserter *ServerInserter) Flush(ctx context.Context) {
	_, err := inserter.pipeline.Exec(ctx)
	if err != nil {
		core.Error("server insert error: %v", err)
	}
	inserter.numPending = 0
	inserter.lastFlushTime = time.Now()
	inserter.pipeline = inserter.redisClient.Pipeline()
}

// ------------------------------------------------------------------------------------------------------------

func GetServerCount(ctx context.Context, redisClient *redis.ClusterClient, minutes int64) int {

	pipeline := redisClient.Pipeline()

	pipeline.ZCard(ctx, fmt.Sprintf("sv-%d", minutes-1))
	pipeline.ZCard(ctx, fmt.Sprintf("sv-%d", minutes))

	cmds, err := pipeline.Exec(ctx)
	if err != nil {
		core.Error("failed to get server counts: %v", err)
		return 0
	}

	var totalServerCount_a int
	var totalServerCount_b int

	total_a := int(cmds[0].(*redis.IntCmd).Val())
	total_b := int(cmds[1].(*redis.IntCmd).Val())

	totalServerCount_a += total_a
	totalServerCount_b += total_b

	totalServerCount := totalServerCount_a
	if totalServerCount_b > totalServerCount {
		totalServerCount = totalServerCount_b
	}

	return totalServerCount
}

// ------------------------------------------------------------------------------------------------------------

type RelayInserter struct {
	redisClient   *redis.ClusterClient
	lastFlushTime time.Time
	batchSize     int
	numPending    int
	pipeline      redis.Pipeliner
}

func CreateRelayInserter(redisClient *redis.ClusterClient, batchSize int) *RelayInserter {
	inserter := RelayInserter{}
	inserter.redisClient = redisClient
	inserter.lastFlushTime = time.Now()
	inserter.batchSize = batchSize
	inserter.pipeline = redisClient.Pipeline()
	return &inserter
}

func (inserter *RelayInserter) Insert(ctx context.Context, relayData *RelayData) {

	currentTime := time.Now()

	minutes := currentTime.Unix() / 60

	score := uint32(relayData.RelayId) ^ uint32(relayData.RelayId>>32)

	key := fmt.Sprintf("r-%d", minutes)
	inserter.pipeline.ZAdd(ctx, key, redis.Z{Score: float64(score), Member: relayData.RelayAddress})

	inserter.pipeline.Set(ctx, fmt.Sprintf("rd-%s", relayData.RelayAddress), relayData.Value(), 0)

	inserter.numPending++

	inserter.CheckForFlush(ctx, currentTime)
}

func (inserter *RelayInserter) CheckForFlush(ctx context.Context, currentTime time.Time) {
	if inserter.numPending > inserter.batchSize || currentTime.Sub(inserter.lastFlushTime) >= time.Second {
		inserter.Flush(ctx)
	}
}

func (inserter *RelayInserter) Flush(ctx context.Context) {
	_, err := inserter.pipeline.Exec(ctx)
	if err != nil {
		core.Error("relay insert error: %v", err)
	}
	inserter.numPending = 0
	inserter.lastFlushTime = time.Now()
	inserter.pipeline = inserter.redisClient.Pipeline()
}

// ------------------------------------------------------------------------------------------------------------

func GetRelayCount(ctx context.Context, redisClient *redis.ClusterClient, minutes int64) int {

	pipeline := redisClient.Pipeline()

	pipeline.ZCard(ctx, fmt.Sprintf("r-%d", minutes-1))
	pipeline.ZCard(ctx, fmt.Sprintf("r-%d", minutes))

	cmds, err := pipeline.Exec(ctx)
	if err != nil {
		core.Error("failed to get relay counts: %v", err)
		return 0
	}

	var totalRelayCount_a int
	var totalRelayCount_b int

	total_a := int(cmds[0].(*redis.IntCmd).Val())
	total_b := int(cmds[1].(*redis.IntCmd).Val())

	totalRelayCount_a += total_a
	totalRelayCount_b += total_b

	totalRelayCount := totalRelayCount_a
	if totalRelayCount_b > totalRelayCount {
		totalRelayCount = totalRelayCount_b
	}

	return totalRelayCount
}

// ------------------------------------------------------------------------------------------------------------
