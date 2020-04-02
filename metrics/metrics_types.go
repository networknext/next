package metrics

import (
	"context"
)

type DecisionMetrics struct {
	NoChange            Counter
	ForceDirect         Counter
	ForceNext           Counter
	NoNextRoute         Counter
	ABTestDirect        Counter
	RTTReduction        Counter
	PacketLossMultipath Counter
	JitterMultipath     Counter
	VetoRTT             Counter
	RTTMultipath        Counter
	VetoPacketLoss      Counter
	FallbackToDirect    Counter
	VetoYOLO            Counter
	VetoNoRoute         Counter
	InitialSlice        Counter
	VetoRTTYOLO         Counter
	VetoPacketLossYOLO  Counter
	RTTIncrease         Counter
}

var EmptyDecisionMetrics DecisionMetrics = DecisionMetrics{
	NoChange:            &EmptyCounter{},
	ForceDirect:         &EmptyCounter{},
	ForceNext:           &EmptyCounter{},
	NoNextRoute:         &EmptyCounter{},
	ABTestDirect:        &EmptyCounter{},
	RTTReduction:        &EmptyCounter{},
	PacketLossMultipath: &EmptyCounter{},
	JitterMultipath:     &EmptyCounter{},
	VetoRTT:             &EmptyCounter{},
	RTTMultipath:        &EmptyCounter{},
	VetoPacketLoss:      &EmptyCounter{},
	FallbackToDirect:    &EmptyCounter{},
	VetoYOLO:            &EmptyCounter{},
	VetoNoRoute:         &EmptyCounter{},
	InitialSlice:        &EmptyCounter{},
	VetoRTTYOLO:         &EmptyCounter{},
	VetoPacketLossYOLO:  &EmptyCounter{},
	RTTIncrease:         &EmptyCounter{},
}

type SessionErrorMetrics struct {
	ReadPacketFailure           Counter
	FallbackToDirect            Counter
	PipelineExecFailure         Counter
	GetServerDataFailure        Counter
	UnmarshalServerDataFailure  Counter
	GetSessionDataFailure       Counter
	UnmarshalSessionDataFailure Counter
	BuyerNotFound               Counter
	VerifyFailure               Counter
	OldSequence                 Counter
	WriteCachedResponseFailure  Counter
	ClientLocateFailure         Counter
	NearRelaysLocateFailure     Counter
	RouteFailure                Counter
	EncryptionFailure           Counter
	WriteResponseFailure        Counter
	UpdateSessionFailure        Counter
	BillingFailure              Counter
}

var EmptySessionErrorMetrics SessionErrorMetrics = SessionErrorMetrics{
	ReadPacketFailure:           &EmptyCounter{},
	FallbackToDirect:            &EmptyCounter{},
	PipelineExecFailure:         &EmptyCounter{},
	GetServerDataFailure:        &EmptyCounter{},
	UnmarshalServerDataFailure:  &EmptyCounter{},
	GetSessionDataFailure:       &EmptyCounter{},
	UnmarshalSessionDataFailure: &EmptyCounter{},
	BuyerNotFound:               &EmptyCounter{},
	VerifyFailure:               &EmptyCounter{},
	OldSequence:                 &EmptyCounter{},
	WriteCachedResponseFailure:  &EmptyCounter{},
	ClientLocateFailure:         &EmptyCounter{},
	NearRelaysLocateFailure:     &EmptyCounter{},
	RouteFailure:                &EmptyCounter{},
	EncryptionFailure:           &EmptyCounter{},
	WriteResponseFailure:        &EmptyCounter{},
	UpdateSessionFailure:        &EmptyCounter{},
	BillingFailure:              &EmptyCounter{},
}

type SessionMetrics struct {
	Invocations         Counter
	DirectSessions      Counter
	NextSessions        Counter
	DurationGauge       Gauge
	DecisionMetrics     DecisionMetrics
	SessionErrorMetrics SessionErrorMetrics
}

var EmptySessionMetrics SessionMetrics = SessionMetrics{
	Invocations:         &EmptyCounter{},
	DirectSessions:      &EmptyCounter{},
	NextSessions:        &EmptyCounter{},
	DurationGauge:       &EmptyGauge{},
	DecisionMetrics:     EmptyDecisionMetrics,
	SessionErrorMetrics: EmptySessionErrorMetrics,
}

type ServerUpdateMetrics struct {
	Invocations   Counter
	DurationGauge Gauge
}

var EmptyServerUpdateMetrics ServerUpdateMetrics = ServerUpdateMetrics{
	Invocations:   &EmptyCounter{},
	DurationGauge: &EmptyGauge{},
}

type RelayInitMetrics struct {
	Invocations   Counter
	DurationGauge Gauge
}

var EmptyRelayInitMetrics RelayInitMetrics = RelayInitMetrics{
	Invocations:   &EmptyCounter{},
	DurationGauge: &EmptyGauge{},
}

type RelayUpdateMetrics struct {
	Invocations   Counter
	DurationGauge Gauge
}

var EmptyRelayUpdateMetrics RelayUpdateMetrics = RelayUpdateMetrics{
	Invocations:   &EmptyCounter{},
	DurationGauge: &EmptyGauge{},
}

type RelayHandlerMetrics struct {
	Invocations   Counter
	DurationGauge Gauge
}

var EmptyRelayHandlerMetrics RelayHandlerMetrics = RelayHandlerMetrics{
	Invocations:   &EmptyCounter{},
	DurationGauge: &EmptyGauge{},
}

type RelayStatMetrics struct {
	NumRelays Gauge
	NumRoutes Gauge
}

var EmptyRelayStatMetrics RelayStatMetrics = RelayStatMetrics{
	NumRelays: &EmptyGauge{},
	NumRoutes: &EmptyGauge{},
}

func NewSessionMetrics(ctx context.Context, metricsHandler Handler) (*SessionMetrics, error) {
	var err error

	sessionMetrics := SessionMetrics{}

	sessionMetrics.Invocations, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Total session update invocations",
		ServiceName: "server_backend",
		ID:          "session.count",
		Unit:        "invocations",
		Description: "The total number of concurrent sessions",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.DirectSessions, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Total direct session count",
		ServiceName: "server_backend",
		ID:          "session.direct.count",
		Unit:        "sessions",
		Description: "The total number of sessions that are currently being served a direct route",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.NextSessions, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Total next session count",
		ServiceName: "server_backend",
		ID:          "session.next.count",
		Unit:        "sessions",
		Description: "The total number of sessions that are currently being served a network next route",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.DurationGauge, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Session update duration",
		ServiceName: "server_backend",
		ID:          "session.duration",
		Unit:        "milliseconds",
		Description: "How long it takes to process a session update request",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.DecisionMetrics.NoChange, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Decision no change",
		ServiceName: "server_backend",
		ID:          "session.route_decision.no_change",
		Unit:        "decisions",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.DecisionMetrics.ForceDirect, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Decision force direct",
		ServiceName: "server_backend",
		ID:          "session.route_decision.force_direct",
		Unit:        "decisions",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.DecisionMetrics.ForceNext, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Decision force next",
		ServiceName: "server_backend",
		ID:          "session.route_decision.force_next",
		Unit:        "decisions",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.DecisionMetrics.NoNextRoute, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Decision no next route",
		ServiceName: "server_backend",
		ID:          "session.route_decision.no_next_route",
		Unit:        "decisions",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.DecisionMetrics.ABTestDirect, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Decision AB test direct",
		ServiceName: "server_backend",
		ID:          "session.route_decision.ab_test_direct",
		Unit:        "decisions",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.DecisionMetrics.RTTReduction, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Decision RTT reduction",
		ServiceName: "server_backend",
		ID:          "session.route_decision.rtt_reduction",
		Unit:        "decisions",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.DecisionMetrics.PacketLossMultipath, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Decision packet loss multipath",
		ServiceName: "server_backend",
		ID:          "session.route_decision.packet_loss_multipath",
		Unit:        "decisions",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.DecisionMetrics.JitterMultipath, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Decision jitter multipath",
		ServiceName: "server_backend",
		ID:          "session.route_decision.jitter_multipath",
		Unit:        "decisions",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.DecisionMetrics.VetoRTT, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Decision veto RTT",
		ServiceName: "server_backend",
		ID:          "session.route_decision.veto_rtt",
		Unit:        "decisions",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.DecisionMetrics.RTTMultipath, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Decision RTT multipath",
		ServiceName: "server_backend",
		ID:          "session.route_decision.rtt_multipath",
		Unit:        "decisions",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.DecisionMetrics.VetoPacketLoss, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Decision veto packet loss",
		ServiceName: "server_backend",
		ID:          "session.route_decision.veto_packet_loss",
		Unit:        "decisions",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.DecisionMetrics.FallbackToDirect, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Decision fallback to direct",
		ServiceName: "server_backend",
		ID:          "session.route_decision.fallback_to_direct",
		Unit:        "decisions",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.DecisionMetrics.VetoYOLO, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Decision veto YOLO",
		ServiceName: "server_backend",
		ID:          "session.route_decision.veto_yolo",
		Unit:        "decisions",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.DecisionMetrics.VetoNoRoute, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Decision veto no route",
		ServiceName: "server_backend",
		ID:          "session.route_decision.veto_no_route",
		Unit:        "decisions",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.DecisionMetrics.InitialSlice, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Decision initial slice",
		ServiceName: "server_backend",
		ID:          "session.route_decision.initial_slice",
		Unit:        "decisions",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.DecisionMetrics.VetoRTTYOLO, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Decision veto RTT YOLO",
		ServiceName: "server_backend",
		ID:          "session.route_decision.veto_rtt_yolo",
		Unit:        "decisions",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.DecisionMetrics.VetoPacketLossYOLO, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Decision veto packet loss yolo",
		ServiceName: "server_backend",
		ID:          "session.route_decision.veto_packet_loss_yolo",
		Unit:        "decisions",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.DecisionMetrics.RTTIncrease, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Decision RTT increase",
		ServiceName: "server_backend",
		ID:          "session.route_decision.rtt_increase",
		Unit:        "decisions",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.SessionErrorMetrics.BillingFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Session Billing Failure",
		ServiceName: "server_backend",
		ID:          "session.error.billing_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.SessionErrorMetrics.BuyerNotFound, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Session Buyer Not Found",
		ServiceName: "server_backend",
		ID:          "session.error.buyer_not_found",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.SessionErrorMetrics.ClientLocateFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Session Client Locate Failure",
		ServiceName: "server_backend",
		ID:          "session.error.client_locate_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.SessionErrorMetrics.EncryptionFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Session Encryption Failure",
		ServiceName: "server_backend",
		ID:          "session.error.encryption_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.SessionErrorMetrics.FallbackToDirect, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Session Fallback To Direct",
		ServiceName: "server_backend",
		ID:          "session.error.fallback_to_direct",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.SessionErrorMetrics.GetServerDataFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Session Get Server Data Failure",
		ServiceName: "server_backend",
		ID:          "session.error.get_server_data_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.SessionErrorMetrics.GetSessionDataFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Session Get Session Data Failure",
		ServiceName: "server_backend",
		ID:          "session.error.get_session_data_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.SessionErrorMetrics.NearRelaysLocateFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Session Near Relays Locate Failure",
		ServiceName: "server_backend",
		ID:          "session.error.near_relays_locate_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.SessionErrorMetrics.OldSequence, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Session Old Sequence",
		ServiceName: "server_backend",
		ID:          "session.error.old_sequence",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.SessionErrorMetrics.PipelineExecFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Session Redis Pipeline Exec Failure",
		ServiceName: "server_backend",
		ID:          "session.error.redis_pipeline_exec_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.SessionErrorMetrics.ReadPacketFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Session Read Packet Failure",
		ServiceName: "server_backend",
		ID:          "session.error.read_packet_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.SessionErrorMetrics.RouteFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Session Route Failure",
		ServiceName: "server_backend",
		ID:          "session.error.route_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.SessionErrorMetrics.UnmarshalServerDataFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Session Unmarshal Server Data Failure",
		ServiceName: "server_backend",
		ID:          "session.error.unmarshal_server_data_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.SessionErrorMetrics.UnmarshalSessionDataFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Session Unmarshal Session Data Failure",
		ServiceName: "server_backend",
		ID:          "session.error.unmarshal_session_data_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.SessionErrorMetrics.UpdateSessionFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Session Update Session Failure",
		ServiceName: "server_backend",
		ID:          "session.error.update_session_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.SessionErrorMetrics.VerifyFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Session Verify Failure",
		ServiceName: "server_backend",
		ID:          "session.error.verify_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.SessionErrorMetrics.WriteCachedResponseFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Session Write Cached Response Failure",
		ServiceName: "server_backend",
		ID:          "session.error.write_cached_response_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.SessionErrorMetrics.WriteResponseFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Session Write Response Failure",
		ServiceName: "server_backend",
		ID:          "session.error.write_response_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	return &sessionMetrics, nil
}

func NewServerUpdateMetrics(ctx context.Context, metricsHandler Handler) (*ServerUpdateMetrics, error) {
	updateDurationGauge, err := metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Server update duration",
		ServiceName: "server_backend",
		ID:          "server.duration",
		Unit:        "milliseconds",
		Description: "How long it takes to process a server update request.",
	})
	if err != nil {
		return nil, err
	}

	updateInvocationsCounter, err := metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Total server update invocations",
		ServiceName: "server_backend",
		ID:          "server.count",
		Unit:        "invocations",
		Description: "The total number of concurrent servers",
	})
	if err != nil {
		return nil, err
	}

	updateMetrics := ServerUpdateMetrics{
		Invocations:   updateInvocationsCounter,
		DurationGauge: updateDurationGauge,
	}

	return &updateMetrics, nil
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
	}

	return &updateMetrics, nil
}

func NewRelayHandlerMetrics(ctx context.Context, metricsHandler Handler) (*RelayHandlerMetrics, error) {
	handlerCount, err := metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Total relay handler count",
		ServiceName: "relay_backend",
		ID:          "relay.handler.count",
		Unit:        "requests",
		Description: "The total number of received relay requests",
	})
	if err != nil {
		return nil, err
	}

	handlerDuration, err := metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Relay handler duration",
		ServiceName: "relay_backend",
		ID:          "relay.handler.duration",
		Unit:        "milliseconds",
		Description: "How long it takes to process a relay request",
	})
	if err != nil {
		return nil, err
	}

	handerMetrics := RelayHandlerMetrics{
		Invocations:   handlerCount,
		DurationGauge: handlerDuration,
	}

	return &handerMetrics, nil
}

func NewRelayStatMetrics(ctx context.Context, metricsHandler Handler) (*RelayStatMetrics, error) {
	numRelays, err := metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Relays num relays",
		ServiceName: "relay_backend",
		ID:          "relays.num.relays",
		Unit:        "relays",
		Description: "How many relays are active",
	})
	if err != nil {
		return nil, err
	}

	numRoutes, err := metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Route Matrix num routes",
		ServiceName: "relay_backend",
		ID:          "route.matrix.num.routes",
		Unit:        "routes",
		Description: "How many routes are being generated",
	})
	if err != nil {
		return nil, err
	}

	statMetrics := RelayStatMetrics{
		NumRelays: numRelays,
		NumRoutes: numRoutes,
	}

	return &statMetrics, nil
}
