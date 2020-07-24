package metrics

import (
	"context"
)

type SessionMetrics struct {
	Invocations     Counter
	DirectSessions  Counter
	NextSessions    Counter
	DurationGauge   Gauge
	LongDuration    Counter
	DecisionMetrics DecisionMetrics
	ErrorMetrics    SessionErrorMetrics
}

var EmptySessionMetrics SessionMetrics = SessionMetrics{
	Invocations:     &EmptyCounter{},
	DirectSessions:  &EmptyCounter{},
	NextSessions:    &EmptyCounter{},
	DurationGauge:   &EmptyGauge{},
	LongDuration:    &EmptyCounter{},
	DecisionMetrics: EmptyDecisionMetrics,
	ErrorMetrics:    EmptySessionErrorMetrics,
}

type SessionErrorMetrics struct {
	UnserviceableUpdate         Counter
	ReadPacketHeaderFailure     Counter
	ReadPacketFailure           Counter
	EarlyFallbackToDirect       Counter
	PipelineExecFailure         Counter
	ServerDataMissing           Counter
	GetServerDataFailure        Counter
	UnmarshalServerDataFailure  Counter
	GetSessionDataFailure       Counter
	UnmarshalSessionDataFailure Counter
	GetVetoDataFailure          Counter
	UnmarshalVetoDataFailure    Counter
	BuyerNotFound               Counter
	VerifyFailure               Counter
	OldSequence                 Counter
	WriteCachedResponseFailure  Counter
	ClientLocateFailure         Counter
	ClientIPAnonymizeFailure    Counter
	NearRelaysLocateFailure     Counter
	DatacenterDisabled          Counter
	NoRelaysInDatacenter        Counter
	RouteFailure                Counter
	RouteSelectFailure          Counter
	EncryptionFailure           Counter
	MarshalResponseFailure      Counter
	WriteResponseFailure        Counter
	UpdateCacheFailure          Counter
	BillingFailure              Counter
	UpdatePortalFailure         Counter
}

var EmptySessionErrorMetrics SessionErrorMetrics = SessionErrorMetrics{
	UnserviceableUpdate:         &EmptyCounter{},
	ReadPacketFailure:           &EmptyCounter{},
	ReadPacketHeaderFailure:     &EmptyCounter{},
	EarlyFallbackToDirect:       &EmptyCounter{},
	PipelineExecFailure:         &EmptyCounter{},
	ServerDataMissing:           &EmptyCounter{},
	GetServerDataFailure:        &EmptyCounter{},
	UnmarshalServerDataFailure:  &EmptyCounter{},
	GetSessionDataFailure:       &EmptyCounter{},
	UnmarshalSessionDataFailure: &EmptyCounter{},
	GetVetoDataFailure:          &EmptyCounter{},
	UnmarshalVetoDataFailure:    &EmptyCounter{},
	BuyerNotFound:               &EmptyCounter{},
	VerifyFailure:               &EmptyCounter{},
	OldSequence:                 &EmptyCounter{},
	WriteCachedResponseFailure:  &EmptyCounter{},
	ClientLocateFailure:         &EmptyCounter{},
	ClientIPAnonymizeFailure:    &EmptyCounter{},
	NearRelaysLocateFailure:     &EmptyCounter{},
	DatacenterDisabled:          &EmptyCounter{},
	NoRelaysInDatacenter:        &EmptyCounter{},
	RouteFailure:                &EmptyCounter{},
	RouteSelectFailure:          &EmptyCounter{},
	EncryptionFailure:           &EmptyCounter{},
	MarshalResponseFailure:      &EmptyCounter{},
	WriteResponseFailure:        &EmptyCounter{},
	UpdateCacheFailure:          &EmptyCounter{},
	BillingFailure:              &EmptyCounter{},
	UpdatePortalFailure:         &EmptyCounter{},
}

type DecisionMetrics struct {
	NoReason            Counter
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
	InitialSlice        Counter
	VetoRTTYOLO         Counter
	VetoPacketLossYOLO  Counter
	RTTHysteresis       Counter
	VetoCommit          Counter
	BuyerNotLive        Counter
}

var EmptyDecisionMetrics DecisionMetrics = DecisionMetrics{
	NoReason:            &EmptyCounter{},
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
	InitialSlice:        &EmptyCounter{},
	VetoRTTYOLO:         &EmptyCounter{},
	VetoPacketLossYOLO:  &EmptyCounter{},
	RTTHysteresis:       &EmptyCounter{},
	VetoCommit:          &EmptyCounter{},
	BuyerNotLive:        &EmptyCounter{},
}

type OptimizeMetrics struct {
	Invocations     Counter
	DurationGauge   Gauge
	LongUpdateCount Counter
}

var EmptyOptimizeMetrics OptimizeMetrics = OptimizeMetrics{
	Invocations:     &EmptyCounter{},
	DurationGauge:   &EmptyGauge{},
	LongUpdateCount: &EmptyCounter{},
}

type ServerInitMetrics struct {
	Invocations   Counter
	DurationGauge Gauge
	LongDuration  Counter
	ErrorMetrics  ServerInitErrorMetrics
}

var EmptyServerInitMetrics ServerInitMetrics = ServerInitMetrics{
	Invocations:   &EmptyCounter{},
	DurationGauge: &EmptyGauge{},
	LongDuration:  &EmptyCounter{},
	ErrorMetrics:  EmptyServerInitErrorMetrics,
}

type ServerInitErrorMetrics struct {
	ReadPacketFailure    Counter
	SDKTooOld            Counter
	BuyerNotFound        Counter
	VerificationFailure  Counter
	DatacenterNotFound   Counter
	WriteResponseFailure Counter
}

var EmptyServerInitErrorMetrics ServerInitErrorMetrics = ServerInitErrorMetrics{
	ReadPacketFailure:    &EmptyCounter{},
	SDKTooOld:            &EmptyCounter{},
	BuyerNotFound:        &EmptyCounter{},
	DatacenterNotFound:   &EmptyCounter{},
	VerificationFailure:  &EmptyCounter{},
	WriteResponseFailure: &EmptyCounter{},
}

type ServerUpdateMetrics struct {
	Invocations   Counter
	DurationGauge Gauge
	LongDuration  Counter
	ErrorMetrics  ServerUpdateErrorMetrics
}

var EmptyServerUpdateMetrics ServerUpdateMetrics = ServerUpdateMetrics{
	Invocations:   &EmptyCounter{},
	DurationGauge: &EmptyGauge{},
	LongDuration:  &EmptyCounter{},
	ErrorMetrics:  EmptyServerUpdateErrorMetrics,
}

type ServerUpdateErrorMetrics struct {
	UnserviceableUpdate  Counter
	ReadPacketFailure    Counter
	SDKTooOld            Counter
	BuyerNotFound        Counter
	DatacenterNotFound   Counter
	VerificationFailure  Counter
	PacketSequenceTooOld Counter
}

var EmptyServerUpdateErrorMetrics ServerUpdateErrorMetrics = ServerUpdateErrorMetrics{
	UnserviceableUpdate:  &EmptyCounter{},
	ReadPacketFailure:    &EmptyCounter{},
	SDKTooOld:            &EmptyCounter{},
	BuyerNotFound:        &EmptyCounter{},
	DatacenterNotFound:   &EmptyCounter{},
	VerificationFailure:  &EmptyCounter{},
	PacketSequenceTooOld: &EmptyCounter{},
}

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
	RedisFailure       Counter
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
	RedisFailure:       &EmptyCounter{},
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
	UnmarshalFailure      Counter
	InvalidVersion        Counter
	ExceedMaxRelays       Counter
	RedisFailure          Counter
	RelayNotFound         Counter
	RelayUnmarshalFailure Counter
	InvalidToken          Counter
	RelayNotEnabled       Counter
}

var EmptyRelayUpdateErrorMetrics RelayUpdateErrorMetrics = RelayUpdateErrorMetrics{
	UnmarshalFailure:      &EmptyCounter{},
	InvalidVersion:        &EmptyCounter{},
	ExceedMaxRelays:       &EmptyCounter{},
	RedisFailure:          &EmptyCounter{},
	RelayNotFound:         &EmptyCounter{},
	RelayUnmarshalFailure: &EmptyCounter{},
	InvalidToken:          &EmptyCounter{},
	RelayNotEnabled:       &EmptyCounter{},
}

type RelayHandlerMetrics struct {
	Invocations   Counter
	DurationGauge Gauge
	ErrorMetrics  RelayHandlerErrorMetrics
}

var EmptyRelayHandlerMetrics RelayHandlerMetrics = RelayHandlerMetrics{
	Invocations:   &EmptyCounter{},
	DurationGauge: &EmptyGauge{},
	ErrorMetrics:  EmptyRelayHandlerErrorMetrics,
}

type RelayHandlerErrorMetrics struct {
	UnmarshalFailure      Counter
	ExceedMaxRelays       Counter
	RelayNotFound         Counter
	RelayQuarantined      Counter
	NoAuthHeader          Counter
	BadAuthHeaderLength   Counter
	BadAuthHeaderToken    Counter
	BadNonce              Counter
	BadEncryptedAddress   Counter
	DecryptFailure        Counter
	RedisFailure          Counter
	RelayUnmarshalFailure Counter
}

var EmptyRelayHandlerErrorMetrics RelayHandlerErrorMetrics = RelayHandlerErrorMetrics{
	UnmarshalFailure:      &EmptyCounter{},
	ExceedMaxRelays:       &EmptyCounter{},
	RelayNotFound:         &EmptyCounter{},
	RelayQuarantined:      &EmptyCounter{},
	NoAuthHeader:          &EmptyCounter{},
	BadAuthHeaderLength:   &EmptyCounter{},
	BadAuthHeaderToken:    &EmptyCounter{},
	BadNonce:              &EmptyCounter{},
	BadEncryptedAddress:   &EmptyCounter{},
	DecryptFailure:        &EmptyCounter{},
	RedisFailure:          &EmptyCounter{},
	RelayUnmarshalFailure: &EmptyCounter{},
}

type RelayStatMetrics struct {
	NumRelays Gauge
	NumRoutes Gauge
}

var EmptyRelayStatMetrics RelayStatMetrics = RelayStatMetrics{
	NumRelays: &EmptyGauge{},
	NumRoutes: &EmptyGauge{},
}

type CostMatrixMetrics struct {
	Invocations     Counter
	DurationGauge   Gauge
	LongUpdateCount Counter
	Bytes           Gauge
	ErrorMetrics    CostMatrixErrorMetrics
}

var EmptyCostMatrixMetrics CostMatrixMetrics = CostMatrixMetrics{
	Invocations:     &EmptyCounter{},
	DurationGauge:   &EmptyGauge{},
	LongUpdateCount: &EmptyCounter{},
	Bytes:           &EmptyGauge{},
	ErrorMetrics:    EmptyCostMatrixErrorMetrics,
}

type CostMatrixErrorMetrics struct {
	GenFailure Counter
}

var EmptyCostMatrixErrorMetrics CostMatrixErrorMetrics = CostMatrixErrorMetrics{
	GenFailure: &EmptyCounter{},
}

type MaxmindSyncMetrics struct {
	Invocations   Counter
	DurationGauge Gauge
	ErrorMetrics  MaxmindSyncErrorMetrics
}

var EmptyMaxmindSyncMetrics MaxmindSyncMetrics = MaxmindSyncMetrics{
	Invocations:   &EmptyCounter{},
	DurationGauge: &EmptyGauge{},
}

type MaxmindSyncErrorMetrics struct {
	FailedToSync    Counter
	FailedToSyncISP Counter
}

var EmptyMaxmindSyncErrorMetrics MaxmindSyncErrorMetrics = MaxmindSyncErrorMetrics{
	FailedToSync:    &EmptyCounter{},
	FailedToSyncISP: &EmptyCounter{},
}

type BillingMetrics struct {
	Goroutines       Gauge
	MemoryAllocated  Gauge
	EntriesReceived  Counter
	EntriesSubmitted Counter
	EntriesQueued    Counter
	EntriesFlushed   Counter
	ErrorMetrics     BillingErrorMetrics
}

var EmptyBillingMetrics BillingMetrics = BillingMetrics{
	Goroutines:       &EmptyGauge{},
	MemoryAllocated:  &EmptyGauge{},
	EntriesReceived:  &EmptyCounter{},
	EntriesSubmitted: &EmptyCounter{},
	EntriesQueued:    &EmptyCounter{},
	EntriesFlushed:   &EmptyCounter{},
	ErrorMetrics:     EmptyBillingErrorMetrics,
}

type BillingErrorMetrics struct {
	BillingPublishFailure Counter
	BillingReadFailure    Counter
	BillingWriteFailure   Counter
}

var EmptyBillingErrorMetrics BillingErrorMetrics = BillingErrorMetrics{
	BillingPublishFailure: &EmptyCounter{},
	BillingReadFailure:    &EmptyCounter{},
	BillingWriteFailure:   &EmptyCounter{},
}

type RelayBackendMetrics struct {
	Goroutines      Gauge
	MemoryAllocated Gauge
}

var EmptyRelayBackendMetrics RelayBackendMetrics = RelayBackendMetrics{
	Goroutines:      &EmptyGauge{},
	MemoryAllocated: &EmptyGauge{},
}

type RouteMatrixMetrics struct {
	DatacenterCount Gauge
	RelayCount      Gauge
	RouteCount      Gauge
	Bytes           Gauge
}

var EmptyRouteMatrixMetrics RouteMatrixMetrics = RouteMatrixMetrics{
	DatacenterCount: &EmptyGauge{},
	RelayCount:      &EmptyGauge{},
	RouteCount:      &EmptyGauge{},
	Bytes:           &EmptyGauge{},
}

type ServerBackendMetrics struct {
	SessionCount Gauge
}

var EmptyServerBackendMetrics ServerBackendMetrics = ServerBackendMetrics{
	SessionCount: &EmptyGauge{},
}

func NewServerBackendMetrics(ctx context.Context, metricsHandler Handler) (*ServerBackendMetrics, error) {
	var err error

	serverBackendMetrics := ServerBackendMetrics{}

	serverBackendMetrics.SessionCount, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Total session count",
		ServiceName: "server_backend",
		ID:          "server_backend.sessions",
		Unit:        "sessions",
		Description: "The total number of concurrent sessions",
	})
	if err != nil {
		return nil, err
	}

	return &serverBackendMetrics, nil
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

	sessionMetrics.LongDuration, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Long Session Update Durations",
		ServiceName: "server_backend",
		ID:          "session.long_durations",
		Unit:        "durations",
		Description: "The number of session update calls that took longer than 100ms to complete",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.DecisionMetrics.NoReason, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Decision no reason",
		ServiceName: "server_backend",
		ID:          "session.route_decision.no_reason",
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

	sessionMetrics.DecisionMetrics.RTTHysteresis, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Decision RTT hysteresis",
		ServiceName: "server_backend",
		ID:          "session.route_decision.rtt_hysteresis",
		Unit:        "decisions",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.DecisionMetrics.VetoCommit, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Decision veto commit",
		ServiceName: "server_backend",
		ID:          "session.route_decision.veto_commit",
		Unit:        "decisions",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.DecisionMetrics.BuyerNotLive, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Session Buyer Not Live",
		ServiceName: "server_backend",
		ID:          "session.route_decision.buyer_not_live",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.ErrorMetrics.UnserviceableUpdate, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Unserviceable Session Updates",
		ServiceName: "server_backend",
		ID:          "session.error.unserviceable_sessions",
		Unit:        "sessions",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.ErrorMetrics.BillingFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Session Billing Failure",
		ServiceName: "server_backend",
		ID:          "session.error.billing_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.ErrorMetrics.UpdatePortalFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Session Update Portal Failure",
		ServiceName: "server_backend",
		ID:          "session.error.update_portal_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.ErrorMetrics.BuyerNotFound, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Session Buyer Not Found",
		ServiceName: "server_backend",
		ID:          "session.error.buyer_not_found",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.ErrorMetrics.ClientLocateFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Session Client Locate Failure",
		ServiceName: "server_backend",
		ID:          "session.error.client_locate_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.ErrorMetrics.ClientIPAnonymizeFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Session Client IP Anonymize Failure",
		ServiceName: "server_backend",
		ID:          "session.error.client_ip_anonymize_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.ErrorMetrics.EncryptionFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Session Encryption Failure",
		ServiceName: "server_backend",
		ID:          "session.error.encryption_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.ErrorMetrics.EarlyFallbackToDirect, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Session Early Fallback To Direct",
		ServiceName: "server_backend",
		ID:          "session.error.early_fallback_to_direct",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.ErrorMetrics.ServerDataMissing, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Session Server Data Missing",
		ServiceName: "server_backend",
		ID:          "session.error.server_data_missing",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.ErrorMetrics.GetServerDataFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Session Get Server Data Failure",
		ServiceName: "server_backend",
		ID:          "session.error.get_server_data_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.ErrorMetrics.GetSessionDataFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Session Get Session Data Failure",
		ServiceName: "server_backend",
		ID:          "session.error.get_session_data_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.ErrorMetrics.GetVetoDataFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Session Get Veto Data Failure",
		ServiceName: "server_backend",
		ID:          "session.error.get_veto_data_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.ErrorMetrics.NearRelaysLocateFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Session Near Relays Locate Failure",
		ServiceName: "server_backend",
		ID:          "session.error.near_relays_locate_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.ErrorMetrics.DatacenterDisabled, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Session Datacenter Disabled",
		ServiceName: "server_backend",
		ID:          "session.error.datacenter_disabled",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.ErrorMetrics.NoRelaysInDatacenter, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Session No Relays In Datacenter",
		ServiceName: "server_backend",
		ID:          "session.error.no_relays_in_datacenter",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.ErrorMetrics.OldSequence, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Session Old Sequence",
		ServiceName: "server_backend",
		ID:          "session.error.old_sequence",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.ErrorMetrics.PipelineExecFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Session Redis Pipeline Exec Failure",
		ServiceName: "server_backend",
		ID:          "session.error.redis_pipeline_exec_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.ErrorMetrics.ReadPacketHeaderFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Session Read Packet Header Failure",
		ServiceName: "server_backend",
		ID:          "session.error.read_packet_header_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.ErrorMetrics.ReadPacketFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Session Read Packet Failure",
		ServiceName: "server_backend",
		ID:          "session.error.read_packet_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.ErrorMetrics.RouteFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Session Route Failure",
		ServiceName: "server_backend",
		ID:          "session.error.route_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.ErrorMetrics.RouteSelectFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Session Route Select Failure",
		ServiceName: "server_backend",
		ID:          "session.error.route_select_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.ErrorMetrics.UnmarshalServerDataFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Session Unmarshal Server Data Failure",
		ServiceName: "server_backend",
		ID:          "session.error.unmarshal_server_data_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.ErrorMetrics.UnmarshalSessionDataFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Session Unmarshal Session Data Failure",
		ServiceName: "server_backend",
		ID:          "session.error.unmarshal_session_data_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.ErrorMetrics.UnmarshalVetoDataFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Session Unmarshal Veto Data Failure",
		ServiceName: "server_backend",
		ID:          "session.error.unmarshal_veto_data_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.ErrorMetrics.UpdateCacheFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Session Update Cache Failure",
		ServiceName: "server_backend",
		ID:          "session.error.update_cache_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.ErrorMetrics.VerifyFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Session Verify Failure",
		ServiceName: "server_backend",
		ID:          "session.error.verify_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.ErrorMetrics.WriteCachedResponseFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Session Write Cached Response Failure",
		ServiceName: "server_backend",
		ID:          "session.error.write_cached_response_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.ErrorMetrics.MarshalResponseFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Session Marshal Response Failure",
		ServiceName: "server_backend",
		ID:          "session.error.marshal_response_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	sessionMetrics.ErrorMetrics.WriteResponseFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
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

func NewServerInitMetrics(ctx context.Context, metricsHandler Handler) (*ServerInitMetrics, error) {
	initMetrics := ServerInitMetrics{}
	var err error

	initMetrics.DurationGauge, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Server init duration",
		ServiceName: "server_backend",
		ID:          "server.init.duration",
		Unit:        "milliseconds",
		Description: "How long it takes to process a server init request.",
	})
	if err != nil {
		return nil, err
	}

	initMetrics.Invocations, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Total server init invocations",
		ServiceName: "server_backend",
		ID:          "server.init.count",
		Unit:        "invocations",
		Description: "The total number of concurrent servers",
	})
	if err != nil {
		return nil, err
	}

	initMetrics.LongDuration, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Long Server Init Durations",
		ServiceName: "server_backend",
		ID:          "server.init.long_durations",
		Unit:        "durations",
		Description: "The number of server init calls that took longer than 100ms to complete",
	})
	if err != nil {
		return nil, err
	}

	initMetrics.ErrorMetrics.ReadPacketFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Server Init Read Packet Failure",
		ServiceName: "server_backend",
		ID:          "server.init.error.read_packet_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	initMetrics.ErrorMetrics.SDKTooOld, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Server Init SDK Too Old",
		ServiceName: "server_backend",
		ID:          "server.init.error.sdk_too_old",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	initMetrics.ErrorMetrics.BuyerNotFound, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Server Init Buyer Not Found",
		ServiceName: "server_backend",
		ID:          "server.init.error.buyer_not_found",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	initMetrics.ErrorMetrics.VerificationFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Server Init Verification Failure",
		ServiceName: "server_backend",
		ID:          "server.init.error.verification_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	initMetrics.ErrorMetrics.DatacenterNotFound, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Server Init Datacenter Not Found",
		ServiceName: "server_backend",
		ID:          "server.init.error.datacenter_not_found",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	initMetrics.ErrorMetrics.WriteResponseFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Server Init Write Response Failure",
		ServiceName: "server_backend",
		ID:          "server.init.error.write_response_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	return &initMetrics, nil
}

func NewServerUpdateMetrics(ctx context.Context, metricsHandler Handler) (*ServerUpdateMetrics, error) {
	var err error

	updateMetrics := ServerUpdateMetrics{}

	updateMetrics.DurationGauge, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Server update duration",
		ServiceName: "server_backend",
		ID:          "server.duration",
		Unit:        "milliseconds",
		Description: "How long it takes to process a server update request.",
	})
	if err != nil {
		return nil, err
	}

	updateMetrics.Invocations, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Total server update invocations",
		ServiceName: "server_backend",
		ID:          "server.count",
		Unit:        "invocations",
		Description: "The total number of concurrent servers",
	})
	if err != nil {
		return nil, err
	}

	updateMetrics.LongDuration, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Long Server Update Durations",
		ServiceName: "server_backend",
		ID:          "server.long_durations",
		Unit:        "durations",
		Description: "The number of server update calls that took longer than 100ms to complete",
	})
	if err != nil {
		return nil, err
	}

	updateMetrics.ErrorMetrics.UnserviceableUpdate, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Unserviceable Server Updates",
		ServiceName: "server_backend",
		ID:          "server.error.unserviceable_servers",
		Unit:        "servers",
	})
	if err != nil {
		return nil, err
	}

	updateMetrics.ErrorMetrics.ReadPacketFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Server Update Read Packet Failure",
		ServiceName: "server_backend",
		ID:          "server.error.read_packet_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	updateMetrics.ErrorMetrics.SDKTooOld, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Server Update SDK Too Old",
		ServiceName: "server_backend",
		ID:          "server.error.sdk_too_old",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	updateMetrics.ErrorMetrics.BuyerNotFound, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Server Update Buyer Not Found",
		ServiceName: "server_backend",
		ID:          "server.error.buyer_not_found",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	updateMetrics.ErrorMetrics.DatacenterNotFound, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Server Update Datacenter Not Found",
		ServiceName: "server_backend",
		ID:          "server.error.datacenter_not_found",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	updateMetrics.ErrorMetrics.VerificationFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Server Update Verification Failure",
		ServiceName: "server_backend",
		ID:          "server.error.verification_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	updateMetrics.ErrorMetrics.PacketSequenceTooOld, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Server Update Packet Sequence Too Old",
		ServiceName: "server_backend",
		ID:          "server.error.packet_sequence_too_old",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
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
		ErrorMetrics:  EmptyRelayHandlerErrorMetrics,
	}

	return &handerMetrics, nil
}

func NewCostMatrixMetrics(ctx context.Context, metricsHandler Handler) (*CostMatrixMetrics, error) {
	costMatrixDurationGauge, err := metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "StatsDB -> GetCostMatrix duration",
		ServiceName: "relay_backend",
		ID:          "cost_matrix.duration",
		Unit:        "milliseconds",
		Description: "How long it takes to generate a cost matrix from the stats database.",
	})
	if err != nil {
		return nil, err
	}

	costMatrixInvocationsCounter, err := metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Total StatsDB -> CostMatrix invocations",
		ServiceName: "relay_backend",
		ID:          "cost_matrix.count",
		Unit:        "invocations",
		Description: "The total number of StatsDB -> CostMatrix invocations",
	})
	if err != nil {
		return nil, err
	}

	costMatrixLongUpdateCounter, err := metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Cost Matrix Long Updates",
		ServiceName: "relay_backend",
		ID:          "cost_matrix.long.updates",
		Unit:        "updates",
		Description: "The number of cost matrix gen calls that took longer than 1 second",
	})
	if err != nil {
		return nil, err
	}

	costMatrixBytes, err := metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Cost Matrix Size",
		ServiceName: "relay_backend",
		ID:          "cost_matrix.bytes",
		Unit:        "bytes",
		Description: "How large the cost matrix is in bytes",
	})
	if err != nil {
		return nil, err
	}

	costMatrixGenFailure, err := metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Cost Matrix Gen Failure",
		ServiceName: "relay_backend",
		ID:          "cost_matrix.failure",
		Unit:        "errors",
		Description: "How many times the relay backend failed to generate the cost matrix",
	})
	if err != nil {
		return nil, err
	}

	costMatrixMetrics := CostMatrixMetrics{
		Invocations:     costMatrixInvocationsCounter,
		DurationGauge:   costMatrixDurationGauge,
		LongUpdateCount: costMatrixLongUpdateCounter,
		Bytes:           costMatrixBytes,
		ErrorMetrics: CostMatrixErrorMetrics{
			GenFailure: costMatrixGenFailure,
		},
	}

	return &costMatrixMetrics, nil
}

func NewOptimizeMetrics(ctx context.Context, metricsHandler Handler) (*OptimizeMetrics, error) {
	optimizeDurationGauge, err := metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Optimize duration",
		ServiceName: "relay_backend",
		ID:          "optimize.duration",
		Unit:        "milliseconds",
		Description: "How long it takes to optimize a cost matrix.",
	})
	if err != nil {
		return nil, err
	}

	optimizeInvocationsCounter, err := metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Total cost matrix optimize invocations",
		ServiceName: "relay_backend",
		ID:          "optimize.count",
		Unit:        "invocations",
		Description: "The total number of cost matrix optimizers",
	})
	if err != nil {
		return nil, err
	}

	optimizeMetrics := OptimizeMetrics{
		Invocations:   optimizeInvocationsCounter,
		DurationGauge: optimizeDurationGauge,
	}

	optimizeMetrics.LongUpdateCount, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Optimize Long Updates",
		ServiceName: "relay_backend",
		ID:          "optimize.long.updates",
		Unit:        "updates",
		Description: "The number of optimize calls that took longer than 1 second",
	})
	if err != nil {
		return nil, err
	}

	return &optimizeMetrics, nil
}

func NewMaxmindSyncMetrics(ctx context.Context, metricsHandler Handler) (*MaxmindSyncMetrics, error) {
	duration, err := metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Maxmind Sync Duration",
		ServiceName: "relay_backend",
		ID:          "maxmind.duration",
		Unit:        "milliseconds",
		Description: "How long it takes to sync the maxmind database from Maxmind.com",
	})
	if err != nil {
		return nil, err
	}

	invocations, err := metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Total Maxmind Sync Invocations",
		ServiceName: "relay_backend",
		ID:          "maxmind.count",
		Unit:        "invocations",
		Description: "The total number of Maxmind sync invocations",
	})
	if err != nil {
		return nil, err
	}

	maxmindSyncMetrics := MaxmindSyncMetrics{
		Invocations:   invocations,
		DurationGauge: duration,
	}

	maxmindSyncMetrics.ErrorMetrics.FailedToSync, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Failed To Sync MaxmindDB",
		ServiceName: "relay_backend",
		ID:          "maxmind.error.failed_to_sync",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	maxmindSyncMetrics.ErrorMetrics.FailedToSyncISP, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Failed To Sync MaxmindDB ISP",
		ServiceName: "relay_backend",
		ID:          "maxmind.error.failed_to_sync_isp",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	return &maxmindSyncMetrics, nil
}

func NewBillingMetrics(ctx context.Context, metricsHandler Handler) (*BillingMetrics, error) {
	billingMetrics := BillingMetrics{}
	var err error

	billingMetrics.Goroutines, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Billing Goroutine Count",
		ServiceName: "billing",
		ID:          "billing.goroutines",
		Unit:        "goroutines",
		Description: "The number of goroutines the billing service is using",
	})
	if err != nil {
		return nil, err
	}

	billingMetrics.MemoryAllocated, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Billing Memory Allocated",
		ServiceName: "billing",
		ID:          "billing.memory",
		Unit:        "MB",
		Description: "The amount of memory the billing service has allocated in MB",
	})
	if err != nil {
		return nil, err
	}

	billingMetrics.EntriesReceived, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Billing Entries Received",
		ServiceName: "billing",
		ID:          "billing.entries",
		Unit:        "entries",
		Description: "The total number of billing entries received through pubsub",
	})
	if err != nil {
		return nil, err
	}

	billingMetrics.EntriesSubmitted, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Billing Entries Submitted",
		ServiceName: "billing",
		ID:          "billing.entries.submitted",
		Unit:        "entries",
		Description: "The total number of billing entries submitted to BigQuery",
	})
	if err != nil {
		return nil, err
	}

	billingMetrics.EntriesQueued, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Billing Entries Queued",
		ServiceName: "billing",
		ID:          "billing.entries.queued",
		Unit:        "entries",
		Description: "The total number of billing entries waiting to be sent to BigQuery",
	})
	if err != nil {
		return nil, err
	}

	billingMetrics.EntriesFlushed, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Billing Entries Written",
		ServiceName: "billing",
		ID:          "billing.entries.written",
		Unit:        "entries",
		Description: "The total number of billing entries written to bigquery",
	})
	if err != nil {
		return nil, err
	}

	billingMetrics.ErrorMetrics.BillingPublishFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Billing Publish Failure",
		ServiceName: "billing",
		ID:          "billing.error.publish_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	billingMetrics.ErrorMetrics.BillingReadFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Billing Read Failure",
		ServiceName: "billing",
		ID:          "billing.error.read_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	billingMetrics.ErrorMetrics.BillingWriteFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Billing Write Failure",
		ServiceName: "billing",
		ID:          "billing.error.write_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	return &billingMetrics, nil
}

func NewRelayBackendMetrics(ctx context.Context, metricsHandler Handler) (*RelayBackendMetrics, error) {
	relayBackendMetrics := RelayBackendMetrics{}
	var err error

	relayBackendMetrics.Goroutines, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Relay Backend Goroutine Count",
		ServiceName: "relay_backend",
		ID:          "relay_backend.goroutines",
		Unit:        "goroutines",
		Description: "The number of goroutines the relay backend service is using",
	})
	if err != nil {
		return nil, err
	}

	relayBackendMetrics.MemoryAllocated, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Relay Backend Memory Allocated",
		ServiceName: "relay_backend",
		ID:          "relay_backend.memory",
		Unit:        "MB",
		Description: "The amount of memory the relay backend service has allocated in MB",
	})
	if err != nil {
		return nil, err
	}

	return &relayBackendMetrics, nil
}

func NewRouteMatrixMetrics(ctx context.Context, metricsHandler Handler) (*RouteMatrixMetrics, error) {
	routeMatrixMetrics := RouteMatrixMetrics{}
	var err error

	routeMatrixMetrics.DatacenterCount, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Route Matrix Datacenter Count",
		ServiceName: "relay_backend",
		ID:          "route_matrix.datacenter.count",
		Unit:        "datacenters",
		Description: "The number of datacenters the route matrix contains",
	})
	if err != nil {
		return nil, err
	}

	routeMatrixMetrics.RelayCount, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Route Matrix Relay Count",
		ServiceName: "relay_backend",
		ID:          "route_matrix.relay.count",
		Unit:        "relays",
		Description: "The number of relays the route matrix contains",
	})
	if err != nil {
		return nil, err
	}

	routeMatrixMetrics.RouteCount, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Route Matrix Route Count",
		ServiceName: "relay_backend",
		ID:          "route_matrix.route.count",
		Unit:        "routes",
		Description: "The number of routes the route matrix contains",
	})
	if err != nil {
		return nil, err
	}

	routeMatrixMetrics.Bytes, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Route Matrix Size",
		ServiceName: "relay_backend",
		ID:          "route_matrix.bytes",
		Unit:        "bytes",
		Description: "How large the route matrix is in bytes",
	})
	if err != nil {
		return nil, err
	}

	return &routeMatrixMetrics, nil
}
