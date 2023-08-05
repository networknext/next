package portal

import (
	"fmt"
	"math/rand"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/constants"
	"github.com/networknext/next/modules/core"

	"github.com/gomodule/redigo/redis"
)

// --------------------------------------------------------------------------------------------------

type SliceData struct {
	Timestamp        uint64  `json:"timestamp"`
	SliceNumber      uint32  `json:"slice_number"`
	DirectRTT        uint32  `json:"direct_rtt"`
	NextRTT          uint32  `json:"next_rtt"`
	PredictedRTT     uint32  `json:"predicted_rtt"`
	DirectJitter     uint32  `json:"direct_jitter"`
	NextJitter       uint32  `json:"next_jitter"`
	RealJitter       uint32  `json:"real_jitter"`
	DirectPacketLoss float32 `json:"direct_packet_loss"`
	NextPacketLoss   float32 `json:"next_packet_loss"`
	RealPacketLoss   float32 `json:"real_packet_loss"`
	RealOutOfOrder   float32 `json:"real_out_of_order"`
	InternalEvents   uint64  `json:"internal_events"`
	SessionEvents    uint64  `json:"session_events"`
	DirectKbpsUp     uint32  `json:"direct_kbps_up"`
	DirectKbpsDown   uint32  `json:"direct_kbps_down"`
	NextKbpsUp       uint32  `json:"next_kbps_up"`
	NextKbpsDown     uint32  `json:"next_kbps_down"`
}

func (data *SliceData) Value() string {
	return fmt.Sprintf("%x|%d|%d|%d|%d|%d|%d|%d|%.2f|%.2f|%.2f|%.2f|%x|%x|%d|%d|%d|%d",
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
	)
}

func (data *SliceData) Parse(value string) {
	values := strings.Split(value, "|")
	if len(values) != 18 {
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
	return &data
}

// --------------------------------------------------------------------------------------------------

type NearRelayData struct {
	Timestamp           uint64                           `json:"timestamp"`
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

type SessionData struct {
	SessionId      uint64  `json:"session_id"`
	ISP            string  `json:"isp"`
	ConnectionType uint8   `json:"connection_type"`
	PlatformType   uint8   `json:"platform_type"`
	Latitude       float32 `json:"latitude"`
	Longitude      float32 `json:"longitude"`
	DirectRTT      uint32  `json:"direct_rtt"`
	NextRTT        uint32  `json:"next_rtt"`
	MatchId        uint64  `json:"match_id"`
	BuyerId        uint64  `json:"buyer_id"`
	DatacenterId   uint64  `json:"datacenter_id"`
	ServerAddress  string  `json:"server_address"`
}

func (data *SessionData) Value() string {
	return fmt.Sprintf("%x|%s|%d|%d|%.2f|%.2f|%d|%d|%x|%x|%x|%s",
		data.SessionId,
		data.ISP,
		data.ConnectionType,
		data.PlatformType,
		data.Latitude,
		data.Longitude,
		data.DirectRTT,
		data.NextRTT,
		data.MatchId,
		data.BuyerId,
		data.DatacenterId,
		data.ServerAddress,
	)
}

func (data *SessionData) Parse(value string) {
	values := strings.Split(value, "|")
	if len(values) != 12 {
		return
	}
	sessionId, err := strconv.ParseUint(values[0], 16, 64)
	if err != nil {
		return
	}
	isp := values[1]
	connectionType, err := strconv.ParseUint(values[2], 10, 32)
	if err != nil {
		return
	}
	platformType, err := strconv.ParseUint(values[3], 10, 32)
	if err != nil {
		return
	}
	latitude, err := strconv.ParseFloat(values[4], 32)
	if err != nil {
		return
	}
	longitude, err := strconv.ParseFloat(values[5], 32)
	if err != nil {
		return
	}
	directRTT, err := strconv.ParseUint(values[6], 10, 32)
	if err != nil {
		return
	}
	nextRTT, err := strconv.ParseUint(values[7], 10, 32)
	if err != nil {
		return
	}
	matchId, err := strconv.ParseUint(values[8], 16, 64)
	if err != nil {
		return
	}
	buyerId, err := strconv.ParseUint(values[9], 16, 64)
	if err != nil {
		return
	}
	datacenterId, err := strconv.ParseUint(values[10], 16, 64)
	if err != nil {
		return
	}
	serverAddress := values[11]

	data.SessionId = sessionId
	data.ISP = isp
	data.ConnectionType = uint8(connectionType)
	data.PlatformType = uint8(platformType)
	data.Latitude = float32(latitude)
	data.Longitude = float32(longitude)
	data.DirectRTT = uint32(directRTT)
	data.NextRTT = uint32(nextRTT)
	data.MatchId = matchId
	data.BuyerId = buyerId
	data.DatacenterId = datacenterId
	data.ServerAddress = serverAddress
}

func GenerateRandomSessionData() *SessionData {
	data := SessionData{}
	data.SessionId = rand.Uint64()
	data.ISP = "Comcast Internet Company, LLC"
	data.ConnectionType = uint8(common.RandomInt(0, constants.MaxConnectionType))
	data.PlatformType = uint8(common.RandomInt(0, constants.MaxPlatformType))
	data.Latitude = float32(common.RandomInt(-9000, +9000)) / 100.0
	data.Longitude = float32(common.RandomInt(-18000, +18000)) / 100.0
	data.DirectRTT = rand.Uint32()
	data.NextRTT = rand.Uint32()
	data.MatchId = rand.Uint64()
	data.BuyerId = rand.Uint64()
	data.DatacenterId = rand.Uint64()
	data.ServerAddress = fmt.Sprintf("127.0.0.1:%d", common.RandomInt(1000, 65535))
	return &data
}

// --------------------------------------------------------------------------------------------------

type ServerData struct {
	ServerAddress    string
	SDKVersion_Major uint8
	SDKVersion_Minor uint8
	SDKVersion_Patch uint8
	MatchId          uint64
	BuyerId          uint64
	DatacenterId     uint64
	NumSessions      uint32
	StartTime        uint64
}

func (data *ServerData) Value() string {
	return fmt.Sprintf("%s|%d|%d|%d|%x|%x|%x|%d|%x",
		data.ServerAddress,
		data.SDKVersion_Major,
		data.SDKVersion_Minor,
		data.SDKVersion_Patch,
		data.MatchId,
		data.BuyerId,
		data.DatacenterId,
		data.NumSessions,
		data.StartTime,
	)
}

func (data *ServerData) Parse(value string) {
	values := strings.Split(value, "|")
	if len(values) != 9 {
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
	matchId, err := strconv.ParseUint(values[4], 16, 64)
	if err != nil {
		return
	}
	buyerId, err := strconv.ParseUint(values[5], 16, 64)
	if err != nil {
		return
	}
	datacenterId, err := strconv.ParseUint(values[6], 16, 64)
	if err != nil {
		return
	}
	numSessions, err := strconv.ParseUint(values[7], 10, 32)
	if err != nil {
		return
	}
	startTime, err := strconv.ParseUint(values[8], 16, 64)
	if err != nil {
		return
	}
	data.ServerAddress = serverAddress
	data.SDKVersion_Major = uint8(sdkVersionMajor)
	data.SDKVersion_Minor = uint8(sdkVersionMinor)
	data.SDKVersion_Patch = uint8(sdkVersionPatch)
	data.MatchId = matchId
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
	data.MatchId = rand.Uint64()
	data.BuyerId = rand.Uint64()
	data.DatacenterId = rand.Uint64()
	data.NumSessions = rand.Uint32()
	data.StartTime = rand.Uint64()
	return &data
}

// --------------------------------------------------------------------------------------------------

type RelayData struct {
	RelayId      uint64 `json:"relay_id"`
	RelayAddress string `json:"relay_address"`
	NumSessions  uint32 `json:"num_sessions"`
	MaxSessions  uint32 `json:"max_sessions"`
	StartTime    uint64 `json:"start_time"`
	RelayFlags   uint64 `json:"relay_flags"`
	Version      string `json:"version"`
}

func (data *RelayData) Value() string {
	return fmt.Sprintf("%x|%s|%d|%d|%x|%x|%s",
		data.RelayId,
		data.RelayAddress,
		data.NumSessions,
		data.MaxSessions,
		data.StartTime,
		data.RelayFlags,
		data.Version,
	)
}

func (data *RelayData) Parse(value string) {
	values := strings.Split(value, "|")
	if len(values) != 7 {
		return
	}
	relayId, err := strconv.ParseUint(values[0], 16, 64)
	if err != nil {
		return
	}
	relayAddress := values[1]
	numSessions, err := strconv.ParseUint(values[2], 10, 32)
	if err != nil {
		return
	}
	maxSessions, err := strconv.ParseUint(values[3], 10, 32)
	if err != nil {
		return
	}
	startTime, err := strconv.ParseUint(values[4], 16, 64)
	if err != nil {
		return
	}
	relayFlags, err := strconv.ParseUint(values[5], 16, 64)
	if err != nil {
		return
	}
	version := values[6]
	data.RelayId = relayId
	data.RelayAddress = relayAddress
	data.NumSessions = uint32(numSessions)
	data.MaxSessions = uint32(maxSessions)
	data.StartTime = startTime
	data.RelayFlags = relayFlags
	data.Version = version
}

func GenerateRandomRelayData() *RelayData {
	data := RelayData{}
	data.RelayId = rand.Uint64()
	data.RelayAddress = fmt.Sprintf("127.0.0.1:%d", common.RandomInt(1000, 65535))
	data.NumSessions = rand.Uint32()
	data.MaxSessions = rand.Uint32()
	data.StartTime = rand.Uint64()
	data.RelayFlags = rand.Uint64()
	data.Version = common.RandomString(constants.MaxRelayVersionLength)
	return &data
}

// --------------------------------------------------------------------------------------------------

type RelaySample struct {
	Timestamp                 uint64
	NumSessions               uint32
	EnvelopeBandwidthUpKbps   uint32
	EnvelopeBandwidthDownKbps uint32
	PacketsSentPerSecond      float32
	PacketsReceivedPerSecond  float32
	BandwidthSentKbps         float32
	BandwidthReceivedKbps     float32
	NearPingsPerSecond        float32
	RelayPingsPerSecond       float32
	RelayFlags                uint64
	NumRoutable               uint32
	NumUnroutable             uint32
	CurrentTime               uint64
}

func (data *RelaySample) Value() string {
	return fmt.Sprintf("%x|%d|%d|%d|%.2f|%.2f|%.2f|%.2f|%.2f|%.2f|%x|%d|%d|%x",
		data.Timestamp,
		data.NumSessions,
		data.EnvelopeBandwidthUpKbps,
		data.EnvelopeBandwidthDownKbps,
		data.PacketsSentPerSecond,
		data.PacketsReceivedPerSecond,
		data.BandwidthSentKbps,
		data.BandwidthReceivedKbps,
		data.NearPingsPerSecond,
		data.RelayPingsPerSecond,
		data.RelayFlags,
		data.NumRoutable,
		data.NumUnroutable,
		data.CurrentTime,
	)
}

func (data *RelaySample) Parse(value string) {
	values := strings.Split(value, "|")
	if len(values) != 14 {
		return
	}
	timestamp, err := strconv.ParseUint(values[0], 16, 64)
	if err != nil {
		return
	}
	numSessions, err := strconv.ParseUint(values[1], 10, 32)
	if err != nil {
		return
	}
	envelopeBandwidthUpKbps, err := strconv.ParseUint(values[2], 10, 32)
	if err != nil {
		return
	}
	envelopeBandwidthDownKbps, err := strconv.ParseUint(values[3], 10, 32)
	if err != nil {
		return
	}
	packetsSentPerSecond, err := strconv.ParseFloat(values[4], 32)
	if err != nil {
		return
	}
	packetsReceivedPerSecond, err := strconv.ParseFloat(values[5], 32)
	if err != nil {
		return
	}
	bandwidthSentKbps, err := strconv.ParseFloat(values[6], 32)
	if err != nil {
		return
	}
	bandwidthReceivedKbps, err := strconv.ParseFloat(values[7], 32)
	if err != nil {
		return
	}
	nearPingsPerSecond, err := strconv.ParseFloat(values[8], 32)
	if err != nil {
		return
	}
	relayPingsPerSecond, err := strconv.ParseFloat(values[9], 32)
	if err != nil {
		return
	}
	relayFlags, err := strconv.ParseUint(values[10], 16, 64)
	if err != nil {
		return
	}
	numRoutable, err := strconv.ParseUint(values[11], 10, 32)
	if err != nil {
		return
	}
	numUnroutable, err := strconv.ParseUint(values[12], 10, 32)
	if err != nil {
		return
	}
	currentTime, err := strconv.ParseUint(values[13], 16, 64)
	if err != nil {
		return
	}
	data.Timestamp = timestamp
	data.NumSessions = uint32(numSessions)
	data.EnvelopeBandwidthUpKbps = uint32(envelopeBandwidthUpKbps)
	data.EnvelopeBandwidthDownKbps = uint32(envelopeBandwidthDownKbps)
	data.PacketsSentPerSecond = float32(packetsSentPerSecond)
	data.PacketsReceivedPerSecond = float32(packetsReceivedPerSecond)
	data.BandwidthSentKbps = float32(bandwidthSentKbps)
	data.BandwidthReceivedKbps = float32(bandwidthReceivedKbps)
	data.NearPingsPerSecond = float32(nearPingsPerSecond)
	data.RelayPingsPerSecond = float32(relayPingsPerSecond)
	data.RelayFlags = relayFlags
	data.NumRoutable = uint32(numRoutable)
	data.NumUnroutable = uint32(numUnroutable)
	data.CurrentTime = uint64(currentTime)
}

func GenerateRandomRelaySample() *RelaySample {
	data := RelaySample{}
	data.Timestamp = rand.Uint64()
	data.NumSessions = rand.Uint32()
	data.EnvelopeBandwidthUpKbps = rand.Uint32()
	data.EnvelopeBandwidthDownKbps = rand.Uint32()
	data.PacketsSentPerSecond = float32(common.RandomInt(0, 1000))
	data.PacketsReceivedPerSecond = float32(common.RandomInt(0, 1000))
	data.BandwidthSentKbps = float32(common.RandomInt(0, 1000))
	data.BandwidthReceivedKbps = float32(common.RandomInt(0, 1000))
	data.NearPingsPerSecond = float32(common.RandomInt(0, 1000))
	data.RelayPingsPerSecond = float32(common.RandomInt(0, 1000))
	data.RelayFlags = rand.Uint64()
	data.NumRoutable = rand.Uint32()
	data.NumUnroutable = rand.Uint32()
	data.CurrentTime = rand.Uint64()
	return &data
}

// --------------------------------------------------------------------------------------------------

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

func GetSessions(pool *redis.Pool, minutes int64, begin int, end int) []SessionData {

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

	sessionEntries := make([]SessionEntry, len(sessionsMap))
	index := 0
	for _, v := range sessionsMap {
		sessionEntries[index] = v
		index++
	}

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
		args = args.Add(sessionEntries[i].SessionId)
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

	fmt.Printf("session %016x has %d slices\n", sessionId, len(redis_slice_data))

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
	Score   uint32 `json:"score"`
}

func GetServers(pool *redis.Pool, minutes int64, begin int, end int) []ServerData {

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

	// get the set of server addresses in the range [begin,end]

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
		score, _ := strconv.ParseUint(servers_b[i+1], 10, 32)
		serverMap[servers_b[i]] = ServerEntry{
			Address: address,
			Score:   uint32(score),
		}
	}

	for i := 0; i < len(servers_a); i += 2 {
		address := servers_a[i]
		score, _ := strconv.ParseUint(servers_a[i+1], 10, 32)
		serverMap[servers_a[i]] = ServerEntry{
			Address: address,
			Score:   uint32(score),
		}
	}

	serverEntries := make([]ServerEntry, len(serverMap))
	index := 0
	for _, v := range serverMap {
		serverEntries[index] = v
		index++
	}

	sort.SliceStable(serverEntries, func(i, j int) bool { return serverEntries[i].Score > serverEntries[j].Score })

	maxSize := end - begin
	if len(serverEntries) > maxSize {
		serverEntries = serverEntries[:maxSize]
	}

	// now get server data for the set of server addresses in [begin, end]

	if len(serverEntries) == 0 {
		return nil
	}

	redisClient = pool.Get()

	args := redis.Args{}
	for i := range serverEntries {
		args = args.Add(serverEntries[i].Address)
	}

	redisClient.Send("MGET", args...)

	redisClient.Flush()

	redis_server_data, err := redis.Strings(redisClient.Receive())
	if err != nil {
		core.Error("redis mget get server data failed: %v", err)
		return nil
	}

	redisClient.Close()

	servers := make([]ServerData, len(redis_server_data))

	for i := range servers {
		servers[i].Parse(redis_server_data[i])
		servers[i].ServerAddress = serverEntries[i].Address
	}

	return servers
}

func GetServerData(pool *redis.Pool, serverAddress string, minutes int64) (*ServerData, []uint64) {

	redisClient := pool.Get()

	redisClient.Send("GET", fmt.Sprintf("svd-%s", serverAddress))
	redisClient.Send("HGETALL", fmt.Sprintf("svs-%s-%d", serverAddress, minutes))
	redisClient.Send("HGETALL", fmt.Sprintf("svs-%s-%d", serverAddress, minutes-1))

	redisClient.Flush()

	redis_server_data, err := redis.String(redisClient.Receive())
	if err != nil {
		return nil, nil
	}

	redis_sessions_a, err := redis.Strings(redisClient.Receive())

	redis_sessions_b, err := redis.Strings(redisClient.Receive())

	redisClient.Close()

	serverData := ServerData{}
	serverData.Parse(redis_server_data)

	sessionMap := make(map[uint64]bool)

	currentTime := uint64(time.Now().Unix())

	for i := 0; i < len(redis_sessions_a); i += 2 {
		session_id, _ := strconv.ParseUint(redis_sessions_a[i], 16, 64)
		timestamp, _ := strconv.ParseUint(redis_sessions_a[i+1], 10, 64)
		if currentTime-timestamp > 30 {
			continue
		}
		sessionMap[session_id] = true
	}

	for i := 0; i < len(redis_sessions_b); i += 2 {
		session_id, _ := strconv.ParseUint(redis_sessions_b[i], 16, 64)
		timestamp, _ := strconv.ParseUint(redis_sessions_b[i+1], 10, 64)
		if currentTime-timestamp > 30 {
			continue
		}
		sessionMap[session_id] = true
	}

	serverSessions := make([]uint64, len(sessionMap))
	index := 0
	for k := range sessionMap {
		serverSessions[index] = k
		index++
	}

	sort.SliceStable(serverSessions, func(i, j int) bool { return serverSessions[i] < serverSessions[j] })

	return &serverData, serverSessions
}

// ------------------------------------------------------------------------------------------------------

func GetRelayCount(pool *redis.Pool, minutes int64) int {

	redisClient := pool.Get()

	redisClient.Send("ZCARD", fmt.Sprintf("r-%d", minutes-1))
	redisClient.Send("ZCARD", fmt.Sprintf("r-%d", minutes))

	redisClient.Flush()

	relayCount_a, err := redis.Int(redisClient.Receive())
	if err != nil {
		core.Error("redis get relay count a failed: %v", err)
		return 0
	}

	relayCount_b, err := redis.Int(redisClient.Receive())
	if err != nil {
		core.Error("redis get relay count b failed: %v", err)
		return 0
	}

	redisClient.Close()

	relayCount := relayCount_a
	if relayCount_b > relayCount {
		relayCount = relayCount_b
	}

	return relayCount
}

type RelayEntry struct {
	Address string
	Score   uint32
}

func GetRelays(pool *redis.Pool, minutes int64, begin int, end int) []RelayData {

	if begin < 0 {
		core.Error("invalid begin passed to get relays: %d", begin)
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

	// get the set of relay addresses in order in the range [begin,end]

	redisClient := pool.Get()

	redisClient.Send("ZREVRANGE", fmt.Sprintf("r-%d", minutes-1), begin, end-1, "WITHSCORES")
	redisClient.Send("ZREVRANGE", fmt.Sprintf("r-%d", minutes), begin, end-1, "WITHSCORES")

	redisClient.Flush()

	relays_a, err := redis.Strings(redisClient.Receive())
	if err != nil {
		core.Error("redis get relays a failed: %v", err)
		return nil
	}

	relays_b, err := redis.Strings(redisClient.Receive())
	if err != nil {
		core.Error("redis get relays b failed: %v", err)
		return nil
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

	relayEntries := make([]RelayEntry, len(relayMap))
	index := 0
	for _, v := range relayMap {
		relayEntries[index] = v
		index++
	}

	sort.SliceStable(relayEntries, func(i, j int) bool { return relayEntries[i].Score > relayEntries[j].Score })

	maxSize := end - begin
	if len(relayEntries) > maxSize {
		relayEntries = relayEntries[:maxSize]
	}

	// now get relay data for the set of relay addresses in [begin, end]

	if len(relayEntries) == 0 {
		return nil
	}

	redisClient = pool.Get()

	args := redis.Args{}
	for i := range relayEntries {
		args = args.Add(fmt.Sprintf("rd-%s", relayEntries[i].Address))
	}

	redisClient.Send("MGET", args...)

	redisClient.Flush()

	redis_relay_data, err := redis.Strings(redisClient.Receive())
	if err != nil {
		core.Error("redis mget get relay data failed: %v", err)
		return nil
	}

	redisClient.Close()

	relays := make([]RelayData, len(redis_relay_data))

	for i := range relays {
		relays[i].Parse(redis_relay_data[i])
		relays[i].RelayAddress = relayEntries[i].Address
	}

	return relays
}

func GetRelayData(pool *redis.Pool, relayAddress string) (*RelayData, []RelaySample) {

	redisClient := pool.Get()

	redisClient.Send("GET", fmt.Sprintf("rd-%s", relayAddress))
	redisClient.Send("LRANGE", fmt.Sprintf("rs-%s", relayAddress), 0, -1)

	redisClient.Flush()

	redis_relay_data, err := redis.String(redisClient.Receive())
	if err != nil {
		return nil, nil
	}

	redis_sample_data, err := redis.Strings(redisClient.Receive())
	if err != nil {
		return nil, nil
	}

	redisClient.Close()

	relayData := RelayData{}
	relayData.Parse(redis_relay_data)

	relaySamples := make([]RelaySample, len(redis_sample_data))
	for i := 0; i < len(redis_sample_data); i++ {
		relaySamples[i].Parse(redis_sample_data[i])
	}

	return &relayData, relaySamples
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

	sessionIdString := fmt.Sprintf("%016x", sessionId)

	key := fmt.Sprintf("sd-%s", sessionIdString)
	inserter.redisClient.Send("SET", key, sessionData.Value())
	inserter.redisClient.Send("EXPIRE", key, 30)

	key = fmt.Sprintf("sl-%s", sessionIdString)
	inserter.redisClient.Send("RPUSH", key, sliceData.Value())
	inserter.redisClient.Send("EXPIRE", key, 30)

	key = fmt.Sprintf("svs-%s-%d", sessionData.ServerAddress, minutes)
	inserter.redisClient.Send("HSET", key, sessionIdString, currentTime.Unix())
	inserter.redisClient.Send("EXPIRE", key, 30)

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

	key := fmt.Sprintf("nr-%016x", sessionId)
	inserter.redisClient.Send("RPUSH", key, nearRelayData.Value())
	inserter.redisClient.Send("EXPIRE", key, 3600)

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

	serverId := common.HashString(serverData.ServerAddress)

	score := uint32(serverId) ^ uint32(serverId>>32)

	inserter.servers = inserter.servers.Add(score)
	inserter.servers = inserter.servers.Add(serverData.ServerAddress)

	inserter.redisClient.Send("SET", fmt.Sprintf("svd-%s", serverData.ServerAddress), serverData.Value())
	inserter.redisClient.Send("EXPIRE", fmt.Sprintf("svd-%s", serverData.ServerAddress), 30)

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

func (inserter *RelayInserter) Insert(relayData *RelayData, relaySample *RelaySample) {

	currentTime := time.Now()

	minutes := currentTime.Unix() / 60

	if len(inserter.relays) == 0 {
		inserter.relays = redis.Args{}.Add(fmt.Sprintf("r-%d", minutes))
	}

	score := uint32(relayData.RelayId) ^ uint32(relayData.RelayId>>32)

	inserter.relays = inserter.relays.Add(score)
	inserter.relays = inserter.relays.Add(relayData.RelayAddress)

	key := fmt.Sprintf("rd-%s", relayData.RelayAddress)
	inserter.redisClient.Send("SET", key, relayData.Value())
	inserter.redisClient.Send("EXPIRE", key, "30")

	key = fmt.Sprintf("rs-%s", relayData.RelayAddress)
	inserter.redisClient.Send("RPUSH", key, relaySample.Value())
	inserter.redisClient.Send("LTRIM", key, "-3600", "-1")
	inserter.redisClient.Send("EXPIRE", key, "3600")

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
