package relay_gateway

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/modules/metrics"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

type GatewayConfig struct {
	ChannelBufferSize     int
	BinSyncInterval       time.Duration
	UseHTTP               bool
	RelayBackendAddresses []string
	HTTPTimeout           time.Duration
	BatchSize             int
	NumGoroutines         int

	PublishToHosts        []string
	PublisherSendBuffer   int
	PublisherRefreshTimer time.Duration
}

type GatewayHTTPClient struct {
	Cfg *GatewayConfig

	updateChan     chan []byte
	gatewayMetrics *metrics.RelayGatewayMetrics
	client         *http.Client
	logger         log.Logger
}

func NewGatewayHTTPClient(cfg *GatewayConfig, updateChan chan []byte, gatewayMetrics *metrics.RelayGatewayMetrics, logger log.Logger) (*GatewayHTTPClient, error) {
	// Create HTTP client to communicate with relay backends
	client := &http.Client{Timeout: cfg.HTTPTimeout}

	return &GatewayHTTPClient{
		Cfg:            cfg,
		updateChan:     updateChan,
		gatewayMetrics: gatewayMetrics,
		client:         client,
		logger:         logger,
	}, nil
}

// Starts goroutines for batch-sending relay update requests to the relay backends
func (httpClient *GatewayHTTPClient) Start(ctx context.Context) error {
	var wg sync.WaitGroup

	// Create worker goroutines to pull updates from the update channel
	for i := 0; i < httpClient.Cfg.NumGoroutines; i++ {
		wg.Add(1)

		// Handle update requests
		go func() {
			defer wg.Done()
			// Create buffers to store updates for batch-sending
			var updateBufferMessageCount int
			var updateBufferMutex sync.Mutex
			updateBuffer := make([]byte, 0)

			for {
				select {
				case <-ctx.Done():
					// TODO: flush all messages to relay backends
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
							go httpClient.PostRelayUpdate(bufferCopy, address)
						}

						// Set the number of relay update requests sent to the relay backends (not necessarily successful)
						httpClient.gatewayMetrics.UpdatesFlushed.Add(float64(numUpdatesFlushed))
						level.Info(httpClient.logger).Log("msg", fmt.Sprintf("Sent %d relay updates to the relay backends", numUpdatesFlushed))
					}
				}
			}
		}()
	}

	// Wait until either there is an error or the context is done
	select {
	case <-ctx.Done():
		// Let the goroutines finish up
		wg.Wait()
		return ctx.Err()
	}
}

func (httpClient *GatewayHTTPClient) PostRelayUpdate(bufferCopy []byte, address string) {
	// Post to relay backend
	buffer := bytes.NewBuffer(bufferCopy)

	resp, err := httpClient.client.Post(fmt.Sprintf("http://%s/relay_update", address), "application/octet-stream", buffer)
	if err != nil || resp.StatusCode != http.StatusOK {
		httpClient.gatewayMetrics.ErrorMetrics.BackendSendFailure.Add(1)
		level.Error(httpClient.logger).Log("msg", fmt.Sprintf("unable to send update to relay backend %s, response was non-200", address), "err", err)
		return
	} else if resp.StatusCode != http.StatusOK {
		httpClient.gatewayMetrics.ErrorMetrics.BackendSendFailure.Add(1)
		level.Error(httpClient.logger).Log("msg", fmt.Sprintf("unable to send update to relay backend %s, response was %d", address, resp.StatusCode), "err", err)
		return
	}
	resp.Body.Close()
}
