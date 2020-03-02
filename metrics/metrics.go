package metrics

import (
	"context"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
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
	DisplayName string
	ServiceName string
	ID          string
	ValueType   ValueType
	Unit        string
	Description string
}

// Handler handles creating and update metrics
type Handler interface {
	Open(ctx context.Context) error
	WriteLoop(ctx context.Context, logger log.Logger, duration time.Duration, maxMetricsIncrement int)
	NewCounter(ctx context.Context, descriptor *Descriptor) (Counter, error)
	NewGauge(ctx context.Context, descriptor *Descriptor) (Gauge, error)
	NewHistogram(ctx context.Context, descriptor *Descriptor, buckets int) (Histogram, error)
	Close() error
}

// Counter is an interface that represents a metric counter, based on go-kit's generic counter.
type Counter interface {
	With(labelValues ...string) metrics.Counter
	Add(delta float64)
	Value() float64
	ValueReset() float64
	LabelValues() []string
}

// Gauge is an interface that represents a metric gauge, based on go-kit's generic gauge.
type Gauge interface {
	With(labelValues ...string) metrics.Gauge
	Set(value float64)
	Add(delta float64)
	Value() float64
	LabelValues() []string
}

// Histogram is an interface that represents a metric histogram, based on go-kit's generic histogram.
type Histogram interface {
	With(labelValues ...string) metrics.Histogram
	Observe(value float64)
	Quantile(q float64) float64
	LabelValues() []string
}
