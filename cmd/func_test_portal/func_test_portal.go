/*
   Network Next. You control the network.
   Copyright © 2017 - 2023 Network Next, Inc. All rights reserved.
*/

package main

import (
	"bytes"
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

func api() (*exec.Cmd, *bytes.Buffer) {

	cmd := exec.Command("./api")
	if cmd == nil {
		panic("could not create api!\n")
		return nil, nil
	}

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "ENABLE_ADMIN=false")
	cmd.Env = append(cmd.Env, "HTTP_PORT=50000")

	var output bytes.Buffer
	cmd.Stdout = &output
	cmd.Stderr = &output
	cmd.Start()

	return cmd, &output
}

func RunSessionInsertThreads(pool *redis.Pool, threadCount int) {

	for k := 0; k < threadCount; k++ {

		go func(thread int) {

			sessionInserter := portal.CreateSessionInserter(pool, 1000)

			nearRelayInserter := portal.CreateNearRelayInserter(pool, 1000)

			iteration := uint64(0)

			near_relay_max := uint64(0)

			for {

				for j := 0; j < 1000; j++ {

					sessionId := uint64(thread*1000000) + uint64(j) + iteration
					score := uint32(rand.Intn(10000))
					next := ((uint64(j) + iteration) % 10) == 0

					sessionData := portal.GenerateRandomSessionData()

					sliceData := portal.GenerateRandomSliceData()

					sessionInserter.Insert(sessionId, score, next, sessionData, sliceData)

					if sessionId > near_relay_max {
						nearRelayData := portal.GenerateRandomNearRelayData()
						nearRelayInserter.Insert(sessionId, nearRelayData)
						near_relay_max = sessionId
					}
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

			iteration := uint64(0)

			for {

				for j := 0; j < 100; j++ {

					serverData := portal.GenerateRandomServerData()

					id := uint32(iteration + uint64(j))

					serverData.ServerAddress = fmt.Sprintf("%d.%d.%d.%d:%d", id&0xFF, (id>>8)&0xFF, (id>>16)&0xFF, (id>>24)&0xFF, uint64(thread))

					serverInserter.Insert(serverData)
				}

				time.Sleep(time.Second)

				iteration++
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

func test_portal() {

	fmt.Printf("test_portal\n")

	api_cmd, _ := api()

	defer func() {
		api_cmd.Process.Signal(os.Interrupt)
		api_cmd.Wait()
	}()

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

			fmt.Printf("session %x has %d slices, %d near relay data\n", sessionsResponse.Sessions[0].SessionId, len(sessionDataResponse.SliceData), len(sessionDataResponse.NearRelayData))
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

		fmt.Printf("-------------------------------------------------------------\n")

		if ready {
			break
		}

		time.Sleep(time.Second)
	}

	if !ready {
		fmt.Printf("error: portal API is broken\n")
		os.Exit(1)
	}
}

func main() {
	test_portal()
}
