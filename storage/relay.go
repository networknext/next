package storage

import (
	"encoding/json"
	"net"
)

//go:generate moq -out Relay_test_mocks.go . RelayStore
type RelayStore interface{
	Set(RelayStoreData) error
	ExpireReset(relayID uint64) error
	Get(relayID uint64) (*RelayStoreData,error)
	GetAll() ([]*RelayStoreData, error)
	Delete(relayID uint64) error
}

type RelayStoreData struct{
	ID uint64				`json:"id"`
	Address net.UDPAddr		`json:"address"`
	RelayVersion   string	`json:"version"`
}

func NewRelayStoreData(relayID uint64, relayVersion string, address net.UDPAddr) *RelayStoreData {
	data := new(RelayStoreData)
	data.ID = relayID
	data.Address = address
	data.RelayVersion = relayVersion
	return data
}

func RelayToJSON(relayData RelayStoreData) ([]byte, error){
	return json.Marshal(relayData)
}

func RelayFromJSON(data []byte) (*RelayStoreData,error){
	r := new(RelayStoreData)
	err := json.Unmarshal(data, r)
	return r, err
}

