/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/rand"
	"net"
	"net/http"
	"os"
	"sort"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"golang.org/x/sys/unix"

	"github.com/networknext/backend/billing"
	"github.com/networknext/backend/core"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/transport"
)

const NEXT_RELAY_BACKEND_PORT = 30000
const NEXT_SERVER_BACKEND_PORT = 40000

const BACKEND_MODE_FORCE_DIRECT = 1
const BACKEND_MODE_RANDOM = 2
const BACKEND_MODE_MULTIPATH = 3
const BACKEND_MODE_ON_OFF = 4
const BACKEND_MODE_ON_ON_OFF = 5
const BACKEND_MODE_ROUTE_SWITCHING = 6
const BACKEND_MODE_UNCOMMITTED = 7
const BACKEND_MODE_UNCOMMITTED_TO_COMMITTED = 8
const BACKEND_MODE_USER_FLAGS = 9

type Backend struct {
	mutex           sync.RWMutex
	dirty           bool
	mode            int
	serverDatabase  map[string]ServerEntry
	sessionDatabase map[uint64]SessionCacheEntry
	statsDatabase   *routing.StatsDatabase
	costMatrix      *routing.CostMatrix
	routeMatrix     *routing.RouteMatrix
	nearData        []byte

	relayMap *routing.RelayMap
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
	RouteDecision              routing.Decision
	OnNNSliceCounter           uint64
	CommitPending              bool
	CommitObservedSliceCounter uint8
	Committed                  bool
	TimestampStart             time.Time
	TimestampExpire            time.Time
	Version                    uint8
	DirectRTT                  float64
	NextRTT                    float64
	Location                   routing.Location
	Response                   []byte
}

const ThresholdRTT = 1.0
const MaxJitter = float32(10.0)
const MaxPacketLoss = float32(0.1)

func OptimizeThread() {

	for {
		backend.mutex.Lock()

		if err := backend.statsDatabase.GetCostMatrix(backend.costMatrix, backend.relayMap.GetAllRelayData(), MaxJitter, MaxPacketLoss); err != nil {
			fmt.Printf("error generating cost matrix: %v\n", err)
		}

		if err := backend.costMatrix.Optimize(backend.routeMatrix, ThresholdRTT); err != nil {
			fmt.Printf("error generating route matrix: %v\n", err)
		}

		backend.mutex.Unlock()

		time.Sleep(1 * time.Second)
	}
}

func TimeoutThread() {
	for {
		time.Sleep(time.Second * 1)
		backend.mutex.Lock()
		currentTime := time.Now()
		currentTimestamp := currentTime.Unix()
		unixTimeout := int64(routing.RelayTimeout.Seconds())
		for k, v := range backend.serverDatabase {
			if currentTimestamp-v.lastUpdate > unixTimeout {
				delete(backend.serverDatabase, k)
				backend.dirty = true
				continue
			}
		}
		for k, v := range backend.sessionDatabase {
			if v.TimestampExpire.Before(currentTime) {
				delete(backend.sessionDatabase, k)
				backend.dirty = true
				continue
			}
		}

		allRelayData := backend.relayMap.GetAllRelayData()
		for _, relayData := range allRelayData {
			if currentTimestamp-relayData.LastUpdateTime.Unix() > unixTimeout {
				backend.relayMap.RemoveRelayData(relayData.Addr.String())
				backend.dirty = true
				continue
			}
		}
		if backend.dirty {
			fmt.Printf("-----------------------------\n")
			allRelayData := backend.relayMap.GetAllRelayData()
			for _, relayData := range allRelayData {
				fmt.Printf("relay: %v\n", &relayData.Addr)
			}

			for _, v := range backend.serverDatabase {
				fmt.Printf("server: %s\n", v.address)
			}
			for k := range backend.sessionDatabase {
				fmt.Printf("session: %x\n", k)
			}
			backend.dirty = false
		}
		backend.mutex.Unlock()
	}
}

func (backend *Backend) GetNearRelays() []*routing.RelayData {
	backend.mutex.Lock()
	allRelayData := backend.relayMap.GetAllRelayData()
	backend.mutex.Unlock()
	sort.SliceStable(allRelayData, func(i, j int) bool { return allRelayData[i].ID < allRelayData[j].ID })
	if len(allRelayData) > int(transport.MaxNearRelays) {
		allRelayData = allRelayData[:transport.MaxNearRelays]
	}
	return allRelayData
}

func ServerInitHandlerFunc(w io.Writer, incoming *transport.UDPPacket) {
	var initRequest transport.ServerInitRequestPacket4
	if err := transport.UnmarshalPacket(&initRequest, incoming.Data); err != nil {
		fmt.Printf("error: failed to read server init request packet: %v\n", err)
		return
	}

	initResponse := &transport.ServerInitResponsePacket4{
		RequestID: initRequest.RequestID,
		Response:  transport.InitResponseOK,
	}

	initResponseData, err := transport.MarshalPacket(initResponse)
	if err != nil {
		fmt.Printf("error: failed to marshal server init response: %v\n", err)
		return
	}

	packetHeader := append([]byte{transport.PacketTypeServerInitResponse4}, make([]byte, crypto.PacketHashSize)...)
	responseData := append(packetHeader, initResponseData...)
	if _, err := w.Write(responseData); err != nil {
		fmt.Printf("error: failed to write server init response: %v\n", err)
		return
	}
}

func ServerUpdateHandlerFunc(w io.Writer, incoming *transport.UDPPacket) {
	var serverUpdate transport.ServerUpdatePacket4
	if err := transport.UnmarshalPacket(&serverUpdate, incoming.Data); err != nil {
		fmt.Printf("error: failed to read server update packet: %v\n", err)
		return
	}
}

func SessionUpdateHandlerFunc(w io.Writer, incoming *transport.UDPPacket) {
	var sessionUpdate transport.SessionUpdatePacket4
	if err := transport.UnmarshalPacket(&sessionUpdate, incoming.Data); err != nil {
		fmt.Printf("error: failed to read session update packet: %v\n", err)
		return
	}

	if sessionUpdate.FallbackToDirect {
		fmt.Printf("error: fallback to direct %s\n", incoming.SourceAddr.String())
		return
	}

	newSession := sessionUpdate.SliceNumber == 0

	var sessionData transport.SessionData4
	if newSession {
		sessionData.Version = transport.SessionDataVersion4
		sessionData.SessionID = sessionUpdate.SessionID
		sessionData.SliceNumber = uint32(sessionUpdate.SliceNumber + 1)
		sessionData.ExpireTimestamp = uint64(time.Now().Unix()) + billing.BillingSliceSeconds
		sessionData.RouteState.UserID = sessionUpdate.UserHash
		sessionData.Location = routing.LocationNullIsland
	} else {
		if err := transport.UnmarshalSessionData(&sessionData, sessionUpdate.SessionData[:]); err != nil {
			fmt.Printf("could not read session data in session update packet: %v\n", err)
			return
		}

		sessionData.SliceNumber = uint32(sessionUpdate.SliceNumber + 1)
		sessionData.ExpireTimestamp += billing.BillingSliceSeconds
	}

	nearRelays := backend.GetNearRelays()

	var sessionResponse *transport.SessionResponsePacket4

	takeNetworkNext := len(nearRelays) > 0

	if backend.mode == BACKEND_MODE_FORCE_DIRECT {
		takeNetworkNext = false
	}

	if backend.mode == BACKEND_MODE_RANDOM {
		takeNetworkNext = takeNetworkNext && rand.Float32() > 0.5
	}

	if backend.mode == BACKEND_MODE_ON_OFF {
		// Alternate between direct and next routes for each slice
		if (sessionUpdate.SliceNumber & 1) == 0 {
			takeNetworkNext = false
		}
	}

	if backend.mode == BACKEND_MODE_ON_ON_OFF {
		// Alternate between direct, a new route token and a continue route token for every 3 slices
		if (sessionUpdate.SliceNumber & 2) == 0 {
			takeNetworkNext = false
		}
	}

	if backend.mode == BACKEND_MODE_ROUTE_SWITCHING {
		rand.Shuffle(len(nearRelays), func(i, j int) {
			nearRelays[i], nearRelays[j] = nearRelays[j], nearRelays[i]
		})
	}

	multipath := len(nearRelays) > 0 && backend.mode == BACKEND_MODE_MULTIPATH

	committed := true

	if backend.mode == BACKEND_MODE_UNCOMMITTED {
		committed = false
		if sessionUpdate.Committed {
			panic("slices must not be committed in this mode")
		}
	}

	if backend.mode == BACKEND_MODE_UNCOMMITTED_TO_COMMITTED {
		committed = sessionUpdate.SliceNumber > 2
		if sessionUpdate.SliceNumber <= 2 && sessionUpdate.Committed {
			panic("slices 0,1,2,3 should not be committed")
		}
		if sessionUpdate.SliceNumber >= 4 && !sessionUpdate.Committed {
			panic("slices 4 and greater should be committed")
		}
	}

	if backend.mode == BACKEND_MODE_USER_FLAGS {
		if sessionUpdate.SliceNumber >= 2 && sessionUpdate.UserFlags != 0x123 {
			panic("user flags not set on session update")
		}
	}

	// Extract ids and addresses into own list to make response
	var nearRelayIDs = [MaxRelays]uint64{}
	var nearRelayAddresses = [MaxRelays]net.UDPAddr{}
	var nearRelayPublicKeys = [MaxRelays][]byte{}
	for i, relay := range nearRelays {
		nearRelayIDs[i] = relay.ID
		nearRelayAddresses[i] = relay.Addr
		nearRelayPublicKeys[i] = relay.PublicKey
	}

	if !takeNetworkNext {

		// direct route
		sessionResponse = &transport.SessionResponsePacket4{
			SessionID:          sessionUpdate.SessionID,
			SliceNumber:        sessionUpdate.SliceNumber,
			NumNearRelays:      int32(len(nearRelays)),
			NearRelayIDs:       nearRelayIDs[:len(nearRelays)],
			NearRelayAddresses: nearRelayAddresses[:len(nearRelays)],
			RouteType:          int32(routing.RouteTypeDirect),
			NumTokens:          0,
			Tokens:             nil,
		}

	} else {

		// Make next route from near relays (but respect hop limit)
		numRelays := len(nearRelays)
		if numRelays > routing.MaxRelays {
			numRelays = routing.MaxRelays
		}
		nextRoute := routing.Route{
			NumRelays:       numRelays,
			RelayIDs:        nearRelayIDs,
			RelayAddrs:      nearRelayAddresses,
			RelayPublicKeys: nearRelayPublicKeys,
		}

		relayTokens := make([]routing.RelayToken, nextRoute.NumRelays)
		for i := range relayTokens {
			relayTokens[i] = routing.RelayToken{
				ID:        nextRoute.RelayIDs[i],
				Addr:      nextRoute.RelayAddrs[i],
				PublicKey: nextRoute.RelayPublicKeys[i],
			}
		}

		var routeType int32
		sameRoute := nextRoute.NumRelays == int(sessionData.RouteNumRelays) && nextRoute.RelayIDs == sessionData.RouteRelayIDs

		routerPrivateKey := [crypto.KeySize]byte{}
		copy(routerPrivateKey[:], crypto.RouterPrivateKey)

		tokenAddresses := make([]*net.UDPAddr, nextRoute.NumRelays+2)
		tokenAddresses[0] = &sessionUpdate.ClientAddress
		tokenAddresses[len(tokenAddresses)-1] = &sessionUpdate.ServerAddress
		for i := 0; i < nextRoute.NumRelays; i++ {
			tokenAddresses[1+i] = &nearRelayAddresses[i]
		}

		tokenPublicKeys := make([][]byte, nextRoute.NumRelays+2)
		tokenPublicKeys[0] = sessionUpdate.ClientRoutePublicKey
		tokenPublicKeys[len(tokenPublicKeys)-1] = sessionUpdate.ServerRoutePublicKey
		for i := 0; i < nextRoute.NumRelays; i++ {
			tokenPublicKeys[1+i] = nearRelayPublicKeys[i]
		}

		numTokens := nextRoute.NumRelays + 2

		var tokenData []byte
		if sameRoute {
			tokenData = make([]byte, numTokens*routing.EncryptedContinueRouteTokenSize4)
			core.WriteContinueTokens(tokenData, sessionData.ExpireTimestamp, sessionData.SessionID, uint8(sessionData.SessionVersion), int(numTokens), nextRoute.RelayPublicKeys[:], routerPrivateKey)
			routeType = routing.RouteTypeContinue
		} else {
			sessionData.ExpireTimestamp += billing.BillingSliceSeconds
			sessionData.SessionVersion++

			tokenData = make([]byte, numTokens*routing.EncryptedNextRouteTokenSize4)
			core.WriteRouteTokens(tokenData, sessionData.ExpireTimestamp, sessionData.SessionID, uint8(sessionData.SessionVersion), 1024, 1024, int(numTokens), tokenAddresses, tokenPublicKeys, routerPrivateKey)
			routeType = routing.RouteTypeNew
		}

		sessionResponse = &transport.SessionResponsePacket4{
			SessionID:          sessionUpdate.SessionID,
			SliceNumber:        sessionUpdate.SliceNumber,
			NumNearRelays:      int32(len(nearRelays)),
			NearRelayIDs:       nearRelayIDs[:len(nearRelays)],
			NearRelayAddresses: nearRelayAddresses[:len(nearRelays)],
			RouteType:          routeType,
			Multipath:          multipath,
			Committed:          committed,
			NumTokens:          int32(numTokens),
			Tokens:             tokenData,
		}
	}

	if sessionResponse == nil {
		fmt.Printf("error: nil session response\n")
		return
	}

	sessionDataBuffer, err := transport.MarshalSessionData(&sessionData)
	if err != nil {
		fmt.Printf("error: failed to marshal session data: %v\n", err)
		return
	}

	if len(sessionDataBuffer) > transport.MaxSessionDataSize {
		fmt.Printf("session data too large\n")
	}

	sessionResponse.SessionDataBytes = int32(len(sessionDataBuffer))
	copy(sessionResponse.SessionData[:], sessionDataBuffer)

	sessionResponseData, err := transport.MarshalPacket(sessionResponse)
	if err != nil {
		fmt.Printf("error: failed to marshal session response: %v\n", err)
		return
	}

	packetHeader := append([]byte{transport.PacketTypeSessionResponse4}, make([]byte, crypto.PacketHashSize)...)
	responseData := append(packetHeader, sessionResponseData...)
	if _, err := w.Write(responseData); err != nil {
		fmt.Printf("error: failed to write session response: %v\n", err)
		return
	}
}

func main() {

	rand.Seed(time.Now().UnixNano())

	backend.serverDatabase = make(map[string]ServerEntry)
	backend.sessionDatabase = make(map[uint64]SessionCacheEntry)
	backend.statsDatabase = routing.NewStatsDatabase()
	backend.costMatrix = &routing.CostMatrix{}
	backend.routeMatrix = &routing.RouteMatrix{}

	backend.relayMap = routing.NewRelayMap(func(relayData *routing.RelayData) error {
		backend.statsDatabase.DeleteEntry(relayData.ID)
		return nil
	})

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

	if os.Getenv("BACKEND_MODE") == "USER_FLAGS" {
		backend.mode = BACKEND_MODE_USER_FLAGS
	}

	go OptimizeThread()

	go TimeoutThread()

	go WebServer()

	fmt.Printf("started functional backend on ports %d and %d\n", NEXT_RELAY_BACKEND_PORT, NEXT_SERVER_BACKEND_PORT)

	lc := net.ListenConfig{
		Control: func(network string, address string, c syscall.RawConn) error {
			err := c.Control(func(fileDescriptor uintptr) {
				err := unix.SetsockoptInt(int(fileDescriptor), unix.SOL_SOCKET, unix.SO_REUSEADDR, 1)
				if err != nil {
					panic(fmt.Sprintf("failed to set reuse address socket option: %v", err))
				}

				err = unix.SetsockoptInt(int(fileDescriptor), unix.SOL_SOCKET, unix.SO_REUSEPORT, 1)
				if err != nil {
					panic(fmt.Sprintf("failed to set reuse port socket option: %v", err))
				}
			})

			return err
		},
	}

	for {
		lp, err := lc.ListenPacket(context.Background(), "udp", "0.0.0.0:"+fmt.Sprintf("%d", NEXT_SERVER_BACKEND_PORT))
		if err != nil {
			panic(fmt.Sprintf("could not bind socket: %v", err))
		}

		conn := lp.(*net.UDPConn)

		dataArray := [transport.DefaultMaxPacketSize]byte{}
		for {
			data := dataArray[:]
			size, fromAddr, err := conn.ReadFromUDP(data)
			if err != nil {
				fmt.Printf("failed to read udp packet: %v\n", err)
				break
			}

			if size <= 0 {
				continue
			}

			data = data[:size]

			// Check the packet hash is legit and remove the hash from the beginning of the packet
			// to continue processing the packet as normal
			if !crypto.IsNetworkNextPacket(crypto.PacketHashKey, data) {
				fmt.Println("received non network next packet")
				continue
			}

			packetType := data[0]
			data = data[crypto.PacketHashSize+1 : size]

			var buffer bytes.Buffer
			packet := transport.UDPPacket{SourceAddr: *fromAddr, Data: data}

			switch packetType {
			case transport.PacketTypeServerInitRequest4:
				ServerInitHandlerFunc(&buffer, &packet)
			case transport.PacketTypeServerUpdate4:
				ServerUpdateHandlerFunc(&buffer, &packet)
			case transport.PacketTypeSessionUpdate4:
				SessionUpdateHandlerFunc(&buffer, &packet)
			default:
				fmt.Printf("unknown packet type %d\n", packet.Data[0])
			}

			if buffer.Len() > 0 {
				response := buffer.Bytes()

				// Sign and hash the response
				response = crypto.SignPacket(crypto.BackendPrivateKey, response)
				crypto.HashPacket(crypto.PacketHashKey, response)

				if _, err := conn.WriteToUDP(response, fromAddr); err != nil {
					fmt.Printf("failed to write UDP response: %v\n", err)
				}
			}
		}
	}
}

// -----------------------------------------------------------

const InitRequestMagic = uint32(0x9083708f)
const InitRequestVersion = 0
const InitResponseVersion = 0
const UpdateRequestVersion = 0
const UpdateResponseVersion = 0
const MaxRelayAddressLength = 256
const RelayTokenBytes = 32
const MaxRelays = 5

func ReadUint32(data []byte, index *int, value *uint32) bool {
	if *index+4 > len(data) {
		return false
	}
	*value = binary.LittleEndian.Uint32(data[*index:])
	*index += 4
	return true
}

func ReadUint64(data []byte, index *int, value *uint64) bool {
	if *index+8 > len(data) {
		return false
	}
	*value = binary.LittleEndian.Uint64(data[*index:])
	*index += 8
	return true
}

func ReadFloat32(data []byte, index *int, value *float32) bool {
	var int_value uint32
	if !ReadUint32(data, index, &int_value) {
		return false
	}
	*value = math.Float32frombits(int_value)
	return true
}

func ReadString(data []byte, index *int, value *string, maxStringLength uint32) bool {
	var stringLength uint32
	if !ReadUint32(data, index, &stringLength) {
		return false
	}
	if stringLength > maxStringLength {
		return false
	}
	if *index+int(stringLength) > len(data) {
		return false
	}
	stringData := make([]byte, stringLength)
	for i := uint32(0); i < stringLength; i++ {
		stringData[i] = data[*index]
		*index += 1
	}
	*value = string(stringData)
	return true
}

func ReadBytes(data []byte, index *int, value *[]byte, bytes uint32) bool {
	if *index+int(bytes) > len(data) {
		return false
	}
	*value = make([]byte, bytes)
	for i := uint32(0); i < bytes; i++ {
		(*value)[i] = data[*index]
		*index += 1
	}
	return true
}

func WriteUint32(data []byte, index *int, value uint32) {
	binary.LittleEndian.PutUint32(data[*index:], value)
	*index += 4
}

func WriteUint64(data []byte, index *int, value uint64) {
	binary.LittleEndian.PutUint64(data[*index:], value)
	*index += 8
}

func WriteString(data []byte, index *int, value string, maxStringLength uint32) {
	stringLength := uint32(len(value))
	if stringLength > maxStringLength {
		panic("string is too long!\n")
	}
	binary.LittleEndian.PutUint32(data[*index:], stringLength)
	*index += 4
	for i := 0; i < int(stringLength); i++ {
		data[*index] = value[i]
		*index++
	}
}

func WriteBytes(data []byte, index *int, value []byte, numBytes int) {
	for i := 0; i < numBytes; i++ {
		data[*index] = value[i]
		*index++
	}
}

func RelayInitHandler(writer http.ResponseWriter, request *http.Request) {
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return
	}
	defer request.Body.Close()

	index := 0

	var magic uint32
	if !ReadUint32(body, &index, &magic) || magic != InitRequestMagic {
		return
	}

	var version uint32
	if !ReadUint32(body, &index, &version) || version != InitRequestVersion {
		return
	}

	var nonce []byte
	if !ReadBytes(body, &index, &nonce, crypto.NonceSize) {
		return
	}

	var relay_address string
	if !ReadString(body, &index, &relay_address, MaxRelayAddressLength) {
		return
	}

	var encrypted_token []byte
	if !ReadBytes(body, &index, &encrypted_token, RelayTokenBytes+crypto.MACSize) {
		return
	}

	if _, success := crypto.Open(encrypted_token, nonce, crypto.RelayPublicKey[:], crypto.RouterPrivateKey[:]); !success {
		return
	}

	// New redis entry
	udpAddr, err := net.ResolveUDPAddr("udp", relay_address)
	if err != nil {
		return
	}

	relay := &routing.RelayData{
		ID:             crypto.HashID(relay_address),
		Addr:           *udpAddr,
		PublicKey:      crypto.RelayPublicKey[:],
		LastUpdateTime: time.Now(),
	}

	backend.mutex.Lock()
	relayData := backend.relayMap.GetRelayData(relay.Addr.String())
	backend.mutex.Unlock()
	if relayData != nil {
		writer.WriteHeader(http.StatusConflict)
		return
	}

	backend.mutex.Lock()
	backend.relayMap.UpdateRelayData(relay.Addr.String(), relay)
	backend.dirty = true
	backend.mutex.Unlock()

	writer.Header().Set("Content-Type", "application/octet-stream")

	responseData := make([]byte, 64)
	index = 0
	WriteUint32(responseData, &index, InitResponseVersion)
	WriteUint64(responseData, &index, uint64(time.Now().Unix()))
	WriteBytes(responseData, &index, relay.PublicKey, RelayTokenBytes)
	responseData = responseData[:index]
	writer.Write(responseData)
}

func RelayUpdateHandler(writer http.ResponseWriter, request *http.Request) {
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return
	}
	defer request.Body.Close()

	index := 0

	var version uint32
	if !ReadUint32(body, &index, &version) || version != UpdateRequestVersion {
		return
	}

	var relay_address string
	if !ReadString(body, &index, &relay_address, MaxRelayAddressLength) {
		return
	}

	var token []byte
	if !ReadBytes(body, &index, &token, RelayTokenBytes) {
		return
	}

	udpAddr, err := net.ResolveUDPAddr("udp", relay_address)
	if err != nil {
		return
	}

	relay := &routing.RelayData{
		ID:             crypto.HashID(relay_address),
		Addr:           *udpAddr,
		PublicKey:      token,
		LastUpdateTime: time.Now(),
	}

	var numRelays uint32
	if !ReadUint32(body, &index, &numRelays) {
		return
	}

	if numRelays > MaxRelays {
		return
	}

	statsUpdate := &routing.RelayStatsUpdate{}
	statsUpdate.ID = relay.ID

	for i := 0; i < int(numRelays); i++ {
		var id uint64
		var rtt, jitter, packetLoss float32
		if !ReadUint64(body, &index, &id) {
			return
		}
		if !ReadFloat32(body, &index, &rtt) {
			return
		}
		if !ReadFloat32(body, &index, &jitter) {
			return
		}
		if !ReadFloat32(body, &index, &packetLoss) {
			return
		}
		ping := routing.RelayStatsPing{}
		ping.RelayID = id
		ping.RTT = rtt
		ping.Jitter = jitter
		ping.PacketLoss = packetLoss
		statsUpdate.PingStats = append(statsUpdate.PingStats, ping)
	}

	backend.mutex.Lock()
	backend.statsDatabase.ProcessStats(statsUpdate)
	backend.mutex.Unlock()

	relaysToPing := make([]routing.RelayPingData, 0)

	backend.mutex.Lock()
	allRelayData := backend.relayMap.GetAllRelayData()
	for _, v := range allRelayData {
		if v.Addr.String() != relay.Addr.String() {
			relaysToPing = append(relaysToPing, routing.RelayPingData{ID: uint64(v.ID), Address: v.Addr.String()})
		}
	}
	relayData := backend.relayMap.GetRelayData(relay.Addr.String())
	if relayData == nil {
		backend.mutex.Unlock()
		writer.WriteHeader(http.StatusNotFound)
		return
	}
	backend.relayMap.UpdateRelayData(relay.Addr.String(), relay)
	backend.mutex.Unlock()

	responseData := make([]byte, 10*1024)

	index = 0

	WriteUint32(responseData, &index, UpdateResponseVersion)

	WriteUint64(responseData, &index, uint64(time.Now().Unix()))

	WriteUint32(responseData, &index, uint32(len(relaysToPing)))

	for i := range relaysToPing {
		WriteUint64(responseData, &index, relaysToPing[i].ID)
		WriteString(responseData, &index, relaysToPing[i].Address, MaxRelayAddressLength)
	}

	responseLength := index

	responseData = responseData[:responseLength]

	writer.Header().Set("Content-Type", "application/octet-stream")

	writer.Write(responseData)
}

func NearHandler(writer http.ResponseWriter, request *http.Request) {
	backend.mutex.RLock()
	nearData := backend.nearData
	backend.mutex.RUnlock()
	writer.WriteHeader(http.StatusOK)
	writer.Header().Set("Content-Type", "application/octet-stream")
	writer.Write(nearData)
}

func WebServer() {
	router := mux.NewRouter()
	router.HandleFunc("/relay_init", RelayInitHandler).Methods("POST")
	router.HandleFunc("/relay_update", RelayUpdateHandler).Methods("POST")
	router.HandleFunc("/near", NearHandler).Methods("GET")
	http.ListenAndServe(fmt.Sprintf(":%d", NEXT_RELAY_BACKEND_PORT), router)
}
