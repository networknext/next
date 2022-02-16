package match_data

import (
	"context"
	"fmt"
)

type ErrEntriesBufferFull struct{}

func (e *ErrEntriesBufferFull) Error() string {
	return fmt.Sprintf("match data entries buffer full")
}

// Matcher is an interface that handles sending match data entries to remote services
type Matcher interface {
	Match(ctx context.Context, matchData *MatchDataEntry) error
	FlushBuffer(ctx context.Context)
	Close()
}
