package metrics

import (
	"context"
	"fmt"
)

// OptimizerRelayInitMetrics defines the set of metrics that the optimizer needs when initializing a relay.
type OptimizerRelayInitMetrics struct {
	RelayNotFound      Counter
	RelayAlreadyExists Counter
}

// EmptyOptimizerRelayInitMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyOptimizerRelayInitMetrics = OptimizerRelayInitMetrics{
	RelayNotFound:      &EmptyCounter{},
	RelayAlreadyExists: &EmptyCounter{},
}

// OptimizerRelayUpdateMetrics defines the set of metrics that the optimizer needs when updating a relay.
type OptimizerRelayUpdateMetrics struct {
	UnmarshalFailure Counter
	RelayNotFound    Counter
}

// EmptyOptimizerRelayUpdateMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyOptimizerRelayUpdateMetrics = OptimizerRelayUpdateMetrics{
	UnmarshalFailure: &EmptyCounter{},
	RelayNotFound:    &EmptyCounter{},
}

// OptimizerMetrics defines the set of metrics for the optimizer.
type OptimizerMetrics struct {
	ServiceMetrics ServiceMetrics

	OptimizerRelayInitMetrics   OptimizerRelayInitMetrics
	OptimizerRelayUpdateMetrics OptimizerRelayUpdateMetrics

	CostMatrixMetrics  CostMatrixMetrics
	RouteMatrixMetrics RouteMatrixMetrics

	OptimizeMetrics RoutineMetrics

	ValveCostMatrixMetrics  CostMatrixMetrics
	ValveRouteMatrixMetrics RouteMatrixMetrics

	ValveOptimizeMetrics RoutineMetrics

	PingStatsMetrics  PublisherMetrics
	RelayStatsMetrics PublisherMetrics
}

// EmptyOptimizerMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyOptimizerMetrics = OptimizerMetrics{
	CostMatrixMetrics:  EmptyCostMatrixMetrics,
	RouteMatrixMetrics: EmptyRouteMatrixMetrics,

	OptimizeMetrics: EmptyRoutineMetrics,

	ValveCostMatrixMetrics:  EmptyCostMatrixMetrics,
	ValveRouteMatrixMetrics: EmptyRouteMatrixMetrics,

	ValveOptimizeMetrics: EmptyRoutineMetrics,
}

// NewOptimizerMetrics creates the metrics that the optimizer will use.
func NewOptimizerMetrics(ctx context.Context, handler Handler) (OptimizerMetrics, error) {
	serviceName := "optimizer"

	var err error
	m := OptimizerMetrics{}

	m.ServiceMetrics, err = NewServiceMetrics(ctx, handler, serviceName)
	if err != nil {
		return EmptyOptimizerMetrics, fmt.Errorf("failed to create service metrics: %v", err)
	}

	m.OptimizerRelayInitMetrics, err = newOptimizerRelayInitMetrics(ctx, handler, serviceName, "relay_init", "Relay Init", "relay init entry")
	if err != nil {
		return EmptyOptimizerMetrics, fmt.Errorf("failed to create optimizer relay init metrics: %v", err)
	}

	m.OptimizerRelayUpdateMetrics, err = newOptimizerRelayUpdateMetrics(ctx, handler, serviceName, "relay_update", "Relay Update", "relay update message")
	if err != nil {
		return EmptyOptimizerMetrics, fmt.Errorf("failed to create optimizer relay update metrics: %v", err)
	}

	m.CostMatrixMetrics, err = NewCostMatrixMetrics(ctx, handler, serviceName, "cost_matrix", "Cost Matrix", "cost matrix")
	if err != nil {
		return EmptyOptimizerMetrics, fmt.Errorf("failed to create cost matrix metrics: %v", err)
	}

	m.OptimizeMetrics, err = NewRoutineMetrics(ctx, handler, serviceName, "optimize", "Optimize", "optimize call")
	if err != nil {
		return EmptyOptimizerMetrics, fmt.Errorf("failed to create optimize metrics: %v", err)
	}

	m.RouteMatrixMetrics, err = NewRouteMatrixMetrics(ctx, handler, serviceName, "route_matrix", "Route Matrix", "route matrix")
	if err != nil {
		return EmptyOptimizerMetrics, fmt.Errorf("failed to create route matrix metrics: %v", err)
	}

	m.ValveCostMatrixMetrics, err = NewCostMatrixMetrics(ctx, handler, serviceName, "valve_cost_matrix", "Valve Cost Matrix", "valve cost matrix")
	if err != nil {
		return EmptyOptimizerMetrics, fmt.Errorf("failed to create valve cost matrix metrics: %v", err)
	}

	m.ValveOptimizeMetrics, err = NewRoutineMetrics(ctx, handler, serviceName, "valve_optimize", "Valve Optimize", "valve optimize call")
	if err != nil {
		return EmptyOptimizerMetrics, fmt.Errorf("failed to create valve optimize metrics: %v", err)
	}

	m.ValveRouteMatrixMetrics, err = NewRouteMatrixMetrics(ctx, handler, serviceName, "valve_route_matrix", "Valve Route Matrix", "valve route matrix")
	if err != nil {
		return EmptyOptimizerMetrics, fmt.Errorf("failed to create valve route matrix metrics: %v", err)
	}

	m.PingStatsMetrics, err = NewPublisherMetrics(ctx, handler, serviceName, "ping_stats", "Ping Stats", "ping stats")
	if err != nil {
		return EmptyOptimizerMetrics, fmt.Errorf("failed to create ping stats metrics: %v", err)
	}

	m.RelayStatsMetrics, err = NewPublisherMetrics(ctx, handler, serviceName, "relay_stats", "Relay Stats", "relay stats")
	if err != nil {
		return EmptyOptimizerMetrics, fmt.Errorf("failed to create relay stats metrics: %v", err)
	}

	return m, nil
}

func newOptimizerRelayInitMetrics(ctx context.Context, handler Handler, serviceName string, handlerID string, handlerName string, handlerDescription string) (OptimizerRelayInitMetrics, error) {
	var err error
	m := OptimizerRelayInitMetrics{}

	m.RelayNotFound, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Relay Not Found",
		ServiceName: serviceName,
		ID:          handlerID + ".relay_not_found",
		Unit:        "errors",
		Description: "The number of times a " + handlerDescription + " contains an unknown relay.",
	})
	if err != nil {
		return EmptyOptimizerRelayInitMetrics, err
	}

	m.RelayAlreadyExists, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Relay Already Exists",
		ServiceName: serviceName,
		ID:          handlerID + ".relay_already_exists",
		Unit:        "errors",
		Description: "The number of times a " + handlerDescription + " contains a relay that is already initialized.",
	})
	if err != nil {
		return EmptyOptimizerRelayInitMetrics, err
	}

	return m, nil
}

func newOptimizerRelayUpdateMetrics(ctx context.Context, handler Handler, serviceName string, handlerID string, handlerName string, handlerDescription string) (OptimizerRelayUpdateMetrics, error) {
	var err error
	m := OptimizerRelayUpdateMetrics{}

	m.UnmarshalFailure, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Unmarshal Failure",
		ServiceName: serviceName,
		ID:          handlerID + ".unmarshal_failure",
		Unit:        "errors",
		Description: "The number of times a " + handlerDescription + " could not be unmarshaled.",
	})
	if err != nil {
		return EmptyOptimizerRelayUpdateMetrics, err
	}

	m.RelayNotFound, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Relay Not Found",
		ServiceName: serviceName,
		ID:          handlerID + ".relay_not_found",
		Unit:        "errors",
		Description: "The number of times a " + handlerDescription + " contains an unknown relay.",
	})
	if err != nil {
		return EmptyOptimizerRelayUpdateMetrics, err
	}

	return m, nil
}
