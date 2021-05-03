package main

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
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
	buildtime               string
	commitMessage           string
	sha                     string
	tag                     string
	RELAY_PUBLIC_KEY        []byte
	RELAY_PRIVATE_KEY       []byte
	RELAY_ROUTER_PUBLIC_KEY []byte
	relayUpdateVersion      int
)

const (
	maxRTT               = 300
	maxJitter            = 10
	maxMultiplierPercent = 10

	// pLChance is 1 in n
	relayDisabled = 1
	relayEnabled  = 2
	relayShutdown = 3

	// chances are 1 in n
	pLChance = 100000
	pLValue  = .3
)

func main() {
	os.Exit(mainReturnWithCode())
}

func mainReturnWithCode() int {

	// todo metrics??
	RELAY_PUBLIC_KEY, _ = base64.StdEncoding.DecodeString("8hUCRvzKh2aknL9RErM/Vj22+FGJW0tWMRz5KlHKryE=")
	RELAY_PRIVATE_KEY, _ = base64.StdEncoding.DecodeString("ZiCSchVFo6T5gJvbQfcwU7yfELsNJaYIC2laQm9DSuA=")
	RELAY_ROUTER_PUBLIC_KEY, _ = base64.StdEncoding.DecodeString("SS55dEl9nTSnVVDrqwPeqRv/YcYOZZLXCWTpNBIyX0Y=")

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

	featureNoInit, err := envvar.GetBool("FEATURE_NO_INIT", true)
	if err != nil {
		level.Error(logger).Log("err", err)
	}

	// get and verify relayGatewayAddr
	var relayGatewayAddr string
	if gcpProjectID != "" {
		relayGatewayAddr= envvar.Get("RELAY_GATEWAY_LB_IP", "")
		if net.ParseIP(relayGatewayAddr) == nil {
			level.Error(logger).Log("err", err)
			return 1
		}
	} else {
		relayGatewayAddr = "127.0.0.1:30000"
	}

	relayUpdateVersion, err = envvar.GetInt("RELAY_UPDATE_VERSION", 2)
	if err != nil {
		level.Error(logger).Log("err", err)
		return 1
	}

	// create fake relays
	storageRelayArr := make([]routing.Relay, numRelays)
	for i := 0; i < numRelays; i++ {
		storageRelayArr[i] = fakeRelay(i)
	}

	// create fake relay route Bases
	relayArr := make([]*Relay, len(storageRelayArr))
	for i := 0; i < numRelays; i++ {
		relayI := storageRelayArr[i]
		newRelay := &Relay{
			data:         relayI,
			state:        relayDisabled,
			stateChanged: time.Now().Add(-5 * time.Minute),
			RouteBaseMap: make(map[uint64]RouteBase),
		}

		for j := 0; j < numRelays; j++ {
			if i == j {
				continue
			}

			relayJ := storageRelayArr[j]

			base := newRouteBase()
			newRelay.RouteBaseMap[relayJ.ID] = base
		}

		relayArr[i] = newRelay
	}

	fmt.Printf("starting main logic\n")

	// core logic
	shutdown := false
	// initAddress := fmt.Sprintf("http://%s/relay_init", relayBackendAddr)
	updateAddress := fmt.Sprintf("http://%s/relay_update", relayGatewayAddr)

	for i := 0; i < numRelays; i++ {
		go func(relay *Relay, relayArr []*Relay) {

			var relaysToPing []uint64
			syncTimer := helpers.NewSyncTimer(1 * time.Second)
			for {
				syncTimer.Run()

				if shutdown {
					err := sendShutdown(*relay, updateAddress)
					if err != nil {
						level.Error(logger).Log("err", err)
						continue
					}
					return
				}

				if relay.state == relayDisabled {
					if featureNoInit {
						err := sendUpdateInit(*relay, updateAddress)
						if err != nil {
							level.Error(logger).Log("err", err)
							continue
						}
						relay.state = relayEnabled

					} else {
						level.Error(logger).Log("err", "turn on no init")
						// err := sendInit(*relay, initAddress)
						// if err != nil {
						// 	level.Error(logger).Log("err", err)
						// 	continue
						// }
						// relay.state = relayEnabled
					}
					continue
				}

				if relay.state == relayShutdown {
					err = sendShutdown(*relay, updateAddress)
					if err != nil {
						level.Error(logger).Log("err", err)
						continue
					}
					relay.state = relayDisabled
				}

				relaysToPing = relaysToPingFromRelayList(relayArr, relay.data.ID)

				_, err := sendUpdate(*relay, relaysToPing, updateAddress)
				if err != nil {
					level.Error(logger).Log("err", err)
					relay.state = relayShutdown
					continue
				}

			}

		}(relayArr[i], relayArr)
	}

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<-sigint
	shutdown = true
	time.Sleep(5 * time.Second)
	return 0
}

func fakeRelay(i int) routing.Relay {
	firstIpPart := i / 255
	secondIpPart := i % 255
	IP := fmt.Sprintf("100.0.%v.%v:40000", firstIpPart, secondIpPart)
	addr, _ := net.ResolveUDPAddr("udp", IP)

	id := crypto.HashID(IP)

	return routing.Relay{
		Name:      fmt.Sprintf("fake_relay_%v", i),
		ID:        id,
		Addr:      *addr,
		PublicKey: RELAY_PUBLIC_KEY,
	}
}

type Relay struct {
	data         routing.Relay
	state        int
	stateChanged time.Time
	RouteBaseMap map[uint64]RouteBase
}

type RouteBase struct {
	rtt        float32
	jitter     float32
	packetLoss float32
}

func newRouteBase() RouteBase {
	rb := new(RouteBase)
	rb.rtt = float32(rand.Int31n(maxRTT))
	rb.jitter = float32(rand.Int31n(maxJitter))
	rb.packetLoss = 0.0

	return *rb
}

// func sendInit(relay Relay, addr string) error {
// 	nonce, token := makeToken()

// 	initRequest := transport.RelayInitRequest{
// 		Magic:          transport.InitRequestMagic,
// 		Version:        transport.VersionNumberInitRequest,
// 		Nonce:          nonce,
// 		Address:        relay.data.Addr,
// 		EncryptedToken: token,
// 		RelayVersion:   "1.1.0",
// 	}

// 	initBinary, err := initRequest.MarshalBinary()
// 	if err != nil {
// 		return err
// 	}
// 	buffer := bytes.NewBuffer(initBinary)
// 	resp, err := http.Post(addr, "application/octet-stream", buffer)
// 	if err != nil {
// 		return err
// 	}

// 	if resp.StatusCode != http.StatusOK {
// 		return fmt.Errorf("response was non 200: %v", resp.StatusCode)
// 	}
// 	resp.Body.Close()
// 	return nil
// }

func sendUpdateInit(relay Relay, addr string) error {
	updateRequest := baseUpdate(relay)
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
		return fmt.Errorf("response was non 200: %v", resp.StatusCode)
	}
	resp.Body.Close()
	return nil
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
	resp.Body.Close()
	return nil
}

func sendUpdate(relay Relay, relaysToPing []uint64, addr string) ([]uint64, error) {
	updateRequest := baseUpdate(relay)

	numRelays := len(relaysToPing)
	statsData := make([]routing.RelayStatsPing, len(relaysToPing))
	for i := 0; i < numRelays; i++ {
		if base, ok := relay.RouteBaseMap[relaysToPing[i]]; ok {
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
	resp.Body.Close()

	return relaysToPing, nil
}

func relaysToPingFromRelayList(relayArr []*Relay, skipID uint64) []uint64 {
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

func baseUpdate(relay Relay) transport.RelayUpdateRequest {

	req := transport.RelayUpdateRequest{
		Version:      uint32(relayUpdateVersion),
		RelayVersion: "1.1.0",
		Address:      relay.data.Addr,
	}

	if relayUpdateVersion == 2 {
		req.Token = relay.data.PublicKey
	}

	return req
}

func makeToken() ([]byte, []byte) {
	nonce := []byte("123456781234567812345678")
	data := []byte("12345678123456781234567812345678")
	token := crypto.Seal(data, nonce, RELAY_ROUTER_PUBLIC_KEY, RELAY_PRIVATE_KEY)

	return nonce, token
}

func newPacketData(id uint64, base RouteBase) routing.RelayStatsPing {
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
