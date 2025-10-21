/*
   Network Next. Copyright 2017 - 2025 Network Next, Inc.
   Licensed under the Network Next Source Available License 1.0
*/

package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"

	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/constants"
	"github.com/networknext/next/modules/core"
	"github.com/networknext/next/modules/crypto"
	"github.com/networknext/next/modules/encoding"
	"github.com/networknext/next/modules/packets"
)

func Base64String(value string) []byte {
	data, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		panic(err)
	}
	return data
}

var TestRelayPublicKey = Base64String("1nTj7bQmo8gfIDqG+o//GFsak/g1TRo4hl6XXw1JkyI=")
var TestRelayPrivateKey = Base64String("cwvK44Pr5aHI3vE3siODS7CUgdPI/l1VwjVZ2FvEyAo=")
var TestRelayBackendPublicKey = Base64String("IsjRpWEz9H7qslhWWupW4A9LIpVh+PzWoLleuXL1NUE=")
var TestRelayBackendPrivateKey = Base64String("qXeUdLPZxaMnZ/zFHLHkmgkQOmunWq1AmRv55nqTYMg=")
var TestServerBackendPublicKey = Base64String("1wXeogqOEL/UuMnHy3lwpdkdklcg4IktO/39mJiYfgc=")
var TestServerBackendPrivateKey = Base64String("peZ17P29VgtnOiEv5wwNPDDo9lWweFV7dBVac0KoaXHXBd6iCo4Qv9S4ycfLeXCl2R2SVyDgiS07/f2YmJh+Bw==")
var TestPingKey = Base64String("xsBL4b6PO4ESADcc69kERzLXxs9ESOrX1kSHJH0m9D0=")

const NEXT_RELAY_BACKEND_PORT = 30000
const NEXT_SERVER_BACKEND_PORT = 45000

const BACKEND_MODE_FORCE_DIRECT = 1
const BACKEND_MODE_RANDOM = 2
const BACKEND_MODE_ON_OFF = 3
const BACKEND_MODE_ON_ON_OFF = 4
const BACKEND_MODE_ROUTE_SWITCHING = 5
const BACKEND_MODE_SERVER_EVENTS = 6
const BACKEND_MODE_FORCE_RETRY = 7
const BACKEND_MODE_BANDWIDTH = 8
const BACKEND_MODE_JITTER = 9
const BACKEND_MODE_DIRECT_STATS = 10
const BACKEND_MODE_NEXT_STATS = 11
const BACKEND_MODE_CLIENT_RELAY_STATS = 12
const BACKEND_MODE_SERVER_RELAY_STATS = 13
const BACKEND_MODE_ZERO_MAGIC = 14

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
	BuyerID                    uint64
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
		relayPrice := make([]uint8, numRelays)
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

		core.Optimize2(numRelays, numSegments, costMatrix, relayPrice, relayDatacenterIds, destRelays)

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
	if backend.mode != BACKEND_MODE_ZERO_MAGIC {
		core.RandomBytes(newMagic)
	}
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
	if len(activeRelays) > constants.MaxClientRelays {
		activeRelays = activeRelays[:constants.MaxClientRelays]
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

	body, err := io.ReadAll(request.Body)

	if err != nil {
		core.Error("could not read body")
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

	if backend.mode == BACKEND_MODE_ZERO_MAGIC {
		if (requestPacket.RelayFlags & constants.RelayFlags_ShuttingDown) != 0 {
			fmt.Printf("relay is shutting down\n")
		}
	}

	if backend.mode == BACKEND_MODE_ZERO_MAGIC {
		for i := 0; i < constants.NumRelayCounters; i++ {
			if requestPacket.RelayCounters[i] != 0 {
				fmt.Printf("counter %d: %d\n", i, requestPacket.RelayCounters[i])
			}
		}
	}

	if backend.mode == BACKEND_MODE_ZERO_MAGIC {

		if requestPacket.SessionCount > 0 {
			fmt.Printf("session count = %d\n", requestPacket.SessionCount)
		}

		if requestPacket.EnvelopeBandwidthUpKbps > 0 {
			fmt.Printf("envelope bandwidth up kbps = %d\n", requestPacket.EnvelopeBandwidthUpKbps)
		}

		if requestPacket.EnvelopeBandwidthDownKbps > 0 {
			fmt.Printf("envelope bandwidth down kbps = %d\n", requestPacket.EnvelopeBandwidthDownKbps)
		}

		if requestPacket.PacketsSentPerSecond > 0 {
			fmt.Printf("packets sent per-second = %.2f\n", requestPacket.PacketsSentPerSecond)
		}

		if requestPacket.PacketsReceivedPerSecond > 0 {
			fmt.Printf("packets received per-second = %.2f\n", requestPacket.PacketsReceivedPerSecond)
		}

		if requestPacket.BandwidthSentKbps > 0 {
			fmt.Printf("bandwidth sent kbps = %.2f\n", requestPacket.BandwidthSentKbps)
		}

		if requestPacket.BandwidthReceivedKbps > 0 {
			fmt.Printf("bandwidth received kbps = %.2f\n", requestPacket.BandwidthReceivedKbps)
		}

		if requestPacket.ClientPingsPerSecond > 0 {
			fmt.Printf("client pings per-second = %.2f\n", requestPacket.ClientPingsPerSecond)
		}

		if requestPacket.RelayPingsPerSecond > 0 {
			fmt.Printf("relay pings per-second = %.2f\n", requestPacket.RelayPingsPerSecond)
		}
	}

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

	testSecretKey, _ := crypto.SecretKey_GenerateLocal(TestRelayPublicKey, TestRelayPrivateKey, TestRelayBackendPublicKey)

	token := core.RouteToken{}
	token.NextAddress = net.UDPAddr{IP: net.IPv4(0, 0, 0, 0), Port: 10000}
	token.PrevAddress = net.UDPAddr{IP: net.IPv4(0, 0, 0, 0), Port: 20000}
	core.WriteEncryptedRouteToken(&token, responsePacket.TestToken[:], testSecretKey)

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

	relayPrice := make([]byte, len(activeRelays))

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
		RelayPrice:         relayPrice,
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

	credentials := handlers.AllowCredentials()
	methods := handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE"})
	origins := handlers.AllowedOrigins([]string{"127.0.0.1"})

	http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", NEXT_RELAY_BACKEND_PORT), handlers.CORS(credentials, methods, origins)(router))
}

func StartUDPServer() {

	bindAddress := core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", NEXT_SERVER_BACKEND_PORT))

	serverBackendAddress = core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", NEXT_SERVER_BACKEND_PORT))

	config := common.UDPServerConfig{}
	config.Port = NEXT_SERVER_BACKEND_PORT
	config.NumThreads = 2
	config.SocketReadBuffer = 1024 * 1024
	config.SocketWriteBuffer = 1024 * 1024
	config.MaxPacketSize = 1384
	config.BindAddress = bindAddress

	common.CreateUDPServer(context.Background(), config, packetHandler)
}

// -------------------------------------------------

var serverBackendAddress net.UDPAddr

func packetHandler(conn *net.UDPConn, from *net.UDPAddr, packetData []byte) {

	// ignore packets that are too small

	if len(packetData) < packets.SDK_MinPacketBytes {
		fmt.Printf("packet is too small")
		return
	}

	// ignore packet types we don't support

	packetType := packetData[0]

	if packetType < packets.SDK_SERVER_INIT_REQUEST_PACKET {
		fmt.Printf("unsupported packet type %d\n", packetType)
		return
	}

	// check basic packet filetr

	if !core.BasicPacketFilter(packetData, len(packetData)) {
		fmt.Printf("basic packet filter failed\n")
		return
	}

	// check advanced packet filter

	var emptyMagic [constants.MagicBytes]byte

	fromAddressData := core.GetAddressData(from)
	toAddressData := core.GetAddressData(&serverBackendAddress)

	if !core.AdvancedPacketFilter(packetData, emptyMagic[:], fromAddressData, toAddressData, len(packetData)) {
		fmt.Printf("advanced packet filter failed\n")
		return
	}

	// process the packet according to type

	packetData = packetData[18:]

	switch packetType {

	case packets.SDK_SERVER_INIT_REQUEST_PACKET:
		packet := packets.SDK_ServerInitRequestPacket{}
		if err := packets.ReadPacket(packetData, &packet); err != nil {
			core.Error("could not read server init request packet from %s: %v", from.String(), err)
			return
		}
		ProcessServerInitRequestPacket(conn, from, &packet)
		break

	case packets.SDK_SERVER_UPDATE_REQUEST_PACKET:
		packet := packets.SDK_ServerUpdateRequestPacket{}
		if err := packets.ReadPacket(packetData, &packet); err != nil {
			core.Error("could not read server update request packet from %s: %v", from.String(), err)
			return
		}
		ProcessServerUpdateRequestPacket(conn, from, &packet)
		break

	case packets.SDK_SESSION_UPDATE_REQUEST_PACKET:
		packet := packets.SDK_SessionUpdateRequestPacket{}
		if err := packets.ReadPacket(packetData, &packet); err != nil {
			core.Error("could not read session update request packet from %s: %v", from.String(), err)
			return
		}
		ProcessSessionUpdateRequestPacket(conn, from, &packet)
		break

	case packets.SDK_CLIENT_RELAY_REQUEST_PACKET:
		packet := packets.SDK_ClientRelayRequestPacket{}
		if err := packets.ReadPacket(packetData, &packet); err != nil {
			core.Error("could not read client relay request packet from %s: %v", from.String(), err)
			return
		}
		ProcessClientRelayRequestPacket(conn, from, &packet)
		break

	case packets.SDK_SERVER_RELAY_REQUEST_PACKET:
		packet := packets.SDK_ServerRelayRequestPacket{}
		if err := packets.ReadPacket(packetData, &packet); err != nil {
			core.Error("could not read server relay request packet from %s: %v", from.String(), err)
			return
		}
		ProcessServerRelayRequestPacket(conn, from, &packet)
		break

	default:
		panic("unknown packet type")
	}
}

func SendResponsePacket[P packets.Packet](conn *net.UDPConn, to *net.UDPAddr, packetType int, packet P) {

	packetData, err := packets.SDK_WritePacket(packet, packetType, constants.MaxPacketBytes, &serverBackendAddress, to, TestServerBackendPrivateKey)
	if err != nil {
		core.Error("failed to write response packet: %v", err)
		return
	}

	if _, err := conn.WriteToUDP(packetData, to); err != nil {
		core.Error("failed to send response packet: %v", err)
		return
	}
}

func ProcessServerInitRequestPacket(conn *net.UDPConn, from *net.UDPAddr, requestPacket *packets.SDK_ServerInitRequestPacket) {

	fmt.Printf("server init request from %s\n", from.String())

	responsePacket := &packets.SDK_ServerInitResponsePacket{
		RequestId: requestPacket.RequestId,
		Response:  packets.SDK_ServerInitResponseOK,
	}

	responsePacket.UpcomingMagic, responsePacket.CurrentMagic, responsePacket.PreviousMagic = GetMagic()

	SendResponsePacket(conn, from, packets.SDK_SERVER_INIT_RESPONSE_PACKET, responsePacket)
}

func ProcessServerUpdateRequestPacket(conn *net.UDPConn, from *net.UDPAddr, requestPacket *packets.SDK_ServerUpdateRequestPacket) {

	fmt.Printf("server update request from %s\n", from.String())

	responsePacket := &packets.SDK_ServerUpdateResponsePacket{
		RequestId: requestPacket.RequestId,
	}

	responsePacket.UpcomingMagic, responsePacket.CurrentMagic, responsePacket.PreviousMagic = GetMagic()

	SendResponsePacket(conn, from, packets.SDK_SERVER_UPDATE_RESPONSE_PACKET, responsePacket)
}

func ProcessClientRelayRequestPacket(conn *net.UDPConn, from *net.UDPAddr, requestPacket *packets.SDK_ClientRelayRequestPacket) {

	fmt.Printf("client relay request from %s\n", from.String())

	relayIds, relayAddresses := backend.GetRelays()

	numRelays := len(relayIds)
	if numRelays > constants.MaxClientRelays {
		numRelays = constants.MaxClientRelays
	}

	relayIds = relayIds[:numRelays]
	relayAddresses = relayAddresses[:numRelays]

	responsePacket := &packets.SDK_ClientRelayResponsePacket{
		RequestId:       requestPacket.RequestId,
		ClientAddress:   requestPacket.ClientAddress,
		Latitude:        41, // iowa
		Longitude:       -93,
		NumClientRelays: int32(numRelays),
		ExpireTimestamp: uint64(time.Now().Unix()) + 15,
	}

	clientAddressWithoutPort := requestPacket.ClientAddress
	clientAddressWithoutPort.Port = 0

	for i := 0; i < numRelays; i++ {
		responsePacket.ClientRelayIds[i] = relayIds[i]
		responsePacket.ClientRelayAddresses[i] = relayAddresses[i]
		core.GeneratePingToken(responsePacket.ExpireTimestamp, &clientAddressWithoutPort, &responsePacket.ClientRelayAddresses[i], TestPingKey, responsePacket.ClientRelayPingTokens[i][:])
	}

	SendResponsePacket(conn, from, packets.SDK_CLIENT_RELAY_RESPONSE_PACKET, responsePacket)
}

func ProcessServerRelayRequestPacket(conn *net.UDPConn, from *net.UDPAddr, requestPacket *packets.SDK_ServerRelayRequestPacket) {

	fmt.Printf("server relay request from %s\n", from.String())

	relayIds, relayAddresses := backend.GetRelays()

	numRelays := len(relayIds)
	if numRelays > constants.MaxServerRelays {
		numRelays = constants.MaxServerRelays
	}

	relayIds = relayIds[:numRelays]
	relayAddresses = relayAddresses[:numRelays]

	responsePacket := &packets.SDK_ServerRelayResponsePacket{
		RequestId:       requestPacket.RequestId,
		NumServerRelays: int32(numRelays),
		ExpireTimestamp: uint64(time.Now().Unix()) + 15,
	}

	for i := 0; i < numRelays; i++ {
		responsePacket.ServerRelayIds[i] = relayIds[i]
		responsePacket.ServerRelayAddresses[i] = relayAddresses[i]
		core.GeneratePingToken(responsePacket.ExpireTimestamp, from, &responsePacket.ServerRelayAddresses[i], TestPingKey, responsePacket.ServerRelayPingTokens[i][:])
	}

	SendResponsePacket(conn, from, packets.SDK_SERVER_RELAY_RESPONSE_PACKET, responsePacket)
}

func ProcessSessionUpdateRequestPacket(conn *net.UDPConn, from *net.UDPAddr, requestPacket *packets.SDK_SessionUpdateRequestPacket) {

	fmt.Printf("session update request from %s\n", from.String())

	if backend.mode == BACKEND_MODE_FORCE_RETRY && requestPacket.RetryNumber < 4 {
		return
	}

	if requestPacket.PlatformType == packets.SDK_PlatformTypeUnknown {
		panic("platform type is unknown")
	}

	if requestPacket.ConnectionType == packets.SDK_ConnectionTypeUnknown {
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

	if requestPacket.PacketsLostClientToServer > 0 {
		fmt.Printf("%d client to server packets lost\n", requestPacket.PacketsLostClientToServer)
	}

	if requestPacket.PacketsLostServerToClient > 0 {
		fmt.Printf("%d server to client packets lost\n", requestPacket.PacketsLostServerToClient)
	}

	if backend.mode == BACKEND_MODE_BANDWIDTH {
		if requestPacket.BandwidthKbpsUp > 0 {
			fmt.Printf("%d bandwidth kbps up\n", requestPacket.BandwidthKbpsUp)
		}
		if requestPacket.BandwidthKbpsDown > 0 {
			fmt.Printf("%d bandwidth kbps down\n", requestPacket.BandwidthKbpsDown)
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

	if backend.mode == BACKEND_MODE_CLIENT_RELAY_STATS {
		for i := 0; i <= int(requestPacket.NumClientRelays); i++ {
			fmt.Printf("client relay: id = %x, rtt = %d, jitter = %d, packet loss = %.2f\n", requestPacket.ClientRelayIds[i], requestPacket.ClientRelayRTT[i], requestPacket.ClientRelayJitter[i], requestPacket.ClientRelayPacketLoss[i])
		}
	}

	if backend.mode == BACKEND_MODE_SERVER_RELAY_STATS {
		for i := 0; i <= int(requestPacket.NumServerRelays); i++ {
			fmt.Printf("server relay: id = %x, rtt = %d, jitter = %d, packet loss = %.2f\n", requestPacket.ServerRelayIds[i], requestPacket.ServerRelayRTT[i], requestPacket.ServerRelayJitter[i], requestPacket.ServerRelayPacketLoss[i])
		}
	}

	// read the session data

	newSession := requestPacket.SliceNumber == 0

	var sessionData packets.SDK_SessionData

	if newSession {

		sessionData.Version = packets.SDK_SessionDataVersion_Write
		sessionData.SessionId = requestPacket.SessionId
		sessionData.SliceNumber = uint32(requestPacket.SliceNumber + 1)
		sessionData.ExpireTimestamp = uint64(time.Now().Unix()) + packets.SDK_SliceSeconds

	} else {

		readStream := encoding.CreateReadStream(requestPacket.SessionData[:])

		err := sessionData.Serialize(readStream)
		if err != nil {
			fmt.Printf("could not read session data in session update packet: %v\n", err)
			return
		}

		sessionData.SliceNumber = uint32(requestPacket.SliceNumber + 1)
		sessionData.ExpireTimestamp += packets.SDK_SliceSeconds
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

	if backend.mode == BACKEND_MODE_SERVER_EVENTS {
		if requestPacket.SliceNumber >= 2 && requestPacket.SessionEvents != 0x123 {
			panic("session events not set on session update")
		}
	}

	if requestPacket.SessionEvents > 0 {
		fmt.Printf("session events %x\n", requestPacket.SessionEvents)
	}

	// build response packet

	var responsePacket *packets.SDK_SessionUpdateResponsePacket

	if !takeNetworkNext {

		// direct route

		responsePacket = &packets.SDK_SessionUpdateResponsePacket{
			SessionId:   requestPacket.SessionId,
			SliceNumber: requestPacket.SliceNumber,
			RouteType:   int32(packets.SDK_RouteTypeDirect),
			NumTokens:   0,
			Tokens:      nil,
			Multipath:   true,
		}

	} else {

		// next

		const MaxRouteRelays = packets.SDK_MaxRelaysPerRoute

		var routeRelayIds [MaxRouteRelays]uint64
		var routePublicAddresses [MaxRouteRelays]net.UDPAddr
		var routePublicKeys [MaxRouteRelays][]byte

		numRouteRelays := numRelays
		if numRouteRelays > MaxRouteRelays {
			numRouteRelays = MaxRouteRelays
		}

		for i := 0; i < numRouteRelays; i++ {
			routeRelayIds[i] = relayIds[i]
			routePublicAddresses[i] = relayAddresses[i]
			routePublicKeys[i] = TestRelayPublicKey
		}

		// is this a continue route, or a new route?

		var routeType int32

		sameRoute := numRouteRelays == int(sessionData.RouteNumRelays) && routeRelayIds == sessionData.RouteRelayIds

		// build token data

		routerPrivateKey := [packets.SDK_PrivateKeyBytes]byte{}
		copy(routerPrivateKey[:], TestRelayBackendPrivateKey)

		numTokens := numRouteRelays + 2

		tokenPublicAddresses := make([]net.UDPAddr, numTokens)
		tokenHasInternalAddresses := make([]bool, numTokens)
		tokenInternalAddresses := make([]net.UDPAddr, numTokens)
		tokenInternalGroups := make([]uint64, numTokens)
		tokenSellers := make([]int, numTokens)

		tokenPublicAddresses[0] = requestPacket.ClientAddress
		tokenPublicAddresses[len(tokenPublicAddresses)-1] = requestPacket.ServerAddress
		for i := 0; i < numRouteRelays; i++ {
			tokenPublicAddresses[1+i] = relayAddresses[i]
		}

		tokenPublicKeys := make([][]byte, numTokens)
		tokenPublicKeys[0] = requestPacket.ClientRoutePublicKey[:]
		tokenPublicKeys[len(tokenPublicKeys)-1] = requestPacket.ServerRoutePublicKey[:]
		for i := 0; i < numRouteRelays; i++ {
			tokenPublicKeys[1+i] = TestRelayPublicKey
		}

		var routeSecretKeys [constants.NextMaxNodes][]byte
		routeSecretKeys[0], _ = crypto.SecretKey_GenerateRemote(TestRelayBackendPublicKey, TestRelayBackendPrivateKey, requestPacket.ClientRoutePublicKey[:])
		relaySecretKeys := routeSecretKeys[1 : numTokens-1]
		for i := 0; i < numRouteRelays; i++ {
			relaySecretKeys[i], _ = crypto.SecretKey_GenerateRemote(TestRelayBackendPublicKey, TestRelayBackendPrivateKey, TestRelayPublicKey)
		}
		routeSecretKeys[numTokens-1], _ = crypto.SecretKey_GenerateRemote(TestRelayBackendPublicKey, TestRelayBackendPrivateKey, requestPacket.ServerRoutePublicKey[:])

		var tokenData []byte

		if sameRoute {
			tokenData = make([]byte, numTokens*packets.SDK_EncryptedContinueRouteTokenSize)
			core.WriteContinueTokens(tokenData, sessionData.ExpireTimestamp, sessionData.SessionId, uint8(sessionData.SessionVersion), int(numTokens), routeSecretKeys[:])
			routeType = packets.SDK_RouteTypeContinue
		} else {
			sessionData.ExpireTimestamp += packets.SDK_SliceSeconds
			sessionData.SessionVersion++
			tokenData = make([]byte, numTokens*packets.SDK_EncryptedNextRouteTokenSize)
			core.WriteRouteTokens(tokenData, sessionData.ExpireTimestamp, sessionData.SessionId, uint8(sessionData.SessionVersion), 256, 256, int(numTokens), tokenPublicAddresses, tokenHasInternalAddresses, tokenInternalAddresses, tokenInternalGroups, tokenSellers, routeSecretKeys[:])
			routeType = packets.SDK_RouteTypeNew
		}

		// contruct the session update response packet

		responsePacket = &packets.SDK_SessionUpdateResponsePacket{
			SessionId:   requestPacket.SessionId,
			SliceNumber: requestPacket.SliceNumber,
			RouteType:   routeType,
			Multipath:   true,
			NumTokens:   int32(numTokens),
			Tokens:      tokenData,
		}
	}

	if responsePacket == nil {
		fmt.Printf("error: nil session response\n")
		return
	}

	packetSessionData, err := packets.WritePacket[*packets.SDK_SessionData](responsePacket.SessionData[:], &sessionData)

	if err != nil {
		fmt.Printf("error: failed to write session data\n")
		return
	}

	responsePacket.SessionDataBytes = int32(len(packetSessionData))

	SendResponsePacket(conn, from, packets.SDK_SESSION_UPDATE_RESPONSE_PACKET, responsePacket)
}

// -----------------------------------------------

func main() {

	rand.Seed(time.Now().UnixNano())

	backend.relayManager = common.CreateRelayManager(false) // IMPORTANT: Create without history

	if os.Getenv("BACKEND_MODE") == "FORCE_DIRECT" {
		backend.mode = BACKEND_MODE_FORCE_DIRECT
	}

	if os.Getenv("BACKEND_MODE") == "RANDOM" {
		backend.mode = BACKEND_MODE_RANDOM
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

	if os.Getenv("BACKEND_MODE") == "ZERO_MAGIC" {
		backend.mode = BACKEND_MODE_ZERO_MAGIC
	}

	GenerateMagic(magicUpcoming[:])
	GenerateMagic(magicCurrent[:])
	GenerateMagic(magicPrevious[:])

	go OptimizeThread()

	go UpdateMagic()

	go StartWebServer()

	go StartUDPServer()

	fmt.Printf("started functional backend on ports %d and %d\n", NEXT_RELAY_BACKEND_PORT, NEXT_SERVER_BACKEND_PORT)

	// Wait for shutdown signal
	termChan := make(chan os.Signal, 1)
	signal.Notify(termChan, os.Interrupt, syscall.SIGTERM)

	select {
	case <-termChan:
		core.Debug("received shutdown signal")
		return
	}
}
