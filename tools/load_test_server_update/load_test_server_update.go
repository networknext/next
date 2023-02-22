package main

import (
	"fmt"
	"math/rand"
	"time"
	"net"
	"runtime"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/packets"
	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/modules/envvar"
)

func RunServerUpdateThreads(threadCount int, updateChannels []chan []byte) {

	for k := 0; k < threadCount; k++ {

		go func(thread int) {

			time.Sleep(time.Duration(rand.Intn(10000)) * time.Millisecond)

			const NumServers = 10000

			serverAddresses := make([]net.UDPAddr, NumServers)
			for i := range serverAddresses {
				serverAddresses[i] = core.ParseAddress(fmt.Sprintf("127.0.%d.%d:%d", i>>8, i&0xFF, 2000+thread))
			}

			for {

				for j := 0; j < NumServers; j++ {

					packet := packets.SDK5_ServerUpdateRequestPacket{
						Version:      packets.SDKVersion{5, 0, 0},
						BuyerId:      rand.Uint64(),
						RequestId:    rand.Uint64(),
						DatacenterId: rand.Uint64(),
					}

					const BufferSize = 1024

					buffer := [BufferSize]byte{}

					stream := encoding.CreateWriteStream(buffer[:])

					err := packet.Serialize(stream)
					if err != nil {
						panic("packet failed to serialize")
					}
					stream.Flush()
					packetBytes := stream.GetBytesProcessed()

					packetData := buffer[:packetBytes]

					// todo: need to encrypt, sign, pittle/chonkle the packet...

					updateChannel := updateChannels[j%len(updateChannels)]

					updateChannel <- packetData
				}

				time.Sleep(10 * time.Second)
			}
		}(k)
	}
}

func RunHandlerThreads(threadCount int, updateChannels []chan []byte) {

	for k := 0; k < threadCount; k++ {

		go func(thread int) {
	
			updateChannel := updateChannels[thread]

			for {
				// todo
				_ = updateChannel
			}

		}(k)
	}
}

/*
	readStream := encoding.CreateReadStream(buffer[:packetBytes])
	err = readPacket.Serialize(readStream)
	assert.Nil(t, err)
*/

func RunWatcherThread() {

	go func() {

		ticker := time.NewTicker(time.Second)

		iteration := uint64(0)

		for {

			select {

			case <-ticker.C:
				// numSent := atomic.LoadUint64(numMessagesSent)
				// numReceived := atomic.LoadUint64(numMessagesReceived)
				fmt.Printf("iteration %d\n", iteration)
				iteration++
			}
		}
	}()
}

func main() {

	numServerUpdateThreads := envvar.GetInt("NUM_SERVER_UPDATE_THREADS", 1) // todo: 1000 or so
	numHandlerThreads := envvar.GetInt("NUM_HANDLER_THREADS", runtime.NumCPU())

	updateChannels := make([]chan []byte, numHandlerThreads)
	for i := range updateChannels {
		updateChannels[i] = make(chan []byte, 1024*1024)
	}

	RunServerUpdateThreads(numServerUpdateThreads, updateChannels)

	RunHandlerThreads(numHandlerThreads, updateChannels)

	RunWatcherThread()

	time.Sleep(time.Minute)
}
