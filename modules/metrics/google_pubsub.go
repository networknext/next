package metrics

import "context"

type GooglePublisherMetrics struct {
    EntriesQueuedToPublish Counter
    EntriesPublished Counter
    ErrorMetrics GooglePublisherErrorMetrics
}

// EmptyBeaconMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyGooglePublisherMetrics GooglePublisherMetrics = GooglePublisherMetrics{
    EntriesQueuedToPublish: &EmptyCounter{},
    EntriesPublished:          &EmptyCounter{},
    ErrorMetrics:              EmptyGooglePublisherErrorMetrics,
}

type GooglePublisherErrorMetrics struct {
    PublishFailure Counter
}

var EmptyGooglePublisherErrorMetrics = GooglePublisherErrorMetrics{
    PublishFailure: &EmptyCounter{},
}

type GoogleSubscriberMetrics struct {
    EntriesReceived Counter
    EntriesQueuedToWrite Counter
    EntriesSubmitted Counter
    ErrorMetrics GoogleSubscriberErrorMetrics
}

var EmptyGoogleSubscriberMetrics GoogleSubscriberMetrics = GoogleSubscriberMetrics{
    EntriesReceived: &EmptyCounter{},
    EntriesQueuedToWrite: &EmptyCounter{},
    EntriesSubmitted: &EmptyCounter{},
    ErrorMetrics: EmptyGoogleSubscriberErrorMetrics,
}

type GoogleSubscriberErrorMetrics struct {
    BatchedReadFailure Counter
    ReadFailure Counter
    EntriesWithNaN Counter
    InvalidEntries Counter
    BigQueryWriteFailure Counter
}

var EmptyGoogleSubscriberErrorMetrics GoogleSubscriberErrorMetrics = GoogleSubscriberErrorMetrics{
    BatchedReadFailure: &EmptyCounter{},
    ReadFailure: &EmptyCounter{},
    EntriesWithNaN: &EmptyCounter{},
    InvalidEntries: &EmptyCounter{},
    BigQueryWriteFailure: &EmptyCounter{},
}

func NewGooglePublisherMetrics(ctx context.Context, metricsHandler Handler, serviceName string, handlerName string, handlerID string) (*GooglePublisherMetrics, error) {
    publisherMetrics := GooglePublisherMetrics{}

    publisherMetrics.EntriesQueuedToPublish, err = handler.NewCounter(ctx, &Descriptor{
        DisplayName: handlerName + " Entries Queued To Publish",
        ServiceName: serviceName,
        ID:          handlerID + ".entries_submitted",
        Unit:        "errors",
        Description: "The number of " + handlerID + " entries that have been queued for publishing through Google PubSub.",
    })
    if err != nil {
        return nil, err
    }

    publisherMetrics.EntriesPublished, err = handler.NewCounter(ctx, &Descriptor{
        DisplayName: handlerName + " Entries Published",
        ServiceName: serviceName,
        ID:          handlerID + ".entries.written",
        Unit:        "errors",
        Description: "The number of " + handlerID + " entries that have been published through Google PubSub.",
    })
    if err != nil {
        return nil, err
    }

    publisherMetrics.ErrorMetrics.
}
