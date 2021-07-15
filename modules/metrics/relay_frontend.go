package metrics

import "context"

type RelayFrontendMetrics struct {
	FrontendServiceMetrics *ServiceMetrics
	MasterSelect           Counter
	MasterSelectSuccess    Counter
	CostMatrix             FrontendMatrix
	RouteMatrix            FrontendMatrix
	PingStatsMetrics       AnalyticsMetrics
	RelayStatsMetrics      AnalyticsMetrics
	ErrorMetrics           RelayFrontendErrorMetrics
}

var EmptyRelayFrontendMetrics = &RelayFrontendMetrics{
	FrontendServiceMetrics: &EmptyServiceMetrics,
	MasterSelect:           &EmptyCounter{},
	MasterSelectSuccess:    &EmptyCounter{},
	CostMatrix:             *EmptyFrontendMatrix,
	RouteMatrix:            *EmptyFrontendMatrix,
	PingStatsMetrics:       EmptyAnalyticsMetrics,
	RelayStatsMetrics:      EmptyAnalyticsMetrics,
	ErrorMetrics:           *EmptyRelayFrontendErrorMetrics,
}

type RelayFrontendErrorMetrics struct {
	MasterSelectError Counter
}

var EmptyRelayFrontendErrorMetrics = &RelayFrontendErrorMetrics{
	MasterSelectError: &EmptyCounter{},
}

type FrontendMatrix struct {
	Invocations Counter
	Success     Counter
	Error       Counter
}

var EmptyFrontendMatrix = &FrontendMatrix{
	Invocations: &EmptyCounter{},
	Success:     &EmptyCounter{},
	Error:       &EmptyCounter{},
}

func NewRelayFrontendMetrics(ctx context.Context, metricsHandler Handler) (*RelayFrontendMetrics, error) {
	m := new(RelayFrontendMetrics)

	var err error

	m.FrontendServiceMetrics, err = NewServiceMetrics(ctx, metricsHandler, "relay_frontend")
	if err != nil {
		return nil, err
	}

	m.MasterSelect, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Master Select Invocations",
		ServiceName: "relay_frontend",
		ID:          "relay_frontend.master_select.count",
		Unit:        "invocations",
		Description: "The Number of times the relay_frontend was asked to chose the master relay backend.",
	})
	if err != nil {
		return nil, err
	}

	m.MasterSelectSuccess, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Master Select Success",
		ServiceName: "relay_frontend",
		ID:          "relay_frontend.master_select.success",
		Unit:        "invocations",
		Description: "The Number of times the relay_frontend successfully chose the master relay backend.",
	})
	if err != nil {
		return nil, err
	}

	m.ErrorMetrics.MasterSelectError, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Master Select Failure Count",
		ServiceName: "relay_frontend",
		ID:          "relay_frontend.master_select.error",
		Unit:        "errors",
		Description: "The Number of times the relay_frontend failed to choose the master relay backend",
	})
	if err != nil {
		return nil, err
	}

	m.CostMatrix.Invocations, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Cost Matrix Invocations",
		ServiceName: "relay_frontend",
		ID:          "relay_frontend.cost_matrix.count",
		Unit:        "invocations",
		Description: "The Number of times the relay_frontend was asked to get the cached cost matrix",
	})
	if err != nil {
		return nil, err
	}

	m.CostMatrix.Success, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Cost Matrix Cache Success",
		ServiceName: "relay_frontend",
		ID:          "relay_frontend.cost_matrix.success_count",
		Unit:        "invocations",
		Description: "The Number of times the relay_frontend successfully got the cached cost matrix",
	})
	if err != nil {
		return nil, err
	}

	m.CostMatrix.Error, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Cost Matrix Cache Failure",
		ServiceName: "relay_frontend",
		ID:          "relay_frontend.cost_matrix.cache_failure",
		Unit:        "errors",
		Description: "The Number of times the relay_frontend failed to get the cached cost matrix",
	})
	if err != nil {
		return nil, err
	}

	m.RouteMatrix.Invocations, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Route Matrix Invocations",
		ServiceName: "relay_frontend",
		ID:          "relay_frontend.route_matrix.count",
		Unit:        "invocations",
		Description: "The Number of times the relay_frontend was asked to get the cached route matrix",
	})
	if err != nil {
		return nil, err
	}

	m.RouteMatrix.Success, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Route Matrix Cache Success",
		ServiceName: "relay_frontend",
		ID:          "relay_frontend.route_matrix.success_count",
		Unit:        "invocations",
		Description: "The Number of times the relay_frontend successfully got the cached route matrix",
	})
	if err != nil {
		return nil, err
	}

	m.RouteMatrix.Error, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Route Matrix Cache Failure",
		ServiceName: "relay_frontend",
		ID:          "relay_frontend.route_matrix.cache_failure",
		Unit:        "errors",
		Description: "The Number of times the relay_frontend failed to get the cached route matrix",
	})
	if err != nil {
		return nil, err
	}

	m.PingStatsMetrics.EntriesReceived = &EmptyCounter{}

	m.PingStatsMetrics.EntriesSubmitted, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Relay Backend Ping Stats Entries Written",
		ServiceName: "relay_backend",
		ID:          "relay_backend.ping_stats.entries.submitted",
		Unit:        "entries",
		Description: "The number of ping stats entries the relay backend submitted to be published",
	})
	if err != nil {
		return nil, err
	}

	m.PingStatsMetrics.EntriesQueued, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Relay Backend Ping Stats Entries Queued",
		ServiceName: "relay_backend",
		ID:          "relay_backend.ping_stats.entries.queued",
		Unit:        "entries",
		Description: "The number of ping stats entries the relay backend has queued. This should always be 0",
	})
	if err != nil {
		return nil, err
	}

	m.PingStatsMetrics.EntriesFlushed, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Relay Backend Ping Stats Entries Flushed",
		ServiceName: "relay_backend",
		ID:          "relay_backend.ping_stats.entries.flushed",
		Unit:        "entries",
		Description: "The number of ping stats entries the relay backend has flushed",
	})
	if err != nil {
		return nil, err
	}

	m.PingStatsMetrics.ErrorMetrics.PublishFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Relay Backend Ping Stats Publish Failure",
		ServiceName: "relay_backend",
		ID:          "relay_backend.ping_stats.error.publish_failure",
		Unit:        "entries",
		Description: "The number of ping stats entries the relay backend has failed to publish",
	})
	if err != nil {
		return nil, err
	}

	m.PingStatsMetrics.ErrorMetrics.ReadFailure = &EmptyCounter{}

	m.PingStatsMetrics.ErrorMetrics.WriteFailure = &EmptyCounter{}

	m.RelayStatsMetrics.EntriesReceived = &EmptyCounter{}

	m.RelayStatsMetrics.EntriesSubmitted, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Relay Backend Relay Stats Entries Written",
		ServiceName: "relay_backend",
		ID:          "relay_backend.relay_stats.entries.submitted",
		Unit:        "entries",
		Description: "The number of relay stats entries the relay backend submitted to be published",
	})
	if err != nil {
		return nil, err
	}

	m.RelayStatsMetrics.EntriesQueued, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Relay Backend Relay Stats Entries Queued",
		ServiceName: "relay_backend",
		ID:          "relay_backend.relay_stats.entries.queued",
		Unit:        "entries",
		Description: "The number of relay stats entries the relay backend has queued. This should always be 0",
	})
	if err != nil {
		return nil, err
	}

	m.RelayStatsMetrics.EntriesFlushed, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Relay Backend Relay Stats Entries Flushed",
		ServiceName: "relay_backend",
		ID:          "relay_backend.relay_stats.entries.flushed",
		Unit:        "entries",
		Description: "The number of relay stats entries the relay backend has flushed",
	})
	if err != nil {
		return nil, err
	}

	m.RelayStatsMetrics.ErrorMetrics.PublishFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Relay Backend Relay Stats Publish Failure",
		ServiceName: "relay_backend",
		ID:          "relay_backend.relay_stats.error.publish_failure",
		Unit:        "entries",
		Description: "The number of relay stats entries the relay backend has failed to publish",
	})
	if err != nil {
		return nil, err
	}

	m.RelayStatsMetrics.ErrorMetrics.ReadFailure = &EmptyCounter{}

	m.RelayStatsMetrics.ErrorMetrics.WriteFailure = &EmptyCounter{}

	return m, nil
}
