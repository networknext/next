package analytics

import (
	"cloud.google.com/go/bigquery"
	"github.com/networknext/backend/encoding"
)

const (
	StatsEntryVersion = uint8(1)
)

type StatsEntry struct {
	Timestamp uint64

	RelayA     uint64
	RelayB     uint64
	RTT        float32
	Jitter     float32
	PacketLoss float32
}

func WriteStatsEntries(entries []StatsEntry) []byte {
	length := 1 + 8 + len(entries)*(8+8+4+4+4)
	data := make([]byte, length)

	index := 0
	encoding.WriteUint8(data, &index, StatsEntryVersion)
	encoding.WriteUint64(data, &index, uint64(len(entries)))

	for i := range entries {
		entry := &entries[i]
		encoding.WriteUint64(data, &index, entry.RelayA)
		encoding.WriteUint64(data, &index, entry.RelayB)
		encoding.WriteFloat32(data, &index, entry.RTT)
		encoding.WriteFloat32(data, &index, entry.Jitter)
		encoding.WriteFloat32(data, &index, entry.PacketLoss)
	}

	return data
}

func ReadStatsEntries(data []byte) ([]StatsEntry, bool) {
	index := 0

	var version uint8
	if !encoding.ReadUint8(data, &index, &version) {
		return nil, false
	}

	var length uint64
	if !encoding.ReadUint64(data, &index, &length) {
		return nil, false
	}

	entries := make([]StatsEntry, length)

	for i := range entries {
		entry := &entries[i]

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
	}

	return entries, true
}

// Save implements the bigquery.ValueSaver interface for an Entry
// so it can be used in Put()
func (e *StatsEntry) Save() (map[string]bigquery.Value, string, error) {
	bqEntry := make(map[string]bigquery.Value)

	bqEntry["timestamp"] = int(e.Timestamp)
	bqEntry["relayA"] = int(e.RelayA)
	bqEntry["relayB"] = int(e.RelayB)
	bqEntry["rtt"] = e.RTT
	bqEntry["jitter"] = e.Jitter
	bqEntry["packetLoss"] = e.PacketLoss

	return bqEntry, "", nil
}
