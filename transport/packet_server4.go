package transport

import (
	"errors"
	"fmt"
	"net"

	"github.com/networknext/backend/encoding"
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

type ServerInitRequestPacket4 struct {
	Version        SDKVersion
	RequestID      uint64
	CustomerID     uint64
	DatacenterID   uint64
	DatacenterName string
}

func (packet *ServerInitRequestPacket4) UnmarshalBinary(data []byte) error {
	var index int

	var packetType uint8
	if !encoding.ReadUint8(data, &index, &packetType) {
		return errors.New("[ServerInitRequestPacket4] failed to read packet type")
	}

	if packetType != PacketTypeServerInitRequest4 {
		return fmt.Errorf("[ServerInitRequestPacket4] wrong packet type %d, expected %d", packetType, PacketTypeServerInitRequest4)
	}

	var versionMajor uint8
	if !encoding.ReadUint8(data, &index, &versionMajor) {
		return errors.New("[ServerInitRequestPacket4] failed to read version major")
	}

	var versionMinor uint8
	if !encoding.ReadUint8(data, &index, &versionMinor) {
		return errors.New("[ServerInitRequestPacket4] failed to read version minor")
	}

	var versionPatch uint8
	if !encoding.ReadUint8(data, &index, &versionPatch) {
		return errors.New("[ServerInitRequestPacket4] failed to read version patch")
	}

	packet.Version = SDKVersion{
		Major: int32(versionMajor),
		Minor: int32(versionMinor),
		Patch: int32(versionPatch),
	}

	if !encoding.ReadUint64(data, &index, &packet.RequestID) {
		return errors.New("[ServerInitRequestPacket4] failed to read request ID")
	}

	if !encoding.ReadUint64(data, &index, &packet.CustomerID) {
		return errors.New("[ServerInitRequestPacket4] failed to read customer ID")
	}

	if !encoding.ReadUint64(data, &index, &packet.DatacenterID) {
		return errors.New("[ServerInitRequestPacket4] failed to read datacenter ID")
	}

	var datacenterNameLength uint8
	if !encoding.ReadUint8(data, &index, &datacenterNameLength) {
		return errors.New("[ServerInitRequestPacket4] failed to read datacenter name length")
	}

	var datacenterNameBytes []byte
	if !encoding.ReadBytes(data, &index, &datacenterNameBytes, uint32(datacenterNameLength)) {
		return errors.New("[ServerInitRequestPacket4] failed to read datacenter name")
	}
	packet.DatacenterName = string(datacenterNameBytes)

	return nil
}

func (packet ServerInitRequestPacket4) MarshalBinary() ([]byte, error) {
	data := make([]byte, packet.Size())
	var index int

	encoding.WriteUint8(data, &index, PacketTypeServerInitRequest4)
	encoding.WriteUint8(data, &index, uint8(packet.Version.Major))
	encoding.WriteUint8(data, &index, uint8(packet.Version.Minor))
	encoding.WriteUint8(data, &index, uint8(packet.Version.Patch))
	encoding.WriteUint64(data, &index, packet.RequestID)
	encoding.WriteUint64(data, &index, packet.CustomerID)
	encoding.WriteUint64(data, &index, packet.DatacenterID)

	encoding.WriteUint8(data, &index, uint8(len(packet.DatacenterName)))
	encoding.WriteBytes(data, &index, []byte(packet.DatacenterName), len(packet.DatacenterName))

	return data, nil
}

func (packet ServerInitRequestPacket4) Size() uint64 {
	return uint64(1 + 1*3 + 8 + 8 + 8 + 4 + len(packet.DatacenterName))
}

type ServerInitResponsePacket4 struct {
	RequestID uint64
	Response  uint32
}

func (packet *ServerInitResponsePacket4) UnmarshalBinary(data []byte) error {
	var index int

	var packetType uint8
	if !encoding.ReadUint8(data, &index, &packetType) {
		return errors.New("[ServerInitResponsePacket4] failed to read packet type")
	}

	if packetType != PacketTypeServerInitResponse4 {
		return fmt.Errorf("[ServerInitResponsePacket4] wrong packet type %d, expected %d", packetType, PacketTypeServerInitResponse4)
	}

	if !encoding.ReadUint64(data, &index, &packet.RequestID) {
		return errors.New("[ServerInitResponsePacket4] failed to read request ID")
	}

	if !encoding.ReadUint32(data, &index, &packet.Response) {
		return errors.New("[ServerInitResponsePacket4] failed to read response code")
	}

	return nil
}

func (packet ServerInitResponsePacket4) MarshalBinary() ([]byte, error) {
	data := make([]byte, packet.Size())
	var index int

	encoding.WriteUint8(data, &index, PacketTypeServerInitResponse4)
	encoding.WriteUint64(data, &index, packet.RequestID)
	encoding.WriteUint32(data, &index, packet.Response)

	return data, nil
}

func (packet ServerInitResponsePacket4) Size() uint64 {
	return 1 + 8 + 4
}

type ServerUpdatePacket4 struct {
	Version       SDKVersion
	Sequence      uint64
	CustomerID    uint64
	DatacenterID  uint64
	NumSessions   uint32
	ServerAddress net.UDPAddr
}

func (packet *ServerUpdatePacket4) UnmarshalBinary(data []byte) error {
	var index int

	var packetType uint8
	if !encoding.ReadUint8(data, &index, &packetType) {
		return errors.New("[ServerUpdatePacket4] failed to read packet type")
	}

	if packetType != PacketTypeServerUpdate4 {
		return fmt.Errorf("[ServerUpdatePacket4] wrong packet type %d, expected %d", packetType, PacketTypeServerUpdate4)
	}

	var versionMajor uint8
	if !encoding.ReadUint8(data, &index, &versionMajor) {
		return errors.New("[ServerUpdatePacket4] failed to read version major")
	}

	var versionMinor uint8
	if !encoding.ReadUint8(data, &index, &versionMinor) {
		return errors.New("[ServerUpdatePacket4] failed to read version minor")
	}

	var versionPatch uint8
	if !encoding.ReadUint8(data, &index, &versionPatch) {
		return errors.New("[ServerUpdatePacket4] failed to read version patch")
	}

	packet.Version = SDKVersion{
		Major: int32(versionMajor),
		Minor: int32(versionMinor),
		Patch: int32(versionPatch),
	}

	if !encoding.ReadUint64(data, &index, &packet.Sequence) {
		return errors.New("[ServerUpdatePacket4] failed to read sequence number")
	}

	if !encoding.ReadUint64(data, &index, &packet.CustomerID) {
		return errors.New("[ServerUpdatePacket4] failed to read customer ID")
	}

	if !encoding.ReadUint64(data, &index, &packet.DatacenterID) {
		return errors.New("[ServerUpdatePacket4] failed to read datacenter ID")
	}

	if !encoding.ReadUint32(data, &index, &packet.NumSessions) {
		return errors.New("[ServerUpdatePacket4] failed to read number of sessions")
	}

	packet.ServerAddress = *encoding.ReadAddress(data[index:])
	return nil
}

func (packet ServerUpdatePacket4) MarshalBinary() ([]byte, error) {
	data := make([]byte, packet.Size())
	var index int

	encoding.WriteUint8(data, &index, PacketTypeServerUpdate4)
	encoding.WriteUint8(data, &index, uint8(packet.Version.Major))
	encoding.WriteUint8(data, &index, uint8(packet.Version.Minor))
	encoding.WriteUint8(data, &index, uint8(packet.Version.Patch))
	encoding.WriteUint64(data, &index, packet.Sequence)
	encoding.WriteUint64(data, &index, packet.CustomerID)
	encoding.WriteUint64(data, &index, packet.DatacenterID)
	encoding.WriteUint32(data, &index, packet.NumSessions)
	encoding.WriteAddress(data[index:], &packet.ServerAddress)

	return data, nil
}

func (packet ServerUpdatePacket4) Size() uint64 {
	return 1 + 1*3 + 8 + 8 + 8 + 4 + 19
}

type SessionUpdatePacket4 struct {
	VersionMajor              uint32
	VersionMinor              uint32
	VersionPatch              uint32
	Sequence                  uint64
	CustomerID                uint64
	SessionID                 uint64
	UserHash                  uint64
	PlatformID                int32
	Tag                       uint64
	Flags                     uint32
	Flagged                   bool
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
	ServerAddress             net.UDPAddr
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
