package portal

import (
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/constants"
	"github.com/networknext/backend/modules/core"
)

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
}

func (data *SliceData) Value() string {
	return fmt.Sprintf("%x|%d|%d|%d|%d|%d|%d|%d|%.2f|%.2f|%.2f|%.2f|%x|%x",
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
	)
}

func (data *SliceData) Parse(value string) {
	values := strings.Split(value, "|")
	if len(values) != 14 {
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
	return &data
}

type NearRelayData struct {
	Timestamp           uint64                           `json:"timestamp"`
	NumNearRelays       int                              `json:"num_near_relays"`
	NearRelayId         [constants.MaxNearRelays]uint64  `json:"near_relay_id"`
	NearRelayRTT        [constants.MaxNearRelays]uint8   `json:"near_relay_rtt"`
	NearRelayJitter     [constants.MaxNearRelays]uint8   `json:"near_relay_jitter"`
	NearRelayPacketLoss [constants.MaxNearRelays]float32 `json:"near_relay_packet_loss"`
}

func (data *NearRelayData) Value() string {
	output := fmt.Sprintf("%x|%d", data.Timestamp, data.NumNearRelays)
	for i := 0; i < data.NumNearRelays; i++ {
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
	data.NumNearRelays = int(numNearRelays)
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
	for i := 0; i < data.NumNearRelays; i++ {
		data.NearRelayId[i] = rand.Uint64()
		data.NearRelayRTT[i] = uint8(common.RandomInt(5, 20))
		data.NearRelayJitter[i] = uint8(common.RandomInt(5, 20))
		data.NearRelayPacketLoss[i] = float32(common.RandomInt(0, 10000)) / 100.0
	}
	return &data
}

type SessionData struct {
	SessionId      uint64      `json:"session_id"`
	ISP            string      `json:"isp"`
	ConnectionType uint8       `json:"connection_type"`
	PlatformType   uint8       `json:"platform_type"`
	Latitude       float32     `json:"latitude"`
	Longitude      float32     `json:"longitude"`
	DirectRTT      uint32      `json:"direct_rtt"`
	NextRTT        uint32      `json:"next_rtt"`
	MatchId        uint64      `json:"match_id"`
	BuyerId        uint64      `json:"buyer_id"`
	DatacenterId   uint64      `json:"datacenter_id"`
	ServerAddress  net.UDPAddr `json:"server_address"`
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
		data.ServerAddress.String(),
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
	serverAddress := core.ParseAddress(values[11])

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
	data.ServerAddress = common.RandomAddress()
	return &data
}

type ServerData struct {
	ServerAddress    net.UDPAddr
	SDKVersion_Major uint8
	SDKVersion_Minor uint8
	SDKVersion_Patch uint8
	MatchId          uint64
	BuyerId          uint64
	DatacenterId     uint64
	NumPlayers       uint32
	StartTime        uint64
}

func (data *ServerData) Value() string {
	return fmt.Sprintf("%s|%d|%d|%d|%x|%x|%x|%d|%x",
		data.ServerAddress.String(),
		data.SDKVersion_Major,
		data.SDKVersion_Minor,
		data.SDKVersion_Patch,
		data.MatchId,
		data.BuyerId,
		data.DatacenterId,
		data.NumPlayers,
		data.StartTime,
	)
}

func (data *ServerData) Parse(value string) {
	values := strings.Split(value, "|")
	if len(values) != 9 {
		return
	}
	serverAddress := core.ParseAddress(values[0])
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
	numPlayers, err := strconv.ParseUint(values[7], 10, 32)
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
	data.NumPlayers = uint32(numPlayers)
	data.StartTime = startTime
}

func GenerateRandomServerData() *ServerData {
	data := ServerData{}
	data.ServerAddress = common.RandomAddress()
	data.SDKVersion_Major = uint8(common.RandomInt(0, 255))
	data.SDKVersion_Minor = uint8(common.RandomInt(0, 255))
	data.SDKVersion_Patch = uint8(common.RandomInt(0, 255))
	data.MatchId = rand.Uint64()
	data.BuyerId = rand.Uint64()
	data.DatacenterId = rand.Uint64()
	data.NumPlayers = rand.Uint32()
	data.StartTime = rand.Uint64()
	return &data
}

type RelayData struct {
	RelayId      uint64
	RelayAddress net.UDPAddr
	DatacenterId uint64
	NumSessions  uint32
	MaxSessions  uint32
	StartTime    uint64
	Version      string
}

func (data *RelayData) Value() string {
	return fmt.Sprintf("%x|%s|%x|%d|%d|%x|%s",
		data.RelayId,
		data.RelayAddress.String(),
		data.DatacenterId,
		data.NumSessions,
		data.MaxSessions,
		data.StartTime,
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
	relayAddress := core.ParseAddress(values[1])
	datacenterId, err := strconv.ParseUint(values[2], 16, 64)
	if err != nil {
		return
	}
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
	version := values[6]
	data.RelayId = relayId
	data.RelayAddress = relayAddress
	data.DatacenterId = datacenterId
	data.NumSessions = uint32(numSessions)
	data.MaxSessions = uint32(maxSessions)
	data.StartTime = startTime
	data.Version = version
}

func GenerateRandomRelayData() *RelayData {
	data := RelayData{}
	data.RelayId = rand.Uint64()
	data.RelayAddress = common.RandomAddress()
	data.DatacenterId = rand.Uint64()
	data.NumSessions = rand.Uint32()
	data.MaxSessions = rand.Uint32()
	data.StartTime = rand.Uint64()
	data.Version = common.RandomString(constants.MaxRelayVersionLength)
	return &data
}
