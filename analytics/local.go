package analytics

import (
	"context"
	"errors"
	"fmt"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

type LocalPubSubPublisher struct {
	Logger log.Logger

	written uint64
}

func (publisher *LocalPubSubPublisher) Publish(ctx context.Context, entry *StatsEntry) error {
	publisher.written++

	if publisher.Logger == nil {
		return errors.New("no logger for local pubsub publisher, can't display entry")
	}

	level.Info(publisher.Logger).Log("msg", "wrote analytics pubsub entry")

	output := fmt.Sprintf("%#v", entry)
	level.Debug(publisher.Logger).Log("entry", output)

	return nil
}

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

func (local *LocalBigQueryWriter) NumSubmitted() uint64 {
	return local.submitted
}

func (local *LocalBigQueryWriter) NumQueued() uint64 {
	return 0
}

func (local *LocalBigQueryWriter) NumFlushed() uint64 {
	return local.submitted
}
