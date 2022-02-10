package billing

import (
	"fmt"
	"math"
	"time"

	"cloud.google.com/go/bigquery"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/encoding"
)

const (
	BillingEntryVersion2 = uint32(8)

	MaxBillingEntry2Bytes = 4096

	BillingEntryMaxRelays           = 5
	BillingEntryMaxAddressLength    = 256
	BillingEntryMaxISPLength        = 64
	BillingEntryMaxSDKVersionLength = 11
	BillingEntryMaxDebugLength      = 2048
	BillingEntryMaxNearRelays       = 32
	BillingEntryMaxTags             = 8
)

type BillingEntry2 struct {

	// always

	Version             uint32
	Timestamp           uint32
	SessionID           uint64
	SliceNumber         uint32
	DirectMinRTT        int32
	DirectMaxRTT        int32
	DirectPrimeRTT      int32
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
	UserFlags           uint64

	// first slice and summary slice only

	DatacenterID      uint64
	BuyerID           uint64
	UserHash          uint64
	EnvelopeBytesUp   uint64
	EnvelopeBytesDown uint64
	Latitude          float32
	Longitude         float32
	ClientAddress     string
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
	EverOnNext                      bool
	SessionDuration                 uint32
	TotalPriceSum                   uint64
	EnvelopeBytesUpSum              uint64
	EnvelopeBytesDownSum            uint64
	DurationOnNext                  uint32
	StartTimestamp                  uint32

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

// A stripped down struct of BillingEntry2 containing important fields from the
// first and summary slice. This is written to a separate table in BigQuery.
type BillingEntry2Summary struct {
	SessionID                       uint64
	Summary                         bool
	BuyerID                         uint64
	UserHash                        uint64
	DatacenterID                    uint64
	StartTimestamp                  uint32
	Latitude                        float32
	Longitude                       float32
	ISP                             string
	ConnectionType                  int32
	PlatformType                    int32
	NumTags                         int32
	Tags                            [BillingEntryMaxTags]uint64
	ABTest                          bool
	Pro                             bool
	SDKVersion                      string
	EnvelopeBytesUp                 uint64
	EnvelopeBytesDown               uint64
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
	EverOnNext                      bool
	SessionDuration                 uint32
	TotalPriceSum                   uint64
	EnvelopeBytesUpSum              uint64
	EnvelopeBytesDownSum            uint64
	DurationOnNext                  uint32
	ClientAddress                   string
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

	stream.SerializeInteger(&entry.DirectMinRTT, 0, 1023)

	/*
		Version 7

		Includes DirectMaxRTT and DirectPrimeRTT stats from SDK 4.0.18.
		DirectRTT was changed to DirectMinRTT.
	*/
	if entry.Version >= uint32(7) {
		stream.SerializeInteger(&entry.DirectMaxRTT, 0, 1023)
		stream.SerializeInteger(&entry.DirectPrimeRTT, 0, 1023)
	}

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
		Version 8

		Includes UserFlags from SDK 4.20.0 driven by next_server_event().
	*/
	if entry.Version >= uint32(8) {
		stream.SerializeUint64(&entry.UserFlags)
	}

	/*
		2. First slice and summary slice only

		These values are serialized only for slice 0 and summary slice.

		NOTE: Prior to version 3, these fields were only serialized for slice 0.
	*/

	if entry.Version >= uint32(3) {
		if entry.SliceNumber == 0 || entry.Summary {

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
	} else {
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

	/*
		Version 2

		Includes the following in the summary slice:
			- Sum of TotalPrice
			- Sum of EnvelopeBytesUp
			- Sum of EnvelopeBytesDown
			- Duration of the session (in seconds)
			- EverOnNext
	*/
	if entry.Version >= uint32(2) {
		if entry.Summary {

			stream.SerializeBool(&entry.EverOnNext)

			stream.SerializeUint32(&entry.SessionDuration)

			if entry.EverOnNext {

				stream.SerializeUint64(&entry.TotalPriceSum)

				stream.SerializeUint64(&entry.EnvelopeBytesUpSum)
				stream.SerializeUint64(&entry.EnvelopeBytesDownSum)

			}

		}
	}

	/*
		Version 4

		Includes the following in the summary slice:
			- Duration of the session on next (in seconds)
	*/
	if entry.Version >= uint32(4) {
		if entry.Summary && entry.EverOnNext {
			stream.SerializeUint32(&entry.DurationOnNext)
		}
	}

	/*
		Version 5

		Includes client IP address in first and summary slice
	*/
	if entry.Version >= uint32(5) {
		if entry.SliceNumber == 0 || entry.Summary {
			stream.SerializeString(&entry.ClientAddress, BillingEntryMaxAddressLength)
		}
	}

	/*
		Version 6

		Includes session start timestamp in the summary slice

		NOTE: Prior to version 6, summary slices were written to the billing2 table
	*/
	if entry.Version >= uint32(6) {
		if entry.Summary {
			stream.SerializeUint32(&entry.StartTimestamp)
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

	if entry.DirectMinRTT < 0 || entry.DirectMinRTT > 1023 {
		fmt.Printf("invalid direct min rtt\n")
		return false
	}

	if entry.DirectMaxRTT < 0 || entry.DirectMaxRTT > 1023 {
		fmt.Printf("invalid direct max rtt\n")
		return false
	}

	if entry.DirectPrimeRTT < 0 || entry.DirectPrimeRTT > 1023 {
		fmt.Printf("invalid direct prime rtt\n")
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

	if entry.DirectMinRTT < 0 {
		core.Error("BillingEntry2 DirectMinRTT (%d) < 0. Clamping to 0.", entry.DirectMinRTT)
		entry.DirectMinRTT = 0
	}

	if entry.DirectMinRTT > 1023 {
		core.Debug("BillingEntry2 DirectMinRTT (%d) > 1023. Clamping to 1023.", entry.DirectMinRTT)
		entry.DirectMinRTT = 1023
	}

	if entry.DirectMaxRTT < 0 {
		core.Error("BillingEntry2 DirectMaxRTT (%d) < 0. Clamping to 0.", entry.DirectMaxRTT)
		entry.DirectMaxRTT = 0
	}

	if entry.DirectMaxRTT > 1023 {
		core.Debug("BillingEntry2 DirectMaxRTT (%d) > 1023. Clamping to 1023.", entry.DirectMaxRTT)
		entry.DirectMaxRTT = 1023
	}

	if entry.DirectPrimeRTT < 0 {
		core.Error("BillingEntry2 DirectPrimeRTT (%d) < 0. Clamping to 0.", entry.DirectPrimeRTT)
		entry.DirectPrimeRTT = 0
	}

	if entry.DirectPrimeRTT > 1023 {
		core.Debug("BillingEntry2 DirectPrimeRTT (%d) > 1023. Clamping to 1023.", entry.DirectPrimeRTT)
		entry.DirectPrimeRTT = 1023
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

	// first slice and summary slice only

	if entry.SliceNumber == 0 || entry.Summary {

		if len(entry.ClientAddress) >= BillingEntryMaxAddressLength {
			core.Debug("BillingEntry2 Client IP Address length (%d) >= BillingEntryMaxAddressLength (%d). Clamping to BillingEntryMaxAddressLength - 1 (%d)", len(entry.ClientAddress), BillingEntryMaxAddressLength, BillingEntryMaxAddressLength-1)
			entry.ClientAddress = entry.ClientAddress[:BillingEntryMaxAddressLength-1]
		}

		if len(entry.ISP) >= BillingEntryMaxISPLength {
			core.Debug("BillingEntry2 ISP length (%d) >= BillingEntryMaxISPLength (%d). Clamping to BillingEntryMaxISPLength - 1 (%d)", len(entry.ISP), BillingEntryMaxISPLength, BillingEntryMaxISPLength-1)
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
	e["directRTT"] = int(entry.DirectMinRTT) // NOTE: directRTT refers to DirectMinRTT as of version 7
	e["directMaxRTT"] = int(entry.DirectMaxRTT)
	e["directPrimeRTT"] = int(entry.DirectPrimeRTT)
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

	if entry.UserFlags > 0 {
		e["userFlags"] = int(entry.UserFlags)
	}

	/*
		2. First slice and summary slice only

		These values are serialized only for slice 0 and the summary slice.
	*/

	if entry.SliceNumber == 0 || entry.Summary {

		e["datacenterID"] = int(entry.DatacenterID)
		e["buyerID"] = int(entry.BuyerID)
		e["userHash"] = int(entry.UserHash)
		e["envelopeBytesUp"] = int(entry.EnvelopeBytesUp)
		e["envelopeBytesDown"] = int(entry.EnvelopeBytesDown)
		e["latitude"] = entry.Latitude
		e["longitude"] = entry.Longitude
		e["clientAddress"] = entry.ClientAddress
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

		e["everOnNext"] = entry.EverOnNext

		e["sessionDuration"] = int(entry.SessionDuration)

		if entry.EverOnNext {
			e["totalPriceSum"] = int(entry.TotalPriceSum)
			e["envelopeBytesUpSum"] = int(entry.EnvelopeBytesUpSum)
			e["envelopeBytesDownSum"] = int(entry.EnvelopeBytesDownSum)
			e["durationOnNext"] = int(entry.DurationOnNext)
		}

		e["startTimestamp"] = int(entry.StartTimestamp)

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

func (entry *BillingEntry2) GetSummaryStruct() *BillingEntry2Summary {
	return &BillingEntry2Summary{
		SessionID:                       entry.SessionID,
		Summary:                         entry.Summary,
		BuyerID:                         entry.BuyerID,
		UserHash:                        entry.UserHash,
		DatacenterID:                    entry.DatacenterID,
		StartTimestamp:                  entry.StartTimestamp,
		Latitude:                        entry.Latitude,
		Longitude:                       entry.Longitude,
		ISP:                             entry.ISP,
		ConnectionType:                  entry.ConnectionType,
		PlatformType:                    entry.PlatformType,
		NumTags:                         entry.NumTags,
		Tags:                            entry.Tags,
		ABTest:                          entry.ABTest,
		Pro:                             entry.Pro,
		SDKVersion:                      entry.SDKVersion,
		EnvelopeBytesUp:                 entry.EnvelopeBytesUp,
		EnvelopeBytesDown:               entry.EnvelopeBytesDown,
		ClientToServerPacketsSent:       entry.ClientToServerPacketsSent,
		ServerToClientPacketsSent:       entry.ServerToClientPacketsSent,
		ClientToServerPacketsLost:       entry.ClientToServerPacketsLost,
		ServerToClientPacketsLost:       entry.ServerToClientPacketsLost,
		ClientToServerPacketsOutOfOrder: entry.ClientToServerPacketsOutOfOrder,
		ServerToClientPacketsOutOfOrder: entry.ServerToClientPacketsOutOfOrder,
		NumNearRelays:                   entry.NumNearRelays,
		NearRelayIDs:                    entry.NearRelayIDs,
		NearRelayRTTs:                   entry.NearRelayRTTs,
		NearRelayJitters:                entry.NearRelayJitters,
		NearRelayPacketLosses:           entry.NearRelayPacketLosses,
		EverOnNext:                      entry.EverOnNext,
		SessionDuration:                 entry.SessionDuration,
		TotalPriceSum:                   entry.TotalPriceSum,
		EnvelopeBytesUpSum:              entry.EnvelopeBytesUpSum,
		EnvelopeBytesDownSum:            entry.EnvelopeBytesDownSum,
		DurationOnNext:                  entry.DurationOnNext,
		ClientAddress:                   entry.ClientAddress,
	}
}

func (entry *BillingEntry2Summary) Save() (map[string]bigquery.Value, string, error) {

	e := make(map[string]bigquery.Value)

	/*
		1. Always

		These values are written for every slice.
	*/

	e["sessionID"] = int(entry.SessionID)

	if entry.Summary {

		/*
			2. First slice and summary slice only

			These values are serialized only for slice 0 and the summary slice.
		*/

		e["datacenterID"] = int(entry.DatacenterID)
		e["buyerID"] = int(entry.BuyerID)
		e["userHash"] = int(entry.UserHash)
		e["envelopeBytesUp"] = int(entry.EnvelopeBytesUp)
		e["envelopeBytesDown"] = int(entry.EnvelopeBytesDown)
		e["latitude"] = entry.Latitude
		e["longitude"] = entry.Longitude
		e["clientAddress"] = entry.ClientAddress
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

		/*
			3. Summary slice only

			These values are serialized only for the summary slice (at the end of the session)
		*/

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

		e["everOnNext"] = entry.EverOnNext

		e["sessionDuration"] = int(entry.SessionDuration)

		if entry.EverOnNext {
			e["totalPriceSum"] = int(entry.TotalPriceSum)
			e["envelopeBytesUpSum"] = int(entry.EnvelopeBytesUpSum)
			e["envelopeBytesDownSum"] = int(entry.EnvelopeBytesDownSum)
			e["durationOnNext"] = int(entry.DurationOnNext)
		}

		if entry.StartTimestamp == 0 {
			// In case startTimestamp is 0 during transition
			e["startTimestamp"] = int(time.Now().Add(time.Duration(entry.SessionDuration) * time.Second * -1).Unix())
		} else {
			e["startTimestamp"] = int(entry.StartTimestamp)
		}
	}

	return e, "", nil
}
