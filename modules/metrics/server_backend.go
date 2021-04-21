package metrics

import (
	"context"
)

// ServerInitMetrics defines the set of metrics for the server init handler in the server backend.
type ServerInitMetrics struct {
	HandlerMetrics *PacketHandlerMetrics

	ReadPacketFailure            Counter
	BuyerNotFound                Counter
	BuyerNotActive               Counter
	SignatureCheckFailed         Counter
	SDKTooOld                    Counter
	DatacenterMapNotFound        Counter
	DatacenterNotFound           Counter
	MisconfiguredDatacenterAlias Counter
	DatacenterNotAllowed         Counter
	WriteResponseFailure         Counter
}

// EmptyServerInitMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyServerInitMetrics = ServerInitMetrics{
	HandlerMetrics:               &EmptyPacketHandlerMetrics,
	ReadPacketFailure:            &EmptyCounter{},
	BuyerNotFound:                &EmptyCounter{},
	BuyerNotActive:               &EmptyCounter{},
	SignatureCheckFailed:         &EmptyCounter{},
	SDKTooOld:                    &EmptyCounter{},
	DatacenterMapNotFound:        &EmptyCounter{},
	DatacenterNotFound:           &EmptyCounter{},
	MisconfiguredDatacenterAlias: &EmptyCounter{},
	DatacenterNotAllowed:         &EmptyCounter{},
	WriteResponseFailure:         &EmptyCounter{},
}

// ServerUpdateMetrics defines the set of metrics for the server update handler in the server backend.
type ServerUpdateMetrics struct {
	HandlerMetrics *PacketHandlerMetrics

	ReadPacketFailure            Counter
	BuyerNotFound                Counter
	SignatureCheckFailed         Counter
	SDKTooOld                    Counter
	DatacenterMapNotFound        Counter
	DatacenterNotFound           Counter
	MisconfiguredDatacenterAlias Counter
	DatacenterNotAllowed         Counter
}

// EmptyServerUpdateMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyServerUpdateMetrics = ServerUpdateMetrics{
	HandlerMetrics:               &EmptyPacketHandlerMetrics,
	ReadPacketFailure:            &EmptyCounter{},
	BuyerNotFound:                &EmptyCounter{},
	SignatureCheckFailed:         &EmptyCounter{},
	SDKTooOld:                    &EmptyCounter{},
	DatacenterMapNotFound:        &EmptyCounter{},
	DatacenterNotFound:           &EmptyCounter{},
	MisconfiguredDatacenterAlias: &EmptyCounter{},
	DatacenterNotAllowed:         &EmptyCounter{},
}

// SessionUpdateMetrics defines the set of metrics for the session update handler in the server backend.
type SessionUpdateMetrics struct {
	HandlerMetrics *PacketHandlerMetrics

	DirectSlices Counter
	NextSlices   Counter

	ReadPacketFailure                          Counter
	FallbackToDirectUnknownReason              Counter
	FallbackToDirectBadRouteToken              Counter
	FallbackToDirectNoNextRouteToContinue      Counter
	FallbackToDirectPreviousUpdateStillPending Counter
	FallbackToDirectBadContinueToken           Counter
	FallbackToDirectRouteExpired               Counter
	FallbackToDirectRouteRequestTimedOut       Counter
	FallbackToDirectContinueRequestTimedOut    Counter
	FallbackToDirectClientTimedOut             Counter
	FallbackToDirectUpgradeResponseTimedOut    Counter
	FallbackToDirectRouteUpdateTimedOut        Counter
	FallbackToDirectDirectPongTimedOut         Counter
	FallbackToDirectNextPongTimedOut           Counter
	BuyerNotFound                              Counter
	SignatureCheckFailed                       Counter
	ClientLocateFailure                        Counter
	ReadSessionDataFailure                     Counter
	BadSessionID                               Counter
	BadSliceNumber                             Counter
	BuyerNotLive                               Counter
	ClientPingTimedOut                         Counter
	DatacenterMapNotFound                      Counter
	DatacenterNotFound                         Counter
	MisconfiguredDatacenterAlias               Counter
	DatacenterNotAllowed                       Counter
	NearRelaysLocateFailure                    Counter
	NearRelaysChanged                          Counter
	NoRelaysInDatacenter                       Counter
	RouteDoesNotExist                          Counter
	RouteSwitched                              Counter
	SDKAborted                                 Counter
	NoRoute                                    Counter
	MultipathOverload                          Counter
	LatencyWorse                               Counter
	MispredictVeto                             Counter
	WriteResponseFailure                       Counter
}

// EmptySessionUpdateMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptySessionUpdateMetrics = SessionUpdateMetrics{
	HandlerMetrics:                             &EmptyPacketHandlerMetrics,
	DirectSlices:                               &EmptyCounter{},
	NextSlices:                                 &EmptyCounter{},
	ReadPacketFailure:                          &EmptyCounter{},
	FallbackToDirectUnknownReason:              &EmptyCounter{},
	FallbackToDirectBadRouteToken:              &EmptyCounter{},
	FallbackToDirectNoNextRouteToContinue:      &EmptyCounter{},
	FallbackToDirectPreviousUpdateStillPending: &EmptyCounter{},
	FallbackToDirectBadContinueToken:           &EmptyCounter{},
	FallbackToDirectRouteExpired:               &EmptyCounter{},
	FallbackToDirectRouteRequestTimedOut:       &EmptyCounter{},
	FallbackToDirectContinueRequestTimedOut:    &EmptyCounter{},
	FallbackToDirectClientTimedOut:             &EmptyCounter{},
	FallbackToDirectUpgradeResponseTimedOut:    &EmptyCounter{},
	FallbackToDirectRouteUpdateTimedOut:        &EmptyCounter{},
	FallbackToDirectDirectPongTimedOut:         &EmptyCounter{},
	FallbackToDirectNextPongTimedOut:           &EmptyCounter{},
	BuyerNotFound:                              &EmptyCounter{},
	SignatureCheckFailed:                       &EmptyCounter{},
	ClientLocateFailure:                        &EmptyCounter{},
	ReadSessionDataFailure:                     &EmptyCounter{},
	BadSessionID:                               &EmptyCounter{},
	BadSliceNumber:                             &EmptyCounter{},
	BuyerNotLive:                               &EmptyCounter{},
	ClientPingTimedOut:                         &EmptyCounter{},
	DatacenterMapNotFound:                      &EmptyCounter{},
	DatacenterNotFound:                         &EmptyCounter{},
	MisconfiguredDatacenterAlias:               &EmptyCounter{},
	DatacenterNotAllowed:                       &EmptyCounter{},
	NearRelaysLocateFailure:                    &EmptyCounter{},
	NearRelaysChanged:                          &EmptyCounter{},
	NoRelaysInDatacenter:                       &EmptyCounter{},
	RouteDoesNotExist:                          &EmptyCounter{},
	RouteSwitched:                              &EmptyCounter{},
	SDKAborted:                                 &EmptyCounter{},
	NoRoute:                                    &EmptyCounter{},
	MultipathOverload:                          &EmptyCounter{},
	LatencyWorse:                               &EmptyCounter{},
	MispredictVeto:                             &EmptyCounter{},
	WriteResponseFailure:                       &EmptyCounter{},
}

// ServerBackendMetrics defines the set of metrics for the server backend.
type ServerBackendMetrics struct {
	ServiceMetrics *ServiceMetrics

	ServerInitMetrics    *ServerInitMetrics
	ServerUpdateMetrics  *ServerUpdateMetrics
	SessionUpdateMetrics *SessionUpdateMetrics

	PostSessionMetrics *PostSessionMetrics

	BillingMetrics *BillingMetrics

	RouteMatrixUpdateDuration     Gauge
	RouteMatrixUpdateLongDuration Counter
	RouteMatrixNumRoutes          Gauge
	RouteMatrixBytes              Gauge

	BinWrapperEmpty   Counter
	BinWrapperFailure Counter
	StaleRouteMatrix  Counter
}

// EmptyServerBackendMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyServerBackendMetrics = ServerBackendMetrics{
	ServiceMetrics:                &EmptyServiceMetrics,
	ServerInitMetrics:             &EmptyServerInitMetrics,
	ServerUpdateMetrics:           &EmptyServerUpdateMetrics,
	SessionUpdateMetrics:          &EmptySessionUpdateMetrics,
	PostSessionMetrics:            &EmptyPostSessionMetrics,
	BillingMetrics:                &EmptyBillingMetrics,
	RouteMatrixUpdateDuration:     &EmptyGauge{},
	RouteMatrixUpdateLongDuration: &EmptyCounter{},
	RouteMatrixNumRoutes:          &EmptyGauge{},
	RouteMatrixBytes:              &EmptyGauge{},
	BinWrapperEmpty:			   &EmptyCounter{},
	BinWrapperFailure:             &EmptyCounter{},
	StaleRouteMatrix:              &EmptyCounter{},
}

// NewServerBackendMetrics creates the metrics that the server backend will use.
func NewServerBackendMetrics(ctx context.Context, handler Handler) (*ServerBackendMetrics, error) {
	serviceName := "server_backend"

	var err error
	m := &ServerBackendMetrics{}

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

	m.BinWrapperEmpty, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Server Backend Bin Wrapper Empty",
		ServiceName: serviceName,
		ID:          "server_backend.bin_wrapper_empty",
		Unit:        "errors",
		Description: "The number of times the " + serviceName + " received an empty database bin wrapper from the route matrix.",
	})
	if err != nil {
		return nil, err
	}

	m.BinWrapperFailure, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Server Backend Bin Wrapper Failure",
		ServiceName: serviceName,
		ID:          "server_backend.bin_wrapper_failure",
		Unit:        "errors",
		Description: "The number of times the " + serviceName + " failed to decode the database bin wrapper from the route matrix.",
	})
	if err != nil {
		return nil, err
	}

	m.StaleRouteMatrix, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Route Matrix Stale",
		ServiceName: serviceName,
		ID:          "route_matrix_update.stale",
		Unit:        "count",
		Description: "The number of times the route matrix has gone stale from updates.",
	})
	if err != nil {
		return nil, err
	}

	return m, nil
}

func newServerInitMetrics(ctx context.Context, handler Handler, serviceName string, handlerID string, handlerName string, packetDescription string) (*ServerInitMetrics, error) {
	var err error
	m := &ServerInitMetrics{}

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

	m.BuyerNotActive, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Buyer Not Active",
		ServiceName: serviceName,
		ID:          handlerID + ".buyer_not_active",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " contained an inactive customer account.",
	})
	if err != nil {
		return nil, err
	}

	m.SignatureCheckFailed, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Signature Check Failed",
		ServiceName: serviceName,
		ID:          handlerID + ".signature_check_failed",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " failed the signature check to verify the customer's identity.",
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

	m.DatacenterMapNotFound, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Datacenter Map Not Found",
		ServiceName: serviceName,
		ID:          handlerID + ".datacenter_map_not_found",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " could not find a datacenter map for a buyer.",
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

	m.MisconfiguredDatacenterAlias, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Misconfigured Datacenter Alias",
		ServiceName: serviceName,
		ID:          handlerID + ".misconfigured_datacenter_alias",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " contained a valid datacenter alias but the datacenter map was misconfigured in our database.",
	})
	if err != nil {
		return nil, err
	}

	m.DatacenterNotAllowed, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Datacenter Not Allowed",
		ServiceName: serviceName,
		ID:          handlerID + ".datacenter_not_allowed",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " contained a valid datacenter but the buyer was not configured to use it.",
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

func newServerUpdateMetrics(ctx context.Context, handler Handler, serviceName string, handlerID string, handlerName string, packetDescription string) (*ServerUpdateMetrics, error) {
	var err error
	m := &ServerUpdateMetrics{}

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

	m.SignatureCheckFailed, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Signature Check Failed",
		ServiceName: serviceName,
		ID:          handlerID + ".signature_check_failed",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " failed the signature check to verify the customer's identity.",
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

	m.DatacenterMapNotFound, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Datacenter Map Not Found",
		ServiceName: serviceName,
		ID:          handlerID + ".datacenter_map_not_found",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " could not find a datacenter map for a buyer.",
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

	m.MisconfiguredDatacenterAlias, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Misconfigured Datacenter Alias",
		ServiceName: serviceName,
		ID:          handlerID + ".misconfigured_datacenter_alias",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " contained a valid datacenter alias but the datacenter map was misconfigured in our database.",
	})
	if err != nil {
		return nil, err
	}

	m.DatacenterNotAllowed, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Datacenter Not Allowed",
		ServiceName: serviceName,
		ID:          handlerID + ".datacenter_not_allowed",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " contained a valid datacenter but the buyer was not configured to use it.",
	})
	if err != nil {
		return nil, err
	}

	return m, nil
}

func newSessionUpdateMetrics(ctx context.Context, handler Handler, serviceName string, handlerID string, handlerName string, packetDescription string) (*SessionUpdateMetrics, error) {
	var err error
	m := &SessionUpdateMetrics{}

	m.HandlerMetrics, err = NewPacketHandlerMetrics(ctx, handler, serviceName, handlerID, handlerName, packetDescription)
	if err != nil {
		return nil, err
	}

	m.DirectSlices, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Direct Slices",
		ServiceName: serviceName,
		ID:          handlerID + ".direct_slices",
		Unit:        "slices",
		Description: "The number of slices taking a direct route.",
	})
	if err != nil {
		return nil, err
	}

	m.NextSlices, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Next Slices",
		ServiceName: serviceName,
		ID:          handlerID + ".next_slices",
		Unit:        "slices",
		Description: "The number of slices taking a network next route.",
	})
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

	m.FallbackToDirectUnknownReason, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Fallback To Direct Unknown Reason",
		ServiceName: serviceName,
		ID:          handlerID + ".fallback_to_direct",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " fell back to direct for some unknown reason.",
	})
	if err != nil {
		return nil, err
	}

	m.FallbackToDirectBadRouteToken, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Fallback To Direct Bad Route Token",
		ServiceName: serviceName,
		ID:          handlerID + ".fallback_to_direct.bad_route_token",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " fell back to direct due to a bad route token.",
	})
	if err != nil {
		return nil, err
	}

	m.FallbackToDirectNoNextRouteToContinue, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Fallback To Direct No Next Route To Continue",
		ServiceName: serviceName,
		ID:          handlerID + ".fallback_to_direct.no_next_route_to_continue",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " fell back to direct due to no next route to continue.",
	})
	if err != nil {
		return nil, err
	}

	m.FallbackToDirectPreviousUpdateStillPending, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Fallback To Direct Previous Update Still Pending",
		ServiceName: serviceName,
		ID:          handlerID + ".fallback_to_direct.previous_update_still_pending",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " fell back to direct due to the previous update still pending.",
	})
	if err != nil {
		return nil, err
	}

	m.FallbackToDirectBadContinueToken, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Fallback To Direct Bad Continue Token",
		ServiceName: serviceName,
		ID:          handlerID + ".fallback_to_direct.bad_continue_token",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " fell back to direct due to a bad continue token.",
	})
	if err != nil {
		return nil, err
	}

	m.FallbackToDirectRouteExpired, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Fallback To Direct Route Expired",
		ServiceName: serviceName,
		ID:          handlerID + ".fallback_to_direct.route_expired",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " fell back to direct due to the route expiring.",
	})
	if err != nil {
		return nil, err
	}

	m.FallbackToDirectRouteRequestTimedOut, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Fallback To Direct Route Request Timed Out",
		ServiceName: serviceName,
		ID:          handlerID + ".fallback_to_direct.route_request_timed_out",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " fell back to direct due to the route request timing out.",
	})
	if err != nil {
		return nil, err
	}

	m.FallbackToDirectContinueRequestTimedOut, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Fallback To Direct Continue Request Timed Out",
		ServiceName: serviceName,
		ID:          handlerID + ".fallback_to_direct.continue_request_timed_out",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " fell back to direct due to the continue request timing out.",
	})
	if err != nil {
		return nil, err
	}

	m.FallbackToDirectClientTimedOut, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Fallback To Direct Client Timed Out",
		ServiceName: serviceName,
		ID:          handlerID + ".fallback_to_direct.client_timed_out",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " fell back to direct due to the client timing out.",
	})
	if err != nil {
		return nil, err
	}

	m.FallbackToDirectUpgradeResponseTimedOut, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Fallback To Direct Upgradr Response Timed Out",
		ServiceName: serviceName,
		ID:          handlerID + ".fallback_to_direct.upgrade_response_timed_out",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " fell back to direct due to the upgrade response timing out.",
	})
	if err != nil {
		return nil, err
	}

	m.FallbackToDirectRouteUpdateTimedOut, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Fallback To Direct Route Update Timed Out",
		ServiceName: serviceName,
		ID:          handlerID + ".fallback_to_direct.route_update_timed_out",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " fell back to direct due to the route update timing out.",
	})
	if err != nil {
		return nil, err
	}

	m.FallbackToDirectDirectPongTimedOut, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Fallback To Direct Direct Pong Timed Out",
		ServiceName: serviceName,
		ID:          handlerID + ".fallback_to_direct.direct_pong_timed_out",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " fell back to direct due to the direct pong timing out.",
	})
	if err != nil {
		return nil, err
	}

	m.FallbackToDirectNextPongTimedOut, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Fallback To Direct Next Pong Timed Out",
		ServiceName: serviceName,
		ID:          handlerID + ".fallback_to_direct.next_pong_timed_out",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " fell back to direct due to the next pong timing out.",
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

	m.SignatureCheckFailed, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Signature Check Failed",
		ServiceName: serviceName,
		ID:          handlerID + ".signature_check_failed",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " failed the signature check to verify the customer's identity.",
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

	m.ClientPingTimedOut, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Client Ping Timed Out",
		ServiceName: serviceName,
		ID:          handlerID + ".client_ping_timed_out",
		Unit:        "timeouts",
		Description: "The number of times a " + packetDescription + " contained a client ping timeout reported up from the server.",
	})
	if err != nil {
		return nil, err
	}

	m.DatacenterMapNotFound, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Datacenter Map Not Found",
		ServiceName: serviceName,
		ID:          handlerID + ".datacenter_map_not_found",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " could not find a datacenter map for a buyer.",
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

	m.MisconfiguredDatacenterAlias, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Misconfigured Datacenter Alias",
		ServiceName: serviceName,
		ID:          handlerID + ".misconfigured_datacenter_alias",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " contained a valid datacenter alias but the datacenter map was misconfigured in our database.",
	})
	if err != nil {
		return nil, err
	}

	m.DatacenterNotAllowed, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Datacenter Not Allowed",
		ServiceName: serviceName,
		ID:          handlerID + ".datacenter_not_allowed",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " contained a valid datacenter but the buyer was not configured to use it.",
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

	m.NearRelaysChanged, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Near Relays Changed",
		ServiceName: serviceName,
		ID:          handlerID + ".near_relays_changed",
		Unit:        "errors",
		Description: "The number of times the near relays changed after the first slice for a " + packetDescription + ".",
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

	m.RouteDoesNotExist, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Route Does Not Exist",
		ServiceName: serviceName,
		ID:          handlerID + ".route_does_not_exist",
		Unit:        "errors",
		Description: "The number of times a route no longer exists for a " + packetDescription + ".",
	})
	if err != nil {
		return nil, err
	}

	m.RouteSwitched, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Route Switched",
		ServiceName: serviceName,
		ID:          handlerID + ".route_switched",
		Unit:        "errors",
		Description: "The number of times a route switched for a " + packetDescription + ".",
	})
	if err != nil {
		return nil, err
	}

	m.SDKAborted, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " SDK Aborted",
		ServiceName: serviceName,
		ID:          handlerID + ".sdk_aborted",
		Unit:        "errors",
		Description: "The number of times the SDK aborted the session for a " + packetDescription + ".",
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

	m.MispredictVeto, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Mispredict Veto",
		ServiceName: serviceName,
		ID:          handlerID + ".mispredict_veto",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + "was vetoed due too mispredicting too many times.",
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
