package main

import (
	"context"
	"fmt"
	"time"
	"sync"

	"github.com/networknext/next/modules/envvar"
	"github.com/networknext/next/modules/common"
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

func (publisher *RedisTimeSeriesPublisher) Publish(keys []string, values []float64) {
	timestamp := uint64(time.Now().Unix())
	_ = timestamp
	// todo
	fmt.Printf("%d: publish %v\n", timestamp, values)
}

// -------------------------------------------------------------------------------

type RedisTimeSeriesWatcher struct {
	config RedisTimeSeriesConfig
	keys []string
	mutex sync.Mutex
}

func CreateRedisTimeSeriesWatcher(ctx context.Context, config RedisTimeSeriesConfig, keys[] string) (*RedisTimeSeriesWatcher, error) {

	watcher := &RedisTimeSeriesWatcher{}

	watcher.config = config
	watcher.keys = keys

	// todo: create watcher thread

	return watcher, nil 
}

func (watcher *RedisTimeSeriesWatcher) GetTimeSeries() (keys []string, timestamps []uint64, values [][]float64) {
	// todo
	return watcher.keys, []uint64{}, [][]float64{}
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

	keys := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"}

	values := make([]float64, len(keys))

	go func() {

		ticker := time.NewTicker(time.Millisecond)

		for {

			select {

			case <-ctx.Done():
				return

			case <-ticker.C:
				for i := range values {
					values[i] = float64(common.RandomInt(0,1000000))/10000.0
				}
				publisher.Publish(keys, values)
			}
		}

	}()
}

func RunWatcherThread(ctx context.Context, redisHostname string) {

	fmt.Printf("watcher\n")

	config := RedisTimeSeriesConfig{
		RedisHostname: redisHostname,
	}

	keys := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"}

	watcher, err := CreateRedisTimeSeriesWatcher(context.Background(), config, keys)
	if err != nil {
		panic("could not create redis time series watcher")
	}

	go func() {

		ticker := time.NewTicker(time.Second)

		iteration := uint64(0)

		for {

			select {

			case <-ctx.Done():
				return

			case <-ticker.C:
				fmt.Printf("iteration %d\n", iteration)
				keys, timestamps, values := watcher.GetTimeSeries()
				_ = keys
				_ = timestamps 
				_ = values
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
