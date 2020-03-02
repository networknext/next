package metrics_test

import (
	"context"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/networknext/backend/metrics"

	"github.com/go-kit/kit/log"
	"github.com/stretchr/testify/assert"
)

func TestStackDriverMetrics(t *testing.T) {
	ctx, cancelWriteLoop := context.WithCancel(context.Background())

	stackdrivercreds, ok := os.LookupEnv("GCP_CREDENTIALS_METRICS")
	if !ok {
		t.Skip() // Skip the test if GCP credentials aren't configured
	}

	projectID, ok := os.LookupEnv("GCP_METRICS_PROJECT")
	if !ok {
		t.Skip() // Skip the test if GCP metrics project ID isn't set
	}

	var stackdrivercredsjson []byte
	_, err := os.Stat(stackdrivercreds)
	assert.NoError(t, err)

	stackdrivercredsjson, err = ioutil.ReadFile(stackdrivercreds)
	assert.NoError(t, err)

	// Create the metrics handler
	handler := &metrics.StackDriverHandler{
		ProjectID:          projectID,
		Credentials:        stackdrivercredsjson,
		OverwriteFrequency: time.Second,
		OverwriteTimeout:   10 * time.Second,
	}

	// Open the StackDriver metrics client
	err = handler.Open(ctx)
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

	histogram, err := handler.NewHistogram(ctx, &metrics.Descriptor{
		DisplayName: "Test Metric Histogram",
		ServiceName: "service",
		ID:          "test-metric-histogram",
		Unit:        "{units}",
		Description: "A dummy metric to test the metrics package.",
	}, 50)

	assert.NoError(t, err)

	// Test duplicate metric
	_, err = handler.NewHistogram(ctx, &metrics.Descriptor{
		DisplayName: "Test Metric Histogram Duplicate",
		ServiceName: "service",
		ID:          "test-metric-histogram",
		Unit:        "{units}",
		Description: "A dummy metric to test the metrics package.",
	}, 50)

	assert.EqualError(t, err, "Metric test-metric-histogram already created")

	// Test counter functions
	labels := []string{"label1", "value1", "label2", "value2"}
	counter = counter.With(labels...).(metrics.Counter)
	labelsResult := counter.LabelValues()
	assert.Equal(t, labels, labelsResult)

	assert.Equal(t, 0.0, counter.Value())
	counter.Add(5)
	assert.Equal(t, 5.0, counter.Value())
	counter.Add(1.112)
	assert.Equal(t, 6.112, counter.Value())

	// Test gauge functions
	labels = []string{"label1", "value1", "label2", "value2"}
	gauge = gauge.With(labels...).(metrics.Gauge)
	labelsResult = gauge.LabelValues()
	assert.Equal(t, labels, labelsResult)

	assert.Equal(t, 0.0, gauge.Value())
	gauge.Add(5)
	assert.Equal(t, 5.0, gauge.Value())
	gauge.Add(1.112)
	assert.Equal(t, 6.112, gauge.Value())
	gauge.Set(4)
	assert.Equal(t, 4.0, gauge.Value())

	// Test histogram functions
	labels = []string{"label1", "value1", "label2", "value2"}
	histogram = histogram.With(labels...).(metrics.Histogram)
	labelsResult = histogram.LabelValues()
	assert.Equal(t, labels, labelsResult)

	assert.Equal(t, -1.0, histogram.Quantile(0.5))
	histogram.Observe(5)
	assert.Equal(t, 5.0, histogram.Quantile(0.5))
	histogram.Observe(1.112)
	assert.Equal(t, 1.112, histogram.Quantile(0.5))
	histogram.Observe(5)
	assert.Equal(t, 5.0, histogram.Quantile(0.5))

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
