package optimizer

import (
	"context"

	"github.com/networknext/backend/modules/metrics"
)
type Metrics struct{
	CostMatrixMetrics *metrics.CostMatrixMetrics
	OptimizeMetrics *metrics.OptimizeMetrics
	RelayBackendMetrics *metrics.RelayBackendMetrics
	RelayInitMetrics *metrics.RelayInitMetrics
	RelayUpdateMetrics *metrics.RelayUpdateMetrics
	ValveCostMatrixMetrics *metrics.CostMatrixMetrics
	ValveOptimizeMetrics    *metrics.OptimizeMetrics
	ValveRouteMatrixMetrics *metrics.RouteMatrixMetrics
}

func NewMetrics(ctx context.Context, metricsHandler metrics.Handler) (*Metrics, error, string){
	m := &Metrics{}
	costMatrixMetrics, err := metrics.NewCostMatrixMetrics(ctx, metricsHandler)
	if err != nil {
		return nil, err, "failed to create cost matrix metrics"
	}
	m.CostMatrixMetrics = costMatrixMetrics

	optimizeMetrics, err := metrics.NewOptimizeMetrics(ctx, metricsHandler)
	if err != nil {
		return nil, err, "failed to create optimize metrics"
	}
	m.OptimizeMetrics = optimizeMetrics

	relayBackendMetrics, err := metrics.NewRelayBackendMetrics(ctx, metricsHandler)
	if err != nil {
		return nil, err, "failed to create relay backend metrics"
	}
	m.RelayBackendMetrics = relayBackendMetrics

	relayInitMetrics, err := metrics.NewRelayInitMetrics(ctx, metricsHandler)
	if err != nil {
		return nil, err, "failed to create relay init metrics"
	}
	m.RelayInitMetrics = relayInitMetrics

	relayUpdateMetrics, err := metrics.NewRelayUpdateMetrics(ctx, metricsHandler)
	if err != nil {
		return nil, err, "failed to create relay update metrics"
	}
	m.RelayUpdateMetrics = relayUpdateMetrics

	valveCostMatrixMetrics, err := metrics.NewValveCostMatrixMetrics(ctx, metricsHandler)
	if err != nil {
		return nil, err, "failed to create valve cost matrix metrics"
	}
	m.ValveCostMatrixMetrics = valveCostMatrixMetrics

	valveOptimizeMetrics, err := metrics.NewValveOptimizeMetrics(ctx, metricsHandler)
	if err != nil {
		return nil, err, "failed to create valve optimize metrics"
	}
	m.ValveOptimizeMetrics = valveOptimizeMetrics

	valveRouteMatrixMetrics, err := metrics.NewValveRouteMatrixMetrics(ctx, metricsHandler)
	if err != nil {
		return nil, err, "failed to create valve route matrix metrics"
	}
	m.ValveRouteMatrixMetrics = valveRouteMatrixMetrics

	return m,nil,""
}
