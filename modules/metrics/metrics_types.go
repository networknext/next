package metrics

import (
	"context"
)

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
	EntriesReceived        Counter
	EntriesSubmitted       Counter
	EntriesQueued          Gauge
	EntriesFlushed         Counter
	PubsubBillingEntrySize Gauge
	BillingEntrySize       Gauge

	Entries2Received         Counter
	Entries2Submitted        Counter
	Entries2Queued           Gauge
	Entries2Flushed          Counter
	SummaryEntries2Submitted Counter
	SummaryEntries2Queued    Gauge
	SummaryEntries2Flushed   Counter
	PubsubBillingEntry2Size  Gauge
	BillingEntry2Size        Gauge

	ErrorMetrics BillingErrorMetrics
}

var EmptyBillingMetrics BillingMetrics = BillingMetrics{
	EntriesReceived:        &EmptyCounter{},
	EntriesSubmitted:       &EmptyCounter{},
	EntriesQueued:          &EmptyGauge{},
	EntriesFlushed:         &EmptyCounter{},
	PubsubBillingEntrySize: &EmptyGauge{},
	BillingEntrySize:       &EmptyGauge{},

	Entries2Received:         &EmptyCounter{},
	Entries2Submitted:        &EmptyCounter{},
	Entries2Queued:           &EmptyGauge{},
	Entries2Flushed:          &EmptyCounter{},
	SummaryEntries2Submitted: &EmptyCounter{},
	SummaryEntries2Queued:    &EmptyGauge{},
	SummaryEntries2Flushed:   &EmptyCounter{},
	PubsubBillingEntry2Size:  &EmptyGauge{},
	BillingEntry2Size:        &EmptyGauge{},

	ErrorMetrics: EmptyBillingErrorMetrics,
}

type BillingErrorMetrics struct {
	BillingPublishFailure     Counter
	BillingReadFailure        Counter
	BillingBatchedReadFailure Counter
	BillingWriteFailure       Counter
	BillingInvalidEntries     Counter
	BillingEntriesWithNaN     Counter

	Billing2PublishFailure     Counter
	Billing2ReadFailure        Counter
	Billing2BatchedReadFailure Counter
	Billing2WriteFailure       Counter
	Billing2InvalidEntries     Counter
	Billing2EntriesWithNaN     Counter
}

var EmptyBillingErrorMetrics BillingErrorMetrics = BillingErrorMetrics{
	BillingPublishFailure:     &EmptyCounter{},
	BillingReadFailure:        &EmptyCounter{},
	BillingBatchedReadFailure: &EmptyCounter{},
	BillingWriteFailure:       &EmptyCounter{},
	BillingInvalidEntries:     &EmptyCounter{},
	BillingEntriesWithNaN:     &EmptyCounter{},

	Billing2PublishFailure:     &EmptyCounter{},
	Billing2ReadFailure:        &EmptyCounter{},
	Billing2BatchedReadFailure: &EmptyCounter{},
	Billing2WriteFailure:       &EmptyCounter{},
	Billing2InvalidEntries:     &EmptyCounter{},
	Billing2EntriesWithNaN:     &EmptyCounter{},
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

type AnalyticsServiceMetrics struct {
	Goroutines              Gauge
	MemoryAllocated         Gauge
	PingStatsMetrics        AnalyticsMetrics
	RelayStatsMetrics       AnalyticsMetrics
	RouteMatrixStatsMetrics AnalyticsMetrics
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

type BuyerEndpointMetrics struct {
	NoSlicesFailure Counter
}

var EmptyBuyerEndpointMetrics = BuyerEndpointMetrics{
	NoSlicesFailure: &EmptyCounter{},
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

type VanityServiceMetrics struct {
	Goroutines               Gauge
	MemoryAllocated          Gauge
	ReceivedVanityCount      Counter
	UpdateVanitySuccessCount Counter
	UpdateVanityFailureCount Counter
	ReadVanitySuccessCount   Counter
	ReadVanityFailureCount   Counter
}

var EmptyVanityServiceMetrics = VanityServiceMetrics{
	Goroutines:               &EmptyGauge{},
	MemoryAllocated:          &EmptyGauge{},
	ReceivedVanityCount:      &EmptyCounter{},
	UpdateVanitySuccessCount: &EmptyCounter{},
	UpdateVanityFailureCount: &EmptyCounter{},
	ReadVanitySuccessCount:   &EmptyCounter{},
	ReadVanityFailureCount:   &EmptyCounter{},
}

func NewVanityServiceMetrics(ctx context.Context, metricsHandler Handler) (*VanityServiceMetrics, error) {
	var err error

	vanityMetrics := VanityServiceMetrics{}

	vanityMetrics.Goroutines, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Vanity Metrics Goroutine Count",
		ServiceName: "vanity_metrics",
		ID:          "vanity_metrics.goroutines",
		Unit:        "goroutines",
		Description: "The number of goroutines the vanity_metrics is using",
	})
	if err != nil {
		return nil, err
	}

	vanityMetrics.MemoryAllocated, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Vanity Metrics Memory Allocated",
		ServiceName: "vanity_metrics",
		ID:          "vanity_metrics.memory",
		Unit:        "MB",
		Description: "The amount of memory the vanity_metrics has allocated in MB",
	})
	if err != nil {
		return nil, err
	}

	vanityMetrics.ReceivedVanityCount, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Vanity Metrics Received Count",
		ServiceName: "vanity_metrics",
		ID:          "vanity.metrics.received.count",
		Unit:        "reads",
		Description: "The number of successful vanity metrics received from ZeroMQ",
	})
	if err != nil {
		return nil, err
	}

	vanityMetrics.UpdateVanitySuccessCount, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Vanity Metrics Update Success Count",
		ServiceName: "vanity_metrics",
		ID:          "vanity.metrics.update.success.count",
		Unit:        "updates",
		Description: "The number of successful vanity metric updates to the metrics handler",
	})
	if err != nil {
		return nil, err
	}

	vanityMetrics.UpdateVanityFailureCount, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Vanity Metrics Update Failure Count",
		ServiceName: "vanity_metrics",
		ID:          "vanity.metrics.update.failure.count",
		Unit:        "updates",
		Description: "The number of failed vanity metric updates to the metrics handler",
	})
	if err != nil {
		return nil, err
	}

	vanityMetrics.ReadVanitySuccessCount, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Vanity Metrics Read Success Count",
		ServiceName: "vanity_metrics",
		ID:          "vanity.metrics.read.success.count",
		Unit:        "reads",
		Description: "The number of successful vanity metric reads from StackDriver",
	})

	vanityMetrics.ReadVanityFailureCount, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Vanity Metrics Read Failure Count",
		ServiceName: "vanity_metrics",
		ID:          "vanity.metrics.read.failure.count",
		Unit:        "reads",
		Description: "The number of failed vanity metric reads from StackDriver",
	})

	return &vanityMetrics, nil
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

	billingServiceMetrics.BillingMetrics.BillingEntrySize, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Billing Entry Size",
		ServiceName: "billing",
		ID:          "billing.entry.size",
		Unit:        "bytes",
		Description: "The size of a billing entry",
	})
	if err != nil {
		return nil, err
	}

	billingServiceMetrics.BillingMetrics.PubsubBillingEntrySize, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Pubsub Billing Entry Size",
		ServiceName: "billing",
		ID:          "pubsub.billing.entry.size",
		Unit:        "bytes",
		Description: "The size of a pubsub billing entry",
	})
	if err != nil {
		return nil, err
	}

	billingServiceMetrics.BillingMetrics.Entries2Received, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Billing Entries 2 Received",
		ServiceName: "billing",
		ID:          "billing.entries.2",
		Unit:        "entries",
		Description: "The total number of billing entries 2 received through Google Pub/Sub",
	})
	if err != nil {
		return nil, err
	}

	billingServiceMetrics.BillingMetrics.Entries2Submitted, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Billing Entries 2 Submitted",
		ServiceName: "billing",
		ID:          "billing.entries.2.submitted",
		Unit:        "entries",
		Description: "The total number of billing entries 2 submitted to BigQuery",
	})
	if err != nil {
		return nil, err
	}

	billingServiceMetrics.BillingMetrics.Entries2Queued, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Billing Entries 2 Queued",
		ServiceName: "billing",
		ID:          "billing.entries.2.queued",
		Unit:        "entries",
		Description: "The total number of billing entries 2 waiting to be sent to BigQuery",
	})
	if err != nil {
		return nil, err
	}

	billingServiceMetrics.BillingMetrics.Entries2Flushed, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Billing Entries 2 Written",
		ServiceName: "billing",
		ID:          "billing.entries.2.written",
		Unit:        "entries",
		Description: "The total number of billing entries 2 written to BigQuery",
	})
	if err != nil {
		return nil, err
	}

	billingServiceMetrics.BillingMetrics.SummaryEntries2Submitted, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Billing Summary Entries 2 Submitted",
		ServiceName: "billing",
		ID:          "billing.summary.entries.2.submitted",
		Unit:        "entries",
		Description: "The total number of billing summary entries 2 submitted to BigQuery",
	})
	if err != nil {
		return nil, err
	}

	billingServiceMetrics.BillingMetrics.SummaryEntries2Queued, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Billing Summary Entries 2 Queued",
		ServiceName: "billing",
		ID:          "billing.summary.entries.2.queued",
		Unit:        "entries",
		Description: "The total number of billing summary entries 2 waiting to be sent to BigQuery",
	})
	if err != nil {
		return nil, err
	}

	billingServiceMetrics.BillingMetrics.SummaryEntries2Flushed, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Billing Summary Entries 2 Written",
		ServiceName: "billing",
		ID:          "billing.summary.entries.2.written",
		Unit:        "entries",
		Description: "The total number of billing summary entries 2 written to BigQuery",
	})
	if err != nil {
		return nil, err
	}

	billingServiceMetrics.BillingMetrics.BillingEntry2Size, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Billing Entry 2 Size",
		ServiceName: "billing",
		ID:          "billing.entry.2.size",
		Unit:        "bytes",
		Description: "The size of a billing entry 2",
	})
	if err != nil {
		return nil, err
	}

	billingServiceMetrics.BillingMetrics.PubsubBillingEntry2Size, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Pubsub Billing Entry 2 Size",
		ServiceName: "billing",
		ID:          "pubsub.billing.entry.2.size",
		Unit:        "bytes",
		Description: "The size of a pubsub billing entry 2",
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

	billingServiceMetrics.BillingMetrics.ErrorMetrics.BillingInvalidEntries, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Billing Invalid Entries",
		ServiceName: "billing",
		ID:          "billing.error.invalid_entries",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	billingServiceMetrics.BillingMetrics.ErrorMetrics.BillingEntriesWithNaN, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Billing Entries with NaN",
		ServiceName: "billing",
		ID:          "billing.error.billing_entries_with_nan",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	billingServiceMetrics.BillingMetrics.ErrorMetrics.Billing2PublishFailure = &EmptyCounter{}

	billingServiceMetrics.BillingMetrics.ErrorMetrics.Billing2ReadFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Billing 2 Read Failure",
		ServiceName: "billing",
		ID:          "billing.error.read_failure_2",
		Unit:        "errors",
		Description: "The number of times a billing entry 2 failed to be read",
	})
	if err != nil {
		return nil, err
	}

	billingServiceMetrics.BillingMetrics.ErrorMetrics.Billing2BatchedReadFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Billing 2 Batched Read Failure",
		ServiceName: "billing",
		ID:          "billing.error.batched_read_failure_2",
		Unit:        "errors",
		Description: "The number of times a batch of billing entry 2s failed to be read",
	})
	if err != nil {
		return nil, err
	}

	billingServiceMetrics.BillingMetrics.ErrorMetrics.Billing2WriteFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Billing 2 Write Failure",
		ServiceName: "billing",
		ID:          "billing.error.write_failure_2",
		Unit:        "errors",
		Description: "The number of times a billing entry 2 failed to be written",
	})
	if err != nil {
		return nil, err
	}

	billingServiceMetrics.BillingMetrics.ErrorMetrics.Billing2InvalidEntries, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Billing 2 Invalid Entries",
		ServiceName: "billing",
		ID:          "billing.error.invalid_entries_2",
		Unit:        "errors",
		Description: "The number of times a billing entry 2 had invalid values",
	})
	if err != nil {
		return nil, err
	}

	billingServiceMetrics.BillingMetrics.ErrorMetrics.Billing2EntriesWithNaN, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Billing 2 Entries with NaN",
		ServiceName: "billing",
		ID:          "billing.error.billing_entries_with_nan_2",
		Unit:        "errors",
		Description: "The number of times a billing entry 2 had NaN values",
	})
	if err != nil {
		return nil, err
	}

	return &billingServiceMetrics, nil
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

	//RelayNamesHash
	analyticsMetrics.RouteMatrixStatsMetrics.EntriesReceived, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Route Matrix Stats Entries Received",
		ServiceName: "analytics",
		ID:          "analytics.route_matrix_stats.entries",
		Unit:        "entries",
		Description: "The total number of Route Matrix Stats entries received through Google Pub/Sub",
	})
	if err != nil {
		return nil, err
	}

	analyticsMetrics.RouteMatrixStatsMetrics.EntriesSubmitted, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Route Matrix Stats Entries Submitted",
		ServiceName: "analytics",
		ID:          "analytics.route_matrix_stats.entries.submitted",
		Unit:        "entries",
		Description: "The total number of relay stats entries submitted to BigQuery",
	})
	if err != nil {
		return nil, err
	}

	analyticsMetrics.RouteMatrixStatsMetrics.EntriesQueued, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Route Matrix Stats Entries Queued",
		ServiceName: "analytics",
		ID:          "analytics.route_matrix_stats.entries.queued",
		Unit:        "entries",
		Description: "The total number of relay stats entries waiting to be sent to BigQuery",
	})
	if err != nil {
		return nil, err
	}

	analyticsMetrics.RouteMatrixStatsMetrics.EntriesFlushed, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Route Matrix Stats Entries Flushed",
		ServiceName: "analytics",
		ID:          "analytics.route_matrix_stats.entries.flushed",
		Unit:        "entries",
		Description: "The total number of relay stats entries written to BigQuery",
	})
	if err != nil {
		return nil, err
	}

	analyticsMetrics.RouteMatrixStatsMetrics.ErrorMetrics.PublishFailure = &EmptyCounter{}

	analyticsMetrics.RouteMatrixStatsMetrics.ErrorMetrics.ReadFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Route Matrix Stats Read Failure",
		ServiceName: "analytics",
		ID:          "analytics.route_matrix_stats.error.read_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	analyticsMetrics.RouteMatrixStatsMetrics.ErrorMetrics.WriteFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Route Matrix Stats Write Failure",
		ServiceName: "analytics",
		ID:          "analytics.route_matrix_stats.error.write_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	return &analyticsMetrics, nil
}

func NewBuyerEndpointMetrics(ctx context.Context, metricsHandler Handler) (*BuyerEndpointMetrics, error) {
	buyerEndpointMetrics := BuyerEndpointMetrics{}
	var err error

	buyerEndpointMetrics.NoSlicesFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Session has no slice data",
		ServiceName: "buyerEndpoint",
		ID:          "buyerEndpoint.slices.empty",
		Unit:        "sessions",
		Description: "The total number of sessions with out slices",
	})
	if err != nil {
		return nil, err
	}

	return &buyerEndpointMetrics, nil
}
