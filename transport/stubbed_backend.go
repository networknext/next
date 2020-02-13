package transport

import (
	"sync"

	"github.com/networknext/backend/routing"
)

type StubbedBackend struct {
	mutex           sync.RWMutex
	dirty           bool
	mode            int
	relayDatabase   map[string]routing.Relay
	serverDatabase  map[string]ServerCacheEntry
	sessionDatabase map[uint64]SessionCacheEntry
	statsDatabase   *routing.StatsDatabase
	costMatrix      *routing.CostMatrix
	costMatrixData  []byte
	routeMatrix     *routing.RouteMatrix
	routeMatrixData []byte
	nearData        []byte
}

func NewStubbedBackend() *StubbedBackend {
	backend := new(StubbedBackend)
	backend.relayDatabase = make(map[string]routing.Relay)
	backend.serverDatabase = make(map[string]ServerCacheEntry)
	backend.sessionDatabase = make(map[uint64]SessionCacheEntry)
	backend.statsDatabase = new(routing.StatsDatabase)
	backend.costMatrix = new(routing.CostMatrix)
	backend.routeMatrix = new(routing.RouteMatrix)
	return backend
}
