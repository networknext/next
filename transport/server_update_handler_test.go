package transport_test

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"net"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-kit/kit/log"
	"github.com/go-redis/redis/v7"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/metrics"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport"
	"github.com/stretchr/testify/assert"
)

func TestFailToUnmarshalServerUpdate(t *testing.T) {
	redisServer, _ := miniredis.Run()
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
	assert.NoError(t, err)

	updateMetrics := metrics.EmptyServerUpdateMetrics
	localMetrics := metrics.LocalHandler{}

	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
	assert.NoError(t, err)

	updateMetrics.ErrorMetrics.UnmarshalFailure = metric

	handler := transport.ServerUpdateHandlerFunc(log.NewNopLogger(), redisClient, nil, &updateMetrics)
	handler(&bytes.Buffer{}, &transport.UDPPacket{SourceAddr: addr, Data: []byte("this is not a proper packet")})

	_, err = redisServer.Get("SERVER-0.0.0.0:13")
	assert.Error(t, err)

	assert.Equal(t, 1.0, updateMetrics.ErrorMetrics.UnmarshalFailure.Value())
}

func TestSDKVersionTooOld(t *testing.T) {
	redisServer, _ := miniredis.Run()
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
	assert.NoError(t, err)

	updateMetrics := metrics.EmptyServerUpdateMetrics
	localMetrics := metrics.LocalHandler{}

	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
	assert.NoError(t, err)

	updateMetrics.ErrorMetrics.SDKTooOld = metric

	packet := transport.ServerUpdatePacket{
		Sequence:             13,
		ServerAddress:        net.UDPAddr{IP: net.IPv4zero, Port: 13},
		ServerPrivateAddress: net.UDPAddr{IP: net.IPv4zero, Port: 13},
		ServerRoutePublicKey: TestServerBackendPublicKey,

		DatacenterID: 13,

		Version: transport.SDKVersion{1, 2, 3},

		Signature: make([]byte, ed25519.SignatureSize),
	}

	data, err := packet.MarshalBinary()
	assert.NoError(t, err)

	handler := transport.ServerUpdateHandlerFunc(log.NewNopLogger(), redisClient, nil, &updateMetrics)
	handler(&bytes.Buffer{}, &transport.UDPPacket{SourceAddr: addr, Data: data})

	_, err = redisServer.Get("SERVER-0.0.0.0:13")
	assert.Error(t, err)

	assert.Equal(t, 1.0, updateMetrics.ErrorMetrics.SDKTooOld.Value())
}

func TestBuyerNotFound(t *testing.T) {
	redisServer, _ := miniredis.Run()
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	db := storage.InMemory{}

	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
	assert.NoError(t, err)

	updateMetrics := metrics.EmptyServerUpdateMetrics
	localMetrics := metrics.LocalHandler{}

	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
	assert.NoError(t, err)

	updateMetrics.ErrorMetrics.BuyerNotFound = metric

	packet := transport.ServerUpdatePacket{
		Sequence:             13,
		ServerAddress:        net.UDPAddr{IP: net.IPv4zero, Port: 13},
		ServerPrivateAddress: net.UDPAddr{IP: net.IPv4zero, Port: 13},
		ServerRoutePublicKey: TestServerBackendPublicKey,

		DatacenterID: 13,

		Version: transport.SDKVersionMin,

		Signature: make([]byte, ed25519.SignatureSize),
	}

	data, err := packet.MarshalBinary()
	assert.NoError(t, err)

	handler := transport.ServerUpdateHandlerFunc(log.NewNopLogger(), redisClient, &db, &updateMetrics)
	handler(&bytes.Buffer{}, &transport.UDPPacket{SourceAddr: addr, Data: data})

	_, err = redisServer.Get("SERVER-0.0.0.0:13")
	assert.Error(t, err)

	assert.Equal(t, 1.0, updateMetrics.ErrorMetrics.BuyerNotFound.Value())
}

func TestVerificationFailure(t *testing.T) {
	_, buyersServerPrivKey, err := ed25519.GenerateKey(nil)
	assert.NoError(t, err)

	redisServer, _ := miniredis.Run()
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	db := storage.InMemory{
		LocalBuyer: &routing.Buyer{},
	}

	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
	assert.NoError(t, err)

	updateMetrics := metrics.EmptyServerUpdateMetrics
	localMetrics := metrics.LocalHandler{}

	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
	assert.NoError(t, err)

	updateMetrics.ErrorMetrics.VerificationFailure = metric

	packet := transport.ServerUpdatePacket{
		Sequence:             13,
		ServerAddress:        net.UDPAddr{IP: net.IPv4zero, Port: 13},
		ServerPrivateAddress: net.UDPAddr{IP: net.IPv4zero, Port: 13},
		ServerRoutePublicKey: TestServerBackendPublicKey,

		DatacenterID: 13,

		Version: transport.SDKVersionMin,

		Signature: make([]byte, ed25519.SignatureSize),
	}
	packet.Signature = crypto.Sign(buyersServerPrivKey, packet.GetSignData())

	data, err := packet.MarshalBinary()
	assert.NoError(t, err)

	handler := transport.ServerUpdateHandlerFunc(log.NewNopLogger(), redisClient, &db, &updateMetrics)
	handler(&bytes.Buffer{}, &transport.UDPPacket{SourceAddr: addr, Data: data})

	_, err = redisServer.Get("SERVER-0.0.0.0:13")
	assert.Error(t, err)

	assert.Equal(t, 1.0, updateMetrics.ErrorMetrics.VerificationFailure.Value())
}

func TestServerPacketSequenceTooOld(t *testing.T) {
	buyersServerPubKey, buyersServerPrivKey, err := ed25519.GenerateKey(nil)
	assert.NoError(t, err)

	redisServer, _ := miniredis.Run()
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	db := storage.InMemory{
		LocalBuyer: &routing.Buyer{
			PublicKey: buyersServerPubKey,
		},
	}

	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
	assert.NoError(t, err)

	updateMetrics := metrics.EmptyServerUpdateMetrics
	localMetrics := metrics.LocalHandler{}

	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
	assert.NoError(t, err)

	updateMetrics.ErrorMetrics.PacketSequenceTooOld = metric

	packet := transport.ServerUpdatePacket{
		Sequence:             0,
		ServerAddress:        net.UDPAddr{IP: net.IPv4zero, Port: 13},
		ServerPrivateAddress: net.UDPAddr{IP: net.IPv4zero, Port: 13},
		ServerRoutePublicKey: TestServerBackendPublicKey,

		DatacenterID: 13,

		Version: transport.SDKVersionMin,
	}
	packet.Signature = crypto.Sign(buyersServerPrivKey, packet.GetSignData())

	data, err := packet.MarshalBinary()
	assert.NoError(t, err)

	expected := transport.ServerCacheEntry{
		Sequence: 1,
	}
	se, err := expected.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SERVER-0.0.0.0:13", string(se))
	assert.NoError(t, err)

	handler := transport.ServerUpdateHandlerFunc(log.NewNopLogger(), redisClient, &db, &updateMetrics)
	handler(&bytes.Buffer{}, &transport.UDPPacket{SourceAddr: addr, Data: data})

	ds, err := redisServer.Get("SERVER-0.0.0.0:13")
	assert.NoError(t, err)

	var actual transport.ServerCacheEntry
	err = actual.UnmarshalBinary([]byte(ds))
	assert.NoError(t, err)

	assert.Equal(t, expected.Sequence, actual.Sequence)

	assert.Equal(t, 1.0, updateMetrics.ErrorMetrics.PacketSequenceTooOld.Value())
}

func TestSuccessfulUpdate(t *testing.T) {
	buyersServerPubKey, buyersServerPrivKey, err := ed25519.GenerateKey(nil)
	assert.NoError(t, err)

	// Get an in-memory redis server and a client that is connected to it
	redisServer, _ := miniredis.Run()
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	db := storage.InMemory{
		LocalBuyer: &routing.Buyer{
			PublicKey: buyersServerPubKey,
		},
	}

	// Create a ServerUpdatePacket and marshal it to binary so sent it into the UDP handler
	packet := transport.ServerUpdatePacket{
		Sequence:             13,
		ServerAddress:        net.UDPAddr{IP: net.IPv4zero, Port: 13},
		ServerPrivateAddress: net.UDPAddr{IP: net.IPv4zero, Port: 13},
		ServerRoutePublicKey: make([]byte, 32),

		DatacenterID: 13,

		Version: transport.SDKVersionMin,
	}
	packet.Signature = crypto.Sign(buyersServerPrivKey, packet.GetSignData())

	data, err := packet.MarshalBinary()
	assert.NoError(t, err)

	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
	assert.NoError(t, err)

	// Create an in-memory buffer to give to the hander since it implements io.Writer
	var buf bytes.Buffer

	// Create a UDPPacket for the handler
	incoming := transport.UDPPacket{
		SourceAddr: addr,
		Data:       data,
	}

	// Initialize the UDP handler with the required redis client
	handler := transport.ServerUpdateHandlerFunc(log.NewNopLogger(), redisClient, &db, &metrics.EmptyServerUpdateMetrics)

	// Invoke the handler with the data packet and address it is coming from
	handler(&buf, &incoming)

	// Get the server entry directly from the in-memory redis and assert there is no error
	ds, err := redisServer.Get("SERVER-0.0.0.0:13")
	assert.NoError(t, err)

	// Create an "expected" ServerEntry based on the incoming ServerUpdatePacket above
	expected := transport.ServerCacheEntry{
		Sequence:   13,
		Server:     routing.Server{Addr: *addr, PublicKey: packet.ServerRoutePublicKey},
		Datacenter: routing.Datacenter{ID: packet.DatacenterID},
		SDKVersion: packet.Version,
	}

	// Unmarshal the data in redis to the actual ServerEntry saved
	var actual transport.ServerCacheEntry
	err = actual.UnmarshalBinary([]byte(ds))
	assert.NoError(t, err)

	// Finally compare both ServerEntry struct to ensure we saved the right data in redis
	assert.Equal(t, expected, actual)

	assert.Equal(t, 0, buf.Len())
}
