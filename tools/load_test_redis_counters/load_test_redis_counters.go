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

	config := common.RedisCountersConfig{
		RedisHostname: redisHostname,
		SumWindow: 1000,
	}

	publisher, err := common.CreateRedisCountersPublisher(context.Background(), config)
	if err != nil {
		panic("could not create redis counters publisher")
	}

	keys := []string{"a", "b", "c", "d", "e", "f", "g", "h", "i", "j", "k"}

	go func() {

		ticker := time.NewTicker(time.Millisecond*10)

		for {

			select {

			case <-ctx.Done():
				return

			case <-ticker.C:
				for i := range keys {
					publisher.MessageChannel <- keys[i]
				}
			}
		}

	}()
}

func RunWatcherThread(ctx context.Context, redisHostname string) {

	fmt.Printf("watcher\n")

	config := common.RedisCountersConfig{
		RedisHostname: redisHostname,
		DisplayWindow: 5000, // 5 second window in milliseconds
	}

	watcher, err := common.CreateRedisCountersWatcher(context.Background(), config)
	if err != nil {
		panic("could not create redis counters watcher")
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
				fmt.Printf("--------------------------------------------------------\n")
				watcher.Lock()
				for i := range keys {
					values := make([]int, 0)
					timestamps := make([]uint64, 0)
					watcher.GetIntValues(&timestamps, &values, keys[i])
					fmt.Printf("%s: %v => %v\n", keys[i], timestamps, values)
				}
				watcher.Unlock()
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
