package packets

type ServerInitRequestPacket struct {
	Version        SDKVersion
	BuyerID        uint64
	DatacenterID   uint64
	RequestID      uint64
	DatacenterName string
}

func (packet *ServerInitRequestPacket) Serialize(stream encoding.Stream) error {
	versionMajor := uint32(packet.Version.Major)
	versionMinor := uint32(packet.Version.Minor)
	versionPatch := uint32(packet.Version.Patch)
	stream.SerializeBits(&versionMajor, 8)
	stream.SerializeBits(&versionMinor, 8)
	stream.SerializeBits(&versionPatch, 8)
	packet.Version = SDKVersion{int32(versionMajor), int32(versionMinor), int32(versionPatch)}
	stream.SerializeUint64(&packet.BuyerID)
	stream.SerializeUint64(&packet.DatacenterID)
	stream.SerializeUint64(&packet.RequestID)
	stream.SerializeString(&packet.DatacenterName, MaxDatacenterNameLength)
	return stream.Error()
}
