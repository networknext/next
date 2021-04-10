package relay_gateway

import (
	"context"
	"net/http"
	"sync"

	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/transport"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

type GatewayConfig struct {
	ChannelBufferSize     int
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
	cfg				 GatewayConfig
	updateChan chan []byte
	gatewayMetrics *metrics.RelayGatewayMetrics
	client 		*http.Client
	logger		log.Logger
}

func NewGatewayHTTPClient(cfg *GatewayConfig, updateChan chan []byte, gatewayMetrics *metrics.RelayGatewayMetrics, logger log.Logger) (*GatewayHTTPClient, error) {
	// Create HTTP client to communicate with relay backends
	client := &http.Client{Timeout: cfg.HTTPTimeout}

	return &GatewayHTTPClient{
		cfg: cfg,
		updateChan: updateChan,
		gatewayMetrics: gatewayMetrics,
		client: client,
		logger: logger,
	}, nil
}

// Starts goroutines for batch-sending relay update requests to the relay backends
func (httpClient *GatewayHTTPClient) Start(ctx context.Context) error {
	var wg sync.WaitGroup

	// Create worker goroutines to pull updates from the update channel
	for i := 0; i < httpClient.cfg.NumGoroutines; i++ {
		wg.Add(1)

		// Handle update requests
		go func() {
			defer wg.Done()
			// Create a buffer to hold the requests that will go to the relay backends
			var updateBuffer *transport.RelayUpdateRequestList
			var updateBufferMutex sync.RWMutex

			for {
				select {
				case <-ctx.Done():
					// TODO: flush all messages to relay backends
					return
				case update := <-httpClient.updateChan:
					// Add the update to the buffer and get the buffer's length
					updateBufferMutex.Lock()
					updateBuffer.Requests = append(updateBuffer.Requests, update)
					bufferLength := len(updateBuffer.Requests)
					updateBufferMutex.Unlock()

					// Increment the updates received metric
					httpClient.gatewayMetrics.UpdatesReceived.Add(1)
					// Set the number of updates queued for batch-sending
					httpClient.gatewayMetrics.UpdatesQueued.Set(bufferLength)

					// Check if we have reached the batch size
					if bufferLength >= httpClient.cfg.BatchSize {
						// Copy the buffer so we can clear it without affecting the worker goroutines
						updateBufferMutex.Lock()
						bufferCopy := updateBuffer
						updateBuffer = updateBuffer[:0]
						updateBufferMutex.Unlock()

						// Send the buffer to all relay backends
						for _, address := range httpClient.cfg.RelayBackendAddresses {
							go func() {
								// Marshal the buffer copy
								body, err := bufferCopy.MarshalBinary()
								if err != nil {
									httpClient.gatewayMetrics.ErrorMetrics.MarshalBinaryFailure.Add(1)
									_ = level.Error(httpClient.logger).Log("msg", "unable to marshal buffer copy", "err", err)
									return
								}

								// Post to relay backend
								buffer := bytes.NewBuffer(body)
								resp, err := httpClient.client.Post(fmt.Sprintf("http://%s/relay_update", address), "application/octet-stream", buffer)
								if err != nil || resp.StatusCode != http.StatusOK {
									httpClient.gatewayMetrics.ErrorMetrics.BackendSendFailure.Add(1)
									_ = level.Error(httpClient.logger).Log("msg", fmt.Sprintf("unable to send update to relay backend %s, response %d", address, resp.StatusCode), "err", err)
									return
								}
								resp.Body.Close()
							}()
						}

						// Set the number of relay update requests sent to the relay backends (not necessarily successful)
						httpClient.gatewayMetrics.UpdatesFlushed.Add(bufferLength)
						level.Info(httpClient.logger).Log("msg", fmt.Sprintf("Sent %d relay updates to the relay backends", bufferLength))
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
