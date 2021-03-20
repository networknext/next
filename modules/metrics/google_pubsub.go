package metrics

import "context"

// GooglePublisherMetrics defines the metrics for publishing to Google Pub/Sub.
type GooglePublisherMetrics struct {
	EntriesQueuedToPublish Counter
	EntriesPublished       Counter
	ErrorMetrics           GooglePublisherErrorMetrics
}

// EmptyGooglePublisherMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyGooglePublisherMetrics GooglePublisherMetrics = GooglePublisherMetrics{
	EntriesQueuedToPublish: &EmptyCounter{},
	EntriesPublished:       &EmptyCounter{},
	ErrorMetrics:           EmptyGooglePublisherErrorMetrics,
}

// GooglePublisherErrorMetrics contains the error metrics for the Google Pub/Sub publisher.
type GooglePublisherErrorMetrics struct {
	PublishFailure Counter
}

// EmptyGooglePublisherErrorMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyGooglePublisherErrorMetrics = GooglePublisherErrorMetrics{
	PublishFailure: &EmptyCounter{},
}

// GoogleSubscriberMetrics defines the metrics for subscribing to Google Pub/Sub.
type GoogleSubscriberMetrics struct {
	EntriesReceived      Counter
	EntriesQueuedToWrite Gauge
	EntriesSubmitted     Counter
	ErrorMetrics         GoogleSubscriberErrorMetrics
}

// EmptyGoogleSubscriberMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyGoogleSubscriberMetrics GoogleSubscriberMetrics = GoogleSubscriberMetrics{
	EntriesReceived:      &EmptyCounter{},
	EntriesQueuedToWrite: &EmptyGauge{},
	EntriesSubmitted:     &EmptyCounter{},
	ErrorMetrics:         EmptyGoogleSubscriberErrorMetrics,
}

// GoogleSubscriberErrorMetrics contains the error metrics for the Google Pub/Sub subscriber.
type GoogleSubscriberErrorMetrics struct {
	BatchedReadFailure   Counter
	ReadFailure          Counter
	QueueFailure         Counter
	BigQueryWriteFailure Counter
}

// EmptyGoogleSubscriberErrorMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyGoogleSubscriberErrorMetrics GoogleSubscriberErrorMetrics = GoogleSubscriberErrorMetrics{
	BatchedReadFailure:   &EmptyCounter{},
	ReadFailure:          &EmptyCounter{},
	QueueFailure:         &EmptyCounter{},
	BigQueryWriteFailure: &EmptyCounter{},
}

func NewGooglePublisherMetrics(ctx context.Context, handler Handler, serviceName string, handlerName string, handlerID string) (*GooglePublisherMetrics, error) {
	publisherMetrics := GooglePublisherMetrics{}
	var err error

	publisherMetrics.EntriesQueuedToPublish, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Entries Queued To Publish",
		ServiceName: serviceName,
		ID:          handlerID + ".entries_submitted",
		Unit:        "entries",
		Description: "The number of " + handlerID + " entries that have been queued for publishing through Google Pub/Sub.",
	})
	if err != nil {
		return nil, err
	}

	publisherMetrics.EntriesPublished, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Entries Published",
		ServiceName: serviceName,
		ID:          handlerID + ".entries.written",
		Unit:        "entries",
		Description: "The number of " + handlerID + " entries that have been published through Google Pub/Sub.",
	})
	if err != nil {
		return nil, err
	}

	publisherMetrics.ErrorMetrics.PublishFailure, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Entries Publish Failure",
		ServiceName: serviceName,
		ID:          handlerID + ".error.publish_failure",
		Unit:        "errors",
		Description: "The number of " + handlerID + " entries that failed to be published through Google Pub/Sub.",
	})
	if err != nil {
		return nil, err
	}

	return &publisherMetrics, nil
}

func NewGoogleSubscriberMetircs(ctx context.Context, handler Handler, serviceName string, handlerName string, handlerID string) (*GoogleSubscriberMetrics, error) {
	subscriberMetrics := GoogleSubscriberMetrics{}
	var err error

	subscriberMetrics.EntriesReceived, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Entries Received",
		ServiceName: serviceName,
		ID:          handlerID + ".entries",
		Unit:        "entries",
		Description: "The total number of " + handlerID + " entries received through Google Pub/Sub",
	})
	if err != nil {
		return nil, err
	}

	subscriberMetrics.EntriesQueuedToWrite, err = handler.NewGauge(ctx, &Descriptor{
		DisplayName: handlerName + " Entries Queued To Write",
		ServiceName: serviceName,
		ID:          handlerID + ".entries.queued",
		Unit:        "entries",
		Description: "The total number of " + handlerID + " entries queued to be written to BigQuery",
	})
	if err != nil {
		return nil, err
	}

	subscriberMetrics.EntriesSubmitted, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Entries Submitted",
		ServiceName: serviceName,
		ID:          handlerID + ".entries.written",
		Unit:        "entries",
		Description: "The total number of " + handlerID + " entries successfully written to BigQuery",
	})
	if err != nil {
		return nil, err
	}

	subscriberMetrics.ErrorMetrics.BatchedReadFailure, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Batched Read Failure",
		ServiceName: serviceName,
		ID:          handlerID + ".error.batched_read_failure",
		Unit:        "errors",
		Description: "The total number of " + handlerID + " entries that could not be batch read from Google Pub/Sub",
	})
	if err != nil {
		return nil, err
	}

	subscriberMetrics.ErrorMetrics.ReadFailure, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Read Failure",
		ServiceName: serviceName,
		ID:          handlerID + ".error.read_failure",
		Unit:        "errors",
		Description: "The total number of " + handlerID + " entries that could not be read",
	})
	if err != nil {
		return nil, err
	}

	subscriberMetrics.ErrorMetrics.QueueFailure, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " Queue Failure",
		ServiceName: serviceName,
		ID:          handlerID + ".error.queue_failure",
		Unit:        "errors",
		Description: "The total number of " + handlerID + " entries that could not be queued for submission to BigQuery",
	})
	if err != nil {
		return nil, err
	}

	subscriberMetrics.ErrorMetrics.BigQueryWriteFailure, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: handlerName + " BigQuery Write Failures",
		ServiceName: serviceName,
		ID:          handlerID + ".error.write_failure",
		Unit:        "errors",
		Description: "The total number of " + handlerID + " entries that failed to write to BigQuery",
	})
	if err != nil {
		return nil, err
	}

	return &subscriberMetrics, nil
}
