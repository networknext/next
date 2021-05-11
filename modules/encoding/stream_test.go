package encoding

import (
	"net"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

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

type TestObject struct {
	a           int32
	b           int32
	c           int32
	d           uint32
	e           uint32
	f           uint32
	g           bool
	items       [16]uint32
	floatValue  float32
	doubleValue float64
	uint64Value uint64
	bytes       [32]byte
	str         string
	addressA    net.UDPAddr
	addressB    net.UDPAddr
	addressC    net.UDPAddr
}

func parseAddress(input string) *net.UDPAddr {
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

func createTestObject() *TestObject {
	testObject := TestObject{}
	for i := range testObject.bytes {
		testObject.bytes[i] = byte(65 + i)
	}
	for i := range testObject.items {
		testObject.items[i] = uint32(i)
	}
	testObject.a = 1
	testObject.b = -2
	testObject.c = 150
	testObject.d = 55
	testObject.e = 255
	testObject.f = 127
	testObject.g = true
	testObject.floatValue = 3.1415926
	testObject.doubleValue = 1.0 / 3.0
	testObject.uint64Value = 0x1234567898765432
	testObject.str = "hello world!"
	testObject.addressA = net.UDPAddr{}
	testObject.addressB = *parseAddress("127.0.0.1:50000")
	testObject.addressC = *parseAddress("[::1]:50000")
	return &testObject
}

func (obj *TestObject) Serialize(stream Stream) error {

	stream.SerializeInteger(&obj.a, -10, 10)
	stream.SerializeInteger(&obj.b, -10, 10)

	stream.SerializeInteger(&obj.c, -100, 10000)

	stream.SerializeBits(&obj.d, 6)
	stream.SerializeBits(&obj.e, 8)
	stream.SerializeBits(&obj.f, 7)

	stream.SerializeAlign()

	stream.SerializeBool(&obj.g)

	for i := range obj.items {
		stream.SerializeBits(&obj.items[i], 8)
	}

	stream.SerializeFloat32(&obj.floatValue)

	stream.SerializeFloat64(&obj.doubleValue)

	stream.SerializeUint64(&obj.uint64Value)

	stream.SerializeBytes(obj.bytes[:])

	stream.SerializeString(&obj.str, 256)

	stream.SerializeAddress(&obj.addressA)
	stream.SerializeAddress(&obj.addressB)
	stream.SerializeAddress(&obj.addressC)

	return stream.Error()
}

func TestStream(t *testing.T) {

	t.Parallel()

	const BufferSize = 256

	buffer := [BufferSize]byte{}

	writeStream, err := CreateWriteStream(buffer[:])
	assert.Nil(t, err)

	writeObject := createTestObject()
	err = writeObject.Serialize(writeStream)
	assert.Nil(t, err)

	writeStream.Flush()

	readStream := CreateReadStream(buffer[:])
	readObject := &TestObject{}
	err = readObject.Serialize(readStream)
	assert.Nil(t, err)

	assert.Equal(t, writeObject, readObject)
}

// ------------------------------------------
