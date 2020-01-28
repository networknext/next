package storage

import "log"

type InMemoryRelayStore struct {
	RelaysToDatacenterName map[uint32]string
}

type InMemoryDatacenterStore struct {
}

type InMemory struct {
	RelayStore      InMemoryRelayStore
	DatacenterStore InMemoryDatacenterStore
}

func NewInMemory() InMemory {
	return InMemory{
		RelayStore: InMemoryRelayStore{
			RelaysToDatacenterName: make(map[uint32]string),
		},
	}
}

func (m *InMemory) GetAndCheckBySdkVersion3PublicKeyId(key uint64) (*Buyer, bool) {
	return &Buyer{
		Name:                   "Network Next",
		SdkVersion3PublicKeyId: 13672574147039585173,
		SdkVersion3PublicKeyData: []byte{
			0xb8, 0xb9, 0x3e, 0x1f, 0xd4, 0x16, 0xbc, 0x3c, 0x41, 0x2f, 0x21, 0x00, 0x3f, 0x52, 0x41, 0x99,
			0x2e, 0x54, 0xfe, 0xb1, 0xd7, 0x8b, 0x44, 0x72, 0xac, 0x56, 0xa6, 0x99, 0xab, 0xbd, 0xb7, 0x67,
		},
	}, true
}

func (m *InMemoryDatacenterStore) GetAndCheck(key *Key) (*Datacenter, bool) {
	return &Datacenter{
		Name: key.PartitionId.Namespace,
	}, true
}

func (m *InMemoryRelayStore) GetAndCheckByRelayCoreId(key uint32) (*Relay, bool) {
	name, ok := m.RelaysToDatacenterName[key]
	if ok {
		log.Printf("Found stubbed data for relay: %d", key)
	}

	return &Relay{
		Datacenter: &Key{
			PartitionId: &PartitionId{
				Namespace: name,
			},
		},
	}, true
}
