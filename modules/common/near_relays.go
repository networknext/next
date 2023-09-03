package common

import (
	"net"
	"sort"

	"github.com/networknext/next/modules/core"
)

func GetNearRelays(maxNearRelays int, distanceThreshold int, latencyThreshold float32, relayIds []uint64, relayAddresses []net.UDPAddr, relayLatitudes []float32, relayLongitudes []float32, sourceLatitude float32, sourceLongitude float32, destLatitude float32, destLongitude float32) ([]uint64, []net.UDPAddr) {

	// Are there no relays in the route matrix? Return empty set

	if len(relayIds) == 0 {
		nearRelayIds := make([]uint64, 0)
		nearRelayAddresses := make([]net.UDPAddr, 0)
		return nearRelayIds, nearRelayAddresses
	}

	// Work with the near relays as an array of structs first for easier sorting

	type NearRelayData struct {
		Id        uint64
		Address   net.UDPAddr
		Latitude  float64
		Longitude float64
		Distance  int
	}

	nearRelayData := make([]NearRelayData, len(relayIds))

	for i, relayId := range relayIds {
		nearRelayData[i].Id = relayId
		nearRelayData[i].Address = relayAddresses[i]
		nearRelayData[i].Latitude = float64(int64(relayLatitudes[i]))
		nearRelayData[i].Longitude = float64(int64(relayLongitudes[i]))
		nearRelayData[i].Distance = int(core.HaversineDistance(float64(sourceLatitude), float64(sourceLongitude), float64(nearRelayData[i].Latitude), float64(nearRelayData[i].Longitude)))
	}

	sort.SliceStable(nearRelayData, func(i, j int) bool { return nearRelayData[i].Id < nearRelayData[j].Id })

	sort.SliceStable(nearRelayData, func(i, j int) bool { return nearRelayData[i].Distance < nearRelayData[j].Distance })

	// Estimate direct latency

	directLatency := float32(3.0 / 2.0 * core.SpeedOfLightTimeMilliseconds_AB(float64(sourceLatitude), float64(sourceLongitude), float64(destLatitude), float64(destLongitude)))

	// Select near relays within distance threshold, provided estimated latency through near relay does not exceed direct latency (avoids out of way relays)

	relayMap := make(map[uint64]NearRelayData)

	for i := 0; i < len(nearRelayData); i++ {

		if len(relayMap) == maxNearRelays {
			break
		}

		if nearRelayData[i].Distance > distanceThreshold {
			break
		}

		nearRelayLatency := 3.0 / 2.0 * float32(core.SpeedOfLightTimeMilliseconds_ABC(float64(sourceLatitude), float64(sourceLongitude), float64(nearRelayData[i].Latitude), float64(nearRelayData[i].Longitude), float64(destLatitude), float64(destLongitude)))

		if nearRelayLatency > directLatency+latencyThreshold {
			continue
		}

		relayMap[nearRelayData[i].Id] = nearRelayData[i]
	}

	// We already have enough relays? Just stop now and return them

	if len(relayMap) == maxNearRelays {
		nearRelayIds := make([]uint64, maxNearRelays)
		nearRelayAddresses := make([]net.UDPAddr, maxNearRelays)
		index := 0
		for k, v := range relayMap {
			nearRelayIds[index] = k
			nearRelayAddresses[index] = v.Address
			index++
		}
		return nearRelayIds, nearRelayAddresses
	}

	// We need more relays. Look for near relays around the *destination*
	// Paradoxically, this can really help, especially for cases like South America <-> Miami

	for i := range nearRelayData {
		nearRelayData[i].Distance = int(core.HaversineDistance(float64(destLatitude), float64(destLongitude), float64(nearRelayData[i].Latitude), float64(nearRelayData[i].Longitude)))
	}

	sort.SliceStable(nearRelayData, func(i, j int) bool { return nearRelayData[i].Distance < nearRelayData[j].Distance })

	for i := 0; i < len(nearRelayData); i++ {

		if len(relayMap) == maxNearRelays {
			break
		}

		nearRelayLatency := 3.0 / 2.0 * float32(core.SpeedOfLightTimeMilliseconds_ABC(float64(sourceLatitude), float64(sourceLongitude), float64(nearRelayData[i].Latitude), float64(nearRelayData[i].Longitude), float64(destLatitude), float64(destLongitude)))

		if nearRelayLatency > directLatency+latencyThreshold {
			continue
		}

		relayMap[nearRelayData[i].Id] = nearRelayData[i]
	}

	// Return results

	numNearRelays := len(relayMap)

	nearRelayIds := make([]uint64, numNearRelays)
	nearRelayAddresses := make([]net.UDPAddr, numNearRelays)
	index := 0
	for k, v := range relayMap {
		nearRelayIds[index] = k
		nearRelayAddresses[index] = v.Address
		index++
	}

	return nearRelayIds, nearRelayAddresses
}
