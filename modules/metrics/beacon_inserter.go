package metrics

import "context"

// BeaconInserterMetrics defines a set of metrics for the beacon insertion service.
type BeaconInserterMetrics struct {
	ServiceMetrics ServiceMetrics

	ReceiveMetrics ReceiverMetrics
	PublishMetrics PublisherMetrics
}

// EmptyBeaconInserterMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyBeaconInserterMetrics BeaconInserterMetrics = BeaconInserterMetrics{
	ServiceMetrics: EmptyServiceMetrics,

	ReceiveMetrics: EmptyReceiverMetrics,
	PublishMetrics: EmptyPublisherMetrics,
}

// NewBeaconInserterMetrics creates the metrics that the beacon insertion service will use.
func NewBeaconInserterMetrics(ctx context.Context, handler Handler) (BeaconInserterMetrics, error) {
	serviceName := "beacon_inserter"

	var err error
	m := BeaconInserterMetrics{}

	m.ServiceMetrics, err = NewServiceMetrics(ctx, handler, serviceName)
	if err != nil {
		return EmptyBeaconInserterMetrics, err
	}

	m.ReceiveMetrics, err = NewReceiverMetrics(ctx, handler, serviceName, "beacon_inserter", "Beacon Inserter", "beacon")
	if err != nil {
		return EmptyBeaconInserterMetrics, err
	}

	m.PublishMetrics, err = NewPublisherMetrics(ctx, handler, serviceName, "beacon_inserter", "Beacon Inserter", "beacon")
	if err != nil {
		return EmptyBeaconInserterMetrics, err
	}

	return m, nil
}
