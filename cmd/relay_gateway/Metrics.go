package main

import (
	"context"
	"github.com/networknext/backend/modules/metrics"
)

type Metrics struct{
	RelayInitMetrics *metrics.RelayInitMetrics
	RelayUpdateMetrics *metrics.RelayUpdateMetrics
	RelayBackendMetrics *metrics.RelayBackendMetrics
}

func NewMetrics(ctx context.Context, metricsHandler metrics.Handler) (*Metrics, error ,string){
	m := new(Metrics)

	// Create relay init metrics
	relayInitMetrics, err := metrics.NewRelayInitMetrics(ctx, metricsHandler)
	if err != nil {
		return nil, err, "failed to create relay init metrics"
	}
	m.RelayInitMetrics = relayInitMetrics

	// Create relay update metrics
	relayUpdateMetrics, err := metrics.NewRelayUpdateMetrics(ctx, metricsHandler)
	if err != nil {
		return nil, err, "failed to create relay update metrics"
	}
	m.RelayUpdateMetrics = relayUpdateMetrics

	relayBackendMetrics, err := metrics.NewRelayBackendMetrics(ctx, metricsHandler)
	if err != nil {
		return nil,err, "failed to create relay backend metrics"
	}
	m.RelayBackendMetrics = relayBackendMetrics

	return m, nil,""
}