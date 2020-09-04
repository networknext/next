// Package ghostarmy contains types shared between the ghost_army program & the generator
package ghostarmy

import (
	"math"
	"time"

	"github.com/networknext/backend/encoding"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/transport"
)

const (
	// All the same for now, but just in case they change later
	LocalBuyerID   = 0
	DevBuyerID     = 0
	ProdBuyerID    = 0
	StagingBuyerID = 0
)

func GhostArmyBuyerID(env string) uint64 {
	switch env {
	case "local":
		return LocalBuyerID
	case "dev":
		return DevBuyerID
	case "staging":
		return StagingBuyerID
	case "prod":
		return ProdBuyerID
	}

	return 0
}

type StrippedDatacenter struct {
	Name string
	Lat  float64
	Long float64
}

type DatacenterMap = map[uint64]StrippedDatacenter

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
	Latitude                  float64
	Longitude                 float64
	ISP                       string
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

	if !encoding.ReadBool(bin, index, &self.Next) {
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

	if !encoding.ReadBool(bin, index, &self.Committed) {
		return false
	}

	if !encoding.ReadBool(bin, index, &self.Flagged) {
		return false
	}

	if !encoding.ReadBool(bin, index, &self.Multipath) {
		return false
	}

	if !casterInt64(&self.NextBytesUp) {
		return false
	}

	if !casterInt64(&self.NextBytesDown) {
		return false
	}

	if !encoding.ReadBool(bin, index, &self.Initial) {
		return false
	}

	if !casterInt64(&self.DatacenterID) {
		return false
	}

	if !encoding.ReadBool(bin, index, &self.RttReduction) {
		return false
	}

	if !encoding.ReadBool(bin, index, &self.PacketLossReduction) {
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

	casterInt64(&self.Userhash)

	if !encoding.ReadFloat64(bin, index, &self.Latitude) {
		return false
	}

	if !encoding.ReadFloat64(bin, index, &self.Longitude) {
		return false
	}

	return encoding.ReadString(bin, index, &self.ISP, math.MaxInt32)
}

func (self *Entry) MarshalBinary() ([]byte, error) {
	size := 8 + 8 + 8 + 8 + 1 + 8 + 8 + 8 + 8 + 8 + 8 + 8 + 8*len(self.NextRelays) + 8 + 8 + 8 + 1 + 1 + 1 + 8 + 8 + 1 + 8 + 1 + 1 + 8 + 8*len(self.NextRelaysPrice) + 8 + 8 + 8 + 4 + len(self.ISP)
	bin := make([]byte, size)
	index := 0

	encoding.WriteUint64(bin, &index, uint64(self.SessionID))
	encoding.WriteUint64(bin, &index, uint64(self.Timestamp))
	encoding.WriteUint64(bin, &index, uint64(self.BuyerID))
	encoding.WriteUint64(bin, &index, uint64(self.SliceNumber))
	encoding.WriteBool(bin, &index, self.Next)
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
	encoding.WriteBool(bin, &index, self.Committed)
	encoding.WriteBool(bin, &index, self.Flagged)
	encoding.WriteBool(bin, &index, self.Multipath)
	encoding.WriteUint64(bin, &index, uint64(self.NextBytesUp))
	encoding.WriteUint64(bin, &index, uint64(self.NextBytesDown))
	encoding.WriteBool(bin, &index, self.Initial)
	encoding.WriteUint64(bin, &index, uint64(self.DatacenterID))
	encoding.WriteBool(bin, &index, self.RttReduction)
	encoding.WriteBool(bin, &index, self.PacketLossReduction)
	encoding.WriteUint64(bin, &index, uint64(int64(len(self.NextRelaysPrice))))
	for _, price := range self.NextRelaysPrice {
		encoding.WriteUint64(bin, &index, uint64(price))
	}
	encoding.WriteUint64(bin, &index, uint64(self.Userhash))
	encoding.WriteFloat64(bin, &index, self.Latitude)
	encoding.WriteFloat64(bin, &index, self.Longitude)
	encoding.WriteString(bin, &index, self.ISP, uint32(len(self.ISP)))

	return bin, nil
}

func (self *Entry) Into(data *transport.SessionPortalData, dcmap DatacenterMap, buyerID uint64) {
	var dc StrippedDatacenter
	if v, ok := dcmap[uint64(self.DatacenterID)]; ok {
		dc.Name = v.Name
	} else {
		dc.Name = "Unknown"
	}

	// meta
	{
		meta := &data.Meta
		meta.ID = uint64(self.SessionID)
		meta.UserHash = uint64(self.Userhash)
		meta.DatacenterName = dc.Name
		meta.DatacenterAlias = ""
		meta.OnNetworkNext = self.Next
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

		meta.Location = routing.Location{
			Latitude:  self.Latitude,
			Longitude: self.Longitude,
			ISP:       self.ISP,
		}

		meta.ClientAddr = "?"
		meta.ServerAddr = "?"
		meta.Hops = make([]transport.RelayHop, len(self.NextRelays))
		for i, id := range self.NextRelays {
			meta.Hops[i].ID = uint64(id)
			meta.Hops[i].Name = "?"
		}
		meta.SDK = "4.0.0"
		meta.Connection = transport.ParseConnectionType("wired")
		meta.NearbyRelays = make([]transport.NearRelayPortalData, 0)
		meta.Platform = transport.ParsePlatformType("Windows")
		meta.BuyerID = buyerID
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

		var sliceDuration int64
		if self.Initial {
			sliceDuration = 20
		} else {
			sliceDuration = 10
		}

		// multiply by 5 at the end to inflate the numbers 5x
		slice.Envelope = routing.Envelope{
			Up:   self.NextBytesUp / 1000 / sliceDuration * 8 * 5,
			Down: self.NextBytesDown / 1000 / sliceDuration * 8 * 5,
		}
		slice.OnNetworkNext = self.Next
		slice.IsMultiPath = self.Multipath
		slice.IsTryBeforeYouBuy = false
	}

	// map point
	{
		pt := &data.Point
		pt.Latitude = self.Latitude
		pt.Longitude = self.Longitude
	}
}
