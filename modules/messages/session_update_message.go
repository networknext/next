package messages

import (
	"fmt"
	// "net"

	// "cloud.google.com/go/bigquery"

	// "github.com/networknext/backend/modules/common"
	// "github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/encoding"
)

const (
	SessionUpdateMessageVersion_Min   = 1
	SessionUpdateMessageVersion_Max   = 1
	SessionUpdateMessageVersion_Write = 1

	MaxSessionUpdateMessageBytes = 4096

	// todo: constants module
	SessionUpdateMessageMaxRelays           = 5
	SessionUpdateMessageMaxAddressLength    = 256
	SessionUpdateMessageMaxISPLength        = 64
	SessionUpdateMessageMaxSDKVersionLength = 11
	SessionUpdateMessageMaxDebugLength      = 2048
	SessionUpdateMessageMaxTags             = 8
	SessionUpdateMessageMaxRTT              = 1023
	SessionUpdateMessageMaxJitter           = 255
	SessionUpdateMessageMaxPacketLoss       = 100
	SessionUpdateMessageMaxRouteDiversity   = 31
	SessionUpdateMessageMaxConnectionType   = 3
	SessionUpdateMessageMaxPlatformType     = 10
	SessionUpdateMessageMaxRouteRelays      = 5
	SessionUpdateMessageMaxNearRelayRTT     = 255

	SessionFlags_Next                 = (1<<0)
	SessionFlags_Reported             = (1<<1)
	SessionFlags_Debug                = (1<<2)
	SessionFlags_FallbackToDirect     = (1<<3)
	SessionFlags_Mispredict           = (1<<4)
	SessionFlags_LatencyWorse         = (1<<5)
	SessionFlags_NoRoute              = (1<<6)
	SessionFlags_NextLatencyTooHigh   = (1<<7)
	SessionFlags_UnknownDatacenter    = (1<<8)
	SessionFlags_DatacenterNotEnabled = (1<<9)
	SessionFlags_StaleRouteMatrix     = (1<<10)
	SessionFlags_ABTest               = (1<<11)
	SessionFlags_Aborted              = (1<<12)
	SessionFlags_LatencyReduction     = (1<<13)
	SessionFlags_PacketLossReduction  = (1<<14)
	SessionFlags_EverOnNext           = (1<<15)
	SessionFlags_Summary              = (1<<16)
)

type SessionUpdateMessage struct {

	// always

	Version             uint8
	Timestamp           uint64
	SessionId           uint64
	SliceNumber         uint32
	RealPacketLoss      float32
	RealJitter          float32
	RealOutOfOrder      float32
	SessionFlags        uint64
	GameEvents          uint64
	DirectRTT           float32
	DirectJitter        float32
	DirectPacketLoss    float32
	DirectBytesUp       uint64
	DirectBytesDown     uint64

	// next only

	NextRTT             float32
	NextJitter          float32
	NextPacketLoss      float32
	NextPredictedRTT    float32
	NextNearRelayRTT    float32
	NextNumRouteRelays  uint32
	NextRouteRelays     [SessionUpdateMessageMaxRouteRelays]uint64
	NextBytesUp         uint64
	NextBytesDown       uint64

/*
	// first slice and summary slice only

	DatacenterId      uint64
	BuyerId           uint64
	UserHash          uint64
	EnvelopeBytesUp   uint64
	EnvelopeBytesDown uint64
	Latitude          float32
	Longitude         float32
	ClientAddress     net.UDPAddr
	ServerAddress     net.UDPAddr
	ISP               string
	ConnectionType    int32
	PlatformType      int32
	SDKVersion_Major  uint32
	SDKVersion_Minor  uint32
	SDKVersion_Patch  uint32
	NumTags           int32
	Tags              [SessionUpdateMessageMaxTags]uint64

	// summary slice only

	ClientToServerPacketsSent       uint64
	ServerToClientPacketsSent       uint64
	ClientToServerPacketsLost       uint64
	ServerToClientPacketsLost       uint64
	ClientToServerPacketsOutOfOrder uint64
	ServerToClientPacketsOutOfOrder uint64
	SessionDuration                 uint32
	EnvelopeBytesUpSum              uint64
	EnvelopeBytesDownSum            uint64
	DurationOnNext                  uint32
	StartTimestamp                  uint64
*/
}

func (message *SessionUpdateMessage) Write(buffer []byte) []byte {

	index := 0

	if message.Version < PortalMessageVersion_Min || message.Version > PortalMessageVersion_Max {
		panic(fmt.Sprintf("invalid portal message version %d", message.Version))
	}

	encoding.WriteUint8(buffer, &index, message.Version)

	// ...

	return buffer[:index]
}

func (message *SessionUpdateMessage) Read(buffer []byte) error {

	index := 0

	if !encoding.ReadUint8(buffer, &index, &message.Version) {
		return fmt.Errorf("failed to read session update message version")
	}

	if message.Version < SessionUpdateMessageVersion_Min || message.Version > SessionUpdateMessageVersion_Max {
		return fmt.Errorf("invalid session update message version %d", message.Version)
	}

	// ...

	return nil
}


/*
func (message *SessionUpdateMessage) Serialize(stream encoding.Stream) error {

	// always

	if stream.IsWriting() && (message.Version < SessionUpdateMessageVersion_Min || message.Version > SessionUpdateMessageVersion_Max) {
		panic(fmt.Sprintf("invalid session update message version %d", message.Version))
	}

	stream.SerializeBits(&message.Version, 8)

	if stream.IsReading() && (message.Version < SessionUpdateMessageVersion_Min || message.Version > SessionUpdateMessageVersion_Max) {
		return fmt.Errorf("invalid session update message version %d", message.Version)
	}

	stream.SerializeUint64(&message.Timestamp)
	stream.SerializeUint64(&message.SessionId)

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

	stream.SerializeInteger(&message.DirectMinRTT, 0, SessionUpdateMessageMaxRTT)
	stream.SerializeInteger(&message.DirectMaxRTT, 0, SessionUpdateMessageMaxRTT)
	stream.SerializeInteger(&message.DirectPrimeRTT, 0, SessionUpdateMessageMaxRTT)
	stream.SerializeInteger(&message.DirectJitter, 0, SessionUpdateMessageMaxJitter)
	stream.SerializeInteger(&message.DirectPacketLoss, 0, SessionUpdateMessageMaxPacketLoss)

	stream.SerializeInteger(&message.RealPacketLoss, 0, SessionUpdateMessageMaxPacketLoss)
	stream.SerializeBits(&message.RealPacketLoss_Frac, 8)
	stream.SerializeUint32(&message.RealJitter)

	stream.SerializeBool(&message.Next)
	stream.SerializeBool(&message.Flagged)
	stream.SerializeBool(&message.Summary)

	stream.SerializeBool(&message.UseDebug)
	stream.SerializeString(&message.Debug, SessionUpdateMessageMaxDebugLength)

	stream.SerializeInteger(&message.RouteDiversity, 0, SessionUpdateMessageMaxRouteDiversity)

	stream.SerializeUint64(&message.UserFlags)

	stream.SerializeBool(&message.TryBeforeYouBuy)

	// first slice and summary slice only

	if message.SliceNumber == 0 || message.Summary {

		stream.SerializeUint64(&message.DatacenterId)
		stream.SerializeUint64(&message.BuyerId)
		stream.SerializeUint64(&message.UserHash)
		stream.SerializeUint64(&message.EnvelopeBytesUp)
		stream.SerializeUint64(&message.EnvelopeBytesDown)
		stream.SerializeFloat32(&message.Latitude)
		stream.SerializeFloat32(&message.Longitude)
		stream.SerializeString(&message.ISP, SessionUpdateMessageMaxISPLength)
		stream.SerializeInteger(&message.ConnectionType, 0, SessionUpdateMessageMaxConnectionType)
		stream.SerializeInteger(&message.PlatformType, 0, SessionUpdateMessageMaxPlatformType)
		stream.SerializeBits(&message.SDKVersion_Major, 8)
		stream.SerializeBits(&message.SDKVersion_Minor, 8)
		stream.SerializeBits(&message.SDKVersion_Patch, 8)
		stream.SerializeInteger(&message.NumTags, 0, SessionUpdateMessageMaxTags)
		for i := 0; i < int(message.NumTags); i++ {
			stream.SerializeUint64(&message.Tags[i])
		}
		stream.SerializeBool(&message.ABTest)
		stream.SerializeBool(&message.Pro)
		stream.SerializeAddress(&message.ClientAddress)
		stream.SerializeAddress(&message.ServerAddress)
	}

	// summary slice only

	if message.Summary {

		stream.SerializeUint64(&message.ClientToServerPacketsSent)
		stream.SerializeUint64(&message.ServerToClientPacketsSent)
		stream.SerializeUint64(&message.ClientToServerPacketsLost)
		stream.SerializeUint64(&message.ServerToClientPacketsLost)
		stream.SerializeUint64(&message.ClientToServerPacketsOutOfOrder)
		stream.SerializeUint64(&message.ServerToClientPacketsOutOfOrder)

		stream.SerializeUint64(&message.StartTimestamp)
		stream.SerializeUint32(&message.SessionDuration)

		stream.SerializeBool(&message.EverOnNext)

		if message.EverOnNext {

			stream.SerializeUint64(&message.EnvelopeBytesUpSum)
			stream.SerializeUint64(&message.EnvelopeBytesDownSum)
			stream.SerializeUint32(&message.DurationOnNext)
		}
	}

	// next only

	if message.Next {

		stream.SerializeInteger(&message.NextRTT, 0, SessionUpdateMessageMaxRTT)
		stream.SerializeInteger(&message.NextJitter, 0, SessionUpdateMessageMaxJitter)
		stream.SerializeInteger(&message.NextPacketLoss, 0, SessionUpdateMessageMaxPacketLoss)
		stream.SerializeInteger(&message.PredictedNextRTT, 0, SessionUpdateMessageMaxRTT)
		stream.SerializeInteger(&message.NearRelayRTT, 0, SessionUpdateMessageMaxNearRelayRTT)

		stream.SerializeInteger(&message.NumNextRelays, 0, SessionUpdateMessageMaxRelays)
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

		stream.SerializeUint64(&message.NextBytesUp)
		stream.SerializeUint64(&message.NextBytesDown)
	}

	// error state only

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

	return stream.Error()
}

func (message *SessionUpdateMessage) Read(buffer []byte) error {

	readStream := encoding.CreateReadStream(buffer)

	return message.Serialize(readStream)
}

func (message *SessionUpdateMessage) Write(buffer []byte) []byte {

	writeStream := encoding.CreateWriteStream(buffer[:])

	err := message.Serialize(writeStream)
	if err != nil {
		panic(err)
	}

	writeStream.Flush()

	packetBytes := writeStream.GetBytesProcessed()

	return buffer[:packetBytes]
}

func (message *SessionUpdateMessage) Save() (map[string]bigquery.Value, string, error) {

	e := make(map[string]bigquery.Value)

	// always

	e["timestamp"] = int(message.Timestamp)
	e["sessionID"] = int(message.SessionId)
	e["sliceNumber"] = int(message.SliceNumber)
	e["directRTT"] = int(message.DirectMinRTT)
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

	// first slice and summary slice only

	if message.SliceNumber == 0 || message.Summary {

		e["datacenterID"] = int(message.DatacenterId)
		e["buyerID"] = int(message.BuyerId)
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
		e["sdkVersion"] = fmt.Sprintf("%d.%d.%d", message.SDKVersion_Major, message.SDKVersion_Minor, message.SDKVersion_Patch)

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

	// summary slice only

	if message.Summary {

		e["clientToServerPacketsSent"] = int(message.ClientToServerPacketsSent)
		e["serverToClientPacketsSent"] = int(message.ServerToClientPacketsSent)
		e["clientToServerPacketsLost"] = int(message.ClientToServerPacketsLost)
		e["serverToClientPacketsLost"] = int(message.ServerToClientPacketsLost)
		e["clientToServerPacketsOutOfOrder"] = int(message.ClientToServerPacketsOutOfOrder)
		e["serverToClientPacketsOutOfOrder"] = int(message.ServerToClientPacketsOutOfOrder)

		if message.EverOnNext {
			e["everOnNext"] = true
		}

		e["sessionDuration"] = int(message.SessionDuration)

		if message.EverOnNext {
			e["envelopeBytesUpSum"] = int(message.EnvelopeBytesUpSum)
			e["envelopeBytesDownSum"] = int(message.EnvelopeBytesDownSum)
			e["durationOnNext"] = int(message.DurationOnNext)
		}

		e["startTimestamp"] = int(message.StartTimestamp)
	}

	// next only

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

	// error state only
	
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

func (message *SessionUpdateMessage) Clamp() {

	if common.Clamp(&message.DirectMinRTT, 0, SessionUpdateMessageMaxRTT) {
		core.Warn("DirectMinRTT was clamped!")
	}

	if common.Clamp(&message.DirectMaxRTT, 0, SessionUpdateMessageMaxRTT) {
		core.Warn("DirectMaxRTT was clamped!")
	}

	if common.Clamp(&message.DirectPrimeRTT, 0, SessionUpdateMessageMaxRTT) {
		core.Warn("DirectPrimeRTT was clamped!")
	}

	if common.Clamp(&message.DirectJitter, 0, SessionUpdateMessageMaxJitter) {
		core.Warn("DirectMinRTT was clamped!")
	}

	if common.Clamp(&message.DirectPacketLoss, 0, SessionUpdateMessageMaxPacketLoss) {
		core.Warn("DirectPacketLoss was clamped!")
	}

	if common.Clamp(&message.RealPacketLoss, 0, SessionUpdateMessageMaxPacketLoss) {
		core.Warn("RealPacketLoss was clamped!")
	}

	if common.Clamp(&message.RealPacketLoss_Frac, 0, 255) {
		core.Warn("RealPacketLoss_Frac was clamped!")
	}

	if common.Clamp(&message.RealJitter, 0, SessionUpdateMessageMaxJitter) {
		core.Warn("RealPacketJitter was clamped!")
	}

	if common.ClampString(&message.Debug, SessionUpdateMessageMaxDebugLength) {
		core.Warn("Debug was clamped!")
	}

	if common.Clamp(&message.RouteDiversity, 0, SessionUpdateMessageMaxRouteDiversity) {
		core.Warn("RouteDiversity was clamped!")
	}

	if common.ClampString(&message.ISP, SessionUpdateMessageMaxISPLength) {
		core.Warn("ISP was clamped!")
	}

	if common.Clamp(&message.ConnectionType, 0, SessionUpdateMessageMaxConnectionType) {
		core.Warn("RealPacketLoss was clamped!")
	}

	if common.Clamp(&message.PlatformType, 0, SessionUpdateMessageMaxPlatformType) {
		core.Warn("PlatformType was clamped!")
	}

	if common.Clamp(&message.NumTags, 0, SessionUpdateMessageMaxTags) {
		core.Warn("NumTags was clamped!")
	}

	if common.Clamp(&message.NextRTT, 0, SessionUpdateMessageMaxRTT) {
		core.Warn("NextRTT was clamped!")
	}

	if common.Clamp(&message.NextJitter, 0, SessionUpdateMessageMaxJitter) {
		core.Warn("NextJitter was clamped!")
	}

	if common.Clamp(&message.NextPacketLoss, 0, SessionUpdateMessageMaxPacketLoss) {
		core.Warn("NextPacketLoss was clamped!")
	}

	if common.Clamp(&message.PredictedNextRTT, 0, SessionUpdateMessageMaxRTT) {
		core.Warn("PredictedNextRTT was clamped!")
	}

	if common.Clamp(&message.NearRelayRTT, 0, SessionUpdateMessageMaxNearRelayRTT) {
		core.Warn("NearRelayRTT was clamped!")
	}

	if common.Clamp(&message.NumNextRelays, 0, SessionUpdateMessageMaxRelays) {
		core.Warn("NumNextRelays was clamped!")
	}
}
*/