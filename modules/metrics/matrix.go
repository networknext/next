package metrics

import "context"

// CostMatrixMetrics defines the metrics that measure the data stored in the cost matrix.
type CostMatrixMetrics struct {
	DatacenterCount Gauge
	RelayCount      Gauge
	Bytes           Gauge
}

// EmptyCostMatrixMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyCostMatrixMetrics CostMatrixMetrics = CostMatrixMetrics{
	DatacenterCount: &EmptyGauge{},
	RelayCount:      &EmptyGauge{},
	Bytes:           &EmptyGauge{},
}

// RouteMatrixMetrics defines the metrics that measure the data stored in the route matrix.
type RouteMatrixMetrics struct {
	DatacenterCount Gauge
	RelayCount      Gauge
	RouteCount      Gauge
	Bytes           Gauge
}

// EmptyRouteMatrixMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyRouteMatrixMetrics RouteMatrixMetrics = RouteMatrixMetrics{
	DatacenterCount: &EmptyGauge{},
	RelayCount:      &EmptyGauge{},
	RouteCount:      &EmptyGauge{},
	Bytes:           &EmptyGauge{},
}

// NewCostMatrixMetrics creates the metrics for measuring the data stored in the cost matrix.
func NewCostMatrixMetrics(ctx context.Context, handler Handler, serviceName string, matrixID string, matrixName string, matrixDescription string) (CostMatrixMetrics, error) {
	var err error
	m := CostMatrixMetrics{}

	m.DatacenterCount, err = handler.NewGauge(ctx, &Descriptor{
		DisplayName: matrixName + " Datacenter Count",
		ServiceName: serviceName,
		ID:          matrixID + ".datacenters",
		Unit:        "datacenters",
		Description: "The number of datacenters the " + matrixDescription + " contains.",
	})
	if err != nil {
		return EmptyCostMatrixMetrics, err
	}

	m.RelayCount, err = handler.NewGauge(ctx, &Descriptor{
		DisplayName: matrixName + " Relay Count",
		ServiceName: serviceName,
		ID:          matrixID + ".relays",
		Unit:        "relays",
		Description: "The number of relays the " + matrixDescription + " contains.",
	})
	if err != nil {
		return EmptyCostMatrixMetrics, err
	}

	m.Bytes, err = handler.NewGauge(ctx, &Descriptor{
		DisplayName: matrixName + " Size",
		ServiceName: serviceName,
		ID:          matrixID + ".bytes",
		Unit:        "bytes",
		Description: "How large the " + matrixDescription + " is in bytes.",
	})
	if err != nil {
		return EmptyCostMatrixMetrics, err
	}

	return m, nil
}

// NewRouteMatrixMetrics creates the metrics for measuring the data stored in the route matrix.
func NewRouteMatrixMetrics(ctx context.Context, handler Handler, serviceName string, matrixID string, matrixName string, matrixDescription string) (RouteMatrixMetrics, error) {
	var err error
	m := RouteMatrixMetrics{}

	m.DatacenterCount, err = handler.NewGauge(ctx, &Descriptor{
		DisplayName: matrixName + " Datacenter Count",
		ServiceName: serviceName,
		ID:          matrixID + ".datacenters",
		Unit:        "datacenters",
		Description: "The number of datacenters the " + matrixDescription + " contains.",
	})
	if err != nil {
		return EmptyRouteMatrixMetrics, err
	}

	m.RelayCount, err = handler.NewGauge(ctx, &Descriptor{
		DisplayName: matrixName + " Relay Count",
		ServiceName: serviceName,
		ID:          matrixID + ".relays",
		Unit:        "relays",
		Description: "The number of relays the " + matrixDescription + " contains.",
	})
	if err != nil {
		return EmptyRouteMatrixMetrics, err
	}

	m.RouteCount, err = handler.NewGauge(ctx, &Descriptor{
		DisplayName: matrixName + " Route Count",
		ServiceName: serviceName,
		ID:          matrixID + ".routes",
		Unit:        "routes",
		Description: "The number of routes the " + matrixDescription + " contains.",
	})
	if err != nil {
		return EmptyRouteMatrixMetrics, err
	}

	m.Bytes, err = handler.NewGauge(ctx, &Descriptor{
		DisplayName: matrixName + " Size",
		ServiceName: serviceName,
		ID:          matrixID + ".bytes",
		Unit:        "bytes",
		Description: "How large the " + matrixDescription + " is in bytes.",
	})
	if err != nil {
		return EmptyRouteMatrixMetrics, err
	}

	return m, nil
}
