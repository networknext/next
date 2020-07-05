package billing

import (
	"github.com/networknext/backend/encoding"
)

const BillingEntryVersion = 0

const BillingEntryMaxRelays = 5

const BillingEntryBytes = 4 + 8 + 8 + 4 + 8 + 1 + (6*4) + 1 + (BillingEntryMaxRelays*4) + 8

type BillingEntry struct {
	Version uint32
	Timestamp uint64
	SessionID uint64
	SliceNumber uint32
	BuyerID uint64
	Next bool
	DirectRTT float32
	DirectJitter float32
	DirectPacketLoss float32
	NextRTT float32
	NextJitter float32
	NextPacketLoss float32
	NumNextRelays uint8
	NextRelays [BillingEntryMaxRelays]uint64
	TotalPrice uint64
}

func WriteBillingEntry(entry *BillingEntry) []byte {
	data := make([]byte, BillingEntryBytes)
	index := 0
	encoding.WriteUint32(data, &index, entry.Version)
	encoding.WriteUint64(data, &index, entry.Timestamp)
	encoding.WriteUint64(data, &index, entry.SessionID)
	encoding.WriteUint32(data, &index, entry.SliceNumber)
	encoding.WriteUint64(data, &index, entry.BuyerID)
	if entry.Next {
		encoding.WriteUint8(data, &index, 1)
	} else {
		encoding.WriteUint8(data, &index, 0)
	}
	encoding.WriteFloat32(data, &index, entry.DirectRTT)
	encoding.WriteFloat32(data, &index, entry.DirectJitter)
	encoding.WriteFloat32(data, &index, entry.DirectPacketLoss)
	encoding.WriteFloat32(data, &index, entry.NextRTT)
	encoding.WriteFloat32(data, &index, entry.NextJitter)
	encoding.WriteFloat32(data, &index, entry.NextPacketLoss)
	encoding.WriteUint8(data, &index, entry.NumNextRelays)
	for i := 0; i < BillingEntryMaxRelays; i++ {
		encoding.WriteUint64(data, &index, entry.NextRelays[i])
	}
	encoding.WriteUint64(data, &index, entry.TotalPrice)
	return data
}

func ReadBillingEntry(entry *BillingEntry, data []byte)  {
	index := 0
	encoding.ReadUint32(data, &index, &entry.Version)
	encoding.ReadUint64(data, &index, &entry.Timestamp)
	encoding.ReadUint64(data, &index, &entry.SessionID)
	encoding.ReadUint32(data, &index, &entry.SliceNumber)
	encoding.ReadUint64(data, &index, &entry.BuyerID)
	var next uint8
	encoding.ReadUint8(data, &index, &next)
	if next != 0 {
		entry.Next = true
	}
	encoding.ReadFloat32(data, &index, &entry.DirectRTT)
	encoding.ReadFloat32(data, &index, &entry.DirectJitter)
	encoding.ReadFloat32(data, &index, &entry.DirectPacketLoss)
	encoding.ReadFloat32(data, &index, &entry.NextRTT)
	encoding.ReadFloat32(data, &index, &entry.NextJitter)
	encoding.ReadFloat32(data, &index, &entry.NextPacketLoss)
	encoding.ReadUint8(data, &index, &entry.NumNextRelays)
	for i := 0; i < BillingEntryMaxRelays; i++ {
		encoding.ReadUint64(data, &index, &entry.NextRelays[i])
	}
	encoding.ReadUint64(data, &index, &entry.TotalPrice)
}
