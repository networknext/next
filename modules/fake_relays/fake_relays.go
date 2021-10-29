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

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/transport"
)

const (
	MaxRTT               = 300
	MaxJitter            = 10
	MaxPacketLoss        = 30
	MaxMultiplierPercent = 10

	// Chances are 1 in N
	PLChance = 100
)

// FakeRelay represents a single fake relay that simulates a real relay.
// It mocks relay_update and pinging near relays to load test the relay backend.
// It does NOT actually carry session traffic, and thus should only be used
// with the Fake Server for load testing.
type FakeRelay struct {
	data            routing.Relay
	state           routing.RelayState
	routeBaseMap    map[uint64]RouteBase
	relaysToPing    []routing.RelayPingData
	backendHostname string
	updateVersion   int
	relayMetrics    *metrics.FakeRelayMetrics
}

// RouteBase represents the ping stats between a pair of Fake Relays.
type RouteBase struct {
	rtt        float32
	jitter     float32
	packetLoss float32
}

// NewFakeRelays() creates numRelays fake relays for load testing
func NewFakeRelays(numRelays int, relayPublicKey []byte, gatewayAddr string, updateVersion int, fakeRelayMetrics *metrics.FakeRelayMetrics) ([]*FakeRelay, error) {
	relayArr := make([]*FakeRelay, numRelays)
	// Fill up the relay array with fake relays
	for i := 0; i < numRelays; i++ {
		routingRelay, err := newRoutingRelay(i, relayPublicKey)
		if err != nil {
			fakeRelayMetrics.ErrorMetrics.ResolveUDPAddressError.Add(1)
			return nil, err
		}

		relayArr[i] = &FakeRelay{
			data:            routingRelay,
			state:           routing.RelayStateDisabled,
			routeBaseMap:    make(map[uint64]RouteBase),
			backendHostname: gatewayAddr,
			updateVersion:   updateVersion,
			relayMetrics:    fakeRelayMetrics,
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

// StartLoop() enables a relay to send relay updates to the Relay Gateway
// until a shutdown signal is received.
func (relay *FakeRelay) StartLoop(ctx context.Context, wg *sync.WaitGroup) {
	defer wg.Done()

	// Fake Relays update once a second
	ticker := time.NewTicker(time.Second)

	var err error
	for {
		relay.relayMetrics.UpdateInvocations.Add(1)

		select {
		case <-ctx.Done():
			// Shutdown signal received
			err = relay.sendShutdownRequest()
			if err != nil {
				core.Error("relay: failed to send shutdown request: %v", err)
			} else {
				relay.relayMetrics.SuccessfulUpdateInvocations.Add(1)
			}
			relay.state = routing.RelayStateMaintenance
			return
		case <-ticker.C:
			if relay.state == routing.RelayStateDisabled {
				// Send initial update request if Fake Relay is disabled
				err = relay.sendInitialUpdateRequest()
				if err != nil {
					core.Error("relay: failed to send initial update request: %v", err)
					continue
				}
				// Change relay state to enabled
				relay.state = routing.RelayStateEnabled
				relay.relayMetrics.SuccessfulUpdateInvocations.Add(1)
			} else if relay.state == routing.RelayStateEnabled {
				// Send standard update request
				err = relay.sendStandardUpdateRequest()
				if err != nil {
					core.Error("relay: failed to send standard update request: %v", err)
					continue
				}
				relay.relayMetrics.SuccessfulUpdateInvocations.Add(1)
			}
		}
	}
}

func (relay *FakeRelay) sendShutdownRequest() error {
	shutdownReq := relay.baseUpdate()
	shutdownReq.ShuttingDown = true
	shutdownReqBin, err := shutdownReq.MarshalBinary()
	if err != nil {
		relay.relayMetrics.ErrorMetrics.MarshalBinaryError.Add(1)
		return fmt.Errorf("error marshaling shutdown request: %v", err)
	}

	buffer := bytes.NewBuffer(shutdownReqBin)
	addr := fmt.Sprintf("http://%s/relay_update", relay.backendHostname)
	resp, err := http.Post(addr, "application/octet-stream", buffer)
	if err != nil {
		relay.relayMetrics.ErrorMetrics.UpdatePostError.Add(1)
		return err
	}

	if resp.StatusCode != http.StatusOK {
		relay.relayMetrics.ErrorMetrics.NotOKResponseError.Add(1)
		return fmt.Errorf("shutdown response was non 200: %v", resp.StatusCode)
	}

	return nil
}

func (relay *FakeRelay) sendInitialUpdateRequest() error {
	// Get the inital update request
	updateReq := relay.baseUpdate()
	updateReqBin, err := updateReq.MarshalBinary()
	if err != nil {
		relay.relayMetrics.ErrorMetrics.MarshalBinaryError.Add(1)
		return fmt.Errorf("error marshaling update request: %v", err)
	}

	return relay.sendUpdateRequest(updateReqBin)
}

func (relay *FakeRelay) sendStandardUpdateRequest() error {
	// Get the update request
	updateReq := relay.baseUpdate()

	// Simulate relay pings
	updateReq.PingStats = relay.simulateRelayPings()
	updateReqBin, err := updateReq.MarshalBinary()
	if err != nil {
		relay.relayMetrics.ErrorMetrics.MarshalBinaryError.Add(1)
		return fmt.Errorf("error marshaling update request: %v", err)
	}

	return relay.sendUpdateRequest(updateReqBin)
}

func (relay *FakeRelay) sendUpdateRequest(updateRequestBin []byte) error {
	// POST the update to the Relay Gateway
	buffer := bytes.NewBuffer(updateRequestBin)
	addr := fmt.Sprintf("http://%s/relay_update", relay.backendHostname)
	resp, err := http.Post(addr, "application/octet-stream", buffer)
	if err != nil {
		relay.relayMetrics.ErrorMetrics.UpdatePostError.Add(1)
		return err
	}

	// Read the response to get the relays to ping
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		relay.relayMetrics.ErrorMetrics.UpdatePostError.Add(1)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		relay.relayMetrics.ErrorMetrics.NotOKResponseError.Add(1)
		return fmt.Errorf("response was non 200: %v", resp.StatusCode)
	}

	updateResponse := &transport.RelayUpdateResponse{}
	err = updateResponse.UnmarshalBinary(body)
	if err != nil {
		relay.relayMetrics.ErrorMetrics.UnmarshalBinaryError.Add(1)
		return fmt.Errorf("error unmarshaling update response: %v", err)
	}

	// Set the relays to ping
	relay.relaysToPing = updateResponse.RelaysToPing

	return nil
}

// Simulates relay pings between this relay and all other Fake Relays based on probabilities.
func (relay *FakeRelay) simulateRelayPings() []routing.RelayStatsPing {
	numRelays := len(relay.relaysToPing)
	statsData := make([]routing.RelayStatsPing, numRelays)

	for i := 0; i < numRelays; i++ {
		if base, ok := relay.routeBaseMap[relay.relaysToPing[i].ID]; ok {
			statsData[i] = relay.newRelayPingStats(relay.relaysToPing[i].ID, base)
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

	if rand.Int31n(PLChance) == 1 {
		pingStat.PacketLoss = float32(rand.Int31n(MaxPacketLoss)) / 100.0
	} else {
		pingStat.PacketLoss = float32(0)
	}

	return pingStat
}

func (relay *FakeRelay) baseUpdate() transport.RelayUpdateRequest {
	req := transport.RelayUpdateRequest{
		Version:      uint32(relay.updateVersion),
		RelayVersion: "2.0.8",
		Address:      relay.data.Addr,
	}

	if relay.updateVersion >= 3 {
		req.Token = relay.data.PublicKey
	}

	return req
}

func newRoutingRelay(index int, relayPublicKey []byte) (routing.Relay, error) {
	ipAddress := fmt.Sprintf("127.0.0.1:%d", 10000+index)
	udpAddr, err := net.ResolveUDPAddr("udp", ipAddress)
	if err != nil {
		return routing.Relay{}, fmt.Errorf("error resolving UDP address: %v\n", err)
	}

	return routing.Relay{
		Name:      fmt.Sprintf("staging.relay.%d", index+1),
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

// Returns the float multiplier at +/- MaxMultiplierPercent
func calcMultiplier() float32 {
	base := rand.Int31n(MaxMultiplierPercent * 2)
	return 1.0 + float32(base-MaxMultiplierPercent)/100.0
}
