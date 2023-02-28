/*
   Network Next. You control the network.
   Copyright © 2017 - 2023 Network Next, Inc. All rights reserved.
*/

package main

import (
	"fmt"
	"os"
	"os/exec"
	"time"
	"math/rand"
	"net/http"
	"io/ioutil"
	"encoding/json"

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/portal"
	"github.com/networknext/backend/modules/envvar"

	"github.com/gomodule/redigo/redis"
)

func bash(command string) {

	cmd := exec.Command("bash", "-c", command)
	if cmd == nil {
		fmt.Printf("error: could not run bash!\n")
		os.Exit(1)
	}

	if err := cmd.Run(); err != nil {
		fmt.Printf("error: failed to run command: %v\n", err)
		os.Exit(1)
	}

	cmd.Wait()
}

func api() *exec.Cmd {

	cmd := exec.Command("./api")
	if cmd == nil {
		panic("could not create api!\n")
		return nil
	}

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "ENABLE_ADMIN=false")
	cmd.Env = append(cmd.Env, "HTTP_PORT=50000")

	cmd.Start()

	return cmd
}

func RunSessionInsertThreads(pool *redis.Pool, threadCount int) {

	for k := 0; k < threadCount; k++ {

		go func(thread int) {

			sessionInserter := portal.CreateSessionInserter(pool, 1000)

			nearRelayInserter := portal.CreateNearRelayInserter(pool, 1000)

			iteration := uint64(0)

			for {

				for j := 0; j < 1000; j++ {

					sessionId := uint64(thread*1000000) + uint64(j) + iteration
					score := uint32(rand.Intn(10000))
					next := ((uint64(j) + iteration) % 10) == 0

					sessionData := portal.GenerateRandomSessionData()

					sessionData.ServerAddress = "127.0.0.1:50000"

					sliceData := portal.GenerateRandomSliceData()

					sessionInserter.Insert(sessionId, score, next, sessionData, sliceData)

					nearRelayData := portal.GenerateRandomNearRelayData()
					nearRelayInserter.Insert(sessionId, nearRelayData)
				}

				time.Sleep(time.Second)

				iteration++
			}
		}(k)
	}
}

func RunServerInsertThreads(pool *redis.Pool, threadCount int) {

	for k := 0; k < threadCount; k++ {

		go func(thread int) {

			serverInserter := portal.CreateServerInserter(pool, 1000)

			for {

				serverData := portal.GenerateRandomServerData()

				serverData.ServerAddress = "127.0.0.1:50000"

				serverInserter.Insert(serverData)

				time.Sleep(time.Second)
			}
		}(k)
	}
}

func RunRelayInsertThreads(pool *redis.Pool, threadCount int) {

	for k := 0; k < threadCount; k++ {

		go func(thread int) {

			relayInserter := portal.CreateRelayInserter(pool, 1000)

			iteration := uint64(0)

			for {

				for j := 0; j < 10; j++ {

					relayData := portal.GenerateRandomRelayData()
					relaySample := portal.GenerateRandomRelaySample()

					id := uint32(iteration + uint64(j))

					relayData.RelayAddress = fmt.Sprintf("%d.%d.%d.%d:%d", id&0xFF, (id>>8)&0xFF, (id>>16)&0xFF, (id>>24)&0xFF, uint64(thread))

					relayInserter.Insert(relayData, relaySample)
				}

				time.Sleep(time.Second)

				iteration++
			}
		}(k)
	}
}

func Get(url string, object interface{}) {

	var err error
	var response *http.Response
	for i := 0; i < 30; i++ {
		response, err = http.Get(url)
		if err == nil {
			break
		}
		time.Sleep(time.Second)
	}

	if err != nil {
		panic(fmt.Sprintf("failed to read %s: %v", url, err))
	}

	body, error := ioutil.ReadAll(response.Body)
	if error != nil {
		panic(fmt.Sprintf("could not read response body for %s: %v", url, err))
	}

	response.Body.Close()

	err = json.Unmarshal([]byte(body), &object)
	if err != nil {
		panic(fmt.Sprintf("could not parse json response for %s: %v", url, err))
	}
}

type PortalSessionCountsResponse struct {
	TotalSessionCount int `json:"total_session_count"`
	NextSessionCount  int `json:"next_session_count"`
}

type PortalSessionsResponse struct {
	Sessions []portal.SessionEntry `json:"sessions"`
}

type PortalSessionDataResponse struct {
	SessionData   *portal.SessionData    `json:"session_data"`
	SliceData     []portal.SliceData     `json:"slice_data"`
	NearRelayData []portal.NearRelayData `json:"near_relay_data"`
}

type PortalServerCountResponse struct {
	ServerCount int `json:"server_count"`
}

type PortalServersResponse struct {
	Servers []portal.ServerEntry `json:"servers"`
}

type PortalServerDataResponse struct {
	ServerData       *portal.ServerData `json:"server_data"`
	ServerSessionIds []uint64           `json:"server_session_ids"`
}

type PortalRelayCountResponse struct {
	RelayCount int `json:"relay_count"`
}

type PortalRelaysResponse struct {
	Relays []portal.RelayEntry `json:"relays"`
}

type PortalRelayDataResponse struct {
	RelayData    *portal.RelayData    `json:"relay_data"`
	RelaySamples []portal.RelaySample `json:"relay_samples"`
}

func test_portal() {

	fmt.Printf("test_portal\n")

	redisClient := common.CreateRedisClient("127.0.0.1:6379")

	redisClient.Do("FLUSHALL")

	api_cmd := api()

	redisHostname := envvar.GetString("REDIS_HOSTNAME", "127.0.0.1:6379")

	redisPool := common.CreateRedisPool(redisHostname, 100, 1000)

	threadCount := envvar.GetInt("REDIS_THREAD_COUNT", 10)

	RunSessionInsertThreads(redisPool, threadCount)
	RunServerInsertThreads(redisPool, threadCount)
	RunRelayInsertThreads(redisPool, threadCount)

	var ready bool

	for i := 0; i < 10; i++ {

		sessionCountsResponse := PortalSessionCountsResponse{}

		Get("http://127.0.0.1:50000/portal/session_counts", &sessionCountsResponse)

		fmt.Printf("-------------------------------------------------------------\n")

		fmt.Printf("next sessions = %d, total sessions = %d\n", sessionCountsResponse.NextSessionCount, sessionCountsResponse.TotalSessionCount)

		sessionsResponse := PortalSessionsResponse{}

		Get("http://127.0.0.1:50000/portal/sessions/0/1000", &sessionsResponse)

		sessionDataResponse := PortalSessionDataResponse{}

		if len(sessionsResponse.Sessions) > 0 {

			Get(fmt.Sprintf("http://127.0.0.1:50000/portal/session/%d", sessionsResponse.Sessions[0].SessionId), &sessionDataResponse)

			fmt.Printf("session %016x has %d slices, %d near relay data\n", sessionsResponse.Sessions[0].SessionId, len(sessionDataResponse.SliceData), len(sessionDataResponse.NearRelayData))
		}

		serverCountResponse := PortalServerCountResponse{}

		Get("http://127.0.0.1:50000/portal/server_count", &serverCountResponse)

		fmt.Printf("servers = %d\n", serverCountResponse.ServerCount)

		serversResponse := PortalServersResponse{}

		Get("http://127.0.0.1:50000/portal/servers/0/1000", &serversResponse)

		serverDataResponse := PortalServerDataResponse{}

		if len(serversResponse.Servers) > 0 {

			Get(fmt.Sprintf("http://127.0.0.1:50000/portal/server/%s", serversResponse.Servers[0].Address), &serverDataResponse)

			fmt.Printf("server %s has %d sessions\n", serversResponse.Servers[0].Address, len(serverDataResponse.ServerSessionIds))
		}

		Get("http://127.0.0.1:50000/portal/server_count", &serverCountResponse)

		relayCountResponse := PortalRelayCountResponse{}

		Get("http://127.0.0.1:50000/portal/relay_count", &relayCountResponse)

		fmt.Printf("relays = %d\n", relayCountResponse.RelayCount)

		relaysResponse := PortalRelaysResponse{}

		Get("http://127.0.0.1:50000/portal/relays/0/1000", &relaysResponse)

		relayDataResponse := PortalRelayDataResponse{}

		if len(relaysResponse.Relays) > 0 {

			Get(fmt.Sprintf("http://127.0.0.1:50000/portal/relay/%s", relaysResponse.Relays[0].Address), &relayDataResponse)

			fmt.Printf("relay %s has %d samples\n", relaysResponse.Relays[0].Address, len(relayDataResponse.RelaySamples))
		}

		ready = true

		if sessionCountsResponse.NextSessionCount < 100 {
			ready = false
		}

		if sessionCountsResponse.TotalSessionCount < 1000 {
			ready = false
		}

		if len(sessionsResponse.Sessions) < 1000 {
			ready = false
		}

		if sessionDataResponse.SessionData == nil {
			ready = false
		}

		if len(sessionDataResponse.SliceData) == 0 {
			ready = false
		}

		if len(sessionDataResponse.NearRelayData) == 0 {
			ready = false
		}

		if serverCountResponse.ServerCount != 1 {
			ready = false
		}

		if len(serverDataResponse.ServerSessionIds) < 10000 {
			ready = false
		}

		if relayCountResponse.RelayCount < 100 {
			ready = false
		}

		if len(relayDataResponse.RelaySamples) == 0 {
			ready = false
		}

		fmt.Printf("-------------------------------------------------------------\n")

		if ready {
			break
		}

		time.Sleep(time.Second)
	}

	api_cmd.Process.Signal(os.Interrupt)
	api_cmd.Wait()

	if !ready {
		fmt.Printf("error: portal API is broken\n")
		os.Exit(1)
	}
}

func main() {
	test_portal()
}
