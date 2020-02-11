package transport

import (
	"bytes"
	"crypto/ed25519"
	"encoding/binary"
	"net"

	"github.com/networknext/backend/encoding"
	"github.com/networknext/backend/routing"
)

const (
	DefaultMaxPacketSize = 1500

	PacketTypeServerUpdate    = 200
	PacketTypeSessionUpdate   = 201
	PacketTypeSessionResponse = 202

	MaxNearRelays = 32
	MaxTokens     = 7

	// EncryptedTokenRouteSize    = 117
	// EncryptedTokenContinueSize = 58
	MTUSize = 1300

	ConnectionTypeUnknown  = 0
	ConnectionTypeWired    = 1
	ConnectionTypeWifi     = 2
	ConnectionTypeCellular = 3

	PlatformUnknown = 0
	PlatformWindows = 1
	PlatformMac     = 2
	PlatformUnix    = 3
	PlatformSwitch  = 4
	PlatformPS4     = 5
	PlatformIOS     = 6
	PlatformXboxOne = 7

	RouteSliceFlagNext                = (uint64(1) << 1)
	RouteSliceFlagReported            = (uint64(1) << 2)
	RouteSliceFlagVetoed              = (uint64(1) << 3)
	RouteSliceFlagFallbackToDirect    = (uint64(1) << 4)
	RouteSliceFlagPacketLossMultipath = (uint64(1) << 5)
	RouteSliceFlagJitterMultipath     = (uint64(1) << 6)
	RouteSliceFlagRTTMultipath        = (uint64(1) << 7)

	FlagBadRouteToken           = uint32(1 << 0)
	FlagNoRouteToContinue       = uint32(1 << 1)
	FlagPreviousUpdatePending   = uint32(1 << 2)
	FlagBadContinueToken        = uint32(1 << 3)
	FlagRouteExpired            = uint32(1 << 4)
	FlagRouteRequestTimedOut    = uint32(1 << 5)
	FlagContinueRequestTimedOut = uint32(1 << 6)
	FlagClientTimedOut          = uint32(1 << 7)
	FlagTryBeforeYouBuyAbort    = uint32(1 << 8)
	FlagDirectRouteExpired      = uint32(1 << 9)
	FlagTotalCount              = 10

	AddressSize = 19
)

type ServerUpdatePacket struct {
	Sequence             uint64
	VersionMajor         int32
	VersionMinor         int32
	VersionPatch         int32
	CustomerId           uint64
	DatacenterId         uint64
	NumSessionsPending   uint32
	NumSessionsUpgraded  uint32
	ServerAddress        net.UDPAddr
	ServerPrivateAddress net.UDPAddr
	ServerRoutePublicKey []byte
	Signature            []byte
}

func (packet *ServerUpdatePacket) UnmarshalBinary(data []byte) error {
	if err := packet.Serialize(encoding.CreateReadStream(data)); err != nil {
		return err
	}
	return nil
}

func (packet *ServerUpdatePacket) MarshalBinary() ([]byte, error) {
	ws, err := encoding.CreateWriteStream(DefaultMaxPacketSize)
	if err != nil {
		return nil, err
	}

	if err := packet.Serialize(ws); err != nil {
		return nil, err
	}
	ws.Flush()

	return ws.GetData(), nil
}

func (packet *ServerUpdatePacket) Serialize(stream encoding.Stream) error {
	packetType := uint32(PacketTypeServerUpdate)
	stream.SerializeBits(&packetType, 8)

	stream.SerializeUint64(&packet.Sequence)
	stream.SerializeInteger(&packet.VersionMajor, 0, SDKVersionMax.Major)
	stream.SerializeInteger(&packet.VersionMinor, 0, SDKVersionMax.Minor)
	stream.SerializeInteger(&packet.VersionPatch, 0, SDKVersionMax.Patch)
	stream.SerializeUint64(&packet.CustomerId)
	stream.SerializeUint64(&packet.DatacenterId)
	stream.SerializeUint32(&packet.NumSessionsPending)
	stream.SerializeUint32(&packet.NumSessionsUpgraded)
	stream.SerializeAddress(&packet.ServerAddress)
	stream.SerializeAddress(&packet.ServerPrivateAddress)
	if stream.IsReading() {
		packet.ServerRoutePublicKey = make([]byte, ed25519.PublicKeySize)
		packet.Signature = make([]byte, ed25519.SignatureSize)
	}
	stream.SerializeBytes(packet.ServerRoutePublicKey)
	stream.SerializeBytes(packet.Signature)
	return stream.Error()
}

func (packet *ServerUpdatePacket) GetSignData() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, packet.Sequence)
	binary.Write(buf, binary.LittleEndian, uint64(packet.VersionMajor))
	binary.Write(buf, binary.LittleEndian, uint64(packet.VersionMinor))
	binary.Write(buf, binary.LittleEndian, uint64(packet.VersionPatch))
	binary.Write(buf, binary.LittleEndian, packet.CustomerId)
	binary.Write(buf, binary.LittleEndian, packet.DatacenterId)
	binary.Write(buf, binary.LittleEndian, packet.NumSessionsPending)
	binary.Write(buf, binary.LittleEndian, packet.NumSessionsUpgraded)

	address := make([]byte, AddressSize)
	encoding.WriteAddress(address, &packet.ServerAddress)
	binary.Write(buf, binary.LittleEndian, address)

	privateAddress := make([]byte, AddressSize)
	encoding.WriteAddress(privateAddress, &packet.ServerPrivateAddress)
	binary.Write(buf, binary.LittleEndian, privateAddress)

	binary.Write(buf, binary.LittleEndian, packet.ServerRoutePublicKey)
	return buf.Bytes()
}

type SessionUpdatePacket struct {
	Sequence                  uint64
	CustomerId                uint64
	SessionId                 uint64
	UserHash                  uint64
	PlatformId                uint64
	Tag                       uint64
	Flags                     uint32
	Flagged                   bool
	FallbackToDirect          bool
	TryBeforeYouBuy           bool     	// removed in SDK 3.3.5
	ConnectionType            int32
	OnNetworkNext             bool
	DirectMinRtt              float32
	DirectMaxRtt              float32
	DirectMeanRtt             float32
	DirectJitter              float32
	DirectPacketLoss          float32
	NextMinRtt                float32
	NextMaxRtt                float32
	NextMeanRtt               float32
	NextJitter                float32
	NextPacketLoss            float32
	NumNearRelays             int32
	NearRelayIds              []uint64
	NearRelayMinRtt           []float32
	NearRelayMaxRtt           []float32
	NearRelayMeanRtt          []float32
	NearRelayJitter           []float32
	NearRelayPacketLoss       []float32
	ClientAddress             net.UDPAddr
	ServerAddress             net.UDPAddr
	ClientRoutePublicKey      []byte
	KbpsUp                    uint32
	KbpsDown                  uint32
	PacketsLostClientToServer uint64
	PacketsLostServerToClient uint64
	Signature                 []byte
}

func (packet *SessionUpdatePacket) UnmarshalBinary(data []byte) error {
	if err := packet.Serialize(encoding.CreateReadStream(data), SDKVersionMin); err != nil {
		return err
	}
	return nil
}

func (packet *SessionUpdatePacket) MarshalBinary() ([]byte, error) {
	ws, err := encoding.CreateWriteStream(DefaultMaxPacketSize)
	if err != nil {
		return nil, err
	}

	if err := packet.Serialize(ws, SDKVersionMin); err != nil {
		return nil, err
	}
	ws.Flush()

	return ws.GetData(), nil
}

func (packet *SessionUpdatePacket) Serialize(stream encoding.Stream, version SDKVersion) error {
	packetType := uint32(PacketTypeSessionUpdate)
	stream.SerializeBits(&packetType, 8)

	stream.SerializeUint64(&packet.Sequence)
	stream.SerializeUint64(&packet.CustomerId)
	stream.SerializeAddress(&packet.ServerAddress)
	stream.SerializeUint64(&packet.SessionId)
	stream.SerializeUint64(&packet.UserHash)
	stream.SerializeUint64(&packet.PlatformId)
	stream.SerializeUint64(&packet.Tag)
	if version.Compare(SDKVersion{3, 3, 4}) == SDKVersionEqual ||
		version.Compare(SDKVersion{3, 3, 4}) == SDKVersionNewer {
		stream.SerializeBits(&packet.Flags, FlagTotalCount)
	}
	stream.SerializeBool(&packet.Flagged)
	stream.SerializeBool(&packet.FallbackToDirect)
	if version.Compare(SDKVersion{3, 3, 5}) == SDKVersionOlder {
		stream.SerializeBool(&packet.TryBeforeYouBuy)
	}
	stream.SerializeInteger(&packet.ConnectionType, ConnectionTypeUnknown, ConnectionTypeCellular)
	stream.SerializeFloat32(&packet.DirectMinRtt)
	stream.SerializeFloat32(&packet.DirectMaxRtt)
	stream.SerializeFloat32(&packet.DirectMeanRtt)
	stream.SerializeFloat32(&packet.DirectJitter)
	stream.SerializeFloat32(&packet.DirectPacketLoss)
	stream.SerializeBool(&packet.OnNetworkNext)
	if packet.OnNetworkNext {
		stream.SerializeFloat32(&packet.NextMinRtt)
		stream.SerializeFloat32(&packet.NextMaxRtt)
		stream.SerializeFloat32(&packet.NextMeanRtt)
		stream.SerializeFloat32(&packet.NextJitter)
		stream.SerializeFloat32(&packet.NextPacketLoss)
	}
	stream.SerializeInteger(&packet.NumNearRelays, 0, MaxNearRelays)
	if stream.IsReading() {
		packet.NearRelayIds = make([]uint64, packet.NumNearRelays)
		packet.NearRelayMinRtt = make([]float32, packet.NumNearRelays)
		packet.NearRelayMaxRtt = make([]float32, packet.NumNearRelays)
		packet.NearRelayMeanRtt = make([]float32, packet.NumNearRelays)
		packet.NearRelayJitter = make([]float32, packet.NumNearRelays)
		packet.NearRelayPacketLoss = make([]float32, packet.NumNearRelays)
	}
	var i int32
	for i = 0; i < packet.NumNearRelays; i++ {
		stream.SerializeUint64(&packet.NearRelayIds[i])
		stream.SerializeFloat32(&packet.NearRelayMinRtt[i])
		stream.SerializeFloat32(&packet.NearRelayMaxRtt[i])
		stream.SerializeFloat32(&packet.NearRelayMeanRtt[i])
		stream.SerializeFloat32(&packet.NearRelayJitter[i])
		stream.SerializeFloat32(&packet.NearRelayPacketLoss[i])
	}
	stream.SerializeAddress(&packet.ClientAddress)
	if stream.IsReading() {
		packet.ClientRoutePublicKey = make([]byte, ed25519.PublicKeySize)
		packet.Signature = make([]byte, ed25519.SignatureSize)
	}
	stream.SerializeBytes(packet.ClientRoutePublicKey)
	stream.SerializeUint32(&packet.KbpsUp)
	stream.SerializeUint32(&packet.KbpsDown)
	if version.Compare(SDKVersionMin) == SDKVersionEqual ||
		version.Compare(SDKVersionMin) == SDKVersionNewer {
		stream.SerializeUint64(&packet.PacketsLostClientToServer)
		stream.SerializeUint64(&packet.PacketsLostServerToClient)
	}
	stream.SerializeBytes(packet.Signature)
	return stream.Error()
}

func (packet *SessionUpdatePacket) HeaderSerialize(stream encoding.Stream) error {
	stream.SerializeUint64(&packet.Sequence)
	stream.SerializeUint64(&packet.CustomerId)
	stream.SerializeAddress(&packet.ServerAddress)
	return stream.Error()
}

func (packet *SessionUpdatePacket) GetSignData(version SDKVersion) []byte {

	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, packet.Sequence)
	binary.Write(buf, binary.LittleEndian, packet.CustomerId)
	binary.Write(buf, binary.LittleEndian, packet.SessionId)
	binary.Write(buf, binary.LittleEndian, packet.UserHash)
	binary.Write(buf, binary.LittleEndian, packet.PlatformId)
	binary.Write(buf, binary.LittleEndian, packet.Tag)

	if version.IsInternal() ||
		version.Compare(SDKVersion{3, 3, 4}) == SDKVersionEqual ||
		version.Compare(SDKVersion{3, 3, 4}) == SDKVersionNewer {
		binary.Write(buf, binary.LittleEndian, packet.Flags)
	}
	binary.Write(buf, binary.LittleEndian, packet.Flagged)
	binary.Write(buf, binary.LittleEndian, packet.FallbackToDirect)
	if version.Compare(SDKVersion{3, 3, 5}) == SDKVersionOlder {
		binary.Write(buf, binary.LittleEndian, packet.TryBeforeYouBuy)
	}
	binary.Write(buf, binary.LittleEndian, uint8(packet.ConnectionType))

	var onNetworkNext uint8
	onNetworkNext = 0
	if packet.OnNetworkNext {
		onNetworkNext = 1
	}

	binary.Write(buf, binary.LittleEndian, onNetworkNext)

	binary.Write(buf, binary.LittleEndian, packet.DirectMinRtt)
	binary.Write(buf, binary.LittleEndian, packet.DirectMaxRtt)
	binary.Write(buf, binary.LittleEndian, packet.DirectMeanRtt)
	binary.Write(buf, binary.LittleEndian, packet.DirectJitter)
	binary.Write(buf, binary.LittleEndian, packet.DirectPacketLoss)

	binary.Write(buf, binary.LittleEndian, packet.NextMinRtt)
	binary.Write(buf, binary.LittleEndian, packet.NextMaxRtt)
	binary.Write(buf, binary.LittleEndian, packet.NextMeanRtt)
	binary.Write(buf, binary.LittleEndian, packet.NextJitter)
	binary.Write(buf, binary.LittleEndian, packet.NextPacketLoss)

	binary.Write(buf, binary.LittleEndian, uint32(packet.NumNearRelays))
	var i int32
	for i = 0; i < packet.NumNearRelays; i++ {
		binary.Write(buf, binary.LittleEndian, packet.NearRelayIds[i])
		binary.Write(buf, binary.LittleEndian, packet.NearRelayMinRtt[i])
		binary.Write(buf, binary.LittleEndian, packet.NearRelayMaxRtt[i])
		binary.Write(buf, binary.LittleEndian, packet.NearRelayMeanRtt[i])
		binary.Write(buf, binary.LittleEndian, packet.NearRelayJitter[i])
		binary.Write(buf, binary.LittleEndian, packet.NearRelayPacketLoss[i])
	}

	clientAddress := make([]byte, AddressSize)
	encoding.WriteAddress(clientAddress, &packet.ClientAddress)
	binary.Write(buf, binary.LittleEndian, clientAddress)

	serverAddress := make([]byte, AddressSize)
	encoding.WriteAddress(serverAddress, &packet.ServerAddress)
	binary.Write(buf, binary.LittleEndian, serverAddress)

	binary.Write(buf, binary.LittleEndian, packet.KbpsUp)
	binary.Write(buf, binary.LittleEndian, packet.KbpsDown)

	if version.IsInternal() ||
		version.Compare(SDKVersion{3, 3, 4}) == SDKVersionEqual ||
		version.Compare(SDKVersion{3, 3, 4}) == SDKVersionNewer {
		binary.Write(buf, binary.LittleEndian, packet.PacketsLostClientToServer)
		binary.Write(buf, binary.LittleEndian, packet.PacketsLostServerToClient)
	}

	binary.Write(buf, binary.LittleEndian, packet.ClientRoutePublicKey)

	return buf.Bytes()
}

type SessionResponsePacket struct {
	Sequence             uint64
	SessionId            uint64
	NumNearRelays        int32
	NearRelayIds         []uint64
	NearRelayAddresses   []net.UDPAddr
	RouteType            int32
	Multipath            bool
	NumTokens            int32
	Tokens               []byte
	ServerRoutePublicKey []byte
	Signature            []byte
}

func (packet *SessionResponsePacket) UnmarshalBinary(data []byte) error {
	if err := packet.Serialize(encoding.CreateReadStream(data), SDKVersionMin); err != nil {
		return err
	}
	return nil
}

func (packet *SessionResponsePacket) MarshalBinary() ([]byte, error) {
	ws, err := encoding.CreateWriteStream(DefaultMaxPacketSize)
	if err != nil {
		return nil, err
	}

	if err := packet.Serialize(ws, SDKVersionMin); err != nil {
		return nil, err
	}
	ws.Flush()

	return ws.GetData(), nil
}

func (packet *SessionResponsePacket) Serialize(stream encoding.Stream, version SDKVersion) error {
	packetType := uint32(PacketTypeSessionResponse)
	stream.SerializeBits(&packetType, 8)

	stream.SerializeUint64(&packet.Sequence)
	stream.SerializeUint64(&packet.SessionId)
	stream.SerializeInteger(&packet.NumNearRelays, 0, MaxNearRelays)
	if stream.IsReading() {
		packet.NearRelayIds = make([]uint64, packet.NumNearRelays)
		packet.NearRelayAddresses = make([]net.UDPAddr, packet.NumNearRelays)
	}
	var i int32
	for i = 0; i < packet.NumNearRelays; i++ {
		stream.SerializeUint64(&packet.NearRelayIds[i])
		stream.SerializeAddress(&packet.NearRelayAddresses[i])
	}
	stream.SerializeInteger(&packet.RouteType, 0, routing.DecisionTypeContinue)
	if packet.RouteType != routing.DecisionTypeDirect {
		stream.SerializeBool(&packet.Multipath)
		stream.SerializeInteger(&packet.NumTokens, 0, MaxTokens)
	}
	if stream.IsReading() {
		packet.Tokens = make([]byte, packet.NumTokens*routing.EncryptedNextRouteTokenSize)
	}
	if packet.RouteType == routing.DecisionTypeNew {
		stream.SerializeBytes(packet.Tokens)
	}
	if packet.RouteType == routing.DecisionTypeContinue {
		stream.SerializeBytes(packet.Tokens)
	}
	if stream.IsReading() {
		packet.ServerRoutePublicKey = make([]byte, ed25519.PublicKeySize)
		packet.Signature = make([]byte, ed25519.SignatureSize)
	}
	stream.SerializeBytes(packet.ServerRoutePublicKey)
	stream.SerializeBytes(packet.Signature)

	return stream.Error()
}

func (packet *SessionResponsePacket) GetSignData() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, packet.Sequence)
	binary.Write(buf, binary.LittleEndian, packet.SessionId)
	binary.Write(buf, binary.LittleEndian, uint8(packet.NumNearRelays))
	var i int32
	for i = 0; i < packet.NumNearRelays; i++ {
		binary.Write(buf, binary.LittleEndian, packet.NearRelayIds[i])
		address := make([]byte, AddressSize)
		encoding.WriteAddress(address, &packet.NearRelayAddresses[i])
		binary.Write(buf, binary.LittleEndian, address)
	}
	binary.Write(buf, binary.LittleEndian, uint8(packet.RouteType))
	if packet.RouteType != routing.DecisionTypeDirect {
		if packet.Multipath {
			binary.Write(buf, binary.LittleEndian, uint8(1))
		} else {
			binary.Write(buf, binary.LittleEndian, uint8(0))
		}
		binary.Write(buf, binary.LittleEndian, uint8(packet.NumTokens))
	}
	if packet.RouteType == routing.DecisionTypeNew {
		binary.Write(buf, binary.LittleEndian, packet.Tokens)
	}
	if packet.RouteType == routing.DecisionTypeContinue {
		binary.Write(buf, binary.LittleEndian, packet.Tokens)
	}
	binary.Write(buf, binary.LittleEndian, packet.ServerRoutePublicKey)

	return buf.Bytes()
}
