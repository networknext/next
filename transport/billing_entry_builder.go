package transport

import (
	"net"
	"strconv"

	"github.com/networknext/backend/billing"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
)

// Convert new representation of data into old for billing entry
func BuildRouteRequest(updatePacket SessionUpdatePacket, buyer routing.Buyer, serverData ServerCacheEntry, location routing.Location, storer storage.Storer, clientRelays []routing.Relay) billing.RouteRequest {

	return billing.RouteRequest{
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
		ConnectionType:         billing.SessionConnectionType(updatePacket.ConnectionType),
		DatacenterId:           makeEntityID("Datacenter", serverData.Datacenter.ID),
		SequenceNumber:         updatePacket.Sequence,
		FallbackToDirect:       updatePacket.FallbackToDirect,
		VersionMajor:           serverData.SDKVersion.Major,
		VersionMinor:           serverData.SDKVersion.Minor,
		VersionPatch:           serverData.SDKVersion.Patch,
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
func buildIssuedNearRelayList(nearRelays []routing.Relay) []*billing.IssuedNearRelay {
	var issuedNearRelays []*billing.IssuedNearRelay
	for idx, nearRelay := range nearRelays {
		issuedNearRelays = append(issuedNearRelays, &billing.IssuedNearRelay{
			Index:          int32(idx),
			RelayId:        makeEntityID("Relay", nearRelay.ID),
			RelayIpAddress: udpAddrToAddress(nearRelay.Addr),
		})
	}

	return issuedNearRelays
}

func udpAddrToAddress(addr net.UDPAddr) *billing.Address {
	if addr.IP == nil {
		return &billing.Address{
			Ip:        nil,
			Type:      billing.Address_NONE,
			Port:      0,
			Formatted: "",
		}
	}

	ipv4 := addr.IP.To4()
	if ipv4 == nil {
		ipv6 := addr.IP.To16()
		if ipv6 == nil {
			return &billing.Address{
				Ip:        nil,
				Type:      billing.Address_NONE,
				Port:      0,
				Formatted: "",
			}
		}

		return &billing.Address{
			Ip:        []byte(ipv6),
			Type:      billing.Address_IPV6,
			Port:      uint32(addr.Port),
			Formatted: addr.String(),
		}
	}

	return &billing.Address{
		Ip:        []byte(ipv4),
		Type:      billing.Address_IPV4,
		Port:      uint32(addr.Port),
		Formatted: addr.String(),
	}
}

func makeEntityID(kind string, ID uint64) *billing.EntityId {
	return &billing.EntityId{
		Kind: kind,
		Name: strconv.FormatUint(ID, 10),
	}
}
