package metrics

import (
	"context"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
)

// NoOpHandler is a metric handler that doesn't do anything. Useful for testing and error handling.
type NoOpHandler struct{}

func (handler *NoOpHandler) Open(ctx context.Context) error { return nil }
func (handler *NoOpHandler) WriteLoop(ctx context.Context, logger log.Logger, duration time.Duration, maxMetricsIncrement int) {
}
func (handler *NoOpHandler) NewCounter(ctx context.Context, descriptor *Descriptor) (Counter, error) {
	return &EmptyCounter{}, nil
}
func (handler *NoOpHandler) NewGauge(ctx context.Context, descriptor *Descriptor) (Gauge, error) {
	return &EmptyGauge{}, nil
}
func (handler *NoOpHandler) NewHistogram(ctx context.Context, descriptor *Descriptor, buckets int) (Histogram, error) {
	return &EmptyHistogram{}, nil
}
func (handler *NoOpHandler) Close() error { return nil }

// EmptyCounter is a counter that does nothing. Useful for testing and error handling.
type EmptyCounter struct{}

func (c *EmptyCounter) With(labelValues ...string) metrics.Counter { return c }
func (c *EmptyCounter) Add(delta float64)                          {}
func (c *EmptyCounter) Value() float64                             { return 0.0 }
func (c *EmptyCounter) LabelValues() []string                      { return nil }
func (c *EmptyCounter) ValueReset() float64                        { return 0.0 }

// EmptyGauge is a gauge that does nothing. Useful for testing and error handling.
type EmptyGauge struct{}

func (g *EmptyGauge) With(labelValues ...string) metrics.Gauge { return g }
func (g *EmptyGauge) Set(value float64)                        {}
func (g *EmptyGauge) Add(delta float64)                        {}
func (g *EmptyGauge) Value() float64                           { return 0.0 }
func (g *EmptyGauge) LabelValues() []string                    { return nil }

// EmptyHistogram is a histogram that does nothing. Useful for testing and error handling.
type EmptyHistogram struct{}

func (h *EmptyHistogram) With(labelValues ...string) metrics.Histogram { return h }
func (h *EmptyHistogram) Observe(value float64)                        {}
func (h *EmptyHistogram) Quantile(q float64) float64                   { return 0.0 }
func (h *EmptyHistogram) LabelValues() []string                        { return nil }
