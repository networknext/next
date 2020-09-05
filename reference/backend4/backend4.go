/*
	Network Next Reference Backend (SDK4)
	Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package main

// #cgo pkg-config: libsodium
// #include <sodium.h>
import "C"

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"math"
	"math/bits"
	"math/rand"
	"net"
	"net/http"
	"runtime/debug"
	"sort"
	"strconv"
	"sync"
	"time"
	"unsafe"

	"github.com/gorilla/mux"
)

const NEXT_MAX_ROUTE_RELAYS = 5

const NEXT_MAX_SESSION_DATA_BYTES = 511

const NEXT_MAX_SESSION_UPDATE_RETRIES = 10

const NEXT_MAX_NEAR_RELAYS = 32
const NEXT_RELAY_BACKEND_PORT = 30000
const NEXT_SERVER_BACKEND_PORT = 40000

const NEXT_BACKEND_SERVER_UPDATE_PACKET = 220
const NEXT_BACKEND_SESSION_UPDATE_PACKET = 221
const NEXT_BACKEND_SESSION_RESPONSE_PACKET = 222
const NEXT_BACKEND_SERVER_INIT_REQUEST_PACKET = 223
const NEXT_BACKEND_SERVER_INIT_RESPONSE_PACKET = 224

const NEXT_MAX_PACKET_BYTES = 4096
const NEXT_MTU = 1300
const NEXT_ADDRESS_BYTES = 19
const NEXT_MAX_NODES = 7
const NEXT_ROUTE_TOKEN_BYTES = 76
const NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES = 116
const NEXT_CONTINUE_TOKEN_BYTES = 17
const NEXT_ENCRYPTED_CONTINUE_TOKEN_BYTES = 57

const NEXT_ROUTE_TYPE_DIRECT = 0
const NEXT_ROUTE_TYPE_NEW = 1
const NEXT_ROUTE_TYPE_CONTINUE = 2

const NEXT_FLAGS_BAD_ROUTE_TOKEN = uint32(1 << 0)
const NEXT_FLAGS_NO_ROUTE_TO_CONTINUE = uint32(1 << 1)
const NEXT_FLAGS_PREVIOUS_UPDATE_STILL_PENDING = uint32(1 << 2)
const NEXT_FLAGS_BAD_CONTINUE_TOKEN = uint32(1 << 3)
const NEXT_FLAGS_ROUTE_EXPIRED = uint32(1 << 4)
const NEXT_FLAGS_ROUTE_REQUEST_TIMED_OUT = uint32(1 << 5)
const NEXT_FLAGS_CONTINUE_REQUEST_TIMED_OUT = uint32(1 << 6)
const NEXT_FLAGS_CLIENT_TIMED_OUT = uint32(1 << 7)
const NEXT_FLAGS_UPGRADE_RESPONSE_TIMED_OUT = uint32(1 << 8)
const NEXT_FLAGS_COUNT = 9

const NEXT_RELAY_INIT_REQUEST_MAGIC = uint32(0x9083708f)
const NEXT_RELAY_INIT_REQUEST_VERSION = 0
const NEXT_RELAY_INIT_RESPONSE_VERSION = 0
const NEXT_RELAY_UPDATE_REQUEST_VERSION = 0
const NEXT_RELAY_UPDATE_RESPONSE_VERSION = 0
const NEXT_MAX_RELAY_ADDRESS_LENGTH = 256
const NEXT_RELAY_TOKEN_BYTES = 32
const NEXT_MAX_RELAYS = 1024

const NEXT_SERVER_INIT_RESPONSE_OK = 0
const NEXT_SERVER_INIT_RESPONSE_UNKNOWN_CUSTOMER = 1
const NEXT_SERVER_INIT_RESPONSE_UNKNOWN_DATACENTER = 2
const NEXT_SERVER_INIT_RESPONSE_SDK_VERSION_TOO_OLD = 3
const NEXT_SERVER_INIT_RESPONSE_SIGNATURE_CHECK_FAILED = 4

const NEXT_PACKET_HASH_BYTES = 8

const (
	ADDRESS_NONE = 0
	ADDRESS_IPV4 = 1
	ADDRESS_IPV6 = 2
)

const (
	NEXT_CONNECTION_TYPE_UNKNOWN  = 0
	NEXT_CONNECTION_TYPE_WIRED    = 1
	NEXT_CONNECTION_TYPE_WIFI     = 2
	NEXT_CONNECTION_TYPE_CELLULAR = 3
	NEXT_CONNECTION_TYPE_MAX = 3
)

const (
	NEXT_PLATFORM_UNKNOWN  = 0
	NEXT_PLATFORM_WINDOWS  = 1
	NEXT_PLATFORM_MAC      = 2
	NEXT_PLATFORM_UNIX     = 3
	NEXT_PLATFORM_SWITCH   = 4
	NEXT_PLATFORM_PS4      = 5
	NEXT_PLATFORM_IOS      = 6
	NEXT_PLATFORM_XBOX_ONE = 7
	NEXT_PLATFORM_MAX      = 7
)

var relayPublicKey = []byte{
	0xf5, 0x22, 0xad, 0xc1, 0xee, 0x04, 0x6a, 0xbe,
	0x7d, 0x89, 0x0c, 0x81, 0x3a, 0x08, 0x31, 0xba,
	0xdc, 0xdd, 0xb5, 0x52, 0xcb, 0x73, 0x56, 0x10,
	0xda, 0xa9, 0xc0, 0xae, 0x08, 0xa2, 0xcf, 0x5e,
}

var routerPrivateKey = [...]byte{0x96, 0xce, 0x57, 0x8b, 0x00, 0x19, 0x44, 0x27, 0xf2, 0xb9, 0x90, 0x1b, 0x43, 0x56, 0xfd, 0x4f, 0x56, 0xe1, 0xd9, 0x56, 0x58, 0xf2, 0xf4, 0x3b, 0x86, 0x9f, 0x12, 0x75, 0x24, 0xd2, 0x47, 0xb3}

var backendPrivateKey = []byte{21, 124, 5, 171, 56, 198, 148, 140, 20, 15, 8, 170, 212, 222, 84, 155, 149, 84, 122, 199, 107, 225, 243, 246, 133, 85, 118, 114, 114, 126, 200, 4, 76, 97, 202, 140, 71, 135, 62, 212, 160, 181, 151, 195, 202, 224, 207, 113, 8, 45, 37, 60, 145, 14, 212, 111, 25, 34, 175, 186, 37, 150, 163, 64}

var packetHashKey = []byte{0xe3, 0x18, 0x61, 0x72, 0xee, 0x70, 0x62, 0x37, 0x40, 0xf6, 0x0a, 0xea, 0xe0, 0xb5, 0x1a, 0x2c, 0x2a, 0x47, 0x98, 0x8f, 0x27, 0xec, 0x63, 0x2c, 0x25, 0x04, 0x74, 0x89, 0xaf, 0x5a, 0xeb, 0x24}

// ===================================================================================================================

type NextBackendServerInitRequestPacket struct {
	VersionMajor uint32
	VersionMinor uint32
	VersionPatch uint32
	CustomerId   uint64
	DatacenterId uint64
	RequestId    uint64
}

func (packet *NextBackendServerInitRequestPacket) Serialize(stream Stream) error {
	stream.SerializeBits(&packet.VersionMajor, 8)
	stream.SerializeBits(&packet.VersionMinor, 8)
	stream.SerializeBits(&packet.VersionPatch, 8)
	stream.SerializeUint64(&packet.CustomerId)
	stream.SerializeUint64(&packet.DatacenterId)
	stream.SerializeUint64(&packet.RequestId)
	return stream.Error()
}

// -------------------------------------------------------------------------------------

type NextBackendServerInitResponsePacket struct {
	RequestId uint64
	Response  uint32
}

func (packet *NextBackendServerInitResponsePacket) Serialize(stream Stream) error {
	stream.SerializeUint64(&packet.RequestId)
	stream.SerializeBits(&packet.Response, 8)
	return stream.Error()
}

// -------------------------------------------------------------------------------------

type NextBackendServerUpdatePacket struct {
	VersionMajor         uint32
	VersionMinor         uint32
	VersionPatch         uint32
	CustomerId           uint64
	DatacenterId         uint64
	NumSessions   		 uint32
	ServerAddress        net.UDPAddr
}

func (packet *NextBackendServerUpdatePacket) Serialize(stream Stream) error {
	stream.SerializeBits(&packet.VersionMajor, 8)
	stream.SerializeBits(&packet.VersionMinor, 8)
	stream.SerializeBits(&packet.VersionPatch, 8)
	stream.SerializeUint64(&packet.CustomerId)
	stream.SerializeUint64(&packet.DatacenterId)
	stream.SerializeUint32(&packet.NumSessions)
	stream.SerializeAddress(&packet.ServerAddress)
	return stream.Error()
}

// -----------------------------------------------------------------------------

type NextBackendSessionUpdatePacket struct {
	VersionMajor         	  uint32
	VersionMinor              uint32
	VersionPatch              uint32
	CustomerId                uint64
	SessionId                 uint64
	SliceNumber               uint32
	RetryNumber               int32
	SessionDataBytes          int32
	SessionData               [NEXT_MAX_SESSION_DATA_BYTES]byte
	ClientAddress             net.UDPAddr
	ServerAddress             net.UDPAddr
	ClientRoutePublicKey      []byte
	ServerRoutePublicKey      []byte
	UserHash                  uint64
	PlatformId                int32
	ConnectionType            int32
	Next             		  bool
	Committed                 bool
	Reported                  bool
	Tag                       uint64
	Flags                     uint32
	UserFlags                 uint64
	DirectRTT                 float32
	DirectJitter              float32
	DirectPacketLoss          float32
	NextRTT                   float32
	NextJitter                float32
	NextPacketLoss            float32
	NumNearRelays             int32
	NearRelayIds              []uint64
	NearRelayRTT              []float32
	NearRelayJitter           []float32
	NearRelayPacketLoss       []float32
	KbpsUp                    uint32
	KbpsDown                  uint32
	PacketsSentClientToServer uint64
	PacketsSentServerToClient uint64
	PacketsLostClientToServer uint64
	PacketsLostServerToClient uint64
}

func (packet *NextBackendSessionUpdatePacket) Serialize(stream Stream) error {
	
	stream.SerializeBits(&packet.VersionMajor, 8)
	stream.SerializeBits(&packet.VersionMinor, 8)
	stream.SerializeBits(&packet.VersionPatch, 8)
	
	stream.SerializeUint64(&packet.CustomerId)
	
	stream.SerializeUint64(&packet.SessionId)
	
	stream.SerializeBits(&packet.SliceNumber, 32)
	
	stream.SerializeInteger(&packet.RetryNumber, 0, NEXT_MAX_SESSION_UPDATE_RETRIES)
	
	stream.SerializeInteger(&packet.SessionDataBytes, 0, NEXT_MAX_SESSION_DATA_BYTES)
	if packet.SessionDataBytes > 0 {
		sessionData := packet.SessionData[:packet.SessionDataBytes]
		stream.SerializeBytes(sessionData)
	}
	
	stream.SerializeAddress(&packet.ClientAddress)
	stream.SerializeAddress(&packet.ServerAddress)

	if stream.IsReading() {
		packet.ClientRoutePublicKey = make([]byte, Crypto_box_PUBLICKEYBYTES)
		packet.ServerRoutePublicKey = make([]byte, Crypto_box_PUBLICKEYBYTES)
	}
	stream.SerializeBytes(packet.ClientRoutePublicKey)
	stream.SerializeBytes(packet.ServerRoutePublicKey)

	stream.SerializeUint64(&packet.UserHash)

	stream.SerializeInteger(&packet.PlatformId, 0, NEXT_PLATFORM_MAX)

	stream.SerializeInteger(&packet.ConnectionType, NEXT_CONNECTION_TYPE_UNKNOWN, NEXT_CONNECTION_TYPE_MAX)

	stream.SerializeBool(&packet.Next)
	stream.SerializeBool(&packet.Committed)
	stream.SerializeBool(&packet.Reported)

	hasTag := stream.IsWriting() && packet.Tag != 0
	hasFlags := stream.IsWriting() && packet.Flags != 0
	hasUserFlags := stream.IsWriting() && packet.UserFlags != 0
	hasLostPackets := stream.IsWriting() && ( packet.PacketsLostClientToServer + packet.PacketsLostServerToClient ) > 0

	stream.SerializeBool( &hasTag )
	stream.SerializeBool( &hasFlags )
	stream.SerializeBool( &hasUserFlags )
	stream.SerializeBool( &hasLostPackets )

	if hasTag {
		stream.SerializeUint64(&packet.Tag)
	}

	if hasFlags {
		stream.SerializeBits(&packet.Flags, NEXT_FLAGS_COUNT)
	}

	if hasUserFlags {
		stream.SerializeUint64(&packet.UserFlags)
	}

	stream.SerializeFloat32(&packet.DirectRTT)
	stream.SerializeFloat32(&packet.DirectJitter)
	stream.SerializeFloat32(&packet.DirectPacketLoss)

	if packet.Next {
		stream.SerializeFloat32(&packet.NextRTT)
		stream.SerializeFloat32(&packet.NextJitter)
		stream.SerializeFloat32(&packet.NextPacketLoss)
	}

	stream.SerializeInteger(&packet.NumNearRelays, 0, NEXT_MAX_NEAR_RELAYS)
	if stream.IsReading() {
		packet.NearRelayIds = make([]uint64, packet.NumNearRelays)
		packet.NearRelayRTT = make([]float32, packet.NumNearRelays)
		packet.NearRelayJitter = make([]float32, packet.NumNearRelays)
		packet.NearRelayPacketLoss = make([]float32, packet.NumNearRelays)
	}
	var i int32
	for i = 0; i < packet.NumNearRelays; i++ {
		stream.SerializeUint64(&packet.NearRelayIds[i])
		stream.SerializeFloat32(&packet.NearRelayRTT[i])
		stream.SerializeFloat32(&packet.NearRelayJitter[i])
		stream.SerializeFloat32(&packet.NearRelayPacketLoss[i])
	}

	if packet.Next {
		stream.SerializeUint32(&packet.KbpsUp)
		stream.SerializeUint32(&packet.KbpsDown)
	}

	stream.SerializeUint64(&packet.PacketsSentClientToServer)
	stream.SerializeUint64(&packet.PacketsSentServerToClient)

	if hasLostPackets {
		stream.SerializeUint64(&packet.PacketsLostClientToServer)
		stream.SerializeUint64(&packet.PacketsLostServerToClient)
	}

	return stream.Error()
}

// --------------------------------------------------------------------------------

type NextBackendSessionResponsePacket struct {
	SessionId            uint64
	SliceNumber          uint32
	SessionDataBytes     int32
	SessionData          [NEXT_MAX_SESSION_DATA_BYTES]byte
	RouteType            int32
	NumNearRelays        int32
	NearRelayIds         []uint64
	NearRelayAddresses   []net.UDPAddr
	NumTokens            int32
	Tokens               []byte
	Multipath            bool
	Committed            bool
}

func (packet *NextBackendSessionResponsePacket) Serialize(stream Stream, versionMajor uint32, versionMinor uint32, versionPatch uint32) error {

	stream.SerializeUint64(&packet.SessionId)

	stream.SerializeBits(&packet.SliceNumber, 32)

	stream.SerializeInteger(&packet.SessionDataBytes, 0, NEXT_MAX_SESSION_DATA_BYTES)
	if packet.SessionDataBytes > 0 {
		sessionData := packet.SessionData[:packet.SessionDataBytes]
		stream.SerializeBytes(sessionData)
	}

	stream.SerializeInteger(&packet.RouteType, 0, NEXT_ROUTE_TYPE_CONTINUE)

	stream.SerializeInteger(&packet.NumNearRelays, 0, NEXT_MAX_NEAR_RELAYS)

	if stream.IsReading() {
		packet.NearRelayIds = make([]uint64, packet.NumNearRelays)
		packet.NearRelayAddresses = make([]net.UDPAddr, packet.NumNearRelays)
	}
	var i int32
	for i = 0; i < packet.NumNearRelays; i++ {
		stream.SerializeUint64(&packet.NearRelayIds[i])
		stream.SerializeAddress(&packet.NearRelayAddresses[i])
	}

	if packet.RouteType != NEXT_ROUTE_TYPE_DIRECT {
		stream.SerializeBool(&packet.Multipath)
		stream.SerializeBool(&packet.Committed)
		stream.SerializeInteger(&packet.NumTokens, 0, NEXT_MAX_NODES)
	}

	if packet.RouteType == NEXT_ROUTE_TYPE_NEW {
		if stream.IsReading() {
			packet.Tokens = make([]byte, packet.NumTokens*NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES)
		}
		stream.SerializeBytes(packet.Tokens)
	}

	if packet.RouteType == NEXT_ROUTE_TYPE_CONTINUE {
		if stream.IsReading() {
			packet.Tokens = make([]byte, packet.NumTokens*NEXT_ENCRYPTED_CONTINUE_TOKEN_BYTES)
		}
		stream.SerializeBytes(packet.Tokens)
	}

	return stream.Error()
}

// ------------------------------------------------------------------------------------------

const SessionDataVersion = 0

type SessionData struct {
	Version uint32
	SessionId uint64
	SessionVersion uint32
	SliceNumber uint32
	ExpireTimestamp uint64
	Route []uint64
}

func (packet *SessionData) Serialize(stream Stream) error {

	stream.SerializeBits(&packet.Version, 8)
	if stream.IsReading() && packet.Version != SessionDataVersion {
		return fmt.Errorf("bad session data version %d, expected %d", packet.Version, SessionDataVersion)
	}
	
	stream.SerializeUint64(&packet.SessionId)
	
	stream.SerializeBits(&packet.SliceNumber, 32)
	
	stream.SerializeBits(&packet.SessionVersion, 8)

	stream.SerializeUint64(&packet.ExpireTimestamp)

	numRelays := int32(0)
	hasRoute := false
	if stream.IsWriting() {
		numRelays = int32(len(packet.Route))
		hasRoute = numRelays > 0
	}
	
	stream.SerializeBool(&hasRoute) 
	if hasRoute {
		stream.SerializeInteger(&numRelays, 0, NEXT_MAX_ROUTE_RELAYS)
		if stream.IsReading() {
			packet.Route = make([]uint64, numRelays)
		}
		for i := 0; i < int(numRelays); i++ {
			stream.SerializeUint64(&packet.Route[i])
		}
	}
	
	return stream.Error()
}

// ===================================================================================================================

type Backend struct {
	mutex           sync.RWMutex
	dirty           bool
	relayDatabase   map[string]RelayEntry
	serverDatabase  map[string]ServerEntry
	sessionDatabase map[uint64]SessionEntry
	nearData        []byte
}

var backend Backend

type RelayEntry struct {
	id         uint64
	name       string
	address    *net.UDPAddr
	lastUpdate int64
	token      []byte
}

type ServerEntry struct {
	address    *net.UDPAddr
	publicKey  []byte
	lastUpdate int64
}

type SessionEntry struct {
	id              uint64
	expireTimestamp uint64
}

func TimeoutThread() {
	for {
		time.Sleep(time.Second * 1)
		backend.mutex.Lock()
		currentTimestamp := time.Now().Unix()
		for k, v := range backend.relayDatabase {
			if currentTimestamp-v.lastUpdate > 15 {
				backend.dirty = true
				delete(backend.relayDatabase, k)
				continue
			}
		}
		for k, v := range backend.serverDatabase {
			if currentTimestamp-v.lastUpdate > 15 {
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
		if backend.dirty {
			fmt.Printf("-----------------------------\n")
			for _, v := range backend.relayDatabase {
				fmt.Printf("relay: %s\n", v.address)
			}
			for _, v := range backend.serverDatabase {
				fmt.Printf("server: %s\n", v.address)
			}
			for k := range backend.sessionDatabase {
				fmt.Printf("session: %x\n", k)
			}
			if len(backend.relayDatabase) == 0 && len(backend.serverDatabase) == 0 && len(backend.sessionDatabase) == 0 {
				fmt.Printf("(nil)\n")
			}
			backend.dirty = false
		}
		backend.mutex.Unlock()
	}
}

func GetRelayId(name string) uint64 {
	hash := fnv.New64a()
	hash.Write([]byte(name))
	return hash.Sum64()
}

func GetNearRelays() ([]uint64, []net.UDPAddr) {
	nearRelays := make([]RelayEntry, 0)
	backend.mutex.RLock()
	for _, v := range backend.relayDatabase {
		nearRelays = append(nearRelays, v)
	}
	backend.mutex.RUnlock()
	sort.SliceStable(nearRelays[:], func(i, j int) bool { return nearRelays[i].id < nearRelays[j].id })
	if len(nearRelays) > NEXT_MAX_NEAR_RELAYS {
		nearRelays = nearRelays[:NEXT_MAX_NEAR_RELAYS]
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

func IsNetworkNextPacket(packetData []byte) bool {
	packetBytes := len(packetData)
	if packetBytes <= NEXT_PACKET_HASH_BYTES {
		fmt.Printf("packet too small\n")
		return false
	}
	if packetBytes > NEXT_MAX_PACKET_BYTES {
		fmt.Printf("packet too big\n")
		return false
	}
	messageLength := packetBytes - NEXT_PACKET_HASH_BYTES
	if messageLength > 32 {
		messageLength = 32
	}
	hash := make([]byte, NEXT_PACKET_HASH_BYTES)
	C.crypto_generichash(
		(*C.uchar)(&hash[0]),
		C.ulong(NEXT_PACKET_HASH_BYTES),
		(*C.uchar)(&packetData[NEXT_PACKET_HASH_BYTES]),
		C.ulonglong(messageLength),
		(*C.uchar)(&packetHashKey[0]),
		C.ulong(C.crypto_generichash_KEYBYTES),
	)
	for i := 0; i < NEXT_PACKET_HASH_BYTES; i++ {
		if hash[i] != packetData[i] {
			fmt.Printf("signature check failed\n")
			return false
		}
	}
	return true
}

func SignNetworkNextPacket(packetData []byte, privateKey []byte) []byte {
	signedPacketData := make([]byte, len(packetData)+C.crypto_sign_BYTES)
	for i := 0; i < len(packetData); i++ {
		signedPacketData[i] = packetData[i]
	}
	messageLength := len(packetData)
	var state C.crypto_sign_state
	C.crypto_sign_init(&state)
	C.crypto_sign_update(&state, (*C.uchar)(&signedPacketData[0]), C.ulonglong(messageLength))
	C.crypto_sign_final_create(&state, (*C.uchar)(&signedPacketData[len(packetData)]), nil, (*C.uchar)(&privateKey[0]))
	return signedPacketData
}

func HashNetworkNextPacket(packetData []byte) []byte {
	hashedPacketData := make([]byte, len(packetData)+NEXT_PACKET_HASH_BYTES)
	messageLength := len(packetData)
	if messageLength > 32 {
		messageLength = 32
	}
	C.crypto_generichash(
		(*C.uchar)(&hashedPacketData[0]),
		C.ulong(NEXT_PACKET_HASH_BYTES),
		(*C.uchar)(&packetData[0]),
		C.ulonglong(messageLength),
		(*C.uchar)(&packetHashKey[0]),
		C.ulong(C.crypto_generichash_KEYBYTES),
	)
	for i := 0; i < len(packetData); i++ {
		hashedPacketData[NEXT_PACKET_HASH_BYTES+i] = packetData[i]
	}
	return hashedPacketData
}

// -----------------------------------------------------------

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

func GetRelayPublicKey(relay_address string) []byte {
	return []byte{0x06, 0xb0, 0x4d, 0x9e, 0xa6, 0xf5, 0x7c, 0x0b, 0x3c, 0x6a, 0x2d, 0x9d, 0xbf, 0x34, 0x32, 0xb6, 0x66, 0x00, 0xa0, 0x3b, 0x2b, 0x5b, 0x5d, 0x00, 0x91, 0x4a, 0x32, 0xee, 0xf2, 0x36, 0xc2, 0x9c}
}

func CryptoCheck(data []byte, nonce []byte, publicKey []byte, privateKey []byte) bool {
	return C.crypto_box_open((*C.uchar)(&data[0]), (*C.uchar)(&data[0]), C.ulonglong(len(data)), (*C.uchar)(&nonce[0]), (*C.uchar)(&publicKey[0]), (*C.uchar)(&privateKey[0])) != 0
}

func RelayInitHandler(writer http.ResponseWriter, request *http.Request) {
	body, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return
	}
	defer request.Body.Close()

	index := 0

	var magic uint32
	if !ReadUint32(body, &index, &magic) || magic != NEXT_RELAY_INIT_REQUEST_MAGIC {
		return
	}

	var version uint32
	if !ReadUint32(body, &index, &version) || version != NEXT_RELAY_INIT_REQUEST_VERSION {
		return
	}

	var nonce []byte
	if !ReadBytes(body, &index, &nonce, C.crypto_box_NONCEBYTES) {
		return
	}

	var relay_address string
	if !ReadString(body, &index, &relay_address, NEXT_MAX_RELAY_ADDRESS_LENGTH) {
		return
	}

	var encrypted_token []byte
	if !ReadBytes(body, &index, &encrypted_token, NEXT_RELAY_TOKEN_BYTES+C.crypto_box_MACBYTES) {
		return
	}

	if !CryptoCheck(encrypted_token, nonce, relayPublicKey[:], routerPrivateKey[:]) {
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

	relayEntry := RelayEntry{}
	relayEntry.name = relay_address
	relayEntry.id = GetRelayId(relay_address)
	relayEntry.address = ParseAddress(relay_address)
	relayEntry.lastUpdate = time.Now().Unix()
	relayEntry.token = RandomBytes(NEXT_RELAY_TOKEN_BYTES)

	backend.mutex.Lock()
	backend.relayDatabase[key] = relayEntry
	backend.dirty = true
	backend.mutex.Unlock()

	writer.Header().Set("Content-Type", "application/octet-stream")

	responseData := make([]byte, 64)
	index = 0
	WriteUint32(responseData, &index, NEXT_RELAY_INIT_RESPONSE_VERSION)
	WriteUint64(responseData, &index, uint64(time.Now().Unix()))
	WriteBytes(responseData, &index, relayEntry.token, NEXT_RELAY_TOKEN_BYTES)
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
	defer request.Body.Close()

	index := 0

	var version uint32
	if !ReadUint32(body, &index, &version) || version != NEXT_RELAY_UPDATE_REQUEST_VERSION {
		return
	}

	var relay_address string
	if !ReadString(body, &index, &relay_address, NEXT_MAX_RELAY_ADDRESS_LENGTH) {
		return
	}

	var token []byte
	if !ReadBytes(body, &index, &token, NEXT_RELAY_TOKEN_BYTES) {
		return
	}

	key := relay_address

	backend.mutex.RLock()
	relayEntry, ok := backend.relayDatabase[key]
	found := false
	if ok && CompareTokens(token, relayEntry.token) {
		found = true
	}
	backend.mutex.RUnlock()

	if !found {
		writer.WriteHeader(http.StatusNotFound)
		return
	}

	var num_relays uint32
	if !ReadUint32(body, &index, &num_relays) {
		return
	}

	if num_relays > NEXT_MAX_RELAYS {
		return
	}

	for i := 0; i < int(num_relays); i++ {
		var id uint64
		var rtt, jitter, packet_loss float32
		if !ReadUint64(body, &index, &id) {
			return
		}
		if !ReadFloat32(body, &index, &rtt) {
			return
		}
		if !ReadFloat32(body, &index, &jitter) {
			return
		}
		if !ReadFloat32(body, &index, &packet_loss) {
			return
		}
	}

	relayEntry = RelayEntry{}
	relayEntry.name = relay_address
	relayEntry.id = GetRelayId(relay_address)
	relayEntry.address = ParseAddress(relay_address)
	relayEntry.lastUpdate = time.Now().Unix()
	relayEntry.token = token

	type RelayPingData struct {
		id      uint64
		address string
	}

	relaysToPing := make([]RelayPingData, 0)

	backend.mutex.Lock()
	backend.relayDatabase[key] = relayEntry
	for k, v := range backend.relayDatabase {
		if k != relay_address {
			if k != relay_address {
				relaysToPing = append(relaysToPing, RelayPingData{id: v.id, address: k})
			}
		}
	}
	backend.mutex.Unlock()

	responseData := make([]byte, 10*1024)

	index = 0

	WriteUint32(responseData, &index, NEXT_RELAY_UPDATE_RESPONSE_VERSION)

	WriteUint32(responseData, &index, uint32(len(relaysToPing)))

	for i := range relaysToPing {
		WriteUint64(responseData, &index, relaysToPing[i].id)
		WriteString(responseData, &index, relaysToPing[i].address, NEXT_MAX_RELAY_ADDRESS_LENGTH)
	}

	responseLength := index

	responseData = responseData[:responseLength]

	writer.Header().Set("Content-Type", "application/octet-stream")

	writer.Write(responseData)
}

func WebServer() {
	router := mux.NewRouter()
	router.HandleFunc("/relay_init", RelayInitHandler).Methods("POST")
	router.HandleFunc("/relay_update", RelayUpdateHandler).Methods("POST")
	http.ListenAndServe(fmt.Sprintf(":%d", NEXT_RELAY_BACKEND_PORT), router)
}

// ========================================================================================================

func uintsToBytes(vs []uint32) []byte {
	buf := make([]byte, len(vs)*4)
	for i, v := range vs {
		binary.LittleEndian.PutUint32(buf[i*4:], v)
	}
	return buf
}

func bytesToUints(vs []byte) []uint32 {
	out := make([]uint32, (len(vs)+3)/4)
	for i := range out {
		tmp := [4]byte{}
		copy(tmp[:], vs[i*4:])
		out[i] = binary.LittleEndian.Uint32(tmp[:])
	}
	return out
}

func log2(x uint32) int {
	a := x | (x >> 1)
	b := a | (a >> 2)
	c := b | (b >> 4)
	d := c | (c >> 8)
	e := d | (d >> 16)
	f := e >> 1
	return bits.OnesCount32(f)
}

func bitsRequired(min uint32, max uint32) int {
	if min == max {
		return 0
	} else {
		return log2(max-min) + 1
	}
}

func bitsRequiredSigned(min int32, max int32) int {
	if min == max {
		return 0
	} else {
		return log2(uint32(max-min)) + 1
	}
}

func sequenceGreaterThan(s1 uint16, s2 uint16) bool {
	return ((s1 > s2) && (s1-s2 <= 32768)) ||
		((s1 < s2) && (s2-s1 > 32768))
}

func sequenceLessThan(s1 uint16, s2 uint16) bool {
	return sequenceGreaterThan(s2, s1)
}

func signedToUnsigned(n int32) uint32 {
	return uint32((n << 1) ^ (n >> 31))
}

func unsignedToSigned(n uint32) int32 {
	return int32(n>>1) ^ (-int32(n & 1))
}

// -----------------------------------------------------------------------------------

type BitWriter struct {
	data        []uint32
	scratch     uint64
	numBits     int
	bitsWritten int
	wordIndex   int
	scratchBits int
	numWords    int
}

func CreateBitWriter(bytes int) (*BitWriter, error) {
	if bytes%4 != 0 {
		return nil, fmt.Errorf("bitwriter bytes must be a multiple of 4")
	}
	numWords := bytes / 4
	return &BitWriter{
		data:     make([]uint32, numWords),
		numBits:  numWords * 32,
		numWords: numWords,
	}, nil
}

func hostToNetwork(x uint32) uint32 {
	return x
}

func networkToHost(x uint32) uint32 {
	return x
}

func (writer *BitWriter) WriteBits(value uint32, bits int) error {
	if bits <= 0 || bits > 32 {
		return fmt.Errorf("expected between 1 and 32 bits, got %d bits", bits)
	}
	if writer.bitsWritten+bits > writer.numBits {
		return fmt.Errorf("buffer overflow")
	}
	if uint64(value) > ((1 << uint64(bits)) - 1) {
		return fmt.Errorf("%d is not representable in %d bits", value, bits)
	}

	writer.scratch |= uint64(value) << uint(writer.scratchBits)

	writer.scratchBits += bits

	if writer.scratchBits >= 32 {
		writer.data[writer.wordIndex] = hostToNetwork(uint32(writer.scratch & 0xFFFFFFFF))
		writer.scratch >>= 32
		writer.scratchBits -= 32
		writer.wordIndex++
	}

	writer.bitsWritten += bits

	return nil
}

func (writer *BitWriter) WriteAlign() error {
	remainderBits := writer.bitsWritten % 8

	if remainderBits != 0 {
		err := writer.WriteBits(uint32(0), 8-remainderBits)
		if err != nil {
			return err
		}
		if writer.bitsWritten%8 != 0 {
			panic("WriteAlign() failed to align BitWriter")
		}
	}
	return nil
}

func (writer *BitWriter) WriteBytes(data []byte) error {
	if writer.GetAlignBits() != 0 {
		panic("writer must be aligned before calling WriteBytes()")
	}

	{
		bitIndex := writer.bitsWritten % 32
		if bitIndex != 0 && bitIndex != 8 && bitIndex != 16 && bitIndex != 24 {
			panic("bit index should be aligned before calling WriteBytes()")
		}
	}

	if writer.bitsWritten+len(data)*8 > writer.numBits {
		return fmt.Errorf("buffer overflow")
	}

	headBytes := (4 - (writer.bitsWritten%32)/8) % 4
	if headBytes > len(data) {
		headBytes = len(data)
	}
	for i := 0; i < headBytes; i++ {
		writer.WriteBits(uint32(data[i]), 8)
	}
	if headBytes == len(data) {
		return nil
	}

	if err := writer.FlushBits(); err != nil {
		return err
	}

	if writer.GetAlignBits() != 0 {
		panic("writer should be aligned")
	}

	numWords := (len(data) - headBytes) / 4
	if numWords > 0 {
		if (writer.bitsWritten % 32) != 0 {
			panic("bits written should be aligned")
		}
		copy(writer.data[writer.wordIndex:], bytesToUints(data[headBytes:headBytes+numWords*4]))
		writer.bitsWritten += numWords * 32
		writer.wordIndex += numWords
		writer.scratch = 0
	}

	if writer.GetAlignBits() != 0 {
		panic("writer should be aligned")
	}

	tailStart := headBytes + numWords*4
	tailBytes := len(data) - tailStart
	if tailBytes < 0 || tailBytes >= 4 {
		panic(fmt.Sprintf("tail bytes out of range: %d, should be between 0 and 4", tailBytes))
	}

	for i := 0; i < tailBytes; i++ {
		err := writer.WriteBits(uint32(data[tailStart+i]), 8)
		if err != nil {
			return err
		}
	}

	if writer.GetAlignBits() != 0 {
		panic("writer should be aligned")
	}

	if headBytes+numWords*4+tailBytes != len(data) {
		panic("everything should add up")
	}
	return nil
}

func (writer *BitWriter) FlushBits() error {
	if writer.scratchBits != 0 {
		if writer.scratchBits > 32 {
			panic("scratch bits should be 32 or less")
		}
		if writer.wordIndex >= writer.numWords {
			return fmt.Errorf("buffer overflow")
		}
		writer.data[writer.wordIndex] = hostToNetwork(uint32(writer.scratch & 0xFFFFFFFF))
		writer.scratch >>= 32
		writer.scratchBits = 0
		writer.wordIndex++
	}
	return nil
}

func (writer *BitWriter) GetAlignBits() int {
	return (8 - (writer.bitsWritten % 8)) % 8
}

func (writer *BitWriter) GetBitsWritten() int {
	return writer.bitsWritten
}

func (writer *BitWriter) GetBitsAvailable() int {
	return writer.numBits - writer.bitsWritten
}

func (writer *BitWriter) GetData() []byte {
	return uintsToBytes(writer.data)
}

func (writer *BitWriter) GetBytesWritten() int {
	return (writer.bitsWritten + 7) / 8
}

// -----------------------------------------------------------------

type BitReader struct {
	data        []uint32
	numBits     int
	numBytes    int
	numWords    int
	bitsRead    int
	scratch     uint64
	scratchBits int
	wordIndex   int
}

func CreateBitReader(data []byte) *BitReader {
	return &BitReader{
		numBits:  len(data) * 8,
		numBytes: len(data),
		numWords: (len(data) + 3) / 4,
		data:     bytesToUints(data),
	}
}

func (reader *BitReader) WouldReadPastEnd(bits int) bool {
	return reader.bitsRead+bits > reader.numBits
}

func (reader *BitReader) ReadBits(bits int) (uint32, error) {
	if bits < 0 || bits > 32 {
		return 0, fmt.Errorf("expected between 0 and 32 bits")
	}
	if reader.bitsRead+bits > reader.numBits {
		return 0, fmt.Errorf("call would read past end of buffer")
	}

	reader.bitsRead += bits

	if reader.scratchBits < 0 || reader.scratchBits > 64 {
		panic("scratch bits should be between 0 and 64")
	}

	if reader.scratchBits < bits {
		if reader.wordIndex >= reader.numWords {
			return 0, fmt.Errorf("would read past end of buffer")
		}
		reader.scratch |= uint64(networkToHost(reader.data[reader.wordIndex])) << uint(reader.scratchBits)
		reader.scratchBits += 32
		reader.wordIndex++
	}

	if reader.scratchBits < bits {
		panic(fmt.Sprintf("should have written at least %d scratch bits", bits))
	}

	output := reader.scratch & ((uint64(1) << uint(bits)) - 1)

	reader.scratch >>= uint(bits)
	reader.scratchBits -= bits

	return uint32(output), nil
}

func (reader *BitReader) ReadAlign() error {
	remainderBits := reader.bitsRead % 8
	if remainderBits != 0 {
		value, err := reader.ReadBits(8 - remainderBits)
		if err != nil {
			return fmt.Errorf("ReadAlign() failed: %v", err)
		}
		if reader.bitsRead%8 != 0 {
			panic("reader should be aligned at this point")
		}
		if value != 0 {
			return fmt.Errorf("tried to read align; value should be 0")
		}
	}
	return nil
}

func (reader *BitReader) ReadBytes(buffer []byte) error {
	if reader.GetAlignBits() != 0 {
		return fmt.Errorf("reader should be aligned before calling ReadBytes()")
	}
	if reader.bitsRead+len(buffer)*8 > reader.numBits {
		return fmt.Errorf("would read past end of buffer")
	}
	{
		bitIndex := reader.bitsRead % 32
		if bitIndex != 0 && bitIndex != 8 && bitIndex != 16 && bitIndex != 24 {
			return fmt.Errorf("reader should be aligned before calling ReadBytes()")
		}
	}

	headBytes := (4 - (reader.bitsRead%32)/8) % 4
	if headBytes > len(buffer) {
		headBytes = len(buffer)
	}
	for i := 0; i < headBytes; i++ {
		value, err := reader.ReadBits(8)
		if err != nil {
			return err
		}
		buffer[i] = byte(value)
	}
	if headBytes == len(buffer) {
		return nil
	}

	if reader.GetAlignBits() != 0 {
		panic("reader should be aligned at this point")
	}

	numWords := (len(buffer) - headBytes) / 4
	if numWords > 0 {
		if (reader.bitsRead % 32) != 0 {
			panic("reader should be word aligned at this point")
		}
		copy(buffer[headBytes:], uintsToBytes(reader.data[reader.wordIndex:reader.wordIndex+numWords]))
		reader.bitsRead += numWords * 32
		reader.wordIndex += numWords
		reader.scratchBits = 0
	}

	if reader.GetAlignBits() != 0 {
		panic("reader should be aligned at this point")
	}

	tailStart := headBytes + numWords*4
	tailBytes := len(buffer) - tailStart
	if tailBytes < 0 || tailBytes >= 4 {
		panic(fmt.Sprintf("tail bytes out of range: %d, should be between 0 and 4", tailBytes))
	}
	for i := 0; i < tailBytes; i++ {
		value, err := reader.ReadBits(8)
		if err != nil {
			return err
		}
		buffer[tailStart+i] = byte(value)
	}

	if reader.GetAlignBits() != 0 {
		panic("reader should be aligned at this point")
	}

	if headBytes+numWords*4+tailBytes != len(buffer) {
		panic("everything should add up")
	}
	return nil
}

func (reader *BitReader) GetAlignBits() int {
	return (8 - reader.bitsRead%8) % 8
}

func (reader *BitReader) GetBitsRead() int {
	return reader.bitsRead
}

func (reader *BitReader) GetBitsRemaining() int {
	return reader.numBits - reader.bitsRead
}

// ----------------------------------------------------------------------------

type Stream interface {
	IsWriting() bool
	IsReading() bool
	SerializeInteger(value *int32, min int32, max int32)
	SerializeBits(value *uint32, bits int)
	SerializeUint32(value *uint32)
	SerializeBool(value *bool)
	SerializeFloat32(value *float32)
	SerializeUint64(value *uint64)
	SerializeFloat64(value *float64)
	SerializeBytes(data []byte)
	SerializeString(value *string, maxSize int)
	SerializeAlign()
	SerializeAddress(addr *net.UDPAddr)
	GetAlignBits() int
	GetBytesProcessed() int
	GetBitsProcessed() int
	Error() error
	Flush()
}

// ---------------------------------------------------------------------------

type WriteStream struct {
	writer *BitWriter
	err    error
}

func CreateWriteStream(bytes int) (*WriteStream, error) {
	writer, err := CreateBitWriter(bytes)
	if err != nil {
		return nil, err
	}
	return &WriteStream{
		writer: writer,
	}, nil
}

func (stream *WriteStream) IsWriting() bool {
	return true
}

func (stream *WriteStream) IsReading() bool {
	return false
}

func (stream *WriteStream) error(err error) {
	if err != nil && stream.err == nil {
		stream.err = fmt.Errorf("%v\n%s", err, string(debug.Stack()))
	}
}

func (stream *WriteStream) Error() error {
	return stream.err
}

func (stream *WriteStream) SerializeInteger(value *int32, min int32, max int32) {
	if stream.err != nil {
		return
	}
	if min >= max {
		stream.error(fmt.Errorf("min (%d) should be less than max (%d)", min, max))
		return
	}
	if value == nil {
		stream.error(fmt.Errorf("value is nil"))
		return
	}
	if *value < min {
		stream.error(fmt.Errorf("value (%d) should be at least min (%d)", *value, min))
		return
	}
	if *value > max {
		stream.error(fmt.Errorf("value (%d) should be no more than max (%d)", *value, max))
		return
	}
	bits := bitsRequired(uint32(min), uint32(max))
	unsignedValue := uint32(*value - min)
	stream.error(stream.writer.WriteBits(unsignedValue, bits))
}

func (stream *WriteStream) SerializeBits(value *uint32, bits int) {
	if stream.err != nil {
		return
	}
	if value == nil {
		stream.error(fmt.Errorf("value is nil"))
		return
	}
	if bits < 0 || bits > 32 {
		stream.error(fmt.Errorf("bits (%d) should be between 0 and 32", bits))
		return
	}
	stream.error(stream.writer.WriteBits(*value, bits))
}

func (stream *WriteStream) SerializeUint32(value *uint32) {
	stream.SerializeBits(value, 32)
}

func (stream *WriteStream) SerializeBool(value *bool) {
	if stream.err != nil {
		return
	}
	if value == nil {
		stream.error(fmt.Errorf("value is nil"))
		return
	}
	uint32Value := uint32(0)
	if *value {
		uint32Value = 1
	}
	stream.error(stream.writer.WriteBits(uint32Value, 1))
}

func (stream *WriteStream) SerializeFloat32(value *float32) {
	if stream.err != nil {
		return
	}
	if value == nil {
		stream.error(fmt.Errorf("value is nil"))
		return
	}
	stream.error(stream.writer.WriteBits(math.Float32bits(*value), 32))
}

func (stream *WriteStream) SerializeUint64(value *uint64) {
	if stream.err != nil {
		return
	}
	if value == nil {
		stream.error(fmt.Errorf("value is nil"))
		return
	}
	lo := uint32(*value & 0xFFFFFFFF)
	hi := uint32(*value >> 32)
	stream.SerializeBits(&lo, 32)
	stream.SerializeBits(&hi, 32)
}

func (stream *WriteStream) SerializeFloat64(value *float64) {
	if stream.err != nil {
		return
	}
	if value == nil {
		stream.error(fmt.Errorf("value is nil"))
		return
	}
	uint64Value := math.Float64bits(*value)
	stream.SerializeUint64(&uint64Value)
}

func (stream *WriteStream) SerializeBytes(data []byte) {
	if stream.err != nil {
		return
	}
	if len(data) == 0 {
		stream.error(fmt.Errorf("byte buffer should have more than 0 bytes"))
		return
	}
	stream.SerializeAlign()
	stream.error(stream.writer.WriteBytes(data))
}

func (stream *WriteStream) SerializeString(value *string, maxSize int) {
	if stream.err != nil {
		return
	}
	if value == nil {
		stream.error(fmt.Errorf("value is nil"))
		return
	}
	if maxSize <= 0 {
		stream.error(fmt.Errorf("maxSize (%d) should be > 0", maxSize))
		return
	}
	length := int32(len(*value))
	if length > int32(maxSize) {
		stream.error(fmt.Errorf("string is longer than maxSize"))
		return
	}
	min := int32(0)
	max := int32(maxSize - 1)
	stream.SerializeInteger(&length, min, max)
	stream.SerializeBytes([]byte(*value))
}

func (stream *WriteStream) SerializeIntRelative(previous *int32, current *int32) {
	if stream.err != nil {
		return
	}
	if previous == nil {
		stream.error(fmt.Errorf("previous is nil"))
		return
	}
	if current == nil {
		stream.error(fmt.Errorf("current is nil"))
		return
	}
	if *previous >= *current {
		stream.error(fmt.Errorf("previous value should be less than current value"))
		return
	}

	difference := *current - *previous

	oneBit := difference == 1
	stream.SerializeBool(&oneBit)
	if oneBit {
		return
	}

	twoBits := difference <= 6
	stream.SerializeBool(&twoBits)
	if twoBits {
		min := int32(2)
		max := int32(6)
		stream.SerializeInteger(&difference, min, max)
		return
	}

	fourBits := difference <= 23
	stream.SerializeBool(&fourBits)
	if fourBits {
		min := int32(7)
		max := int32(23)
		stream.SerializeInteger(&difference, min, max)
		return
	}

	eightBits := difference <= 280
	stream.SerializeBool(&eightBits)
	if eightBits {
		min := int32(24)
		max := int32(280)
		stream.SerializeInteger(&difference, min, max)
		return
	}

	twelveBits := difference <= 4377
	stream.SerializeBool(&twelveBits)
	if twelveBits {
		min := int32(281)
		max := int32(4377)
		stream.SerializeInteger(&difference, min, max)
		return
	}

	sixteenBits := difference <= 4377
	stream.SerializeBool(&sixteenBits)
	if sixteenBits {
		min := int32(4378)
		max := int32(69914)
		stream.SerializeInteger(&difference, min, max)
		return
	}

	uint32Value := uint32(*current)
	stream.SerializeUint32(&uint32Value)
}

func (stream *WriteStream) SerializeAddress(addr *net.UDPAddr) {
	if stream.err != nil {
		return
	}
	if addr == nil {
		stream.error(fmt.Errorf("addr is nil"))
		return
	}

	addrType := uint32(0)
	if addr.IP == nil {
		addrType = ADDRESS_NONE
	} else if addr.IP.To4() == nil {
		addrType = ADDRESS_IPV6
	} else {
		addrType = ADDRESS_IPV4
	}

	stream.SerializeBits(&addrType, 2)
	if stream.err != nil {
		return
	}
	if addrType == uint32(ADDRESS_IPV4) {
		stream.SerializeBytes(addr.IP[12:])
		if stream.err != nil {
			return
		}
		port := uint32(addr.Port)
		stream.SerializeBits(&port, 16)
	} else if addrType == uint32(ADDRESS_IPV6) {
		addr.IP = make([]byte, 16)
		for i := 0; i < 8; i++ {
			uint32Value := uint32(binary.BigEndian.Uint16(addr.IP[i*2:]))
			stream.SerializeBits(&uint32Value, 16)
			if stream.err != nil {
				return
			}
		}
		uint32Value := uint32(addr.Port)
		stream.SerializeBits(&uint32Value, 16)
	}
}

func (stream *WriteStream) SerializeAlign() {
	if stream.err != nil {
		return
	}
	stream.error(stream.writer.WriteAlign())
}

func (stream *WriteStream) GetAlignBits() int {
	return stream.writer.GetAlignBits()
}

func (stream *WriteStream) Flush() {
	if stream.err != nil {
		return
	}
	stream.error(stream.writer.FlushBits())
}

func (stream *WriteStream) GetData() []byte {
	return stream.writer.GetData()
}

func (stream *WriteStream) GetBytesProcessed() int {
	return stream.writer.GetBytesWritten()
}

func (stream *WriteStream) GetBitsProcessed() int {
	return stream.writer.GetBitsWritten()
}

// --------------------------------------------------------------------

type ReadStream struct {
	reader *BitReader
	err    error
}

func CreateReadStream(buffer []byte) *ReadStream {
	return &ReadStream{
		reader: CreateBitReader(buffer),
	}
}

func (stream *ReadStream) error(err error) {
	if err != nil && stream.err == nil {
		stream.err = fmt.Errorf("%v\n%s", err, string(debug.Stack()))
	}
}

func (stream *ReadStream) Error() error {
	return stream.err
}

func (stream *ReadStream) IsWriting() bool {
	return false
}

func (stream *ReadStream) IsReading() bool {
	return true
}

func (stream *ReadStream) SerializeInteger(value *int32, min int32, max int32) {
	if stream.err != nil {
		return
	}
	if value == nil {
		stream.error(fmt.Errorf("value is nil"))
		return
	}
	if min >= max {
		stream.error(fmt.Errorf("min (%d) should be less than max (%d)", min, max))
		return
	}
	bits := bitsRequiredSigned(min, max)
	if stream.reader.WouldReadPastEnd(bits) {
		stream.error(fmt.Errorf("would read past end of buffer"))
		return
	}
	unsignedValue, err := stream.reader.ReadBits(bits)
	if err != nil {
		stream.error(err)
		return
	}
	candidateValue := int32(unsignedValue) + min
	if candidateValue > max {
		stream.error(fmt.Errorf("value (%d) above max (%d)", candidateValue, max))
		return
	}
	*value = candidateValue
}

func (stream *ReadStream) SerializeBits(value *uint32, bits int) {
	if stream.err != nil {
		return
	}
	if value == nil {
		stream.error(fmt.Errorf("value is nil"))
		return
	}
	if bits < 0 || bits > 32 {
		stream.error(fmt.Errorf("bits (%d) should be between 0 and 32 bits", bits))
		return
	}
	if stream.reader.WouldReadPastEnd(bits) {
		stream.error(fmt.Errorf("would read past end of buffer"))
		return
	}
	readValue, err := stream.reader.ReadBits(bits)
	if err != nil {
		stream.error(err)
		return
	}
	*value = readValue
}

func (stream *ReadStream) SerializeUint32(value *uint32) {
	stream.SerializeBits(value, 32)
}

func (stream *ReadStream) SerializeBool(value *bool) {
	if stream.err != nil {
		return
	}
	if value == nil {
		stream.error(fmt.Errorf("value is nil"))
		return
	}
	if stream.reader.WouldReadPastEnd(1) {
		stream.error(fmt.Errorf("would read past end of buffer"))
		return
	}
	readValue, err := stream.reader.ReadBits(1)
	if err != nil {
		stream.error(err)
		return
	}
	*value = readValue != 0
}

func (stream *ReadStream) SerializeFloat32(value *float32) {
	if stream.err != nil {
		return
	}
	if value == nil {
		stream.error(fmt.Errorf("value is nil"))
		return
	}
	if stream.reader.WouldReadPastEnd(32) {
		stream.error(fmt.Errorf("would read past end of buffer"))
		return
	}
	readValue, err := stream.reader.ReadBits(32)
	if err != nil {
		stream.error(err)
		return
	}
	*value = math.Float32frombits(readValue)
}

func (stream *ReadStream) SerializeUint64(value *uint64) {
	if stream.err != nil {
		return
	}
	if value == nil {
		stream.error(fmt.Errorf("value is nil"))
		return
	}
	if stream.reader.WouldReadPastEnd(64) {
		stream.error(fmt.Errorf("would read past end of buffer"))
		return
	}
	lo, err := stream.reader.ReadBits(32)
	if err != nil {
		stream.error(err)
		return
	}
	hi, err := stream.reader.ReadBits(32)
	if err != nil {
		stream.error(err)
		return
	}
	*value = (uint64(hi) << 32) | uint64(lo)
}

func (stream *ReadStream) SerializeFloat64(value *float64) {
	if stream.err != nil {
		return
	}
	if value == nil {
		stream.error(fmt.Errorf("value is nil"))
		return
	}
	readValue := uint64(0)
	stream.SerializeUint64(&readValue)
	if stream.err != nil {
		return
	}
	*value = math.Float64frombits(readValue)
}

func (stream *ReadStream) SerializeBytes(data []byte) {
	if stream.err != nil {
		return
	}
	if len(data) == 0 {
		stream.error(fmt.Errorf("buffer should contain more than 0 bytes"))
		return
	}
	stream.SerializeAlign()
	if stream.err != nil {
		return
	}
	if stream.reader.WouldReadPastEnd(len(data) * 8) {
		stream.error(fmt.Errorf("SerializeBytes() would read past end of buffer"))
		return
	}
	stream.error(stream.reader.ReadBytes(data))
}

func (stream *ReadStream) SerializeString(value *string, maxSize int) {
	if stream.err != nil {
		return
	}
	if maxSize < 0 {
		stream.error(fmt.Errorf("maxSize (%d) should be > 0", maxSize))
		return
	}
	length := int32(0)
	min := int32(0)
	max := int32(maxSize - 1)
	stream.SerializeInteger(&length, min, max)
	if stream.err != nil {
		return
	}
	stringBytes := make([]byte, length)
	stream.SerializeBytes(stringBytes)
	*value = string(stringBytes)
}

func (stream *ReadStream) SerializeIntRelative(previous *int32, current *int32) {
	if stream.err != nil {
		return
	}
	if previous == nil {
		stream.error(fmt.Errorf("previous is nil"))
		return
	}
	if current == nil {
		stream.error(fmt.Errorf("current is nil"))
		return
	}
	oneBit := false
	stream.SerializeBool(&oneBit)
	if stream.err != nil {
		return
	}
	if oneBit {
		*current = *previous + 1
		return
	}

	twoBits := false
	stream.SerializeBool(&twoBits)
	if stream.err != nil {
		return
	}
	if twoBits {
		difference := int32(0)
		min := int32(2)
		max := int32(6)
		stream.SerializeInteger(&difference, min, max)
		*current = *previous + difference
		return
	}

	fourBits := false
	stream.SerializeBool(&fourBits)
	if stream.err != nil {
		return
	}
	if fourBits {
		difference := int32(0)
		min := int32(7)
		max := int32(32)
		stream.SerializeInteger(&difference, min, max)
		*current = *previous + difference
		return
	}

	eightBits := false
	stream.SerializeBool(&eightBits)
	if stream.err != nil {
		return
	}
	if eightBits {
		difference := int32(0)
		min := int32(24)
		max := int32(280)
		stream.SerializeInteger(&difference, min, max)
		*current = *previous + difference
		return
	}

	twelveBits := false
	stream.SerializeBool(&twelveBits)
	if stream.err != nil {
		return
	}
	if twelveBits {
		difference := int32(0)
		min := int32(281)
		max := int32(4377)
		stream.SerializeInteger(&difference, min, max)
		*current = *previous + difference
		return
	}

	sixteenBits := false
	stream.SerializeBool(&sixteenBits)
	if stream.err != nil {
		return
	}
	if sixteenBits {
		difference := int32(0)
		min := int32(4378)
		max := int32(69914)
		stream.SerializeInteger(&difference, min, max)
		*current = *previous + difference
		return
	}

	uint32Value := uint32(0)
	stream.SerializeUint32(&uint32Value)
	if stream.err != nil {
		return
	}
	*current = int32(uint32Value)
}

func (stream *ReadStream) SerializeAddress(addr *net.UDPAddr) {
	if stream.err != nil {
		return
	}
	if addr == nil {
		stream.error(fmt.Errorf("addr is nil"))
		return
	}
	addrType := uint32(0)
	stream.SerializeBits(&addrType, 2)
	if stream.err != nil {
		return
	}
	if addrType == uint32(ADDRESS_IPV4) {
		addr.IP = make([]byte, 16)
		addr.IP[10] = 255
		addr.IP[11] = 255
		stream.SerializeBytes(addr.IP[12:])
		if stream.err != nil {
			return
		}
		port := uint32(0)
		stream.SerializeBits(&port, 16)
		if stream.err != nil {
			return
		}
		addr.Port = int(port)
	} else if addrType == uint32(ADDRESS_IPV6) {
		addr.IP = make([]byte, 16)
		for i := 0; i < 8; i++ {
			uint32Value := uint32(0)
			stream.SerializeBits(&uint32Value, 16)
			if stream.err != nil {
				return
			}
			binary.BigEndian.PutUint16(addr.IP[i*2:], uint16(uint32Value))
		}
		uint32Value := uint32(0)
		stream.SerializeBits(&uint32Value, 16)
		if stream.err != nil {
			return
		}
		addr.Port = int(uint32Value)
	} else {
		*addr = net.UDPAddr{}
	}
}

func (stream *ReadStream) SerializeAlign() {
	alignBits := stream.reader.GetAlignBits()
	if stream.reader.WouldReadPastEnd(alignBits) {
		stream.error(fmt.Errorf("SerializeAlign() would read past end of buffer"))
		return
	}
	stream.error(stream.reader.ReadAlign())
}

func (stream *ReadStream) Flush() {
}

func (stream *ReadStream) GetAlignBits() int {
	return stream.reader.GetAlignBits()
}

func (stream *ReadStream) GetBitsProcessed() int {
	return stream.reader.GetBitsRead()
}

func (stream *ReadStream) GetBytesProcessed() int {
	return (stream.reader.GetBitsRead() + 7) / 8
}

// -------------------------------------------------------------------------------------

func ProtocolVersionAtLeast(serverMajor int32, serverMinor int32, serverPatch int32, targetMajor int32, targetMinor int32, targetPatch int32) bool {
	serverVersion := ( (serverMajor&0xFF) << 16 ) | ( (serverMinor&0xFF) << 8 ) | (serverPatch&0xFF);
	targetVersion := ( (targetMajor&0xFF) << 16 ) | ( (targetMinor&0xFF) << 8 ) | (targetPatch&0xFF);
	return serverVersion >= targetVersion
}

// -----------------------------------------------------------------------------

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

func ReadAddress(buffer []byte) *net.UDPAddr {
	addressType := buffer[0]
	if addressType == ADDRESS_IPV4 {
		return &net.UDPAddr{IP: net.IPv4(buffer[1], buffer[2], buffer[3], buffer[4]), Port: ((int)(binary.LittleEndian.Uint16(buffer[5:])))}
	} else if addressType == ADDRESS_IPV6 {
		return &net.UDPAddr{IP: buffer[1:], Port: ((int)(binary.LittleEndian.Uint16(buffer[17:])))}
	}
	return nil
}

func ParseAddress(input string) *net.UDPAddr {
	address := &net.UDPAddr{}
	ip_string, port_string, err := net.SplitHostPort(input)
	if err != nil {
		address.IP = net.ParseIP(input)
		address.Port = 0
		return address
	}
	address.IP = net.ParseIP(ip_string)
	address.Port, _ = strconv.Atoi(port_string)
	return address
}

func ParseKeyFromBase64(input_base64 string) []byte {
	input, err := base64.StdEncoding.DecodeString(input_base64)
	if err != nil {
		return nil
	}
	return CheckKey(input)
}

func CheckKey(input []byte) []byte {
	if len(input) != KeyBytes {
		return nil
	}
	return input
}

func ParseAddressFromBase64(input_base64 string) *net.UDPAddr {
	input, err := base64.StdEncoding.DecodeString(input_base64)
	if err != nil {
		return nil
	}
	return ParseAddress(string(input))
}

func Checksum(data []byte) []byte {
	hasher := sha256.New()
	hasher.Write(data)
	return hasher.Sum(nil)
}

// -----------------------------------------------------------------------------

const Crypto_kx_PUBLICKEYBYTES = C.crypto_kx_PUBLICKEYBYTES
const Crypto_box_PUBLICKEYBYTES = C.crypto_box_PUBLICKEYBYTES

const KeyBytes = 32
const NonceBytes = 24
const SignatureBytes = C.crypto_sign_BYTES
const PublicKeyBytes = C.crypto_sign_PUBLICKEYBYTES

func Encrypt(senderPrivateKey []byte, receiverPublicKey []byte, nonce []byte, buffer []byte, bytes int) error {
	result := C.crypto_box_easy((*C.uchar)(&buffer[0]),
		(*C.uchar)(&buffer[0]),
		C.ulonglong(bytes),
		(*C.uchar)(&nonce[0]),
		(*C.uchar)(&receiverPublicKey[0]),
		(*C.uchar)(&senderPrivateKey[0]))
	if result != 0 {
		return fmt.Errorf("failed to encrypt: result = %d", result)
	} else {
		return nil
	}
}

func Decrypt(senderPublicKey []byte, receiverPrivateKey []byte, nonce []byte, buffer []byte, bytes int) error {
	result := C.crypto_box_open_easy(
		(*C.uchar)(&buffer[0]),
		(*C.uchar)(&buffer[0]),
		C.ulonglong(bytes),
		(*C.uchar)(&nonce[0]),
		(*C.uchar)(&senderPublicKey[0]),
		(*C.uchar)(&receiverPrivateKey[0]))
	if result != 0 {
		return fmt.Errorf("failed to decrypt: result = %d", result)
	} else {
		return nil
	}
}

func Encrypt_ChaCha20(buffer []byte, additional []byte, privateKey []byte) ([]byte, []byte, error) {
	nonce := RandomBytes(C.crypto_aead_xchacha20poly1305_ietf_NPUBBYTES)
	encrypted := make([]byte, len(buffer)+C.crypto_aead_xchacha20poly1305_ietf_ABYTES)
	var encryptedLength = C.ulonglong(0)
	result := C.crypto_aead_xchacha20poly1305_ietf_encrypt((*C.uchar)(&encrypted[0]), &encryptedLength,
		(*C.uchar)(&buffer[0]), C.ulonglong(len(buffer)),
		(*C.uchar)(&additional[0]), C.ulonglong(len(additional)),
		nil, (*C.uchar)(&nonce[0]), (*C.uchar)(&privateKey[0]))
	if result != 0 {
		return nil, nil, fmt.Errorf("failed to encrypt chacha20: result = %d", result)
	} else {
		return encrypted, nonce, nil
	}
}

func Decrypt_ChaCha20(encrypted []byte, additional []byte, nonce []byte, privateKey []byte) ([]byte, error) {
	if len(encrypted) <= C.crypto_aead_xchacha20poly1305_ietf_ABYTES {
		return nil, fmt.Errorf("failed to decrypt chacha20: encrypted data is too small")
	}
	decrypted := make([]byte, len(encrypted)-C.crypto_aead_xchacha20poly1305_ietf_ABYTES)
	var decryptedLength = C.ulonglong(0)
	result := C.crypto_aead_xchacha20poly1305_ietf_decrypt((*C.uchar)(&decrypted[0]), &decryptedLength, nil,
		(*C.uchar)(&encrypted[0]), C.ulonglong(len(encrypted)),
		(*C.uchar)(&additional[0]), C.ulonglong(len(additional)),
		(*C.uchar)(&nonce[0]), (*C.uchar)(&privateKey[0]))
	if result != 0 {
		return nil, fmt.Errorf("failed to decrypt chacha20: result = %d", result)
	} else {
		return decrypted, nil
	}
}

func RandomBytes(bytes int) []byte {
	buffer := make([]byte, bytes)
	C.randombytes_buf(unsafe.Pointer(&buffer[0]), C.size_t(bytes))
	return buffer
}

func WriteRouteToken(token *RouteToken, buffer []byte) {
	binary.LittleEndian.PutUint64(buffer[0:], token.expireTimestamp)
	binary.LittleEndian.PutUint64(buffer[8:], token.sessionId)
	buffer[8+8] = token.sessionVersion
	binary.LittleEndian.PutUint32(buffer[8+8+1:], token.kbpsUp)
	binary.LittleEndian.PutUint32(buffer[8+8+1+4:], token.kbpsDown)
	WriteAddress(buffer[8+8+1+4+4:], token.nextAddress)
	copy(buffer[8+8+1+4+4+NEXT_ADDRESS_BYTES:], token.privateKey)
}

func WriteContinueToken(token *ContinueToken, buffer []byte) {
	binary.LittleEndian.PutUint64(buffer[0:], token.expireTimestamp)
	binary.LittleEndian.PutUint64(buffer[8:], token.sessionId)
	buffer[8+8] = token.sessionVersion
}

func ReadContinueToken(buffer []byte) (*ContinueToken, error) {
	if len(buffer) < NEXT_CONTINUE_TOKEN_BYTES {
		return nil, fmt.Errorf("buffer too small to read continue token")
	}
	token := &ContinueToken{}
	token.expireTimestamp = binary.LittleEndian.Uint64(buffer[0:])
	token.sessionId = binary.LittleEndian.Uint64(buffer[8:])
	token.sessionVersion = buffer[8+8]
	return token, nil
}

func WriteEncryptedContinueToken(buffer []byte, token *ContinueToken, senderPrivateKey []byte, receiverPublicKey []byte) error {
	nonce := RandomBytes(NonceBytes)
	copy(buffer, nonce)
	WriteContinueToken(token, buffer[NonceBytes:])
	result := Encrypt(senderPrivateKey, receiverPublicKey, nonce, buffer[NonceBytes:], NEXT_CONTINUE_TOKEN_BYTES)
	return result
}

func WriteEncryptedRouteToken(buffer []byte, token *RouteToken, senderPrivateKey []byte, receiverPublicKey []byte, nonce []byte) error {
	copy(buffer, nonce)
	WriteRouteToken(token, buffer[NonceBytes:])
	result := Encrypt(senderPrivateKey, receiverPublicKey, nonce, buffer[NonceBytes:], NEXT_ROUTE_TOKEN_BYTES)
	return result
}

func ReadEncryptedContinueToken(tokenData []byte, senderPublicKey []byte, receiverPrivateKey []byte) (*ContinueToken, error) {
	if len(tokenData) < NEXT_ENCRYPTED_CONTINUE_TOKEN_BYTES {
		return nil, fmt.Errorf("not enough bytes for encrypted continue token")
	}
	nonce := tokenData[0 : C.crypto_box_NONCEBYTES-1]
	tokenData = tokenData[C.crypto_box_NONCEBYTES:]
	if err := Decrypt(senderPublicKey, receiverPrivateKey, nonce, tokenData, NEXT_CONTINUE_TOKEN_BYTES+C.crypto_box_MACBYTES); err != nil {
		return nil, err
	}
	return ReadContinueToken(tokenData)
}

func WriteRouteTokens(expireTimestamp uint64, sessionId uint64, sessionVersion uint8, kbpsUp uint32, kbpsDown uint32, numNodes int, addresses []*net.UDPAddr, publicKeys [][]byte, masterPrivateKey [KeyBytes]byte) ([]byte, error) {
	if numNodes < 1 || numNodes > NEXT_MAX_NODES {
		return nil, fmt.Errorf("invalid numNodes %d. expected value in range [1,%d]", numNodes, NEXT_MAX_NODES)
	}
	privateKey := RandomBytes(KeyBytes)
	tokenData := make([]byte, numNodes*NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES)
	for i := 0; i < numNodes; i++ {
		nonce := RandomBytes(NonceBytes)
		token := &RouteToken{}
		token.expireTimestamp = expireTimestamp
		token.sessionId = sessionId
		token.sessionVersion = sessionVersion
		token.kbpsUp = kbpsUp
		token.kbpsDown = kbpsDown
		if i != numNodes-1 {
			token.nextAddress = addresses[i+1]
		}
		token.privateKey = privateKey
		err := WriteEncryptedRouteToken(tokenData[i*NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES:], token, masterPrivateKey[:], publicKeys[i], nonce)
		if err != nil {
			return nil, err
		}
	}
	return tokenData, nil
}

func WriteContinueTokens(expireTimestamp uint64, sessionId uint64, sessionVersion uint8, numNodes int, publicKeys [][]byte, masterPrivateKey [KeyBytes]byte) ([]byte, error) {
	if numNodes < 1 || numNodes > NEXT_MAX_NODES {
		return nil, fmt.Errorf("invalid numNodes %d. expected value in range [1,%d]", numNodes, NEXT_MAX_NODES)
	}
	tokenData := make([]byte, numNodes*NEXT_ENCRYPTED_CONTINUE_TOKEN_BYTES)
	for i := 0; i < numNodes; i++ {
		token := &ContinueToken{}
		token.expireTimestamp = expireTimestamp
		token.sessionId = sessionId
		token.sessionVersion = sessionVersion
		err := WriteEncryptedContinueToken(tokenData[i*NEXT_ENCRYPTED_CONTINUE_TOKEN_BYTES:], token, masterPrivateKey[:], publicKeys[i])
		if err != nil {
			return nil, err
		}
	}
	return tokenData, nil
}

// --------------------------------------------------------

type RouteToken struct {
	expireTimestamp uint64
	sessionId       uint64
	sessionVersion  uint8
	kbpsUp          uint32
	kbpsDown        uint32
	nextAddress     *net.UDPAddr
	privateKey      []byte
}

type ContinueToken struct {
	expireTimestamp uint64
	sessionId       uint64
	sessionVersion  uint8
}

// -------------------------------------------------------

func main() {

	rand.Seed(time.Now().UnixNano())

	backend.relayDatabase = make(map[string]RelayEntry)
	backend.serverDatabase = make(map[string]ServerEntry)
	backend.sessionDatabase = make(map[uint64]SessionEntry)

	go TimeoutThread()

	go WebServer()

	listenAddress := net.UDPAddr{
		Port: NEXT_SERVER_BACKEND_PORT,
		IP:   net.ParseIP("0.0.0.0"),
	}

	connection, err := net.ListenUDP("udp", &listenAddress)
	if err != nil {
		fmt.Printf("error: could not listen on %s\n", listenAddress.String())
		return
	}

	defer connection.Close()

	fmt.Printf("\nreference backend (sdk4)\n\n")
	
	for {

		packetData := make([]byte, NEXT_MAX_PACKET_BYTES)

		packetBytes, from, err := connection.ReadFromUDP(packetData)

		packetData = packetData[:packetBytes]

		if err != nil {
			fmt.Printf("socket error: %v\n", err)
			continue
		}

		if packetBytes <= 0 {
			continue
		}

		if !IsNetworkNextPacket(packetData) {
			fmt.Printf("error: not network next packet (%d)\n", packetData[8])
			continue
		}

		packetData = packetData[NEXT_PACKET_HASH_BYTES:]
		packetBytes -= NEXT_PACKET_HASH_BYTES

		packetType := packetData[0]

		if packetType == NEXT_BACKEND_SERVER_INIT_REQUEST_PACKET {

			readStream := CreateReadStream(packetData[1:])

			serverInitRequest := &NextBackendServerInitRequestPacket{}
			if err := serverInitRequest.Serialize(readStream); err != nil {
				fmt.Printf("error: failed to read server init request packet: %v\n", err)
				continue
			}

			initResponse := &NextBackendServerInitResponsePacket{}
			initResponse.RequestId = serverInitRequest.RequestId
			initResponse.Response = NEXT_SERVER_INIT_RESPONSE_OK

			writeStream, err := CreateWriteStream(NEXT_MAX_PACKET_BYTES)
			if err != nil {
				fmt.Printf("error: failed to write server init response packet: %v\n", err)
				continue
			}

			responsePacketType := uint32(NEXT_BACKEND_SERVER_INIT_RESPONSE_PACKET)
			writeStream.SerializeBits(&responsePacketType, 8)
			if err := initResponse.Serialize(writeStream); err != nil {
				fmt.Printf("error: failed to write server init response packet: %v\n", err)
				continue
			}
			writeStream.Flush()

			responsePacketData := writeStream.GetData()[0:writeStream.GetBytesProcessed()]

			signedResponsePacketData := SignNetworkNextPacket(responsePacketData, backendPrivateKey[:])

			hashedResponsePacketData := HashNetworkNextPacket(signedResponsePacketData)

			_, err = connection.WriteToUDP(hashedResponsePacketData, from)
			if err != nil {
				fmt.Printf("error: failed to send udp response: %v\n", err)
				continue
			}
		
		} else if packetType == NEXT_BACKEND_SERVER_UPDATE_PACKET {

			readStream := CreateReadStream(packetData[1:])

			serverUpdate := &NextBackendServerUpdatePacket{}
			if err := serverUpdate.Serialize(readStream); err != nil {
				fmt.Printf("error: failed to read server update packet: %v\n", err)
				continue
			}

			serverEntry := ServerEntry{}
			serverEntry.address = from
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

			readStream := CreateReadStream(packetData[1:])
			sessionUpdate := &NextBackendSessionUpdatePacket{}
			if err := sessionUpdate.Serialize(readStream); err != nil {
				fmt.Printf("error: failed to read server session update packet: %v\n", err)
				continue
			}

			sessionDataReadStream := CreateReadStream(sessionUpdate.SessionData[:sessionUpdate.SessionDataBytes])
			var sessionData SessionData
			sessionData.Version = SessionDataVersion
			if sessionUpdate.SliceNumber != 0 {
				err := sessionData.Serialize(sessionDataReadStream)
				if err != nil {
					fmt.Printf("error: could not read session data: %v\n", err)
					continue
				}
			}

			// todo: various checks on the session data to make sure it's valid

			sessionData.Version = SessionDataVersion
			sessionData.SessionId = sessionUpdate.SessionId
			sessionData.SliceNumber = sessionUpdate.SliceNumber + 1

			var sessionResponse *NextBackendSessionResponsePacket

			backend.mutex.RLock()
			sessionEntry := backend.sessionDatabase[sessionUpdate.SessionId]
			sessionEntry.expireTimestamp = uint64(time.Now().Unix()) + 15
			backend.mutex.RUnlock()

			nearRelayIds, nearRelayAddresses := GetNearRelays()

			takeNetworkNext := len(nearRelayIds) > 0

			if !takeNetworkNext {

				// direct route

				sessionResponse = &NextBackendSessionResponsePacket{
					SessionId:            sessionUpdate.SessionId,
					SliceNumber:          sessionUpdate.SliceNumber,
					RouteType:            int32(NEXT_ROUTE_TYPE_DIRECT),
					NumNearRelays:        int32(len(nearRelayIds)),
					NearRelayIds:         nearRelayIds,
					NearRelayAddresses:   nearRelayAddresses,
					NumTokens:            0,
					Tokens:               nil,
				}

			} else {

				// next route

				numRelays := len(nearRelayIds)
				if numRelays > 5 {
					numRelays = 5
				}

				route := make([]uint64, numRelays)
				for i := 0; i < numRelays; i++ {
					route[i] = nearRelayIds[i]
				}

				routeChanged := RouteChanged(sessionData.Route, route)

				numNodes := numRelays + 2

				addresses := make([]*net.UDPAddr, numNodes)
				publicKeys := make([][]byte, numNodes)
				publicKeys[0] = sessionUpdate.ClientRoutePublicKey

				for i := 0; i < numRelays; i++ {
					addresses[1+i] = &nearRelayAddresses[i]
					publicKeys[1+i] = relayPublicKey
				}

				addresses[numNodes-1] = from
				publicKeys[numNodes-1] = sessionUpdate.ServerRoutePublicKey

				var tokens []byte

				var routeType int32

				if sessionData.ExpireTimestamp == 0 {
					sessionData.ExpireTimestamp = uint64(time.Now().Unix())
				}

				if routeChanged {

					// new route

					routeType = NEXT_ROUTE_TYPE_NEW

					sessionData.ExpireTimestamp += 20
					sessionData.SessionVersion += 1

					tokens, err = WriteRouteTokens(sessionData.ExpireTimestamp, sessionData.SessionId, uint8(sessionData.SessionVersion), 256, 256, numNodes, addresses, publicKeys, routerPrivateKey)

					if err != nil {
						fmt.Printf("error: could not write route tokens: %v\n", err)
						continue
					}


				} else {

					// continue route

					routeType = NEXT_ROUTE_TYPE_CONTINUE

					sessionData.ExpireTimestamp += 10

					tokens, err = WriteContinueTokens(sessionData.ExpireTimestamp, sessionData.SessionId, uint8(sessionData.SessionVersion), numNodes, publicKeys, routerPrivateKey)

					if err != nil {
						fmt.Printf("error: could not write continue tokens: %v\n", err)
						continue
					}

				}

				sessionResponse = &NextBackendSessionResponsePacket{
					SliceNumber:          sessionUpdate.SliceNumber,
					SessionId:            sessionUpdate.SessionId,
					NumNearRelays:        int32(len(nearRelayIds)),
					NearRelayIds:         nearRelayIds,
					NearRelayAddresses:   nearRelayAddresses,
					RouteType:            routeType,
					NumTokens:            int32(numNodes),
					Tokens:               tokens,
					Multipath:            false,
					Committed:            true,
				}

				sessionData.Route = route
			}

			if sessionResponse == nil {
				fmt.Printf("error: nil session response\n")
				continue
			}

			backend.mutex.Lock()
			if sessionData.SliceNumber == 0 {
				backend.dirty = true
			}
			backend.sessionDatabase[sessionUpdate.SessionId] = sessionEntry
			backend.mutex.Unlock()

			sessionData.Version = SessionDataVersion
			sessionData.SessionId = sessionUpdate.SessionId
			sessionData.SliceNumber = sessionUpdate.SliceNumber + 1

			sessionDataWriteStream, err := CreateWriteStream(1024)
			if err != nil {
				fmt.Printf("error: failed to create write stream for session data: %v\n", err)
				continue
			}
			if err := sessionData.Serialize(sessionDataWriteStream); err != nil {
				fmt.Printf("error: failed to write session data: %v\n", err)
				continue
			}
			sessionDataWriteStream.Flush()

			if sessionResponse.SessionDataBytes > NEXT_MAX_SESSION_DATA_BYTES {
				panic("session data is too large")
			}
			
			sessionResponse.SessionDataBytes = int32(sessionDataWriteStream.GetBytesProcessed())
			copy(sessionResponse.SessionData[:], sessionDataWriteStream.GetData()[0:sessionDataWriteStream.GetBytesProcessed()])

			writeStream, err := CreateWriteStream(NEXT_MAX_PACKET_BYTES)
			if err != nil {
				fmt.Printf("error: failed to create write stream for session response packet: %v\n", err)
				continue
			}
			responsePacketType := uint32(NEXT_BACKEND_SESSION_RESPONSE_PACKET)
			writeStream.SerializeBits(&responsePacketType, 8)
			if err := sessionResponse.Serialize(writeStream, sessionUpdate.VersionMajor, sessionUpdate.VersionMinor, sessionUpdate.VersionPatch); err != nil {
				fmt.Printf("error: failed to write session response packet: %v\n", err)
				continue
			}
			writeStream.Flush()

			responsePacketData := writeStream.GetData()[0:writeStream.GetBytesProcessed()]

			signedResponsePacketData := SignNetworkNextPacket(responsePacketData, backendPrivateKey[:])

			hashedResponsePacketData := HashNetworkNextPacket(signedResponsePacketData)

			_, err = connection.WriteToUDP(hashedResponsePacketData, from)
			if err != nil {
				fmt.Printf("error: failed to send udp response: %v\n", err)
				continue
			}
		}
	}
}
