package metrics

import (
	"context"
)

// MatchDataStatus defines the metrics reported by the service's status endpoint
type MatchDataStatus struct {
	// Service Information
	ServiceName string `json:"service_name"`
	GitHash     string `json:"git_hash"`
	Started     string `json:"started"`
	Uptime      string `json:"uptime"`

	// Metrics
	Goroutines      int     `json:"goroutines"`
	MemoryAllocated float64 `json:"mb_allocated"`

	// Match Data Entries
	MatchDataEntriesReceived  int `json:"match_data_entries_received"`
	MatchDataEntriesSubmitted int `json:"match_data_entries_submitted"`
	MatchDataEntriesQueued    int `json:"match_data_entries_queued"`
	MatchDataEntriesFlushed   int `json:"match_data_entries_flushed"`

	// Match Data Errors
	MatchDataEntriesWithNaN int `json:"match_data_entries_with_nan"`
	MatchDataInvalidEntries int `json:"match_data_invalid_entries"`
	MatchDataReadFailures   int `json:"match_data_read_failures"`
	MatchDataWriteFailures  int `json:"match_data_write_failures"`
}

type MatchDataServiceMetrics struct {
	ServiceMetrics    *ServiceMetrics
	MatchDataMetrics  *MatchDataMetrics
}

var EmptyMatchDataServiceMetrics MatchDataServiceMetrics = MatchDataServiceMetrics{
	ServiceMetrics:    &EmptyServiceMetrics,
	MatchDataMetrics:  &EmptyMatchDataMetrics,
}

type MatchDataMetrics struct {
	EntriesReceived         Counter
	EntriesSubmitted        Counter
	EntriesQueued           Gauge
	EntriesFlushed          Counter
	MatchDataEntryPubsubSize  Gauge
	MatchDataEntrySize        Gauge

	ErrorMetrics *MatchDataErrorMetrics
}

var EmptyMatchDataMetrics MatchDataMetrics = MatchDataMetrics{
	EntriesReceived:         &EmptyCounter{},
	EntriesSubmitted:        &EmptyCounter{},
	EntriesQueued:           &EmptyGauge{},
	EntriesFlushed:          &EmptyCounter{},
	MatchDataEntryPubsubSize:  &EmptyGauge{},
	MatchDataEntrySize:        &EmptyGauge{},

	ErrorMetrics: &EmptyMatchDataErrorMetrics,
}

type MatchDataErrorMetrics struct {
	MatchDataPublishFailure     Counter
	MatchDataReadFailure        Counter
	MatchDataBatchedReadFailure Counter
	MatchDataWriteFailure       Counter
	MatchDataInvalidEntries     Counter
	MatchDataEntriesWithNaN     Counter
	MatchDataRetryLimitReached  Counter
}

var EmptyMatchDataErrorMetrics MatchDataErrorMetrics = MatchDataErrorMetrics{
	MatchDataPublishFailure:     &EmptyCounter{},
	MatchDataReadFailure:        &EmptyCounter{},
	MatchDataBatchedReadFailure: &EmptyCounter{},
	MatchDataWriteFailure:       &EmptyCounter{},
	MatchDataInvalidEntries:     &EmptyCounter{},
	MatchDataEntriesWithNaN:     &EmptyCounter{},
	MatchDataRetryLimitReached:  &EmptyCounter{},
}

func NewMatchDataServiceMetrics(ctx context.Context, metricsHandler Handler, serviceName string, handlerID string, handlerName string, packetDescription string) (*MatchDataServiceMetrics, error) {
	matchDataServiceMetrics := MatchDataServiceMetrics{}
	var err error

	matchDataServiceMetrics.ServiceMetrics, err = NewServiceMetrics(ctx, metricsHandler, serviceName)
	if err != nil {
		return nil, err
	}

	matchDataServiceMetrics.MatchDataMetrics, err = NewMatchDataMetrics(ctx, metricsHandler, serviceName, handlerID, handlerName, packetDescription)
	if err != nil {
		return nil, err
	}
	matchDataServiceMetrics.MatchDataMetrics.ErrorMetrics.MatchDataPublishFailure = &EmptyCounter{}

	return &matchDataServiceMetrics, nil
}

func NewMatchDataMetrics(ctx context.Context, metricsHandler Handler, serviceName string, handlerID string, handlerName string, packetDescription string) (*MatchDataMetrics, error) {
	matchDataMetrics := MatchDataMetrics{}
	var err error

	matchDataMetrics.EntriesReceived, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Entries Received",
		ServiceName: serviceName,
		ID:          handlerID + ".entries",
		Unit:        "entries",
		Description: "The total number of " + packetDescription + "s received through Google Pub/Sub",
	})
	if err != nil {
		return nil, err
	}

	matchDataMetrics.EntriesSubmitted, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Entries Submitted",
		ServiceName: serviceName,
		ID:          handlerID + ".entries.submitted",
		Unit:        "entries",
		Description: "The total number of " + packetDescription + "s submitted to BigQuery",
	})
	if err != nil {
		return nil, err
	}

	matchDataMetrics.EntriesQueued, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: handlerName + " Entries Queued",
		ServiceName: serviceName,
		ID:          handlerID + ".entries.queued",
		Unit:        "entries",
		Description: "The total number of " + packetDescription + "s waiting to be sent to BigQuery",
	})
	if err != nil {
		return nil, err
	}

	matchDataMetrics.EntriesFlushed, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Entries Written",
		ServiceName: serviceName,
		ID:          handlerID + ".entries.written",
		Unit:        "entries",
		Description: "The total number of " + packetDescription + "s written to BigQuery",
	})
	if err != nil {
		return nil, err
	}

	matchDataMetrics.MatchDataEntrySize, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: handlerName + " Entry Size",
		ServiceName: serviceName,
		ID:          handlerID + ".entry.size",
		Unit:        "bytes",
		Description: "The size of a " + packetDescription,
	})
	if err != nil {
		return nil, err
	}

	matchDataMetrics.MatchDataEntryPubsubSize, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: handlerName + " Entry Pubsub Size",
		ServiceName: serviceName,
		ID:          handlerID + ".pubsub.size",
		Unit:        "bytes",
		Description: "The size of a pubsub " + packetDescription + "",
	})
	if err != nil {
		return nil, err
	}

	matchDataMetrics.ErrorMetrics.MatchDataPublishFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Publish Failure",
		ServiceName: serviceName,
		ID:          handlerID + ".error.publish_failure",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " message failed to be published to Google Pub/Sub",
	})
	if err != nil {
		return nil, err
	}

	matchDataMetrics.ErrorMetrics.MatchDataReadFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Read Failure",
		ServiceName: serviceName,
		ID:          handlerID + ".error.read_failure",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " failed to be read",
	})
	if err != nil {
		return nil, err
	}

	matchDataMetrics.ErrorMetrics.MatchDataBatchedReadFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Batched Read Failure",
		ServiceName: serviceName,
		ID:          handlerID + ".error.batched_read_failure",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " message failed to be read",
	})
	if err != nil {
		return nil, err
	}

	matchDataMetrics.ErrorMetrics.MatchDataWriteFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Write Failure",
		ServiceName: serviceName,
		ID:          handlerID + ".error.write_failure",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " failed to be written",
	})
	if err != nil {
		return nil, err
	}

	matchDataMetrics.ErrorMetrics.MatchDataInvalidEntries, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Invalid Entries",
		ServiceName: serviceName,
		ID:          handlerID + ".error.invalid_entries",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " had invalid values",
	})
	if err != nil {
		return nil, err
	}

	matchDataMetrics.ErrorMetrics.MatchDataEntriesWithNaN, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Entries with NaN",
		ServiceName: serviceName,
		ID:          handlerID + ".error.entries_with_nan",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " had NaN values",
	})
	if err != nil {
		return nil, err
	}

	matchDataMetrics.ErrorMetrics.MatchDataRetryLimitReached, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Retry Limit Reached",
		ServiceName: serviceName,
		ID:          handlerID + ".error.retry_limit_reached",
		Unit:        "errors",
		Description: "The number of times a " + packetDescription + " message could not be fully submitted to the internal buffer and was nacked",
	})
	if err != nil {
		return nil, err
	}

	return &matchDataMetrics, nil
}
