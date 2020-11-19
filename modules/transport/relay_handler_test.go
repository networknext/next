package transport_test

// todo: come back and fix these tests

// import (
// 	"bytes"
// 	"context"
// 	"crypto/rand"
// 	"encoding/base64"
// 	"fmt"
// 	"math"
// 	mrand "math/rand"
// 	"net"
// 	"net/http"
// 	"net/http/httptest"
// 	"reflect"
// 	"strconv"
// 	"testing"
// 	"time"

// 	"github.com/go-kit/kit/log"
// 	"github.com/networknext/backend/crypto"
// 	"github.com/networknext/backend/metrics"
// 	"github.com/networknext/backend/routing"
// 	"github.com/networknext/backend/storage"
// 	"github.com/networknext/backend/transport"
// 	"github.com/stretchr/testify/assert"
// 	"golang.org/x/crypto/nacl/box"
// )

// func pingRelayBackendHandler(t *testing.T, headers map[string]string, body []byte, metrics metrics.RelayHandlerMetrics, inMemory *storage.InMemory, statsdb *routing.StatsDatabase, routerPrivateKey []byte) *httptest.ResponseRecorder {
// 	if statsdb == nil {
// 		statsdb = routing.NewStatsDatabase()
// 	}

// 	recorder := httptest.NewRecorder()
// 	request, err := http.NewRequest("POST", "/relays", bytes.NewBuffer(body))
// 	assert.NoError(t, err)

// 	request.Header.Add("Content-Type", "application/json")
// 	for key, val := range headers {
// 		request.Header.Add(key, val)
// 	}

// 	cleanupCallback := func(relayID uint64) error {
// 		statsdb.DeleteEntry(relayID)
// 		return nil
// 	}

// 	handler := transport.RelayHandlerFunc(log.NewNopLogger(), log.NewNopLogger(), &transport.RelayHandlerConfig{
// 		RelayMap:         routing.NewRelayMap(cleanupCallback),
// 		Storer:           inMemory,
// 		StatsDb:          statsdb,
// 		Metrics:          &metrics,
// 		RouterPrivateKey: routerPrivateKey,
// 	})

// 	handler(recorder, request)
// 	return recorder
// }

// func relayHandlerErrorAssertions(t *testing.T, recorder *httptest.ResponseRecorder, expectedCode int, errMetric metrics.Counter) {
// 	assert.Equal(t, expectedCode, recorder.Code)
// 	assert.Equal(t, 1.0, errMetric.Value())
// }

// func relayHandlerShutdownAssertions(t *testing.T, errMetrics metrics.RelayHandlerErrorMetrics, relayMap *routing.RelayMap, expected *routing.RelayData, inMemory *storage.InMemory, statsdb *routing.StatsDatabase) {
// 	relay, err := inMemory.Relay(crypto.HashID(expected.Addr.String()))
// 	assert.NoError(t, err)

// 	assert.Equal(t, routing.RelayStateMaintenance, relay.State)

// 	for i, stat := range statsdb.Entries {
// 		assert.NotEqual(t, i, relay.ID)
// 		for j := range stat.Relays {
// 			assert.NotEqual(t, j, relay.ID)
// 		}
// 	}

// 	actual := relayMap.GetRelayData(expected.Addr.String())
// 	assert.Nil(t, actual)

// 	errMetricsStruct := reflect.ValueOf(errMetrics)
// 	for i := 0; i < errMetricsStruct.NumField(); i++ {
// 		if errMetricsStruct.Field(i).CanInterface() {
// 			assert.Equal(t, 0.0, errMetricsStruct.Field(i).Interface().(metrics.Counter).ValueReset())
// 		}
// 	}
// }

// func relayHandlerSuccessAssertions(t *testing.T, recorder *httptest.ResponseRecorder, errMetrics metrics.RelayHandlerErrorMetrics, relayMap *routing.RelayMap, location routing.Location, inMemory *storage.InMemory, statsdb *routing.StatsDatabase, expected routing.RelayData, relaysToPing []routing.RelayData) {
// 	assert.Equal(t, http.StatusOK, recorder.Code)

// 	// Validate redis entry is correct
// 	actual := relayMap.GetRelayData(expected.Addr.String())

// 	assert.Equal(t, expected.ID, actual.ID)
// 	assert.Equal(t, expected.Name, actual.Name)
// 	assert.Equal(t, expected.Addr, actual.Addr)
// 	assert.NotZero(t, actual.LastUpdateTime)
// 	assert.NotEqual(t, expected.LastUpdateTime, actual.LastUpdateTime)
// 	assert.Equal(t, expected.PublicKey, actual.PublicKey)

// 	// Validate response header is correct
// 	header := recorder.Header()

// 	contentType := header.Get("Content-Type")
// 	assert.Equal(t, "application/json", contentType)

// 	// Validate response body is correct
// 	body := recorder.Body.Bytes()

// 	var response transport.RelayRequest
// 	response.UnmarshalJSON(body)

// 	assert.Equal(t, len(relaysToPing), len(response.PingStats))

// 	relaysToPingIDs := make([]uint64, 0)
// 	relaysToPingAddrs := make([]string, 0)

// 	for _, data := range response.PingStats {
// 		relaysToPingIDs = append(relaysToPingIDs, data.ID)
// 		relaysToPingAddrs = append(relaysToPingAddrs, data.Address)
// 	}

// 	// Validate statsDB is correct
// 	assert.Contains(t, statsdb.Entries, expected.ID)
// 	relations := statsdb.Entries[expected.ID]
// 	for _, relayData := range relaysToPing {
// 		id := crypto.HashID(relayData.Addr.String())
// 		assert.Contains(t, relaysToPingIDs, id)
// 		assert.Contains(t, relaysToPingAddrs, relayData.Addr.String())
// 		assert.Contains(t, relations.Relays, id)
// 	}

// 	assert.NotContains(t, relaysToPingIDs, expected.ID)
// 	assert.NotContains(t, relaysToPingAddrs, expected.Addr.String())

// 	// Validate relay state is correct
// 	relay, err := inMemory.Relay(expected.ID)
// 	assert.NoError(t, err)
// 	assert.Equal(t, routing.RelayStateEnabled, relay.State)

// 	errMetricsStruct := reflect.ValueOf(errMetrics)
// 	for i := 0; i < errMetricsStruct.NumField(); i++ {
// 		if errMetricsStruct.Field(i).CanInterface() {
// 			assert.Equal(t, 0.0, errMetricsStruct.Field(i).Interface().(metrics.Counter).ValueReset())
// 		}
// 	}
// }

// func TestRelayHandlerUnmarshalFailure(t *testing.T) {
// 	handlerMetrics := metrics.EmptyRelayHandlerMetrics
// 	localMetrics := metrics.LocalHandler{}

// 	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
// 	assert.NoError(t, err)

// 	handlerMetrics.ErrorMetrics.UnmarshalFailure = metric

// 	buff := []byte("{")
// 	recorder := pingRelayBackendHandler(t, nil, buff, handlerMetrics, nil, nil, nil)
// 	relayHandlerErrorAssertions(t, recorder, http.StatusBadRequest, metric)
// }

// func TestRelayHandlerExceedMaxRelays(t *testing.T) {
// 	addr := "127.0.0.1:40000"
// 	udpAddr, err := net.ResolveUDPAddr("udp", addr)
// 	assert.NoError(t, err)

// 	request := transport.RelayRequest{
// 		Address:   *udpAddr,
// 		PingStats: make([]transport.RelayPingStats, 1025),
// 	}

// 	handlerMetrics := metrics.EmptyRelayHandlerMetrics
// 	localMetrics := metrics.LocalHandler{}

// 	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
// 	assert.NoError(t, err)

// 	handlerMetrics.ErrorMetrics.ExceedMaxRelays = metric

// 	buff, err := request.MarshalJSON()
// 	assert.NoError(t, err)
// 	recorder := pingRelayBackendHandler(t, nil, buff, handlerMetrics, nil, nil, nil)
// 	relayHandlerErrorAssertions(t, recorder, http.StatusBadRequest, metric)
// }

// func TestRelayHandlerRelayNotFound(t *testing.T) {
// 	addr := "127.0.0.1:40000"
// 	udpAddr, err := net.ResolveUDPAddr("udp", addr)
// 	assert.NoError(t, err)

// 	inMemory := &storage.InMemory{} // Empty DB storage

// 	request := transport.RelayRequest{
// 		Address: *udpAddr,
// 	}

// 	handlerMetrics := metrics.EmptyRelayHandlerMetrics
// 	localMetrics := metrics.LocalHandler{}

// 	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
// 	assert.NoError(t, err)

// 	handlerMetrics.ErrorMetrics.RelayNotFound = metric

// 	buff, err := request.MarshalJSON()
// 	assert.NoError(t, err)
// 	recorder := pingRelayBackendHandler(t, nil, buff, handlerMetrics, inMemory, nil, nil)
// 	relayHandlerErrorAssertions(t, recorder, http.StatusNotFound, metric)
// }

// func TestRelayHandlerQuarantinedRelay(t *testing.T) {
// 	addr := "127.0.0.1:40000"
// 	udpAddr, err := net.ResolveUDPAddr("udp", addr)
// 	assert.NoError(t, err)

// 	relay := routing.Relay{
// 		ID:   crypto.HashID(addr),
// 		Addr: *udpAddr,
// 		Seller: routing.Seller{
// 			ID:   "sellerID",
// 			Name: "seller name",
// 		},
// 		Datacenter: routing.Datacenter{
// 			ID:   crypto.HashID("some datacenter"),
// 			Name: "some datacenter",
// 		},
// 		State: routing.RelayStateQuarantine,
// 	}

// 	inMemory := &storage.InMemory{}
// 	err = inMemory.AddSeller(context.Background(), relay.Seller)
// 	assert.NoError(t, err)
// 	err = inMemory.AddDatacenter(context.Background(), relay.Datacenter)
// 	assert.NoError(t, err)
// 	err = inMemory.AddRelay(context.Background(), relay)
// 	assert.NoError(t, err)

// 	request := transport.RelayRequest{
// 		Address: *udpAddr,
// 	}

// 	handlerMetrics := metrics.EmptyRelayHandlerMetrics
// 	localMetrics := metrics.LocalHandler{}

// 	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
// 	assert.NoError(t, err)

// 	handlerMetrics.ErrorMetrics.RelayQuarantined = metric

// 	buff, err := request.MarshalJSON()
// 	assert.NoError(t, err)
// 	recorder := pingRelayBackendHandler(t, nil, buff, handlerMetrics, inMemory, nil, nil)
// 	relayHandlerErrorAssertions(t, recorder, http.StatusUnauthorized, metric)
// }

// func TestRelayHandlerNoAuthHeader(t *testing.T) {
// 	addr := "127.0.0.1:40000"
// 	udpAddr, err := net.ResolveUDPAddr("udp", addr)
// 	assert.NoError(t, err)

// 	relay := routing.Relay{
// 		ID:   crypto.HashID(addr),
// 		Addr: *udpAddr,
// 		Seller: routing.Seller{
// 			ID:   "sellerID",
// 			Name: "seller name",
// 		},
// 		Datacenter: routing.Datacenter{
// 			ID:   crypto.HashID("some datacenter"),
// 			Name: "some datacenter",
// 		},
// 	}

// 	inMemory := &storage.InMemory{}
// 	err = inMemory.AddSeller(context.Background(), relay.Seller)
// 	assert.NoError(t, err)
// 	err = inMemory.AddDatacenter(context.Background(), relay.Datacenter)
// 	assert.NoError(t, err)
// 	err = inMemory.AddRelay(context.Background(), relay)
// 	assert.NoError(t, err)

// 	request := transport.RelayRequest{
// 		Address: *udpAddr,
// 	}

// 	handlerMetrics := metrics.EmptyRelayHandlerMetrics
// 	localMetrics := metrics.LocalHandler{}

// 	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
// 	assert.NoError(t, err)

// 	handlerMetrics.ErrorMetrics.NoAuthHeader = metric

// 	buff, err := request.MarshalJSON()
// 	assert.NoError(t, err)
// 	recorder := pingRelayBackendHandler(t, nil, buff, handlerMetrics, inMemory, nil, nil)
// 	relayHandlerErrorAssertions(t, recorder, http.StatusUnauthorized, metric)
// }

// func TestRelayHandlerBadAuthHeaderLength(t *testing.T) {
// 	addr := "127.0.0.1:40000"
// 	udpAddr, err := net.ResolveUDPAddr("udp", addr)
// 	assert.NoError(t, err)

// 	relay := routing.Relay{
// 		ID:   crypto.HashID(addr),
// 		Addr: *udpAddr,
// 		Seller: routing.Seller{
// 			ID:   "sellerID",
// 			Name: "seller name",
// 		},
// 		Datacenter: routing.Datacenter{
// 			ID:   crypto.HashID("some datacenter"),
// 			Name: "some datacenter",
// 		},
// 	}

// 	inMemory := &storage.InMemory{}
// 	err = inMemory.AddSeller(context.Background(), relay.Seller)
// 	assert.NoError(t, err)
// 	err = inMemory.AddDatacenter(context.Background(), relay.Datacenter)
// 	assert.NoError(t, err)
// 	err = inMemory.AddRelay(context.Background(), relay)
// 	assert.NoError(t, err)

// 	request := transport.RelayRequest{
// 		Address: *udpAddr,
// 	}

// 	handlerMetrics := metrics.EmptyRelayHandlerMetrics
// 	localMetrics := metrics.LocalHandler{}

// 	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
// 	assert.NoError(t, err)

// 	handlerMetrics.ErrorMetrics.BadAuthHeaderLength = metric

// 	buff, err := request.MarshalJSON()
// 	assert.NoError(t, err)

// 	// Set auth HTTP header
// 	headers := make(map[string]string)
// 	headers["Authorization"] = "bad"

// 	recorder := pingRelayBackendHandler(t, headers, buff, handlerMetrics, inMemory, nil, nil)
// 	relayHandlerErrorAssertions(t, recorder, http.StatusBadRequest, metric)
// }

// func TestRelayHandlerBadAuthHeaderToken(t *testing.T) {
// 	addr := "127.0.0.1:40000"
// 	udpAddr, err := net.ResolveUDPAddr("udp", addr)
// 	assert.NoError(t, err)

// 	relay := routing.Relay{
// 		ID:   crypto.HashID(addr),
// 		Addr: *udpAddr,
// 		Seller: routing.Seller{
// 			ID:   "sellerID",
// 			Name: "seller name",
// 		},
// 		Datacenter: routing.Datacenter{
// 			ID:   crypto.HashID("some datacenter"),
// 			Name: "some datacenter",
// 		},
// 	}

// 	inMemory := &storage.InMemory{}
// 	err = inMemory.AddSeller(context.Background(), relay.Seller)
// 	assert.NoError(t, err)
// 	err = inMemory.AddDatacenter(context.Background(), relay.Datacenter)
// 	assert.NoError(t, err)
// 	err = inMemory.AddRelay(context.Background(), relay)
// 	assert.NoError(t, err)

// 	request := transport.RelayRequest{
// 		Address: *udpAddr,
// 	}

// 	handlerMetrics := metrics.EmptyRelayHandlerMetrics
// 	localMetrics := metrics.LocalHandler{}

// 	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
// 	assert.NoError(t, err)

// 	handlerMetrics.ErrorMetrics.BadAuthHeaderToken = metric

// 	buff, err := request.MarshalJSON()
// 	assert.NoError(t, err)

// 	// Set auth HTTP header
// 	headers := make(map[string]string)
// 	headers["Authorization"] = "Bearer bad token"

// 	recorder := pingRelayBackendHandler(t, headers, buff, handlerMetrics, inMemory, nil, nil)
// 	relayHandlerErrorAssertions(t, recorder, http.StatusBadRequest, metric)
// }

// func TestRelayHandlerBadNonce(t *testing.T) {
// 	addr := "127.0.0.1:40000"
// 	udpAddr, err := net.ResolveUDPAddr("udp", addr)
// 	assert.NoError(t, err)

// 	relay := routing.Relay{
// 		ID:   crypto.HashID(addr),
// 		Addr: *udpAddr,
// 		Seller: routing.Seller{
// 			ID:   "sellerID",
// 			Name: "seller name",
// 		},
// 		Datacenter: routing.Datacenter{
// 			ID:   crypto.HashID("some datacenter"),
// 			Name: "some datacenter",
// 		},
// 	}

// 	inMemory := &storage.InMemory{}
// 	err = inMemory.AddSeller(context.Background(), relay.Seller)
// 	assert.NoError(t, err)
// 	err = inMemory.AddDatacenter(context.Background(), relay.Datacenter)
// 	assert.NoError(t, err)
// 	err = inMemory.AddRelay(context.Background(), relay)
// 	assert.NoError(t, err)

// 	request := transport.RelayRequest{
// 		Address: *udpAddr,
// 	}

// 	handlerMetrics := metrics.EmptyRelayHandlerMetrics
// 	localMetrics := metrics.LocalHandler{}

// 	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
// 	assert.NoError(t, err)

// 	handlerMetrics.ErrorMetrics.BadNonce = metric

// 	buff, err := request.MarshalJSON()
// 	assert.NoError(t, err)

// 	// Set auth HTTP header
// 	headers := make(map[string]string)
// 	headers["Authorization"] = "Bearer invalid:base64"

// 	recorder := pingRelayBackendHandler(t, headers, buff, handlerMetrics, inMemory, nil, nil)
// 	relayHandlerErrorAssertions(t, recorder, http.StatusBadRequest, metric)
// }

// func TestRelayHandlerBadEncryptedAddress(t *testing.T) {
// 	addr := "127.0.0.1:40000"
// 	udpAddr, err := net.ResolveUDPAddr("udp", addr)
// 	assert.NoError(t, err)

// 	relay := routing.Relay{
// 		ID:   crypto.HashID(addr),
// 		Addr: *udpAddr,
// 		Seller: routing.Seller{
// 			ID:   "sellerID",
// 			Name: "seller name",
// 		},
// 		Datacenter: routing.Datacenter{
// 			ID:   crypto.HashID("some datacenter"),
// 			Name: "some datacenter",
// 		},
// 	}

// 	inMemory := &storage.InMemory{}
// 	err = inMemory.AddSeller(context.Background(), relay.Seller)
// 	assert.NoError(t, err)
// 	err = inMemory.AddDatacenter(context.Background(), relay.Datacenter)
// 	assert.NoError(t, err)
// 	err = inMemory.AddRelay(context.Background(), relay)
// 	assert.NoError(t, err)

// 	nonce := make([]byte, crypto.NonceSize)
// 	rand.Read(nonce)

// 	nonceBase64 := base64.StdEncoding.EncodeToString(nonce)
// 	encryptedAddressBase64 := "badaddress"

// 	token := nonceBase64 + ":" + encryptedAddressBase64

// 	request := transport.RelayRequest{
// 		Address: *udpAddr,
// 	}

// 	handlerMetrics := metrics.EmptyRelayHandlerMetrics
// 	localMetrics := metrics.LocalHandler{}

// 	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
// 	assert.NoError(t, err)

// 	handlerMetrics.ErrorMetrics.BadEncryptedAddress = metric

// 	buff, err := request.MarshalJSON()
// 	assert.NoError(t, err)

// 	// Set auth HTTP header
// 	headers := make(map[string]string)
// 	headers["Authorization"] = "Bearer " + token

// 	recorder := pingRelayBackendHandler(t, headers, buff, handlerMetrics, inMemory, nil, nil)
// 	relayHandlerErrorAssertions(t, recorder, http.StatusBadRequest, metric)
// }

// func TestRelayHandlerDecryptFailure(t *testing.T) {
// 	addr := "127.0.0.1:40000"
// 	udpAddr, err := net.ResolveUDPAddr("udp", addr)
// 	assert.NoError(t, err)

// 	// Don't use the other key in the key pairs to fail decryption
// 	_, relayPrivateKey, err := box.GenerateKey(rand.Reader)
// 	assert.NoError(t, err)
// 	routerPublicKey, _, err := box.GenerateKey(rand.Reader)
// 	assert.NoError(t, err)

// 	relay := routing.Relay{
// 		ID:   crypto.HashID(addr),
// 		Addr: *udpAddr,
// 		Seller: routing.Seller{
// 			ID:   "sellerID",
// 			Name: "seller name",
// 		},
// 		Datacenter: routing.Datacenter{
// 			ID:   crypto.HashID("some datacenter"),
// 			Name: "some datacenter",
// 		},
// 	}

// 	inMemory := &storage.InMemory{}
// 	err = inMemory.AddSeller(context.Background(), relay.Seller)
// 	assert.NoError(t, err)
// 	err = inMemory.AddDatacenter(context.Background(), relay.Datacenter)
// 	assert.NoError(t, err)
// 	err = inMemory.AddRelay(context.Background(), relay)
// 	assert.NoError(t, err)

// 	nonce := make([]byte, crypto.NonceSize)
// 	rand.Read(nonce)

// 	// Encrypt the address
// 	encryptedAddress := crypto.Seal([]byte(addr), nonce, routerPublicKey[:], relayPrivateKey[:])

// 	nonceBase64 := base64.StdEncoding.EncodeToString(nonce)
// 	encryptedAddressBase64 := base64.StdEncoding.EncodeToString(encryptedAddress)

// 	token := nonceBase64 + ":" + encryptedAddressBase64

// 	request := transport.RelayRequest{
// 		Address: *udpAddr,
// 	}

// 	handlerMetrics := metrics.EmptyRelayHandlerMetrics
// 	localMetrics := metrics.LocalHandler{}

// 	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
// 	assert.NoError(t, err)

// 	handlerMetrics.ErrorMetrics.DecryptFailure = metric

// 	buff, err := request.MarshalJSON()
// 	assert.NoError(t, err)

// 	// Set auth HTTP header
// 	headers := make(map[string]string)
// 	headers["Authorization"] = "Bearer " + token

// 	recorder := pingRelayBackendHandler(t, headers, buff, handlerMetrics, inMemory, nil, nil)
// 	relayHandlerErrorAssertions(t, recorder, http.StatusUnauthorized, metric)
// }

// func TestRelayHandlerShuttingDown(t *testing.T) {
// 	addr := "127.0.0.1:40000"
// 	udpAddr, err := net.ResolveUDPAddr("udp", addr)
// 	assert.NoError(t, err)

// 	relayPublicKey, relayPrivateKey, err := box.GenerateKey(rand.Reader)
// 	assert.NoError(t, err)
// 	routerPublicKey, routerPrivateKey, err := box.GenerateKey(rand.Reader)
// 	assert.NoError(t, err)

// 	statsdb := routing.NewStatsDatabase()

// 	callback := func(relayID uint64) error {
// 		statsdb.DeleteEntry(relayID)
// 		return nil
// 	}
// 	relayMap := routing.NewRelayMap(callback)

// 	relaysToPing := make([]routing.RelayData, 0)
// 	for i := 0; i < 4; i++ {
// 		addr, err := net.ResolveUDPAddr("udp", "127.0.0."+strconv.FormatInt(int64(i+2), 10)+":40000")
// 		assert.NoError(t, err)

// 		relayToPing := routing.RelayData{
// 			ID:   crypto.HashID(addr.String()),
// 			Addr: *addr,
// 		}
// 		relaysToPing = append(relaysToPing, relayToPing)
// 		relayMap.UpdateRelayData(addr.String(), &relayToPing)
// 	}

// 	// Create relay in DB storage
// 	relay := routing.Relay{
// 		ID:        crypto.HashID(addr),
// 		Addr:      *udpAddr,
// 		PublicKey: relayPublicKey[:],
// 		Seller: routing.Seller{
// 			ID:   "sellerID",
// 			Name: "seller name",
// 		},
// 		Datacenter: routing.Datacenter{
// 			ID:   crypto.HashID("some datacenter"),
// 			Name: "some datacenter",
// 		},
// 	}

// 	customerPublicKey := make([]byte, crypto.KeySize)
// 	rand.Read(customerPublicKey)

// 	inMemory := &storage.InMemory{
// 		LocalMode: true,
// 	}
// 	err = inMemory.AddBuyer(context.Background(), routing.Buyer{
// 		PublicKey: customerPublicKey[8:],
// 	})
// 	assert.NoError(t, err)
// 	err = inMemory.AddSeller(context.Background(), relay.Seller)
// 	assert.NoError(t, err)
// 	err = inMemory.AddDatacenter(context.Background(), relay.Datacenter)
// 	assert.NoError(t, err)
// 	err = inMemory.AddRelay(context.Background(), relay)
// 	assert.NoError(t, err)

// 	nonce := make([]byte, crypto.NonceSize)
// 	rand.Read(nonce)

// 	// Encrypt the address
// 	encryptedAddress := crypto.Seal([]byte(addr), nonce, routerPublicKey[:], relayPrivateKey[:])

// 	nonceBase64 := base64.StdEncoding.EncodeToString(nonce)
// 	encryptedAddressBase64 := base64.StdEncoding.EncodeToString(encryptedAddress)

// 	token := nonceBase64 + ":" + encryptedAddressBase64

// 	request := transport.RelayRequest{
// 		Address: *udpAddr,
// 		PingStats: []transport.RelayPingStats{
// 			{
// 				ID:         relaysToPing[0].ID,
// 				Address:    relaysToPing[0].Addr.String(),
// 				RTT:        1,
// 				Jitter:     2,
// 				PacketLoss: 3,
// 			},

// 			{
// 				ID:         relaysToPing[1].ID,
// 				Address:    relaysToPing[1].Addr.String(),
// 				RTT:        4,
// 				Jitter:     5,
// 				PacketLoss: 6,
// 			},

// 			{
// 				ID:         relaysToPing[2].ID,
// 				Address:    relaysToPing[2].Addr.String(),
// 				RTT:        7,
// 				Jitter:     8,
// 				PacketLoss: 9,
// 			},

// 			{
// 				ID:         relaysToPing[3].ID,
// 				Address:    relaysToPing[3].Addr.String(),
// 				RTT:        10,
// 				Jitter:     11,
// 				PacketLoss: 12,
// 			},
// 		},
// 		TrafficStats: routing.RelayTrafficStats{
// 			SessionCount:  10,
// 			BytesSent:     1000000,
// 			BytesReceived: 1000000,
// 		},
// 		ShuttingDown: true,
// 	}

// 	expected := routing.RelayData{
// 		ID:   crypto.HashID(addr),
// 		Addr: *udpAddr,
// 		Datacenter: routing.Datacenter{
// 			ID:   1,
// 			Name: "some name",
// 		},
// 		PublicKey:      relayPublicKey[:],
// 		LastUpdateTime: time.Now().Add(-time.Second),
// 	}

// 	localMetrics := metrics.LocalHandler{}
// 	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
// 	assert.NoError(t, err)

// 	handlerMetrics := metrics.RelayHandlerMetrics{
// 		Invocations:   &metrics.EmptyCounter{},
// 		DurationGauge: &metrics.EmptyGauge{},
// 	}
// 	v := reflect.ValueOf(&handlerMetrics.ErrorMetrics).Elem()
// 	for i := 0; i < v.NumField(); i++ {
// 		if v.Field(i).CanSet() {
// 			v.Field(i).Set(reflect.ValueOf(metric))
// 		}
// 	}

// 	buff, err := request.MarshalJSON()
// 	assert.NoError(t, err)

// 	// Set auth HTTP header
// 	headers := make(map[string]string)
// 	headers["Authorization"] = "Bearer " + token

// 	pingRelayBackendHandler(t, headers, buff, handlerMetrics, inMemory, statsdb, routerPrivateKey[:])
// 	relayHandlerShutdownAssertions(t, handlerMetrics.ErrorMetrics, relayMap, &expected, inMemory, statsdb)

// 	// Now make the same request again, with the relay now initialized
// 	pingRelayBackendHandler(t, headers, buff, handlerMetrics, inMemory, statsdb, routerPrivateKey[:])
// 	relayHandlerShutdownAssertions(t, handlerMetrics.ErrorMetrics, relayMap, &expected, inMemory, statsdb)
// }

// func TestRelayHandlerSuccess(t *testing.T) {
// 	addr := "127.0.0.1:40000"
// 	udpAddr, err := net.ResolveUDPAddr("udp", addr)
// 	assert.NoError(t, err)

// 	relayPublicKey, relayPrivateKey, err := box.GenerateKey(rand.Reader)
// 	assert.NoError(t, err)
// 	routerPublicKey, routerPrivateKey, err := box.GenerateKey(rand.Reader)
// 	assert.NoError(t, err)

// 	statsdb := routing.NewStatsDatabase()

// 	location := routing.Location{
// 		Latitude:  math.Round(mrand.Float64()*1000) / 1000,
// 		Longitude: math.Round(mrand.Float64()*1000) / 1000,
// 	}

// 	callback := func(relayID uint64) error {
// 		statsdb.DeleteEntry(relayID)
// 		return nil
// 	}
// 	relayMap := routing.NewRelayMap(callback)

// 	relaysToPing := make([]routing.RelayData, 0)
// 	for i := 0; i < 4; i++ {
// 		addr, err := net.ResolveUDPAddr("udp", "127.0.0."+strconv.FormatInt(int64(i+2), 10)+":40000")
// 		assert.NoError(t, err)

// 		relayToPing := routing.RelayData{
// 			ID:   crypto.HashID(addr.String()),
// 			Addr: *addr,
// 		}
// 		relaysToPing = append(relaysToPing, relayToPing)
// 		relayMap.UpdateRelayData(addr.String(), &relayToPing)
// 	}

// 	// Create relay in DB storage
// 	relay := routing.Relay{
// 		ID:        crypto.HashID(addr),
// 		Addr:      *udpAddr,
// 		PublicKey: relayPublicKey[:],
// 		Seller: routing.Seller{
// 			ID:   "sellerID",
// 			Name: "seller name",
// 		},
// 		Datacenter: routing.Datacenter{
// 			ID:       crypto.HashID("some datacenter"),
// 			Name:     "some datacenter",
// 			Location: location,
// 		},
// 	}

// 	customerPublicKey := make([]byte, crypto.KeySize)
// 	rand.Read(customerPublicKey)

// 	inMemory := &storage.InMemory{
// 		LocalMode: true,
// 	}
// 	err = inMemory.AddBuyer(context.Background(), routing.Buyer{
// 		PublicKey: customerPublicKey[8:],
// 	})
// 	assert.NoError(t, err)
// 	err = inMemory.AddSeller(context.Background(), relay.Seller)
// 	assert.NoError(t, err)
// 	err = inMemory.AddDatacenter(context.Background(), relay.Datacenter)
// 	assert.NoError(t, err)
// 	err = inMemory.AddRelay(context.Background(), relay)
// 	assert.NoError(t, err)

// 	nonce := make([]byte, crypto.NonceSize)
// 	rand.Read(nonce)

// 	// Encrypt the address
// 	encryptedAddress := crypto.Seal([]byte(addr), nonce, routerPublicKey[:], relayPrivateKey[:])

// 	nonceBase64 := base64.StdEncoding.EncodeToString(nonce)
// 	encryptedAddressBase64 := base64.StdEncoding.EncodeToString(encryptedAddress)

// 	token := nonceBase64 + ":" + encryptedAddressBase64

// 	request := transport.RelayRequest{
// 		Address: *udpAddr,
// 		PingStats: []transport.RelayPingStats{
// 			{
// 				ID:         relaysToPing[0].ID,
// 				Address:    relaysToPing[0].Addr.String(),
// 				RTT:        1,
// 				Jitter:     2,
// 				PacketLoss: 3,
// 			},

// 			{
// 				ID:         relaysToPing[1].ID,
// 				Address:    relaysToPing[1].Addr.String(),
// 				RTT:        4,
// 				Jitter:     5,
// 				PacketLoss: 6,
// 			},

// 			{
// 				ID:         relaysToPing[2].ID,
// 				Address:    relaysToPing[2].Addr.String(),
// 				RTT:        7,
// 				Jitter:     8,
// 				PacketLoss: 9,
// 			},

// 			{
// 				ID:         relaysToPing[3].ID,
// 				Address:    relaysToPing[3].Addr.String(),
// 				RTT:        10,
// 				Jitter:     11,
// 				PacketLoss: 12,
// 			},
// 		},
// 		TrafficStats: routing.RelayTrafficStats{
// 			SessionCount:  10,
// 			BytesSent:     1000000,
// 			BytesReceived: 1000000,
// 		},
// 	}

// 	expected := routing.RelayData{
// 		ID:   crypto.HashID(addr),
// 		Addr: *udpAddr,
// 		Datacenter: routing.Datacenter{
// 			ID:   1,
// 			Name: "some name",
// 		},
// 		PublicKey:      relayPublicKey[:],
// 		LastUpdateTime: time.Now().Add(-time.Second),
// 	}

// 	localMetrics := metrics.LocalHandler{}
// 	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
// 	assert.NoError(t, err)

// 	handlerMetrics := metrics.RelayHandlerMetrics{
// 		Invocations:   &metrics.EmptyCounter{},
// 		DurationGauge: &metrics.EmptyGauge{},
// 	}
// 	v := reflect.ValueOf(&handlerMetrics.ErrorMetrics).Elem()
// 	for i := 0; i < v.NumField(); i++ {
// 		if v.Field(i).CanSet() {
// 			v.Field(i).Set(reflect.ValueOf(metric))
// 		}
// 	}

// 	buff, err := request.MarshalJSON()
// 	assert.NoError(t, err)

// 	// Set auth HTTP header
// 	headers := make(map[string]string)
// 	headers["Authorization"] = "Bearer " + token

// 	recorder := pingRelayBackendHandler(t, headers, buff, handlerMetrics, inMemory, statsdb, routerPrivateKey[:])
// 	relayHandlerSuccessAssertions(t, recorder, handlerMetrics.ErrorMetrics, relayMap, location, inMemory, statsdb, expected, relaysToPing)

// 	// Now make the same request again, with the relay now initialized
// 	recorder = pingRelayBackendHandler(t, headers, buff, handlerMetrics, inMemory, statsdb, routerPrivateKey[:])
// 	relayHandlerSuccessAssertions(t, recorder, handlerMetrics.ErrorMetrics, relayMap, location, inMemory, statsdb, expected, relaysToPing)
// }
