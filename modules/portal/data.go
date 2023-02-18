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
	sessionId uint64
	latitude  float32
	longitude float32
	next      bool
}

func (data *MapData) Value() string {
	nextInt := 0
	if data.next {
		nextInt = 1
	}
	return fmt.Sprintf("%.2f|%.2f|%d", data.latitude, data.longitude, nextInt)
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
	data.sessionId = sessionId
	data.longitude = float32(longitude)
	data.latitude = float32(latitude)
	data.next = values[2] == "1"
	return
}

type SliceData struct {
	timestamp uint64
	sliceNumber uint32
	directRTT int32
	nextRTT int32
	predictedRTT int32
	directJitter int32
	nextJitter int32
	realJitter int32
	directPacketLoss float32
	nextPacketLoss float32
	realPacketLoss float32
	realOutOfOrder float32
	internalEvents uint64
	sessionEvents uint64
}

func (data * SliceData) Value() string {
	return fmt.Sprintf("%x|%d|%d|%d|%d|%d|%d|%d|%.2f|%.2f|%.2f|%.2f|%x|%x",
		data.timestamp,
		data.sliceNumber,
		data.directRTT,
		data.nextRTT,
		data.predictedRTT,
		data.directJitter,
		data.nextJitter,
		data.realJitter,
		data.directPacketLoss,
		data.nextPacketLoss,
		data.realPacketLoss,
		data.realOutOfOrder,
		data.internalEvents,
		data.sessionEvents,
	)
}

// todo: parse

func GenerateRandomSliceData() *SliceData {
	data := SliceData{}
	// todo...	
	return &data
}

type NearRelayData struct {
	timestamp           uint64
	numNearRelays       int
	nearRelayId         [constants.MaxNearRelays]uint64
	nearRelayRTT        [constants.MaxNearRelays]uint8
	nearRelayJitter     [constants.MaxNearRelays]uint8
	nearRelayPacketLoss [constants.MaxNearRelays]float32
}

func (data *NearRelayData) Value() string {
	output := fmt.Sprintf("%x|%d", data.timestamp, data.numNearRelays)
	for i := 0; i < data.numNearRelays; i++ {
		output += fmt.Sprintf("|%d|%d|%.2f", data.nearRelayId[i], data.nearRelayJitter[i], data.nearRelayPacketLoss[i])
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
	data.timestamp = timestamp
	data.numNearRelays = int(numNearRelays)
	for i := 0; i < int(numNearRelays); i++ {
		data.nearRelayId[i] = nearRelayId[i]
		data.nearRelayRTT[i] = uint8(nearRelayRTT[i])
		data.nearRelayJitter[i] = uint8(nearRelayJitter[i])
		data.nearRelayPacketLoss[i] = float32(nearRelayPacketLoss[i])
	}
	return
}

func GenerateRandomNearRelayData() *NearRelayData {
	data := NearRelayData{}
	data.timestamp = uint64(time.Now().Unix())
	data.numNearRelays = constants.MaxNearRelays
	for i := 0; i < data.numNearRelays; i++ {
		data.nearRelayId[i] = rand.Uint64()
		data.nearRelayRTT[i] = uint8(common.RandomInt(5,20))
		data.nearRelayJitter[i] = uint8(common.RandomInt(5,20))
		data.nearRelayPacketLoss[i] = float32(common.RandomInt(0,1000000)) / 10000.0
	}
	return &data
}

type SessionData struct {
	sessionId uint64
	isp string
	connectionType int
	platformType int
	latitude float32
	longitude float32
	directRTT int32
	nextRTT int32
	matchId uint64
	buyerId uint64
	datacenterId uint64
	serverAddress net.UDPAddr
}

func (data *SessionData) Value() string {
	return fmt.Sprintf("%x|%s|%d|%d|%.2f|%.2f|%d|%d|%x|%x|%x|%s",
		data.sessionId,
		data.isp,
		data.connectionType,
		data.platformType,
		data.latitude,
		data.longitude,
		data.directRTT,
		data.nextRTT,
		data.matchId,
		data.buyerId,
		data.datacenterId,
		data.serverAddress.String(),
	)
}

// todo: parse

// todo: generate random
