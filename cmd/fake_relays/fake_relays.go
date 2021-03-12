package main

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/networknext/backend/modules/crypto"

	"github.com/go-kit/kit/log/level"
	"github.com/networknext/backend/modules/backend"
	"github.com/networknext/backend/modules/common/helpers"
	"github.com/networknext/backend/modules/envvar"
	"github.com/networknext/backend/modules/transport"

	"github.com/networknext/backend/modules/routing"
)

var (
	buildtime             string
	commitMessage         string
	sha                   string
	tag                   string
	relayPrivatekey       string
	relayBackendPublicKey string
)

const (
	maxRTT               = 300
	maxJitter            = 30
	maxMultiplierPercent = 10

	// pLChance is 1 in n

	relayDisabled = 1
	relayEnabled  = 2
	relayCrashed  = 3

	// chances are 1 in n
	pLChance             = 10000
	pLValue              = .3
	relayCrashChance     = 100000
	relayCrashTimeout    = 60 * time.Second
	relayShutdownChance  = 100000
	relayShutdownTimeout = 30 * time.Second
)

func main() {
	os.Exit(mainReturnWithCode())
}

func mainReturnWithCode() int {
	serviceName := "fake_relays"
	fmt.Printf("%s: Git Hash: %s - Commit: %s\n", serviceName, sha, commitMessage)

	ctx := context.Background()

	gcpProjectID := backend.GetGCPProjectID()

	logger, err := backend.GetLogger(ctx, gcpProjectID, serviceName)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	// numRelays to fake
	numRelays, err := envvar.GetInt("NUM_FAKE_RELAYS", 5)
	if err != nil {
		level.Error(logger).Log("err", err)
	}

	featureNoInit, err := envvar.GetBool("FEATURE_NO_INIT", false)
	if err != nil {
		level.Error(logger).Log("err", err)
	}

	// get and verify relayBackendAddr
	relayBackendAddr := envvar.Get("RELAY_BACKEND_ADDR", "")
	if net.ParseIP(relayBackendAddr) == nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	relayPrivatekey = envvar.Get("RELAY_PRIVATE_KEY", "")
	if relayPrivatekey == "" {
		level.Error(logger).Log("msg", "relay private key not set")
		return 1
	}

	relayBackendPublicKey = envvar.Get("RELAY_BACKEND_PUBLIC_KEY", "")
	if relayBackendPublicKey == "" {
		level.Error(logger).Log("msg", "relay backend public key not set")
		return 1
	}

	// create fake relays
	storageRelayArr := make([]routing.Relay, numRelays)
	for i := 0; i < numRelays; i++ {
		storageRelayArr[i] = fakeRelay(i)
	}

	// create fake relay route Bases
	relayArr := make([]Relay, len(storageRelayArr))
	for i := 0; i < numRelays; i++ {
		relayI := storageRelayArr[i]
		newRelay := Relay{
			data:         relayI,
			state:        relayDisabled,
			stateChanged: time.Now().Add(-5 * time.Minute),
			routeBaseMap: make(map[uint64]routeBase),
		}

		for j := 0; j < numRelays; j++ {
			if i == j {
				continue
			}

			relayJ := storageRelayArr[j]

			base := newRouteBase()
			newRelay.routeBaseMap[relayJ.ID] = base
		}

		relayArr[i] = newRelay
	}

	// core logic
	shutdown := false
	initAddress := fmt.Sprintf("%s/relay_init", relayBackendAddr)
	updateAddress := fmt.Sprintf("%s/relay_update", relayBackendAddr)

	for i := 0; i < numRelays; i++ {

		relay := relayArr[i]

		go func(relay Relay) {

			var relaysToPing []uint64
			syncTimer := helpers.NewSyncTimer(1 * time.Second)
			for {
				syncTimer.Run()

				if shutdown {
					err := sendShutdown(relay, updateAddress)
					level.Error(logger).Log("err", err)
					return
				}

				if relay.state != relayEnabled {

					// WIP relay chaos
					// if relay.state == relayCrashed{
					//	if relay.stateChanged.Sinc
					//
					// }

					if featureNoInit {
						_, err := sendUpdateInit(relay, updateAddress)
						if err != nil {
							level.Error(logger).Log("err", err)
							continue
						}
						relaysToPing = relaysToPingFromRelayList(relayArr, relay.data.ID)

					} else {
						err := sendInit(relay, initAddress)
						if err != nil {
							level.Error(logger).Log("err", err)
							continue
						}
						relaysToPing = relaysToPingFromRelayList(relayArr, relay.data.ID)
					}
					continue
				}

				// WIP determine state change
				//isCrashed := rand.Int31n(relayCrashChance)
				//if isCrashed == 1 {
				//	relay.state = relayCrashed
				//	relay.stateChanged = time.Now()
				//	continue
				//}
				//
				//isShutdown := rand.Int31n(relayShutdownChance)
				//if isShutdown == 1 {
				//	relay.state = relayDisabled
				//	err := sendShutdown(relay, updateAddress)
				//	if err != nil {
				//		level.Error(logger).Log("err", err)
				//	}
				//	continue
				//}
				//if no state change populate packet update

				newRelaysToPing, err := sendUpdate(relay, relaysToPing, updateAddress)
				if err != nil {
					level.Error(logger).Log("err", err)
				}
				relaysToPing = newRelaysToPing
			}

		}(relay)
	}

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint
	shutdown = true
	time.Sleep(5 * time.Second)
	os.Exit(0)
	return 0
}

func fakeRelay(i int) routing.Relay {

	firstIpPart := i / 255
	secondIpPart := i % 255
	IP := fmt.Sprintf("100.0.%v.%v:40000", firstIpPart, secondIpPart)
	addr, _ := net.ResolveUDPAddr("UDP", IP)

	id := crypto.HashID(IP)

	return routing.Relay{
		Name: fmt.Sprintf("fake_relay_%v", i),
		ID:   id,
		Addr: *addr,
	}
}

type Relay struct {
	data         routing.Relay
	state        int
	stateChanged time.Time
	routeBaseMap map[uint64]routeBase
}

type routeBase struct {
	rtt        float32
	jitter     float32
	packetLoss float32
}

func newRouteBase() routeBase {
	rb := new(routeBase)
	rb.rtt = float32(rand.Int31n(maxRTT))
	rb.jitter = float32(rand.Int31n(maxJitter))
	rb.packetLoss = 0.0

	return *rb
}

func sendInit(relay Relay, addr string) error {
	//todo needs public/private key, one key for all
	initRequest := transport.RelayInitRequest{
		Magic:          transport.InitRequestMagic,
		Version:        transport.VersionNumberInitRequest,
		Nonce:          []byte{}, //todo
		Address:        relay.data.Addr,
		EncryptedToken: []byte{}, //todo
		RelayVersion:   "0.0.0",
	}

	initBinary, err := initRequest.MarshalBinary()
	if err != nil {
		return err
	}
	buffer := bytes.NewBuffer(initBinary)
	resp, err := http.Post(addr, "application/octet-stream", buffer)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("response was non 200: %v", resp.StatusCode)
	}

	return nil
}

func sendUpdateInit(relay Relay, addr string) ([]uint64, error) {
	updateRequest := baseUpdate(relay)
	updateBinary, err := updateRequest.MarshalBinary()
	if err != nil {
		return []uint64{}, err
	}

	buffer := bytes.NewBuffer(updateBinary)
	resp, err := http.Post(addr, "application/octet-stream", buffer)
	if err != nil {
		return []uint64{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return []uint64{}, fmt.Errorf("response was non 200: %v", resp.StatusCode)
	}

	// WIP return relaysToPingFromUpdateResponse(resp)
	return []uint64{}, nil
}

func sendShutdown(relay Relay, addr string) error {
	updateRequest := baseUpdate(relay)
	updateRequest.ShuttingDown = true
	updateBinary, err := updateRequest.MarshalBinary()
	if err != nil {
		return err
	}

	buffer := bytes.NewBuffer(updateBinary)
	resp, err := http.Post(addr, "application/octet-stream", buffer)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("shutdown response was non 200: %v", resp.StatusCode)
	}

	return nil
}

func sendUpdate(relay Relay, relaysToPing []uint64, addr string) ([]uint64, error) {
	updateRequest := baseUpdate(relay)

	numRelays := len(relaysToPing)
	statsData := make([]routing.RelayStatsPing, len(relaysToPing))
	for i := 0; i < numRelays; i++ {
		if base, ok := relay.routeBaseMap[relaysToPing[i]]; ok {
			statsData[i] = newPacketData(relaysToPing[i], base)
		}
	}
	updateRequest.PingStats = statsData
	updateBinary, err := updateRequest.MarshalBinary()
	if err != nil {
		return []uint64{}, err
	}

	buffer := bytes.NewBuffer(updateBinary)
	resp, err := http.Post(addr, "application/octet-stream", buffer)
	if err != nil {
		return []uint64{}, err
	}

	if resp.StatusCode != http.StatusOK {
		return []uint64{}, fmt.Errorf("shutdown response was non 200: %v", resp.StatusCode)
	}

	// WIP return relaysToPingFromUpdateResponse(resp)
	return relaysToPing, nil
}

func relaysToPingFromRelayList(relayArr []Relay, skipID uint64) []uint64 {
	relays := make([]uint64, 0)
	for _, relay := range relayArr {
		if relay.data.ID == skipID {
			continue
		}

		if relay.state != relayEnabled {
			continue
		}
		relays = append(relays, relay.data.ID)
	}

	return relays

}

func relaysToPingFromUpdateResponse(resp *http.Response) ([]uint64, error) {

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []uint64{}, err
	}
	defer resp.Body.Close()

	var relayUpdateResponse transport.RelayUpdateResponse
	switch resp.Header.Get("Content-Type") {
	case "application/octet-stream":
		err = relayUpdateResponse.UnmarshalBinary(body)
	default:
		err = fmt.Errorf("unsupported content type")
	}
	if err != nil {
		return []uint64{}, err
	}

	relaysToPing := make([]uint64, len(relayUpdateResponse.RelaysToPing))
	for i, relay := range relayUpdateResponse.RelaysToPing {
		if relay.ID != 0 {
			relaysToPing[i] = relay.ID
		}
	}

	return relaysToPing, nil
}

func baseUpdate(relay Relay) transport.RelayUpdateRequest {
	return transport.RelayUpdateRequest{
		Version:      2,
		RelayVersion: "2",
		Address:      relay.data.Addr,
		Token:        relay.data.PublicKey,
	}
}

func newPacketData(id uint64, base routeBase) routing.RelayStatsPing {
	pingStat := routing.RelayStatsPing{}
	pingStat.RelayID = id

	rttMultiplier := calcMultiplier()
	pingStat.RTT = base.rtt * rttMultiplier

	jitterMultiplier := calcMultiplier()
	pingStat.Jitter = base.jitter * jitterMultiplier

	hasPL := rand.Int31n(pLChance)
	if hasPL == 1 {
		pingStat.PacketLoss = pLValue
	} else {
		pingStat.PacketLoss = 0.0
	}

	return pingStat
}

// this returns the float multiplier at +/- maxMultiplierPercent
func calcMultiplier() float32 {
	base := rand.Int31n(maxMultiplierPercent * 2)
	return 1.0 + float32(base-maxMultiplierPercent)/100.0

}
