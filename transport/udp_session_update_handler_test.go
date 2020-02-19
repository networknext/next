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

func ValidateDirectResponsePacket(resbuf bytes.Buffer, t *testing.T) {
	assert.Greater(t, resbuf.Len(), 0)

	var actual transport.SessionResponsePacket
	err := actual.UnmarshalBinary(resbuf.Bytes())
	assert.NoError(t, err)

	verified := crypto.Verify(TestServerPublicKey, actual.GetSignData(), actual.Signature)
	assert.True(t, verified)

	assert.Equal(t, int(actual.RouteType), routing.RouteTypeDirect)
}

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
			PublicKey: TestServerPublicKey,
		},
	}
	sceData, err := serverCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SERVER-0.0.0.0:13", string(sceData))
	assert.NoError(t, err)

	packet := transport.SessionUpdatePacket{
		Sequence:             13,
		ServerAddress:        net.UDPAddr{IP: net.IPv4zero, Port: 13},
		ClientRoutePublicKey: make([]byte, crypto.KeySize),
	}

	packet.Signature = crypto.Sign(TestBuyerPrivateKey, packet.GetSignData())

	data, err := packet.MarshalBinary()
	assert.NoError(t, err)

	var resbuf bytes.Buffer

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, &db, nil, nil, nil, TestServerPrivateKey, nil)
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	ValidateDirectResponsePacket(resbuf, t)
}

func TestVerificationFailed(t *testing.T) {
	redisServer, _ := miniredis.Run()
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	db := storage.InMemory{
		LocalBuyer: &routing.Buyer{
			PublicKey: TestServerPublicKey, // normally would be buyers public key, intentionally using servers to cause error
		},
	}

	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
	assert.NoError(t, err)

	serverCacheEntry := transport.ServerCacheEntry{
		Sequence: 13,
		Server: routing.Server{
			Addr:      *addr,
			PublicKey: TestServerPublicKey,
		},
	}
	sceData, err := serverCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SERVER-0.0.0.0:13", string(sceData))
	assert.NoError(t, err)

	packet := transport.SessionUpdatePacket{
		Sequence:             13,
		ServerAddress:        net.UDPAddr{IP: net.IPv4zero, Port: 13},
		ClientRoutePublicKey: make([]byte, crypto.KeySize),
	}
	packet.Signature = crypto.Sign(TestBuyerPrivateKey, packet.GetSignData())

	data, err := packet.MarshalBinary()
	assert.NoError(t, err)

	var resbuf bytes.Buffer

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, &db, nil, nil, nil, TestServerPrivateKey, nil)
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	ValidateDirectResponsePacket(resbuf, t)
}

func TestPacketSequenceTooOld(t *testing.T) {
	redisServer, _ := miniredis.Run()
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	db := storage.InMemory{
		LocalBuyer: &routing.Buyer{
			PublicKey: TestBuyerPublicKey,
		},
	}

	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
	assert.NoError(t, err)

	serverCacheEntry := transport.ServerCacheEntry{
		Sequence: 13,
		Server: routing.Server{
			Addr:      *addr,
			PublicKey: TestServerPublicKey,
		},
	}
	sceData, err := serverCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SERVER-0.0.0.0:13", string(sceData))
	assert.NoError(t, err)

	expectedsession := transport.SessionCacheEntry{
		SessionID: 9999,
		Sequence:  13,
	}
	sce, err := expectedsession.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SESSION-9999", string(sce))
	assert.NoError(t, err)

	packet := transport.SessionUpdatePacket{
		SessionId:     9999,
		Sequence:      1,
		ServerAddress: net.UDPAddr{IP: net.IPv4zero, Port: 13},

		ClientRoutePublicKey: make([]byte, crypto.KeySize),

		Signature: make([]byte, ed25519.SignatureSize),
	}
	packet.Signature = crypto.Sign(TestBuyerPrivateKey, packet.GetSignData())

	data, err := packet.MarshalBinary()
	assert.NoError(t, err)

	var resbuf bytes.Buffer

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, &db, nil, nil, nil, TestServerPrivateKey, nil)
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	ValidateDirectResponsePacket(resbuf, t)
}
