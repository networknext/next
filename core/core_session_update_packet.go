package core

import (
	"bytes"
	"encoding/binary"
	"net"
)

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
	TryBeforeYouBuy           bool
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
	if err := packet.Serialize(CreateReadStream(data), SDKVersionMajorMin, SDKVersionMinorMin, SDKVersionPatchMin); err != nil {
		return err
	}
	return nil
}

func (packet *SessionUpdatePacket) MarshalBinary() ([]byte, error) {
	ws, err := CreateWriteStream(1500)
	if err != nil {
		return nil, err
	}

	if err := packet.Serialize(ws, SDKVersionMajorMin, SDKVersionMinorMin, SDKVersionPatchMin); err != nil {
		return nil, err
	}
	ws.Flush()

	return ws.GetData(), nil
}

func (packet *SessionUpdatePacket) Serialize(stream Stream, versionMajor int32, versionMinor int32, versionPatch int32) error {
	stream.SerializeUint64(&packet.Sequence)
	stream.SerializeUint64(&packet.CustomerId)
	stream.SerializeAddress(&packet.ServerAddress)
	stream.SerializeUint64(&packet.SessionId)
	stream.SerializeUint64(&packet.UserHash)
	stream.SerializeUint64(&packet.PlatformId)
	stream.SerializeUint64(&packet.Tag)
	if ProtocolVersionAtLeast(versionMajor, versionMinor, versionPatch, 3, 3, 4) {
		stream.SerializeBits(&packet.Flags, FlagTotalCount)
	}
	stream.SerializeBool(&packet.Flagged)
	stream.SerializeBool(&packet.FallbackToDirect)
	stream.SerializeBool(&packet.TryBeforeYouBuy)
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
	stream.SerializeInteger(&packet.NumNearRelays, 0, NEXT_MAX_NEAR_RELAYS)
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
		packet.ClientRoutePublicKey = make([]byte, Crypto_box_PUBLICKEYBYTES)
		packet.Signature = make([]byte, SignatureBytes)
	}
	stream.SerializeBytes(packet.ClientRoutePublicKey)
	stream.SerializeUint32(&packet.KbpsUp)
	stream.SerializeUint32(&packet.KbpsDown)
	if ProtocolVersionAtLeast(versionMajor, versionMinor, versionPatch, 3, 3, 2) {
		stream.SerializeUint64(&packet.PacketsLostClientToServer)
		stream.SerializeUint64(&packet.PacketsLostServerToClient)
	}
	stream.SerializeBytes(packet.Signature)
	return stream.Error()
}

func (packet *SessionUpdatePacket) HeaderSerialize(stream Stream) error {
	stream.SerializeUint64(&packet.Sequence)
	stream.SerializeUint64(&packet.CustomerId)
	stream.SerializeAddress(&packet.ServerAddress)
	return stream.Error()
}

func (packet *SessionUpdatePacket) GetSignData(versionMajor int32, versionMinor int32, versionPatch int32) []byte {

	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, packet.Sequence)
	binary.Write(buf, binary.LittleEndian, packet.CustomerId)
	binary.Write(buf, binary.LittleEndian, packet.SessionId)
	binary.Write(buf, binary.LittleEndian, packet.UserHash)
	binary.Write(buf, binary.LittleEndian, packet.PlatformId)
	binary.Write(buf, binary.LittleEndian, packet.Tag)

	if ProtocolVersionAtLeast(versionMajor, versionMinor, versionPatch, 3, 3, 4) {
		binary.Write(buf, binary.LittleEndian, packet.Flags)
	}
	binary.Write(buf, binary.LittleEndian, packet.Flagged)
	binary.Write(buf, binary.LittleEndian, packet.FallbackToDirect)
	binary.Write(buf, binary.LittleEndian, packet.TryBeforeYouBuy)
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

	clientAddress := make([]byte, AddressBytes)
	WriteAddress(clientAddress, &packet.ClientAddress)
	binary.Write(buf, binary.LittleEndian, clientAddress)

	serverAddress := make([]byte, AddressBytes)
	WriteAddress(serverAddress, &packet.ServerAddress)
	binary.Write(buf, binary.LittleEndian, serverAddress)

	binary.Write(buf, binary.LittleEndian, packet.KbpsUp)
	binary.Write(buf, binary.LittleEndian, packet.KbpsDown)

	if ProtocolVersionAtLeast(versionMajor, versionMinor, versionPatch, 3, 3, 2) {
		binary.Write(buf, binary.LittleEndian, packet.PacketsLostClientToServer)
		binary.Write(buf, binary.LittleEndian, packet.PacketsLostServerToClient)
	}

	binary.Write(buf, binary.LittleEndian, packet.ClientRoutePublicKey)

	return buf.Bytes()
}
