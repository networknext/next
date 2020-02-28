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
		ProjectID: projectID,
	}

	// Open the StackDriver metrics client
	err = handler.Open(ctx, stackdrivercredsjson)
	assert.NoError(t, err)

	// Attempt to delete the metric before creating it, since it may still exist from
	// the last time the test was run
	handler.DeleteMetric(ctx, &metrics.Descriptor{
		DisplayName: "Test Metric",
		ServiceName: "service",
		ID:          "test-metric",
	})

	// Test handle creation
	var handle metrics.Handle
	handle, err = handler.CreateMetric(ctx, &metrics.Descriptor{
		DisplayName: "Test Metric",
		ServiceName: "service",
		ID:          "test-metric",
		ValueType:   metrics.ValueType{ValueType: metrics.TypeDouble{}},
		Unit:        "{units}",
		Description: "A dummy metric to test the new metrics package.",
	})

	assert.NotEmpty(t, handle)
	assert.NoError(t, err)

	// Wait a second for StackDriver to process the metric creation
	time.Sleep(2 * time.Second)

	// Attempt to create a metric again with the same ID
	// This should just retrive the same metric with the original values
	var handle2 metrics.Handle
	handle2, err = handler.CreateMetric(ctx, &metrics.Descriptor{
		DisplayName: "Test Metric",
		ServiceName: "service",
		ID:          "test-metric",
		ValueType:   metrics.ValueType{ValueType: metrics.TypeInt64{}},
		Unit:        "{units}",
		Description: "A second dummy metric to test metric creation.",
	})

	assert.Equal(t, handle, handle2)
	assert.NoError(t, err)

	// Test gauge functions
	gauge := handle.Gauge
	labels := []string{"label1", "value1", "label2", "value2"}
	gauge = gauge.With(labels...).(metrics.Gauge)
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
	go handler.WriteLoop(ctx, log.NewNopLogger(), time.Second, 200)

	// Sleep for 2 seconds to allow the metric to be pushed to StackDriver
	time.Sleep(2 * time.Second)

	// Delete the test metric
	err = handler.DeleteMetric(ctx, &metrics.Descriptor{
		ServiceName: "service",
		ID:          "test-metric",
	})
	assert.NoError(t, err)

	// Stop the submit routine
	cancelWriteLoop()

	// Close the metric client
	err = handler.Close()
	assert.NoError(t, err)
}
