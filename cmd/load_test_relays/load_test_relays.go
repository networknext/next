package main

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/constants"
	"github.com/networknext/next/modules/core"
	"github.com/networknext/next/modules/envvar"
	"github.com/networknext/next/modules/packets"
)

var service *common.Service
var relayAddress string
var relayBackendHostname string
var numRelays int

func main() {

	service = common.CreateService("load_test_relays")

	numRelays = envvar.GetInt("NUM_RELAYS", 1000)

	relayAddress = envvar.GetString("RELAY_ADDRESS", "127.0.0.1")

	relayBackendHostname = envvar.GetString("RELAY_BACKEND_HOSTNAME", "http://127.0.0.1:30000")

	core.Log("simulating %d relays", numRelays)

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
		sampleRelayIds[i] = common.RelayId(fmt.Sprintf("%s:%d", relayAddress, 1000+i))
	}

	ticker := time.NewTicker(time.Second)

	go func() {
		for {
			select {

			case <-service.Context.Done():
				return

			case <-ticker.C:

				fmt.Printf("update relay %d\n", index)

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
					packet.SampleRTT[i] = uint8(common.RandomInt(0, 100))
					packet.SampleJitter[i] = uint8(common.RandomInt(0, 10))
					packet.SamplePacketLoss[i] = uint16(common.RandomInt(0, 500))
				}

				// write relay update packet

				const BufferSize = 16 * 1024

				var buffer [BufferSize]byte

				packetData := packet.Write(buffer[:])

				// encrypt relay update

/*
    // encrypt data after relay address

    const int encrypt_buffer_length = (int) ( p - encrypt_buffer );

    uint8_t nonce[crypto_box_NONCEBYTES];
    relay_random_bytes( nonce, crypto_box_NONCEBYTES );

    if ( crypto_box_easy( encrypt_buffer, encrypt_buffer, encrypt_buffer_length, nonce, main->relay_backend_public_key, main->relay_private_key ) != 0 )
    {
        printf( "error: failed to encrypt relay update\n" );
        return RELAY_ERROR;
    }
    
    p += crypto_box_MACBYTES;

    memcpy( p, nonce, crypto_box_NONCEBYTES );

    p += crypto_box_NONCEBYTES;

    const int update_data_length = p - update_data;
*/

				// post to relay backend

				err := PostBinary(fmt.Sprintf("%s/relay_update", relayBackendHostname), packetData)
				if err != nil {
					core.Error("failed to post relay update to relay backend: %v", err)
					os.Exit(1)
				}
			}
		}
	}()
}

func PostBinary(url string, data []byte) error {

	fmt.Printf("post binary to %s\n", url)

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
