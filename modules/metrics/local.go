package metrics

import (
	"context"
	"expvar"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
)

// LocalHandler handles metrics for local development by using gokit's metrics expvar package,
// creating a local endpoint to view all metrics as JSON in the browser.
type LocalHandler struct {
	counters map[string]counterMapData
	gauges   map[string]gaugeMapData

	counterMapMutex sync.Mutex
	gaugeMapMutex   sync.Mutex

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

	counter := &LocalCounter{f: value, labels: make(map[string]string)}
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

	gauge := &LocalGauge{f: value, labels: make(map[string]string)}
	local.gauges[descriptor.ID] = gaugeMapData{
		descriptor: descriptor,
		gauge:      gauge,
	}

	return gauge, nil
}

// Close is a no-op.
func (local *LocalHandler) Close() error { return nil }

func (local *LocalHandler) init() {
	local.counters = make(map[string]counterMapData)
	local.gauges = make(map[string]gaugeMapData)

	result := expvar.Get("Local Metrics")
	if result != nil {
		local.customMetricsMap = result.(*expvar.Map)
	} else {
		local.customMetricsMap = expvar.NewMap("Local Metrics")
	}
}

// LocalCounter is an implementation of a counter for running in the happy path.
type LocalCounter struct {
	f           *expvar.Float
	labels      map[string]string
	labelsMutex sync.RWMutex
}

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

// AddLabels adds more labels to the counter. If a label being added already exists, the value is overwritten.
func (c *LocalCounter) AddLabels(labels map[string]string) {
	c.labelsMutex.Lock()
	defer c.labelsMutex.Unlock()

	for k, v := range labels {
		c.labels[k] = v
	}
}

// Labels is a returns the current list of labels attached to the counter.
func (c *LocalCounter) Labels() map[string]string {
	labelsCopy := make(map[string]string)
	c.labelsMutex.RLock()
	defer c.labelsMutex.RUnlock()

	for k, v := range c.labels {
		labelsCopy[k] = v
	}

	return labelsCopy
}

// ClearLabels removes all existing labels attached to the counter
func (c *LocalCounter) ClearLabels() {
	c.labelsMutex.Lock()
	defer c.labelsMutex.Unlock()

	c.labels = make(map[string]string)
}

// LocalGauge mimics go-kit's expvar.Gauge, but adds methods to satisfy this package's Gauge.
// Label values aren't supported.
type LocalGauge struct {
	f           *expvar.Float
	labels      map[string]string
	labelsMutex sync.RWMutex
}

// Set sets the gauge's value directly.
func (g *LocalGauge) Set(value float64) { g.f.Set(value) }

// Add adds the delta to the gauge's value.
func (g *LocalGauge) Add(delta float64) { g.f.Add(delta) }

// Value returns the gauge's current value.
func (g *LocalGauge) Value() float64 { return g.f.Value() }

// ValueReset returns the gauge's current value and resets the gauge.
func (g *LocalGauge) ValueReset() float64 {
	v := g.f.Value()
	g.f.Set(0)
	return v
}

// AddLabels adds more labels to the gauge. If a label being added already exists, the value is overwritten.
func (g *LocalGauge) AddLabels(labels map[string]string) {
	g.labelsMutex.Lock()
	defer g.labelsMutex.Unlock()

	for k, v := range labels {
		g.labels[k] = v
	}
}

// Labels is a returns a copy of the current list of labels attached to the gauge.
func (g *LocalGauge) Labels() map[string]string {
	labelsCopy := make(map[string]string)
	g.labelsMutex.RLock()
	defer g.labelsMutex.RUnlock()

	for k, v := range g.labels {
		labelsCopy[k] = v
	}

	return labelsCopy
}

// ClearLabels removes all existing labels attached to the gauge
func (g *LocalGauge) ClearLabels() {
	g.labelsMutex.Lock()
	defer g.labelsMutex.Unlock()

	g.labels = make(map[string]string)
}
