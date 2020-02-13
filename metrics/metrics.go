package metrics

import "github.com/go-kit/kit/metrics/generic"

// MetricValueType represents the type of the metric's value
type MetricValueType struct {
	ValueType IMetricValueType
}

// IMetricValueType is an interface to support multiple types
type IMetricValueType interface {
	getTypeName() string
}

// MetricTypeBool represents a boolean type value
type MetricTypeBool struct {
	Value bool
}

// MetricTypeInt64 represents an 8 byte integer type value
type MetricTypeInt64 struct {
	Value int64
}

// MetricTypeDouble represents a double precision floating point type value
type MetricTypeDouble struct {
	Value float64
}

func (mv *MetricTypeBool) getTypeName() string   { return "bool" }
func (mv *MetricTypeInt64) getTypeName() string  { return "int64" }
func (mv *MetricTypeDouble) getTypeName() string { return "double" }

// MetricDescriptor describes metric metadata
type MetricDescriptor struct {
	PackageName string
	MetricID    string
	ValueType   MetricValueType
	Unit        string
	Description string
}

// Metric is a handle that is passed to the MetricHandler to publish the metric
type Metric struct {
	MetricDescriptor *MetricDescriptor
	Gauge            *generic.Gauge
}

// MetricHandler handles creating and update metrics
type MetricHandler interface {
	CreateMetric(*MetricDescriptor, *generic.Gauge) Metric
	SubmitMetric(Metric)
	SubmitMetrics([]Metric)
	DeleteMetric(*MetricDescriptor) error
	Close() error
}
