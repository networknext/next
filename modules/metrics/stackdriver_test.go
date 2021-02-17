package metrics_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"

	"github.com/networknext/backend/modules/metrics"
)

func TestStackDriverMetrics(t *testing.T) {
	ctx, cancelWriteLoop := context.WithCancel(context.Background())

	projectID, ok := os.LookupEnv("GOOGLE_PROJECT_ID")
	if !ok {
		t.Skip() // Skip the test if GCP project ID isn't set
	}

	// Create the metrics handler
	handler := &metrics.StackDriverHandler{
		ProjectID:          projectID,
		OverwriteFrequency: time.Second,
		OverwriteTimeout:   10 * time.Second,
	}

	// Open the StackDriver metrics client
	err := handler.Open(ctx)
	assert.NoError(t, err)

	// Test metric creation
	counter, err := handler.NewCounter(ctx, &metrics.Descriptor{
		DisplayName: "Test Metric Counter",
		ServiceName: "service",
		ID:          "test-metric-counter",
		Unit:        "{units}",
		Description: "A dummy metric to test the metrics package.",
	})

	assert.NoError(t, err)

	gauge, err := handler.NewGauge(ctx, &metrics.Descriptor{
		DisplayName: "Test Metric Gauge",
		ServiceName: "service",
		ID:          "test-metric-gauge",
		Unit:        "{units}",
		Description: "A dummy metric to test the metrics package.",
	})

	assert.NoError(t, err)

	// Test counter functions
	labels := map[string]string{"label1": "value1", "label2": "value2"}
	counter.AddLabels(labels)
	labelsResult := counter.Labels()
	assert.Equal(t, labels, labelsResult)

	assert.Equal(t, 0.0, counter.Value())
	counter.Add(5)
	assert.Equal(t, 5.0, counter.Value())
	counter.Add(1.112)
	assert.Equal(t, 6.112, counter.Value())

	// Test gauge functions
	labels = map[string]string{"label1": "value1", "label2": "value2"}
	gauge.AddLabels(labels)
	labelsResult = gauge.Labels()
	assert.Equal(t, labels, labelsResult)

	assert.Equal(t, 0.0, gauge.Value())
	gauge.Add(5)
	assert.Equal(t, 5.0, gauge.Value())
	gauge.Add(1.112)
	assert.Equal(t, 6.112, gauge.Value())
	gauge.Set(4)
	assert.Equal(t, 4.0, gauge.Value())

	// Start the submit routine
	go handler.WriteLoop(ctx, log.NewNopLogger(), time.Second, 200)

	// Sleep for 2 seconds to allow the metric to be pushed to StackDriver
	time.Sleep(2 * time.Second)

	// Stop the submit routine
	cancelWriteLoop()

	// Close the metric client
	err = handler.Close()
	assert.NoError(t, err)
}
