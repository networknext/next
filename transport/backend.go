package transport

import (
	"sync"

	"github.com/networknext/backend/core"
)

// Backend in-memory version of the actual backend, replace usage with actual backend calls when known
type Backend struct {
	Mutex           sync.RWMutex
	Dirty           bool
	Mode            int
	RelayDatabase   map[string]RelayEntry
	ServerDatabase  map[string]ServerEntry
	SessionDatabase map[uint64]SessionEntry
	StatsDatabase   *core.StatsDatabase
	CostMatrix      *core.CostMatrix
	CostMatrixData  []byte
	RouteMatrix     *core.RouteMatrix
	RouteMatrixData []byte
	NearData        []byte
}

// NewBackend creates a new backend initialized properly
func NewBackend() *Backend {
	backend := new(Backend)
	backend.RelayDatabase = make(map[string]RelayEntry)
	backend.ServerDatabase = make(map[string]ServerEntry)
	backend.SessionDatabase = make(map[uint64]SessionEntry)
	return backend
}
