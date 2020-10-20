package analytics

import (
	"cloud.google.com/go/bigquery"
	"github.com/networknext/backend/encoding"
	"github.com/networknext/backend/routing"
)

const (
	PingStatsEntryVersion  = uint8(2)
	RelayStatsEntryVersion = uint8(2)
)

type PingStatsEntry struct {
	Timestamp uint64

	RelayA     uint64
	RelayB     uint64
	RTT        float32
	Jitter     float32
	PacketLoss float32
	Routable   bool
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
					Routable:   rtt != routing.InvalidRouteValue && jitter != routing.InvalidRouteValue && pl != routing.InvalidRouteValue,
				}
			}
		}
	}

	return entries
}

func WritePingStatsEntries(entries []PingStatsEntry) []byte {
	length := 1 + 8 + len(entries)*(8+8+4+4+4+1)
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

		if !encoding.ReadBool(data, &index, &entry.Routable) {
			return nil, false
		}
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

type RelayStatsEntry struct {
	Timestamp uint64

	ID uint64

	CPUUsage float32
	MemUsage float32

	// percent = (sent||received) / nic speed

	BandwidthSentPercent     float32
	BandwidthReceivedPercent float32

	// percent = bandwidth_(sent||received) / envelope_(sent||received)

	EnvelopeSentPercent     float32
	EnvelopeReceivedPercent float32

	BandwidthSentMbps     float32
	BandwidthReceivedMbps float32

	EnvelopeSentMbps     float32
	EnvelopeReceivedMbps float32

	NumSessions uint32
	MaxSessions uint32

	NumRoutable   uint32
	NumUnroutable uint32
}

func WriteRelayStatsEntries(entries []RelayStatsEntry) []byte {
	length := 1 + 8 + len(entries)*int(8+4+4+4+4+4+4+4+4+4+4+4+4+4+4)
	data := make([]byte, length)

	index := 0
	encoding.WriteUint8(data, &index, PingStatsEntryVersion)
	encoding.WriteUint64(data, &index, uint64(len(entries)))

	for i := range entries {
		entry := &entries[i]
		encoding.WriteUint64(data, &index, entry.ID)
		encoding.WriteFloat32(data, &index, entry.CPUUsage)
		encoding.WriteFloat32(data, &index, entry.MemUsage)
		encoding.WriteFloat32(data, &index, entry.BandwidthSentPercent)
		encoding.WriteFloat32(data, &index, entry.BandwidthReceivedPercent)
		encoding.WriteFloat32(data, &index, entry.EnvelopeSentPercent)
		encoding.WriteFloat32(data, &index, entry.EnvelopeReceivedPercent)
		encoding.WriteFloat32(data, &index, entry.BandwidthSentMbps)
		encoding.WriteFloat32(data, &index, entry.BandwidthReceivedMbps)
		encoding.WriteFloat32(data, &index, entry.EnvelopeSentMbps)
		encoding.WriteFloat32(data, &index, entry.EnvelopeReceivedMbps)
		encoding.WriteUint32(data, &index, entry.NumSessions)
		encoding.WriteUint32(data, &index, entry.MaxSessions)
		encoding.WriteUint32(data, &index, entry.NumRoutable)
		encoding.WriteUint32(data, &index, entry.NumUnroutable)
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

		if !encoding.ReadFloat32(data, &index, &entry.CPUUsage) {
			return nil, false
		}

		if !encoding.ReadFloat32(data, &index, &entry.MemUsage) {
			return nil, false
		}

		if !encoding.ReadFloat32(data, &index, &entry.BandwidthSentPercent) {
			return nil, false
		}

		if !encoding.ReadFloat32(data, &index, &entry.BandwidthReceivedPercent) {
			return nil, false
		}

		if !encoding.ReadFloat32(data, &index, &entry.EnvelopeSentPercent) {
			return nil, false
		}

		if !encoding.ReadFloat32(data, &index, &entry.EnvelopeReceivedPercent) {
			return nil, false
		}

		if !encoding.ReadFloat32(data, &index, &entry.BandwidthSentMbps) {
			return nil, false
		}

		if !encoding.ReadFloat32(data, &index, &entry.BandwidthReceivedMbps) {
			return nil, false
		}

		if !encoding.ReadFloat32(data, &index, &entry.EnvelopeSentMbps) {
			return nil, false
		}

		if !encoding.ReadFloat32(data, &index, &entry.EnvelopeReceivedMbps) {
			return nil, false
		}

		if !encoding.ReadUint32(data, &index, &entry.NumSessions) {
			return nil, false
		}

		if !encoding.ReadUint32(data, &index, &entry.MaxSessions) {
			return nil, false
		}

		if !encoding.ReadUint32(data, &index, &entry.NumRoutable) {
			return nil, false
		}

		if !encoding.ReadUint32(data, &index, &entry.NumUnroutable) {
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
	bqEntry["relay_id"] = int(e.ID)
	bqEntry["cpu_percent"] = e.CPUUsage
	bqEntry["memory_percent"] = e.MemUsage
	bqEntry["actual_bandwidth_send_percent"] = e.BandwidthSentPercent
	bqEntry["actual_bandwidth_receive_percent"] = e.BandwidthReceivedPercent
	bqEntry["envelope_bandwidth_send_percent"] = e.EnvelopeSentPercent
	bqEntry["envelope_bandwidth_received_percent"] = e.EnvelopeReceivedPercent
	bqEntry["actual_bandwidth_send_mbps"] = e.BandwidthSentMbps
	bqEntry["actual_bandwidth_receive_mbps"] = e.BandwidthReceivedMbps
	bqEntry["envelope_bandwidth_send_mbps"] = e.EnvelopeSentMbps
	bqEntry["envelope_bandwidth_receive_mbps"] = e.EnvelopeReceivedMbps
	bqEntry["num_sessions"] = int(e.NumSessions)
	bqEntry["max_sessions"] = int(e.MaxSessions)
	bqEntry["num_routable"] = int(e.NumRoutable)
	bqEntry["num_unroutable"] = int(e.NumUnroutable)

	return bqEntry, "", nil
}
