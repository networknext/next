package metrics

import "context"

// GhostArmyStatus defines the metrics reported by the service's status endpoint
type GhostArmyStatus struct {
	// Service Information
	ServiceName string `json:"service_name"`
	GitHash     string `json:"git_hash"`
	Started     string `json:"started"`
	Uptime      string `json:"uptime"`

	// Service Metrics
	Goroutines      int     `json:"goroutines"`
	MemoryAllocated float64 `json:"mb_allocated"`

	// Error Metrics
	EntryPublishingFailure     int `json:"failed_entries_published"`
	SessionsOverEstimate       int `json:"sessions_over_estimate"`
	SessionEntryMarshalFailure int `json:"session_entry_marshal_failure"`

	// Success Metrics
	EntriesPublished int `json:"entries_published"`
}

type GhostArmyMetrics struct {
	ServiceMetrics *ServiceMetrics

	SuccessMetrics *GhostArmySuccessMetrics

	ErrorMetrics *GhostArmyErrorMetrics
}

// NewGhostArmyMetrics creates the metrics that the ghost army service will use.
func NewGhostArmyMetrics(ctx context.Context, handler Handler) (*GhostArmyMetrics, error) {
	serviceName := "ghost_army"

	var err error
	m := &GhostArmyMetrics{}

	m.ServiceMetrics, err = NewServiceMetrics(ctx, handler, serviceName)
	if err != nil {
		return nil, err
	}

	m.SuccessMetrics, err = newGhostArmySuccessMetrics(ctx, handler, serviceName)
	if err != nil {
		return nil, err
	}

	m.ErrorMetrics, err = newGhostArmyErrorMetrics(ctx, handler, serviceName)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func newGhostArmySuccessMetrics(ctx context.Context, handler Handler, serviceName string) (*GhostArmySuccessMetrics, error) {
	var err error
	m := &GhostArmySuccessMetrics{}

	m.SessionEntriesPublished, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Successfully published session entries",
		ServiceName: serviceName,
		ID:          "ghost_army.session_entries_published",
		Unit:        "entries",
		Description: "The number of times the a batch of session entries is published to the portal crunchers",
	})
	if err != nil {
		return nil, err
	}

	return m, nil
}

func newGhostArmyErrorMetrics(ctx context.Context, handler Handler, serviceName string) (*GhostArmyErrorMetrics, error) {
	var err error
	m := &GhostArmyErrorMetrics{}

	m.SessionEntryPublishFailure, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Failed to publish session entries",
		ServiceName: serviceName,
		ID:          "ghost_army.session_entry_publishing_failures",
		Unit:        "errors",
		Description: "The number of times the a batch of session entries fail to publish to portal crunchers",
	})
	if err != nil {
		return nil, err
	}

	m.PublishedSessionsOverEstimate, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Published sessions exceed estimated",
		ServiceName: serviceName,
		ID:          "ghost_army.session_batch_size_exceeded",
		Unit:        "errors",
		Description: "The number of times the a batch of session entries exceeds the estimated size",
	})
	if err != nil {
		return nil, err
	}

	m.SessionEntryMarshalFailure, err = handler.NewCounter(ctx, &Descriptor{
		DisplayName: "Session entry marshal failure",
		ServiceName: serviceName,
		ID:          "ghost_army.marshal_session_failure",
		Unit:        "errors",
		Description: "The number of times a session entry fails to be marshalled",
	})
	if err != nil {
		return nil, err
	}

	return m, nil
}

// EmptyGhostArmyMetrics is used for testing when we want to pass in metrics but don't care about their value.
var EmptyGhostArmyMetrics = GhostArmyMetrics{
	ServiceMetrics: &EmptyServiceMetrics,
	SuccessMetrics: &EmptyGhostArmySuccessMetrics,
	ErrorMetrics:   &EmptyGhostArmyErrorMetrics,
}

// EmptyGhostArmyMetrics is used for collecting successes in the service.
type GhostArmySuccessMetrics struct {
	SessionEntriesPublished Counter
}

var EmptyGhostArmySuccessMetrics = GhostArmySuccessMetrics{
	SessionEntriesPublished: &EmptyCounter{},
}

// EmptyGhostArmyMetrics is used for collecting failures in the service.
type GhostArmyErrorMetrics struct {
	SessionEntryPublishFailure    Counter
	PublishedSessionsOverEstimate Counter
	SessionEntryMarshalFailure    Counter
}

var EmptyGhostArmyErrorMetrics = GhostArmyErrorMetrics{
	SessionEntryPublishFailure:    &EmptyCounter{},
	PublishedSessionsOverEstimate: &EmptyCounter{},
	SessionEntryMarshalFailure:    &EmptyCounter{},
}
