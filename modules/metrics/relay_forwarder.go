package metrics

import "context"

type RelayForwarderMetrics struct {
	ForwarderServiceMetrics *ServiceMetrics
	ForwardPostSuccess      Counter
	ErrorMetrics            RelayForwarderErrorMetrics
}

var EmptyRelayForwarderMetrics = &RelayForwarderMetrics{
	ForwarderServiceMetrics: &EmptyServiceMetrics,
	ForwardPostSuccess:     &EmptyCounter{},
	ErrorMetrics:           *EmptyRelayForwarderErrorMetrics,
}

type RelayForwarderErrorMetrics struct {
	ForwardPostError Counter
}

var EmptyRelayForwarderErrorMetrics = &RelayForwarderErrorMetrics{
	ForwardPostError: &EmptyCounter{},
}

func NewRelayForwarderMetrics(ctx context.Context, metricsHandler Handler, serviceName string) (*RelayForwarderMetrics, error) {
	m := new(RelayForwarderMetrics)

	var err error

	m.ForwarderServiceMetrics, err = NewServiceMetrics(ctx, metricsHandler, serviceName)
	if err != nil {
		return nil, err
	}

	m.ForwardPostSuccess, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Forward Post Success Count",
		ServiceName: serviceName,
		ID:          serviceName + ".forward_post_success",
		Unit:        "invocations",
		Description: "The Number of times the " + serviceName + " successfully made a post request to the relay_gateway",
	})
	if err != nil {
		return nil, err
    }

	m.ErrorMetrics.ForwardPostError, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Forward Post Error Count",
		ServiceName: serviceName,
		ID:          serviceName + ".forward_post_error",
		Unit:        "errors",
		Description: "The Number of times the " + serviceName + " failed to make a post request to the relay_gateway",
	})
	if err != nil {
		return nil, err
	}

	return m, nil
}
