package common

import (
	"context"
	"sync"
	"time"

	"cloud.google.com/go/bigquery"
	"github.com/networknext/backend/modules/core"
)

type GoogleBigQueryConfig struct {
	// todo: bigquery specific config here
	BatchSize          int
	BatchDuration      time.Duration
	PublishChannelSize int
}

type GoogleBigQueryPublisher struct {
	PublishChannel chan *bigquery.ValueSaver
	config         GoogleBigQueryConfig
	// todo: bigquery specific variables here
	messageBatch    []*bigquery.ValueSaver
	batchStartTime  time.Time
	mutex           sync.RWMutex
	numMessagesSent int
	numBatchesSent  int
}

func CreateGoogleBigQueryPublisher(ctx context.Context, config GoogleBigQueryConfig) (*GoogleBigQueryPublisher, error) {

	// todo: create bigquery stuff

	publisher := &GoogleBigQueryPublisher{}

	if config.PublishChannelSize == 0 {
		config.PublishChannelSize = 10 * 1024
	}

	publisher.config = config
	if publisher.config.BatchDuration == 0 {
		publisher.config.BatchDuration = time.Second
	}

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

	// todo: publish batch to bigquery

	batchId := publisher.numBatchesSent
	batchNumMessages := len(publisher.messageBatch)

	publisher.mutex.Lock()
	publisher.numBatchesSent++
	publisher.numMessagesSent += batchNumMessages
	publisher.mutex.Unlock()

	publisher.messageBatch = []*bigquery.ValueSaver{}

	core.Debug("published batch %d containing %d messages", batchId, batchNumMessages)
}
