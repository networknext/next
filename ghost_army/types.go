// Package ghostarmy contains types shared between the ghost_army program & the generator
package ghostarmy

import (
	"fmt"

	"github.com/networknext/backend/encoding"
)

// Entry is the contents of the csv
type Entry struct {
	SessionID                 int64
	Timestamp                 int64
	BuyerID                   int64
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

func (self *Entry) ReadFrom(bin []byte, index *int) bool {
	casterInt64 := func(x *int64) bool {
		var y uint64
		if !encoding.ReadUint64(bin, index, &y) {
			return false
		}
		*x = int64(y)
		return true
	}

	casterBool := func(x *bool) bool {
		var y uint8
		if !encoding.ReadUint8(bin, index, &y) {
			return false
		}
		*x = y == 1
		return true
	}

	if !casterInt64(&self.SessionID) {
		return false
	}

	fmt.Printf("session id = %d\n", self.SessionID)

	if !casterInt64(&self.Timestamp) {
		return false
	}

	fmt.Printf("timestamp = %d\n", self.Timestamp)

	if !casterInt64(&self.BuyerID) {
		return false
	}

	fmt.Printf("buyer id = %d\n", self.BuyerID)

	if !casterInt64(&self.SliceNumber) {
		return false
	}

	fmt.Printf("slice number = %d\n", self.SliceNumber)

	if !casterBool(&self.Next) {
		return false
	}

	fmt.Printf("next = %v\n", self.Next)

	if !encoding.ReadFloat64(bin, index, &self.DirectRTT) {
		return false
	}

	fmt.Printf("dir rtt = %f\n", self.DirectRTT)

	if !encoding.ReadFloat64(bin, index, &self.DirectJitter) {
		return false
	}

	fmt.Printf("dir jitter = %f\n", self.DirectJitter)

	if !encoding.ReadFloat64(bin, index, &self.DirectPacketLoss) {
		return false
	}

	fmt.Printf("dir pl = %f\n", self.DirectPacketLoss)

	if !encoding.ReadFloat64(bin, index, &self.NextRTT) {
		return false
	}

	fmt.Printf("next rtt = %f\n", self.NextRTT)

	if !encoding.ReadFloat64(bin, index, &self.NextJitter) {
		return false
	}

	fmt.Printf("next jitter = %f\n", self.NextJitter)

	if !encoding.ReadFloat64(bin, index, &self.NextPacketLoss) {
		return false
	}

	fmt.Printf("next pl = %f\n", self.NextPacketLoss)

	fmt.Printf("index = %d\n", *index)

	var relayCount uint64
	if !encoding.ReadUint64(bin, index, &relayCount) {
		return false
	}

	fmt.Printf("count = %d\n", relayCount)
	self.NextRelays = make([]int64, relayCount)

	for i := uint64(0); i < relayCount; i++ {
		if !casterInt64(&self.NextRelays[i]) {
			return false
		}
	}

	if !casterInt64(&self.TotalPrice) {
		return false
	}

	if !casterInt64(&self.ClientToServerPacketsLost) {
		return false
	}

	if !casterInt64(&self.ServerToClientPacketsLost) {
		return false
	}

	if !casterBool(&self.Committed) {
		return false
	}

	if !casterBool(&self.Flagged) {
		return false
	}

	if !casterBool(&self.Multipath) {
		return false
	}

	if !casterInt64(&self.NextBytesUp) {
		return false
	}

	if !casterInt64(&self.NextBytesDown) {
		return false
	}

	if !casterBool(&self.Initial) {
		return false
	}

	if !casterInt64(&self.DatacenterID) {
		return false
	}

	if !casterBool(&self.RttReduction) {
		return false
	}

	if !casterBool(&self.PacketLossReduction) {
		return false
	}

	var priceCount uint64
	if !encoding.ReadUint64(bin, index, &priceCount) {
		return false
	}

	self.NextRelaysPrice = make([]int64, priceCount)

	for i := uint64(0); i < priceCount; i++ {
		if !casterInt64(&self.NextRelaysPrice[i]) {
			return false
		}
	}

	return casterInt64(&self.Userhash)
}

func (self *Entry) MarshalBinary() ([]byte, error) {
	size := 8 + 8 + 8 + 8 + 1 + 8 + 8 + 8 + 8 + 8 + 8 + 8 + 8*len(self.NextRelays) + 8 + 8 + 8 + 1 + 1 + 1 + 8 + 8 + 1 + 8 + 1 + 1 + 8 + 8*len(self.NextRelaysPrice) + 8
	bin := make([]byte, size)
	index := 0

	encoding.WriteUint64(bin, &index, uint64(self.SessionID))
	encoding.WriteUint64(bin, &index, uint64(self.Timestamp))
	encoding.WriteUint64(bin, &index, uint64(self.BuyerID))
	encoding.WriteUint64(bin, &index, uint64(self.SliceNumber))
	if self.Next {
		encoding.WriteUint8(bin, &index, 1)
	} else {
		encoding.WriteUint8(bin, &index, 0)
	}
	encoding.WriteFloat64(bin, &index, self.DirectRTT)
	encoding.WriteFloat64(bin, &index, self.DirectJitter)
	encoding.WriteFloat64(bin, &index, self.DirectPacketLoss)
	encoding.WriteFloat64(bin, &index, self.NextRTT)
	encoding.WriteFloat64(bin, &index, self.NextJitter)
	encoding.WriteFloat64(bin, &index, self.NextPacketLoss)
	encoding.WriteUint64(bin, &index, uint64(len(self.NextRelays)))
	for _, relay := range self.NextRelays {
		encoding.WriteUint64(bin, &index, uint64(relay))
	}
	encoding.WriteUint64(bin, &index, uint64(self.TotalPrice))
	encoding.WriteUint64(bin, &index, uint64(self.ClientToServerPacketsLost))
	encoding.WriteUint64(bin, &index, uint64(self.ServerToClientPacketsLost))
	if self.Committed {
		encoding.WriteUint8(bin, &index, 1)
	} else {
		encoding.WriteUint8(bin, &index, 0)
	}
	if self.Flagged {
		encoding.WriteUint8(bin, &index, 1)
	} else {
		encoding.WriteUint8(bin, &index, 0)
	}
	if self.Multipath {
		encoding.WriteUint8(bin, &index, 1)
	} else {
		encoding.WriteUint8(bin, &index, 0)
	}
	encoding.WriteUint64(bin, &index, uint64(self.NextBytesUp))
	encoding.WriteUint64(bin, &index, uint64(self.NextBytesDown))
	if self.Initial {
		encoding.WriteUint8(bin, &index, 1)
	} else {
		encoding.WriteUint8(bin, &index, 0)
	}
	encoding.WriteUint64(bin, &index, uint64(self.DatacenterID))
	if self.RttReduction {
		encoding.WriteUint8(bin, &index, 1)
	} else {
		encoding.WriteUint8(bin, &index, 0)
	}
	if self.PacketLossReduction {
		encoding.WriteUint8(bin, &index, 1)
	} else {
		encoding.WriteUint8(bin, &index, 0)
	}
	encoding.WriteUint64(bin, &index, uint64(int64(len(self.NextRelaysPrice))))
	for _, price := range self.NextRelaysPrice {
		encoding.WriteUint64(bin, &index, uint64(price))
	}
	encoding.WriteUint64(bin, &index, uint64(self.Userhash))

	return bin, nil
}
