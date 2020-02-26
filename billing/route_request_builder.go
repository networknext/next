package billing

import (
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport"
)

// Convert new representation of data into old for billing entry
func BuildRouteRequest(updatePacket transport.SessionUpdatePacket, buyer routing.Buyer, serverData transport.ServerCacheEntry, location routing.Location, storer storage.Storer, clientRelays []routing.Relay) RouteRequest {

	return RouteRequest{
		BuyerId:                makeEntityID("Buyer", buyer.ID),
		SessionId:              updatePacket.SessionId,
		UserHash:               updatePacket.UserHash,
		PlatformId:             updatePacket.PlatformId,
		DirectRtt:              updatePacket.DirectMinRtt,
		DirectJitter:           updatePacket.DirectJitter,
		DirectPacketLoss:       updatePacket.DirectPacketLoss,
		NextRtt:                updatePacket.NextMinRtt,
		NextJitter:             updatePacket.NextJitter,
		NextPacketLoss:         updatePacket.NextPacketLoss,
		ClientIpAddress:        udpAddrToAddress(updatePacket.ClientAddress),
		ServerIpAddress:        udpAddrToAddress(updatePacket.ServerAddress),
		ServerPrivateIpAddress: udpAddrToAddress(serverData.Server.Addr),
		ClientRoutePublicKey:   updatePacket.ClientRoutePublicKey,
		ServerRoutePublicKey:   serverData.Server.PublicKey,
		Tag:                    updatePacket.Tag,
		NearRelays:             buildNearRelayList(updatePacket, storer),
		IssuedNearRelays:       buildIssuedNearRelayList(clientRelays),
		ConnectionType:         SessionConnectionType(updatePacket.ConnectionType),
		DatacenterId:           makeEntityID("Datacenter", serverData.Datacenter.ID),
		SequenceNumber:         updatePacket.Sequence,
		FallbackToDirect:       updatePacket.FallbackToDirect,
		VersionMajor:           serverData.VersionMajor,
		VersionMinor:           serverData.VersionMinor,
		VersionPatch:           serverData.VersionPatch,
		Location: &Location{
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
func buildNearRelayList(updatePacket transport.SessionUpdatePacket, storer storage.Storer) []*NearRelay {
	var nearRelays []*NearRelay
	var i int32
	for i = 0; i < updatePacket.NumNearRelays; i++ {
		relay, ok := storer.Relay(updatePacket.NearRelayIds[i])
		if !ok {
			continue
		}

		nearRelays = append(
			nearRelays,
			&NearRelay{
				RelayId:    makeEntityID("Relay", relay.ID),
				Rtt:        float64(updatePacket.NearRelayMinRtt[i]),
				Jitter:     float64(updatePacket.NearRelayJitter[i]),
				PacketLoss: float64(updatePacket.NearRelayPacketLoss[i]),
			},
		)
	}

	return nearRelays
}

// The list of relays we are telling the client is close to
func buildIssuedNearRelayList(nearRelays []routing.Relay) []*IssuedNearRelay {
	var issuedNearRelays []*IssuedNearRelay
	for idx, nearRelay := range nearRelays {
		issuedNearRelays = append(issuedNearRelays, &IssuedNearRelay{
			Index:          int32(idx),
			RelayId:        makeEntityID("Relay", nearRelay.ID),
			RelayIpAddress: udpAddrToAddress(nearRelay.Addr),
		})
	}

	return issuedNearRelays
}
