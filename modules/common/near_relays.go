package common

import (
	"net"
	"sort"

	"github.com/networknext/backend/modules/core"
)

// todo: this function needs to be unit tested

func GetNearRelays(routeMatrix *RouteMatrix, directLatency float32, sourceLatitude_in float32, sourceLongitude_in float32, destLatitude_in float32, destLongitude_in float32, maxNearRelays int, destDatacenterId uint64, distanceThreshold int, latencyThreshold float32) ([]uint64, []net.UDPAddr) {

	// Quantize to integer values so we don't have noise in low bits

	sourceLatitude := float64(int64(sourceLatitude_in))
	sourceLongitude := float64(int64(sourceLongitude_in))

	destLatitude := float64(int64(destLatitude_in))
	destLongitude := float64(int64(destLongitude_in))

	// If direct latency is 0, it's the first slice and we don't know it yet. Approximate it via speed of light * 2

	if directLatency <= 0.0 {
		directDistanceKilometers := core.HaversineDistance(sourceLatitude, sourceLongitude, destLatitude, destLongitude)
		directLatency = float32(directDistanceKilometers/299792.458*1000.0) * 2
	}

	// Work with the near relays as an array of structs first for easier sorting

	type NearRelayData struct {
		Id        uint64
		Address   net.UDPAddr
		Latitude  float64
		Longitude float64
		Distance  int
	}

	nearRelayData := make([]NearRelayData, len(routeMatrix.RelayIds))

	for i, relayId := range routeMatrix.RelayIds {
		nearRelayData[i].Id = relayId
		nearRelayData[i].Address = routeMatrix.RelayAddresses[i]
		nearRelayData[i].Latitude = float64(int64(routeMatrix.RelayLatitudes[i]))
		nearRelayData[i].Longitude = float64(int64(routeMatrix.RelayLongitudes[i]))
		nearRelayData[i].Distance = int(core.HaversineDistance(sourceLatitude, sourceLongitude, nearRelayData[i].Latitude, nearRelayData[i].Longitude))
	}

	sort.SliceStable(nearRelayData, func(i, j int) bool { return nearRelayData[i].Id < nearRelayData[j].Id })

	sort.SliceStable(nearRelayData, func(i, j int) bool { return nearRelayData[i].Distance < nearRelayData[j].Distance })

	// Select near relays

	relayMap := make(map[uint64]*NearRelayData)

	for i := 0; i < len(nearRelayData); i++ {

		if len(relayMap) == maxNearRelays {
			break
		}

		if nearRelayData[i].Distance > distanceThreshold {
			break
		}

		nearRelayLatency := 3.0 / 2.0 * float32(core.SpeedOfLightTimeMilliseconds(sourceLatitude, sourceLongitude, nearRelayData[i].Latitude, nearRelayData[i].Longitude, destLatitude, destLongitude))

		if nearRelayLatency > directLatency+latencyThreshold {
			continue
		}

		relayMap[nearRelayData[i].Id] = &nearRelayData[i]
	}

	// We already have enough relays? Just stop now and return them

	if len(relayMap) == maxNearRelays {
		nearRelayIds := make([]uint64, 0, maxNearRelays)
		nearRelayAddresses := make([]net.UDPAddr, 0, maxNearRelays)
		index := 0
		for k, v := range relayMap {
			nearRelayIds[index] = k
			nearRelayAddresses[index] = v.Address
		}
	}

	// We need more relays. Look for near relays around the *destination*
	// Paradoxically, this can really help, especially for cases like South America <-> Miami

	for i := range nearRelayData {
		nearRelayData[i].Distance = int(core.HaversineDistance(destLatitude, destLongitude, nearRelayData[i].Latitude, nearRelayData[i].Longitude))
	}

	sort.SliceStable(nearRelayData, func(i, j int) bool { return nearRelayData[i].Distance < nearRelayData[j].Distance })

	for i := 0; i < len(nearRelayData); i++ {

		if len(relayMap) == maxNearRelays {
			break
		}

		nearRelayLatency := 3.0 / 2.0 * float32(core.SpeedOfLightTimeMilliseconds(sourceLatitude, sourceLongitude, nearRelayData[i].Latitude, nearRelayData[i].Longitude, destLatitude, destLongitude))

		if nearRelayLatency > directLatency+latencyThreshold {
			continue
		}

		relayMap[nearRelayData[i].Id] = &nearRelayData[i]
	}

	// Return results, including -- potentially -- some relays around the destination datacenter

	numNearRelays := len(relayMap)

	nearRelayIds := make([]uint64, 0, numNearRelays)
	nearRelayAddresses := make([]net.UDPAddr, 0, numNearRelays)
	index := 0
	for k, v := range relayMap {
		nearRelayIds[index] = k
		nearRelayAddresses[index] = v.Address
	}

	return nearRelayIds, nearRelayAddresses
}
