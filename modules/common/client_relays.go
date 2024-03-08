package common

import (
	"net"
	"sort"

	"github.com/networknext/next/modules/core"
)

func GetClientRelays(maxClientRelays int, distanceThreshold int, latencyThreshold float32, relayIds []uint64, relayAddresses []net.UDPAddr, relayLatitudes []float32, relayLongitudes []float32, sourceLatitude float32, sourceLongitude float32, destLatitude float32, destLongitude float32) ([]uint64, []net.UDPAddr) {

	// Are there no relays in the route matrix? Return empty set

	if len(relayIds) == 0 {
		clientRelayIds := make([]uint64, 0)
		clientRelayAddresses := make([]net.UDPAddr, 0)
		return clientRelayIds, clientRelayAddresses
	}

	// Work with the client relays as an array of structs first for easier sorting

	type ClientRelayData struct {
		Id        uint64
		Address   net.UDPAddr
		Latitude  float64
		Longitude float64
		Distance  int
	}

	clientRelayData := make([]ClientRelayData, len(relayIds))

	for i, relayId := range relayIds {
		clientRelayData[i].Id = relayId
		clientRelayData[i].Address = relayAddresses[i]
		clientRelayData[i].Latitude = float64(int64(relayLatitudes[i]))
		clientRelayData[i].Longitude = float64(int64(relayLongitudes[i]))
		clientRelayData[i].Distance = int(core.HaversineDistance(float64(sourceLatitude), float64(sourceLongitude), float64(clientRelayData[i].Latitude), float64(clientRelayData[i].Longitude)))
	}

	sort.SliceStable(clientRelayData, func(i, j int) bool { return clientRelayData[i].Id < clientRelayData[j].Id })

	sort.SliceStable(clientRelayData, func(i, j int) bool { return clientRelayData[i].Distance < clientRelayData[j].Distance })

	// Estimate direct latency

	directLatency := float32(3.0 / 2.0 * core.SpeedOfLightTimeMilliseconds_AB(float64(sourceLatitude), float64(sourceLongitude), float64(destLatitude), float64(destLongitude)))

	// Select client relays within distance threshold, provided estimated latency through client relay does not exceed direct latency (avoids out of way relays)

	relayMap := make(map[uint64]ClientRelayData)

	for i := 0; i < len(clientRelayData); i++ {

		if len(relayMap) == maxClientRelays {
			break
		}

		if clientRelayData[i].Distance > distanceThreshold {
			break
		}

		clientRelayLatency := 3.0 / 2.0 * float32(core.SpeedOfLightTimeMilliseconds_ABC(float64(sourceLatitude), float64(sourceLongitude), float64(clientRelayData[i].Latitude), float64(clientRelayData[i].Longitude), float64(destLatitude), float64(destLongitude)))

		if clientRelayLatency > directLatency+latencyThreshold {
			continue
		}

		relayMap[clientRelayData[i].Id] = clientRelayData[i]
	}

	// We already have enough relays? Just stop now and return them

	if len(relayMap) == maxClientRelays {
		clientRelayIds := make([]uint64, maxClientRelays)
		clientRelayAddresses := make([]net.UDPAddr, maxClientRelays)
		index := 0
		for k, v := range relayMap {
			clientRelayIds[index] = k
			clientRelayAddresses[index] = v.Address
			index++
		}
		return clientRelayIds, clientRelayAddresses
	}

	// We need more relays. Look for client relays around the *destination*
	// Paradoxically, this can really help, especially for cases like South America <-> Miami

	for i := range clientRelayData {
		clientRelayData[i].Distance = int(core.HaversineDistance(float64(destLatitude), float64(destLongitude), float64(clientRelayData[i].Latitude), float64(clientRelayData[i].Longitude)))
	}

	sort.SliceStable(clientRelayData, func(i, j int) bool { return clientRelayData[i].Distance < clientRelayData[j].Distance })

	for i := 0; i < len(clientRelayData); i++ {

		if len(relayMap) == maxClientRelays {
			break
		}

		clientRelayLatency := 3.0 / 2.0 * float32(core.SpeedOfLightTimeMilliseconds_ABC(float64(sourceLatitude), float64(sourceLongitude), float64(clientRelayData[i].Latitude), float64(clientRelayData[i].Longitude), float64(destLatitude), float64(destLongitude)))

		if clientRelayLatency > directLatency+latencyThreshold {
			continue
		}

		relayMap[clientRelayData[i].Id] = clientRelayData[i]
	}

	// Return results

	numClientRelays := len(relayMap)

	clientRelayIds := make([]uint64, numClientRelays)
	clientRelayAddresses := make([]net.UDPAddr, numClientRelays)
	index := 0
	for k, v := range relayMap {
		clientRelayIds[index] = k
		clientRelayAddresses[index] = v.Address
		index++
	}

	return clientRelayIds, clientRelayAddresses
}
