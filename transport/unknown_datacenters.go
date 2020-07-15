package transport

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

type UnknownDatacenters struct {
	unknownDatacentersMap map[uint64]time.Time
	mutex                 sync.Mutex
}

func NewUnknownDatacenters() *UnknownDatacenters {
	return &UnknownDatacenters{
		unknownDatacentersMap: make(map[uint64]time.Time),
	}
}

func (ud *UnknownDatacenters) Add(datacenterID uint64) {
	ud.mutex.Lock()
	defer ud.mutex.Unlock()

	ud.unknownDatacentersMap[datacenterID] = time.Now()
}

func (ud *UnknownDatacenters) Length() int {
	ud.mutex.Lock()
	defer ud.mutex.Unlock()

	return len(ud.unknownDatacentersMap)
}

func (ud *UnknownDatacenters) GetUnknownDatacenters() []string {
	ud.mutex.Lock()

	unknownDatacenters := make([]string, len(ud.unknownDatacentersMap))

	var i int
	for key := range ud.unknownDatacentersMap {
		unknownDatacenters[i] = fmt.Sprintf("%016x", key)
		i++
	}

	ud.mutex.Unlock()

	sort.Slice(unknownDatacenters, func(i, j int) bool {
		return unknownDatacenters[i] < unknownDatacenters[j]
	})

	return unknownDatacenters
}

func (ud *UnknownDatacenters) TimeoutLoop(ctx context.Context, timeout time.Duration, c <-chan time.Time) {
	for {
		select {
		case <-c:
			ud.mutex.Lock()

			numIterations := 0
			for datacenterID, expireTime := range ud.unknownDatacentersMap {
				if numIterations > 3 {
					break
				}

				if time.Since(expireTime) >= timeout {
					delete(ud.unknownDatacentersMap, datacenterID)
				}

				numIterations++
			}

			ud.mutex.Unlock()
		case <-ctx.Done():
			return
		}
	}
}
