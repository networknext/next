package transport

import (
	"github.com/networknext/backend/billing"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
)

// Convert new representation of data into old for billing entry
func BuildRouteRequest(updatePacket SessionUpdatePacket, buyer routing.Buyer, serverData ServerCacheEntry, location routing.Location, storer storage.Storer, clientRelays []*routing.Relay) billing.RouteRequest {
	return billing.RouteRequest{
		BuyerId:                billing.MakeEntityID("Buyer", buyer.ID),
		SessionId:              updatePacket.SessionId,
		UserHash:               updatePacket.UserHash,
		PlatformId:             updatePacket.PlatformId,
		DirectRtt:              updatePacket.DirectMinRtt,
		DirectJitter:           updatePacket.DirectJitter,
		DirectPacketLoss:       updatePacket.DirectPacketLoss,
		NextRtt:                updatePacket.NextMinRtt,
		NextJitter:             updatePacket.NextJitter,
		NextPacketLoss:         updatePacket.NextPacketLoss,
		ClientIpAddress:        billing.UdpAddrToAddress(updatePacket.ClientAddress),
		ServerIpAddress:        billing.UdpAddrToAddress(updatePacket.ServerAddress),
		ServerPrivateIpAddress: billing.UdpAddrToAddress(serverData.Server.Addr),
		ClientRoutePublicKey:   updatePacket.ClientRoutePublicKey,
		ServerRoutePublicKey:   serverData.Server.PublicKey,
		Tag:                    updatePacket.Tag,
		NearRelays:             buildNearRelayList(updatePacket, storer),
		IssuedNearRelays:       buildIssuedNearRelayList(clientRelays),
		ConnectionType:         billing.SessionConnectionType(updatePacket.ConnectionType),
		DatacenterId:           billing.MakeEntityID("Datacenter", serverData.Datacenter.ID),
		SequenceNumber:         updatePacket.Sequence,
		FallbackToDirect:       updatePacket.FallbackToDirect,
		VersionMajor:           serverData.SDKVersion.Major,
		VersionMinor:           serverData.SDKVersion.Minor,
		VersionPatch:           serverData.SDKVersion.Patch,
		// Can't get commented out fields unless we pay for MaxMind Pro, see Monday task for more details:
		// https://network-next.monday.com/boards/359802870/pulses/474491141
		Location: &billing.Location{
			// CountryCode: location.CountryCode,
			Country:   location.Country,
			Region:    location.Region,
			City:      location.City,
			Latitude:  float32(location.Latitude),
			Longitude: float32(location.Longitude),
			// Isp: location.ISP,
			// Asn: location.Asn,
			Continent: location.Continent,
		},
		UsageKbpsUp:               updatePacket.KbpsUp,
		UsageKbpsDown:             updatePacket.KbpsDown,
		Flagged:                   updatePacket.Flagged,
		TryBeforeYouBuy:           updatePacket.TryBeforeYouBuy,
		OnNetworkNext:             updatePacket.OnNetworkNext,
		PacketsLostClientToServer: updatePacket.PacketsLostClientToServer,
		PacketsLostServerToClient: updatePacket.PacketsLostServerToClient,
		FallbackFlags:             updatePacket.Flags,
		Committed:                 updatePacket.Committed,
	}
}

// The list of relays the client actually believes it is close to / is using (should match issued near relays)
func buildNearRelayList(updatePacket SessionUpdatePacket, storer storage.Storer) []*billing.NearRelay {
	var nearRelays []*billing.NearRelay
	var i int32
	for i = 0; i < updatePacket.NumNearRelays; i++ {
		relay, ok := storer.Relay(updatePacket.NearRelayIds[i])
		if !ok {
			continue
		}

		nearRelays = append(
			nearRelays,
			&billing.NearRelay{
				RelayId:    billing.MakeEntityID("Relay", relay.ID),
				Rtt:        float64(updatePacket.NearRelayMinRtt[i]),
				Jitter:     float64(updatePacket.NearRelayJitter[i]),
				PacketLoss: float64(updatePacket.NearRelayPacketLoss[i]),
			},
		)
	}

	return nearRelays
}

// The list of relays we are telling the client is close to
func buildIssuedNearRelayList(nearRelays []*routing.Relay) []*billing.IssuedNearRelay {
	var issuedNearRelays []*billing.IssuedNearRelay
	for idx, nearRelay := range nearRelays {
		issuedNearRelays = append(issuedNearRelays, &billing.IssuedNearRelay{
			Index:          int32(idx),
			RelayId:        billing.MakeEntityID("Relay", nearRelay.ID),
			RelayIpAddress: billing.UdpAddrToAddress(nearRelay.Addr),
		})
	}

	return issuedNearRelays
}
