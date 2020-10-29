package main

import (
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	"net"
	"runtime"
	"time"
)

type Optimizer struct{
	cfg         *Config
	relayMap 	*routing.RelayMap
	store  		storage.Storer
	statsDB 	*routing.StatsDatabase
	metrics     *Metrics
	logger 		log.Logger
}

func New(cfg *Config, metrics *Metrics, relayMap *routing.RelayMap, storer storage.Storer, statsDB *routing.StatsDatabase, logger log.Logger) *Optimizer{
	o := new(Optimizer)
	o.cfg = cfg
	o.relayMap = relayMap
	o.store = storer
	o.statsDB = statsDB
	o.metrics = metrics
	o.logger = logger

	return o
}

func (o *Optimizer) GetRelayIDs(excludeList []string) []uint64{
	return o.relayMap.GetAllRelayIDs(excludeList)
}

func (o *Optimizer) GetRouteMatrix() (*routing.CostMatrix, *routing.RouteMatrix){
	relayIDs := o.relayMap.GetAllRelayIDs([]string{"valve"})
	costMatrix, routeMatrix := o.NewCostAndRouteMatrixBaseRelayData(relayIDs)

	o.metrics.costMatrixMetrics.Invocations.Add(1)
	costMatrixDurationStart := time.Now()
	costMatrix.Costs = o.statsDB.GetCosts(relayIDs, o.cfg.maxJitter, o.cfg.maxPacketLoss)
	costMatrixDurationSince := time.Since(costMatrixDurationStart)
	o.metrics.costMatrixMetrics.DurationGauge.Set(float64(costMatrixDurationSince.Milliseconds()))
	if costMatrixDurationSince.Seconds() > 1.0 {
		o.metrics.costMatrixMetrics.LongUpdateCount.Add(1)
	}

	if err := costMatrix.WriteResponseData(o.cfg.matrixBufferSize); err != nil {
		level.Error(o.logger).Log("matrix", "cost", "op", "write_response", "msg", "could not write response data", "err", err)
		return nil, nil
	}

	o.metrics.costMatrixMetrics.Bytes.Set(float64(len(costMatrix.GetResponseData())))


	numCPUs := runtime.NumCPU()
	numRelays := len(relayIDs)
	numSegments := numRelays
	if numCPUs < numRelays {
		numSegments = numRelays / 5
		if numSegments == 0 {
			numSegments = 1
		}
	}

	o.metrics.optimizeMetrics.Invocations.Add(1)
	optimizeDurationStart := time.Now()

	routeEntries := core.Optimize(numRelays, numSegments, costMatrix.Costs, 5, costMatrix.RelayDatacenterIDs)
	if len(routeEntries) == 0 {
		level.Warn(o.logger).Log("matrix", "cost", "op", "optimize", "warn", "no route entries generated from cost matrix")
		return nil, nil
	}

	optimizeDurationSince := time.Since(optimizeDurationStart)
	o.metrics.optimizeMetrics.DurationGauge.Set(float64(optimizeDurationSince.Milliseconds()))

	if optimizeDurationSince.Seconds() > 1.0 {
		o.metrics.optimizeMetrics.LongUpdateCount.Add(1)
	}

	if err := routeMatrix.WriteResponseData(o.cfg.matrixBufferSize); err != nil {
		level.Error(o.logger).Log("matrix", "route", "op", "write_response", "msg", "could not write response data", "err", err)
		return nil, nil
	}

	routeMatrix.WriteAnalysisData()


	o.metrics.relayBackendMetrics.RouteMatrix.Bytes.Set(float64(len(routeMatrix.GetResponseData())))
	o.metrics.relayBackendMetrics.RouteMatrix.RelayCount.Set(float64(len(routeMatrix.RelayIDs)))
	o.metrics.relayBackendMetrics.RouteMatrix.DatacenterCount.Set(float64(len(routeMatrix.RelayDatacenterIDs)))

	// todo: calculate this in optimize and store in route matrix so we don't have to calc this here
	numRoutes := int32(0)
	for i := range routeMatrix.RouteEntries {
		numRoutes += routeMatrix.RouteEntries[i].NumRoutes
	}
	o.metrics.relayBackendMetrics.RouteMatrix.RouteCount.Set(float64(numRoutes))

	memoryUsed := func() float64 {
		var m runtime.MemStats
		runtime.ReadMemStats(&m)
		return float64(m.Alloc) / (1000.0 * 1000.0)
	}

	o.metrics.relayBackendMetrics.Goroutines.Set(float64(runtime.NumGoroutine()))
	o.metrics.relayBackendMetrics.MemoryAllocated.Set(memoryUsed())

	return costMatrix, routeMatrix
}

func (o *Optimizer) GetValveRouteMatrix() (*routing.CostMatrix, *routing.RouteMatrix){
	relayIDs := o.relayMap.GetAllRelayIDs([]string{})

	costMatrix, routeMatrix := o.NewCostAndRouteMatrixBaseRelayData(relayIDs)

	o.metrics.valveCostMatrixMetrics.Invocations.Add(1)
	costMatrixDurationStart := time.Now()
	costMatrix.Costs = o.statsDB.GetCosts(relayIDs, o.cfg.maxJitter, o.cfg.maxPacketLoss)
	costMatrixDurationSince := time.Since(costMatrixDurationStart)
	o.metrics.valveCostMatrixMetrics.DurationGauge.Set(float64(costMatrixDurationSince.Milliseconds()))
	if costMatrixDurationSince.Seconds() > 1.0 {
		o.metrics.valveCostMatrixMetrics.LongUpdateCount.Add(1)
	}

	o.metrics.valveCostMatrixMetrics.Bytes.Set(float64(len(costMatrix.GetResponseData()) * 4))

	numCPUs := runtime.NumCPU()
	numRelays := len(relayIDs)
	numSegments := numRelays
	if numCPUs < numRelays {
		numSegments = numRelays / 5
		if numSegments == 0 {
			numSegments = 1
		}
	}

	o.metrics.valveOptimizeMetrics.Invocations.Add(1)
	optimizeDurationStart := time.Now()

	routeEntries := core.Optimize(numRelays, numSegments, costMatrix.Costs, 5, costMatrix.RelayDatacenterIDs)
	if len(routeEntries) == 0 {
		level.Warn(o.logger).Log("matrix", "cost", "op", "optimize", "warn", "no route entries generated from cost matrix")
		return nil, nil
	}

	optimizeDurationSince := time.Since(optimizeDurationStart)
	o.metrics.valveOptimizeMetrics.DurationGauge.Set(float64(optimizeDurationSince.Milliseconds()))

	if optimizeDurationSince.Seconds() > 1.0 {
		o.metrics.valveOptimizeMetrics.LongUpdateCount.Add(1)
	}

	routeMatrix.RouteEntries = routeEntries


	if err := routeMatrix.WriteResponseData(o.cfg.matrixBufferSize); err != nil {
		level.Error(o.logger).Log("matrix", "route", "op", "write_response", "msg", "could not write response data", "err", err)
		return nil, nil
	}

	routeMatrix.WriteAnalysisData()

	o.metrics.valveRouteMatrixMetrics.Bytes.Set(float64(len(routeMatrix.GetResponseData())))
	o.metrics.valveRouteMatrixMetrics.RelayCount.Set(float64(len(routeMatrix.RelayIDs)))
	o.metrics.valveRouteMatrixMetrics.DatacenterCount.Set(float64(len(routeMatrix.RelayDatacenterIDs)))

		// todo: calculate this in optimize and store in route matrix so we don't have to calc this here
		numRoutes := int32(0)
		for i := range routeMatrix.RouteEntries {
			numRoutes += routeMatrix.RouteEntries[i].NumRoutes
		}
	o.metrics.valveRouteMatrixMetrics.RouteCount.Set(float64(numRoutes))

}

func (o *Optimizer) NewCostAndRouteMatrixBaseRelayData(relayIDs []uint64) (*routing.CostMatrix, *routing.RouteMatrix) {
	numRelays := len(relayIDs)
	relayAddresses := make([]net.UDPAddr, numRelays)
	relayNames := make([]string, numRelays)
	relayLatitudes := make([]float32, numRelays)
	relayLongitudes := make([]float32, numRelays)
	relayDatacenterIDs := make([]uint64, numRelays)

	for i, relayID := range relayIDs {
		relay, err := o.store.Relay(relayID)
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
	fmt.Printf("%.2f mb allocated\n", o.metrics.relayBackendMetrics.MemoryAllocated.Value())
	fmt.Printf("%d goroutines\n", int(o.metrics.relayBackendMetrics.Goroutines.Value()))
	fmt.Printf("%d datacenters\n", int(o.metrics.relayBackendMetrics.RouteMatrix.DatacenterCount.Value()))
	fmt.Printf("%d relays\n", int(o.metrics.relayBackendMetrics.RouteMatrix.RelayCount.Value()))
	fmt.Printf("%d relays in map\n", o.relayMap.GetRelayCount())
	fmt.Printf("%d routes\n", int(o.metrics.relayBackendMetrics.RouteMatrix.RouteCount.Value()))
	fmt.Printf("%d long cost matrix updates\n", int(o.metrics.costMatrixMetrics.LongUpdateCount.Value()))
	fmt.Printf("%d long route matrix updates\n", int(o.metrics.optimizeMetrics.LongUpdateCount.Value()))
	fmt.Printf("cost matrix update: %.2f milliseconds\n", o.metrics.costMatrixMetrics.DurationGauge.Value())
	fmt.Printf("route matrix update: %.2f milliseconds\n", o.metrics.optimizeMetrics.DurationGauge.Value())
	fmt.Printf("cost matrix bytes: %d\n", int(o.metrics.costMatrixMetrics.Bytes.Value()))
	fmt.Printf("route matrix bytes: %d\n", int(o.metrics.relayBackendMetrics.RouteMatrix.Bytes.Value()))
	fmt.Printf("%d ping stats entries submitted\n", int(o.metrics.relayBackendMetrics.PingStatsMetrics.EntriesSubmitted.Value()))
	fmt.Printf("%d ping stats entries queued\n", int(o.metrics.relayBackendMetrics.PingStatsMetrics.EntriesQueued.Value()))
	fmt.Printf("%d ping stats entries flushed\n", int(o.metrics.relayBackendMetrics.PingStatsMetrics.EntriesFlushed.Value()))
	fmt.Printf("%d relay stats entries submitted\n", int(o.metrics.relayBackendMetrics.RelayStatsMetrics.EntriesSubmitted.Value()))
	fmt.Printf("%d relay stats entries queued\n", int(o.metrics.relayBackendMetrics.RelayStatsMetrics.EntriesQueued.Value()))
	fmt.Printf("%d relay stats entries flushed\n", int(o.metrics.relayBackendMetrics.RelayStatsMetrics.EntriesFlushed.Value()))
	fmt.Printf("-----------------------------\n")
}