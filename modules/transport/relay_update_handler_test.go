package transport_test

/* import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/storage"
	"github.com/networknext/backend/modules/transport"
	"github.com/stretchr/testify/assert"
)

func testRelayUpdateHandlerConfig(RelayMap *routing.RelayMap, statsdb *routing.StatsDatabase, storer storage.Storer, metrics metrics.RelayUpdateMetrics, noInit bool) *transport.RelayUpdateHandlerConfig {

	if RelayMap == nil {
		RelayMap = routing.NewRelayMap(func(relay *routing.RelayData) error {
			return nil
		})
	}

	return &transport.RelayUpdateHandlerConfig{
		RelayMap: RelayMap,
		StatsDB:  statsdb,
		Metrics:  &metrics,
		Storer:   storer,
	}
}

func pingRelayBackendUpdate(t *testing.T, contentType string, body []byte, handlerConfig *transport.RelayUpdateHandlerConfig) *httptest.ResponseRecorder {

	recorder := httptest.NewRecorder()
	request, err := http.NewRequest("POST", "/relay_update", bytes.NewBuffer(body))
	assert.NoError(t, err)
	request.Header.Add("Content-Type", contentType)

	handler := transport.RelayUpdateHandlerFunc(log.NewNopLogger(), log.NewNopLogger(), handlerConfig)

	handler(recorder, request)
	return recorder
}

func relayUpdateErrorAssertions(t *testing.T, recorder *httptest.ResponseRecorder, expectedCode int, errMetric metrics.Counter) {
	assert.Equal(t, expectedCode, recorder.Code)
	assert.Equal(t, 1.0, errMetric.ValueReset())
}

func relayUpdateShutdownAssertions(t *testing.T, recorder *httptest.ResponseRecorder, handlerConfig *transport.RelayUpdateHandlerConfig, addr string) {
	relayData := handlerConfig.RelayMap.GetRelayData(addr)
	assert.Nil(t, relayData)

	if recorder.Code != 200 {
		body, err := ioutil.ReadAll(recorder.Body)
		assert.Nil(t, err)
		fmt.Println(string(body))
	}

	relay, err := handlerConfig.Storer.Relay(crypto.HashID(addr))
	assert.NoError(t, err)
	fmt.Println(relay.State)
	assert.Equal(t, routing.RelayStateMaintenance, relay.State)

	for i, stat := range handlerConfig.StatsDB.Entries {
		assert.NotEqual(t, i, relay.ID)
		for j := range stat.Relays {
			assert.NotEqual(t, j, relay.ID)
		}
	}

	errMetricsStruct := reflect.ValueOf(handlerConfig.Metrics.ErrorMetrics)
	for i := 0; i < errMetricsStruct.NumField(); i++ {
		if errMetricsStruct.Field(i).CanInterface() {
			assert.Equal(t, 0.0, errMetricsStruct.Field(i).Interface().(metrics.Counter).ValueReset())
		}
	}
}

func relayUpdateSuccessAssertions(t *testing.T, recorder *httptest.ResponseRecorder, expectedContentType string, entry *routing.RelayData, handlerConfig *transport.RelayUpdateHandlerConfig, statIps []string, addr string) {
	assert.Equal(t, http.StatusOK, recorder.Code)

	numRelays := handlerConfig.RelayMap.GetRelayCount()
	assert.Equal(t, uint64(5), numRelays)

	actual := handlerConfig.RelayMap.GetRelayData(addr)
	assert.NotNil(t, actual)

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

	assert.Contains(t, handlerConfig.StatsDB.Entries, entry.ID)
	relations := handlerConfig.StatsDB.Entries[entry.ID]
	for _, addr := range statIps {
		id := crypto.HashID(addr)
		assert.Contains(t, relaysToPingIDs, id)
		assert.Contains(t, relaysToPingAddrs, addr)
		assert.Contains(t, relations.Relays, id)
	}

	assert.NotContains(t, relaysToPingIDs, entry.ID)
	assert.NotContains(t, relaysToPingAddrs, addr)

	relay, err := handlerConfig.Storer.Relay(crypto.HashID(addr))
	assert.NoError(t, err)

	assert.Equal(t, routing.RelayStateEnabled, relay.State)

	errMetricsStruct := reflect.ValueOf(handlerConfig.Metrics.ErrorMetrics)
	for i := 0; i < errMetricsStruct.NumField(); i++ {
		if errMetricsStruct.Field(i).CanInterface() {
			assert.Equal(t, 0.0, errMetricsStruct.Field(i).Interface().(metrics.Counter).ValueReset())
		}
	}
}

func TestRelayUpdateUnmarshalFailure(t *testing.T) {
	updateMetrics := metrics.EmptyRelayUpdateMetrics
	localMetrics := metrics.LocalHandler{}

	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
	assert.NoError(t, err)

	updateMetrics.ErrorMetrics.UnmarshalFailure = metric

	handlerConfig := testRelayUpdateHandlerConfig(nil, nil, &storage.StorerMock{}, updateMetrics, false)

	buff := make([]byte, 10) // invalid relay packet size
	recorder := pingRelayBackendUpdate(t, "application/octet-stream", buff, handlerConfig)
	relayUpdateErrorAssertions(t, recorder, http.StatusBadRequest, metric)
}

func TestRelayUpdateInvalidAddress(t *testing.T) {
	udp, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
	assert.NoError(t, err)
	packet := transport.RelayUpdateRequest{
		Address: *udp,
		Token:   make([]byte, crypto.KeySize),
	}

	updateMetrics := metrics.EmptyRelayUpdateMetrics
	localMetrics := metrics.LocalHandler{}

	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
	assert.NoError(t, err)

	updateMetrics.ErrorMetrics.UnmarshalFailure = metric

	buff, err := packet.MarshalBinary()
	assert.NoError(t, err)
	badAddr := "invalid address"        // "invalid address" is luckily the same number of characters as "127.0.0.1:40000"
	for i := 0; i < len(badAddr); i++ { // Replace the address with the bad address character by character
		buff[8+i] = badAddr[i]
	}
	handlerConfig := testRelayUpdateHandlerConfig(nil, nil, &storage.StorerMock{}, updateMetrics, false)

	recorder := pingRelayBackendUpdate(t, "application/octet-stream", buff, handlerConfig)
	relayUpdateErrorAssertions(t, recorder, http.StatusBadRequest, metric)
}

func TestRelayUpdateExceedMaxRelays(t *testing.T) {
	udp, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
	assert.NoError(t, err)
	packet := transport.RelayUpdateRequest{
		Address:   *udp,
		Token:     make([]byte, crypto.KeySize),
		PingStats: make([]routing.RelayStatsPing, 1025),
	}

	updateMetrics := metrics.EmptyRelayUpdateMetrics
	localMetrics := metrics.LocalHandler{}

	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
	assert.NoError(t, err)

	updateMetrics.ErrorMetrics.ExceedMaxRelays = metric

	buff, err := packet.MarshalBinary()
	assert.NoError(t, err)
	handlerConfig := testRelayUpdateHandlerConfig(nil, nil, &storage.StorerMock{}, updateMetrics, false)
	recorder := pingRelayBackendUpdate(t, "application/octet-stream", buff, handlerConfig)
	relayUpdateErrorAssertions(t, recorder, http.StatusBadRequest, metric)

}

func TestRelayUpdateRelayNotFoundInRelayMapNoInitOff(t *testing.T) {
	addr := "127.0.0.1:40000"
	udp, err := net.ResolveUDPAddr("udp", addr)
	assert.NoError(t, err)
	packet := transport.RelayUpdateRequest{
		Address:   *udp,
		Token:     make([]byte, crypto.KeySize),
		PingStats: make([]routing.RelayStatsPing, 3),
	}

	updateMetrics := metrics.EmptyRelayUpdateMetrics
	localMetrics := metrics.LocalHandler{}

	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
	assert.NoError(t, err)

	updateMetrics.ErrorMetrics.RelayNotFound = metric

	relay := routing.Relay{
		ID:   crypto.HashID(addr),
		Addr: *udp,
		Datacenter: routing.Datacenter{
			ID:   1,
			Name: "some name",
		},
		PublicKey: make([]byte, crypto.KeySize),
		Seller: routing.Seller{
			ID:   "sellerID",
			Name: "seller name",
		},
		State: routing.RelayStateEnabled,
	}

	inMemory := &storage.InMemory{}

	err = inMemory.AddSeller(context.Background(), relay.Seller)
	assert.NoError(t, err)
	err = inMemory.AddDatacenter(context.Background(), relay.Datacenter)
	assert.NoError(t, err)
	err = inMemory.AddRelay(context.Background(), relay)
	assert.NoError(t, err)

	buff, err := packet.MarshalBinary()
	assert.NoError(t, err)
	handlerConfig := testRelayUpdateHandlerConfig(nil, nil, inMemory, updateMetrics, false)
	recorder := pingRelayBackendUpdate(t, "application/octet-stream", buff, handlerConfig)
	relayUpdateErrorAssertions(t, recorder, http.StatusNotFound, metric)

}

func TestGhostRelayIgnore(t *testing.T) {
	addr := "127.0.0.1:40000"
	udp, err := net.ResolveUDPAddr("udp", addr)
	assert.NoError(t, err)

	packet := transport.RelayUpdateRequest{
		Address:   *udp,
		Token:     make([]byte, crypto.KeySize),
		PingStats: make([]routing.RelayStatsPing, 0),
	}

	updateMetrics := metrics.EmptyRelayUpdateMetrics
	localMetrics := metrics.LocalHandler{}

	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
	assert.NoError(t, err)

	updateMetrics.ErrorMetrics.RelayNotFound = metric
	inMemory := &storage.InMemory{}

	buff, err := packet.MarshalBinary()
	assert.NoError(t, err)
	handlerConfig := testRelayUpdateHandlerConfig(nil, nil, inMemory, updateMetrics, false)
	recorder := pingRelayBackendUpdate(t, "application/octet-stream", buff, handlerConfig)
	relayUpdateErrorAssertions(t, recorder, http.StatusNotFound, metric)
}

func TestRelayShuttingDown(t *testing.T) {
	addr := "127.0.0.1:40000"
	udp, err := net.ResolveUDPAddr("udp", addr)
	assert.NoError(t, err)

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

	entry := &routing.RelayData{
		ID:   crypto.HashID(addr),
		Addr: *udp,
		Datacenter: routing.Datacenter{
			ID:   1,
			Name: "some name",
		},
		PublicKey:      make([]byte, crypto.KeySize),
		LastUpdateTime: time.Now().Add(-time.Second),
	}

	relay := routing.Relay{
		ID:   entry.ID,
		Addr: entry.Addr,
		Seller: routing.Seller{
			ID:   "sellerID",
			Name: "seller name",
		},
		Datacenter: entry.Datacenter,
		PublicKey:  entry.PublicKey,
		State:      routing.RelayStateEnabled,
	}

	inMemory := &storage.InMemory{}

	err = inMemory.AddSeller(context.Background(), relay.Seller)
	assert.NoError(t, err)
	err = inMemory.AddDatacenter(context.Background(), relay.Datacenter)
	assert.NoError(t, err)
	err = inMemory.AddRelay(context.Background(), relay)
	assert.NoError(t, err)

	seedStorage(t, inMemory, statIps)

	localMetrics := metrics.LocalHandler{}
	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
	assert.NoError(t, err)

	updateMetrics := metrics.RelayUpdateMetrics{
		Invocations:   &metrics.EmptyCounter{},
		DurationGauge: &metrics.EmptyGauge{},
	}
	v := reflect.ValueOf(&updateMetrics.ErrorMetrics).Elem()
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).CanSet() {
			v.Field(i).Set(reflect.ValueOf(metric))
		}
	}

	handlerConfig := testRelayUpdateHandlerConfig(nil, statsdb, inMemory, updateMetrics, false)
	handlerConfig.RelayMap.AddRelayDataEntry(addr, entry)

	buff, err := packet.MarshalBinary()
	assert.NoError(t, err)
	recorder := pingRelayBackendUpdate(t, "application/octet-stream", buff, handlerConfig)
	relayUpdateShutdownAssertions(t, recorder, handlerConfig, addr)

	// Check that the relay cache entry is gone
	relayData := handlerConfig.RelayMap.GetRelayData(addr)
	assert.Nil(t, relayData)

}

func TestRelayUpdateRelayUnmarshalFailure(t *testing.T) {

	addr := "127.0.0.1:40000"
	udp, err := net.ResolveUDPAddr("udp", addr)
	assert.NoError(t, err)

	packet := transport.RelayUpdateRequest{
		Address:   *udp,
		Token:     make([]byte, crypto.KeySize),
		PingStats: make([]routing.RelayStatsPing, 0),
	}

	storedToken := make([]byte, crypto.KeySize)
	entry := &routing.RelayData{
		ID:   crypto.HashID(addr),
		Addr: *udp,
		Datacenter: routing.Datacenter{
			ID:   1,
			Name: "some name",
		},
		PublicKey:      storedToken,
		LastUpdateTime: time.Now().Add(-time.Second),
	}

	updateMetrics := metrics.EmptyRelayUpdateMetrics
	localMetrics := metrics.LocalHandler{}

	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
	assert.NoError(t, err)

	updateMetrics.ErrorMetrics.UnmarshalFailure = metric

	// add a relay to storage to pass the ghost checks in RelayUpdateHandlerFunc
	relay := routing.Relay{
		ID:   entry.ID,
		Addr: entry.Addr,
		Seller: routing.Seller{
			ID:   "sellerID",
			Name: "seller name",
		},
		Datacenter: entry.Datacenter,
		PublicKey:  entry.PublicKey,
		State:      routing.RelayStateEnabled,
	}

	inMemory := &storage.InMemory{}
	err = inMemory.AddSeller(context.Background(), relay.Seller)
	assert.NoError(t, err)
	err = inMemory.AddDatacenter(context.Background(), relay.Datacenter)
	assert.NoError(t, err)
	err = inMemory.AddRelay(context.Background(), relay)
	assert.NoError(t, err)

	handlerConfig := testRelayUpdateHandlerConfig(nil, nil, inMemory, updateMetrics, false)
	handlerConfig.RelayMap.AddRelayDataEntry(addr, entry)
	buff, err := packet.MarshalBinary()
	buff[3] = 'a'
	assert.NoError(t, err)

	recorder := pingRelayBackendUpdate(t, "application/octet-stream", buff, handlerConfig)
	relayUpdateErrorAssertions(t, recorder, http.StatusBadRequest, metric)
}

func TestRelayUpdateInvalidToken(t *testing.T) {
	addr := "127.0.0.1:40000"
	udp, err := net.ResolveUDPAddr("udp", addr)
	assert.NoError(t, err)

	incomingToken := make([]byte, crypto.KeySize)
	rand.Read(incomingToken)
	storedToken := make([]byte, crypto.KeySize)
	rand.Read(storedToken)
	packet := transport.RelayUpdateRequest{
		Address:   *udp,
		Token:     incomingToken,
		PingStats: make([]routing.RelayStatsPing, 0),
	}

	entry := &routing.RelayData{
		ID:   crypto.HashID(addr),
		Addr: *udp,
		Datacenter: routing.Datacenter{
			ID:   1,
			Name: "some name",
		},
		PublicKey:      storedToken,
		LastUpdateTime: time.Now().Add(-time.Second),
	}

	updateMetrics := metrics.EmptyRelayUpdateMetrics
	localMetrics := metrics.LocalHandler{}

	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
	assert.NoError(t, err)

	updateMetrics.ErrorMetrics.InvalidToken = metric

	// add a relay to storage to pass the ghost checks in RelayUpdateHandlerFunc
	relay := routing.Relay{
		ID:   entry.ID,
		Addr: entry.Addr,
		Seller: routing.Seller{
			ID:   "sellerID",
			Name: "seller name",
		},
		Datacenter: entry.Datacenter,
		PublicKey:  entry.PublicKey,
		State:      routing.RelayStateEnabled,
	}

	inMemory := &storage.InMemory{}
	err = inMemory.AddSeller(context.Background(), relay.Seller)
	assert.NoError(t, err)
	err = inMemory.AddDatacenter(context.Background(), relay.Datacenter)
	assert.NoError(t, err)
	err = inMemory.AddRelay(context.Background(), relay)
	assert.NoError(t, err)

	handlerConfig := testRelayUpdateHandlerConfig(nil, nil, inMemory, updateMetrics, false)
	handlerConfig.RelayMap.AddRelayDataEntry(addr, entry)

	buff, err := packet.MarshalBinary()
	assert.NoError(t, err)
	recorder := pingRelayBackendUpdate(t, "application/octet-stream", buff, handlerConfig)
	relayUpdateErrorAssertions(t, recorder, http.StatusBadRequest, metric)

}

func TestRelayUpdateInvalidState(t *testing.T) {
	addr := "127.0.0.1:40000"
	udp, err := net.ResolveUDPAddr("udp", addr)
	assert.NoError(t, err)
	statsdb := routing.NewStatsDatabase()
	statIps := []string{"127.0.0.2:40000", "127.0.0.3:40000", "127.0.0.4:40000", "127.0.0.5:40000"}
	packet := transport.RelayUpdateRequest{
		Address: *udp,
		Token:   make([]byte, crypto.KeySize),
	}

	packet.PingStats = make([]routing.RelayStatsPing, len(statIps))
	for i, addr := range statIps {
		stats := &packet.PingStats[i]
		stats.RelayID = crypto.HashID(addr)
		stats.RTT = rand.Float32()
		stats.Jitter = rand.Float32()
		stats.PacketLoss = rand.Float32()
	}

	entry := &routing.RelayData{
		ID:   crypto.HashID(addr),
		Addr: *udp,
		Datacenter: routing.Datacenter{
			ID:   1,
			Name: "some name",
		},
		PublicKey:      make([]byte, crypto.KeySize),
		LastUpdateTime: time.Now().Add(-time.Second),
	}

	relay := routing.Relay{
		ID:   entry.ID,
		Addr: entry.Addr,
		Seller: routing.Seller{
			ID:   "sellerID",
			Name: "seller name",
		},
		Datacenter: entry.Datacenter,
		PublicKey:  entry.PublicKey,
		State:      routing.RelayStateQuarantine,
	}

	inMemory := &storage.InMemory{}

	err = inMemory.AddSeller(context.Background(), relay.Seller)
	assert.NoError(t, err)
	err = inMemory.AddDatacenter(context.Background(), relay.Datacenter)
	assert.NoError(t, err)
	err = inMemory.AddRelay(context.Background(), relay)
	assert.NoError(t, err)

	seedStorage(t, inMemory, statIps)

	updateMetrics := metrics.EmptyRelayUpdateMetrics
	localMetrics := metrics.LocalHandler{}

	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
	assert.NoError(t, err)

	updateMetrics.ErrorMetrics.RelayNotEnabled = metric

	handlerConfig := testRelayUpdateHandlerConfig(nil, statsdb, inMemory, updateMetrics, false)
	handlerConfig.RelayMap.AddRelayDataEntry(addr, entry)
	buff, err := packet.MarshalBinary()
	assert.NoError(t, err)
	recorder := pingRelayBackendUpdate(t, "application/octet-stream", buff, handlerConfig)
	relayUpdateErrorAssertions(t, recorder, http.StatusUnauthorized, metric)
}

func TestRelayUpdateSuccess(t *testing.T) {
	addr := "127.0.0.1:40000"
	udp, err := net.ResolveUDPAddr("udp", addr)
	assert.NoError(t, err)
	statsdb := routing.NewStatsDatabase()
	statIps := []string{"127.0.0.2:40000", "127.0.0.3:40000", "127.0.0.4:40000", "127.0.0.5:40000"}
	packet := transport.RelayUpdateRequest{
		Address: *udp,
		Token:   make([]byte, crypto.KeySize),
	}

	packet.PingStats = make([]routing.RelayStatsPing, len(statIps))
	for i, addr := range statIps {
		stats := &packet.PingStats[i]
		stats.RelayID = crypto.HashID(addr)
		stats.RTT = rand.Float32()
		stats.Jitter = rand.Float32()
		stats.PacketLoss = rand.Float32()
	}

	entry := &routing.RelayData{
		ID:   crypto.HashID(addr),
		Addr: *udp,
		Datacenter: routing.Datacenter{
			ID:   1,
			Name: "some name",
		},
		PublicKey:      make([]byte, crypto.KeySize),
		LastUpdateTime: time.Now().Add(-time.Second),
	}
	orgEntry := *entry

	relay := routing.Relay{
		ID:   entry.ID,
		Addr: entry.Addr,
		Seller: routing.Seller{
			ID:   "sellerID",
			Name: "seller name",
		},
		Datacenter: entry.Datacenter,
		PublicKey:  entry.PublicKey,
		State:      routing.RelayStateEnabled,
	}

	inMemory := &storage.InMemory{}
	err = inMemory.AddSeller(context.Background(), relay.Seller)
	assert.NoError(t, err)
	err = inMemory.AddDatacenter(context.Background(), relay.Datacenter)
	assert.NoError(t, err)
	err = inMemory.AddRelay(context.Background(), relay)
	assert.NoError(t, err)

	seedStorage(t, inMemory, statIps)

	localMetrics := metrics.LocalHandler{}
	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
	assert.NoError(t, err)

	updateMetrics := metrics.RelayUpdateMetrics{
		Invocations:   &metrics.EmptyCounter{},
		DurationGauge: &metrics.EmptyGauge{},
	}
	v := reflect.ValueOf(&updateMetrics.ErrorMetrics).Elem()
	for i := 0; i < v.NumField(); i++ {
		if v.Field(i).CanSet() {
			v.Field(i).Set(reflect.ValueOf(metric))
		}
	}

	handlerConfig := testRelayUpdateHandlerConfig(nil, statsdb, inMemory, updateMetrics, false)
	handlerConfig.RelayMap.AddRelayDataEntry(addr, entry)
	seedRelayMap(t, handlerConfig.RelayMap, statIps)

	buff, err := packet.MarshalBinary()
	assert.NoError(t, err)
	recorder := pingRelayBackendUpdate(t, "application/octet-stream", buff, handlerConfig)
	relayUpdateSuccessAssertions(t, recorder, "application/octet-stream", &orgEntry, handlerConfig, statIps, addr)
}
*/
