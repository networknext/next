package messages

import (
	"cloud.google.com/go/bigquery"
	"github.com/networknext/backend/modules/encoding"
)

const (
	PingStatsEntryVersion  = uint8(4)
	MaxPingStatsEntrySize  = 128
	MaxInstanceIDLength    = 64        // todo: remove
)

type PingStatsEntry struct {
	Timestamp  uint64
	RelayA     uint64
	RelayB     uint64
	RTT        float32
	Jitter     float32
	PacketLoss float32
	Routable   bool
}

func WritePingStatsEntries(entries []PingStatsEntry) []byte {
	
	length := 1 + 8 + len(entries)*int(MaxRelayStatsEntrySize)
	
	data := make([]byte, length)

	index := 0
	encoding.WriteUint8(data, &index, PingStatsEntryVersion)
	encoding.WriteUint64(data, &index, uint64(len(entries)))

	for i := range entries {
		entry := &entries[i]
		encoding.WriteUint64(data, &index, entry.RelayA)
		encoding.WriteUint64(data, &index, entry.RelayB)
		encoding.WriteFloat32(data, &index, entry.RTT)
		encoding.WriteFloat32(data, &index, entry.Jitter)
		encoding.WriteFloat32(data, &index, entry.PacketLoss)
		encoding.WriteBool(data, &index, entry.Routable)
	}

	return data[:index]
}

func ReadPingStatsEntries(data []byte) ([]*PingStatsEntry, bool) {
	
	index := 0

	var version uint8
	if !encoding.ReadUint8(data, &index, &version) {
		return nil, false
	}

	var length uint64
	if !encoding.ReadUint64(data, &index, &length) {
		return nil, false
	}

	entries := make([]*PingStatsEntry, length)

	for i := range entries {
		entry := new(PingStatsEntry)

		if !encoding.ReadUint64(data, &index, &entry.RelayA) {
			return nil, false
		}

		if !encoding.ReadUint64(data, &index, &entry.RelayB) {
			return nil, false
		}

		if !encoding.ReadFloat32(data, &index, &entry.RTT) {
			return nil, false
		}

		if !encoding.ReadFloat32(data, &index, &entry.Jitter) {
			return nil, false
		}

		if !encoding.ReadFloat32(data, &index, &entry.PacketLoss) {
			return nil, false
		}

		if version >= 2 {
			if !encoding.ReadBool(data, &index, &entry.Routable) {
				return nil, false
			}
		}

		if version == 3 {
			// we don't support these anymore
			var InstanceID string
			var Debug bool
			if !encoding.ReadString(data, &index, &InstanceID, uint32(MaxInstanceIDLength)) {
				return nil, false
			}
			if !encoding.ReadBool(data, &index, &Debug) {
				return nil, false
			}
		}

		entries[i] = entry
	}

	return entries, true
}

func (e *PingStatsEntry) Save() (map[string]bigquery.Value, string, error) {

	bqEntry := make(map[string]bigquery.Value)

	bqEntry["timestamp"] = int(e.Timestamp)
	bqEntry["relay_a"] = int(e.RelayA)
	bqEntry["relay_b"] = int(e.RelayB)
	bqEntry["rtt"] = e.RTT
	bqEntry["jitter"] = e.Jitter
	bqEntry["packet_loss"] = e.PacketLoss
	bqEntry["routable"] = e.Routable

	return bqEntry, "", nil
}
