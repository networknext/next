package main

import (
	"fmt"
	"github.com/networknext/backend/modules/transport"
	"math/rand"
	"net/http"
	"sync"

	"github.com/networknext/backend/modules/routing"
)

func main() {
	//WIP

	//numRelays to fake
	//todo envvar numRelays
	
	numRelays := 5

	//run time
	//time the load test runs for before calling shutdown to the relay backend.
	//todo envvar runtime

	//no init
	FeatureNoInit := false

	//get relay info from firestore
	//todo get "real" fake relays from firestore, create them first
	relayArr := make([]routing.Relay, numRelays)
	for i := 0; i < numRelays; i++ {
		relayArr[i] = tempFakeRelay(i)
	}

	//create fake relay route Bases
	relayRoutes := make(map[uint64]map[uint64]routeBase)
	for i := 0; i < numRelays; i++ {
		relayI := relayArr[i]

		for j := 0; j < numRelays; j++ {
			relayJ := relayArr[j]

			if relayI.ID == relayJ.ID {
				continue
			}

			base := newRouteBase()
			relayRoutes[relayI.ID][relayJ.ID] = base
			relayRoutes[relayJ.ID][relayI.ID] = base
		}
	}

	//send first update/init
	//first update is different because it has no packet data
	//todo send init

	var wgInit sync.WaitGroup
	if !FeatureNoInit {
		for _, relay := range relayArr {
			wgInit.Add(1)
			go func() {
				initRequest := newInit(relay)
				initBinary,err := initRequest.MarshalBinary()
				resp, err := http.Post()
			}()
		}
	}

	//todo send first update/ noinit

	//send update
	//todo go routine for said runtime

	//send shutdowns
	//todo send shutdowns for each relay

}

func tempFakeRelay(i int) routing.Relay {
	return routing.Relay{
		Name: fmt.Sprintf("fake_relay_%v", i),
		ID:   uint64(i),
	}
}

type routeBase struct {
	rtt        float32
	jitter     float32
	packetLoss float32
}

func newRouteBase() routeBase {
	rb := new(routeBase)
	rb.rtt = float32(rand.Int31n(300))
	rb.jitter = float32(rand.Int31n(30))
	rb.packetLoss = 0.0

	return *rb
}

func newInit(relay routing.Relay) transport.RelayInitRequest{

	//todo needs public/private key

	return transport.RelayInitRequest{

		Magic:         transport.InitRequestMagic,
		Version:        transport.VersionNumberInitRequest,
		Nonce:          []byte, //todo
		Address:        relay.Addr,
		EncryptedToken: []byte, //todo
		RelayVersion:   "0.0.0",
	}
}

func newPacketData(base routeBase) (float32, float32, float32){

}

