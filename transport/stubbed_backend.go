package transport

import (
	"net"
	"sync"

	"github.com/networknext/backend/core"
)

type StubbedRelayEntry struct {
	id         uint64
	name       string
	address    *net.UDPAddr
	lastUpdate int64
	token      []byte
}

type StubbedServerEntry struct {
	address    *net.UDPAddr
	publicKey  []byte
	lastUpdate int64
}

type StubbedSessionEntry struct {
	id              uint64
	version         uint8
	expireTimestamp uint64
	route           []uint64
	next            bool
	slice           uint64
}

type StubbedBackend struct {
	mutex           sync.RWMutex
	dirty           bool
	mode            int
	relayDatabase   map[string]StubbedRelayEntry
	serverDatabase  map[string]StubbedServerEntry
	sessionDatabase map[uint64]StubbedSessionEntry
	statsDatabase   *core.StatsDatabase
	costMatrix      *core.CostMatrix
	costMatrixData  []byte
	routeMatrix     *core.RouteMatrix
	routeMatrixData []byte
	nearData        []byte
}

func NewStubbedBackend() *StubbedBackend {
	backend := new(StubbedBackend)
	backend.relayDatabase = make(map[string]StubbedRelayEntry)
	backend.serverDatabase = make(map[string]StubbedServerEntry)
	backend.sessionDatabase = make(map[uint64]StubbedSessionEntry)
	backend.statsDatabase = new(core.StatsDatabase)
	backend.costMatrix = new(core.CostMatrix)
	backend.routeMatrix = new(core.RouteMatrix)
	return backend
}
