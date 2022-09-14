package common

import (
	"context"
	"sync"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/networknext/backend/modules/core"
)

type GoogleBigQueryConfig struct {
	ProjectId          string
	Dataset            string
	TableName          string
	BatchSize          int
	BatchDuration      time.Duration
	PublishChannelSize int
}

type GoogleBigQueryPublisher struct {
	PublishChannel  chan *bigquery.ValueSaver
	config          GoogleBigQueryConfig
	bigqueryClient  *bigquery.Client
	TableInserter   *bigquery.Inserter
	messageBatch    []*bigquery.ValueSaver
	batchStartTime  time.Time
	mutex           sync.RWMutex
	numMessagesSent int
	numBatchesSent  int
}

func CreateGoogleBigQueryPublisher(ctx context.Context, config GoogleBigQueryConfig) (*GoogleBigQueryPublisher, error) {

	bigqueryClient, err := bigquery.NewClient(ctx, config.ProjectId)
	if err != nil {
		core.Error("failed to create google bigquery client: %v", err)
		return nil, err
	}

	publisher := &GoogleBigQueryPublisher{}

	if config.PublishChannelSize == 0 {
		config.PublishChannelSize = 10 * 1024
	}

	publisher.config = config
	if publisher.config.BatchDuration == 0 {
		publisher.config.BatchDuration = time.Second
	}

	tableInserter := publisher.bigqueryClient.Dataset(config.Dataset).Table(config.TableName).Inserter()

	publisher.bigqueryClient = bigqueryClient
	publisher.TableInserter = tableInserter
	publisher.PublishChannel = make(chan *bigquery.ValueSaver, config.PublishChannelSize)

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
			if len(publisher.messageBatch) > 0 {
				publisher.publishBatch(ctx)
			}
			break

		case message := <-publisher.PublishChannel:
			publisher.messageBatch = append(publisher.messageBatch, message)
			if len(publisher.messageBatch) >= publisher.config.BatchSize {
				publisher.publishBatch(ctx)
			}
			break
		}
	}
}

func (publisher *GoogleBigQueryPublisher) publishBatch(ctx context.Context) {

	publisher.mutex.RLock()
	entriesToSend := publisher.messageBatch
	publisher.mutex.RUnlock()

	err := publisher.TableInserter.Put(ctx, entriesToSend)
	if err != nil {
		core.Error("failed to publish bigquery entry: %v", err) // todo: update this failure case to something applicable to billing
		return
	}

	batchId := publisher.numBatchesSent
	batchNumMessages := len(publisher.messageBatch)

	publisher.mutex.Lock()
	publisher.numBatchesSent++
	publisher.numMessagesSent += batchNumMessages
	publisher.mutex.Unlock()

	publisher.messageBatch = []*bigquery.ValueSaver{}

	core.Debug("published batch %d containing %d messages", batchId, batchNumMessages)
}
