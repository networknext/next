package metrics

import "context"

type RelayGatewayMetrics struct {
	HandlerMetrics        *PacketHandlerMetrics
	GatewayServiceMetrics *ServiceMetrics
	UpdatesReceived       Counter
	UpdatesQueued         Gauge
	UpdatesFlushed        Counter
	ErrorMetrics          RelayGatewayErrorMetrics
}

var EmptyRelayGatewayMetrics = RelayGatewayMetrics{
	HandlerMetrics:        &EmptyPacketHandlerMetrics,
	GatewayServiceMetrics: &EmptyServiceMetrics,
	UpdatesReceived:       &EmptyCounter{},
	UpdatesQueued:         &EmptyGauge{},
	UpdatesFlushed:        &EmptyCounter{},
	ErrorMetrics:          EmptyRelayGatewayErrorMetrics,
}

type RelayGatewayErrorMetrics struct {
	ReadPacketFailure            Counter
	ContentTypeFailure           Counter
	UnmarshalFailure             Counter
	InvalidVersion               Counter
	ExceedMaxRelays              Counter
	RelayNotFound                Counter
	MarshalBinaryResponseFailure Counter
	MarshalBinaryFailure         Counter
	BackendSendFailure           Counter
}

var EmptyRelayGatewayErrorMetrics = RelayGatewayErrorMetrics{
	ReadPacketFailure:            &EmptyCounter{},
	ContentTypeFailure:           &EmptyCounter{},
	UnmarshalFailure:             &EmptyCounter{},
	InvalidVersion:               &EmptyCounter{},
	ExceedMaxRelays:              &EmptyCounter{},
	RelayNotFound:                &EmptyCounter{},
	MarshalBinaryResponseFailure: &EmptyCounter{},
	MarshalBinaryFailure:         &EmptyCounter{},
	BackendSendFailure:           &EmptyCounter{},
}

func NewRelayGatewayMetrics(ctx context.Context, metricsHandler Handler, serviceName string, handlerID string, handlerName string, packetDescription string) (*RelayGatewayMetrics, error) {
	m := &RelayGatewayMetrics{}
	var err error

	m.HandlerMetrics, err = NewPacketHandlerMetrics(ctx, metricsHandler, serviceName, handlerID, handlerName, packetDescription)
	if err != nil {
		return nil, err
	}

	m.GatewayServiceMetrics, err = NewServiceMetrics(ctx, metricsHandler, serviceName)
	if err != nil {
		return nil, err
	}

	m.UpdatesReceived, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Total Updates Received",
		ServiceName: serviceName,
		ID:          handlerID + "updates_received",
		Unit:        "updates",
		Description: "The total number of received relay update requests",
	})
	if err != nil {
		return nil, err
	}

	m.UpdatesQueued, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: handlerName + " Current Updates Queued",
		ServiceName: serviceName,
		ID:          handlerID + "updates_queued",
		Unit:        "updates",
		Description: "The current number of relay update requests queued for batch-writing to the relay backends via HTTP",
	})
	if err != nil {
		return nil, err
	}

	m.UpdatesFlushed, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Total Updates Flushed",
		ServiceName: serviceName,
		ID:          handlerID + "updates_flushed",
		Unit:        "updates",
		Description: "The total number of unique relay update requests that were sent to the relay backends via HTTP (not necessarily successful)",
	})
	if err != nil {
		return nil, err
	}

	m.ErrorMetrics.ReadPacketFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Total Read Packet Failures",
		ServiceName: serviceName,
		ID:          handlerID + ".error.read_failure",
		Unit:        "errors",
		Description: "The total number of relay update request packets that could not be read",
	})
	if err != nil {
		return nil, err
	}

	m.ErrorMetrics.ContentTypeFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Total Content Type Failures",
		ServiceName: serviceName,
		ID:          handlerID + ".error.content_type_failure",
		Unit:        "errors",
		Description: "The total number of relay update request packets that had unsupported content types",
	})
	if err != nil {
		return nil, err
	}

	m.ErrorMetrics.UnmarshalFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Total Unmarshal Failures",
		ServiceName: serviceName,
		ID:          handlerID + ".error.unmarshal_failure",
		Unit:        "errors",
		Description: "The total number of relay update requests that failed to be unmarshaled",
	})
	if err != nil {
		return nil, err
	}

	m.ErrorMetrics.InvalidVersion, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Total Invalid Versions",
		ServiceName: serviceName,
		ID:          handlerID + ".error.invalid_version",
		Unit:        "errors",
		Description: "The total number of relay update requests that had invalid versions",
	})
	if err != nil {
		return nil, err
	}

	m.ErrorMetrics.ExceedMaxRelays, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Total Exceed Max Relays",
		ServiceName: serviceName,
		ID:          handlerID + ".error.exceed_max_relays",
		Unit:        "errors",
		Description: "The total number of relay update requests that had too many relays in ping stats",
	})
	if err != nil {
		return nil, err
	}

	m.ErrorMetrics.RelayNotFound, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Total Relays Not Found",
		ServiceName: serviceName,
		ID:          handlerID + ".error.relay_not_found",
		Unit:        "errors",
		Description: "The total number of relay update requests that could not be found in the internal map",
	})
	if err != nil {
		return nil, err
	}

	m.ErrorMetrics.MarshalBinaryResponseFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Total Marshal Binary Response Failures",
		ServiceName: serviceName,
		ID:          handlerID + ".error.marhsal_binary_response_failure",
		Unit:        "errors",
		Description: "The total number of relay update responses that could not be marshaled",
	})
	if err != nil {
		return nil, err
	}

	m.ErrorMetrics.MarshalBinaryFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Total Marshal Binary Failures",
		ServiceName: serviceName,
		ID:          handlerID + ".error.marhsal_binary_failure",
		Unit:        "errors",
		Description: "The total number of relay update request batches that could not be marshaled",
	})
	if err != nil {
		return nil, err
	}

	m.ErrorMetrics.BackendSendFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Total Backend Send Failures",
		ServiceName: serviceName,
		ID:          handlerID + ".error.backend_send_failure",
		Unit:        "errors",
		Description: "The total number of relay update request batches that failed to be sent to the relay backends",
	})
	if err != nil {
		return nil, err
	}

	return m, nil
}
