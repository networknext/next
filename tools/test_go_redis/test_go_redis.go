package main

import (
	"fmt"
	"time"
	"math/rand"
	"context"
	"strings"
	"strconv"

	"github.com/networknext/next/modules/constants"
	"github.com/networknext/next/modules/core"
	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/envvar"

	"github.com/redis/go-redis/v9"
)

const NumBuckets = 256

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

			_ = redisClient

			/*
			relayInserter := CreateRelayInserter(pool, 1000)

			iteration := uint64(0)

			time.Sleep(time.Duration(rand.Intn(10000)) * time.Millisecond)

			for {

				for j := 0; j < 10; j++ {

					relayData := GenerateRandomRelayData()
					relaySample := GenerateRandomRelaySample()

					id := uint32(iteration + uint64(j))

					relayData.RelayAddress = fmt.Sprintf("%d.%d.%d.%d:%d", id&0xFF, (id>>8)&0xFF, (id>>16)&0xFF, (id>>24)&0xFF, uint64(thread))

					relayInserter.Insert(relayData, relaySample)
				}

				time.Sleep(10 * time.Second)

				iteration++
			}
			*/
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

	bucket := sessionId % NumBuckets

	sessionIdString := fmt.Sprintf("%016x", sessionId)

	key := fmt.Sprintf("s-%03x-%d", bucket, minutes)
	inserter.pipeline.ZAdd(ctx, key, redis.Z{Score: float64(score), Member: sessionIdString})

	if next {
		key = fmt.Sprintf("n-%03x-%d", bucket, minutes)
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

	for i := 0; i < NumBuckets; i++ {
		pipeline.ZCard(ctx, fmt.Sprintf("s-%03x-%d", i, minutes-1))
		pipeline.ZCard(ctx, fmt.Sprintf("s-%03x-%d", i, minutes))
		pipeline.ZCard(ctx, fmt.Sprintf("n-%03x-%d", i, minutes-1))
		pipeline.ZCard(ctx, fmt.Sprintf("n-%03x-%d", i, minutes))
	}

	cmds, err := pipeline.Exec(ctx)
	if err != nil {
		core.Error("failed to get session counts: %v", err)
		return 0,0
	}

	var totalSessionCount_a int
	var totalSessionCount_b int
	var nextSessionCount_a int
	var nextSessionCount_b int

	for i := 0; i < NumBuckets*4; i+=4 {

		total_a := int(cmds[i].(*redis.IntCmd).Val())
		total_b := int(cmds[i+1].(*redis.IntCmd).Val())
		next_a := int(cmds[i+2].(*redis.IntCmd).Val())
		next_b := int(cmds[i+3].(*redis.IntCmd).Val())

		totalSessionCount_a += total_a
		totalSessionCount_b += total_b
		nextSessionCount_a += next_a
		nextSessionCount_b += next_b
	}

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

	bucket := serverId % NumBuckets

	score := uint32(serverId) ^ uint32(serverId>>32)

	key := fmt.Sprintf("sv-%03x-%d", bucket, minutes)
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

	for i := 0; i < NumBuckets; i++ {
		pipeline.ZCard(ctx, fmt.Sprintf("sv-%03x-%d", i, minutes-1))
		pipeline.ZCard(ctx, fmt.Sprintf("sv-%03x-%d", i, minutes))
	}

	cmds, err := pipeline.Exec(ctx)
	if err != nil {
		core.Error("failed to get server counts: %v", err)
		return 0
	}

	var totalServerCount_a int
	var totalServerCount_b int

	for i := 0; i < NumBuckets*2; i+=2 {

		total_a := int(cmds[i].(*redis.IntCmd).Val())
		total_b := int(cmds[i+1].(*redis.IntCmd).Val())

		totalServerCount_a += total_a
		totalServerCount_b += total_b
	}

	totalServerCount := totalServerCount_a
	if totalServerCount_b > totalServerCount {
		totalServerCount = totalServerCount_b
	}

	return totalServerCount
}
// ------------------------------------------------------------------------------------------------------------
