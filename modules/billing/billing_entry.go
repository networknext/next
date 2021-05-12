package billing

import (
	"fmt"

	"github.com/networknext/backend/modules/encoding"
)

const (
	BillingEntryVersion = uint8(27)

	BillingEntryMaxRelays           = 5
	BillingEntryMaxISPLength        = 64
	BillingEntryMaxSDKVersionLength = 11
	BillingEntryMaxDebugLength      = 2048
	BillingEntryMaxNearRelays       = 32
	BillingEntryMaxTags             = 8

	MaxBillingEntryBytes = 1 + // Version
		8 + // Timestamp
		8 + // BuyerID
		8 + // UserHash
		8 + // SessionID
		4 + // SliceNumber
		4 + // DirectRTT
		4 + // DirectJitter
		4 + // DirectPacketLoss
		1 + // Next
		4 + // NextRTT
		4 + // NextJitter
		4 + // NextPacketLoss
		1 + // NumNextRelays
		BillingEntryMaxRelays*8 + // NextRelays
		8 + // TotalPrice
		8 + // ClientToServerPacketsLost
		8 + // ServerToClientPacketsLost
		1 + // Committed
		1 + // Flagged
		1 + // Multipath
		1 + // Initial
		8 + // NextBytesUp
		8 + // NextBytesDown
		8 + // EnvelopeBytesUp
		8 + // EnvelopeBytesDown
		8 + // DatacenterID
		1 + // RTTReduction
		1 + // PacketLossReduction
		BillingEntryMaxRelays*8 + // NextRelaysPrice
		4 + // Latitude
		4 + // Longitude
		4 + BillingEntryMaxISPLength + // ISP
		1 + // ABTest
		1 + // ConnectionType
		1 + // PlatformType
		4 + BillingEntryMaxSDKVersionLength + // SDKVersion
		4 + // PacketLoss
		4 + // PredictedNextRTT
		1 + // MultipathVetoed
		1 + // UseDebug
		4 + BillingEntryMaxDebugLength + // Debug
		1 + // FallbackToDirect
		4 + // ClientFlags
		8 + // UserFlags
		4 + // NearRelayRTT
		8 + // PacketsOutOfOrderClientToServer
		8 + // PacketsOutOfOrderServerToClient
		4 + // JitterClientToServer
		4 + // JitterServerToClient
		1 + // NumNearRelays
		BillingEntryMaxNearRelays*8 + // NearRelayIDs
		BillingEntryMaxNearRelays*4 + // NearRelayRTTs
		BillingEntryMaxNearRelays*4 + // NearRelayJitters
		BillingEntryMaxNearRelays*4 + // NearRelayPacketLosses
		1 + // RelayWentAway
		1 + // RouteLost
		1 + // NumTags
		BillingEntryMaxTags*8 + // Tags
		1 + // Mispredicted
		1 + // Vetoed
		1 + // LatencyWorse
		1 + // NoRoute
		1 + // NextLatencyTooHigh
		1 + // RouteChanged
		1 + // CommitVeto
		4 + // RouteDiversity
		1 + // LackOfDiversity
		1 + // Pro
		1 + // MultipathRestricted
		8 + // ClientToServerPacketsSent
		8 + // ServerToClientPacketsSent
		1 + // UnknownDatacenter
		1 + // DatacenterNotEnabled
		1 + // BuyerNotLive
		1 // StaleRouteMatrix
)

type BillingEntry struct {
	Version                         uint8
	Timestamp                       uint64
	BuyerID                         uint64
	UserHash                        uint64
	SessionID                       uint64
	SliceNumber                     uint32
	DirectRTT                       float32
	DirectJitter                    float32
	DirectPacketLoss                float32
	Next                            bool
	NextRTT                         float32
	NextJitter                      float32
	NextPacketLoss                  float32
	NumNextRelays                   uint8
	NextRelays                      [BillingEntryMaxRelays]uint64
	TotalPrice                      uint64
	ClientToServerPacketsLost       uint64
	ServerToClientPacketsLost       uint64
	Committed                       bool
	Flagged                         bool
	Multipath                       bool
	Initial                         bool
	NextBytesUp                     uint64
	NextBytesDown                   uint64
	EnvelopeBytesUp                 uint64
	EnvelopeBytesDown               uint64
	DatacenterID                    uint64
	RTTReduction                    bool
	PacketLossReduction             bool
	NextRelaysPrice                 [BillingEntryMaxRelays]uint64
	Latitude                        float32
	Longitude                       float32
	ISP                             string
	ABTest                          bool
	RouteDecision                   uint64
	ConnectionType                  uint8
	PlatformType                    uint8
	SDKVersion                      string
	PacketLoss                      float32
	PredictedNextRTT                float32
	MultipathVetoed                 bool
	UseDebug                        bool
	Debug                           string
	FallbackToDirect                bool
	ClientFlags                     uint32
	UserFlags                       uint64
	NearRelayRTT                    float32
	PacketsOutOfOrderClientToServer uint64
	PacketsOutOfOrderServerToClient uint64
	JitterClientToServer            float32
	JitterServerToClient            float32
	NumNearRelays                   uint8
	NearRelayIDs                    [BillingEntryMaxNearRelays]uint64
	NearRelayRTTs                   [BillingEntryMaxNearRelays]float32
	NearRelayJitters                [BillingEntryMaxNearRelays]float32
	NearRelayPacketLosses           [BillingEntryMaxNearRelays]float32
	RelayWentAway                   bool
	RouteLost                       bool
	NumTags                         uint8
	Tags                            [BillingEntryMaxTags]uint64
	Mispredicted                    bool
	Vetoed                          bool
	LatencyWorse                    bool
	NoRoute                         bool
	NextLatencyTooHigh              bool
	RouteChanged                    bool
	CommitVeto                      bool
	RouteDiversity                  uint32
	LackOfDiversity                 bool
	Pro                             bool
	MultipathRestricted             bool
	ClientToServerPacketsSent       uint64
	ServerToClientPacketsSent       uint64
	UnknownDatacenter               bool
	DatacenterNotEnabled            bool
	BuyerNotLive                    bool
	StaleRouteMatrix                bool
}

func WriteBillingEntry(entry *BillingEntry) []byte {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("recovered from panic during billing entry write: %v\n", r)
		}
	}()

	data := make([]byte, MaxBillingEntryBytes)
	index := 0
	encoding.WriteUint8(data, &index, BillingEntryVersion)
	encoding.WriteUint64(data, &index, entry.Timestamp)
	encoding.WriteUint64(data, &index, entry.BuyerID)
	encoding.WriteUint64(data, &index, entry.UserHash)
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
		for i := uint8(0); i < entry.NumNextRelays; i++ {
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

	encoding.WriteBool(data, &index, entry.Initial)

	if entry.Next {
		encoding.WriteUint64(data, &index, entry.NextBytesUp)
		encoding.WriteUint64(data, &index, entry.NextBytesDown)
		encoding.WriteUint64(data, &index, entry.EnvelopeBytesUp)
		encoding.WriteUint64(data, &index, entry.EnvelopeBytesDown)
	}

	encoding.WriteUint64(data, &index, entry.DatacenterID)

	if entry.Next {
		encoding.WriteBool(data, &index, entry.RTTReduction)
		encoding.WriteBool(data, &index, entry.PacketLossReduction)

		encoding.WriteUint8(data, &index, entry.NumNextRelays)
		for i := uint8(0); i < entry.NumNextRelays; i++ {
			encoding.WriteUint64(data, &index, entry.NextRelaysPrice[i])
		}
	}

	encoding.WriteFloat32(data, &index, entry.Latitude)
	encoding.WriteFloat32(data, &index, entry.Longitude)
	encoding.WriteString(data, &index, entry.ISP, BillingEntryMaxISPLength)
	encoding.WriteBool(data, &index, entry.ABTest)

	encoding.WriteUint8(data, &index, entry.ConnectionType)
	encoding.WriteUint8(data, &index, entry.PlatformType)
	encoding.WriteString(data, &index, entry.SDKVersion, BillingEntryMaxSDKVersionLength)

	encoding.WriteFloat32(data, &index, entry.PacketLoss)

	encoding.WriteFloat32(data, &index, entry.PredictedNextRTT)

	encoding.WriteBool(data, &index, entry.MultipathVetoed)

	encoding.WriteBool(data, &index, entry.UseDebug)
	encoding.WriteString(data, &index, entry.Debug, BillingEntryMaxDebugLength)

	encoding.WriteBool(data, &index, entry.FallbackToDirect)
	encoding.WriteUint32(data, &index, entry.ClientFlags)
	encoding.WriteUint64(data, &index, entry.UserFlags)

	encoding.WriteFloat32(data, &index, entry.NearRelayRTT)

	encoding.WriteUint64(data, &index, entry.PacketsOutOfOrderClientToServer)
	encoding.WriteUint64(data, &index, entry.PacketsOutOfOrderServerToClient)
	encoding.WriteFloat32(data, &index, entry.JitterClientToServer)
	encoding.WriteFloat32(data, &index, entry.JitterServerToClient)

	encoding.WriteUint8(data, &index, entry.NumNearRelays)
	for i := uint8(0); i < entry.NumNearRelays; i++ {
		encoding.WriteUint64(data, &index, entry.NearRelayIDs[i])
		encoding.WriteFloat32(data, &index, entry.NearRelayRTTs[i])
		encoding.WriteFloat32(data, &index, entry.NearRelayJitters[i])
		encoding.WriteFloat32(data, &index, entry.NearRelayPacketLosses[i])
	}

	encoding.WriteBool(data, &index, entry.RelayWentAway)
	encoding.WriteBool(data, &index, entry.RouteLost)

	encoding.WriteUint8(data, &index, entry.NumTags)
	for i := uint8(0); i < entry.NumTags; i++ {
		encoding.WriteUint64(data, &index, entry.Tags[i])
	}

	encoding.WriteBool(data, &index, entry.Mispredicted)
	encoding.WriteBool(data, &index, entry.Vetoed)

	encoding.WriteBool(data, &index, entry.LatencyWorse)
	encoding.WriteBool(data, &index, entry.NoRoute)
	encoding.WriteBool(data, &index, entry.NextLatencyTooHigh)
	encoding.WriteBool(data, &index, entry.RouteChanged)
	encoding.WriteBool(data, &index, entry.CommitVeto)

	encoding.WriteUint32(data, &index, entry.RouteDiversity)

	encoding.WriteBool(data, &index, entry.LackOfDiversity)

	encoding.WriteBool(data, &index, entry.Pro)

	encoding.WriteBool(data, &index, entry.MultipathRestricted)

	encoding.WriteUint64(data, &index, entry.ClientToServerPacketsSent)
	encoding.WriteUint64(data, &index, entry.ServerToClientPacketsSent)

	encoding.WriteBool(data, &index, entry.UnknownDatacenter)
	encoding.WriteBool(data, &index, entry.DatacenterNotEnabled)
	encoding.WriteBool(data, &index, entry.BuyerNotLive)

	encoding.WriteBool(data, &index, entry.StaleRouteMatrix)

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

	if entry.Version >= 13 {
		if !encoding.ReadUint64(data, &index, &entry.Timestamp) {
			return false
		}
	}

	if !encoding.ReadUint64(data, &index, &entry.BuyerID) {
		return false
	}

	if entry.Version >= 6 {
		if !encoding.ReadUint64(data, &index, &entry.UserHash) {
			return false
		}
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

		if entry.Next {
			if !encoding.ReadUint64(data, &index, &entry.NextBytesUp) {
				return false
			}

			if !encoding.ReadUint64(data, &index, &entry.NextBytesDown) {
				return false
			}

			if entry.Version >= 10 {
				if !encoding.ReadUint64(data, &index, &entry.EnvelopeBytesUp) {
					return false
				}

				if !encoding.ReadUint64(data, &index, &entry.EnvelopeBytesDown) {
					return false
				}
			}
		}
	}

	if entry.Version >= 4 {
		if !encoding.ReadUint64(data, &index, &entry.DatacenterID) {
			return false
		}

		if entry.Next {
			if !encoding.ReadBool(data, &index, &entry.RTTReduction) {
				return false
			}

			if !encoding.ReadBool(data, &index, &entry.PacketLossReduction) {
				return false
			}
		}
	}

	if entry.Version >= 5 {
		if entry.Next {
			if !encoding.ReadUint8(data, &index, &entry.NumNextRelays) {
				return false
			}

			if entry.NumNextRelays > BillingEntryMaxRelays {
				return false
			}

			for i := 0; i < int(entry.NumNextRelays); i++ {
				if !encoding.ReadUint64(data, &index, &entry.NextRelaysPrice[i]) {
					return false
				}
			}
		}
	}

	if entry.Version >= 7 {
		if !encoding.ReadFloat32(data, &index, &entry.Latitude) {
			return false
		}

		if !encoding.ReadFloat32(data, &index, &entry.Longitude) {
			return false
		}

		if !encoding.ReadString(data, &index, &entry.ISP, BillingEntryMaxISPLength) {
			return false
		}

		if !encoding.ReadBool(data, &index, &entry.ABTest) {
			return false
		}

		if entry.Version < 14 {
			if !encoding.ReadUint64(data, &index, &entry.RouteDecision) {
				return false
			}
		}
	}

	if entry.Version >= 8 {
		if !encoding.ReadUint8(data, &index, &entry.ConnectionType) {
			return false
		}

		if !encoding.ReadUint8(data, &index, &entry.PlatformType) {
			return false
		}

		if !encoding.ReadString(data, &index, &entry.SDKVersion, BillingEntryMaxSDKVersionLength) {
			return false
		}
	}

	if entry.Version >= 9 {
		if !encoding.ReadFloat32(data, &index, &entry.PacketLoss) {
			return false
		}
	}

	if entry.Version >= 11 {
		if !encoding.ReadFloat32(data, &index, &entry.PredictedNextRTT) {
			return false
		}
	}

	if entry.Version >= 12 {
		if !encoding.ReadBool(data, &index, &entry.MultipathVetoed) {
			return false
		}
	}

	if entry.Version >= 14 {
		if entry.Version >= 17 {
			if !encoding.ReadBool(data, &index, &entry.UseDebug) {
				return false
			}
		}

		if !encoding.ReadString(data, &index, &entry.Debug, BillingEntryMaxDebugLength) {
			return false
		}
	}

	if entry.Version >= 15 {
		if !encoding.ReadBool(data, &index, &entry.FallbackToDirect) {
			return false
		}

		if !encoding.ReadUint32(data, &index, &entry.ClientFlags) {
			return false
		}

		if !encoding.ReadUint64(data, &index, &entry.UserFlags) {
			return false
		}
	}

	if entry.Version >= 16 {
		if !encoding.ReadFloat32(data, &index, &entry.NearRelayRTT) {
			return false
		}
	}

	if entry.Version >= 17 {
		if !encoding.ReadUint64(data, &index, &entry.PacketsOutOfOrderClientToServer) {
			return false
		}

		if !encoding.ReadUint64(data, &index, &entry.PacketsOutOfOrderServerToClient) {
			return false
		}

		if !encoding.ReadFloat32(data, &index, &entry.JitterClientToServer) {
			return false
		}

		if !encoding.ReadFloat32(data, &index, &entry.JitterServerToClient) {
			return false
		}

		if !encoding.ReadUint8(data, &index, &entry.NumNearRelays) {
			return false
		}

		var numNearRelays int
		if entry.Version >= 18 {
			numNearRelays = int(entry.NumNearRelays)
		} else {
			numNearRelays = BillingEntryMaxNearRelays
		}

		for i := 0; i < numNearRelays; i++ {
			if !encoding.ReadUint64(data, &index, &entry.NearRelayIDs[i]) {
				return false
			}

			if !encoding.ReadFloat32(data, &index, &entry.NearRelayRTTs[i]) {
				return false
			}

			if !encoding.ReadFloat32(data, &index, &entry.NearRelayJitters[i]) {
				return false
			}

			if !encoding.ReadFloat32(data, &index, &entry.NearRelayPacketLosses[i]) {
				return false
			}
		}
	}

	if entry.Version >= 18 {
		if !encoding.ReadBool(data, &index, &entry.RelayWentAway) {
			return false
		}

		if !encoding.ReadBool(data, &index, &entry.RouteLost) {
			return false
		}

		if !encoding.ReadUint8(data, &index, &entry.NumTags) {
			return false
		}

		for i := uint8(0); i < entry.NumTags; i++ {
			if !encoding.ReadUint64(data, &index, &entry.Tags[i]) {
				return false
			}
		}
	}

	if entry.Version >= 19 {
		if !encoding.ReadBool(data, &index, &entry.Mispredicted) {
			return false
		}

		if !encoding.ReadBool(data, &index, &entry.Vetoed) {
			return false
		}
	}

	if entry.Version >= 20 {
		if !encoding.ReadBool(data, &index, &entry.LatencyWorse) {
			return false
		}

		if !encoding.ReadBool(data, &index, &entry.NoRoute) {
			return false
		}

		if !encoding.ReadBool(data, &index, &entry.NextLatencyTooHigh) {
			return false
		}

		if !encoding.ReadBool(data, &index, &entry.RouteChanged) {
			return false
		}

		if !encoding.ReadBool(data, &index, &entry.CommitVeto) {
			return false
		}
	}

	if entry.Version >= 21 {
		if !encoding.ReadUint32(data, &index, &entry.RouteDiversity) {
			return false
		}
	}

	if entry.Version >= 22 {
		if !encoding.ReadBool(data, &index, &entry.LackOfDiversity) {
			return false
		}
	}

	if entry.Version >= 23 {
		if !encoding.ReadBool(data, &index, &entry.Pro) {
			return false
		}
	}

	if entry.Version >= 24 {
		if !encoding.ReadBool(data, &index, &entry.MultipathRestricted) {
			return false
		}
	}

	if entry.Version >= 25 {
		if !encoding.ReadUint64(data, &index, &entry.ClientToServerPacketsSent) {
			return false
		}

		if !encoding.ReadUint64(data, &index, &entry.ServerToClientPacketsSent) {
			return false
		}
	}

	if entry.Version >= 26 {

		if !encoding.ReadBool(data, &index, &entry.UnknownDatacenter) {
			return false
		}

		if !encoding.ReadBool(data, &index, &entry.DatacenterNotEnabled) {
			return false
		}

		if !encoding.ReadBool(data, &index, &entry.BuyerNotLive) {
			return false
		}

	}

	if entry.Version >= 27 {
		if !encoding.ReadBool(data, &index, &entry.StaleRouteMatrix) {
			return false
		}
	}

	return true
}

// ------------------------------------------------------------------------

const BillingEntryVersion2 = uint8(0)

type BillingEntry2 struct {

	// always

	Version             uint32
	Timestamp           uint32
	SessionID           uint64
	SliceNumber         uint32
	DirectRTT           int32
	DirectJitter        int32
	DirectPacketLoss    int32
	RealPacketLoss      int32
	RealPacketLoss_Frac uint32
	RealJitter          uint32
	Next                bool
	Flagged             bool
	Summary             bool
	UseDebug            bool
	Debug               string

	// first slice only

	DatacenterID      uint64
	BuyerID           uint64
	UserHash          uint64
	EnvelopeBytesUp   uint64
	EnvelopeBytesDown uint64
	Latitude          float32
	Longitude         float32
	ISP               string
	ConnectionType    int32
	PlatformType      int32
	SDKVersion        string
	NumTags           int32
	Tags              [BillingEntryMaxTags]uint64
	ABTest            bool
	Pro               bool

	// summary slice only

	ClientToServerPacketsSent       uint64
	ServerToClientPacketsSent       uint64
	ClientToServerPacketsLost       uint64
	ServerToClientPacketsLost       uint64
	ClientToServerPacketsOutOfOrder uint64
	ServerToClientPacketsOutOfOrder uint64
	NumNearRelays                   int32
	NearRelayIDs                    [BillingEntryMaxNearRelays]uint64
	NearRelayRTTs                   [BillingEntryMaxNearRelays]int32
	NearRelayJitters                [BillingEntryMaxNearRelays]int32
	NearRelayPacketLosses           [BillingEntryMaxNearRelays]int32

	// network next only

	NextRTT             int32
	NextJitter          int32
	NextPacketLoss      int32
	PredictedNextRTT    int32
	NearRelayRTT        int32
	NumNextRelays       int32
	NextRelays          [BillingEntryMaxRelays]uint64
	NextRelayPrice      [BillingEntryMaxRelays]uint64
	TotalPrice          uint64
	RouteDiversity      int32
	Committed           bool
	Multipath           bool
	RTTReduction        bool
	PacketLossReduction bool

	// error state only

	FallbackToDirect     bool
	MultipathVetoed      bool
	Mispredicted         bool
	Vetoed               bool
	LatencyWorse         bool
	NoRoute              bool
	NextLatencyTooHigh   bool
	RouteChanged         bool
	CommitVeto           bool
	LackOfDiversity      bool
	MultipathRestricted  bool
	UnknownDatacenter    bool
	DatacenterNotEnabled bool
	BuyerNotLive         bool
	StaleRouteMatrix     bool
}

func (entry *BillingEntry2) Serialize(stream encoding.Stream) error {

	/*
		1. Always

		These values are serialized in every slice
	*/

	stream.SerializeBits(&entry.Version, 8)
	stream.SerializeBits(&entry.Timestamp, 32)
	stream.SerializeUint64(&entry.SessionID)

	// // todo: serialize intelligently, eg. up to 1hr, up to 16 bits, 32 bits worst case only.
	stream.SerializeBits(&entry.SliceNumber, 32)

	stream.SerializeInteger(&entry.DirectRTT, 0, 1023)
	stream.SerializeInteger(&entry.DirectJitter, 0, 255)
	stream.SerializeInteger(&entry.DirectPacketLoss, 0, 100)

	stream.SerializeInteger(&entry.RealPacketLoss, 0, 100)
	stream.SerializeBits(&entry.RealPacketLoss_Frac, 8)
	stream.SerializeBits(&entry.RealJitter, 8)

	stream.SerializeBool(&entry.Next)
	stream.SerializeBool(&entry.Flagged)
	stream.SerializeBool(&entry.Summary)

	stream.SerializeBool(&entry.UseDebug)
	stream.SerializeString(&entry.Debug, BillingEntryMaxDebugLength)

	// todo: Summary

	/*
		2. First slice only

		These values are serialized only for slice 0.
	*/

	if entry.SliceNumber == 0 {

		stream.SerializeUint64(&entry.DatacenterID)
		stream.SerializeUint64(&entry.BuyerID)
		stream.SerializeUint64(&entry.UserHash)
		stream.SerializeUint64(&entry.EnvelopeBytesUp)
		stream.SerializeUint64(&entry.EnvelopeBytesDown)
		stream.SerializeFloat32(&entry.Latitude)
		stream.SerializeFloat32(&entry.Longitude)
		stream.SerializeString(&entry.ISP, BillingEntryMaxISPLength)
		stream.SerializeInteger(&entry.ConnectionType, 0, 3) // todo: constant
		stream.SerializeInteger(&entry.PlatformType, 0, 10)  // todo: constant
		stream.SerializeString(&entry.SDKVersion, 10)        // todo: constant
		stream.SerializeInteger(&entry.NumTags, 0, BillingEntryMaxTags)
		for i := 0; i < int(entry.NumTags); i++ {
			stream.SerializeUint64(&entry.Tags[i])
		}
		stream.SerializeBool(&entry.ABTest)
		stream.SerializeBool(&entry.Pro)

	}

	/*
		3. Summary slice only

		These values are serialized only for the summary slice (at the end of the session)
	*/

	if entry.Summary {

		stream.SerializeUint64(&entry.ClientToServerPacketsSent)
		stream.SerializeUint64(&entry.ServerToClientPacketsSent)
		stream.SerializeUint64(&entry.ClientToServerPacketsLost)
		stream.SerializeUint64(&entry.ServerToClientPacketsLost)
		stream.SerializeUint64(&entry.ClientToServerPacketsOutOfOrder)
		stream.SerializeUint64(&entry.ServerToClientPacketsOutOfOrder)
		stream.SerializeInteger(&entry.NumNearRelays, 0, BillingEntryMaxNearRelays)
		for i := 0; i < int(entry.NumNearRelays); i++ {
			stream.SerializeUint64(&entry.NearRelayIDs[i])
			stream.SerializeInteger(&entry.NearRelayRTTs[i], 0, 255)
			stream.SerializeInteger(&entry.NearRelayJitters[i], 0, 255)
			stream.SerializeInteger(&entry.NearRelayPacketLosses[i], 0, 100)
		}

	}

	/*
		4. Network Next Only

		These values are serialized only when a slice is on network next.
	*/

	if entry.Next {

		stream.SerializeInteger(&entry.NextRTT, 0, 255)
		stream.SerializeInteger(&entry.NextJitter, 0, 255)
		stream.SerializeInteger(&entry.NextPacketLoss, 0, 100)
		stream.SerializeInteger(&entry.PredictedNextRTT, 0, 255)
		stream.SerializeInteger(&entry.NearRelayRTT, 0, 255)

		stream.SerializeInteger(&entry.NumNextRelays, 0, BillingEntryMaxRelays)
		for i := 0; i < int(entry.NumNextRelays); i++ {
			stream.SerializeUint64(&entry.NextRelays[i])
			stream.SerializeUint64(&entry.NextRelayPrice[i])
		}

		stream.SerializeUint64(&entry.TotalPrice)
		stream.SerializeInteger(&entry.RouteDiversity, 1, 31)
		stream.SerializeBool(&entry.Committed)
		stream.SerializeBool(&entry.Multipath)
		stream.SerializeBool(&entry.RTTReduction)
		stream.SerializeBool(&entry.PacketLossReduction)
	}

	/*
		5. Error State Only

		These values are only serialized when the session is in an error state (rare).
	*/

	errorState := false

	if stream.IsWriting() {
		errorState =
			entry.FallbackToDirect ||
				entry.MultipathVetoed ||
				entry.Mispredicted ||
				entry.Vetoed ||
				entry.LatencyWorse ||
				entry.NoRoute ||
				entry.NextLatencyTooHigh ||
				entry.RouteChanged ||
				entry.CommitVeto ||
				entry.LackOfDiversity ||
				entry.MultipathRestricted ||
				entry.UnknownDatacenter ||
				entry.DatacenterNotEnabled ||
				entry.BuyerNotLive ||
				entry.StaleRouteMatrix
	}

	stream.SerializeBool(&errorState)

	if errorState {

		stream.SerializeBool(&entry.FallbackToDirect)
		stream.SerializeBool(&entry.MultipathVetoed)
		stream.SerializeBool(&entry.Mispredicted)
		stream.SerializeBool(&entry.Vetoed)
		stream.SerializeBool(&entry.LatencyWorse)
		stream.SerializeBool(&entry.NoRoute)
		stream.SerializeBool(&entry.NextLatencyTooHigh)
		stream.SerializeBool(&entry.RouteChanged)
		stream.SerializeBool(&entry.CommitVeto)
		stream.SerializeBool(&entry.LackOfDiversity)
		stream.SerializeBool(&entry.MultipathRestricted)
		stream.SerializeBool(&entry.UnknownDatacenter)
		stream.SerializeBool(&entry.DatacenterNotEnabled)
		stream.SerializeBool(&entry.BuyerNotLive)
		stream.SerializeBool(&entry.StaleRouteMatrix)

	}

	return stream.Error()
}
