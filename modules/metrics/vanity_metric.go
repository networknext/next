package metrics

import (
	"context"
	"fmt"
)

// VanityStatus defines the metrics reported by the service's status endpoint
type VanityStatus struct {
	// Service Information
	ServiceName string `json:"service_name"`
	GitHash     string `json:"git_hash"`
	Started     string `json:"started"`
	Uptime      string `json:"uptime"`

	// Metrics
	Goroutines        int     `json:"goroutines"`
	MemoryAllocated   float64 `json:"mb_allocated"`
	MessagesReceived  int     `json:"messages_received"`
	SuccessfulUpdates int     `json:"successful_updates"`
	FailedUpdates     int     `json:"failed_updates"`
}

// VanityMetric defines the set of metrics for each vanity metric to be recorded.
type VanityMetric struct {
	SlicesAccelerated       Counter
	SlicesLatencyReduced    Counter
	SlicesPacketLossReduced Counter
	SlicesJitterReduced     Counter
	SessionsAccelerated     Counter
}

// EmptyVanityMetric is used for testing when we want to pass in metrics but don't care about their value,
var EmptyVanityMetric = VanityMetric{
	SlicesAccelerated:       &EmptyCounter{},
	SlicesLatencyReduced:    &EmptyCounter{},
	SlicesPacketLossReduced: &EmptyCounter{},
	SlicesJitterReduced:     &EmptyCounter{},
	SessionsAccelerated:     &EmptyCounter{},
}

// NewVanityMetric creates the metrics the vanity metrics service will use.
// Uses the buyerID as the service name
func NewVanityMetric(ctx context.Context, handler Handler, buyerID string) (*VanityMetric, error) {
	var err error
	m := &VanityMetric{}

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
