// Package ghostarmy contains types shared between the ghost_army program & the generator
package ghostarmy

import (
	"time"

	"github.com/networknext/backend/encoding"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/transport"
)

// Entry is the contents of the csv
type Entry struct {
	SessionID                 int64
	Timestamp                 int64 // not actual timestamp, but # of seconds into the day
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

	if !casterInt64(&self.Timestamp) {
		return false
	}

	if !casterInt64(&self.BuyerID) {
		return false
	}

	if !casterInt64(&self.SliceNumber) {
		return false
	}

	if !casterBool(&self.Next) {
		return false
	}

	if !encoding.ReadFloat64(bin, index, &self.DirectRTT) {
		return false
	}

	if !encoding.ReadFloat64(bin, index, &self.DirectJitter) {
		return false
	}

	if !encoding.ReadFloat64(bin, index, &self.DirectPacketLoss) {
		return false
	}

	if !encoding.ReadFloat64(bin, index, &self.NextRTT) {
		return false
	}

	if !encoding.ReadFloat64(bin, index, &self.NextJitter) {
		return false
	}

	if !encoding.ReadFloat64(bin, index, &self.NextPacketLoss) {
		return false
	}

	var relayCount uint64
	if !encoding.ReadUint64(bin, index, &relayCount) {
		return false
	}

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

func (self *Entry) Into(data *transport.SessionPortalData) {
	// meta
	{
		meta := &data.Meta
		meta.ID = uint64(self.SessionID)
		meta.UserHash = uint64(self.Userhash)
		meta.DatacenterName = "TODO"   // TODO
		meta.DatacenterAlias = "TODO"  // TODO
		meta.OnNetworkNext = self.Next // TODO check
		meta.NextRTT = self.NextRTT
		meta.DirectRTT = self.DirectRTT

		var deltaRTT float64
		if !self.Next {
			deltaRTT = 0
		} else {
			deltaRTT = self.DirectRTT - self.NextRTT
			if self.NextRTT == 0 || deltaRTT < 0 {
				deltaRTT = 0
			}
		}

		meta.DeltaRTT = deltaRTT

		meta.Location = routing.Location{} // TODO

		meta.ClientAddr = "TODO" // TODO
		meta.ServerAddr = "TODO" // TODO
		meta.Hops = make([]transport.RelayHop, len(self.NextRelays))
		for i, id := range self.NextRelays {
			meta.Hops[i].ID = uint64(id)
			meta.Hops[i].Name = "TODO" // TODO
		}
		meta.SDK = "1.2.3"                                           // TODO get valid version
		meta.Connection = 0                                          // TODO what is this?
		meta.NearbyRelays = make([]transport.NearRelayPortalData, 0) // TODO somehow come up with a list
		meta.Platform = 0                                            // TODO what is this?
		meta.BuyerID = uint64(self.BuyerID)
	}

	// slice
	{
		slice := &data.Slice
		slice.Timestamp = time.Unix(self.Timestamp, 0)
		slice.Next = routing.Stats{
			RTT:        self.NextRTT,
			Jitter:     self.NextJitter,
			PacketLoss: self.NextPacketLoss,
		}
		slice.Direct = routing.Stats{
			RTT:        self.DirectRTT,
			Jitter:     self.DirectJitter,
			PacketLoss: self.DirectPacketLoss,
		}
		slice.Envelope = routing.Envelope{
			Up:   self.NextBytesUp,   // TODO check
			Down: self.NextBytesDown, // TODO check
		}
		slice.OnNetworkNext = self.Next
		slice.IsMultiPath = self.Multipath
		slice.IsTryBeforeYouBuy = false // TODO check
	}
}
