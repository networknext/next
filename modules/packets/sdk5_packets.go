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

// ...

// ------------------------------------------------------------