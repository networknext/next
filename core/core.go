// Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.

package core

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"log"
	"math"
	"net"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
)

const (
	SDKVersionMajorMin = 3
	SDKVersionMinorMin = 3
	SDKVersionPatchMin = 2
	SDKVersionMajorMax = 254
	SDKVersionMinorMax = 1023
	SDKVersionPatchMax = 254

	MaxRelays                   = 5
	MaxNodes                    = 7
	MaxNearRelays               = 10
	BillingSliceSeconds         = 10
	MinimumKbps                 = 100
	AddressBytes                = 19
	SessionTokenBytes           = 77
	EncryptedSessionTokenBytes  = 117
	ContinueTokenBytes          = 18
	EncryptedContinueTokenBytes = 58
	MaxRoutesPerRelayPair       = 8

	IPAddressNone = 0
	IPAddressIPv4 = 1
	IPAddressIPv6 = 2

	ConnectionTypeUnknown  = 0
	ConnectionTypeWired    = 1
	ConnectionTypeWifi     = 2
	ConnectionTypeCellular = 3

	PlatformUnknown = 0
	PlatformWindows = 1
	PlatformMac     = 2
	PlatformUnix    = 3
	PlatformSwitch  = 4
	PlatformPS4     = 5
	PlatformIOS     = 6
	PlatformXboxOne = 7

	RouteSliceFlagNext                = (uint64(1) << 1)
	RouteSliceFlagReported            = (uint64(1) << 2)
	RouteSliceFlagVetoed              = (uint64(1) << 3)
	RouteSliceFlagFallbackToDirect    = (uint64(1) << 4)
	RouteSliceFlagPacketLossMultipath = (uint64(1) << 5)
	RouteSliceFlagJitterMultipath     = (uint64(1) << 6)
	RouteSliceFlagRTTMultipath        = (uint64(1) << 7)

	FlagBadRouteToken           = uint32(1 << 0)
	FlagNoRouteToContinue       = uint32(1 << 1)
	FlagPreviousUpdatePending   = uint32(1 << 2)
	FlagBadContinueToken        = uint32(1 << 3)
	FlagRouteExpired            = uint32(1 << 4)
	FlagRouteRequestTimedOut    = uint32(1 << 5)
	FlagContinueRequestTimedOut = uint32(1 << 6)
	FlagClientTimedOut          = uint32(1 << 7)
	FlagTryBeforeYouBuyAbort    = uint32(1 << 8)
	FlagDirectRouteExpired      = uint32(1 << 9)
	FlagTotalCount              = 10
)

var RouterPrivateKey = [...]byte{0x96, 0xce, 0x57, 0x8b, 0x00, 0x19, 0x44, 0x27, 0xf2, 0xb9, 0x90, 0x1b, 0x43, 0x56, 0xfd, 0x4f, 0x56, 0xe1, 0xd9, 0x56, 0x58, 0xf2, 0xf4, 0x3b, 0x86, 0x9f, 0x12, 0x75, 0x24, 0xd2, 0x47, 0xb3}

var BackendPrivateKey = []byte{21, 124, 5, 171, 56, 198, 148, 140, 20, 15, 8, 170, 212, 222, 84, 155, 149, 84, 122, 199, 107, 225, 243, 246, 133, 85, 118, 114, 114, 126, 200, 4, 76, 97, 202, 140, 71, 135, 62, 212, 160, 181, 151, 195, 202, 224, 207, 113, 8, 45, 37, 60, 145, 14, 212, 111, 25, 34, 175, 186, 37, 150, 163, 64}

var MaxJitter float32 = 10.0
var MaxPacketLoss float32 = 0.1

const InvalidRouteValue = 10000.0
const InvalidHistoryValue = -1
const RelayTimeoutSeconds = 60

// ============================================================================

func GenerateRelayKeyPair() ([]byte, []byte) {
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		log.Fatalln(err)
	}
	return publicKey, privateKey
}

func GenerateCustomerKeyPair() ([]byte, []byte) {
	customerId := make([]byte, 8)
	rand.Read(customerId)
	publicKey, privateKey, err := ed25519.GenerateKey(nil)
	if err != nil {
		log.Fatalln(err)
	}
	customerPublicKey := make([]byte, 0)
	customerPublicKey = append(customerPublicKey, customerId...)
	customerPublicKey = append(customerPublicKey, publicKey...)
	customerPrivateKey := make([]byte, 0)
	customerPrivateKey = append(customerPrivateKey, customerId...)
	customerPrivateKey = append(customerPrivateKey, privateKey...)
	return customerPublicKey, customerPrivateKey
}

// ============================================================================

type EntityId struct {
	Kind string
	Name string
}

type RelayId uint64

func GetRelayId(id *EntityId) (RelayId, error) {
	if id.Kind != "Relay" {
		return RelayId(0), fmt.Errorf("not a valid relay: %+v", id)
	}
	hash := fnv.New64a()
	hash.Write([]byte(id.Name))
	return RelayId(hash.Sum64()), nil
}

type DatacenterId uint64

func GetDatacenterId(id *EntityId) (DatacenterId, error) {
	if id.Kind != "Datacenter" {
		return DatacenterId(0), fmt.Errorf("not a valid datacenter: %+v", id)
	}
	hash := fnv.New64a()
	hash.Write([]byte(id.Name))
	return DatacenterId(hash.Sum64()), nil
}

// ============================================================================

func WriteString(buffer []byte, value string) int {
	binary.LittleEndian.PutUint32(buffer, uint32(len(value)))
	copy(buffer[4:], []byte(value))
	return 4 + len([]byte(value))
}

func ReadString(buffer []byte) (string, int) {
	stringLength := binary.LittleEndian.Uint32(buffer)
	stringData := make([]byte, stringLength)
	copy(stringData, buffer[4:4+stringLength])
	return string(stringData), int(4 + stringLength)
}

func WriteBytes(buffer []byte, value []byte) int {
	binary.LittleEndian.PutUint32(buffer, uint32(len(value)))
	copy(buffer[4:], value)
	return 4 + len(value)
}

func ReadBytes(buffer []byte) ([]byte, int) {
	length := binary.LittleEndian.Uint32(buffer)
	data := make([]byte, length)
	copy(data, buffer[4:4+length])
	return data, int(4 + length)
}

func WriteAddress(buffer []byte, address *net.UDPAddr) {
	if address == nil {
		buffer[0] = IPAddressNone
		return
	}
	ipv4 := address.IP.To4()
	port := address.Port
	if ipv4 != nil {
		buffer[0] = IPAddressIPv4
		buffer[1] = ipv4[0]
		buffer[2] = ipv4[1]
		buffer[3] = ipv4[2]
		buffer[4] = ipv4[3]
		buffer[5] = (byte)(port & 0xFF)
		buffer[6] = (byte)(port >> 8)
	} else {
		buffer[0] = IPAddressIPv6
		copy(buffer[1:], address.IP)
		buffer[17] = (byte)(port & 0xFF)
		buffer[18] = (byte)(port >> 8)
	}
}

func ReadAddress(buffer []byte) *net.UDPAddr {
	addressType := buffer[0]
	switch addressType {
	case IPAddressIPv4:
		return &net.UDPAddr{IP: net.IPv4(buffer[1], buffer[2], buffer[3], buffer[4]), Port: ((int)(binary.LittleEndian.Uint16(buffer[5:])))}
	case IPAddressIPv6:
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

// ===========================================================================

type SessionToken struct {
	expireTimestamp uint64
	sessionId       uint64
	sessionVersion  uint8
	sessionFlags    uint8
	kbpsUp          uint32
	kbpsDown        uint32
	nextAddress     *net.UDPAddr
	privateKey      []byte
}

type ContinueToken struct {
	expireTimestamp uint64
	sessionId       uint64
	sessionVersion  uint8
	sessionFlags    uint8
}

// =====================================================================================================

func HaversineDistance(lat1 float64, long1 float64, lat2 float64, long2 float64) float64 {
	lat1 *= math.Pi / 180
	lat2 *= math.Pi / 180
	long1 *= math.Pi / 180
	long2 *= math.Pi / 180
	delta_lat := lat2 - lat1
	delta_long := long2 - long1
	lat_sine := math.Sin(delta_lat / 2)
	long_sine := math.Sin(delta_long / 2)
	a := lat_sine*lat_sine + math.Cos(lat1)*math.Cos(lat2)*long_sine*long_sine
	c := 2 * math.Atan2(math.Sqrt(a), math.Sqrt(1-a))
	r := 6371.0
	d := r * c
	return d // kilometers
}

func TriMatrixLength(size int) int {
	return (size * (size - 1)) / 2
}

func TriMatrixIndex(i, j int) int {
	if i <= j {
		i, j = j, i
	}
	return i*(i+1)/2 - i + j
}

func NibblinsToDollarString(nibblins int64) string {
	cents := float64(nibblins) / 1e9
	dollars := cents / 100
	return fmt.Sprintf("%f", dollars)
}

func DollarStringToNibblins(str string) (int64, error) {
	if len(str) == 0 {
		return 0, nil
	}
	decimal := strings.Index(str, ".")
	if decimal == -1 {
		decimal = len(str)
	}

	start := 0
	if str[0] == '-' {
		start = 1
	}

	dollars := int64(0)
	if decimal > start {
		var err error
		dollars, err = strconv.ParseInt(str[start:decimal], 10, 64)
		if err != nil {
			return 0, err
		}
	}
	nibblins := int64(0)
	if decimal+1 < len(str) {
		length := len(str) - (decimal + 1)
		if length < 11 {
			length = 11
		}
		for i := 0; i < length; i += 1 {
			if i < 11 {
				nibblins *= 10
			}
			index := decimal + 1 + i
			if index < len(str) {
				char := str[index]
				if char < byte('0') || char > byte('9') {
					return 0, fmt.Errorf("invalid dollar string: %s", str)
				}
				if i < 11 {
					nibblins += int64(char - byte('0'))
				}
			}
		}
	}
	if str[0] == '-' {
		dollars = -dollars
		nibblins = -nibblins
	}
	return (dollars * 1e11) + nibblins, nil
}

// ====================================================================

func ProtocolVersionAtLeast(serverVersionMajor int32, serverVersionMinor int32, serverVersionPatch int32, targetProtocolVersionMajor int32, targetProtocolVersionMinor int32, targetProtocolVersionPatch int32) bool {
	if serverVersionMajor == 0 && serverVersionMinor == 0 && serverVersionPatch == 0 {
		// This is an internal build, assume latest version.
		return true
	}

	if serverVersionMajor > targetProtocolVersionMajor {
		// The server has a major version newer than the target, ignore minor and patch numbers and pass.
		return true
	}

	if serverVersionMajor == targetProtocolVersionMajor {
		// The server has a matching major version, now check minor version.

		if serverVersionMinor > targetProtocolVersionMinor {
			// The server has a minor version newer than the target, ignore patch number and pass.
			return true
		}

		if serverVersionMinor == targetProtocolVersionMinor {
			// The server has a matching minor version, now check patch version.

			if serverVersionPatch >= targetProtocolVersionPatch {
				// The patch version is newer or equal to the desired version, pass.
				return true
			}
		}
	}

	// Server version is not new enough.
	return false
}

// ===========================================================

const NEXT_MAX_NEAR_RELAYS = 32
const NEXT_UPDATE_TYPE_DIRECT = 0
const NEXT_UPDATE_TYPE_ROUTE = 1
const NEXT_UPDATE_TYPE_CONTINUE = 2
const NEXT_MAX_TOKENS = 7
const NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES = 117
const NEXT_ENCRYPTED_CONTINUE_TOKEN_BYTES = 58
const NEXT_MTU = 1300

type ServerUpdatePacket struct {
	Sequence             uint64
	VersionMajor         int32
	VersionMinor         int32
	VersionPatch         int32
	CustomerId           uint64
	DatacenterId         uint64
	NumSessionsPending   uint32
	NumSessionsUpgraded  uint32
	ServerAddress        net.UDPAddr
	ServerPrivateAddress net.UDPAddr
	ServerRoutePublicKey []byte
	Signature            []byte
}

func (packet *ServerUpdatePacket) UnmarshalBinary(data []byte) error {
	if err := packet.Serialize(CreateReadStream(data)); err != nil {
		return err
	}
	return nil
}

func (packet *ServerUpdatePacket) MarshalBinary() ([]byte, error) {
	ws, err := CreateWriteStream(1500)
	if err != nil {
		return nil, err
	}

	if err := packet.Serialize(ws); err != nil {
		return nil, err
	}
	ws.Flush()

	return ws.GetData(), nil
}

func (packet *ServerUpdatePacket) Serialize(stream Stream) error {
	stream.SerializeUint64(&packet.Sequence)
	stream.SerializeInteger(&packet.VersionMajor, 0, SDKVersionMajorMax)
	stream.SerializeInteger(&packet.VersionMinor, 0, SDKVersionMinorMax)
	stream.SerializeInteger(&packet.VersionPatch, 0, SDKVersionPatchMax)
	stream.SerializeUint64(&packet.CustomerId)
	stream.SerializeUint64(&packet.DatacenterId)
	stream.SerializeUint32(&packet.NumSessionsPending)
	stream.SerializeUint32(&packet.NumSessionsUpgraded)
	stream.SerializeAddress(&packet.ServerAddress)
	stream.SerializeAddress(&packet.ServerPrivateAddress)
	if stream.IsReading() {
		packet.ServerRoutePublicKey = make([]byte, Crypto_box_PUBLICKEYBYTES)
		packet.Signature = make([]byte, SignatureBytes)
	}
	stream.SerializeBytes(packet.ServerRoutePublicKey)
	stream.SerializeBytes(packet.Signature)
	return stream.Error()
}

func (packet *ServerUpdatePacket) GetSignData() []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, packet.Sequence)
	binary.Write(buf, binary.LittleEndian, uint64(packet.VersionMajor))
	binary.Write(buf, binary.LittleEndian, uint64(packet.VersionMinor))
	binary.Write(buf, binary.LittleEndian, uint64(packet.VersionPatch))
	binary.Write(buf, binary.LittleEndian, packet.CustomerId)
	binary.Write(buf, binary.LittleEndian, packet.DatacenterId)
	binary.Write(buf, binary.LittleEndian, packet.NumSessionsPending)
	binary.Write(buf, binary.LittleEndian, packet.NumSessionsUpgraded)

	address := make([]byte, AddressBytes)
	WriteAddress(address, &packet.ServerAddress)
	binary.Write(buf, binary.LittleEndian, address)

	privateAddress := make([]byte, AddressBytes)
	WriteAddress(privateAddress, &packet.ServerPrivateAddress)
	binary.Write(buf, binary.LittleEndian, privateAddress)

	binary.Write(buf, binary.LittleEndian, packet.ServerRoutePublicKey)
	return buf.Bytes()
}

type SessionUpdatePacket struct {
	Sequence                  uint64
	CustomerId                uint64
	SessionId                 uint64
	UserHash                  uint64
	PlatformId                uint64
	Tag                       uint64
	Flags                     uint32
	Flagged                   bool
	FallbackToDirect          bool
	TryBeforeYouBuy           bool
	ConnectionType            int32
	OnNetworkNext             bool
	DirectMinRtt              float32
	DirectMaxRtt              float32
	DirectMeanRtt             float32
	DirectJitter              float32
	DirectPacketLoss          float32
	NextMinRtt                float32
	NextMaxRtt                float32
	NextMeanRtt               float32
	NextJitter                float32
	NextPacketLoss            float32
	NumNearRelays             int32
	NearRelayIds              []uint64
	NearRelayMinRtt           []float32
	NearRelayMaxRtt           []float32
	NearRelayMeanRtt          []float32
	NearRelayJitter           []float32
	NearRelayPacketLoss       []float32
	ClientAddress             net.UDPAddr
	ServerAddress             net.UDPAddr
	ClientRoutePublicKey      []byte
	KbpsUp                    uint32
	KbpsDown                  uint32
	PacketsLostClientToServer uint64
	PacketsLostServerToClient uint64
	Signature                 []byte
}

func (packet *SessionUpdatePacket) UnmarshalBinary(data []byte) error {
	if err := packet.Serialize(CreateReadStream(data), SDKVersionMajorMin, SDKVersionMinorMin, SDKVersionPatchMin); err != nil {
		return err
	}
	return nil
}

func (packet *SessionUpdatePacket) MarshalBinary() ([]byte, error) {
	ws, err := CreateWriteStream(1500)
	if err != nil {
		return nil, err
	}

	if err := packet.Serialize(ws, SDKVersionMajorMin, SDKVersionMinorMin, SDKVersionPatchMin); err != nil {
		return nil, err
	}
	ws.Flush()

	return ws.GetData(), nil
}

func (packet *SessionUpdatePacket) Serialize(stream Stream, versionMajor int32, versionMinor int32, versionPatch int32) error {
	stream.SerializeUint64(&packet.Sequence)
	stream.SerializeUint64(&packet.CustomerId)
	stream.SerializeAddress(&packet.ServerAddress)
	stream.SerializeUint64(&packet.SessionId)
	stream.SerializeUint64(&packet.UserHash)
	stream.SerializeUint64(&packet.PlatformId)
	stream.SerializeUint64(&packet.Tag)
	if ProtocolVersionAtLeast(versionMajor, versionMinor, versionPatch, 3, 3, 4) {
		stream.SerializeBits(&packet.Flags, FlagTotalCount)
	}
	stream.SerializeBool(&packet.Flagged)
	stream.SerializeBool(&packet.FallbackToDirect)
	stream.SerializeBool(&packet.TryBeforeYouBuy)
	stream.SerializeInteger(&packet.ConnectionType, ConnectionTypeUnknown, ConnectionTypeCellular)
	stream.SerializeFloat32(&packet.DirectMinRtt)
	stream.SerializeFloat32(&packet.DirectMaxRtt)
	stream.SerializeFloat32(&packet.DirectMeanRtt)
	stream.SerializeFloat32(&packet.DirectJitter)
	stream.SerializeFloat32(&packet.DirectPacketLoss)
	stream.SerializeBool(&packet.OnNetworkNext)
	if packet.OnNetworkNext {
		stream.SerializeFloat32(&packet.NextMinRtt)
		stream.SerializeFloat32(&packet.NextMaxRtt)
		stream.SerializeFloat32(&packet.NextMeanRtt)
		stream.SerializeFloat32(&packet.NextJitter)
		stream.SerializeFloat32(&packet.NextPacketLoss)
	}
	stream.SerializeInteger(&packet.NumNearRelays, 0, NEXT_MAX_NEAR_RELAYS)
	if stream.IsReading() {
		packet.NearRelayIds = make([]uint64, packet.NumNearRelays)
		packet.NearRelayMinRtt = make([]float32, packet.NumNearRelays)
		packet.NearRelayMaxRtt = make([]float32, packet.NumNearRelays)
		packet.NearRelayMeanRtt = make([]float32, packet.NumNearRelays)
		packet.NearRelayJitter = make([]float32, packet.NumNearRelays)
		packet.NearRelayPacketLoss = make([]float32, packet.NumNearRelays)
	}
	var i int32
	for i = 0; i < packet.NumNearRelays; i++ {
		stream.SerializeUint64(&packet.NearRelayIds[i])
		stream.SerializeFloat32(&packet.NearRelayMinRtt[i])
		stream.SerializeFloat32(&packet.NearRelayMaxRtt[i])
		stream.SerializeFloat32(&packet.NearRelayMeanRtt[i])
		stream.SerializeFloat32(&packet.NearRelayJitter[i])
		stream.SerializeFloat32(&packet.NearRelayPacketLoss[i])
	}
	stream.SerializeAddress(&packet.ClientAddress)
	if stream.IsReading() {
		packet.ClientRoutePublicKey = make([]byte, Crypto_box_PUBLICKEYBYTES)
		packet.Signature = make([]byte, SignatureBytes)
	}
	stream.SerializeBytes(packet.ClientRoutePublicKey)
	stream.SerializeUint32(&packet.KbpsUp)
	stream.SerializeUint32(&packet.KbpsDown)
	if ProtocolVersionAtLeast(versionMajor, versionMinor, versionPatch, 3, 3, 2) {
		stream.SerializeUint64(&packet.PacketsLostClientToServer)
		stream.SerializeUint64(&packet.PacketsLostServerToClient)
	}
	stream.SerializeBytes(packet.Signature)
	return stream.Error()
}

func (packet *SessionUpdatePacket) HeaderSerialize(stream Stream) error {
	stream.SerializeUint64(&packet.Sequence)
	stream.SerializeUint64(&packet.CustomerId)
	stream.SerializeAddress(&packet.ServerAddress)
	return stream.Error()
}

func (packet *SessionUpdatePacket) GetSignData(versionMajor int32, versionMinor int32, versionPatch int32) []byte {

	buf := new(bytes.Buffer)

	binary.Write(buf, binary.LittleEndian, packet.Sequence)
	binary.Write(buf, binary.LittleEndian, packet.CustomerId)
	binary.Write(buf, binary.LittleEndian, packet.SessionId)
	binary.Write(buf, binary.LittleEndian, packet.UserHash)
	binary.Write(buf, binary.LittleEndian, packet.PlatformId)
	binary.Write(buf, binary.LittleEndian, packet.Tag)

	if ProtocolVersionAtLeast(versionMajor, versionMinor, versionPatch, 3, 3, 4) {
		binary.Write(buf, binary.LittleEndian, packet.Flags)
	}
	binary.Write(buf, binary.LittleEndian, packet.Flagged)
	binary.Write(buf, binary.LittleEndian, packet.FallbackToDirect)
	binary.Write(buf, binary.LittleEndian, packet.TryBeforeYouBuy)
	binary.Write(buf, binary.LittleEndian, uint8(packet.ConnectionType))

	var onNetworkNext uint8
	onNetworkNext = 0
	if packet.OnNetworkNext {
		onNetworkNext = 1
	}

	binary.Write(buf, binary.LittleEndian, onNetworkNext)

	binary.Write(buf, binary.LittleEndian, packet.DirectMinRtt)
	binary.Write(buf, binary.LittleEndian, packet.DirectMaxRtt)
	binary.Write(buf, binary.LittleEndian, packet.DirectMeanRtt)
	binary.Write(buf, binary.LittleEndian, packet.DirectJitter)
	binary.Write(buf, binary.LittleEndian, packet.DirectPacketLoss)

	binary.Write(buf, binary.LittleEndian, packet.NextMinRtt)
	binary.Write(buf, binary.LittleEndian, packet.NextMaxRtt)
	binary.Write(buf, binary.LittleEndian, packet.NextMeanRtt)
	binary.Write(buf, binary.LittleEndian, packet.NextJitter)
	binary.Write(buf, binary.LittleEndian, packet.NextPacketLoss)

	binary.Write(buf, binary.LittleEndian, uint32(packet.NumNearRelays))
	var i int32
	for i = 0; i < packet.NumNearRelays; i++ {
		binary.Write(buf, binary.LittleEndian, packet.NearRelayIds[i])
		binary.Write(buf, binary.LittleEndian, packet.NearRelayMinRtt[i])
		binary.Write(buf, binary.LittleEndian, packet.NearRelayMaxRtt[i])
		binary.Write(buf, binary.LittleEndian, packet.NearRelayMeanRtt[i])
		binary.Write(buf, binary.LittleEndian, packet.NearRelayJitter[i])
		binary.Write(buf, binary.LittleEndian, packet.NearRelayPacketLoss[i])
	}

	clientAddress := make([]byte, AddressBytes)
	WriteAddress(clientAddress, &packet.ClientAddress)
	binary.Write(buf, binary.LittleEndian, clientAddress)

	serverAddress := make([]byte, AddressBytes)
	WriteAddress(serverAddress, &packet.ServerAddress)
	binary.Write(buf, binary.LittleEndian, serverAddress)

	binary.Write(buf, binary.LittleEndian, packet.KbpsUp)
	binary.Write(buf, binary.LittleEndian, packet.KbpsDown)

	if ProtocolVersionAtLeast(versionMajor, versionMinor, versionPatch, 3, 3, 2) {
		binary.Write(buf, binary.LittleEndian, packet.PacketsLostClientToServer)
		binary.Write(buf, binary.LittleEndian, packet.PacketsLostServerToClient)
	}

	binary.Write(buf, binary.LittleEndian, packet.ClientRoutePublicKey)

	return buf.Bytes()
}

type SessionResponsePacket struct {
	Sequence             uint64
	SessionId            uint64
	NumNearRelays        int32
	NearRelayIds         []uint64
	NearRelayAddresses   []net.UDPAddr
	ResponseType         int32
	Multipath            bool
	NumTokens            int32
	Tokens               []byte
	ServerRoutePublicKey []byte
	Signature            []byte
}

func (packet *SessionResponsePacket) Serialize(stream Stream, versionMajor int32, versionMinor int32, versionPatch int32) error {
	stream.SerializeUint64(&packet.Sequence)
	stream.SerializeUint64(&packet.SessionId)
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
	stream.SerializeInteger(&packet.ResponseType, 0, NEXT_UPDATE_TYPE_CONTINUE)
	if packet.ResponseType != NEXT_UPDATE_TYPE_DIRECT {
		stream.SerializeBool(&packet.Multipath)
		stream.SerializeInteger(&packet.NumTokens, 0, NEXT_MAX_TOKENS)
	}
	if packet.ResponseType == NEXT_UPDATE_TYPE_ROUTE {
		stream.SerializeBytes(packet.Tokens)
	}
	if packet.ResponseType == NEXT_UPDATE_TYPE_CONTINUE {
		stream.SerializeBytes(packet.Tokens)
	}
	if stream.IsReading() {
		packet.ServerRoutePublicKey = make([]byte, Crypto_box_PUBLICKEYBYTES)
		packet.Signature = make([]byte, SignatureBytes)
	}
	stream.SerializeBytes(packet.ServerRoutePublicKey)
	stream.SerializeBytes(packet.Signature)
	return stream.Error()
}

func (packet *SessionResponsePacket) GetSignData(versionMajor int32, versionMinor int32, versionPatch int32) []byte {
	buf := new(bytes.Buffer)
	binary.Write(buf, binary.LittleEndian, packet.Sequence)
	binary.Write(buf, binary.LittleEndian, packet.SessionId)
	binary.Write(buf, binary.LittleEndian, uint8(packet.NumNearRelays))
	var i int32
	for i = 0; i < packet.NumNearRelays; i++ {
		binary.Write(buf, binary.LittleEndian, packet.NearRelayIds[i])
		address := make([]byte, AddressBytes)
		WriteAddress(address, &packet.NearRelayAddresses[i])
		binary.Write(buf, binary.LittleEndian, address)
	}
	binary.Write(buf, binary.LittleEndian, uint8(packet.ResponseType))
	if packet.ResponseType != NEXT_UPDATE_TYPE_DIRECT {
		if packet.Multipath {
			binary.Write(buf, binary.LittleEndian, uint8(1))
		} else {
			binary.Write(buf, binary.LittleEndian, uint8(0))
		}
		binary.Write(buf, binary.LittleEndian, uint8(packet.NumTokens))
	}
	if packet.ResponseType == NEXT_UPDATE_TYPE_ROUTE {
		binary.Write(buf, binary.LittleEndian, packet.Tokens)
	}
	if packet.ResponseType == NEXT_UPDATE_TYPE_CONTINUE {
		binary.Write(buf, binary.LittleEndian, packet.Tokens)
	}
	binary.Write(buf, binary.LittleEndian, packet.ServerRoutePublicKey)
	return buf.Bytes()
}

// =========================================================================

type CostMatrix struct {
	RelayIds         []RelayId
	RelayNames       []string
	RelayAddresses   [][]byte
	RelayPublicKeys  [][]byte
	DatacenterIds    []DatacenterId
	DatacenterNames  []string
	DatacenterRelays map[DatacenterId][]RelayId
	RTT              []int32
}

// IMPORTANT: Bump this version whenever you change the binary format
const CostMatrixVersion = 2

func WriteCostMatrix(buffer []byte, costMatrix *CostMatrix) []byte {

	var index int

	// todo: update this to the new way of reading/writing binary as per-backend.go

	binary.LittleEndian.PutUint32(buffer[index:], CostMatrixVersion)
	index += 4

	numRelays := len(costMatrix.RelayIds)
	binary.LittleEndian.PutUint32(buffer[index:], uint32(numRelays))
	index += 4

	for i := range costMatrix.RelayIds {
		binary.LittleEndian.PutUint32(buffer[index:], uint32(costMatrix.RelayIds[i]))
		index += 4
	}

	for i := range costMatrix.RelayNames {
		index += WriteString(buffer[index:], costMatrix.RelayNames[i])
	}

	if len(costMatrix.DatacenterIds) != len(costMatrix.DatacenterNames) {
		panic("datacenter ids length does not match datacenter names length")
	}

	binary.LittleEndian.PutUint32(buffer[index:], uint32(len(costMatrix.DatacenterIds)))
	index += 4

	for i := 0; i < len(costMatrix.DatacenterIds); i++ {
		binary.LittleEndian.PutUint32(buffer[index:], uint32(costMatrix.DatacenterIds[i]))
		index += 4
		index += WriteString(buffer[index:], costMatrix.DatacenterNames[i])
	}

	for i := range costMatrix.RelayAddresses {
		index += WriteBytes(buffer[index:], costMatrix.RelayAddresses[i])
	}

	for i := range costMatrix.RelayPublicKeys {
		index += WriteBytes(buffer[index:], costMatrix.RelayPublicKeys[i])
	}

	numDatacenters := int32(len(costMatrix.DatacenterRelays))
	binary.LittleEndian.PutUint32(buffer[index:], uint32(numDatacenters))
	index += 4

	for k, v := range costMatrix.DatacenterRelays {

		binary.LittleEndian.PutUint32(buffer[index:], uint32(k))
		index += 4

		numRelaysInDatacenter := len(v)
		binary.LittleEndian.PutUint32(buffer[index:], uint32(numRelaysInDatacenter))
		index += 4

		for i := 0; i < numRelaysInDatacenter; i++ {
			binary.LittleEndian.PutUint32(buffer[index:], uint32(v[i]))
			index += 4
		}
	}

	for i := range costMatrix.RTT {
		binary.LittleEndian.PutUint32(buffer[index:], uint32(costMatrix.RTT[i]))
		index += 4
	}

	return buffer[:index]
}

func ReadCostMatrix(buffer []byte) (*CostMatrix, error) {

	var index int

	var costMatrix CostMatrix

	// todo: update to new way to read/write binary as per backend.go

	version := binary.LittleEndian.Uint32(buffer[index:])
	index += 4

	if version > CostMatrixVersion {
		return nil, fmt.Errorf("unknown cost matrix version %d", version)
	}

	numRelays := int32(binary.LittleEndian.Uint32(buffer[index:]))
	index += 4

	costMatrix.RelayIds = make([]RelayId, numRelays)
	for i := 0; i < int(numRelays); i++ {
		costMatrix.RelayIds[i] = RelayId(binary.LittleEndian.Uint32(buffer[index:]))
		index += 4
	}

	var bytes_read int

	costMatrix.RelayNames = make([]string, numRelays)
	if version >= 1 {
		for i := range costMatrix.RelayNames {
			costMatrix.RelayNames[i], bytes_read = ReadString(buffer[index:])
			index += bytes_read
		}
	}

	if version >= 2 {
		datacenterCount := binary.LittleEndian.Uint32(buffer[index:])
		index += 4

		costMatrix.DatacenterIds = make([]DatacenterId, datacenterCount)
		costMatrix.DatacenterNames = make([]string, datacenterCount)
		for i := 0; i < int(datacenterCount); i++ {
			costMatrix.DatacenterIds[i] = DatacenterId(binary.LittleEndian.Uint32(buffer[index:]))
			index += 4
			costMatrix.DatacenterNames[i], bytes_read = ReadString(buffer[index:])
			index += bytes_read
		}
	}

	costMatrix.RelayAddresses = make([][]byte, numRelays)
	for i := range costMatrix.RelayAddresses {
		costMatrix.RelayAddresses[i], bytes_read = ReadBytes(buffer[index:])
		index += bytes_read
	}

	costMatrix.RelayPublicKeys = make([][]byte, numRelays)
	for i := range costMatrix.RelayPublicKeys {
		costMatrix.RelayPublicKeys[i], bytes_read = ReadBytes(buffer[index:])
		index += bytes_read
	}

	numDatacenters := int32(binary.LittleEndian.Uint32(buffer[index:]))
	index += 4

	costMatrix.DatacenterRelays = make(map[DatacenterId][]RelayId)

	for i := 0; i < int(numDatacenters); i++ {

		datacenterId := DatacenterId(binary.LittleEndian.Uint32(buffer[index:]))
		index += 4

		numRelaysInDatacenter := int32(binary.LittleEndian.Uint32(buffer[index:]))
		index += 4

		costMatrix.DatacenterRelays[datacenterId] = make([]RelayId, numRelaysInDatacenter)

		for j := 0; j < int(numRelaysInDatacenter); j++ {
			costMatrix.DatacenterRelays[datacenterId][j] = RelayId(binary.LittleEndian.Uint32(buffer[index:]))
			index += 4
		}
	}

	entryCount := TriMatrixLength(int(numRelays))
	costMatrix.RTT = make([]int32, entryCount)
	for i := range costMatrix.RTT {
		costMatrix.RTT[i] = int32(binary.LittleEndian.Uint32(buffer[index:]))
		index += 4
	}

	return &costMatrix, nil
}

// =============================================================================

type RouteMatrix struct {
	RelayIds         []RelayId
	RelayNames       []string
	RelayAddresses   [][]byte
	RelayPublicKeys  [][]byte
	DatacenterRelays map[DatacenterId][]RelayId
	DatacenterIds    []DatacenterId
	DatacenterNames  []string
	Entries          []RouteMatrixEntry
}

type RouteMatrixEntry struct {
	DirectRTT      int32
	NumRoutes      int32
	RouteRTT       [MaxRoutesPerRelayPair]int32
	RouteNumRelays [MaxRoutesPerRelayPair]int32
	RouteRelays    [MaxRoutesPerRelayPair][MaxRelays]uint32
}

// IMPORTANT: Increment this when you change the binary format
const RouteMatrixVersion = 2

func WriteRouteMatrix(buffer []byte, routeMatrix *RouteMatrix) []byte {

	var index int

	// todo: update to new way to read/write binary as per backend.go

	binary.LittleEndian.PutUint32(buffer[index:], RouteMatrixVersion)
	index += 4

	numRelays := len(routeMatrix.RelayIds)
	binary.LittleEndian.PutUint32(buffer[index:], uint32(numRelays))
	index += 4

	for i := range routeMatrix.RelayIds {
		binary.LittleEndian.PutUint32(buffer[index:], uint32(routeMatrix.RelayIds[i]))
		index += 4
	}

	for i := range routeMatrix.RelayNames {
		index += WriteString(buffer[index:], routeMatrix.RelayNames[i])
	}

	if len(routeMatrix.DatacenterIds) != len(routeMatrix.DatacenterNames) {
		panic("datacenter ids length does not match datacenter names length")
	}

	binary.LittleEndian.PutUint32(buffer[index:], uint32(len(routeMatrix.DatacenterIds)))
	index += 4

	for i := 0; i < len(routeMatrix.DatacenterIds); i++ {
		binary.LittleEndian.PutUint32(buffer[index:], uint32(routeMatrix.DatacenterIds[i]))
		index += 4
		index += WriteString(buffer[index:], routeMatrix.DatacenterNames[i])
	}

	for i := range routeMatrix.RelayAddresses {
		index += WriteBytes(buffer[index:], routeMatrix.RelayAddresses[i])
	}

	for i := range routeMatrix.RelayPublicKeys {
		index += WriteBytes(buffer[index:], routeMatrix.RelayPublicKeys[i])
	}

	numDatacenters := int32(len(routeMatrix.DatacenterRelays))
	binary.LittleEndian.PutUint32(buffer[index:], uint32(numDatacenters))
	index += 4

	for k, v := range routeMatrix.DatacenterRelays {

		binary.LittleEndian.PutUint32(buffer[index:], uint32(k))
		index += 4

		numRelaysInDatacenter := len(v)
		binary.LittleEndian.PutUint32(buffer[index:], uint32(numRelaysInDatacenter))
		index += 4

		for i := 0; i < numRelaysInDatacenter; i++ {
			binary.LittleEndian.PutUint32(buffer[index:], uint32(v[i]))
			index += 4
		}
	}

	for i := range routeMatrix.Entries {

		binary.LittleEndian.PutUint32(buffer[index:], uint32(routeMatrix.Entries[i].DirectRTT))
		index += 4

		binary.LittleEndian.PutUint32(buffer[index:], uint32(routeMatrix.Entries[i].NumRoutes))
		index += 4

		for j := 0; j < int(routeMatrix.Entries[i].NumRoutes); j++ {

			binary.LittleEndian.PutUint32(buffer[index:], uint32(routeMatrix.Entries[i].RouteRTT[j]))
			index += 4

			binary.LittleEndian.PutUint32(buffer[index:], uint32(routeMatrix.Entries[i].RouteNumRelays[j]))
			index += 4

			for k := 0; k < int(routeMatrix.Entries[i].RouteNumRelays[j]); k++ {
				binary.LittleEndian.PutUint32(buffer[index:], uint32(routeMatrix.Entries[i].RouteRelays[j][k]))
				index += 4
			}
		}
	}

	return buffer[:index]
}

func ReadRouteMatrix(buffer []byte) (*RouteMatrix, error) {

	var index int

	var routeMatrix RouteMatrix

	// todo: update to new and better way to read/write binary

	version := binary.LittleEndian.Uint32(buffer[index:])
	index += 4

	if version > RouteMatrixVersion {
		return nil, fmt.Errorf("unknown route matrix version: %d", version)
	}

	var numRelays int32
	numRelays = int32(binary.LittleEndian.Uint32(buffer[index:]))
	index += 4

	routeMatrix.RelayIds = make([]RelayId, numRelays)
	for i := 0; i < int(numRelays); i++ {
		routeMatrix.RelayIds[i] = RelayId(binary.LittleEndian.Uint32(buffer[index:]))
		index += 4
	}

	var bytes_read int

	routeMatrix.RelayNames = make([]string, numRelays)
	if version >= 1 {
		for i := range routeMatrix.RelayNames {
			routeMatrix.RelayNames[i], bytes_read = ReadString(buffer[index:])
			index += bytes_read
		}
	}

	if version >= 2 {
		datacenterCount := binary.LittleEndian.Uint32(buffer[index:])
		index += 4

		routeMatrix.DatacenterIds = make([]DatacenterId, datacenterCount)
		routeMatrix.DatacenterNames = make([]string, datacenterCount)
		for i := 0; i < int(datacenterCount); i++ {
			routeMatrix.DatacenterIds[i] = DatacenterId(binary.LittleEndian.Uint32(buffer[index:]))
			index += 4
			routeMatrix.DatacenterNames[i], bytes_read = ReadString(buffer[index:])
			index += bytes_read
		}
	}

	routeMatrix.RelayAddresses = make([][]byte, numRelays)
	for i := range routeMatrix.RelayAddresses {
		routeMatrix.RelayAddresses[i], bytes_read = ReadBytes(buffer[index:])
		index += bytes_read
	}

	routeMatrix.RelayPublicKeys = make([][]byte, numRelays)
	for i := range routeMatrix.RelayPublicKeys {
		routeMatrix.RelayPublicKeys[i], bytes_read = ReadBytes(buffer[index:])
		index += bytes_read
	}

	numDatacenters := int32(binary.LittleEndian.Uint32(buffer[index:]))
	index += 4

	routeMatrix.DatacenterRelays = make(map[DatacenterId][]RelayId)

	for i := 0; i < int(numDatacenters); i++ {

		datacenterId := DatacenterId(binary.LittleEndian.Uint32(buffer[index:]))
		index += 4

		numRelaysInDatacenter := int32(binary.LittleEndian.Uint32(buffer[index:]))
		index += 4

		routeMatrix.DatacenterRelays[datacenterId] = make([]RelayId, numRelaysInDatacenter)

		for j := 0; j < int(numRelaysInDatacenter); j++ {
			routeMatrix.DatacenterRelays[datacenterId][j] = RelayId(binary.LittleEndian.Uint32(buffer[index:]))
			index += 4
		}
	}

	entryCount := TriMatrixLength(int(numRelays))

	routeMatrix.Entries = make([]RouteMatrixEntry, entryCount)

	for i := range routeMatrix.Entries {

		routeMatrix.Entries[i].DirectRTT = int32(binary.LittleEndian.Uint32(buffer[index:]))
		index += 4

		routeMatrix.Entries[i].NumRoutes = int32(binary.LittleEndian.Uint32(buffer[index:]))
		index += 4

		for j := 0; j < int(routeMatrix.Entries[i].NumRoutes); j++ {

			routeMatrix.Entries[i].RouteRTT[j] = int32(binary.LittleEndian.Uint32(buffer[index:]))
			index += 4

			routeMatrix.Entries[i].RouteNumRelays[j] = int32(binary.LittleEndian.Uint32(buffer[index:]))
			index += 4

			for k := 0; k < int(routeMatrix.Entries[i].RouteNumRelays[j]); k++ {
				routeMatrix.Entries[i].RouteRelays[j][k] = binary.LittleEndian.Uint32(buffer[index:])
				index += 4
			}
		}
	}

	return &routeMatrix, nil
}

// =============================================================================

func RouteHash(relays ...uint32) uint32 {
	hash := uint32(0)
	for i := range relays {
		hash *= uint32(0x811C9DC5)
		hash ^= (relays[i] >> 24) & 0xFF
		hash *= uint32(0x811C9DC5)
		hash ^= (relays[i] >> 16) & 0xFF
		hash *= uint32(0x811C9DC5)
		hash ^= (relays[i] >> 8) & 0xFF
		hash *= uint32(0x811C9DC5)
		hash ^= relays[i] & 0xFF
	}
	return hash
}

type RouteManager struct {
	NumRoutes      int
	RouteRTT       [MaxRoutesPerRelayPair]int32
	RouteHash      [MaxRoutesPerRelayPair]uint32
	RouteNumRelays [MaxRoutesPerRelayPair]int32
	RouteRelays    [MaxRoutesPerRelayPair][MaxRelays]uint32
}

func NewRouteManager() *RouteManager {
	manager := &RouteManager{}
	return manager
}

func (manager *RouteManager) AddRoute(rtt int32, relays ...uint32) {
	if rtt < 0 {
		return
	}
	if manager.NumRoutes == 0 {

		// no routes yet. add the route

		manager.NumRoutes = 1
		manager.RouteRTT[0] = rtt
		manager.RouteHash[0] = RouteHash(relays...)
		manager.RouteNumRelays[0] = int32(len(relays))
		for i := range relays {
			manager.RouteRelays[0][i] = relays[i]
		}

	} else if manager.NumRoutes < MaxRoutesPerRelayPair {

		// not at max routes yet. insert according RTT sort order

		routeHash := RouteHash(relays...)
		for i := 0; i < manager.NumRoutes; i++ {
			if routeHash == manager.RouteHash[i] {
				return
			}
		}

		if rtt >= manager.RouteRTT[manager.NumRoutes-1] {

			// RTT is greater than existing entries. append.

			manager.RouteRTT[manager.NumRoutes] = rtt
			manager.RouteHash[manager.NumRoutes] = routeHash
			manager.RouteNumRelays[manager.NumRoutes] = int32(len(relays))
			for i := range relays {
				manager.RouteRelays[manager.NumRoutes][i] = relays[i]
			}
			manager.NumRoutes++

		} else {

			// RTT is lower than at least one entry. insert.

			insertIndex := manager.NumRoutes - 1
			for {
				if insertIndex == 0 || rtt > manager.RouteRTT[insertIndex-1] {
					break
				}
				insertIndex--
			}
			manager.NumRoutes++
			for i := manager.NumRoutes - 1; i > insertIndex; i-- {
				manager.RouteRTT[i] = manager.RouteRTT[i-1]
				manager.RouteHash[i] = manager.RouteHash[i-1]
				manager.RouteNumRelays[i] = manager.RouteNumRelays[i-1]
				for j := 0; j < int(manager.RouteNumRelays[i]); j++ {
					manager.RouteRelays[i][j] = manager.RouteRelays[i-1][j]
				}
			}
			manager.RouteRTT[insertIndex] = rtt
			manager.RouteHash[insertIndex] = routeHash
			manager.RouteNumRelays[insertIndex] = int32(len(relays))
			for i := range relays {
				manager.RouteRelays[insertIndex][i] = relays[i]
			}

		}

	} else {

		// route set is full. only insert if lower RTT than at least one current route.

		if rtt >= manager.RouteRTT[manager.NumRoutes-1] {
			return
		}

		routeHash := RouteHash(relays...)
		for i := 0; i < manager.NumRoutes; i++ {
			if routeHash == manager.RouteHash[i] {
				return
			}
		}

		insertIndex := manager.NumRoutes - 1
		for {
			if insertIndex == 0 || rtt > manager.RouteRTT[insertIndex-1] {
				break
			}
			insertIndex--
		}

		for i := manager.NumRoutes - 1; i > insertIndex; i-- {
			manager.RouteRTT[i] = manager.RouteRTT[i-1]
			manager.RouteHash[i] = manager.RouteHash[i-1]
			manager.RouteNumRelays[i] = manager.RouteNumRelays[i-1]
			for j := 0; j < int(manager.RouteNumRelays[i]); j++ {
				manager.RouteRelays[i][j] = manager.RouteRelays[i-1][j]
			}
		}

		manager.RouteRTT[insertIndex] = rtt
		manager.RouteHash[insertIndex] = routeHash
		manager.RouteNumRelays[insertIndex] = int32(len(relays))

		for i := range relays {
			manager.RouteRelays[insertIndex][i] = relays[i]
		}

	}
}

func Optimize(costMatrix *CostMatrix, thresholdRTT int32) *RouteMatrix {

	numRelays := len(costMatrix.RelayIds)

	entryCount := TriMatrixLength(numRelays)

	result := &RouteMatrix{}
	result.RelayIds = costMatrix.RelayIds
	result.RelayNames = costMatrix.RelayNames
	result.RelayAddresses = costMatrix.RelayAddresses
	result.RelayPublicKeys = costMatrix.RelayPublicKeys
	result.DatacenterIds = costMatrix.DatacenterIds
	result.DatacenterNames = costMatrix.DatacenterNames
	result.DatacenterRelays = costMatrix.DatacenterRelays
	result.Entries = make([]RouteMatrixEntry, entryCount)

	type Indirect struct {
		relay int32
		rtt   int32
	}

	rtt := costMatrix.RTT

	indirect := make([][][]Indirect, numRelays)

	// phase 1: build a matrix of indirect routes from relays i -> j that have lower rtt than direct, eg. i -> (x) -> j, where x is every other relay

	numCPUs := runtime.NumCPU()

	numSegments := numRelays
	if numCPUs < numRelays {
		numSegments = numRelays / 5
		if numSegments == 0 {
			numSegments = 1
		}
	}

	var wg sync.WaitGroup

	wg.Add(numSegments)

	for segment := 0; segment < numSegments; segment++ {

		startIndex := segment * numRelays / numSegments
		endIndex := (segment+1)*numRelays/numSegments - 1
		if segment == numSegments-1 {
			endIndex = numRelays - 1
		}

		go func(startIndex int, endIndex int) {

			defer wg.Done()

			working := make([]Indirect, numRelays)

			for i := startIndex; i <= endIndex; i++ {

				indirect[i] = make([][]Indirect, numRelays)

				for j := 0; j < numRelays; j++ {

					// can't route to self
					if i == j {
						continue
					}

					ij_index := TriMatrixIndex(i, j)

					numRoutes := 0
					rtt_direct := rtt[ij_index]

					if rtt_direct < 0 {

						// no direct route exists between i,j. subdivide valid routes so we don't miss indirect paths.

						for k := 0; k < numRelays; k++ {
							if k == i || k == j {
								continue
							}
							ik_index := TriMatrixIndex(i, k)
							kj_index := TriMatrixIndex(k, j)
							ik_rtt := rtt[ik_index]
							kj_rtt := rtt[kj_index]
							if ik_rtt < 0 || kj_rtt < 0 {
								continue
							}
							working[numRoutes].relay = int32(k)
							working[numRoutes].rtt = ik_rtt + kj_rtt
							numRoutes++
						}

					} else {

						// direct route exists between i,j. subdivide only when a significant rtt reduction occurs.

						for k := 0; k < numRelays; k++ {
							if k == i || k == j {
								continue
							}
							ik_index := TriMatrixIndex(i, k)
							kj_index := TriMatrixIndex(k, j)
							ik_rtt := rtt[ik_index]
							kj_rtt := rtt[kj_index]
							if ik_rtt < 0 || kj_rtt < 0 {
								continue
							}
							indirectRTT := ik_rtt + kj_rtt
							if indirectRTT > rtt_direct-thresholdRTT {
								continue
							}
							working[numRoutes].relay = int32(k)
							working[numRoutes].rtt = indirectRTT
							numRoutes++
						}

					}

					if numRoutes > 0 {
						indirect[i][j] = make([]Indirect, numRoutes)
						copy(indirect[i][j], working)
						sort.Slice(indirect[i][j], func(a, b int) bool { return indirect[i][j][a].rtt < indirect[i][j][b].rtt })
					}
				}
			}

		}(startIndex, endIndex)
	}

	wg.Wait()

	// phase 2: use the indirect matrix to subdivide a route up to 5 hops

	wg.Add(numSegments)

	for segment := 0; segment < numSegments; segment++ {

		startIndex := segment * numRelays / numSegments
		endIndex := (segment+1)*numRelays/numSegments - 1
		if segment == numSegments-1 {
			endIndex = numRelays - 1
		}

		go func(startIndex int, endIndex int) {

			defer wg.Done()

			for i := startIndex; i <= endIndex; i++ {

				for j := 0; j < i; j++ {

					ij_index := TriMatrixIndex(i, j)

					if indirect[i][j] == nil {

						if rtt[ij_index] >= 0 {

							// only direct route from i -> j exists, and it is suitable

							result.Entries[ij_index].DirectRTT = rtt[ij_index]
							result.Entries[ij_index].NumRoutes = 1
							result.Entries[ij_index].RouteRTT[0] = rtt[ij_index]
							result.Entries[ij_index].RouteNumRelays[0] = 2
							result.Entries[ij_index].RouteRelays[0][0] = uint32(i)
							result.Entries[ij_index].RouteRelays[0][1] = uint32(j)

						}

					} else {

						// subdivide routes from i -> j as follows: i -> (x) -> (y) -> (z) -> j, where the subdivision improves significantly on RTT

						routeManager := NewRouteManager()

						for k := range indirect[i][j] {

							routeManager.AddRoute(rtt[ij_index], uint32(i), uint32(j))

							y := indirect[i][j][k]

							routeManager.AddRoute(y.rtt, uint32(i), uint32(y.relay), uint32(j))

							var x *Indirect
							if indirect[i][y.relay] != nil {
								x = &indirect[i][y.relay][0]
							}

							var z *Indirect
							if indirect[j][y.relay] != nil {
								z = &indirect[j][y.relay][0]
							}

							if x != nil {
								ix_index := TriMatrixIndex(i, int(x.relay))
								xy_index := TriMatrixIndex(int(x.relay), int(y.relay))
								yj_index := TriMatrixIndex(int(y.relay), j)

								routeManager.AddRoute(rtt[ix_index]+rtt[xy_index]+rtt[yj_index],
									uint32(i), uint32(x.relay), uint32(y.relay), uint32(j))
							}

							if z != nil {
								iy_index := TriMatrixIndex(i, int(y.relay))
								yz_index := TriMatrixIndex(int(y.relay), int(z.relay))
								zj_index := TriMatrixIndex(int(z.relay), j)

								routeManager.AddRoute(rtt[iy_index]+rtt[yz_index]+rtt[zj_index],
									uint32(i), uint32(y.relay), uint32(z.relay), uint32(j))
							}

							if x != nil && z != nil {
								ix_index := TriMatrixIndex(i, int(x.relay))
								xy_index := TriMatrixIndex(int(x.relay), int(y.relay))
								yz_index := TriMatrixIndex(int(y.relay), int(z.relay))
								zj_index := TriMatrixIndex(int(z.relay), j)

								routeManager.AddRoute(rtt[ix_index]+rtt[xy_index]+rtt[yz_index]+rtt[zj_index],
									uint32(i), uint32(x.relay), uint32(y.relay), uint32(z.relay), uint32(j))
							}

							numRoutes := routeManager.NumRoutes

							result.Entries[ij_index].DirectRTT = rtt[ij_index]
							result.Entries[ij_index].NumRoutes = int32(numRoutes)

							for u := 0; u < numRoutes; u++ {
								result.Entries[ij_index].RouteRTT[u] = routeManager.RouteRTT[u]
								result.Entries[ij_index].RouteNumRelays[u] = routeManager.RouteNumRelays[u]
								numRelays := int(result.Entries[ij_index].RouteNumRelays[u])
								for v := 0; v < numRelays; v++ {
									result.Entries[ij_index].RouteRelays[u][v] = routeManager.RouteRelays[u][v]
								}
							}
						}
					}
				}
			}

		}(startIndex, endIndex)
	}

	wg.Wait()

	return result
}
