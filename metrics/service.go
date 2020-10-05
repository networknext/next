package metrics

import "context"

// ServiceMetrics defines the set of metrics for any backend service.
type ServiceMetrics struct {
	Goroutines      Gauge
	MemoryAllocated Gauge
}

// EmptyServiceMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyServiceMetrics = ServiceMetrics{
	Goroutines:      &EmptyGauge{},
	MemoryAllocated: &EmptyGauge{},
}

// NewServiceMetrics creates the metrics a service will use.
func NewServiceMetrics(ctx context.Context, handler Handler, serviceName string) (*ServiceMetrics, error) {
	var err error
	m := &ServiceMetrics{}

	m.Goroutines, err = handler.NewGauge(ctx, &Descriptor{
		DisplayName: "Goroutines",
		ServiceName: serviceName,
		ID:          "goroutines",
		Unit:        "goroutines",
		Description: "The number of goroutines that the service is running.",
	})
	if err != nil {
		return nil, err
	}

	m.MemoryAllocated, err = handler.NewGauge(ctx, &Descriptor{
		DisplayName: "Memory Allocated",
		ServiceName: serviceName,
		ID:          "memory_allocated",
		Unit:        "MB",
		Description: "The amount of memory the service has allocated in megabytes.",
	})
	if err != nil {
		return nil, err
	}

	return m, nil
}
