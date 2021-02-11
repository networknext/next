package metrics

import "context"

// PortalCruncherMetrics defines the set of metrics for the portal cruncher.
type PortalCruncherMetrics struct {
	ServiceMetrics ServiceMetrics

	ReceiveMetrics ReceiverMetrics

	BigTableMetrics BigTableWriteMetrics
}

// EmptyPortalCruncherMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyPortalCruncherMetrics = PortalCruncherMetrics{
	ServiceMetrics: EmptyServiceMetrics,

	ReceiveMetrics: EmptyReceiverMetrics,

	BigTableMetrics: EmptyBigTableWriteMetrics,
}

// NewPortalCruncherMetrics creates the metrics that the portal cruncher will use.
func NewPortalCruncherMetrics(ctx context.Context, handler Handler) (PortalCruncherMetrics, error) {
	serviceName := "portal_cruncher"

	var err error
	m := PortalCruncherMetrics{}

	m.ServiceMetrics, err = NewServiceMetrics(ctx, handler, serviceName)
	if err != nil {
		return EmptyPortalCruncherMetrics, err
	}

	m.ReceiveMetrics, err = NewReceiverMetrics(ctx, handler, serviceName, "portal_cruncher", "Portal Cruncher", "portal data")
	if err != nil {
		return EmptyPortalCruncherMetrics, err
	}

	m.BigTableMetrics, err = NewBigTableWriteMetrics(ctx, handler, serviceName)
	if err != nil {
		return EmptyPortalCruncherMetrics, err
	}

	return m, nil
}
