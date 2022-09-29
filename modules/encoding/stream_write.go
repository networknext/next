package encoding

import (
	"encoding/binary"
	"fmt"
	"math"
	"net"
	"runtime/debug"
)

type WriteStream struct {
	writer *BitWriter
	err    error
}

func CreateWriteStream(buffer []byte) *WriteStream {
	return &WriteStream{
		writer: CreateBitWriter(buffer),
	}
}

func (stream *WriteStream) IsWriting() bool {
	return true
}

func (stream *WriteStream) IsReading() bool {
	return false
}

// todo: lame function name. also, should handle printf formatting. simpler
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
	max := int32(maxSize)
	stream.SerializeInteger(&length, min, max)
	if length > 0 {
		stream.SerializeBytes([]byte(*value))
	}
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
		addrType = IPAddressNone
	} else if addr.IP.To4() == nil {
		addrType = IPAddressIPv6
	} else {
		addrType = IPAddressIPv4
	}

	stream.SerializeBits(&addrType, 2)
	if stream.err != nil {
		return
	}
	if addrType == uint32(IPAddressIPv4) {
		stream.SerializeBytes(addr.IP[12:])
		if stream.err != nil {
			return
		}
		port := uint32(addr.Port)
		stream.SerializeBits(&port, 16)
	} else if addrType == uint32(IPAddressIPv6) {
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
