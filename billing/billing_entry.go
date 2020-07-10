package billing

import (
	"github.com/networknext/backend/encoding"
)

const BillingEntryVersion = uint8(3)

const BillingEntryMaxRelays = 5

const MaxBillingEntryBytes = 1 + 8 + 4 + 8 + 1 + (6 * 4) + 1 + (BillingEntryMaxRelays * 8) + 8 + 8 + 8 + 4 + 8 + 8

type BillingEntry struct {
	Timestamp                 uint64 // IMPORTANT: Timestamp is not serialized. Pubsub already has the timestamp so we use that instead.
	Version                   uint8
	BuyerID                   uint64
	SessionID                 uint64
	SliceNumber               uint32
	DirectRTT                 float32
	DirectJitter              float32
	DirectPacketLoss          float32
	Next                      bool
	NextRTT                   float32
	NextJitter                float32
	NextPacketLoss            float32
	NumNextRelays             uint8
	NextRelays                [BillingEntryMaxRelays]uint64
	TotalPrice                uint64
	ClientToServerPacketsLost uint64
	ServerToClientPacketsLost uint64
	Committed                 bool
	Flagged                   bool
	Multipath                 bool
	Initial                   bool
	NextBytesUp               uint64
	NextBytesDown             uint64
}

func WriteBillingEntry(entry *BillingEntry) []byte {
	data := make([]byte, MaxBillingEntryBytes)
	index := 0
	encoding.WriteUint8(data, &index, BillingEntryVersion)
	encoding.WriteUint64(data, &index, entry.BuyerID)
	encoding.WriteUint64(data, &index, entry.SessionID)
	encoding.WriteUint32(data, &index, entry.SliceNumber)
	encoding.WriteFloat32(data, &index, entry.DirectRTT)
	encoding.WriteFloat32(data, &index, entry.DirectJitter)
	encoding.WriteFloat32(data, &index, entry.DirectPacketLoss)
	if entry.Next {
		encoding.WriteBool(data, &index, true)
		encoding.WriteFloat32(data, &index, entry.NextRTT)
		encoding.WriteFloat32(data, &index, entry.NextJitter)
		encoding.WriteFloat32(data, &index, entry.NextPacketLoss)
		encoding.WriteUint8(data, &index, entry.NumNextRelays)
		for i := 0; i < int(entry.NumNextRelays); i++ {
			encoding.WriteUint64(data, &index, entry.NextRelays[i])
		}
		encoding.WriteUint64(data, &index, entry.TotalPrice)
	} else {
		encoding.WriteBool(data, &index, false)
	}
	encoding.WriteUint64(data, &index, entry.ClientToServerPacketsLost)
	encoding.WriteUint64(data, &index, entry.ServerToClientPacketsLost)

	encoding.WriteBool(data, &index, entry.Committed)
	encoding.WriteBool(data, &index, entry.Flagged)
	encoding.WriteBool(data, &index, entry.Multipath)

	if entry.Initial {
		encoding.WriteBool(data, &index, true)
		encoding.WriteUint64(data, &index, entry.NextBytesUp)
		encoding.WriteUint64(data, &index, entry.NextBytesDown)
	} else {
		encoding.WriteBool(data, &index, false)
	}

	return data[:index]
}

func ReadBillingEntry(entry *BillingEntry, data []byte) bool {
	index := 0
	if !encoding.ReadUint8(data, &index, &entry.Version) {
		return false
	}
	if entry.Version > BillingEntryVersion {
		return false
	}
	if !encoding.ReadUint64(data, &index, &entry.BuyerID) {
		return false
	}
	if !encoding.ReadUint64(data, &index, &entry.SessionID) {
		return false
	}
	if !encoding.ReadUint32(data, &index, &entry.SliceNumber) {
		return false
	}
	if !encoding.ReadFloat32(data, &index, &entry.DirectRTT) {
		return false
	}
	if !encoding.ReadFloat32(data, &index, &entry.DirectJitter) {
		return false
	}
	if !encoding.ReadFloat32(data, &index, &entry.DirectPacketLoss) {
		return false
	}
	if !encoding.ReadBool(data, &index, &entry.Next) {
		return false
	}

	if entry.Next {
		if !encoding.ReadFloat32(data, &index, &entry.NextRTT) {
			return false
		}
		if !encoding.ReadFloat32(data, &index, &entry.NextJitter) {
			return false
		}
		if !encoding.ReadFloat32(data, &index, &entry.NextPacketLoss) {
			return false
		}
		if !encoding.ReadUint8(data, &index, &entry.NumNextRelays) {
			return false
		}
		if entry.NumNextRelays > BillingEntryMaxRelays {
			return false
		}
		for i := 0; i < int(entry.NumNextRelays); i++ {
			if !encoding.ReadUint64(data, &index, &entry.NextRelays[i]) {
				return false
			}
		}
		if !encoding.ReadUint64(data, &index, &entry.TotalPrice) {
			return false
		}
	}
	if entry.Version >= 2 {
		if !encoding.ReadUint64(data, &index, &entry.ClientToServerPacketsLost) {
			return false
		}
		if !encoding.ReadUint64(data, &index, &entry.ServerToClientPacketsLost) {
			return false
		}
	}

	if entry.Version >= 3 {
		if !encoding.ReadBool(data, &index, &entry.Committed) {
			return false
		}

		if !encoding.ReadBool(data, &index, &entry.Flagged) {
			return false
		}

		if !encoding.ReadBool(data, &index, &entry.Multipath) {
			return false
		}

		if !encoding.ReadBool(data, &index, &entry.Initial) {
			return false
		}

		if entry.Initial {
			if !encoding.ReadUint64(data, &index, &entry.NextBytesUp) {
				return false
			}
			if !encoding.ReadUint64(data, &index, &entry.NextBytesDown) {
				return false
			}
		}
	}

	return true
}
