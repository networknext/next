package metrics

import "context"

// BigTableWriteMetrics defines the set of metrics for writing portal data to BigTable.
type BigTableWriteMetrics struct {
	WriteMetaSuccessCount  Counter
	WriteSliceSuccessCount Counter
	WriteMetaFailureCount  Counter
	WriteSliceFailureCount Counter
}

// EmptyBigTableWriteMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyBigTableWriteMetrics = BigTableWriteMetrics{
	WriteMetaSuccessCount:  &EmptyCounter{},
	WriteSliceSuccessCount: &EmptyCounter{},
	WriteMetaFailureCount:  &EmptyCounter{},
	WriteSliceFailureCount: &EmptyCounter{},
}

// BigTableReadMetrics defines the set of metrics for reading portal data from BigTable.
type BigTableReadMetrics struct {
	ReadMetaSuccessCount  Counter
	ReadSliceSuccessCount Counter
	ReadMetaFailureCount  Counter
	ReadSliceFailureCount Counter
}

// EmptyBigTableReadMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyBigTableReadMetrics = BigTableReadMetrics{
	ReadMetaSuccessCount:  &EmptyCounter{},
	ReadSliceSuccessCount: &EmptyCounter{},
	ReadMetaFailureCount:  &EmptyCounter{},
	ReadSliceFailureCount: &EmptyCounter{},
}

// NewBigTableWriteMetrics creates the metrics for writing portal data to BigTable.
func NewBigTableWriteMetrics(ctx context.Context, handler Handler, serviceName string) (BigTableWriteMetrics, error) {
	var err error
	m := BigTableWriteMetrics{}

	m.WriteMetaSuccessCount, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "BigTable Write Meta Success Count",
		ServiceName: serviceName,
		ID:          "bigtable.write_meta_success_count",
		Unit:        "writes",
		Description: "The number of successful meta writes to BigTable.",
	})
	if err != nil {
		return EmptyBigTableWriteMetrics, err
	}

	m.WriteSliceSuccessCount, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "BigTable Write Slice Success Count",
		ServiceName: serviceName,
		ID:          "bigtable.write_slice_success_count",
		Unit:        "writes",
		Description: "The number of successful slice writes to BigTable.",
	})
	if err != nil {
		return EmptyBigTableWriteMetrics, err
	}

	m.WriteMetaFailureCount, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "BigTable Write Meta Failure Count",
		ServiceName: serviceName,
		ID:          "bigtable.write_meta_failure_count",
		Unit:        "writes",
		Description: "The number of failed meta writes to BigTable.",
	})
	if err != nil {
		return EmptyBigTableWriteMetrics, err
	}

	m.WriteSliceFailureCount, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "BigTable Write Slice Failure Count",
		ServiceName: serviceName,
		ID:          "bigtable.write_slice_failure_count",
		Unit:        "writes",
		Description: "The number of failed slice writes to BigTable.",
	})
	if err != nil {
		return EmptyBigTableWriteMetrics, err
	}

	return m, nil
}

// NewBigTableReadMetrics creates the metrics for reading portal data from BigTable.
func NewBigTableReadMetrics(ctx context.Context, handler Handler, serviceName string) (BigTableReadMetrics, error) {
	var err error
	m := BigTableReadMetrics{}

	m.ReadMetaSuccessCount, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "BigTable Read Meta Success Count",
		ServiceName: serviceName,
		ID:          "bigtable.read_meta_success_count",
		Unit:        "reads",
		Description: "The number of successful meta reads from BigTable.",
	})
	if err != nil {
		return EmptyBigTableReadMetrics, err
	}

	m.ReadSliceSuccessCount, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "BigTable Read Slice Success Count",
		ServiceName: serviceName,
		ID:          "bigtable.read_slice_success_count",
		Unit:        "reads",
		Description: "The number of successful slice reads from BigTable.",
	})
	if err != nil {
		return EmptyBigTableReadMetrics, err
	}

	m.ReadMetaFailureCount, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "BigTable Read Meta Failure Count",
		ServiceName: serviceName,
		ID:          "bigtable.read_meta_failure_count",
		Unit:        "reads",
		Description: "The number of failed meta reads from BigTable.",
	})
	if err != nil {
		return EmptyBigTableReadMetrics, err
	}

	m.ReadSliceFailureCount, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "BigTable Read Slice Failure Count",
		ServiceName: serviceName,
		ID:          "bigtable.read_slice_failure_count",
		Unit:        "reads",
		Description: "The number of failed slice reads from BigTable.",
	})
	if err != nil {
		return EmptyBigTableReadMetrics, err
	}

	return m, nil
}
