package metrics

import (
	"context"
	"expvar"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/metrics"
	"github.com/go-kit/kit/metrics/generic"
)

// LocalHandler handles metrics for local development by using gokit's metrics expvar package,
// creating a local endpoint to view all metrics as JSON in the browser.
type LocalHandler struct {
	counters   map[string]counterMapData
	gauges     map[string]gaugeMapData
	histograms map[string]histogramMapData

	counterMapMutex   sync.Mutex
	gaugeMapMutex     sync.Mutex
	histogramMapMutex sync.Mutex

	customMetricsMap *expvar.Map
}

// Open is a no-op.
func (local *LocalHandler) Open(ctx context.Context) error { return nil }

// WriteLoop is a no-op, since writing is handled by expvar.
func (local *LocalHandler) WriteLoop(ctx context.Context, logger log.Logger, duration time.Duration, maxMetricsIncrement int) {
}

// NewCounter creates a local metric that is visible as JSON in the browser update by the returned counter.
// It is automatically updated by the expvar package so calling WriteLoop() is unnecessary.
// If the metric already exists, NewCounter will return the counter to update it. The error is not used.
func (local *LocalHandler) NewCounter(ctx context.Context, descriptor *Descriptor) (Counter, error) {
	if local.counters == nil || local.customMetricsMap == nil {
		local.init()
	}

	local.counterMapMutex.Lock()
	defer local.counterMapMutex.Unlock()

	if mapData, contains := local.counters[descriptor.ID]; contains {
		return mapData.counter, nil
	}

	value := new(expvar.Float)
	local.customMetricsMap.Set(descriptor.ID, value)

	counter := &LocalCounter{f: value}
	local.counters[descriptor.ID] = counterMapData{
		descriptor: descriptor,
		counter:    counter,
	}

	return counter, nil
}

// NewGauge creates a local metric that is visible as JSON in the browser update by the returned gauge.
// It is automatically updated by the expvar package so calling WriteLoop() is unnecessary.
// If the metric already exists, NewGauge will return the gauge to update it. The error is not used.
func (local *LocalHandler) NewGauge(ctx context.Context, descriptor *Descriptor) (Gauge, error) {
	if local.gauges == nil || local.customMetricsMap == nil {
		local.init()
	}

	local.gaugeMapMutex.Lock()
	defer local.gaugeMapMutex.Unlock()

	if mapData, contains := local.gauges[descriptor.ID]; contains {
		return mapData.gauge, nil
	}

	value := new(expvar.Float)
	local.customMetricsMap.Set(descriptor.ID, value)

	gauge := &LocalGauge{f: value}
	local.gauges[descriptor.ID] = gaugeMapData{
		descriptor: descriptor,
		gauge:      gauge,
	}

	return gauge, nil
}

// NewHistogram creates a local metric that is visible as JSON in the browser observed by the returned histogram.
// It is automatically updated by the expvar package so calling WriteLoop() is unnecessary.
// If the metric already exists, NewHistogram will return the histogram observing it. The error is not used.
func (local *LocalHandler) NewHistogram(ctx context.Context, descriptor *Descriptor, buckets int) (Histogram, error) {
	if local.histograms == nil || local.customMetricsMap == nil {
		local.init()
	}

	local.histogramMapMutex.Lock()
	defer local.histogramMapMutex.Unlock()

	if mapData, contains := local.histograms[descriptor.ID]; contains {
		return mapData.histogram, nil
	}

	p50 := new(expvar.Float)
	p90 := new(expvar.Float)
	p95 := new(expvar.Float)
	p99 := new(expvar.Float)

	local.customMetricsMap.Set(descriptor.ID+".p50", p50)
	local.customMetricsMap.Set(descriptor.ID+".p90", p90)
	local.customMetricsMap.Set(descriptor.ID+".p95", p95)
	local.customMetricsMap.Set(descriptor.ID+".p99", p99)

	histogram := &LocalHistogram{
		h:       generic.NewHistogram(descriptor.ID, buckets),
		buckets: buckets,
		p50:     p50,
		p90:     p90,
		p95:     p95,
		p99:     p99,
	}

	local.histograms[descriptor.ID] = histogramMapData{
		descriptor: descriptor,
		histogram:  histogram,
		buckets:    buckets,
	}

	return histogram, nil
}

// Close is a no-op.
func (local *LocalHandler) Close() error { return nil }

func (local *LocalHandler) init() {
	local.counters = make(map[string]counterMapData)
	local.gauges = make(map[string]gaugeMapData)
	local.histograms = make(map[string]histogramMapData)

	result := expvar.Get("Local Metrics")
	if result != nil {
		local.customMetricsMap = result.(*expvar.Map)
	} else {
		local.customMetricsMap = expvar.NewMap("Local Metrics")
	}
}

// LocalCounter mimics go-kit's expvar.Counter, but adds methods to satisfy this package's Counter.
// Label values aren't supported.
type LocalCounter struct {
	f *expvar.Float
}

// With is a no-op.
func (c *LocalCounter) With(labelValues ...string) metrics.Counter { return c }

// Add adds the delta to the counter's value.
func (c *LocalCounter) Add(delta float64) { c.f.Add(delta) }

// Value returns the counter's current value.
func (c *LocalCounter) Value() float64 { return c.f.Value() }

// ValueReset returns the counter's current value and resets the counter.
func (c *LocalCounter) ValueReset() float64 {
	v := c.f.Value()
	c.f.Set(0)
	return v
}

// LabelValues is a no-op.
func (c *LocalCounter) LabelValues() []string { return nil }

// Reset sets the counter's value to 0.
func (c *LocalCounter) Reset() { c.f.Set(0) }

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

// Reset sets the gauge's value to 0.
func (g *LocalGauge) Reset() { g.f.Set(0) }

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
	h.h.Observe(value)

	h.mtx.Lock()
	defer h.mtx.Unlock()
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
