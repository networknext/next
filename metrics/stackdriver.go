package metrics

import (
	"context"
	"fmt"
	"math"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-kit/kit/metrics/generic"

	metadataapi "cloud.google.com/go/compute/metadata"
	monitoring "cloud.google.com/go/monitoring/apiv3"
	googlepb "github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/api/option"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
	monitoredrespb "google.golang.org/genproto/googleapis/api/monitoredres"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// valueTypeMap is a map from the value's named type to StackDriver's MetricDescriptor value type
var valueTypeMap = map[string]metricpb.MetricDescriptor_ValueType{
	"BOOL":   metricpb.MetricDescriptor_BOOL,
	"INT64":  metricpb.MetricDescriptor_INT64,
	"DOUBLE": metricpb.MetricDescriptor_DOUBLE,
}

// StackDriverHandler is an implementation of the Handler interface that handles metrics for StackDriver
type StackDriverHandler struct {
	ProjectID   string
	Credentials []byte

	// Optional kubernetes container data. If these are set, the client will know that the monitored resource is running in a kubernetes container.
	// If they are not set, the client will check to see if the monitored resource is running in a GCE instance. If it's not, it will default to global.
	ClusterLocation string
	ClusterName     string
	PodName         string
	ContainerName   string
	NamespaceName   string // If this is not set, it will default to "default"

	client *monitoring.MetricClient

	counters   map[string]counterMapData
	gauges     map[string]gaugeMapData
	histograms map[string]histogramMapData

	counterMapMutex   sync.Mutex
	gaugeMapMutex     sync.Mutex
	histogramMapMutex sync.Mutex
}

type counterMapData struct {
	descriptor *Descriptor
	counter    Counter
}

type gaugeMapData struct {
	descriptor *Descriptor
	gauge      Gauge
}

type histogramMapData struct {
	descriptor *Descriptor
	histogram  Histogram
	buckets    int
}

// Open opens the client connection to StackDriver. This must be done before any metrics are created, deleted, or fetched.
func (handler *StackDriverHandler) Open(ctx context.Context) error {
	// Lock the map mutexes just in case Open is called more than once
	handler.counterMapMutex.Lock()
	handler.gaugeMapMutex.Lock()
	handler.histogramMapMutex.Lock()

	handler.counters = make(map[string]counterMapData)
	handler.gauges = make(map[string]gaugeMapData)
	handler.histograms = make(map[string]histogramMapData)

	handler.counterMapMutex.Unlock()
	handler.gaugeMapMutex.Unlock()
	handler.histogramMapMutex.Unlock()

	// Create a Stackdriver metrics client
	var err error
	handler.client, err = monitoring.NewMetricClient(ctx, option.WithCredentialsJSON(handler.Credentials))
	return err
}

// WriteLoop is responsible for sending the metrics up to StackDriver. Call in a separate goroutine.
// Pass a duration in seconds to have the routine send metrics up to StackDriver periodically.
// If the duration is less than or equal to 0, a default of 1 minute is used.
// maxMetricsCount is the maximum number of metrics to send in one push to StackDriver. 200 is the maximum number of time series allowed in a single request.
func (handler *StackDriverHandler) WriteLoop(ctx context.Context, logger log.Logger, duration time.Duration, maxMetricsIncrement int) {
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
			handler.histogramMapMutex.Lock()

			metricsCount := len(handler.counters) + len(handler.gauges) + len(handler.histograms)
			timeSeries := make([]*monitoringpb.TimeSeries, metricsCount)

			for _, mapData := range handler.counters {
				labels := convertLabelValues(mapData.counter.LabelValues())
				value := convertValue(mapData.descriptor, mapData.counter.Value())
				timeSeries[index] = handler.newTimeSeries(mapData.descriptor, labels, value)

				mapData.counter.ValueReset()
				index++
			}

			for _, mapData := range handler.gauges {
				labels := convertLabelValues(mapData.gauge.LabelValues())
				value := convertValue(mapData.descriptor, mapData.gauge.Value())
				timeSeries[index] = handler.newTimeSeries(mapData.descriptor, labels, value)

				mapData.gauge.Set(0)
				index++
			}

			for _, mapData := range handler.histograms {
				labels := convertLabelValues(mapData.histogram.LabelValues())
				value := convertValue(mapData.descriptor, mapData.histogram.Quantile(0.5))
				timeSeries[index] = handler.newTimeSeries(mapData.descriptor, labels, value)

				// Don't reset the histogram
				index++
			}

			handler.counterMapMutex.Unlock()
			handler.gaugeMapMutex.Unlock()
			handler.histogramMapMutex.Unlock()

			// Send the time series objects to StackDriver with a maximum send size to avoid overloading
			for i := 0; i < metricsCount; i += maxMetricsIncrement {
				// Calculate the number of metrics to process this iteration
				e := i + maxMetricsIncrement
				if e > metricsCount {
					e = metricsCount
				}

				if err := handler.client.CreateTimeSeries(ctx, &monitoringpb.CreateTimeSeriesRequest{
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
	if err := handler.createMetric(ctx, descriptor, metricpb.MetricDescriptor_CUMULATIVE); err != nil {
		return nil, err
	}

	// Create the counter for updating the metric
	counter := generic.NewCounter(descriptor.ID)

	// Add the metric to the map data
	handler.counterMapMutex.Lock()
	handler.counters[descriptor.ID] = counterMapData{
		descriptor: descriptor,
		counter:    counter,
	}
	handler.counterMapMutex.Unlock()

	return counter, nil
}

// NewGauge creates a metric and returns a new gauge to update it.
// If the metric already exists, then the returned new gauge updates the metric instead of any other existing gauge.
func (handler *StackDriverHandler) NewGauge(ctx context.Context, descriptor *Descriptor) (Gauge, error) {
	// Create the metric in StackDriver
	if err := handler.createMetric(ctx, descriptor, metricpb.MetricDescriptor_GAUGE); err != nil {
		return nil, err
	}

	// Create the gauge for updating the metric
	gauge := generic.NewGauge(descriptor.ID)

	// Add the metric to the map data
	handler.gaugeMapMutex.Lock()
	handler.gauges[descriptor.ID] = gaugeMapData{
		descriptor: descriptor,
		gauge:      gauge,
	}
	handler.gaugeMapMutex.Unlock()

	return gauge, nil
}

// NewHistogram creates a metric and returns a new histogram to observe it.
// If the metric already exists, then the returned new histogram observes the metric instead of any other existing histogram.
func (handler *StackDriverHandler) NewHistogram(ctx context.Context, descriptor *Descriptor, buckets int) (Histogram, error) {
	// Create the metric in StackDriver
	if err := handler.createMetric(ctx, descriptor, metricpb.MetricDescriptor_DELTA); err != nil {
		return nil, err
	}

	// Create the gauge for updating the metric
	histogram := generic.NewHistogram(descriptor.ID, buckets)

	// Add the metric to the map data
	handler.histogramMapMutex.Lock()
	handler.histograms[descriptor.ID] = histogramMapData{
		descriptor: descriptor,
		histogram:  histogram,
		buckets:    buckets,
	}
	handler.histogramMapMutex.Unlock()

	return histogram, nil
}

// createMetric creates the metric on StackDriver using the given metric descriptor.
// If the metric already exists on StackDriver, it will overwrite it.
func (handler *StackDriverHandler) createMetric(ctx context.Context, descriptor *Descriptor, kind metricpb.MetricDescriptor_MetricKind) error {
	stackdriverDescriptor, err := handler.client.CreateMetricDescriptor(ctx, &monitoringpb.CreateMetricDescriptorRequest{
		Name: monitoring.MetricProjectPath(handler.ProjectID),
		MetricDescriptor: &metricpb.MetricDescriptor{
			Name:        descriptor.ID,
			Type:        fmt.Sprintf("custom.googleapis.com/%s/%s", descriptor.ServiceName, descriptor.ID),
			MetricKind:  kind,
			ValueType:   valueTypeMap[descriptor.ValueType.ValueType.getTypeName()],
			Unit:        descriptor.Unit,
			Description: descriptor.Description,
			DisplayName: descriptor.DisplayName,
		},
	})

	if err != nil {
		if status.Code(err) != codes.AlreadyExists {
			return err
		}

		// The metric already exists, check if it needs to be overwritten
		if kind != stackdriverDescriptor.MetricKind ||
			valueTypeMap[descriptor.ValueType.ValueType.getTypeName()] != stackdriverDescriptor.ValueType ||
			descriptor.Unit != stackdriverDescriptor.Unit ||
			descriptor.Description != stackdriverDescriptor.Description ||
			descriptor.DisplayName != stackdriverDescriptor.DisplayName {
			// The given descriptor differs from stackdriver, so overwrite the metric
			if err = handler.client.DeleteMetricDescriptor(ctx, &monitoringpb.DeleteMetricDescriptorRequest{
				Name: fmt.Sprintf("projects/%s/metricDescriptors/custom.googleapis.com/%s/%s", handler.ProjectID, descriptor.ServiceName, descriptor.ID),
			}); err != nil {
				return err
			}

			// Check if the metric has been deleted, and when it has, create it again
			timer := time.Now().Unix()
			timeout := timer + 1
			for timeout <= timer {
				if _, err = handler.client.GetMetricDescriptor(ctx, &monitoringpb.GetMetricDescriptorRequest{
					Name: fmt.Sprintf("projects/%s/metricDescriptors/custom.googleapis.com/%s/%s", handler.ProjectID, descriptor.ServiceName, descriptor.ID),
				}); err == nil {
					break
				}

				timer = time.Now().Unix()
			}

			_, err = handler.client.CreateMetricDescriptor(ctx, &monitoringpb.CreateMetricDescriptorRequest{
				Name: monitoring.MetricProjectPath(handler.ProjectID),
				MetricDescriptor: &metricpb.MetricDescriptor{
					Name:        descriptor.ID,
					Type:        fmt.Sprintf("custom.googleapis.com/%s/%s", descriptor.ServiceName, descriptor.ID),
					MetricKind:  kind,
					ValueType:   valueTypeMap[descriptor.ValueType.ValueType.getTypeName()],
					Unit:        descriptor.Unit,
					Description: descriptor.Description,
					DisplayName: descriptor.DisplayName,
				},
			})
		}
	}

	return nil
}

// Close closes the client connection to StackDriver
func (handler *StackDriverHandler) Close() error {
	// Closes the client and flushes the data to Stackdriver.
	var err error
	if err = handler.client.Close(); err != nil {
		handler.counterMapMutex.Lock()
		handler.counters = make(map[string]counterMapData)
		handler.gauges = make(map[string]gaugeMapData)
		handler.histograms = make(map[string]histogramMapData)
		handler.counterMapMutex.Unlock()
	}

	return err
}

func (handler *StackDriverHandler) getMonitoredResource() *monitoredrespb.MonitoredResource {
	var monitoredResource *monitoredrespb.MonitoredResource

	if handler.ClusterLocation != "" && handler.ClusterName != "" && handler.PodName != "" && handler.ContainerName != "" {
		if handler.NamespaceName == "" {
			handler.NamespaceName = "default"
		}

		monitoredResource = &monitoredrespb.MonitoredResource{
			Type: "k8s_container",
			Labels: map[string]string{
				"project_id":     handler.ProjectID,
				"location":       handler.ClusterLocation,
				"cluster_name":   handler.ClusterName,
				"namespace_name": handler.NamespaceName,
				"pod_name":       handler.PodName,
				"container_name": handler.ContainerName,
			},
		}
	} else if metadataapi.OnGCE() {
		projectIDFromMetadata, err1 := metadataapi.ProjectID()
		instanceID, err2 := metadataapi.InstanceID()
		zone, err3 := metadataapi.Zone()
		if err1 == nil && err2 == nil && err3 == nil {
			monitoredResource = &monitoredrespb.MonitoredResource{
				Type: "gce_instance",
				Labels: map[string]string{
					"project_id":  projectIDFromMetadata,
					"instance_id": instanceID,
					"zone":        zone,
				},
			}
		} else {
			monitoredResource = &monitoredrespb.MonitoredResource{
				Type: "global",
				Labels: map[string]string{
					"project_id": handler.ProjectID,
				},
			}
		}
	} else {
		monitoredResource = &monitoredrespb.MonitoredResource{
			Type: "global",
			Labels: map[string]string{
				"project_id": handler.ProjectID,
			},
		}
	}

	return monitoredResource
}

// Converts string slice to map
func convertLabelValues(labelValues []string) map[string]string {
	labels := make(map[string]string)

	// Convert the labels from a 1D string slice to a map
	// Label values are guaranteed to be even
	for i := 0; i < len(labelValues); i += 2 {
		labels[labelValues[i]] = labelValues[i+1]
	}

	return labels
}

// Converts the descriptor value type to stackdriver's metric value type
func convertValue(descriptor *Descriptor, value float64) *monitoringpb.TypedValue {
	var valueType *monitoringpb.TypedValue
	switch descriptor.ValueType.ValueType.(type) {
	case TypeBool:
		var b bool
		if value != 0.0 {
			b = true
		} else {
			b = false
		}

		valueType = &monitoringpb.TypedValue{
			Value: &monitoringpb.TypedValue_BoolValue{
				BoolValue: b,
			},
		}
	case TypeInt64:
		valueType = &monitoringpb.TypedValue{
			Value: &monitoringpb.TypedValue_Int64Value{
				Int64Value: int64(math.Round(value)),
			},
		}
	case TypeDouble:
		valueType = &monitoringpb.TypedValue{
			Value: &monitoringpb.TypedValue_DoubleValue{
				DoubleValue: value,
			},
		}
	}

	return valueType
}

func (handler *StackDriverHandler) newTimeSeries(descriptor *Descriptor, labels map[string]string, value *monitoringpb.TypedValue) *monitoringpb.TimeSeries {
	return &monitoringpb.TimeSeries{
		Metric: &metricpb.Metric{
			Type:   fmt.Sprintf("custom.googleapis.com/%s/%s", descriptor.ServiceName, descriptor.ID),
			Labels: labels,
		},
		Resource: handler.getMonitoredResource(),
		Points: []*monitoringpb.Point{
			&monitoringpb.Point{
				Interval: &monitoringpb.TimeInterval{
					EndTime: &googlepb.Timestamp{
						Seconds: time.Now().Unix(),
					},
				},
				Value: value,
			},
		},
	}
}
