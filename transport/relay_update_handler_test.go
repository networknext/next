package transport_test

// todo: come back and fix these tests

// import (
// 	"bytes"
// 	"context"
// 	"encoding/json"
// 	"math/rand"
// 	"net"
// 	"net/http"
// 	"net/http/httptest"
// 	"reflect"
// 	"strings"
// 	"testing"
// 	"time"

// 	"github.com/alicebob/miniredis/v2"
// 	"github.com/go-kit/kit/log"
// 	"github.com/go-redis/redis/v7"
// 	"github.com/networknext/backend/crypto"
// 	"github.com/networknext/backend/metrics"
// 	"github.com/networknext/backend/routing"
// 	"github.com/networknext/backend/storage"

// 	"github.com/networknext/backend/transport"
// 	"github.com/stretchr/testify/assert"
// )

// func pingRelayBackendUpdate(t *testing.T, contentType string, body []byte, metrics metrics.RelayUpdateMetrics, redisClient *redis.Client, geoClient *routing.GeoClient, statsdb *routing.StatsDatabase, inMemory *storage.InMemory) *httptest.ResponseRecorder {
// 	if redisClient == nil {
// 		redisServer, err := miniredis.Run()
// 		assert.NoError(t, err)
// 		redisClient = redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
// 	}

// 	if statsdb == nil {
// 		statsdb = routing.NewStatsDatabase()
// 	}

// 	if inMemory == nil {
// 		inMemory = &storage.InMemory{}

// 	}

// 	recorder := httptest.NewRecorder()
// 	request, err := http.NewRequest("POST", "/relay_update", bytes.NewBuffer(body))
// 	assert.NoError(t, err)
// 	request.Header.Add("Content-Type", contentType)

// 	handler := transport.RelayUpdateHandlerFunc(log.NewNopLogger(), log.NewNopLogger(), &transport.RelayUpdateHandlerConfig{
// 		RedisClient: redisClient,
// 		GeoClient:   geoClient,
// 		StatsDb:     statsdb,
// 		Metrics:     &metrics,
// 		Storer:      inMemory,
// 	})

// 	handler(recorder, request)
// 	return recorder
// }

// func relayUpdateErrorAssertions(t *testing.T, recorder *httptest.ResponseRecorder, expectedCode int, errMetric metrics.Counter) {
// 	assert.Equal(t, expectedCode, recorder.Code)
// 	assert.Equal(t, 1.0, errMetric.ValueReset())
// }

// func relayUpdateShutdownAssertions(t *testing.T, errMetrics metrics.RelayUpdateErrorMetrics, entry routing.RelayCacheEntry, redisClient *redis.Client, inMemory *storage.InMemory, statsdb *routing.StatsDatabase, addr string) {
// 	res := redisClient.HGet(routing.HashKeyAllRelays, entry.Key())
// 	assert.NotNil(t, res.Err())
// 	assert.EqualError(t, redis.Nil, res.Err().Error())

// 	relay, err := inMemory.Relay(crypto.HashID(addr))
// 	assert.NoError(t, err)

// 	assert.Equal(t, routing.RelayStateMaintenance, relay.State)

// 	for i, stat := range statsdb.Entries {
// 		assert.NotEqual(t, i, relay.ID)
// 		for j := range stat.Relays {
// 			assert.NotEqual(t, j, relay.ID)
// 		}
// 	}

// 	errMetricsStruct := reflect.ValueOf(errMetrics)
// 	for i := 0; i < errMetricsStruct.NumField(); i++ {
// 		if errMetricsStruct.Field(i).CanInterface() {
// 			assert.Equal(t, 0.0, errMetricsStruct.Field(i).Interface().(metrics.Counter).ValueReset())
// 		}
// 	}
// }

// func relayUpdateSuccessAssertions(t *testing.T, recorder *httptest.ResponseRecorder, expectedContentType string, errMetrics metrics.RelayUpdateErrorMetrics, entry routing.RelayCacheEntry, redisClient *redis.Client, inMemory *storage.InMemory, statsdb *routing.StatsDatabase, statIps []string, addr string) {
// 	assert.Equal(t, http.StatusOK, recorder.Code)

// 	res := redisClient.HGet(routing.HashKeyAllRelays, entry.Key())
// 	var actual routing.RelayCacheEntry
// 	raw, err := res.Bytes()
// 	assert.NoError(t, err)
// 	actual.UnmarshalBinary(raw)

// 	assert.Equal(t, entry.ID, actual.ID)
// 	assert.Equal(t, entry.Name, actual.Name)
// 	assert.Equal(t, entry.Addr, actual.Addr)
// 	assert.Equal(t, entry.Datacenter.ID, actual.Datacenter.ID)
// 	assert.Equal(t, entry.Datacenter.Name, actual.Datacenter.Name)
// 	assert.Equal(t, entry.PublicKey, actual.PublicKey)
// 	assert.NotEqual(t, entry.LastUpdateTime, actual.LastUpdateTime)

// 	// response assertions
// 	header := recorder.Header()
// 	contentType, ok := header["Content-Type"]
// 	assert.True(t, ok)
// 	if assert.NotNil(t, contentType) && assert.Len(t, contentType, 1) {
// 		assert.Equal(t, expectedContentType, contentType[0])
// 	}

// 	body := recorder.Body.Bytes()

// 	var response transport.RelayUpdateResponse
// 	switch expectedContentType {
// 	case "application/octet-stream":
// 		err := response.UnmarshalBinary(body)
// 		assert.NoError(t, err)
// 	case "application/json":
// 		err := json.Unmarshal(body, &response)
// 		assert.NoError(t, err)
// 	default:
// 		assert.FailNow(t, "Invalid expected content type")
// 	}

// 	assert.Equal(t, len(statIps), len(response.RelaysToPing))

// 	relaysToPingIDs := make([]uint64, 0)
// 	relaysToPingAddrs := make([]string, 0)

// 	for _, data := range response.RelaysToPing {
// 		relaysToPingIDs = append(relaysToPingIDs, data.ID)
// 		relaysToPingAddrs = append(relaysToPingAddrs, data.Address)
// 	}

// 	assert.Contains(t, statsdb.Entries, entry.ID)
// 	relations := statsdb.Entries[entry.ID]
// 	for _, addr := range statIps {
// 		id := crypto.HashID(addr)
// 		assert.Contains(t, relaysToPingIDs, id)
// 		assert.Contains(t, relaysToPingAddrs, addr)
// 		assert.Contains(t, relations.Relays, id)
// 	}

// 	assert.NotContains(t, relaysToPingIDs, entry.ID)
// 	assert.NotContains(t, relaysToPingAddrs, addr)

// 	relay, err := inMemory.Relay(crypto.HashID(addr))
// 	assert.NoError(t, err)

// 	assert.Equal(t, routing.RelayStateEnabled, relay.State)

// 	errMetricsStruct := reflect.ValueOf(errMetrics)
// 	for i := 0; i < errMetricsStruct.NumField(); i++ {
// 		if errMetricsStruct.Field(i).CanInterface() {
// 			assert.Equal(t, 0.0, errMetricsStruct.Field(i).Interface().(metrics.Counter).ValueReset())
// 		}
// 	}
// }

// func TestRelayUpdateUnmarshalFailure(t *testing.T) {
// 	updateMetrics := metrics.EmptyRelayUpdateMetrics
// 	localMetrics := metrics.LocalHandler{}

// 	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
// 	assert.NoError(t, err)

// 	updateMetrics.ErrorMetrics.UnmarshalFailure = metric

// 	// Binary version
// 	{
// 		buff := make([]byte, 10) // invalid relay packet size
// 		recorder := pingRelayBackendUpdate(t, "application/octet-stream", buff, updateMetrics, nil, nil, nil, nil)
// 		relayUpdateErrorAssertions(t, recorder, http.StatusBadRequest, metric)
// 	}

// 	// JSON version
// 	{
// 		buff := []byte("{") // basic but gets the job done
// 		recorder := pingRelayBackendUpdate(t, "application/json", buff, updateMetrics, nil, nil, nil, nil)
// 		relayUpdateErrorAssertions(t, recorder, http.StatusBadRequest, metric)
// 	}
// }

// func TestRelayUpdateInvalidAddress(t *testing.T) {
// 	udp, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
// 	assert.NoError(t, err)
// 	packet := transport.RelayUpdateRequest{
// 		Address: *udp,
// 		Token:   make([]byte, crypto.KeySize),
// 	}

// 	updateMetrics := metrics.EmptyRelayUpdateMetrics
// 	localMetrics := metrics.LocalHandler{}

// 	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
// 	assert.NoError(t, err)

// 	updateMetrics.ErrorMetrics.UnmarshalFailure = metric

// 	// Binary version
// 	{
// 		buff, err := packet.MarshalBinary()
// 		assert.NoError(t, err)
// 		badAddr := "invalid address"        // "invalid address" is luckily the same number of characters as "127.0.0.1:40000"
// 		for i := 0; i < len(badAddr); i++ { // Replace the address with the bad address character by character
// 			buff[8+i] = badAddr[i]
// 		}
// 		recorder := pingRelayBackendUpdate(t, "application/octet-stream", buff, updateMetrics, nil, nil, nil, nil)
// 		relayUpdateErrorAssertions(t, recorder, http.StatusBadRequest, metric)
// 	}

// 	// JSON version
// 	{
// 		buff, err := packet.MarshalJSON()
// 		assert.NoError(t, err)

// 		offset := strings.Index(string(buff), "127.0.0.1:40000")
// 		assert.GreaterOrEqual(t, offset, 0)
// 		badAddr := "invalid address"        // "invalid address" is luckily the same number of characters as "127.0.0.1:40000"
// 		for i := 0; i < len(badAddr); i++ { // Replace the address with the bad address character by character
// 			buff[offset+i] = badAddr[i]
// 		}
// 		recorder := pingRelayBackendUpdate(t, "application/json", buff, updateMetrics, nil, nil, nil, nil)
// 		relayUpdateErrorAssertions(t, recorder, http.StatusBadRequest, metric)
// 	}
// }

// func TestRelayUpdateInvalidVersion(t *testing.T) {
// 	udp, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
// 	assert.NoError(t, err)
// 	packet := transport.RelayUpdateRequest{
// 		Version: 1,
// 		Address: *udp,
// 		Token:   make([]byte, crypto.KeySize),
// 	}

// 	updateMetrics := metrics.EmptyRelayUpdateMetrics
// 	localMetrics := metrics.LocalHandler{}

// 	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
// 	assert.NoError(t, err)

// 	updateMetrics.ErrorMetrics.InvalidVersion = metric

// 	// Binary version
// 	{
// 		buff, err := packet.MarshalBinary()
// 		assert.NoError(t, err)
// 		recorder := pingRelayBackendUpdate(t, "application/octet-stream", buff, updateMetrics, nil, nil, nil, nil)
// 		relayUpdateErrorAssertions(t, recorder, http.StatusBadRequest, metric)
// 	}

// 	// JSON version
// 	{
// 		buff, err := packet.MarshalJSON()
// 		assert.NoError(t, err)
// 		recorder := pingRelayBackendUpdate(t, "application/json", buff, updateMetrics, nil, nil, nil, nil)
// 		relayUpdateErrorAssertions(t, recorder, http.StatusBadRequest, metric)
// 	}
// }

// func TestRelayUpdateExceedMaxRelays(t *testing.T) {
// 	udp, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
// 	assert.NoError(t, err)
// 	packet := transport.RelayUpdateRequest{
// 		Address:   *udp,
// 		Token:     make([]byte, crypto.KeySize),
// 		PingStats: make([]routing.RelayStatsPing, 1025),
// 	}

// 	updateMetrics := metrics.EmptyRelayUpdateMetrics
// 	localMetrics := metrics.LocalHandler{}

// 	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
// 	assert.NoError(t, err)

// 	updateMetrics.ErrorMetrics.ExceedMaxRelays = metric

// 	// Binary version
// 	{
// 		buff, err := packet.MarshalBinary()
// 		assert.NoError(t, err)
// 		recorder := pingRelayBackendUpdate(t, "application/octet-stream", buff, updateMetrics, nil, nil, nil, nil)
// 		relayUpdateErrorAssertions(t, recorder, http.StatusBadRequest, metric)
// 	}

// 	// JSON version
// 	{
// 		buff, err := packet.MarshalJSON()
// 		assert.NoError(t, err)
// 		recorder := pingRelayBackendUpdate(t, "application/json", buff, updateMetrics, nil, nil, nil, nil)
// 		relayUpdateErrorAssertions(t, recorder, http.StatusBadRequest, metric)
// 	}
// }

// func TestGhostRelayIgnore(t *testing.T) {
// 	addr := "127.0.0.1:40000"
// 	udp, err := net.ResolveUDPAddr("udp", addr)
// 	assert.NoError(t, err)

// 	packet := transport.RelayUpdateRequest{
// 		Address:   *udp,
// 		Token:     make([]byte, crypto.KeySize),
// 		PingStats: make([]routing.RelayStatsPing, 0),
// 	}

// 	updateMetrics := metrics.EmptyRelayUpdateMetrics
// 	localMetrics := metrics.LocalHandler{}

// 	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
// 	assert.NoError(t, err)

// 	updateMetrics.ErrorMetrics.RelayNotFound = metric

// 	// No InMemory provided, no storage, so 404
// 	// Binary version
// 	{
// 		buff, err := packet.MarshalBinary()
// 		assert.NoError(t, err)
// 		recorder := pingRelayBackendUpdate(t, "application/octet-stream", buff, updateMetrics, nil, nil, nil, nil)
// 		relayUpdateErrorAssertions(t, recorder, http.StatusNotFound, metric)
// 	}

// 	// JSON version
// 	{
// 		buff, err := packet.MarshalJSON()
// 		assert.NoError(t, err)
// 		recorder := pingRelayBackendUpdate(t, "application/json", buff, updateMetrics, nil, nil, nil, nil)
// 		relayUpdateErrorAssertions(t, recorder, http.StatusNotFound, metric)
// 	}
// }

// func TestRelayUpdateRedisFailure(t *testing.T) {
// 	redisClient := redis.NewClient(&redis.Options{Addr: "0.0.0.0"})

// 	udp, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
// 	assert.NoError(t, err)
// 	packet := transport.RelayUpdateRequest{
// 		Address: *udp,
// 		Token:   make([]byte, crypto.KeySize),
// 	}

// 	updateMetrics := metrics.EmptyRelayUpdateMetrics
// 	localMetrics := metrics.LocalHandler{}

// 	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
// 	assert.NoError(t, err)

// 	updateMetrics.ErrorMetrics.RedisFailure = metric

// 	// add a relay to storage to pass the ghost checks in RelayUpdateHandlerFunc
// 	relay := routing.Relay{
// 		ID:   crypto.HashID("127.0.0.1:40000"),
// 		Addr: *udp,
// 		Seller: routing.Seller{
// 			ID:   "sellerID",
// 			Name: "seller name",
// 		},
// 		Datacenter: routing.Datacenter{
// 			ID:   1,
// 			Name: "some name",
// 		},
// 		PublicKey: make([]byte, crypto.KeySize),
// 		State:     routing.RelayStateEnabled,
// 	}

// 	inMemory := &storage.InMemory{}
// 	err = inMemory.AddSeller(context.Background(), relay.Seller)
// 	assert.NoError(t, err)
// 	err = inMemory.AddDatacenter(context.Background(), relay.Datacenter)
// 	assert.NoError(t, err)
// 	err = inMemory.AddRelay(context.Background(), relay)
// 	assert.NoError(t, err)

// 	// Binary version
// 	{
// 		buff, err := packet.MarshalBinary()
// 		assert.NoError(t, err)
// 		recorder := pingRelayBackendUpdate(t, "application/octet-stream", buff, updateMetrics, redisClient, nil, nil, inMemory)
// 		relayUpdateErrorAssertions(t, recorder, http.StatusInternalServerError, metric)
// 	}

// 	// JSON version
// 	{
// 		buff, err := packet.MarshalJSON()
// 		assert.NoError(t, err)
// 		recorder := pingRelayBackendUpdate(t, "application/json", buff, updateMetrics, redisClient, nil, nil, inMemory)
// 		relayUpdateErrorAssertions(t, recorder, http.StatusInternalServerError, metric)
// 	}
// }

// func TestRelayUpdateRelayNotFound(t *testing.T) {
// 	udp, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
// 	assert.NoError(t, err)
// 	packet := transport.RelayUpdateRequest{
// 		Address:   *udp,
// 		Token:     make([]byte, crypto.KeySize),
// 		PingStats: make([]routing.RelayStatsPing, 3),
// 	}

// 	updateMetrics := metrics.EmptyRelayUpdateMetrics
// 	localMetrics := metrics.LocalHandler{}

// 	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
// 	assert.NoError(t, err)

// 	updateMetrics.ErrorMetrics.RelayNotFound = metric

// 	// Binary version
// 	{
// 		buff, err := packet.MarshalBinary()
// 		assert.NoError(t, err)
// 		recorder := pingRelayBackendUpdate(t, "application/octet-stream", buff, updateMetrics, nil, nil, nil, nil)
// 		relayUpdateErrorAssertions(t, recorder, http.StatusNotFound, metric)
// 	}

// 	// JSON version
// 	{
// 		buff, err := packet.MarshalJSON()
// 		assert.NoError(t, err)
// 		recorder := pingRelayBackendUpdate(t, "application/json", buff, updateMetrics, nil, nil, nil, nil)
// 		relayUpdateErrorAssertions(t, recorder, http.StatusNotFound, metric)
// 	}
// }

// func TestRelayShuttingDown(t *testing.T) {
// 	redisServer, err := miniredis.Run()
// 	assert.NoError(t, err)
// 	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
// 	addr := "127.0.0.1:40000"
// 	udp, err := net.ResolveUDPAddr("udp", addr)
// 	assert.NoError(t, err)
// 	statsdb := routing.NewStatsDatabase()
// 	statIps := []string{"127.0.0.2:40000", "127.0.0.3:40000", "127.0.0.4:40000", "127.0.0.5:40000"}
// 	packet := transport.RelayUpdateRequest{
// 		Address:      *udp,
// 		Token:        make([]byte, crypto.KeySize),
// 		ShuttingDown: true,
// 	}

// 	packet.PingStats = make([]routing.RelayStatsPing, len(statIps))
// 	for i, addr := range statIps {
// 		stats := &packet.PingStats[i]
// 		stats.RelayID = crypto.HashID(addr)
// 		stats.RTT = rand.Float32()
// 		stats.Jitter = rand.Float32()
// 		stats.PacketLoss = rand.Float32()
// 	}

// 	seedRedis(t, redisServer, statIps)

// 	entry := routing.RelayCacheEntry{
// 		ID:   crypto.HashID(addr),
// 		Addr: *udp,
// 		Datacenter: routing.Datacenter{
// 			ID:   1,
// 			Name: "some name",
// 		},
// 		PublicKey:      make([]byte, crypto.KeySize),
// 		LastUpdateTime: time.Now().Add(-time.Second),
// 	}

// 	raw, err := entry.MarshalBinary()
// 	assert.NoError(t, err)
// 	redisServer.Set(entry.Key(), "0")
// 	redisServer.HSet(routing.HashKeyAllRelays, entry.Key(), string(raw))

// 	geoClient := &routing.GeoClient{
// 		RedisClient: redisClient,
// 	}

// 	relay := routing.Relay{
// 		ID:   entry.ID,
// 		Addr: entry.Addr,
// 		Seller: routing.Seller{
// 			ID:   "sellerID",
// 			Name: "seller name",
// 		},
// 		Datacenter: entry.Datacenter,
// 		PublicKey:  entry.PublicKey,
// 		State:      routing.RelayStateEnabled,
// 	}

// 	inMemory := &storage.InMemory{}

// 	err = inMemory.AddSeller(context.Background(), relay.Seller)
// 	assert.NoError(t, err)
// 	err = inMemory.AddDatacenter(context.Background(), relay.Datacenter)
// 	assert.NoError(t, err)
// 	err = inMemory.AddRelay(context.Background(), relay)
// 	assert.NoError(t, err)

// 	seedStorage(t, inMemory, statIps)

// 	localMetrics := metrics.LocalHandler{}
// 	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
// 	assert.NoError(t, err)

// 	updateMetrics := metrics.RelayUpdateMetrics{
// 		Invocations:   &metrics.EmptyCounter{},
// 		DurationGauge: &metrics.EmptyGauge{},
// 	}
// 	v := reflect.ValueOf(&updateMetrics.ErrorMetrics).Elem()
// 	for i := 0; i < v.NumField(); i++ {
// 		if v.Field(i).CanSet() {
// 			v.Field(i).Set(reflect.ValueOf(metric))
// 		}
// 	}

// 	// Binary version
// 	{
// 		buff, err := packet.MarshalBinary()
// 		assert.NoError(t, err)
// 		pingRelayBackendUpdate(t, "application/octet-stream", buff, updateMetrics, redisClient, geoClient, statsdb, inMemory)
// 		relayUpdateShutdownAssertions(t, updateMetrics.ErrorMetrics, entry, redisClient, inMemory, statsdb, addr)
// 	}

// 	// Check that the relay cache entry is gone
// 	_, err = redisServer.Get(entry.Key())
// 	assert.Error(t, err)

// 	// Add the relay cache entry back for the JSON version
// 	redisServer.Set(entry.Key(), "0")
// 	redisServer.HSet(routing.HashKeyAllRelays, entry.Key(), string(raw))

// 	// JSON version
// 	{
// 		buff, err := packet.MarshalJSON()
// 		assert.NoError(t, err)
// 		pingRelayBackendUpdate(t, "application/json", buff, updateMetrics, redisClient, geoClient, statsdb, inMemory)
// 		relayUpdateShutdownAssertions(t, updateMetrics.ErrorMetrics, entry, redisClient, inMemory, statsdb, addr)
// 	}

// 	// Check that the relay cache entry is gone
// 	_, err = redisServer.Get(entry.Key())
// 	assert.Error(t, err)
// }

// func TestRelayUpdateRelayUnmarshalFailure(t *testing.T) {
// 	redisServer, err := miniredis.Run()
// 	assert.NoError(t, err)
// 	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
// 	addr := "127.0.0.1:40000"
// 	udp, err := net.ResolveUDPAddr("udp", addr)
// 	assert.NoError(t, err)

// 	packet := transport.RelayUpdateRequest{
// 		Address:   *udp,
// 		Token:     make([]byte, crypto.KeySize),
// 		PingStats: make([]routing.RelayStatsPing, 0),
// 	}

// 	storedToken := make([]byte, crypto.KeySize)
// 	entry := routing.RelayCacheEntry{
// 		ID:   crypto.HashID(addr),
// 		Addr: *udp,
// 		Datacenter: routing.Datacenter{
// 			ID:   1,
// 			Name: "some name",
// 		},
// 		PublicKey:      storedToken,
// 		LastUpdateTime: time.Now().Add(-time.Second)}

// 	redisServer.Set(entry.Key(), "0")
// 	redisServer.HSet(routing.HashKeyAllRelays, entry.Key(), "invalid relay data")

// 	geoClient := &routing.GeoClient{
// 		RedisClient: redisClient,
// 	}

// 	updateMetrics := metrics.EmptyRelayUpdateMetrics
// 	localMetrics := metrics.LocalHandler{}

// 	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
// 	assert.NoError(t, err)

// 	updateMetrics.ErrorMetrics.RelayUnmarshalFailure = metric

// 	// add a relay to storage to pass the ghost checks in RelayUpdateHandlerFunc
// 	relay := routing.Relay{
// 		ID:   entry.ID,
// 		Addr: entry.Addr,
// 		Seller: routing.Seller{
// 			ID:   "sellerID",
// 			Name: "seller name",
// 		},
// 		Datacenter: entry.Datacenter,
// 		PublicKey:  entry.PublicKey,
// 		State:      routing.RelayStateEnabled,
// 	}

// 	inMemory := &storage.InMemory{}
// 	err = inMemory.AddSeller(context.Background(), relay.Seller)
// 	assert.NoError(t, err)
// 	err = inMemory.AddDatacenter(context.Background(), relay.Datacenter)
// 	assert.NoError(t, err)
// 	err = inMemory.AddRelay(context.Background(), relay)
// 	assert.NoError(t, err)

// 	// Binary version
// 	{
// 		buff, err := packet.MarshalBinary()
// 		assert.NoError(t, err)
// 		recorder := pingRelayBackendUpdate(t, "application/octet-stream", buff, updateMetrics, redisClient, geoClient, nil, inMemory)
// 		relayUpdateErrorAssertions(t, recorder, http.StatusInternalServerError, metric)
// 	}

// 	// JSON version
// 	{
// 		buff, err := packet.MarshalJSON()
// 		assert.NoError(t, err)
// 		recorder := pingRelayBackendUpdate(t, "application/json", buff, updateMetrics, redisClient, geoClient, nil, inMemory)
// 		relayUpdateErrorAssertions(t, recorder, http.StatusInternalServerError, metric)
// 	}
// }

// func TestRelayUpdateInvalidToken(t *testing.T) {
// 	redisServer, err := miniredis.Run()
// 	assert.NoError(t, err)
// 	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
// 	addr := "127.0.0.1:40000"
// 	udp, err := net.ResolveUDPAddr("udp", addr)
// 	assert.NoError(t, err)

// 	incomingToken := make([]byte, crypto.KeySize)
// 	rand.Read(incomingToken)
// 	storedToken := make([]byte, crypto.KeySize)
// 	rand.Read(storedToken)
// 	packet := transport.RelayUpdateRequest{
// 		Address:   *udp,
// 		Token:     incomingToken,
// 		PingStats: make([]routing.RelayStatsPing, 0),
// 	}

// 	entry := routing.RelayCacheEntry{
// 		ID:   crypto.HashID(addr),
// 		Addr: *udp,
// 		Datacenter: routing.Datacenter{
// 			ID:   1,
// 			Name: "some name",
// 		},
// 		PublicKey:      storedToken,
// 		LastUpdateTime: time.Now().Add(-time.Second),
// 	}

// 	raw, err := entry.MarshalBinary()
// 	assert.NoError(t, err)
// 	redisServer.Set(entry.Key(), "0")
// 	redisServer.HSet(routing.HashKeyAllRelays, entry.Key(), string(raw))

// 	geoClient := &routing.GeoClient{
// 		RedisClient: redisClient,
// 	}

// 	updateMetrics := metrics.EmptyRelayUpdateMetrics
// 	localMetrics := metrics.LocalHandler{}

// 	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
// 	assert.NoError(t, err)

// 	updateMetrics.ErrorMetrics.InvalidToken = metric

// 	// add a relay to storage to pass the ghost checks in RelayUpdateHandlerFunc
// 	relay := routing.Relay{
// 		ID:   entry.ID,
// 		Addr: entry.Addr,
// 		Seller: routing.Seller{
// 			ID:   "sellerID",
// 			Name: "seller name",
// 		},
// 		Datacenter: entry.Datacenter,
// 		PublicKey:  entry.PublicKey,
// 		State:      routing.RelayStateEnabled,
// 	}

// 	inMemory := &storage.InMemory{}
// 	err = inMemory.AddSeller(context.Background(), relay.Seller)
// 	assert.NoError(t, err)
// 	err = inMemory.AddDatacenter(context.Background(), relay.Datacenter)
// 	assert.NoError(t, err)
// 	err = inMemory.AddRelay(context.Background(), relay)
// 	assert.NoError(t, err)

// 	// Binary version
// 	{
// 		buff, err := packet.MarshalBinary()
// 		assert.NoError(t, err)
// 		recorder := pingRelayBackendUpdate(t, "application/octet-stream", buff, updateMetrics, redisClient, geoClient, nil, inMemory)
// 		relayUpdateErrorAssertions(t, recorder, http.StatusBadRequest, metric)
// 	}

// 	// JSON version
// 	{
// 		buff, err := packet.MarshalJSON()
// 		assert.NoError(t, err)
// 		recorder := pingRelayBackendUpdate(t, "application/json", buff, updateMetrics, redisClient, geoClient, nil, inMemory)
// 		relayUpdateErrorAssertions(t, recorder, http.StatusBadRequest, metric)
// 	}
// }

// func TestRelayUpdateInvalidState(t *testing.T) {
// 	redisServer, err := miniredis.Run()
// 	assert.NoError(t, err)
// 	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
// 	addr := "127.0.0.1:40000"
// 	udp, err := net.ResolveUDPAddr("udp", addr)
// 	assert.NoError(t, err)
// 	statsdb := routing.NewStatsDatabase()
// 	statIps := []string{"127.0.0.2:40000", "127.0.0.3:40000", "127.0.0.4:40000", "127.0.0.5:40000"}
// 	packet := transport.RelayUpdateRequest{
// 		Address: *udp,
// 		Token:   make([]byte, crypto.KeySize),
// 	}

// 	packet.PingStats = make([]routing.RelayStatsPing, len(statIps))
// 	for i, addr := range statIps {
// 		stats := &packet.PingStats[i]
// 		stats.RelayID = crypto.HashID(addr)
// 		stats.RTT = rand.Float32()
// 		stats.Jitter = rand.Float32()
// 		stats.PacketLoss = rand.Float32()
// 	}

// 	seedRedis(t, redisServer, statIps)

// 	entry := routing.RelayCacheEntry{
// 		ID:   crypto.HashID(addr),
// 		Addr: *udp,
// 		Datacenter: routing.Datacenter{
// 			ID:   1,
// 			Name: "some name",
// 		},
// 		PublicKey:      make([]byte, crypto.KeySize),
// 		LastUpdateTime: time.Now().Add(-time.Second),
// 	}

// 	raw, err := entry.MarshalBinary()
// 	assert.NoError(t, err)
// 	redisServer.Set(entry.Key(), "0")
// 	redisServer.HSet(routing.HashKeyAllRelays, entry.Key(), string(raw))

// 	geoClient := &routing.GeoClient{
// 		RedisClient: redisClient,
// 	}

// 	relay := routing.Relay{
// 		ID:   entry.ID,
// 		Addr: entry.Addr,
// 		Seller: routing.Seller{
// 			ID:   "sellerID",
// 			Name: "seller name",
// 		},
// 		Datacenter: entry.Datacenter,
// 		PublicKey:  entry.PublicKey,
// 		State:      routing.RelayStateQuarantine,
// 	}

// 	inMemory := &storage.InMemory{}

// 	err = inMemory.AddSeller(context.Background(), relay.Seller)
// 	assert.NoError(t, err)
// 	err = inMemory.AddDatacenter(context.Background(), relay.Datacenter)
// 	assert.NoError(t, err)
// 	err = inMemory.AddRelay(context.Background(), relay)
// 	assert.NoError(t, err)

// 	seedStorage(t, inMemory, statIps)

// 	updateMetrics := metrics.EmptyRelayUpdateMetrics
// 	localMetrics := metrics.LocalHandler{}

// 	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
// 	assert.NoError(t, err)

// 	updateMetrics.ErrorMetrics.RelayNotEnabled = metric

// 	// Binary version
// 	{
// 		buff, err := packet.MarshalBinary()
// 		assert.NoError(t, err)
// 		recorder := pingRelayBackendUpdate(t, "application/octet-stream", buff, updateMetrics, redisClient, geoClient, statsdb, inMemory)
// 		relayUpdateErrorAssertions(t, recorder, http.StatusUnauthorized, metric)
// 	}

// 	// JSON version
// 	{
// 		buff, err := packet.MarshalJSON()
// 		assert.NoError(t, err)
// 		recorder := pingRelayBackendUpdate(t, "application/json", buff, updateMetrics, redisClient, geoClient, statsdb, inMemory)
// 		relayUpdateErrorAssertions(t, recorder, http.StatusUnauthorized, metric)
// 	}
// }

// func TestRelayUpdateSuccess(t *testing.T) {
// 	redisServer, err := miniredis.Run()
// 	assert.NoError(t, err)
// 	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})
// 	addr := "127.0.0.1:40000"
// 	udp, err := net.ResolveUDPAddr("udp", addr)
// 	assert.NoError(t, err)
// 	statsdb := routing.NewStatsDatabase()
// 	statIps := []string{"127.0.0.2:40000", "127.0.0.3:40000", "127.0.0.4:40000", "127.0.0.5:40000"}
// 	packet := transport.RelayUpdateRequest{
// 		Address: *udp,
// 		Token:   make([]byte, crypto.KeySize),
// 	}

// 	packet.PingStats = make([]routing.RelayStatsPing, len(statIps))
// 	for i, addr := range statIps {
// 		stats := &packet.PingStats[i]
// 		stats.RelayID = crypto.HashID(addr)
// 		stats.RTT = rand.Float32()
// 		stats.Jitter = rand.Float32()
// 		stats.PacketLoss = rand.Float32()
// 	}

// 	seedRedis(t, redisServer, statIps)

// 	entry := routing.RelayCacheEntry{
// 		ID:   crypto.HashID(addr),
// 		Addr: *udp,
// 		Datacenter: routing.Datacenter{
// 			ID:   1,
// 			Name: "some name",
// 		},
// 		PublicKey:      make([]byte, crypto.KeySize),
// 		LastUpdateTime: time.Now().Add(-time.Second),
// 	}

// 	raw, err := entry.MarshalBinary()
// 	assert.NoError(t, err)
// 	redisServer.Set(entry.Key(), "0")
// 	redisServer.HSet(routing.HashKeyAllRelays, entry.Key(), string(raw))

// 	geoClient := &routing.GeoClient{
// 		RedisClient: redisClient,
// 	}

// 	relay := routing.Relay{
// 		ID:   entry.ID,
// 		Addr: entry.Addr,
// 		Seller: routing.Seller{
// 			ID:   "sellerID",
// 			Name: "seller name",
// 		},
// 		Datacenter: entry.Datacenter,
// 		PublicKey:  entry.PublicKey,
// 		State:      routing.RelayStateEnabled,
// 	}

// 	inMemory := &storage.InMemory{}

// 	err = inMemory.AddSeller(context.Background(), relay.Seller)
// 	assert.NoError(t, err)
// 	err = inMemory.AddDatacenter(context.Background(), relay.Datacenter)
// 	assert.NoError(t, err)
// 	err = inMemory.AddRelay(context.Background(), relay)
// 	assert.NoError(t, err)

// 	seedStorage(t, inMemory, statIps)

// 	localMetrics := metrics.LocalHandler{}
// 	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
// 	assert.NoError(t, err)

// 	updateMetrics := metrics.RelayUpdateMetrics{
// 		Invocations:   &metrics.EmptyCounter{},
// 		DurationGauge: &metrics.EmptyGauge{},
// 	}
// 	v := reflect.ValueOf(&updateMetrics.ErrorMetrics).Elem()
// 	for i := 0; i < v.NumField(); i++ {
// 		if v.Field(i).CanSet() {
// 			v.Field(i).Set(reflect.ValueOf(metric))
// 		}
// 	}

// 	// Binary version
// 	{
// 		buff, err := packet.MarshalBinary()
// 		assert.NoError(t, err)
// 		recorder := pingRelayBackendUpdate(t, "application/octet-stream", buff, updateMetrics, redisClient, geoClient, statsdb, inMemory)
// 		relayUpdateSuccessAssertions(t, recorder, "application/octet-stream", updateMetrics.ErrorMetrics, entry, redisClient, inMemory, statsdb, statIps, addr)
// 	}

// 	// JSON version
// 	{
// 		buff, err := packet.MarshalJSON()
// 		assert.NoError(t, err)
// 		recorder := pingRelayBackendUpdate(t, "application/json", buff, updateMetrics, redisClient, geoClient, statsdb, inMemory)
// 		relayUpdateSuccessAssertions(t, recorder, "application/json", updateMetrics.ErrorMetrics, entry, redisClient, inMemory, statsdb, statIps, addr)
// 	}
// }
