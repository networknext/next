/*
   Network Next. Copyright 2017 - 2026 Network Next, Inc.
   Licensed under the Network Next Source Available License 1.0
*/

package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/exec"
	"reflect"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/constants"
	"github.com/networknext/next/modules/core"
	"github.com/networknext/next/modules/crypto"
	db "github.com/networknext/next/modules/database"
	"github.com/networknext/next/modules/encoding"
	"github.com/networknext/next/modules/packets"
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

func Base64String(value string) []byte {
	data, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		panic(err)
	}
	return data
}

const TestRelayPublicKey = "1nTj7bQmo8gfIDqG+o//GFsak/g1TRo4hl6XXw1JkyI="
const TestRelayPrivateKey = "cwvK44Pr5aHI3vE3siODS7CUgdPI/l1VwjVZ2FvEyAo="
const TestRelayBackendPublicKey = "IsjRpWEz9H7qslhWWupW4A9LIpVh+PzWoLleuXL1NUE="
const TestRelayBackendPrivateKey = "qXeUdLPZxaMnZ/zFHLHkmgkQOmunWq1AmRv55nqTYMg="
const TestServerBackendPublicKey = "1wXeogqOEL/UuMnHy3lwpdkdklcg4IktO/39mJiYfgc="
const TestServerBackendPrivateKey = "peZ17P29VgtnOiEv5wwNPDDo9lWweFV7dBVac0KoaXHXBd6iCo4Qv9S4ycfLeXCl2R2SVyDgiS07/f2YmJh+Bw=="
const TestPingKey = "xsBL4b6PO4ESADcc69kERzLXxs9ESOrX1kSHJH0m9D0="

func check_output(substring string, cmd *exec.Cmd, stdout bytes.Buffer, stderr bytes.Buffer) {
	if !strings.Contains(stdout.String(), substring) {
		fmt.Printf("\nerror: missing output '%s'\n\n", substring)
		fmt.Printf("--------------------------------------------------\n")
		fmt.Printf("%s", stdout.String())
		fmt.Printf("--------------------------------------------------\n")
		if len(stderr.String()) > 0 {
			fmt.Printf("%s", stderr.String())
			fmt.Printf("--------------------------------------------------\n")
		}
		fmt.Printf("\n")
		cmd.Process.Signal(syscall.SIGTERM)
		os.Exit(1)
	}
}

func test_magic_backend() {

	fmt.Printf("test_magic_backend\n")

	// run the magic backend and make sure it runs and does things it's expected to do

	cmd := exec.Command("./magic_backend")

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	cmd.Env = make([]string, 0)
	cmd.Env = append(cmd.Env, "ENV=local")
	cmd.Env = append(cmd.Env, "DEBUG_LOGS=1")
	cmd.Env = append(cmd.Env, "HTTP_PORT=40000")
	cmd.Env = append(cmd.Env, "MAGIC_UPDATE_SECONDS=5")

	err := cmd.Start()
	if err != nil {
		fmt.Printf("\nerror: failed to run magic backend!\n\n")
		fmt.Printf("%s", stdout.String())
		fmt.Printf("%s", stderr.String())
		os.Exit(1)
	}

	time.Sleep(20 * time.Second)

	check_output("magic_backend", cmd, stdout, stderr)
	check_output("starting http server on port 40000", cmd, stdout, stderr)

	// test the vm health check

	response, err := http.Get("http://127.0.0.1:40000/vm_health")
	if err != nil || response.StatusCode != 200 {
		fmt.Printf("error: vm health check failed\n")
		cmd.Process.Signal(syscall.SIGTERM)
		os.Exit(1)
	}

	// test the lb health check

	response, err = http.Get("http://127.0.0.1:40000/lb_health")
	if err != nil || response.StatusCode != 200 {
		fmt.Printf("error: lb health check failed\n")
		cmd.Process.Signal(syscall.SIGTERM)
		os.Exit(1)
	}

	// test the version endpoint

	_, err = http.Get("http://127.0.0.1:40000/version")
	if err != nil || response.StatusCode != 200 {
		fmt.Printf("error: version endpoint failed\n")
		cmd.Process.Signal(syscall.SIGTERM)
		os.Exit(1)
	}

	// test the magic values endpoint

	response, err = http.Get("http://127.0.0.1:40000/magic")
	if err != nil || response.StatusCode != 200 {
		fmt.Printf("error: magic endpoint failed\n")
		cmd.Process.Signal(syscall.SIGTERM)
		os.Exit(1)
	}

	magicData, error := io.ReadAll(response.Body)
	if error != nil {
		fmt.Printf("error: failed to read magic response data\n")
		cmd.Process.Signal(syscall.SIGTERM)
		os.Exit(1)
	}

	if len(magicData) != 32 {
		fmt.Printf("error: magic data should be 32 bytes long\n")
		cmd.Process.Signal(syscall.SIGTERM)
		os.Exit(1)
	}

	time.Sleep(time.Second)

	check_output("served magic values", cmd, stdout, stderr)

	// test that the magic values shuffle from upcoming -> current -> previous over time

	magicCounter := binary.LittleEndian.Uint64(magicData[0:8])

	var upcomingMagic [8]byte
	var currentMagic [8]byte
	var previousMagic [8]byte

	copy(upcomingMagic[:], magicData[8:16])
	copy(currentMagic[:], magicData[16:24])
	copy(previousMagic[:], magicData[24:32])

	magicUpdates := 0

	for i := 0; i < 30; i++ {

		response, err = http.Get("http://127.0.0.1:40000/magic")
		if err != nil || response.StatusCode != 200 {
			fmt.Printf("error: magic endpoint failed\n")
			cmd.Process.Signal(syscall.SIGTERM)
			os.Exit(1)
		}

		magicData, error := io.ReadAll(response.Body)
		if error != nil {
			fmt.Printf("error: failed to read magic response data\n")
			cmd.Process.Signal(syscall.SIGTERM)
			os.Exit(1)
		}

		if len(magicData) != 32 {
			fmt.Printf("error: magic data should be 32 bytes long\n")
			cmd.Process.Signal(syscall.SIGTERM)
			os.Exit(1)
		}

		newMagicCounter := binary.LittleEndian.Uint64(magicData[0:8])
		if newMagicCounter < magicCounter {
			fmt.Printf("error: magic counter must not decrease\n")
			cmd.Process.Signal(syscall.SIGTERM)
			os.Exit(1)
		}

		if newMagicCounter != magicCounter {

			magicCounter = newMagicCounter
			magicUpdates++

			if bytes.Compare(magicData[16:24], upcomingMagic[:]) != 0 {
				fmt.Printf("error: did not see upcoming magic shuffle to current magic\n")
				cmd.Process.Signal(syscall.SIGTERM)
				os.Exit(1)
			}

			if bytes.Compare(magicData[24:32], currentMagic[:]) != 0 {
				fmt.Printf("error: did not see current magic shuffle to previous magic\n")
				cmd.Process.Signal(syscall.SIGTERM)
				os.Exit(1)
			}

			copy(upcomingMagic[:], magicData[8:16])
			copy(currentMagic[:], magicData[16:24])
			copy(previousMagic[:], magicData[24:32])

		}

		time.Sleep(time.Second)

	}

	// we should see 5,6 or 7 magic updates (30 seconds with updates once every 5 seconds...)

	if magicUpdates != 5 && magicUpdates != 6 && magicUpdates != 7 {
		fmt.Printf("error: did not see magic values update every ~5 seconds (%d magic updates)", magicUpdates)
		cmd.Process.Signal(syscall.SIGTERM)
		os.Exit(1)
	}

	// run a second magic backend. it should match the same magic values

	cmd2 := exec.Command("./magic_backend")

	var stdout2 bytes.Buffer
	var stderr2 bytes.Buffer
	cmd2.Stdout = &stdout2
	cmd2.Stderr = &stderr2

	cmd2.Env = make([]string, 0)
	cmd2.Env = append(cmd2.Env, "ENV=local")
	cmd2.Env = append(cmd2.Env, "DEBUG_LOGS=1")
	cmd2.Env = append(cmd2.Env, "HTTP_PORT=40001")
	cmd2.Env = append(cmd2.Env, "MAGIC_UPDATE_SECONDS=5")

	err = cmd2.Start()
	if err != nil {
		fmt.Printf("\nerror: failed to run magic backend #2!\n\n")
		fmt.Printf("%s", stdout2.String())
		fmt.Printf("%s", stderr2.String())
		os.Exit(1)
	}

	time.Sleep(time.Second)

	for i := 0; i < 10; i++ {

		response1, err := http.Get("http://127.0.0.1:40000/magic")
		if err != nil || response1.StatusCode != 200 {
			fmt.Printf("error: magic endpoint failed (1)\n")
			cmd.Process.Signal(syscall.SIGTERM)
			os.Exit(1)
		}

		magicData1, error := io.ReadAll(response.Body)
		if error != nil {
			fmt.Printf("error: failed to read magic response data (1)\n")
			cmd.Process.Signal(syscall.SIGTERM)
			os.Exit(1)
		}

		response2, err := http.Get("http://127.0.0.1:40001/magic")
		if err != nil || response2.StatusCode != 200 {
			fmt.Printf("error: magic endpoint failed (2)\n")
			cmd.Process.Signal(syscall.SIGTERM)
			os.Exit(1)
		}

		magicData2, error := io.ReadAll(response.Body)
		if error != nil {
			fmt.Printf("error: failed to read magic response data (2)\n")
			cmd.Process.Signal(syscall.SIGTERM)
			os.Exit(1)
		}

		if bytes.Compare(magicData1, magicData2) != 0 && !(bytes.Compare(magicData1[8:24], magicData2[16:32]) == 0 || bytes.Compare(magicData2[8:24], magicData1[16:32]) == 0) {
			fmt.Printf("error: magic data mismatch between two magic backends\n")
			cmd.Process.Signal(syscall.SIGTERM)
			os.Exit(1)
		}

		time.Sleep(time.Second)

	}

	// test that the service shuts down cleanly

	cmd.Process.Signal(os.Interrupt)
	cmd2.Process.Signal(os.Interrupt)

	cmd.Wait()
	cmd2.Wait()

	check_output("received shutdown signal", cmd, stdout, stderr)
	check_output("successfully shutdown", cmd, stdout, stderr)

	check_output("received shutdown signal", cmd2, stdout, stderr)
	check_output("successfully shutdown", cmd2, stdout, stderr)
}

func test_cost_matrix_read_write() {

	fmt.Printf("test_cost_matrix_read_write\n")

	for numRelays := 0; numRelays <= constants.MaxRelays; numRelays++ {

		fmt.Printf("read/write cost matrix with %d relays\n", numRelays)

		writeMessage := common.GenerateRandomCostMatrix(numRelays)

		readMessage := common.CostMatrix{}

		buffer, err := writeMessage.Write()
		if err != nil {
			panic(err)
		}

		err = readMessage.Read(buffer)

		if !reflect.DeepEqual(writeMessage, readMessage) {
			panic("cost matrix read write failure")
		}
	}
}

func test_route_matrix_read_write() {

	fmt.Printf("test_route_matrix_read_write\n")

	const step = 21

	for numRelays := 0; numRelays <= constants.MaxRelays+step-1; numRelays += step {

		if numRelays > constants.MaxRelays {
			numRelays = constants.MaxRelays
		}

		fmt.Printf("read/write route matrix with %d relays\n", numRelays)

		writeMessage := common.GenerateRandomRouteMatrix(numRelays)

		readMessage := common.RouteMatrix{}

		buffer, err := writeMessage.Write()
		if err != nil {
			panic(err)
		}

		err = readMessage.Read(buffer)

		if !reflect.DeepEqual(writeMessage, readMessage) {
			panic("route matrix read write failure")
		}
	}
}

func test_session_data_serialize() {

	fmt.Printf("test_session_data_serialize\n")

	startTime := time.Now()

	for {

		if time.Since(startTime) > 60*time.Second {
			break
		}

		writePacket := packets.GenerateRandomSessionData()

		readPacket := packets.SDK_SessionData{}

		const BufferSize = 10 * 1024

		buffer := [BufferSize]byte{}

		writeStream := encoding.CreateWriteStream(buffer[:])

		err := writePacket.Serialize(writeStream)
		if err != nil {
			panic(err)
		}
		writeStream.Flush()
		packetBytes := writeStream.GetBytesProcessed()

		readStream := encoding.CreateReadStream(buffer[:packetBytes])
		err = readPacket.Serialize(readStream)
		if err != nil {
			panic(err)
		}

		if !reflect.DeepEqual(writePacket, readPacket) {
			panic("session data serialize failure")
		}
	}
}

func test_relay_manager() {

	fmt.Printf("test_relay_manager\n")

	relayManager := common.CreateRelayManager(true)

	ctx, contextCancelFunc := context.WithCancel(context.Background())

	// setup a lot of relays

	const NumRelays = constants.MaxRelays

	relayNames := make([]string, NumRelays)
	relayIds := make([]uint64, NumRelays)
	relayAddresses := make([]net.UDPAddr, NumRelays)

	for i := range relayIds {
		relayIds[i] = common.RelayId(relayNames[i])
		relayAddresses[i] = core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", 2000+i))
	}

	// get costs once per-second

	go func() {
		counter := 0
		ticker := time.NewTicker(time.Second)
		for {
			select {
			case <-ctx.Done():
				if counter < 30 {
					panic("get costs deadlocked!")
				}
				return
			case <-ticker.C:
				fmt.Printf("costs %d\n", counter)
				const MaxJitter = 100
				const MaxPacketLoss = 1
				currentTime := time.Now().Unix()
				costs := relayManager.GetCosts(currentTime, relayIds, MaxJitter, MaxPacketLoss)
				_ = costs
				counter++
			}
		}
	}()

	// really slam in the relay updates once per-second, randomly for 1000 relays

	numSamples := NumRelays
	sampleRelayId := make([]uint64, numSamples)
	sampleRTT := make([]uint8, numSamples)
	sampleJitter := make([]uint8, numSamples)
	samplePacketLoss := make([]uint16, numSamples)
	counters := make([]uint64, constants.NumRelayCounters)

	for i := 0; i < numSamples; i++ {
		sampleRelayId[i] = uint64(i)
		sampleRTT[i] = 10
		sampleJitter[i] = 5
		samplePacketLoss[i] = 0
	}

	for i := 0; i < NumRelays; i++ {

		go func(index int) {

			ticker := time.NewTicker(time.Second)
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					currentTime := time.Now().Unix()
					fmt.Printf("relay update\n")
					relayManager.ProcessRelayUpdate(currentTime, relayIds[index], relayNames[index], relayAddresses[index], 0, "test", 0, numSamples, sampleRelayId, sampleRTT, sampleJitter, samplePacketLoss, counters)
				}
			}

		}(i)

	}

	time.Sleep(20 * time.Second)

	contextCancelFunc()
}

func test_optimize() {

	fmt.Printf("test_optimize\n")

	relayManager := common.CreateRelayManager(true)

	ctx, contextCancelFunc := context.WithCancel(context.Background())

	// setup a lot of relays

	const NumRelays = 100

	relayNames := make([]string, NumRelays)
	relayIds := make([]uint64, NumRelays)
	relayAddresses := make([]net.UDPAddr, NumRelays)
	relayLatitudes := make([]float32, NumRelays)
	relayLongitudes := make([]float32, NumRelays)
	relayDatacenterIds := make([]uint64, NumRelays)
	destRelays := make([]bool, NumRelays)

	for i := range relayIds {
		relayNames[i] = fmt.Sprintf("relay%d", i)
		relayIds[i] = common.RelayId(relayNames[i])
		relayAddresses[i] = core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", 2000+i))
		relayLatitudes[i] = float32(common.RandomInt(-90, +90))
		relayLongitudes[i] = float32(common.RandomInt(-90, +90))
		relayDatacenterIds[i] = uint64(common.RandomInt(0, 5))
		destRelays[i] = true
	}

	// get costs once per-second

	go func() {
		counter := 0
		ticker := time.NewTicker(time.Second)
		for {
			select {

			case <-ctx.Done():
				if counter < 30 {
					panic("optimize deadlocked!")
				}
				return

			case <-ticker.C:

				const MaxJitter = 100
				const MaxPacketLoss = 1

				currentTime := time.Now().Unix()

				relayPrice := make([]uint8, NumRelays)

				costs := relayManager.GetCosts(currentTime, relayIds, MaxJitter, MaxPacketLoss)

				costMatrix := &common.CostMatrix{
					Version:            common.CostMatrixVersion_Write,
					RelayIds:           relayIds,
					RelayAddresses:     relayAddresses,
					RelayNames:         relayNames,
					RelayLatitudes:     relayLatitudes,
					RelayLongitudes:    relayLongitudes,
					RelayDatacenterIds: relayDatacenterIds,
					DestRelays:         destRelays,
					Costs:              costs,
					RelayPrice:         relayPrice,
				}

				costMatrixData, err := costMatrix.Write()
				if err != nil {
					panic("could not write cost matrix")
				}
				_ = costMatrixData

				numCPUs := runtime.NumCPU()
				numSegments := NumRelays
				if numCPUs < NumRelays {
					numSegments = NumRelays / 5
					if numSegments == 0 {
						numSegments = 1
					}
				}

				binFileData := make([]byte, 256*1024)

				routeMatrix := &common.RouteMatrix{
					CreatedAt:          uint64(time.Now().Unix()),
					Version:            common.RouteMatrixVersion_Write,
					RelayIds:           relayIds,
					RelayAddresses:     relayAddresses,
					RelayNames:         relayNames,
					RelayLatitudes:     relayLatitudes,
					RelayLongitudes:    relayLongitudes,
					RelayDatacenterIds: relayDatacenterIds,
					DestRelays:         destRelays,
					RouteEntries:       core.Optimize2(NumRelays, numSegments, costs, relayPrice, relayDatacenterIds, destRelays),
					BinFileBytes:       int32(len(binFileData)),
					BinFileData:        binFileData,
					Costs:              costs,
					RelayPrice:         relayPrice,
				}

				routeMatrixData, err := routeMatrix.Write()
				if err != nil {
					panic(fmt.Sprintf("could not write route matrix: %v", err))
					continue
				}
				_ = routeMatrixData

				fmt.Printf("optimize %d\n", counter)

				counter++
			}
		}
	}()

	// relay updates once per-second for each relay

	numSamples := NumRelays
	sampleRelayId := make([]uint64, numSamples)
	sampleRTT := make([]uint8, numSamples)
	sampleJitter := make([]uint8, numSamples)
	samplePacketLoss := make([]uint16, numSamples)
	counters := make([]uint64, constants.NumRelayCounters)

	for i := 0; i < numSamples; i++ {
		sampleRelayId[i] = uint64(i)
		sampleRTT[i] = uint8(common.RandomInt(0, 255))
		sampleJitter[i] = uint8(common.RandomInt(0, 255))
		samplePacketLoss[i] = uint16(common.RandomInt(0, 255))
	}

	for i := 0; i < NumRelays; i++ {

		go func(index int) {

			ticker := time.NewTicker(time.Second)
			for {
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
					currentTime := time.Now().Unix()
					relayManager.ProcessRelayUpdate(currentTime, relayIds[index], relayNames[index], relayAddresses[index], 0, "test", 0, numSamples, sampleRelayId, sampleRTT, sampleJitter, samplePacketLoss, counters)
				}
			}

		}(i)

	}

	time.Sleep(60 * time.Second)

	contextCancelFunc()
}

const (
	magicBackendBin = "./magic_backend"
	relayGatewayBin = "./relay_gateway"
	relayBackendBin = "./relay_backend"
)

func test_relay_backend() {

	fmt.Printf("test_relay_backend\n")

	cancelContext, cancelFunc := context.WithTimeout(context.Background(), time.Duration(60*time.Second))

	// setup datacenters

	const NumDatacenters = 10

	datacenterIds := make([]uint64, NumDatacenters)
	datacenterNames := make([]string, NumDatacenters)
	datacenterLatitudes := make([]float32, NumDatacenters)
	datacenterLongitudes := make([]float32, NumDatacenters)

	for i := 0; i < NumDatacenters; i++ {
		datacenterIds[i] = uint64(i)
		datacenterNames[i] = fmt.Sprintf("datacenter%d", i)
		datacenterLatitudes[i] = float32(common.RandomInt(-90, +90))
		datacenterLongitudes[i] = float32(common.RandomInt(-90, +90))
	}

	// setup relays

	const NumRelays = 100

	relayIds := make([]uint64, NumRelays)
	relayNames := make([]string, NumRelays)
	relayAddresses := make([]net.UDPAddr, NumRelays)
	relayDatacenterIds := make([]uint64, NumRelays)
	destRelays := make([]bool, NumRelays)

	for i := range relayIds {
		relayNames[i] = fmt.Sprintf("relay%d", i)
		relayAddresses[i] = core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", 2000+i))
		relayIds[i] = common.RelayId(relayAddresses[i].String())
		relayDatacenterIds[i] = uint64(common.RandomInt(0, NumDatacenters-1))
		destRelays[i] = true
	}

	// setup a database containing the relays

	database := db.CreateDatabase()

	database.CreationTime = time.Now().String()
	database.Creator = "test"

	seller := db.Seller{}
	seller.Id = 1
	seller.Name = "seller"
	database.SellerMap[1] = &seller

	datacenter := db.Datacenter{}
	datacenter.Id = 1
	datacenter.Name = "test"
	database.DatacenterMap[1] = &datacenter

	for i := 0; i < NumRelays; i++ {

		relay := db.Relay{}

		relay.Id = relayIds[i]
		relay.Name = relayNames[i]
		relay.PublicAddress = relayAddresses[i]
		relay.SSHAddress = relayAddresses[i]
		relay.Version = "test"
		relay.Seller = &seller
		relay.Datacenter = &datacenter
		relay.PublicKey = Base64String(TestRelayPublicKey)

		database.Relays = append(database.Relays, relay)

		database.DatacenterRelays[datacenter.Id] = append(database.DatacenterRelays[datacenter.Id], relay.Id)
	}

	database.Fixup()

	// write the database out to a temporary file

	file, err := os.CreateTemp(".", "temp-database-")
	if err != nil {
		panic("could not create temporary database file")
	}

	databaseFilename := file.Name()

	defer os.Remove(databaseFilename)

	fmt.Println(databaseFilename)

	err = database.Validate()
	if err != nil {
		fmt.Printf("error: database did not validate: %v\n", err)
		os.Exit(1)
	}

	database.Save(databaseFilename)

	// start the magic backend

	fmt.Printf("starting magic backend\n")

	magic_backend_cmd := exec.Command(magicBackendBin)
	if magic_backend_cmd == nil {
		panic("could not create magic backend!\n")
	}

	magic_backend_cmd.Env = os.Environ()
	magic_backend_cmd.Env = append(magic_backend_cmd.Env, "HTTP_PORT=41007")

	var magic_backend_output bytes.Buffer
	magic_backend_cmd.Stdout = &magic_backend_output
	magic_backend_cmd.Stderr = &magic_backend_output
	magic_backend_cmd.Start()

	// run the relay gateway, such that it loads the temporary database file

	fmt.Printf("starting relay gateway\n")

	relay_gateway_cmd := exec.Command(relayGatewayBin)
	if relay_gateway_cmd == nil {
		panic("could not create relay gateway!\n")
	}

	relay_gateway_cmd.Env = os.Environ()
	relay_gateway_cmd.Env = append(relay_gateway_cmd.Env, fmt.Sprintf("DATABASE_PATH=%s", databaseFilename))
	relay_gateway_cmd.Env = append(relay_gateway_cmd.Env, "HTTP_PORT=30000")
	relay_gateway_cmd.Env = append(relay_gateway_cmd.Env, fmt.Sprintf("RELAY_BACKEND_PUBLIC_KEY=%s", TestRelayBackendPublicKey))
	relay_gateway_cmd.Env = append(relay_gateway_cmd.Env, fmt.Sprintf("RELAY_BACKEND_PRIVATE_KEY=%s", TestRelayBackendPrivateKey))
	relay_gateway_cmd.Env = append(relay_gateway_cmd.Env, fmt.Sprintf("PING_KEY=%s", TestPingKey))
	relay_gateway_cmd.Env = append(relay_gateway_cmd.Env, "DEBUG_LOGS=0")

	relay_gateway_cmd.Stdout = os.Stdout
	relay_gateway_cmd.Stderr = os.Stderr

	relay_gateway_cmd.Start()

	// run the relay backend, such that it loads the temporary database file

	fmt.Printf("starting relay backend\n")

	relay_backend_cmd := exec.Command(relayBackendBin)

	if relay_backend_cmd == nil {
		panic("could not create relay backend!\n")
	}

	relay_backend_cmd.Env = os.Environ()
	relay_backend_cmd.Env = append(relay_backend_cmd.Env, fmt.Sprintf("DATABASE_PATH=%s", databaseFilename))
	relay_backend_cmd.Env = append(relay_backend_cmd.Env, "HTTP_PORT=30001")
	relay_backend_cmd.Env = append(relay_backend_cmd.Env, "INITIAL_DELAY=5")
	relay_backend_cmd.Env = append(relay_backend_cmd.Env, fmt.Sprintf("RELAY_BACKEND_PUBLIC_KEY=%s", TestRelayBackendPublicKey))
	relay_backend_cmd.Env = append(relay_backend_cmd.Env, fmt.Sprintf("RELAY_BACKEND_PRIVATE_KEY=%s", TestRelayBackendPrivateKey))
	relay_backend_cmd.Env = append(relay_backend_cmd.Env, "DEBUG_LOGS=0")

	relay_backend_cmd.Stdout = os.Stdout
	relay_backend_cmd.Stderr = os.Stderr

	relay_backend_cmd.Start()

	// hammer the relay backend with relay updates

	var waitGroup sync.WaitGroup

	waitGroup.Add(NumRelays)

	var errorCount uint64

	for i := 0; i < NumRelays; i++ {

		go func(index int) {

			// create http client

			transport := &http.Transport{
				MaxIdleConns:        1,
				MaxIdleConnsPerHost: 1,
			}

			client := &http.Client{Transport: transport}

			ticker := time.NewTicker(1 * time.Second)

			for {
				select {

				case <-cancelContext.Done():
					waitGroup.Done()
					return

				case <-ticker.C:

					requestPacket := packets.RelayUpdateRequestPacket{}

					requestPacket.Version = packets.RelayUpdateRequestPacket_VersionWrite
					requestPacket.CurrentTime = uint64(time.Now().Unix())
					requestPacket.Address = relayAddresses[index]
					requestPacket.NumSamples = NumRelays
					requestPacket.NumRelayCounters = constants.NumRelayCounters

					for i := 0; i < NumRelays; i++ {
						requestPacket.SampleRelayId[i] = relayIds[i]
						requestPacket.SampleRTT[i] = uint8(common.RandomInt(0, 255))
						requestPacket.SampleJitter[i] = uint8(common.RandomInt(0, 255))
						requestPacket.SamplePacketLoss[i] = uint16(common.RandomInt(0, 65535))
					}

					buffer := make([]byte, 100*1024)

					packetData := requestPacket.Write(buffer)
					packetBytes := len(packetData)

					encryptData := packetData[1+1+4+2:]

					nonce := make([]byte, crypto.Box_NonceSize)

					crypto.Box_Encrypt(Base64String(TestRelayPrivateKey), Base64String(TestRelayBackendPublicKey), nonce, encryptData, len(encryptData))

					body := packetData[:packetBytes+crypto.Box_MacSize+crypto.Box_NonceSize]

					bodyLength := len(body)

					copy(body[bodyLength-crypto.Box_NonceSize:], nonce)

					request, err := http.NewRequest("POST", "http://127.0.0.1:30000/relay_update", bytes.NewBuffer(body))
					if err != nil {
						fmt.Printf("error creating http request: %v\n", err)
						atomic.AddUint64(&errorCount, 1)
						break
					}

					request.Header.Set("Content-Type", "application/octet-stream")

					response, err := client.Do(request)
					if err != nil {
						fmt.Printf("error running http request: %v\n", err)
						atomic.AddUint64(&errorCount, 1)
						break
					}

					if response.StatusCode != 200 {
						fmt.Printf("bad http response %d\n", response.StatusCode)
						atomic.AddUint64(&errorCount, 1)
						break
					}

					body, err = io.ReadAll(response.Body)

					if err != nil {
						fmt.Printf("error reading http response: %v\n", err)
						atomic.AddUint64(&errorCount, 1)
						break
					}

					defer response.Body.Close()

					// read the relay response packet

					var responsePacket packets.RelayUpdateResponsePacket

					err = responsePacket.Read(body)
					if err != nil {
						fmt.Printf("could not read relay response: %v", err)
						atomic.AddUint64(&errorCount, 1)
						break
					}

					_ = responsePacket
				}
			}

		}(i)
	}

	// run a goroutine to pull down the route matrix once per-second from the relay backend

	waitGroup.Add(1)

	routeMatrixCounter := 0

	go func() {

		transport := &http.Transport{
			MaxIdleConns:        1,
			MaxIdleConnsPerHost: 1,
		}

		client := &http.Client{Transport: transport}

		// wait until the relay backend health checks both pass

		fmt.Printf("waiting until health checks pass...\n")

		for {

			readyCount := 0

			response, err := client.Get("http://127.0.0.1:30001/vm_health")
			if err == nil && response.StatusCode == 200 {
				readyCount++
			} else {
				fmt.Printf("vm_health is not ready\n")
			}

			response, err = client.Get("http://127.0.0.1:30001/lb_health")
			if err == nil && response.StatusCode == 200 {
				readyCount++
			} else {
				fmt.Printf("lb_health is not ready\n")
			}

			if readyCount == 2 {
				break
			}

			time.Sleep(time.Second)
		}

		// request route matrix once per-second

		fmt.Printf("requesting route matrix once per-second\n")

		ticker := time.NewTicker(1 * time.Second)

		for {
			select {

			case <-cancelContext.Done():
				waitGroup.Done()
				return

			case <-ticker.C:

				response, err := client.Get("http://127.0.0.1:30001/route_matrix")
				if err != nil {
					core.Error("failed to http get route matrix: %v", err)
					atomic.AddUint64(&errorCount, 1)
					break
				}

				buffer, err := io.ReadAll(response.Body)
				if err != nil {
					core.Error("failed to read route matrix: %v", err)
					atomic.AddUint64(&errorCount, 1)
					break
				}

				response.Body.Close()

				routeMatrix := common.RouteMatrix{}

				err = routeMatrix.Read(buffer)
				if err != nil {
					core.Error("failed to read route matrix: %v", err)
					atomic.AddUint64(&errorCount, 1)
					break
				}

				if len(routeMatrix.RelayIds) != NumRelays {
					core.Error("wrong num relays in route matrix: %d", len(routeMatrix.RelayIds))
					atomic.AddUint64(&errorCount, 1)
					break
				}

				fmt.Printf("route matrix %d\n", routeMatrixCounter)

				routeMatrixCounter++
			}
		}
	}()

	// wait for 60 seconds

	time.Sleep(60 * time.Second)

	// wait for all goroutines to finish

	cancelFunc()

	fmt.Printf("waiting for goroutines\n")

	waitGroup.Wait()

	// print output from services

	fmt.Printf("waiting for magic backend\n")

	magic_backend_cmd.Process.Signal(os.Interrupt)
	magic_backend_cmd.Wait()

	fmt.Printf("waiting for relay gateway\n")

	relay_gateway_cmd.Process.Signal(os.Interrupt)
	relay_gateway_cmd.Wait()

	fmt.Printf("waiting for relay backend\n")

	relay_backend_cmd.Process.Signal(os.Kill)
	relay_backend_cmd.Wait()

	if errorCount != 0 {
		panic("error count is not zero")
	}

	if routeMatrixCounter < 45 {
		panic("not enough valid route matrices")
	}
}

type test_function func()

var googleProjectID string

func main() {

	googleProjectID = "local"

	allTests := []test_function{
		test_magic_backend,
		test_cost_matrix_read_write,
		test_route_matrix_read_write,
		test_session_data_serialize,
		test_relay_manager,
		test_optimize,
		test_relay_backend,
	}

	var tests []test_function

	if len(os.Args) > 1 {
		funcName := os.Args[1]
		for _, test := range allTests {
			name := runtime.FuncForPC(reflect.ValueOf(test).Pointer()).Name()
			name = name[len("main."):]
			if funcName == name {
				tests = append(tests, test)
				break
			}
		}
		if len(tests) == 0 {
			panic(fmt.Sprintf("could not find any test: '%s'", funcName))
		}
	} else {
		tests = allTests // No command line args, run all tests
	}

	go func() {
		time.Sleep(time.Duration(len(tests)*120) * time.Second)
		panic("tests took too long!")
	}()

	fmt.Printf("\n")

	for i := range tests {
		tests[i]()
	}
}
