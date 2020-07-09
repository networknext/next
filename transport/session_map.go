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

	NumSessionSliceMutexes = 8
)

type SessionData struct {
	timestamp            int64
	location             routing.Location
	sequence             uint64
	nearRelays           []routing.Relay
	routeHash            uint64
	routeDecision        routing.Decision
	onNNSliceCounter     uint64
	committedData        routing.CommittedData
	routeExpireTimestamp int64
	tokenVersion         uint8
	cachedResponse       []byte
	sliceMutexes         []sync.Mutex
}

type SessionMapShard struct {
	mutex       sync.Mutex
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
		sliceMutexes: make([]sync.Mutex, NumSessionSliceMutexes),
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
	for {
		select {
		case <-c:
			timeoutTimestamp := time.Now().Add(-timeout).Unix()

			for index := 0; index < NumSessionMapShards; index++ {
				sessionTimeoutStart := time.Now()
				sessionMap.shard[index].mutex.Lock()
				numSessionIterations := 0
				for k, v := range sessionMap.shard[index].sessions {
					if numSessionIterations > 3 {
						break
					}
					if v.timestamp < timeoutTimestamp {
						// fmt.Printf("timed out session: %x\n", k)
						delete(sessionMap.shard[index].sessions, k)
						atomic.AddUint64(&sessionMap.shard[index].numSessions, ^uint64(0))
					}
					numSessionIterations++
				}
				sessionMap.shard[index].mutex.Unlock()
				if time.Since(sessionTimeoutStart).Seconds() > 0.1 {
					// fmt.Printf("long session timeout check [%d]\n", index)
				}
			}
		case <-ctx.Done():
			return
		}
	}
}
