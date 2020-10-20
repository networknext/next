package analytics_test

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"testing"

	"cloud.google.com/go/pubsub"
	"github.com/go-kit/kit/log"
	"github.com/networknext/backend/analytics"
	"github.com/networknext/backend/metrics"
	"github.com/networknext/backend/routing"
	"github.com/stretchr/testify/assert"
)

func checkGooglePubsubEmulator(t *testing.T) {
	pubsubEmulatorHost := os.Getenv("PUBSUB_EMULATOR_HOST")
	if pubsubEmulatorHost == "" {
		t.Skip("Pub/Sub emulator not set up, skipping analytics pub/sub tests")
	}
}

func TestNewGooglePubSubPublisher(t *testing.T) {
	checkGooglePubsubEmulator(t)
	_, err := analytics.NewGooglePubSubPingStatsPublisher(context.Background(), &metrics.EmptyAnalyticsMetrics, log.NewNopLogger(), "default", "analytics", pubsub.DefaultPublishSettings)
	assert.NoError(t, err)
}

func TestGooglePubSubPublisher(t *testing.T) {
	checkGooglePubsubEmulator(t)
	ctx := context.Background()

	t.Run("uninitialized writing client", func(t *testing.T) {
		publisher := &analytics.GooglePubSubPingStatsPublisher{}
		err := publisher.Publish(ctx, []analytics.PingStatsEntry{})
		assert.EqualError(t, err, "analytics: client not initialized")
	})

	t.Run("success", func(t *testing.T) {
		publisher, err := analytics.NewGooglePubSubPingStatsPublisher(ctx, &metrics.EmptyAnalyticsMetrics, log.NewNopLogger(), "default", "analytics", pubsub.DefaultPublishSettings)
		assert.NoError(t, err)
		err = publisher.Publish(ctx, []analytics.PingStatsEntry{})
		assert.NoError(t, err)
	})
}

func TestLocalBigQueryWriter(t *testing.T) {
	t.Run("no logger", func(t *testing.T) {
		writer := analytics.LocalPingStatsWriter{}
		err := writer.Write(context.Background(), &analytics.PingStatsEntry{})
		assert.EqualError(t, err, "no logger for local big query writer, can't display entry")
	})

	t.Run("success", func(t *testing.T) {
		writer := analytics.LocalPingStatsWriter{
			Logger: log.NewNopLogger(),
		}

		err := writer.Write(context.Background(), &analytics.PingStatsEntry{})
		assert.NoError(t, err)
	})
}

func TestNoOp(t *testing.T) {
	t.Run("pubsub", func(t *testing.T) {
		publisher := analytics.NoOpPingStatsPublisher{}
		publisher.Publish(context.Background(), []analytics.PingStatsEntry{})
	})

	t.Run("bigquery", func(t *testing.T) {
		writer := analytics.NoOpPingStatsWriter{}
		writer.Write(context.Background(), &analytics.PingStatsEntry{})
	})
}

func TestExtractPingStats(t *testing.T) {
	numRelays := 10

	statsdb := routing.NewStatsDatabase()

	for i := 0; i < numRelays; i++ {
		var update routing.RelayStatsUpdate
		update.ID = uint64(i)
		update.PingStats = make([]routing.RelayStatsPing, numRelays-1)

		for j, idx := 0, 0; j < numRelays; j++ {
			if i == j {
				continue
			}

			update.PingStats[idx].RelayID = uint64(j)
			update.PingStats[idx].RTT = rand.Float32()
			update.PingStats[idx].Jitter = rand.Float32()
			update.PingStats[idx].PacketLoss = rand.Float32()

			idx++
		}

		statsdb.ProcessStats(&update)
	}

	pairs := analytics.ExtractPingStats(statsdb)

	assert.Len(t, pairs, numRelays*(numRelays-1)/2)

	for i := 0; i < numRelays; i++ {
		for j := 0; j < numRelays; j++ {
			var expectedTimesFound int
			if i == j {
				// this pair should not be in the list
				expectedTimesFound = 0
			} else {
				// this pair should be in the list only once
				expectedTimesFound = 1
			}

			timesFound := 0

			for k := range pairs {
				pair := &pairs[k]
				if (pair.RelayA == uint64(i) && pair.RelayB == uint64(j)) || (pair.RelayA == uint64(j) && pair.RelayB == uint64(i)) {
					timesFound++
				}
			}

			assert.Equal(t, expectedTimesFound, timesFound, fmt.Sprintf("i = %d, j = %d, pairs = %v", i, j, pairs))
		}
	}
}
