package metrics

import "context"

// BeaconMetrics defines a set of metrics for the beacon service.
type BeaconMetrics struct {
	ServiceMetrics ServiceMetrics

	HandlerMetrics RoutineMetrics

	ReadPacketFailure Counter
	UnmarshalFailure  Counter
	BeaconSendFailure Counter
	BeaconChannelFull Counter

	EntriesReceived          Counter
	NonBeaconPacketsReceived Counter
	EntriesSent              Counter
	NextEntries              Counter
	DirectEntries            Counter
	UpgradedEntries          Counter
	NotUpgradedEntries       Counter
	EnabledEntries           Counter
	NotEnabledEntries        Counter
	FallbackToDirect         Counter

	PublishMetrics PublisherMetrics
}

// EmptyBeaconMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyBeaconMetrics BeaconMetrics = BeaconMetrics{
	ServiceMetrics: EmptyServiceMetrics,

	HandlerMetrics: EmptyRoutineMetrics,

	ReadPacketFailure: &EmptyCounter{},
	UnmarshalFailure:  &EmptyCounter{},
	BeaconSendFailure: &EmptyCounter{},
	BeaconChannelFull: &EmptyCounter{},

	NonBeaconPacketsReceived: &EmptyCounter{},
	EntriesSent:              &EmptyCounter{},
	NextEntries:              &EmptyCounter{},
	DirectEntries:            &EmptyCounter{},
	UpgradedEntries:          &EmptyCounter{},
	NotUpgradedEntries:       &EmptyCounter{},
	EnabledEntries:           &EmptyCounter{},
	NotEnabledEntries:        &EmptyCounter{},
	FallbackToDirect:         &EmptyCounter{},

	PublishMetrics: EmptyPublisherMetrics,
}

// NewBeaconMetrics creates the metrics that the beacon service will use.
func NewBeaconMetrics(ctx context.Context, handler Handler) (*BeaconMetrics, error) {
	serviceName := "beacon"

	m := BeaconMetrics{}
	var err error

	m.ServiceMetrics, err = NewServiceMetrics(ctx, handler, serviceName)
	if err != nil {
		return nil, err
	}

	m.HandlerMetrics, err = NewRoutineMetrics(ctx, handler, serviceName, "beacon", "Beacon", "beacon update request")
	if err != nil {
		return nil, err
	}

	m.ReadPacketFailure, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Beacon Read Packet Failure",
		ServiceName: serviceName,
		ID:          "read_packet_failure",
		Unit:        "packets",
		Description: "The total number of packets the beacon service failed to read.",
	})
	if err != nil {
		return nil, err
	}

	m.UnmarshalFailure, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Beacon Unmarshal Failure",
		ServiceName: serviceName,
		ID:          "unmarshal_failure",
		Unit:        "packets",
		Description: "The total number of beacon packets that could not be unmarshaled.",
	})
	if err != nil {
		return nil, err
	}

	m.BeaconSendFailure, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Beacon Send Failure",
		ServiceName: serviceName,
		ID:          "send_failure",
		Unit:        "errors",
		Description: "The total number of beacon entries that could not be submitted to Google Pubsub.",
	})
	if err != nil {
		return nil, err
	}

	m.BeaconChannelFull, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Beacon Channel Full",
		ServiceName: serviceName,
		ID:          "channel_full",
		Unit:        "errors",
		Description: "The total number of beacon entries that could not be inserted into the internal channel for submission to Google Pubsub.",
	})
	if err != nil {
		return nil, err
	}

	m.EntriesReceived, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Beacon Entries Received",
		ServiceName: serviceName,
		ID:          "entries_received",
		Unit:        "entries",
		Description: "The total number of beacon entries successfully received from the server.",
	})
	if err != nil {
		return nil, err
	}

	m.NonBeaconPacketsReceived, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Non Beacon Packets Received",
		ServiceName: serviceName,
		ID:          "non_beacon_packets_received",
		Unit:        "packets",
		Description: "The total number of non beacon packets received by the beacon service.",
	})
	if err != nil {
		return nil, err
	}

	m.EntriesSent, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Beacon Entries Sent",
		ServiceName: serviceName,
		ID:          "entries_sent",
		Unit:        "entries",
		Description: "The total number of beacon entries sent to be submitted to Google Pubsub.",
	})
	if err != nil {
		return nil, err
	}

	m.NextEntries, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Beacon Entries on Next",
		ServiceName: serviceName,
		ID:          "entries_on_next",
		Unit:        "entries",
		Description: "The total number of beacon entries received that are on network next.",
	})
	if err != nil {
		return nil, err
	}

	m.DirectEntries, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Beacon Entries on Direct",
		ServiceName: serviceName,
		ID:          "entries_on_direct",
		Unit:        "entries",
		Description: "The total number of beacon entries received that are on direct.",
	})
	if err != nil {
		return nil, err
	}

	m.UpgradedEntries, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Beacon Entries Upgraded",
		ServiceName: serviceName,
		ID:          "entries_upgraded",
		Unit:        "entries",
		Description: "The total number of beacon entries received that are upgraded.",
	})
	if err != nil {
		return nil, err
	}

	m.NotUpgradedEntries, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Beacon Entries Not Upgraded",
		ServiceName: serviceName,
		ID:          "entries_not_upgraded",
		Unit:        "entries",
		Description: "The total number of beacon entries received that are on not upgraded.",
	})
	if err != nil {
		return nil, err
	}

	m.EnabledEntries, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Beacon Entries Enabled",
		ServiceName: serviceName,
		ID:          "entries_enabled",
		Unit:        "entries",
		Description: "The total number of beacon entries received that are enabled.",
	})
	if err != nil {
		return nil, err
	}

	m.NotEnabledEntries, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Beacon Entries Not Enabled",
		ServiceName: serviceName,
		ID:          "entries_not_enabled",
		Unit:        "entries",
		Description: "The total number of beacon entries received that are on not enabled.",
	})
	if err != nil {
		return nil, err
	}

	m.FallbackToDirect, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Beacon Entries Fallback To Direct",
		ServiceName: serviceName,
		ID:          "entries_fallback_to_direct",
		Unit:        "entries",
		Description: "The total number of beacon entries received that have fallen back to direct.",
	})
	if err != nil {
		return nil, err
	}

	m.PublishMetrics, err = NewPublisherMetrics(ctx, handler, serviceName, "beacon", "Beacon", "beacon")
	if err != nil {
		return nil, err
	}

	return &m, nil
}
