package common

import (
	"net"
)

func GetNearRelays(routeMatrix *RouteMatrix, directLatency float32, sourceLatitude float32, sourceLongitude float32, destLatitude float32, destLongitude float32, maxNearRelays int, destDatacenterId uint64) ([]uint64, []net.UDPAddr) {

	// todo: we have to port this across

	nearRelayIds := make([]uint64, 0, maxNearRelays)
	nearRelayAddresses := make([]net.UDPAddr, 0, maxNearRelays)

	/*
	// Quantize to integer values so we don't have noise in low bits

	sourceLatitude := float64(int64(source_latitude))
	sourceLongitude := float64(int64(source_longitude))

	destLatitude := float64(int64(dest_latitude))
	destLongitude := float64(int64(dest_longitude))

	// If direct latency is 0, we don't know it yet. Approximate it via speed of light * 2

	if directLatency <= 0.0 {
		directDistanceKilometers := core.HaversineDistance(sourceLatitude, sourceLongitude, destLatitude, destLongitude)
		directLatency = float32(directDistanceKilometers/299792.458*1000.0) * 2
	}

	// Work with the near relays as an array of structs first for easier sorting

	type NearRelayData struct {
		ID        uint64
		Addr      net.UDPAddr
		Name      string
		Distance  int
		Latitude  float64
		Longitude float64
	}

	nearRelayData := make([]NearRelayData, len(m.RelayIDs))

	nearRelayIDs := make([]uint64, 0, maxNearRelays)
	nearRelayAddresses := make([]net.UDPAddr, 0, maxNearRelays)

	nearRelayIDMap := map[uint64]struct{}{}

	for i, relayID := range m.RelayIDs {
		nearRelayData[i].ID = relayID
		nearRelayData[i].Addr = m.RelayAddresses[i]
		nearRelayData[i].Name = m.RelayNames[i]
		nearRelayData[i].Latitude = float64(int64(m.RelayLatitudes[i]))
		nearRelayData[i].Longitude = float64(int64(m.RelayLongitudes[i]))
		nearRelayData[i].Distance = int(core.HaversineDistance(sourceLatitude, sourceLongitude, nearRelayData[i].Latitude, nearRelayData[i].Longitude))

		if _, isDestFirstRelay := m.DestFirstRelayIDsSet[relayID]; isDestFirstRelay && destDatacenterID == m.RelayDatacenterIDs[i] {
			// Always add "destination first" relays if we have a relay with the flag enabled in the destination datacenter
			if len(nearRelayIDs) == maxNearRelays {
				break
			}

			nearRelayIDs = append(nearRelayIDs, nearRelayData[i].ID)
			nearRelayIDMap[nearRelayData[i].ID] = struct{}{}

			if internalAddr, exists := m.InternalAddressClientRoutableRelayAddrMap[nearRelayData[i].ID]; exists {
				// Client should ping this relay using its internal address
				nearRelayAddresses = append(nearRelayAddresses, internalAddr)
			} else {
				nearRelayAddresses = append(nearRelayAddresses, nearRelayData[i].Addr)
			}
		}
	}

	// If we already have enough relays, stop and return them

	if len(nearRelayIDs) == maxNearRelays {
		return nearRelayIDs, nearRelayAddresses
	}

	sort.SliceStable(nearRelayData, func(i, j int) bool { return nearRelayData[i].Distance < nearRelayData[j].Distance })

	// Exclude any near relays whose 2/3rds speed of light * 2/3rds route through them is greater than direct + threshold
	// or who are further than x kilometers away from the player's location

	distanceThreshold := 2500

	latencyThreshold := float32(30.0)

	for i := 0; i < len(nearRelayData); i++ {

		if len(nearRelayIDs) == maxNearRelays {
			break
		}

		if nearRelayData[i].Distance > distanceThreshold {
			break
		}

		// don't add the same relay twice
		if _, ok := nearRelayIDMap[nearRelayData[i].ID]; ok {
			continue
		}

		nearRelayLatency := 3.0 / 2.0 * float32(core.SpeedOfLightTimeMilliseconds(sourceLatitude, sourceLongitude, nearRelayData[i].Latitude, nearRelayData[i].Longitude, destLatitude, destLongitude))
		if nearRelayLatency > directLatency+latencyThreshold {
			continue
		}

		nearRelayIDs = append(nearRelayIDs, nearRelayData[i].ID)
		nearRelayIDMap[nearRelayData[i].ID] = struct{}{}

		if internalAddr, exists := m.InternalAddressClientRoutableRelayAddrMap[nearRelayData[i].ID]; exists {
			// Client should ping this relay using its internal address
			nearRelayAddresses = append(nearRelayAddresses, internalAddr)
		} else {
			nearRelayAddresses = append(nearRelayAddresses, nearRelayData[i].Addr)
		}
	}

	// If we already have enough relays, stop and return them

	if len(nearRelayIDs) == maxNearRelays {
		return nearRelayIDs, nearRelayAddresses
	}

	// We need more relays. Look for near relays around the *destination*
	// Paradoxically, this can really help, especially for cases like South America <-> Miami

	for i := range m.RelayIDs {
		nearRelayData[i].Distance = int(core.HaversineDistance(destLatitude, destLongitude, nearRelayData[i].Latitude, nearRelayData[i].Longitude))
	}

	sort.SliceStable(nearRelayData, func(i, j int) bool { return nearRelayData[i].Distance < nearRelayData[j].Distance })

	for i := 0; i < len(nearRelayData); i++ {

		if len(nearRelayIDs) == maxNearRelays {
			break
		}

		// don't add the same relay twice
		if _, ok := nearRelayIDMap[nearRelayData[i].ID]; ok {
			continue
		}

		nearRelayLatency := 3.0 / 2.0 * float32(core.SpeedOfLightTimeMilliseconds(sourceLatitude, sourceLongitude, nearRelayData[i].Latitude, nearRelayData[i].Longitude, destLatitude, destLongitude))
		if nearRelayLatency > directLatency+latencyThreshold {
			continue
		}

		nearRelayIDs = append(nearRelayIDs, nearRelayData[i].ID)

		if internalAddr, exists := m.InternalAddressClientRoutableRelayAddrMap[nearRelayData[i].ID]; exists {
			// Client should ping this relay using its internal address
			nearRelayAddresses = append(nearRelayAddresses, internalAddr)
		} else {
			nearRelayAddresses = append(nearRelayAddresses, nearRelayData[i].Addr)
		}
	}
	*/

	return nearRelayIds, nearRelayAddresses
}
