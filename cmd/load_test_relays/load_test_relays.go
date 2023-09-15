package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
	"os"
	"os/exec"
	"sync"
	"bufio"
	"strings"

	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/constants"
	"github.com/networknext/next/modules/core"
	"github.com/networknext/next/modules/envvar"
	"github.com/networknext/next/modules/packets"
	"github.com/networknext/next/modules/crypto"
)

var service *common.Service
var relayAddress string
var relayBackendHostname string
var relayBackendPublicKey []byte
var relayPrivateKey []byte
var numRelays int

func main() {

	service = common.CreateService("load_test_relays")

	numRelays = envvar.GetInt("NUM_RELAYS", 1000)
	
	relayAddress = envvar.GetString("RELAY_ADDRESS", "127.0.0.1")

	relayBackendHostname = envvar.GetString("RELAY_BACKEND_HOSTNAME", "http://127.0.0.1:30000")

	relayBackendPublicKey = envvar.GetBase64("RELAY_BACKEND_PUBLIC_KEY", []byte{})

	if len(relayBackendPublicKey) == 0 {
		panic("you must supply the relay backend public key")
	}

	relayPrivateKey = envvar.GetBase64("RELAY_PRIVATE_KEY", []byte{})

	if len(relayPrivateKey) == 0 {
		panic("you must supply the relay private key")
	}

	core.Log("simulating %d relays", numRelays)

	relayAddress = DetectGoogleRelayAddress(relayAddress)

	go SimulateRelays(service)

	service.WaitForShutdown()
}

func SimulateRelays(service *common.Service) {
	for i := 0; i < numRelays; i++ {
		go RunRelay(service, i)
	}
}

func RunRelay(service *common.Service, index int) {

	time.Sleep(time.Duration(common.RandomInt(0, 1000)) * time.Millisecond)

	startTime := time.Now().Unix()

	address := core.ParseAddress(fmt.Sprintf("%s:%d", relayAddress, 10000+index))

	sampleRelayIds := make([]uint64, numRelays)
	for i := 0; i < numRelays; i++ {
		sampleRelayIds[i] = common.RelayId(fmt.Sprintf("%s:%d", relayAddress, 10000+i))
	}

	ticker := time.NewTicker(time.Second)

	go func() {
		for {
			select {

			case <-service.Context.Done():
				return

			case <-ticker.C:

				fmt.Printf("update relay %03d\n", index)

				// construct relay update. it has random samples for all the other relays which should result in a worse case route matrix optimize

				packet := packets.RelayUpdateRequestPacket{
					Version:                   uint8(packets.RelayUpdateRequestPacket_VersionMax),
					CurrentTime:               uint64(time.Now().Unix()),
					StartTime:                 uint64(startTime),
					Address:                   address,
					NumSamples:                uint32(numRelays),
					SessionCount:              100,
					EnvelopeBandwidthUpKbps:   uint32(common.RandomInt(10000, 20000)),
					EnvelopeBandwidthDownKbps: uint32(common.RandomInt(10000, 20000)),
					PacketsSentPerSecond:      float32(common.RandomInt(1000, 2000)),
					PacketsReceivedPerSecond:  float32(common.RandomInt(1000, 2000)),
					BandwidthSentKbps:         float32(common.RandomInt(10000, 20000)),
					BandwidthReceivedKbps:     float32(common.RandomInt(10000, 20000)),
					NearPingsPerSecond:        float32(common.RandomInt(10000, 20000)),
					RelayPingsPerSecond:       float32(common.RandomInt(10000, 20000)),
					RelayFlags:                0,
					RelayVersion:              "load test",
					NumRelayCounters:          constants.NumRelayCounters,
				}

				copy(packet.SampleRelayId[:], sampleRelayIds)

				for i := 0; i < int(packet.NumSamples); i++ {
					packet.SampleRTT[i] = uint8(common.RandomInt(1, 10))
				}

				// write relay update packet

				const BufferSize = 16 * 1024

				var buffer [BufferSize]byte

				packetData := packet.Write(buffer[:])

				// encrypt relay update

				nonce := make([]byte, crypto.Box_NonceSize)

				common.RandomBytes(nonce)

				encryptedData := buffer[8:len(packetData)]

				encryptedBytes := crypto.Box_Encrypt(relayPrivateKey[:], relayBackendPublicKey[:], nonce, encryptedData, len(encryptedData))

				packetData = buffer[:8+encryptedBytes+crypto.Box_NonceSize]

				copy(packetData[8+encryptedBytes:], nonce)

				// post to relay backend

				err := PostBinary(fmt.Sprintf("%s/relay_update", relayBackendHostname), packetData)
				if err != nil {
					core.Error("failed to post relay update to relay backend: %v", err)
				}
			}
		}
	}()
}

func PostBinary(url string, data []byte) error {

	buffer := bytes.NewBuffer(data)

	request, _ := http.NewRequest("POST", url, buffer)

	request.Header.Add("Content-Type", "application/octet-stream")

	httpClient := &http.Client{}
	response, err := httpClient.Do(request)
	if err != nil {
		return err
	}

	if response.StatusCode != 200 {
		return fmt.Errorf("got response %d", response.StatusCode)
	}

	body, error := ioutil.ReadAll(response.Body)
	if error != nil {
		return fmt.Errorf("could not read response: %v", err)
	}

	response.Body.Close()

	_ = body

	return nil
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

func Bash(command string) (bool,string) {
	return RunCommand("bash", []string{"-c", command})
}

func DetectGoogleRelayAddress(input string) string {
	result, output := Bash("curl -s http://metadata/computeMetadata/v1/instance/network-interfaces/0/access-configs/0/external-ip -H \"Metadata-Flavor: Google\" --max-time 10 -vs 2>/dev/null")
	if !result {
		return input
	}
	output = strings.TrimSuffix(output, "\n")
	core.Log("google cloud relay address is '%s'", output)
	return output
}
