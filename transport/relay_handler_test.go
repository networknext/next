package transport_test

import (
	"bytes"
	crand "crypto/rand"
	"encoding/base64"
	"math"
	mrand "math/rand"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-kit/kit/log"
	"github.com/go-redis/redis/v7"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/metrics"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/stats"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/nacl/box"
)

func relayHandlerAssertions(t *testing.T, token string, relay routing.Relay, headers map[string]string, body []byte, expectedCode int, geoClient *routing.GeoClient, ipfunc routing.LocateIPFunc, inMemory *storage.InMemory, redisClient *redis.Client, statsdb *routing.StatsDatabase, routerPrivateKey []byte) *httptest.ResponseRecorder {
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

	if statsdb == nil {
		statsdb = routing.NewStatsDatabase()
	}

	recorder := httptest.NewRecorder()
	request, err := http.NewRequest("POST", "/relays", bytes.NewBuffer(body))
	assert.NoError(t, err)

	request.Header.Add("Content-Type", "application/json")
	for key, val := range headers {
		request.Header.Add(key, val)
	}

	handler := transport.RelayHandlerFunc(log.NewNopLogger(), &transport.RelayHandlerConfig{
		RedisClient:           redisClient,
		GeoClient:             geoClient,
		IpLocator:             ipfunc,
		Storer:                inMemory,
		StatsDb:               statsdb,
		TrafficStatsPublisher: &stats.NoOpTrafficStatsPublisher{},
		Metrics:               &metrics.EmptyRelayHandlerMetrics,
		RouterPrivateKey:      routerPrivateKey,
	})

	handler(recorder, request)

	assert.Equal(t, expectedCode, recorder.Code)

	return recorder
}

func validateRelayHandlerSuccess(t *testing.T, recorder *httptest.ResponseRecorder, geoClient routing.GeoClient, redisClient *redis.Client, location routing.Location, statsdb *routing.StatsDatabase, addr string, expected routing.Relay, statIps []string) {
	// Validate redis entry is correct
	entry := redisClient.HGet(routing.HashKeyAllRelays, expected.Key())

	var actual routing.Relay
	entryBytes, err := entry.Bytes()
	assert.NoError(t, err)

	err = actual.UnmarshalBinary(entryBytes)
	assert.NoError(t, err)

	assert.Equal(t, expected.ID, actual.ID)
	assert.Equal(t, expected.Name, actual.Name)
	assert.Equal(t, expected.Addr, actual.Addr)
	assert.NotZero(t, actual.LastUpdateTime)
	assert.NotEqual(t, expected.LastUpdateTime, actual.LastUpdateTime)
	assert.Equal(t, expected.PublicKey, actual.PublicKey)

	// Validate geoclient entry is correct
	// only added one relay so it should be the only one returned by this
	relaysInLocation, err := geoClient.RelaysWithin(location.Latitude, location.Longitude, 1, "km")
	assert.NoError(t, err)
	if assert.Len(t, relaysInLocation, 1) {
		relay := relaysInLocation[0]

		assert.Equal(t, crypto.HashID(addr), relay.ID)
		assert.Equal(t, location.Latitude, math.Round(relay.Latitude*1000)/1000)
		assert.Equal(t, location.Longitude, math.Round(relay.Longitude*1000)/1000)
	}

	// Validate response header is correct
	header := recorder.Header()

	contentType := header.Get("Content-Type")
	assert.Equal(t, "application/json", contentType)

	// Validate response body is correct
	body := recorder.Body.Bytes()

	var response transport.RelayRequest
	response.UnmarshalJSON(body)

	assert.Equal(t, len(statIps), len(response.PingStats))

	relaysToPingIDs := make([]uint64, 0)
	relaysToPingAddrs := make([]string, 0)

	for _, data := range response.PingStats {
		relaysToPingIDs = append(relaysToPingIDs, data.ID)
		relaysToPingAddrs = append(relaysToPingAddrs, data.Address)
	}

	// Validate statsDB is correct
	assert.Contains(t, statsdb.Entries, expected.ID)
	relations := statsdb.Entries[expected.ID]
	for _, addr := range statIps {
		id := crypto.HashID(addr)
		assert.Contains(t, relaysToPingIDs, id)
		assert.Contains(t, relaysToPingAddrs, addr)
		assert.Contains(t, relations.Relays, id)
	}

	assert.NotContains(t, relaysToPingIDs, expected.ID)
	assert.NotContains(t, relaysToPingAddrs, addr)
}

func TestRelayHandlerUnreadableRequest(t *testing.T) {
	addr := "127.0.0.1:40000"
	relay := routing.Relay{
		ID: crypto.HashID(addr),
		Datacenter: routing.Datacenter{
			Name: "some datacenter",
		},
	}

	buff := []byte("bad packet")
	relayHandlerAssertions(t, "", relay, nil, buff, http.StatusBadRequest, nil, nil, nil, nil, nil, nil)
}

func TestRelayHandlerExceedMaxRelays(t *testing.T) {
	addr := "127.0.0.1:40000"
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	assert.NoError(t, err)

	relay := routing.Relay{
		ID:   crypto.HashID(addr),
		Addr: *udpAddr,
		Datacenter: routing.Datacenter{
			Name: "some datacenter",
		},
	}

	request := transport.RelayRequest{
		Address:   *udpAddr,
		PingStats: make([]transport.RelayPingStats, 1025),
	}

	buff, err := request.MarshalJSON()
	assert.NoError(t, err)
	relayHandlerAssertions(t, "", relay, nil, buff, http.StatusBadRequest, nil, nil, nil, nil, nil, nil)
}

func TestRelayHandlerRelayNotFound(t *testing.T) {
	addr := "127.0.0.1:40000"
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	assert.NoError(t, err)

	relay := routing.Relay{
		ID:   crypto.HashID(addr),
		Addr: *udpAddr,
		Datacenter: routing.Datacenter{
			Name: "some datacenter",
		},
	}

	inMemory := &storage.InMemory{} // Empty DB storage

	request := transport.RelayRequest{
		Address: *udpAddr,
	}

	buff, err := request.MarshalJSON()
	assert.NoError(t, err)
	relayHandlerAssertions(t, "", relay, nil, buff, http.StatusInternalServerError, nil, nil, inMemory, nil, nil, nil)
}

func TestRelayHandlerNoAuthHeader(t *testing.T) {
	addr := "127.0.0.1:40000"
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	assert.NoError(t, err)

	relay := routing.Relay{
		ID:   crypto.HashID(addr),
		Addr: *udpAddr,
		Datacenter: routing.Datacenter{
			Name: "some datacenter",
		},
	}

	inMemory := &storage.InMemory{
		LocalRelays: []routing.Relay{relay},
	}

	request := transport.RelayRequest{
		Address: *udpAddr,
	}

	buff, err := request.MarshalJSON()
	assert.NoError(t, err)
	relayHandlerAssertions(t, "", relay, nil, buff, http.StatusUnauthorized, nil, nil, inMemory, nil, nil, nil)
}

func TestRelayHandlerBadAuthHeaderLength(t *testing.T) {
	addr := "127.0.0.1:40000"
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	assert.NoError(t, err)

	relay := routing.Relay{
		ID:   crypto.HashID(addr),
		Addr: *udpAddr,
		Datacenter: routing.Datacenter{
			Name: "some datacenter",
		},
	}

	inMemory := &storage.InMemory{
		LocalRelays: []routing.Relay{relay},
	}

	request := transport.RelayRequest{
		Address: *udpAddr,
	}

	buff, err := request.MarshalJSON()
	assert.NoError(t, err)

	// Set auth HTTP header
	headers := make(map[string]string)
	headers["Authorization"] = "bad"

	relayHandlerAssertions(t, "", relay, headers, buff, http.StatusBadRequest, nil, nil, inMemory, nil, nil, nil)
}

func TestRelayHandlerBadAuthHeaderToken(t *testing.T) {
	addr := "127.0.0.1:40000"
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	assert.NoError(t, err)

	relay := routing.Relay{
		ID:   crypto.HashID(addr),
		Addr: *udpAddr,
		Datacenter: routing.Datacenter{
			Name: "some datacenter",
		},
	}

	inMemory := &storage.InMemory{
		LocalRelays: []routing.Relay{relay},
	}

	request := transport.RelayRequest{
		Address: *udpAddr,
	}

	buff, err := request.MarshalJSON()
	assert.NoError(t, err)

	// Set auth HTTP header
	headers := make(map[string]string)
	headers["Authorization"] = "Bearer bad token"

	relayHandlerAssertions(t, "", relay, headers, buff, http.StatusBadRequest, nil, nil, inMemory, nil, nil, nil)
}

func TestRelayHandlerBadNonce(t *testing.T) {
	addr := "127.0.0.1:40000"
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	assert.NoError(t, err)

	relay := routing.Relay{
		ID:   crypto.HashID(addr),
		Addr: *udpAddr,
		Datacenter: routing.Datacenter{
			Name: "some datacenter",
		},
	}

	inMemory := &storage.InMemory{
		LocalRelays: []routing.Relay{relay},
	}

	request := transport.RelayRequest{
		Address: *udpAddr,
	}

	buff, err := request.MarshalJSON()
	assert.NoError(t, err)

	// Set auth HTTP header
	headers := make(map[string]string)
	headers["Authorization"] = "Bearer invalid:base64"

	relayHandlerAssertions(t, "", relay, headers, buff, http.StatusBadRequest, nil, nil, inMemory, nil, nil, nil)
}

func TestRelayHandlerBadEncryptedAddress(t *testing.T) {
	addr := "127.0.0.1:40000"
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	assert.NoError(t, err)

	relay := routing.Relay{
		ID:   crypto.HashID(addr),
		Addr: *udpAddr,
		Datacenter: routing.Datacenter{
			Name: "some datacenter",
		},
	}

	inMemory := &storage.InMemory{
		LocalRelays: []routing.Relay{relay},
	}

	nonce := make([]byte, crypto.NonceSize)
	crand.Read(nonce)

	nonceBase64 := base64.StdEncoding.EncodeToString(nonce)
	encryptedAddressBase64 := "badaddress"

	token := nonceBase64 + ":" + encryptedAddressBase64

	request := transport.RelayRequest{
		Address: *udpAddr,
	}

	buff, err := request.MarshalJSON()
	assert.NoError(t, err)

	// Set auth HTTP header
	headers := make(map[string]string)
	headers["Authorization"] = "Bearer " + token

	relayHandlerAssertions(t, "", relay, headers, buff, http.StatusBadRequest, nil, nil, inMemory, nil, nil, nil)
}

func TestRelayHandlerDecryptFailure(t *testing.T) {
	addr := "127.0.0.1:40000"
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	assert.NoError(t, err)

	// Don't use the other key in the key pairs to fail decryption
	_, relayPrivateKey := getRelayKeyPair(t)
	routerPublicKey, _, err := box.GenerateKey(crand.Reader)
	assert.NoError(t, err)

	relay := routing.Relay{
		ID:   crypto.HashID(addr),
		Addr: *udpAddr,
		Datacenter: routing.Datacenter{
			Name: "some datacenter",
		},
	}

	inMemory := &storage.InMemory{
		LocalRelays: []routing.Relay{relay},
	}

	nonce := make([]byte, crypto.NonceSize)
	crand.Read(nonce)

	// Encrypt the address
	encryptedAddress := crypto.Seal([]byte(addr), nonce, routerPublicKey[:], relayPrivateKey)

	nonceBase64 := base64.StdEncoding.EncodeToString(nonce)
	encryptedAddressBase64 := base64.StdEncoding.EncodeToString(encryptedAddress)

	token := nonceBase64 + ":" + encryptedAddressBase64

	request := transport.RelayRequest{
		Address: *udpAddr,
	}

	buff, err := request.MarshalJSON()
	assert.NoError(t, err)

	// Set auth HTTP header
	headers := make(map[string]string)
	headers["Authorization"] = "Bearer " + token

	relayHandlerAssertions(t, "", relay, headers, buff, http.StatusUnauthorized, nil, nil, inMemory, nil, nil, nil)
}

func TestRelayHandlerRedisUnmarshalFailure(t *testing.T) {
	addr := "127.0.0.1:40000"
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	assert.NoError(t, err)

	// Don't use the other key in the key pairs to fail decryption
	relayPublicKey, relayPrivateKey := getRelayKeyPair(t)
	routerPublicKey, routerPrivateKey, err := box.GenerateKey(crand.Reader)
	assert.NoError(t, err)

	redisServer, _ := miniredis.Run()
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	relay := routing.Relay{
		ID:   crypto.HashID(addr),
		Addr: *udpAddr,
		Datacenter: routing.Datacenter{
			Name: "some datacenter",
		},
		PublicKey: relayPublicKey,
	}

	// Set a bad entry in redis
	entry := "bad relay entry"
	redisServer.HSet(routing.HashKeyAllRelays, relay.Key(), entry)

	inMemory := &storage.InMemory{
		LocalRelays: []routing.Relay{relay},
	}

	nonce := make([]byte, crypto.NonceSize)
	crand.Read(nonce)

	// Encrypt the address
	encryptedAddress := crypto.Seal([]byte(addr), nonce, routerPublicKey[:], relayPrivateKey)

	nonceBase64 := base64.StdEncoding.EncodeToString(nonce)
	encryptedAddressBase64 := base64.StdEncoding.EncodeToString(encryptedAddress)

	token := nonceBase64 + ":" + encryptedAddressBase64

	request := transport.RelayRequest{
		Address: *udpAddr,
	}

	buff, err := request.MarshalJSON()
	assert.NoError(t, err)

	// Set auth HTTP header
	headers := make(map[string]string)
	headers["Authorization"] = "Bearer " + token

	relayHandlerAssertions(t, "", relay, headers, buff, http.StatusInternalServerError, nil, nil, inMemory, redisClient, nil, routerPrivateKey[:])
}

func TestRelayHandlerSuccess(t *testing.T) {
	addr := "127.0.0.1:40000"
	udpAddr, err := net.ResolveUDPAddr("udp", addr)
	assert.NoError(t, err)

	relayPublicKey, relayPrivateKey := getRelayKeyPair(t)
	routerPublicKey, routerPrivateKey, err := box.GenerateKey(crand.Reader)
	assert.NoError(t, err)

	redisServer, _ := miniredis.Run()
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	statsdb := routing.NewStatsDatabase()

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

	statIps := []string{"127.0.0.2:40000", "127.0.0.3:40000", "127.0.0.4:40000", "127.0.0.5:40000"}

	// Populate redis with the relays to ping
	seedRedis(t, redisServer, statIps)

	// Create relay in DB storage
	relay := routing.Relay{
		ID:   crypto.HashID(addr),
		Addr: *udpAddr,
		Datacenter: routing.Datacenter{
			ID:   1,
			Name: "some name",
		},
		PublicKey:      relayPublicKey,
		Latitude:       13,
		Longitude:      13,
		LastUpdateTime: time.Now().Add(-time.Second),
	}

	var customerPublicKey []byte
	{
		if key := os.Getenv("NEXT_CUSTOMER_PUBLIC_KEY"); len(key) != 0 {
			customerPublicKey, _ = base64.StdEncoding.DecodeString(key)
		}
	}

	inMemory := &storage.InMemory{
		LocalBuyer: &routing.Buyer{
			PublicKey: customerPublicKey[8:],
		},

		LocalRelays: []routing.Relay{relay},
	}

	nonce := make([]byte, crypto.NonceSize)
	crand.Read(nonce)

	// Encrypt the address
	encryptedAddress := crypto.Seal([]byte(addr), nonce, routerPublicKey[:], relayPrivateKey)

	nonceBase64 := base64.StdEncoding.EncodeToString(nonce)
	encryptedAddressBase64 := base64.StdEncoding.EncodeToString(encryptedAddress)

	token := nonceBase64 + ":" + encryptedAddressBase64

	request := transport.RelayRequest{
		Address: *udpAddr,
		PingStats: []transport.RelayPingStats{
			transport.RelayPingStats{
				ID:         crypto.HashID(statIps[0]),
				Address:    statIps[0],
				RTT:        1,
				Jitter:     2,
				PacketLoss: 3,
			},

			transport.RelayPingStats{
				ID:         crypto.HashID(statIps[1]),
				Address:    statIps[1],
				RTT:        4,
				Jitter:     5,
				PacketLoss: 6,
			},

			transport.RelayPingStats{
				ID:         crypto.HashID(statIps[2]),
				Address:    statIps[2],
				RTT:        7,
				Jitter:     8,
				PacketLoss: 9,
			},

			transport.RelayPingStats{
				ID:         crypto.HashID(statIps[3]),
				Address:    statIps[3],
				RTT:        10,
				Jitter:     11,
				PacketLoss: 12,
			},
		},
		TrafficStats: transport.RelayTrafficStats{
			SessionCount:  10,
			BytesSent:     1000000,
			BytesReceived: 1000000,
		},
	}

	expected := routing.Relay{
		ID:   crypto.HashID(addr),
		Addr: *udpAddr,
		Datacenter: routing.Datacenter{
			ID:   1,
			Name: "some name",
		},
		PublicKey: relayPublicKey,
	}

	buff, err := request.MarshalJSON()
	assert.NoError(t, err)

	// Set auth HTTP header
	headers := make(map[string]string)
	headers["Authorization"] = "Bearer " + token

	recorder := relayHandlerAssertions(t, token, relay, headers, buff, http.StatusOK, &geoClient, ipfunc, inMemory, redisClient, statsdb, routerPrivateKey[:])
	validateRelayHandlerSuccess(t, recorder, geoClient, redisClient, location, statsdb, addr, expected, statIps)

	// Now make the same request again, with the relay now initialized
	recorder = relayHandlerAssertions(t, token, relay, headers, buff, http.StatusOK, &geoClient, ipfunc, inMemory, redisClient, statsdb, routerPrivateKey[:])
	validateRelayHandlerSuccess(t, recorder, geoClient, redisClient, location, statsdb, addr, expected, statIps)
}
