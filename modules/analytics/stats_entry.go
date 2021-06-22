package analytics

import (
	"strconv"

	"cloud.google.com/go/bigquery"
	"github.com/networknext/backend/modules/encoding"
)

const (
	PingStatsEntryVersion      = uint8(3)
	RelayStatsEntryVersion     = uint8(3)
	RelayNamesHashEntryVersion = uint8(1)

	MaxInstanceIDLength = 64
)

type PingStatsEntry struct {
	Timestamp uint64

	RelayA     uint64
	RelayB     uint64
	RTT        float32
	Jitter     float32
	PacketLoss float32
	Routable   bool

	InstanceID string
	Debug      bool
}

func WritePingStatsEntries(entries []PingStatsEntry) []byte {
	length := 1 + 8 + len(entries)*(8+8+4+4+4+1+MaxInstanceIDLength+1)
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
		encoding.WriteString(data, &index, entry.InstanceID, uint32(MaxInstanceIDLength))
		encoding.WriteBool(data, &index, entry.Debug)
	}

	return data
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
		if version >= 3 {
			if !encoding.ReadString(data, &index, &entry.InstanceID, uint32(MaxInstanceIDLength)) {
				return nil, false
			}
			if !encoding.ReadBool(data, &index, &entry.Debug) {
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
	bqEntry["instanceID"] = e.InstanceID
	bqEntry["debug"] = e.Debug

	return bqEntry, "", nil
}

type RelayStatsEntry struct {
	Timestamp uint64

	ID uint64

	NumSessions uint32
	MaxSessions uint32

	NumRoutable   uint32
	NumUnroutable uint32

	Full bool

	// all of below are deprecated
	Tx                        uint64
	Rx                        uint64
	PeakSessions              uint64
	PeakSentBandwidthMbps     float32
	PeakReceivedBandwidthMbps float32
	CPUUsage                  float32
	MemUsage                  float32

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
}

func WriteRelayStatsEntries(entries []RelayStatsEntry) []byte {
	length := 1 + 8 + len(entries)*int(8+4+4+4+4+4+4+4+4+4+4+4+4+4+4+1)
	data := make([]byte, length)

	index := 0
	encoding.WriteUint8(data, &index, RelayStatsEntryVersion)
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
		encoding.WriteBool(data, &index, entry.Full)
	}

	return data
}

func ReadRelayStatsEntries(data []byte) ([]*RelayStatsEntry, bool) {
	index := 0

	var version uint8
	if !encoding.ReadUint8(data, &index, &version) {
		return nil, false
	}

	var length uint64
	if !encoding.ReadUint64(data, &index, &length) {
		return nil, false
	}

	entries := make([]*RelayStatsEntry, length)

	for i := range entries {
		entry := new(RelayStatsEntry)

		if version >= 2 {
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
		} else {
			if !encoding.ReadUint64(data, &index, &entry.ID) {
				return nil, false
			}

			var numSessions uint64
			if !encoding.ReadUint64(data, &index, &numSessions) {
				return nil, false
			}
			entry.NumSessions = uint32(numSessions)

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

		if version >= 3 {
			if !encoding.ReadBool(data, &index, &entry.Full) {
				return nil, false
			}
		}

		entries[i] = entry
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
	bqEntry["envelope_bandwidth_receive_percent"] = e.EnvelopeReceivedPercent
	bqEntry["actual_bandwidth_send_mbps"] = e.BandwidthSentMbps
	bqEntry["actual_bandwidth_receive_mbps"] = e.BandwidthReceivedMbps
	bqEntry["envelope_bandwidth_send_mbps"] = e.EnvelopeSentMbps
	bqEntry["envelope_bandwidth_receive_mbps"] = e.EnvelopeReceivedMbps
	bqEntry["num_sessions"] = int(e.NumSessions)
	bqEntry["max_sessions"] = int(e.MaxSessions)
	bqEntry["num_routable"] = int(e.NumRoutable)
	bqEntry["num_unroutable"] = int(e.NumUnroutable)

	if e.Full {
		bqEntry["full"] = e.Full
	}

	return bqEntry, "", nil
}

type RouteMatrixStatsEntry struct {
	Timestamp uint64
	Hash      uint64
	IDs       []uint64
}

func WriteRouteMatrixStatsEntry(entry RouteMatrixStatsEntry) []byte {

	length := 1 + 8 + 8 + 4 + (len(entry.IDs) * 8)

	data := make([]byte, length)
	index := 0
	encoding.WriteUint8(data, &index, RelayNamesHashEntryVersion)
	encoding.WriteUint64(data, &index, uint64(entry.Timestamp))
	encoding.WriteUint64(data, &index, uint64(entry.Hash))
	encoding.WriteUint32(data, &index, uint32(len(entry.IDs)))

	for _, ids := range entry.IDs {
		encoding.WriteUint64(data, &index, ids)
	}
	return data
}

func ReadRouteMatrixStatsEntry(data []byte) (*RouteMatrixStatsEntry, bool) {
	index := 0

	entry := new(RouteMatrixStatsEntry)
	var version uint8
	if !encoding.ReadUint8(data, &index, &version) {
		return nil, false
	}

	if !encoding.ReadUint64(data, &index, &entry.Timestamp) {
		return nil, false
	}

	if !encoding.ReadUint64(data, &index, &entry.Hash) {
		return nil, false
	}

	var length uint32
	if !encoding.ReadUint32(data, &index, &length) {
		return nil, false
	}

	entry.IDs = make([]uint64, length)

	for i := range entry.IDs {
		var id uint64
		if !encoding.ReadUint64(data, &index, &id) {
			return nil, false
		}
		entry.IDs[i] = id
	}

	return entry, true
}

// Save implements the bigquery.ValueSaver interface for an Entry
// so it can be used in Put()
func (e *RouteMatrixStatsEntry) Save() (map[string]bigquery.Value, string, error) {
	bqEntry := make(map[string]bigquery.Value)

	bqEntry["timestamp"] = int(e.Timestamp)
	bqEntry["down_relay_hash"] = int64(e.Hash)

	idStrings := make([]string, len(e.IDs))
	for i, id := range e.IDs {
		idStrings[i] = strconv.FormatUint(id, 10)
	}

	bqEntry["down_relay_IDs"] = idStrings
	return bqEntry, "", nil
}
