package storage

import (
	"encoding/base64"
	"encoding/binary"
)

type InMemory struct {
	RelayDatacenterNames map[uint32]string
}

func NewInMemory() InMemory {
	return InMemory{
		RelayDatacenterNames: make(map[uint32]string),
	}
}

func (m *InMemory) GetAndCheckBySdkVersion3PublicKeyId(key uint64) (*Buyer, bool) {
	publicKey, _ := base64.StdEncoding.DecodeString("leN7D7+9vr24uT4f1Ba8PEEvIQA/UkGZLlT+sdeLRHKsVqaZq723Zw==")

	return &Buyer{
		Name:                     "Network Next",
		SdkVersion3PublicKeyId:   binary.LittleEndian.Uint64(publicKey[:8]),
		SdkVersion3PublicKeyData: publicKey[8:],
		Active:                   true,
	}, true
}

func (m *InMemory) GetAndCheck(key *Key) (*Datacenter, bool) {
	// can't simulate a failed lookup here as well, at least not yet
	return &Datacenter{
		Name: key.PartitionId.Namespace,
	}, true
}

func (m *InMemory) GetAndCheckByRelayCoreId(key uint32) (*Relay, bool) {
	name, ok := m.RelayDatacenterNames[key]
	if !ok {
		return nil, false
	}

	return &Relay{
		Datacenter: &Key{
			PartitionId: &PartitionId{
				Namespace: name,
			},
		},
	}, true
}
