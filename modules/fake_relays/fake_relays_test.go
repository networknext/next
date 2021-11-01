package fake_relays

import (
	"encoding/base64"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/transport"
	"github.com/stretchr/testify/assert"
)

func TestFakeNewRelays(t *testing.T) {
	t.Parallel()

	fakeRelayMetrics := metrics.EmptyFakeRelayMetrics

	t.Run("failed to resolve udp address", func(t *testing.T) {
		relays, err := NewFakeRelays(100000, []byte{}, "", 4, fakeRelayMetrics)
		assert.Error(t, err)
		assert.Equal(t, fmt.Errorf("error resolving UDP address: address 65536: invalid port\n"), err)
		assert.Equal(t, 0, len(relays))
	})

	t.Run("success", func(t *testing.T) {
		relays, err := NewFakeRelays(1, []byte{}, "", 3, fakeRelayMetrics)
		assert.NoError(t, err)
		assert.NotNil(t, relays)
		assert.Equal(t, 1, len(relays))
	})
}

func TestSendRequests(t *testing.T) {
	t.Parallel()

	fakeRelayMetrics := metrics.EmptyFakeRelayMetrics
	relayPublicKey, err := base64.StdEncoding.DecodeString("8hUCRvzKh2aknL9RErM/Vj22+FGJW0tWMRz5KlHKryE=")
	assert.NoError(t, err)

	t.Run("shutdown marshal error", func(t *testing.T) {
		relays, err := NewFakeRelays(1, relayPublicKey, "", 2, fakeRelayMetrics)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(relays))

		err = relays[0].sendShutdownRequest()
		assert.Error(t, err)
		assert.Equal(t, fmt.Errorf("error marshaling shutdown request: invalid update request version: 2"), err)
	})

	t.Run("shutdown response not OK", func(t *testing.T) {
		gatewayHandler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("test"))
		}

		svr := httptest.NewServer(http.HandlerFunc(gatewayHandler))
		defer svr.Close()

		gatewayAddr := strings.TrimLeft(svr.URL, "http://")
		assert.NotEqual(t, gatewayAddr, svr.URL)

		relays, err := NewFakeRelays(1, relayPublicKey, gatewayAddr, 4, fakeRelayMetrics)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(relays))
		assert.NotNil(t, relays[0])

		err = relays[0].sendShutdownRequest()
		assert.Error(t, err)
		assert.Equal(t, fmt.Errorf("shutdown response was non 200: 404"), err)
	})

	t.Run("shutdown response OK", func(t *testing.T) {
		gatewayHandler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("test"))
		}

		svr := httptest.NewServer(http.HandlerFunc(gatewayHandler))
		defer svr.Close()

		gatewayAddr := strings.TrimLeft(svr.URL, "http://")
		assert.NotEqual(t, gatewayAddr, svr.URL)

		relays, err := NewFakeRelays(1, relayPublicKey, gatewayAddr, 4, fakeRelayMetrics)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(relays))
		assert.NotNil(t, relays[0])

		err = relays[0].sendShutdownRequest()
		assert.NoError(t, err)
	})

	t.Run("initial update request marshal error", func(t *testing.T) {
		relays, err := NewFakeRelays(1, relayPublicKey, "", 2, fakeRelayMetrics)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(relays))

		err = relays[0].sendInitialUpdateRequest()
		assert.Error(t, err)
		assert.Equal(t, fmt.Errorf("error marshaling update request: invalid update request version: 2"), err)
	})

	t.Run("initial update request response not OK", func(t *testing.T) {
		gatewayHandler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("test"))
		}

		svr := httptest.NewServer(http.HandlerFunc(gatewayHandler))
		defer svr.Close()

		gatewayAddr := strings.TrimLeft(svr.URL, "http://")
		assert.NotEqual(t, gatewayAddr, svr.URL)

		relays, err := NewFakeRelays(1, relayPublicKey, gatewayAddr, 4, fakeRelayMetrics)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(relays))
		assert.NotNil(t, relays[0])

		err = relays[0].sendInitialUpdateRequest()
		assert.Error(t, err)
		assert.Equal(t, fmt.Errorf("response was non 200: 404"), err)
	})

	t.Run("initial update request unmarshaling error", func(t *testing.T) {
		gatewayHandler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("test"))
		}

		svr := httptest.NewServer(http.HandlerFunc(gatewayHandler))
		defer svr.Close()

		gatewayAddr := strings.TrimLeft(svr.URL, "http://")
		assert.NotEqual(t, gatewayAddr, svr.URL)

		relays, err := NewFakeRelays(1, relayPublicKey, gatewayAddr, 3, fakeRelayMetrics)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(relays))
		assert.NotNil(t, relays[0])

		err = relays[0].sendInitialUpdateRequest()
		assert.Error(t, err)
		assert.Equal(t, fmt.Errorf("error unmarshaling update response: invalid version number"), err)
	})

	t.Run("initial update request success", func(t *testing.T) {
		gatewayHandler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			emptyResponse := &transport.RelayUpdateResponse{}
			bin, err := emptyResponse.MarshalBinary()
			assert.NoError(t, err)
			w.Write(bin)
		}

		svr := httptest.NewServer(http.HandlerFunc(gatewayHandler))
		defer svr.Close()

		gatewayAddr := strings.TrimLeft(svr.URL, "http://")
		assert.NotEqual(t, gatewayAddr, svr.URL)

		relays, err := NewFakeRelays(1, relayPublicKey, gatewayAddr, 4, fakeRelayMetrics)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(relays))
		assert.NotNil(t, relays[0])

		err = relays[0].sendInitialUpdateRequest()
		assert.NoError(t, err)
	})

	t.Run("standard update request marshal error", func(t *testing.T) {
		relays, err := NewFakeRelays(1, relayPublicKey, "", 2, fakeRelayMetrics)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(relays))

		err = relays[0].sendStandardUpdateRequest()
		assert.Error(t, err)
		assert.Equal(t, fmt.Errorf("error marshaling update request: invalid update request version: 2"), err)
	})

	t.Run("standard update request response not OK", func(t *testing.T) {
		gatewayHandler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte("test"))
		}

		svr := httptest.NewServer(http.HandlerFunc(gatewayHandler))
		defer svr.Close()

		gatewayAddr := strings.TrimLeft(svr.URL, "http://")
		assert.NotEqual(t, gatewayAddr, svr.URL)

		relays, err := NewFakeRelays(1, relayPublicKey, gatewayAddr, 4, fakeRelayMetrics)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(relays))
		assert.NotNil(t, relays[0])

		err = relays[0].sendStandardUpdateRequest()
		assert.Error(t, err)
		assert.Equal(t, fmt.Errorf("response was non 200: 404"), err)
	})

	t.Run("standard update request unmarshaling error", func(t *testing.T) {
		gatewayHandler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("test"))
		}

		svr := httptest.NewServer(http.HandlerFunc(gatewayHandler))
		defer svr.Close()

		gatewayAddr := strings.TrimLeft(svr.URL, "http://")
		assert.NotEqual(t, gatewayAddr, svr.URL)

		relays, err := NewFakeRelays(1, relayPublicKey, gatewayAddr, 4, fakeRelayMetrics)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(relays))
		assert.NotNil(t, relays[0])

		err = relays[0].sendStandardUpdateRequest()
		assert.Error(t, err)
		assert.Equal(t, fmt.Errorf("error unmarshaling update response: invalid version number"), err)
	})

	t.Run("standard update request success", func(t *testing.T) {
		gatewayHandler := func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusOK)
			emptyResponse := &transport.RelayUpdateResponse{}
			bin, err := emptyResponse.MarshalBinary()
			assert.NoError(t, err)
			w.Write(bin)
		}

		svr := httptest.NewServer(http.HandlerFunc(gatewayHandler))
		defer svr.Close()

		gatewayAddr := strings.TrimLeft(svr.URL, "http://")
		assert.NotEqual(t, gatewayAddr, svr.URL)

		relays, err := NewFakeRelays(1, relayPublicKey, gatewayAddr, 4, fakeRelayMetrics)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(relays))
		assert.NotNil(t, relays[0])

		err = relays[0].sendStandardUpdateRequest()
		assert.NoError(t, err)
	})
}

func TestSimulatedPingStats(t *testing.T) {
	fakeRelayMetrics := metrics.EmptyFakeRelayMetrics
	relayPublicKey, err := base64.StdEncoding.DecodeString("8hUCRvzKh2aknL9RErM/Vj22+FGJW0tWMRz5KlHKryE=")
	assert.NoError(t, err)

	relays, err := NewFakeRelays(1, relayPublicKey, "", 4, fakeRelayMetrics)
	assert.NoError(t, err)

	// Seed the RNG
	rand.Seed(1000)

	base := RouteBase{
		rtt:        float32(100),
		jitter:     float32(5),
		packetLoss: float32(0),
	}
	expectedStats := routing.RelayStatsPing{
		RelayID:    0,
		RTT:        float32(95),
		Jitter:     float32(5.05),
		PacketLoss: float32(0),
	}
	stats := relays[0].newRelayPingStats(0, base)

	assert.Equal(t, expectedStats, stats)
}
