package metrics

import "context"

type BuyerUserSessionsEndpointMetrics struct {
	NoSlicesRedisFailure    Counter
	NoSlicesBigTableFailure Counter
}

var EmptyBuyerUserSessionsEndpointMetrics = BuyerUserSessionsEndpointMetrics{
	NoSlicesRedisFailure:    &EmptyCounter{},
	NoSlicesBigTableFailure: &EmptyCounter{},
}

type PortalMetrics struct {
	ServiceMetrics ServiceMetrics

	BuyerUserSessionsMetrics BuyerUserSessionsEndpointMetrics

	BigTableMetrics BigTableReadMetrics
}

var EmptyPortalMetrics = PortalMetrics{
	ServiceMetrics: EmptyServiceMetrics,

	BuyerUserSessionsMetrics: EmptyBuyerUserSessionsEndpointMetrics,

	BigTableMetrics: EmptyBigTableReadMetrics,
}

func NewPortalMetrics(ctx context.Context, handler Handler) (PortalMetrics, error) {
	serviceName := "portal"

	var err error
	m := PortalMetrics{}

	m.BuyerUserSessionsMetrics, err = newBuyerUserSessionsMetrics(ctx, handler, serviceName, "user_sessions", "User Sessions", "sessions")
	if err != nil {
		return EmptyPortalMetrics, err
	}

	m.BigTableMetrics, err = NewBigTableReadMetrics(ctx, handler, serviceName)
	if err != nil {
		return EmptyPortalMetrics, err
	}

	return m, nil
}

func newBuyerUserSessionsMetrics(ctx context.Context, metricsHandler Handler, serviceName string, endpointID string, endpointName string, endpointDescription string) (BuyerUserSessionsEndpointMetrics, error) {
	var err error
	m := BuyerUserSessionsEndpointMetrics{}

	m.NoSlicesRedisFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: endpointName + " No Slices Redis Failure",
		ServiceName: serviceName,
		ID:          endpointID + ".no_slices_redis_failure",
		Unit:        "sessions",
		Description: "The total number of " + endpointDescription + " without slices in Redis.",
	})
	if err != nil {
		return EmptyBuyerUserSessionsEndpointMetrics, err
	}

	m.NoSlicesBigTableFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: endpointName + " No Slices BigTable Failure",
		ServiceName: serviceName,
		ID:          endpointID + ".no_slices_bigtable_failure",
		Unit:        "sessions",
		Description: "The total number of " + endpointDescription + " without slices in BigTable.",
	})
	if err != nil {
		return EmptyBuyerUserSessionsEndpointMetrics, err
	}

	return m, nil
}
