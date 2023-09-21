/*
   Network Next Accelerate.
   Copyright © 2017 - 2023 Network Next, Inc. All rights reserved.
*/

package main

// todo
/*
import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"time"

	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/constants"
	"github.com/networknext/next/modules/core"
	db "github.com/networknext/next/modules/database"
	"github.com/networknext/next/modules/envvar"
	"github.com/networknext/next/modules/portal"

	"github.com/gomodule/redigo/redis"
)

var apiKey = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZG1pbiI6dHJ1ZSwiZGF0YWJhc2UiOnRydWUsInBvcnRhbCI6dHJ1ZX0.QFPdb-RcP8wyoaOIBYeB_X6uA7jefGPVxm2VevJvpwU"

var apiPrivateKey = "this is the private key that generates API keys. make sure you change this value in production"

// ----------------------------------------------------------------------------------------

type PortalSliceData struct {
	Timestamp        string  `json:"timestamp"`
	SliceNumber      uint32  `json:"slice_number"`
	DirectRTT        uint32  `json:"direct_rtt"`
	NextRTT          uint32  `json:"next_rtt"`
	PredictedRTT     uint32  `json:"predicted_rtt"`
	DirectJitter     uint32  `json:"direct_jitter"`
	NextJitter       uint32  `json:"next_jitter"`
	RealJitter       uint32  `json:"real_jitter"`
	DirectPacketLoss float32 `json:"direct_packet_loss"`
	NextPacketLoss   float32 `json:"next_packet_loss"`
	RealPacketLoss   float32 `json:"real_packet_loss"`
	RealOutOfOrder   float32 `json:"real_out_of_order"`
	InternalEvents   string  `json:"internal_events"`
	SessionEvents    string  `json:"session_events"`
	DirectKbpsUp     uint32  `json:"direct_kbps_up"`
	DirectKbpsDown   uint32  `json:"direct_kbps_down"`
	NextKbpsUp       uint32  `json:"next_kbps_up"`
	NextKbpsDown     uint32  `json:"next_kbps_down"`
}

type PortalRelayData struct {
	RelayName    string `json:"relay_name"`
	RelayId      string `json:"relay_id"`
	RelayAddress string `json:"relay_address"`
	NumSessions  uint32 `json:"num_sessions"`
	MaxSessions  uint32 `json:"max_sessions"`
	StartTime    string `json:"start_time"`
	RelayFlags   string `json:"relay_flags"`
	RelayVersion string `json:"relay_version"`
}

type PortalNearRelayData struct {
	Timestamp           string                           `json:"timestamp"`
	NumNearRelays       uint32                           `json:"num_near_relays"`
	NearRelayId         [constants.MaxNearRelays]uint64  `json:"near_relay_id"`
	NearRelayRTT        [constants.MaxNearRelays]uint8   `json:"near_relay_rtt"`
	NearRelayJitter     [constants.MaxNearRelays]uint8   `json:"near_relay_jitter"`
	NearRelayPacketLoss [constants.MaxNearRelays]float32 `json:"near_relay_packet_loss"`
}

type PortalSessionData struct {
	SessionId      string  `json:"session_id"`
	ISP            string  `json:"isp"`
	ConnectionType uint8   `json:"connection_type"`
	PlatformType   uint8   `json:"platform_type"`
	Latitude       float32 `json:"latitude"`
	Longitude      float32 `json:"longitude"`
	DirectRTT      uint32  `json:"direct_rtt"`
	NextRTT        uint32  `json:"next_rtt"`
	BuyerId        string  `json:"buyer_id"`
	DatacenterId   string  `json:"datacenter_id"`
	ServerAddress  string  `json:"server_address"`
}

type PortalServerData struct {
	ServerAddress    string `json:"server_address"`
	SDKVersion_Major uint8  `json:"sdk_version_major"`
	SDKVersion_Minor uint8  `json:"sdk_version_minor"`
	SDKVersion_Patch uint8  `json:"sdk_version_patch"`
	BuyerId          string `json:"buyer_id"`
	DatacenterId     string `json:"datacenter_id"`
	NumSessions      uint32 `json:"num_sessions"`
	StartTime        string `json:"start_time"`
}

type PortalRelaySample struct {
	Timestamp                 string  `json:"timestamp"`
	NumSessions               uint32  `json:"num_sessions"`
	EnvelopeBandwidthUpKbps   uint32  `json:"envelope_bandwidth_up_kbps"`
	EnvelopeBandwidthDownKbps uint32  `json:"envelope_bandwidth_down_kbps"`
	PacketsSentPerSecond      float32 `json:"packets_sent_per_second"`
	PacketsReceivedPerSecond  float32 `json:"packets_recieved_per_second"`
	BandwidthSentKbps         float32 `json:"bandwidth_sent_kbps"`
	BandwidthReceivedKbps     float32 `json:"bandwidth_received_kbps"`
	NearPingsPerSecond        float32 `json:"near_pings_per_second"`
	RelayPingsPerSecond       float32 `json:"relay_pings_per_second"`
	RelayFlags                string  `json:"relay_flags"`
	NumRoutable               uint32  `json:"num_routable"`
	NumUnroutable             uint32  `json:"num_unroutable"`
	CurrentTime               string  `json:"current_time"`
}

// ----------------------------------------------------------------------------------------

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
	cmd.Env = append(cmd.Env, "ENABLE_DATABASE=false")
	cmd.Env = append(cmd.Env, "HTTP_PORT=50000")
	cmd.Env = append(cmd.Env, fmt.Sprintf("API_PRIVATE_KEY=%s", apiPrivateKey))

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

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
					userHash := uint64(j) + iteration
					score := uint32(rand.Intn(10000))
					next := ((uint64(j) + iteration) % 10) == 0

					sessionData := portal.GenerateRandomSessionData()

					sessionData.ServerAddress = "127.0.0.1:50000"

					sliceData := portal.GenerateRandomSliceData()

					sessionInserter.Insert(sessionId, userHash, score, next, sessionData, sliceData)

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

					id := 1 + uint64(k*threadCount+j)

					relayData.RelayName = fmt.Sprintf("local-%d", id)

					relayData.RelayAddress = fmt.Sprintf("127.0.0.1:%d", 2000+id)

					relayInserter.Insert(relayData, relaySample)
				}

				time.Sleep(time.Second)

				iteration++
			}
		}(k)
	}
}

func Get(url string, object interface{}) {

	fmt.Printf("Get URL: %s\n", url)

	buffer := new(bytes.Buffer)

	json.NewEncoder(buffer).Encode(object)

	request, err := http.NewRequest("GET", url, buffer)

	if err != nil {
		panic(err)
	}

	request.Header.Set("Authorization", "Bearer "+apiKey)

	client := &http.Client{}

	var response *http.Response
	for i := 0; i < 30; i++ {
		response, err = client.Do(request)
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

	fmt.Printf("--------------------------------------------------------------------\n")
	fmt.Printf("%s", body)
	fmt.Printf("--------------------------------------------------------------------\n")

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
	Sessions []PortalSessionData `json:"sessions"`
}

type PortalSessionDataResponse struct {
	SessionData   *PortalSessionData    `json:"session_data"`
	SliceData     []PortalSliceData     `json:"slice_data"`
	NearRelayData []PortalNearRelayData `json:"near_relay_data"`
}

type PortalServerCountResponse struct {
	ServerCount int `json:"server_count"`
}

type PortalServersResponse struct {
	Servers []PortalServerData `json:"servers"`
}

type PortalServerDataResponse struct {
	ServerData       *PortalServerData `json:"server_data"`
	ServerSessionIds []uint64          `json:"server_session_ids"`
}

type PortalRelayCountResponse struct {
	RelayCount int `json:"relay_count"`
}

type PortalRelaysResponse struct {
	Relays []PortalRelayData `json:"relays"`
}

type PortalRelayDataResponse struct {
	RelayData    *PortalRelayData    `json:"relay_data"`
	RelaySamples []PortalRelaySample `json:"relay_samples"`
}

func test_portal() {

	fmt.Printf("test_portal\n")

	redisClient := common.CreateRedisClient("127.0.0.1:6379")

	redisClient.Do("FLUSHALL")

	// create a dummy database

	database := db.CreateDatabase()

	database.CreationTime = "now"
	database.Creator = "test"
	database.BuyerMap[1] = &db.Buyer{Id: 1, Name: "buyer", Live: true, Debug: true}
	database.SellerMap[1] = &db.Seller{Id: 1, Name: "seller"}
	database.DatacenterMap[1] = &db.Datacenter{Id: 1, Name: "local", Latitude: 100, Longitude: 200}
	for i := 0; i < 1000; i++ {
		relayId := uint64(1 + i)
		relay := db.Relay{
			Id:            relayId,
			Name:          fmt.Sprintf("local-%d", i+1),
			PublicAddress: core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", 2000+i)),
			SSHAddress:    core.ParseAddress("127.0.0.1:22"),
			Datacenter:    database.DatacenterMap[1],
			Seller:        database.SellerMap[1],
		}
		database.Relays = append(database.Relays, relay)
		database.DatacenterRelays[1] = append(database.DatacenterRelays[1], uint64(1+i))
	}

	database.Fixup()

	fmt.Printf("database is %s\n", database.String())

	err := database.Validate()
	if err != nil {
		fmt.Printf("error: database did not validate: %v\n", err)
		os.Exit(1)
	}

	// save it to database.bin

	database.Save("database.bin")

	// run the API service, it will load the database

	api_cmd := api()

	// run redis insertion threads

	redisHostname := envvar.GetString("REDIS_HOSTNAME", "127.0.0.1:6379")

	redisPool := common.CreateRedisPool(redisHostname, 100, 1000)

	threadCount := envvar.GetInt("REDIS_THREAD_COUNT", 10)

	RunSessionInsertThreads(redisPool, threadCount)
	RunServerInsertThreads(redisPool, threadCount)
	RunRelayInsertThreads(redisPool, threadCount)

	// simulate portal activity

	var ready bool

	for i := 0; i < 60; i++ {

		fmt.Printf("iteration %d\n", i)

		sessionCountsResponse := PortalSessionCountsResponse{}

		Get("http://127.0.0.1:50000/portal/session_counts", &sessionCountsResponse)

		fmt.Printf("-------------------------------------------------------------\n")

		fmt.Printf("next sessions = %d, total sessions = %d\n", sessionCountsResponse.NextSessionCount, sessionCountsResponse.TotalSessionCount)

		sessionsResponse := PortalSessionsResponse{}

		Get("http://127.0.0.1:50000/portal/sessions/0/10", &sessionsResponse)

		fmt.Printf("got data for %d sessions\n", len(sessionsResponse.Sessions))

		sessionDataResponse := PortalSessionDataResponse{}

		if len(sessionsResponse.Sessions) > 0 {

			sessionId, err := strconv.ParseInt(sessionsResponse.Sessions[0].SessionId, 10, 64)
			if err != nil {
				panic(err)
			}

			fmt.Printf("first session id is %016x\n", sessionId)

			Get(fmt.Sprintf("http://127.0.0.1:50000/portal/session/%016x", sessionId), &sessionDataResponse)

			fmt.Printf("session %016x has %d slices, %d near relay data\n", sessionId, len(sessionDataResponse.SliceData), len(sessionDataResponse.NearRelayData))
		}

		serverCountResponse := PortalServerCountResponse{}

		Get("http://127.0.0.1:50000/portal/server_count", &serverCountResponse)

		fmt.Printf("servers = %d\n", serverCountResponse.ServerCount)

		serversResponse := PortalServersResponse{}

		Get("http://127.0.0.1:50000/portal/servers/0/10", &serversResponse)

		serverDataResponse := PortalServerDataResponse{}

		fmt.Printf("got data for %d servers\n", len(serversResponse.Servers))

		if len(serversResponse.Servers) > 0 {

			fmt.Printf("first server address is '%s'\n", serversResponse.Servers[0].ServerAddress)

			Get(fmt.Sprintf("http://127.0.0.1:50000/portal/server/%s", serversResponse.Servers[0].ServerAddress), &serverDataResponse)

			fmt.Printf("server %s has %d sessions\n", serversResponse.Servers[0].ServerAddress, len(serverDataResponse.ServerSessionIds))
		}

		Get("http://127.0.0.1:50000/portal/server_count", &serverCountResponse)

		relayCountResponse := PortalRelayCountResponse{}
		Get("http://127.0.0.1:50000/portal/relay_count", &relayCountResponse)

		fmt.Printf("relays = %d\n", relayCountResponse.RelayCount)

		relaysResponse := PortalRelaysResponse{}

		Get("http://127.0.0.1:50000/portal/relays/0/10", &relaysResponse)

		fmt.Printf("got data for %d relays\n", len(relaysResponse.Relays))

		relayDataResponse := PortalRelayDataResponse{}

		if len(relaysResponse.Relays) > 0 {

			fmt.Printf("first relay is '%s'\n", relaysResponse.Relays[0].RelayName)

			Get(fmt.Sprintf("http://127.0.0.1:50000/portal/relay/%s", relaysResponse.Relays[0].RelayName), &relayDataResponse)

			fmt.Printf("relay %s has %d samples\n", relaysResponse.Relays[0].RelayName, len(relayDataResponse.RelaySamples))
		}

		ready = true

		if sessionCountsResponse.NextSessionCount < 100 {
			fmt.Printf("A\n")
			ready = false
		}

		if sessionCountsResponse.TotalSessionCount < 1000 {
			fmt.Printf("B\n")
			ready = false
		}

		if len(sessionsResponse.Sessions) < 10 {
			fmt.Printf("C\n")
			ready = false
		}

		if sessionDataResponse.SessionData == nil {
			fmt.Printf("D\n")
			ready = false
		}

		if len(sessionDataResponse.SliceData) == 0 {
			fmt.Printf("E\n")
			ready = false
		}

		if len(sessionDataResponse.NearRelayData) == 0 {
			fmt.Printf("F\n")
			ready = false
		}

		if serverCountResponse.ServerCount != 1 {
			fmt.Printf("G\n")
			ready = false
		}

		if len(serverDataResponse.ServerSessionIds) < 10000 {
			fmt.Printf("H\n")
			ready = false
		}

		if relayCountResponse.RelayCount < 10 {
			fmt.Printf("I\n")
			ready = false
		}

		if len(relayDataResponse.RelaySamples) == 0 {
			fmt.Printf("J\n")
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
*/

// todo: dummy out
func main() {
	// ...
}
