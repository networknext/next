package main

import (
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-kit/kit/util/conn"
	"github.com/networknext/backend/modules/common/helpers"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport"
	"math/rand"
	"net"
	"runtime"
	"time"
)

type Optimizer struct{
	id			uint64
	cfg         *Config
	createdAt   time.Time
	relayCache  *storage.RelayCache
	shutdown    bool

	Logger 		log.Logger
	Metrics     *Metrics
	MatrixStore storage.MatrixStore
	RelayMap 	*routing.RelayMap
	RelayStore  storage.RelayStore
	StatsDB 	*routing.StatsDatabase
	Store  		storage.Storer
}

func NewBaseOptimizer(cfg *Config) *Optimizer {
	o := new(Optimizer)
	o.id = rand.Uint64()
	o.createdAt = time.Now()
	o.shutdown = false

	return o
}

func (o *Optimizer) GetRelayIDs(excludeList []string) []uint64{
	return o.RelayMap.GetAllRelayIDs(excludeList)
}

func (o *Optimizer) GetRouteMatrix() (*routing.CostMatrix, *routing.RouteMatrix){

	relayIDs := o.RelayMap.GetAllRelayIDs([]string{"valve"})
	costMatrix, routeMatrix := o.NewCostAndRouteMatrixBaseRelayData(relayIDs)

	o.Metrics.costMatrixMetrics.Invocations.Add(1)
	costMatrixDurationStart := time.Now()
	costMatrix.Costs = o.StatsDB.GetCosts(relayIDs, o.cfg.maxJitter, o.cfg.maxPacketLoss)
	costMatrixDurationSince := time.Since(costMatrixDurationStart)
	o.Metrics.costMatrixMetrics.DurationGauge.Set(float64(costMatrixDurationSince.Milliseconds()))
	if costMatrixDurationSince.Seconds() > 1.0 {
		o.Metrics.costMatrixMetrics.LongUpdateCount.Add(1)
	}

	if err := costMatrix.WriteResponseData(o.cfg.matrixBufferSize); err != nil {
		level.Error(o.Logger).Log("matrix", "cost", "op", "write_response", "msg", "could not write response data", "err", err)
		return nil, nil
	}

	o.Metrics.costMatrixMetrics.Bytes.Set(float64(len(costMatrix.GetResponseData())))


	numCPUs := runtime.NumCPU()
	numRelays := len(relayIDs)
	numSegments := numRelays
	if numCPUs < numRelays {
		numSegments = numRelays / 5
		if numSegments == 0 {
			numSegments = 1
		}
	}

	o.Metrics.optimizeMetrics.Invocations.Add(1)
	optimizeDurationStart := time.Now()

	routeEntries := core.Optimize(numRelays, numSegments, costMatrix.Costs, 5, costMatrix.RelayDatacenterIDs)
	if len(routeEntries) == 0 {
		level.Warn(o.Logger).Log("matrix", "cost", "op", "optimize", "warn", "no route entries generated from cost matrix")
		return nil, nil
	}

	optimizeDurationSince := time.Since(optimizeDurationStart)
	o.Metrics.optimizeMetrics.DurationGauge.Set(float64(optimizeDurationSince.Milliseconds()))

	if optimizeDurationSince.Seconds() > 1.0 {
		o.Metrics.optimizeMetrics.LongUpdateCount.Add(1)
	}

	if err := routeMatrix.WriteResponseData(o.cfg.matrixBufferSize); err != nil {
		level.Error(o.Logger).Log("matrix", "route", "op", "write_response", "msg", "could not write response data", "err", err)
		return nil, nil
	}

	routeMatrix.WriteAnalysisData()


	o.Metrics.relayBackendMetrics.RouteMatrix.Bytes.Set(float64(len(routeMatrix.GetResponseData())))
	o.Metrics.relayBackendMetrics.RouteMatrix.RelayCount.Set(float64(len(routeMatrix.RelayIDs)))
	o.Metrics.relayBackendMetrics.RouteMatrix.DatacenterCount.Set(float64(len(routeMatrix.RelayDatacenterIDs)))

	// todo: calculate this in optimize and Store in route matrix so we don't have to calc this here
	numRoutes := int32(0)
	for i := range routeMatrix.RouteEntries {
		numRoutes += routeMatrix.RouteEntries[i].NumRoutes
	}
	o.Metrics.relayBackendMetrics.RouteMatrix.RouteCount.Set(float64(numRoutes))

	memoryUsed := func() float64 {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		return float64(m.Alloc) / (1000.0 * 1000.0)
	}

	o.Metrics.relayBackendMetrics.Goroutines.Set(float64(runtime.NumGoroutine()))
	o.Metrics.relayBackendMetrics.MemoryAllocated.Set(memoryUsed())

	return costMatrix, routeMatrix
}

func (o *Optimizer) GetValveRouteMatrix() (*routing.CostMatrix, *routing.RouteMatrix){
	relayIDs := o.RelayMap.GetAllRelayIDs([]string{})

	costMatrix, routeMatrix := o.NewCostAndRouteMatrixBaseRelayData(relayIDs)

	o.Metrics.valveCostMatrixMetrics.Invocations.Add(1)
	costMatrixDurationStart := time.Now()
	costMatrix.Costs = o.StatsDB.GetCosts(relayIDs, o.cfg.maxJitter, o.cfg.maxPacketLoss)
	costMatrixDurationSince := time.Since(costMatrixDurationStart)
	o.Metrics.valveCostMatrixMetrics.DurationGauge.Set(float64(costMatrixDurationSince.Milliseconds()))
	if costMatrixDurationSince.Seconds() > 1.0 {
		o.Metrics.valveCostMatrixMetrics.LongUpdateCount.Add(1)
	}

	o.Metrics.valveCostMatrixMetrics.Bytes.Set(float64(len(costMatrix.GetResponseData()) * 4))

	numCPUs := runtime.NumCPU()
	numRelays := len(relayIDs)
	numSegments := numRelays
	if numCPUs < numRelays {
		numSegments = numRelays / 5
		if numSegments == 0 {
			numSegments = 1
		}
	}

	o.Metrics.valveOptimizeMetrics.Invocations.Add(1)
	optimizeDurationStart := time.Now()

	routeEntries := core.Optimize(numRelays, numSegments, costMatrix.Costs, 5, costMatrix.RelayDatacenterIDs)
	if len(routeEntries) == 0 {
		level.Warn(o.Logger).Log("matrix", "cost", "op", "optimize", "warn", "no route entries generated from cost matrix")
		return nil, nil
	}

	optimizeDurationSince := time.Since(optimizeDurationStart)
	o.Metrics.valveOptimizeMetrics.DurationGauge.Set(float64(optimizeDurationSince.Milliseconds()))

	if optimizeDurationSince.Seconds() > 1.0 {
		o.Metrics.valveOptimizeMetrics.LongUpdateCount.Add(1)
	}

	routeMatrix.RouteEntries = routeEntries


	if err := routeMatrix.WriteResponseData(o.cfg.matrixBufferSize); err != nil {
		level.Error(o.Logger).Log("matrix", "route", "op", "write_response", "msg", "could not write response data", "err", err)
		return nil, nil
	}

	routeMatrix.WriteAnalysisData()

	o.Metrics.valveRouteMatrixMetrics.Bytes.Set(float64(len(routeMatrix.GetResponseData())))
	o.Metrics.valveRouteMatrixMetrics.RelayCount.Set(float64(len(routeMatrix.RelayIDs)))
	o.Metrics.valveRouteMatrixMetrics.DatacenterCount.Set(float64(len(routeMatrix.RelayDatacenterIDs)))

		// todo: calculate this in optimize and Store in route matrix so we don't have to calc this here
		numRoutes := int32(0)
		for i := range routeMatrix.RouteEntries {
			numRoutes += routeMatrix.RouteEntries[i].NumRoutes
		}
	o.Metrics.valveRouteMatrixMetrics.RouteCount.Set(float64(numRoutes))

	return costMatrix,routeMatrix
}

func (o *Optimizer) UpdateMatrix(routeMatrix routing.RouteMatrix, matrixType string) error{
	matrix := storage.NewMatrix(o.id, o.createdAt, time.Now(),matrixType, routeMatrix.GetResponseData())

	err := o.MatrixStore.UpdateOptimizerMatrix(matrix)
	if err != nil {
		level.Error(o.Logger).Log("msg", "failed to route matrix in MatrixStore", "err", err)
		return err
	}
	return nil
}

func (o *Optimizer) initializeRelay(data *storage.RelayStoreData) {

	relay, err := o.Store.Relay(data.ID)
	if err != nil {
		level.Error(o.Logger).Log("msg", "failed to get relay from storage", "err", err)
		o.Metrics.RelayInitMetrics.ErrorMetrics.RelayNotFound.Add(1)
		return
	}
	o.RelayMap.Lock()
	defer o.RelayMap.Unlock()
	relayData := o.RelayMap.GetRelayData(data.Address.String())
	if relayData != nil {
		level.Warn(o.Logger).Log("msg", "relay already initialized")
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
	return
}

func (o *Optimizer) removeRelay(relayAddress string) {
	o.RelayMap.Lock()
	o.RelayMap.RemoveRelayData(relayAddress)
	o.RelayMap.Unlock()
}


func (o *Optimizer) UpdateRelay(requestBody []byte) {
	var relayUpdateRequest transport.RelayUpdateRequest
	err := relayUpdateRequest.UnmarshalBinary(requestBody)
	if err != nil {
		level.Error(o.Logger).Log("msg", "error unmarshaling relay update request", "err", err)
		o.Metrics.RelayUpdateMetrics.ErrorMetrics.UnmarshalFailure.Add(1)
		return
	}

	relayData := o.RelayMap.GetRelayData(relayUpdateRequest.Address.String())
	if relayData == nil {
		level.Warn(o.Logger).Log("msg", "relay not initialized")
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

func (o *Optimizer) NewCostAndRouteMatrixBaseRelayData(relayIDs []uint64) (*routing.CostMatrix, *routing.RouteMatrix) {
	numRelays := len(relayIDs)
	relayAddresses := make([]net.UDPAddr, numRelays)
	relayNames := make([]string, numRelays)
	relayLatitudes := make([]float32, numRelays)
	relayLongitudes := make([]float32, numRelays)
	relayDatacenterIDs := make([]uint64, numRelays)

	for i, relayID := range relayIDs {
		relay, err := o.Store.Relay(relayID)
		if err != nil {
			continue
		}

		relayAddresses[i] = relay.Addr
		relayNames[i] = relay.Name
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

func (o *Optimizer) MetricsOutput(){
	fmt.Printf("-----------------------------\n")
	fmt.Printf("%.2f mb allocated\n", o.Metrics.relayBackendMetrics.MemoryAllocated.Value())
	fmt.Printf("%d goroutines\n", int(o.Metrics.relayBackendMetrics.Goroutines.Value()))
	fmt.Printf("%d datacenters\n", int(o.Metrics.relayBackendMetrics.RouteMatrix.DatacenterCount.Value()))
	fmt.Printf("%d relays\n", int(o.Metrics.relayBackendMetrics.RouteMatrix.RelayCount.Value()))
	fmt.Printf("%d relays in map\n", o.RelayMap.GetRelayCount())
	fmt.Printf("%d routes\n", int(o.Metrics.relayBackendMetrics.RouteMatrix.RouteCount.Value()))
	fmt.Printf("%d long cost matrix updates\n", int(o.Metrics.costMatrixMetrics.LongUpdateCount.Value()))
	fmt.Printf("%d long route matrix updates\n", int(o.Metrics.optimizeMetrics.LongUpdateCount.Value()))
	fmt.Printf("cost matrix update: %.2f milliseconds\n", o.Metrics.costMatrixMetrics.DurationGauge.Value())
	fmt.Printf("route matrix update: %.2f milliseconds\n", o.Metrics.optimizeMetrics.DurationGauge.Value())
	fmt.Printf("cost matrix bytes: %d\n", int(o.Metrics.costMatrixMetrics.Bytes.Value()))
	fmt.Printf("route matrix bytes: %d\n", int(o.Metrics.relayBackendMetrics.RouteMatrix.Bytes.Value()))
	fmt.Printf("%d ping stats entries submitted\n", int(o.Metrics.relayBackendMetrics.PingStatsMetrics.EntriesSubmitted.Value()))
	fmt.Printf("%d ping stats entries queued\n", int(o.Metrics.relayBackendMetrics.PingStatsMetrics.EntriesQueued.Value()))
	fmt.Printf("%d ping stats entries flushed\n", int(o.Metrics.relayBackendMetrics.PingStatsMetrics.EntriesFlushed.Value()))
	fmt.Printf("%d relay stats entries submitted\n", int(o.Metrics.relayBackendMetrics.RelayStatsMetrics.EntriesSubmitted.Value()))
	fmt.Printf("%d relay stats entries queued\n", int(o.Metrics.relayBackendMetrics.RelayStatsMetrics.EntriesQueued.Value()))
	fmt.Printf("%d relay stats entries flushed\n", int(o.Metrics.relayBackendMetrics.RelayStatsMetrics.EntriesFlushed.Value()))
	fmt.Printf("-----------------------------\n")
}

func (o *Optimizer) RelayCacheRunner() error{

	o.relayCache = storage.NewRelayCache()

	errCount := 0
	syncTimer := helpers.NewSyncTimer(o.cfg.relayCacheUpdate)
	for !o.shutdown{
		syncTimer.Run()

		if errCount > 10 {
			return fmt.Errorf("relay cached errored %v in a row", conn.ErrConnectionUnavailable)
		}

		relayArr, err := o.RelayStore.GetAll()
		if err != nil {
			level.Error(o.Logger).Log("msg", "unable to get relays from Relay Store", "err", err.Error())
			errCount++
			continue
		}

		addArr, removeArr, err := o.relayCache.SetAllWithAddRemove(relayArr)
		if err != nil {
			level.Error(o.Logger).Log("msg", "unable to get relays from Relay Store", "err", err.Error())
			errCount++
			continue
		}

		for _, relay := range addArr{
			 o.initializeRelay(relay)
		}

		for _, id := range removeArr{
			 o.removeRelay(id)
		}

		errCount = 0
	}

	return nil
}