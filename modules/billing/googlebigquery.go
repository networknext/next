package billing

import (
	"context"
	"errors"
	"sync"

	"cloud.google.com/go/bigquery"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/networknext/backend/modules/metrics"
)

const (
	DefaultBigQueryBatchSize   = 1000
	DefaultBigQueryChannelSize = 10000
)

type GoogleBigQueryClient struct {
	Metrics       metrics.PublisherMetrics
	Logger        log.Logger
	TableInserter *bigquery.Inserter
	BatchSize     int

	buffer      []*BillingEntry
	bufferMutex sync.RWMutex

	entries chan *BillingEntry
}

// Bill pushes an Entry to the channel
func (bq *GoogleBigQueryClient) Bill(ctx context.Context, entry *BillingEntry) error {
	bq.Metrics.EntriesSubmitted.Add(1)
	if bq.entries == nil {
		bq.entries = make(chan *BillingEntry, DefaultBigQueryChannelSize)
	}

	bq.bufferMutex.RLock()
	bufferLength := len(bq.buffer)
	bq.bufferMutex.RUnlock()

	if bufferLength >= bq.BatchSize {
		return errors.New("entries buffer full")
	}

	select {
	case bq.entries <- entry:
		return nil
	default:
		return errors.New("entries channel full")
	}
}

// WriteLoop ranges over the incoming channel of Entry types and fills an internal buffer.
// Once the buffer fills to the BatchSize it will write all entries to BigQuery. This should
// only be called from 1 goroutine to avoid using a mutex around the internal buffer slice
func (bq *GoogleBigQueryClient) WriteLoop(ctx context.Context) error {
	if bq.entries == nil {
		bq.entries = make(chan *BillingEntry, DefaultBigQueryChannelSize)
	}
	for entry := range bq.entries {
		bq.Metrics.EntriesQueued.Set(float64(len(bq.entries)))

		bq.bufferMutex.Lock()
		bq.buffer = append(bq.buffer, entry)
		bufferLength := len(bq.buffer)
		if bufferLength >= bq.BatchSize {
			if err := bq.TableInserter.Put(ctx, bq.buffer); err != nil {
				bq.bufferMutex.Unlock()

				level.Error(bq.Logger).Log("msg", "failed to write to BigQuery", "err", err)
				bq.Metrics.PublishFailure.Add(float64(bufferLength))
				continue
			}

			bq.buffer = bq.buffer[:0]

			level.Info(bq.Logger).Log("msg", "flushed entries to BigQuery", "size", bq.BatchSize, "total", bufferLength)
			bq.Metrics.EntriesFlushed.Add(float64(bufferLength))
		}

		bq.bufferMutex.Unlock()
	}
	return nil
}

// Save implements the bigquery.ValueSaver interface for an Entry
// so it can be used in Put()
func (entry *BillingEntry) Save() (map[string]bigquery.Value, string, error) {

	e := make(map[string]bigquery.Value)

	e["timestamp"] = int(entry.Timestamp)
	e["buyerId"] = int(entry.BuyerID)
	e["sessionId"] = int(entry.SessionID)
	e["sliceNumber"] = int(entry.SliceNumber)
	e["directRTT"] = entry.DirectRTT
	e["directJitter"] = entry.DirectJitter
	e["directPacketLoss"] = entry.DirectPacketLoss
	e["userHash"] = int(entry.UserHash)

	if entry.Next {
		e["next"] = entry.Next
		e["nextRTT"] = entry.NextRTT
		e["nextJitter"] = entry.NextJitter
		e["nextPacketLoss"] = entry.NextPacketLoss
	}

	nextRelays := make([]bigquery.Value, entry.NumNextRelays)
	for i := 0; i < int(entry.NumNextRelays); i++ {
		nextRelays[i] = int(entry.NextRelays[i])
	}
	e["nextRelays"] = nextRelays

	e["totalPrice"] = int(entry.TotalPrice)

	if entry.ClientToServerPacketsLost > 0 {
		e["clientToServerPacketsLost"] = int(entry.ClientToServerPacketsLost)
	}

	if entry.ServerToClientPacketsLost > 0 {
		e["serverToClientPacketsLost"] = int(entry.ServerToClientPacketsLost)
	}

	e["committed"] = entry.Committed
	e["flagged"] = entry.Flagged
	e["multipath"] = entry.Multipath

	if entry.Next {
		e["initial"] = entry.Initial
		e["nextBytesUp"] = int(entry.NextBytesUp)
		e["nextBytesDown"] = int(entry.NextBytesDown)
		e["envelopeBytesUp"] = int(entry.EnvelopeBytesUp)
		e["envelopeBytesDown"] = int(entry.EnvelopeBytesDown)
	}

	e["datacenterID"] = int(entry.DatacenterID)

	if entry.Next {
		e["rttReduction"] = entry.RTTReduction
		e["packetLossReduction"] = entry.PacketLossReduction
	}

	nextRelaysPrice := make([]bigquery.Value, entry.NumNextRelays)
	for i := 0; i < int(entry.NumNextRelays); i++ {
		nextRelaysPrice[i] = int(entry.NextRelaysPrice[i])
	}
	e["nextRelaysPrice"] = nextRelaysPrice

	e["latitude"] = entry.Latitude
	e["longitude"] = entry.Longitude
	e["isp"] = entry.ISP
	e["abTest"] = entry.ABTest
	e["routeDecision"] = int(entry.RouteDecision)

	e["connectionType"] = int(entry.ConnectionType)
	e["platformType"] = int(entry.PlatformType)
	e["sdkVersion"] = entry.SDKVersion

	if entry.PacketLoss > 0.0 {
		e["packetLoss"] = entry.PacketLoss
	}

	if entry.PredictedNextRTT > 0.0 {
		e["predictedNextRTT"] = entry.PredictedNextRTT
	}

	e["multipathVetoed"] = entry.MultipathVetoed

	if entry.UseDebug && entry.Debug != "" {
		e["debug"] = entry.Debug
	}

	e["fallbackToDirect"] = entry.FallbackToDirect

	if entry.ClientFlags != 0 {
		e["clientFlags"] = int(entry.ClientFlags)
	}

	if entry.UserFlags != 0 {
		e["userFlags"] = int(entry.UserFlags)
	}

	if entry.NearRelayRTT != 0 {
		e["nearRelayRTT"] = entry.NearRelayRTT
	}

	if entry.PacketsOutOfOrderClientToServer != 0 {
		e["packetsOutOfOrderClientToServer"] = int(entry.PacketsOutOfOrderClientToServer)
	}

	if entry.PacketsOutOfOrderServerToClient != 0 {
		e["packetsOutOfOrderServerToClient"] = int(entry.PacketsOutOfOrderServerToClient)
	}

	if entry.JitterClientToServer != 0 {
		e["jitterClientToServer"] = entry.JitterClientToServer
	}

	if entry.JitterServerToClient != 0 {
		e["jitterServerToClient"] = entry.JitterServerToClient
	}

	if entry.UseDebug {
		if entry.NumNearRelays != 0 {
			e["numNearRelays"] = int(entry.NumNearRelays)

			nearRelayIDs := make([]bigquery.Value, entry.NumNearRelays)
			for i := 0; i < int(entry.NumNearRelays); i++ {
				nearRelayIDs[i] = int(entry.NearRelayIDs[i])
			}
			e["nearRelayIDs"] = nearRelayIDs

			nearRelayRTTs := make([]bigquery.Value, entry.NumNearRelays)
			for i := 0; i < int(entry.NumNearRelays); i++ {
				nearRelayRTTs[i] = entry.NearRelayRTTs[i]
			}
			e["nearRelayRTTs"] = nearRelayRTTs

			nearRelayJitters := make([]bigquery.Value, entry.NumNearRelays)
			for i := 0; i < int(entry.NumNearRelays); i++ {
				nearRelayJitters[i] = entry.NearRelayJitters[i]
			}
			e["nearRelayJitters"] = nearRelayJitters

			nearRelayPacketLosses := make([]bigquery.Value, entry.NumNearRelays)
			for i := 0; i < int(entry.NumNearRelays); i++ {
				nearRelayPacketLosses[i] = entry.NearRelayPacketLosses[i]
			}
			e["nearRelayPacketLosses"] = nearRelayPacketLosses
		}
	}

	e["relayWentAway"] = entry.RelayWentAway
	e["routeLost"] = entry.RouteLost

	if entry.NumTags > 0 {
		tags := make([]bigquery.Value, entry.NumTags)
		for i := 0; i < int(entry.NumTags); i++ {
			tags[i] = int(entry.Tags[i])
		}
		e["tags"] = tags
	}

	e["mispredicted"] = entry.Mispredicted
	e["vetoed"] = entry.Vetoed

	e["latencyWorse"] = entry.LatencyWorse
	e["noRoute"] = entry.NoRoute
	e["nextLatencyTooHigh"] = entry.NextLatencyTooHigh
	e["routeChanged"] = entry.RouteChanged
	e["commitVeto"] = entry.CommitVeto

	if entry.RouteDiversity > 0 {
		e["routeDiversity"] = entry.RouteDiversity
	}

	if entry.LackOfDiversity {
		e["lackOfDiversity"] = entry.LackOfDiversity
	}

	if entry.Pro {
		e["pro"] = entry.Pro
	}

	if entry.MultipathRestricted {
		e["multipathRestricted"] = entry.MultipathRestricted
	}

	return e, "", nil
}
