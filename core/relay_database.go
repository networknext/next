package core

type RelayUpdate struct {
	Id             RelayId
	Name           string
	Address        string
	Datacenter     DatacenterId
	DatacenterName string
	PublicKey      []byte
	Shutdown       bool
}

type RelayData struct {
	Id             RelayId
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

func (database *RelayDatabase) UpdateRelay(update *RelayUpdate) bool {
	id := update.Id
	if update.Shutdown == true {
		delete(database.Relays, id)
		return false
	}
	relayData, relayExistedAlready := database.Relays[id]
	relayData.Id = update.Id
	relayData.Name = update.Name
	relayData.Address = update.Address
	relayData.PublicKey = update.PublicKey
	relayData.LastUpdateTime = uint64(time.Now().Unix())
	relayData.Datacenter = update.Datacenter
	relayData.DatacenterName = update.DatacenterName
	database.Relays[id] = relayData
	return !relayExistedAlready
}

func (database *RelayDatabase) CheckForTimeouts(timeoutSeconds int) []RelayId {
	disconnected := make([]RelayId, 0)
	currentTime := uint64(time.Now().Unix())
	for k, v := range database.Relays {
		if v.LastUpdateTime+uint64(timeoutSeconds) <= currentTime {
			disconnected = append(disconnected, v.Id)
			delete(database.Relays, k)
		}
	}
	return disconnected
}

func (database *RelayDatabase) MakeCopy() *RelayDatabase {
	database_copy := NewRelayDatabase()
	for k, v := range database.Relays {
		database_copy.Relays[k] = v
	}
	return database_copy
}
