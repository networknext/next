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

type RelayBackendMetrics struct {
	Goroutines              Gauge
	MemoryAllocated         Gauge
	RouteMatrix             RouteMatrixMetrics
	RouteMatrixStatsMetrics AnalyticsMetrics
}

var EmptyRelayBackendMetrics RelayBackendMetrics = RelayBackendMetrics{
	Goroutines:      &EmptyGauge{},
	MemoryAllocated: &EmptyGauge{},
	RouteMatrix:     EmptyRouteMatrixMetrics,
}

type RelayUpdateMetrics struct {
	Invocations      Counter
	DurationGauge    Gauge
	InitErrorMetrics RelayInitErrorMetrics
	ErrorMetrics     RelayUpdateErrorMetrics
}

var EmptyRelayUpdateMetrics RelayUpdateMetrics = RelayUpdateMetrics{
	Invocations:      &EmptyCounter{},
	DurationGauge:    &EmptyGauge{},
	InitErrorMetrics: EmptyRelayInitErrorMetrics,
	ErrorMetrics:     EmptyRelayUpdateErrorMetrics,
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

	//RelayNamesHash
	relayBackendMetrics.RouteMatrixStatsMetrics.EntriesReceived, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Relay Backend Route Matrix Stats Entries Received",
		ServiceName: "relay_backend",
		ID:          "relay_backend.route_matrix_stats.entries",
		Unit:        "entries",
		Description: "The total number of Route Matrix Stats entries received through Google Pub/Sub",
	})
	if err != nil {
		return nil, err
	}

	relayBackendMetrics.RouteMatrixStatsMetrics.EntriesSubmitted, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Relay Backend Route Matrix Stats Entries Submitted",
		ServiceName: "relay_backend",
		ID:          "relay_backend.route_matrix_stats.entries.submitted",
		Unit:        "entries",
		Description: "The total number of relay stats entries submitted to BigQuery",
	})
	if err != nil {
		return nil, err
	}

	relayBackendMetrics.RouteMatrixStatsMetrics.EntriesQueued, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Relay Backend Route Matrix Stats Entries Queued",
		ServiceName: "relay_backend",
		ID:          "relay_backend.route_matrix_stats.entries.queued",
		Unit:        "entries",
		Description: "The total number of relay stats entries waiting to be sent to BigQuery",
	})
	if err != nil {
		return nil, err
	}

	relayBackendMetrics.RouteMatrixStatsMetrics.EntriesFlushed, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Relay Backend Route Matrix Stats Entries Flushed",
		ServiceName: "relay_backend",
		ID:          "relay_backend.route_matrix_stats.entries.flushed",
		Unit:        "entries",
		Description: "The total number of relay stats entries written to BigQuery",
	})
	if err != nil {
		return nil, err
	}

	relayBackendMetrics.RouteMatrixStatsMetrics.ErrorMetrics.PublishFailure = &EmptyCounter{}

	relayBackendMetrics.RouteMatrixStatsMetrics.ErrorMetrics.ReadFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Relay Backend Route Matrix Stats Read Failure",
		ServiceName: "relay_backend",
		ID:          "relay_backend.route_matrix_stats.error.read_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	relayBackendMetrics.RouteMatrixStatsMetrics.ErrorMetrics.WriteFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Relay Backend Route Matrix Stats Write Failure",
		ServiceName: "relay_backend",
		ID:          "relay_backend.route_matrix_stats.error.write_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	return &relayBackendMetrics, nil
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

	var initErrorMetrics RelayInitErrorMetrics
	initErrorMetrics.UnmarshalFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Total relay init unmarshal failure count",
		ServiceName: "relay_backend",
		ID:          "relay.init.errors.unmarshal_failure.count",
		Unit:        "unmarshal_failure",
		Description: "The total number of received relay init requests that resulted in unmarshal failure",
	})
	if err != nil {
		return nil, err
	}

	initErrorMetrics.InvalidMagic, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Total relay init invalid magic error count",
		ServiceName: "relay_backend",
		ID:          "relay.init.errors.invalid_magic.count",
		Unit:        "invalid_magic",
		Description: "The total number of received relay init requests that resulted in invalid magic error",
	})
	if err != nil {
		return nil, err
	}

	initErrorMetrics.InvalidVersion, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Total relay init invalid version error count",
		ServiceName: "relay_backend",
		ID:          "relay.init.errors.invalid_version.count",
		Unit:        "invalid_version",
		Description: "The total number of received relay init requests that resulted in invalid version error",
	})
	if err != nil {
		return nil, err
	}

	initErrorMetrics.RelayNotFound, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Total relay init relay not found error count",
		ServiceName: "relay_backend",
		ID:          "relay.init.errors.not_found.count",
		Unit:        "relay_not_found",
		Description: "The total number of received relay init requests that resulted in relay not found error",
	})
	if err != nil {
		return nil, err
	}

	initErrorMetrics.RelayQuarantined, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Total relay init relay quarantined error count",
		ServiceName: "relay_backend",
		ID:          "relay.init.errors.quarantined.count",
		Unit:        "relay_quarantined",
		Description: "The total number of received relay init requests that resulted in relay quarantined error",
	})
	if err != nil {
		return nil, err
	}

	initErrorMetrics.DecryptionFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Total relay init decryption failure count",
		ServiceName: "relay_backend",
		ID:          "relay.init.errors.decryption_failure.count",
		Unit:        "decryption_failure",
		Description: "The total number of received relay init requests that resulted in decryption failure",
	})
	if err != nil {
		return nil, err
	}

	initErrorMetrics.RelayAlreadyExists, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Total relay init relay already exists count",
		ServiceName: "relay_backend",
		ID:          "relay.init.errors.already_exists.count",
		Unit:        "relay_already_exists",
		Description: "The total number of received relay init requests that resulted in relay already exists",
	})
	if err != nil {
		return nil, err
	}

	initErrorMetrics.IPLookupFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Total relay init IP lookup failure count",
		ServiceName: "relay_backend",
		ID:          "relay.init.errors.ip_lookup_failure.count",
		Unit:        "ip_lookup_failure",
		Description: "The total number of received relay init requests that resulted in IP lookup failure",
	})
	if err != nil {
		return nil, err
	}

	initMetrics := RelayInitMetrics{
		Invocations:   initCount,
		DurationGauge: initDuration,
		ErrorMetrics:  initErrorMetrics,
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

	var em RelayUpdateErrorMetrics
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

	em.InvalidVersion, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Total relay update invalid version error count",
		ServiceName: "relay_backend",
		ID:          "relay.update.errors.invalid_version.count",
		Unit:        "invalid_version",
		Description: "The total number of received relay update requests that resulted in invalid version error",
	})
	if err != nil {
		return nil, err
	}

	em.ExceedMaxRelays, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Total relay upgrade exceed max relays error count",
		ServiceName: "relay_backend",
		ID:          "relay.update.errors.exceed_max_relays.count",
		Unit:        "exceed_max_relays",
		Description: "The total number of received relay update requests that resulted in exceed max relays error",
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

	em.InvalidToken, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Total relay update invalid token error count",
		ServiceName: "relay_backend",
		ID:          "relay.update.errors.invalid_token.count",
		Unit:        "invalid_token",
		Description: "The total number of received relay init requests that resulted in invalid token error",
	})
	if err != nil {
		return nil, err
	}

	em.RelayNotEnabled, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Total relay update relay not enabled error count",
		ServiceName: "relay_backend",
		ID:          "relay.init_errors.not_enabled.count",
		Unit:        "relay_not_enabled",
		Description: "The total number of received relay init requests that resulted in relay not enabled",
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
