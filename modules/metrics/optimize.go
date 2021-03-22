package metrics

import (
	"context"
)

type OptimizeMetrics struct {
	Invocations     Counter
	DurationGauge   Gauge
	LongUpdateCount Counter
}

func NewTestOptimizeMetrics() *OptimizeMetrics {
	m := new(OptimizeMetrics)
	m.DurationGauge = new(LocalGauge)
	m.Invocations = new(LocalCounter)
	m.LongUpdateCount = new(LocalCounter)
	return m
}

var EmptyOptimizeMetrics OptimizeMetrics = OptimizeMetrics{
	Invocations:     &EmptyCounter{},
	DurationGauge:   &EmptyGauge{},
	LongUpdateCount: &EmptyCounter{},
}

type CostMatrixMetrics struct {
	Invocations        Counter
	DurationGauge      Gauge
	LongUpdateCount    Counter
	Bytes              Gauge
	WriteResponseError Counter
}

func NewTestCostMatrixMetrics() *CostMatrixMetrics {
	return &CostMatrixMetrics{
		Invocations:        new(LocalCounter),
		DurationGauge:      new(LocalGauge),
		LongUpdateCount:    new(LocalCounter),
		Bytes:              new(LocalGauge),
		WriteResponseError: new(LocalCounter),
	}
}

var EmptyCostMatrixMetrics CostMatrixMetrics = CostMatrixMetrics{
	Invocations:        &EmptyCounter{},
	DurationGauge:      &EmptyGauge{},
	LongUpdateCount:    &EmptyCounter{},
	Bytes:              &EmptyGauge{},
	WriteResponseError: &EmptyCounter{},
}

type RouteMatrixMetrics struct {
	DatacenterCount Gauge
	RelayCount      Gauge
	RouteCount      Gauge
	Bytes           Gauge
}

func NewTestRouteMatrixMetrics() *RouteMatrixMetrics {
	return &RouteMatrixMetrics{
		DatacenterCount: new(LocalGauge),
		RelayCount:      new(LocalGauge),
		RouteCount:      new(LocalGauge),
		Bytes:           new(LocalGauge),
	}
}

var EmptyRouteMatrixMetrics RouteMatrixMetrics = RouteMatrixMetrics{
	DatacenterCount: &EmptyGauge{},
	RelayCount:      &EmptyGauge{},
	RouteCount:      &EmptyGauge{},
	Bytes:           &EmptyGauge{},
}

func NewCostMatrixMetrics(ctx context.Context, metricsHandler Handler) (*CostMatrixMetrics, error) {
	costMatrixDurationGauge, err := metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "StatsDB -> GetCostMatrix duration",
		ServiceName: "relay_backend",
		ID:          "cost_matrix.duration",
		Unit:        "milliseconds",
		Description: "How long it takes to generate a cost matrix from the stats database.",
	})
	if err != nil {
		return nil, err
	}

	costMatrixInvocationsCounter, err := metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Total StatsDB -> CostMatrix invocations",
		ServiceName: "relay_backend",
		ID:          "cost_matrix.count",
		Unit:        "invocations",
		Description: "The total number of StatsDB -> CostMatrix invocations",
	})
	if err != nil {
		return nil, err
	}

	costMatrixLongUpdateCounter, err := metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Cost Matrix Long Updates",
		ServiceName: "relay_backend",
		ID:          "cost_matrix.long.updates",
		Unit:        "updates",
		Description: "The number of cost matrix gen calls that took longer than 1 second",
	})
	if err != nil {
		return nil, err
	}

	costMatrixBytes, err := metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Cost Matrix Size",
		ServiceName: "relay_backend",
		ID:          "cost_matrix.bytes",
		Unit:        "bytes",
		Description: "How large the cost matrix is in bytes",
	})
	if err != nil {
		return nil, err
	}

	costMatrixMetrics := CostMatrixMetrics{
		Invocations:     costMatrixInvocationsCounter,
		DurationGauge:   costMatrixDurationGauge,
		LongUpdateCount: costMatrixLongUpdateCounter,
		Bytes:           costMatrixBytes,
	}

	return &costMatrixMetrics, nil
}

func NewOptimizeMetrics(ctx context.Context, metricsHandler Handler) (*OptimizeMetrics, error) {
	var err error
	m := new(OptimizeMetrics)
	m.DurationGauge, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Optimize duration",
		ServiceName: "relay_backend",
		ID:          "optimize.duration",
		Unit:        "milliseconds",
		Description: "How long it takes to optimize a cost matrix.",
	})
	if err != nil {
		return nil, err
	}

	m.Invocations, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Total cost matrix optimize invocations",
		ServiceName: "relay_backend",
		ID:          "optimize.count",
		Unit:        "invocations",
		Description: "The total number of cost matrix optimize calls",
	})
	if err != nil {
		return nil, err
	}

	m.LongUpdateCount, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Optimize Long Updates",
		ServiceName: "relay_backend",
		ID:          "optimize.long.updates",
		Unit:        "updates",
		Description: "The number of optimize calls that took longer than 1 second",
	})
	if err != nil {
		return nil, err
	}

	return m, nil
}

func NewValveCostMatrixMetrics(ctx context.Context, metricsHandler Handler) (*CostMatrixMetrics, error) {
	costMatrixDurationGauge, err := metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Valve StatsDB -> GetCostMatrix duration",
		ServiceName: "relay_backend",
		ID:          "cost_matrix.valve.duration",
		Unit:        "milliseconds",
		Description: "How long it takes to generate a valve cost matrix from the stats database.",
	})
	if err != nil {
		return nil, err
	}

	costMatrixInvocationsCounter, err := metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Valve Total StatsDB -> CostMatrix invocations",
		ServiceName: "relay_backend",
		ID:          "cost_matrix.valve.count",
		Unit:        "invocations",
		Description: "The total number of valve StatsDB -> CostMatrix invocations",
	})
	if err != nil {
		return nil, err
	}

	costMatrixLongUpdateCounter, err := metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Valve Cost Matrix Long Updates",
		ServiceName: "relay_backend",
		ID:          "cost_matrix.valve.long.updates",
		Unit:        "updates",
		Description: "The number of valve cost matrix gen calls that took longer than 1 second",
	})
	if err != nil {
		return nil, err
	}

	costMatrixBytes, err := metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Valve Cost Matrix Size",
		ServiceName: "relay_backend",
		ID:          "cost_matrix.valve.bytes",
		Unit:        "bytes",
		Description: "How large the valve cost matrix is in bytes",
	})
	if err != nil {
		return nil, err
	}

	costMatrixMetrics := CostMatrixMetrics{
		Invocations:     costMatrixInvocationsCounter,
		DurationGauge:   costMatrixDurationGauge,
		LongUpdateCount: costMatrixLongUpdateCounter,
		Bytes:           costMatrixBytes,
	}

	return &costMatrixMetrics, nil
}

func NewValveOptimizeMetrics(ctx context.Context, metricsHandler Handler) (*OptimizeMetrics, error) {
	optimizeDurationGauge, err := metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Valve optimize duration",
		ServiceName: "relay_backend",
		ID:          "optimize.valve.duration",
		Unit:        "milliseconds",
		Description: "How long it takes to optimize a valve cost matrix.",
	})
	if err != nil {
		return nil, err
	}

	optimizeInvocationsCounter, err := metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Valve total cost matrix optimize invocations",
		ServiceName: "relay_backend",
		ID:          "optimize.valve.count",
		Unit:        "invocations",
		Description: "The total number of valve cost matrix optimize calls",
	})
	if err != nil {
		return nil, err
	}

	optimizeMetrics := OptimizeMetrics{
		Invocations:   optimizeInvocationsCounter,
		DurationGauge: optimizeDurationGauge,
	}

	optimizeMetrics.LongUpdateCount, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Valve Optimize Long Updates",
		ServiceName: "relay_backend",
		ID:          "optimize.valve.long.updates",
		Unit:        "updates",
		Description: "The number of valve optimize calls that took longer than 1 second",
	})
	if err != nil {
		return nil, err
	}

	return &optimizeMetrics, nil
}

func NewValveRouteMatrixMetrics(ctx context.Context, metricsHandler Handler) (*RouteMatrixMetrics, error) {
	routeMatrixMetrics := RouteMatrixMetrics{}
	var err error

	routeMatrixMetrics.DatacenterCount, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Valve Route Matrix Datacenter Count",
		ServiceName: "relay_backend",
		ID:          "route_matrix.valve.datacenter.count",
		Unit:        "datacenters",
		Description: "The number of datacenters the valve route matrix contains",
	})
	if err != nil {
		return nil, err
	}

	routeMatrixMetrics.RelayCount, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Valve Route Matrix Relay Count",
		ServiceName: "relay_backend",
		ID:          "route_matrix.valve.relay.count",
		Unit:        "relays",
		Description: "The number of relays the valve route matrix contains",
	})
	if err != nil {
		return nil, err
	}

	routeMatrixMetrics.RouteCount, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Valve Route Matrix Route Count",
		ServiceName: "relay_backend",
		ID:          "route_matrix.valve.route.count",
		Unit:        "routes",
		Description: "The number of routes the valve route matrix contains",
	})
	if err != nil {
		return nil, err
	}

	routeMatrixMetrics.Bytes, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Valve Route Matrix Size",
		ServiceName: "relay_backend",
		ID:          "route_matrix.valve.bytes",
		Unit:        "bytes",
		Description: "How large the valve route matrix is in bytes",
	})
	if err != nil {
		return nil, err
	}

	return &routeMatrixMetrics, nil
}
