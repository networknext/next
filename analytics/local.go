package analytics

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

type LocalPingStatsWriter struct {
	Logger log.Logger

	written uint64
}

func (writer *LocalPingStatsWriter) Write(ctx context.Context, entry *PingStatsEntry) error {
	writer.written++

	if writer.Logger == nil {
		return errors.New("no logger for local big query writer, can't display entry")
	}

	level.Info(writer.Logger).Log("msg", "wrote analytics bigquery entry")

	output := fmt.Sprintf("%#v", entry)
	level.Debug(writer.Logger).Log("entry", output)

	return nil
}

type LocalRelayStatsWriter struct {
	Logger log.Logger

	written uint64
}

func (writer *LocalRelayStatsWriter) Write(ctx context.Context, entry *RelayStatsEntry) error {
	writer.written++

	if writer.Logger == nil {
		return errors.New("no logger for local big query writer, can't display entry")
	}

	level.Info(writer.Logger).Log("msg", "wrote analytics bigquery entry")

	output := fmt.Sprintf("%#v", entry)
	level.Debug(writer.Logger).Log("entry", output)

	return nil
}
