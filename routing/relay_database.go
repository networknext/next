package routing

import (
	"time"
)

type RelayDatabase struct {
	Relays map[uint64]Relay
}

func NewRelayDatabase() *RelayDatabase {
	database := &RelayDatabase{}
	database.Relays = make(map[uint64]Relay)
	return database
}

// UpdateRelay updates the relay who's id is within the update data
func (database *RelayDatabase) UpdateRelay(update *RelayUpdate) bool {
	id := update.ID
	if update.Shutdown == true {
		delete(database.Relays, id)
		return false
	}
	relay, relayExistedAlready := database.Relays[id]
	relay.ID = update.ID
	relay.Name = update.Name
	relay.Addr = update.Address
	relay.PublicKey = update.PublicKey
	relay.LastUpdateTime = uint64(time.Now().Unix())
	relay.Datacenter = update.Datacenter
	relay.DatacenterName = update.DatacenterName
	database.Relays[id] = relay
	return !relayExistedAlready
}

// CheckForTimeouts loops over all relays and if any exceed the timout then they are removed
func (database *RelayDatabase) CheckForTimeouts(timeoutSeconds int) []uint64 {
	disconnected := make([]uint64, 0)
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
