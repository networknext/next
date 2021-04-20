package metrics

import "context"

type RelayFrontendMetrics struct {
	FrontendServiceMetrics *ServiceMetrics
	MasterSelect           Counter
	MasterSelectSuccess    Counter
	CostMatrix             FrontendMatrix
	RouteMatrix            FrontendMatrix
	ErrorMetrics           RelayFrontendErrorMetrics
}

var EmptyRelayFrontendMetrics = &RelayFrontendMetrics{
	FrontendServiceMetrics: &EmptyServiceMetrics,
	MasterSelect:           &EmptyCounter{},
	MasterSelectSuccess:    &EmptyCounter{},
	CostMatrix:             *EmptyFrontendMatrix,
	RouteMatrix:            *EmptyFrontendMatrix,
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

	return m, nil
}
