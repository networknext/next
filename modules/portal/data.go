package portal

import (
	"fmt"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"

	"github.com/networknext/backend/modules/constants"
	"github.com/networknext/backend/modules/common"
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
	return
}

type SliceData struct {
	Timestamp uint64
	SliceNumber uint32
	DirectRTT int32
	NextRTT int32
	PredictedRTT int32
	DirectJitter int32
	NextJitter int32
	RealJitter int32
	DirectPacketLoss float32
	NextPacketLoss float32
	RealPacketLoss float32
	RealOutOfOrder float32
	InternalEvents uint64
	SessionEvents uint64
}

func (data * SliceData) Value() string {
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

// todo: parse

func GenerateRandomSliceData() *SliceData {
	data := SliceData{}
	// todo...	
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
		output += fmt.Sprintf("|%d|%d|%.2f", data.NearRelayId[i], data.NearRelayJitter[i], data.NearRelayPacketLoss[i])
	}
	return output
}

func (data *NearRelayData) Parse(key string, value string) {
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
	if len(values) != 2 + int(numNearRelays) * 4 {
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
		data.NearRelayRTT[i] = uint8(common.RandomInt(5,20))
		data.NearRelayJitter[i] = uint8(common.RandomInt(5,20))
		data.NearRelayPacketLoss[i] = float32(common.RandomInt(0,1000000)) / 10000.0
	}
	return &data
}

type SessionData struct {
	SessionId uint64
	ISP string
	ConnectionType int
	PlatformType int
	Latitude float32
	Longitude float32
	DirectRTT int32
	NextRTT int32
	MatchId uint64
	BuyerId uint64
	DatacenterId uint64
	ServerAddress net.UDPAddr
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

// todo: parse

// todo: generate random
