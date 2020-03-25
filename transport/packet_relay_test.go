package transport_test

import (
	crand "crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"math"
	"math/rand"
	"net"
	"testing"

	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/transport"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/nacl/box"
)

func TestRelayInitRequest(t *testing.T) {
	t.Run("UnmarshalJSON()", func(t *testing.T) {
		t.Run("nonce is invalid base64", func(t *testing.T) {
			jsonRequest := []byte(`{
				"magic_request_protection": 1,
				"version": 1,
				"relay_address": "127.0.0.1:1111",
				"relay_port": 1111,
				"nonce": "\n\ninvalid\t\t",
				"encrypted_token": ""
			}`)

			var packet transport.RelayInitRequest

			assert.Error(t, packet.UnmarshalJSON(jsonRequest))
		})

		t.Run("address is invalid", func(t *testing.T) {
			nonce := make([]byte, crypto.NonceSize)
			crand.Read(nonce)

			jsonRequest := []byte(`{
				"magic_request_protection": 1,
				"version": 1,
				"relay_address": "invalid",
				"relay_port": 0,
				"nonce": "` + base64.StdEncoding.EncodeToString(nonce) + `",
				"encrypted_token": ""
			}`)

			var packet transport.RelayInitRequest

			assert.Error(t, packet.UnmarshalJSON(jsonRequest))
		})

		t.Run("token is invalid base64", func(t *testing.T) {
			nonce := make([]byte, crypto.NonceSize)
			crand.Read(nonce)

			jsonRequest := []byte(`{
				"magic_request_protection": 1,
				"version": 1,
				"relay_address": "127.0.0.1",
				"relay_port": 0,
				"nonce": "` + base64.StdEncoding.EncodeToString(nonce) + `",
				"encrypted_token": "\n\ninvalid\t\t"
			}`)

			var packet transport.RelayInitRequest

			assert.Error(t, packet.UnmarshalJSON(jsonRequest))
		})

		t.Run("valid", func(t *testing.T) {
			nonce := make([]byte, crypto.NonceSize)
			crand.Read(nonce)
			b64Nonce := base64.StdEncoding.EncodeToString(nonce)

			routerPublicKey, _, err := box.GenerateKey(crand.Reader)
			assert.NoError(t, err)

			_, relayPrivateKey, err := box.GenerateKey(crand.Reader)
			assert.NoError(t, err)

			token := make([]byte, crypto.KeySize)
			crand.Read(token)
			encryptedToken := crypto.Seal(token, nonce, routerPublicKey[:], relayPrivateKey[:])
			b64EncToken := base64.StdEncoding.EncodeToString(encryptedToken)

			expectedPacket := transport.RelayInitRequest{
				Magic:          1,
				Version:        1,
				Address:        net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 1111},
				Nonce:          nonce,
				EncryptedToken: encryptedToken,
			}

			jsonRequest := []byte(`{
				"magic_request_protection": 1,
				"version": 1,
				"relay_address": "127.0.0.1:1111",
				"relay_port": 1111,
				"nonce": "` + b64Nonce + `",
				"encrypted_token": "` + b64EncToken + `"
			}`)

			var packet transport.RelayInitRequest

			assert.NoError(t, packet.UnmarshalJSON(jsonRequest))
			assert.Equal(t, expectedPacket, packet)
		})
	})

	t.Run("UnmarshalBinary()", func(t *testing.T) {
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
			buff := make([]byte, 8+crypto.NonceSize+4+len(addr)+routing.EncryptedTokenSize)
			binary.LittleEndian.PutUint32(buff, rand.Uint32())
			binary.LittleEndian.PutUint32(buff[4:], rand.Uint32())
			binary.LittleEndian.PutUint32(buff[8+crypto.NonceSize:], uint32(len(addr)))
			copy(buff[12+crypto.NonceSize:], addr)
			assert.Equal(t, packet.UnmarshalBinary(buff), errors.New("could not resolve init packet with address 'invalid' with reason: address invalid: missing port in address"))
		})

		t.Run("valid", func(t *testing.T) {
			var packet transport.RelayInitRequest
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
	})
}

func TestRelayInitResponse(t *testing.T) {
	t.Run("MarshalJSON()", func(t *testing.T) {
		res := transport.RelayInitResponse{
			Version:   12345,
			Timestamp: 12345,
			PublicKey: []byte{0x13, 0x14},
		}

		jsonRes, err := res.MarshalJSON()
		assert.NoError(t, err)

		assert.JSONEq(t, `{"Timestamp":12345}`, string(jsonRes))
	})

	t.Run("MarshalBinary()", func(t *testing.T) {
		res := transport.RelayInitResponse{
			Version:   12345,
			Timestamp: 12345,
			PublicKey: make([]byte, crypto.KeySize),
		}

		binaryRes, err := res.MarshalBinary()
		assert.NoError(t, err)

		expected := []byte{0x0, 0x0, 0x0, 0x0, 0x39, 0x30, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0, 0x0}
		assert.Equal(t, expected, binaryRes)
	})
}

func TestRelayUpdateRequest(t *testing.T) {
	t.Run("UnmarshalBinary()", func(t *testing.T) {
		t.Run("missing request version", func(t *testing.T) {
			var packet transport.RelayUpdateRequest
			assert.Equal(t, packet.UnmarshalBinary(make([]byte, 0)), errors.New("invalid packet"))
		})

		t.Run("missing relay address", func(t *testing.T) {
			var packet transport.RelayUpdateRequest
			buff := make([]byte, 4)
			binary.LittleEndian.PutUint32(buff, rand.Uint32()) //version
			assert.Equal(t, packet.UnmarshalBinary(buff), errors.New("invalid packet"))
		})

		t.Run("missing relay token", func(t *testing.T) {
			var packet transport.RelayUpdateRequest
			buff := make([]byte, 4+13)
			binary.LittleEndian.PutUint32(buff, rand.Uint32())
			binary.LittleEndian.PutUint32(buff[4:], 13) // address length
			assert.Equal(t, packet.UnmarshalBinary(buff), errors.New("invalid packet"))
		})

		t.Run("missing number of relays", func(t *testing.T) {
			var packet transport.RelayUpdateRequest
			buff := make([]byte, 4+4+13+crypto.KeySize)
			binary.LittleEndian.PutUint32(buff, rand.Uint32())
			binary.LittleEndian.PutUint32(buff[4:], 13)
			assert.Equal(t, packet.UnmarshalBinary(buff), errors.New("invalid packet"))
		})

		t.Run("address is not formatted correctly", func(t *testing.T) {
			var packet transport.RelayUpdateRequest
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
				var packet transport.RelayUpdateRequest
				addr := "127.0.0.1:40000"
				buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4)
				binary.LittleEndian.PutUint32(buff, rand.Uint32())
				binary.LittleEndian.PutUint32(buff[4:], uint32(len(addr)))
				copy(buff[8:], addr)
				binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize:], 1) // number of relays
				assert.Equal(t, packet.UnmarshalBinary(buff), errors.New("invalid packet"))
			})

			t.Run("missing the rtt", func(t *testing.T) {
				var packet transport.RelayUpdateRequest
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
				var packet transport.RelayUpdateRequest
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
				var packet transport.RelayUpdateRequest
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

		t.Run("missing received stats", func(t *testing.T) {
			var packet transport.RelayUpdateRequest
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
			assert.Equal(t, packet.UnmarshalBinary(buff), errors.New("invalid packet"))
		})

		t.Run("valid", func(t *testing.T) {
			var packet transport.RelayUpdateRequest
			addr := "127.0.0.1:40000"
			buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4+8+4+4+4+8)
			binary.LittleEndian.PutUint32(buff, rand.Uint32())
			binary.LittleEndian.PutUint32(buff[4:], uint32(len(addr)))
			copy(buff[8:], addr)
			binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize:], 1)
			binary.LittleEndian.PutUint64(buff[8+len(addr)+crypto.KeySize+4:], rand.Uint64())
			binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize+12:], math.Float32bits(rand.Float32()))
			binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize+16:], math.Float32bits(rand.Float32()))
			binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize+20:], math.Float32bits(rand.Float32()))
			binary.LittleEndian.PutUint64(buff[8+len(addr)+crypto.KeySize+24:], rand.Uint64()) // bytes received
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
		expected := transport.RelayUpdateRequest{
			Version:   rand.Uint32(),
			Address:   *udp,
			Token:     token,
			PingStats: stats,
		}

		data, _ := expected.MarshalBinary()

		var actual transport.RelayUpdateRequest
		assert.Nil(t, actual.UnmarshalBinary(data))
		assert.Equal(t, expected, actual)
	})
}

// func TestRelayUpdateRequestJSON(t *testing.T) {
// 	t.Run("ToUpdatePacket()", func(t *testing.T) {
// 		t.Run("invalid address", func(t *testing.T) {
// 			var jsonRequest transport.RelayUpdateRequestJSON
// 			jsonRequest.StringAddr = "invalid"
// 			var packet transport.RelayUpdateRequest

// 			assert.IsType(t, &net.AddrError{}, jsonRequest.ToUpdatePacket(&packet))
// 		})

// 		t.Run("token is invalid base64", func(t *testing.T) {
// 			var jsonRequest transport.RelayUpdateRequestJSON
// 			jsonRequest.Metadata.TokenBase64 = "\t\ninvalid\n\n\t"
// 			var packet transport.RelayUpdateRequest

// 			assert.IsType(t, base64.CorruptInputError(0), jsonRequest.ToUpdatePacket(&packet))
// 		})

// 		t.Run("valid", func(t *testing.T) {
// 			statIps := []string{"127.0.0.2:40000", "127.0.0.3:40000", "127.0.0.4:40000", "127.0.0.5:40000"}
// 			token := make([]byte, crypto.KeySize)
// 			b64Token := base64.StdEncoding.EncodeToString(token)
// 			var request transport.RelayUpdateRequestJSON
// 			request.StringAddr = "127.0.0.1:40000"
// 			request.Metadata.TokenBase64 = b64Token
// 			request.PingStats = make([]routing.RelayStatsPing, uint32(len(statIps)))
// 			request.TrafficStats.BytesMeasurementRx = rand.Uint64()

// 			for i, addr := range statIps {
// 				stats := &request.PingStats[i]
// 				stats.RelayID = crypto.HashID(addr)
// 				stats.RTT = rand.Float32()
// 				stats.Jitter = rand.Float32()
// 				stats.PacketLoss = rand.Float32()
// 			}

// 			var packet transport.RelayUpdateRequest
// 			assert.Nil(t, request.ToUpdatePacket(&packet))

// 			assert.Equal(t, request.StringAddr, packet.Address.String())
// 			assert.True(t, bytes.Equal(packet.Token, token))
// 			assert.Equal(t, request.PingStats, packet.PingStats)
// 			assert.Equal(t, request.TrafficStats.BytesMeasurementRx, packet.BytesReceived)
// 		})
// 	})
// }
