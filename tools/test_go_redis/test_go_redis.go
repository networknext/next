package main

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"strings"
	"time"
	"sort"

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

			nearRelayInserter := CreateNearRelayInserter(redisClient, 1000)

			near_relay_max := uint64(0)

			time.Sleep(time.Duration(rand.Intn(10000)) * time.Millisecond)

			for {

				for j := 0; j < 10000; j++ {

					sessionId := uint64(thread*1000000) + uint64(j) + iteration

					userHash := uint64(j) + iteration

					sessionData := GenerateRandomSessionData()

					sessionData.SessionId = sessionId

					sliceData := GenerateRandomSliceData()

					sessionInserter.Insert(ctx, sessionId, userHash, sessionData, sliceData)

					if sessionId > near_relay_max {
						nearRelayData := GenerateRandomNearRelayData()
						nearRelayInserter.Insert(ctx, sessionId, nearRelayData)
						near_relay_max = sessionId
					}
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

	iteration := uint64(0)

	go func() {

		redisClient := CreateRedisClusterClient()

		for {

			// ------------------------------------------------------------------------------------------

			fmt.Printf("------------------------------------------------------------------------------------------------\n")

			start := time.Now()

			sessionIds := make([]uint64, 1000)
			for i := 0; i < len(sessionIds); i++ {
				sessionIds[i] = uint64(1000000) + uint64(i)
			}

			sessionList := GetSessionList(ctx, redisClient, sessionIds)
			if sessionList != nil {
				fmt.Printf("session list %d (%.3fms)\n", len(sessionList), float64(time.Since(start).Milliseconds()))
			}

			// ------------------------------------------------------------------------------------------

			start = time.Now()

			sessionId := uint64(1000000) + iteration

			sessionData, sliceData, nearRelayData := GetSessionData(ctx, redisClient, sessionId)
			if sessionData != nil {
				fmt.Printf("session data %x -> %d slices, %d near relay data (%.3fms)\n", sessionData.SessionId, len(sliceData), len(nearRelayData), float64(time.Since(start).Milliseconds()))
			}

			// ------------------------------------------------------------------------------------------

			start = time.Now()

			minutes := start.Unix() / 60

			serverCount := GetServerCount(ctx, redisClient, minutes)

			fmt.Printf("servers: %d (%.3fms)\n", serverCount, float64(time.Since(start).Milliseconds()))

			// ------------------------------------------------------------------------------------------

			start = time.Now()

			serverAddress := "208.3.0.0:15"

			serverData, serverSessions := GetServerData(ctx, redisClient, serverAddress, minutes)

			if serverData != nil {
				fmt.Printf("server data %s -> %d sessions (%.3fms)\n", serverData.ServerAddress, len(serverSessions), float64(time.Since(start).Milliseconds()))
			}

			// ------------------------------------------------------------------------------------------

			start = time.Now()

			serverAddresses := GetServerAddresses(ctx, redisClient, minutes, 0, 100)

			fmt.Printf("server addresses -> %d server addresses (%.3fms)\n", len(serverAddresses), float64(time.Since(start).Milliseconds()))

			// ------------------------------------------------------------------------------------------

			start = time.Now()

			/*
			serverAddresses := make([]string, 100)
			for i := range serverAddresses {
				serverAddresses[i] = fmt.Sprintf("208.3.0.0:%d", 15+i)
			}
			*/

			serverList := GetServerList(ctx, redisClient, serverAddresses)
			if serverList != nil {
				fmt.Printf("server list %d (%.3fms)\n", len(serverList), float64(time.Since(start).Milliseconds()))
			}

			// ------------------------------------------------------------------------------------------

			start = time.Now()

			relayCount := GetRelayCount(ctx, redisClient, minutes)

			fmt.Printf("relays: %d (%.3fms)\n", relayCount, float64(time.Since(start).Milliseconds()))

			// ------------------------------------------------------------------------------------------

			// todo: get relay data

			// ------------------------------------------------------------------------------------------

			// todo: get relay list

			// ------------------------------------------------------------------------------------------

			time.Sleep(time.Second)

			iteration++
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

type NearRelayData struct {
	Timestamp           uint64                           `json:"timestamp,string"`
	NumNearRelays       uint32                           `json:"num_near_relays"`
	NearRelayId         [constants.MaxNearRelays]uint64  `json:"near_relay_id"`
	NearRelayRTT        [constants.MaxNearRelays]uint8   `json:"near_relay_rtt"`
	NearRelayJitter     [constants.MaxNearRelays]uint8   `json:"near_relay_jitter"`
	NearRelayPacketLoss [constants.MaxNearRelays]float32 `json:"near_relay_packet_loss"`
}

func (data *NearRelayData) Value() string {
	output := fmt.Sprintf("%x|%d", data.Timestamp, data.NumNearRelays)
	for i := 0; i < int(data.NumNearRelays); i++ {
		output += fmt.Sprintf("|%x|%d|%d|%.2f", data.NearRelayId[i], data.NearRelayRTT[i], data.NearRelayJitter[i], data.NearRelayPacketLoss[i])
	}
	return output
}

func (data *NearRelayData) Parse(value string) {
	values := strings.Split(value, "|")
	if len(values) < 2 {
		return
	}
	timestamp, err := strconv.ParseUint(values[0], 16, 64)
	if err != nil {
		return
	}
	numNearRelays, err := strconv.ParseInt(values[1], 10, 8)
	if err != nil || numNearRelays < 0 || numNearRelays > constants.MaxNearRelays {
		return
	}
	if len(values) != 2+int(numNearRelays)*4 {
		return
	}
	nearRelayId := make([]uint64, numNearRelays)
	nearRelayRTT := make([]uint64, numNearRelays)
	nearRelayJitter := make([]uint64, numNearRelays)
	nearRelayPacketLoss := make([]float64, numNearRelays)
	for i := 0; i < int(numNearRelays); i++ {
		nearRelayId[i], err = strconv.ParseUint(values[2+i*4], 16, 64)
		if err != nil {
			return
		}
		nearRelayRTT[i], err = strconv.ParseUint(values[2+i*4+1], 10, 8)
		if err != nil {
			return
		}
		nearRelayJitter[i], err = strconv.ParseUint(values[2+i*4+2], 10, 8)
		if err != nil {
			return
		}
		nearRelayPacketLoss[i], err = strconv.ParseFloat(values[2+i*4+3], 32)
		if err != nil {
			return
		}
	}
	data.Timestamp = timestamp
	data.NumNearRelays = uint32(numNearRelays)
	for i := 0; i < int(numNearRelays); i++ {
		data.NearRelayId[i] = nearRelayId[i]
		data.NearRelayRTT[i] = uint8(nearRelayRTT[i])
		data.NearRelayJitter[i] = uint8(nearRelayJitter[i])
		data.NearRelayPacketLoss[i] = float32(nearRelayPacketLoss[i])
	}
	return
}

func GenerateRandomNearRelayData() *NearRelayData {
	data := NearRelayData{}
	data.Timestamp = uint64(time.Now().Unix())
	data.NumNearRelays = constants.MaxNearRelays
	for i := 0; i < int(data.NumNearRelays); i++ {
		data.NearRelayId[i] = rand.Uint64()
		data.NearRelayRTT[i] = uint8(common.RandomInt(5, 20))
		data.NearRelayJitter[i] = uint8(common.RandomInt(5, 20))
		data.NearRelayPacketLoss[i] = float32(common.RandomInt(0, 10000)) / 100.0
	}
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

func (inserter *SessionInserter) Insert(ctx context.Context, sessionId uint64, userHash uint64, sessionData *SessionData, sliceData *SliceData) {

	currentTime := time.Now()

	minutes := currentTime.Unix() / 60

	sessionIdString := fmt.Sprintf("%016x", sessionId)

	key := fmt.Sprintf("sd-%s", sessionIdString)
	inserter.pipeline.Set(ctx, key, sessionData.Value(), 0)

	key = fmt.Sprintf("sl-%s", sessionIdString)
	inserter.pipeline.RPush(ctx, key, sliceData.Value())

	key = fmt.Sprintf("u-%016x", userHash)
	inserter.pipeline.ZAdd(ctx, key, redis.Z{Score: float64(sessionData.StartTime), Member: sessionIdString})

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

func GetSessionData(ctx context.Context, redisClient *redis.ClusterClient, sessionId uint64) (*SessionData, []SliceData, []NearRelayData) {

	pipeline := redisClient.Pipeline()

	pipeline.Get(ctx, fmt.Sprintf("sd-%016x", sessionId))
	pipeline.LRange(ctx, fmt.Sprintf("sl-%016x", sessionId), 0, -1)
	pipeline.LRange(ctx, fmt.Sprintf("nr-%016x", sessionId), 0, -1)

	cmds, err := pipeline.Exec(ctx)
	if err != nil {
		core.Error("failed to get session data: %v", err)
		return nil, nil, nil
	}

	redis_session_data := cmds[0].(*redis.StringCmd).Val()
	redis_slice_data := cmds[1].(*redis.StringSliceCmd).Val()
	redis_near_relay_data := cmds[2].(*redis.StringSliceCmd).Val()

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

func GetSessionList(ctx context.Context, redisClient *redis.ClusterClient, sessionIds []uint64) ([]*SessionData) {

	pipeline := redisClient.Pipeline()

	for i := range sessionIds {
		pipeline.Get(ctx, fmt.Sprintf("sd-%016x", sessionIds[i]))
	}

	cmds, err := pipeline.Exec(ctx)
	if err != nil {
		core.Error("failed to get session list: %v", err)
		return nil
	}

	sessionList := make([]*SessionData, 0)

	for i := range sessionIds {

		redis_session_data := cmds[i].(*redis.StringCmd).Val()

		sessionData := SessionData{}
		sessionData.Parse(redis_session_data)

		if sessionData.SessionId != sessionIds[i] {
			continue
		}

		sessionList = append(sessionList, &sessionData)
	}

	return sessionList
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
*/

// ------------------------------------------------------------------------------------------------------------

type NearRelayInserter struct {
	redisClient   *redis.ClusterClient
	lastFlushTime time.Time
	batchSize     int
	numPending    int
	pipeline      redis.Pipeliner
}

func CreateNearRelayInserter(redisClient *redis.ClusterClient, batchSize int) *NearRelayInserter {
	inserter := NearRelayInserter{}
	inserter.redisClient = redisClient
	inserter.lastFlushTime = time.Now()
	inserter.batchSize = batchSize
	inserter.pipeline = redisClient.Pipeline()
	return &inserter
}

func (inserter *NearRelayInserter) Insert(ctx context.Context, sessionId uint64, nearRelayData *NearRelayData) {

	currentTime := time.Now()

	key := fmt.Sprintf("nr-%016x", sessionId)
	inserter.pipeline.RPush(ctx, key, nearRelayData.Value())

	inserter.numPending++

	inserter.CheckForFlush(ctx, currentTime)
}

func (inserter *NearRelayInserter) CheckForFlush(ctx context.Context, currentTime time.Time) {
	if inserter.numPending > inserter.batchSize || currentTime.Sub(inserter.lastFlushTime) >= time.Second {
		inserter.Flush(ctx)
	}
}

func (inserter *NearRelayInserter) Flush(ctx context.Context) {
	_, err := inserter.pipeline.Exec(ctx)
	if err != nil {
		core.Error("near relay insert error: %v", err)
	}
	inserter.numPending = 0
	inserter.lastFlushTime = time.Now()
	inserter.pipeline = inserter.redisClient.Pipeline()
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

func GetServerData(ctx context.Context, redisClient *redis.ClusterClient, serverAddress string, minutes int64) (*ServerData, []*SessionData) {

	pipeline := redisClient.Pipeline()

	pipeline.Get(ctx, fmt.Sprintf("svd-%s", serverAddress))
	pipeline.HGetAll(ctx, fmt.Sprintf("svs-%s-%d", serverAddress, minutes-1))
	pipeline.HGetAll(ctx, fmt.Sprintf("svs-%s-%d", serverAddress, minutes))

	cmds, err := pipeline.Exec(ctx)
	if err != nil {
		core.Error("failed to get server data: %v", err)
		return nil, nil
	}

	redis_server_data := cmds[0].(*redis.StringCmd).Val()

	redis_server_sessions_a := cmds[1].(*redis.MapStringStringCmd).Val()

	redis_server_sessions_b := cmds[2].(*redis.MapStringStringCmd).Val()

	serverData := ServerData{}
	serverData.Parse(redis_server_data)

	if serverData.ServerAddress != serverAddress {
		return nil, nil
	}

	currentTime := uint64(time.Now().Unix())

	sessionMap := make(map[uint64]bool)

	for k,v := range redis_server_sessions_a {
		session_id, _ := strconv.ParseUint(k, 16, 64)
		timestamp, _ := strconv.ParseUint(v, 10, 64)
		if currentTime-timestamp > 30 {
			continue
		}
		sessionMap[session_id] = true
	}

	for k,v := range redis_server_sessions_b {
		session_id, _ := strconv.ParseUint(k, 16, 64)
		timestamp, _ := strconv.ParseUint(v, 10, 64)
		if currentTime-timestamp > 30 {
			continue
		}
		sessionMap[session_id] = true
	}

	serverSessionIds := make([]uint64, len(sessionMap))
	index := 0
	for k := range sessionMap {
		serverSessionIds[index] = k
		index++
	}

	sort.SliceStable(serverSessionIds, func(i, j int) bool { return serverSessionIds[i] < serverSessionIds[j] })

	serverSessionData := GetSessionList(ctx, redisClient, serverSessionIds)

	return &serverData, serverSessionData
}

func GetServerList(ctx context.Context, redisClient *redis.ClusterClient, serverAddresses []string) ([]*ServerData) {

	pipeline := redisClient.Pipeline()

	for i := range serverAddresses {
		pipeline.Get(ctx, fmt.Sprintf("svd-%s", serverAddresses[i]))
	}

	cmds, err := pipeline.Exec(ctx)
	if err != nil {
		core.Error("failed to get server list: %v", err)
		return nil
	}

	serverList := make([]*ServerData, 0)

	for i := range serverAddresses {

		redis_server_data := cmds[i].(*redis.StringCmd).Val()

		serverData := ServerData{}
		serverData.Parse(redis_server_data)

		if serverData.ServerAddress != serverAddresses[i] {
			continue
		}

		serverList = append(serverList, &serverData)
	}

	return serverList
}

func GetServerAddresses(ctx context.Context, redisClient *redis.ClusterClient, minutes int64, begin int, end int) []string {

	if begin < 0 {
		core.Error("invalid begin passed to get server addresses: %d", begin)
		return nil
	}

	if end < 0 {
		core.Error("invalid end passed to get server addresses: %d", end)
		return nil
	}

	if end <= begin {
		core.Error("end must be greater than begin")
		return nil
	}

	// get the set of server addresses in the range [begin,end]

	pipeline := redisClient.Pipeline()

	pipeline.ZRevRangeWithScores(ctx, fmt.Sprintf("sv-%d", minutes-1), int64(begin), int64(end-1))
	pipeline.ZRevRangeWithScores(ctx, fmt.Sprintf("sv-%d", minutes), int64(begin), int64(end-1))

	cmds, err := pipeline.Exec(ctx)
	if err != nil {
		core.Error("failed to get server addresses: %v", err)
		return nil
	}

	redis_server_addresses_a, err := cmds[0].(*redis.ZSliceCmd).Result()
	if err != nil {
		core.Error("failed to get redis server addresses a: %v", err)
		return nil
	}

	redis_server_addresses_b, err := cmds[1].(*redis.ZSliceCmd).Result()
	if err != nil {
		core.Error("failed to get redis server addresses b: %v", err)
		return nil
	}

	serverMap := make(map[string]int32)

	for i := range redis_server_addresses_a {
		address := redis_server_addresses_a[i].Member.(string)
		score := int32(redis_server_addresses_a[i].Score)
		serverMap[address] = score
	}

	for i := range redis_server_addresses_b {
		address := redis_server_addresses_b[i].Member.(string)
		score := int32(redis_server_addresses_b[i].Score)
		serverMap[address] = score
	}

	type ServerEntry struct {
		address string
		score   int32
	}

	serverEntries := make([]ServerEntry, len(serverMap))
	index := 0
	for k,v := range serverMap {
		serverEntries[index] = ServerEntry{k,v}
		index++
	}

	sort.SliceStable(serverEntries, func(i, j int) bool { return serverEntries[i].score > serverEntries[j].score })

	maxSize := end - begin
	if len(serverEntries) > maxSize {
		serverEntries = serverEntries[:maxSize]
	}

	serverAddresses := make([]string, len(serverEntries))
	for i := range serverEntries {
		serverAddresses[i] = serverEntries[i].address
	}

	return serverAddresses
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
