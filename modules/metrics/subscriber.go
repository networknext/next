package metrics

import "context"

// SubscriberMetrics defines the metrics that track receiving subscribed data.
type SubscriberMetrics struct {
	EntriesReceived Counter

	UnmarshalFailure Counter
}

// EmptySubscriberMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptySubscriberMetrics SubscriberMetrics = SubscriberMetrics{
	EntriesReceived: &EmptyCounter{},

	UnmarshalFailure: &EmptyCounter{},
}

// NewSubscriberMetrics creates the metrics that track receiving subscribed data.
func NewSubscriberMetrics(ctx context.Context, handler Handler, serviceName string, handlerID string, handlerName string, handlerDescription string) (SubscriberMetrics, error) {
	var err error
	m := SubscriberMetrics{}

	m.EntriesReceived, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Entries Received",
		ServiceName: serviceName,
		ID:          handlerID + ".entries_received",
		Unit:        "entries",
		Description: "The number of " + handlerDescription + " entries that have been received.",
	})
	if err != nil {
		return EmptySubscriberMetrics, err
	}

	m.UnmarshalFailure, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Unmarshal Failure",
		ServiceName: serviceName,
		ID:          handlerID + ".unmarshal_failure",
		Unit:        "entries",
		Description: "The number of " + handlerDescription + " entries that failed to be unmarshaled.",
	})
	if err != nil {
		return EmptySubscriberMetrics, err
	}

	return m, nil
}
