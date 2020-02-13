package metrics

import (
	"context"
	"fmt"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-kit/kit/metrics/generic"

	metadataapi "cloud.google.com/go/compute/metadata"
	monitoring "cloud.google.com/go/monitoring/apiv3"
	googlepb "github.com/golang/protobuf/ptypes/timestamp"
	"google.golang.org/genproto/googleapis/api/metric"
	metricpb "google.golang.org/genproto/googleapis/api/metric"
	monitoredrespb "google.golang.org/genproto/googleapis/api/monitoredres"
	monitoringpb "google.golang.org/genproto/googleapis/monitoring/v3"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// ValueTypeMap is a map from the value's named type to an integer reprentation for use with the StackDriver API
var ValueTypeMap = map[string]int32{
	"bool":   1,
	"int64":  2,
	"double": 3,
}

// StackDriverHandler is an implementation of the Handler interface that handles metrics for StackDriver
type StackDriverHandler struct {
	Client *monitoring.MetricClient

	ClusterLocation string
	ClusterName     string
	PodName         string
	ContainerName   string
	NamespaceName   string
	ProjectID       string

	PushMetricsChan chan []Handle
}

// MetricSubmitRoutine is responsible for sending the metrics up to StackDriver. Call in a separate goroutine.
// maxMetricsCount is the maximum number of metrics to send in one push to StackDriver. If you're unsure, 200 is a good number.
// MetricSubmitRoutine will push a value of true through the channel when it first begins to indicate that it is ready to receive
func (handler *StackDriverHandler) MetricSubmitRoutine(logger log.Logger, maxMetricsCount int) {
	for metrics := range handler.PushMetricsChan {
		metricsCount := len(metrics)
		timeSeries := make([]*monitoringpb.TimeSeries, metricsCount)
		labels := make(map[string]string)

		// Preprocess all metrics in the slice to create time series objects
		for metricIndex := 0; metricIndex < metricsCount; metricIndex++ {
			// Convert the labels from a 1D string slice to a map
			labelValues := metrics[metricIndex].Gauge.LabelValues()
			for labelIndex := 0; labelIndex < len(labelValues); labelIndex++ {
				labels[labelValues[labelIndex]] = labelValues[labelIndex+1]
			}

			// Gets the metric value from the metric descriptor type
			var value *monitoringpb.TypedValue
			switch metrics[metricIndex].MetricDescriptor.ValueType.ValueType.(type) {
			case *TypeBool:
				var b bool
				if metrics[metricIndex].Gauge.Value() != 0 {
					b = true
				} else {
					b = false
				}

				value = &monitoringpb.TypedValue{
					Value: &monitoringpb.TypedValue_BoolValue{
						BoolValue: b,
					},
				}
			case *TypeInt64:
				value = &monitoringpb.TypedValue{
					Value: &monitoringpb.TypedValue_Int64Value{
						Int64Value: int64(metrics[metricIndex].Gauge.Value()),
					},
				}
			case *TypeDouble:
				value = &monitoringpb.TypedValue{
					Value: &monitoringpb.TypedValue_DoubleValue{
						DoubleValue: metrics[metricIndex].Gauge.Value(),
					},
				}
			}

			// Create a time series object for each metric
			timeSeries[metricIndex] = &monitoringpb.TimeSeries{
				Metric: &metricpb.Metric{
					Type:   fmt.Sprintf("custom.googleapis.com/%s/%s", metrics[metricIndex].MetricDescriptor.PackageName, metrics[metricIndex].MetricDescriptor.ID),
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

		// Send the time series objects to StackDriver with a maximum send size to avoid overloading
		for i := 0; i < metricsCount; i += maxMetricsCount {
			// Calculate the number of metrics to process this iteration
			e := i + maxMetricsCount
			if e > metricsCount {
				e = metricsCount
			}

			if err := handler.Client.CreateTimeSeries(context.Background(), &monitoringpb.CreateTimeSeriesRequest{
				Name:       monitoring.MetricProjectPath(handler.ProjectID),
				TimeSeries: timeSeries[i:e],
			}); err != nil {
				level.Error(logger).Log("msg", "Failed to write time series data", "err", err)
			}
		}
	}
}

// CreateMetric creates the metric on StackDriver using the given metric descriptor.
func (handler *StackDriverHandler) CreateMetric(descriptor *Descriptor, gauge *generic.Gauge) (Handle, error) {
	_, err := handler.Client.CreateMetricDescriptor(context.Background(), &monitoringpb.CreateMetricDescriptorRequest{
		Name: monitoring.MetricProjectPath(handler.ProjectID),
		MetricDescriptor: &metricpb.MetricDescriptor{
			Name:        gauge.Name,
			Type:        fmt.Sprintf("custom.googleapis.com/%s/%s", descriptor.PackageName, descriptor.ID),
			MetricKind:  metric.MetricDescriptor_GAUGE,
			ValueType:   metricpb.MetricDescriptor_ValueType(ValueTypeMap[descriptor.ValueType.ValueType.getTypeName()]),
			Unit:        descriptor.Unit,
			Description: descriptor.Description,
			DisplayName: gauge.Name,
		},
	})

	if err != nil {
		if status.Code(err) == codes.AlreadyExists {
			return Handle{}, fmt.Errorf("Attempted to create metric with name '%s' but a metric with that name already exists in StackDriver", gauge.Name)
		}

		return Handle{}, err
	}

	return Handle{
		MetricDescriptor: descriptor,
		Gauge:            gauge,
	}, nil
}

// SubmitMetric updates the metric on StackDriver
func (handler *StackDriverHandler) SubmitMetric(handle Handle) error {
	return handler.SubmitMetrics([]Handle{handle})
}

// SubmitMetrics updates a list of metrics on StackDriver
func (handler *StackDriverHandler) SubmitMetrics(handles []Handle) error {
	select {
	case handler.PushMetricsChan <- handles:
		return nil
	case <-time.After(5 * time.Second):
		return fmt.Errorf("Submit routine not running yet. Call MetricSubmitRoutine() in a goroutine before calling SubmitMetrics()")
	}
}

// DeleteMetric deletes the metric represented by the given descriptor.
// Only the Descriptor's PackageName and ID need to be filled in for the delete to work.
// This shouldn't ever be called in production because metrics shouldn't ever be deleted.
func (handler *StackDriverHandler) DeleteMetric(descriptor *Descriptor) error {
	return handler.Client.DeleteMetricDescriptor(context.Background(), &monitoringpb.DeleteMetricDescriptorRequest{
		Name: fmt.Sprintf("%s/metricDescriptors/custom.googleapis.com/%s/%s", monitoring.MetricProjectPath(handler.ProjectID), descriptor.PackageName, descriptor.ID),
	})
}

// Close closes the client connection to StackDriver
func (handler *StackDriverHandler) Close() error {
	// Closes the client and flushes the data to Stackdriver.
	if err := handler.Client.Close(); err != nil {
		return err
	}

	return nil
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
