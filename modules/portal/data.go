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

type MapData struct {
	SessionId uint64
	Latitude  float32
	Longitude float32
	Next      bool
}

func (data *MapData) Value() string {
	nextInt := 0
	if data.Next {
		nextInt = 1
	}
	return fmt.Sprintf("%.2f|%.2f|%d", data.Latitude, data.Longitude, nextInt)
}

func (data *MapData) Parse(key string, value string) {
	sessionId, err := strconv.ParseUint(key[2:], 16, 64)
	if err != nil {
		return
	}
	values := strings.Split(value, "|")
	if len(values) != 3 {
		return
	}
	latitude, err := strconv.ParseFloat(values[0], 32)
	if err != nil {
		return
	}
	longitude, err := strconv.ParseFloat(values[1], 32)
	if err != nil {
		return
	}
	data.SessionId = sessionId
	data.Longitude = float32(longitude)
	data.Latitude = float32(latitude)
	data.Next = values[2] == "1"
}

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
	Timestamp           uint64
	NumNearRelays       int
	NearRelayId         [constants.MaxNearRelays]uint64
	NearRelayRTT        [constants.MaxNearRelays]uint8
	NearRelayJitter     [constants.MaxNearRelays]uint8
	NearRelayPacketLoss [constants.MaxNearRelays]float32
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
	SessionId      uint64
	ISP            string
	ConnectionType uint32
	PlatformType   uint32
	Latitude       float32
	Longitude      float32
	DirectRTT      uint32
	NextRTT        uint32
	MatchId        uint64
	BuyerId        uint64
	DatacenterId   uint64
	ServerAddress  net.UDPAddr
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
	data.ConnectionType = uint32(connectionType)
	data.PlatformType = uint32(platformType)
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
	data.ConnectionType = uint32(common.RandomInt(0, constants.MaxConnectionType))
	data.PlatformType = uint32(common.RandomInt(0, constants.MaxPlatformType))
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
	ServerAddress net.UDPAddr
	SDKVersion_Major uint8
	SDKVersion_Minor uint8
	SDKVersion_Patch uint8
	MatchId uint64
	BuyerId uint64
	DatacenterId uint64
	NumPlayers uint32
	StartTime uint64
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
	data.SDKVersion_Major = uint8(common.RandomInt(0,255))
	data.SDKVersion_Minor = uint8(common.RandomInt(0,255))
	data.SDKVersion_Patch = uint8(common.RandomInt(0,255))
	data.MatchId = rand.Uint64()
	data.BuyerId = rand.Uint64()
	data.DatacenterId = rand.Uint64()
	data.NumPlayers = rand.Uint32()
	data.StartTime = rand.Uint64()
	return &data
}
