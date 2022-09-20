package packets

import (
	"errors"
	"net"
	"fmt"

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/crypto"
)

// ------------------------------------------------------------

type SDK4_ServerInitRequestPacket struct {
	Version        SDKVersion
	BuyerId        uint64
	DatacenterId   uint64
	RequestId      uint64
	DatacenterName string
}

func (packet *SDK4_ServerInitRequestPacket) Serialize(stream common.Stream) error {
	packet.Version.Serialize(stream)
	stream.SerializeUint64(&packet.BuyerId)
	stream.SerializeUint64(&packet.DatacenterId)
	stream.SerializeUint64(&packet.RequestId)
	stream.SerializeString(&packet.DatacenterName, SDK4_MaxDatacenterNameLength)
	return stream.Error()
}

// ------------------------------------------------------------

type SDK4_ServerInitResponsePacket struct {
	RequestId uint64
	Response  uint32
}

func (packet *SDK4_ServerInitResponsePacket) Serialize(stream common.Stream) error {
	stream.SerializeUint64(&packet.RequestId)
	stream.SerializeBits(&packet.Response, 8)
	return stream.Error()
}

// ------------------------------------------------------------

type SDK4_ServerUpdatePacket struct {
	Version       SDKVersion
	BuyerId       uint64
	DatacenterId  uint64
	NumSessions   uint32
	ServerAddress net.UDPAddr
}

func (packet *SDK4_ServerUpdatePacket) Serialize(stream common.Stream) error {
	packet.Version.Serialize(stream)
	stream.SerializeUint64(&packet.BuyerId)
	stream.SerializeUint64(&packet.DatacenterId)
	stream.SerializeUint32(&packet.NumSessions)
	stream.SerializeAddress(&packet.ServerAddress)
	return stream.Error()
}

// ------------------------------------------------------------

type SDK4_SessionUpdatePacket struct {
	Version                         SDKVersion
	BuyerId                         uint64
	DatacenterId                    uint64
	SessionId                       uint64
	SliceNumber                     uint32
	RetryNumber                     int32
	SessionDataBytes                int32
	SessionData                     [SDK4_MaxSessionDataSize]byte
	ClientAddress                   net.UDPAddr
	ServerAddress                   net.UDPAddr
	ClientRoutePublicKey            [crypto.KeySize]byte
	ServerRoutePublicKey            [crypto.KeySize]byte
	UserHash                        uint64
	PlatformType                    int32
	ConnectionType                  int32
	Next                            bool
	Committed                       bool
	Reported                        bool
	FallbackToDirect                bool
	ClientBandwidthOverLimit        bool
	ServerBandwidthOverLimit        bool
	ClientPingTimedOut              bool
	NumTags                         int32
	Tags                            [SDK4_MaxTags]uint64
	Flags                           uint32
	UserFlags                       uint64
	DirectMinRTT                    float32
	DirectMaxRTT                    float32
	DirectPrimeRTT                  float32
	DirectJitter                    float32
	DirectPacketLoss                float32
	NextRTT                         float32
	NextJitter                      float32
	NextPacketLoss                  float32
	NumNearRelays                   int32
	NearRelayIds                    [SDK4_MaxNearRelays]uint64
	NearRelayRTT                    [SDK4_MaxNearRelays]int32
	NearRelayJitter                 [SDK4_MaxNearRelays]int32
	NearRelayPacketLoss             [SDK4_MaxNearRelays]int32
	NextKbpsUp                      uint32
	NextKbpsDown                    uint32
	PacketsSentClientToServer       uint64
	PacketsSentServerToClient       uint64
	PacketsLostClientToServer       uint64
	PacketsLostServerToClient       uint64
	PacketsOutOfOrderClientToServer uint64
	PacketsOutOfOrderServerToClient uint64
	JitterClientToServer            float32
	JitterServerToClient            float32
}

func (packet *SDK4_SessionUpdatePacket) Serialize(stream common.Stream) error {

	packet.Version.Serialize(stream)

	stream.SerializeUint64(&packet.BuyerId)
	stream.SerializeUint64(&packet.DatacenterId)
	stream.SerializeUint64(&packet.SessionId)
	stream.SerializeUint32(&packet.SliceNumber)
	stream.SerializeInteger(&packet.RetryNumber, 0, SDK4_MaxSessionUpdateRetries)

	stream.SerializeInteger(&packet.SessionDataBytes, 0, SDK4_MaxSessionDataSize)
	if packet.SessionDataBytes > 0 {
		sessionData := packet.SessionData[:packet.SessionDataBytes]
		stream.SerializeBytes(sessionData)
	}

	stream.SerializeAddress(&packet.ClientAddress)

	stream.SerializeAddress(&packet.ServerAddress)

	stream.SerializeBytes(packet.ClientRoutePublicKey[:])

	stream.SerializeBytes(packet.ServerRoutePublicKey[:])

	stream.SerializeUint64(&packet.UserHash)

	stream.SerializeInteger(&packet.PlatformType, SDK4_PlatformTypeUnknown, SDK4_PlatformTypeMax)

	stream.SerializeInteger(&packet.ConnectionType, SDK4_ConnectionTypeUnknown, SDK4_ConnectionTypeMax)

	stream.SerializeBool(&packet.Next)
	stream.SerializeBool(&packet.Committed)
	stream.SerializeBool(&packet.Reported)
	stream.SerializeBool(&packet.FallbackToDirect)
	stream.SerializeBool(&packet.ClientBandwidthOverLimit)
	stream.SerializeBool(&packet.ServerBandwidthOverLimit)
	stream.SerializeBool(&packet.ClientPingTimedOut)

	hasTags := stream.IsWriting() && packet.NumTags > 0
	hasFlags := stream.IsWriting() && packet.Flags != 0
	hasUserFlags := stream.IsWriting() && packet.UserFlags != 0
	hasLostPackets := stream.IsWriting() && (packet.PacketsLostClientToServer+packet.PacketsLostServerToClient) > 0
	hasOutOfOrderPackets := stream.IsWriting() && (packet.PacketsOutOfOrderClientToServer+packet.PacketsOutOfOrderServerToClient) > 0

	stream.SerializeBool(&hasTags)
	stream.SerializeBool(&hasFlags)
	stream.SerializeBool(&hasUserFlags)
	stream.SerializeBool(&hasLostPackets)
	stream.SerializeBool(&hasOutOfOrderPackets)

	if hasTags {
		stream.SerializeInteger(&packet.NumTags, 0, SDK4_MaxTags)
		for i := 0; i < int(packet.NumTags); i++ {
			stream.SerializeUint64(&packet.Tags[i])
		}
	}

	if hasFlags {
		stream.SerializeBits(&packet.Flags, SDK4_FallbackFlagsCount)
	}

	if hasUserFlags {
		stream.SerializeUint64(&packet.UserFlags)
	}

	stream.SerializeFloat32(&packet.DirectMinRTT)
	stream.SerializeFloat32(&packet.DirectMaxRTT)
	stream.SerializeFloat32(&packet.DirectPrimeRTT)
	stream.SerializeFloat32(&packet.DirectJitter)
	stream.SerializeFloat32(&packet.DirectPacketLoss)

	if packet.Next {
		stream.SerializeFloat32(&packet.NextRTT)
		stream.SerializeFloat32(&packet.NextJitter)
		stream.SerializeFloat32(&packet.NextPacketLoss)
	}

	stream.SerializeInteger(&packet.NumNearRelays, 0, SDK4_MaxNearRelays)

	for i := int32(0); i < packet.NumNearRelays; i++ {
		stream.SerializeUint64(&packet.NearRelayIds[i])
		stream.SerializeInteger(&packet.NearRelayRTT[i], 0, 255)
		stream.SerializeInteger(&packet.NearRelayJitter[i], 0, 255)
		stream.SerializeInteger(&packet.NearRelayPacketLoss[i], 0, 100)
	}

	if packet.Next {
		stream.SerializeUint32(&packet.NextKbpsUp)
		stream.SerializeUint32(&packet.NextKbpsDown)
	}

	stream.SerializeUint64(&packet.PacketsSentClientToServer)
	stream.SerializeUint64(&packet.PacketsSentServerToClient)

	if hasLostPackets {
		stream.SerializeUint64(&packet.PacketsLostClientToServer)
		stream.SerializeUint64(&packet.PacketsLostServerToClient)
	}

	if hasOutOfOrderPackets {
		stream.SerializeUint64(&packet.PacketsOutOfOrderClientToServer)
		stream.SerializeUint64(&packet.PacketsOutOfOrderServerToClient)
	}

	stream.SerializeFloat32(&packet.JitterClientToServer)
	stream.SerializeFloat32(&packet.JitterServerToClient)

	return stream.Error()
}

// ------------------------------------------------------------

type SDK4_SessionResponsePacket struct {
	Version            SDKVersion
	SessionId          uint64
	SliceNumber        uint32
	SessionDataBytes   int32
	SessionData        [SDK4_MaxSessionDataSize]byte
	RouteType          int32
	NearRelaysChanged  bool
	NumNearRelays      int32
	NearRelayIds       [SDK4_MaxNearRelays]uint64
	NearRelayAddresses [SDK4_MaxNearRelays]net.UDPAddr
	NumTokens          int32
	Tokens             []byte
	Multipath          bool
	Committed          bool
	HasDebug           bool
	Debug              string
	ExcludeNearRelays  bool
	NearRelayExcluded  [SDK4_MaxNearRelays]bool
	HighFrequencyPings bool
}

func (packet *SDK4_SessionResponsePacket) Serialize(stream common.Stream) error {

	packet.Version.Serialize(stream)

	stream.SerializeUint64(&packet.SessionId)

	stream.SerializeUint32(&packet.SliceNumber)

	stream.SerializeInteger(&packet.SessionDataBytes, 0, SDK4_MaxSessionDataSize)
	if packet.SessionDataBytes > 0 {
		sessionData := packet.SessionData[:packet.SessionDataBytes]
		stream.SerializeBytes(sessionData)
	}

	stream.SerializeInteger(&packet.RouteType, 0, SDK4_RouteTypeContinue)

	stream.SerializeBool(&packet.NearRelaysChanged)

	if packet.NearRelaysChanged {
		stream.SerializeInteger(&packet.NumNearRelays, 0, SDK4_MaxNearRelays)
		for i := int32(0); i < packet.NumNearRelays; i++ {
			stream.SerializeUint64(&packet.NearRelayIds[i])
			stream.SerializeAddress(&packet.NearRelayAddresses[i])
		}
	}

	if packet.RouteType != SDK4_RouteTypeDirect {
		stream.SerializeBool(&packet.Multipath)
		stream.SerializeBool(&packet.Committed)
		stream.SerializeInteger(&packet.NumTokens, 0, SDK4_MaxTokens)
	}

	if packet.RouteType == SDK4_RouteTypeNew {
		if stream.IsReading() {
			packet.Tokens = make([]byte, packet.NumTokens*SDK4_EncryptedNextRouteTokenSize)
		}
		stream.SerializeBytes(packet.Tokens)
	}

	if packet.RouteType == SDK4_RouteTypeContinue {
		if stream.IsReading() {
			packet.Tokens = make([]byte, packet.NumTokens*SDK4_EncryptedContinueRouteTokenSize)
		}
		stream.SerializeBytes(packet.Tokens)
	}

	stream.SerializeBool(&packet.HasDebug)
	stream.SerializeString(&packet.Debug, SDK4_MaxSessionDebug)

	stream.SerializeBool(&packet.ExcludeNearRelays)
	if packet.ExcludeNearRelays {
		for i := range packet.NearRelayExcluded {
			stream.SerializeBool(&packet.NearRelayExcluded[i])
		}
	}

	stream.SerializeBool(&packet.HighFrequencyPings)

	return stream.Error()
}

// ------------------------------------------------------------

type SDK4_LocationData struct {
	Latitude    float32
	Longitude   float32
	ISP         string
	ASN         uint32
}

func (location *SDK4_LocationData) Read(data []byte) error {
	
	index := 0

	var version uint32
	if !common.ReadUint32(data, &index, &version) {
		return errors.New("invalid read at version number")
	}

	if version > SDK4_LocationVersion {
		return fmt.Errorf("unknown location version: %d", version)
	}

	if !common.ReadFloat32(data, &index, &location.Latitude) {
		return errors.New("invalid read at latitude")
	}

	if !common.ReadFloat32(data, &index, &location.Longitude) {
		return errors.New("invalid read at longitude")
	}

	if !common.ReadString(data, &index, &location.ISP, SDK4_MaxISPNameLength) {
		return errors.New("invalid read at ISP")
	}

	if !common.ReadUint32(data, &index, &location.ASN) {
		return errors.New("invalid read at ASN")
	}

	return nil
}

func (location *SDK4_LocationData) Write(buffer []byte) ([]byte, error) {
	index := 0
	common.WriteUint32(buffer, &index, SDK4_LocationVersion)
	common.WriteFloat32(buffer, &index, location.Latitude)
	common.WriteFloat32(buffer, &index, location.Longitude)
	common.WriteString(buffer, &index, location.ISP, SDK4_MaxISPNameLength)
	common.WriteUint32(buffer, &index, location.ASN)
	return buffer[:index], nil
}

// ------------------------------------------------------------

type SDK4_SessionData struct {
	Version         uint32
	SessionId       uint64
	SessionVersion  uint32
	SliceNumber     uint32
	ExpireTimestamp uint64
	Initial         bool
	// Location                      routing.Location
	RouteChanged                  bool
	RouteNumRelays                int32
	RouteCost                     int32
	RouteRelayIds                 [SDK4_MaxRelaysPerRoute]uint64
	RouteState                    core.RouteState
	EverOnNext                    bool
	FellBackToDirect              bool
	PrevPacketsSentClientToServer uint64
	PrevPacketsSentServerToClient uint64
	PrevPacketsLostClientToServer uint64
	PrevPacketsLostServerToClient uint64
	HoldNearRelays                bool
	HoldNearRelayRTT              [SDK4_MaxNearRelays]int32
	WroteSummary                  bool
	TotalPriceSum                 uint64
	NextEnvelopeBytesUpSum        uint64
	NextEnvelopeBytesDownSum      uint64
	DurationOnNext                uint32
}

func (sessionData *SDK4_SessionData) Serialize(stream common.Stream) error {

	// IMPORTANT: DO NOT CHANGE CODE IN THIS FUNCTION BELOW HERE.

	stream.SerializeBits(&sessionData.Version, 8)

	if sessionData.Version < 8 {
		return errors.New("session data is too old")
	}

	stream.SerializeUint64(&sessionData.SessionId)
	stream.SerializeBits(&sessionData.SessionVersion, 8)

	stream.SerializeUint32(&sessionData.SliceNumber)

	stream.SerializeUint64(&sessionData.ExpireTimestamp)

	stream.SerializeBool(&sessionData.Initial)

	// todo
	/*
		locationSize := uint32(sessionData.Location.Size())
		stream.SerializeUint32(&locationSize)
		if stream.IsReading() {
			locationBytes := make([]byte, locationSize)
			stream.SerializeBytes(locationBytes)
			if err := sessionData.Location.UnmarshalBinary(locationBytes); err != nil {
				return err
			}
		} else {
			locationBytes, err := sessionData.Location.MarshalBinary()
			if err != nil {
				return err
			}
			stream.SerializeBytes(locationBytes)
		}
	*/

	stream.SerializeBool(&sessionData.RouteChanged)

	hasRoute := sessionData.RouteNumRelays > 0

	stream.SerializeBool(&hasRoute)

	stream.SerializeInteger(&sessionData.RouteCost, 0, SDK4_InvalidRouteValue)

	if hasRoute {
		stream.SerializeInteger(&sessionData.RouteNumRelays, 0, SDK4_MaxTokens)
		for i := int32(0); i < sessionData.RouteNumRelays; i++ {
			stream.SerializeUint64(&sessionData.RouteRelayIds[i])
		}
	}

	stream.SerializeUint64(&sessionData.RouteState.UserID)
	stream.SerializeBool(&sessionData.RouteState.Next)
	stream.SerializeBool(&sessionData.RouteState.Veto)
	stream.SerializeBool(&sessionData.RouteState.Banned)
	stream.SerializeBool(&sessionData.RouteState.Disabled)
	stream.SerializeBool(&sessionData.RouteState.NotSelected)
	stream.SerializeBool(&sessionData.RouteState.ABTest)
	stream.SerializeBool(&sessionData.RouteState.A)
	stream.SerializeBool(&sessionData.RouteState.B)
	stream.SerializeBool(&sessionData.RouteState.ForcedNext)
	stream.SerializeBool(&sessionData.RouteState.ReduceLatency)
	stream.SerializeBool(&sessionData.RouteState.ReducePacketLoss)
	stream.SerializeBool(&sessionData.RouteState.ProMode)
	stream.SerializeBool(&sessionData.RouteState.Multipath)
	stream.SerializeBool(&sessionData.RouteState.Committed)
	stream.SerializeBool(&sessionData.RouteState.CommitVeto)
	stream.SerializeInteger(&sessionData.RouteState.CommitCounter, 0, 4)
	stream.SerializeBool(&sessionData.RouteState.LatencyWorse)
	stream.SerializeBool(&sessionData.RouteState.MultipathOverload)
	stream.SerializeBool(&sessionData.RouteState.NoRoute)
	stream.SerializeBool(&sessionData.RouteState.NextLatencyTooHigh)
	stream.SerializeBool(&sessionData.RouteState.Mispredict)
	stream.SerializeBool(&sessionData.EverOnNext)
	stream.SerializeBool(&sessionData.FellBackToDirect)

	stream.SerializeInteger(&sessionData.RouteState.NumNearRelays, 0, core.MaxNearRelays)

	for i := int32(0); i < sessionData.RouteState.NumNearRelays; i++ {
		stream.SerializeInteger(&sessionData.RouteState.NearRelayRTT[i], 0, 255)
		stream.SerializeInteger(&sessionData.RouteState.NearRelayJitter[i], 0, 255)
		nearRelayPLHistory := int32(sessionData.RouteState.NearRelayPLHistory[i])
		stream.SerializeInteger(&nearRelayPLHistory, 0, 255)
		sessionData.RouteState.NearRelayPLHistory[i] = uint32(nearRelayPLHistory)
	}

	directPLHistory := int32(sessionData.RouteState.DirectPLHistory)
	stream.SerializeInteger(&directPLHistory, 0, 255)
	sessionData.RouteState.DirectPLHistory = uint32(directPLHistory)

	stream.SerializeInteger(&sessionData.RouteState.PLHistoryIndex, 0, 7)
	stream.SerializeInteger(&sessionData.RouteState.PLHistorySamples, 0, 8)

	stream.SerializeBool(&sessionData.RouteState.RelayWentAway)
	stream.SerializeBool(&sessionData.RouteState.RouteLost)
	stream.SerializeInteger(&sessionData.RouteState.DirectJitter, 0, 255)

	stream.SerializeUint32(&sessionData.RouteState.DirectPLCount)

	for i := int32(0); i < sessionData.RouteState.NumNearRelays; i++ {
		stream.SerializeUint32(&sessionData.RouteState.NearRelayPLCount[i])
	}

	stream.SerializeBool(&sessionData.RouteState.LackOfDiversity)

	stream.SerializeBits(&sessionData.RouteState.MispredictCounter, 2)

	stream.SerializeBits(&sessionData.RouteState.LatencyWorseCounter, 2)

	stream.SerializeBool(&sessionData.RouteState.MultipathRestricted)

	stream.SerializeUint64(&sessionData.PrevPacketsSentClientToServer)
	stream.SerializeUint64(&sessionData.PrevPacketsSentServerToClient)
	stream.SerializeUint64(&sessionData.PrevPacketsLostClientToServer)
	stream.SerializeUint64(&sessionData.PrevPacketsLostServerToClient)

	stream.SerializeBool(&sessionData.RouteState.LocationVeto)

	stream.SerializeBool(&sessionData.HoldNearRelays)
	if sessionData.HoldNearRelays {
		for i := 0; i < core.MaxNearRelays; i++ {
			stream.SerializeInteger(&sessionData.HoldNearRelayRTT[i], 0, 255)
		}
	}

	stream.SerializeInteger(&sessionData.RouteState.PLSustainedCounter, 0, 3)

	stream.SerializeBool(&sessionData.WroteSummary)

	stream.SerializeUint64(&sessionData.TotalPriceSum)

	stream.SerializeUint64(&sessionData.NextEnvelopeBytesUpSum)
	stream.SerializeUint64(&sessionData.NextEnvelopeBytesDownSum)

	stream.SerializeUint32(&sessionData.DurationOnNext)

	// IMPORTANT: DO NOT CHANGE CODE IN THIS FUNCTION ABOVE HERE

	// ADD NEW FIELDS HERE ONLY. ONCE COMPLETED MOVE THIS COMMENT SECTION BELOW THE NEW FIELDS

	// FAILING TO FOLLOW THIS WILL BREAK PRODUCTION!!!

	// >>> NEW FIELDS GO HERE <<<

	return stream.Error()
}

// ------------------------------------------------------------

type SDK4_MatchDataRequestPacket struct {
	Version        SDKVersion
	BuyerId        uint64
	ServerAddress  net.UDPAddr
	DatacenterId   uint64
	UserHash       uint64
	SessionId      uint64
	RetryNumber    uint32
	MatchId        uint64
	NumMatchValues int32
	MatchValues    [SDK4_MaxMatchValues]float64
}

func (packet *SDK4_MatchDataRequestPacket) Serialize(stream common.Stream) error {

	packet.Version.Serialize(stream)

	stream.SerializeUint64(&packet.BuyerId)
	stream.SerializeAddress(&packet.ServerAddress)
	stream.SerializeUint64(&packet.DatacenterId)
	stream.SerializeUint64(&packet.UserHash)
	stream.SerializeUint64(&packet.SessionId)
	stream.SerializeUint32(&packet.RetryNumber)
	stream.SerializeUint64(&packet.MatchId)

	hasMatchValues := stream.IsWriting() && packet.NumMatchValues > 0

	stream.SerializeBool(&hasMatchValues)

	if hasMatchValues {
		stream.SerializeInteger(&packet.NumMatchValues, 0, SDK4_MaxMatchValues)
		for i := 0; i < int(packet.NumMatchValues); i++ {
			stream.SerializeFloat64(&packet.MatchValues[i])
		}
	}

	return stream.Error()
}

// ------------------------------------------------------------

type SDK4_MatchDataResponsePacket struct {
	SessionId uint64
	Response  uint32
}

func (packet *SDK4_MatchDataResponsePacket) Serialize(stream common.Stream) error {
	stream.SerializeUint64(&packet.SessionId)
	stream.SerializeBits(&packet.Response, 8)
	return stream.Error()
}

// ------------------------------------------------------------
