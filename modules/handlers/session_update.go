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
	RelayBackendPublicKey   []byte
	RelayBackendPrivateKey  []byte
	ServerBackendAddress    *net.UDPAddr
	ServerBackendPrivateKey []byte
	ServerBackendPublicKey  []byte

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

	// for route planning
	DestRelayIds   []uint64
	DestRelays     []int32
	SourceRelays   []int32
	SourceRelayRTT []int32

	// error flags for this update
	Error uint64

	// ip2location function to get ISP and country from IP
	GetISPAndCountry func(ip net.IP) (string, string)

	// track start time of handler
	StartTimestamp     uint64
	StartTimestampNano uint64

	// if true, only network next sessions are sent to portal
	PortalNextSessionsOnly bool

	// codepath flags (for unit testing etc...)
	ClientPingTimedOut                          bool
	RouteChanged                                bool
	RouteContinued                              bool
	TakeNetworkNext                             bool
	StayDirect                                  bool
	FirstUpdate                                 bool
	ReadSessionData                             bool
	NotUpdatingClientRelaysDatacenterNotEnabled bool
	NotUpdatingServerRelaysDatacenterNotEnabled bool
	SentPortalSessionUpdateMessage              bool
	SentPortalClientRelayUpdateMessage          bool
	SentPortalServerRelayUpdateMessage          bool
	SentAnalyticsClientRelayPingMessage         bool
	SentAnalyticsServerRelayPingMessage         bool
	SentAnalyticsSessionUpdateMessage           bool
	SentAnalyticsSessionSummaryMessage          bool
	WroteResponsePacket                         bool
	LongSessionUpdate                           bool

	FallbackToDirectChannel chan<- uint64

	PortalSessionUpdateMessageChannel     chan<- *messages.PortalSessionUpdateMessage
	PortalClientRelayUpdateMessageChannel chan<- *messages.PortalClientRelayUpdateMessage
	PortalServerRelayUpdateMessageChannel chan<- *messages.PortalServerRelayUpdateMessage

	AnalyticsSessionUpdateMessageChannel   chan<- *messages.AnalyticsSessionUpdateMessage
	AnalyticsSessionSummaryMessageChannel  chan<- *messages.AnalyticsSessionSummaryMessage
	AnalyticsClientRelayPingMessageChannel chan<- *messages.AnalyticsClientRelayPingMessage
	AnalyticsServerRelayPingMessageChannel chan<- *messages.AnalyticsServerRelayPingMessage
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

	state.StartTimestamp = state.StartTimestampNano / 1000000000 // nano -> seconds

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

	if state.Request.SliceNumber >= 1 {
		score := core.GetSessionScore(int32(state.Request.DirectRTT), int32(state.Request.NextRTT))
		if uint32(score) <= state.Input.BestScore {
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
		state.Input.BestScore = uint32(constants.MaxScore)
	}

	/*
		Fallback to direct is a state where the SDK has met some fatal error condition.

		When this happens, the session will go direct from this point forward.
	*/

	if state.Request.FallbackToDirect {
		state.Error |= constants.SessionError_FallbackToDirect
		datacenter := state.Database.GetDatacenter(state.Request.DatacenterId)
		if datacenter != nil {
			core.Debug("fallback to direct: session_id = %016x, datacenter = %s [%016x]", state.Request.SessionId, datacenter.Name, state.Request.DatacenterId)
		} else {
			core.Debug("fallback to direct: session_id = %016x, datacenter = %016x", state.Request.SessionId, state.Request.DatacenterId)
		}
		if state.FallbackToDirectChannel != nil {
			state.FallbackToDirectChannel <- state.Request.SessionId
		}
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
	state.Output.ExpireTimestamp = state.Input.ExpireTimestamp + packets.SDK_SliceSeconds*2 + 5
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

func SessionUpdate_UpdateClientRelays(state *SessionUpdateState) bool {

	if (state.Error & constants.SessionError_DatacenterNotEnabled) != 0 {
		core.Debug("datacenter not enabled, not updating client relays")
		state.NotUpdatingClientRelaysDatacenterNotEnabled = true
		return false
	}

	/*
		If the client relays have changed, filter them to work out which ones need to be excluded moving forward.

		These client relays will be excluded from routing, until the client relays change again.
	*/

	numClientRelays := state.Request.NumClientRelays

	sourceRelayIds := state.Request.ClientRelayIds[:numClientRelays]
	sourceRelayLatency := state.Request.ClientRelayRTT[:numClientRelays]
	sourceRelayJitter := state.Request.ClientRelayJitter[:numClientRelays]
	sourceRelayPacketLoss := state.Request.ClientRelayPacketLoss[:numClientRelays]

	if state.Request.ClientRelayPingsHaveChanged {

		directLatency := int32(math.Ceil(float64(state.Request.DirectRTT)))
		directJitter := int32(math.Ceil(float64(state.Request.DirectJitter)))
		directPacketLoss := state.Request.DirectMaxPacketLossSeen

		core.FilterSourceRelays(
			directLatency,
			directJitter,
			directPacketLoss,
			sourceRelayIds,
			sourceRelayLatency,
			sourceRelayJitter,
			sourceRelayPacketLoss,
			state.Output.ExcludeClientRelay[:],
		)

		if core.DebugLogs && len(sourceRelayIds) > 0 {
			core.Debug("------------------------------------------------------------------------------------------------")
			core.Debug("client relays changed:")
			for i := range sourceRelayIds {
				if state.Output.ExcludeClientRelay[i] {
					core.Debug("%d: %016x | rtt = %d, jitter = %d, packet loss = %.2f (excluded)", i, sourceRelayIds[i], sourceRelayLatency[i], sourceRelayJitter[i], sourceRelayPacketLoss[i])
				} else {
					core.Debug("%d: %016x | rtt = %d, jitter = %d, packet loss = %.2f", i, sourceRelayIds[i], sourceRelayLatency[i], sourceRelayJitter[i], sourceRelayPacketLoss[i])
				}
			}
			core.Debug("------------------------------------------------------------------------------------------------")
		}

		/*
			If all client relays are > 60ms RTT, this is likely a VPN or cross-region session

			If we don't find any valid client relays, set a flag.

			If we don't find any non-zero client relays, set a flag.
		*/

	    foundValidRelay := false
		foundLowLatency := false
		foundNonZeroRelay := false

		for i := 0; i < len(sourceRelayIds); i++ {
			if sourceRelayLatency[i] != 0 || sourceRelayPacketLoss[i] != 0 {
				foundNonZeroRelay = true
			}
			if state.Output.ExcludeClientRelay[i] {
				continue
			}
			foundValidRelay = true
			if sourceRelayLatency[i] <= 60 {
				foundLowLatency = true
			}
		}

		if !foundLowLatency && foundValidRelay {
			core.Debug("session %016x is likely vpn or cross region", state.Request.SessionId)
			state.Output.LikelyVPNOrCrossRegion = true
		}

		if !foundValidRelay {
			core.Debug("session %016x has no client relays", state.Request.SessionId)
			state.Output.NoClientRelays = true
		}

		if !foundNonZeroRelay {
			core.Debug("session %016x client relays are all zero", state.Request.SessionId)
			state.Output.AllClientRelaysAreZero = true;
		}
	}

	/*
		Reframe the client relays from relay ids to relay indices relative to the current route matrix.

		The route matrix and the position of relays in it can change over time, thus we need to perform this step each update.

		At the same time we also set any source relays that are excluded, or do not exist in the route matrix to have RTT = 255.

		(RTT 255 from this point forward means "do not route through this source relay")
	*/

	outputSourceRelays := make([]int32, numClientRelays)
	outputSourceRelayRTT := make([]int32, numClientRelays)

	core.ReframeSourceRelays(state.RouteMatrix.RelayIdToIndex, sourceRelayIds, sourceRelayLatency[:], state.Output.ExcludeClientRelay[:], outputSourceRelays, outputSourceRelayRTT)

	state.SourceRelays = outputSourceRelays
	state.SourceRelayRTT = outputSourceRelayRTT

	return true
}

func SessionUpdate_UpdateServerRelays(state *SessionUpdateState) bool {

	if (state.Error & constants.SessionError_DatacenterNotEnabled) != 0 {
		core.Debug("datacenter not enabled, not updating server relay stats")
		state.NotUpdatingServerRelaysDatacenterNotEnabled = true
		return false
	}

	/*
		If the server relays have changed, filter them to work out which ones need to be excluded moving forward.

		These server relays will be excluded from routing, until the server relays change again.
	*/

	numServerRelays := state.Request.NumServerRelays

	destRelayIds := state.Request.ServerRelayIds[:numServerRelays]
	destRelayLatency := state.Request.ServerRelayRTT[:numServerRelays]
	destRelayJitter := state.Request.ServerRelayJitter[:numServerRelays]
	destRelayPacketLoss := state.Request.ServerRelayPacketLoss[:numServerRelays]

	if state.Request.ServerRelayPingsHaveChanged {

		core.FilterDestRelays(
			destRelayIds,
			destRelayLatency,
			destRelayJitter,
			destRelayPacketLoss,
			state.Output.ExcludeServerRelay[:],
		)

		if core.DebugLogs && len(destRelayIds) > 0 {
			core.Debug("------------------------------------------------------------------------------------------------")
			core.Debug("server relays changed:")
			for i := range destRelayIds {
				if state.Output.ExcludeServerRelay[i] {
					core.Debug("%d: %016x | rtt = %d, jitter = %d, packet loss = %.2f (excluded)", i, destRelayIds[i], destRelayLatency[i], destRelayJitter[i], destRelayPacketLoss[i])
				} else {
					core.Debug("%d: %016x | rtt = %d, jitter = %d, packet loss = %.2f", i, destRelayIds[i], destRelayLatency[i], destRelayJitter[i], destRelayPacketLoss[i])
				}
			}
			core.Debug("------------------------------------------------------------------------------------------------")
		}

	    foundValidRelay := false
		for i := 0; i < len(destRelayIds); i++ {
			if !state.Output.ExcludeServerRelay[i] {
				foundValidRelay = true
				break
			}
		}

		if !foundValidRelay {
			core.Debug("session %016x has no server relays", state.Request.SessionId)
			state.Output.NoServerRelays = true
		}
	}

	/*
		Reframe the server relays from relay ids to relay indices relative to the current route matrix.

		The route matrix and the position of relays in it can change over time, thus we need to perform this step each update.

		This step also excludes dest relays that are marked to be excluded, or no longer exist in the route matrix.
	*/

	outputDestRelays := make([]int32, 0, numServerRelays)

	core.ReframeDestRelays(state.RouteMatrix.RelayIdToIndex, destRelayIds, state.Output.ExcludeServerRelay[:], &outputDestRelays)

	state.DestRelays = outputDestRelays

	return true
}

func SessionUpdate_BuildNextTokens(state *SessionUpdateState, routeNumRelays int32, routeRelays []int32) {

	state.Output.SessionVersion++

	numTokens := routeNumRelays + 2

	var routePublicAddresses [constants.NextMaxNodes]net.UDPAddr
	var routeHasInternalAddresses [constants.NextMaxNodes]bool
	var routeInternalAddresses [constants.NextMaxNodes]net.UDPAddr
	var routeInternalGroups [constants.NextMaxNodes]uint64
	var routeSellers [constants.NextMaxNodes]int
	var routeSecretKeys [constants.NextMaxNodes][]byte

	// client node

	routeSecretKeys[0], _ = crypto.SecretKey_GenerateRemote(state.RelayBackendPublicKey, state.RelayBackendPrivateKey, state.Request.ClientRoutePublicKey[:])
	routePublicAddresses[0] = state.Request.ClientAddress
	routePublicAddresses[0].Port = 0 // IMPORTANT: Set client port to zero, it will be replaced with whatever port is in from addr by the relay

	// relay nodes

	relayPublicAddresses := routePublicAddresses[1 : numTokens-1]
	relayHasInternalAddresses := routeHasInternalAddresses[1 : numTokens-1]
	relayInternalAddresses := routeInternalAddresses[1 : numTokens-1]
	relayInternalGroups := routeInternalGroups[1 : numTokens-1]
	relaySellers := routeSellers[1 : numTokens-1]
	relaySecretKeys := routeSecretKeys[1 : numTokens-1]

	numRouteRelays := len(routeRelays)

	for i := 0; i < numRouteRelays; i++ {

		relayIndex := routeRelays[i]

		relay := &state.Database.Relays[relayIndex]

		relayPublicAddresses[i] = relay.PublicAddress
		relayHasInternalAddresses[i] = relay.HasInternalAddress
		relayInternalAddresses[i] = relay.InternalAddress
		relayInternalGroups[i] = relay.InternalGroup
		relaySellers[i] = int(relay.Seller.Id)
		relaySecretKeys[i] = state.Database.RelaySecretKeys[relay.Id]
	}

	// server node

	routePublicAddresses[numTokens-1] = *state.From
	routeSecretKeys[numTokens-1], _ = crypto.SecretKey_GenerateRemote(state.RelayBackendPublicKey, state.RelayBackendPrivateKey, state.Request.ServerRoutePublicKey[:])

	// write the tokens

	tokenData := make([]byte, numTokens*packets.SDK_EncryptedNextRouteTokenSize)

	sessionId := state.Output.SessionId
	sessionVersion := uint8(state.Output.SessionVersion)
	expireTimestamp := state.Output.ExpireTimestamp
	envelopeUpKbps := uint32(state.Buyer.RouteShader.BandwidthEnvelopeUpKbps)
	envelopeDownKbps := uint32(state.Buyer.RouteShader.BandwidthEnvelopeDownKbps)

	core.WriteRouteTokens(tokenData, expireTimestamp, sessionId, sessionVersion, envelopeUpKbps, envelopeDownKbps, int(numTokens), routePublicAddresses[:], routeHasInternalAddresses[:], routeInternalAddresses[:], routeInternalGroups[:], routeSellers[:], routeSecretKeys[:])

	state.Response.RouteType = packets.SDK_RouteTypeNew
	state.Response.NumTokens = numTokens
	state.Response.Tokens = tokenData
}

func SessionUpdate_BuildContinueTokens(state *SessionUpdateState, routeNumRelays int32, routeRelays []int32) {

	numTokens := routeNumRelays + 2

	var routeSecretKeys [constants.NextMaxNodes][]byte

	// client node

	routeSecretKeys[0], _ = crypto.SecretKey_GenerateRemote(state.RelayBackendPublicKey, state.RelayBackendPrivateKey, state.Request.ClientRoutePublicKey[:])

	// relay nodes

	relaySecretKeys := routeSecretKeys[1 : numTokens-1]

	numRouteRelays := len(routeRelays)

	for i := 0; i < numRouteRelays; i++ {
		relayIndex := routeRelays[i]
		relay := &state.Database.Relays[relayIndex]
		relaySecretKeys[i] = state.Database.RelaySecretKeys[relay.Id]
	}

	// server node

	routeSecretKeys[numTokens-1], _ = crypto.SecretKey_GenerateRemote(state.RelayBackendPublicKey, state.RelayBackendPrivateKey, state.Request.ServerRoutePublicKey[:])

	// build the tokens

	tokenData := make([]byte, numTokens*packets.SDK_EncryptedContinueRouteTokenSize)

	sessionId := state.Output.SessionId
	sessionVersion := uint8(state.Output.SessionVersion)
	expireTimestamp := state.Output.ExpireTimestamp

	core.WriteContinueTokens(tokenData, expireTimestamp, sessionId, sessionVersion, int(numTokens), routeSecretKeys[:])

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
		}
	}

	/*
		Multipath means to send packets across both the direct and the network
		next route at the same time, which reduces packet loss.
	*/

	state.Response.Multipath = true

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
		Logic for sending client relays to portal.

		This is somewhat complicated because we only want to send client relays when they change (once every 5-10 minutes)
		and at scale we usually only send accelerated sessions to the portal, thus if we are going direct, and we see new
		client relay data, we have to remember to send the client relays to the portal later on, when the session is accelerated.
	*/

	if state.Request.ClientRelayPingsHaveChanged {
		state.Output.ShouldSendClientRelaysToPortal = true
		state.Input.SentClientRelaysToPortal = false
		state.Output.SentClientRelaysToPortal = false
	}

	if !state.Input.SentClientRelaysToPortal && state.Output.ShouldSendClientRelaysToPortal && (!state.PortalNextSessionsOnly || state.Output.RouteState.Next) {
		state.Output.SentClientRelaysToPortal = true
	}

	/*
		Logic for sending server relays to portal.

		Same complications here as per-client relays...
	*/

	if state.Request.ServerRelayPingsHaveChanged {
		state.Output.ShouldSendServerRelaysToPortal = true
		state.Input.SentServerRelaysToPortal = false
		state.Output.SentServerRelaysToPortal = false
	}

	if !state.Input.SentServerRelaysToPortal && state.Output.ShouldSendServerRelaysToPortal && (!state.PortalNextSessionsOnly || state.Output.RouteState.Next) {
		state.Output.SentServerRelaysToPortal = true
	}

	/*
		Accumulate error flags from input state, and from this session update, then write it to output.

		This lets us write error flags in the session summary only, and we capture all errors that occurred for a session.
	*/

	state.Output.Error = state.Input.Error | state.Error

	/*
		The first slice always goes direct, because we do not have client relay stats yet.
	*/

	if state.Request.SliceNumber == 0 {
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
		If we have debug, print it here.
		We no longer send it down to the SDK because of MTU.
	*/

	if state.Debug != nil {
		core.Debug("%s", *state.Debug)
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
		sendPortalClientRelayUpdateMessage(state)
		sendPortalServerRelayUpdateMessage(state)

		sendAnalyticsSessionUpdateMessage(state)
		sendAnalyticsSessionSummaryMessage(state)
		sendAnalyticsClientRelayPingMessages(state)
		sendAnalyticsServerRelayPingMessages(state)
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

	message.ClientAddress = core.AnonymizeAddress(state.Request.ClientAddress)
	message.ServerAddress = state.Request.ServerAddress
	message.ServerId = common.HashString(state.Request.ServerAddress.String())
	message.MatchId = state.Request.MatchId

	message.SDKVersion_Major = byte(state.Request.Version.Major)
	message.SDKVersion_Minor = byte(state.Request.Version.Minor)
	message.SDKVersion_Patch = byte(state.Request.Version.Patch)

	message.SessionId = state.Request.SessionId
	message.UserHash = state.Request.UserHash
	message.StartTime = state.Input.StartTimestamp
	message.BuyerId = state.Request.BuyerId
	message.DatacenterId = state.Request.DatacenterId
	message.Latitude = state.Request.Latitude
	message.Longitude = state.Request.Longitude
	message.SliceNumber = state.Input.SliceNumber - 1 // IMPORTANT: Line it up with data coming from the SDK
	message.SessionEvents = state.Request.SessionEvents
	message.InternalEvents = state.Request.InternalEvents
	message.ConnectionType = uint8(state.Request.ConnectionType)
	message.PlatformType = uint8(state.Request.PlatformType)

	message.DirectRTT = state.Request.DirectRTT
	message.DirectJitter = state.Request.DirectJitter
	message.DirectPacketLoss = state.Request.DirectPacketLoss

	message.BandwidthKbpsUp = state.Request.BandwidthKbpsUp
	message.BandwidthKbpsDown = state.Request.BandwidthKbpsDown

	message.Next = state.Request.Next

	if message.Next {
		message.NextRTT = state.Request.NextRTT
		message.NextJitter = state.Request.NextJitter
		message.NextPacketLoss = state.Request.NextPacketLoss
		message.NextPredictedRTT = uint32(state.Input.RouteCost)
		message.NextNumRouteRelays = uint32(state.Input.RouteNumRelays)
		for i := 0; i < int(message.NextNumRouteRelays); i++ {
			message.NextRouteRelayId[i] = state.Input.RouteRelayIds[i]
		}
	}

	message.RealJitter = state.RealJitter
	message.RealPacketLoss = state.RealPacketLoss
	message.RealOutOfOrder = state.RealOutOfOrder

	message.DeltaTimeMin = state.Request.DeltaTimeMin
	message.DeltaTimeMax = state.Request.DeltaTimeMax
	message.DeltaTimeAvg = state.Request.DeltaTimeAvg

	message.GameRTT = state.Request.GameRTT
	message.GameJitter = state.Request.GameJitter
	message.GamePacketLoss = state.Request.GamePacketLoss

	message.NumClientRelays = uint32(state.Request.NumClientRelays)
	for i := 0; i < int(message.NumClientRelays); i++ {
		message.ClientRelayId[i] = state.Request.ClientRelayIds[i]
		message.ClientRelayRTT[i] = byte(state.Request.ClientRelayRTT[i])
		message.ClientRelayJitter[i] = byte(state.Request.ClientRelayJitter[i])
		message.ClientRelayPacketLoss[i] = state.Request.ClientRelayPacketLoss[i]
		message.ClientRelayRoutable[i] = state.Request.ClientRelayRTT[i] != 255
	}

	message.NumServerRelays = uint32(state.Request.NumServerRelays)
	for i := 0; i < int(message.NumServerRelays); i++ {
		message.ServerRelayId[i] = state.Request.ServerRelayIds[i]
		message.ServerRelayRTT[i] = byte(state.Request.ServerRelayRTT[i])
		message.ServerRelayJitter[i] = byte(state.Request.ServerRelayJitter[i])
		message.ServerRelayPacketLoss[i] = state.Request.ServerRelayPacketLoss[i]
		message.ServerRelayRoutable[i] = state.Request.ServerRelayRTT[i] != 255
	}

	message.BestScore = state.Output.BestScore
	message.BestDirectRTT = state.Output.BestDirectRTT
	message.BestNextRTT = state.Output.BestNextRTT

	message.Retry = state.Request.RetryNumber != 0
	message.FallbackToDirect = state.Request.FallbackToDirect
	message.SendToPortal = !state.PortalNextSessionsOnly || (state.PortalNextSessionsOnly && state.Output.DurationOnNext > 0)

	if state.PortalSessionUpdateMessageChannel != nil {
		state.PortalSessionUpdateMessageChannel <- &message
		state.SentPortalSessionUpdateMessage = true
	}
}

func sendPortalClientRelayUpdateMessage(state *SessionUpdateState) {

	if !(state.Input.SentClientRelaysToPortal == false && state.Output.SentClientRelaysToPortal == true) {
		return
	}

	message := messages.PortalClientRelayUpdateMessage{}

	message.Timestamp = state.StartTimestamp
	message.BuyerId = state.Request.BuyerId
	message.SessionId = state.Output.SessionId
	message.NumClientRelays = uint32(state.Request.NumClientRelays)
	for i := 0; i < int(state.Request.NumClientRelays); i++ {
		message.ClientRelayId[i] = state.Request.ClientRelayIds[i]
		message.ClientRelayRTT[i] = byte(state.Request.ClientRelayRTT[i])
		message.ClientRelayJitter[i] = byte(state.Request.ClientRelayJitter[i])
		message.ClientRelayPacketLoss[i] = state.Request.ClientRelayPacketLoss[i]
	}

	if state.PortalClientRelayUpdateMessageChannel != nil {
		state.PortalClientRelayUpdateMessageChannel <- &message
		state.SentPortalClientRelayUpdateMessage = true
	}
}

func sendPortalServerRelayUpdateMessage(state *SessionUpdateState) {

	if !(state.Input.SentServerRelaysToPortal == false && state.Output.SentServerRelaysToPortal == true) {
		return
	}

	message := messages.PortalServerRelayUpdateMessage{}

	message.Timestamp = state.StartTimestamp
	message.BuyerId = state.Request.BuyerId
	message.SessionId = state.Output.SessionId
	message.NumServerRelays = uint32(state.Request.NumServerRelays)
	for i := 0; i < int(state.Request.NumServerRelays); i++ {
		message.ServerRelayId[i] = state.Request.ServerRelayIds[i]
		message.ServerRelayRTT[i] = byte(state.Request.ServerRelayRTT[i])
		message.ServerRelayJitter[i] = byte(state.Request.ServerRelayJitter[i])
		message.ServerRelayPacketLoss[i] = state.Request.ServerRelayPacketLoss[i]
	}

	if state.PortalServerRelayUpdateMessageChannel != nil {
		state.PortalServerRelayUpdateMessageChannel <- &message
		state.SentPortalServerRelayUpdateMessage = true
	}
}

// ---------------------------------------------------------------------------------

func sendAnalyticsClientRelayPingMessages(state *SessionUpdateState) {

	if !state.Request.ClientRelayPingsHaveChanged {
		return
	}

	for i := 0; i < int(state.Request.NumClientRelays); i++ {

		message := messages.AnalyticsClientRelayPingMessage{}

		message.Timestamp = int64(state.StartTimestampNano / 1000) // nano -> micro
		message.BuyerId = int64(state.Request.BuyerId)
		message.SessionId = int64(state.Output.SessionId)
		message.UserHash = int64(state.Request.UserHash)
		message.Latitude = state.Request.Latitude
		message.Longitude = state.Request.Longitude
		anonymizedClientAddress := core.AnonymizeAddress(state.Request.ClientAddress)
		message.ClientAddress = anonymizedClientAddress.String()
		message.ConnectionType = int32(state.Request.ConnectionType)
		message.PlatformType = int32(state.Request.PlatformType)
		message.ClientRelayId = int64(state.Request.ClientRelayIds[i])
		message.ClientRelayRTT = int32(state.Request.ClientRelayRTT[i])
		message.ClientRelayJitter = int32(state.Request.ClientRelayJitter[i])
		message.ClientRelayPacketLoss = state.Request.ClientRelayPacketLoss[i]

		if state.AnalyticsClientRelayPingMessageChannel != nil {
			state.AnalyticsClientRelayPingMessageChannel <- &message
			state.SentAnalyticsClientRelayPingMessage = true
		}

	}
}

func sendAnalyticsServerRelayPingMessages(state *SessionUpdateState) {

	if !state.Request.ServerRelayPingsHaveChanged {
		return
	}

	for i := 0; i < int(state.Request.NumServerRelays); i++ {

		message := messages.AnalyticsServerRelayPingMessage{}

		message.Timestamp = int64(state.StartTimestampNano / 1000) // nano -> micro
		message.BuyerId = int64(state.Request.BuyerId)
		message.ServerAddress = state.Request.ServerAddress.String()
		message.DatacenterId = int64(state.Request.DatacenterId)
		message.ServerRelayId = int64(state.Request.ServerRelayIds[i])
		message.ServerRelayRTT = int32(state.Request.ServerRelayRTT[i])
		message.ServerRelayJitter = int32(state.Request.ServerRelayJitter[i])
		message.ServerRelayPacketLoss = state.Request.ServerRelayPacketLoss[i]

		if state.AnalyticsServerRelayPingMessageChannel != nil {
			state.AnalyticsServerRelayPingMessageChannel <- &message
			state.SentAnalyticsServerRelayPingMessage = true
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
	message.ServerId = int64(common.HashString(state.Request.ServerAddress.String()))
	message.SliceNumber = int32(state.Request.SliceNumber - 1) // IMPORTANT: Line it up with data coming from the SDK
	message.RealPacketLoss = state.RealPacketLoss
	message.RealJitter = state.RealJitter
	message.RealOutOfOrder = state.RealOutOfOrder
	message.SessionEvents = int64(state.Request.SessionEvents)
	message.InternalEvents = int64(state.Request.InternalEvents)
	message.DirectRTT = state.Request.DirectRTT
	message.DirectJitter = state.Request.DirectJitter
	message.DirectPacketLoss = state.Request.DirectPacketLoss
	message.BandwidthKbpsUp = int32(state.Request.BandwidthKbpsUp)
	message.BandwidthKbpsDown = int32(state.Request.BandwidthKbpsDown)
	message.DeltaTimeMin = state.Request.DeltaTimeMin
	message.DeltaTimeMax = state.Request.DeltaTimeMax
	message.DeltaTimeAvg = state.Request.DeltaTimeAvg
	message.GameRTT = state.Request.GameRTT
	message.GameJitter = state.Request.GameJitter
	message.GamePacketLoss = state.Request.GamePacketLoss

	// next only

	message.Next = state.Request.Next
	if message.Next {
		message.NextRTT = state.Request.NextRTT
		message.NextJitter = state.Request.NextJitter
		message.NextPacketLoss = state.Request.NextPacketLoss
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
	message.Veto = state.Input.RouteState.Veto
	message.Disabled = state.Input.RouteState.Disabled
	message.NotSelected = state.Input.RouteState.NotSelected
	message.A = state.Input.RouteState.A
	message.B = state.Input.RouteState.B
	message.Flags = int64(state.Request.Flags)

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
	message.MatchId = int64(state.Request.MatchId)
	message.UserHash = int64(state.Request.UserHash)
	message.Latitude = state.Request.Latitude
	message.Longitude = state.Request.Longitude
	anonymizedClientAddress := core.AnonymizeAddress(state.Request.ClientAddress)
	message.ClientAddress = anonymizedClientAddress.String()
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
	message.DurationOnNext = int32(state.Input.DurationOnNext)           // seconds
	message.StartTimestamp = int64(state.Input.StartTimestamp * 1000000) // seconds -> microseconds
	message.Error = int64(state.Input.Error)

	// get isp and country from client IP address via ip2location db (optional)

	if state.GetISPAndCountry != nil {
		isp, country := state.GetISPAndCountry(state.Request.ClientAddress.IP)
		message.ISP = isp
		message.Country = country
	}

	// calculate best latency reduction for this session

	if message.DurationOnNext > 0 && state.Output.BestNextRTT > 0 {
		message.BestLatencyReduction = int64(state.Output.BestDirectRTT - state.Output.BestNextRTT)
	}

	// flags

	message.Reported = state.Request.Reported
	message.LatencyReduction = state.Input.RouteState.ReduceLatency
	message.PacketLossReduction = state.Input.RouteState.ReducePacketLoss
	message.ForceNext = state.Input.RouteState.ForcedNext
	message.LongSessionUpdate = state.LongSessionUpdate
	message.Veto = state.Input.RouteState.Veto
	message.Disabled = state.Input.RouteState.Disabled
	message.NotSelected = state.Input.RouteState.NotSelected
	message.A = state.Input.RouteState.A
	message.B = state.Input.RouteState.B
	message.FallbackToDirect = state.Request.FallbackToDirect
	message.LikelyVPNOrCrossRegion = state.Input.LikelyVPNOrCrossRegion
	message.NoClientRelays = state.Input.NoClientRelays
	message.NoServerRelays = state.Input.NoServerRelays
	message.AllClientRelaysAreZero = state.Input.AllClientRelaysAreZero
	message.Flags = int64(state.Request.Flags)

	// send it

	if state.AnalyticsSessionSummaryMessageChannel != nil {
		state.AnalyticsSessionSummaryMessageChannel <- &message
		state.SentAnalyticsSessionSummaryMessage = true
	}
}

// -----------------------------------------
