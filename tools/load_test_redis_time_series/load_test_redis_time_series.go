package main

import (
	"context"
	"fmt"
	"time"

	"github.com/networknext/next/modules/envvar"
)

type RedisTimeSeriesConfig struct {
	RedisHostname      string
	RedisCluster       []string
	BatchSize          int
	BatchDuration      time.Duration
}

type RedisTimeSeriesPublisher struct {
	config          RedisTimeSeriesConfig
	// redisClient     redis.StreamCmdable
}

func CreateRedisTimeSeriesPublisher(ctx context.Context, config RedisTimeSeriesConfig) (*RedisTimeSeriesPublisher, error) {

	/*
	var redisClient redis.StreamCmdable
	if len(config.RedisCluster) > 0 {
		client := CreateRedisClusterClient(config.RedisCluster)
		_, err := client.Ping(ctx).Result()
		if err != nil {
			return nil, err
		}
		redisClient = client
	} else {
		client := CreateRedisClient(config.RedisHostname)
		_, err := client.Ping(ctx).Result()
		if err != nil {
			return nil, err
		}
		redisClient = client
	}
	*/

	publisher := &RedisTimeSeriesPublisher{}

	if config.BatchDuration == 0 {
		config.BatchDuration = time.Second
	}

	if config.BatchSize == 0 {
		config.BatchSize = 10000
	}

	publisher.config = config

	/*
	producer.redisClient = redisClient
	producer.MessageChannel = make(chan []byte, config.MessageChannelSize)

	go producer.updateMessageChannel(ctx)
	*/

	return publisher, nil
}

// -------------------------------------------------------------------------------

func RunPublisherThread(ctx context.Context, redisHostname string) {

	fmt.Printf("publisher\n")

	config := RedisTimeSeriesConfig{
		RedisHostname: redisHostname,
	}

	publisher, err := CreateRedisTimeSeriesPublisher(context.Background(), config)
	if err != nil {
		panic("could not create redis time series publisher")
	}

	go func() {

		ticker := time.NewTicker(time.Millisecond)

		for {

			select {

			case <-ctx.Done():
				return

			case <-ticker.C:

				// todo
				_ = publisher
			}
		}

	}()
}

func RunWatcherThread(ctx context.Context, redisHostname string) {

	fmt.Printf("watcher\n")

	go func() {

		ticker := time.NewTicker(time.Second)

		iteration := uint64(0)

		for {

			select {

			case <-ctx.Done():
				return

			case <-ticker.C:
				fmt.Printf("iteration %d\n", iteration)
				iteration++
			}
		}
	}()
}

func main() {

	redisHostname := envvar.GetString("REDIS_HOSTNAME", "127.0.0.1:6379")

	RunPublisherThread(context.Background(), redisHostname)

	RunWatcherThread(context.Background(), redisHostname)

	time.Sleep(time.Minute)
}
