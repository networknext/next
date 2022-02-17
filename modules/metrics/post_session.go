package metrics

import "context"

// PostSessionMetrics defines the set of metrics for the post session update processing.
type PostSessionMetrics struct {
	BillingEntries2Sent     Counter
	BillingEntries2Finished Counter
	Billing2BufferLength    Gauge
	Billing2BufferFull      Counter

	PortalEntriesSent     Counter
	PortalEntriesFinished Counter
	PortalBufferLength    Gauge
	PortalBufferFull      Counter

	VanityMetricsSent     Counter
	VanityMetricsFinished Counter
	VanityBufferLength    Gauge
	VanityBufferFull      Counter

	MatchDataEntriesSent         Counter
	MatchDataEntriesFinished     Counter
	MatchDataEntriesBufferLength Gauge
	MatchDataEntriesBufferFull   Counter

	Billing2Failure         Counter
	PortalFailure           Counter
	VanityMarshalFailure    Counter
	VanityTransmitFailure   Counter
	MatchDataEntriesFailure Counter
}

// EmptyPostSessionMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyPostSessionMetrics = PostSessionMetrics{
	BillingEntries2Sent:          &EmptyCounter{},
	BillingEntries2Finished:      &EmptyCounter{},
	Billing2BufferLength:         &EmptyGauge{},
	Billing2BufferFull:           &EmptyCounter{},
	PortalEntriesSent:            &EmptyCounter{},
	PortalEntriesFinished:        &EmptyCounter{},
	PortalBufferLength:           &EmptyGauge{},
	PortalBufferFull:             &EmptyCounter{},
	VanityMetricsSent:            &EmptyCounter{},
	VanityMetricsFinished:        &EmptyCounter{},
	VanityBufferLength:           &EmptyGauge{},
	VanityBufferFull:             &EmptyCounter{},
	MatchDataEntriesSent:         &EmptyCounter{},
	MatchDataEntriesFinished:     &EmptyCounter{},
	MatchDataEntriesBufferLength: &EmptyGauge{},
	MatchDataEntriesBufferFull:   &EmptyCounter{},
	Billing2Failure:              &EmptyCounter{},
	PortalFailure:                &EmptyCounter{},
	VanityMarshalFailure:         &EmptyCounter{},
	VanityTransmitFailure:        &EmptyCounter{},
	MatchDataEntriesFailure:      &EmptyCounter{},
}

// NewPostSessionMetrics creates the metrics the post session processor will use.
func NewPostSessionMetrics(ctx context.Context, handler Handler, serviceName string) (*PostSessionMetrics, error) {
	var err error
	m := &PostSessionMetrics{}

	m.BillingEntries2Sent, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Post Session Update Billing Entries 2 Sent",
		ServiceName: serviceName,
		ID:          "post_session_update.billing_entries_sent_2",
		Unit:        "entries",
		Description: "The number of billing entries 2 sent to the post session billing 2 channel.",
	})
	if err != nil {
		return nil, err
	}

	m.BillingEntries2Finished, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Post Session Update Billing Entries 2 Finished",
		ServiceName: serviceName,
		ID:          "post_session_update.billing_entries_finished_2",
		Unit:        "entries",
		Description: "The number of billing entries 2 finished sending to Google Pub/Sub.",
	})
	if err != nil {
		return nil, err
	}

	m.Billing2BufferLength, err = handler.NewGauge(ctx, &Descriptor{
		DisplayName: "Post Session Update Billing Entries 2 Length",
		ServiceName: serviceName,
		ID:          "post_session_update.billing_entries_length_2",
		Unit:        "entries",
		Description: "The number of billing entries 2 in queue waiting to be sent.",
	})
	if err != nil {
		return nil, err
	}

	m.Billing2BufferFull, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Post Session Update Billing 2 Buffer Full",
		ServiceName: serviceName,
		ID:          "post_session_update.billing_buffer_full_2",
		Unit:        "entries",
		Description: "The number of billing entries 2 dropped because the billing queue was full.",
	})
	if err != nil {
		return nil, err
	}

	m.Billing2Failure, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Post Session Update Billing 2 Failure",
		ServiceName: serviceName,
		ID:          "post_session_update.billing_failure_2",
		Unit:        "errors",
		Description: "The number of billing entries 2 that failed to be sent to Google Pub/Sub.",
	})
	if err != nil {
		return nil, err
	}

	m.PortalEntriesSent, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Post Session Update Portal Entries Sent",
		ServiceName: serviceName,
		ID:          "post_session_update.portal_entries_sent",
		Unit:        "entries",
		Description: "The number of portal entries sent to the post session portal channel.",
	})
	if err != nil {
		return nil, err
	}

	m.PortalEntriesFinished, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Post Session Update Portal Entries Finished",
		ServiceName: serviceName,
		ID:          "post_session_update.portal_entries_finished",
		Unit:        "entries",
		Description: "The number of billing entries finished sending to the portal_cruncher service.",
	})
	if err != nil {
		return nil, err
	}

	m.PortalBufferLength, err = handler.NewGauge(ctx, &Descriptor{
		DisplayName: "Post Session Update Portal Entries Length",
		ServiceName: serviceName,
		ID:          "post_session_update.portal_entries_length",
		Unit:        "entries",
		Description: "The number of portal entries in queue waiting to be sent.",
	})
	if err != nil {
		return nil, err
	}

	m.PortalBufferFull, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Post Session Update Portal Buffer Full",
		ServiceName: serviceName,
		ID:          "post_session_update.portal_buffer_full",
		Unit:        "entries",
		Description: "The number of portal entries dropped because the portal queue was full.",
	})
	if err != nil {
		return nil, err
	}

	m.PortalFailure, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Post Session Update Portal Failure",
		ServiceName: serviceName,
		ID:          "post_session_update.portal_failure",
		Unit:        "errors",
		Description: "The number of portal entries that failed to be sent to the portal_cruncher service.",
	})
	if err != nil {
		return nil, err
	}

	m.VanityMetricsSent, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Post Session Update Vanity Metrics Sent",
		ServiceName: serviceName,
		ID:          "post_session_update.vanity_metrics_sent",
		Unit:        "entries",
		Description: "The number of billing entries sent to the post session vanity metrics channel.",
	})
	if err != nil {
		return nil, err
	}

	m.VanityMetricsFinished, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Post Session Update Vanity Metrics Finished",
		ServiceName: serviceName,
		ID:          "post_session_update.vanity_metrics_finished",
		Unit:        "metrics",
		Description: "The number of vanity metric structs finished pushing onto ZeroMQ.",
	})
	if err != nil {
		return nil, err
	}

	m.VanityBufferLength, err = handler.NewGauge(ctx, &Descriptor{
		DisplayName: "Post Session Update Vanity Metrics Length",
		ServiceName: serviceName,
		ID:          "post_session_update.vanity_metrics_length",
		Unit:        "entries",
		Description: "The number of billing entries for vanity metrics in queue waiting to be sent.",
	})
	if err != nil {
		return nil, err
	}

	m.VanityBufferFull, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Post Session Update Vanity Buffer Full",
		ServiceName: serviceName,
		ID:          "post_session_update.vanity_buffer_full",
		Unit:        "entries",
		Description: "The number of billing entries dropped because the vanity queue was full.",
	})
	if err != nil {
		return nil, err
	}

	m.VanityMarshalFailure, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Post Session Update Vanity Marshal Failure",
		ServiceName: serviceName,
		ID:          "post_session_update.vanity_marshal_failure",
		Unit:        "errors",
		Description: "The number of entries for vanity metrics that failed to be marshaled.",
	})
	if err != nil {
		return nil, err
	}

	m.VanityTransmitFailure, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Post Session Update Vanity Transmit Failure",
		ServiceName: serviceName,
		ID:          "post_session_update.vanity_transmit_failure",
		Unit:        "errors",
		Description: "The number of marshaled vanity metrics that failed to be pushed onto ZeroMQ.",
	})
	if err != nil {
		return nil, err
	}

	m.MatchDataEntriesSent, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Post Session Update Match Data Entries Sent",
		ServiceName: serviceName,
		ID:          "post_session_update.match_data_entries_sent",
		Unit:        "entries",
		Description: "The number of match data entries sent to the post session match data channel.",
	})
	if err != nil {
		return nil, err
	}

	m.MatchDataEntriesFinished, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Post Session Update Match Data Entries Finished",
		ServiceName: serviceName,
		ID:          "post_session_update.match_data_entries_finished",
		Unit:        "metrics",
		Description: "The number of match data entries finished sending to Google Pub/Sub.",
	})
	if err != nil {
		return nil, err
	}

	m.MatchDataEntriesBufferLength, err = handler.NewGauge(ctx, &Descriptor{
		DisplayName: "Post Session Update Match Data Entries Length",
		ServiceName: serviceName,
		ID:          "post_session_update.match_data_entries_length",
		Unit:        "entries",
		Description: "The number of match data entries in queue waiting to be sent.",
	})
	if err != nil {
		return nil, err
	}

	m.MatchDataEntriesBufferFull, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Post Session Update Match Data Entries Buffer Full",
		ServiceName: serviceName,
		ID:          "post_session_update.match_data_entries_buffer_full",
		Unit:        "entries",
		Description: "The number of match data entries dropped because the match data queue was full.",
	})
	if err != nil {
		return nil, err
	}

	m.MatchDataEntriesFailure, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Post Session Update Match Data Entries Failure",
		ServiceName: serviceName,
		ID:          "post_session_update.match_data_entries_failure",
		Unit:        "errors",
		Description: "The number of match data entries that failed to be sent to Google Pub/Sub.",
	})
	if err != nil {
		return nil, err
	}

	return m, nil
}
