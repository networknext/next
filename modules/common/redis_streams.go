package common

// import (
// 	"context"
// 	"fmt"
// 	"strings"
// 	"sync/atomic"
// 	"time"

// 	"github.com/go-redis/redis/v9"
// 	"github.com/networknext/backend/modules/core"
// 	"github.com/networknext/backend/modules/encoding"
// 	uuid "github.com/satori/go.uuid"
// )

// /**
// ********     PRODUCER     *********
// **/
// type Producer struct {
// 	config               ProducerConfig
// 	RedisDB              *redis.Client
// 	numberOfMessagesSent int64
// 	numberOfBatchesSent  int64
// }

// type ProducerConfig struct {
// 	RedisHostname string
// 	RedisPassword string
// 	StreamName    string
// 	BatchSize     int
// 	BatchBytes    int
// 	TimeInterval  time.Duration
// }

// func NewProducer(
// 	config ProducerConfig,
// ) *Producer {

// 	return &Producer{
// 		config:               config,
// 		RedisDB:              &redis.Client{},
// 		numberOfMessagesSent: 0,
// 		numberOfBatchesSent:  0,
// 	}

// }

// func (producer *Producer) Connect(ctx context.Context) error {
// 	producer.RedisDB = redis.NewClient(&redis.Options{
// 		Addr:     producer.config.RedisHostname,
// 		Password: producer.config.RedisPassword,
// 	})
// 	_, err := producer.RedisDB.Ping(ctx).Result()
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (producer *Producer) SendMessages(
// 	ctx context.Context,
// 	messagesBatch [][]byte,
// 	start time.Time,
// ) ([][]byte, time.Time, error) {

// 	//check if timeout or batchSize reached -> send the messages block
// 	if time.Since(start) > producer.config.TimeInterval || len(messagesBatch) >= producer.config.BatchSize {
// 		messageToSend := buildMessages(messagesBatch, producer.config.BatchBytes, uint32(producer.NumBatchesSent()))

// 		if _, err := producer.RedisDB.XAdd(ctx, &redis.XAddArgs{
// 			Stream:     producer.config.StreamName,
// 			NoMkStream: false,
// 			Approx:     false,
// 			Limit:      0,
// 			ID:         "",
// 			Values:     map[string]interface{}{"type": "message", "data": messageToSend},
// 		}).Result(); err != nil {
// 			return messagesBatch, start, err
// 		}

// 		core.Debug("Sent batch of %v messages (%v bytes) with batch ID: %d\n", len(messagesBatch), len(messageToSend), producer.NumBatchesSent())

// 		atomic.AddInt64(&producer.numberOfBatchesSent, 1)
// 		atomic.AddInt64(&producer.numberOfMessagesSent, int64(len(messagesBatch)))

// 		messagesBatch = messagesBatch[:0]
// 		start = time.Now()

// 	}
// 	return messagesBatch, start, nil

// }

// func (producer *Producer) NumMessagesSent() int64 {
// 	return atomic.LoadInt64(&producer.numberOfMessagesSent)
// }

// func (producer *Producer) NumBatchesSent() int64 {
// 	return atomic.LoadInt64(&producer.numberOfBatchesSent)
// }

// func buildMessages(messages [][]byte, batchBytes int, batchNum uint32) []byte {
// 	data := make([]byte, batchBytes)
// 	index := 0
// 	encoding.WriteUint32(data, &index, batchNum)
// 	encoding.WriteUint32(data, &index, uint32(len(messages)))
// 	for i := range messages {
// 		encoding.WriteUint32(data, &index, uint32(len(messages[i])))
// 		encoding.WriteBytes(data, &index, messages[i], len(messages[i]))
// 	}
// 	return data[:index]
// }

// /**
// ********     CONSUMER     *********
// **/

// type Consumer struct {
// 	Config                   ConsumerConfig
// 	RedisDB                  *redis.Client
// 	numberOfMessagesReceived int64
// 	numberOfBatchesReceived  int64
// }

// type ConsumerConfig struct {
// 	RedisHostname     string
// 	RedisPassword     string
// 	StreamName        string
// 	ConsumerGroup     string
// 	BlockTimeout      time.Duration
// 	ConsumerBatchSize int
// 	ConsumerName      string
// }

// func NewConsumer(
// 	config ConsumerConfig,
// ) *Consumer {

// 	return &Consumer{
// 		Config:                   config,
// 		RedisDB:                  &redis.Client{},
// 		numberOfMessagesReceived: 0,
// 		numberOfBatchesReceived:  0,
// 	}
// }

// func (consumer *Consumer) Connect(ctx context.Context) error {
// 	consumer.RedisDB = redis.NewClient(&redis.Options{
// 		Addr:     consumer.Config.RedisHostname,
// 		Password: consumer.Config.RedisPassword,
// 	})
// 	_, err := consumer.RedisDB.Ping(ctx).Result()
// 	if err != nil {
// 		return err
// 	}
// 	return nil
// }

// func (consumer *Consumer) CreateConsumerGroup(ctx context.Context) error {
// 	//create consumerGroup with length of if group no created yet, if the group existed, cmd returns BUSYGROUP
// 	_, err := consumer.RedisDB.XGroupCreateMkStream(ctx, consumer.Config.StreamName, consumer.Config.ConsumerGroup, "0").Result()

// 	if !strings.Contains(fmt.Sprint(err), "BUSYGROUP") {
// 		//do not need to handle this error
// 		fmt.Printf("Consumer Group: %v already existed\n", consumer.Config.ConsumerGroup)
// 	}

// 	if err == context.Canceled {
// 		return err
// 	}

// 	consumer.Config.ConsumerName = uuid.NewV4().String()
// 	fmt.Printf("consumerName = %s\n", consumer.Config.ConsumerName)
// 	return nil
// }

// func (consumer *Consumer) ReceiveMessages(ctx context.Context) error {

// 	start := ">"

// 	streamMsgs, err := consumer.RedisDB.XReadGroup(ctx, &redis.XReadGroupArgs{
// 		Streams:  []string{consumer.Config.StreamName, start},
// 		Group:    consumer.Config.ConsumerGroup,
// 		Consumer: consumer.Config.ConsumerName,
// 		Count:    int64(consumer.Config.ConsumerBatchSize),
// 		Block:    consumer.Config.BlockTimeout,
// 		NoAck:    false,
// 	}).Result()

// 	if err == context.Canceled {
// 		return context.Canceled
// 	}

// 	if err == error(redis.Nil) {
// 		return err
// 	}

// 	if err != nil {
// 		core.Error("error reading redis stream: %s", err)
// 		return err
// 	}

// 	// Consume messages batch
// 	for _, stream := range streamMsgs[0].Messages {
// 		// split messages block
// 		messages := parseMessages([]byte(stream.Values["data"].(string)))
// 		core.Debug("Received %v messages (%v bytes)", len(messages), len([]byte(stream.Values["data"].(string))))
// 		consumer.RedisDB.XAck(ctx, consumer.Config.StreamName, consumer.Config.ConsumerGroup, stream.ID)
// 		// todo: do messages processing here, if it fails "continue"

// 		atomic.AddInt64(&consumer.numberOfBatchesReceived, 1)
// 		atomic.AddInt64(&consumer.numberOfMessagesReceived, int64(len(messages)))
// 	}

// 	return nil

// }
// func parseMessages(messages []byte) [][]byte {
// 	index := 0
// 	var batchNum uint32
// 	var numMessages uint32
// 	if !encoding.ReadUint32(messages, &index, &batchNum) {
// 		core.Debug("could not read batch number")
// 	}
// 	if !encoding.ReadUint32(messages, &index, &numMessages) {
// 		core.Debug("could not read number of messages")
// 	}
// 	messagesData := make([][]byte, numMessages)
// 	for i := 0; i < int(numMessages); i++ {
// 		var messageLength uint32
// 		if !encoding.ReadUint32(messages, &index, &messageLength) {
// 			core.Debug("could not read length of the message")
// 		}
// 		if !encoding.ReadBytes(messages, &index, &messagesData[i], uint32(messageLength)) {
// 			core.Debug("could not read message data")
// 		}
// 	}
// 	return messagesData
// }

// func (consumer *Consumer) NumMessageReceived() int64 {
// 	return atomic.LoadInt64(&consumer.numberOfMessagesReceived)
// }

// func (consumer *Consumer) NumBatchesReceived() int64 {
// 	return atomic.LoadInt64(&consumer.numberOfBatchesReceived)
// }
