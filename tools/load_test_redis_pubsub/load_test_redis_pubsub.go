package main

import (
	"math/rand"
	"time"

	"github.com/networknext/backend/modules/envvar"
)

func RunProducerThreads(hostname string, threadCount int) {

	for k := 0; k < threadCount; k++ {

		go func(thread int) {

			iteration := uint64(0)

			time.Sleep(time.Duration(rand.Intn(10000)) * time.Millisecond)

			for {

				for j := 0; j < 1000; j++ {

					// todo
				}

				time.Sleep(10 * time.Second)

				iteration++
			}
		}(k)
	}
}

func RunConsumerThreads(hostname string, threadCount int) {

	for k := 0; k < threadCount; k++ {

		go func(thread int) {

			// todo

		}(k)
	}
}

func main() {

	redisHostname := envvar.GetString("REDIS_HOSTNAME", "127.0.0.1:6379")

	producerThreadCount := envvar.GetInt("PRODUCER_THREAD_COUNT", 1000)
	consumerThreadCount := envvar.GetInt("CONSUMER_THREAD_COUNT", 1000)

	RunProducerThreads(redisHostname, producerThreadCount)

	RunConsumerThreads(redisHostname, consumerThreadCount)

	time.Sleep(time.Minute)
}
