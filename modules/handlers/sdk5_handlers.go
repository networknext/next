package handlers

import (
	"net"
	"time"
	
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

	// track the length of session update handlers

	timeStart := time.Now()
	defer func() {
		milliseconds := float64(time.Since(timeStart).Milliseconds())
		if milliseconds > 100 {
			// todo: set long duration 
		}
		core.Debug("session update duration: %fms\n-----------------------------------------", milliseconds)
	}()

	// log stuff we want to see with each session update (debug only)

	core.Debug("buyer id is %x", requestPacket.BuyerId)
	core.Debug("datacenter id is %x", requestPacket.DatacenterId)
	core.Debug("session id is %x", requestPacket.SessionId)
	core.Debug("slice number is %d", requestPacket.SliceNumber)
	core.Debug("retry number is %d", requestPacket.RetryNumber)

	/*
	   Build session handler state. Putting everything in a struct makes calling subroutines much easier.
	*/

	var state SessionUpdateState

	state.Request = requestPacket

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
