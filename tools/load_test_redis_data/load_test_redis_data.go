package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"time"

	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/core"
	"github.com/networknext/next/modules/envvar"
)

func RunRedisLeaderThreads(hostname string, threadCount int) {

	redisHostname := envvar.GetString("REDIS_HOSTNAME", "127.0.0.1:6379")

	redisClient := common.CreateRedisClient(redisHostname)

	ctx := context.Background()

	for k := 0; k < threadCount; k++ {

		go func(thread int) {

			iteration := uint64(0)

			time.Sleep(time.Duration(rand.Intn(10000)) * time.Millisecond)

			leaderElection, err := common.CreateRedisLeaderElection(redisClient, common.RedisLeaderElectionConfig{
				RedisHostname: hostname,
				ServiceName:   "load_test_redis_data",
			})
			if err != nil {
				core.Error("failed to create redis leader")
				os.Exit(1)
			}

			for {

				start := time.Now()

				leaderElection.Update(ctx)

				if leaderElection.IsLeader() {
					leaderElection.Store(ctx, "a", make([]byte, 1024*1024))
					leaderElection.Store(ctx, "b", make([]byte, 10*1024*1024))
					leaderElection.Store(ctx, "c", make([]byte, 100*1024*1024))
				}

				a := leaderElection.Load(ctx, "a")
				b := leaderElection.Load(ctx, "b")
				c := leaderElection.Load(ctx, "c")

				if a != nil && len(a) != 1024*1024 {
					panic("a should be 1mb")
				}

				if b != nil && len(b) != 10*1024*1024 {
					panic("b should be 10mb")
				}

				if c != nil && len(c) != 100*1024*1024 {
					panic("c should be 100mb")
				}

				fmt.Printf("iteration %d: update instance %d (%dms)\n", iteration, thread, time.Since(start).Milliseconds())

				time.Sleep(10 * time.Second)

				iteration++
			}
		}(k)
	}
}

func main() {

	redisHostname := envvar.GetString("REDIS_HOSTNAME", "127.0.0.1:6379")

	RunRedisLeaderThreads(redisHostname, 32)

	time.Sleep(time.Minute)
}
