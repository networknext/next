package transport_test

/*
import (
	"bytes"
	"context"
	crand "crypto/rand"
	"errors"
	"fmt"
	"math"
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
	unrouteableRTT rttType = 1
	goodRTT        rttType = 2
	overloadedRTT  rttType = 3
	badRTT         rttType = 4
	slightlyBadRTT rttType = 5
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

func getRelays(t *testing.T, numRelays int32, rttType rttType) relayGroup {
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

		switch rttType {
		case noRTT:
			relays.RTTs = append(relays.RTTs, 0)

		case unrouteableRTT:
			relays.RTTs = append(relays.RTTs, 255)

		case goodRTT:
			relays.RTTs = append(relays.RTTs, int32(15+i*5))

		case overloadedRTT:
			relays.RTTs = append(relays.RTTs, 500)

		case badRTT:
			relays.RTTs = append(relays.RTTs, int32(100+i*5))

		case slightlyBadRTT:
			relays.RTTs = append(relays.RTTs, int32(55+i*5))
		}

		relays.Jitters = append(relays.Jitters, 0)
		relays.PacketLosses = append(relays.PacketLosses, 0)
	}

	return relays
}

func getRouteRelays(t *testing.T, numRelays int32, internal bool, alternateRoute bool) relayGroup {
	relays := relayGroup{
		Count:        numRelays,
		IDs:          make([]uint64, 0),
		Addresses:    make([]net.UDPAddr, 0),
		RTTs:         make([]int32, 0),
		Jitters:      make([]int32, 0),
		PacketLosses: make([]int32, 0),
	}

	var relayAddress *net.UDPAddr
	var err error

	// We want an alternate route that excludes the second to last relay from the regular route
	if alternateRoute {
		for i := numRelays; i > 1; i-- {
			if internal && i < numRelays-1 {
				relayAddress, err = net.ResolveUDPAddr("udp", "10.128.0.1:4000"+fmt.Sprintf("%d", i))
				assert.NoError(t, err)
			} else {
				relayAddress, err = net.ResolveUDPAddr("udp", "127.0.0.1:4000"+fmt.Sprintf("%d", i))
				assert.NoError(t, err)
			}

			relays.IDs = append(relays.IDs, uint64(i+1))
			relays.Addresses = append(relays.Addresses, *relayAddress)
			relays.RTTs = append(relays.RTTs, int32(15+i*5))
			relays.Jitters = append(relays.Jitters, 0)
			relays.PacketLosses = append(relays.PacketLosses, 0)
		}

		relayAddress, err = net.ResolveUDPAddr("udp", "127.0.0.1:40000")
		assert.NoError(t, err)

		relays.IDs = append(relays.IDs, 1)
		relays.Addresses = append(relays.Addresses, *relayAddress)
		relays.RTTs = append(relays.RTTs, 15)
		relays.Jitters = append(relays.Jitters, 0)
		relays.PacketLosses = append(relays.PacketLosses, 0)

		return relays
	}

	for i := numRelays - 1; i >= 0; i-- {
		if internal && i < numRelays-1 {
			relayAddress, err = net.ResolveUDPAddr("udp", "10.128.0.1:4000"+fmt.Sprintf("%d", i))
			assert.NoError(t, err)
		} else {
			relayAddress, err = net.ResolveUDPAddr("udp", "127.0.0.1:4000"+fmt.Sprintf("%d", i))
			assert.NoError(t, err)
		}

		relays.IDs = append(relays.IDs, uint64(i+1))
		relays.Addresses = append(relays.Addresses, *relayAddress)
		relays.RTTs = append(relays.RTTs, int32(15+i*5))
		relays.Jitters = append(relays.Jitters, 0)
		relays.PacketLosses = append(relays.PacketLosses, 0)
	}

	return relays
}

func getStats(rttType rttType) routing.Stats {
	switch rttType {
	case unrouteableRTT:
		return routing.Stats{RTT: 255}

	case goodRTT:
		return routing.Stats{RTT: 15}

	case overloadedRTT:
		return routing.Stats{RTT: 500}

	case badRTT:
		return routing.Stats{RTT: 100}

	case slightlyBadRTT:
		return routing.Stats{RTT: 55}
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

func getPrevRouteInfo(
	t *testing.T,
	routeMatrix *routing.RouteMatrix,
	numNearRelays int32,
	nearRelayRTTType rttType,
	destRelay uint64,
	sessionVersion uint32,
	routeNumRelays int32,
	routeType int,
	internal bool,
	badNearRelay bool,
	badRoute bool,

) routeInfo {
	// Get the previous update's near relays and route relays
	nearRelays := getRelays(t, numNearRelays, nearRelayRTTType)
	routeRelays := getRouteRelays(t, routeNumRelays, internal, false)

	// Calculate the near and dest relay indices
	nearRelayIndices := make([]int32, nearRelays.Count)
	for i := 0; i < len(nearRelayIndices); i++ {
		nearRelayIndices[i] = routeMatrix.RelayIDsToIndices[nearRelays.IDs[i]]
	}

	routeRelayIndices := [core.MaxRelaysPerRoute]int32{}
	for i := int32(0); i < routeRelays.Count; i++ {
		routeRelayIndices[i] = routeMatrix.RelayIDsToIndices[routeRelays.IDs[i]]
	}

	destRelayIndex := routeMatrix.RelayIDsToIndices[destRelay]
	destRelayIndices := append([]int32{}, destRelayIndex)

	var routeCost int32

	if routeRelays.Count > 0 {
		routeCost = core.GetCurrentRouteCost(routeMatrix.RouteEntries, routeRelays.Count, routeRelayIndices, nearRelayIndices, nearRelays.RTTs, destRelayIndices, nil)
	}

	if routeCost <= 0 {
		routeCost = core.GetBestRouteCost(routeMatrix.RouteEntries, nearRelayIndices, nearRelays.RTTs, destRelayIndices)
	}

	if routeCost > 10000 {
		routeCost = 10000
	}

	// To simulate a near relay becoming unroutable, set it to 255 RTT
	if badNearRelay {
		if nearRelays.Count > 1 {
			nearRelays.RTTs[1] = 255
		}
	}

	// If the route is bad, make the second relay an unknown relay
	if routeRelays.Count > 1 && badRoute {
		routeRelays.IDs[1] = math.MaxUint64
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
		routeCost:           routeCost,
		next:                false,
	}

	return routeInfo
}

func getRouteInfo(
	t *testing.T,
	routeMatrix *routing.RouteMatrix,
	numNearRelays int32,
	nearRelayRTTType rttType,
	destRelay uint64,
	sessionVersion uint32,
	routeNumRelays int32,
	routeType int,
	internal bool,
	alternateRoute bool,
	vetoed bool,
) routeInfo {
	nearRelays := getRelays(t, numNearRelays, nearRelayRTTType)
	routeRelays := getRouteRelays(t, routeNumRelays, internal, alternateRoute)

	nearRelayIndices := make([]int32, nearRelays.Count)
	for i := 0; i < len(nearRelayIndices); i++ {
		nearRelayIndices[i] = routeMatrix.RelayIDsToIndices[nearRelays.IDs[i]]
	}

	routeRelayIndices := [core.MaxRelaysPerRoute]int32{}
	for i := int32(0); i < routeRelays.Count; i++ {
		routeRelayIndices[i] = routeMatrix.RelayIDsToIndices[routeRelays.IDs[i]]
	}

	destRelayIndex := routeMatrix.RelayIDsToIndices[destRelay]
	destRelayIndices := append([]int32{}, destRelayIndex)

	var routeCost int32

	if routeRelays.Count > 0 {
		var debug string
		routeCost = core.GetCurrentRouteCost(routeMatrix.RouteEntries, routeRelays.Count, routeRelayIndices, nearRelayIndices, nearRelays.RTTs, destRelayIndices, &debug)
	}

	if routeCost <= 0 {
		routeCost = core.GetBestRouteCost(routeMatrix.RouteEntries, nearRelayIndices, nearRelays.RTTs, destRelayIndices)
	}

	if routeCost > 10000 {
		routeCost = 10000
	}

	if vetoed {
		routeCost = 0
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
		routeCost:           routeCost,
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
	request *sessionUpdateRequestConfig,
	routeInfo routeInfo,
	response *sessionUpdateResponseConfig,
	committed bool,
	commitCounter int32,
	multipath bool,
	mispredictCounter uint32,
) *transport.SessionData {
	var sessionData *transport.SessionData

	nearRelays := getRelays(t, request.numNearRelays, request.nearRelayRTTType)
	if request.badNearRelay {
		if nearRelays.Count > 1 {
			nearRelays.RTTs[1] = 255
		}
	}

	if request.badSessionDataSessionID {
		sessionData = &transport.SessionData{}
	} else if request.badSessionDataSliceNumber {
		sessionData = &transport.SessionData{
			SessionID: request.sessionID,
		}
	} else if request.sliceNumber > 1 {
		prevRequest := *request
		prevRequest.sliceNumber -= 1

		sd := getExpectedSessionData(
			t,
			transport.SessionData{},
			uint64(time.Now().Unix()),
			&prevRequest,
			routeInfo,
			response,
			request.numNearRelays,
			request.nearRelayRTTType,
			false,
			request.badRoute,
			false,
			false,
			false,
			false,
			false,
			false,
			committed,
			commitCounter,
			multipath,
			mispredictCounter,
		)
		sessionData = &sd
	} else if request.sliceNumber > 0 {
		sessionData = &transport.SessionData{
			Version:         transport.SessionDataVersion,
			SessionID:       request.sessionID,
			SliceNumber:     request.sliceNumber,
			Location:        routing.LocationNullIsland,
			ExpireTimestamp: uint64(time.Now().Unix()),
			RouteState: core.RouteState{
				UserID:        request.userHash,
				NumNearRelays: nearRelays.Count,
			},
		}

		copy(sessionData.RouteState.NearRelayRTT[:], nearRelays.RTTs)
		copy(sessionData.RouteState.NearRelayJitter[:], nearRelays.Jitters)

		// Convert the near relay PL history from []int32 to []uint32
		nearRelayPLHistory := make([]uint32, nearRelays.Count)
		for i := 0; int32(i) < nearRelays.Count; i++ {
			nearRelayPLHistory[i] = uint32(nearRelays.PacketLosses[i])
		}

		copy(sessionData.RouteState.NearRelayPLHistory[:], nearRelayPLHistory)
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
	numNearRelays int32,
	nearRelayRTTType rttType,
	badNearRelay bool,
	badRoute bool,
	noRoute bool,
	multipathOverload bool,
	latencyWorse bool,
	commitVeto bool,
	mispredictVeto bool,
	sdkAborted bool,
	committed bool,
	commitCounter int32,
	multipath bool,
	mispredictCounter uint32,
) transport.SessionData {
	if response.unchangedSessionData {
		initialSessionData.Version = transport.SessionDataVersion
		return initialSessionData
	}

	packetsSent := uint64(0)
	if request.sliceNumber > 0 {
		packetsSent = uint64(request.sliceNumber) * 600
	}

	if request.badSessionData || request.badSessionDataSessionID {
		return transport.SessionData{
			Version:                       transport.SessionDataVersion,
			PrevPacketsSentClientToServer: packetsSent,
			PrevPacketsSentServerToClient: packetsSent,
			PrevPacketsLostClientToServer: 0,
			PrevPacketsLostServerToClient: 0,
		}
	}

	if request.badSessionDataSliceNumber {
		return transport.SessionData{
			Version:                       transport.SessionDataVersion,
			SessionID:                     request.sessionID,
			PrevPacketsSentClientToServer: packetsSent,
			PrevPacketsSentServerToClient: packetsSent,
			PrevPacketsLostClientToServer: 0,
			PrevPacketsLostServerToClient: 0,
		}
	}

	nearRelays := getRelays(t, numNearRelays, nearRelayRTTType)
	if request.badNearRelay {
		if nearRelays.Count > 1 {
			nearRelays.RTTs[1] = 255
		}
	}

	var everOnNext bool
	var reduceLatency bool
	if request.prevRouteType != routing.RouteTypeDirect || response.routeType != routing.RouteTypeDirect {
		// These fields are "sticky" - they will remain true as long as we take network next once
		everOnNext = true

		if !response.proMode {
			reduceLatency = true
		}
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
			Veto:              noRoute || multipathOverload || latencyWorse || commitVeto || mispredictVeto || sdkAborted,
			ReduceLatency:     reduceLatency,
			ProMode:           response.proMode,
			Committed:         committed,
			CommitCounter:     commitCounter,
			CommitVeto:        commitVeto,
			Multipath:         multipath || response.proMode,
			RelayWentAway:     badRoute,
			RouteLost:         badNearRelay || badRoute,
			NoRoute:           noRoute,
			MultipathOverload: multipathOverload,
			LatencyWorse:      latencyWorse,
			MispredictCounter: mispredictCounter,
			Mispredict:        mispredictVeto,
		},
		Initial:                       routeInfo.initial,
		FellBackToDirect:              request.fallbackToDirect,
		EverOnNext:                    everOnNext,
		RouteChanged:                  request.prevRouteType == routing.RouteTypeContinue && response.routeType == routing.RouteTypeNew,
		PrevPacketsSentClientToServer: packetsSent,
		PrevPacketsSentServerToClient: packetsSent,
		PrevPacketsLostClientToServer: 0,
		PrevPacketsLostServerToClient: 0,
	}

	if response.attemptUpdateNearRelays && !request.desyncedNearRelays {
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
	tags                      []uint64
	clientPingTimedOut        bool
	fallbackToDirect          bool
	numNearRelays             int32
	nearRelayRTTType          rttType
	directStats               routing.Stats
	nextStats                 routing.Stats
	prevRouteType             int
	prevRouteNumRelays        int32
	badSessionData            bool // The request sends up bad session data (fails to unmarshal)
	badSessionDataSessionID   bool // The request sends up a mismatched session ID in the session data
	badSessionDataSliceNumber bool // The request sends up a mismatched slice number in the session data
	badRoute                  bool // The request sends up a route in the session data with relays that no longer exist
	badNearRelay              bool // The request sends up a near relay with unroutable RTT
	desyncedNearRelays        bool // The request sends up a different number of near relays than what is stored in the session data
	sdkAborted                bool // The request sends up "next = false" when we expect the session to be on next
}

func NewSessionUpdateRequestConfig(t *testing.T) *sessionUpdateRequestConfig {
	return &sessionUpdateRequestConfig{
		sdkVersion:                transport.SDKVersionMin,
		sessionID:                 1111,
		sliceNumber:               0,
		buyerID:                   123,
		datacenterID:              10,
		userHash:                  12345,
		tags:                      nil,
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
		badRoute:                  false,
		badNearRelay:              false,
		desyncedNearRelays:        false,
		sdkAborted:                false,
	}
}

type sessionUpdateBackendConfig struct {
	ipLocator            routing.IPLocator
	numRouteMatrixRelays int32
	internalAddresses    bool
	buyer                *routing.Buyer
	datacenters          []routing.Datacenter
	datacenterMaps       []routing.DatacenterMap
	failSignatureCheck   bool
}

func NewSessionUpdateBackendConfig(t *testing.T) *sessionUpdateBackendConfig {
	return &sessionUpdateBackendConfig{
		ipLocator:            &goodIPLocator{},
		numRouteMatrixRelays: 0,
		internalAddresses:    false,
		buyer:                nil,
		datacenters:          nil,
		datacenterMaps:       nil,
		failSignatureCheck:   false,
	}
}

type sessionUpdateResponseConfig struct {
	numNearRelays           int32
	unchangedSessionData    bool // The backend should not have altered the session data
	attemptUpdateNearRelays bool // Should the backend have even attempted to update the near relays?
	attemptFindRoute        bool // Should the backend have even attempted to find a route (run through the core logic)?
	routeType               int
	routeNumRelays          int32
	proMode                 bool
	debug                   bool // The backend should have sent down a debug string
}

func NewSessionUpdateResponseConfig(t *testing.T) *sessionUpdateResponseConfig {
	return &sessionUpdateResponseConfig{
		numNearRelays:           0,
		unchangedSessionData:    false,
		attemptUpdateNearRelays: false,
		attemptFindRoute:        false,
		routeType:               routing.RouteTypeDirect,
		routeNumRelays:          0,
		proMode:                 false,
		debug:                   false,
	}
}

func runSessionUpdateTest(t *testing.T, request *sessionUpdateRequestConfig, backend *sessionUpdateBackendConfig, response *sessionUpdateResponseConfig, expectedMetrics *metrics.SessionUpdateMetrics) {
	customerPublicKey, customerPrivateKey, err := crypto.GenerateCustomerKeyPair()
	assert.NoError(t, err)

	customerPublicKey = customerPublicKey[8:]
	customerPrivateKey = customerPrivateKey[8:]

	// Set up route matrix

	relays := getRelays(t, backend.numRouteMatrixRelays, goodRTT)

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

	// Set up request packet

	var nearRelays relayGroup
	if request.desyncedNearRelays {
		nearRelays = getRelays(t, 0, noRTT)
	} else {
		nearRelays = getRelays(t, request.numNearRelays, request.nearRelayRTTType)
		if request.badNearRelay {
			if nearRelays.Count > 1 {
				nearRelays.RTTs[1] = 255
			}
		}
	}

	sessionVersion := uint32(1)
	if request.prevRouteType == routing.RouteTypeDirect {
		sessionVersion = 0
	}

	var destRelay uint64
	if nearRelays.Count > 0 {
		destRelay = nearRelays.IDs[0]
	}

	var prevRouteInfo routeInfo
	if request.sliceNumber > 0 {
		prevRouteInfo = getPrevRouteInfo(
			t,
			&routeMatrix,
			request.numNearRelays,
			request.nearRelayRTTType,
			destRelay,
			sessionVersion,
			request.prevRouteNumRelays,
			request.prevRouteType,
			false,
			request.badNearRelay,
			request.badRoute,
		)
	}

	routeShader := core.NewRouteShader()
	if backend.buyer != nil {
		routeShader = backend.buyer.RouteShader
	}

	internalConfig := core.NewInternalConfig()
	if backend.buyer != nil {
		internalConfig = backend.buyer.InternalConfig
	}

	var commitCounter int32
	if response.attemptFindRoute && internalConfig.TryBeforeYouBuy {
		commitCounter = int32(request.sliceNumber) - 1
		if commitCounter < 0 {
			commitCounter = 0
		}
	}

	var initialCommitted bool
	if request.prevRouteType != routing.RouteTypeDirect {
		if !internalConfig.TryBeforeYouBuy {
			initialCommitted = true
		}
	}

	var mispredictCounter uint32
	if initialCommitted && request.sliceNumber > 1 {
		mispredictCounter = request.sliceNumber - 2
	}

	initialSessionData := getInitialSessionData(t, request, prevRouteInfo, response, initialCommitted, commitCounter, routeShader.Multipath, mispredictCounter)

	var sessionDataSlice []byte
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

	tags := [transport.MaxTags]uint64{}
	copy(tags[:], request.tags)

	requestPacket := transport.SessionUpdatePacket{
		Version:                         request.sdkVersion,
		SessionID:                       request.sessionID,
		UserHash:                        request.userHash,
		CustomerID:                      request.buyerID,
		DatacenterID:                    request.datacenterID,
		SliceNumber:                     request.sliceNumber,
		ClientRoutePublicKey:            publicKey[:],
		ServerRoutePublicKey:            privateKey[:],
		ClientAddress:                   *clientAddr,
		ServerAddress:                   *serverAddr,
		ClientPingTimedOut:              request.clientPingTimedOut,
		NumTags:                         int32(len(request.tags)),
		Tags:                            tags,
		SessionDataBytes:                int32(len(sessionDataSlice)),
		SessionData:                     requestSessionData,
		NumNearRelays:                   nearRelays.Count,
		NearRelayIDs:                    nearRelays.IDs,
		NearRelayRTT:                    nearRelays.RTTs,
		NearRelayJitter:                 nearRelays.Jitters,
		NearRelayPacketLoss:             nearRelays.PacketLosses,
		FallbackToDirect:                request.fallbackToDirect,
		Next:                            request.prevRouteType != routing.RouteTypeDirect,
		DirectRTT:                       float32(request.directStats.RTT),
		DirectJitter:                    float32(request.directStats.Jitter),
		DirectPacketLoss:                float32(request.directStats.PacketLoss),
		NextRTT:                         float32(request.nextStats.RTT),
		NextJitter:                      float32(request.nextStats.Jitter),
		NextPacketLoss:                  float32(request.nextStats.PacketLoss),
		PacketsSentClientToServer:       uint64(request.sliceNumber) * 600,
		PacketsSentServerToClient:       uint64(request.sliceNumber) * 600,
		PacketsLostClientToServer:       0,
		PacketsLostServerToClient:       0,
		PacketsOutOfOrderClientToServer: 0,
		PacketsOutOfOrderServerToClient: 0,
	}

	if request.sdkAborted {
		requestPacket.Next = false
	}

	requestData, err := transport.MarshalPacket(&requestPacket)
	assert.NoError(t, err)

	if initialSessionData == nil {
		initialSessionData = &transport.SessionData{}
	}

	// Add the packet type byte and hash bytes to the request data so we can sign it properly
	requestDataHeader := append([]byte{transport.PacketTypeSessionUpdate}, make([]byte, crypto.PacketHashSize)...)
	requestData = append(requestDataHeader, requestData...)

	// Sign the packet
	requestData = crypto.SignPacket(customerPrivateKey, requestData)

	// Once the packet is signed, we need to remove the header before passing to the session update handler
	requestData = requestData[1+crypto.PacketHashSize:]

	// Set up backend

	ctx := context.Background()
	storer := &storage.InMemory{}

	if backend.buyer != nil {
		backend.buyer.PublicKey = nil
		if !backend.failSignatureCheck {
			backend.buyer.PublicKey = customerPublicKey
		}
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
	postSessionHandler := transport.NewPostSessionHandler(0, 0, nil, 0, nil, 0, false, nil, logger, &metrics.EmptyPostSessionMetrics)

	var routeInfo routeInfo
	if response.attemptFindRoute {
		routeInfo = getRouteInfo(
			t,
			&routeMatrix,
			request.numNearRelays,
			request.nearRelayRTTType,
			destRelay,
			initialSessionData.SessionVersion,
			response.routeNumRelays,
			response.routeType,
			backend.internalAddresses,
			request.badNearRelay,
			request.prevRouteType != routing.RouteTypeDirect && response.routeType == routing.RouteTypeDirect,
		)
	} else if !request.sdkAborted {
		routeInfo = prevRouteInfo
	} else {
		routeInfo.sessionVersion = prevRouteInfo.sessionVersion
	}

	expireTime := time.Now()
	if response.routeType == routing.RouteTypeNew {
		expireTime = expireTime.Add(time.Second * billing.BillingSliceSeconds * 2)
	} else {
		expireTime = expireTime.Add(time.Second * billing.BillingSliceSeconds)
	}

	committed := initialSessionData.RouteState.Committed

	if response.attemptFindRoute {
		// Committed will stay true once it's set
		if response.routeType != routing.RouteTypeDirect {
			committed = !internalConfig.Uncommitted && (!internalConfig.TryBeforeYouBuy || routeShader.Multipath)
		}

		if response.routeType == routing.RouteTypeContinue {
			if internalConfig.TryBeforeYouBuy {
				if request.nextStats.RTT <= request.directStats.RTT && request.nextStats.PacketLoss <= request.directStats.PacketLoss {
					committed = true
				}
			}
		}

		if request.prevRouteType != routing.RouteTypeDirect && internalConfig.TryBeforeYouBuy {
			commitCounter++
		}
	}

	if response.attemptFindRoute && int32(request.nextStats.RTT) >= routeInfo.routeCost+10 && !request.badRoute {
		mispredictCounter++
	}

	// Determine veto reason
	var noRoute bool
	var multipathOverload bool
	var latencyWorse bool
	var commitVeto bool
	var mispredictVeto bool

	if response.attemptFindRoute && request.prevRouteType != routing.RouteTypeDirect && response.routeType == routing.RouteTypeDirect {
		if request.directStats.RTT < request.nextStats.RTT {
			if internalConfig.TryBeforeYouBuy && commitCounter > 3 {
				commitVeto = true
			} else if mispredictCounter >= 3 {
				mispredictVeto = true
			} else {
				latencyWorse = true
			}
		} else if request.directStats == getStats(overloadedRTT) {
			multipathOverload = true
		} else if response.routeNumRelays == 0 {
			noRoute = true
		}
	}

	numNearRelays := response.numNearRelays
	if response.attemptUpdateNearRelays {
		numNearRelays = request.numNearRelays
	}

	expectedSessionData := getExpectedSessionData(
		t,
		*initialSessionData,
		uint64(expireTime.Unix()),
		request,
		routeInfo,
		response,
		numNearRelays,
		request.nearRelayRTTType,
		request.badNearRelay,
		request.badRoute,
		noRoute,
		multipathOverload,
		latencyWorse,
		commitVeto,
		mispredictVeto,
		request.sdkAborted,
		committed,
		commitCounter,
		routeShader.Multipath,
		mispredictCounter,
	)

	sessionDataSlice, err = transport.MarshalSessionData(&expectedSessionData)
	assert.NoError(t, err)

	nearRelays = getRelays(t, response.numNearRelays, noRTT)

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
		expectedResponse.Multipath = expectedSessionData.RouteState.Multipath
		expectedResponse.NumTokens = 2 + int32(len(routeInfo.routeRelayIDs))
	}

	if request.sdkVersion.Compare(transport.SDKVersion{4, 0, 4}) == transport.SDKVersionOlder || request.sliceNumber == 0 {
		expectedResponse.NumNearRelays = nearRelays.Count
		expectedResponse.NearRelayIDs = nearRelays.IDs

		if response.numNearRelays > backend.numRouteMatrixRelays {
			for i := backend.numRouteMatrixRelays; i < nearRelays.Count; i++ {
				nearRelays.Addresses[i] = net.UDPAddr{}
			}
		}

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

func TestSessionUpdateHandlerSignatureCheckFailed(t *testing.T) {
	request := NewSessionUpdateRequestConfig(t)

	backend := NewSessionUpdateBackendConfig(t)
	backend.buyer = &routing.Buyer{ID: request.buyerID, Live: true}
	backend.failSignatureCheck = true

	response := NewSessionUpdateResponseConfig(t)
	response.unchangedSessionData = true

	expectedMetrics := getBlankSessionUpdateMetrics(t)
	expectedMetrics.DirectSlices.Add(1)
	expectedMetrics.SignatureCheckFailed.Add(1)

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

func TestSessionUpdateDesyncedNearRelays(t *testing.T) {
	request := NewSessionUpdateRequestConfig(t)
	request.sliceNumber = 3
	request.directStats = getStats(goodRTT)
	request.nextStats = getStats(goodRTT)
	request.numNearRelays = 2
	request.nearRelayRTTType = goodRTT
	request.prevRouteType = routing.RouteTypeContinue
	request.prevRouteNumRelays = 2
	request.desyncedNearRelays = true

	backend := NewSessionUpdateBackendConfig(t)
	backend.buyer = &routing.Buyer{ID: request.buyerID, Live: true, Debug: true, RouteShader: core.NewRouteShader(), InternalConfig: core.NewInternalConfig()}
	backend.datacenters = []routing.Datacenter{{ID: request.datacenterID}, {ID: request.datacenterID + 1}}
	backend.datacenterMaps = []routing.DatacenterMap{{BuyerID: request.buyerID, DatacenterID: request.datacenterID}}
	backend.numRouteMatrixRelays = 2

	response := NewSessionUpdateResponseConfig(t)
	response.routeType = routing.RouteTypeDirect
	response.attemptUpdateNearRelays = true

	expectedMetrics := getBlankSessionUpdateMetrics(t)
	expectedMetrics.DirectSlices.Add(1)
	expectedMetrics.NearRelaysChanged.Add(1)

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
	request.directStats = getStats(badRTT)

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

// We want to ensure that the "local" datacenter is always usable by
// customers, as this is the datacenter that the SDK will default to
// when no datacenter is provided. This should be a valid datacenter
// so that customers can integrate easily and see their sessions
// in the portal after following the SDK documentation.
func TestSessionUpdateHandlerLocalDatacenter(t *testing.T) {
	request := NewSessionUpdateRequestConfig(t)
	request.sliceNumber = 1
	request.directStats = getStats(goodRTT)
	request.numNearRelays = 2
	request.nearRelayRTTType = goodRTT
	request.datacenterID = crypto.HashID("local")

	backend := NewSessionUpdateBackendConfig(t)
	backend.buyer = &routing.Buyer{ID: request.buyerID, Live: true, RouteShader: core.NewRouteShader(), InternalConfig: core.NewInternalConfig()}
	backend.datacenters = []routing.Datacenter{{ID: request.datacenterID}, {ID: request.datacenterID + 1}}
	backend.datacenterMaps = []routing.DatacenterMap{{BuyerID: request.buyerID, DatacenterID: request.datacenterID}}
	backend.numRouteMatrixRelays = 2

	response := NewSessionUpdateResponseConfig(t)
	response.numNearRelays = 2
	response.attemptUpdateNearRelays = true
	response.attemptFindRoute = true

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
	request.nearRelayRTTType = goodRTT

	backend := NewSessionUpdateBackendConfig(t)
	backend.buyer = &routing.Buyer{ID: request.buyerID, Live: true, RouteShader: core.NewRouteShader(), InternalConfig: core.NewInternalConfig()}
	backend.datacenters = []routing.Datacenter{{ID: request.datacenterID}, {ID: request.datacenterID + 1}}
	backend.datacenterMaps = []routing.DatacenterMap{{BuyerID: request.buyerID, DatacenterID: request.datacenterID}}
	backend.numRouteMatrixRelays = 2

	response := NewSessionUpdateResponseConfig(t)
	response.numNearRelays = 2
	response.attemptUpdateNearRelays = true
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
	request.directStats = getStats(unrouteableRTT)
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
	response.attemptUpdateNearRelays = true
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
	request.directStats = getStats(unrouteableRTT)
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
	response.attemptUpdateNearRelays = true
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
	response.attemptUpdateNearRelays = true
	response.attemptFindRoute = true

	expectedMetrics := getBlankSessionUpdateMetrics(t)
	expectedMetrics.NextSlices.Add(1)

	runSessionUpdateTest(t, request, backend, response, expectedMetrics)
}

// The session is continuing a network next route, but the route no longer
// exists because one or more relays no longer exist in the route matrix.
// The server backend should find another suitable route to take instead,
// and maintain the same number of near relays that the request already had,
// but set any missing near relays as unroutable.
func TestSessionUpdateHandlerRouteRelayWentAway(t *testing.T) {
	request := NewSessionUpdateRequestConfig(t)
	request.sliceNumber = 2
	request.directStats = getStats(unrouteableRTT)
	request.nextStats = getStats(goodRTT)
	request.numNearRelays = 2
	request.nearRelayRTTType = goodRTT
	request.prevRouteType = routing.RouteTypeContinue
	request.prevRouteNumRelays = 2
	request.badRoute = true

	backend := NewSessionUpdateBackendConfig(t)
	backend.buyer = &routing.Buyer{ID: request.buyerID, Live: true, RouteShader: core.NewRouteShader(), InternalConfig: core.NewInternalConfig()}
	backend.datacenters = []routing.Datacenter{{ID: request.datacenterID}, {ID: request.datacenterID + 1}, {ID: request.datacenterID + 2}}
	backend.datacenterMaps = []routing.DatacenterMap{{BuyerID: request.buyerID, DatacenterID: request.datacenterID}}
	backend.numRouteMatrixRelays = 3

	response := NewSessionUpdateResponseConfig(t)
	response.routeType = routing.RouteTypeNew
	response.routeNumRelays = 2
	response.numNearRelays = 2
	response.attemptUpdateNearRelays = true
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
	request.directStats = getStats(unrouteableRTT)
	request.nextStats = getStats(goodRTT)
	request.numNearRelays = 3
	request.nearRelayRTTType = goodRTT
	request.prevRouteType = routing.RouteTypeContinue
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
	response.attemptUpdateNearRelays = true
	response.attemptFindRoute = true

	expectedMetrics := getBlankSessionUpdateMetrics(t)
	expectedMetrics.NextSlices.Add(1)
	expectedMetrics.RouteSwitched.Add(1)

	runSessionUpdateTest(t, request, backend, response, expectedMetrics)
}

// The SDK can sometimes "abort" a session by sending up "next = false"
// in the session update request without falling back to direct.
// For this case, we should veto the session.
func TestSessionUpdateHandlerSDKAborted(t *testing.T) {
	request := NewSessionUpdateRequestConfig(t)
	request.sliceNumber = 2
	request.directStats = getStats(badRTT)
	request.nextStats = getStats(goodRTT)
	request.numNearRelays = 2
	request.nearRelayRTTType = goodRTT
	request.prevRouteType = routing.RouteTypeContinue
	request.prevRouteNumRelays = 2
	request.sdkAborted = true

	backend := NewSessionUpdateBackendConfig(t)
	backend.buyer = &routing.Buyer{ID: request.buyerID, Live: true, RouteShader: core.NewRouteShader(), InternalConfig: core.NewInternalConfig()}
	backend.datacenters = []routing.Datacenter{{ID: request.datacenterID}, {ID: request.datacenterID + 1}, {ID: request.datacenterID + 2}}
	backend.datacenterMaps = []routing.DatacenterMap{{BuyerID: request.buyerID, DatacenterID: request.datacenterID}}
	backend.numRouteMatrixRelays = 3

	response := NewSessionUpdateResponseConfig(t)
	response.numNearRelays = 2
	response.attemptUpdateNearRelays = true

	expectedMetrics := getBlankSessionUpdateMetrics(t)
	expectedMetrics.DirectSlices.Add(1)
	expectedMetrics.SDKAborted.Add(1)

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
	request.badRoute = true

	backend := NewSessionUpdateBackendConfig(t)
	backend.buyer = &routing.Buyer{ID: request.buyerID, Live: true, RouteShader: core.NewRouteShader(), InternalConfig: core.NewInternalConfig()}
	backend.datacenters = []routing.Datacenter{{ID: request.datacenterID}, {ID: request.datacenterID + 1}, {ID: request.datacenterID + 2}}
	backend.datacenterMaps = []routing.DatacenterMap{{BuyerID: request.buyerID, DatacenterID: request.datacenterID}}
	backend.numRouteMatrixRelays = 3

	response := NewSessionUpdateResponseConfig(t)
	response.numNearRelays = 2
	response.attemptUpdateNearRelays = true
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
	response.attemptUpdateNearRelays = true
	response.attemptFindRoute = true

	expectedMetrics := getBlankSessionUpdateMetrics(t)
	expectedMetrics.DirectSlices.Add(1)
	expectedMetrics.MultipathOverload.Add(1)

	runSessionUpdateTest(t, request, backend, response, expectedMetrics)
}

// The session was vetoed from taking network next because
// network next route actually increased latency over the direct route.
func TestSessionUpdateHandlerVetoLatencyWorse(t *testing.T) {
	request := NewSessionUpdateRequestConfig(t)
	request.sliceNumber = 2
	request.directStats = getStats(goodRTT)
	request.nextStats = getStats(unrouteableRTT)
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
	response.attemptUpdateNearRelays = true
	response.attemptFindRoute = true

	expectedMetrics := getBlankSessionUpdateMetrics(t)
	expectedMetrics.DirectSlices.Add(1)
	expectedMetrics.LatencyWorse.Add(1)

	runSessionUpdateTest(t, request, backend, response, expectedMetrics)
}

// The session was vetoed from taking network next because
// we mispredicted the route RTT three times in a row.
func TestSessionUpdateHandlerVetoMispredict(t *testing.T) {
	request := NewSessionUpdateRequestConfig(t)
	request.sliceNumber = 4
	request.directStats = getStats(goodRTT)
	request.nextStats = getStats(slightlyBadRTT)
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
	response.attemptUpdateNearRelays = true
	response.attemptFindRoute = true

	expectedMetrics := getBlankSessionUpdateMetrics(t)
	expectedMetrics.DirectSlices.Add(1)
	expectedMetrics.MispredictVeto.Add(1)

	runSessionUpdateTest(t, request, backend, response, expectedMetrics)
}

// The session has try before you buy enabled, so this test
// validates that we don't actually commit to the route
// until we know that it is satisfactory
func TestSessionUpdateHandlerCommitPending(t *testing.T) {
	request := NewSessionUpdateRequestConfig(t)
	request.sliceNumber = 2
	request.directStats = getStats(goodRTT)
	request.nextStats = getStats(badRTT)
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
	response.attemptUpdateNearRelays = true
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
	request.sliceNumber = 4
	request.directStats = getStats(goodRTT)
	request.nextStats = getStats(badRTT)
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
	response.attemptUpdateNearRelays = true
	response.attemptFindRoute = true

	expectedMetrics := getBlankSessionUpdateMetrics(t)
	expectedMetrics.DirectSlices.Add(1)

	runSessionUpdateTest(t, request, backend, response, expectedMetrics)
}

// When the buyer has multipath enabled, we want to make sure that the
// routing logic works correctly and the response packet has multipath set to true
func TestSessionUpdateMultipath(t *testing.T) {
	request := NewSessionUpdateRequestConfig(t)
	request.sliceNumber = 1
	request.directStats = getStats(slightlyBadRTT)
	request.numNearRelays = 2
	request.nearRelayRTTType = goodRTT

	routeShader := core.NewRouteShader()
	routeShader.Multipath = true

	backend := NewSessionUpdateBackendConfig(t)
	backend.buyer = &routing.Buyer{ID: request.buyerID, Live: true, Debug: true, RouteShader: routeShader, InternalConfig: core.NewInternalConfig()}
	backend.datacenters = []routing.Datacenter{{ID: request.datacenterID}, {ID: request.datacenterID + 1}}
	backend.datacenterMaps = []routing.DatacenterMap{{BuyerID: request.buyerID, DatacenterID: request.datacenterID}}
	backend.numRouteMatrixRelays = 2

	response := NewSessionUpdateResponseConfig(t)
	response.routeType = routing.RouteTypeNew
	response.routeNumRelays = 2
	response.numNearRelays = 2
	response.attemptUpdateNearRelays = true
	response.attemptFindRoute = true

	expectedMetrics := getBlankSessionUpdateMetrics(t)
	expectedMetrics.NextSlices.Add(1)

	runSessionUpdateTest(t, request, backend, response, expectedMetrics)
}

// When the buyer has the Debug flag set, we want to ensure that
// we send back whatever debug information is necessary in the response
func TestSessionUpdateDebugResponse(t *testing.T) {
	request := NewSessionUpdateRequestConfig(t)
	request.sdkVersion = transport.SDKVersion{4, 0, 4}
	request.sliceNumber = 1
	request.directStats = getStats(unrouteableRTT)
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
	response.attemptUpdateNearRelays = true
	response.attemptFindRoute = true
	response.debug = true

	expectedMetrics := getBlankSessionUpdateMetrics(t)
	expectedMetrics.NextSlices.Add(1)

	runSessionUpdateTest(t, request, backend, response, expectedMetrics)
}

// Verify that the pro tag works correctly.
// When a session update comes in with a tag
// that contains "pro", enable pro mode.
func TestSessionUpdateESLProMode(t *testing.T) {
	request := NewSessionUpdateRequestConfig(t)
	request.sliceNumber = 1
	request.directStats = getStats(slightlyBadRTT)
	request.numNearRelays = 2
	request.nearRelayRTTType = goodRTT
	request.tags = []uint64{crypto.HashID("pro")}

	backend := NewSessionUpdateBackendConfig(t)
	backend.buyer = &routing.Buyer{ID: request.buyerID, Live: true, Debug: true, RouteShader: core.NewRouteShader(), InternalConfig: core.NewInternalConfig()}
	backend.datacenters = []routing.Datacenter{{ID: request.datacenterID}, {ID: request.datacenterID + 1}}
	backend.datacenterMaps = []routing.DatacenterMap{{BuyerID: request.buyerID, DatacenterID: request.datacenterID}}
	backend.numRouteMatrixRelays = 2

	response := NewSessionUpdateResponseConfig(t)
	response.routeType = routing.RouteTypeNew
	response.routeNumRelays = 2
	response.numNearRelays = 2
	response.attemptUpdateNearRelays = true
	response.attemptFindRoute = true
	response.proMode = true

	expectedMetrics := getBlankSessionUpdateMetrics(t)
	expectedMetrics.NextSlices.Add(1)

	runSessionUpdateTest(t, request, backend, response, expectedMetrics)
}
*/
