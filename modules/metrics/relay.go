package metrics

import (
	"context"
)

type RelayInitMetrics struct {
	Invocations   Counter
	DurationGauge Gauge
	ErrorMetrics  RelayInitErrorMetrics
}

var EmptyRelayInitMetrics RelayInitMetrics = RelayInitMetrics{
	Invocations:   &EmptyCounter{},
	DurationGauge: &EmptyGauge{},
	ErrorMetrics:  EmptyRelayInitErrorMetrics,
}

type RelayInitErrorMetrics struct {
	UnmarshalFailure   Counter
	InvalidMagic       Counter
	InvalidVersion     Counter
	RelayNotFound      Counter
	RelayQuarantined   Counter
	DecryptionFailure  Counter
	RelayAlreadyExists Counter
	IPLookupFailure    Counter
}

var EmptyRelayInitErrorMetrics RelayInitErrorMetrics = RelayInitErrorMetrics{
	UnmarshalFailure:   &EmptyCounter{},
	InvalidMagic:       &EmptyCounter{},
	InvalidVersion:     &EmptyCounter{},
	RelayNotFound:      &EmptyCounter{},
	RelayQuarantined:   &EmptyCounter{},
	DecryptionFailure:  &EmptyCounter{},
	RelayAlreadyExists: &EmptyCounter{},
	IPLookupFailure:    &EmptyCounter{},
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
	UnmarshalFailure Counter
	InvalidVersion   Counter
	ExceedMaxRelays  Counter
	RelayNotFound    Counter
	InvalidToken     Counter
	RelayNotEnabled  Counter
}

var EmptyRelayUpdateErrorMetrics RelayUpdateErrorMetrics = RelayUpdateErrorMetrics{
	UnmarshalFailure: &EmptyCounter{},
	InvalidVersion:   &EmptyCounter{},
	ExceedMaxRelays:  &EmptyCounter{},
	RelayNotFound:    &EmptyCounter{},
	InvalidToken:     &EmptyCounter{},
	RelayNotEnabled:  &EmptyCounter{},
}

type RelayBackendMetrics struct {
	Goroutines        Gauge
	MemoryAllocated   Gauge
	RouteMatrix       RouteMatrixMetrics
	PingStatsMetrics  AnalyticsMetrics
	RelayStatsMetrics AnalyticsMetrics
}

var EmptyRelayBackendMetrics RelayBackendMetrics = RelayBackendMetrics{
	Goroutines:        &EmptyGauge{},
	MemoryAllocated:   &EmptyGauge{},
	RouteMatrix:       EmptyRouteMatrixMetrics,
	PingStatsMetrics:  EmptyAnalyticsMetrics,
	RelayStatsMetrics: EmptyAnalyticsMetrics,
}

func NewRelayInitMetrics(ctx context.Context, metricsHandler Handler) (*RelayInitMetrics, error) {
	initCount, err := metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Total relay init count",
		ServiceName: "relay_backend",
		ID:          "relay.init.count",
		Unit:        "requests",
		Description: "The total number of received relay init requests",
	})
	if err != nil {
		return nil, err
	}

	initDuration, err := metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Relay init duration",
		ServiceName: "relay_backend",
		ID:          "relay.init.duration",
		Unit:        "milliseconds",
		Description: "How long it takes to process a relay init request",
	})
	if err != nil {
		return nil, err
	}

	initMetrics := RelayInitMetrics{
		Invocations:   initCount,
		DurationGauge: initDuration,
		ErrorMetrics:  EmptyRelayInitErrorMetrics,
	}

	return &initMetrics, nil
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

	updateMetrics := RelayUpdateMetrics{
		Invocations:   updateCount,
		DurationGauge: updateDuration,
		ErrorMetrics:  EmptyRelayUpdateErrorMetrics,
	}

	return &updateMetrics, nil
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

	relayBackendMetrics.PingStatsMetrics.EntriesReceived = &EmptyCounter{}

	relayBackendMetrics.PingStatsMetrics.EntriesSubmitted, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Relay Backend Ping Stats Entries Written",
		ServiceName: "relay_backend",
		ID:          "relay_backend.ping_stats.entries.submitted",
		Unit:        "entries",
		Description: "The number of ping stats entries the relay backend submitted to be published",
	})
	if err != nil {
		return nil, err
	}

	relayBackendMetrics.PingStatsMetrics.EntriesQueued, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Relay Backend Ping Stats Entries Queued",
		ServiceName: "relay_backend",
		ID:          "relay_backend.ping_stats.entries.queued",
		Unit:        "entries",
		Description: "The number of ping stats entries the relay backend has queued. This should always be 0",
	})
	if err != nil {
		return nil, err
	}

	relayBackendMetrics.PingStatsMetrics.EntriesFlushed, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Relay Backend Ping Stats Entries Flushed",
		ServiceName: "relay_backend",
		ID:          "relay_backend.ping_stats.entries.flushed",
		Unit:        "entries",
		Description: "The number of ping stats entries the relay backend has flushed",
	})
	if err != nil {
		return nil, err
	}

	relayBackendMetrics.PingStatsMetrics.ErrorMetrics.PublishFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Relay Backend Ping Stats Publish Failure",
		ServiceName: "relay_backend",
		ID:          "relay_backend.ping_stats.error.publish_failure",
		Unit:        "entries",
		Description: "The number of ping stats entries the relay backend has failed to publish",
	})
	if err != nil {
		return nil, err
	}

	relayBackendMetrics.PingStatsMetrics.ErrorMetrics.ReadFailure = &EmptyCounter{}

	relayBackendMetrics.PingStatsMetrics.ErrorMetrics.WriteFailure = &EmptyCounter{}

	relayBackendMetrics.RelayStatsMetrics.EntriesReceived = &EmptyCounter{}

	relayBackendMetrics.RelayStatsMetrics.EntriesSubmitted, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Relay Backend Relay Stats Entries Written",
		ServiceName: "relay_backend",
		ID:          "relay_backend.relay_stats.entries.submitted",
		Unit:        "entries",
		Description: "The number of relay stats entries the relay backend submitted to be published",
	})
	if err != nil {
		return nil, err
	}

	relayBackendMetrics.RelayStatsMetrics.EntriesQueued, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Relay Backend Relay Stats Entries Queued",
		ServiceName: "relay_backend",
		ID:          "relay_backend.relay_stats.entries.queued",
		Unit:        "entries",
		Description: "The number of relay stats entries the relay backend has queued. This should always be 0",
	})
	if err != nil {
		return nil, err
	}

	relayBackendMetrics.RelayStatsMetrics.EntriesFlushed, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Relay Backend Relay Stats Entries Flushed",
		ServiceName: "relay_backend",
		ID:          "relay_backend.relay_stats.entries.flushed",
		Unit:        "entries",
		Description: "The number of relay stats entries the relay backend has flushed",
	})
	if err != nil {
		return nil, err
	}

	relayBackendMetrics.RelayStatsMetrics.ErrorMetrics.PublishFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Relay Backend Relay Stats Publish Failure",
		ServiceName: "relay_backend",
		ID:          "relay_backend.relay_stats.error.publish_failure",
		Unit:        "entries",
		Description: "The number of relay stats entries the relay backend has failed to publish",
	})
	if err != nil {
		return nil, err
	}

	relayBackendMetrics.RelayStatsMetrics.ErrorMetrics.ReadFailure = &EmptyCounter{}

	relayBackendMetrics.RelayStatsMetrics.ErrorMetrics.WriteFailure = &EmptyCounter{}

	return &relayBackendMetrics, nil
}