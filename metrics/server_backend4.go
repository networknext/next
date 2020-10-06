package metrics

import (
	"context"
)

// ServerInit4Metrics defines the set of metrics for the server init handler in server_backend4.
type ServerInit4Metrics struct {
	HandlerMetrics *PacketHandlerMetrics

	ReadPacketFailure    Counter
	BuyerNotFound        Counter
	SDKTooOld            Counter
	DatacenterNotFound   Counter
	WriteResponseFailure Counter
}

// EmptyServerInit4Metrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyServerInit4Metrics = ServerInit4Metrics{
	HandlerMetrics:       &EmptyPacketHandlerMetrics,
	ReadPacketFailure:    &EmptyCounter{},
	BuyerNotFound:        &EmptyCounter{},
	SDKTooOld:            &EmptyCounter{},
	DatacenterNotFound:   &EmptyCounter{},
	WriteResponseFailure: &EmptyCounter{},
}

// ServerUpdate4Metrics defines the set of metrics for the server update handler in server_backend4.
type ServerUpdate4Metrics struct {
	HandlerMetrics *PacketHandlerMetrics

	ReadPacketFailure  Counter
	BuyerNotFound      Counter
	SDKTooOld          Counter
	DatacenterNotFound Counter
}

// EmptyServerUpdate4Metrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyServerUpdate4Metrics = ServerUpdate4Metrics{
	HandlerMetrics:     &EmptyPacketHandlerMetrics,
	ReadPacketFailure:  &EmptyCounter{},
	BuyerNotFound:      &EmptyCounter{},
	SDKTooOld:          &EmptyCounter{},
	DatacenterNotFound: &EmptyCounter{},
}

// SessionUpdate4Metrics defines the set of metrics for the session update handler in server_backend4.
type SessionUpdate4Metrics struct {
	HandlerMetrics *PacketHandlerMetrics

	ReadPacketFailure       Counter
	BuyerNotFound           Counter
	ClientLocateFailure     Counter
	ReadSessionDataFailure  Counter
	BadSessionID            Counter
	BadSliceNumber          Counter
	BuyerNotLive            Counter
	DatacenterNotFound      Counter
	NearRelaysLocateFailure Counter
	NoRelaysInDatacenter    Counter
	NoRoute                 Counter
	MultipathOverload       Counter
	LatencyWorse            Counter
	WriteResponseFailure    Counter
}

// EmptySessionUpdate4Metrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptySessionUpdate4Metrics = SessionUpdate4Metrics{
	HandlerMetrics:          &EmptyPacketHandlerMetrics,
	ReadPacketFailure:       &EmptyCounter{},
	BuyerNotFound:           &EmptyCounter{},
	ClientLocateFailure:     &EmptyCounter{},
	ReadSessionDataFailure:  &EmptyCounter{},
	BadSessionID:            &EmptyCounter{},
	BadSliceNumber:          &EmptyCounter{},
	BuyerNotLive:            &EmptyCounter{},
	DatacenterNotFound:      &EmptyCounter{},
	NearRelaysLocateFailure: &EmptyCounter{},
	NoRelaysInDatacenter:    &EmptyCounter{},
	NoRoute:                 &EmptyCounter{},
	MultipathOverload:       &EmptyCounter{},
	LatencyWorse:            &EmptyCounter{},
	WriteResponseFailure:    &EmptyCounter{},
}

// ServerBackend4Metrics defines the set of metrics for the server_backend4 service.
type ServerBackend4Metrics struct {
	ServiceMetrics *ServiceMetrics

	ServerInitMetrics    *ServerInit4Metrics
	ServerUpdateMetrics  *ServerUpdate4Metrics
	SessionUpdateMetrics *SessionUpdate4Metrics

	PostSessionMetrics *PostSessionMetrics

	BillingMetrics *BillingMetrics

	RouteMatrixUpdateDuration     Gauge
	RouteMatrixUpdateLongDuration Counter
	RouteMatrixNumRoutes          Gauge
	RouteMatrixBytes              Gauge
}

// EmptyServerBackend4Metrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyServerBackend4Metrics = ServerBackend4Metrics{
	ServiceMetrics:                &EmptyServiceMetrics,
	ServerInitMetrics:             &EmptyServerInit4Metrics,
	ServerUpdateMetrics:           &EmptyServerUpdate4Metrics,
	SessionUpdateMetrics:          &EmptySessionUpdate4Metrics,
	PostSessionMetrics:            &EmptyPostSessionMetrics,
	BillingMetrics:                &EmptyBillingMetrics,
	RouteMatrixUpdateDuration:     &EmptyGauge{},
	RouteMatrixUpdateLongDuration: &EmptyCounter{},
	RouteMatrixNumRoutes:          &EmptyGauge{},
	RouteMatrixBytes:              &EmptyGauge{},
}

// NewServerBackend4Metrics creates the metrics that the server_backend4 will use.
func NewServerBackend4Metrics(ctx context.Context, handler Handler) (*ServerBackend4Metrics, error) {
	serviceName := "server_backend4"

	var err error
	m := &ServerBackend4Metrics{}

	m.ServiceMetrics, err = NewServiceMetrics(ctx, handler, serviceName)
	if err != nil {
		return nil, err
	}

	m.ServerInitMetrics, err = newServerInitMetrics(ctx, handler, serviceName, "server_init", "Server Init", "server init request")
	if err != nil {
		return nil, err
	}

	m.ServerUpdateMetrics, err = newServerUpdateMetrics(ctx, handler, serviceName, "server_update", "Server Update", "server update")
	if err != nil {
		return nil, err
	}

	m.SessionUpdateMetrics, err = newSessionUpdateMetrics(ctx, handler, serviceName, "session_update", "Session Update", "session update request")
	if err != nil {
		return nil, err
	}

	m.PostSessionMetrics, err = NewPostSessionMetrics(ctx, handler, serviceName)
	if err != nil {
		return nil, err
	}

	m.BillingMetrics = &BillingMetrics{}
	m.BillingMetrics.EntriesReceived = &EmptyCounter{}

	m.BillingMetrics.EntriesSubmitted, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Server Backend Billing Entries Submitted",
		ServiceName: serviceName,
		ID:          "billing.entries_submitted",
		Unit:        "entries",
		Description: "The number of billing entries the server_backend has submitted to be published",
	})
	if err != nil {
		return nil, err
	}

	m.BillingMetrics.EntriesQueued, err = handler.NewGauge(ctx, &Descriptor{
		DisplayName: "Server Backend Billing Entries Queued",
		ServiceName: serviceName,
		ID:          "billing.entries_queued",
		Unit:        "entries",
		Description: "The number of billing entries the server_backend has queued waiting to be published",
	})
	if err != nil {
		return nil, err
	}

	m.BillingMetrics.EntriesFlushed, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Server Backend Billing Entries Flushed",
		ServiceName: serviceName,
		ID:          "billing.entries_flushed",
		Unit:        "entries",
		Description: "The number of billing entries the server_backend has flushed after publishing",
	})
	if err != nil {
		return nil, err
	}

	m.BillingMetrics.ErrorMetrics.BillingPublishFailure, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Server Backend Billing Publish Failure",
		ServiceName: serviceName,
		ID:          "billing.publish_failure",
		Unit:        "entries",
		Description: "The number of billing entries the server_backend has failed to publish",
	})
	if err != nil {
		return nil, err
	}

	m.BillingMetrics.ErrorMetrics.BillingBatchedReadFailure = &EmptyCounter{}
	m.BillingMetrics.ErrorMetrics.BillingReadFailure = &EmptyCounter{}
	m.BillingMetrics.ErrorMetrics.BillingWriteFailure = &EmptyCounter{}

	m.RouteMatrixUpdateDuration, err = handler.NewGauge(ctx, &Descriptor{
		DisplayName: "Route Matrix Update Duration",
		ServiceName: serviceName,
		ID:          "route_matrix_update.duration",
		Unit:        "ms",
		Description: "The amount of time the route matrix update takes to complete in milliseconds.",
	})
	if err != nil {
		return nil, err
	}

	m.RouteMatrixUpdateLongDuration, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Route Matrix Update Long Durations",
		ServiceName: serviceName,
		ID:          "route_matrix_update.long_durations",
		Unit:        "invocations",
		Description: "The number of times the route matrix update takes longer than 100 milliseconds to complete.",
	})
	if err != nil {
		return nil, err
	}

	m.RouteMatrixNumRoutes, err = handler.NewGauge(ctx, &Descriptor{
		DisplayName: "Route Matrix Number of Routes",
		ServiceName: serviceName,
		ID:          "route_matrix_update.num_routes",
		Unit:        "routes",
		Description: "The number of routes read from the route matrix.",
	})
	if err != nil {
		return nil, err
	}

	m.RouteMatrixBytes, err = handler.NewGauge(ctx, &Descriptor{
		DisplayName: "Route Matrix Bytes",
		ServiceName: serviceName,
		ID:          "route_matrix_update.bytes",
		Unit:        "bytes",
		Description: "The number of bytes read from the route matrix.",
	})
	if err != nil {
		return nil, err
	}

	return m, nil
}

func newServerInitMetrics(ctx context.Context, handler Handler, serviceName string, handlerID string, handlerName string, packetDescription string) (*ServerInit4Metrics, error) {
	var err error
	m := &ServerInit4Metrics{}

	m.HandlerMetrics, err = NewPacketHandlerMetrics(ctx, handler, serviceName, handlerID, handlerName, packetDescription)
	if err != nil {
		return nil, err
	}

	m.ReadPacketFailure, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Read Packet Failure",
		ServiceName: serviceName,
		ID:          handlerID + ".read_packet_failure",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " failed to read.",
	})
	if err != nil {
		return nil, err
	}

	m.BuyerNotFound, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Buyer Not Found",
		ServiceName: serviceName,
		ID:          handlerID + ".buyer_not_found",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " contained an unknown customer ID.",
	})
	if err != nil {
		return nil, err
	}

	m.SDKTooOld, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " SDK Too Old",
		ServiceName: serviceName,
		ID:          handlerID + ".sdk_too_old",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " contained an out of date SDK version.",
	})
	if err != nil {
		return nil, err
	}

	m.DatacenterNotFound, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Datacenter Not Found",
		ServiceName: serviceName,
		ID:          handlerID + ".datacenter_not_found",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " contained an unknown datacenter name.",
	})
	if err != nil {
		return nil, err
	}

	m.WriteResponseFailure, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Write Response Failure",
		ServiceName: serviceName,
		ID:          handlerID + ".write_response_failure",
		Unit:        "errors",
		Description: "The number of times we failed to write a response to a " + packetDescription + ".",
	})
	if err != nil {
		return nil, err
	}

	return m, nil
}

func newServerUpdateMetrics(ctx context.Context, handler Handler, serviceName string, handlerID string, handlerName string, packetDescription string) (*ServerUpdate4Metrics, error) {
	var err error
	m := &ServerUpdate4Metrics{}

	m.HandlerMetrics, err = NewPacketHandlerMetrics(ctx, handler, serviceName, handlerID, handlerName, packetDescription)
	if err != nil {
		return nil, err
	}

	m.ReadPacketFailure, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Read Packet Failure",
		ServiceName: serviceName,
		ID:          handlerID + ".read_packet_failure",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " failed to read.",
	})
	if err != nil {
		return nil, err
	}

	m.BuyerNotFound, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Buyer Not Found",
		ServiceName: serviceName,
		ID:          handlerID + ".buyer_not_found",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " contained an unknown customer ID.",
	})
	if err != nil {
		return nil, err
	}

	m.SDKTooOld, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " SDK Too Old",
		ServiceName: serviceName,
		ID:          handlerID + ".sdk_too_old",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " contained an out of date SDK version.",
	})
	if err != nil {
		return nil, err
	}

	m.DatacenterNotFound, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Datacenter Not Found",
		ServiceName: serviceName,
		ID:          handlerID + ".datacenter_not_found",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " contained an unknown datacenter ID.",
	})
	if err != nil {
		return nil, err
	}

	return m, nil
}

func newSessionUpdateMetrics(ctx context.Context, handler Handler, serviceName string, handlerID string, handlerName string, packetDescription string) (*SessionUpdate4Metrics, error) {
	var err error
	m := &SessionUpdate4Metrics{}

	m.HandlerMetrics, err = NewPacketHandlerMetrics(ctx, handler, serviceName, handlerID, handlerName, packetDescription)
	if err != nil {
		return nil, err
	}

	m.ReadPacketFailure, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Read Packet Failure",
		ServiceName: serviceName,
		ID:          handlerID + ".read_packet_failure",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " failed to read.",
	})
	if err != nil {
		return nil, err
	}

	m.BuyerNotFound, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Buyer Not Found",
		ServiceName: serviceName,
		ID:          handlerID + ".buyer_not_found",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " contained an unknown customer ID.",
	})
	if err != nil {
		return nil, err
	}

	m.ClientLocateFailure, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Client Locate Failure",
		ServiceName: serviceName,
		ID:          handlerID + ".client_locate_failure",
		Unit:        "errors",
		Description: "The number of times we failed to locate the client for a " + packetDescription + ".",
	})
	if err != nil {
		return nil, err
	}

	m.ReadSessionDataFailure, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Read Session Data Failure",
		ServiceName: serviceName,
		ID:          handlerID + ".read_session_data_failure",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " contained unreadable session data.",
	})
	if err != nil {
		return nil, err
	}

	m.BadSessionID, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Bad Session ID",
		ServiceName: serviceName,
		ID:          handlerID + ".bad_session_id",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " contained an invalid session ID in its session data.",
	})
	if err != nil {
		return nil, err
	}

	m.BadSliceNumber, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Bad Slice Number",
		ServiceName: serviceName,
		ID:          handlerID + ".bad_slice_number",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " contained an invalid slice number in its session data.",
	})
	if err != nil {
		return nil, err
	}

	m.BuyerNotLive, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Buyer Not Live",
		ServiceName: serviceName,
		ID:          handlerID + ".buyer_not_live",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " contained a buyer that is not yet paying for acceleration.",
	})
	if err != nil {
		return nil, err
	}

	m.DatacenterNotFound, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Datacenter Not Found",
		ServiceName: serviceName,
		ID:          handlerID + ".datacenter_not_found",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " contained an unknown datacenter ID.",
	})
	if err != nil {
		return nil, err
	}

	m.NearRelaysLocateFailure, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Near Relays Locate Failure",
		ServiceName: serviceName,
		ID:          handlerID + ".near_relays_locate_failure",
		Unit:        "errors",
		Description: "The number of times we failed to locate any near relays for a " + packetDescription + ".",
	})
	if err != nil {
		return nil, err
	}

	m.NoRelaysInDatacenter, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " No Relays In Datacenter",
		ServiceName: serviceName,
		ID:          handlerID + ".no_relays_in_datacenter",
		Unit:        "errors",
		Description: "The number of times we found no relays in the game server's datacenter for a " + packetDescription + ".",
	})
	if err != nil {
		return nil, err
	}

	m.NoRoute, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " No Route",
		ServiceName: serviceName,
		ID:          handlerID + ".no_route",
		Unit:        "errors",
		Description: "The number of times we could not find a route for a " + packetDescription + ".",
	})
	if err != nil {
		return nil, err
	}

	m.MultipathOverload, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Multipath Overload",
		ServiceName: serviceName,
		ID:          handlerID + ".multipath_overload",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + "'s connection was overloaded due to multipath.",
	})
	if err != nil {
		return nil, err
	}

	m.LatencyWorse, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Latency Worse",
		ServiceName: serviceName,
		ID:          handlerID + ".latency_worse",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + "'s latency was made worse by network next.",
	})
	if err != nil {
		return nil, err
	}

	m.WriteResponseFailure, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Write Response Failure",
		ServiceName: serviceName,
		ID:          handlerID + ".write_response_failure",
		Unit:        "errors",
		Description: "The number of times we failed to write a response to a " + packetDescription + ".",
	})
	if err != nil {
		return nil, err
	}

	return m, nil
}
