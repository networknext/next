package metrics

import (
	"context"
	"fmt"
	"os"
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

// StackDriverMetricHandler is an implementation of the MetricHandler interface that handles metrics for StackDriver
type StackDriverMetricHandler struct {
	client *monitoring.MetricClient

	clusterLocation string
	clusterName     string
	podName         string
	containerName   string
	namespaceName   string
	projectID       string

	pushMetricsChan chan []Metric
}

// NewStackDriverMetricHandler sets up a new stackdriver metric handler
func NewStackDriverMetricHandler() (*StackDriverMetricHandler, error) {
	metricHandler := &StackDriverMetricHandler{
		clusterLocation: os.Getenv("GOOGLE_CLOUD_METRICS_CLUSTER_LOCATION"),
		clusterName:     os.Getenv("GOOGLE_CLOUD_METRICS_CLUSTER_NAME"),
		podName:         os.Getenv("GOOGLE_CLOUD_METRICS_POD_NAME"),
		containerName:   os.Getenv("GOOGLE_CLOUD_METRICS_CONTAINER_NAME"),
		namespaceName:   os.Getenv("GOOGLE_CLOUD_METRICS_NAMESPACE_NAME"),
		projectID:       os.Getenv("GOOGLE_CLOUD_METRICS_PROJECT"),
		pushMetricsChan: make(chan []Metric),
	}

	// Configure logging
	logger := log.NewLogfmtLogger(os.Stdout)
	{
		switch os.Getenv("BACKEND_LOG_LEVEL") {
		case "none":
			logger = level.NewFilter(logger, level.AllowNone())
		case level.ErrorValue().String():
			logger = level.NewFilter(logger, level.AllowError())
		case level.WarnValue().String():
			logger = level.NewFilter(logger, level.AllowWarn())
		case level.InfoValue().String():
			logger = level.NewFilter(logger, level.AllowInfo())
		case level.DebugValue().String():
			logger = level.NewFilter(logger, level.AllowDebug())
		default:
			logger = level.NewFilter(logger, level.AllowWarn())
		}

		logger = log.With(logger, "ts", log.DefaultTimestampUTC)
	}

	ctx := context.Background()

	// Create a Stackdriver metrics client
	var err error
	metricHandler.client, err = monitoring.NewMetricClient(ctx)
	if err != nil {
		return nil, err
	}

	// Sends the metrics to StackDriver in a separate goroutine
	go func() {
		// The maximum number of metrics to send up to StackDriver in one push
		metricsMaxProcessCount := 200

		for metrics := range metricHandler.pushMetricsChan {
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
				case *MetricTypeBool:
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
				case *MetricTypeInt64:
					value = &monitoringpb.TypedValue{
						Value: &monitoringpb.TypedValue_Int64Value{
							Int64Value: int64(metrics[metricIndex].Gauge.Value()),
						},
					}
				case *MetricTypeDouble:
					value = &monitoringpb.TypedValue{
						Value: &monitoringpb.TypedValue_DoubleValue{
							DoubleValue: metrics[metricIndex].Gauge.Value(),
						},
					}
				}

				// Create a time series object for each metric
				timeSeries[metricIndex] = &monitoringpb.TimeSeries{
					Metric: &metricpb.Metric{
						Type:   fmt.Sprintf("custom.googleapis.com/%s/%s", metrics[metricIndex].MetricDescriptor.PackageName, metrics[metricIndex].MetricDescriptor.MetricID),
						Labels: labels,
					},
					Resource: metricHandler.getMonitoredResource(),
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
			for i := 0; i < metricsCount; i += metricsMaxProcessCount {
				// Calculate the number of metrics to process this iteration
				e := i + metricsMaxProcessCount
				if e > metricsCount {
					e = metricsCount
				}

				if err = metricHandler.client.CreateTimeSeries(ctx, &monitoringpb.CreateTimeSeriesRequest{
					Name:       monitoring.MetricProjectPath(metricHandler.projectID),
					TimeSeries: timeSeries[i:e],
				}); err != nil {
					level.Error(logger).Log("msg", "Failed to write time series data", "err", err)
				}
			}
		}
	}()

	return metricHandler, nil
}

// CreateMetric creates the metric on StackDriver using the given metric descriptor.
func (metricHandler *StackDriverMetricHandler) CreateMetric(metricDescriptor *MetricDescriptor, gauge *generic.Gauge) (Metric, error) {
	_, err := metricHandler.client.CreateMetricDescriptor(context.Background(), &monitoringpb.CreateMetricDescriptorRequest{
		Name: monitoring.MetricProjectPath(metricHandler.projectID),
		MetricDescriptor: &metricpb.MetricDescriptor{
			Name:        gauge.Name,
			Type:        fmt.Sprintf("custom.googleapis.com/%s/%s", metricDescriptor.PackageName, metricDescriptor.MetricID),
			MetricKind:  metric.MetricDescriptor_GAUGE,
			ValueType:   metricpb.MetricDescriptor_ValueType(ValueTypeMap[metricDescriptor.ValueType.ValueType.getTypeName()]),
			Unit:        metricDescriptor.Unit,
			Description: metricDescriptor.Description,
			DisplayName: gauge.Name,
		},
	})

	if err != nil {
		if status.Code(err) == codes.AlreadyExists {
			return Metric{}, fmt.Errorf("Attempted to create metric with name '%s' but a metric with that name already exists in StackDriver", gauge.Name)
		}

		return Metric{}, err
	}

	return Metric{
		MetricDescriptor: metricDescriptor,
		Gauge:            gauge,
	}, nil
}

// SubmitMetric updates the metric on StackDriver
func (metricHandler *StackDriverMetricHandler) SubmitMetric(metric Metric) {
	metricHandler.pushMetricsChan <- []Metric{metric}
}

// SubmitMetrics updates a list of metrics on StackDriver
func (metricHandler *StackDriverMetricHandler) SubmitMetrics(metrics []Metric) {
	metricHandler.pushMetricsChan <- metrics
}

// DeleteMetric deletes the metric represented by the given metric descriptor.
// Only the MetricDescriptor's PackageName and MetricID need to be filled in for the delete to work.
// This shouldn't ever be called in production because metrics shouldn't ever be deleted.
func (metricHandler *StackDriverMetricHandler) DeleteMetric(metricDescriptor *MetricDescriptor) error {
	return metricHandler.client.DeleteMetricDescriptor(context.Background(), &monitoringpb.DeleteMetricDescriptorRequest{
		Name: fmt.Sprintf("%s/metricDescriptors/custom.googleapis.com/%s/%s", monitoring.MetricProjectPath(metricHandler.projectID), metricDescriptor.PackageName, metricDescriptor.MetricID),
	})
}

// Close closes the client connection to StackDriver
func (metricHandler *StackDriverMetricHandler) Close() error {
	// Closes the client and flushes the data to Stackdriver.
	if err := metricHandler.client.Close(); err != nil {
		return err
	}

	return nil
}

func (metricHandler *StackDriverMetricHandler) getMonitoredResource() *monitoredrespb.MonitoredResource {
	var monitoredResource *monitoredrespb.MonitoredResource

	if metricHandler.clusterLocation != "" && metricHandler.clusterName != "" && metricHandler.podName != "" && metricHandler.containerName != "" {
		if metricHandler.namespaceName == "" {
			metricHandler.namespaceName = "default"
		}

		monitoredResource = &monitoredrespb.MonitoredResource{
			Type: "k8s_container",
			Labels: map[string]string{
				"project_id":     metricHandler.projectID,
				"location":       metricHandler.clusterLocation,
				"cluster_name":   metricHandler.clusterName,
				"namespace_name": metricHandler.namespaceName,
				"pod_name":       metricHandler.podName,
				"container_name": metricHandler.containerName,
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
					"project_id": metricHandler.projectID,
				},
			}
		}
	} else {
		monitoredResource = &monitoredrespb.MonitoredResource{
			Type: "global",
			Labels: map[string]string{
				"project_id": metricHandler.projectID,
			},
		}
	}

	return monitoredResource
}
