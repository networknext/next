package metrics

import (
	"context"
	"fmt"
)

// RelayGatewayMetrics defines the set of metrics for the relay gateway.
type RelayGatewayMetrics struct {
	ServiceMetrics ServiceMetrics

	RelayInitMetrics   RelayInitMetrics
	RelayUpdateMetrics RelayUpdateMetrics
}

// EmptyRelayGatewayMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyRelayGatewayMetrics = RelayGatewayMetrics{
	ServiceMetrics: EmptyServiceMetrics,

	RelayInitMetrics:   EmptyRelayInitMetrics,
	RelayUpdateMetrics: EmptyRelayUpdateMetrics,
}

// NewRelayGatewayMetrics creates the metrics that the relay gateway will use.
func NewRelayGatewayMetrics(ctx context.Context, handler Handler) (RelayGatewayMetrics, error) {
	serviceName := "relay_gateway"

	var err error
	m := RelayGatewayMetrics{}

	m.ServiceMetrics, err = NewServiceMetrics(ctx, handler, serviceName)
	if err != nil {
		return EmptyRelayGatewayMetrics, err
	}

	m.RelayInitMetrics, err = newRelayInitMetrics(ctx, handler, serviceName, "relay_init", "Relay Init", "relay init request")
	if err != nil {
		return EmptyRelayGatewayMetrics, fmt.Errorf("failed to create relay init metrics: %v", err)
	}

	m.RelayUpdateMetrics, err = newRelayUpdateMetrics(ctx, handler, serviceName, "relay_update", "Relay Update", "relay update request")
	if err != nil {
		return EmptyRelayGatewayMetrics, fmt.Errorf("failed to create relay update metrics: %v", err)
	}

	return m, nil
}

func newRelayInitMetrics(ctx context.Context, handler Handler, serviceName string, handlerID string, handlerName string, packetDescription string) (RelayInitMetrics, error) {
	var err error
	m := RelayInitMetrics{}

	m.HandlerMetrics, err = NewRoutineMetrics(ctx, handler, serviceName, handlerID, handlerName, packetDescription)
	if err != nil {
		return EmptyRelayInitMetrics, err
	}

	m.UnmarshalFailure, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Unmarshal Failure",
		ServiceName: serviceName,
		ID:          handlerID + ".unmarshal_failure",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " could not be unmarshaled.",
	})
	if err != nil {
		return EmptyRelayInitMetrics, err
	}

	m.InvalidMagic, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Invalid Magic",
		ServiceName: serviceName,
		ID:          handlerID + ".invalid_magic",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " contained in an invalid magic number.",
	})
	if err != nil {
		return EmptyRelayInitMetrics, err
	}

	m.InvalidVersion, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Invalid Version",
		ServiceName: serviceName,
		ID:          handlerID + ".invalid_version",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " contained in an invalid version number.",
	})
	if err != nil {
		return EmptyRelayInitMetrics, err
	}

	m.RelayNotFound, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Relay Not Found",
		ServiceName: serviceName,
		ID:          handlerID + ".relay_not_found",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " came from an unknown relay.",
	})
	if err != nil {
		return EmptyRelayInitMetrics, err
	}

	m.RelayQuarantined, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Relay Quarantined",
		ServiceName: serviceName,
		ID:          handlerID + ".relay_quarantined",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " came from a quarantined relay.",
	})
	if err != nil {
		return EmptyRelayInitMetrics, err
	}

	m.RelayAlreadyExists, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Relay Already Exists",
		ServiceName: serviceName,
		ID:          handlerID + ".relay_already_exists",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " came from a relay that has already initialized.",
	})
	if err != nil {
		return EmptyRelayInitMetrics, err
	}

	m.DecryptionFailure, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Decryption Failure",
		ServiceName: serviceName,
		ID:          handlerID + ".decryption_failure",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " contained a token that could not be decrypted.",
	})
	if err != nil {
		return EmptyRelayInitMetrics, err
	}

	return m, nil
}

func newRelayUpdateMetrics(ctx context.Context, handler Handler, serviceName string, handlerID string, handlerName string, packetDescription string) (RelayUpdateMetrics, error) {
	var err error
	m := RelayUpdateMetrics{}

	m.HandlerMetrics, err = NewRoutineMetrics(ctx, handler, serviceName, handlerID, handlerName, packetDescription)
	if err != nil {
		return EmptyRelayUpdateMetrics, err
	}

	m.UnmarshalFailure, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Unmarshal Failure",
		ServiceName: serviceName,
		ID:          handlerID + ".unmarshal_failure",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " could not be unmarshaled.",
	})
	if err != nil {
		return EmptyRelayUpdateMetrics, err
	}

	m.InvalidVersion, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Invalid Version",
		ServiceName: serviceName,
		ID:          handlerID + ".invalid_version",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " contained in an invalid version number.",
	})
	if err != nil {
		return EmptyRelayUpdateMetrics, err
	}

	m.ExceedMaxRelays, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Exceeded Max Relays",
		ServiceName: serviceName,
		ID:          handlerID + ".exceeded_max_relays",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " contained too many relay ping stats.",
	})
	if err != nil {
		return EmptyRelayUpdateMetrics, err
	}

	m.RelayNotFound, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Relay Not Found",
		ServiceName: serviceName,
		ID:          handlerID + ".relay_not_found",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " came from an unknown relay.",
	})
	if err != nil {
		return EmptyRelayUpdateMetrics, err
	}

	m.InvalidToken, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Invalid Token",
		ServiceName: serviceName,
		ID:          handlerID + ".invalid_token",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " contained an invalid token.",
	})
	if err != nil {
		return EmptyRelayUpdateMetrics, err
	}

	m.RelayNotEnabled, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Relay Not Enabled",
		ServiceName: serviceName,
		ID:          handlerID + ".relay_not_enabled",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " came from a relay that is not enabled.",
	})
	if err != nil {
		return EmptyRelayUpdateMetrics, err
	}

	return m, nil
}
