package metrics_test

import (
	"fmt"
	"testing"

	"github.com/networknext/backend/metrics"

	"github.com/go-kit/kit/metrics/generic"
	"github.com/stretchr/testify/assert"
)

func TestStackDriverMetrics(t *testing.T) {
	// Initialize the metric handler
	metricHandler, err := metrics.NewStackDriverMetricHandler()
	assert.NoError(t, err)

	metricName := "Test Metric"

	// Create a gauge to track a dummy metric
	gauge := generic.NewGauge(metricName)
	assert.Equal(t, metricName, gauge.Name)

	// Attempt to delete the metric before creating it, since it may still exist from
	// the last time the test was run
	metricHandler.DeleteMetric(&metrics.MetricDescriptor{
		PackageName: "package",
		MetricID:    "test-metric",
	})

	// Test metric creation
	var metric metrics.Metric
	metric, err = metricHandler.CreateMetric(&metrics.MetricDescriptor{
		PackageName: "package",
		MetricID:    "test-metric",
		ValueType:   metrics.MetricValueType{ValueType: &metrics.MetricTypeDouble{}},
		Unit:        "{units}",
		Description: "A dummy metric to test the new metrics package.",
	}, gauge)

	assert.NoError(t, err)

	// Attempt to create a metric again with the same ID and gauge
	var metric2 metrics.Metric
	metric2, err = metricHandler.CreateMetric(&metrics.MetricDescriptor{
		PackageName: "package",
		MetricID:    "test-metric",
		ValueType:   metrics.MetricValueType{ValueType: &metrics.MetricTypeInt64{}},
		Unit:        "{units}",
		Description: "A second dummy metric to test metric creation.",
	}, gauge)

	assert.Empty(t, metric2)
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

	// Send the metric to StackDriver
	metricHandler.SubmitMetric(metric)

	// Delete the test metric
	err = metricHandler.DeleteMetric(&metrics.MetricDescriptor{
		PackageName: "package",
		MetricID:    "test-metric",
	})

	assert.NoError(t, err)

	// Close the metric client
	err = metricHandler.Close()
	assert.NoError(t, err)
}
