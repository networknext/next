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

func (writer *LocalPingStatsWriter) Write(ctx context.Context, entries []*PingStatsEntry) error {
	writer.written++

	if writer.Logger == nil {
		return errors.New("no logger for local big query writer, can't display entry")
	}

	level.Info(writer.Logger).Log("msg", "wrote analytics bigquery entries")

	for i := range entries {
		entry := entries[i]
		level.Debug(writer.Logger).Log(fmt.Sprintf("entry %d", i), fmt.Sprintf("%+v", *entry))
	}

	return nil
}

type LocalRelayStatsWriter struct {
	Logger log.Logger

	written uint64
}

func (writer *LocalRelayStatsWriter) Write(ctx context.Context, entries []*RelayStatsEntry) error {
	writer.written++

	if writer.Logger == nil {
		return errors.New("no logger for local big query writer, can't display entry")
	}

	level.Info(writer.Logger).Log("msg", "wrote analytics bigquery entries")

	for i := range entries {
		entry := entries[i]
		level.Debug(writer.Logger).Log(fmt.Sprintf("entry %d", i), fmt.Sprintf("%+v", *entry))
	}

	return nil
}
