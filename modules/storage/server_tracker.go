package storage

import (
	"fmt"
	"net"
	"sync"
	"time"
)

type ServerInfo struct {
	Timestamp      int64
	DatacenterID   string
	DatacenterName string
}

// Maps a buyerID to a map of Server Address to ServerInfo
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

	buyerHexID := fmt.Sprintf("%016x", buyerID)
	datacenterHexID := fmt.Sprintf("%016x", datacenterID)
	serverAddressStr := serverAddress.String()

	t.TrackerMutex.RLock()

	_, exists = t.Tracker[buyerHexID]

	t.TrackerMutex.RUnlock()

	if !exists {
		// Add the new buyer to the top-level list
		var addressList = make(map[string]ServerInfo)
		addressList[serverAddressStr] = ServerInfo{
			Timestamp:      time.Now().Unix(),
			DatacenterID:   datacenterHexID,
			DatacenterName: datacenterName,
		}

		t.TrackerMutex.Lock()
		t.Tracker[buyerHexID] = addressList
		t.TrackerMutex.Unlock()
		return
	}

	// Buyer already exists, add server to existing list

	t.TrackerMutex.RLock()

	prevInfo, serverExists := t.Tracker[buyerHexID][serverAddressStr]

	t.TrackerMutex.RUnlock()

	if !serverExists {
		// Add the new server to this buyer
		t.TrackerMutex.Lock()

		t.Tracker[buyerHexID][serverAddressStr] = ServerInfo{
			Timestamp:      time.Now().Unix(),
			DatacenterID:   datacenterHexID,
			DatacenterName: datacenterName,
		}

		t.TrackerMutex.Unlock()

		return
	}

	// Server already exists, update timestamp
	t.TrackerMutex.Lock()

	t.Tracker[buyerHexID][serverAddressStr] = ServerInfo{
		Timestamp:      time.Now().Unix(),
		DatacenterID:   prevInfo.DatacenterID,
		DatacenterName: prevInfo.DatacenterName,
	}

	t.TrackerMutex.Unlock()
}
