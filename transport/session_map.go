package transport

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/networknext/backend/routing"
)

const (
	NumSessionMapShards = 100000

	// todo: disable session locks for the moment
	/*
		// todo: ryan, this seems incredibly low... holy mutex contention batman!
		NumSessionSliceMutexes = 8
	*/
)

type SessionData struct {
	Timestamp            int64
	BuyerID              uint64
	Location             routing.Location
	Sequence             uint64
	NearRelays           []routing.Relay
	RouteHash            uint64
	Initial              bool
	RouteDecision        routing.Decision
	NextSliceCounter     uint64
	CommittedData        routing.CommittedData
	RouteExpireTimestamp uint64
	TokenVersion         uint8
	CachedResponse       []byte
	SliceMutexes         []sync.Mutex
}

type SessionMapShard struct {
	mutex    sync.RWMutex
	sessions map[uint64]*SessionData
}

type SessionMap struct {
	numSessions                    uint64
	numNextSessions                uint64
	numDirectSessions              uint64
	numNextSessionsPerBuyer        map[uint64]uint64
	numNextSessionsPerBuyerMutex   sync.RWMutex
	numDirectSessionsPerBuyer      map[uint64]uint64
	numDirectSessionsPerBuyerMutex sync.RWMutex
	timeoutShard                   int
	shard                          [NumSessionMapShards]*SessionMapShard
}

func NewSessionMap() *SessionMap {
	sessionMap := &SessionMap{
		numNextSessionsPerBuyer:   make(map[uint64]uint64),
		numDirectSessionsPerBuyer: make(map[uint64]uint64),
	}
	for i := 0; i < NumSessionMapShards; i++ {
		sessionMap.shard[i] = &SessionMapShard{}
		sessionMap.shard[i].sessions = make(map[uint64]*SessionData)
	}
	return sessionMap
}

func (sessionMap *SessionMap) GetSessionCount() uint64 {
	return atomic.LoadUint64(&sessionMap.numSessions)
}

func (sessionMap *SessionMap) GetDirectSessionCount() uint64 {
	return atomic.LoadUint64(&sessionMap.numDirectSessions)
}

func (sessionMap *SessionMap) GetNextSessionCount() uint64 {
	return atomic.LoadUint64(&sessionMap.numNextSessions)
}

func (sessionMap *SessionMap) GetDirectSessionCountPerBuyer() map[uint64]uint64 {
	sessionMap.numDirectSessionsPerBuyerMutex.RLock()
	defer sessionMap.numDirectSessionsPerBuyerMutex.RUnlock()

	// make a copy of the map for thread safety
	copy := make(map[uint64]uint64)
	for k, v := range sessionMap.numDirectSessionsPerBuyer {
		copy[k] = v
	}

	return copy
}

func (sessionMap *SessionMap) GetNextSessionCountPerBuyer() map[uint64]uint64 {
	sessionMap.numNextSessionsPerBuyerMutex.RLock()
	defer sessionMap.numNextSessionsPerBuyerMutex.RUnlock()

	// make a copy of the map for thread safety
	copy := make(map[uint64]uint64)
	for k, v := range sessionMap.numNextSessionsPerBuyer {
		copy[k] = v
	}

	return copy
}

func NewSessionData() *SessionData {
	return &SessionData{
		// todo
		// sliceMutexes: make([]sync.Mutex, NumSessionSliceMutexes),
	}
}

func (sessionMap *SessionMap) Lock(sessionId uint64) {
	index := sessionId % NumSessionMapShards
	sessionMap.shard[index].mutex.Lock()
}

func (sessionMap *SessionMap) Unlock(sessionId uint64) {
	index := sessionId % NumSessionMapShards
	sessionMap.shard[index].mutex.Unlock()
}

func (sessionMap *SessionMap) RLock(sessionId uint64) {
	index := sessionId % NumSessionMapShards
	sessionMap.shard[index].mutex.RLock()
}

func (sessionMap *SessionMap) RUnlock(sessionId uint64) {
	index := sessionId % NumSessionMapShards
	sessionMap.shard[index].mutex.RUnlock()
}

func (sessionMap *SessionMap) UpdateSessionData(sessionId uint64, sessionData *SessionData) {
	index := sessionId % NumSessionMapShards
	_, exists := sessionMap.shard[index].sessions[sessionId]

	next := sessionData.NextSliceCounter > 0

	if !exists {
		atomic.AddUint64(&sessionMap.numSessions, 1)

		if next {
			atomic.AddUint64(&sessionMap.numNextSessions, 1)
			sessionMap.numNextSessionsPerBuyerMutex.Lock()
			sessionMap.numNextSessionsPerBuyer[sessionData.BuyerID]++
			sessionMap.numNextSessionsPerBuyerMutex.Unlock()
		} else {
			atomic.AddUint64(&sessionMap.numDirectSessions, 1)
			sessionMap.numDirectSessionsPerBuyerMutex.Lock()
			sessionMap.numDirectSessionsPerBuyer[sessionData.BuyerID]++
			sessionMap.numDirectSessionsPerBuyerMutex.Unlock()
		}
	} else {
		prevNext := sessionMap.shard[index].sessions[sessionId].NextSliceCounter > 0

		// detect next -> direct
		if prevNext && !next {
			atomic.AddUint64(&sessionMap.numNextSessions, ^uint64(0))
			sessionMap.numNextSessionsPerBuyerMutex.Lock()
			sessionMap.numNextSessionsPerBuyer[sessionData.BuyerID]--
			sessionMap.numNextSessionsPerBuyerMutex.Unlock()

			atomic.AddUint64(&sessionMap.numDirectSessions, 1)
			sessionMap.numDirectSessionsPerBuyerMutex.Lock()
			sessionMap.numDirectSessionsPerBuyer[sessionData.BuyerID]++
			sessionMap.numDirectSessionsPerBuyerMutex.Unlock()
		}

		// detect direct -> next
		if !prevNext && next {
			atomic.AddUint64(&sessionMap.numDirectSessions, ^uint64(0))
			sessionMap.numDirectSessionsPerBuyerMutex.Lock()
			sessionMap.numDirectSessionsPerBuyer[sessionData.BuyerID]--
			sessionMap.numDirectSessionsPerBuyerMutex.Unlock()

			atomic.AddUint64(&sessionMap.numNextSessions, 1)
			sessionMap.numNextSessionsPerBuyerMutex.Lock()
			sessionMap.numNextSessionsPerBuyer[sessionData.BuyerID]++
			sessionMap.numNextSessionsPerBuyerMutex.Unlock()
		}
	}

	sessionMap.shard[index].sessions[sessionId] = sessionData
}

func (sessionMap *SessionMap) GetSessionData(sessionId uint64) *SessionData {
	index := sessionId % NumSessionMapShards
	sessionData, _ := sessionMap.shard[index].sessions[sessionId]
	return sessionData
}

func (sessionMap *SessionMap) TimeoutLoop(ctx context.Context, timeoutSeconds int64, c <-chan time.Time) {
	maxShards := 100
	maxIterations := 10
	deleteList := make([]uint64, maxIterations)
	for {
		select {
		case <-c:
			timeoutTimestamp := time.Now().Unix() - timeoutSeconds
			for i := 0; i < maxShards; i++ {
				index := (sessionMap.timeoutShard + i) % NumSessionMapShards
				deleteList = deleteList[:0]
				sessionMap.shard[index].mutex.RLock()
				numIterations := 0
				for k, v := range sessionMap.shard[index].sessions {
					if numIterations > maxIterations || numIterations > len(sessionMap.shard[index].sessions) {
						break
					}
					if v.Timestamp < timeoutTimestamp {
						deleteList = append(deleteList, k)

						atomic.AddUint64(&sessionMap.numSessions, ^uint64(0))
						if next := v.NextSliceCounter > 0; next {
							atomic.AddUint64(&sessionMap.numNextSessions, ^uint64(0))
							sessionMap.numNextSessionsPerBuyerMutex.Lock()
							sessionMap.numNextSessionsPerBuyer[v.BuyerID]--
							sessionMap.numNextSessionsPerBuyerMutex.Unlock()
						} else {
							atomic.AddUint64(&sessionMap.numDirectSessions, ^uint64(0))
							sessionMap.numDirectSessionsPerBuyerMutex.Lock()
							sessionMap.numDirectSessionsPerBuyer[v.BuyerID]--
							sessionMap.numDirectSessionsPerBuyerMutex.Unlock()
						}
					}
					numIterations++
				}
				sessionMap.shard[index].mutex.RUnlock()
				sessionMap.shard[index].mutex.Lock()
				for i := range deleteList {
					// fmt.Printf("timeout session %x\n", deleteList[i])
					delete(sessionMap.shard[index].sessions, deleteList[i])
				}
				sessionMap.shard[index].mutex.Unlock()
			}
			sessionMap.timeoutShard = (sessionMap.timeoutShard + maxShards) % NumSessionMapShards
		case <-ctx.Done():
			return
		}
	}
}
