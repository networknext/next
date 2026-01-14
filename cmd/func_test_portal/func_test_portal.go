/*
   Network Next. Copyright 2017 - 2026 Network Next, Inc.
   Licensed under the Network Next Source Available License 1.0
*/

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
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
)

var SessionCruncherURL = "http://127.0.0.1:40200"
var ServerCruncherURL = "http://127.0.0.1:40300"

const TestAPIKey = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhZG1pbiI6dHJ1ZSwicG9ydGFsIjp0cnVlLCJpc3MiOiJuZXh0IGtleWdlbiIsImlhdCI6MTc0OTczODE4OX0.I89NXJCRMU_pIjnSleAnbux5HNsHhymzQ_SVatFo3b4"
const TestAPIPrivateKey = "uKUsmTySUVEssqBmVNciJWWolchcGGhFzRWMpydwOtVExvqYpHMotnkanNTaGHHh"

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
	DirectKbpsUp     uint32  `json:"bandwidth_kbps_up"`
	DirectKbpsDown   uint32  `json:"bandwidth_kbps_down"`
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

type PortalClientRelayData struct {
	Timestamp             string                             `json:"timestamp"`
	NumClientRelays       uint32                             `json:"num_client_relays"`
	ClientRelayId         [constants.MaxClientRelays]uint64  `json:"relay_id"`
	ClientRelayRTT        [constants.MaxClientRelays]uint8   `json:"relay_rtt"`
	ClientRelayJitter     [constants.MaxClientRelays]uint8   `json:"relay_jitter"`
	ClientRelayPacketLoss [constants.MaxClientRelays]float32 `json:"relay_packet_loss"`
}

type PortalServerRelayData struct {
	Timestamp             string                             `json:"timestamp"`
	NumServerRelays       uint32                             `json:"num_server_relays"`
	ServerRelayId         [constants.MaxServerRelays]uint64  `json:"relay_id"`
	ServerRelayRTT        [constants.MaxServerRelays]uint8   `json:"relay_rtt"`
	ServerRelayJitter     [constants.MaxServerRelays]uint8   `json:"relay_jitter"`
	ServerRelayPacketLoss [constants.MaxServerRelays]float32 `json:"relay_packet_loss"`
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
	ServerId       uint64  `json:"server_id,string"`
}

type PortalServerData struct {
	ServerId         uint64 `json:"server_id,string"`
	SDKVersion_Major uint8  `json:"sdk_version_major"`
	SDKVersion_Minor uint8  `json:"sdk_version_minor"`
	SDKVersion_Patch uint8  `json:"sdk_version_patch"`
	BuyerId          string `json:"buyer_id"`
	DatacenterId     string `json:"datacenter_id"`
	NumSessions      uint32 `json:"num_sessions"`
	StartTime        string `json:"start_time"`
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
	cmd.Env = append(cmd.Env, fmt.Sprintf("API_PRIVATE_KEY=%s", TestAPIPrivateKey))

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	cmd.Start()

	return cmd
}

func session_cruncher() *exec.Cmd {

	cmd := exec.Command("./session_cruncher")
	if cmd == nil {
		panic("could not create session cruncher!\n")
		return nil
	}

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "HTTP_PORT=40200")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	cmd.Start()

	return cmd
}

func server_cruncher() *exec.Cmd {

	cmd := exec.Command("./server_cruncher")
	if cmd == nil {
		panic("could not create server cruncher!\n")
		return nil
	}

	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "HTTP_PORT=40300")

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stdout

	cmd.Start()

	return cmd
}

func RunSessionInsertThreads(threadCount int) {

	for k := 0; k < threadCount; k++ {

		go func(thread int) {

			redisClient := common.CreateRedisClient("127.0.0.1:6379")

			sessionInserter := portal.CreateSessionInserter(context.Background(), redisClient, SessionCruncherURL, 1000)

			clientRelayInserter := portal.CreateClientRelayInserter(redisClient, 1000)

			serverRelayInserter := portal.CreateServerRelayInserter(redisClient, 1000)

			iteration := uint64(0)

			for {

				for j := 0; j < 1000; j++ {

					sessionId := uint64(thread*1000000) + uint64(j) + iteration
					next := ((uint64(j) + iteration) % 10) == 0

					sessionData := portal.GenerateRandomSessionData()

					sessionData.SessionId = sessionId
					sessionData.BuyerId = uint64(common.RandomInt(0, 9))
					sessionData.ServerId = common.HashString("127.0.0.1:50000")

					sliceData := portal.GenerateRandomSliceData()

					score := uint32(sessionId % 1000)

					sessionInserter.Insert(context.Background(), sessionId, next, score, sessionData, sliceData)

					clientRelayData := portal.GenerateRandomClientRelayData()
					clientRelayInserter.Insert(context.Background(), sessionId, clientRelayData)

					serverRelayData := portal.GenerateRandomServerRelayData()
					serverRelayInserter.Insert(context.Background(), sessionId, serverRelayData)
				}

				time.Sleep(time.Second)

				iteration++
			}
		}(k)
	}
}

func RunServerInsertThreads(threadCount int) {

	for k := 0; k < threadCount; k++ {

		go func(thread int) {

			redisClient := common.CreateRedisClient("127.0.0.1:6379")

			serverInserter := portal.CreateServerInserter(context.Background(), redisClient, ServerCruncherURL, 1000)

			for {

				serverData := portal.GenerateRandomServerData()

				serverData.ServerId = common.HashString("127.0.0.1:50000")

				serverInserter.Insert(context.Background(), serverData)

				time.Sleep(time.Second)
			}
		}(k)
	}
}

func RunRelayInsertThreads(threadCount int) {

	for k := 0; k < threadCount; k++ {

		go func(thread int) {

			redisClient := common.CreateRedisClient("127.0.0.1:6379")

			relayInserter := portal.CreateRelayInserter(redisClient, 1000)

			iteration := uint64(0)

			for {

				for j := 0; j < 10; j++ {

					relayData := portal.GenerateRandomRelayData()

					id := 1 + uint64(k*threadCount+j)

					relayData.RelayName = fmt.Sprintf("local-%03d", id)

					relayInserter.Insert(context.Background(), relayData)
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

	request.Header.Set("Authorization", "Bearer "+TestAPIKey)

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

	fmt.Printf("status code is %d\n", response.StatusCode)

	if response.StatusCode != 200 {
		fmt.Sprintf("warning: status code %d for %s", response.StatusCode, url)
		return
	}

	body, error := io.ReadAll(response.Body)
	if error != nil {
		panic(fmt.Sprintf("could not read response body for %s: %v", url, err))
	}

	response.Body.Close()

	// fmt.Printf("--------------------------------------------------------------------\n")
	// fmt.Printf("%s", body)
	// fmt.Printf("--------------------------------------------------------------------\n")

	err = json.Unmarshal([]byte(body), &object)
	if err != nil {
		panic(fmt.Sprintf("could not parse json response for %s: %v", url, err))
	}
}

func GetBinary(url string) []byte {

	var err error
	var response *http.Response
	req, err := http.NewRequest("GET", url, bytes.NewBuffer(nil))
	req.Header.Set("Authorization", "Bearer "+TestAPIKey)
	client := &http.Client{}
	response, err = client.Do(req)

	if err != nil {
		panic(fmt.Sprintf("failed to read %s: %v", url, err))
	}

	if response == nil {
		core.Error("no response from %s", url)
		os.Exit(1)
	}

	if response.StatusCode != 200 {
		panic(fmt.Sprintf("got %d response for %s", response.StatusCode, url))
	}

	body, error := io.ReadAll(response.Body)
	if error != nil {
		panic(fmt.Sprintf("could not read response body for %s: %v", url, err))
	}

	response.Body.Close()

	return body
}

type PortalSessionCountsResponse struct {
	NextSessionCount  int `json:"next_session_count"`
	TotalSessionCount int `json:"total_session_count"`
}

type PortalSessionsResponse struct {
	Sessions []PortalSessionData `json:"sessions"`
}

type PortalSessionDataResponse struct {
	SessionData     *PortalSessionData      `json:"session_data"`
	SliceData       []PortalSliceData       `json:"slice_data"`
	ClientRelayData []PortalClientRelayData `json:"client_relay_data"`
	ServerRelayData []PortalServerRelayData `json:"server_relay_data"`
}

type PortalServerCountResponse struct {
	ServerCount int `json:"server_count"`
}

type PortalServersResponse struct {
	Servers []PortalServerData `json:"servers"`
}

type PortalServerDataResponse struct {
	ServerData       *PortalServerData    `json:"server_data"`
	ServerSessionIds []*PortalSessionData `json:"server_sessions"`
}

type PortalRelayCountResponse struct {
	RelayCount int `json:"relay_count"`
}

type PortalRelaysResponse struct {
	Relays []PortalRelayData `json:"relays"`
}

type PortalRelayDataResponse struct {
	RelayData *PortalRelayData `json:"relay_data"`
}

func test_portal() {

	fmt.Printf("test_portal\n")

	redisClient := common.CreateRedisClient("127.0.0.1:6379")

	redisClient.FlushAll(context.Background())

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
			Name:          fmt.Sprintf("local-%03d", i+1),
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

	// run the session and server crunchers, they handles high load work that is too intense for redis

	session_cruncher_cmd := session_cruncher()

	server_cruncher_cmd := server_cruncher()

	time.Sleep(10 * time.Second)

	// run the API service

	api_cmd := api()

	// run redis insertion threads

	threadCount := envvar.GetInt("REDIS_THREAD_COUNT", 10)

	RunSessionInsertThreads(threadCount)
	RunServerInsertThreads(threadCount)
	RunRelayInsertThreads(threadCount)

	// simulate portal activity

	var ready bool

	for i := 0; i < 120; i++ {

		fmt.Printf("iteration %d\n", i)

		sessionCountsResponse := PortalSessionCountsResponse{}

		Get("http://127.0.0.1:50000/portal/session_counts", &sessionCountsResponse)

		fmt.Printf("-------------------------------------------------------------\n")

		fmt.Printf("next sessions = %d, total sessions = %d\n", sessionCountsResponse.NextSessionCount, sessionCountsResponse.TotalSessionCount)

		sessionsResponse := PortalSessionsResponse{}

		Get("http://127.0.0.1:50000/portal/sessions/0", &sessionsResponse)

		fmt.Printf("got data for %d sessions\n", len(sessionsResponse.Sessions))

		sessionDataResponse := PortalSessionDataResponse{}

		if len(sessionsResponse.Sessions) > 0 {

			sessionId, err := strconv.ParseUint(sessionsResponse.Sessions[0].SessionId, 10, 64)
			if err != nil {
				panic(err)
			}

			fmt.Printf("first session id is %016x\n", sessionId)

			Get(fmt.Sprintf("http://127.0.0.1:50000/portal/session/%016x", sessionId), &sessionDataResponse)

			fmt.Printf("session %016x has %d slices, %d client relay data, %d server relay data\n", sessionId, len(sessionDataResponse.SliceData), len(sessionDataResponse.ClientRelayData), len(sessionDataResponse.ServerRelayData))
		}

		serversResponse := PortalServersResponse{}

		Get("http://127.0.0.1:50000/portal/servers/0", &serversResponse)

		fmt.Printf("got data for %d servers\n", len(serversResponse.Servers))

		serverDataResponse := PortalServerDataResponse{}

		if len(serversResponse.Servers) > 0 {

			fmt.Printf("first server id is %016x\n", serversResponse.Servers[0].ServerId)

			Get(fmt.Sprintf("http://127.0.0.1:50000/portal/server/%016x", serversResponse.Servers[0].ServerId), &serverDataResponse)

			fmt.Printf("server %016x has %d sessions\n", serversResponse.Servers[0].ServerId, len(serverDataResponse.ServerSessionIds))
		}

		relayCountResponse := PortalRelayCountResponse{}
		Get("http://127.0.0.1:50000/portal/relay_count", &relayCountResponse)

		fmt.Printf("relays = %d\n", relayCountResponse.RelayCount)

		relaysResponse := PortalRelaysResponse{}

		Get("http://127.0.0.1:50000/portal/relays/0", &relaysResponse)

		fmt.Printf("got data for %d relays\n", len(relaysResponse.Relays))

		relayDataResponse := PortalRelayDataResponse{}

		if len(relaysResponse.Relays) > 0 {

			fmt.Printf("first relay is '%s'\n", relaysResponse.Relays[0].RelayName)

			Get(fmt.Sprintf("http://127.0.0.1:50000/portal/relay/%s", relaysResponse.Relays[0].RelayName), &relayDataResponse)
		}

		mapData := GetBinary("http://127.0.0.1:50000/portal/map_data")

		ready = true

		if len(sessionsResponse.Sessions) < 10 {
			fmt.Printf("A\n")
			ready = false
		}

		if sessionDataResponse.SessionData == nil {
			fmt.Printf("B\n")
			ready = false
		}

		if len(sessionDataResponse.SliceData) == 0 {
			fmt.Printf("C\n")
			ready = false
		}

		if len(sessionDataResponse.ClientRelayData) == 0 {
			fmt.Printf("D\n")
			ready = false
		}

		if len(sessionDataResponse.ServerRelayData) == 0 {
			fmt.Printf("E\n")
			ready = false
		}

		if len(serverDataResponse.ServerSessionIds) < 10 {
			fmt.Printf("F\n")
			ready = false
		}

		if relayCountResponse.RelayCount < 10 {
			fmt.Printf("G\n")
			ready = false
		}

		if len(mapData) == 0 {
			fmt.Printf("H\n")
			ready = false
		}

		fmt.Printf("-------------------------------------------------------------\n")

		if ready {
			break
		}

		time.Sleep(time.Second)
	}

	api_cmd.Process.Signal(os.Interrupt)
	session_cruncher_cmd.Process.Signal(os.Interrupt)
	server_cruncher_cmd.Process.Signal(os.Interrupt)

	api_cmd.Wait()
	session_cruncher_cmd.Wait()
	server_cruncher_cmd.Wait()

	if !ready {
		fmt.Printf("error: portal API is broken\n")
		os.Exit(1)
	}
}

func main() {
	test_portal()
}
