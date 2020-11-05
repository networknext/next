package main

import (
	"context"

	"github.com/networknext/backend/modules/metrics"
)
type Metrics struct{
	costMatrixMetrics *metrics.CostMatrixMetrics
	optimizeMetrics *metrics.OptimizeMetrics
	relayBackendMetrics *metrics.RelayBackendMetrics
	valveCostMatrixMetrics *metrics.CostMatrixMetrics
	valveOptimizeMetrics    *metrics.OptimizeMetrics
	valveRouteMatrixMetrics *metrics.RouteMatrixMetrics
}

func NewMetrics(ctx context.Context, metricsHandler metrics.Handler) (*Metrics, error, string){
	costMatrixMetrics, err := metrics.NewCostMatrixMetrics(ctx, metricsHandler)
	if err != nil {
		return nil, err, "failed to create cost matrix metrics"
	}

	optimizeMetrics, err := metrics.NewOptimizeMetrics(ctx, metricsHandler)
	if err != nil {
		return nil, err, "failed to create optimize metrics"
	}

	relayBackendMetrics, err := metrics.NewRelayBackendMetrics(ctx, metricsHandler)
	if err != nil {
		return nil, err, "failed to create relay backend metrics"
	}

	valveCostMatrixMetrics, err := metrics.NewValveCostMatrixMetrics(ctx, metricsHandler)
	if err != nil {
		return nil, err, "failed to create valve cost matrix metrics"
	}

	valveOptimizeMetrics, err := metrics.NewValveOptimizeMetrics(ctx, metricsHandler)
	if err != nil {
		return nil, err, "failed to create valve optimize metrics"
	}

	valveRouteMatrixMetrics, err := metrics.NewValveRouteMatrixMetrics(ctx, metricsHandler)
	if err != nil {
		return nil, err, "failed to create valve route matrix metrics"
	}

	return &Metrics{
		costMatrixMetrics: costMatrixMetrics,
		optimizeMetrics: optimizeMetrics,
		relayBackendMetrics: relayBackendMetrics,
		valveCostMatrixMetrics: valveCostMatrixMetrics,
		valveOptimizeMetrics: valveOptimizeMetrics,
		valveRouteMatrixMetrics: valveRouteMatrixMetrics,
	},nil,""
}
