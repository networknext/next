package analytics

import (
	"cloud.google.com/go/bigquery"
	"github.com/networknext/backend/encoding"
	"github.com/networknext/backend/routing"
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

func ExtractPingStats(statsdb *routing.StatsDatabase) []PingStatsEntry {
	length := routing.TriMatrixLength(len(statsdb.Entries))
	entries := make([]PingStatsEntry, length)

	if length > 0 { // prevent crash with only 1 relay
		ids := make([]uint64, len(statsdb.Entries))

		idx := 0
		for k := range statsdb.Entries {
			ids[idx] = k
			idx++
		}

		for i := 1; i < len(ids); i++ {
			for j := 0; j < i; j++ {
				idA := ids[i]
				idB := ids[j]

				rtt, jitter, pl := statsdb.GetSample(idA, idB)

				entries[routing.TriMatrixIndex(i, j)] = PingStatsEntry{
					RelayA:     idA,
					RelayB:     idB,
					RTT:        rtt,
					Jitter:     jitter,
					PacketLoss: pl,
				}
			}
		}
	}

	return entries
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
	MemUsage    float32
	Tx          uint64
	Rx          uint64

	// Maximum traffic stats since the last upload for this relay
	PeakSessions              uint64
	PeakSentBandwidthMbps     float32
	PeakReceivedBandwidthMbps float32
}

func WriteRelayStatsEntries(entries []RelayStatsEntry) []byte {
	length := 1 + 8 + len(entries)*(8+8+4+4+8+8+8+4+4)
	data := make([]byte, length)

	index := 0
	encoding.WriteUint8(data, &index, PingStatsEntryVersion)
	encoding.WriteUint64(data, &index, uint64(len(entries)))

	for i := range entries {
		entry := &entries[i]
		encoding.WriteUint64(data, &index, entry.ID)
		encoding.WriteUint64(data, &index, entry.NumSessions)
		encoding.WriteFloat32(data, &index, entry.CPUUsage)
		encoding.WriteFloat32(data, &index, entry.MemUsage)
		encoding.WriteUint64(data, &index, entry.Tx)
		encoding.WriteUint64(data, &index, entry.Rx)
		encoding.WriteUint64(data, &index, entry.PeakSessions)
		encoding.WriteFloat32(data, &index, entry.PeakSentBandwidthMbps)
		encoding.WriteFloat32(data, &index, entry.PeakReceivedBandwidthMbps)
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

		if !encoding.ReadFloat32(data, &index, &entry.MemUsage) {
			return nil, false
		}

		if !encoding.ReadUint64(data, &index, &entry.Tx) {
			return nil, false
		}

		if !encoding.ReadUint64(data, &index, &entry.Rx) {
			return nil, false
		}

		if !encoding.ReadUint64(data, &index, &entry.PeakSessions) {
			return nil, false
		}

		if !encoding.ReadFloat32(data, &index, &entry.PeakSentBandwidthMbps) {
			return nil, false
		}

		if !encoding.ReadFloat32(data, &index, &entry.PeakReceivedBandwidthMbps) {
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
	bqEntry["relayID"] = int(e.ID)
	bqEntry["numSessions"] = int(e.NumSessions)
	bqEntry["cpu"] = e.CPUUsage
	bqEntry["mem"] = e.MemUsage
	bqEntry["tx"] = e.Tx
	bqEntry["rx"] = e.Rx
	bqEntry["peakSessions"] = e.PeakSessions
	bqEntry["peakSentBandwidthMbps"] = e.PeakSentBandwidthMbps
	bqEntry["peakReceivedBandwidthMbps"] = e.PeakReceivedBandwidthMbps

	return bqEntry, "", nil
}
