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
	RouteDecision                   uint64 // Deprecated with server_backend4
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
