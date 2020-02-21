package metrics_test

import (
	"context"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/networknext/backend/metrics"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics/generic"
	"github.com/stretchr/testify/assert"
)

func TestStackDriverMetrics(t *testing.T) {
	ctx, cancelMetricSubmitRoutine := context.WithCancel(context.Background())

	gcpcreds, ok := os.LookupEnv("GCP_CREDENTIALS")
	if !ok {
		t.Skip() // Skip the test if GCP credentials aren't configured
	}

	projectID, ok := os.LookupEnv("GCP_METRICS_PROJECT")
	if !ok {
		t.Skip() // Skip the test if GCP metrics project ID isn't set
	}

	var gcpcredsjson []byte
	_, err := os.Stat(gcpcreds)
	assert.NoError(t, err)

	gcpcredsjson, err = ioutil.ReadFile(gcpcreds)
	assert.NoError(t, err)

	// Create the metrics handler
	handler := &metrics.StackDriverHandler{
		ProjectID: projectID,
	}

	// Open the StackDriver metrics client
	err = handler.Open(ctx, gcpcredsjson)
	assert.NoError(t, err)

	// Create a gauge to track a dummy metric
	metricName := "Test Metric"
	gauge := generic.NewGauge(metricName)
	assert.Equal(t, metricName, gauge.Name)

	// Attempt to delete the metric before creating it, since it may still exist from
	// the last time the test was run
	handler.DeleteMetric(ctx, &metrics.Descriptor{
		ServiceName: "service",
		ID:          "test-metric",
	})

	// Test handle creation
	var handle metrics.Handle
	handle, err = handler.CreateMetric(ctx, &metrics.Descriptor{
		ServiceName: "service",
		ID:          "test-metric",
		ValueType:   metrics.ValueType{ValueType: metrics.TypeDouble{}},
		Unit:        "{units}",
		Description: "A dummy metric to test the new metrics package.",
	}, gauge)

	assert.NotEmpty(t, handle)
	assert.NoError(t, err)

	// Wait a second for StackDriver to process the metric creation
	time.Sleep(2 * time.Second)

	// Attempt to create a metric again with the same ID and gauge
	// This should just retrive the same metric with the original values
	var handle2 metrics.Handle
	handle2, err = handler.CreateMetric(ctx, &metrics.Descriptor{
		ServiceName: "service",
		ID:          "test-metric",
		ValueType:   metrics.ValueType{ValueType: metrics.TypeInt64{}},
		Unit:        "{units}",
		Description: "A second dummy metric to test metric creation.",
	}, gauge)

	assert.Equal(t, handle, handle2)
	assert.NoError(t, err)

	// Test gauge functions
	labels := []string{"label1", "value1", "label2", "value2"}
	gauge = gauge.With(labels...).(*generic.Gauge)
	labelsResult := gauge.LabelValues()
	assert.Equal(t, labels, labelsResult)

	assert.Equal(t, 0.0, gauge.Value())
	gauge.Add(5)
	assert.Equal(t, 5.0, gauge.Value())
	gauge.Add(1.112)
	assert.Equal(t, 6.112, gauge.Value())
	gauge.Set(4)
	assert.Equal(t, 4.0, gauge.Value())

	// Start the submit routine
	go handler.MetricSubmitRoutine(ctx, log.NewNopLogger(), time.Second, 200)

	// Sleep for 2 seconds to allow the metric to be pushed to StackDriver
	time.Sleep(2 * time.Second)

	// Delete the test metric
	err = handler.DeleteMetric(ctx, &metrics.Descriptor{
		ServiceName: "service",
		ID:          "test-metric",
	})
	assert.NoError(t, err)

	// Stop the submit routine
	cancelMetricSubmitRoutine()

	// Close the metric client
	err = handler.Close()
	assert.NoError(t, err)
}
