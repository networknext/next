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
	"github.com/networknext/next/modules/crypto"
)

const (
	SDK_HandlerEvent_PacketTooSmall             = 0
	SDK_HandlerEvent_BasicPacketFilterFailed    = 1
	SDK_HandlerEvent_AdvancedPacketFilterFailed = 2
	SDK_HandlerEvent_NoRouteMatrix              = 3
	SDK_HandlerEvent_NoDatabase                 = 4
	SDK_HandlerEvent_UnknownBuyer               = 5
	SDK_HandlerEvent_SignatureCheckFailed       = 6
	SDK_HandlerEvent_BuyerNotLive               = 7
	SDK_HandlerEvent_SDKTooOld                  = 8
	SDK_HandlerEvent_UnknownDatacenter          = 9
	SDK_HandlerEvent_UnknownRelay               = 10
	SDK_HandlerEvent_DatacenterNotEnabled       = 11

	SDK_HandlerEvent_CouldNotReadServerInitRequestPacket    = 11
	SDK_HandlerEvent_CouldNotReadServerUpdateRequestPacket  = 12
	SDK_HandlerEvent_CouldNotReadSessionUpdateRequestPacket = 13
	SDK_HandlerEvent_CouldNotReadClientRelayRequestPacket   = 14
	SDK_HandlerEvent_CouldNotReadServerRelayRequestPacket   = 15

	SDK_HandlerEvent_ProcessServerInitRequestPacket    = 16
	SDK_HandlerEvent_ProcessServerUpdateRequestPacket  = 17
	SDK_HandlerEvent_ProcessClientRelayRequestPacket   = 18
	SDK_HandlerEvent_ProcessServerRelayRequestPacket   = 19
	SDK_HandlerEvent_ProcessSessionUpdateRequestPacket = 20

	SDK_HandlerEvent_SentServerInitResponsePacket    = 21
	SDK_HandlerEvent_SentServerUpdateResponsePacket  = 22
	SDK_HandlerEvent_SentClientRelayResponsePacket   = 23
	SDK_HandlerEvent_SentServerRelayResponsePacket   = 24
	SDK_HandlerEvent_SentSessionUpdateResponsePacket = 25

	SDK_HandlerEvent_SentAnalyticsServerInitMessage    = 26
	SDK_HandlerEvent_SentAnalyticsServerUpdateMessage  = 27
	SDK_HandlerEvent_SentAnalyticsSessionUpdateMessage = 28

	SDK_HandlerEvent_SentPortalServerUpdateMessage = 29

	SDK_HandlerEvent_NumEvents = 30
)

type SDK_Handler struct {
	Database                *database.Database
	RouteMatrix             *common.RouteMatrix
	MaxPacketSize           int
	ServerBackendAddress    net.UDPAddr
	PingKey                 []byte
	ServerBackendPublicKey  []byte
	ServerBackendPrivateKey []byte
	RelayBackendPublicKey   []byte
	RelayBackendPrivateKey  []byte
	GetMagicValues          func() ([constants.MagicBytes]byte, [constants.MagicBytes]byte, [constants.MagicBytes]byte)
	Events                  [SDK_HandlerEvent_NumEvents]bool
	LocateIP                func(ip net.IP) (float32, float32)
	GetISPAndCountry        func(ip net.IP) (string, string)

	PortalNextSessionsOnly bool

	FallbackToDirectChannel chan<- uint64

	PortalServerUpdateMessageChannel      chan<- *messages.PortalServerUpdateMessage
	PortalSessionUpdateMessageChannel     chan<- *messages.PortalSessionUpdateMessage
	PortalClientRelayUpdateMessageChannel chan<- *messages.PortalClientRelayUpdateMessage
	PortalServerRelayUpdateMessageChannel chan<- *messages.PortalServerRelayUpdateMessage

	AnalyticsServerInitMessageChannel      chan<- *messages.AnalyticsServerInitMessage
	AnalyticsServerUpdateMessageChannel    chan<- *messages.AnalyticsServerUpdateMessage
	AnalyticsClientRelayPingMessageChannel chan<- *messages.AnalyticsClientRelayPingMessage
	AnalyticsServerRelayPingMessageChannel chan<- *messages.AnalyticsServerRelayPingMessage
	AnalyticsSessionUpdateMessageChannel   chan<- *messages.AnalyticsSessionUpdateMessage
	AnalyticsSessionSummaryMessageChannel  chan<- *messages.AnalyticsSessionSummaryMessage
}

func SDK_PacketHandler(handler *SDK_Handler, conn *net.UDPConn, from *net.UDPAddr, packetData []byte) {

	// ignore packets that are too small

	if len(packetData) < packets.SDK_MinPacketBytes {
		core.Debug("packet is too small")
		handler.Events[SDK_HandlerEvent_PacketTooSmall] = true
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

	fromAddressData := core.GetAddressData(from)
	toAddressData := core.GetAddressData(to)

	if !core.AdvancedPacketFilter(packetData, emptyMagic[:], fromAddressData, toAddressData, len(packetData)) {
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
	index := 18 + 3
	encoding.ReadUint64(packetData, &index, &buyerId)

	buyer, ok := handler.Database.BuyerMap[buyerId]
	if !ok {
		core.Error("unknown buyer id: %016x", buyerId)
		handler.Events[SDK_HandlerEvent_UnknownBuyer] = true
		return
	}

	publicKey := buyer.PublicKey

	if !crypto.SDK_CheckPacketSignature(packetData, publicKey) {
		core.Debug("packet signature check failed")
		handler.Events[SDK_HandlerEvent_SignatureCheckFailed] = true
		return
	}

	// process the packet according to type

	packetType := packetData[0]

	packetData = packetData[18:]

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

	case packets.SDK_CLIENT_RELAY_REQUEST_PACKET:
		packet := packets.SDK_ClientRelayRequestPacket{}
		if err := packets.ReadPacket(packetData, &packet); err != nil {
			core.Error("could not read client relay request packet from %s: %v", from.String(), err)
			handler.Events[SDK_HandlerEvent_CouldNotReadClientRelayRequestPacket] = true
			return
		}
		SDK_ProcessClientRelayRequestPacket(handler, conn, from, &packet)
		break

	case packets.SDK_SERVER_RELAY_REQUEST_PACKET:
		packet := packets.SDK_ServerRelayRequestPacket{}
		if err := packets.ReadPacket(packetData, &packet); err != nil {
			core.Error("could not read server relay request packet from %s: %v", from.String(), err)
			handler.Events[SDK_HandlerEvent_CouldNotReadServerRelayRequestPacket] = true
			return
		}
		SDK_ProcessServerRelayRequestPacket(handler, conn, from, &packet)
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

	case packets.SDK_CLIENT_RELAY_RESPONSE_PACKET:
		handler.Events[SDK_HandlerEvent_SentClientRelayResponsePacket] = true
		break

	case packets.SDK_SERVER_RELAY_RESPONSE_PACKET:
		handler.Events[SDK_HandlerEvent_SentServerRelayResponsePacket] = true
		break

	default:
		panic("unknown response packet type")
	}
}

func SDK_ProcessServerInitRequestPacket(handler *SDK_Handler, conn *net.UDPConn, from *net.UDPAddr, requestPacket *packets.SDK_ServerInitRequestPacket) {

	startTimestamp := time.Now().UnixNano()

	handler.Events[SDK_HandlerEvent_ProcessServerInitRequestPacket] = true

	if core.DebugLogs {
		core.Debug("---------------------------------------------------------------------------")
		core.Debug("received server init request packet from %s [%016x]", from.String(), common.HashString(from.String()))
		core.Debug("version: %d.%d.%d", requestPacket.Version.Major, requestPacket.Version.Minor, requestPacket.Version.Patch)
		core.Debug("buyer id: %016x", requestPacket.BuyerId)
		core.Debug("match id: %016x", requestPacket.MatchId)
		core.Debug("request id: %016x", requestPacket.RequestId)
		core.Debug("datacenter: \"%s\" [%016x]", requestPacket.DatacenterName, requestPacket.DatacenterId)
		core.Debug("---------------------------------------------------------------------------")
	}

	upcomingMagic, currentMagic, previousMagic := handler.GetMagicValues()

	responsePacket := &packets.SDK_ServerInitResponsePacket{}
	responsePacket.RequestId = requestPacket.RequestId
	responsePacket.Response = packets.SDK_ServerInitResponseOK
	copy(responsePacket.UpcomingMagic[:], upcomingMagic[:])
	copy(responsePacket.CurrentMagic[:], currentMagic[:])
	copy(responsePacket.PreviousMagic[:], previousMagic[:])

	buyer, exists := handler.Database.BuyerMap[requestPacket.BuyerId]
	if !exists {
		core.Warn("unknown buyer: %016x", requestPacket.BuyerId)
		responsePacket.Response = packets.SDK_ServerInitResponseUnknownBuyer
		handler.Events[SDK_HandlerEvent_UnknownBuyer] = true
	}

	if !buyer.Live {
		core.Warn("buyer not live: %016x", requestPacket.BuyerId)
		responsePacket.Response = packets.SDK_ServerInitResponseBuyerNotActive
		handler.Events[SDK_HandlerEvent_BuyerNotLive] = true
	}

	if !requestPacket.Version.AtLeast(packets.SDKVersion{1, 0, 0}) {
		core.Warn("sdk version is too old: %s", requestPacket.Version.String())
		responsePacket.Response = packets.SDK_ServerInitResponseSDKVersionTooOld
		handler.Events[SDK_HandlerEvent_SDKTooOld] = true
	}

	buyerSettings, exists := handler.Database.BuyerDatacenterSettings[requestPacket.BuyerId]

	if !exists {

		core.Warn("datacenter '%s' [%016x] is not enabled for buyer %016x (1) [%s]", requestPacket.DatacenterName, requestPacket.DatacenterId, requestPacket.BuyerId, from.String())
		responsePacket.Response = packets.SDK_ServerInitResponseDatacenterNotEnabled
		handler.Events[SDK_HandlerEvent_DatacenterNotEnabled] = true

	} else {

		datacenterSettings, exists := buyerSettings[requestPacket.DatacenterId]
		if !exists || !datacenterSettings.EnableAcceleration {
			core.Warn("datacenter '%s' [%016x] is not enabled for buyer %016x (2) [%s]", requestPacket.DatacenterName, requestPacket.DatacenterId, requestPacket.BuyerId, from.String())
			responsePacket.Response = packets.SDK_ServerInitResponseDatacenterNotEnabled
			handler.Events[SDK_HandlerEvent_DatacenterNotEnabled] = true
		}

	}

	SDK_SendResponsePacket(handler, conn, from, packets.SDK_SERVER_INIT_RESPONSE_PACKET, responsePacket)

	if handler.AnalyticsServerInitMessageChannel != nil {

		message := messages.AnalyticsServerInitMessage{}

		message.Timestamp = startTimestamp / 1000 // nano -> milliseconds
		message.SDKVersion_Major = int32(requestPacket.Version.Major)
		message.SDKVersion_Minor = int32(requestPacket.Version.Minor)
		message.SDKVersion_Patch = int32(requestPacket.Version.Patch)
		message.BuyerId = int64(requestPacket.BuyerId)
		message.MatchId = int64(requestPacket.MatchId)
		message.DatacenterId = int64(requestPacket.DatacenterId)
		message.DatacenterName = requestPacket.DatacenterName
		message.ServerAddress = from.String()
		message.ServerId = int64(common.HashString(from.String()))

		handler.AnalyticsServerInitMessageChannel <- &message

		handler.Events[SDK_HandlerEvent_SentAnalyticsServerInitMessage] = true
	}
}

func SDK_ProcessServerUpdateRequestPacket(handler *SDK_Handler, conn *net.UDPConn, from *net.UDPAddr, requestPacket *packets.SDK_ServerUpdateRequestPacket) {

	startTimestamp := time.Now().UnixNano()

	handler.Events[SDK_HandlerEvent_ProcessServerUpdateRequestPacket] = true

	if core.DebugLogs {
		core.Debug("---------------------------------------------------------------------------")
		core.Debug("received server update request packet from %s", from.String())
		core.Debug("version: %d.%d.%d", requestPacket.Version.Major, requestPacket.Version.Minor, requestPacket.Version.Patch)
		core.Debug("buyer id: %016x", requestPacket.BuyerId)
		core.Debug("match id: %016x", requestPacket.MatchId)
		core.Debug("request id: %016x", requestPacket.RequestId)
		core.Debug("server id: %016x", requestPacket.ServerId)
		core.Debug("datacenter id: %016x", requestPacket.DatacenterId)
		core.Debug("---------------------------------------------------------------------------")
	}

	defer func() {

		if handler.AnalyticsServerUpdateMessageChannel != nil {

			message := messages.AnalyticsServerUpdateMessage{}

			message.Timestamp = startTimestamp / 1000 // nano -> milliseconds
			message.SDKVersion_Major = int32(requestPacket.Version.Major)
			message.SDKVersion_Minor = int32(requestPacket.Version.Minor)
			message.SDKVersion_Patch = int32(requestPacket.Version.Patch)
			message.BuyerId = int64(requestPacket.BuyerId)
			message.MatchId = int64(requestPacket.MatchId)
			message.DatacenterId = int64(requestPacket.DatacenterId)
			message.NumSessions = int32(requestPacket.NumSessions)
			message.ServerId = int64(requestPacket.ServerId)
			message.ServerAddress = from.String()
			message.DeltaTimeMin = requestPacket.DeltaTimeMin
			message.DeltaTimeMax = requestPacket.DeltaTimeMax
			message.DeltaTimeAvg = requestPacket.DeltaTimeAvg

			handler.AnalyticsServerUpdateMessageChannel <- &message

			handler.Events[SDK_HandlerEvent_SentAnalyticsServerUpdateMessage] = true
		}
	}()

	defer func() {

		if handler.PortalServerUpdateMessageChannel != nil {

			message := messages.PortalServerUpdateMessage{}

			message.Timestamp = uint64(startTimestamp / 1000000000) // nanoseconds -> seconds
			message.SDKVersion_Major = byte(requestPacket.Version.Major)
			message.SDKVersion_Minor = byte(requestPacket.Version.Minor)
			message.SDKVersion_Patch = byte(requestPacket.Version.Patch)
			message.BuyerId = requestPacket.BuyerId
			message.ServerId = requestPacket.ServerId
			message.DatacenterId = requestPacket.DatacenterId
			message.NumSessions = requestPacket.NumSessions
			message.Uptime = requestPacket.Uptime

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

	if !requestPacket.Version.AtLeast(packets.SDKVersion{1, 0, 0}) {
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
		core.Warn("unknown datacenter %016x", requestPacket.DatacenterId)
		handler.Events[SDK_HandlerEvent_UnknownDatacenter] = true
	}

	SDK_SendResponsePacket(handler, conn, from, packets.SDK_SERVER_UPDATE_RESPONSE_PACKET, responsePacket)
}

func SDK_ProcessSessionUpdateRequestPacket(handler *SDK_Handler, conn *net.UDPConn, from *net.UDPAddr, requestPacket *packets.SDK_SessionUpdateRequestPacket) {

	handler.Events[SDK_HandlerEvent_ProcessSessionUpdateRequestPacket] = true

	if core.DebugLogs {
		core.Debug("---------------------------------------------------------------------------")
		core.Debug("received session update request packet from %s [%x]", from.String(), common.HashString(from.String()))
		core.Debug("---------------------------------------------------------------------------")
	}

	/*
	   Build session handler state. Putting everything in a struct makes calling subroutines much easier.
	*/

	var state SessionUpdateState

	state.PingKey = handler.PingKey
	state.RelayBackendPublicKey = handler.RelayBackendPublicKey
	state.RelayBackendPrivateKey = handler.RelayBackendPrivateKey
	state.ServerBackendPublicKey = handler.ServerBackendPublicKey
	state.ServerBackendPrivateKey = handler.ServerBackendPrivateKey
	state.ServerBackendAddress = &handler.ServerBackendAddress
	state.From = from
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

	state.PortalNextSessionsOnly = handler.PortalNextSessionsOnly

	state.FallbackToDirectChannel = handler.FallbackToDirectChannel

	state.PortalSessionUpdateMessageChannel = handler.PortalSessionUpdateMessageChannel
	state.PortalClientRelayUpdateMessageChannel = handler.PortalClientRelayUpdateMessageChannel
	state.PortalServerRelayUpdateMessageChannel = handler.PortalServerRelayUpdateMessageChannel

	state.AnalyticsClientRelayPingMessageChannel = handler.AnalyticsClientRelayPingMessageChannel
	state.AnalyticsServerRelayPingMessageChannel = handler.AnalyticsServerRelayPingMessageChannel
	state.AnalyticsSessionUpdateMessageChannel = handler.AnalyticsSessionUpdateMessageChannel
	state.AnalyticsSessionSummaryMessageChannel = handler.AnalyticsSessionSummaryMessageChannel

	state.GetISPAndCountry = handler.GetISPAndCountry

	// track the length of session update handlers

	timeStart := time.Now()
	defer func() {
		milliseconds := int(time.Since(timeStart).Milliseconds())
		if milliseconds > 100 {
			state.LongSessionUpdate = true
			core.Warn("long session update: %dms", milliseconds)
		}
		core.Debug("---------------------------------------------------------------------------")
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
	   Process client and server relay ping stats.

	   We use client and server relay latency, jitter and packet loss for route planning.
	*/

	SessionUpdate_UpdateClientRelays(&state)

	SessionUpdate_UpdateServerRelays(&state)

	/*
	   Decide whether we should take network next or not.
	*/

	SessionUpdate_MakeRouteDecision(&state)

	core.Debug("session updated successfully")
}

func SDK_ProcessClientRelayRequestPacket(handler *SDK_Handler, conn *net.UDPConn, from *net.UDPAddr, requestPacket *packets.SDK_ClientRelayRequestPacket) {

	handler.Events[SDK_HandlerEvent_ProcessClientRelayRequestPacket] = true

	if core.DebugLogs {
		core.Debug("---------------------------------------------------------------------------")
		core.Debug("received client relay request packet from %s", from.String())
		core.Debug("version: %d.%d.%d", requestPacket.Version.Major, requestPacket.Version.Minor, requestPacket.Version.Patch)
		core.Debug("buyer id: %016x", requestPacket.BuyerId)
		core.Debug("request id: %016x", requestPacket.RequestId)
		core.Debug("---------------------------------------------------------------------------")
	}

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

	if !requestPacket.Version.AtLeast(packets.SDKVersion{1, 0, 0}) {
		core.Debug("sdk version is too old: %s", requestPacket.Version.String())
		handler.Events[SDK_HandlerEvent_SDKTooOld] = true
		return
	}

	datacenter := handler.Database.GetDatacenter(requestPacket.DatacenterId)

	if datacenter == nil {
		core.Debug("datacenter is nil, not getting client relays")
		handler.Events[SDK_HandlerEvent_UnknownDatacenter] = true
		return
	}

	clientLatitude, clientLongitude := handler.LocateIP(requestPacket.ClientAddress.IP)

	serverLatitude := datacenter.Latitude
	serverLongitude := datacenter.Longitude

	const distanceThreshold = 2500
	const latencyThreshold = 30.0

	clientRelayIds, clientRelayAddresses := common.GetClientRelays(constants.MaxClientRelays,
		distanceThreshold,
		latencyThreshold,
		handler.RouteMatrix.RelayIds,
		handler.RouteMatrix.RelayAddresses,
		handler.RouteMatrix.RelayLatitudes,
		handler.RouteMatrix.RelayLongitudes,
		clientLatitude,
		clientLongitude,
		serverLatitude,
		serverLongitude,
	)

	numClientRelays := len(clientRelayIds)

	core.Debug("found %d client relays", numClientRelays)

	responsePacket := &packets.SDK_ClientRelayResponsePacket{}
	responsePacket.RequestId = requestPacket.RequestId
	responsePacket.ClientAddress = requestPacket.ClientAddress
	responsePacket.Latitude = clientLatitude
	responsePacket.Longitude = clientLongitude
	responsePacket.NumClientRelays = int32(numClientRelays)
	responsePacket.ExpireTimestamp = uint64(time.Now().Unix()) + 15

	clientAddressWithoutPort := requestPacket.ClientAddress
	clientAddressWithoutPort.Port = 0

	for i := 0; i < numClientRelays; i++ {
		responsePacket.ClientRelayIds[i] = clientRelayIds[i]
		responsePacket.ClientRelayAddresses[i] = clientRelayAddresses[i]

		core.GeneratePingToken(responsePacket.ExpireTimestamp, &clientAddressWithoutPort, &responsePacket.ClientRelayAddresses[i], handler.PingKey, responsePacket.ClientRelayPingTokens[i][:])
	}

	SDK_SendResponsePacket(handler, conn, from, packets.SDK_CLIENT_RELAY_RESPONSE_PACKET, responsePacket)
}

func SDK_ProcessServerRelayRequestPacket(handler *SDK_Handler, conn *net.UDPConn, from *net.UDPAddr, requestPacket *packets.SDK_ServerRelayRequestPacket) {

	handler.Events[SDK_HandlerEvent_ProcessServerRelayRequestPacket] = true

	if core.DebugLogs {
		core.Debug("---------------------------------------------------------------------------")
		core.Debug("received server relay request packet from %s", from.String())
		core.Debug("version: %d.%d.%d", requestPacket.Version.Major, requestPacket.Version.Minor, requestPacket.Version.Patch)
		core.Debug("buyer id: %016x", requestPacket.BuyerId)
		core.Debug("datacenter id: %016x", requestPacket.DatacenterId)
		core.Debug("request id: %016x", requestPacket.RequestId)
		core.Debug("---------------------------------------------------------------------------")
	}

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

	if !requestPacket.Version.AtLeast(packets.SDKVersion{1, 0, 0}) {
		core.Debug("sdk version is too old: %s", requestPacket.Version.String())
		handler.Events[SDK_HandlerEvent_SDKTooOld] = true
		return
	}

	datacenter := handler.Database.GetDatacenter(requestPacket.DatacenterId)

	if datacenter == nil {
		core.Debug("datacenter is nil, not getting server relays")
		handler.Events[SDK_HandlerEvent_UnknownDatacenter] = true
		return
	}

	datacenterRelays := handler.Database.GetDatacenterRelays(requestPacket.DatacenterId)

	if len(datacenterRelays) > constants.MaxDestRelays {
		datacenterRelays = datacenterRelays[:constants.MaxDestRelays]
	}

	numServerRelays := len(datacenterRelays)

	core.Debug("found %d server relays", numServerRelays)

	responsePacket := &packets.SDK_ServerRelayResponsePacket{}
	responsePacket.RequestId = requestPacket.RequestId
	responsePacket.NumServerRelays = int32(numServerRelays)
	responsePacket.ExpireTimestamp = uint64(time.Now().Unix()) + 15

	for i := 0; i < numServerRelays; i++ {
		relay := handler.Database.GetRelay(datacenterRelays[i])
		if relay == nil {
			core.Debug("unknown relay %x", datacenterRelays[i])
			handler.Events[SDK_HandlerEvent_UnknownRelay] = true
			return
		}
		responsePacket.ServerRelayIds[i] = datacenterRelays[i]
		responsePacket.ServerRelayAddresses[i] = relay.PublicAddress

		core.GeneratePingToken(responsePacket.ExpireTimestamp, from, &responsePacket.ServerRelayAddresses[i], handler.PingKey, responsePacket.ServerRelayPingTokens[i][:])
	}

	SDK_SendResponsePacket(handler, conn, from, packets.SDK_SERVER_RELAY_RESPONSE_PACKET, responsePacket)
}
