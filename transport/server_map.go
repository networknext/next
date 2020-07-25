package transport

import (
	"context"
	"sync"
	"sync/atomic"
	"time"
	"fmt"

	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/routing"
)

const NumServerMapShards = 4096

type ServerData struct {
	timestamp      int64
	routePublicKey []byte
	version        SDKVersion
	datacenter     routing.Datacenter
	sequence       uint64
}

type ServerMapShard struct {
	mutex      sync.Mutex
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

func (serverMap *ServerMap) UpdateServerData(buyerID uint64, serverAddress string, serverData *ServerData) {
	serverHash := crypto.HashID(serverAddress)
	index := (buyerID + serverHash) % NumServerMapShards
	serverMap.shard[index].mutex.Lock()
	_, exists := serverMap.shard[index].servers[serverAddress]
	serverMap.shard[index].servers[serverAddress] = serverData
	serverMap.shard[index].mutex.Unlock()
	if !exists {
		atomic.AddUint64(&serverMap.shard[index].numServers, 1)
	}
}

func (serverMap *ServerMap) GetServerData(buyerID uint64, serverAddress string) *ServerData {
	serverHash := crypto.HashID(serverAddress)
	index := (buyerID + serverHash) % NumServerMapShards
	serverMap.shard[index].mutex.Lock()
	serverData, _ := serverMap.shard[index].servers[serverAddress]
	serverMap.shard[index].mutex.Unlock()
	return serverData
}

func (serverMap *ServerMap) TimeoutLoop(ctx context.Context, timeout time.Duration, c <-chan time.Time) {
	for {
		select {
		case <-c:
			timeoutTimestamp := time.Now().Add(-timeout).Unix()

			for index := 0; index < NumServerMapShards; index++ {
				serverTimeoutStart := time.Now()
				serverMap.shard[index].mutex.Lock()
				numServerIterations := 0
				for k, v := range serverMap.shard[index].servers {
					if numServerIterations > 100 {
						break
					}
					if v.timestamp < timeoutTimestamp {
						fmt.Printf("timed out server: %x\n", k)
						delete(serverMap.shard[index].servers, k)
						atomic.AddUint64(&serverMap.shard[index].numServers, ^uint64(0))
					}
					numServerIterations++
				}
				serverMap.shard[index].mutex.Unlock()
				if time.Since(serverTimeoutStart).Seconds() > 0.1 {
					// fmt.Printf("long server timeout check [%d]\n", index)
				}
			}
		case <-ctx.Done():
			return
		}
	}
}
