package transport

import (
	"fmt"
	"io"
	"math"
	"net"
	"time"

	"github.com/go-kit/kit/log"

	"github.com/networknext/backend/modules/billing"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/storage"
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
	// todo: this should be reworked so that it's a map lookup to check if a datacenter is enabled
	// the linear walk below is not acceptable (this is in the hot path!!!)
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

func ServerInitHandlerFunc(logger log.Logger, getDatabase func() *routing.DatabaseBinWrapper, metrics *metrics.ServerInitMetrics) UDPHandlerFunc {

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
			fmt.Printf("error: unknown datacenter %s [%x] for buyer id %x", packet.DatacenterName, packet.DatacenterID, packet.BuyerID)
			metrics.DatacenterNotFound.Add(1)
			return
		}

		core.Debug("server is in datacenter \"%s\" [%x]", packet.DatacenterName, packet.DatacenterID)

		core.Debug("server initialized successfully")
	}
}

// ----------------------------------------------------------------------------

func ServerUpdateHandlerFunc(logger log.Logger, getDatabase func() *routing.DatabaseBinWrapper, postSessionHandler *PostSessionHandler, metrics *metrics.ServerUpdateMetrics) UDPHandlerFunc {

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
			// todo: add buyer not active metric here
			// metrics.BuyerNotLive.Add(1)
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
			ServerID:    crypto.HashID(packet.ServerAddress.String()),
			BuyerID:     buyer.ID,
			NumSessions: packet.NumSessions,
		}
		postSessionHandler.SendPortalCounts(countData)

		if !datacenterExists(database, packet.DatacenterID) {
			core.Debug("datacenter does not exist %x", packet.DatacenterID)
			// todo: add metric for this
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

		relay, _ := database.RelayMap[relayID]

		/*
			If the relay has a private address defined and the previous relay in the route
			is from the same seller, prefer to send to the relay private address instead.
			These private addresses often have better performance than the private addresses,
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
	input SessionData // sent down to the SDK

	output SessionData // sent down to the SDK

	writer        io.Writer
	packet        SessionUpdatePacket
	response      SessionResponsePacket
	packetData    []byte
	metrics       *metrics.SessionUpdateMetrics
	database      *routing.DatabaseBinWrapper
	routeMatrix   *routing.RouteMatrix
	datacenter    routing.Datacenter
	buyer         routing.Buyer
	debug         *string
	ipLocator     routing.IPLocator
	staleDuration time.Duration

	// flags
	signatureCheckFailed bool
	unknownDatacenter    bool
	datacenterNotEnabled bool
	buyerNotFound        bool
	buyerNotLive         bool
	staleRouteMatrix     bool

	// real packet loss (from actual game packets). high precision %
	realPacketLoss float32

	// for route planning (comes from SDK and route matrix)
	nearRelayRTTs    [core.MaxNearRelays]int32
	nearRelayJitters [core.MaxNearRelays]int32
	numDestRelays    int32
	destRelays       []int32

	// for display (sent to billing, portal etc...)
	outputNearRelayCount      int
	outputNearRelayIDs        [core.MaxNearRelays]uint64
	outputNearRelayNames      [core.MaxNearRelays]string
	outputNearRelayAddresses  [core.MaxNearRelays]net.UDPAddr
	outputNearRelayRTT        [core.MaxNearRelays]int32
	outputNearRelayJitter     [core.MaxNearRelays]int32
	outputNearRelayPacketLoss [core.MaxNearRelays]int32

	// todo
	/*
		routeDiversity int32
		postSessionHandler *PostSessionHandler
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
		state.metrics.ClientPingTimedOut.Add(1)
		return true
	}

	if !datacenterExists(state.database, state.packet.DatacenterID) {
		core.Debug("unknown datacenter")
		// todo: add a metric for this condition
		// state.metrics.UnknownDatacenter.Add(1)
		state.unknownDatacenter = true
		return true
	}

	if !datacenterEnabled(state.database, state.packet.BuyerID, state.packet.DatacenterID) {
		core.Debug("datacenter not enabled")
		// todo: add a metric for this condition
		// state.metrics.DatacenterNotEnabled.Add(1)
		state.datacenterNotEnabled = true
		return true
	}

	destRelayIDs := state.routeMatrix.GetDatacenterRelayIDs(state.datacenter.ID)
	if len(destRelayIDs) == 0 {
		core.Debug("no relays in datacenter %x", state.datacenter.ID)
		state.metrics.NoRelaysInDatacenter.Add(1)
		return true
	}

	if state.routeMatrix.CreatedAt+uint64(state.staleDuration.Seconds()) < uint64(time.Now().Unix()) {
		core.Debug("stale route matrix")
		state.staleRouteMatrix = true
		// todo: add a metric for this
		// state.metrics.StaleRouteMatrix.Add(1)
		return true
	}

	if state.input.RouteState.NumNearRelays != state.packet.NumNearRelays {
		core.Debug("near relay count diverged")
		// todo: add a metric for this
		// state.metrics.NearRelayCountDiverged.Add(1)
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

	state.datacenter = getDatacenter(state.database, state.packet.DatacenterID)

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

	err := UnmarshalSessionData(&state.input, state.packet.SessionData[:])

	if err != nil {
		core.Debug("could not read session data:\n\n%s\n", err)
		state.metrics.ReadSessionDataFailure.Add(1)
		return
	}

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

	state.output = state.input
	state.output.SliceNumber += 1
	state.output.ExpireTimestamp += billing.BillingSliceSeconds

	// calculate real packet loss (driven from actual game packets, not ping packets)

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
}

func sessionHandleFallbackToDirect(state *SessionHandlerState) bool {

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
	state.response.HighFrequencyPings = state.buyer.InternalConfig.HighFrequencyPings
	state.response.NearRelaysChanged = true

	return true
}

func sessionUpdateNearRelayStats(state *SessionHandlerState) bool {

	routeShader := &state.buyer.RouteShader

	routeState := &state.output.RouteState

	directLatency := int32(math.Ceil(float64(state.packet.DirectRTT)))
	directJitter := int32(math.Ceil(float64(state.packet.DirectJitter)))
	directPacketLoss := math.Floor(float64(state.packet.DirectPacketLoss) + 0.5)
	nextPacketLoss := math.Floor(float64(state.packet.NextPacketLoss) + 0.5)

	/*
		IMPORTANT: If we are not currently on network next, replace the direct packet loss
		that comes from pings (@10HZ), with the real packet loss from real game packets (60HZ)
		This gives us a much higher precision view of packet loss, which is important because
		it's used as an input to decide if we should take network next to reduce packet loss!
	*/
	if !state.packet.Next {
		realPacketLoss := float64(state.realPacketLoss)
		if realPacketLoss > directPacketLoss {
			directPacketLoss = realPacketLoss
		}
	}

	destRelayIDs := state.routeMatrix.GetDatacenterRelayIDs(state.datacenter.ID)
	if len(destRelayIDs) == 0 {
		core.Debug("no relays in datacenter %x", state.datacenter.ID)
		state.metrics.NoRelaysInDatacenter.Add(1)
		return false
	}

	sliceNumber := int32(state.packet.SliceNumber)

	state.destRelays = make([]int32, len(destRelayIDs))

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

	return true

}

func sessionMakeRouteDecision(state *SessionHandlerState) {

	var routeCost int32

	routeRelays := [core.MaxRelaysPerRoute]int32{}

	// todo: why would we copy such a potentially large map here?
	// multipathVetoMap := multipathVetoHandler.GetMapCopy(buyer.CompanyCode)

	// todo
	/*
	// todo: don't allocate in hot path
	nearRelayIndices := make([]int32, nearRelays.Count)
	nearRelayCosts := make([]int32, nearRelays.Count)
	for i := int32(0); i < nearRelays.Count; i++ {
		nearRelayIndex, ok := routeMatrix.RelayIDsToIndices[nearRelays.IDs[i]]
		if !ok {
			continue
		}

		nearRelayIndices[i] = nearRelayIndex
		nearRelayCosts[i] = nearRelays.RTTs[i]
	}
	*/

	var routeNumRelays int32

	/*
		If we are on on network next but don't have any relays in our route, something is WRONG.
		Veto the session and go direct.
	*/

	if state.input.RouteState.Next && state.input.RouteNumRelays == 0 {
		core.Debug("on network next, but no route relays?")
		state.output.RouteState.Next = false
		state.output.RouteState.Veto = true
		// todo: add metric for this condition
		// state.metrics.NextWithoutRouteRelays.Add(1)
		return
	}

	routeChanged := false

	if !state.input.RouteState.Next {

		// currently going direct. should we take network next?

		// todo
		/*
		if core.MakeRouteDecision_TakeNetworkNext(routeMatrix.RouteEntries, &buyer.RouteShader, &sessionData.RouteState, multipathVetoMap, &buyer.InternalConfig, int32(packet.DirectRTT), slicePacketLoss, nearRelayIndices, nearRelayCosts, reframedDestRelays, &routeCost, &routeNumRelays, routeRelays[:], &routeDiversity, state.debug) {
			BuildNextTokens(&sessionData, state.database, &buyer, &packet, routeNumRelays, routeRelays[:], routeMatrix.RelayIDs, routerPrivateKey, &response)
		}
		*/

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

		// todo
		/*
			if !core.ReframeRoute(&sessionData.RouteState, routeMatrix.RelayIDsToIndices, sessionData.RouteRelayIDs[:sessionData.RouteNumRelays], &routeRelays) {
				routeRelays = [core.MaxRelaysPerRoute]int32{}
				core.Debug("one or more relays in the route no longer exist")
				metrics.RouteDoesNotExist.Add(1)
			}
		*/

		// todo
		// if stayOnNext, routeChanged := core.MakeRouteDecision_StayOnNetworkNext(routeMatrix.RouteEntries, routeMatrix.RelayNames, &buyer.RouteShader, &sessionData.RouteState, &buyer.InternalConfig, int32(packet.DirectRTT), int32(packet.NextRTT), sessionData.RouteCost, slicePacketLoss, packet.NextPacketLoss, sessionData.RouteNumRelays, routeRelays, nearRelayIndices, nearRelayCosts, reframedDestRelays, &routeCost, &routeNumRelays, routeRelays[:], state.debug)

		stayOnNext := false

		if stayOnNext {

			// stay on network next

			if routeChanged {
				core.Debug("route changed")
				state.metrics.RouteSwitched.Add(1)
				// todo
				// BuildNextTokens(&sessionData, state.database, &buyer, &packet, routeNumRelays, routeRelays[:], routeMatrix.RelayIDs, routerPrivateKey, &response)
			} else {
				core.Debug("route continued")
				// todo
				// BuildContinueTokens(&sessionData, state.database, &buyer, &packet, routeNumRelays, routeRelays[:], routeMatrix.RelayIDs, routerPrivateKey, &response)
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

	state.response.Committed = state.output.RouteState.Committed
	state.response.Multipath = state.output.RouteState.Multipath

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
		so we can use them to calculate delta packets sent and delta packets lost
		for the next slice in our real packet loss calculation.
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
		Write the session response packet and send it back to the caller.
	*/

	if err := writeSessionResponse(state.writer, &state.response, &state.output); err != nil {
		core.Debug("failed to write session update response: %s", err)
		state.metrics.WriteResponseFailure.Add(1)
		return
	}

	/*
		Build route relay data.
	*/

	routeRelayNames := [core.MaxRelaysPerRoute]string{}
	routeRelaySellers := [core.MaxRelaysPerRoute]routing.Seller{}

	for i := int32(0); i < state.input.RouteNumRelays; i++ {
		relay, ok := state.database.RelayMap[state.input.RouteRelayIDs[i]]
		if ok {
			routeRelayNames[i] = relay.Name
			routeRelaySellers[i] = relay.Seller
		}
	}

	/*
		Build output near relay data (for portal, billing etc...)
	*/

	state.outputNearRelayCount = int(state.input.RouteState.NumNearRelays)

	for i := 0; i < state.outputNearRelayCount; i++ {

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
			Fill in information for near relays.

			We grab this data from the input route state, which corresponds
			to the previous slice. This makes sure all values for a slice in
			billing and the portal line up temporally.
		*/

		state.outputNearRelayIDs[i] = relayID
		state.outputNearRelayNames[i] = state.routeMatrix.RelayNames[relayIndex]
		state.outputNearRelayAddresses[i] = state.routeMatrix.RelayAddresses[relayIndex]
		state.outputNearRelayRTT[i] = state.input.RouteState.NearRelayRTT[i]
		state.outputNearRelayJitter[i] = state.input.RouteState.NearRelayJitter[i]

		/*
			We have to be a bit tricky to get packet loss for near relays,
			since we store a history of near relay PL in a sliding window,
			not just a single value.

			Take the previous entry in the sliding window and it corresponds
			to the input slice number that we want, however we can't use -1
			because modulo negative numbers doesn't do what we want, so add 7 instead...
		*/
		index := (state.input.RouteState.PLHistoryIndex + 7) % 8
		state.outputNearRelayPacketLoss[i] = int32(state.input.RouteState.NearRelayPLHistory[index])
	}

	/*
		The client times out at the end of each session, and holds on for 60 seconds.
		These slices at the end have no useful information, so we drop them here.
	*/

	if state.packet.ClientPingTimedOut {
		return
	}

	// todo: pass in routing state not individual data pieces

	/*
		go PostSessionUpdate(postSessionHandler,
			&packet,
			&prevSessionData,
			&buyer,
			multipathVetoHandler,
			routeRelayNames,
			routeRelaySellers,
			nearRelays,
			&datacenter,
			routeDiversity,
			slicePacketLossClientToServer,
			slicePacketLossServerToClient,
			debug,
			unknownDatacenter,
			datacenterNotEnabled,
			buyerNotLive,
			staleRouteMatrix,
		)
	*/
}

func buildPortalData(state *SessionHandlerState, portalData *SessionPortalData) {

	/*
		// todo: we should try to avoid allocation here
		hops := make([]RelayHop, sessionData.RouteNumRelays)
		for i := int32(0); i < sessionData.RouteNumRelays; i++ {
			hops[i] = RelayHop{
				ID:   sessionData.RouteRelayIDs[i],
				Name: routeRelayNames[i],
			}
		}

		// todo: we should try to avoid allocation here
		nearRelayPortalData := make([]NearRelayPortalData, nearRelays.Count)
		for i := range nearRelayPortalData {
			nearRelayPortalData[i] = NearRelayPortalData{
				ID:   nearRelays.IDs[i],
				Name: nearRelays.Names[i],
				ClientStats: routing.Stats{
					RTT:        float64(nearRelays.RTTs[i]),
					Jitter:     float64(nearRelays.Jitters[i]),
					PacketLoss: float64(nearRelays.PacketLosses[i]),
				},
			}
		}

		// todo: sorting below should be done by the portal instead. here we are in hot path and must do as little work as possible

		// Sort the near relays for display purposes
		sort.Slice(nearRelayPortalData, func(i, j int) bool {
			return nearRelayPortalData[i].Name < nearRelayPortalData[j].Name
		})

		var deltaRTT float32
		if packet.Next && packet.NextRTT != 0 && packet.DirectRTT >= packet.NextRTT {
			deltaRTT = packet.DirectRTT - packet.NextRTT
		}

		predictedRTT := float64(sessionData.RouteCost)
		if sessionData.RouteCost >= routing.InvalidRouteValue {
			predictedRTT = 0
		}

		*portalData = SessionPortalData{
			Meta: SessionMeta{
				ID:              packet.SessionID,
				UserHash:        packet.UserHash,
				DatacenterName:  datacenter.Name,
				DatacenterAlias: datacenter.AliasName,
				OnNetworkNext:   packet.Next,
				NextRTT:         float64(packet.NextRTT),
				DirectRTT:       float64(packet.DirectRTT),
				DeltaRTT:        float64(deltaRTT),
				Location:        sessionData.Location,
				ClientAddr:      packet.ClientAddress.String(),
				ServerAddr:      packet.ServerAddress.String(),
				Hops:            hops,
				SDK:             packet.Version.String(),
				Connection:      uint8(packet.ConnectionType),
				NearbyRelays:    nearRelayPortalData,
				Platform:        uint8(packet.PlatformType),
				BuyerID:         packet.BuyerID,
			},
			Slice: SessionSlice{
				Timestamp: time.Now(),
				Next: routing.Stats{
					RTT:        float64(packet.NextRTT),
					Jitter:     float64(packet.NextJitter),
					PacketLoss: float64(packet.NextPacketLoss),
				},
				Direct: routing.Stats{
					RTT:        float64(packet.DirectRTT),
					Jitter:     float64(packet.DirectJitter),
					PacketLoss: float64(packet.DirectPacketLoss),
				},
				Predicted: routing.Stats{
					RTT: predictedRTT,
				},
				ClientToServerStats: routing.Stats{
					Jitter:     float64(packet.JitterClientToServer),
					PacketLoss: float64(slicePacketLossClientToServer),
				},
				ServerToClientStats: routing.Stats{
					Jitter:     float64(packet.JitterServerToClient),
					PacketLoss: float64(slicePacketLossServerToClient),
				},
				RouteDiversity: uint32(routeDiversity),
				Envelope: routing.Envelope{
					Up:   int64(packet.NextKbpsUp),
					Down: int64(packet.NextKbpsDown),
				},
				IsMultiPath:       sessionData.RouteState.Multipath,
				IsTryBeforeYouBuy: !sessionData.RouteState.Committed,
				OnNetworkNext:     packet.Next,
			},
			Point: SessionMapPoint{
				Latitude:  float64(sessionData.Location.Latitude),
				Longitude: float64(sessionData.Location.Longitude),
			},
			LargeCustomer:	   buyer.InternalConfig.LargeCustomer,
			EverOnNext:    sessionData.EverOnNext,
		}
	*/
}

/*
func PostSessionUpdate(
	postSessionHandler *PostSessionHandler,
	packet *SessionUpdatePacket,
	sessionData *SessionData,
	buyer *routing.Buyer,
	multipathVetoHandler storage.MultipathVetoHandler,
	routeRelayNames [core.MaxRelaysPerRoute]string,
	routeRelaySellers [core.MaxRelaysPerRoute]routing.Seller,
	nearRelays nearRelayGroup,
	datacenter *routing.Datacenter,
	routeDiversity int32,
	slicePacketLossClientToServer float32,
	slicePacketLossServerToClient float32,
	debug *string,
	unknownDatacenter bool,
	datacenterNotEnabled bool,
	buyerNotLive bool,
	staleRouteMatrix bool,
) {
	// todo: move the function below into its own "build billing entry" function

	sliceDuration := uint64(billing.BillingSliceSeconds)
	if sessionData.Initial {
		sliceDuration *= 2
	}
	nextBytesUp, nextBytesDown := CalculateNextBytesUpAndDown(uint64(packet.NextKbpsUp), uint64(packet.NextKbpsDown), sliceDuration)
	nextEnvelopeBytesUp, nextEnvelopeBytesDown := CalculateNextBytesUpAndDown(uint64(buyer.RouteShader.BandwidthEnvelopeUpKbps), uint64(buyer.RouteShader.BandwidthEnvelopeDownKbps), sliceDuration)
	totalPrice := CalculateTotalPriceNibblins(int(sessionData.RouteNumRelays), routeRelaySellers, nextEnvelopeBytesUp, nextEnvelopeBytesDown)
	routeRelayPrices := CalculateRouteRelaysPrice(int(sessionData.RouteNumRelays), routeRelaySellers, nextEnvelopeBytesUp, nextEnvelopeBytesDown)

	// Check if we should multipath veto the user
	if packet.Next && sessionData.RouteState.MultipathOverload {
		if err := multipathVetoHandler.MultipathVetoUser(buyer.CompanyCode, packet.UserHash); err != nil {
			level.Error(postSessionHandler.logger).Log("err", err)
		}
	}

	nextRelaysPrice := [core.MaxRelaysPerRoute]uint64{}
	for i := 0; i < core.MaxRelaysPerRoute; i++ {
		nextRelaysPrice[i] = uint64(routeRelayPrices[i])
	}

	var routeCost int32 = sessionData.RouteCost
	if sessionData.RouteCost == math.MaxInt32 {
		routeCost = 0
	}

	var nearRelayRTT float32
	if sessionData.RouteNumRelays > 0 {
		for i, nearRelayID := range nearRelays.IDs {
			if nearRelayID == sessionData.RouteRelayIDs[0] {
				nearRelayRTT = float32(nearRelays.RTTs[i])
				break
			}
		}
	}

	debugString := ""
	if debug != nil {
		debugString = *debug
	}

	var numNearRelays uint8
	nearRelayIDs := [billing.BillingEntryMaxNearRelays]uint64{}
	nearRelayRTTs := [billing.BillingEntryMaxNearRelays]float32{}
	nearRelayJitters := [billing.BillingEntryMaxNearRelays]float32{}
	nearRelayPacketLosses := [billing.BillingEntryMaxNearRelays]float32{}

	if buyer.Debug {
		numNearRelays = uint8(nearRelays.Count)
		for i := uint8(0); i < numNearRelays; i++ {
			nearRelayIDs[i] = nearRelays.IDs[i]
			nearRelayRTTs[i] = float32(nearRelays.RTTs[i])
			nearRelayJitters[i] = float32(nearRelays.Jitters[i])
			nearRelayPacketLosses[i] = float32(nearRelays.PacketLosses[i])
		}
	}

	// // todo
	// slicePacketLoss := slicePacketLossClientToServer
	// if slicePacketLossServerToClient > slicePacketLossClientToServer {
	// 	slicePacketLoss = slicePacketLossServerToClient
	// }

	// Clamp jitter between client <-> server at 1000 (it is meaningless beyond that)
	if packet.JitterClientToServer > 1000.0 {
		packet.JitterClientToServer = float32(1000)
	}
	if packet.JitterServerToClient > 1000.0 {
		packet.JitterServerToClient = float32(1000)
	}

	billingEntry := &billing.BillingEntry{
		Timestamp:                       uint64(time.Now().Unix()),
		BuyerID:                         packet.BuyerID,
		UserHash:                        packet.UserHash,
		SessionID:                       packet.SessionID,
		SliceNumber:                     packet.SliceNumber,
		DirectRTT:                       packet.DirectRTT,
		DirectJitter:                    packet.DirectJitter,
		DirectPacketLoss:                packet.DirectPacketLoss,
		Next:                            packet.Next,
		NextRTT:                         packet.NextRTT,
		NextJitter:                      packet.NextJitter,
		NextPacketLoss:                  packet.NextPacketLoss,
		NumNextRelays:                   uint8(sessionData.RouteNumRelays),
		NextRelays:                      sessionData.RouteRelayIDs,
		TotalPrice:                      uint64(totalPrice),
		ClientToServerPacketsLost:       packet.PacketsLostClientToServer,
		ServerToClientPacketsLost:       packet.PacketsLostServerToClient,
		Committed:                       packet.Committed,
		Flagged:                         packet.Reported,
		Multipath:                       sessionData.RouteState.Multipath,
		Initial:                         sessionData.Initial,
		NextBytesUp:                     nextBytesUp,
		NextBytesDown:                   nextBytesDown,
		EnvelopeBytesUp:                 nextEnvelopeBytesUp,
		EnvelopeBytesDown:               nextEnvelopeBytesDown,
		DatacenterID:                    datacenter.ID,
		RTTReduction:                    sessionData.RouteState.ReduceLatency,
		PacketLossReduction:             sessionData.RouteState.ReducePacketLoss,
		NextRelaysPrice:                 nextRelaysPrice,
		Latitude:                        float32(sessionData.Location.Latitude),
		Longitude:                       float32(sessionData.Location.Longitude),
		ISP:                             sessionData.Location.ISP,
		ABTest:                          sessionData.RouteState.ABTest,
		RouteDecision:                   0, // todo: deprecated
		ConnectionType:                  uint8(packet.ConnectionType),
		PlatformType:                    uint8(packet.PlatformType),
		SDKVersion:                      packet.Version.String(),
		// todo
		// PacketLoss:                      slicePacketLoss,
		PredictedNextRTT:                float32(routeCost),
		MultipathVetoed:                 sessionData.RouteState.MultipathOverload,
		UseDebug:                        buyer.Debug,
		Debug:                           debugString,
		FallbackToDirect:                packet.FallbackToDirect,
		ClientFlags:                     packet.Flags,
		UserFlags:                       packet.UserFlags,
		NearRelayRTT:                    nearRelayRTT,
		PacketsOutOfOrderClientToServer: packet.PacketsOutOfOrderClientToServer,
		PacketsOutOfOrderServerToClient: packet.PacketsOutOfOrderServerToClient,
		JitterClientToServer:            packet.JitterClientToServer,
		JitterServerToClient:            packet.JitterServerToClient,
		NumNearRelays:                   numNearRelays,
		NearRelayIDs:                    nearRelayIDs,
		NearRelayRTTs:                   nearRelayRTTs,
		NearRelayJitters:                nearRelayJitters,
		NearRelayPacketLosses:           nearRelayPacketLosses,
		RelayWentAway:                   sessionData.RouteState.RelayWentAway,
		RouteLost:                       sessionData.RouteState.RouteLost,
		NumTags:                         uint8(packet.NumTags),
		Tags:                            packet.Tags,
		Mispredicted:                    sessionData.RouteState.Mispredict,
		Vetoed:                          sessionData.RouteState.Veto,
		LatencyWorse:                    sessionData.RouteState.LatencyWorse,
		NoRoute:                         sessionData.RouteState.NoRoute,
		NextLatencyTooHigh:              sessionData.RouteState.NextLatencyTooHigh,
		RouteChanged:                    sessionData.RouteChanged,
		CommitVeto:                      sessionData.RouteState.CommitVeto,
		RouteDiversity:                  uint32(routeDiversity),
		LackOfDiversity:                 sessionData.RouteState.LackOfDiversity,
		Pro:                             buyer.RouteShader.ProMode && !sessionData.RouteState.MultipathRestricted,
		MultipathRestricted:             sessionData.RouteState.MultipathRestricted,
		ClientToServerPacketsSent:       packet.PacketsSentClientToServer,
		ServerToClientPacketsSent:       packet.PacketsSentServerToClient,
		BuyerNotLive:                    buyerNotLive,
		UnknownDatacenter:               unknownDatacenter,
		DatacenterNotEnabled:            datacenterNotEnabled,
		StaleRouteMatrix:                staleRouteMatrix,
	}

	// send to the billing system (non-realtime path)

	postSessionHandler.SendBillingEntry(billingEntry)

	// send to vanity metrics (real-time path)

	if postSessionHandler.useVanityMetrics {
		postSessionHandler.SendVanityMetric(billingEntry)
	}

	// send data to the portal (real-time path)

	// todo
	// var portalData SessionPortalData

	// buildPortalData(&portalData,
	// 	packet,
	// 	sessionData,
	// 	buyer,
	// 	routeRelayNames,
	// 	routeRelaySellers,
	// 	nearRelays,
	// 	datacenter,
	// 	routeDiversity,
	// 	slicePacketLossClientToServer,
	// 	slicePacketLossServerToClient,
	// 	debug,
	// 	unknownDatacenter,
	// 	datacenterNotEnabled,
	// 	buyerNotLive,
	// 	staleRouteMatrix,
	// )

	// if portalData.Meta.NextRTT != 0 || portalData.Meta.DirectRTT != 0 {
	// 	postSessionHandler.SendPortalData(&portalData)
	// }

}
*/

// ------------------------------------------------------------------

func writeSessionResponse(w io.Writer, response *SessionResponsePacket, sessionData *SessionData) error {
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
	if _, err := w.Write(responseData); err != nil {
		return err
	}
	return nil
}

// ------------------------------------------------------------------

func SessionUpdateHandlerFunc(
	logger log.Logger,
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
		state.response = SessionResponsePacket{
			Version:     state.packet.Version,
			SessionID:   state.packet.SessionID,
			SliceNumber: state.packet.SliceNumber,
			RouteType:   routing.RouteTypeDirect,
		}

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
			Build the set of near relays for the SDK to ping.

			The SDK pings these near relays and reports up the results in the next session update.

			We hold the set of near relays fixed for the session, so we only do this work on the first slice.
		*/

		if state.packet.SliceNumber == 0 {
			sessionGetNearRelays(&state)
			core.Debug("first slice always goes direct")
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
