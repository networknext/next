package transport

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
	// "fmt"

	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/routing"
)

const NumServerMapShards = 1000000

type ServerData struct {
	Timestamp      int64
	RoutePublicKey []byte
	Version        SDKVersion
	Datacenter     routing.Datacenter
	Sequence       uint64
}

type ServerMapShard struct {
	mutex      sync.RWMutex
	servers    map[string]*ServerData
	numServers uint64
}

type ServerMap struct {
	shard [NumServerMapShards]*ServerMapShard
}

func NewServerMap() *ServerMap {
	serverMap := &ServerMap{}
	for i := 0; i < NumServerMapShards; i++ {
		serverMap.shard[i] = &ServerMapShard{}
		serverMap.shard[i].servers = make(map[string]*ServerData)
	}
	return serverMap
}

func (serverMap *ServerMap) NumServers() uint64 {
	var total uint64
	for i := 0; i < NumServerMapShards; i++ {
		numServersInShard := atomic.LoadUint64(&serverMap.shard[i].numServers)
		total += numServersInShard
	}
	return total
}

func (serverMap *ServerMap) RLock(buyerID uint64, serverAddress string) {
	serverHash := crypto.HashID(serverAddress)
	index := (buyerID + serverHash) % NumServerMapShards
	serverMap.shard[index].mutex.RLock()
}

func (serverMap *ServerMap) RUnlock(buyerID uint64, serverAddress string) {
	serverHash := crypto.HashID(serverAddress)
	index := (buyerID + serverHash) % NumServerMapShards
	serverMap.shard[index].mutex.RUnlock()
}

func (serverMap *ServerMap) Lock(buyerID uint64, serverAddress string) {
	serverHash := crypto.HashID(serverAddress)
	index := (buyerID + serverHash) % NumServerMapShards
	serverMap.shard[index].mutex.Lock()
}

func (serverMap *ServerMap) Unlock(buyerID uint64, serverAddress string) {
	serverHash := crypto.HashID(serverAddress)
	index := (buyerID + serverHash) % NumServerMapShards
	serverMap.shard[index].mutex.Unlock()
}

func (serverMap *ServerMap) UpdateServerData(buyerID uint64, serverAddress string, serverData *ServerData) {
	serverHash := crypto.HashID(serverAddress)
	index := (buyerID + serverHash) % NumServerMapShards
	_, exists := serverMap.shard[index].servers[serverAddress]
	serverMap.shard[index].servers[serverAddress] = serverData
	if !exists {
		atomic.AddUint64(&serverMap.shard[index].numServers, 1)
	}
}

func (serverMap *ServerMap) GetServerData(buyerID uint64, serverAddress string) *ServerData {
	serverHash := crypto.HashID(serverAddress)
	index := (buyerID + serverHash) % NumServerMapShards
	serverData, _ := serverMap.shard[index].servers[serverAddress]
	return serverData
}

func (serverMap *ServerMap) TimeoutLoop(ctx context.Context, timeoutSeconds int64, c <-chan time.Time) {
	maxIterations := 100
	deleteList := make([]string, maxIterations)
	for {
		select {
		case <-c:
			timeoutTimestamp := time.Now().Unix() - timeoutSeconds

			for index := 0; index < NumServerMapShards; index++ {
				deleteList = deleteList[:0]
				serverMap.shard[index].mutex.RLock()
				numIterations := 0
				for k, v := range serverMap.shard[index].servers {
					if numIterations > maxIterations || numIterations > len(serverMap.shard[index].servers) {
						break
					}
					if v.Timestamp < timeoutTimestamp {
						atomic.AddUint64(&serverMap.shard[index].numServers, ^uint64(0))
						deleteList = append(deleteList, k)
					}
					numIterations++
				}
				serverMap.shard[index].mutex.RUnlock()
				serverMap.shard[index].mutex.Lock()
				for i := range deleteList {
					// fmt.Printf("timeout server %x\n", deleteList[i])
					delete(serverMap.shard[index].servers, deleteList[i])
				}
				serverMap.shard[index].mutex.Unlock()
			}
		case <-ctx.Done():
			return
		}
	}
}
