package metrics_test

import (
	"context"
	"testing"
	"github.com/stretchr/testify/assert"
	"github.com/networknext/backend/modules/metrics"
)

func TestLocalMetrics(t *testing.T) {
	// Test metric creation
	{
		localHandler := &metrics.LocalHandler{}

		counter, err := localHandler.NewCounter(context.Background(), &metrics.Descriptor{ID: "test-metric"})
		assert.NoError(t, err)
		assert.NotNil(t, counter)

		gauge, err := localHandler.NewGauge(context.Background(), &metrics.Descriptor{ID: "test-metric"})
		assert.NoError(t, err)
		assert.NotNil(t, gauge)

		histogram, err := localHandler.NewHistogram(context.Background(), &metrics.Descriptor{ID: "test-metric"}, 50)
		assert.NoError(t, err)
		assert.NotNil(t, histogram)
	}
}
