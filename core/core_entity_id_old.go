package core

import (
	"fmt"
	"hash/fnv"
)

// todo: this whole old entity id / relay id etc. is incompatible with how the new backend should work. talk to me to learn more. -- glenn

type EntityID struct {
	Kind string
	Name string
}

type RelayID uint64

func GetRelayIDFromEntityID(id *EntityID) (RelayID, error) {
	if id.Kind != "Relay" {
		return RelayID(0), fmt.Errorf("not a valid relay: %+v", id)
	}
	hash := fnv.New64a()
	hash.Write([]byte(id.Name))
	return RelayID(hash.Sum64()), nil
}

type DatacenterID uint64

func GetDatacenterID(id *EntityID) (DatacenterID, error) {
	if id.Kind != "Datacenter" {
		return DatacenterID(0), fmt.Errorf("not a valid datacenter: %+v", id)
	}
	hash := fnv.New64a()
	hash.Write([]byte(id.Name))
	return DatacenterID(hash.Sum64()), nil
}
