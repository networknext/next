package transport

import (
	"fmt"
	"io"
	"math"
	"net"
	"time"

	"github.com/networknext/backend/modules/billing"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/storage"
)

const (
	UDPIPPacketHeaderSize = 28 // IP: 20, UDP: 8
)

type UDPPacket struct {
	From net.UDPAddr
	Data []byte
}

type UDPHandlerFunc func(io.Writer, *UDPPacket)

func datacenterExists(database *routing.DatabaseBinWrapper, datacenterID uint64) bool {
	_, exists := database.DatacenterMap[datacenterID]
	return exists
}

func datacenterEnabled(database *routing.DatabaseBinWrapper, buyerID uint64, datacenterID uint64) bool {
	datacenterAliases, ok := database.DatacenterMaps[buyerID]
	if !ok {
		return false
	}
	// todo: this should be a hash look up, not a linear walk!
	for _, dcMap := range datacenterAliases {
		if datacenterID == dcMap.DatacenterID {
			return true
		}
	}
	return false
}

func getDatacenter(database *routing.DatabaseBinWrapper, datacenterID uint64) routing.Datacenter {
	value, _ := database.DatacenterMap[datacenterID]
	return value
}

func writeServerInitResponse(w io.Writer, packet *ServerInitRequestPacket, response uint32) error {
	responsePacket := ServerInitResponsePacket{
		RequestID: packet.RequestID,
		Response:  response,
	}
	responsePacketData, err := MarshalPacket(&responsePacket)
	if err != nil {
		return err
	}
	packetHeader := append([]byte{PacketTypeServerInitResponse}, make([]byte, crypto.PacketHashSize)...)
	responseData := append(packetHeader, responsePacketData...)
	if _, err := w.Write(responseData); err != nil {
		return err
	}
	return nil
}

// ----------------------------------------------------------------------------

func ServerInitHandlerFunc(getDatabase func() *routing.DatabaseBinWrapper, metrics *metrics.ServerInitMetrics) UDPHandlerFunc {

	return func(w io.Writer, incoming *UDPPacket) {

		core.Debug("-----------------------------------------")
		core.Debug("server init packet from %s", incoming.From.String())

		metrics.HandlerMetrics.Invocations.Add(1)

		timeStart := time.Now()
		defer func() {
			milliseconds := float64(time.Since(timeStart).Milliseconds())
			metrics.HandlerMetrics.Duration.Set(milliseconds)
			if milliseconds > 100 {
				metrics.HandlerMetrics.LongDuration.Add(1)
			}
			core.Debug("server init duration: %fms\n-----------------------------------------", milliseconds)
		}()

		var packet ServerInitRequestPacket
		if err := UnmarshalPacket(&packet, incoming.Data); err != nil {
			core.Debug("could not read server init packet:\n\n%v\n", err)
			metrics.ReadPacketFailure.Add(1)
			return
		}

		core.Debug("server buyer id is %x", packet.BuyerID)

		database := getDatabase()

		responseType := InitResponseOK

		defer func() {
			if err := writeServerInitResponse(w, &packet, uint32(responseType)); err != nil {
				core.Debug("failed to write server init response: %s", err)
				metrics.WriteResponseFailure.Add(1)
			}
		}()

		buyer, exists := database.BuyerMap[packet.BuyerID]
		if !exists {
			core.Debug("unknown buyer")
			metrics.BuyerNotFound.Add(1)
			responseType = InitResponseUnknownBuyer
			return
		}

		if !buyer.Live {
			core.Debug("buyer not active")
			metrics.BuyerNotActive.Add(1)
			responseType = InitResponseBuyerNotActive
			return
		}

		if !crypto.VerifyPacket(buyer.PublicKey, incoming.Data) {
			core.Debug("signature check failed")
			metrics.SignatureCheckFailed.Add(1)
			responseType = InitResponseSignatureCheckFailed
			return
		}

		if !packet.Version.AtLeast(SDKVersion{4, 0, 0}) {
			core.Debug("sdk version is too old: %s", packet.Version.String())
			metrics.SDKTooOld.Add(1)
			responseType = InitResponseOldSDKVersion
			return
		}

		/*
			IMPORTANT: When the datacenter doesn't exist, we intentionally let the server init succeed anyway
			and just log here, so we can map the datacenter name to the datacenter id, when we are tracking it down.
		*/

		if !datacenterExists(database, packet.DatacenterID) {
			core.Error("unknown datacenter %s [%016x, %s] for buyer id %016x", packet.DatacenterName, packet.DatacenterID, incoming.From.String(), packet.BuyerID)
			metrics.DatacenterNotFound.Add(1)
			return
		}

		core.Debug("server is in datacenter \"%s\" [%x]", packet.DatacenterName, packet.DatacenterID)

		core.Debug("server initialized successfully")
	}
}

// ----------------------------------------------------------------------------

func ServerUpdateHandlerFunc(getDatabase func() *routing.DatabaseBinWrapper, postSessionHandler *PostSessionHandler, metrics *metrics.ServerUpdateMetrics) UDPHandlerFunc {

	return func(w io.Writer, incoming *UDPPacket) {

		core.Debug("-----------------------------------------")
		core.Debug("server update packet from %s", incoming.From.String())

		metrics.HandlerMetrics.Invocations.Add(1)

		timeStart := time.Now()
		defer func() {
			milliseconds := float64(time.Since(timeStart).Milliseconds())
			metrics.HandlerMetrics.Duration.Set(milliseconds)
			if milliseconds > 100 {
				metrics.HandlerMetrics.LongDuration.Add(1)
			}
			core.Debug("server update duration: %fms\n-----------------------------------------", milliseconds)
		}()

		metrics.ServerUpdatePacketSize.Set(float64(len(incoming.Data)))

		var packet ServerUpdatePacket
		if err := UnmarshalPacket(&packet, incoming.Data); err != nil {
			core.Debug("could not read server update packet:\n\n%v\n", err)
			metrics.ReadPacketFailure.Add(1)
			return
		}

		core.Debug("server buyer id is %x", packet.BuyerID)

		database := getDatabase()

		buyer, exists := database.BuyerMap[packet.BuyerID]
		if !exists {
			core.Debug("unknown buyer")
			metrics.BuyerNotFound.Add(1)
			return
		}

		if !buyer.Live {
			core.Debug("buyer not active")
			metrics.BuyerNotLive.Add(1)
			return
		}

		if !crypto.VerifyPacket(buyer.PublicKey, incoming.Data) {
			core.Debug("signature check failed")
			metrics.SignatureCheckFailed.Add(1)
			return
		}

		if !packet.Version.AtLeast(SDKVersion{4, 0, 0}) && !buyer.Debug {
			core.Debug("sdk version is too old: %s", packet.Version.String())
			metrics.SDKTooOld.Add(1)
			return
		}

		// Send the number of sessions on the server to the portal cruncher
		countData := &SessionCountData{
			Version:     SessionCountDataVersion,
			ServerID:    crypto.HashID(packet.ServerAddress.String()),
			BuyerID:     buyer.ID,
			NumSessions: packet.NumSessions,
		}
		postSessionHandler.SendPortalCounts(countData)

		if !datacenterExists(database, packet.DatacenterID) {
			core.Debug("datacenter does not exist %x", packet.DatacenterID)
			metrics.DatacenterNotFound.Add(1)
			return
		}

		core.Debug("server is in datacenter %x", packet.DatacenterID)

		core.Debug("server has %d sessions", packet.NumSessions)

		core.Debug("server updated successfully")
	}
}

// ----------------------------------------------------------------------------

func CalculateNextBytesUpAndDown(kbpsUp uint64, kbpsDown uint64, sliceDuration uint64) (uint64, uint64) {
	bytesUp := (((1000 * kbpsUp) / 8) * sliceDuration)
	bytesDown := (((1000 * kbpsDown) / 8) * sliceDuration)
	return bytesUp, bytesDown
}

func CalculateTotalPriceNibblins(routeNumRelays int, relaySellers [core.MaxRelaysPerRoute]routing.Seller, envelopeBytesUp uint64, envelopeBytesDown uint64) routing.Nibblin {

	if routeNumRelays == 0 {
		return 0
	}

	envelopeUpGB := float64(envelopeBytesUp) / 1000000000.0
	envelopeDownGB := float64(envelopeBytesDown) / 1000000000.0

	sellerPriceNibblinsPerGB := routing.Nibblin(0)
	for _, seller := range relaySellers {
		sellerPriceNibblinsPerGB += seller.EgressPriceNibblinsPerGB
	}

	nextPriceNibblinsPerGB := routing.Nibblin(1e9)
	totalPriceNibblins := float64(sellerPriceNibblinsPerGB+nextPriceNibblinsPerGB) * (envelopeUpGB + envelopeDownGB)

	return routing.Nibblin(totalPriceNibblins)
}

func CalculateRouteRelaysPrice(routeNumRelays int, relaySellers [core.MaxRelaysPerRoute]routing.Seller, envelopeBytesUp uint64, envelopeBytesDown uint64) [core.MaxRelaysPerRoute]routing.Nibblin {

	relayPrices := [core.MaxRelaysPerRoute]routing.Nibblin{}

	if routeNumRelays == 0 {
		return relayPrices
	}

	envelopeUpGB := float64(envelopeBytesUp) / 1000000000.0
	envelopeDownGB := float64(envelopeBytesDown) / 1000000000.0

	for i := 0; i < len(relayPrices); i++ {
		relayPriceNibblins := float64(relaySellers[i].EgressPriceNibblinsPerGB) * (envelopeUpGB + envelopeDownGB)
		relayPrices[i] = routing.Nibblin(relayPriceNibblins)
	}

	return relayPrices
}

func BuildNextTokens(
	sessionData *SessionData,
	database *routing.DatabaseBinWrapper,
	buyer *routing.Buyer,
	packet *SessionUpdatePacket,
	routeNumRelays int32,
	routeRelays []int32,
	allRelayIDs []uint64,
	routerPrivateKey [crypto.KeySize]byte,
	response *SessionResponsePacket,
) {
	/*
		This is either the first network next route, or we have changed network next route.

		We add an extra 10 seconds to the session expire timestamp, taking it to a total of 20 seconds.

		This means that each time we get a new route, we purchase ahead an extra 10 seconds, and renew
		the route 10 seconds early from this point, avoiding race conditions at the end of the 10 seconds
		when we continue the route.

		However, this also means that each time we switch routes, we burn the tail (10 seconds),
		so we want to minimize route switching where possible, for our customer's benefit.

		We also increase the session version here. This ensures that the new route is considered
		distinct from the old route, even if there are common relays in the old and the new routes.
	*/

	sessionData.ExpireTimestamp += billing.BillingSliceSeconds
	sessionData.SessionVersion++
	sessionData.Initial = true

	/*
		Build the cryptographic tokens that describe the route.

		The first token in the array always corresponds to the client.

		The last token in the array always corresponds to the server.

		The tokens in the middle correspond to relays.

		Each token is encrypted with the private key of the router (known only to us),
		and the public key of the corresponding node (client, server or relay).

		This gives us the following properties:

			1. Nobody can generate routes except us

			2. Only the corresponding node can decrypt the token

		While we are not currently a DDoS protection solution, property #2 means that
		we could use our technology to build one, if we choose, since we can construct
		a route and the client would only know the address of the next hop, and nothing more...
	*/

	numTokens := routeNumRelays + 2 // client + relays + server -> 1 + numRelays + 1 -> numRelays + 2

	routeAddresses, routePublicKeys := GetRouteAddressesAndPublicKeys(&packet.ClientAddress, packet.ClientRoutePublicKey, &packet.ServerAddress, packet.ServerRoutePublicKey, numTokens, routeRelays, allRelayIDs, database)

	tokenData := make([]byte, numTokens*routing.EncryptedNextRouteTokenSize)
	core.WriteRouteTokens(tokenData, sessionData.ExpireTimestamp, sessionData.SessionID, uint8(sessionData.SessionVersion), uint32(buyer.RouteShader.BandwidthEnvelopeUpKbps), uint32(buyer.RouteShader.BandwidthEnvelopeDownKbps), int(numTokens), routeAddresses, routePublicKeys, routerPrivateKey)
	response.RouteType = routing.RouteTypeNew
	response.NumTokens = numTokens
	response.Tokens = tokenData
}

func BuildContinueTokens(
	sessionData *SessionData,
	database *routing.DatabaseBinWrapper,
	buyer *routing.Buyer,
	packet *SessionUpdatePacket,
	routeNumRelays int32,
	routeRelays []int32,
	allRelayIDs []uint64,
	routerPrivateKey [crypto.KeySize]byte,
	response *SessionResponsePacket,
) {

	/*
		Continue tokens are used when we hold the same route from one slice to the next.

		Continue tokens just extend the expire time for the route across each relay by 10 seconds.

		It is smaller than the full initial description of the route, and is the common case.
	*/

	numTokens := routeNumRelays + 2 // client + relays + server -> 1 + numRelays + 1 -> numRelays + 2

	_, routePublicKeys := GetRouteAddressesAndPublicKeys(&packet.ClientAddress, packet.ClientRoutePublicKey, &packet.ServerAddress, packet.ServerRoutePublicKey, numTokens, routeRelays, allRelayIDs, database)

	tokenData := make([]byte, numTokens*routing.EncryptedContinueRouteTokenSize)
	core.WriteContinueTokens(tokenData, sessionData.ExpireTimestamp, sessionData.SessionID, uint8(sessionData.SessionVersion), int(numTokens), routePublicKeys, routerPrivateKey)
	response.RouteType = routing.RouteTypeContinue
	response.NumTokens = numTokens
	response.Tokens = tokenData
}

func GetRouteAddressesAndPublicKeys(
	clientAddress *net.UDPAddr,
	clientPublicKey []byte,
	serverAddress *net.UDPAddr,
	serverPublicKey []byte,
	numTokens int32,
	routeRelays []int32,
	allRelayIDs []uint64,
	database *routing.DatabaseBinWrapper,
) ([]*net.UDPAddr, [][]byte) {

	var routeAddresses [core.NEXT_MAX_NODES]*net.UDPAddr
	var routePublicKeys [core.NEXT_MAX_NODES][]byte

	// client node

	routeAddresses[0] = clientAddress
	routePublicKeys[0] = clientPublicKey

	// relay nodes

	relayAddresses := routeAddresses[1 : numTokens-1]
	relayPublicKeys := routePublicKeys[1 : numTokens-1]

	numRouteRelays := len(routeRelays)

	for i := 0; i < numRouteRelays; i++ {

		relayIndex := routeRelays[i]

		relayID := allRelayIDs[relayIndex]

		/*
			IMPORTANT: By this point, all relays in the route have been verified to exist
			so we don't need to check that it exists in the relay map here. It *DOES*
		*/

		relay, exists := database.RelayMap[relayID]

		if !exists {
			core.Debug("relay %x doesn't exist?!\n", relayID)
		}

		/*
			If the relay has a private address defined and the previous relay in the route
			is from the same seller, prefer to send to the relay private address instead.
			These private addresses often have better performance than the public addresses,
			and in the case of google cloud, have cheaper bandwidth prices.
		*/

		relayAddresses[i] = &relay.Addr

		if i > 0 {
			prevRelayIndex := routeRelays[i-1]
			prevID := allRelayIDs[prevRelayIndex]
			prev, _ := database.RelayMap[prevID] // IMPORTANT: Relay DOES exist.
			if prev.Seller.ID == relay.Seller.ID && relay.InternalAddr.String() != ":0" {
				relayAddresses[i] = &relay.InternalAddr
			}
		}

		relayPublicKeys[i] = relay.PublicKey
	}

	// server node

	routeAddresses[numTokens-1] = serverAddress
	routePublicKeys[numTokens-1] = serverPublicKey

	return routeAddresses[:numTokens], routePublicKeys[:numTokens]
}

// ----------------------------------------------------------------------------

type SessionHandlerState struct {

	/*
		Convenience state struct for the session update handler.

		We put all the state in here so it's easy to call out to functions to do work.

		Otherwise we have to pass a million parameters into every function and it gets old fast.
	*/

	input SessionData // sent up from the SDK. previous slice.

	output SessionData // sent down to the SDK. current slice.

	writer             io.Writer
	packet             SessionUpdatePacket
	response           SessionResponsePacket
	packetData         []byte
	metrics            *metrics.SessionUpdateMetrics
	database           *routing.DatabaseBinWrapper
	routeMatrix        *routing.RouteMatrix
	datacenter         routing.Datacenter
	buyer              routing.Buyer
	debug              *string
	ipLocator          routing.IPLocator
	staleDuration      time.Duration
	routerPrivateKey   [crypto.KeySize]byte
	postSessionHandler *PostSessionHandler

	// flags
	signatureCheckFailed bool
	unknownDatacenter    bool
	datacenterNotEnabled bool
	buyerNotFound        bool
	buyerNotLive         bool
	staleRouteMatrix     bool

	// real packet loss (from actual game packets). high precision %
	realPacketLoss float32

	// real jitter (from actual game packets).
	realJitter float32

	// route diversity is the number unique near relays with viable routes
	routeDiversity int32

	// for route planning (comes from SDK and route matrix)
	numNearRelays    int
	nearRelayIndices [core.MaxNearRelays]int32
	nearRelayRTTs    [core.MaxNearRelays]int32
	nearRelayJitters [core.MaxNearRelays]int32
	numDestRelays    int32
	destRelays       []int32

	// for session post (billing, portal etc...)
	postNearRelayCount               int
	postNearRelayIDs                 [core.MaxNearRelays]uint64
	postNearRelayNames               [core.MaxNearRelays]string
	postNearRelayAddresses           [core.MaxNearRelays]net.UDPAddr
	postNearRelayRTT                 [core.MaxNearRelays]float32
	postNearRelayJitter              [core.MaxNearRelays]float32
	postNearRelayPacketLoss          [core.MaxNearRelays]float32
	postRouteRelayNames              [core.MaxRelaysPerRoute]string
	postRouteRelaySellers            [core.MaxRelaysPerRoute]routing.Seller
	postRealPacketLossClientToServer float32
	postRealPacketLossServerToClient float32

	// todo
	/*
		multipathVetoHandler storage.MultipathVetoHandler
	*/
}

func sessionPre(state *SessionHandlerState) bool {

	var exists bool
	state.buyer, exists = state.database.BuyerMap[state.packet.BuyerID]
	if !exists {
		core.Debug("buyer not found")
		state.metrics.BuyerNotFound.Add(1)
		state.buyerNotFound = true
		return true
	}

	if !state.buyer.Live {
		core.Debug("buyer not live")
		state.metrics.BuyerNotLive.Add(1)
		state.buyerNotLive = true
		return true
	}

	if !crypto.VerifyPacket(state.buyer.PublicKey, state.packetData) {
		core.Debug("signature check failed")
		state.metrics.SignatureCheckFailed.Add(1)
		state.signatureCheckFailed = true
		return true
	}

	if state.packet.ClientPingTimedOut {
		core.Debug("client ping timed out")

		if state.postSessionHandler.featureBilling2 {
			// Unmarshal the session data into the input to verify if we wrote the summary slice in sessionPost()
			err := UnmarshalSessionData(&state.input, state.packet.SessionData[:])

			if err != nil {
				core.Debug("could not read session data:\n\n%s\n", err)
				state.metrics.ReadSessionDataFailure.Add(1)
				return true
			}
		}

		state.metrics.ClientPingTimedOut.Add(1)
		return true
	}

	if !datacenterExists(state.database, state.packet.DatacenterID) {
		core.Debug("unknown datacenter")
		state.metrics.DatacenterNotFound.Add(1)
		state.unknownDatacenter = true
		return true
	}

	if !datacenterEnabled(state.database, state.packet.BuyerID, state.packet.DatacenterID) {
		core.Debug("datacenter not enabled")
		state.metrics.DatacenterNotEnabled.Add(1)
		state.datacenterNotEnabled = true
		return true
	}

	state.datacenter = getDatacenter(state.database, state.packet.DatacenterID)

	destRelayIDs := state.routeMatrix.GetDatacenterRelayIDs(state.packet.DatacenterID)
	if len(destRelayIDs) == 0 {
		core.Debug("no relays in datacenter %x", state.packet.DatacenterID)
		state.metrics.NoRelaysInDatacenter.Add(1)
		return true
	}

	if state.routeMatrix.CreatedAt+uint64(state.staleDuration.Seconds()) < uint64(time.Now().Unix()) {
		core.Debug("stale route matrix")
		state.staleRouteMatrix = true
		state.metrics.StaleRouteMatrix.Add(1)
		return true
	}

	if state.buyer.Debug {
		core.Debug("debug enabled")
		state.debug = new(string)
	}

	for i := int32(0); i < state.packet.NumTags; i++ {
		if state.packet.Tags[i] == crypto.HashID("pro") {
			core.Debug("pro mode enabled")
			state.buyer.RouteShader.ProMode = true
		}
	}

	state.output.Initial = false

	return false
}

func sessionUpdateNewSession(state *SessionHandlerState) {

	core.Debug("new session")

	var err error

	state.output.Location, err = state.ipLocator.LocateIP(state.packet.ClientAddress.IP)

	if err != nil || state.output.Location == routing.LocationNullIsland {
		core.Debug("location veto")
		state.metrics.ClientLocateFailure.Add(1)
		state.output.RouteState.LocationVeto = true
		return
	}

	state.output.Version = SessionDataVersion
	state.output.SessionID = state.packet.SessionID
	state.output.SliceNumber = state.packet.SliceNumber + 1
	state.output.ExpireTimestamp = uint64(time.Now().Unix()) + billing.BillingSliceSeconds
	state.output.RouteState.UserID = state.packet.UserHash
	state.output.RouteState.ABTest = state.buyer.RouteShader.ABTest

	state.input = state.output
}

func sessionUpdateExistingSession(state *SessionHandlerState) {

	core.Debug("existing session")

	/*
		Read in the input state from the session data

		This is the state.output from the previous slice.
	*/

	err := UnmarshalSessionData(&state.input, state.packet.SessionData[:])

	if err != nil {
		core.Debug("could not read session data:\n\n%s\n", err)
		state.metrics.ReadSessionDataFailure.Add(1)
		return
	}

	/*
		Check for some obviously divergent data between the session request packet
		and the stored session data. If there is a mismatch, just return a direct route.
	*/

	if state.input.SessionID != state.packet.SessionID {
		core.Debug("bad session id")
		state.metrics.BadSessionID.Add(1)
		return
	}

	if state.input.SliceNumber != state.packet.SliceNumber {
		core.Debug("bad slice number")
		state.metrics.BadSliceNumber.Add(1)
		return
	}

	/*
		Copy input state to output and go to next slice.

		During the rest of the session update we transform session.output in place,
		before sending it back to the SDK in the session response packet.
	*/

	state.output = state.input
	state.output.Initial = false
	state.output.SliceNumber += 1
	state.output.ExpireTimestamp += billing.BillingSliceSeconds

	/*
		Calculate real packet loss.

		This is driven from actual game packets, not ping packets.

		This value is typically much higher precision (60HZ), vs. ping packets (10HZ).
	*/

	slicePacketsSentClientToServer := state.packet.PacketsSentClientToServer - state.input.PrevPacketsSentClientToServer
	slicePacketsSentServerToClient := state.packet.PacketsSentServerToClient - state.input.PrevPacketsSentServerToClient

	slicePacketsLostClientToServer := state.packet.PacketsLostClientToServer - state.input.PrevPacketsLostClientToServer
	slicePacketsLostServerToClient := state.packet.PacketsLostServerToClient - state.input.PrevPacketsLostServerToClient

	var realPacketLossClientToServer float32
	if slicePacketsSentClientToServer != uint64(0) {
		realPacketLossClientToServer = float32(float64(slicePacketsLostClientToServer)/float64(slicePacketsSentClientToServer)) * 100.0
	}

	var realPacketLossServerToClient float32
	if slicePacketsSentServerToClient != uint64(0) {
		realPacketLossServerToClient = float32(float64(slicePacketsLostServerToClient)/float64(slicePacketsSentServerToClient)) * 100.0
	}

	state.realPacketLoss = realPacketLossClientToServer
	if realPacketLossServerToClient > realPacketLossClientToServer {
		state.realPacketLoss = realPacketLossServerToClient
	}

	state.postRealPacketLossClientToServer = realPacketLossClientToServer
	state.postRealPacketLossServerToClient = realPacketLossServerToClient

	/*
		Calculate real jitter.

		This is driven from actual game packets, not ping packets.

		Clamp jitter between client and server at 1000.

		It is meaningless beyond that...
	*/

	if state.packet.JitterClientToServer > 1000.0 {
		state.packet.JitterClientToServer = float32(1000)
	}

	if state.packet.JitterServerToClient > 1000.0 {
		state.packet.JitterServerToClient = float32(1000)
	}

	state.realJitter = state.packet.JitterClientToServer
	if state.packet.JitterServerToClient > state.packet.JitterClientToServer {
		state.realJitter = state.packet.JitterServerToClient
	}
}

func sessionHandleFallbackToDirect(state *SessionHandlerState) bool {

	/*
		Fallback to direct is a state where the SDK has met some fatal error condition.

		When this happens, the session will go direct from that point forward.

		Here we look at flags sent up from the SDK, and send them to stackdriver metrics,
		so we can diagnose what caused any fallback to directs to happen.
	*/

	if state.packet.FallbackToDirect && !state.output.FellBackToDirect {

		core.Debug("fallback to direct")

		state.output.FellBackToDirect = true

		reported := false

		if state.packet.Flags&FallbackFlagsBadRouteToken != 0 {
			state.metrics.FallbackToDirectBadRouteToken.Add(1)
			reported = true
		}

		if state.packet.Flags&FallbackFlagsNoNextRouteToContinue != 0 {
			state.metrics.FallbackToDirectNoNextRouteToContinue.Add(1)
			reported = true
		}

		if state.packet.Flags&FallbackFlagsPreviousUpdateStillPending != 0 {
			state.metrics.FallbackToDirectPreviousUpdateStillPending.Add(1)
			reported = true
		}

		if state.packet.Flags&FallbackFlagsBadContinueToken != 0 {
			state.metrics.FallbackToDirectBadContinueToken.Add(1)
			reported = true
		}

		if state.packet.Flags&FallbackFlagsRouteExpired != 0 {
			state.metrics.FallbackToDirectRouteExpired.Add(1)
			reported = true
		}

		if state.packet.Flags&FallbackFlagsRouteRequestTimedOut != 0 {
			state.metrics.FallbackToDirectRouteRequestTimedOut.Add(1)
			reported = true
		}

		if state.packet.Flags&FallbackFlagsContinueRequestTimedOut != 0 {
			state.metrics.FallbackToDirectContinueRequestTimedOut.Add(1)
			reported = true
		}

		if state.packet.Flags&FallbackFlagsClientTimedOut != 0 {
			state.metrics.FallbackToDirectClientTimedOut.Add(1)
			reported = true
		}

		if state.packet.Flags&FallbackFlagsUpgradeResponseTimedOut != 0 {
			state.metrics.FallbackToDirectUpgradeResponseTimedOut.Add(1)
			reported = true
		}

		if state.packet.Flags&FallbackFlagsRouteUpdateTimedOut != 0 {
			state.metrics.FallbackToDirectRouteUpdateTimedOut.Add(1)
			reported = true
		}

		if state.packet.Flags&FallbackFlagsDirectPongTimedOut != 0 {
			state.metrics.FallbackToDirectDirectPongTimedOut.Add(1)
			reported = true
		}

		if state.packet.Flags&FallbackFlagsNextPongTimedOut != 0 {
			state.metrics.FallbackToDirectNextPongTimedOut.Add(1)
			reported = true
		}

		if !reported {
			state.metrics.FallbackToDirectUnknownReason.Add(1)
		}

		return true
	}

	return false
}

func sessionGetNearRelays(state *SessionHandlerState) bool {

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
	*/

	directLatency := state.packet.DirectRTT

	clientLatitude := state.output.Location.Latitude
	clientLongitude := state.output.Location.Longitude

	serverLatitude := state.datacenter.Location.Latitude
	serverLongitude := state.datacenter.Location.Longitude

	state.response.NearRelayIDs, state.response.NearRelayAddresses = state.routeMatrix.GetNearRelays(directLatency, clientLatitude, clientLongitude, serverLatitude, serverLongitude, core.MaxNearRelays)
	if len(state.response.NearRelayIDs) == 0 {
		core.Debug("no near relays :(")
		state.metrics.NearRelaysLocateFailure.Add(1)
		return false
	}

	state.response.NumNearRelays = int32(len(state.response.NearRelayIDs))
	state.response.HighFrequencyPings = state.buyer.InternalConfig.HighFrequencyPings && !state.buyer.InternalConfig.LargeCustomer
	state.response.NearRelaysChanged = true

	return true
}

func sessionUpdateNearRelayStats(state *SessionHandlerState) bool {

	/*
		This function is called once every seconds for all slices
		in a session after slice 0 (first slice).

		It takes the ping statistics for each near relay, and collates them
		into a format suitable for route planning later on in the session
		update.

		It also runs various filters inside core.ReframeRelays, which look at
		the history of latency, jitter and packet loss across the entire session
		in order to exclude near relays with bad performance from being selected.
	*/

	routeShader := &state.buyer.RouteShader

	routeState := &state.output.RouteState

	directLatency := int32(math.Ceil(float64(state.packet.DirectRTT)))
	directJitter := int32(math.Ceil(float64(state.packet.DirectJitter)))
	directPacketLoss := int32(math.Floor(float64(state.packet.DirectPacketLoss) + 0.5))
	nextPacketLoss := int32(math.Floor(float64(state.packet.NextPacketLoss) + 0.5))

	destRelayIDs := state.routeMatrix.GetDatacenterRelayIDs(state.datacenter.ID)
	if len(destRelayIDs) == 0 {
		core.Debug("no relays in datacenter %x", state.datacenter.ID)
		state.metrics.NoRelaysInDatacenter.Add(1)
		return false
	}

	sliceNumber := int32(state.packet.SliceNumber)

	state.destRelays = make([]int32, len(destRelayIDs))

	/*
		If we are holding near relays, use the held near relay RTT as input
		instead of the near relay ping data sent up from the SDK.
	*/

	if state.input.HoldNearRelays {
		core.Debug("using held near relay RTTs")
		for i := range state.packet.NearRelayIDs {
			state.packet.NearRelayRTT[i] = state.input.HoldNearRelayRTT[i] // when set to 255, near relay is excluded from routing
			state.packet.NearRelayJitter[i] = 0
			state.packet.NearRelayPacketLoss[i] = 0
		}
	}

	/*
		Reframe the near relays to get them in a relay index form relative to the current route matrix.
	*/

	core.ReframeRelays(

		// input
		routeShader,
		routeState,
		state.routeMatrix.RelayIDsToIndices,
		directLatency,
		directJitter,
		directPacketLoss,
		nextPacketLoss,
		sliceNumber,
		state.packet.NearRelayIDs,
		state.packet.NearRelayRTT,
		state.packet.NearRelayJitter,
		state.packet.NearRelayPacketLoss,
		destRelayIDs,

		// output
		state.nearRelayRTTs[:],
		state.nearRelayJitters[:],
		&state.numDestRelays,
		state.destRelays,
	)

	state.numNearRelays = len(state.packet.NearRelayIDs)

	for i := range state.packet.NearRelayIDs {
		relayIndex, exists := state.routeMatrix.RelayIDsToIndices[state.packet.NearRelayIDs[i]]
		if exists {
			state.nearRelayIndices[i] = relayIndex
		} else {
			state.nearRelayIndices[i] = -1 // near relay no longer exists in route matrix
		}
	}

	sessionFilterNearRelays(state) // IMPORTANT: Reduce % of sessions that run near relay pings for large customers

	return true

}

func sessionFilterNearRelays(state *SessionHandlerState) {

	/*
		Reduce the % of sessions running near relay pings for large customers.

		We do this by only running near relay pings for the first 3 slices, and then holding
		the near relay ping results fixed for the rest of the session.
	*/

	if !state.buyer.InternalConfig.LargeCustomer {
		return
	}

	if state.packet.SliceNumber < 4 {
		return
	}

	// IMPORTANT: On slice 4, grab the *processed* near relay RTTs from ReframeRelays,
	// which are set to 255 for any near relays excluded because of high jitter or PL
	// and hold them as the near relay RTTs to use from now on.

	if state.packet.SliceNumber == 4 {
		core.Debug("holding near relays")
		state.output.HoldNearRelays = true
		for i := 0; i < len(state.packet.NearRelayIDs); i++ {
			state.output.HoldNearRelayRTT[i] = state.nearRelayRTTs[i]
		}
	}

	// tell the SDK to stop pinging near relays

	state.response.ExcludeNearRelays = true
	for i := 0; i < core.MaxNearRelays; i++ {
		state.response.NearRelayExcluded[i] = true
	}
}

func sessionMakeRouteDecision(state *SessionHandlerState) {

	// todo: why would we copy such a potentially large map here? really bad idea...
	// multipathVetoMap := multipathVetoHandler.GetMapCopy(buyer.CompanyCode)
	multipathVetoMap := map[uint64]bool{}

	/*
		If we are on on network next but don't have any relays in our route, something is WRONG.
		Veto the session and go direct.
	*/

	if state.input.RouteState.Next && state.input.RouteNumRelays == 0 {
		core.Debug("on network next, but no route relays?")
		state.output.RouteState.Next = false
		state.output.RouteState.Veto = true
		state.metrics.NextWithoutRouteRelays.Add(1)
		return
	}

	var routeChanged bool
	var routeCost int32
	var routeNumRelays int32

	routeRelays := [core.MaxRelaysPerRoute]int32{}

	sliceNumber := int32(state.packet.SliceNumber)

	if !state.input.RouteState.Next {

		// currently going direct. should we take network next?

		if core.MakeRouteDecision_TakeNetworkNext(state.routeMatrix.RouteEntries, &state.buyer.RouteShader, &state.output.RouteState, multipathVetoMap, &state.buyer.InternalConfig, int32(state.packet.DirectRTT), state.realPacketLoss, state.nearRelayIndices[:], state.nearRelayRTTs[:], state.destRelays, &routeCost, &routeNumRelays, routeRelays[:], &state.routeDiversity, state.debug, sliceNumber) {
			BuildNextTokens(&state.output, state.database, &state.buyer, &state.packet, routeNumRelays, routeRelays[:routeNumRelays], state.routeMatrix.RelayIDs, state.routerPrivateKey, &state.response)
		}

	} else {

		// currently taking network next

		if !state.packet.Next {

			// the sdk aborted this session

			core.Debug("aborted")
			state.output.RouteState.Next = false
			state.output.RouteState.Veto = true
			state.metrics.SDKAborted.Add(1)
			return
		}

		/*
			Reframe the current route in terms of relay indices in the current route matrix

			This is necessary because the set of relays in the route matrix change over time.
		*/

		if !core.ReframeRoute(&state.output.RouteState, state.routeMatrix.RelayIDsToIndices, state.output.RouteRelayIDs[:state.output.RouteNumRelays], &routeRelays) {
			routeRelays = [core.MaxRelaysPerRoute]int32{}
			core.Debug("one or more relays in the route no longer exist")
			state.metrics.RouteDoesNotExist.Add(1)
		}

		stayOnNext, routeChanged := core.MakeRouteDecision_StayOnNetworkNext(state.routeMatrix.RouteEntries, state.routeMatrix.RelayNames, &state.buyer.RouteShader, &state.output.RouteState, &state.buyer.InternalConfig, int32(state.packet.DirectRTT), int32(state.packet.NextRTT), state.output.RouteCost, state.realPacketLoss, state.packet.NextPacketLoss, state.output.RouteNumRelays, routeRelays, state.nearRelayIndices[:], state.nearRelayRTTs[:], state.destRelays[:], &routeCost, &routeNumRelays, routeRelays[:], state.debug)

		if stayOnNext {

			// stay on network next

			if routeChanged {
				core.Debug("route changed")
				state.metrics.RouteSwitched.Add(1)
				BuildNextTokens(&state.output, state.database, &state.buyer, &state.packet, routeNumRelays, routeRelays[:routeNumRelays], state.routeMatrix.RelayIDs, state.routerPrivateKey, &state.response)
			} else {
				core.Debug("route continued")
				BuildContinueTokens(&state.output, state.database, &state.buyer, &state.packet, routeNumRelays, routeRelays[:routeNumRelays], state.routeMatrix.RelayIDs, state.routerPrivateKey, &state.response)
			}

		} else {

			// leave network next

			if state.output.RouteState.NoRoute {
				core.Debug("route no longer exists")
				state.metrics.NoRoute.Add(1)
			}

			if state.output.RouteState.MultipathOverload {
				core.Debug("multipath overload")
				state.metrics.MultipathOverload.Add(1)
			}

			if state.output.RouteState.Mispredict {
				core.Debug("mispredict")
				state.metrics.MispredictVeto.Add(1)
			}

			if state.output.RouteState.LatencyWorse {
				core.Debug("latency worse")
				state.metrics.LatencyWorse.Add(1)
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

	state.response.Committed = state.output.RouteState.Committed
	state.response.Multipath = state.output.RouteState.Multipath

	/*
		Stick the route cost, whether the route changed, and the route relay data
		in the output state. This output state is serialized into the route state
		in the route response, and sent back up to us, allowing us to know the
		current network next route, when we plan the next 10 second slice.
	*/

	if routeCost > routing.InvalidRouteValue {
		routeCost = routing.InvalidRouteValue
	}

	state.output.RouteCost = routeCost
	state.output.RouteChanged = routeChanged
	state.output.RouteNumRelays = routeNumRelays

	for i := int32(0); i < routeNumRelays; i++ {
		relayID := state.routeMatrix.RelayIDs[routeRelays[i]]
		state.output.RouteRelayIDs[i] = relayID
	}
}

func sessionPost(state *SessionHandlerState) {

	/*
		If the buyer doesn't exist, or the signature check failed,
		this is potentially a malicious request. Don't respond to it.
	*/

	if state.buyerNotFound || state.signatureCheckFailed {
		core.Debug("not responding")
		return
	}

	/*
		Build the set of near relays for the SDK to ping.

		The SDK pings these near relays and reports up the results in the next session update.

		We hold the set of near relays fixed for the session, so we only do this work on the first slice.
	*/

	if state.packet.SliceNumber == 0 {
		sessionGetNearRelays(state)
		core.Debug("first slice always goes direct")
	}

	/*
		Since post runs at the end of every session handler, run logic
		here that must run if we are taking network next vs. direct
	*/

	if state.response.RouteType != routing.RouteTypeDirect {
		core.Debug("session takes network next")
		state.metrics.NextSlices.Add(1)
		state.output.EverOnNext = true
	} else {
		core.Debug("session goes direct")
		state.metrics.DirectSlices.Add(1)
	}

	/*
		Store the packets sent and packets lost counters in the route state,
		so we can use them to calculate real packet loss next session update.
	*/

	state.output.PrevPacketsSentClientToServer = state.packet.PacketsSentClientToServer
	state.output.PrevPacketsSentServerToClient = state.packet.PacketsSentServerToClient
	state.output.PrevPacketsLostClientToServer = state.packet.PacketsLostClientToServer
	state.output.PrevPacketsLostServerToClient = state.packet.PacketsLostServerToClient

	/*
		If the core routing logic generated a debug string, include it in the response.
	*/

	if state.debug != nil {
		state.response.Debug = *state.debug
		if state.response.Debug != "" {
			state.response.HasDebug = true
		}
	}

	/*
		Determine if we should write the summary slice. Should only happen
		when the session is finished.

		The end of a session occurs when the client ping times out.

		We always set the output flag to true so that it remains recorded as true on
		subsequent slices where the client ping has timed out. Instead, we check
		the input when deciding to write billing entry 2.
	*/

	if state.postSessionHandler.featureBilling2 && state.packet.ClientPingTimedOut {
		state.output.WroteSummary = true
	}

	/*
		Write the session response packet and send it back to the caller.
	*/

	if err := writeSessionResponse(state.writer, &state.response, &state.output, state.metrics); err != nil {
		core.Debug("failed to write session update response: %s", err)
		state.metrics.WriteResponseFailure.Add(1)
		return
	}

	/*
		Check if we should multipath veto this user.

		Multipath veto detects users who spike up RTT while on multipath, indicating
		that multipath is sending too much bandwidth for their connection.

		Multipath veto users immediately leave network next (go direct), and are
		disallowed from taking multipath for future next routes for some period
		of time.

		After this time elapses, they are allowed to try multipath again.
	*/

	// todo: bring back multipath veto, but fix the weird copy the entire multipath database thing first =p
	/*
		if packet.Next && sessionData.RouteState.MultipathOverload {
			if err := multipathVetoHandler.MultipathVetoUser(buyer.CompanyCode, packet.UserHash); err != nil {
				level.Error(postSessionHandler.logger).Log("err", err)
			}
		}
	*/

	/*
		Build route relay data (for portal, billing etc...)
	*/

	buildPostRouteRelayData(state)

	/*
		Build post near relay data (for portal, billing etc...)
	*/

	buildPostNearRelayData(state)

	/*
		Build billing 2 data and send it to the billing system via pubsub (non-realtime path)

		Check the input state to see if we wrote the summary slice since
		the output state is not set to input state if we early out in sessionPre()
		when the client ping times out.

		Doing this ensures that we only write the summary slice once since the first time the
		client ping times out, input flag will be false and the output flag will be true,
		and on the following slices, both will be true.
	*/

	if state.postSessionHandler.featureBilling2 && !state.input.WroteSummary {
		billingEntry2 := buildBillingEntry2(state)

		state.postSessionHandler.SendBillingEntry2(billingEntry2)
	}

	/*
		The client times out at the end of each session, and holds on for 60 seconds.
		These slices at the end have no useful information for the portal, so we drop
		them here. Billing2 needs to know if the client times out to write the summary
		portion of the billing entry 2, but Billing1 does not.
	*/

	if state.packet.ClientPingTimedOut {
		return
	}

	/*
		Build billing data and send it to the billing system via pubsub (non-realtime path)
	*/

	if state.postSessionHandler.featureBilling {
		billingEntry := buildBillingEntry(state)

		state.postSessionHandler.SendBillingEntry(billingEntry)

		/*
			Send the billing entry to the vanity metrics system (real-time path)

			TODO: once buildBillingEntry() is deprecated, modify vanity metrics to use BillingEntry2
		*/

		if state.postSessionHandler.useVanityMetrics {
			state.postSessionHandler.SendVanityMetric(billingEntry)
		}
	}

	/*
		Send data to the portal (real-time path)
	*/

	portalData := buildPortalData(state)

	if portalData.Meta.NextRTT != 0 || portalData.Meta.DirectRTT != 0 {
		state.postSessionHandler.SendPortalData(portalData)
	}
}

func buildPostRouteRelayData(state *SessionHandlerState) {

	/*
		Build information about the relays involved in the current route.

		This data is sent to the portal, billing and the vanity metrics system.
	*/

	for i := int32(0); i < state.input.RouteNumRelays; i++ {
		relay, ok := state.database.RelayMap[state.input.RouteRelayIDs[i]]
		if ok {
			state.postRouteRelayNames[i] = relay.Name
			state.postRouteRelaySellers[i] = relay.Seller
		}
	}
}

func buildPostNearRelayData(state *SessionHandlerState) {

	state.postNearRelayCount = int(state.packet.NumNearRelays)

	for i := 0; i < state.postNearRelayCount; i++ {

		/*
			The set of near relays is held fixed at the start of a session.
			Therefore it is possible that a near relay may no longer exist.
		*/

		relayID := state.packet.NearRelayIDs[i]
		relayIndex, ok := state.routeMatrix.RelayIDsToIndices[relayID]
		if !ok {
			continue
		}

		/*
			Fill in information for near relays needed by billing and the portal.

			We grab this data from the session update packet, which corresponds to the previous slice (input).

			This makes sure all values for a slice in billing and the portal line up temporally.
		*/

		state.postNearRelayIDs[i] = relayID
		state.postNearRelayNames[i] = state.routeMatrix.RelayNames[relayIndex]
		state.postNearRelayAddresses[i] = state.routeMatrix.RelayAddresses[relayIndex]
		state.postNearRelayRTT[i] = float32(state.packet.NearRelayRTT[i])
		state.postNearRelayJitter[i] = float32(state.packet.NearRelayJitter[i])
		state.postNearRelayPacketLoss[i] = float32(state.packet.NearRelayPacketLoss[i])
	}
}

func buildBillingEntry(state *SessionHandlerState) *billing.BillingEntry {

	/*
		Each slice is 10 seconds long except for the first slice with a given network next route,
		which is 20 seconds long. Each time we change network next route, we burn the 10 second tail
		that we pre-bought at the start of the previous route.
	*/

	sliceDuration := uint64(billing.BillingSliceSeconds)
	if state.input.Initial {
		sliceDuration *= 2
	}

	/*
		Calculate the actual amounts of bytes sent up and down along the network next route
		for the duration of the previous slice (just being reported up from the SDK).

		This is *not* what we bill on.
	*/

	nextBytesUp, nextBytesDown := CalculateNextBytesUpAndDown(uint64(state.packet.NextKbpsUp), uint64(state.packet.NextKbpsDown), sliceDuration)

	/*
		Calculate the envelope bandwidth in bytes up and down for the duration of the previous slice.

		This is what we bill on.
	*/

	nextEnvelopeBytesUp, nextEnvelopeBytesDown := CalculateNextBytesUpAndDown(uint64(state.buyer.RouteShader.BandwidthEnvelopeUpKbps), uint64(state.buyer.RouteShader.BandwidthEnvelopeDownKbps), sliceDuration)

	/*
		Calculate the total price for this slice of bandwidth envelope.

		This is the sum of all relay hop prices, plus our rake, multiplied by the envelope up/down
		and the length of the session in seconds.
	*/

	totalPrice := CalculateTotalPriceNibblins(int(state.input.RouteNumRelays), state.postRouteRelaySellers, nextEnvelopeBytesUp, nextEnvelopeBytesDown)

	/*
		Calculate the per-relay hop price that sums up to the total price, minus our rake.
	*/

	routeRelayPrices := CalculateRouteRelaysPrice(int(state.input.RouteNumRelays), state.postRouteRelaySellers, nextEnvelopeBytesUp, nextEnvelopeBytesDown)

	// todo: not really sure why we transform it like this? seems wasteful
	nextRelaysPrice := [core.MaxRelaysPerRoute]uint64{}
	for i := 0; i < core.MaxRelaysPerRoute; i++ {
		nextRelaysPrice[i] = uint64(routeRelayPrices[i])
	}

	// todo: not really sure why we need to do this...
	var routeCost int32 = state.input.RouteCost
	if state.input.RouteCost == math.MaxInt32 {
		routeCost = 0
	}

	/*
		Save the first hop RTT from the client to the first relay in the route.

		This is useful for analysis and saves data science some work.
	*/

	var nearRelayRTT float32
	if state.input.RouteNumRelays > 0 {
		for i, nearRelayID := range state.postNearRelayIDs {
			if nearRelayID == state.input.RouteRelayIDs[0] {
				nearRelayRTT = float32(state.postNearRelayRTT[i])
				break
			}
		}
	}

	/*
		If the debug string is set to something by the core routing system, put it in the billing entry.
	*/

	debugString := ""
	if state.debug != nil {
		debugString = *state.debug
	}

	/*
		Clamp jitter between client and server at 1000.

		It is meaningless beyond that...
	*/

	if state.packet.JitterClientToServer > 1000.0 {
		state.packet.JitterClientToServer = float32(1000)
	}

	if state.packet.JitterServerToClient > 1000.0 {
		state.packet.JitterServerToClient = float32(1000)
	}

	/*
		Create the billing entry and return it to the caller
	*/

	billingEntry := billing.BillingEntry{
		Timestamp:                       uint64(time.Now().Unix()),
		BuyerID:                         state.packet.BuyerID,
		UserHash:                        state.packet.UserHash,
		SessionID:                       state.packet.SessionID,
		SliceNumber:                     state.packet.SliceNumber,
		DirectRTT:                       state.packet.DirectRTT,
		DirectJitter:                    state.packet.DirectJitter,
		DirectPacketLoss:                state.packet.DirectPacketLoss,
		Next:                            state.packet.Next,
		NextRTT:                         state.packet.NextRTT,
		NextJitter:                      state.packet.NextJitter,
		NextPacketLoss:                  state.packet.NextPacketLoss,
		NumNextRelays:                   uint8(state.input.RouteNumRelays),
		NextRelays:                      state.input.RouteRelayIDs,
		TotalPrice:                      uint64(totalPrice),
		ClientToServerPacketsLost:       state.packet.PacketsLostClientToServer,
		ServerToClientPacketsLost:       state.packet.PacketsLostServerToClient,
		Committed:                       state.packet.Committed,
		Flagged:                         state.packet.Reported,
		Multipath:                       state.input.RouteState.Multipath,
		Initial:                         state.input.Initial,
		NextBytesUp:                     nextBytesUp,
		NextBytesDown:                   nextBytesDown,
		EnvelopeBytesUp:                 nextEnvelopeBytesUp,
		EnvelopeBytesDown:               nextEnvelopeBytesDown,
		DatacenterID:                    state.datacenter.ID,
		RTTReduction:                    state.input.RouteState.ReduceLatency,
		PacketLossReduction:             state.input.RouteState.ReducePacketLoss,
		NextRelaysPrice:                 nextRelaysPrice,
		Latitude:                        float32(state.input.Location.Latitude),
		Longitude:                       float32(state.input.Location.Longitude),
		ISP:                             state.input.Location.ISP,
		ABTest:                          state.input.RouteState.ABTest,
		RouteDecision:                   0, // deprecated
		ConnectionType:                  uint8(state.packet.ConnectionType),
		PlatformType:                    uint8(state.packet.PlatformType),
		SDKVersion:                      state.packet.Version.String(),
		PacketLoss:                      state.realPacketLoss,
		PredictedNextRTT:                float32(routeCost),
		MultipathVetoed:                 state.input.RouteState.MultipathOverload,
		UseDebug:                        state.buyer.Debug,
		Debug:                           debugString,
		FallbackToDirect:                state.packet.FallbackToDirect,
		ClientFlags:                     state.packet.Flags,
		UserFlags:                       state.packet.UserFlags,
		NearRelayRTT:                    nearRelayRTT,
		PacketsOutOfOrderClientToServer: state.packet.PacketsOutOfOrderClientToServer,
		PacketsOutOfOrderServerToClient: state.packet.PacketsOutOfOrderServerToClient,
		JitterClientToServer:            state.packet.JitterClientToServer,
		JitterServerToClient:            state.packet.JitterServerToClient,
		NumNearRelays:                   uint8(state.postNearRelayCount),
		NearRelayIDs:                    state.postNearRelayIDs,
		NearRelayRTTs:                   state.postNearRelayRTT,
		NearRelayJitters:                state.postNearRelayJitter,
		NearRelayPacketLosses:           state.postNearRelayPacketLoss,
		RelayWentAway:                   state.input.RouteState.RelayWentAway,
		RouteLost:                       state.input.RouteState.RouteLost,
		NumTags:                         uint8(state.packet.NumTags),
		Tags:                            state.packet.Tags,
		Mispredicted:                    state.input.RouteState.Mispredict,
		Vetoed:                          state.input.RouteState.Veto,
		LatencyWorse:                    state.input.RouteState.LatencyWorse,
		NoRoute:                         state.input.RouteState.NoRoute,
		NextLatencyTooHigh:              state.input.RouteState.NextLatencyTooHigh,
		RouteChanged:                    state.input.RouteChanged,
		CommitVeto:                      state.input.RouteState.CommitVeto,
		RouteDiversity:                  uint32(state.routeDiversity),
		LackOfDiversity:                 state.input.RouteState.LackOfDiversity,
		Pro:                             state.buyer.RouteShader.ProMode && !state.input.RouteState.MultipathRestricted,
		MultipathRestricted:             state.input.RouteState.MultipathRestricted,
		ClientToServerPacketsSent:       state.packet.PacketsSentClientToServer,
		ServerToClientPacketsSent:       state.packet.PacketsSentServerToClient,
		BuyerNotLive:                    state.buyerNotLive,
		UnknownDatacenter:               state.unknownDatacenter,
		DatacenterNotEnabled:            state.datacenterNotEnabled,
		StaleRouteMatrix:                state.staleRouteMatrix,
	}

	return &billingEntry
}

func buildBillingEntry2(state *SessionHandlerState) *billing.BillingEntry2 {
	/*
		Each slice is 10 seconds long except for the first slice with a given network next route,
		which is 20 seconds long. Each time we change network next route, we burn the 10 second tail
		that we pre-bought at the start of the previous route.
	*/

	sliceDuration := uint64(billing.BillingSliceSeconds)
	if state.input.Initial {
		sliceDuration *= 2
	}

	/*
		Calculate the envelope bandwidth in bytes up and down for the duration of the previous slice.

		This is what we bill on.
	*/

	nextEnvelopeBytesUp, nextEnvelopeBytesDown := CalculateNextBytesUpAndDown(uint64(state.buyer.RouteShader.BandwidthEnvelopeUpKbps), uint64(state.buyer.RouteShader.BandwidthEnvelopeDownKbps), sliceDuration)

	/*
		Calculate the total price for this slice of bandwidth envelope.

		This is the sum of all relay hop prices, plus our rake, multiplied by the envelope up/down
		and the length of the session in seconds.
	*/

	totalPrice := CalculateTotalPriceNibblins(int(state.input.RouteNumRelays), state.postRouteRelaySellers, nextEnvelopeBytesUp, nextEnvelopeBytesDown)

	/*
		Calculate the per-relay hop price that sums up to the total price, minus our rake.
	*/

	routeRelayPrices := CalculateRouteRelaysPrice(int(state.input.RouteNumRelays), state.postRouteRelaySellers, nextEnvelopeBytesUp, nextEnvelopeBytesDown)

	// todo: not really sure why we transform it like this? seems wasteful
	nextRelayPrice := [core.MaxRelaysPerRoute]uint64{}
	for i := 0; i < core.MaxRelaysPerRoute; i++ {
		nextRelayPrice[i] = uint64(routeRelayPrices[i])
	}

	// todo: not really sure why we need to do this...
	var routeCost int32 = state.input.RouteCost
	if state.input.RouteCost == math.MaxInt32 {
		routeCost = 0
	}

	/*
		Save the first hop RTT from the client to the first relay in the route.

		This is useful for analysis and saves data science some work.
	*/

	var nearRelayRTT int32
	if state.input.RouteNumRelays > 0 {
		for i, nearRelayID := range state.postNearRelayIDs {
			if nearRelayID == state.input.RouteRelayIDs[0] {
				nearRelayRTT = int32(state.postNearRelayRTT[i])
				break
			}
		}
	}

	/*
		If the debug string is set to something by the core routing system, put it in the billing entry.
	*/

	debugString := ""
	if state.debug != nil {
		debugString = *state.debug
	}

	/*
		Separate the integer and fractional portions of real packet loss to
		allow for more efficient bitpacking while maintaining precision.
	*/

	realPacketLoss, realPacketLoss_Frac := math.Modf(float64(state.realPacketLoss))
	realPacketLoss_Frac = math.Round(realPacketLoss_Frac * 255.0)

	/*
		Recast near relay RTT, Jitter, and Packet Loss to int32.

		TODO: once buildBillingEntry() is deprecated, modify buildPostNearRelayData() to use int32 instead of float32.
	*/

	var nearRelayRTTs [core.MaxNearRelays]int32
	var nearRelayJitters [core.MaxNearRelays]int32
	var nearRelayPacketLosses [core.MaxNearRelays]int32
	for i := 0; i < state.postNearRelayCount; i++ {
		nearRelayRTTs[i] = int32(state.postNearRelayRTT[i])
		nearRelayJitters[i] = int32(state.postNearRelayJitter[i])
		nearRelayPacketLosses[i] = int32(state.postNearRelayPacketLoss[i])
	}

	/*
		Create the billing entry 2 and return it to the caller.
	*/

	billingEntry2 := billing.BillingEntry2{
		Version:                         uint32(billing.BillingEntryVersion2),
		Timestamp:                       uint32(time.Now().Unix()),
		SessionID:                       state.packet.SessionID,
		SliceNumber:                     state.packet.SliceNumber,
		DirectRTT:                       int32(state.packet.DirectRTT),
		DirectJitter:                    int32(state.packet.DirectJitter),
		DirectPacketLoss:                int32(state.packet.DirectPacketLoss),
		RealPacketLoss:                  int32(realPacketLoss),
		RealPacketLoss_Frac:             uint32(realPacketLoss_Frac),
		RealJitter:                      uint32(state.realJitter),
		Next:                            state.packet.Next,
		Flagged:                         state.packet.Reported,
		Summary:                         state.output.WroteSummary,
		UseDebug:                        state.buyer.Debug,
		Debug:                           debugString,
		DatacenterID:                    state.datacenter.ID,
		BuyerID:                         state.packet.BuyerID,
		UserHash:                        state.packet.UserHash,
		EnvelopeBytesDown:               nextEnvelopeBytesDown,
		EnvelopeBytesUp:                 nextEnvelopeBytesUp,
		Latitude:                        float32(state.input.Location.Latitude),
		Longitude:                       float32(state.input.Location.Longitude),
		ISP:                             state.input.Location.ISP,
		ConnectionType:                  int32(state.packet.ConnectionType),
		PlatformType:                    int32(state.packet.PlatformType),
		SDKVersion:                      state.packet.Version.String(),
		NumTags:                         int32(state.packet.NumTags),
		Tags:                            state.packet.Tags,
		ABTest:                          state.input.RouteState.ABTest,
		Pro:                             state.buyer.RouteShader.ProMode && !state.input.RouteState.MultipathRestricted,
		ClientToServerPacketsSent:       state.packet.PacketsSentClientToServer,
		ServerToClientPacketsSent:       state.packet.PacketsSentServerToClient,
		ClientToServerPacketsLost:       state.packet.PacketsLostClientToServer,
		ServerToClientPacketsLost:       state.packet.PacketsLostServerToClient,
		ClientToServerPacketsOutOfOrder: state.packet.PacketsOutOfOrderClientToServer,
		ServerToClientPacketsOutOfOrder: state.packet.PacketsOutOfOrderServerToClient,
		NumNearRelays:                   int32(state.postNearRelayCount),
		NearRelayIDs:                    state.postNearRelayIDs,
		NearRelayRTTs:                   nearRelayRTTs,
		NearRelayJitters:                nearRelayJitters,
		NearRelayPacketLosses:           nearRelayPacketLosses,
		NextRTT:                         int32(state.packet.NextRTT),
		NextJitter:                      int32(state.packet.NextJitter),
		NextPacketLoss:                  int32(state.packet.NextPacketLoss),
		PredictedNextRTT:                routeCost,
		NearRelayRTT:                    nearRelayRTT,
		NumNextRelays:                   int32(state.input.RouteNumRelays),
		NextRelays:                      state.input.RouteRelayIDs,
		NextRelayPrice:                  nextRelayPrice,
		TotalPrice:                      uint64(totalPrice),
		RouteDiversity:                  int32(state.routeDiversity),
		Uncommitted:                     !state.packet.Committed,
		Multipath:                       state.input.RouteState.Multipath,
		RTTReduction:                    state.input.RouteState.ReduceLatency,
		PacketLossReduction:             state.input.RouteState.ReducePacketLoss,
		RouteChanged:                    state.input.RouteChanged,
		FallbackToDirect:                state.packet.FallbackToDirect,
		MultipathVetoed:                 state.input.RouteState.MultipathOverload,
		Mispredicted:                    state.input.RouteState.Mispredict,
		Vetoed:                          state.input.RouteState.Veto,
		LatencyWorse:                    state.input.RouteState.LatencyWorse,
		NoRoute:                         state.input.RouteState.NoRoute,
		NextLatencyTooHigh:              state.input.RouteState.NextLatencyTooHigh,
		CommitVeto:                      state.input.RouteState.CommitVeto,
		UnknownDatacenter:               state.unknownDatacenter,
		DatacenterNotEnabled:            state.datacenterNotEnabled,
		BuyerNotLive:                    state.buyerNotLive,
		StaleRouteMatrix:                state.staleRouteMatrix,
	}

	return &billingEntry2
}

func buildPortalData(state *SessionHandlerState) *SessionPortalData {

	/*
		Build the relay hops for the portal
	*/

	// todo: we should try to avoid allocations
	hops := make([]RelayHop, state.input.RouteNumRelays)
	for i := int32(0); i < state.input.RouteNumRelays; i++ {
		hops[i] = RelayHop{
			Version: RelayHopVersion,
			ID:      state.input.RouteRelayIDs[i],
			Name:    state.postRouteRelayNames[i],
		}
	}

	/*
		Build the near relay data for the portal
	*/

	// todo: we should try to avoid allocations
	nearRelayPortalData := make([]NearRelayPortalData, state.postNearRelayCount)
	for i := range nearRelayPortalData {
		nearRelayPortalData[i] = NearRelayPortalData{
			Version: NearRelayPortalDataVersion,
			ID:      state.postNearRelayIDs[i],
			Name:    state.postNearRelayNames[i],
			ClientStats: routing.Stats{
				RTT:        float64(state.postNearRelayRTT[i]),
				Jitter:     float64(state.postNearRelayJitter[i]),
				PacketLoss: float64(state.postNearRelayPacketLoss[i]),
			},
		}
	}

	/*
		Calculate the delta between network next and direct.

		Clamp the delta RTT above 0. This is used for the top sessions page.
	*/

	var deltaRTT float32
	if state.packet.Next && state.packet.NextRTT != 0 && state.packet.DirectRTT >= state.packet.NextRTT {
		deltaRTT = state.packet.DirectRTT - state.packet.NextRTT
	}

	/*
		Predicted RTT is the round trip time that we predict, even if we don't
		take network next. It's a conservative prodiction.
	*/

	predictedRTT := float64(state.input.RouteCost)
	if state.input.RouteCost >= routing.InvalidRouteValue {
		predictedRTT = 0
	}

	/*
		Build the portal data and return it to the caller.
	*/

	portalData := SessionPortalData{
		Version: SessionPortalDataVersion,
		Meta: SessionMeta{
			Version:         SessionMetaVersion,
			ID:              state.packet.SessionID,
			UserHash:        state.packet.UserHash,
			DatacenterName:  state.datacenter.Name,
			DatacenterAlias: state.datacenter.AliasName,
			OnNetworkNext:   state.packet.Next,
			NextRTT:         float64(state.packet.NextRTT),
			DirectRTT:       float64(state.packet.DirectRTT),
			DeltaRTT:        float64(deltaRTT),
			Location:        state.input.Location,
			ClientAddr:      state.packet.ClientAddress.String(),
			ServerAddr:      state.packet.ServerAddress.String(),
			Hops:            hops,
			SDK:             state.packet.Version.String(),
			Connection:      uint8(state.packet.ConnectionType),
			NearbyRelays:    nearRelayPortalData,
			Platform:        uint8(state.packet.PlatformType),
			BuyerID:         state.packet.BuyerID,
		},
		Slice: SessionSlice{
			Version:   SessionSliceVersion,
			Timestamp: time.Now(),
			Next: routing.Stats{
				RTT:        float64(state.packet.NextRTT),
				Jitter:     float64(state.packet.NextJitter),
				PacketLoss: float64(state.packet.NextPacketLoss),
			},
			Direct: routing.Stats{
				RTT:        float64(state.packet.DirectRTT),
				Jitter:     float64(state.packet.DirectJitter),
				PacketLoss: float64(state.packet.DirectPacketLoss),
			},
			Predicted: routing.Stats{
				RTT: predictedRTT,
			},
			ClientToServerStats: routing.Stats{
				Jitter:     float64(state.packet.JitterClientToServer),
				PacketLoss: float64(state.postRealPacketLossClientToServer),
			},
			ServerToClientStats: routing.Stats{
				Jitter:     float64(state.packet.JitterServerToClient),
				PacketLoss: float64(state.postRealPacketLossServerToClient),
			},
			RouteDiversity: uint32(state.routeDiversity),
			Envelope: routing.Envelope{
				Up:   int64(state.packet.NextKbpsUp),
				Down: int64(state.packet.NextKbpsDown),
			},
			IsMultiPath:       state.input.RouteState.Multipath,
			IsTryBeforeYouBuy: !state.input.RouteState.Committed,
			OnNetworkNext:     state.packet.Next,
		},
		Point: SessionMapPoint{
			Version:   SessionMapPointVersion,
			Latitude:  float64(state.input.Location.Latitude),
			Longitude: float64(state.input.Location.Longitude),
			SessionID: state.input.SessionID,
		},
		LargeCustomer: state.buyer.InternalConfig.LargeCustomer,
		EverOnNext:    state.input.EverOnNext,
	}

	return &portalData
}

// ------------------------------------------------------------------

func writeSessionResponse(w io.Writer, response *SessionResponsePacket, sessionData *SessionData, metrics *metrics.SessionUpdateMetrics) error {
	sessionDataBuffer, err := MarshalSessionData(sessionData)
	if err != nil {
		return err
	}

	if len(sessionDataBuffer) > MaxSessionDataSize {
		return fmt.Errorf("session data of %d exceeds limit of %d bytes", len(sessionDataBuffer), MaxSessionDataSize)
	}
	response.SessionDataBytes = int32(len(sessionDataBuffer))
	copy(response.SessionData[:], sessionDataBuffer)
	responsePacketData, err := MarshalPacket(response)
	if err != nil {
		return err
	}
	packetHeader := append([]byte{PacketTypeSessionResponse}, make([]byte, crypto.PacketHashSize)...)
	responseData := append(packetHeader, responsePacketData...)

	if sessionData.RouteState.Next {
		metrics.NextSessionResponsePacketSize.Set(float64(len(responseData) + UDPIPPacketHeaderSize))
	} else {
		metrics.DirectSessionResponsePacketSize.Set(float64(len(responseData) + UDPIPPacketHeaderSize))
	}

	if _, err := w.Write(responseData); err != nil {
		return err
	}

	return nil
}

// ------------------------------------------------------------------

func SessionUpdateHandlerFunc(
	getIPLocator func(sessionID uint64) routing.IPLocator,
	getRouteMatrix func() *routing.RouteMatrix,
	multipathVetoHandler storage.MultipathVetoHandler,
	getDatabase func() *routing.DatabaseBinWrapper,
	routerPrivateKey [crypto.KeySize]byte,
	postSessionHandler *PostSessionHandler,
	metrics *metrics.SessionUpdateMetrics,
	staleDuration time.Duration,
) UDPHandlerFunc {

	return func(w io.Writer, incoming *UDPPacket) {

		core.Debug("-----------------------------------------")
		core.Debug("session update packet from %s", incoming.From.String())

		metrics.HandlerMetrics.Invocations.Add(1)

		// make sure we track the length of session update handlers

		timeStart := time.Now()
		defer func() {
			milliseconds := float64(time.Since(timeStart).Milliseconds())
			metrics.HandlerMetrics.Duration.Set(milliseconds)
			if milliseconds > 100 {
				metrics.HandlerMetrics.LongDuration.Add(1)
			}
			core.Debug("session update duration: %fms\n-----------------------------------------", milliseconds)
		}()

		// read in the session update packet

		metrics.SessionUpdatePacketSize.Set(float64(len(incoming.Data)))

		var state SessionHandlerState

		if err := UnmarshalPacket(&state.packet, incoming.Data); err != nil {
			core.Debug("could not read session update packet:\n\n%v\n", err)
			metrics.ReadPacketFailure.Add(1)
			return
		}

		// log stuff we want to see with each session update (debug only)

		core.Debug("buyer id is %x", state.packet.BuyerID)
		core.Debug("datacenter id is %x", state.packet.DatacenterID)
		core.Debug("session id is %x", state.packet.SessionID)
		core.Debug("slice number is %d", state.packet.SliceNumber)
		core.Debug("retry number is %d", state.packet.RetryNumber)

		/*
			Build session handler state. Putting everything in a struct makes calling subroutines much easier.
		*/

		state.writer = w
		state.metrics = metrics
		state.database = getDatabase()
		state.datacenter = routing.UnknownDatacenter
		state.packetData = incoming.Data
		state.ipLocator = getIPLocator(state.packet.SessionID)
		state.routeMatrix = getRouteMatrix()
		state.staleDuration = staleDuration
		state.routerPrivateKey = routerPrivateKey
		state.response = SessionResponsePacket{
			Version:     state.packet.Version,
			SessionID:   state.packet.SessionID,
			SliceNumber: state.packet.SliceNumber,
			RouteType:   routing.RouteTypeDirect,
		}
		state.postSessionHandler = postSessionHandler

		/*
			Session post *always* runs at the end of this function

			It writes and sends the response packet back to the sender,
			and sends session data to billing, vanity metrics and the portal.
		*/

		defer sessionPost(&state)

		/*
			Call session pre function

			This function checks for early out conditions and does some setup of the handler state.

			If it returns true, one of the early out conditions has been met, so we return early.
		*/

		if sessionPre(&state) {
			return
		}

		/*
			Update the session

			Do setup on slice 0, then for subsequent slices transform state.input -> state.output

			state.output is sent down to the SDK in the session response packet, and next slice
			it is sent back up to us in the subsequent session update packet for this session.

			This is how we make this handler stateless. Without this, we need to store per-session
			data somewhere and this is extremely difficult at scale, given the real-time nature of
			this handler.
		*/

		if state.packet.SliceNumber == 0 {
			sessionUpdateNewSession(&state)
		} else {
			sessionUpdateExistingSession(&state)
		}

		/*
			Handle fallback to direct.

			Fallback to direct is a condition where the SDK indicates that it has seen
			some fatal error, like not getting a session response from the backend,
			and has decided to go direct for the rest of the session.

			When this happens, we early out to save processing time.
		*/

		if sessionHandleFallbackToDirect(&state) {
			return
		}

		/*
			Process near relay ping statistics after the first slice.

			We use near relay latency, jitter and packet loss for route planning.
		*/

		sessionUpdateNearRelayStats(&state)

		/*
			Decide whether we should take network next or not.
		*/

		sessionMakeRouteDecision(&state)

		core.Debug("session updated successfully")
	}
}

// ----------------------------------------------------------------------------
