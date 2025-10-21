package main

import (
	"fmt"
	"math/rand"
	"net"
	"runtime"
	"sync/atomic"
	"time"

	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/constants"
	"github.com/networknext/next/modules/core"
	"github.com/networknext/next/modules/crypto"
	db "github.com/networknext/next/modules/database"
	"github.com/networknext/next/modules/encoding"
	"github.com/networknext/next/modules/envvar"
	"github.com/networknext/next/modules/handlers"
	"github.com/networknext/next/modules/packets"
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

var SellerId uint64

var BuyerId uint64
var BuyerPublicKey []byte
var BuyerPrivateKey []byte

var ClientPublicKey []byte
var ClientPrivateKey []byte

var ServerPublicKey []byte
var ServerPrivateKey []byte

var RelayPublicKey []byte
var RelayPrivateKey []byte

type Update struct {
	from       net.UDPAddr
	packetData []byte
}

func RunSessionUpdateThreads(threadCount int, updateChannels []chan *Update) {

	for k := 0; k < threadCount; k++ {

		go func(thread int) {

			time.Sleep(time.Duration(rand.Intn(10000)) * time.Millisecond)

			clientAddress := core.ParseAddress("127.0.0.1:40000")
			serverAddress := core.ParseAddress("127.0.0.1:50000")

			sessionData := packets.SDK_SessionData{}
			sessionData.Version = packets.SDK_SessionDataVersion_Write
			sessionData.SessionId = SessionId
			sessionData.SliceNumber = 10
			sessionData.ExpireTimestamp = uint64(time.Now().Unix() + 1000)

			sessionData_Output := make([]byte, packets.SDK_MaxSessionDataSize)
			sessionData_Signature := make([]byte, crypto.Sign_SignatureSize)
			{
				stream := encoding.CreateWriteStream(sessionData_Output)
				err := sessionData.Serialize(stream)
				if err != nil {
					panic("failed to write session data")
				}
				stream.Flush()
				sessionDataBytes := stream.GetBytesProcessed()
				sessionData_Output = sessionData_Output[:sessionDataBytes]
				copy(sessionData_Signature, crypto.Sign(sessionData_Output, ServerBackendPrivateKey))
			}

			for {

				for j := 0; j < NumSessions; j++ {

					packet := packets.SDK_SessionUpdateRequestPacket{
						Version:             packets.SDKVersion{1, 0, 0},
						BuyerId:             BuyerId,
						DatacenterId:        uint64(j),
						SessionId:           SessionId,
						SliceNumber:         10,
						SessionDataBytes:    int32(len(sessionData_Output)),
						ClientAddress:       clientAddress,
						ServerAddress:       serverAddress,
						HasClientRelayPings: true,
						HasServerRelayPings: true,
					}

					if (j % 10) == 0 {
						packet.DirectRTT = 200
					}

					copy(packet.SessionData[:], sessionData_Output)
					copy(packet.SessionDataSignature[:], sessionData_Signature)

					copy(packet.ClientRoutePublicKey[:], ClientPublicKey)
					copy(packet.ServerRoutePublicKey[:], ServerPublicKey)

					packet.NumClientRelays = constants.MaxClientRelays
					for i := 0; i < constants.MaxClientRelays; i++ {
						packet.ClientRelayIds[i] = uint64((j + i) % NumRelays)
						packet.ClientRelayRTT[i] = int32(common.RandomInt(0, 10))
					}

					packet.NumServerRelays = constants.MaxServerRelays
					for i := 0; i < constants.MaxServerRelays; i++ {
						packet.ServerRelayIds[i] = uint64((j + i) % NumRelays)
						packet.ServerRelayRTT[i] = int32(common.RandomInt(0, 1))
					}

					packetData, err := packets.SDK_WritePacket(&packet, packets.SDK_SESSION_UPDATE_REQUEST_PACKET, packets.SDK_MaxPacketBytes, &serverAddress, &ServerBackendAddress, BuyerPrivateKey[:])
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

	buyer := db.Buyer{}
	buyer.Id = BuyerId
	buyer.Name = "buyer"
	buyer.Live = true
	buyer.Debug = false
	buyer.PublicKey = BuyerPublicKey[:]
	buyer.RouteShader = core.NewRouteShader()

	datacenters := make([]db.Datacenter, NumRelays)
	for i := range datacenters {
		datacenters[i].Id = uint64(i)
		datacenters[i].Name = fmt.Sprintf("datacenter-%d", i)
		datacenters[i].Latitude = 100
		datacenters[i].Longitude = 200
	}

	seller := db.Seller{}
	seller.Name = "seller"
	seller.Id = SellerId

	relays := make([]db.Relay, NumRelays)
	for i := range relays {
		relays[i].Id = uint64(i)
		relays[i].Name = fmt.Sprintf("relay-%d", i)
		relays[i].DatacenterId = uint64(i)
		relays[i].PublicAddress = core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", 2000+i))
		relays[i].PublicKey = RelayPublicKey
		relays[i].PrivateKey = RelayPrivateKey
		relays[i].Seller = &seller
		relays[i].Datacenter = &datacenters[i]
		relays[i].SSHUser = "ubuntu"
		relays[i].SSHAddress = core.ParseAddress("127.0.0.1:22")
	}

	database := db.CreateDatabase()
	database.Relays = relays
	for i := range relays {
		database.RelayMap[relays[i].Id] = &relays[i]
		array := [1]uint64{uint64(i)}
		database.DatacenterRelays[uint64(i)] = array[:]
	}
	database.SellerMap[SellerId] = &seller
	database.BuyerMap[BuyerId] = &buyer
	database.BuyerDatacenterSettings[BuyerId] = make(map[uint64]*db.BuyerDatacenterSettings)
	for i := range datacenters {
		database.DatacenterMap[datacenters[i].Id] = &datacenters[i]
		database.BuyerDatacenterSettings[BuyerId][datacenters[i].Id] = &db.BuyerDatacenterSettings{DatacenterId: uint64(i), BuyerId: BuyerId, EnableAcceleration: true}
	}

	database.Fixup()

	err := database.Validate()
	if err != nil {
		panic(fmt.Sprintf("database did not validate: %v\n", err))
	}

	size := core.TriMatrixLength(NumRelays)
	costs := make([]uint8, size)
	var entries []core.RouteEntry
	{
		for i := 0; i < NumRelays; i++ {
			for j := 0; j < i; j++ {
				index := core.TriMatrixIndex(i, j)
				costs[index] = uint8(common.RandomInt(0, 255))
			}
		}

		numSegments := 256
		relayDatacenterIds := make([]uint64, NumRelays)
		for i := range relayDatacenterIds {
			relayDatacenterIds[i] = uint64(i)
		}

		destRelays := make([]bool, NumRelays)
		for i := range destRelays {
			destRelays[i] = true
		}

		relayPrice := make([]byte, NumRelays)

		entries = core.Optimize2(NumRelays, numSegments, costs, relayPrice, relayDatacenterIds, destRelays)
	}

	routeMatrix := common.RouteMatrix{}
	routeMatrix.RelayIds = make([]uint64, NumRelays)
	routeMatrix.RelayIdToIndex = make(map[uint64]int32)
	routeMatrix.RelayAddresses = make([]net.UDPAddr, NumRelays)
	routeMatrix.RelayNames = make([]string, NumRelays)
	routeMatrix.RelayLatitudes = make([]float32, NumRelays)
	routeMatrix.RelayLongitudes = make([]float32, NumRelays)
	routeMatrix.RelayDatacenterIds = make([]uint64, NumRelays)
	routeMatrix.DestRelays = make([]bool, NumRelays)
	routeMatrix.RouteEntries = entries
	for i := 0; i < NumRelays; i++ {
		routeMatrix.RelayIds[i] = uint64(i)
		routeMatrix.RelayIdToIndex[uint64(i)] = int32(i)
		routeMatrix.RelayAddresses[i] = core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", 2000+i))
		routeMatrix.RelayNames[i] = fmt.Sprintf("relay-%d", i)
		routeMatrix.RelayLatitudes[i] = 100
		routeMatrix.RelayLongitudes[i] = 200
		routeMatrix.RelayDatacenterIds[i] = uint64(i)
		routeMatrix.DestRelays[i] = true
	}

	handler := handlers.SDK_Handler{}
	handler.Database = database
	handler.RouteMatrix = &routeMatrix
	handler.ServerBackendAddress = ServerBackendAddress
	handler.ServerBackendPublicKey = ServerBackendPublicKey
	handler.RelayBackendPublicKey = RelayBackendPublicKey
	handler.RelayBackendPrivateKey = RelayBackendPrivateKey
	handler.ServerBackendPrivateKey = ServerBackendPrivateKey
	handler.MaxPacketSize = packets.SDK_MaxPacketBytes
	handler.GetMagicValues = func() ([constants.MagicBytes]byte, [constants.MagicBytes]byte, [constants.MagicBytes]byte) {
		return [constants.MagicBytes]byte{}, [constants.MagicBytes]byte{}, [constants.MagicBytes]byte{}
	}

	for k := 0; k < threadCount; k++ {

		go func(thread int) {

			updateChannel := updateChannels[thread]

			for {
				select {
				case update := <-updateChannel:
					routeMatrix.CreatedAt = uint64(time.Now().Unix())
					handlers.SDK_PacketHandler(&handler, nil, &update.from, update.packetData)
					if !handler.Events[handlers.SDK_HandlerEvent_SentSessionUpdateResponsePacket] {
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

	core.DebugLogs = false

	SellerId = rand.Uint64()

	BuyerId = rand.Uint64()

	SessionId = rand.Uint64()

	DatacenterId = rand.Uint64()

	BuyerPublicKey, BuyerPrivateKey = crypto.Sign_KeyPair()

	ServerBackendPublicKey, ServerBackendPrivateKey = crypto.Sign_KeyPair()

	RelayBackendPublicKey, RelayBackendPrivateKey = crypto.Box_KeyPair()

	ClientPublicKey, ClientPrivateKey = crypto.Box_KeyPair()

	ServerPublicKey, ServerPrivateKey = crypto.Box_KeyPair()

	RelayPublicKey, RelayPrivateKey = crypto.Box_KeyPair()

	numSessionUpdateThreads := envvar.GetInt("NUM_SESSION_UPDATE_THREADS", 1000)

	numHandlerThreads := envvar.GetInt("NUM_HANDLER_THREADS", runtime.NumCPU())

	updateChannels := make([]chan *Update, numHandlerThreads)
	for i := range updateChannels {
		updateChannels[i] = make(chan *Update, 1024)
	}

	var numSessionUpdatesProcessed uint64

	RunSessionUpdateThreads(numSessionUpdateThreads, updateChannels)

	RunHandlerThreads(numHandlerThreads, updateChannels, &numSessionUpdatesProcessed)

	RunWatcherThread(&numSessionUpdatesProcessed)

	time.Sleep(time.Minute)
}
