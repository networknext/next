package billing

import (
	"context"

	"cloud.google.com/go/bigquery"
)

type GoogleBigQueryClient struct {
	TableInserter *bigquery.Inserter
}

func (bq *GoogleBigQueryClient) Bill(ctx context.Context, sessionID uint64, entry *Entry) error {
	return bq.TableInserter.Put(ctx, entry)
}

// Save implements the bigquery.ValueSaver interface for an Entry
// so it can be used in Put()
func (entry *Entry) Save() (map[string]bigquery.Value, string, error) {
	e := make(map[string]bigquery.Value)

	e["buyerId"] = entry.Request.BuyerID.String()
	e["sessionId"] = int(entry.Request.SessionID)
	e["userId"] = entry.Request.UserHash
	e["platformId"] = int(entry.Request.PlatformID)

	e["directRtt"] = entry.Request.DirectRTT
	e["directJitter"] = entry.Request.DirectJitter
	e["directPacketLoss"] = entry.Request.DirectPacketLoss
	e["nextRtt"] = entry.Request.NextRTT
	e["nextJitter"] = entry.Request.NextJitter
	e["nextPacketLoss"] = entry.Request.NextPacketLoss
	e["predictedRtt"] = entry.PredictedRTT
	e["predictedJitter"] = entry.PredictedJitter
	e["predictedPacketLoss"] = entry.PredictedPacketLoss

	e["packetsLostClientToServer"] = int(entry.Request.PacketsLostClientToServer)
	e["packetsLostServerToClient"] = int(entry.Request.PacketsLostServerToClient)

	e["clientIpAddress"] = entry.Request.ClientIpAddress.String()
	e["serverIpAddress"] = entry.Request.ServerIpAddress.String()
	e["serverPrivateIpAddress"] = entry.Request.ServerPrivateIpAddress.String()

	e["tag"] = int(entry.Request.Tag)

	nrs := make([]map[string]bigquery.Value, len(entry.Request.NearRelays))
	for idx := range entry.Request.NearRelays {
		nrs[idx] = make(map[string]bigquery.Value)
		nrs[idx]["id"] = entry.Request.NearRelays[idx].RelayID.String()
		nrs[idx]["rtt"] = entry.Request.NearRelays[idx].RTT
		nrs[idx]["jitter"] = entry.Request.NearRelays[idx].Jitter
		nrs[idx]["packetLoss"] = entry.Request.NearRelays[idx].PacketLoss
	}
	e["nearRelays"] = nrs

	inrs := make([]map[string]bigquery.Value, len(entry.Request.IssuedNearRelays))
	for idx := range entry.Request.IssuedNearRelays {
		inrs[idx] = make(map[string]bigquery.Value)
		inrs[idx]["index"] = int(entry.Request.IssuedNearRelays[idx].Index)
		inrs[idx]["id"] = entry.Request.IssuedNearRelays[idx].RelayID.String()
		inrs[idx]["ipAddress"] = entry.Request.IssuedNearRelays[idx].RelayIpAddress.String()
	}
	e["issuedNearRelays"] = inrs

	e["connectionType"] = int(entry.Request.ConnectionType)
	e["sequenceNumber"] = int(entry.Request.SequenceNumber)
	e["datacenterId"] = entry.Request.DatacenterID.String()
	e["fallbackToDirect"] = entry.Request.FallbackToDirect

	e["versionMajor"] = int(entry.Request.VersionMajor)
	e["versionMinor"] = int(entry.Request.VersionMinor)
	e["versionPatch"] = int(entry.Request.VersionPatch)

	e["kbpsUp"] = int(entry.Request.UsageKbpsUp)
	e["kbpsDown"] = int(entry.Request.UsageKbpsDown)

	e["countryCode"] = entry.Request.Location.CountryCode
	e["country"] = entry.Request.Location.Country
	e["region"] = entry.Request.Location.Region
	e["city"] = entry.Request.Location.City
	e["latitude"] = entry.Request.Location.Latitude
	e["longitude"] = entry.Request.Location.Longitude
	e["isp"] = entry.Request.Location.Isp

	selectedRoute := make([]map[string]bigquery.Value, len(entry.Route))
	for idx := range entry.Route {
		selectedRoute[idx] = make(map[string]bigquery.Value)
		selectedRoute[idx]["id"] = entry.Route[idx].RelayID.String()
		selectedRoute[idx]["sellerId"] = entry.Route[idx].SellerID.String()
		selectedRoute[idx]["priceIngress"] = int(entry.Route[idx].PriceIngress)
		selectedRoute[idx]["priceEgress"] = int(entry.Route[idx].PriceEgress)
	}
	e["route"] = selectedRoute

	acceptableRoutes := make([]map[string]map[string]bigquery.Value, len(entry.AcceptableRoutes))
	for ridx, route := range entry.AcceptableRoutes {
		acceptableRoutes[ridx] = make(map[string]map[string]bigquery.Value)
		acceptableRoutes[ridx]["route"] = make(map[string]bigquery.Value)
		for idx := range route.Route {
			acceptableRoutes[ridx]["route"]["id"] = entry.Route[idx].RelayID.String()
			acceptableRoutes[ridx]["route"]["sellerId"] = entry.Route[idx].SellerID.String()
			acceptableRoutes[ridx]["route"]["priceIngress"] = int(entry.Route[idx].PriceIngress)
			acceptableRoutes[ridx]["route"]["priceEgress"] = int(entry.Route[idx].PriceEgress)
		}
	}
	e["acceptableRoutes"] = acceptableRoutes

	consideredRoutes := make([]map[string]map[string]bigquery.Value, len(entry.ConsideredRoutes))
	for ridx, route := range entry.ConsideredRoutes {
		consideredRoutes[ridx] = make(map[string]map[string]bigquery.Value)
		consideredRoutes[ridx]["route"] = make(map[string]bigquery.Value)
		for idx := range route.Route {
			consideredRoutes[ridx]["route"]["id"] = entry.Route[idx].RelayID.String()
			consideredRoutes[ridx]["route"]["sellerId"] = entry.Route[idx].SellerID.String()
			consideredRoutes[ridx]["route"]["priceIngress"] = int(entry.Route[idx].PriceIngress)
			consideredRoutes[ridx]["route"]["priceEgress"] = int(entry.Route[idx].PriceEgress)
		}
	}
	e["consideredRoutes"] = consideredRoutes

	e["routeDecision"] = int(entry.RouteDecision)
	e["routeChanged"] = entry.RouteChanged
	e["sameRoute"] = entry.SameRoute
	e["networkNext"] = entry.NetworkNextAvailable
	e["initial"] = entry.Initial
	e["flagged"] = entry.Request.Flagged
	e["tryBeforeYouBuy"] = entry.Request.TryBeforeYouBuy

	e["duration"] = int(entry.Duration)
	e["bytesUp"] = int(entry.EnvelopeBytesUp)
	e["bytesDown"] = int(entry.EnvelopeBytesDown)

	e["timestamp"] = entry.Timestamp
	e["timestampStart"] = entry.TimestampStart

	e["committed"] = entry.Request.Committed
	e["user_flags"] = entry.Request.UserFlags

	return e, "", nil
}
