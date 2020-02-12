package metrics

import (
	"context"
	"errors"
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

// MetricDescriptor describes a custom Stackdriver metric
type MetricDescriptor struct {
	packageName string
	metricType  string
	valueType   metricpb.MetricDescriptor_ValueType
	unit        string
	description string
}

// Metric is just a wrapper of a Stackdriver TimeSeries
type Metric struct {
	timeSeries *monitoringpb.TimeSeries
}

// Init sets up the metric handler
func (metricHandler *StackDriverMetricHandler) Init() error {
	metricHandler.clusterLocation = os.Getenv("GOOGLE_CLOUD_METRICS_CLUSTER_LOCATION")
	metricHandler.clusterName = os.Getenv("GOOGLE_CLOUD_METRICS_CLUSTER_NAME")
	metricHandler.podName = os.Getenv("GOOGLE_CLOUD_METRICS_POD_NAME")
	metricHandler.containerName = os.Getenv("GOOGLE_CLOUD_METRICS_CONTAINER_NAME")
	metricHandler.namespaceName = os.Getenv("GOOGLE_CLOUD_METRICS_NAMESPACE_NAME")
	metricHandler.projectID = os.Getenv("GOOGLE_CLOUD_METRICS_PROJECT")

	metricHandler.pushMetricsChan = make(chan []Metric)

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
		return err
	}

	// Sends the metrics to StackDriver
	go func() {
		for metric := range metricHandler.pushMetricsChan {
			for i := 0; i < len(metric); i += 200 {
				e := i + 200
				if e > len(metric) {
					e = len(metric)
				}

				// Since each metric only holds one timeseries and the create time series requests expects a slice of timeseries,
				// copy the timeseries into their own slice
				length := e - i
				timeSeries := make([]*monitoringpb.TimeSeries, length)
				for n := 0; n < length; n++ {
					timeSeries[n] = metric[n].timeSeries
				}

				if err = metricHandler.client.CreateTimeSeries(ctx, &monitoringpb.CreateTimeSeriesRequest{
					Name:       monitoring.MetricProjectPath(metricHandler.projectID),
					TimeSeries: timeSeries,
				}); err != nil {
					level.Error(logger).Log("msg", "Failed to write time series data", "err", err)
					os.Exit(1)
				}
			}
		}
	}()

	return nil
}

// CreateMetric creates the metric on StackDriver using the given metric descriptor
func (metricHandler *StackDriverMetricHandler) CreateMetric(metricDescriptor *MetricDescriptor, counter *generic.Counter) (Metric, error) {
	_, err := metricHandler.client.GetMetricDescriptor(context.Background(), &monitoringpb.GetMetricDescriptorRequest{
		Name: fmt.Sprintf("projects/%s/metricDescriptors/custom.googleapis.com/%s/%s", metricHandler.projectID, metricDescriptor.packageName, metricDescriptor.metricType),
	})

	if err != nil && status.Code(err) == codes.NotFound {
		_, err = metricHandler.client.CreateMetricDescriptor(context.Background(), &monitoringpb.CreateMetricDescriptorRequest{
			Name: monitoring.MetricProjectPath(metricHandler.projectID),
			MetricDescriptor: &metricpb.MetricDescriptor{
				Name:        counter.Name,
				Type:        fmt.Sprintf("custom.googleapis.com/%s/%s", metricDescriptor.packageName, metricDescriptor.metricType),
				MetricKind:  metric.MetricDescriptor_GAUGE,
				ValueType:   metricDescriptor.valueType,
				Unit:        metricDescriptor.unit,
				Description: metricDescriptor.description,
				DisplayName: counter.Name,
			},
		})
	} else {
		return Metric{}, errors.New("Attempted to create metric descriptor '%s' of type %s when that descriptor already exists")
	}

	// Since the label values are stored as an array, convert them to a map
	var labels map[string]string
	labelValues := counter.LabelValues()
	for i := 0; i < len(labelValues); i += 2 {
		labels[labelValues[i]] = labelValues[i+1]
	}

	return Metric{
		&monitoringpb.TimeSeries{
			Metric: &metricpb.Metric{
				Type:   fmt.Sprintf("custom.googleapis.com/%s/%s", metricDescriptor.packageName, metricDescriptor.metricType),
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
					Value: &monitoringpb.TypedValue{
						Value: &monitoringpb.TypedValue_Int64Value{
							Int64Value: int64(counter.Value()),
						},
					},
				},
			},
		},
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
