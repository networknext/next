package metrics

import (
	"context"
)

// ServerBackendStatus defines the metrics reported by the service's status endpoint
type ServerBackendStatus struct {
	// Service Information
	ServiceName string `json:"service_name"`
	GitHash     string `json:"git_hash"`
	Started     string `json:"started"`
	Uptime      string `json:"uptime"`

	// Service Metrics
	Goroutines      int     `json:"goroutines"`
	MemoryAllocated float64 `json:"mb_allocated"`

	// Server Init Metrics
	ServerInitInvocations           int `json:"server_init_invocations"`
	ServerInitReadPacketFailure     int `json:"server_init_read_packet_failure"`
	ServerInitBuyerNotFound         int `json:"server_init_buyer_not_found"`
	ServerInitBuyerNotActive        int `json:"server_init_buyer_not_active"`
	ServerInitSignatureCheckFailed  int `json:"server_init_signature_check_failed"`
	ServerInitSDKTooOld             int `json:"server_init_sdk_too_old"`
	ServerInitDatacenterMapNotFound int `json:"server_init_datacneter_map_not_found"`
	ServerInitDatacenterNotFound    int `json:"server_init_datacenter_not_found"`
	ServerInitWriteResponseFailure  int `json:"server_init_write_response_failure"`

	// Server Update Metrics
	ServerUpdateInvocations           int `json:"server_update_invocations"`
	ServerUpdateReadPacketFailure     int `json:"server_update_read_packet_failure"`
	ServerUpdateBuyerNotFound         int `json:"server_update_buyer_not_found"`
	ServerUpdateBuyerNotLive          int `json:"server_update_buyer_not_live"`
	ServerUpdateSignatureCheckFailed  int `json:"server_update_signature_check_failed"`
	ServerUpdateSDKTooOld             int `json:"server_update_sdk_too_old"`
	ServerUpdateDatacenterMapNotFound int `json:"server_update_datacneter_map_not_found"`
	ServerUpdateDatacenterNotFound    int `json:"server_update_datacenter_not_found"`

	// Session Update Metrics
	SessionUpdateInvocations                                int `json:"session_update_invocations"`
	SessionUpdateDirectSlices                               int `json:"session_update_direct_slices"`
	SessionUpdateNextSlices                                 int `json:"session_update_next_slices"`
	SessionUpdateReadPacketFailure                          int `json:"session_update_read_packet_failure"`
	SessionUpdateFallbackToDirectUnknownReason              int `json:"session_update_fallback_to_direct_unknown_reason"`
	SessionUpdateFallbackToDirectBadRouteToken              int `json:"session_update_fallback_to_direct_bad_route_token"`
	SessionUpdateFallbackToDirectNoNextRouteToContinue      int `json:"session_update_fallback_to_direct_no_next_route_token_to_continue"`
	SessionUpdateFallbackToDirectPreviousUpdateStillPending int `json:"session_update_fallback_to_direct_previous_update_still_pending"`
	SessionUpdateFallbackToDirectBadContinueToken           int `json:"session_update_fallback_to_direct_bad_continue_token"`
	SessionUpdateFallbackToDirectRouteExpired               int `json:"session_update_fallback_to_direct_route_expired"`
	SessionUpdateFallbackToDirectRouteRequestTimedOut       int `json:"session_update_fallback_to_direct_route_request_timed_out"`
	SessionUpdateFallbackToDirectContinueRequestTimedOut    int `json:"session_update_fallback_to_direct_continue_request_timed_out"`
	SessionUpdateFallbackToDirectClientTimedOut             int `json:"session_update_fallback_to_direct_client_timed_out"`
	SessionUpdateFallbackToDirectUpgradeResponseTimedOut    int `json:"session_update_fallback_to_direct_upgrade_response_timed_out"`
	SessionUpdateFallbackToDirectRouteUpdateTimedOut        int `json:"session_update_fallback_to_direct_route_update_timed_out"`
	SessionUpdateFallbackToDirectPongTimedOut               int `json:"session_update_fallback_to_direct_pong_timed_out"`
	SessionUpdateFallbackToDirectNextPongTimedOut           int `json:"session_update_fallback_to_direct_next_pong_timed_out"`
	SessionUpdateBuyerNotFound                              int `json:"session_update_buyer_not_found"`
	SessionUpdateSignatureCheckFailed                       int `json:"session_update_signature_check_failed"`
	SessionUpdateClientLocateFailure                        int `json:"session_update_client_locate_failure"`
	SessionUpdateReadSessionDataFailure                     int `json:"session_update_read_session_data_failure"`
	SessionUpdateBadSessionID                               int `json:"session_update_bad_session_id"`
	SessionUpdateBadSliceNumber                             int `json:"session_update_bad_slice_number"`
	SessionUpdateBuyerNotLive                               int `json:"session_update_buyer_not_live"`
	SessionUpdateClientPingTimedOut                         int `json:"session_update_client_ping_timed_out"`
	SessionUpdateDatacenterMapNotFound                      int `json:"session_update_datacenter_map_not_found"`
	SessionUpdateDatacenterNotFound                         int `json:"session_update_datacenter_not_found"`
	SessionUpdateDatacenterNotEnabled                       int `json:"session_update_datacenter_not_enabled"`
	SessionUpdateNearRelaysLocateFailure                    int `json:"session_update_near_relays_locate_failure"`
	SessionUpdateNearRelaysChanged                          int `json:"session_update_near_relays_changed"`
	SessionUpdateNoRelaysInDatacenter                       int `json:"session_update_no_relays_in_datacenter"`
	SessionUpdateRouteDoesNotExist                          int `json:"session_update_route_does_not_exist"`
	SessionUpdateRouteSwitched                              int `json:"session_update_route_switched"`
	SessionUpdateNextWithoutRouteRelays                     int `json:"session_update_next_without_route_relays"`
	SessionUpdateSDKAborted                                 int `json:"session_update_sdk_aborted"`
	SessionUpdateNoRoute                                    int `json:"session_update_no_route"`
	SessionUpdateMultipathOverload                          int `json:"session_update_multipath_overload"`
	SessionUpdateLatencyWorse                               int `json:"session_update_latency_worse"`
	SessionUpdateMispredictVeto                             int `json:"session_update_mispredict_veto"`
	SessionUpdateWriteResponseFailure                       int `json:"session_update_writeresponse_failure"`
	SessionUpdateStaleRouteMatrix                           int `json:"session_update_stale_route_matrix"`

	// Match Data Handler Metrics
	MatchDataHandlerInvocations          int `json:"match_data_handler_invocations"`
	MatchDataHandlerReadPacketFailure    int `json:"match_data_handler_read_packet_failure"`
	MatchDataHandlerBuyerNotFound        int `json:"match_data_handler_buyer_not_found"`
	MatchDataHandlerBuyerNotActive       int `json:"match_data_handler_buyer_not_active"`
	MatchDataHandlerSignatureCheckFailed int `json:"match_data_handler_signature_check_failed"`
	MatchDataHandlerWriteResponseFailure int `json:"match_data_handler_write_response_failure"`

	// Post Session Metrics
	PostSessionBillingEntries2Sent        int `json:"post_session_billing_entries_2_sent"`
	PostSessionBillingEntries2Finished    int `json:"post_session_billing_entires_2_finished"`
	PostSessionBilling2BufferFull         int `json:"post_session_billing_2_buffer_full"`
	PostSessionPortalEntriesSent          int `json:"post_session_portal_entries_sent"`
	PostSessionPortalEntriesFinished      int `json:"post_session_portal_entries_finished"`
	PostSessionPortalBufferFull           int `json:"post_session_portal_buffer_full"`
	PostSessionMatchDataEntriesSent       int `json:"post_session_match_data_entries_sent"`
	PostSessionMatchDataEntriesFinished   int `json:"post_session_match_data_entries_finished"`
	PostSessionMatchDataEntriesBufferFull int `json:"post_session_match_data_entries_buffer_full"`
	PostSessionBilling2Failure            int `json:"post_session_billing_2_failure"`
	PostSessionPortalFailure              int `json:"post_session_portal_failure"`
	PostSessionMatchDataEntriesFailure    int `json:"post_session_match_data_entries_failure"`

	// Billing Metrics
	BillingEntries2Submitted int `json:"billing_entries_2_submitted"`
	BillingEntries2Queued    int `json:"billing_entries_2_queued"`
	BillingEntries2Flushed   int `json:"billing_entries_2_flushed"`
	Billing2PublishFailure   int `json:"billing_2_publish_failure"`

	// Match Data Metrics
	MatchDataEntriesSubmitted      int `json:"match_data_entries_submitted"`
	MatchDataEntriesQueued         int `json:"match_data_entries_queued"`
	MatchDataEntriesFlushed        int `json:"match_data_entries_flushed"`
	MatchDataEntriesPublishFailure int `json:"match_data_entries_publish_failure"`

	// Route Matrix Metrics
	RouteMatrixNumRoutes int `json:"route_matrix_num_routes"`
	RouteMatrixBytes     int `json:"route_matrix_bytes"`

	// Error Metrics
	RouteMatrixReaderNil        int `json:"route_matrix_reader_nil"`
	RouteMatrixReadFailure      int `json:"route_matrix_read_failure"`
	RouteMatrixBufferEmpty      int `json:"route_matrix_buffer_empty"`
	RouteMatrixSerializeFailure int `json:"route_matrix_serialize_failure"`
	BinWrapperEmpty             int `json:"bin_wrapper_empty"`
	BinWrapperFailure           int `json:"bin_wrapper_failure"`
	StaleRouteMatrix            int `json:"stale_route_matrix"`
}

// ServerInitMetrics defines the set of metrics for the server init handler in the server backend.
type ServerInitMetrics struct {
	HandlerMetrics *PacketHandlerMetrics

	ReadPacketFailure     Counter
	BuyerNotFound         Counter
	BuyerNotActive        Counter
	SignatureCheckFailed  Counter
	SDKTooOld             Counter
	DatacenterMapNotFound Counter
	DatacenterNotFound    Counter
	WriteResponseFailure  Counter
}

// EmptyServerInitMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyServerInitMetrics = ServerInitMetrics{
	HandlerMetrics:        &EmptyPacketHandlerMetrics,
	ReadPacketFailure:     &EmptyCounter{},
	BuyerNotFound:         &EmptyCounter{},
	BuyerNotActive:        &EmptyCounter{},
	SignatureCheckFailed:  &EmptyCounter{},
	SDKTooOld:             &EmptyCounter{},
	DatacenterMapNotFound: &EmptyCounter{},
	DatacenterNotFound:    &EmptyCounter{},
	WriteResponseFailure:  &EmptyCounter{},
}

// ServerUpdateMetrics defines the set of metrics for the server update handler in the server backend.
type ServerUpdateMetrics struct {
	HandlerMetrics *PacketHandlerMetrics

	ReadPacketFailure      Counter
	BuyerNotFound          Counter
	BuyerNotLive           Counter
	SignatureCheckFailed   Counter
	SDKTooOld              Counter
	DatacenterMapNotFound  Counter
	DatacenterNotFound     Counter
	ServerUpdatePacketSize Gauge
}

// EmptyServerUpdateMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyServerUpdateMetrics = ServerUpdateMetrics{
	HandlerMetrics:         &EmptyPacketHandlerMetrics,
	ReadPacketFailure:      &EmptyCounter{},
	BuyerNotFound:          &EmptyCounter{},
	BuyerNotLive:           &EmptyCounter{},
	SignatureCheckFailed:   &EmptyCounter{},
	SDKTooOld:              &EmptyCounter{},
	DatacenterMapNotFound:  &EmptyCounter{},
	DatacenterNotFound:     &EmptyCounter{},
	ServerUpdatePacketSize: &EmptyGauge{},
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
	DatacenterNotEnabled                       Counter
	NearRelaysLocateFailure                    Counter
	NearRelaysChanged                          Counter
	NoRelaysInDatacenter                       Counter
	RouteDoesNotExist                          Counter
	RouteSwitched                              Counter
	NextWithoutRouteRelays                     Counter
	SDKAborted                                 Counter
	NoRoute                                    Counter
	MultipathOverload                          Counter
	LatencyWorse                               Counter
	MispredictVeto                             Counter
	WriteResponseFailure                       Counter
	StaleRouteMatrix                           Counter
	SessionUpdatePacketSize                    Gauge
	NextSessionResponsePacketSize              Gauge
	DirectSessionResponsePacketSize            Gauge
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
	DatacenterNotEnabled:                       &EmptyCounter{},
	NearRelaysLocateFailure:                    &EmptyCounter{},
	NearRelaysChanged:                          &EmptyCounter{},
	NoRelaysInDatacenter:                       &EmptyCounter{},
	RouteDoesNotExist:                          &EmptyCounter{},
	RouteSwitched:                              &EmptyCounter{},
	NextWithoutRouteRelays:                     &EmptyCounter{},
	SDKAborted:                                 &EmptyCounter{},
	NoRoute:                                    &EmptyCounter{},
	MultipathOverload:                          &EmptyCounter{},
	LatencyWorse:                               &EmptyCounter{},
	MispredictVeto:                             &EmptyCounter{},
	WriteResponseFailure:                       &EmptyCounter{},
	StaleRouteMatrix:                           &EmptyCounter{},
	SessionUpdatePacketSize:                    &EmptyGauge{},
	NextSessionResponsePacketSize:              &EmptyGauge{},
	DirectSessionResponsePacketSize:            &EmptyGauge{},
}

// MatchDataHandlerMetrics defines the set of metrics for the match data handler in the server backend.
type MatchDataHandlerMetrics struct {
	HandlerMetrics *PacketHandlerMetrics

	ReadPacketFailure    Counter
	BuyerNotFound        Counter
	BuyerNotActive       Counter
	SignatureCheckFailed Counter
	WriteResponseFailure Counter
}

// EmptyMatchDataHandlerMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyMatchDataHandlerMetrics = MatchDataHandlerMetrics{
	HandlerMetrics:       &EmptyPacketHandlerMetrics,
	ReadPacketFailure:    &EmptyCounter{},
	BuyerNotFound:        &EmptyCounter{},
	BuyerNotActive:       &EmptyCounter{},
	SignatureCheckFailed: &EmptyCounter{},
	WriteResponseFailure: &EmptyCounter{},
}

// ServerBackendMetrics defines the set of metrics for the server backend.
type ServerBackendMetrics struct {
	ServiceMetrics *ServiceMetrics

	ServerInitMetrics       *ServerInitMetrics
	ServerUpdateMetrics     *ServerUpdateMetrics
	SessionUpdateMetrics    *SessionUpdateMetrics
	MatchDataHandlerMetrics *MatchDataHandlerMetrics

	PostSessionMetrics *PostSessionMetrics

	BillingMetrics   *BillingMetrics
	MatchDataMetrics *MatchDataMetrics

	RouteMatrixUpdateDuration     Gauge
	RouteMatrixUpdateLongDuration Counter
	RouteMatrixNumRoutes          Gauge
	RouteMatrixBytes              Gauge

	ErrorMetrics *ServerBackendErrorMetrics
}

// EmptyServerBackendMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyServerBackendMetrics = ServerBackendMetrics{
	ServiceMetrics:                &EmptyServiceMetrics,
	ServerInitMetrics:             &EmptyServerInitMetrics,
	ServerUpdateMetrics:           &EmptyServerUpdateMetrics,
	SessionUpdateMetrics:          &EmptySessionUpdateMetrics,
	PostSessionMetrics:            &EmptyPostSessionMetrics,
	BillingMetrics:                &EmptyBillingMetrics,
	MatchDataMetrics:              &EmptyMatchDataMetrics,
	RouteMatrixUpdateDuration:     &EmptyGauge{},
	RouteMatrixUpdateLongDuration: &EmptyCounter{},
	RouteMatrixNumRoutes:          &EmptyGauge{},
	RouteMatrixBytes:              &EmptyGauge{},
	ErrorMetrics:                  &EmptyServerBackendErrorMetrics,
}

type ServerBackendErrorMetrics struct {
	RouteMatrixReaderNil        Counter
	RouteMatrixReadFailure      Counter
	RouteMatrixBufferEmpty      Counter
	RouteMatrixSerializeFailure Counter
	BinWrapperEmpty             Counter
	BinWrapperFailure           Counter
	StaleRouteMatrix            Counter
}

var EmptyServerBackendErrorMetrics = ServerBackendErrorMetrics{
	RouteMatrixReaderNil:        &EmptyCounter{},
	RouteMatrixReadFailure:      &EmptyCounter{},
	RouteMatrixBufferEmpty:      &EmptyCounter{},
	RouteMatrixSerializeFailure: &EmptyCounter{},
	BinWrapperEmpty:             &EmptyCounter{},
	BinWrapperFailure:           &EmptyCounter{},
	StaleRouteMatrix:            &EmptyCounter{},
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

	m.MatchDataHandlerMetrics, err = newMatchDataHandlerMetrics(ctx, handler, serviceName, "match_data", "Match Data", "match data entry")
	if err != nil {
		return nil, err
	}

	m.PostSessionMetrics, err = NewPostSessionMetrics(ctx, handler, serviceName)
	if err != nil {
		return nil, err
	}

	m.BillingMetrics = &BillingMetrics{}
	m.BillingMetrics.Entries2Received = &EmptyCounter{}

	m.BillingMetrics.Entries2Submitted, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Server Backend Billing Entries 2 Submitted",
		ServiceName: serviceName,
		ID:          "billing.entries_submitted_2",
		Unit:        "entries",
		Description: "The number of billing entries 2 the server_backend has submitted to be published",
	})
	if err != nil {
		return nil, err
	}

	m.BillingMetrics.Entries2Queued, err = handler.NewGauge(ctx, &Descriptor{
		DisplayName: "Server Backend Billing Entries 2 Queued",
		ServiceName: serviceName,
		ID:          "billing.entries_queued_2",
		Unit:        "entries",
		Description: "The number of billing entries 2 the server_backend has queued waiting to be published",
	})
	if err != nil {
		return nil, err
	}

	m.BillingMetrics.Entries2Flushed, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Server Backend Billing Entries 2 Flushed",
		ServiceName: serviceName,
		ID:          "billing.entries_flushed_2",
		Unit:        "entries",
		Description: "The number of billing entries 2 the server_backend has flushed after publishing",
	})
	if err != nil {
		return nil, err
	}

	m.BillingMetrics.SummaryEntries2Submitted, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Server Backend Billing Summary Entries 2 Submitted",
		ServiceName: serviceName,
		ID:          "billing.summary_entries_submitted_2",
		Unit:        "entries",
		Description: "The number of billing summary entries 2 the server_backend has submitted to be published",
	})
	if err != nil {
		return nil, err
	}

	m.BillingMetrics.SummaryEntries2Queued = &EmptyGauge{}
	m.BillingMetrics.SummaryEntries2Flushed = &EmptyCounter{}

	m.BillingMetrics.ErrorMetrics.Billing2PublishFailure, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Server Backend Billing 2 Publish Failure",
		ServiceName: serviceName,
		ID:          "billing.publish_failure_2",
		Unit:        "entries",
		Description: "The number of billing entries 2 the server_backend has failed to publish",
	})
	if err != nil {
		return nil, err
	}

	m.BillingMetrics.ErrorMetrics.Billing2BatchedReadFailure = &EmptyCounter{}
	m.BillingMetrics.ErrorMetrics.Billing2ReadFailure = &EmptyCounter{}
	m.BillingMetrics.ErrorMetrics.Billing2WriteFailure = &EmptyCounter{}

	m.BillingMetrics.BillingEntry2Size, err = handler.NewGauge(ctx, &Descriptor{
		DisplayName: "Billing Entry Size",
		ServiceName: "billing",
		ID:          "billing.entry.2.size",
		Unit:        "bytes",
		Description: "The size of a billing entry 2",
	})
	if err != nil {
		return nil, err
	}

	m.BillingMetrics.PubsubBillingEntry2Size, err = handler.NewGauge(ctx, &Descriptor{
		DisplayName: "Pubsub Billing Entry 2 Size",
		ServiceName: "billing",
		ID:          "pubsub.billing.entry.2.size",
		Unit:        "bytes",
		Description: "The size of a pubsub billing entry 2",
	})
	if err != nil {
		return nil, err
	}

	m.MatchDataMetrics, err = NewMatchDataMetrics(ctx, handler, serviceName, "match_data", "Server Backend Match Data", "match data entry")
	if err != nil {
		return nil, err
	}
	m.MatchDataMetrics.EntriesReceived = &EmptyCounter{}
	m.MatchDataMetrics.ErrorMetrics.MatchDataReadFailure = &EmptyCounter{}
	m.MatchDataMetrics.ErrorMetrics.MatchDataBatchedReadFailure = &EmptyCounter{}
	m.MatchDataMetrics.ErrorMetrics.MatchDataWriteFailure = &EmptyCounter{}
	m.MatchDataMetrics.ErrorMetrics.MatchDataInvalidEntries = &EmptyCounter{}
	m.MatchDataMetrics.ErrorMetrics.MatchDataEntriesWithNaN = &EmptyCounter{}
	m.MatchDataMetrics.ErrorMetrics.MatchDataRetryLimitReached = &EmptyCounter{}

	// used: entries submitted, queued, flushed, publish failure, entry size, pubsub entry size

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

	m.ErrorMetrics, err = newServerBackendErrorMetrics(ctx, handler, serviceName)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func newServerBackendErrorMetrics(ctx context.Context, handler Handler, serviceName string) (*ServerBackendErrorMetrics, error) {
	var err error
	m := &ServerBackendErrorMetrics{}

	m.RouteMatrixReaderNil, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Server Backend Route Matrix Reader Nil",
		ServiceName: serviceName,
		ID:          "server_backend.route_matrix_reader_nil",
		Unit:        "errors",
		Description: "The number of times the " + serviceName + "'s route matrix reader was nil.",
	})
	if err != nil {
		return nil, err
	}

	m.RouteMatrixReadFailure, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Server Backend Route Matrix Read Failure",
		ServiceName: serviceName,
		ID:          "server_backend.route_matrix_read_failure",
		Unit:        "errors",
		Description: "The number of times the " + serviceName + " failed to read the route matrix data.",
	})
	if err != nil {
		return nil, err
	}

	m.RouteMatrixBufferEmpty, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Server Backend Route Matrix Buffer Empty",
		ServiceName: serviceName,
		ID:          "server_backend.route_matrix_buffer_empty",
		Unit:        "errors",
		Description: "The number of times the " + serviceName + "'s route matrix buffer was empty.",
	})
	if err != nil {
		return nil, err
	}

	m.RouteMatrixSerializeFailure, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Server Backend Route Matrix Serialize Failure",
		ServiceName: serviceName,
		ID:          "server_backend.route_matrix_serialize_failure",
		Unit:        "errors",
		Description: "The number of times the " + serviceName + " failed to serialize the route matrix.",
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
		ID:          "server_backend.route_matrix_stale",
		Unit:        "count",
		Description: "The number of times the " + serviceName + " received a stale route matrix.",
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

	m.BuyerNotLive, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Buyer Not Live",
		ServiceName: serviceName,
		ID:          handlerID + ".buyer_not_live",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " contained a customer ID that was not live.",
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

	m.ServerUpdatePacketSize, err = handler.NewGauge(ctx, &Descriptor{
		DisplayName: handlerName + " Server Update Packet Size",
		ServiceName: serviceName,
		ID:          handlerID + ".server_update_packet_size",
		Unit:        "bytes",
		Description: "The size of a server update packet",
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

	m.DatacenterNotEnabled, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Datacenter Not Enabled",
		ServiceName: serviceName,
		ID:          handlerID + ".datacenter_not_enabled",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " contained a datacenter ID that was not enabled.",
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

	m.NextWithoutRouteRelays, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Next Without Route Relays",
		ServiceName: serviceName,
		ID:          handlerID + ".next_without_route_relays",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " was on next without any route relays.",
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

	m.StaleRouteMatrix, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Stale Route Matrix",
		ServiceName: serviceName,
		ID:          handlerID + ".stale_route_matrix",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " was using a stale route matrix.",
	})
	if err != nil {
		return nil, err
	}

	m.SessionUpdatePacketSize, err = handler.NewGauge(ctx, &Descriptor{
		DisplayName: handlerName + " Session Update Packet Size",
		ServiceName: serviceName,
		ID:          handlerID + ".session_update_packet_size",
		Unit:        "bytes",
		Description: "The size of the incoming session update packet.",
	})
	if err != nil {
		return nil, err
	}

	m.DirectSessionResponsePacketSize, err = handler.NewGauge(ctx, &Descriptor{
		DisplayName: handlerName + " Direct Session Response Packet Size",
		ServiceName: serviceName,
		ID:          handlerID + ".direct_session_response_packet_size",
		Unit:        "bytes",
		Description: "The size of the incoming direct session response packet.",
	})
	if err != nil {
		return nil, err
	}

	m.NextSessionResponsePacketSize, err = handler.NewGauge(ctx, &Descriptor{
		DisplayName: handlerName + " Next Session Response Packet Size",
		ServiceName: serviceName,
		ID:          handlerID + ".next_session_response_packet_size",
		Unit:        "bytes",
		Description: "The size of the incoming next session response packet.",
	})
	if err != nil {
		return nil, err
	}

	return m, nil
}

func newMatchDataHandlerMetrics(ctx context.Context, handler Handler, serviceName string, handlerID string, handlerName string, packetDescription string) (*MatchDataHandlerMetrics, error) {
	var err error
	m := &MatchDataHandlerMetrics{}

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
