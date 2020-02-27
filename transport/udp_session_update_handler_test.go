package transport_test

import (
	"bytes"
	"crypto/ed25519"
	"errors"
	"net"
	"testing"

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

func ValidateDirectResponsePacket(resbuf bytes.Buffer, t *testing.T) {
	assert.Greater(t, resbuf.Len(), 0)

	var actual transport.SessionResponsePacket
	err := actual.UnmarshalBinary(resbuf.Bytes())
	assert.NoError(t, err)

	verified := crypto.Verify(TestServerBackendPublicKey, actual.GetSignData(), actual.Signature)
	assert.True(t, verified)

	assert.Equal(t, int(actual.RouteType), routing.RouteTypeDirect)
}

func TestFailToUnmarshalSessionUpdate(t *testing.T) {
	redisServer, _ := miniredis.Run()
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
	assert.NoError(t, err)

	var resbuf bytes.Buffer

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, nil, nil, nil, nil, &metrics.NoOpHandler{}, nil, nil, nil)
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
			PublicKey: TestServerBackendPublicKey,
		},
	}
	serverCacheEntryData, err := serverCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SERVER-0.0.0.0:13", string(serverCacheEntryData))
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

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, &db, nil, nil, nil, &metrics.NoOpHandler{}, nil, TestServerBackendPrivateKey, nil)
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	ValidateDirectResponsePacket(resbuf, t)
}

func TestVerificationFailed(t *testing.T) {
	redisServer, _ := miniredis.Run()
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	db := storage.InMemory{
		LocalBuyer: &routing.Buyer{
			PublicKey: TestServerBackendPublicKey, // normally would be buyers public key, intentionally using servers to cause error
		},
	}

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

	err = redisServer.Set("SERVER-0.0.0.0:13", string(serverCacheEntryData))
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

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, &db, nil, nil, nil, &metrics.NoOpHandler{}, nil, TestServerBackendPrivateKey, nil)
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	ValidateDirectResponsePacket(resbuf, t)
}

func TestSessionPacketSequenceTooOld(t *testing.T) {
	redisServer, _ := miniredis.Run()
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	db := storage.InMemory{
		LocalBuyer: &routing.Buyer{
			PublicKey: TestBuyersServerPublicKey,
		},
	}

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

	err = redisServer.Set("SERVER-0.0.0.0:13", string(serverCacheEntryData))
	assert.NoError(t, err)

	sessionCacheEntry := transport.SessionCacheEntry{
		SessionID: 9999,
		Sequence:  13,
	}
	sessionCacheEntryData, err := sessionCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SESSION-9999", string(sessionCacheEntryData))
	assert.NoError(t, err)

	packet := transport.SessionUpdatePacket{
		SessionId:     9999,
		Sequence:      1,
		ServerAddress: net.UDPAddr{IP: net.IPv4zero, Port: 13},

		ClientRoutePublicKey: make([]byte, crypto.KeySize),

		Signature: make([]byte, ed25519.SignatureSize),
	}
	packet.Signature = crypto.Sign(TestBuyersServerPrivateKey, packet.GetSignData())

	data, err := packet.MarshalBinary()
	assert.NoError(t, err)

	var resbuf bytes.Buffer

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, &db, nil, nil, nil, &metrics.NoOpHandler{}, nil, TestServerBackendPrivateKey, nil)
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	ValidateDirectResponsePacket(resbuf, t)
}

func TestClientIPLookupFail(t *testing.T) {
	buyersServerPubKey, buyersServerPrivKey, err := ed25519.GenerateKey(nil)
	assert.NoError(t, err)

	redisServer, _ := miniredis.Run()
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	db := storage.InMemory{
		LocalBuyer: &routing.Buyer{
			PublicKey: buyersServerPubKey,
		},
	}

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

	err = redisServer.Set("SERVER-0.0.0.0:13", string(serverCacheEntryData))
	assert.NoError(t, err)

	sessionCacheEntry := transport.SessionCacheEntry{
		SessionID: 9999,
		Sequence:  13,
	}
	sessionCacheEntryData, err := sessionCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SESSION-9999", string(sessionCacheEntryData))
	assert.NoError(t, err)

	packet := transport.SessionUpdatePacket{
		SessionId:     9999,
		Sequence:      14,
		ServerAddress: net.UDPAddr{IP: net.IPv4zero, Port: 13},

		ClientRoutePublicKey: make([]byte, crypto.KeySize),

		Signature: make([]byte, ed25519.SignatureSize),
	}
	packet.Signature = crypto.Sign(buyersServerPrivKey, packet.GetSignData())

	data, err := packet.MarshalBinary()
	assert.NoError(t, err)

	var resbuf bytes.Buffer

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, &db, nil, &iploc, nil, &metrics.NoOpHandler{}, nil, TestServerBackendPrivateKey, nil)
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	ValidateDirectResponsePacket(resbuf, t)
}

func TestNoRelaysNearClient(t *testing.T) {
	redisServer, _ := miniredis.Run()
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	db := storage.InMemory{
		LocalBuyer: &routing.Buyer{
			PublicKey: TestBuyersServerPublicKey,
		},
	}

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

	err = redisServer.Set("SERVER-0.0.0.0:13", string(serverCacheEntryData))
	assert.NoError(t, err)

	sessionCacheEntry := transport.SessionCacheEntry{
		SessionID: 9999,
		Sequence:  13,
	}
	sessionCacheEntryData, err := sessionCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SESSION-9999", string(sessionCacheEntryData))
	assert.NoError(t, err)

	packet := transport.SessionUpdatePacket{
		SessionId:     9999,
		Sequence:      14,
		ServerAddress: net.UDPAddr{IP: net.IPv4zero, Port: 13},

		ClientRoutePublicKey: make([]byte, crypto.KeySize),

		Signature: make([]byte, ed25519.SignatureSize),
	}
	packet.Signature = crypto.Sign(TestBuyersServerPrivateKey, packet.GetSignData())

	data, err := packet.MarshalBinary()
	assert.NoError(t, err)

	var resbuf bytes.Buffer

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, &db, &rp, &iploc, &geoClient, &metrics.NoOpHandler{}, &billing.NoOpBiller{}, TestServerBackendPrivateKey, nil)
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	ValidateDirectResponsePacket(resbuf, t)
}

func TestNoRoutesFound(t *testing.T) {
	redisServer, _ := miniredis.Run()
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	db := storage.InMemory{
		LocalBuyer: &routing.Buyer{
			PublicKey: TestBuyersServerPublicKey,
		},
	}

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
		ID:        1,
		Latitude:  0,
		Longitude: 0,
	}
	err := geoClient.Add(nearbyRelay)

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

	err = redisServer.Set("SERVER-0.0.0.0:13", string(serverCacheEntryData))
	assert.NoError(t, err)

	sessionCacheEntry := transport.SessionCacheEntry{
		SessionID: 9999,
		Sequence:  13,
	}
	sessionCacheEntryData, err := sessionCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SESSION-9999", string(sessionCacheEntryData))
	assert.NoError(t, err)

	packet := transport.SessionUpdatePacket{
		SessionId:     9999,
		Sequence:      14,
		ServerAddress: net.UDPAddr{IP: net.IPv4zero, Port: 13},

		ClientRoutePublicKey: make([]byte, crypto.KeySize),

		Signature: make([]byte, ed25519.SignatureSize),
	}
	packet.Signature = crypto.Sign(TestBuyersServerPrivateKey, packet.GetSignData())

	data, err := packet.MarshalBinary()
	assert.NoError(t, err)

	var resbuf bytes.Buffer

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, &db, &rp, &iploc, &geoClient, &metrics.NoOpHandler{}, &billing.NoOpBiller{}, TestServerBackendPrivateKey, nil)
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	ValidateDirectResponsePacket(resbuf, t)
}

func TestNextRouteResponse(t *testing.T) {
	redisServer, _ := miniredis.Run()
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	db := storage.InMemory{
		LocalBuyer: &routing.Buyer{
			PublicKey: TestBuyersServerPublicKey,
		},
	}

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
		ID:        1,
		Latitude:  0,
		Longitude: 0,
	}
	err := geoClient.Add(nearbyRelay)

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

	err = redisServer.Set("SERVER-0.0.0.0:13", string(serverCacheEntryData))
	assert.NoError(t, err)

	sessionCacheEntry := transport.SessionCacheEntry{
		SessionID: 9999,
		Sequence:  13,
	}
	sessionCacheEntryData, err := sessionCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SESSION-9999", string(sessionCacheEntryData))
	assert.NoError(t, err)

	packet := transport.SessionUpdatePacket{
		SessionId:     9999,
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

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, &db, &rp, &iploc, &geoClient, &metrics.NoOpHandler{}, &billing.NoOpBiller{}, TestServerBackendPrivateKey[:], TestRouterPrivateKey[:])
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	assert.Greater(t, resbuf.Len(), 0)

	var actual transport.SessionResponsePacket
	err = actual.UnmarshalBinary(resbuf.Bytes())
	assert.NoError(t, err)

	verified := crypto.Verify(TestServerBackendPublicKey, actual.GetSignData(), actual.Signature)
	assert.True(t, verified)

	assert.Equal(t, packet.SessionId, actual.SessionId)
	assert.Equal(t, packet.Sequence, actual.Sequence)
	assert.Equal(t, int32(routing.RouteTypeNew), actual.RouteType)
	assert.Equal(t, int32(5), actual.NumTokens)
	assert.Equal(t, TestBuyersServerPublicKey[:], actual.ServerRoutePublicKey)
}

func TestContinueRouteResponse(t *testing.T) {
	redisServer, _ := miniredis.Run()
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	db := storage.InMemory{
		LocalBuyer: &routing.Buyer{
			PublicKey: TestBuyersServerPublicKey,
		},
	}

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
		ID:        1,
		Latitude:  0,
		Longitude: 0,
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

	err = redisServer.Set("SERVER-0.0.0.0:13", string(se))
	assert.NoError(t, err)

	expectedsession := transport.SessionCacheEntry{
		SessionID: 9999,
		Sequence:  13,
		RouteHash: 1511739644222804357,
	}
	sce, err := expectedsession.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SESSION-9999", string(sce))
	assert.NoError(t, err)

	packet := transport.SessionUpdatePacket{
		SessionId:     9999,
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

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, &db, &rp, &iploc, &geoClient, &metrics.NoOpHandler{}, &billing.NoOpBiller{}, TestServerBackendPrivateKey[:], TestRouterPrivateKey[:])
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	var actual transport.SessionResponsePacket
	err = actual.UnmarshalBinary(resbuf.Bytes())
	assert.NoError(t, err)

	verified := crypto.Verify(TestServerBackendPublicKey, actual.GetSignData(), actual.Signature)
	assert.True(t, verified)

	assert.Equal(t, packet.SessionId, actual.SessionId)
	assert.Equal(t, packet.Sequence, actual.Sequence)
	assert.Equal(t, int32(routing.RouteTypeContinue), actual.RouteType)
	assert.Equal(t, int32(5), actual.NumTokens)
	assert.Equal(t, TestBuyersServerPublicKey[:], actual.ServerRoutePublicKey)
}

func TestCachedRouteResponse(t *testing.T) {
	redisServer, _ := miniredis.Run()
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	db := storage.InMemory{
		LocalBuyer: &routing.Buyer{
			PublicKey: TestBuyersServerPublicKey,
		},
	}

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
		ID:        1,
		Latitude:  0,
		Longitude: 0,
	}
	err := geoClient.Add(nearbyRelay)

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

	err = redisServer.Set("SERVER-0.0.0.0:13", string(serverCacheEntryData))
	assert.NoError(t, err)

	tokens := make([]byte, routing.EncryptedNextRouteTokenSize)
	copy(tokens, []byte("TEST TOKENS"))

	cachedRouteResponse := transport.SessionResponsePacket{
		SessionId:            9999,
		Sequence:             13,
		NumNearRelays:        0,
		NearRelayIds:         make([]uint64, 0),
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

	err = redisServer.Set("SESSION-9999", string(sessionCacheEntryData))
	assert.NoError(t, err)

	packet := transport.SessionUpdatePacket{
		SessionId:     9999,
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

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, &db, &rp, &iploc, &geoClient, &metrics.NoOpHandler{}, &billing.NoOpBiller{}, TestServerBackendPrivateKey[:], TestRouterPrivateKey[:])
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	assert.Greater(t, resbuf.Len(), 0)

	var actual transport.SessionResponsePacket
	err = actual.UnmarshalBinary(resbuf.Bytes())
	assert.NoError(t, err)

	verified := crypto.Verify(TestServerBackendPublicKey, actual.GetSignData(), actual.Signature)
	assert.True(t, verified)

	assert.Equal(t, packet.SessionId, actual.SessionId)
	assert.Equal(t, packet.Sequence, actual.Sequence)
	assert.Equal(t, int32(routing.RouteTypeNew), actual.RouteType)
	assert.Equal(t, int32(1), actual.NumTokens)
	assert.Equal(t, tokens, actual.Tokens)
	assert.Equal(t, TestBuyersServerPublicKey[:], actual.ServerRoutePublicKey)
}

func TestTokenEncryptionFailure(t *testing.T) {
	redisServer, _ := miniredis.Run()
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	db := storage.InMemory{
		LocalBuyer: &routing.Buyer{
			PublicKey: TestBuyersServerPublicKey,
		},
	}

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
		ID:        1,
		Latitude:  0,
		Longitude: 0,
	}
	err := geoClient.Add(nearbyRelay)

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

	err = redisServer.Set("SERVER-0.0.0.0:13", string(serverCacheEntryData))
	assert.NoError(t, err)

	sessionCacheEntry := transport.SessionCacheEntry{
		SessionID: 9999,
		Sequence:  13,
	}
	sessionCacheEntryData, err := sessionCacheEntry.MarshalBinary()
	assert.NoError(t, err)

	err = redisServer.Set("SESSION-9999", string(sessionCacheEntryData))
	assert.NoError(t, err)

	packet := transport.SessionUpdatePacket{
		SessionId:     9999,
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

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, &db, &rp, &iploc, &geoClient, &metrics.NoOpHandler{}, &billing.NoOpBiller{}, TestServerBackendPrivateKey[:], TestRouterPrivateKey[:])
	handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

	assert.Greater(t, resbuf.Len(), 0)

	var actual transport.SessionResponsePacket
	err = actual.UnmarshalBinary(resbuf.Bytes())
	assert.NoError(t, err)

	ValidateDirectResponsePacket(resbuf, t)
}
