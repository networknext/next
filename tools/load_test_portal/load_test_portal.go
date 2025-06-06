package main

import (
	"context"
	"fmt"
	"math/rand"
	"os"
	"os/exec"
	"time"

	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/envvar"
	"github.com/networknext/next/modules/portal"
)

var SessionCruncherURL = "http://127.0.0.1:40200"
var ServerCruncherURL = "http://127.0.0.1:40300"

func RunSessionInsertThreads(ctx context.Context, threadCount int) {

	for k := 0; k < threadCount; k++ {

		go func(thread int) {

			redisClient := common.CreateRedisClient("127.0.0.1:6379")

			sessionInserter := portal.CreateSessionInserter(ctx, redisClient, SessionCruncherURL, 10000)

			iteration := uint64(0)

			clientRelayInserter := portal.CreateClientRelayInserter(redisClient, 10000)

			serverRelayInserter := portal.CreateServerRelayInserter(redisClient, 10000)

			relay_max := uint64(0)

			time.Sleep(time.Duration(rand.Intn(10000)) * time.Millisecond)

			for {

				for j := 0; j < 10000; j++ {

					sessionId := uint64(thread*10000000000) + uint64(j) + iteration

					userHash := uint64(j) + iteration

					sessionData := portal.GenerateRandomSessionData()

					sessionData.SessionId = sessionId
					sessionData.UserHash = userHash
					sessionData.BuyerId = uint64(common.RandomInt(0, 9))

					sliceData := portal.GenerateRandomSliceData()

					next := (sessionId % 2) == 0

					score := uint32(sessionId % 1000)

					sessionInserter.Insert(ctx, sessionId, userHash, next, score, sessionData, sliceData)

					if sessionId > relay_max {

						clientRelayData := portal.GenerateRandomClientRelayData()
						clientRelayInserter.Insert(ctx, sessionId, clientRelayData)

						serverRelayData := portal.GenerateRandomServerRelayData()
						serverRelayInserter.Insert(ctx, sessionId, serverRelayData)

						relay_max = sessionId
					}
				}

				time.Sleep(10 * time.Second)

				iteration++
			}
		}(k)
	}
}

func RunServerInsertThreads(ctx context.Context, threadCount int) {

	for k := 0; k < threadCount; k++ {

		go func(thread int) {

			redisClient := common.CreateRedisClient("127.0.0.1:6379")

			serverInserter := portal.CreateServerInserter(context.Background(), redisClient, ServerCruncherURL, 10000)

			iteration := uint64(0)

			time.Sleep(time.Duration(rand.Intn(10000)) * time.Millisecond)

			for {

				for j := 0; j < 1000; j++ {

					serverData := portal.GenerateRandomServerData()

					id := uint32(iteration + uint64(j))

					serverData.ServerAddress = fmt.Sprintf("%d.%d.%d.%d:%d", id&0xFF, (id>>8)&0xFF, (id>>16)&0xFF, (id>>24)&0xFF, uint64(thread))
					serverData.BuyerId = uint64(common.RandomInt(0, 9))

					serverInserter.Insert(ctx, serverData)
				}

				time.Sleep(10 * time.Second)

				iteration++
			}
		}(k)
	}
}

func RunRelayInsertThreads(ctx context.Context, threadCount int) {

	for k := 0; k < threadCount; k++ {

		go func(thread int) {

			redisClient := common.CreateRedisClient("127.0.0.1:6379")

			relayInserter := portal.CreateRelayInserter(redisClient, 10000)

			iteration := uint64(0)

			time.Sleep(time.Duration(rand.Intn(10000)) * time.Millisecond)

			for {

				for j := 0; j < 10; j++ {

					relayData := portal.GenerateRandomRelayData()

					id := uint32(iteration + uint64(j))

					relayData.RelayAddress = fmt.Sprintf("%d.%d.%d.%d:%d", id&0xFF, (id>>8)&0xFF, (id>>16)&0xFF, (id>>24)&0xFF, uint64(thread))

					relayInserter.Insert(ctx, relayData)
				}

				time.Sleep(10 * time.Second)

				iteration++
			}
		}(k)
	}
}

func RunPollThread(ctx context.Context) {

	iteration := uint64(0)

	go func() {

		redisClient := common.CreateRedisClient("127.0.0.1:6379")

		topSessionsWatcher := portal.CreateTopSessionsWatcher(SessionCruncherURL)

		topServersWatcher := portal.CreateTopServersWatcher(ServerCruncherURL)

		mapDataWatcher := portal.CreateMapDataWatcher(SessionCruncherURL)

		for {

			// ------------------------------------------------------------------------------------------

			fmt.Printf("------------------------------------------------------------------------------------------------\n")

			begin := 0
			end := 1000

			sessions := topSessionsWatcher.GetSessions(begin, end)

			fmt.Printf("top sessions: %d\n", len(sessions))

			// ------------------------------------------------------------------------------------------

			start := time.Now()

			sessionList := portal.GetSessionList(ctx, redisClient, sessions)
			if sessionList != nil {
				fmt.Printf("session list %d (%.3fms)\n", len(sessionList), float64(time.Since(start).Milliseconds()))
			}

			// ------------------------------------------------------------------------------------------

			if len(sessionList) > 0 {

				start = time.Now()

				sessionId := sessionList[0].SessionId

				sessionData, sliceData, clientRelayData, serverRelayData := portal.GetSessionData(ctx, redisClient, sessionId)
				if sessionData != nil {
					fmt.Printf("session data %x -> %d slices, %d client relay data, %d server relay data (%.3fms)\n", sessionData.SessionId, len(sliceData), len(clientRelayData), len(serverRelayData), float64(time.Since(start).Milliseconds()))
				}
			}

			// ------------------------------------------------------------------------------------------

			start = time.Now()

			minutes := start.Unix() / 60

			userHash := uint64(0x131e)

			userSessionList := portal.GetUserSessionList(ctx, redisClient, userHash, minutes, 100)
			if userSessionList != nil {
				fmt.Printf("user session list %d (%.3fms)\n", len(userSessionList), float64(time.Since(start).Milliseconds()))
			}

			// ------------------------------------------------------------------------------------------

			start = time.Now()

			servers := topServersWatcher.GetTopServers()

			fmt.Printf("top servers: %d (%.3fms)\n", len(servers), float64(time.Since(start).Milliseconds()))

			// ------------------------------------------------------------------------------------------

			start = time.Now()

			serverAddresses := topServersWatcher.GetServers(0, 100)

			fmt.Printf("server addresses -> %d server addresses (%.3fms)\n", len(serverAddresses), float64(time.Since(start).Milliseconds()))

			// ------------------------------------------------------------------------------------------

			start = time.Now()

			mapData := mapDataWatcher.GetMapData()

			fmt.Printf("map data: %d bytes (%.3fms)\n", len(mapData), float64(time.Since(start).Milliseconds()))

			// ------------------------------------------------------------------------------------------

			if len(serverAddresses) > 0 {

				start := time.Now()

				serverAddress := serverAddresses[0]

				serverData, serverSessions := portal.GetServerData(ctx, redisClient, serverAddress, minutes)

				if serverData != nil {
					fmt.Printf("server data %s -> %d sessions (%.3fms)\n", serverData.ServerAddress, len(serverSessions), float64(time.Since(start).Milliseconds()))
				}
			}

			// ------------------------------------------------------------------------------------------

			start = time.Now()

			serverList := portal.GetServerList(ctx, redisClient, serverAddresses)
			if serverList != nil {
				fmt.Printf("server list %d (%.3fms)\n", len(serverList), float64(time.Since(start).Milliseconds()))
			}

			// ------------------------------------------------------------------------------------------

			start = time.Now()

			relayCount := portal.GetRelayCount(ctx, redisClient, minutes)

			fmt.Printf("relays: %d (%.3fms)\n", relayCount, float64(time.Since(start).Milliseconds()))

			// ------------------------------------------------------------------------------------------

			start = time.Now()

			relayAddresses := portal.GetRelayAddresses(ctx, redisClient, minutes, 0, 100)

			fmt.Printf("relay addresses -> %d relay addresses (%.3fms)\n", len(relayAddresses), float64(time.Since(start).Milliseconds()))

			// ------------------------------------------------------------------------------------------

			if len(relayAddresses) > 0 {

				start := time.Now()

				relayAddress := relayAddresses[0]

				relayData := portal.GetRelayData(ctx, redisClient, relayAddress)

				if relayData != nil {
					fmt.Printf("relay data %s (%.3fms)\n", relayData.RelayAddress, float64(time.Since(start).Milliseconds()))
				}
			}

			// ------------------------------------------------------------------------------------------

			start = time.Now()

			relayList := portal.GetRelayList(ctx, redisClient, relayAddresses)
			if serverList != nil {
				fmt.Printf("relay list %d (%.3fms)\n", len(relayList), float64(time.Since(start).Milliseconds()))
			}

			// ------------------------------------------------------------------------------------------

			time.Sleep(time.Second)

			iteration++
		}
	}()
}

func session_cruncher() *exec.Cmd {

	cmd := exec.Command("./dist/session_cruncher")
	if cmd == nil {
		panic("could not create session cruncher!\n")
		return nil
	}

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "ENABLE_REDIS_TIME_SERIES=true")
	cmd.Env = append(cmd.Env, "HTTP_PORT=40200")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	cmd.Start()

	return cmd
}

func server_cruncher() *exec.Cmd {

	cmd := exec.Command("./dist/server_cruncher")
	if cmd == nil {
		panic("could not create server cruncher!\n")
		return nil
	}

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "ENABLE_REDIS_TIME_SERIES=true")
	cmd.Env = append(cmd.Env, "HTTP_PORT=40300")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	cmd.Start()

	return cmd
}

func main() {

	threadCount := envvar.GetInt("THREAD_COUNT", 1)

	ctx := context.Background()

	session_cruncher_cmd := session_cruncher()
	server_cruncher_cmd := server_cruncher()

	RunSessionInsertThreads(ctx, threadCount)

	RunServerInsertThreads(ctx, threadCount)

	RunRelayInsertThreads(ctx, threadCount)

	RunPollThread(ctx)

	time.Sleep(time.Minute * 2)

	session_cruncher_cmd.Process.Signal(os.Interrupt)
	server_cruncher_cmd.Process.Signal(os.Interrupt)

	session_cruncher_cmd.Wait()
	server_cruncher_cmd.Wait()
}
