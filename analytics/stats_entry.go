package analytics

import (
	"cloud.google.com/go/bigquery"
	"github.com/networknext/backend/encoding"
)

const (
	PingStatsEntryVersion  = uint8(1)
	RelayStatsEntryVersion = uint8(1)
)

type PingStatsEntry struct {
	Timestamp uint64

	RelayA     uint64
	RelayB     uint64
	RTT        float32
	Jitter     float32
	PacketLoss float32
}

func WritePingStatsEntries(entries []PingStatsEntry) []byte {
	length := 1 + 8 + len(entries)*(8+8+4+4+4)
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
	}

	return data
}

func ReadPingStatsEntries(data []byte) ([]PingStatsEntry, bool) {
	index := 0

	var version uint8
	if !encoding.ReadUint8(data, &index, &version) {
		return nil, false
	}

	var length uint64
	if !encoding.ReadUint64(data, &index, &length) {
		return nil, false
	}

	entries := make([]PingStatsEntry, length)

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

func (e *PingStatsEntry) Save() (map[string]bigquery.Value, string, error) {
	bqEntry := make(map[string]bigquery.Value)

	bqEntry["timestamp"] = int(e.Timestamp)
	bqEntry["relayA"] = int(e.RelayA)
	bqEntry["relayB"] = int(e.RelayB)
	bqEntry["rtt"] = e.RTT
	bqEntry["jitter"] = e.Jitter
	bqEntry["packetLoss"] = e.PacketLoss

	return bqEntry, "", nil
}

type RelayStatsEntry struct {
	Timestamp uint64

	ID          uint64
	NumSessions uint64
	CPUUsage    float32
	RAMUsage    float32
	Tx          uint64
	Rx          uint64
}

func WriteRelayStatsEntries(entries []RelayStatsEntry) []byte {
	length := 1 + 8 + len(entries)*(8+8+4+4+8+8)
	data := make([]byte, length)

	index := 0
	encoding.WriteUint8(data, &index, PingStatsEntryVersion)
	encoding.WriteUint64(data, &index, uint64(len(entries)))

	for i := range entries {
		entry := &entries[i]
		encoding.WriteUint64(data, &index, entry.ID)
		encoding.WriteUint64(data, &index, entry.NumSessions)
		encoding.WriteFloat32(data, &index, entry.CPUUsage)
		encoding.WriteFloat32(data, &index, entry.RAMUsage)
		encoding.WriteUint64(data, &index, entry.Tx)
		encoding.WriteUint64(data, &index, entry.Rx)
	}

	return data
}

func ReadRelayStatsEntries(data []byte) ([]RelayStatsEntry, bool) {
	index := 0

	var version uint8
	if !encoding.ReadUint8(data, &index, &version) {
		return nil, false
	}

	var length uint64
	if !encoding.ReadUint64(data, &index, &length) {
		return nil, false
	}

	entries := make([]RelayStatsEntry, length)

	for i := range entries {
		entry := &entries[i]

		if !encoding.ReadUint64(data, &index, &entry.ID) {
			return nil, false
		}

		if !encoding.ReadUint64(data, &index, &entry.NumSessions) {
			return nil, false
		}

		if !encoding.ReadFloat32(data, &index, &entry.CPUUsage) {
			return nil, false
		}

		if !encoding.ReadFloat32(data, &index, &entry.RAMUsage) {
			return nil, false
		}

		if !encoding.ReadUint64(data, &index, &entry.Tx) {
			return nil, false
		}

		if !encoding.ReadUint64(data, &index, &entry.Rx) {
			return nil, false
		}
	}

	return entries, true
}

// Save implements the bigquery.ValueSaver interface for an Entry
// so it can be used in Put()
func (e *RelayStatsEntry) Save() (map[string]bigquery.Value, string, error) {
	bqEntry := make(map[string]bigquery.Value)

	bqEntry["timestamp"] = int(e.Timestamp)
	bqEntry["realyID"] = int(e.ID)
	bqEntry["numSessions"] = int(e.NumSessions)
	bqEntry["cpuUsage"] = e.CPUUsage
	bqEntry["ramUsage"] = e.RAMUsage
	bqEntry["Tx"] = e.Tx
	bqEntry["Rx"] = e.Rx

	return bqEntry, "", nil
}
