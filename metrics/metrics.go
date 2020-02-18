package metrics

import (
	"context"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics/generic"
)

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

func (mv TypeBool) getTypeName() string   { return "BOOL" }
func (mv TypeInt64) getTypeName() string  { return "INT64" }
func (mv TypeDouble) getTypeName() string { return "DOUBLE" }

// Descriptor describes metric metadata
type Descriptor struct {
	ServiceName string
	ID          string
	ValueType   ValueType
	Unit        string
	Description string
}

// Handler handles creating and update metrics
type Handler interface {
	Open(ctx context.Context) error
	MetricSubmitRoutine(ctx context.Context, logger log.Logger, c <-chan time.Time, maxMetricsIncrement int)
	CreateMetric(ctx context.Context, descriptor *Descriptor, gauge *generic.Gauge) (Handle, error)
	GetMetric(id string) (Handle, bool)
	DeleteMetric(ctx context.Context, descriptor *Descriptor) error
	Close() error
}

// Handle is the return result of creating or fetching a metric. It allows access to the metric's descriptor and gauge.
type Handle struct {
	Descriptor *Descriptor
	Gauge      *generic.Gauge
}
