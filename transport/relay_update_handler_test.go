package transport_test

import (
	"bytes"
	"encoding/json"
	"math/rand"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
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
)

func relayUpdateAssertions(t *testing.T, contentType string, body []byte, expectedCode int, redisClient *redis.Client, statsdb *routing.StatsDatabase) *httptest.ResponseRecorder {
	if redisClient == nil {
		redisServer, _ := miniredis.Run()
		redisClient = redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	}

	if statsdb == nil {
		statsdb = routing.NewStatsDatabase()
	}

	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("POST", "/relay_update", bytes.NewBuffer(body))
	request.Header.Add("Content-Type", contentType)

	handler := transport.RelayUpdateHandlerFunc(log.NewNopLogger(), &transport.RelayUpdateHandlerConfig{
		RedisClient:           redisClient,
		StatsDb:               statsdb,
		Metrics:               &metrics.EmptyRelayUpdateMetrics,
		TrafficStatsPublisher: &stats.NoOpTrafficStatsPublisher{},
		Storer:                &storage.InMemory{},
	})

	handler(recorder, request)

	assert.Equal(t, expectedCode, recorder.Code)

	return recorder
}

func validateRelayUpdateSuccess(t *testing.T, expectedContentType string, recorder *httptest.ResponseRecorder, entry routing.Relay, redisClient *redis.Client, statsdb *routing.StatsDatabase, statIps []string, addr string) {
	res := redisClient.HGet(routing.HashKeyAllRelays, entry.Key())
	var actual routing.Relay
	raw, err := res.Bytes()
	assert.NoError(t, err)
	actual.UnmarshalBinary(raw)

	assert.Equal(t, entry.ID, actual.ID)
	assert.Equal(t, entry.Name, actual.Name)
	assert.Equal(t, entry.Addr, actual.Addr)
	assert.Equal(t, entry.Datacenter.ID, actual.Datacenter.ID)
	assert.Equal(t, entry.Datacenter.Name, actual.Datacenter.Name)
	assert.Equal(t, entry.PublicKey, actual.PublicKey)
	assert.NotEqual(t, entry.LastUpdateTime, actual.LastUpdateTime)

	// response assertions
	header := recorder.Header()
	contentType, ok := header["Content-Type"]
	assert.True(t, ok)
	if assert.NotNil(t, contentType) && assert.Len(t, contentType, 1) {
		assert.Equal(t, expectedContentType, contentType[0])
	}

	body := recorder.Body.Bytes()

	var response transport.RelayUpdateResponse
	switch expectedContentType {
	case "application/octet-stream":
		err := response.UnmarshalBinary(body)
		assert.NoError(t, err)
	case "application/json":
		err := json.Unmarshal(body, &response)
		assert.NoError(t, err)
	default:
		assert.FailNow(t, "Invalid expected content type")
	}

	assert.Equal(t, len(statIps), len(response.RelaysToPing))

	relaysToPingIDs := make([]uint64, 0)
	relaysToPingAddrs := make([]string, 0)

	for _, data := range response.RelaysToPing {
		relaysToPingIDs = append(relaysToPingIDs, data.ID)
		relaysToPingAddrs = append(relaysToPingAddrs, data.Address)
	}

	assert.Contains(t, statsdb.Entries, entry.ID)
	relations := statsdb.Entries[entry.ID]
	for _, addr := range statIps {
		id := crypto.HashID(addr)
		assert.Contains(t, relaysToPingIDs, id)
		assert.Contains(t, relaysToPingAddrs, addr)
		assert.Contains(t, relations.Relays, id)
	}

	assert.NotContains(t, relaysToPingIDs, entry.ID)
	assert.NotContains(t, relaysToPingAddrs, addr)
	assert.Equal(t, uint32(routing.RelayStateShuttingDown), actual.State)
}

func TestRelayUpdateRelayInvalid(t *testing.T) {
	// Binary version
	{
		buff := make([]byte, 10) // invalid relay packet size
		relayUpdateAssertions(t, "application/octet-stream", buff, http.StatusBadRequest, nil, nil)
	}

	// JSON version
	{
		buff := []byte("{") // basic but gets the job done
		relayUpdateAssertions(t, "application/json", buff, http.StatusBadRequest, nil, nil)
	}
}

func TestRelayUpdateUnequalTokens(t *testing.T) {
	redisServer, _ := miniredis.Run()
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	addr := "127.0.0.1:40000"
	udp, _ := net.ResolveUDPAddr("udp", addr)

	incomingToken := make([]byte, crypto.KeySize)
	rand.Read(incomingToken)
	storedToken := make([]byte, crypto.KeySize)
	rand.Read(storedToken)
	packet := transport.RelayUpdateRequest{
		Address:   *udp,
		Token:     incomingToken,
		PingStats: make([]routing.RelayStatsPing, 0),
	}

	entry := routing.Relay{
		ID:   crypto.HashID(addr),
		Addr: *udp,
		Datacenter: routing.Datacenter{
			ID:   1,
			Name: "some name",
		},
		PublicKey:      storedToken,
		LastUpdateTime: uint64(time.Now().Unix() - 1),
	}

	raw, err := entry.MarshalBinary()
	assert.NoError(t, err)
	redisServer.HSet(routing.HashKeyAllRelays, entry.Key(), string(raw))

	// Binary version
	{
		buff, err := packet.MarshalBinary()
		assert.NoError(t, err)
		relayUpdateAssertions(t, "application/octet-stream", buff, http.StatusBadRequest, redisClient, nil)
	}

	// JSON version
	{
		buff, err := json.Marshal(packet)
		assert.NoError(t, err)
		relayUpdateAssertions(t, "application/json", buff, http.StatusBadRequest, redisClient, nil)
	}
}

func TestRelayUpdateInvalidAddress(t *testing.T) {
	t.Skip("Test can fail on certain machines due to relay address being unmarshaled and interpreted as correct. Needs more work to determine the cause.")

	udp, _ := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
	packet := transport.RelayUpdateRequest{
		Address: *udp,
		Token:   make([]byte, crypto.KeySize),
	}

	// Binary version
	{
		buff, err := packet.MarshalBinary()
		assert.NoError(t, err)
		buff[8] = 'x' // assign this index (which should be the first item in the address) as the letter 'x' making it invalid
		relayUpdateAssertions(t, "application/octet-stream", buff, http.StatusBadRequest, nil, nil)
	}

	// JSON version
	{
		buff, err := json.Marshal(packet)
		assert.NoError(t, err)

		offset := strings.Index(string(buff), "127.0.0.1:40000")
		assert.GreaterOrEqual(t, offset, 0)
		buff[offset] = 'x' // assign this index (which should be the first item in the address) as the letter 'x' making it invalid
		relayUpdateAssertions(t, "application/json", buff, http.StatusBadRequest, nil, nil)
	}
}

func TestRelayUpdateExceedMaxRelays(t *testing.T) {
	udp, _ := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
	packet := transport.RelayUpdateRequest{
		Address:   *udp,
		Token:     make([]byte, crypto.KeySize),
		PingStats: make([]routing.RelayStatsPing, 1025),
	}

	// Binary version
	{
		buff, err := packet.MarshalBinary()
		assert.NoError(t, err)
		relayUpdateAssertions(t, "application/octet-stream", buff, http.StatusBadRequest, nil, nil)
	}

	// JSON version
	{
		buff, err := json.Marshal(packet)
		assert.NoError(t, err)
		relayUpdateAssertions(t, "application/json", buff, http.StatusBadRequest, nil, nil)
	}
}

func TestRelayUpdateRelayNotFound(t *testing.T) {
	udp, _ := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
	packet := transport.RelayUpdateRequest{
		Address:   *udp,
		Token:     make([]byte, crypto.KeySize),
		PingStats: make([]routing.RelayStatsPing, 3),
	}

	// Binary version
	{
		buff, err := packet.MarshalBinary()
		assert.NoError(t, err)
		relayUpdateAssertions(t, "application/octet-stream", buff, http.StatusNotFound, nil, nil)
	}

	// JSON version
	{
		buff, err := json.Marshal(packet)
		assert.NoError(t, err)
		relayUpdateAssertions(t, "application/json", buff, http.StatusNotFound, nil, nil)
	}
}

func TestRelayUpdateSuccess(t *testing.T) {
	redisServer, _ := miniredis.Run()
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
	addr := "127.0.0.1:40000"
	udp, _ := net.ResolveUDPAddr("udp", addr)
	statsdb := routing.NewStatsDatabase()
	statIps := []string{"127.0.0.2:40000", "127.0.0.3:40000", "127.0.0.4:40000", "127.0.0.5:40000"}
	packet := transport.RelayUpdateRequest{
		Address:      *udp,
		Token:        make([]byte, crypto.KeySize),
		ShuttingDown: true,
	}

	packet.PingStats = make([]routing.RelayStatsPing, len(statIps))
	for i, addr := range statIps {
		stats := &packet.PingStats[i]
		stats.RelayID = crypto.HashID(addr)
		stats.RTT = rand.Float32()
		stats.Jitter = rand.Float32()
		stats.PacketLoss = rand.Float32()
	}

	seedRedis(t, redisServer, statIps)

	entry := routing.Relay{
		ID:   crypto.HashID(addr),
		Addr: *udp,
		Datacenter: routing.Datacenter{
			ID:   1,
			Name: "some name",
		},
		PublicKey:      make([]byte, crypto.KeySize),
		LastUpdateTime: uint64(time.Now().Unix() - 1),
		State:          routing.RelayStateOnline,
	}

	raw, err := entry.MarshalBinary()
	assert.NoError(t, err)
	redisServer.HSet(routing.HashKeyAllRelays, entry.Key(), string(raw))

	// Binary version
	{
		buff, err := packet.MarshalBinary()
		assert.NoError(t, err)
		recorder := relayUpdateAssertions(t, "application/octet-stream", buff, http.StatusOK, redisClient, statsdb)
		validateRelayUpdateSuccess(t, "application/octet-stream", recorder, entry, redisClient, statsdb, statIps, addr)
	}

	// JSON version
	{
		buff, err := json.Marshal(packet)
		assert.NoError(t, err)
		recorder := relayUpdateAssertions(t, "application/json", buff, http.StatusOK, redisClient, statsdb)
		validateRelayUpdateSuccess(t, "application/json", recorder, entry, redisClient, statsdb, statIps, addr)
	}
}
