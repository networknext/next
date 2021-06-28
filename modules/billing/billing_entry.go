package billing

import (
	"fmt"
	"math"

	"cloud.google.com/go/bigquery"

	"github.com/networknext/backend/modules/core"
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

// Validate a billing entry. Returns true if the billing entry is valid, false if invalid.
func (entry *BillingEntry) Validate() bool {

	if entry.Timestamp < 0 {
		fmt.Printf("invalid timestamp\n")
		return false
	}

	if entry.BuyerID == 0 {
		fmt.Printf("invalid buyer id\n")
		return false
	}

	if entry.SessionID == 0 {
		fmt.Printf("invalid session id\n")
		return false
	}

	// IMPORTANT: Logic inverted because comparing a NaN float value always returns false
	if !(entry.Latitude >= -90.0 && entry.Latitude <= +90.0) {
		fmt.Printf("invalid latitude\n")
		return false
	}

	if !(entry.Longitude >= -180.0 && entry.Longitude <= +180.0) {
		fmt.Printf("invalid longitude\n")
		return false
	}

	// IMPORTANT: You must update this check if you ever add a new connection type in the SDK
	if entry.ConnectionType < 0 || entry.ConnectionType > 3 {
		fmt.Printf("invalid connection type\n")
		return false
	}

	// IMPORTANT: You must update this check when you add new platforms to the SDK
	if entry.PlatformType < 0 || entry.PlatformType > 10 {
		fmt.Printf("invalid platform type\n")
		return false
	}

	if entry.RouteDiversity < 0 || entry.RouteDiversity > 32 {
		fmt.Printf("invalid route diversity\n")
		return false
	}

	if !(entry.DirectRTT >= 0.0 && entry.DirectRTT <= 10000.0) {
		fmt.Printf("invalid direct rtt\n")
		return false
	}

	if !(entry.DirectJitter >= 0.0 && entry.DirectJitter <= 10000.0) {
		fmt.Printf("invalid direct jitter\n")
		return false
	}

	if !(entry.DirectPacketLoss >= 0.0 && entry.DirectPacketLoss <= 100.0) {
		fmt.Printf("invalid direct packet loss\n")
		return false
	}

	if !(entry.PacketLoss >= 0.0 && entry.PacketLoss <= 100.0) {
		if entry.PacketLoss > 100.0 {
			fmt.Printf("PacketLoss %v > 100.0. Clamping to 100.0\n%+v\n", entry.PacketLoss, entry)
			entry.PacketLoss = 100.0
		} else {
			fmt.Printf("invalid packet loss\n")
			return false
		}
	}

	if !(entry.JitterClientToServer >= 0.0 && entry.JitterClientToServer <= 1000.0) {
		if entry.JitterClientToServer > 1000.0 {
			fmt.Printf("JitterClientToServer %v > 1000.0. Clamping to 1000.0\n%+v\n", entry.JitterClientToServer, entry)
			entry.JitterClientToServer = 1000.0
		} else {
			fmt.Printf("invalid jitter client to server\n")
			return false
		}
	}

	if !(entry.JitterServerToClient >= 0.0 && entry.JitterServerToClient <= 1000.0) {
		if entry.JitterServerToClient > 1000.0 {
			fmt.Printf("JitterServerToClient %v > 1000.0. Clamping to 1000.0.\n%+v\n", entry.JitterServerToClient, entry)
			entry.JitterServerToClient = 1000.0
		} else {
			fmt.Printf("invalid jitter server to client\n")
			return false
		}
	}

	if entry.NumTags < 0 || entry.NumTags > 8 {
		fmt.Printf("invalid num tags\n")
		return false
	}

	if entry.Next {

		if !(entry.NextRTT >= 0.0 && entry.NextRTT <= 10000.0) {
			fmt.Printf("invalid next rtt\n")
			return false
		}

		if !(entry.NextJitter >= 0.0 && entry.NextJitter <= 10000.0) {
			fmt.Printf("invalid next jitter\n")
			return false
		}

		if !(entry.NextPacketLoss >= 0.0 && entry.NextPacketLoss <= 100.0) {
			fmt.Printf("invalid next packet loss\n")
			return false
		}

		if !(entry.PredictedNextRTT >= 0.0 && entry.PredictedNextRTT <= 10000.0) {
			fmt.Printf("invalid predicted next rtt\n")
			return false
		}

		if entry.NumNearRelays < 0 || entry.NumNearRelays > 32 {
			fmt.Printf("invalid num near relays\n")
			return false
		}

		if entry.NumNearRelays > 0 {

			for i := 0; i < int(entry.NumNearRelays); i++ {

				if entry.NearRelayIDs[i] == 0 {
					// Log this but do not return false
					// TODO: investigate why nearRelayID is 0
					fmt.Printf("NearRelayIDs[%d] is 0.\n%+v\n", i, entry)
				}

				if !(entry.NearRelayRTTs[i] >= 0.0 && entry.NearRelayRTTs[i] <= 255.0) {
					fmt.Printf("invalid near relay rtt\n")
					return false
				}

				if !(entry.NearRelayJitters[i] >= 0.0 && entry.NearRelayJitters[i] <= 255.0) {
					fmt.Printf("invalid near relay jitter\n")
					return false
				}

				if !(entry.NearRelayPacketLosses[i] >= 0.0 && entry.NearRelayPacketLosses[i] <= 100.0) {
					fmt.Printf("invalid near relay packet loss\n")
					return false
				}
			}
		}
	}

	return true
}

// Checks all floating point numbers in the BillingEntry for NaN and +-Inf and forces them to 0
func (entry *BillingEntry) CheckNaNOrInf() (bool, []string) {
	var nanOrInfExists bool
	var nanOrInfFields []string

	if math.IsNaN(float64(entry.NextRTT)) || math.IsInf(float64(entry.NextRTT), 0) {
		nanOrInfFields = append(nanOrInfFields, "NextRTT")
		nanOrInfExists = true
		entry.NextRTT = float32(0)
	}

	if math.IsNaN(float64(entry.NextJitter)) || math.IsInf(float64(entry.NextJitter), 0) {
		nanOrInfFields = append(nanOrInfFields, "NextJitter")
		nanOrInfExists = true
		entry.NextJitter = float32(0)
	}

	if math.IsNaN(float64(entry.NextPacketLoss)) || math.IsInf(float64(entry.NextPacketLoss), 0) {
		nanOrInfFields = append(nanOrInfFields, "NextPacketLoss")
		nanOrInfExists = true
		entry.NextPacketLoss = float32(0)
	}

	if math.IsNaN(float64(entry.Latitude)) || math.IsInf(float64(entry.Latitude), 0) {
		nanOrInfFields = append(nanOrInfFields, "Latitude")
		nanOrInfExists = true
		entry.Latitude = float32(0)
	}

	if math.IsNaN(float64(entry.Longitude)) || math.IsInf(float64(entry.Longitude), 0) {
		nanOrInfFields = append(nanOrInfFields, "Longitude")
		nanOrInfExists = true
		entry.Longitude = float32(0)
	}

	if math.IsNaN(float64(entry.PacketLoss)) || math.IsInf(float64(entry.PacketLoss), 0) {
		nanOrInfFields = append(nanOrInfFields, "PacketLoss")
		nanOrInfExists = true
		entry.PacketLoss = float32(0)
	}

	if math.IsNaN(float64(entry.PredictedNextRTT)) || math.IsInf(float64(entry.PredictedNextRTT), 0) {
		nanOrInfFields = append(nanOrInfFields, "PredictedNextRTT")
		nanOrInfExists = true
		entry.PredictedNextRTT = float32(0)
	}

	if math.IsNaN(float64(entry.NearRelayRTT)) || math.IsInf(float64(entry.NearRelayRTT), 0) {
		nanOrInfFields = append(nanOrInfFields, "NearRelayRTT")
		nanOrInfExists = true
		entry.NearRelayRTT = float32(0)
	}

	if math.IsNaN(float64(entry.JitterClientToServer)) || math.IsInf(float64(entry.JitterClientToServer), 0) {
		nanOrInfFields = append(nanOrInfFields, "JitterClientToServer")
		nanOrInfExists = true
		entry.JitterClientToServer = float32(0)
	}

	if math.IsNaN(float64(entry.JitterServerToClient)) || math.IsInf(float64(entry.JitterServerToClient), 0) {
		nanOrInfFields = append(nanOrInfFields, "JitterServerToClient")
		nanOrInfExists = true
		entry.JitterServerToClient = float32(0)
	}

	if entry.NumNearRelays > 0 {
		for i := 0; i < int(entry.NumNearRelays); i++ {
			if math.IsNaN(float64(entry.NearRelayRTTs[i])) || math.IsInf(float64(entry.NearRelayRTTs[i]), 0) {
				nanOrInfFields = append(nanOrInfFields, fmt.Sprintf("NearRelayRTTs[%d]", i))
				nanOrInfExists = true
				entry.NearRelayRTTs[i] = float32(0)
			}

			if math.IsNaN(float64(entry.NearRelayJitters[i])) || math.IsInf(float64(entry.NearRelayJitters[i]), 0) {
				nanOrInfFields = append(nanOrInfFields, fmt.Sprintf("NearRelayJitters[%d]", i))
				nanOrInfExists = true
				entry.NearRelayJitters[i] = float32(0)
			}

			if math.IsNaN(float64(entry.NearRelayPacketLosses[i])) || math.IsInf(float64(entry.NearRelayPacketLosses[i]), 0) {
				nanOrInfFields = append(nanOrInfFields, fmt.Sprintf("NearRelayPacketLosses[%d]", i))
				nanOrInfExists = true
				entry.NearRelayPacketLosses[i] = float32(0)
			}
		}
	}

	return nanOrInfExists, nanOrInfFields
}

// Implements the bigquery.ValueSaver interface for a billing entry so it can be used in Put()
func (entry *BillingEntry) Save() (map[string]bigquery.Value, string, error) {

	e := make(map[string]bigquery.Value)

	e["timestamp"] = int(entry.Timestamp)

	e["buyerId"] = int(entry.BuyerID)
	e["sessionId"] = int(entry.SessionID)
	e["datacenterID"] = int(entry.DatacenterID)
	e["userHash"] = int(entry.UserHash)
	e["latitude"] = entry.Latitude
	e["longitude"] = entry.Longitude
	e["isp"] = entry.ISP
	e["connectionType"] = int(entry.ConnectionType)
	e["platformType"] = int(entry.PlatformType)
	e["sdkVersion"] = entry.SDKVersion

	e["sliceNumber"] = int(entry.SliceNumber)

	e["flagged"] = entry.Flagged
	e["fallbackToDirect"] = entry.FallbackToDirect
	e["multipathVetoed"] = entry.MultipathVetoed
	e["abTest"] = entry.ABTest
	e["committed"] = entry.Committed
	e["multipath"] = entry.Multipath
	e["rttReduction"] = entry.RTTReduction
	e["packetLossReduction"] = entry.PacketLossReduction
	e["relayWentAway"] = entry.RelayWentAway
	e["routeLost"] = entry.RouteLost
	e["mispredicted"] = entry.Mispredicted
	e["vetoed"] = entry.Vetoed
	e["latencyWorse"] = entry.LatencyWorse
	e["noRoute"] = entry.NoRoute
	e["nextLatencyTooHigh"] = entry.NextLatencyTooHigh
	e["routeChanged"] = entry.RouteChanged
	e["commitVeto"] = entry.CommitVeto

	if entry.Pro {
		e["pro"] = entry.Pro
	}

	if entry.BuyerNotLive {
		e["buyerNotLive"] = entry.BuyerNotLive
	}

	if entry.UnknownDatacenter {
		e["unknownDatacenter"] = entry.UnknownDatacenter
	}

	if entry.DatacenterNotEnabled {
		e["datacenterNotEnabled"] = entry.DatacenterNotEnabled
	}

	if entry.StaleRouteMatrix {
		e["staleRouteMatrix"] = entry.StaleRouteMatrix
	}

	if entry.RouteDiversity > 0 {
		e["routeDiversity"] = entry.RouteDiversity
	}

	if entry.LackOfDiversity {
		e["lackOfDiversity"] = entry.LackOfDiversity
	}

	if entry.MultipathRestricted {
		e["multipathRestricted"] = entry.MultipathRestricted
	}

	e["directRTT"] = entry.DirectRTT
	e["directJitter"] = entry.DirectJitter
	e["directPacketLoss"] = entry.DirectPacketLoss

	if entry.ClientToServerPacketsSent > 0 {
		e["clientToServerPacketsSent"] = int(entry.ClientToServerPacketsSent)
	}

	if entry.ServerToClientPacketsSent > 0 {
		e["serverToClientPacketsSent"] = int(entry.ServerToClientPacketsSent)
	}

	if entry.ClientToServerPacketsLost > 0 {
		e["clientToServerPacketsLost"] = int(entry.ClientToServerPacketsLost)
	}

	if entry.ServerToClientPacketsLost > 0 {
		e["serverToClientPacketsLost"] = int(entry.ServerToClientPacketsLost)
	}

	// IMPORTANT: This is derived from *PacketsSent and *PacketsLost, and is valid for both next and direct
	if entry.PacketLoss > 0.0 {
		e["packetLoss"] = entry.PacketLoss
	}

	if entry.PacketsOutOfOrderClientToServer != 0 {
		e["packetsOutOfOrderClientToServer"] = int(entry.PacketsOutOfOrderClientToServer)
	}

	if entry.PacketsOutOfOrderServerToClient != 0 {
		e["packetsOutOfOrderServerToClient"] = int(entry.PacketsOutOfOrderServerToClient)
	}

	if entry.JitterClientToServer != 0 {
		e["jitterClientToServer"] = entry.JitterClientToServer
	}

	if entry.JitterServerToClient != 0 {
		e["jitterServerToClient"] = entry.JitterServerToClient
	}

	if entry.UseDebug && entry.Debug != "" {
		e["debug"] = entry.Debug
	}

	if entry.ClientFlags != 0 {
		e["clientFlags"] = int(entry.ClientFlags)
	}

	if entry.UserFlags != 0 {
		e["userFlags"] = int(entry.UserFlags)
	}

	if entry.NumTags > 0 {
		tags := make([]bigquery.Value, entry.NumTags)
		for i := 0; i < int(entry.NumTags); i++ {
			tags[i] = int(entry.Tags[i])
		}
		e["tags"] = tags
	}

	if entry.Next {

		e["next"] = entry.Next

		e["nextRTT"] = entry.NextRTT
		e["nextJitter"] = entry.NextJitter
		e["nextPacketLoss"] = entry.NextPacketLoss

		e["totalPrice"] = int(entry.TotalPrice)

		e["nextBytesUp"] = int(entry.NextBytesUp)
		e["nextBytesDown"] = int(entry.NextBytesDown)
		e["envelopeBytesUp"] = int(entry.EnvelopeBytesUp)
		e["envelopeBytesDown"] = int(entry.EnvelopeBytesDown)

		if entry.PredictedNextRTT > 0.0 {
			e["predictedNextRTT"] = entry.PredictedNextRTT
		}

		if entry.NearRelayRTT != 0 {
			e["nearRelayRTT"] = entry.NearRelayRTT
		}

		nextRelays := make([]bigquery.Value, entry.NumNextRelays)
		nextRelaysPrice := make([]bigquery.Value, entry.NumNextRelays)

		for i := 0; i < int(entry.NumNextRelays); i++ {
			nextRelays[i] = int(entry.NextRelays[i])
			nextRelaysPrice[i] = int(entry.NextRelaysPrice[i])
		}

		e["nextRelays"] = nextRelays
		e["nextRelaysPrice"] = nextRelaysPrice

		if entry.NumNearRelays > 0 && entry.UseDebug {
			// IMPORTANT: Only write this data if debug is on because it is very large

			e["numNearRelays"] = int(entry.NumNearRelays)

			nearRelayIDs := make([]bigquery.Value, entry.NumNearRelays)
			nearRelayRTTs := make([]bigquery.Value, entry.NumNearRelays)
			nearRelayJitters := make([]bigquery.Value, entry.NumNearRelays)
			nearRelayPacketLosses := make([]bigquery.Value, entry.NumNearRelays)

			for i := 0; i < int(entry.NumNearRelays); i++ {
				nearRelayIDs[i] = int(entry.NearRelayIDs[i])
				nearRelayRTTs[i] = entry.NearRelayRTTs[i]
				nearRelayJitters[i] = entry.NearRelayJitters[i]
				nearRelayPacketLosses[i] = entry.NearRelayPacketLosses[i]
			}

			e["nearRelayIDs"] = nearRelayIDs
			e["nearRelayRTTs"] = nearRelayRTTs
			e["nearRelayJitters"] = nearRelayJitters
			e["nearRelayPacketLosses"] = nearRelayPacketLosses
		}
	}

	// todo: this is deprecated. we don't really need this anymore. should
	// be made nullable in the schema and we just stop writing this.
	e["initial"] = entry.Initial

	// todo: this is deprecated and should be made nullable in the schema
	// at this point we should stop writing this
	e["routeDecision"] = int(entry.RouteDecision)

	return e, "", nil
}

// ------------------------------------------------------------------------

const (
	BillingEntryVersion2 = uint32(1)

	MaxBillingEntry2Bytes = 4096
)

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
	RouteDiversity      int32

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
	Uncommitted         bool
	Multipath           bool
	RTTReduction        bool
	PacketLossReduction bool
	RouteChanged        bool
	NextBytesUp         uint64
	NextBytesDown       uint64

	// error state only

	FallbackToDirect     bool
	MultipathVetoed      bool
	Mispredicted         bool
	Vetoed               bool
	LatencyWorse         bool
	NoRoute              bool
	NextLatencyTooHigh   bool
	CommitVeto           bool
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

	stream.SerializeBits(&entry.Version, 32)
	stream.SerializeBits(&entry.Timestamp, 32)
	stream.SerializeUint64(&entry.SessionID)

	small := false
	if entry.SliceNumber < 1024 {
		small = true
	}
	stream.SerializeBool(&small)
	if small {
		stream.SerializeBits(&entry.SliceNumber, 10)
	} else {
		stream.SerializeBits(&entry.SliceNumber, 32)
	}

	stream.SerializeInteger(&entry.DirectRTT, 0, 1023)
	stream.SerializeInteger(&entry.DirectJitter, 0, 255)
	stream.SerializeInteger(&entry.DirectPacketLoss, 0, 100)

	stream.SerializeInteger(&entry.RealPacketLoss, 0, 100)
	stream.SerializeBits(&entry.RealPacketLoss_Frac, 8)
	stream.SerializeUint32(&entry.RealJitter)

	stream.SerializeBool(&entry.Next)
	stream.SerializeBool(&entry.Flagged)
	stream.SerializeBool(&entry.Summary)

	stream.SerializeBool(&entry.UseDebug)
	stream.SerializeString(&entry.Debug, BillingEntryMaxDebugLength)

	stream.SerializeInteger(&entry.RouteDiversity, 0, 32)

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
		stream.SerializeString(&entry.SDKVersion, BillingEntryMaxSDKVersionLength)
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
		stream.SerializeBool(&entry.Uncommitted)
		stream.SerializeBool(&entry.Multipath)
		stream.SerializeBool(&entry.RTTReduction)
		stream.SerializeBool(&entry.PacketLossReduction)
		stream.SerializeBool(&entry.RouteChanged)
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
				entry.CommitVeto ||
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
		stream.SerializeBool(&entry.CommitVeto)
		stream.SerializeBool(&entry.UnknownDatacenter)
		stream.SerializeBool(&entry.DatacenterNotEnabled)
		stream.SerializeBool(&entry.BuyerNotLive)
		stream.SerializeBool(&entry.StaleRouteMatrix)

	}

	/*
		Version 1

		Include Next Bytes Up/Down for slices on next.
	*/
	if entry.Version >= uint32(1) {

		if entry.Next {

			stream.SerializeUint64(&entry.NextBytesUp)
			stream.SerializeUint64(&entry.NextBytesDown)

		}
	}

	return stream.Error()
}

func WriteBillingEntry2(entry *BillingEntry2) ([]byte, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("recovered from panic during BillingEntry2 packet entry write: %v\n", r)
		}
	}()

	buffer := [MaxBillingEntry2Bytes]byte{}

	ws, err := encoding.CreateWriteStream(buffer[:])
	if err != nil {
		return nil, err
	}

	if err := entry.Serialize(ws); err != nil {
		return nil, err
	}
	ws.Flush()

	return buffer[:ws.GetBytesProcessed()], nil
}

func ReadBillingEntry2(entry *BillingEntry2, data []byte) error {
	if err := entry.Serialize(encoding.CreateReadStream(data)); err != nil {
		return err
	}
	return nil
}

// Validate a billing entry 2. Returns true if the billing entry 2 is valid, false if invalid.
func (entry *BillingEntry2) Validate() bool {

	// always

	if entry.Version < 0 {
		fmt.Printf("invalid version\n")
		return false
	}

	if entry.Timestamp < 0 {
		fmt.Printf("invalid timestamp\n")
		return false
	}

	if entry.SessionID == 0 {
		fmt.Printf("invalid session id\n")
		return false
	}

	if entry.SliceNumber < 0 {
		fmt.Printf("invalid slice number\n")
		return false
	}

	if entry.DirectRTT < 0 || entry.DirectRTT > 1023 {
		fmt.Printf("invalid direct rtt\n")
		return false
	}

	if entry.DirectJitter < 0 || entry.DirectJitter > 255 {
		fmt.Printf("invalid direct jitter\n")
		return false
	}

	if entry.DirectPacketLoss < 0 || entry.DirectPacketLoss > 100 {
		fmt.Printf("invalid direct packet loss\n")
		return false
	}

	if entry.RealPacketLoss < 0 || entry.RealPacketLoss > 100 {
		if entry.RealPacketLoss > 100 {
			fmt.Printf("RealPacketLoss %v > 100. Clamping to 100\n%+v\n", entry.RealPacketLoss, entry)
			entry.RealPacketLoss = 100
		} else {
			fmt.Printf("invalid real packet loss\n")
			return false
		}
	}

	if entry.RealJitter < 0 || entry.RealJitter > 1000 {
		if entry.RealJitter > 1000 {
			fmt.Printf("RealJitter %v > 1000. Clamping to 1000\n%+v\n", entry.RealJitter, entry)
			entry.RealJitter = 100
		} else {
			fmt.Printf("invalid real jitter\n")
			return false
		}
	}

	if entry.RouteDiversity < 0 || entry.RouteDiversity > 32 {
		fmt.Printf("invalid route diversity\n")
		return false
	}

	// first slice only

	if entry.SliceNumber == 0 {

		if entry.BuyerID == 0 {
			fmt.Printf("invalid buyer id\n")
			return false
		}

		// IMPORTANT: Logic inverted because comparing a NaN float value always returns false
		if !(entry.Latitude >= -90.0 && entry.Latitude <= +90.0) {
			fmt.Printf("invalid latitude\n")
			return false
		}

		if !(entry.Longitude >= -180.0 && entry.Longitude <= +180.0) {
			fmt.Printf("invalid longitude\n")
			return false
		}

		// IMPORTANT: You must update this check if you ever add a new connection type in the SDK
		if entry.ConnectionType < 0 || entry.ConnectionType > 3 {
			fmt.Printf("invalid connection type\n")
			return false
		}

		// IMPORTANT: You must update this check when you add new platforms to the SDK
		if entry.PlatformType < 0 || entry.PlatformType > 10 {
			fmt.Printf("invalid platform type\n")
			return false
		}

		if entry.NumTags < 0 || entry.NumTags > 8 {
			fmt.Printf("invalid num tags\n")
			return false
		}
	}

	// summary slice only

	if entry.Summary {

		if entry.NumNearRelays < 0 || entry.NumNearRelays > 32 {
			fmt.Printf("invalid num near relays\n")
			return false
		}

		if entry.NumNearRelays > 0 {

			for i := 0; i < int(entry.NumNearRelays); i++ {

				if entry.NearRelayIDs[i] == 0 {
					// Log this but do not return false
					// TODO: investigate why nearRelayID is 0
					fmt.Printf("NearRelayIDs[%d] is 0.\n%+v\n", i, entry)
				}

				if entry.NearRelayRTTs[i] < 0 || entry.NearRelayRTTs[i] > 255 {
					fmt.Printf("invalid near relay rtt\n")
					return false
				}

				if entry.NearRelayJitters[i] < 0 || entry.NearRelayJitters[i] > 255 {
					fmt.Printf("invalid near relay jitter\n")
					return false
				}

				if entry.NearRelayPacketLosses[i] < 0 || entry.NearRelayPacketLosses[i] > 100 {
					fmt.Printf("invalid near relay packet loss\n")
					return false
				}
			}
		}
	}

	// network next only

	if entry.Next {

		if entry.NextRTT < 0 || entry.NextRTT > 255 {
			fmt.Printf("invalid next rtt\n")
			return false
		}

		if entry.NextJitter < 0 || entry.NextJitter > 255 {
			fmt.Printf("invalid next jitter\n")
			return false
		}

		if entry.NextPacketLoss < 0 || entry.NextPacketLoss > 100 {
			fmt.Printf("invalid next packet loss\n")
			return false
		}

		if entry.PredictedNextRTT < 0 || entry.PredictedNextRTT > 255 {
			fmt.Printf("invalid predicted next rtt\n")
			return false
		}

		if entry.NearRelayRTT < 0 || entry.NearRelayRTT > 255 {
			fmt.Printf("invalid near relay rtt\n")
			return false
		}

		if entry.NumNextRelays < 0 || entry.NumNextRelays > 32 {
			fmt.Printf("invalid num next relays\n")
			return false
		}
	}

	return true
}

// Checks all floating point numbers in the BillingEntry2 for NaN and +-Inf and forces them to 0
func (entry *BillingEntry2) CheckNaNOrInf() (bool, []string) {
	var nanOrInfExists bool
	var nanOrInfFields []string

	if math.IsNaN(float64(entry.Latitude)) || math.IsInf(float64(entry.Latitude), 0) {
		nanOrInfFields = append(nanOrInfFields, "Latitude")
		nanOrInfExists = true
		entry.Latitude = float32(0)
	}

	if math.IsNaN(float64(entry.Longitude)) || math.IsInf(float64(entry.Longitude), 0) {
		nanOrInfFields = append(nanOrInfFields, "Longitude")
		nanOrInfExists = true
		entry.Longitude = float32(0)
	}

	return nanOrInfExists, nanOrInfFields
}

// To save bits during serialization, clamp integer and string fields if they go beyond the min
// or max values as defined in BillingEntry2.Serialize()
func (entry *BillingEntry2) ClampEntry() {

	// always

	if entry.DirectRTT < 0 {
		core.Error("BillingEntry2 DirectRTT (%d) < 0. Clamping to 0.", entry.DirectRTT)
		entry.DirectRTT = 0
	}

	if entry.DirectRTT > 1023 {
		core.Debug("BillingEntry2 DirectRTT (%d) > 1023. Clamping to 1023.", entry.DirectRTT)
		entry.DirectRTT = 1023
	}

	if entry.DirectJitter < 0 {
		core.Error("BillingEntry2 DirectJitter (%d) < 0. Clamping to 0.", entry.DirectJitter)
		entry.DirectJitter = 0
	}

	if entry.DirectJitter > 255 {
		core.Debug("BillingEntry2 DirectJitter (%d) > 255. Clamping to 255.", entry.DirectJitter)
		entry.DirectJitter = 255
	}

	if entry.DirectPacketLoss < 0 {
		core.Error("BillingEntry2 DirectPacketLoss (%d) < 0. Clamping to 0.", entry.DirectPacketLoss)
		entry.DirectPacketLoss = 0
	}

	if entry.DirectPacketLoss > 100 {
		core.Debug("BillingEntry2 DirectPacketLoss (%d) > 100. Clamping to 100.", entry.DirectPacketLoss)
		entry.DirectPacketLoss = 100
	}

	if entry.RealPacketLoss < 0 {
		core.Error("BillingEntry2 RealPacketLoss (%d) < 0. Clamping to 0.", entry.RealPacketLoss)
		entry.RealPacketLoss = 0
	}

	if entry.RealPacketLoss > 100 {
		core.Debug("BillingEntry2 RealPacketLoss (%d) > 100. Clamping to 100.", entry.RealPacketLoss)
		entry.RealPacketLoss = 100
	}

	if entry.RealJitter > 1000 {
		core.Debug("BillingEntry2 RealJitter (%d) > 1000. Clamping to 1000.", entry.RealJitter)
		entry.RealJitter = uint32(1000)
	}

	if len(entry.Debug) >= BillingEntryMaxDebugLength {
		core.Debug("BillingEntry2 Debug length (%d) >= BillingEntryMaxDebugLength (%d). Clamping to BillingEntryMaxDebugLength - 1 (%d)", len(entry.Debug), BillingEntryMaxDebugLength, BillingEntryMaxDebugLength-1)
		entry.Debug = entry.Debug[:BillingEntryMaxDebugLength-1]
	}

	if entry.RouteDiversity < 0 {
		core.Error("BillingEntry2 RouteDiversity (%d) < 0. Clamping to 0.", entry.RouteDiversity)
		entry.RouteDiversity = 0
	}

	if entry.RouteDiversity > 32 {
		core.Debug("BillingEntry2 RouteDiversity (%d) > 32. Clamping to 32.", entry.RouteDiversity)
		entry.RouteDiversity = 32
	}

	// first slice only

	if entry.SliceNumber == 0 {

		if len(entry.ISP) >= BillingEntryMaxISPLength {
			core.Debug("BillingEntry2 ISP length (%d) >= BillingEntryMaxISPLength (%d). Clamping to BillingEntryMaxISPLength - 1(%d)", len(entry.ISP), BillingEntryMaxISPLength, BillingEntryMaxISPLength-1)
			entry.ISP = entry.ISP[:BillingEntryMaxISPLength-1]
		}

		if entry.ConnectionType < 0 {
			core.Error("BillingEntry2 ConnectionType (%d) < 0. Clamping to 0.", entry.ConnectionType)
			entry.ConnectionType = 0
		}

		// IMPORTANT: You must update this check if you ever add a new connection type in the SDK
		if entry.ConnectionType > 3 {
			core.Debug("BillingEntry2 ConnectionType (%d) > 3. Clamping to 0 (unknown).", entry.ConnectionType)
			entry.ConnectionType = 0
		}

		if entry.PlatformType < 0 {
			core.Error("BillingEntry2 PlatformType (%d) < 0. Clamping to 0.", entry.PlatformType)
			entry.PlatformType = 0
		}

		// IMPORTANT: You must update this check when you add new platforms to the SDK
		if entry.PlatformType > 10 {
			core.Debug("BillingEntry2 PlatformType (%d) > 10. Clamping to 0 (unknown).", entry.PlatformType)
			entry.PlatformType = 0
		}

		if len(entry.SDKVersion) >= BillingEntryMaxSDKVersionLength {
			core.Debug("BillingEntry2 SDKVersion length (%d) >= BillingEntryMaxSDKVersionLength (%d). Clamping to BillingEntryMaxSDKVersionLength - 1 (%d)", len(entry.SDKVersion), BillingEntryMaxSDKVersionLength, BillingEntryMaxSDKVersionLength-1)
			entry.SDKVersion = entry.SDKVersion[:BillingEntryMaxSDKVersionLength-1]
		}

		if entry.NumTags < 0 {
			core.Error("BillingEntry2 NumTags (%d) < 0. Clamping to 0.", entry.NumTags)
			entry.NumTags = 0
		}

		if entry.NumTags > BillingEntryMaxTags {
			core.Debug("BillingEntry2 NumTags (%d) > BillingEntryMaxTags (%d). Clamping to BillingEntryMaxTags (%d).", entry.NumTags, BillingEntryMaxTags, BillingEntryMaxTags)
			entry.NumTags = BillingEntryMaxTags
		}
	}

	// summary slice only

	if entry.Summary {

		if entry.NumNearRelays < 0 {
			core.Error("BillingEntry2 NumNearRelays (%d) < 0. Clamping to 0.", entry.NumNearRelays)
			entry.NumNearRelays = 0
		}

		if entry.NumNearRelays > BillingEntryMaxNearRelays {
			core.Debug("BillingEntry2 NumNearRelays (%d) > BillingEntryMaxNearRelays (%d). Clamping to BillingEntryMaxNearRelays (%d).", entry.NumNearRelays, BillingEntryMaxNearRelays, BillingEntryMaxNearRelays)
			entry.NumNearRelays = BillingEntryMaxNearRelays
		}

		for i := 0; i < int(entry.NumNearRelays); i++ {
			if entry.NearRelayRTTs[i] < 0 {
				core.Error("BillingEntry2 NearRelayRTT[%d] (%d) < 0. Clamping to 0.", i, entry.NearRelayRTTs[i])
				entry.NearRelayRTTs[i] = 0
			}

			if entry.NearRelayRTTs[i] > 255 {
				core.Debug("BillingEntry2 NearRelayRTTs[%d] (%d) > 255. Clamping to 255.", i, entry.NearRelayRTTs[i])
				entry.NearRelayRTTs[i] = 255
			}

			if entry.NearRelayJitters[i] < 0 {
				core.Error("BillingEntry2 NearRelayRTT[%d] (%d) < 0. Clamping to 0.", i, entry.NearRelayJitters[i])
				entry.NearRelayJitters[i] = 0
			}

			if entry.NearRelayJitters[i] > 255 {
				core.Debug("BillingEntry2 NearRelayJitters[%d] (%d) > 255. Clamping to 255.", i, entry.NearRelayJitters[i])
				entry.NearRelayJitters[i] = 255
			}

			if entry.NearRelayPacketLosses[i] < 0 {
				core.Error("BillingEntry2 NearRelayRTT[%d] (%d) < 0. Clamping to 0.", i, entry.NearRelayPacketLosses[i])
				entry.NearRelayPacketLosses[i] = 0
			}

			if entry.NearRelayPacketLosses[i] > 100 {
				core.Debug("BillingEntry2 NearRelayPacketLosses[%d] (%d) > 100. Clamping to 100.", i, entry.NearRelayPacketLosses[i])
				entry.NearRelayPacketLosses[i] = 100
			}
		}
	}

	// network next only

	if entry.Next {

		if entry.NextRTT < 0 {
			core.Error("BillingEntry2 NextRTT (%d) < 0. Clamping to 0.", entry.NextRTT)
			entry.NextRTT = 0
		}

		if entry.NextRTT > 255 {
			core.Debug("BillingEntry2 NextRTT (%d) > 255. Clamping to 255.", entry.NextRTT)
			entry.NextRTT = 255
		}

		if entry.NextJitter < 0 {
			core.Error("BillingEntry2 NextJitter (%d) < 0. Clamping to 0.", entry.NextJitter)
			entry.NextJitter = 0
		}

		if entry.NextJitter > 255 {
			core.Debug("BillingEntry2 NextJitter (%d) > 255. Clamping to 255.", entry.NextJitter)
			entry.NextJitter = 255
		}

		if entry.NextPacketLoss < 0 {
			core.Error("BillingEntry2 NextPacketLoss (%d) < 0. Clamping to 0.", entry.NextPacketLoss)
			entry.NextPacketLoss = 0
		}

		if entry.NextPacketLoss > 100 {
			core.Debug("BillingEntry2 NextPacketLoss (%d) > 100. Clamping to 100.", entry.NextPacketLoss)
			entry.NextPacketLoss = 100
		}

		if entry.PredictedNextRTT < 0 {
			core.Error("BillingEntry2 PredictedNextRTT (%d) < 0. Clamping to 0.", entry.PredictedNextRTT)
			entry.PredictedNextRTT = 0
		}

		if entry.PredictedNextRTT > 255 {
			core.Debug("BillingEntry2 PredictedNextRTT (%d) > 255. Clamping to 255.", entry.PredictedNextRTT)
			entry.PredictedNextRTT = 255
		}

		if entry.NearRelayRTT < 0 {
			core.Error("BillingEntry2 NearRelayRTT (%d) < 0. Clamping to 0.", entry.NearRelayRTT)
			entry.NearRelayRTT = 0
		}

		if entry.NearRelayRTT > 255 {
			core.Debug("BillingEntry2 NearRelayRTT (%d) > 255. Clamping to 255.", entry.NearRelayRTT)
			entry.NearRelayRTT = 255
		}

		if entry.NumNextRelays < 0 {
			core.Error("BillingEntry2 NumNextRelays (%d) < 0. Clamping to 0.", entry.NumNextRelays)
			entry.NumNextRelays = 0
		}

		if entry.NumNextRelays > BillingEntryMaxRelays {
			core.Debug("BillingEntry2 NumNextRelays (%d) > BillingEntryMaxRelays (%d). Clamping to BillingEntryMaxRelays (%d).", entry.NumNextRelays, BillingEntryMaxRelays, BillingEntryMaxRelays)
			entry.NumNextRelays = BillingEntryMaxRelays
		}
	}
}

func (entry *BillingEntry2) Save() (map[string]bigquery.Value, string, error) {

	e := make(map[string]bigquery.Value)

	/*
		1. Always

		These values are written for every slice.
	*/

	e["timestamp"] = int(entry.Timestamp)
	e["sessionID"] = int(entry.SessionID)
	e["sliceNumber"] = int(entry.SliceNumber)
	e["directRTT"] = int(entry.DirectRTT)
	e["directJitter"] = int(entry.DirectJitter)
	e["directPacketLoss"] = int(entry.DirectPacketLoss)
	e["realPacketLoss"] = float64(entry.RealPacketLoss) + float64(entry.RealPacketLoss_Frac)/256.0
	e["realJitter"] = int(entry.RealJitter)

	if entry.Next {
		e["next"] = true
	}

	if entry.Flagged {
		e["flagged"] = true
	}

	if entry.Summary {
		e["summary"] = true
	}

	if entry.UseDebug {
		e["debug"] = entry.Debug
	}

	if entry.RouteDiversity > 0 {
		e["routeDiversity"] = int(entry.RouteDiversity)
	}

	/*
		2. First slice only

		These values are serialized only for slice 0.
	*/

	if entry.SliceNumber == 0 {

		e["datacenterID"] = int(entry.DatacenterID)
		e["buyerID"] = int(entry.BuyerID)
		e["userHash"] = int(entry.UserHash)
		e["envelopeBytesUp"] = int(entry.EnvelopeBytesUp)
		e["envelopeBytesDown"] = int(entry.EnvelopeBytesDown)
		e["latitude"] = entry.Latitude
		e["longitude"] = entry.Longitude
		e["isp"] = entry.ISP
		e["connectionType"] = int(entry.ConnectionType)
		e["platformType"] = int(entry.PlatformType)
		e["sdkVersion"] = entry.SDKVersion

		if entry.NumTags > 0 {
			tags := make([]bigquery.Value, entry.NumTags)
			for i := 0; i < int(entry.NumTags); i++ {
				tags[i] = int(entry.Tags[i])
			}
			e["tags"] = tags
		}

		if entry.ABTest {
			e["abTest"] = true
		}

		if entry.Pro {
			e["pro"] = true
		}
	}

	/*
		3. Summary slice only

		These values are serialized only for the summary slice (at the end of the session)
	*/

	if entry.Summary {

		e["clientToServerPacketsSent"] = int(entry.ClientToServerPacketsSent)
		e["serverToClientPacketsSent"] = int(entry.ServerToClientPacketsSent)
		e["clientToServerPacketsLost"] = int(entry.ClientToServerPacketsLost)
		e["serverToClientPacketsLost"] = int(entry.ServerToClientPacketsLost)
		e["clientToServerPacketsOutOfOrder"] = int(entry.ClientToServerPacketsOutOfOrder)
		e["serverToClientPacketsOutOfOrder"] = int(entry.ServerToClientPacketsOutOfOrder)

		if entry.NumNearRelays > 0 {

			nearRelayIDs := make([]bigquery.Value, entry.NumNearRelays)
			nearRelayRTTs := make([]bigquery.Value, entry.NumNearRelays)
			nearRelayJitters := make([]bigquery.Value, entry.NumNearRelays)
			nearRelayPacketLosses := make([]bigquery.Value, entry.NumNearRelays)

			for i := 0; i < int(entry.NumNearRelays); i++ {
				nearRelayIDs[i] = int(entry.NearRelayIDs[i])
				nearRelayRTTs[i] = int(entry.NearRelayRTTs[i])
				nearRelayJitters[i] = int(entry.NearRelayJitters[i])
				nearRelayPacketLosses[i] = int(entry.NearRelayPacketLosses[i])
			}

			e["nearRelayIDs"] = nearRelayIDs
			e["nearRelayRTTs"] = nearRelayRTTs
			e["nearRelayJitters"] = nearRelayJitters
			e["nearRelayPacketLosses"] = nearRelayPacketLosses

		}

	}

	/*
		4. Network Next Only

		These values are serialized only when a slice is on network next.
	*/

	if entry.Next {

		e["nextRTT"] = int(entry.NextRTT)
		e["nextJitter"] = int(entry.NextJitter)
		e["nextPacketLoss"] = int(entry.NextPacketLoss)
		e["predictedNextRTT"] = int(entry.PredictedNextRTT)
		e["nearRelayRTT"] = int(entry.NearRelayRTT)

		if entry.NumNextRelays > 0 {

			nextRelays := make([]bigquery.Value, entry.NumNextRelays)
			nextRelayPrice := make([]bigquery.Value, entry.NumNextRelays)

			for i := 0; i < int(entry.NumNextRelays); i++ {
				nextRelays[i] = int(entry.NextRelays[i])
				nextRelayPrice[i] = int(entry.NextRelayPrice[i])
			}

			e["nextRelays"] = nextRelays
			e["nextRelayPrice"] = nextRelayPrice
		}

		e["totalPrice"] = int(entry.TotalPrice)

		if entry.Uncommitted {
			e["uncommitted"] = true
		}

		if entry.Multipath {
			e["multipath"] = true
		}

		if entry.RTTReduction {
			e["rttReduction"] = true
		}

		if entry.PacketLossReduction {
			e["packetLossReduction"] = true
		}

		if entry.RouteChanged {
			e["routeChanged"] = true
		}

		e["nextBytesUp"] = int(entry.NextBytesUp)
		e["nextBytesDown"] = int(entry.NextBytesDown)

	}

	/*
		5. Error State Only

		These values are only serialized when the session is in an error state (rare).
	*/

	if entry.FallbackToDirect {
		e["fallbackToDirect"] = true
	}

	if entry.MultipathVetoed {
		e["multipathVetoed"] = true
	}

	if entry.Mispredicted {
		e["mispredicted"] = true
	}

	if entry.Vetoed {
		e["vetoed"] = true
	}

	if entry.LatencyWorse {
		e["latencyWorse"] = true
	}

	if entry.NoRoute {
		e["noRoute"] = true
	}

	if entry.NextLatencyTooHigh {
		e["nextLatencyTooHigh"] = true
	}

	if entry.CommitVeto {
		e["commitVeto"] = true
	}

	if entry.UnknownDatacenter {
		e["unknownDatacenter"] = true
	}

	if entry.DatacenterNotEnabled {
		e["datacenterNotEnabled"] = true
	}

	if entry.BuyerNotLive {
		e["buyerNotLive"] = true
	}

	if entry.StaleRouteMatrix {
		e["staleRouteMatrix"] = true
	}

	return e, "", nil
}
