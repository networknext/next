package metrics

import "context"

type BeaconServiceMetrics struct {
	ServiceMetrics 	ServiceMetrics
	BeaconMetrics   BeaconMetrics
}

var EmptyBeaconServiceMetrics BeaconServiceMetrics = BeaconServiceMetrics{
	ServiceMetrics:  EmptyServiceMetrics,
	BeaconMetrics:   EmptyBeaconMetrics,
}

type BeaconMetrics struct {
	EntriesReceived  Counter
	EntriesSubmitted Counter
	EntriesQueued    Gauge
	EntriesFlushed   Counter
	ErrorMetrics     BeaconErrorMetrics
}

var EmptyBeaconMetrics BeaconMetrics = BeaconMetrics{
	EntriesReceived:  &EmptyCounter{},
	EntriesSubmitted: &EmptyCounter{},
	EntriesQueued:    &EmptyGauge{},
	EntriesFlushed:   &EmptyCounter{},
	ErrorMetrics:     EmptyBeaconErrorMetrics,
}

type BeaconErrorMetrics struct {
	BeaconPublishFailure     Counter
	BeaconReadFailure        Counter
	BeaconBatchedReadFailure Counter
	BeaconWriteFailure       Counter
}

var EmptyBeaconErrorMetrics BeaconErrorMetrics = BeaconErrorMetrics{
	BeaconPublishFailure:     &EmptyCounter{},
	BeaconReadFailure:        &EmptyCounter{},
	BeaconBatchedReadFailure: &EmptyCounter{},
	BeaconWriteFailure:       &EmptyCounter{},
}


func NewBeaconServiceMetrics(ctx context.Context, metricsHandler Handler) (*BeaconServiceMetrics, error) {
	beaconServiceMetrics := BeaconServiceMetrics{}
	var err error

	beaconServiceMetrics.ServiceMetrics, err := &NewServiceMetrics(ctx, metricsHandler, "beacon")
	if err != nil {
		return nil, err
	}

	beaconServiceMetrics.BeaconMetrics.EntriesReceived, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Beacon Entries Received",
		ServiceName: "beacon",
		ID:          "beacon.entries",
		Unit:        "entries",
		Description: "The total number of beacon entries received through Google Pub/Sub",
	})
	if err != nil {
		return nil, err
	}

	beaconServiceMetrics.BeaconMetrics.EntriesSubmitted, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Beacon Entries Submitted",
		ServiceName: "beacon",
		ID:          "beacon.entries.submitted",
		Unit:        "entries",
		Description: "The total number of beacon entries submitted to BigQuery",
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

	beaconServiceMetrics.BeaconMetrics.ErrorMetrics.BeaconPublishFailure = &EmptyCounter{}

	beaconServiceMetrics.BeaconMetrics.ErrorMetrics.BeaconReadFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Beacon Read Failure",
		ServiceName: "beacon",
		ID:          "beacon.error.read_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	beaconServiceMetrics.BeaconMetrics.ErrorMetrics.BeaconBatchedReadFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Beacon Batched Read Failure",
		ServiceName: "beacon",
		ID:          "beacon.error.batched_read_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	beaconServiceMetrics.BeaconMetrics.ErrorMetrics.BeaconWriteFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Beacon Write Failure",
		ServiceName: "beacon",
		ID:          "beacon.error.write_failure",
		Unit:        "errors",
	})
	if err != nil {
		return nil, err
	}

	return &beaconServiceMetrics, nil
}
