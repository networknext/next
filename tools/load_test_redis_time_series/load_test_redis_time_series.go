package main

import (
	"context"
	"fmt"
	"time"

	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/envvar"
)

func RunPublisherThread(ctx context.Context, redisHostname string) {

	fmt.Printf("publisher\n")

	config := common.RedisTimeSeriesConfig{
		RedisHostname: redisHostname,
	}

	publisher, err := common.CreateRedisTimeSeriesPublisher(context.Background(), config)
	if err != nil {
		panic("could not create redis time series publisher")
	}

	keys := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"}

	go func() {

		ticker := time.NewTicker(time.Millisecond)

		for {

			select {

			case <-ctx.Done():
				return

			case <-ticker.C:
				message := common.RedisTimeSeriesMessage{}
				message.Timestamp = uint64(time.Now().Unix())
				message.Keys = keys
				message.Values = make([]float64, len(keys))
				for i := range message.Values {
					message.Values[i] = float64(common.RandomInt(0, 1000000)) / 10000.0
				}
				publisher.MessageChannel <- &message
			}
		}

	}()
}

func RunWatcherThread(ctx context.Context, redisHostname string) {

	fmt.Printf("watcher\n")

	config := common.RedisTimeSeriesConfig{
		RedisHostname: redisHostname,
	}

	watcher, err := common.CreateRedisTimeSeriesWatcher(context.Background(), config)
	if err != nil {
		panic("could not create redis time series watcher")
	}

	keys := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"}

	watcher.SetKeys(keys)

	go func() {

		ticker := time.NewTicker(time.Second)

		iteration := uint64(0)

		for {

			select {

			case <-ctx.Done():
				return

			case <-ticker.C:
				fmt.Printf("iteration %d\n", iteration)
				keyToIndex, timestamps, values := watcher.GetTimeSeries()
				_ = keyToIndex
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
