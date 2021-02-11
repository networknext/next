package metrics

import "context"

// ReceiverMetrics defines the metrics that track receiving data.
type ReceiverMetrics struct {
	EntriesReceived Counter

	UnmarshalFailure Counter
}

// EmptyReceiverMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyReceiverMetrics ReceiverMetrics = ReceiverMetrics{
	EntriesReceived: &EmptyCounter{},

	UnmarshalFailure: &EmptyCounter{},
}

// NewReceiverMetrics creates the metrics that track receiving data.
func NewReceiverMetrics(ctx context.Context, handler Handler, serviceName string, handlerID string, handlerName string, handlerDescription string) (ReceiverMetrics, error) {
	var err error
	m := ReceiverMetrics{}

	m.EntriesReceived, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Entries Received",
		ServiceName: serviceName,
		ID:          handlerID + ".entries_received",
		Unit:        "entries",
		Description: "The number of " + handlerDescription + " entries that have been received.",
	})
	if err != nil {
		return EmptyReceiverMetrics, err
	}

	m.UnmarshalFailure, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Unmarshal Failure",
		ServiceName: serviceName,
		ID:          handlerID + ".unmarshal_failure",
		Unit:        "entries",
		Description: "The number of " + handlerDescription + " entries that failed to be unmarshaled.",
	})
	if err != nil {
		return EmptyReceiverMetrics, err
	}

	return m, nil
}
