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
	EntriesReceived   Counter
	EntriesTransfered Counter
	EntriesSent       Counter
	EntriesSubmitted  Counter
	EntriesQueued     Gauge
	EntriesFlushed    Counter
	ErrorMetrics      BeaconErrorMetrics
}

var EmptyBeaconMetrics BeaconMetrics = BeaconMetrics{
	EntriesReceived:   &EmptyCounter{},
	EntriesTransfered: &EmptyCounter{},
	EntriesSent:       &EmptyCounter{},
	EntriesSubmitted:  &EmptyCounter{},
	EntriesQueued:     &EmptyGauge{},
	EntriesFlushed:    &EmptyCounter{},
	ErrorMetrics:      EmptyBeaconErrorMetrics,
}

type BeaconErrorMetrics struct {
	BeaconPublishFailure     Counter
	BeaconReadFailure        Counter
	BeaconBatchedReadFailure Counter
	BeaconSendFailure        Counter
	BeaconChannelFull        Counter
	BeaconWriteFailure       Counter
}

var EmptyBeaconErrorMetrics BeaconErrorMetrics = BeaconErrorMetrics{
	BeaconPublishFailure:     &EmptyCounter{},
	BeaconReadFailure:        &EmptyCounter{},
	BeaconBatchedReadFailure: &EmptyCounter{},
	BeaconSendFailure:        &EmptyCounter{},
	BeaconChannelFull:        &EmptyCounter{},
	BeaconWriteFailure:       &EmptyCounter{},
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
		Description: "The total number of beacon entries received from the server",
	})
	if err != nil {
		return nil, err
	}

	beaconServiceMetrics.BeaconMetrics.EntriesTransfered, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Beacon Entries Transfered",
		ServiceName: "beacon",
		ID:          "beacon.entries.transfered",
		Unit:        "entries",
		Description: "The total number of beacon entries successfully received through Google Pubsub",
	})
	if err != nil {
		return nil, err
	}

	beaconServiceMetrics.BeaconMetrics.EntriesSent, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Beacon Entries Sent",
		ServiceName: "beacon",
		ID:          "beacon.entries.sent",
		Unit:        "entries",
		Description: "The total number of beacon entries sent to be submitted to Google Pubsub",
	})
	if err != nil {
		return nil, err
	}

	beaconServiceMetrics.BeaconMetrics.EntriesSubmitted, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Beacon Entries Submitted",
		ServiceName: "beacon",
		ID:          "beacon.entries.submitted",
		Unit:        "entries",
		Description: "The total number of beacon entries submitted to Google Pubsub",
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
		Description: "The total number of beacon entries written to Google Pubsub",
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
		Description: "The total number of beacon entries that could not be read by Google Pubsub Forwarder",
	})
	if err != nil {
		return nil, err
	}

	beaconServiceMetrics.BeaconMetrics.ErrorMetrics.BeaconBatchedReadFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Beacon Batched Read Failure",
		ServiceName: "beacon",
		ID:          "beacon.error.batched_read_failure",
		Unit:        "errors",
		Description: "The total number of batched beacon entries that could not be unbatched by Google Pubsub Forwarder",
	})
	if err != nil {
		return nil, err
	}

	beaconServiceMetrics.BeaconMetrics.ErrorMetrics.BeaconSendFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Beacon Send Failure",
		ServiceName: "beacon",
		ID:          "beacon.error.send_failure",
		Unit:        "errors",
		Description: "The total number of beacon entries that could not be submitted to Google Pubsub",
	})
	if err != nil {
		return nil, err
	}

	beaconServiceMetrics.BeaconMetrics.ErrorMetrics.BeaconChannelFull, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Beacon Channel Full",
		ServiceName: "beacon",
		ID:          "beacon.error.channel_full",
		Unit:        "errors",
		Description: "The total number of beacon entries that could not be inserted into the internal channel for submission to Google Pubsub",
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
