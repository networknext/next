package metrics

import (
	"context"
	"errors"
	"fmt"
	"math"
	"sync"
	"sync/atomic"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	metadataapi "cloud.google.com/go/compute/metadata"
	monitoring "cloud.google.com/go/monitoring/apiv3"
	googlepb "github.com/golang/protobuf/ptypes/timestamp"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
	monitoredrespb "google.golang.org/genproto/googleapis/api/monitoredres"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// StackDriverHandler is an implementation of the Handler interface that handles metrics for StackDriver.
type StackDriverHandler struct {
	ProjectID string

	// When creating metrics, if these overwrite values are greater than zero, then the created metric will overwrite any existing metric with the same service/ID combination.
	// If these values are both <= 0, then the metric won't be overwritten and the descriptor will be updated to match the version in StackDriver.

	OverwriteFrequency time.Duration // The frequency at which to attempt to overwrite metrics.
	OverwriteTimeout   time.Duration // The max amount of time to spend attempting to overwrite a metric before returning an error.

	Client *monitoring.MetricClient

	counters map[string]counterMapData
	gauges   map[string]gaugeMapData

	counterMapMutex sync.Mutex
	gaugeMapMutex   sync.Mutex
}

type counterMapData struct {
	descriptor *Descriptor
	counter    Counter
}

type gaugeMapData struct {
	descriptor *Descriptor
	gauge      Gauge
}

// Open opens the client connection to StackDriver. This must be done before any metrics are created, deleted, or fetched.
func (handler *StackDriverHandler) Open(ctx context.Context) error {
	handler.counters = make(map[string]counterMapData)
	handler.gauges = make(map[string]gaugeMapData)

	// Create a Stackdriver metrics client
	var err error
	handler.Client, err = monitoring.NewMetricClient(ctx)
	return err
}

// WriteLoop is responsible for sending the metrics up to StackDriver. Call in a separate goroutine.
// Pass a duration in seconds to have the routine send metrics up to StackDriver periodically.
// If the duration is less than or equal to 0, a default of 1 minute is used.
// maxMetricsIncrement is the maximum number of metrics to send in one push to StackDriver. 200 is the maximum number of time series allowed in a single request.
func (handler *StackDriverHandler) WriteLoop(ctx context.Context, logger log.Logger, duration time.Duration, maxMetricsIncrement int) {
	monitoredResource, err := handler.getMonitoredResource()
	if err != nil {
		level.Error(logger).Log("msg", "error when detecting monitored resource, defaulting to global", "err", err)
	}

	if duration <= 0 {
		duration = time.Minute
	}

	ticker := time.NewTicker(duration)

	for {
		select {
		case <-ticker.C:
			// Preprocess all metrics in the map to create time series objects
			index := 0

			handler.counterMapMutex.Lock()
			handler.gaugeMapMutex.Lock()

			metricsCount := len(handler.counters) + len(handler.gauges)
			timeSeries := make([]*monitoringpb.TimeSeries, metricsCount)

			for _, mapData := range handler.counters {
				timeSeries[index] = &monitoringpb.TimeSeries{
					Metric: &metricpb.Metric{
						Type:   fmt.Sprintf("custom.googleapis.com/%s/%s", mapData.descriptor.ServiceName, mapData.descriptor.ID),
						Labels: mapData.counter.Labels(),
					},
					Resource: monitoredResource,
					Points: []*monitoringpb.Point{
						{
							Interval: &monitoringpb.TimeInterval{
								EndTime: &googlepb.Timestamp{
									Seconds: time.Now().Unix(),
								},
							},
							Value: &monitoringpb.TypedValue{
								Value: &monitoringpb.TypedValue_Int64Value{
									Int64Value: int64(math.Round(mapData.counter.ValueReset())),
								},
							},
						},
					},
				}

				mapData.counter.ClearLabels()
				index++
			}

			for _, mapData := range handler.gauges {
				timeSeries[index] = &monitoringpb.TimeSeries{
					Metric: &metricpb.Metric{
						Type:   fmt.Sprintf("custom.googleapis.com/%s/%s", mapData.descriptor.ServiceName, mapData.descriptor.ID),
						Labels: mapData.gauge.Labels(),
					},
					Resource: monitoredResource,
					Points: []*monitoringpb.Point{
						{
							Interval: &monitoringpb.TimeInterval{
								EndTime: &googlepb.Timestamp{
									Seconds: time.Now().Unix(),
								},
							},
							Value: &monitoringpb.TypedValue{
								Value: &monitoringpb.TypedValue_DoubleValue{
									DoubleValue: mapData.gauge.ValueReset(),
								},
							},
						},
					},
				}

				mapData.gauge.ClearLabels()
				index++
			}

			handler.counterMapMutex.Unlock()
			handler.gaugeMapMutex.Unlock()

			// Send the time series objects to StackDriver with a maximum send size to avoid overloading
			for i := 0; i < metricsCount; i += maxMetricsIncrement {
				// Calculate the number of metrics to process this iteration
				e := i + maxMetricsIncrement
				if e > metricsCount {
					e = metricsCount
				}

				if err := handler.Client.CreateTimeSeries(ctx, &monitoringpb.CreateTimeSeriesRequest{
					Name:       monitoring.MetricProjectPath(handler.ProjectID),
					TimeSeries: timeSeries[i:e],
				}); err != nil {
					level.Error(logger).Log("msg", "Failed to write time series data", "err", err)
				} else {
					level.Debug(logger).Log("msg", "Metrics pushed to StackDriver")
				}
			}
		case <-ctx.Done():
			return
		}
	}
}

// NewCounter creates a metric and returns a new counter to update it.
// If the metric already exists, then the returned new counter updates the metric instead of any other existing counter.
func (handler *StackDriverHandler) NewCounter(ctx context.Context, descriptor *Descriptor) (Counter, error) {
	// Create the metric in StackDriver
	if err := handler.createMetric(ctx, descriptor, metricpb.MetricDescriptor_INT64, metricpb.MetricDescriptor_GAUGE); err != nil {
		return nil, err
	}

	// Create the counter for updating the metric
	counter := StackDriverCounter{
		labels: make(map[string]string),
	}

	// Add the metric to the map data
	handler.counterMapMutex.Lock()
	handler.counters[descriptor.ID] = counterMapData{
		descriptor: descriptor,
		counter:    &counter,
	}
	handler.counterMapMutex.Unlock()

	return &counter, nil
}

// NewGauge creates a metric and returns a new gauge to update it.
// If the metric already exists, then the returned new gauge updates the metric instead of any other existing gauge.
func (handler *StackDriverHandler) NewGauge(ctx context.Context, descriptor *Descriptor) (Gauge, error) {
	// Create the metric in StackDriver
	if err := handler.createMetric(ctx, descriptor, metricpb.MetricDescriptor_DOUBLE, metricpb.MetricDescriptor_GAUGE); err != nil {
		return nil, err
	}

	// Create the gauge for updating the metric
	gauge := StackDriverGauge{
		labels: make(map[string]string),
	}

	// Add the metric to the map data
	handler.gaugeMapMutex.Lock()
	handler.gauges[descriptor.ID] = gaugeMapData{
		descriptor: descriptor,
		gauge:      &gauge,
	}
	handler.gaugeMapMutex.Unlock()

	return &gauge, nil
}

// createMetric creates the metric on StackDriver using the given metric descriptor.
// If the metric already exists on StackDriver, it will overwrite it.
func (handler *StackDriverHandler) createMetric(ctx context.Context, descriptor *Descriptor, valueType metricpb.MetricDescriptor_ValueType, kind metricpb.MetricDescriptor_MetricKind) error {
	// Check if the metric ID already exists in the local set of metrics
	if _, contains := handler.counters[descriptor.ID]; contains {
		return errors.New("Metric " + descriptor.ID + " already created")
	}

	if _, contains := handler.gauges[descriptor.ID]; contains {
		return errors.New("Metric " + descriptor.ID + " already created")
	}

	// Create the metric in StackDriver
	_, err := handler.Client.CreateMetricDescriptor(ctx, &monitoringpb.CreateMetricDescriptorRequest{
		Name: monitoring.MetricProjectPath(handler.ProjectID),
		MetricDescriptor: &metricpb.MetricDescriptor{
			Name:        descriptor.ID,
			Type:        fmt.Sprintf("custom.googleapis.com/%s/%s", descriptor.ServiceName, descriptor.ID),
			MetricKind:  kind,
			ValueType:   valueType,
			Unit:        descriptor.Unit,
			Description: descriptor.Description,
			DisplayName: descriptor.DisplayName,
		},
	})

	if err != nil {
		if status.Code(err) != codes.AlreadyExists {
			return err
		}

		// The metric already exists in StackDriver (from a previous run), check if it needs to be overwritten
		var stackdriverDescriptor *metricpb.MetricDescriptor
		if stackdriverDescriptor, err = handler.Client.GetMetricDescriptor(ctx, &monitoringpb.GetMetricDescriptorRequest{
			Name: fmt.Sprintf("projects/%s/metricDescriptors/custom.googleapis.com/%s/%s", handler.ProjectID, descriptor.ServiceName, descriptor.ID),
		}); err != nil {
			return err
		}

		if kind != stackdriverDescriptor.MetricKind ||
			valueType != stackdriverDescriptor.ValueType ||
			descriptor.Unit != stackdriverDescriptor.Unit ||
			descriptor.Description != stackdriverDescriptor.Description ||
			descriptor.DisplayName != stackdriverDescriptor.DisplayName {
			// The given descriptor differs from StackDriver, so overwrite the metric
			if err := handler.overwriteMetric(ctx, descriptor, valueType, kind); err != nil {
				if err.Error() == "Overwrite not set" {
					// User doesn't want to overwrite metrics, so update the descriptor with the version from StackDriver
					descriptor.Unit = stackdriverDescriptor.Unit
					descriptor.Description = stackdriverDescriptor.Description
					descriptor.DisplayName = stackdriverDescriptor.DisplayName

					return nil
				}

				return err
			}
		}
	}

	return nil
}

func (handler *StackDriverHandler) overwriteMetric(ctx context.Context, descriptor *Descriptor, valueType metricpb.MetricDescriptor_ValueType, kind metricpb.MetricDescriptor_MetricKind) error {
	if handler.OverwriteFrequency <= 0 || handler.OverwriteTimeout <= 0 {
		return errors.New("Overwrite not set") // Overwriting values not set, user doesn't want to overwrite metrics.
	}

	// Start by deleting the existing metric
	if err := handler.Client.DeleteMetricDescriptor(ctx, &monitoringpb.DeleteMetricDescriptorRequest{
		Name: fmt.Sprintf("projects/%s/metricDescriptors/custom.googleapis.com/%s/%s", handler.ProjectID, descriptor.ServiceName, descriptor.ID),
	}); err != nil {
		return err
	}

	// Periodically check if the metric has been deleted, and when it has, create it again
	ticker := time.NewTicker(handler.OverwriteFrequency)
	timer := time.NewTimer(handler.OverwriteTimeout)

	loop := true
	for loop {
		select {
		case <-ticker.C:
			if _, err := handler.Client.GetMetricDescriptor(ctx, &monitoringpb.GetMetricDescriptorRequest{
				Name: fmt.Sprintf("projects/%s/metricDescriptors/custom.googleapis.com/%s/%s", handler.ProjectID, descriptor.ServiceName, descriptor.ID),
			}); err != nil {
				if status.Code(err) == codes.NotFound {
					// Metric has been deleted
					ticker.Stop()
					timer.Stop()
					loop = false
				}
			}

		case <-timer.C:
			return errors.New("Failed to create metric: overwrite timeout exceeded")
		}
	}

	// Recreate the metric
	_, err := handler.Client.CreateMetricDescriptor(ctx, &monitoringpb.CreateMetricDescriptorRequest{
		Name: monitoring.MetricProjectPath(handler.ProjectID),
		MetricDescriptor: &metricpb.MetricDescriptor{
			Name:        descriptor.ID,
			Type:        fmt.Sprintf("custom.googleapis.com/%s/%s", descriptor.ServiceName, descriptor.ID),
			MetricKind:  kind,
			ValueType:   valueType,
			Unit:        descriptor.Unit,
			Description: descriptor.Description,
			DisplayName: descriptor.DisplayName,
		},
	})

	return err
}

// Close closes the client connection to StackDriver
func (handler *StackDriverHandler) Close() error {
	// Closes the client and flushes the data to Stackdriver.
	var err error
	if err = handler.Client.Close(); err != nil {
		handler.counterMapMutex.Lock()
		handler.counters = make(map[string]counterMapData)
		handler.gauges = make(map[string]gaugeMapData)
		handler.counterMapMutex.Unlock()
	}

	return err
}

func (handler *StackDriverHandler) getMonitoredResource() (*monitoredrespb.MonitoredResource, error) {
	monitoredResource := &monitoredrespb.MonitoredResource{
		Type: "global",
		Labels: map[string]string{
			"project_id": handler.ProjectID,
		},
	}

	if metadataapi.OnGCE() {
		projectIDFromMetadata, err := metadataapi.ProjectID()
		if err != nil {
			return monitoredResource, err
		}

		instanceID, err := metadataapi.InstanceID()
		if err != nil {
			return monitoredResource, err
		}

		zone, err := metadataapi.Zone()
		if err != nil {
			return monitoredResource, err
		}

		monitoredResource = &monitoredrespb.MonitoredResource{
			Type: "gce_instance",
			Labels: map[string]string{
				"project_id":  projectIDFromMetadata,
				"instance_id": instanceID,
				"zone":        zone,
			},
		}

		return monitoredResource, nil
	}

	monitoredResource = &monitoredrespb.MonitoredResource{
		Type: "global",
		Labels: map[string]string{
			"project_id": handler.ProjectID,
		},
	}

	return monitoredResource, nil
}

// StackDriverCounter is an atomic implementation of a counter based on go-kit's generic counter to fit our Counter interface
type StackDriverCounter struct {
	bits        uint64
	labels      map[string]string
	labelsMutex sync.RWMutex
}

// Add adds the delta to the counter's value.
func (c *StackDriverCounter) Add(delta float64) {
	for {
		var (
			old  = atomic.LoadUint64(&c.bits)
			newf = math.Float64frombits(old) + delta
			new  = math.Float64bits(newf)
		)
		if atomic.CompareAndSwapUint64(&c.bits, old, new) {
			break
		}
	}
}

// Value returns the counter's current value.
func (c *StackDriverCounter) Value() float64 {
	return math.Float64frombits(atomic.LoadUint64(&c.bits))
}

// ValueReset returns the counter's current value and resets the counter.
func (c *StackDriverCounter) ValueReset() float64 {
	for {
		var (
			old  = atomic.LoadUint64(&c.bits)
			newf = 0.0
			new  = math.Float64bits(newf)
		)
		if atomic.CompareAndSwapUint64(&c.bits, old, new) {
			return math.Float64frombits(old)
		}
	}
}

// AddLabels adds more labels to the counter. If a label being added already exists, the value is overwritten.
func (c *StackDriverCounter) AddLabels(labels map[string]string) {
	c.labelsMutex.Lock()
	defer c.labelsMutex.Unlock()

	for k, v := range labels {
		c.labels[k] = v
	}
}

// Labels is a returns the a copy of current list of labels attached to the counter.
func (c *StackDriverCounter) Labels() map[string]string {
	labelsCopy := make(map[string]string)
	c.labelsMutex.RLock()
	defer c.labelsMutex.RUnlock()

	for k, v := range c.labels {
		labelsCopy[k] = v
	}

	return labelsCopy
}

// ClearLabels removes all existing labels attached to the counter
func (c *StackDriverCounter) ClearLabels() {
	c.labelsMutex.Lock()
	defer c.labelsMutex.Unlock()

	c.labels = make(map[string]string)
}

// StackDriverGauge is an atomic implementation of a gauge based on go-kit's generic gauge to fit our Gauge interface
type StackDriverGauge struct {
	bits        uint64
	labels      map[string]string
	labelsMutex sync.RWMutex
}

// Set sets the gauge's value.
func (g *StackDriverGauge) Set(value float64) {
	atomic.StoreUint64(&g.bits, math.Float64bits(value))
}

// Add adds the delta to the counter's value.
func (g *StackDriverGauge) Add(delta float64) {
	for {
		var (
			old  = atomic.LoadUint64(&g.bits)
			newf = math.Float64frombits(old) + delta
			new  = math.Float64bits(newf)
		)
		if atomic.CompareAndSwapUint64(&g.bits, old, new) {
			break
		}
	}
}

// Value returns the counter's current value.
func (g *StackDriverGauge) Value() float64 {
	return math.Float64frombits(atomic.LoadUint64(&g.bits))
}

// ValueReset returns the counter's current value and resets the counter.
func (g *StackDriverGauge) ValueReset() float64 {
	for {
		var (
			old  = atomic.LoadUint64(&g.bits)
			newf = 0.0
			new  = math.Float64bits(newf)
		)
		if atomic.CompareAndSwapUint64(&g.bits, old, new) {
			return math.Float64frombits(old)
		}
	}
}

// AddLabels adds more labels to the counter. If a label being added already exists, the value is overwritten.
func (g *StackDriverGauge) AddLabels(labels map[string]string) {
	g.labelsMutex.Lock()
	defer g.labelsMutex.Unlock()

	for k, v := range labels {
		g.labels[k] = v
	}
}

// Labels is a returns the a copy of current list of labels attached to the counter.
func (g *StackDriverGauge) Labels() map[string]string {
	labelsCopy := make(map[string]string)
	g.labelsMutex.RLock()
	defer g.labelsMutex.RUnlock()

	for k, v := range g.labels {
		labelsCopy[k] = v
	}

	return labelsCopy
}

// ClearLabels removes all existing labels attached to the counter
func (g *StackDriverGauge) ClearLabels() {
	g.labelsMutex.Lock()
	defer g.labelsMutex.Unlock()

	g.labels = make(map[string]string)
}
