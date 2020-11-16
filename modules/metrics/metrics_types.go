package metrics

import (
	"context"
)

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
	UnmarshalFailure    Counter
	ExceedMaxRelays     Counter
	RelayNotFound       Counter
	RelayQuarantined    Counter
	NoAuthHeader        Counter
	BadAuthHeaderLength Counter
	BadAuthHeaderToken  Counter
	BadNonce            Counter
	BadEncryptedAddress Counter
	DecryptFailure      Counter
}

var EmptyRelayHandlerErrorMetrics RelayHandlerErrorMetrics = RelayHandlerErrorMetrics{
	UnmarshalFailure:    &EmptyCounter{},
	ExceedMaxRelays:     &EmptyCounter{},
	RelayNotFound:       &EmptyCounter{},
	RelayQuarantined:    &EmptyCounter{},
	NoAuthHeader:        &EmptyCounter{},
	BadAuthHeaderLength: &EmptyCounter{},
	BadAuthHeaderToken:  &EmptyCounter{},
	BadNonce:            &EmptyCounter{},
	BadEncryptedAddress: &EmptyCounter{},
	DecryptFailure:      &EmptyCounter{},
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
}

var EmptyCostMatrixMetrics CostMatrixMetrics = CostMatrixMetrics{
	Invocations:     &EmptyCounter{},
	DurationGauge:   &EmptyGauge{},
	LongUpdateCount: &EmptyCounter{},
	Bytes:           &EmptyGauge{},
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

type BillingServiceMetrics struct {
	Goroutines      Gauge
	MemoryAllocated Gauge
	BillingMetrics  BillingMetrics
}

var EmptyBillingServiceMetrics BillingServiceMetrics = BillingServiceMetrics{
	Goroutines:      &EmptyGauge{},
	MemoryAllocated: &EmptyGauge{},
	BillingMetrics:  EmptyBillingMetrics,
}

type BillingMetrics struct {
	EntriesReceived  Counter
	EntriesSubmitted Counter
	EntriesQueued    Gauge
	EntriesFlushed   Counter
	ErrorMetrics     BillingErrorMetrics
}

var EmptyBillingMetrics BillingMetrics = BillingMetrics{
	EntriesReceived:  &EmptyCounter{},
	EntriesSubmitted: &EmptyCounter{},
	EntriesQueued:    &EmptyGauge{},
	EntriesFlushed:   &EmptyCounter{},
	ErrorMetrics:     EmptyBillingErrorMetrics,
}

type BillingErrorMetrics struct {
	BillingPublishFailure     Counter
	BillingReadFailure        Counter
	BillingBatchedReadFailure Counter
	BillingWriteFailure       Counter
}

var EmptyBillingErrorMetrics BillingErrorMetrics = BillingErrorMetrics{
	BillingPublishFailure:     &EmptyCounter{},
	BillingReadFailure:        &EmptyCounter{},
	BillingBatchedReadFailure: &EmptyCounter{},
	BillingWriteFailure:       &EmptyCounter{},
}

type AnalyticsMetrics struct {
	EntriesReceived  Counter
	EntriesSubmitted Counter
	EntriesQueued    Counter
	EntriesFlushed   Counter
	ErrorMetrics     AnalyticsErrorMetrics
}

type AnalyticsErrorMetrics struct {
	PublishFailure Counter
	ReadFailure    Counter
	WriteFailure   Counter
}

var EmptyAnalyticsErrorMetrics AnalyticsErrorMetrics = AnalyticsErrorMetrics{
	PublishFailure: &EmptyCounter{},
	ReadFailure:    &EmptyCounter{},
	WriteFailure:   &EmptyCounter{},
}

var EmptyAnalyticsMetrics AnalyticsMetrics = AnalyticsMetrics{
	EntriesReceived:  &EmptyCounter{},
	EntriesSubmitted: &EmptyCounter{},
	EntriesQueued:    &EmptyCounter{},
	EntriesFlushed:   &EmptyCounter{},
	ErrorMetrics:     EmptyAnalyticsErrorMetrics,
}

type RelayBackendMetrics struct {
	Goroutines        Gauge
	MemoryAllocated   Gauge
	RouteMatrix       RouteMatrixMetrics
	PingStatsMetrics  AnalyticsMetrics
	RelayStatsMetrics AnalyticsMetrics
}

var EmptyRelayBackendMetrics RelayBackendMetrics = RelayBackendMetrics{
	Goroutines:        &EmptyGauge{},
	MemoryAllocated:   &EmptyGauge{},
	RouteMatrix:       EmptyRouteMatrixMetrics,
	PingStatsMetrics:  EmptyAnalyticsMetrics,
	RelayStatsMetrics: EmptyAnalyticsMetrics,
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

type AnalyticsServiceMetrics struct {
	Goroutines        Gauge
	MemoryAllocated   Gauge
	PingStatsMetrics  AnalyticsMetrics
	RelayStatsMetrics AnalyticsMetrics
}

var EmptyAnalyticsServiceMetrics = AnalyticsServiceMetrics{
	Goroutines:        &EmptyGauge{},
	MemoryAllocated:   &EmptyGauge{},
	PingStatsMetrics:  EmptyAnalyticsMetrics,
	RelayStatsMetrics: EmptyAnalyticsMetrics,
}

type PortalCruncherMetrics struct {
	Goroutines           Gauge
	MemoryAllocated      Gauge
	ReceivedMessageCount Counter
}

var EmptyPortalCruncherMetrics = PortalCruncherMetrics{
	Goroutines:           &EmptyGauge{},
	MemoryAllocated:      &EmptyGauge{},
	ReceivedMessageCount: &EmptyCounter{},
}

type BigTableMetrics struct {
	WriteMetaSuccessCount  Counter
	WriteSliceSuccessCount Counter
	WriteMetaFailureCount  Counter
	WriteSliceFailureCount Counter
	ReadMetaSuccessCount   Counter
	ReadSliceSuccessCount  Counter
	ReadMetaFailureCount   Counter
	ReadSliceFailureCount  Counter
}

var EmptyBigTableMetrics = BigTableMetrics{
	WriteMetaSuccessCount:  &EmptyCounter{},
	WriteSliceSuccessCount: &EmptyCounter{},
	WriteMetaFailureCount:  &EmptyCounter{},
	WriteSliceFailureCount: &EmptyCounter{},
	ReadMetaSuccessCount:   &EmptyCounter{},
	ReadSliceSuccessCount:  &EmptyCounter{},
	ReadMetaFailureCount:   &EmptyCounter{},
	ReadSliceFailureCount:  &EmptyCounter{},
}

func NewBigTableMetrics(ctx context.Context, metricsHandler Handler) (*BigTableMetrics, error) {
	var err error

	bigtableMetrics := BigTableMetrics{}

	bigtableMetrics.WriteMetaSuccessCount, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Bigtable Write Meta Success Count",
		ServiceName: "bigtable",
		ID:          "bigtable.write.meta.success.count",
		Unit:        "writes",
		Description: "The number of successful meta writes to bigtable",
	})
	if err != nil {
		return nil, err
	}

	bigtableMetrics.WriteSliceSuccessCount, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Bigtable Write Slice Success Count",
		ServiceName: "bigtable",
		ID:          "bigtable.write.slice.success.count",
		Unit:        "writes",
		Description: "The number of successful slice writes to bigtable",
	})
	if err != nil {
		return nil, err
	}

	bigtableMetrics.WriteMetaFailureCount, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Bigtable Write Meta Failure Count",
		ServiceName: "bigtable",
		ID:          "bigtable.write.meta.failure.count",
		Unit:        "writes",
		Description: "The number of failed meta writes to bigtable",
	})
	if err != nil {
		return nil, err
	}

	bigtableMetrics.WriteSliceFailureCount, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Bigtable Write Slice Failure Count",
		ServiceName: "bigtable",
		ID:          "bigtable.write.slice.failure.count",
		Unit:        "writes",
		Description: "The number of failed slice writes to bigtable",
	})
	if err != nil {
		return nil, err
	}

	bigtableMetrics.ReadMetaSuccessCount, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Bigtable Read Meta Success Count",
		ServiceName: "bigtable",
		ID:          "bigtable.read.meta.success.count",
		Unit:        "writes",
		Description: "The number of successful meta reads from bigtable",
	})
	if err != nil {
		return nil, err
	}

	bigtableMetrics.ReadSliceSuccessCount, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Bigtable Read Slice Success Count",
		ServiceName: "bigtable",
		ID:          "bigtable.read.slice.success.count",
		Unit:        "writes",
		Description: "The number of successful slice reads from bigtable",
	})
	if err != nil {
		return nil, err
	}

	bigtableMetrics.ReadMetaFailureCount, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Bigtable Read Meta Failure Count",
		ServiceName: "bigtable",
		ID:          "bigtable.read.meta.failure.count",
		Unit:        "writes",
		Description: "The number of failed meta reads from bigtable",
	})
	if err != nil {
		return nil, err
	}

	bigtableMetrics.ReadSliceFailureCount, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Bigtable Read Slice Failure Count",
		ServiceName: "bigtable",
		ID:          "bigtable.read.slice.failure.count",
		Unit:        "writes",
		Description: "The number of failed slice reads from bigtable",
	})
	if err != nil {
		return nil, err
	}

	return &bigtableMetrics, nil
}

func NewPortalCruncherMetrics(ctx context.Context, metricsHandler Handler) (*PortalCruncherMetrics, error) {
	var err error

	portalCruncherMetrics := PortalCruncherMetrics{}

	portalCruncherMetrics.Goroutines, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Portal Cruncher Goroutine Count",
		ServiceName: "portal_cruncher",
		ID:          "portal_cruncher.goroutines",
		Unit:        "goroutines",
		Description: "The number of goroutines the portal_cruncher is using",
	})
	if err != nil {
		return nil, err
	}

	portalCruncherMetrics.MemoryAllocated, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Portal Cruncher Memory Allocated",
		ServiceName: "portal_cruncher",
		ID:          "portal_cruncher.memory",
		Unit:        "MB",
		Description: "The amount of memory the portal_cruncher has allocated in MB",
	})
	if err != nil {
		return nil, err
	}

	portalCruncherMetrics.ReceivedMessageCount, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Portal Cruncher Received Message Count",
		ServiceName: "portal_cruncher",
		ID:          "portal_cruncher.received.message.count",
		Unit:        "messages",
		Description: "The amount of messages the portal_cruncher has received",
	})
	if err != nil {
		return nil, err
	}

	return &portalCruncherMetrics, nil
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

	costMatrixMetrics := CostMatrixMetrics{
		Invocations:     costMatrixInvocationsCounter,
		DurationGauge:   costMatrixDurationGauge,
		LongUpdateCount: costMatrixLongUpdateCounter,
		Bytes:           costMatrixBytes,
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
		Description: "The total number of cost matrix optimize calls",
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

func NewValveCostMatrixMetrics(ctx context.Context, metricsHandler Handler) (*CostMatrixMetrics, error) {
	costMatrixDurationGauge, err := metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Valve StatsDB -> GetCostMatrix duration",
		ServiceName: "relay_backend",
		ID:          "cost_matrix.valve.duration",
		Unit:        "milliseconds",
		Description: "How long it takes to generate a valve cost matrix from the stats database.",
	})
	if err != nil {
		return nil, err
	}

	costMatrixInvocationsCounter, err := metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Valve Total StatsDB -> CostMatrix invocations",
		ServiceName: "relay_backend",
		ID:          "cost_matrix.valve.count",
		Unit:        "invocations",
		Description: "The total number of valve StatsDB -> CostMatrix invocations",
	})
	if err != nil {
		return nil, err
	}

	costMatrixLongUpdateCounter, err := metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Valve Cost Matrix Long Updates",
		ServiceName: "relay_backend",
		ID:          "cost_matrix.valve.long.updates",
		Unit:        "updates",
		Description: "The number of valve cost matrix gen calls that took longer than 1 second",
	})
	if err != nil {
		return nil, err
	}

	costMatrixBytes, err := metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Valve Cost Matrix Size",
		ServiceName: "relay_backend",
		ID:          "cost_matrix.valve.bytes",
		Unit:        "bytes",
		Description: "How large the valve cost matrix is in bytes",
	})
	if err != nil {
		return nil, err
	}

	costMatrixMetrics := CostMatrixMetrics{
		Invocations:     costMatrixInvocationsCounter,
		DurationGauge:   costMatrixDurationGauge,
		LongUpdateCount: costMatrixLongUpdateCounter,
		Bytes:           costMatrixBytes,
	}

	return &costMatrixMetrics, nil
}

func NewValveOptimizeMetrics(ctx context.Context, metricsHandler Handler) (*OptimizeMetrics, error) {
	optimizeDurationGauge, err := metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Valve optimize duration",
		ServiceName: "relay_backend",
		ID:          "optimize.valve.duration",
		Unit:        "milliseconds",
		Description: "How long it takes to optimize a valve cost matrix.",
	})
	if err != nil {
		return nil, err
	}

	optimizeInvocationsCounter, err := metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Valve total cost matrix optimize invocations",
		ServiceName: "relay_backend",
		ID:          "optimize.valve.count",
		Unit:        "invocations",
		Description: "The total number of valve cost matrix optimize calls",
	})
	if err != nil {
		return nil, err
	}

	optimizeMetrics := OptimizeMetrics{
		Invocations:   optimizeInvocationsCounter,
		DurationGauge: optimizeDurationGauge,
	}

	optimizeMetrics.LongUpdateCount, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Valve Optimize Long Updates",
		ServiceName: "relay_backend",
		ID:          "optimize.valve.long.updates",
		Unit:        "updates",
		Description: "The number of valve optimize calls that took longer than 1 second",
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

func NewBillingServiceMetrics(ctx context.Context, metricsHandler Handler) (*BillingServiceMetrics, error) {
	billingServiceMetrics := BillingServiceMetrics{}
	var err error

	billingServiceMetrics.Goroutines, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Billing Goroutine Count",
		ServiceName: "billing",
		ID:          "billing.goroutines",
		Unit:        "goroutines",
		Description: "The number of goroutines the billing service is using",
	})
	if err != nil {
		return nil, err
	}

	billingServiceMetrics.MemoryAllocated, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Billing Memory Allocated",
		ServiceName: "billing",
		ID:          "billing.memory",
		Unit:        "MB",
		Description: "The amount of memory the billing service has allocated in MB",
	})
	if err != nil {
		return nil, err
	}

	billingServiceMetrics.BillingMetrics.EntriesReceived, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Billing Entries Received",
		ServiceName: "billing",
		ID:          "billing.entries",
		Unit:        "entries",
		Description: "The total number of billing entries received through Google Pub/Sub",
	})
	if err != nil {
		return nil, err
	}

	billingServiceMetrics.BillingMetrics.EntriesSubmitted, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Billing Entries Submitted",
		ServiceName: "billing",
		ID:          "billing.entries.submitted",
		Unit:        "entries",
		Description: "The total number of billing entries submitted to BigQuery",
	})
	if err != nil {
		return nil, err
	}

	billingServiceMetrics.BillingMetrics.EntriesQueued, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Billing Entries Queued",
		ServiceName: "billing",
		ID:          "billing.entries.queued",
		Unit:        "entries",
		Description: "The total number of billing entries waiting to be sent to BigQuery",
	})
	if err != nil {
		return nil, err
	}

	billingServiceMetrics.BillingMetrics.EntriesFlushed, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Billing Entries Written",
		ServiceName: "billing",
		ID:          "billing.entries.written",
		Unit:        "entries",
		Description: "The total number of billing entries written to BigQuery",
	})
	if err != nil {
		return nil, err
	}

	billingServiceMetrics.BillingMetrics.ErrorMetrics.BillingPublishFailure = &EmptyCounter{}

	billingServiceMetrics.BillingMetrics.ErrorMetrics.BillingReadFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Billing Read Failure",
		ServiceName: "billing",
		ID:          "billing.error.read_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	billingServiceMetrics.BillingMetrics.ErrorMetrics.BillingBatchedReadFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Billing Batched Read Failure",
		ServiceName: "billing",
		ID:          "billing.error.batched_read_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	billingServiceMetrics.BillingMetrics.ErrorMetrics.BillingWriteFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Billing Write Failure",
		ServiceName: "billing",
		ID:          "billing.error.write_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	return &billingServiceMetrics, nil
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

	relayBackendMetrics.RouteMatrix.DatacenterCount, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Route Matrix Datacenter Count",
		ServiceName: "relay_backend",
		ID:          "route_matrix.datacenter.count",
		Unit:        "datacenters",
		Description: "The number of datacenters the route matrix contains",
	})
	if err != nil {
		return nil, err
	}

	relayBackendMetrics.RouteMatrix.RelayCount, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Route Matrix Relay Count",
		ServiceName: "relay_backend",
		ID:          "route_matrix.relay.count",
		Unit:        "relays",
		Description: "The number of relays the route matrix contains",
	})
	if err != nil {
		return nil, err
	}

	relayBackendMetrics.RouteMatrix.RouteCount, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Route Matrix Route Count",
		ServiceName: "relay_backend",
		ID:          "route_matrix.route.count",
		Unit:        "routes",
		Description: "The number of routes the route matrix contains",
	})
	if err != nil {
		return nil, err
	}

	relayBackendMetrics.RouteMatrix.Bytes, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Route Matrix Size",
		ServiceName: "relay_backend",
		ID:          "route_matrix.bytes",
		Unit:        "bytes",
		Description: "How large the route matrix is in bytes",
	})
	if err != nil {
		return nil, err
	}

	relayBackendMetrics.PingStatsMetrics.EntriesReceived = &EmptyCounter{}

	relayBackendMetrics.PingStatsMetrics.EntriesSubmitted, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Relay Backend Ping Stats Entries Written",
		ServiceName: "relay_backend",
		ID:          "relay_backend.ping_stats.entries.submitted",
		Unit:        "entries",
		Description: "The number of ping stats entries the relay backend submitted to be published",
	})
	if err != nil {
		return nil, err
	}

	relayBackendMetrics.PingStatsMetrics.EntriesQueued, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Relay Backend Ping Stats Entries Queued",
		ServiceName: "relay_backend",
		ID:          "relay_backend.ping_stats.entries.queued",
		Unit:        "entries",
		Description: "The number of ping stats entries the relay backend has queued. This should always be 0",
	})
	if err != nil {
		return nil, err
	}

	relayBackendMetrics.PingStatsMetrics.EntriesFlushed, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Relay Backend Ping Stats Entries Flushed",
		ServiceName: "relay_backend",
		ID:          "relay_backend.ping_stats.entries.flushed",
		Unit:        "entries",
		Description: "The number of ping stats entries the relay backend has flushed",
	})
	if err != nil {
		return nil, err
	}

	relayBackendMetrics.PingStatsMetrics.ErrorMetrics.PublishFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Relay Backend Ping Stats Publish Failure",
		ServiceName: "relay_backend",
		ID:          "relay_backend.ping_stats.error.publish_failure",
		Unit:        "entries",
		Description: "The number of ping stats entries the relay backend has failed to publish",
	})
	if err != nil {
		return nil, err
	}

	relayBackendMetrics.PingStatsMetrics.ErrorMetrics.ReadFailure = &EmptyCounter{}

	relayBackendMetrics.PingStatsMetrics.ErrorMetrics.WriteFailure = &EmptyCounter{}

	relayBackendMetrics.RelayStatsMetrics.EntriesReceived = &EmptyCounter{}

	relayBackendMetrics.RelayStatsMetrics.EntriesSubmitted, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Relay Backend Relay Stats Entries Written",
		ServiceName: "relay_backend",
		ID:          "relay_backend.relay_stats.entries.submitted",
		Unit:        "entries",
		Description: "The number of relay stats entries the relay backend submitted to be published",
	})
	if err != nil {
		return nil, err
	}

	relayBackendMetrics.RelayStatsMetrics.EntriesQueued, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Relay Backend Relay Stats Entries Queued",
		ServiceName: "relay_backend",
		ID:          "relay_backend.relay_stats.entries.queued",
		Unit:        "entries",
		Description: "The number of relay stats entries the relay backend has queued. This should always be 0",
	})
	if err != nil {
		return nil, err
	}

	relayBackendMetrics.RelayStatsMetrics.EntriesFlushed, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Relay Backend Relay Stats Entries Flushed",
		ServiceName: "relay_backend",
		ID:          "relay_backend.relay_stats.entries.flushed",
		Unit:        "entries",
		Description: "The number of relay stats entries the relay backend has flushed",
	})
	if err != nil {
		return nil, err
	}

	relayBackendMetrics.RelayStatsMetrics.ErrorMetrics.PublishFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Relay Backend Relay Stats Publish Failure",
		ServiceName: "relay_backend",
		ID:          "relay_backend.relay_stats.error.publish_failure",
		Unit:        "entries",
		Description: "The number of relay stats entries the relay backend has failed to publish",
	})
	if err != nil {
		return nil, err
	}

	relayBackendMetrics.RelayStatsMetrics.ErrorMetrics.ReadFailure = &EmptyCounter{}

	relayBackendMetrics.RelayStatsMetrics.ErrorMetrics.WriteFailure = &EmptyCounter{}

	return &relayBackendMetrics, nil
}

func NewValveRouteMatrixMetrics(ctx context.Context, metricsHandler Handler) (*RouteMatrixMetrics, error) {
	routeMatrixMetrics := RouteMatrixMetrics{}
	var err error

	routeMatrixMetrics.DatacenterCount, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Valve Route Matrix Datacenter Count",
		ServiceName: "relay_backend",
		ID:          "route_matrix.valve.datacenter.count",
		Unit:        "datacenters",
		Description: "The number of datacenters the valve route matrix contains",
	})
	if err != nil {
		return nil, err
	}

	routeMatrixMetrics.RelayCount, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Valve Route Matrix Relay Count",
		ServiceName: "relay_backend",
		ID:          "route_matrix.valve.relay.count",
		Unit:        "relays",
		Description: "The number of relays the valve route matrix contains",
	})
	if err != nil {
		return nil, err
	}

	routeMatrixMetrics.RouteCount, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Valve Route Matrix Route Count",
		ServiceName: "relay_backend",
		ID:          "route_matrix.valve.route.count",
		Unit:        "routes",
		Description: "The number of routes the valve route matrix contains",
	})
	if err != nil {
		return nil, err
	}

	routeMatrixMetrics.Bytes, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Valve Route Matrix Size",
		ServiceName: "relay_backend",
		ID:          "route_matrix.valve.bytes",
		Unit:        "bytes",
		Description: "How large the valve route matrix is in bytes",
	})
	if err != nil {
		return nil, err
	}

	return &routeMatrixMetrics, nil
}

func NewAnalyticsServiceMetrics(ctx context.Context, metricsHandler Handler) (*AnalyticsServiceMetrics, error) {
	analyticsMetrics := AnalyticsServiceMetrics{}
	var err error

	analyticsMetrics.Goroutines, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Analytics Goroutine Count",
		ServiceName: "analytics",
		ID:          "analytics.goroutines",
		Unit:        "goroutines",
		Description: "The number of goroutines the analytics service is using",
	})
	if err != nil {
		return nil, err
	}

	analyticsMetrics.MemoryAllocated, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Analytics Memory Allocated",
		ServiceName: "analytics",
		ID:          "analytics.memory",
		Unit:        "MB",
		Description: "The amount of memory the analytics service has allocated in MB",
	})
	if err != nil {
		return nil, err
	}

	analyticsMetrics.PingStatsMetrics.EntriesReceived, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Ping Stats Entries Received",
		ServiceName: "analytics",
		ID:          "analytics.ping_stats.entries",
		Unit:        "entries",
		Description: "The total number of ping stats entries received through Google Pub/Sub",
	})
	if err != nil {
		return nil, err
	}

	analyticsMetrics.PingStatsMetrics.EntriesSubmitted, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Ping Stats Entries Submitted",
		ServiceName: "analytics",
		ID:          "analytics.ping_stats.entries.submitted",
		Unit:        "entries",
		Description: "The total number of ping stats entries submitted to BigQuery",
	})
	if err != nil {
		return nil, err
	}

	analyticsMetrics.PingStatsMetrics.EntriesQueued, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Ping Stats Entries Queued",
		ServiceName: "analytics",
		ID:          "analytics.ping_stats.entries.queued",
		Unit:        "entries",
		Description: "The total number of ping stats entries waiting to be sent to BigQuery",
	})
	if err != nil {
		return nil, err
	}

	analyticsMetrics.PingStatsMetrics.EntriesFlushed, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Ping Stats Entries Flushed",
		ServiceName: "analytics",
		ID:          "analytics.ping_stats.entries.flushed",
		Unit:        "entries",
		Description: "The total number of ping stats entries written to BigQuery",
	})
	if err != nil {
		return nil, err
	}

	analyticsMetrics.PingStatsMetrics.ErrorMetrics.PublishFailure = &EmptyCounter{}

	analyticsMetrics.PingStatsMetrics.ErrorMetrics.ReadFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Ping Stats Read Failure",
		ServiceName: "analytics",
		ID:          "analytics.ping_stats.error.read_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	analyticsMetrics.PingStatsMetrics.ErrorMetrics.WriteFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Ping Stats Write Failure",
		ServiceName: "analytics",
		ID:          "analytics.ping_stats.error.write_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	analyticsMetrics.RelayStatsMetrics.EntriesReceived, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Relay Stats Entries Received",
		ServiceName: "analytics",
		ID:          "analytics.relay_stats.entries",
		Unit:        "entries",
		Description: "The total number of relay stats entries received through Google Pub/Sub",
	})
	if err != nil {
		return nil, err
	}

	analyticsMetrics.RelayStatsMetrics.EntriesSubmitted, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Relay Stats Entries Submitted",
		ServiceName: "analytics",
		ID:          "analytics.relay_stats.entries.submitted",
		Unit:        "entries",
		Description: "The total number of relay stats entries submitted to BigQuery",
	})
	if err != nil {
		return nil, err
	}

	analyticsMetrics.RelayStatsMetrics.EntriesQueued, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Relay Stats Entries Queued",
		ServiceName: "analytics",
		ID:          "analytics.relay_stats.entries.queued",
		Unit:        "entries",
		Description: "The total number of relay stats entries waiting to be sent to BigQuery",
	})
	if err != nil {
		return nil, err
	}

	analyticsMetrics.RelayStatsMetrics.EntriesFlushed, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Relay Stats Entries Flushed",
		ServiceName: "analytics",
		ID:          "analytics.relay_stats.entries.flushed",
		Unit:        "entries",
		Description: "The total number of relay stats entries written to BigQuery",
	})
	if err != nil {
		return nil, err
	}

	analyticsMetrics.RelayStatsMetrics.ErrorMetrics.PublishFailure = &EmptyCounter{}

	analyticsMetrics.RelayStatsMetrics.ErrorMetrics.ReadFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Relay Stats Read Failure",
		ServiceName: "analytics",
		ID:          "analytics.relay_stats.error.read_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	analyticsMetrics.RelayStatsMetrics.ErrorMetrics.WriteFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Relay Stats Write Failure",
		ServiceName: "analytics",
		ID:          "analytics.relay_stats.error.write_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	return &analyticsMetrics, nil
}
