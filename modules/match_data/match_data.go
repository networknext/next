package match_data

import (
	"context"
	"fmt"
)

type ErrMatchDataBufferFull struct{}

func (e *ErrMatchDataBufferFull) Error() string {
	return fmt.Sprintf("match data buffer full")
}

// Matcher is an interface that handles sending match data entries to remote services
type Matcher interface {
	Match(ctx context.Context, matchData *MatchDataEntry) error
	FlushBuffer(ctx context.Context)
	Close()
}
