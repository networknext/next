package packets

import (
	"github.com/networknext/backend/modules/common"
)

// ------------------------------------------------------------

type SDK5_ServerInitRequestPacket struct {
	Version        SDKVersion
	BuyerId        uint64
	DatacenterId   uint64
	RequestId      uint64
	DatacenterName string
}

func (packet *SDK5_ServerInitRequestPacket) Serialize(stream common.Stream) error {
	packet.Version.Serialize(stream)
	stream.SerializeUint64(&packet.BuyerId)
	stream.SerializeUint64(&packet.DatacenterId)
	stream.SerializeUint64(&packet.RequestId)
	stream.SerializeString(&packet.DatacenterName, SDK5_MaxDatacenterNameLength)
	return stream.Error()
}

// ------------------------------------------------------------

type SDK5_ServerInitResponsePacket struct {
	// ...
}

func (packet *SDK5_ServerInitResponsePacket) Serialize(stream common.Stream) error {
	// ...
	return stream.Error()
}

// ------------------------------------------------------------

type SDK5_ServerUpdateRequestPacket struct {
	Version        SDKVersion
	BuyerId        uint64
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
	Version        SDKVersion
	BuyerId        uint64
	// ...
}

func (packet *SDK5_ServerUpdateResponsePacket) Serialize(stream common.Stream) error {
	// ...
	return stream.Error()
}

// ------------------------------------------------------------

type SDK5_SessionUpdateRequestPacket struct {
	Version        SDKVersion
	BuyerId        uint64
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
	Version        SDKVersion
	BuyerId        uint64
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
