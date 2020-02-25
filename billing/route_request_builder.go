package billing

import (
	"net"
	"strconv"

	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport"
)

// Convert new representation of data into old for billing entry
func BuildRouteRequest(updatePacket transport.SessionUpdatePacket, buyer routing.Buyer, serverData transport.ServerCacheEntry, location routing.Location, storer storage.Storer, nearRelays []routing.Relay) (RouteRequest, error) {
	issuedNearRelays, err := buildIssuedNearRelayList(nearRelays)
	if err != nil {
		return RouteRequest{}, err
	}

	return RouteRequest{
		BuyerId:                makeEntityID("Buyer", buyer.Key),
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
		IssuedNearRelays:       issuedNearRelays,
		ConnectionType:         SessionConnectionType(updatePacket.ConnectionType),
		DatacenterId:           makeEntityID("Datacenter", serverData.Datacenter.ID),
		SequenceNumber:         updatePacket.Sequence,
		FallbackToDirect:       updatePacket.FallbackToDirect,
		VersionMajor:           serverData.VersionMajor,
		VersionMinor:           serverData.VersionMinor,
		VersionPatch:           serverData.VersionPatch,
		Location: Location{
			// CountryCode: location.CountryCode,
			Country:   location.Country,
			Region:    location.Region,
			City:      location.City,
			Latitude:  location.Latitude,
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
	}, nil
}

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

func buildIssuedNearRelayList(nearRelays []routing.Relay) ([]*IssuedNearRelay, error) {
	var issuedNearRelays []*IssuedNearRelay
	for idx, nearRelay := range nearRelays {
		var address *Address
		ipStr, portStr, err := net.SplitHostPort(string(nearRelay.Addr))
		if err == nil {
			ip := net.ParseIP(ipStr)
			port, err := strconv.Atoi(portStr)
			if ip != nil && err == nil {
				address = udpAddrToAddress(net.UDPAddr{
					IP:   ip,
					Port: port,
				})
			} else {
				if err != nil {
					common.Error(ctx, "ServerIngress", "near relay is missing port, addr '%s', err: %v", string(nearRelay.Address), err)
				}
				if ip == nil {
					common.Error(ctx, "ServerIngress", "near relay is missing IP, addr '%s'", string(nearRelay.Address))
				}
			}
		} else {
			common.Error(ctx, "ServerIngress", "near relay address is not parsable, addr '%s', err: %v", string(nearRelay.Address), err)
		}

		issuedNearRelays = append(issuedNearRelays, &IssuedNearRelay{
			Index:          int32(idx),
			RelayId:        nearRelay.RelayId,
			RelayIpAddress: address,
		})
	}
}
