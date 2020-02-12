package metrics

// MetricHandler handles creating and update metrics
type MetricHandler interface {
	Init() error
	CreateMetric(*MetricDescriptor) Metric
	SubmitMetric(Metric)
	SubmitMetrics([]Metric)
	Close() error
}
