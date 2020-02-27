package metrics

import (
	"context"
	"time"

	"github.com/go-kit/kit/log"
)

type NoOpHandler struct{}

func (handler *NoOpHandler) Open(ctx context.Context, credentials []byte) error { return nil }
func (handler *NoOpHandler) MetricSubmitRoutine(ctx context.Context, logger log.Logger, duration time.Duration, maxMetricsIncrement int) {
}
func (handler *NoOpHandler) GetSubmitFrequency() float64 { return 0 }
func (handler *NoOpHandler) CreateMetric(ctx context.Context, descriptor *Descriptor) (Handle, error) {
	return Handle{}, nil
}
func (handler *NoOpHandler) GetMetric(id string) (Handle, bool) { return Handle{}, true }
func (handler *NoOpHandler) DeleteMetric(ctx context.Context, descriptor *Descriptor) error {
	return nil
}
func (handler *NoOpHandler) Close() error { return nil }
