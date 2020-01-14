package core

import (
	"bytes"
	"encoding/binary"
	"net"
)

type SessionResponsePacket struct {
	Sequence             uint64
	SessionId            uint64
	NumNearRelays        int32
	NearRelayIds         []uint64
	NearRelayAddresses   []net.UDPAddr
	ResponseType         int32
	Multipath            bool
	NumTokens            int32
	Tokens               []byte
	ServerRoutePublicKey []byte
	signature            []byte
}

func (packet *SessionResponsePacket) UnmarshalBinary(data []byte) error {
	if err := packet.Serialize(CreateReadStream(data), SDKVersionMajorMin, SDKVersionMinorMin, SDKVersionPatchMin); err != nil {
		return err
	}
	return nil
}

func (packet *SessionResponsePacket) MarshalBinary() ([]byte, error) {
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

func (packet *SessionResponsePacket) Serialize(stream Stream, versionMajor int32, versionMinor int32, versionPatch int32) error {
	packetType := uint32(202)
	stream.SerializeBits(&packetType, 8)

	stream.SerializeUint64(&packet.Sequence)
	stream.SerializeUint64(&packet.SessionId)
	stream.SerializeInteger(&packet.NumNearRelays, 0, NEXT_MAX_NEAR_RELAYS)
	if stream.IsReading() {
		packet.NearRelayIds = make([]uint64, packet.NumNearRelays)
		packet.NearRelayAddresses = make([]net.UDPAddr, packet.NumNearRelays)
	}
	var i int32
	for i = 0; i < packet.NumNearRelays; i++ {
		stream.SerializeUint64(&packet.NearRelayIds[i])
		stream.SerializeAddress(&packet.NearRelayAddresses[i])
	}
	stream.SerializeInteger(&packet.ResponseType, 0, NEXT_UPDATE_TYPE_CONTINUE)
	if packet.ResponseType != NEXT_UPDATE_TYPE_DIRECT {
		stream.SerializeBool(&packet.Multipath)
		stream.SerializeInteger(&packet.NumTokens, 0, NEXT_MAX_TOKENS)
	}
	if packet.ResponseType == NEXT_UPDATE_TYPE_ROUTE {
		stream.SerializeBytes(packet.Tokens)
	}
	if packet.ResponseType == NEXT_UPDATE_TYPE_CONTINUE {
		stream.SerializeBytes(packet.Tokens)
	}
	if stream.IsReading() {
		packet.ServerRoutePublicKey = make([]byte, Crypto_box_PUBLICKEYBYTES)
		packet.signature = make([]byte, SignatureBytes)
	}
	stream.SerializeBytes(packet.ServerRoutePublicKey)
	stream.SerializeBytes(packet.signature)
	return stream.Error()
}

func (packet *SessionResponsePacket) Sign(versionMajor int32, versionMinor int32, versionPatch int32) {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, packet.Sequence)
	binary.Write(buf, binary.LittleEndian, packet.SessionId)
	binary.Write(buf, binary.LittleEndian, uint8(packet.NumNearRelays))
	var i int32
	for i = 0; i < packet.NumNearRelays; i++ {
		binary.Write(buf, binary.LittleEndian, packet.NearRelayIds[i])
		address := make([]byte, AddressBytes)
		WriteAddress(address, &packet.NearRelayAddresses[i])
		binary.Write(buf, binary.LittleEndian, address)
	}
	binary.Write(buf, binary.LittleEndian, uint8(packet.ResponseType))
	if packet.ResponseType != NEXT_UPDATE_TYPE_DIRECT {
		if packet.Multipath {
			binary.Write(buf, binary.LittleEndian, uint8(1))
		} else {
			binary.Write(buf, binary.LittleEndian, uint8(0))
		}
		binary.Write(buf, binary.LittleEndian, uint8(packet.NumTokens))
	}
	if packet.ResponseType == NEXT_UPDATE_TYPE_ROUTE {
		binary.Write(buf, binary.LittleEndian, packet.Tokens)
	}
	if packet.ResponseType == NEXT_UPDATE_TYPE_CONTINUE {
		binary.Write(buf, binary.LittleEndian, packet.Tokens)
	}
	binary.Write(buf, binary.LittleEndian, packet.ServerRoutePublicKey)

	packet.signature = CryptoSignCreate(buf.Bytes(), BackendPrivateKey)
}
