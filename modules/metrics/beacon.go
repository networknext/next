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
	EntriesSent       Counter
	EntriesSubmitted  Counter
	EntriesFlushed    Counter
	ErrorMetrics      BeaconErrorMetrics
}

var EmptyBeaconMetrics BeaconMetrics = BeaconMetrics{
	EntriesReceived:   &EmptyCounter{},
	EntriesSent:       &EmptyCounter{},
	EntriesSubmitted:  &EmptyCounter{},
	EntriesFlushed:    &EmptyCounter{},
	ErrorMetrics:      EmptyBeaconErrorMetrics,
}

type BeaconErrorMetrics struct {
	BeaconPublishFailure     Counter
	BeaconSendFailure        Counter
	BeaconChannelFull        Counter
}

var EmptyBeaconErrorMetrics BeaconErrorMetrics = BeaconErrorMetrics{
	BeaconPublishFailure:     &EmptyCounter{},
	BeaconSendFailure:        &EmptyCounter{},
	BeaconChannelFull:        &EmptyCounter{},
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

	return &beaconServiceMetrics, nil
}
