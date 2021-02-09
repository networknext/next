package metrics

import "context"

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

	var initErrorMetrics RelayInitErrorMetrics
	initErrorMetrics.UnmarshalFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Total relay init unmarshal failure count",
		ServiceName: "relay_gateway",
		ID:          "relay_gateway.init.errors.unmarshal_failure.count",
		Unit:        "unmarshal_failure",
		Description: "The total number of received relay init requests that resulted in unmarshal failure",
	})
	if err != nil {
		return nil, err
	}

	initErrorMetrics.InvalidMagic, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Total relay init invalid magic error count",
		ServiceName: "relay_gateway",
		ID:          "relay_gateway.init.errors.invalid_magic.count",
		Unit:        "invalid_magic",
		Description: "The total number of received relay init requests that resulted in invalid magic error",
	})
	if err != nil {
		return nil, err
	}

	initErrorMetrics.InvalidVersion, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Total relay init invalid version error count",
		ServiceName: "relay_gateway",
		ID:          "relay_gateway.init.errors.invalid_version.count",
		Unit:        "invalid_version",
		Description: "The total number of received relay init requests that resulted in invalid version error",
	})
	if err != nil {
		return nil, err
	}

	initErrorMetrics.RelayNotFound, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Total relay init relay not found error count",
		ServiceName: "relay_gateway",
		ID:          "relay_gateway.init.errors.not_found.count",
		Unit:        "relay_not_found",
		Description: "The total number of received relay init requests that resulted in relay not found error",
	})
	if err != nil {
		return nil, err
	}

	initErrorMetrics.RelayQuarantined, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Total relay init relay quarantined error count",
		ServiceName: "relay_gateway",
		ID:          "relay_gateway.init.errors.quarantined.count",
		Unit:        "relay_quarantined",
		Description: "The total number of received relay init requests that resulted in relay quarantined error",
	})
	if err != nil {
		return nil, err
	}

	initErrorMetrics.DecryptionFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Total relay init decryption failure count",
		ServiceName: "relay_gateway",
		ID:          "relay_gateway.init.errors.decryption_failure.count",
		Unit:        "decryption_failure",
		Description: "The total number of received relay init requests that resulted in decryption failure",
	})
	if err != nil {
		return nil, err
	}

	initErrorMetrics.RelayAlreadyExists, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Total relay init relay already exists count",
		ServiceName: "relay_gateway",
		ID:          "relay_gateway.init.errors.already_exists.count",
		Unit:        "relay_already_exists",
		Description: "The total number of received relay init requests that resulted in relay already exists",
	})
	if err != nil {
		return nil, err
	}

	initErrorMetrics.IPLookupFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Total relay init IP lookup failure count",
		ServiceName: "relay_gateway",
		ID:          "relay_gateway.init.errors.ip_lookup_failure.count",
		Unit:        "ip_lookup_failure",
		Description: "The total number of received relay init requests that resulted in IP lookup failure",
	})
	if err != nil {
		return nil, err
	}

	initMetrics := RelayInitMetrics{
		Invocations:   initCount,
		DurationGauge: initDuration,
		ErrorMetrics:  initErrorMetrics,
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

	var em RelayUpdateErrorMetrics
	em.UnmarshalFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Total relay update unmarshal failure count",
		ServiceName: "relay_gateway",
		ID:          "relay_gateway.update.errors.unmarshal_failure.count",
		Unit:        "unmarshal_failure",
		Description: "The total number of received relay update requests that resulted in unmarshal failure",
	})
	if err != nil {
		return nil, err
	}

	em.InvalidVersion, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Total relay update invalid version error count",
		ServiceName: "relay_gateway",
		ID:          "relay_gateway.update.errors.invalid_version.count",
		Unit:        "invalid_version",
		Description: "The total number of received relay update requests that resulted in invalid version error",
	})
	if err != nil {
		return nil, err
	}

	em.ExceedMaxRelays, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Total relay upgrade exceed max relays error count",
		ServiceName: "relay_gateway",
		ID:          "relay_gateway.update.errors.exceed_max_relays.count",
		Unit:        "exceed_max_relays",
		Description: "The total number of received relay update requests that resulted in exceed max relays error",
	})
	if err != nil {
		return nil, err
	}

	em.RelayNotFound, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Total relay update relay not found error count",
		ServiceName: "relay_gateway",
		ID:          "relay_gateway.update.errors.not_found.count",
		Unit:        "relay_not_found",
		Description: "The total number of received relay update requests that resulted in relay not found error",
	})
	if err != nil {
		return nil, err
	}

	em.InvalidToken, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Total relay update invalid token error count",
		ServiceName: "relay_gateway",
		ID:          "relay_gateway.update.errors.invalid_token.count",
		Unit:        "invalid_token",
		Description: "The total number of received relay init requests that resulted in invalid token error",
	})
	if err != nil {
		return nil, err
	}

	em.RelayNotEnabled, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Total relay update relay not enabled error count",
		ServiceName: "relay_gateway",
		ID:          "relay_gateway.init_errors.not_enabled.count",
		Unit:        "relay_not_enabled",
		Description: "The total number of received relay init requests that resulted in relay not enabled",
	})
	if err != nil {
		return nil, err
	}

	updateMetrics := RelayUpdateMetrics{
		Invocations:   updateCount,
		DurationGauge: updateDuration,
		ErrorMetrics:  em,
	}

	return &updateMetrics, nil
}
