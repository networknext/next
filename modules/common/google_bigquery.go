package common

import (
	"context"
	"time"

	"github.com/networknext/accelerate/modules/core"
	"github.com/networknext/accelerate/modules/envvar"

	"cloud.google.com/go/bigquery"
	"google.golang.org/api/option"
)

type GoogleBigQueryConfig struct {
	ProjectId          string
	Dataset            string
	TableName          string
	ClientOptions      []option.ClientOption
	BatchSize          int
	BatchDuration      time.Duration
	PublishChannelSize int
}

type GoogleBigQueryPublisher struct {
	PublishChannel      chan bigquery.ValueSaver
	config              GoogleBigQueryConfig
	bigqueryClient      *bigquery.Client
	tableInserter       *bigquery.Inserter
	messageBatch        []bigquery.ValueSaver
	batchStartTime      time.Time
	NumEntriesRecieved  uint64
	NumEntriesPublished uint64
	NumBatchesPublished uint64
}

func CreateGoogleBigQueryPublisher(ctx context.Context, config GoogleBigQueryConfig) (*GoogleBigQueryPublisher, error) {

	if config.ProjectId == "local" {
		config.ClientOptions = []option.ClientOption{
			option.WithEndpoint(envvar.GetString("BIGQUERY_EMULATOR_HOST", "http://127.0.0.1:9050")),
			option.WithoutAuthentication(),
		}
	}

	bigqueryClient, err := bigquery.NewClient(ctx, config.ProjectId, config.ClientOptions...)
	if err != nil {
		core.Error("failed to create google bigquery client: %v", err)
		return nil, err
	}

	publisher := &GoogleBigQueryPublisher{}

	publisher.config = config

	if publisher.config.PublishChannelSize == 0 {
		publisher.config.PublishChannelSize = 10 * 1024
	}

	if publisher.config.BatchDuration == 0 {
		publisher.config.BatchDuration = time.Second
	}

	tableInserter := bigqueryClient.Dataset(config.Dataset).Table(config.TableName).Inserter()

	publisher.bigqueryClient = bigqueryClient
	publisher.tableInserter = tableInserter
	publisher.PublishChannel = make(chan bigquery.ValueSaver, config.PublishChannelSize)

	go publisher.updatePublishChannel(ctx)

	return publisher, nil
}

func (publisher *GoogleBigQueryPublisher) updatePublishChannel(ctx context.Context) {

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

func (publisher *GoogleBigQueryPublisher) publishBatch(ctx context.Context) {

	err := publisher.tableInserter.Put(ctx, publisher.messageBatch)
	if err != nil {
		core.Error("failed to publish bigquery entry: %v", err)
		return
	}

	batchNumMessages := len(publisher.messageBatch)

	publisher.NumBatchesPublished++
	publisher.NumEntriesPublished += uint64(batchNumMessages)

	publisher.messageBatch = []bigquery.ValueSaver{}
}

func (publisher *GoogleBigQueryPublisher) Close() error {
	return publisher.bigqueryClient.Close()
}

// Test entry for making func testing easier
type TestEntry struct {
	Timestamp uint32
}

func (entry *TestEntry) Save() (map[string]bigquery.Value, string, error) {

	e := make(map[string]bigquery.Value)

	e["timestamp"] = int(entry.Timestamp)
	return e, "", nil
}
