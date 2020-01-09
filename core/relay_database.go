package core

import (
	"hash/fnv"
	"time"
)

type RelayUpdate struct {
	ID             RelayId
	Name           string
	Address        string
	Datacenter     DatacenterId
	DatacenterName string
	PublicKey      []byte
	Shutdown       bool
}

type RelayData struct {
	ID             RelayId
	Name           string
	Address        string
	Datacenter     DatacenterId
	DatacenterName string
	PublicKey      []byte
	LastUpdateTime uint64
}

type RelayDatabase struct {
	Relays map[RelayId]RelayData
}

func NewRelayDatabase() *RelayDatabase {
	database := &RelayDatabase{}
	database.Relays = make(map[RelayId]RelayData)
	return database
}

// UpdateRelay updates the relay who's id is within the update data
func (database *RelayDatabase) UpdateRelay(update *RelayUpdate) bool {
	id := update.ID
	if update.Shutdown == true {
		delete(database.Relays, id)
		return false
	}
	relayData, relayExistedAlready := database.Relays[id]
	relayData.ID = update.ID
	relayData.Name = update.Name
	relayData.Address = update.Address
	relayData.PublicKey = update.PublicKey
	relayData.LastUpdateTime = uint64(time.Now().Unix())
	relayData.Datacenter = update.Datacenter
	relayData.DatacenterName = update.DatacenterName
	database.Relays[id] = relayData
	return !relayExistedAlready
}

// CheckForTimeouts loops over all relays and if any exceed the timout then they are removed
func (database *RelayDatabase) CheckForTimeouts(timeoutSeconds int) []RelayId {
	disconnected := make([]RelayId, 0)
	currentTime := uint64(time.Now().Unix())
	for k, v := range database.Relays {
		if v.LastUpdateTime+uint64(timeoutSeconds) <= currentTime {
			disconnected = append(disconnected, v.ID)
			delete(database.Relays, k)
		}
	}
	return disconnected
}

// MakeCopy makes a new relay database whose contents are identical to the calling db
func (database *RelayDatabase) MakeCopy() *RelayDatabase {
	databaseCopy := NewRelayDatabase()
	for k, v := range database.Relays {
		databaseCopy.Relays[k] = v
	}
	return databaseCopy
}

// GetRelayID hashes the name of the relay and returns the result. Typically name is the address of the relay
func GetRelayID(name string) RelayId {
	hash := fnv.New64a()
	hash.Write([]byte(name))
	return RelayId(hash.Sum64())
}
