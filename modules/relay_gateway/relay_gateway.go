package relay_gateway

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/modules/metrics"
)

type GatewayConfig struct {
	ChannelBufferSize     int
	BinSyncInterval       time.Duration
	MagicPollFrequency    time.Duration
	MagicBackendIP        string
	UseHTTP               bool
	RelayBackendAddresses []string
	HTTPTimeout           time.Duration
	BatchSize             int

	PublishToHosts        []string
	PublisherSendBuffer   int
	PublisherRefreshTimer time.Duration
}

type GatewayHTTPClient struct {
	Cfg *GatewayConfig

	updateChan     chan []byte
	gatewayMetrics *metrics.RelayGatewayMetrics
	client         *http.Client
}

func NewGatewayHTTPClient(cfg *GatewayConfig, updateChan chan []byte, gatewayMetrics *metrics.RelayGatewayMetrics) (*GatewayHTTPClient, error) {
	// Create HTTP client to communicate with relay backends
	client := &http.Client{Timeout: cfg.HTTPTimeout}

	return &GatewayHTTPClient{
		Cfg:            cfg,
		updateChan:     updateChan,
		gatewayMetrics: gatewayMetrics,
		client:         client,
	}, nil
}

// Starts goroutines for batch-sending relay update requests to the relay backends
// NOTE: Start() should be called once to avoid race conditions with the internal buffer
func (httpClient *GatewayHTTPClient) Start(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	// Create buffers to store updates for batch-sending
	var updateBufferMessageCount int
	var updateBufferMutex sync.Mutex
	updateBuffer := make([]byte, 0)

	// Handle update requests
	for {
		select {
		case <-ctx.Done():
			// Flush all remaining updates to the relay backends
			core.Debug("received shutdown signal")

			// Copy the buffer so we can clear it without affecting the worker goroutines
			updateBufferMutex.Lock()

			bufferCopy := updateBuffer
			updateBuffer = make([]byte, 0)
			numUpdatesFlushed := updateBufferMessageCount
			updateBufferMessageCount = 0

			updateBufferMutex.Unlock()

			// Send the buffer to all relay backends
			for _, address := range httpClient.Cfg.RelayBackendAddresses {
				wg.Add(1)
				go httpClient.PostRelayUpdate(bufferCopy, address, wg)
			}

			// Set the number of relay update requests sent to the relay backends (not necessarily successful)
			httpClient.gatewayMetrics.UpdatesFlushed.Add(float64(numUpdatesFlushed))
			core.Debug("sent %d relay updates to the relay backends", numUpdatesFlushed)
			core.Debug("finished shutdown")
			return
		case update := <-httpClient.updateChan:
			// Create a byte slice with an offset
			var offset int
			data := make([]byte, 4+len(update))
			encoding.WriteUint32(data, &offset, uint32(len(update)))
			encoding.WriteBytes(data, &offset, update, len(update))

			// Chain the update to previous updates linked via offset
			updateBufferMutex.Lock()
			updateBuffer = append(updateBuffer, data...)
			updateBufferMessageCount++
			updateBufferMutex.Unlock()

			// Increment the updates received metric
			httpClient.gatewayMetrics.UpdatesReceived.Add(1)
			// Set the number of updates queued for batch-sending
			httpClient.gatewayMetrics.UpdatesQueued.Set(float64(updateBufferMessageCount))

			// Check if we have reached the batch size
			if updateBufferMessageCount >= httpClient.Cfg.BatchSize {
				// Copy the buffer so we can clear it without affecting the worker goroutines
				updateBufferMutex.Lock()

				bufferCopy := updateBuffer
				updateBuffer = make([]byte, 0)
				numUpdatesFlushed := updateBufferMessageCount
				updateBufferMessageCount = 0

				updateBufferMutex.Unlock()

				// Send the buffer to all relay backends
				for _, address := range httpClient.Cfg.RelayBackendAddresses {
					wg.Add(1)
					go httpClient.PostRelayUpdate(bufferCopy, address, wg)
				}

				// Set the number of relay update requests sent to the relay backends (not necessarily successful)
				httpClient.gatewayMetrics.UpdatesFlushed.Add(float64(numUpdatesFlushed))
				core.Debug("sent %d relay updates to the relay backends", numUpdatesFlushed)
			}
		}
	}
}

func (httpClient *GatewayHTTPClient) PostRelayUpdate(bufferCopy []byte, address string, wg *sync.WaitGroup) {
	defer wg.Done()

	// Post to relay backend
	buffer := bytes.NewBuffer(bufferCopy)

	resp, err := httpClient.client.Post(fmt.Sprintf("http://%s/relay_update", address), "application/octet-stream", buffer)
	if err != nil || resp.StatusCode != http.StatusOK {
		httpClient.gatewayMetrics.ErrorMetrics.BackendSendFailure.Add(1)
		core.Error("unabled to send update to relay backend %s, response was non-200: %v", address, err)
		return
	} else if resp.StatusCode != http.StatusOK {
		httpClient.gatewayMetrics.ErrorMetrics.BackendSendFailure.Add(1)
		core.Error("unabled to send update to relay backend %s, response was %d: %v", address, resp.StatusCode, err)
		return
	}

	resp.Body.Close()
}
