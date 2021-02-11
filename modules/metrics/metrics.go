package metrics

import (
	"context"
	"time"

	"github.com/go-kit/kit/log"
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
	Close() error
}

// Valuer is an interface that accepts any metric type with a value.
type Valuer interface {
	Value() float64
}

// Counter is an interface that represents a metric counter, based on go-kit's generic counter.
type Counter interface {
	Add(delta float64)
	Value() float64
	ValueReset() float64
	AddLabels(labels map[string]string)
	Labels() map[string]string
	ClearLabels()
}

// Gauge is an interface that represents a metric gauge, based on go-kit's generic gauge.
type Gauge interface {
	Set(value float64)
	Add(delta float64)
	Value() float64
	ValueReset() float64
	AddLabels(labelValues map[string]string)
	Labels() map[string]string
	ClearLabels()
}
