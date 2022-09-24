package packets

import (
	"github.com/networknext/backend/modules/common"
)

// ------------------------------------------------------------

type SDK5_ServerInitRequestPacket struct {
	Version        SDKVersion
	BuyerId        uint64
	RequestId      uint64
	DatacenterId   uint64
	DatacenterName string
}

func (packet *SDK5_ServerInitRequestPacket) Serialize(stream common.Stream) error {
	packet.Version.Serialize(stream)
	stream.SerializeUint64(&packet.BuyerId)
	stream.SerializeUint64(&packet.RequestId)
	stream.SerializeUint64(&packet.DatacenterId)
	stream.SerializeString(&packet.DatacenterName, SDK5_MaxDatacenterNameLength)
	return stream.Error()
}

// ------------------------------------------------------------

type SDK5_ServerInitResponsePacket struct {
	RequestId     uint64
	Response      uint32
	UpcomingMagic [8]byte
	CurrentMagic  [8]byte
	PreviousMagic [8]byte
}

func (packet *SDK5_ServerInitResponsePacket) Serialize(stream common.Stream) error {
	stream.SerializeUint64(&packet.RequestId)
	stream.SerializeBits(&packet.Response, 8)
	stream.SerializeBytes(packet.UpcomingMagic[:])
	stream.SerializeBytes(packet.CurrentMagic[:])
	stream.SerializeBytes(packet.PreviousMagic[:])
	return stream.Error()
}

// ------------------------------------------------------------

type SDK5_ServerUpdateRequestPacket struct {
	Version SDKVersion
	BuyerId uint64
	// ...
}

func (packet *SDK5_ServerUpdateRequestPacket) Serialize(stream common.Stream) error {
	packet.Version.Serialize(stream)
	stream.SerializeUint64(&packet.BuyerId)
	// ...
	return stream.Error()
}

// ------------------------------------------------------------

type SDK5_ServerUpdateResponsePacket struct {
	Version SDKVersion
	BuyerId uint64
	// ...
}

func (packet *SDK5_ServerUpdateResponsePacket) Serialize(stream common.Stream) error {
	// ...
	return stream.Error()
}

// ------------------------------------------------------------

type SDK5_SessionUpdateRequestPacket struct {
	Version SDKVersion
	BuyerId uint64
	// ...
}

func (packet *SDK5_SessionUpdateRequestPacket) Serialize(stream common.Stream) error {
	packet.Version.Serialize(stream)
	stream.SerializeUint64(&packet.BuyerId)
	// ...
	return stream.Error()
}

// ------------------------------------------------------------

type SDK5_SessionUpdateResponsePacket struct {
	// ...
}

func (packet *SDK5_SessionUpdateResponsePacket) Serialize(stream common.Stream) error {
	// ...
	return stream.Error()
}

// ------------------------------------------------------------

type SDK5_MatchDataRequestPacket struct {
	Version SDKVersion
	BuyerId uint64
	// ...
}

func (packet *SDK5_MatchDataRequestPacket) Serialize(stream common.Stream) error {
	packet.Version.Serialize(stream)
	stream.SerializeUint64(&packet.BuyerId)
	// ...
	return stream.Error()
}

// ------------------------------------------------------------

type SDK5_MatchDataResponsePacket struct {
}

func (packet *SDK5_MatchDataResponsePacket) Serialize(stream common.Stream) error {
	// ...
	return stream.Error()
}

// ------------------------------------------------------------
