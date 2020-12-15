package metrics

import (
	"context"
)

type RelayGatewayMetrics struct {
	RelayInitMetrics   *RelayInitMetrics
	RelayUpdateMetrics *RelayUpdateMetrics
}

var EmptyRelayGatewayMetrics = &RelayGatewayMetrics{
	RelayInitMetrics:   &EmptyRelayInitMetrics,
	RelayUpdateMetrics: &EmptyRelayUpdateMetrics,
}

func NewRelayGatewayMetrics(ctx context.Context, metricsHandler Handler) (*RelayGatewayMetrics, error, string) {
	m := new(RelayGatewayMetrics)

	// Create relay init metrics
	relayInitMetrics, err := NewRelayGatewayInitMetrics(ctx, metricsHandler)
	if err != nil {
		return nil, err, "failed to create relay init metrics"
	}
	m.RelayInitMetrics = relayInitMetrics

	// Create relay update metrics
	relayUpdateMetrics, err := NewRelayGatewayUpdateMetrics(ctx, metricsHandler)
	if err != nil {
		return nil, err, "failed to create relay update metrics"
	}
	m.RelayUpdateMetrics = relayUpdateMetrics

	return m, nil, ""
}

func NewRelayGatewayInitMetrics(ctx context.Context, metricsHandler Handler) (*RelayInitMetrics, error) {
	initCount, err := metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Total relay init count",
		ServiceName: "relay_gateway",
		ID:          "relay_gateway.init.count",
		Unit:        "requests",
		Description: "The total number of received relay init requests",
	})
	if err != nil {
		return nil, err
	}

	initDuration, err := metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Relay init duration",
		ServiceName: "relay_gateway",
		ID:          "relay_gateway.init.duration",
		Unit:        "milliseconds",
		Description: "How long it takes to process a relay init request",
	})
	if err != nil {
		return nil, err
	}

	//todo: add error metrics when approved for relay backend
	initMetrics := RelayInitMetrics{
		Invocations:   initCount,
		DurationGauge: initDuration,
		ErrorMetrics:  EmptyRelayInitErrorMetrics,
	}

	return &initMetrics, nil
}

func NewRelayGatewayUpdateMetrics(ctx context.Context, metricsHandler Handler) (*RelayUpdateMetrics, error) {
	updateCount, err := metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Total relay update count",
		ServiceName: "relay_gateway",
		ID:          "relay_gateway.update.count",
		Unit:        "requests",
		Description: "The total number of received relay update requests",
	})
	if err != nil {
		return nil, err
	}

	updateDuration, err := metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Relay update duration",
		ServiceName: "relay_gateway",
		ID:          "relay_gateway.update.duration",
		Unit:        "milliseconds",
		Description: "How long it takes to process a relay update request.",
	})
	if err != nil {
		return nil, err
	}

	//todo: add error metrics when approved for relay backend
	updateMetrics := RelayUpdateMetrics{
		Invocations:   updateCount,
		DurationGauge: updateDuration,
		ErrorMetrics:  EmptyRelayUpdateErrorMetrics,
	}

	return &updateMetrics, nil
}
