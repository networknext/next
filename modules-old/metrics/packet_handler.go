package metrics

import "context"

// PacketHandlerMetrics defines the set of metrics for any incoming packet or endpoint handler.
type PacketHandlerMetrics struct {
	Invocations  Counter
	Duration     Gauge
	LongDuration Counter
}

// EmptyPacketHandlerMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyPacketHandlerMetrics = PacketHandlerMetrics{
	Invocations:  &EmptyCounter{},
	Duration:     &EmptyGauge{},
	LongDuration: &EmptyCounter{},
}

// NewPacketHandlerMetrics creates the metrics a packet handler will use.
func NewPacketHandlerMetrics(ctx context.Context, handler Handler, serviceName string, handlerID string, handlerName string, packetDescription string) (*PacketHandlerMetrics, error) {
	var err error
	m := &PacketHandlerMetrics{}

	m.Invocations, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Invocations",
		ServiceName: serviceName,
		ID:          handlerID + ".invocations",
		Unit:        "invocations",
		Description: "The number of times a " + packetDescription + " is received.",
	})
	if err != nil {
		return nil, err
	}

	m.Duration, err = handler.NewGauge(ctx, &Descriptor{
		DisplayName: handlerName + " Duration",
		ServiceName: serviceName,
		ID:          handlerID + ".duration",
		Unit:        "ms",
		Description: "The amount of time a " + packetDescription + " takes to complete in milliseconds.",
	})
	if err != nil {
		return nil, err
	}

	m.LongDuration, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Long Durations",
		ServiceName: serviceName,
		ID:          handlerID + ".long_durations",
		Unit:        "invocations",
		Description: "The number of times a " + packetDescription + " takes longer than 100 milliseconds to complete.",
	})
	if err != nil {
		return nil, err
	}

	return m, nil
}
