/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

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
	"math/bits"
	"net"
	"runtime"
	"runtime/debug"
	"sort"
	"strconv"
	"strings"
	"sync"
)

const MaxRelays = 5
const MaxNodes = 7
const MaxNearRelays = 10
const BillingSliceSeconds = 10
const MinimumKbps = 100
const AddressBytes = 19
const SessionTokenBytes = 77
const EncryptedSessionTokenBytes = 117
const ContinueTokenBytes = 18
const EncryptedContinueTokenBytes = 58
const MaxRoutesPerRelayPair = 8

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
)

const (
	SDKVersionMajorMin = 3
	SDKVersionMinorMin = 3
	SDKVersionPatchMin = 2
	SDKVersionMajorMax = 254
	SDKVersionMinorMax = 1023
	SDKVersionPatchMax = 254
)

var RouterPrivateKey = [...]byte{0x96, 0xce, 0x57, 0x8b, 0x00, 0x19, 0x44, 0x27, 0xf2, 0xb9, 0x90, 0x1b, 0x43, 0x56, 0xfd, 0x4f, 0x56, 0xe1, 0xd9, 0x56, 0x58, 0xf2, 0xf4, 0x3b, 0x86, 0x9f, 0x12, 0x75, 0x24, 0xd2, 0x47, 0xb3}

var BackendPrivateKey = []byte{21, 124, 5, 171, 56, 198, 148, 140, 20, 15, 8, 170, 212, 222, 84, 155, 149, 84, 122, 199, 107, 225, 243, 246, 133, 85, 118, 114, 114, 126, 200, 4, 76, 97, 202, 140, 71, 135, 62, 212, 160, 181, 151, 195, 202, 224, 207, 113, 8, 45, 37, 60, 145, 14, 212, 111, 25, 34, 175, 186, 37, 150, 163, 64}

// ============================================================================

var MaxJitter float32 = 10.0
var MaxPacketLoss float32 = 0.1

const InvalidHistoryValue = -1
const RelayTimeoutSeconds = 60

// ============================================================================

type RouteContext struct {
	RelayAddresses  []*net.UDPAddr
	RelayPublicKeys [][]byte
	RouteMatrix     *RouteMatrix
	RelayIdToIndex  map[RelayId]int
}

type Route struct {
	RTT        float32
	Jitter     float32
	PacketLoss float32
	RelayIds   []RelayId
}

func GetRouteHash(relayIds []RelayId) uint64 {
	hash := fnv.New64a()
	for _, v := range relayIds {
		a := make([]byte, 4)
		binary.LittleEndian.PutUint32(a, uint32(v))
		hash.Write(a)
	}
	return hash.Sum64()
}

type RouteStats struct {
	RTT        float32
	Jitter     float32
	PacketLoss float32
}

type RelayStats struct {
	Id RelayId
	RouteStats
}

type RouteSample struct {
	DirectStats RouteStats
	NextStats   RouteStats
	NearRelays  []RelayStats
}

const ROUTE_SLICE_FLAG_NEXT = (uint64(1) << 1)
const ROUTE_SLICE_FLAG_REPORTED = (uint64(1) << 2)
const ROUTE_SLICE_FLAG_VETOED = (uint64(1) << 3)
const ROUTE_SLICE_FLAG_FALLBACK_TO_DIRECT = (uint64(1) << 4)
const ROUTE_SLICE_FLAG_PACKET_LOSS_MULTIPATH = (uint64(1) << 5)
const ROUTE_SLICE_FLAG_JITTER_MULTIPATH = (uint64(1) << 6)
const ROUTE_SLICE_FLAG_RTT_MULTIPATH = (uint64(1) << 7)

type RouteSlice struct {
	Flags          uint64
	RouteSample    RouteSample
	PredictedRoute *Route
}

const RouteSliceVersion = uint32(0)

const RouteSliceMagic = uint32(0x12345678)

func (slice *RouteSlice) Serialize(stream Stream) error {

	var magic uint32
	if stream.IsWriting() {
		magic = RouteSliceMagic
	}
	stream.SerializeUint32(&magic)
	if stream.IsReading() && magic != RouteSliceMagic {
		return fmt.Errorf("expected route slice magic %x, got %x", RouteSliceMagic, magic)
	}

	var version uint32
	if stream.IsWriting() {
		version = RouteSliceVersion
	}
	stream.SerializeUint32(&version)
	if stream.IsReading() && version != RouteSliceVersion {
		return fmt.Errorf("expected route slice version %d, got %d", RouteSliceVersion, version)
	}

	stream.SerializeUint64(&slice.Flags)

	stream.SerializeFloat32(&slice.RouteSample.DirectStats.RTT)
	stream.SerializeFloat32(&slice.RouteSample.DirectStats.Jitter)
	stream.SerializeFloat32(&slice.RouteSample.DirectStats.PacketLoss)

	stream.SerializeFloat32(&slice.RouteSample.NextStats.RTT)
	stream.SerializeFloat32(&slice.RouteSample.NextStats.Jitter)
	stream.SerializeFloat32(&slice.RouteSample.NextStats.PacketLoss)

	hasPredictedRoute := stream.IsWriting() && slice.PredictedRoute != nil
	stream.SerializeBool(&hasPredictedRoute)
	if hasPredictedRoute {
		if stream.IsReading() {
			slice.PredictedRoute = &Route{}
		}
		stream.SerializeFloat32(&slice.PredictedRoute.RTT)
		stream.SerializeFloat32(&slice.PredictedRoute.Jitter)
		stream.SerializeFloat32(&slice.PredictedRoute.PacketLoss)
		numRelayIds := uint32(len(slice.PredictedRoute.RelayIds))
		stream.SerializeUint32(&numRelayIds)
		if stream.IsReading() {
			if numRelayIds > MaxRelays {
				return fmt.Errorf("too many relays in route: %d", numRelayIds)
			}
			slice.PredictedRoute.RelayIds = make([]RelayId, numRelayIds)
		}
		for i := range slice.PredictedRoute.RelayIds {
			relayId := uint64(slice.PredictedRoute.RelayIds[i])
			stream.SerializeUint64(&relayId)
			if stream.IsReading() {
				slice.PredictedRoute.RelayIds[i] = RelayId(relayId)
			}
		}

		numNearRelays := uint32(len(slice.RouteSample.NearRelays))
		stream.SerializeUint32(&numNearRelays)
		if stream.IsReading() {
			if numNearRelays > MaxNearRelays {
				return fmt.Errorf("too many near relays in route slice: %d", numNearRelays)
			}
			slice.RouteSample.NearRelays = make([]RelayStats, numNearRelays)
		}
		for i := 0; i < int(numNearRelays); i++ {
			relayId := uint64(slice.RouteSample.NearRelays[i].Id)
			stream.SerializeUint64(&relayId)
			if stream.IsReading() {
				slice.RouteSample.NearRelays[i].Id = RelayId(relayId)
			}
			stream.SerializeFloat32(&slice.RouteSample.NearRelays[i].RTT)
			stream.SerializeFloat32(&slice.RouteSample.NearRelays[i].Jitter)
			stream.SerializeFloat32(&slice.RouteSample.NearRelays[i].PacketLoss)
		}
	}

	return stream.Error()
}

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

func GetRelayIdOld(id *EntityId) (RelayId, error) {
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

// ===========================================================

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

func Log2(x uint32) int {
	a := x | (x >> 1)
	b := a | (a >> 2)
	c := b | (b >> 4)
	d := c | (c >> 8)
	e := d | (d >> 16)
	f := e >> 1
	return bits.OnesCount32(f)
}

func BitsRequired(min uint32, max uint32) int {
	if min == max {
		return 0
	} else {
		return Log2(max-min) + 1
	}
}

func BitsRequiredSigned(min int32, max int32) int {
	if min == max {
		return 0
	} else {
		return Log2(uint32(max-min)) + 1
	}
}

func SequenceGreaterThan(s1 uint16, s2 uint16) bool {
	return ((s1 > s2) && (s1-s2 <= 32768)) ||
		((s1 < s2) && (s2-s1 > 32768))
}

func SequenceLessThan(s1 uint16, s2 uint16) bool {
	return SequenceGreaterThan(s2, s1)
}

func SignedToUnsigned(n int32) uint32 {
	return uint32((n << 1) ^ (n >> 31))
}

func UnsignedToSigned(n uint32) int32 {
	return int32(n>>1) ^ (-int32(n & 1))
}

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

func HostToNetwork(x uint32) uint32 {
	return x
}

func NetworkToHost(x uint32) uint32 {
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
		writer.data[writer.wordIndex] = HostToNetwork(uint32(writer.scratch & 0xFFFFFFFF))
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
		writer.data[writer.wordIndex] = HostToNetwork(uint32(writer.scratch & 0xFFFFFFFF))
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
		reader.scratch |= uint64(NetworkToHost(reader.data[reader.wordIndex])) << uint(reader.scratchBits)
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

// ====================================================================

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
	SerializeIntRelative(previous *int32, current *int32)
	SerializeAckRelative(sequence uint16, ack *uint16)
	SerializeSequenceRelative(sequence1 uint16, sequence2 *uint16)
	SerializeAlign()
	SerializeAddress(addr *net.UDPAddr)
	GetAlignBits() int
	GetBytesProcessed() int
	GetBitsProcessed() int
	Error() error
	Flush()
}

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
	bits := BitsRequired(uint32(min), uint32(max))
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

func (stream *WriteStream) SerializeAckRelative(sequence uint16, ack *uint16) {
	if ack == nil {
		stream.error(fmt.Errorf("ack is nil"))
		return
	}

	ackDelta := int32(0)
	if *ack < sequence {
		ackDelta = int32(sequence) - int32(*ack)
	} else {
		ackDelta = int32(sequence) + 65536 - int32(*ack)
	}

	if ackDelta < 0 {
		panic("ackDelta should never be < 0")
	}

	if ackDelta == 0 {
		stream.error(fmt.Errorf("ack should not equal sequence"))
		return
	}

	ackInRange := ackDelta <= 64
	stream.SerializeBool(&ackInRange)
	if stream.err != nil {
		return
	}
	if ackInRange {
		stream.SerializeInteger(&ackDelta, 1, 64)
	} else {
		uint32Value := uint32(*ack)
		stream.SerializeBits(&uint32Value, 16)
	}
}

func (stream *WriteStream) SerializeSequenceRelative(sequence1 uint16, sequence2 *uint16) {
	if stream.err != nil {
		return
	}
	if sequence2 == nil {
		stream.error(fmt.Errorf("sequence2 is nil"))
		return
	}
	a := int32(sequence1)
	b := int32(*sequence2)
	if sequence1 > *sequence2 {
		b += 65536
	}
	stream.SerializeIntRelative(&a, &b)
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
	bits := BitsRequiredSigned(min, max)
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

func (stream *ReadStream) SerializeAckRelative(sequence uint16, ack *uint16) {
	if ack == nil {
		stream.error(fmt.Errorf("ack is nil"))
		return
	}
	ackDelta := int32(0)
	ackInRange := false
	stream.SerializeBool(&ackInRange)
	if ackInRange {
		stream.SerializeInteger(&ackDelta, 1, 64)
		*ack = sequence - uint16(ackDelta)
	} else {
		uint32Value := uint32(0)
		stream.SerializeBits(&uint32Value, 16)
		*ack = uint16(uint32Value)
	}
}

func (stream *ReadStream) SerializeSequenceRelative(sequence1 uint16, sequence2 *uint16) {
	if stream.err != nil {
		return
	}
	if sequence2 == nil {
		stream.error(fmt.Errorf("sequence2 is nil"))
		return
	}
	a := int32(sequence1)
	b := int32(0)
	stream.SerializeIntRelative(&a, &b)
	if stream.err != nil {
		return
	}
	if b >= 65536 {
		b -= 65536
	}
	*sequence2 = uint16(b)
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

// ===========================================================

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

const NEXT_FLAGS_BAD_ROUTE_TOKEN = uint32(1 << 0)
const NEXT_FLAGS_NO_ROUTE_TO_CONTINUE = uint32(1 << 1)
const NEXT_FLAGS_PREVIOUS_UPDATE_STILL_PENDING = uint32(1 << 2)
const NEXT_FLAGS_BAD_CONTINUE_TOKEN = uint32(1 << 3)
const NEXT_FLAGS_ROUTE_EXPIRED = uint32(1 << 4)
const NEXT_FLAGS_ROUTE_REQUEST_TIMED_OUT = uint32(1 << 5)
const NEXT_FLAGS_CONTINUE_REQUEST_TIMED_OUT = uint32(1 << 6)
const NEXT_FLAGS_CLIENT_TIMED_OUT = uint32(1 << 7)
const NEXT_FLAGS_TRY_BEFORE_YOU_BUY_ABORT = uint32(1 << 8)
const NEXT_FLAGS_DIRECT_ROUTE_EXPIRED = uint32(1 << 9)
const NEXT_FLAGS_COUNT = 10

func (packet *SessionUpdatePacket) Serialize(stream Stream, versionMajor int32, versionMinor int32, versionPatch int32) error {
	stream.SerializeUint64(&packet.Sequence)
	stream.SerializeUint64(&packet.CustomerId)
	stream.SerializeAddress(&packet.ServerAddress)
	stream.SerializeUint64(&packet.SessionId)
	stream.SerializeUint64(&packet.UserHash)
	stream.SerializeUint64(&packet.PlatformId)
	stream.SerializeUint64(&packet.Tag)
	if ProtocolVersionAtLeast(versionMajor, versionMinor, versionPatch, 3, 3, 4) {
		stream.SerializeBits(&packet.Flags, NEXT_FLAGS_COUNT)
	}
	stream.SerializeBool(&packet.Flagged)
	stream.SerializeBool(&packet.FallbackToDirect)
	stream.SerializeBool(&packet.TryBeforeYouBuy)
	stream.SerializeInteger(&packet.ConnectionType, NEXT_CONNECTION_TYPE_UNKNOWN, NEXT_CONNECTION_TYPE_CELLULAR)
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
