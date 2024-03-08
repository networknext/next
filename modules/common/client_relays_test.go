package common_test

import (
	"fmt"
	"net"
	"testing"

	"github.com/networknext/next/modules/common"
	"github.com/networknext/next/modules/core"

	"github.com/stretchr/testify/assert"
)

func TestGetClientRelays_Basic(t *testing.T) {

	t.Parallel()

	// setup locations

	const PlayerLatitude = 0.0
	const PlayerLongitude = 0.0

	const DestinationLatitude = 0.0
	const DestinationLongitude = +100.0

	// create a bunch of relays

	const NumRelays = 100

	relayIds := make([]uint64, NumRelays)
	relayAddresses := make([]net.UDPAddr, NumRelays)
	relayLatitudes := make([]float32, NumRelays)
	relayLongitudes := make([]float32, NumRelays)

	for i := range relayIds {
		relayIds[i] = uint64(i)
		relayAddresses[i] = core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", 2000+i))
	}

	// setup half the relays near the player, and the other half near the destination

	for i := range relayIds {
		if i < 50 {
			// near player
			relayLatitudes[i] = PlayerLatitude + float32(common.RandomInt(-10, +10))
			relayLongitudes[i] = PlayerLongitude + float32(common.RandomInt(-10, +10))
		} else {
			// near destination
			relayLatitudes[i] = DestinationLatitude + float32(common.RandomInt(-10, +10))
			relayLongitudes[i] = DestinationLongitude + float32(common.RandomInt(-10, +10))
		}
	}

	// get client relays -- we should find that all client relays are near the player

	const MaxClientRelays = 20
	const DistanceThreshold = 2500
	const LatencyThreshold = 30

	clientRelayIds, clientRelayAddresses := common.GetClientRelays(MaxClientRelays, DistanceThreshold, LatencyThreshold, relayIds, relayAddresses, relayLatitudes, relayLongitudes, PlayerLatitude, PlayerLongitude, DestinationLatitude, DestinationLongitude)

	for i := range clientRelayIds {
		assert.True(t, clientRelayIds[i] < 50)
		assert.Equal(t, clientRelayAddresses[i].String(), relayAddresses[clientRelayIds[i]].String())
	}
}

func TestGetClientRelays_Dest(t *testing.T) {

	t.Parallel()

	// setup locations

	const PlayerLatitude = 0.0
	const PlayerLongitude = 0.0

	const DestinationLatitude = 0.0
	const DestinationLongitude = +100.0

	// create a bunch of relays

	const NumRelays = 100

	relayIds := make([]uint64, NumRelays)
	relayAddresses := make([]net.UDPAddr, NumRelays)
	relayLatitudes := make([]float32, NumRelays)
	relayLongitudes := make([]float32, NumRelays)

	for i := range relayIds {
		relayIds[i] = uint64(i)
		relayAddresses[i] = core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", 2000+i))
		relayLatitudes[i] = DestinationLatitude + float32(common.RandomInt(-10, +10))
		relayLongitudes[i] = DestinationLongitude + float32(common.RandomInt(-10, +10))
	}

	// get client relays -- we should not be able to find any near relays, but find dest relays

	const MaxClientRelays = 20
	const DistanceThreshold = 100
	const LatencyThreshold = 30

	clientRelayIds, clientRelayAddresses := common.GetClientRelays(MaxClientRelays, DistanceThreshold, LatencyThreshold, relayIds, relayAddresses, relayLatitudes, relayLongitudes, PlayerLatitude, PlayerLongitude, DestinationLatitude, DestinationLongitude)

	for i := range clientRelayIds {
		assert.Equal(t, clientRelayAddresses[i].String(), relayAddresses[clientRelayIds[i]].String())
	}
}

func TestGetClientRelays_OutOfWay(t *testing.T) {

	t.Parallel()

	// setup locations

	const PlayerLatitude = 0.0
	const PlayerLongitude = -0.0

	const DestinationLatitude = 0.0
	const DestinationLongitude = +100.0

	// create a bunch of relays

	const NumRelays = 150

	relayIds := make([]uint64, NumRelays)
	relayAddresses := make([]net.UDPAddr, NumRelays)
	relayLatitudes := make([]float32, NumRelays)
	relayLongitudes := make([]float32, NumRelays)

	for i := range relayIds {
		relayIds[i] = uint64(i)
		relayAddresses[i] = core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", 2000+i))
		if i < 50 {
			relayLatitudes[i] = PlayerLatitude + float32(common.RandomInt(0, +10))
			relayLongitudes[i] = PlayerLongitude + float32(common.RandomInt(0, +10))
		} else if i < 100 {
			relayLatitudes[i] = DestinationLatitude + float32(common.RandomInt(-10, 0))
			relayLongitudes[i] = DestinationLongitude + float32(common.RandomInt(-10, 0))
		} else {
			// out of the way
			relayLatitudes[i] = PlayerLatitude + 50
			relayLongitudes[i] = PlayerLongitude - 20
		}
	}

	// get client relays -- the out of the way relays should be excluded

	const MaxClientRelays = 250
	const DistanceThreshold = 2500
	const LatencyThreshold = 30

	clientRelayIds, clientRelayAddresses := common.GetClientRelays(MaxClientRelays, DistanceThreshold, LatencyThreshold, relayIds, relayAddresses, relayLatitudes, relayLongitudes, PlayerLatitude, PlayerLongitude, DestinationLatitude, DestinationLongitude)

	assert.Equal(t, 100, len(clientRelayIds))

	for i := range clientRelayIds {
		assert.Equal(t, clientRelayAddresses[i].String(), relayAddresses[clientRelayIds[i]].String())
	}
}
