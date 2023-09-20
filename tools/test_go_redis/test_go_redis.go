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

					sessionInserter.Insert(sessionId, userHash, score, next, sessionData, sliceData)
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

			time.Sleep(time.Second)
		}
	}()
}

func main() {

	redisNodes = []string{"127.0.0.1:10000", "127.0.0.1:10001", "127.0.0.1:10002", "127.0.0.1:10003", "127.0.0.1:10004", "127.0.0.1:10005"}

	threadCount := envvar.GetInt("REDIS_THREAD_COUNT", 100)

	ctx := context.Background()

	RunSessionInsertThreads(ctx, threadCount)

	RunPollThread(ctx)

	time.Sleep(time.Minute)
}

// --------------------------------------------------------------------------------------------------

const NumBuckets = 256

func GetSessionCounts(ctx context.Context, redisClient *redis.ClusterClient, minutes int64) (int, int) {

	pipeline := redisClient.Pipeline()

	for i := 0; i < NumBuckets; i++ {
		pipeline.ZCard(ctx, fmt.Sprintf("s-%02x-%d", minutes-1))
		pipeline.ZCard(ctx, fmt.Sprintf("s-%02x-%d", minutes))
		pipeline.ZCard(ctx, fmt.Sprintf("n-%02x-%d", minutes-1))
		pipeline.ZCard(ctx, fmt.Sprintf("n-%02x-%d", minutes))
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

// ------------------------------------------------------------------------------------------------------------

type SessionInserter struct {
	redisClient   *redis.ClusterClient
	lastFlushTime time.Time
	batchSize     int
	numPending    int
	/*
	allSessions   redis.Args
	nextSessions  redis.Args
	userSessions  redis.Args
	*/
}

func CreateSessionInserter(redisClient *redis.ClusterClient, batchSize int) *SessionInserter {
	inserter := SessionInserter{}
	inserter.redisClient = redisClient
	inserter.lastFlushTime = time.Now()
	inserter.batchSize = batchSize
	return &inserter
}

func (inserter *SessionInserter) Insert(sessionId uint64, userHash uint64, score uint32, next bool, sessionData *SessionData, sliceData *SliceData) {

	currentTime := time.Now()

	minutes := currentTime.Unix() / 60

	_ = minutes

	/*
	if len(inserter.allSessions) == 0 {
		inserter.allSessions = redis.Args{}.Add(fmt.Sprintf("s-%d", minutes))
		inserter.nextSessions = redis.Args{}.Add(fmt.Sprintf("n-%d", minutes))
		inserter.userSessions = redis.Args{}.Add(fmt.Sprintf("u-%016x-%d", userHash, minutes))
	}

	sessionIdString := fmt.Sprintf("%016x", sessionId)

	inserter.allSessions = inserter.allSessions.Add(score)
	inserter.allSessions = inserter.allSessions.Add(sessionIdString)

	if next {
		inserter.nextSessions = inserter.nextSessions.Add(score)
		inserter.nextSessions = inserter.nextSessions.Add(sessionIdString)
	}

	inserter.userSessions = inserter.userSessions.Add(sessionData.StartTime)
	inserter.userSessions = inserter.userSessions.Add(sessionIdString)

	key := fmt.Sprintf("sd-%s", sessionIdString)
	inserter.redisClient.Send("SET", key, sessionData.Value())

	key = fmt.Sprintf("sl-%s", sessionIdString)
	inserter.redisClient.Send("RPUSH", key, sliceData.Value())

	key = fmt.Sprintf("svs-%s-%d", sessionData.ServerAddress, minutes)
	inserter.redisClient.Send("HSET", key, sessionIdString, currentTime.Unix())

	inserter.numPending++

	inserter.CheckForFlush(currentTime)
	*/
}

func (inserter *SessionInserter) CheckForFlush(currentTime time.Time) {
	/*
	if inserter.numPending > inserter.batchSize || currentTime.Sub(inserter.lastFlushTime) >= time.Second {
		if len(inserter.allSessions) > 1 {
			inserter.redisClient.Send("ZADD", inserter.allSessions...)
		}
		if len(inserter.nextSessions) > 1 {
			inserter.redisClient.Send("ZADD", inserter.nextSessions...)
		}
		if len(inserter.userSessions) > 1 {
			inserter.redisClient.Send("ZADD", inserter.userSessions...)
		}
		inserter.redisClient.Do("")
		inserter.redisClient.Close()
		inserter.numPending = 0
		inserter.lastFlushTime = time.Now()
		inserter.redisClient = inserter.redisPool.Get()
		inserter.allSessions = redis.Args{}
		inserter.nextSessions = redis.Args{}
		inserter.userSessions = redis.Args{}
	}
	*/
}

// ------------------------------------------------------------------------------------------------------------
