package handlers

// #cgo pkg-config: libsodium
// #include <sodium.h>
import "C"

import (
	"net"
	"time"
	"math"

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/database"
	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/modules/messages"
	"github.com/networknext/backend/modules/packets"
)

const (
	SDK5_HandlerEvent_PacketTooSmall             = 0
	SDK5_HandlerEvent_UnsupportedPacketType      = 1
	SDK5_HandlerEvent_BasicPacketFilterFailed    = 2
	SDK5_HandlerEvent_AdvancedPacketFilterFailed = 3
	SDK5_HandlerEvent_NoRouteMatrix              = 4
	SDK5_HandlerEvent_NoDatabase                 = 5
	SDK5_HandlerEvent_UnknownBuyer               = 6
	SDK5_HandlerEvent_SignatureCheckFailed       = 7
	SDK5_HandlerEvent_BuyerNotLive               = 8
	SDK5_HandlerEvent_SDKTooOld                  = 9
	SDK5_HandlerEvent_UnknownDatacenter          = 10

	SDK5_HandlerEvent_CouldNotReadServerInitRequestPacket    = 11
	SDK5_HandlerEvent_CouldNotReadServerUpdateRequestPacket  = 12
	SDK5_HandlerEvent_CouldNotReadSessionUpdateRequestPacket = 13
	SDK5_HandlerEvent_CouldNotReadMatchDataRequestPacket     = 14

	SDK5_HandlerEvent_ProcessServerInitRequestPacket    = 15
	SDK5_HandlerEvent_ProcessServerUpdateRequestPacket  = 16
	SDK5_HandlerEvent_ProcessSessionUpdateRequestPacket = 17
	SDK5_HandlerEvent_ProcessMatchDataRequestPacket     = 18

	SDK5_HandlerEvent_SentServerInitResponsePacket    = 19
	SDK5_HandlerEvent_SentServerUpdateResponsePacket  = 20
	SDK5_HandlerEvent_SentSessionUpdateResponsePacket = 21
	SDK5_HandlerEvent_SentMatchDataResponsePacket     = 22

	SDK5_HandlerEvent_SentServerInitMessage    = 23
	SDK5_HandlerEvent_SentServerUpdateMessage  = 24
	SDK5_HandlerEvent_SentSessionUpdateMessage = 25
	SDK5_HandlerEvent_SentMatchDataMessage     = 26

	SDK5_HandlerEvent_NumEvents = 27
)

type SDK5_Handler struct {
	Database             *database.Database
	RouteMatrix          *common.RouteMatrix
	MaxPacketSize        int
	ServerBackendAddress net.UDPAddr
	PrivateKey           []byte
	GetMagicValues       func() ([]byte, []byte, []byte)
	Events               [SDK5_HandlerEvent_NumEvents]bool

	ServerInitMessageChannel    chan<- *messages.ServerInitMessage
	ServerUpdateMessageChannel  chan<- *messages.ServerUpdateMessage
	SessionUpdateMessageChannel chan<- *messages.SessionUpdateMessage
	MatchDataMessageChannel     chan<- *messages.MatchDataMessage
}

func SDK5_PacketHandler(handler *SDK5_Handler, conn *net.UDPConn, from *net.UDPAddr, packetData []byte) {

	// ignore packets that are too small

	if len(packetData) < packets.SDK5_MinPacketBytes {
		core.Debug("packet is too small")
		handler.Events[SDK5_HandlerEvent_PacketTooSmall] = true
		return
	}

	// ignore packet types we don't support

	packetType := packetData[0]

	if packetType != packets.SDK5_SERVER_INIT_REQUEST_PACKET && packetType != packets.SDK5_SERVER_UPDATE_REQUEST_PACKET && packetType != packets.SDK5_SESSION_UPDATE_REQUEST_PACKET && packetType != packets.SDK5_MATCH_DATA_REQUEST_PACKET {
		core.Debug("unsupported packet type %d", packetType)
		handler.Events[SDK5_HandlerEvent_UnsupportedPacketType] = true
		return
	}

	// make sure the basic packet filter passes

	if !core.BasicPacketFilter(packetData[:], len(packetData)) {
		core.Debug("basic packet filter failed for %d byte packet from %s", len(packetData), from.String())
		handler.Events[SDK5_HandlerEvent_BasicPacketFilterFailed] = true
		return
	}

	// make sure the advanced packet filter passes

	to := &handler.ServerBackendAddress

	var emptyMagic [8]byte

	var fromAddressBuffer [32]byte
	var toAddressBuffer [32]byte

	fromAddressData, fromAddressPort := core.GetAddressData(from, fromAddressBuffer[:])
	toAddressData, toAddressPort := core.GetAddressData(to, toAddressBuffer[:])

	if !core.AdvancedPacketFilter(packetData, emptyMagic[:], fromAddressData, fromAddressPort, toAddressData, toAddressPort, len(packetData)) {
		core.Debug("advanced packet filter failed for %d byte packet from %s to %s", len(packetData), from.String(), to.String())
		handler.Events[SDK5_HandlerEvent_AdvancedPacketFilterFailed] = true
		return
	}

	// we can't process any packets without these

	if handler.RouteMatrix == nil {
		core.Debug("ignoring packet because we don't have a route matrix")
		handler.Events[SDK5_HandlerEvent_NoRouteMatrix] = true
		return
	}

	if handler.Database == nil {
		core.Debug("ignoring packet because we don't have a database")
		handler.Events[SDK5_HandlerEvent_NoDatabase] = true
		return
	}

	// check packet signature

	var buyerId uint64
	index := 16 + 3
	encoding.ReadUint64(packetData, &index, &buyerId)

	buyer, ok := handler.Database.BuyerMap[buyerId]
	if !ok {
		core.Error("unknown buyer id: %016x", buyerId)
		handler.Events[SDK5_HandlerEvent_UnknownBuyer] = true
		return
	}

	publicKey := buyer.PublicKey

	if !packets.SDK5_CheckPacketSignature(packetData, publicKey) {
		core.Debug("packet signature check failed")
		handler.Events[SDK5_HandlerEvent_SignatureCheckFailed] = true
		return
	}

	// process the packet according to type

	packetData = packetData[16:]

	switch packetType {

	case packets.SDK5_SERVER_INIT_REQUEST_PACKET:
		packet := packets.SDK5_ServerInitRequestPacket{}
		if err := packets.ReadPacket(packetData, &packet); err != nil {
			core.Error("could not read server init request packet from %s", from.String())
			handler.Events[SDK5_HandlerEvent_CouldNotReadServerInitRequestPacket] = true
			return
		}
		SDK5_ProcessServerInitRequestPacket(handler, conn, from, &packet)
		break

	case packets.SDK5_SERVER_UPDATE_REQUEST_PACKET:
		packet := packets.SDK5_ServerUpdateRequestPacket{}
		if err := packets.ReadPacket(packetData, &packet); err != nil {
			core.Error("could not read server update request packet from %s", from.String())
			handler.Events[SDK5_HandlerEvent_CouldNotReadServerUpdateRequestPacket] = true
			return
		}
		SDK5_ProcessServerUpdateRequestPacket(handler, conn, from, &packet)
		break

	case packets.SDK5_SESSION_UPDATE_REQUEST_PACKET:
		packet := packets.SDK5_SessionUpdateRequestPacket{}
		if err := packets.ReadPacket(packetData, &packet); err != nil {
			core.Error("could not read session update request packet from %s", from.String())
			handler.Events[SDK5_HandlerEvent_CouldNotReadSessionUpdateRequestPacket] = true
			return
		}
		SDK5_ProcessSessionUpdateRequestPacket(handler, conn, from, &packet)
		break

	case packets.SDK5_MATCH_DATA_REQUEST_PACKET:
		packet := packets.SDK5_MatchDataRequestPacket{}
		if err := packets.ReadPacket(packetData, &packet); err != nil {
			core.Error("could not read match data request packet from %s", from.String())
			handler.Events[SDK5_HandlerEvent_CouldNotReadMatchDataRequestPacket] = true
			return
		}
		SDK5_ProcessMatchDataRequestPacket(handler, conn, from, &packet)
		break

	default:
		panic("unknown packet type")
	}
}

func SDK5_SendResponsePacket[P packets.Packet](handler *SDK5_Handler, conn *net.UDPConn, to *net.UDPAddr, packetType int, packet P) {

	packetData, err := packets.SDK5_WritePacket(packet, packetType, handler.MaxPacketSize, &handler.ServerBackendAddress, to, handler.PrivateKey)
	if err != nil {
		core.Error("failed to write response packet: %v", err)
		return
	}

	if _, err := conn.WriteToUDP(packetData, to); err != nil {
		core.Error("failed to send response packet: %v", err)
		return
	}

	switch packetType {

	case packets.SDK5_SERVER_INIT_RESPONSE_PACKET:
		handler.Events[SDK5_HandlerEvent_SentServerInitResponsePacket] = true
		break

	case packets.SDK5_SERVER_UPDATE_RESPONSE_PACKET:
		handler.Events[SDK5_HandlerEvent_SentServerUpdateResponsePacket] = true
		break

	case packets.SDK5_SESSION_UPDATE_RESPONSE_PACKET:
		handler.Events[SDK5_HandlerEvent_SentSessionUpdateResponsePacket] = true
		break

	case packets.SDK5_MATCH_DATA_RESPONSE_PACKET:
		handler.Events[SDK5_HandlerEvent_SentMatchDataResponsePacket] = true
		break

	default:
		panic("unknown response packet type")
	}
}

func SDK5_ProcessServerInitRequestPacket(handler *SDK5_Handler, conn *net.UDPConn, from *net.UDPAddr, requestPacket *packets.SDK5_ServerInitRequestPacket) {

	handler.Events[SDK5_HandlerEvent_ProcessServerInitRequestPacket] = true

	core.Debug("---------------------------------------------------------------------------")
	core.Debug("received server init request packet from %s", from.String())
	core.Debug("version: %d.%d.%d", requestPacket.Version.Major, requestPacket.Version.Minor, requestPacket.Version.Patch)
	core.Debug("buyer id: %016x", requestPacket.BuyerId)
	core.Debug("request id: %016x", requestPacket.RequestId)
	core.Debug("datacenter: \"%s\" [%016x]", requestPacket.DatacenterName, requestPacket.DatacenterId)
	core.Debug("---------------------------------------------------------------------------")

	upcomingMagic, currentMagic, previousMagic := handler.GetMagicValues()

	responsePacket := &packets.SDK5_ServerInitResponsePacket{}
	responsePacket.RequestId = requestPacket.RequestId
	responsePacket.Response = packets.SDK5_ServerInitResponseOK
	copy(responsePacket.UpcomingMagic[:], upcomingMagic[:])
	copy(responsePacket.CurrentMagic[:], currentMagic[:])
	copy(responsePacket.PreviousMagic[:], previousMagic[:])

	buyer, exists := handler.Database.BuyerMap[requestPacket.BuyerId]
	if !exists {
		core.Debug("unknown buyer: %016x", requestPacket.BuyerId)
		responsePacket.Response = packets.SDK5_ServerInitResponseUnknownBuyer
		handler.Events[SDK5_HandlerEvent_UnknownBuyer] = true
	}

	if !buyer.Live {
		core.Debug("buyer not live: %016x", requestPacket.BuyerId)
		responsePacket.Response = packets.SDK5_ServerInitResponseBuyerNotActive
		handler.Events[SDK5_HandlerEvent_BuyerNotLive] = true
	}

	if !requestPacket.Version.AtLeast(packets.SDKVersion{5, 0, 0}) {
		core.Debug("sdk version is too old: %s", requestPacket.Version.String())
		responsePacket.Response = packets.SDK5_ServerInitResponseOldSDKVersion
		handler.Events[SDK5_HandlerEvent_SDKTooOld] = true
	}

	_, exists = handler.Database.DatacenterMap[requestPacket.DatacenterId]
	if !exists {
		// IMPORTANT: Let the server init succeed even if the datacenter is unknown!
		core.Debug("unknown datacenter '%s' [%016x]", requestPacket.DatacenterName, requestPacket.DatacenterId)
		handler.Events[SDK5_HandlerEvent_UnknownDatacenter] = true
	}

	SDK5_SendResponsePacket(handler, conn, from, packets.SDK5_SERVER_INIT_RESPONSE_PACKET, responsePacket)

	if handler.ServerInitMessageChannel != nil {

		message := messages.ServerInitMessage{}

		message.SDKVersion_Major = byte(requestPacket.Version.Major)
		message.SDKVersion_Minor = byte(requestPacket.Version.Minor)
		message.SDKVersion_Patch = byte(requestPacket.Version.Patch)
		message.BuyerId = requestPacket.BuyerId
		message.DatacenterId = requestPacket.DatacenterId
		message.DatacenterName = requestPacket.DatacenterName

		handler.ServerInitMessageChannel <- &message

		handler.Events[SDK5_HandlerEvent_SentServerInitMessage] = true
	}
}

func SDK5_ProcessServerUpdateRequestPacket(handler *SDK5_Handler, conn *net.UDPConn, from *net.UDPAddr, requestPacket *packets.SDK5_ServerUpdateRequestPacket) {

	handler.Events[SDK5_HandlerEvent_ProcessServerUpdateRequestPacket] = true

	core.Debug("---------------------------------------------------------------------------")
	core.Debug("received server update request packet from %s", from.String())
	core.Debug("version: %d.%d.%d", requestPacket.Version.Major, requestPacket.Version.Minor, requestPacket.Version.Patch)
	core.Debug("buyer id: %016x", requestPacket.BuyerId)
	core.Debug("request id: %016x", requestPacket.RequestId)
	core.Debug("datacenter id: %016x", requestPacket.DatacenterId)
	core.Debug("---------------------------------------------------------------------------")

	defer func() {

		if handler.ServerUpdateMessageChannel != nil {

			message := messages.ServerUpdateMessage{}

			message.SDKVersion_Major = byte(requestPacket.Version.Major)
			message.SDKVersion_Minor = byte(requestPacket.Version.Minor)
			message.SDKVersion_Patch = byte(requestPacket.Version.Patch)
			message.BuyerId = requestPacket.BuyerId
			message.DatacenterId = requestPacket.DatacenterId

			handler.ServerUpdateMessageChannel <- &message

			handler.Events[SDK5_HandlerEvent_SentServerUpdateMessage] = true
		}
	}()

	buyer, exists := handler.Database.BuyerMap[requestPacket.BuyerId]
	if !exists {
		core.Debug("unknown buyer: %016x", requestPacket.BuyerId)
		handler.Events[SDK5_HandlerEvent_UnknownBuyer] = true
		return
	}

	if !buyer.Live {
		core.Debug("buyer not live: %016x", requestPacket.BuyerId)
		handler.Events[SDK5_HandlerEvent_BuyerNotLive] = true
		return
	}

	if !requestPacket.Version.AtLeast(packets.SDKVersion{5, 0, 0}) {
		core.Debug("sdk version is too old: %s", requestPacket.Version.String())
		handler.Events[SDK5_HandlerEvent_SDKTooOld] = true
		return
	}

	upcomingMagic, currentMagic, previousMagic := handler.GetMagicValues()

	responsePacket := &packets.SDK5_ServerUpdateResponsePacket{}
	responsePacket.RequestId = requestPacket.RequestId
	copy(responsePacket.UpcomingMagic[:], upcomingMagic[:])
	copy(responsePacket.CurrentMagic[:], currentMagic[:])
	copy(responsePacket.PreviousMagic[:], previousMagic[:])

	_, exists = handler.Database.DatacenterMap[requestPacket.DatacenterId]
	if !exists {
		// IMPORTANT: Let the server update succeed, even if the datacenter is unknown
		core.Debug("unknown datacenter %016x", requestPacket.DatacenterId)
		handler.Events[SDK5_HandlerEvent_UnknownDatacenter] = true
	}

	SDK5_SendResponsePacket(handler, conn, from, packets.SDK5_SERVER_UPDATE_RESPONSE_PACKET, responsePacket)
}

func SDK5_ProcessMatchDataRequestPacket(handler *SDK5_Handler, conn *net.UDPConn, from *net.UDPAddr, requestPacket *packets.SDK5_MatchDataRequestPacket) {

	handler.Events[SDK5_HandlerEvent_ProcessMatchDataRequestPacket] = true

	core.Debug("---------------------------------------------------------------------------")
	core.Debug("received match data request packet from %s", from.String())
	core.Debug("---------------------------------------------------------------------------")

	buyer, exists := handler.Database.BuyerMap[requestPacket.BuyerId]
	if !exists {
		core.Debug("unknown buyer: %016x", requestPacket.BuyerId)
		handler.Events[SDK5_HandlerEvent_UnknownBuyer] = true
		return
	}

	if !buyer.Live {
		core.Debug("buyer not live: %016x", requestPacket.BuyerId)
		handler.Events[SDK5_HandlerEvent_BuyerNotLive] = true
		return
	}

	if !requestPacket.Version.AtLeast(packets.SDKVersion{5, 0, 0}) {
		core.Debug("sdk version is too old: %s", requestPacket.Version.String())
		handler.Events[SDK5_HandlerEvent_SDKTooOld] = true
		return
	}

	responsePacket := &packets.SDK5_MatchDataResponsePacket{}
	responsePacket.SessionId = requestPacket.SessionId

	SDK5_SendResponsePacket(handler, conn, from, packets.SDK5_MATCH_DATA_RESPONSE_PACKET, responsePacket)

	if handler.MatchDataMessageChannel != nil {

		message := messages.MatchDataMessage{}

		message.Timestamp = uint64(time.Now().Unix())
		message.BuyerId = requestPacket.BuyerId
		message.ServerAddress = requestPacket.ServerAddress
		message.DatacenterId = requestPacket.DatacenterId
		message.UserHash = requestPacket.UserHash
		message.SessionId = requestPacket.SessionId
		message.MatchId = requestPacket.MatchId
		message.NumMatchValues = uint32(requestPacket.NumMatchValues)

		for i := 0; i < int(requestPacket.NumMatchValues); i++ {
			message.MatchValues[i] = requestPacket.MatchValues[i]
		}

		handler.MatchDataMessageChannel <- &message

		handler.Events[SDK5_HandlerEvent_SentMatchDataMessage] = true
	}
}

func SDK5_ProcessSessionUpdateRequestPacket(handler *SDK5_Handler, conn *net.UDPConn, from *net.UDPAddr, requestPacket *packets.SDK5_SessionUpdateRequestPacket) {

	handler.Events[SDK5_HandlerEvent_ProcessSessionUpdateRequestPacket] = true

	core.Debug("---------------------------------------------------------------------------")
	core.Debug("received session update request packet from %s", from.String())
	core.Debug("---------------------------------------------------------------------------")

	// make sure we track the length of session update handlers

	timeStart := time.Now()
	defer func() {
		milliseconds := float64(time.Since(timeStart).Milliseconds())
		if milliseconds > 100 {
			// todo: set long duration 
		}
		core.Debug("session update duration: %fms\n-----------------------------------------", milliseconds)
	}()

	// setup session update handler state

	var state SessionUpdateState

	state.Request = requestPacket

	// UnmarshalPacket(&state.Packet, incoming.Data); err != nil {

	// log stuff we want to see with each session update (debug only)

	core.Debug("buyer id is %x", state.Request.BuyerId)
	core.Debug("datacenter id is %x", state.Request.DatacenterId)
	core.Debug("session id is %x", state.Request.SessionId)
	core.Debug("slice number is %d", state.Request.SliceNumber)
	core.Debug("retry number is %d", state.Request.RetryNumber)
	
	/*
	   Build session handler state. Putting everything in a struct makes calling subroutines much easier.
	*/

	/*
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
	*/

	/*
	   Session post *always* runs at the end of this function

	   It writes and sends the response packet back to the sender,
	   and sends session data to billing and the portal.
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

	if state.Request.SliceNumber == 0 {
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

// ------------------------------------------------------------------------

type SessionUpdateState struct {

	/*
	   Convenience state struct for the session update handler.

	   We put all the state in here so it's easy to call out to functions to do work.
z
	   Otherwise we have to pass a million parameters into every function and it gets old fast.
	*/

	Input packets.SDK5_SessionData // sent up from the SDK. previous slice.

	Output packets.SDK5_SessionData // sent down to the SDK. current slice.

	Request            *packets.SDK5_SessionUpdateRequestPacket
	Response           packets.SDK5_SessionUpdateResponsePacket
	Database           *database.Database
	RouteMatrix        *common.RouteMatrix
	Datacenter         database.Datacenter
	Buyer              database.Buyer
	Debug              *string
	
	/*
	IpLocator          *routing.MaxmindDB
	StaleDuration      time.Duration
	RouterPrivateKey   [crypto_old.KeySize]byte
	PostSessionHandler *PostSessionHandler
	*/

	// flags
	SignatureCheckFailed          bool
	UnknownDatacenter             bool
	DatacenterNotEnabled          bool
	BuyerNotFound                 bool
	BuyerNotLive                  bool
	StaleRouteMatrix              bool
	DatacenterAccelerationEnabled bool

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
	PostRouteRelaySellers             [core.MaxRelaysPerRoute]database.Seller
	PostRealPacketLossClientToServer  float32
	PostRealPacketLossServerToClient  float32

	// for convenience
	UnmarshaledSessionData bool
}

func SessionPre(state *SessionUpdateState) bool {

	// var exists bool
	// var err error

	if state.Request.ClientPingTimedOut {

		core.Debug("client ping timed out")

		// todo
		// state.Metrics.ClientPingTimedOut.Add(1)

		// todo
		/*
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
		*/

		return true
	}

	if state.Request.SliceNumber == 0 {

		// todo
		/*
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
		*/

	} else {

		/*
		   For existing sessions, read in the input state from the session data.

		   This is the state.Output from the previous slice.

		   We do this in SessionPre() rather than SessionUpdateExistingSession()
		   in case we early out later on in SessionPre() to ensure location is
		   written back to the SDK.
		*/

		// todo
		/*
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
		*/
	}

	/*
	   If the buyer is "Analysis Only", allow the session to proceed
	   even if the datacenter does not exist, is not enabled, or has zero
	   destination relays in the database.

	   The session will always go direct since the Route State will be disabled.

	   The billing entry will still contain the UnknownDatacenter flag to let
	   us know if we need to add this datacenter for the buyer.
	*/

	// todo
	/*
	if !datacenterExists(state.Database, state.Request.DatacenterId) {
		core.Debug("unknown datacenter")
		state.UnknownDatacenter = true
		if !state.Buyer.RouteShader.AnalysisOnly {
			return true
		}
	}
	*/

	// todo
	/*
	if !datacenterEnabled(state.Database, state.Request.BuyerId, state.Request.DatacenterId) && !state.Buyer.RouteShader.AnalysisOnly {
		core.Debug("datacenter not enabled")
		state.DatacenterNotEnabled = true
		return true
	}
	*/

	// todo
	// state.DatacenterAccelerationEnabled = accelerateDatacenter(state.Database, state.Buyer.ID, state.Packet.DatacenterId)

	// todo
	// state.Datacenter = getDatacenter(state.Database, state.Request.DatacenterID)

	// todo
	// core.Debug("SessionPre(): Datacenter: %s will be accelerated: %v", state.Datacenter.Name, state.DatacenterAccelerationEnabled)

	// todo
	/*
	destRelayIDs := state.RouteMatrix.GetDatacenterRelayIDs(state.Packet.DatacenterID)
	if len(destRelayIDs) == 0 && !state.Buyer.RouteShader.AnalysisOnly && state.DatacenterAccelerationEnabled {
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
		if state.Packet.Tags[i] == crypto_old.HashID("pro") {
			core.Debug("pro mode enabled")
			state.Buyer.RouteShader.ProMode = true
		}
	}
	*/

	state.Output.Initial = false

	return false
}

func SessionUpdateNewSession(state *SessionUpdateState) {

	core.Debug("new session")

	// todo
	// state.Output.Version = SessionDataVersion
	state.Output.SessionId = state.Request.SessionId
	state.Output.SliceNumber = state.Request.SliceNumber + 1
	state.Output.ExpireTimestamp = uint64(time.Now().Unix()) + packets.SDK5_BillingSliceSeconds
	state.Output.RouteState.UserID = state.Request.UserHash
	state.Output.RouteState.ABTest = state.Buyer.RouteShader.ABTest

	state.Input = state.Output
}

func SessionUpdateExistingSession(state *SessionUpdateState) {

	core.Debug("existing session")

	if !state.UnmarshaledSessionData {

		/*
		   If for some reason we did not unmarshal the SessionData
		   already, we must do it here for existing sessions.
		*/

		// todo
		/*
		err := UnmarshalSessionData(&state.Input, state.Packet.SessionData[:])

		if err != nil {
			core.Error("SessionUpdateExistingSession(): could not read session data for buyer %016x:\n\n%s\n", state.Buyer.ID, err)
			state.Metrics.ReadSessionDataFailure.Add(1)
			return
		}
		*/
	}

	/*
	   Check for some obviously divergent data between the session request packet
	   and the stored session data. If there is a mismatch, just return a direct route.
	*/

	if state.Input.SessionId != state.Request.SessionId {
		core.Debug("bad session id")
		// todo
		// state.Metrics.BadSessionID.Add(1)
		return
	}

	if state.Input.SliceNumber != state.Request.SliceNumber {
		core.Debug("bad slice number")
		// todo
		// state.Metrics.BadSliceNumber.Add(1)
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

func SessionHandleFallbackToDirect(state *SessionUpdateState) bool {

	/*
	   Fallback to direct is a state where the SDK has met some fatal error condition.

	   When this happens, the session will go direct from that point forward.

	   Here we look at flags sent up from the SDK, so we can diagnose what caused any fallback to directs to happen.
	*/

	if state.Request.FallbackToDirect && !state.Output.FellBackToDirect {

		core.Debug("fallback to direct")

		state.Output.FellBackToDirect = true

		// todo: laaaaaame
		/*
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
		*/

		return true
	}

	return false
}

// func SessionGetNearRelays(state *SessionHandlerState) bool {

// 	/*
// 	   This function selects up to 32 near relays for the session,
// 	   according to the players latitude and longitude determined by
// 	   ip2location.

// 	   These near relays are selected only on the first slice (slice 0)
// 	   of a session, and are held fixed for the duration of the session.

// 	   The SDK pings the near relays, and reports up the latency, jitter
// 	   and packet loss to each near relay, with each subsequent session
// 	   update (every 10 seconds).

// 	   Network Next uses the relay ping statistics in route planning,
// 	   by adding the latency to the first relay to the total route cost,
// 	   and by excluding near relays with higher jitter or packet loss
// 	   than the default internet route.

// 	   This function is skipped for "Analysis Only" buyers because sessions
// 	   will always take direct.

// 	   This function is skipped for datacenters that are not enabled for
// 	   acceleration, forcing all connected clients to go direct.
// 	*/

// 	if state.Buyer.RouteShader.AnalysisOnly {
// 		core.Debug("analysis only, not getting near relays")
// 		return false
// 	}

// 	if !state.DatacenterAccelerationEnabled {
// 		core.Debug("datacenter acceleration disabled, not getting near relays")
// 		return false
// 	}

// 	directLatency := state.Packet.DirectMinRTT

// 	clientLatitude := state.Output.Location.Latitude
// 	clientLongitude := state.Output.Location.Longitude

// 	serverLatitude := state.Datacenter.Location.Latitude
// 	serverLongitude := state.Datacenter.Location.Longitude

// 	state.Response.NearRelayIDs, state.Response.NearRelayAddresses = state.RouteMatrix.GetNearRelays(directLatency, clientLatitude, clientLongitude, serverLatitude, serverLongitude, core.MaxNearRelays, state.Datacenter.ID)
// 	if len(state.Response.NearRelayIDs) == 0 {
// 		core.Debug("no near relays :(")
// 		state.Metrics.NearRelaysLocateFailure.Add(1)
// 		return false
// 	}

// 	state.Response.NumNearRelays = int32(len(state.Response.NearRelayIDs))
// 	state.Response.HighFrequencyPings = state.Buyer.InternalConfig.HighFrequencyPings && !state.Buyer.InternalConfig.LargeCustomer
// 	state.Response.NearRelaysChanged = true

// 	return true
// }

func SessionUpdateNearRelayStats(state *SessionUpdateState) bool {

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

	   This function is skipped for datacenters that are not enabled for
	   accelerations because sessions will always go direct.
	*/

	routeShader := &state.Buyer.RouteShader

	if routeShader.AnalysisOnly {
		core.Debug("analysis only, not updating near relay stats")
		return false
	}

	if !state.DatacenterAccelerationEnabled {
		core.Debug("datacenter acceleration disabled, not updating near relay stats")
		return false
	}

	routeState := &state.Output.RouteState

	directLatency := int32(math.Ceil(float64(state.Request.DirectMinRTT)))
	directJitter := int32(math.Ceil(float64(state.Request.DirectJitter)))
	directPacketLoss := int32(math.Floor(float64(state.Request.DirectPacketLoss) + 0.5))
	nextPacketLoss := int32(math.Floor(float64(state.Request.NextPacketLoss) + 0.5))

	// todo
	_ = routeState
	_ = directLatency
	_ = directJitter
	_ = directPacketLoss
	_ = nextPacketLoss

	// todo
	/*
	destRelayIDs := state.RouteMatrix.GetDatacenterRelayIDs(state.Datacenter.ID)
	if len(destRelayIDs) == 0 {
		core.Debug("no relays in datacenter %x", state.Datacenter.ID)
		state.Metrics.NoRelaysInDatacenter.Add(1)
		return false
	}

	state.DestRelays = make([]int32, len(destRelayIDs))
	*/

	sliceNumber := int32(state.Request.SliceNumber)

	// todo
	_ = sliceNumber

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
	}

	/*
	   Reframe the near relays to get them in a relay index form relative to the current route matrix.
	*/

	// todo
	/*
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
	*/

	state.NumNearRelays = len(state.Request.NearRelayIds)

	for i := range state.Request.NearRelayIds {
		// todo
		_ = i
		/*
		relayIndex, exists := state.RouteMatrix.RelayIDsToIndices[state.Packet.NearRelayIDs[i]]
		if exists {
			state.NearRelayIndices[i] = relayIndex
		} else {
			state.NearRelayIndices[i] = -1 // near relay no longer exists in route matrix
		}
		*/
	}

	SessionFilterNearRelays(state) // IMPORTANT: Reduce % of sessions that run near relay pings for large customers

	return true
}

func SessionFilterNearRelays(state *SessionUpdateState) {

	/*
	   Reduce the % of sessions running near relay pings for large customers.

	   We do this by only running near relay pings for the first 3 slices, and then holding
	   the near relay ping results fixed for the rest of the session.
	*/

	if !state.Buyer.InternalConfig.LargeCustomer {
		return
	}

	if state.Request.SliceNumber < 4 {
		return
	}

	// IMPORTANT: On any slice after 4, if we haven't already, grab the *processed*
	// near relay RTTs from ReframeRelays, which are set to 255 for any near relays
	// excluded because of high jitter or PL and hold them as the near relay RTTs to use from now on.

	if !state.Input.HoldNearRelays {
		core.Debug("holding near relays")
		state.Output.HoldNearRelays = true
		for i := 0; i < len(state.Request.NearRelayIds); i++ {
			state.Output.HoldNearRelayRTT[i] = state.NearRelayRTTs[i]
		}
	}

	// tell the SDK to stop pinging near relays

	state.Response.ExcludeNearRelays = true
	for i := 0; i < core.MaxNearRelays; i++ {
		state.Response.NearRelayExcluded[i] = true
	}
}

func SessionMakeRouteDecision(state *SessionUpdateState) {

	/*
	   If we are on on network next but don't have any relays in our route, something is WRONG.
	   Veto the session and go direct.
	*/

	if state.Input.RouteState.Next && state.Input.RouteNumRelays == 0 {
		core.Debug("on network next, but no route relays?")
		state.Output.RouteState.Next = false
		state.Output.RouteState.Veto = true
		// todo
		// state.Metrics.NextWithoutRouteRelays.Add(1)
		return
	}

	// todo
	// var stayOnNext bool
	var routeChanged bool
	var routeCost int32
	var routeNumRelays int32

	routeRelays := [core.MaxRelaysPerRoute]int32{}

	// todo
	// sliceNumber := int32(state.Request.SliceNumber)

	if !state.Input.RouteState.Next {

		// currently going direct. should we take network next?

		// todo
		/*
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
		*/

	} else {

		// currently taking network next

		if !state.Request.Next {

			// the sdk aborted this session

			core.Debug("aborted")
			state.Output.RouteState.Next = false
			state.Output.RouteState.Veto = true
			// todo
			// state.Metrics.SDKAborted.Add(1)
			return
		}

		/*
		   Reframe the current route in terms of relay indices in the current route matrix

		   This is necessary because the set of relays in the route matrix change over time.
		*/

		// todo
		/*
		if !core.ReframeRoute(&state.Output.RouteState, state.RouteMatrix.RelayIDsToIndices, state.Output.RouteRelayIDs[:state.Output.RouteNumRelays], &routeRelays) {
			routeRelays = [core.MaxRelaysPerRoute]int32{}
			core.Debug("one or more relays in the route no longer exist")
			// todo
			// state.Metrics.RouteDoesNotExist.Add(1)
		}
		*/

		// stayOnNext, routeChanged = core.MakeRouteDecision_StayOnNetworkNext(state.RouteMatrix.RouteEntries, state.RouteMatrix.FullRelayIndicesSet, state.RouteMatrix.RelayNames, &state.Buyer.RouteShader, &state.Output.RouteState, &state.Buyer.InternalConfig, int32(state.Packet.DirectMinRTT), int32(state.Packet.NextRTT), state.Output.RouteCost, state.RealPacketLoss, state.Packet.NextPacketLoss, state.Output.RouteNumRelays, routeRelays, state.NearRelayIndices[:], state.NearRelayRTTs[:], state.DestRelays[:], &routeCost, &routeNumRelays, routeRelays[:], state.Debug)

 		// todo
		stayOnNext := false
		routeChanged := false

		if stayOnNext {

			// stay on network next

			if routeChanged {
				core.Debug("route changed")
				// todo
				// state.Metrics.RouteSwitched.Add(1)
				// todo
				// BuildNextTokens(&state.Output, state.Database, &state.Buyer, &state.Packet, routeNumRelays, routeRelays[:routeNumRelays], state.RouteMatrix.RelayIDs, state.RouterPrivateKey, &state.Response)
			} else {
				core.Debug("route continued")
				// todo
				// BuildContinueTokens(&state.Output, state.Database, &state.Buyer, &state.Packet, routeNumRelays, routeRelays[:routeNumRelays], state.RouteMatrix.RelayIDs, state.RouterPrivateKey, &state.Response)
			}

		} else {

			// leave network next

			if state.Output.RouteState.NoRoute {
				core.Debug("route no longer exists")
				// todo
				// state.Metrics.NoRoute.Add(1)
			}

			if state.Output.RouteState.MultipathOverload {
				core.Debug("multipath overload")
				// todo
				// state.Metrics.MultipathOverload.Add(1)
			}

			if state.Output.RouteState.Mispredict {
				core.Debug("mispredict")
				// todo
				// state.Metrics.MispredictVeto.Add(1)
			}

			if state.Output.RouteState.LatencyWorse {
				core.Debug("latency worse")
				// todo
				// state.Metrics.LatencyWorse.Add(1)
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

	// todo
	/*
	if routeCost > routing.InvalidRouteValue {
		routeCost = routing.InvalidRouteValue
	}
	*/

	state.Output.RouteCost = routeCost
	state.Output.RouteChanged = routeChanged
	state.Output.RouteNumRelays = routeNumRelays

	for i := int32(0); i < routeNumRelays; i++ {
		relayId := state.RouteMatrix.RelayIds[routeRelays[i]]
		state.Output.RouteRelayIds[i] = relayId
	}
}

func SessionPost(state *SessionUpdateState) {

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

	if state.Request.SliceNumber == 0 {
		// todo
		// SessionGetNearRelays(state)
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
	   Decide if the session was ever on next.

	   We avoid using route type to verify if a session was ever on next
	   in case the route decision to take next was made on the final slice.
	*/

	if state.Request.Next {
		state.Output.EverOnNext = true
	}

	/*
	   Store the packets sent and packets lost counters in the route state,
	   so we can use them to calculate real packet loss next session update.
	*/

	state.Output.PrevPacketsSentClientToServer = state.Request.PacketsSentClientToServer
	state.Output.PrevPacketsSentServerToClient = state.Request.PacketsSentServerToClient
	state.Output.PrevPacketsLostClientToServer = state.Request.PacketsLostClientToServer
	state.Output.PrevPacketsLostServerToClient = state.Request.PacketsLostServerToClient

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

	// todo
	// BuildPostRouteRelayData(state)

	/*
	   Determine if we should write the summary slice. Should only happen
	   when the session is finished.

	   The end of a session occurs when the client ping times out.

	   We always set the output flag to true so that it remains recorded as true on
	   subsequent slices where the client ping has timed out. Instead, we check
	   the input when deciding to write billing entry 2.
	*/

	if state.Request.ClientPingTimedOut {
		state.Output.WroteSummary = true   // todo: weird logic and use of flag
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

	sliceDuration := uint64(packets.SDK5_BillingSliceSeconds)

	// todo: this logic is totally wrong. it should be 20 seconds if initial OR route changed
	if state.Input.Initial && !(state.Output.WroteSummary && state.Input.RouteChanged) {  // todo: weird logic here
		sliceDuration *= 2
	}

	/*
	   Calculate the envelope bandwidth in bytes up and down for the duration of the previous slice.
	*/

	// todo: this is not really necessary any more
	// nextEnvelopeBytesUp, nextEnvelopeBytesDown := CalculateNextBytesUpAndDown(uint64(state.Buyer.RouteShader.BandwidthEnvelopeUpKbps), uint64(state.Buyer.RouteShader.BandwidthEnvelopeDownKbps), sliceDuration)

	/*
	   Calculate the total price for this slice of bandwidth envelope.

	   This is the sum of all relay hop prices, plus our rake, multiplied by the envelope up/down
	   and the length of the session in seconds.
	*/

	// todo: this is not really necessary any more
	// totalPrice := CalculateTotalPriceNibblins(int(state.Input.RouteNumRelays), state.PostRouteRelaySellers, state.PostRouteRelayEgressPriceOverride, nextEnvelopeBytesUp, nextEnvelopeBytesDown)

	/*
	   Store the cumulative sum of totalPrice, nextEnvelopeBytesUp, and nextEnvelopeBytesDown in
	   the output session data. Used in the summary slice.

	   If this is the summary slice, then we do NOT want to include this slice's values in the
	   cumulative sum since this session is finished.

	   This saves datascience some work when analyzing sessions across days.
	*/

	if !state.Output.WroteSummary && state.Request.Next {
		// todo: not needed any more
		/*
		state.Output.TotalPriceSum = state.Input.TotalPriceSum + uint64(totalPrice)
		state.Output.NextEnvelopeBytesUpSum = state.Input.NextEnvelopeBytesUpSum + nextEnvelopeBytesUp
		state.Output.NextEnvelopeBytesDownSum = state.Input.NextEnvelopeBytesDownSum + nextEnvelopeBytesDown
		*/
		state.Output.DurationOnNext = state.Input.DurationOnNext + packets.SDK5_BillingSliceSeconds
	}

	/*
	   Write the session response packet and send it back to the caller.
	*/

	// todo
	/*
	if err := WriteSessionResponse(state.Writer, &state.Response, &state.Output, state.Metrics); err != nil {
		core.Debug("failed to write session update response: %s", err)
		state.Metrics.WriteResponseFailure.Add(1)
		return
	}
	*/

	/*
	   Build post near relay data (for portal, billing etc...)
	*/

	// todo
	// BuildPostNearRelayData(state)

	/*
	   Build billing 2 data and send it to the billing system via pubsub (non-realtime path)

	   Check the input state to see if we wrote the summary slice since
	   the output state is not set to input state if we early out in sessionPre()
	   when the client ping times out.

	   Doing this ensures that we only write the summary slice once since the first time the
	   client ping times out, input flag will be false and the output flag will be true,
	   and on the following slices, both will be true.
	*/

	// todo
	/*
	if state.PostSessionHandler.featureBilling2 && !state.Input.WroteSummary {
		billingEntry2 := BuildBillingEntry2(state, sliceDuration, nextEnvelopeBytesUp, nextEnvelopeBytesDown, totalPrice)

		state.PostSessionHandler.SendBillingEntry2(billingEntry2)
	}
	*/

	/*
	   The client times out at the end of each session, and holds on for 60 seconds.
	   These slices at the end have no useful information for the portal, so we drop
	   them here.
	*/

	if state.Request.ClientPingTimedOut {
		return
	}

	/*
	   Send data to the portal (real-time path)
	*/

	// todo
	/*
	portalData := BuildPortalData(state)

	if portalData.Meta.NextRTT != 0 || portalData.Meta.DirectRTT != 0 {
		state.PostSessionHandler.SendPortalData(portalData)
	}
	*/
}

// func BuildPostRouteRelayData(state *SessionHandlerState) {

// 	/*
// 	   Build information about the relays involved in the current route.

// 	   This data is sent to the portal, and billing system.
// 	*/

// 	for i := int32(0); i < state.Input.RouteNumRelays; i++ {
// 		relay, ok := state.Database.RelayMap[state.Input.RouteRelayIDs[i]]
// 		if ok {
// 			state.PostRouteRelayNames[i] = relay.Name
// 			state.PostRouteRelaySellers[i] = relay.Seller
// 			state.PostRouteRelayEgressPriceOverride[i] = relay.EgressPriceOverride
// 		}
// 	}
// }

// func BuildPostNearRelayData(state *SessionHandlerState) {

// 	state.PostNearRelayCount = int(state.Packet.NumNearRelays)

// 	for i := 0; i < state.PostNearRelayCount; i++ {

// 		/*
// 		   The set of near relays is held fixed at the start of a session.
// 		   Therefore it is possible that a near relay may no longer exist.
// 		*/

// 		relayID := state.Packet.NearRelayIDs[i]
// 		relayIndex, ok := state.RouteMatrix.RelayIDsToIndices[relayID]
// 		if !ok {
// 			continue
// 		}

// 		/*
// 		   Fill in information for near relays needed by billing and the portal.

// 		   We grab this data from the session update packet, which corresponds to the previous slice (input).

// 		   This makes sure all values for a slice in billing and the portal line up temporally.
// 		*/

// 		state.PostNearRelayIDs[i] = relayID
// 		state.PostNearRelayNames[i] = state.RouteMatrix.RelayNames[relayIndex]
// 		state.PostNearRelayAddresses[i] = state.RouteMatrix.RelayAddresses[relayIndex]
// 		state.PostNearRelayRTT[i] = float32(state.Packet.NearRelayRTT[i])
// 		state.PostNearRelayJitter[i] = float32(state.Packet.NearRelayJitter[i])
// 		state.PostNearRelayPacketLoss[i] = float32(state.Packet.NearRelayPacketLoss[i])
// 	}
// }

// func BuildBillingEntry2(state *SessionHandlerState, sliceDuration uint64, nextEnvelopeBytesUp uint64, nextEnvelopeBytesDown uint64, totalPrice routing.Nibblin) *billing.BillingEntry2 {
// 	/*
// 	   Calculate the actual amounts of bytes sent up and down along the network next route
// 	   for the duration of the previous slice (just being reported up from the SDK).

// 	   This is *not* what we bill on.
// 	*/

// 	nextBytesUp, nextBytesDown := CalculateNextBytesUpAndDown(uint64(state.Packet.NextKbpsUp), uint64(state.Packet.NextKbpsDown), sliceDuration)

// 	/*
// 	   Calculate the per-relay hop price that sums up to the total price, minus our rake.
// 	*/

// 	routeRelayPrices := CalculateRouteRelaysPrice(int(state.Input.RouteNumRelays), state.PostRouteRelaySellers, state.PostRouteRelayEgressPriceOverride, nextEnvelopeBytesUp, nextEnvelopeBytesDown)

// 	nextRelayPrice := [core.MaxRelaysPerRoute]uint64{}
// 	for i := 0; i < core.MaxRelaysPerRoute; i++ {
// 		nextRelayPrice[i] = uint64(routeRelayPrices[i])
// 	}

// 	var routeCost int32 = state.Input.RouteCost
// 	if state.Input.RouteCost == math.MaxInt32 {
// 		routeCost = 0
// 	}

// 	/*
// 	   Save the first hop RTT from the client to the first relay in the route.

// 	   This is useful for analysis and saves data science some work.
// 	*/

// 	var nearRelayRTT int32
// 	if state.Input.RouteNumRelays > 0 {
// 		for i, nearRelayID := range state.PostNearRelayIDs {
// 			if nearRelayID == state.Input.RouteRelayIDs[0] {
// 				nearRelayRTT = int32(state.PostNearRelayRTT[i])
// 				break
// 			}
// 		}
// 	}

// 	/*
// 	   If the debug string is set to something by the core routing system, put it in the billing entry.
// 	*/

// 	debugString := ""
// 	if state.Debug != nil {
// 		debugString = *state.Debug
// 	}

// 	/*
// 	   Separate the integer and fractional portions of real packet loss to
// 	   allow for more efficient bitpacking while maintaining precision.
// 	*/

// 	RealPacketLoss, RealPacketLoss_Frac := math.Modf(float64(state.RealPacketLoss))
// 	RealPacketLoss_Frac = math.Round(RealPacketLoss_Frac * 255.0)

// 	/*
// 	   Recast near relay RTT, Jitter, and Packet Loss to int32.
// 	   We do this here since the portal data requires float level precision.
// 	*/

// 	var NearRelayRTTs [core.MaxNearRelays]int32
// 	var NearRelayJitters [core.MaxNearRelays]int32
// 	var nearRelayPacketLosses [core.MaxNearRelays]int32
// 	for i := 0; i < state.PostNearRelayCount; i++ {
// 		NearRelayRTTs[i] = int32(state.PostNearRelayRTT[i])
// 		NearRelayJitters[i] = int32(state.PostNearRelayJitter[i])
// 		nearRelayPacketLosses[i] = int32(state.PostNearRelayPacketLoss[i])
// 	}

// 	/*
// 	   Calculate the session duration in seconds for the summary slice.

// 	   Slice numbers start at 0, so the length of a session is the
// 	   summary slice's slice number * 10 seconds.
// 	*/
// 	var sessionDuration uint32
// 	if state.Output.WroteSummary && state.Packet.SliceNumber != 0 {
// 		sessionDuration = state.Packet.SliceNumber * billing.BillingSliceSeconds
// 	}

// 	/*
// 	   Calculate the starting timestamp of the session to include in the summary slice.
// 	*/
// 	var startTime time.Time
// 	if state.Output.WroteSummary {
// 		secondsToSub := int(sessionDuration)
// 		startTime = time.Now().Add(time.Duration(-secondsToSub) * time.Second)
// 	}

// 	/*
// 	   Create the billing entry 2 and return it to the caller.
// 	*/

// 	billingEntry2 := billing.BillingEntry2{
// 		Version:                         uint32(billing.BillingEntryVersion2),
// 		Timestamp:                       uint32(time.Now().Unix()),
// 		SessionID:                       state.Packet.SessionID,
// 		SliceNumber:                     state.Packet.SliceNumber,
// 		DirectMinRTT:                    int32(state.Packet.DirectMinRTT),
// 		DirectMaxRTT:                    int32(state.Packet.DirectMaxRTT),
// 		DirectPrimeRTT:                  int32(state.Packet.DirectPrimeRTT),
// 		DirectJitter:                    int32(state.Packet.DirectJitter),
// 		DirectPacketLoss:                int32(state.Packet.DirectPacketLoss),
// 		RealPacketLoss:                  int32(RealPacketLoss),
// 		RealPacketLoss_Frac:             uint32(RealPacketLoss_Frac),
// 		RealJitter:                      uint32(state.RealJitter),
// 		Next:                            state.Packet.Next,
// 		Flagged:                         state.Packet.Reported,
// 		Summary:                         state.Output.WroteSummary,
// 		UseDebug:                        state.Buyer.Debug,
// 		Debug:                           debugString,
// 		RouteDiversity:                  int32(state.RouteDiversity),
// 		UserFlags:                       state.Packet.UserFlags,
// 		DatacenterID:                    state.Packet.DatacenterID,
// 		BuyerID:                         state.Packet.BuyerID,
// 		UserHash:                        state.Packet.UserHash,
// 		EnvelopeBytesDown:               nextEnvelopeBytesDown,
// 		EnvelopeBytesUp:                 nextEnvelopeBytesUp,
// 		Latitude:                        float32(state.Input.Location.Latitude),
// 		Longitude:                       float32(state.Input.Location.Longitude),
// 		ClientAddress:                   state.Packet.ClientAddress.String(),
// 		ServerAddress:                   state.Packet.ServerAddress.String(),
// 		ISP:                             state.Input.Location.ISP,
// 		ConnectionType:                  int32(state.Packet.ConnectionType),
// 		PlatformType:                    int32(state.Packet.PlatformType),
// 		SDKVersion:                      state.Packet.Version.String(),
// 		NumTags:                         int32(state.Packet.NumTags),
// 		Tags:                            state.Packet.Tags,
// 		ABTest:                          state.Input.RouteState.ABTest,
// 		Pro:                             state.Buyer.RouteShader.ProMode && !state.Input.RouteState.MultipathRestricted,
// 		ClientToServerPacketsSent:       state.Packet.PacketsSentClientToServer,
// 		ServerToClientPacketsSent:       state.Packet.PacketsSentServerToClient,
// 		ClientToServerPacketsLost:       state.Packet.PacketsLostClientToServer,
// 		ServerToClientPacketsLost:       state.Packet.PacketsLostServerToClient,
// 		ClientToServerPacketsOutOfOrder: state.Packet.PacketsOutOfOrderClientToServer,
// 		ServerToClientPacketsOutOfOrder: state.Packet.PacketsOutOfOrderServerToClient,
// 		NumNearRelays:                   int32(state.PostNearRelayCount),
// 		NearRelayIDs:                    state.PostNearRelayIDs,
// 		NearRelayRTTs:                   NearRelayRTTs,
// 		NearRelayJitters:                NearRelayJitters,
// 		NearRelayPacketLosses:           nearRelayPacketLosses,
// 		EverOnNext:                      state.Input.EverOnNext,
// 		SessionDuration:                 sessionDuration,
// 		TotalPriceSum:                   state.Input.TotalPriceSum,
// 		EnvelopeBytesUpSum:              state.Input.NextEnvelopeBytesUpSum,
// 		EnvelopeBytesDownSum:            state.Input.NextEnvelopeBytesDownSum,
// 		DurationOnNext:                  state.Input.DurationOnNext,
// 		StartTimestamp:                  uint32(startTime.Unix()),
// 		NextRTT:                         int32(state.Packet.NextRTT),
// 		NextJitter:                      int32(state.Packet.NextJitter),
// 		NextPacketLoss:                  int32(state.Packet.NextPacketLoss),
// 		PredictedNextRTT:                routeCost,
// 		NearRelayRTT:                    nearRelayRTT,
// 		NumNextRelays:                   int32(state.Input.RouteNumRelays),
// 		NextRelays:                      state.Input.RouteRelayIDs,
// 		NextRelayPrice:                  nextRelayPrice,
// 		TotalPrice:                      uint64(totalPrice),
// 		Uncommitted:                     !state.Packet.Committed,
// 		Multipath:                       state.Input.RouteState.Multipath,
// 		RTTReduction:                    state.Input.RouteState.ReduceLatency,
// 		PacketLossReduction:             state.Input.RouteState.ReducePacketLoss,
// 		RouteChanged:                    state.Input.RouteChanged,
// 		NextBytesUp:                     nextBytesUp,
// 		NextBytesDown:                   nextBytesDown,
// 		FallbackToDirect:                state.Packet.FallbackToDirect,
// 		MultipathVetoed:                 state.Input.RouteState.MultipathOverload,
// 		Mispredicted:                    state.Input.RouteState.Mispredict,
// 		Vetoed:                          state.Input.RouteState.Veto,
// 		LatencyWorse:                    state.Input.RouteState.LatencyWorse,
// 		NoRoute:                         state.Input.RouteState.NoRoute,
// 		NextLatencyTooHigh:              state.Input.RouteState.NextLatencyTooHigh,
// 		CommitVeto:                      state.Input.RouteState.CommitVeto,
// 		UnknownDatacenter:               state.UnknownDatacenter,
// 		DatacenterNotEnabled:            state.DatacenterNotEnabled,
// 		BuyerNotLive:                    state.BuyerNotLive,
// 		StaleRouteMatrix:                state.StaleRouteMatrix,
// 		TryBeforeYouBuy:                 !state.Input.RouteState.Committed,
// 	}

// 	// Clamp any values to ensure the entry is serialized properly
// 	billingEntry2.ClampEntry()

// 	return &billingEntry2
// }

// func BuildPortalData(state *SessionHandlerState) *SessionPortalData {

// 	/*
// 	   Build the relay hops for the portal
// 	*/

// 	hops := make([]RelayHop, state.Input.RouteNumRelays)
// 	for i := int32(0); i < state.Input.RouteNumRelays; i++ {
// 		hops[i] = RelayHop{
// 			Version: RelayHopVersion,
// 			ID:      state.Input.RouteRelayIDs[i],
// 			Name:    state.PostRouteRelayNames[i],
// 		}
// 	}

// 	/*
// 	   Build the near relay data for the portal
// 	*/

// 	nearRelayPortalData := make([]NearRelayPortalData, state.PostNearRelayCount)
// 	for i := range nearRelayPortalData {
// 		nearRelayPortalData[i] = NearRelayPortalData{
// 			Version: NearRelayPortalDataVersion,
// 			ID:      state.PostNearRelayIDs[i],
// 			Name:    state.PostNearRelayNames[i],
// 			ClientStats: routing.Stats{
// 				RTT:        float64(state.PostNearRelayRTT[i]),
// 				Jitter:     float64(state.PostNearRelayJitter[i]),
// 				PacketLoss: float64(state.PostNearRelayPacketLoss[i]),
// 			},
// 		}
// 	}

// 	/*
// 	   Calculate the delta between network next and direct.

// 	   Clamp the delta RTT above 0. This is used for the top sessions page.
// 	*/

// 	var deltaRTT float32
// 	if state.Packet.Next && state.Packet.NextRTT != 0 && state.Packet.DirectMinRTT >= state.Packet.NextRTT {
// 		deltaRTT = state.Packet.DirectMinRTT - state.Packet.NextRTT
// 	}

// 	/*
// 	   Predicted RTT is the round trip time that we predict, even if we don't
// 	   take network next. It's a conservative prodiction.
// 	*/

// 	predictedRTT := float64(state.Input.RouteCost)
// 	if state.Input.RouteCost >= routing.InvalidRouteValue {
// 		predictedRTT = 0
// 	}

// 	/*
// 	   Build the portal data and return it to the caller.
// 	*/

// 	portalData := SessionPortalData{
// 		Version: SessionPortalDataVersion,
// 		Meta: SessionMeta{
// 			Version:         SessionMetaVersion,
// 			ID:              state.Packet.SessionID,
// 			UserHash:        state.Packet.UserHash,
// 			DatacenterName:  state.Datacenter.Name,
// 			DatacenterAlias: state.Datacenter.AliasName,
// 			OnNetworkNext:   state.Packet.Next,
// 			NextRTT:         float64(state.Packet.NextRTT),
// 			DirectRTT:       float64(state.Packet.DirectMinRTT),
// 			DeltaRTT:        float64(deltaRTT),
// 			Location:        state.Input.Location,
// 			ClientAddr:      state.Packet.ClientAddress.String(),
// 			ServerAddr:      state.Packet.ServerAddress.String(),
// 			Hops:            hops,
// 			SDK:             state.Packet.Version.String(),
// 			Connection:      uint8(state.Packet.ConnectionType),
// 			NearbyRelays:    nearRelayPortalData,
// 			Platform:        uint8(state.Packet.PlatformType),
// 			BuyerID:         state.Packet.BuyerID,
// 		},
// 		Slice: SessionSlice{
// 			Version:   SessionSliceVersion,
// 			Timestamp: time.Now(),
// 			Next: routing.Stats{
// 				RTT:        float64(state.Packet.NextRTT),
// 				Jitter:     float64(state.Packet.NextJitter),
// 				PacketLoss: float64(state.Packet.NextPacketLoss),
// 			},
// 			Direct: routing.Stats{
// 				RTT:        float64(state.Packet.DirectMinRTT),
// 				Jitter:     float64(state.Packet.DirectJitter),
// 				PacketLoss: float64(state.Packet.DirectPacketLoss),
// 			},
// 			Predicted: routing.Stats{
// 				RTT: predictedRTT,
// 			},
// 			ClientToServerStats: routing.Stats{
// 				Jitter:     float64(state.Packet.JitterClientToServer),
// 				PacketLoss: float64(state.PostRealPacketLossClientToServer),
// 			},
// 			ServerToClientStats: routing.Stats{
// 				Jitter:     float64(state.Packet.JitterServerToClient),
// 				PacketLoss: float64(state.PostRealPacketLossServerToClient),
// 			},
// 			RouteDiversity: uint32(state.RouteDiversity),
// 			Envelope: routing.Envelope{
// 				Up:   int64(state.Packet.NextKbpsUp),
// 				Down: int64(state.Packet.NextKbpsDown),
// 			},
// 			IsMultiPath:       state.Input.RouteState.Multipath,
// 			IsTryBeforeYouBuy: !state.Input.RouteState.Committed,
// 			OnNetworkNext:     state.Packet.Next,
// 		},
// 		Point: SessionMapPoint{
// 			Version:   SessionMapPointVersion,
// 			Latitude:  float64(state.Input.Location.Latitude),
// 			Longitude: float64(state.Input.Location.Longitude),
// 			SessionID: state.Input.SessionID,
// 		},
// 		LargeCustomer: state.Buyer.InternalConfig.LargeCustomer,
// 		EverOnNext:    state.Input.EverOnNext,
// 	}

// 	return &portalData
// }
