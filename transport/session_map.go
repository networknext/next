package transport

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/networknext/backend/routing"
)

const (
	NumSessionMapShards = 4096

	// todo: disable session locks for the moment
	/*
		// todo: ryan, this seems incredibly low... holy mutex contention batman!
		NumSessionSliceMutexes = 8
	*/
)

type SessionData struct {
	timestamp            int64
	location             routing.Location
	sequence             uint64
	nearRelays           []routing.Relay
	routeHash            uint64
	initial              bool
	routeDecision        routing.Decision
	nextSliceCounter     uint64
	committedData        routing.CommittedData
	routeExpireTimestamp uint64
	tokenVersion         uint8
	cachedResponse       []byte
	sliceMutexes         []sync.Mutex
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

func (sessionMap *SessionMap) UpdateSessionData(sessionId uint64, sessionData *SessionData) {
	index := sessionId % NumSessionMapShards
	sessionMap.shard[index].mutex.Lock()
	_, exists := sessionMap.shard[index].sessions[sessionId]
	sessionMap.shard[index].sessions[sessionId] = sessionData
	sessionMap.shard[index].mutex.Unlock()
	if !exists {
		atomic.AddUint64(&sessionMap.shard[index].numSessions, 1)
	}
}

func (sessionMap *SessionMap) GetSessionData(sessionId uint64) *SessionData {
	index := sessionId % NumServerMapShards
	sessionMap.shard[index].mutex.Lock()
	sessionData, _ := sessionMap.shard[index].sessions[sessionId]
	sessionMap.shard[index].mutex.Unlock()
	return sessionData
}

func (sessionMap *SessionMap) TimeoutLoop(ctx context.Context, timeout time.Duration, c <-chan time.Time) {
	numIterations := 100
	deleteList := make([]uint64, numIterations)
	for {
		select {
		case <-c:
			timeoutTimestamp := time.Now().Add(-timeout).Unix()

			deleteList = deleteList[:0]

			for index := 0; index < NumSessionMapShards; index++ {
				sessionMap.shard[index].mutex.RLock()
				numSessionIterations := 0
				for k, v := range sessionMap.shard[index].sessions {
					if numSessionIterations > numIterations {
						break
					}
					if v.timestamp < timeoutTimestamp {
						atomic.AddUint64(&sessionMap.shard[index].numSessions, ^uint64(0))
						deleteList = append(deleteList, k)
					}
					numSessionIterations++
				}
				sessionMap.shard[index].mutex.RUnlock()
				sessionMap.shard[index].mutex.Lock()
				for i := range deleteList {
					delete(sessionMap.shard[index].sessions, deleteList[i])
				}
				sessionMap.shard[index].mutex.Unlock()
			}

		case <-ctx.Done():
			return
		}
	}
}
