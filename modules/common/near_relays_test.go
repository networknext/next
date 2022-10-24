package common_test

import (
	"testing"
	"net"
	"fmt"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/common"

	"github.com/stretchr/testify/assert"
)

func TestGetNearRelays_Basic(t *testing.T) {

	t.Parallel()

	// setup locations

	const PlayerLatitude = 0.0
	const PlayerLongitude = -100.0

	const DestinationLatitude = 0.0
	const DestinationLongitude = +100.0

	// create a bunch of relays

	const NumRelays = 100

	relayIds := make([]uint64, NumRelays)
	relayAddresses := make([]net.UDPAddr, NumRelays)
	relayLatitudes := make([]float64, NumRelays)
	relayLongitudes := make([]float64, NumRelays)

	for i := range relayIds {
		relayIds[i] = uint64(i)
		relayAddresses[i] = *core.ParseAddress(fmt.Sprintf("127.0.0.1:%d", 2000+i))
	}

	// setup half the relays near the player, and the other half near the destination

	for i := range relayIds {
		if i < 50 {
			// near player
			relayLatitudes[i] = PlayerLatitude + float64(common.RandomInt(-10,+10))
			relayLongitudes[i] = PlayerLongitude + float64(common.RandomInt(-10,+10))
		} else {
			// near destination
			relayLatitudes[i] = DestinationLatitude + float64(common.RandomInt(-10,+10))
			relayLongitudes[i] = DestinationLongitude + float64(common.RandomInt(-10,+10))
		}
	}

	// get near relays -- we should find that all near relays are near the player

	const MaxNearRelays = 20
	const DistanceThreshold = 2500
	const LatencyThreshold = 30

	nearRelayIds, nearRelayAddresses := common.GetNearRelays(MaxNearRelays, DistanceThreshold, LatencyThreshold, relayIds, relayAddresses, relayLatitudes, relayLongitudes, PlayerLatitude, PlayerLongitude, DestinationLatitude, DestinationLongitude)

	for i := range nearRelayIds {
		assert.True(t, nearRelayIds[i] < 50)
		assert.Equal(t, nearRelayAddresses[i].String(), relayAddresses[nearRelayIds[i]].String())
	}
}
