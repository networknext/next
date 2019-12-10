/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

// #cgo pkg-config: libsodium
// #include <sodium.h>
import "C"

import (
	"os"
	"net"
	"fmt"
	"time"
	"sync"
	"sort"
	"net/http"
	"math/rand"
	"hash/fnv"
	"io/ioutil"
	"encoding/json"
	"encoding/binary"
	"github.com/gorilla/mux"
	"github.com/networknext/backend/core"
)

const NEXT_BACKEND_PORT = 40001
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

type Backend struct {
	mutex sync.RWMutex
	dirty bool
	mode int
	relayDatabase map[string]RelayEntry
	serverDatabase map[string]ServerEntry
	sessionDatabase map[uint64]SessionEntry
}

var backend Backend

type RelayEntry struct {
	id uint64
	name string
	address *net.UDPAddr
	lastUpdate int64
}

type ServerEntry struct {
	address *net.UDPAddr
	publicKey []byte
	lastUpdate int64
}

type SessionEntry struct {
	id 				uint64
	version 		uint8
	expireTimestamp uint64
	route           []uint64
	next 			bool
	slice           uint64
}

func TimeoutThread() {
	for {
		time.Sleep(time.Second*1)
		backend.mutex.Lock()
		currentTimestamp := time.Now().Unix()
		for k,v := range backend.relayDatabase {
			if currentTimestamp - v.lastUpdate > 15 {
				backend.dirty = true
				delete(backend.relayDatabase, k)
				continue
			}
		}
		for k,v := range backend.serverDatabase {
			if currentTimestamp - v.lastUpdate > 15 {
				delete(backend.serverDatabase, k)
				backend.dirty = true
				continue
			}
		}
		for k,v := range backend.sessionDatabase {
			if uint64(currentTimestamp) >= v.expireTimestamp {
				delete(backend.sessionDatabase, k)
				backend.dirty = true
				continue
			}
		}
		if backend.dirty {
			fmt.Printf( "-----------------------------\n" )
			for _,v := range backend.relayDatabase {
				fmt.Printf( "relay: %s\n", v.address );
			}
			for _,v := range backend.serverDatabase {
				fmt.Printf( "server: %s\n", v.address );
			}
			for k,_ := range backend.sessionDatabase {
				fmt.Printf( "session: %x\n", k );
			}
			if len(backend.relayDatabase) == 0 && len(backend.serverDatabase) == 0 {
				fmt.Printf( "(nil)\n" )
			}
			backend.dirty = false
		}
		backend.mutex.Unlock()
	}
}

func GetRelayId(name string) uint32 {
	hash := fnv.New32a()
	hash.Write([]byte(name))
	return hash.Sum32()
}

func GetNearRelays() ([]uint64, []net.UDPAddr) {
	nearRelays := make([]RelayEntry, 0)
	backend.mutex.RLock()
	for _,v := range backend.relayDatabase {
		nearRelays = append(nearRelays, v)
	}
	backend.mutex.RUnlock()
	sort.SliceStable(nearRelays[:], func(i, j int) bool { return nearRelays[i].id < nearRelays[j].id })
	if len(nearRelays) > int(core.MaxNearRelays) {
		nearRelays = nearRelays[:core.MaxNearRelays]
	}
	nearRelayIds := make([]uint64, len(nearRelays))
	nearRelayAddresses := make([]net.UDPAddr, len(nearRelays))
	for i := range nearRelays {
		nearRelayIds[i] = nearRelays[i].id
		nearRelayAddresses[i] = *nearRelays[i].address
	}
	return nearRelayIds, nearRelayAddresses
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

	backend.relayDatabase = make(map[string]RelayEntry)
	backend.serverDatabase = make(map[string]ServerEntry)
	backend.sessionDatabase = make(map[uint64]SessionEntry)

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

	go TimeoutThread()

	go TerribleOldShite()

	go WebServer()

	listenAddress := net.UDPAddr{
    	Port: NEXT_BACKEND_PORT,
    	IP:   net.ParseIP("0.0.0.0"),
	}

	connection, err := net.ListenUDP("udp", &listenAddress)
	if err != nil {
		fmt.Printf( "error: could not listen on %s\n", listenAddress.String())
		return
	}

	defer connection.Close()

	fmt.Printf( "started local backend on port %d\n", NEXT_BACKEND_PORT )

	packetData := make([]byte, NEXT_MAX_PACKET_BYTES)

	for {

		packetBytes, from, _ := connection.ReadFromUDP(packetData)

		if packetBytes <= 0 {
			continue
		}

		packetType := packetData[0]

		if packetType == NEXT_BACKEND_SERVER_UPDATE_PACKET {

			readStream := core.CreateReadStream(packetData[1:])

			serverUpdate := &core.NextBackendServerUpdatePacket{}

			if err := serverUpdate.Serialize(readStream); err != nil {
				fmt.Printf( "error: failed to read server update packet: %v\n", err)
				continue
			}

			serverEntry := ServerEntry{}
			serverEntry.address = from
			serverEntry.publicKey = serverUpdate.ServerRoutePublicKey
			serverEntry.lastUpdate = time.Now().Unix()

			key := string(from.String())

			backend.mutex.Lock()
			_, ok := backend.serverDatabase[key]
			if !ok {
				backend.dirty = true
			}
			backend.serverDatabase[key] = serverEntry
			backend.mutex.Unlock()

		} else if packetType == NEXT_BACKEND_SESSION_UPDATE_PACKET {

			readStream := core.CreateReadStream(packetData[1:])
			sessionUpdate := &core.NextBackendSessionUpdatePacket{}
			if err := sessionUpdate.Serialize(readStream, NEXT_VERSION_MAJOR, NEXT_VERSION_MINOR, NEXT_VERSION_PATCH); err != nil {
				fmt.Printf("error: failed to read server session update packet: %v\n", err)
				continue
			}

			if sessionUpdate.FallbackToDirect {
				fmt.Printf( "error: fallback to direct %s\n", from )
				continue
			}

			backend.mutex.RLock()
			serverEntry, ok := backend.serverDatabase[string(from.String())]
			backend.mutex.RUnlock()
			if !ok {
				fmt.Printf( "error: could not find server %s\n", from )
				continue
			}

			nearRelayIds, nearRelayAddresses := GetNearRelays()

			var sessionResponse *core.NextBackendSessionResponsePacket

			backend.mutex.RLock()
			sessionEntry, ok := backend.sessionDatabase[sessionUpdate.SessionId]
			backend.mutex.RUnlock()

			newSession := !ok

			if newSession {
				sessionEntry.id = sessionUpdate.SessionId
				sessionEntry.version = 0
                sessionEntry.expireTimestamp = uint64(time.Now().Unix()) + 20
			} else {
				sessionEntry.expireTimestamp += 10		
				sessionEntry.slice++		
			}

			takeNetworkNext := len(nearRelayIds) > 0

			if backend.mode == BACKEND_MODE_FORCE_DIRECT {
				takeNetworkNext = false
			}

			if backend.mode == BACKEND_MODE_RANDOM {
				takeNetworkNext = takeNetworkNext && rand.Float32() > 0.5
			}

			if backend.mode == BACKEND_MODE_ON_OFF {
				if ( sessionEntry.slice & 1 ) == 0 {
					takeNetworkNext = false
				}
			}

			if backend.mode == BACKEND_MODE_ROUTE_SWITCHING {
				rand.Shuffle(len(nearRelayIds), func(i, j int) { nearRelayIds[i], nearRelayIds[j] = nearRelayIds[j], nearRelayIds[i]; nearRelayAddresses[i], nearRelayAddresses[j] = nearRelayAddresses[j], nearRelayAddresses[i] })
			}

			multipath := len(nearRelayIds) > 0 && backend.mode == BACKEND_MODE_MULTIPATH

			if !takeNetworkNext {

				// direct route

				sessionResponse = &core.NextBackendSessionResponsePacket{
					Sequence:             sessionUpdate.Sequence,
					SessionId:            sessionUpdate.SessionId,
					NumNearRelays:        int32(len(nearRelayIds)),
					NearRelayIds:         nearRelayIds,
					NearRelayAddresses:   nearRelayAddresses,
					ResponseType:         int32(core.NEXT_UPDATE_TYPE_DIRECT),
					NumTokens:            0,
					Tokens:               nil,
					ServerRoutePublicKey: serverEntry.publicKey,
					Signature:            nil,
				}

				sessionEntry.route = nil
				sessionEntry.next = false

			} else {

				// next route

				numRelays := len(nearRelayIds)
				if ( numRelays > 5 ) {
					numRelays = 5
				}

				route := make([]uint64, numRelays)
				for i := 0; i < numRelays; i++ {
					route[i] = nearRelayIds[i]
				}

				routeChanged := RouteChanged(sessionEntry.route, route)

				numNodes := numRelays + 2

				addresses := make([]*net.UDPAddr, numNodes)
				publicKeys := make([][]byte, numNodes)
				publicKeys[0] = sessionUpdate.ClientRoutePublicKey[:]

				for i := 0; i < numRelays; i++ {
					addresses[1+i] = &nearRelayAddresses[i]
					publicKeys[1+i] = relayPublicKey
				}

				addresses[numNodes-1] = from
				publicKeys[numNodes-1] = serverEntry.publicKey

				var tokens []byte

				var responseType int32

				if !sessionEntry.next || routeChanged {

					// new route

					sessionEntry.version++
					tokens, err = core.WriteRouteTokens(sessionEntry.expireTimestamp, sessionEntry.id, sessionEntry.version, 0, 256, 256, numNodes, addresses, publicKeys, core.RouterPrivateKey)
					if err != nil {
						fmt.Printf("error: could not write route tokens: %v\n", err)
						continue
					}
					responseType = core.NEXT_UPDATE_TYPE_ROUTE

				} else {

					// continue route

					tokens, err = core.WriteContinueTokens(sessionEntry.expireTimestamp, sessionEntry.id, sessionEntry.version, 0, numNodes, publicKeys, core.RouterPrivateKey)
					if err != nil {
						fmt.Printf("error: could not write continue tokens: %v\n", err)
						continue
					}
					responseType = core.NEXT_UPDATE_TYPE_CONTINUE

				}

				sessionResponse = &core.NextBackendSessionResponsePacket{
					Sequence:             sessionUpdate.Sequence,
					SessionId:            sessionUpdate.SessionId,
					NumNearRelays:        int32(len(nearRelayIds)),
					NearRelayIds:         nearRelayIds,
					NearRelayAddresses:   nearRelayAddresses,
					ResponseType:         responseType,
					Multipath:            multipath,
					NumTokens:            int32(numNodes),
					Tokens:               tokens,
					ServerRoutePublicKey: serverEntry.publicKey,
					Signature:            nil,
				}

				sessionEntry.route = route
				sessionEntry.next = true
			}

			if sessionResponse == nil {
				fmt.Printf( "error: nil session response\n" );
				continue
			}

			backend.mutex.Lock()
			if newSession {
				backend.dirty = true
			}
			backend.sessionDatabase[sessionUpdate.SessionId] = sessionEntry
			backend.mutex.Unlock()

			signResponseData := sessionResponse.GetSignData(NEXT_VERSION_MAJOR, NEXT_VERSION_MINOR, NEXT_VERSION_PATCH)

			sessionResponse.Signature = core.CryptoSignCreate(signResponseData, core.BackendPrivateKey)
			if sessionResponse.Signature == nil {
				fmt.Printf("error: failed to sign session response packet")
				continue
			}

			writeStream, err := core.CreateWriteStream(NEXT_MAX_PACKET_BYTES)
			if err != nil {
				fmt.Printf("error: failed to write session response packet: %v\n", err)
				continue
			}
			responsePacketType := uint32(NEXT_BACKEND_SESSION_RESPONSE_PACKET)
			writeStream.SerializeBits(&responsePacketType, 8)
			if err := sessionResponse.Serialize(writeStream, NEXT_VERSION_MAJOR, NEXT_VERSION_MINOR, NEXT_VERSION_PATCH); err != nil {
				fmt.Printf("error: failed to write session response packet: %v\n", err)
				continue
			}
			writeStream.Flush()

			responsePacketData := writeStream.GetData()[0:writeStream.GetBytesProcessed()]

			_, err = connection.WriteToUDP(responsePacketData, from)
			if err != nil {
				fmt.Printf("error: failed to send udp response: %v\n", err)
				continue
			}

		}
	}
}

// -----------------------------------------------------------

const InitMagic = uint32(0x9083708f)
const InitVersion = 0
const MaxRelayIdLength = 256
const MaxRelayAddressLength = 256

func ReadUint32(data []byte, index *int, value *uint32) bool {
	if *index + 4 > len(data) {
		return false
	}
	*value = binary.LittleEndian.Uint32(data[*index:])
	*index += 4
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
	if *index + int(stringLength) > len(data) {
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

func GetRelayPublicKey(relay_id string) []byte {
	if relay_id == "local" {
		return []byte{0x06, 0xb0, 0x4d, 0x9e, 0xa6, 0xf5, 0x7c, 0x0b, 0x3c, 0x6a, 0x2d, 0x9d, 0xbf, 0x34, 0x32, 0xb6, 0x66, 0x00, 0xa0, 0x3b, 0x2b, 0x5b, 0x5d, 0x00, 0x91, 0x4a, 0x32, 0xee, 0xf2, 0x36, 0xc2, 0x9c}
	} else {
		return []byte{}
	}
}

func GetRelayAddress(relay_id string) string {
	if relay_id == "local" {
		return "127.0.0.1:50000"
	} else {
		return ""
	}
}

func SignatureCheck(data []byte, publicKey []byte) bool {
    dummy := make([]byte, len(data))
    dummy_length := uint64(0)
	if C.crypto_sign_open((*C.uchar)(&dummy[0]), (*C.ulonglong)(&dummy_length), (*C.uchar)(&data[0]), C.ulonglong(len(data)), (*C.uchar)(&publicKey[0])) != 0 {
		return false
	}
	return true
}

func RelayInitHandler(writer http.ResponseWriter, request *http.Request) {

    body, err := ioutil.ReadAll(request.Body)
    if err != nil {
        return
    }

	index := C.crypto_sign_BYTES

	var magic uint32
	if !ReadUint32(body, &index, &magic) || magic != InitMagic {
		return
	}

	var version uint32
	if !ReadUint32(body, &index, &version) || version != InitVersion {
		return
	}

	var relay_id string
	if !ReadString(body, &index, &relay_id, MaxRelayIdLength) {
		return
	}

	relay_public_key := GetRelayPublicKey(relay_id)
	if len(relay_public_key) == 0 {
		return
	}

	if !SignatureCheck(body, relay_public_key) {
		return
	}

	var relay_address string
	if !ReadString(body, &index, &relay_address, MaxRelayAddressLength) {
		return
	}

	if relay_address != GetRelayAddress(relay_id) {
		return
	}

	key := relay_address

	backend.mutex.RLock()
	_, relayAlreadyExists := backend.relayDatabase[key]
	backend.mutex.RUnlock()

	if relayAlreadyExists {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	fmt.Printf("%s: %s\n", relay_id, relay_address)

	relayEntry := RelayEntry{}
	relayEntry.name = relay_address
	relayEntry.id = uint64(GetRelayId(relay_address))			// todo: not strictly correct, but works for now
	relayEntry.address = core.ParseAddress(relay_address)
	relayEntry.lastUpdate = time.Now().Unix()

	backend.mutex.Lock()
	backend.relayDatabase[key] = relayEntry
	backend.dirty = true
	backend.mutex.Unlock()

	// todo: we actually need to send a nonce down (random 8 bytes), to be passed back up

	writer.WriteHeader(http.StatusOK)
}

func WebServer() {
	router := mux.NewRouter()
	router.HandleFunc("/relay_init", RelayInitHandler).Methods("POST")
	http.ListenAndServe(":30000", router)
}

// -----------------------------------------------------------

const NEXT_PACKET_TYPE_RELAY_INIT_REQUEST = 43
const NEXT_PACKET_TYPE_RELAY_INIT_RESPONSE = 52
const NEXT_PACKET_TYPE_RELAY_CONFIG_REQUEST  = 50
const NEXT_PACKET_TYPE_RELAY_CONFIG_RESPONSE = 51
const NEXT_PACKET_TYPE_RELAY_REPORT = 48

var MasterTokenSignKey = []byte{
	0x15, 0xa0, 0x59, 0x84, 0x51, 0x1e, 0xf7, 0x96,
	0xed, 0x4b, 0x82, 0xd2, 0x44, 0xec, 0x04, 0x65,
	0x0c, 0x55, 0x71, 0xa0, 0xfd, 0xf8, 0x0a, 0xc3,
	0x64, 0x90, 0x0f, 0x16, 0x24, 0xb7, 0x8f, 0x3a,
}

var MasterUDPSignPrivateKey = []byte{
	0x84, 0xc7, 0x24, 0xfa, 0x5f, 0x94, 0x86, 0x99,
	0x0d, 0x22, 0x40, 0xaf, 0xa1, 0x62, 0x8c, 0x24,
	0x51, 0xef, 0xfc, 0x10, 0x6f, 0xef, 0x04, 0xb3,
	0x50, 0x9b, 0xbc, 0xb0, 0xce, 0xcb, 0xc3, 0x03,
	0x60, 0x45, 0x96, 0x52, 0x4f, 0x1c, 0x00, 0xda,
	0x35, 0x1b, 0x6c, 0x17, 0x8b, 0xa8, 0xaa, 0xac,
	0xb4, 0x8c, 0x76, 0xb1, 0x72, 0xa6, 0xfa, 0x7f,
	0x52, 0x28, 0xd8, 0x6d, 0x9e, 0x2b, 0x91, 0xec,
}
var MasterUDPSignPublicKey = []byte{
	0x60, 0x45, 0x96, 0x52, 0x4f, 0x1c, 0x00, 0xda,
	0x35, 0x1b, 0x6c, 0x17, 0x8b, 0xa8, 0xaa, 0xac,
	0xb4, 0x8c, 0x76, 0xb1, 0x72, 0xa6, 0xfa, 0x7f,
	0x52, 0x28, 0xd8, 0x6d, 0x9e, 0x2b, 0x91, 0xec,
}
var MasterUDPSealPrivateKey = []byte{
	0xb7, 0xca, 0x67, 0x4b, 0x12, 0xe7, 0x6a, 0x19,
	0xab, 0x69, 0xbc, 0x32, 0x31, 0xf9, 0x9b, 0x29,
	0x49, 0xe8, 0xa9, 0x5b, 0x7e, 0xb6, 0xe8, 0x4c,
	0x8a, 0x8a, 0x9e, 0xb3, 0xc2, 0x7b, 0x1f, 0x98,
}
var MasterUDPSealPublicKey = []byte{
	0x77, 0x9f, 0xf2, 0xeb, 0x45, 0xfb, 0xe8, 0x25,
	0x7a, 0xf3, 0x78, 0xf9, 0x26, 0x22, 0x29, 0xc0,
	0xa8, 0xd0, 0x66, 0x92, 0x8b, 0xf9, 0x47, 0xcc,
	0x8b, 0x93, 0x62, 0xbe, 0xb3, 0x88, 0xf9, 0x6f,
}

const (
	ADDRESS_NONE = 0
	ADDRESS_IPV4 = 1
	ADDRESS_IPV6 = 2
)

func WriteAddress(buffer []byte, address *net.UDPAddr) {
	if address == nil {
		buffer[0] = ADDRESS_NONE
		return
	}
	ipv4 := address.IP.To4()
	port := address.Port
	if ipv4 != nil {
		buffer[0] = ADDRESS_IPV4
		buffer[1] = ipv4[0]
		buffer[2] = ipv4[1]
		buffer[3] = ipv4[2]
		buffer[4] = ipv4[3]
		buffer[5] = (byte)(port & 0xFF)
		buffer[6] = (byte)(port >> 8)
	} else {
		buffer[0] = ADDRESS_IPV6
		copy(buffer[1:], address.IP)
		buffer[17] = (byte)(port & 0xFF)
		buffer[18] = (byte)(port >> 8)
	}
}

func WriteMasterToken(buffer []byte, address *net.UDPAddr) error {
	if len(buffer) < MasterTokenBytes {
		return fmt.Errorf("expected %d byte buffer, got %d bytes", MasterTokenBytes, len(buffer))
	}
	WriteAddress(buffer, address)
	hmac, err := CryptoAuth(buffer[0:AddressBytes], MasterTokenSignKey)
	if err != nil {
		return fmt.Errorf("failed to sign master token: %v", err)
	}
	if len(hmac) != 32 {
		panic("wrong hmac size")
	}
	copy(buffer[AddressBytes:], hmac[:])
	return nil
}

func CryptoAuth(data []byte, key []byte) ([]byte, error) {
	if len(key) != C.crypto_auth_KEYBYTES {
		return nil, fmt.Errorf("expected %d byte key, got %d bytes", C.crypto_auth_KEYBYTES, len(key))
	}
	var signature [C.crypto_auth_BYTES]byte
	if C.crypto_auth((*C.uchar)(&signature[0]), (*C.uchar)(&data[0]), (C.ulonglong)(len(data)), (*C.uchar)(&key[0])) != 0 {
		return nil, fmt.Errorf("failed to sign data with key")
	}
	return signature[:], nil
}

type InitResponseJSON struct {
	Timestamp   uint64
	IP          []byte
	IP2Location string
	Token       []byte
}

type RelayConfigRequest struct {
	RelayId   uint64
	Timestamp uint64
	Signature []byte
}

type RelayJSON struct {
	Name              string
	UpdateKey         []byte
	Group             string
	Role              string
	State             int
	Address           string
	ManagementAddress string
}

func TerribleOldShite() {

	listener := UDPListenerMasterCreate(MasterTokenSignKey, MasterUDPSealPublicKey, MasterUDPSealPrivateKey)

	builder := UDPPacketToClientBuilderCreate(MasterUDPSignPrivateKey)

	var packetsReceivedCount int64

	go listener.Listen(
		&packetsReceivedCount,
		func(packet *UDPPacketToMaster, from *net.UDPAddr, conn *net.UDPConn) error {

			if packet.Type == NEXT_PACKET_TYPE_RELAY_INIT_REQUEST {

				var token [MasterTokenBytes]byte
				err := WriteMasterToken(token[:], &net.UDPAddr{IP: from.IP, Port: 0})
				if err != nil {
					return fmt.Errorf("could not write master token: %v", err)
					return nil
				}

				response := &InitResponseJSON{
					Timestamp:   uint64(time.Now().UnixNano() / 1000000), // milliseconds
					IP2Location: "",
					IP:          []byte(from.String()),
					Token:       token[:],
				}

				responseData, _ := json.Marshal(response)

				packets, err := builder.Build(&UDPPacketToClient{Type: NEXT_PACKET_TYPE_RELAY_INIT_RESPONSE, ID: packet.ID, Status: uint16(200), Data: responseData})
				if err != nil {
					return fmt.Errorf("could not build relay init response packet: %v", err)
				}

				for _, packet := range packets {
					conn.WriteToUDP(packet, from)
				}
				return nil

			} else if packet.Type == NEXT_PACKET_TYPE_RELAY_CONFIG_REQUEST {

				var request RelayConfigRequest
				if err := json.Unmarshal(packet.Data, &request); err != nil {
					fmt.Printf("could not parse relay config request json: %s", err)
					return nil
				}

				response := &RelayJSON{
					Name:              "local",
					UpdateKey:         make([]byte, 32),
					Group:             "local",
					Role:              "default",
					State:             0,
					Address:           from.String(),
					ManagementAddress: from.String(),
				}

				responseData, _ := json.Marshal(response)

				packets, err := builder.Build(&UDPPacketToClient{Type: NEXT_PACKET_TYPE_RELAY_CONFIG_RESPONSE, ID: packet.ID, Status: uint16(200), Data: responseData})
				if err != nil {
					return fmt.Errorf("could not build relay config response packet: %v", err)
				}

				for _, packet := range packets {
					conn.WriteToUDP(packet, from)
				}
				return nil

			} else if packet.Type == NEXT_PACKET_TYPE_RELAY_REPORT {

				relayEntry := RelayEntry{}
				relayEntry.name = from.String()
				relayEntry.id = uint64(GetRelayId(from.String()))
				relayEntry.address = from
				relayEntry.lastUpdate = time.Now().Unix()

				key := string(from.String())

				backend.mutex.Lock()
				_, ok := backend.relayDatabase[key]
				if !ok {
					backend.dirty = true
				}
				backend.relayDatabase[key] = relayEntry
				backend.mutex.Unlock()

			}

			return nil
		},
		":40000",
	)
}
