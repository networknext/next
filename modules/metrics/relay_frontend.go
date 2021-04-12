package metrics

import "context"

type RelayFrontendMetrics struct {
	MasterSelect      Counter
	MasterSelectError Counter
	CostMatrix        FrontendMatrix
	RouteMatrix       FrontendMatrix
	ValveMatrix       FrontendMatrix
}

var EmptyRelayFrontendMetrics = &RelayFrontendMetrics{
	MasterSelect:      &EmptyCounter{},
	MasterSelectError: &EmptyCounter{},
	CostMatrix:        *EmptyFrontendMatrix,
	RouteMatrix:       *EmptyFrontendMatrix,
	ValveMatrix:       *EmptyFrontendMatrix,
}

type FrontendMatrix struct {
	Invocations Counter
	Error       Counter
}

var EmptyFrontendMatrix = &FrontendMatrix{
	Invocations: &EmptyCounter{},
	Error:       &EmptyCounter{},
}

func NewRelayFrontendMetrics(ctx context.Context, metricsHandler Handler) (*RelayFrontendMetrics, error) {
	m := new(RelayFrontendMetrics)

	var err error

	m.MasterSelect, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Master Select Invocations Count",
		ServiceName: "relay_Frontend",
		ID:          "frontend.master_select.count",
		Unit:        "Invocations",
		Description: "The Number of Master Select Invocations",
	})
	if err != nil {
		return nil, err
	}

	m.MasterSelectError, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Master Select Error Count",
		ServiceName: "relay_Frontend",
		ID:          "frontend.master_select.error",
		Unit:        "error",
		Description: "The Number of Master Select Errors",
	})
	if err != nil {
		return nil, err
	}

	m.CostMatrix.Invocations, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Cost Matrix Invocations Count",
		ServiceName: "relay_Frontend",
		ID:          "frontend.cost_matrix.count",
		Unit:        "Invocations",
		Description: "The Number of cache Cost Matrix calls",
	})
	if err != nil {
		return nil, err
	}

	m.CostMatrix.Error, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Cost Matrix Error Count",
		ServiceName: "relay_Frontend",
		ID:          "frontend.cost_matrix.error",
		Unit:        "error",
		Description: "The Number of cache Cost Matrix Errors",
	})
	if err != nil {
		return nil, err
	}

	m.RouteMatrix.Invocations, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Route Matrix Invocations Count",
		ServiceName: "relay_Frontend",
		ID:          "frontend.route_matrix.count",
		Unit:        "Invocations",
		Description: "The Number of cache Route Matrix calls",
	})
	if err != nil {
		return nil, err
	}

	m.RouteMatrix.Error, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Route Matrix Error Count",
		ServiceName: "relay_Frontend",
		ID:          "frontend.route_matrix.error",
		Unit:        "error",
		Description: "The Number of cache Route Matrix Errors",
	})
	if err != nil {
		return nil, err
	}

	m.ValveMatrix.Invocations, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Valve Matrix Invocations Count",
		ServiceName: "relay_Frontend",
		ID:          "frontend.valve_matrix.count",
		Unit:        "Invocations",
		Description: "The Number of cache Valve Matrix calls",
	})
	if err != nil {
		return nil, err
	}

	m.ValveMatrix.Error, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Valve Matrix Error Count",
		ServiceName: "relay_Frontend",
		ID:          "frontend.valve_matrix.error",
		Unit:        "error",
		Description: "The Number of cache Valve Matrix Errors",
	})
	if err != nil {
		return nil, err
	}

	return m, nil
}
