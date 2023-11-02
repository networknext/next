package handlers

import (
	"fmt"
	"math"
	"net"
	"time"

	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/constants"
	"github.com/networknext/next/modules/core"
	"github.com/networknext/next/modules/crypto"
	db "github.com/networknext/next/modules/database"
	"github.com/networknext/next/modules/encoding"
	"github.com/networknext/next/modules/messages"
	"github.com/networknext/next/modules/packets"
)

type SessionUpdateState struct {

	/*
		Convenience state struct for the session update handler.

		We put all the state in here so it's easy to call out to functions to do work.

		Otherwise we have to pass a million parameters into every function and it gets old fast.
	*/

	PingKey                 []byte
	RelayBackendPrivateKey  []byte
	ServerBackendAddress    *net.UDPAddr
	ServerBackendPrivateKey []byte
	ServerBackendPublicKey  []byte

	LocateIP func(ip net.IP) (float32, float32)

	From *net.UDPAddr

	Input packets.SDK_SessionData // sent up from the SDK. previous slice.

	Output packets.SDK_SessionData // sent down to the SDK. current slice.

	ResponsePacket []byte // response packet sent back to the "from" if non-zero length.

	Request       *packets.SDK_SessionUpdateRequestPacket
	Response      packets.SDK_SessionUpdateResponsePacket
	Database      *db.Database
	RouteMatrix   *common.RouteMatrix
	Datacenter    *db.Datacenter
	BuyerId       uint64
	Buyer         *db.Buyer
	Debug         *string
	StaleDuration time.Duration

	RealPacketLoss float32
	RealJitter     float32
	RealOutOfOrder float32

	// route diversity is the number of unique near relays with viable routes
	RouteDiversity int32

	// for route planning
	DestRelayIds   []uint64
	DestRelays     []int32
	SourceRelays   []int32
	SourceRelayRTT []int32

	// error flags for this update
	Error uint64

	// lat/long if we looked it up this update
	Latitude  float32
	Longitude float32

	// track start time of handler
	StartTimestamp     uint64
	StartTimestampNano uint64

	// true if we fellback to direct on this update
	FallbackToDirect bool

	// if true, only network next sessions are sent to portal
	PortalNextSessionsOnly bool

	// codepath flags (for unit testing etc...)
	ClientPingTimedOut                        bool
	AnalysisOnly                              bool
	RouteChanged                              bool
	RouteContinued                            bool
	TakeNetworkNext                           bool
	StayDirect                                bool
	FirstUpdate                               bool
	ReadSessionData                           bool
	NotGettingNearRelaysDatacenterIsNil       bool
	NotGettingNearRelaysAnalysisOnly          bool
	NotGettingNearRelaysDatacenterNotEnabled  bool
	NotUpdatingNearRelaysAnalysisOnly         bool
	NotUpdatingNearRelaysDatacenterNotEnabled bool
	SentPortalSessionUpdateMessage            bool
	SentPortalNearRelayUpdateMessage          bool
	SentAnalyticsNearRelayPingMessage         bool
	SentAnalyticsSessionUpdateMessage         bool
	SentAnalyticsSessionSummaryMessage        bool
	LocatedIP                                 bool
	GetNearRelays                             bool
	WroteResponsePacket                       bool
	LongSessionUpdate                         bool

	FallbackToDirectChannel chan<- uint64

	PortalSessionUpdateMessageChannel   chan<- *messages.PortalSessionUpdateMessage
	PortalNearRelayUpdateMessageChannel chan<- *messages.PortalNearRelayUpdateMessage

	AnalyticsSessionUpdateMessageChannel  chan<- *messages.AnalyticsSessionUpdateMessage
	AnalyticsSessionSummaryMessageChannel chan<- *messages.AnalyticsSessionSummaryMessage
	AnalyticsNearRelayPingMessageChannel  chan<- *messages.AnalyticsNearRelayPingMessage
}

func SessionUpdate_ReadSessionData(state *SessionUpdateState) bool {

	if state.ReadSessionData {
		return true
	}

	if !crypto.Verify(state.Request.SessionData[:state.Request.SessionDataBytes], state.ServerBackendPublicKey[:], state.Request.SessionDataSignature[:]) {
		core.Error("session data signature check failed")
		state.Error |= constants.SessionError_SessionDataSignatureCheckFailed
		return false
	}

	readStream := encoding.CreateReadStream(state.Request.SessionData[:])

	err := state.Input.Serialize(readStream)
	if err != nil {
		core.Debug("failed to read session data: %v", err)
		state.Error |= constants.SessionError_FailedToReadSessionData
		return false
	}

	state.ReadSessionData = true

	return true
}

func SessionUpdate_Pre(state *SessionUpdateState) bool {

	state.StartTimestampNano = uint64(time.Now().UnixNano())

	state.StartTimestamp = state.StartTimestampNano / 1000000000	// nano -> seconds

	/*
		Read session data first

		We always need to read this, because we have to process and return it in the output.
	*/

	if state.Request.SliceNumber != 0 && !SessionUpdate_ReadSessionData(state) {
		return true
	}

	/*
		Update scores

		We track the best score seen per-session to keep scores (and session ordering in portal) stable.

		The lowest scores are the best scores, so we check if the new score is lower than the current best score.
	*/

	score := core.GetSessionScore(state.Request.Next, int32(state.Request.DirectRTT), int32(state.Request.NextRTT))

	if state.Request.SliceNumber >= 1 {
		if uint32(score) < state.Input.BestScore {
			state.Input.BestScore = uint32(score)
			state.Input.BestDirectRTT = uint32(state.Request.DirectRTT)
			state.Input.BestNextRTT = uint32(state.Request.NextRTT)
			if state.Input.BestDirectRTT > 1000 {
				state.Input.BestDirectRTT = 1000
			}
			if state.Input.BestNextRTT > 1000 {
				state.Input.BestNextRTT = 1000
			}
		}
	} else {
		state.Input.BestScore = uint32(999)
	}

	/*
		Fallback to direct is a state where the SDK has met some fatal error condition.

		When this happens, the session will go direct from this point forward.
	*/

	if state.Request.FallbackToDirect {
		if (state.Error & constants.SessionError_FallbackToDirect) == 0 {
			core.Warn("fallback to direct [%016x]", state.Request.SessionId)
			state.Error |= constants.SessionError_FallbackToDirect
			state.FallbackToDirect = true
			if state.FallbackToDirectChannel != nil {
				state.FallbackToDirectChannel <- state.Request.SessionId
			}
		}
		return true
	}

	/*
		If the route shader is in analysis only mode, set the analysis only flag in the state

		We don't acceleration sessions in analysis only mode.
	*/

	if state.Buyer.RouteShader.AnalysisOnly {
		core.Debug("analysis only")
		state.AnalysisOnly = true
		return true
	}

	/*
		If the client ping has timed out, don't do any further processing.
	*/

	if state.Request.ClientPingTimedOut {
		core.Debug("client ping timed out")
		state.ClientPingTimedOut = true
		return true
	}

	/*
		On the initial slice, we look up the lat/long for the player using ip2location.

		On subsequent slices, we use the cached location data from the session state.
	*/

	if state.Request.SliceNumber == 0 {

		state.LocatedIP = true

		state.FirstUpdate = true

		state.Latitude, state.Longitude = state.LocateIP(state.Request.ClientAddress.IP)

		if state.Latitude == 0.0 && state.Longitude == 0.0 {
			core.Error("location veto")
			state.Input.RouteState.LocationVeto = true
			return true
		}
	}

	/*
		Routing with an old route matrix runs a serious risk of sending players across routes that are WORSE
		than their default internet route, so it's best to just go direct if the route matrix is stale.
	*/

	if state.RouteMatrix.CreatedAt+uint64(state.StaleDuration.Seconds()) < uint64(time.Now().Unix()) {
		core.Debug("stale route matrix")
		state.Error |= constants.SessionError_StaleRouteMatrix
		return true
	}

	/*
		Check if the datacenter is unknown, and flag it.

		This is important so that we can quickly see when we need to add new datacenters for buyers.
	*/

	if !state.Database.DatacenterExists(state.Request.DatacenterId) {
		core.Debug("unknown datacenter")
		state.Error |= constants.SessionError_UnknownDatacenter
	}

	/*
		Check if the datacenter is enabled for this buyer.

		If the datacenter is not enabled, we just wont accelerate the player.
	*/

	if !state.Database.DatacenterEnabled(state.Request.BuyerId, state.Request.DatacenterId) {
		core.Debug("datacenter not enabled: %x, %x", state.Request.BuyerId, state.Request.DatacenterId)
		state.Error |= constants.SessionError_DatacenterNotEnabled
	}

	/*
		Get the datacenter information and store it in the handler state.
	*/

	state.Datacenter = state.Database.GetDatacenter(state.Request.DatacenterId)

	/*
		Get the set of relay ids that are in the destination datacenter (if applicable).

		If anything goes wrong, this is an empty set.
	*/

	destRelayIds := state.Database.GetDatacenterRelays(state.Request.DatacenterId)
	if len(destRelayIds) == 0 {
		core.Debug("no relays in datacenter %x", state.Request.DatacenterId)
		state.Error |= constants.SessionError_NoRelaysInDatacenter
	}

	state.DestRelayIds = destRelayIds

	/*
		The debug string is appended to during the rest of the handler and sent down to the SDK
		when Buyer.Debug is true. We use this to debug route decisions when something is not working.
	*/

	if state.Buyer.Debug {
		core.Debug("debug enabled")
		state.Debug = new(string)
	}

	return false
}

func SessionUpdate_NewSession(state *SessionUpdateState) {

	core.Debug("new session")

	state.Input.Version = packets.SDK_SessionDataVersion_Write
	state.Input.SessionId = state.Request.SessionId
	state.Input.SliceNumber = 0
	state.Input.StartTimestamp = uint64(time.Now().Unix())
	state.Input.ExpireTimestamp = state.Input.StartTimestamp
	state.Input.RouteState.ABTest = state.Buyer.RouteShader.ABTest

	state.Output = state.Input
	state.Output.SliceNumber += 1
	state.Output.ExpireTimestamp = state.Input.ExpireTimestamp + packets.SDK_SliceSeconds*2 + 1
}

func SessionUpdate_ExistingSession(state *SessionUpdateState) {

	core.Debug("existing session")

	/*
		Check for some obviously divergent data between the session request packet
		and the stored session data. If there is a mismatch, just return a direct route.
	*/

	if state.Input.SessionId != state.Request.SessionId {
		core.Debug("bad session id")
		state.Error |= constants.SessionError_BadSessionId
		return
	}

	if state.Input.SliceNumber != state.Request.SliceNumber {
		core.Debug("bad slice number")
		state.Error |= constants.SessionError_BadSliceNumber
		return
	}

	/*
		Copy input state to output and go to next slice.

		During the rest of the session update we transform session.output in place,
		before sending it back to the SDK in the session response packet.
	*/

	state.Output = state.Input
	state.Output.SliceNumber += 1
	state.Output.ExpireTimestamp += packets.SDK_SliceSeconds

	/*
		Track total next envelope bandwidth sent up and down
	*/

	if state.Request.Next {
		state.Output.NextEnvelopeBytesUpSum += uint64(state.Buyer.RouteShader.BandwidthEnvelopeUpKbps) * 1000 * packets.SDK_SliceSeconds / 8
		state.Output.NextEnvelopeBytesDownSum += uint64(state.Buyer.RouteShader.BandwidthEnvelopeDownKbps) * 1000 * packets.SDK_SliceSeconds / 8
	}

	/*
		Calculate real packet loss %

		This is driven from actual game packets, not ping packets.

		This value is typically much higher precision (60HZ), vs. ping packets (10HZ).
	*/

	slicePacketsSentClientToServer := state.Request.PacketsSentClientToServer - state.Input.PrevPacketsSentClientToServer
	slicePacketsSentServerToClient := state.Request.PacketsSentServerToClient - state.Input.PrevPacketsSentServerToClient

	slicePacketsLostClientToServer := state.Request.PacketsLostClientToServer - state.Input.PrevPacketsLostClientToServer
	slicePacketsLostServerToClient := state.Request.PacketsLostServerToClient - state.Input.PrevPacketsLostServerToClient

	var RealPacketLossClientToServer float32
	if slicePacketsSentClientToServer != uint64(0) {
		RealPacketLossClientToServer = float32(float64(slicePacketsLostClientToServer)/float64(slicePacketsSentClientToServer)) * 100.0
	}

	var RealPacketLossServerToClient float32
	if slicePacketsSentServerToClient != uint64(0) {
		RealPacketLossServerToClient = float32(float64(slicePacketsLostServerToClient)/float64(slicePacketsSentServerToClient)) * 100.0
	}

	state.RealPacketLoss = RealPacketLossClientToServer
	if RealPacketLossServerToClient > RealPacketLossClientToServer {
		state.RealPacketLoss = RealPacketLossServerToClient
	}

	/*
		Calculate real out of order packet %

		This is driven from actual game packets, not ping packets.
	*/

	slicePacketsOutOfOrderClientToServer := state.Request.PacketsOutOfOrderClientToServer - state.Input.PrevPacketsOutOfOrderClientToServer
	slicePacketsOutOfOrderServerToClient := state.Request.PacketsOutOfOrderServerToClient - state.Input.PrevPacketsOutOfOrderServerToClient

	var RealOutOfOrderClientToServer float32
	if slicePacketsSentClientToServer != uint64(0) {
		RealOutOfOrderClientToServer = float32(float64(slicePacketsOutOfOrderClientToServer)/float64(slicePacketsSentClientToServer)) * 100.0
	}

	var RealOutOfOrderServerToClient float32
	if slicePacketsSentServerToClient != uint64(0) {
		RealOutOfOrderServerToClient = float32(float64(slicePacketsOutOfOrderServerToClient)/float64(slicePacketsSentServerToClient)) * 100.0
	}

	state.RealOutOfOrder = RealOutOfOrderClientToServer
	if RealOutOfOrderServerToClient > RealOutOfOrderClientToServer {
		state.RealOutOfOrder = RealOutOfOrderServerToClient
	}

	/*
		Calculate real jitter.

		This is driven from actual game packets, not ping packets.

		Clamp jitter between client and server at 1000.

		It is meaningless beyond that...
	*/

	if state.Request.JitterClientToServer > 1000.0 {
		state.Request.JitterClientToServer = float32(1000)
	}

	if state.Request.JitterServerToClient > 1000.0 {
		state.Request.JitterServerToClient = float32(1000)
	}

	state.RealJitter = state.Request.JitterClientToServer
	if state.Request.JitterServerToClient > state.Request.JitterClientToServer {
		state.RealJitter = state.Request.JitterServerToClient
	}
}

func SessionUpdate_GetNearRelays(state *SessionUpdateState) bool {

	/*
		This function selects up to constants.MaxNearRelays near relays for the session,
		according to the players latitude and longitude determined by
		ip2location.

		These near relays are selected only on the first slice (slice 0)
		of a session, and are held fixed for the duration of the session.

		The SDK pings the near relays, and reports up the latency, jitter
		and packet loss to each near relay.

		Network Next uses this data in route planning, adding the latency
		to the first relay to the total route cost, and by excluding near relays
		with higher jitter or packet loss.
	*/

	state.GetNearRelays = true

	if state.Datacenter == nil {
		core.Debug("datacenter is nil, not getting near relays")
		state.NotGettingNearRelaysDatacenterIsNil = true
		return false
	}

	if state.Buyer.RouteShader.AnalysisOnly {
		core.Debug("analysis only, not getting near relays")
		state.NotGettingNearRelaysAnalysisOnly = true
		return false
	}

	if (state.Error & constants.SessionError_DatacenterNotEnabled) != 0 {
		core.Debug("datacenter not enabled, not getting near relays")
		state.NotGettingNearRelaysDatacenterNotEnabled = true
		return false
	}

	clientLatitude := state.Output.Latitude
	clientLongitude := state.Output.Longitude

	serverLatitude := state.Datacenter.Latitude
	serverLongitude := state.Datacenter.Longitude

	const distanceThreshold = 2500
	const latencyThreshold = 30.0

	nearRelayIds, nearRelayAddresses := common.GetNearRelays(constants.MaxNearRelays,
		distanceThreshold,
		latencyThreshold,
		state.RouteMatrix.RelayIds,
		state.RouteMatrix.RelayAddresses,
		state.RouteMatrix.RelayLatitudes,
		state.RouteMatrix.RelayLongitudes,
		clientLatitude,
		clientLongitude,
		serverLatitude,
		serverLongitude,
	)

	numNearRelays := len(nearRelayIds)

	if numNearRelays == 0 {
		core.Debug("no near relays :(")
		state.Error |= constants.SessionError_NoNearRelays
		return false
	}

	core.Debug("found %d near relays", numNearRelays)

	state.Response.HasNearRelays = true
	state.Response.NumNearRelays = int32(numNearRelays)

	for i := 0; i < numNearRelays; i++ {
		state.Response.NearRelayIds[i] = nearRelayIds[i]
		state.Response.NearRelayAddresses[i] = nearRelayAddresses[i]
	}

	// generate ping tokens

	if len(state.PingKey) == 0 {
		panic("missing ping key")
	}

	expireTimestamp := uint64(time.Now().Unix()) + 15

	core.GeneratePingTokens(expireTimestamp, &state.Request.ClientAddress, state.Response.NearRelayAddresses[:state.Response.NumNearRelays], state.PingKey, state.Response.NearRelayPingTokens[:])

	state.Response.NearRelayExpireTimestamp = expireTimestamp

	return true
}

func SessionUpdate_UpdateNearRelays(state *SessionUpdateState) bool {

	if state.Buyer.RouteShader.AnalysisOnly {
		core.Debug("analysis only, not updating near relay stats")
		state.NotUpdatingNearRelaysAnalysisOnly = true
		return false
	}

	if (state.Error & constants.SessionError_DatacenterNotEnabled) != 0 {
		core.Debug("datacenter not enabled, not updating near relay stats")
		state.NotUpdatingNearRelaysDatacenterNotEnabled = true
		return false
	}

	/*
		Debug print near relay ping results on slice 1. This is when the SDK tells us the near relay ping results.
	*/

	if state.Request.SliceNumber == 1 {
		core.Debug("sdk uploaded near relay stats for %d relays:", state.Request.NumNearRelays)
		for i := 0; i < int(state.Request.NumNearRelays); i++ {
			relayId := state.Request.NearRelayIds[i]
			relayIndex, exists := state.RouteMatrix.RelayIdToIndex[relayId]
			var relayName string
			if exists {
				relayName = state.RouteMatrix.RelayNames[relayIndex]
			} else {
				relayName = "???" // near relay no longer exists in route matrix
			}
			rtt := state.Request.NearRelayRTT[i]
			jitter := state.Request.NearRelayJitter[i]
			pl := state.Request.NearRelayPacketLoss[i]
			core.Debug(" + %s [%016x] rtt = %d, jitter = %d, pl = %.2f", relayName, relayId, rtt, jitter, pl)
		}
	}

	/*
		Reframe dest relays to get them relative to the current route matrix.
	*/

	outputNumDestRelays := 0
	outputDestRelays := make([]int32, len(state.DestRelayIds))

	core.ReframeDestRelays(state.RouteMatrix.RelayIdToIndex, state.DestRelayIds, &outputNumDestRelays, outputDestRelays)

	state.DestRelays = outputDestRelays[:outputNumDestRelays]

	core.Debug("reframed dest relays %d -> %d", len(state.DestRelayIds), outputNumDestRelays)

	/*
		Filter source relays and get them in a form relative to the current route matrix
	*/

	directLatency := int32(math.Ceil(float64(state.Request.DirectRTT)))
	directJitter := int32(math.Ceil(float64(state.Request.DirectJitter)))
	directPacketLoss := state.Request.DirectMaxPacketLossSeen

	sourceRelayIds := state.Request.NearRelayIds[:state.Request.NumNearRelays]
	sourceRelayLatency := state.Request.NearRelayRTT[:state.Request.NumNearRelays]
	sourceRelayJitter := state.Request.NearRelayJitter[:state.Request.NumNearRelays]
	sourceRelayPacketLoss := state.Request.NearRelayPacketLoss[:state.Request.NumNearRelays]

	filteredSourceRelayLatency := [constants.MaxNearRelays]int32{}

	core.FilterSourceRelays(state.RouteMatrix.RelayIdToIndex,
		directLatency,
		directJitter,
		directPacketLoss,
		sourceRelayIds,
		sourceRelayLatency,
		sourceRelayJitter,
		sourceRelayPacketLoss,
		filteredSourceRelayLatency[:])

	outputSourceRelays := make([]int32, len(sourceRelayIds))
	outputSourceRelayLatency := make([]int32, len(sourceRelayIds))

	core.Debug("filtered near relays %d -> %d", state.Request.NumNearRelays, len(sourceRelayIds))

	core.ReframeSourceRelays(state.RouteMatrix.RelayIdToIndex, sourceRelayIds, filteredSourceRelayLatency[:], outputSourceRelays, outputSourceRelayLatency)

	state.SourceRelays = outputSourceRelays
	state.SourceRelayRTT = outputSourceRelayLatency

	core.Debug("reframed near relays %d -> %d", len(sourceRelayIds), len(state.SourceRelays))

	return true
}

func SessionUpdate_BuildNextTokens(state *SessionUpdateState, routeNumRelays int32, routeRelays []int32) {

	state.Output.SessionVersion++

	numTokens := routeNumRelays + 2

	var routePublicAddresses [constants.NEXT_MAX_NODES]net.UDPAddr
	var routeHasInternalAddresses [constants.NEXT_MAX_NODES]bool
	var routeInternalAddresses [constants.NEXT_MAX_NODES]net.UDPAddr
	var routeInternalGroups [constants.NEXT_MAX_NODES]uint64
	var routeSellers [constants.NEXT_MAX_NODES]int
	var routePublicKeys [constants.NEXT_MAX_NODES][]byte

	// client node

	routePublicKeys[0] = state.Request.ClientRoutePublicKey[:]
	routePublicAddresses[0] = state.Request.ClientAddress

	// relay nodes

	relayPublicAddresses := routePublicAddresses[1 : numTokens-1]
	relayHasInternalAddresses := routeHasInternalAddresses[1 : numTokens-1]
	relayInternalAddresses := routeInternalAddresses[1 : numTokens-1]
	relayInternalGroups := routeInternalGroups[1 : numTokens-1]
	relaySellers := routeSellers[1 : numTokens-1]
	relayPublicKeys := routePublicKeys[1 : numTokens-1]

	numRouteRelays := len(routeRelays)

	for i := 0; i < numRouteRelays; i++ {

		relayIndex := routeRelays[i]

		relay := &state.Database.Relays[relayIndex]

		relayPublicAddresses[i] = relay.PublicAddress
		relayHasInternalAddresses[i] = relay.HasInternalAddress
		relayInternalAddresses[i] = relay.InternalAddress
		relayInternalGroups[i] = relay.InternalGroup
		relaySellers[i] = int(relay.Seller.Id)
		relayPublicKeys[i] = relay.PublicKey
	}

	// server node

	routePublicAddresses[numTokens-1] = *state.From
	routePublicKeys[numTokens-1] = state.Request.ServerRoutePublicKey[:]

	// write the tokens

	tokenData := make([]byte, numTokens*packets.SDK_EncryptedNextRouteTokenSize)

	sessionId := state.Output.SessionId
	sessionVersion := uint8(state.Output.SessionVersion)
	expireTimestamp := state.Output.ExpireTimestamp
	envelopeUpKbps := uint32(state.Buyer.RouteShader.BandwidthEnvelopeUpKbps)
	envelopeDownKbps := uint32(state.Buyer.RouteShader.BandwidthEnvelopeDownKbps)

	core.WriteRouteTokens(tokenData, expireTimestamp, sessionId, sessionVersion, envelopeUpKbps, envelopeDownKbps, int(numTokens), routePublicAddresses[:], routeHasInternalAddresses[:], routeInternalAddresses[:], routeInternalGroups[:], routeSellers[:], routePublicKeys[:], state.RelayBackendPrivateKey[:])

	state.Response.RouteType = packets.SDK_RouteTypeNew
	state.Response.NumTokens = numTokens
	state.Response.Tokens = tokenData
}

func SessionUpdate_BuildContinueTokens(state *SessionUpdateState, routeNumRelays int32, routeRelays []int32) {

	numTokens := routeNumRelays + 2

	var routePublicKeys [constants.NEXT_MAX_NODES][]byte

	// client node

	routePublicKeys[0] = state.Request.ClientRoutePublicKey[:]

	// relay nodes

	relayPublicKeys := routePublicKeys[1 : numTokens-1]

	numRouteRelays := len(routeRelays)

	for i := 0; i < numRouteRelays; i++ {
		relayIndex := routeRelays[i]
		relay := &state.Database.Relays[relayIndex]
		relayPublicKeys[i] = relay.PublicKey
	}

	// server node

	routePublicKeys[numTokens-1] = state.Request.ServerRoutePublicKey[:]

	// build the tokens

	tokenData := make([]byte, numTokens*packets.SDK_EncryptedContinueRouteTokenSize)

	sessionId := state.Output.SessionId
	sessionVersion := uint8(state.Output.SessionVersion)
	expireTimestamp := state.Output.ExpireTimestamp

	core.WriteContinueTokens(tokenData, expireTimestamp, sessionId, sessionVersion, int(numTokens), routePublicKeys[:], state.RelayBackendPrivateKey[:])

	state.Response.RouteType = packets.SDK_RouteTypeContinue
	state.Response.NumTokens = numTokens
	state.Response.Tokens = tokenData
}

func SessionUpdate_MakeRouteDecision(state *SessionUpdateState) {

	/*
		If we are on on network next but don't have any relays in our route, something is WRONG.
		Veto the session and go direct.
	*/

	if state.Input.RouteState.Next && state.Input.RouteNumRelays == 0 {
		core.Debug("on network next, but no route relays?")
		state.Output.RouteState.Next = false
		state.Output.RouteState.Veto = true
		state.Error |= constants.SessionError_NoRouteRelays
		if state.Debug != nil {
			*state.Debug += "no route relays?!\n"
		}
		return
	}

	var stayOnNext bool
	var routeChanged bool
	var routeCost int32
	var routeNumRelays int32

	routeRelays := [constants.MaxRouteRelays]int32{}

	sliceNumber := int32(state.Request.SliceNumber)

	if !state.Input.RouteState.Next {

		// currently going direct. should we take network next?

		if core.MakeRouteDecision_TakeNetworkNext(state.Request.UserHash,
			state.RouteMatrix.RouteEntries,
			&state.Buyer.RouteShader,
			&state.Output.RouteState,
			int32(state.Request.DirectRTT),
			state.RealPacketLoss,
			state.SourceRelays,
			state.SourceRelayRTT,
			state.DestRelays,
			&routeCost,
			&routeNumRelays,
			routeRelays[:],
			&state.RouteDiversity,
			state.Debug,
			sliceNumber) {

			state.TakeNetworkNext = true

			SessionUpdate_BuildNextTokens(state, routeNumRelays, routeRelays[:routeNumRelays])

			if state.Debug != nil {

				*state.Debug += "take network next: "

				for i, routeRelay := range routeRelays[:routeNumRelays] {
					if i != int(routeNumRelays-1) {
						*state.Debug += fmt.Sprintf("%s - ", state.RouteMatrix.RelayNames[routeRelay])
					} else {
						*state.Debug += fmt.Sprintf("%s\n", state.RouteMatrix.RelayNames[routeRelay])
					}
				}
			}

		} else {

			state.StayDirect = true

			if state.Debug != nil {
				*state.Debug += "staying direct\n"
			}

		}

	} else {

		// currently taking network next

		if !state.Request.Next {

			// the sdk aborted this session

			core.Debug("aborted")
			state.Output.RouteState.Next = false
			state.Output.RouteState.Veto = true
			state.Error |= constants.SessionError_Aborted
			if state.Debug != nil {
				*state.Debug += "aborted\n"
			}
			return
		}

		// reframe the current route in terms of relay indices in the current route matrix

		if !core.ReframeRoute(state.RouteMatrix.RelayIdToIndex, state.Output.RouteRelayIds[:state.Output.RouteNumRelays], &routeRelays) {
			routeRelays = [constants.MaxRouteRelays]int32{}
			core.Debug("one or more relays in the route no longer exist")
			state.Error |= constants.SessionError_RouteRelayNoLongerExists
			if state.Debug != nil {
				*state.Debug += "route relay no longer exists\n"
			}
		}

		// make route decision

		directLatency := int32(state.Request.DirectRTT)
		nextLatency := int32(state.Request.NextRTT)
		predictedLatency := state.Input.RouteCost

		stayOnNext, routeChanged = core.MakeRouteDecision_StayOnNetworkNext(state.Request.UserHash,
			state.RouteMatrix.RouteEntries,
			state.RouteMatrix.RelayNames,
			&state.Buyer.RouteShader,
			&state.Output.RouteState,
			directLatency,
			nextLatency,
			predictedLatency,
			state.RealPacketLoss,
			state.Request.NextPacketLoss,
			state.Output.RouteNumRelays,
			routeRelays,
			state.SourceRelays,
			state.SourceRelayRTT,
			state.DestRelays,
			&routeCost,
			&routeNumRelays,
			routeRelays[:],
			state.Debug)

		if stayOnNext {

			// stay on network next

			if routeChanged {

				core.Debug("route changed")

				state.RouteChanged = true

				SessionUpdate_BuildNextTokens(state, routeNumRelays, routeRelays[:routeNumRelays])

				if state.Debug != nil {

					*state.Debug += "route changed: "

					for i, routeRelay := range routeRelays[:routeNumRelays] {
						if i != int(routeNumRelays-1) {
							*state.Debug += fmt.Sprintf("%s - ", state.RouteMatrix.RelayNames[routeRelay])
						} else {
							*state.Debug += fmt.Sprintf("%s\n", state.RouteMatrix.RelayNames[routeRelay])
						}
					}
				}

			} else {

				core.Debug("route continued")

				state.RouteContinued = true

				SessionUpdate_BuildContinueTokens(state, routeNumRelays, routeRelays[:routeNumRelays])
				if state.Debug != nil {
					*state.Debug += "route continued\n"
				}

			}

		} else {

			// leave network next

			if state.Output.RouteState.NoRoute {
				core.Debug("route no longer exists")
				state.Error |= constants.SessionError_RouteNoLongerExists
				if state.Debug != nil {
					*state.Debug += "route no longer exists\n"
				}
			}

			if state.Output.RouteState.Mispredict {
				core.Debug("mispredict")
				if state.Debug != nil {
					*state.Debug += "mispredict\n"
				}
			}

			if state.Output.RouteState.LatencyWorse {
				core.Debug("latency worse")
				if state.Debug != nil {
					*state.Debug += "latency worse\n"
				}
			}
		}
	}

	/*
		Multipath means to send packets across both the direct and the network
		next route at the same time, which reduces packet loss.
	*/

	state.Response.Multipath = state.Output.RouteState.Multipath

	/*
		Stick the route cost, whether the route changed, and the route relay data
		in the output state. This output state is serialized into the route state
		in the route response, and sent back up to us, allowing us to know the
		current network next route, when we plan the next 10 second slice.
	*/

	if routeCost > constants.MaxRouteCost {
		routeCost = constants.MaxRouteCost
	}

	if state.Debug != nil {
		if routeCost != 0 {
			*state.Debug += fmt.Sprintf("route cost is %d\n", routeCost)
		}
	}

	state.Output.RouteCost = routeCost
	state.Output.RouteChanged = routeChanged
	state.Output.RouteNumRelays = routeNumRelays

	for i := int32(0); i < routeNumRelays; i++ {
		relayId := state.RouteMatrix.RelayIds[routeRelays[i]]
		state.Output.RouteRelayIds[i] = relayId
	}
}

func SessionUpdate_Post(state *SessionUpdateState) {

	/*
		If we have already wrote the session summary, there is nothing more to do...
	*/

	if state.Input.WroteSummary {
		return
	}

	/*
		Logic for sending near relay messages to portal
		This is somewhat complicated because at scale we usually only send next sessions to the portal
		Thus, we send it right away if that flag is off, OR, if we are at scale, we must remember to send it only once
		the first slice we are on next.
	*/

	shouldSendNearRelays := state.Input.SliceNumber >= 1 && (!state.PortalNextSessionsOnly || state.Output.RouteState.Next)

	if !state.Input.SentNearRelaysToPortal && shouldSendNearRelays {
		state.Output.SentNearRelaysToPortal = true
	}

	/*
		Accumulate error flags from input state, and from this session update, then write it to output.

		This lets us write error flags in the session summary only, and we capture all errors that occurred for a session.
	*/

	state.Output.Error = state.Input.Error | state.Error

	/*
		Build the set of near relays for the SDK to ping.

		The SDK pings these near relays and reports up the results in the next session update.

		We hold the set of near relays fixed for the session, so we only do this work on the first slice.
	*/

	if state.Request.SliceNumber == 0 {
		state.Output.Latitude = state.Latitude
		state.Output.Longitude = state.Longitude
		SessionUpdate_GetNearRelays(state)
		core.Debug("first slice always goes direct")
	}

	/*
		Since post runs at the end of every session handler, run logic
		here that must run if we are taking network next vs. direct
	*/

	if state.Response.RouteType != packets.SDK_RouteTypeDirect {
		core.Debug("session takes network next")
	} else {
		core.Debug("session goes direct")
	}

	/*
		Track duration of time spent on network next, and if the session has ever been on network next.
	*/

	if state.Input.RouteState.Next {
		state.Output.DurationOnNext += packets.SDK_SliceSeconds
		core.Debug("session has been on network next for %d seconds", state.Output.DurationOnNext)
	}

	/*
		Store the *previous* packets sent and packets lost counters in the route state,

		This lets us perform a delta each slice to calculate real packet loss in high precision, per-slice.
	*/

	state.Output.PrevPacketsSentClientToServer = state.Request.PacketsSentClientToServer
	state.Output.PrevPacketsSentServerToClient = state.Request.PacketsSentServerToClient
	state.Output.PrevPacketsLostClientToServer = state.Request.PacketsLostClientToServer
	state.Output.PrevPacketsLostServerToClient = state.Request.PacketsLostServerToClient
	state.Output.PrevPacketsOutOfOrderClientToServer = state.Request.PacketsOutOfOrderClientToServer
	state.Output.PrevPacketsOutOfOrderServerToClient = state.Request.PacketsOutOfOrderServerToClient

	/*
		If the core routing logic generated a debug string, include it in the response packet
	*/

	if state.Debug != nil {
		state.Response.Debug = *state.Debug
		if state.Response.Debug != "" {
			state.Response.HasDebug = true
		}
	}

	/*
		The session ends when the client ping times out or the client falls back to direct.

		At this point we write a summary slice to bigquery, with more information than regular slices.

		This saves a lot of bandwidth and bigquery cost, by only writing this information once per-session.
	*/

	if state.Request.ClientPingTimedOut || state.Request.FallbackToDirect { // IMPORTANT: once set, these values will each remain true thereafter

		// write summary only once

		if !state.Output.WroteSummary && !state.Output.WriteSummary {
			state.Output.WriteSummary = true
		} else if state.Output.WriteSummary && !state.Output.WroteSummary {
			state.Output.WroteSummary = true
			state.Output.WriteSummary = false
		}
	}

	/*
		Take note of when we send near relays down to the client. Useful for debugging.
	*/

	if state.Response.HasNearRelays {
		core.Debug("sending %d near relays down to client:", state.Response.NumNearRelays)
		for i := 0; i < int(state.Response.NumNearRelays); i++ {
			relayIndex := state.RouteMatrix.RelayIdToIndex[state.Response.NearRelayIds[i]]
			relayName := state.RouteMatrix.RelayNames[relayIndex]
			core.Debug(" + %s [%016x]", relayName, state.Response.NearRelayIds[i])
		}
	}

	if state.Output.SliceNumber == 1 && state.Response.HasNearRelays == false {
		core.Debug("no near relays sent down to client for slice 1?!!!")
	}

	if state.Output.SliceNumber == 1 && state.Response.NumNearRelays == 0 {
		core.Debug("num near relays is zero for slice 1?!!!!")
	}

	/*
		Write session data
	*/

	writeStream := encoding.CreateWriteStream(state.Response.SessionData[:])

	state.Output.Version = packets.SDK_SessionDataVersion_Write

	err := state.Output.Serialize(writeStream)
	if err != nil {
		core.Error("failed to write session data: %v", err)
		state.Error |= constants.SessionError_FailedToWriteSessionData
		return
	}

	writeStream.Flush()

	state.Response.SessionDataBytes = int32(writeStream.GetBytesProcessed())

	copy(state.Response.SessionDataSignature[:], crypto.Sign(state.Response.SessionData[:state.Response.SessionDataBytes], state.ServerBackendPrivateKey))

	/*
		Write the session update response packet.
	*/

	if state.Debug != nil {
		state.Response.Debug = *state.Debug
		core.Debug("-------------------------------------\n%s-------------------------------------", *state.Debug)
	}

	state.ResponsePacket, err = packets.SDK_WritePacket(&state.Response, packets.SDK_SESSION_UPDATE_RESPONSE_PACKET, packets.SDK_MaxPacketBytes, state.ServerBackendAddress, state.From, state.ServerBackendPrivateKey[:])
	if err != nil {
		core.Error("failed to write session update response packet: %v", err)
		state.Error |= constants.SessionError_FailedToWriteResponsePacket
		return
	}

	state.WroteResponsePacket = true

	/*
		Send various messages to drive the portal and analytics
	*/

	if !state.FirstUpdate {

		sendPortalSessionUpdateMessage(state)
		sendPortalNearRelayUpdateMessage(state)

		sendAnalyticsSessionUpdateMessage(state)
		sendAnalyticsSessionSummaryMessage(state)
		sendAnalyticsNearRelayPingMessages(state)
	}
}

// -----------------------------------------

func sendPortalSessionUpdateMessage(state *SessionUpdateState) {

	if state.Input.SliceNumber < 1 {
		return
	}

	if state.Request.ClientPingTimedOut {
		return
	}

	message := messages.PortalSessionUpdateMessage{}

	message.Timestamp = state.StartTimestamp

	message.ClientAddress = state.Request.ClientAddress
	message.ServerAddress = state.Request.ServerAddress

	message.SDKVersion_Major = byte(state.Request.Version.Major)
	message.SDKVersion_Minor = byte(state.Request.Version.Minor)
	message.SDKVersion_Patch = byte(state.Request.Version.Patch)

	message.SessionId = state.Request.SessionId
	message.UserHash = state.Request.UserHash
	message.StartTime = state.Input.StartTimestamp
	message.BuyerId = state.Request.BuyerId
	message.DatacenterId = state.Request.DatacenterId
	message.Latitude = state.Input.Latitude
	message.Longitude = state.Input.Longitude
	message.SliceNumber = state.Input.SliceNumber - 1 // IMPORTANT: Line it up with data coming from the SDK
	message.SessionEvents = state.Request.SessionEvents
	message.InternalEvents = state.Request.InternalEvents
	message.ConnectionType = uint8(state.Request.ConnectionType)
	message.PlatformType = uint8(state.Request.PlatformType)

	message.DirectRTT = state.Request.DirectRTT
	message.DirectJitter = state.Request.DirectJitter
	message.DirectPacketLoss = state.Request.DirectPacketLoss
	message.DirectKbpsUp = state.Request.DirectKbpsUp
	message.DirectKbpsDown = state.Request.DirectKbpsDown

	message.Next = state.Request.Next

	if message.Next {
		message.NextRTT = state.Request.NextRTT
		message.NextJitter = state.Request.NextJitter
		message.NextPacketLoss = state.Request.NextPacketLoss
		message.NextKbpsUp = state.Request.NextKbpsUp
		message.NextKbpsDown = state.Request.NextKbpsDown
		message.NextPredictedRTT = uint32(state.Input.RouteCost)
		message.NextNumRouteRelays = uint32(state.Input.RouteNumRelays)
		for i := 0; i < int(message.NextNumRouteRelays); i++ {
			message.NextRouteRelayId[i] = state.Input.RouteRelayIds[i]
		}
	}

	message.RealJitter = state.RealJitter
	message.RealPacketLoss = state.RealPacketLoss
	message.RealOutOfOrder = state.RealOutOfOrder

	message.NumNearRelays = uint32(state.Request.NumNearRelays)
	for i := 0; i < int(message.NumNearRelays); i++ {
		message.NearRelayId[i] = state.Request.NearRelayIds[i]
		message.NearRelayRTT[i] = byte(state.Request.NearRelayRTT[i])
		message.NearRelayJitter[i] = byte(state.Request.NearRelayJitter[i])
		message.NearRelayPacketLoss[i] = state.Request.NearRelayPacketLoss[i]
		message.NearRelayRoutable[i] = state.Request.NearRelayRTT[i] != 255
	}

	message.BestScore = state.Output.BestScore
	message.BestDirectRTT = state.Output.BestDirectRTT
	message.BestNextRTT = state.Output.BestNextRTT

	message.Retry = state.Request.RetryNumber != 0
	message.FallbackToDirect = state.FallbackToDirect
	message.SendToPortal = !state.PortalNextSessionsOnly || (state.PortalNextSessionsOnly && state.Output.DurationOnNext > 0)

	if state.PortalSessionUpdateMessageChannel != nil {
		state.PortalSessionUpdateMessageChannel <- &message
		state.SentPortalSessionUpdateMessage = true
	}
}

func sendPortalNearRelayUpdateMessage(state *SessionUpdateState) {

	if !(state.Input.SentNearRelaysToPortal == false && state.Output.SentNearRelaysToPortal == true) {
		return
	}

	message := messages.PortalNearRelayUpdateMessage{}

	message.Timestamp = state.StartTimestamp
	message.BuyerId = state.Request.BuyerId
	message.SessionId = state.Output.SessionId
	message.NumNearRelays = uint32(state.Request.NumNearRelays)
	for i := 0; i < int(state.Request.NumNearRelays); i++ {
		message.NearRelayId[i] = state.Request.NearRelayIds[i]
		message.NearRelayRTT[i] = byte(state.Request.NearRelayRTT[i])
		message.NearRelayJitter[i] = byte(state.Request.NearRelayJitter[i])
		message.NearRelayPacketLoss[i] = state.Request.NearRelayPacketLoss[i]
	}

	if state.PortalNearRelayUpdateMessageChannel != nil {
		state.PortalNearRelayUpdateMessageChannel <- &message
		state.SentPortalNearRelayUpdateMessage = true
	}
}

// ---------------------------------------------------------------------------------

func sendAnalyticsNearRelayPingMessages(state *SessionUpdateState) {

	if state.Request.SliceNumber != 1 {
		return
	}

	for i := 0; i < int(state.Request.NumNearRelays); i++ {

		message := messages.AnalyticsNearRelayPingMessage{}

		message.Timestamp = int64(state.StartTimestampNano / 1000) // nano -> micro
		message.BuyerId = int64(state.Request.BuyerId)
		message.SessionId = int64(state.Output.SessionId)
		message.UserHash = int64(state.Request.UserHash)
		message.Latitude = state.Output.Latitude
		message.Longitude = state.Output.Longitude
		message.ClientAddress = state.Request.ClientAddress.String()
		message.ConnectionType = int32(state.Request.ConnectionType)
		message.PlatformType = int32(state.Request.PlatformType)
		message.NearRelayId = int64(state.Request.NearRelayIds[i])
		message.NearRelayRTT = int32(state.Request.NearRelayRTT[i])
		message.NearRelayJitter = int32(state.Request.NearRelayJitter[i])
		message.NearRelayPacketLoss = state.Request.NearRelayPacketLoss[i]

		if state.AnalyticsNearRelayPingMessageChannel != nil {
			state.AnalyticsNearRelayPingMessageChannel <- &message
			state.SentAnalyticsNearRelayPingMessage = true
		}

	}
}

func sendAnalyticsSessionUpdateMessage(state *SessionUpdateState) {

	if state.Request.SliceNumber < 1 {
		return
	}

	message := messages.AnalyticsSessionUpdateMessage{}

	// always

	message.Timestamp = int64(state.StartTimestampNano / 1000) // nano -> micro
	message.SessionId = int64(state.Request.SessionId)
	message.SliceNumber = int32(state.Request.SliceNumber - 1) // IMPORTANT: Line it up with data coming from the SDK
	message.RealPacketLoss = state.RealPacketLoss
	message.RealJitter = state.RealJitter
	message.RealOutOfOrder = state.RealOutOfOrder
	message.SessionEvents = int64(state.Request.SessionEvents)
	message.InternalEvents = int64(state.Request.InternalEvents)
	message.DirectRTT = state.Request.DirectRTT
	message.DirectJitter = state.Request.DirectJitter
	message.DirectPacketLoss = state.Request.DirectPacketLoss
	message.DirectKbpsUp = int32(state.Request.DirectKbpsUp)
	message.DirectKbpsDown = int32(state.Request.DirectKbpsDown)

	// next only

	message.Next = state.Request.Next
	if message.Next {
		message.NextRTT = state.Request.NextRTT
		message.NextJitter = state.Request.NextJitter
		message.NextPacketLoss = state.Request.NextPacketLoss
		message.NextKbpsUp = int32(state.Request.NextKbpsUp)
		message.NextKbpsDown = int32(state.Request.NextKbpsDown)
		message.NextPredictedRTT = float32(state.Input.RouteCost)
		message.NextRouteRelays = make([]int64, len(state.Input.RouteRelayIds))
		for i := range state.Input.RouteRelayIds {
			message.NextRouteRelays[i] = int64(state.Input.RouteRelayIds[i])
		}
	}

	// flags

	message.FallbackToDirect = state.Request.FallbackToDirect
	message.Reported = state.Request.Reported
	message.LatencyReduction = state.Input.RouteState.ReduceLatency
	message.PacketLossReduction = state.Input.RouteState.ReducePacketLoss
	message.ForceNext = state.Input.RouteState.ForcedNext
	message.LongSessionUpdate = state.LongSessionUpdate
	message.ClientNextBandwidthOverLimit = state.Request.ClientNextBandwidthOverLimit
	message.ServerNextBandwidthOverLimit = state.Request.ServerNextBandwidthOverLimit
	message.Veto = state.Input.RouteState.Veto
	message.Disabled = state.Input.RouteState.Disabled
	message.NotSelected = state.Input.RouteState.NotSelected
	message.A = state.Input.RouteState.A
	message.B = state.Input.RouteState.B
	message.LatencyWorse = state.Input.RouteState.LatencyWorse
	message.LocationVeto = state.Input.RouteState.LocationVeto
	message.Mispredict = state.Input.RouteState.Mispredict
	message.LackOfDiversity = state.Input.RouteState.LackOfDiversity

	// send message

	if state.AnalyticsSessionUpdateMessageChannel != nil {
		state.AnalyticsSessionUpdateMessageChannel <- &message
		state.SentAnalyticsSessionUpdateMessage = true
	}
}

func sendAnalyticsSessionSummaryMessage(state *SessionUpdateState) {

	if !state.Output.WriteSummary {
		return
	}

	message := messages.AnalyticsSessionSummaryMessage{}

	message.Timestamp = int64(state.StartTimestampNano / 1000) // nano -> micro
	message.SessionId = int64(state.Request.SessionId)
	message.DatacenterId = int64(state.Request.DatacenterId)
	message.BuyerId = int64(state.Request.BuyerId)
	message.UserHash = int64(state.Request.UserHash)
	message.Latitude = state.Input.Latitude
	message.Longitude = state.Input.Longitude
	message.ClientAddress = state.Request.ClientAddress.String()
	message.ServerAddress = state.Request.ServerAddress.String()
	message.ConnectionType = int32(state.Request.ConnectionType)
	message.PlatformType = int32(state.Request.PlatformType)
	message.SDKVersion_Major = int32(state.Request.Version.Major)
	message.SDKVersion_Minor = int32(state.Request.Version.Minor)
	message.SDKVersion_Patch = int32(state.Request.Version.Patch)
	message.ClientToServerPacketsSent = int64(state.Request.PacketsSentClientToServer)
	message.ServerToClientPacketsSent = int64(state.Request.PacketsSentServerToClient)
	message.ClientToServerPacketsLost = int64(state.Request.PacketsLostClientToServer)
	message.ServerToClientPacketsLost = int64(state.Request.PacketsLostServerToClient)
	message.ClientToServerPacketsOutOfOrder = int64(state.Request.PacketsOutOfOrderClientToServer)
	message.ServerToClientPacketsOutOfOrder = int64(state.Request.PacketsOutOfOrderServerToClient)
	message.SessionDuration = int32((state.Request.SliceNumber - 1) * packets.SDK_SliceSeconds)
	message.TotalNextEnvelopeBytesUp = int64(state.Input.NextEnvelopeBytesUpSum)
	message.TotalNextEnvelopeBytesDown = int64(state.Input.NextEnvelopeBytesDownSum)
	message.DurationOnNext = int32(state.Input.DurationOnNext) // seconds
	message.StartTimestamp = int64(state.Input.StartTimestamp * 1000000) // seconds -> microseconds
	message.Error = int64(state.Input.Error)

	// flags

	message.Reported = state.Request.Reported
	message.LatencyReduction = state.Input.RouteState.ReduceLatency
	message.PacketLossReduction = state.Input.RouteState.ReducePacketLoss
	message.ForceNext = state.Input.RouteState.ForcedNext
	message.LongSessionUpdate = state.LongSessionUpdate
	message.ClientNextBandwidthOverLimit = state.Request.ClientNextBandwidthOverLimit
	message.ServerNextBandwidthOverLimit = state.Request.ServerNextBandwidthOverLimit
	message.Veto = state.Input.RouteState.Veto
	message.Disabled = state.Input.RouteState.Disabled
	message.NotSelected = state.Input.RouteState.NotSelected
	message.A = state.Input.RouteState.A
	message.B = state.Input.RouteState.B
	message.LatencyWorse = state.Input.RouteState.LatencyWorse
	message.LocationVeto = state.Input.RouteState.LocationVeto
	message.Mispredict = state.Input.RouteState.Mispredict
	message.LackOfDiversity = state.Input.RouteState.LackOfDiversity

	// send it

	if state.AnalyticsSessionSummaryMessageChannel != nil {
		state.AnalyticsSessionSummaryMessageChannel <- &message
		state.SentAnalyticsSessionSummaryMessage = true
	}
}

// -----------------------------------------
