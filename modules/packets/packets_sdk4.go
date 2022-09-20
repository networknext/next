package packets

import (
	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/crypto"
	"net"
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

type SessionResponsePacket struct {
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

func (packet *SessionResponsePacket) Serialize(stream common.Stream) error {

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
