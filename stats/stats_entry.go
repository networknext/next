package stats

import (
	"github.com/networknext/backend/encoding"
)

const (
	StatsEntryVersion = uint8(1)
	MaxStatEntryBytes = 1 + 8 + 8 + 4 + 4 + 4
)

type StatsEntry struct {
	Version    uint8
	RelayA     uint64
	RelayB     uint64
	RTT        float32
	Jitter     float32
	PacketLoss float32
}

func WriteStatsEntry(entry StatsEntry) []byte {
	index := 0
	data := make([]byte, MaxStatEntryBytes)

	encoding.WriteUint8(data, &index, StatsEntryVersion)
	encoding.WriteUint64(data, &index, entry.RelayA)
	encoding.WriteUint64(data, &index, entry.RelayB)
	encoding.WriteFloat32(data, &index, entry.RTT)
	encoding.WriteFloat32(data, &index, entry.Jitter)
	encoding.WriteFloat32(data, &index, entry.PacketLoss)

	return data[:index]
}

func ReadStatsEntry(entry *StatsEntry, data []byte) bool {
	index := 0

	if !encoding.ReadUint8(data, &index, &entry.Version) {
		return false
	}

	if !encoding.ReadUint64(data, &index, &entry.RelayA) {
		return false
	}

	if !encoding.ReadUint64(data, &index, &entry.RelayB) {
		return false
	}

	if !encoding.ReadFloat32(data, &index, &entry.RTT) {
		return false
	}

	if !encoding.ReadFloat32(data, &index, &entry.Jitter) {
		return false
	}

	if !encoding.ReadFloat32(data, &index, &entry.PacketLoss) {
		return false
	}

	return true
}
