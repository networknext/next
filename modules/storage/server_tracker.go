package storage

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type ServerInfo struct {
	Timestamp      uint64
	DatacenterID   string
	DatacenterName string
}

type ServerTracker struct {
	Tracker      map[string]map[string]ServerInfo
	TrackerMutex sync.RWMutex
}

func NewServerTracker() *ServerTracker {
	tracker := &ServerTracker{
		Tracker:      make(map[string]map[string]ServerInfo),
		TrackerMutex: sync.RWMutex{},
	}

	return tracker
}

func (t *ServerTracker) AddServer(buyerID uint64, datacenterID uint64, serverAddress net.UDPAddr, datacenterName string) {
	var exists bool

	t.TrackerMutex.RLock()

	_, exists = t.Tracker[fmt.Sprintf("%016x", buyerID)]

	t.TrackerMutex.RUnlock()

	if !exists {
		// Add the new buyer to the top-level list
		var addressList = make(map[string]ServerInfo)
		addressList[serverAddress.String()] = ServerInfo{
			Timestamp:      uint64(time.Now().Unix()),
			DatacenterID:   fmt.Sprintf("%016x", datacenterID),
			DatacenterName: datacenterName,
		}

		t.TrackerMutex.Lock()
		t.Tracker[fmt.Sprintf("%016x", buyerID)] = addressList
		t.TrackerMutex.Unlock()
		return
	}

	// Buyer already exists, add server to existing list

	t.TrackerMutex.RLock()

	prevInfo := t.Tracker[fmt.Sprintf("%016x", buyerID)][serverAddress.String()]

	t.TrackerMutex.RUnlock()

	t.TrackerMutex.Lock()

	t.Tracker[fmt.Sprintf("%016x", buyerID)][serverAddress.String()] = ServerInfo{
		Timestamp:      uint64(time.Now().Unix()),
		DatacenterID:   fmt.Sprintf("%016x", datacenterID),
		DatacenterName: prevInfo.DatacenterName,
	}

	t.TrackerMutex.Unlock()
}
