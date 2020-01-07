package transport

import (
	"sync"

	"github.com/networknext/backend/core"
)

// Backend ? need help commenting on this one, "internal structure for handling connections" maybe?
type Backend struct {
	mutex           sync.RWMutex
	dirty           bool
	mode            int
	relayDatabase   map[string]RelayEntry
	serverDatabase  map[string]ServerEntry
	sessionDatabase map[uint64]SessionEntry
	statsDatabase   *core.StatsDatabase
	costMatrix      *core.CostMatrix
	costMatrixData  []byte
	routeMatrix     *core.RouteMatrix
	routeMatrixData []byte
	nearData        []byte
}

// NewBackend creates a new backend initialized properly
func NewBackend() *Backend {
	backend := new(Backend)
	backend.relayDatabase = make(map[string]RelayEntry)
	backend.serverDatabase = make(map[string]ServerEntry)
	backend.sessionDatabase = make(map[uint64]SessionEntry)
	return backend
}
