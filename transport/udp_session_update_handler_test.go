package transport_test

import (
	"bytes"
	"crypto/ed25519"
	"net"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-kit/kit/log"
	"github.com/go-redis/redis/v7"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport"
	"github.com/stretchr/testify/assert"
)

func TestFailToUnmarshal(t *testing.T) {
	redisServer, _ := miniredis.Run()
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
	assert.NoError(t, err)

	var resbuf bytes.Buffer

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, nil, nil, nil, nil, nil, nil)
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: []byte("this is not a proper packet")})

	assert.Equal(t, 0, resbuf.Len())
}

func TestNoBuyerFound(t *testing.T) {
	redisServer, _ := miniredis.Run()
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	db := storage.InMemory{}

	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
	assert.NoError(t, err)

	serverCacheEntry := transport.ServerCacheEntry{
		Sequence: 13,
		Server: routing.Server{
			Addr:      *addr,
			PublicKey: TestPublicKey,
		},
	}
	sceData, err := serverCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SERVER-0.0.0.0:13", string(sceData))
	assert.NoError(t, err)

	packet := transport.SessionUpdatePacket{
		Sequence:      13,
		ServerAddress: net.UDPAddr{IP: net.IPv4zero, Port: 13},

		ClientRoutePublicKey: make([]byte, crypto.KeySize),

		Signature: make([]byte, ed25519.SignatureSize),
	}

	data, err := packet.MarshalBinary()
	assert.NoError(t, err)

	var resbuf bytes.Buffer

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, &db, nil, nil, nil, TestPrivateKey, nil)
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	assert.Greater(t, resbuf.Len(), 0)

	var actual transport.SessionResponsePacket
	err = actual.UnmarshalBinary(resbuf.Bytes())
	assert.NoError(t, err)

	assert.Equal(t, int(actual.RouteType), routing.RouteTypeDirect)
}

func TestVerificationFailed(t *testing.T) {
	_, buyersServerPrivKey, err := ed25519.GenerateKey(nil)
	assert.NoError(t, err)

	redisServer, _ := miniredis.Run()
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	db := storage.InMemory{
		LocalBuyer: &routing.Buyer{
			PublicKey: TestPublicKey, // intentionally different public key to private key generated above to cause error
		},
	}

	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
	assert.NoError(t, err)

	serverCacheEntry := transport.ServerCacheEntry{
		Sequence: 13,
		Server: routing.Server{
			Addr:      *addr,
			PublicKey: TestPublicKey,
		},
	}
	sceData, err := serverCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SERVER-0.0.0.0:13", string(sceData))
	assert.NoError(t, err)

	packet := transport.SessionUpdatePacket{
		Sequence:      13,
		ServerAddress: net.UDPAddr{IP: net.IPv4zero, Port: 13},

		ClientRoutePublicKey: make([]byte, crypto.KeySize),

		Signature: make([]byte, ed25519.SignatureSize),
	}
	packet.Signature = crypto.Sign(buyersServerPrivKey, packet.GetSignData())

	data, err := packet.MarshalBinary()
	assert.NoError(t, err)

	var resbuf bytes.Buffer

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, &db, nil, nil, nil, TestPrivateKey, nil)
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	assert.Greater(t, resbuf.Len(), 0)

	var actual transport.SessionResponsePacket
	err = actual.UnmarshalBinary(resbuf.Bytes())
	assert.NoError(t, err)

	assert.Equal(t, int(actual.RouteType), routing.RouteTypeDirect)
}
