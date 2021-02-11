package metrics

import (
	"context"
)

// RelayInitMetrics defines the set of metrics for the relay init handler in the relay backend.
type RelayInitMetrics struct {
	HandlerMetrics RoutineMetrics

	UnmarshalFailure   Counter
	InvalidMagic       Counter
	InvalidVersion     Counter
	RelayNotFound      Counter
	RelayQuarantined   Counter
	DecryptionFailure  Counter
	RelayAlreadyExists Counter
}

// EmptyRelayInitMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyRelayInitMetrics RelayInitMetrics = RelayInitMetrics{
	HandlerMetrics: EmptyRoutineMetrics,

	UnmarshalFailure:   &EmptyCounter{},
	InvalidMagic:       &EmptyCounter{},
	InvalidVersion:     &EmptyCounter{},
	RelayNotFound:      &EmptyCounter{},
	RelayQuarantined:   &EmptyCounter{},
	DecryptionFailure:  &EmptyCounter{},
	RelayAlreadyExists: &EmptyCounter{},
}

// RelayUpdateMetrics defines the set of metrics for the relay update handler in the relay backend.
type RelayUpdateMetrics struct {
	HandlerMetrics RoutineMetrics

	UnmarshalFailure Counter
	InvalidVersion   Counter
	ExceedMaxRelays  Counter
	RelayNotFound    Counter
	InvalidToken     Counter
	RelayNotEnabled  Counter
}

// EmptyRelayUpdateMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyRelayUpdateMetrics RelayUpdateMetrics = RelayUpdateMetrics{
	HandlerMetrics: EmptyRoutineMetrics,

	UnmarshalFailure: &EmptyCounter{},
	InvalidVersion:   &EmptyCounter{},
	ExceedMaxRelays:  &EmptyCounter{},
	RelayNotFound:    &EmptyCounter{},
	InvalidToken:     &EmptyCounter{},
	RelayNotEnabled:  &EmptyCounter{},
}

// RelayBackendMetrics defines the set of metrics for the relay backend.
type RelayBackendMetrics struct {
	ServiceMetrics ServiceMetrics

	RelayInitMetrics   RelayInitMetrics
	RelayUpdateMetrics RelayUpdateMetrics

	OptimizeMetrics      RoutineMetrics
	ValveOptimizeMetrics RoutineMetrics

	CostMatrixMetrics      CostMatrixMetrics
	ValveCostMatrixMetrics CostMatrixMetrics

	RouteMatrixMetrics      RouteMatrixMetrics
	ValveRouteMatrixMetrics RouteMatrixMetrics

	PingStatsMetrics        PublisherMetrics
	RelayStatsMetrics       PublisherMetrics
	RouteMatrixStatsMetrics PublisherMetrics
}

// EmptyRelayBackendMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyRelayBackendMetrics RelayBackendMetrics = RelayBackendMetrics{
	ServiceMetrics: EmptyServiceMetrics,

	RelayInitMetrics:   EmptyRelayInitMetrics,
	RelayUpdateMetrics: EmptyRelayUpdateMetrics,

	OptimizeMetrics:      EmptyRoutineMetrics,
	ValveOptimizeMetrics: EmptyRoutineMetrics,

	CostMatrixMetrics:      EmptyCostMatrixMetrics,
	ValveCostMatrixMetrics: EmptyCostMatrixMetrics,

	RouteMatrixMetrics:      EmptyRouteMatrixMetrics,
	ValveRouteMatrixMetrics: EmptyRouteMatrixMetrics,

	PingStatsMetrics:        EmptyPublisherMetrics,
	RelayStatsMetrics:       EmptyPublisherMetrics,
	RouteMatrixStatsMetrics: EmptyPublisherMetrics,
}

// NewRelayBackendMetrics creates the metrics that the relay backend will use.
func NewRelayBackendMetrics(ctx context.Context, handler Handler) (RelayBackendMetrics, error) {
	serviceName := "relay_backend"

	var err error
	m := RelayBackendMetrics{}

	m.ServiceMetrics, err = NewServiceMetrics(ctx, handler, serviceName)
	if err != nil {
		return EmptyRelayBackendMetrics, err
	}

	m.RelayInitMetrics, err = newRelayInitMetrics(ctx, handler, serviceName, "relay_init", "Relay Init", "relay init request")
	if err != nil {
		return EmptyRelayBackendMetrics, err
	}

	m.RelayUpdateMetrics, err = newRelayUpdateMetrics(ctx, handler, serviceName, "relay_update", "Relay Update", "relay update request")
	if err != nil {
		return EmptyRelayBackendMetrics, err
	}

	m.OptimizeMetrics, err = NewRoutineMetrics(ctx, handler, serviceName, "optimize", "Optimize", "optimize call")
	if err != nil {
		return EmptyRelayBackendMetrics, err
	}

	m.ValveOptimizeMetrics, err = NewRoutineMetrics(ctx, handler, serviceName, "valve_optimize", "Valve Optimize", "valve optimize call")
	if err != nil {
		return EmptyRelayBackendMetrics, err
	}

	m.CostMatrixMetrics, err = NewCostMatrixMetrics(ctx, handler, serviceName, "cost_matrix", "Cost Matrix", "cost matrix")
	if err != nil {
		return EmptyRelayBackendMetrics, err
	}

	m.ValveCostMatrixMetrics, err = NewCostMatrixMetrics(ctx, handler, serviceName, "valve_cost_matrix", "Valve Cost Matrix", "valve cost matrix")
	if err != nil {
		return EmptyRelayBackendMetrics, err
	}

	m.RouteMatrixMetrics, err = NewRouteMatrixMetrics(ctx, handler, serviceName, "route_matrix", "Route Matrix", "route matrix")
	if err != nil {
		return EmptyRelayBackendMetrics, err
	}

	m.ValveRouteMatrixMetrics, err = NewRouteMatrixMetrics(ctx, handler, serviceName, "valve_route_matrix", "Valve Route Matrix", "valve route matrix")
	if err != nil {
		return EmptyRelayBackendMetrics, err
	}

	m.PingStatsMetrics, err = NewPublisherMetrics(ctx, handler, serviceName, "ping_stats", "Ping Stats", "ping stats")
	if err != nil {
		return EmptyRelayBackendMetrics, err
	}

	m.RelayStatsMetrics, err = NewPublisherMetrics(ctx, handler, serviceName, "relay_stats", "Relay Stats", "relay stats")
	if err != nil {
		return EmptyRelayBackendMetrics, err
	}

	m.RouteMatrixStatsMetrics, err = NewPublisherMetrics(ctx, handler, serviceName, "route_matrix_stats", "Route Matrix Stats", "route matrix stats")
	if err != nil {
		return EmptyRelayBackendMetrics, err
	}

	return m, nil
}
