package metrics

import "context"

// RelayPusherServiceMetrics defines a set of metrics for the beacon insertion service.
type RelayPusherServiceMetrics struct {
	ServiceMetrics     *ServiceMetrics
	RelayPusherMetrics *RelayPusherMetrics
}

// EmptyRelayPusherServiceMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyRelayPusherServiceMetrics RelayPusherServiceMetrics = RelayPusherServiceMetrics{
	ServiceMetrics:     &EmptyServiceMetrics,
	RelayPusherMetrics: &EmptyRelayPusherMetrics,
}

// RelayPusherMetrics defines a set of metrics for monitoring the beacon insertion service.
type RelayPusherMetrics struct {
	ErrorMetrics RelayPusherErrorMetrics
}

// EmptyRelayPusherMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyRelayPusherMetrics RelayPusherMetrics = RelayPusherMetrics{
	ErrorMetrics: EmptyRelayPusherErrorMetrics,
}

// RelayPusherErrorMetrics defines a set of metrics for recording errors for the beacon insertion service.
type RelayPusherErrorMetrics struct {
}

// EmptyRelayPusherErrorMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyRelayPusherErrorMetrics RelayPusherErrorMetrics = RelayPusherErrorMetrics{}

// NewRelayPusherServiceMetrics creates the metrics that the beacon insertion service will use.
func NewRelayPusherServiceMetrics(ctx context.Context, metricsHandler Handler) (*RelayPusherServiceMetrics, error) {
	RelayPusherServiceMetrics := &RelayPusherServiceMetrics{}
	var err error

	RelayPusherServiceMetrics.ServiceMetrics, err = NewServiceMetrics(ctx, metricsHandler, "relay_pusher")
	if err != nil {
		return nil, err
	}

	RelayPusherServiceMetrics.RelayPusherMetrics = &RelayPusherMetrics{}
	RelayPusherServiceMetrics.RelayPusherMetrics.ErrorMetrics = RelayPusherErrorMetrics{}

	return RelayPusherServiceMetrics, nil
}
