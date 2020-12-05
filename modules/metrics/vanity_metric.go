package metrics

import "context"

// VanityMetric defines the set of metrics for each vanity metric to be recorded.
type VanityMetric struct {
	NumSlicesGlobal			Counter
	NumSlicesPerCustomer	Counter
	NumSessionsGlobal		Counter
	NumSessionsPerCustomer	Counter
	NumPlayHoursPerCustomer	Counter
	RTTReduction			Gauge
	PacketLossReduction		Counter
}

// EmptyVanityMetric is used for testing when we want to pass in metrics but don't care about their value,
var EmptyVanityMetric = VanityMetric{
	NumSlicesGlobal:			&EmptyCounter{},
	NumSlicesPerCustomer:		&EmptyCounter{},
	NumSessionsGlobal:			&EmptyCounter{},
	NumSessionsPerCustomer:		&EmptyCounter{},
	NumPlayHoursPerCustomer:	&EmptyCounter{},
	RTTReduction:				&EmptyGauge{},
	PacketLossReduction:		&EmptyCounter{},
}

// NewVanityMetric creates the metrics the vanity metrics service will use.
// User the buyerID as the service name 
func NewVanityMetric(ctx context.Context, handler TSHandler, buyerID string) (*VanityMetric, error) {
	var err error
	m := &VanityMetric{}

	m.NumSlicesGlobal, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Vanity Metric Num Slices Global",
		ServiceName: buyerID,
		ID:          "vanity_metric.num_slices_global",
		Unit:        "slices",
		Description: "The total number of slices through Network Next",
	})
	if err != nil {
		return nil, err
	}

	m.NumSlicesPerCustomer, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Vanity Metric Num Slices Per Customer",
		ServiceName: buyerID,
		ID:          "vanity_metric.num_slices_per_customer",
		Unit:        "slices",
		Description: fmt.Sprintf("The total number of slices for customer %s" buyerID),
	})
	if err != nil {
		return nil, err
	}

	m.NumSessionsGlobal, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Vanity Metric Num Sessions Global",
		ServiceName: buyerID,
		ID:          "vanity_metric.num_sessions_global",
		Unit:        "sessions",
		Description: "The total number of sessions through Network Next",
	})
	if err != nil {
		return nil, err
	}

	m.NumSessionsPerCustomer, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Vanity Metric Num Sessions Per Customer",
		ServiceName: buyerID,
		ID:          "vanity_metric.num_sessions_per_customer",
		Unit:        "sessions",
		Description: fmt.Sprintf("The total number of sessions for customer %s" buyerID),
	})
	if err != nil {
		return nil, err
	}

	m.NumPlayHoursPerCustomer, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Vanity Metric Num Play Hours Per Customer",
		ServiceName: buyerID,
		ID:          "vanity_metric.num_play_hours_per_customer",
		Unit:        "hours",
		Description: fmt.Sprintf("The total number of play hours for customer %s" buyerID),
	})
	if err != nil {
		return nil, err
	}

	m.RTTReduction, err = handler.NewGauge(ctx, &Descriptor{
		DisplayName: "Vanity Metric RTT Reduction",
		ServiceName: buyerID,
		ID:          "vanity_metric.rtt_reduction",
		Unit:        "ms",
		Description: "The ms of RTT that have been reduced through Network Next.",
	})
	if err != nil {
		return nil, err
	}

	m.PacketLossReduction, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Vanity Metric Packet Loss Reduction",
		ServiceName: buyerID,
		ID:          "vanity_metric.packet_loss_reduction",
		Unit:        "packets",
		Description: "The number of packets that have not been lost due to Network Next.",
	})
	if err != nil {
		return nil, err
	}

	return m, nil
}
