package metrics

import "github.com/go-kit/kit/metrics/generic"

// ValueType represents the type of the metric's value
type ValueType struct {
	ValueType Type
}

// Type is an interface to support multiple types for the metric
type Type interface {
	getTypeName() string
}

// TypeBool represents a boolean type value
type TypeBool struct {
	Value bool
}

// TypeInt64 represents an 8 byte integer type value
type TypeInt64 struct {
	Value int64
}

// TypeDouble represents a double precision floating point type value
type TypeDouble struct {
	Value float64
}

func (mv *TypeBool) getTypeName() string   { return "bool" }
func (mv *TypeInt64) getTypeName() string  { return "int64" }
func (mv *TypeDouble) getTypeName() string { return "double" }

// Descriptor describes metric metadata
type Descriptor struct {
	PackageName string
	ID          string
	ValueType   ValueType
	Unit        string
	Description string
}

// Handle is a handle that is passed to the Handler to publish the metric
type Handle struct {
	MetricDescriptor *Descriptor
	Gauge            *generic.Gauge
}

// Handler handles creating and update metrics
type Handler interface {
	MetricSubmitRoutine(maxMetricsCount int)
	CreateMetric(*Descriptor, *generic.Gauge) Handle
	SubmitMetric(Handle)
	SubmitMetrics([]Handle)
	DeleteMetric(*Descriptor) error
	Close() error
}
