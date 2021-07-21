package analytics_pusher

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/networknext/backend/modules/analytics"
	"github.com/networknext/backend/modules/common/helpers"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/routing"
)

type AnalyticsPusher struct {
	relayStatsPublisher analytics.RelayStatsPublisher
	pingStatsPublisher  analytics.PingStatsPublisher

	relayStatsPublishInterval time.Duration
	pingStatsPublishInterval  time.Duration

	httpClient     http.Client
	routeMatrixURI string
	staleDuration  time.Duration

	metrics *metrics.AnalyticsPusherMetrics
}

func NewAnalyticsPusher(
	relayStatsPublisher analytics.RelayStatsPublisher,
	pingStatsPublisher analytics.PingStatsPublisher,
	relayStatsPublishInterval time.Duration,
	pingStatsPublishInterval time.Duration,
	httpTimeout time.Duration,
	routeMatrixURI string,
	routeMatrixStaleDuration time.Duration,
	metrics *metrics.AnalyticsPusherMetrics,
) (*AnalyticsPusher, error) {

	// Ensure durations are not negative
	if relayStatsPublishInterval < 0 {
		return nil, fmt.Errorf("relay stats publish interval is negative")
	} else if pingStatsPublishInterval < 0 {
		return nil, fmt.Errorf("ping stats publish interval is negative")
	} else if httpTimeout < 0 {
		return nil, fmt.Errorf("http timeout is negative")
	} else if routeMatrixStaleDuration < 0 {
		return nil, fmt.Errorf("route matrix stale duration is negative")
	}

	// Create HTTP client for getting route matrix
	httpClient := http.Client{Timeout: httpTimeout}

	analyticsPusher := &AnalyticsPusher{
		relayStatsPublisher:       relayStatsPublisher,
		pingStatsPublisher:        pingStatsPublisher,
		relayStatsPublishInterval: relayStatsPublishInterval,
		pingStatsPublishInterval:  pingStatsPublishInterval,
		httpClient:                httpClient,
		routeMatrixURI:            routeMatrixURI,
		staleDuration:             routeMatrixStaleDuration,
        metrics:                   metrics,
	}

	return analyticsPusher, nil
}

func (ap *AnalyticsPusher) Start(ctx context.Context, wg *sync.WaitGroup, errChan chan error) error {
    defer wg.Done()

	// Start the relay stats publishing goroutine
	wg.Add(1)
	ap.StartRelayStatsPublisher(ctx, wg, errChan)

	// Start the ping stats publishing goroutine
	wg.Add(1)
	ap.StartPingStatsPublisher(ctx, wg, errChan)

	return nil
}

func (ap *AnalyticsPusher) StartRelayStatsPublisher(ctx context.Context, wg *sync.WaitGroup, errChan chan error) {
	defer wg.Done()

	syncTimer := helpers.NewSyncTimer(ap.relayStatsPublishInterval)
	for {
		syncTimer.Run()
		select {
		case <-ctx.Done():
			return
		default:
			routeMatrix, err := ap.getRouteMatrix()
			if err != nil {
				core.Error("error getting route matrix: %v", err)
				continue
			}

			numRelayStats := len(routeMatrix.RelayStats)

			core.Debug("Number of relay stats to be published: %d", len(routeMatrix.RelayStats))
			if numRelayStats > 0 {

				ap.metrics.RelayStatsMetrics.EntriesReceived.Add(float64(numRelayStats))

				if err := ap.relayStatsPublisher.Publish(ctx, routeMatrix.RelayStats); err != nil {
					core.Error("error publishing relay stats: %v", err)
					errChan <- err
				}
			}
		}
	}
}

func (ap *AnalyticsPusher) StartPingStatsPublisher(ctx context.Context, wg *sync.WaitGroup, errChan chan error) {
	defer wg.Done()

	syncTimer := helpers.NewSyncTimer(ap.pingStatsPublishInterval)
	for {
		syncTimer.Run()
		select {
		case <-ctx.Done():
			return
		default:
			routeMatrix, err := ap.getRouteMatrix()
			if err != nil {
                core.Error("error getting route matrix: %v", err)
				continue
			}

			numPingStats := len(routeMatrix.PingStats)

			core.Debug("Number of ping stats to be published: %d", numPingStats)
			if numPingStats > 0 {

				ap.metrics.PingStatsMetrics.EntriesReceived.Add(float64(numPingStats))

				if err := ap.pingStatsPublisher.Publish(ctx, routeMatrix.PingStats); err != nil {
					core.Error("error publishing ping stats: %v", err)
					errChan <- err
				}
			}
		}
	}
}

func (ap *AnalyticsPusher) getRouteMatrix() (*routing.RouteMatrix, error) {
	ap.metrics.RouteMatrixInvocations.Add(1)

    var err error
	var buffer []byte
	start := time.Now()

	var routeMatrixReader io.ReadCloser

	if f, err := os.Open(ap.routeMatrixURI); err == nil {
		routeMatrixReader = f
	}

	if r, err := ap.httpClient.Get(ap.routeMatrixURI); err == nil {
		routeMatrixReader = r.Body
	}

	if routeMatrixReader == nil {
		err = fmt.Errorf("route matrix reader is nil")
		ap.metrics.ErrorMetrics.RouteMatrixReaderNil.Add(1)
		return &routing.RouteMatrix{}, err
	}

	buffer, err = ioutil.ReadAll(routeMatrixReader)

	routeMatrixReader.Close()

	if err != nil {
		err = fmt.Errorf("failed to read route matrix data: %v", err)
		ap.metrics.ErrorMetrics.RouteMatrixReadFailure.Add(1)
		return &routing.RouteMatrix{}, err

	}

	if len(buffer) == 0 {
		err = fmt.Errorf("route matrix buffer is empty")
		ap.metrics.ErrorMetrics.RouteMatrixBufferEmpty.Add(1)
		return &routing.RouteMatrix{}, err
	}

	var newRouteMatrix routing.RouteMatrix
	readStream := encoding.CreateReadStream(buffer)
	if err := newRouteMatrix.Serialize(readStream); err != nil {
		err = fmt.Errorf("failed to serialize route matrix: %v", err)
		ap.metrics.ErrorMetrics.RouteMatrixSerializeFailure.Add(1)
		return &routing.RouteMatrix{}, err
	}

	if newRouteMatrix.CreatedAt+uint64(ap.staleDuration.Seconds()) < uint64(time.Now().Unix()) {
		err = fmt.Errorf("route matrix is stale")
		ap.metrics.ErrorMetrics.StaleRouteMatrix.Add(1)
		return &routing.RouteMatrix{}, err
	}

	routeEntriesTime := time.Since(start)
	duration := float64(routeEntriesTime.Milliseconds())
	ap.metrics.RouteMatrixUpdateDuration.Set(duration)
	if duration > 250 {
		core.Error("long route matrix duration %dms", int(duration))
		ap.metrics.RouteMatrixUpdateLongDuration.Add(1)
	}

	ap.metrics.RouteMatrixSuccess.Add(1)

	return &newRouteMatrix, nil
}
