package transport_test

import (
    "bytes"
    "context"
    "net"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/networknext/backend/modules/core"
    "github.com/networknext/backend/modules/encoding"

    "github.com/networknext/backend/modules-old/crypto"
    "github.com/networknext/backend/modules-old/metrics"
    "github.com/networknext/backend/modules-old/routing"
    "github.com/networknext/backend/modules-old/transport"

    "github.com/stretchr/testify/assert"
)

func getRelayUpdateHandlerConfig(t *testing.T, relays []routing.Relay) *transport.RelayUpdateHandlerConfig {
    statsdb := routing.NewStatsDatabase()

    cleanupCallback := func(relayData routing.RelayData) error {
        statsdb.DeleteEntry(relayData.ID)
        return nil
    }

    relayMap := routing.NewRelayMap(cleanupCallback)

    relayUpdateMetrics, err := metrics.NewRelayUpdateMetrics(context.Background(), &metrics.LocalHandler{})
    assert.NoError(t, err)

    relayHash := make(map[uint64]routing.Relay)
    for _, relay := range relays {
        relayHash[relay.ID] = relay
    }

    getRelayData := func() ([]routing.Relay, map[uint64]routing.Relay) {
        return relays, relayHash
    }

    return &transport.RelayUpdateHandlerConfig{
        RelayMap:     relayMap,
        StatsDB:      statsdb,
        Metrics:      relayUpdateMetrics,
        GetRelayData: getRelayData,
    }
}

func TestRelayUpdateHandlerFunc_ContentTypeFailure(t *testing.T) {
    t.Parallel()

    config := getRelayUpdateHandlerConfig(t, []routing.Relay{})

    svr := httptest.NewServer(http.HandlerFunc(transport.RelayUpdateHandlerFunc(config)))
    defer svr.Close()

    client := svr.Client()
    res, err := client.Post(svr.URL, "application/json", nil)
    assert.NoError(t, err)

    assert.Equal(t, 400, res.StatusCode)
    assert.Equal(t, float64(1), config.Metrics.ErrorMetrics.ContentTypeFailure.Value())
}

func TestRelayUpdateHandlerFunc_UnbatchFailure(t *testing.T) {
    t.Parallel()

    config := getRelayUpdateHandlerConfig(t, []routing.Relay{})

    svr := httptest.NewServer(http.HandlerFunc(transport.RelayUpdateHandlerFunc(config)))
    defer svr.Close()

    addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
    assert.NoError(t, err)

    updateRequest := &transport.RelayUpdateRequest{
        Version:      5,
        RelayVersion: "2.1.0",
        Address:      *addr,
        Token:        make([]byte, crypto.KeySize),
    }

    // Sending update request without offset causes unbatch failure
    bin, err := updateRequest.MarshalBinary()
    assert.NoError(t, err)

    client := svr.Client()
    res, err := client.Post(svr.URL, "application/octet-stream", bytes.NewBuffer(bin))
    assert.NoError(t, err)

    assert.Equal(t, 400, res.StatusCode)
    assert.Equal(t, float64(1), config.Metrics.ErrorMetrics.UnbatchFailure.Value())
}

func TestRelayUpdateHandlerFunc_UnmarshalFailure(t *testing.T) {
    t.Parallel()

    config := getRelayUpdateHandlerConfig(t, []routing.Relay{})

    svr := httptest.NewServer(http.HandlerFunc(transport.RelayUpdateHandlerFunc(config)))
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

    // Update the version to cause unmarshal failure
    bin = append([]byte{0, 0, 0, 0}, bin[4:]...)

    // Create a byte slice with an offset
    var offset int
    data := make([]byte, 4+len(bin))
    encoding.WriteUint32(data, &offset, uint32(len(bin)))
    encoding.WriteBytes(data, &offset, bin, len(bin))

    client := svr.Client()
    res, err := client.Post(svr.URL, "application/octet-stream", bytes.NewBuffer(data))
    assert.NoError(t, err)

    assert.Equal(t, 200, res.StatusCode)
    assert.Equal(t, float64(1), config.Metrics.ErrorMetrics.UnmarshalFailure.Value())
}

func TestRelayUpdateHandlerFunc_RelayNotFound(t *testing.T) {
    t.Parallel()

    config := getRelayUpdateHandlerConfig(t, []routing.Relay{})

    svr := httptest.NewServer(http.HandlerFunc(transport.RelayUpdateHandlerFunc(config)))
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

    // Create a byte slice with an offset
    var offset int
    data := make([]byte, 4+len(bin))
    encoding.WriteUint32(data, &offset, uint32(len(bin)))
    encoding.WriteBytes(data, &offset, bin, len(bin))

    client := svr.Client()
    res, err := client.Post(svr.URL, "application/octet-stream", bytes.NewBuffer(data))
    assert.NoError(t, err)

    assert.Equal(t, 200, res.StatusCode)
    assert.Equal(t, float64(1), config.Metrics.ErrorMetrics.RelayNotFound.Value())
}

func TestRelayUpdateHandlerFunc_Success(t *testing.T) {
    t.Parallel()

    addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
    assert.NoError(t, err)

    relay1 := routing.Relay{
        ID:   crypto.HashID("127.0.0.1:40000"),
        Name: "relay1",
        Addr: *addr,
        Datacenter: routing.Datacenter{
            ID:   1,
            Name: "some name",
        },
        NICSpeedMbps:     int32(1000),
        MaxBandwidthMbps: int32(900),
        PublicKey:        make([]byte, crypto.KeySize),
        Seller: routing.Seller{
            ID:   "sellerID",
            Name: "seller name",
        },
        State:       routing.RelayStateEnabled,
        MaxSessions: uint32(10),

        Version: "2.1.0",
    }

    config := getRelayUpdateHandlerConfig(t, []routing.Relay{relay1})

    svr := httptest.NewServer(http.HandlerFunc(transport.RelayUpdateHandlerFunc(config)))
    defer svr.Close()

    updateRequest := &transport.RelayUpdateRequest{
        Version:           5,
        RelayVersion:      relay1.Version,
        Address:           relay1.Addr,
        Token:             make([]byte, crypto.KeySize),
        PingStats:         make([]routing.RelayStatsPing, core.MaxNearRelays),
        SessionCount:      uint64(5),
        ShuttingDown:      false,
        CPU:               uint8(16),
        EnvelopeUpKbps:    uint64(200),
        EnvelopeDownKbps:  uint64(200),
        BandwidthSentKbps: uint64(100),
        BandwidthRecvKbps: uint64(100),
    }

    bin, err := updateRequest.MarshalBinary()
    assert.NoError(t, err)

    // Create a byte slice with an offset
    var offset int
    data := make([]byte, 4+len(bin))
    encoding.WriteUint32(data, &offset, uint32(len(bin)))
    encoding.WriteBytes(data, &offset, bin, len(bin))

    client := svr.Client()
    res, err := client.Post(svr.URL, "application/octet-stream", bytes.NewBuffer(data))
    assert.NoError(t, err)

    assert.Equal(t, 200, res.StatusCode)

    relayData, exists := config.RelayMap.GetRelayData(addr.String())
    assert.True(t, exists)
    assert.Equal(t, relayData.ID, relay1.ID)
    assert.Equal(t, relayData.Addr, relay1.Addr)
    assert.Equal(t, relayData.Name, relay1.Name)
    assert.Equal(t, relayData.PublicKey, relay1.PublicKey)
    assert.Equal(t, relayData.MaxSessions, relay1.MaxSessions)
    assert.Equal(t, uint64(relayData.SessionCount), updateRequest.SessionCount)
    assert.Equal(t, relayData.ShuttingDown, updateRequest.ShuttingDown)
    assert.Equal(t, relayData.CPU, updateRequest.CPU)
    assert.Equal(t, relayData.NICSpeedMbps, relay1.NICSpeedMbps)
    assert.Equal(t, relayData.MaxBandwidthMbps, relay1.MaxBandwidthMbps)
    assert.Equal(t, relayData.EnvelopeUpMbps, float32(float64(updateRequest.EnvelopeUpKbps)/1000.0))
    assert.Equal(t, relayData.EnvelopeDownMbps, float32(float64(updateRequest.EnvelopeDownKbps)/1000.0))
    assert.Equal(t, relayData.BandwidthSentMbps, float32(float64(updateRequest.BandwidthSentKbps)/1000.0))
    assert.Equal(t, relayData.BandwidthRecvMbps, float32(float64(updateRequest.BandwidthRecvKbps)/1000.0))

    statsEntry, exists := config.StatsDB.Entries[relayData.ID]
    assert.True(t, exists)
    assert.Greater(t, len(statsEntry.Relays), 0)
}
