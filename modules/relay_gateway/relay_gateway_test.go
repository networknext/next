package relay_gateway_test

import (
	"context"
	"encoding/base64"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/modules/metrics"
	gateway "github.com/networknext/backend/modules/relay_gateway"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/transport"

	"github.com/stretchr/testify/assert"
)

func TestRelayGateway(t *testing.T) {
	// Create a gateway config
	cfg := &gateway.GatewayConfig{
		ChannelBufferSize:     1,
		UseHTTP:               true,
		RelayBackendAddresses: []string{"127.0.0.1:30000"},
		HTTPTimeout:           time.Second,
		BatchSize:             1,
	}

	testChan := make(chan []byte, 1)
	gatewayMetrics := &metrics.EmptyRelayGatewayMetrics

	g, err := gateway.NewGatewayHTTPClient(cfg, testChan, gatewayMetrics)
	assert.NoError(t, err)
	assert.NotNil(t, g)
	assert.Equal(t, cfg, g.Cfg)
}

func TestRelayGatewayStart(t *testing.T) {
	// Setup gateway config
	cfg := &gateway.GatewayConfig{
		ChannelBufferSize:     5,
		UseHTTP:               true,
		RelayBackendAddresses: []string{"127.0.0.1:30000"},
		HTTPTimeout:           time.Second,
		BatchSize:             2,
	}

	// Create relay update request
	udp, _ := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
	relayPrivateKey, _ := base64.StdEncoding.DecodeString("ZiCSchVFo6T5gJvbQfcwU7yfELsNJaYIC2laQm9DSuA=")
	relayRouterPublicKey, _ := base64.StdEncoding.DecodeString("SS55dEl9nTSnVVDrqwPeqRv/YcYOZZLXCWTpNBIyX0Y=")
	nonce := []byte("123456781234567812345678")
	data := []byte("12345678123456781234567812345678")
	token := crypto.Seal(data, nonce, relayRouterPublicKey, relayPrivateKey)

	updateRequest := transport.RelayUpdateRequest{
		Version:      4,
		RelayVersion: "2.0.8",
		Address:      *udp,
		Token:        token,
		PingStats: []routing.RelayStatsPing{
			{
				RelayID:    0,
				RTT:        1,
				Jitter:     1,
				PacketLoss: 1,
			},
		},
		SessionCount: 0,
		ShuttingDown: false,
		CPU:          0,
	}

	t.Run("test single update", func(t *testing.T) {
		ctx, cancelFunc := context.WithCancel(context.Background())
		defer cancelFunc()

		updateChan := make(chan []byte, cfg.ChannelBufferSize)
		metricsHandler := &metrics.LocalHandler{}
		gatewayMetrics, err := metrics.NewRelayGatewayMetrics(ctx, metricsHandler, "relay_gateway_test", "relay_gateway", "Relay Gateway", "relay update request")
		assert.NoError(t, err)

		g, err := gateway.NewGatewayHTTPClient(cfg, updateChan, gatewayMetrics)
		assert.NoError(t, err)
		assert.NotNil(t, g)

		var wg sync.WaitGroup
		wg.Add(1)
		// Start the goroutine for receiving messages
		go g.Start(ctx, &wg)

		requestBin, err := updateRequest.MarshalBinary()
		assert.NoError(t, err)

		// Mimic the gateway's update handler
		updateChan <- requestBin

		time.Sleep(100 * time.Millisecond)

		// Check the metrics to ensure the update was received
		assert.Equal(t, 1, int(gatewayMetrics.UpdatesReceived.Value()))
		assert.Equal(t, 1, int(gatewayMetrics.UpdatesQueued.Value()))
		assert.Equal(t, 0, int(gatewayMetrics.UpdatesFlushed.Value()))
	})

	t.Run("test full batch update", func(t *testing.T) {
		ctx, cancelFunc := context.WithCancel(context.Background())
		defer cancelFunc()

		updateChan := make(chan []byte, cfg.ChannelBufferSize)
		metricsHandler := &metrics.LocalHandler{}
		gatewayMetrics, err := metrics.NewRelayGatewayMetrics(ctx, metricsHandler, "relay_gateway_test", "relay_gateway", "Relay Gateway", "relay update request")
		assert.NoError(t, err)

		g, err := gateway.NewGatewayHTTPClient(cfg, updateChan, gatewayMetrics)
		assert.NoError(t, err)
		assert.NotNil(t, g)

		var wg sync.WaitGroup
		wg.Add(1)
		// Start the goroutine for receiving messages
		go g.Start(ctx, &wg)

		requestBin, err := updateRequest.MarshalBinary()
		assert.NoError(t, err)
		requestBin2, err := updateRequest.MarshalBinary()
		assert.NoError(t, err)

		// Mimic the gateway's update handler
		updateChan <- requestBin
		updateChan <- requestBin2

		time.Sleep(100 * time.Millisecond)

		// Check the metrics to ensure the update was sent
		// HTTP server isn't set up, so we expect 1 error
		assert.Equal(t, 2, int(gatewayMetrics.UpdatesReceived.Value()))
		assert.Equal(t, 2, int(gatewayMetrics.UpdatesQueued.Value()))
		assert.Equal(t, 2, int(gatewayMetrics.UpdatesFlushed.Value()))
		assert.Equal(t, 0, int(gatewayMetrics.ErrorMetrics.MarshalBinaryFailure.Value()))
		assert.Equal(t, 1, int(gatewayMetrics.ErrorMetrics.BackendSendFailure.Value()))
	})

	t.Run("pack and unpack batched binary", func(t *testing.T) {
		requestBin, err := updateRequest.MarshalBinary()
		assert.NoError(t, err)
		requestBin2, err := updateRequest.MarshalBinary()
		assert.NoError(t, err)

		// Pack updates into updateBuffer
		updateBuffer := make([]byte, 0)
		offset := 0
		data := make([]byte, 4+len(requestBin))
		encoding.WriteUint32(data, &offset, uint32(len(requestBin)))
		encoding.WriteBytes(data, &offset, requestBin, len(requestBin))
		updateBuffer = append(updateBuffer, data...)

		offset = 0
		data = make([]byte, 4+len(requestBin2))
		encoding.WriteUint32(data, &offset, uint32(len(requestBin2)))
		encoding.WriteBytes(data, &offset, requestBin2, len(requestBin2))
		updateBuffer = append(updateBuffer, data...)

		// Unpack updates
		updates := make([][]byte, 0)
		offset = 0
		for {
			if offset >= len(updateBuffer) {
				break
			}

			var updateLength uint32
			var updateRequest []byte
			assert.True(t, encoding.ReadUint32(updateBuffer, &offset, &updateLength))

			assert.True(t, encoding.ReadBytes(updateBuffer, &offset, &updateRequest, updateLength))

			updates = append(updates, updateRequest)
		}
		assert.Equal(t, 2, len(updates))
		assert.Equal(t, requestBin, updates[0])
		assert.Equal(t, requestBin2, updates[1])
	})
}
