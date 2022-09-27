package encoding

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestBitpacker(t *testing.T) {

	t.Parallel()

	const BufferSize = 256

	buffer := [BufferSize]byte{}

	writer, err := CreateBitWriter(buffer[:])
	assert.Nil(t, err)

	assert.Equal(t, 0, writer.GetBitsWritten())
	assert.Equal(t, 0, writer.GetBytesWritten())
	assert.Equal(t, BufferSize*8, writer.GetBitsAvailable())

	data := [32]byte{}
	for i := range data {
		data[i] = byte(65 + i)
	}

	writer.WriteBits(0, 1)
	writer.WriteBits(1, 1)
	writer.WriteBits(10, 8)
	writer.WriteBits(255, 8)
	writer.WriteBits(1000, 10)
	writer.WriteBits(50000, 16)
	writer.WriteBits(9999999, 32)
	writer.WriteBytes(data[:])
	writer.FlushBits()

	bitsWritten := 1 + 1 + 8 + 8 + 10 + 16 + 32 + 32*8

	assert.Equal(t, 42, writer.GetBytesWritten())
	assert.Equal(t, bitsWritten, writer.GetBitsWritten())
	assert.Equal(t, BufferSize*8-bitsWritten, writer.GetBitsAvailable())

	reader := CreateBitReader(buffer[:])

	assert.Equal(t, 0, reader.GetBitsRead())
	assert.Equal(t, BufferSize*8, reader.GetBitsRemaining())

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

	readData := [32]byte{}
	err = reader.ReadBytes(readData[:])
	assert.Nil(t, err)

	assert.Equal(t, uint32(0), a)
	assert.Equal(t, uint32(1), b)
	assert.Equal(t, uint32(10), c)
	assert.Equal(t, uint32(255), d)
	assert.Equal(t, uint32(1000), e)
	assert.Equal(t, uint32(50000), f)
	assert.Equal(t, uint32(9999999), g)

	assert.Equal(t, bitsWritten, reader.GetBitsRead())
	assert.Equal(t, BufferSize*8-bitsWritten, reader.GetBitsRemaining())
}
