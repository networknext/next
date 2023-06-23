/*
   Network Next Accelerate.
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
	"bytes"

	"github.com/gorilla/mux"

	"github.com/networknext/accelerate/modules/common"
	"github.com/networknext/accelerate/modules/constants"
	"github.com/networknext/accelerate/modules/core"
	"github.com/networknext/accelerate/modules/crypto"
	"github.com/networknext/accelerate/modules/encoding"
	"github.com/networknext/accelerate/modules/packets"
)

func Base64String(value string) []byte {
	data, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		panic(err)
	}
	return data
}

var TestRelayPublicKey = Base64String("9SKtwe4Ear59iQyBOggxutzdtVLLc1YQ2qnArgiiz14=")
var TestRelayPrivateKey = Base64String("lypnDfozGRHepukundjYAF5fKY1Tw2g7Dxh0rAgMCt8=")
var TestRelayBackendPublicKey = Base64String("SS55dEl9nTSnVVDrqwPeqRv/YcYOZZLXCWTpNBIyX0Y=")
var TestRelayBackendPrivateKey = Base64String("ls5XiwAZRCfyuZAbQ1b9T1bh2VZY8vQ7hp8SdSTSR7M=")
var TestServerBackendPublicKey = Base64String("TGHKjEeHPtSgtZfDyuDPcQgtJTyRDtRvGSKvuiWWo0A=")
var TestServerBackendPrivateKey = Base64String("FXwFqzjGlIwUDwiq1N5Um5VUesdr4fP2hVV2cnJ+yARMYcqMR4c+1KC1l8PK4M9xCC0lPJEO1G8ZIq+6JZajQA==")

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
const BACKEND_MODE_DIRECT_STATS = 13
const BACKEND_MODE_NEXT_STATS = 14
const BACKEND_MODE_NEAR_RELAY_STATS = 15
const BACKEND_MODE_MATCH_ID = 16
const BACKEND_MODE_MATCH_VALUES = 17

type Backend struct {
	mutex        sync.RWMutex
	dirty        bool
	mode         int
	relayManager *common.RelayManager
}

var backend Backend

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
	Latitude                   float32
	Longitude                  float32
	Response                   []byte
}

const ThresholdRTT = 1.0
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

		costMatrix := backend.relayManager.GetCosts(currentTime, relayIds, MaxJitter, MaxPacketLoss)

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

		core.Optimize2(numRelays, numSegments, costMatrix, relayDatacenterIds, destRelays)

		backend.mutex.Unlock()

		time.Sleep(1 * time.Second)
	}
}

var (
	magicMutex    sync.RWMutex
	magicUpcoming [constants.MagicBytes]byte
	magicCurrent  [constants.MagicBytes]byte
	magicPrevious [constants.MagicBytes]byte
)

func GetMagic() ([constants.MagicBytes]byte, [constants.MagicBytes]byte, [constants.MagicBytes]byte) {
	magicMutex.RLock()
	upcoming := magicUpcoming
	current := magicCurrent
	previous := magicPrevious
	magicMutex.RUnlock()
	return upcoming, current, previous
}

func GenerateMagic(magic []byte) {
	newMagic := make([]byte, constants.MagicBytes)
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
	if len(activeRelays) > constants.MaxNearRelays {
		activeRelays = activeRelays[:constants.MaxNearRelays]
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

	body, err := ioutil.ReadAll(request.Body)

	if err != nil {
		fmt.Printf("could not read body\n")
		return
	}

	defer request.Body.Close()

	// ignore the relay update if it's too small to be valid

	packetBytes := len(body)

	if packetBytes < 1+1+4+2+crypto.Box_MacSize+crypto.Box_NonceSize {
		core.Error("relay update packet is too small to be valid")
		writer.WriteHeader(http.StatusBadRequest) // 400
		return
	}

	// read the version and decide if we can handle it

	index := 0
	packetData := body
	var packetVersion uint8
	encoding.ReadUint8(packetData, &index, &packetVersion)

	if packetVersion < packets.RelayUpdateRequestPacket_VersionMin || packetVersion > packets.RelayUpdateRequestPacket_VersionMax {
		core.Error("invalid relay update packet version: %d", request.RemoteAddr, packetVersion)
		writer.WriteHeader(http.StatusBadRequest) // 400
		return
	}

	// read the relay address

	var relayAddress net.UDPAddr
	if !encoding.ReadAddress(packetData, &index, &relayAddress) {
		core.Debug("could not read relay address")
		writer.WriteHeader(http.StatusBadRequest) // 400
		return
	}

	// decrypt the relay update

	nonce := packetData[packetBytes-crypto.Box_NonceSize:]

	encryptedData := packetData[index : packetBytes-crypto.Box_NonceSize]
	encryptedBytes := len(encryptedData)

	err = crypto.Box_Decrypt(TestRelayPublicKey, TestRelayBackendPrivateKey, nonce, encryptedData, encryptedBytes)
	if err != nil {
		core.Debug("[%s] failed to decrypt relay update", request.RemoteAddr)
		writer.WriteHeader(http.StatusBadRequest) // 400
		return
	}

	// read the relay update request packet

	var requestPacket packets.RelayUpdateRequestPacket

	err = requestPacket.Read(body)
	if err != nil {
		core.Error("could not read relay update: %v", err)
		writer.WriteHeader(http.StatusBadRequest) // 400
		return
	}

	currentTimestamp := uint64(time.Now().Unix())

	if requestPacket.CurrentTime < currentTimestamp-10 {
		core.Error("relay update is too old")
		writer.WriteHeader(http.StatusBadRequest) // 400
		return
	}

	if requestPacket.CurrentTime > currentTimestamp+10 {
		core.Error("relay update is in the future")
		writer.WriteHeader(http.StatusBadRequest) // 400
		return
	}

	// process the relay update

	relayId := common.RelayId(relayAddress.String())

	relayPort := relayAddress.Port

	relayName := fmt.Sprintf("local.%d", relayPort)

	numSamples := requestPacket.NumSamples

	currentTime := time.Now().Unix()

	backend.mutex.Lock()

	backend.relayManager.ProcessRelayUpdate(currentTime,
		relayId,
		relayName,
		requestPacket.Address,
		int(requestPacket.SessionCount),
		requestPacket.RelayVersion,
		requestPacket.RelayFlags,
		int(numSamples),
		requestPacket.SampleRelayId[:numSamples],
		requestPacket.SampleRTT[:numSamples],
		requestPacket.SampleJitter[:numSamples],
		requestPacket.SamplePacketLoss[:numSamples],
		requestPacket.RelayCounters[:],
	)

	backend.mutex.Unlock()

	// build response packet

	var responsePacket packets.RelayUpdateResponsePacket

	responsePacket.Version = packets.RelayUpdateResponsePacket_VersionWrite
	responsePacket.Timestamp = uint64(time.Now().Unix())
	responsePacket.TargetVersion = "func test"

	relayIds, relayAddresses := backend.GetRelays()

	relayIndex := 0

	for i := range relayIds {

		if relayIds[i] == relayId {
			continue
		}

		address := relayAddresses[i]

		responsePacket.RelayId[relayIndex] = relayIds[i]
		responsePacket.RelayAddress[relayIndex] = address

		relayIndex++
	}

	responsePacket.NumRelays = uint32(relayIndex)

	responsePacket.UpcomingMagic, responsePacket.CurrentMagic, responsePacket.PreviousMagic = GetMagic()

	responsePacket.ExpectedPublicAddress = relayAddress

	copy(responsePacket.ExpectedRelayPublicKey[:], TestRelayPublicKey)
	copy(responsePacket.ExpectedRelayBackendPublicKey[:], TestRelayBackendPublicKey)

	token := core.RouteToken{}
	core.WriteEncryptedRouteToken(&token, responsePacket.TestToken[:], TestRelayBackendPrivateKey, TestRelayPublicKey)

	// send the response packet

	responseData := make([]byte, 1024*1024)

	responseData = responsePacket.Write(responseData)

	writer.Header().Set("Content-Type", request.Header.Get("Content-Type"))

	writer.Write(responseData)
}

func CostMatrixHandler(writer http.ResponseWriter, request *http.Request) {

	backend.mutex.Lock()

	currentTime := time.Now().Unix()

	activeRelays := backend.relayManager.GetActiveRelays(currentTime)

	relayIds := make([]uint64, len(activeRelays))
	relayAddresses := make([]net.UDPAddr, len(activeRelays))
	relayNames := make([]string, len(activeRelays))
	relayLatitudes := make([]float32, len(activeRelays))
	relayLongitudes := make([]float32, len(activeRelays))
	relayDatacenterIds := make([]uint64, len(activeRelays))
	destRelays := make([]bool, len(activeRelays))

	for i := range activeRelays {
		relayIds[i] = activeRelays[i].Id
		relayAddresses[i] = activeRelays[i].Address
		relayNames[i] = activeRelays[i].Name
		destRelays[i] = true
	}

	costs := backend.relayManager.GetCosts(currentTime, relayIds, 100.0, 0.1)

	costMatrix := &common.CostMatrix{
		Version:            common.CostMatrixVersion_Write,
		RelayIds:           relayIds,
		RelayAddresses:     relayAddresses,
		RelayNames:         relayNames,
		RelayLatitudes:     relayLatitudes,
		RelayLongitudes:    relayLongitudes,
		RelayDatacenterIds: relayDatacenterIds,
		DestRelays:         destRelays,
		Costs:              costs,
	}

	costMatrixData, err := costMatrix.Write()
	if err != nil {
		core.Error("could not write cost matrix: %v", err)
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	backend.mutex.Unlock()

	writer.Header().Set("Content-Type", "application/octet-stream")

	buffer := bytes.NewBuffer(costMatrixData)
	_, err = buffer.WriteTo(writer)
	if err != nil {
		return
	}
}

func StartWebServer() {
	router := mux.NewRouter()
	router.HandleFunc("/relay_update", RelayUpdateHandler).Methods("POST")
	router.HandleFunc("/cost_matrix", CostMatrixHandler).Methods("GET")
	http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", NEXT_RELAY_BACKEND_PORT), router)
}

func StartUDPServer() {

	bindAddress := core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", NEXT_SERVER_BACKEND_PORT))

	serverBackendAddress = core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", NEXT_SERVER_BACKEND_PORT))

	config := common.UDPServerConfig{}
	config.Port = NEXT_SERVER_BACKEND_PORT
	config.NumThreads = 2
	config.SocketReadBuffer = 1024 * 1024
	config.SocketWriteBuffer = 1024 * 1024
	config.MaxPacketSize = 4096
	config.BindAddress = bindAddress

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

	var emptyMagic [constants.MagicBytes]byte

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

	packetData, err := packets.SDK5_WritePacket(packet, packetType, 4096, &serverBackendAddress, to, TestServerBackendPrivateKey)
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
			fmt.Printf("near relay: id = %x, rtt = %d, jitter = %d, packet loss = %.2f\n", requestPacket.NearRelayIds[i], requestPacket.NearRelayRTT[i], requestPacket.NearRelayJitter[i], requestPacket.NearRelayPacketLoss[i])
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
		if requestPacket.SliceNumber >= 2 && requestPacket.SessionEvents != 0x123 {
			panic("session events not set on session update")
		}
	}

	if requestPacket.SessionEvents > 0 {
		fmt.Printf("session events %x\n", requestPacket.SessionEvents)
	}

	// build response packet

	var responsePacket *packets.SDK5_SessionUpdateResponsePacket

	if !takeNetworkNext {

		// direct route

		responsePacket = &packets.SDK5_SessionUpdateResponsePacket{
			SessionId:     requestPacket.SessionId,
			SliceNumber:   requestPacket.SliceNumber,
			NumNearRelays: int32(numRelays),
			RouteType:     int32(packets.SDK5_RouteTypeDirect),
			NumTokens:     0,
			Tokens:        nil,
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
			routeRelayPublicKeys[i] = TestRelayPublicKey
		}

		// is this a continue route, or a new route?

		var routeType int32

		sameRoute := numRouteRelays == int(sessionData.RouteNumRelays) && routeRelayIds == sessionData.RouteRelayIds

		// build token data

		routerPrivateKey := [packets.SDK5_PrivateKeyBytes]byte{}
		copy(routerPrivateKey[:], TestRelayBackendPrivateKey)

		numTokens := numRouteRelays + 2

		tokenAddresses := make([]net.UDPAddr, numTokens)
		tokenAddresses[0] = requestPacket.ClientAddress
		tokenAddresses[len(tokenAddresses)-1] = requestPacket.ServerAddress
		for i := 0; i < numRouteRelays; i++ {
			tokenAddresses[1+i] = relayAddresses[i]
		}

		tokenPublicKeys := make([][]byte, numTokens)
		tokenPublicKeys[0] = requestPacket.ClientRoutePublicKey[:]
		tokenPublicKeys[len(tokenPublicKeys)-1] = requestPacket.ServerRoutePublicKey[:]
		for i := 0; i < numRouteRelays; i++ {
			tokenPublicKeys[1+i] = TestRelayPublicKey
		}

		tokenInternal := make([]bool, numTokens)

		var tokenData []byte

		if sameRoute {
			tokenData = make([]byte, numTokens*packets.SDK5_EncryptedContinueRouteTokenSize)
			core.WriteContinueTokens(tokenData, sessionData.ExpireTimestamp, sessionData.SessionId, uint8(sessionData.SessionVersion), int(numTokens), tokenPublicKeys, routerPrivateKey[:])
			routeType = packets.SDK5_RouteTypeContinue
		} else {
			sessionData.ExpireTimestamp += packets.SDK5_BillingSliceSeconds
			sessionData.SessionVersion++

			tokenData = make([]byte, numTokens*packets.SDK5_EncryptedNextRouteTokenSize)
			core.WriteRouteTokens(tokenData, sessionData.ExpireTimestamp, sessionData.SessionId, uint8(sessionData.SessionVersion), 256, 256, int(numTokens), tokenAddresses, tokenPublicKeys, tokenInternal, routerPrivateKey[:])
			routeType = packets.SDK5_RouteTypeNew
		}

		// contruct the session update response packet

		responsePacket = &packets.SDK5_SessionUpdateResponsePacket{
			SessionId:   requestPacket.SessionId,
			SliceNumber: requestPacket.SliceNumber,
			RouteType:   routeType,
			Multipath:   multipath,
			NumTokens:   int32(numTokens),
			Tokens:      tokenData,
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

	backend.relayManager = common.CreateRelayManager(true)

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
