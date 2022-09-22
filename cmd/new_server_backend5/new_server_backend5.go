package main

import (
	"net"
	"time"

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/packets"
)

var serverBackendAddress net.UDPAddr
var routeMatrixURI string
var routeMatrixInterval time.Duration

func main() {

	service := common.CreateService("new_server_backend5")

	serverBackendAddress = *envvar.GetAddress("SERVER_BACKEND_ADDRESS", core.ParseAddress("127.0.0.1:45000"))
	routeMatrixURI = envvar.GetString("ROUTE_MATRIX_URI", "http://127.0.0.1:30001/route_matrix")
	routeMatrixInterval = envvar.GetDuration("ROUTE_MATRIX_INTERVAL", time.Second)

	core.Log("server backend address: %s", serverBackendAddress.String())
	core.Log("route matrix uri: %s", routeMatrixURI)
	core.Log("route matrix interval: %s", routeMatrixInterval.String())

	service.StartUDPServer(packetHandler)

	service.StartWebServer()

	service.UpdateMagic()

	service.WaitForShutdown()
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

	// check packet signature

	if !CheckPacketSignature(packetData) {
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

func CheckPacketSignature(packetData []byte) bool {

	var buyerId uint64
	index := 16 + 3
	common.ReadUint64(packetData, &index, &buyerId)

	core.Debug("signature buyer id is %016x", buyerId)

	// todo: look up buyer in database

	// todo: get buyer public key

	// todo: run signature check on packet with buyer public key

	return true
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

	maxBytes := envvar.GetInt("COST_MATRIX_STATS_MESSAGE_MAX_BYTES", 1024)

	core.Log("cost matrix stats message max bytes: %d", maxBytes)

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

				if !service.IsLeader() {
					break
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

				routeMatrix := common.RouteMatrix{}

				err = routeMatrix.Read(buffer)
				if err != nil {
					core.Error("failed to read route matrix: %v", err)
					continue
				}

				logMutex.Lock()

				routeMatrixBytes := len(buffer)
				routeMatrixNumRelays := len(routeMatrix.RelayIds)

				routeMatrixNumDestRelays := 0
				for i := range routeMatrix.DestRelays {
					if routeMatrix.DestRelays[i] {
						routeMatrixNumDestRelays++
					}
				}

				datacenterMap := make(map[uint64]bool)
				for i := range routeMatrix.RelayDatacenterIds {
					datacenterMap[routeMatrix.RelayDatacenterIds[i]] = true
				}
				routeMatrixNumDatacenters := len(datacenterMap)

				routeMatrixNumFullRelays := len(routeMatrix.FullRelayIds)

				analysis := routeMatrix.Analyze()

				core.Debug("---------------------------------------------")

				core.Debug("route matrix bytes: %d", routeMatrixBytes)

				core.Debug("route matrix num relays: %d", routeMatrixNumRelays)
				core.Debug("route matrix num dest relays: %d", routeMatrixNumDestRelays)
				core.Debug("route matrix num full relays: %d", routeMatrixNumFullRelays)
				core.Debug("route matrix num datacenters: %d", routeMatrixNumDatacenters)

				core.Debug("route matrix total routes: %d", analysis.TotalRoutes)
				core.Debug("route matrix average num routes: %.1f", analysis.AverageNumRoutes)
				core.Debug("route matrix average route length: %.1f", analysis.AverageRouteLength)
				core.Debug("no route percent: %.1f%%", analysis.NoRoutePercent)
				core.Debug("one route percent: %.1f%%", analysis.OneRoutePercent)
				core.Debug("no direct route percent: %.1f%%", analysis.NoDirectRoutePercent)

				core.Debug("route matrix rtt bucket no improvement: %.1f%%", analysis.RTTBucket_NoImprovement)
				core.Debug("route matrix rtt bucket 0-5ms: %.1f%%", analysis.RTTBucket_0_5ms)
				core.Debug("route matrix rtt bucket 5-10ms: %.1f%%", analysis.RTTBucket_5_10ms)
				core.Debug("route matrix rtt bucket 10-15ms: %.1f%%", analysis.RTTBucket_10_15ms)
				core.Debug("route matrix rtt bucket 15-20ms: %.1f%%", analysis.RTTBucket_15_20ms)
				core.Debug("route matrix rtt bucket 20-25ms: %.1f%%", analysis.RTTBucket_20_25ms)
				core.Debug("route matrix rtt bucket 25-30ms: %.1f%%", analysis.RTTBucket_25_30ms)
				core.Debug("route matrix rtt bucket 30-35ms: %.1f%%", analysis.RTTBucket_30_35ms)
				core.Debug("route matrix rtt bucket 35-40ms: %.1f%%", analysis.RTTBucket_35_40ms)
				core.Debug("route matrix rtt bucket 40-45ms: %.1f%%", analysis.RTTBucket_40_45ms)
				core.Debug("route matrix rtt bucket 45-50ms: %.1f%%", analysis.RTTBucket_45_50ms)
				core.Debug("route matrix rtt bucket 50ms+: %.1f%%", analysis.RTTBucket_50ms_Plus)

				totalPercent := analysis.RTTBucket_NoImprovement +
					analysis.RTTBucket_0_5ms +
					analysis.RTTBucket_5_10ms +
					analysis.RTTBucket_10_15ms +
					analysis.RTTBucket_15_20ms +
					analysis.RTTBucket_20_25ms +
					analysis.RTTBucket_25_30ms +
					analysis.RTTBucket_30_35ms +
					analysis.RTTBucket_35_40ms +
					analysis.RTTBucket_40_45ms +
					analysis.RTTBucket_45_50ms +
					analysis.RTTBucket_50ms_Plus

				core.Debug("route matrix rtt bucket total percent: %.1f%%", totalPercent)

				core.Debug("---------------------------------------------")

				logMutex.Unlock()

				// send route matrix stats via pubsub

				routeMatrixStatsEntry := messages.RouteMatrixStatsMessage{}

				routeMatrixStatsEntry.Version = messages.RouteMatrixStatsMessageVersion
				routeMatrixStatsEntry.Bytes = routeMatrixBytes
				routeMatrixStatsEntry.NumRelays = routeMatrixNumRelays
				routeMatrixStatsEntry.NumDestRelays = routeMatrixNumDestRelays
				routeMatrixStatsEntry.NumDatacenters = routeMatrixNumDatacenters

				message := routeMatrixStatsEntry.Write(make([]byte, maxBytes))

				statsPubsubProducer.MessageChannel <- message
			}
		}
	}()	
}