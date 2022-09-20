package packets

import (
	"github.com/networknext/backend/modules/common"
	"github.com/stretchr/testify/assert"
	"testing"
)

// ------------------------------------------------------------------------

func TestVersionCompare(t *testing.T) {

	t.Parallel()

	t.Run("equal", func(t *testing.T) {
		a := SDKVersion{1, 2, 3}
		b := SDKVersion{1, 2, 3}

		assert.Equal(t, SDKVersionEqual, a.Compare(b))
	})

	t.Run("older", func(t *testing.T) {
		a := SDKVersion{1, 1, 1}
		b := SDKVersion{2, 0, 0}

		assert.Equal(t, SDKVersionOlder, a.Compare(b))

		a = SDKVersion{1, 1, 1}
		b = SDKVersion{1, 2, 0}

		assert.Equal(t, SDKVersionOlder, a.Compare(b))

		a = SDKVersion{1, 1, 1}
		b = SDKVersion{1, 1, 2}

		assert.Equal(t, SDKVersionOlder, a.Compare(b))
	})

	t.Run("newer", func(t *testing.T) {
		a := SDKVersion{1, 1, 1}
		b := SDKVersion{0, 0, 0}

		assert.Equal(t, SDKVersionNewer, a.Compare(b))

		a = SDKVersion{1, 2, 3}
		b = SDKVersion{1, 1, 3}

		assert.Equal(t, SDKVersionNewer, a.Compare(b))

		a = SDKVersion{1, 2, 3}
		b = SDKVersion{1, 2, 2}

		assert.Equal(t, SDKVersionNewer, a.Compare(b))
	})
}

func TestVersionAtLeast(t *testing.T) {

	t.Parallel()

	t.Run("equal", func(t *testing.T) {
		a := SDKVersion{0, 0, 0}
		b := SDKVersion{0, 0, 0}

		assert.True(t, a.AtLeast(b))
	})

	t.Run("newer", func(t *testing.T) {
		a := SDKVersion{0, 0, 1}
		b := SDKVersion{0, 0, 0}

		assert.True(t, a.AtLeast(b))
	})

	t.Run("older", func(t *testing.T) {
		a := SDKVersion{0, 0, 0}
		b := SDKVersion{0, 0, 1}

		assert.False(t, a.AtLeast(b))
	})
}

func PacketSerializationTest[P Packet](writePacket Packet, readPacket Packet, t *testing.T) {

	t.Parallel()

	const BufferSize = 1024

	buffer := [BufferSize]byte{}

	writeStream, err := common.CreateWriteStream(buffer[:])
	assert.Nil(t, err)

	err = writePacket.Serialize(writeStream)
	assert.Nil(t, err)
	writeStream.Flush()

	readStream := common.CreateReadStream(buffer[:])
	err = readPacket.Serialize(readStream)
	assert.Nil(t, err)

	assert.Equal(t, writePacket, readPacket)
}

// ------------------------------------------------------------------------

func Test_SDK4_ServerInitRequestPacket(t *testing.T) {

	writePacket := SDK4_ServerInitRequestPacket{
		Version:        SDKVersion{1,2,3},
		BuyerId:        1234567,
		DatacenterId:   5124111,
		RequestId:      234198347,
		DatacenterName: "test",
	}

	readPacket := SDK4_ServerInitRequestPacket{}

	PacketSerializationTest[*SDK4_ServerInitRequestPacket](&writePacket, &readPacket, t)
}

func Test_SDK4_ServerInitResponsePacket(t *testing.T) {

	writePacket := SDK4_ServerInitResponsePacket{
		RequestId:      234198347,
		Response:       1,
	}

	readPacket := SDK4_ServerInitResponsePacket{}

	PacketSerializationTest[*SDK4_ServerInitResponsePacket](&writePacket, &readPacket, t)
}

func Test_SDK4_SessionUpdatePacket(t *testing.T) {

	writePacket := SDK4_SessionUpdatePacket{
		Version:        SDKVersion{1,2,3},
		// todo
	}

	readPacket := SDK4_SessionUpdatePacket{}

	PacketSerializationTest[*SDK4_SessionUpdatePacket](&writePacket, &readPacket, t)
}

// ------------------------------------------------------------------------

func Test_SDK5_ServerInitRequestPacket(t *testing.T) {

	writePacket := SDK5_ServerInitRequestPacket{
		BuyerId:        1234567,
		DatacenterId:   5124111,
		RequestId:      234198347,
		DatacenterName: "test",
	}

	readPacket := SDK5_ServerInitRequestPacket{}

	PacketSerializationTest[*SDK5_ServerInitRequestPacket](&writePacket, &readPacket, t)
}

// ...

// ------------------------------------------------------------------------
