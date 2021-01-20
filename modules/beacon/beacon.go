/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package beacon

import (
	"github.com/networknext/backend/modules/encoding"
)

type NextBeaconPacket struct {
	Version          uint32
	CustomerId       uint64
	DatacenterId     uint64
	UserHash         uint64
	AddressHash      uint64
	SessionId        uint64
	PlatformId       int32
	ConnectionType   int32
	Enabled          bool
	Upgraded         bool
	Next             bool
	FallbackToDirect bool
}

func (packet *NextBeaconPacket) Serialize(stream encoding.Stream) error {

	stream.SerializeBits(&packet.Version, 8)

	stream.SerializeBool(&packet.Enabled)
	stream.SerializeBool(&packet.Upgraded)
	stream.SerializeBool(&packet.Next)
	stream.SerializeBool(&packet.FallbackToDirect)

	hasDatacenterId := stream.IsWriting() && packet.DatacenterId != 0
	stream.SerializeBool(&hasDatacenterId)

	stream.SerializeUint64(&packet.CustomerId)

	if hasDatacenterId {
		stream.SerializeUint64(&packet.DatacenterId)
	}

	if packet.Upgraded {
		stream.SerializeUint64(&packet.UserHash)
		stream.SerializeUint64(&packet.AddressHash)
		stream.SerializeUint64(&packet.SessionId)
	}

	stream.SerializeInteger(&packet.PlatformId, NEXT_PLATFORM_UNKNOWN, NEXT_PLATFORM_MAX)

	stream.SerializeInteger(&packet.ConnectionType, NEXT_CONNECTION_TYPE_UNKNOWN, NEXT_CONNECTION_TYPE_MAX)

	return stream.Error()
}
