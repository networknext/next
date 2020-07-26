package transport

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
	"fmt"

	"github.com/networknext/backend/routing"
)

const NumVetoMapShards = 10000

type VetoData struct {
	timestamp int64
	reason    routing.DecisionReason
}

type VetoMapShard struct {
	mutex     sync.RWMutex
	vetoes    map[uint64]VetoData
	numVetoes uint64
}

type VetoMap struct {
	shard [NumVetoMapShards]VetoMapShard
}

func NewVetoMap() *VetoMap {
	vetoMap := &VetoMap{}
	for i := 0; i < NumVetoMapShards; i++ {
		vetoMap.shard[i] = VetoMapShard{}
		vetoMap.shard[i].vetoes = make(map[uint64]VetoData)
	}
	return vetoMap
}

func (vetoMap *VetoMap) NumVetoes() uint64 {
	var total uint64
	for i := 0; i < NumVetoMapShards; i++ {
		numVetoesInShard := atomic.LoadUint64(&vetoMap.shard[i].numVetoes)
		total += numVetoesInShard
	}
	return total
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
		atomic.AddUint64(&vetoMap.shard[index].numVetoes, 1)
	}
}

func (vetoMap *VetoMap) GetVeto(vetoId uint64) routing.DecisionReason {
	index := vetoId % NumVetoMapShards
	vetoData, exists := vetoMap.shard[index].vetoes[vetoId]
	if exists {
		vetoData.timestamp = time.Now().Unix()
	}
	vetoMap.shard[index].vetoes[vetoId] = vetoData
	if exists {
		return vetoData.reason
	}
	return routing.DecisionNoReason
}

func (vetoMap *VetoMap) TimeoutLoop(ctx context.Context, timeoutSeconds int64, c <-chan time.Time) {
	maxIterations := 100
	deleteList := make([]uint64, maxIterations)
	for {
		select {
		case <-c:
			timeoutTimestamp := time.Now().Unix() - timeoutSeconds

			deleteList = deleteList[:0]

			for index := 0; index < NumVetoMapShards; index++ {
				vetoMap.shard[index].mutex.RLock()
				numIterations := 0
				for k, v := range vetoMap.shard[index].vetoes {
					if numIterations > maxIterations || numIterations > len(vetoMap.shard[index].vetoes) {
						break
					}
					if v.timestamp < timeoutTimestamp {
						fmt.Printf("timed out veto: %x\n", k)
						atomic.AddUint64(&vetoMap.shard[index].numVetoes, ^uint64(0))
						deleteList = append(deleteList, k)
					}
					numIterations++
				}
				vetoMap.shard[index].mutex.RUnlock()
				vetoMap.shard[index].mutex.Lock()
				for i := range deleteList {
					fmt.Printf("timeout veto %x\n", deleteList[i])
					delete(vetoMap.shard[index].vetoes, deleteList[i])
				}
				vetoMap.shard[index].mutex.Unlock()
			}
		case <-ctx.Done():
			return
		}
	}
}
