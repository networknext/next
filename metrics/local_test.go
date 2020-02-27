package metrics_test

import (
	"context"
	"testing"

	"github.com/networknext/backend/metrics"
	"github.com/stretchr/testify/assert"
)

func TestLocalMetrics(t *testing.T) {
	// Test handle creation
	{
		localHandler := &metrics.LocalHandler{}

		handle, err := localHandler.CreateMetric(context.Background(), &metrics.Descriptor{ID: "test-metric"})
		assert.NoError(t, err)
		assert.NotNil(t, handle)
		assert.Equal(t, &metrics.Descriptor{ID: "test-metric"}, handle.Descriptor)
		assert.NotNil(t, handle.Histogram)
		assert.NotNil(t, handle.Gauge)
	}

	// Test handle retrieval
	{
		localHandler := &metrics.LocalHandler{}

		// Try to get the metric before it's created
		handle, ok := localHandler.GetMetric("test-metric")
		assert.False(t, ok)
		assert.NotNil(t, handle)
		assert.Nil(t, handle.Descriptor)
		assert.Nil(t, handle.Histogram)
		assert.Nil(t, handle.Gauge)

		// Create it and try again
		handle, err := localHandler.CreateMetric(context.Background(), &metrics.Descriptor{ID: "test-metric"})
		assert.NoError(t, err)
		assert.NotNil(t, handle)
		assert.Equal(t, &metrics.Descriptor{ID: "test-metric"}, handle.Descriptor)
		assert.NotNil(t, handle.Histogram)
		assert.NotNil(t, handle.Gauge)

		handle, ok = localHandler.GetMetric("test-metric")
		assert.True(t, ok)
		assert.NotNil(t, handle)
		assert.Equal(t, &metrics.Descriptor{ID: "test-metric"}, handle.Descriptor)
		assert.NotNil(t, handle.Histogram)
		assert.NotNil(t, handle.Gauge)
	}

	// Test handle deletion
	{
		localHandler := &metrics.LocalHandler{}

		// Try to delete the metric before it's created
		err := localHandler.DeleteMetric(context.Background(), &metrics.Descriptor{ID: "test-metric"})
		assert.EqualError(t, err, "Attemped to delete unknown metric test-metric")

		// Create it and try again
		handle, err := localHandler.CreateMetric(context.Background(), &metrics.Descriptor{ID: "test-metric"})
		assert.NoError(t, err)
		assert.NotNil(t, handle)
		assert.Equal(t, &metrics.Descriptor{ID: "test-metric"}, handle.Descriptor)
		assert.NotNil(t, handle.Histogram)
		assert.NotNil(t, handle.Gauge)

		err = localHandler.DeleteMetric(context.Background(), &metrics.Descriptor{ID: "test-metric"})
		assert.NoError(t, err)
	}

}
