package metrics

import (
	"context"
	"time"

	"github.com/go-kit/kit/log"
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
func (handler *NoOpHandler) Close() error { return nil }

// EmptyCounter is a counter that does nothing. Useful for testing and error handling.
type EmptyCounter struct{}

func (c *EmptyCounter) Add(delta float64)                  {}
func (c *EmptyCounter) Value() float64                     { return 0.0 }
func (c *EmptyCounter) ValueReset() float64                { return 0.0 }
func (c *EmptyCounter) AddLabels(labels map[string]string) {}
func (c *EmptyCounter) Labels() map[string]string          { return nil }
func (c *EmptyCounter) ClearLabels()                       {}

// EmptyGauge is a gauge that does nothing. Useful for testing and error handling.
type EmptyGauge struct{}

func (g *EmptyGauge) Set(value float64)                  {}
func (g *EmptyGauge) Add(delta float64)                  {}
func (g *EmptyGauge) Value() float64                     { return 0.0 }
func (g *EmptyGauge) ValueReset() float64                { return 0.0 }
func (g *EmptyGauge) AddLabels(labels map[string]string) {}
func (g *EmptyGauge) Labels() map[string]string          { return nil }
func (g *EmptyGauge) ClearLabels()                       {}
