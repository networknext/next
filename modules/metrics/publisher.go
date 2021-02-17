package metrics

import "context"

// PublisherMetrics defines the metrics that track the buffered publishing of data.
type PublisherMetrics struct {
	EntriesSubmitted Counter
	EntriesQueued    Gauge
	EntriesFlushed   Counter

	MarshalFailure Counter
	PublishFailure Counter
}

// EmptyPublisherMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyPublisherMetrics PublisherMetrics = PublisherMetrics{
	EntriesSubmitted: &EmptyCounter{},
	EntriesQueued:    &EmptyGauge{},
	EntriesFlushed:   &EmptyCounter{},

	MarshalFailure: &EmptyCounter{},
	PublishFailure: &EmptyCounter{},
}

// NewPublisherMetrics creates the metrics that track the buffered publishing of data.
func NewPublisherMetrics(ctx context.Context, handler Handler, serviceName string, handlerID string, handlerName string, handlerDescription string) (PublisherMetrics, error) {
	var err error
	m := PublisherMetrics{}

	m.EntriesSubmitted, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Entries Submitted",
		ServiceName: serviceName,
		ID:          handlerID + ".entries_submitted",
		Unit:        "entries",
		Description: "The number of " + handlerDescription + " entries that have been submitted to be published.",
	})
	if err != nil {
		return EmptyPublisherMetrics, err
	}

	m.EntriesQueued, err = handler.NewGauge(ctx, &Descriptor{
		DisplayName: handlerName + " Entries Queued",
		ServiceName: serviceName,
		ID:          handlerID + ".entries_queued",
		Unit:        "entries",
		Description: "The number of " + handlerDescription + " entries that have been queued for publishing.",
	})
	if err != nil {
		return EmptyPublisherMetrics, err
	}

	m.EntriesFlushed, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Entries Flushed",
		ServiceName: serviceName,
		ID:          handlerID + ".entries_flushed",
		Unit:        "entries",
		Description: "The number of " + handlerDescription + " entries that have been flushed from the internal publish buffer.",
	})
	if err != nil {
		return EmptyPublisherMetrics, err
	}

	m.MarshalFailure, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Marshal Failure",
		ServiceName: serviceName,
		ID:          handlerID + ".marshal_failure",
		Unit:        "entries",
		Description: "The number of " + handlerDescription + " entries that failed to be marshaled.",
	})
	if err != nil {
		return EmptyPublisherMetrics, err
	}

	m.PublishFailure, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Publish Failure",
		ServiceName: serviceName,
		ID:          handlerID + ".publish_failure",
		Unit:        "entries",
		Description: "The number of times some " + handlerDescription + " entries have failed to publish.",
	})
	if err != nil {
		return EmptyPublisherMetrics, err
	}

	return m, nil
}
