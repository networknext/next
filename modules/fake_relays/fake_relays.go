package fake_relays

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/networknext/backend/modules/common/helpers"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/transport"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
)

const (
	MaxRTT               = 300
	MaxJitter            = 10
	MaxMultiplierPercent = 10

	// RelayDisabled = 1
	// RelayEnabled  = 2
	// RelayShutdown = 3

	// Chances are 1 in N
	PLChance = 100
	PLValue  = 0.3
)

// FakeRelay represents a single fake relay that simulates a real relay.
// It mocks relay_update and pinging near relays to load test the relay backend.
// It does NOT actually carry session traffic, and thus should only be used
// with the Fake Server for load testing.
type FakeRelay struct {
	data            routing.Relay
	state           routing.RelayState
	stateChanged    time.Time
	routeBaseMap    map[uint64]RouteBase
	backendHostname string
	updateVersion   int
	logger          log.Logger
}

// RouteBase represents the ping stats between a pair of Fake Relays.
type RouteBase struct {
	rtt        float32
	jitter     float32
	packetLoss float32
}

func NewFakeRelays(numRelays int, relayPublicKey []byte, gatewayAddr string, updateVersion int, logger log.Logger) ([]*FakeRelay, error) {
	relayArr := make([]*FakeRelay, numRelays)
	// Fill up the relay array with fake relays
	for i := 0; i < numRelays; i++ {
		routingRelay, err := newRoutingRelay(i, relayPublicKey)
		if err != nil {
			return relayArr, err
		}

		relayArr[i] = &FakeRelay{
			data:            routingRelay,
			state:           routing.RelayStateDisabled,
			stateChanged:    time.Now().Add(-5 * time.Minute),
			routeBaseMap:    make(map[uint64]RouteBase),
			backendHostname: gatewayAddr,
			updateVersion:   updateVersion,
			logger:          logger,
		}
	}
	// Create route bases for each fake relay
	for i := 0; i < numRelays; i++ {
		for j := 0; j < numRelays; j++ {
			if i == j {
				continue
			}
			relayArr[i].routeBaseMap[relayArr[j].data.ID] = newRouteBase()
		}
	}

	return relayArr, nil
}

func (relay *FakeRelay) StartLoop(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	// Fake Relays update once a second
	syncTimer := helpers.NewSyncTimer(1 * time.Second)

	// Retain a list of fake relays to ping
	var relaysToPing []routing.RelayPingData
	var err error
	for {
		syncTimer.Run()

		select {
		case <-ctx.Done():
			// Shutdown signal received
			err = relay.sendShutdownRequest()
			if err != nil {
				level.Error(relay.logger).Log("err", err)
			}
			relay.state = routing.RelayStateMaintenance
			return
		default:
			if relay.state == routing.RelayStateDisabled {
				// Send initial update request if Fake Relay is disabled
				relaysToPing, err = relay.sendInitialUpdateRequest()
				if err != nil {
					level.Error(relay.logger).Log("err", err)
					continue
				}
				// Change relay state to enabled
				relay.state = routing.RelayStateEnabled
			} else if relay.state == routing.RelayStateEnabled {
				// Send standard update request
				relaysToPing, err = relay.sendStandardUpdateRequest(relaysToPing)
				if err != nil {
					level.Error(relay.logger).Log("err", err)
					continue
				}
			}
		}
	}
}

func (relay *FakeRelay) sendShutdownRequest() error {
	shutdownReq := relay.baseUpdate()
	shutdownReq.ShuttingDown = true
	shutdownReqBin, err := shutdownReq.MarshalBinary()
	if err != nil {
		return fmt.Errorf("error marshaling shutdown request: %v", err)
	}

	buffer := bytes.NewBuffer(shutdownReqBin)
	addr := fmt.Sprintf("http://%s/relay_update", relay.backendHostname)
	resp, err := http.Post(addr, "application/octet-stream", buffer)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("shutdown response was non 200: %v", resp.StatusCode)
	}

	return nil
}

func (relay *FakeRelay) sendInitialUpdateRequest() ([]routing.RelayPingData, error) {
	// Get the inital update request
	updateReq := relay.baseUpdate()
	updateReqBin, err := updateReq.MarshalBinary()
	if err != nil {
		return []routing.RelayPingData{}, fmt.Errorf("error marshaling update request: %v", err)
	}

	return relay.sendUpdateRequest(updateReqBin)
}

func (relay *FakeRelay) sendStandardUpdateRequest(relaysToPing []routing.RelayPingData) ([]routing.RelayPingData, error) {
	// Get the inital update request
	updateReq := relay.baseUpdate()

	// Simulate relay pings
	updateReq.PingStats = relay.simulateRelayPings(relaysToPing)
	updateReqBin, err := updateReq.MarshalBinary()
	if err != nil {
		return []routing.RelayPingData{}, fmt.Errorf("error marshaling update request: %v", err)
	}

	return relay.sendUpdateRequest(updateReqBin)
}

func (relay *FakeRelay) sendUpdateRequest(updateRequestBin []byte) ([]routing.RelayPingData, error) {
	// POST the update to the Relay Gateway
	buffer := bytes.NewBuffer(updateRequestBin)
	addr := fmt.Sprintf("http://%s/relay_update", relay.backendHostname)
	resp, err := http.Post(addr, "application/octet-stream", buffer)
	if err != nil {
		return []routing.RelayPingData{}, err
	}

	// Read the response to get the relays to ping
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return []routing.RelayPingData{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return []routing.RelayPingData{}, fmt.Errorf("response was non 200: %v", resp.StatusCode)
	}

	updateResponse := &transport.RelayUpdateResponse{}
	err = updateResponse.UnmarshalBinary(body)
	if err != nil {
		return []routing.RelayPingData{}, fmt.Errorf("error unmarshaling update response: %v", err)
	}

	return updateResponse.RelaysToPing, nil
}

func (relay *FakeRelay) simulateRelayPings(relaysToPing []routing.RelayPingData) []routing.RelayStatsPing {
	numRelays := len(relaysToPing)
	statsData := make([]routing.RelayStatsPing, numRelays)

	for i := 0; i < numRelays; i++ {
		if base, ok := relay.routeBaseMap[relaysToPing[i].ID]; ok {
			statsData[i] = relay.newRelayPingStats(relaysToPing[i].ID, base)
		}
	}

	return statsData
}

func (relay *FakeRelay) newRelayPingStats(id uint64, base RouteBase) routing.RelayStatsPing {
	pingStat := routing.RelayStatsPing{}
	pingStat.RelayID = id

	rttMultiplier := calcMultiplier()
	pingStat.RTT = base.rtt * rttMultiplier

	jitterMultiplier := calcMultiplier()
	pingStat.Jitter = base.jitter * jitterMultiplier

	hasPL := rand.Int31n(PLChance)
	if hasPL == 1 {
		pingStat.PacketLoss = PLValue
	} else {
		pingStat.PacketLoss = float32(0)
	}

	return pingStat
}

func (relay *FakeRelay) baseUpdate() transport.RelayUpdateRequest {

	req := transport.RelayUpdateRequest{
		Version:      uint32(relay.updateVersion),
		RelayVersion: "2.0.6", // TODO: change to constant in modules/tranpsort/packet_relay
		Address:      relay.data.Addr,
	}

	if relay.updateVersion >= 2 {
		req.Token = relay.data.PublicKey
	}

	return req
}

func newRoutingRelay(index int, relayPublicKey []byte) (routing.Relay, error) {
	// firstIpPart := i / 255
	// secondIpPart := i % 255
	// IP := fmt.Sprintf("100.0.%v.%v:40000", firstIpPart, secondIpPart)
	// TODO: fix local hack
	ipAddress := fmt.Sprintf("127.0.0.1:%d", 10000+index)
	udpAddr, err := net.ResolveUDPAddr("udp", ipAddress)
	if err != nil {
		return routing.Relay{}, fmt.Errorf("error creating IP: %v\n", err)
	}

	return routing.Relay{
		Name:      fmt.Sprintf("fake_relay_%d", index),
		ID:        crypto.HashID(ipAddress),
		Addr:      *udpAddr,
		PublicKey: relayPublicKey,
	}, nil
}

func newRouteBase() RouteBase {
	return RouteBase{
		rtt:        float32(rand.Int31n(MaxRTT)),
		jitter:     float32(rand.Int31n(MaxJitter)),
		packetLoss: float32(0),
	}
}

// this returns the float multiplier at +/- maxMultiplierPercent
func calcMultiplier() float32 {
	base := rand.Int31n(MaxMultiplierPercent * 2)
	return 1.0 + float32(base-MaxMultiplierPercent)/100.0

}
