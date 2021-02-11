package metrics

import "context"

// BillingMetrics defines the set of metrics for the billing service.
type BillingMetrics struct {
	ServiceMetrics ServiceMetrics

	BillingSubscriberMetrics SubscriberMetrics
	BillingPublisherMetrics  PublisherMetrics
}

// EmptyBillingMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyBillingMetrics BillingMetrics = BillingMetrics{
	ServiceMetrics:           EmptyServiceMetrics,
	BillingSubscriberMetrics: EmptySubscriberMetrics,
	BillingPublisherMetrics:  EmptyPublisherMetrics,
}

// NewBillingMetrics creates the metrics that the billing service will use.
func NewBillingMetrics(ctx context.Context, handler Handler) (BillingMetrics, error) {
	serviceName := "billing"

	var err error
	m := BillingMetrics{}

	m.ServiceMetrics, err = NewServiceMetrics(ctx, handler, serviceName)
	if err != nil {
		return EmptyBillingMetrics, err
	}

	m.BillingPublisherMetrics, err = NewPublisherMetrics(ctx, handler, serviceName, "billing", "Billing", "billing")
	if err != nil {
		return EmptyBillingMetrics, err
	}

	m.BillingSubscriberMetrics, err = NewSubscriberMetrics(ctx, handler, serviceName, "billing", "Billing", "billing")
	if err != nil {
		return EmptyBillingMetrics, err
	}

	return m, nil
}
