package encoding

import (
	"net"
)

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
