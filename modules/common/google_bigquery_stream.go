package common

import (
	"context"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/networknext/backend/modules/core"
	"google.golang.org/api/option"
)

type GoogleBigQueryStreamConfig struct {
	ProjectId          string
	Dataset            string
	TableName          string
	BatchSize          int
	BatchDuration      time.Duration
	PublishChannelSize int
	ClientOptions      []option.ClientOption
}

type GoogleBigQueryStreamPublisher struct {
	PublishChannel      chan bigquery.ValueSaver
	config              GoogleBigQueryConfig
	bigqueryClient      *bigquery.Client
	TableInserter       *bigquery.Inserter
	messageBatch        []bigquery.ValueSaver
	batchStartTime      time.Time
	NumEntriesRecieved  uint64
	NumEntriesPublished uint64
	NumBatchesPublished uint64
}

func CreateGoogleBigQueryStreamPublisher(ctx context.Context, config GoogleBigQueryConfig) (*GoogleBigQueryStreamPublisher, error) {

	bigqueryClient, err := bigquery.NewClient(ctx, config.ProjectId, config.ClientOptions...)
	if err != nil {
		core.Error("failed to create google bigquery client: %v", err)
		return nil, err
	}

	publisher := &GoogleBigQueryStreamPublisher{}

	publisher.config = config

	if publisher.config.PublishChannelSize == 0 {
		publisher.config.PublishChannelSize = 10 * 1024
	}

	if publisher.config.BatchDuration == 0 {
		publisher.config.BatchDuration = time.Second
	}

	tableInserter := bigqueryClient.Dataset(config.Dataset).Table(config.TableName).Inserter()

	publisher.bigqueryClient = bigqueryClient
	publisher.TableInserter = tableInserter
	publisher.PublishChannel = make(chan bigquery.ValueSaver, config.PublishChannelSize)

	go publisher.updatePublishChannel(ctx)

	return publisher, nil
}

func (publisher *GoogleBigQueryStreamPublisher) updatePublishChannel(ctx context.Context) {

	ticker := time.NewTicker(publisher.config.BatchDuration)

	for {

		select {

		case <-ctx.Done():
			return

		case <-ticker.C:
			// Took too long to fill the batch
			if len(publisher.messageBatch) > 0 {
				publisher.publishBatch(ctx)
			}
			break

		case message := <-publisher.PublishChannel:

			publisher.messageBatch = append(publisher.messageBatch, message)
			publisher.NumEntriesRecieved++

			if len(publisher.messageBatch) >= publisher.config.BatchSize {
				publisher.publishBatch(ctx)
			}
		}
	}
}

func (publisher *GoogleBigQueryStreamPublisher) publishBatch(ctx context.Context) {

	err := publisher.TableInserter.Put(ctx, publisher.messageBatch)
	if err != nil {
		core.Error("failed to publish bigquery entry: %v", err)
		return
	}

	batchId := publisher.NumBatchesPublished
	batchNumMessages := len(publisher.messageBatch)

	publisher.NumBatchesPublished++
	publisher.NumEntriesPublished += uint64(batchNumMessages)

	publisher.messageBatch = []bigquery.ValueSaver{}

	core.Debug("published batch %d containing %d messages", batchId, batchNumMessages)
}
