package main

import (
	"context"

	"github.com/networknext/backend/modules/metrics"
)
type Metrics struct{
	costMatrixMetrics *metrics.CostMatrixMetrics
	optimizeMetrics *metrics.OptimizeMetrics
	relayBackendMetrics *metrics.RelayBackendMetrics
	RelayInitMetrics *metrics.RelayInitMetrics
	RelayUpdateMetrics *metrics.RelayUpdateMetrics
	valveCostMatrixMetrics *metrics.CostMatrixMetrics
	valveOptimizeMetrics    *metrics.OptimizeMetrics
	valveRouteMatrixMetrics *metrics.RouteMatrixMetrics
}

func NewMetrics(ctx context.Context, metricsHandler metrics.Handler) (*Metrics, error, string){
	m := &Metrics{}
	costMatrixMetrics, err := metrics.NewCostMatrixMetrics(ctx, metricsHandler)
	if err != nil {
		return nil, err, "failed to create cost matrix metrics"
	}
	m.costMatrixMetrics = costMatrixMetrics

	optimizeMetrics, err := metrics.NewOptimizeMetrics(ctx, metricsHandler)
	if err != nil {
		return nil, err, "failed to create optimize metrics"
	}
	m.optimizeMetrics = optimizeMetrics

	relayBackendMetrics, err := metrics.NewRelayBackendMetrics(ctx, metricsHandler)
	if err != nil {
		return nil, err, "failed to create relay backend metrics"
	}
	m.relayBackendMetrics = relayBackendMetrics

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
	m.valveCostMatrixMetrics = valveCostMatrixMetrics

	valveOptimizeMetrics, err := metrics.NewValveOptimizeMetrics(ctx, metricsHandler)
	if err != nil {
		return nil, err, "failed to create valve optimize metrics"
	}
	m.valveOptimizeMetrics = valveOptimizeMetrics

	valveRouteMatrixMetrics, err := metrics.NewValveRouteMatrixMetrics(ctx, metricsHandler)
	if err != nil {
		return nil, err, "failed to create valve route matrix metrics"
	}
	m.valveRouteMatrixMetrics = valveRouteMatrixMetrics

	return m,nil,""
}
