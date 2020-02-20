package billing

import "context"

// NoOp does not perform any billing actions. Useful for when billing is not configured or for testing.
type NoOp struct{}

// Bill does nothing
func (noop *NoOp) Bill(ctx context.Context, sessionID uint64, entry *Entry) error {
	return nil
}
