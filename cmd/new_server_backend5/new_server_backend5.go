package main

// #cgo pkg-config: libsodium
// #include <sodium.h>
import "C"

import (
	"net"
	"time"
	"net/http"
	"io/ioutil"
	"sync"
	"bytes"
	"encoding/gob"

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/packets"
	"github.com/networknext/backend/modules/routing"
)

var maxPacketSize int
var serverBackendAddress net.UDPAddr
var routeMatrixURI string
var routeMatrixInterval time.Duration

var routeMatrixMutex sync.RWMutex
var routeMatrix *common.RouteMatrix
var database *routing.DatabaseBinWrapper

func main() {

	service := common.CreateService("new_server_backend5")

	maxPacketSize = envvar.GetInt("UDP_MAX_PACKET_SIZE", 4096)
	serverBackendAddress = *envvar.GetAddress("SERVER_BACKEND_ADDRESS", core.ParseAddress("127.0.0.1:45000"))
	routeMatrixURI = envvar.GetString("ROUTE_MATRIX_URI", "http://127.0.0.1:30001/route_matrix")
	routeMatrixInterval = envvar.GetDuration("ROUTE_MATRIX_INTERVAL", time.Second)

	core.Log("max packet size: %d bytes", maxPacketSize)
	core.Log("server backend address: %s", serverBackendAddress.String())
	core.Log("route matrix uri: %s", routeMatrixURI)
	core.Log("route matrix interval: %s", routeMatrixInterval.String())

	UpdateRouteMatrix(service)

	service.OverrideHealthHandler(healthHandler)

	service.StartUDPServer(packetHandler)

	service.UpdateMagic()

	service.StartWebServer()

	service.WaitForShutdown()
}

func healthHandler(w http.ResponseWriter, r *http.Request) {

	routeMatrixMutex.RLock()
	not_ready := routeMatrix == nil
	routeMatrixMutex.RUnlock()

	if not_ready {
		w.WriteHeader(http.StatusInternalServerError)
	} else {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(http.StatusText(http.StatusOK)))
	}
}

func packetHandler(conn *net.UDPConn, from *net.UDPAddr, packetData []byte) {

	// ignore packets that are too small

	if len(packetData) < 16+3+4+packets.NEXT_CRYPTO_SIGN_BYTES+2 {
		core.Debug("packet is too small")
		return
	}

	// ignore packet types we don't support

	packetType := packetData[0]

	if packetType != packets.SDK5_SERVER_INIT_REQUEST_PACKET && packetType != packets.SDK5_SERVER_UPDATE_REQUEST_PACKET && packetType != packets.SDK5_SESSION_UPDATE_REQUEST_PACKET && packetType != packets.SDK5_MATCH_DATA_REQUEST_PACKET {
		core.Debug("unsupported packet type %d", packetType)
		return
	}

	// make sure the basic packet filter passes

	if !core.BasicPacketFilter(packetData[:], len(packetData)) {
		core.Debug("basic packet filter failed for %d byte packet from %s", len(packetData), from.String())
		return
	}

	// make sure the advanced packet filter passes

	to := &serverBackendAddress

	var emptyMagic [8]byte

	var fromAddressBuffer [32]byte
	var toAddressBuffer [32]byte

	fromAddressData, fromAddressPort := core.GetAddressData(from, fromAddressBuffer[:])
	toAddressData, toAddressPort := core.GetAddressData(to, toAddressBuffer[:])

	if !core.AdvancedPacketFilter(packetData, emptyMagic[:], fromAddressData, fromAddressPort, toAddressData, toAddressPort, len(packetData)) {
		core.Debug("advanced packet filter failed for %d byte packet from %s to %s", len(packetData), from.String(), to.String())
		return
	}

	// get route matrix and database (under same lock)

	routeMatrix, database := GetRouteMatrixAndDatabase()

	if routeMatrix == nil {
		core.Debug("ignoring packet because we don't have a route matrix")
		return
	}

	if database == nil {
		core.Debug("ignoring packet because we don't have a database")
		return
	}

	// check packet signature

	if !CheckPacketSignature(packetData, routeMatrix, database) {
		core.Debug("packet signature check failed")
		return
	}

	// process the packet according to type

	packetData = packetData[16 : len(packetData)-(2+packets.NEXT_CRYPTO_SIGN_BYTES)]

	switch packetType {

	case packets.SDK5_SERVER_INIT_REQUEST_PACKET:
		packet := packets.SDK5_ServerInitRequestPacket{}
		if err := packets.ReadPacket(packetData, &packet); err != nil {
			core.Error("could not read server init request packet from %s", from.String())
			return
		}
		ProcessServerInitRequestPacket(conn, from, &packet)
		break

	case packets.SDK5_SERVER_UPDATE_REQUEST_PACKET:
		packet := packets.SDK5_ServerUpdateRequestPacket{}
		if err := packets.ReadPacket(packetData, &packet); err != nil {
			core.Error("could not read server update request packet from %s", from.String())
			return
		}
		ProcessServerUpdateRequestPacket(conn, from, &packet)
		break

	case packets.SDK5_SESSION_UPDATE_REQUEST_PACKET:
		packet := packets.SDK5_SessionUpdateRequestPacket{}
		if err := packets.ReadPacket(packetData, &packet); err != nil {
			core.Error("could not read session update request packet from %s", from.String())
			return
		}
		ProcessSessionUpdateRequestPacket(conn, from, &packet)
		break

	case packets.SDK5_MATCH_DATA_REQUEST_PACKET:
		packet := packets.SDK5_MatchDataRequestPacket{}
		if err := packets.ReadPacket(packetData, &packet); err != nil {
			core.Error("could not read match data request packet from %s", from.String())
			return
		}
		ProcessMatchDataRequestPacket(conn, from, &packet)
		break

	default:
		core.Debug("received unknown packet type %d from %s", packetType, from.String())
	}
}

func CheckPacketSignature(packetData []byte, routeMatrix *common.RouteMatrix, database *routing.DatabaseBinWrapper) bool {

	var buyerId uint64
	index := 16 + 3
	common.ReadUint64(packetData, &index, &buyerId)

	buyer, ok := database.BuyerMap[buyerId]
	if !ok {
		core.Error("unknown buyer id: %016x", buyerId)
		return false;
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

func SendResponsePacket[P packets.Packet](conn *net.UDPConn, to *net.UDPAddr, packet P) {

	buffer := make([]byte, maxPacketSize)

	writeStream := common.CreateWriteStream(buffer[:])

	err := packet.Serialize(writeStream)
	if err != nil {
		core.Error("failed to write response packet: %v", err)
	}

	writeStream.Flush()

	packetBytes := writeStream.GetBytesProcessed()
	packetData := buffer[:packetBytes]

	// todo: [packet type][chonkle](packet data)[signature](pittle)

	if _, err := conn.WriteToUDP(packetData, to); err != nil {
		core.Error("failed to send response packet: %v", err)
	}
}

func ProcessServerInitRequestPacket(conn *net.UDPConn, from *net.UDPAddr, packet *packets.SDK5_ServerInitRequestPacket) {
	core.Debug("---------------------------------------------------------------------------")
	core.Debug("received server init request packet from %s", from.String())
	core.Debug("version: %d.%d.%d", packet.Version.Major, packet.Version.Minor, packet.Version.Patch)
	core.Debug("buyer id: %016x", packet.BuyerId)
	core.Debug("request id: %016x", packet.RequestId)
	core.Debug("datacenter: \"%s\" [%016x]", packet.DatacenterName, packet.DatacenterId)
	core.Debug("---------------------------------------------------------------------------")
}

func ProcessServerUpdateRequestPacket(conn *net.UDPConn, from *net.UDPAddr, packet *packets.SDK5_ServerUpdateRequestPacket) {
	core.Debug("---------------------------------------------------------------------------")
	core.Debug("received server update request packet from %s", from.String())
	// ...
	core.Debug("---------------------------------------------------------------------------")
}

func ProcessSessionUpdateRequestPacket(conn *net.UDPConn, from *net.UDPAddr, packet *packets.SDK5_SessionUpdateRequestPacket) {
	core.Debug("---------------------------------------------------------------------------")
	core.Debug("received session update request packet from %s", from.String())
	// ...
	core.Debug("---------------------------------------------------------------------------")
}

func ProcessMatchDataRequestPacket(conn *net.UDPConn, from *net.UDPAddr, packet *packets.SDK5_MatchDataRequestPacket) {
	core.Debug("---------------------------------------------------------------------------")
	core.Debug("received match data request packet from %s", from.String())
	// ...
	core.Debug("---------------------------------------------------------------------------")
}

func UpdateRouteMatrix(service *common.Service) {

	httpClient := &http.Client{
		Timeout: routeMatrixInterval,
	}

	ticker := time.NewTicker(routeMatrixInterval)

	go func() {
		for {
			select {

			case <-service.Context.Done():
				return

			case <-ticker.C:

				routeMatrixMutex.Lock()
				currentRouteMatrix := routeMatrix
				routeMatrixMutex.Unlock()

				if currentRouteMatrix != nil && time.Now().Unix() - int64(currentRouteMatrix.CreatedAt) > 30 {
					core.Error("route matrix is stale")
					routeMatrixMutex.Lock()
					routeMatrix = nil
					routeMatrixMutex.Unlock()
				}

				response, err := httpClient.Get(routeMatrixURI)
				if err != nil {
					core.Error("failed to http get route matrix: %v", err)
					continue
				}

				buffer, err := ioutil.ReadAll(response.Body)
				if err != nil {
					core.Error("failed to read route matrix: %v", err)
					continue
				}

				response.Body.Close()

				newRouteMatrix := common.RouteMatrix{}

				err = newRouteMatrix.Read(buffer)
				if err != nil {
					core.Error("failed to read route matrix: %v", err)
					continue
				}

				var newDatabase routing.DatabaseBinWrapper

				databaseBuffer := bytes.NewBuffer(newRouteMatrix.BinFileData)
				decoder := gob.NewDecoder(databaseBuffer)
				err = decoder.Decode(&newDatabase)
				if err != nil {
					core.Error("failed to read database: %v", err)
					continue
				}

				routeMatrixMutex.Lock()
				routeMatrix = &newRouteMatrix
				database = &newDatabase
				routeMatrixMutex.Unlock()

				core.Debug("updated route matrix: %d relays", len(routeMatrix.RelayIds))
			}
		}
	}()	
}

func GetRouteMatrixAndDatabase() (*common.RouteMatrix, *routing.DatabaseBinWrapper) {
	routeMatrixMutex.Lock()
	a := routeMatrix
	b := database
	routeMatrixMutex.Unlock()
	return a,b
}
