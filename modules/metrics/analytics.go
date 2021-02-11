package metrics

import "context"

// AnalyticsMetrics defines the set of metrics for the analytics service.
type AnalyticsMetrics struct {
	ServiceMetrics ServiceMetrics

	PingStatsSubscriberMetrics        SubscriberMetrics
	RelayStatsSubscriberMetrics       SubscriberMetrics
	RouteMatrixStatsSubscriberMetrics SubscriberMetrics

	PingStatsPublisherMetrics        PublisherMetrics
	RelayStatsPublisherMetrics       PublisherMetrics
	RouteMatrixStatsPublisherMetrics PublisherMetrics
}

// EmptyAnalyticsMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyAnalyticsMetrics = AnalyticsMetrics{
	ServiceMetrics: EmptyServiceMetrics,

	PingStatsSubscriberMetrics:        EmptySubscriberMetrics,
	RelayStatsSubscriberMetrics:       EmptySubscriberMetrics,
	RouteMatrixStatsSubscriberMetrics: EmptySubscriberMetrics,

	PingStatsPublisherMetrics:        EmptyPublisherMetrics,
	RelayStatsPublisherMetrics:       EmptyPublisherMetrics,
	RouteMatrixStatsPublisherMetrics: EmptyPublisherMetrics,
}

// NewAnalyticsMetrics creates the metrics that the analytics service will use.
func NewAnalyticsMetrics(ctx context.Context, handler Handler) (AnalyticsMetrics, error) {
	serviceName := "analytics"

	var err error
	m := AnalyticsMetrics{}

	m.ServiceMetrics, err = NewServiceMetrics(ctx, handler, serviceName)
	if err != nil {
		return EmptyAnalyticsMetrics, err
	}

	m.PingStatsPublisherMetrics, err = NewPublisherMetrics(ctx, handler, serviceName, "ping_stats", "Ping Stats", "ping stats")
	if err != nil {
		return EmptyAnalyticsMetrics, err
	}

	m.RelayStatsPublisherMetrics, err = NewPublisherMetrics(ctx, handler, serviceName, "relay_stats", "Relay Stats", "relay stats")
	if err != nil {
		return EmptyAnalyticsMetrics, err
	}

	m.RouteMatrixStatsPublisherMetrics, err = NewPublisherMetrics(ctx, handler, serviceName, "route_matrix_stats", "Route Matrix Stats", "route matrix stats")
	if err != nil {
		return EmptyAnalyticsMetrics, err
	}

	m.PingStatsSubscriberMetrics, err = NewSubscriberMetrics(ctx, handler, serviceName, "ping_stats", "Ping Stats", "ping stats")
	if err != nil {
		return EmptyAnalyticsMetrics, err
	}

	m.RelayStatsSubscriberMetrics, err = NewSubscriberMetrics(ctx, handler, serviceName, "relay_stats", "Relay Stats", "relay stats")
	if err != nil {
		return EmptyAnalyticsMetrics, err
	}

	m.RouteMatrixStatsSubscriberMetrics, err = NewSubscriberMetrics(ctx, handler, serviceName, "route_matrix_stats", "Route Matrix Stats", "route matrix stats")
	if err != nil {
		return EmptyAnalyticsMetrics, err
	}

	return m, nil
}
