package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"sync/atomic"
	"time"

	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/core"
	"github.com/networknext/next/modules/envvar"
)

func RunProducerThreads(ctx context.Context, hostname string, threadCount int, numMessagesSent *uint64) {

	producers := make([]*common.RedisPubsubProducer, threadCount)

	for i := 0; i < threadCount; i++ {

		var err error

		producers[i], err = common.CreateRedisPubsubProducer(ctx, common.RedisPubsubConfig{
			RedisHostname:      hostname,
			PubsubChannelName:  "test-channel",
			MessageChannelSize: 1024 * 1024,
			BatchSize:          10000,
			BatchDuration:      time.Second,
		})

		if err != nil {
			core.Error("failed to create redis pubsub producer: %v", err)
			os.Exit(1)
		}
	}

	for k := 0; k < threadCount; k++ {

		go func(thread int) {

			producer := producers[thread]

			time.Sleep(time.Duration(rand.Intn(10000)) * time.Millisecond)

			ticker := time.NewTicker(time.Second)

			for {

				select {

				case <-ctx.Done():
					return

				case <-ticker.C:
					const NumMessages = 1000
					for i := 0; i < NumMessages; i++ {
						messageData := [1024]byte{}
						producer.MessageChannel <- messageData[:]
					}
					atomic.AddUint64(numMessagesSent, NumMessages)
				}
			}
		}(k)
	}
}

func RunConsumerThreads(ctx context.Context, hostname string, threadCount int, numMessagesReceived *uint64) {

	consumers := make([]*common.RedisPubsubConsumer, threadCount)

	for i := 0; i < threadCount; i++ {

		var err error

		consumers[i], err = common.CreateRedisPubsubConsumer(ctx, common.RedisPubsubConfig{
			RedisHostname:      hostname,
			PubsubChannelName:  "test-channel",
			MessageChannelSize: 1024 * 1024,
		})

		if err != nil {
			core.Error("failed to create redis pubsub consumer: %v", err)
			os.Exit(1)
		}
	}

	for k := 0; k < threadCount; k++ {

		go func(thread int) {

			consumer := consumers[thread]

			for {

				select {

				case <-ctx.Done():
					fmt.Printf("consumer %d finished\n", thread)
					return

				case <-consumer.MessageChannel:
					atomic.AddUint64(numMessagesReceived, 1)
				}
			}

		}(k)
	}
}

func RunWatcherThread(ctx context.Context, numMessagesSent *uint64, numMessagesReceived *uint64) {

	go func() {

		ticker := time.NewTicker(time.Second)

		iteration := uint64(0)

		for {

			select {

			case <-ctx.Done():
				return

			case <-ticker.C:
				numSent := atomic.LoadUint64(numMessagesSent)
				numReceived := atomic.LoadUint64(numMessagesReceived)
				fmt.Printf("iteration %d: %d messages sent, %d messages received\n", iteration, numSent, numReceived)
				iteration++
			}
		}
	}()
}

func main() {

	redisHostname := envvar.GetString("REDIS_HOSTNAME", "127.0.0.1:6379")

	producerThreadCount := envvar.GetInt("PRODUCER_THREAD_COUNT", 10)
	consumerThreadCount := envvar.GetInt("CONSUMER_THREAD_COUNT", 20)

	var numMessagesSent uint64
	var numMessagesReceived uint64

	RunProducerThreads(context.Background(), redisHostname, producerThreadCount, &numMessagesSent)

	RunConsumerThreads(context.Background(), redisHostname, consumerThreadCount, &numMessagesReceived)

	RunWatcherThread(context.Background(), &numMessagesSent, &numMessagesReceived)

	time.Sleep(time.Minute)
}
