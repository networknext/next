package transport_test

import (
	"bytes"
	crand "crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"math"
	"math/rand"
	"net"
	"os"
	"testing"

	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/transport"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/nacl/box"
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

func TestRelayInitRequestJSON(t *testing.T) {
	t.Run("ToInitPacket()", func(t *testing.T) {
		t.Run("nonce is invalid base64", func(t *testing.T) {
			var jsonRequest transport.RelayInitRequestJSON
			jsonRequest.NonceBase64 = "\n\ninvalid\t\t"
			var packet transport.RelayInitPacket

			assert.IsType(t, base64.CorruptInputError(9), jsonRequest.ToInitPacket(&packet)) // (9) because the first invalid char is at position 9, print out the error message for more details
		})

		t.Run("address is invalid", func(t *testing.T) {
			var jsonRequest transport.RelayInitRequestJSON
			nonce := make([]byte, crypto.NonceSize)
			crand.Read(nonce)
			b64Nonce := base64.StdEncoding.EncodeToString(nonce)
			jsonRequest.NonceBase64 = b64Nonce
			jsonRequest.StringAddr = "invalid"
			jsonRequest.PortNum = 0
			var packet transport.RelayInitPacket

			assert.IsType(t, &net.AddrError{}, jsonRequest.ToInitPacket(&packet))
		})

		t.Run("token is invalid base64", func(t *testing.T) {
			var jsonRequest transport.RelayInitRequestJSON
			nonce := make([]byte, crypto.NonceSize)
			crand.Read(nonce)
			b64Nonce := base64.StdEncoding.EncodeToString(nonce)
			jsonRequest.NonceBase64 = b64Nonce
			jsonRequest.StringAddr = "127.0.0.1:40000"
			jsonRequest.PortNum = 40000
			jsonRequest.EncryptedTokenBase64 = "\n\ninvalid\t\t"
			var packet transport.RelayInitPacket

			assert.Equal(t, base64.CorruptInputError(9), jsonRequest.ToInitPacket(&packet))
		})

		t.Run("valid", func(t *testing.T) {
			var jsonRequest transport.RelayInitRequestJSON

			nonce := make([]byte, crypto.NonceSize)
			crand.Read(nonce)
			b64Nonce := base64.StdEncoding.EncodeToString(nonce)

			routerPublicKey, _, err := box.GenerateKey(crand.Reader)
			assert.NoError(t, err)

			key := os.Getenv("RELAY_PRIVATE_KEY")
			assert.NotEqual(t, 0, len(key))
			relayPrivateKey, err := base64.StdEncoding.DecodeString(key)
			assert.NoError(t, err)

			token := make([]byte, crypto.KeySize)
			crand.Read(token)
			encryptedToken := crypto.Seal(token, nonce, routerPublicKey[:], relayPrivateKey[:])
			b64EncToken := base64.StdEncoding.EncodeToString(encryptedToken)

			jsonRequest.Magic = transport.InitRequestMagic
			jsonRequest.Version = transport.VersionNumberInitRequest
			jsonRequest.NonceBase64 = b64Nonce
			jsonRequest.StringAddr = "127.0.0.1:40000"
			jsonRequest.PortNum = 40000
			jsonRequest.EncryptedTokenBase64 = b64EncToken
			var packet transport.RelayInitPacket

			assert.Nil(t, jsonRequest.ToInitPacket(&packet))
			assert.Equal(t, jsonRequest.Magic, packet.Magic)
			assert.Equal(t, jsonRequest.Version, packet.Version)
			assert.Equal(t, jsonRequest.StringAddr, packet.Address.String())
			assert.True(t, bytes.Equal(packet.Nonce, nonce))
			assert.True(t, bytes.Equal(packet.EncryptedToken, encryptedToken))
		})
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

func TestRelayUpdateRequestJSON(t *testing.T) {
	t.Run("ToUpdatePacket()", func(t *testing.T) {
		t.Run("invalid address", func(t *testing.T) {
			var jsonRequest transport.RelayUpdateRequestJSON
			jsonRequest.StringAddr = "invalid"
			var packet transport.RelayUpdatePacket

			assert.IsType(t, &net.AddrError{}, jsonRequest.ToUpdatePacket(&packet))
		})

		t.Run("token is invalid base64", func(t *testing.T) {
			var jsonRequest transport.RelayUpdateRequestJSON
			jsonRequest.Metadata.TokenBase64 = "\t\ninvalid\n\n\t"
			var packet transport.RelayUpdatePacket

			assert.IsType(t, base64.CorruptInputError(0), jsonRequest.ToUpdatePacket(&packet))
		})

		t.Run("valid", func(t *testing.T) {
			statIps := []string{"127.0.0.2:40000", "127.0.0.3:40000", "127.0.0.4:40000", "127.0.0.5:40000"}
			token := make([]byte, crypto.KeySize)
			b64Token := base64.StdEncoding.EncodeToString(token)
			var request transport.RelayUpdateRequestJSON
			request.StringAddr = "127.0.0.1:40000"
			request.Metadata.TokenBase64 = b64Token
			request.PingStats = make([]routing.RelayStatsPing, uint32(len(statIps)))

			for i, addr := range statIps {
				stats := &request.PingStats[i]
				stats.RelayID = crypto.HashID(addr)
				stats.RTT = rand.Float32()
				stats.Jitter = rand.Float32()
				stats.PacketLoss = rand.Float32()
			}

			var packet transport.RelayUpdatePacket
			assert.Nil(t, request.ToUpdatePacket(&packet))

			assert.Equal(t, request.StringAddr, packet.Address.String())
			assert.True(t, bytes.Equal(packet.Token, token))
			assert.Equal(t, request.PingStats, packet.PingStats)
		})
	})
}
