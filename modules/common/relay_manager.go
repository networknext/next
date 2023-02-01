package common

import (
	"fmt"
	"net"
	"sort"
	"sync"
	"math"

	"github.com/huandu/go-clone"
)

const RelayTimeout = 10

const HistorySize = 300 // 5 minutes @ one relay update per-second

const NumRelayCounters = 100

func TriMatrixLength(size int) int {
	return (size * (size - 1)) / 2
}

func TriMatrixIndex(i, j int) int {
	if i <= j {
		i, j = j, i
	}
	return i*(i+1)/2 - i + j
}

func historyMax(history []float32) float32 {
	var max float32
	for i := 0; i < len(history); i++ {
		if history[i] > max {
			max = history[i]
		}
	}
	return max
}

func historyMean(history []float32) float32 {
	var sum float64
	for i := 0; i < len(history); i++ {
		sum += float64(history[i])
	}
	return float32(sum / float64(HistorySize))
}

type RelayManagerDestEntry struct {
	LastUpdateTime    int64
	RTT               float32
	Jitter            float32
	PacketLoss        float32
	HistoryIndex      int32
	HistoryRTT        [HistorySize]float32
	HistoryJitter     [HistorySize]float32
	HistoryPacketLoss [HistorySize]float32
}

type RelayManagerSourceEntry struct {
	LastUpdateTime int64
	RelayId        uint64
	RelayName      string
	RelayAddress   net.UDPAddr
	Sessions       int
	RelayVersion   string
	ShuttingDown   bool
	DestEntries    map[uint64]*RelayManagerDestEntry
	Counters       [NumRelayCounters]uint64
}

type RelayManager struct {
	mutex         sync.RWMutex
	SourceEntries map[uint64]*RelayManagerSourceEntry
	TotalCounters [NumRelayCounters]uint64
}

func CreateRelayManager() *RelayManager {
	relayManager := &RelayManager{}
	relayManager.SourceEntries = make(map[uint64]*RelayManagerSourceEntry)
	return relayManager
}

func (relayManager *RelayManager) ProcessRelayUpdate(currentTime int64, relayId uint64, relayName string, relayAddress net.UDPAddr, sessions int, relayVersion string, shuttingDown bool, numSamples int, sampleRelayId []uint64, sampleRTT []float32, sampleJitter []float32, samplePacketLoss []float32, counters []uint64) {

	// look up the entry corresponding to the source relay, or create it if it doesn't exist

	relayManager.mutex.Lock()

	sourceEntry, exists := relayManager.SourceEntries[relayId]
	if !exists {
		sourceEntry = &RelayManagerSourceEntry{}
		sourceEntry.DestEntries = make(map[uint64]*RelayManagerDestEntry)
		relayManager.SourceEntries[relayId] = sourceEntry
	}

	// update stats for the source relay, then...
	// iterate across all samples and insert them into the history buffer
	// in the dest entry corresponding to their relay pair (source,dest)

	sourceEntry.LastUpdateTime = currentTime
	sourceEntry.RelayId = relayId
	sourceEntry.RelayName = relayName
	sourceEntry.RelayAddress = relayAddress
	sourceEntry.Sessions = sessions
	sourceEntry.RelayVersion = relayVersion
	sourceEntry.ShuttingDown = shuttingDown

	for i := 0; i < numSamples; i++ {

		destRelayId := sampleRelayId[i]

		destEntry, exists := sourceEntry.DestEntries[destRelayId]
		if !exists {
			destEntry = &RelayManagerDestEntry{}
			sourceEntry.DestEntries[destRelayId] = destEntry
		}

		destEntry.HistoryRTT[destEntry.HistoryIndex] = sampleRTT[i]
		destEntry.HistoryJitter[destEntry.HistoryIndex] = sampleJitter[i]
		destEntry.HistoryPacketLoss[destEntry.HistoryIndex] = samplePacketLoss[i]

		destEntry.RTT = historyMax(destEntry.HistoryRTT[:])
		destEntry.Jitter = historyMean(destEntry.HistoryJitter[:])
		destEntry.PacketLoss = historyMean(destEntry.HistoryPacketLoss[:])

		destEntry.HistoryIndex = (destEntry.HistoryIndex + 1) % HistorySize

		destEntry.LastUpdateTime = currentTime
	}

	// update relay counters

	for i := 0; i < NumRelayCounters; i++ {
		sourceEntry.Counters[i] = counters[i]
	}

	relayManager.mutex.Unlock()
}

func Max(a float32, b float32) float32 {
	if a > b {
		return a
	} else {
		return b
	}
}

func (relayManager *RelayManager) getSample(currentTime int64, sourceRelayId uint64, destRelayId uint64) (float32, float32, float32) {

	sourceRTT := float32(InvalidRouteValue)
	sourceJitter := float32(InvalidRouteValue)
	sourcePacketLoss := float32(InvalidRouteValue)

	destRTT := float32(InvalidRouteValue)
	destJitter := float32(InvalidRouteValue)
	destPacketLoss := float32(InvalidRouteValue)

	// get source ping values
	{
		sourceEntry := relayManager.SourceEntries[sourceRelayId]
		if sourceEntry != nil {
			destEntry := sourceEntry.DestEntries[destRelayId]
			if destEntry != nil {
				sourceRTT = destEntry.RTT
				sourceJitter = destEntry.Jitter
				sourcePacketLoss = destEntry.PacketLoss
			}
		}
	}

	// get dest ping values
	{
		sourceEntry := relayManager.SourceEntries[destRelayId]
		if sourceEntry != nil {
			destEntry := sourceEntry.DestEntries[sourceRelayId]
			if destEntry != nil {
				destRTT = destEntry.RTT
				destJitter = destEntry.Jitter
				destPacketLoss = destEntry.PacketLoss
			}
		}
	}

	// take maximum of source and dest values

	rtt := Max(sourceRTT, destRTT)
	jitter := Max(sourceJitter, destJitter)
	packetLoss := Max(sourcePacketLoss, destPacketLoss)

	return rtt, jitter, packetLoss
}

func (relayManager *RelayManager) GetCosts(currentTime int64, relayIds []uint64, maxRTT float32, maxJitter float32, maxPacketLoss float32) []int32 {

	numRelays := len(relayIds)

	costs := make([]int32, TriMatrixLength(numRelays))

	for i := range costs {
		costs[i] = -1
	}

	activeRelayHash := relayManager.GetActiveRelayHash(currentTime)

	for i := 0; i < numRelays; i++ {
		sourceRelayId := uint64(relayIds[i])
		_, sourceActive := activeRelayHash[sourceRelayId]
		if sourceActive {
			for j := 0; j < i; j++ {
				destRelayId := uint64(relayIds[j])
				_, destActive := activeRelayHash[destRelayId]
				if destActive {
					relayManager.mutex.RLock()
					rtt, jitter, packetLoss := relayManager.getSample(currentTime, sourceRelayId, destRelayId)
					relayManager.mutex.RUnlock()
					if rtt < maxRTT && jitter < maxJitter && packetLoss < maxPacketLoss {
						index := TriMatrixIndex(i, j)
						costs[index] = int32(math.Ceil(float64(rtt)))
					}
				}
			}
		}
	}

	return costs
}

const RELAY_STATUS_OFFLINE = 0
const RELAY_STATUS_ONLINE = 1
const RELAY_STATUS_SHUTTING_DOWN = 2

var RelayStatusStrings = [3]string{"offline", "online", "shutting down"}

type Relay struct {
	Id       uint64
	Name     string
	Address  net.UDPAddr
	Status   int
	Sessions int
	Version  string
}

func (relayManager *RelayManager) GetRelays(currentTime int64, relayIds []uint64, relayNames []string, relayAddresses []net.UDPAddr) []Relay {

	relayManager.mutex.RLock()

	keys := make([]uint64, len(relayManager.SourceEntries))
	index := 0
	for k := range relayManager.SourceEntries {
		keys[index] = k
		index++
	}

	relays := make([]Relay, 0, len(keys))

	for i := range keys {

		sourceEntry, ok := relayManager.SourceEntries[keys[i]]

		if !ok {
			continue
		}

		relay := Relay{}

		relay.Id = sourceEntry.RelayId
		relay.Name = sourceEntry.RelayName
		relay.Address = sourceEntry.RelayAddress
		relay.Sessions = sourceEntry.Sessions

		relay.Status = RELAY_STATUS_ONLINE

		if sourceEntry.ShuttingDown {
			relay.Status = RELAY_STATUS_SHUTTING_DOWN
		}

		expired := currentTime-sourceEntry.LastUpdateTime > RelayTimeout

		if expired {
			relay.Status = RELAY_STATUS_OFFLINE
		}

		if relay.Status == RELAY_STATUS_ONLINE {
			relay.Version = sourceEntry.RelayVersion
		}

		if relay.Status != RELAY_STATUS_ONLINE {
			relay.Sessions = 0
		}

		relays = append(relays, relay)
	}

	// pick up any relays that the relay manager doesn't know about as offline

	for i := 0; i < len(relayIds); i++ {

		_, exists := relayManager.SourceEntries[relayIds[i]]
		if exists {
			continue
		}

		relay := Relay{}

		relay.Id = relayIds[i]
		relay.Name = relayNames[i]
		relay.Address = relayAddresses[i]
		relay.Sessions = 0
		relay.Version = ""
		relay.Status = RELAY_STATUS_OFFLINE

		relays = append(relays, relay)
	}

	relayManager.mutex.RUnlock()

	// sort to make sure the set of relays is stable order over time

	sort.SliceStable(relays, func(i, j int) bool { return relays[i].Name < relays[j].Name })

	return relays
}

func (relayManager *RelayManager) GetActiveRelays(currentTime int64) []Relay {

	relayManager.mutex.RLock()

	keys := make([]uint64, len(relayManager.SourceEntries))
	index := 0
	for k := range relayManager.SourceEntries {
		keys[index] = k
		index++
	}

	activeRelays := make([]Relay, 0, len(keys))

	for i := range keys {

		sourceEntry, ok := relayManager.SourceEntries[keys[i]]

		if !ok {
			continue
		}

		activeRelay := Relay{}
		activeRelay.Status = RELAY_STATUS_ONLINE
		activeRelay.Name = sourceEntry.RelayName
		activeRelay.Address = sourceEntry.RelayAddress
		activeRelay.Id = sourceEntry.RelayId
		activeRelay.Sessions = sourceEntry.Sessions
		activeRelay.Version = sourceEntry.RelayVersion

		expired := currentTime-sourceEntry.LastUpdateTime > RelayTimeout

		shuttingDown := sourceEntry.ShuttingDown

		if expired || shuttingDown {
			continue
		}

		activeRelays = append(activeRelays, activeRelay)
	}

	relayManager.mutex.RUnlock()

	sort.SliceStable(activeRelays, func(i, j int) bool { return activeRelays[i].Name < activeRelays[j].Name })

	return activeRelays
}

func (relayManager *RelayManager) GetActiveRelayHash(currentTime int64) map[uint64]Relay {

	activeRelays := relayManager.GetActiveRelays(currentTime)

	activeRelayHash := make(map[uint64]Relay)
	for i := range activeRelays {
		activeRelayHash[activeRelays[i].Id] = activeRelays[i]
	}

	return activeRelayHash
}

func (relayManager *RelayManager) GetRelaysCSV(currentTime int64, relayIds []uint64, relayNames []string, relayAddresses []net.UDPAddr) []byte {

	relaysCSV := "name,address,id,status,sessions,version\n"

	relays := relayManager.GetRelays(currentTime, relayIds, relayNames, relayAddresses)

	for i := range relays {
		relay := relays[i]
		relaysCSV += fmt.Sprintf("%s,%s,%016x,%s,%d,%s\n",
			relay.Name,
			relay.Address.String(),
			relay.Id,
			RelayStatusStrings[relay.Status],
			relay.Sessions,
			relay.Version)
	}

	return []byte(relaysCSV)
}

func (relayManager *RelayManager) GetRelayCounters(relayId uint64) []uint64 {
	relayManager.mutex.RLock()
	sourceEntry, ok := relayManager.SourceEntries[relayId]
	relayManager.mutex.RUnlock()
	if !ok {
		return []uint64{}
	}
	return sourceEntry.Counters[:]
}

func (relayManager *RelayManager) GetTotalCounters() []uint64 {
	return relayManager.TotalCounters[:]
}

func (relayManager *RelayManager) Copy() *RelayManager {
	relayManager.mutex.Lock()
	copy := clone.Clone(relayManager).(*RelayManager)
	relayManager.mutex.Unlock()
	return copy
}
