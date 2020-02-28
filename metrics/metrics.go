package metrics

import (
	"context"
	"io"
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
	Open(ctx context.Context, credentials []byte) error
	WriteLoop(ctx context.Context, logger log.Logger, duration time.Duration, maxMetricsIncrement int)
	GetWriteFrequency() float64
	CreateMetric(ctx context.Context, descriptor *Descriptor) (Handle, error)
	GetMetric(id string) (Handle, bool)
	DeleteMetric(ctx context.Context, descriptor *Descriptor) error
	Close() error
}

// Handle is the return result of creating or fetching a metric. It allows access to the metric's descriptor and gauge.
type Handle struct {
	Descriptor *Descriptor
	Histogram  Histogram
	Gauge      Gauge
}

// EmptyHandle is a metric handle with no data. Useful for testing and error handling.
var EmptyHandle = Handle{
	Descriptor: &Descriptor{},
	Histogram:  &EmptyHistogram{},
	Gauge:      &EmptyGauge{},
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
	Print(w io.Writer)
	Reset()
}
