package billing

import (
	"net"

	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/transport"
)

func udpAddrToAddress(addr net.UDPAddr) *Address {
	if addr.IP == nil {
		return &Address{
			Ip:        nil,
			Type:      Address_NONE,
			Port:      0,
			Formatted: "",
		}
	}

	ipv4 := addr.IP.To4()
	if ipv4 == nil {
		ipv6 := addr.IP.To16()
		if ipv6 == nil {
			return &Address{
				Ip:        nil,
				Type:      Address_NONE,
				Port:      0,
				Formatted: "",
			}
		}

		return &Address{
			Ip:        []byte(ipv6),
			Type:      Address_IPV6,
			Port:      uint32(addr.Port),
			Formatted: addr.String(),
		}
	}

	return &Address{
		Ip:        []byte(ipv4),
		Type:      Address_IPV4,
		Port:      uint32(addr.Port),
		Formatted: addr.String(),
	}
}

func makeEntityID(kind string, ID uint64) EntityId{
	return EntityId{
		Kind: kind,
		Name: strconv.Itoa(ID)
	}
}

// Convert new representation of data into old for billing entry
func BuildRouteRequest(updatePacket transport.SessionUpdatePacket, buyer routing.Buyer, serverData transport.ServerCacheEntry, location routing.Location) RouteRequest {
	return RouteRequest{
		BuyerId:                   makeEntityID("Buyer", buyer.Key),
		SessionId:                 updatePacket.SessionId,
		UserHash:                  updatePacket.UserHash,
		PlatformId:                updatePacket.PlatformId,
		DirectRtt:                 updatePacket.DirectMinRtt,
		DirectJitter:              updatePacket.DirectJitter,
		DirectPacketLoss:          updatePacket.DirectPacketLoss,
		NextRtt:                   updatePacket.NextMinRtt,
		NextJitter:                updatePacket.NextJitter,
		NextPacketLoss:            updatePacket.NextPacketLoss,
		ClientIpAddress:           udpAddrToAddress(updatePacket.ClientAddress),
		ServerIpAddress:           udpAddrToAddress(updatePacket.ServerAddress),
		ServerPrivateIpAddress:    udpAddrToAddress(serverData.Server.Addr),
		ClientRoutePublicKey:      updatePacket.ClientRoutePublicKey,
		ServerRoutePublicKey:      serverData.Server.PublicKey,
		Tag:                       updatePacket.Tag,
		NearRelays:                nearRelays,
		IssuedNearRelays:          issuedNearRelays,
		ConnectionType:            SessionConnectionType(updatePacket.ConnectionType),
		DatacenterId:              makeEntityID("Datacenter", serverData.Datacenter.ID),
		SequenceNumber:            updatePacket.Sequence,
		FallbackToDirect:          updatePacket.FallbackToDirect,
		VersionMajor:              serverData.VersionMajor,
		VersionMinor:              serverData.VersionMinor,
		VersionPatch:              serverData.VersionPatch,
		Location:                  Location{
			// CountryCode: location.CountryCode,
			Country: location.Country,
			Region: location.Region,
			City: location.City,
			Latitude: location.Latitude,
			Longitude: location.Longitude,
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
