package metrics

import "context"

// BeaconInserterServiceMetrics defines a set of metrics for the beacon insertion service.
type BeaconInserterServiceMetrics struct {
	ServiceMetrics        *ServiceMetrics
	BeaconInserterMetrics *BeaconInserterMetrics
}

// EmptyBeaconInserterServiceMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyBeaconInserterServiceMetrics BeaconInserterServiceMetrics = BeaconInserterServiceMetrics{
	ServiceMetrics:        &EmptyServiceMetrics,
	BeaconInserterMetrics: &EmptyBeaconInserterMetrics,
}

// BeaconInserterMetrics defines a set of metrics for monitoring the beacon insertion service.
type BeaconInserterMetrics struct {
	EntriesTransfered Counter
	EntriesSubmitted  Counter
	EntriesQueued     Gauge
	EntriesFlushed    Counter
	ErrorMetrics      BeaconInserterErrorMetrics
}

// EmptyBeaconInserterMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyBeaconInserterMetrics BeaconInserterMetrics = BeaconInserterMetrics{
	EntriesSubmitted: &EmptyCounter{},
	EntriesQueued:    &EmptyGauge{},
	EntriesFlushed:   &EmptyCounter{},
	ErrorMetrics:     EmptyBeaconInserterErrorMetrics,
}

// BeaconInserterErrorMetrics defines a set of metrics for recording errors for the beacon insertion service.
type BeaconInserterErrorMetrics struct {
	BeaconInserterReadFailure        Counter
	BeaconInserterBatchedReadFailure Counter
	BeaconInserterWriteFailure       Counter
}

// EmptyBeaconInserterErrorMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyBeaconInserterErrorMetrics BeaconInserterErrorMetrics = BeaconInserterErrorMetrics{
	BeaconInserterReadFailure:        &EmptyCounter{},
	BeaconInserterBatchedReadFailure: &EmptyCounter{},
	BeaconInserterWriteFailure:       &EmptyCounter{},
}

// NewBeaconInserterServiceMetrics creates the metrics that the beacon insertion service will use.
func NewBeaconInserterServiceMetrics(ctx context.Context, metricsHandler Handler) (*BeaconInserterServiceMetrics, error) {
	beaconInserterServiceMetrics := &BeaconInserterServiceMetrics{}
	var err error

	beaconInserterServiceMetrics.ServiceMetrics, err = NewServiceMetrics(ctx, metricsHandler, "beacon_inserter")
	if err != nil {
		return nil, err
	}

	beaconInserterServiceMetrics.BeaconInserterMetrics = &BeaconInserterMetrics{}
	beaconInserterServiceMetrics.BeaconInserterMetrics.ErrorMetrics = BeaconInserterErrorMetrics{}

	beaconInserterServiceMetrics.BeaconInserterMetrics.EntriesTransfered, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Beacon Inserter Entries Transfered",
		ServiceName: "beacon_inserter",
		ID:          "beacon_inserter.entries.transfered",
		Unit:        "entries",
		Description: "The total number of beacon entries successfully received through Google Pubsub to the Google Pubsub Forwarder",
	})
	if err != nil {
		return nil, err
	}

	beaconInserterServiceMetrics.BeaconInserterMetrics.EntriesSubmitted, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Beacon Inserter Entries Submitted",
		ServiceName: "beacon_inserter",
		ID:          "beacon_inserter.entries.submitted",
		Unit:        "entries",
		Description: "The total number of beacon entries submitted to be written to BigQuery",
	})
	if err != nil {
		return nil, err
	}

	beaconInserterServiceMetrics.BeaconInserterMetrics.EntriesQueued, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Beacon Inserter Entries Queued",
		ServiceName: "beacon_inserter",
		ID:          "beacon_inserter.entries.queued",
		Unit:        "entries",
		Description: "The total number of beacon entries waiting to be sent to BigQuery",
	})
	if err != nil {
		return nil, err
	}

	beaconInserterServiceMetrics.BeaconInserterMetrics.EntriesFlushed, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Beacon Inserter Entries Written",
		ServiceName: "beacon_inserter",
		ID:          "beacon_inserter.entries.written",
		Unit:        "entries",
		Description: "The total number of beacon entries written to BigQuery",
	})
	if err != nil {
		return nil, err
	}

	beaconInserterServiceMetrics.BeaconInserterMetrics.ErrorMetrics.BeaconInserterReadFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Beacon Inserter Read Failure",
		ServiceName: "beacon_inserter",
		ID:          "beacon_inserter.error.read_failure",
		Unit:        "errors",
		Description: "The total number of beacon entries that could not be read by Google Pubsub Forwarder",
	})
	if err != nil {
		return nil, err
	}

	beaconInserterServiceMetrics.BeaconInserterMetrics.ErrorMetrics.BeaconInserterBatchedReadFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Beacon Inserter Batched Read Failure",
		ServiceName: "beacon_inserter",
		ID:          "beacon_inserter.error.batched_read_failure",
		Unit:        "errors",
		Description: "The total number of batched beacon entries that could not be unbatched by Google Pubsub Forwarder",
	})
	if err != nil {
		return nil, err
	}

	beaconInserterServiceMetrics.BeaconInserterMetrics.ErrorMetrics.BeaconInserterWriteFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Beacon Inserter Write Failure",
		ServiceName: "beacon_inserter",
		ID:          "beacon_inserter.error.write_failure",
		Unit:        "errors",
		Description: "The total number of beacon entries that could not be written to BigQuery",
	})
	if err != nil {
		return nil, err
	}

	return beaconInserterServiceMetrics, nil
}
