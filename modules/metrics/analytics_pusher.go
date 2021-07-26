package metrics

import "context"

type AnalyticsPusherMetrics struct {
	AnalyticsPusherServiceMetrics *ServiceMetrics
	RouteMatrixInvocations        Counter
	RouteMatrixSuccess            Counter
	RouteMatrixUpdateDuration     Gauge
	RouteMatrixUpdateLongDuration Counter
	PingStatsMetrics              AnalyticsMetrics
	RelayStatsMetrics             AnalyticsMetrics
	ErrorMetrics                  AnalyticsPusherErrorMetrics
}

var EmptyAnalyticsPusherMetrics = &AnalyticsPusherMetrics{
	AnalyticsPusherServiceMetrics: &EmptyServiceMetrics,
	RouteMatrixInvocations:        &EmptyCounter{},
	RouteMatrixSuccess:            &EmptyCounter{},
	RouteMatrixUpdateDuration:     &EmptyGauge{},
	RouteMatrixUpdateLongDuration: &EmptyCounter{},
	PingStatsMetrics:              EmptyAnalyticsMetrics,
	RelayStatsMetrics:             EmptyAnalyticsMetrics,
	ErrorMetrics:                  EmptyAnalyticsPusherErrorMetrics,
}

type AnalyticsPusherErrorMetrics struct {
	RouteMatrixReaderNil        Counter
	RouteMatrixReadFailure      Counter
	RouteMatrixBufferEmpty      Counter
	RouteMatrixSerializeFailure Counter
	StaleRouteMatrix            Counter
}

var EmptyAnalyticsPusherErrorMetrics = AnalyticsPusherErrorMetrics{
	RouteMatrixReaderNil:        &EmptyCounter{},
	RouteMatrixReadFailure:      &EmptyCounter{},
	RouteMatrixBufferEmpty:      &EmptyCounter{},
	RouteMatrixSerializeFailure: &EmptyCounter{},
	StaleRouteMatrix:            &EmptyCounter{},
}

func NewAnalyticsPusherMetrics(ctx context.Context, metricsHandler Handler, serviceName string) (*AnalyticsPusherMetrics, error) {
	m := new(AnalyticsPusherMetrics)

	var err error

	m.AnalyticsPusherServiceMetrics, err = NewServiceMetrics(ctx, metricsHandler, serviceName)
	if err != nil {
		return nil, err
	}

	m.RouteMatrixInvocations, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Route Matrix Invocations",
		ServiceName: serviceName,
		ID:          serviceName + ".route_matrix.invocations",
		Unit:        "invocations",
		Description: "The number of times the " + serviceName + " has been requested to get the route matrix.",
	})
	if err != nil {
		return nil, err
	}

	m.RouteMatrixSuccess, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Route Matrix Success",
		ServiceName: serviceName,
		ID:          serviceName + ".route_matrix.success",
		Unit:        "invocations",
		Description: "The number of times the " + serviceName + " has been successfully received and serialized the route matrix.",
	})
	if err != nil {
		return nil, err
	}

	m.RouteMatrixUpdateDuration, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Route Matrix Update Duration",
		ServiceName: serviceName,
		ID:          serviceName + ".route_matrix.update_duration",
		Unit:        "ms",
		Description: "The number of ms the " + serviceName + " took to receive and serialize route matrix.",
	})
	if err != nil {
		return nil, err
	}

	m.RouteMatrixUpdateLongDuration, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Route Matrix Update Long Duration",
		ServiceName: serviceName,
		ID:          serviceName + ".route_matrix.update_long_duration",
		Unit:        "invocations",
		Description: "The number of times the " + serviceName + " took more than 250 ms to receive and serialize the route matrix.",
	})
	if err != nil {
		return nil, err
	}

	m.PingStatsMetrics.EntriesReceived, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Ping Stats Entries Received",
		ServiceName: serviceName,
		ID:          serviceName + ".ping_stats.entries.received",
		Unit:        "entries",
		Description: "The number of relay stats entries the " + serviceName + " received from the route matrix.",
	})
	if err != nil {
		return nil, err
	}

	m.PingStatsMetrics.EntriesSubmitted, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Ping Stats Entries Submitted",
		ServiceName: serviceName,
		ID:          serviceName + ".ping_stats.entries.submitted",
		Unit:        "entries",
		Description: "The number of relay stats entries the " + serviceName + " submitted to be published to Google Pub/Sub.",
	})
	if err != nil {
		return nil, err
	}

	m.PingStatsMetrics.EntriesQueued = &EmptyCounter{}

	m.PingStatsMetrics.EntriesFlushed, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Ping Stats Entries Flushed",
		ServiceName: serviceName,
		ID:          serviceName + ".ping_stats.entries.flushed",
		Unit:        "entries",
		Description: "The number of relay stats entries the " + serviceName + " flushed to Google Pub/Sub.",
	})
	if err != nil {
		return nil, err
	}

	m.PingStatsMetrics.ErrorMetrics.PublishFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Ping Stats Publish Failure",
		ServiceName: serviceName,
		ID:          serviceName + ".ping_stats.error.publish_failure",
		Unit:        "errors",
		Description: "The number of relay stats entries the " + serviceName + " has failed to publish to Google Pub/Sub",
	})
	if err != nil {
		return nil, err
	}

	m.PingStatsMetrics.ErrorMetrics.ReadFailure = &EmptyCounter{}

	m.PingStatsMetrics.ErrorMetrics.WriteFailure = &EmptyCounter{}

	m.RelayStatsMetrics.EntriesReceived, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Relay Stats Entries Received",
		ServiceName: serviceName,
		ID:          serviceName + ".relay_stats.entries.received",
		Unit:        "entries",
		Description: "The number of relay stats entries the " + serviceName + " received from the route matrix.",
	})
	if err != nil {
		return nil, err
	}

	m.RelayStatsMetrics.EntriesSubmitted, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Relay Stats Entries Submitted",
		ServiceName: serviceName,
		ID:          serviceName + ".relay_stats.entries.submitted",
		Unit:        "entries",
		Description: "The number of relay stats entries the " + serviceName + " submitted to be published to Google Pub/Sub.",
	})
	if err != nil {
		return nil, err
	}

	m.RelayStatsMetrics.EntriesQueued = &EmptyCounter{}

	m.RelayStatsMetrics.EntriesFlushed, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Relay Stats Entries Flushed",
		ServiceName: serviceName,
		ID:          serviceName + ".relay_stats.entries.flushed",
		Unit:        "entries",
		Description: "The number of relay stats entries the " + serviceName + " flushed to Google Pub/Sub.",
	})
	if err != nil {
		return nil, err
	}

	m.RelayStatsMetrics.ErrorMetrics.PublishFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Relay Stats Publish Failure",
		ServiceName: serviceName,
		ID:          serviceName + ".relay_stats.error.publish_failure",
		Unit:        "errors",
		Description: "The number of relay stats entries the " + serviceName + " has failed to publish to Google Pub/Sub",
	})
	if err != nil {
		return nil, err
	}

	m.RelayStatsMetrics.ErrorMetrics.ReadFailure = &EmptyCounter{}

	m.RelayStatsMetrics.ErrorMetrics.WriteFailure = &EmptyCounter{}

	m.ErrorMetrics.RouteMatrixReaderNil, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Route Matrix Reader Nil",
		ServiceName: serviceName,
		ID:          serviceName + ".route_matrix.error.reader_nil",
		Unit:        "errors",
		Description: "The number of times the route matrix reader was nil when getting the route matrix.",
	})
	if err != nil {
		return nil, err
	}

	m.ErrorMetrics.RouteMatrixReadFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Route Matrix Read Failure",
		ServiceName: serviceName,
		ID:          serviceName + ".route_matrix.error.read_failure",
		Unit:        "errors",
		Description: "The number of times the " + serviceName + " failed to read the route matrix data.",
	})
	if err != nil {
		return nil, err
	}

	m.ErrorMetrics.RouteMatrixBufferEmpty, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Route Matrix Buffer Empty",
		ServiceName: serviceName,
		ID:          serviceName + ".route_matrix.error.buffer_empty",
		Unit:        "errors",
		Description: "The number of times the route matrix buffer was empty after reading the route matrix data.",
	})
	if err != nil {
		return nil, err
	}

	m.ErrorMetrics.RouteMatrixSerializeFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Route Matrix Serialize Failure",
		ServiceName: serviceName,
		ID:          serviceName + ".route_matrix.error.serialize_failure",
		Unit:        "errors",
		Description: "The number of times the " + serviceName + " failed to serialize the route matrix.",
	})
	if err != nil {
		return nil, err
	}

	m.ErrorMetrics.StaleRouteMatrix, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Route Matrix Stale",
		ServiceName: serviceName,
		ID:          serviceName + ".route_matrix.error.stale_matrix",
		Unit:        "errors",
		Description: "The number of times the " + serviceName + " serialized a stale route matrix.",
	})
	if err != nil {
		return nil, err
	}

	return m, nil
}
