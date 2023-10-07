package portal

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/constants"
	"github.com/networknext/next/modules/core"
	"github.com/networknext/next/modules/encoding"

	"github.com/redis/go-redis/v9"
)

// --------------------------------------------------------------------------------------------------

type SessionData struct {
	SessionId      uint64                           `json:"session_id,string"`
	UserHash       uint64                           `json:"user_hash,string"`
	StartTime      uint64                           `json:"start_time,string"`
	ISP            string                           `json:"isp"`
	ConnectionType uint8                            `json:"connection_type"`
	PlatformType   uint8                            `json:"platform_type"`
	Latitude       float32                          `json:"latitude"`
	Longitude      float32                          `json:"longitude"`
	DirectRTT      uint32                           `json:"direct_rtt"`
	NextRTT        uint32                           `json:"next_rtt"`
	BuyerId        uint64                           `json:"buyer_id,string"`
	DatacenterId   uint64                           `json:"datacenter_id,string"`
	ServerAddress  string                           `json:"server_address"`
	NumRouteRelays int                              `json:"num_route_relays"`
	RouteRelays    [constants.MaxRouteRelays]uint64 `json:"route_relays,string"`
}

func (data *SessionData) Value() string {
	value := fmt.Sprintf("%x|%x|%x|%s|%d|%d|%.2f|%.2f|%d|%d|%x|%x|%s|%d|",
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
		data.NumRouteRelays,
	)
	for i := 0; i < data.NumRouteRelays; i++ {
		value += fmt.Sprintf("%x|", data.RouteRelays[i])
	}
	return value
}

func (data *SessionData) Parse(value string) {
	values := strings.Split(value, "|")
	if len(values) < 14 {
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
	numRouteRelays, err := strconv.ParseUint(values[13], 10, 32)
	if err != nil {
		return
	}
	if len(values) != 14+int(numRouteRelays)+1 {
		return
	}
	routeRelays := make([]uint64, numRouteRelays)
	for i := range routeRelays {
		routeRelays[i], err = strconv.ParseUint(values[14+i], 16, 64)
		if err != nil {
			return
		}
	}
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
	data.NumRouteRelays = int(numRouteRelays)
	copy(data.RouteRelays[:], routeRelays)
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
	data.NumRouteRelays = common.RandomInt(0, constants.MaxRouteRelays-1)
	for i := 0; i < data.NumRouteRelays; i++ {
		data.RouteRelays[i] = rand.Uint64()
	}
	return &data
}

// --------------------------------------------------------------------------------------------------

type SliceData struct {
	Timestamp        uint64  `json:"timestamp,string"`
	SliceNumber      uint32  `json:"session_id"`
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
	InternalEvents   uint64  `json:"internal_events,string"`
	SessionEvents    uint64  `json:"session_events,string"`
	DirectKbpsUp     uint32  `json:"direct_kbps_up"`
	DirectKbpsDown   uint32  `json:"direct_kbps_down"`
	NextKbpsUp       uint32  `json:"next_kbps_up"`
	NextKbpsDown     uint32  `json:"next_kbps_down"`
	Next             bool    `json:"next"`
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
	ServerAddress    string `json:"server_address"`
	SDKVersion_Major uint8  `json:"sdk_version_major"`
	SDKVersion_Minor uint8  `json:"sdk_version_minor"`
	SDKVersion_Patch uint8  `json:"sdk_version_patch"`
	BuyerId          uint64 `json:"buyer_id,string"`
	DatacenterId     uint64 `json:"datacenter_id,string"`
	NumSessions      uint32 `json:"num_sessions"`
	Uptime           uint64 `json:"uptime,string"`
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
		data.Uptime,
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
	uptime, err := strconv.ParseUint(values[7], 16, 64)
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
	data.Uptime = uptime
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
	data.Uptime = rand.Uint64()
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

type RelaySample struct {
	Timestamp                 uint64  `json:"timestamp,string"`
	NumSessions               uint32  `json:"num_sessions"`
	EnvelopeBandwidthUpKbps   uint32  `json:"envelope_bandwidth_up_kbps"`
	EnvelopeBandwidthDownKbps uint32  `json:"envelope_bandwidth_down_kbps"`
	PacketsSentPerSecond      float32 `json:"packets_sent_per_second"`
	PacketsReceivedPerSecond  float32 `json:"packets_recieved_per_second"`
	BandwidthSentKbps         float32 `json:"bandwidth_sent_kbps"`
	BandwidthReceivedKbps     float32 `json:"bandwidth_received_kbps"`
	NearPingsPerSecond        float32 `json:"near_pings_per_second"`
	RelayPingsPerSecond       float32 `json:"relay_pings_per_second"`
	RelayFlags                uint64  `json:"relay_flags,string"`
	NumRoutable               uint32  `json:"num_routable"`
	NumUnroutable             uint32  `json:"num_unroutable"`
	CurrentTime               uint64  `json:"current_time,string"`
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

// ------------------------------------------------------------------------------------------------------------

const NumBuckets = 1000

type SessionCruncherEntry struct {
	SessionId uint64
	BuyerId   uint64
	Score     uint32
	Next      uint8
	Latitude  float32
	Longitude float32
}

type SessionCruncherPublisherConfig struct {
	URL                string
	BatchSize          int
	BatchDuration      time.Duration
	MessageChannelSize int
}

type SessionCruncherPublisher struct {
	MessageChannel    chan SessionCruncherEntry
	config            SessionCruncherPublisherConfig
	batchStartTime    time.Time
	mutex             sync.RWMutex
	lastBatchSendTime time.Time
	batchMessages     []SessionCruncherEntry
	numMessagesSent   int
	numBatchesSent    int
}

func CreateSessionCruncherPublisher(ctx context.Context, config SessionCruncherPublisherConfig) *SessionCruncherPublisher {

	if config.MessageChannelSize == 0 {
		config.MessageChannelSize = 1024 * 1024
	}

	if config.BatchDuration == 0 {
		config.BatchDuration = time.Second
	}

	if config.BatchSize == 0 {
		config.BatchSize = 10000
	}

	publisher := &SessionCruncherPublisher{}

	publisher.config = config
	publisher.MessageChannel = make(chan SessionCruncherEntry, config.MessageChannelSize)

	go publisher.updateMessageChannel(ctx)

	return publisher
}

const SessionBatchVersion_Write = uint64(0)

func (publisher *SessionCruncherPublisher) updateMessageChannel(ctx context.Context) {

	for {
		select {

		case <-ctx.Done():
			return

		case message := <-publisher.MessageChannel:
			publisher.mutex.Lock()
			publisher.numMessagesSent++
			publisher.batchMessages = append(publisher.batchMessages, message)
			if len(publisher.batchMessages) >= publisher.config.BatchSize || (len(publisher.batchMessages) > 0 && time.Since(publisher.lastBatchSendTime) >= publisher.config.BatchDuration) {
				publisher.sendBatch()
			}
			publisher.mutex.Unlock()
		}
	}
}

func (publisher *SessionCruncherPublisher) sendBatch() {

	batchSize := [NumBuckets]uint32{}

	for i := range publisher.batchMessages {
		batchSize[publisher.batchMessages[i].Score]++
	}

	batch := make([][]SessionCruncherEntry, NumBuckets)

	for i := range batchSize {
		batch[i] = make([]SessionCruncherEntry, 0, batchSize[i])
	}

	for i := range publisher.batchMessages {
		index := int(publisher.batchMessages[i].Score)
		batch[index] = append(batch[index], publisher.batchMessages[i])
	}

	size := 8 + 4*NumBuckets
	for i := range batchSize {
		size += int(batchSize[i]) * (8 + 8 + 1 + 4 + 4)
	}

	data := make([]byte, size)

	index := 0

	encoding.WriteUint64(data[:], &index, SessionBatchVersion_Write)

	for i := 0; i < NumBuckets; i++ {
		encoding.WriteUint32(data[:], &index, uint32(batchSize[i]))
		for j := range batch[i] {
			encoding.WriteUint64(data[:], &index, batch[i][j].SessionId)
			encoding.WriteUint64(data[:], &index, batch[i][j].BuyerId)
			encoding.WriteUint8(data[:], &index, batch[i][j].Next)
			encoding.WriteFloat32(data[:], &index, batch[i][j].Latitude)
			encoding.WriteFloat32(data[:], &index, batch[i][j].Longitude)
		}
	}

	err := postBinary(publisher.config.URL, data)

	if err != nil {
		core.Error("failed to post session cruncher batch: %v", err)
	}

	publisher.batchMessages = publisher.batchMessages[:0]
	publisher.numBatchesSent++
	publisher.lastBatchSendTime = time.Now()
}

func postBinary(url string, data []byte) error {

	buffer := bytes.NewBuffer(data)

	request, _ := http.NewRequest("POST", url, buffer)

	request.Header.Add("Content-Type", "application/octet-stream")

	httpClient := &http.Client{}
	response, err := httpClient.Do(request)
	if err != nil {
		return err
	}

	if response.StatusCode != 200 {
		return fmt.Errorf("got response %d", response.StatusCode)
	}

	body, error := ioutil.ReadAll(response.Body)
	if error != nil {
		return fmt.Errorf("could not read response: %v", err)
	}

	response.Body.Close()

	_ = body

	return nil
}

func (publisher *SessionCruncherPublisher) NumMessagesSent() int {
	publisher.mutex.RLock()
	value := publisher.numMessagesSent
	publisher.mutex.RUnlock()
	return value
}

func (publisher *SessionCruncherPublisher) NumBatchesSent() int {
	publisher.mutex.RLock()
	value := publisher.numBatchesSent
	publisher.mutex.RUnlock()
	return value
}

// ------------------------------------------------------------------------------------------------------------

type SessionInserter struct {
	redisClient   redis.Cmdable
	lastFlushTime time.Time
	batchSize     int
	numPending    int
	pipeline      redis.Pipeliner
	publisher     *SessionCruncherPublisher
}

func CreateSessionInserter(ctx context.Context, redisClient redis.Cmdable, sessionCruncherURL string, batchSize int) *SessionInserter {
	inserter := SessionInserter{}
	inserter.redisClient = redisClient
	inserter.lastFlushTime = time.Now()
	inserter.batchSize = batchSize
	inserter.pipeline = redisClient.Pipeline()
	inserter.publisher = CreateSessionCruncherPublisher(ctx, SessionCruncherPublisherConfig{URL: sessionCruncherURL + "/session_batch", BatchSize: batchSize})
	return &inserter
}

func (inserter *SessionInserter) Insert(ctx context.Context, sessionId uint64, userHash uint64, next bool, score uint32, sessionData *SessionData, sliceData *SliceData) {

	currentTime := time.Now()

	minutes := currentTime.Unix() / 60

	entry := SessionCruncherEntry{
		SessionId: sessionId,
		BuyerId:   sessionData.BuyerId,
		Score:     score,
		Latitude:  sessionData.Latitude,
		Longitude: sessionData.Longitude,
	}

	if next {
		entry.Next = 1
	}

	inserter.publisher.MessageChannel <- entry

	sessionIdString := fmt.Sprintf("%016x", sessionId)

	key := fmt.Sprintf("sd-%s", sessionIdString)
	inserter.pipeline.Set(ctx, key, sessionData.Value(), 0)

	key = fmt.Sprintf("sl-%s", sessionIdString)
	inserter.pipeline.RPush(ctx, key, sliceData.Value())

	key = fmt.Sprintf("u-%016x-%d", userHash, minutes)
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

const TopSessionsVersion = uint64(0)

type TopSessionsWatcher struct {
	url           string
	mutex         sync.RWMutex
	nextSessions  int
	totalSessions int
	topSessions   []uint64
}

func CreateTopSessionsWatcher(sessionCruncherURL string) *TopSessionsWatcher {
	watcher := TopSessionsWatcher{}
	watcher.url = sessionCruncherURL + "/top_sessions"
	go watcher.watchTopSessions()
	return &watcher
}

func (watcher *TopSessionsWatcher) watchTopSessions() {
	ticker := time.NewTicker(time.Second)
	for {
		select {

		case <-ticker.C:

			data := getBinary(watcher.url)

			if data == nil {
				break
			}

			if len(data) < 8+4+4+4 {
				core.Error("top session response is too small")
				break
			}

			index := 0

			var version uint64
			encoding.ReadUint64(data[:], &index, &version)
			if version != TopSessionsVersion {
				core.Error("bad top sessions version. expected %d, got %d", version, TopSessionsVersion)
				break
			}

			var nextSessions, totalSessions uint32
			encoding.ReadUint32(data[:], &index, &nextSessions)
			encoding.ReadUint32(data[:], &index, &totalSessions)

			numSessions := (len(data) - (8 + 4 + 4 + 4)) / 8
			sessions := make([]uint64, numSessions)
			for i := 0; i < numSessions; i++ {
				encoding.ReadUint64(data[:], &index, &sessions[i])
			}

			watcher.mutex.Lock()
			watcher.nextSessions = int(nextSessions)
			watcher.totalSessions = int(totalSessions)
			watcher.topSessions = sessions
			watcher.mutex.Unlock()
		}
	}
}

func getBinary(url string) []byte {

	var err error
	var response *http.Response
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(nil))

	client := &http.Client{}

	response, err = client.Do(req)

	if err != nil {
		core.Error("failed to read %s: %v", url, err)
		return nil
	}

	if response == nil {
		core.Error("no response from %s", url)
		return nil
	}

	if response.StatusCode != 200 {
		core.Error("got %d response for %s", response.StatusCode, url)
		return nil
	}

	body, error := ioutil.ReadAll(response.Body)
	if error != nil {
		core.Error("could not read response body for %s: %v", url, err)
		return nil
	}

	response.Body.Close()

	return body
}

func (watcher *TopSessionsWatcher) GetSessionCounts() (int, int) {
	watcher.mutex.RLock()
	next := watcher.nextSessions
	total := watcher.totalSessions
	watcher.mutex.RUnlock()
	return next, total
}

func (watcher *TopSessionsWatcher) GetSessions(begin int, end int) []uint64 {
	if begin < 0 {
		return nil
	}
	if end <= begin {
		return nil
	}
	watcher.mutex.RLock()
	sessions := watcher.topSessions
	watcher.mutex.RUnlock()
	if end >= len(sessions) {
		end = len(sessions)
	}
	return sessions[begin:end]
}

func (watcher *TopSessionsWatcher) GetTopSessions() []uint64 {
	watcher.mutex.RLock()
	sessions := watcher.topSessions
	watcher.mutex.RUnlock()
	return sessions
}

// --------------------------------------------------------------------------------------------------

const BuyerSessionDataVersion = uint64(0)
const BuyerServerDataVersion = uint64(0)

type BuyerDataWatcher struct {
	sessionUrl     string
	serverUrl      string
	mutex          sync.RWMutex
	buyerIds       []uint64
	buyerIdToIndex map[uint64]int
	nextSessions   []uint32
	totalSessions  []uint32
	serverCounts   []uint32
}

func CreateBuyerDataWatcher(sessionCruncherURL string, serverCruncherURL string) *BuyerDataWatcher {
	watcher := BuyerDataWatcher{}
	watcher.sessionUrl = sessionCruncherURL + "/buyer_data"
	watcher.serverUrl = serverCruncherURL + "/buyer_data"
	go watcher.watchBuyerData()
	return &watcher
}

func (watcher *BuyerDataWatcher) watchBuyerData() {
	ticker := time.NewTicker(time.Second)
	for {
		select {

		case <-ticker.C:

			// get buyer data from *both* the session cruncher (total/next session counts, accelerated % per-buyer), and the server cruncher (server count per-buyer)

			sessionData := getBinary(watcher.sessionUrl)

			serverData := getBinary(watcher.serverUrl)

			// process session buyer data

			if sessionData == nil {
				break
			}

			if len(sessionData) < 8+4 {
				core.Error("session buyer data response is too small")
				break
			}

			index := 0

			var version uint64
			encoding.ReadUint64(sessionData[:], &index, &version)
			if version != BuyerSessionDataVersion {
				core.Error("bad session buyer data version. expected %d, got %d", version, BuyerSessionDataVersion)
				break
			}

			var numBuyers uint32
			encoding.ReadUint32(sessionData[:], &index, &numBuyers)
			if numBuyers > constants.MaxBuyers {
				core.Error("too many session buyers. got %d, max is %d", numBuyers, constants.MaxBuyers)
				break
			}

			buyerIds := make([]uint64, numBuyers)
			buyerIdToIndex := make(map[uint64]int, len(buyerIds))
			totalSessions := make([]uint32, numBuyers)
			nextSessions := make([]uint32, numBuyers)

			for i := 0; i < int(numBuyers); i++ {
				encoding.ReadUint64(sessionData[:], &index, &buyerIds[i])
				encoding.ReadUint32(sessionData[:], &index, &totalSessions[i])
				encoding.ReadUint32(sessionData[:], &index, &nextSessions[i])
				buyerIdToIndex[buyerIds[i]] = i
			}

			if len(buyerIdToIndex) != len(buyerIds) {
				core.Error("duplicate buyer id detected")
				break
			}

			// process the server buyer data

			serverCounts := make([]uint32, len(buyerIds))
			{
				if serverData == nil {
					break
				}

				if len(serverData) < 8+4 {
					core.Error("server buyer data response is too small")
					break
				}

				index := 0

				var version uint64
				encoding.ReadUint64(serverData[:], &index, &version)
				if version != BuyerServerDataVersion {
					core.Error("bad server buyer data version. expected %d, got %d", version, BuyerServerDataVersion)
					break
				}

				var serverNumBuyers uint32
				encoding.ReadUint32(serverData[:], &index, &serverNumBuyers)
				if serverNumBuyers > constants.MaxBuyers {
					core.Error("too many server buyers. got %d, max is %d", serverNumBuyers, constants.MaxBuyers)
					break
				}

				serverBuyerIds := make([]uint64, serverNumBuyers)
				serverBuyerServerCounts := make([]uint32, serverNumBuyers)

				for i := 0; i < int(serverNumBuyers); i++ {
					encoding.ReadUint64(serverData[:], &index, &serverBuyerIds[i])
					encoding.ReadUint32(serverData[:], &index, &serverBuyerServerCounts[i])
				}

				for i := 0; i < int(serverNumBuyers); i++ {
					index, found := buyerIdToIndex[serverBuyerIds[i]]
					if found {
						serverCounts[index] = serverBuyerServerCounts[i]
					}
				}
			}

			// stash the processed data

			watcher.mutex.Lock()
			watcher.buyerIds = buyerIds
			watcher.buyerIdToIndex = buyerIdToIndex
			watcher.totalSessions = totalSessions
			watcher.nextSessions = nextSessions
			watcher.serverCounts = serverCounts
			watcher.mutex.Unlock()
		}
	}
}

func (watcher *BuyerDataWatcher) GetBuyerData() (buyerIds []uint64, buyerIdToIndex map[uint64]int, totalSessions []uint32, nextSessions []uint32, serverCounts []uint32) {
	watcher.mutex.RLock()
	buyerIds = watcher.buyerIds
	buyerIdToIndex = watcher.buyerIdToIndex
	totalSessions = watcher.totalSessions
	nextSessions = watcher.nextSessions
	serverCounts = watcher.serverCounts
	watcher.mutex.RUnlock()
	return
}

// --------------------------------------------------------------------------------------------------

const MapDataVersion = uint64(0)

type MapDataWatcher struct {
	url     string
	mutex   sync.RWMutex
	mapData []byte
}

func CreateMapDataWatcher(sessionCruncherURL string) *MapDataWatcher {
	watcher := MapDataWatcher{}
	watcher.url = sessionCruncherURL + "/map_data"
	go watcher.watchMapData()
	return &watcher
}

func (watcher *MapDataWatcher) watchMapData() {
	ticker := time.NewTicker(time.Second)
	for {
		select {

		case <-ticker.C:

			data := getBinary(watcher.url)

			if data == nil {
				break
			}

			if len(data) < 8+4 {
				core.Error("map data response is too small")
				break
			}

			watcher.mutex.Lock()
			watcher.mapData = data
			watcher.mutex.Unlock()
		}
	}
}

func (watcher *MapDataWatcher) GetMapData() []byte {
	watcher.mutex.RLock()
	data := watcher.mapData
	watcher.mutex.RUnlock()
	return data
}

// --------------------------------------------------------------------------------------------------

func GetSessionData(ctx context.Context, redisClient redis.Cmdable, sessionId uint64) (*SessionData, []SliceData, []NearRelayData) {

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

func GetSessionList(ctx context.Context, redisClient redis.Cmdable, sessionIds []uint64) []*SessionData {

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

func GetUserSessionList(ctx context.Context, redisClient redis.Cmdable, userHash uint64, minutes int64, max int) []*SessionData {

	if max < 1 {
		core.Error("invalid max passed in to get user session list: %d", max)
		return nil
	}

	// get user session ids in order in the range [0,max)

	pipeline := redisClient.Pipeline()

	pipeline.ZRevRangeWithScores(ctx, fmt.Sprintf("u-%016x-%d", userHash, minutes-1), 0, int64(max-1))
	pipeline.ZRevRangeWithScores(ctx, fmt.Sprintf("u-%016x-%d", userHash, minutes), 0, int64(max-1))

	cmds, err := pipeline.Exec(ctx)
	if err != nil {
		core.Error("failed to get user sessions: %v", err)
		return nil
	}

	redis_user_sessions_a, err := cmds[0].(*redis.ZSliceCmd).Result()
	if err != nil {
		core.Error("failed to get redis user sessions a: %v", err)
		return nil
	}

	redis_user_sessions_b, err := cmds[1].(*redis.ZSliceCmd).Result()
	if err != nil {
		core.Error("failed to get redis user sessions b: %v", err)
		return nil
	}

	sessionMap := make(map[uint64]uint64)

	for i := range redis_user_sessions_a {
		sessionId, _ := strconv.ParseUint(redis_user_sessions_a[i].Member.(string), 16, 64)
		score := uint64(redis_user_sessions_a[i].Score)
		sessionMap[sessionId] = score
	}

	for i := range redis_user_sessions_b {
		sessionId, _ := strconv.ParseUint(redis_user_sessions_b[i].Member.(string), 16, 64)
		score := uint64(redis_user_sessions_b[i].Score)
		sessionMap[sessionId] = score
	}

	type SessionEntry struct {
		sessionId uint64
		score     uint64
	}

	sessionEntries := make([]SessionEntry, len(sessionMap))
	index := 0
	for k, v := range sessionMap {
		sessionEntries[index].sessionId = k
		sessionEntries[index].score = v
		index++
	}

	sort.Slice(sessionEntries, func(i, j int) bool { return sessionEntries[i].sessionId < sessionEntries[j].sessionId })
	sort.SliceStable(sessionEntries, func(i, j int) bool { return sessionEntries[i].score > sessionEntries[j].score })

	if len(sessionEntries) > max {
		sessionEntries = sessionEntries[:max]
	}

	userSessionIds := make([]uint64, len(sessionEntries))
	index = 0
	for i := range sessionEntries {
		userSessionIds[index] = sessionEntries[i].sessionId
		index++
	}

	userSessions := GetSessionList(ctx, redisClient, userSessionIds)

	return userSessions
}

// ------------------------------------------------------------------------------------------------------------

type NearRelayInserter struct {
	redisClient   redis.Cmdable
	lastFlushTime time.Time
	batchSize     int
	numPending    int
	pipeline      redis.Pipeliner
}

func CreateNearRelayInserter(redisClient redis.Cmdable, batchSize int) *NearRelayInserter {
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

const MaxServerAddressLength = 64

type ServerCruncherEntry struct {
	ServerAddress string
	BuyerId       uint64
	Score         uint32
}

type ServerCruncherPublisherConfig struct {
	URL                string
	BatchSize          int
	BatchDuration      time.Duration
	MessageChannelSize int
}

type ServerCruncherPublisher struct {
	MessageChannel    chan ServerCruncherEntry
	config            ServerCruncherPublisherConfig
	batchStartTime    time.Time
	mutex             sync.RWMutex
	lastBatchSendTime time.Time
	batchMessages     []ServerCruncherEntry
	numMessagesSent   int
	numBatchesSent    int
}

func CreateServerCruncherPublisher(ctx context.Context, config ServerCruncherPublisherConfig) *ServerCruncherPublisher {

	if config.MessageChannelSize == 0 {
		config.MessageChannelSize = 1024 * 1024
	}

	if config.BatchDuration == 0 {
		config.BatchDuration = time.Second
	}

	if config.BatchSize == 0 {
		config.BatchSize = 10000
	}

	publisher := &ServerCruncherPublisher{}

	publisher.config = config
	publisher.MessageChannel = make(chan ServerCruncherEntry, config.MessageChannelSize)

	go publisher.updateMessageChannel(ctx)

	return publisher
}

const ServerBatchVersion_Write = uint64(0)

func (publisher *ServerCruncherPublisher) updateMessageChannel(ctx context.Context) {

	for {
		select {

		case <-ctx.Done():
			return

		case message := <-publisher.MessageChannel:
			publisher.mutex.Lock()
			publisher.numMessagesSent++
			publisher.batchMessages = append(publisher.batchMessages, message)
			if len(publisher.batchMessages) >= publisher.config.BatchSize || (len(publisher.batchMessages) > 0 && time.Since(publisher.lastBatchSendTime) >= publisher.config.BatchDuration) {
				publisher.sendBatch()
			}
			publisher.mutex.Unlock()
		}
	}
}

func (publisher *ServerCruncherPublisher) sendBatch() {

	batchSize := [NumBuckets]uint32{}

	for i := range publisher.batchMessages {
		batchSize[publisher.batchMessages[i].Score]++
	}

	batch := make([][]ServerCruncherEntry, NumBuckets)

	for i := range batchSize {
		batch[i] = make([]ServerCruncherEntry, 0, batchSize[i])
	}

	for i := range publisher.batchMessages {
		index := int(publisher.batchMessages[i].Score)
		batch[index] = append(batch[index], publisher.batchMessages[i])
	}

	size := 8 + 4*NumBuckets
	for i := range batchSize {
		size += 4 + int(batchSize[i])*(MaxServerAddressLength+8)
	}

	data := make([]byte, size)

	index := 0

	encoding.WriteUint64(data[:], &index, ServerBatchVersion_Write)

	for i := 0; i < NumBuckets; i++ {
		encoding.WriteUint32(data[:], &index, uint32(batchSize[i]))
		for j := range batch[i] {
			encoding.WriteString(data, &index, batch[i][j].ServerAddress, MaxServerAddressLength)
			encoding.WriteUint64(data, &index, batch[i][j].BuyerId)
		}
	}

	data = data[:index]

	err := postBinary(publisher.config.URL, data)

	if err != nil {
		core.Error("failed to post server cruncher batch: %v", err)
	}

	publisher.batchMessages = publisher.batchMessages[:0]
	publisher.numBatchesSent++
	publisher.lastBatchSendTime = time.Now()
}

func (publisher *ServerCruncherPublisher) NumMessagesSent() int {
	publisher.mutex.RLock()
	value := publisher.numMessagesSent
	publisher.mutex.RUnlock()
	return value
}

func (publisher *ServerCruncherPublisher) NumBatchesSent() int {
	publisher.mutex.RLock()
	value := publisher.numBatchesSent
	publisher.mutex.RUnlock()
	return value
}

// ------------------------------------------------------------------------------------------------------------

type ServerInserter struct {
	redisClient   redis.Cmdable
	lastFlushTime time.Time
	batchSize     int
	numPending    int
	pipeline      redis.Pipeliner
	publisher     *ServerCruncherPublisher
}

func CreateServerInserter(ctx context.Context, redisClient redis.Cmdable, serverCruncherURL string, batchSize int) *ServerInserter {
	inserter := ServerInserter{}
	inserter.redisClient = redisClient
	inserter.lastFlushTime = time.Now()
	inserter.batchSize = batchSize
	inserter.pipeline = redisClient.Pipeline()
	inserter.publisher = CreateServerCruncherPublisher(ctx, ServerCruncherPublisherConfig{URL: serverCruncherURL + "/server_batch", BatchSize: batchSize})
	return &inserter
}

func (inserter *ServerInserter) Insert(ctx context.Context, serverData *ServerData) {

	currentTime := time.Now()

	serverId := common.HashString(serverData.ServerAddress)

	score := (uint32(serverId) ^ uint32(serverId>>32)) % NumBuckets

	entry := ServerCruncherEntry{
		ServerAddress: serverData.ServerAddress,
		BuyerId:       serverData.BuyerId,
		Score:         score,
	}

	inserter.publisher.MessageChannel <- entry

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

func GetServerData(ctx context.Context, redisClient redis.Cmdable, serverAddress string, minutes int64) (*ServerData, []*SessionData) {

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
		core.Error("server address mismatch: got '%s', expected '%s'", serverData.ServerAddress, serverAddress)
		return nil, nil
	}

	currentTime := uint64(time.Now().Unix())

	sessionMap := make(map[uint64]bool)

	for k, v := range redis_server_sessions_a {
		session_id, _ := strconv.ParseUint(k, 16, 64)
		timestamp, _ := strconv.ParseUint(v, 10, 64)
		if currentTime-timestamp > 30 {
			continue
		}
		sessionMap[session_id] = true
	}

	for k, v := range redis_server_sessions_b {
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

	sort.Slice(serverSessionIds, func(i, j int) bool { return serverSessionIds[i] < serverSessionIds[j] })

	serverSessionData := GetSessionList(ctx, redisClient, serverSessionIds)

	return &serverData, serverSessionData
}

func GetServerList(ctx context.Context, redisClient redis.Cmdable, serverAddresses []string) []*ServerData {

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

// ------------------------------------------------------------------------------------------------------------

const TopServersVersion = uint64(0)

type TopServersWatcher struct {
	url              string
	mutex            sync.RWMutex
	totalServerCount int
	topServers       []string
}

func CreateTopServersWatcher(serverCruncherURL string) *TopServersWatcher {
	watcher := TopServersWatcher{}
	watcher.url = serverCruncherURL + "/top_servers"
	go watcher.watchTopServers()
	return &watcher
}

func (watcher *TopServersWatcher) watchTopServers() {
	ticker := time.NewTicker(time.Second)
	for {
		select {

		case <-ticker.C:

			data := getBinary(watcher.url)

			if data == nil {
				break
			}

			if len(data) < 8+4 {
				core.Error("top server response is too small")
				break
			}

			index := 0

			var version uint64
			encoding.ReadUint64(data[:], &index, &version)
			if version != TopServersVersion {
				core.Error("bad top servers version. expected %d, got %d", version, TopServersVersion)
				break
			}

			var totalServerCount uint32
			encoding.ReadUint32(data[:], &index, &totalServerCount)

			var numTopServers uint32
			encoding.ReadUint32(data[:], &index, &numTopServers)

			servers := make([]string, numTopServers)
			for i := 0; i < int(numTopServers); i++ {
				encoding.ReadString(data[:], &index, &servers[i], MaxServerAddressLength)
			}

			watcher.mutex.Lock()
			watcher.totalServerCount = int(totalServerCount)
			watcher.topServers = servers
			watcher.mutex.Unlock()
		}
	}
}

func (watcher *TopServersWatcher) GetServerCount() int {
	watcher.mutex.RLock()
	total := watcher.totalServerCount
	watcher.mutex.RUnlock()
	return total
}

func (watcher *TopServersWatcher) GetServers(begin int, end int) []string {
	if begin < 0 {
		return nil
	}
	if end <= begin {
		return nil
	}
	watcher.mutex.RLock()
	servers := watcher.topServers
	watcher.mutex.RUnlock()
	if end >= len(servers) {
		end = len(servers)
	}
	return servers[begin:end]
}

func (watcher *TopServersWatcher) GetTopServers() []string {
	watcher.mutex.RLock()
	servers := watcher.topServers
	watcher.mutex.RUnlock()
	return servers
}

// ------------------------------------------------------------------------------------------------------------

type RelayInserter struct {
	redisClient   redis.Cmdable
	lastFlushTime time.Time
	batchSize     int
	numPending    int
	pipeline      redis.Pipeliner
}

func CreateRelayInserter(redisClient redis.Cmdable, batchSize int) *RelayInserter {
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

func GetRelayCount(ctx context.Context, redisClient redis.Cmdable, minutes int64) int {

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

func GetRelayAddresses(ctx context.Context, redisClient redis.Cmdable, minutes int64, begin int, end int) []string {

	if begin < 0 {
		core.Error("invalid begin passed to get relay addresses: %d", begin)
		return nil
	}

	if end < 0 {
		core.Error("invalid end passed to get relay addresses: %d", end)
		return nil
	}

	if end <= begin {
		core.Error("end must be greater than begin")
		return nil
	}

	// get the set of relay addresses in the range [begin,end]

	pipeline := redisClient.Pipeline()

	pipeline.ZRevRangeWithScores(ctx, fmt.Sprintf("r-%d", minutes-1), int64(begin), int64(end-1))
	pipeline.ZRevRangeWithScores(ctx, fmt.Sprintf("r-%d", minutes), int64(begin), int64(end-1))

	cmds, err := pipeline.Exec(ctx)
	if err != nil {
		core.Error("failed to get relay addresses: %v", err)
		return nil
	}

	redis_relay_addresses_a, err := cmds[0].(*redis.ZSliceCmd).Result()
	if err != nil {
		core.Error("failed to get redis relay addresses a: %v", err)
		return nil
	}

	redis_relay_addresses_b, err := cmds[1].(*redis.ZSliceCmd).Result()
	if err != nil {
		core.Error("failed to get redis relay addresses b: %v", err)
		return nil
	}

	relayMap := make(map[string]int32)

	for i := range redis_relay_addresses_a {
		address := redis_relay_addresses_a[i].Member.(string)
		score := int32(redis_relay_addresses_a[i].Score)
		relayMap[address] = score
	}

	for i := range redis_relay_addresses_b {
		address := redis_relay_addresses_b[i].Member.(string)
		score := int32(redis_relay_addresses_b[i].Score)
		relayMap[address] = score
	}

	type RelayEntry struct {
		address string
		score   int32
	}

	relayEntries := make([]RelayEntry, len(relayMap))
	index := 0
	for k, v := range relayMap {
		relayEntries[index] = RelayEntry{k, v}
		index++
	}

	maxSize := end - begin
	if len(relayEntries) > maxSize {
		relayEntries = relayEntries[:maxSize]
	}

	relayAddresses := make([]string, len(relayEntries))
	for i := range relayEntries {
		relayAddresses[i] = relayEntries[i].address
	}

	return relayAddresses
}

func GetRelayData(ctx context.Context, redisClient redis.Cmdable, relayAddress string) *RelayData {

	pipeline := redisClient.Pipeline()

	pipeline.Get(ctx, fmt.Sprintf("rd-%s", relayAddress))

	cmds, err := pipeline.Exec(ctx)
	if err != nil {
		core.Error("failed to get relay data: %v", err)
		return nil
	}

	redis_relay_data := cmds[0].(*redis.StringCmd).Val()

	relayData := RelayData{}
	relayData.Parse(redis_relay_data)

	if relayData.RelayAddress != relayAddress {
		return nil
	}

	return &relayData
}

func GetRelayList(ctx context.Context, redisClient redis.Cmdable, relayAddresses []string) []*RelayData {

	pipeline := redisClient.Pipeline()

	for i := range relayAddresses {
		pipeline.Get(ctx, fmt.Sprintf("rd-%s", relayAddresses[i]))
	}

	cmds, err := pipeline.Exec(ctx)
	if err != nil {
		core.Error("failed to get relay list: %v", err)
		return nil
	}

	relayList := make([]*RelayData, 0)

	for i := range relayAddresses {

		redis_relay_data := cmds[i].(*redis.StringCmd).Val()

		relayData := RelayData{}
		relayData.Parse(redis_relay_data)

		if relayData.RelayAddress != relayAddresses[i] {
			continue
		}

		relayList = append(relayList, &relayData)
	}

	return relayList
}

// ------------------------------------------------------------------------------------------------------------
