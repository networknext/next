package transport

import (
	"sync"

	"github.com/networknext/backend/core"
	"github.com/networknext/backend/routing"
)

type StubbedBackend struct {
	mutex           sync.RWMutex
	dirty           bool
	mode            int
	relayDatabase   map[string]routing.Relay
	serverDatabase  map[string]ServerCacheEntry
	sessionDatabase map[uint64]SessionEntry
	statsDatabase   *core.StatsDatabase
	costMatrix      *core.CostMatrix
	costMatrixData  []byte
	routeMatrix     *core.RouteMatrix
	routeMatrixData []byte
	nearData        []byte
}

func NewStubbedBackend() *StubbedBackend {
	backend := new(StubbedBackend)
	backend.relayDatabase = make(map[string]routing.Relay)
	backend.serverDatabase = make(map[string]ServerCacheEntry)
	backend.sessionDatabase = make(map[uint64]SessionEntry)
	backend.statsDatabase = new(core.StatsDatabase)
	backend.costMatrix = new(core.CostMatrix)
	backend.routeMatrix = new(core.RouteMatrix)
	return backend
}
