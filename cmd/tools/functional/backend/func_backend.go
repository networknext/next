/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math"
	"math/rand"
	"net"
	"net/http"
	"os"
	"runtime"
	"sort"
	"sync"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v7"
	"github.com/gorilla/mux"

	"github.com/networknext/backend/core"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/transport"
)

const NEXT_RELAY_BACKEND_PORT = 30000
const NEXT_SERVER_BACKEND_PORT = 40000
const NEXT_BACKEND_SERVER_UPDATE_PACKET = 200
const NEXT_BACKEND_SESSION_UPDATE_PACKET = 201
const NEXT_BACKEND_SESSION_RESPONSE_PACKET = 202
const NEXT_VERSION_MAJOR = 0
const NEXT_VERSION_MINOR = 0
const NEXT_VERSION_PATCH = 0
const NEXT_MAX_PACKET_BYTES = 1500

var relayPublicKey = []byte{
	0xf5, 0x22, 0xad, 0xc1, 0xee, 0x04, 0x6a, 0xbe,
	0x7d, 0x89, 0x0c, 0x81, 0x3a, 0x08, 0x31, 0xba,
	0xdc, 0xdd, 0xb5, 0x52, 0xcb, 0x73, 0x56, 0x10,
	0xda, 0xa9, 0xc0, 0xae, 0x08, 0xa2, 0xcf, 0x5e,
}

const BACKEND_MODE_DEFAULT = 0
const BACKEND_MODE_FORCE_DIRECT = 1
const BACKEND_MODE_RANDOM = 2
const BACKEND_MODE_MULTIPATH = 3
const BACKEND_MODE_ON_OFF = 4
const BACKEND_MODE_ROUTE_SWITCHING = 5
const BACKEND_MODE_UNCOMMITTED = 6
const BACKEND_MODE_UNCOMMITTED_TO_COMMITTED = 7
const BACKEND_MODE_USER_FLAGS = 8
const BACKEND_MODE_IDEMPOTENT = 9

type Backend struct {
	mutex           sync.RWMutex
	dirty           bool
	mode            int
	serverDatabase  map[string]ServerEntry
	sessionDatabase map[uint64]SessionEntry
	statsDatabase   *routing.StatsDatabase
	costMatrix      *routing.CostMatrix
	routeMatrix     *routing.RouteMatrix
	nearData        []byte

	redisClient *redis.Client
}

var backend Backend

type ServerEntry struct {
	address    *net.UDPAddr
	publicKey  []byte
	lastUpdate int64
}

type SessionEntry struct {
	id              uint64
	sequence        uint64
	version         uint8
	expireTimestamp uint64
	route           []uint64
	next            bool
	slice           uint64
	response        []byte
}

const RTT_Threshold = 1.0
const CostMatrixBytes = 10 * 1024 * 1024
const RouteMatrixBytes = 32 * 1024 * 1024

func OptimizeThread() {
	for {
		if err := backend.statsDatabase.GetCostMatrix(backend.costMatrix, backend.redisClient); err != nil {
			fmt.Printf("error generating cost matrix: %v\n", err)
		}

		if err := backend.costMatrix.Optimize(backend.routeMatrix, RTT_Threshold); err != nil {
			fmt.Printf("error generating route matrix: %v\n", err)
		}

		time.Sleep(1 * time.Second)
	}
}

func TimeoutThread() {
	for {
		time.Sleep(time.Second * 1)
		backend.mutex.Lock()
		currentTimestamp := time.Now().Unix()
		unixTimeout := int64(routing.RelayTimeout.Seconds())
		for k, v := range backend.serverDatabase {
			if currentTimestamp-v.lastUpdate > unixTimeout {
				delete(backend.serverDatabase, k)
				backend.dirty = true
				continue
			}
		}
		for k, v := range backend.sessionDatabase {
			if uint64(currentTimestamp) >= v.expireTimestamp {
				delete(backend.sessionDatabase, k)
				backend.dirty = true
				continue
			}
		}
		hgetallResult := backend.redisClient.HGetAll(routing.HashKeyAllRelays)
		for _, raw := range hgetallResult.Val() {
			var r routing.Relay
			r.UnmarshalBinary([]byte(raw))
			if currentTimestamp-int64(r.LastUpdateTime) > unixTimeout {
				fmt.Println("Deleting redis relay")
				backend.redisClient.HDel(routing.HashKeyAllRelays, r.Key())
				backend.dirty = true
				continue
			}
		}

		if backend.dirty {
			fmt.Printf("-----------------------------\n")
			hgetallResult := backend.redisClient.HGetAll(routing.HashKeyAllRelays)
			for _, raw := range hgetallResult.Val() {
				var r routing.Relay
				r.UnmarshalBinary([]byte(raw))
				fmt.Printf("relay: %v\n", &r.Addr)
			}

			for _, v := range backend.serverDatabase {
				fmt.Printf("server: %s\n", v.address)
			}
			for k := range backend.sessionDatabase {
				fmt.Printf("session: %x\n", k)
			}
			if len(hgetallResult.Val()) == 0 && len(backend.serverDatabase) == 0 {
				fmt.Printf("No relay or server entries\n")
			}
			backend.dirty = false
		}
		backend.mutex.Unlock()
	}
}

func (backend *Backend) GetNearRelays() []routing.Relay {
	var nearRelays = make([]routing.Relay, 0)

	// Get relays
	hgetallResult := backend.redisClient.HGetAll(routing.HashKeyAllRelays)
	for _, raw := range hgetallResult.Val() {
		var r routing.Relay
		r.UnmarshalBinary([]byte(raw))
		nearRelays = append(nearRelays, r)
	}

	// sort them by ID for consistency
	sort.SliceStable(nearRelays[:], func(i, j int) bool { return nearRelays[i].ID < nearRelays[j].ID })

	// Clamp relay count to max // TODO: actually do this in real backend?
	if len(nearRelays) > int(core.MaxNearRelays) {
		nearRelays = nearRelays[:core.MaxNearRelays]
	}

	return nearRelays
}

func RouteChanged(previous []uint64, current []uint64) bool {
	if len(previous) != len(current) {
		return true
	}
	for i := range current {
		if current[i] != previous[i] {
			return true
		}
	}
	return false
}

func main() {

	rand.Seed(time.Now().UnixNano())

	backend.serverDatabase = make(map[string]ServerEntry)
	backend.sessionDatabase = make(map[uint64]SessionEntry)
	backend.statsDatabase = routing.NewStatsDatabase()
	backend.costMatrix = &routing.CostMatrix{}
	backend.routeMatrix = &routing.RouteMatrix{}

	redisServer, err := miniredis.Run()
	if err != nil {
		fmt.Printf("failed to run redis, err: %v", err)
		return
	}
	backend.redisClient = redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

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

	if os.Getenv("BACKEND_MODE") == "IDEMPOTENT" {
		backend.mode = BACKEND_MODE_IDEMPOTENT
	}

	go OptimizeThread()

	go TimeoutThread()

	go WebServer()

	listenAddress := net.UDPAddr{
		Port: NEXT_SERVER_BACKEND_PORT,
		IP:   net.ParseIP("0.0.0.0"),
	}

	fmt.Printf("started functional backend on ports %d and %d\n", NEXT_RELAY_BACKEND_PORT, NEXT_SERVER_BACKEND_PORT)

	connection, err := net.ListenUDP("udp", &listenAddress)
	if err != nil {
		fmt.Printf("error: could not listen on %s\n", listenAddress.String())
		return
	}

	defer connection.Close()

	mux := transport.UDPServerMux{
		Conn:          connection,
		MaxPacketSize: transport.DefaultMaxPacketSize,

		ServerUpdateHandlerFunc: func(w io.Writer, incoming *transport.UDPPacket) {
			serverUpdate := &transport.ServerUpdatePacket{}
			if err = serverUpdate.UnmarshalBinary(incoming.Data); err != nil {
				fmt.Printf("error: failed to read server update packet: %v\n", err)
				return
			}

			serverEntry := ServerEntry{}
			serverEntry.address = incoming.SourceAddr
			serverEntry.publicKey = serverUpdate.ServerRoutePublicKey
			serverEntry.lastUpdate = time.Now().Unix()

			key := string(incoming.SourceAddr.String())

			backend.mutex.Lock()
			_, ok := backend.serverDatabase[key]
			if !ok {
				backend.dirty = true
			}
			backend.serverDatabase[key] = serverEntry
			backend.mutex.Unlock()
		},

		SessionUpdateHandlerFunc: func(w io.Writer, incoming *transport.UDPPacket) {
			sessionUpdate := &transport.SessionUpdatePacket{}
			if err = sessionUpdate.UnmarshalBinary(incoming.Data); err != nil {
				fmt.Printf("error: failed to read server session update packet: %v\n", err)
				return
			}

			if sessionUpdate.FallbackToDirect {
				fmt.Printf("error: fallback to direct %s\n", incoming.SourceAddr)
				return
			}

			backend.mutex.RLock()
			serverEntry, ok := backend.serverDatabase[string(incoming.SourceAddr.String())]
			backend.mutex.RUnlock()
			if !ok {
				fmt.Printf("error: could not find server %s\n", incoming.SourceAddr)
				return
			}

			nearRelays := backend.GetNearRelays()

			var sessionResponse *transport.SessionResponsePacket

			backend.mutex.RLock()
			sessionEntry, ok := backend.sessionDatabase[sessionUpdate.SessionID]
			backend.mutex.RUnlock()

			newSession := !ok

			if newSession {
				sessionEntry.id = sessionUpdate.SessionID
				sessionEntry.version = 0
				sessionEntry.expireTimestamp = uint64(time.Now().Unix()) + 20
			} else {
				sessionEntry.expireTimestamp += 10
				sessionEntry.slice++
			}

			if backend.mode == BACKEND_MODE_IDEMPOTENT {
				if sessionUpdate.Sequence == sessionEntry.sequence {
					_, err = w.Write(sessionEntry.response)
					if err != nil {
						fmt.Printf("error: failed to send udp response: %v\n", err)
					}
					return
				}
			}

			takeNetworkNext := len(nearRelays) > 0

			if backend.mode == BACKEND_MODE_FORCE_DIRECT {
				takeNetworkNext = false
			}

			if backend.mode == BACKEND_MODE_RANDOM {
				takeNetworkNext = takeNetworkNext && rand.Float32() > 0.5
			}

			if backend.mode == BACKEND_MODE_ON_OFF {
				if (sessionEntry.slice & 1) == 0 {
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
				committed = sessionUpdate.Sequence > 2
				if sessionUpdate.Sequence <= 2 && sessionUpdate.Committed == true {
					panic("slices 0,1,2,3 should not be committed")
				}
				if sessionUpdate.Sequence >= 4 && sessionUpdate.Committed == false {
					panic("slices 4 and greater should be committed")
				}
			}

			if backend.mode == BACKEND_MODE_USER_FLAGS {
				if sessionUpdate.Sequence >= 2 && sessionUpdate.UserFlags != 0x123 {
					panic("user flags not set on session update")
				}
			}

			// Extract ids and addresses into own list to make response
			var nearRelayIDs = make([]uint64, 0)
			var nearRelayAddresses = make([]net.UDPAddr, 0)
			for _, relay := range nearRelays {
				nearRelayIDs = append(nearRelayIDs, relay.ID)
				nearRelayAddresses = append(nearRelayAddresses, relay.Addr)
			}

			if !takeNetworkNext {

				// direct route
				sessionResponse = &transport.SessionResponsePacket{
					Sequence:             sessionUpdate.Sequence,
					SessionID:            sessionUpdate.SessionID,
					NumNearRelays:        int32(len(nearRelays)),
					NearRelayIDs:         nearRelayIDs,
					NearRelayAddresses:   nearRelayAddresses,
					RouteType:            int32(core.NEXT_UPDATE_TYPE_DIRECT),
					NumTokens:            0,
					Tokens:               nil,
					ServerRoutePublicKey: serverEntry.publicKey,
				}

				sessionEntry.route = nil
				sessionEntry.next = false

			} else {

				// next route

				numRelays := len(nearRelayIDs)
				if numRelays > 5 {
					numRelays = 5
				}

				route := make([]uint64, numRelays)
				for i := 0; i < numRelays; i++ {
					route[i] = nearRelayIDs[i]
				}

				routeChanged := RouteChanged(sessionEntry.route, route)

				numNodes := numRelays + 2

				addresses := make([]*net.UDPAddr, numNodes)
				publicKeys := make([][]byte, numNodes)
				publicKeys[0] = sessionUpdate.ClientRoutePublicKey[:]

				for i := 0; i < numRelays; i++ {
					addresses[1+i] = &nearRelayAddresses[i]
					publicKeys[1+i] = crypto.RelayPublicKey[:]
				}

				addresses[numNodes-1] = incoming.SourceAddr
				publicKeys[numNodes-1] = serverEntry.publicKey

				var tokens []byte

				var responseType int32

				if !sessionEntry.next || routeChanged {

					// new route

					sessionEntry.version++
					tokens, err = core.WriteRouteTokens(sessionEntry.expireTimestamp, sessionEntry.id, sessionEntry.version, 0, 256, 256, numNodes, addresses, publicKeys, crypto.RouterPrivateKey)
					if err != nil {
						fmt.Printf("error: could not write route tokens: %v\n", err)
						return
					}
					responseType = core.NEXT_UPDATE_TYPE_ROUTE

				} else {

					// continue route

					tokens, err = core.WriteContinueTokens(sessionEntry.expireTimestamp, sessionEntry.id, sessionEntry.version, 0, numNodes, publicKeys, crypto.RouterPrivateKey)
					if err != nil {
						fmt.Printf("error: could not write continue tokens: %v\n", err)
						return
					}
					responseType = core.NEXT_UPDATE_TYPE_CONTINUE

				}

				sessionResponse = &transport.SessionResponsePacket{
					Sequence:             sessionUpdate.Sequence,
					SessionID:            sessionUpdate.SessionID,
					NumNearRelays:        int32(len(nearRelayIDs)),
					NearRelayIDs:         nearRelayIDs,
					NearRelayAddresses:   nearRelayAddresses,
					RouteType:            responseType,
					Multipath:            multipath,
					Committed:            committed,
					NumTokens:            int32(numNodes),
					Tokens:               tokens,
					ServerRoutePublicKey: serverEntry.publicKey,
				}

				sessionEntry.route = route
				sessionEntry.next = true
			}

			if sessionResponse == nil {
				fmt.Printf("error: nil session response\n")
				return
			}

			sessionResponse.Signature = crypto.Sign(crypto.BackendPrivateKey, sessionResponse.GetSignData())

			responsePacketData, err := sessionResponse.MarshalBinary()
			if err != nil {
				fmt.Printf("error: failed to write session response packet: %v\n", err)
				return
			}

			sessionEntry.sequence = sessionResponse.Sequence
			sessionEntry.response = responsePacketData

			backend.mutex.Lock()
			if newSession {
				backend.dirty = true
			}
			backend.sessionDatabase[sessionUpdate.SessionID] = sessionEntry
			backend.mutex.Unlock()

			_, err = w.Write(responsePacketData)
			if err != nil {
				fmt.Printf("error: failed to send udp response: %v\n", err)
				return
			}
		},
	}

	if err := mux.Start(context.Background(), runtime.NumCPU()); err != nil {
		log.Fatalf("failed to start udp server: %v", err)
	}
}

// -----------------------------------------------------------

const InitRequestMagic = uint32(0x9083708f)
const InitRequestVersion = 0
const InitResponseVersion = 0
const UpdateRequestVersion = 0
const UpdateResponseVersion = 0
const MaxRelayIDLength = 256
const MaxRelayAddressLength = 256
const RelayTokenBytes = 32
const MaxRelays = 1024

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
	relay := routing.Relay{
		ID:             crypto.HashID(relay_address),
		Addr:           *udpAddr,
		PublicKey:      core.RandomBytes(RelayTokenBytes),
		LastUpdateTime: uint64(time.Now().Unix()),
	}

	if backend.redisClient.HExists(routing.HashKeyAllRelays, relay.Key()).Val() {
		writer.WriteHeader(http.StatusConflict)
		return
	}

	backend.redisClient.HSet(routing.HashKeyAllRelays, relay.Key(), relay)
	backend.dirty = true

	writer.Header().Set("Content-Type", "application/octet-stream")

	responseData := make([]byte, 64)
	index = 0
	WriteUint32(responseData, &index, InitResponseVersion)
	WriteUint64(responseData, &index, uint64(time.Now().Unix()))
	WriteBytes(responseData, &index, relay.PublicKey, RelayTokenBytes)
	responseData = responseData[:index]
	writer.Write(responseData)
}

func CompareTokens(a []byte, b []byte) bool {
	if len(a) != len(b) {
		fmt.Printf("token length is wrong\n")
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			fmt.Printf("token value is wrong: %d vs. %d\n", a[i], b[i])
			return false
		}
	}
	return true
}

func RelayUpdateHandler(writer http.ResponseWriter, request *http.Request) {
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return
	}

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
	relay := routing.Relay{
		ID:             crypto.HashID(relay_address),
		Addr:           *udpAddr,
		PublicKey:      token,
		LastUpdateTime: uint64(time.Now().Unix()),
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

	backend.statsDatabase.ProcessStats(statsUpdate)
	relaysToPing := make([]routing.RelayPingData, 0)

	hgetallResult := backend.redisClient.HGetAll(routing.HashKeyAllRelays)
	for k, v := range hgetallResult.Val() {
		if k != relay.Key() {
			var unmarshaledValue routing.Relay
			unmarshaledValue.UnmarshalBinary([]byte(v))
			relaysToPing = append(relaysToPing, routing.RelayPingData{ID: uint64(unmarshaledValue.ID), Address: unmarshaledValue.Addr.String()})
		}
	}

	if !backend.redisClient.HExists(routing.HashKeyAllRelays, relay.Key()).Val() {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	backend.redisClient.HSet(routing.HashKeyAllRelays, relay.Key(), relay)

	// Back to old code
	responseData := make([]byte, 10*1024)

	index = 0

	WriteUint32(responseData, &index, UpdateResponseVersion)

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
	router.Handle("/cost_matrix", backend.costMatrix).Methods("GET")
	router.Handle("/route_matrix", backend.routeMatrix).Methods("GET")
	router.HandleFunc("/near", NearHandler).Methods("GET")
	http.ListenAndServe(fmt.Sprintf(":%d", NEXT_RELAY_BACKEND_PORT), router)
}
