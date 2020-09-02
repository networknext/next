package transport

import (
	"fmt"
	"net"

	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/encoding"
	"github.com/networknext/backend/routing"
)

const (
	PacketHashMessageSize = 32

	MaxDatacenterNameLength = 256

	MaxSessionDataSize = 511

	PacketTypeServerUpdate4       = 220
	PacketTypeSessionUpdate4      = 221
	PacketTypeSessionResponse4    = 222
	PacketTypeServerInitRequest4  = 223
	PacketTypeServerInitResponse4 = 224
)

type Packet interface {
	Serialize(stream encoding.Stream) error
}

func UnmarshalPacket(packet Packet, data []byte) error {
	if err := packet.Serialize(encoding.CreateReadStream(data)); err != nil {
		return err
	}
	return nil
}

func MarshalPacket(packet Packet) ([]byte, error) {
	ws, err := encoding.CreateWriteStream(DefaultMaxPacketSize)
	if err != nil {
		return nil, err
	}

	if err := packet.Serialize(ws); err != nil {
		return nil, err
	}
	ws.Flush()

	return ws.GetData()[:ws.GetBytesProcessed()], nil
}

type ServerInitRequestPacket4 struct {
	Version        SDKVersion
	RequestID      uint64
	CustomerID     uint64
	DatacenterID   uint64
	DatacenterName string
}

func (packet *ServerInitRequestPacket4) Serialize(stream encoding.Stream) error {
	packetType := uint32(PacketTypeServerInitRequest4)
	stream.SerializeBits(&packetType, 8)

	if packetType != PacketTypeServerInitRequest4 {
		return fmt.Errorf("[ServerInitRequestPacket4] wrong packet type %d, expected %d", packetType, PacketTypeServerInitRequest4)
	}

	versionMajor := uint32(packet.Version.Major)
	versionMinor := uint32(packet.Version.Minor)
	versionPatch := uint32(packet.Version.Patch)
	stream.SerializeBits(&versionMajor, 8)
	stream.SerializeBits(&versionMinor, 8)
	stream.SerializeBits(&versionPatch, 8)
	packet.Version = SDKVersion{int32(versionMajor), int32(versionMinor), int32(versionPatch)}
	stream.SerializeUint64(&packet.RequestID)
	stream.SerializeUint64(&packet.CustomerID)
	stream.SerializeUint64(&packet.DatacenterID)
	stream.SerializeString(&packet.DatacenterName, MaxDatacenterNameLength)
	return stream.Error()
}

type ServerInitResponsePacket4 struct {
	RequestID uint64
	Response  uint32
}

func (packet *ServerInitResponsePacket4) Serialize(stream encoding.Stream) error {
	packetType := uint32(PacketTypeServerInitResponse4)
	stream.SerializeBits(&packetType, 8)

	if packetType != PacketTypeServerInitResponse4 {
		return fmt.Errorf("[ServerInitResponsePacket4] wrong packet type %d, expected %d", packetType, PacketTypeServerInitResponse4)
	}

	stream.SerializeUint64(&packet.RequestID)
	stream.SerializeUint32(&packet.Response)
	return stream.Error()
}

type ServerUpdatePacket4 struct {
	Version       SDKVersion
	Sequence      uint64
	CustomerID    uint64
	DatacenterID  uint64
	NumSessions   uint32
	ServerAddress net.UDPAddr
}

func (packet *ServerUpdatePacket4) Serialize(stream encoding.Stream) error {
	packetType := uint32(PacketTypeServerUpdate4)
	stream.SerializeBits(&packetType, 8)

	if packetType != PacketTypeServerUpdate4 {
		return fmt.Errorf("[ServerUpdatePacket4] wrong packet type %d, expected %d", packetType, PacketTypeServerUpdate4)
	}

	versionMajor := uint32(packet.Version.Major)
	versionMinor := uint32(packet.Version.Minor)
	versionPatch := uint32(packet.Version.Patch)
	stream.SerializeBits(&versionMajor, 8)
	stream.SerializeBits(&versionMinor, 8)
	stream.SerializeBits(&versionPatch, 8)
	packet.Version = SDKVersion{int32(versionMajor), int32(versionMinor), int32(versionPatch)}
	stream.SerializeUint64(&packet.Sequence)
	stream.SerializeUint64(&packet.CustomerID)
	stream.SerializeUint64(&packet.DatacenterID)
	stream.SerializeUint32(&packet.NumSessions)
	stream.SerializeAddress(&packet.ServerAddress)
	return stream.Error()
}

type SessionUpdatePacket4 struct {
	Version                   SDKVersion
	Sequence                  uint64
	CustomerID                uint64
	ServerAddress             net.UDPAddr
	SessionID                 uint64
	UserHash                  uint64
	PlatformID                int32
	Tag                       uint64
	Flags                     uint32
	Flagged                   bool
	FallbackToDirect          bool
	ConnectionType            int32
	OnNetworkNext             bool
	Committed                 bool
	DirectRTT                 float32
	DirectJitter              float32
	DirectPacketLoss          float32
	NextRTT                   float32
	NextJitter                float32
	NextPacketLoss            float32
	NumNearRelays             int32
	NearRelayIDs              []uint64
	NearRelayRTT              []float32
	NearRelayJitter           []float32
	NearRelayPacketLoss       []float32
	ClientAddress             net.UDPAddr
	ClientRoutePublicKey      []byte
	ServerRoutePublicKey      []byte
	KbpsUp                    uint32
	KbpsDown                  uint32
	PacketsSentClientToServer uint64
	PacketsSentServerToClient uint64
	PacketsLostClientToServer uint64
	PacketsLostServerToClient uint64
	UserFlags                 uint64
	SessionDataBytes          int32
	SessionData               [MaxSessionDataSize]byte
}

func (packet *SessionUpdatePacket4) Serialize(stream encoding.Stream) error {
	packetType := uint32(PacketTypeSessionUpdate4)
	stream.SerializeBits(&packetType, 8)

	if packetType != PacketTypeSessionUpdate4 {
		return fmt.Errorf("[SessionUpdatePacket4] wrong packet type %d, expected %d", packetType, PacketTypeSessionUpdate4)
	}

	versionMajor := uint32(packet.Version.Major)
	versionMinor := uint32(packet.Version.Minor)
	versionPatch := uint32(packet.Version.Patch)
	stream.SerializeBits(&versionMajor, 8)
	stream.SerializeBits(&versionMinor, 8)
	stream.SerializeBits(&versionPatch, 8)
	packet.Version = SDKVersion{int32(versionMajor), int32(versionMinor), int32(versionPatch)}
	stream.SerializeUint64(&packet.Sequence)
	stream.SerializeUint64(&packet.CustomerID)
	stream.SerializeAddress(&packet.ServerAddress)
	stream.SerializeUint64(&packet.SessionID)
	stream.SerializeUint64(&packet.UserHash)
	stream.SerializeInteger(&packet.PlatformID, PlatformTypeUnknown, PlatformTypeMax)
	stream.SerializeUint64(&packet.Tag)
	hasFlags := stream.IsWriting() && packet.Flags != 0
	stream.SerializeBool(&hasFlags)
	if hasFlags {
		stream.SerializeBits(&packet.Flags, 9)
	}
	stream.SerializeBool(&packet.Flagged)
	stream.SerializeBool(&packet.FallbackToDirect)
	stream.SerializeInteger(&packet.ConnectionType, ConnectionTypeUnknown, ConnectionTypeMax)
	stream.SerializeFloat32(&packet.DirectRTT)
	stream.SerializeFloat32(&packet.DirectJitter)
	stream.SerializeFloat32(&packet.DirectPacketLoss)
	stream.SerializeBool(&packet.OnNetworkNext)
	stream.SerializeBool(&packet.Committed)
	if packet.OnNetworkNext {
		stream.SerializeFloat32(&packet.NextRTT)
		stream.SerializeFloat32(&packet.NextJitter)
		stream.SerializeFloat32(&packet.NextPacketLoss)
	}
	stream.SerializeInteger(&packet.NumNearRelays, 0, MaxNearRelays)
	if stream.IsReading() {
		packet.NearRelayIDs = make([]uint64, packet.NumNearRelays)
		packet.NearRelayRTT = make([]float32, packet.NumNearRelays)
		packet.NearRelayJitter = make([]float32, packet.NumNearRelays)
		packet.NearRelayPacketLoss = make([]float32, packet.NumNearRelays)
	}
	var i int32
	for i = 0; i < packet.NumNearRelays; i++ {
		stream.SerializeUint64(&packet.NearRelayIDs[i])
		stream.SerializeFloat32(&packet.NearRelayRTT[i])
		stream.SerializeFloat32(&packet.NearRelayJitter[i])
		stream.SerializeFloat32(&packet.NearRelayPacketLoss[i])
	}
	stream.SerializeAddress(&packet.ClientAddress)
	if stream.IsReading() {
		packet.ClientRoutePublicKey = make([]byte, crypto.KeySize)
		packet.ServerRoutePublicKey = make([]byte, crypto.KeySize)
	}
	stream.SerializeBytes(packet.ClientRoutePublicKey)
	stream.SerializeBytes(packet.ServerRoutePublicKey)
	stream.SerializeUint32(&packet.KbpsUp)
	stream.SerializeUint32(&packet.KbpsDown)
	stream.SerializeUint64(&packet.PacketsSentClientToServer)
	stream.SerializeUint64(&packet.PacketsSentServerToClient)
	stream.SerializeUint64(&packet.PacketsLostClientToServer)
	stream.SerializeUint64(&packet.PacketsLostServerToClient)
	hasUserFlags := stream.IsWriting() && packet.UserFlags != 0
	stream.SerializeBool(&hasUserFlags)
	if hasUserFlags {
		stream.SerializeUint64(&packet.UserFlags)
	}
	stream.SerializeInteger(&packet.SessionDataBytes, 0, MaxSessionDataSize)
	if packet.SessionDataBytes > 0 {
		sessionData := packet.SessionData[:packet.SessionDataBytes]
		stream.SerializeBytes(sessionData)
	}
	return stream.Error()
}

type SessionResponsePacket4 struct {
	Sequence             uint64
	SessionID            uint64
	NumNearRelays        int32
	NearRelayIDs         []uint64
	NearRelayAddresses   []net.UDPAddr
	RouteType            int32
	Multipath            bool
	Committed            bool
	NumTokens            int32
	Tokens               []byte
	ServerRoutePublicKey []byte
	SessionDataBytes     int32
	SessionData          [MaxSessionDataSize]byte
}

func (packet *SessionResponsePacket4) Serialize(stream encoding.Stream) error {
	packetType := uint32(PacketTypeSessionResponse4)
	stream.SerializeBits(&packetType, 8)

	if packetType != PacketTypeSessionResponse4 {
		return fmt.Errorf("[SessionResponsePacket4] wrong packet type %d, expected %d", packetType, PacketTypeSessionResponse4)
	}

	stream.SerializeUint64(&packet.Sequence)
	stream.SerializeUint64(&packet.SessionID)
	stream.SerializeInteger(&packet.NumNearRelays, 0, MaxNearRelays)
	if stream.IsReading() {
		packet.NearRelayIDs = make([]uint64, packet.NumNearRelays)
		packet.NearRelayAddresses = make([]net.UDPAddr, packet.NumNearRelays)
	}
	for i := 0; i < int(packet.NumNearRelays); i++ {
		stream.SerializeUint64(&packet.NearRelayIDs[i])
		stream.SerializeAddress(&packet.NearRelayAddresses[i])
	}
	stream.SerializeInteger(&packet.RouteType, routing.RouteTypeDirect, routing.RouteTypeContinue)
	if packet.RouteType != routing.RouteTypeDirect {
		stream.SerializeBool(&packet.Multipath)
		stream.SerializeBool(&packet.Committed)
		stream.SerializeInteger(&packet.NumTokens, 0, MaxTokens)
	}
	if packet.RouteType == routing.RouteTypeNew {
		if stream.IsReading() {
			packet.Tokens = make([]byte, packet.NumTokens*routing.EncryptedNextRouteTokenSize4)
		}
		stream.SerializeBytes(packet.Tokens)
	}
	if packet.RouteType == routing.RouteTypeContinue {
		if stream.IsReading() {
			packet.Tokens = make([]byte, packet.NumTokens*routing.EncryptedContinueRouteTokenSize4)
		}
		stream.SerializeBytes(packet.Tokens)
	}
	if stream.IsReading() {
		packet.ServerRoutePublicKey = make([]byte, crypto.KeySize)
	}
	stream.SerializeBytes(packet.ServerRoutePublicKey)
	stream.SerializeInteger(&packet.SessionDataBytes, 0, MaxSessionDataSize)
	if packet.SessionDataBytes > 0 {
		sessionData := packet.SessionData[:packet.SessionDataBytes]
		stream.SerializeBytes(sessionData)
	}
	return stream.Error()
}
