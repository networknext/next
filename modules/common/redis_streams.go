package common

import (
	"context"
	"sync"
	"time"

	"github.com/go-redis/redis/v9"
	"github.com/networknext/backend/modules/core"
)

type RedisStreamsConfig struct {
	RedisHostname      string
	RedisPassword      string
	StreamName         string
	BatchSize          int
	BatchDuration      time.Duration
	MessageChannelSize int
}

type RedisStreamsProducer struct {
	MessageChannel  chan []byte
	config          RedisStreamsConfig
	redisClient     *redis.Client
	messageBatch    [][]byte
	batchStartTime  time.Time
	mutex           sync.RWMutex
	numMessagesSent int
	numBatchesSent  int
}

func CreateRedisStreamsProducer(ctx context.Context, config RedisStreamsConfig) (*RedisStreamsProducer, error) {

	redisClient := redis.NewClient(&redis.Options{
		Addr:     config.RedisHostname,
		Password: config.RedisPassword,
	})
	_, err := redisClient.Ping(ctx).Result()
	if err != nil {
		return nil, err
	}

	producer := &RedisStreamsProducer{}

	producer.config = config
	producer.redisClient = redisClient
	producer.MessageChannel = make(chan []byte, config.MessageChannelSize)

	go producer.updateMessageChannel(ctx)

	return producer, nil
}

func (producer *RedisStreamsProducer) updateMessageChannel(ctx context.Context) {

	ticker := time.NewTicker(producer.config.BatchDuration)

	for {
		select {

		case <-ctx.Done():
			return

		case <-ticker.C:
			if len(producer.messageBatch) > 0 {
				producer.sendBatchToRedis(ctx)
			}
			break

		case message := <-producer.MessageChannel:
			producer.messageBatch = append(producer.messageBatch, message)
			if len(producer.messageBatch) >= producer.config.BatchSize {
				producer.sendBatchToRedis(ctx)
			}
			break
		}
	}
}

func (producer *RedisStreamsProducer) sendBatchToRedis(ctx context.Context) {

	messageToSend := batchMessages(producer.numBatchesSent, producer.messageBatch)

	args := &redis.XAddArgs{
		Stream:     producer.config.StreamName,
		NoMkStream: false,
		Approx:     false,
		Limit:      0,
		ID:         "",
		Values:     map[string]interface{}{"type": "message", "data": messageToSend},
	}

	if _, err := producer.redisClient.XAdd(ctx, args).Result(); err != nil {
		core.Error("failed to send batch to redis: %v", err)
		return
	}

	batchId := producer.numBatchesSent
	batchNumMessages := len(producer.messageBatch)

	producer.mutex.Lock()
	producer.numBatchesSent++
	producer.numMessagesSent += batchNumMessages
	producer.mutex.Unlock()

	producer.messageBatch = [][]byte{}

	core.Debug("sent batch %d containing %d messages (%d bytes)", batchId, batchNumMessages, len(messageToSend))
}

/*
type Consumer struct {
	Config                   ConsumerConfig
	RedisDB                  *redis.Client
	numberOfMessagesReceived int64
	numberOfBatchesReceived  int64
}

type ConsumerConfig struct {
	RedisHostname     string
	RedisPassword     string
	StreamName        string
	ConsumerGroup     string
	BlockTimeout      time.Duration
	ConsumerBatchSize int
	ConsumerName      string
}

func NewConsumer(
	config ConsumerConfig,
) *Consumer {

	return &Consumer{
		Config:                   config,
		RedisDB:                  &redis.Client{},
		numberOfMessagesReceived: 0,
		numberOfBatchesReceived:  0,
	}
}

func (consumer *Consumer) Connect(ctx context.Context) error {
	consumer.RedisDB = redis.NewClient(&redis.Options{
		Addr:     consumer.Config.RedisHostname,
		Password: consumer.Config.RedisPassword,
	})
	_, err := consumer.RedisDB.Ping(ctx).Result()
	if err != nil {
		return err
	}
	return nil
}

func (consumer *Consumer) CreateConsumerGroup(ctx context.Context) error {
	//create consumerGroup with length of if group no created yet, if the group existed, cmd returns BUSYGROUP
	_, err := consumer.RedisDB.XGroupCreateMkStream(ctx, consumer.Config.StreamName, consumer.Config.ConsumerGroup, "0").Result()

	if !strings.Contains(fmt.Sprint(err), "BUSYGROUP") {
		//do not need to handle this error
		fmt.Printf("Consumer Group: %v already existed\n", consumer.Config.ConsumerGroup)
	}

	if err == context.Canceled {
		return err
	}

	consumer.Config.ConsumerName = uuid.NewV4().String()
	fmt.Printf("consumerName = %s\n", consumer.Config.ConsumerName)
	return nil
}

func (consumer *Consumer) ReceiveMessages(ctx context.Context) error {

	start := ">"

	streamMsgs, err := consumer.RedisDB.XReadGroup(ctx, &redis.XReadGroupArgs{
		Streams:  []string{consumer.Config.StreamName, start},
		Group:    consumer.Config.ConsumerGroup,
		Consumer: consumer.Config.ConsumerName,
		Count:    int64(consumer.Config.ConsumerBatchSize),
		Block:    consumer.Config.BlockTimeout,
		NoAck:    false,
	}).Result()

	if err == context.Canceled {
		return context.Canceled
	}

	if err == error(redis.Nil) {
		return err
	}

	if err != nil {
		core.Error("error reading redis stream: %s", err)
		return err
	}

	// Consume messages batch
	for _, stream := range streamMsgs[0].Messages {
		// split messages block
		messages := parseMessages([]byte(stream.Values["data"].(string)))
		core.Debug("Received %v messages (%v bytes)", len(messages), len([]byte(stream.Values["data"].(string))))
		consumer.RedisDB.XAck(ctx, consumer.Config.StreamName, consumer.Config.ConsumerGroup, stream.ID)
		// todo: do messages processing here, if it fails "continue"

		atomic.AddInt64(&consumer.numberOfBatchesReceived, 1)
		atomic.AddInt64(&consumer.numberOfMessagesReceived, int64(len(messages)))
	}

	return nil

}
func parseMessages(messages []byte) [][]byte {
	index := 0
	var batchNum uint32
	var numMessages uint32
	if !encoding.ReadUint32(messages, &index, &batchNum) {
		core.Debug("could not read batch number")
	}
	if !encoding.ReadUint32(messages, &index, &numMessages) {
		core.Debug("could not read number of messages")
	}
	messagesData := make([][]byte, numMessages)
	for i := 0; i < int(numMessages); i++ {
		var messageLength uint32
		if !encoding.ReadUint32(messages, &index, &messageLength) {
			core.Debug("could not read length of the message")
		}
		if !encoding.ReadBytes(messages, &index, &messagesData[i], uint32(messageLength)) {
			core.Debug("could not read message data")
		}
	}
	return messagesData
}

func (consumer *Consumer) NumMessageReceived() int64 {
	return atomic.LoadInt64(&consumer.numberOfMessagesReceived)
}

func (consumer *Consumer) NumBatchesReceived() int64 {
	return atomic.LoadInt64(&consumer.numberOfBatchesReceived)
}
*/
