package main

import (
	"fmt"
	"time"

	"github.com/networknext/next/modules/envvar"

	"github.com/redis/go-redis/v9"
)

var redisNodes []string

func RunSessionInsertThreads(threadCount int) {

	for k := 0; k < threadCount; k++ {

		go func(thread int) {

			// todo
			/*
			sessionInserter := portal.CreateSessionInserter(pool, 1000)

			iteration := uint64(0)

			time.Sleep(time.Duration(rand.Intn(10000)) * time.Millisecond)

			near_relay_max := uint64(0)

			for {

				for j := 0; j < 10000; j++ {

					sessionId := uint64(thread*1000000) + uint64(j) + iteration
					userHash := uint64(j) + iteration
					score := uint32(rand.Intn(10000))
					next := ((uint64(j) + iteration) % 10) == 0

					sessionData := portal.GenerateRandomSessionData()

					sliceData := portal.GenerateRandomSliceData()

					sessionInserter.Insert(sessionId, userHash, score, next, sessionData, sliceData)
				}

				time.Sleep(10 * time.Second)

				iteration++
			}
			*/
		}(k)
	}
}

func RunPollThread() {

	go func() {

		clusterOptions := redis.ClusterOptions{Addrs: redisNodes}

		redisClient := redis.NewClusterClient(&clusterOptions)

		for {

			fmt.Printf("-------------------------------------------------\n")

			start := time.Now()
			secs := start.Unix()

			minutes := secs / 60

			totalSessionCount, nextSessionCount := GetSessionCounts(redisClient, minutes)

			fmt.Printf("sessions: %d/%d (%.1fms)\n", nextSessionCount, totalSessionCount, float64(time.Since(start).Milliseconds()))

			// ------------------------------------------------------------------------------------------

			time.Sleep(time.Second)
		}
	}()
}

func main() {

	redisNodes = []string{"127.0.0.1:7000", "127.0.0.1:7001", "127.0.0.1:7002", "127.0.0.1:7003", "127.0.0.1:7004", "127.0.0.1:7005"}

	threadCount := envvar.GetInt("REDIS_THREAD_COUNT", 100)

	RunSessionInsertThreads(threadCount)

	RunPollThread()

	time.Sleep(time.Minute)
}

// --------------------------------------------------------------------------------------------------

func GetSessionCounts(client *redis.ClusterClient, minutes int64) (int, int) {

	/*
	redisClient := pool.Get()

	redisClient.Send("ZCARD", fmt.Sprintf("s-%d", minutes-1))
	redisClient.Send("ZCARD", fmt.Sprintf("s-%d", minutes))
	redisClient.Send("ZCARD", fmt.Sprintf("n-%d", minutes-1))
	redisClient.Send("ZCARD", fmt.Sprintf("n-%d", minutes))

	redisClient.Flush()

	totalSessionCount_a, err := redis.Int(redisClient.Receive())
	if err != nil {
		core.Error("redis get total sessions count a failed: %v", err)
		return 0, 0
	}

	totalSessionCount_b, err := redis.Int(redisClient.Receive())
	if err != nil {
		core.Error("redis get total sessions count b failed: %v", err)
		return 0, 0
	}

	nextSessionCount_a, err := redis.Int(redisClient.Receive())
	if err != nil {
		core.Error("redis get next sessions count a failed: %v", err)
		return 0, 0
	}

	nextSessionCount_b, err := redis.Int(redisClient.Receive())
	if err != nil {
		core.Error("redis get next sessions count b failed: %v", err)
		return 0, 0
	}

	redisClient.Close()

	totalSessionCount := totalSessionCount_a
	if totalSessionCount_b > totalSessionCount {
		totalSessionCount = totalSessionCount_b
	}

	nextSessionCount := nextSessionCount_a
	if nextSessionCount_b > nextSessionCount {
		nextSessionCount = nextSessionCount_b
	}

	return totalSessionCount, nextSessionCount
	*/

	return 0, 0
}

// ------------------------------------------------------------------------------------------------------------
