package metrics

import (
	"context"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
)

// Descriptor describes metric metadata
type Descriptor struct {
	DisplayName string
	ServiceName string
	ID          string
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

type Valuer interface {
	Value() float64
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
