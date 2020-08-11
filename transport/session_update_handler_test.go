package transport_test

// todo: disabled
/*
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

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, 10*time.Second, nil, nil, nil, nil, &sessionMetrics, nil, nil, nil)
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	assert.Equal(t, 0, resbuf.Len())

	assert.Equal(t, 1.0, sessionMetrics.ErrorMetrics.PipelineExecFailure.Value())
}

func TestServerDataMissing(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
	assert.NoError(t, err)

	sessionMetrics := metrics.EmptySessionMetrics
	localMetrics := metrics.LocalHandler{}

	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
	assert.NoError(t, err)

	sessionMetrics.ErrorMetrics.ServerDataMissing = metric

	packet := transport.SessionUpdatePacket{
		Sequence:             13,
		ServerAddress:        net.UDPAddr{IP: net.IPv4zero, Port: 13},
		ClientRoutePublicKey: make([]byte, crypto.KeySize),
	}

	packet.Signature = crypto.Sign(TestBuyersServerPrivateKey, packet.GetSignData())

	data, err := packet.MarshalBinary()
	assert.NoError(t, err)

	var resbuf bytes.Buffer

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, 10*time.Second, nil, nil, nil, nil, &sessionMetrics, nil, nil, nil)
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	assert.Equal(t, 1.0, sessionMetrics.ErrorMetrics.ServerDataMissing.Value())
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

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, 10*time.Second, nil, nil, nil, nil, &sessionMetrics, nil, nil, nil)
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

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, 10*time.Second, nil, nil, nil, nil, &sessionMetrics, nil, TestServerBackendPrivateKey, nil)
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	validateDirectResponsePacket(t, resbuf, sessionMetrics.DirectSessions, sessionMetrics.ErrorMetrics.UnmarshalSessionDataFailure)
}

func TestFailToUnmarshalVetoData(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	sessionMetrics := metrics.EmptySessionMetrics
	localMetrics := metrics.LocalHandler{}

	errMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "err metric"})
	assert.NoError(t, err)
	directMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "direct metric"})
	assert.NoError(t, err)

	sessionMetrics.ErrorMetrics.UnmarshalVetoDataFailure = errMetric
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

	sessionCacheEntry := transport.SessionCacheEntry{
		SessionID: 9999,
		Sequence:  13,
	}

	sessionCacheEntryData, err := sessionCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SESSION-0-9999", string(sessionCacheEntryData))
	assert.NoError(t, err)

	err = redisServer.Set("VETO-0-9999", "bad veto data")
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

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, 10*time.Second, nil, nil, nil, nil, &sessionMetrics, nil, TestServerBackendPrivateKey, nil)
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	validateDirectResponsePacket(t, resbuf, sessionMetrics.DirectSessions, sessionMetrics.ErrorMetrics.UnmarshalVetoDataFailure)
}

func TestFailToUnmarshalPacket(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	sessionMetrics := metrics.EmptySessionMetrics
	localMetrics := metrics.LocalHandler{}

	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "err metric"})
	assert.NoError(t, err)

	sessionMetrics.ErrorMetrics.ReadPacketFailure = metric

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
		CustomerID:           0,
		ServerAddress:        *addr,
		ClientRoutePublicKey: make([]byte, crypto.KeySize),
	}

	packet.Signature = crypto.Sign(TestBuyersServerPrivateKey, packet.GetSignData())

	data, err := packet.MarshalBinary()
	assert.NoError(t, err)

	// Malform the packet
	data = data[:200]

	var resbuf bytes.Buffer

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, 10*time.Second, &db, nil, nil, nil, &sessionMetrics, nil, TestServerBackendPrivateKey, nil)
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	assert.Equal(t, 1.0, sessionMetrics.ErrorMetrics.ReadPacketFailure.Value())
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

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, 10*time.Second, &db, nil, nil, nil, &sessionMetrics, nil, TestServerBackendPrivateKey, nil)
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

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, 10*time.Second, &db, nil, nil, nil, &sessionMetrics, nil, TestServerBackendPrivateKey, nil)
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

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, 10*time.Second, &db, nil, nil, nil, &sessionMetrics, nil, TestServerBackendPrivateKey, nil)
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

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, 10*time.Second, &db, nil, nil, nil, &sessionMetrics, nil, TestServerBackendPrivateKey, nil)
	handler(&badWriter, &transport.UDPPacket{SourceAddr: addr, Data: data})

	var actual transport.SessionResponsePacket
	err = actual.UnmarshalBinary(badWriter.Bytes())
	assert.Error(t, err)

	assert.Equal(t, 1.0, sessionMetrics.ErrorMetrics.WriteCachedResponseFailure.Value())
}

func TestFallbackToDirect(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	sessionMetrics := metrics.EmptySessionMetrics
	localMetrics := metrics.LocalHandler{}

	errMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "err metric"})
	assert.NoError(t, err)
	directMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "direct metric"})
	assert.NoError(t, err)

	sessionMetrics.DecisionMetrics.FallbackToDirect = errMetric
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
		Sequence:      14,
		ServerAddress: net.UDPAddr{IP: net.IPv4zero, Port: 13},

		ClientRoutePublicKey: make([]byte, crypto.KeySize),

		FallbackToDirect: true,

		Signature: make([]byte, ed25519.SignatureSize),
	}
	packet.Signature = crypto.Sign(TestBuyersServerPrivateKey, packet.GetSignData())

	data, err := packet.MarshalBinary()
	assert.NoError(t, err)

	var resbuf bytes.Buffer

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, 10*time.Second, &db, nil, nil, nil, &sessionMetrics, &billing.NoOpBiller{}, TestServerBackendPrivateKey, nil)
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	validateDirectResponsePacket(t, resbuf, sessionMetrics.DirectSessions, sessionMetrics.DecisionMetrics.FallbackToDirect)
}

func TestEarlyFallbackToDirect(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	sessionMetrics := metrics.EmptySessionMetrics
	localMetrics := metrics.LocalHandler{}

	errMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "err metric"})
	assert.NoError(t, err)
	directMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "direct metric"})
	assert.NoError(t, err)

	sessionMetrics.ErrorMetrics.EarlyFallbackToDirect = errMetric
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
		SessionID:      9999,
		Sequence:       13,
		TimestampStart: time.Now().Add(-10 * time.Second),
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

		FallbackToDirect: true,

		Signature: make([]byte, ed25519.SignatureSize),
	}
	packet.Signature = crypto.Sign(TestBuyersServerPrivateKey, packet.GetSignData())

	data, err := packet.MarshalBinary()
	assert.NoError(t, err)

	var resbuf bytes.Buffer

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, 10*time.Second, &db, nil, nil, nil, &sessionMetrics, &billing.NoOpBiller{}, TestServerBackendPrivateKey, nil)
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	validateDirectResponsePacket(t, resbuf, sessionMetrics.DirectSessions, sessionMetrics.ErrorMetrics.EarlyFallbackToDirect)
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

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, 10*time.Second, &db, nil, &iploc, nil, &sessionMetrics, &billing.NoOpBiller{}, TestServerBackendPrivateKey, nil)
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

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, 10*time.Second, &db, &rp, &iploc, &geoClient, &sessionMetrics, &billing.NoOpBiller{}, TestServerBackendPrivateKey, nil)
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	validateDirectResponsePacket(t, resbuf, sessionMetrics.DirectSessions, sessionMetrics.ErrorMetrics.NearRelaysLocateFailure)
}

func TestDatacenterDisabled(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	sessionMetrics := metrics.EmptySessionMetrics
	localMetrics := metrics.LocalHandler{}

	errMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "err metric"})
	assert.NoError(t, err)
	directMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "direct metric"})
	assert.NoError(t, err)

	sessionMetrics.ErrorMetrics.DatacenterDisabled = errMetric
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

	err = geoClient.Add(1, 0, 0)
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

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, 10*time.Second, &db, &rp, &iploc, &geoClient, &sessionMetrics, &billing.NoOpBiller{}, TestServerBackendPrivateKey, nil)
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	validateDirectResponsePacket(t, resbuf, sessionMetrics.DirectSessions, sessionMetrics.ErrorMetrics.DatacenterDisabled)
}

func TestNoRelaysInDatacenter(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	sessionMetrics := metrics.EmptySessionMetrics
	localMetrics := metrics.LocalHandler{}

	errMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "err metric"})
	assert.NoError(t, err)
	directMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "direct metric"})
	assert.NoError(t, err)

	sessionMetrics.ErrorMetrics.NoRelaysInDatacenter = errMetric
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

	err = geoClient.Add(1, 0, 0)
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
		Datacenter: routing.Datacenter{
			Enabled: true,
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

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, 10*time.Second, &db, &rp, &iploc, &geoClient, &sessionMetrics, &billing.NoOpBiller{}, TestServerBackendPrivateKey, nil)
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	validateDirectResponsePacket(t, resbuf, sessionMetrics.DirectSessions, sessionMetrics.ErrorMetrics.NoRelaysInDatacenter)
}

func TestNoRoutesInRouteMatrix(t *testing.T) {
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
		RouteShader: routing.LocalRouteShader,
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

	err = geoClient.Add(1, 0, 0)
	assert.NoError(t, err)

	rp := mockRouteProvider{
		datacenterRelays: []routing.Relay{{}},
	}

	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
	assert.NoError(t, err)

	serverCacheEntry := transport.ServerCacheEntry{
		Sequence: 13,
		Server: routing.Server{
			Addr:      *addr,
			PublicKey: TestServerBackendPublicKey,
		},
		Datacenter: routing.Datacenter{
			Enabled: true,
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

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, 10*time.Second, &db, &rp, &iploc, &geoClient, &sessionMetrics, &billing.NoOpBiller{}, TestServerBackendPrivateKey, nil)
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	validateDirectResponsePacket(t, resbuf, sessionMetrics.DirectSessions, sessionMetrics.ErrorMetrics.RouteFailure)
}

func TestNoRoutesInRouteMatrixVeto(t *testing.T) {
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
		RouteShader: routing.LocalRouteShader,
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

	err = geoClient.Add(1, 0, 0)
	assert.NoError(t, err)

	rp := mockRouteProvider{
		datacenterRelays: []routing.Relay{{}},
	}

	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
	assert.NoError(t, err)

	serverCacheEntry := transport.ServerCacheEntry{
		Sequence: 13,
		Server: routing.Server{
			Addr:      *addr,
			PublicKey: TestServerBackendPublicKey,
		},
		Datacenter: routing.Datacenter{
			Enabled: true,
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

		OnNetworkNext: true,

		ClientRoutePublicKey: make([]byte, crypto.KeySize),

		Signature: make([]byte, ed25519.SignatureSize),
	}
	packet.Signature = crypto.Sign(TestBuyersServerPrivateKey, packet.GetSignData())

	data, err := packet.MarshalBinary()
	assert.NoError(t, err)

	var resbuf bytes.Buffer

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, 10*time.Second, &db, &rp, &iploc, &geoClient, &sessionMetrics, &billing.NoOpBiller{}, TestServerBackendPrivateKey, nil)
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	validateDirectResponsePacket(t, resbuf, sessionMetrics.DirectSessions, sessionMetrics.ErrorMetrics.RouteFailure)

	vetoDataString, err := redisServer.Get("VETO-0-9999")
	assert.NoError(t, err)

	vetoData := []byte(vetoDataString)

	var vetoCacheEntry transport.VetoCacheEntry
	err = vetoCacheEntry.UnmarshalBinary(vetoData)
	assert.NoError(t, err)

	assert.Equal(t, routing.DecisionVetoNoRoute, vetoCacheEntry.Reason)
}

func TestNoRoutesSelected(t *testing.T) {
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
		RouteShader: routing.LocalRouteShader,
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

	err = geoClient.Add(1, 0, 0)
	assert.NoError(t, err)

	relayAddr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
	assert.NoError(t, err)

	dsRelayAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
	assert.NoError(t, err)

	relay := routing.Relay{
		ID:          crypto.HashID(relayAddr.String()),
		Name:        "client.relay",
		Addr:        *relayAddr,
		MaxSessions: 3000,
	}

	dsRelay := routing.Relay{
		ID:          crypto.HashID(dsRelayAddr.String()),
		Name:        "datacenter.relay",
		Addr:        *dsRelayAddr,
		MaxSessions: 0,
	}

	rp := mockRouteProvider{
		relay:            relay,
		datacenterRelays: []routing.Relay{dsRelay},
		routes: []routing.Route{
			{
				Relays: []routing.Relay{
					relay,
					dsRelay,
				},
				Stats: routing.Stats{
					RTT:        30,
					Jitter:     0,
					PacketLoss: 0,
				},
			},
		},
	}

	serverCacheEntry := transport.ServerCacheEntry{
		Sequence: 13,
		Server: routing.Server{
			Addr:      *relayAddr,
			PublicKey: TestServerBackendPublicKey,
		},
		Datacenter: routing.Datacenter{
			Enabled: true,
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

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, 10*time.Second, &db, &rp, &iploc, &geoClient, &sessionMetrics, &billing.NoOpBiller{}, TestServerBackendPrivateKey, nil)
	handler(&resbuf, &transport.UDPPacket{SourceAddr: relayAddr, Data: data})

	validateDirectResponsePacket(t, resbuf, sessionMetrics.DirectSessions, sessionMetrics.ErrorMetrics.RouteFailure)
}

func TestNoRoutesSelectedVeto(t *testing.T) {
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
		RouteShader: routing.LocalRouteShader,
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

	err = geoClient.Add(1, 0, 0)
	assert.NoError(t, err)

	relayAddr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
	assert.NoError(t, err)

	dsRelayAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
	assert.NoError(t, err)

	relay := routing.Relay{
		ID:          crypto.HashID(relayAddr.String()),
		Name:        "client.relay",
		Addr:        *relayAddr,
		MaxSessions: 3000,
	}

	dsRelay := routing.Relay{
		ID:          crypto.HashID(dsRelayAddr.String()),
		Name:        "datacenter.relay",
		Addr:        *dsRelayAddr,
		MaxSessions: 0,
	}

	rp := mockRouteProvider{
		relay:            relay,
		datacenterRelays: []routing.Relay{dsRelay},
		routes: []routing.Route{
			{
				Relays: []routing.Relay{
					relay,
					dsRelay,
				},
				Stats: routing.Stats{
					RTT:        30,
					Jitter:     0,
					PacketLoss: 0,
				},
			},
		},
	}

	serverCacheEntry := transport.ServerCacheEntry{
		Sequence: 13,
		Server: routing.Server{
			Addr:      *relayAddr,
			PublicKey: TestServerBackendPublicKey,
		},
		Datacenter: routing.Datacenter{
			Enabled: true,
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

		OnNetworkNext: true,

		ClientRoutePublicKey: make([]byte, crypto.KeySize),

		Signature: make([]byte, ed25519.SignatureSize),
	}
	packet.Signature = crypto.Sign(TestBuyersServerPrivateKey, packet.GetSignData())

	data, err := packet.MarshalBinary()
	assert.NoError(t, err)

	var resbuf bytes.Buffer

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, 10*time.Second, &db, &rp, &iploc, &geoClient, &sessionMetrics, &billing.NoOpBiller{}, TestServerBackendPrivateKey, nil)
	handler(&resbuf, &transport.UDPPacket{SourceAddr: relayAddr, Data: data})

	validateDirectResponsePacket(t, resbuf, sessionMetrics.DirectSessions, sessionMetrics.ErrorMetrics.RouteFailure)

	vetoDataString, err := redisServer.Get("VETO-0-9999")
	assert.NoError(t, err)

	vetoData := []byte(vetoDataString)

	var vetoCacheEntry transport.VetoCacheEntry
	err = vetoCacheEntry.UnmarshalBinary(vetoData)
	assert.NoError(t, err)

	assert.Equal(t, routing.DecisionVetoNoRoute, vetoCacheEntry.Reason)
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
	buyer.RouteShader.SelectionPercentage = 100

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

	err = geoClient.Add(1, 0, 0)
	assert.NoError(t, err)

	rp := mockRouteProvider{
		datacenterRelays: []routing.Relay{{}},
		routes: []routing.Route{
			{
				Relays: []routing.Relay{
					{ID: 1, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 123}, PublicKey: TestRelayPublicKey[:], MaxSessions: 3000},
					{ID: 2, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.2"), Port: 123}, MaxSessions: 3000},
					{ID: 3, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.3"), Port: 123}, PublicKey: TestRelayPublicKey[:], MaxSessions: 3000},
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
		Datacenter: routing.Datacenter{
			Enabled: true,
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

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, 10*time.Second, &db, &rp, &iploc, &geoClient, &sessionMetrics, &billing.NoOpBiller{}, TestServerBackendPrivateKey[:], TestRouterPrivateKey[:])
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

	err = geoClient.Add(1, 0, 0)
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

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, 10*time.Second, &db, &rp, &iploc, &geoClient, &sessionMetrics, &billing.NoOpBiller{}, TestServerBackendPrivateKey[:], TestRouterPrivateKey[:])
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

	err = geoClient.Add(1, 0, 0)
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

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, 10*time.Second, &db, &rp, &iploc, &geoClient, &sessionMetrics, &badBiller{}, TestServerBackendPrivateKey[:], TestRouterPrivateKey[:])
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
		RouteShader: routing.LocalRouteShader,
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

	err = geoClient.Add(1, 0, 0)
	assert.NoError(t, err)

	rp := mockRouteProvider{
		datacenterRelays: []routing.Relay{{}},
		routes: []routing.Route{
			{
				Relays: []routing.Relay{
					{ID: 1, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 123}, PublicKey: TestRelayPublicKey[:], MaxSessions: 3000},
					{ID: 2, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.2"), Port: 123}, PublicKey: TestRelayPublicKey[:], MaxSessions: 3000},
					{ID: 3, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.3"), Port: 123}, PublicKey: TestRelayPublicKey[:], MaxSessions: 3000},
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
		Datacenter: routing.Datacenter{
			Enabled: true,
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
		RouteDecision: routing.Decision{
			Reason:        routing.DecisionInitialSlice,
			OnNetworkNext: true,
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

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, 10*time.Second, &db, &rp, &iploc, &geoClient, &sessionMetrics, &billing.NoOpBiller{}, TestServerBackendPrivateKey[:], TestRouterPrivateKey[:])
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	globalDelta, err := redisServer.ZScore("total-next", fmt.Sprintf("%016x", packet.SessionID))
	assert.NoError(t, err)
	assert.Equal(t, globalDelta, float64(20))

	assert.True(t, redisServer.Exists(fmt.Sprintf("session-%016x-meta", packet.SessionID)))
	assert.Greater(t, redisServer.TTL(fmt.Sprintf("session-%016x-meta", packet.SessionID)).Hours(), float64(-1))

	assert.True(t, redisServer.Exists(fmt.Sprintf("session-%016x-slices", packet.SessionID)))
	assert.Greater(t, redisServer.TTL(fmt.Sprintf("session-%016x-slices", packet.SessionID)).Hours(), float64(-1))

	assert.True(t, redisServer.Exists(fmt.Sprintf("user-%016x-sessions", 0)))
	assert.Greater(t, redisServer.TTL(fmt.Sprintf("user-%016x-sessions", 0)).Hours(), float64(-1))

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
		RouteShader: routing.LocalRouteShader,
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

	err = geoClient.Add(1, 0, 0)
	assert.NoError(t, err)

	rp := mockRouteProvider{
		datacenterRelays: []routing.Relay{{}},
		routes: []routing.Route{
			{
				Relays: []routing.Relay{
					{ID: 1, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 123}, PublicKey: TestRelayPublicKey[:], MaxSessions: 3000},
					{ID: 2, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.2"), Port: 123}, PublicKey: TestRelayPublicKey[:], MaxSessions: 3000},
					{ID: 3, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.3"), Port: 123}, PublicKey: TestRelayPublicKey[:], MaxSessions: 3000},
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
		Datacenter: routing.Datacenter{
			Enabled: true,
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
		RouteDecision: routing.Decision{
			Reason:        routing.DecisionRTTReduction,
			OnNetworkNext: true,
		},
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

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, 10*time.Second, &db, &rp, &iploc, &geoClient, &sessionMetrics, &billing.NoOpBiller{}, TestServerBackendPrivateKey[:], TestRouterPrivateKey[:])
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	globalDelta, err := redisServer.ZScore("total-next", fmt.Sprintf("%016x", packet.SessionID))
	assert.NoError(t, err)
	assert.Equal(t, globalDelta, float64(20))

	assert.True(t, redisServer.Exists(fmt.Sprintf("session-%016x-meta", packet.SessionID)))
	assert.Greater(t, redisServer.TTL(fmt.Sprintf("session-%016x-meta", packet.SessionID)).Hours(), float64(-1))

	assert.True(t, redisServer.Exists(fmt.Sprintf("session-%016x-slices", packet.SessionID)))
	assert.Greater(t, redisServer.TTL(fmt.Sprintf("session-%016x-slices", packet.SessionID)).Hours(), float64(-1))

	assert.True(t, redisServer.Exists(fmt.Sprintf("user-%016x-sessions", 0)))
	assert.Greater(t, redisServer.TTL(fmt.Sprintf("user-%016x-sessions", 0)).Hours(), float64(-1))

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
		RouteShader: routing.LocalRouteShader,
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

	err = geoClient.Add(1, 0, 0)
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
		Datacenter: routing.Datacenter{
			Enabled: true,
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

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, 10*time.Second, &db, &rp, &iploc, &geoClient, &sessionMetrics, &billing.NoOpBiller{}, TestServerBackendPrivateKey[:], TestRouterPrivateKey[:])
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
		RouteShader: routing.LocalRouteShader,
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

	err = geoClient.Add(1, 0, 0)
	assert.NoError(t, err)

	rp := mockRouteProvider{
		datacenterRelays: []routing.Relay{{}},
		routes: []routing.Route{
			{
				Relays: []routing.Relay{
					{ID: 1, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 123}, PublicKey: TestRelayPublicKey[:], MaxSessions: 3000},
					{ID: 2, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.2"), Port: 123}, PublicKey: TestRelayPublicKey[:], MaxSessions: 3000},
					{ID: 3, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.3"), Port: 123}, PublicKey: TestRelayPublicKey[:], MaxSessions: 3000},
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
		Datacenter: routing.Datacenter{
			Enabled: true,
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
		RouteDecision: routing.Decision{
			OnNetworkNext: false,
			Reason:        routing.DecisionVetoRTT,
		},
	}
	sessionCacheEntryData, err := sessionCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SESSION-0-9999", string(sessionCacheEntryData))
	assert.NoError(t, err)

	vetoCacheEntry := transport.VetoCacheEntry{
		VetoTimestamp: time.Now().Add(5 * time.Second),
		Reason:        routing.DecisionVetoRTT,
	}
	vetoCacheEntryData, err := vetoCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("VETO-0-9999", string(vetoCacheEntryData))
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

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, 10*time.Second, &db, &rp, &iploc, &geoClient, &sessionMetrics, &billing.NoOpBiller{}, TestServerBackendPrivateKey[:], TestRouterPrivateKey[:])
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
		RouteShader: routing.LocalRouteShader,
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

	err = geoClient.Add(1, 0, 0)
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
		Datacenter: routing.Datacenter{
			Enabled: true,
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
		RouteDecision: routing.Decision{
			OnNetworkNext: false,
			Reason:        routing.DecisionVetoRTT,
		},
	}
	sessionCacheEntryData, err := sessionCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SESSION-0-9999", string(sessionCacheEntryData))
	assert.NoError(t, err)

	vetoCacheEntry := transport.VetoCacheEntry{
		VetoTimestamp: time.Now().Add(-5 * time.Second),
	}
	vetoCacheEntryData, err := vetoCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("VETO-0-9999", string(vetoCacheEntryData))
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

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, 10*time.Second, &db, &rp, &iploc, &geoClient, &sessionMetrics, &billing.NoOpBiller{}, TestServerBackendPrivateKey[:], TestRouterPrivateKey[:])
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
		RouteShader: routing.LocalRouteShader,
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

	err = geoClient.Add(1, 0, 0)
	assert.NoError(t, err)

	rp := mockRouteProvider{
		datacenterRelays: []routing.Relay{{}},
		routes: []routing.Route{
			{
				Relays: []routing.Relay{
					{ID: 1, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 123}, PublicKey: TestRelayPublicKey[:], MaxSessions: 3000},
					{ID: 2, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.2"), Port: 123}, PublicKey: TestRelayPublicKey[:], MaxSessions: 3000},
					{ID: 3, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.3"), Port: 123}, PublicKey: TestRelayPublicKey[:], MaxSessions: 3000},
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
		Datacenter: routing.Datacenter{
			Enabled: true,
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
		RouteDecision: routing.Decision{
			OnNetworkNext: false,
			Reason:        routing.DecisionVetoPacketLoss,
		},
	}
	sessionCacheEntryData, err := sessionCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SESSION-0-9999", string(sessionCacheEntryData))
	assert.NoError(t, err)

	vetoCacheEntry := transport.VetoCacheEntry{
		VetoTimestamp: time.Now().Add(5 * time.Second),
		Reason:        routing.DecisionVetoPacketLoss,
	}
	vetoCacheEntryData, err := vetoCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("VETO-0-9999", string(vetoCacheEntryData))
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

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, 10*time.Second, &db, &rp, &iploc, &geoClient, &sessionMetrics, &billing.NoOpBiller{}, TestServerBackendPrivateKey[:], TestRouterPrivateKey[:])
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
		RouteShader: routing.LocalRouteShader,
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

	err = geoClient.Add(1, 0, 0)
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
		Datacenter: routing.Datacenter{
			Enabled: true,
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
		RouteDecision: routing.Decision{
			OnNetworkNext: false,
			Reason:        routing.DecisionVetoPacketLoss,
		},
	}
	sessionCacheEntryData, err := sessionCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SESSION-0-9999", string(sessionCacheEntryData))
	assert.NoError(t, err)

	vetoCacheEntry := transport.VetoCacheEntry{
		VetoTimestamp: time.Now().Add(-5 * time.Second),
	}
	vetoCacheEntryData, err := vetoCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("VETO-0-9999", string(vetoCacheEntryData))
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

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, 10*time.Second, &db, &rp, &iploc, &geoClient, &sessionMetrics, &billing.NoOpBiller{}, TestServerBackendPrivateKey[:], TestRouterPrivateKey[:])
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

	rrs := routing.LocalRouteShader
	rrs.EnableYouOnlyLiveOnce = true

	db := storage.InMemory{}
	db.AddBuyer(context.Background(), routing.Buyer{
		PublicKey:            TestBuyersServerPublicKey,
		RouteShader: rrs,
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

	err = geoClient.Add(1, 0, 0)
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
		Datacenter: routing.Datacenter{
			Enabled: true,
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
		RouteDecision: routing.Decision{
			OnNetworkNext: false,
			Reason:        routing.DecisionVetoRTT | routing.DecisionVetoYOLO,
		},
	}
	sessionCacheEntryData, err := sessionCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SESSION-0-9999", string(sessionCacheEntryData))
	assert.NoError(t, err)

	vetoCacheEntry := transport.VetoCacheEntry{
		VetoTimestamp: time.Now().Add(-5 * time.Second),
	}
	vetoCacheEntryData, err := vetoCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("VETO-0-9999", string(vetoCacheEntryData))
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

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, 10*time.Second, &db, &rp, &iploc, &geoClient, &sessionMetrics, &billing.NoOpBiller{}, TestServerBackendPrivateKey[:], TestRouterPrivateKey[:])
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	validateDirectResponsePacket(t, resbuf, sessionMetrics.DirectSessions, sessionMetrics.DecisionMetrics.VetoRTTYOLO)
}

func TestEarlySlice(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	sessionMetrics := metrics.EmptySessionMetrics
	localMetrics := metrics.LocalHandler{}

	decisionMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "decision metric"})
	assert.NoError(t, err)
	nextMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "route metric"})
	assert.NoError(t, err)

	sessionMetrics.DecisionMetrics.RTTReduction = decisionMetric
	sessionMetrics.NextSessions = nextMetric

	routeRuleSettings := routing.LocalRouteShader
	routeRuleSettings.SelectionPercentage = 100
	db := storage.InMemory{}
	db.AddBuyer(context.Background(), routing.Buyer{
		PublicKey:            TestBuyersServerPublicKey,
		RouteShader: routeRuleSettings,
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

	err = geoClient.Add(1, 0, 0)
	assert.NoError(t, err)

	rp := mockRouteProvider{
		datacenterRelays: []routing.Relay{{}},
		routes: []routing.Route{
			{
				Relays: []routing.Relay{
					{ID: 1, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 123}, PublicKey: TestRelayPublicKey[:], MaxSessions: 3000},
					{ID: 2, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.2"), Port: 123}, PublicKey: TestRelayPublicKey[:], MaxSessions: 3000},
					{ID: 3, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.3"), Port: 123}, PublicKey: TestRelayPublicKey[:], MaxSessions: 3000},
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
		Datacenter: routing.Datacenter{
			Enabled: true,
		},
	}
	serverCacheEntryData, err := serverCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SERVER-0-0.0.0.0:13", string(serverCacheEntryData))
	assert.NoError(t, err)

	sessionCacheEntry := transport.SessionCacheEntry{
		SessionID:        9999,
		Sequence:         13,
		OnNNSliceCounter: 2,
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

		OnNetworkNext: true,
		DirectMinRTT:  40,
		NextMinRTT:    30,

		NextPacketLoss: 1.0, // Should server a NN route even though we have high packet loss
	}
	packet.Signature = crypto.Sign(TestBuyersServerPrivateKey, packet.GetSignData())

	data, err := packet.MarshalBinary()
	assert.NoError(t, err)

	var resbuf bytes.Buffer

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, 10*time.Second, &db, &rp, &iploc, &geoClient, &sessionMetrics, &billing.NoOpBiller{}, TestServerBackendPrivateKey[:], TestRouterPrivateKey[:])
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	validateNextResponsePacket(t, resbuf, packet.SessionID, packet.Sequence, 5, routing.RouteTypeNew, sessionMetrics.NextSessions, sessionMetrics.DecisionMetrics.RTTReduction)
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

		routeRuleSettings := routing.LocalRouteShader
		routeRuleSettings.SelectionPercentage = 50
		db := storage.InMemory{}
		db.AddBuyer(context.Background(), routing.Buyer{
			PublicKey:            TestBuyersServerPublicKey,
			RouteShader: routeRuleSettings,
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

		err = geoClient.Add(1, 0, 0)
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
			Datacenter: routing.Datacenter{
				Enabled: true,
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

		handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, 10*time.Second, &db, &rp, &iploc, &geoClient, &sessionMetrics, &billing.NoOpBiller{}, TestServerBackendPrivateKey[:], TestRouterPrivateKey[:])
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

		routeRuleSettings := routing.LocalRouteShader
		routeRuleSettings.Mode = routing.ModeForceDirect
		db := storage.InMemory{}
		db.AddBuyer(context.Background(), routing.Buyer{
			PublicKey:            TestBuyersServerPublicKey,
			RouteShader: routeRuleSettings,
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

		err = geoClient.Add(1, 0, 0)
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
			Datacenter: routing.Datacenter{
				Enabled: true,
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

		handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, 10*time.Second, &db, &rp, &iploc, &geoClient, &sessionMetrics, &billing.NoOpBiller{}, TestServerBackendPrivateKey[:], TestRouterPrivateKey[:])
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

	routeRuleSettings := routing.LocalRouteShader
	routeRuleSettings.Mode = routing.ModeForceNext
	routeRuleSettings.SelectionPercentage = 100
	db := storage.InMemory{}
	db.AddBuyer(context.Background(), routing.Buyer{
		PublicKey:            TestBuyersServerPublicKey,
		RouteShader: routeRuleSettings,
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

	err = geoClient.Add(1, 0, 0)
	assert.NoError(t, err)

	rp := mockRouteProvider{
		datacenterRelays: []routing.Relay{{}},
		routes: []routing.Route{
			{
				Relays: []routing.Relay{
					{ID: 1, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 123}, PublicKey: TestRelayPublicKey[:], MaxSessions: 3000},
					{ID: 2, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.2"), Port: 123}, PublicKey: TestRelayPublicKey[:], MaxSessions: 3000},
					{ID: 3, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.3"), Port: 123}, PublicKey: TestRelayPublicKey[:], MaxSessions: 3000},
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
		Datacenter: routing.Datacenter{
			Enabled: true,
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

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, 10*time.Second, &db, &rp, &iploc, &geoClient, &sessionMetrics, &billing.NoOpBiller{}, TestServerBackendPrivateKey[:], TestRouterPrivateKey[:])
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

	routeRuleSettings := routing.LocalRouteShader
	routeRuleSettings.EnableABTest = true
	db := storage.InMemory{}
	db.AddBuyer(context.Background(), routing.Buyer{
		PublicKey:            TestBuyersServerPublicKey,
		RouteShader: routeRuleSettings,
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

	err = geoClient.Add(1, 0, 0)
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
		Datacenter: routing.Datacenter{
			Enabled: true,
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

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, 10*time.Second, &db, &rp, &iploc, &geoClient, &sessionMetrics, &billing.NoOpBiller{}, TestServerBackendPrivateKey[:], TestRouterPrivateKey[:])
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

	rrs := routing.LocalRouteShader
	rrs.EnableTryBeforeYouBuy = true

	db := storage.InMemory{}
	err = db.AddBuyer(context.Background(), routing.Buyer{
		PublicKey:            TestBuyersServerPublicKey,
		RouteShader: rrs,
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

	err = geoClient.Add(1, 0, 0)
	assert.NoError(t, err)

	rp := mockRouteProvider{
		datacenterRelays: []routing.Relay{{}},
		routes: []routing.Route{
			{
				Relays: []routing.Relay{
					{ID: 1, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 123}, PublicKey: TestRelayPublicKey[:], MaxSessions: 3000},
					{ID: 2, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.2"), Port: 123}, PublicKey: TestRelayPublicKey[:], MaxSessions: 3000},
					{ID: 3, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.3"), Port: 123}, PublicKey: TestRelayPublicKey[:], MaxSessions: 3000},
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
		Datacenter: routing.Datacenter{
			Enabled: true,
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

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, 10*time.Second, &db, &rp, &iploc, &geoClient, &sessionMetrics, &billing.NoOpBiller{}, TestServerBackendPrivateKey[:], TestRouterPrivateKey[:])
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

	rrs := routing.LocalRouteShader
	rrs.EnableTryBeforeYouBuy = true

	db := storage.InMemory{}
	err = db.AddBuyer(context.Background(), routing.Buyer{
		PublicKey:            TestBuyersServerPublicKey,
		RouteShader: rrs,
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

	err = geoClient.Add(1, 0, 0)
	assert.NoError(t, err)

	rp := mockRouteProvider{
		datacenterRelays: []routing.Relay{{}},
		routes: []routing.Route{
			{
				Relays: []routing.Relay{
					{ID: 1, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 123}, PublicKey: TestRelayPublicKey[:], MaxSessions: 3000},
					{ID: 2, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.2"), Port: 123}, PublicKey: TestRelayPublicKey[:], MaxSessions: 3000},
					{ID: 3, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.3"), Port: 123}, PublicKey: TestRelayPublicKey[:], MaxSessions: 3000},
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
		Datacenter: routing.Datacenter{
			Enabled: true,
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

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, 10*time.Second, &db, &rp, &iploc, &geoClient, &sessionMetrics, &billing.NoOpBiller{}, TestServerBackendPrivateKey[:], TestRouterPrivateKey[:])
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	actual := validateNextResponsePacket(t, resbuf, packet.SessionID, packet.Sequence, 5, routing.RouteTypeNew, sessionMetrics.NextSessions, sessionMetrics.DecisionMetrics.NoReason)
	assert.Equal(t, true, actual.Committed)
}

func TestMultipathRTT(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	sessionMetrics := metrics.EmptySessionMetrics
	localMetrics := metrics.LocalHandler{}

	decisionMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "decision metric"})
	assert.NoError(t, err)
	nextMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "route metric"})
	assert.NoError(t, err)

	sessionMetrics.DecisionMetrics.RTTMultipath = decisionMetric
	sessionMetrics.NextSessions = nextMetric

	rrs := routing.LocalRouteShader
	rrs.EnableMultipathForRTT = true

	db := storage.InMemory{}
	err = db.AddBuyer(context.Background(), routing.Buyer{
		PublicKey:            TestBuyersServerPublicKey,
		RouteShader: rrs,
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

	err = geoClient.Add(1, 0, 0)
	assert.NoError(t, err)

	rp := mockRouteProvider{
		datacenterRelays: []routing.Relay{{}},
		routes: []routing.Route{
			{
				Relays: []routing.Relay{
					{ID: 1, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 123}, PublicKey: TestRelayPublicKey[:], MaxSessions: 3000},
					{ID: 2, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.2"), Port: 123}, PublicKey: TestRelayPublicKey[:], MaxSessions: 3000},
					{ID: 3, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.3"), Port: 123}, PublicKey: TestRelayPublicKey[:], MaxSessions: 3000},
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
		Datacenter: routing.Datacenter{
			Enabled: true,
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

		DirectMinRTT: 30,

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

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, 10*time.Second, &db, &rp, &iploc, &geoClient, &sessionMetrics, &billing.NoOpBiller{}, TestServerBackendPrivateKey[:], TestRouterPrivateKey[:])
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	actual := validateNextResponsePacket(t, resbuf, packet.SessionID, packet.Sequence, 5, routing.RouteTypeNew, sessionMetrics.NextSessions, sessionMetrics.DecisionMetrics.RTTMultipath)
	assert.Equal(t, true, actual.Multipath)
}

func TestMultipathJitter(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	sessionMetrics := metrics.EmptySessionMetrics
	localMetrics := metrics.LocalHandler{}

	decisionMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "decision metric"})
	assert.NoError(t, err)
	nextMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "route metric"})
	assert.NoError(t, err)

	sessionMetrics.DecisionMetrics.JitterMultipath = decisionMetric
	sessionMetrics.NextSessions = nextMetric

	rrs := routing.LocalRouteShader
	rrs.EnableMultipathForJitter = true

	db := storage.InMemory{}
	err = db.AddBuyer(context.Background(), routing.Buyer{
		PublicKey:            TestBuyersServerPublicKey,
		RouteShader: rrs,
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

	err = geoClient.Add(1, 0, 0)
	assert.NoError(t, err)

	rp := mockRouteProvider{
		datacenterRelays: []routing.Relay{{}},
		routes: []routing.Route{
			{
				Relays: []routing.Relay{
					{ID: 1, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 123}, PublicKey: TestRelayPublicKey[:], MaxSessions: 3000},
					{ID: 2, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.2"), Port: 123}, PublicKey: TestRelayPublicKey[:], MaxSessions: 3000},
					{ID: 3, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.3"), Port: 123}, PublicKey: TestRelayPublicKey[:], MaxSessions: 3000},
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
		Datacenter: routing.Datacenter{
			Enabled: true,
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

		DirectJitter: 50,

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

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, 10*time.Second, &db, &rp, &iploc, &geoClient, &sessionMetrics, &billing.NoOpBiller{}, TestServerBackendPrivateKey[:], TestRouterPrivateKey[:])
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	actual := validateNextResponsePacket(t, resbuf, packet.SessionID, packet.Sequence, 5, routing.RouteTypeNew, sessionMetrics.NextSessions, sessionMetrics.DecisionMetrics.JitterMultipath)
	assert.Equal(t, true, actual.Multipath)
}

func TestMultipathPacketLoss(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	sessionMetrics := metrics.EmptySessionMetrics
	localMetrics := metrics.LocalHandler{}

	decisionMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "decision metric"})
	assert.NoError(t, err)
	nextMetric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "route metric"})
	assert.NoError(t, err)

	sessionMetrics.DecisionMetrics.PacketLossMultipath = decisionMetric
	sessionMetrics.NextSessions = nextMetric

	rrs := routing.LocalRouteShader
	rrs.EnableMultipathForPacketLoss = true

	db := storage.InMemory{}
	err = db.AddBuyer(context.Background(), routing.Buyer{
		PublicKey:            TestBuyersServerPublicKey,
		RouteShader: rrs,
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

	err = geoClient.Add(1, 0, 0)
	assert.NoError(t, err)

	rp := mockRouteProvider{
		datacenterRelays: []routing.Relay{{}},
		routes: []routing.Route{
			{
				Relays: []routing.Relay{
					{ID: 1, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 123}, PublicKey: TestRelayPublicKey[:], MaxSessions: 3000},
					{ID: 2, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.2"), Port: 123}, PublicKey: TestRelayPublicKey[:], MaxSessions: 3000},
					{ID: 3, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.3"), Port: 123}, PublicKey: TestRelayPublicKey[:], MaxSessions: 3000},
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
		Datacenter: routing.Datacenter{
			Enabled: true,
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

		DirectPacketLoss: 1,

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

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, redisClient, 10*time.Second, &db, &rp, &iploc, &geoClient, &sessionMetrics, &billing.NoOpBiller{}, TestServerBackendPrivateKey[:], TestRouterPrivateKey[:])
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	actual := validateNextResponsePacket(t, resbuf, packet.SessionID, packet.Sequence, 5, routing.RouteTypeNew, sessionMetrics.NextSessions, sessionMetrics.DecisionMetrics.PacketLossMultipath)
	assert.Equal(t, true, actual.Multipath)
}
*/
