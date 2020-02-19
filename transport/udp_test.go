package transport_test

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"errors"
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
	"golang.org/x/crypto/nacl/box"
)

type mockRouteProvider struct {
	relay            routing.Relay
	datacenterRelays []routing.Relay
	routes           []routing.Route
}

func (rp *mockRouteProvider) ResolveRelay(id uint64) (routing.Relay, error) {
	return rp.relay, nil
}

func (rp *mockRouteProvider) RelaysIn(ds routing.Datacenter) []routing.Relay {
	return rp.datacenterRelays
}

func (rp *mockRouteProvider) Routes(from []routing.Relay, to []routing.Relay) []routing.Route {
	return rp.routes
}

func TestServerUpdateHandlerFunc(t *testing.T) {
	t.Run("failed to unmarshal packet", func(t *testing.T) {
		redisServer, _ := miniredis.Run()
		redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

		addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
		assert.NoError(t, err)

		handler := transport.ServerUpdateHandlerFunc(log.NewNopLogger(), redisClient, nil)
		handler(&bytes.Buffer{}, &transport.UDPPacket{SourceAddr: addr, Data: []byte("this is not a proper packet")})

		_, err = redisServer.Get("SERVER-0.0.0.0:13")
		assert.Error(t, err)
	})

	t.Run("sdk version too old", func(t *testing.T) {
		redisServer, _ := miniredis.Run()
		redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

		addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
		assert.NoError(t, err)

		packet := transport.ServerUpdatePacket{
			Sequence:             13,
			ServerAddress:        net.UDPAddr{IP: net.IPv4zero, Port: 13},
			ServerPrivateAddress: net.UDPAddr{IP: net.IPv4zero, Port: 13},
			ServerRoutePublicKey: TestPublicKey,

			DatacenterId: 13,

			Version: transport.SDKVersion{1, 2, 3},

			Signature: make([]byte, ed25519.SignatureSize),
		}

		data, err := packet.MarshalBinary()
		assert.NoError(t, err)

		handler := transport.ServerUpdateHandlerFunc(log.NewNopLogger(), redisClient, nil)
		handler(&bytes.Buffer{}, &transport.UDPPacket{SourceAddr: addr, Data: data})

		_, err = redisServer.Get("SERVER-0.0.0.0:13")
		assert.Error(t, err)
	})

	t.Run("did not get a buyer", func(t *testing.T) {
		redisServer, _ := miniredis.Run()
		redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

		db := storage.InMemory{}

		addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
		assert.NoError(t, err)

		packet := transport.ServerUpdatePacket{
			Sequence:             13,
			ServerAddress:        net.UDPAddr{IP: net.IPv4zero, Port: 13},
			ServerPrivateAddress: net.UDPAddr{IP: net.IPv4zero, Port: 13},
			ServerRoutePublicKey: TestPublicKey,

			DatacenterId: 13,

			Version: transport.SDKVersionMin,

			Signature: make([]byte, ed25519.SignatureSize),
		}

		data, err := packet.MarshalBinary()
		assert.NoError(t, err)

		handler := transport.ServerUpdateHandlerFunc(log.NewNopLogger(), redisClient, &db)
		handler(&bytes.Buffer{}, &transport.UDPPacket{SourceAddr: addr, Data: data})

		_, err = redisServer.Get("SERVER-0.0.0.0:13")
		assert.Error(t, err)
	})

	t.Run("buyer's public key failed verification", func(t *testing.T) {
		_, buyersServerPrivKey, err := ed25519.GenerateKey(nil)
		assert.NoError(t, err)

		redisServer, _ := miniredis.Run()
		redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

		db := storage.InMemory{}

		addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
		assert.NoError(t, err)

		packet := transport.ServerUpdatePacket{
			Sequence:             13,
			ServerAddress:        net.UDPAddr{IP: net.IPv4zero, Port: 13},
			ServerPrivateAddress: net.UDPAddr{IP: net.IPv4zero, Port: 13},
			ServerRoutePublicKey: TestPublicKey,

			DatacenterId: 13,

			Version: transport.SDKVersionMin,

			Signature: make([]byte, ed25519.SignatureSize),
		}
		packet.Signature = crypto.Sign(buyersServerPrivKey, packet.GetSignData())

		data, err := packet.MarshalBinary()
		assert.NoError(t, err)

		handler := transport.ServerUpdateHandlerFunc(log.NewNopLogger(), redisClient, &db)
		handler(&bytes.Buffer{}, &transport.UDPPacket{SourceAddr: addr, Data: data})

		_, err = redisServer.Get("SERVER-0.0.0.0:13")
		assert.Error(t, err)
	})

	t.Run("packet sequence too old", func(t *testing.T) {
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

		packet := transport.ServerUpdatePacket{
			Sequence:             1,
			ServerAddress:        net.UDPAddr{IP: net.IPv4zero, Port: 13},
			ServerPrivateAddress: net.UDPAddr{IP: net.IPv4zero, Port: 13},
			ServerRoutePublicKey: TestPublicKey,

			DatacenterId: 13,

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

		handler := transport.ServerUpdateHandlerFunc(log.NewNopLogger(), redisClient, &db)
		handler(&bytes.Buffer{}, &transport.UDPPacket{SourceAddr: addr, Data: data})

		ds, err := redisServer.Get("SERVER-0.0.0.0:13")
		assert.NoError(t, err)

		var actual transport.ServerCacheEntry
		err = actual.UnmarshalBinary([]byte(ds))
		assert.NoError(t, err)

		assert.Equal(t, expected.Sequence, actual.Sequence)
	})

	t.Run("success", func(t *testing.T) {
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

			DatacenterId: 13,

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
		handler := transport.ServerUpdateHandlerFunc(log.NewNopLogger(), redisClient, &db)

		// Invoke the handler with the data packet and address it is coming from
		handler(&buf, &incoming)

		// Get the server entry directly from the in-memory redis and assert there is no error
		ds, err := redisServer.Get("SERVER-0.0.0.0:13")
		assert.NoError(t, err)

		// Create an "expected" ServerEntry based on the incoming ServerUpdatePacket above
		expected := transport.ServerCacheEntry{
			Sequence:   13,
			Server:     routing.Server{Addr: *addr, PublicKey: packet.ServerRoutePublicKey},
			Datacenter: routing.Datacenter{ID: packet.DatacenterId},
			SDKVersion: packet.Version,
		}

		// Unmarshal the data in redis to the actual ServerEntry saved
		var actual transport.ServerCacheEntry
		err = actual.UnmarshalBinary([]byte(ds))
		assert.NoError(t, err)

		// Finally compare both ServerEntry struct to ensure we saved the right data in redis
		assert.Equal(t, expected, actual)

		assert.Equal(t, 0, buf.Len())
	})
}

func TestSessionUpdateHandlerFunc(t *testing.T) {
	t.Run("failed to unmarshal packet", func(t *testing.T) {

		redisServer, _ := miniredis.Run()
		redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

		addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
		assert.NoError(t, err)

		var resbuf bytes.Buffer

		handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, nil, nil, nil, nil, nil, nil)
		handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: []byte("this is not a proper packet")})

		assert.Equal(t, 0, resbuf.Len())
	})

	t.Run("did not get a buyer", func(t *testing.T) {
		redisServer, _ := miniredis.Run()
		redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

		db := storage.InMemory{}

		addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
		assert.NoError(t, err)

		expected := transport.ServerCacheEntry{
			Sequence: 13,
		}
		se, err := expected.MarshalBinary()
		assert.NoError(t, err)

		err = redisServer.Set("SERVER-0.0.0.0:13", string(se))
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

		handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, &db, nil, nil, nil, nil, nil)
		handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

		assert.Equal(t, 0, resbuf.Len())
	})

	t.Run("buyer's public key failed verification", func(t *testing.T) {
		_, buyersServerPrivKey, err := ed25519.GenerateKey(nil)
		assert.NoError(t, err)

		redisServer, _ := miniredis.Run()
		redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

		db := storage.InMemory{}

		addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
		assert.NoError(t, err)

		expected := transport.ServerCacheEntry{
			Sequence: 13,
		}
		se, err := expected.MarshalBinary()
		assert.NoError(t, err)

		err = redisServer.Set("SERVER-0.0.0.0:13", string(se))
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

		handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, &db, nil, nil, nil, nil, nil)
		handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

		assert.Equal(t, 0, resbuf.Len())
	})

	t.Run("packet sequence too old", func(t *testing.T) {

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

		expected := transport.ServerCacheEntry{
			Sequence: 13,
		}
		se, err := expected.MarshalBinary()
		assert.NoError(t, err)

		err = redisServer.Set("SERVER-0.0.0.0:13", string(se))
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
		packet.Signature = crypto.Sign(buyersServerPrivKey, packet.GetSignData())

		data, err := packet.MarshalBinary()
		assert.NoError(t, err)

		var resbuf bytes.Buffer

		handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, &db, nil, nil, nil, nil, nil)
		handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

		assert.Equal(t, 0, resbuf.Len())
	})

	t.Run("client ip lookup failed", func(t *testing.T) {
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

		expected := transport.ServerCacheEntry{
			Sequence: 13,
		}
		se, err := expected.MarshalBinary()
		assert.NoError(t, err)

		err = redisServer.Set("SERVER-0.0.0.0:13", string(se))
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
			Sequence:      14,
			ServerAddress: net.UDPAddr{IP: net.IPv4zero, Port: 13},

			ClientRoutePublicKey: make([]byte, crypto.KeySize),

			Signature: make([]byte, ed25519.SignatureSize),
		}
		packet.Signature = crypto.Sign(buyersServerPrivKey, packet.GetSignData())

		data, err := packet.MarshalBinary()
		assert.NoError(t, err)

		var resbuf bytes.Buffer

		handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, &db, nil, &iploc, nil, nil, nil)
		handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

		assert.Equal(t, 0, resbuf.Len())
	})

	t.Run("no routes found", func(t *testing.T) {
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

		expected := transport.ServerCacheEntry{
			Sequence: 13,
		}
		se, err := expected.MarshalBinary()
		assert.NoError(t, err)

		err = redisServer.Set("SERVER-0.0.0.0:13", string(se))
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
			Sequence:      14,
			ServerAddress: net.UDPAddr{IP: net.IPv4zero, Port: 13},

			ClientRoutePublicKey: make([]byte, crypto.KeySize),

			Signature: make([]byte, ed25519.SignatureSize),
		}
		packet.Signature = crypto.Sign(buyersServerPrivKey, packet.GetSignData())

		data, err := packet.MarshalBinary()
		assert.NoError(t, err)

		var resbuf bytes.Buffer

		handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, &db, &rp, &iploc, &geoClient, nil, nil)
		handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

		assert.Equal(t, 0, resbuf.Len())
	})

	t.Run("next route response", func(t *testing.T) {
		_, routerPrivKey, err := box.GenerateKey(rand.Reader)
		assert.NoError(t, err)

		clientPubKey, _, err := box.GenerateKey(rand.Reader)
		assert.NoError(t, err)

		serverPubKey, _, err := box.GenerateKey(rand.Reader)
		assert.NoError(t, err)

		relayPubKey, _, err := box.GenerateKey(rand.Reader)
		assert.NoError(t, err)

		buyersServerPubKey, buyersServerPrivKey, err := ed25519.GenerateKey(nil)
		assert.NoError(t, err)

		_, serverBackendPrivKey, err := ed25519.GenerateKey(nil)
		assert.NoError(t, err)

		redisServer, _ := miniredis.Run()
		redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

		db := storage.InMemory{
			LocalBuyer: &routing.Buyer{
				PublicKey: buyersServerPubKey,
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

		rp := mockRouteProvider{
			routes: []routing.Route{
				{
					Relays: []routing.Relay{
						{ID: 1, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 123}, PublicKey: relayPubKey[:]},
						{ID: 2, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.2"), Port: 123}, PublicKey: relayPubKey[:]},
						{ID: 3, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.3"), Port: 123}, PublicKey: relayPubKey[:]},
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
				PublicKey: serverPubKey[:],
			},
		}
		se, err := expected.MarshalBinary()
		assert.NoError(t, err)

		err = redisServer.Set("SERVER-0.0.0.0:13", string(se))
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
			Sequence:      14,
			ServerAddress: net.UDPAddr{IP: net.IPv4zero, Port: 13},

			ClientAddress: net.UDPAddr{
				IP:   net.ParseIP("0.0.0.0"),
				Port: 1234,
			},
			ClientRoutePublicKey: clientPubKey[:],
		}
		packet.Signature = crypto.Sign(buyersServerPrivKey, packet.GetSignData())

		data, err := packet.MarshalBinary()
		assert.NoError(t, err)

		var resbuf bytes.Buffer

		handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, &db, &rp, &iploc, &geoClient, serverBackendPrivKey[:], routerPrivKey[:])
		handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

		var actual transport.SessionResponsePacket
		err = actual.UnmarshalBinary(resbuf.Bytes())
		assert.NoError(t, err)

		assert.Equal(t, packet.SessionId, actual.SessionId)
		assert.Equal(t, packet.Sequence, actual.Sequence)
		assert.Equal(t, int32(routing.RouteTypeNew), actual.RouteType)
		assert.Equal(t, int32(5), actual.NumTokens)
		assert.Equal(t, serverPubKey[:], actual.ServerRoutePublicKey)
	})

	t.Run("continue route response", func(t *testing.T) {
		_, routerPrivKey, err := box.GenerateKey(rand.Reader)
		assert.NoError(t, err)

		clientPubKey, _, err := box.GenerateKey(rand.Reader)
		assert.NoError(t, err)

		serverPubKey, _, err := box.GenerateKey(rand.Reader)
		assert.NoError(t, err)

		relayPubKey, _, err := box.GenerateKey(rand.Reader)
		assert.NoError(t, err)

		buyersServerPubKey, buyersServerPrivKey, err := ed25519.GenerateKey(nil)
		assert.NoError(t, err)

		_, serverBackendPrivKey, err := ed25519.GenerateKey(nil)
		assert.NoError(t, err)

		redisServer, _ := miniredis.Run()
		redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

		db := storage.InMemory{
			LocalBuyer: &routing.Buyer{
				PublicKey: buyersServerPubKey,
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

		rp := mockRouteProvider{
			routes: []routing.Route{
				{
					Relays: []routing.Relay{
						{ID: 1, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 123}, PublicKey: relayPubKey[:]},
						{ID: 2, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.2"), Port: 123}, PublicKey: relayPubKey[:]},
						{ID: 3, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.3"), Port: 123}, PublicKey: relayPubKey[:]},
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
				PublicKey: serverPubKey[:],
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
			ClientRoutePublicKey: clientPubKey[:],
		}
		packet.Signature = crypto.Sign(buyersServerPrivKey, packet.GetSignData())

		data, err := packet.MarshalBinary()
		assert.NoError(t, err)

		var resbuf bytes.Buffer

		handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), redisClient, &db, &rp, &iploc, &geoClient, serverBackendPrivKey[:], routerPrivKey[:])
		handler(&resbuf, &transport.UDPPacket{SourceAddr: addr, Data: data})

		var actual transport.SessionResponsePacket
		err = actual.UnmarshalBinary(resbuf.Bytes())
		assert.NoError(t, err)

		assert.Equal(t, packet.SessionId, actual.SessionId)
		assert.Equal(t, packet.Sequence, actual.Sequence)
		assert.Equal(t, int32(routing.RouteTypeContinue), actual.RouteType)
		assert.Equal(t, int32(5), actual.NumTokens)
		assert.Equal(t, serverPubKey[:], actual.ServerRoutePublicKey)
		assert.True(t, crypto.Verify(buyersServerPubKey, packet.GetSignData(), packet.Signature))
	})
}
