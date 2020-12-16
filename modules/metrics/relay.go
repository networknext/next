package metrics

import (
	"context"
)

type RelayInitMetrics struct {
	Invocations   Counter
	DurationGauge Gauge
	ErrorMetrics  RelayInitErrorMetrics
}

var EmptyRelayInitMetrics RelayInitMetrics = RelayInitMetrics{
	Invocations:   &EmptyCounter{},
	DurationGauge: &EmptyGauge{},
	ErrorMetrics:  EmptyRelayInitErrorMetrics,
}

type RelayInitErrorMetrics struct {
	UnmarshalFailure   Counter
	InvalidMagic       Counter
	InvalidVersion     Counter
	RelayNotFound      Counter
	RelayQuarantined   Counter
	DecryptionFailure  Counter
	RelayAlreadyExists Counter
	IPLookupFailure    Counter
}

var EmptyRelayInitErrorMetrics RelayInitErrorMetrics = RelayInitErrorMetrics{
	UnmarshalFailure:   &EmptyCounter{},
	InvalidMagic:       &EmptyCounter{},
	InvalidVersion:     &EmptyCounter{},
	RelayNotFound:      &EmptyCounter{},
	RelayQuarantined:   &EmptyCounter{},
	DecryptionFailure:  &EmptyCounter{},
	RelayAlreadyExists: &EmptyCounter{},
	IPLookupFailure:    &EmptyCounter{},
}

type RelayUpdateMetrics struct {
	Invocations   Counter
	DurationGauge Gauge
	ErrorMetrics  RelayUpdateErrorMetrics
}

var EmptyRelayUpdateMetrics RelayUpdateMetrics = RelayUpdateMetrics{
	Invocations:   &EmptyCounter{},
	DurationGauge: &EmptyGauge{},
	ErrorMetrics:  EmptyRelayUpdateErrorMetrics,
}

type RelayUpdateErrorMetrics struct {
	UnmarshalFailure Counter
	InvalidVersion   Counter
	ExceedMaxRelays  Counter
	RelayNotFound    Counter
	InvalidToken     Counter
	RelayNotEnabled  Counter
}

var EmptyRelayUpdateErrorMetrics RelayUpdateErrorMetrics = RelayUpdateErrorMetrics{
	UnmarshalFailure: &EmptyCounter{},
	InvalidVersion:   &EmptyCounter{},
	ExceedMaxRelays:  &EmptyCounter{},
	RelayNotFound:    &EmptyCounter{},
	InvalidToken:     &EmptyCounter{},
	RelayNotEnabled:  &EmptyCounter{},
}

func NewRelayInitMetrics(ctx context.Context, metricsHandler Handler) (*RelayInitMetrics, error) {
	initCount, err := metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Total relay init count",
		ServiceName: "relay_backend",
		ID:          "relay.init.count",
		Unit:        "requests",
		Description: "The total number of received relay init requests",
	})
	if err != nil {
		return nil, err
	}

	initDuration, err := metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Relay init duration",
		ServiceName: "relay_backend",
		ID:          "relay.init.duration",
		Unit:        "milliseconds",
		Description: "How long it takes to process a relay init request",
	})
	if err != nil {
		return nil, err
	}

	initMetrics := RelayInitMetrics{
		Invocations:   initCount,
		DurationGauge: initDuration,
		ErrorMetrics:  EmptyRelayInitErrorMetrics,
	}

	return &initMetrics, nil
}

func NewRelayUpdateMetrics(ctx context.Context, metricsHandler Handler) (*RelayUpdateMetrics, error) {
	updateCount, err := metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Total relay update count",
		ServiceName: "relay_backend",
		ID:          "relay.update.count",
		Unit:        "requests",
		Description: "The total number of received relay update requests",
	})
	if err != nil {
		return nil, err
	}

	updateDuration, err := metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Relay update duration",
		ServiceName: "relay_backend",
		ID:          "relay.update.duration",
		Unit:        "milliseconds",
		Description: "How long it takes to process a relay update request.",
	})
	if err != nil {
		return nil, err
	}

	updateMetrics := RelayUpdateMetrics{
		Invocations:   updateCount,
		DurationGauge: updateDuration,
		ErrorMetrics:  EmptyRelayUpdateErrorMetrics,
	}

	return &updateMetrics, nil
}
