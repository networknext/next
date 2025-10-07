package portal

import (
	"bytes"
	"context"
	"fmt"
	"io"
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
	value := fmt.Sprintf("%x|%x|%s|%d|%d|%.2f|%.2f|%d|%d|%x|%x|%s|%d|",
		data.SessionId,
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
	if len(values) < 13 {
		return
	}
	sessionId, err := strconv.ParseUint(values[0], 16, 64)
	if err != nil {
		return
	}
	startTime, err := strconv.ParseUint(values[1], 16, 64)
	if err != nil {
		return
	}
	isp := values[2]
	connectionType, err := strconv.ParseUint(values[3], 10, 32)
	if err != nil {
		return
	}
	platformType, err := strconv.ParseUint(values[4], 10, 32)
	if err != nil {
		return
	}
	latitude, err := strconv.ParseFloat(values[5], 32)
	if err != nil {
		return
	}
	longitude, err := strconv.ParseFloat(values[6], 32)
	if err != nil {
		return
	}
	directRTT, err := strconv.ParseUint(values[7], 10, 32)
	if err != nil {
		return
	}
	nextRTT, err := strconv.ParseUint(values[8], 10, 32)
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
	numRouteRelays, err := strconv.ParseUint(values[12], 10, 32)
	if err != nil {
		return
	}
	if len(values) != 13+int(numRouteRelays)+1 {
		return
	}
	routeRelays := make([]uint64, numRouteRelays)
	for i := range routeRelays {
		routeRelays[i], err = strconv.ParseUint(values[13+i], 16, 64)
		if err != nil {
			return
		}
	}
	data.SessionId = sessionId
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
	DeltaTimeMin     float32 `json:"delta_time_min"`
	DeltaTimeMax     float32 `json:"delta_time_max"`
	DeltaTimeAvg     float32 `json:"delta_time_avg"`
	InternalEvents   uint64  `json:"internal_events,string"`
	SessionEvents    uint64  `json:"session_events,string"`
	DirectKbpsUp     uint32  `json:"direct_kbps_up"`
	DirectKbpsDown   uint32  `json:"direct_kbps_down"`
	NextKbpsUp       uint32  `json:"next_kbps_up"`
	NextKbpsDown     uint32  `json:"next_kbps_down"`
	Next             bool    `json:"next"`
}

func (data *SliceData) Value() string {
	return fmt.Sprintf("%x|%d|%d|%d|%d|%d|%d|%d|%.2f|%.2f|%.2f|%.2f|%x|%x|%d|%d|%d|%d|%v|%.3f|%.3f|%.3f",
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
		data.DeltaTimeMin,
		data.DeltaTimeMax,
		data.DeltaTimeAvg,
	)
}

func (data *SliceData) Parse(value string) {
	values := strings.Split(value, "|")
	if len(values) != 22 {
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
	deltaTimeMin, err := strconv.ParseFloat(values[19], 32)
	if err != nil {
		return
	}
	deltaTimeMax, err := strconv.ParseFloat(values[20], 32)
	if err != nil {
		return
	}
	deltaTimeAvg, err := strconv.ParseFloat(values[21], 32)
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
	data.Next = next
	data.DeltaTimeMin = float32(deltaTimeMin)
	data.DeltaTimeMax = float32(deltaTimeMax)
	data.DeltaTimeAvg = float32(deltaTimeAvg)
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
	data.DeltaTimeMin = 1.0
	data.DeltaTimeMax = 2.0
	data.DeltaTimeAvg = 4.0
	return &data
}

// --------------------------------------------------------------------------------------------------

type ClientRelayData struct {
	Timestamp             uint64                             `json:"timestamp,string"`
	NumClientRelays       uint32                             `json:"num_client_relays"`
	ClientRelayId         [constants.MaxClientRelays]uint64  `json:"client_relay_id"`
	ClientRelayRTT        [constants.MaxClientRelays]uint8   `json:"client_relay_rtt"`
	ClientRelayJitter     [constants.MaxClientRelays]uint8   `json:"client_relay_jitter"`
	ClientRelayPacketLoss [constants.MaxClientRelays]float32 `json:"client_relay_packet_loss"`
}

func (data *ClientRelayData) Value() string {
	output := fmt.Sprintf("%x|%d", data.Timestamp, data.NumClientRelays)
	for i := 0; i < int(data.NumClientRelays); i++ {
		output += fmt.Sprintf("|%x|%d|%d|%.2f", data.ClientRelayId[i], data.ClientRelayRTT[i], data.ClientRelayJitter[i], data.ClientRelayPacketLoss[i])
	}
	return output
}

func (data *ClientRelayData) Parse(value string) {
	values := strings.Split(value, "|")
	if len(values) < 2 {
		return
	}
	timestamp, err := strconv.ParseUint(values[0], 16, 64)
	if err != nil {
		return
	}
	numClientRelays, err := strconv.ParseInt(values[1], 10, 8)
	if err != nil || numClientRelays < 0 || numClientRelays > constants.MaxClientRelays {
		return
	}
	if len(values) != 2+int(numClientRelays)*4 {
		return
	}
	clientRelayId := make([]uint64, numClientRelays)
	clientRelayRTT := make([]uint64, numClientRelays)
	clientRelayJitter := make([]uint64, numClientRelays)
	clientRelayPacketLoss := make([]float64, numClientRelays)
	for i := 0; i < int(numClientRelays); i++ {
		clientRelayId[i], err = strconv.ParseUint(values[2+i*4], 16, 64)
		if err != nil {
			return
		}
		clientRelayRTT[i], err = strconv.ParseUint(values[2+i*4+1], 10, 8)
		if err != nil {
			return
		}
		clientRelayJitter[i], err = strconv.ParseUint(values[2+i*4+2], 10, 8)
		if err != nil {
			return
		}
		clientRelayPacketLoss[i], err = strconv.ParseFloat(values[2+i*4+3], 32)
		if err != nil {
			return
		}
	}
	data.Timestamp = timestamp
	data.NumClientRelays = uint32(numClientRelays)
	for i := 0; i < int(numClientRelays); i++ {
		data.ClientRelayId[i] = clientRelayId[i]
		data.ClientRelayRTT[i] = uint8(clientRelayRTT[i])
		data.ClientRelayJitter[i] = uint8(clientRelayJitter[i])
		data.ClientRelayPacketLoss[i] = float32(clientRelayPacketLoss[i])
	}
	return
}

func GenerateRandomClientRelayData() *ClientRelayData {
	data := ClientRelayData{}
	data.Timestamp = uint64(time.Now().Unix())
	data.NumClientRelays = constants.MaxClientRelays
	for i := 0; i < int(data.NumClientRelays); i++ {
		data.ClientRelayId[i] = rand.Uint64()
		data.ClientRelayRTT[i] = uint8(common.RandomInt(5, 20))
		data.ClientRelayJitter[i] = uint8(common.RandomInt(5, 20))
		data.ClientRelayPacketLoss[i] = float32(common.RandomInt(0, 10000)) / 100.0
	}
	return &data
}

// --------------------------------------------------------------------------------------------------

type ServerRelayData struct {
	Timestamp             uint64                             `json:"timestamp,string"`
	NumServerRelays       uint32                             `json:"num_client_relays"`
	ServerRelayId         [constants.MaxServerRelays]uint64  `json:"server_relay_id"`
	ServerRelayRTT        [constants.MaxServerRelays]uint8   `json:"server_relay_rtt"`
	ServerRelayJitter     [constants.MaxServerRelays]uint8   `json:"server_relay_jitter"`
	ServerRelayPacketLoss [constants.MaxServerRelays]float32 `json:"server_relay_packet_loss"`
}

func (data *ServerRelayData) Value() string {
	output := fmt.Sprintf("%x|%d", data.Timestamp, data.NumServerRelays)
	for i := 0; i < int(data.NumServerRelays); i++ {
		output += fmt.Sprintf("|%x|%d|%d|%.2f", data.ServerRelayId[i], data.ServerRelayRTT[i], data.ServerRelayJitter[i], data.ServerRelayPacketLoss[i])
	}
	return output
}

func (data *ServerRelayData) Parse(value string) {
	values := strings.Split(value, "|")
	if len(values) < 2 {
		return
	}
	timestamp, err := strconv.ParseUint(values[0], 16, 64)
	if err != nil {
		return
	}
	numServerRelays, err := strconv.ParseInt(values[1], 10, 8)
	if err != nil || numServerRelays < 0 || numServerRelays > constants.MaxServerRelays {
		return
	}
	if len(values) != 2+int(numServerRelays)*4 {
		return
	}
	serverRelayId := make([]uint64, numServerRelays)
	serverRelayRTT := make([]uint64, numServerRelays)
	serverRelayJitter := make([]uint64, numServerRelays)
	serverRelayPacketLoss := make([]float64, numServerRelays)
	for i := 0; i < int(numServerRelays); i++ {
		serverRelayId[i], err = strconv.ParseUint(values[2+i*4], 16, 64)
		if err != nil {
			return
		}
		serverRelayRTT[i], err = strconv.ParseUint(values[2+i*4+1], 10, 8)
		if err != nil {
			return
		}
		serverRelayJitter[i], err = strconv.ParseUint(values[2+i*4+2], 10, 8)
		if err != nil {
			return
		}
		serverRelayPacketLoss[i], err = strconv.ParseFloat(values[2+i*4+3], 32)
		if err != nil {
			return
		}
	}
	data.Timestamp = timestamp
	data.NumServerRelays = uint32(numServerRelays)
	for i := 0; i < int(numServerRelays); i++ {
		data.ServerRelayId[i] = serverRelayId[i]
		data.ServerRelayRTT[i] = uint8(serverRelayRTT[i])
		data.ServerRelayJitter[i] = uint8(serverRelayJitter[i])
		data.ServerRelayPacketLoss[i] = float32(serverRelayPacketLoss[i])
	}
	return
}

func GenerateRandomServerRelayData() *ServerRelayData {
	data := ServerRelayData{}
	data.Timestamp = uint64(time.Now().Unix())
	data.NumServerRelays = constants.MaxServerRelays
	for i := 0; i < int(data.NumServerRelays); i++ {
		data.ServerRelayId[i] = rand.Uint64()
		data.ServerRelayRTT[i] = uint8(common.RandomInt(5, 20))
		data.ServerRelayJitter[i] = uint8(common.RandomInt(5, 20))
		data.ServerRelayPacketLoss[i] = float32(common.RandomInt(0, 10000)) / 100.0
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
	ServerId         uint64 `json:"server_id,string"`
	DatacenterId     uint64 `json:"datacenter_id,string"`
	NumSessions      uint32 `json:"num_sessions"`
	Uptime           uint64 `json:"uptime,string"`
}

func (data *ServerData) Value() string {
	return fmt.Sprintf("%s|%d|%d|%d|%x|%x|%x|%d|%x",
		data.ServerAddress,
		data.SDKVersion_Major,
		data.SDKVersion_Minor,
		data.SDKVersion_Patch,
		data.BuyerId,
		data.ServerId,
		data.DatacenterId,
		data.NumSessions,
		data.Uptime,
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
	buyerId, err := strconv.ParseUint(values[4], 16, 64)
	if err != nil {
		return
	}
	serverId, err := strconv.ParseUint(values[5], 16, 64)
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
	uptime, err := strconv.ParseUint(values[8], 16, 64)
	if err != nil {
		return
	}
	data.ServerAddress = serverAddress
	data.SDKVersion_Major = uint8(sdkVersionMajor)
	data.SDKVersion_Minor = uint8(sdkVersionMinor)
	data.SDKVersion_Patch = uint8(sdkVersionPatch)
	data.BuyerId = buyerId
	data.ServerId = serverId
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
	data.ServerId = rand.Uint64()
	data.DatacenterId = rand.Uint64()
	data.NumSessions = rand.Uint32()
	data.Uptime = rand.Uint64()
	return &data
}

// --------------------------------------------------------------------------------------------------

type RelayData struct {
	RelayName    string `json:"relay_name"`
	RelayId      uint64 `json:"relay_id,string"`
	NumSessions  uint32 `json:"num_sessions"`
	MaxSessions  uint32 `json:"max_sessions"`
	StartTime    uint64 `json:"start_time,string"`
	RelayFlags   uint64 `json:"relay_flags,string"`
	RelayVersion string `json:"relay_version"`
}

func (data *RelayData) Value() string {
	return fmt.Sprintf("%s|%x|%d|%d|%x|%x|%s",
		data.RelayName,
		data.RelayId,
		data.NumSessions,
		data.MaxSessions,
		data.StartTime,
		data.RelayFlags,
		data.RelayVersion,
	)
}

func (data *RelayData) Parse(value string) {

	values := strings.Split(value, "|")
	if len(values) != 7 {
		return
	}
	relayName := values[0]
	relayId, err := strconv.ParseUint(values[1], 16, 64)
	if err != nil {
		return
	}
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
	relayVersion := values[6]

	data.RelayName = relayName
	data.RelayId = relayId
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
	ClientPingsPerSecond      float32 `json:"client_pings_per_second"`
	ServerPingsPerSecond      float32 `json:"server_pings_per_second"`
	RelayPingsPerSecond       float32 `json:"relay_pings_per_second"`
	RelayFlags                uint64  `json:"relay_flags,string"`
	NumRoutable               uint32  `json:"num_routable"`
	NumUnroutable             uint32  `json:"num_unroutable"`
	CurrentTime               uint64  `json:"current_time,string"`
}

func (data *RelaySample) Value() string {
	return fmt.Sprintf("%x|%d|%d|%d|%.2f|%.2f|%.2f|%.2f|%.2f|%.2f|%.2f|%x|%d|%d|%x",
		data.Timestamp,
		data.NumSessions,
		data.EnvelopeBandwidthUpKbps,
		data.EnvelopeBandwidthDownKbps,
		data.PacketsSentPerSecond,
		data.PacketsReceivedPerSecond,
		data.BandwidthSentKbps,
		data.BandwidthReceivedKbps,
		data.ClientPingsPerSecond,
		data.ServerPingsPerSecond,
		data.RelayPingsPerSecond,
		data.RelayFlags,
		data.NumRoutable,
		data.NumUnroutable,
		data.CurrentTime,
	)
}

func (data *RelaySample) Parse(value string) {
	values := strings.Split(value, "|")
	if len(values) != 15 {
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
	clientPingsPerSecond, err := strconv.ParseFloat(values[8], 32)
	if err != nil {
		return
	}
	serverPingsPerSecond, err := strconv.ParseFloat(values[9], 32)
	if err != nil {
		return
	}
	relayPingsPerSecond, err := strconv.ParseFloat(values[10], 32)
	if err != nil {
		return
	}
	relayFlags, err := strconv.ParseUint(values[11], 16, 64)
	if err != nil {
		return
	}
	numRoutable, err := strconv.ParseUint(values[12], 10, 32)
	if err != nil {
		return
	}
	numUnroutable, err := strconv.ParseUint(values[13], 10, 32)
	if err != nil {
		return
	}
	currentTime, err := strconv.ParseUint(values[14], 16, 64)
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
	data.ClientPingsPerSecond = float32(clientPingsPerSecond)
	data.ServerPingsPerSecond = float32(serverPingsPerSecond)
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
	data.ClientPingsPerSecond = float32(common.RandomInt(0, 1000))
	data.ServerPingsPerSecond = float32(common.RandomInt(0, 1000))
	data.RelayPingsPerSecond = float32(common.RandomInt(0, 1000))
	data.RelayFlags = rand.Uint64()
	data.NumRoutable = rand.Uint32()
	data.NumUnroutable = rand.Uint32()
	data.CurrentTime = rand.Uint64()
	return &data
}

// ------------------------------------------------------------------------------------------------------------

type SessionCruncherEntry struct {
	SessionId uint64
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

	batchSize := make([]uint32, constants.NumBuckets)

	for i := range publisher.batchMessages {
		batchIndex := int(publisher.batchMessages[i].Score)
		if batchIndex > constants.NumBuckets-1 {
			batchIndex = constants.NumBuckets - 1
		}
		batchSize[batchIndex]++
	}

	batch := make([][]SessionCruncherEntry, constants.NumBuckets)

	for i := range batchSize {
		batch[i] = make([]SessionCruncherEntry, 0, batchSize[i])
	}

	for i := range publisher.batchMessages {
		batchIndex := int(publisher.batchMessages[i].Score)
		if batchIndex > constants.NumBuckets-1 {
			batchIndex = constants.NumBuckets - 1
		}
		batch[batchIndex] = append(batch[batchIndex], publisher.batchMessages[i])
	}

	size := 8 + 4*constants.NumBuckets
	for i := range batchSize {
		size += int(batchSize[i]) * (8 + 1 + 4 + 4)
	}

	data := make([]byte, size)

	index := 0

	encoding.WriteUint64(data[:], &index, SessionBatchVersion_Write)

	for i := 0; i < constants.NumBuckets; i++ {
		encoding.WriteUint32(data[:], &index, uint32(batchSize[i]))
		for j := range batch[i] {
			encoding.WriteUint64(data[:], &index, batch[i][j].SessionId)
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

	body, error := io.ReadAll(response.Body)
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

func (inserter *SessionInserter) Insert(ctx context.Context, sessionId uint64, next bool, score uint32, sessionData *SessionData, sliceData *SliceData) {

	currentTime := time.Now()

	minutes := currentTime.Unix() / 60

	entry := SessionCruncherEntry{
		SessionId: sessionId,
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
	url         string
	mutex       sync.RWMutex
	topSessions []uint64
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

			if len(data) < 8 {
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

			numSessions := (len(data) - 8) / 8
			sessions := make([]uint64, numSessions)
			for i := 0; i < numSessions; i++ {
				encoding.ReadUint64(data[:], &index, &sessions[i])
			}

			watcher.mutex.Lock()
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

	body, error := io.ReadAll(response.Body)
	if error != nil {
		core.Error("could not read response body for %s: %v", url, err)
		return nil
	}

	response.Body.Close()

	return body
}

func (watcher *TopSessionsWatcher) GetSessions(begin int, end int) []uint64 {
	if begin < 0 {
		return nil
	}
	if end <= begin {
		return nil
	}
	watcher.mutex.RLock()
	if end > len(watcher.topSessions) {
		end = len(watcher.topSessions)
	}
	sessions := make([]uint64, end-begin)
	copy(sessions, watcher.topSessions[begin:end])
	watcher.mutex.RUnlock()
	return sessions
}

func (watcher *TopSessionsWatcher) GetTopSessions() []uint64 {
	watcher.mutex.RLock()
	sessions := watcher.topSessions
	watcher.mutex.RUnlock()
	return sessions
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

func GetSessionData(ctx context.Context, redisClient redis.Cmdable, sessionId uint64) (*SessionData, []SliceData, []ClientRelayData, []ServerRelayData) {

	pipeline := redisClient.Pipeline()

	pipeline.Get(ctx, fmt.Sprintf("sd-%016x", sessionId))
	pipeline.LRange(ctx, fmt.Sprintf("sl-%016x", sessionId), 0, -1)
	pipeline.LRange(ctx, fmt.Sprintf("crd-%016x", sessionId), 0, -1)
	pipeline.LRange(ctx, fmt.Sprintf("srd-%016x", sessionId), 0, -1)

	cmds, err := pipeline.Exec(ctx)
	if err != nil {
		core.Error("failed to get session data: %v", err)
		return nil, nil, nil, nil
	}

	redis_session_data := cmds[0].(*redis.StringCmd).Val()
	redis_slice_data := cmds[1].(*redis.StringSliceCmd).Val()
	redis_client_relay_data := cmds[2].(*redis.StringSliceCmd).Val()
	redis_server_relay_data := cmds[3].(*redis.StringSliceCmd).Val()

	sessionData := SessionData{}
	sessionData.Parse(redis_session_data)

	sliceData := make([]SliceData, len(redis_slice_data))
	for i := 0; i < len(redis_slice_data); i++ {
		sliceData[i].Parse(redis_slice_data[i])
	}

	clientRelayData := make([]ClientRelayData, len(redis_client_relay_data))
	for i := 0; i < len(redis_client_relay_data); i++ {
		clientRelayData[i].Parse(redis_client_relay_data[i])
	}

	serverRelayData := make([]ServerRelayData, len(redis_server_relay_data))
	for i := 0; i < len(redis_server_relay_data); i++ {
		serverRelayData[i].Parse(redis_server_relay_data[i])
	}

	return &sessionData, sliceData, clientRelayData, serverRelayData
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

// ------------------------------------------------------------------------------------------------------------

type ClientRelayInserter struct {
	redisClient   redis.Cmdable
	lastFlushTime time.Time
	batchSize     int
	numPending    int
	pipeline      redis.Pipeliner
}

func CreateClientRelayInserter(redisClient redis.Cmdable, batchSize int) *ClientRelayInserter {
	inserter := ClientRelayInserter{}
	inserter.redisClient = redisClient
	inserter.lastFlushTime = time.Now()
	inserter.batchSize = batchSize
	inserter.pipeline = redisClient.Pipeline()
	return &inserter
}

func (inserter *ClientRelayInserter) Insert(ctx context.Context, sessionId uint64, clientRelayData *ClientRelayData) {

	currentTime := time.Now()

	key := fmt.Sprintf("crd-%016x", sessionId)
	inserter.pipeline.RPush(ctx, key, clientRelayData.Value())

	inserter.numPending++

	inserter.CheckForFlush(ctx, currentTime)
}

func (inserter *ClientRelayInserter) CheckForFlush(ctx context.Context, currentTime time.Time) {
	if inserter.numPending > inserter.batchSize || currentTime.Sub(inserter.lastFlushTime) >= time.Second {
		inserter.Flush(ctx)
	}
}

func (inserter *ClientRelayInserter) Flush(ctx context.Context) {
	_, err := inserter.pipeline.Exec(ctx)
	if err != nil {
		core.Error("client relay insert error: %v", err)
	}
	inserter.numPending = 0
	inserter.lastFlushTime = time.Now()
	inserter.pipeline = inserter.redisClient.Pipeline()
}

// ------------------------------------------------------------------------------------------------------------

type ServerRelayInserter struct {
	redisClient   redis.Cmdable
	lastFlushTime time.Time
	batchSize     int
	numPending    int
	pipeline      redis.Pipeliner
}

func CreateServerRelayInserter(redisClient redis.Cmdable, batchSize int) *ServerRelayInserter {
	inserter := ServerRelayInserter{}
	inserter.redisClient = redisClient
	inserter.lastFlushTime = time.Now()
	inserter.batchSize = batchSize
	inserter.pipeline = redisClient.Pipeline()
	return &inserter
}

func (inserter *ServerRelayInserter) Insert(ctx context.Context, sessionId uint64, serverRelayData *ServerRelayData) {

	currentTime := time.Now()

	key := fmt.Sprintf("srd-%016x", sessionId)
	inserter.pipeline.RPush(ctx, key, serverRelayData.Value())

	inserter.numPending++

	inserter.CheckForFlush(ctx, currentTime)
}

func (inserter *ServerRelayInserter) CheckForFlush(ctx context.Context, currentTime time.Time) {
	if inserter.numPending > inserter.batchSize || currentTime.Sub(inserter.lastFlushTime) >= time.Second {
		inserter.Flush(ctx)
	}
}

func (inserter *ServerRelayInserter) Flush(ctx context.Context) {
	_, err := inserter.pipeline.Exec(ctx)
	if err != nil {
		core.Error("server relay insert error: %v", err)
	}
	inserter.numPending = 0
	inserter.lastFlushTime = time.Now()
	inserter.pipeline = inserter.redisClient.Pipeline()
}

// ------------------------------------------------------------------------------------------------------------

const MaxServerAddressLength = 64

type ServerCruncherEntry struct {
	ServerAddress string
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

	batchSize := make([]uint32, constants.NumBuckets)

	for i := range publisher.batchMessages {
		batchIndex := int(publisher.batchMessages[i].Score)
		if batchIndex > constants.NumBuckets-1 {
			batchIndex = constants.NumBuckets - 1
		}
		batchSize[batchIndex]++
	}

	batch := make([][]ServerCruncherEntry, constants.NumBuckets)

	for i := range batchSize {
		batch[i] = make([]ServerCruncherEntry, 0, batchSize[i])
	}

	for i := range publisher.batchMessages {
		batchIndex := int(publisher.batchMessages[i].Score)
		if batchIndex > constants.NumBuckets-1 {
			batchIndex = constants.NumBuckets - 1
		}
		batch[batchIndex] = append(batch[batchIndex], publisher.batchMessages[i])
	}

	size := 8 + 4*constants.NumBuckets
	for i := range batchSize {
		size += 4 + int(batchSize[i])*(MaxServerAddressLength+8)
	}

	data := make([]byte, size)

	index := 0

	encoding.WriteUint64(data[:], &index, ServerBatchVersion_Write)

	for i := 0; i < constants.NumBuckets; i++ {
		encoding.WriteUint32(data[:], &index, uint32(batchSize[i]))
		for j := range batch[i] {
			encoding.WriteString(data, &index, batch[i][j].ServerAddress, MaxServerAddressLength)
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

	score := (uint32(serverId) ^ uint32(serverId>>32)) % uint32(constants.MaxScore+1)

	entry := ServerCruncherEntry{
		ServerAddress: serverData.ServerAddress,
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
	url        string
	mutex      sync.RWMutex
	topServers []string
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

			var numTopServers uint32
			encoding.ReadUint32(data[:], &index, &numTopServers)

			servers := make([]string, numTopServers)
			for i := 0; i < int(numTopServers); i++ {
				encoding.ReadString(data[:], &index, &servers[i], MaxServerAddressLength)
			}

			watcher.mutex.Lock()
			watcher.topServers = servers
			watcher.mutex.Unlock()
		}
	}
}

func (watcher *TopServersWatcher) GetServers(begin int, end int) []string {
	if begin < 0 {
		return nil
	}
	if end <= begin {
		return nil
	}
	watcher.mutex.RLock()
	if end > len(watcher.topServers) {
		end = len(watcher.topServers)
	}
	servers := make([]string, end-begin)
	copy(servers, watcher.topServers[begin:end])
	watcher.mutex.RUnlock()
	return servers
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

	key := fmt.Sprintf("r-%x", minutes)
	inserter.pipeline.ZAdd(ctx, key, redis.Z{Score: float64(score), Member: relayData.RelayName})

	inserter.pipeline.Set(ctx, fmt.Sprintf("rd-%s", relayData.RelayName), relayData.Value(), 0)

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

func GetRelayNames(ctx context.Context, redisClient redis.Cmdable, minutes int64, begin int, end int) []string {

	if begin < 0 {
		core.Error("invalid begin passed to get relay names: %d", begin)
		return nil
	}

	if end < 0 {
		core.Error("invalid end passed to get relay names: %d", end)
		return nil
	}

	if end <= begin {
		core.Error("end must be greater than begin")
		return nil
	}

	// get the set of relay names in the range [begin,end]

	pipeline := redisClient.Pipeline()

	pipeline.ZRevRangeWithScores(ctx, fmt.Sprintf("r-%d", minutes-1), int64(begin), int64(end-1))
	pipeline.ZRevRangeWithScores(ctx, fmt.Sprintf("r-%d", minutes), int64(begin), int64(end-1))

	cmds, err := pipeline.Exec(ctx)
	if err != nil {
		core.Error("failed to get relay names: %v", err)
		return nil
	}

	redis_relay_names_a, err := cmds[0].(*redis.ZSliceCmd).Result()
	if err != nil {
		core.Error("failed to get redis relay names a: %v", err)
		return nil
	}

	redis_relay_names_b, err := cmds[1].(*redis.ZSliceCmd).Result()
	if err != nil {
		core.Error("failed to get redis relay names b: %v", err)
		return nil
	}

	relayMap := make(map[string]int32)

	for i := range redis_relay_names_a {
		name := redis_relay_names_a[i].Member.(string)
		score := int32(redis_relay_names_a[i].Score)
		relayMap[name] = score
	}

	for i := range redis_relay_names_b {
		name := redis_relay_names_b[i].Member.(string)
		score := int32(redis_relay_names_b[i].Score)
		relayMap[name] = score
	}

	type RelayEntry struct {
		name  string
		score int32
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

	relayNames := make([]string, len(relayEntries))
	for i := range relayNames {
		relayNames[i] = relayEntries[i].name
	}

	return relayNames
}

func GetRelayData(ctx context.Context, redisClient redis.Cmdable, relayName string) *RelayData {

	pipeline := redisClient.Pipeline()

	pipeline.Get(ctx, fmt.Sprintf("rd-%s", relayName))

	cmds, err := pipeline.Exec(ctx)
	if err != nil {
		core.Error("failed to get relay data: %v", err)
		return nil
	}

	redis_relay_data := cmds[0].(*redis.StringCmd).Val()

	relayData := RelayData{}
	relayData.Parse(redis_relay_data)

	if relayData.RelayName != relayName {
		return nil
	}

	return &relayData
}

func GetRelayList(ctx context.Context, redisClient redis.Cmdable, relayNames []string) []*RelayData {

	pipeline := redisClient.Pipeline()

	for i := range relayNames {
		pipeline.Get(ctx, fmt.Sprintf("rd-%s", relayNames[i]))
	}

	cmds, err := pipeline.Exec(ctx)
	if err != nil {
		core.Error("failed to get relay list: %v", err)
		return nil
	}

	relayList := make([]*RelayData, 0)

	for i := range relayNames {

		redis_relay_data := cmds[i].(*redis.StringCmd).Val()

		relayData := RelayData{}
		relayData.Parse(redis_relay_data)

		if relayData.RelayName != relayNames[i] {
			continue
		}

		relayList = append(relayList, &relayData)
	}

	return relayList
}

// ------------------------------------------------------------------------------------------------------------
