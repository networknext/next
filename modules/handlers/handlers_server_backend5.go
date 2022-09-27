package handlers

// #cgo pkg-config: libsodium
// #include <sodium.h>
import "C"

import (
	"net"
	"fmt"

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/packets"

	"github.com/networknext/backend/modules-old/routing"
)

const (
	SDK5_HandlerEvent_PacketTooSmall                         = 0
	SDK5_HandlerEvent_UnsupportedPacketType                  = 1
	SDK5_HandlerEvent_BasicPacketFilterFailed                = 2
	SDK5_HandlerEvent_AdvancedPacketFilterFailed             = 3
	SDK5_HandlerEvent_NoRouteMatrix                          = 4
	SDK5_HandlerEvent_NoDatabase                             = 5
	SDK5_HandlerEvent_UnknownBuyer                           = 6
	SDK5_HandlerEvent_SignatureCheckFailed                   = 7
	SDK5_HandlerEvent_BuyerNotLive                           = 8
	SDK5_HandlerEvent_SDKTooOld                              = 9
	SDK5_HandlerEvent_UnknownDatacenter                      = 10

	SDK5_HandlerEvent_CouldNotReadServerInitRequestPacket    = 11
	SDK5_HandlerEvent_CouldNotReadServerUpdateRequestPacket  = 12
	SDK5_HandlerEvent_CouldNotReadSessionUpdateRequestPacket = 13
	SDK5_HandlerEvent_CouldNotReadMatchDataRequestPacket     = 14

	SDK5_HandlerEvent_ProcessServerInitRequestPacket         = 15
	SDK5_HandlerEvent_ProcessServerUpdateRequestPacket       = 16
	SDK5_HandlerEvent_ProcessSessionUpdateRequestPacket      = 17
	SDK5_HandlerEvent_ProcessMatchDataRequestPacket          = 18

	SDK5_HandlerEvent_SentServerInitResponsePacket           = 19
	SDK5_HandlerEvent_SentServerUpdateResponsePacket         = 20
	SDK5_HandlerEvent_SentSessionUpdateResponsePacket        = 21
	SDK5_HandlerEvent_SentMatchDataResponsePacket            = 22

	SDK5_HandlerEvent_NumEvents = 23
)

type SDK5_Handler struct {
	Database             *routing.DatabaseBinWrapper
	RouteMatrix          *common.RouteMatrix
	MaxPacketSize        int
	ServerBackendAddress net.UDPAddr
	PrivateKey           []byte
	GetMagicValues       func() ([]byte, []byte, []byte)
	Events               [SDK5_HandlerEvent_NumEvents]bool
}

func SDK5_PacketHandler(handler *SDK5_Handler, conn *net.UDPConn, from *net.UDPAddr, packetData []byte) {

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

	var buyerId uint64
	index := 16 + 3
	common.ReadUint64(packetData, &index, &buyerId)

	buyer, ok := handler.Database.BuyerMap[buyerId]
	if !ok {
		core.Error("unknown buyer id: %016x", buyerId)
		handler.Events[SDK5_HandlerEvent_UnknownBuyer] = true
		return
	}

	publicKey := buyer.PublicKey

	if !SDK5_CheckPacketSignature(packetData, publicKey) {
		core.Debug("packet signature check failed")
		handler.Events[SDK5_HandlerEvent_SignatureCheckFailed] = true
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

func SDK5_CheckPacketSignature(packetData []byte, publicKey []byte) bool {

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

func SDK5_SignKeypair(publicKey []byte, privateKey []byte) int {
	result := C.crypto_sign_keypair((*C.uchar)(&publicKey[0]), (*C.uchar)(&privateKey[0]))
	return int(result)
}

func SDK5_SignPacket(packetData []byte, privateKey []byte) {
	var state C.crypto_sign_state
	C.crypto_sign_init(&state)
	C.crypto_sign_update(&state, (*C.uchar)(&packetData[0]), C.ulonglong(1))
	C.crypto_sign_update(&state, (*C.uchar)(&packetData[16]), C.ulonglong(len(packetData)-16-2-packets.NEXT_CRYPTO_SIGN_BYTES))
	C.crypto_sign_final_create(&state, (*C.uchar)(&packetData[len(packetData)-2-packets.NEXT_CRYPTO_SIGN_BYTES]), nil, (*C.uchar)(&privateKey[0]))
}

func SDK5_WritePacket[P packets.Packet](packet P, packetType int, maxPacketSize int, from *net.UDPAddr, to *net.UDPAddr, privateKey []byte) ([]byte, error) {

 	buffer := make([]byte, maxPacketSize)

	writeStream := common.CreateWriteStream(buffer[:])

	var dummy [16]byte
	writeStream.SerializeBytes(dummy[:])

	err := packet.Serialize(writeStream)
	if err != nil {
		return nil, fmt.Errorf("failed to write response packet: %v", err)
	}

	writeStream.Flush()

	packetBytes := writeStream.GetBytesProcessed() + packets.NEXT_CRYPTO_SIGN_BYTES + 2

	packetData := buffer[:packetBytes]

	packetData[0] = uint8(packetType)

	var state C.crypto_sign_state
	C.crypto_sign_init(&state)
	C.crypto_sign_update(&state, (*C.uchar)(&packetData[0]), C.ulonglong(1))
	C.crypto_sign_update(&state, (*C.uchar)(&packetData[16]), C.ulonglong(len(packetData)-16-2-packets.NEXT_CRYPTO_SIGN_BYTES))
	result := C.crypto_sign_final_create(&state, (*C.uchar)(&packetData[len(packetData)-2-packets.NEXT_CRYPTO_SIGN_BYTES]), nil, (*C.uchar)(&privateKey[0]))

	if result != 0 {
		return nil, fmt.Errorf("failed to sign response packet: %d", result)
	}

	var magic [8]byte
	var fromAddressBuffer [32]byte
	var toAddressBuffer [32]byte

	fromAddressData, fromAddressPort := core.GetAddressData(from, fromAddressBuffer[:])
	toAddressData, toAddressPort := core.GetAddressData(to, toAddressBuffer[:])

	core.GenerateChonkle(packetData[1:16], magic[:], fromAddressData, fromAddressPort, toAddressData, toAddressPort, packetBytes)

	core.GeneratePittle(packetData[packetBytes-2:], fromAddressData, fromAddressPort, toAddressData, toAddressPort, packetBytes)

	return packetData, nil
}

func SDK5_SendResponsePacket[P packets.Packet](handler *SDK5_Handler, conn *net.UDPConn, to *net.UDPAddr, packetType int, packet P) {

	packetData, err := SDK5_WritePacket(packet, packetType, handler.MaxPacketSize, &handler.ServerBackendAddress, to, handler.PrivateKey)
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

	// todo: send a server init message via pubsub
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

	// todo: send server update message via google pubsub
}

func SDK5_ProcessSessionUpdateRequestPacket(handler *SDK5_Handler, conn *net.UDPConn, from *net.UDPAddr, requestPacket *packets.SDK5_SessionUpdateRequestPacket) {

	handler.Events[SDK5_HandlerEvent_ProcessSessionUpdateRequestPacket] = true

	core.Debug("---------------------------------------------------------------------------")
	core.Debug("received session update request packet from %s", from.String())
	core.Debug("---------------------------------------------------------------------------")

	// todo

	// ...
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

	// todo: build a match data message and send it to pubsub
}
