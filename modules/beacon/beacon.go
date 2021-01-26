/*
   Network Next. You control the network.
   Copyright © 2017 - 2021 Network Next, Inc. All rights reserved.
*/

package beacon

import (
	"context"
	"fmt"

	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/modules/transport"
)

const (
	NEXT_CONNECTION_TYPE_UNKNOWN  = 0
	NEXT_CONNECTION_TYPE_WIRED    = 1
	NEXT_CONNECTION_TYPE_WIFI     = 2
	NEXT_CONNECTION_TYPE_CELLULAR = 3
	NEXT_CONNECTION_TYPE_MAX      = 3

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

	MaxNextBeaconPacketBytes = 4 + // Version
		8 + // Timestamp
		8 + // CustomerID
		8 + // DatacenterID
		8 + // UserHash
		8 + // AddressHash
		8 + // SessionID
		4 + // PlatformID
		4 + // ConnectionType
		1 + // Enabled
		1 + // Upgraded
		1 + // Next
		1 // FallbackToDirect
)

// type RawBeaconPacket struct {
// 	Version          uint32
// 	CustomerID       uint64
// 	DatacenterID     uint64
// 	UserHash         uint64
// 	AddressHash      uint64
// 	SessionID        uint64
// 	PlatformID       int32
// 	ConnectionType   int32
// 	Enabled          bool
// 	Upgraded         bool
// 	Next             bool
// 	FallbackToDirect bool
// }

type NextBeaconPacket struct {
	Version          uint32
	Timestamp        uint64
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

// Beaconer is a beacon service interface that handles sending beacon packet entries through google pubsub to bigquery
type Beaconer interface {
	Submit(ctx context.Context, entry *NextBeaconPacket) error
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

func WriteBeaconEntry(entry *NextBeaconPacket) ([]byte, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("recovered from panic during beacon packet entry write: %v\n", r)
		}
	}()

	// data := make([]byte, MaxNextBeaconPacketBytes)
	// index := 0

	// encoding.WriteUint32(data, &index, entry.Version)

	// encoding.WriteUint64(data, &index, entry.Timestamp)
	// encoding.WriteUint64(data, &index, entry.CustomerID)
	// encoding.WriteUint64(data, &index, entry.DatacenterID)
	// encoding.WriteUint64(data, &index, entry.UserHash)
	// encoding.WriteUint64(data, &index, entry.AddressHash)
	// encoding.WriteUint64(data, &index, entry.SessionID)

	// encoding.WriteUint32(data, &index, entry.PlatformID)
	// encoding.WriteUint32(data, &index, entry.ConnectionType)

	// encoding.WriteBool(data, &index, entry.Enabled)
	// encoding.WriteBool(data, &index, entry.Upgraded)
	// encoding.WriteBool(data, &index, entry.Next)
	// encoding.WriteBool(data, &index, entry.FallbackToDirect)

	// return data

	ws, err := encoding.CreateWriteStream(transport.DefaultMaxPacketSize)
	if err != nil {
		return nil, err
	}

	if err := entry.Serialize(ws); err != nil {
		return nil, err
	}
	ws.Flush()

	return ws.GetData()[:ws.GetBytesProcessed()], nil
}

func ReadBeaconEntry(entry *NextBeaconPacket, data []byte) error {
	// index := 0
	// if !encoding.ReadUint32(data, &index, &entry.Version) {
	// 	return false
	// }

	// if !encoding.ReadUint64(data, &index, &entry.Timestamp) {
	// 	return false
	// }

	// if !encoding.ReadUint64(data, &index, &entry.CustomerID) {
	// 	return false
	// }

	// if !encoding.ReadUint64(data, &index, &entry.DatacenterID) {
	// 	return false
	// }

	// if !encoding.ReadUint64(data, &index, &entry.UserHash) {
	// 	return false
	// }

	// if !encoding.ReadUint64(data, &index, &entry.AddressHash) {
	// 	return false
	// }

	// if !encoding.ReadUint64(data, &index, &entry.SessionID) {
	// 	return false
	// }

	// if !encoding.ReadUint32(data, &index, &entry.PlatformID) {
	// 	return false
	// }

	// if !encoding.ReadUint32(data, &index, &entry.ConnectionType) {
	// 	return false
	// }

	// if !encoding.ReadBool(data, &index, &entry.Enabled) {
	// 	return false
	// }

	// if !encoding.ReadBool(data, &index, &entry.Upgraded) {
	// 	return false
	// }

	// if !encoding.ReadBool(data, &index, &entry.Next) {
	// 	return false
	// }

	// if !encoding.ReadBool(data, &index, &entry.FallbackToDirect) {
	// 	return false
	// }

	// return true

	if err := entry.Serialize(encoding.CreateReadStream(data)); err != nil {
		return err
	}
	return nil
}
