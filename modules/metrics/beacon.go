package metrics

import "context"

type BeaconServiceMetrics struct {
	ServiceMetrics ServiceMetrics
	BeaconMetrics  BeaconMetrics
}

var EmptyBeaconServiceMetrics BeaconServiceMetrics = BeaconServiceMetrics{
	ServiceMetrics: EmptyServiceMetrics,
	BeaconMetrics:  EmptyBeaconMetrics,
}

type BeaconMetrics struct {
	EntriesReceived  Counter
	EntriesSent      Counter
	EntriesSubmitted Counter
	EntriesQueued    Gauge
	EntriesFlushed   Counter
	ErrorMetrics     BeaconErrorMetrics
}

var EmptyBeaconMetrics BeaconMetrics = BeaconMetrics{
	EntriesReceived:  &EmptyCounter{},
	EntriesSent:      &EmptyCounter{},
	EntriesSubmitted: &EmptyCounter{},
	EntriesQueued:    &EmptyGauge{},
	EntriesFlushed:   &EmptyCounter{},
	ErrorMetrics:     EmptyBeaconErrorMetrics,
}

type BeaconErrorMetrics struct {
	// BeaconPublishFailure     Counter
	// BeaconReadFailure        Counter
	// BeaconBatchedReadFailure Counter
	BeaconSubmitFailure Counter
	BeaconChannelFull   Counter
	BeaconWriteFailure  Counter
}

var EmptyBeaconErrorMetrics BeaconErrorMetrics = BeaconErrorMetrics{
	// BeaconPublishFailure:     &EmptyCounter{},
	// BeaconReadFailure:        &EmptyCounter{},
	// BeaconBatchedReadFailure: &EmptyCounter{},
	BeaconSubmitFailure: &EmptyCounter{},
	BeaconChannelFull:   &EmptyCounter{},
	BeaconWriteFailure:  &EmptyCounter{},
}

func NewBeaconServiceMetrics(ctx context.Context, metricsHandler Handler) (*BeaconServiceMetrics, error) {
	beaconServiceMetrics := BeaconServiceMetrics{}
	var err error

	serviceMetrics, err := NewServiceMetrics(ctx, metricsHandler, "beacon")
	if err != nil {
		return nil, err
	}
	beaconServiceMetrics.ServiceMetrics = *serviceMetrics

	beaconServiceMetrics.BeaconMetrics.EntriesReceived, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Beacon Entries Received",
		ServiceName: "beacon",
		ID:          "beacon.entries.received",
		Unit:        "entries",
		Description: "The total number of beacon entries received",
	})
	if err != nil {
		return nil, err
	}

	beaconServiceMetrics.BeaconMetrics.EntriesSent, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Beacon Entries Sent",
		ServiceName: "beacon",
		ID:          "beacon.entries.sent",
		Unit:        "entries",
		Description: "The total number of beacon entries sent to be submitted to BigQuery",
	})
	if err != nil {
		return nil, err
	}

	beaconServiceMetrics.BeaconMetrics.EntriesSubmitted, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Beacon Entries Submitted",
		ServiceName: "beacon",
		ID:          "beacon.entries.submitted",
		Unit:        "entries",
		Description: "The total number of beacon entries submitted to be batch written to BigQuery",
	})
	if err != nil {
		return nil, err
	}

	beaconServiceMetrics.BeaconMetrics.EntriesQueued, err = metricsHandler.NewGauge(ctx, &Descriptor{
		DisplayName: "Beacon Entries Queued",
		ServiceName: "beacon",
		ID:          "beacon.entries.queued",
		Unit:        "entries",
		Description: "The total number of beacon entries waiting to be sent to BigQuery",
	})
	if err != nil {
		return nil, err
	}

	beaconServiceMetrics.BeaconMetrics.EntriesFlushed, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Beacon Entries Written",
		ServiceName: "beacon",
		ID:          "beacon.entries.written",
		Unit:        "entries",
		Description: "The total number of beacon entries written to BigQuery",
	})
	if err != nil {
		return nil, err
	}

	// beaconServiceMetrics.BeaconMetrics.ErrorMetrics.BeaconPublishFailure = &EmptyCounter{}

	beaconServiceMetrics.BeaconMetrics.ErrorMetrics.BeaconSubmitFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Beacon Submit Failure",
		ServiceName: "beacon",
		ID:          "beacon.error.submit_failure",
		Unit:        "errors",
		Description: "The total number of beacon entries that could not be submitted to BigQuery",
	})
	if err != nil {
		return nil, err
	}

	beaconServiceMetrics.BeaconMetrics.ErrorMetrics.BeaconChannelFull, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Beacon Channel Full",
		ServiceName: "beacon",
		ID:          "beacon.error.channel_full",
		Unit:        "errors",
		Description: "The total number of beacon entries that could not be inserted into the internal channel for submission to BigQuery",
	})
	if err != nil {
		return nil, err
	}

	beaconServiceMetrics.BeaconMetrics.ErrorMetrics.BeaconWriteFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Beacon Write Failure",
		ServiceName: "beacon",
		ID:          "beacon.error.write_failure",
		Unit:        "errors",
		Description: "The total number of beacon entries that could not be written to BigQuery",
	})
	if err != nil {
		return nil, err
	}

	return &beaconServiceMetrics, nil
}
