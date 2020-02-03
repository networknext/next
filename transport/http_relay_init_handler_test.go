package transport_test

import (
	"bytes"
	crand "crypto/rand"
	"errors"
	"log"
	"math"
	mrand "math/rand"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v7"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/encoding"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/nacl/box"
)

func relayInitAssertions(t *testing.T, relay routing.Relay, body []byte, expectedCode int, geoClient *routing.GeoClient, ipfunc routing.LocateIPFunc, inMemory *storage.InMemory, redisClient *redis.Client, relayPublicKey []byte, routerPrivateKey []byte) *httptest.ResponseRecorder {
	if redisClient == nil {
		redisServer, _ := miniredis.Run()
		redisClient = redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	}

	if geoClient == nil {
		serv, _ := miniredis.Run()
		cli := redis.NewClient(&redis.Options{Addr: serv.Addr()})
		geoClient = &routing.GeoClient{
			RedisClient: cli,
			Namespace:   "RELAY_LOCATIONS",
		}
	}

	if ipfunc == nil {
		ipfunc = func(ip net.IP) (routing.Location, error) {
			return routing.Location{
				Continent: "a continent on the Earth",
				Country:   "a country in the continent",
				Region:    "a region in the country",
				City:      "a city in the region",
				Latitude:  mrand.Float64(),
				Longitude: mrand.Float64(),
			}, nil
		}
	}

	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("POST", "/relay_init", bytes.NewBuffer(body))

	if inMemory == nil {
		rtodcnameMap := make(map[uint32]string)
		rtodcnameMap[uint32(relay.ID)] = relay.DatacenterName
		inMemory = &storage.InMemory{
			RelayDatacenterNames: rtodcnameMap,
		}
	}

	handler := transport.RelayInitHandlerFunc(redisClient, geoClient, ipfunc, inMemory, inMemory, relayPublicKey, routerPrivateKey)

	handler(recorder, request)

	assert.Equal(t, expectedCode, recorder.Code)

	return recorder
}

func TestRelayInitHandler(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	const addr = "127.0.0.1:40000"
	t.Run("magic is invalid", func(t *testing.T) {
		udp, _ := net.ResolveUDPAddr("udp", addr)
		packet := transport.RelayInitPacket{
			Magic:          0xFFFFFFFF,
			Version:        0,
			Address:        *udp,
			Nonce:          make([]byte, crypto.NonceSize),
			EncryptedToken: make([]byte, routing.EncryptedTokenSize),
		}
		buff, _ := packet.MarshalBinary()
		relay := routing.Relay{
			ID:             crypto.HashID(addr),
			DatacenterName: "some datacenter",
		}
		relayInitAssertions(t, relay, buff, http.StatusBadRequest, nil, nil, nil, nil, nil, nil)
	})

	t.Run("version is invalid", func(t *testing.T) {
		udp, _ := net.ResolveUDPAddr("udp", addr)
		packet := transport.RelayInitPacket{
			Magic:          transport.InitRequestMagic,
			Version:        1,
			Address:        *udp,
			Nonce:          make([]byte, crypto.NonceSize),
			EncryptedToken: make([]byte, routing.EncryptedTokenSize),
		}
		buff, _ := packet.MarshalBinary()
		relay := routing.Relay{
			ID:             crypto.HashID(addr),
			DatacenterName: "some datacenter",
		}
		relayInitAssertions(t, relay, buff, http.StatusBadRequest, nil, nil, nil, nil, nil, nil)
	})

	t.Run("address is invalid", func(t *testing.T) {
		// generate keys
		relayPublicKey, relayPrivateKey, _ := box.GenerateKey(crand.Reader)
		routerPublicKey, routerPrivateKey, _ := box.GenerateKey(crand.Reader)

		// generate nonce
		nonce := make([]byte, crypto.NonceSize)
		crand.Read(nonce)

		// generate token
		token := make([]byte, crypto.KeySize)
		crand.Read(token)

		// encrypt token
		encryptedToken := crypto.Seal(token, nonce, routerPublicKey[:], relayPrivateKey[:])

		udp, _ := net.ResolveUDPAddr("udp", addr)
		packet := transport.RelayInitPacket{
			Magic:          transport.InitRequestMagic,
			Version:        0,
			Nonce:          nonce,
			Address:        *udp,
			EncryptedToken: encryptedToken,
		}
		buff, _ := packet.MarshalBinary()
		buff[8+crypto.NonceSize] = 'x' // first number in ip address is now 'x'
		relay := routing.Relay{
			ID:             crypto.HashID(addr),
			DatacenterName: "some datacenter",
		}
		relayInitAssertions(t, relay, buff, http.StatusBadRequest, nil, nil, nil, nil, relayPublicKey[:], routerPrivateKey[:])
	})

	t.Run("encryption token is 0'ed", func(t *testing.T) {
		// generate keys
		relayPublicKey, _, _ := box.GenerateKey(crand.Reader)
		_, routerPrivateKey, _ := box.GenerateKey(crand.Reader)

		// generate nonce
		nonce := make([]byte, crypto.NonceSize)
		crand.Read(nonce)

		// generate token but leave it as 0's
		token := make([]byte, routing.EncryptedTokenSize)

		udp, _ := net.ResolveUDPAddr("udp", addr)
		packet := transport.RelayInitPacket{
			Magic:          transport.InitRequestMagic,
			Version:        0,
			Nonce:          nonce,
			Address:        *udp,
			EncryptedToken: token,
		}
		buff, _ := packet.MarshalBinary()
		relay := routing.Relay{
			ID:             crypto.HashID(addr),
			DatacenterName: "some datacenter",
		}
		relayInitAssertions(t, relay, buff, http.StatusUnauthorized, nil, nil, nil, nil, relayPublicKey[:], routerPrivateKey[:])
	})

	t.Run("nonce bytes are 0'ed", func(t *testing.T) {
		// generate keys
		relayPublicKey, relayPrivateKey, _ := box.GenerateKey(crand.Reader)
		routerPublicKey, routerPrivateKey, _ := box.GenerateKey(crand.Reader)

		// generate nonce but leave it as 0's
		nonce := make([]byte, crypto.NonceSize)

		// generate random token
		token := make([]byte, crypto.KeySize)
		crand.Read(token)

		// seal it with the bad nonce
		encryptedToken := crypto.Seal(token, nonce, routerPublicKey[:], relayPrivateKey[:])

		udp, _ := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
		packet := transport.RelayInitPacket{
			Magic:          transport.InitRequestMagic,
			Version:        0,
			Nonce:          nonce,
			Address:        *udp,
			EncryptedToken: encryptedToken,
		}

		buff, _ := packet.MarshalBinary()
		relay := routing.Relay{
			ID:             crypto.HashID(addr),
			DatacenterName: "some datacenter",
		}
		relayInitAssertions(t, relay, buff, http.StatusOK, nil, nil, nil, nil, relayPublicKey[:], routerPrivateKey[:])
	})

	t.Run("relay already exists", func(t *testing.T) {
		redisServer, _ := miniredis.Run()
		redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

		// generate keys
		relayPublicKey, relayPrivateKey, _ := box.GenerateKey(crand.Reader)
		routerPublicKey, routerPrivateKey, _ := box.GenerateKey(crand.Reader)

		// generate nonce
		nonce := make([]byte, crypto.NonceSize)
		crand.Read(nonce)

		// generate token
		token := make([]byte, crypto.KeySize)
		crand.Read(token)

		// encrypt token
		encryptedToken := crypto.Seal(token, nonce, routerPublicKey[:], relayPrivateKey[:])

		name := "some name"
		addr := "127.0.0.1:40000"
		udpAddr, _ := net.ResolveUDPAddr("udp", addr)
		dcname := "another name"

		packet := transport.RelayInitPacket{
			Magic:          transport.InitRequestMagic,
			Version:        0,
			Nonce:          nonce,
			Address:        *udpAddr,
			EncryptedToken: encryptedToken,
		}

		buff, _ := packet.MarshalBinary()

		entry := routing.Relay{
			ID:             crypto.HashID(addr),
			Name:           name,
			Addr:           *udpAddr,
			Datacenter:     32,
			DatacenterName: dcname,
			PublicKey:      token,
			LastUpdateTime: 1234,
		}

		// get the binary data from the entry
		data, _ := entry.MarshalBinary()

		// set it in the redis instance
		redisServer.HSet(routing.HashKeyAllRelays, entry.Key(), string(data))
		relay := routing.Relay{
			ID:             crypto.HashID(addr),
			DatacenterName: "some datacenter",
		}
		relayInitAssertions(t, relay, buff, http.StatusNotFound, nil, nil, nil, redisClient, relayPublicKey[:], routerPrivateKey[:])
	})

	t.Run("could not lookup relay location", func(t *testing.T) {
		redisServer, _ := miniredis.Run()
		redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

		relayPublicKey, relayPrivateKey, _ := box.GenerateKey(crand.Reader)
		routerPublicKey, routerPrivateKey, _ := box.GenerateKey(crand.Reader)

		ipfunc := func(ip net.IP) (routing.Location, error) {
			return routing.Location{}, errors.New("descriptive error")
		}

		nonce := make([]byte, crypto.NonceSize)
		crand.Read(nonce)

		addr := "127.0.0.1:40000"
		udpAddr, _ := net.ResolveUDPAddr("udp", addr)

		token := make([]byte, crypto.KeySize)
		crand.Read(token)

		encryptedToken := crypto.Seal(token, nonce, routerPublicKey[:], relayPrivateKey[:])

		packet := transport.RelayInitPacket{
			Magic:          transport.InitRequestMagic,
			Nonce:          nonce,
			Address:        *udpAddr,
			EncryptedToken: encryptedToken,
		}
		buff, _ := packet.MarshalBinary()

		relay := routing.Relay{
			ID:             crypto.HashID(addr),
			DatacenterName: "some datacenter",
		}
		relayInitAssertions(t, relay, buff, http.StatusInternalServerError, nil, ipfunc, nil, redisClient, relayPublicKey[:], routerPrivateKey[:])
	})

	t.Run("failed to get relay from configstore", func(t *testing.T) {
		redisServer, _ := miniredis.Run()
		redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

		relayPublicKey, relayPrivateKey, _ := box.GenerateKey(crand.Reader)
		routerPublicKey, routerPrivateKey, _ := box.GenerateKey(crand.Reader)

		nonce := make([]byte, crypto.NonceSize)
		crand.Read(nonce)

		addr := "127.0.0.1:40000"
		udpAddr, _ := net.ResolveUDPAddr("udp", addr)

		token := make([]byte, crypto.KeySize)
		crand.Read(token)

		encryptedToken := crypto.Seal(token, nonce, routerPublicKey[:], relayPrivateKey[:])

		packet := transport.RelayInitPacket{
			Magic:          transport.InitRequestMagic,
			Nonce:          nonce,
			Address:        *udpAddr,
			EncryptedToken: encryptedToken,
		}
		buff, _ := packet.MarshalBinary()

		relay := routing.Relay{
			ID:             crypto.HashID(addr),
			DatacenterName: "some datacenter",
		}

		inMemory := &storage.InMemory{} // Have empty storage to fail lookup

		log.Println(inMemory)
		relayInitAssertions(t, relay, buff, http.StatusInternalServerError, nil, nil, inMemory, redisClient, relayPublicKey[:], routerPrivateKey[:])
	})

	t.Run("Failed to get relay from redis", func(t *testing.T) {
		// Don't establish a redis server to simulate the client being unable to find the relay
		redisClient := redis.NewClient(&redis.Options{Addr: "0.0.0.0"})

		relayPublicKey, relayPrivateKey, _ := box.GenerateKey(crand.Reader)
		routerPublicKey, routerPrivateKey, _ := box.GenerateKey(crand.Reader)

		nonce := make([]byte, crypto.NonceSize)
		crand.Read(nonce)

		addr := "127.0.0.1:40000"
		udpAddr, _ := net.ResolveUDPAddr("udp", addr)

		token := make([]byte, crypto.KeySize)
		crand.Read(token)

		encryptedToken := crypto.Seal(token, nonce, routerPublicKey[:], relayPrivateKey[:])

		packet := transport.RelayInitPacket{
			Magic:          transport.InitRequestMagic,
			Nonce:          nonce,
			Address:        *udpAddr,
			EncryptedToken: encryptedToken,
		}
		buff, _ := packet.MarshalBinary()

		relay := routing.Relay{
			ID:             crypto.HashID(addr),
			DatacenterName: "some datacenter",
		}
		relayInitAssertions(t, relay, buff, http.StatusNotFound, nil, nil, nil, redisClient, relayPublicKey[:], routerPrivateKey[:])
	})

	t.Run("valid", func(t *testing.T) {
		redisServer, _ := miniredis.Run()
		redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

		relayPublicKey, relayPrivateKey, _ := box.GenerateKey(crand.Reader)
		routerPublicKey, routerPrivateKey, _ := box.GenerateKey(crand.Reader)

		var geoClient routing.GeoClient
		{
			redisServer, _ := miniredis.Run()
			redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
			geoClient = routing.GeoClient{
				RedisClient: redisClient,
				Namespace:   "RELAY_LOCATIONS",
			}
		}

		location := routing.Location{
			Latitude:  math.Round(mrand.Float64()*1000) / 1000,
			Longitude: math.Round(mrand.Float64()*1000) / 1000,
		}

		ipfunc := func(ip net.IP) (routing.Location, error) {
			return location, nil
		}

		nonce := make([]byte, crypto.NonceSize)
		crand.Read(nonce)

		addr := "127.0.0.1:40000"
		udpAddr, _ := net.ResolveUDPAddr("udp", addr)

		token := make([]byte, crypto.KeySize)
		crand.Read(token)

		encryptedToken := crypto.Seal(token, nonce, routerPublicKey[:], relayPrivateKey[:])

		before := uint64(time.Now().Unix())

		packet := transport.RelayInitPacket{
			Magic:          transport.InitRequestMagic,
			Nonce:          nonce,
			Address:        *udpAddr,
			EncryptedToken: encryptedToken,
		}
		buff, _ := packet.MarshalBinary()

		relay := routing.Relay{
			ID:             crypto.HashID(addr),
			DatacenterName: "some datacenter",
		}
		recorder := relayInitAssertions(t, relay, buff, http.StatusOK, &geoClient, ipfunc, nil, redisClient, relayPublicKey[:], routerPrivateKey[:])

		header := recorder.Header()
		contentType, _ := header["Content-Type"]
		expected := routing.Relay{
			ID:   crypto.HashID(addr),
			Addr: *udpAddr,
		}

		resp := redisClient.HGet(routing.HashKeyAllRelays, expected.Key())

		var actual routing.Relay
		bin, _ := resp.Bytes()
		assert.Nil(t, actual.UnmarshalBinary(bin))

		indx := 0
		body := recorder.Body.Bytes()

		var version uint32
		encoding.ReadUint32(body, &indx, &version)

		var timestamp uint64
		encoding.ReadUint64(body, &indx, &timestamp)

		var publicKey []byte
		encoding.ReadBytes(body, &indx, &publicKey, crypto.KeySize)

		assert.Equal(t, "application/octet-stream", contentType[0])
		assert.Equal(t, transport.VersionNumberInitResponse, int(version))
		assert.LessOrEqual(t, before, timestamp)
		assert.GreaterOrEqual(t, uint64(time.Now().Unix()), timestamp)
		assert.Equal(t, actual.PublicKey, publicKey) // entry gets a public key assigned at init which is returned in the response

		assert.Equal(t, expected.ID, actual.ID)
		assert.Equal(t, expected.Name, actual.Name)
		assert.Equal(t, expected.Addr, actual.Addr)
		assert.NotZero(t, actual.LastUpdateTime)
		assert.Len(t, actual.PublicKey, 32)

		// only added one relay so it should be the only one returned by this
		relaysInLocation, _ := geoClient.RelaysWithin(location.Latitude, location.Longitude, 1, "km")
		if assert.Len(t, relaysInLocation, 1) {
			relay := relaysInLocation[0]

			assert.Equal(t, crypto.HashID(addr), relay.ID)
			assert.Equal(t, location.Latitude, math.Round(relay.Latitude*1000)/1000)
			assert.Equal(t, location.Longitude, math.Round(relay.Longitude*1000)/1000)
		}
	})
}
