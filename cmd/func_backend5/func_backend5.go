/*
   Network Next. You control the network.
   Copyright © 2017 - 2023 Network Next, Inc. All rights reserved.
*/

package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/mux"

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/packets"
)

var TestRouterPrivateKey = []byte{}

var TestBackendPrivateKey = []byte{}

const NEXT_RELAY_BACKEND_PORT = 30000
const NEXT_SERVER_BACKEND_PORT = 45000

const BACKEND_MODE_FORCE_DIRECT = 1
const BACKEND_MODE_RANDOM = 2
const BACKEND_MODE_MULTIPATH = 3
const BACKEND_MODE_ON_OFF = 4
const BACKEND_MODE_ON_ON_OFF = 5
const BACKEND_MODE_ROUTE_SWITCHING = 6
const BACKEND_MODE_UNCOMMITTED = 7
const BACKEND_MODE_UNCOMMITTED_TO_COMMITTED = 8
const BACKEND_MODE_SERVER_EVENTS = 9
const BACKEND_MODE_FORCE_RETRY = 10
const BACKEND_MODE_BANDWIDTH = 11
const BACKEND_MODE_JITTER = 12
const BACKEND_MODE_TAGS = 13
const BACKEND_MODE_DIRECT_STATS = 14
const BACKEND_MODE_NEXT_STATS = 15
const BACKEND_MODE_NEAR_RELAY_STATS = 16
const BACKEND_MODE_MATCH_ID = 17
const BACKEND_MODE_MATCH_VALUES = 18

type Backend struct {
	mutex        sync.RWMutex
	dirty        bool
	mode         int
	relayManager *common.RelayManager
	routeMatrix  *common.RouteMatrix
}

var backend Backend

var relayPublicKey []byte

type ServerEntry struct {
	address    *net.UDPAddr
	publicKey  []byte
	lastUpdate int64
}

type SessionCacheEntry struct {
	CustomerID                 uint64
	SessionID                  uint64
	UserHash                   uint64
	Sequence                   uint64
	RouteHash                  uint64
	OnNNSliceCounter           uint64
	CommitPending              bool
	CommitObservedSliceCounter uint8
	TimestampStart             time.Time
	TimestampExpire            time.Time
	Version                    uint8
	DirectRTT                  float64
	NextRTT                    float64
	Location                   packets.SDK5_LocationData
	Response                   []byte
}

const ThresholdRTT = 1.0
const MaxRTT = float32(100)
const MaxJitter = float32(10.0)
const MaxPacketLoss = float32(0.1)

func OptimizeThread() {

	for {

		backend.mutex.Lock()

		currentTime := time.Now().Unix()

		activeRelays := backend.relayManager.GetActiveRelays(currentTime)

		numRelays := len(activeRelays)
		relayIds := make([]uint64, numRelays)
		relayDatacenterIds := make([]uint64, numRelays)
		for i := 0; i < numRelays; i++ {
			relayIds[i] = activeRelays[i].Id
			relayDatacenterIds[i] = common.DatacenterId("local")
		}

		costMatrix := backend.relayManager.GetCosts(currentTime, relayIds, MaxRTT, MaxJitter, MaxPacketLoss, false)

		numCPUs := runtime.NumCPU()
		numSegments := numRelays
		if numCPUs < numRelays {
			numSegments = numRelays / 5
			if numSegments == 0 {
				numSegments = 1
			}
		}

		destRelays := make([]bool, numRelays)
		for i := 0; i < numRelays; i++ {
			destRelays[i] = true
		}

		core.Optimize2(numRelays, numSegments, costMatrix, 1, relayDatacenterIds, destRelays)

		backend.mutex.Unlock()

		time.Sleep(1 * time.Second)
	}
}

var (
	magicMutex    sync.RWMutex
	magicUpcoming [8]byte
	magicCurrent  [8]byte
	magicPrevious [8]byte
)

func GetMagic() ([8]byte, [8]byte, [8]byte) {
	magicMutex.RLock()
	upcoming := magicUpcoming
	current := magicCurrent
	previous := magicPrevious
	magicMutex.RUnlock()
	return upcoming, current, previous
}

func GenerateMagic(magic []byte) {
	newMagic := make([]byte, 8)
	core.RandomBytes(newMagic)
	for i := range newMagic {
		magic[i] = newMagic[i]
	}
}

func UpdateMagic() {
	ticker := time.NewTicker(time.Second * 60)
	for {
		select {
		case <-ticker.C:
			magicMutex.Lock()
			magicPrevious = magicCurrent
			magicCurrent = magicUpcoming
			GenerateMagic(magicUpcoming[:])
			magicMutex.Unlock()
		}
	}
}

func (backend *Backend) GetRelays() (relayIds []uint64, relayAddresses []net.UDPAddr) {
	backend.mutex.Lock()
	currentTime := time.Now().Unix()
	activeRelays := backend.relayManager.GetActiveRelays(currentTime)
	backend.mutex.Unlock()
	if len(activeRelays) > core.MaxNearRelays {
		activeRelays = activeRelays[:core.MaxNearRelays]
	}
	numRelays := len(activeRelays)
	relayIds = make([]uint64, numRelays)
	relayAddresses = make([]net.UDPAddr, numRelays)
	for i := range activeRelays {
		relayIds[i] = activeRelays[i].Id
		relayAddresses[i] = activeRelays[i].Address
	}
	return
}

// -----------------------------------------------------------

const UpdateRequestVersion = 5
const UpdateResponseVersion = 1
const MaxRelayAddressLength = 256
const RelayTokenBytes = 32

func RelayUpdateHandler(writer http.ResponseWriter, request *http.Request) {

	// parse the relay update request

	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		fmt.Printf("could not read body\n")
		return
	}
	defer request.Body.Close()

	index := 0

	var version uint32
	if !encoding.ReadUint32(body, &index, &version) || version != UpdateRequestVersion {
		fmt.Printf("bad version\n")
		return
	}

	var relay_address string
	if !encoding.ReadString(body, &index, &relay_address, MaxRelayAddressLength) {
		fmt.Printf("address\n")
		return
	}

	relayId := common.RelayId(relay_address)

	var token []byte
	if !encoding.ReadBytes(body, &index, &token, RelayTokenBytes) {
		fmt.Printf("bad token\n")
		return
	}

	udpAddr, err := net.ResolveUDPAddr("udp", relay_address)
	if err != nil {
		fmt.Printf("bad resolve addr %s\n", relay_address)
		return
	}

	var numSamples uint32
	if !encoding.ReadUint32(body, &index, &numSamples) {
		fmt.Printf("could not read num samples\n")
		return
	}

	sampleRelayId := make([]uint64, numSamples)
	sampleRTT := make([]float32, numSamples)
	sampleJitter := make([]float32, numSamples)
	samplePacketLoss := make([]float32, numSamples)

	for i := 0; i < int(numSamples); i++ {

		var id uint64
		var rtt, jitter, packetLoss float32

		if !encoding.ReadUint64(body, &index, &id) {
			fmt.Printf("bad relay id\n")
			return
		}

		if !encoding.ReadFloat32(body, &index, &rtt) {
			fmt.Printf("bad relay rtt\n")
			return
		}

		if !encoding.ReadFloat32(body, &index, &jitter) {
			fmt.Printf("bad relay jitter\n")
			return
		}

		if !encoding.ReadFloat32(body, &index, &packetLoss) {
			fmt.Printf("bad relay packet loss\n")
			return
		}

		sampleRelayId[i] = id
		sampleRTT[i] = rtt
		sampleJitter[i] = jitter
		samplePacketLoss[i] = packetLoss
	}

	var sessionCount uint64
	if !encoding.ReadUint64(body, &index, &sessionCount) {
		fmt.Printf("could not read session count\n")
		return
	}

	var shutdown bool
	if !encoding.ReadBool(body, &index, &shutdown) {
		fmt.Printf("could not read shutdown\n")
		return
	}

	var relayVersion string
	if !encoding.ReadString(body, &index, &relayVersion, uint32(32)) {
		fmt.Printf("could not read relay version\n")
		return
	}

	var cpu uint8
	if !encoding.ReadUint8(body, &index, &cpu) {
		fmt.Printf("could not read cpu\n")
		return
	}

	var envelopeUpKbps uint64
	if !encoding.ReadUint64(body, &index, &envelopeUpKbps) {
		fmt.Printf("could not read envelope up kbps\n")
		return
	}

	var envelopeDownKbps uint64
	if !encoding.ReadUint64(body, &index, &envelopeDownKbps) {
		fmt.Printf("could not read envelope down kbps\n")
		return
	}

	var bandwidthSentKbps uint64
	if !encoding.ReadUint64(body, &index, &bandwidthSentKbps) {
		fmt.Printf("could not read bandwidth sent kbps\n")
		return
	}

	var bandwidthRecvKbps uint64
	if !encoding.ReadUint64(body, &index, &bandwidthRecvKbps) {
		fmt.Printf("could not read bandwidth recv kbps\n")
		return
	}

	// process the relay update

	relayPort := udpAddr.Port

	relayName := fmt.Sprintf("local.%d", relayPort)

	currentTime := time.Now().Unix()

	backend.mutex.Lock()
	backend.relayManager.ProcessRelayUpdate(currentTime, relayId, relayName, *udpAddr, int(sessionCount), relayVersion, shutdown, int(numSamples), sampleRelayId, sampleRTT, sampleJitter, samplePacketLoss)
	backend.mutex.Unlock()

	// get relays to ping

	relayIds, relayAddresses := backend.GetRelays()

	numRelays := uint32(len(relayIds))

	// write response packet

	magicUpcoming, magicCurrent, magicPrevious := GetMagic()

	responseData := make([]byte, 10*1024)

	index = 0

	encoding.WriteUint32(responseData, &index, UpdateResponseVersion)

	encoding.WriteUint64(responseData, &index, uint64(time.Now().Unix()))

	encoding.WriteUint32(responseData, &index, uint32(numRelays))

	for i := range relayIds {
		encoding.WriteUint64(responseData, &index, relayIds[i])
		encoding.WriteString(responseData, &index, relayAddresses[i].String(), MaxRelayAddressLength)
	}

	encoding.WriteString(responseData, &index, relayVersion, uint32(32))

	encoding.WriteBytes(responseData, &index, magicUpcoming[:], 8)

	encoding.WriteBytes(responseData, &index, magicCurrent[:], 8)

	encoding.WriteBytes(responseData, &index, magicPrevious[:], 8)

	encoding.WriteUint32(responseData, &index, 0)

	responseLength := index

	responseData = responseData[:responseLength]

	writer.Header().Set("Content-Type", "application/octet-stream")

	writer.Write(responseData)
}

func StartWebServer() {
	router := mux.NewRouter()
	router.HandleFunc("/relay_update", RelayUpdateHandler).Methods("POST")
	http.ListenAndServe(fmt.Sprintf(":%d", NEXT_RELAY_BACKEND_PORT), router)
}

func StartUDPServer() {

	serverBackendAddress = *core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", NEXT_SERVER_BACKEND_PORT))

	config := common.UDPServerConfig{}
	config.Port = NEXT_SERVER_BACKEND_PORT
	config.NumThreads = 2
	config.SocketReadBuffer = 1024 * 1024
	config.SocketWriteBuffer = 1024 * 1024
	config.MaxPacketSize = 4096

	common.CreateUDPServer(context.Background(), config, packetHandler)
}

// -------------------------------------------------

var serverBackendAddress net.UDPAddr

func packetHandler(conn *net.UDPConn, from *net.UDPAddr, packetData []byte) {

	// ignore packets that are too small

	if len(packetData) < packets.SDK5_MinPacketBytes {
		fmt.Printf("packet is too small")
		return
	}

	// ignore packet types we don't support

	packetType := packetData[0]

	if packetType != packets.SDK5_SERVER_INIT_REQUEST_PACKET && packetType != packets.SDK5_SERVER_UPDATE_REQUEST_PACKET && packetType != packets.SDK5_SESSION_UPDATE_REQUEST_PACKET && packetType != packets.SDK5_MATCH_DATA_REQUEST_PACKET {
		fmt.Printf("unsupported packet type %d", packetType)
		return
	}

	// check basic packet filetr

	if !core.BasicPacketFilter(packetData, len(packetData)) {
		fmt.Printf("basic packet filter failed\n")
		return
	}

	// check advanced packet filter

	var emptyMagic [8]byte

	var fromAddressBuffer [32]byte
	var toAddressBuffer [32]byte

	fromAddressData, fromAddressPort := core.GetAddressData(from, fromAddressBuffer[:])
	toAddressData, toAddressPort := core.GetAddressData(&serverBackendAddress, toAddressBuffer[:])

	if !core.AdvancedPacketFilter(packetData, emptyMagic[:], fromAddressData, fromAddressPort, toAddressData, toAddressPort, len(packetData)) {
		fmt.Printf("advanced packet filter failed\n")
		return
	}

	// process the packet according to type

	packetData = packetData[16:]

	switch packetType {

	case packets.SDK5_SERVER_INIT_REQUEST_PACKET:
		packet := packets.SDK5_ServerInitRequestPacket{}
		if err := packets.ReadPacket(packetData, &packet); err != nil {
			core.Error("could not read server init request packet from %s: %v", from.String(), err)
			return
		}
		ProcessServerInitRequestPacket(conn, from, &packet)
		break

	case packets.SDK5_SERVER_UPDATE_REQUEST_PACKET:
		packet := packets.SDK5_ServerUpdateRequestPacket{}
		if err := packets.ReadPacket(packetData, &packet); err != nil {
			core.Error("could not read server update request packet from %s: %v", from.String(), err)
			return
		}
		ProcessServerUpdateRequestPacket(conn, from, &packet)
		break

	case packets.SDK5_SESSION_UPDATE_REQUEST_PACKET:
		packet := packets.SDK5_SessionUpdateRequestPacket{}
		if err := packets.ReadPacket(packetData, &packet); err != nil {
			core.Error("could not read session update request packet from %s: %v", from.String(), err)
			return
		}
		ProcessSessionUpdateRequestPacket(conn, from, &packet)
		break

	case packets.SDK5_MATCH_DATA_REQUEST_PACKET:
		packet := packets.SDK5_MatchDataRequestPacket{}
		if err := packets.ReadPacket(packetData, &packet); err != nil {
			core.Error("could not read match data request packet from %s: %v", from.String(), err)
			return
		}
		ProcessMatchDataRequestPacket(conn, from, &packet)
		break

	default:
		panic("unknown packet type")
	}
}

func SendResponsePacket[P packets.Packet](conn *net.UDPConn, to *net.UDPAddr, packetType int, packet P) {

	packetData, err := packets.SDK5_WritePacket(packet, packetType, 4096, &serverBackendAddress, to, TestBackendPrivateKey)
	if err != nil {
		core.Error("failed to write response packet: %v", err)
		return
	}

	if _, err := conn.WriteToUDP(packetData, to); err != nil {
		core.Error("failed to send response packet: %v", err)
		return
	}
}

func ProcessServerInitRequestPacket(conn *net.UDPConn, from *net.UDPAddr, requestPacket *packets.SDK5_ServerInitRequestPacket) {

	fmt.Printf("server init request from %s\n", from.String())

	responsePacket := &packets.SDK5_ServerInitResponsePacket{
		RequestId: requestPacket.RequestId,
		Response:  packets.SDK5_ServerInitResponseOK,
	}

	responsePacket.UpcomingMagic, responsePacket.CurrentMagic, responsePacket.PreviousMagic = GetMagic()

	SendResponsePacket(conn, from, packets.SDK5_SERVER_INIT_RESPONSE_PACKET, responsePacket)
}

func ProcessServerUpdateRequestPacket(conn *net.UDPConn, from *net.UDPAddr, requestPacket *packets.SDK5_ServerUpdateRequestPacket) {

	fmt.Printf("server update request from %s\n", from.String())

	responsePacket := &packets.SDK5_ServerUpdateResponsePacket{
		RequestId: requestPacket.RequestId,
	}

	responsePacket.UpcomingMagic, responsePacket.CurrentMagic, responsePacket.PreviousMagic = GetMagic()

	SendResponsePacket(conn, from, packets.SDK5_SERVER_UPDATE_RESPONSE_PACKET, responsePacket)
}

func ProcessSessionUpdateRequestPacket(conn *net.UDPConn, from *net.UDPAddr, requestPacket *packets.SDK5_SessionUpdateRequestPacket) {

	fmt.Printf("session update request from %s\n", from.String())

	if backend.mode == BACKEND_MODE_FORCE_RETRY && requestPacket.RetryNumber < 4 {
		return
	}

	if requestPacket.PlatformType == packets.SDK5_PlatformTypeUnknown {
		panic("platform type is unknown")
	}

	if requestPacket.ConnectionType == packets.SDK5_ConnectionTypeUnknown {
		panic("connection type is unknown")
	}

	if requestPacket.FallbackToDirect {
		fmt.Printf("error: fallback to direct %s\n", from.String())
		return
	}

	if requestPacket.Reported {
		fmt.Printf("client reported session\n")
	}

	if requestPacket.ClientPingTimedOut {
		fmt.Printf("client ping timed out\n")
	}

	if requestPacket.ClientNextBandwidthOverLimit {
		fmt.Printf("client next bandwidth over limit\n")
	}

	if requestPacket.ServerNextBandwidthOverLimit {
		fmt.Printf("server next bandwidth over limit\n")
	}

	if requestPacket.PacketsLostClientToServer > 0 {
		fmt.Printf("%d client to server packets lost\n", requestPacket.PacketsLostClientToServer)
	}

	if requestPacket.PacketsLostServerToClient > 0 {
		fmt.Printf("%d server to client packets lost\n", requestPacket.PacketsLostServerToClient)
	}

	if backend.mode == BACKEND_MODE_BANDWIDTH {
		if requestPacket.DirectKbpsUp > 0 {
			fmt.Printf("%d direct kbps up\n", requestPacket.DirectKbpsUp)
		}
		if requestPacket.DirectKbpsDown > 0 {
			fmt.Printf("%d direct kbps down\n", requestPacket.DirectKbpsDown)
		}
		if requestPacket.NextKbpsUp > 0 {
			fmt.Printf("%d next kbps up\n", requestPacket.NextKbpsUp)
		}
		if requestPacket.NextKbpsDown > 0 {
			fmt.Printf("%d next kbps down\n", requestPacket.NextKbpsDown)
		}
	}

	if backend.mode == BACKEND_MODE_JITTER {
		if requestPacket.JitterClientToServer > 0 {
			fmt.Printf("%f jitter up\n", requestPacket.JitterClientToServer)
			if requestPacket.JitterClientToServer > 100 {
				panic("jitter up too high")
			}
		}
		if requestPacket.JitterServerToClient > 0 {
			fmt.Printf("%f jitter down\n", requestPacket.JitterServerToClient)
			if requestPacket.JitterServerToClient > 100 {
				panic("jitter down too high")
			}
		}
	}

	if backend.mode == BACKEND_MODE_TAGS {
		if requestPacket.NumTags > 0 {
			for i := 0; i < int(requestPacket.NumTags); i++ {
				fmt.Printf("tag %x\n", requestPacket.Tags[i])
			}
		} else {
			fmt.Printf("tag cleared\n")
		}
	}

	if backend.mode == BACKEND_MODE_DIRECT_STATS {
		if requestPacket.DirectRTT > 0 && requestPacket.DirectJitter > 0 && requestPacket.DirectPacketLoss > 0 {
			fmt.Printf("direct rtt = %f, direct jitter = %f, direct packet loss = %f\n", requestPacket.DirectRTT, requestPacket.DirectJitter, requestPacket.DirectPacketLoss)
		}
	}

	if backend.mode == BACKEND_MODE_NEXT_STATS {
		if requestPacket.NextRTT > 0 && requestPacket.NextJitter > 0 && requestPacket.NextPacketLoss > 0 {
			fmt.Printf("next rtt = %f, next jitter = %f, next packet loss = %f\n", requestPacket.NextRTT, requestPacket.NextJitter, requestPacket.NextPacketLoss)
		}
	}

	if backend.mode == BACKEND_MODE_NEAR_RELAY_STATS {
		for i := 0; i <= int(requestPacket.NumNearRelays); i++ {
			fmt.Printf("near relay: id = %x, rtt = %d, jitter = %d, packet loss = %d\n", requestPacket.NearRelayIds[i], requestPacket.NearRelayRTT[i], requestPacket.NearRelayJitter[i], requestPacket.NearRelayPacketLoss[i])
		}
	}

	// read the session data

	newSession := requestPacket.SliceNumber == 0

	var sessionData packets.SDK5_SessionData

	if newSession {

		sessionData.Version = packets.SDK5_SessionDataVersion_Write
		sessionData.SessionId = requestPacket.SessionId
		sessionData.SliceNumber = uint32(requestPacket.SliceNumber + 1)
		sessionData.ExpireTimestamp = uint64(time.Now().Unix()) + packets.SDK5_BillingSliceSeconds
		sessionData.RouteState.UserID = requestPacket.UserHash
		sessionData.Location.Version = packets.SDK5_LocationVersion_Write

	} else {

		readStream := encoding.CreateReadStream(requestPacket.SessionData[:])

		err := sessionData.Serialize(readStream)
		if err != nil {
			fmt.Printf("could not read session data in session update packet: %v\n", err)
			return
		}

		sessionData.SliceNumber = uint32(requestPacket.SliceNumber + 1)
		sessionData.ExpireTimestamp += packets.SDK5_BillingSliceSeconds
	}

	// get data about all active relays on the relay backend

	relayIds, relayAddresses := backend.GetRelays()

	numRelays := len(relayIds)

	// decide if we should take network next or not

	takeNetworkNext := numRelays > 0

	if backend.mode == BACKEND_MODE_FORCE_DIRECT {
		takeNetworkNext = false
	}

	if backend.mode == BACKEND_MODE_RANDOM {
		takeNetworkNext = takeNetworkNext && rand.Float32() > 0.5
	}

	if backend.mode == BACKEND_MODE_ON_OFF {
		// Alternate between direct and next routes for each slice
		if (requestPacket.SliceNumber & 1) == 0 {
			takeNetworkNext = false
		}
	}

	if backend.mode == BACKEND_MODE_ON_ON_OFF {
		// Alternate between direct, a new route token and a continue route token for every 3 slices
		if (requestPacket.SliceNumber & 2) == 0 {
			takeNetworkNext = false
		}
	}

	if backend.mode == BACKEND_MODE_ROUTE_SWITCHING {
		rand.Shuffle(numRelays, func(i, j int) {
			relayIds[i], relayIds[j] = relayIds[j], relayIds[i]
			relayAddresses[i], relayAddresses[j] = relayAddresses[j], relayAddresses[i]
		})
	}

	// run various checks and prints for special func test modes

	multipath := len(relayIds) > 0 && backend.mode == BACKEND_MODE_MULTIPATH

	if backend.mode == BACKEND_MODE_SERVER_EVENTS {
		if requestPacket.SliceNumber >= 2 && requestPacket.ServerEvents != 0x123 {
			panic("server events not set on session update")
		}
	}

	if requestPacket.ServerEvents > 0 {
		fmt.Printf("server events %x\n", requestPacket.ServerEvents)
	}

	// build response packet

	var responsePacket *packets.SDK5_SessionUpdateResponsePacket

	if !takeNetworkNext {

		// direct route

		responsePacket = &packets.SDK5_SessionUpdateResponsePacket{
			SessionId:          requestPacket.SessionId,
			SliceNumber:        requestPacket.SliceNumber,
			NumNearRelays:      int32(numRelays),
			RouteType:          int32(packets.SDK5_RouteTypeDirect),
			NumTokens:          0,
			Tokens:             nil,
			HighFrequencyPings: false,
		}

		for i := 0; i < numRelays; i++ {
			responsePacket.NearRelayIds[i] = relayIds[i]
			responsePacket.NearRelayAddresses[i] = relayAddresses[i]
		}

	} else {

		// next

		const MaxRouteRelays = packets.SDK5_MaxRelaysPerRoute

		var routeRelayIds [MaxRouteRelays]uint64
		var routeRelayAddresses [MaxRouteRelays]net.UDPAddr
		var routeRelayPublicKeys [MaxRouteRelays][]byte

		numRouteRelays := numRelays
		if numRouteRelays > MaxRouteRelays {
			numRouteRelays = MaxRouteRelays
		}

		for i := 0; i < numRouteRelays; i++ {
			routeRelayIds[i] = relayIds[i]
			routeRelayAddresses[i] = relayAddresses[i]
			routeRelayPublicKeys[i] = relayPublicKey
		}

		// is this a continue route, or a new route?

		var routeType int32

		sameRoute := numRouteRelays == int(sessionData.RouteNumRelays) && routeRelayIds == sessionData.RouteRelayIds

		// build token data

		routerPrivateKey := [packets.SDK5_PrivateKeyBytes]byte{}
		copy(routerPrivateKey[:], TestRouterPrivateKey)

		tokenAddresses := make([]*net.UDPAddr, numRouteRelays+2)
		tokenAddresses[0] = &requestPacket.ClientAddress
		tokenAddresses[len(tokenAddresses)-1] = &requestPacket.ServerAddress
		for i := 0; i < numRouteRelays; i++ {
			tokenAddresses[1+i] = &relayAddresses[i]
		}

		tokenPublicKeys := make([][]byte, numRouteRelays+2)
		tokenPublicKeys[0] = requestPacket.ClientRoutePublicKey[:]
		tokenPublicKeys[len(tokenPublicKeys)-1] = requestPacket.ServerRoutePublicKey[:]
		for i := 0; i < numRouteRelays; i++ {
			tokenPublicKeys[1+i] = relayPublicKey
		}

		numTokens := numRouteRelays + 2

		var tokenData []byte

		if sameRoute {
			tokenData = make([]byte, numTokens*packets.SDK5_EncryptedContinueRouteTokenSize)
			core.WriteContinueTokens(tokenData, sessionData.ExpireTimestamp, sessionData.SessionId, uint8(sessionData.SessionVersion), int(numTokens), tokenPublicKeys, routerPrivateKey[:])
			routeType = packets.SDK5_RouteTypeContinue
		} else {
			sessionData.ExpireTimestamp += packets.SDK5_BillingSliceSeconds
			sessionData.SessionVersion++

			tokenData = make([]byte, numTokens*packets.SDK5_EncryptedNextRouteTokenSize)
			core.WriteRouteTokens(tokenData, sessionData.ExpireTimestamp, sessionData.SessionId, uint8(sessionData.SessionVersion), 256, 256, int(numTokens), tokenAddresses, tokenPublicKeys, routerPrivateKey[:])
			routeType = packets.SDK5_RouteTypeNew
		}

		// contruct the session update response packet

		responsePacket = &packets.SDK5_SessionUpdateResponsePacket{
			SessionId:          requestPacket.SessionId,
			SliceNumber:        requestPacket.SliceNumber,
			RouteType:          routeType,
			Multipath:          multipath,
			NumTokens:          int32(numTokens),
			Tokens:             tokenData,
			HighFrequencyPings: false,
		}

		if numRelays > packets.SDK5_MaxNearRelays {
			numRelays = packets.SDK5_MaxNearRelays
		}

		responsePacket.NumNearRelays = int32(numRelays)

		for i := 0; i < numRelays; i++ {
			responsePacket.NearRelayIds[i] = relayIds[i]
			responsePacket.NearRelayAddresses[i] = relayAddresses[i]
		}
	}

	if responsePacket == nil {
		fmt.Printf("error: nil session response\n")
		return
	}

	packetSessionData, err := packets.WritePacket[*packets.SDK5_SessionData](responsePacket.SessionData[:], &sessionData)

	if err != nil {
		fmt.Printf("error: failed to write session data\n")
		return
	}

	responsePacket.SessionDataBytes = int32(len(packetSessionData))

	SendResponsePacket(conn, from, packets.SDK5_SESSION_UPDATE_RESPONSE_PACKET, responsePacket)
}

func ProcessMatchDataRequestPacket(conn *net.UDPConn, from *net.UDPAddr, requestPacket *packets.SDK5_MatchDataRequestPacket) {

	fmt.Printf("server match data request\n")

	if backend.mode == BACKEND_MODE_FORCE_RETRY && requestPacket.RetryNumber < 4 {
		fmt.Printf("force retry for match data request packet\n")
		return
	}

	if backend.mode == BACKEND_MODE_MATCH_ID || backend.mode == BACKEND_MODE_FORCE_RETRY {
		fmt.Printf("match id %x\n", requestPacket.MatchId)
	}

	if backend.mode == BACKEND_MODE_MATCH_VALUES || backend.mode == BACKEND_MODE_FORCE_RETRY {
		for i := 0; i < int(requestPacket.NumMatchValues); i++ {
			fmt.Printf("match value %.2f\n", requestPacket.MatchValues[i])
		}
	}

	responsePacket := &packets.SDK5_MatchDataResponsePacket{
		SessionId: requestPacket.SessionId,
	}

	SendResponsePacket(conn, from, packets.SDK5_MATCH_DATA_RESPONSE_PACKET, responsePacket)
}

// -----------------------------------------------

func main() {

	rand.Seed(time.Now().UnixNano())

	backend.relayManager = common.CreateRelayManager()

	backend.routeMatrix = &common.RouteMatrix{}

	if os.Getenv("BACKEND_MODE") == "FORCE_DIRECT" {
		backend.mode = BACKEND_MODE_FORCE_DIRECT
	}

	if os.Getenv("BACKEND_MODE") == "RANDOM" {
		backend.mode = BACKEND_MODE_RANDOM
	}

	if os.Getenv("BACKEND_MODE") == "MULTIPATH" {
		backend.mode = BACKEND_MODE_MULTIPATH
	}

	if os.Getenv("BACKEND_MODE") == "ON_OFF" {
		backend.mode = BACKEND_MODE_ON_OFF
	}

	if os.Getenv("BACKEND_MODE") == "ON_ON_OFF" {
		backend.mode = BACKEND_MODE_ON_ON_OFF
	}

	if os.Getenv("BACKEND_MODE") == "ROUTE_SWITCHING" {
		backend.mode = BACKEND_MODE_ROUTE_SWITCHING
	}

	if os.Getenv("BACKEND_MODE") == "UNCOMMITTED" {
		backend.mode = BACKEND_MODE_UNCOMMITTED
	}

	if os.Getenv("BACKEND_MODE") == "UNCOMMITTED_TO_COMMITTED" {
		backend.mode = BACKEND_MODE_UNCOMMITTED_TO_COMMITTED
	}

	if os.Getenv("BACKEND_MODE") == "SERVER_EVENTS" {
		backend.mode = BACKEND_MODE_SERVER_EVENTS
	}

	if os.Getenv("BACKEND_MODE") == "FORCE_RETRY" {
		backend.mode = BACKEND_MODE_FORCE_RETRY
	}

	if os.Getenv("BACKEND_MODE") == "BANDWIDTH" {
		backend.mode = BACKEND_MODE_BANDWIDTH
	}

	if os.Getenv("BACKEND_MODE") == "JITTER" {
		backend.mode = BACKEND_MODE_JITTER
	}

	if os.Getenv("BACKEND_MODE") == "TAGS" {
		backend.mode = BACKEND_MODE_TAGS
	}

	if os.Getenv("BACKEND_MODE") == "DIRECT_STATS" {
		backend.mode = BACKEND_MODE_DIRECT_STATS
	}

	if os.Getenv("BACKEND_MODE") == "NEXT_STATS" {
		backend.mode = BACKEND_MODE_NEXT_STATS
	}

	if os.Getenv("BACKEND_MODE") == "MATCH_ID" {
		backend.mode = BACKEND_MODE_MATCH_ID
	}

	if os.Getenv("BACKEND_MODE") == "MATCH_VALUES" {
		backend.mode = BACKEND_MODE_MATCH_VALUES
	}

	TestRouterPrivateKey = envvar.GetBase64("TEST_ROUTER_PRIVATE_KEY", []byte{})

	TestBackendPrivateKey = envvar.GetBase64("TEST_BACKEND_PRIVATE_KEY", []byte{})

	relayPublicKey, _ = base64.StdEncoding.DecodeString("9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=")

	GenerateMagic(magicUpcoming[:])
	GenerateMagic(magicCurrent[:])
	GenerateMagic(magicPrevious[:])

	go OptimizeThread()

	go UpdateMagic()

	go StartWebServer()

	go StartUDPServer()

	fmt.Printf("started functional backend on ports %d and %d (sdk5)\n", NEXT_RELAY_BACKEND_PORT, NEXT_SERVER_BACKEND_PORT)

	// Wait for shutdown signal
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-termChan:
		core.Debug("received shutdown signal")
		return
	}
}
