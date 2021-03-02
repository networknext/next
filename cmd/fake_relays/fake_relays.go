package main

import (
	"bytes"
	"context"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"os"
	"time"

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
	maxRTT    = 300
	maxJitter = 30

	relayDisabled = 1
	relayEnabled  = 2
	relayCrashed  = 3

	relayCrashValue      = 10000
	relayCrashTimeout    = 60
	relayShutdownValue   = 1000
	relayShutdownTimeout = 30
)

func main() {
	os.Exit(mainReturnWithCode())
}

func mainReturnWithCode() int {
	//todo WIP
	//Setup-------------------------------------------------------------------------------------------------------------
	serviceName := "fake_relays"
	fmt.Printf("%s: Git Hash: %s - Commit: %s\n", serviceName, sha, commitMessage)

	ctx := context.Background()

	gcpProjectID := backend.GetGCPProjectID()

	logger, err := backend.GetLogger(ctx, gcpProjectID, serviceName)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	//numRelays to fake
	numRelays, err := envvar.GetInt("NUM_FAKE_RELAYS", 5)
	if err != nil {
		level.Error(logger).Log("err", err)
	}

	//time the load test runs for before calling shutdown to the relay backend.
	timeToRun, err := envvar.GetDuration("TIME_TO_RUN", 20*time.Minute)
	if err != nil {
		level.Error(logger).Log("err", err)
	}

	FeatureNoInit, err := envvar.GetBool("FEATURE_NO_INIT", false)
	if err != nil {
		level.Error(logger).Log("err", err)
	}

	//get and verify relayBackendAddr
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

	//get relay info from firestore
	//todo get "real" fake relays from firestore, create them first
	storageRelayArr := make([]routing.Relay, numRelays)
	for i := 0; i < numRelays; i++ {
		storageRelayArr[i] = tempFakeRelay(i)
	}

	//create fake relay route Bases
	relayArr := make([]Relay, len(storageRelayArr))
	for i := 0; i < numRelays; i++ {
		relayI := storageRelayArr[i]
		newRelay := Relay{
			data:         relayI,
			state:        relayDisabled,
			stateChanged: time.Now(),
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

	// core logic ----------------------------------------------------------------------------------------------------------
	shutdown := false
	initAddress := fmt.Sprintf("%s/relay_init", relayBackendAddr)
	updateAddress := fmt.Sprintf("%s/relay_update", relayBackendAddr)

	//send first update/init
	//first update is different because it has no packet data
	for i := 0; i < numRelays; i++ {

		relay := relayArr[i]

		go func(relay Relay) {

			var relaysToPing []uint64
			stateChangeTicker := 0
			syncTimer := helpers.NewSyncTimer(1 * time.Second)
			for {
				syncTimer.Run()

				if shutdown {
					err := sendShutdown(relay, updateAddress)
					level.Error(logger).Log("err", err)
					return
				}

				if stateChangeTicker > 0 {
					stateChangeTicker--
					continue
				}

				if stateChangeTicker == 0 {
					if FeatureNoInit {
						newRelaysToPing, err := sendUpdateInit(relay, updateAddress)
						if err != nil {
							level.Error(logger).Log("err", err)
							continue
						}
						relaysToPing = newRelaysToPing

					} else {
						newRelaysToPing, err := sendInit(relay, initAddress)
						if err != nil {
							level.Error(logger).Log("err", err)
							continue
						}
						relaysToPing = newRelaysToPing
					}
					stateChangeTicker = -1
				}

				//determine state change
				isCrashed := rand.Int31n(relayCrashValue)
				if isCrashed == 1 {
					stateChangeTicker = relayCrashTimeout
					continue
				}

				isShutdown := rand.Int31n(relayShutdownValue)
				if isShutdown == 1 {
					stateChangeTicker = relayShutdownTimeout
					err := sendShutdown(relay, updateAddress)
					if err != nil {
						level.Error(logger).Log("err", err)
					}
					continue
				}
				//if no state change populate packet update

				newRelaysToPing, err := sendUpdate(relay, relaysToPing, updateAddress)
				if err != nil {
					level.Error(logger).Log("err", err)
				}
				relaysToPing = newRelaysToPing
			}

		}(relay)
	}

	//shutdown and close application
	go func() {

		// todo system shutdowns / early shutdown
	}()

	time.Sleep(timeToRun)
	shutdown = true

	time.Sleep(1 * time.Minute)
	return 0
}

func tempFakeRelay(i int) routing.Relay {
	return routing.Relay{
		Name: fmt.Sprintf("fake_relay_%v", i),
		ID:   uint64(i),
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

func sendInit(relay Relay, addr string) ([]uint64, error) {
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
		return []uint64{}, err
	}
	buffer := bytes.NewBuffer(initBinary)
	resp, err := http.Post(addr, "application/octet-stream", buffer)

	//todo handle response

}

func sendUpdateInit(relay Relay, addr string) ([]uint64, error) {
	//todo finish base
	updateRequest := baseUpdate(relay)
	updateBinary, err := updateRequest.MarshalBinary()
	if err != nil {
		return []uint64{}, err
	}

	buffer := bytes.NewBuffer(updateBinary)
	resp, err := http.Post(addr, "application/octet-stream", buffer)

	//todo handle response get relays to ping

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
	//todo handle resp

	return nil
}

func sendUpdate(relay Relay, relaysToPing []uint64, addr string) ([]uint64, error) {
	updateRequest := baseUpdate(relay)

	numRelays := len(relaysToPing)
	statsData := make([]routing.RelayStatsPing, len(relaysToPing))
	for i := 0; i < numRelays; i++ {
		if base, ok := relay.routeBaseMap[relaysToPing[i]]; ok {
			statsData[i] = newPacketData(base)
		}
	}
	updateRequest.PingStats = statsData
	updateBinary, err := updateRequest.MarshalBinary()
	if err != nil {
		return []uint64{}, err
	}

	buffer := bytes.NewBuffer(updateBinary)
	resp, err := http.Post(addr, "application/octet-stream", buffer)

	//todo handle response get relays to ping

	return []uint64{}, nil
}

func baseUpdate(relay Relay) transport.RelayUpdateRequest {
	return transport.RelayUpdateRequest{
		Version:      2,
		RelayVersion: "2",
		Address:      relay.data.Addr,
		Token:        relay.data.PublicKey,
	}
}

func newPacketData(base routeBase) routing.RelayStatsPing {
	//todo handle packetData
	return routing.RelayStatsPing{}
}
