package transport

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
	// "fmt"

	"github.com/networknext/backend/routing"
)

const NumVetoMapShards = 100000

type VetoData struct {
	timestamp int64
	reason    routing.DecisionReason
}

type VetoMapShard struct {
	mutex     sync.RWMutex
	vetoes    map[uint64]VetoData
}

type VetoMap struct {
	numVetoes 		uint64
	timeoutShard 	int
	shard 			[NumVetoMapShards]VetoMapShard
}

func NewVetoMap() *VetoMap {
	vetoMap := &VetoMap{}
	for i := 0; i < NumVetoMapShards; i++ {
		vetoMap.shard[i] = VetoMapShard{}
		vetoMap.shard[i].vetoes = make(map[uint64]VetoData)
	}
	return vetoMap
}

func (vetoMap *VetoMap) GetVetoCount() uint64 {
	return atomic.LoadUint64(&vetoMap.numVetoes)
}

func (vetoMap *VetoMap) RLock(vetoId uint64) {
	index := vetoId % NumVetoMapShards
	vetoMap.shard[index].mutex.RLock()
}

func (vetoMap *VetoMap) RUnlock(vetoId uint64) {
	index := vetoId % NumVetoMapShards
	vetoMap.shard[index].mutex.RUnlock()
}

func (vetoMap *VetoMap) Lock(vetoId uint64) {
	index := vetoId % NumVetoMapShards
	vetoMap.shard[index].mutex.Lock()
}

func (vetoMap *VetoMap) Unlock(vetoId uint64) {
	index := vetoId % NumVetoMapShards
	vetoMap.shard[index].mutex.Unlock()
}

func (vetoMap *VetoMap) SetVeto(vetoId uint64, reason routing.DecisionReason) {
	index := vetoId % NumVetoMapShards
	vetoData := VetoData{
		timestamp: time.Now().Unix(),
		reason:    reason,
	}
	_, exists := vetoMap.shard[index].vetoes[vetoId]
	vetoMap.shard[index].vetoes[vetoId] = vetoData
	if !exists {
		atomic.AddUint64(&vetoMap.numVetoes, uint64(1))
	}
}

func (vetoMap *VetoMap) GetVeto(vetoId uint64) routing.DecisionReason {
	index := vetoId % NumVetoMapShards
	vetoData, exists := vetoMap.shard[index].vetoes[vetoId]
	if exists {
		vetoData.timestamp = time.Now().Unix()
		vetoMap.shard[index].vetoes[vetoId] = vetoData
		return vetoData.reason
	}
	return routing.DecisionNoReason
}

func (vetoMap *VetoMap) TimeoutLoop(ctx context.Context, timeoutSeconds int64, c <-chan time.Time) {
	maxShards := 100
	maxIterations := 10
	deleteList := make([]uint64, maxIterations)
	for {
		select {
		case <-c:
			timeoutTimestamp := time.Now().Unix() - timeoutSeconds
			for i := 0; i < maxShards; i++ {
				index := ( vetoMap.timeoutShard + i ) % NumVetoMapShards
				deleteList = deleteList[:0]
				vetoMap.shard[index].mutex.RLock()
				numIterations := 0
				for k, v := range vetoMap.shard[index].vetoes {
					if numIterations > maxIterations || numIterations > len(vetoMap.shard[index].vetoes) {
						break
					}
					if v.timestamp < timeoutTimestamp {
						deleteList = append(deleteList, k)
					}
					numIterations++
				}
				vetoMap.shard[index].mutex.RUnlock()
				if len(deleteList) > 0 {
					vetoMap.shard[index].mutex.Lock()
					for i := range deleteList {
						// fmt.Printf("timeout veto %x\n", deleteList[i])
						delete(vetoMap.shard[index].vetoes, deleteList[i])
						atomic.AddUint64(&vetoMap.numVetoes, ^uint64(0))
					}
					vetoMap.shard[index].mutex.Unlock()
				}
				vetoMap.timeoutShard = ( vetoMap.timeoutShard + maxShards ) % NumVetoMapShards
			}
		case <-ctx.Done():
			return
		}
	}
}
