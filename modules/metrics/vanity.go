package metrics

import (
	"context"
	"fmt"
)

// VanityMetrics defines the set of vanity metrics to be displayed for customers.
type VanityMetrics struct {
	SlicesAccelerated       Counter
	SlicesLatencyReduced    Counter
	SlicesPacketLossReduced Counter
	SlicesJitterReduced     Counter
	SessionsAccelerated     Counter
}

// EmptyVanityMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyVanityMetrics = VanityMetrics{
	SlicesAccelerated:       &EmptyCounter{},
	SlicesLatencyReduced:    &EmptyCounter{},
	SlicesPacketLossReduced: &EmptyCounter{},
	SlicesJitterReduced:     &EmptyCounter{},
	SessionsAccelerated:     &EmptyCounter{},
}

// NewVanityMetrics creates the metrics the vanity metrics service will use.
// Uses the buyerID as the service name
func NewVanityMetrics(ctx context.Context, handler Handler, buyerID string) (*VanityMetrics, error) {
	var err error
	m := &VanityMetrics{}

	m.SlicesAccelerated, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Slices Accelerated",
		ServiceName: buyerID,
		ID:          fmt.Sprintf("vanity_metric.%s.slices_accelerated", buyerID),
		Unit:        "slices",
		Description: fmt.Sprintf("The number of slices that have been accelerated for customer %s", buyerID),
	})
	if err != nil {
		return nil, err
	}

	m.SlicesLatencyReduced, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Slices Latency Reduced",
		ServiceName: buyerID,
		ID:          fmt.Sprintf("vanity_metric.%s.slices_latency_reduced", buyerID),
		Unit:        "slices",
		Description: fmt.Sprintf("The number of slices where latency was reduced over Network Next for customer %s", buyerID),
	})
	if err != nil {
		return nil, err
	}

	m.SlicesPacketLossReduced, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Slices Packet Loss Reduced",
		ServiceName: buyerID,
		ID:          fmt.Sprintf("vanity_metric.%s.slices_packet_loss_reduced", buyerID),
		Unit:        "slices",
		Description: fmt.Sprintf("The number of slices where packet loss was reduced over Network Next for customer %s", buyerID),
	})
	if err != nil {
		return nil, err
	}

	m.SlicesJitterReduced, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Slices Jitter Reduced",
		ServiceName: buyerID,
		ID:          fmt.Sprintf("vanity_metric.%s.slices_jitter_reduced", buyerID),
		Unit:        "slices",
		Description: fmt.Sprintf("The number of slices where jitter was reduced over Network Next for customer %s", buyerID),
	})
	if err != nil {
		return nil, err
	}

	m.SessionsAccelerated, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Sessions Accelerated",
		ServiceName: buyerID,
		ID:          fmt.Sprintf("vanity_metric.%s.sessions_accelerated", buyerID),
		Unit:        "sessions",
		Description: fmt.Sprintf("The number of sessions that have been accelerated for customer %s", buyerID),
	})
	if err != nil {
		return nil, err
	}

	return m, nil
}

// VanityServiceMetrics defines the set of metrics for monitoring the vanity service.
type VanityServiceMetrics struct {
	ServiceMetrics ServiceMetrics

	ReceiveMetrics ReceiverMetrics

	VanityUpdateSuccessCount Counter
	VanityUpdateFailureCount Counter
}

// EmptyVanityServiceMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyVanityServiceMetrics = VanityServiceMetrics{
	ServiceMetrics: EmptyServiceMetrics,

	ReceiveMetrics: EmptyReceiverMetrics,

	VanityUpdateSuccessCount: &EmptyCounter{},
	VanityUpdateFailureCount: &EmptyCounter{},
}

// NewVanityServiceMetrics creates the metrics that the vanity service will use.
func NewVanityServiceMetrics(ctx context.Context, handler Handler) (VanityServiceMetrics, error) {
	serviceName := "vanity"

	var err error
	m := VanityServiceMetrics{}

	m.ServiceMetrics, err = NewServiceMetrics(ctx, handler, serviceName)
	if err != nil {
		return EmptyVanityServiceMetrics, err
	}

	m.ReceiveMetrics, err = NewReceiverMetrics(ctx, handler, serviceName, "vanity", "Vanity", "vanity metrics")
	if err != nil {
		return EmptyVanityServiceMetrics, err
	}

	m.VanityUpdateSuccessCount, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Vanity Metrics Update Success Count",
		ServiceName: serviceName,
		ID:          "vanity_metrics_update_success_count",
		Unit:        "updates",
		Description: "The number of successful vanity metrics updates to the metrics handler.",
	})
	if err != nil {
		return EmptyVanityServiceMetrics, err
	}

	m.VanityUpdateFailureCount, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Vanity Metrics Update Failure Count",
		ServiceName: serviceName,
		ID:          "vanity_metrics_update_failure_count",
		Unit:        "updates",
		Description: "The number of failed vanity metrics updates to the metrics handler.",
	})
	if err != nil {
		return EmptyVanityServiceMetrics, err
	}

	return m, nil
}
