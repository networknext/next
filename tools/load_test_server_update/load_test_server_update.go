package main

import (
	"fmt"
	"math/rand"
	"net"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/networknext/backend/modules/constants"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/packets"
	"github.com/networknext/backend/modules/handlers"
)

var ServerBackendAddress = core.ParseAddress("127.0.0.1:50000")
var ServerBackendPublicKey []byte
var ServerBackendPrivateKey []byte

type Update struct {
	from       net.UDPAddr
	packetData []byte
}

func RunServerUpdateThreads(threadCount int, updateChannels []chan *Update) {

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

					buffer := [packets.SDK5_MaxPacketBytes]byte{}

					stream := encoding.CreateWriteStream(buffer[:])

					err := packet.Serialize(stream)
					if err != nil {
						panic("packet failed to serialize write")
					}
					stream.Flush()
					packetBytes := stream.GetBytesProcessed()

					packetData := buffer[:packetBytes]

					// todo: need to sign the packet and apply the pittle and chonkle

					updateChannel := updateChannels[j%len(updateChannels)]

					update := Update{}
					update.from = serverAddresses[j]
					update.packetData = packetData

					updateChannel <- &update
				}

				time.Sleep(10 * time.Second)
			}
		}(k)
	}
}

func RunHandlerThreads(threadCount int, updateChannels []chan *Update, numServerUpdatesProcessed *uint64) {

	for k := 0; k < threadCount; k++ {

		go func(thread int) {

			updateChannel := updateChannels[thread]

			handler := handlers.SDK5_Handler{}
			// todo: need to set handler.Database here to one with the buyer id and keypair in it
			handler.ServerBackendAddress = ServerBackendAddress
			handler.ServerBackendPublicKey = ServerBackendPublicKey
			handler.ServerBackendPrivateKey = ServerBackendPrivateKey
			handler.MaxPacketSize = packets.SDK5_MaxPacketBytes
			handler.GetMagicValues = func() ([constants.MagicBytes]byte, [constants.MagicBytes]byte, [constants.MagicBytes]byte) {
				return [constants.MagicBytes]byte{}, [constants.MagicBytes]byte{}, [constants.MagicBytes]byte{}
			}

			for {
				select {
				case update := <-updateChannel:
					handlers.SDK5_PacketHandler(&handler, nil, &update.from, update.packetData)
					if !handler.Events[handlers.SDK5_HandlerEvent_SentServerUpdateResponsePacket] {
						panic("failed to process server update")
					}
					atomic.AddUint64(numServerUpdatesProcessed, 1)
				}
			}

		}(k)
	}
}

func RunWatcherThread(numServerUpdatesProcessed *uint64) {

	go func() {

		ticker := time.NewTicker(time.Second)

		iteration := uint64(0)

		start := time.Now()

		for {
			select {
			case <-ticker.C:
				numUpdates := atomic.LoadUint64(numServerUpdatesProcessed)
				fmt.Printf("iteration %d: %7d server updates per-second\n", iteration, uint64(float64(numUpdates)/time.Since(start).Seconds()))
				iteration++
			}
		}
	}()
}

func main() {

	ServerBackendPublicKey, ServerBackendPrivateKey = crypto.Sign_KeyPair()

	numServerUpdateThreads := envvar.GetInt("NUM_SERVER_UPDATE_THREADS", 1000)

	numHandlerThreads := envvar.GetInt("NUM_HANDLER_THREADS", runtime.NumCPU())

	// todo: we're going to need to generate a buyer keypair here

	updateChannels := make([]chan *Update, numHandlerThreads)
	for i := range updateChannels {
		updateChannels[i] = make(chan *Update, 1024*1024)
	}

	var numServerUpdatesProcessed uint64

	RunServerUpdateThreads(numServerUpdateThreads, updateChannels)

	RunHandlerThreads(numHandlerThreads, updateChannels, &numServerUpdatesProcessed)

	RunWatcherThread(&numServerUpdatesProcessed)

	time.Sleep(time.Minute)
}
