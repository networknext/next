package core

import (
	"fmt"
	"hash/fnv"
)

// todo: this whole old entity id / relay id etc. is incompatible with how the new backend should work. talk to me to learn more. -- glenn

type EntityId struct {
	Kind string
	Name string
}

type RelayId uint64

func GetRelayId(id *EntityId) (RelayId, error) {
	if id.Kind != "Relay" {
		return RelayId(0), fmt.Errorf("not a valid relay: %+v", id)
	}
	hash := fnv.New64a()
	hash.Write([]byte(id.Name))
	return RelayId(hash.Sum64()), nil
}

type DatacenterId uint64

func GetDatacenterId(id *EntityId) (DatacenterId, error) {
	if id.Kind != "Datacenter" {
		return DatacenterId(0), fmt.Errorf("not a valid datacenter: %+v", id)
	}
	hash := fnv.New64a()
	hash.Write([]byte(id.Name))
	return DatacenterId(hash.Sum64()), nil
}
