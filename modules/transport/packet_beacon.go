package transport

import (
	"fmt"
	"time"

	"cloud.google.com/go/bigquery"

	"github.com/networknext/backend/modules/encoding"
)

const (
	PacketTypeBeacon = 118 // Magic number to indicate beacon packet

	BeaconPacketVersion = 0

	MaxNextBeaconPacketBytes = 4 + // Version
		8 + // Timestamp
		8 + // BuyerID
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

type NextBeaconPacket struct {
	Version          uint32
	Timestamp        uint64
	BuyerID          uint64
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

	stream.SerializeUint64(&packet.BuyerID)

	if hasDatacenterID {
		stream.SerializeUint64(&packet.DatacenterID)
	}

	if packet.Upgraded {
		stream.SerializeUint64(&packet.UserHash)
		stream.SerializeUint64(&packet.AddressHash)
		stream.SerializeUint64(&packet.SessionID)
	}

	stream.SerializeInteger(&packet.PlatformID, PlatformTypeUnknown, PlatformTypeMax)

	stream.SerializeInteger(&packet.ConnectionType, ConnectionTypeUnknown, ConnectionTypeMax)

	hasTimestamp := stream.IsReading() && packet.Timestamp != 0 || stream.IsWriting()
	stream.SerializeBool(&hasTimestamp)

	if hasTimestamp {
		stream.SerializeUint64(&packet.Timestamp)
	}

	return stream.Error()
}

func WriteBeaconEntry(entry *NextBeaconPacket) ([]byte, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("recovered from panic during beacon packet entry write: %v\n", r)
		}
	}()

	// Set the timestamp if needed so that we can serialize it properly
	if entry.Timestamp == 0 {
		entry.Timestamp = uint64(time.Now().Unix())
	}

	buffer := [MaxNextBeaconPacketBytes]byte{}

	ws, err := encoding.CreateWriteStream(buffer[:])
	if err != nil {
		return nil, err
	}

	if err := entry.Serialize(ws); err != nil {
		return nil, err
	}
	ws.Flush()

	return buffer[:ws.GetBytesProcessed()], nil
}

func ReadBeaconEntry(entry *NextBeaconPacket, data []byte) error {
	if err := entry.Serialize(encoding.CreateReadStream(data)); err != nil {
		return err
	}
	return nil
}

// Save implements the bigquery.ValueSaver interface for an Entry
// so it can be used in Put()
func (entry *NextBeaconPacket) Save() (map[string]bigquery.Value, string, error) {
	e := make(map[string]bigquery.Value)

	e["version"] = int(entry.Version)
	e["timestamp"] = int(entry.Timestamp)
	e["customerID"] = int(entry.BuyerID) // todo: should rename to buyer id at some point
	e["datacenterID"] = int(entry.DatacenterID)
	e["userHash"] = int(entry.UserHash)
	e["addressHash"] = int(entry.AddressHash)
	e["sessionID"] = int(entry.SessionID)
	e["platformID"] = int(entry.PlatformID)
	e["connectionType"] = int(entry.ConnectionType)

	e["enabled"] = entry.Enabled
	e["upgraded"] = entry.Upgraded
	e["next"] = entry.Next
	e["fallbackToDirect"] = entry.FallbackToDirect

	return e, "", nil
}
