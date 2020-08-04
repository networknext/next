package transport

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/routing"
)

const NumServerMapShards = 100000

type ServerData struct {
	Timestamp      int64
	RoutePublicKey []byte
	Version        routing.SDKVersion
	Datacenter     routing.Datacenter
	Sequence       uint64
}

type ServerMapShard struct {
	mutex   sync.RWMutex
	servers map[string]*ServerData
}

type ServerMap struct {
	numServers   uint64
	timeoutShard int
	shard        [NumServerMapShards]*ServerMapShard
}

func NewServerMap() *ServerMap {
	serverMap := &ServerMap{}
	for i := 0; i < NumServerMapShards; i++ {
		serverMap.shard[i] = &ServerMapShard{}
		serverMap.shard[i].servers = make(map[string]*ServerData)
	}
	return serverMap
}

func (serverMap *ServerMap) GetServerCount() uint64 {
	return atomic.LoadUint64(&serverMap.numServers)
}

func (serverMap *ServerMap) RLock(buyerID uint64, serverAddress string) {
	serverHash := crypto.HashID(serverAddress)
	index := serverHash % NumServerMapShards
	serverMap.shard[index].mutex.RLock()
}

func (serverMap *ServerMap) RUnlock(buyerID uint64, serverAddress string) {
	serverHash := crypto.HashID(serverAddress)
	index := serverHash % NumServerMapShards
	serverMap.shard[index].mutex.RUnlock()
}

func (serverMap *ServerMap) Lock(buyerID uint64, serverAddress string) {
	serverHash := crypto.HashID(serverAddress)
	index := serverHash % NumServerMapShards
	serverMap.shard[index].mutex.Lock()
}

func (serverMap *ServerMap) Unlock(buyerID uint64, serverAddress string) {
	serverHash := crypto.HashID(serverAddress)
	index := serverHash % NumServerMapShards
	serverMap.shard[index].mutex.Unlock()
}

func (serverMap *ServerMap) UpdateServerData(buyerID uint64, serverAddress string, serverData *ServerData) {
	serverHash := crypto.HashID(serverAddress)
	index := serverHash % NumServerMapShards
	_, exists := serverMap.shard[index].servers[serverAddress]
	serverMap.shard[index].servers[serverAddress] = serverData
	if !exists {
		atomic.AddUint64(&serverMap.numServers, 1)
	}
}

func (serverMap *ServerMap) GetServerData(buyerID uint64, serverAddress string) *ServerData {
	serverHash := crypto.HashID(serverAddress)
	index := serverHash % NumServerMapShards
	serverData, _ := serverMap.shard[index].servers[serverAddress]
	return serverData
}

func (serverMap *ServerMap) TimeoutLoop(ctx context.Context, timeoutSeconds int64, c <-chan time.Time) {
	maxShards := 100
	maxIterations := 10
	deleteList := make([]string, maxIterations)
	for {
		select {
		case <-c:
			timeoutTimestamp := time.Now().Unix() - timeoutSeconds
			for i := 0; i < maxShards; i++ {
				index := (serverMap.timeoutShard + i) % NumServerMapShards
				deleteList = deleteList[:0]
				serverMap.shard[index].mutex.RLock()
				numIterations := 0
				for k, v := range serverMap.shard[index].servers {
					if numIterations >= maxIterations || numIterations >= len(serverMap.shard[index].servers) {
						break
					}
					if v.Timestamp < timeoutTimestamp {
						deleteList = append(deleteList, k)
					}
					numIterations++
				}
				serverMap.shard[index].mutex.RUnlock()
				if len(deleteList) > 0 {
					serverMap.shard[index].mutex.Lock()
					for i := range deleteList {
						// fmt.Printf("timeout server %x\n", deleteList[i])
						delete(serverMap.shard[index].servers, deleteList[i])
						atomic.AddUint64(&serverMap.numServers, ^uint64(0))
					}
					serverMap.shard[index].mutex.Unlock()
				}
			}
			serverMap.timeoutShard = (serverMap.timeoutShard + maxShards) % NumServerMapShards
		case <-ctx.Done():
			return
		}
	}
}
