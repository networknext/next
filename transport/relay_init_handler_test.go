package transport_test

import (
	"bytes"
	crand "crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"math"
	mrand "math/rand"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-kit/kit/log"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v7"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/metrics"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/nacl/box"
)

func getRelayKeyPair(t *testing.T) (pubKey []byte, privKey []byte) {
	key := os.Getenv("RELAY_PUBLIC_KEY")
	assert.NotEqual(t, 0, len(key))
	pubKey, err := base64.StdEncoding.DecodeString(key)
	assert.NoError(t, err)

	key = os.Getenv("RELAY_PRIVATE_KEY")
	assert.NotEqual(t, 0, len(key))
	privKey, err = base64.StdEncoding.DecodeString(key)
	assert.NoError(t, err)

	return pubKey, privKey
}

func relayInitAssertions(t *testing.T, contentType string, relay routing.Relay, body []byte, expectedCode int, geoClient *routing.GeoClient, ipfunc routing.LocateIPFunc, inMemory *storage.InMemory, redisClient *redis.Client, routerPrivateKey []byte) *httptest.ResponseRecorder {
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

	var customerPublicKey []byte
	{
		if key := os.Getenv("NEXT_CUSTOMER_PUBLIC_KEY"); len(key) != 0 {
			customerPublicKey, _ = base64.StdEncoding.DecodeString(key)
		}
	}

	var relayPublicKey []byte
	{
		if key := os.Getenv("RELAY_PUBLIC_KEY"); len(key) != 0 {
			relayPublicKey, _ = base64.StdEncoding.DecodeString(key)
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
	request.Header.Add("Content-Type", contentType)

	if inMemory == nil {
		rtodcnameMap := make(map[uint32]string)
		rtodcnameMap[uint32(relay.ID)] = relay.Datacenter.Name
		rpubkeyMap := make(map[uint32][]byte)
		rpubkeyMap[uint32(relay.ID)] = relay.PublicKey
		inMemory = &storage.InMemory{
			LocalBuyer: &routing.Buyer{
				PublicKey: customerPublicKey[8:],
			},

			LocalRelays: []routing.Relay{
				routing.Relay{
					ID:        crypto.HashID("127.0.0.1:40000"),
					PublicKey: relayPublicKey,
					Latitude:  13,
					Longitude: 13,
				}},
		}
	}

	handler := transport.RelayInitHandlerFunc(log.NewNopLogger(), &transport.RelayInitHandlerConfig{
		RedisClient:      redisClient,
		GeoClient:        geoClient,
		IpLocator:        ipfunc,
		Storer:           inMemory,
		Duration:         &metrics.EmptyGauge{},
		Counter:          &metrics.EmptyCounter{},
		RouterPrivateKey: routerPrivateKey,
	})

	handler(recorder, request)

	assert.Equal(t, expectedCode, recorder.Code)

	return recorder
}

func validateRelayInitSuccess(t *testing.T, expectedContentType string, recorder *httptest.ResponseRecorder, geoClient routing.GeoClient, redisClient *redis.Client, location routing.Location, addr string, before uint64, expected routing.Relay) {
	header := recorder.Header()

	contentType, ok := header["Content-Type"]
	assert.True(t, ok)

	body := recorder.Body.Bytes()

	entry := redisClient.HGet(routing.HashKeyAllRelays, expected.Key())

	var actual routing.Relay
	entryBytes, err := entry.Bytes()
	assert.NoError(t, err)

	err = actual.UnmarshalBinary(entryBytes)
	assert.NoError(t, err)

	var response transport.RelayInitResponse
	switch expectedContentType {
	case "application/octet-stream":
		err = response.UnmarshalBinary(body)
		assert.NoError(t, err)
	case "application/json":
		err = json.Unmarshal(body, &response)
		assert.NoError(t, err)
	default:
		assert.FailNow(t, "Invalid expected content type")
	}

	assert.Equal(t, expectedContentType, contentType[0])
	assert.Equal(t, transport.VersionNumberInitResponse, int(response.Version))
	assert.LessOrEqual(t, before, response.Timestamp)
	assert.GreaterOrEqual(t, uint64(time.Now().Unix()*1000), response.Timestamp)
	assert.Equal(t, actual.PublicKey, response.PublicKey) // entry gets a public key assigned at init which is returned in the response

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

	assert.Equal(t, uint32(routing.RelayStateOnline), actual.State)
}

func TestRelayInitBadPacket(t *testing.T) {
	addr := "127.0.0.1:40000"
	relay := routing.Relay{
		ID: crypto.HashID(addr),
		Datacenter: routing.Datacenter{
			Name: "some datacenter",
		},
	}

	// Binary version
	{
		buff := []byte("bad packet")
		relayInitAssertions(t, "application/octet-stream", relay, buff, http.StatusBadRequest, nil, nil, nil, nil, nil)
	}

	// JSON version
	{
		buff := []byte("{")
		relayInitAssertions(t, "application/json", relay, buff, http.StatusBadRequest, nil, nil, nil, nil, nil)
	}
}

func TestRelayInitMagicIsInvalid(t *testing.T) {
	_, relayPrivateKey := getRelayKeyPair(t)
	routerPublicKey, _, err := box.GenerateKey(crand.Reader)
	assert.NoError(t, err)

	// generate nonce
	nonce := make([]byte, crypto.NonceSize)
	crand.Read(nonce)

	// generate token
	token := make([]byte, crypto.KeySize)
	crand.Read(token)

	// encrypt token
	encryptedToken := crypto.Seal(token, nonce, routerPublicKey[:], relayPrivateKey[:])

	addr := "127.0.0.1:40000"
	relay := routing.Relay{
		ID: crypto.HashID(addr),
		Datacenter: routing.Datacenter{
			Name: "some datacenter",
		},
	}

	packet := transport.RelayInitRequest{
		Magic:          0xFFFFFFFF,
		Nonce:          nonce, // nonce and token just need to be set to pass marshalling
		EncryptedToken: encryptedToken,
	}

	// Binary version
	{
		buff, err := packet.MarshalBinary()
		assert.NoError(t, err)
		relayInitAssertions(t, "application/octet-stream", relay, buff, http.StatusBadRequest, nil, nil, nil, nil, nil)
	}

	// JSON version
	{
		buff, err := packet.MarshalJSON()
		assert.NoError(t, err)
		relayInitAssertions(t, "application/json", relay, buff, http.StatusBadRequest, nil, nil, nil, nil, nil)
	}
}

func TestRelayInitVersionIsInvalid(t *testing.T) {
	_, relayPrivateKey := getRelayKeyPair(t)
	routerPublicKey, _, err := box.GenerateKey(crand.Reader)
	assert.NoError(t, err)

	// generate nonce
	nonce := make([]byte, crypto.NonceSize)
	crand.Read(nonce)

	// generate token
	token := make([]byte, crypto.KeySize)
	crand.Read(token)

	// encrypt token
	encryptedToken := crypto.Seal(token, nonce, routerPublicKey[:], relayPrivateKey[:])

	addr := "127.0.0.1:40000"
	relay := routing.Relay{
		ID: crypto.HashID(addr),
		Datacenter: routing.Datacenter{
			Name: "some datacenter",
		},
	}
	packet := transport.RelayInitRequest{
		Magic:          transport.InitRequestMagic,
		Version:        1,
		Nonce:          nonce, // nonce and token just need to be set to pass marshalling
		EncryptedToken: encryptedToken,
	}

	// Binary version
	{
		buff, err := packet.MarshalBinary()
		assert.NoError(t, err)
		relayInitAssertions(t, "application/octet-stream", relay, buff, http.StatusBadRequest, nil, nil, nil, nil, nil)
	}

	// JSON version
	{
		buff, err := packet.MarshalJSON()
		assert.NoError(t, err)
		relayInitAssertions(t, "application/json", relay, buff, http.StatusBadRequest, nil, nil, nil, nil, nil)
	}
}

// todo: disable test until fixed
/*
func TestRelayInitAddressIsInvalid(t *testing.T) {
	relayPublicKey, relayPrivateKey := getRelayKeyPair(t)
	routerPublicKey, routerPrivateKey, err := box.GenerateKey(crand.Reader)
	assert.NoError(t, err)

	// generate nonce
	nonce := make([]byte, crypto.NonceSize)
	crand.Read(nonce)

	// generate token
	token := make([]byte, crypto.KeySize)
	crand.Read(token)

	// encrypt token
	encryptedToken := crypto.Seal(token, nonce, routerPublicKey[:], relayPrivateKey[:])

	addr := "127.0.0.1:40000"
	udp, _ := net.ResolveUDPAddr("udp", addr)
	relay := routing.Relay{
		ID: crypto.HashID(addr),
		Datacenter: routing.Datacenter{
			Name: "some datacenter",
		},
		PublicKey: relayPublicKey,
	}
	packet := transport.RelayInitRequest{
		Magic:          transport.InitRequestMagic,
		Version:        0,
		Nonce:          nonce,
		Address:        *udp,
		EncryptedToken: encryptedToken,
	}

	// Binary version
	{
		buff, err := packet.MarshalBinary()
		assert.NoError(t, err)
		buff[8+crypto.NonceSize] = 'x' // first number in ip address is now 'x'
		relayInitAssertions(t, "application/octet-stream", relay, buff, http.StatusBadRequest, nil, nil, nil, nil, routerPrivateKey[:])
	}

	// JSON version
	{
		buff, err := packet.MarshalJSON()
		assert.NoError(t, err)

		offset := strings.Index(string(buff), "127.0.0.1:40000")
		assert.GreaterOrEqual(t, offset, 0)
		buff[offset] = 'x' // first number in ip address is now 'x'
		relayInitAssertions(t, "application/json", relay, buff, http.StatusBadRequest, nil, nil, nil, nil, routerPrivateKey[:])
	}
}
*/

func TestRelayInitInvalidToken(t *testing.T) {
	_, routerPrivateKey, err := box.GenerateKey(crand.Reader)
	assert.NoError(t, err)

	// generate nonce
	nonce := make([]byte, crypto.NonceSize)
	crand.Read(nonce)

	// generate token but leave it as 0's
	token := make([]byte, routing.EncryptedRelayTokenSize)

	addr := "127.0.0.1:40000"
	udp, _ := net.ResolveUDPAddr("udp", addr)
	relay := routing.Relay{
		ID: crypto.HashID(addr),
		Datacenter: routing.Datacenter{
			Name: "some datacenter",
		},
	}
	packet := transport.RelayInitRequest{
		Magic:          transport.InitRequestMagic,
		Version:        0,
		Nonce:          nonce,
		Address:        *udp,
		EncryptedToken: token,
	}

	// Binary version
	{
		buff, err := packet.MarshalBinary()
		assert.NoError(t, err)
		relayInitAssertions(t, "application/octet-stream", relay, buff, http.StatusUnauthorized, nil, nil, nil, nil, routerPrivateKey[:])
	}

	// JSON version
	{
		buff, err := packet.MarshalJSON()
		assert.NoError(t, err)
		relayInitAssertions(t, "application/json", relay, buff, http.StatusUnauthorized, nil, nil, nil, nil, routerPrivateKey[:])
	}
}

func TestRelayInitInvalidNonce(t *testing.T) {
	relayPublicKey, relayPrivateKey := getRelayKeyPair(t)
	routerPublicKey, routerPrivateKey, err := box.GenerateKey(crand.Reader)
	assert.NoError(t, err)

	// generate nonce but leave it as 0's
	nonce := make([]byte, crypto.NonceSize)

	// generate random token
	token := make([]byte, crypto.KeySize)
	crand.Read(token)

	// seal it with the bad nonce
	encryptedToken := crypto.Seal(token, nonce, routerPublicKey[:], relayPrivateKey[:])

	addr := "127.0.0.1:40000"
	udp, _ := net.ResolveUDPAddr("udp", addr)
	relay := routing.Relay{
		ID: crypto.HashID(addr),
		Datacenter: routing.Datacenter{
			Name: "some datacenter",
		},
		PublicKey: relayPublicKey,
	}
	packet := transport.RelayInitRequest{
		Magic:          transport.InitRequestMagic,
		Version:        0,
		Nonce:          nonce,
		Address:        *udp,
		EncryptedToken: encryptedToken,
	}

	// Binary version
	{
		buff, err := packet.MarshalBinary()
		assert.NoError(t, err)
		relayInitAssertions(t, "application/octet-stream", relay, buff, http.StatusOK, nil, nil, nil, nil, routerPrivateKey[:])
	}

	// JSON version
	{
		buff, err := packet.MarshalJSON()
		assert.NoError(t, err)
		relayInitAssertions(t, "application/json", relay, buff, http.StatusOK, nil, nil, nil, nil, routerPrivateKey[:])
	}

}

func TestRelayInitRelayExists(t *testing.T) {
	redisServer, _ := miniredis.Run()
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	relayPublicKey, relayPrivateKey := getRelayKeyPair(t)
	routerPublicKey, routerPrivateKey, err := box.GenerateKey(crand.Reader)
	assert.NoError(t, err)

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

	entry := routing.Relay{
		ID:   crypto.HashID(addr),
		Name: name,
		Addr: *udpAddr,
		Datacenter: routing.Datacenter{
			ID:   32,
			Name: dcname,
		},
		PublicKey:      token,
		LastUpdateTime: 1234,
	}

	relay := routing.Relay{
		ID: crypto.HashID(addr),
		Datacenter: routing.Datacenter{
			Name: "some datacenter",
		},
		PublicKey: relayPublicKey,
	}

	packet := transport.RelayInitRequest{
		Magic:          transport.InitRequestMagic,
		Version:        0,
		Nonce:          nonce,
		Address:        *udpAddr,
		EncryptedToken: encryptedToken,
	}

	// get the binary data from the entry
	data, err := entry.MarshalBinary()
	assert.NoError(t, err)

	// set it in the redis instance
	redisServer.HSet(routing.HashKeyAllRelays, entry.Key(), string(data))

	// Binary version
	{
		buff, err := packet.MarshalBinary()
		assert.NoError(t, err)
		relayInitAssertions(t, "application/octet-stream", relay, buff, http.StatusConflict, nil, nil, nil, redisClient, routerPrivateKey[:])
	}

	// JSON version
	{
		buff, err := packet.MarshalJSON()
		assert.NoError(t, err)
		relayInitAssertions(t, "application/json", relay, buff, http.StatusConflict, nil, nil, nil, redisClient, routerPrivateKey[:])
	}
}

func TestRelayInitRelayIPLookupFailure(t *testing.T) {
	redisServer, _ := miniredis.Run()
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	relayPublicKey, relayPrivateKey := getRelayKeyPair(t)
	routerPublicKey, routerPrivateKey, err := box.GenerateKey(crand.Reader)
	assert.NoError(t, err)

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

	relay := routing.Relay{
		ID: crypto.HashID(addr),
		Datacenter: routing.Datacenter{
			Name: "some datacenter",
		},
		PublicKey: relayPublicKey,
		Latitude:  13,
		Longitude: 13,
	}

	packet := transport.RelayInitRequest{
		Magic:          transport.InitRequestMagic,
		Nonce:          nonce,
		Address:        *udpAddr,
		EncryptedToken: encryptedToken,
	}

	// Binary version
	{
		buff, err := packet.MarshalBinary()
		assert.NoError(t, err)
		relayInitAssertions(t, "application/octet-stream", relay, buff, http.StatusOK, nil, ipfunc, nil, redisClient, routerPrivateKey[:])
	}

	// clear redis
	redisServer, _ = miniredis.Run()
	redisClient = redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	// JSON version
	{
		buff, err := packet.MarshalJSON()
		assert.NoError(t, err)
		relayInitAssertions(t, "application/json", relay, buff, http.StatusOK, nil, ipfunc, nil, redisClient, routerPrivateKey[:])
	}
}

func TestRelayInitRelayDBLookupFailure(t *testing.T) {
	redisServer, _ := miniredis.Run()
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	relayPublicKey, relayPrivateKey := getRelayKeyPair(t)
	routerPublicKey, routerPrivateKey, err := box.GenerateKey(crand.Reader)
	assert.NoError(t, err)

	nonce := make([]byte, crypto.NonceSize)
	crand.Read(nonce)

	addr := "127.0.0.1:40000"
	udpAddr, _ := net.ResolveUDPAddr("udp", addr)

	token := make([]byte, crypto.KeySize)
	crand.Read(token)

	encryptedToken := crypto.Seal(token, nonce, routerPublicKey[:], relayPrivateKey[:])

	inMemory := &storage.InMemory{} // Have empty storage to fail lookup

	relay := routing.Relay{
		ID: crypto.HashID(addr),
		Datacenter: routing.Datacenter{
			Name: "some datacenter",
		},
		PublicKey: relayPublicKey,
	}

	packet := transport.RelayInitRequest{
		Magic:          transport.InitRequestMagic,
		Nonce:          nonce,
		Address:        *udpAddr,
		EncryptedToken: encryptedToken,
	}

	// Binary version
	{
		buff, err := packet.MarshalBinary()
		assert.NoError(t, err)
		relayInitAssertions(t, "application/octet-stream", relay, buff, http.StatusInternalServerError, nil, nil, inMemory, redisClient, routerPrivateKey[:])
	}

	// clear redis
	redisServer, _ = miniredis.Run()
	redisClient = redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	// JSON version
	{
		buff, err := packet.MarshalJSON()
		assert.NoError(t, err)
		relayInitAssertions(t, "application/json", relay, buff, http.StatusInternalServerError, nil, nil, inMemory, redisClient, routerPrivateKey[:])
	}
}

func TestRelayInitRelayRedisFailure(t *testing.T) {
	// Don't establish a redis server to simulate the client being unable to find the relay
	redisClient := redis.NewClient(&redis.Options{Addr: "0.0.0.0"})

	relayPublicKey, relayPrivateKey := getRelayKeyPair(t)
	routerPublicKey, routerPrivateKey, err := box.GenerateKey(crand.Reader)
	assert.NoError(t, err)

	nonce := make([]byte, crypto.NonceSize)
	crand.Read(nonce)

	addr := "127.0.0.1:40000"
	udpAddr, _ := net.ResolveUDPAddr("udp", addr)

	token := make([]byte, crypto.KeySize)
	crand.Read(token)

	encryptedToken := crypto.Seal(token, nonce, routerPublicKey[:], relayPrivateKey[:])

	relay := routing.Relay{
		ID: crypto.HashID(addr),
		Datacenter: routing.Datacenter{
			Name: "some datacenter",
		},
		PublicKey: relayPublicKey,
	}

	packet := transport.RelayInitRequest{
		Magic:          transport.InitRequestMagic,
		Nonce:          nonce,
		Address:        *udpAddr,
		EncryptedToken: encryptedToken,
	}

	// Binary version
	{
		buff, err := packet.MarshalBinary()
		assert.NoError(t, err)
		relayInitAssertions(t, "application/octet-stream", relay, buff, http.StatusNotFound, nil, nil, nil, redisClient, routerPrivateKey[:])
	}

	// JSON version
	{
		buff, err := packet.MarshalJSON()
		assert.NoError(t, err)
		relayInitAssertions(t, "application/json", relay, buff, http.StatusNotFound, nil, nil, nil, redisClient, routerPrivateKey[:])
	}

}

func TestRelayInitSuccess(t *testing.T) {
	redisServer, _ := miniredis.Run()
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	relayPublicKey, relayPrivateKey := getRelayKeyPair(t)
	routerPublicKey, routerPrivateKey, err := box.GenerateKey(crand.Reader)
	assert.NoError(t, err)

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

	relay := routing.Relay{
		ID: crypto.HashID(addr),
		Datacenter: routing.Datacenter{
			Name: "some datacenter",
		},
		PublicKey: relayPublicKey,
	}

	packet := transport.RelayInitRequest{
		Magic:          transport.InitRequestMagic,
		Nonce:          nonce,
		Address:        *udpAddr,
		EncryptedToken: encryptedToken,
	}

	expected := routing.Relay{
		ID:   crypto.HashID(addr),
		Addr: *udpAddr,
	}

	var recorder *httptest.ResponseRecorder
	// Binary version
	{
		buff, err := packet.MarshalBinary()
		assert.NoError(t, err)
		recorder = relayInitAssertions(t, "application/octet-stream", relay, buff, http.StatusOK, &geoClient, ipfunc, nil, redisClient, routerPrivateKey[:])
		header := recorder.Header()

		contentType, ok := header["Content-Type"]
		assert.True(t, ok)

		assert.Equal(t, "application/octet-stream", contentType[0])
		validateRelayInitSuccess(t, "application/octet-stream", recorder, geoClient, redisClient, location, addr, before, expected)
	}

	// clear redis
	redisServer, _ = miniredis.Run()
	redisClient = redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	// JSON version
	{
		buff, err := packet.MarshalJSON()
		assert.NoError(t, err)
		recorder = relayInitAssertions(t, "application/json", relay, buff, http.StatusOK, &geoClient, ipfunc, nil, redisClient, routerPrivateKey[:])

		validateRelayInitSuccess(t, "application/json", recorder, geoClient, redisClient, location, addr, before*1000, expected)
	}
}
