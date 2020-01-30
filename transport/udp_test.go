package transport_test

import (
	"bytes"
	"crypto/ed25519"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v7"
	"github.com/networknext/backend/core"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/nacl/box"
)

type mockBuyerProvider struct {
	buyer *storage.Buyer
	ok    bool
}

func (bp *mockBuyerProvider) GetAndCheckBySdkVersion3PublicKeyId(id uint64) (*storage.Buyer, bool) {
	return bp.buyer, bp.ok
}

type mockRouteProvider struct{}

func (rp *mockRouteProvider) Route(d routing.Datacenter, rs []routing.Relay) ([]routing.Route, error) {
	return []routing.Route{
		{
			Relays: []routing.Relay{
				{ID: 1, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.1"), Port: 123}, PublicKey: make([]byte, crypto.KeySize)},
				{ID: 2, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.2"), Port: 123}, PublicKey: make([]byte, crypto.KeySize)},
				{ID: 3, Addr: net.UDPAddr{IP: net.ParseIP("127.0.0.3"), Port: 123}, PublicKey: make([]byte, crypto.KeySize)},
			},
		},
	}, nil
}

func TestServerUpdateHandlerFunc(t *testing.T) {
	t.Run("failed to unmarshal packet", func(t *testing.T) {
		redisServer, _ := miniredis.Run()
		redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

		addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
		assert.NoError(t, err)

		handler := transport.ServerUpdateHandlerFunc(redisClient, nil)
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

			VersionMajor: 1,
			VersionMinor: 2,
			VersionPatch: 3,

			Signature: make([]byte, ed25519.SignatureSize),
		}

		data, err := packet.MarshalBinary()
		assert.NoError(t, err)

		handler := transport.ServerUpdateHandlerFunc(redisClient, nil)
		handler(&bytes.Buffer{}, &transport.UDPPacket{SourceAddr: addr, Data: data})

		_, err = redisServer.Get("SERVER-0.0.0.0:13")
		assert.Error(t, err)
	})

	t.Run("did not get a buyer", func(t *testing.T) {
		redisServer, _ := miniredis.Run()
		redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

		bp := mockBuyerProvider{
			ok: false,
		}

		addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
		assert.NoError(t, err)

		packet := transport.ServerUpdatePacket{
			Sequence:             13,
			ServerAddress:        net.UDPAddr{IP: net.IPv4zero, Port: 13},
			ServerPrivateAddress: net.UDPAddr{IP: net.IPv4zero, Port: 13},
			ServerRoutePublicKey: TestPublicKey,

			DatacenterId: 13,

			VersionMajor: transport.SDKVersionMin.Major,
			VersionMinor: transport.SDKVersionMin.Minor,
			VersionPatch: transport.SDKVersionMin.Patch,

			Signature: make([]byte, ed25519.SignatureSize),
		}

		data, err := packet.MarshalBinary()
		assert.NoError(t, err)

		handler := transport.ServerUpdateHandlerFunc(redisClient, &bp)
		handler(&bytes.Buffer{}, &transport.UDPPacket{SourceAddr: addr, Data: data})

		_, err = redisServer.Get("SERVER-0.0.0.0:13")
		assert.Error(t, err)
	})

	t.Run("buyer's public key failed verification", func(t *testing.T) {
		_, buyersServerPrivKey, err := ed25519.GenerateKey(nil)
		assert.NoError(t, err)

		redisServer, _ := miniredis.Run()
		redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

		bp := mockBuyerProvider{
			buyer: &storage.Buyer{
				SdkVersion3PublicKeyData: make([]byte, ed25519.PublicKeySize),
			},
			ok: true,
		}

		addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
		assert.NoError(t, err)

		packet := transport.ServerUpdatePacket{
			Sequence:             13,
			ServerAddress:        net.UDPAddr{IP: net.IPv4zero, Port: 13},
			ServerPrivateAddress: net.UDPAddr{IP: net.IPv4zero, Port: 13},
			ServerRoutePublicKey: TestPublicKey,

			DatacenterId: 13,

			VersionMajor: transport.SDKVersionMin.Major,
			VersionMinor: transport.SDKVersionMin.Minor,
			VersionPatch: transport.SDKVersionMin.Patch,

			Signature: make([]byte, ed25519.SignatureSize),
		}
		packet.Signature = ed25519.Sign(buyersServerPrivKey, packet.GetSignData())

		data, err := packet.MarshalBinary()
		assert.NoError(t, err)

		handler := transport.ServerUpdateHandlerFunc(redisClient, &bp)
		handler(&bytes.Buffer{}, &transport.UDPPacket{SourceAddr: addr, Data: data})

		_, err = redisServer.Get("SERVER-0.0.0.0:13")
		assert.Error(t, err)
	})

	t.Run("packet sequence too old", func(t *testing.T) {
		buyersServerPubKey, buyersServerPrivKey, err := ed25519.GenerateKey(nil)
		assert.NoError(t, err)

		redisServer, _ := miniredis.Run()
		redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

		bp := mockBuyerProvider{
			buyer: &storage.Buyer{
				SdkVersion3PublicKeyData: buyersServerPubKey,
			},
			ok: true,
		}

		addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:13")
		assert.NoError(t, err)

		packet := transport.ServerUpdatePacket{
			Sequence:             1,
			ServerAddress:        net.UDPAddr{IP: net.IPv4zero, Port: 13},
			ServerPrivateAddress: net.UDPAddr{IP: net.IPv4zero, Port: 13},
			ServerRoutePublicKey: TestPublicKey,

			DatacenterId: 13,

			VersionMajor: transport.SDKVersionMin.Major,
			VersionMinor: transport.SDKVersionMin.Minor,
			VersionPatch: transport.SDKVersionMin.Patch,
		}
		packet.Signature = ed25519.Sign(buyersServerPrivKey, packet.GetSignData())

		data, err := packet.MarshalBinary()
		assert.NoError(t, err)

		expected := transport.ServerCacheEntry{
			Sequence: 13,
		}
		se, err := expected.MarshalBinary()
		assert.NoError(t, err)

		err = redisServer.Set("SERVER-0.0.0.0:13", string(se))
		assert.NoError(t, err)

		handler := transport.ServerUpdateHandlerFunc(redisClient, &bp)
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

		bp := mockBuyerProvider{
			buyer: &storage.Buyer{
				SdkVersion3PublicKeyData: buyersServerPubKey,
				Active:                   true,
			},
			ok: true,
		}

		// Create a ServerUpdatePacket and marshal it to binary so sent it into the UDP handler
		packet := transport.ServerUpdatePacket{
			Sequence:             13,
			ServerAddress:        net.UDPAddr{IP: net.IPv4zero, Port: 13},
			ServerPrivateAddress: net.UDPAddr{IP: net.IPv4zero, Port: 13},
			ServerRoutePublicKey: TestPublicKey,

			DatacenterId: 13,

			VersionMajor: transport.SDKVersionMin.Major,
			VersionMinor: transport.SDKVersionMin.Minor,
			VersionPatch: transport.SDKVersionMin.Patch,
		}
		packet.Signature = ed25519.Sign(buyersServerPrivKey, packet.GetSignData())

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
		handler := transport.ServerUpdateHandlerFunc(redisClient, &bp)

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
			SDKVersion: transport.SDKVersion{packet.VersionMajor, packet.VersionMinor, packet.VersionPatch},
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
	_, serverBackendPrivKey, err := ed25519.GenerateKey(nil)
	assert.NoError(t, err)

	_, routerServerPrivKey, err := box.GenerateKey(rand.Reader)
	assert.NoError(t, err)

	buyersServerPubKey, buyersServerPrivKey, err := ed25519.GenerateKey(nil)
	assert.NoError(t, err)

	redisServer, _ := miniredis.Run()
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	bp := mockBuyerProvider{
		buyer: &storage.Buyer{
			SdkVersion3PublicKeyData: buyersServerPubKey,
		},
		ok: true,
	}

	rp := mockRouteProvider{}

	// Define a static LocateIPFunc so it will satisfy the IPLocator interface required by the UDP handler
	iploc := routing.LocateIPFunc(func(ip net.IP) (routing.Location, error) {
		return routing.Location{
			Continent: "NA",
			Country:   "US",
			Region:    "NY",
			City:      "Troy",
			Latitude:  43.05036163330078,
			Longitude: -73.75393676757812,
		}, nil
	})

	geoClient := routing.GeoClient{
		RedisClient: redisClient,
		Namespace:   "GEO_TEST",
	}

	addr, err := net.ResolveUDPAddr("udp", "0.0.0.0:12345")
	assert.NoError(t, err)

	// Create a ServerEntry to put into redis for a SessionUpdate to read
	serverentry := transport.ServerCacheEntry{
		Server:     routing.Server{Addr: *addr, PublicKey: TestPublicKey},
		SDKVersion: transport.SDKVersionMin,
	}
	serverdata, err := serverentry.MarshalBinary()
	assert.NoError(t, err)

	// Set the ServerEntry in redis
	err = redisServer.Set("SERVER-0.0.0.0:12345", string(serverdata))
	assert.NoError(t, err)

	// Create an incoming SessionUpdatePacket for the handler
	packet := transport.SessionUpdatePacket{
		Sequence:   13,
		CustomerId: 13,
		SessionId:  13,
		UserHash:   13,
		PlatformId: core.PlatformUnknown,

		ConnectionType: core.ConnectionTypeUnknown,

		Tag:   13,
		Flags: 0,

		Flagged:          true,
		FallbackToDirect: true,
		TryBeforeYouBuy:  true,
		OnNetworkNext:    true,

		DirectMinRtt:     1.0,
		DirectMaxRtt:     2.0,
		DirectMeanRtt:    1.5,
		DirectJitter:     3.0,
		DirectPacketLoss: 4.0,
		NextMinRtt:       1.0,
		NextMaxRtt:       2.0,
		NextMeanRtt:      1.5,
		NextJitter:       3.0,
		NextPacketLoss:   4.0,

		KbpsUp:   10,
		KbpsDown: 20,

		PacketsLostServerToClient: 0,
		PacketsLostClientToServer: 0,

		NumNearRelays:       1,
		NearRelayIds:        []uint64{1},
		NearRelayMinRtt:     []float32{1.0},
		NearRelayMaxRtt:     []float32{2.0},
		NearRelayMeanRtt:    []float32{1.5},
		NearRelayJitter:     []float32{3.0},
		NearRelayPacketLoss: []float32{4.0},

		ServerAddress:        *addr,
		ClientAddress:        *addr,
		ClientRoutePublicKey: TestPublicKey,
	}
	packet.Signature = ed25519.Sign(buyersServerPrivKey, packet.GetSignData(serverentry.SDKVersion))

	data, err := packet.MarshalBinary()
	assert.NoError(t, err)

	var resbuf bytes.Buffer
	incoming := transport.UDPPacket{
		SourceAddr: addr,
		Data:       data,
	}

	// Create and invoke the handler with the packet and from addr
	handler := transport.SessionUpdateHandlerFunc(redisClient, &bp, &rp, iploc, &geoClient, routerServerPrivKey[:], serverBackendPrivKey)
	handler(&resbuf, &incoming)

	{
		// Create the expected SessionEntry
		expected := transport.SessionResponsePacket{
			Sequence:             packet.Sequence,
			SessionId:            packet.SessionId,
			RouteType:            int32(routing.DecisionTypeNew),
			NumTokens:            5,
			Tokens:               make([]byte, 5*routing.EncryptedNextRouteTokenSize),
			Multipath:            false,
			ServerRoutePublicKey: serverentry.Server.PublicKey,
			NumNearRelays:        0,
			NearRelayIds:         make([]uint64, 0),
			NearRelayAddresses:   make([]net.UDPAddr, 0),
		}
		expected.Signature = ed25519.Sign(crypto.BackendPrivateKey, expected.GetSignData())

		data := resbuf.Bytes()
		fmt.Println(hex.Dump(data))
		var actual transport.SessionResponsePacket
		err = actual.UnmarshalBinary(data)
		assert.NoError(t, err)

		assert.Equal(t, expected.Sequence, actual.Sequence)
		assert.Equal(t, expected.SessionId, actual.SessionId)
		assert.Equal(t, expected.RouteType, actual.RouteType)
		assert.Equal(t, expected.NumTokens, actual.NumTokens)
		assert.Equal(t, len(expected.Tokens), len(actual.Tokens))
		assert.Equal(t, expected.Multipath, actual.Multipath)
		assert.Equal(t, expected.ServerRoutePublicKey, actual.ServerRoutePublicKey)
		assert.Equal(t, expected.NumNearRelays, actual.NumNearRelays)
		assert.Equal(t, expected.NearRelayIds, actual.NearRelayIds)
		assert.Equal(t, expected.NearRelayAddresses, actual.NearRelayAddresses)
	}
}
