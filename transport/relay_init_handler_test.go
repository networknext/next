package transport_test

import (
	"bytes"
	"context"
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
	"reflect"
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

func pingRelayBackendInit(t *testing.T, contentType string, relay routing.Relay, body []byte, initMetrics metrics.RelayInitMetrics, geoClient *routing.GeoClient, ipfunc routing.LocateIPFunc, inMemory *storage.InMemory, redisClient *redis.Client, routerPrivateKey []byte) *httptest.ResponseRecorder {
	if redisClient == nil {
		redisServer, err := miniredis.Run()
		assert.NoError(t, err)
		redisClient = redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	}

	if geoClient == nil {
		serv, err := miniredis.Run()
		assert.NoError(t, err)
		cli := redis.NewClient(&redis.Options{Addr: serv.Addr()})
		geoClient = &routing.GeoClient{
			RedisClient: cli,
			Namespace:   "RELAY_LOCATIONS",
		}
	}

	var customerPublicKey []byte
	{
		if key := os.Getenv("NEXT_CUSTOMER_PUBLIC_KEY"); len(key) != 0 {
			var err error
			customerPublicKey, err = base64.StdEncoding.DecodeString(key)
			assert.NoError(t, err)
		}
	}

	var relayPublicKey []byte
	{
		if key := os.Getenv("RELAY_PUBLIC_KEY"); len(key) != 0 {
			var err error
			relayPublicKey, err = base64.StdEncoding.DecodeString(key)
			assert.NoError(t, err)
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
	request, err := http.NewRequest("POST", "/relay_init", bytes.NewBuffer(body))
	assert.NoError(t, err)
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
		Metrics:          &initMetrics,
		RouterPrivateKey: routerPrivateKey,
	})

	handler(recorder, request)
	return recorder
}

func relayInitErrorAssertions(t *testing.T, recorder *httptest.ResponseRecorder, expectedCode int, errMetric metrics.Counter) {
	assert.Equal(t, expectedCode, recorder.Code)
	assert.Equal(t, 1.0, errMetric.ValueReset())
}

func relayInitSuccessAssertions(t *testing.T, recorder *httptest.ResponseRecorder, expectedContentType string, errMetrics metrics.RelayInitErrorMetrics, geoClient *routing.GeoClient, redisClient *redis.Client, location routing.Location, addr string, before uint64, expected routing.Relay) {
	assert.Equal(t, http.StatusOK, recorder.Code)

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
	relaysInLocation, err := geoClient.RelaysWithin(location.Latitude, location.Longitude, 1, "km")
	assert.NoError(t, err)
	if assert.Len(t, relaysInLocation, 1) {
		relay := relaysInLocation[0]

		assert.Equal(t, crypto.HashID(addr), relay.ID)
		assert.Equal(t, location.Latitude, math.Round(relay.Latitude*1000)/1000)
		assert.Equal(t, location.Longitude, math.Round(relay.Longitude*1000)/1000)
	}

	assert.Equal(t, uint32(routing.RelayStateOnline), actual.State)

	errMetricsStruct := reflect.ValueOf(errMetrics)
	for i := 0; i < errMetricsStruct.NumField(); i++ {
		if errMetricsStruct.Field(i).CanInterface() {
			assert.Equal(t, 0.0, errMetricsStruct.Field(i).Interface().(metrics.Counter).ValueReset())
		}
	}
}

func TestRelayInitUnmarshalFailure(t *testing.T) {
	addr := "127.0.0.1:40000"
	relay := routing.Relay{
		ID: crypto.HashID(addr),
		Datacenter: routing.Datacenter{
			Name: "some datacenter",
		},
	}

	initMetrics := metrics.EmptyRelayInitMetrics
	localMetrics := metrics.LocalHandler{}

	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
	assert.NoError(t, err)

	initMetrics.ErrorMetrics.UnmarshalFailure = metric

	// Binary version
	{
		buff := []byte("bad packet")
		recorder := pingRelayBackendInit(t, "application/octet-stream", relay, buff, initMetrics, nil, nil, nil, nil, nil)
		relayInitErrorAssertions(t, recorder, http.StatusBadRequest, metric)
	}

	// JSON version
	{
		buff := []byte("{")
		recorder := pingRelayBackendInit(t, "application/json", relay, buff, initMetrics, nil, nil, nil, nil, nil)
		relayInitErrorAssertions(t, recorder, http.StatusBadRequest, metric)
	}
}

func TestRelayInitInvalidMagic(t *testing.T) {
	addr := "127.0.0.1:40000"
	relay := routing.Relay{
		ID: crypto.HashID(addr),
		Datacenter: routing.Datacenter{
			Name: "some datacenter",
		},
	}

	packet := transport.RelayInitRequest{
		Magic:          0xFFFFFFFF,
		Nonce:          make([]byte, crypto.NonceSize),
		EncryptedToken: make([]byte, routing.EncryptedRelayTokenSize),
	}

	initMetrics := metrics.EmptyRelayInitMetrics
	localMetrics := metrics.LocalHandler{}

	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
	assert.NoError(t, err)

	initMetrics.ErrorMetrics.InvalidMagic = metric

	// Binary version
	{
		buff, err := packet.MarshalBinary()
		assert.NoError(t, err)
		recorder := pingRelayBackendInit(t, "application/octet-stream", relay, buff, initMetrics, nil, nil, nil, nil, nil)
		relayInitErrorAssertions(t, recorder, http.StatusBadRequest, metric)
	}

	// JSON version
	{
		buff, err := packet.MarshalJSON()
		assert.NoError(t, err)
		recorder := pingRelayBackendInit(t, "application/json", relay, buff, initMetrics, nil, nil, nil, nil, nil)
		relayInitErrorAssertions(t, recorder, http.StatusBadRequest, metric)
	}
}

func TestRelayInitInvalidVersion(t *testing.T) {
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
		Nonce:          make([]byte, crypto.NonceSize),
		EncryptedToken: make([]byte, routing.EncryptedRelayTokenSize),
	}

	initMetrics := metrics.EmptyRelayInitMetrics
	localMetrics := metrics.LocalHandler{}

	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
	assert.NoError(t, err)

	initMetrics.ErrorMetrics.InvalidVersion = metric

	// Binary version
	{
		buff, err := packet.MarshalBinary()
		assert.NoError(t, err)
		recorder := pingRelayBackendInit(t, "application/octet-stream", relay, buff, initMetrics, nil, nil, nil, nil, nil)
		relayInitErrorAssertions(t, recorder, http.StatusBadRequest, metric)
	}

	// JSON version
	{
		buff, err := packet.MarshalJSON()
		assert.NoError(t, err)
		recorder := pingRelayBackendInit(t, "application/json", relay, buff, initMetrics, nil, nil, nil, nil, nil)
		relayInitErrorAssertions(t, recorder, http.StatusBadRequest, metric)
	}
}

func TestRelayInitInvalidAddress(t *testing.T) {
	relayPublicKey, _ := getRelayKeyPair(t)
	_, routerPrivateKey, err := box.GenerateKey(crand.Reader)
	assert.NoError(t, err)

	addr := "127.0.0.1:40000"
	udp, err := net.ResolveUDPAddr("udp", addr)
	assert.NoError(t, err)
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
		Nonce:          make([]byte, crypto.NonceSize),
		Address:        *udp,
		EncryptedToken: make([]byte, routing.EncryptedRelayTokenSize),
	}

	initMetrics := metrics.EmptyRelayInitMetrics
	localMetrics := metrics.LocalHandler{}

	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
	assert.NoError(t, err)

	initMetrics.ErrorMetrics.UnmarshalFailure = metric

	// Binary version
	{
		buff, err := packet.MarshalBinary()
		assert.NoError(t, err)
		badAddr := "invalid address"        // "invalid address" is luckily the same number of characters as "127.0.0.1:40000"
		for i := 0; i < len(badAddr); i++ { // Replace the address with the bad address character by character
			buff[4+4+crypto.NonceSize+4+i] = badAddr[i]
		}
		recorder := pingRelayBackendInit(t, "application/octet-stream", relay, buff, initMetrics, nil, nil, nil, nil, routerPrivateKey[:])
		relayInitErrorAssertions(t, recorder, http.StatusBadRequest, metric)
	}

	// JSON version
	{
		buff, err := packet.MarshalJSON()
		assert.NoError(t, err)

		offset := strings.Index(string(buff), addr)
		assert.GreaterOrEqual(t, offset, 0)
		badAddr := "invalid address"        // "invalid address" is luckily the same number of characters as "127.0.0.1:40000"
		for i := 0; i < len(badAddr); i++ { // Replace the address with the bad address character by character
			buff[offset+i] = badAddr[i]
		}
		recorder := pingRelayBackendInit(t, "application/json", relay, buff, initMetrics, nil, nil, nil, nil, routerPrivateKey[:])
		relayInitErrorAssertions(t, recorder, http.StatusBadRequest, metric)
	}
}

func TestRelayInitRelayNotFound(t *testing.T) {
	addr := "127.0.0.1:40000"
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	assert.NoError(t, err)

	inMemory := &storage.InMemory{} // Have empty storage to fail lookup

	relay := routing.Relay{
		ID: crypto.HashID(addr),
		Datacenter: routing.Datacenter{
			Name: "some datacenter",
		},
	}

	packet := transport.RelayInitRequest{
		Magic:          transport.InitRequestMagic,
		Nonce:          make([]byte, crypto.NonceSize),
		Address:        *udpAddr,
		EncryptedToken: make([]byte, routing.EncryptedRelayTokenSize),
	}

	initMetrics := metrics.EmptyRelayInitMetrics
	localMetrics := metrics.LocalHandler{}

	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
	assert.NoError(t, err)

	initMetrics.ErrorMetrics.RelayNotFound = metric

	// Binary version
	{
		buff, err := packet.MarshalBinary()
		assert.NoError(t, err)
		recorder := pingRelayBackendInit(t, "application/octet-stream", relay, buff, initMetrics, nil, nil, inMemory, nil, nil)
		relayInitErrorAssertions(t, recorder, http.StatusNotFound, metric)
	}

	// JSON version
	{
		buff, err := packet.MarshalJSON()
		assert.NoError(t, err)
		recorder := pingRelayBackendInit(t, "application/json", relay, buff, initMetrics, nil, nil, inMemory, nil, nil)
		relayInitErrorAssertions(t, recorder, http.StatusNotFound, metric)
	}
}

func TestRelayInitInvalidToken(t *testing.T) {
	_, routerPrivateKey, err := box.GenerateKey(crand.Reader)
	assert.NoError(t, err)

	// generate nonce
	nonce := make([]byte, crypto.NonceSize)
	crand.Read(nonce)

	// generate token but leave it as 0's
	token := make([]byte, routing.EncryptedRelayTokenSize)

	addr := "127.0.0.1:40000"
	udp, err := net.ResolveUDPAddr("udp", addr)
	assert.NoError(t, err)
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

	initMetrics := metrics.EmptyRelayInitMetrics
	localMetrics := metrics.LocalHandler{}

	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
	assert.NoError(t, err)

	initMetrics.ErrorMetrics.DecryptionFailure = metric

	// Binary version
	{
		buff, err := packet.MarshalBinary()
		assert.NoError(t, err)
		recorder := pingRelayBackendInit(t, "application/octet-stream", relay, buff, initMetrics, nil, nil, nil, nil, routerPrivateKey[:])
		relayInitErrorAssertions(t, recorder, http.StatusUnauthorized, metric)
	}

	// JSON version
	{
		buff, err := packet.MarshalJSON()
		assert.NoError(t, err)
		recorder := pingRelayBackendInit(t, "application/json", relay, buff, initMetrics, nil, nil, nil, nil, routerPrivateKey[:])
		relayInitErrorAssertions(t, recorder, http.StatusUnauthorized, metric)
	}
}

func TestRelayInitInvalidNonce(t *testing.T) {
	relayPublicKey, relayPrivateKey := getRelayKeyPair(t)
	routerPublicKey, routerPrivateKey, err := box.GenerateKey(crand.Reader)
	assert.NoError(t, err)

	// generate nonce
	nonce := make([]byte, crypto.NonceSize)
	crand.Read(nonce)

	// generate random token
	token := make([]byte, crypto.KeySize)
	crand.Read(token)

	// seal it with the bad nonce
	encryptedToken := crypto.Seal(token, nonce, routerPublicKey[:], relayPrivateKey[:])

	addr := "127.0.0.1:40000"
	udp, err := net.ResolveUDPAddr("udp", addr)
	assert.NoError(t, err)
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
		Nonce:          make([]byte, crypto.NonceSize), // Send a different nonce than the one used
		Address:        *udp,
		EncryptedToken: encryptedToken,
	}

	initMetrics := metrics.EmptyRelayInitMetrics
	localMetrics := metrics.LocalHandler{}

	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
	assert.NoError(t, err)

	initMetrics.ErrorMetrics.DecryptionFailure = metric

	// Binary version
	{
		buff, err := packet.MarshalBinary()
		assert.NoError(t, err)
		recorder := pingRelayBackendInit(t, "application/octet-stream", relay, buff, initMetrics, nil, nil, nil, nil, routerPrivateKey[:])
		relayInitErrorAssertions(t, recorder, http.StatusUnauthorized, metric)
	}

	// JSON version
	{
		buff, err := packet.MarshalJSON()
		assert.NoError(t, err)
		recorder := pingRelayBackendInit(t, "application/json", relay, buff, initMetrics, nil, nil, nil, nil, routerPrivateKey[:])
		relayInitErrorAssertions(t, recorder, http.StatusUnauthorized, metric)
	}
}

func TestRelayInitRelayRedisFailure(t *testing.T) {
	// Don't establish a redis server so redis calls fail
	redisClient := redis.NewClient(&redis.Options{Addr: "0.0.0.0"})

	relayPublicKey, relayPrivateKey := getRelayKeyPair(t)
	routerPublicKey, routerPrivateKey, err := box.GenerateKey(crand.Reader)
	assert.NoError(t, err)

	nonce := make([]byte, crypto.NonceSize)
	crand.Read(nonce)

	addr := "127.0.0.1:40000"
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	assert.NoError(t, err)

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

	initMetrics := metrics.EmptyRelayInitMetrics
	localMetrics := metrics.LocalHandler{}

	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
	assert.NoError(t, err)

	initMetrics.ErrorMetrics.RedisFailure = metric

	// Binary version
	{
		buff, err := packet.MarshalBinary()
		assert.NoError(t, err)
		recorder := pingRelayBackendInit(t, "application/octet-stream", relay, buff, initMetrics, nil, nil, nil, redisClient, routerPrivateKey[:])
		relayInitErrorAssertions(t, recorder, http.StatusInternalServerError, metric)
	}

	// JSON version
	{
		buff, err := packet.MarshalJSON()
		assert.NoError(t, err)
		recorder := pingRelayBackendInit(t, "application/json", relay, buff, initMetrics, nil, nil, nil, redisClient, routerPrivateKey[:])
		relayInitErrorAssertions(t, recorder, http.StatusInternalServerError, metric)
	}
}

func TestRelayInitRelayExists(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)
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
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	assert.NoError(t, err)
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

	initMetrics := metrics.EmptyRelayInitMetrics
	localMetrics := metrics.LocalHandler{}

	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
	assert.NoError(t, err)

	initMetrics.ErrorMetrics.RelayAlreadyExists = metric

	// Binary version
	{
		buff, err := packet.MarshalBinary()
		assert.NoError(t, err)
		recorder := pingRelayBackendInit(t, "application/octet-stream", relay, buff, initMetrics, nil, nil, nil, redisClient, routerPrivateKey[:])
		relayInitErrorAssertions(t, recorder, http.StatusConflict, metric)
	}

	// JSON version
	{
		buff, err := packet.MarshalJSON()
		assert.NoError(t, err)
		recorder := pingRelayBackendInit(t, "application/json", relay, buff, initMetrics, nil, nil, nil, redisClient, routerPrivateKey[:])
		relayInitErrorAssertions(t, recorder, http.StatusConflict, metric)
	}
}

func TestRelayInitRelayIPLookupFailure(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)
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
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	assert.NoError(t, err)

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

	initMetrics := metrics.EmptyRelayInitMetrics
	localMetrics := metrics.LocalHandler{}

	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
	assert.NoError(t, err)

	initMetrics.ErrorMetrics.IPLookupFailure = metric

	// Binary version
	{
		buff, err := packet.MarshalBinary()
		assert.NoError(t, err)
		recorder := pingRelayBackendInit(t, "application/octet-stream", relay, buff, initMetrics, nil, ipfunc, nil, redisClient, routerPrivateKey[:])
		relayInitErrorAssertions(t, recorder, http.StatusOK, metric)
	}

	// clear redis
	redisServer, err = miniredis.Run()
	assert.NoError(t, err)
	redisClient = redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	// JSON version
	{
		buff, err := packet.MarshalJSON()
		assert.NoError(t, err)
		recorder := pingRelayBackendInit(t, "application/json", relay, buff, initMetrics, nil, ipfunc, nil, redisClient, routerPrivateKey[:])
		relayInitErrorAssertions(t, recorder, http.StatusOK, metric)
	}
}

func TestRelayInitSuccess(t *testing.T) {
	redisServer, err := miniredis.Run()
	assert.NoError(t, err)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	relayPublicKey, relayPrivateKey := getRelayKeyPair(t)
	routerPublicKey, routerPrivateKey, err := box.GenerateKey(crand.Reader)
	assert.NoError(t, err)

	var geoClient routing.GeoClient
	{
		redisServer, err := miniredis.Run()
		assert.NoError(t, err)
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
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	assert.NoError(t, err)

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

	localMetrics := metrics.LocalHandler{}
	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
	assert.NoError(t, err)

	initMetrics := metrics.RelayInitMetrics{
		Invocations:   &metrics.EmptyCounter{},
		DurationGauge: &metrics.EmptyGauge{},
	}
	v := reflect.ValueOf(&initMetrics.ErrorMetrics).Elem()
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).CanSet() {
			v.Field(i).Set(reflect.ValueOf(metric))
		}
	}

	// Binary version
	{
		buff, err := packet.MarshalBinary()
		assert.NoError(t, err)

		recorder := pingRelayBackendInit(t, "application/octet-stream", relay, buff, initMetrics, &geoClient, ipfunc, nil, redisClient, routerPrivateKey[:])
		relayInitSuccessAssertions(t, recorder, "application/octet-stream", initMetrics.ErrorMetrics, &geoClient, redisClient, location, addr, before, expected)
	}

	// clear redis
	redisServer, err = miniredis.Run()
	assert.NoError(t, err)
	redisClient = redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	// JSON version
	{
		buff, err := packet.MarshalJSON()
		assert.NoError(t, err)

		recorder := pingRelayBackendInit(t, "application/json", relay, buff, initMetrics, &geoClient, ipfunc, nil, redisClient, routerPrivateKey[:])
		relayInitSuccessAssertions(t, recorder, "application/json", initMetrics.ErrorMetrics, &geoClient, redisClient, location, addr, before, expected)
	}
}
