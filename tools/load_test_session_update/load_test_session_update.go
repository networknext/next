package main

import (
	"fmt"
	"math/rand"
	"net"
	// "runtime"
	"sync/atomic"
	"time"

	"github.com/networknext/backend/modules/common"
	"github.com/networknext/backend/modules/constants"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/crypto"
	db "github.com/networknext/backend/modules/database"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/handlers"
	"github.com/networknext/backend/modules/packets"
)

const NumRelays = constants.MaxRelays
const NumSessions = 1000

var ServerBackendAddress = core.ParseAddress("127.0.0.1:50000")
var ServerBackendPublicKey []byte
var ServerBackendPrivateKey []byte

var RelayBackendPublicKey []byte
var RelayBackendPrivateKey []byte

var SessionId uint64

var DatacenterId uint64

var BuyerId uint64
var BuyerPublicKey []byte
var BuyerPrivateKey []byte

var ClientPublicKey []byte
var ClientPrivateKey []byte

var ServerPublicKey []byte
var ServerPrivateKey []byte

type Update struct {
	from       net.UDPAddr
	packetData []byte
}

func RunSessionUpdateThreads(threadCount int, updateChannels []chan *Update) {

	for k := 0; k < threadCount; k++ {

		go func(thread int) {

			time.Sleep(time.Duration(rand.Intn(10000)) * time.Millisecond)

			// todo: create session data and write to byte array

			clientAddress := core.ParseAddress("127.0.0.1:40000")
			serverAddress := core.ParseAddress("127.0.0.1:50000")

			for {

				for j := 0; j < NumSessions; j++ {

					packet := packets.SDK5_SessionUpdateRequestPacket{
						Version:      packets.SDKVersion{5, 0, 0},
						BuyerId:      BuyerId,
						DatacenterId: uint64(j),
						SessionId:    SessionId,
						SliceNumber:  10,
						// todo: SessionDataBytes
						// todo: SessionData
						// todo: SessionDataSignature
						ClientAddress: clientAddress,
						ServerAddress: serverAddress,
						HasNearRelayPings: true,
						DirectRTT: 200,
					}

					copy(packet.ClientRoutePublicKey[:], ClientPublicKey)
					copy(packet.ServerRoutePublicKey[:], ServerPublicKey)

					packet.NumNearRelays = constants.MaxNearRelays
					for i := 0; i < constants.MaxNearRelays; i++ {
						packet.NearRelayIds[i] = uint64((j+i) % NumRelays)
						packet.NearRelayRTT[i] = int32(common.RandomInt(0,10))
					}

					packetData, err := packets.SDK5_WritePacket(&packet, packets.SDK5_SESSION_UPDATE_REQUEST_PACKET, packets.SDK5_MaxPacketBytes, &serverAddress, &ServerBackendAddress, BuyerPrivateKey[:])
					if err != nil {
						panic("failed to write server update packet")
					}

					updateChannel := updateChannels[j%len(updateChannels)]

					update := Update{}
					update.from = serverAddress
					update.packetData = packetData

					updateChannel <- &update
				}

				time.Sleep(10 * time.Second)
			}
		}(k)
	}
}

func RunHandlerThreads(threadCount int, updateChannels []chan *Update, numSessionUpdatesProcessed *uint64) {

	for k := 0; k < threadCount; k++ {

		go func(thread int) {

			updateChannel := updateChannels[thread]

			buyer := db.Buyer{}
			buyer.Id = BuyerId
			buyer.Name = "buyer"
			buyer.Live = true
			buyer.Debug = false
			buyer.PublicKey = BuyerPublicKey[:]
			buyer.RouteShader = core.NewRouteShader()
			buyer.RouteShader.AnalysisOnly = false

			datacenters := make([]db.Datacenter, NumRelays)
			for i := range datacenters {
				datacenters[i].Id = uint64(i)
				datacenters[i].Name = fmt.Sprintf("datacenter-%d", i)
				datacenters[i].Latitude = 100
				datacenters[i].Longitude = 200
			}

			database := db.CreateDatabase()
			database.BuyerMap[BuyerId] = &buyer
			database.DatacenterMaps[BuyerId] = make(map[uint64]*db.DatacenterMap)
			for i := range datacenters {
				database.DatacenterMap[datacenters[i].Id] = &datacenters[i]
				database.DatacenterMaps[BuyerId][datacenters[i].Id] = &db.DatacenterMap{DatacenterId: uint64(i), BuyerId: BuyerId, EnableAcceleration: true}
			}

			err := database.Validate()
			if err != nil {
				panic(fmt.Sprintf("database did not validate: %v\n", err))
			}

			routeMatrix := common.RouteMatrix{}

			handler := handlers.SDK5_Handler{}
			handler.Database = database
			handler.RouteMatrix = &routeMatrix
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
					routeMatrix.CreatedAt = uint64(time.Now().Unix())
					handlers.SDK5_PacketHandler(&handler, nil, &update.from, update.packetData)
					if !handler.Events[handlers.SDK5_HandlerEvent_SentSessionUpdateResponsePacket] {
						panic("failed to process session update")
					}
					atomic.AddUint64(numSessionUpdatesProcessed, 1)
				}
			}

		}(k)
	}
}

func RunWatcherThread(numSessionUpdatesProcessed *uint64) {

	go func() {

		ticker := time.NewTicker(time.Second)

		iteration := uint64(0)

		start := time.Now()

		for {
			select {
			case <-ticker.C:
				numUpdates := atomic.LoadUint64(numSessionUpdatesProcessed)
				fmt.Printf("iteration %d: %8d session updates | %7d session updates per-second\n", iteration, numUpdates, uint64(float64(numUpdates)/time.Since(start).Seconds()))
				iteration++
			}
		}
	}()
}

func main() {

	// core.DebugLogs = false

	BuyerId = rand.Uint64()

	SessionId = rand.Uint64()

	DatacenterId = rand.Uint64()

	BuyerPublicKey, BuyerPrivateKey = crypto.Sign_KeyPair()

	ServerBackendPublicKey, ServerBackendPrivateKey = crypto.Sign_KeyPair()

	RelayBackendPublicKey, RelayBackendPrivateKey = crypto.Box_KeyPair()

	ClientPublicKey, ClientPrivateKey = crypto.Box_KeyPair()

	ServerPublicKey, ServerPrivateKey = crypto.Box_KeyPair()

	numSessionUpdateThreads := envvar.GetInt("NUM_SESSION_UPDATE_THREADS", 1) //000)

	numHandlerThreads := envvar.GetInt("NUM_HANDLER_THREADS", 1)//runtime.NumCPU())

	updateChannels := make([]chan *Update, numHandlerThreads)
	for i := range updateChannels {
		updateChannels[i] = make(chan *Update, 1024*1024)
	}

	var numSessionUpdatesProcessed uint64

	RunSessionUpdateThreads(numSessionUpdateThreads, updateChannels)

	RunHandlerThreads(numHandlerThreads, updateChannels, &numSessionUpdatesProcessed)

	RunWatcherThread(&numSessionUpdatesProcessed)

	time.Sleep(time.Minute)
}
