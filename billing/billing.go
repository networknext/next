package billing

import (
	"context"
	"time"
)

// Biller is a billing interface that handles sending billing entries to remote services
type Biller interface {
	Bill(ctx context.Context, sessionID uint64, entry *Entry) error
	NumSubmitted() uint64
	NumQueued() uint64
	NumFlushed() uint64
}

// Descriptor contains metadata on how to send billing data to the service.
type Descriptor struct {
	// The number of billing clients to run concurrently.
	ClientCount int

	// Publish a non-empty batch after this delay has passed.
	// Note that if an implementation has a maximum threshold, this value will be modified
	// to adhere to that maximum.
	DelayThreshold time.Duration

	// Publish a batch when it has this many messages.
	// Note that if an implementation has a maximum threshold, this value will be modified
	// to adhere to that maximum.
	CountThreshold int

	// Publish a batch when its size in bytes reaches this value.
	// Note that if an implementation has a maximum threshold, this value will be modified
	// to adhere to that maximum.
	ByteThreshold int

	// The number of goroutines that send billing info concurrently per client.
	NumGoroutines int

	// The maximum time that a client will attempt to publish a bundle of messages.
	Timeout time.Duration

	// The size of a client's result channel buffer for receiving messages back from the billing service.
	ResultChannelBuffer int
}
