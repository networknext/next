package handlers

import (
	"net"
	"time"

	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/constants"
	"github.com/networknext/next/modules/core"
	"github.com/networknext/next/modules/database"
	"github.com/networknext/next/modules/encoding"
	"github.com/networknext/next/modules/messages"
	"github.com/networknext/next/modules/packets"
)

const (
	SDK_HandlerEvent_PacketTooSmall             = 0
	SDK_HandlerEvent_UnsupportedPacketType      = 1
	SDK_HandlerEvent_BasicPacketFilterFailed    = 2
	SDK_HandlerEvent_AdvancedPacketFilterFailed = 3
	SDK_HandlerEvent_NoRouteMatrix              = 4
	SDK_HandlerEvent_NoDatabase                 = 5
	SDK_HandlerEvent_UnknownBuyer               = 6
	SDK_HandlerEvent_SignatureCheckFailed       = 7
	SDK_HandlerEvent_BuyerNotLive               = 8
	SDK_HandlerEvent_SDKTooOld                  = 9
	SDK_HandlerEvent_UnknownDatacenter          = 10

	SDK_HandlerEvent_CouldNotReadServerInitRequestPacket    = 11
	SDK_HandlerEvent_CouldNotReadServerUpdateRequestPacket  = 12
	SDK_HandlerEvent_CouldNotReadSessionUpdateRequestPacket = 13
	SDK_HandlerEvent_CouldNotReadMatchDataRequestPacket     = 14

	SDK_HandlerEvent_ProcessServerInitRequestPacket    = 15
	SDK_HandlerEvent_ProcessServerUpdateRequestPacket  = 16
	SDK_HandlerEvent_ProcessSessionUpdateRequestPacket = 17
	SDK_HandlerEvent_ProcessMatchDataRequestPacket     = 18

	SDK_HandlerEvent_SentServerInitResponsePacket    = 19
	SDK_HandlerEvent_SentServerUpdateResponsePacket  = 20
	SDK_HandlerEvent_SentSessionUpdateResponsePacket = 21
	SDK_HandlerEvent_SentMatchDataResponsePacket     = 22

	SDK_HandlerEvent_SentAnalyticsServerInitMessage    = 23
	SDK_HandlerEvent_SentAnalyticsServerUpdateMessage  = 24
	SDK_HandlerEvent_SentAnalyticsSessionUpdateMessage = 25
	SDK_HandlerEvent_SentAnalyticsMatchDataMessage     = 26

	SDK_HandlerEvent_SentPortalServerUpdateMessage = 27

	SDK_HandlerEvent_NumEvents = 28
)

type SDK_Handler struct {
	Database                *database.Database
	RouteMatrix             *common.RouteMatrix
	MaxPacketSize           int
	ServerBackendAddress    net.UDPAddr
	PingKey                 []byte
	ServerBackendPublicKey  []byte
	ServerBackendPrivateKey []byte
	RelayBackendPrivateKey  []byte
	GetMagicValues          func() ([constants.MagicBytes]byte, [constants.MagicBytes]byte, [constants.MagicBytes]byte)
	Events                  [SDK_HandlerEvent_NumEvents]bool
	LocateIP                func(ip net.IP) (float32, float32)

	PortalServerUpdateMessageChannel    chan<- *messages.PortalServerUpdateMessage
	PortalSessionUpdateMessageChannel   chan<- *messages.PortalSessionUpdateMessage
	PortalNearRelayUpdateMessageChannel chan<- *messages.PortalNearRelayUpdateMessage
	PortalMapUpdateMessageChannel       chan<- *messages.PortalMapUpdateMessage

	AnalyticsServerInitMessageChannel     chan<- *messages.AnalyticsServerInitMessage
	AnalyticsServerUpdateMessageChannel   chan<- *messages.AnalyticsServerUpdateMessage
	AnalyticsNearRelayPingMessageChannel  chan<- *messages.AnalyticsNearRelayPingMessage
	AnalyticsSessionUpdateMessageChannel  chan<- *messages.AnalyticsSessionUpdateMessage
	AnalyticsSessionSummaryMessageChannel chan<- *messages.AnalyticsSessionSummaryMessage
	AnalyticsMatchDataMessageChannel      chan<- *messages.AnalyticsMatchDataMessage
}

func SDK_PacketHandler(handler *SDK_Handler, conn *net.UDPConn, from *net.UDPAddr, packetData []byte) {

	// ignore packets that are too small

	if len(packetData) < packets.SDK_MinPacketBytes {
		core.Debug("packet is too small")
		handler.Events[SDK_HandlerEvent_PacketTooSmall] = true
		return
	}

	// ignore packet types we don't support

	packetType := packetData[0]

	if packetType != packets.SDK_SERVER_INIT_REQUEST_PACKET && packetType != packets.SDK_SERVER_UPDATE_REQUEST_PACKET && packetType != packets.SDK_SESSION_UPDATE_REQUEST_PACKET && packetType != packets.SDK_MATCH_DATA_REQUEST_PACKET {
		core.Debug("unsupported packet type %d", packetType)
		handler.Events[SDK_HandlerEvent_UnsupportedPacketType] = true
		return
	}

	// make sure the basic packet filter passes

	if !core.BasicPacketFilter(packetData[:], len(packetData)) {
		core.Debug("basic packet filter failed for %d byte packet from %s", len(packetData), from.String())
		handler.Events[SDK_HandlerEvent_BasicPacketFilterFailed] = true
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
		handler.Events[SDK_HandlerEvent_AdvancedPacketFilterFailed] = true
		return
	}

	// we can't process any packets without these

	if handler.RouteMatrix == nil {
		core.Debug("ignoring packet because we don't have a route matrix")
		handler.Events[SDK_HandlerEvent_NoRouteMatrix] = true
		return
	}

	if handler.Database == nil {
		core.Debug("ignoring packet because we don't have a database")
		handler.Events[SDK_HandlerEvent_NoDatabase] = true
		return
	}

	// check packet signature

	var buyerId uint64
	index := 16 + 3
	encoding.ReadUint64(packetData, &index, &buyerId)

	buyer, ok := handler.Database.BuyerMap[buyerId]
	if !ok {
		core.Error("unknown buyer id: %016x", buyerId)
		handler.Events[SDK_HandlerEvent_UnknownBuyer] = true
		return
	}

	publicKey := buyer.PublicKey

	if !packets.SDK_CheckPacketSignature(packetData, publicKey) {
		core.Debug("packet signature check failed")
		handler.Events[SDK_HandlerEvent_SignatureCheckFailed] = true
		return
	}

	// process the packet according to type

	packetData = packetData[16:]

	switch packetType {

	case packets.SDK_SERVER_INIT_REQUEST_PACKET:
		packet := packets.SDK_ServerInitRequestPacket{}
		if err := packets.ReadPacket(packetData, &packet); err != nil {
			core.Error("could not read server init request packet from %s: %v", from.String(), err)
			handler.Events[SDK_HandlerEvent_CouldNotReadServerInitRequestPacket] = true
			return
		}
		SDK_ProcessServerInitRequestPacket(handler, conn, from, &packet)
		break

	case packets.SDK_SERVER_UPDATE_REQUEST_PACKET:
		packet := packets.SDK_ServerUpdateRequestPacket{}
		if err := packets.ReadPacket(packetData, &packet); err != nil {
			core.Error("could not read server update request packet from %s: %v", from.String(), err)
			handler.Events[SDK_HandlerEvent_CouldNotReadServerUpdateRequestPacket] = true
			return
		}
		SDK_ProcessServerUpdateRequestPacket(handler, conn, from, &packet)
		break

	case packets.SDK_SESSION_UPDATE_REQUEST_PACKET:
		packet := packets.SDK_SessionUpdateRequestPacket{}
		if err := packets.ReadPacket(packetData, &packet); err != nil {
			core.Error("could not read session update request packet from %s: %v", from.String(), err)
			handler.Events[SDK_HandlerEvent_CouldNotReadSessionUpdateRequestPacket] = true
			return
		}
		SDK_ProcessSessionUpdateRequestPacket(handler, conn, from, &packet)
		break

	case packets.SDK_MATCH_DATA_REQUEST_PACKET:
		packet := packets.SDK_MatchDataRequestPacket{}
		if err := packets.ReadPacket(packetData, &packet); err != nil {
			core.Error("could not read match data request packet from %s: %v", from.String(), err)
			handler.Events[SDK_HandlerEvent_CouldNotReadMatchDataRequestPacket] = true
			return
		}
		SDK_ProcessMatchDataRequestPacket(handler, conn, from, &packet)
		break

	default:
		panic("unknown packet type")
	}
}

func SDK_SendResponsePacket[P packets.Packet](handler *SDK_Handler, conn *net.UDPConn, to *net.UDPAddr, packetType int, packet P) {

	packetData, err := packets.SDK_WritePacket(packet, packetType, handler.MaxPacketSize, &handler.ServerBackendAddress, to, handler.ServerBackendPrivateKey)
	if err != nil {
		core.Error("failed to write response packet: %v", err)
		return
	}

	if conn != nil {
		if _, err := conn.WriteToUDP(packetData, to); err != nil {
			core.Error("failed to send response packet: %v", err)
			return
		}
	}

	switch packetType {

	case packets.SDK_SERVER_INIT_RESPONSE_PACKET:
		handler.Events[SDK_HandlerEvent_SentServerInitResponsePacket] = true
		break

	case packets.SDK_SERVER_UPDATE_RESPONSE_PACKET:
		handler.Events[SDK_HandlerEvent_SentServerUpdateResponsePacket] = true
		break

	case packets.SDK_SESSION_UPDATE_RESPONSE_PACKET:
		handler.Events[SDK_HandlerEvent_SentSessionUpdateResponsePacket] = true
		break

	case packets.SDK_MATCH_DATA_RESPONSE_PACKET:
		handler.Events[SDK_HandlerEvent_SentMatchDataResponsePacket] = true
		break

	default:
		panic("unknown response packet type")
	}
}

func SDK_ProcessServerInitRequestPacket(handler *SDK_Handler, conn *net.UDPConn, from *net.UDPAddr, requestPacket *packets.SDK_ServerInitRequestPacket) {

	handler.Events[SDK_HandlerEvent_ProcessServerInitRequestPacket] = true

	core.Debug("---------------------------------------------------------------------------")
	core.Debug("received server init request packet from %s", from.String())
	core.Debug("version: %d.%d.%d", requestPacket.Version.Major, requestPacket.Version.Minor, requestPacket.Version.Patch)
	core.Debug("buyer id: %016x", requestPacket.BuyerId)
	core.Debug("request id: %016x", requestPacket.RequestId)
	core.Debug("datacenter: \"%s\" [%016x]", requestPacket.DatacenterName, requestPacket.DatacenterId)
	core.Debug("---------------------------------------------------------------------------")

	upcomingMagic, currentMagic, previousMagic := handler.GetMagicValues()

	responsePacket := &packets.SDK_ServerInitResponsePacket{}
	responsePacket.RequestId = requestPacket.RequestId
	responsePacket.Response = packets.SDK_ServerInitResponseOK
	copy(responsePacket.UpcomingMagic[:], upcomingMagic[:])
	copy(responsePacket.CurrentMagic[:], currentMagic[:])
	copy(responsePacket.PreviousMagic[:], previousMagic[:])

	buyer, exists := handler.Database.BuyerMap[requestPacket.BuyerId]
	if !exists {
		core.Debug("unknown buyer: %016x", requestPacket.BuyerId)
		responsePacket.Response = packets.SDK_ServerInitResponseUnknownBuyer
		handler.Events[SDK_HandlerEvent_UnknownBuyer] = true
	}

	if !buyer.Live {
		core.Debug("buyer not live: %016x", requestPacket.BuyerId)
		responsePacket.Response = packets.SDK_ServerInitResponseBuyerNotActive
		handler.Events[SDK_HandlerEvent_BuyerNotLive] = true
	}

	if !requestPacket.Version.AtLeast(packets.SDKVersion{5, 0, 0}) {
		core.Debug("sdk version is too old: %s", requestPacket.Version.String())
		responsePacket.Response = packets.SDK_ServerInitResponseOldSDKVersion
		handler.Events[SDK_HandlerEvent_SDKTooOld] = true
	}

	_, exists = handler.Database.DatacenterMap[requestPacket.DatacenterId]
	if !exists {
		// IMPORTANT: Let the server init succeed even if the datacenter is unknown!
		core.Debug("unknown datacenter '%s' [%016x]", requestPacket.DatacenterName, requestPacket.DatacenterId)
		handler.Events[SDK_HandlerEvent_UnknownDatacenter] = true
	}

	SDK_SendResponsePacket(handler, conn, from, packets.SDK_SERVER_INIT_RESPONSE_PACKET, responsePacket)

	if handler.AnalyticsServerInitMessageChannel != nil {

		message := messages.AnalyticsServerInitMessage{}

		message.Version = messages.AnalyticsServerInitMessageVersion_Write
		message.Timestamp = uint64(time.Now().Unix())
		message.SDKVersion_Major = byte(requestPacket.Version.Major)
		message.SDKVersion_Minor = byte(requestPacket.Version.Minor)
		message.SDKVersion_Patch = byte(requestPacket.Version.Patch)
		message.BuyerId = requestPacket.BuyerId
		message.DatacenterId = requestPacket.DatacenterId
		message.DatacenterName = requestPacket.DatacenterName
		message.ServerAddress = *from

		handler.AnalyticsServerInitMessageChannel <- &message

		handler.Events[SDK_HandlerEvent_SentAnalyticsServerInitMessage] = true
	}
}

func SDK_ProcessServerUpdateRequestPacket(handler *SDK_Handler, conn *net.UDPConn, from *net.UDPAddr, requestPacket *packets.SDK_ServerUpdateRequestPacket) {

	handler.Events[SDK_HandlerEvent_ProcessServerUpdateRequestPacket] = true

	core.Debug("---------------------------------------------------------------------------")
	core.Debug("received server update request packet from %s", from.String())
	core.Debug("version: %d.%d.%d", requestPacket.Version.Major, requestPacket.Version.Minor, requestPacket.Version.Patch)
	core.Debug("buyer id: %016x", requestPacket.BuyerId)
	core.Debug("request id: %016x", requestPacket.RequestId)
	core.Debug("datacenter id: %016x", requestPacket.DatacenterId)
	core.Debug("---------------------------------------------------------------------------")

	defer func() {

		if handler.AnalyticsServerUpdateMessageChannel != nil {

			message := messages.AnalyticsServerUpdateMessage{}

			message.Version = messages.AnalyticsServerUpdateMessageVersion_Write
			message.Timestamp = uint64(time.Now().Unix())
			message.SDKVersion_Major = byte(requestPacket.Version.Major)
			message.SDKVersion_Minor = byte(requestPacket.Version.Minor)
			message.SDKVersion_Patch = byte(requestPacket.Version.Patch)
			message.BuyerId = requestPacket.BuyerId
			message.MatchId = requestPacket.MatchId
			message.DatacenterId = requestPacket.DatacenterId
			message.NumSessions = requestPacket.NumSessions
			message.ServerAddress = *from

			handler.AnalyticsServerUpdateMessageChannel <- &message

			handler.Events[SDK_HandlerEvent_SentAnalyticsServerUpdateMessage] = true
		}
	}()

	defer func() {

		if handler.PortalServerUpdateMessageChannel != nil {

			message := messages.PortalServerUpdateMessage{}

			message.Version = messages.PortalServerUpdateMessageVersion_Write
			message.Timestamp = uint64(time.Now().Unix())
			message.SDKVersion_Major = byte(requestPacket.Version.Major)
			message.SDKVersion_Minor = byte(requestPacket.Version.Minor)
			message.SDKVersion_Patch = byte(requestPacket.Version.Patch)
			message.BuyerId = requestPacket.BuyerId
			message.MatchId = requestPacket.MatchId
			message.DatacenterId = requestPacket.DatacenterId
			message.NumSessions = requestPacket.NumSessions
			message.StartTime = requestPacket.StartTime
			message.ServerAddress = *from

			handler.PortalServerUpdateMessageChannel <- &message

			handler.Events[SDK_HandlerEvent_SentPortalServerUpdateMessage] = true
		}
	}()

	buyer, exists := handler.Database.BuyerMap[requestPacket.BuyerId]
	if !exists {
		core.Debug("unknown buyer: %016x", requestPacket.BuyerId)
		handler.Events[SDK_HandlerEvent_UnknownBuyer] = true
		return
	}

	if !buyer.Live {
		core.Debug("buyer not live: %016x", requestPacket.BuyerId)
		handler.Events[SDK_HandlerEvent_BuyerNotLive] = true
		return
	}

	if !requestPacket.Version.AtLeast(packets.SDKVersion{5, 0, 0}) {
		core.Debug("sdk version is too old: %s", requestPacket.Version.String())
		handler.Events[SDK_HandlerEvent_SDKTooOld] = true
		return
	}

	upcomingMagic, currentMagic, previousMagic := handler.GetMagicValues()

	responsePacket := &packets.SDK_ServerUpdateResponsePacket{}
	responsePacket.RequestId = requestPacket.RequestId
	copy(responsePacket.UpcomingMagic[:], upcomingMagic[:])
	copy(responsePacket.CurrentMagic[:], currentMagic[:])
	copy(responsePacket.PreviousMagic[:], previousMagic[:])

	_, exists = handler.Database.DatacenterMap[requestPacket.DatacenterId]
	if !exists {
		// IMPORTANT: Let the server update succeed, even if the datacenter is unknown
		core.Debug("unknown datacenter %016x", requestPacket.DatacenterId)
		handler.Events[SDK_HandlerEvent_UnknownDatacenter] = true
	}

	SDK_SendResponsePacket(handler, conn, from, packets.SDK_SERVER_UPDATE_RESPONSE_PACKET, responsePacket)
}

func SDK_ProcessMatchDataRequestPacket(handler *SDK_Handler, conn *net.UDPConn, from *net.UDPAddr, requestPacket *packets.SDK_MatchDataRequestPacket) {

	handler.Events[SDK_HandlerEvent_ProcessMatchDataRequestPacket] = true

	core.Debug("---------------------------------------------------------------------------")
	core.Debug("received match data request packet from %s", from.String())
	core.Debug("---------------------------------------------------------------------------")

	buyer, exists := handler.Database.BuyerMap[requestPacket.BuyerId]
	if !exists {
		core.Debug("unknown buyer: %016x", requestPacket.BuyerId)
		handler.Events[SDK_HandlerEvent_UnknownBuyer] = true
		return
	}

	if !buyer.Live {
		core.Debug("buyer not live: %016x", requestPacket.BuyerId)
		handler.Events[SDK_HandlerEvent_BuyerNotLive] = true
		return
	}

	if !requestPacket.Version.AtLeast(packets.SDKVersion{5, 0, 0}) {
		core.Debug("sdk version is too old: %s", requestPacket.Version.String())
		handler.Events[SDK_HandlerEvent_SDKTooOld] = true
		return
	}

	responsePacket := &packets.SDK_MatchDataResponsePacket{}
	responsePacket.SessionId = requestPacket.SessionId

	SDK_SendResponsePacket(handler, conn, from, packets.SDK_MATCH_DATA_RESPONSE_PACKET, responsePacket)

	if handler.AnalyticsMatchDataMessageChannel != nil {

		message := messages.AnalyticsMatchDataMessage{}

		message.Timestamp = uint64(time.Now().Unix())
		message.BuyerId = requestPacket.BuyerId
		message.ServerAddress = requestPacket.ServerAddress
		message.DatacenterId = requestPacket.DatacenterId
		message.SessionId = requestPacket.SessionId
		message.MatchId = requestPacket.MatchId
		message.NumMatchValues = uint32(requestPacket.NumMatchValues)

		for i := 0; i < int(requestPacket.NumMatchValues); i++ {
			message.MatchValues[i] = requestPacket.MatchValues[i]
		}

		handler.AnalyticsMatchDataMessageChannel <- &message

		handler.Events[SDK_HandlerEvent_SentAnalyticsMatchDataMessage] = true
	}
}

func SDK_ProcessSessionUpdateRequestPacket(handler *SDK_Handler, conn *net.UDPConn, from *net.UDPAddr, requestPacket *packets.SDK_SessionUpdateRequestPacket) {

	handler.Events[SDK_HandlerEvent_ProcessSessionUpdateRequestPacket] = true

	core.Debug("---------------------------------------------------------------------------")
	core.Debug("received session update request packet from %s", from.String())
	core.Debug("---------------------------------------------------------------------------")

	/*
	   Build session handler state. Putting everything in a struct makes calling subroutines much easier.
	*/

	var state SessionUpdateState

	state.PingKey = handler.PingKey
	state.RelayBackendPrivateKey = handler.RelayBackendPrivateKey
	state.ServerBackendPublicKey = handler.ServerBackendPublicKey
	state.ServerBackendPrivateKey = handler.ServerBackendPrivateKey
	state.ServerBackendAddress = &handler.ServerBackendAddress
	state.From = from
	state.LocateIP = handler.LocateIP
	state.Buyer = handler.Database.BuyerMap[requestPacket.BuyerId]

	state.Request = requestPacket
	state.Database = handler.Database
	state.RouteMatrix = handler.RouteMatrix
	state.StaleDuration = 30 * time.Second
	state.Response = packets.SDK_SessionUpdateResponsePacket{
		SessionId:   state.Request.SessionId,
		SliceNumber: state.Request.SliceNumber,
		RouteType:   packets.SDK_RouteTypeDirect,
	}

	state.PortalSessionUpdateMessageChannel = handler.PortalSessionUpdateMessageChannel
	state.PortalNearRelayUpdateMessageChannel = handler.PortalNearRelayUpdateMessageChannel
	state.PortalMapUpdateMessageChannel = handler.PortalMapUpdateMessageChannel

	state.AnalyticsNearRelayPingMessageChannel = handler.AnalyticsNearRelayPingMessageChannel
	state.AnalyticsSessionUpdateMessageChannel = handler.AnalyticsSessionUpdateMessageChannel
	state.AnalyticsSessionSummaryMessageChannel = handler.AnalyticsSessionSummaryMessageChannel

	// track the length of session update handlers

	timeStart := time.Now()
	defer func() {
		milliseconds := float64(time.Since(timeStart).Milliseconds())
		if milliseconds > 100 {
			state.LongSessionUpdate = true
		}
		core.Warn("long session update: %fms\n-----------------------------------------", milliseconds)
	}()

	// log stuff we want to see with each session update (debug only)

	core.Debug("buyer id is %x", requestPacket.BuyerId)
	core.Debug("datacenter id is %x", requestPacket.DatacenterId)
	core.Debug("session id is %x", requestPacket.SessionId)
	core.Debug("slice number is %d", requestPacket.SliceNumber)
	core.Debug("retry number is %d", requestPacket.RetryNumber)

	/*
	   Session post *always* runs at the end of this function
	*/

	defer func() {
		SessionUpdate_Post(&state)
		if len(state.ResponsePacket) > 0 {
			handler.Events[SDK_HandlerEvent_SentSessionUpdateResponsePacket] = true
			if conn != nil {
				if _, err := conn.WriteToUDP(state.ResponsePacket, state.From); err != nil {
					core.Error("failed to send session update response packet: %v", err)
					return
				}
			}
		}
	}()

	/*
	   Call session pre function

	   This function checks for early out conditions and does some setup of the handler state.

	   If it returns true, one of the early out conditions has been met, so we return early.
	*/

	if SessionUpdate_Pre(&state) {
		state.Output = state.Input
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
		SessionUpdate_NewSession(&state)
	} else {
		SessionUpdate_ExistingSession(&state)
	}

	/*
	   Process near relay ping statistics after the first slice.

	   We use near relay latency, jitter and packet loss for route planning.
	*/

	if state.Request.SliceNumber > 0 {
		SessionUpdate_UpdateNearRelays(&state)
	}

	/*
	   Decide whether we should take network next or not.
	*/

	SessionUpdate_MakeRouteDecision(&state)

	core.Debug("session updated successfully")
}
