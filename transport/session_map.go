package transport

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
	// "fmt"

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
	mutex       sync.RWMutex
	sessions    map[uint64]*SessionData
	numSessions uint64
}

type SessionMap struct {
	shard [NumSessionMapShards]*SessionMapShard
}

func NewSessionMap() *SessionMap {
	sessionMap := &SessionMap{}
	for i := 0; i < NumSessionMapShards; i++ {
		sessionMap.shard[i] = &SessionMapShard{}
		sessionMap.shard[i].sessions = make(map[uint64]*SessionData)
	}
	return sessionMap
}

func (sessionMap *SessionMap) NumSessions() uint64 {
	var total uint64
	for i := 0; i < NumSessionMapShards; i++ {
		numSessionsInShard := atomic.LoadUint64(&sessionMap.shard[i].numSessions)
		total += numSessionsInShard
	}
	return total
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

func (sessionMap *SessionMap) UpdateSessionData(sessionId uint64, sessionData *SessionData) {
	index := sessionId % NumSessionMapShards
	_, exists := sessionMap.shard[index].sessions[sessionId]
	sessionMap.shard[index].sessions[sessionId] = sessionData
	if !exists {
		atomic.AddUint64(&sessionMap.shard[index].numSessions, 1)
	}
}

func (sessionMap *SessionMap) GetSessionData(sessionId uint64) *SessionData {
	index := sessionId % NumSessionMapShards
	sessionData, _ := sessionMap.shard[index].sessions[sessionId]
	return sessionData
}

func (sessionMap *SessionMap) TimeoutLoop(ctx context.Context, timeoutSeconds int64, c <-chan time.Time) {
	maxIterations := 100
	deleteList := make([]uint64, maxIterations)
	for {
		select {
		case <-c:
			timeoutTimestamp := time.Now().Unix() - timeoutSeconds
			for index := 0; index < NumSessionMapShards; index++ {
				deleteList = deleteList[:0]
				sessionMap.shard[index].mutex.RLock()
				numIterations := 0
				for k, v := range sessionMap.shard[index].sessions {
					if numIterations > maxIterations || numIterations > len(sessionMap.shard[index].sessions) {
						break
					}
					if v.Timestamp < timeoutTimestamp {
						atomic.AddUint64(&sessionMap.shard[index].numSessions, ^uint64(0))
						deleteList = append(deleteList, k)
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

		case <-ctx.Done():
			return
		}
	}
}
