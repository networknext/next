package metrics

import "context"

// BeaconServiceMetrics defines a set of metrics for the beacon service.
type BeaconServiceMetrics struct {
	ServiceMetrics *ServiceMetrics
	HandlerMetrics *PacketHandlerMetrics
	BeaconMetrics  *BeaconMetrics
}

// EmptyBeaconServiceMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyBeaconServiceMetrics BeaconServiceMetrics = BeaconServiceMetrics{
	ServiceMetrics: &EmptyServiceMetrics,
	HandlerMetrics: &EmptyPacketHandlerMetrics,
	BeaconMetrics:  &EmptyBeaconMetrics,
}

// BeconMetrics defines a set of metrics for monitoring beacon packets.
type BeaconMetrics struct {
	NonBeaconPacketsReceived Counter
	EntriesReceived          Counter
	EntriesSent              Counter
	EntriesSubmitted         Counter
	EntriesFlushed           Counter
	NextEntries              Counter
	DirectEntries            Counter
	UpgradedEntries          Counter
	NotUpgradedEntries       Counter
	EnabledEntries           Counter
	NotEnabledEntries        Counter
	FallbackToDirect         Counter
	ErrorMetrics             BeaconErrorMetrics
}

// EmptyBeaconMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyBeaconMetrics BeaconMetrics = BeaconMetrics{
	NonBeaconPacketsReceived: &EmptyCounter{},
	EntriesReceived:          &EmptyCounter{},
	EntriesSent:              &EmptyCounter{},
	EntriesSubmitted:         &EmptyCounter{},
	EntriesFlushed:           &EmptyCounter{},
	NextEntries:              &EmptyCounter{},
	DirectEntries:            &EmptyCounter{},
	UpgradedEntries:          &EmptyCounter{},
	NotUpgradedEntries:       &EmptyCounter{},
	EnabledEntries:           &EmptyCounter{},
	NotEnabledEntries:        &EmptyCounter{},
	FallbackToDirect:         &EmptyCounter{},
	ErrorMetrics:             EmptyBeaconErrorMetrics,
}

// BeaconErrorMetrics defines a set of metrics for recording errors for the beacon service.
type BeaconErrorMetrics struct {
	BeaconReadPacketFailure      Counter
	BeaconSerializePacketFailure Counter
	BeaconPublishFailure         Counter
	BeaconSendFailure            Counter
	BeaconChannelFull            Counter
}

// EmptyBeaconErrorMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyBeaconErrorMetrics BeaconErrorMetrics = BeaconErrorMetrics{
	BeaconReadPacketFailure:      &EmptyCounter{},
	BeaconSerializePacketFailure: &EmptyCounter{},
	BeaconPublishFailure:         &EmptyCounter{},
	BeaconSendFailure:            &EmptyCounter{},
	BeaconChannelFull:            &EmptyCounter{},
}

// NewBeaconServiceMetrics creates the metrics that the beacon service will use.
func NewBeaconServiceMetrics(ctx context.Context, metricsHandler Handler) (*BeaconServiceMetrics, error) {
	beaconServiceMetrics := BeaconServiceMetrics{}
	var err error

	beaconServiceMetrics.ServiceMetrics, err = NewServiceMetrics(ctx, metricsHandler, "beacon")
	if err != nil {
		return nil, err
	}

	beaconServiceMetrics.HandlerMetrics, err = NewPacketHandlerMetrics(ctx, metricsHandler, "beacon", "beacon", "Beacon", "beacon packet")
	if err != nil {
		return nil, err
	}

	beaconServiceMetrics.BeaconMetrics = &BeaconMetrics{}
	beaconServiceMetrics.BeaconMetrics.ErrorMetrics = BeaconErrorMetrics{}

	beaconServiceMetrics.BeaconMetrics.NonBeaconPacketsReceived, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Non Beacon Packets Received",
		ServiceName: "beacon",
		ID:          "beacon.non.beacon.packets.received",
		Unit:        "packets",
		Description: "The total number of non beacon packets received by the beacon service",
	})
	if err != nil {
		return nil, err
	}

	beaconServiceMetrics.BeaconMetrics.EntriesReceived, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Beacon Entries Received",
		ServiceName: "beacon",
		ID:          "beacon.entries.received",
		Unit:        "entries",
		Description: "The total number of beacon entries successfully received from the server",
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

	beaconServiceMetrics.BeaconMetrics.NextEntries, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Beacon Entries on Next",
		ServiceName: "beacon",
		ID:          "beacon.entries.next",
		Unit:        "entries",
		Description: "The total number of beacon entries received that are on Next",
	})
	if err != nil {
		return nil, err
	}

	beaconServiceMetrics.BeaconMetrics.DirectEntries, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Beacon Entries on Direct",
		ServiceName: "beacon",
		ID:          "beacon.entries.direct",
		Unit:        "entries",
		Description: "The total number of beacon entries received that are on Direct",
	})
	if err != nil {
		return nil, err
	}

	beaconServiceMetrics.BeaconMetrics.UpgradedEntries, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Beacon Entries Upgraded",
		ServiceName: "beacon",
		ID:          "beacon.entries.upgraded",
		Unit:        "entries",
		Description: "The total number of beacon entries received that are Upgraded",
	})
	if err != nil {
		return nil, err
	}

	beaconServiceMetrics.BeaconMetrics.NotUpgradedEntries, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Beacon Entries Not Upgraded",
		ServiceName: "beacon",
		ID:          "beacon.entries.not_upgraded",
		Unit:        "entries",
		Description: "The total number of beacon entries received that are on not Upgraded",
	})
	if err != nil {
		return nil, err
	}

	beaconServiceMetrics.BeaconMetrics.EnabledEntries, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Beacon Entries Enabled",
		ServiceName: "beacon",
		ID:          "beacon.entries.enabled",
		Unit:        "entries",
		Description: "The total number of beacon entries received that are Enabled",
	})
	if err != nil {
		return nil, err
	}

	beaconServiceMetrics.BeaconMetrics.NotEnabledEntries, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Beacon Entries Not Enabled",
		ServiceName: "beacon",
		ID:          "beacon.entries.not_enabled",
		Unit:        "entries",
		Description: "The total number of beacon entries received that are on not Enabled",
	})
	if err != nil {
		return nil, err
	}

	beaconServiceMetrics.BeaconMetrics.FallbackToDirect, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Beacon Entries Fallback To Direct",
		ServiceName: "beacon",
		ID:          "beacon.entries.fallback_to_direct",
		Unit:        "entries",
		Description: "The total number of beacon entries received that have fallen back to Direct",
	})
	if err != nil {
		return nil, err
	}

	beaconServiceMetrics.BeaconMetrics.ErrorMetrics.BeaconReadPacketFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Beacon Read Packet Failure",
		ServiceName: "beacon",
		ID:          "beacon.read.packet.failure",
		Unit:        "packets",
		Description: "The total number of packets the beacon service failed to read",
	})
	if err != nil {
		return nil, err
	}

	beaconServiceMetrics.BeaconMetrics.ErrorMetrics.BeaconSerializePacketFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Beacon Serialize Packet Failure",
		ServiceName: "beacon",
		ID:          "beacon.serialize.packet.failure",
		Unit:        "packets",
		Description: "The total number of beacon packets that could not be serialized",
	})
	if err != nil {
		return nil, err
	}

	beaconServiceMetrics.BeaconMetrics.ErrorMetrics.BeaconPublishFailure, err = metricsHandler.NewCounter(ctx, &Descriptor{
		DisplayName: "Beacon Publish Failure",
		ServiceName: "beacon",
		ID:          "beacon.error.publish_failure",
		Unit:        "errors",
		Description: "The total number of batched beacon entries that could not be published to Google Pubsub",
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

	return &beaconServiceMetrics, nil
}
