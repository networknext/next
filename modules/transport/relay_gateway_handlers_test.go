package transport_test

import (
	"bytes"
	"context"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/transport"

	"github.com/stretchr/testify/assert"
)

func TestGatewayRelayInit(t *testing.T) {
	t.Parallel()

	svr := httptest.NewServer(http.HandlerFunc(transport.GatewayRelayInitHandlerFunc()))
	defer svr.Close()

	client := svr.Client()
	res, err := client.Post(svr.URL, "application/octet-stream", nil)
	assert.NoError(t, err)

	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, "application/octet-stream", res.Header.Get("Content-Type"))

	body, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)
	defer res.Body.Close()

	index := 0

	var version uint32
	assert.True(t, encoding.ReadUint32(body, &index, &version))
	assert.Equal(t, uint32(0), version)

	var respTime uint64
	assert.True(t, encoding.ReadUint64(body, &index, &respTime))
	assert.Greater(t, respTime, uint64(0))

	var pubKey []byte
	assert.True(t, encoding.ReadBytes(body, &index, &pubKey, uint32(32)))
	assert.Equal(t, pubKey, make([]byte, 32))
}

func getGatewayRelayUpdateHandlerConfig(t *testing.T, relays []routing.Relay) transport.GatewayRelayUpdateHandlerConfig {
	requestChan := make(chan []byte, 10000)

	gatewayMetrics, err := metrics.NewRelayGatewayMetrics(context.Background(), &metrics.LocalHandler{}, "relay_gateway", "relay_gateway", "Relay Gateway", "relay update request")
	assert.NoError(t, err)

	relayHash := make(map[uint64]routing.Relay)
	for _, relay := range relays {
		relayHash[relay.ID] = relay
	}

	getRelayData := func() ([]routing.Relay, map[uint64]routing.Relay) {
		return relays, relayHash
	}

	return transport.GatewayRelayUpdateHandlerConfig{
		RequestChan:  requestChan,
		Metrics:      gatewayMetrics,
		GetRelayData: getRelayData,
	}
}

func TestGatewayRelayUpdate_ContentTypeFailure(t *testing.T) {
	t.Parallel()

	config := getGatewayRelayUpdateHandlerConfig(t, []routing.Relay{})

	svr := httptest.NewServer(http.HandlerFunc(transport.GatewayRelayUpdateHandlerFunc(config)))
	defer svr.Close()

	client := svr.Client()
	res, err := client.Post(svr.URL, "application/json", nil)
	assert.NoError(t, err)

	assert.Equal(t, 400, res.StatusCode)
	assert.Equal(t, float64(1), config.Metrics.ErrorMetrics.ContentTypeFailure.Value())
}

func TestGatewayRelayUpdate_UnmarshalFailure(t *testing.T) {
	t.Parallel()

	config := getGatewayRelayUpdateHandlerConfig(t, []routing.Relay{})

	svr := httptest.NewServer(http.HandlerFunc(transport.GatewayRelayUpdateHandlerFunc(config)))
	defer svr.Close()

	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
	assert.NoError(t, err)

	updateRequest := &transport.RelayUpdateRequest{
		Version:      5,
		RelayVersion: "2.1.0",
		Address:      *addr,
		Token:        make([]byte, crypto.KeySize),
	}

	bin, err := updateRequest.MarshalBinary()
	assert.NoError(t, err)

	client := svr.Client()
	res, err := client.Post(svr.URL, "application/octet-stream", bytes.NewBuffer(bin[15:]))
	assert.NoError(t, err)

	assert.Equal(t, 400, res.StatusCode)
	assert.Equal(t, float64(1), config.Metrics.ErrorMetrics.UnmarshalFailure.Value())
}

func TestGatewayRelayUpdate_ExceedMaxRelays(t *testing.T) {
	t.Parallel()

	config := getGatewayRelayUpdateHandlerConfig(t, []routing.Relay{})

	svr := httptest.NewServer(http.HandlerFunc(transport.GatewayRelayUpdateHandlerFunc(config)))
	defer svr.Close()

	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
	assert.NoError(t, err)

	updateRequest := &transport.RelayUpdateRequest{
		Version:      5,
		RelayVersion: "2.1.0",
		Address:      *addr,
		Token:        make([]byte, crypto.KeySize),
		PingStats:    make([]routing.RelayStatsPing, transport.MaxRelays+1),
	}

	bin, err := updateRequest.MarshalBinary()
	assert.NoError(t, err)

	client := svr.Client()
	res, err := client.Post(svr.URL, "application/octet-stream", bytes.NewBuffer(bin))
	assert.NoError(t, err)

	assert.Equal(t, 400, res.StatusCode)
	assert.Equal(t, float64(1), config.Metrics.ErrorMetrics.ExceedMaxRelays.Value())
}

func TestGatewayRelayUpdate_RelayNotFound(t *testing.T) {
	t.Parallel()

	config := getGatewayRelayUpdateHandlerConfig(t, []routing.Relay{})

	svr := httptest.NewServer(http.HandlerFunc(transport.GatewayRelayUpdateHandlerFunc(config)))
	defer svr.Close()

	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
	assert.NoError(t, err)

	updateRequest := &transport.RelayUpdateRequest{
		Version:      5,
		RelayVersion: "2.1.0",
		Address:      *addr,
		Token:        make([]byte, crypto.KeySize),
	}

	bin, err := updateRequest.MarshalBinary()
	assert.NoError(t, err)

	client := svr.Client()
	res, err := client.Post(svr.URL, "application/octet-stream", bytes.NewBuffer(bin))
	assert.NoError(t, err)

	assert.Equal(t, 404, res.StatusCode)
	assert.Equal(t, float64(1), config.Metrics.ErrorMetrics.RelayNotFound.Value())
}

func TestGatewayRelayUpdate_Success(t *testing.T) {
	t.Parallel()

	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
	assert.NoError(t, err)

	relay1 := routing.Relay{
		ID:   crypto.HashID("127.0.0.1:40000"),
		Addr: *addr,
		Datacenter: routing.Datacenter{
			ID:   1,
			Name: "some name",
		},
		PublicKey: make([]byte, crypto.KeySize),
		Seller: routing.Seller{
			ID:   "sellerID",
			Name: "seller name",
		},
		State:   routing.RelayStateEnabled,
		Version: "2.1.0",
	}

	config := getGatewayRelayUpdateHandlerConfig(t, []routing.Relay{relay1})

	svr := httptest.NewServer(http.HandlerFunc(transport.GatewayRelayUpdateHandlerFunc(config)))
	defer svr.Close()

	updateRequest := &transport.RelayUpdateRequest{
		Version:      5,
		RelayVersion: relay1.Version,
		Address:      relay1.Addr,
		Token:        make([]byte, crypto.KeySize),
	}

	bin, err := updateRequest.MarshalBinary()
	assert.NoError(t, err)

	client := svr.Client()
	res, err := client.Post(svr.URL, "application/octet-stream", bytes.NewBuffer(bin))
	assert.NoError(t, err)

	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, 1, len(config.RequestChan))

	body, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)
	defer res.Body.Close()

	response := &transport.RelayUpdateResponse{}
	err = response.UnmarshalBinary(body)
	assert.NoError(t, err)
}

func TestGatewayRelayUpdate_PingInternalAddr_NotSameSeller(t *testing.T) {
	t.Parallel()

	addr1, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
	assert.NoError(t, err)
	intAddr1, err := net.ResolveUDPAddr("udp", "10.128.0.1:40000")
	assert.NoError(t, err)

	relay1 := routing.Relay{
		ID:           crypto.HashID("127.0.0.1:40000"),
		Addr:         *addr1,
		InternalAddr: *intAddr1,
		Datacenter: routing.Datacenter{
			ID:   1,
			Name: "some name",
		},
		PublicKey: make([]byte, crypto.KeySize),
		Seller: routing.Seller{
			ID:   "sellerID1",
			Name: "seller name 1",
		},
		State:   routing.RelayStateEnabled,
		Version: "2.1.0",
	}

	addr2, err := net.ResolveUDPAddr("udp", "127.0.0.2:40000")
	assert.NoError(t, err)
	intAddr2, err := net.ResolveUDPAddr("udp", "10.128.0.2:40000")
	assert.NoError(t, err)

	relay2 := routing.Relay{
		ID:           crypto.HashID("127.0.0.2:40000"),
		Addr:         *addr2,
		InternalAddr: *intAddr2,
		Datacenter: routing.Datacenter{
			ID:   2,
			Name: "some name 2",
		},
		PublicKey: make([]byte, crypto.KeySize),
		Seller: routing.Seller{
			ID:   "sellerID2",
			Name: "seller name 2",
		},
		State:   routing.RelayStateEnabled,
		Version: "2.1.0",
	}

	config := getGatewayRelayUpdateHandlerConfig(t, []routing.Relay{relay1, relay2})

	svr := httptest.NewServer(http.HandlerFunc(transport.GatewayRelayUpdateHandlerFunc(config)))
	defer svr.Close()

	updateRequest := &transport.RelayUpdateRequest{
		Version:      5,
		RelayVersion: relay1.Version,
		Address:      relay1.Addr,
		Token:        make([]byte, crypto.KeySize),
	}

	bin, err := updateRequest.MarshalBinary()
	assert.NoError(t, err)

	client := svr.Client()
	res, err := client.Post(svr.URL, "application/octet-stream", bytes.NewBuffer(bin))
	assert.NoError(t, err)

	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, 1, len(config.RequestChan))

	body, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)
	defer res.Body.Close()

	response := &transport.RelayUpdateResponse{}
	err = response.UnmarshalBinary(body)
	assert.NoError(t, err)

	assert.Equal(t, 1, len(response.RelaysToPing))
	assert.Equal(t, relay2.ID, response.RelaysToPing[0].ID)
	assert.Equal(t, relay2.Addr.String(), response.RelaysToPing[0].Address)
}

func TestGatewayRelayUpdate_PingInternalAddr_NoInternalIP(t *testing.T) {
	t.Parallel()

	addr1, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
	assert.NoError(t, err)
	intAddr1, err := net.ResolveUDPAddr("udp", "10.128.0.1:40000")
	assert.NoError(t, err)

	relay1 := routing.Relay{
		ID:           crypto.HashID("127.0.0.1:40000"),
		Addr:         *addr1,
		InternalAddr: *intAddr1,
		Datacenter: routing.Datacenter{
			ID:   1,
			Name: "some name",
		},
		PublicKey: make([]byte, crypto.KeySize),
		Seller: routing.Seller{
			ID:   "sellerID1",
			Name: "seller name 1",
		},
		State:   routing.RelayStateEnabled,
		Version: "2.1.0",
	}

	addr2, err := net.ResolveUDPAddr("udp", "127.0.0.2:40000")
	assert.NoError(t, err)

	relay2 := routing.Relay{
		ID:   crypto.HashID("127.0.0.2:40000"),
		Addr: *addr2,
		Datacenter: routing.Datacenter{
			ID:   2,
			Name: "some name 2",
		},
		PublicKey: make([]byte, crypto.KeySize),
		Seller: routing.Seller{
			ID:   "sellerID1",
			Name: "seller name 1",
		},
		State:   routing.RelayStateEnabled,
		Version: "2.1.0",
	}

	config := getGatewayRelayUpdateHandlerConfig(t, []routing.Relay{relay1, relay2})

	svr := httptest.NewServer(http.HandlerFunc(transport.GatewayRelayUpdateHandlerFunc(config)))
	defer svr.Close()

	updateRequest := &transport.RelayUpdateRequest{
		Version:      5,
		RelayVersion: relay1.Version,
		Address:      relay1.Addr,
		Token:        make([]byte, crypto.KeySize),
	}

	bin, err := updateRequest.MarshalBinary()
	assert.NoError(t, err)

	client := svr.Client()
	res, err := client.Post(svr.URL, "application/octet-stream", bytes.NewBuffer(bin))
	assert.NoError(t, err)

	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, 1, len(config.RequestChan))

	body, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)
	defer res.Body.Close()

	response := &transport.RelayUpdateResponse{}
	err = response.UnmarshalBinary(body)
	assert.NoError(t, err)

	assert.Equal(t, 1, len(response.RelaysToPing))
	assert.Equal(t, relay2.ID, response.RelaysToPing[0].ID)
	assert.Equal(t, relay2.Addr.String(), response.RelaysToPing[0].Address)
}

func TestGatewayRelayUpdate_PingInternalAddr_Success(t *testing.T) {
	t.Parallel()

	addr1, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
	assert.NoError(t, err)
	intAddr1, err := net.ResolveUDPAddr("udp", "10.128.0.1:40000")
	assert.NoError(t, err)

	relay1 := routing.Relay{
		ID:           crypto.HashID("127.0.0.1:40000"),
		Addr:         *addr1,
		InternalAddr: *intAddr1,
		Datacenter: routing.Datacenter{
			ID:   1,
			Name: "some name",
		},
		PublicKey: make([]byte, crypto.KeySize),
		Seller: routing.Seller{
			ID:   "sellerID1",
			Name: "seller name 1",
		},
		State:   routing.RelayStateEnabled,
		Version: "2.1.0",
	}

	addr2, err := net.ResolveUDPAddr("udp", "127.0.0.2:40000")
	assert.NoError(t, err)
	intAddr2, err := net.ResolveUDPAddr("udp", "10.128.0.2:40000")
	assert.NoError(t, err)

	relay2 := routing.Relay{
		ID:           crypto.HashID("127.0.0.2:40000"),
		Addr:         *addr2,
		InternalAddr: *intAddr2,
		Datacenter: routing.Datacenter{
			ID:   2,
			Name: "some name 2",
		},
		PublicKey: make([]byte, crypto.KeySize),
		Seller: routing.Seller{
			ID:   "sellerID1",
			Name: "seller name 1",
		},
		State:   routing.RelayStateEnabled,
		Version: "2.1.0",
	}

	addr3, err := net.ResolveUDPAddr("udp", "127.0.0.3:40000")
	assert.NoError(t, err)
	intAddr3, err := net.ResolveUDPAddr("udp", "10.128.0.3:40000")
	assert.NoError(t, err)

	relay3 := routing.Relay{
		ID:           crypto.HashID("127.0.0.3:40000"),
		Addr:         *addr3,
		InternalAddr: *intAddr3,
		Datacenter: routing.Datacenter{
			ID:   3,
			Name: "some name 3",
		},
		PublicKey: make([]byte, crypto.KeySize),
		Seller: routing.Seller{
			ID:   "sellerID2",
			Name: "seller name 2",
		},
		State:   routing.RelayStateEnabled,
		Version: "2.1.0",
	}

	config := getGatewayRelayUpdateHandlerConfig(t, []routing.Relay{relay1, relay2, relay3})

	svr := httptest.NewServer(http.HandlerFunc(transport.GatewayRelayUpdateHandlerFunc(config)))
	defer svr.Close()

	updateRequest := &transport.RelayUpdateRequest{
		Version:      5,
		RelayVersion: relay1.Version,
		Address:      relay1.Addr,
		Token:        make([]byte, crypto.KeySize),
	}

	bin, err := updateRequest.MarshalBinary()
	assert.NoError(t, err)

	client := svr.Client()
	res, err := client.Post(svr.URL, "application/octet-stream", bytes.NewBuffer(bin))
	assert.NoError(t, err)

	assert.Equal(t, 200, res.StatusCode)
	assert.Equal(t, 1, len(config.RequestChan))

	body, err := ioutil.ReadAll(res.Body)
	assert.NoError(t, err)
	defer res.Body.Close()

	response := &transport.RelayUpdateResponse{}
	err = response.UnmarshalBinary(body)
	assert.NoError(t, err)

	assert.Equal(t, 2, len(response.RelaysToPing))
	assert.Equal(t, relay2.ID, response.RelaysToPing[0].ID)
	assert.Equal(t, relay2.InternalAddr.String(), response.RelaysToPing[0].Address)
	assert.Equal(t, relay3.ID, response.RelaysToPing[1].ID)
	assert.Equal(t, relay3.Addr.String(), response.RelaysToPing[1].Address)
}
