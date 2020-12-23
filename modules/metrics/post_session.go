package metrics

import "context"

// PostSessionMetrics defines the set of metrics for the post session update processing.
type PostSessionMetrics struct {
	BillingEntriesSent     Counter
	BillingEntriesFinished Counter
	BillingBufferLength    Gauge
	BillingBufferFull      Counter
	PortalEntriesSent      Counter
	PortalEntriesFinished  Counter
	PortalBufferLength     Gauge
	PortalBufferFull       Counter
	VanityMetricsSent      Counter
	VanityMetricsFinished  Counter
	VanityBufferLength     Gauge
	VanityBufferFull       Counter

	BillingFailure        Counter
	PortalFailure         Counter
	VanityMarshalFailure  Counter
	VanityTransmitFailure Counter
}

// EmptyPostSessionMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyPostSessionMetrics = PostSessionMetrics{
	BillingEntriesSent:     &EmptyCounter{},
	BillingEntriesFinished: &EmptyCounter{},
	BillingBufferLength:    &EmptyGauge{},
	BillingBufferFull:      &EmptyCounter{},
	PortalEntriesSent:      &EmptyCounter{},
	PortalEntriesFinished:  &EmptyCounter{},
	PortalBufferLength:     &EmptyGauge{},
	PortalBufferFull:       &EmptyCounter{},
	VanityMetricsSent:      &EmptyCounter{},
	VanityMetricsFinished:  &EmptyCounter{},
	VanityBufferLength:     &EmptyGauge{},
	VanityBufferFull:       &EmptyCounter{},
	BillingFailure:         &EmptyCounter{},
	PortalFailure:          &EmptyCounter{},
	VanityMarshalFailure:   &EmptyCounter{},
	VanityTransmitFailure:  &EmptyCounter{},
}

// NewPostSessionMetrics creates the metrics the post session processor will use.
func NewPostSessionMetrics(ctx context.Context, handler Handler, serviceName string) (*PostSessionMetrics, error) {
	var err error
	m := &PostSessionMetrics{}

	m.BillingEntriesSent, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Post Session Update Billing Entries Sent",
		ServiceName: serviceName,
		ID:          "post_session_update.billing_entries_sent",
		Unit:        "entries",
		Description: "The number of billing entries sent to the post session billing channel.",
	})
	if err != nil {
		return nil, err
	}

	m.BillingEntriesFinished, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Post Session Update Billing Entries Finished",
		ServiceName: serviceName,
		ID:          "post_session_update.billing_entries_finished",
		Unit:        "entries",
		Description: "The number of billing entries finished sending to Google Pub/Sub.",
	})
	if err != nil {
		return nil, err
	}

	m.BillingBufferLength, err = handler.NewGauge(ctx, &Descriptor{
		DisplayName: "Post Session Update Billing Entries Length",
		ServiceName: serviceName,
		ID:          "post_session_update.billing_entries_length",
		Unit:        "entries",
		Description: "The number of billing entries in queue waiting to be sent.",
	})
	if err != nil {
		return nil, err
	}

	m.BillingBufferFull, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Post Session Update Billing Buffer Full",
		ServiceName: serviceName,
		ID:          "post_session_update.billing_buffer_full",
		Unit:        "entries",
		Description: "The number of billing entries dropped because the billing queue was full.",
	})
	if err != nil {
		return nil, err
	}

	m.BillingFailure, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Post Session Update Billing Failure",
		ServiceName: serviceName,
		ID:          "post_session_update.billing_failure",
		Unit:        "errors",
		Description: "The number of billing entries that failed to be sent to Google Pub/Sub.",
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

	return m, nil
}
