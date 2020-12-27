package encoding

import (
	"encoding/binary"
	"fmt"
	"math/bits"
)

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
