package analytics_pusher_test

import (
	"context"
	"os"
	"sync"
	"testing"
	"time"

	"cloud.google.com/go/pubsub"

	"github.com/networknext/backend/modules/analytics"
	pusher "github.com/networknext/backend/modules/analytics_pusher"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/routing"

	"github.com/stretchr/testify/assert"
)

func checkGooglePubsubEmulator(t *testing.T) {
	pubsubEmulatorHost := os.Getenv("PUBSUB_EMULATOR_HOST")
	if pubsubEmulatorHost == "" {
		t.Skip("Pub/Sub emulator not set up, skipping analytics pub/sub tests")
	}
}

func TestCreateAnalyticsPusher(t *testing.T) {
	t.Parallel()

	t.Run("negative relay stats publish interval", func(t *testing.T) {

		var relayStatsPublisher analytics.RelayStatsPublisher = &analytics.NoOpRelayStatsPublisher{}
		var pingStatsPublisher analytics.PingStatsPublisher = &analytics.NoOpPingStatsPublisher{}
		relayStatsPublishInterval := -1 * time.Second

		_, err := pusher.NewAnalyticsPusher(relayStatsPublisher, pingStatsPublisher, relayStatsPublishInterval, time.Second, time.Second, "", time.Second, metrics.EmptyAnalyticsPusherMetrics)
		assert.EqualError(t, err, "relay stats publish interval is negative")
	})

	t.Run("negative ping stats publish interval", func(t *testing.T) {

		var relayStatsPublisher analytics.RelayStatsPublisher = &analytics.NoOpRelayStatsPublisher{}
		var pingStatsPublisher analytics.PingStatsPublisher = &analytics.NoOpPingStatsPublisher{}
		pingStatsPublishInterval := -1 * time.Second

		_, err := pusher.NewAnalyticsPusher(relayStatsPublisher, pingStatsPublisher, time.Second, pingStatsPublishInterval, time.Second, "", time.Second, metrics.EmptyAnalyticsPusherMetrics)
		assert.EqualError(t, err, "ping stats publish interval is negative")
	})

	t.Run("negative http timeout", func(t *testing.T) {

		var relayStatsPublisher analytics.RelayStatsPublisher = &analytics.NoOpRelayStatsPublisher{}
		var pingStatsPublisher analytics.PingStatsPublisher = &analytics.NoOpPingStatsPublisher{}
		httpTimeout := -1 * time.Second

		_, err := pusher.NewAnalyticsPusher(relayStatsPublisher, pingStatsPublisher, time.Second, time.Second, httpTimeout, "", time.Second, metrics.EmptyAnalyticsPusherMetrics)
		assert.EqualError(t, err, "http timeout is negative")
	})

	t.Run("negative route matrix stale duration", func(t *testing.T) {

		var relayStatsPublisher analytics.RelayStatsPublisher = &analytics.NoOpRelayStatsPublisher{}
		var pingStatsPublisher analytics.PingStatsPublisher = &analytics.NoOpPingStatsPublisher{}
		routeMatrixStaleDuration := -1 * time.Second

		_, err := pusher.NewAnalyticsPusher(relayStatsPublisher, pingStatsPublisher, time.Second, time.Second, time.Second, "", routeMatrixStaleDuration, metrics.EmptyAnalyticsPusherMetrics)
		assert.EqualError(t, err, "route matrix stale duration is negative")
	})

	t.Run("success", func(t *testing.T) {

		var relayStatsPublisher analytics.RelayStatsPublisher = &analytics.NoOpRelayStatsPublisher{}
		var pingStatsPublisher analytics.PingStatsPublisher = &analytics.NoOpPingStatsPublisher{}

		ap, err := pusher.NewAnalyticsPusher(relayStatsPublisher, pingStatsPublisher, time.Second, time.Second, time.Second, "", time.Second, metrics.EmptyAnalyticsPusherMetrics)
		assert.NoError(t, err)
		assert.NotNil(t, ap)
	})
}

// We are unable to test the following errors since it would require disk I/O
//	- RouteMatrixReadFailure
//	- RouteMatrixBufferEmpty
//  - RouteMatrixSerializeFailure
func TestGetRouteMatrix(t *testing.T) {

	t.Parallel()

	var relayStatsPublisher analytics.RelayStatsPublisher = &analytics.NoOpRelayStatsPublisher{}
	var pingStatsPublisher analytics.PingStatsPublisher = &analytics.NoOpPingStatsPublisher{}

	t.Run("route matrix reader nil", func(t *testing.T) {
		pusherMetrics, err := metrics.NewAnalyticsPusherMetrics(context.Background(), &metrics.LocalHandler{}, "analytics_pusher")
		assert.NoError(t, err)

		ap, err := pusher.NewAnalyticsPusher(relayStatsPublisher, pingStatsPublisher, time.Second, time.Second, time.Second, "", time.Second, pusherMetrics)
		assert.NoError(t, err)

		rm, err := ap.GetRouteMatrix()
		assert.Equal(t, rm, &routing.RouteMatrix{})
		assert.EqualError(t, err, "route matrix reader is nil")
		assert.Equal(t, float64(1), pusherMetrics.ErrorMetrics.RouteMatrixReaderNil.Value())
	})

	t.Run("route matrix stale", func(t *testing.T) {
		pusherMetrics, err := metrics.NewAnalyticsPusherMetrics(context.Background(), &metrics.LocalHandler{}, "analytics_pusher")
		assert.NoError(t, err)

		ap, err := pusher.NewAnalyticsPusher(relayStatsPublisher, pingStatsPublisher, time.Second, time.Second, time.Second, "../../testdata/route_matrix_dev.bin", time.Second, pusherMetrics)
		assert.NoError(t, err)

		rm, err := ap.GetRouteMatrix()
		assert.Equal(t, rm, &routing.RouteMatrix{})
		assert.EqualError(t, err, "route matrix is stale")
		assert.Equal(t, float64(1), pusherMetrics.ErrorMetrics.StaleRouteMatrix.Value())
	})

	t.Run("success", func(t *testing.T) {
		pusherMetrics, err := metrics.NewAnalyticsPusherMetrics(context.Background(), &metrics.LocalHandler{}, "analytics_pusher")
		assert.NoError(t, err)

		ap, err := pusher.NewAnalyticsPusher(relayStatsPublisher, pingStatsPublisher, time.Second, time.Second, time.Second, "../../testdata/route_matrix_dev.bin", time.Hour*87600, pusherMetrics)
		assert.NoError(t, err)

		rm, err := ap.GetRouteMatrix()
		assert.NotEqual(t, rm, &routing.RouteMatrix{})
		assert.NoError(t, err)

	})
}

func TestStartRelayStatsPublisher(t *testing.T) {

	t.Run("error getting route matrix", func(t *testing.T) {
		pusherMetrics, err := metrics.NewAnalyticsPusherMetrics(context.Background(), &metrics.LocalHandler{}, "analytics_pusher")
		assert.NoError(t, err)

		var relayStatsPublisher analytics.RelayStatsPublisher = &analytics.NoOpRelayStatsPublisher{}
		var pingStatsPublisher analytics.PingStatsPublisher = &analytics.NoOpPingStatsPublisher{}

		ap, err := pusher.NewAnalyticsPusher(relayStatsPublisher, pingStatsPublisher, time.Millisecond, time.Second, time.Second, "", time.Second, pusherMetrics)
		assert.NoError(t, err)

		ctx, cancelFunc := context.WithCancel(context.Background())
		wg := &sync.WaitGroup{}
		errChan := make(chan error, 1)

		wg.Add(1)
		go ap.StartRelayStatsPublisher(ctx, wg, errChan)

		time.Sleep(time.Millisecond * 200)
		cancelFunc()
		wg.Wait()

		assert.Greater(t, pusherMetrics.ErrorMetrics.RouteMatrixReaderNil.Value(), float64(0))
		assert.Equal(t, 0, len(errChan))
	})

	t.Run("success", func(t *testing.T) {
		checkGooglePubsubEmulator(t)

		pusherMetrics, err := metrics.NewAnalyticsPusherMetrics(context.Background(), &metrics.LocalHandler{}, "analytics_pusher")
		assert.NoError(t, err)

		var pingStatsPublisher analytics.PingStatsPublisher = &analytics.NoOpPingStatsPublisher{}
		relayStatsPublisher, err := analytics.NewGooglePubSubRelayStatsPublisher(context.Background(), &pusherMetrics.RelayStatsMetrics, "local", "analytics", pubsub.DefaultPublishSettings)
		assert.NoError(t, err)

		ap, err := pusher.NewAnalyticsPusher(relayStatsPublisher, pingStatsPublisher, time.Millisecond, time.Second, time.Second, "../../testdata/route_matrix_dev.bin", time.Hour*87600, pusherMetrics)
		assert.NoError(t, err)

		ctx, cancelFunc := context.WithCancel(context.Background())
		wg := &sync.WaitGroup{}
		errChan := make(chan error, 1)

		wg.Add(1)
		go ap.StartRelayStatsPublisher(ctx, wg, errChan)

		time.Sleep(time.Millisecond * 200)
		cancelFunc()
		wg.Wait()

		assert.Greater(t, pusherMetrics.RelayStatsMetrics.EntriesReceived.Value(), float64(0))
		assert.Greater(t, pusherMetrics.RelayStatsMetrics.EntriesSubmitted.Value(), float64(0))
		assert.Greater(t, pusherMetrics.RelayStatsMetrics.EntriesFlushed.Value(), float64(0))
		assert.Equal(t, 0, len(errChan))
	})
}

func TestStartPingStatsPublisher(t *testing.T) {

	t.Run("error getting route matrix", func(t *testing.T) {
		pusherMetrics, err := metrics.NewAnalyticsPusherMetrics(context.Background(), &metrics.LocalHandler{}, "analytics_pusher")
		assert.NoError(t, err)

		var relayStatsPublisher analytics.RelayStatsPublisher = &analytics.NoOpRelayStatsPublisher{}
		var pingStatsPublisher analytics.PingStatsPublisher = &analytics.NoOpPingStatsPublisher{}

		ap, err := pusher.NewAnalyticsPusher(relayStatsPublisher, pingStatsPublisher, time.Second, time.Millisecond, time.Second, "", time.Second, pusherMetrics)
		assert.NoError(t, err)

		ctx, cancelFunc := context.WithCancel(context.Background())
		wg := &sync.WaitGroup{}
		errChan := make(chan error, 1)

		wg.Add(1)
		go ap.StartPingStatsPublisher(ctx, wg, errChan)

		checks := 4
		for {
			time.Sleep(time.Millisecond * 200)

			if checks <= 0 {
				break
			}

			if pusherMetrics.ErrorMetrics.RouteMatrixReaderNil.Value() > 0 && len(errChan) == 0 {
				break
			}

			checks--
		}

		cancelFunc()
		wg.Wait()

		assert.Greater(t, pusherMetrics.ErrorMetrics.RouteMatrixReaderNil.Value(), float64(0))
		assert.Equal(t, 0, len(errChan))
	})

	t.Run("success", func(t *testing.T) {
		checkGooglePubsubEmulator(t)

		pusherMetrics, err := metrics.NewAnalyticsPusherMetrics(context.Background(), &metrics.LocalHandler{}, "analytics_pusher")
		assert.NoError(t, err)

		var relayStatsPublisher analytics.RelayStatsPublisher = &analytics.NoOpRelayStatsPublisher{}
		pingStatsPublisher, err := analytics.NewGooglePubSubPingStatsPublisher(context.Background(), &pusherMetrics.PingStatsMetrics, "local", "analytics", pubsub.DefaultPublishSettings)
		assert.NoError(t, err)

		ap, err := pusher.NewAnalyticsPusher(relayStatsPublisher, pingStatsPublisher, time.Second, time.Millisecond, time.Second, "../../testdata/route_matrix_dev.bin", time.Hour*87600, pusherMetrics)
		assert.NoError(t, err)

		ctx, cancelFunc := context.WithCancel(context.Background())
		wg := &sync.WaitGroup{}
		errChan := make(chan error, 1)

		wg.Add(1)
		go ap.StartPingStatsPublisher(ctx, wg, errChan)

		checks := 4
		for {
			time.Sleep(time.Millisecond * 200)

			if checks <= 0 {
				break
			}

			if pusherMetrics.PingStatsMetrics.EntriesReceived.Value() > 0 &&
				pusherMetrics.PingStatsMetrics.EntriesSubmitted.Value() > 0 &&
				pusherMetrics.PingStatsMetrics.ErrorMetrics.PublishFailure.Value() == 0 &&
				len(errChan) == 0 {
				break
			}

			checks--
		}

		cancelFunc()
		wg.Wait()

		assert.Greater(t, pusherMetrics.PingStatsMetrics.EntriesReceived.Value(), float64(0))
		assert.Greater(t, pusherMetrics.PingStatsMetrics.EntriesSubmitted.Value(), float64(0))
		assert.Zero(t, pusherMetrics.PingStatsMetrics.ErrorMetrics.PublishFailure.Value())
		assert.Equal(t, 0, len(errChan))
	})
}
