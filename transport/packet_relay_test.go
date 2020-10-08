package transport_test

import (
	crand "crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"net"
	"testing"
	"time"

	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/encoding"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/transport"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/nacl/box"
)

func TestRelayInitRequestUnmarshalBinary(t *testing.T) {
	t.Parallel()

	t.Run("returns 'invalid packet' when missing magic number", func(t *testing.T) {
		var packet transport.RelayInitRequest
		assert.Equal(t, packet.UnmarshalBinary(make([]byte, 0)), errors.New("invalid packet"))
	})

	t.Run("missing request version", func(t *testing.T) {
		var packet transport.RelayInitRequest
		buff := make([]byte, 4)
		binary.LittleEndian.PutUint32(buff, rand.Uint32()) // can be anything for testing purposes
		assert.Equal(t, packet.UnmarshalBinary(buff), errors.New("invalid packet"))
	})

	t.Run("missing nonce bytes", func(t *testing.T) {
		var packet transport.RelayInitRequest
		buff := make([]byte, 8)
		binary.LittleEndian.PutUint32(buff, rand.Uint32())
		binary.LittleEndian.PutUint32(buff[4:], rand.Uint32())
		assert.Equal(t, packet.UnmarshalBinary(buff), errors.New("invalid packet"))
	})

	t.Run("missing relay address", func(t *testing.T) {
		var packet transport.RelayInitRequest
		buff := make([]byte, 8+crypto.NonceSize)
		binary.LittleEndian.PutUint32(buff, rand.Uint32())
		binary.LittleEndian.PutUint32(buff[4:], rand.Uint32())
		assert.Equal(t, packet.UnmarshalBinary(buff), errors.New("invalid packet"))
	})

	t.Run("missing encryption token", func(t *testing.T) {
		var packet transport.RelayInitRequest
		addr := "127.0.0.1:40000"
		buff := make([]byte, 8+crypto.NonceSize+4+len(addr)) // 4 is the uint32 for address length
		binary.LittleEndian.PutUint32(buff, rand.Uint32())
		binary.LittleEndian.PutUint32(buff[4:], rand.Uint32())
		binary.LittleEndian.PutUint32(buff[8+crypto.NonceSize:], uint32(len(addr)))
		copy(buff[12+crypto.NonceSize:], addr)
		assert.Equal(t, packet.UnmarshalBinary(buff), errors.New("invalid packet"))
	})

	t.Run("address not formatted correctly", func(t *testing.T) {
		var packet transport.RelayInitRequest
		addr := "invalid"
		buff := make([]byte, 8+crypto.NonceSize+4+len(addr)+routing.EncryptedRelayTokenSize)
		binary.LittleEndian.PutUint32(buff, rand.Uint32())
		binary.LittleEndian.PutUint32(buff[4:], rand.Uint32())
		binary.LittleEndian.PutUint32(buff[8+crypto.NonceSize:], uint32(len(addr)))
		copy(buff[12+crypto.NonceSize:], addr)
		assert.Equal(t, packet.UnmarshalBinary(buff), errors.New("could not resolve init packet with address 'invalid' with reason: address invalid: missing port in address"))
	})

	t.Run("valid", func(t *testing.T) {
		var packet transport.RelayInitRequest
		addr := "127.0.0.1:40000"
		buff := make([]byte, 8+crypto.NonceSize+4+len(addr)+routing.EncryptedRelayTokenSize)
		binary.LittleEndian.PutUint32(buff, rand.Uint32())
		binary.LittleEndian.PutUint32(buff[4:], rand.Uint32())
		binary.LittleEndian.PutUint32(buff[8+crypto.NonceSize:], uint32(len(addr)))
		copy(buff[12+crypto.NonceSize:], addr)
		assert.Nil(t, packet.UnmarshalBinary(buff))
	})
}

func TestRelayInitRequestMarshalBinary(t *testing.T) {
	nonce := make([]byte, crypto.NonceSize)
	token := make([]byte, routing.EncryptedRelayTokenSize)
	rand.Read(nonce)
	rand.Read(token)

	udp, _ := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
	expected := transport.RelayInitRequest{
		Magic:          rand.Uint32(),
		Version:        rand.Uint32(),
		Nonce:          nonce,
		Address:        *udp,
		EncryptedToken: token,
	}

	var actual transport.RelayInitRequest

	data, _ := expected.MarshalBinary()

	assert.Nil(t, actual.UnmarshalBinary(data))
	assert.Equal(t, expected, actual)
}

func TestRelayInitResponseMarshalBinary(t *testing.T) {
	res := transport.RelayInitResponse{
		Version:   12345,
		Timestamp: 12345,
		PublicKey: make([]byte, crypto.KeySize),
	}

	binaryRes, err := res.MarshalBinary()
	assert.NoError(t, err)

	expected := []byte{0x0, 0x0, 0x0, 0x0, 0x39, 0x30, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
	assert.Equal(t, expected, binaryRes)
}

func TestRelayInitResponseUnmarshalBinary(t *testing.T) {
	t.Parallel()

	t.Run("bad version", func(t *testing.T) {
		var response transport.RelayInitResponse

		buff := []byte("")

		err := response.UnmarshalBinary(buff)
		assert.EqualError(t, err, "failed to unmarshal relay init response version")
	})

	t.Run("bad timestamp", func(t *testing.T) {
		var response transport.RelayInitResponse

		buff := make([]byte, 5)
		index := 0

		var version uint32 = 0
		encoding.WriteUint32(buff, &index, version)

		err := response.UnmarshalBinary(buff)
		assert.EqualError(t, err, "failed to unmarshal relay init response timestamp")
	})

	t.Run("bad public key", func(t *testing.T) {
		var response transport.RelayInitResponse

		buff := make([]byte, 13)
		index := 0

		var version uint32 = 0
		encoding.WriteUint32(buff, &index, version)

		var timestamp uint64 = uint64(time.Now().Unix())
		encoding.WriteUint64(buff, &index, timestamp)

		err := response.UnmarshalBinary(buff)
		assert.EqualError(t, err, "failed to unmarshal relay init response public key")
	})

	t.Run("valid", func(t *testing.T) {
		var response transport.RelayInitResponse

		buff := make([]byte, 44)
		index := 0

		var version uint32 = 0
		encoding.WriteUint32(buff, &index, version)

		var timestamp uint64 = uint64(time.Now().Unix())
		encoding.WriteUint64(buff, &index, timestamp)

		publicKey, _, err := box.GenerateKey(crand.Reader)
		assert.NoError(t, err)
		encoding.WriteBytes(buff, &index, publicKey[:], crypto.KeySize)

		err = response.UnmarshalBinary(buff)
		assert.NoError(t, err)
	})
}

func TestRelayUpdateRequestUnmarshalBinary(t *testing.T) {
	t.Parallel()

	t.Run("missing request version", func(t *testing.T) {
		var packet transport.RelayUpdateRequest
		assert.Equal(t, errors.New("invalid packet, could not read packet version"), packet.UnmarshalBinary(make([]byte, 0)))
	})

	t.Run("version 0", func(t *testing.T) {
		t.Parallel()

		t.Run("missing relay address", func(t *testing.T) {
			var packet transport.RelayUpdateRequest
			buff := make([]byte, 4)
			binary.LittleEndian.PutUint32(buff, 0) //version
			assert.Equal(t, packet.UnmarshalBinary(buff), errors.New("invalid packet"))
		})

		t.Run("missing relay token", func(t *testing.T) {
			var packet transport.RelayUpdateRequest
			buff := make([]byte, 4+13)
			binary.LittleEndian.PutUint32(buff, 0)
			binary.LittleEndian.PutUint32(buff[4:], 13) // address length
			assert.Equal(t, packet.UnmarshalBinary(buff), errors.New("invalid packet"))
		})

		t.Run("missing number of relays", func(t *testing.T) {
			var packet transport.RelayUpdateRequest
			buff := make([]byte, 4+4+13+crypto.KeySize)
			binary.LittleEndian.PutUint32(buff, 0)
			binary.LittleEndian.PutUint32(buff[4:], 13)
			assert.Equal(t, packet.UnmarshalBinary(buff), errors.New("invalid packet"))
		})

		t.Run("address is not formatted correctly", func(t *testing.T) {
			var packet transport.RelayUpdateRequest
			addr := "invalid"
			buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4)
			binary.LittleEndian.PutUint32(buff, 0)
			binary.LittleEndian.PutUint32(buff[4:], uint32(len(addr)))
			copy(buff[8:], addr)
			binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize:], 1) // number of relays
			assert.Equal(t, packet.UnmarshalBinary(buff), errors.New("could not resolve init packet with address 'invalid' with reason: address invalid: missing port in address"))
		})

		t.Run("missing various relay ping stats", func(t *testing.T) {
			t.Run("missing the id", func(t *testing.T) {
				var packet transport.RelayUpdateRequest
				addr := "127.0.0.1:40000"
				buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4)
				binary.LittleEndian.PutUint32(buff, 0)
				binary.LittleEndian.PutUint32(buff[4:], uint32(len(addr)))
				copy(buff[8:], addr)
				binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize:], 1) // number of relays
				assert.Equal(t, packet.UnmarshalBinary(buff), errors.New("invalid packet, could not read a ping stat"))
			})

			t.Run("missing the rtt", func(t *testing.T) {
				var packet transport.RelayUpdateRequest
				addr := "127.0.0.1:40000"
				buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4+8)
				binary.LittleEndian.PutUint32(buff, 0)
				binary.LittleEndian.PutUint32(buff[4:], uint32(len(addr)))
				copy(buff[8:], addr)
				binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize:], 1)
				binary.LittleEndian.PutUint64(buff[8+len(addr)+crypto.KeySize+4:], rand.Uint64()) // relay id
				assert.Equal(t, packet.UnmarshalBinary(buff), errors.New("invalid packet, could not read a ping stat"))
			})

			t.Run("missing the jitter", func(t *testing.T) {
				var packet transport.RelayUpdateRequest
				addr := "127.0.0.1:40000"
				buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4+8+4)
				binary.LittleEndian.PutUint32(buff, 0)
				binary.LittleEndian.PutUint32(buff[4:], uint32(len(addr)))
				copy(buff[8:], addr)
				binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize:], 1)
				binary.LittleEndian.PutUint64(buff[8+len(addr)+crypto.KeySize+4:], rand.Uint64())
				binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize+12:], math.Float32bits(rand.Float32())) // rtt
				assert.Equal(t, packet.UnmarshalBinary(buff), errors.New("invalid packet, could not read a ping stat"))
			})

			t.Run("missing the packet loss", func(t *testing.T) {
				var packet transport.RelayUpdateRequest
				addr := "127.0.0.1:40000"
				buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4+8+4+4)
				binary.LittleEndian.PutUint32(buff, 0)
				binary.LittleEndian.PutUint32(buff[4:], uint32(len(addr)))
				copy(buff[8:], addr)
				binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize:], 1)
				binary.LittleEndian.PutUint64(buff[8+len(addr)+crypto.KeySize+4:], rand.Uint64())
				binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize+12:], math.Float32bits(rand.Float32()))
				binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize+16:], math.Float32bits(rand.Float32())) // jitter
				assert.Equal(t, packet.UnmarshalBinary(buff), errors.New("invalid packet, could not read a ping stat"))
			})
		})

		t.Run("missing session count", func(t *testing.T) {
			var packet transport.RelayUpdateRequest
			addr := "127.0.0.1:40000"
			buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4+8+4+4+4)
			binary.LittleEndian.PutUint32(buff, 0)
			binary.LittleEndian.PutUint32(buff[4:], uint32(len(addr)))
			copy(buff[8:], addr)
			binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize:], 1)
			binary.LittleEndian.PutUint64(buff[8+len(addr)+crypto.KeySize+4:], rand.Uint64())
			binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize+12:], math.Float32bits(rand.Float32()))
			binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize+16:], math.Float32bits(rand.Float32()))
			binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize+20:], math.Float32bits(rand.Float32())) // packet loss
			assert.Equal(t, packet.UnmarshalBinary(buff), errors.New("invalid packet, could not read session count"))
		})

		t.Run("missing bytes sent", func(t *testing.T) {
			var packet transport.RelayUpdateRequest
			addr := "127.0.0.1:40000"
			buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4+8+4+4+4+8)
			binary.LittleEndian.PutUint32(buff, 0)
			binary.LittleEndian.PutUint32(buff[4:], uint32(len(addr)))
			copy(buff[8:], addr)
			binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize:], 1)
			binary.LittleEndian.PutUint64(buff[8+len(addr)+crypto.KeySize+4:], rand.Uint64())
			binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize+12:], math.Float32bits(rand.Float32()))
			binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize+16:], math.Float32bits(rand.Float32()))
			binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize+20:], math.Float32bits(rand.Float32()))
			binary.LittleEndian.PutUint64(buff[8+len(addr)+crypto.KeySize+24:], rand.Uint64()) // session count
			assert.Equal(t, packet.UnmarshalBinary(buff), errors.New("invalid packet, could not read bytes sent"))
		})

		t.Run("missing bytes received", func(t *testing.T) {
			var packet transport.RelayUpdateRequest
			addr := "127.0.0.1:40000"
			buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4+8+4+4+4+8+8)
			binary.LittleEndian.PutUint32(buff, 0)
			binary.LittleEndian.PutUint32(buff[4:], uint32(len(addr)))
			copy(buff[8:], addr)
			binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize:], 1)
			binary.LittleEndian.PutUint64(buff[8+len(addr)+crypto.KeySize+4:], rand.Uint64())
			binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize+12:], math.Float32bits(rand.Float32()))
			binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize+16:], math.Float32bits(rand.Float32()))
			binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize+20:], math.Float32bits(rand.Float32()))
			binary.LittleEndian.PutUint64(buff[8+len(addr)+crypto.KeySize+24:], rand.Uint64())
			binary.LittleEndian.PutUint64(buff[8+len(addr)+crypto.KeySize+32:], rand.Uint64()) // bytes sent
			assert.Equal(t, packet.UnmarshalBinary(buff), errors.New("invalid packet, could not read bytes received"))
		})

		t.Run("missing shutdown flag", func(t *testing.T) {
			var packet transport.RelayUpdateRequest
			addr := "127.0.0.1:40000"
			buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4+8+4+4+4+8+8+8)
			binary.LittleEndian.PutUint32(buff, 0)
			binary.LittleEndian.PutUint32(buff[4:], uint32(len(addr)))
			copy(buff[8:], addr)
			binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize:], 1)
			binary.LittleEndian.PutUint64(buff[8+len(addr)+crypto.KeySize+4:], rand.Uint64())
			binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize+12:], math.Float32bits(rand.Float32()))
			binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize+16:], math.Float32bits(rand.Float32()))
			binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize+20:], math.Float32bits(rand.Float32()))
			binary.LittleEndian.PutUint64(buff[8+len(addr)+crypto.KeySize+24:], rand.Uint64())
			binary.LittleEndian.PutUint64(buff[8+len(addr)+crypto.KeySize+32:], rand.Uint64())
			binary.LittleEndian.PutUint64(buff[8+len(addr)+crypto.KeySize+40:], rand.Uint64()) // bytes received
			assert.Equal(t, packet.UnmarshalBinary(buff), errors.New("invalid packet, could not read shutdown flag"))
		})

		t.Run("missing cpu usage", func(t *testing.T) {
			var packet transport.RelayUpdateRequest
			packet.ShuttingDown = true
			addr := "127.0.0.1:40000"
			buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4+8+4+4+4+8+8+8+1)
			binary.LittleEndian.PutUint32(buff, 0)
			binary.LittleEndian.PutUint32(buff[4:], uint32(len(addr)))
			copy(buff[8:], addr)
			binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize:], 1)
			binary.LittleEndian.PutUint64(buff[8+len(addr)+crypto.KeySize+4:], rand.Uint64())
			binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize+12:], math.Float32bits(rand.Float32()))
			binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize+16:], math.Float32bits(rand.Float32()))
			binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize+20:], math.Float32bits(rand.Float32()))
			binary.LittleEndian.PutUint64(buff[8+len(addr)+crypto.KeySize+24:], rand.Uint64())
			binary.LittleEndian.PutUint64(buff[8+len(addr)+crypto.KeySize+32:], rand.Uint64())
			binary.LittleEndian.PutUint64(buff[8+len(addr)+crypto.KeySize+40:], rand.Uint64())
			buff[8+len(addr)+crypto.KeySize+48] = 1 // shutdown flag
			assert.Equal(t, packet.UnmarshalBinary(buff), errors.New("invalid packet, could not read cpu usage"))
		})

		t.Run("missing memory usage", func(t *testing.T) {
			var packet transport.RelayUpdateRequest
			packet.ShuttingDown = true
			addr := "127.0.0.1:40000"
			buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4+8+4+4+4+8+8+8+1+8)
			binary.LittleEndian.PutUint32(buff, 0)
			binary.LittleEndian.PutUint32(buff[4:], uint32(len(addr)))
			copy(buff[8:], addr)
			binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize:], 1)
			binary.LittleEndian.PutUint64(buff[8+len(addr)+crypto.KeySize+4:], rand.Uint64())
			binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize+12:], math.Float32bits(rand.Float32()))
			binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize+16:], math.Float32bits(rand.Float32()))
			binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize+20:], math.Float32bits(rand.Float32()))
			binary.LittleEndian.PutUint64(buff[8+len(addr)+crypto.KeySize+24:], rand.Uint64())
			binary.LittleEndian.PutUint64(buff[8+len(addr)+crypto.KeySize+32:], rand.Uint64())
			binary.LittleEndian.PutUint64(buff[8+len(addr)+crypto.KeySize+40:], rand.Uint64())
			buff[8+len(addr)+crypto.KeySize+48] = 1
			binary.LittleEndian.PutUint64(buff[8+len(addr)+crypto.KeySize+49:], math.Float64bits(rand.Float64())) // cpu usage
			assert.Equal(t, packet.UnmarshalBinary(buff), errors.New("invalid packet, could not read memory usage"))
		})

		t.Run("missing relay version", func(t *testing.T) {
			var packet transport.RelayUpdateRequest
			packet.ShuttingDown = true
			addr := "127.0.0.1:40000"
			buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4+8+4+4+4+8+8+8+1+8+8)
			binary.LittleEndian.PutUint32(buff, 0)
			binary.LittleEndian.PutUint32(buff[4:], uint32(len(addr)))
			copy(buff[8:], addr)
			binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize:], 1)
			binary.LittleEndian.PutUint64(buff[8+len(addr)+crypto.KeySize+4:], rand.Uint64())
			binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize+12:], math.Float32bits(rand.Float32()))
			binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize+16:], math.Float32bits(rand.Float32()))
			binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize+20:], math.Float32bits(rand.Float32()))
			binary.LittleEndian.PutUint64(buff[8+len(addr)+crypto.KeySize+24:], rand.Uint64())
			binary.LittleEndian.PutUint64(buff[8+len(addr)+crypto.KeySize+32:], rand.Uint64())
			binary.LittleEndian.PutUint64(buff[8+len(addr)+crypto.KeySize+40:], rand.Uint64())
			buff[8+len(addr)+crypto.KeySize+48] = 1
			binary.LittleEndian.PutUint64(buff[8+len(addr)+crypto.KeySize+49:], math.Float64bits(rand.Float64()))
			binary.LittleEndian.PutUint64(buff[8+len(addr)+crypto.KeySize+57:], math.Float64bits(rand.Float64())) // memory usage
			assert.Equal(t, packet.UnmarshalBinary(buff), errors.New("invalid packet, could not read relay version"))
		})

		t.Run("valid", func(t *testing.T) {
			var packet transport.RelayUpdateRequest
			packet.ShuttingDown = true
			addr := "127.0.0.1:40000"
			buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4+8+4+4+4+8+8+8+1+8+8+4+len("1.0.0"))
			binary.LittleEndian.PutUint32(buff, 0)
			binary.LittleEndian.PutUint32(buff[4:], uint32(len(addr)))
			copy(buff[8:], addr)
			binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize:], 1)
			binary.LittleEndian.PutUint64(buff[8+len(addr)+crypto.KeySize+4:], rand.Uint64())
			binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize+12:], math.Float32bits(rand.Float32()))
			binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize+16:], math.Float32bits(rand.Float32()))
			binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize+20:], math.Float32bits(rand.Float32()))
			binary.LittleEndian.PutUint64(buff[8+len(addr)+crypto.KeySize+24:], rand.Uint64())
			binary.LittleEndian.PutUint64(buff[8+len(addr)+crypto.KeySize+32:], rand.Uint64())
			binary.LittleEndian.PutUint64(buff[8+len(addr)+crypto.KeySize+40:], rand.Uint64())
			buff[8+len(addr)+crypto.KeySize+48] = 1
			binary.LittleEndian.PutUint64(buff[8+len(addr)+crypto.KeySize+49:], math.Float64bits(rand.Float64()))
			binary.LittleEndian.PutUint64(buff[8+len(addr)+crypto.KeySize+57:], math.Float64bits(rand.Float64()))
			binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize+65:], uint32(len("1.0.0"))) // relay version
			copy(buff[8+len(addr)+crypto.KeySize+69:], "1.0.0")
			assert.Nil(t, packet.UnmarshalBinary(buff))
		})
	})

	t.Run("version 1", func(t *testing.T) {
		t.Parallel()

		t.Run("missing relay address", func(t *testing.T) {
			var packet transport.RelayUpdateRequest
			buff := make([]byte, 4)
			index := 0
			encoding.WriteUint32(buff, &index, 1)
			assert.Equal(t, errors.New("invalid packet"), packet.UnmarshalBinary(buff))
		})

		t.Run("missing relay token", func(t *testing.T) {
			var packet transport.RelayUpdateRequest
			addr := "invalid"
			buff := make([]byte, 4+4+len(addr))
			index := 0
			encoding.WriteUint32(buff, &index, 1)
			encoding.WriteString(buff, &index, addr, uint32(len(addr)))
			assert.Equal(t, errors.New("invalid packet"), packet.UnmarshalBinary(buff))
		})

		t.Run("missing number of relays", func(t *testing.T) {
			var packet transport.RelayUpdateRequest
			addr := "invalid"
			buff := make([]byte, 4+4+len(addr)+crypto.KeySize)
			index := 0
			encoding.WriteUint32(buff, &index, 1)
			encoding.WriteString(buff, &index, addr, uint32(len(addr)))
			index += crypto.KeySize
			assert.Equal(t, errors.New("invalid packet"), packet.UnmarshalBinary(buff))
		})

		t.Run("address is not formatted correctly", func(t *testing.T) {
			var packet transport.RelayUpdateRequest
			addr := "invalid"
			buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4)
			index := 0
			encoding.WriteUint32(buff, &index, 1)
			encoding.WriteString(buff, &index, addr, uint32(len(addr)))
			index += crypto.KeySize
			encoding.WriteUint32(buff, &index, 1)
			assert.Equal(t, errors.New("could not resolve init packet with address 'invalid' with reason: address invalid: missing port in address"), packet.UnmarshalBinary(buff))
		})

		t.Run("missing various relay ping stats", func(t *testing.T) {
			t.Run("missing the id", func(t *testing.T) {
				var packet transport.RelayUpdateRequest
				addr := "127.0.0.1:40000"
				buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4)
				index := 0
				encoding.WriteUint32(buff, &index, 1)
				encoding.WriteString(buff, &index, addr, uint32(len(addr)))
				index += crypto.KeySize
				encoding.WriteUint32(buff, &index, 1)
				assert.Equal(t, errors.New("invalid packet, could not read a ping stat"), packet.UnmarshalBinary(buff))
			})

			t.Run("missing the rtt", func(t *testing.T) {
				var packet transport.RelayUpdateRequest
				addr := "127.0.0.1:40000"
				buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4+8)
				index := 0
				encoding.WriteUint32(buff, &index, 1)
				encoding.WriteString(buff, &index, addr, uint32(len(addr)))
				index += crypto.KeySize
				encoding.WriteUint32(buff, &index, 1)
				assert.Equal(t, errors.New("invalid packet, could not read a ping stat"), packet.UnmarshalBinary(buff))
			})

			t.Run("missing the jitter", func(t *testing.T) {
				var packet transport.RelayUpdateRequest
				addr := "127.0.0.1:40000"
				buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4+8+4)
				index := 0
				encoding.WriteUint32(buff, &index, 1)
				encoding.WriteString(buff, &index, addr, uint32(len(addr)))
				index += crypto.KeySize
				encoding.WriteUint32(buff, &index, 1)
				assert.Equal(t, errors.New("invalid packet, could not read a ping stat"), packet.UnmarshalBinary(buff))
			})

			t.Run("missing the packet loss", func(t *testing.T) {
				var packet transport.RelayUpdateRequest
				addr := "127.0.0.1:40000"
				buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4+8+4+4)
				index := 0
				encoding.WriteUint32(buff, &index, 1)
				encoding.WriteString(buff, &index, addr, uint32(len(addr)))
				index += crypto.KeySize
				encoding.WriteUint32(buff, &index, 1)
				assert.Equal(t, errors.New("invalid packet, could not read a ping stat"), packet.UnmarshalBinary(buff))
			})
		})

		t.Run("missing session count", func(t *testing.T) {
			var packet transport.RelayUpdateRequest
			addr := "127.0.0.1:40000"
			buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4+8+4+4+4)
			index := 0
			encoding.WriteUint32(buff, &index, 1)
			encoding.WriteString(buff, &index, addr, uint32(len(addr)))
			index += crypto.KeySize
			encoding.WriteUint32(buff, &index, 1)
			assert.Equal(t, errors.New("invalid packet, could not read session count"), packet.UnmarshalBinary(buff))
		})

		t.Run("missing outbound ping tx", func(t *testing.T) {
			var packet transport.RelayUpdateRequest
			addr := "127.0.0.1:40000"
			buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4+8+4+4+4+8)
			index := 0
			encoding.WriteUint32(buff, &index, 1)
			encoding.WriteString(buff, &index, addr, uint32(len(addr)))
			index += crypto.KeySize
			encoding.WriteUint32(buff, &index, 1)
			assert.Equal(t, errors.New("invalid packet, could not read outbound ping tx"), packet.UnmarshalBinary(buff))
		})

		t.Run("missing route request rx", func(t *testing.T) {
			var packet transport.RelayUpdateRequest
			addr := "127.0.0.1:40000"
			buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4+8+4+4+4+8+8*1)
			index := 0
			encoding.WriteUint32(buff, &index, 1)
			encoding.WriteString(buff, &index, addr, uint32(len(addr)))
			index += crypto.KeySize
			encoding.WriteUint32(buff, &index, 1)
			assert.Equal(t, errors.New("invalid packet, could not read route request rx"), packet.UnmarshalBinary(buff))
		})

		t.Run("missing route request tx", func(t *testing.T) {
			var packet transport.RelayUpdateRequest
			addr := "127.0.0.1:40000"
			buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4+8+4+4+4+8+8*2)
			index := 0
			encoding.WriteUint32(buff, &index, 1)
			encoding.WriteString(buff, &index, addr, uint32(len(addr)))
			index += crypto.KeySize
			encoding.WriteUint32(buff, &index, 1)
			assert.Equal(t, errors.New("invalid packet, could not read route request tx"), packet.UnmarshalBinary(buff))
		})

		t.Run("missing route response rx", func(t *testing.T) {
			var packet transport.RelayUpdateRequest
			addr := "127.0.0.1:40000"
			buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4+8+4+4+4+8+8*3)
			index := 0
			encoding.WriteUint32(buff, &index, 1)
			encoding.WriteString(buff, &index, addr, uint32(len(addr)))
			index += crypto.KeySize
			encoding.WriteUint32(buff, &index, 1)
			assert.Equal(t, errors.New("invalid packet, could not read route response rx"), packet.UnmarshalBinary(buff))
		})

		t.Run("missing route response tx", func(t *testing.T) {
			var packet transport.RelayUpdateRequest
			addr := "127.0.0.1:40000"
			buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4+8+4+4+4+8+8*4)
			index := 0
			encoding.WriteUint32(buff, &index, 1)
			encoding.WriteString(buff, &index, addr, uint32(len(addr)))
			index += crypto.KeySize
			encoding.WriteUint32(buff, &index, 1)
			assert.Equal(t, errors.New("invalid packet, could not read route response tx"), packet.UnmarshalBinary(buff))
		})

		t.Run("missing client to server rx", func(t *testing.T) {
			var packet transport.RelayUpdateRequest
			addr := "127.0.0.1:40000"
			buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4+8+4+4+4+8+8*5)
			index := 0
			encoding.WriteUint32(buff, &index, 1)
			encoding.WriteString(buff, &index, addr, uint32(len(addr)))
			index += crypto.KeySize
			encoding.WriteUint32(buff, &index, 1)
			assert.Equal(t, errors.New("invalid packet, could not read client to server rx"), packet.UnmarshalBinary(buff))
		})

		t.Run("missing client to server tx", func(t *testing.T) {
			var packet transport.RelayUpdateRequest
			addr := "127.0.0.1:40000"
			buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4+8+4+4+4+8+8*6)
			index := 0
			encoding.WriteUint32(buff, &index, 1)
			encoding.WriteString(buff, &index, addr, uint32(len(addr)))
			index += crypto.KeySize
			encoding.WriteUint32(buff, &index, 1)
			assert.Equal(t, errors.New("invalid packet, could not read client to server tx"), packet.UnmarshalBinary(buff))
		})

		t.Run("missing server to client rx", func(t *testing.T) {
			var packet transport.RelayUpdateRequest
			addr := "127.0.0.1:40000"
			buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4+8+4+4+4+8+8*7)
			index := 0
			encoding.WriteUint32(buff, &index, 1)
			encoding.WriteString(buff, &index, addr, uint32(len(addr)))
			index += crypto.KeySize
			encoding.WriteUint32(buff, &index, 1)
			assert.Equal(t, errors.New("invalid packet, could not read server to client rx"), packet.UnmarshalBinary(buff))
		})

		t.Run("missing server to client tx", func(t *testing.T) {
			var packet transport.RelayUpdateRequest
			addr := "127.0.0.1:40000"
			buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4+8+4+4+4+8+8*8)
			index := 0
			encoding.WriteUint32(buff, &index, 1)
			encoding.WriteString(buff, &index, addr, uint32(len(addr)))
			index += crypto.KeySize
			encoding.WriteUint32(buff, &index, 1)
			assert.Equal(t, errors.New("invalid packet, could not read server to client tx"), packet.UnmarshalBinary(buff))
		})

		t.Run("missing inbound ping rx", func(t *testing.T) {
			var packet transport.RelayUpdateRequest
			addr := "127.0.0.1:40000"
			buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4+8+4+4+4+8+8*9)
			index := 0
			encoding.WriteUint32(buff, &index, 1)
			encoding.WriteString(buff, &index, addr, uint32(len(addr)))
			index += crypto.KeySize
			encoding.WriteUint32(buff, &index, 1)
			assert.Equal(t, errors.New("invalid packet, could not read inbound ping rx"), packet.UnmarshalBinary(buff))
		})

		t.Run("missing inbound ping tx", func(t *testing.T) {
			var packet transport.RelayUpdateRequest
			addr := "127.0.0.1:40000"
			buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4+8+4+4+4+8+8*10)
			index := 0
			encoding.WriteUint32(buff, &index, 1)
			encoding.WriteString(buff, &index, addr, uint32(len(addr)))
			index += crypto.KeySize
			encoding.WriteUint32(buff, &index, 1)
			assert.Equal(t, errors.New("invalid packet, could not read inbound ping tx"), packet.UnmarshalBinary(buff))
		})

		t.Run("missing pong rx", func(t *testing.T) {
			var packet transport.RelayUpdateRequest
			addr := "127.0.0.1:40000"
			buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4+8+4+4+4+8+8*11)
			index := 0
			encoding.WriteUint32(buff, &index, 1)
			encoding.WriteString(buff, &index, addr, uint32(len(addr)))
			index += crypto.KeySize
			encoding.WriteUint32(buff, &index, 1)
			assert.Equal(t, errors.New("invalid packet, could not read pong rx"), packet.UnmarshalBinary(buff))
		})

		t.Run("missing session ping rx", func(t *testing.T) {
			var packet transport.RelayUpdateRequest
			addr := "127.0.0.1:40000"
			buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4+8+4+4+4+8+8*12)
			index := 0
			encoding.WriteUint32(buff, &index, 1)
			encoding.WriteString(buff, &index, addr, uint32(len(addr)))
			index += crypto.KeySize
			encoding.WriteUint32(buff, &index, 1)
			assert.Equal(t, errors.New("invalid packet, could not read session ping rx"), packet.UnmarshalBinary(buff))
		})

		t.Run("missing session ping tx", func(t *testing.T) {
			var packet transport.RelayUpdateRequest
			addr := "127.0.0.1:40000"
			buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4+8+4+4+4+8+8*13)
			index := 0
			encoding.WriteUint32(buff, &index, 1)
			encoding.WriteString(buff, &index, addr, uint32(len(addr)))
			index += crypto.KeySize
			encoding.WriteUint32(buff, &index, 1)
			assert.Equal(t, errors.New("invalid packet, could not read session ping tx"), packet.UnmarshalBinary(buff))
		})

		t.Run("missing session pong rx", func(t *testing.T) {
			var packet transport.RelayUpdateRequest
			addr := "127.0.0.1:40000"
			buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4+8+4+4+4+8+8*14)
			index := 0
			encoding.WriteUint32(buff, &index, 1)
			encoding.WriteString(buff, &index, addr, uint32(len(addr)))
			index += crypto.KeySize
			encoding.WriteUint32(buff, &index, 1)
			assert.Equal(t, errors.New("invalid packet, could not read session pong rx"), packet.UnmarshalBinary(buff))
		})

		t.Run("missing session pong tx", func(t *testing.T) {
			var packet transport.RelayUpdateRequest
			addr := "127.0.0.1:40000"
			buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4+8+4+4+4+8+8*15)
			index := 0
			encoding.WriteUint32(buff, &index, 1)
			encoding.WriteString(buff, &index, addr, uint32(len(addr)))
			index += crypto.KeySize
			encoding.WriteUint32(buff, &index, 1)
			assert.Equal(t, errors.New("invalid packet, could not read session pong tx"), packet.UnmarshalBinary(buff))
		})

		t.Run("missing continue request rx", func(t *testing.T) {
			var packet transport.RelayUpdateRequest
			addr := "127.0.0.1:40000"
			buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4+8+4+4+4+8+8*16)
			index := 0
			encoding.WriteUint32(buff, &index, 1)
			encoding.WriteString(buff, &index, addr, uint32(len(addr)))
			index += crypto.KeySize
			encoding.WriteUint32(buff, &index, 1)
			assert.Equal(t, errors.New("invalid packet, could not read continue request rx"), packet.UnmarshalBinary(buff))
		})

		t.Run("missing continue request tx", func(t *testing.T) {
			var packet transport.RelayUpdateRequest
			addr := "127.0.0.1:40000"
			buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4+8+4+4+4+8+8*17)
			index := 0
			encoding.WriteUint32(buff, &index, 1)
			encoding.WriteString(buff, &index, addr, uint32(len(addr)))
			index += crypto.KeySize
			encoding.WriteUint32(buff, &index, 1)
			assert.Equal(t, errors.New("invalid packet, could not read continue request tx"), packet.UnmarshalBinary(buff))
		})

		t.Run("missing continue response rx", func(t *testing.T) {
			var packet transport.RelayUpdateRequest
			addr := "127.0.0.1:40000"
			buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4+8+4+4+4+8+8*18)
			index := 0
			encoding.WriteUint32(buff, &index, 1)
			encoding.WriteString(buff, &index, addr, uint32(len(addr)))
			index += crypto.KeySize
			encoding.WriteUint32(buff, &index, 1)
			assert.Equal(t, errors.New("invalid packet, could not read continue response rx"), packet.UnmarshalBinary(buff))
		})

		t.Run("missing continue response tx", func(t *testing.T) {
			var packet transport.RelayUpdateRequest
			addr := "127.0.0.1:40000"
			buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4+8+4+4+4+8+8*19)
			index := 0
			encoding.WriteUint32(buff, &index, 1)
			encoding.WriteString(buff, &index, addr, uint32(len(addr)))
			index += crypto.KeySize
			encoding.WriteUint32(buff, &index, 1)
			assert.Equal(t, errors.New("invalid packet, could not read continue response tx"), packet.UnmarshalBinary(buff))
		})

		t.Run("missing near ping rx", func(t *testing.T) {
			var packet transport.RelayUpdateRequest
			addr := "127.0.0.1:40000"
			buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4+8+4+4+4+8+8*20)
			index := 0
			encoding.WriteUint32(buff, &index, 1)
			encoding.WriteString(buff, &index, addr, uint32(len(addr)))
			index += crypto.KeySize
			encoding.WriteUint32(buff, &index, 1)
			assert.Equal(t, errors.New("invalid packet, could not read near ping rx"), packet.UnmarshalBinary(buff))
		})

		t.Run("missing near ping tx", func(t *testing.T) {
			var packet transport.RelayUpdateRequest
			addr := "127.0.0.1:40000"
			buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4+8+4+4+4+8+8*21)
			index := 0
			encoding.WriteUint32(buff, &index, 1)
			encoding.WriteString(buff, &index, addr, uint32(len(addr)))
			index += crypto.KeySize
			encoding.WriteUint32(buff, &index, 1)
			assert.Equal(t, errors.New("invalid packet, could not read near ping tx"), packet.UnmarshalBinary(buff))
		})

		t.Run("missing near unknown rx", func(t *testing.T) {
			var packet transport.RelayUpdateRequest
			addr := "127.0.0.1:40000"
			buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4+8+4+4+4+8+8*22)
			index := 0
			encoding.WriteUint32(buff, &index, 1)
			encoding.WriteString(buff, &index, addr, uint32(len(addr)))
			index += crypto.KeySize
			encoding.WriteUint32(buff, &index, 1)
			assert.Equal(t, errors.New("invalid packet, could not read unknown rx"), packet.UnmarshalBinary(buff))
		})

		t.Run("missing shutdown flag", func(t *testing.T) {
			var packet transport.RelayUpdateRequest
			addr := "127.0.0.1:40000"
			buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4+8+4+4+4+8+8*23)
			index := 0
			encoding.WriteUint32(buff, &index, 1)
			encoding.WriteString(buff, &index, addr, uint32(len(addr)))
			index += crypto.KeySize
			encoding.WriteUint32(buff, &index, 1)
			assert.Equal(t, errors.New("invalid packet, could not read shutdown flag"), packet.UnmarshalBinary(buff))
		})

		t.Run("missing cpu usage", func(t *testing.T) {
			var packet transport.RelayUpdateRequest
			addr := "127.0.0.1:40000"
			buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4+8+4+4+4+8+8*23+1)
			index := 0
			encoding.WriteUint32(buff, &index, 1)
			encoding.WriteString(buff, &index, addr, uint32(len(addr)))
			index += crypto.KeySize
			encoding.WriteUint32(buff, &index, 1)
			assert.Equal(t, errors.New("invalid packet, could not read cpu usage"), packet.UnmarshalBinary(buff))
		})

		t.Run("missing memory usage", func(t *testing.T) {
			var packet transport.RelayUpdateRequest
			addr := "127.0.0.1:40000"
			buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4+8+4+4+4+8+8*23+1+8)
			index := 0
			encoding.WriteUint32(buff, &index, 1)
			encoding.WriteString(buff, &index, addr, uint32(len(addr)))
			index += crypto.KeySize
			encoding.WriteUint32(buff, &index, 1)
			assert.Equal(t, errors.New("invalid packet, could not read memory usage"), packet.UnmarshalBinary(buff))
		})

		t.Run("missing relay version", func(t *testing.T) {
			var packet transport.RelayUpdateRequest
			addr := "127.0.0.1:40000"
			buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4+8+4+4+4+8+8*23+1+8+8)
			index := 0
			encoding.WriteUint32(buff, &index, 1)
			encoding.WriteString(buff, &index, addr, uint32(len(addr)))
			index += crypto.KeySize
			encoding.WriteUint32(buff, &index, 1)
			assert.Equal(t, errors.New("invalid packet, could not read relay version"), packet.UnmarshalBinary(buff))
		})

		t.Run("valid", func(t *testing.T) {
			var packet transport.RelayUpdateRequest
			addr := "127.0.0.1:40000"
			buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4+8+4+4+4+8+8*23+1+8+8+4+len("1.0.0"))
			index := 0
			encoding.WriteUint32(buff, &index, 1)
			encoding.WriteString(buff, &index, addr, uint32(len(addr)))
			index += crypto.KeySize
			encoding.WriteUint32(buff, &index, 1)
			index += 8 + 4 + 4 + 4 + 8 + 8*23 + 1 + 8 + 8
			encoding.WriteString(buff, &index, "1.0.0", 5)
			assert.NoError(t, packet.UnmarshalBinary(buff))
		})
	})
}

func TestRelayUpdateRequestMarshalBinary(t *testing.T) {
	t.Parallel()

	t.Run("version 0", func(t *testing.T) {
		stats := make([]routing.RelayStatsPing, 5)

		for i := 0; i < 5; i++ {
			stat := &stats[i]
			stat.RelayID = rand.Uint64()
			stat.RTT = rand.Float32()
			stat.Jitter = rand.Float32()
			stat.PacketLoss = rand.Float32()
		}

		token := make([]byte, crypto.KeySize)
		rand.Read(token)

		udp, _ := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
		expected := transport.RelayUpdateRequest{
			Version:      0,
			Address:      *udp,
			Token:        token,
			PingStats:    stats,
			CPUUsage:     10.0,
			MemUsage:     20.0,
			RelayVersion: "1.0.0",
			TrafficStats: routing.TrafficStats{
				SessionCount:  1,
				BytesSent:     10,
				BytesReceived: 20,
			},
		}

		data, err := expected.MarshalBinary()
		assert.NoError(t, err)

		var actual transport.RelayUpdateRequest
		err = actual.UnmarshalBinary(data)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("version 1", func(t *testing.T) {
		stats := make([]routing.RelayStatsPing, 5)

		for i := 0; i < 5; i++ {
			stat := &stats[i]
			stat.RelayID = rand.Uint64()
			stat.RTT = rand.Float32()
			stat.Jitter = rand.Float32()
			stat.PacketLoss = rand.Float32()
		}

		token := make([]byte, crypto.KeySize)
		rand.Read(token)

		udp, _ := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
		expected := transport.RelayUpdateRequest{
			Version:      1,
			Address:      *udp,
			Token:        token,
			PingStats:    stats,
			CPUUsage:     10.0,
			MemUsage:     20.0,
			RelayVersion: "1.0.0",
			TrafficStats: routing.TrafficStats{
				SessionCount:       1,
				OutboundPingTx:     2,
				RouteRequestRx:     3,
				RouteRequestTx:     4,
				RouteResponseRx:    5,
				RouteResponseTx:    6,
				ClientToServerRx:   7,
				ClientToServerTx:   8,
				ServerToClientRx:   9,
				ServerToClientTx:   10,
				InboundPingRx:      11,
				InboundPingTx:      12,
				PongRx:             13,
				SessionPingRx:      14,
				SessionPingTx:      15,
				SessionPongRx:      16,
				SessionPongTx:      17,
				ContinueRequestRx:  18,
				ContinueRequestTx:  19,
				ContinueResponseRx: 20,
				ContinueResponseTx: 21,
				NearPingRx:         22,
				NearPingTx:         23,
				UnknownRx:          24,
				BytesSent:          137,
				BytesReceived:      162,
			},
		}

		data, err := expected.MarshalBinary()
		assert.NoError(t, err)

		var actual transport.RelayUpdateRequest
		err = actual.UnmarshalBinary(data)
		assert.NoError(t, err)
		assert.Equal(t, expected, actual)
	})

	t.Run("bad version", func(t *testing.T) {
		expected := transport.RelayUpdateRequest{
			Version: transport.VersionNumberUpdateRequest + 1,
		}

		data, err := expected.MarshalBinary()
		assert.Nil(t, data)
		assert.Errorf(t, err, fmt.Sprintf("invalid update request version: %d", transport.VersionNumberUpdateRequest+1))
	})
}

func TestRelayUpdateResponseUnmarshalBinary(t *testing.T) {
	t.Parallel()

	t.Run("missing response version", func(t *testing.T) {
		var packet transport.RelayUpdateResponse
		buff := make([]byte, 0)
		assert.EqualError(t, packet.UnmarshalBinary(buff), "failed to unmarshal relay update response version")
	})

	t.Run("missing response timestamp", func(t *testing.T) {
		var packet transport.RelayUpdateResponse
		index := 0
		buff := make([]byte, 4)
		encoding.WriteUint32(buff, &index, rand.Uint32()) //version
		assert.EqualError(t, packet.UnmarshalBinary(buff), "failed to unmarshal relay update response timestamp")
	})

	t.Run("missing response number of relays to ping", func(t *testing.T) {
		var packet transport.RelayUpdateResponse
		index := 0
		buff := make([]byte, 4+8)
		encoding.WriteUint32(buff, &index, rand.Uint32())
		encoding.WriteUint64(buff, &index, rand.Uint64()) //timestamp
		assert.EqualError(t, packet.UnmarshalBinary(buff), "failed to unmarshal relay update response number of relays to ping")
	})

	t.Run("missing response relay id", func(t *testing.T) {
		var packet transport.RelayUpdateResponse
		index := 0
		buff := make([]byte, 4+8+4)
		encoding.WriteUint32(buff, &index, rand.Uint32()) //version
		encoding.WriteUint64(buff, &index, rand.Uint64()) //timestamp
		encoding.WriteUint32(buff, &index, 1)             //numRelaysToPing
		assert.EqualError(t, packet.UnmarshalBinary(buff), "failed to unmarshal relay update response relay id")
	})

	t.Run("missing response relay address", func(t *testing.T) {
		var packet transport.RelayUpdateResponse
		index := 0
		buff := make([]byte, 4+8+4+8)
		encoding.WriteUint32(buff, &index, rand.Uint32()) //version
		encoding.WriteUint64(buff, &index, rand.Uint64()) //timestamp
		encoding.WriteUint32(buff, &index, 1)             //numRelaysToPing
		encoding.WriteUint64(buff, &index, rand.Uint64()) //id
		assert.EqualError(t, packet.UnmarshalBinary(buff), "failed to unmarshal relay update response relay address")
	})

	t.Run("valid", func(t *testing.T) {
		var packet transport.RelayUpdateResponse
		addr := "127.0.0.1:40000"
		index := 0
		buff := make([]byte, 4+8+4+8+(4+len(addr)))
		encoding.WriteUint32(buff, &index, rand.Uint32())                       //version
		encoding.WriteUint64(buff, &index, rand.Uint64())                       //timestamp
		encoding.WriteUint32(buff, &index, 1)                                   //numRelaysToPing
		encoding.WriteUint64(buff, &index, rand.Uint64())                       //id
		encoding.WriteString(buff, &index, addr, routing.MaxRelayAddressLength) //address
		assert.NoError(t, packet.UnmarshalBinary(buff))
	})
}

func TestRelayUpdateResponseMarshalBinary(t *testing.T) {
	addr := "127.0.0.1:40000"

	expected := transport.RelayUpdateResponse{
		RelaysToPing: []routing.RelayPingData{
			{
				ID:      crypto.HashID(addr),
				Address: addr,
			},
		},
	}

	data, err := expected.MarshalBinary()
	assert.NoError(t, err)

	var actual transport.RelayUpdateResponse
	err = actual.UnmarshalBinary(data)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}
