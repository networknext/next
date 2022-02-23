package match_data_test

import (
	"context"
	"os"
	"testing"

	"cloud.google.com/go/pubsub"

	md "github.com/networknext/backend/modules/match_data"
	"github.com/networknext/backend/modules/metrics"
	"github.com/stretchr/testify/assert"
)

func checkGooglePubsubEmulator(t *testing.T) {
	pubsubEmulatorHost := os.Getenv("PUBSUB_EMULATOR_HOST")
	if pubsubEmulatorHost == "" {
		t.Skip("Pub/Sub emulator not set up, skipping match_data pub/sub tests")
	}
}

func TestNewGooglePubSubMatcher(t *testing.T) {
	checkGooglePubsubEmulator(t)

	t.Run("no publish settings", func(t *testing.T) {
		_, err := md.NewGooglePubSubMatcher(context.Background(), &metrics.EmptyMatchDataMetrics, "", "", 0, 0, 0, nil)
		assert.EqualError(t, err, "nil google pubsub publish settings")
	})

	t.Run("success", func(t *testing.T) {
		_, err := md.NewGooglePubSubMatcher(context.Background(), &metrics.EmptyMatchDataMetrics, "default", "match_data", 1, 0, 0, &pubsub.DefaultPublishSettings)
		assert.NoError(t, err)
	})
}

func TestGooglePubSubMatch(t *testing.T) {
	checkGooglePubsubEmulator(t)
	ctx := context.Background()

	t.Run("uninitialized match data clients", func(t *testing.T) {
		matcher := &md.GooglePubSubMatcher{}
		err := matcher.Match(ctx, &md.MatchDataEntry{})
		assert.EqualError(t, err, "match_data: clients not initialized")
	})

	t.Run("success", func(t *testing.T) {
		matcher, err := md.NewGooglePubSubMatcher(context.Background(), &metrics.EmptyMatchDataMetrics, "default", "match_data", 1, 0, 0, &pubsub.DefaultPublishSettings)
		assert.NoError(t, err)

		err = matcher.Match(ctx, &md.MatchDataEntry{})
		assert.NoError(t, err)
	})
}

func TestLocalMatch(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		matcher := md.LocalMatcher{
			Metrics: &metrics.EmptyMatchDataMetrics,
		}

		err := matcher.Match(context.Background(), &md.MatchDataEntry{})
		assert.NoError(t, err)
	})
}

func TestNoOpMatch(t *testing.T) {
	matcher := md.NoOpMatcher{}
	err := matcher.Match(context.Background(), nil)
	assert.NoError(t, err)
}
