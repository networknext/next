package analytics

import (
	"context"
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

type LocalPubSubWriter struct {
	Logger log.Logger

	written uint64
}

func (writer *LocalPubSubWriter) Write(ctx context.Context, entry *StatsEntry) {
	writer.written++

	if writer.Logger != nil {
		level.Info(writer.Logger).Log("msg", "wrote analytics pubsub entry")
	}

	output := fmt.Sprintf("%#v", entry)
	level.Debug(writer.Logger).Log("entry", output)
}

type LocalBigQueryWriter struct {
	Logger log.Logger

	written uint64
}

func (writer *LocalBigQueryWriter) Write(ctx context.Context, entry *StatsEntry) {
	writer.written++

	if writer.Logger != nil {
		level.Info(writer.Logger).Log("msg", "wrote analytics bigquery entry")
	}

	output := fmt.Sprintf("%#v", entry)
	level.Debug(writer.Logger).Log("entry", output)
}
