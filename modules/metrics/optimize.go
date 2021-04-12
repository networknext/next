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

//func NewTestingCostMatrixMetrics
