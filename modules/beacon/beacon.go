/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package beacon

import (
	"github.com/networknext/backend/modules/encoding"
)

const (
	NEXT_CONNECTION_TYPE_UNKNOWN  = 0
	NEXT_CONNECTION_TYPE_WIRED    = 1
	NEXT_CONNECTION_TYPE_WIFI     = 2
	NEXT_CONNECTION_TYPE_CELLULAR = 3
	NEXT_CONNECTION_TYPE_MAX      = 3
)

const (
	NEXT_PLATFORM_UNKNOWN       = 0
	NEXT_PLATFORM_WINDOWS       = 1
	NEXT_PLATFORM_MAC           = 2
	NEXT_PLATFORM_UNIX          = 3
	NEXT_PLATFORM_SWITCH        = 4
	NEXT_PLATFORM_PS4           = 5
	NEXT_PLATFORM_IOS           = 6
	NEXT_PLATFORM_XBOX_ONE      = 7
	NEXT_PLATFORM_XBOX_SERIES_X = 8
	NEXT_PLATFORM_PS5           = 9
	NEXT_PLATFORM_MAX           = 9
)

type NextBeaconPacket struct {
	Version          uint32
	CustomerID       uint64
	DatacenterID     uint64
	UserHash         uint64
	AddressHash      uint64
	SessionID        uint64
	PlatformID       int32
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

	hasDatacenterID := stream.IsWriting() && packet.DatacenterID != 0
	stream.SerializeBool(&hasDatacenterID)

	stream.SerializeUint64(&packet.CustomerID)

	if hasDatacenterID {
		stream.SerializeUint64(&packet.DatacenterID)
	}

	if packet.Upgraded {
		stream.SerializeUint64(&packet.UserHash)
		stream.SerializeUint64(&packet.AddressHash)
		stream.SerializeUint64(&packet.SessionID)
	}

	stream.SerializeInteger(&packet.PlatformID, NEXT_PLATFORM_UNKNOWN, NEXT_PLATFORM_MAX)

	stream.SerializeInteger(&packet.ConnectionType, NEXT_CONNECTION_TYPE_UNKNOWN, NEXT_CONNECTION_TYPE_MAX)

	return stream.Error()
}
