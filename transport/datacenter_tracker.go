package transport

import (
	"context"
	"fmt"
	"sort"
	"sync"
	"time"
)

type DatacenterTracker struct {
	unknownDatacentersMap   map[uint64]time.Time
	unknownDatacentersMutex sync.Mutex

	emptyDatacentersMap   map[string]time.Time
	emptyDatacentersMutex sync.Mutex
}

func NewDatacenterTracker() *DatacenterTracker {
	return &DatacenterTracker{
		unknownDatacentersMap: make(map[uint64]time.Time),
		emptyDatacentersMap:   make(map[string]time.Time),
	}
}

func (tracker *DatacenterTracker) AddUnknownDatacenter(datacenterID uint64) {
	tracker.unknownDatacentersMutex.Lock()
	defer tracker.unknownDatacentersMutex.Unlock()

	tracker.unknownDatacentersMap[datacenterID] = time.Now()
}

func (tracker *DatacenterTracker) AddEmptyDatacenter(datacenterName string) {
	tracker.emptyDatacentersMutex.Lock()
	defer tracker.emptyDatacentersMutex.Unlock()

	tracker.emptyDatacentersMap[datacenterName] = time.Now()
}

func (tracker *DatacenterTracker) UnknownDatacenterLength() int {
	tracker.unknownDatacentersMutex.Lock()
	defer tracker.unknownDatacentersMutex.Unlock()

	return len(tracker.unknownDatacentersMap)
}

func (tracker *DatacenterTracker) EmptyDatacenterLength() int {
	tracker.emptyDatacentersMutex.Lock()
	defer tracker.emptyDatacentersMutex.Unlock()

	return len(tracker.emptyDatacentersMap)
}

func (tracker *DatacenterTracker) GetUnknownDatacenters() []string {
	tracker.unknownDatacentersMutex.Lock()

	unknownDatacenters := make([]string, len(tracker.unknownDatacentersMap))

	var i int
	for key := range tracker.unknownDatacentersMap {
		unknownDatacenters[i] = fmt.Sprintf("%016x", key)
		i++
	}

	tracker.unknownDatacentersMutex.Unlock()

	sort.Slice(unknownDatacenters, func(i, j int) bool {
		return unknownDatacenters[i] < unknownDatacenters[j]
	})

	return unknownDatacenters
}

func (tracker *DatacenterTracker) GetEmptyDatacenters() []string {
	tracker.emptyDatacentersMutex.Lock()

	emptyDatacenters := make([]string, len(tracker.emptyDatacentersMap))

	var i int
	for key := range tracker.emptyDatacentersMap {
		emptyDatacenters[i] = key
		i++
	}

	tracker.emptyDatacentersMutex.Unlock()

	sort.Slice(emptyDatacenters, func(i, j int) bool {
		return emptyDatacenters[i] < emptyDatacenters[j]
	})

	return emptyDatacenters
}

func (tracker *DatacenterTracker) TimeoutLoop(ctx context.Context, timeout time.Duration, c <-chan time.Time) {
	for {
		select {
		case <-c:
			{
				tracker.unknownDatacentersMutex.Lock()

				numIterations := 0
				for datacenterID, expireTime := range tracker.unknownDatacentersMap {
					if numIterations > 3 {
						break
					}

					if time.Since(expireTime) >= timeout {
						delete(tracker.unknownDatacentersMap, datacenterID)
					}

					numIterations++
				}

				tracker.unknownDatacentersMutex.Unlock()
			}

			{
				tracker.emptyDatacentersMutex.Lock()

				numIterations := 0
				for datacenterID, expireTime := range tracker.emptyDatacentersMap {
					if numIterations > 3 {
						break
					}

					if time.Since(expireTime) >= timeout {
						delete(tracker.emptyDatacentersMap, datacenterID)
					}

					numIterations++
				}

				tracker.emptyDatacentersMutex.Unlock()
			}
		case <-ctx.Done():
			return
		}
	}
}
