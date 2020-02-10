package storage

import (
	"encoding/binary"
)

type InMemory struct {
	LocalDatacenter bool

	LocalCustomerPublicKey []byte
	LocalRelayPublicKey    []byte

	RelayDatacenterNames map[uint32]string
	RelayPublicKeys      map[uint32][]byte
}

func (m *InMemory) GetAndCheckBySdkVersion3PublicKeyId(key uint64) (*Buyer, bool) {
	return &Buyer{
		Name:                     "Network Next",
		SdkVersion3PublicKeyId:   binary.LittleEndian.Uint64(m.LocalCustomerPublicKey[:8]),
		SdkVersion3PublicKeyData: m.LocalCustomerPublicKey[8:],
		Active:                   true,
	}, true
}

func (m *InMemory) GetAndCheck(key *Key) (*Datacenter, bool) {
	// can't simulate a failed lookup here as well, at least not yet
	return &Datacenter{
		Name: key.PartitionId.Namespace,
	}, true
}

// GetAndCheckByRelayCoreId gets the Relay from Configstore based on the hash
// of the Relay's address. The in memory implementation can be seeded with
// predetermined Relay address and IDs, but when running a relay to randomize
// a port this is imposible to guess. If the relay is not found in the seeded
// map we assume it is running locally.
//
// This is ONLY used in testing and local dev and in memory will NEVER run in
// production.
func (m *InMemory) GetAndCheckByRelayCoreId(key uint32) (*Relay, bool) {
	if m.LocalDatacenter {
		return &Relay{
			Datacenter: &Key{
				PartitionId: &PartitionId{
					Namespace: "local",
				},
			},
			UpdateKey: m.LocalRelayPublicKey,
		}, true
	}

	name, ok := m.RelayDatacenterNames[key]
	if !ok {
		return nil, false
	}

	relayPublicKey, ok := m.RelayPublicKeys[key]
	if !ok {
		return nil, false
	}

	return &Relay{
		Datacenter: &Key{
			PartitionId: &PartitionId{
				Namespace: name,
			},
		},
		UpdateKey: relayPublicKey,
	}, true
}
