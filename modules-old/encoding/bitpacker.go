package encoding

import (
	"fmt"
	"math/bits"
	"unsafe"
)

// ---------------------------------------------------

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

// ------------------------------------------------------

type BitWriter struct {
	buffer      []byte
	scratch     uint64
	numBits     int
	bitsWritten int
	wordIndex   int
	scratchBits int
	numWords    int
}

func CreateBitWriter(buffer []byte) (*BitWriter, error) {
	bytes := len(buffer)
	if bytes%4 != 0 {
		return nil, fmt.Errorf("bitwriter bytes must be a multiple of 4")
	}
	numWords := bytes / 4
	return &BitWriter{
		buffer:   buffer,
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

	writer.scratch |= uint64(value) << uint(writer.scratchBits)

	writer.scratchBits += bits

	if writer.scratchBits >= 32 {
		*(*uint32)(unsafe.Pointer(&writer.buffer[writer.wordIndex*4])) = HostToNetwork(uint32(writer.scratch & 0xFFFFFFFF))
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
	}
	return nil
}

func (writer *BitWriter) WriteBytes(data []byte) error {

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

	numWords := (len(data) - headBytes) / 4

	for i := 0; i < numWords; i++ {
		*(*uint32)(unsafe.Pointer(&writer.buffer[writer.wordIndex*4])) = *(*uint32)(unsafe.Pointer(&data[headBytes+i*4]))
		writer.bitsWritten += 32
		writer.wordIndex += 1
	}

	writer.scratch = 0

	tailStart := headBytes + numWords*4
	tailBytes := len(data) - tailStart

	for i := 0; i < tailBytes; i++ {
		err := writer.WriteBits(uint32(data[tailStart+i]), 8)
		if err != nil {
			return err
		}
	}

	return nil
}

func (writer *BitWriter) FlushBits() error {
	if writer.scratchBits != 0 {
		*(*uint32)(unsafe.Pointer(&writer.buffer[writer.wordIndex*4])) = HostToNetwork(uint32(writer.scratch & 0xFFFFFFFF))
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
	return writer.buffer
}

func (writer *BitWriter) GetBytesWritten() int {
	return (writer.bitsWritten + 7) / 8
}

// ------------------------------------------------------

type BitReader struct {
	buffer      []byte
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
		buffer:   data,
		numBits:  len(data) * 8,
		numBytes: len(data),
		numWords: (len(data) + 3) / 4,
	}
}

func (reader *BitReader) WouldReadPastEnd(bits int) bool {
	return reader.bitsRead+bits > reader.numBits
}

func (reader *BitReader) ReadBits(bits int) (uint32, error) {

	if reader.bitsRead+bits > reader.numBits {
		return 0, fmt.Errorf("call would read past end of buffer")
	}

	reader.bitsRead += bits

	if reader.scratchBits < bits {
		if reader.wordIndex >= reader.numWords {
			return 0, fmt.Errorf("would read past end of buffer")
		}
		reader.scratch |= uint64(NetworkToHost(*(*uint32)(unsafe.Pointer(&reader.buffer[reader.wordIndex*4])))) << uint(reader.scratchBits)
		reader.scratchBits += 32
		reader.wordIndex++
	}

	output := reader.scratch & ((uint64(1) << uint(bits)) - 1)

	reader.scratch >>= uint(bits)
	reader.scratchBits -= bits

	return uint32(output), nil
}

func (reader *BitReader) ReadAlign() error {
	remainderBits := reader.bitsRead % 8
	if remainderBits != 0 {
		_, err := reader.ReadBits(8 - remainderBits)
		if err != nil {
			return err
		}
	}
	return nil
}

func (reader *BitReader) ReadBytes(buffer []byte) error {

	if reader.bitsRead+len(buffer)*8 > reader.numBits {
		return fmt.Errorf("would read past end of buffer")
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

	numWords := (len(buffer) - headBytes) / 4

	for i := 0; i < numWords; i++ {
		*(*uint32)(unsafe.Pointer(&buffer[headBytes+i*4])) = *(*uint32)(unsafe.Pointer(&reader.buffer[reader.wordIndex*4]))
		reader.bitsRead += 32
		reader.wordIndex += 1
	}

	reader.scratchBits = 0

	tailStart := headBytes + numWords*4
	tailBytes := len(buffer) - tailStart

	for i := 0; i < tailBytes; i++ {
		value, err := reader.ReadBits(8)
		if err != nil {
			return err
		}
		buffer[tailStart+i] = byte(value)
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

// --------------------------------------------------------
