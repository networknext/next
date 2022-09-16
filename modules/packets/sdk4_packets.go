package packets

import (
	"github.com/networknext/backend/modules/common"
)

// ------------------------------------------------------------

type SDK4_ServerInitRequestPacket struct {
	Version        SDKVersion
	BuyerId        uint64
	DatacenterId   uint64
	RequestId      uint64
	DatacenterName string
}

func (packet *SDK4_ServerInitRequestPacket) Serialize(stream common.Stream) error {
	packet.Version.Serialize(stream)
	stream.SerializeUint64(&packet.BuyerId)
	stream.SerializeUint64(&packet.DatacenterId)
	stream.SerializeUint64(&packet.RequestId)
	stream.SerializeString(&packet.DatacenterName, SDK4_MaxDatacenterNameLength)
	return stream.Error()
}

// ------------------------------------------------------------

// ...

// ------------------------------------------------------------
