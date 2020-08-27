package ghost_army

import (
	"time"

	"github.com/networknext/backend/encoding"
)

// Follows the csv schema
type Entry struct {
	SessionID                 int64
	Timestamp                 time.Time
	BuyerId                   int64
	SliceNumber               int64
	Next                      bool
	DirectRTT                 float64
	DirectJitter              float64
	DirectPacketLoss          float64
	NextRTT                   float64
	NextJitter                float64
	NextPacketLoss            float64
	NextRelays                []int64
	TotalPrice                int64
	ClientToServerPacketsLost int64
	ServerToClientPacketsLost int64
	Committed                 bool
	Flagged                   bool
	Multipath                 bool
	NextBytesUp               int64
	NextBytesDown             int64
	Initial                   bool
	DatacenterID              int64
	RttReduction              bool
	PacketLossReduction       bool
	NextRelaysPrice           []int64
	Userhash                  int64
}

func (self *Entry) MarshalBinary() ([]byte, error) {
	size := 8 + 64 + 8 + 8 + 1 + 8 + 8 + 8 + 8 + 8 + 8 + 8*len(self.NextRelays) + 8 + 8 + 8 + 1 + 1 + 1 + 8 + 8 + 1 + 8 + 1 + 1 + 8*len(self.NextRelaysPrice) + 8
	bin := make([]byte, size)
	index := 0

	casterInt64 := func(x int64) {
		encoding.WriteUint64(bin, &index, uint64(self.SessionID))
	}

	casterBool := func(x bool) {
		if x {
			encoding.WriteUint8(bin, &index, 1)
		} else {
			encoding.WriteUint8(bin, &index, 0)
		}
	}

	casterInt64(self.SessionID)
	casterInt64(self.Timestamp.Unix())
	casterInt64(self.BuyerId)
	encoding.WriteFloat64(bin, &index, self.DirectRTT)
	encoding.WriteFloat64(bin, &index, self.DirectJitter)
	encoding.WriteFloat64(bin, &index, self.DirectPacketLoss)
	encoding.WriteFloat64(bin, &index, self.NextRTT)
	encoding.WriteFloat64(bin, &index, self.NextJitter)
	encoding.WriteFloat64(bin, &index, self.NextPacketLoss)
	casterInt64(int64(len(self.NextRelays)))
	for _, relay := range self.NextRelays {
		casterInt64(relay)
	}
	casterInt64(self.TotalPrice)
	casterInt64(self.ClientToServerPacketsLost)
	casterInt64(self.ServerToClientPacketsLost)
	casterBool(self.Committed)
	casterBool(self.Flagged)
	casterBool(self.Multipath)
	casterInt64(self.NextBytesUp)
	casterInt64(self.NextBytesDown)
	casterBool(self.Initial)
	casterInt64(self.DatacenterID)
	casterBool(self.RttReduction)
	casterBool(self.PacketLossReduction)
	casterInt64(int64(len(self.NextRelaysPrice)))
	for _, price := range self.NextRelaysPrice {
		casterInt64(price)
	}
	casterInt64(self.Userhash)

	return bin, nil
}
