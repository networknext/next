package analytics

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

type LocalBigQueryWriter struct {
	Logger log.Logger

	submitted uint64
}

func (writer *LocalBigQueryWriter) Write(ctx context.Context, entry *StatsEntry) error {
	writer.submitted++

	if writer.Logger == nil {
		return errors.New("no logger for local big query writer, can't display entry")
	}

	level.Info(writer.Logger).Log("msg", "wrote analytics bigquery entry")

	output := fmt.Sprintf("%#v", entry)
	level.Debug(writer.Logger).Log("entry", output)

	return nil
}
