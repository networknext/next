package storage

import (
	"encoding/base64"
	"encoding/binary"
	"log"
	"os"
)

type InMemory struct {
	LocalDatacenter bool

	RelayDatacenterNames map[uint32]string
	RelayPublicKeys      map[uint32][]byte
}

func (m *InMemory) GetAndCheckBySdkVersion3PublicKeyId(key uint64) (*Buyer, bool) {
	if k := os.Getenv("NEXT_CUSTOMER_PUBLIC_KEY"); len(k) != 0 {
		publicKey, _ := base64.StdEncoding.DecodeString(k)

		return &Buyer{
			Name:                     "Network Next",
			SdkVersion3PublicKeyId:   binary.LittleEndian.Uint64(publicKey[:8]),
			SdkVersion3PublicKeyData: publicKey[8:],
			Active:                   true,
		}, true
	}

	return nil, false
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
	var localRelayPublicKey []byte
	{
		if key := os.Getenv("RELAY_PUBLIC_KEY"); len(key) != 0 {
			localRelayPublicKey, _ = base64.StdEncoding.DecodeString(key)
		} else {
			log.Fatal("env var 'RELAY_PUBLIC_KEY' is not set")
		}
	}

	if m.LocalDatacenter {
		return &Relay{
			Datacenter: &Key{
				PartitionId: &PartitionId{
					Namespace: "local",
				},
			},
			UpdateKey: localRelayPublicKey,
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
