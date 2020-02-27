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
	"google.golang.org/genproto/googleapis/api/metric"
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

// valueTypeMapReverse is a reverse map to convert from StackDriver's MetricDescriptor value type to the metric package's value type
var valueTypeMapReverse = map[metricpb.MetricDescriptor_ValueType]Type{
	metricpb.MetricDescriptor_BOOL:   TypeBool{},
	metricpb.MetricDescriptor_INT64:  TypeInt64{},
	metricpb.MetricDescriptor_DOUBLE: TypeDouble{},
}

// StackDriverHandler is an implementation of the Handler interface that handles metrics for StackDriver
type StackDriverHandler struct {
	Client    *monitoring.MetricClient
	ProjectID string

	// Optional kubernetes container data. If these are set, the client will know that the monitored resource is running in a kubernetes container.
	// If they are not set, the client will check to see if the monitored resource is running in a GCE instance. If it's not, it will default to global.
	ClusterLocation string
	ClusterName     string
	PodName         string
	ContainerName   string
	NamespaceName   string // If this is not set, it will default to "default"

	submitFrequency float64

	metricsMap      map[string]Handle
	metricsMapMutex sync.Mutex
}

// Open opens the client connection to StackDriver. This must be done before any metrics are created, deleted, or fetched.
func (handler *StackDriverHandler) Open(ctx context.Context, credentials []byte) error {
	handler.metricsMap = make(map[string]Handle)

	// Create a Stackdriver metrics client
	var err error
	handler.Client, err = monitoring.NewMetricClient(ctx, option.WithCredentialsJSON(credentials))
	return err
}

// MetricSubmitRoutine is responsible for sending the metrics up to StackDriver. Call in a separate goroutine.
// Pass a duration in seconds to have the routine send metrics up to StackDriver periodically.
// If the duration is less than or equal to 0, a default of 1 minute is used.
// maxMetricsCount is the maximum number of metrics to send in one push to StackDriver. 200 is the maximum number of time series allowed in a single request.
func (handler *StackDriverHandler) MetricSubmitRoutine(ctx context.Context, logger log.Logger, duration time.Duration, maxMetricsIncrement int) {
	if duration <= 0 {
		duration = time.Minute
	}

	ticker := time.NewTicker(duration)
	handler.submitFrequency = 1.0 / duration.Seconds()

	for {
		select {
		case <-ticker.C:
			labels := make(map[string]string)

			// Preprocess all metrics in the map to create time series objects
			index := 0
			handler.metricsMapMutex.Lock()
			metricsCount := len(handler.metricsMap)
			timeSeries := make([]*monitoringpb.TimeSeries, metricsCount)
			for _, handle := range handler.metricsMap {
				// Convert the labels from a 1D string slice to a map
				labelValues := handle.Gauge.LabelValues() // LabelValues guarantees that there is an even number of label values
				for i := 0; i < len(labelValues); i += 2 {
					labels[labelValues[i]] = labelValues[i+1]
				}

				// Gets the metric value from the metric descriptor type
				var value *monitoringpb.TypedValue
				switch handle.Descriptor.ValueType.ValueType.(type) {
				case TypeBool:
					var b bool
					if handle.Gauge.Value() != 0 {
						b = true
					} else {
						b = false
					}

					value = &monitoringpb.TypedValue{
						Value: &monitoringpb.TypedValue_BoolValue{
							BoolValue: b,
						},
					}
				case TypeInt64:
					value = &monitoringpb.TypedValue{
						Value: &monitoringpb.TypedValue_Int64Value{
							Int64Value: int64(math.Round(handle.Gauge.Value())),
						},
					}
				case TypeDouble:
					value = &monitoringpb.TypedValue{
						Value: &monitoringpb.TypedValue_DoubleValue{
							DoubleValue: handle.Gauge.Value(),
						},
					}
				}

				// Create a time series object for each metric
				timeSeries[index] = &monitoringpb.TimeSeries{
					Metric: &metricpb.Metric{
						Type:   fmt.Sprintf("custom.googleapis.com/%s/%s", handle.Descriptor.ServiceName, handle.Descriptor.ID),
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

				// Reset metric data
				handle.Gauge.Set(0.0)
				handle.Histogram.Reset()

				index++
			}

			handler.metricsMapMutex.Unlock()

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
			handler.submitFrequency = 0
			return
		}
	}
}

// GetSubmitFrequency returns the frequency of how often the submit routine submits metrics in seconds.
// This will return 0 if the submit routine hasn't been started yet.
func (handler *StackDriverHandler) GetSubmitFrequency() float64 {
	return handler.submitFrequency
}

// CreateMetric creates the metric on StackDriver using the given metric descriptor.
// If the metric already exists on StackDriver, it will return a handle to it.
func (handler *StackDriverHandler) CreateMetric(ctx context.Context, descriptor *Descriptor) (Handle, error) {
	_, err := handler.Client.CreateMetricDescriptor(ctx, &monitoringpb.CreateMetricDescriptorRequest{
		Name: monitoring.MetricProjectPath(handler.ProjectID),
		MetricDescriptor: &metricpb.MetricDescriptor{
			Name:        descriptor.ID,
			Type:        fmt.Sprintf("custom.googleapis.com/%s/%s", descriptor.ServiceName, descriptor.ID),
			MetricKind:  metric.MetricDescriptor_GAUGE,
			ValueType:   valueTypeMap[descriptor.ValueType.ValueType.getTypeName()],
			Unit:        descriptor.Unit,
			Description: descriptor.Description,
			DisplayName: descriptor.DisplayName,
		},
	})

	// Create a StackDriver metric descriptor with the values copied from our descriptor.
	// If the metric doesn't exist yet, then it will use these values that we passed in.
	// If the metric already exists, then these will be overwritten with the values already in StackDriver.
	stackdriverDescriptor := &metricpb.MetricDescriptor{
		ValueType:   valueTypeMap[descriptor.ValueType.ValueType.getTypeName()],
		Unit:        descriptor.Unit,
		Description: descriptor.Description,
	}

	if err != nil {
		if status.Code(err) != codes.AlreadyExists {
			return Handle{}, err
		}

		stackdriverDescriptor, err = handler.Client.GetMetricDescriptor(ctx, &monitoringpb.GetMetricDescriptorRequest{
			Name: fmt.Sprintf("projects/%s/metricDescriptors/custom.googleapis.com/%s/%s", handler.ProjectID, descriptor.ServiceName, descriptor.ID),
		})

		if err != nil {
			return Handle{}, err
		}
	}

	handle := Handle{
		// After retrieving the existing StackDriver metric if necessary, use the resulting StackDriver descriptor values.
		// ServiceName and ID should always be the same for the same metrics so that can be taken from the passed descriptor.
		Descriptor: &Descriptor{
			ServiceName: descriptor.ServiceName,
			ID:          descriptor.ID,
			ValueType:   ValueType{ValueType: valueTypeMapReverse[stackdriverDescriptor.ValueType]},
			Unit:        stackdriverDescriptor.Unit,
			Description: stackdriverDescriptor.Description,
		},
		Histogram: &EmptyHistogram{}, // StackDriver doesn't need histograms
		Gauge:     generic.NewGauge(descriptor.ID),
	}

	handler.metricsMapMutex.Lock()
	handler.metricsMap[descriptor.ID] = handle
	handler.metricsMapMutex.Unlock()
	return handle, nil
}

// GetMetric returns a metric from the handler's metric map. Note that this fetches the metric from local memory,
// so the metric must always be created before it is fetched, even if it already exists in StackDriver.
func (handler *StackDriverHandler) GetMetric(id string) (Handle, bool) {
	handler.metricsMapMutex.Lock()
	defer handler.metricsMapMutex.Unlock()

	handle, success := handler.metricsMap[id]
	return handle, success
}

// DeleteMetric deletes the metric represented by the given descriptor.
// Only the Descriptor's ServiceName and ID need to be filled in for the delete to work.
// In production this always returns a permission denied error because metrics can't and shouldn't be deleted.
func (handler *StackDriverHandler) DeleteMetric(ctx context.Context, descriptor *Descriptor) error {
	err := handler.Client.DeleteMetricDescriptor(ctx, &monitoringpb.DeleteMetricDescriptorRequest{
		Name: fmt.Sprintf("%s/metricDescriptors/custom.googleapis.com/%s/%s", monitoring.MetricProjectPath(handler.ProjectID), descriptor.ServiceName, descriptor.ID),
	})

	if err == nil {
		handler.metricsMapMutex.Lock()
		delete(handler.metricsMap, descriptor.ID)
		handler.metricsMapMutex.Unlock()
	}

	return err
}

// Close closes the client connection to StackDriver
func (handler *StackDriverHandler) Close() error {
	// Closes the client and flushes the data to Stackdriver.
	var err error
	if err = handler.Client.Close(); err != nil {
		handler.metricsMapMutex.Lock()
		handler.metricsMap = make(map[string]Handle)
		handler.metricsMapMutex.Unlock()
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
