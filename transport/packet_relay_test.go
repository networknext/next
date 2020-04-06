package transport_test

import (
	crand "crypto/rand"
	"encoding/base64"
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

func TestRelayRequestUnmarshalJSON(t *testing.T) {
	t.Run("unparsable json", func(t *testing.T) {
		jsonRequest := []byte("{")

		var packet transport.RelayRequest

		err := packet.UnmarshalJSON(jsonRequest)
		assert.EqualError(t, err, "unexpected end of JSON input")
	})

	t.Run("relay address is invalid", func(t *testing.T) {
		jsonRequest := []byte(`{
			"address": "invalid"
		}`)

		var packet transport.RelayRequest

		err := packet.UnmarshalJSON(jsonRequest)
		assert.EqualError(t, err, "address invalid: missing port in address")
	})

	t.Run("relay id to ping is invalid", func(t *testing.T) {
		jsonRequest := []byte(`{
			"address": "127.0.0.1:40000",
			"ping_stats": [
				{
					"id": xxx,
					"address": "invalid"
				}	
			]
		}`)

		var packet transport.RelayRequest

		err := packet.UnmarshalJSON(jsonRequest)
		assert.EqualError(t, err, "invalid character 'x' looking for beginning of value")
	})

	t.Run("relay address to ping is invalid", func(t *testing.T) {
		jsonRequest := []byte(`{
			"address": "127.0.0.1:40000",
			"ping_stats": [
				{
					"id": 999999,
					"address": "invalid"
				}	
			]
		}`)

		var packet transport.RelayRequest

		err := packet.UnmarshalJSON(jsonRequest)
		assert.EqualError(t, err, "address invalid: missing port in address")
	})

	t.Run("valid", func(t *testing.T) {
		jsonRequest := []byte(`{
			"address": "127.0.0.1:40000",
			"ping_stats": [
				{
					"id": 999999,
					"address": "127.0.0.2:40000",
					"rtt": 1,
					"jitter": 1,
					"packet_loss": 1
				},
				{
					"id": 12345,
					"address": "127.0.0.3:40000",
					"rtt": 1,
					"jitter": 1,
					"packet_loss": 1
				}
			],
			"traffic_stats": {
				"session_count": 100,
				"bytes_tx": 200,
				"bytes_rx": 100
			}
		}`)

		var packet transport.RelayRequest

		err := packet.UnmarshalJSON(jsonRequest)
		assert.NoError(t, err)
	})
}

func TestRelayRequestMarshalJSON(t *testing.T) {
	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
	assert.NoError(t, err)

	statIps := []string{"127.0.0.2:40000", "127.0.0.3:40000", "127.0.0.4:40000", "127.0.0.5:40000"}

	expected := transport.RelayRequest{
		Address: *addr,
		PingStats: []transport.RelayPingStats{
			transport.RelayPingStats{
				ID:         1,
				Address:    statIps[0],
				RTT:        1,
				Jitter:     2,
				PacketLoss: 3,
			},

			transport.RelayPingStats{
				ID:         2,
				Address:    statIps[1],
				RTT:        4,
				Jitter:     5,
				PacketLoss: 6,
			},

			transport.RelayPingStats{
				ID:         3,
				Address:    statIps[2],
				RTT:        7,
				Jitter:     8,
				PacketLoss: 9,
			},

			transport.RelayPingStats{
				ID:         4,
				Address:    statIps[3],
				RTT:        10,
				Jitter:     11,
				PacketLoss: 12,
			},
		},
		TrafficStats: transport.RelayTrafficStats{
			SessionCount:  10,
			BytesSent:     1000000,
			BytesReceived: 1000000,
		},
	}

	var actual transport.RelayRequest

	data, err := expected.MarshalJSON()
	assert.NoError(t, err)

	err = actual.UnmarshalJSON(data)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestRelayInitRequestUnmarshalJSON(t *testing.T) {
	t.Run("unparsable json", func(t *testing.T) {
		jsonRequest := []byte("{")

		var packet transport.RelayInitRequest

		assert.Error(t, packet.UnmarshalJSON(jsonRequest))
	})

	t.Run("nonce is invalid base64", func(t *testing.T) {
		jsonRequest := []byte(`{
			"magic_request_protection": 1,
			"version": 1,
			"relay_address": "127.0.0.1:1111",
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
			"relay_address": "127.0.0.1:1111",
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
			"nonce": "` + b64Nonce + `",
			"encrypted_token": "` + b64EncToken + `"
		}`)

		var packet transport.RelayInitRequest

		assert.NoError(t, packet.UnmarshalJSON(jsonRequest))
		assert.Equal(t, expectedPacket, packet)
	})
}

func TestRelayInitRequestMarshalJSON(t *testing.T) {
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

	data, _ := expected.MarshalJSON()

	assert.Nil(t, actual.UnmarshalJSON(data))
	assert.Equal(t, expected, actual)
}

func TestRelayInitRequestUnmarshalBinary(t *testing.T) {
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

func TestRelayInitResponseMarshalJSON(t *testing.T) {
	res := transport.RelayInitResponse{
		Version:   12345,
		Timestamp: 12345,
		PublicKey: []byte{0x13, 0x14},
	}

	jsonRes, err := res.MarshalJSON()
	assert.NoError(t, err)

	assert.JSONEq(t, `{"PublicKey": "ExQ=", "Timestamp":12345, "Version":0}`, string(jsonRes))
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

func TestRelayUpdateRequestUnmarshalJSON(t *testing.T) {
	t.Run("invalid address", func(t *testing.T) {
		var packet transport.RelayUpdateRequest

		// no port
		jsonRequest := []byte(`{
			"version": 1,
			"relay_address": "127.0.0.1"
		}`)
		assert.EqualError(t, packet.UnmarshalJSON(jsonRequest), "address 127.0.0.1: missing port in address")

		// bad port
		jsonRequest = []byte(`{
			"version": 1,
			"relay_address": "127.0.0.1:x"
		}`)
		assert.Error(t, packet.UnmarshalJSON(jsonRequest))

		// invalid address
		jsonRequest = []byte(`{
			"version": 1,
			"relay_address": "127.0:1111"
		}`)
		assert.EqualError(t, packet.UnmarshalJSON(jsonRequest), "invalid relay_address")
	})

	t.Run("invalid token size", func(t *testing.T) {
		var packet transport.RelayUpdateRequest

		jsonRequest := []byte(`{
			"version": 1,
			"relay_address": "127.0.0.1:40000",
			"Metadata": {
				"PublicKey": "this is a test"
			}
		}`)
		assert.EqualError(t, packet.UnmarshalJSON(jsonRequest), "illegal base64 data at input byte 4")

		jsonRequest = []byte(`{
			"version": 1,
			"relay_address": "127.0.0.1:40000",
			"Metadata": {
				"PublicKey": "AAAA"
			}
		}`)
		assert.EqualError(t, packet.UnmarshalJSON(jsonRequest), "invalid token size")
	})

	t.Run("invalid ping stats", func(t *testing.T) {
		var packet transport.RelayUpdateRequest

		jsonRequest := []byte(`{
			"version": 1,
			"relay_address": "127.0.0.1:40000",
			"Metadata": {
				"PublicKey": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
			}
		}`)
		assert.EqualError(t, packet.UnmarshalJSON(jsonRequest), "unexpected end of JSON input")

		jsonRequest = []byte(`{
			"version": 1,
			"relay_address": "127.0.0.1:40000",
			"Metadata": {
				"PublicKey": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
			}
			"PingStats": 0
		}`)
		assert.EqualError(t, packet.UnmarshalJSON(jsonRequest), "json: cannot unmarshal number into Go value of type []routing.RelayStatsPing")
	})

	t.Run("valid", func(t *testing.T) {
		var packet transport.RelayUpdateRequest

		jsonRequest := []byte(`{
			"version": 1,
			"relay_address": "127.0.0.1:40000",
			"Metadata": {
				"PublicKey": "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="
			}
			"TrafficStats": {
				"BytesMeasurementRx": 100
			},
			"PingStats": [{
				"RelayID": 1,
				"RTT": 2,
				"Jitter": 3,
				"PacketLoss": 4
			}],
			"shutting_down": true
		}`)
		assert.NoError(t, packet.UnmarshalJSON(jsonRequest))

		expected := transport.RelayUpdateRequest{
			Version:       1,
			Address:       net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 40000},
			Token:         make([]byte, crypto.KeySize),
			BytesReceived: 100,
			PingStats: []routing.RelayStatsPing{
				{RelayID: 1, RTT: 2, Jitter: 3, PacketLoss: 4},
			},
			ShuttingDown: true,
		}
		assert.Equal(t, expected, packet)
	})
}

func TestRelayUpdateRequestUnmarshalBinary(t *testing.T) {
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
			assert.Equal(t, packet.UnmarshalBinary(buff), errors.New("invalid packet, could not read a ping stat"))
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
			assert.Equal(t, packet.UnmarshalBinary(buff), errors.New("invalid packet, could not read a ping stat"))
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
			assert.Equal(t, packet.UnmarshalBinary(buff), errors.New("invalid packet, could not read a ping stat"))
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
			assert.Equal(t, packet.UnmarshalBinary(buff), errors.New("invalid packet, could not read a ping stat"))
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
		assert.Equal(t, packet.UnmarshalBinary(buff), errors.New("invalid packet, could not read bytes received"))
	})

	t.Run("missing shutdown flag", func(t *testing.T) {
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
		assert.Equal(t, packet.UnmarshalBinary(buff), errors.New("invalid packet, could not read shutdown flag"))
	})

	t.Run("valid", func(t *testing.T) {
		var packet transport.RelayUpdateRequest
		packet.ShuttingDown = true
		addr := "127.0.0.1:40000"
		buff := make([]byte, 4+4+len(addr)+crypto.KeySize+4+8+4+4+4+8+1)
		binary.LittleEndian.PutUint32(buff, rand.Uint32())
		binary.LittleEndian.PutUint32(buff[4:], uint32(len(addr)))
		copy(buff[8:], addr)
		binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize:], 1)
		binary.LittleEndian.PutUint64(buff[8+len(addr)+crypto.KeySize+4:], rand.Uint64())
		binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize+12:], math.Float32bits(rand.Float32()))
		binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize+16:], math.Float32bits(rand.Float32()))
		binary.LittleEndian.PutUint32(buff[8+len(addr)+crypto.KeySize+20:], math.Float32bits(rand.Float32()))
		binary.LittleEndian.PutUint64(buff[8+len(addr)+crypto.KeySize+24:], rand.Uint64())
		buff[8+len(addr)+crypto.KeySize+32] = 1
		assert.Nil(t, packet.UnmarshalBinary(buff))
	})
}

func TestRelayUpdateRequestMarshalJSON(t *testing.T) {
	stats := make([]routing.RelayStatsPing, 1)

	stat := &stats[0]
	stat.RelayID = rand.Uint64()
	stat.RTT = rand.Float32()
	stat.Jitter = rand.Float32()
	stat.PacketLoss = rand.Float32()

	token := make([]byte, crypto.KeySize)
	rand.Read(token)

	udp, _ := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
	version := rand.Uint32()
	req := transport.RelayUpdateRequest{
		Version:       version,
		Address:       *udp,
		Token:         token,
		PingStats:     stats,
		BytesReceived: rand.Uint64(),
	}

	jsonRes, err := req.MarshalJSON()
	assert.NoError(t, err)

	assert.JSONEq(t, fmt.Sprintf(`{
		"version":%d,
		"relay_address":"127.0.0.1:40000",
		"PingStats":[{
			"RelayId":%d,
			"RTT":%v,
			"Jitter":%v,
			"PacketLoss":%v
		}],
		"TrafficStats":{
			"BytesMeasurementRx":%d
		},
		"Metadata":{
			"PublicKey":"%s"
		},
		"shutting_down":false
	}`, version, stat.RelayID, stat.RTT, stat.Jitter, stat.PacketLoss, req.BytesReceived, base64.StdEncoding.EncodeToString(token)), string(jsonRes))
}

func TestRelayUpdateRequestMarshalBinary(t *testing.T) {
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

	data, err := expected.MarshalBinary()
	assert.NoError(t, err)

	var actual transport.RelayUpdateRequest
	err = actual.UnmarshalBinary(data)
	assert.NoError(t, err)
	assert.Equal(t, expected, actual)
}

func TestRelayUpdateResponseUnmarshalBinary(t *testing.T) {
	t.Run("missing response version", func(t *testing.T) {
		var packet transport.RelayUpdateResponse
		buff := make([]byte, 0)
		assert.EqualError(t, packet.UnmarshalBinary(buff), "failed to unmarshal relay update response version")
	})

	t.Run("missing response number of relays to ping", func(t *testing.T) {
		var packet transport.RelayUpdateResponse
		index := 0
		buff := make([]byte, 4)
		encoding.WriteUint32(buff, &index, rand.Uint32()) //version
		assert.EqualError(t, packet.UnmarshalBinary(buff), "failed to unmarshal relay update response number of relays to ping")
	})

	t.Run("missing response relay id", func(t *testing.T) {
		var packet transport.RelayUpdateResponse
		index := 0
		buff := make([]byte, 4+4)
		encoding.WriteUint32(buff, &index, rand.Uint32()) //version
		encoding.WriteUint32(buff, &index, 1)             //numRelaysToPing
		assert.EqualError(t, packet.UnmarshalBinary(buff), "failed to unmarshal relay update response relay id")
	})

	t.Run("missing response relay address", func(t *testing.T) {
		var packet transport.RelayUpdateResponse
		index := 0
		buff := make([]byte, 4+4+8)
		encoding.WriteUint32(buff, &index, rand.Uint32()) //version
		encoding.WriteUint32(buff, &index, 1)             //numRelaysToPing
		encoding.WriteUint64(buff, &index, rand.Uint64()) //id
		assert.EqualError(t, packet.UnmarshalBinary(buff), "failed to unmarshal relay update response relay address")
	})

	t.Run("valid", func(t *testing.T) {
		var packet transport.RelayUpdateResponse
		addr := "127.0.0.1:40000"
		index := 0
		buff := make([]byte, 4+4+8+(4+len(addr)))
		encoding.WriteUint32(buff, &index, rand.Uint32())                         //version
		encoding.WriteUint32(buff, &index, 1)                                     //numRelaysToPing
		encoding.WriteUint64(buff, &index, rand.Uint64())                         //id
		encoding.WriteString(buff, &index, addr, transport.MaxRelayAddressLength) //address
		assert.NoError(t, packet.UnmarshalBinary(buff))
	})
}

func TestRelayUpdateResponseMarshalJSON(t *testing.T) {
	addr := "127.0.0.1:40000"

	pingToken := routing.LegacyPingToken{
		Timeout: 0,
		RelayID: 0,
		HMac:    [32]byte{},
	}

	pingTokenData, err := pingToken.MarshalBinary()
	assert.NoError(t, err)

	relaysToPing := []routing.LegacyPingData{
		routing.LegacyPingData{
			RelayPingData: routing.RelayPingData{
				ID:      crypto.HashID(addr),
				Address: addr,
			},
			PingToken: base64.StdEncoding.EncodeToString(pingTokenData),
		},
	}

	response := transport.RelayUpdateResponse{
		RelaysToPing: relaysToPing,
	}

	expected := `{
		"ping_data":[{
			"relay_id":14990044260459612264,
			"relay_address":"127.0.0.1:40000",
			"ping_info":"AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA"
		}],
		"version":0
	}`

	buff, err := response.MarshalJSON()
	assert.NoError(t, err)
	assert.JSONEq(t, expected, string(buff))
}

func TestRelayUpdateResponseMarshalBinary(t *testing.T) {
	addr := "127.0.0.1:40000"

	expected := transport.RelayUpdateResponse{
		RelaysToPing: []routing.LegacyPingData{
			routing.LegacyPingData{
				RelayPingData: routing.RelayPingData{
					ID:      crypto.HashID(addr),
					Address: addr,
				},
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
