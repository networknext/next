package transport

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/networknext/backend/routing"
)

const NumVetoMapShards = 4096

type VetoData struct {
	timestamp int64
	reason    routing.DecisionReason
}

type VetoMapShard struct {
	mutex     sync.Mutex
	vetoes    map[uint64]*VetoData
	numVetoes uint64
}

type VetoMap struct {
	shard [NumVetoMapShards]*VetoMapShard
}

func NewVetoMap() *VetoMap {
	vetoMap := &VetoMap{}
	for i := 0; i < NumVetoMapShards; i++ {
		vetoMap.shard[i] = &VetoMapShard{}
		vetoMap.shard[i].vetoes = make(map[uint64]*VetoData)
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

// IMPORTANT: Only needs to be called once. Each time you call "GetVeto" it refreshes the timestamp so you don't have to get and set each time. Efficient.
func (vetoMap *VetoMap) SetVeto(vetoId uint64, reason routing.DecisionReason) {
	index := vetoId % NumVetoMapShards
	vetoData := VetoData{
		timestamp: time.Now().Unix(),
		reason:    reason,
	}
	vetoMap.shard[index].mutex.Lock()
	_, exists := vetoMap.shard[index].vetoes[vetoId]
	vetoMap.shard[index].vetoes[vetoId] = &vetoData
	vetoMap.shard[index].mutex.Unlock()
	if !exists {
		atomic.AddUint64(&vetoMap.shard[index].numVetoes, 1)
	}
}

func (vetoMap *VetoMap) GetVeto(vetoId uint64) routing.DecisionReason {
	index := vetoId % NumServerMapShards
	vetoMap.shard[index].mutex.Lock()
	vetoData, _ := vetoMap.shard[index].vetoes[vetoId]
	if vetoData != nil {
		vetoData.timestamp = time.Now().Unix()
	}
	vetoMap.shard[index].mutex.Unlock()
	if vetoData != nil {
		return vetoData.reason
	}

	return routing.DecisionNoReason
}

func (vetoMap *VetoMap) TimeoutLoop(ctx context.Context, timeout time.Duration, c <-chan time.Time) {
	for {
		select {
		case <-c:
			timeoutTimestamp := time.Now().Add(-timeout).Unix()

			for index := 0; index < NumVetoMapShards; index++ {
				vetoTimeoutStart := time.Now()
				vetoMap.shard[index].mutex.Lock()
				numVetoIterations := 0
				for k, v := range vetoMap.shard[index].vetoes {
					if numVetoIterations > 100 {
						break
					}
					if v.timestamp < timeoutTimestamp {
						// fmt.Printf("timed out veto: %x\n", k)
						delete(vetoMap.shard[index].vetoes, k)
						atomic.AddUint64(&vetoMap.shard[index].numVetoes, ^uint64(0))
					}
					numVetoIterations++
				}
				vetoMap.shard[index].mutex.Unlock()
				if time.Since(vetoTimeoutStart).Seconds() > 0.1 {
					// fmt.Printf("long veto timeout check [%d]\n", index)
				}
			}
		case <-ctx.Done():
			return
		}
	}
}
