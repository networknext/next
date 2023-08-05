package main

import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"time"

	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/constants"
	"github.com/networknext/next/modules/core"
	"github.com/networknext/next/modules/envvar"
)

type Update struct {
	relayId          uint64
	relayName        string
	relayAddress     net.UDPAddr
	relayVersion     string
	relayFlags       uint64
	sessions         int
	numSamples       int
	sampleRelayId    []uint64
	sampleRTT        []uint8
	sampleJitter     []uint8
	samplePacketLoss []uint16
	counters         [constants.NumRelayCounters]uint64
}

func RunInsertThreads(ctx context.Context, numRelays int, updateChan chan *Update) {

	for k := 0; k < numRelays; k++ {

		go func(relayIndex int) {

			time.Sleep(time.Duration(rand.Intn(1000)) * time.Millisecond)

			ticker := time.NewTicker(time.Second)

			relayName := fmt.Sprintf("local-%d", relayIndex)
			relayAddress := core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", 2000+relayIndex))

			for {

				select {

				case <-ctx.Done():
					return

				case <-ticker.C:
					update := Update{}
					update.relayId = uint64(relayIndex)
					update.relayName = relayName
					update.relayAddress = relayAddress
					update.relayVersion = "load-test"
					update.sessions = common.RandomInt(0, 1000)
					update.numSamples = numRelays - 1
					update.sampleRelayId = make([]uint64, update.numSamples)
					update.sampleRTT = make([]uint8, update.numSamples)
					update.sampleJitter = make([]uint8, update.numSamples)
					update.samplePacketLoss = make([]uint16, update.numSamples)
					index := 0
					for i := 0; i < update.numSamples; i++ {
						if i == relayIndex {
							continue
						}
						update.sampleRelayId[index] = uint64(index)
						update.sampleRTT[index] = uint8(common.RandomInt(0, 255))
						update.sampleJitter[index] = uint8(common.RandomInt(0, 255))
						index++
					}
					updateChan <- &update
				}
			}
		}(k)
	}
}

func RunRelayManagerThread(ctx context.Context, numRelays int, updateChan chan *Update) {

	go func() {

		local := true

		relayManager := common.CreateRelayManager(local)

		relayIds := make([]uint64, numRelays)
		for i := 0; i < numRelays; i++ {
			relayIds[i] = uint64(i)
		}

		maxJitter := float32(10.0)
		maxPacketLoss := float32(0.1)

		ticker := time.NewTicker(time.Second)

		iteration := uint64(0)

		for {

			select {

			case <-ctx.Done():
				return

			case update := <-updateChan:
				relayManager.ProcessRelayUpdate(time.Now().Unix(), update.relayId, update.relayName, update.relayAddress, update.sessions, update.relayVersion, update.relayFlags, update.numSamples, update.sampleRelayId, update.sampleRTT, update.sampleJitter, update.samplePacketLoss, update.counters[:])
				break

			case <-ticker.C:
				start := time.Now()
				costs := relayManager.GetCosts(start.Unix(), relayIds, maxJitter, maxPacketLoss)
				fmt.Printf("iteration %d: %d relays, cost array is %d bytes (%dms)\n", iteration, numRelays, len(costs), time.Since(start).Milliseconds())
				iteration++
			}
		}
	}()
}

func main() {

	numRelays := envvar.GetInt("NUM_RELAYS", constants.MaxRelays)

	updateChan := make(chan *Update, 1024*numRelays)

	RunInsertThreads(context.Background(), numRelays, updateChan)

	RunRelayManagerThread(context.Background(), numRelays, updateChan)

	time.Sleep(time.Minute)
}
