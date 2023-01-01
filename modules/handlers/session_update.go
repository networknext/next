package handlers

import (
	"fmt"
	"math"
	"net"
	"time"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/modules/packets"
	db "github.com/networknext/backend/modules/database"
)

type SessionUpdateState struct {

	/*
	   Convenience state struct for the session update handler.

	   We put all the state in here so it's easy to call out to functions to do work.

	   Otherwise we have to pass a million parameters into every function and it gets old fast.
	*/

	RoutingPrivateKey [crypto.Box_KeySize]byte

	ServerBackendAddress *net.UDPAddr

	LocateIP func(ip net.IP) (packets.SDK5_LocationData, error)

	Connection *net.UDPConn
	From       *net.UDPAddr

	Input packets.SDK5_SessionData // sent up from the SDK. previous slice.

	Output packets.SDK5_SessionData // sent down to the SDK. current slice.

	Request       *packets.SDK5_SessionUpdateRequestPacket
	Response      packets.SDK5_SessionUpdateResponsePacket
	Database      *db.Database
	RouteMatrix   *common.RouteMatrix
	Datacenter    db.Datacenter
	Buyer         db.Buyer
	Debug         *string
	StaleDuration time.Duration

	// real packet loss (from actual game packets). high precision %
	RealPacketLoss float32

	// real jitter (from actual game packets).
	RealJitter float32

	// route diversity is the number unique near relays with viable routes
	RouteDiversity int32

	// for route planning (comes from SDK and route matrix)
	NumNearRelays    int
	NearRelayIndices [core.MaxNearRelays]int32
	NearRelayRTTs    [core.MaxNearRelays]int32
	NearRelayJitters [core.MaxNearRelays]int32
	NumDestRelays    int32
	DestRelays       []int32

	// for session post (billing, portal etc...)
	PostNearRelayCount               int
	PostNearRelayIDs                 [core.MaxNearRelays]uint64
	PostNearRelayNames               [core.MaxNearRelays]string
	PostNearRelayAddresses           [core.MaxNearRelays]net.UDPAddr
	PostNearRelayRTT                 [core.MaxNearRelays]float32
	PostNearRelayJitter              [core.MaxNearRelays]float32
	PostNearRelayPacketLoss          [core.MaxNearRelays]float32
	PostRouteRelayNames              [core.MaxRelaysPerRoute]string
	PostRouteRelaySellers            [core.MaxRelaysPerRoute]db.Seller
	PostRealPacketLossClientToServer float32
	PostRealPacketLossServerToClient float32

	// flags
	ReadSessionData                                    bool
	LongDuration                                       bool
	ClientPingTimedOut                                 bool
	Pro                                                bool
	BadSessionId                                       bool
	BadSliceNumber                                     bool
	AnalysisOnly                                       bool
	NoRelaysInDatacenter                               bool
	HoldingNearRelays                                  bool
	NearRelaysExcluded                                 bool
	UsingHeldNearRelays                                bool
	NotGettingNearRelaysAnalysisOnly                   bool
	NotGettingNearRelaysDatacenterAccelerationDisabled bool
	FallbackToDirect                                   bool
	NoNearRelays                                       bool
	LargeCustomer                                      bool
	NoRouteRelays                                      bool
	Aborted                                            bool
	RouteRelayNoLongerExists                           bool
	RouteChanged                                       bool
	RouteContinued                                     bool
	RouteNoLongerExists                                bool
	Mispredict                                         bool
	LatencyWorse                                       bool
	FailedToReadSessionData                            bool
	StaleRouteMatrix                                   bool
	UnknownDatacenter                                  bool
	DatacenterNotEnabled                               bool
	TakeNetworkNext                                    bool
	LeftNetworkNext                                    bool
	WroteResponsePacket                                bool
	FailedToWriteResponsePacket                        bool
	FailedToSendResponsePacket                         bool
	LocationVeto                                       bool
	SentSessionUpdateMessage                           bool
	SentPortalData                                     bool
	LocatedIP                                          bool
}

func SessionUpdate_Pre(state *SessionUpdateState) bool {

	/*
		If the route shader is in analysis only mode, set the analysis only flag in the state

		We don't acceleration sessions in analysis only mode.
	*/

	if state.Buyer.RouteShader.AnalysisOnly {
		core.Debug("analysis only") // tested
		state.AnalysisOnly = true
	}

	/*
		When a client disconnects from the server, the server reports this up to us via the "ClientPingTimedOut" flag
		When this happens we don't need to do any complicated route logic, just exit early.
	*/

	if state.Request.ClientPingTimedOut {
		core.Debug("client ping timed out")
		state.ClientPingTimedOut = true // tested
		return true
	}

	/*
		On the initial slice, we look up the lat/long for the player using ip2location.

		On subsequent slices, we use the cached location data from the session state.
	*/

	if state.Request.SliceNumber == 0 {

		var err error

		state.LocatedIP = true // tested

		state.Output.Location, err = state.LocateIP(state.Request.ClientAddress.IP)

		if err != nil {
			core.Error("location veto: %s", err)
			state.Output.RouteState.LocationVeto = true // tested
			state.LocationVeto = true
			return true
		}

		state.Input.Location = state.Output.Location // tested

	} else {

		// use location data stored in session data

		readStream := encoding.CreateReadStream(state.Request.SessionData[:])

		err := state.Input.Serialize(readStream)
		if err != nil {
			core.Debug("failed to read session data: %v", err)
			state.FailedToReadSessionData = true // tested
			return true
		}

		state.ReadSessionData = true // tested

		state.Output.Location = state.Input.Location
	}

	/*
		Routing with an old route matrix runs a serious risk of sending players across routes that are WORSE
		than their default internet route, so it's best to just go direct if the route matrix is stale.
	*/

	if state.RouteMatrix.CreatedAt+uint64(state.StaleDuration.Seconds()) < uint64(time.Now().Unix()) {
		core.Debug("stale route matrix")
		state.StaleRouteMatrix = true // tested
		return true
	}

	/*
		Check if the datacenter is unknown, and flag it.

		This is important so that we can quickly check if we need to add new datacenters for customers.
	*/

	if !datacenterExists(state.Database, state.Request.DatacenterId) {
		core.Debug("unknown datacenter")
		state.UnknownDatacenter = true // tested
	}

	/*
		Check if the datacenter is enabled for this customer.

		If the datacenter is not enabled, we just wont accelerate the player.
	*/

	if !datacenterEnabled(state.Database, state.Request.BuyerId, state.Request.DatacenterId) {
		core.Debug("datacenter not enabled: %x, %x", state.Request.BuyerId, state.Request.DatacenterId)
		state.DatacenterNotEnabled = true // tested
	}

	/*
		Get the datacenter information and store it in the handler state.

		If the datacenter is unknown, this datacenter will have a zero id and be named "unknown".
	*/

	state.Datacenter = getDatacenter(state.Database, state.Request.DatacenterId)

	/*
		Get the set of relay ids that are in the destination datacenter (if applicable).

		If anything goes wrong, this is an empty set.
	*/

	destRelayIds := state.RouteMatrix.GetDatacenterRelays(state.Request.DatacenterId)
	if len(destRelayIds) == 0 {
		core.Debug("no relays in datacenter %x", state.Request.DatacenterId)
		state.NoRelaysInDatacenter = true
	}

	/*
		Check for various tags on the first slice only (tags are only set on the first slice).

		These tags enable specific behavior, like "pro" mode (accelerate always)

		It's an easy way for our customers to indicate that certain sessions should be treated differently.
	*/

	const ProTag = 0x77FD571956A1F7F8

	if state.Request.SliceNumber == 0 {
		for i := int32(0); i < state.Request.NumTags; i++ {
			if state.Request.Tags[i] == ProTag {
				core.Debug("pro mode enabled")
				state.Buyer.RouteShader.ProMode = true // tested
				state.Pro = true
			}
		}
	}

	/*
		The debug string is appended to during the rest of the handler and sent down to the SDK
		when Buyer.Debug is true. We use this to debug route decisions when something is not working.
	*/

	if state.Buyer.Debug {
		core.Debug("debug enabled")
		state.Debug = new(string) // tested
	}

	return false
}

func SessionUpdate_NewSession(state *SessionUpdateState) {

	core.Debug("new session")

	state.Output.Version = packets.SDK5_SessionDataVersion_Write
	state.Output.SessionId = state.Request.SessionId
	state.Output.SliceNumber = 1
	state.Output.ExpireTimestamp = uint64(time.Now().Unix()) + packets.SDK5_BillingSliceSeconds
	state.Output.RouteState.UserID = state.Request.UserHash
	state.Output.RouteState.ABTest = state.Buyer.RouteShader.ABTest

	state.Input = state.Output
}

func SessionUpdate_ExistingSession(state *SessionUpdateState) {

	core.Debug("existing session")

	/*
	   Read the session data, if it has not already been read.

	   This data contains state that persists across the session, it is sent up from the SDK,
	   we transform it, and then send it back down -- and it gets sent up by the SDK in the next
	   update.

	   This way we don't have to store state per-session in the backend.
	*/

	// todo: the session data must not be modifiable by the client. where is the check that ensures this is the case?

	if !state.ReadSessionData {

		readStream := encoding.CreateReadStream(state.Request.SessionData[:])

		err := state.Input.Serialize(readStream)
		if err != nil {
			core.Debug("failed to read session data: %v", err)
			state.FailedToReadSessionData = true
			return
		}

		state.ReadSessionData = true
	}

	/*
	   Check for some obviously divergent data between the session request packet
	   and the stored session data. If there is a mismatch, just return a direct route.
	*/

	if state.Input.SessionId != state.Request.SessionId {
		core.Debug("bad session id")
		state.BadSessionId = true
		return
	}

	if state.Input.SliceNumber != state.Request.SliceNumber {
		core.Debug("bad slice number")
		state.BadSliceNumber = true
		return
	}

	/*
	   Copy input state to output and go to next slice.

	   During the rest of the session update we transform session.output in place,
	   before sending it back to the SDK in the session response packet.
	*/

	state.Output = state.Input
	state.Output.Initial = false
	state.Output.SliceNumber += 1
	state.Output.ExpireTimestamp += packets.SDK5_BillingSliceSeconds

	/*
	   Calculate real packet loss.

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

	state.PostRealPacketLossClientToServer = RealPacketLossClientToServer
	state.PostRealPacketLossServerToClient = RealPacketLossServerToClient

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

func SessionUpdate_HandleFallbackToDirect(state *SessionUpdateState) bool {

	/*
	   Fallback to direct is a state where the SDK has met some fatal error condition.

	   When this happens, the session will go direct from that point forward.
	*/

	if state.Request.FallbackToDirect && !state.Output.FallbackToDirect {
		core.Debug("fallback to direct")
		state.Output.FallbackToDirect = true
		state.FallbackToDirect = true
		return true
	}

	return false
}

func SessionUpdate_GetNearRelays(state *SessionUpdateState) bool {

	/*
	   This function selects up to 32 near relays for the session,
	   according to the players latitude and longitude determined by
	   ip2location.

	   These near relays are selected only on the first slice (slice 0)
	   of a session, and are held fixed for the duration of the session.

	   The SDK pings the near relays, and reports up the latency, jitter
	   and packet loss to each near relay, with each subsequent session
	   update (every 10 seconds).

	   Network Next uses the relay ping statistics in route planning,
	   by adding the latency to the first relay to the total route cost,
	   and by excluding near relays with higher jitter or packet loss
	   than the default internet route.

	   This function is skipped for "Analysis Only" buyers because sessions
	   will always take direct.

	   This function is skipped for datacenters that are not enabled for
	   acceleration, forcing all connected clients to go direct.
	*/

	if state.Buyer.RouteShader.AnalysisOnly {
		core.Debug("analysis only, not getting near relays")
		state.NotGettingNearRelaysAnalysisOnly = true
		return false
	}

	if !state.DatacenterNotEnabled {
		core.Debug("datacenter not enabled, not getting near relays")
		state.NotGettingNearRelaysDatacenterAccelerationDisabled = true
		return false
	}

	clientLatitude := state.Output.Location.Latitude
	clientLongitude := state.Output.Location.Longitude

	serverLatitude := state.Datacenter.Latitude
	serverLongitude := state.Datacenter.Longitude

	const distanceThreshold = 2500
	const latencyThreshold = 30.0

	nearRelayIds, nearRelayAddresses := common.GetNearRelays(core.MaxNearRelays,
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
		state.NoNearRelays = true
		return false
	}

	for i := 0; i < numNearRelays; i++ {
		state.Response.NearRelayIds[i] = nearRelayIds[i]
		state.Response.NearRelayAddresses[i] = nearRelayAddresses[i]
	}

	state.Response.NumNearRelays = int32(numNearRelays)
	state.Response.HighFrequencyPings = state.Buyer.InternalConfig.HighFrequencyPings && !state.Buyer.InternalConfig.LargeCustomer
	state.Response.NearRelaysChanged = true

	return true
}

func SessionUpdate_UpdateNearRelays(state *SessionUpdateState) bool {

	/*
	   This function is CalculateNextBytesUpAndDown once every 10 seconds for all slices
	   in a session after slice 0 (first slice).

	   It takes the ping statistics for each near relay, and collates them
	   into a format suitable for route planning later on in the session
	   update.

	   It also runs various filters inside core.ReframeRelays, which look at
	   the history of latency, jitter and packet loss across the entire session
	   in order to exclude near relays with bad performance from being selected.

	   This function exits early if the session will not be accelerated.
	*/

	routeShader := &state.Buyer.RouteShader

	if state.AnalysisOnly {
		core.Debug("analysis only, not updating near relay stats")
		return false
	}

	if state.DatacenterNotEnabled {
		core.Debug("datacenter not disabled, not updating near relay stats")
		return false
	}

	destRelayIds := state.RouteMatrix.GetDatacenterRelays(state.Datacenter.ID)

	if len(destRelayIds) == 0 {
		core.Debug("no relays in datacenter %x", state.Datacenter.ID)
		state.NoRelaysInDatacenter = true
		return false
	}

	state.DestRelays = make([]int32, len(destRelayIds))

	/*
	   If we are holding near relays, use the held near relay RTT as input
	   instead of the near relay ping data sent up from the SDK.
	*/

	if state.Input.HoldNearRelays {
		core.Debug("using held near relay RTTs")
		for i := range state.Request.NearRelayIds {
			state.Request.NearRelayRTT[i] = state.Input.HoldNearRelayRTT[i] // when set to 255, near relay is excluded from routing
			state.Request.NearRelayJitter[i] = 0
			state.Request.NearRelayPacketLoss[i] = 0
		}
		state.UsingHeldNearRelays = true
	}

	/*
	   Reframe the near relays to get them in a relay index form relative to the current route matrix.
	*/

	routeState := &state.Output.RouteState

	directLatency := int32(math.Ceil(float64(state.Request.DirectMinRTT)))
	directJitter := int32(math.Ceil(float64(state.Request.DirectJitter)))
	directPacketLoss := int32(math.Floor(float64(state.Request.DirectPacketLoss) + 0.5))
	nextPacketLoss := int32(math.Floor(float64(state.Request.NextPacketLoss) + 0.5))

	numNearRelays := state.Request.NumNearRelays

	core.ReframeRelays(

		// input
		routeShader,
		routeState,
		state.RouteMatrix.RelayIdToIndex,
		directLatency,
		directJitter,
		directPacketLoss,
		nextPacketLoss,
		int32(state.Request.SliceNumber),
		state.Request.NearRelayIds[:numNearRelays],
		state.Request.NearRelayRTT[:numNearRelays],
		state.Request.NearRelayJitter[:numNearRelays],
		state.Request.NearRelayPacketLoss[:numNearRelays],
		destRelayIds,

		// output
		state.NearRelayRTTs[:],
		state.NearRelayJitters[:],
		&state.NumDestRelays,
		state.DestRelays,
	)

	state.NumNearRelays = int(numNearRelays)

	for i := range state.Request.NearRelayIds {
		relayIndex, exists := state.RouteMatrix.RelayIdToIndex[state.Request.NearRelayIds[i]]
		if exists {
			state.NearRelayIndices[i] = int32(relayIndex)
		} else {
			state.NearRelayIndices[i] = -1 // near relay no longer exists in route matrix
		}
	}

	SessionUpdate_FilterNearRelays(state) // IMPORTANT: Reduce % of sessions that run near relay pings for large customers

	return true
}

func SessionUpdate_FilterNearRelays(state *SessionUpdateState) {

	/*
	   Reduce the % of sessions running near relay pings for large customers.

	   We do this by only running near relay pings for the first 3 slices, and then holding
	   the near relay ping results fixed for the rest of the session.
	*/

	if !state.Buyer.InternalConfig.LargeCustomer {
		return
	}

	state.LargeCustomer = true

	if state.Request.SliceNumber < 4 {
		return
	}

	// IMPORTANT: On any slice after 4, if we haven't already, grab the *processed* (255 if not routable)
	// near relay RTTs from ReframeRelays and hold them as the near relay RTTs to use from now on.

	if !state.Input.HoldNearRelays {
		core.Debug("holding near relays")
		state.Output.HoldNearRelays = true
		state.HoldingNearRelays = true
		for i := 0; i < len(state.Request.NearRelayIds); i++ {
			state.Output.HoldNearRelayRTT[i] = state.NearRelayRTTs[i]
		}
	}

	// tell the SDK to stop pinging near relays

	state.Response.ExcludeNearRelays = true
	for i := 0; i < core.MaxNearRelays; i++ {
		state.Response.NearRelayExcluded[i] = true
	}
	state.NearRelaysExcluded = true
}

func SessionUpdate_BuildNextTokens(state *SessionUpdateState, routeNumRelays int32, routeRelays []int32) {

	state.Output.ExpireTimestamp += packets.SDK5_BillingSliceSeconds
	state.Output.SessionVersion++
	state.Output.Initial = true

	numTokens := routeNumRelays + 2

	var routeAddresses [core.NEXT_MAX_NODES]*net.UDPAddr
	var routePublicKeys [core.NEXT_MAX_NODES][]byte

	// client node (no address specified...)

	routePublicKeys[0] = state.Request.ClientRoutePublicKey[:]

	// relay nodes

	relayAddresses := routeAddresses[1 : numTokens-1]
	relayPublicKeys := routePublicKeys[1 : numTokens-1]

	numRouteRelays := len(routeRelays)

	for i := 0; i < numRouteRelays; i++ {

		relayIndex := routeRelays[i]

		relayId := state.RouteMatrix.RelayIds[relayIndex]

		relay, _ := state.Database.RelayMap[relayId]

		relayAddresses[i] = &relay.Addr

		// use private address (when it exists) when sending between two relays belonging to the same seller
		if i > 0 {
			prevRelayIndex := routeRelays[i-1]
			prevId := state.RouteMatrix.RelayIds[prevRelayIndex]
			prev, _ := state.Database.RelayMap[prevId]
			if prev.Seller.ID == relay.Seller.ID && relay.InternalAddr.String() != ":0" {
				relayAddresses[i] = &relay.InternalAddr
			}
		}

		relayPublicKeys[i] = relay.PublicKey
	}

	// server node

	routeAddresses[numTokens-1] = state.From
	routePublicKeys[numTokens-1] = state.Request.ServerRoutePublicKey[:]

	// debug print the route

	core.Debug("----------------------------------------------------")
	for index, address := range routeAddresses[:numTokens] {
		core.Debug("route address %d: %s", index, address.String())
	}
	core.Debug("----------------------------------------------------")

	// write the tokens

	tokenData := make([]byte, numTokens*packets.SDK5_EncryptedNextRouteTokenSize)

	sessionId := state.Output.SessionId
	sessionVersion := uint8(state.Output.SessionVersion)
	expireTimestamp := state.Output.ExpireTimestamp
	envelopeUpKbps := uint32(state.Buyer.RouteShader.BandwidthEnvelopeUpKbps)
	envelopeDownKbps := uint32(state.Buyer.RouteShader.BandwidthEnvelopeDownKbps)

	core.WriteRouteTokens(tokenData, expireTimestamp, sessionId, sessionVersion, envelopeUpKbps, envelopeDownKbps, int(numTokens), routeAddresses[:], routePublicKeys[:], state.RoutingPrivateKey[:])

	state.Response.RouteType = packets.SDK5_RouteTypeNew
	state.Response.NumTokens = numTokens
	state.Response.Tokens = tokenData
}

func SessionUpdate_BuildContinueTokens(state *SessionUpdateState, routeNumRelays int32, routeRelays []int32) {

	numTokens := routeNumRelays + 2

	var routePublicKeys [core.NEXT_MAX_NODES][]byte

	// client node

	routePublicKeys[0] = state.Request.ClientRoutePublicKey[:]

	// relay nodes

	relayPublicKeys := routePublicKeys[1 : numTokens-1]

	numRouteRelays := len(routeRelays)

	for i := 0; i < numRouteRelays; i++ {

		relayIndex := routeRelays[i]

		relayId := state.RouteMatrix.RelayIds[relayIndex]

		relay, _ := state.Database.RelayMap[relayId]

		relayPublicKeys[i] = relay.PublicKey
	}

	// server node

	routePublicKeys[numTokens-1] = state.Request.ServerRoutePublicKey[:]

	// build the tokens

	tokenData := make([]byte, numTokens*packets.SDK5_EncryptedContinueRouteTokenSize)

	sessionId := state.Output.SessionId
	sessionVersion := uint8(state.Output.SessionVersion)
	expireTimestamp := state.Output.ExpireTimestamp

	core.WriteContinueTokens(tokenData, expireTimestamp, sessionId, sessionVersion, int(numTokens), routePublicKeys[:], state.RoutingPrivateKey[:])

	state.Response.RouteType = packets.SDK5_RouteTypeContinue
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
		state.NoRouteRelays = true
		return
	}

	var stayOnNext bool
	var routeChanged bool
	var routeCost int32
	var routeNumRelays int32

	routeRelays := [core.MaxRelaysPerRoute]int32{}

	sliceNumber := int32(state.Request.SliceNumber)

	if !state.Input.RouteState.Next {

		// currently going direct. should we take network next?

		if core.MakeRouteDecision_TakeNetworkNext(state.RouteMatrix.RouteEntries, state.RouteMatrix.FullRelayIndexSet, &state.Buyer.RouteShader, &state.Output.RouteState, &state.Buyer.InternalConfig, int32(state.Request.DirectMinRTT), state.RealPacketLoss, state.NearRelayIndices[:], state.NearRelayRTTs[:], state.DestRelays, &routeCost, &routeNumRelays, routeRelays[:], &state.RouteDiversity, state.Debug, sliceNumber) {

			state.TakeNetworkNext = true

			SessionUpdate_BuildNextTokens(state, routeNumRelays, routeRelays[:routeNumRelays])

			if state.Debug != nil {

				*state.Debug += "route relays: "

				for i, routeRelay := range routeRelays[:routeNumRelays] {
					if i != int(routeNumRelays-1) {
						*state.Debug += fmt.Sprintf("%s - ", state.RouteMatrix.RelayNames[routeRelay])
					} else {
						*state.Debug += fmt.Sprintf("%s\n", state.RouteMatrix.RelayNames[routeRelay])
					}
				}
			}
		}

	} else {

		// currently taking network next

		if !state.Request.Next {

			// the sdk aborted this session

			core.Debug("aborted")
			state.Output.RouteState.Next = false
			state.Output.RouteState.Veto = true
			state.Aborted = true
			return
		}

		/*
		   Reframe the current route in terms of relay indices in the current route matrix

		   This is necessary because the set of relays in the route matrix change over time.
		*/

		if !core.ReframeRoute(&state.Output.RouteState, state.RouteMatrix.RelayIdToIndex, state.Output.RouteRelayIds[:state.Output.RouteNumRelays], &routeRelays) {
			routeRelays = [core.MaxRelaysPerRoute]int32{}
			core.Debug("one or more relays in the route no longer exist")
			state.RouteRelayNoLongerExists = true
		}

		stayOnNext, routeChanged = core.MakeRouteDecision_StayOnNetworkNext(state.RouteMatrix.RouteEntries, state.RouteMatrix.FullRelayIndexSet, state.RouteMatrix.RelayNames, &state.Buyer.RouteShader, &state.Output.RouteState, &state.Buyer.InternalConfig, int32(state.Request.DirectMinRTT), int32(state.Request.NextRTT), state.Output.RouteCost, state.RealPacketLoss, state.Request.NextPacketLoss, state.Output.RouteNumRelays, routeRelays, state.NearRelayIndices[:], state.NearRelayRTTs[:], state.DestRelays[:], &routeCost, &routeNumRelays, routeRelays[:], state.Debug)

		if stayOnNext {

			// stay on network next

			if routeChanged {

				core.Debug("route changed")
				state.RouteChanged = true
				SessionUpdate_BuildNextTokens(state, routeNumRelays, routeRelays[:routeNumRelays])

			} else {

				core.Debug("route continued")
				state.RouteContinued = true
				SessionUpdate_BuildContinueTokens(state, routeNumRelays, routeRelays[:routeNumRelays])

			}

		} else {

			// leave network next

			if state.Output.RouteState.NoRoute {
				core.Debug("route no longer exists")
				state.RouteNoLongerExists = true
			}

			if state.Output.RouteState.Mispredict {
				core.Debug("mispredict")
				state.Mispredict = true
			}

			if state.Output.RouteState.LatencyWorse {
				core.Debug("latency worse")
				state.LatencyWorse = true
			}
		}
	}

	/*
	   Stash key route parameters in the response so the SDK recieves them.

	   Committed means to actually send packets across the network next route,
	   if false, then the route just has ping packets sent across it, but no
	   game packets.

	   Multipath means to send packets across both the direct and the network
	   next route at the same time, which reduces packet loss.
	*/

	state.Response.Committed = state.Output.RouteState.Committed
	state.Response.Multipath = state.Output.RouteState.Multipath

	/*
	   Stick the route cost, whether the route changed, and the route relay data
	   in the output state. This output state is serialized into the route state
	   in the route response, and sent back up to us, allowing us to know the
	   current network next route, when we plan the next 10 second slice.
	*/

	if routeCost > common.InvalidRouteValue {
		routeCost = common.InvalidRouteValue
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
	   Build the set of near relays for the SDK to ping.

	   The SDK pings these near relays and reports up the results in the next session update.

	   We hold the set of near relays fixed for the session, so we only do this work on the first slice.
	*/

	if state.Request.SliceNumber == 0 {
		SessionUpdate_GetNearRelays(state)
		core.Debug("first slice always goes direct")
	}

	/*
	   Since post runs at the end of every session handler, run logic
	   here that must run if we are taking network next vs. direct
	*/

	if state.Response.RouteType != packets.SDK5_RouteTypeDirect {
		core.Debug("session takes network next")
	} else {
		core.Debug("session goes direct")
	}

	/*
	   Track duration of time spent on network next, and if the session has ever been on network next.
	*/

	if state.Request.Next {
		state.Output.EverOnNext = true
		state.Output.DurationOnNext += packets.SDK5_BillingSliceSeconds
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
	   The session ends when the client ping times out.

	   At this point we write a summary slice to bigquery, with more information than regular slices.

	   This saves a lot of bandwidth and bigquery cost, by only writing this information once per-session.
	*/

	if state.Request.ClientPingTimedOut {
		state.Output.WroteSummary = true
	}

	/*
	   Write the session update response packet and send it back to the caller.
	*/

	packetData, err := packets.SDK5_WritePacket(&state.Response, packets.SDK5_SESSION_UPDATE_RESPONSE_PACKET, packets.SDK5_MaxPacketBytes, state.ServerBackendAddress, state.From, state.RoutingPrivateKey[:])
	if err != nil {
		core.Error("failed to write session update response packet: %v", err)
		return
	}

	if err == nil {

		if _, err := state.Connection.WriteToUDP(packetData, state.From); err != nil {
			core.Error("failed to send session update response packet: %v", err)
			state.FailedToSendResponsePacket = true
			return
		}

		state.WroteResponsePacket = true

	} else {

		core.Error("failed to write response packet: %v", err)

		state.FailedToWriteResponsePacket = true

	}

	/*
		Build the data for the relays in the route.
	*/

	buildRouteRelayData(state)

	/*
		Build the data for the near relays.
	*/

	buildNearRelayData(state)

	/*
		Send this slice to the portal via the real-time path (redis streams).
	*/

	sendPortalData(state)

	/*
		Send this slice billing system (bigquery) via the non-realtime path (google pubsub).
	*/

	sendSessionUpdateMessage(state)
}

// -----------------------------------------

func datacenterExists(database *db.Database, datacenterId uint64) bool {
	_, exists := database.DatacenterMap[datacenterId]
	return exists
}

func datacenterEnabled(database *db.Database, buyerId uint64, datacenterId uint64) bool {
	datacenterAliases, ok := database.DatacenterMaps[buyerId]
	if !ok {
		return false
	}
	_, ok = datacenterAliases[datacenterId]
	return ok
}

func getDatacenter(database *db.Database, datacenterId uint64) db.Datacenter {
	value, exists := database.DatacenterMap[datacenterId]
	if !exists {
		return db.Datacenter{
			Name: "unknown",
		}
	}
	return value
}

func buildRouteRelayData(state *SessionUpdateState) {

	for i := int32(0); i < state.Input.RouteNumRelays; i++ {
		relay, ok := state.Database.RelayMap[state.Input.RouteRelayIds[i]]
		if ok {
			state.PostRouteRelayNames[i] = relay.Name
			state.PostRouteRelaySellers[i] = relay.Seller
		}
	}
}

func buildNearRelayData(state *SessionUpdateState) {

	// todo

}

func sendPortalData(state *SessionUpdateState) {

	// no point sending data to the portal, once the client has timed out

	if state.Request.ClientPingTimedOut {
		return
	}

	state.SentPortalData = true

	// todo
	/*
		portalData := buildPortalData(state)

		if portalData.Meta.NextRTT != 0 || portalData.Meta.DirectRTT != 0 {
			state.PostSessionHandler.SendPortalData(portalData)
		}
	*/
}

func sendSessionUpdateMessage(state *SessionUpdateState) {

	// todo

	state.SentSessionUpdateMessage = true

}
