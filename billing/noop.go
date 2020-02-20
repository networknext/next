package billing

import "context"

// NoOpBiller does not perform any billing actions. Useful for when billing is not configured or for testing.
type NoOpBiller struct{}

// Bill does nothing
func (noop *NoOpBiller) Bill(ctx context.Context, sessionID uint64, entry *Entry) error {
	return nil
}
