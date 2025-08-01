package main

import (
	"bufio"
	"context"
	"encoding/binary"
	"fmt"
	"math/rand"
	"net"
	"os"
	"os/exec"
	"strings"
	"sync"
	"time"

	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/constants"
	"github.com/networknext/next/modules/core"
	"github.com/networknext/next/modules/envvar"
	"github.com/networknext/next/modules/packets"
)

var service *common.Service
var numServers int
var numRelays int
var serverAddress string
var serverBackendAddress net.UDPAddr
var buyerId uint64
var buyerPrivateKey []byte

func main() {

	service = common.CreateService("load_test_servers")

	numServers = envvar.GetInt("NUM_SERVERS", 1000)

	numRelays = envvar.GetInt("NUM_RELAYS", 1000)

	serverAddress = envvar.GetString("SERVER_ADDRESS", "127.0.0.1")

	serverBackendAddress = envvar.GetAddress("SERVER_BACKEND_ADDRESS", core.ParseAddress("127.0.0.1:40000"))

	buyerPrivateKey = envvar.GetBase64("NEXT_BUYER_PRIVATE_KEY", nil)

	core.Log("num relays = %d", numRelays)
	core.Log("num servers = %d", numServers)
	core.Log("server address = %s", serverAddress)
	core.Log("server backend address = %s", serverBackendAddress.String())

	if buyerPrivateKey == nil {
		panic("you must supply the buyer private key")
	}

	serverAddress = DetectGoogleServerAddress(serverAddress)

	buyerId = binary.LittleEndian.Uint64(buyerPrivateKey[0:8])

	buyerPrivateKey = buyerPrivateKey[8:]

	core.Log("simulating %d servers", numServers)

	go SimulateServers()

	service.WaitForShutdown()
}

func RunCommand(command string, args []string) (bool, string) {

	cmd := exec.Command(command, args...)

	stdoutReader, err := cmd.StdoutPipe()
	if err != nil {
		return false, ""
	}

	var wait sync.WaitGroup
	var mutex sync.Mutex

	output := ""

	stdoutScanner := bufio.NewScanner(stdoutReader)
	wait.Add(1)
	go func() {
		for stdoutScanner.Scan() {
			mutex.Lock()
			output += stdoutScanner.Text() + "\n"
			mutex.Unlock()
		}
		wait.Done()
	}()

	cmd.Stderr = os.Stderr

	err = cmd.Start()
	if err != nil {
		return false, output
	}

	wait.Wait()

	err = cmd.Wait()
	if err != nil {
		return false, output
	}

	return true, output
}

func Bash(command string) (bool, string) {
	return RunCommand("bash", []string{"-c", command})
}

func DetectGoogleServerAddress(input string) string {
	result, output := Bash("curl -s http://metadata/computeMetadata/v1/instance/network-interfaces/0/access-configs/0/external-ip -H \"Metadata-Flavor: Google\" --max-time 10 -s 2>/dev/null")
	if !result {
		return input
	}
	output = strings.TrimSuffix(output, "\n")
	core.Log("google cloud server address is '%s'", output)
	return output
}

func SimulateServers() {
	for i := 0; i < numServers; i++ {
		go RunServer(i)
	}
}

func RunServer(index int) {

	var r = rand.New(rand.NewSource(time.Now().UnixNano()))

	time.Sleep(time.Duration(r.Intn(1000)) * time.Millisecond) // jitter delay

	time.Sleep(time.Duration(r.Intn(360)) * time.Second) // initial delay

	address := core.ParseAddress(fmt.Sprintf("%s:%d", serverAddress, 10000+index))

	var requestId uint64

	datacenterId := common.DatacenterId(fmt.Sprintf("test.%03d", index%numRelays))

	bindAddress := fmt.Sprintf("0.0.0.0:%d", 10000+index)

	lc := net.ListenConfig{}
	lp, err := lc.ListenPacket(context.Background(), "udp", bindAddress)
	if err != nil {
		panic(fmt.Sprintf("could not bind socket: %v", err))
	}

	conn := lp.(*net.UDPConn)

	go func() {

		for {

			buffer := make([]byte, constants.MaxPacketBytes)
			_, _, err := conn.ReadFromUDP(buffer[:])
			if err != nil {
				fmt.Printf("udp receive error: %v\n", err)
				break
			}
		}

		conn.Close()

	}()

	ticker := time.NewTicker(10 * time.Second)

	go func() {
		for {
			select {

			case <-service.Context.Done():
				return

			case <-ticker.C:

				fmt.Printf("update server %03d\n", index)

				packet := packets.SDK_ServerUpdateRequestPacket{
					Version:       packets.SDKVersion{255, 255, 255},
					BuyerId:       buyerId,
					RequestId:     requestId,
					DatacenterId:  datacenterId,
					NumSessions:   uint32(common.RandomInt(100, 200)),
					ServerAddress: address,
					Uptime:        uint64(requestId * 10),
				}

				requestId += 1

				packetData, err := packets.SDK_WritePacket(&packet, packets.SDK_SERVER_UPDATE_REQUEST_PACKET, constants.MaxPacketBytes, &address, &serverBackendAddress, buyerPrivateKey)
				if err != nil {
					core.Error("failed to write response packet: %v", err)
					return
				}

				if _, err := conn.WriteToUDP(packetData, &serverBackendAddress); err != nil {
					core.Error("failed to send packet: %v", err)
					return
				}
			}
		}
	}()

}
