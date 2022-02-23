package transport

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net"
	"net/http"
	"strings"
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
	value, exists := database.DatacenterMap[datacenterID]

	if !exists {
		return routing.UnknownDatacenter
	}

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

func ServerInitHandlerFunc(getDatabase func() *routing.DatabaseBinWrapper, ServerTracker *storage.ServerTracker, metrics *metrics.ServerInitMetrics) UDPHandlerFunc {

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

		// Track which servers are initing
		// This is where we get the datacenter name
		if strings.TrimSpace(packet.DatacenterName) == "" {
			ServerTracker.AddServer(packet.BuyerID, packet.DatacenterID, incoming.From, "unknown_init")
		} else {
			ServerTracker.AddServer(packet.BuyerID, packet.DatacenterID, incoming.From, packet.DatacenterName)
		}

		/*
			IMPORTANT: When the datacenter doesn't exist, we intentionally let the server init succeed anyway
			and just log here, so we can map the datacenter name to the datacenter id, when we are tracking it down.
		*/

		if !datacenterExists(database, packet.DatacenterID) {
			// core.Error("unknown datacenter %s [%016x, %s] for buyer id %016x", packet.DatacenterName, packet.DatacenterID, incoming.From.String(), packet.BuyerID)
			metrics.DatacenterNotFound.Add(1)
			return
		}

		core.Debug("server is in datacenter \"%s\" [%x]", packet.DatacenterName, packet.DatacenterID)

		core.Debug("server initialized successfully")
	}
}

// ----------------------------------------------------------------------------

func ServerTrackerHandlerFunc(serverTracker *storage.ServerTracker) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		defer r.Body.Close()

		w.Header().Set("Content-Type", "application/json")

		serverTracker.TrackerMutex.RLock()
		defer serverTracker.TrackerMutex.RUnlock()
		json.NewEncoder(w).Encode(serverTracker.Tracker)
	}
}

// ----------------------------------------------------------------------------

func ServerUpdateHandlerFunc(getDatabase func() *routing.DatabaseBinWrapper, PostSessionHandler *PostSessionHandler, ServerTracker *storage.ServerTracker, metrics *metrics.ServerUpdateMetrics) UDPHandlerFunc {

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
		PostSessionHandler.SendPortalCounts(countData)

		if !datacenterExists(database, packet.DatacenterID) {
			core.Debug("datacenter does not exist %x", packet.DatacenterID)
			metrics.DatacenterNotFound.Add(1)
			// Track this server with unknown datacenter name
			ServerTracker.AddServer(buyer.ID, packet.DatacenterID, packet.ServerAddress, "unknown_update")
			return
		}

		// The server is a known datacenter, track it using the correct datacenter name from the bin file
		if datacenter, exists := database.DatacenterMap[packet.DatacenterID]; exists {
			ServerTracker.AddServer(buyer.ID, packet.DatacenterID, packet.ServerAddress, datacenter.Name)
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

func CalculateTotalPriceNibblins(routeNumRelays int, relaySellers [core.MaxRelaysPerRoute]routing.Seller, relayEgressPriceOverride [core.MaxRelaysPerRoute]routing.Nibblin, envelopeBytesUp uint64, envelopeBytesDown uint64) routing.Nibblin {

	if routeNumRelays == 0 {
		return 0
	}

	envelopeUpGB := float64(envelopeBytesUp) / 1000000000.0
	envelopeDownGB := float64(envelopeBytesDown) / 1000000000.0

	relayPriceNibblinsPerGB := routing.Nibblin(0)
	for i, seller := range relaySellers {
		if relayEgressPriceOverride[i] > 0 {
			relayPriceNibblinsPerGB += relayEgressPriceOverride[i]
		} else {
			relayPriceNibblinsPerGB += seller.EgressPriceNibblinsPerGB
		}
	}

	nextPriceNibblinsPerGB := routing.Nibblin(1e9)
	totalPriceNibblins := float64(relayPriceNibblinsPerGB+nextPriceNibblinsPerGB) * (envelopeUpGB + envelopeDownGB)

	return routing.Nibblin(totalPriceNibblins)
}

func CalculateRouteRelaysPrice(routeNumRelays int, relaySellers [core.MaxRelaysPerRoute]routing.Seller, relayEgressPriceOverride [core.MaxRelaysPerRoute]routing.Nibblin, envelopeBytesUp uint64, envelopeBytesDown uint64) [core.MaxRelaysPerRoute]routing.Nibblin {

	relayPrices := [core.MaxRelaysPerRoute]routing.Nibblin{}

	if routeNumRelays == 0 {
		return relayPrices
	}

	envelopeUpGB := float64(envelopeBytesUp) / 1000000000.0
	envelopeDownGB := float64(envelopeBytesDown) / 1000000000.0

	for i := 0; i < len(relayPrices); i++ {
		var basePrice float64

		if relayEgressPriceOverride[i] > 0 {
			basePrice = float64(relayEgressPriceOverride[i])
		} else {
			basePrice = float64(relaySellers[i].EgressPriceNibblinsPerGB)
		}

		relayPriceNibblins := basePrice * (envelopeUpGB + envelopeDownGB)
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

		relayAddresses[i] = &relay.Addr

		if i == 0 && relay.InternalAddressClientRoutable && relay.InternalAddr.String() != ":0" {
			/*
				If the relay is the first hop and has an internal address
				that can be pinged by the client, prefer to use this address
				instead of the external address.
			*/

			relayAddresses[i] = &relay.InternalAddr
		} else if i > 0 {
			/*
				If the relay has a private address defined and the previous relay in the route
				is from the same seller, prefer to send to the relay private address instead.
				These private addresses often have better performance than the public addresses,
				and in the case of google cloud, have cheaper bandwidth prices.
			*/

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

	Input SessionData // sent up from the SDK. previous slice.

	Output SessionData // sent down to the SDK. current slice.

	Writer             io.Writer
	Packet             SessionUpdatePacket
	Response           SessionResponsePacket
	PacketData         []byte
	Metrics            *metrics.SessionUpdateMetrics
	Database           *routing.DatabaseBinWrapper
	RouteMatrix        *routing.RouteMatrix
	Datacenter         routing.Datacenter
	Buyer              routing.Buyer
	Debug              *string
	IpLocator          *routing.MaxmindDB
	StaleDuration      time.Duration
	RouterPrivateKey   [crypto.KeySize]byte
	PostSessionHandler *PostSessionHandler

	// flags
	SignatureCheckFailed bool
	UnknownDatacenter    bool
	DatacenterNotEnabled bool
	BuyerNotFound        bool
	BuyerNotLive         bool
	StaleRouteMatrix     bool

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
	PostNearRelayCount                int
	PostNearRelayIDs                  [core.MaxNearRelays]uint64
	PostNearRelayNames                [core.MaxNearRelays]string
	PostNearRelayAddresses            [core.MaxNearRelays]net.UDPAddr
	PostNearRelayRTT                  [core.MaxNearRelays]float32
	PostNearRelayJitter               [core.MaxNearRelays]float32
	PostNearRelayPacketLoss           [core.MaxNearRelays]float32
	PostRouteRelayNames               [core.MaxRelaysPerRoute]string
	PostRouteRelaySellers             [core.MaxRelaysPerRoute]routing.Seller
	PostRouteRelayEgressPriceOverride [core.MaxRelaysPerRoute]routing.Nibblin
	PostRealPacketLossClientToServer  float32
	PostRealPacketLossServerToClient  float32

	// for convenience
	UnmarshaledSessionData bool

	// todo
	/*
		multipathVetoHandler storage.MultipathVetoHandler
	*/
}

func SessionPre(state *SessionHandlerState) bool {

	var exists bool
	var err error

	state.Buyer, exists = state.Database.BuyerMap[state.Packet.BuyerID]
	if !exists {
		core.Debug("buyer not found")
		state.Metrics.BuyerNotFound.Add(1)
		state.BuyerNotFound = true
		return true
	}

	if !state.Buyer.Live {
		core.Debug("buyer not live")
		state.Metrics.BuyerNotLive.Add(1)
		state.BuyerNotLive = true
		return true
	}

	if !crypto.VerifyPacket(state.Buyer.PublicKey, state.PacketData) {
		core.Debug("signature check failed")
		state.Metrics.SignatureCheckFailed.Add(1)
		state.SignatureCheckFailed = true
		return true
	}

	if state.Packet.ClientPingTimedOut {
		core.Debug("client ping timed out")
		state.Metrics.ClientPingTimedOut.Add(1)

		if state.PostSessionHandler.featureBilling2 {
			// Unmarshal the session data into the input to verify if we wrote the summary slice in sessionPost()
			err := UnmarshalSessionData(&state.Input, state.Packet.SessionData[:])

			if err != nil {
				core.Error("SessionPre(): ClientPingTimedOut: could not read session data for buyer %016x:\n\n%s\n", state.Buyer.ID, err)
				state.Metrics.ReadSessionDataFailure.Add(1)
				return true
			}

			state.UnmarshaledSessionData = true
		}

		return true
	}

	if state.Packet.SliceNumber == 0 {
		state.Output.Location, err = state.IpLocator.LocateIP(state.Packet.ClientAddress.IP, state.Packet.SessionID)

		if err != nil || state.Output.Location == routing.LocationNullIsland {
			core.Error("location veto: %s\n", err)
			state.Metrics.ClientLocateFailure.Add(1)
			state.Input.Location = routing.LocationNullIsland
			state.Output.RouteState.LocationVeto = true
			return true
		}

		// Always assign location no matter the outcome of SessionPre() on the first slice
		defer func() {
			state.Input.Location = state.Output.Location
		}()
	} else {

		/*
			For existing sessions, read in the input state from the session data.

			This is the state.Output from the previous slice.

			We do this in SessionPre() rather than SessionUpdateExistingSession()
			in case we early out later on in SessionPre() to ensure location is
			written back to the SDK.
		*/

		defer func() {
			err := UnmarshalSessionData(&state.Input, state.Packet.SessionData[:])

			if err != nil {
				core.Error("SessionPre(): could not read session data for buyer %016x:\n\n%s\n", state.Buyer.ID, err)
				state.Metrics.ReadSessionDataFailure.Add(1)
			} else {
				state.Output.Location = state.Input.Location
				state.UnmarshaledSessionData = true
			}
		}()
	}

	/*
		If the buyer is "Analysis Only", allow the session to proceed
		even if the datacenter does not exist, is not enabled, or has zero
		destination relays in the database.

		The session will always go direct since the Route State will be disabled.

		The billing entry will still contain the UnknownDatacenter flag to let
		us know if we need to add this datacenter for the buyer.

		It does not make sense to record the DatacenterNotEnabled flag or
		NoRelaysInDatacenter metric for an "Analysis Only" buyer.
	*/

	if !datacenterExists(state.Database, state.Packet.DatacenterID) {
		core.Debug("unknown datacenter")
		state.Metrics.DatacenterNotFound.Add(1)
		state.UnknownDatacenter = true

		if !state.Buyer.RouteShader.AnalysisOnly {
			return true
		}
	}

	if !datacenterEnabled(state.Database, state.Packet.BuyerID, state.Packet.DatacenterID) && !state.Buyer.RouteShader.AnalysisOnly {
		core.Debug("datacenter not enabled")
		state.Metrics.DatacenterNotEnabled.Add(1)
		state.DatacenterNotEnabled = true
		return true
	}

	state.Datacenter = getDatacenter(state.Database, state.Packet.DatacenterID)

	destRelayIDs := state.RouteMatrix.GetDatacenterRelayIDs(state.Packet.DatacenterID)
	if len(destRelayIDs) == 0 && !state.Buyer.RouteShader.AnalysisOnly {
		core.Debug("no relays in datacenter %x", state.Packet.DatacenterID)
		state.Metrics.NoRelaysInDatacenter.Add(1)
		return true
	}

	if state.RouteMatrix.CreatedAt+uint64(state.StaleDuration.Seconds()) < uint64(time.Now().Unix()) {
		core.Debug("stale route matrix")
		state.StaleRouteMatrix = true
		state.Metrics.StaleRouteMatrix.Add(1)
		return true
	}

	if state.Buyer.Debug {
		core.Debug("debug enabled")
		state.Debug = new(string)
	}

	for i := int32(0); i < state.Packet.NumTags; i++ {
		if state.Packet.Tags[i] == crypto.HashID("pro") {
			core.Debug("pro mode enabled")
			state.Buyer.RouteShader.ProMode = true
		}
	}

	state.Output.Initial = false

	return false
}

func SessionUpdateNewSession(state *SessionHandlerState) {

	core.Debug("new session")

	state.Output.Version = SessionDataVersion
	state.Output.SessionID = state.Packet.SessionID
	state.Output.SliceNumber = state.Packet.SliceNumber + 1
	state.Output.ExpireTimestamp = uint64(time.Now().Unix()) + billing.BillingSliceSeconds
	state.Output.RouteState.UserID = state.Packet.UserHash
	state.Output.RouteState.ABTest = state.Buyer.RouteShader.ABTest

	state.Input = state.Output
}

func SessionUpdateExistingSession(state *SessionHandlerState) {

	core.Debug("existing session")

	if !state.UnmarshaledSessionData {

		/*
			If for some reason we did not unmarshal the SessionData
			already, we must do it here for existing sessions.
		*/

		err := UnmarshalSessionData(&state.Input, state.Packet.SessionData[:])

		if err != nil {
			core.Error("SessionUpdateExistingSession(): could not read session data for buyer %016x:\n\n%s\n", state.Buyer.ID, err)
			state.Metrics.ReadSessionDataFailure.Add(1)
			return
		}
	}

	/*
		Check for some obviously divergent data between the session request packet
		and the stored session data. If there is a mismatch, just return a direct route.
	*/

	if state.Input.SessionID != state.Packet.SessionID {
		core.Debug("bad session id")
		state.Metrics.BadSessionID.Add(1)
		return
	}

	if state.Input.SliceNumber != state.Packet.SliceNumber {
		core.Debug("bad slice number")
		state.Metrics.BadSliceNumber.Add(1)
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
	state.Output.ExpireTimestamp += billing.BillingSliceSeconds

	/*
		Calculate real packet loss.

		This is driven from actual game packets, not ping packets.

		This value is typically much higher precision (60HZ), vs. ping packets (10HZ).
	*/

	slicePacketsSentClientToServer := state.Packet.PacketsSentClientToServer - state.Input.PrevPacketsSentClientToServer
	slicePacketsSentServerToClient := state.Packet.PacketsSentServerToClient - state.Input.PrevPacketsSentServerToClient

	slicePacketsLostClientToServer := state.Packet.PacketsLostClientToServer - state.Input.PrevPacketsLostClientToServer
	slicePacketsLostServerToClient := state.Packet.PacketsLostServerToClient - state.Input.PrevPacketsLostServerToClient

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

	if state.Packet.JitterClientToServer > 1000.0 {
		state.Packet.JitterClientToServer = float32(1000)
	}

	if state.Packet.JitterServerToClient > 1000.0 {
		state.Packet.JitterServerToClient = float32(1000)
	}

	state.RealJitter = state.Packet.JitterClientToServer
	if state.Packet.JitterServerToClient > state.Packet.JitterClientToServer {
		state.RealJitter = state.Packet.JitterServerToClient
	}
}

func SessionHandleFallbackToDirect(state *SessionHandlerState) bool {

	/*
		Fallback to direct is a state where the SDK has met some fatal error condition.

		When this happens, the session will go direct from that point forward.

		Here we look at flags sent up from the SDK, and send them to stackdriver metrics,
		so we can diagnose what caused any fallback to directs to happen.
	*/

	if state.Packet.FallbackToDirect && !state.Output.FellBackToDirect {

		core.Debug("fallback to direct")

		state.Output.FellBackToDirect = true

		reported := false

		if state.Packet.Flags&FallbackFlagsBadRouteToken != 0 {
			state.Metrics.FallbackToDirectBadRouteToken.Add(1)
			reported = true
		}

		if state.Packet.Flags&FallbackFlagsNoNextRouteToContinue != 0 {
			state.Metrics.FallbackToDirectNoNextRouteToContinue.Add(1)
			reported = true
		}

		if state.Packet.Flags&FallbackFlagsPreviousUpdateStillPending != 0 {
			state.Metrics.FallbackToDirectPreviousUpdateStillPending.Add(1)
			reported = true
		}

		if state.Packet.Flags&FallbackFlagsBadContinueToken != 0 {
			state.Metrics.FallbackToDirectBadContinueToken.Add(1)
			reported = true
		}

		if state.Packet.Flags&FallbackFlagsRouteExpired != 0 {
			state.Metrics.FallbackToDirectRouteExpired.Add(1)
			reported = true
		}

		if state.Packet.Flags&FallbackFlagsRouteRequestTimedOut != 0 {
			state.Metrics.FallbackToDirectRouteRequestTimedOut.Add(1)
			reported = true
		}

		if state.Packet.Flags&FallbackFlagsContinueRequestTimedOut != 0 {
			state.Metrics.FallbackToDirectContinueRequestTimedOut.Add(1)
			reported = true
		}

		if state.Packet.Flags&FallbackFlagsClientTimedOut != 0 {
			state.Metrics.FallbackToDirectClientTimedOut.Add(1)
			reported = true
		}

		if state.Packet.Flags&FallbackFlagsUpgradeResponseTimedOut != 0 {
			state.Metrics.FallbackToDirectUpgradeResponseTimedOut.Add(1)
			reported = true
		}

		if state.Packet.Flags&FallbackFlagsRouteUpdateTimedOut != 0 {
			state.Metrics.FallbackToDirectRouteUpdateTimedOut.Add(1)
			reported = true
		}

		if state.Packet.Flags&FallbackFlagsDirectPongTimedOut != 0 {
			state.Metrics.FallbackToDirectDirectPongTimedOut.Add(1)
			reported = true
		}

		if state.Packet.Flags&FallbackFlagsNextPongTimedOut != 0 {
			state.Metrics.FallbackToDirectNextPongTimedOut.Add(1)
			reported = true
		}

		if !reported {
			state.Metrics.FallbackToDirectUnknownReason.Add(1)
		}

		return true
	}

	return false
}

func SessionGetNearRelays(state *SessionHandlerState) bool {

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
	*/

	if state.Buyer.RouteShader.AnalysisOnly {
		core.Debug("analysis only, not getting near relays")
		return false
	}

	directLatency := state.Packet.DirectMinRTT

	clientLatitude := state.Output.Location.Latitude
	clientLongitude := state.Output.Location.Longitude

	serverLatitude := state.Datacenter.Location.Latitude
	serverLongitude := state.Datacenter.Location.Longitude

	state.Response.NearRelayIDs, state.Response.NearRelayAddresses = state.RouteMatrix.GetNearRelays(directLatency, clientLatitude, clientLongitude, serverLatitude, serverLongitude, core.MaxNearRelays, state.Datacenter.ID)
	if len(state.Response.NearRelayIDs) == 0 {
		core.Debug("no near relays :(")
		state.Metrics.NearRelaysLocateFailure.Add(1)
		return false
	}

	state.Response.NumNearRelays = int32(len(state.Response.NearRelayIDs))
	state.Response.HighFrequencyPings = state.Buyer.InternalConfig.HighFrequencyPings && !state.Buyer.InternalConfig.LargeCustomer
	state.Response.NearRelaysChanged = true

	return true
}

func SessionUpdateNearRelayStats(state *SessionHandlerState) bool {

	/*
		This function is called once every seconds for all slices
		in a session after slice 0 (first slice).

		It takes the ping statistics for each near relay, and collates them
		into a format suitable for route planning later on in the session
		update.

		It also runs various filters inside core.ReframeRelays, which look at
		the history of latency, jitter and packet loss across the entire session
		in order to exclude near relays with bad performance from being selected.

		This function is skipped for "Analysis Only" buyers because sessions
		will always take direct.
	*/

	routeShader := &state.Buyer.RouteShader

	if routeShader.AnalysisOnly {
		core.Debug("analysis only, not updating near relay stats")
		return false
	}

	routeState := &state.Output.RouteState

	directLatency := int32(math.Ceil(float64(state.Packet.DirectMinRTT)))
	directJitter := int32(math.Ceil(float64(state.Packet.DirectJitter)))
	directPacketLoss := int32(math.Floor(float64(state.Packet.DirectPacketLoss) + 0.5))
	nextPacketLoss := int32(math.Floor(float64(state.Packet.NextPacketLoss) + 0.5))

	destRelayIDs := state.RouteMatrix.GetDatacenterRelayIDs(state.Datacenter.ID)
	if len(destRelayIDs) == 0 {
		core.Debug("no relays in datacenter %x", state.Datacenter.ID)
		state.Metrics.NoRelaysInDatacenter.Add(1)
		return false
	}

	sliceNumber := int32(state.Packet.SliceNumber)

	state.DestRelays = make([]int32, len(destRelayIDs))

	/*
		If we are holding near relays, use the held near relay RTT as input
		instead of the near relay ping data sent up from the SDK.
	*/

	if state.Input.HoldNearRelays {
		core.Debug("using held near relay RTTs")
		for i := range state.Packet.NearRelayIDs {
			state.Packet.NearRelayRTT[i] = state.Input.HoldNearRelayRTT[i] // when set to 255, near relay is excluded from routing
			state.Packet.NearRelayJitter[i] = 0
			state.Packet.NearRelayPacketLoss[i] = 0
		}
	}

	/*
		Reframe the near relays to get them in a relay index form relative to the current route matrix.
	*/

	core.ReframeRelays(

		// input
		routeShader,
		routeState,
		state.RouteMatrix.RelayIDsToIndices,
		directLatency,
		directJitter,
		directPacketLoss,
		nextPacketLoss,
		sliceNumber,
		state.Packet.NearRelayIDs,
		state.Packet.NearRelayRTT,
		state.Packet.NearRelayJitter,
		state.Packet.NearRelayPacketLoss,
		destRelayIDs,

		// output
		state.NearRelayRTTs[:],
		state.NearRelayJitters[:],
		&state.NumDestRelays,
		state.DestRelays,
	)

	state.NumNearRelays = len(state.Packet.NearRelayIDs)

	for i := range state.Packet.NearRelayIDs {
		relayIndex, exists := state.RouteMatrix.RelayIDsToIndices[state.Packet.NearRelayIDs[i]]
		if exists {
			state.NearRelayIndices[i] = relayIndex
		} else {
			state.NearRelayIndices[i] = -1 // near relay no longer exists in route matrix
		}
	}

	SessionFilterNearRelays(state) // IMPORTANT: Reduce % of sessions that run near relay pings for large customers

	return true

}

func SessionFilterNearRelays(state *SessionHandlerState) {

	/*
		Reduce the % of sessions running near relay pings for large customers.

		We do this by only running near relay pings for the first 3 slices, and then holding
		the near relay ping results fixed for the rest of the session.
	*/

	if !state.Buyer.InternalConfig.LargeCustomer {
		return
	}

	if state.Packet.SliceNumber < 4 {
		return
	}

	// IMPORTANT: On any slice after 4, if we haven't already, grab the *processed*
	// near relay RTTs from ReframeRelays, which are set to 255 for any near relays
	// excluded because of high jitter or PL and hold them as the near relay RTTs to use from now on.

	if !state.Input.HoldNearRelays {
		core.Debug("holding near relays")
		state.Output.HoldNearRelays = true
		for i := 0; i < len(state.Packet.NearRelayIDs); i++ {
			state.Output.HoldNearRelayRTT[i] = state.NearRelayRTTs[i]
		}
	}

	// tell the SDK to stop pinging near relays

	state.Response.ExcludeNearRelays = true
	for i := 0; i < core.MaxNearRelays; i++ {
		state.Response.NearRelayExcluded[i] = true
	}
}

func SessionMakeRouteDecision(state *SessionHandlerState) {

	// todo: why would we copy such a potentially large map here? really bad idea...
	// multipathVetoMap := multipathVetoHandler.GetMapCopy(buyer.CompanyCode)
	multipathVetoMap := map[uint64]bool{}

	/*
		If we are on on network next but don't have any relays in our route, something is WRONG.
		Veto the session and go direct.
	*/

	if state.Input.RouteState.Next && state.Input.RouteNumRelays == 0 {
		core.Debug("on network next, but no route relays?")
		state.Output.RouteState.Next = false
		state.Output.RouteState.Veto = true
		state.Metrics.NextWithoutRouteRelays.Add(1)
		return
	}

	var stayOnNext bool
	var routeChanged bool
	var routeCost int32
	var routeNumRelays int32

	routeRelays := [core.MaxRelaysPerRoute]int32{}

	sliceNumber := int32(state.Packet.SliceNumber)

	if !state.Input.RouteState.Next {

		// currently going direct. should we take network next?

		if core.MakeRouteDecision_TakeNetworkNext(state.RouteMatrix.RouteEntries, state.RouteMatrix.FullRelayIndicesSet, &state.Buyer.RouteShader, &state.Output.RouteState, multipathVetoMap, &state.Buyer.InternalConfig, int32(state.Packet.DirectMinRTT), state.RealPacketLoss, state.NearRelayIndices[:], state.NearRelayRTTs[:], state.DestRelays, &routeCost, &routeNumRelays, routeRelays[:], &state.RouteDiversity, state.Debug, sliceNumber) {
			BuildNextTokens(&state.Output, state.Database, &state.Buyer, &state.Packet, routeNumRelays, routeRelays[:routeNumRelays], state.RouteMatrix.RelayIDs, state.RouterPrivateKey, &state.Response)

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

		if !state.Packet.Next {

			// the sdk aborted this session

			core.Debug("aborted")
			state.Output.RouteState.Next = false
			state.Output.RouteState.Veto = true
			state.Metrics.SDKAborted.Add(1)
			return
		}

		/*
			Reframe the current route in terms of relay indices in the current route matrix

			This is necessary because the set of relays in the route matrix change over time.
		*/

		if !core.ReframeRoute(&state.Output.RouteState, state.RouteMatrix.RelayIDsToIndices, state.Output.RouteRelayIDs[:state.Output.RouteNumRelays], &routeRelays) {
			routeRelays = [core.MaxRelaysPerRoute]int32{}
			core.Debug("one or more relays in the route no longer exist")
			state.Metrics.RouteDoesNotExist.Add(1)
		}

		stayOnNext, routeChanged = core.MakeRouteDecision_StayOnNetworkNext(state.RouteMatrix.RouteEntries, state.RouteMatrix.FullRelayIndicesSet, state.RouteMatrix.RelayNames, &state.Buyer.RouteShader, &state.Output.RouteState, &state.Buyer.InternalConfig, int32(state.Packet.DirectMinRTT), int32(state.Packet.NextRTT), state.Output.RouteCost, state.RealPacketLoss, state.Packet.NextPacketLoss, state.Output.RouteNumRelays, routeRelays, state.NearRelayIndices[:], state.NearRelayRTTs[:], state.DestRelays[:], &routeCost, &routeNumRelays, routeRelays[:], state.Debug)

		if stayOnNext {

			// stay on network next

			if routeChanged {
				core.Debug("route changed")
				state.Metrics.RouteSwitched.Add(1)
				BuildNextTokens(&state.Output, state.Database, &state.Buyer, &state.Packet, routeNumRelays, routeRelays[:routeNumRelays], state.RouteMatrix.RelayIDs, state.RouterPrivateKey, &state.Response)
			} else {
				core.Debug("route continued")
				BuildContinueTokens(&state.Output, state.Database, &state.Buyer, &state.Packet, routeNumRelays, routeRelays[:routeNumRelays], state.RouteMatrix.RelayIDs, state.RouterPrivateKey, &state.Response)
			}

		} else {

			// leave network next

			if state.Output.RouteState.NoRoute {
				core.Debug("route no longer exists")
				state.Metrics.NoRoute.Add(1)
			}

			if state.Output.RouteState.MultipathOverload {
				core.Debug("multipath overload")
				state.Metrics.MultipathOverload.Add(1)
			}

			if state.Output.RouteState.Mispredict {
				core.Debug("mispredict")
				state.Metrics.MispredictVeto.Add(1)
			}

			if state.Output.RouteState.LatencyWorse {
				core.Debug("latency worse")
				state.Metrics.LatencyWorse.Add(1)
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

	if routeCost > routing.InvalidRouteValue {
		routeCost = routing.InvalidRouteValue
	}

	state.Output.RouteCost = routeCost
	state.Output.RouteChanged = routeChanged
	state.Output.RouteNumRelays = routeNumRelays

	for i := int32(0); i < routeNumRelays; i++ {
		relayID := state.RouteMatrix.RelayIDs[routeRelays[i]]
		state.Output.RouteRelayIDs[i] = relayID
	}
}

func SessionPost(state *SessionHandlerState) {

	/*
		If the buyer doesn't exist, or the signature check failed,
		this is potentially a malicious request. Don't respond to it.
	*/

	if state.BuyerNotFound || state.SignatureCheckFailed {
		core.Debug("not responding")
		return
	}

	/*
		Build the set of near relays for the SDK to ping.

		The SDK pings these near relays and reports up the results in the next session update.

		We hold the set of near relays fixed for the session, so we only do this work on the first slice.
	*/

	if state.Packet.SliceNumber == 0 {
		SessionGetNearRelays(state)
		core.Debug("first slice always goes direct")
	}

	/*
		Since post runs at the end of every session handler, run logic
		here that must run if we are taking network next vs. direct
	*/

	if state.Response.RouteType != routing.RouteTypeDirect {
		core.Debug("session takes network next")
		state.Metrics.NextSlices.Add(1)
	} else {
		core.Debug("session goes direct")
		state.Metrics.DirectSlices.Add(1)
	}

	/*
		Decide if the session was ever on next.

		We avoid using route type to verify if a session was ever on next
		in case the route decision to take next was made on the final slice.
	*/

	if state.Packet.Next {
		state.Output.EverOnNext = true
	}

	/*
		Store the packets sent and packets lost counters in the route state,
		so we can use them to calculate real packet loss next session update.
	*/

	state.Output.PrevPacketsSentClientToServer = state.Packet.PacketsSentClientToServer
	state.Output.PrevPacketsSentServerToClient = state.Packet.PacketsSentServerToClient
	state.Output.PrevPacketsLostClientToServer = state.Packet.PacketsLostClientToServer
	state.Output.PrevPacketsLostServerToClient = state.Packet.PacketsLostServerToClient

	/*
		If the core routing logic generated a debug string, include it in the response.
	*/

	if state.Debug != nil {
		state.Response.Debug = *state.Debug
		if state.Response.Debug != "" {
			state.Response.HasDebug = true
		}
	}

	/*
		Build route relay data (for portal, billing etc...).

		This is done here to get the post route relay sellers egress price override for
		calculating total price and route relay price when building the billing entry.
	*/

	BuildPostRouteRelayData(state)

	/*
		Determine if we should write the summary slice. Should only happen
		when the session is finished.

		The end of a session occurs when the client ping times out.

		We always set the output flag to true so that it remains recorded as true on
		subsequent slices where the client ping has timed out. Instead, we check
		the input when deciding to write billing entry 2.
	*/

	if state.PostSessionHandler.featureBilling2 && state.Packet.ClientPingTimedOut {
		state.Output.WroteSummary = true
	}

	/*
		Each slice is 10 seconds long except for the first slice with a given network next route,
		which is 20 seconds long. Each time we change network next route, we burn the 10 second tail
		that we pre-bought at the start of the previous route.

		If the route changed on the final session slice, the slice duration
		should be 10 seconds, not 20 seconds, since the session ended using
		the previous route.

		Otherwise the first and summary slices will have different values for
		the envelope bandwidth, total price, etc.
	*/

	sliceDuration := uint64(billing.BillingSliceSeconds)
	if state.Input.Initial && !(state.Output.WroteSummary && state.Input.RouteChanged) {
		sliceDuration *= 2
	}

	/*
		Calculate the envelope bandwidth in bytes up and down for the duration of the previous slice.

		This is what we bill on.
	*/

	nextEnvelopeBytesUp, nextEnvelopeBytesDown := CalculateNextBytesUpAndDown(uint64(state.Buyer.RouteShader.BandwidthEnvelopeUpKbps), uint64(state.Buyer.RouteShader.BandwidthEnvelopeDownKbps), sliceDuration)

	/*
		Calculate the total price for this slice of bandwidth envelope.

		This is the sum of all relay hop prices, plus our rake, multiplied by the envelope up/down
		and the length of the session in seconds.
	*/

	totalPrice := CalculateTotalPriceNibblins(int(state.Input.RouteNumRelays), state.PostRouteRelaySellers, state.PostRouteRelayEgressPriceOverride, nextEnvelopeBytesUp, nextEnvelopeBytesDown)

	/*
		Store the cumulative sum of totalPrice, nextEnvelopeBytesUp, and nextEnvelopeBytesDown in
		the output session data. Used in the summary slice.

		If this is the summary slice, then we do NOT want to include this slice's values in the
		cumulative sum since this session is finished.

		This saves datascience some work when analyzing sessions across days.
	*/

	if !state.Output.WroteSummary && state.Packet.Next {
		state.Output.TotalPriceSum = state.Input.TotalPriceSum + uint64(totalPrice)
		state.Output.NextEnvelopeBytesUpSum = state.Input.NextEnvelopeBytesUpSum + nextEnvelopeBytesUp
		state.Output.NextEnvelopeBytesDownSum = state.Input.NextEnvelopeBytesDownSum + nextEnvelopeBytesDown
		state.Output.DurationOnNext = state.Input.DurationOnNext + billing.BillingSliceSeconds
	}

	/*
		Write the session response packet and send it back to the caller.
	*/

	if err := WriteSessionResponse(state.Writer, &state.Response, &state.Output, state.Metrics); err != nil {
		core.Debug("failed to write session update response: %s", err)
		state.Metrics.WriteResponseFailure.Add(1)
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
				level.Error(PostSessionHandler.logger).Log("err", err)
			}
		}
	*/

	/*
		Build post near relay data (for portal, billing etc...)
	*/

	BuildPostNearRelayData(state)

	/*
		Build billing 2 data and send it to the billing system via pubsub (non-realtime path)

		Check the input state to see if we wrote the summary slice since
		the output state is not set to input state if we early out in sessionPre()
		when the client ping times out.

		Doing this ensures that we only write the summary slice once since the first time the
		client ping times out, input flag will be false and the output flag will be true,
		and on the following slices, both will be true.
	*/

	if state.PostSessionHandler.featureBilling2 && !state.Input.WroteSummary {
		billingEntry2 := BuildBillingEntry2(state, sliceDuration, nextEnvelopeBytesUp, nextEnvelopeBytesDown, totalPrice)

		state.PostSessionHandler.SendBillingEntry2(billingEntry2)

		/*
			Send the billing entry to the vanity metrics system (real-time path)
			except for the summary slice.
		*/

		if state.PostSessionHandler.useVanityMetrics && !state.Output.WroteSummary {
			state.PostSessionHandler.SendVanityMetric(billingEntry2)
		}
	}

	/*
		The client times out at the end of each session, and holds on for 60 seconds.
		These slices at the end have no useful information for the portal, so we drop
		them here. Billing2 needs to know if the client times out to write the summary
		portion of the billing entry 2, but Billing1 does not.
	*/

	if state.Packet.ClientPingTimedOut {
		return
	}

	/*
		Send data to the portal (real-time path)
	*/

	portalData := BuildPortalData(state)

	if portalData.Meta.NextRTT != 0 || portalData.Meta.DirectRTT != 0 {
		state.PostSessionHandler.SendPortalData(portalData)
	}
}

func BuildPostRouteRelayData(state *SessionHandlerState) {

	/*
		Build information about the relays involved in the current route.

		This data is sent to the portal, billing and the vanity metrics system.
	*/

	for i := int32(0); i < state.Input.RouteNumRelays; i++ {
		relay, ok := state.Database.RelayMap[state.Input.RouteRelayIDs[i]]
		if ok {
			state.PostRouteRelayNames[i] = relay.Name
			state.PostRouteRelaySellers[i] = relay.Seller
			state.PostRouteRelayEgressPriceOverride[i] = relay.EgressPriceOverride
		}
	}
}

func BuildPostNearRelayData(state *SessionHandlerState) {

	state.PostNearRelayCount = int(state.Packet.NumNearRelays)

	for i := 0; i < state.PostNearRelayCount; i++ {

		/*
			The set of near relays is held fixed at the start of a session.
			Therefore it is possible that a near relay may no longer exist.
		*/

		relayID := state.Packet.NearRelayIDs[i]
		relayIndex, ok := state.RouteMatrix.RelayIDsToIndices[relayID]
		if !ok {
			continue
		}

		/*
			Fill in information for near relays needed by billing and the portal.

			We grab this data from the session update packet, which corresponds to the previous slice (input).

			This makes sure all values for a slice in billing and the portal line up temporally.
		*/

		state.PostNearRelayIDs[i] = relayID
		state.PostNearRelayNames[i] = state.RouteMatrix.RelayNames[relayIndex]
		state.PostNearRelayAddresses[i] = state.RouteMatrix.RelayAddresses[relayIndex]
		state.PostNearRelayRTT[i] = float32(state.Packet.NearRelayRTT[i])
		state.PostNearRelayJitter[i] = float32(state.Packet.NearRelayJitter[i])
		state.PostNearRelayPacketLoss[i] = float32(state.Packet.NearRelayPacketLoss[i])
	}
}

func BuildBillingEntry2(state *SessionHandlerState, sliceDuration uint64, nextEnvelopeBytesUp uint64, nextEnvelopeBytesDown uint64, totalPrice routing.Nibblin) *billing.BillingEntry2 {
	/*
		Calculate the actual amounts of bytes sent up and down along the network next route
		for the duration of the previous slice (just being reported up from the SDK).

		This is *not* what we bill on.
	*/

	nextBytesUp, nextBytesDown := CalculateNextBytesUpAndDown(uint64(state.Packet.NextKbpsUp), uint64(state.Packet.NextKbpsDown), sliceDuration)

	/*
		Calculate the per-relay hop price that sums up to the total price, minus our rake.
	*/

	routeRelayPrices := CalculateRouteRelaysPrice(int(state.Input.RouteNumRelays), state.PostRouteRelaySellers, state.PostRouteRelayEgressPriceOverride, nextEnvelopeBytesUp, nextEnvelopeBytesDown)

	// todo: not really sure why we transform it like this? seems wasteful
	nextRelayPrice := [core.MaxRelaysPerRoute]uint64{}
	for i := 0; i < core.MaxRelaysPerRoute; i++ {
		nextRelayPrice[i] = uint64(routeRelayPrices[i])
	}

	// todo: not really sure why we need to do this...
	var routeCost int32 = state.Input.RouteCost
	if state.Input.RouteCost == math.MaxInt32 {
		routeCost = 0
	}

	/*
		Save the first hop RTT from the client to the first relay in the route.

		This is useful for analysis and saves data science some work.
	*/

	var nearRelayRTT int32
	if state.Input.RouteNumRelays > 0 {
		for i, nearRelayID := range state.PostNearRelayIDs {
			if nearRelayID == state.Input.RouteRelayIDs[0] {
				nearRelayRTT = int32(state.PostNearRelayRTT[i])
				break
			}
		}
	}

	/*
		If the debug string is set to something by the core routing system, put it in the billing entry.
	*/

	debugString := ""
	if state.Debug != nil {
		debugString = *state.Debug
	}

	/*
		Separate the integer and fractional portions of real packet loss to
		allow for more efficient bitpacking while maintaining precision.
	*/

	RealPacketLoss, RealPacketLoss_Frac := math.Modf(float64(state.RealPacketLoss))
	RealPacketLoss_Frac = math.Round(RealPacketLoss_Frac * 255.0)

	/*
		Recast near relay RTT, Jitter, and Packet Loss to int32.
		We do this here since the portal data requires float level precision.
	*/

	var NearRelayRTTs [core.MaxNearRelays]int32
	var NearRelayJitters [core.MaxNearRelays]int32
	var nearRelayPacketLosses [core.MaxNearRelays]int32
	for i := 0; i < state.PostNearRelayCount; i++ {
		NearRelayRTTs[i] = int32(state.PostNearRelayRTT[i])
		NearRelayJitters[i] = int32(state.PostNearRelayJitter[i])
		nearRelayPacketLosses[i] = int32(state.PostNearRelayPacketLoss[i])
	}

	/*
		Calculate the session duration in seconds for the summary slice.

		Slice numbers start at 0, so the length of a session is the
		summary slice's slice number * 10 seconds.
	*/
	var sessionDuration uint32
	if state.Output.WroteSummary && state.Packet.SliceNumber != 0 {
		sessionDuration = state.Packet.SliceNumber * billing.BillingSliceSeconds
	}

	/*
		Calculate the starting timestamp of the session to include in the summary slice.
	*/
	var startTime time.Time
	if state.Output.WroteSummary {
		secondsToSub := int(sessionDuration)
		startTime = time.Now().Add(time.Duration(-secondsToSub) * time.Second)
	}

	/*
		Create the billing entry 2 and return it to the caller.
	*/

	billingEntry2 := billing.BillingEntry2{
		Version:                         uint32(billing.BillingEntryVersion2),
		Timestamp:                       uint32(time.Now().Unix()),
		SessionID:                       state.Packet.SessionID,
		SliceNumber:                     state.Packet.SliceNumber,
		DirectMinRTT:                    int32(state.Packet.DirectMinRTT),
		DirectMaxRTT:                    int32(state.Packet.DirectMaxRTT),
		DirectPrimeRTT:                  int32(state.Packet.DirectPrimeRTT),
		DirectJitter:                    int32(state.Packet.DirectJitter),
		DirectPacketLoss:                int32(state.Packet.DirectPacketLoss),
		RealPacketLoss:                  int32(RealPacketLoss),
		RealPacketLoss_Frac:             uint32(RealPacketLoss_Frac),
		RealJitter:                      uint32(state.RealJitter),
		Next:                            state.Packet.Next,
		Flagged:                         state.Packet.Reported,
		Summary:                         state.Output.WroteSummary,
		UseDebug:                        state.Buyer.Debug,
		Debug:                           debugString,
		RouteDiversity:                  int32(state.RouteDiversity),
		UserFlags:                       state.Packet.UserFlags,
		DatacenterID:                    state.Packet.DatacenterID,
		BuyerID:                         state.Packet.BuyerID,
		UserHash:                        state.Packet.UserHash,
		EnvelopeBytesDown:               nextEnvelopeBytesDown,
		EnvelopeBytesUp:                 nextEnvelopeBytesUp,
		Latitude:                        float32(state.Input.Location.Latitude),
		Longitude:                       float32(state.Input.Location.Longitude),
		ClientAddress:                   state.Packet.ClientAddress.String(),
		ServerAddress:                   state.Packet.ServerAddress.String(),
		ISP:                             state.Input.Location.ISP,
		ConnectionType:                  int32(state.Packet.ConnectionType),
		PlatformType:                    int32(state.Packet.PlatformType),
		SDKVersion:                      state.Packet.Version.String(),
		NumTags:                         int32(state.Packet.NumTags),
		Tags:                            state.Packet.Tags,
		ABTest:                          state.Input.RouteState.ABTest,
		Pro:                             state.Buyer.RouteShader.ProMode && !state.Input.RouteState.MultipathRestricted,
		ClientToServerPacketsSent:       state.Packet.PacketsSentClientToServer,
		ServerToClientPacketsSent:       state.Packet.PacketsSentServerToClient,
		ClientToServerPacketsLost:       state.Packet.PacketsLostClientToServer,
		ServerToClientPacketsLost:       state.Packet.PacketsLostServerToClient,
		ClientToServerPacketsOutOfOrder: state.Packet.PacketsOutOfOrderClientToServer,
		ServerToClientPacketsOutOfOrder: state.Packet.PacketsOutOfOrderServerToClient,
		NumNearRelays:                   int32(state.PostNearRelayCount),
		NearRelayIDs:                    state.PostNearRelayIDs,
		NearRelayRTTs:                   NearRelayRTTs,
		NearRelayJitters:                NearRelayJitters,
		NearRelayPacketLosses:           nearRelayPacketLosses,
		EverOnNext:                      state.Input.EverOnNext,
		SessionDuration:                 sessionDuration,
		TotalPriceSum:                   state.Input.TotalPriceSum,
		EnvelopeBytesUpSum:              state.Input.NextEnvelopeBytesUpSum,
		EnvelopeBytesDownSum:            state.Input.NextEnvelopeBytesDownSum,
		DurationOnNext:                  state.Input.DurationOnNext,
		StartTimestamp:                  uint32(startTime.Unix()),
		NextRTT:                         int32(state.Packet.NextRTT),
		NextJitter:                      int32(state.Packet.NextJitter),
		NextPacketLoss:                  int32(state.Packet.NextPacketLoss),
		PredictedNextRTT:                routeCost,
		NearRelayRTT:                    nearRelayRTT,
		NumNextRelays:                   int32(state.Input.RouteNumRelays),
		NextRelays:                      state.Input.RouteRelayIDs,
		NextRelayPrice:                  nextRelayPrice,
		TotalPrice:                      uint64(totalPrice),
		Uncommitted:                     !state.Packet.Committed,
		Multipath:                       state.Input.RouteState.Multipath,
		RTTReduction:                    state.Input.RouteState.ReduceLatency,
		PacketLossReduction:             state.Input.RouteState.ReducePacketLoss,
		RouteChanged:                    state.Input.RouteChanged,
		NextBytesUp:                     nextBytesUp,
		NextBytesDown:                   nextBytesDown,
		FallbackToDirect:                state.Packet.FallbackToDirect,
		MultipathVetoed:                 state.Input.RouteState.MultipathOverload,
		Mispredicted:                    state.Input.RouteState.Mispredict,
		Vetoed:                          state.Input.RouteState.Veto,
		LatencyWorse:                    state.Input.RouteState.LatencyWorse,
		NoRoute:                         state.Input.RouteState.NoRoute,
		NextLatencyTooHigh:              state.Input.RouteState.NextLatencyTooHigh,
		CommitVeto:                      state.Input.RouteState.CommitVeto,
		UnknownDatacenter:               state.UnknownDatacenter,
		DatacenterNotEnabled:            state.DatacenterNotEnabled,
		BuyerNotLive:                    state.BuyerNotLive,
		StaleRouteMatrix:                state.StaleRouteMatrix,
		TryBeforeYouBuy:                 !state.Input.RouteState.Committed,
	}

	// Clamp any values to ensure the entry is serialized properly
	billingEntry2.ClampEntry()

	return &billingEntry2
}

func BuildPortalData(state *SessionHandlerState) *SessionPortalData {

	/*
		Build the relay hops for the portal
	*/

	// todo: we should try to avoid allocations
	hops := make([]RelayHop, state.Input.RouteNumRelays)
	for i := int32(0); i < state.Input.RouteNumRelays; i++ {
		hops[i] = RelayHop{
			Version: RelayHopVersion,
			ID:      state.Input.RouteRelayIDs[i],
			Name:    state.PostRouteRelayNames[i],
		}
	}

	/*
		Build the near relay data for the portal
	*/

	// todo: we should try to avoid allocations
	nearRelayPortalData := make([]NearRelayPortalData, state.PostNearRelayCount)
	for i := range nearRelayPortalData {
		nearRelayPortalData[i] = NearRelayPortalData{
			Version: NearRelayPortalDataVersion,
			ID:      state.PostNearRelayIDs[i],
			Name:    state.PostNearRelayNames[i],
			ClientStats: routing.Stats{
				RTT:        float64(state.PostNearRelayRTT[i]),
				Jitter:     float64(state.PostNearRelayJitter[i]),
				PacketLoss: float64(state.PostNearRelayPacketLoss[i]),
			},
		}
	}

	/*
		Calculate the delta between network next and direct.

		Clamp the delta RTT above 0. This is used for the top sessions page.
	*/

	var deltaRTT float32
	if state.Packet.Next && state.Packet.NextRTT != 0 && state.Packet.DirectMinRTT >= state.Packet.NextRTT {
		deltaRTT = state.Packet.DirectMinRTT - state.Packet.NextRTT
	}

	/*
		Predicted RTT is the round trip time that we predict, even if we don't
		take network next. It's a conservative prodiction.
	*/

	predictedRTT := float64(state.Input.RouteCost)
	if state.Input.RouteCost >= routing.InvalidRouteValue {
		predictedRTT = 0
	}

	/*
		Build the portal data and return it to the caller.
	*/

	portalData := SessionPortalData{
		Version: SessionPortalDataVersion,
		Meta: SessionMeta{
			Version:         SessionMetaVersion,
			ID:              state.Packet.SessionID,
			UserHash:        state.Packet.UserHash,
			DatacenterName:  state.Datacenter.Name,
			DatacenterAlias: state.Datacenter.AliasName,
			OnNetworkNext:   state.Packet.Next,
			NextRTT:         float64(state.Packet.NextRTT),
			DirectRTT:       float64(state.Packet.DirectMinRTT),
			DeltaRTT:        float64(deltaRTT),
			Location:        state.Input.Location,
			ClientAddr:      state.Packet.ClientAddress.String(),
			ServerAddr:      state.Packet.ServerAddress.String(),
			Hops:            hops,
			SDK:             state.Packet.Version.String(),
			Connection:      uint8(state.Packet.ConnectionType),
			NearbyRelays:    nearRelayPortalData,
			Platform:        uint8(state.Packet.PlatformType),
			BuyerID:         state.Packet.BuyerID,
		},
		Slice: SessionSlice{
			Version:   SessionSliceVersion,
			Timestamp: time.Now(),
			Next: routing.Stats{
				RTT:        float64(state.Packet.NextRTT),
				Jitter:     float64(state.Packet.NextJitter),
				PacketLoss: float64(state.Packet.NextPacketLoss),
			},
			Direct: routing.Stats{
				RTT:        float64(state.Packet.DirectMinRTT),
				Jitter:     float64(state.Packet.DirectJitter),
				PacketLoss: float64(state.Packet.DirectPacketLoss),
			},
			Predicted: routing.Stats{
				RTT: predictedRTT,
			},
			ClientToServerStats: routing.Stats{
				Jitter:     float64(state.Packet.JitterClientToServer),
				PacketLoss: float64(state.PostRealPacketLossClientToServer),
			},
			ServerToClientStats: routing.Stats{
				Jitter:     float64(state.Packet.JitterServerToClient),
				PacketLoss: float64(state.PostRealPacketLossServerToClient),
			},
			RouteDiversity: uint32(state.RouteDiversity),
			Envelope: routing.Envelope{
				Up:   int64(state.Packet.NextKbpsUp),
				Down: int64(state.Packet.NextKbpsDown),
			},
			IsMultiPath:       state.Input.RouteState.Multipath,
			IsTryBeforeYouBuy: !state.Input.RouteState.Committed,
			OnNetworkNext:     state.Packet.Next,
		},
		Point: SessionMapPoint{
			Version:   SessionMapPointVersion,
			Latitude:  float64(state.Input.Location.Latitude),
			Longitude: float64(state.Input.Location.Longitude),
			SessionID: state.Input.SessionID,
		},
		LargeCustomer: state.Buyer.InternalConfig.LargeCustomer,
		EverOnNext:    state.Input.EverOnNext,
	}

	return &portalData
}

// ------------------------------------------------------------------

func WriteSessionResponse(w io.Writer, response *SessionResponsePacket, sessionData *SessionData, metrics *metrics.SessionUpdateMetrics) error {
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
	getIPLocator func() *routing.MaxmindDB,
	getRouteMatrix func() *routing.RouteMatrix,
	multipathVetoHandler storage.MultipathVetoHandler,
	getDatabase func() *routing.DatabaseBinWrapper,
	routerPrivateKey [crypto.KeySize]byte,
	PostSessionHandler *PostSessionHandler,
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

		if err := UnmarshalPacket(&state.Packet, incoming.Data); err != nil {
			core.Debug("could not read session update packet:\n\n%v\n", err)
			metrics.ReadPacketFailure.Add(1)
			return
		}

		// log stuff we want to see with each session update (debug only)

		core.Debug("buyer id is %x", state.Packet.BuyerID)
		core.Debug("datacenter id is %x", state.Packet.DatacenterID)
		core.Debug("session id is %x", state.Packet.SessionID)
		core.Debug("slice number is %d", state.Packet.SliceNumber)
		core.Debug("retry number is %d", state.Packet.RetryNumber)

		/*
			Build session handler state. Putting everything in a struct makes calling subroutines much easier.
		*/

		state.Writer = w
		state.Metrics = metrics
		state.Database = getDatabase()
		state.Datacenter = routing.UnknownDatacenter
		state.PacketData = incoming.Data
		state.IpLocator = getIPLocator()
		state.RouteMatrix = getRouteMatrix()
		state.StaleDuration = staleDuration
		state.RouterPrivateKey = routerPrivateKey
		state.Response = SessionResponsePacket{
			Version:     state.Packet.Version,
			SessionID:   state.Packet.SessionID,
			SliceNumber: state.Packet.SliceNumber,
			RouteType:   routing.RouteTypeDirect,
		}
		state.PostSessionHandler = PostSessionHandler

		/*
			Session post *always* runs at the end of this function

			It writes and sends the response packet back to the sender,
			and sends session data to billing, vanity metrics and the portal.
		*/

		defer SessionPost(&state)

		/*
			Call session pre function

			This function checks for early out conditions and does some setup of the handler state.

			If it returns true, one of the early out conditions has been met, so we return early.
		*/

		if SessionPre(&state) {
			return
		}

		/*
			Update the session

			Do setup on slice 0, then for subsequent slices transform state.Input -> state.Output

			state.Output is sent down to the SDK in the session response packet, and next slice
			it is sent back up to us in the subsequent session update packet for this session.

			This is how we make this handler stateless. Without this, we need to store per-session
			data somewhere and this is extremely difficult at scale, given the real-time nature of
			this handler.
		*/

		if state.Packet.SliceNumber == 0 {
			SessionUpdateNewSession(&state)
		} else {
			SessionUpdateExistingSession(&state)
		}

		/*
			Handle fallback to direct.

			Fallback to direct is a condition where the SDK indicates that it has seen
			some fatal error, like not getting a session response from the backend,
			and has decided to go direct for the rest of the session.

			When this happens, we early out to save processing time.
		*/

		if SessionHandleFallbackToDirect(&state) {
			return
		}

		/*
			Process near relay ping statistics after the first slice.

			We use near relay latency, jitter and packet loss for route planning.
		*/

		SessionUpdateNearRelayStats(&state)

		/*
			Decide whether we should take network next or not.
		*/

		SessionMakeRouteDecision(&state)

		core.Debug("session updated successfully")
	}
}

// ----------------------------------------------------------------------------
