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

	PacketTypeServerUpdate       = 200
	PacketTypeSessionUpdate      = 201
	PacketTypeSessionResponse    = 202
	PacketTypeServerInitRequest  = 203
	PacketTypeServerInitResponse = 204

	InitResponseOK                   = 0
	InitResponseUnknownCustomer      = 1
	InitResponseUnknownDatacenter    = 2
	InitResponseOldSDKVersion        = 3
	InitResponseSignatureCheckFailed = 4

	MaxNearRelays = 32
	MaxTokens     = 7
)

type ServerInitRequestPacket struct {
	RequestID    uint64
	CustomerID   uint64
	DatacenterID uint64
	Signature    []byte

	Version SDKVersion
}

func (packet *ServerInitRequestPacket) Serialize(stream encoding.Stream) error {
	packetType := uint32(PacketTypeServerInitRequest)
	stream.SerializeBits(&packetType, 8)
	stream.SerializeInteger(&packet.Version.Major, 0, SDKVersionMax.Major)
	stream.SerializeInteger(&packet.Version.Minor, 0, SDKVersionMax.Minor)
	stream.SerializeInteger(&packet.Version.Patch, 0, SDKVersionMax.Patch)
	stream.SerializeUint64(&packet.RequestID)
	stream.SerializeUint64(&packet.CustomerID)
	stream.SerializeUint64(&packet.DatacenterID)
	if stream.IsReading() {
		packet.Signature = make([]byte, ed25519.SignatureSize)
	}
	stream.SerializeBytes(packet.Signature)
	return stream.Error()
}

func (packet *ServerInitRequestPacket) GetSignData() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, uint64(packet.Version.Major))
	binary.Write(buf, binary.LittleEndian, uint64(packet.Version.Minor))
	binary.Write(buf, binary.LittleEndian, uint64(packet.Version.Patch))
	binary.Write(buf, binary.LittleEndian, packet.RequestID)
	binary.Write(buf, binary.LittleEndian, packet.CustomerID)
	binary.Write(buf, binary.LittleEndian, packet.DatacenterID)
	return buf.Bytes()
}

func (packet *ServerInitRequestPacket) UnmarshalBinary(data []byte) error {
	if err := packet.Serialize(encoding.CreateReadStream(data)); err != nil {
		return err
	}
	return nil
}

func (packet *ServerInitRequestPacket) MarshalBinary() ([]byte, error) {
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

type ServerInitResponsePacket struct {
	RequestID uint64
	Response  uint32
	Signature []byte
}

func (packet *ServerInitResponsePacket) Serialize(stream encoding.Stream) error {
	packetType := uint32(PacketTypeServerInitResponse)
	stream.SerializeBits(&packetType, 8)
	stream.SerializeUint64(&packet.RequestID)
	stream.SerializeUint32(&packet.Response)
	if stream.IsReading() {
		packet.Signature = make([]byte, ed25519.SignatureSize)
	}
	stream.SerializeBytes(packet.Signature)
	return stream.Error()
}

func (packet *ServerInitResponsePacket) GetSignData() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, packet.RequestID)
	binary.Write(buf, binary.LittleEndian, packet.Response)
	return buf.Bytes()
}

func (packet *ServerInitResponsePacket) UnmarshalBinary(data []byte) error {
	if err := packet.Serialize(encoding.CreateReadStream(data)); err != nil {
		return err
	}
	return nil
}

func (packet *ServerInitResponsePacket) MarshalBinary() ([]byte, error) {
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

type ServerUpdatePacket struct {
	Sequence             uint64
	CustomerID           uint64
	DatacenterID         uint64
	NumSessionsPending   uint32
	NumSessionsUpgraded  uint32
	ServerAddress        net.UDPAddr
	ServerPrivateAddress net.UDPAddr // no longer used in 3.4.* SDK. please remove field when convenient
	ServerRoutePublicKey []byte
	Signature            []byte

	Version SDKVersion
}

func (packet *ServerUpdatePacket) Serialize(stream encoding.Stream) error {
	packetType := uint32(PacketTypeServerUpdate)
	stream.SerializeBits(&packetType, 8)

	stream.SerializeUint64(&packet.Sequence)
	stream.SerializeInteger(&packet.Version.Major, 0, SDKVersionMax.Major)
	stream.SerializeInteger(&packet.Version.Minor, 0, SDKVersionMax.Minor)
	stream.SerializeInteger(&packet.Version.Patch, 0, SDKVersionMax.Patch)
	stream.SerializeUint64(&packet.CustomerID)
	stream.SerializeUint64(&packet.DatacenterID)
	stream.SerializeUint32(&packet.NumSessionsPending)
	stream.SerializeUint32(&packet.NumSessionsUpgraded)
	stream.SerializeAddress(&packet.ServerAddress)
	if !packet.Version.AtLeast(SDKVersion{3, 4, 4}) {
		stream.SerializeAddress(&packet.ServerPrivateAddress)
	}
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
	binary.Write(buf, binary.LittleEndian, uint64(packet.Version.Major))
	binary.Write(buf, binary.LittleEndian, uint64(packet.Version.Minor))
	binary.Write(buf, binary.LittleEndian, uint64(packet.Version.Patch))
	binary.Write(buf, binary.LittleEndian, packet.CustomerID)
	binary.Write(buf, binary.LittleEndian, packet.DatacenterID)
	binary.Write(buf, binary.LittleEndian, packet.NumSessionsPending)
	binary.Write(buf, binary.LittleEndian, packet.NumSessionsUpgraded)

	address := make([]byte, encoding.AddressSize)
	encoding.WriteAddress(address, &packet.ServerAddress)
	binary.Write(buf, binary.LittleEndian, address)

	if !packet.Version.AtLeast(SDKVersion{3, 4, 4}) {
		privateAddress := make([]byte, encoding.AddressSize)
		encoding.WriteAddress(privateAddress, &packet.ServerPrivateAddress)
		binary.Write(buf, binary.LittleEndian, privateAddress)
	}

	binary.Write(buf, binary.LittleEndian, packet.ServerRoutePublicKey)
	return buf.Bytes()
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

type SessionUpdatePacket struct {
	Sequence                  uint64
	CustomerID                uint64
	SessionID                 uint64
	UserHash                  uint64
	PlatformID                uint64
	Tag                       uint64
	Flags                     uint32
	Flagged                   bool
	FallbackToDirect          bool
	TryBeforeYouBuy           bool
	ConnectionType            int32
	OnNetworkNext             bool
	Committed                 bool
	DirectMinRTT              float32
	DirectMaxRTT              float32
	DirectMeanRTT             float32
	DirectJitter              float32
	DirectPacketLoss          float32
	NextMinRTT                float32
	NextMaxRTT                float32
	NextMeanRTT               float32
	NextJitter                float32
	NextPacketLoss            float32
	NumNearRelays             int32
	NearRelayIDs              []uint64
	NearRelayMinRTT           []float32
	NearRelayMaxRTT           []float32
	NearRelayMeanRTT          []float32
	NearRelayJitter           []float32
	NearRelayPacketLoss       []float32
	ClientAddress             net.UDPAddr
	ServerAddress             net.UDPAddr
	ClientRoutePublicKey      []byte
	KbpsUp                    uint32
	KbpsDown                  uint32
	PacketsLostClientToServer uint64
	PacketsLostServerToClient uint64
	UserFlags                 uint64
	Signature                 []byte

	Version SDKVersion
}

func (packet *SessionUpdatePacket) Serialize(stream encoding.Stream) error {
	packetType := uint32(PacketTypeSessionUpdate)
	stream.SerializeBits(&packetType, 8)

	stream.SerializeUint64(&packet.Sequence)
	stream.SerializeUint64(&packet.CustomerID)
	stream.SerializeAddress(&packet.ServerAddress)
	stream.SerializeUint64(&packet.SessionID)
	stream.SerializeUint64(&packet.UserHash)
	stream.SerializeUint64(&packet.PlatformID)
	stream.SerializeUint64(&packet.Tag)

	if packet.Version.AtLeast(SDKVersion{3, 3, 4}) {
		if packet.Version.AtLeast(SDKVersion{3, 4, 0}) {
			stream.SerializeBits(&packet.Flags, 11)
		} else {
			stream.SerializeBits(&packet.Flags, 10)
		}
	}

	stream.SerializeBool(&packet.Flagged)
	stream.SerializeBool(&packet.FallbackToDirect)

	if !packet.Version.AtLeast(SDKVersion{3, 4, 0}) {
		stream.SerializeBool(&packet.TryBeforeYouBuy)
	}

	stream.SerializeInteger(&packet.ConnectionType, routing.ConnectionTypeUnknown, routing.ConnectionTypeCellular)
	stream.SerializeFloat32(&packet.DirectMinRTT)
	stream.SerializeFloat32(&packet.DirectMaxRTT)
	stream.SerializeFloat32(&packet.DirectMeanRTT)
	stream.SerializeFloat32(&packet.DirectJitter)
	stream.SerializeFloat32(&packet.DirectPacketLoss)
	stream.SerializeBool(&packet.OnNetworkNext)
	if packet.Version.AtLeast(SDKVersion{3, 4, 0}) {
		stream.SerializeBool(&packet.Committed)
	}
	if packet.OnNetworkNext {
		stream.SerializeFloat32(&packet.NextMinRTT)
		stream.SerializeFloat32(&packet.NextMaxRTT)
		stream.SerializeFloat32(&packet.NextMeanRTT)
		stream.SerializeFloat32(&packet.NextJitter)
		stream.SerializeFloat32(&packet.NextPacketLoss)
	}
	stream.SerializeInteger(&packet.NumNearRelays, 0, MaxNearRelays)
	if stream.IsReading() {
		packet.NearRelayIDs = make([]uint64, packet.NumNearRelays)
		packet.NearRelayMinRTT = make([]float32, packet.NumNearRelays)
		packet.NearRelayMaxRTT = make([]float32, packet.NumNearRelays)
		packet.NearRelayMeanRTT = make([]float32, packet.NumNearRelays)
		packet.NearRelayJitter = make([]float32, packet.NumNearRelays)
		packet.NearRelayPacketLoss = make([]float32, packet.NumNearRelays)
	}
	var i int32
	for i = 0; i < packet.NumNearRelays; i++ {
		stream.SerializeUint64(&packet.NearRelayIDs[i])
		stream.SerializeFloat32(&packet.NearRelayMinRTT[i])
		stream.SerializeFloat32(&packet.NearRelayMaxRTT[i])
		stream.SerializeFloat32(&packet.NearRelayMeanRTT[i])
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
	if packet.Version.AtLeast(SDKVersion{3, 3, 2}) {
		stream.SerializeUint64(&packet.PacketsLostClientToServer)
		stream.SerializeUint64(&packet.PacketsLostServerToClient)
	}

	if packet.Version.AtLeast(SDKVersion{3, 4, 0}) {
		stream.SerializeUint64(&packet.UserFlags)
	}

	stream.SerializeBytes(packet.Signature)
	return stream.Error()
}

func (packet *SessionUpdatePacket) GetSignData() []byte {

	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, packet.Sequence)
	binary.Write(buf, binary.LittleEndian, packet.CustomerID)
	binary.Write(buf, binary.LittleEndian, packet.SessionID)
	binary.Write(buf, binary.LittleEndian, packet.UserHash)
	binary.Write(buf, binary.LittleEndian, packet.PlatformID)
	binary.Write(buf, binary.LittleEndian, packet.Tag)

	if packet.Version.AtLeast(SDKVersion{3, 3, 4}) {
		binary.Write(buf, binary.LittleEndian, packet.Flags)
	}

	binary.Write(buf, binary.LittleEndian, packet.Flagged)
	binary.Write(buf, binary.LittleEndian, packet.FallbackToDirect)
	if !packet.Version.AtLeast(SDKVersion{3, 4, 0}) {
		binary.Write(buf, binary.LittleEndian, packet.TryBeforeYouBuy)
	}
	binary.Write(buf, binary.LittleEndian, uint8(packet.ConnectionType))

	var onNetworkNext uint8
	if packet.OnNetworkNext {
		onNetworkNext = 1
	}
	binary.Write(buf, binary.LittleEndian, onNetworkNext)

	if packet.Version.AtLeast(SDKVersion{3, 4, 0}) {
		var committed uint8
		if packet.Committed {
			committed = 1
		}
		binary.Write(buf, binary.LittleEndian, committed)
	}

	binary.Write(buf, binary.LittleEndian, packet.DirectMinRTT)
	binary.Write(buf, binary.LittleEndian, packet.DirectMaxRTT)
	binary.Write(buf, binary.LittleEndian, packet.DirectMeanRTT)
	binary.Write(buf, binary.LittleEndian, packet.DirectJitter)
	binary.Write(buf, binary.LittleEndian, packet.DirectPacketLoss)

	binary.Write(buf, binary.LittleEndian, packet.NextMinRTT)
	binary.Write(buf, binary.LittleEndian, packet.NextMaxRTT)
	binary.Write(buf, binary.LittleEndian, packet.NextMeanRTT)
	binary.Write(buf, binary.LittleEndian, packet.NextJitter)
	binary.Write(buf, binary.LittleEndian, packet.NextPacketLoss)

	binary.Write(buf, binary.LittleEndian, uint32(packet.NumNearRelays))
	var i int32
	for i = 0; i < packet.NumNearRelays; i++ {
		binary.Write(buf, binary.LittleEndian, packet.NearRelayIDs[i])
		binary.Write(buf, binary.LittleEndian, packet.NearRelayMinRTT[i])
		binary.Write(buf, binary.LittleEndian, packet.NearRelayMaxRTT[i])
		binary.Write(buf, binary.LittleEndian, packet.NearRelayMeanRTT[i])
		binary.Write(buf, binary.LittleEndian, packet.NearRelayJitter[i])
		binary.Write(buf, binary.LittleEndian, packet.NearRelayPacketLoss[i])
	}

	clientAddress := make([]byte, encoding.AddressSize)
	encoding.WriteAddress(clientAddress, &packet.ClientAddress)
	binary.Write(buf, binary.LittleEndian, clientAddress)

	serverAddress := make([]byte, encoding.AddressSize)
	encoding.WriteAddress(serverAddress, &packet.ServerAddress)
	binary.Write(buf, binary.LittleEndian, serverAddress)

	binary.Write(buf, binary.LittleEndian, packet.KbpsUp)
	binary.Write(buf, binary.LittleEndian, packet.KbpsDown)

	if packet.Version.AtLeast(SDKVersion{3, 3, 4}) {
		binary.Write(buf, binary.LittleEndian, packet.PacketsLostClientToServer)
		binary.Write(buf, binary.LittleEndian, packet.PacketsLostServerToClient)
	}

	if packet.Version.AtLeast(SDKVersion{3, 4, 0}) {
		binary.Write(buf, binary.LittleEndian, packet.UserFlags)
	}

	binary.Write(buf, binary.LittleEndian, packet.ClientRoutePublicKey)

	return buf.Bytes()
}

func (packet *SessionUpdatePacket) UnmarshalBinary(data []byte) error {
	if err := packet.Serialize(encoding.CreateReadStream(data)); err != nil {
		return err
	}
	return nil
}

func (packet *SessionUpdatePacket) MarshalBinary() ([]byte, error) {
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

type SessionResponsePacket struct {
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
	Signature            []byte

	Version SDKVersion
}

func (packet *SessionResponsePacket) Serialize(stream encoding.Stream) error {
	packetType := uint32(PacketTypeSessionResponse)
	stream.SerializeBits(&packetType, 8)

	stream.SerializeUint64(&packet.Sequence)
	stream.SerializeUint64(&packet.SessionID)
	stream.SerializeInteger(&packet.NumNearRelays, 0, MaxNearRelays)
	if stream.IsReading() {
		packet.NearRelayIDs = make([]uint64, packet.NumNearRelays)
		packet.NearRelayAddresses = make([]net.UDPAddr, packet.NumNearRelays)
	}
	var i int32
	for i = 0; i < packet.NumNearRelays; i++ {
		stream.SerializeUint64(&packet.NearRelayIDs[i])
		stream.SerializeAddress(&packet.NearRelayAddresses[i])
	}
	stream.SerializeInteger(&packet.RouteType, 0, routing.RouteTypeContinue)
	if packet.RouteType != routing.RouteTypeDirect {
		stream.SerializeBool(&packet.Multipath)
		if packet.Version.AtLeast(SDKVersion{3, 4, 0}) {
			stream.SerializeBool(&packet.Committed)
		}
		stream.SerializeInteger(&packet.NumTokens, 0, MaxTokens)
	}
	if packet.RouteType == routing.RouteTypeNew {
		if stream.IsReading() {
			packet.Tokens = make([]byte, packet.NumTokens*routing.EncryptedNextRouteTokenSize)
		}
		stream.SerializeBytes(packet.Tokens)
	}
	if packet.RouteType == routing.RouteTypeContinue {
		if stream.IsReading() {
			packet.Tokens = make([]byte, packet.NumTokens*routing.EncryptedContinueRouteTokenSize)
		}
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
	binary.Write(buf, binary.LittleEndian, packet.SessionID)
	binary.Write(buf, binary.LittleEndian, uint8(packet.NumNearRelays))
	var i int32
	for i = 0; i < packet.NumNearRelays; i++ {
		binary.Write(buf, binary.LittleEndian, packet.NearRelayIDs[i])
		address := make([]byte, encoding.AddressSize)
		encoding.WriteAddress(address, &packet.NearRelayAddresses[i])
		binary.Write(buf, binary.LittleEndian, address)
	}
	binary.Write(buf, binary.LittleEndian, uint8(packet.RouteType))
	if packet.RouteType != routing.RouteTypeDirect {
		if packet.Multipath {
			binary.Write(buf, binary.LittleEndian, uint8(1))
		} else {
			binary.Write(buf, binary.LittleEndian, uint8(0))
		}

		if packet.Version.AtLeast(SDKVersion{3, 4, 0}) {
			if packet.Committed {
				binary.Write(buf, binary.LittleEndian, uint8(1))
			} else {
				binary.Write(buf, binary.LittleEndian, uint8(0))
			}
		}

		binary.Write(buf, binary.LittleEndian, uint8(packet.NumTokens))
	}
	if packet.RouteType == routing.RouteTypeNew {
		binary.Write(buf, binary.LittleEndian, packet.Tokens)
	}
	if packet.RouteType == routing.RouteTypeContinue {
		binary.Write(buf, binary.LittleEndian, packet.Tokens)
	}
	binary.Write(buf, binary.LittleEndian, packet.ServerRoutePublicKey)

	return buf.Bytes()
}

func (packet *SessionResponsePacket) UnmarshalBinary(data []byte) error {
	if err := packet.Serialize(encoding.CreateReadStream(data)); err != nil {
		return err
	}
	return nil
}

func (packet *SessionResponsePacket) MarshalBinary() ([]byte, error) {
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
