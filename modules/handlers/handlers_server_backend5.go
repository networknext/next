package handlers

// #cgo pkg-config: libsodium
// #include <sodium.h>
import "C"

import (
	"net"

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/packets"
	"github.com/networknext/backend/modules/routing"
)

const (
	SDK5_HandlerEvent_PacketTooSmall                         = 0
	SDK5_HandlerEvent_UnsupportedPacketType                  = 1
	SDK5_HandlerEvent_BasicPacketFilterFailed                = 2
	SDK5_HandlerEvent_AdvancedPacketFilterFailed             = 3
	SDK5_HandlerEvent_PacketSignatureCheckFailed             = 4
	SDK5_HandlerEvent_NoRouteMatrix                          = 5
	SDK5_HandlerEvent_NoDatabase                             = 6
	SDK5_HandlerEvent_CouldNotReadServerInitRequestPacket    = 7
	SDK5_HandlerEvent_CouldNotReadServerUpdateRequestPacket  = 8
	SDK5_HandlerEvent_CouldNotReadSessionUpdateRequestPacket = 9
	SDK5_HandlerEvent_CouldNotReadMatchDataRequestPacket     = 10
	SDK5_HandlerEvent_UnknownPacketType                      = 11

	SDK5_HandlerEvent_NumEvents = 12
)

type SDK5_HandlerData struct {
	Database             *routing.DatabaseBinWrapper
	RouteMatrix          *common.RouteMatrix
	MaxPacketSize        int
	ServerBackendAddress net.UDPAddr
	PrivateKey           []byte
	GetMagicValues       func() ([]byte, []byte, []byte)
	Events               [SDK5_HandlerEvent_NumEvents]bool
}

func SDK5_PacketHandler(handler *SDK5_HandlerData, conn *net.UDPConn, from *net.UDPAddr, packetData []byte) {

	// ignore packets that are too small

	if len(packetData) < 16+3+4+packets.NEXT_CRYPTO_SIGN_BYTES+2 {
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

	if !CheckPacketSignature(packetData, handler.RouteMatrix, handler.Database) {
		core.Debug("packet signature check failed")
		handler.Events[SDK5_HandlerEvent_PacketSignatureCheckFailed] = true
		return
	}

	// process the packet according to type

	packetData = packetData[16 : len(packetData)-(2+packets.NEXT_CRYPTO_SIGN_BYTES)]

	switch packetType {

	case packets.SDK5_SERVER_INIT_REQUEST_PACKET:
		packet := packets.SDK5_ServerInitRequestPacket{}
		if err := packets.ReadPacket(packetData, &packet); err != nil {
			core.Error("could not read server init request packet from %s", from.String())
			handler.Events[SDK5_HandlerEvent_CouldNotReadServerInitRequestPacket] = true
			return
		}
		ProcessServerInitRequestPacket(handler, conn, from, &packet)
		break

	case packets.SDK5_SERVER_UPDATE_REQUEST_PACKET:
		packet := packets.SDK5_ServerUpdateRequestPacket{}
		if err := packets.ReadPacket(packetData, &packet); err != nil {
			core.Error("could not read server update request packet from %s", from.String())
			handler.Events[SDK5_HandlerEvent_CouldNotReadServerUpdateRequestPacket] = true
			return
		}
		ProcessServerUpdateRequestPacket(handler, conn, from, &packet)
		break

	case packets.SDK5_SESSION_UPDATE_REQUEST_PACKET:
		packet := packets.SDK5_SessionUpdateRequestPacket{}
		if err := packets.ReadPacket(packetData, &packet); err != nil {
			core.Error("could not read session update request packet from %s", from.String())
			handler.Events[SDK5_HandlerEvent_CouldNotReadSessionUpdateRequestPacket] = true
			return
		}
		ProcessSessionUpdateRequestPacket(handler, conn, from, &packet)
		break

	case packets.SDK5_MATCH_DATA_REQUEST_PACKET:
		packet := packets.SDK5_MatchDataRequestPacket{}
		if err := packets.ReadPacket(packetData, &packet); err != nil {
			core.Error("could not read match data request packet from %s", from.String())
			handler.Events[SDK5_HandlerEvent_CouldNotReadMatchDataRequestPacket] = true
			return
		}
		ProcessMatchDataRequestPacket(handler, conn, from, &packet)
		break

	default:
		core.Debug("received unknown packet type %d from %s", packetType, from.String())
		handler.Events[SDK5_HandlerEvent_UnknownPacketType] = true
	}
}

func CheckPacketSignature(packetData []byte, routeMatrix *common.RouteMatrix, database *routing.DatabaseBinWrapper) bool {

	var buyerId uint64
	index := 16 + 3
	common.ReadUint64(packetData, &index, &buyerId)

	buyer, ok := database.BuyerMap[buyerId]
	if !ok {
		core.Error("unknown buyer id: %016x", buyerId)
		return false
	}

	publicKey := buyer.PublicKey

	var state C.crypto_sign_state
	C.crypto_sign_init(&state)
	C.crypto_sign_update(&state, (*C.uchar)(&packetData[0]), C.ulonglong(1))
	C.crypto_sign_update(&state, (*C.uchar)(&packetData[16]), C.ulonglong(len(packetData)-16-2-packets.NEXT_CRYPTO_SIGN_BYTES))
	result := C.crypto_sign_final_verify(&state, (*C.uchar)(&packetData[len(packetData)-2-packets.NEXT_CRYPTO_SIGN_BYTES]), (*C.uchar)(&publicKey[0]))

	if result != 0 {
		core.Error("signed packet did not verify")
		return false
	}

	return true
}

func SendResponsePacket[P packets.Packet](handler *SDK5_HandlerData, conn *net.UDPConn, to *net.UDPAddr, packetType int, packet P) {

	buffer := make([]byte, handler.MaxPacketSize)

	writeStream := common.CreateWriteStream(buffer[:])

	var dummy [16]byte
	writeStream.SerializeBytes(dummy[:])

	err := packet.Serialize(writeStream)
	if err != nil {
		core.Error("failed to write response packet: %v", err)
		return
	}

	writeStream.Flush()

	packetBytes := writeStream.GetBytesProcessed() + packets.NEXT_CRYPTO_SIGN_BYTES + 2

	packetData := buffer[:packetBytes]

	packetData[0] = uint8(packetType)

	var state C.crypto_sign_state
	C.crypto_sign_init(&state)
	C.crypto_sign_update(&state, (*C.uchar)(&packetData[0]), C.ulonglong(1))
	C.crypto_sign_update(&state, (*C.uchar)(&packetData[16]), C.ulonglong(len(packetData)-16-2-packets.NEXT_CRYPTO_SIGN_BYTES))
	result := C.crypto_sign_final_create(&state, (*C.uchar)(&packetData[len(packetData)-2-packets.NEXT_CRYPTO_SIGN_BYTES]), nil, (*C.uchar)(&handler.PrivateKey[0]))

	if result != 0 {
		core.Error("failed to sign response packet")
		return
	}

	var magic [8]byte
	var fromAddressBuffer [32]byte
	var toAddressBuffer [32]byte

	fromAddressData, fromAddressPort := core.GetAddressData(&handler.ServerBackendAddress, fromAddressBuffer[:])
	toAddressData, toAddressPort := core.GetAddressData(to, toAddressBuffer[:])

	core.GenerateChonkle(packetData[1:16], magic[:], fromAddressData, fromAddressPort, toAddressData, toAddressPort, packetBytes)

	core.GeneratePittle(packetData[packetBytes-2:], fromAddressData, fromAddressPort, toAddressData, toAddressPort, packetBytes)

	if _, err := conn.WriteToUDP(packetData, to); err != nil {
		core.Error("failed to send response packet: %v", err)
		return
	}
}

func ProcessServerInitRequestPacket(handler *SDK5_HandlerData, conn *net.UDPConn, from *net.UDPAddr, requestPacket *packets.SDK5_ServerInitRequestPacket) {

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
	}

	if !buyer.Live {
		core.Debug("buyer not live: %016x", requestPacket.BuyerId)
		responsePacket.Response = packets.SDK5_ServerInitResponseBuyerNotActive
	}

	if !requestPacket.Version.AtLeast(packets.SDKVersion{5, 0, 0}) {
		core.Debug("sdk version is too old: %s", requestPacket.Version.String())
		responsePacket.Response = packets.SDK5_ServerInitResponseOldSDKVersion
	}

	_, exists = handler.Database.DatacenterMap[requestPacket.DatacenterId]
	if !exists {
		core.Debug("unknown datacenter %s [%016x]", requestPacket.DatacenterName, requestPacket.DatacenterId)
	}

	SendResponsePacket(handler, conn, from, packets.SDK5_SERVER_INIT_RESPONSE_PACKET, responsePacket)
}

func ProcessServerUpdateRequestPacket(handler *SDK5_HandlerData, conn *net.UDPConn, from *net.UDPAddr, requestPacket *packets.SDK5_ServerUpdateRequestPacket) {

	core.Debug("---------------------------------------------------------------------------")
	core.Debug("received server update request packet from %s", from.String())
	core.Debug("version: %d.%d.%d", requestPacket.Version.Major, requestPacket.Version.Minor, requestPacket.Version.Patch)
	core.Debug("buyer id: %016x", requestPacket.BuyerId)
	core.Debug("request id: %016x", requestPacket.RequestId)
	core.Debug("datacenter id: %016x", requestPacket.DatacenterId)
	core.Debug("---------------------------------------------------------------------------")

	upcomingMagic, currentMagic, previousMagic := handler.GetMagicValues()

	responsePacket := &packets.SDK5_ServerInitResponsePacket{}
	responsePacket.RequestId = requestPacket.RequestId
	copy(responsePacket.UpcomingMagic[:], upcomingMagic[:])
	copy(responsePacket.CurrentMagic[:], currentMagic[:])
	copy(responsePacket.PreviousMagic[:], previousMagic[:])

	buyer, exists := handler.Database.BuyerMap[requestPacket.BuyerId]
	if !exists {
		core.Debug("unknown buyer: %016x", requestPacket.BuyerId)
		return
	}

	if !buyer.Live {
		core.Debug("buyer not live: %016x", requestPacket.BuyerId)
		responsePacket.Response = packets.SDK5_ServerInitResponseBuyerNotActive
		return
	}

	if !requestPacket.Version.AtLeast(packets.SDKVersion{5, 0, 0}) {
		core.Debug("sdk version is too old: %s", requestPacket.Version.String())
		responsePacket.Response = packets.SDK5_ServerInitResponseOldSDKVersion
		return
	}

	_, exists = handler.Database.DatacenterMap[requestPacket.DatacenterId]
	if !exists {
		core.Debug("unknown datacenter %016x", requestPacket.DatacenterId)
	}

	// todo: send server update message to bigquery via google pubsub

	SendResponsePacket(handler, conn, from, packets.SDK5_SERVER_UPDATE_RESPONSE_PACKET, responsePacket)
}

func ProcessSessionUpdateRequestPacket(handler *SDK5_HandlerData, conn *net.UDPConn, from *net.UDPAddr, requestPacket *packets.SDK5_SessionUpdateRequestPacket) {

	core.Debug("---------------------------------------------------------------------------")
	core.Debug("received session update request packet from %s", from.String())
	core.Debug("---------------------------------------------------------------------------")

	// ...
}

func ProcessMatchDataRequestPacket(handler *SDK5_HandlerData, conn *net.UDPConn, from *net.UDPAddr, requestPacket *packets.SDK5_MatchDataRequestPacket) {
	core.Debug("---------------------------------------------------------------------------")
	core.Debug("received match data request packet from %s", from.String())
	core.Debug("---------------------------------------------------------------------------")

	// ...
}
