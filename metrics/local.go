package metrics

import (
	"context"
	"errors"
	"expvar"
	"io"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/generic"
)

// LocalHandler handles metrics for local development by using gokit's metrics expvar package,
// creating a local endpoint to view all metrics as JSON in the browser.
type LocalHandler struct {
	metrics          map[string]Handle
	customMetricsMap *expvar.Map
}

// Open is a no-op.
func (local *LocalHandler) Open(ctx context.Context, credentials []byte) error { return nil }

// MetricSubmitRoutine is a no-op.
func (local *LocalHandler) MetricSubmitRoutine(ctx context.Context, logger log.Logger, duration time.Duration, maxMetricsIncrement int) {
}

// GetSubmitFrequency is a no-op.
func (local *LocalHandler) GetSubmitFrequency() float64 {
	return 1 / 60.0 // Simulate publishing every minute
}

// CreateMetric creates a local metric that is visible as JSON in the browser.
// It is automatically updated by the expvar package so calling MetricSubmitRoutine() is unnecessary.
// If the metric already exists, CreateMetric will return it. The error is not used.
func (local *LocalHandler) CreateMetric(ctx context.Context, descriptor *Descriptor) (Handle, error) {
	if local.metrics == nil {
		local.init()
	}

	if handle, contains := local.metrics[descriptor.ID]; contains {
		return handle, nil
	}

	value := new(expvar.Float)
	p50 := new(expvar.Float)
	p90 := new(expvar.Float)
	p95 := new(expvar.Float)
	p99 := new(expvar.Float)

	local.customMetricsMap.Set(descriptor.ID, value)
	local.customMetricsMap.Set(descriptor.ID+".p50", p50)
	local.customMetricsMap.Set(descriptor.ID+".p90", p90)
	local.customMetricsMap.Set(descriptor.ID+".p95", p95)
	local.customMetricsMap.Set(descriptor.ID+".p99", p99)

	handle := Handle{
		Descriptor: descriptor,
		Histogram: &LocalHistogram{
			h:       generic.NewHistogram(descriptor.ID, 50),
			buckets: 50,
			p50:     p50,
			p90:     p90,
			p95:     p95,
			p99:     p99,
		},
		Gauge: &LocalGauge{
			f: value,
		},
	}
	local.metrics[descriptor.ID] = handle

	return handle, nil
}

// GetMetric returns the metric handle by the given ID.
func (local *LocalHandler) GetMetric(id string) (Handle, bool) {
	if local.metrics == nil {
		local.init()
	}

	handle, contains := local.metrics[id]
	return handle, contains
}

// DeleteMetric removes the metric from the map of tracked metrics.
func (local *LocalHandler) DeleteMetric(ctx context.Context, descriptor *Descriptor) error {
	if local.metrics == nil {
		local.init()
	}

	if _, contains := local.metrics[descriptor.ID]; contains {
		delete(local.metrics, descriptor.ID)
		return nil
	}

	return errors.New("Attemped to delete unknown metric " + descriptor.ID)
}

// Close is a no-op.
func (local *LocalHandler) Close() error { return nil }

func (local *LocalHandler) init() {
	local.metrics = map[string]Handle{}
	result := expvar.Get("Local Metrics")
	if result != nil {
		local.customMetricsMap = result.(*expvar.Map)
	} else {
		local.customMetricsMap = expvar.NewMap("Local Metrics")
	}
}

// LocalGauge mimics go-kit's expvar.Gauge, but adds methods to satisfy this package's Gauge.
// Label values aren't supported.
type LocalGauge struct {
	f *expvar.Float
}

// With is a no-op.
func (g *LocalGauge) With(labelValues ...string) metrics.Gauge { return g }

// Set sets the gauge's value directly.
func (g *LocalGauge) Set(value float64) { g.f.Set(value) }

// Add adds the delta to the gauge's value.
func (g *LocalGauge) Add(delta float64) { g.f.Add(delta) }

// Value returns the gauge's current value.
func (g *LocalGauge) Value() float64 { return g.f.Value() }

// LabelValues is a no-op.
func (g *LocalGauge) LabelValues() []string { return nil }

// LocalHistogram mimics go-kit's expvar.Histogram, but adds methods to satisfy this package's Histogram.
// Label values aren't supported.
type LocalHistogram struct {
	mtx     sync.Mutex
	h       *generic.Histogram
	buckets int
	p50     *expvar.Float
	p90     *expvar.Float
	p95     *expvar.Float
	p99     *expvar.Float
}

// With is a no-op.
func (h *LocalHistogram) With(labelValues ...string) metrics.Histogram { return h }

// Observe adds the given value to the histogram.
func (h *LocalHistogram) Observe(value float64) {
	h.mtx.Lock()
	defer h.mtx.Unlock()
	h.h.Observe(value)
	h.p50.Set(h.Quantile(0.50))
	h.p90.Set(h.Quantile(0.90))
	h.p95.Set(h.Quantile(0.95))
	h.p99.Set(h.Quantile(0.99))
}

// Quantile returns the value of the quantile, between 0.0 and 1.0
func (h *LocalHistogram) Quantile(q float64) float64 {
	return h.h.Quantile(q)
}

// LabelValues is a no-op.
func (h *LocalHistogram) LabelValues() []string { return nil }

// Print writes a string representation of the histogram.
func (h *LocalHistogram) Print(w io.Writer) {
	h.h.Print(w)
}

// Reset resets all histogram data.
func (h *LocalHistogram) Reset() {
	h.mtx.Lock()
	defer h.mtx.Unlock()

	h.h = generic.NewHistogram(h.h.Name, h.buckets)
}
