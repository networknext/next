/*
   Network Next. You control the network.
   Copyright © 2017 - 2020 Network Next, Inc. All rights reserved.
*/

package core

import (
	"fmt"
	"hash/fnv"
	"io/ioutil"
	"net"
	"reflect"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func StopBeingAnnoyingGolang() {
	fmt.Printf("you are annoying\n")
}

func CompareContinueTokens(a *ContinueToken, b *ContinueToken, t *testing.T) {

	if a.expireTimestamp != b.expireTimestamp {
		t.Errorf("expire timestamp mismatch: got %x, expected %x", a.expireTimestamp, b.expireTimestamp)
	}

	if a.sessionId != b.sessionId {
		t.Errorf("session id mismatch: got %x, expected %x", a.sessionId, b.sessionId)
	}

	if a.sessionVersion != b.sessionVersion {
		t.Errorf("session version mismatch: got %d, expected %d", a.sessionVersion, b.sessionVersion)
	}

	if a.sessionFlags != b.sessionFlags {
		t.Errorf("session flags mismatch: got %d, expected %d", a.sessionFlags, b.sessionFlags)
	}
}

func TestContinueToken(t *testing.T) {

	t.Parallel()

	relayPublicKey := [...]byte{0x71, 0x16, 0xce, 0xc5, 0x16, 0x1a, 0xda, 0xc7, 0xa5, 0x89, 0xb2, 0x51, 0x2b, 0x67, 0x4f, 0x8f, 0x98, 0x21, 0xad, 0x8f, 0xe6, 0x2d, 0x39, 0xca, 0xe3, 0x9b, 0xec, 0xdf, 0x3e, 0xfc, 0x2c, 0x24}
	relayPrivateKey := [...]byte{0xb6, 0x7d, 0x01, 0x0d, 0xaf, 0xba, 0xd1, 0x40, 0x75, 0x99, 0x08, 0x15, 0x0d, 0x3a, 0xce, 0x7b, 0x82, 0x28, 0x01, 0x5f, 0x7d, 0xa0, 0x75, 0xb6, 0xc1, 0x15, 0x56, 0x33, 0xe1, 0x01, 0x99, 0xd6}
	masterPublicKey := [...]byte{0x6f, 0x58, 0xb4, 0xd7, 0x3d, 0xdc, 0x73, 0x06, 0xb8, 0x97, 0x3d, 0x22, 0x4d, 0xe6, 0xf1, 0xfd, 0x2a, 0xf0, 0x26, 0x7e, 0x8b, 0x1d, 0x93, 0x73, 0xd0, 0x40, 0xa9, 0x8b, 0x86, 0x75, 0xcd, 0x43}
	masterPrivateKey := [...]byte{0x2a, 0xad, 0xd5, 0x43, 0x4e, 0x52, 0xbf, 0x62, 0x0b, 0x76, 0x24, 0x18, 0xe1, 0x26, 0xfb, 0xda, 0x89, 0x95, 0x32, 0xde, 0x1d, 0x39, 0x7f, 0xcd, 0x7b, 0x7a, 0xd5, 0x96, 0x3b, 0x0d, 0x23, 0xe5}

	continueToken := &ContinueToken{}
	continueToken.expireTimestamp = uint64(time.Now().Unix() + 10)
	continueToken.sessionId = 0x123131231313131
	continueToken.sessionVersion = 100
	continueToken.sessionFlags = 1

	buffer := make([]byte, EncryptedContinueTokenBytes)

	WriteContinueToken(continueToken, buffer[:])

	readContinueToken, err := ReadContinueToken(buffer)
	if err != nil {
		t.Errorf("ReadContinueToken failed: %s", err)
	}

	CompareContinueTokens(readContinueToken, continueToken, t)

	err = WriteEncryptedContinueToken(buffer, continueToken, masterPrivateKey[:], relayPublicKey[:])
	if err != nil {
		t.Errorf("WriteEncryptedContinueToken failed: %s", err)
	}

	readContinueToken, err = ReadEncryptedContinueToken(buffer, masterPublicKey[:], relayPrivateKey[:])
	if err != nil {
		t.Errorf("ReadEncryptedContinueToken failed: %s", err)
	}

	CompareContinueTokens(readContinueToken, continueToken, t)
}

func CentsToNibblins(cents int64) int64 {
	return cents * 1e9
}

func NibblinsToCents(nibblins int64) int64 {
	return nibblins / 1e9
}

func NibblinsToDollars(nibblins int64) float64 {
	return float64(nibblins) / 1e11
}

func TestDollarString(t *testing.T) {

	t.Parallel()

	x, err := DollarStringToNibblins("0.08372111111") // maximum precision supported by nibblins
	if err != nil {
		t.Errorf("max precision dollar string failed to parse: %v", err)
	}
	if x != 8372111111 {
		t.Errorf("max precision dollar string parsed incorrectly: %v", x)
	}

	x, err = DollarStringToNibblins("0.0837211111111111111") // over maximum precision
	if err != nil {
		t.Errorf("over max precision dollar string failed to parse: %v", err)
	}
	if x != 8372111111 {
		t.Errorf("over max precision dollar string parsed incorrectly: %v", x)
	}

	x, err = DollarStringToNibblins("0.083721111111111111d") // should cause error
	if err == nil {
		t.Errorf("invalid dollar string parsed to: %v", x)
	}

	x, err = DollarStringToNibblins("0.01")
	if err != nil {
		t.Errorf("%v", err)
	}
	if x != CentsToNibblins(1) {
		t.Errorf("dollar string with no leading zero parsed incorrectly: %v", x)
	}

	x, err = DollarStringToNibblins(".01")
	if err != nil {
		t.Errorf("%v", err)
	}
	if x != CentsToNibblins(1) {
		t.Errorf("dollar string with no leading zero parsed incorrectly: %v", x)
	}

	x, err = DollarStringToNibblins("90000000") // max value
	if err != nil {
		t.Errorf("%v", err)
	}
	if NibblinsToDollars(x) != 90000000 {
		t.Errorf("max value dollar string parsed incorrectly: %v", x)
	}

	x, err = DollarStringToNibblins("-90000000") // min value
	if err != nil {
		t.Errorf("%v", err)
	}
	if NibblinsToDollars(x) != -90000000 {
		t.Errorf("min value dollar string parsed incorrectly: %v", x)
	}

	x, err = DollarStringToNibblins("-.01") // negative decimal value
	if err != nil {
		t.Errorf("%v", err)
	}
	if x != CentsToNibblins(-1) {
		t.Errorf("negative decimal dollar string parsed incorrectly: %v", x)
	}

	x, err = DollarStringToNibblins("-1.01") // negative value
	if err != nil {
		t.Errorf("%v", err)
	}
	if x != CentsToNibblins(-101) {
		t.Errorf("negative dollar string parsed incorrectly: %v", x)
	}
}

func TestBitpacker(t *testing.T) {

	t.Parallel()

	const BufferSize = 256

	writer, err := CreateBitWriter(BufferSize)
	assert.Nil(t, err)

	assert.Equal(t, 0, writer.GetBitsWritten())
	assert.Equal(t, 0, writer.GetBytesWritten())
	assert.Equal(t, BufferSize*8, writer.GetBitsAvailable())

	writer.WriteBits(0, 1)
	writer.WriteBits(1, 1)
	writer.WriteBits(10, 8)
	writer.WriteBits(255, 8)
	writer.WriteBits(1000, 10)
	writer.WriteBits(50000, 16)
	writer.WriteBits(9999999, 32)
	writer.FlushBits()

	bitsWritten := 1 + 1 + 8 + 8 + 10 + 16 + 32

	assert.Equal(t, 10, writer.GetBytesWritten())
	assert.Equal(t, bitsWritten, writer.GetBitsWritten())
	assert.Equal(t, BufferSize*8-bitsWritten, writer.GetBitsAvailable())

	bytesWritten := writer.GetBytesWritten()

	assert.Equal(t, 10, bytesWritten)

	buffer := writer.GetData()

	reader := CreateBitReader(buffer[:bytesWritten])

	assert.Equal(t, 0, reader.GetBitsRead())
	assert.Equal(t, bytesWritten*8, reader.GetBitsRemaining())

	a, err := reader.ReadBits(1)
	assert.Nil(t, err)

	b, err := reader.ReadBits(1)
	assert.Nil(t, err)

	c, err := reader.ReadBits(8)
	assert.Nil(t, err)

	d, err := reader.ReadBits(8)
	assert.Nil(t, err)

	e, err := reader.ReadBits(10)
	assert.Nil(t, err)

	f, err := reader.ReadBits(16)
	assert.Nil(t, err)

	g, err := reader.ReadBits(32)
	assert.Nil(t, err)

	assert.Equal(t, uint32(0), a)
	assert.Equal(t, uint32(1), b)
	assert.Equal(t, uint32(10), c)
	assert.Equal(t, uint32(255), d)
	assert.Equal(t, uint32(1000), e)
	assert.Equal(t, uint32(50000), f)
	assert.Equal(t, uint32(9999999), g)

	assert.Equal(t, bitsWritten, reader.GetBitsRead())
	assert.Equal(t, bytesWritten*8-bitsWritten, reader.GetBitsRemaining())
}

func TestBitsRequired(t *testing.T) {
	t.Parallel()
	assert.Equal(t, 0, BitsRequired(0, 0))
	assert.Equal(t, 1, BitsRequired(0, 1))
	assert.Equal(t, 2, BitsRequired(0, 2))
	assert.Equal(t, 2, BitsRequired(0, 3))
	assert.Equal(t, 3, BitsRequired(0, 4))
	assert.Equal(t, 3, BitsRequired(0, 5))
	assert.Equal(t, 3, BitsRequired(0, 6))
	assert.Equal(t, 3, BitsRequired(0, 7))
	assert.Equal(t, 4, BitsRequired(0, 8))
	assert.Equal(t, 8, BitsRequired(0, 255))
	assert.Equal(t, 16, BitsRequired(0, 65535))
	assert.Equal(t, 32, BitsRequired(0, 4294967295))
}

const maxItems = 11
const testByteCount = 17

type testObject struct {
	a           int32
	b           int32
	c           int32
	d           uint32
	e           uint32
	f           uint32
	g           bool
	items       []uint32
	floatValue  float32
	doubleValue float64
	uint64Value uint64
	bytes       []byte
	str         string
	addressA    net.UDPAddr
	addressB    net.UDPAddr
	addressC    net.UDPAddr
}

func createTestObject() *testObject {
	items := make([]uint32, maxItems/2)
	for i := range items {
		items[i] = uint32(i + 10)
	}

	bytes := make([]byte, testByteCount)
	for i := range bytes {
		bytes[i] = byte((i * 37) % 255)
	}

	return &testObject{
		a:           1,
		b:           -2,
		c:           150,
		d:           55,
		e:           255,
		f:           127,
		g:           true,
		items:       items,
		floatValue:  3.1415926,
		doubleValue: 1.0 / 3.0,
		uint64Value: 0x1234567898765432,
		bytes:       bytes,
		str:         "hello world!",
		addressA:    net.UDPAddr{},
		addressB:    *ParseAddress("127.0.0.1:50000"),
		addressC:    *ParseAddress("[::1]:50000"),
	}
}

func (obj *testObject) Serialize(stream Stream) error {

	stream.SerializeInteger(&obj.a, -10, 10)
	stream.SerializeInteger(&obj.b, -10, 10)

	stream.SerializeInteger(&obj.c, -100, 10000)

	stream.SerializeBits(&obj.d, 6)
	stream.SerializeBits(&obj.e, 8)
	stream.SerializeBits(&obj.f, 7)

	stream.SerializeAlign()

	stream.SerializeBool(&obj.g)

	numItems := int32(len(obj.items))
	stream.SerializeInteger(&numItems, 0, maxItems-1)
	if stream.IsReading() {
		obj.items = make([]uint32, numItems)
	}
	for i := range obj.items {
		stream.SerializeBits(&obj.items[i], 8)
	}

	stream.SerializeFloat32(&obj.floatValue)

	stream.SerializeFloat64(&obj.doubleValue)

	stream.SerializeUint64(&obj.uint64Value)

	stream.SerializeBytes(obj.bytes)

	stream.SerializeString(&obj.str, 256)

	stream.SerializeAddress(&obj.addressA)
	stream.SerializeAddress(&obj.addressB)
	stream.SerializeAddress(&obj.addressC)

	stream.Flush()

	return stream.Error()
}

func TestStream(t *testing.T) {

	t.Parallel()

	const BufferSize = 1024

	writeStream, err := CreateWriteStream(BufferSize)
	assert.Nil(t, err)

	writeObject := createTestObject()
	err = writeObject.Serialize(writeStream)
	assert.Nil(t, err)

	bytesWritten := writeStream.GetBytesProcessed()

	buffer := writeStream.GetData()

	readStream := CreateReadStream(buffer[:bytesWritten])
	readObject := &testObject{
		bytes: make([]byte, testByteCount),
	}
	err = readObject.Serialize(readStream)
	assert.Nil(t, err)

	assert.True(t, reflect.DeepEqual(readObject, writeObject))
}

func TestStreamCpp(t *testing.T) {

	t.Parallel()

	// this byte buffer is copy-pasted from C++ next_test()
	buffer := []byte{
		0x0b, 0xe9, 0x03, 0xf7, 0xff, 0x1f, 0x4b, 0x61, 0x81, 0xa1,
		0xc1, 0x41, 0xfb, 0x21, 0x09, 0xa8, 0xaa, 0xaa, 0xaa, 0xaa,
		0xaa, 0xaa, 0xfa, 0x47, 0x86, 0xca, 0x0e, 0x13, 0xcf, 0x8a,
		0x46, 0x02, 0x00, 0x25, 0x4a, 0x6f, 0x94, 0xb9, 0xde, 0x04,
		0x29, 0x4e, 0x73, 0x98, 0xbd, 0xe2, 0x08, 0x2d, 0x52, 0x0c,
		0x68, 0x65, 0x6c, 0x6c, 0x6f, 0x20, 0x77, 0x6f, 0x72, 0x6c,
		0x64, 0x21, 0x04, 0x7f, 0x00, 0x00, 0x01, 0x50, 0xc3, 0x02,
		0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00,
		0x00, 0x00, 0x00, 0x04, 0x00, 0x40, 0x0d, 0x03,
	}

	writeObject := createTestObject()

	readStream := CreateReadStream(buffer)
	readObject := &testObject{
		bytes: make([]byte, testByteCount),
	}
	err := readObject.Serialize(readStream)
	assert.Nil(t, err)

	assert.True(t, reflect.DeepEqual(readObject, writeObject))
}

func TestTriMatrix(t *testing.T) {
	t.Parallel()
	assert.Equal(t, 0, TriMatrixLength(0))
	assert.Equal(t, 0, TriMatrixLength(1))
	assert.Equal(t, 1, TriMatrixLength(2))
	assert.Equal(t, 3, TriMatrixLength(3))
	assert.Equal(t, 6, TriMatrixLength(4))
	size := 100
	length := TriMatrixLength(size)
	values := make([]int, length)
	for i := 0; i < size; i++ {
		for j := 0; j < i; j++ {
			index := TriMatrixIndex(i, j)
			values[index] = index
		}
	}
	for i := 0; i < size; i++ {
		for j := 0; j < i; j++ {
			index := TriMatrixIndex(i, j)
			assert.Equal(t, index, values[index])
		}
	}
}

func TestProtocolVersionAtLeast(t *testing.T) {

	t.Parallel()

	assert.True(t, ProtocolVersionAtLeast(0, 0, 0, 3, 1, 2))
	assert.True(t, ProtocolVersionAtLeast(0, 0, 0, -1, -1, -1))

	assert.True(t, ProtocolVersionAtLeast(3, 0, 0, 3, 0, 0))
	assert.True(t, ProtocolVersionAtLeast(4, 0, 0, 3, 0, 0))
	assert.True(t, ProtocolVersionAtLeast(3, 1, 0, 3, 0, 0))
	assert.True(t, ProtocolVersionAtLeast(3, 0, 1, 3, 0, 0))

	assert.True(t, ProtocolVersionAtLeast(3, 4, 5, 3, 4, 5))
	assert.True(t, ProtocolVersionAtLeast(4, 0, 0, 3, 4, 5))
	assert.True(t, ProtocolVersionAtLeast(3, 5, 0, 3, 4, 5))
	assert.True(t, ProtocolVersionAtLeast(3, 4, 6, 3, 4, 5))

	assert.True(t, !ProtocolVersionAtLeast(3, 0, 99, 3, 1, 1))
	assert.True(t, !ProtocolVersionAtLeast(3, 1, 0, 3, 1, 1))
	assert.True(t, !ProtocolVersionAtLeast(2, 0, 0, 3, 1, 1))

	assert.True(t, !ProtocolVersionAtLeast(3, 0, 5, 3, 1, 0))
	assert.True(t, ProtocolVersionAtLeast(3, 1, 0, 3, 1, 0))
}

func TestRouteManager(t *testing.T) {

	t.Parallel()

	routeManager := NewRouteManager()

	assert.Equal(t, 0, routeManager.NumRoutes)

	routeManager.AddRoute(100, 1, 2, 3)
	assert.Equal(t, 1, routeManager.NumRoutes)
	assert.Equal(t, int32(100), routeManager.RouteRTT[0])
	assert.Equal(t, int32(3), routeManager.RouteNumRelays[0])
	assert.Equal(t, uint32(1), routeManager.RouteRelays[0][0])
	assert.Equal(t, uint32(2), routeManager.RouteRelays[0][1])
	assert.Equal(t, uint32(3), routeManager.RouteRelays[0][2])

	routeManager.AddRoute(200, 4, 5, 6)
	assert.Equal(t, 2, routeManager.NumRoutes)

	routeManager.AddRoute(100, 4, 5, 6)
	assert.Equal(t, 2, routeManager.NumRoutes)

	routeManager.AddRoute(190, 5, 6, 7, 8, 9)
	assert.Equal(t, 3, routeManager.NumRoutes)

	routeManager.AddRoute(180, 6, 7, 8)
	assert.Equal(t, 4, routeManager.NumRoutes)

	routeManager.AddRoute(175, 8, 9)
	assert.Equal(t, 5, routeManager.NumRoutes)

	routeManager.AddRoute(160, 9, 10, 11)
	assert.Equal(t, 6, routeManager.NumRoutes)

	routeManager.AddRoute(165, 10, 11, 12, 13, 14)
	assert.Equal(t, 7, routeManager.NumRoutes)

	routeManager.AddRoute(150, 11, 12)
	assert.Equal(t, 8, routeManager.NumRoutes)

	for i := 0; i < routeManager.NumRoutes-1; i++ {
		assert.True(t, routeManager.RouteRTT[i] <= routeManager.RouteRTT[i+1])
	}

	routeManager.AddRoute(1000, 12, 13, 14)
	assert.Equal(t, routeManager.NumRoutes, 8)
	for i := 0; i < routeManager.NumRoutes; i++ {
		assert.True(t, routeManager.RouteRTT[i] != 1000)
	}

	routeManager.AddRoute(177, 13, 14, 15, 16, 17)
	assert.Equal(t, routeManager.NumRoutes, 8)
	for i := 0; i < routeManager.NumRoutes-1; i++ {
		assert.True(t, routeManager.RouteRTT[i] <= routeManager.RouteRTT[i+1])
	}
	found := false
	for i := 0; i < routeManager.NumRoutes; i++ {
		if routeManager.RouteRTT[i] == 177 {
			found = true
		}
	}
	assert.True(t, found)

	assert.Equal(t, int32(100), routeManager.RouteRTT[0])
	assert.Equal(t, int32(3), routeManager.RouteNumRelays[0])
	assert.Equal(t, uint32(1), routeManager.RouteRelays[0][0])
	assert.Equal(t, uint32(2), routeManager.RouteRelays[0][1])
	assert.Equal(t, uint32(3), routeManager.RouteRelays[0][2])
	assert.Equal(t, RouteHash(1, 2, 3), routeManager.RouteHash[0])

	assert.Equal(t, int32(150), routeManager.RouteRTT[1])
	assert.Equal(t, int32(2), routeManager.RouteNumRelays[1])
	assert.Equal(t, uint32(11), routeManager.RouteRelays[1][0])
	assert.Equal(t, uint32(12), routeManager.RouteRelays[1][1])
	assert.Equal(t, RouteHash(11, 12), routeManager.RouteHash[1])

	assert.Equal(t, int32(160), routeManager.RouteRTT[2])
	assert.Equal(t, int32(3), routeManager.RouteNumRelays[2])
	assert.Equal(t, uint32(9), routeManager.RouteRelays[2][0])
	assert.Equal(t, uint32(10), routeManager.RouteRelays[2][1])
	assert.Equal(t, uint32(11), routeManager.RouteRelays[2][2])
	assert.Equal(t, RouteHash(9, 10, 11), routeManager.RouteHash[2])

	assert.Equal(t, int32(165), routeManager.RouteRTT[3])
	assert.Equal(t, int32(5), routeManager.RouteNumRelays[3])
	assert.Equal(t, uint32(10), routeManager.RouteRelays[3][0])
	assert.Equal(t, uint32(11), routeManager.RouteRelays[3][1])
	assert.Equal(t, uint32(12), routeManager.RouteRelays[3][2])
	assert.Equal(t, uint32(13), routeManager.RouteRelays[3][3])
	assert.Equal(t, uint32(14), routeManager.RouteRelays[3][4])
	assert.Equal(t, RouteHash(10, 11, 12, 13, 14), routeManager.RouteHash[3])

	assert.Equal(t, int32(175), routeManager.RouteRTT[4])
	assert.Equal(t, int32(2), routeManager.RouteNumRelays[4])
	assert.Equal(t, uint32(8), routeManager.RouteRelays[4][0])
	assert.Equal(t, uint32(9), routeManager.RouteRelays[4][1])
	assert.Equal(t, RouteHash(8, 9), routeManager.RouteHash[4])

	assert.Equal(t, int32(177), routeManager.RouteRTT[5])
	assert.Equal(t, int32(5), routeManager.RouteNumRelays[5])
	assert.Equal(t, uint32(13), routeManager.RouteRelays[5][0])
	assert.Equal(t, uint32(14), routeManager.RouteRelays[5][1])
	assert.Equal(t, uint32(15), routeManager.RouteRelays[5][2])
	assert.Equal(t, uint32(16), routeManager.RouteRelays[5][3])
	assert.Equal(t, uint32(17), routeManager.RouteRelays[5][4])
	assert.Equal(t, RouteHash(13, 14, 15, 16, 17), routeManager.RouteHash[5])

	assert.Equal(t, int32(180), routeManager.RouteRTT[6])
	assert.Equal(t, int32(3), routeManager.RouteNumRelays[6])
	assert.Equal(t, uint32(6), routeManager.RouteRelays[6][0])
	assert.Equal(t, uint32(7), routeManager.RouteRelays[6][1])
	assert.Equal(t, uint32(8), routeManager.RouteRelays[6][2])
	assert.Equal(t, RouteHash(6, 7, 8), routeManager.RouteHash[6])

	assert.Equal(t, int32(190), routeManager.RouteRTT[7])
	assert.Equal(t, int32(5), routeManager.RouteNumRelays[7])
	assert.Equal(t, uint32(5), routeManager.RouteRelays[7][0])
	assert.Equal(t, uint32(6), routeManager.RouteRelays[7][1])
	assert.Equal(t, uint32(7), routeManager.RouteRelays[7][2])
	assert.Equal(t, uint32(8), routeManager.RouteRelays[7][3])
	assert.Equal(t, uint32(9), routeManager.RouteRelays[7][4])
	assert.Equal(t, RouteHash(5, 6, 7, 8, 9), routeManager.RouteHash[7])
}

func TestCostMatrix(t *testing.T) {

	raw, err := ioutil.ReadFile("test_data/cost.bin")
	assert.Nil(t, err)
	assert.Equal(t, len(raw), 355188, "cost.bin should be 355188 bytes")

	costMatrix, err := ReadCostMatrix(raw)
	assert.Nil(t, err)

	buffer := make([]byte, 10*1024*1024)

	costMatrixData := WriteCostMatrix(buffer, costMatrix)

	readCostMatrix, err := ReadCostMatrix(costMatrixData)
	assert.Nil(t, err)
	assert.NotNil(t, readCostMatrix)
	if readCostMatrix == nil {
		return
	}

	assert.Equal(t, costMatrix.RelayIds, readCostMatrix.RelayIds, "relay id mismatch")
	// todo: costMatrix.RelayNames once the version is updated
	assert.Equal(t, costMatrix.RelayAddresses, readCostMatrix.RelayAddresses, "relay address mismatch")
	assert.Equal(t, costMatrix.RelayPublicKeys, readCostMatrix.RelayPublicKeys, "relay public key mismatch")
	assert.Equal(t, costMatrix.DatacenterRelays, readCostMatrix.DatacenterRelays, "datacenter relays mismatch")
	assert.Equal(t, costMatrix.RTT, readCostMatrix.RTT, "relay rtt mismatch")
}

func TestRouteMatrixSanity(t *testing.T) {

	raw, err := ioutil.ReadFile("test_data/cost-for-sanity-check.bin")
	assert.Nil(t, err)

	costMatrix, err := ReadCostMatrix(raw)
	assert.Nil(t, err)

	routeMatrix := Optimize(costMatrix, 1.0)

	src := routeMatrix.RelayIds
	dest := routeMatrix.RelayIds

	for i := range src {
		for j := range dest {
			if j < i {
				ijFlatIndex := TriMatrixIndex(i, j)

				entries := routeMatrix.Entries[ijFlatIndex]
				for k := 0; k < int(entries.NumRoutes); k++ {
					numRelays := entries.RouteNumRelays[k]
					firstRelay := entries.RouteRelays[k][0]
					lastRelay := entries.RouteRelays[k][numRelays-1]

					assert.Equal(t, src[firstRelay], dest[i], "invalid route entry #%d at (%d,%d), near relay %d (idx %d) != %d (idx %d)\n", k, i, j, src[firstRelay], firstRelay, dest[i], i)
					assert.Equal(t, src[lastRelay], dest[j], "invalid route entry #%d at (%d,%d), dest relay %d (idx %d) != %d (idx %d)\n", k, i, j, src[lastRelay], lastRelay, dest[j], j)
				}
			}
		}
	}

}

func TestRouteMatrix(t *testing.T) {

	raw, err := ioutil.ReadFile("test_data/cost.bin")
	assert.Nil(t, err)
	assert.Equal(t, 355188, len(raw), "cost.bin should be 355188 bytes")

	costMatrix, err := ReadCostMatrix(raw)
	assert.Nil(t, err)

	buffer := make([]byte, 10*1024*1024)

	costMatrixData := WriteCostMatrix(buffer, costMatrix)

	readCostMatrix, err := ReadCostMatrix(costMatrixData)
	assert.Nil(t, err)
	assert.NotNil(t, readCostMatrix)
	if readCostMatrix == nil {
		return
	}

	routeMatrix := Optimize(costMatrix, 5)
	assert.NotNil(t, routeMatrix)
	assert.Equal(t, costMatrix.RelayIds, routeMatrix.RelayIds, "relay id mismatch")
	assert.Equal(t, costMatrix.RelayAddresses, routeMatrix.RelayAddresses, "relay address mismatch")
	assert.Equal(t, costMatrix.RelayPublicKeys, routeMatrix.RelayPublicKeys, "relay public key mismatch")

	buffer = make([]byte, 20*1024*1024)
	buffer = WriteRouteMatrix(buffer, routeMatrix)

	readRouteMatrix, err := ReadRouteMatrix(buffer)
	assert.Nil(t, err)
	assert.NotNil(t, readRouteMatrix)
	if readRouteMatrix == nil {
		return
	}

	assert.Equal(t, routeMatrix.RelayIds, readRouteMatrix.RelayIds, "relay id mismatch")
	// todo: relay names soon
	assert.Equal(t, routeMatrix.RelayAddresses, readRouteMatrix.RelayAddresses, "relay address mismatch")
	assert.Equal(t, routeMatrix.RelayPublicKeys, readRouteMatrix.RelayPublicKeys, "relay public key mismatch")
	assert.Equal(t, routeMatrix.DatacenterRelays, readRouteMatrix.DatacenterRelays, "datacenter relays mismatch")

	equal := true

	for i := 0; i < len(routeMatrix.Entries); i++ {

		if routeMatrix.Entries[i].DirectRTT != readRouteMatrix.Entries[i].DirectRTT {
			fmt.Printf("DirectRTT mismatch\n")
			equal = false
			break
		}

		if routeMatrix.Entries[i].NumRoutes != readRouteMatrix.Entries[i].NumRoutes {
			fmt.Printf("NumRoutes mismatch\n")
			equal = false
			break
		}

		for j := 0; j < int(routeMatrix.Entries[i].NumRoutes); j++ {

			if routeMatrix.Entries[i].RouteRTT[j] != readRouteMatrix.Entries[i].RouteRTT[j] {
				fmt.Printf("RouteRTT mismatch\n")
				equal = false
				break
			}

			if routeMatrix.Entries[i].RouteNumRelays[j] != readRouteMatrix.Entries[i].RouteNumRelays[j] {
				fmt.Printf("RouteNumRelays mismatch\n")
				equal = false
				break
			}

			for k := 0; k < int(routeMatrix.Entries[i].RouteNumRelays[j]); k++ {
				if routeMatrix.Entries[i].RouteRelays[j][k] != readRouteMatrix.Entries[i].RouteRelays[j][k] {
					fmt.Printf("RouteRelayId mismatch\n")
					equal = false
					break
				}
			}
		}
	}

	assert.True(t, equal, "route matrix entries mismatch")

	Analyze(t, readRouteMatrix)
}

func Analyze(t *testing.T, route_matrix *RouteMatrix) {

	src := route_matrix.RelayIds
	dest := route_matrix.RelayIds

	entries := make([]int32, 0, len(src)*len(dest))

	numRelayPairs := 0
	numValidRelayPairs := 0
	numValidRelayPairsWithoutImprovement := 0

	buckets := make([]int, 11)

	for i := range src {
		for j := range dest {
			if j < i {
				numRelayPairs++
				abFlatIndex := TriMatrixIndex(i, j)
				if len(route_matrix.Entries[abFlatIndex].RouteRTT) > 0 {
					numValidRelayPairs++
					improvement := route_matrix.Entries[abFlatIndex].DirectRTT - route_matrix.Entries[abFlatIndex].RouteRTT[0]
					if improvement > 0.0 {
						entries = append(entries, improvement)
						if improvement <= 5 {
							buckets[0]++
						} else if improvement <= 10 {
							buckets[1]++
						} else if improvement <= 15 {
							buckets[2]++
						} else if improvement <= 20 {
							buckets[3]++
						} else if improvement <= 25 {
							buckets[4]++
						} else if improvement <= 30 {
							buckets[5]++
						} else if improvement <= 35 {
							buckets[6]++
						} else if improvement <= 40 {
							buckets[7]++
						} else if improvement <= 45 {
							buckets[8]++
						} else if improvement <= 50 {
							buckets[9]++
						} else {
							buckets[10]++
						}
					} else {
						numValidRelayPairsWithoutImprovement++
					}
				}
			}
		}
	}

	assert.Equal(t, 43916, numValidRelayPairsWithoutImprovement, "optimizer is broken")

	expected := []int{2561, 8443, 6531, 4690, 3208, 2336, 1775, 1364, 1078, 749, 5159}

	assert.Equal(t, expected, buckets, "optimizer is broken")
}

// -----------------------------------------------------

func GetTestRelayId(name string) RelayId {
	hash := fnv.New32a()
	hash.Write([]byte(name))
	return RelayId(hash.Sum32())
}

type TestRelayData struct {
	id         RelayId
	name       string
	address    *net.UDPAddr
	publicKey  []byte
	privateKey []byte
	index      int
}

type TestEnvironment struct {
	relayArray []*TestRelayData
	relays     map[string]*TestRelayData
	rtt        [][]int32
}

func NewTestEnvironment() *TestEnvironment {
	env := &TestEnvironment{}
	env.relays = make(map[string]*TestRelayData)
	return env
}

func (env *TestEnvironment) Clear() {
	numRelays := len(env.relays)
	env.rtt = make([][]int32, numRelays)
	for i := 0; i < numRelays; i++ {
		env.rtt[i] = make([]int32, numRelays)
		for j := 0; j < numRelays; j++ {
			env.rtt[i][j] = -1
		}
	}
}

func (env *TestEnvironment) AddRelay(relayName string, relayAddress string) {
	relay := &TestRelayData{}
	relay.id = GetTestRelayId(relayName)
	relay.name = relayName
	relay.address = ParseAddress(relayAddress)
	relay.publicKey, relay.privateKey, _ = GenerateRelayKeyPair()
	relay.index = len(env.relayArray)
	env.relays[relayName] = relay
	env.relayArray = append(env.relayArray, relay)
	env.Clear()
}

func (env *TestEnvironment) SetRTT(sourceRelayName string, destRelayName string, rtt int32) {
	i := env.relays[sourceRelayName].index
	j := env.relays[destRelayName].index
	if j > i {
		i, j = j, i
	}
	env.rtt[i][j] = rtt
}

func (env *TestEnvironment) GetRelayData(relayName string) *TestRelayData {
	return env.relays[relayName]
}

func (env *TestEnvironment) GetCostMatrix() *CostMatrix {
	costMatrix := &CostMatrix{}
	numRelays := len(env.relays)
	costMatrix.RelayIds = make([]RelayId, numRelays)
	costMatrix.RelayNames = make([]string, numRelays)
	costMatrix.RelayAddresses = make([][]byte, numRelays)
	costMatrix.RelayPublicKeys = make([][]byte, numRelays)
	for i := range env.relayArray {
		costMatrix.RelayIds[i] = env.relayArray[i].id
		costMatrix.RelayNames[i] = env.relayArray[i].name
		costMatrix.RelayAddresses[i] = []byte(env.relayArray[i].address.String())
		costMatrix.RelayPublicKeys[i] = env.relayArray[i].publicKey
	}
	entryCount := TriMatrixLength(numRelays)
	costMatrix.RTT = make([]int32, entryCount)
	for i := 0; i < numRelays; i++ {
		for j := 0; j < i; j++ {
			index := TriMatrixIndex(i, j)
			costMatrix.RTT[index] = env.rtt[i][j]
		}
	}
	costMatrix.DatacenterRelays = make(map[DatacenterId][]RelayId)
	return costMatrix
}

type TestRouteData struct {
	rtt    int32
	relays []string
}

func (env *TestEnvironment) GetRoutes(routeMatrix *RouteMatrix, sourceRelayName string, destRelayName string) []TestRouteData {
	sourceRelay := env.relays[sourceRelayName]
	destRelay := env.relays[destRelayName]
	i := sourceRelay.index
	j := destRelay.index
	if i == j {
		return nil
	}
	index := TriMatrixIndex(i, j)
	entry := routeMatrix.Entries[index]
	testRouteData := make([]TestRouteData, entry.NumRoutes)
	for k := 0; k < int(entry.NumRoutes); k++ {
		testRouteData[k].rtt = entry.RouteRTT[k]
		testRouteData[k].relays = make([]string, entry.RouteNumRelays[k])
		if j < i {
			for l := 0; l < int(entry.RouteNumRelays[k]); l++ {
				relayIndex := entry.RouteRelays[k][l]
				testRouteData[k].relays[l] = env.relayArray[relayIndex].name
			}
		} else {
			for l := 0; l < int(entry.RouteNumRelays[k]); l++ {
				relayIndex := entry.RouteRelays[k][int(entry.RouteNumRelays[k])-1-l]
				testRouteData[k].relays[l] = env.relayArray[relayIndex].name
			}

		}
	}
	return testRouteData
}

func TestTheTestEnvironment(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")
	env.AddRelay("a", "10.0.0.3")
	env.AddRelay("b", "10.0.0.4")
	env.AddRelay("c", "10.0.0.5")
	env.AddRelay("d", "10.0.0.6")
	env.AddRelay("e", "10.0.0.7")

	env.SetRTT("losangeles", "chicago", 100)

	costMatrix := env.GetCostMatrix()

	routeMatrix := Optimize(costMatrix, 5)

	numRelays := len(costMatrix.RelayNames)

	for i := 0; i < numRelays; i++ {
		for j := 0; j < numRelays; j++ {
			sourceRelayName := costMatrix.RelayNames[i]
			destRelayName := costMatrix.RelayNames[j]
			routes := env.GetRoutes(routeMatrix, sourceRelayName, destRelayName)
			if sourceRelayName == "losangeles" && destRelayName == "chicago" {
				assert.Equal(t, 1, len(routes))
				if len(routes) == 1 {
					assert.Equal(t, int32(100), routes[0].rtt)
					assert.Equal(t, []string{"losangeles", "chicago"}, routes[0].relays)
				}
			} else if sourceRelayName == "chicago" && destRelayName == "losangeles" {
				assert.Equal(t, 1, len(routes))
				if len(routes) == 1 {
					assert.Equal(t, int32(100), routes[0].rtt)
					assert.Equal(t, []string{"chicago", "losangeles"}, routes[0].relays)
				}
			} else {
				assert.Equal(t, 0, len(routes))
			}
		}
	}
}

func TestCostMatrixReadAndWrite(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")
	env.AddRelay("a", "10.0.0.3")
	env.AddRelay("b", "10.0.0.4")
	env.AddRelay("c", "10.0.0.5")
	env.AddRelay("d", "10.0.0.6")
	env.AddRelay("e", "10.0.0.7")

	env.SetRTT("losangeles", "a", 10)
	env.SetRTT("a", "chicago", 10)

	costMatrix := env.GetCostMatrix()
	costMatrix.DatacenterIds = []DatacenterId{
		DatacenterId(0),
		DatacenterId(1),
		DatacenterId(2),
	}
	costMatrix.DatacenterNames = []string{
		"a",
		"b",
		"c",
	}

	buffer := make([]byte, 10*1024*1024)

	costMatrixData := WriteCostMatrix(buffer, costMatrix)

	readCostMatrix, err := ReadCostMatrix(costMatrixData)
	assert.Nil(t, err)
	assert.NotNil(t, readCostMatrix)
	if readCostMatrix == nil {
		return
	}

	assert.Equal(t, costMatrix.RelayIds, readCostMatrix.RelayIds, "relay id mismatch")
	assert.Equal(t, costMatrix.RelayNames, readCostMatrix.RelayNames, "relay name mismatch")
	assert.Equal(t, costMatrix.RelayAddresses, readCostMatrix.RelayAddresses, "relay address mismatch")
	assert.Equal(t, costMatrix.RelayPublicKeys, readCostMatrix.RelayPublicKeys, "relay public key mismatch")
	assert.Equal(t, costMatrix.DatacenterIds, readCostMatrix.DatacenterIds, "datacenter id mismatch")
	assert.Equal(t, costMatrix.DatacenterNames, readCostMatrix.DatacenterNames, "datacenter names mismatch")
	assert.Equal(t, costMatrix.DatacenterRelays, readCostMatrix.DatacenterRelays, "datacenter relays mismatch")
	assert.Equal(t, costMatrix.RTT, readCostMatrix.RTT, "relay rtt mismatch")
}

func TestRouteMatrixReadAndWrite(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")
	env.AddRelay("a", "10.0.0.3")
	env.AddRelay("b", "10.0.0.4")
	env.AddRelay("c", "10.0.0.5")
	env.AddRelay("d", "10.0.0.6")
	env.AddRelay("e", "10.0.0.7")

	env.SetRTT("losangeles", "a", 10)
	env.SetRTT("a", "chicago", 10)

	costMatrix := env.GetCostMatrix()

	routeMatrix := Optimize(costMatrix, 5)

	buffer := make([]byte, 10*1024*1024)

	routeMatrixData := WriteRouteMatrix(buffer, routeMatrix)

	readRouteMatrix, err := ReadRouteMatrix(routeMatrixData)
	assert.Nil(t, err)
	assert.NotNil(t, readRouteMatrix)
	if readRouteMatrix == nil {
		return
	}

	assert.Equal(t, routeMatrix.RelayIds, readRouteMatrix.RelayIds, "relay id mismatch")
	assert.Equal(t, routeMatrix.RelayNames, readRouteMatrix.RelayNames, "relay name mismatch")
	assert.Equal(t, routeMatrix.RelayAddresses, readRouteMatrix.RelayAddresses, "relay address mismatch")
	assert.Equal(t, routeMatrix.RelayPublicKeys, readRouteMatrix.RelayPublicKeys, "relay public key mismatch")
	assert.Equal(t, routeMatrix.DatacenterRelays, readRouteMatrix.DatacenterRelays, "datacenter relays mismatch")

	equal := true

	for i := 0; i < len(routeMatrix.Entries); i++ {

		if routeMatrix.Entries[i].DirectRTT != readRouteMatrix.Entries[i].DirectRTT {
			fmt.Printf("DirectRTT mismatch\n")
			equal = false
			break
		}

		if routeMatrix.Entries[i].NumRoutes != readRouteMatrix.Entries[i].NumRoutes {
			fmt.Printf("NumRoutes mismatch\n")
			equal = false
			break
		}

		for j := 0; j < int(routeMatrix.Entries[i].NumRoutes); j++ {

			if routeMatrix.Entries[i].RouteRTT[j] != readRouteMatrix.Entries[i].RouteRTT[j] {
				fmt.Printf("RouteRTT mismatch\n")
				equal = false
				break
			}

			if routeMatrix.Entries[i].RouteNumRelays[j] != readRouteMatrix.Entries[i].RouteNumRelays[j] {
				fmt.Printf("RouteNumRelays mismatch\n")
				equal = false
				break
			}

			for k := 0; k < int(routeMatrix.Entries[i].RouteNumRelays[j]); k++ {
				if routeMatrix.Entries[i].RouteRelays[j][k] != readRouteMatrix.Entries[i].RouteRelays[j][k] {
					fmt.Printf("RouteRelayId mismatch\n")
					equal = false
					break
				}
			}
		}
	}

	assert.True(t, equal, "route matrix entries mismatch")
}

func TestIndirectRoute3(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")
	env.AddRelay("a", "10.0.0.3")
	env.AddRelay("b", "10.0.0.4")
	env.AddRelay("c", "10.0.0.5")
	env.AddRelay("d", "10.0.0.6")
	env.AddRelay("e", "10.0.0.7")

	env.SetRTT("losangeles", "a", 10)
	env.SetRTT("a", "chicago", 10)

	costMatrix := env.GetCostMatrix()

	routeMatrix := Optimize(costMatrix, 5)

	routes := env.GetRoutes(routeMatrix, "losangeles", "chicago")

	assert.Equal(t, 1, len(routes))
	if len(routes) == 1 {
		assert.Equal(t, int32(20), routes[0].rtt)
		assert.Equal(t, 3, len(routes[0].relays))
		if len(routes[0].relays) == 3 {
			assert.Equal(t, []string{"losangeles", "a", "chicago"}, routes[0].relays)
		}
	}
}

func TestIndirectRoute4(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")
	env.AddRelay("a", "10.0.0.3")
	env.AddRelay("b", "10.0.0.4")
	env.AddRelay("c", "10.0.0.5")
	env.AddRelay("d", "10.0.0.6")
	env.AddRelay("e", "10.0.0.7")

	env.SetRTT("losangeles", "a", 10)
	env.SetRTT("losangeles", "b", 100)
	env.SetRTT("a", "b", 10)
	env.SetRTT("b", "chicago", 10)

	costMatrix := env.GetCostMatrix()

	routeMatrix := Optimize(costMatrix, 5)

	routes := env.GetRoutes(routeMatrix, "losangeles", "chicago")

	assert.True(t, len(routes) >= 1)
	if len(routes) >= 1 {
		assert.Equal(t, int32(30), routes[0].rtt)
		assert.Equal(t, []string{"losangeles", "a", "b", "chicago"}, routes[0].relays)
	}
}

func TestIndirectRoute5(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")
	env.AddRelay("a", "10.0.0.3")
	env.AddRelay("b", "10.0.0.4")
	env.AddRelay("c", "10.0.0.5")
	env.AddRelay("d", "10.0.0.6")
	env.AddRelay("e", "10.0.0.7")

	env.SetRTT("losangeles", "a", 10)
	env.SetRTT("a", "b", 10)
	env.SetRTT("b", "c", 10)
	env.SetRTT("c", "chicago", 10)

	env.SetRTT("losangeles", "b", 100)
	env.SetRTT("b", "chicago", 100)

	costMatrix := env.GetCostMatrix()

	routeMatrix := Optimize(costMatrix, 5)

	routes := env.GetRoutes(routeMatrix, "losangeles", "chicago")

	// fmt.Printf("routes:\n%v\n", routes)

	assert.True(t, len(routes) >= 1)
	if len(routes) >= 1 {
		assert.Equal(t, int32(40), routes[0].rtt)
		assert.Equal(t, []string{"losangeles", "a", "b", "c", "chicago"}, routes[0].relays)
	}
}

func TestFasterRoute3(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")
	env.AddRelay("a", "10.0.0.3")

	env.SetRTT("losangeles", "chicago", 100)
	env.SetRTT("losangeles", "a", 10)
	env.SetRTT("a", "chicago", 10)

	costMatrix := env.GetCostMatrix()

	routeMatrix := Optimize(costMatrix, 5)

	routes := env.GetRoutes(routeMatrix, "losangeles", "chicago")

	assert.Equal(t, 2, len(routes))
	if len(routes) == 2 {
		assert.Equal(t, int32(20), routes[0].rtt)
		assert.Equal(t, []string{"losangeles", "a", "chicago"}, routes[0].relays)
		assert.Equal(t, int32(100), routes[1].rtt)
		assert.Equal(t, []string{"losangeles", "chicago"}, routes[1].relays)
	}
}

func TestFasterRoute4(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")
	env.AddRelay("a", "10.0.0.3")
	env.AddRelay("b", "10.0.0.4")

	env.SetRTT("losangeles", "chicago", 100)
	env.SetRTT("losangeles", "a", 10)
	env.SetRTT("a", "chicago", 50)
	env.SetRTT("a", "b", 10)
	env.SetRTT("b", "chicago", 10)

	costMatrix := env.GetCostMatrix()

	routeMatrix := Optimize(costMatrix, 5)

	routes := env.GetRoutes(routeMatrix, "losangeles", "chicago")

	assert.Equal(t, 3, len(routes))
	if len(routes) == 3 {
		assert.Equal(t, int32(30), routes[0].rtt)
		assert.Equal(t, []string{"losangeles", "a", "b", "chicago"}, routes[0].relays)
	}
}

func TestFasterRoute5(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")
	env.AddRelay("a", "10.0.0.3")
	env.AddRelay("b", "10.0.0.4")
	env.AddRelay("c", "10.0.0.5")

	env.SetRTT("losangeles", "chicago", 1000)
	env.SetRTT("losangeles", "a", 10)
	env.SetRTT("losangeles", "b", 100)
	env.SetRTT("losangeles", "c", 100)
	env.SetRTT("a", "chicago", 100)
	env.SetRTT("b", "chicago", 100)
	env.SetRTT("c", "chicago", 10)
	env.SetRTT("a", "b", 10)
	env.SetRTT("a", "c", 100)
	env.SetRTT("b", "c", 10)

	costMatrix := env.GetCostMatrix()

	routeMatrix := Optimize(costMatrix, 5)

	routes := env.GetRoutes(routeMatrix, "losangeles", "chicago")

	assert.Equal(t, 7, len(routes))
	if len(routes) == 7 {
		assert.Equal(t, int32(40), routes[0].rtt)
		assert.Equal(t, []string{"losangeles", "a", "b", "c", "chicago"}, routes[0].relays)
	}
}

func TestSlowerRoute(t *testing.T) {

	t.Parallel()

	env := NewTestEnvironment()

	env.AddRelay("losangeles", "10.0.0.1")
	env.AddRelay("chicago", "10.0.0.2")
	env.AddRelay("a", "10.0.0.3")
	env.AddRelay("b", "10.0.0.4")
	env.AddRelay("c", "10.0.0.5")

	env.SetRTT("losangeles", "chicago", 10)
	env.SetRTT("losangeles", "a", 100)
	env.SetRTT("a", "chicago", 100)
	env.SetRTT("b", "chicago", 100)
	env.SetRTT("c", "chicago", 100)
	env.SetRTT("a", "b", 100)
	env.SetRTT("a", "c", 100)
	env.SetRTT("b", "c", 100)

	costMatrix := env.GetCostMatrix()

	routeMatrix := Optimize(costMatrix, 5)

	routes := env.GetRoutes(routeMatrix, "losangeles", "chicago")

	assert.Equal(t, 1, len(routes))
	if len(routes) == 1 {
		assert.Equal(t, int32(10), routes[0].rtt)
		assert.Equal(t, []string{"losangeles", "chicago"}, routes[0].relays)
	}
}

func TestHistoryMax(t *testing.T) {

	t.Parallel()

	history := HistoryNotSet()

	assert.Equal(t, float32(0.0), HistoryMax(history[:]))

	history[0] = 5.0
	history[1] = 3.0
	history[2] = 100.0

	assert.Equal(t, float32(100.0), HistoryMax(history[:]))
}

func TestHistoryMean(t *testing.T) {

	t.Parallel()

	history := HistoryNotSet()

	assert.Equal(t, float32(0.0), HistoryMean(history[:]))

	history[0] = 5.0
	history[1] = 3.0
	history[2] = 100.0

	assert.Equal(t, float32(36.0), HistoryMean(history[:]))
}

func TestHaversineDistance(t *testing.T) {

	t.Parallel()

	losangelesLatitude := 34.0522
	losangelesLongitude := -118.2437

	bostonLatitude := 42.3601
	bostonLongitude := -71.0589

	distance := HaversineDistance(losangelesLatitude, losangelesLongitude, bostonLatitude, bostonLongitude)

	assert.Equal(t, 4169.607203810275, distance)
}

func TestRouteSlice(t *testing.T) {

	t.Parallel()

	var slice RouteSlice

	slice.Flags = 1234

	slice.RouteSample.DirectStats.RTT = 100.0
	slice.RouteSample.DirectStats.Jitter = 10.0
	slice.RouteSample.DirectStats.PacketLoss = 1.0

	slice.RouteSample.NextStats.RTT = 50.0
	slice.RouteSample.NextStats.Jitter = 1.0
	slice.RouteSample.NextStats.PacketLoss = 0.1

	slice.RouteSample.NearRelays = make([]RelayStats, 2)

	slice.RouteSample.NearRelays[0].Id = 10
	slice.RouteSample.NearRelays[0].RTT = 5.0
	slice.RouteSample.NearRelays[0].Jitter = 1.0
	slice.RouteSample.NearRelays[0].PacketLoss = 0.1

	slice.RouteSample.NearRelays[1].Id = 11
	slice.RouteSample.NearRelays[1].RTT = 2.0
	slice.RouteSample.NearRelays[1].Jitter = 0.1
	slice.RouteSample.NearRelays[1].PacketLoss = 0.5

	slice.PredictedRoute = &Route{}
	slice.PredictedRoute.RTT = 50.0
	slice.PredictedRoute.Jitter = 10.0
	slice.PredictedRoute.PacketLoss = 0.01
	slice.PredictedRoute.RelayIds = []RelayId{1, 2, 3, 4}

	const BufferSize = 1024

	writeStream, err := CreateWriteStream(BufferSize)
	assert.Nil(t, err)

	err = slice.Serialize(writeStream)
	assert.Nil(t, err)

	writeStream.Flush()

	bytesWritten := writeStream.GetBytesProcessed()

	buffer := writeStream.GetData()

	readStream := CreateReadStream(buffer[:bytesWritten])

	var readSlice RouteSlice

	err = readSlice.Serialize(readStream)
	assert.Nil(t, err)

	assert.Equal(t, slice, readSlice)
}

func TestRouteSliceReadFail(t *testing.T) {

	t.Parallel()

	buffer := make([]byte, 0)
	readStream := CreateReadStream(buffer)
	var readSlice RouteSlice
	err := readSlice.Serialize(readStream)
	assert.NotNil(t, err)

	buffer = make([]byte, 8)
	err = readSlice.Serialize(readStream)
	assert.NotNil(t, err)
}

// todo: there should be test functions to serialize each of the packet types

// todo: add a test to make sure cost matrix datacenters is correct when pulled from stats db
