package packets

import (
	// "github.com/networknext/backend/modules/encoding"	
)

type ServerInitRequestPacket struct {

	// todo
	// Version        SDKVersion

	BuyerId        uint64
	DatacenterId   uint64
	RequestId      uint64
	DatacenterName string
}

/*
func (packet *ServerInitRequestPacket) Serialize[type T common.Stream](stream T) error {

	// todo: version
	// versionMajor := uint32(packet.Version.Major)
	// versionMinor := uint32(packet.Version.Minor)
	// versionPatch := uint32(packet.Version.Patch)
	// stream.SerializeBits(&versionMajor, 8)
	// stream.SerializeBits(&versionMinor, 8)
	// stream.SerializeBits(&versionPatch, 8)
	// packet.Version = SDKVersion{int32(versionMajor), int32(versionMinor), int32(versionPatch)}
	
	
	stream.SerializeUint64(&packet.BuyerId)
	stream.SerializeUint64(&packet.DatacenterId)
	stream.SerializeUint64(&packet.RequestId)
	stream.SerializeString(&packet.DatacenterName, MaxDatacenterNameLength)
	
	return stream.Error()
}
*/