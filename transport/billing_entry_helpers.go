package transport

import (
	"fmt"
	"net"
	"time"

	"github.com/networknext/backend/billing"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
)

func NewBillingEntry(
	routeRequest *billing.RouteRequest,
	route *routing.Route,
	routeType int,
	sameRoute bool,
	routingRulesSettings *routing.RoutingRulesSettings,
	routeDecision routing.Decision,
	packet *SessionUpdatePacket,
	sliceDuration uint64,
	timestampStart time.Time,
	timestampNow time.Time,
	initial bool) *billing.Entry {
	// Create billing slice flags
	sliceFlags := billing.RouteSliceFlagNone
	if routeType == routing.RouteTypeNew || routeType == routing.RouteTypeContinue {
		sliceFlags |= billing.RouteSliceFlagNext
	}

	if routing.IsVetoed(routeDecision) {
		sliceFlags |= billing.RouteSliceFlagVetoed
	}

	if packet.Flagged {
		sliceFlags |= billing.RouteSliceFlagReported
	}

	if packet.FallbackToDirect {
		sliceFlags |= billing.RouteSliceFlagFallbackToDirect
	}

	if (routeDecision.Reason & routing.DecisionHighPacketLossMultipath) != 0 {
		sliceFlags |= billing.RouteSliceFlagPacketLossMultipath
	}

	if (routeDecision.Reason & routing.DecisionHighJitterMultipath) != 0 {
		sliceFlags |= billing.RouteSliceFlagJitterMultipath
	}

	if (routeDecision.Reason & routing.DecisionRTTReductionMultipath) != 0 {
		sliceFlags |= billing.RouteSliceFlagRTTMultipath
	}

	usageBytesUp := (1000 * uint64(packet.KbpsUp)) / 8 * sliceDuration     // Converts Kbps to bytes
	usageBytesDown := (1000 * uint64(packet.KbpsDown)) / 8 * sliceDuration // Converts Kbps to bytes

	return &billing.Entry{
		Request:              routeRequest,
		Route:                NewBillingRoute(route, usageBytesUp, usageBytesDown),
		RouteDecision:        uint64(routeDecision.Reason),
		Duration:             sliceDuration,
		UsageBytesUp:         usageBytesUp,
		UsageBytesDown:       usageBytesDown,
		Timestamp:            uint64(timestampNow.Unix()),
		TimestampStart:       uint64(timestampStart.Unix()),
		PredictedRTT:         float32(route.Stats.RTT),
		PredictedJitter:      float32(route.Stats.Jitter),
		PredictedPacketLoss:  float32(route.Stats.PacketLoss),
		RouteChanged:         routeType != routing.RouteTypeContinue,
		NetworkNextAvailable: routeType == routing.RouteTypeNew || routeType == routing.RouteTypeContinue,
		Initial:              initial,
		EnvelopeBytesUp:      (1000 * uint64(routingRulesSettings.EnvelopeKbpsUp)) / 8 * sliceDuration,   // Converts Kbps to bytes
		EnvelopeBytesDown:    (1000 * uint64(routingRulesSettings.EnvelopeKbpsDown)) / 8 * sliceDuration, // Converts Kbps to bytes
		ConsideredRoutes:     []*billing.Route{},                                                         // Empty since not how new backend works and driven by disabled feature flag in old backend
		AcceptableRoutes:     []*billing.Route{},                                                         // Empty since not how new backend works and driven by disabled feature flag in old backend
		SameRoute:            sameRoute,
		OnNetworkNext:        routeDecision.OnNetworkNext,
		SliceFlags:           uint64(sliceFlags),
	}
}

func NewBillingRoute(route *routing.Route, bytesUp uint64, bytesDown uint64) []*billing.RouteHop {
	// To calculate price per hop, need the route, and to fetch seller + relay for each hop in that route
	var hops []*billing.RouteHop

	for _, relay := range route.Relays {
		// Get seller from relay
		seller := relay.Seller

		upIngress := seller.IngressPriceCents * bytesUp
		upEgress := seller.EgressPriceCents * bytesUp
		downIngress := seller.IngressPriceCents * bytesDown
		downEgress := seller.EgressPriceCents * bytesDown

		hops = append(hops, &billing.RouteHop{
			RelayID:      NewEntityID("Relay", relay.ID),
			SellerID:     &billing.EntityID{Kind: "Seller", Name: seller.ID},
			PriceIngress: int64(upIngress + downIngress),
			PriceEgress:  int64(upEgress + downEgress),
		})
	}

	return hops
}

// Convert new representation of data into old for billing entry
func NewRouteRequest(updatePacket *SessionUpdatePacket, buyer *routing.Buyer, serverData *ServerCacheEntry, location *routing.Location, storer storage.Storer, clientRelays []routing.Relay) *billing.RouteRequest {
	return &billing.RouteRequest{
		BuyerID:                NewEntityID("Buyer", buyer.ID),
		SessionID:              updatePacket.SessionID,
		UserHash:               updatePacket.UserHash,
		PlatformID:             updatePacket.PlatformID,
		DirectRTT:              updatePacket.DirectMinRTT,
		DirectJitter:           updatePacket.DirectJitter,
		DirectPacketLoss:       updatePacket.DirectPacketLoss,
		NextRTT:                updatePacket.NextMinRTT,
		NextJitter:             updatePacket.NextJitter,
		NextPacketLoss:         updatePacket.NextPacketLoss,
		ClientIpAddress:        NewBillingAddress(updatePacket.ClientAddress),
		ServerIpAddress:        NewBillingAddress(updatePacket.ServerAddress),
		ServerPrivateIpAddress: NewBillingAddress(serverData.Server.Addr),
		ClientRoutePublicKey:   updatePacket.ClientRoutePublicKey,
		ServerRoutePublicKey:   serverData.Server.PublicKey,
		Tag:                    updatePacket.Tag,
		NearRelays:             newNearRelayList(updatePacket, storer),
		IssuedNearRelays:       newIssuedNearRelayList(clientRelays),
		ConnectionType:         billing.SessionConnectionType(updatePacket.ConnectionType),
		DatacenterID:           NewEntityID("Datacenter", serverData.Datacenter.ID),
		SequenceNumber:         updatePacket.Sequence,
		FallbackToDirect:       updatePacket.FallbackToDirect,
		VersionMajor:           serverData.SDKVersion.Major,
		VersionMinor:           serverData.SDKVersion.Minor,
		VersionPatch:           serverData.SDKVersion.Patch,
		// Can't get commented out fields unless we pay for MaxMind Pro, see Monday task for more details:
		// https://network-next.monday.com/boards/359802870/pulses/474491141
		Location: &billing.Location{
			CountryCode: location.CountryCode,
			Country:     location.Country,
			Region:      location.Region,
			City:        location.City,
			Latitude:    float32(location.Latitude),
			Longitude:   float32(location.Longitude),
			Isp:         location.ISP,
			Asn:         int64(location.ASN),
			Continent:   location.Continent,
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
		UserFlags:                 updatePacket.UserFlags,
	}
}

// The list of relays the client actually believes it is close to / is using (should match issued near relays)
func newNearRelayList(updatePacket *SessionUpdatePacket, storer storage.Storer) []*billing.NearRelay {
	var nearRelays []*billing.NearRelay
	var i int32
	for i = 0; i < updatePacket.NumNearRelays; i++ {
		relay, err := storer.Relay(updatePacket.NearRelayIDs[i])
		if err != nil {
			continue
		}

		nearRelays = append(
			nearRelays,
			&billing.NearRelay{
				RelayID:    NewEntityID("Relay", relay.ID),
				RTT:        float64(updatePacket.NearRelayMinRTT[i]),
				Jitter:     float64(updatePacket.NearRelayJitter[i]),
				PacketLoss: float64(updatePacket.NearRelayPacketLoss[i]),
			},
		)
	}

	return nearRelays
}

// The list of relays we are telling the client is close to
func newIssuedNearRelayList(nearRelays []routing.Relay) []*billing.IssuedNearRelay {
	var issuedNearRelays []*billing.IssuedNearRelay
	for idx, nearRelay := range nearRelays {
		issuedNearRelays = append(issuedNearRelays, &billing.IssuedNearRelay{
			Index:          int32(idx),
			RelayID:        NewEntityID("Relay", nearRelay.ID),
			RelayIpAddress: NewBillingAddress(nearRelay.Addr),
		})
	}

	return issuedNearRelays
}

func NewBillingAddress(addr net.UDPAddr) *billing.Address {
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

func NewEntityID(kind string, ID uint64) *billing.EntityID {
	return &billing.EntityID{
		Kind: kind,
		Name: fmt.Sprintf("%x", ID),
	}
}
