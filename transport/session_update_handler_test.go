package transport_test

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"errors"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-kit/kit/log"
	"github.com/go-redis/redis/v7"
	"github.com/networknext/backend/billing"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/metrics"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport"
	"github.com/stretchr/testify/assert"
)

type badWriter struct {
	data []byte
}

func (bad *badWriter) Write(p []byte) (int, error) {
	bad.data = []byte("bad wri")
	return len(bad.data), errors.New("bad write")
}

func (bad *badWriter) Bytes() []byte {
	return bad.data
}

type badBiller struct{}

func (bad *badBiller) Bill(ctx context.Context, sessionID uint64, entry *billing.Entry) error {
	return errors.New("bad bill")
}

func validateDirectResponsePacket(t *testing.T, resbuf bytes.Buffer, directCounter metrics.Counter, reasonCounter metrics.Counter) transport.SessionResponsePacket {
	assert.Greater(t, resbuf.Len(), 0)

	var actual transport.SessionResponsePacket
	err := actual.UnmarshalBinary(resbuf.Bytes())
	assert.NoError(t, err)

	verified := crypto.Verify(TestServerBackendPublicKey, actual.GetSignData(), actual.Signature)
	assert.True(t, verified)

	assert.Equal(t, int(actual.RouteType), routing.RouteTypeDirect)

	assert.Equal(t, 1.0, directCounter.Value())
	assert.Equal(t, 1.0, reasonCounter.Value())

	return actual
}

func validateNextResponsePacket(t *testing.T, resbuf bytes.Buffer, sessionID uint64, sequence uint64, numTokens int32, routeType int32, nextCounter metrics.Counter, reasonCounter metrics.Counter) transport.SessionResponsePacket {
	assert.Greater(t, resbuf.Len(), 0)

	var actual transport.SessionResponsePacket
	err := actual.UnmarshalBinary(resbuf.Bytes())
	assert.NoError(t, err)

	verified := crypto.Verify(TestServerBackendPublicKey, actual.GetSignData(), actual.Signature)
	assert.True(t, verified)

	assert.Equal(t, sessionID, actual.SessionID)
	assert.Equal(t, sequence, actual.Sequence)
	assert.Equal(t, routeType, actual.RouteType)
	assert.Equal(t, numTokens, actual.NumTokens)
	assert.Equal(t, TestBuyersServerPublicKey[:], actual.ServerRoutePublicKey)

	assert.Equal(t, 1.0, nextCounter.Value())
	assert.Equal(t, 1.0, reasonCounter.Value())

	return actual
}

func TestFailToUnmarshalPacket(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
	assert.NoError(t, err)

	sessionMetrics := metrics.EmptySessionMetrics
	localMetrics := metrics.LocalHandler{}

	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
	assert.NoError(t, err)

	sessionMetrics.ErrorMetrics.ReadPacketFailure = metric

	var resbuf bytes.Buffer

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, nil, nil, nil, nil, &sessionMetrics, nil, nil, nil)
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: []byte("this is not a proper packet")})

	assert.Equal(t, 0, resbuf.Len())

	assert.Equal(t, 1.0, sessionMetrics.ErrorMetrics.ReadPacketFailure.Value())
}

func TestFallbackToDirect(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
	assert.NoError(t, err)

	sessionMetrics := metrics.EmptySessionMetrics
	localMetrics := metrics.LocalHandler{}

	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
	assert.NoError(t, err)

	sessionMetrics.ErrorMetrics.FallbackToDirect = metric

	packet := transport.SessionUpdatePacket{
		Sequence:             13,
		ServerAddress:        net.UDPAddr{IP: net.IPv4zero, Port: 13},
		ClientRoutePublicKey: make([]byte, crypto.KeySize),
		FallbackToDirect:     true,
	}

	packet.Signature = crypto.Sign(TestBuyersServerPrivateKey, packet.GetSignData())

	data, err := packet.MarshalBinary()
	assert.NoError(t, err)

	var resbuf bytes.Buffer

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, nil, nil, nil, nil, &sessionMetrics, nil, nil, nil)
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	assert.Equal(t, 0, resbuf.Len())

	assert.Equal(t, 1.0, sessionMetrics.ErrorMetrics.FallbackToDirect.Value())
}

func TestFailToExecPipeline(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	redisServer.Close() // Close the server so that the pipeline fails to execute

	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
	assert.NoError(t, err)

	sessionMetrics := metrics.EmptySessionMetrics
	localMetrics := metrics.LocalHandler{}

	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
	assert.NoError(t, err)

	sessionMetrics.ErrorMetrics.PipelineExecFailure = metric

	packet := transport.SessionUpdatePacket{
		Sequence:             13,
		ServerAddress:        net.UDPAddr{IP: net.IPv4zero, Port: 13},
		ClientRoutePublicKey: make([]byte, crypto.KeySize),
	}

	packet.Signature = crypto.Sign(TestBuyersServerPrivateKey, packet.GetSignData())

	data, err := packet.MarshalBinary()
	assert.NoError(t, err)

	var resbuf bytes.Buffer

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, nil, nil, nil, nil, &sessionMetrics, nil, nil, nil)
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	assert.Equal(t, 0, resbuf.Len())

	assert.Equal(t, 1.0, sessionMetrics.ErrorMetrics.PipelineExecFailure.Value())
}

func TestFailToGetServerData(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
	assert.NoError(t, err)

	sessionMetrics := metrics.EmptySessionMetrics
	localMetrics := metrics.LocalHandler{}

	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
	assert.NoError(t, err)

	sessionMetrics.ErrorMetrics.GetServerDataFailure = metric

	packet := transport.SessionUpdatePacket{
		Sequence:             13,
		ServerAddress:        net.UDPAddr{IP: net.IPv4zero, Port: 13},
		ClientRoutePublicKey: make([]byte, crypto.KeySize),
	}

	packet.Signature = crypto.Sign(TestBuyersServerPrivateKey, packet.GetSignData())

	data, err := packet.MarshalBinary()
	assert.NoError(t, err)

	var resbuf bytes.Buffer

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, nil, nil, nil, nil, &sessionMetrics, nil, nil, nil)
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	assert.Equal(t, 1.0, sessionMetrics.ErrorMetrics.GetServerDataFailure.Value())
}

func TestFailToUnmarshalServerData(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
	assert.NoError(t, err)

	sessionMetrics := metrics.EmptySessionMetrics
	localMetrics := metrics.LocalHandler{}

	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
	assert.NoError(t, err)

	sessionMetrics.ErrorMetrics.UnmarshalServerDataFailure = metric

	err = redisServer.Set("SERVER-0-0.0.0.0:13", "Invalid server data")
	assert.NoError(t, err)

	packet := transport.SessionUpdatePacket{
		Sequence:             13,
		ServerAddress:        net.UDPAddr{IP: net.IPv4zero, Port: 13},
		ClientRoutePublicKey: make([]byte, crypto.KeySize),
	}

	packet.Signature = crypto.Sign(TestBuyersServerPrivateKey, packet.GetSignData())

	data, err := packet.MarshalBinary()
	assert.NoError(t, err)

	var resbuf bytes.Buffer

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, nil, nil, nil, nil, &sessionMetrics, nil, nil, nil)
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	assert.Equal(t, 1.0, sessionMetrics.ErrorMetrics.UnmarshalServerDataFailure.Value())
}

func TestFailToUnmarshalSessionData(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	sessionMetrics := metrics.EmptySessionMetrics
	localMetrics := metrics.LocalHandler{}

	errMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "err metric"})
	assert.NoError(t, err)
	directMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "direct metric"})
	assert.NoError(t, err)

	sessionMetrics.ErrorMetrics.UnmarshalSessionDataFailure = errMetric
	sessionMetrics.DirectSessions = directMetric

	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
	assert.NoError(t, err)

	serverCacheEntry := transport.ServerCacheEntry{
		Sequence: 13,
		Server: routing.Server{
			Addr:      *addr,
			PublicKey: TestServerBackendPublicKey,
		},
	}
	serverCacheEntryData, err := serverCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SERVER-0-0.0.0.0:13", string(serverCacheEntryData))
	assert.NoError(t, err)

	err = redisServer.Set("SESSION-0-9999", "Invalid session data")
	assert.NoError(t, err)

	packet := transport.SessionUpdatePacket{
		Sequence:             13,
		SessionID:            9999,
		ServerAddress:        net.UDPAddr{IP: net.IPv4zero, Port: 13},
		ClientRoutePublicKey: make([]byte, crypto.KeySize),
	}

	packet.Signature = crypto.Sign(TestBuyersServerPrivateKey, packet.GetSignData())

	data, err := packet.MarshalBinary()
	assert.NoError(t, err)

	var resbuf bytes.Buffer

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, nil, nil, nil, nil, &sessionMetrics, nil, TestServerBackendPrivateKey, nil)
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	validateDirectResponsePacket(t, resbuf, sessionMetrics.DirectSessions, sessionMetrics.ErrorMetrics.UnmarshalSessionDataFailure)
}

func TestNoBuyerFound(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	sessionMetrics := metrics.EmptySessionMetrics
	localMetrics := metrics.LocalHandler{}

	errMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "err metric"})
	assert.NoError(t, err)
	directMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "direct metric"})
	assert.NoError(t, err)

	sessionMetrics.ErrorMetrics.BuyerNotFound = errMetric
	sessionMetrics.DirectSessions = directMetric

	db := storage.InMemory{}

	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
	assert.NoError(t, err)

	serverCacheEntry := transport.ServerCacheEntry{
		Sequence: 13,
		Server: routing.Server{
			Addr:      *addr,
			PublicKey: TestServerBackendPublicKey,
		},
	}
	serverCacheEntryData, err := serverCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SERVER-0-0.0.0.0:13", string(serverCacheEntryData))
	assert.NoError(t, err)

	packet := transport.SessionUpdatePacket{
		Sequence:             13,
		ServerAddress:        net.UDPAddr{IP: net.IPv4zero, Port: 13},
		ClientRoutePublicKey: make([]byte, crypto.KeySize),
	}

	packet.Signature = crypto.Sign(TestBuyersServerPrivateKey, packet.GetSignData())

	data, err := packet.MarshalBinary()
	assert.NoError(t, err)

	var resbuf bytes.Buffer

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, &db, nil, nil, nil, &sessionMetrics, nil, TestServerBackendPrivateKey, nil)
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	validateDirectResponsePacket(t, resbuf, sessionMetrics.DirectSessions, sessionMetrics.ErrorMetrics.BuyerNotFound)
}

func TestVerificationFailed(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	sessionMetrics := metrics.EmptySessionMetrics
	localMetrics := metrics.LocalHandler{}

	errMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "err metric"})
	assert.NoError(t, err)
	directMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "direct metric"})
	assert.NoError(t, err)

	sessionMetrics.ErrorMetrics.VerifyFailure = errMetric
	sessionMetrics.DirectSessions = directMetric

	db := storage.InMemory{}
	db.AddBuyer(context.Background(), routing.Buyer{
		PublicKey: TestServerBackendPublicKey, // normally would be buyers public key, intentionally using servers to cause error
	})

	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
	assert.NoError(t, err)

	serverCacheEntry := transport.ServerCacheEntry{
		Sequence: 13,
		Server: routing.Server{
			Addr:      *addr,
			PublicKey: TestServerBackendPublicKey,
		},
	}
	serverCacheEntryData, err := serverCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SERVER-0-0.0.0.0:13", string(serverCacheEntryData))
	assert.NoError(t, err)

	packet := transport.SessionUpdatePacket{
		Sequence:             13,
		ServerAddress:        net.UDPAddr{IP: net.IPv4zero, Port: 13},
		ClientRoutePublicKey: make([]byte, crypto.KeySize),
	}
	packet.Signature = crypto.Sign(TestBuyersServerPrivateKey, packet.GetSignData())

	data, err := packet.MarshalBinary()
	assert.NoError(t, err)

	var resbuf bytes.Buffer

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, &db, nil, nil, nil, &sessionMetrics, nil, TestServerBackendPrivateKey, nil)
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	validateDirectResponsePacket(t, resbuf, sessionMetrics.DirectSessions, sessionMetrics.ErrorMetrics.VerifyFailure)
}

func TestSessionPacketSequenceTooOld(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	sessionMetrics := metrics.EmptySessionMetrics
	localMetrics := metrics.LocalHandler{}

	errMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "err metric"})
	assert.NoError(t, err)
	directMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "direct metric"})
	assert.NoError(t, err)

	sessionMetrics.ErrorMetrics.OldSequence = errMetric
	sessionMetrics.DirectSessions = directMetric

	db := storage.InMemory{}
	db.AddBuyer(context.Background(), routing.Buyer{
		PublicKey: TestBuyersServerPublicKey,
	})

	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
	assert.NoError(t, err)

	serverCacheEntry := transport.ServerCacheEntry{
		Sequence: 13,
		Server: routing.Server{
			Addr:      *addr,
			PublicKey: TestServerBackendPublicKey,
		},
	}
	serverCacheEntryData, err := serverCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SERVER-0-0.0.0.0:13", string(serverCacheEntryData))
	assert.NoError(t, err)

	sessionCacheEntry := transport.SessionCacheEntry{
		SessionID: 9999,
		Sequence:  13,
	}
	sessionCacheEntryData, err := sessionCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SESSION-0-9999", string(sessionCacheEntryData))
	assert.NoError(t, err)

	packet := transport.SessionUpdatePacket{
		SessionID:     9999,
		Sequence:      1,
		ServerAddress: net.UDPAddr{IP: net.IPv4zero, Port: 13},

		ClientRoutePublicKey: make([]byte, crypto.KeySize),

		Signature: make([]byte, ed25519.SignatureSize),
	}
	packet.Signature = crypto.Sign(TestBuyersServerPrivateKey, packet.GetSignData())

	data, err := packet.MarshalBinary()
	assert.NoError(t, err)

	var resbuf bytes.Buffer

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, &db, nil, nil, nil, &sessionMetrics, nil, TestServerBackendPrivateKey, nil)
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	validateDirectResponsePacket(t, resbuf, sessionMetrics.DirectSessions, sessionMetrics.ErrorMetrics.OldSequence)
}

func TestBadWriteCachedResponse(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	sessionMetrics := metrics.EmptySessionMetrics
	localMetrics := metrics.LocalHandler{}

	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
	assert.NoError(t, err)

	sessionMetrics.ErrorMetrics.WriteCachedResponseFailure = metric

	db := storage.InMemory{}
	db.AddBuyer(context.Background(), routing.Buyer{
		PublicKey: TestBuyersServerPublicKey,
	})

	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
	assert.NoError(t, err)

	serverCacheEntry := transport.ServerCacheEntry{
		Sequence: 13,
		Server: routing.Server{
			Addr:      *addr,
			PublicKey: TestServerBackendPublicKey,
		},
	}
	serverCacheEntryData, err := serverCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SERVER-0-0.0.0.0:13", string(serverCacheEntryData))
	assert.NoError(t, err)

	sessionCacheEntry := transport.SessionCacheEntry{
		SessionID: 9999,
		Sequence:  13,
		Response:  []byte{1},
	}
	sessionCacheEntryData, err := sessionCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SESSION-0-9999", string(sessionCacheEntryData))
	assert.NoError(t, err)

	packet := transport.SessionUpdatePacket{
		SessionID:     9999,
		Sequence:      13,
		ServerAddress: net.UDPAddr{IP: net.IPv4zero, Port: 13},

		ClientRoutePublicKey: make([]byte, crypto.KeySize),

		Signature: make([]byte, ed25519.SignatureSize),
	}
	packet.Signature = crypto.Sign(TestBuyersServerPrivateKey, packet.GetSignData())

	data, err := packet.MarshalBinary()
	assert.NoError(t, err)

	var badWriter badWriter

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, &db, nil, nil, nil, &sessionMetrics, nil, TestServerBackendPrivateKey, nil)
	handler(&badWriter, &transport.UDPPacket{SourceAddr: addr, Data: data})

	var actual transport.SessionResponsePacket
	err = actual.UnmarshalBinary(badWriter.Bytes())
	assert.Error(t, err)

	assert.Equal(t, 1.0, sessionMetrics.ErrorMetrics.WriteCachedResponseFailure.Value())
}

func TestClientIPLookupFail(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	sessionMetrics := metrics.EmptySessionMetrics
	localMetrics := metrics.LocalHandler{}

	errMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "err metric"})
	assert.NoError(t, err)
	directMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "direct metric"})
	assert.NoError(t, err)

	sessionMetrics.ErrorMetrics.ClientLocateFailure = errMetric
	sessionMetrics.DirectSessions = directMetric

	db := storage.InMemory{}
	db.AddBuyer(context.Background(), routing.Buyer{
		PublicKey: TestBuyersServerPublicKey,
	})

	iploc := routing.LocateIPFunc(func(ip net.IP) (routing.Location, error) {
		return routing.Location{}, errors.New("nope")
	})

	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
	assert.NoError(t, err)

	serverCacheEntry := transport.ServerCacheEntry{
		Sequence: 13,
		Server: routing.Server{
			Addr:      *addr,
			PublicKey: TestServerBackendPublicKey,
		},
	}
	serverCacheEntryData, err := serverCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SERVER-0-0.0.0.0:13", string(serverCacheEntryData))
	assert.NoError(t, err)

	sessionCacheEntry := transport.SessionCacheEntry{
		SessionID: 9999,
		Sequence:  13,
	}
	sessionCacheEntryData, err := sessionCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SESSION-0-9999", string(sessionCacheEntryData))
	assert.NoError(t, err)

	packet := transport.SessionUpdatePacket{
		SessionID:     9999,
		Sequence:      14,
		ServerAddress: net.UDPAddr{IP: net.IPv4zero, Port: 13},

		ClientRoutePublicKey: make([]byte, crypto.KeySize),

		Signature: make([]byte, ed25519.SignatureSize),
	}
	packet.Signature = crypto.Sign(TestBuyersServerPrivateKey, packet.GetSignData())

	data, err := packet.MarshalBinary()
	assert.NoError(t, err)

	var resbuf bytes.Buffer

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, &db, nil, &iploc, nil, &sessionMetrics, nil, TestServerBackendPrivateKey, nil)
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	validateDirectResponsePacket(t, resbuf, sessionMetrics.DirectSessions, sessionMetrics.ErrorMetrics.ClientLocateFailure)
}

func TestNoRelaysNearClient(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	sessionMetrics := metrics.EmptySessionMetrics
	localMetrics := metrics.LocalHandler{}

	errMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "err metric"})
	assert.NoError(t, err)
	directMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "direct metric"})
	assert.NoError(t, err)

	sessionMetrics.ErrorMetrics.NearRelaysLocateFailure = errMetric
	sessionMetrics.DirectSessions = directMetric

	db := storage.InMemory{}
	db.AddBuyer(context.Background(), routing.Buyer{
		PublicKey: TestBuyersServerPublicKey,
	})

	iploc := routing.LocateIPFunc(func(ip net.IP) (routing.Location, error) {
		return routing.Location{
			Continent: "NA",
			Country:   "US",
			Region:    "NY",
			City:      "Troy",
			Latitude:  0,
			Longitude: 0,
		}, nil
	})

	geoClient := routing.GeoClient{
		RedisClient: redisClient,
		Namespace:   "GEO_TEST",
	}

	rp := mockRouteProvider{}

	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
	assert.NoError(t, err)

	serverCacheEntry := transport.ServerCacheEntry{
		Sequence: 13,
		Server: routing.Server{
			Addr:      *addr,
			PublicKey: TestServerBackendPublicKey,
		},
	}
	serverCacheEntryData, err := serverCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SERVER-0-0.0.0.0:13", string(serverCacheEntryData))
	assert.NoError(t, err)

	sessionCacheEntry := transport.SessionCacheEntry{
		SessionID: 9999,
		Sequence:  13,
	}
	sessionCacheEntryData, err := sessionCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SESSION-0-9999", string(sessionCacheEntryData))
	assert.NoError(t, err)

	packet := transport.SessionUpdatePacket{
		SessionID:     9999,
		Sequence:      14,
		ServerAddress: net.UDPAddr{IP: net.IPv4zero, Port: 13},

		ClientRoutePublicKey: make([]byte, crypto.KeySize),

		Signature: make([]byte, ed25519.SignatureSize),
	}
	packet.Signature = crypto.Sign(TestBuyersServerPrivateKey, packet.GetSignData())

	data, err := packet.MarshalBinary()
	assert.NoError(t, err)

	var resbuf bytes.Buffer

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, &db, &rp, &iploc, &geoClient, &sessionMetrics, &billing.NoOpBiller{}, TestServerBackendPrivateKey, nil)
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	validateDirectResponsePacket(t, resbuf, sessionMetrics.DirectSessions, sessionMetrics.ErrorMetrics.NearRelaysLocateFailure)
}

func TestNoRoutesFound(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	sessionMetrics := metrics.EmptySessionMetrics
	localMetrics := metrics.LocalHandler{}

	errMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "err metric"})
	assert.NoError(t, err)
	directMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "direct metric"})
	assert.NoError(t, err)

	sessionMetrics.ErrorMetrics.RouteFailure = errMetric
	sessionMetrics.DirectSessions = directMetric

	db := storage.InMemory{}
	db.AddBuyer(context.Background(), routing.Buyer{
		PublicKey:            TestBuyersServerPublicKey,
		RoutingRulesSettings: routing.LocalRoutingRulesSettings,
	})

	iploc := routing.LocateIPFunc(func(ip net.IP) (routing.Location, error) {
		return routing.Location{
			Continent: "NA",
			Country:   "US",
			Region:    "NY",
			City:      "Troy",
			Latitude:  0,
			Longitude: 0,
		}, nil
	})

	geoClient := routing.GeoClient{
		RedisClient: redisClient,
		Namespace:   "GEO_TEST",
	}

	nearbyRelay := routing.Relay{
		ID: 1,
	}
	err = geoClient.Add(nearbyRelay)
	assert.NoError(t, err)

	rp := mockRouteProvider{}

	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
	assert.NoError(t, err)

	serverCacheEntry := transport.ServerCacheEntry{
		Sequence: 13,
		Server: routing.Server{
			Addr:      *addr,
			PublicKey: TestServerBackendPublicKey,
		},
	}
	serverCacheEntryData, err := serverCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SERVER-0-0.0.0.0:13", string(serverCacheEntryData))
	assert.NoError(t, err)

	sessionCacheEntry := transport.SessionCacheEntry{
		SessionID: 9999,
		Sequence:  13,
	}
	sessionCacheEntryData, err := sessionCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SESSION-0-9999", string(sessionCacheEntryData))
	assert.NoError(t, err)

	packet := transport.SessionUpdatePacket{
		SessionID:     9999,
		Sequence:      14,
		ServerAddress: net.UDPAddr{IP: net.IPv4zero, Port: 13},

		ClientRoutePublicKey: make([]byte, crypto.KeySize),

		Signature: make([]byte, ed25519.SignatureSize),
	}
	packet.Signature = crypto.Sign(TestBuyersServerPrivateKey, packet.GetSignData())

	data, err := packet.MarshalBinary()
	assert.NoError(t, err)

	var resbuf bytes.Buffer

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, &db, &rp, &iploc, &geoClient, &sessionMetrics, &billing.NoOpBiller{}, TestServerBackendPrivateKey, nil)
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	validateDirectResponsePacket(t, resbuf, sessionMetrics.DirectSessions, sessionMetrics.ErrorMetrics.RouteFailure)
}

func TestTokenEncryptionFailure(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	sessionMetrics := metrics.EmptySessionMetrics
	localMetrics := metrics.LocalHandler{}

	errMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "err metric"})
	assert.NoError(t, err)
	directMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "direct metric"})
	assert.NoError(t, err)

	sessionMetrics.ErrorMetrics.EncryptionFailure = errMetric
	sessionMetrics.DirectSessions = directMetric

	buyer := routing.Buyer{
		PublicKey: TestBuyersServerPublicKey,
	}
	buyer.RoutingRulesSettings.SelectionPercentage = 100

	db := storage.InMemory{}
	db.AddBuyer(context.Background(), buyer)

	iploc := routing.LocateIPFunc(func(ip net.IP) (routing.Location, error) {
		return routing.Location{
			Continent: "NA",
			Country:   "US",
			Region:    "NY",
			City:      "Troy",
			Latitude:  0,
			Longitude: 0,
		}, nil
	})

	geoClient := routing.GeoClient{
		RedisClient: redisClient,
		Namespace:   "GEO_TEST",
	}

	nearbyRelay := routing.Relay{
		ID: 1,
	}
	err = geoClient.Add(nearbyRelay)
	assert.NoError(t, err)

	rp := mockRouteProvider{
		routes: []routing.Route{
			{
				Relays: []routing.Relay{
					{ID: 1, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 123}, PublicKey: TestRelayPublicKey[:]},
					{ID: 2, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.2"), Port: 123}},
					{ID: 3, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.3"), Port: 123}, PublicKey: TestRelayPublicKey[:]},
				},
			},
		},
	}

	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
	assert.NoError(t, err)

	serverCacheEntry := transport.ServerCacheEntry{
		Sequence: 13,
		Server: routing.Server{
			Addr:      *addr,
			PublicKey: TestBuyersServerPublicKey[:],
		},
	}
	serverCacheEntryData, err := serverCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SERVER-0-0.0.0.0:13", string(serverCacheEntryData))
	assert.NoError(t, err)

	sessionCacheEntry := transport.SessionCacheEntry{
		SessionID: 9999,
		Sequence:  13,
	}
	sessionCacheEntryData, err := sessionCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SESSION-0-9999", string(sessionCacheEntryData))
	assert.NoError(t, err)

	packet := transport.SessionUpdatePacket{
		SessionID:     9999,
		Sequence:      14,
		ServerAddress: net.UDPAddr{IP: net.IPv4zero, Port: 13},

		ClientAddress: net.UDPAddr{
			IP:   net.ParseIP("0.0.0.0"),
			Port: 1234,
		},
		ClientRoutePublicKey: TestBuyersClientPublicKey[:],
	}
	packet.Signature = crypto.Sign(TestBuyersServerPrivateKey, packet.GetSignData())

	data, err := packet.MarshalBinary()
	assert.NoError(t, err)

	var resbuf bytes.Buffer

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, &db, &rp, &iploc, &geoClient, &sessionMetrics, &billing.NoOpBiller{}, TestServerBackendPrivateKey[:], TestRouterPrivateKey[:])
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	assert.Greater(t, resbuf.Len(), 0)

	var actual transport.SessionResponsePacket
	err = actual.UnmarshalBinary(resbuf.Bytes())
	assert.NoError(t, err)

	validateDirectResponsePacket(t, resbuf, sessionMetrics.DirectSessions, sessionMetrics.ErrorMetrics.EncryptionFailure)
}

func TestBadWriteResponse(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	sessionMetrics := metrics.EmptySessionMetrics
	localMetrics := metrics.LocalHandler{}

	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
	assert.NoError(t, err)

	sessionMetrics.ErrorMetrics.WriteResponseFailure = metric

	db := storage.InMemory{}
	db.AddBuyer(context.Background(), routing.Buyer{
		PublicKey: TestBuyersServerPublicKey,
	})

	iploc := routing.LocateIPFunc(func(ip net.IP) (routing.Location, error) {
		return routing.Location{
			Continent: "NA",
			Country:   "US",
			Region:    "NY",
			City:      "Troy",
			Latitude:  0,
			Longitude: 0,
		}, nil
	})

	geoClient := routing.GeoClient{
		RedisClient: redisClient,
		Namespace:   "GEO_TEST",
	}

	nearbyRelay := routing.Relay{
		ID: 1,
	}
	err = geoClient.Add(nearbyRelay)
	assert.NoError(t, err)

	rp := mockRouteProvider{
		routes: []routing.Route{
			{
				Relays: []routing.Relay{
					{ID: 1, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 123}, PublicKey: TestRelayPublicKey[:]},
					{ID: 2, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.2"), Port: 123}, PublicKey: TestRelayPublicKey[:]},
					{ID: 3, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.3"), Port: 123}, PublicKey: TestRelayPublicKey[:]},
				},
			},
		},
	}

	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
	assert.NoError(t, err)

	serverCacheEntry := transport.ServerCacheEntry{
		Sequence: 13,
		Server: routing.Server{
			Addr:      *addr,
			PublicKey: TestBuyersServerPublicKey[:],
		},
	}
	serverCacheEntryData, err := serverCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SERVER-0-0.0.0.0:13", string(serverCacheEntryData))
	assert.NoError(t, err)

	sessionCacheEntry := transport.SessionCacheEntry{
		SessionID:      9999,
		Sequence:       13,
		TimestampStart: time.Now().Add(-5 * time.Second),
	}
	sessionCacheEntryData, err := sessionCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SESSION-0-9999", string(sessionCacheEntryData))
	assert.NoError(t, err)

	packet := transport.SessionUpdatePacket{
		SessionID:     9999,
		Sequence:      14,
		ServerAddress: net.UDPAddr{IP: net.IPv4zero, Port: 13},

		NumNearRelays:       1,
		NearRelayIDs:        []uint64{1},
		NearRelayMinRTT:     []float32{1},
		NearRelayMaxRTT:     []float32{1},
		NearRelayMeanRTT:    []float32{1},
		NearRelayJitter:     []float32{1},
		NearRelayPacketLoss: []float32{1},

		ClientAddress: net.UDPAddr{
			IP:   net.ParseIP("0.0.0.0"),
			Port: 1234,
		},
		ClientRoutePublicKey: TestBuyersClientPublicKey[:],
	}
	packet.Signature = crypto.Sign(TestBuyersServerPrivateKey, packet.GetSignData())

	data, err := packet.MarshalBinary()
	assert.NoError(t, err)

	var badWriter badWriter

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, &db, &rp, &iploc, &geoClient, &sessionMetrics, &billing.NoOpBiller{}, TestServerBackendPrivateKey[:], TestRouterPrivateKey[:])
	handler(&badWriter, &transport.UDPPacket{SourceAddr: addr, Data: data})

	var actual transport.SessionResponsePacket
	err = actual.UnmarshalBinary(badWriter.Bytes())
	assert.Error(t, err)

	assert.Equal(t, 1.0, sessionMetrics.ErrorMetrics.WriteResponseFailure.Value())
}

func TestBillingFailure(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	sessionMetrics := metrics.EmptySessionMetrics
	localMetrics := metrics.LocalHandler{}

	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
	assert.NoError(t, err)

	sessionMetrics.ErrorMetrics.BillingFailure = metric

	db := storage.InMemory{}
	db.AddBuyer(context.Background(), routing.Buyer{
		PublicKey: TestBuyersServerPublicKey,
	})

	iploc := routing.LocateIPFunc(func(ip net.IP) (routing.Location, error) {
		return routing.Location{
			Continent: "NA",
			Country:   "US",
			Region:    "NY",
			City:      "Troy",
			Latitude:  0,
			Longitude: 0,
		}, nil
	})

	geoClient := routing.GeoClient{
		RedisClient: redisClient,
		Namespace:   "GEO_TEST",
	}

	nearbyRelay := routing.Relay{
		ID: 1,
	}
	err = geoClient.Add(nearbyRelay)
	assert.NoError(t, err)

	rp := mockRouteProvider{
		routes: []routing.Route{
			{
				Relays: []routing.Relay{
					{ID: 1, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 123}, PublicKey: TestRelayPublicKey[:]},
					{ID: 2, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.2"), Port: 123}, PublicKey: TestRelayPublicKey[:]},
					{ID: 3, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.3"), Port: 123}, PublicKey: TestRelayPublicKey[:]},
				},
			},
		},
	}

	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
	assert.NoError(t, err)

	serverCacheEntry := transport.ServerCacheEntry{
		Sequence: 13,
		Server: routing.Server{
			Addr:      *addr,
			PublicKey: TestBuyersServerPublicKey[:],
		},
	}
	serverCacheEntryData, err := serverCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SERVER-0-0.0.0.0:13", string(serverCacheEntryData))
	assert.NoError(t, err)

	sessionCacheEntry := transport.SessionCacheEntry{
		SessionID:      9999,
		Sequence:       13,
		TimestampStart: time.Now().Add(-5 * time.Second),
	}
	sessionCacheEntryData, err := sessionCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SESSION-0-9999", string(sessionCacheEntryData))
	assert.NoError(t, err)

	packet := transport.SessionUpdatePacket{
		SessionID:     9999,
		Sequence:      14,
		ServerAddress: net.UDPAddr{IP: net.IPv4zero, Port: 13},

		NumNearRelays:       1,
		NearRelayIDs:        []uint64{1},
		NearRelayMinRTT:     []float32{1},
		NearRelayMaxRTT:     []float32{1},
		NearRelayMeanRTT:    []float32{1},
		NearRelayJitter:     []float32{1},
		NearRelayPacketLoss: []float32{1},

		ClientAddress: net.UDPAddr{
			IP:   net.ParseIP("0.0.0.0"),
			Port: 1234,
		},
		ClientRoutePublicKey: TestBuyersClientPublicKey[:],
	}
	packet.Signature = crypto.Sign(TestBuyersServerPrivateKey, packet.GetSignData())

	data, err := packet.MarshalBinary()
	assert.NoError(t, err)

	var resbuf bytes.Buffer

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, &db, &rp, &iploc, &geoClient, &sessionMetrics, &badBiller{}, TestServerBackendPrivateKey[:], TestRouterPrivateKey[:])
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	var actual transport.SessionResponsePacket
	err = actual.UnmarshalBinary(resbuf.Bytes())
	assert.NoError(t, err)

	assert.Equal(t, 1.0, sessionMetrics.ErrorMetrics.BillingFailure.Value())
}

func TestNextRouteResponse(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	sessionMetrics := metrics.EmptySessionMetrics
	localMetrics := metrics.LocalHandler{}

	decisionMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "decision metric"})
	assert.NoError(t, err)
	nextMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "next metric"})
	assert.NoError(t, err)

	sessionMetrics.DecisionMetrics.RTTReduction = decisionMetric
	sessionMetrics.NextSessions = nextMetric

	db := storage.InMemory{}
	db.AddBuyer(context.Background(), routing.Buyer{
		PublicKey:            TestBuyersServerPublicKey,
		RoutingRulesSettings: routing.LocalRoutingRulesSettings,
	})

	iploc := routing.LocateIPFunc(func(ip net.IP) (routing.Location, error) {
		return routing.Location{
			Continent: "NA",
			Country:   "US",
			Region:    "NY",
			City:      "Troy",
			Latitude:  0,
			Longitude: 0,
		}, nil
	})

	geoClient := routing.GeoClient{
		RedisClient: redisClient,
		Namespace:   "GEO_TEST",
	}

	nearbyRelay := routing.Relay{
		ID: 1,
	}
	err = geoClient.Add(nearbyRelay)
	assert.NoError(t, err)

	rp := mockRouteProvider{
		routes: []routing.Route{
			{
				Relays: []routing.Relay{
					{ID: 1, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 123}, PublicKey: TestRelayPublicKey[:]},
					{ID: 2, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.2"), Port: 123}, PublicKey: TestRelayPublicKey[:]},
					{ID: 3, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.3"), Port: 123}, PublicKey: TestRelayPublicKey[:]},
				},
			},
		},
	}

	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
	assert.NoError(t, err)

	serverCacheEntry := transport.ServerCacheEntry{
		Sequence: 13,
		Server: routing.Server{
			Addr:      *addr,
			PublicKey: TestBuyersServerPublicKey[:],
		},
	}
	serverCacheEntryData, err := serverCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SERVER-0-0.0.0.0:13", string(serverCacheEntryData))
	assert.NoError(t, err)

	sessionCacheEntry := transport.SessionCacheEntry{
		SessionID:      9999,
		Sequence:       13,
		TimestampStart: time.Now().Add(-5 * time.Second),
	}
	sessionCacheEntryData, err := sessionCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SESSION-0-9999", string(sessionCacheEntryData))
	assert.NoError(t, err)

	packet := transport.SessionUpdatePacket{
		SessionID:     9999,
		Sequence:      14,
		ServerAddress: net.UDPAddr{IP: net.IPv4zero, Port: 13},

		NumNearRelays:       1,
		NearRelayIDs:        []uint64{1},
		NearRelayMinRTT:     []float32{1},
		NearRelayMaxRTT:     []float32{1},
		NearRelayMeanRTT:    []float32{1},
		NearRelayJitter:     []float32{1},
		NearRelayPacketLoss: []float32{1},

		OnNetworkNext: true,
		NextMinRTT:    50,
		DirectMinRTT:  70,

		ClientAddress: net.UDPAddr{
			IP:   net.ParseIP("0.0.0.0"),
			Port: 1234,
		},
		ClientRoutePublicKey: TestBuyersClientPublicKey[:],
	}
	packet.Signature = crypto.Sign(TestBuyersServerPrivateKey, packet.GetSignData())

	data, err := packet.MarshalBinary()
	assert.NoError(t, err)

	var resbuf bytes.Buffer

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, &db, &rp, &iploc, &geoClient, &sessionMetrics, &billing.NoOpBiller{}, TestServerBackendPrivateKey[:], TestRouterPrivateKey[:])
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	globalDelta, err := redisServer.ZScore("top-global", fmt.Sprintf("%x", packet.SessionID))
	assert.NoError(t, err)
	assert.Equal(t, globalDelta, float64(20))

	buyerDelta, err := redisServer.ZScore(fmt.Sprintf("top-buyer-%x", packet.CustomerID), fmt.Sprintf("%x", packet.SessionID))
	assert.NoError(t, err)
	assert.Equal(t, buyerDelta, float64(20))

	assert.True(t, redisServer.Exists(fmt.Sprintf("session-%x-meta", packet.SessionID)))
	assert.Greater(t, redisServer.TTL(fmt.Sprintf("session-%x-meta", packet.SessionID)).Hours(), float64(-1))

	assert.True(t, redisServer.Exists(fmt.Sprintf("session-%x-slices", packet.SessionID)))
	assert.Greater(t, redisServer.TTL(fmt.Sprintf("session-%x-slices", packet.SessionID)).Hours(), float64(-1))

	assert.True(t, redisServer.Exists(fmt.Sprintf("user-%x-sessions", 0)))
	assert.Greater(t, redisServer.TTL(fmt.Sprintf("user-%x-sessions", 0)).Hours(), float64(-1))

	validateNextResponsePacket(t, resbuf, packet.SessionID, packet.Sequence, 5, routing.RouteTypeNew, sessionMetrics.NextSessions, sessionMetrics.DecisionMetrics.RTTReduction)
}

func TestContinueRouteResponse(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	sessionMetrics := metrics.EmptySessionMetrics
	localMetrics := metrics.LocalHandler{}

	decisionMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "decision metric"})
	assert.NoError(t, err)
	nextMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "next metric"})
	assert.NoError(t, err)

	sessionMetrics.DecisionMetrics.RTTReduction = decisionMetric
	sessionMetrics.NextSessions = nextMetric

	db := storage.InMemory{}
	db.AddBuyer(context.Background(), routing.Buyer{
		PublicKey:            TestBuyersServerPublicKey,
		RoutingRulesSettings: routing.LocalRoutingRulesSettings,
	})

	iploc := routing.LocateIPFunc(func(ip net.IP) (routing.Location, error) {
		return routing.Location{
			Continent: "NA",
			Country:   "US",
			Region:    "NY",
			City:      "Troy",
			Latitude:  0,
			Longitude: 0,
		}, nil
	})

	geoClient := routing.GeoClient{
		RedisClient: redisClient,
		Namespace:   "GEO_TEST",
	}

	nearbyRelay := routing.Relay{
		ID: 1,
	}
	geoClient.Add(nearbyRelay)

	rp := mockRouteProvider{
		routes: []routing.Route{
			{
				Relays: []routing.Relay{
					{ID: 1, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 123}, PublicKey: TestRelayPublicKey[:]},
					{ID: 2, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.2"), Port: 123}, PublicKey: TestRelayPublicKey[:]},
					{ID: 3, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.3"), Port: 123}, PublicKey: TestRelayPublicKey[:]},
				},
			},
		},
	}

	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
	assert.NoError(t, err)

	expected := transport.ServerCacheEntry{
		Sequence: 13,
		Server: routing.Server{
			Addr:      *addr,
			PublicKey: TestBuyersServerPublicKey[:],
		},
	}
	se, err := expected.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SERVER-0-0.0.0.0:13", string(se))
	assert.NoError(t, err)

	expectedsession := transport.SessionCacheEntry{
		SessionID:      9999,
		Sequence:       13,
		RouteHash:      1511739644222804357,
		TimestampStart: time.Now().Add(-5 * time.Second),
	}
	sce, err := expectedsession.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SESSION-0-9999", string(sce))
	assert.NoError(t, err)

	packet := transport.SessionUpdatePacket{
		SessionID:     9999,
		Sequence:      14,
		ServerAddress: net.UDPAddr{IP: net.IPv4zero, Port: 13},

		NumNearRelays:       1,
		NearRelayIDs:        []uint64{1},
		NearRelayMinRTT:     []float32{1},
		NearRelayMaxRTT:     []float32{1},
		NearRelayMeanRTT:    []float32{1},
		NearRelayJitter:     []float32{1},
		NearRelayPacketLoss: []float32{1},

		OnNetworkNext: true,
		NextMinRTT:    50,
		DirectMinRTT:  70,

		ClientAddress: net.UDPAddr{
			IP:   net.ParseIP("0.0.0.0"),
			Port: 1234,
		},
		ClientRoutePublicKey: TestBuyersClientPublicKey[:],
	}
	packet.Signature = crypto.Sign(TestBuyersServerPrivateKey, packet.GetSignData())

	data, err := packet.MarshalBinary()
	assert.NoError(t, err)

	var resbuf bytes.Buffer

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, &db, &rp, &iploc, &geoClient, &sessionMetrics, &billing.NoOpBiller{}, TestServerBackendPrivateKey[:], TestRouterPrivateKey[:])
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	globalDelta, err := redisServer.ZScore("top-global", fmt.Sprintf("%x", packet.SessionID))
	assert.NoError(t, err)
	assert.Equal(t, globalDelta, float64(20))

	buyerDelta, err := redisServer.ZScore(fmt.Sprintf("top-buyer-%x", packet.CustomerID), fmt.Sprintf("%x", packet.SessionID))
	assert.NoError(t, err)
	assert.Equal(t, buyerDelta, float64(20))

	assert.True(t, redisServer.Exists(fmt.Sprintf("session-%x-meta", packet.SessionID)))
	assert.Greater(t, redisServer.TTL(fmt.Sprintf("session-%x-meta", packet.SessionID)).Hours(), float64(-1))

	assert.True(t, redisServer.Exists(fmt.Sprintf("session-%x-slices", packet.SessionID)))
	assert.Greater(t, redisServer.TTL(fmt.Sprintf("session-%x-slices", packet.SessionID)).Hours(), float64(-1))

	assert.True(t, redisServer.Exists(fmt.Sprintf("user-%x-sessions", 0)))
	assert.Greater(t, redisServer.TTL(fmt.Sprintf("user-%x-sessions", 0)).Hours(), float64(-1))

	var actual transport.SessionResponsePacket
	err = actual.UnmarshalBinary(resbuf.Bytes())
	assert.NoError(t, err)

	verified := crypto.Verify(TestServerBackendPublicKey, actual.GetSignData(), actual.Signature)
	assert.True(t, verified)

	validateNextResponsePacket(t, resbuf, packet.SessionID, packet.Sequence, 5, routing.RouteTypeContinue, sessionMetrics.NextSessions, sessionMetrics.DecisionMetrics.RTTReduction)
}

func TestCachedRouteResponse(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	sessionMetrics := metrics.EmptySessionMetrics
	localMetrics := metrics.LocalHandler{}

	directMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "direct metric"})
	assert.NoError(t, err)
	nextMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "next metric"})
	assert.NoError(t, err)

	sessionMetrics.DirectSessions = directMetric
	sessionMetrics.NextSessions = nextMetric

	db := storage.InMemory{}
	db.AddBuyer(context.Background(), routing.Buyer{
		PublicKey:            TestBuyersServerPublicKey,
		RoutingRulesSettings: routing.LocalRoutingRulesSettings,
	})

	iploc := routing.LocateIPFunc(func(ip net.IP) (routing.Location, error) {
		return routing.Location{
			Continent: "NA",
			Country:   "US",
			Region:    "NY",
			City:      "Troy",
			Latitude:  0,
			Longitude: 0,
		}, nil
	})

	geoClient := routing.GeoClient{
		RedisClient: redisClient,
		Namespace:   "GEO_TEST",
	}

	nearbyRelay := routing.Relay{
		ID: 1,
	}
	err = geoClient.Add(nearbyRelay)
	assert.NoError(t, err)

	rp := mockRouteProvider{
		routes: []routing.Route{
			{
				Relays: []routing.Relay{
					{ID: 1, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 123}, PublicKey: TestRelayPublicKey[:]},
					{ID: 2, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.2"), Port: 123}, PublicKey: TestRelayPublicKey[:]},
					{ID: 3, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.3"), Port: 123}, PublicKey: TestRelayPublicKey[:]},
				},
			},
		},
	}

	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
	assert.NoError(t, err)

	serverCacheEntry := transport.ServerCacheEntry{
		Sequence: 13,
		Server: routing.Server{
			Addr:      *addr,
			PublicKey: TestBuyersServerPublicKey[:],
		},
	}
	serverCacheEntryData, err := serverCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SERVER-0-0.0.0.0:13", string(serverCacheEntryData))
	assert.NoError(t, err)

	tokens := make([]byte, routing.EncryptedNextRouteTokenSize)
	copy(tokens, []byte("TEST TOKENS"))

	cachedRouteResponse := transport.SessionResponsePacket{
		SessionID:            9999,
		Sequence:             13,
		NumNearRelays:        0,
		NearRelayIDs:         make([]uint64, 0),
		NearRelayAddresses:   make([]net.UDPAddr, 0),
		RouteType:            int32(routing.RouteTypeNew),
		NumTokens:            1,
		Tokens:               tokens,
		ServerRoutePublicKey: TestBuyersServerPublicKey[:],
	}
	cachedRouteResponse.Signature = crypto.Sign(TestServerBackendPrivateKey, cachedRouteResponse.GetSignData())
	cachedRouteResponseData, err := cachedRouteResponse.MarshalBinary()
	assert.NoError(t, err)

	sessionCacheEntry := transport.SessionCacheEntry{
		SessionID: 9999,
		Sequence:  13,
		Response:  cachedRouteResponseData,
	}
	sessionCacheEntryData, err := sessionCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SESSION-0-9999", string(sessionCacheEntryData))
	assert.NoError(t, err)

	packet := transport.SessionUpdatePacket{
		SessionID:     9999,
		Sequence:      13,
		ServerAddress: net.UDPAddr{IP: net.IPv4zero, Port: 13},

		ClientAddress: net.UDPAddr{
			IP:   net.ParseIP("0.0.0.0"),
			Port: 1234,
		},
		ClientRoutePublicKey: TestBuyersClientPublicKey[:],
	}
	packet.Signature = crypto.Sign(TestBuyersServerPrivateKey, packet.GetSignData())

	data, err := packet.MarshalBinary()
	assert.NoError(t, err)

	var resbuf bytes.Buffer

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, &db, &rp, &iploc, &geoClient, &sessionMetrics, &billing.NoOpBiller{}, TestServerBackendPrivateKey[:], TestRouterPrivateKey[:])
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	assert.Greater(t, resbuf.Len(), 0)

	var actual transport.SessionResponsePacket
	err = actual.UnmarshalBinary(resbuf.Bytes())
	assert.NoError(t, err)

	verified := crypto.Verify(TestServerBackendPublicKey, actual.GetSignData(), actual.Signature)
	assert.True(t, verified)

	assert.Equal(t, packet.SessionID, actual.SessionID)
	assert.Equal(t, packet.Sequence, actual.Sequence)
	assert.Equal(t, int32(routing.RouteTypeNew), actual.RouteType)
	assert.Equal(t, int32(1), actual.NumTokens)
	assert.Equal(t, tokens, actual.Tokens)
	assert.Equal(t, TestBuyersServerPublicKey[:], actual.ServerRoutePublicKey)

	assert.Equal(t, 0.0, sessionMetrics.DirectSessions.Value())
	assert.Equal(t, 0.0, sessionMetrics.NextSessions.Value())
}

func TestVetoedRTT(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	sessionMetrics := metrics.EmptySessionMetrics
	localMetrics := metrics.LocalHandler{}

	decisionMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "decision metric"})
	assert.NoError(t, err)
	directMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "route metric"})
	assert.NoError(t, err)

	sessionMetrics.DecisionMetrics.VetoRTT = decisionMetric
	sessionMetrics.DirectSessions = directMetric

	db := storage.InMemory{}
	db.AddBuyer(context.Background(), routing.Buyer{
		PublicKey:            TestBuyersServerPublicKey,
		RoutingRulesSettings: routing.LocalRoutingRulesSettings,
	})

	iploc := routing.LocateIPFunc(func(ip net.IP) (routing.Location, error) {
		return routing.Location{
			Continent: "NA",
			Country:   "US",
			Region:    "NY",
			City:      "Troy",
			Latitude:  0,
			Longitude: 0,
		}, nil
	})

	geoClient := routing.GeoClient{
		RedisClient: redisClient,
		Namespace:   "GEO_TEST",
	}

	nearbyRelay := routing.Relay{
		ID: 1,
	}
	err = geoClient.Add(nearbyRelay)
	assert.NoError(t, err)

	rp := mockRouteProvider{
		routes: []routing.Route{
			{
				Relays: []routing.Relay{
					{ID: 1, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 123}, PublicKey: TestRelayPublicKey[:]},
					{ID: 2, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.2"), Port: 123}, PublicKey: TestRelayPublicKey[:]},
					{ID: 3, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.3"), Port: 123}, PublicKey: TestRelayPublicKey[:]},
				},
			},
		},
	}

	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
	assert.NoError(t, err)

	serverCacheEntry := transport.ServerCacheEntry{
		Sequence: 13,
		Server: routing.Server{
			Addr:      *addr,
			PublicKey: TestBuyersServerPublicKey[:],
		},
	}
	serverCacheEntryData, err := serverCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SERVER-0-0.0.0.0:13", string(serverCacheEntryData))
	assert.NoError(t, err)

	sessionCacheEntry := transport.SessionCacheEntry{
		SessionID:      9999,
		Sequence:       13,
		TimestampStart: time.Now().Add(-5 * time.Second),
		VetoTimestamp:  time.Now().Add(5 * time.Second),
		RouteDecision: routing.Decision{
			OnNetworkNext: false,
			Reason:        routing.DecisionVetoRTT,
		},
	}
	sessionCacheEntryData, err := sessionCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SESSION-0-9999", string(sessionCacheEntryData))
	assert.NoError(t, err)

	packet := transport.SessionUpdatePacket{
		SessionID:     9999,
		Sequence:      14,
		ServerAddress: net.UDPAddr{IP: net.IPv4zero, Port: 13},

		NumNearRelays:       1,
		NearRelayIDs:        []uint64{1},
		NearRelayMinRTT:     []float32{1},
		NearRelayMaxRTT:     []float32{1},
		NearRelayMeanRTT:    []float32{1},
		NearRelayJitter:     []float32{1},
		NearRelayPacketLoss: []float32{1},

		ClientAddress: net.UDPAddr{
			IP:   net.ParseIP("0.0.0.0"),
			Port: 1234,
		},
		ClientRoutePublicKey: TestBuyersClientPublicKey[:],
	}
	packet.Signature = crypto.Sign(TestBuyersServerPrivateKey, packet.GetSignData())

	data, err := packet.MarshalBinary()
	assert.NoError(t, err)

	var resbuf bytes.Buffer

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, &db, &rp, &iploc, &geoClient, &sessionMetrics, &billing.NoOpBiller{}, TestServerBackendPrivateKey[:], TestRouterPrivateKey[:])
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	validateDirectResponsePacket(t, resbuf, sessionMetrics.DirectSessions, sessionMetrics.DecisionMetrics.VetoRTT)
}

func TestVetoExpiredRTT(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	sessionMetrics := metrics.EmptySessionMetrics
	localMetrics := metrics.LocalHandler{}

	decisionMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "decision metric"})
	assert.NoError(t, err)
	directMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "route metric"})
	assert.NoError(t, err)

	sessionMetrics.DecisionMetrics.InitialSlice = decisionMetric
	sessionMetrics.DirectSessions = directMetric

	db := storage.InMemory{}
	db.AddBuyer(context.Background(), routing.Buyer{
		PublicKey:            TestBuyersServerPublicKey,
		RoutingRulesSettings: routing.LocalRoutingRulesSettings,
	})

	iploc := routing.LocateIPFunc(func(ip net.IP) (routing.Location, error) {
		return routing.Location{
			Continent: "NA",
			Country:   "US",
			Region:    "NY",
			City:      "Troy",
			Latitude:  0,
			Longitude: 0,
		}, nil
	})

	geoClient := routing.GeoClient{
		RedisClient: redisClient,
		Namespace:   "GEO_TEST",
	}

	nearbyRelay := routing.Relay{
		ID: 1,
	}
	err = geoClient.Add(nearbyRelay)
	assert.NoError(t, err)

	rp := mockRouteProvider{
		routes: []routing.Route{
			{
				Relays: []routing.Relay{
					{ID: 1, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 123}, PublicKey: TestRelayPublicKey[:]},
					{ID: 2, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.2"), Port: 123}, PublicKey: TestRelayPublicKey[:]},
					{ID: 3, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.3"), Port: 123}, PublicKey: TestRelayPublicKey[:]},
				},
			},
		},
	}

	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
	assert.NoError(t, err)

	serverCacheEntry := transport.ServerCacheEntry{
		Sequence: 13,
		Server: routing.Server{
			Addr:      *addr,
			PublicKey: TestBuyersServerPublicKey[:],
		},
	}
	serverCacheEntryData, err := serverCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SERVER-0-0.0.0.0:13", string(serverCacheEntryData))
	assert.NoError(t, err)

	sessionCacheEntry := transport.SessionCacheEntry{
		SessionID:      9999,
		Sequence:       13,
		TimestampStart: time.Now().Add(-5 * time.Second),
		VetoTimestamp:  time.Now().Add(-5 * time.Second),
		RouteDecision: routing.Decision{
			OnNetworkNext: false,
			Reason:        routing.DecisionVetoRTT,
		},
	}
	sessionCacheEntryData, err := sessionCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SESSION-0-9999", string(sessionCacheEntryData))
	assert.NoError(t, err)

	packet := transport.SessionUpdatePacket{
		SessionID:     9999,
		Sequence:      14,
		ServerAddress: net.UDPAddr{IP: net.IPv4zero, Port: 13},

		NumNearRelays:       1,
		NearRelayIDs:        []uint64{1},
		NearRelayMinRTT:     []float32{1},
		NearRelayMaxRTT:     []float32{1},
		NearRelayMeanRTT:    []float32{1},
		NearRelayJitter:     []float32{1},
		NearRelayPacketLoss: []float32{1},

		ClientAddress: net.UDPAddr{
			IP:   net.ParseIP("0.0.0.0"),
			Port: 1234,
		},
		ClientRoutePublicKey: TestBuyersClientPublicKey[:],
	}
	packet.Signature = crypto.Sign(TestBuyersServerPrivateKey, packet.GetSignData())

	data, err := packet.MarshalBinary()
	assert.NoError(t, err)

	var resbuf bytes.Buffer

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, &db, &rp, &iploc, &geoClient, &sessionMetrics, &billing.NoOpBiller{}, TestServerBackendPrivateKey[:], TestRouterPrivateKey[:])
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	validateDirectResponsePacket(t, resbuf, sessionMetrics.DirectSessions, sessionMetrics.DecisionMetrics.InitialSlice)
}

func TestVetoedPacketLoss(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	sessionMetrics := metrics.EmptySessionMetrics
	localMetrics := metrics.LocalHandler{}

	decisionMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "decision metric"})
	assert.NoError(t, err)
	directMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "route metric"})
	assert.NoError(t, err)

	sessionMetrics.DecisionMetrics.VetoPacketLoss = decisionMetric
	sessionMetrics.DirectSessions = directMetric

	db := storage.InMemory{}
	db.AddBuyer(context.Background(), routing.Buyer{
		PublicKey:            TestBuyersServerPublicKey,
		RoutingRulesSettings: routing.LocalRoutingRulesSettings,
	})

	iploc := routing.LocateIPFunc(func(ip net.IP) (routing.Location, error) {
		return routing.Location{
			Continent: "NA",
			Country:   "US",
			Region:    "NY",
			City:      "Troy",
			Latitude:  0,
			Longitude: 0,
		}, nil
	})

	geoClient := routing.GeoClient{
		RedisClient: redisClient,
		Namespace:   "GEO_TEST",
	}

	nearbyRelay := routing.Relay{
		ID: 1,
	}
	err = geoClient.Add(nearbyRelay)
	assert.NoError(t, err)

	rp := mockRouteProvider{
		routes: []routing.Route{
			{
				Relays: []routing.Relay{
					{ID: 1, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 123}, PublicKey: TestRelayPublicKey[:]},
					{ID: 2, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.2"), Port: 123}, PublicKey: TestRelayPublicKey[:]},
					{ID: 3, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.3"), Port: 123}, PublicKey: TestRelayPublicKey[:]},
				},
			},
		},
	}

	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
	assert.NoError(t, err)

	serverCacheEntry := transport.ServerCacheEntry{
		Sequence: 13,
		Server: routing.Server{
			Addr:      *addr,
			PublicKey: TestBuyersServerPublicKey[:],
		},
	}
	serverCacheEntryData, err := serverCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SERVER-0-0.0.0.0:13", string(serverCacheEntryData))
	assert.NoError(t, err)

	sessionCacheEntry := transport.SessionCacheEntry{
		SessionID:      9999,
		Sequence:       13,
		TimestampStart: time.Now().Add(-5 * time.Second),
		VetoTimestamp:  time.Now().Add(5 * time.Second),
		RouteDecision: routing.Decision{
			OnNetworkNext: false,
			Reason:        routing.DecisionVetoPacketLoss,
		},
	}
	sessionCacheEntryData, err := sessionCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SESSION-0-9999", string(sessionCacheEntryData))
	assert.NoError(t, err)

	packet := transport.SessionUpdatePacket{
		SessionID:     9999,
		Sequence:      14,
		ServerAddress: net.UDPAddr{IP: net.IPv4zero, Port: 13},

		NumNearRelays:       1,
		NearRelayIDs:        []uint64{1},
		NearRelayMinRTT:     []float32{1},
		NearRelayMaxRTT:     []float32{1},
		NearRelayMeanRTT:    []float32{1},
		NearRelayJitter:     []float32{1},
		NearRelayPacketLoss: []float32{1},

		ClientAddress: net.UDPAddr{
			IP:   net.ParseIP("0.0.0.0"),
			Port: 1234,
		},
		ClientRoutePublicKey: TestBuyersClientPublicKey[:],
	}
	packet.Signature = crypto.Sign(TestBuyersServerPrivateKey, packet.GetSignData())

	data, err := packet.MarshalBinary()
	assert.NoError(t, err)

	var resbuf bytes.Buffer

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, &db, &rp, &iploc, &geoClient, &sessionMetrics, &billing.NoOpBiller{}, TestServerBackendPrivateKey[:], TestRouterPrivateKey[:])
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	validateDirectResponsePacket(t, resbuf, sessionMetrics.DirectSessions, sessionMetrics.DecisionMetrics.VetoPacketLoss)
}

func TestVetoExpiredPacketLoss(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	sessionMetrics := metrics.EmptySessionMetrics
	localMetrics := metrics.LocalHandler{}

	decisionMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "decision metric"})
	assert.NoError(t, err)
	directMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "route metric"})
	assert.NoError(t, err)

	sessionMetrics.DecisionMetrics.InitialSlice = decisionMetric
	sessionMetrics.DirectSessions = directMetric

	db := storage.InMemory{}
	db.AddBuyer(context.Background(), routing.Buyer{
		PublicKey:            TestBuyersServerPublicKey,
		RoutingRulesSettings: routing.LocalRoutingRulesSettings,
	})

	iploc := routing.LocateIPFunc(func(ip net.IP) (routing.Location, error) {
		return routing.Location{
			Continent: "NA",
			Country:   "US",
			Region:    "NY",
			City:      "Troy",
			Latitude:  0,
			Longitude: 0,
		}, nil
	})

	geoClient := routing.GeoClient{
		RedisClient: redisClient,
		Namespace:   "GEO_TEST",
	}

	nearbyRelay := routing.Relay{
		ID: 1,
	}
	err = geoClient.Add(nearbyRelay)
	assert.NoError(t, err)

	rp := mockRouteProvider{
		routes: []routing.Route{
			{
				Relays: []routing.Relay{
					{ID: 1, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 123}, PublicKey: TestRelayPublicKey[:]},
					{ID: 2, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.2"), Port: 123}, PublicKey: TestRelayPublicKey[:]},
					{ID: 3, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.3"), Port: 123}, PublicKey: TestRelayPublicKey[:]},
				},
			},
		},
	}

	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
	assert.NoError(t, err)

	serverCacheEntry := transport.ServerCacheEntry{
		Sequence: 13,
		Server: routing.Server{
			Addr:      *addr,
			PublicKey: TestBuyersServerPublicKey[:],
		},
	}
	serverCacheEntryData, err := serverCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SERVER-0-0.0.0.0:13", string(serverCacheEntryData))
	assert.NoError(t, err)

	sessionCacheEntry := transport.SessionCacheEntry{
		SessionID:      9999,
		Sequence:       13,
		TimestampStart: time.Now().Add(-5 * time.Second),
		VetoTimestamp:  time.Now().Add(-5 * time.Second),
		RouteDecision: routing.Decision{
			OnNetworkNext: false,
			Reason:        routing.DecisionVetoPacketLoss,
		},
	}
	sessionCacheEntryData, err := sessionCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SESSION-0-9999", string(sessionCacheEntryData))
	assert.NoError(t, err)

	packet := transport.SessionUpdatePacket{
		SessionID:     9999,
		Sequence:      14,
		ServerAddress: net.UDPAddr{IP: net.IPv4zero, Port: 13},

		NumNearRelays:       1,
		NearRelayIDs:        []uint64{1},
		NearRelayMinRTT:     []float32{1},
		NearRelayMaxRTT:     []float32{1},
		NearRelayMeanRTT:    []float32{1},
		NearRelayJitter:     []float32{1},
		NearRelayPacketLoss: []float32{1},

		ClientAddress: net.UDPAddr{
			IP:   net.ParseIP("0.0.0.0"),
			Port: 1234,
		},
		ClientRoutePublicKey: TestBuyersClientPublicKey[:],
	}
	packet.Signature = crypto.Sign(TestBuyersServerPrivateKey, packet.GetSignData())

	data, err := packet.MarshalBinary()
	assert.NoError(t, err)

	var resbuf bytes.Buffer

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, &db, &rp, &iploc, &geoClient, &sessionMetrics, &billing.NoOpBiller{}, TestServerBackendPrivateKey[:], TestRouterPrivateKey[:])
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	validateDirectResponsePacket(t, resbuf, sessionMetrics.DirectSessions, sessionMetrics.DecisionMetrics.InitialSlice)
}

func TestVetoYOLONoReturn(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	sessionMetrics := metrics.EmptySessionMetrics
	localMetrics := metrics.LocalHandler{}

	decisionMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "decision metric"})
	assert.NoError(t, err)
	directMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "route metric"})
	assert.NoError(t, err)

	sessionMetrics.DecisionMetrics.VetoRTTYOLO = decisionMetric
	sessionMetrics.DirectSessions = directMetric

	rrs := routing.LocalRoutingRulesSettings
	rrs.EnableYouOnlyLiveOnce = true

	db := storage.InMemory{}
	db.AddBuyer(context.Background(), routing.Buyer{
		PublicKey:            TestBuyersServerPublicKey,
		RoutingRulesSettings: rrs,
	})

	iploc := routing.LocateIPFunc(func(ip net.IP) (routing.Location, error) {
		return routing.Location{
			Continent: "NA",
			Country:   "US",
			Region:    "NY",
			City:      "Troy",
			Latitude:  0,
			Longitude: 0,
		}, nil
	})

	geoClient := routing.GeoClient{
		RedisClient: redisClient,
		Namespace:   "GEO_TEST",
	}

	nearbyRelay := routing.Relay{
		ID: 1,
	}
	err = geoClient.Add(nearbyRelay)
	assert.NoError(t, err)

	rp := mockRouteProvider{
		datacenterRelays: []routing.Relay{{}},
		routes: []routing.Route{
			{
				Relays: []routing.Relay{
					{ID: 1, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 123}, PublicKey: TestRelayPublicKey[:]},
					{ID: 2, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.2"), Port: 123}, PublicKey: TestRelayPublicKey[:]},
					{ID: 3, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.3"), Port: 123}, PublicKey: TestRelayPublicKey[:]},
				},
			},
		},
	}

	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
	assert.NoError(t, err)

	serverCacheEntry := transport.ServerCacheEntry{
		Sequence: 13,
		Server: routing.Server{
			Addr:      *addr,
			PublicKey: TestBuyersServerPublicKey[:],
		},
	}
	serverCacheEntryData, err := serverCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SERVER-0-0.0.0.0:13", string(serverCacheEntryData))
	assert.NoError(t, err)

	sessionCacheEntry := transport.SessionCacheEntry{
		SessionID:      9999,
		Sequence:       13,
		TimestampStart: time.Now().Add(-5 * time.Second),
		VetoTimestamp:  time.Now().Add(-5 * time.Second),
		RouteDecision: routing.Decision{
			OnNetworkNext: false,
			Reason:        routing.DecisionVetoRTT | routing.DecisionVetoYOLO,
		},
	}
	sessionCacheEntryData, err := sessionCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SESSION-0-9999", string(sessionCacheEntryData))
	assert.NoError(t, err)

	packet := transport.SessionUpdatePacket{
		SessionID:     9999,
		Sequence:      14,
		ServerAddress: net.UDPAddr{IP: net.IPv4zero, Port: 13},

		NumNearRelays:       1,
		NearRelayIDs:        []uint64{1},
		NearRelayMinRTT:     []float32{1},
		NearRelayMaxRTT:     []float32{1},
		NearRelayMeanRTT:    []float32{1},
		NearRelayJitter:     []float32{1},
		NearRelayPacketLoss: []float32{1},

		ClientAddress: net.UDPAddr{
			IP:   net.ParseIP("0.0.0.0"),
			Port: 1234,
		},
		ClientRoutePublicKey: TestBuyersClientPublicKey[:],
	}
	packet.Signature = crypto.Sign(TestBuyersServerPrivateKey, packet.GetSignData())

	data, err := packet.MarshalBinary()
	assert.NoError(t, err)

	var resbuf bytes.Buffer

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, &db, &rp, &iploc, &geoClient, &sessionMetrics, &billing.NoOpBiller{}, TestServerBackendPrivateKey[:], TestRouterPrivateKey[:])
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	validateDirectResponsePacket(t, resbuf, sessionMetrics.DirectSessions, sessionMetrics.DecisionMetrics.VetoRTTYOLO)
}

func TestForceDirect(t *testing.T) {
	t.Run("by selection percentage", func(t *testing.T) {
		redisServer, err := miniredis.Run()
		assert.NoError(t, err)
		redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

		sessionMetrics := metrics.EmptySessionMetrics
		localMetrics := metrics.LocalHandler{}

		decisionMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "decision metric"})
		assert.NoError(t, err)
		directMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "route metric"})
		assert.NoError(t, err)

		sessionMetrics.DecisionMetrics.ForceDirect = decisionMetric
		sessionMetrics.DirectSessions = directMetric

		routeRuleSettings := routing.LocalRoutingRulesSettings
		routeRuleSettings.SelectionPercentage = 50
		db := storage.InMemory{}
		db.AddBuyer(context.Background(), routing.Buyer{
			PublicKey:            TestBuyersServerPublicKey,
			RoutingRulesSettings: routeRuleSettings,
		})

		iploc := routing.LocateIPFunc(func(ip net.IP) (routing.Location, error) {
			return routing.Location{
				Continent: "NA",
				Country:   "US",
				Region:    "NY",
				City:      "Troy",
				Latitude:  0,
				Longitude: 0,
			}, nil
		})

		geoClient := routing.GeoClient{
			RedisClient: redisClient,
			Namespace:   "GEO_TEST",
		}

		nearbyRelay := routing.Relay{
			ID: 1,
		}
		err = geoClient.Add(nearbyRelay)
		assert.NoError(t, err)

		rp := mockRouteProvider{
			routes: []routing.Route{
				{
					Relays: []routing.Relay{
						{ID: 1, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 123}, PublicKey: TestRelayPublicKey[:]},
						{ID: 2, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.2"), Port: 123}, PublicKey: TestRelayPublicKey[:]},
						{ID: 3, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.3"), Port: 123}, PublicKey: TestRelayPublicKey[:]},
					},
				},
			},
		}

		addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
		assert.NoError(t, err)

		serverCacheEntry := transport.ServerCacheEntry{
			Sequence: 13,
			Server: routing.Server{
				Addr:      *addr,
				PublicKey: TestBuyersServerPublicKey[:],
			},
		}
		serverCacheEntryData, err := serverCacheEntry.MarshalBinary()
		assert.NoError(t, err)

		err = redisServer.Set("SERVER-0-0.0.0.0:13", string(serverCacheEntryData))
		assert.NoError(t, err)

		sessionCacheEntry := transport.SessionCacheEntry{
			SessionID: 9999,
			Sequence:  13,
		}
		sessionCacheEntryData, err := sessionCacheEntry.MarshalBinary()
		assert.NoError(t, err)

		err = redisServer.Set("SESSION-0-9999", string(sessionCacheEntryData))
		assert.NoError(t, err)

		packet := transport.SessionUpdatePacket{
			SessionID:     9999,
			Sequence:      14,
			ServerAddress: net.UDPAddr{IP: net.IPv4zero, Port: 13},

			NumNearRelays:       1,
			NearRelayIDs:        []uint64{1},
			NearRelayMinRTT:     []float32{1},
			NearRelayMaxRTT:     []float32{1},
			NearRelayMeanRTT:    []float32{1},
			NearRelayJitter:     []float32{1},
			NearRelayPacketLoss: []float32{1},

			ClientAddress: net.UDPAddr{
				IP:   net.ParseIP("0.0.0.0"),
				Port: 1234,
			},
			ClientRoutePublicKey: TestBuyersClientPublicKey[:],
		}
		packet.Signature = crypto.Sign(TestBuyersServerPrivateKey, packet.GetSignData())

		data, err := packet.MarshalBinary()
		assert.NoError(t, err)

		var resbuf bytes.Buffer

		handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, &db, &rp, &iploc, &geoClient, &sessionMetrics, &billing.NoOpBiller{}, TestServerBackendPrivateKey[:], TestRouterPrivateKey[:])
		handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

		validateDirectResponsePacket(t, resbuf, sessionMetrics.DirectSessions, sessionMetrics.DecisionMetrics.ForceDirect)
	})

	t.Run("by mode", func(t *testing.T) {
		redisServer, err := miniredis.Run()
		assert.NoError(t, err)
		redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

		sessionMetrics := metrics.EmptySessionMetrics
		localMetrics := metrics.LocalHandler{}

		decisionMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "decision metric"})
		assert.NoError(t, err)
		directMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "route metric"})
		assert.NoError(t, err)

		sessionMetrics.DecisionMetrics.ForceDirect = decisionMetric
		sessionMetrics.DirectSessions = directMetric

		routeRuleSettings := routing.LocalRoutingRulesSettings
		routeRuleSettings.Mode = routing.ModeForceDirect
		db := storage.InMemory{}
		db.AddBuyer(context.Background(), routing.Buyer{
			PublicKey:            TestBuyersServerPublicKey,
			RoutingRulesSettings: routeRuleSettings,
		})

		iploc := routing.LocateIPFunc(func(ip net.IP) (routing.Location, error) {
			return routing.Location{
				Continent: "NA",
				Country:   "US",
				Region:    "NY",
				City:      "Troy",
				Latitude:  0,
				Longitude: 0,
			}, nil
		})

		geoClient := routing.GeoClient{
			RedisClient: redisClient,
			Namespace:   "GEO_TEST",
		}

		nearbyRelay := routing.Relay{
			ID: 1,
		}
		err = geoClient.Add(nearbyRelay)
		assert.NoError(t, err)

		rp := mockRouteProvider{
			routes: []routing.Route{
				{
					Relays: []routing.Relay{
						{ID: 1, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 123}, PublicKey: TestRelayPublicKey[:]},
						{ID: 2, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.2"), Port: 123}, PublicKey: TestRelayPublicKey[:]},
						{ID: 3, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.3"), Port: 123}, PublicKey: TestRelayPublicKey[:]},
					},
				},
			},
		}

		addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
		assert.NoError(t, err)

		serverCacheEntry := transport.ServerCacheEntry{
			Sequence: 13,
			Server: routing.Server{
				Addr:      *addr,
				PublicKey: TestBuyersServerPublicKey[:],
			},
		}
		serverCacheEntryData, err := serverCacheEntry.MarshalBinary()
		assert.NoError(t, err)

		err = redisServer.Set("SERVER-0-0.0.0.0:13", string(serverCacheEntryData))
		assert.NoError(t, err)

		sessionCacheEntry := transport.SessionCacheEntry{
			SessionID: 9999,
			Sequence:  13,
		}
		sessionCacheEntryData, err := sessionCacheEntry.MarshalBinary()
		assert.NoError(t, err)

		err = redisServer.Set("SESSION-0-9999", string(sessionCacheEntryData))
		assert.NoError(t, err)

		packet := transport.SessionUpdatePacket{
			SessionID:     9999,
			Sequence:      14,
			ServerAddress: net.UDPAddr{IP: net.IPv4zero, Port: 13},

			NumNearRelays:       1,
			NearRelayIDs:        []uint64{1},
			NearRelayMinRTT:     []float32{1},
			NearRelayMaxRTT:     []float32{1},
			NearRelayMeanRTT:    []float32{1},
			NearRelayJitter:     []float32{1},
			NearRelayPacketLoss: []float32{1},

			ClientAddress: net.UDPAddr{
				IP:   net.ParseIP("0.0.0.0"),
				Port: 1234,
			},
			ClientRoutePublicKey: TestBuyersClientPublicKey[:],
		}
		packet.Signature = crypto.Sign(TestBuyersServerPrivateKey, packet.GetSignData())

		data, err := packet.MarshalBinary()
		assert.NoError(t, err)

		var resbuf bytes.Buffer

		handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, &db, &rp, &iploc, &geoClient, &sessionMetrics, &billing.NoOpBiller{}, TestServerBackendPrivateKey[:], TestRouterPrivateKey[:])
		handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

		validateDirectResponsePacket(t, resbuf, sessionMetrics.DirectSessions, sessionMetrics.DecisionMetrics.ForceDirect)
	})
}

func TestForceNext(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	sessionMetrics := metrics.EmptySessionMetrics
	localMetrics := metrics.LocalHandler{}

	decisionMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "decision metric"})
	assert.NoError(t, err)
	nextMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "route metric"})
	assert.NoError(t, err)

	sessionMetrics.DecisionMetrics.ForceNext = decisionMetric
	sessionMetrics.NextSessions = nextMetric

	routeRuleSettings := routing.LocalRoutingRulesSettings
	routeRuleSettings.Mode = routing.ModeForceNext
	routeRuleSettings.SelectionPercentage = 100
	db := storage.InMemory{}
	db.AddBuyer(context.Background(), routing.Buyer{
		PublicKey:            TestBuyersServerPublicKey,
		RoutingRulesSettings: routeRuleSettings,
	})

	iploc := routing.LocateIPFunc(func(ip net.IP) (routing.Location, error) {
		return routing.Location{
			Continent: "NA",
			Country:   "US",
			Region:    "NY",
			City:      "Troy",
			Latitude:  0,
			Longitude: 0,
		}, nil
	})

	geoClient := routing.GeoClient{
		RedisClient: redisClient,
		Namespace:   "GEO_TEST",
	}

	nearbyRelay := routing.Relay{
		ID: 1,
	}
	err = geoClient.Add(nearbyRelay)
	assert.NoError(t, err)

	rp := mockRouteProvider{
		routes: []routing.Route{
			{
				Relays: []routing.Relay{
					{ID: 1, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 123}, PublicKey: TestRelayPublicKey[:]},
					{ID: 2, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.2"), Port: 123}, PublicKey: TestRelayPublicKey[:]},
					{ID: 3, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.3"), Port: 123}, PublicKey: TestRelayPublicKey[:]},
				},
			},
		},
	}

	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
	assert.NoError(t, err)

	serverCacheEntry := transport.ServerCacheEntry{
		Sequence: 13,
		Server: routing.Server{
			Addr:      *addr,
			PublicKey: TestBuyersServerPublicKey[:],
		},
	}
	serverCacheEntryData, err := serverCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SERVER-0-0.0.0.0:13", string(serverCacheEntryData))
	assert.NoError(t, err)

	sessionCacheEntry := transport.SessionCacheEntry{
		SessionID: 9999,
		Sequence:  13,
	}
	sessionCacheEntryData, err := sessionCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SESSION-0-9999", string(sessionCacheEntryData))
	assert.NoError(t, err)

	packet := transport.SessionUpdatePacket{
		SessionID:     9999,
		Sequence:      14,
		ServerAddress: net.UDPAddr{IP: net.IPv4zero, Port: 13},

		NumNearRelays:       1,
		NearRelayIDs:        []uint64{1},
		NearRelayMinRTT:     []float32{1},
		NearRelayMaxRTT:     []float32{1},
		NearRelayMeanRTT:    []float32{1},
		NearRelayJitter:     []float32{1},
		NearRelayPacketLoss: []float32{1},

		ClientAddress: net.UDPAddr{
			IP:   net.ParseIP("0.0.0.0"),
			Port: 1234,
		},
		ClientRoutePublicKey: TestBuyersClientPublicKey[:],
	}
	packet.Signature = crypto.Sign(TestBuyersServerPrivateKey, packet.GetSignData())

	data, err := packet.MarshalBinary()
	assert.NoError(t, err)

	var resbuf bytes.Buffer

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, &db, &rp, &iploc, &geoClient, &sessionMetrics, &billing.NoOpBiller{}, TestServerBackendPrivateKey[:], TestRouterPrivateKey[:])
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	validateNextResponsePacket(t, resbuf, packet.SessionID, packet.Sequence, 5, routing.RouteTypeNew, sessionMetrics.NextSessions, sessionMetrics.DecisionMetrics.ForceNext)
}

func TestABTest(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	sessionMetrics := metrics.EmptySessionMetrics
	localMetrics := metrics.LocalHandler{}

	decisionMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "decision metric"})
	assert.NoError(t, err)
	directMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "route metric"})
	assert.NoError(t, err)

	sessionMetrics.DecisionMetrics.ABTestDirect = decisionMetric
	sessionMetrics.DirectSessions = directMetric

	routeRuleSettings := routing.LocalRoutingRulesSettings
	routeRuleSettings.EnableABTest = true
	db := storage.InMemory{}
	db.AddBuyer(context.Background(), routing.Buyer{
		PublicKey:            TestBuyersServerPublicKey,
		RoutingRulesSettings: routeRuleSettings,
	})

	iploc := routing.LocateIPFunc(func(ip net.IP) (routing.Location, error) {
		return routing.Location{
			Continent: "NA",
			Country:   "US",
			Region:    "NY",
			City:      "Troy",
			Latitude:  0,
			Longitude: 0,
		}, nil
	})

	geoClient := routing.GeoClient{
		RedisClient: redisClient,
		Namespace:   "GEO_TEST",
	}

	nearbyRelay := routing.Relay{
		ID: 1,
	}
	err = geoClient.Add(nearbyRelay)
	assert.NoError(t, err)

	rp := mockRouteProvider{
		routes: []routing.Route{
			{
				Relays: []routing.Relay{
					{ID: 1, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 123}, PublicKey: TestRelayPublicKey[:]},
					{ID: 2, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.2"), Port: 123}, PublicKey: TestRelayPublicKey[:]},
					{ID: 3, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.3"), Port: 123}, PublicKey: TestRelayPublicKey[:]},
				},
			},
		},
	}

	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
	assert.NoError(t, err)

	serverCacheEntry := transport.ServerCacheEntry{
		Sequence: 13,
		Server: routing.Server{
			Addr:      *addr,
			PublicKey: TestBuyersServerPublicKey[:],
		},
	}
	serverCacheEntryData, err := serverCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SERVER-0-0.0.0.0:13", string(serverCacheEntryData))
	assert.NoError(t, err)

	sessionCacheEntry := transport.SessionCacheEntry{
		SessionID: 9999,
		Sequence:  13,
	}
	sessionCacheEntryData, err := sessionCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SESSION-0-9999", string(sessionCacheEntryData))
	assert.NoError(t, err)

	packet := transport.SessionUpdatePacket{
		SessionID:     9999,
		Sequence:      14,
		ServerAddress: net.UDPAddr{IP: net.IPv4zero, Port: 13},

		NumNearRelays:       1,
		NearRelayIDs:        []uint64{1},
		NearRelayMinRTT:     []float32{1},
		NearRelayMaxRTT:     []float32{1},
		NearRelayMeanRTT:    []float32{1},
		NearRelayJitter:     []float32{1},
		NearRelayPacketLoss: []float32{1},

		ClientAddress: net.UDPAddr{
			IP:   net.ParseIP("0.0.0.0"),
			Port: 1234,
		},
		ClientRoutePublicKey: TestBuyersClientPublicKey[:],
	}
	packet.Signature = crypto.Sign(TestBuyersServerPrivateKey, packet.GetSignData())

	data, err := packet.MarshalBinary()
	assert.NoError(t, err)

	var resbuf bytes.Buffer

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, &db, &rp, &iploc, &geoClient, &sessionMetrics, &billing.NoOpBiller{}, TestServerBackendPrivateKey[:], TestRouterPrivateKey[:])
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	validateDirectResponsePacket(t, resbuf, sessionMetrics.DirectSessions, sessionMetrics.DecisionMetrics.ABTestDirect)
}

func TestVetoCommit(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	sessionMetrics := metrics.EmptySessionMetrics
	localMetrics := metrics.LocalHandler{}

	decisionMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "decision metric"})
	assert.NoError(t, err)
	directMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "route metric"})
	assert.NoError(t, err)

	sessionMetrics.DecisionMetrics.VetoCommit = decisionMetric
	sessionMetrics.DirectSessions = directMetric

	rrs := routing.LocalRoutingRulesSettings
	rrs.EnableTryBeforeYouBuy = true

	db := storage.InMemory{}
	err = db.AddBuyer(context.Background(), routing.Buyer{
		PublicKey:            TestBuyersServerPublicKey,
		RoutingRulesSettings: rrs,
	})
	assert.NoError(t, err)

	iploc := routing.LocateIPFunc(func(ip net.IP) (routing.Location, error) {
		return routing.Location{
			Continent: "NA",
			Country:   "US",
			Region:    "NY",
			City:      "Troy",
			Latitude:  0,
			Longitude: 0,
		}, nil
	})

	geoClient := routing.GeoClient{
		RedisClient: redisClient,
		Namespace:   "GEO_TEST",
	}

	nearbyRelay := routing.Relay{
		ID: 1,
	}
	err = geoClient.Add(nearbyRelay)
	assert.NoError(t, err)

	rp := mockRouteProvider{
		routes: []routing.Route{
			{
				Relays: []routing.Relay{
					{ID: 1, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 123}, PublicKey: TestRelayPublicKey[:]},
					{ID: 2, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.2"), Port: 123}, PublicKey: TestRelayPublicKey[:]},
					{ID: 3, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.3"), Port: 123}, PublicKey: TestRelayPublicKey[:]},
				},
			},
		},
	}

	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
	assert.NoError(t, err)

	serverCacheEntry := transport.ServerCacheEntry{
		Sequence: 13,
		Server: routing.Server{
			Addr:      *addr,
			PublicKey: TestBuyersServerPublicKey[:],
		},
	}
	serverCacheEntryData, err := serverCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SERVER-0-0.0.0.0:13", string(serverCacheEntryData))
	assert.NoError(t, err)

	sessionCacheEntry := transport.SessionCacheEntry{
		SessionID: 9999,
		Sequence:  13,
		RouteDecision: routing.Decision{
			OnNetworkNext: true,
			Reason:        routing.DecisionNoReason,
		},
		CommitPending:              true,
		CommitObservedSliceCounter: uint8(rrs.TryBeforeYouBuyMaxSlices),
	}
	sessionCacheEntryData, err := sessionCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SESSION-0-9999", string(sessionCacheEntryData))
	assert.NoError(t, err)

	packet := transport.SessionUpdatePacket{
		SessionID:     9999,
		Sequence:      14,
		ServerAddress: net.UDPAddr{IP: net.IPv4zero, Port: 13},

		OnNetworkNext: true,
		DirectMinRTT:  30,
		NextMinRTT:    31,

		NumNearRelays:       1,
		NearRelayIDs:        []uint64{1},
		NearRelayMinRTT:     []float32{1},
		NearRelayMaxRTT:     []float32{1},
		NearRelayMeanRTT:    []float32{1},
		NearRelayJitter:     []float32{1},
		NearRelayPacketLoss: []float32{1},

		ClientAddress: net.UDPAddr{
			IP:   net.ParseIP("0.0.0.0"),
			Port: 1234,
		},
		ClientRoutePublicKey: TestBuyersClientPublicKey[:],
	}
	packet.Signature = crypto.Sign(TestBuyersServerPrivateKey, packet.GetSignData())

	data, err := packet.MarshalBinary()
	assert.NoError(t, err)

	var resbuf bytes.Buffer

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, &db, &rp, &iploc, &geoClient, &sessionMetrics, &billing.NoOpBiller{}, TestServerBackendPrivateKey[:], TestRouterPrivateKey[:])
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	actual := validateDirectResponsePacket(t, resbuf, sessionMetrics.DirectSessions, sessionMetrics.DecisionMetrics.VetoCommit)
	assert.Equal(t, false, actual.Committed)
}

func TestCommitted(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	sessionMetrics := metrics.EmptySessionMetrics
	localMetrics := metrics.LocalHandler{}

	decisionMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "decision metric"})
	assert.NoError(t, err)
	nextMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "route metric"})
	assert.NoError(t, err)

	sessionMetrics.DecisionMetrics.NoReason = decisionMetric
	sessionMetrics.NextSessions = nextMetric

	rrs := routing.LocalRoutingRulesSettings
	rrs.EnableTryBeforeYouBuy = true

	db := storage.InMemory{}
	err = db.AddBuyer(context.Background(), routing.Buyer{
		PublicKey:            TestBuyersServerPublicKey,
		RoutingRulesSettings: rrs,
	})
	assert.NoError(t, err)

	iploc := routing.LocateIPFunc(func(ip net.IP) (routing.Location, error) {
		return routing.Location{
			Continent: "NA",
			Country:   "US",
			Region:    "NY",
			City:      "Troy",
			Latitude:  0,
			Longitude: 0,
		}, nil
	})

	geoClient := routing.GeoClient{
		RedisClient: redisClient,
		Namespace:   "GEO_TEST",
	}

	nearbyRelay := routing.Relay{
		ID: 1,
	}
	err = geoClient.Add(nearbyRelay)
	assert.NoError(t, err)

	rp := mockRouteProvider{
		routes: []routing.Route{
			{
				Relays: []routing.Relay{
					{ID: 1, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 123}, PublicKey: TestRelayPublicKey[:]},
					{ID: 2, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.2"), Port: 123}, PublicKey: TestRelayPublicKey[:]},
					{ID: 3, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.3"), Port: 123}, PublicKey: TestRelayPublicKey[:]},
				},
			},
		},
	}

	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
	assert.NoError(t, err)

	serverCacheEntry := transport.ServerCacheEntry{
		Sequence: 13,
		Server: routing.Server{
			Addr:      *addr,
			PublicKey: TestBuyersServerPublicKey[:],
		},
	}
	serverCacheEntryData, err := serverCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SERVER-0-0.0.0.0:13", string(serverCacheEntryData))
	assert.NoError(t, err)

	sessionCacheEntry := transport.SessionCacheEntry{
		SessionID: 9999,
		Sequence:  13,
		RouteDecision: routing.Decision{
			OnNetworkNext: true,
			Reason:        routing.DecisionNoReason,
		},
		CommitPending:              true,
		CommitObservedSliceCounter: uint8(rrs.TryBeforeYouBuyMaxSlices),
	}
	sessionCacheEntryData, err := sessionCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SESSION-0-9999", string(sessionCacheEntryData))
	assert.NoError(t, err)

	packet := transport.SessionUpdatePacket{
		SessionID:     9999,
		Sequence:      14,
		ServerAddress: net.UDPAddr{IP: net.IPv4zero, Port: 13},

		OnNetworkNext: true,
		DirectMinRTT:  30,
		NextMinRTT:    15,

		NumNearRelays:       1,
		NearRelayIDs:        []uint64{1},
		NearRelayMinRTT:     []float32{1},
		NearRelayMaxRTT:     []float32{1},
		NearRelayMeanRTT:    []float32{1},
		NearRelayJitter:     []float32{1},
		NearRelayPacketLoss: []float32{1},

		ClientAddress: net.UDPAddr{
			IP:   net.ParseIP("0.0.0.0"),
			Port: 1234,
		},
		ClientRoutePublicKey: TestBuyersClientPublicKey[:],
	}
	packet.Signature = crypto.Sign(TestBuyersServerPrivateKey, packet.GetSignData())

	data, err := packet.MarshalBinary()
	assert.NoError(t, err)

	var resbuf bytes.Buffer

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, &db, &rp, &iploc, &geoClient, &sessionMetrics, &billing.NoOpBiller{}, TestServerBackendPrivateKey[:], TestRouterPrivateKey[:])
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	actual := validateNextResponsePacket(t, resbuf, packet.SessionID, packet.Sequence, 5, routing.RouteTypeNew, sessionMetrics.NextSessions, sessionMetrics.DecisionMetrics.NoReason)
	assert.Equal(t, true, actual.Committed)
}
