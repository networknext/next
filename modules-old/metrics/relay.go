package metrics

import (
	"context"
)

// RelayBackendStatus defines the metrics reported by the service's status endpoint
type RelayBackendStatus struct {
	// Service Information
	ServiceName string `json:"service_name"`
	GitHash     string `json:"git_hash"`
	Started     string `json:"started"`
	Uptime      string `json:"uptime"`

	// Service Metrics
	Goroutines      int     `json:"goroutines"`
	MemoryAllocated float64 `json:"mb_allocated"`

	// Relay Information
	DatacenterCount int `json:"datacenter_count"`
	RelayCount      int `json:"relay_count"`
	RouteCount      int `json:"route_count"`

	// Relay Update Information
	RelayUpdateInvocations        int `json:"relay_update_invocations"`
	RelayUpdateContentTypeFailure int `json:"relay_update_content_type_failure"`
	RelayUpdateUnbatchFailure     int `json:"relay_update_unbatch_failure"`
	RelayUpdateUnmarshalFailure   int `json:"relay_update_unmarshal_failure"`
	RelayUpdateRelayNotFound      int `json:"relay_update_relay_not_found"`

	// Durations
	LongCostMatrixUpdates  int     `json:"long_cost_matrix_updates"`
	LongRouteMatrixUpdates int     `json:"long_route_matrix_updates"`
	CostMatrixUpdateMs     float64 `json:"cost_matrix_update_ms"`
	RouteMatrixUpdateMs    float64 `json:"route_matrix_update_ms"`
	RelayUpdateMs          float64 `json:"relay_update_ms"`

	// Size
	CostMatrixBytes  int `json:"cost_matrix_bytes"`
	RouteMatrixBytes int `json:"route_matrix_bytes"`
}

type RelayBackendMetrics struct {
	Goroutines      Gauge
	MemoryAllocated Gauge
	RouteMatrix     RouteMatrixMetrics
}

var EmptyRelayBackendMetrics RelayBackendMetrics = RelayBackendMetrics{
	Goroutines:      &EmptyGauge{},
	MemoryAllocated: &EmptyGauge{},
	RouteMatrix:     EmptyRouteMatrixMetrics,
}

type RelayUpdateMetrics struct {
	Invocations   Counter
	DurationGauge Gauge
	ErrorMetrics  RelayUpdateErrorMetrics
}

var EmptyRelayUpdateMetrics RelayUpdateMetrics = RelayUpdateMetrics{
	Invocations:   &EmptyCounter{},
	DurationGauge: &EmptyGauge{},
	ErrorMetrics:  EmptyRelayUpdateErrorMetrics,
}

type RelayUpdateErrorMetrics struct {
	ContentTypeFailure Counter
	UnbatchFailure     Counter
	UnmarshalFailure   Counter
	RelayNotFound      Counter
}

var EmptyRelayUpdateErrorMetrics RelayUpdateErrorMetrics = RelayUpdateErrorMetrics{
	ContentTypeFailure: &EmptyCounter{},
	UnbatchFailure:     &EmptyCounter{},
	UnmarshalFailure:   &EmptyCounter{},
	RelayNotFound:      &EmptyCounter{},
}

func NewRelayBackendMetrics(ctx context.Context, metricsHandler Handler) (*RelayBackendMetrics, error) {
	relayBackendMetrics := RelayBackendMetrics{}
	var err error

	relayBackendMetrics.Goroutines, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Relay Backend Goroutine Count",
		ServiceName: "relay_backend",
		ID:          "relay_backend.goroutines",
		Unit:        "goroutines",
		Description: "The number of goroutines the relay backend service is using",
	})
	if err != nil {
		return nil, err
	}

	relayBackendMetrics.MemoryAllocated, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Relay Backend Memory Allocated",
		ServiceName: "relay_backend",
		ID:          "relay_backend.memory",
		Unit:        "MB",
		Description: "The amount of memory the relay backend service has allocated in MB",
	})
	if err != nil {
		return nil, err
	}

	relayBackendMetrics.RouteMatrix.DatacenterCount, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Route Matrix Datacenter Count",
		ServiceName: "relay_backend",
		ID:          "route_matrix.datacenter.count",
		Unit:        "datacenters",
		Description: "The number of datacenters the route matrix contains",
	})
	if err != nil {
		return nil, err
	}

	relayBackendMetrics.RouteMatrix.RelayCount, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Route Matrix Relay Count",
		ServiceName: "relay_backend",
		ID:          "route_matrix.relay.count",
		Unit:        "relays",
		Description: "The number of relays the route matrix contains",
	})
	if err != nil {
		return nil, err
	}

	relayBackendMetrics.RouteMatrix.RouteCount, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Route Matrix Route Count",
		ServiceName: "relay_backend",
		ID:          "route_matrix.route.count",
		Unit:        "routes",
		Description: "The number of routes the route matrix contains",
	})
	if err != nil {
		return nil, err
	}

	relayBackendMetrics.RouteMatrix.Bytes, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Route Matrix Size",
		ServiceName: "relay_backend",
		ID:          "route_matrix.bytes",
		Unit:        "bytes",
		Description: "How large the route matrix is in bytes",
	})
	if err != nil {
		return nil, err
	}

	return &relayBackendMetrics, nil
}

func NewRelayUpdateMetrics(ctx context.Context, metricsHandler Handler) (*RelayUpdateMetrics, error) {
	updateCount, err := metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Total relay update count",
		ServiceName: "relay_backend",
		ID:          "relay.update.count",
		Unit:        "requests",
		Description: "The total number of received relay update requests",
	})
	if err != nil {
		return nil, err
	}

	updateDuration, err := metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Relay update duration",
		ServiceName: "relay_backend",
		ID:          "relay.update.duration",
		Unit:        "milliseconds",
		Description: "How long it takes to process a relay update request.",
	})
	if err != nil {
		return nil, err
	}

	var em RelayUpdateErrorMetrics
	em.ContentTypeFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Total relay update content type failure count",
		ServiceName: "relay_backend",
		ID:          "relay.update.errors.content_type_failure.count",
		Unit:        "content_type_failure_failure",
		Description: "The total number of received updates that had the incorrect type content type",
	})
	if err != nil {
		return nil, err
	}

	em.UnbatchFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Total relay update unbatch failure count",
		ServiceName: "relay_backend",
		ID:          "relay.update.errors.unbatch_failure.count",
		Unit:        "unbatch_failure",
		Description: "The total number of received relay update batched requests that failed to be unbatched",
	})
	if err != nil {
		return nil, err
	}

	em.UnmarshalFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Total relay update unmarshal failure count",
		ServiceName: "relay_backend",
		ID:          "relay.update.errors.unmarshal_failure.count",
		Unit:        "unmarshal_failure",
		Description: "The total number of received relay update requests that resulted in unmarshal failure",
	})
	if err != nil {
		return nil, err
	}

	em.RelayNotFound, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Total relay update relay not found error count",
		ServiceName: "relay_backend",
		ID:          "relay.update.errors.not_found.count",
		Unit:        "relay_not_found",
		Description: "The total number of received relay update requests that resulted in relay not found error",
	})
	if err != nil {
		return nil, err
	}

	updateMetrics := RelayUpdateMetrics{
		Invocations:   updateCount,
		DurationGauge: updateDuration,
		ErrorMetrics:  em,
	}

	return &updateMetrics, nil
}
