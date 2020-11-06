package storage

import (
	"fmt"
	"sync"
	)

type RelayCache struct{
	mutex sync.RWMutex
	mapStore map[uint64]RelayStoreData
	arrStore []RelayStoreData
}

func NewRelayCache() *RelayCache{
	rc := new(RelayCache)
	rc.mapStore = make(map[uint64]RelayStoreData)
	return rc
}

func (rc *RelayCache) SetAll(relayArr []*RelayStoreData) error {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()

	var newArr []RelayStoreData
	newMap := make(map[uint64]RelayStoreData)
	for _, relay := range relayArr {
		newArr = append(newArr, *relay)
		newMap[relay.ID] = *relay
	}
	
	rc.mapStore = newMap
	rc.arrStore = newArr
	
	return nil
}

func (rc *RelayCache) SetAllWithAddRemove(relayArr []*RelayStoreData) ([]*RelayStoreData,[]string, error) {
	rc.mutex.Lock()
	defer rc.mutex.Unlock()

	var addArr []*RelayStoreData
	var newArr []RelayStoreData
	newMap := make(map[uint64]RelayStoreData)
	for _, relay := range relayArr {
		newArr = append(newArr, *relay)
		newMap[relay.ID] = *relay

		if _ , ok := rc.mapStore[relay.ID]; ok{
			delete(rc.mapStore, relay.ID)
		}else{
			addArr = append(addArr, relay)
		}
	}

	var removeArr []string
	for _, relay := range rc.mapStore{
		removeArr = append(removeArr,relay.Address.String())
	}

	rc.mapStore = newMap
	rc.arrStore = newArr

	return addArr, removeArr, nil
}

func (rc *RelayCache) GetAll() ([]RelayStoreData,error) {
	rc.mutex.RLock()
	defer rc.mutex.RUnlock()
	
	if len(rc.arrStore)<1{
		return []RelayStoreData{}, fmt.Errorf("no relays stored in cache")
	}
	return rc.arrStore, nil
}

func (rc *RelayCache) Get(relayID uint64) (RelayStoreData, error){
	rc.mutex.RLock()
	defer rc.mutex.RUnlock()
	
	if relay, ok := rc.mapStore[relayID]; ok{
		return relay,nil
	}
	
	return RelayStoreData{}, fmt.Errorf("Relay not found in cache")
}