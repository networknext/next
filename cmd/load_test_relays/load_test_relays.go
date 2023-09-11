package main

import (
	"fmt"
	"time"

	"github.com/networknext/next/modules/constants"
	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/core"
	"github.com/networknext/next/modules/envvar"
	"github.com/networknext/next/modules/packets"
)


func main() {

	service := common.CreateService("load_test_relays")

	numRelays := envvar.GetInt("NUM_RELAYS", 1000)

	core.Log("simulating %d relays", numRelays)

	go SimulateRelays(service, numRelays)

	service.WaitForShutdown()
}

func SimulateRelays(service *common.Service, numRelays int) {
	for i := 0; i < numRelays; i++ {
		go RunRelay(service, numRelays, i)
	}
}

func RunRelay(service *common.Service, numRelays int, index int) {

	time.Sleep(time.Duration(common.RandomInt(0,1000))*time.Millisecond)
	
	startTime := time.Now().Unix()
	
	address := core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", 10000+index))									// todo: need to pass in RELAY_ADDRESS env var

	sampleRelayIds := make([]uint64, numRelays)
	for i := 0; i < numRelays; i++ {
		sampleRelayIds[i] = common.RelayId(fmt.Sprintf("127.0.0.1:%d", 1000+i))								// todo: RELAY_ADDRESS
	}
	
	ticker := time.NewTicker(time.Second)
	
	go func() {
		for {
			select {

			case <-service.Context.Done():
				return

			case <-ticker.C:

				packet := packets.RelayUpdateRequestPacket{
					Version:     uint8(packets.RelayUpdateRequestPacket_VersionMax),
					CurrentTime: uint64(time.Now().Unix()),
					StartTime:   uint64(startTime),
					Address:     address,
					NumSamples:  uint32(numRelays),
					SessionCount: 100,
					EnvelopeBandwidthUpKbps: uint32(common.RandomInt(10000,20000)),
					EnvelopeBandwidthDownKbps: uint32(common.RandomInt(10000,20000)),
					PacketsSentPerSecond: float32(common.RandomInt(1000,2000)),
					PacketsReceivedPerSecond: float32(common.RandomInt(1000,2000)),
					BandwidthSentKbps: float32(common.RandomInt(10000,20000)),
					BandwidthReceivedKbps: float32(common.RandomInt(10000,20000)),
					NearPingsPerSecond: float32(common.RandomInt(10000, 20000)),
					RelayPingsPerSecond: float32(common.RandomInt(10000, 20000)),
					RelayFlags: 0,
					RelayVersion: "load test",
					NumRelayCounters: constants.NumRelayCounters,
				}

				copy(packet.SampleRelayId[:], sampleRelayIds)

				for i := 0; i < int(packet.NumSamples); i++ {
					packet.SampleRTT[i] = uint8(common.RandomInt(0, 100))
					packet.SampleJitter[i] = uint8(common.RandomInt(0, 10))
					packet.SamplePacketLoss[i] = uint16(common.RandomInt(0, 500))
				}

				fmt.Printf("update relay %d\n", index)

				// todo: send request to RELAY_BACKEND0
			}
		}
	}()
}
