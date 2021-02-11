package metrics

import "context"

// RoutineMetrics defines the set of metrics for any routine logic that we want to measure, such as sync routines or endpoint handlers.
type RoutineMetrics struct {
	Invocations  Counter
	Duration     Gauge
	LongDuration Counter
}

// EmptyRoutineMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyRoutineMetrics = RoutineMetrics{
	Invocations:  &EmptyCounter{},
	Duration:     &EmptyGauge{},
	LongDuration: &EmptyCounter{},
}

// NewRoutineMetrics creates the metrics any routine logic will use for measuring performance.
func NewRoutineMetrics(ctx context.Context, handler Handler, serviceName string, handlerID string, handlerName string, handlerDescription string) (RoutineMetrics, error) {
	var err error
	m := RoutineMetrics{}

	m.Invocations, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Invocations",
		ServiceName: serviceName,
		ID:          handlerID + ".invocations",
		Unit:        "invocations",
		Description: "The number of times a " + handlerDescription + " is invoked.",
	})
	if err != nil {
		return EmptyRoutineMetrics, err
	}

	m.Duration, err = handler.NewGauge(ctx, &Descriptor{
		DisplayName: handlerName + " Duration",
		ServiceName: serviceName,
		ID:          handlerID + ".duration",
		Unit:        "ms",
		Description: "The amount of time a " + handlerDescription + " takes to complete in milliseconds.",
	})
	if err != nil {
		return EmptyRoutineMetrics, err
	}

	m.LongDuration, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Long Durations",
		ServiceName: serviceName,
		ID:          handlerID + ".long_durations",
		Unit:        "invocations",
		Description: "The number of times a " + handlerDescription + " takes longer than 100 milliseconds to complete.",
	})
	if err != nil {
		return EmptyRoutineMetrics, err
	}

	return m, nil
}
