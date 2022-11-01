package encoding

import (
	"encoding/binary"
	"fmt"
	"math"
	"net"
	"runtime/debug"
)

type ReadStream struct {
	reader *BitReader
	err    error
}

func CreateReadStream(buffer []byte) *ReadStream {
	return &ReadStream{
		reader: CreateBitReader(buffer),
	}
}

func (stream *ReadStream) SetError(err error) {
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
		stream.SetError(fmt.Errorf("value is nil"))
		return
	}
	if min >= max {
		stream.SetError(fmt.Errorf("min (%d) should be less than max (%d)", min, max))
		return
	}
	bits := BitsRequiredSigned(min, max)
	if stream.reader.WouldReadPastEnd(bits) {
		stream.SetError(fmt.Errorf("would read past end of buffer"))
		return
	}
	unsignedValue, err := stream.reader.ReadBits(bits)
	if err != nil {
		stream.SetError(err)
		return
	}
	candidateValue := int32(unsignedValue) + min
	if candidateValue > max {
		stream.SetError(fmt.Errorf("value (%d) above max (%d)", candidateValue, max))
		return
	}
	*value = candidateValue
}

func (stream *ReadStream) SerializeBits(value *uint32, bits int) {
	if stream.err != nil {
		return
	}
	if value == nil {
		stream.SetError(fmt.Errorf("value is nil"))
		return
	}
	if bits < 0 || bits > 32 {
		stream.SetError(fmt.Errorf("bits (%d) should be between 0 and 32 bits", bits))
		return
	}
	if stream.reader.WouldReadPastEnd(bits) {
		stream.SetError(fmt.Errorf("would read past end of buffer"))
		return
	}
	readValue, err := stream.reader.ReadBits(bits)
	if err != nil {
		stream.SetError(err)
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
		stream.SetError(fmt.Errorf("value is nil"))
		return
	}
	if stream.reader.WouldReadPastEnd(1) {
		stream.SetError(fmt.Errorf("would read past end of buffer"))
		return
	}
	readValue, err := stream.reader.ReadBits(1)
	if err != nil {
		stream.SetError(err)
		return
	}
	*value = readValue != 0
}

func (stream *ReadStream) SerializeFloat32(value *float32) {
	if stream.err != nil {
		return
	}
	if value == nil {
		stream.SetError(fmt.Errorf("value is nil"))
		return
	}
	if stream.reader.WouldReadPastEnd(32) {
		stream.SetError(fmt.Errorf("would read past end of buffer"))
		return
	}
	readValue, err := stream.reader.ReadBits(32)
	if err != nil {
		stream.SetError(err)
		return
	}
	*value = math.Float32frombits(readValue)
}

func (stream *ReadStream) SerializeUint64(value *uint64) {
	if stream.err != nil {
		return
	}
	if value == nil {
		stream.SetError(fmt.Errorf("value is nil"))
		return
	}
	if stream.reader.WouldReadPastEnd(64) {
		stream.SetError(fmt.Errorf("would read past end of buffer"))
		return
	}
	lo, err := stream.reader.ReadBits(32)
	if err != nil {
		stream.SetError(err)
		return
	}
	hi, err := stream.reader.ReadBits(32)
	if err != nil {
		stream.SetError(err)
		return
	}
	*value = (uint64(hi) << 32) | uint64(lo)
}

func (stream *ReadStream) SerializeFloat64(value *float64) {
	if stream.err != nil {
		return
	}
	if value == nil {
		stream.SetError(fmt.Errorf("value is nil"))
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
		stream.SetError(fmt.Errorf("buffer should contain more than 0 bytes"))
		return
	}
	stream.SerializeAlign()
	if stream.err != nil {
		return
	}
	if stream.reader.WouldReadPastEnd(len(data) * 8) {
		stream.SetError(fmt.Errorf("SerializeBytes() would read past end of buffer"))
		return
	}
	stream.SetError(stream.reader.ReadBytes(data))
}

func (stream *ReadStream) SerializeString(value *string, maxSize int) {
	if stream.err != nil {
		return
	}
	if maxSize < 0 {
		stream.SetError(fmt.Errorf("maxSize (%d) should be > 0", maxSize))
		return
	}
	length := int32(0)
	min := int32(0)
	max := int32(maxSize) - 1
	stream.SerializeInteger(&length, min, max)
	if stream.err != nil {
		return
	}
	if length == 0 {
		*value = ""
		return
	}
	stringBytes := make([]byte, length)
	stream.SerializeBytes(stringBytes)
	*value = string(stringBytes)
}

func (stream *ReadStream) SerializeAddress(addr *net.UDPAddr) {
	if stream.err != nil {
		return
	}
	if addr == nil {
		stream.SetError(fmt.Errorf("addr is nil"))
		return
	}
	addrType := uint32(0)
	stream.SerializeBits(&addrType, 2)
	if stream.err != nil {
		return
	}
	if addrType == uint32(IPAddressIPv4) {
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
	} else if addrType == uint32(IPAddressIPv6) {
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
		stream.SetError(fmt.Errorf("SerializeAlign() would read past end of buffer"))
		return
	}
	stream.SetError(stream.reader.ReadAlign())
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
