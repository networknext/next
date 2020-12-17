package transport_test

import (
	"bytes"
	"context"
	crand "crypto/rand"
	"errors"
	"fmt"
	"math/rand"
	"net"
	"os"
	"runtime"
	"testing"
	"time"

	"github.com/alicebob/miniredis"
	"github.com/go-kit/kit/log"
	"github.com/networknext/backend/modules/billing"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/storage"
	"github.com/networknext/backend/modules/transport"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/nacl/box"
)

const (
	noRTT          rttType = 0
	badRTT         rttType = 1
	goodRTT        rttType = 2
	overloadedRTT  rttType = 3
	slightlyBadRTT rttType = 4
)

type rttType int

type relayGroup struct {
	Count        int32
	IDs          []uint64
	Addresses    []net.UDPAddr
	RTTs         []int32
	Jitters      []int32
	PacketLosses []int32
}

func getRelays(t *testing.T, numRelays int32, rttType rttType, singleBadRTT bool) relayGroup {
	relays := relayGroup{
		Count:        numRelays,
		IDs:          make([]uint64, 0),
		Addresses:    make([]net.UDPAddr, 0),
		RTTs:         make([]int32, 0),
		Jitters:      make([]int32, 0),
		PacketLosses: make([]int32, 0),
	}

	var err error

	for i := int32(0); i < numRelays; i++ {
		var relayAddress *net.UDPAddr

		relayAddress, err = net.ResolveUDPAddr("udp", "127.0.0.1:4000"+fmt.Sprintf("%d", i))
		assert.NoError(t, err)

		relays.IDs = append(relays.IDs, uint64(i)+1)
		relays.Addresses = append(relays.Addresses, *relayAddress)

		if singleBadRTT && i == 1 {
			relays.RTTs = append(relays.RTTs, 255)
		} else {
			switch rttType {
			case noRTT:
				relays.RTTs = append(relays.RTTs, 0)

			case badRTT:
				relays.RTTs = append(relays.RTTs, int32(100+i*5))

			case goodRTT:
				relays.RTTs = append(relays.RTTs, int32(20+i*5))
			}
		}

		relays.Jitters = append(relays.Jitters, 0)
		relays.PacketLosses = append(relays.PacketLosses, 0)
	}

	return relays
}

func getRouteRelays(t *testing.T, numRelays int32, internal bool, badRoute bool, badNearRelay bool) relayGroup {
	relays := relayGroup{
		Count:        numRelays,
		IDs:          make([]uint64, 0),
		Addresses:    make([]net.UDPAddr, 0),
		RTTs:         make([]int32, 0),
		Jitters:      make([]int32, 0),
		PacketLosses: make([]int32, 0),
	}

	// Send back relays that won't exist to simulate relays not existing in the route matrix
	if badRoute {
		relayAddress1, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
		assert.NoError(t, err)

		relayAddress2, err := net.ResolveUDPAddr("udp", "127.0.0.1:40001")
		assert.NoError(t, err)

		relays.IDs = append(relays.IDs, 5, 7)
		relays.Addresses = append(relays.Addresses, *relayAddress1, *relayAddress2)
		relays.RTTs = append(relays.RTTs, 20, 25)
		relays.Jitters = append(relays.Jitters, 0, 0)
		relays.PacketLosses = append(relays.PacketLosses, 0, 0)

		return relays
	}

	var err error

	if badNearRelay {

		// If there is a bad near relay, we'll be missing relay index 1
		// so add in all relays > 1 by shifting the loop, then add in 0

		for i := numRelays; i > 1; i-- {
			var relayAddress *net.UDPAddr

			if internal && i < numRelays-1 {
				relayAddress, err = net.ResolveUDPAddr("udp", "10.128.0.1:4000"+fmt.Sprintf("%d", i))
				assert.NoError(t, err)
			} else {
				relayAddress, err = net.ResolveUDPAddr("udp", "127.0.0.1:4000"+fmt.Sprintf("%d", i))
				assert.NoError(t, err)
			}

			relays.IDs = append(relays.IDs, uint64(i+1))
			relays.Addresses = append(relays.Addresses, *relayAddress)

			relays.RTTs = append(relays.RTTs, int32(20+i*5))
			relays.Jitters = append(relays.Jitters, 0)
			relays.PacketLosses = append(relays.PacketLosses, 0)
		}

		relayAddress, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
		assert.NoError(t, err)

		relays.IDs = append(relays.IDs, 1)
		relays.Addresses = append(relays.Addresses, *relayAddress)

		relays.RTTs = append(relays.RTTs, 20)
		relays.Jitters = append(relays.Jitters, 0)
		relays.PacketLosses = append(relays.PacketLosses, 0)

		return relays
	}

	for i := numRelays - 1; i >= 0; i-- {
		var relayAddress *net.UDPAddr

		if internal && i < numRelays-1 {
			relayAddress, err = net.ResolveUDPAddr("udp", "10.128.0.1:4000"+fmt.Sprintf("%d", i))
			assert.NoError(t, err)
		} else {
			relayAddress, err = net.ResolveUDPAddr("udp", "127.0.0.1:4000"+fmt.Sprintf("%d", i))
			assert.NoError(t, err)
		}

		relays.IDs = append(relays.IDs, uint64(i+1))
		relays.Addresses = append(relays.Addresses, *relayAddress)

		relays.RTTs = append(relays.RTTs, int32(20+i*5))
		relays.Jitters = append(relays.Jitters, 0)
		relays.PacketLosses = append(relays.PacketLosses, 0)
	}

	return relays
}

func getStats(rttType rttType) routing.Stats {
	switch rttType {
	case badRTT:
		return routing.Stats{RTT: 100}

	case goodRTT:
		return routing.Stats{RTT: 15}

	case overloadedRTT:
		return routing.Stats{RTT: 500}

	case slightlyBadRTT:
		return routing.Stats{RTT: 16}
	}

	return routing.Stats{}
}

type badIPLocator struct{}

func (locator *badIPLocator) LocateIP(ip net.IP) (routing.Location, error) {
	return routing.LocationNullIsland, errors.New("bad location")
}

type goodIPLocator struct{}

func (locator *goodIPLocator) LocateIP(ip net.IP) (routing.Location, error) {
	return routing.LocationNullIsland, nil
}

type routeInfo struct {
	sessionVersion      uint32
	initial             bool
	routeRelayIDs       []uint64
	routeRelayAddresses []net.UDPAddr
	routeCost           int32
	next                bool
}

func getRouteInfo(
	t *testing.T,
	numNearRelays int32,
	nearRelayRTTType rttType,
	destRelay uint64,
	routeMatrix *routing.RouteMatrix,
	routeType int32,
	routeNumRelays int32,
	sessionVersion uint32,
	internal bool,
	badRouteRelays bool,
	badNearRelay bool,
) routeInfo {
	nearRelays := getRelays(t, numNearRelays, nearRelayRTTType, badNearRelay)
	routeRelays := getRouteRelays(t, routeNumRelays, internal, badRouteRelays, badNearRelay)

	nearRelayIndices := make([]int32, nearRelays.Count)
	var prevBadRelays bool
	var ok bool
	for i := 0; i < len(nearRelayIndices); i++ {
		nearRelayIndices[i], ok = routeMatrix.RelayIDsToIndices[nearRelays.IDs[i]]
		if !ok {
			prevBadRelays = true
			break
		}
	}

	routeRelayIndices := [core.MaxRelaysPerRoute]int32{}
	for i := int32(0); i < routeRelays.Count; i++ {
		routeRelayIndices[i], ok = routeMatrix.RelayIDsToIndices[routeRelays.IDs[i]]
		if !ok {
			prevBadRelays = true
			break
		}
	}

	destRelayIndex, ok := routeMatrix.RelayIDsToIndices[destRelay]
	if !ok {
		prevBadRelays = true
	}

	var routeCost int32

	if !prevBadRelays {
		destRelayIndices := append([]int32{}, destRelayIndex)

		if routeRelays.Count > 0 {
			routeCost = core.GetCurrentRouteCost(routeMatrix.RouteEntries, routeRelays.Count, routeRelayIndices, nearRelayIndices, nearRelays.RTTs, destRelayIndices, nil)
		}
	}

	switch routeType {
	case routing.RouteTypeNew:
		return routeInfo{
			sessionVersion:      sessionVersion + 1,
			initial:             true,
			routeRelayIDs:       routeRelays.IDs,
			routeRelayAddresses: routeRelays.Addresses,
			routeCost:           routeCost,
			next:                true,
		}

	case routing.RouteTypeContinue:
		return routeInfo{
			sessionVersion:      sessionVersion,
			initial:             false,
			routeRelayIDs:       routeRelays.IDs,
			routeRelayAddresses: routeRelays.Addresses,
			routeCost:           routeCost,
			next:                true,
		}
	}

	routeInfo := routeInfo{
		sessionVersion:      sessionVersion,
		initial:             false,
		routeRelayIDs:       nil,
		routeRelayAddresses: nil,
		routeCost:           0,
		next:                false,
	}

	return routeInfo
}

func getBlankSessionUpdateMetrics(t *testing.T) *metrics.SessionUpdateMetrics {
	metricsHandler := metrics.LocalHandler{}
	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)

	return metrics.SessionUpdateMetrics
}

func getInitialSessionData(
	t *testing.T,
	startTime time.Time,
	request *sessionUpdateRequestConfig,
	routeInfo routeInfo,
	response *sessionUpdateResponseConfig,
	committed bool,
	commitCounter int32,
	multipath bool,
) *transport.SessionData {
	var sessionData *transport.SessionData

	if request.badSessionDataSessionID {
		sessionData = &transport.SessionData{}
	} else if request.badSessionDataSliceNumber {
		sessionData = &transport.SessionData{
			SessionID: request.sessionID,
		}
	} else if request.sliceNumber > 1 {
		prevRequest := *request
		prevRequest.sliceNumber -= 1
		sd := getExpectedSessionData(t, transport.SessionData{}, uint64(startTime.Unix()), &prevRequest, routeInfo, response, false, false, false, false, committed, commitCounter, multipath)
		sessionData = &sd
	} else if request.sliceNumber > 0 {
		sessionData = &transport.SessionData{
			SessionID:       request.sessionID,
			SliceNumber:     request.sliceNumber,
			Location:        routing.LocationNullIsland,
			ExpireTimestamp: uint64(startTime.Unix()),
			RouteState: core.RouteState{
				UserID: request.userHash,
			},
		}
	}

	return sessionData
}

func getExpectedSessionData(
	t *testing.T,
	initialSessionData transport.SessionData,
	expireTimestamp uint64,
	request *sessionUpdateRequestConfig,
	routeInfo routeInfo,
	response *sessionUpdateResponseConfig,
	noRoute bool,
	multipathOverload bool,
	latencyWorse bool,
	commitVeto bool,
	committed bool,
	commitCounter int32,
	multipath bool,
) transport.SessionData {
	if response.unchangedSessionData {
		initialSessionData.Version = transport.SessionDataVersion
		return initialSessionData
	}

	if request.badSessionData || request.badSessionDataSessionID {
		return transport.SessionData{
			Version: transport.SessionDataVersion,
		}
	}

	if request.badSessionDataSliceNumber {
		return transport.SessionData{
			Version:   transport.SessionDataVersion,
			SessionID: request.sessionID,
		}
	}

	nearRelays := getRelays(t, response.numNearRelays, request.nearRelayRTTType, request.badNearRelay)

	var everOnNext bool
	var reduceLatency bool
	if request.prevRouteType != routing.RouteTypeDirect || response.routeType != routing.RouteTypeDirect {
		// These fields are "sticky" - they will remain true as long as we take network next once
		everOnNext = true
		reduceLatency = true
	}

	sessionData := transport.SessionData{
		Version:         transport.SessionDataVersion,
		SessionID:       request.sessionID,
		SessionVersion:  routeInfo.sessionVersion,
		SliceNumber:     request.sliceNumber + 1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: expireTimestamp,
		RouteNumRelays:  int32(len(routeInfo.routeRelayIDs)),
		RouteCost:       routeInfo.routeCost,
		RouteState: core.RouteState{
			UserID:            request.userHash,
			NumNearRelays:     int32(nearRelays.Count),
			Next:              routeInfo.next,
			Veto:              noRoute || multipathOverload || latencyWorse || commitVeto,
			ReduceLatency:     reduceLatency,
			Committed:         committed,
			CommitCounter:     commitCounter,
			CommitVeto:        commitVeto,
			Multipath:         multipath,
			RelayWentAway:     request.badRouteRelays,
			RouteLost:         request.badNearRelay || request.badRouteRelays,
			NoRoute:           noRoute,
			MultipathOverload: multipathOverload,
			LatencyWorse:      latencyWorse,
		},
		Initial:          routeInfo.initial,
		FellBackToDirect: request.fallbackToDirect,
		EverOnNext:       everOnNext,
	}

	if response.attemptFindRoute {
		sessionData.RouteState.PLHistoryIndex = initialSessionData.RouteState.PLHistoryIndex + 1
		sessionData.RouteState.PLHistorySamples = initialSessionData.RouteState.PLHistorySamples + 1
	}

	copy(sessionData.RouteRelayIDs[:], routeInfo.routeRelayIDs)

	copy(sessionData.RouteState.NearRelayRTT[:], nearRelays.RTTs)
	copy(sessionData.RouteState.NearRelayJitter[:], nearRelays.Jitters)

	// Convert the near relay PL history from []int32 to []uint32
	nearRelayPLHistory := make([]uint32, nearRelays.Count)
	for i := 0; int32(i) < nearRelays.Count; i++ {
		nearRelayPLHistory[i] = uint32(nearRelays.PacketLosses[i])
	}

	copy(sessionData.RouteState.NearRelayPLHistory[:], nearRelayPLHistory)

	return sessionData
}

type sessionUpdateRequestConfig struct {
	sdkVersion                transport.SDKVersion
	sessionID                 uint64
	sliceNumber               uint32
	buyerID                   uint64
	datacenterID              uint64
	userHash                  uint64
	clientPingTimedOut        bool
	fallbackToDirect          bool
	numNearRelays             int32
	nearRelayRTTType          rttType
	directStats               routing.Stats
	nextStats                 routing.Stats
	prevRouteType             int32
	prevRouteNumRelays        int32
	badSessionData            bool // The request sends up bad session data (fails to unmarshal)
	badSessionDataSessionID   bool // The request sends up a mismatched session ID in the session data
	badSessionDataSliceNumber bool // The request sends up a mismatched slice number in the session data
	badRouteRelays            bool // The request sends up a route in the session data with relays that no longer exist
	badNearRelay              bool // The request sends up a near relay with unroutable RTT
}

func NewSessionUpdateRequestConfig(t *testing.T) *sessionUpdateRequestConfig {
	return &sessionUpdateRequestConfig{
		sdkVersion:                transport.SDKVersionMin,
		sessionID:                 1111,
		sliceNumber:               0,
		buyerID:                   123,
		datacenterID:              10,
		userHash:                  12345,
		clientPingTimedOut:        false,
		fallbackToDirect:          false,
		numNearRelays:             0,
		nearRelayRTTType:          noRTT,
		directStats:               getStats(noRTT),
		nextStats:                 getStats(noRTT),
		prevRouteType:             routing.RouteTypeDirect,
		prevRouteNumRelays:        0,
		badSessionData:            false,
		badSessionDataSessionID:   false,
		badSessionDataSliceNumber: false,
		badRouteRelays:            false,
	}
}

type sessionUpdateBackendConfig struct {
	ipLocator            routing.IPLocator
	numRouteMatrixRelays int32
	internalAddresses    bool
	buyer                *routing.Buyer
	datacenters          []routing.Datacenter
	datacenterMaps       []routing.DatacenterMap
}

func NewSessionUpdateBackendConfig(t *testing.T) *sessionUpdateBackendConfig {
	return &sessionUpdateBackendConfig{
		ipLocator:            &goodIPLocator{},
		numRouteMatrixRelays: 0,
		internalAddresses:    false,
		buyer:                nil,
		datacenters:          nil,
		datacenterMaps:       nil,
	}
}

type sessionUpdateResponseConfig struct {
	numNearRelays        int32
	unchangedSessionData bool // The backend should not have altered the session data
	attemptFindRoute     bool // Should the backend have even attempted to find a route (run through the core logic)?
	routeType            int32
	routeNumRelays       int32
	debug                bool // The backend should have sent down a debug string
}

func NewSessionUpdateResponseConfig(t *testing.T) *sessionUpdateResponseConfig {
	return &sessionUpdateResponseConfig{
		numNearRelays:        0,
		unchangedSessionData: false,
		attemptFindRoute:     false,
		routeType:            routing.RouteTypeDirect,
		routeNumRelays:       0,
		debug:                false,
	}
}

func runSessionUpdateTest(t *testing.T, request *sessionUpdateRequestConfig, backend *sessionUpdateBackendConfig, response *sessionUpdateResponseConfig, expectedMetrics *metrics.SessionUpdateMetrics) {

	// Set up request packet

	nearRelays := getRelays(t, request.numNearRelays, request.nearRelayRTTType, request.badNearRelay)

	startTime := time.Now()

	sessionVersion := uint32(1)
	if request.prevRouteType == routing.RouteTypeDirect {
		sessionVersion = 0
	}

	relays := getRelays(t, backend.numRouteMatrixRelays, goodRTT, false)

	statsdb := routing.NewStatsDatabase()
	var routeMatrix routing.RouteMatrix

	for i := int32(0); i < relays.Count; i++ {
		relayID := relays.IDs[i]
		relayIndex := int32(i)
		relayName := "test.relay." + fmt.Sprintf("%d", i+1)

		if routeMatrix.RelayIDsToIndices == nil {
			routeMatrix.RelayIDsToIndices = make(map[uint64]int32)
		}

		routeMatrix.RelayIDsToIndices[relayID] = relayIndex
		routeMatrix.RelayIDs = append(routeMatrix.RelayIDs, relayID)
		routeMatrix.RelayAddresses = append(routeMatrix.RelayAddresses, relays.Addresses[i])
		routeMatrix.RelayNames = append(routeMatrix.RelayNames, relayName)
		routeMatrix.RelayLatitudes = append(routeMatrix.RelayLatitudes, float32(0+i*5))
		routeMatrix.RelayLongitudes = append(routeMatrix.RelayLongitudes, float32(45+i*10))
		routeMatrix.RelayDatacenterIDs = append(routeMatrix.RelayDatacenterIDs, request.datacenterID+uint64(i))

		// Assign to statsdb directly to avoid the initial invalid history for testing
		statsdb.Entries[relayID] = routing.StatsEntry{Relays: make(map[uint64]*routing.StatsEntryRelay)}
		for j := int32(0); j < relays.Count; j++ {
			otherRelayID := relays.IDs[j]

			if relayID == otherRelayID {
				continue
			}

			statsdb.Entries[relayID].Relays[otherRelayID] = &routing.StatsEntryRelay{
				RTT:        float32(relays.RTTs[i]),
				Jitter:     float32(relays.Jitters[i]),
				PacketLoss: float32(relays.PacketLosses[i]),
			}
		}
	}

	numCPUs := int32(runtime.NumCPU())
	numSegments := relays.Count
	if numCPUs < relays.Count {
		numSegments = relays.Count / 5
		if numSegments == 0 {
			numSegments = 1
		}
	}
	costs := statsdb.GetCosts(relays.IDs, 5, 0.1)
	routeMatrix.RouteEntries = core.Optimize(int(relays.Count), int(numSegments), costs, 1, routeMatrix.RelayDatacenterIDs)

	routeMatrixFunc := func() *routing.RouteMatrix {
		return &routeMatrix
	}

	var destRelay uint64
	if nearRelays.Count > 0 {
		destRelay = nearRelays.IDs[0]
	}

	prevRouteInfo := getRouteInfo(
		t,
		request.numNearRelays,
		request.nearRelayRTTType,
		destRelay,
		&routeMatrix,
		request.prevRouteType,
		request.prevRouteNumRelays,
		sessionVersion,
		false,
		request.badRouteRelays,
		false,
	)

	routeShader := core.NewRouteShader()
	if backend.buyer != nil {
		routeShader = backend.buyer.RouteShader
	}

	internalConfig := core.NewInternalConfig()
	if backend.buyer != nil {
		internalConfig = backend.buyer.InternalConfig
	}

	var commitCounter int32
	if internalConfig.TryBeforeYouBuy {
		commitCounter = int32(request.sliceNumber)
	}

	initialSessionData := getInitialSessionData(t, startTime, request, prevRouteInfo, response, !internalConfig.TryBeforeYouBuy, commitCounter, routeShader.Multipath)

	var sessionDataSlice []byte
	var err error
	if initialSessionData != nil {
		sessionDataSlice, err = transport.MarshalSessionData(initialSessionData)
		assert.NoError(t, err)
	}

	var requestSessionData [transport.MaxSessionDataSize]byte

	if !request.badSessionData {
		copy(requestSessionData[:], sessionDataSlice)
	}

	clientAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:57247")
	assert.NoError(t, err)
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:32202")
	assert.NoError(t, err)

	publicKey, privateKey, err := box.GenerateKey(crand.Reader)
	assert.NoError(t, err)

	requestPacket := transport.SessionUpdatePacket{
		Version:              request.sdkVersion,
		SessionID:            request.sessionID,
		UserHash:             request.userHash,
		CustomerID:           request.buyerID,
		DatacenterID:         request.datacenterID,
		SliceNumber:          request.sliceNumber,
		ClientRoutePublicKey: publicKey[:],
		ServerRoutePublicKey: privateKey[:],
		ClientAddress:        *clientAddr,
		ServerAddress:        *serverAddr,
		ClientPingTimedOut:   request.clientPingTimedOut,
		SessionDataBytes:     int32(len(sessionDataSlice)),
		SessionData:          requestSessionData,
		NumNearRelays:        nearRelays.Count,
		NearRelayIDs:         nearRelays.IDs,
		NearRelayRTT:         nearRelays.RTTs,
		NearRelayJitter:      nearRelays.Jitters,
		NearRelayPacketLoss:  nearRelays.PacketLosses,
		FallbackToDirect:     request.fallbackToDirect,
		Next:                 request.prevRouteType != routing.RouteTypeDirect,
		DirectRTT:            float32(request.directStats.RTT),
		DirectJitter:         float32(request.directStats.Jitter),
		DirectPacketLoss:     float32(request.directStats.PacketLoss),
		NextRTT:              float32(request.nextStats.RTT),
		NextJitter:           float32(request.nextStats.Jitter),
		NextPacketLoss:       float32(request.nextStats.PacketLoss),
	}

	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	if initialSessionData == nil {
		initialSessionData = &transport.SessionData{}
	}

	// Set up backend

	ctx := context.Background()
	storer := &storage.InMemory{}

	if backend.buyer != nil {
		err := storer.AddBuyer(ctx, *backend.buyer)
		assert.NoError(t, err)
	}

	for _, datacenter := range backend.datacenters {
		err := storer.AddDatacenter(ctx, datacenter)
		assert.NoError(t, err)
	}

	for _, datacenterMap := range backend.datacenterMaps {
		err := storer.AddDatacenterMap(ctx, datacenterMap)
		assert.NoError(t, err)
	}

	seller := routing.Seller{ID: "sellerID"}
	if relays.Count > 0 {
		err := storer.AddSeller(ctx, seller)
		assert.NoError(t, err)
	}

	for i := int32(0); i < relays.Count; i++ {
		var datacenter routing.Datacenter

		if i < int32(len(backend.datacenters)) {
			datacenter = backend.datacenters[i]
		} else {
			t.Errorf("Tried to add %d relays but there was only %d datacenters. There should be one datacenter per relay.", relays.Count, len(backend.datacenters))
		}

		relayAddress, err := net.ResolveUDPAddr("udp", "127.0.0.1:4000"+fmt.Sprintf("%d", i))
		assert.NoError(t, err)

		relayInternalAddress := &net.UDPAddr{}
		if backend.internalAddresses {
			relayInternalAddress, err = net.ResolveUDPAddr("udp", "10.128.0.1:4000"+fmt.Sprintf("%d", i))
			assert.NoError(t, err)
		}

		err = storer.AddRelay(ctx, routing.Relay{
			ID:           relays.IDs[i],
			Seller:       seller,
			Datacenter:   datacenter,
			Addr:         *relayAddress,
			InternalAddr: *relayInternalAddress,
			PublicKey:    publicKey[:],
		})
		assert.NoError(t, err)
	}

	ipLocatorFunc := func(sessionID uint64) routing.IPLocator {
		return backend.ipLocator
	}

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
	assert.NoError(t, err)

	logger := log.NewNopLogger()

	sessionUpdateMetrics := getBlankSessionUpdateMetrics(t)

	responseBuffer := &bytes.Buffer{}
	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, logger, &metrics.EmptyPostSessionMetrics)

	routeInfo := getRouteInfo(
		t,
		request.numNearRelays,
		request.nearRelayRTTType,
		destRelay,
		&routeMatrix,
		response.routeType,
		response.routeNumRelays,
		initialSessionData.SessionVersion,
		backend.internalAddresses,
		false,
		request.badNearRelay,
	)

	var expireTimestamp uint64
	if response.routeType == routing.RouteTypeNew {
		expireTimestamp = uint64(startTime.Unix()) + billing.BillingSliceSeconds*2
	} else {
		expireTimestamp = uint64(startTime.Unix()) + billing.BillingSliceSeconds
	}

	// Determine veto reason
	var noRoute bool
	var multipathOverload bool
	var latencyWorse bool
	var commitVeto bool
	if request.prevRouteType != routing.RouteTypeDirect && response.routeType == routing.RouteTypeDirect {
		if request.directStats.RTT < request.nextStats.RTT {
			if internalConfig.TryBeforeYouBuy {
				commitVeto = true
			} else {
				latencyWorse = true
			}

		} else if request.directStats == getStats(overloadedRTT) {
			multipathOverload = true
		} else if response.routeNumRelays == 0 {
			noRoute = true
		}
	}

	committed := response.routeType != routing.RouteTypeDirect || initialSessionData.RouteState.Committed
	if internalConfig.TryBeforeYouBuy && request.directStats.RTT < request.nextStats.RTT {
		committed = false
	}

	if internalConfig.TryBeforeYouBuy {
		commitCounter++
	}

	expectedSessionData := getExpectedSessionData(t, *initialSessionData, expireTimestamp, request, routeInfo, response, noRoute, multipathOverload, latencyWorse, commitVeto, committed, commitCounter, routeShader.Multipath)

	sessionDataSlice, err = transport.MarshalSessionData(&expectedSessionData)
	assert.NoError(t, err)

	nearRelays = getRelays(t, response.numNearRelays, noRTT, false)

	expectedResponse := transport.SessionResponsePacket{
		Version:          request.sdkVersion,
		SessionID:        request.sessionID,
		SliceNumber:      request.sliceNumber,
		SessionDataBytes: int32(len(sessionDataSlice)),
		HasDebug:         response.debug,
	}

	if response.routeType != routing.RouteTypeDirect {
		expectedResponse.RouteType = int32(response.routeType)
		expectedResponse.Committed = expectedSessionData.RouteState.Committed
		expectedResponse.NumTokens = 2 + int32(len(routeInfo.routeRelayIDs))
	}

	if request.sdkVersion.Compare(transport.SDKVersion{4, 0, 4}) == transport.SDKVersionOlder || request.sliceNumber == 0 {
		expectedResponse.NumNearRelays = nearRelays.Count
		expectedResponse.NearRelayIDs = nearRelays.IDs
		expectedResponse.NearRelayAddresses = nearRelays.Addresses

		if request.sliceNumber == 0 && request.sdkVersion.AtLeast(transport.SDKVersion{4, 0, 4}) {
			expectedResponse.NearRelaysChanged = true
		}
	}

	copy(expectedResponse.SessionData[:], sessionDataSlice)

	// Run the session update handler
	handler := transport.SessionUpdateHandlerFunc(logger, ipLocatorFunc, routeMatrixFunc, multipathVetoHandler, storer, 32, *privateKey, postSessionHandler, sessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket
	responsePacket.Version = requestPacket.Version // Make sure we unmarshal the response the same way we marshaled the request
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	var actualSessionData transport.SessionData
	err = transport.UnmarshalSessionData(&actualSessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	// Compare the session data and response packet

	assert.Equal(t, expectedSessionData, actualSessionData)
	assertResponseEqual(t, expectedResponse, responsePacket)

	// Compare metrics

	assertAllMetricsEqual(t, *expectedMetrics, *sessionUpdateMetrics)

	// Check that the tokens sent down to the server have the correct external/internal relay addresses
	routeAddresses := append([]net.UDPAddr{}, *clientAddr)
	routeAddresses = append(routeAddresses, routeInfo.routeRelayAddresses...)
	routeAddresses = append(routeAddresses, *serverAddr)

	if response.routeType == routing.RouteTypeNew {
		ok := assert.Equal(t, core.NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES*(2+len(routeInfo.routeRelayIDs)), len(responsePacket.Tokens))
		if !ok {
			t.FailNow()
		}

		var clientToken core.RouteToken
		assert.NoError(t, core.ReadEncryptedRouteToken(&clientToken, responsePacket.Tokens[0*core.NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES:], publicKey[:], privateKey[:]))
		if clientToken.NextAddress != nil {
			assert.Equal(t, routeAddresses[1], *clientToken.NextAddress)
		}

		for i := 0; i < len(routeInfo.routeRelayIDs); i++ {
			var relayToken core.RouteToken
			assert.NoError(t, core.ReadEncryptedRouteToken(&relayToken, responsePacket.Tokens[(i+1)*core.NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES:], publicKey[:], privateKey[:]))
			if relayToken.NextAddress != nil {
				assert.Equal(t, routeAddresses[i+2], *relayToken.NextAddress)
			}
		}
	}
}

func assertResponseEqual(t *testing.T, expectedResponse transport.SessionResponsePacket, actualResponse transport.SessionResponsePacket) {
	// We can't check if the entire response is equal since the response's tokens will be different each time
	// since the encryption generates random bytes for the nonce
	assert.Equal(t, expectedResponse.Version, actualResponse.Version)
	assert.Equal(t, expectedResponse.SessionID, actualResponse.SessionID)
	assert.Equal(t, expectedResponse.SliceNumber, actualResponse.SliceNumber)
	assert.Equal(t, expectedResponse.RouteType, actualResponse.RouteType)
	assert.Equal(t, expectedResponse.NearRelaysChanged, actualResponse.NearRelaysChanged)
	assert.Equal(t, expectedResponse.NumNearRelays, actualResponse.NumNearRelays)
	assert.Equal(t, expectedResponse.NearRelayIDs, actualResponse.NearRelayIDs)
	assert.Equal(t, expectedResponse.NearRelayAddresses, actualResponse.NearRelayAddresses)
	assert.Equal(t, expectedResponse.NumTokens, actualResponse.NumTokens)
	assert.Equal(t, expectedResponse.Multipath, actualResponse.Multipath)
	assert.Equal(t, expectedResponse.Committed, actualResponse.Committed)
	assert.Equal(t, expectedResponse.HasDebug, actualResponse.HasDebug)

	if expectedResponse.HasDebug {
		assert.NotEmpty(t, actualResponse.Debug)
	} else {
		assert.Empty(t, actualResponse.Debug)
	}
}

/// -------- ERROR TESTS ------------

// These tests check each error case of the session update handler
// and validate that direct slice is returned

// We have to do this test a bit manually since the server backend won't respond on a packet read failure
func TestSessionUpdateHandlerReadPacketFailure(t *testing.T) {
	metricsHandler := metrics.LocalHandler{}

	serverBackendMetrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)

	responseBuffer := &bytes.Buffer{}

	handler := transport.SessionUpdateHandlerFunc(log.NewNopLogger(), nil, nil, nil, nil, 0, [crypto.KeySize]byte{}, nil, serverBackendMetrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: nil,
	})

	assert.Len(t, responseBuffer.Bytes(), 0)

	assert.Equal(t, 1.0, serverBackendMetrics.SessionUpdateMetrics.ReadPacketFailure.Value())
}

func TestSessionUpdateHandlerClientPingTimedOut(t *testing.T) {
	request := NewSessionUpdateRequestConfig(t)
	request.sdkVersion = transport.SDKVersion{4, 0, 2}
	request.clientPingTimedOut = true

	backend := NewSessionUpdateBackendConfig(t)

	response := NewSessionUpdateResponseConfig(t)
	response.unchangedSessionData = true

	expectedMetrics := getBlankSessionUpdateMetrics(t)
	expectedMetrics.DirectSlices.Add(1)
	expectedMetrics.ClientPingTimedOut.Add(1)

	runSessionUpdateTest(t, request, backend, response, expectedMetrics)
}

func TestSessionUpdateHandlerBuyerNotFound(t *testing.T) {
	request := NewSessionUpdateRequestConfig(t)

	backend := NewSessionUpdateBackendConfig(t)

	response := NewSessionUpdateResponseConfig(t)
	response.unchangedSessionData = true

	expectedMetrics := getBlankSessionUpdateMetrics(t)
	expectedMetrics.DirectSlices.Add(1)
	expectedMetrics.BuyerNotFound.Add(1)

	runSessionUpdateTest(t, request, backend, response, expectedMetrics)
}

func TestSessionUpdateHandlerDatacenterNotFound(t *testing.T) {
	request := NewSessionUpdateRequestConfig(t)

	backend := NewSessionUpdateBackendConfig(t)
	backend.buyer = &routing.Buyer{ID: request.buyerID, Live: true}

	response := NewSessionUpdateResponseConfig(t)
	response.unchangedSessionData = true

	expectedMetrics := getBlankSessionUpdateMetrics(t)
	expectedMetrics.DirectSlices.Add(1)
	expectedMetrics.DatacenterNotFound.Add(1)

	runSessionUpdateTest(t, request, backend, response, expectedMetrics)
}

func TestSessionUpdateHandlerMisconfiguredDatacenterAlias(t *testing.T) {
	request := NewSessionUpdateRequestConfig(t)
	request.datacenterID = crypto.HashID("alias")

	backend := NewSessionUpdateBackendConfig(t)
	backend.buyer = &routing.Buyer{ID: request.buyerID, Live: true}
	backend.datacenterMaps = []routing.DatacenterMap{{BuyerID: request.buyerID, Alias: "alias"}}

	response := NewSessionUpdateResponseConfig(t)
	response.unchangedSessionData = true

	expectedMetrics := getBlankSessionUpdateMetrics(t)
	expectedMetrics.DirectSlices.Add(1)
	expectedMetrics.MisconfiguredDatacenterAlias.Add(1)

	runSessionUpdateTest(t, request, backend, response, expectedMetrics)
}

func TestSessionUpdateHandlerDatacenterNotAllowed(t *testing.T) {
	request := NewSessionUpdateRequestConfig(t)

	backend := NewSessionUpdateBackendConfig(t)
	backend.buyer = &routing.Buyer{ID: request.buyerID, Live: true}
	backend.datacenters = []routing.Datacenter{{ID: request.datacenterID}}

	response := NewSessionUpdateResponseConfig(t)
	response.unchangedSessionData = true

	expectedMetrics := getBlankSessionUpdateMetrics(t)
	expectedMetrics.DirectSlices.Add(1)
	expectedMetrics.DatacenterNotAllowed.Add(1)

	runSessionUpdateTest(t, request, backend, response, expectedMetrics)
}

func TestSessionUpdateHandlerClientLocateFailure(t *testing.T) {
	request := NewSessionUpdateRequestConfig(t)

	backend := NewSessionUpdateBackendConfig(t)
	backend.buyer = &routing.Buyer{ID: request.buyerID, Live: true}
	backend.datacenters = []routing.Datacenter{{ID: request.datacenterID}}
	backend.datacenterMaps = []routing.DatacenterMap{{BuyerID: request.buyerID, DatacenterID: request.datacenterID}}
	backend.ipLocator = &badIPLocator{}

	response := NewSessionUpdateResponseConfig(t)

	expectedMetrics := getBlankSessionUpdateMetrics(t)
	expectedMetrics.DirectSlices.Add(1)
	expectedMetrics.ClientLocateFailure.Add(1)

	runSessionUpdateTest(t, request, backend, response, expectedMetrics)
}

func TestSessionUpdateHandlerReadSessionDataFailure(t *testing.T) {
	request := NewSessionUpdateRequestConfig(t)
	request.sliceNumber = 1
	request.badSessionData = true

	backend := NewSessionUpdateBackendConfig(t)
	backend.buyer = &routing.Buyer{ID: request.buyerID, Live: true}
	backend.datacenters = []routing.Datacenter{{ID: request.datacenterID}}
	backend.datacenterMaps = []routing.DatacenterMap{{BuyerID: request.buyerID, DatacenterID: request.datacenterID}}

	response := NewSessionUpdateResponseConfig(t)

	expectedMetrics := getBlankSessionUpdateMetrics(t)
	expectedMetrics.DirectSlices.Add(1)
	expectedMetrics.ReadSessionDataFailure.Add(1)

	runSessionUpdateTest(t, request, backend, response, expectedMetrics)
}

func TestSessionUpdateHandlerSessionDataBadSessionID(t *testing.T) {
	request := NewSessionUpdateRequestConfig(t)
	request.sliceNumber = 1
	request.badSessionDataSessionID = true

	backend := NewSessionUpdateBackendConfig(t)
	backend.buyer = &routing.Buyer{ID: request.buyerID, Live: true}
	backend.datacenters = []routing.Datacenter{{ID: request.datacenterID}}
	backend.datacenterMaps = []routing.DatacenterMap{{BuyerID: request.buyerID, DatacenterID: request.datacenterID}}

	response := NewSessionUpdateResponseConfig(t)

	expectedMetrics := getBlankSessionUpdateMetrics(t)
	expectedMetrics.DirectSlices.Add(1)
	expectedMetrics.BadSessionID.Add(1)

	runSessionUpdateTest(t, request, backend, response, expectedMetrics)
}

func TestSessionUpdateHandlerSessionDataBadSliceNumber(t *testing.T) {
	request := NewSessionUpdateRequestConfig(t)
	request.sliceNumber = 1
	request.badSessionDataSliceNumber = true

	backend := NewSessionUpdateBackendConfig(t)
	backend.buyer = &routing.Buyer{ID: request.buyerID, Live: true}
	backend.datacenters = []routing.Datacenter{{ID: request.datacenterID}}
	backend.datacenterMaps = []routing.DatacenterMap{{BuyerID: request.buyerID, DatacenterID: request.datacenterID}}

	response := NewSessionUpdateResponseConfig(t)

	expectedMetrics := getBlankSessionUpdateMetrics(t)
	expectedMetrics.DirectSlices.Add(1)
	expectedMetrics.BadSliceNumber.Add(1)

	runSessionUpdateTest(t, request, backend, response, expectedMetrics)
}

func TestSessionUpdateHandlerBuyerNotLive(t *testing.T) {
	request := NewSessionUpdateRequestConfig(t)

	backend := NewSessionUpdateBackendConfig(t)
	backend.buyer = &routing.Buyer{ID: request.buyerID, Live: false}
	backend.datacenters = []routing.Datacenter{{ID: request.datacenterID}}
	backend.datacenterMaps = []routing.DatacenterMap{{BuyerID: request.buyerID, DatacenterID: request.datacenterID}}

	response := NewSessionUpdateResponseConfig(t)

	expectedMetrics := getBlankSessionUpdateMetrics(t)
	expectedMetrics.DirectSlices.Add(1)
	expectedMetrics.BuyerNotLive.Add(1)

	runSessionUpdateTest(t, request, backend, response, expectedMetrics)
}

func TestSessionUpdateHandlerFallbackToDirect(t *testing.T) {
	request := NewSessionUpdateRequestConfig(t)
	request.fallbackToDirect = true

	backend := NewSessionUpdateBackendConfig(t)
	backend.buyer = &routing.Buyer{ID: request.buyerID, Live: true}
	backend.datacenters = []routing.Datacenter{{ID: request.datacenterID}}
	backend.datacenterMaps = []routing.DatacenterMap{{BuyerID: request.buyerID, DatacenterID: request.datacenterID}}

	response := NewSessionUpdateResponseConfig(t)

	expectedMetrics := getBlankSessionUpdateMetrics(t)
	expectedMetrics.DirectSlices.Add(1)
	expectedMetrics.FallbackToDirectUnknownReason.Add(1)

	runSessionUpdateTest(t, request, backend, response, expectedMetrics)
}

func TestSessionUpdateHandlerNoDestRelays(t *testing.T) {
	request := NewSessionUpdateRequestConfig(t)

	backend := NewSessionUpdateBackendConfig(t)
	backend.buyer = &routing.Buyer{ID: request.buyerID, Live: true}
	backend.datacenters = []routing.Datacenter{{ID: request.datacenterID}}
	backend.datacenterMaps = []routing.DatacenterMap{{BuyerID: request.buyerID, DatacenterID: request.datacenterID}}

	response := NewSessionUpdateResponseConfig(t)

	expectedMetrics := getBlankSessionUpdateMetrics(t)
	expectedMetrics.DirectSlices.Add(1)
	expectedMetrics.NoRelaysInDatacenter.Add(1)

	runSessionUpdateTest(t, request, backend, response, expectedMetrics)
}

/// -------- END ERROR TESTS ------------

/// -------- SCENARIO TESTS -------------

// These tests check different scenarios where we need to make sure
// that we return a response with the correct values

// The first slice of any session should always return a direct route
// with a set of near relays populated in the response packet
func TestSessionUpdateHandlerFirstSlice(t *testing.T) {
	request := NewSessionUpdateRequestConfig(t)

	backend := NewSessionUpdateBackendConfig(t)
	backend.buyer = &routing.Buyer{ID: request.buyerID, Live: true, RouteShader: core.NewRouteShader(), InternalConfig: core.NewInternalConfig()}
	backend.datacenters = []routing.Datacenter{{ID: request.datacenterID}, {ID: request.datacenterID + 1}}
	backend.datacenterMaps = []routing.DatacenterMap{{BuyerID: request.buyerID, DatacenterID: request.datacenterID}}
	backend.numRouteMatrixRelays = 2

	response := NewSessionUpdateResponseConfig(t)
	response.numNearRelays = 2

	expectedMetrics := getBlankSessionUpdateMetrics(t)
	expectedMetrics.DirectSlices.Add(1)

	runSessionUpdateTest(t, request, backend, response, expectedMetrics)
}

// The core routing logic has decided that we can't improve this session,
// so send back a direct route with the same near relays
func TestSessionUpdateHandlerDirectRoute(t *testing.T) {
	request := NewSessionUpdateRequestConfig(t)
	request.sliceNumber = 1
	request.directStats = getStats(goodRTT)
	request.numNearRelays = 2
	request.nearRelayRTTType = badRTT

	backend := NewSessionUpdateBackendConfig(t)
	backend.buyer = &routing.Buyer{ID: request.buyerID, Live: true, RouteShader: core.NewRouteShader(), InternalConfig: core.NewInternalConfig()}
	backend.datacenters = []routing.Datacenter{{ID: request.datacenterID}, {ID: request.datacenterID + 1}}
	backend.datacenterMaps = []routing.DatacenterMap{{BuyerID: request.buyerID, DatacenterID: request.datacenterID}}
	backend.numRouteMatrixRelays = 2

	response := NewSessionUpdateResponseConfig(t)
	response.numNearRelays = 2
	response.attemptFindRoute = true

	expectedMetrics := getBlankSessionUpdateMetrics(t)
	expectedMetrics.DirectSlices.Add(1)

	runSessionUpdateTest(t, request, backend, response, expectedMetrics)
}

// This session can be improved, so we can send back a new
// network next route for the session to take
func TestSessionUpdateHandlerNextRoute(t *testing.T) {
	request := NewSessionUpdateRequestConfig(t)
	request.sliceNumber = 1
	request.directStats = getStats(badRTT)
	request.numNearRelays = 2
	request.nearRelayRTTType = goodRTT

	backend := NewSessionUpdateBackendConfig(t)
	backend.buyer = &routing.Buyer{ID: request.buyerID, Live: true, RouteShader: core.NewRouteShader(), InternalConfig: core.NewInternalConfig()}
	backend.datacenters = []routing.Datacenter{{ID: request.datacenterID}, {ID: request.datacenterID + 1}}
	backend.datacenterMaps = []routing.DatacenterMap{{BuyerID: request.buyerID, DatacenterID: request.datacenterID}}
	backend.numRouteMatrixRelays = 2

	response := NewSessionUpdateResponseConfig(t)
	response.routeType = routing.RouteTypeNew
	response.routeNumRelays = 2
	response.numNearRelays = 2
	response.attemptFindRoute = true

	expectedMetrics := getBlankSessionUpdateMetrics(t)
	expectedMetrics.NextSlices.Add(1)

	runSessionUpdateTest(t, request, backend, response, expectedMetrics)
}

// When possible (and if the feature is enabled), we want to use relay
// internal IP addresses to communicate across the supplier's internal network
// for lower jitter and a more consistent route.
func TestSessionUpdateNextRouteInternalIPs(t *testing.T) {
	err := os.Setenv("FEATURE_ENABLE_INTERNAL_IPS", "true")
	assert.NoError(t, err)

	request := NewSessionUpdateRequestConfig(t)
	request.sliceNumber = 1
	request.directStats = getStats(badRTT)
	request.numNearRelays = 2
	request.nearRelayRTTType = goodRTT

	backend := NewSessionUpdateBackendConfig(t)
	backend.buyer = &routing.Buyer{ID: request.buyerID, Live: true, RouteShader: core.NewRouteShader(), InternalConfig: core.NewInternalConfig()}
	backend.datacenters = []routing.Datacenter{{ID: request.datacenterID}, {ID: request.datacenterID + 1}}
	backend.datacenterMaps = []routing.DatacenterMap{{BuyerID: request.buyerID, DatacenterID: request.datacenterID}}
	backend.numRouteMatrixRelays = 2
	backend.internalAddresses = true

	response := NewSessionUpdateResponseConfig(t)
	response.routeType = routing.RouteTypeNew
	response.routeNumRelays = 2
	response.numNearRelays = 2
	response.attemptFindRoute = true

	expectedMetrics := getBlankSessionUpdateMetrics(t)
	expectedMetrics.NextSlices.Add(1)

	runSessionUpdateTest(t, request, backend, response, expectedMetrics)
}

// This session has already been using a network next route,
// so we should continue to use the route, even if the direct route
// is as good or even slightly better
func TestSessionUpdateHandlerContinueRoute(t *testing.T) {
	request := NewSessionUpdateRequestConfig(t)
	request.sliceNumber = 2
	request.directStats = getStats(goodRTT)
	request.nextStats = getStats(goodRTT)
	request.numNearRelays = 2
	request.nearRelayRTTType = goodRTT
	request.prevRouteType = routing.RouteTypeNew
	request.prevRouteNumRelays = 2

	backend := NewSessionUpdateBackendConfig(t)
	backend.buyer = &routing.Buyer{ID: request.buyerID, Live: true, RouteShader: core.NewRouteShader(), InternalConfig: core.NewInternalConfig()}
	backend.datacenters = []routing.Datacenter{{ID: request.datacenterID}, {ID: request.datacenterID + 1}, {ID: request.datacenterID + 2}}
	backend.datacenterMaps = []routing.DatacenterMap{{BuyerID: request.buyerID, DatacenterID: request.datacenterID}}
	backend.numRouteMatrixRelays = 3

	response := NewSessionUpdateResponseConfig(t)
	response.routeType = routing.RouteTypeContinue
	response.routeNumRelays = 2
	response.numNearRelays = 2
	response.attemptFindRoute = true

	expectedMetrics := getBlankSessionUpdateMetrics(t)
	expectedMetrics.NextSlices.Add(1)

	runSessionUpdateTest(t, request, backend, response, expectedMetrics)
}

// The session is continuing a network next route, but the route no longer
// exists because one or more relays no longer exist in the route matrix.
// The server backend should find another suitable route to take instead.
func TestSessionUpdateHandlerRelayWentAway(t *testing.T) {
	request := NewSessionUpdateRequestConfig(t)
	request.sliceNumber = 2
	request.directStats = getStats(badRTT)
	request.nextStats = getStats(goodRTT)
	request.numNearRelays = 2
	request.nearRelayRTTType = goodRTT
	request.prevRouteType = routing.RouteTypeNew
	request.badRouteRelays = true

	backend := NewSessionUpdateBackendConfig(t)
	backend.buyer = &routing.Buyer{ID: request.buyerID, Live: true, RouteShader: core.NewRouteShader(), InternalConfig: core.NewInternalConfig()}
	backend.datacenters = []routing.Datacenter{{ID: request.datacenterID}, {ID: request.datacenterID + 1}, {ID: request.datacenterID + 2}}
	backend.datacenterMaps = []routing.DatacenterMap{{BuyerID: request.buyerID, DatacenterID: request.datacenterID}}
	backend.numRouteMatrixRelays = 3

	response := NewSessionUpdateResponseConfig(t)
	response.routeType = routing.RouteTypeNew
	response.routeNumRelays = 2
	response.numNearRelays = 2
	response.attemptFindRoute = true

	expectedMetrics := getBlankSessionUpdateMetrics(t)
	expectedMetrics.NextSlices.Add(1)
	expectedMetrics.RouteDoesNotExist.Add(1)
	expectedMetrics.RouteSwitched.Add(1)

	runSessionUpdateTest(t, request, backend, response, expectedMetrics)
}

// The session is continuing a network next route, but the route is no longer
// suitable to take because a near relay has become unroutable, for example.
// The server backend should find another suitable route to take instead.
func TestSessionUpdateHandlerRouteSwitched(t *testing.T) {
	request := NewSessionUpdateRequestConfig(t)
	request.sliceNumber = 2
	request.directStats = getStats(badRTT)
	request.nextStats = getStats(goodRTT)
	request.numNearRelays = 3
	request.nearRelayRTTType = goodRTT
	request.prevRouteType = routing.RouteTypeNew
	request.prevRouteNumRelays = 2
	request.badNearRelay = true

	backend := NewSessionUpdateBackendConfig(t)
	backend.buyer = &routing.Buyer{ID: request.buyerID, Live: true, RouteShader: core.NewRouteShader(), InternalConfig: core.NewInternalConfig()}
	backend.datacenters = []routing.Datacenter{{ID: request.datacenterID}, {ID: request.datacenterID + 1}, {ID: request.datacenterID + 2}}
	backend.datacenterMaps = []routing.DatacenterMap{{BuyerID: request.buyerID, DatacenterID: request.datacenterID}}
	backend.numRouteMatrixRelays = 3

	response := NewSessionUpdateResponseConfig(t)
	response.routeType = routing.RouteTypeNew
	response.routeNumRelays = 2
	response.numNearRelays = 3
	response.attemptFindRoute = true

	expectedMetrics := getBlankSessionUpdateMetrics(t)
	expectedMetrics.NextSlices.Add(1)
	expectedMetrics.RouteSwitched.Add(1)

	runSessionUpdateTest(t, request, backend, response, expectedMetrics)
}

// The session was vetoed from taking network next because
// the route they were taking no longer exists and there was
// no other suitable route to take to improve over direct.
func TestSessionUpdateHandlerVetoNoRoute(t *testing.T) {
	request := NewSessionUpdateRequestConfig(t)
	request.sliceNumber = 2
	request.directStats = getStats(goodRTT)
	request.nextStats = getStats(goodRTT)
	request.numNearRelays = 2
	request.nearRelayRTTType = goodRTT
	request.prevRouteType = routing.RouteTypeContinue
	request.prevRouteNumRelays = 2
	request.badRouteRelays = true

	backend := NewSessionUpdateBackendConfig(t)
	backend.buyer = &routing.Buyer{ID: request.buyerID, Live: true, RouteShader: core.NewRouteShader(), InternalConfig: core.NewInternalConfig()}
	backend.datacenters = []routing.Datacenter{{ID: request.datacenterID}, {ID: request.datacenterID + 1}, {ID: request.datacenterID + 2}}
	backend.datacenterMaps = []routing.DatacenterMap{{BuyerID: request.buyerID, DatacenterID: request.datacenterID}}
	backend.numRouteMatrixRelays = 3

	response := NewSessionUpdateResponseConfig(t)
	response.numNearRelays = 2
	response.attemptFindRoute = true

	expectedMetrics := getBlankSessionUpdateMetrics(t)
	expectedMetrics.DirectSlices.Add(1)
	expectedMetrics.NoRoute.Add(1)
	expectedMetrics.RouteDoesNotExist.Add(1)

	runSessionUpdateTest(t, request, backend, response, expectedMetrics)
}

// The session was vetoed from taking network next because
// the server backend detected that multipath has overloaded
// the user's connection. This session will be vetoed, and any
// subsequent sessions from this user for the next 7 days will
// not be allowed to use multipath.
func TestSessionUpdateHandlerVetoMultipathOverloaded(t *testing.T) {
	request := NewSessionUpdateRequestConfig(t)
	request.sliceNumber = 2
	request.directStats = getStats(overloadedRTT)
	request.nextStats = getStats(overloadedRTT)
	request.numNearRelays = 2
	request.nearRelayRTTType = goodRTT
	request.prevRouteType = routing.RouteTypeContinue
	request.prevRouteNumRelays = 2

	routeShader := core.NewRouteShader()
	routeShader.Multipath = true

	backend := NewSessionUpdateBackendConfig(t)
	backend.buyer = &routing.Buyer{ID: request.buyerID, Live: true, RouteShader: routeShader, InternalConfig: core.NewInternalConfig()}
	backend.datacenters = []routing.Datacenter{{ID: request.datacenterID}, {ID: request.datacenterID + 1}, {ID: request.datacenterID + 2}}
	backend.datacenterMaps = []routing.DatacenterMap{{BuyerID: request.buyerID, DatacenterID: request.datacenterID}}
	backend.numRouteMatrixRelays = 3

	response := NewSessionUpdateResponseConfig(t)
	response.numNearRelays = 2
	response.attemptFindRoute = true

	expectedMetrics := getBlankSessionUpdateMetrics(t)
	expectedMetrics.DirectSlices.Add(1)
	expectedMetrics.MultipathOverload.Add(1)

	runSessionUpdateTest(t, request, backend, response, expectedMetrics)
}

// The session was vetoed from taking network next because
// nextwork next route actually increased latency over the direct route.
func TestSessionUpdateHandlerVetoLatencyWorse(t *testing.T) {
	request := NewSessionUpdateRequestConfig(t)
	request.sliceNumber = 2
	request.directStats = getStats(goodRTT)
	request.nextStats = getStats(badRTT)
	request.numNearRelays = 2
	request.nearRelayRTTType = goodRTT
	request.prevRouteType = routing.RouteTypeContinue
	request.prevRouteNumRelays = 2

	backend := NewSessionUpdateBackendConfig(t)
	backend.buyer = &routing.Buyer{ID: request.buyerID, Live: true, RouteShader: core.NewRouteShader(), InternalConfig: core.NewInternalConfig()}
	backend.datacenters = []routing.Datacenter{{ID: request.datacenterID}, {ID: request.datacenterID + 1}, {ID: request.datacenterID + 2}}
	backend.datacenterMaps = []routing.DatacenterMap{{BuyerID: request.buyerID, DatacenterID: request.datacenterID}}
	backend.numRouteMatrixRelays = 3

	response := NewSessionUpdateResponseConfig(t)
	response.numNearRelays = 2
	response.attemptFindRoute = true

	expectedMetrics := getBlankSessionUpdateMetrics(t)
	expectedMetrics.DirectSlices.Add(1)
	expectedMetrics.LatencyWorse.Add(1)

	runSessionUpdateTest(t, request, backend, response, expectedMetrics)
}

// The session has try before you buy enabled, so this test
// validates that we don't actually commit to the route
// until we know that it is satisfactory
func TestSessionUpdateHandlerCommitPending(t *testing.T) {
	request := NewSessionUpdateRequestConfig(t)
	request.sliceNumber = 2
	request.directStats = getStats(goodRTT)
	request.nextStats = getStats(slightlyBadRTT)
	request.numNearRelays = 2
	request.nearRelayRTTType = goodRTT
	request.prevRouteType = routing.RouteTypeContinue
	request.prevRouteNumRelays = 2

	internalConfig := core.NewInternalConfig()
	internalConfig.TryBeforeYouBuy = true

	backend := NewSessionUpdateBackendConfig(t)
	backend.buyer = &routing.Buyer{ID: request.buyerID, Live: true, RouteShader: core.NewRouteShader(), InternalConfig: internalConfig}
	backend.datacenters = []routing.Datacenter{{ID: request.datacenterID}, {ID: request.datacenterID + 1}, {ID: request.datacenterID + 2}}
	backend.datacenterMaps = []routing.DatacenterMap{{BuyerID: request.buyerID, DatacenterID: request.datacenterID}}
	backend.numRouteMatrixRelays = 3

	response := NewSessionUpdateResponseConfig(t)
	response.routeType = routing.RouteTypeContinue
	response.routeNumRelays = 2
	response.numNearRelays = 2
	response.attemptFindRoute = true

	expectedMetrics := getBlankSessionUpdateMetrics(t)
	expectedMetrics.NextSlices.Add(1)

	runSessionUpdateTest(t, request, backend, response, expectedMetrics)
}

// The session has try before you buy enabled, but the
// session took 3 slices to try out a network next route
// and couldn't get a good enough improvement. In this case,
// we should veto the session.
func TestSessionUpdateHandlerCommitVeto(t *testing.T) {
	request := NewSessionUpdateRequestConfig(t)
	request.sliceNumber = 3
	request.directStats = getStats(goodRTT)
	request.nextStats = getStats(slightlyBadRTT)
	request.numNearRelays = 2
	request.nearRelayRTTType = goodRTT
	request.prevRouteType = routing.RouteTypeContinue
	request.prevRouteNumRelays = 2

	internalConfig := core.NewInternalConfig()
	internalConfig.TryBeforeYouBuy = true

	backend := NewSessionUpdateBackendConfig(t)
	backend.buyer = &routing.Buyer{ID: request.buyerID, Live: true, RouteShader: core.NewRouteShader(), InternalConfig: internalConfig}
	backend.datacenters = []routing.Datacenter{{ID: request.datacenterID}, {ID: request.datacenterID + 1}, {ID: request.datacenterID + 2}}
	backend.datacenterMaps = []routing.DatacenterMap{{BuyerID: request.buyerID, DatacenterID: request.datacenterID}}
	backend.numRouteMatrixRelays = 3

	response := NewSessionUpdateResponseConfig(t)
	response.routeType = routing.RouteTypeDirect
	response.numNearRelays = 2
	response.attemptFindRoute = true

	expectedMetrics := getBlankSessionUpdateMetrics(t)
	expectedMetrics.DirectSlices.Add(1)

	runSessionUpdateTest(t, request, backend, response, expectedMetrics)
}

// When the buyer has the Debug flag set, we want to ensure that
// we send back whatever debug information is necessary in the response
func TestSessionUpdateDebugResponse(t *testing.T) {
	request := NewSessionUpdateRequestConfig(t)
	request.sdkVersion = transport.SDKVersion{4, 0, 4}
	request.sliceNumber = 1
	request.directStats = getStats(badRTT)
	request.numNearRelays = 2
	request.nearRelayRTTType = goodRTT

	backend := NewSessionUpdateBackendConfig(t)
	backend.buyer = &routing.Buyer{ID: request.buyerID, Live: true, Debug: true, RouteShader: core.NewRouteShader(), InternalConfig: core.NewInternalConfig()}
	backend.datacenters = []routing.Datacenter{{ID: request.datacenterID}, {ID: request.datacenterID + 1}}
	backend.datacenterMaps = []routing.DatacenterMap{{BuyerID: request.buyerID, DatacenterID: request.datacenterID}}
	backend.numRouteMatrixRelays = 2

	response := NewSessionUpdateResponseConfig(t)
	response.routeType = routing.RouteTypeNew
	response.routeNumRelays = 2
	response.numNearRelays = 2
	response.attemptFindRoute = true
	response.debug = true

	expectedMetrics := getBlankSessionUpdateMetrics(t)
	expectedMetrics.NextSlices.Add(1)

	runSessionUpdateTest(t, request, backend, response, expectedMetrics)
}

func TestSessionUpdateDesyncedNearRelays(t *testing.T) {
	// Seed the RNG so we don't get different results from running `make test`
	// and running the test directly in VSCode
	rand.Seed(0)
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}

	expectedMetrics := metrics.EmptyServerBackendMetrics
	var err error
	emptySessionUpdateMetrics := metrics.EmptySessionUpdateMetrics
	expectedMetrics.SessionUpdateMetrics = &emptySessionUpdateMetrics
	expectedMetrics.SessionUpdateMetrics.DirectSlices, err = metricsHandler.NewCounter(context.Background(), &metrics.Descriptor{})
	assert.NoError(t, err)
	expectedMetrics.SessionUpdateMetrics.DirectSlices.Add(1)

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)
	storer := &storage.InMemory{}
	err = storer.AddBuyer(context.Background(), routing.Buyer{
		ID:             100,
		Live:           true,
		RouteShader:    core.NewRouteShader(),
		InternalConfig: core.NewInternalConfig(),
	})
	assert.NoError(t, err)

	err = storer.AddDatacenter(context.Background(), routing.Datacenter{ID: 10})
	assert.NoError(t, err)

	err = storer.AddDatacenter(context.Background(), routing.Datacenter{ID: 11})
	assert.NoError(t, err)

	err = storer.AddDatacenterMap(context.Background(), routing.DatacenterMap{BuyerID: 100, DatacenterID: 11})
	assert.NoError(t, err)

	err = storer.AddSeller(context.Background(), routing.Seller{ID: "seller"})
	assert.NoError(t, err)

	relayAddr1, err := net.ResolveUDPAddr("udp", "127.0.0.1:10000")
	assert.NoError(t, err)
	relayAddr2, err := net.ResolveUDPAddr("udp", "127.0.0.1:10001")
	assert.NoError(t, err)

	publicKey := make([]byte, crypto.KeySize)
	privateKey := [crypto.KeySize]byte{}

	err = storer.AddRelay(context.Background(), routing.Relay{
		ID:         1,
		Name:       "test.relay.1",
		Addr:       *relayAddr1,
		PublicKey:  publicKey,
		Seller:     routing.Seller{ID: "seller"},
		Datacenter: routing.Datacenter{ID: 10},
	})
	assert.NoError(t, err)

	err = storer.AddRelay(context.Background(), routing.Relay{
		ID:         2,
		Name:       "test.relay.2",
		Addr:       *relayAddr2,
		PublicKey:  publicKey,
		Seller:     routing.Seller{ID: "seller"},
		Datacenter: routing.Datacenter{ID: 11},
	})
	assert.NoError(t, err)

	sessionDataStruct := transport.SessionData{
		Version:         transport.SessionDataVersion,
		SessionID:       1111,
		SliceNumber:     1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: uint64(time.Now().Unix()),
		RouteState: core.RouteState{
			NearRelayRTT: [core.MaxNearRelays]int32{10, 15},
		},
	}

	sessionDataSlice, err := transport.MarshalSessionData(&sessionDataStruct)
	assert.NoError(t, err)

	sessionDataArray := [transport.MaxSessionDataSize]byte{}
	copy(sessionDataArray[:], sessionDataSlice)

	clientAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:57247")
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:32202")

	requestPacket := transport.SessionUpdatePacket{
		Version:              transport.SDKVersion{4, 0, 4},
		SessionID:            1111,
		CustomerID:           100,
		DatacenterID:         11,
		SliceNumber:          1,
		SessionDataBytes:     int32(len(sessionDataSlice)),
		SessionData:          sessionDataArray,
		ClientAddress:        *clientAddr,
		ServerAddress:        *serverAddr,
		ClientRoutePublicKey: publicKey,
		ServerRoutePublicKey: publicKey,
		DirectRTT:            80,
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	var goodIPLocator goodIPLocator
	ipLocatorFunc := func(sessionID uint64) routing.IPLocator {
		return &goodIPLocator
	}

	routeMatrix := routing.RouteMatrix{
		RelayIDsToIndices:  map[uint64]int32{1: 0, 2: 1},
		RelayIDs:           []uint64{1, 2},
		RelayAddresses:     []net.UDPAddr{*relayAddr1, *relayAddr2},
		RelayNames:         []string{"test.relay.1", "test.relay.2"},
		RelayLatitudes:     []float32{90, 89},
		RelayLongitudes:    []float32{180, 179},
		RelayDatacenterIDs: []uint64{10, 11},
		RouteEntries: []core.RouteEntry{
			{
				DirectCost:     65,
				NumRoutes:      int32(core.TriMatrixLength(2)),
				RouteCost:      [core.MaxRoutesPerEntry]int32{35},
				RouteNumRelays: [core.MaxRoutesPerEntry]int32{2},
				RouteRelays: [core.MaxRoutesPerEntry][core.MaxRelaysPerRoute]int32{
					{
						1, 0,
					},
				},
				RouteHash: [core.MaxRoutesPerEntry]uint32{core.RouteHash(1, 0)},
			},
		},
	}
	routeMatrixFunc := func() *routing.RouteMatrix {
		return &routeMatrix
	}

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
	assert.NoError(t, err)

	expectedResponse := transport.SessionResponsePacket{
		Version:     requestPacket.Version,
		SessionID:   requestPacket.SessionID,
		SliceNumber: requestPacket.SliceNumber,
		RouteType:   routing.RouteTypeDirect,
	}

	expectedSessionData := transport.SessionData{
		Version:         transport.SessionDataVersion,
		SessionID:       requestPacket.SessionID,
		SessionVersion:  sessionDataStruct.SessionVersion,
		SliceNumber:     requestPacket.SliceNumber + 1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: uint64(time.Now().Unix()) + billing.BillingSliceSeconds,
		RouteState: core.RouteState{
			UserID:           requestPacket.UserHash,
			PLHistoryIndex:   1,
			PLHistorySamples: 1,
		},
	}

	expectedSessionDataSlice, err := transport.MarshalSessionData(&expectedSessionData)
	assert.NoError(t, err)

	expectedResponse.SessionDataBytes = int32(len(expectedSessionDataSlice))
	copy(expectedResponse.SessionData[:], expectedSessionDataSlice)

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, logger, metrics.PostSessionMetrics)
	handler := transport.SessionUpdateHandlerFunc(logger, ipLocatorFunc, routeMatrixFunc, multipathVetoHandler, storer, 32, privateKey, postSessionHandler, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket
	responsePacket.Version = requestPacket.Version
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	var sessionData transport.SessionData
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)

	assertResponseEqual(t, expectedResponse, responsePacket)
	assertAllMetricsEqual(t, *expectedMetrics.SessionUpdateMetrics, *metrics.SessionUpdateMetrics)
}

func TestSessionUpdateOneRelayInRouteMatrix(t *testing.T) {
	// Seed the RNG so we don't get different results from running `make test`
	// and running the test directly in VSCode
	rand.Seed(0)
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}

	expectedMetrics := metrics.EmptyServerBackendMetrics
	var err error
	emptySessionUpdateMetrics := metrics.EmptySessionUpdateMetrics
	expectedMetrics.SessionUpdateMetrics = &emptySessionUpdateMetrics
	expectedMetrics.SessionUpdateMetrics.DirectSlices, err = metricsHandler.NewCounter(context.Background(), &metrics.Descriptor{})
	assert.NoError(t, err)
	expectedMetrics.SessionUpdateMetrics.DirectSlices.Add(1)

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)
	storer := &storage.InMemory{}
	err = storer.AddBuyer(context.Background(), routing.Buyer{
		ID:             100,
		Live:           true,
		RouteShader:    core.NewRouteShader(),
		InternalConfig: core.NewInternalConfig(),
	})
	assert.NoError(t, err)

	err = storer.AddDatacenter(context.Background(), routing.Datacenter{ID: 10})
	assert.NoError(t, err)

	err = storer.AddDatacenterMap(context.Background(), routing.DatacenterMap{BuyerID: 100, DatacenterID: 10})
	assert.NoError(t, err)

	err = storer.AddSeller(context.Background(), routing.Seller{ID: "seller"})
	assert.NoError(t, err)

	relayAddr1, err := net.ResolveUDPAddr("udp", "127.0.0.1:10000")
	assert.NoError(t, err)
	relayAddr2, err := net.ResolveUDPAddr("udp", "127.0.0.1:10001")
	assert.NoError(t, err)

	publicKey := make([]byte, crypto.KeySize)
	privateKey := [crypto.KeySize]byte{}

	err = storer.AddRelay(context.Background(), routing.Relay{
		ID:         1,
		Name:       "test.relay.1",
		Addr:       *relayAddr1,
		PublicKey:  publicKey,
		Seller:     routing.Seller{ID: "seller"},
		Datacenter: routing.Datacenter{ID: 10},
	})
	assert.NoError(t, err)

	err = storer.AddRelay(context.Background(), routing.Relay{
		ID:         2,
		Name:       "test.relay.2",
		Addr:       *relayAddr2,
		PublicKey:  publicKey,
		Seller:     routing.Seller{ID: "seller"},
		Datacenter: routing.Datacenter{ID: 10},
	})
	assert.NoError(t, err)

	sessionDataStruct := transport.SessionData{
		Version:         transport.SessionDataVersion,
		SessionID:       1111,
		SliceNumber:     1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: uint64(time.Now().Unix()),
		RouteState: core.RouteState{
			Next:          true,
			ReduceLatency: true,
			Committed:     true,
			NearRelayRTT:  [core.MaxNearRelays]int32{10, 15},
		},
		EverOnNext: true,
	}

	sessionDataSlice, err := transport.MarshalSessionData(&sessionDataStruct)
	assert.NoError(t, err)

	sessionDataArray := [transport.MaxSessionDataSize]byte{}
	copy(sessionDataArray[:], sessionDataSlice)

	clientAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:57247")
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:32202")

	requestPacket := transport.SessionUpdatePacket{
		Version:              transport.SDKVersion{4, 0, 4},
		SessionID:            1111,
		CustomerID:           100,
		DatacenterID:         10,
		SliceNumber:          1,
		SessionDataBytes:     int32(len(sessionDataSlice)),
		SessionData:          sessionDataArray,
		ClientAddress:        *clientAddr,
		ServerAddress:        *serverAddr,
		ClientRoutePublicKey: publicKey,
		ServerRoutePublicKey: publicKey,
		DirectRTT:            80,
		Next:                 true,
		NextRTT:              60,
		NumNearRelays:        2,
		NearRelayIDs:         []uint64{1, 2},
		NearRelayRTT:         []int32{10, 15},
		NearRelayJitter:      []int32{0, 0},
		NearRelayPacketLoss:  []int32{0, 0},
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	var goodIPLocator goodIPLocator
	ipLocatorFunc := func(sessionID uint64) routing.IPLocator {
		return &goodIPLocator
	}

	routeMatrix := routing.RouteMatrix{
		RelayIDsToIndices:  map[uint64]int32{1: 0},
		RelayIDs:           []uint64{1},
		RelayAddresses:     []net.UDPAddr{*relayAddr1},
		RelayNames:         []string{"test.relay.1"},
		RelayLatitudes:     []float32{90},
		RelayLongitudes:    []float32{180},
		RelayDatacenterIDs: []uint64{10},
	}
	routeMatrixFunc := func() *routing.RouteMatrix {
		return &routeMatrix
	}

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
	assert.NoError(t, err)

	expectedResponse := transport.SessionResponsePacket{
		Version:     requestPacket.Version,
		SessionID:   requestPacket.SessionID,
		SliceNumber: requestPacket.SliceNumber,
		RouteType:   routing.RouteTypeDirect,
	}

	expectedSessionData := transport.SessionData{
		Version:         transport.SessionDataVersion,
		SessionID:       requestPacket.SessionID,
		SessionVersion:  sessionDataStruct.SessionVersion,
		SliceNumber:     requestPacket.SliceNumber + 1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: uint64(time.Now().Unix()) + billing.BillingSliceSeconds,
		RouteState: core.RouteState{
			UserID:           requestPacket.UserHash,
			ReduceLatency:    true,
			Committed:        true,
			NumNearRelays:    2,
			NearRelayRTT:     [core.MaxNearRelays]int32{10, 255},
			PLHistoryIndex:   1,
			PLHistorySamples: 1,
		},
		EverOnNext: true,
	}

	expectedSessionDataSlice, err := transport.MarshalSessionData(&expectedSessionData)
	assert.NoError(t, err)

	expectedResponse.SessionDataBytes = int32(len(expectedSessionDataSlice))
	copy(expectedResponse.SessionData[:], expectedSessionDataSlice)

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, logger, metrics.PostSessionMetrics)
	handler := transport.SessionUpdateHandlerFunc(logger, ipLocatorFunc, routeMatrixFunc, multipathVetoHandler, storer, 32, privateKey, postSessionHandler, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket
	responsePacket.Version = requestPacket.Version
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	var sessionData transport.SessionData
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)

	assertResponseEqual(t, expectedResponse, responsePacket)
	assertAllMetricsEqual(t, *expectedMetrics.SessionUpdateMetrics, *metrics.SessionUpdateMetrics)
}

func TestSessionUpdateHandlerESLProMode(t *testing.T) {
	// Seed the RNG so we don't get different results from running `make test`
	// and running the test directly in VSCode
	rand.Seed(0)
	logger := log.NewNopLogger()
	metricsHandler := metrics.LocalHandler{}

	expectedMetrics := metrics.EmptyServerBackendMetrics
	var err error
	emptySessionUpdateMetrics := metrics.EmptySessionUpdateMetrics
	expectedMetrics.SessionUpdateMetrics = &emptySessionUpdateMetrics
	expectedMetrics.SessionUpdateMetrics.NextSlices, err = metricsHandler.NewCounter(context.Background(), &metrics.Descriptor{})
	assert.NoError(t, err)
	expectedMetrics.SessionUpdateMetrics.NextSlices.Add(1)

	metrics, err := metrics.NewServerBackendMetrics(context.Background(), &metricsHandler)
	assert.NoError(t, err)
	responseBuffer := bytes.NewBuffer(nil)
	storer := &storage.InMemory{}
	err = storer.AddBuyer(context.Background(), routing.Buyer{
		ID:             0x1e4e8804454033c8,
		Live:           true,
		RouteShader:    core.NewRouteShader(),
		InternalConfig: core.NewInternalConfig(),
	})
	assert.NoError(t, err)

	err = storer.AddDatacenter(context.Background(), routing.Datacenter{ID: 10})
	assert.NoError(t, err)

	err = storer.AddDatacenter(context.Background(), routing.Datacenter{ID: 11})
	assert.NoError(t, err)

	err = storer.AddDatacenterMap(context.Background(), routing.DatacenterMap{BuyerID: 0x1e4e8804454033c8, DatacenterID: 11})
	assert.NoError(t, err)

	err = storer.AddSeller(context.Background(), routing.Seller{ID: "seller"})
	assert.NoError(t, err)

	relayAddr1, err := net.ResolveUDPAddr("udp", "127.0.0.1:10000")
	assert.NoError(t, err)
	relayAddr2, err := net.ResolveUDPAddr("udp", "127.0.0.1:10001")
	assert.NoError(t, err)

	publicKey := make([]byte, crypto.KeySize)
	privateKey := [crypto.KeySize]byte{}

	err = storer.AddRelay(context.Background(), routing.Relay{
		ID:         1,
		Addr:       *relayAddr1,
		PublicKey:  publicKey,
		Seller:     routing.Seller{ID: "seller"},
		Datacenter: routing.Datacenter{ID: 10},
	})
	assert.NoError(t, err)

	err = storer.AddRelay(context.Background(), routing.Relay{
		ID:         2,
		Addr:       *relayAddr2,
		PublicKey:  publicKey,
		Seller:     routing.Seller{ID: "seller"},
		Datacenter: routing.Datacenter{ID: 11},
	})
	assert.NoError(t, err)

	sessionDataStruct := transport.SessionData{
		Version:         transport.SessionDataVersion,
		SessionID:       1111,
		SliceNumber:     1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: uint64(time.Now().Unix()),
		RouteState: core.RouteState{
			NearRelayRTT: [core.MaxNearRelays]int32{10, 15},
		},
	}

	sessionDataSlice, err := transport.MarshalSessionData(&sessionDataStruct)
	assert.NoError(t, err)

	sessionDataArray := [transport.MaxSessionDataSize]byte{}
	copy(sessionDataArray[:], sessionDataSlice)

	clientAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:57247")
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:32202")

	requestPacket := transport.SessionUpdatePacket{
		Version:              transport.SDKVersion{4, 0, 4},
		SessionID:            1111,
		CustomerID:           0x1e4e8804454033c8,
		DatacenterID:         11,
		SliceNumber:          1,
		SessionDataBytes:     int32(len(sessionDataSlice)),
		SessionData:          sessionDataArray,
		ClientAddress:        *clientAddr,
		ServerAddress:        *serverAddr,
		ClientRoutePublicKey: publicKey,
		ServerRoutePublicKey: publicKey,
		DirectRTT:            60,
		NumNearRelays:        2,
		NearRelayIDs:         []uint64{1, 2},
		NearRelayRTT:         []int32{30, 35},
		NearRelayJitter:      []int32{0, 0},
		NearRelayPacketLoss:  []int32{0, 0},
		NumTags:              3,
		Tags:                 [transport.MaxTags]uint64{0x0, 0x0, crypto.HashID("pro")},
	}
	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	var goodIPLocator goodIPLocator
	ipLocatorFunc := func(sessionID uint64) routing.IPLocator {
		return &goodIPLocator
	}

	routeMatrix := routing.RouteMatrix{
		RelayIDsToIndices:  map[uint64]int32{1: 0, 2: 1},
		RelayIDs:           []uint64{1, 2},
		RelayAddresses:     []net.UDPAddr{*relayAddr1, *relayAddr2},
		RelayNames:         []string{"test.relay.1", "test.relay.2"},
		RelayLatitudes:     []float32{90, 89},
		RelayLongitudes:    []float32{180, 179},
		RelayDatacenterIDs: []uint64{10, 11},
		RouteEntries: []core.RouteEntry{
			{
				DirectCost:     65,
				NumRoutes:      int32(core.TriMatrixLength(2)),
				RouteCost:      [core.MaxRoutesPerEntry]int32{35},
				RouteNumRelays: [core.MaxRoutesPerEntry]int32{2},
				RouteRelays: [core.MaxRoutesPerEntry][core.MaxRelaysPerRoute]int32{
					{
						0, 1,
					},
				},
				RouteHash: [core.MaxRoutesPerEntry]uint32{core.RouteHash(0, 1)},
			},
		},
	}
	routeMatrixFunc := func() *routing.RouteMatrix {
		return &routeMatrix
	}

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)

	multipathVetoHandler, err := storage.NewMultipathVetoHandler(redisServer.Addr(), storer)
	assert.NoError(t, err)

	expireTimestamp := uint64(time.Now().Unix()) + billing.BillingSliceSeconds*2
	sessionVersion := sessionDataStruct.SessionVersion + 1

	tokenData := make([]byte, core.NEXT_ENCRYPTED_ROUTE_TOKEN_BYTES*4)
	routeAddresses := make([]*net.UDPAddr, 0)
	routeAddresses = append(routeAddresses, clientAddr, relayAddr1, relayAddr2, serverAddr)
	routePublicKeys := make([][]byte, 0)
	routePublicKeys = append(routePublicKeys, publicKey, publicKey, publicKey, publicKey)
	core.WriteRouteTokens(tokenData, expireTimestamp, requestPacket.SessionID, uint8(sessionVersion), 1024, 1024, 4, routeAddresses, routePublicKeys, privateKey)
	expectedResponse := transport.SessionResponsePacket{
		Version:     requestPacket.Version,
		SessionID:   requestPacket.SessionID,
		SliceNumber: requestPacket.SliceNumber,
		RouteType:   routing.RouteTypeNew,
		NumTokens:   4,
		Tokens:      tokenData,
		Multipath:   true,
		Committed:   true,
	}

	expectedSessionData := transport.SessionData{
		Version:         transport.SessionDataVersion,
		SessionID:       requestPacket.SessionID,
		SessionVersion:  sessionVersion,
		SliceNumber:     requestPacket.SliceNumber + 1,
		Location:        routing.LocationNullIsland,
		ExpireTimestamp: expireTimestamp,
		Initial:         true,
		RouteNumRelays:  2,
		RouteCost:       65 + core.CostBias,
		RouteRelayIDs:   [core.MaxRelaysPerRoute]uint64{2, 1},
		RouteState: core.RouteState{
			UserID:           requestPacket.UserHash,
			Next:             true,
			ReduceLatency:    true,
			Committed:        true,
			NumNearRelays:    2,
			NearRelayRTT:     [core.MaxNearRelays]int32{30, 35},
			Multipath:        true,
			ProMode:          true,
			PLHistoryIndex:   1,
			PLHistorySamples: 1,
		},
		EverOnNext: true,
	}

	expectedSessionDataSlice, err := transport.MarshalSessionData(&expectedSessionData)
	assert.NoError(t, err)

	expectedResponse.SessionDataBytes = int32(len(expectedSessionDataSlice))
	copy(expectedResponse.SessionData[:], expectedSessionDataSlice)

	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, logger, metrics.PostSessionMetrics)
	handler := transport.SessionUpdateHandlerFunc(logger, ipLocatorFunc, routeMatrixFunc, multipathVetoHandler, storer, 32, privateKey, postSessionHandler, metrics.SessionUpdateMetrics)
	handler(responseBuffer, &transport.UDPPacket{
		Data: requestData,
	})

	var responsePacket transport.SessionResponsePacket
	responsePacket.Version = requestPacket.Version
	err = transport.UnmarshalPacket(&responsePacket, responseBuffer.Bytes()[1+crypto.PacketHashSize:])
	assert.NoError(t, err)

	var sessionData transport.SessionData
	err = transport.UnmarshalSessionData(&sessionData, responsePacket.SessionData[:])
	assert.NoError(t, err)

	assert.Equal(t, expectedSessionData, sessionData)

	assertResponseEqual(t, expectedResponse, responsePacket)
	assertAllMetricsEqual(t, *expectedMetrics.SessionUpdateMetrics, *metrics.SessionUpdateMetrics)
}
