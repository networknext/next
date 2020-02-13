package metrics_test

import (
	"context"
	"fmt"
	"os"
	"testing"

	monitoring "cloud.google.com/go/monitoring/apiv3"
	"github.com/networknext/backend/metrics"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-kit/kit/metrics/generic"
	"github.com/stretchr/testify/assert"
)

func TestStackDriverMetrics(t *testing.T) {
	// Configure logging
	logger := log.NewLogfmtLogger(os.Stdout)
	{
		switch os.Getenv("BACKEND_LOG_LEVEL") {
		case "none":
			logger = level.NewFilter(logger, level.AllowNone())
		case level.ErrorValue().String():
			logger = level.NewFilter(logger, level.AllowError())
		case level.WarnValue().String():
			logger = level.NewFilter(logger, level.AllowWarn())
		case level.InfoValue().String():
			logger = level.NewFilter(logger, level.AllowInfo())
		case level.DebugValue().String():
			logger = level.NewFilter(logger, level.AllowDebug())
		default:
			logger = level.NewFilter(logger, level.AllowWarn())
		}

		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	}

	// Initialize the metric handler
	handler := &metrics.StackDriverHandler{
		ClusterLocation: os.Getenv("GOOGLE_CLOUD_METRICS_CLUSTER_LOCATION"),
		ClusterName:     os.Getenv("GOOGLE_CLOUD_METRICS_CLUSTER_NAME"),
		PodName:         os.Getenv("GOOGLE_CLOUD_METRICS_POD_NAME"),
		ContainerName:   os.Getenv("GOOGLE_CLOUD_METRICS_CONTAINER_NAME"),
		NamespaceName:   os.Getenv("GOOGLE_CLOUD_METRICS_NAMESPACE_NAME"),
		ProjectID:       os.Getenv("GOOGLE_CLOUD_METRICS_PROJECT"),
		PushMetricsChan: make(chan []metrics.Handle),
	}

	// Create a Stackdriver metrics client
	var err error
	handler.Client, err = monitoring.NewMetricClient(context.Background())
	assert.NoError(t, err)

	metricName := "Test Metric"

	// Create a gauge to track a dummy metric
	gauge := generic.NewGauge(metricName)
	assert.Equal(t, metricName, gauge.Name)

	// Attempt to delete the metric before creating it, since it may still exist from
	// the last time the test was run
	handler.DeleteMetric(&metrics.Descriptor{
		PackageName: "package",
		ID:          "test-metric",
	})

	// Test handle creation
	var handle metrics.Handle
	handle, err = handler.CreateMetric(&metrics.Descriptor{
		PackageName: "package",
		ID:          "test-metric",
		ValueType:   metrics.ValueType{ValueType: &metrics.TypeDouble{}},
		Unit:        "{units}",
		Description: "A dummy metric to test the new metrics package.",
	}, gauge)

	assert.NotEmpty(t, handle)
	assert.NoError(t, err)

	// Attempt to create a metric again with the same ID and gauge
	var handle2 metrics.Handle
	handle2, err = handler.CreateMetric(&metrics.Descriptor{
		PackageName: "package",
		ID:          "test-metric",
		ValueType:   metrics.ValueType{ValueType: &metrics.TypeInt64{}},
		Unit:        "{units}",
		Description: "A second dummy metric to test metric creation.",
	}, gauge)

	assert.Empty(t, handle2)
	assert.EqualError(t, err, fmt.Sprintf("Attempted to create metric with name '%s' but a metric with that name already exists in StackDriver", metricName))

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

	// Call SumbitMetric before the submit routine has been started
	err = handler.SubmitMetric(handle)
	assert.EqualError(t, err, "Submit routine not running yet. Call MetricSubmitRoutine() in a goroutine before calling SubmitMetrics()")

	// Start the submit routine
	go handler.MetricSubmitRoutine(logger, 200)

	// Now send the metric to StackDriver
	err = handler.SubmitMetric(handle)
	assert.NoError(t, err)

	// Delete the test metric
	err = handler.DeleteMetric(&metrics.Descriptor{
		PackageName: "package",
		ID:          "test-metric",
	})

	assert.NoError(t, err)

	// Close the metric client
	err = handler.Close()
	assert.NoError(t, err)
}
