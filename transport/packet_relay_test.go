package transport_test

import (
	"encoding/binary"
	"errors"
	"math"
	"math/rand"
	"net"
	"testing"

	"github.com/networknext/backend/core"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/transport"
	"github.com/stretchr/testify/assert"
)

func TestRelayInitPacket(t *testing.T) {
	t.Run("UnmarshalBinary()", func(t *testing.T) {
		t.Run("returns 'invalid packet' when missing magic number", func(t *testing.T) {
			var packet transport.RelayInitPacket
			assert.Equal(t, packet.UnmarshalBinary(make([]byte, 0)), errors.New("invalid packet"))
		})

		t.Run("missing request version", func(t *testing.T) {
			var packet transport.RelayInitPacket
			buff := make([]byte, 4)
			binary.LittleEndian.PutUint32(buff, rand.Uint32()) // can be anything for testing purposes
			assert.Equal(t, packet.UnmarshalBinary(buff), errors.New("invalid packet"))
		})

		t.Run("missing nonce bytes", func(t *testing.T) {
			var packet transport.RelayInitPacket
			buff := make([]byte, 8)
			binary.LittleEndian.PutUint32(buff, rand.Uint32())
			binary.LittleEndian.PutUint32(buff[4:], rand.Uint32())
			assert.Equal(t, packet.UnmarshalBinary(buff), errors.New("invalid packet"))
		})

		t.Run("missing relay address", func(t *testing.T) {
			var packet transport.RelayInitPacket
			buff := make([]byte, 8+crypto.NonceSize)
			binary.LittleEndian.PutUint32(buff, rand.Uint32())
			binary.LittleEndian.PutUint32(buff[4:], rand.Uint32())
			assert.Equal(t, packet.UnmarshalBinary(buff), errors.New("invalid packet"))
		})

		t.Run("missing encryption token", func(t *testing.T) {
			var packet transport.RelayInitPacket
			addr := "127.0.0.1:40000"
			buff := make([]byte, 8+crypto.NonceSize+4+len(addr)) // 4 is the uint32 for address length
			binary.LittleEndian.PutUint32(buff, rand.Uint32())
			binary.LittleEndian.PutUint32(buff[4:], rand.Uint32())
			binary.LittleEndian.PutUint32(buff[8+crypto.NonceSize:], uint32(len(addr)))
			copy(buff[12+crypto.NonceSize:], addr)
			assert.Equal(t, packet.UnmarshalBinary(buff), errors.New("invalid packet"))
		})

		t.Run("address not formatted correctly", func(t *testing.T) {
			var packet transport.RelayInitPacket
			addr := "invalid"
			buff := make([]byte, 8+crypto.NonceSize+4+len(addr)+routing.EncryptedTokenSize)
			binary.LittleEndian.PutUint32(buff, rand.Uint32())
			binary.LittleEndian.PutUint32(buff[4:], rand.Uint32())
			binary.LittleEndian.PutUint32(buff[8+crypto.NonceSize:], uint32(len(addr)))
			copy(buff[12+crypto.NonceSize:], addr)
			assert.Equal(t, packet.UnmarshalBinary(buff), errors.New("could not resolve init packet with address 'invalid' with reason: address invalid: missing port in address"))
		})

		t.Run("valid", func(t *testing.T) {
			var packet transport.RelayInitPacket
			addr := "127.0.0.1:40000"
			buff := make([]byte, 8+crypto.NonceSize+4+len(addr)+routing.EncryptedTokenSize)
			binary.LittleEndian.PutUint32(buff, rand.Uint32())
			binary.LittleEndian.PutUint32(buff[4:], rand.Uint32())
			binary.LittleEndian.PutUint32(buff[8+crypto.NonceSize:], uint32(len(addr)))
			copy(buff[12+crypto.NonceSize:], addr)
			assert.Nil(t, packet.UnmarshalBinary(buff))
		})
	})

	t.Run("MarshalBinary()", func(t *testing.T) {
		nonce := make([]byte, crypto.NonceSize)
		token := make([]byte, routing.EncryptedTokenSize)
		rand.Read(nonce)
		rand.Read(token)

		udp, _ := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
		expected := transport.RelayInitPacket{
			Magic:          rand.Uint32(),
			Version:        rand.Uint32(),
			Nonce:          nonce,
			Address:        *udp,
			EncryptedToken: token,
		}

		var actual transport.RelayInitPacket

		data, _ := expected.MarshalBinary()

		assert.Nil(t, actual.UnmarshalBinary(data))
		assert.Equal(t, expected, actual)
	})
}

func TestRelayUpdatePacket(t *testing.T) {
	t.Run("UnmarshalBinary()", func(t *testing.T) {
		t.Run("missing request version", func(t *testing.T) {
			var packet transport.RelayUpdatePacket
			assert.Equal(t, packet.UnmarshalBinary(make([]byte, 0)), errors.New("invalid packet"))
		})

		t.Run("missing relay address", func(t *testing.T) {
			var packet transport.RelayUpdatePacket
			buff := make([]byte, 4)
			binary.LittleEndian.PutUint32(buff, rand.Uint32()) //version
			assert.Equal(t, packet.UnmarshalBinary(buff), errors.New("invalid packet"))
		})

		t.Run("missing relay token", func(t *testing.T) {
			var packet transport.RelayUpdatePacket
			buff := make([]byte, 4+13)
			binary.LittleEndian.PutUint32(buff, rand.Uint32())
			binary.LittleEndian.PutUint32(buff[4:], 13) // address length
			assert.Equal(t, packet.UnmarshalBinary(buff), errors.New("invalid packet"))
		})

		t.Run("missing number of relays", func(t *testing.T) {
			var packet transport.RelayUpdatePacket
			buff := make([]byte, 4+4+13+crypto.KeySize)
			binary.LittleEndian.PutUint32(buff, rand.Uint32())
			binary.LittleEndian.PutUint32(buff[4:], 13)
			assert.Equal(t, packet.UnmarshalBinary(buff), errors.New("invalid packet"))
		})

		t.Run("address is not formatted correctly", func(t *testing.T) {
			var packet transport.RelayUpdatePacket
			addr := "invalid"
			buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4)
			binary.LittleEndian.PutUint32(buff, rand.Uint32())
			binary.LittleEndian.PutUint32(buff[4:], uint32(len(addr)))
			copy(buff[8:], addr)
			binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize:], 1) // number of relays
			assert.Equal(t, packet.UnmarshalBinary(buff), errors.New("could not resolve init packet with address 'invalid' with reason: address invalid: missing port in address"))
		})

		t.Run("missing various relay ping stats", func(t *testing.T) {
			t.Run("missing the id", func(t *testing.T) {
				var packet transport.RelayUpdatePacket
				addr := "127.0.0.1:40000"
				buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4)
				binary.LittleEndian.PutUint32(buff, rand.Uint32())
				binary.LittleEndian.PutUint32(buff[4:], uint32(len(addr)))
				copy(buff[8:], addr)
				binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize:], 1) // number of relays
				assert.Equal(t, packet.UnmarshalBinary(buff), errors.New("invalid packet"))
			})

			t.Run("missing the rtt", func(t *testing.T) {
				var packet transport.RelayUpdatePacket
				addr := "127.0.0.1:40000"
				buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4+8)
				binary.LittleEndian.PutUint32(buff, rand.Uint32())
				binary.LittleEndian.PutUint32(buff[4:], uint32(len(addr)))
				copy(buff[8:], addr)
				binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize:], 1)
				binary.LittleEndian.PutUint64(buff[8+len(addr)+crypto.KeySize+4:], rand.Uint64()) // relay id
				assert.Equal(t, packet.UnmarshalBinary(buff), errors.New("invalid packet"))
			})

			t.Run("missing the jitter", func(t *testing.T) {
				var packet transport.RelayUpdatePacket
				addr := "127.0.0.1:40000"
				buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4+8+4)
				binary.LittleEndian.PutUint32(buff, rand.Uint32())
				binary.LittleEndian.PutUint32(buff[4:], uint32(len(addr)))
				copy(buff[8:], addr)
				binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize:], 1)
				binary.LittleEndian.PutUint64(buff[8+len(addr)+crypto.KeySize+4:], rand.Uint64())
				binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize+12:], math.Float32bits(rand.Float32())) // rtt
				assert.Equal(t, packet.UnmarshalBinary(buff), errors.New("invalid packet"))
			})

			t.Run("missing the packet loss", func(t *testing.T) {
				var packet transport.RelayUpdatePacket
				addr := "127.0.0.1:40000"
				buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4+8+4+4)
				binary.LittleEndian.PutUint32(buff, rand.Uint32())
				binary.LittleEndian.PutUint32(buff[4:], uint32(len(addr)))
				copy(buff[8:], addr)
				binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize:], 1)
				binary.LittleEndian.PutUint64(buff[8+len(addr)+crypto.KeySize+4:], rand.Uint64())
				binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize+12:], math.Float32bits(rand.Float32()))
				binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize+16:], math.Float32bits(rand.Float32())) // jitter
				assert.Equal(t, packet.UnmarshalBinary(buff), errors.New("invalid packet"))
			})
		})

		t.Run("valid", func(t *testing.T) {
			var packet transport.RelayUpdatePacket
			addr := "127.0.0.1:40000"
			buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4+8+4+4+4)
			binary.LittleEndian.PutUint32(buff, rand.Uint32())
			binary.LittleEndian.PutUint32(buff[4:], uint32(len(addr)))
			copy(buff[8:], addr)
			binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize:], 1)
			binary.LittleEndian.PutUint64(buff[8+len(addr)+crypto.KeySize+4:], rand.Uint64())
			binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize+12:], math.Float32bits(rand.Float32()))
			binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize+16:], math.Float32bits(rand.Float32()))
			binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize+20:], math.Float32bits(rand.Float32())) // packet loss
			assert.Nil(t, packet.UnmarshalBinary(buff))
		})
	})

	t.Run("MarshalBinary()", func(t *testing.T) {
		stats := make([]core.RelayStatsPing, 5)

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
		expected := transport.RelayUpdatePacket{
			Version:   rand.Uint32(),
			Address:   *udp,
			Token:     token,
			NumRelays: uint32(len(stats)),
			PingStats: stats,
		}

		data, _ := expected.MarshalBinary()

		var actual transport.RelayUpdatePacket
		assert.Nil(t, actual.UnmarshalBinary(data))
		assert.Equal(t, expected, actual)
	})
}
