package transport

import (
	"fmt"
	"io"
	"net"
	"time"

	"github.com/go-kit/kit/log"

	"github.com/networknext/backend/modules/billing"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/envvar"
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
			and log here, so we can map the datacenter name to the datacenter id, when we are tracking it down.
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
			// todo: metrics for this
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

// todo: clean up
func HandleNextToken(
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
	sessionData.Initial = true

	// todo: this should be 2 * BillingSliceSeconds
	sessionData.ExpireTimestamp += billing.BillingSliceSeconds
	sessionData.SessionVersion++

	numTokens := routeNumRelays + 2 // relays + client + server
	routeAddresses, routePublicKeys := GetRouteAddressesAndPublicKeys(&packet.ClientAddress, packet.ClientRoutePublicKey, &packet.ServerAddress, packet.ServerRoutePublicKey, numTokens, routeRelays, allRelayIDs, database)
	if routeAddresses == nil || routePublicKeys == nil {
		response.RouteType = routing.RouteTypeDirect
		response.NumTokens = 0
		response.Tokens = nil
		return
	}

	tokenData := make([]byte, numTokens*routing.EncryptedNextRouteTokenSize)
	core.WriteRouteTokens(tokenData, sessionData.ExpireTimestamp, sessionData.SessionID, uint8(sessionData.SessionVersion), uint32(buyer.RouteShader.BandwidthEnvelopeUpKbps), uint32(buyer.RouteShader.BandwidthEnvelopeDownKbps), int(numTokens), routeAddresses, routePublicKeys, routerPrivateKey)
	response.RouteType = routing.RouteTypeNew
	response.NumTokens = numTokens
	response.Tokens = tokenData
}

// todo: clean up
func HandleContinueToken(
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
	numTokens := routeNumRelays + 2 // relays + client + server
	routeAddresses, routePublicKeys := GetRouteAddressesAndPublicKeys(&packet.ClientAddress, packet.ClientRoutePublicKey, &packet.ServerAddress, packet.ServerRoutePublicKey, numTokens, routeRelays, allRelayIDs, database)
	if routeAddresses == nil || routePublicKeys == nil {
		response.RouteType = routing.RouteTypeDirect
		response.NumTokens = 0
		response.Tokens = nil
		return
	}
	tokenData := make([]byte, numTokens*routing.EncryptedContinueRouteTokenSize)
	core.WriteContinueTokens(tokenData, sessionData.ExpireTimestamp, sessionData.SessionID, uint8(sessionData.SessionVersion), int(numTokens), routePublicKeys, routerPrivateKey)
	response.RouteType = routing.RouteTypeContinue
	response.NumTokens = numTokens
	response.Tokens = tokenData
}

// todo: clean up
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

	routeAddresses := make([]*net.UDPAddr, numTokens)
	routePublicKeys := make([][]byte, numTokens)

	routeAddresses[0] = clientAddress
	routePublicKeys[0] = clientPublicKey
	routeAddresses[numTokens-1] = serverAddress
	routePublicKeys[numTokens-1] = serverPublicKey

	totalNumRelays := int32(len(allRelayIDs))
	foundRelayCount := int32(0)

	enableInternalIPs, _ := envvar.GetBool("FEATURE_ENABLE_INTERNAL_IPS", false)

	for i := int32(0); i < numTokens-2; i++ {
		relayIndex := routeRelays[i]
		if relayIndex < totalNumRelays {
			relayID := allRelayIDs[relayIndex]
			relay, exists := database.RelayMap[relayID]
			if !exists {
				continue
			}

			routeAddresses = AddAddress(enableInternalIPs, i, relay, allRelayIDs, database, routeRelays, routeAddresses)

			routePublicKeys[i+1] = relay.PublicKey
			foundRelayCount++
		}
	}

	if foundRelayCount != numTokens-2 {
		return nil, nil
	}

	return routeAddresses, routePublicKeys
}

// todo: clean up
func AddAddress(enableInternalIPs bool, index int32, relay routing.Relay, allRelayIDs []uint64, database *routing.DatabaseBinWrapper, routeRelays []int32, routeAddresses []*net.UDPAddr) []*net.UDPAddr {
	totalNumRelays := int32(len(allRelayIDs))
	routeAddresses[index+1] = &relay.Addr
	if enableInternalIPs {
		// check if the previous relay is the same seller
		if index >= 1 {
			prevRelayIndex := routeRelays[index-1]
			if prevRelayIndex < totalNumRelays {
				prevID := allRelayIDs[prevRelayIndex]
				prev, exists := database.RelayMap[prevID]
				if exists && prev.Seller.ID == relay.Seller.ID && prev.InternalAddr.String() != ":0" && relay.InternalAddr.String() != ":0" {
					routeAddresses[index+1] = &relay.InternalAddr
				}
			}
		}
	}
	return routeAddresses
}

// ----------------------------------------------------------------------------

func routeMatrixIsStale(routeMatrix *routing.RouteMatrix, staleDuration time.Duration) bool {
	return routeMatrix.CreatedAt+uint64(staleDuration.Seconds()) < uint64(time.Now().Unix())
}

// ----------------------------------------------------------------------------

type SessionHandlerState struct {
	writer               io.Writer
	input                SessionData // sent up from the SDK
	output               SessionData // sent down to the SDK
	packet               SessionUpdatePacket
	response             SessionResponsePacket
	database             *routing.DatabaseBinWrapper
	routeMatrix          *routing.RouteMatrix
	datacenter           routing.Datacenter
	buyer                routing.Buyer
	debug                *string
	ipLocator            routing.IPLocator
	signatureCheckFailed bool
	unknownDatacenter    bool
	datacenterNotEnabled bool
	buyerNotFound        bool
	buyerNotLive         bool
	staleRouteMatrix     bool
	staleDuration        time.Duration
	slicePacketLoss      float32

	/*
		postSessionHandler *PostSessionHandler
		multipathVetoHandler storage.MultipathVetoHandler
		routeRelayNames [core.MaxRelaysPerRoute]string
		routeRelaySellers [core.MaxRelaysPerRoute]routing.Seller
		nearRelays nearRelayGroup
		routeDiversity int32
		slicePacketLossClientToServer float32
		slicePacketLossServerToClient float32
	*/
}

func sessionPre(state *SessionHandlerState) bool {

	if routeMatrixIsStale(state.routeMatrix, state.staleDuration) {
		core.Debug("stale route matrix")
		state.staleRouteMatrix = true
		return true
	}

	if state.packet.ClientPingTimedOut {
		core.Debug("client ping timed out")
		// todo: put metrics in state
		// metrics.ClientPingTimedOut.Add(1)
		return true
	}

	var exists bool
	state.buyer, exists = state.database.BuyerMap[state.packet.BuyerID]
	if !exists {
		core.Debug("buyer not found")
		// todo: put metrics in state
		// metrics.BuyerNotFound.Add(1)
		state.buyerNotFound = true
		return true
	}

	if !state.buyer.Live {
		core.Debug("buyer not live")
		// todo: put metrics in state
		// metrics.BuyerNotLive.Add(1)
		state.buyerNotLive = true
		return true
	}

	// todo: put packet data in handler state
	/*
		if !crypto.VerifyPacket(state.buyer.PublicKey, incoming.Data) {
			core.Debug("signature check failed")
			// todo: put metrics in state
			// metrics.SignatureCheckFailed.Add(1)
			state.signatureCheckFailed = true
			return true
		}
	*/

	if !datacenterExists(state.database, state.packet.DatacenterID) {
		core.Debug("unknown datacenter")
		// todo: add a metric for this condition
		state.unknownDatacenter = true
		return true
	}

	if !datacenterEnabled(state.database, state.packet.BuyerID, state.packet.DatacenterID) {
		core.Debug("datacenter not enabled")
		// todo: add a metric for this condition
		state.datacenterNotEnabled = true
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

	return false
}

func sessionUpdateNewSession(state *SessionHandlerState) {

	core.Debug("new session")

	var err error

	state.output.Location, err = state.ipLocator.LocateIP(state.packet.ClientAddress.IP)

	if err != nil || state.output.Location == routing.LocationNullIsland {
		core.Debug("location veto")
		// todo: put metrics in state
		// metrics.ClientLocateFailure.Add(1)
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
		// todo: put metrics in state
		// metrics.ReadSessionDataFailure.Add(1)
		return
	}

	if state.input.SessionID != state.packet.SessionID {
		core.Debug("bad session id")
		// todo: put metrics in state
		// metrics.BadSessionID.Add(1)
		return
	}

	if state.input.SliceNumber != state.packet.SliceNumber {
		core.Debug("bad slice number")
		// todo: put metrics in state
		// metrics.BadSliceNumber.Add(1)
		return
	}

	state.output = state.input
	state.output.SliceNumber += 1
	state.output.ExpireTimestamp += billing.BillingSliceSeconds

	// calculate real packet loss for this slice

	slicePacketsSentClientToServer := state.packet.PacketsSentClientToServer - state.input.PrevPacketsSentClientToServer
	slicePacketsSentServerToClient := state.packet.PacketsSentServerToClient - state.input.PrevPacketsSentServerToClient

	slicePacketsLostClientToServer := state.packet.PacketsLostClientToServer - state.input.PrevPacketsLostClientToServer
	slicePacketsLostServerToClient := state.packet.PacketsLostServerToClient - state.input.PrevPacketsLostServerToClient

	var slicePacketLossClientToServer float32
	if slicePacketsSentClientToServer != uint64(0) {
		slicePacketLossClientToServer = float32(float64(slicePacketsLostClientToServer)/float64(slicePacketsSentClientToServer)) * 100.0
	} else {
		slicePacketLossClientToServer = float32(0)
	}

	var slicePacketLossServerToClient float32
	if slicePacketsSentServerToClient != uint64(0) {
		slicePacketLossServerToClient = float32(float64(slicePacketsLostServerToClient)/float64(slicePacketsSentServerToClient)) * 100.0
	} else {
		slicePacketLossServerToClient = float32(0)
	}

	state.slicePacketLoss = slicePacketLossClientToServer
	if slicePacketLossServerToClient > slicePacketLossClientToServer {
		state.slicePacketLoss = slicePacketLossServerToClient
	}
}

func sessionHandleFallbackToDirect(state *SessionHandlerState) bool {

	if state.packet.FallbackToDirect && !state.output.FellBackToDirect {

		core.Debug("fallback to direct")

		state.output.FellBackToDirect = true

		reported := false

		if state.packet.Flags&FallbackFlagsBadRouteToken != 0 {
			// todo
			// state.metrics.FallbackToDirectBadRouteToken.Add(1)
			reported = true
		}

		if state.packet.Flags&FallbackFlagsNoNextRouteToContinue != 0 {
			// todo
			// state.metrics.FallbackToDirectNoNextRouteToContinue.Add(1)
			reported = true
		}

		if state.packet.Flags&FallbackFlagsPreviousUpdateStillPending != 0 {
			// todo
			// state.metrics.FallbackToDirectPreviousUpdateStillPending.Add(1)
			reported = true
		}

		if state.packet.Flags&FallbackFlagsBadContinueToken != 0 {
			// todo
			// metrics.FallbackToDirectBadContinueToken.Add(1)
			reported = true
		}

		if state.packet.Flags&FallbackFlagsRouteExpired != 0 {
			// todo
			// metrics.FallbackToDirectRouteExpired.Add(1)
			reported = true
		}

		if state.packet.Flags&FallbackFlagsRouteRequestTimedOut != 0 {
			// todo
			// metrics.FallbackToDirectRouteRequestTimedOut.Add(1)
			reported = true
		}

		if state.packet.Flags&FallbackFlagsContinueRequestTimedOut != 0 {
			// todo
			// metrics.FallbackToDirectContinueRequestTimedOut.Add(1)
			reported = true
		}

		if state.packet.Flags&FallbackFlagsClientTimedOut != 0 {
			// todo
			// metrics.FallbackToDirectClientTimedOut.Add(1)
			reported = true
		}

		if state.packet.Flags&FallbackFlagsUpgradeResponseTimedOut != 0 {
			// todo
			// metrics.FallbackToDirectUpgradeResponseTimedOut.Add(1)
			reported = true
		}

		if state.packet.Flags&FallbackFlagsRouteUpdateTimedOut != 0 {
			// todo
			// metrics.FallbackToDirectRouteUpdateTimedOut.Add(1)
			reported = true
		}

		if state.packet.Flags&FallbackFlagsDirectPongTimedOut != 0 {
			// todo
			// metrics.FallbackToDirectDirectPongTimedOut.Add(1)
			reported = true
		}

		if state.packet.Flags&FallbackFlagsNextPongTimedOut != 0 {
			// todo
			// metrics.FallbackToDirectNextPongTimedOut.Add(1)
			reported = true
		}

		if !reported {
			// todo
			// metrics.FallbackToDirectUnknownReason.Add(1)
		}

		return true
	}

	return false
}

func sessionGetNearRelays(state *SessionHandlerState) {

	// todo

	/*
		nearRelayIDs := routeMatrix.GetNearRelays(float32(directLatency), clientLat, clientLong, serverLat, serverLong, maxNearRelays)
		if len(nearRelayIDs) == 0 {
			core.Debug("no near relays :(")
			return false, nearRelayGroup{}, nil, errors.New("no near relays")
		}

		nearRelays := newNearRelayGroup(int32(len(nearRelayIDs)))
		for i := int32(0); i < nearRelays.Count; i++ {
			relayIndex, ok := routeMatrix.RelayIDsToIndices[nearRelayIDs[i]]
			if !ok {
				continue
			}

			nearRelays.IDs[i] = nearRelayIDs[i]
			nearRelays.Addrs[i] = routeMatrix.RelayAddresses[relayIndex]
			nearRelays.Names[i] = routeMatrix.RelayNames[relayIndex]
		}

		routeState.NumNearRelays = nearRelays.Count
		return true, nearRelays, nil, nil
	*/
}

/*
// todo: simplify the fuck out of this garbage

incomingNearRelays := newNearRelayGroup(packet.NumNearRelays)
for i := int32(0); i < incomingNearRelays.Count; i++ {
	incomingNearRelays.IDs[i] = packet.NearRelayIDs[i]
	incomingNearRelays.RTTs[i] = packet.NearRelayRTT[i]
	incomingNearRelays.Jitters[i] = packet.NearRelayJitter[i]
	incomingNearRelays.PacketLosses[i] = packet.NearRelayPacketLoss[i]

	// The SDK doesn't send up the relay name or relay address, so we have to get those from the route matrix
	relayIndex, ok := routeMatrix.RelayIDsToIndices[packet.NearRelayIDs[i]]
	if !ok {
		// todo: we should catch this condition with  metric
		continue
	}

	incomingNearRelays.Addrs[i] = routeMatrix.RelayAddresses[relayIndex]
	incomingNearRelays.Names[i] = routeMatrix.RelayNames[relayIndex]
}

nearRelaysChanged, nearRelays, reframedDestRelays, err := handleNearAndDestRelays(
	int32(packet.SliceNumber),
	routeMatrix,
	incomingNearRelays,
	&buyer.RouteShader,
	&sessionData.RouteState,
	newSession,
	sessionData.Location.Latitude,
	sessionData.Location.Longitude,
	state.datacenter.Location.Latitude,
	state.datacenter.Location.Longitude,
	maxNearRelays,
	int32(math.Ceil(float64(packet.DirectRTT))),
	int32(math.Ceil(float64(packet.DirectJitter))),
	int32(math.Floor(float64(slicePacketLoss)+0.5)),
	int32(math.Floor(float64(packet.NextPacketLoss)+0.5)),
	sessionData.RouteRelayIDs[0],
	destRelayIDs,
	state.debug,
)

response.NumNearRelays = nearRelays.Count
response.NearRelayIDs = nearRelays.IDs
response.NearRelayAddresses = nearRelays.Addrs
response.NearRelaysChanged = nearRelaysChanged
response.HighFrequencyPings = buyer.InternalConfig.HighFrequencyPings

if err != nil {
	// todo: string comparison in hot path?!
	if strings.HasPrefix(err.Error(), "near relays changed") {
		core.Debug("near relays changed")
		metrics.NearRelaysChanged.Add(1)
	} else {
		core.Debug("failed to get near relays")
		metrics.NearRelaysLocateFailure.Add(1)
	}

	return
}
*/

func sessionPost(state *SessionHandlerState) {

	if state.buyerNotFound || state.signatureCheckFailed {
		core.Debug("not responding")
		return
	}

	if state.response.RouteType != routing.RouteTypeDirect {
		core.Debug("session takes network next")
		// todo: put metrics in state
		// state.metrics.NextSlices.Add(1)
		state.output.EverOnNext = true
	} else {
		core.Debug("session goes direct")
		// todo: put metrics in state
		// metrics.DirectSlices.Add(1)
	}

	state.output.PrevPacketsSentClientToServer = state.packet.PacketsSentClientToServer
	state.output.PrevPacketsSentServerToClient = state.packet.PacketsSentServerToClient
	state.output.PrevPacketsLostClientToServer = state.packet.PacketsLostClientToServer
	state.output.PrevPacketsLostServerToClient = state.packet.PacketsLostServerToClient

	if state.debug != nil {
		state.response.Debug = *state.debug
		if state.response.Debug != "" {
			state.response.HasDebug = true
		}
	}

	if err := writeSessionResponse(state.writer, &state.response, &state.output); err != nil {
		core.Debug("failed to write session update response: %s", err)
		// todo: metrics in state
		// metrics.WriteResponseFailure.Add(1)
		return
	}

	// Rebuild the arrays of route relay names and sellers from the previous session data
	// routeRelayNames := [core.MaxRelaysPerRoute]string{}
	// routeRelaySellers := [core.MaxRelaysPerRoute]routing.Seller{}

	// todo
	/*
		for i := int32(0); i < prevSessionData.RouteNumRelays; i++ {
			for _, relay := range state.database.Relays {
				if relay.ID == prevSessionData.RouteRelayIDs[i] {
					routeRelayNames[i] = relay.Name
					routeRelaySellers[i] = relay.Seller
					break
				}
			}
		}
	*/

	// todo

	// Rebuild the near relays from the previous session data
	// var nearRelays nearRelayGroup

	// Make sure we only rebuild the previous near relays if we haven't gotten out of sync somehow
	/*
		if prevSessionData.RouteState.NumNearRelays == packet.NumNearRelays {
			nearRelays = newNearRelayGroup(prevSessionData.RouteState.NumNearRelays)
		}

		for i := int32(0); i < nearRelays.Count; i++ {

			// Since we now guarantee that the near relay IDs reported up from the SDK each slice don't change,
			// we can use the packet's near relay IDs here instead of storing the near relay IDs in the session data
			relayID := packet.NearRelayIDs[i]

			// Make sure to check if the relay exists in case the near relays are gone
			// this slice compared to the previous slice
			relayIndex, ok := routeMatrix.RelayIDsToIndices[relayID]
			if !ok {
				continue
			}

			nearRelays.IDs[i] = relayID
			nearRelays.Names[i] = routeMatrix.RelayNames[relayIndex]
			nearRelays.Addrs[i] = routeMatrix.RelayAddresses[relayIndex]
			nearRelays.RTTs[i] = prevSessionData.RouteState.NearRelayRTT[i]
			nearRelays.Jitters[i] = prevSessionData.RouteState.NearRelayJitter[i]

			// We don't actually store the packet loss in the session data, so just use the
			// values from the session update packet (no max history)
			if nearRelays.RTTs[i] >= 255 {
				nearRelays.PacketLosses[i] = 100
			} else {
				nearRelays.PacketLosses[i] = packet.NearRelayPacketLoss[i]
			}
		}
	*/

	if !state.packet.ClientPingTimedOut {

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
}

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
	maxNearRelays int,
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
			Build session handler state. Putting everything in a struct makes calling subroutines easier.
		*/

		state.writer = w

		state.database = getDatabase()

		state.datacenter = routing.UnknownDatacenter

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
			Session post always runs at the end of this function

			It sends session data to:

				1. Billing
				2. Vanity Metrics
				3. Portal

			by pushing session data to various queues.

			It also writes and sends the response packet back to the sender.
		*/

		defer sessionPost(&state)

		/*
			Call session pre function

			This function checks for early out conditions and does some setup of the handler state.

			If it returns true, it means that one of the early out conditions has been met, so we return.

			IMPORTANT. Returning here kicks off the sessionPost which was deferred, sending the response packet.
		*/

		if sessionPre(&state) {
			return
		}

		/*
			Update the session

			Do setup on slice 0, then for subsequent slices transform state.input -> state.output
		*/

		if state.packet.SliceNumber == 0 {

			sessionUpdateNewSession(&state)

		} else {

			sessionUpdateExistingSession(&state)

		}

		/*
			Handle fallback to direct.

			Fallback to direct is a condition where the SDK indicates that it has seen
			some fatal error, like not getting a response from the backend in time,
			and has decided to go direct for the rest of the session.

			When this happens, we early out to save processing time.
		*/

		if sessionHandleFallbackToDirect(&state) {
			return
		}

		/*
			Are there any relays in the datacenter? If not then we must go direct.
		*/

		destRelayIDs := state.routeMatrix.GetDatacenterRelayIDs(state.datacenter.ID)
		if len(destRelayIDs) == 0 {
			core.Debug("no relays in datacenter %x", state.datacenter.ID)
			metrics.NoRelaysInDatacenter.Add(1)
			return
		}

		/*
			Build set of near relays to return to the SDK.

			The SDK pings these near relays and reports up the results in the next session update.

			We hold the set of near relays fixed for the session, so we only do this work on the first slice.
		*/

		if state.packet.SliceNumber == 0 {
			sessionGetNearRelays(&state)
			core.Debug("first slice always goes direct")
			return
		}

		// ----------------------

		// todo: break this down

		/*
			var routeCost int32
			routeRelays := [core.MaxRelaysPerRoute]int32{}

			sessionData.Initial = false

			multipathVetoMap := multipathVetoHandler.GetMapCopy(buyer.CompanyCode)

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

			var routeNumRelays int32

			var nextRouteSwitched bool

			if !sessionData.RouteState.Next || sessionData.RouteNumRelays == 0 {
				sessionData.RouteState.Next = false
				if core.MakeRouteDecision_TakeNetworkNext(routeMatrix.RouteEntries, &buyer.RouteShader, &sessionData.RouteState, multipathVetoMap, &buyer.InternalConfig, int32(packet.DirectRTT), slicePacketLoss, nearRelayIndices, nearRelayCosts, reframedDestRelays, &routeCost, &routeNumRelays, routeRelays[:], &routeDiversity, state.debug) {
					HandleNextToken(&sessionData, state.database, &buyer, &packet, routeNumRelays, routeRelays[:], routeMatrix.RelayIDs, routerPrivateKey, &response)
				}
			} else {
				if !core.ReframeRoute(&sessionData.RouteState, routeMatrix.RelayIDsToIndices, sessionData.RouteRelayIDs[:sessionData.RouteNumRelays], &routeRelays) {
					routeRelays = [core.MaxRelaysPerRoute]int32{}
					core.Debug("one or more relays in the route no longer exist")
					metrics.RouteDoesNotExist.Add(1)
				}

				if !packet.Next {

					// the sdk "aborted" this session

					core.Debug("aborted")
					sessionData.RouteState.Next = false
					sessionData.RouteState.Veto = true
					metrics.SDKAborted.Add(1)

				} else {
					var stay bool
					if stay, nextRouteSwitched = core.MakeRouteDecision_StayOnNetworkNext(routeMatrix.RouteEntries, routeMatrix.RelayNames, &buyer.RouteShader, &sessionData.RouteState, &buyer.InternalConfig, int32(packet.DirectRTT), int32(packet.NextRTT), sessionData.RouteCost, slicePacketLoss, packet.NextPacketLoss, sessionData.RouteNumRelays, routeRelays, nearRelayIndices, nearRelayCosts, reframedDestRelays, &routeCost, &routeNumRelays, routeRelays[:], state.debug); stay {

						// stay on network next

						if nextRouteSwitched {
							core.Debug("route changed")
							metrics.RouteSwitched.Add(1)
							// todo: this is where we need to set the double length (Initial)
							HandleNextToken(&sessionData, state.database, &buyer, &packet, routeNumRelays, routeRelays[:], routeMatrix.RelayIDs, routerPrivateKey, &response)
						} else {
							core.Debug("route continued")
							HandleContinueToken(&sessionData, state.database, &buyer, &packet, routeNumRelays, routeRelays[:], routeMatrix.RelayIDs, routerPrivateKey, &response)
						}

					} else {

						// leave network next

						if sessionData.RouteState.NoRoute {
							core.Debug("route no longer exists")
							metrics.NoRoute.Add(1)
						}

						if sessionData.RouteState.MultipathOverload {
							core.Debug("multipath overload")
							metrics.MultipathOverload.Add(1)
						}

						if sessionData.RouteState.Mispredict {
							core.Debug("mispredict")
							metrics.MispredictVeto.Add(1)
						}

						if sessionData.RouteState.LatencyWorse {
							core.Debug("latency worse")
							metrics.LatencyWorse.Add(1)
						}
					}
				}
			}

			if routeCost > routing.InvalidRouteValue {
				routeCost = routing.InvalidRouteValue
			}

			response.Committed = sessionData.RouteState.Committed
			response.Multipath = sessionData.RouteState.Multipath

			// Store the route back into the session data
			sessionData.RouteNumRelays = routeNumRelays
			sessionData.RouteCost = routeCost
			sessionData.RouteChanged = nextRouteSwitched

			for i := int32(0); i < routeNumRelays; i++ {
				relayID := routeMatrix.RelayIDs[routeRelays[i]]
				sessionData.RouteRelayIDs[i] = relayID
			}
		*/

		core.Debug("session updated successfully")
	}
}

// ----------------------------------------------------------------------------

/*
type nearRelayGroup struct {
	Count int32
	// todo: allocation here is bad. we should instead make these fixed sized arrays and have this all on the stack
	IDs          []uint64
	Addrs        []net.UDPAddr
	Names        []string
	RTTs         []int32
	Jitters      []int32
	PacketLosses []int32
}

func newNearRelayGroup(count int32) nearRelayGroup {
	return nearRelayGroup{
		Count:        count,
		IDs:          make([]uint64, count),
		Addrs:        make([]net.UDPAddr, count),
		Names:        make([]string, count),
		RTTs:         make([]int32, count),
		Jitters:      make([]int32, count),
		PacketLosses: make([]int32, count),
	}
}

func (n nearRelayGroup) Copy(other *nearRelayGroup) {

	// todo: allocations galore. we don't want this!

	other.Count = n.Count
	other.IDs = make([]uint64, n.Count)
	other.Addrs = make([]net.UDPAddr, n.Count)
	other.Names = make([]string, n.Count)
	other.RTTs = make([]int32, n.Count)
	other.Jitters = make([]int32, n.Count)
	other.PacketLosses = make([]int32, n.Count)

	copy(other.IDs, n.IDs)
	copy(other.Addrs, n.Addrs)
	copy(other.Names, n.Names)
	copy(other.RTTs, n.RTTs)
	copy(other.Jitters, n.Jitters)
	copy(other.PacketLosses, n.PacketLosses)
}

func handleNearAndDestRelays(
	sliceNumber int32,
	routeMatrix *routing.RouteMatrix,
	incomingNearRelays nearRelayGroup,
	routeShader *core.RouteShader,
	routeState *core.RouteState,
	newSession bool,
	clientLat float32,
	clientLong float32,
	serverLat float32,
	serverLong float32,
	maxNearRelays int,
	directLatency int32,
	directJitter int32,
	directPacketLoss int32,
	nextPacketLoss int32,
	firstRouteRelayID uint64,
	destRelayIDs []uint64,
	debug *string,
) (bool, nearRelayGroup, []int32, error) {
	if newSession {
		nearRelayIDs := routeMatrix.GetNearRelays(float32(directLatency), clientLat, clientLong, serverLat, serverLong, maxNearRelays)
		if len(nearRelayIDs) == 0 {
			core.Debug("no near relays :(")
			return false, nearRelayGroup{}, nil, errors.New("no near relays")
		}

		nearRelays := newNearRelayGroup(int32(len(nearRelayIDs)))
		for i := int32(0); i < nearRelays.Count; i++ {
			relayIndex, ok := routeMatrix.RelayIDsToIndices[nearRelayIDs[i]]
			if !ok {
				continue
			}

			nearRelays.IDs[i] = nearRelayIDs[i]
			nearRelays.Addrs[i] = routeMatrix.RelayAddresses[relayIndex]
			nearRelays.Names[i] = routeMatrix.RelayNames[relayIndex]
		}

		routeState.NumNearRelays = nearRelays.Count
		return true, nearRelays, nil, nil
	}

	var nearRelays nearRelayGroup
	incomingNearRelays.Copy(&nearRelays)

	if nearRelays.Count != routeState.NumNearRelays {
		return false, nearRelayGroup{}, nil, fmt.Errorf("near relays changed from %d to %d", routeState.NumNearRelays, nearRelays.Count)
	}

	var numDestRelays int32
	reframedDestRelays := make([]int32, len(destRelayIDs))

	core.ReframeRelays(routeShader, routeState, routeMatrix.RelayIDsToIndices, directLatency, directJitter, directPacketLoss, nextPacketLoss, firstRouteRelayID, sliceNumber, incomingNearRelays.IDs, incomingNearRelays.RTTs, incomingNearRelays.Jitters, incomingNearRelays.PacketLosses, destRelayIDs, nearRelays.RTTs, nearRelays.Jitters, &numDestRelays, reframedDestRelays)

	return false, nearRelays, reframedDestRelays[:numDestRelays], nil
}
*/

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

func buildPortalData(state *SessionHandlerState, portalData *SessionPortalData) {

	// todo: switch to using session handler state

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
