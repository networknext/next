package metrics

import (
	"context"
	"io"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
)

// NoOpHandler is a metric handler that doesn't do anything. Useful for testing and error handling.
type NoOpHandler struct{}

func (handler *NoOpHandler) Open(ctx context.Context, credentials []byte) error { return nil }
func (handler *NoOpHandler) WriteLoop(ctx context.Context, logger log.Logger, duration time.Duration, maxMetricsIncrement int) {
}
func (handler *NoOpHandler) GetWriteFrequency() float64 { return 0 }
func (handler *NoOpHandler) CreateMetric(ctx context.Context, descriptor *Descriptor) (Handle, error) {
	return EmptyHandle, nil
}
func (handler *NoOpHandler) GetMetric(id string) (Handle, bool) { return EmptyHandle, true }
func (handler *NoOpHandler) DeleteMetric(ctx context.Context, descriptor *Descriptor) error {
	return nil
}
func (handler *NoOpHandler) Close() error { return nil }

// EmptyGauge is a gauge with no data. Useful for testing and error handling.
type EmptyGauge struct{}

func (g *EmptyGauge) With(labelValues ...string) metrics.Gauge { return g }
func (g *EmptyGauge) Set(value float64)                        {}
func (g *EmptyGauge) Add(delta float64)                        {}
func (g *EmptyGauge) Value() float64                           { return 0.0 }
func (g *EmptyGauge) LabelValues() []string                    { return nil }

// EmptyHistogram is a histogram with no data. Useful for testing and error handling.
type EmptyHistogram struct{}

func (h *EmptyHistogram) With(labelValues ...string) metrics.Histogram { return h }
func (h *EmptyHistogram) Observe(value float64)                        {}
func (h *EmptyHistogram) Quantile(q float64) float64                   { return 0.0 }
func (h *EmptyHistogram) LabelValues() []string                        { return nil }
func (h *EmptyHistogram) Print(w io.Writer)                            {}
func (h *EmptyHistogram) Reset()                                       {}
