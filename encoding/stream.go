package encoding

import (
	"encoding/binary"
	"fmt"
	"math"
	"net"
	"runtime/debug"
)

const (
	IPAddressNone = 0
	IPAddressIPv4 = 1
	IPAddressIPv6 = 2
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
	if length > 0 {
		stream.SerializeBytes([]byte(*value))
	}
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
	if length == 0 {
		*value = ""
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
