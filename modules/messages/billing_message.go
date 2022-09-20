package messages

import (
	"cloud.google.com/go/bigquery"
	"github.com/networknext/backend/modules/encoding"
)

const (
	BillingMessageVersion = uint32(9)

	MaxBillingMessageBytes = 4096

	BillingMessageMaxRelays           = 5
	BillingMessageMaxAddressLength    = 256
	BillingMessageMaxISPLength        = 64
	BillingMessageMaxSDKVersionLength = 11
	BillingMessageMaxDebugLength      = 2048
	BillingMessageMaxNearRelays       = 32
	BillingMessageMaxTags             = 8
)

type BillingMessage struct {

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
	TryBeforeYouBuy     bool

	// first slice and summary slice only

	DatacenterID      uint64
	BuyerID           uint64
	UserHash          uint64
	EnvelopeBytesUp   uint64
	EnvelopeBytesDown uint64
	Latitude          float32
	Longitude         float32
	ClientAddress     string
	ServerAddress     string
	ISP               string
	ConnectionType    int32
	PlatformType      int32
	SDKVersion        string
	NumTags           int32
	Tags              [BillingMessageMaxTags]uint64
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
	NearRelayIDs                    [BillingMessageMaxNearRelays]uint64
	NearRelayRTTs                   [BillingMessageMaxNearRelays]int32
	NearRelayJitters                [BillingMessageMaxNearRelays]int32
	NearRelayPacketLosses           [BillingMessageMaxNearRelays]int32
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
	NextRelays          [BillingMessageMaxRelays]uint64
	NextRelayPrice      [BillingMessageMaxRelays]uint64
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

func (message *BillingMessage) Serialize(stream encoding.Stream) error {

	/*
		1. Always

		These values are serialized in every slice
	*/

	stream.SerializeBits(&message.Version, 32)
	stream.SerializeBits(&message.Timestamp, 32)
	stream.SerializeUint64(&message.SessionID)

	small := false
	if message.SliceNumber < 1024 {
		small = true
	}
	stream.SerializeBool(&small)
	if small {
		stream.SerializeBits(&message.SliceNumber, 10)
	} else {
		stream.SerializeBits(&message.SliceNumber, 32)
	}

	stream.SerializeInteger(&message.DirectMinRTT, 0, 1023)

	/*
		Version 7

		Includes DirectMaxRTT and DirectPrimeRTT stats from SDK 4.0.18.
		DirectRTT was changed to DirectMinRTT.
	*/
	if message.Version >= uint32(7) {
		stream.SerializeInteger(&message.DirectMaxRTT, 0, 1023)
		stream.SerializeInteger(&message.DirectPrimeRTT, 0, 1023)
	}

	stream.SerializeInteger(&message.DirectJitter, 0, 255)
	stream.SerializeInteger(&message.DirectPacketLoss, 0, 100)

	stream.SerializeInteger(&message.RealPacketLoss, 0, 100)
	stream.SerializeBits(&message.RealPacketLoss_Frac, 8)
	stream.SerializeUint32(&message.RealJitter)

	stream.SerializeBool(&message.Next)
	stream.SerializeBool(&message.Flagged)
	stream.SerializeBool(&message.Summary)

	stream.SerializeBool(&message.UseDebug)
	stream.SerializeString(&message.Debug, BillingMessageMaxDebugLength)

	stream.SerializeInteger(&message.RouteDiversity, 0, 32)

	/*
		Version 8

		Includes UserFlags from SDK 4.20.0 driven by next_server_event().
	*/
	if message.Version >= uint32(8) {
		stream.SerializeUint64(&message.UserFlags)
	}

	/*
		Version 9

		Includes server IP address in first and summary slices as well as TryBeforeYouBuy for all slices
	*/
	if message.Version >= uint32(9) {
		stream.SerializeBool(&message.TryBeforeYouBuy)
	}

	/*
		2. First slice and summary slice only

		These values are serialized only for slice 0 and summary slice.

		NOTE: Prior to version 3, these fields were only serialized for slice 0.
	*/

	if message.Version >= uint32(3) {
		if message.SliceNumber == 0 || message.Summary {

			stream.SerializeUint64(&message.DatacenterID)
			stream.SerializeUint64(&message.BuyerID)
			stream.SerializeUint64(&message.UserHash)
			stream.SerializeUint64(&message.EnvelopeBytesUp)
			stream.SerializeUint64(&message.EnvelopeBytesDown)
			stream.SerializeFloat32(&message.Latitude)
			stream.SerializeFloat32(&message.Longitude)
			stream.SerializeString(&message.ISP, BillingMessageMaxISPLength)
			stream.SerializeInteger(&message.ConnectionType, 0, 3) // todo: constant
			stream.SerializeInteger(&message.PlatformType, 0, 10)  // todo: constant
			stream.SerializeString(&message.SDKVersion, BillingMessageMaxSDKVersionLength)
			stream.SerializeInteger(&message.NumTags, 0, BillingMessageMaxTags)
			for i := 0; i < int(message.NumTags); i++ {
				stream.SerializeUint64(&message.Tags[i])
			}
			stream.SerializeBool(&message.ABTest)
			stream.SerializeBool(&message.Pro)

		}
	} else {
		if message.SliceNumber == 0 {

			stream.SerializeUint64(&message.DatacenterID)
			stream.SerializeUint64(&message.BuyerID)
			stream.SerializeUint64(&message.UserHash)
			stream.SerializeUint64(&message.EnvelopeBytesUp)
			stream.SerializeUint64(&message.EnvelopeBytesDown)
			stream.SerializeFloat32(&message.Latitude)
			stream.SerializeFloat32(&message.Longitude)
			stream.SerializeString(&message.ISP, BillingMessageMaxISPLength)
			stream.SerializeInteger(&message.ConnectionType, 0, 3) // todo: constant
			stream.SerializeInteger(&message.PlatformType, 0, 10)  // todo: constant
			stream.SerializeString(&message.SDKVersion, BillingMessageMaxSDKVersionLength)
			stream.SerializeInteger(&message.NumTags, 0, BillingMessageMaxTags)
			for i := 0; i < int(message.NumTags); i++ {
				stream.SerializeUint64(&message.Tags[i])
			}
			stream.SerializeBool(&message.ABTest)
			stream.SerializeBool(&message.Pro)

		}
	}

	/*
		3. Summary slice only

		These values are serialized only for the summary slice (at the end of the session)
	*/

	if message.Summary {

		stream.SerializeUint64(&message.ClientToServerPacketsSent)
		stream.SerializeUint64(&message.ServerToClientPacketsSent)
		stream.SerializeUint64(&message.ClientToServerPacketsLost)
		stream.SerializeUint64(&message.ServerToClientPacketsLost)
		stream.SerializeUint64(&message.ClientToServerPacketsOutOfOrder)
		stream.SerializeUint64(&message.ServerToClientPacketsOutOfOrder)
		stream.SerializeInteger(&message.NumNearRelays, 0, BillingMessageMaxNearRelays)
		for i := 0; i < int(message.NumNearRelays); i++ {
			stream.SerializeUint64(&message.NearRelayIDs[i])
			stream.SerializeInteger(&message.NearRelayRTTs[i], 0, 255)
			stream.SerializeInteger(&message.NearRelayJitters[i], 0, 255)
			stream.SerializeInteger(&message.NearRelayPacketLosses[i], 0, 100)
		}

	}

	/*
		4. Network Next Only

		These values are serialized only when a slice is on network next.
	*/

	if message.Next {

		stream.SerializeInteger(&message.NextRTT, 0, 255)
		stream.SerializeInteger(&message.NextJitter, 0, 255)
		stream.SerializeInteger(&message.NextPacketLoss, 0, 100)
		stream.SerializeInteger(&message.PredictedNextRTT, 0, 255)
		stream.SerializeInteger(&message.NearRelayRTT, 0, 255)

		stream.SerializeInteger(&message.NumNextRelays, 0, BillingMessageMaxRelays)
		for i := 0; i < int(message.NumNextRelays); i++ {
			stream.SerializeUint64(&message.NextRelays[i])
			stream.SerializeUint64(&message.NextRelayPrice[i])
		}

		stream.SerializeUint64(&message.TotalPrice)
		stream.SerializeBool(&message.Uncommitted)
		stream.SerializeBool(&message.Multipath)
		stream.SerializeBool(&message.RTTReduction)
		stream.SerializeBool(&message.PacketLossReduction)
		stream.SerializeBool(&message.RouteChanged)
	}

	/*
		5. Error State Only

		These values are only serialized when the session is in an error state (rare).
	*/

	errorState := false

	if stream.IsWriting() {
		errorState =
			message.FallbackToDirect ||
				message.MultipathVetoed ||
				message.Mispredicted ||
				message.Vetoed ||
				message.LatencyWorse ||
				message.NoRoute ||
				message.NextLatencyTooHigh ||
				message.CommitVeto ||
				message.UnknownDatacenter ||
				message.DatacenterNotEnabled ||
				message.BuyerNotLive ||
				message.StaleRouteMatrix
	}

	stream.SerializeBool(&errorState)

	if errorState {

		stream.SerializeBool(&message.FallbackToDirect)
		stream.SerializeBool(&message.MultipathVetoed)
		stream.SerializeBool(&message.Mispredicted)
		stream.SerializeBool(&message.Vetoed)
		stream.SerializeBool(&message.LatencyWorse)
		stream.SerializeBool(&message.NoRoute)
		stream.SerializeBool(&message.NextLatencyTooHigh)
		stream.SerializeBool(&message.CommitVeto)
		stream.SerializeBool(&message.UnknownDatacenter)
		stream.SerializeBool(&message.DatacenterNotEnabled)
		stream.SerializeBool(&message.BuyerNotLive)
		stream.SerializeBool(&message.StaleRouteMatrix)

	}

	/*
		Version 1

		Include Next Bytes Up/Down for slices on next.
	*/
	if message.Version >= uint32(1) {

		if message.Next {

			stream.SerializeUint64(&message.NextBytesUp)
			stream.SerializeUint64(&message.NextBytesDown)

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
	if message.Version >= uint32(2) {
		if message.Summary {

			stream.SerializeBool(&message.EverOnNext)

			stream.SerializeUint32(&message.SessionDuration)

			if message.EverOnNext {

				stream.SerializeUint64(&message.TotalPriceSum)

				stream.SerializeUint64(&message.EnvelopeBytesUpSum)
				stream.SerializeUint64(&message.EnvelopeBytesDownSum)

			}

		}
	}

	/*
		Version 4

		Includes the following in the summary slice:
			- Duration of the session on next (in seconds)
	*/
	if message.Version >= uint32(4) {
		if message.Summary && message.EverOnNext {
			stream.SerializeUint32(&message.DurationOnNext)
		}
	}

	/*
		Version 5

		Includes client IP address in first and summary slice
	*/
	if message.Version >= uint32(5) {
		if message.SliceNumber == 0 || message.Summary {
			stream.SerializeString(&message.ClientAddress, BillingMessageMaxAddressLength)
		}
	}

	/*
		Version 6

		Includes session start timestamp in the summary slice

		NOTE: Prior to version 6, summary slices were written to the billing2 table
	*/
	if message.Version >= uint32(6) {
		if message.Summary {
			stream.SerializeUint32(&message.StartTimestamp)
		}
	}

	/*
		Version 9

		Includes server IP address in first and summary slices as well as TryBeforeYouBuy for all slices
	*/
	if message.Version >= uint32(9) {
		if message.SliceNumber == 0 || message.Summary {
			stream.SerializeString(&message.ServerAddress, BillingMessageMaxAddressLength)
		}
	}
	return stream.Error()
}

/*
func WriteBillingEntry(entry *BillingEntry) ([]byte, error) {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("recovered from panic during BillingEntry packet entry write: %v\n", r)
		}
	}()

	buffer := [MaxBillingEntryBytes]byte{}

	ws, err := encoding.CreateWriteStream(buffer[:])
	if err != nil {
		return nil, err
	}

	if err := message.Serialize(ws); err != nil {
		return nil, err
	}
	ws.Flush()

	return buffer[:ws.GetBytesProcessed()], nil
}

func ReadBillingEntry(entry *BillingEntry, data []byte) error {
	if err := message.Serialize(encoding.CreateReadStream(data)); err != nil {
		return err
	}

	return nil
}

func (entry *BillingEntry) Validate() bool {

	// always

	if message.Version < 0 {
		fmt.Printf("invalid version\n")
		return false
	}

	if message.Timestamp < 0 {
		fmt.Printf("invalid timestamp\n")
		return false
	}

	if message.SessionID == 0 {
		fmt.Printf("invalid session id\n")
		return false
	}

	if message.SliceNumber < 0 {
		fmt.Printf("invalid slice number\n")
		return false
	}

	if message.DirectMinRTT < 0 || message.DirectMinRTT > 1023 {
		fmt.Printf("invalid direct min rtt\n")
		return false
	}

	if message.DirectMaxRTT < 0 || message.DirectMaxRTT > 1023 {
		fmt.Printf("invalid direct max rtt\n")
		return false
	}

	if message.DirectPrimeRTT < 0 || message.DirectPrimeRTT > 1023 {
		fmt.Printf("invalid direct prime rtt\n")
		return false
	}

	if message.DirectJitter < 0 || message.DirectJitter > 255 {
		fmt.Printf("invalid direct jitter\n")
		return false
	}

	if message.DirectPacketLoss < 0 || message.DirectPacketLoss > 100 {
		fmt.Printf("invalid direct packet loss\n")
		return false
	}

	if message.RealPacketLoss < 0 || message.RealPacketLoss > 100 {
		if message.RealPacketLoss > 100 {
			fmt.Printf("RealPacketLoss %v > 100. Clamping to 100\n%+v\n", message.RealPacketLoss, entry)
			message.RealPacketLoss = 100
		} else {
			fmt.Printf("invalid real packet loss\n")
			return false
		}
	}

	if message.RealJitter < 0 || message.RealJitter > 1000 {
		if message.RealJitter > 1000 {
			fmt.Printf("RealJitter %v > 1000. Clamping to 1000\n%+v\n", message.RealJitter, entry)
			message.RealJitter = 100
		} else {
			fmt.Printf("invalid real jitter\n")
			return false
		}
	}

	if message.RouteDiversity < 0 || message.RouteDiversity > 32 {
		fmt.Printf("invalid route diversity\n")
		return false
	}

	// first slice only

	if message.SliceNumber == 0 {

		if message.BuyerID == 0 {
			fmt.Printf("invalid buyer id\n")
			return false
		}

		// IMPORTANT: Logic inverted because comparing a NaN float value always returns false
		if !(message.Latitude >= -90.0 && message.Latitude <= +90.0) {
			fmt.Printf("invalid latitude\n")
			return false
		}

		if !(message.Longitude >= -180.0 && message.Longitude <= +180.0) {
			fmt.Printf("invalid longitude\n")
			return false
		}

		// IMPORTANT: You must update this check if you ever add a new connection type in the SDK
		if message.ConnectionType < 0 || message.ConnectionType > 3 {
			fmt.Printf("invalid connection type\n")
			return false
		}

		// IMPORTANT: You must update this check when you add new platforms to the SDK
		if message.PlatformType < 0 || message.PlatformType > 10 {
			fmt.Printf("invalid platform type\n")
			return false
		}

		if message.NumTags < 0 || message.NumTags > 8 {
			fmt.Printf("invalid num tags\n")
			return false
		}
	}

	// summary slice only

	if message.Summary {

		if message.NumNearRelays < 0 || message.NumNearRelays > 32 {
			fmt.Printf("invalid num near relays\n")
			return false
		}

		if message.NumNearRelays > 0 {

			for i := 0; i < int(message.NumNearRelays); i++ {

				if message.NearRelayIDs[i] == 0 {
					// Log this but do not return false
					// TODO: investigate why nearRelayID is 0
					fmt.Printf("NearRelayIDs[%d] is 0.\n%+v\n", i, entry)
				}

				if message.NearRelayRTTs[i] < 0 || message.NearRelayRTTs[i] > 255 {
					fmt.Printf("invalid near relay rtt\n")
					return false
				}

				if message.NearRelayJitters[i] < 0 || message.NearRelayJitters[i] > 255 {
					fmt.Printf("invalid near relay jitter\n")
					return false
				}

				if message.NearRelayPacketLosses[i] < 0 || message.NearRelayPacketLosses[i] > 100 {
					fmt.Printf("invalid near relay packet loss\n")
					return false
				}
			}
		}
	}

	// network next only

	if message.Next {

		if message.NextRTT < 0 || message.NextRTT > 255 {
			fmt.Printf("invalid next rtt\n")
			return false
		}

		if message.NextJitter < 0 || message.NextJitter > 255 {
			fmt.Printf("invalid next jitter\n")
			return false
		}

		if message.NextPacketLoss < 0 || message.NextPacketLoss > 100 {
			fmt.Printf("invalid next packet loss\n")
			return false
		}

		if message.PredictedNextRTT < 0 || message.PredictedNextRTT > 255 {
			fmt.Printf("invalid predicted next rtt\n")
			return false
		}

		if message.NearRelayRTT < 0 || message.NearRelayRTT > 255 {
			fmt.Printf("invalid near relay rtt\n")
			return false
		}

		if message.NumNextRelays < 0 || message.NumNextRelays > 32 {
			fmt.Printf("invalid num next relays\n")
			return false
		}
	}

	return true
}

func (entry *BillingEntry) CheckNaNOrInf() (bool, []string) {

	var nanOrInfExists bool
	var nanOrInfFields []string

	if math.IsNaN(float64(message.Latitude)) || math.IsInf(float64(message.Latitude), 0) {
		nanOrInfFields = append(nanOrInfFields, "Latitude")
		nanOrInfExists = true
		message.Latitude = float32(0)
	}

	if math.IsNaN(float64(message.Longitude)) || math.IsInf(float64(message.Longitude), 0) {
		nanOrInfFields = append(nanOrInfFields, "Longitude")
		nanOrInfExists = true
		message.Longitude = float32(0)
	}

	return nanOrInfExists, nanOrInfFields
}

// To save bits during serialization, clamp integer and string fields if they go beyond the min
// or max values as defined in BillingEntry.Serialize()

func (entry *BillingEntry) ClampEntry() {

	// always

	if message.DirectMinRTT < 0 {
		core.Error("BillingEntry DirectMinRTT (%d) < 0. Clamping to 0.", message.DirectMinRTT)
		message.DirectMinRTT = 0
	}

	if message.DirectMinRTT > 1023 {
		core.Debug("BillingEntry DirectMinRTT (%d) > 1023. Clamping to 1023.", message.DirectMinRTT)
		message.DirectMinRTT = 1023
	}

	if message.DirectMaxRTT < 0 {
		core.Error("BillingEntry DirectMaxRTT (%d) < 0. Clamping to 0.", message.DirectMaxRTT)
		message.DirectMaxRTT = 0
	}

	if message.DirectMaxRTT > 1023 {
		core.Debug("BillingEntry DirectMaxRTT (%d) > 1023. Clamping to 1023.", message.DirectMaxRTT)
		message.DirectMaxRTT = 1023
	}

	if message.DirectPrimeRTT < 0 {
		core.Error("BillingEntry DirectPrimeRTT (%d) < 0. Clamping to 0.", message.DirectPrimeRTT)
		message.DirectPrimeRTT = 0
	}

	if message.DirectPrimeRTT > 1023 {
		core.Debug("BillingEntry DirectPrimeRTT (%d) > 1023. Clamping to 1023.", message.DirectPrimeRTT)
		message.DirectPrimeRTT = 1023
	}

	if message.DirectJitter < 0 {
		core.Error("BillingEntry DirectJitter (%d) < 0. Clamping to 0.", message.DirectJitter)
		message.DirectJitter = 0
	}

	if message.DirectJitter > 255 {
		core.Debug("BillingEntry DirectJitter (%d) > 255. Clamping to 255.", message.DirectJitter)
		message.DirectJitter = 255
	}

	if message.DirectPacketLoss < 0 {
		core.Error("BillingEntry DirectPacketLoss (%d) < 0. Clamping to 0.", message.DirectPacketLoss)
		message.DirectPacketLoss = 0
	}

	if message.DirectPacketLoss > 100 {
		core.Debug("BillingEntry DirectPacketLoss (%d) > 100. Clamping to 100.", message.DirectPacketLoss)
		message.DirectPacketLoss = 100
	}

	if message.RealPacketLoss < 0 {
		core.Error("BillingEntry RealPacketLoss (%d) < 0. Clamping to 0.", message.RealPacketLoss)
		message.RealPacketLoss = 0
	}

	if message.RealPacketLoss > 100 {
		core.Debug("BillingEntry RealPacketLoss (%d) > 100. Clamping to 100.", message.RealPacketLoss)
		message.RealPacketLoss = 100
	}

	if message.RealJitter > 1000 {
		core.Debug("BillingEntry RealJitter (%d) > 1000. Clamping to 1000.", message.RealJitter)
		message.RealJitter = uint32(1000)
	}

	if len(message.Debug) >= BillingEntryMaxDebugLength {
		core.Debug("BillingEntry Debug length (%d) >= BillingEntryMaxDebugLength (%d). Clamping to BillingEntryMaxDebugLength - 1 (%d)", len(message.Debug), BillingEntryMaxDebugLength, BillingEntryMaxDebugLength-1)
		message.Debug = message.Debug[:BillingEntryMaxDebugLength-1]
	}

	if message.RouteDiversity < 0 {
		core.Error("BillingEntry RouteDiversity (%d) < 0. Clamping to 0.", message.RouteDiversity)
		message.RouteDiversity = 0
	}

	if message.RouteDiversity > 32 {
		core.Debug("BillingEntry RouteDiversity (%d) > 32. Clamping to 32.", message.RouteDiversity)
		message.RouteDiversity = 32
	}

	// first slice and summary slice only

	if message.SliceNumber == 0 || message.Summary {

		if len(message.ClientAddress) >= BillingEntryMaxAddressLength {
			core.Debug("BillingEntry Client IP Address length (%d) >= BillingEntryMaxAddressLength (%d). Clamping to BillingEntryMaxAddressLength - 1 (%d)", len(message.ClientAddress), BillingEntryMaxAddressLength, BillingEntryMaxAddressLength-1)
			message.ClientAddress = message.ClientAddress[:BillingEntryMaxAddressLength-1]
		}

		if len(message.ISP) >= BillingEntryMaxISPLength {
			core.Debug("BillingEntry ISP length (%d) >= BillingEntryMaxISPLength (%d). Clamping to BillingEntryMaxISPLength - 1 (%d)", len(message.ISP), BillingEntryMaxISPLength, BillingEntryMaxISPLength-1)
			message.ISP = message.ISP[:BillingEntryMaxISPLength-1]
		}

		if message.ConnectionType < 0 {
			core.Error("BillingEntry ConnectionType (%d) < 0. Clamping to 0.", message.ConnectionType)
			message.ConnectionType = 0
		}

		// IMPORTANT: You must update this check if you ever add a new connection type in the SDK
		if message.ConnectionType > 3 {
			core.Debug("BillingEntry ConnectionType (%d) > 3. Clamping to 0 (unknown).", message.ConnectionType)
			message.ConnectionType = 0
		}

		if message.PlatformType < 0 {
			core.Error("BillingEntry PlatformType (%d) < 0. Clamping to 0.", message.PlatformType)
			message.PlatformType = 0
		}

		// IMPORTANT: You must update this check when you add new platforms to the SDK
		if message.PlatformType > 10 {
			core.Debug("BillingEntry PlatformType (%d) > 10. Clamping to 0 (unknown).", message.PlatformType)
			message.PlatformType = 0
		}

		if len(message.SDKVersion) >= BillingEntryMaxSDKVersionLength {
			core.Debug("BillingEntry SDKVersion length (%d) >= BillingEntryMaxSDKVersionLength (%d). Clamping to BillingEntryMaxSDKVersionLength - 1 (%d)", len(message.SDKVersion), BillingEntryMaxSDKVersionLength, BillingEntryMaxSDKVersionLength-1)
			message.SDKVersion = message.SDKVersion[:BillingEntryMaxSDKVersionLength-1]
		}

		if message.NumTags < 0 {
			core.Error("BillingEntry NumTags (%d) < 0. Clamping to 0.", message.NumTags)
			message.NumTags = 0
		}

		if message.NumTags > BillingEntryMaxTags {
			core.Debug("BillingEntry NumTags (%d) > BillingEntryMaxTags (%d). Clamping to BillingEntryMaxTags (%d).", message.NumTags, BillingEntryMaxTags, BillingEntryMaxTags)
			message.NumTags = BillingEntryMaxTags
		}

		if len(message.ServerAddress) >= BillingEntryMaxAddressLength {
			core.Debug("BillingEntry Server IP Address length (%d) >= BillingEntryMaxAddressLength (%d). Clamping to BillingEntryMaxAddressLength - 1 (%d)", len(message.ServerAddress), BillingEntryMaxAddressLength, BillingEntryMaxAddressLength-1)
			message.ServerAddress = message.ServerAddress[:BillingEntryMaxAddressLength-1]
		}
	}

	// summary slice only

	if message.Summary {

		if message.NumNearRelays < 0 {
			core.Error("BillingEntry NumNearRelays (%d) < 0. Clamping to 0.", message.NumNearRelays)
			message.NumNearRelays = 0
		}

		if message.NumNearRelays > BillingEntryMaxNearRelays {
			core.Debug("BillingEntry NumNearRelays (%d) > BillingEntryMaxNearRelays (%d). Clamping to BillingEntryMaxNearRelays (%d).", message.NumNearRelays, BillingEntryMaxNearRelays, BillingEntryMaxNearRelays)
			message.NumNearRelays = BillingEntryMaxNearRelays
		}

		for i := 0; i < int(message.NumNearRelays); i++ {
			if message.NearRelayRTTs[i] < 0 {
				core.Error("BillingEntry NearRelayRTT[%d] (%d) < 0. Clamping to 0.", i, message.NearRelayRTTs[i])
				message.NearRelayRTTs[i] = 0
			}

			if message.NearRelayRTTs[i] > 255 {
				core.Debug("BillingEntry NearRelayRTTs[%d] (%d) > 255. Clamping to 255.", i, message.NearRelayRTTs[i])
				message.NearRelayRTTs[i] = 255
			}

			if message.NearRelayJitters[i] < 0 {
				core.Error("BillingEntry NearRelayRTT[%d] (%d) < 0. Clamping to 0.", i, message.NearRelayJitters[i])
				message.NearRelayJitters[i] = 0
			}

			if message.NearRelayJitters[i] > 255 {
				core.Debug("BillingEntry NearRelayJitters[%d] (%d) > 255. Clamping to 255.", i, message.NearRelayJitters[i])
				message.NearRelayJitters[i] = 255
			}

			if message.NearRelayPacketLosses[i] < 0 {
				core.Error("BillingEntry NearRelayRTT[%d] (%d) < 0. Clamping to 0.", i, message.NearRelayPacketLosses[i])
				message.NearRelayPacketLosses[i] = 0
			}

			if message.NearRelayPacketLosses[i] > 100 {
				core.Debug("BillingEntry NearRelayPacketLosses[%d] (%d) > 100. Clamping to 100.", i, message.NearRelayPacketLosses[i])
				message.NearRelayPacketLosses[i] = 100
			}
		}
	}

	// network next only

	if message.Next {

		if message.NextRTT < 0 {
			core.Error("BillingEntry NextRTT (%d) < 0. Clamping to 0.", message.NextRTT)
			message.NextRTT = 0
		}

		if message.NextRTT > 255 {
			core.Debug("BillingEntry NextRTT (%d) > 255. Clamping to 255.", message.NextRTT)
			message.NextRTT = 255
		}

		if message.NextJitter < 0 {
			core.Error("BillingEntry NextJitter (%d) < 0. Clamping to 0.", message.NextJitter)
			message.NextJitter = 0
		}

		if message.NextJitter > 255 {
			core.Debug("BillingEntry NextJitter (%d) > 255. Clamping to 255.", message.NextJitter)
			message.NextJitter = 255
		}

		if message.NextPacketLoss < 0 {
			core.Error("BillingEntry NextPacketLoss (%d) < 0. Clamping to 0.", message.NextPacketLoss)
			message.NextPacketLoss = 0
		}

		if message.NextPacketLoss > 100 {
			core.Debug("BillingEntry NextPacketLoss (%d) > 100. Clamping to 100.", message.NextPacketLoss)
			message.NextPacketLoss = 100
		}

		if message.PredictedNextRTT < 0 {
			core.Error("BillingEntry PredictedNextRTT (%d) < 0. Clamping to 0.", message.PredictedNextRTT)
			message.PredictedNextRTT = 0
		}

		if message.PredictedNextRTT > 255 {
			core.Debug("BillingEntry PredictedNextRTT (%d) > 255. Clamping to 255.", message.PredictedNextRTT)
			message.PredictedNextRTT = 255
		}

		if message.NearRelayRTT < 0 {
			core.Error("BillingEntry NearRelayRTT (%d) < 0. Clamping to 0.", message.NearRelayRTT)
			message.NearRelayRTT = 0
		}

		if message.NearRelayRTT > 255 {
			core.Debug("BillingEntry NearRelayRTT (%d) > 255. Clamping to 255.", message.NearRelayRTT)
			message.NearRelayRTT = 255
		}

		if message.NumNextRelays < 0 {
			core.Error("BillingEntry NumNextRelays (%d) < 0. Clamping to 0.", message.NumNextRelays)
			message.NumNextRelays = 0
		}

		if message.NumNextRelays > BillingEntryMaxRelays {
			core.Debug("BillingEntry NumNextRelays (%d) > BillingEntryMaxRelays (%d). Clamping to BillingEntryMaxRelays (%d).", message.NumNextRelays, BillingEntryMaxRelays, BillingEntryMaxRelays)
			message.NumNextRelays = BillingEntryMaxRelays
		}
	}
}
*/

func (message *BillingMessage) Save() (map[string]bigquery.Value, string, error) {

	e := make(map[string]bigquery.Value)

	/*
		1. Always

		These values are written for every slice.
	*/

	e["timestamp"] = int(message.Timestamp)
	e["sessionID"] = int(message.SessionID)
	e["sliceNumber"] = int(message.SliceNumber)
	e["directRTT"] = int(message.DirectMinRTT) // NOTE: directRTT refers to DirectMinRTT as of version 7
	e["directMaxRTT"] = int(message.DirectMaxRTT)
	e["directPrimeRTT"] = int(message.DirectPrimeRTT)
	e["directJitter"] = int(message.DirectJitter)
	e["directPacketLoss"] = int(message.DirectPacketLoss)
	e["realPacketLoss"] = float64(message.RealPacketLoss) + float64(message.RealPacketLoss_Frac)/256.0
	e["realJitter"] = int(message.RealJitter)

	if message.Next {
		e["next"] = true
	}

	if message.Flagged {
		e["flagged"] = true
	}

	if message.Summary {
		e["summary"] = true
	}

	if message.UseDebug {
		e["debug"] = message.Debug
	}

	if message.RouteDiversity > 0 {
		e["routeDiversity"] = int(message.RouteDiversity)
	}

	if message.UserFlags > 0 {
		e["userFlags"] = int(message.UserFlags)
	}

	if message.TryBeforeYouBuy {
		e["tryBeforeYouBuy"] = message.TryBeforeYouBuy
	}

	/*
		2. First slice and summary slice only

		These values are serialized only for slice 0 and the summary slice.
	*/

	if message.SliceNumber == 0 || message.Summary {

		e["datacenterID"] = int(message.DatacenterID)
		e["buyerID"] = int(message.BuyerID)
		e["userHash"] = int(message.UserHash)
		e["envelopeBytesUp"] = int(message.EnvelopeBytesUp)
		e["envelopeBytesDown"] = int(message.EnvelopeBytesDown)
		e["latitude"] = message.Latitude
		e["longitude"] = message.Longitude
		e["clientAddress"] = message.ClientAddress
		e["serverAddress"] = message.ServerAddress
		e["isp"] = message.ISP
		e["connectionType"] = int(message.ConnectionType)
		e["platformType"] = int(message.PlatformType)
		e["sdkVersion"] = message.SDKVersion

		if message.NumTags > 0 {
			tags := make([]bigquery.Value, message.NumTags)
			for i := 0; i < int(message.NumTags); i++ {
				tags[i] = int(message.Tags[i])
			}
			e["tags"] = tags
		}

		if message.ABTest {
			e["abTest"] = true
		}

		if message.Pro {
			e["pro"] = true
		}
	}

	/*
		3. Summary slice only

		These values are serialized only for the summary slice (at the end of the session)
	*/

	if message.Summary {

		e["clientToServerPacketsSent"] = int(message.ClientToServerPacketsSent)
		e["serverToClientPacketsSent"] = int(message.ServerToClientPacketsSent)
		e["clientToServerPacketsLost"] = int(message.ClientToServerPacketsLost)
		e["serverToClientPacketsLost"] = int(message.ServerToClientPacketsLost)
		e["clientToServerPacketsOutOfOrder"] = int(message.ClientToServerPacketsOutOfOrder)
		e["serverToClientPacketsOutOfOrder"] = int(message.ServerToClientPacketsOutOfOrder)

		if message.NumNearRelays > 0 {

			nearRelayIDs := make([]bigquery.Value, message.NumNearRelays)
			nearRelayRTTs := make([]bigquery.Value, message.NumNearRelays)
			nearRelayJitters := make([]bigquery.Value, message.NumNearRelays)
			nearRelayPacketLosses := make([]bigquery.Value, message.NumNearRelays)

			for i := 0; i < int(message.NumNearRelays); i++ {
				nearRelayIDs[i] = int(message.NearRelayIDs[i])
				nearRelayRTTs[i] = int(message.NearRelayRTTs[i])
				nearRelayJitters[i] = int(message.NearRelayJitters[i])
				nearRelayPacketLosses[i] = int(message.NearRelayPacketLosses[i])
			}

			e["nearRelayIDs"] = nearRelayIDs
			e["nearRelayRTTs"] = nearRelayRTTs
			e["nearRelayJitters"] = nearRelayJitters
			e["nearRelayPacketLosses"] = nearRelayPacketLosses

		}

		if message.EverOnNext {
			e["everOnNext"] = true
		}

		e["sessionDuration"] = int(message.SessionDuration)

		if message.EverOnNext {
			e["totalPriceSum"] = int(message.TotalPriceSum)
			e["envelopeBytesUpSum"] = int(message.EnvelopeBytesUpSum)
			e["envelopeBytesDownSum"] = int(message.EnvelopeBytesDownSum)
			e["durationOnNext"] = int(message.DurationOnNext)
		}

		e["startTimestamp"] = int(message.StartTimestamp)

	}

	/*
		4. Network Next Only

		These values are serialized only when a slice is on network next.
	*/

	if message.Next {

		e["nextRTT"] = int(message.NextRTT)
		e["nextJitter"] = int(message.NextJitter)
		e["nextPacketLoss"] = int(message.NextPacketLoss)
		e["predictedNextRTT"] = int(message.PredictedNextRTT)
		e["nearRelayRTT"] = int(message.NearRelayRTT)

		if message.NumNextRelays > 0 {

			nextRelays := make([]bigquery.Value, message.NumNextRelays)
			nextRelayPrice := make([]bigquery.Value, message.NumNextRelays)

			for i := 0; i < int(message.NumNextRelays); i++ {
				nextRelays[i] = int(message.NextRelays[i])
				nextRelayPrice[i] = int(message.NextRelayPrice[i])
			}

			e["nextRelays"] = nextRelays
			e["nextRelayPrice"] = nextRelayPrice
		}

		e["totalPrice"] = int(message.TotalPrice)

		if message.Uncommitted {
			e["uncommitted"] = true
		}

		if message.Multipath {
			e["multipath"] = true
		}

		if message.RTTReduction {
			e["rttReduction"] = true
		}

		if message.PacketLossReduction {
			e["packetLossReduction"] = true
		}

		if message.RouteChanged {
			e["routeChanged"] = true
		}

		e["nextBytesUp"] = int(message.NextBytesUp)
		e["nextBytesDown"] = int(message.NextBytesDown)

	}

	/*
		5. Error State Only

		These values are only serialized when the session is in an error state (rare).
	*/

	if message.FallbackToDirect {
		e["fallbackToDirect"] = true
	}

	if message.MultipathVetoed {
		e["multipathVetoed"] = true
	}

	if message.Mispredicted {
		e["mispredicted"] = true
	}

	if message.Vetoed {
		e["vetoed"] = true
	}

	if message.LatencyWorse {
		e["latencyWorse"] = true
	}

	if message.NoRoute {
		e["noRoute"] = true
	}

	if message.NextLatencyTooHigh {
		e["nextLatencyTooHigh"] = true
	}

	if message.CommitVeto {
		e["commitVeto"] = true
	}

	if message.UnknownDatacenter {
		e["unknownDatacenter"] = true
	}

	if message.DatacenterNotEnabled {
		e["datacenterNotEnabled"] = true
	}

	if message.BuyerNotLive {
		e["buyerNotLive"] = true
	}

	if message.StaleRouteMatrix {
		e["staleRouteMatrix"] = true
	}

	return e, "", nil
}
