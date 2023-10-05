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

func CreateRedisTimeSeriesPublisher(ctx context.Context, config RedisTimeStreamsConfig) (*RedisTimeStreamsPublisher, error) {

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

	return producer, nil
}

// -------------------------------------------------------------------------------

func RunPublisherThread(ctx context.Context, redisHostname string) {

	config := RedisTimeSeriesConfig{
		RedisHostname: redisHostname,
	}

	publisher := CreateRedisTimeSeriesPublisher(config)

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
}

func RunWatcherThread(ctx context.Context, redisHostname string) {

	go func() {

		ticker := time.NewTicker(time.Second)

		iteration := uint64(0)

		for {

			select {

			case <-ctx.Done():
				return

			case <-ticker.C:
				fmt.Printf("iteration %d\n")
				iteration++
			}
		}
	}()
}

func main() {

	redisHostname := envvar.GetString("REDIS_HOSTNAME", "127.0.0.1:6379")

	RunPublisherThread(context.Background(), redisHostname)

	RunWatcherThread(context.Background(), redisHosthname)

	time.Sleep(time.Minute)
}
