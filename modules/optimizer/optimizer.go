package optimizer

// todo: not today
/*
import (
	"context"
	"fmt"
	"math/rand"
	"net"
	"runtime"
	"sort"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-kit/kit/util/conn"
	"github.com/networknext/backend/modules/analytics"
	"github.com/networknext/backend/modules/common/helpers"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/storage"
	"github.com/networknext/backend/modules/transport"
	"github.com/networknext/backend/modules/transport/pubsub"

	gcStorage "cloud.google.com/go/storage"
)

type Optimizer struct {
	id         uint64
	cfg        *Config
	createdAt  time.Time
	relayCache *storage.RelayCache
	shutdown   bool

	Logger      log.Logger
	Metrics     *Metrics
	MatrixStore storage.MatrixStore
	RelayMap    *routing.RelayMap
	RelayStore  storage.RelayStore
	StatsDB     *routing.StatsDatabase
	Store       storage.Storer
	CloudBucket *gcStorage.BucketHandle
}

func NewBaseOptimizer(cfg *Config) *Optimizer {
	o := new(Optimizer)
	o.id = rand.Uint64()
	o.createdAt = time.Now()
	o.shutdown = false
	o.cfg = cfg
	o.relayCache = storage.NewRelayCache()

	return o
}

func (o *Optimizer) GetRelayIDs(excludeList []string) []uint64 {
	return o.RelayMap.GetAllRelayIDs(excludeList)
}

func (o *Optimizer) costMatrix(relayIDs []uint64, costMatrix *routing.CostMatrix, metrics *metrics.CostMatrixMetrics) *routing.CostMatrix {
	metrics.Invocations.Add(1)
	costMatrixDurationStart := time.Now()
	costMatrix.Costs = o.StatsDB.GetCosts(relayIDs, o.cfg.MaxJitter, o.cfg.MaxPacketLoss)
	costMatrixDurationSince := time.Since(costMatrixDurationStart)
	metrics.DurationGauge.Set(float64(costMatrixDurationSince.Milliseconds()))
	if costMatrixDurationSince.Seconds() > 1.0 {
		metrics.LongUpdateCount.Add(1)
	}

	if err := costMatrix.WriteResponseData(o.cfg.MatrixBufferSize); err != nil {
		_ = level.Error(o.Logger).Log("matrix", "cost", "op", "write_response", "msg", "could not write response data", "err", err)
		return nil
	}

	metrics.Bytes.Set(float64(len(costMatrix.GetResponseData()) * 4))

	return costMatrix
}

func (o *Optimizer) numSegments(numRelays int, numThreads int) int {
	numSegments := numRelays
	if numThreads < numRelays {
		numSegments = numRelays / 5
		if numSegments == 0 {
			numSegments = 1
		}
	}

	return numSegments
}

func (o *Optimizer) optimize(numRelays, numSegments int, costMatrix *routing.CostMatrix, routeMatrix *routing.RouteMatrix, metrics *metrics.OptimizeMetrics) *routing.RouteMatrix {
	metrics.Invocations.Add(1)
	optimizeDurationStart := time.Now()

	costThreshold := int32(1)

	routeEntries := core.Optimize(numRelays, numSegments, costMatrix.Costs, costThreshold, costMatrix.RelayDatacenterIDs)
	if len(routeEntries) == 0 {
		_ = level.Warn(o.Logger).Log("matrix", "cost", "op", "optimize", "warn", "no route entries generated from cost matrix")
		return nil
	}
	routeMatrix.RouteEntries = routeEntries
	optimizeDurationSince := time.Since(optimizeDurationStart)
	metrics.DurationGauge.Set(float64(optimizeDurationSince.Milliseconds()))

	if optimizeDurationSince.Seconds() > 1.0 {
		metrics.LongUpdateCount.Add(1)
	}

	if err := routeMatrix.WriteResponseData(o.cfg.MatrixBufferSize); err != nil {
		_ = level.Error(o.Logger).Log("matrix", "route", "op", "write_response", "msg", "could not write response data", "err", err)
		return nil
	}

	routeMatrix.WriteAnalysisData()

	return routeMatrix
}

func (o *Optimizer) GetRouteMatrix() (*routing.CostMatrix, *routing.RouteMatrix) {
	relayIDs := o.GetRelayIDs([]string{"valve"})
	costMatrix, routeMatrix := o.NewCostAndRouteMatrixBaseRelayData(relayIDs)

	costMatrix = o.costMatrix(relayIDs, costMatrix, o.Metrics.CostMatrixMetrics)
	if costMatrix == nil {
		return nil, nil
	}

	numRelays := len(relayIDs)
	numSegments := o.numSegments(numRelays, o.cfg.NumThreads)

	routeMatrix = o.optimize(numRelays, numSegments, costMatrix, routeMatrix, o.Metrics.OptimizeMetrics)
	if routeMatrix == nil {
		return nil, nil
	}

	o.Metrics.RelayBackendMetrics.RouteMatrix.Bytes.Set(float64(len(routeMatrix.GetResponseData())))
	o.Metrics.RelayBackendMetrics.RouteMatrix.RelayCount.Set(float64(len(routeMatrix.RelayIDs)))
	o.Metrics.RelayBackendMetrics.RouteMatrix.DatacenterCount.Set(float64(len(routeMatrix.RelayDatacenterIDs)))

	// todo: calculate this in optimize and Store in route matrix so we don't have to calc this here
	numRoutes := int32(0)
	for i := range routeMatrix.RouteEntries {
		numRoutes += routeMatrix.RouteEntries[i].NumRoutes
	}
	o.Metrics.RelayBackendMetrics.RouteMatrix.RouteCount.Set(float64(numRoutes))

	memoryUsed := func() float64 {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		return float64(m.Alloc) / (1000.0 * 1000.0)
	}

	o.Metrics.RelayBackendMetrics.Goroutines.Set(float64(runtime.NumGoroutine()))
	o.Metrics.RelayBackendMetrics.MemoryAllocated.Set(memoryUsed())

	return costMatrix, routeMatrix
}

func (o *Optimizer) GetValveRouteMatrix() (*routing.CostMatrix, *routing.RouteMatrix) {
	relayIDs := o.GetRelayIDs([]string{})

	costMatrix, routeMatrix := o.NewCostAndRouteMatrixBaseRelayData(relayIDs)

	costMatrix = o.costMatrix(relayIDs, costMatrix, o.Metrics.ValveCostMatrixMetrics)
	if costMatrix == nil {
		return nil, nil
	}
	numRelays := len(relayIDs)
	numSegments := o.numSegments(numRelays, o.cfg.NumThreads)

	routeMatrix = o.optimize(numRelays, numSegments, costMatrix, routeMatrix, o.Metrics.ValveOptimizeMetrics)
	if routeMatrix == nil {
		return nil, nil
	}

	o.Metrics.ValveRouteMatrixMetrics.Bytes.Set(float64(len(routeMatrix.GetResponseData())))
	o.Metrics.ValveRouteMatrixMetrics.RelayCount.Set(float64(len(routeMatrix.RelayIDs)))
	o.Metrics.ValveRouteMatrixMetrics.DatacenterCount.Set(float64(len(routeMatrix.RelayDatacenterIDs)))

	// todo: calculate this in optimize and Store in route matrix so we don't have to calc this here
	numRoutes := int32(0)
	for i := range routeMatrix.RouteEntries {
		numRoutes += routeMatrix.RouteEntries[i].NumRoutes
	}
	o.Metrics.ValveRouteMatrixMetrics.RouteCount.Set(float64(numRoutes))

	return costMatrix, routeMatrix
}

func (o *Optimizer) CloudStoreMatrix(matrixType string, timestamp time.Time, matrix []byte) error {
	dir := fmt.Sprintf("matrix/optimizer/%d/%d/%d/%d/%d/%d/%s-%d", o.id, timestamp.Year(), timestamp.Month(), timestamp.Day(), timestamp.Hour(), timestamp.Minute(), matrixType, timestamp.Second())
	obj := o.CloudBucket.Object(dir)
	writer := obj.NewWriter(context.Background())
	defer writer.Close()
	_, err := writer.Write(matrix)
	return err
}

func (o *Optimizer) UpdateMatrix(routeMatrix routing.RouteMatrix, matrixType string) error {
	matrix := storage.NewMatrix(o.id, o.createdAt, time.Now(), matrixType, routeMatrix.GetResponseData())
	err := o.MatrixStore.UpdateOptimizerMatrix(matrix)
	if err != nil {
		_ = level.Error(o.Logger).Log("msg", "failed to route matrix in MatrixStore", "err", err)
		return err
	}
	return nil
}

func (o *Optimizer) initializeRelay(data *storage.RelayStoreData) {

	relay, err := o.Store.Relay(data.ID)
	if err != nil {
		_ = level.Error(o.Logger).Log("msg", "failed to get relay from storage", "err", err)
		o.Metrics.RelayInitMetrics.ErrorMetrics.RelayNotFound.Add(1)
		return
	}
	o.RelayMap.Lock()
	defer o.RelayMap.Unlock()
	relayData := o.RelayMap.GetRelayData(data.Address.String())
	if relayData != nil {
		_ = level.Warn(o.Logger).Log("msg", "relay already initialized")
		o.Metrics.RelayInitMetrics.ErrorMetrics.RelayAlreadyExists.Add(1)
		return
	}

	relayData = routing.NewRelayData()
	{
		relayData.ID = data.ID
		relayData.Name = relay.Name
		relayData.Addr = data.Address
		relayData.PublicKey = relay.PublicKey
		relayData.Seller = relay.Seller
		relayData.Datacenter = relay.Datacenter
		relayData.LastUpdateTime = time.Now()
		relayData.MaxSessions = relay.MaxSessions
		relayData.Version = data.RelayVersion
	}

	o.RelayMap.AddRelayDataEntry(relayData.Addr.String(), relayData)
}

func (o *Optimizer) removeRelay(relayAddress string) {
	_ = level.Debug(o.Logger).Log("msg", "relay being removed", "id", relayAddress)
	o.RelayMap.Lock()
	o.RelayMap.RemoveRelayData(relayAddress)
	o.RelayMap.Unlock()
}

func (o *Optimizer) UpdateRelay(requestBody []byte) {
	var relayUpdateRequest transport.RelayUpdateRequest
	err := relayUpdateRequest.UnmarshalBinary(requestBody)
	if err != nil {
		_ = level.Error(o.Logger).Log("msg", "error unmarshaling relay update request", "err", err)
		o.Metrics.RelayUpdateMetrics.ErrorMetrics.UnmarshalFailure.Add(1)
		return
	}

	relayData := o.RelayMap.GetRelayData(relayUpdateRequest.Address.String())
	if relayData == nil {
		_ = level.Warn(o.Logger).Log("msg", "relay not initialized")
		o.Metrics.RelayUpdateMetrics.ErrorMetrics.RelayNotFound.Add(1)
		return
	}

	id := crypto.HashID(relayUpdateRequest.Address.String())
	statsUpdate := &routing.RelayStatsUpdate{}
	statsUpdate.ID = id
	statsUpdate.PingStats = append(statsUpdate.PingStats, relayUpdateRequest.PingStats...)
	o.StatsDB.ProcessStats(statsUpdate)

	// Update the relay data
	o.RelayMap.Lock()
	o.RelayMap.UpdateRelayDataEntry(relayUpdateRequest.Address.String(), relayUpdateRequest.TrafficStats, float32(relayUpdateRequest.CPUUsage)*100.0, float32(relayUpdateRequest.MemUsage)*100.0)
	o.RelayMap.Unlock()
}

func (o *Optimizer) NewCostAndRouteMatrixBaseRelayData(baseRelayIDs []uint64) (*routing.CostMatrix, *routing.RouteMatrix) {
	namesMap := make(map[string]routing.Relay)
	numRelays := len(baseRelayIDs)
	relayAddresses := make([]net.UDPAddr, numRelays)
	relayNames := make([]string, numRelays)
	relayLatitudes := make([]float32, numRelays)
	relayLongitudes := make([]float32, numRelays)
	relayDatacenterIDs := make([]uint64, numRelays)
	relayIDs := make([]uint64, numRelays)

	for i, relayID := range baseRelayIDs {
		relay, err := o.Store.Relay(relayID)
		if err != nil {
			continue
		}

		relayNames[i] = relay.Name
		namesMap[relay.Name] = relay
	}
	//sort relay names then populate other arrays
	sort.Strings(relayNames)
	for i, relayName := range relayNames {
		relay := namesMap[relayName]
		relayIDs[i] = relay.ID
		relayAddresses[i] = relay.Addr
		relayLatitudes[i] = float32(relay.Datacenter.Location.Latitude)
		relayLongitudes[i] = float32(relay.Datacenter.Location.Longitude)
		relayDatacenterIDs[i] = relay.Datacenter.ID
	}

	costMatrix := &routing.CostMatrix{
		RelayIDs:           relayIDs,
		RelayAddresses:     relayAddresses,
		RelayNames:         relayNames,
		RelayLatitudes:     relayLatitudes,
		RelayLongitudes:    relayLongitudes,
		RelayDatacenterIDs: relayDatacenterIDs,
	}

	routeMatrix := &routing.RouteMatrix{
		RelayIDs:           relayIDs,
		RelayAddresses:     relayAddresses,
		RelayNames:         relayNames,
		RelayLatitudes:     relayLatitudes,
		RelayLongitudes:    relayLongitudes,
		RelayDatacenterIDs: relayDatacenterIDs,
	}

	return costMatrix, routeMatrix
}

func (o *Optimizer) MetricsOutput() {
	fmt.Printf("-----------------------------\n")
	fmt.Printf("%.2f mb allocated\n", o.Metrics.RelayBackendMetrics.MemoryAllocated.Value())
	fmt.Printf("%d goroutines\n", int(o.Metrics.RelayBackendMetrics.Goroutines.Value()))
	fmt.Printf("%d datacenters\n", int(o.Metrics.RelayBackendMetrics.RouteMatrix.DatacenterCount.Value()))
	fmt.Printf("%d relays\n", int(o.Metrics.RelayBackendMetrics.RouteMatrix.RelayCount.Value()))
	fmt.Printf("%d relays in map\n", o.RelayMap.GetRelayCount())
	fmt.Printf("%d routes\n", int(o.Metrics.RelayBackendMetrics.RouteMatrix.RouteCount.Value()))
	fmt.Printf("%d long cost matrix updates\n", int(o.Metrics.CostMatrixMetrics.LongUpdateCount.Value()))
	fmt.Printf("%d long route matrix updates\n", int(o.Metrics.OptimizeMetrics.LongUpdateCount.Value()))
	fmt.Printf("cost matrix update: %.2f milliseconds\n", o.Metrics.CostMatrixMetrics.DurationGauge.Value())
	fmt.Printf("route matrix update: %.2f milliseconds\n", o.Metrics.OptimizeMetrics.DurationGauge.Value())
	fmt.Printf("cost matrix bytes: %d\n", int(o.Metrics.CostMatrixMetrics.Bytes.Value()))
	fmt.Printf("route matrix bytes: %d\n", int(o.Metrics.RelayBackendMetrics.RouteMatrix.Bytes.Value()))
	fmt.Printf("%d ping stats entries submitted\n", int(o.Metrics.RelayBackendMetrics.PingStatsMetrics.EntriesSubmitted.Value()))
	fmt.Printf("%d ping stats entries queued\n", int(o.Metrics.RelayBackendMetrics.PingStatsMetrics.EntriesQueued.Value()))
	fmt.Printf("%d ping stats entries flushed\n", int(o.Metrics.RelayBackendMetrics.PingStatsMetrics.EntriesFlushed.Value()))
	fmt.Printf("%d relay stats entries submitted\n", int(o.Metrics.RelayBackendMetrics.RelayStatsMetrics.EntriesSubmitted.Value()))
	fmt.Printf("%d relay stats entries queued\n", int(o.Metrics.RelayBackendMetrics.RelayStatsMetrics.EntriesQueued.Value()))
	fmt.Printf("%d relay stats entries flushed\n", int(o.Metrics.RelayBackendMetrics.RelayStatsMetrics.EntriesFlushed.Value()))
	fmt.Printf("-----------------------------\n")
}

func (o *Optimizer) RelayCacheRunner() error {

	errCount := 0
	syncTimer := helpers.NewSyncTimer(o.cfg.RelayCacheUpdate)
	for !o.shutdown {
		syncTimer.Run()

		if errCount > 10 {
			return fmt.Errorf("relay cached errored %v in a row", conn.ErrConnectionUnavailable)
		}

		relayArr, err := o.RelayStore.GetAll()
		if err != nil {
			_ = level.Error(o.Logger).Log("msg", "unable to get relays from Relay Store", "err", err.Error())
			errCount++
			continue
		}

		addArr, removeArr, err := o.relayCache.SetAllWithAddRemove(relayArr)
		if err != nil {
			_ = level.Error(o.Logger).Log("msg", "unable to get relays from Relay Store", "err", err.Error())
			errCount++
			continue
		}

		for _, relay := range addArr {
			o.initializeRelay(relay)
		}

		for _, id := range removeArr {
			o.removeRelay(id)
		}

		errCount = 0
	}

	return nil
}

func (o *Optimizer) StartSubscriber() error {
	level.Debug(o.Logger).Log("msg", "subscriber starting")

	sub, err := pubsub.NewGenericSubscriber(o.cfg.subscriberPort, o.cfg.subscriberRecieveBufferSize)
	if err != nil {
		return err
	}
	defer sub.Close()

	err = sub.Subscribe(pubsub.RelayUpdateTopic)
	if err != nil {
		return err
	}

	for !o.shutdown {
		msgChan := sub.ReceiveMessage(context.Background())
		msg := <-msgChan
		_ = level.Debug(o.Logger).Log("msg", "meesage recved")
		if msg.Err != nil {
			_ = level.Error(o.Logger).Log("err", err)
		}
		if msg.Topic != pubsub.RelayUpdateTopic {
			_ = level.Error(o.Logger).Log("err", "received the wrong topic")
		}

		o.UpdateRelay(msg.Message)
	}
	return nil
}

func (o *Optimizer) PingPublishRunner(pingStatsPublisher analytics.PingStatsPublisher, ctx context.Context, publishInterval time.Duration) error {
	syncTimer := helpers.NewSyncTimer(publishInterval)
	for !o.shutdown {
		syncTimer.Run()

		cpy := o.StatsDB.MakeCopy()
		entries := analytics.ExtractPingStats(cpy, o.cfg.MaxJitter, o.cfg.MaxPacketLoss)
		if err := pingStatsPublisher.Publish(ctx, entries); err != nil {
			_ = level.Error(o.Logger).Log("err", err)
			return err
		}
	}

	return nil
}

func (o *Optimizer) RelayPublishRunner(relayStatsPublisher analytics.RelayStatsPublisher, ctx context.Context, publishInterval time.Duration) error {

	syncTimer := helpers.NewSyncTimer(publishInterval)
	for !o.shutdown {
		syncTimer.Run()
		allRelayData := o.RelayMap.GetAllRelayData()
		entries := make([]analytics.RelayStatsEntry, len(allRelayData))

		count := 0
		for i := range allRelayData {
			relay := &allRelayData[i]

			// convert peak to mbps

			var traffic routing.TrafficStats

			relay.TrafficMu.Lock()
			for i := range relay.TrafficStatsBuff {
				stats := &relay.TrafficStatsBuff[i]
				traffic = traffic.Add(stats)
			}
			numSessions := relay.PeakTrafficStats.SessionCount
			envUp := relay.PeakTrafficStats.EnvelopeUpKbps
			envDown := relay.PeakTrafficStats.EnvelopeDownKbps
			elapsed := time.Since(relay.LastStatsPublishTime)

			o.RelayMap.ClearRelayData(relay.Addr.String())
			relay.TrafficMu.Unlock()

			fsrelay, err := o.Store.Relay(relay.ID)
			if err != nil {
				continue
			}

			// use the sum of all the stats since the last publish here and convert to mbps
			bwSentMbps := float32(float64(traffic.AllTx()) * 8.0 / 1000000.0 / elapsed.Seconds())
			bwRecvMbps := float32(float64(traffic.AllRx()) * 8.0 / 1000000.0 / elapsed.Seconds())

			// use the peak envelope values here and convert, it's already per second so no need for time adjustment
			envSentMbps := float32(float64(envUp) / 1000.0)
			envRecvMbps := float32(float64(envDown) / 1000.0)

			var numRouteable uint32 = 0
			for i := range allRelayData {
				otherRelay := &allRelayData[i]

				if relay.ID == otherRelay.ID {
					continue
				}

				rtt, jitter, pl := o.StatsDB.GetSample(relay.ID, otherRelay.ID)
				if rtt != routing.InvalidRouteValue && jitter != routing.InvalidRouteValue && pl != routing.InvalidRouteValue {
					if jitter <= float32(o.cfg.MaxJitter) && pl <= float32(o.cfg.MaxPacketLoss) {
						numRouteable++
					}
				}
			}

			var bwSentPercent float32
			var bwRecvPercent float32
			var envSentPercent float32
			var envRecvPercent float32
			if fsrelay.NICSpeedMbps != 0 {
				bwSentPercent = bwSentMbps / float32(fsrelay.NICSpeedMbps) * 100.0
				bwRecvPercent = bwRecvMbps / float32(fsrelay.NICSpeedMbps) * 100.0
				envSentPercent = envSentMbps / float32(fsrelay.NICSpeedMbps) * 100.0
				envRecvPercent = envRecvMbps / float32(fsrelay.NICSpeedMbps) * 100.0
			}

			entries[count] = analytics.RelayStatsEntry{
				ID:                       relay.ID,
				CPUUsage:                 relay.CPUUsage,
				MemUsage:                 relay.MemUsage,
				BandwidthSentPercent:     bwSentPercent,
				BandwidthReceivedPercent: bwRecvPercent,
				EnvelopeSentPercent:      envSentPercent,
				EnvelopeReceivedPercent:  envRecvPercent,
				BandwidthSentMbps:        bwSentMbps,
				BandwidthReceivedMbps:    bwRecvMbps,
				EnvelopeSentMbps:         envSentMbps,
				EnvelopeReceivedMbps:     envRecvMbps,
				NumSessions:              uint32(numSessions),
				MaxSessions:              relay.MaxSessions,
				NumRoutable:              numRouteable,
				NumUnroutable:            uint32(len(allRelayData)) - 1 - numRouteable,
			}

			count++
		}

		entriesToPublish := entries[:count]
		if len(entriesToPublish) > 0 {
			if err := relayStatsPublisher.Publish(ctx, entriesToPublish); err != nil {
				return err
			}
		}
	}
	return nil
}
*/
