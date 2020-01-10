// Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.

package core

import (
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
	"strconv"
	"strings"
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

// todo: this whole old entity id / relay id etc. is incompatible with how the new backend should work.
// talk to me to learn more. -- glenn

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

// =============================================================================
