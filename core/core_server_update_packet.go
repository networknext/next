package core

import (
	"bytes"
	"encoding/binary"
	"net"
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
	if err := packet.Serialize(CreateReadStream(data)); err != nil {
		return err
	}
	return nil
}

func (packet *ServerUpdatePacket) MarshalBinary() ([]byte, error) {
	ws, err := CreateWriteStream(1500)
	if err != nil {
		return nil, err
	}

	if err := packet.Serialize(ws); err != nil {
		return nil, err
	}
	ws.Flush()

	return ws.GetData(), nil
}

func (packet *ServerUpdatePacket) Serialize(stream Stream) error {
	packetType := uint32(200)
	stream.SerializeBits(&packetType, 8)

	stream.SerializeUint64(&packet.Sequence)
	stream.SerializeInteger(&packet.VersionMajor, 0, SDKVersionMajorMax)
	stream.SerializeInteger(&packet.VersionMinor, 0, SDKVersionMinorMax)
	stream.SerializeInteger(&packet.VersionPatch, 0, SDKVersionPatchMax)
	stream.SerializeUint64(&packet.CustomerId)
	stream.SerializeUint64(&packet.DatacenterId)
	stream.SerializeUint32(&packet.NumSessionsPending)
	stream.SerializeUint32(&packet.NumSessionsUpgraded)
	stream.SerializeAddress(&packet.ServerAddress)
	stream.SerializeAddress(&packet.ServerPrivateAddress)
	if stream.IsReading() {
		packet.ServerRoutePublicKey = make([]byte, Crypto_box_PUBLICKEYBYTES)
		packet.Signature = make([]byte, SignatureBytes)
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

	address := make([]byte, AddressBytes)
	WriteAddress(address, &packet.ServerAddress)
	binary.Write(buf, binary.LittleEndian, address)

	privateAddress := make([]byte, AddressBytes)
	WriteAddress(privateAddress, &packet.ServerPrivateAddress)
	binary.Write(buf, binary.LittleEndian, privateAddress)

	binary.Write(buf, binary.LittleEndian, packet.ServerRoutePublicKey)
	return buf.Bytes()
}
