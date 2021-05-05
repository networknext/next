package billing

import (
	"context"
	"errors"
	"fmt"
	"math"
	"strings"
	"sync"

	"cloud.google.com/go/bigquery"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"

	"github.com/networknext/backend/modules/metrics"
)

const (
	DefaultBigQueryBatchSize   = 1000
	DefaultBigQueryChannelSize = 20000
)

type GoogleBigQueryClient struct {
	Metrics       *metrics.BillingMetrics
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

	hasNanOrInf, nanOrInfFields := entry.CheckNaNOrInf()
	if hasNanOrInf {
		bq.Metrics.ErrorMetrics.BillingEntriesWithNaN.Add(1)
		fieldStr := strings.Join(nanOrInfFields, " ")
		fmt.Printf("Warn: billing entry had NaN or Inf values for %v.\n%+v\n", nanOrInfFields, entry)
		level.Warn(bq.Logger).Log("msg", "Billing entry had NaN or Inf values", "fields", fieldStr)
	}

	if !entry.Validate() {
		bq.Metrics.ErrorMetrics.BillingInvalidEntries.Add(1)
		fmt.Printf("Error: billing entry not valid.\n%+v\n", entry)
		return errors.New("invalid billing entry")
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
func (bq *GoogleBigQueryClient) WriteLoop(ctx context.Context, wg *sync.WaitGroup) error {
	defer wg.Done()

	if bq.entries == nil {
		bq.entries = make(chan *BillingEntry, DefaultBigQueryChannelSize)
	}
	for {
		select {
		case entry := <-bq.entries:
			bq.Metrics.EntriesQueued.Set(float64(len(bq.entries)))
			bq.bufferMutex.Lock()
			bq.buffer = append(bq.buffer, entry)
			bufferLength := len(bq.buffer)

			if bufferLength >= bq.BatchSize {
				if err := bq.TableInserter.Put(context.Background(), bq.buffer); err != nil {
					bq.bufferMutex.Unlock()

					level.Error(bq.Logger).Log("msg", "failed to write to BigQuery", "err", err)
					fmt.Printf("Failed to write to BigQuery: %v\n", err)

					bq.Metrics.ErrorMetrics.BillingWriteFailure.Add(float64(bufferLength))
					continue
				}

				bq.buffer = bq.buffer[:0]
				level.Info(bq.Logger).Log("msg", "flushed entries to BigQuery", "size", bq.BatchSize, "total", bufferLength)
				bq.Metrics.BillingEntrySize.Set(float64(bufferLength * MaxBillingEntryBytes))
				bq.Metrics.EntriesFlushed.Add(float64(bufferLength))
			}

			bq.bufferMutex.Unlock()
		case <-ctx.Done():
			var bufferLength int

			// Received shutdown signal, write remaining entries to BigQuery
			bq.bufferMutex.Lock()
			for entry := range bq.entries {
				// Add the remaining entries to the buffer
				bq.buffer = append(bq.buffer, entry)
				bufferLength = len(bq.buffer)
				bq.Metrics.EntriesQueued.Set(float64(bufferLength))
			}

			// Emptied out the entries channel, flush to BigQuery
			if err := bq.TableInserter.Put(context.Background(), bq.buffer); err != nil {
				bq.bufferMutex.Unlock()

				level.Error(bq.Logger).Log("msg", "failed to write to BigQuery", "err", err)
				fmt.Printf("Failed to write to BigQuery: %v\n", err)

				bq.Metrics.ErrorMetrics.BillingWriteFailure.Add(float64(bufferLength))
				return err
			}
			bq.buffer = bq.buffer[:0]
			bq.bufferMutex.Unlock()

			level.Info(bq.Logger).Log("msg", "flushed entries to BigQuery", "size", bufferLength, "total", bufferLength)
			fmt.Printf("Final flush of %d entries to BigQuery.\n", bufferLength)

			bq.Metrics.EntriesFlushed.Add(float64(bufferLength))

			return nil
		}
	}
	return nil
}

// Closes the entries channel. Should only be done by the entry sender.
func (bq *GoogleBigQueryClient) Close() {
	close(bq.entries)
}

// Validate a billing entry. Returns true if the billing entry is valid, false if invalid.
func (entry *BillingEntry) Validate() bool {

	if entry.Timestamp < 0 {
		fmt.Printf("invalid timestamp\n")
		return false
	}

	if entry.BuyerID == 0 {
		fmt.Printf("invalid buyer id\n")
		return false
	}

	if entry.SessionID == 0 {
		fmt.Printf("invalid session id\n")
		return false
	}

	// IMPORTANT: Logic inverted because comparing a NaN float value always returns false
	if !(entry.Latitude >= -90.0 && entry.Latitude <= +90.0) {
		fmt.Printf("invalid latitude\n")
		return false
	}

	if !(entry.Longitude >= -180.0 && entry.Longitude <= +180.0) {
		fmt.Printf("invalid longitude\n")
		return false
	}

	// IMPORTANT: You must update this check if you ever add a new connection type in the SDK
	if entry.ConnectionType < 0 || entry.ConnectionType > 3 {
		fmt.Printf("invalid connection type\n")
		return false
	}

	// IMPORTANT: You must update this check when you add new platforms to the SDK
	if entry.PlatformType < 0 || entry.ConnectionType > 10 {
		fmt.Printf("invalid connection type\n")
		return false
	}

	if entry.RouteDiversity < 0 || entry.RouteDiversity > 32 {
		fmt.Printf("invalid route diversity\n")
		return false
	}

	if !(entry.DirectRTT >= 0.0 && entry.DirectRTT <= 10000.0) {
		fmt.Printf("invalid direct rtt\n")
		return false
	}

	if !(entry.DirectJitter >= 0.0 && entry.DirectJitter <= 10000.0) {
		fmt.Printf("invalid direct jitter\n")
		return false
	}

	if !(entry.DirectPacketLoss >= 0.0 && entry.DirectPacketLoss <= 100.0) {
		fmt.Printf("invalid direct packet loss\n")
		return false
	}

	if !(entry.PacketLoss >= 0.0 && entry.PacketLoss <= 100.0) {
		if entry.PacketLoss > 100.0 {
			fmt.Printf("PacketLoss %v > 100.0. Clamping to 100.0\n%+v\n", entry.PacketLoss, entry)
			entry.PacketLoss = 100.0
		} else {
			fmt.Printf("invalid packet loss\n")
			return false
		}
	}

	if !(entry.JitterClientToServer >= 0.0 && entry.JitterClientToServer <= 1000.0) {
		if entry.JitterClientToServer > 1000.0 {
			fmt.Printf("JitterClientToServer %v > 1000.0. Clamping to 1000.0\n%+v\n", entry.JitterClientToServer, entry)
			entry.JitterClientToServer = 1000.0
		} else {
			fmt.Printf("invalid jitter client to server\n")
			return false
		}
	}

	if !(entry.JitterServerToClient >= 0.0 && entry.JitterServerToClient <= 1000.0) {
		if entry.JitterServerToClient > 1000.0 {
			fmt.Printf("JitterServerToClient %v > 1000.0. Clamping to 1000.0.\n%+v\n", entry.JitterServerToClient, entry)
			entry.JitterServerToClient = 1000.0
		} else {
			fmt.Printf("invalid jitter server to client\n")
			return false
		}
	}

	if entry.NumTags < 0 || entry.NumTags > 8 {
		fmt.Printf("invalid route diversity\n")
		return false
	}

	if entry.Next {

		if !(entry.NextRTT >= 0.0 && entry.NextRTT <= 10000.0) {
			fmt.Printf("invalid next rtt\n")
			return false
		}

		if !(entry.NextJitter >= 0.0 && entry.NextJitter <= 10000.0) {
			fmt.Printf("invalid next jitter\n")
			return false
		}

		if !(entry.NextPacketLoss >= 0.0 && entry.NextPacketLoss <= 100.0) {
			fmt.Printf("invalid next packet loss\n")
			return false
		}

		if !(entry.PredictedNextRTT >= 0.0 && entry.PredictedNextRTT <= 10000.0) {
			fmt.Printf("invalid predicted next rtt\n")
			return false
		}

		if entry.NumNearRelays < 0 || entry.NumNearRelays > 32 {
			fmt.Printf("invalid num near relays\n")
			return false
		}

		if entry.NumNearRelays > 0 {

			for i := 0; i < int(entry.NumNearRelays); i++ {

				if entry.NearRelayIDs[i] == 0 {
					// Log this but do not return false
					// TODO: investigate why nearRelayID is 0
					fmt.Printf("NearRelayIDs[%d] is 0.\n%+v\n", i, entry)
				}

				if !(entry.NearRelayRTTs[i] >= 0.0 && entry.NearRelayRTTs[i] <= 255.0) {
					fmt.Printf("invalid near relay rtt\n")
					return false
				}

				if !(entry.NearRelayJitters[i] >= 0.0 && entry.NearRelayJitters[i] <= 255.0) {
					fmt.Printf("invalid near relay jitter\n")
					return false
				}

				if !(entry.NearRelayPacketLosses[i] >= 0.0 && entry.NearRelayPacketLosses[i] <= 100.0) {
					fmt.Printf("invalid near relay packet loss\n")
					return false
				}
			}
		}
	}

	return true
}

// Checks all floating point numbers for NaN and forces them to 0
func (entry *BillingEntry) CheckNaNOrInf() (bool, []string) {
	var nanOrInfExists bool
	var nanOrInfFields []string

	if math.IsNaN(float64(entry.NextRTT)) || math.IsInf(float64(entry.NextRTT), 0) {
		nanOrInfFields = append(nanOrInfFields, "NextRTT")
		nanOrInfExists = true
		entry.NextRTT = float32(0)
	}

	if math.IsNaN(float64(entry.NextJitter)) || math.IsInf(float64(entry.NextJitter), 0) {
		nanOrInfFields = append(nanOrInfFields, "NextJitter")
		nanOrInfExists = true
		entry.NextJitter = float32(0)
	}

	if math.IsNaN(float64(entry.NextPacketLoss)) || math.IsInf(float64(entry.NextPacketLoss), 0) {
		nanOrInfFields = append(nanOrInfFields, "NextPacketLoss")
		nanOrInfExists = true
		entry.NextPacketLoss = float32(0)
	}

	if math.IsNaN(float64(entry.Latitude)) || math.IsInf(float64(entry.Latitude), 0) {
		nanOrInfFields = append(nanOrInfFields, "Latitude")
		nanOrInfExists = true
		entry.Latitude = float32(0)
	}

	if math.IsNaN(float64(entry.Longitude)) || math.IsInf(float64(entry.Longitude), 0) {
		nanOrInfFields = append(nanOrInfFields, "Longitude")
		nanOrInfExists = true
		entry.Longitude = float32(0)
	}

	if math.IsNaN(float64(entry.PacketLoss)) || math.IsInf(float64(entry.PacketLoss), 0) {
		nanOrInfFields = append(nanOrInfFields, "PacketLoss")
		nanOrInfExists = true
		entry.PacketLoss = float32(0)
	}

	if math.IsNaN(float64(entry.PredictedNextRTT)) || math.IsInf(float64(entry.PredictedNextRTT), 0) {
		nanOrInfFields = append(nanOrInfFields, "PredictedNextRTT")
		nanOrInfExists = true
		entry.PredictedNextRTT = float32(0)
	}

	if math.IsNaN(float64(entry.NearRelayRTT)) || math.IsInf(float64(entry.NearRelayRTT), 0) {
		nanOrInfFields = append(nanOrInfFields, "NearRelayRTT")
		nanOrInfExists = true
		entry.NearRelayRTT = float32(0)
	}

	if math.IsNaN(float64(entry.JitterClientToServer)) || math.IsInf(float64(entry.JitterClientToServer), 0) {
		nanOrInfFields = append(nanOrInfFields, "JitterClientToServer")
		nanOrInfExists = true
		entry.JitterClientToServer = float32(0)
	}

	if math.IsNaN(float64(entry.JitterServerToClient)) || math.IsInf(float64(entry.JitterServerToClient), 0) {
		nanOrInfFields = append(nanOrInfFields, "JitterServerToClient")
		nanOrInfExists = true
		entry.JitterServerToClient = float32(0)
	}

	if entry.NumNearRelays > 0 {
		for i := 0; i < int(entry.NumNearRelays); i++ {
			if math.IsNaN(float64(entry.NearRelayRTTs[i])) || math.IsInf(float64(entry.NearRelayRTTs[i]), 0) {
				nanOrInfFields = append(nanOrInfFields, fmt.Sprintf("NearRelayRTTs[%d]", i))
				nanOrInfExists = true
				entry.NearRelayRTTs[i] = float32(0)
			}

			if math.IsNaN(float64(entry.NearRelayJitters[i])) || math.IsInf(float64(entry.NearRelayJitters[i]), 0) {
				nanOrInfFields = append(nanOrInfFields, fmt.Sprintf("NearRelayJitters[%d]", i))
				nanOrInfExists = true
				entry.NearRelayJitters[i] = float32(0)
			}

			if math.IsNaN(float64(entry.NearRelayPacketLosses[i])) || math.IsInf(float64(entry.NearRelayPacketLosses[i]), 0) {
				nanOrInfFields = append(nanOrInfFields, fmt.Sprintf("NearRelayPacketLosses[%d]", i))
				nanOrInfExists = true
				entry.NearRelayPacketLosses[i] = float32(0)
			}
		}
	}

	return nanOrInfExists, nanOrInfFields
}

// Implements the bigquery.ValueSaver interface for a billing entry so it can be used in Put()
func (entry *BillingEntry) Save() (map[string]bigquery.Value, string, error) {

	e := make(map[string]bigquery.Value)

	e["timestamp"] = int(entry.Timestamp)

	e["buyerId"] = int(entry.BuyerID)
	e["sessionId"] = int(entry.SessionID)
	e["datacenterID"] = int(entry.DatacenterID)
	e["userHash"] = int(entry.UserHash)
	e["latitude"] = entry.Latitude
	e["longitude"] = entry.Longitude
	e["isp"] = entry.ISP
	e["connectionType"] = int(entry.ConnectionType)
	e["platformType"] = int(entry.PlatformType)
	e["sdkVersion"] = entry.SDKVersion

	e["sliceNumber"] = int(entry.SliceNumber)

	e["flagged"] = entry.Flagged
	e["fallbackToDirect"] = entry.FallbackToDirect
	e["multipathVetoed"] = entry.MultipathVetoed
	e["abTest"] = entry.ABTest
	e["committed"] = entry.Committed
	e["multipath"] = entry.Multipath
	e["rttReduction"] = entry.RTTReduction
	e["packetLossReduction"] = entry.PacketLossReduction
	e["relayWentAway"] = entry.RelayWentAway
	e["routeLost"] = entry.RouteLost
	e["mispredicted"] = entry.Mispredicted
	e["vetoed"] = entry.Vetoed
	e["latencyWorse"] = entry.LatencyWorse
	e["noRoute"] = entry.NoRoute
	e["nextLatencyTooHigh"] = entry.NextLatencyTooHigh
	e["routeChanged"] = entry.RouteChanged
	e["commitVeto"] = entry.CommitVeto

	if entry.Pro {
		e["pro"] = entry.Pro
	}

	if entry.BuyerNotLive {
		e["buyerNotLive"] = entry.BuyerNotLive
	}

	if entry.UnknownDatacenter {
		e["unknownDatacenter"] = entry.UnknownDatacenter
	}

	if entry.DatacenterNotEnabled {
		e["datacenterNotEnabled"] = entry.DatacenterNotEnabled
	}

	if entry.StaleRouteMatrix {
		e["staleRouteMatrix"] = entry.StaleRouteMatrix
	}

	if entry.RouteDiversity > 0 {
		e["routeDiversity"] = entry.RouteDiversity
	}

	if entry.LackOfDiversity {
		e["lackOfDiversity"] = entry.LackOfDiversity
	}

	if entry.MultipathRestricted {
		e["multipathRestricted"] = entry.MultipathRestricted
	}

	e["directRTT"] = entry.DirectRTT
	e["directJitter"] = entry.DirectJitter
	e["directPacketLoss"] = entry.DirectPacketLoss

	if entry.ClientToServerPacketsSent > 0 {
		e["clientToServerPacketsSent"] = int(entry.ClientToServerPacketsSent)
	}

	if entry.ServerToClientPacketsSent > 0 {
		e["serverToClientPacketsSent"] = int(entry.ServerToClientPacketsSent)
	}

	if entry.ClientToServerPacketsLost > 0 {
		e["clientToServerPacketsLost"] = int(entry.ClientToServerPacketsLost)
	}

	if entry.ServerToClientPacketsLost > 0 {
		e["serverToClientPacketsLost"] = int(entry.ServerToClientPacketsLost)
	}

	// IMPORTANT: This is derived from *PacketsSent and *PacketsLost, and is valid for both next and direct
	if entry.PacketLoss > 0.0 {
		e["packetLoss"] = entry.PacketLoss
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

	if entry.UseDebug && entry.Debug != "" {
		e["debug"] = entry.Debug
	}

	if entry.ClientFlags != 0 {
		e["clientFlags"] = int(entry.ClientFlags)
	}

	if entry.UserFlags != 0 {
		e["userFlags"] = int(entry.UserFlags)
	}

	if entry.NumTags > 0 {
		tags := make([]bigquery.Value, entry.NumTags)
		for i := 0; i < int(entry.NumTags); i++ {
			tags[i] = int(entry.Tags[i])
		}
		e["tags"] = tags
	}

	if entry.Next {

		e["next"] = entry.Next

		e["nextRTT"] = entry.NextRTT
		e["nextJitter"] = entry.NextJitter
		e["nextPacketLoss"] = entry.NextPacketLoss

		e["totalPrice"] = int(entry.TotalPrice)

		e["nextBytesUp"] = int(entry.NextBytesUp)
		e["nextBytesDown"] = int(entry.NextBytesDown)
		e["envelopeBytesUp"] = int(entry.EnvelopeBytesUp)
		e["envelopeBytesDown"] = int(entry.EnvelopeBytesDown)

		if entry.PredictedNextRTT > 0.0 {
			e["predictedNextRTT"] = entry.PredictedNextRTT
		}

		if entry.NearRelayRTT != 0 {
			e["nearRelayRTT"] = entry.NearRelayRTT
		}

		nextRelays := make([]bigquery.Value, entry.NumNextRelays)
		nextRelaysPrice := make([]bigquery.Value, entry.NumNextRelays)

		for i := 0; i < int(entry.NumNextRelays); i++ {
			nextRelays[i] = int(entry.NextRelays[i])
			nextRelaysPrice[i] = int(entry.NextRelaysPrice[i])
		}

		e["nextRelays"] = nextRelays
		e["nextRelaysPrice"] = nextRelaysPrice

		if entry.NumNearRelays > 0 && entry.UseDebug {
			// IMPORTANT: Only write this data if debug is on because it is very large

			e["numNearRelays"] = int(entry.NumNearRelays)

			nearRelayIDs := make([]bigquery.Value, entry.NumNearRelays)
			nearRelayRTTs := make([]bigquery.Value, entry.NumNearRelays)
			nearRelayJitters := make([]bigquery.Value, entry.NumNearRelays)
			nearRelayPacketLosses := make([]bigquery.Value, entry.NumNearRelays)

			for i := 0; i < int(entry.NumNearRelays); i++ {
				nearRelayIDs[i] = int(entry.NearRelayIDs[i])
				nearRelayRTTs[i] = entry.NearRelayRTTs[i]
				nearRelayJitters[i] = entry.NearRelayJitters[i]
				nearRelayPacketLosses[i] = entry.NearRelayPacketLosses[i]
			}

			e["nearRelayIDs"] = nearRelayIDs
			e["nearRelayRTTs"] = nearRelayRTTs
			e["nearRelayJitters"] = nearRelayJitters
			e["nearRelayPacketLosses"] = nearRelayPacketLosses
		}
	}

	// todo: this is deprecated. we don't really need this anymore. should
	// be made nullable in the schema and we just stop writing this.
	e["initial"] = entry.Initial

	// todo: this is deprecated and should be made nullable in the schema
	// at this point we should stop writing this
	e["routeDecision"] = int(entry.RouteDecision)

	return e, "", nil
}
