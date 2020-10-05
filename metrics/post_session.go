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

	BillingFailure Counter
	PortalFailure  Counter
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
	BillingFailure:         &EmptyCounter{},
	PortalFailure:          &EmptyCounter{},
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

	return m, nil
}
