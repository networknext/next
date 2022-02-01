package routing

import (
	"fmt"
	"net"

	"github.com/networknext/backend/modules/encoding"
)

const (
	MaxDatabaseBinWrapperSize = 100000000
)

// DatabaseBinWrapper contains all the data from the database for
// static use by the relay_gateway, relay_backend, and server_backend
type DatabaseBinWrapper struct {
	CreationTime   string
	Creator        string
	Relays         []Relay
	RelayMap       map[uint64]Relay
	BuyerMap       map[uint64]Buyer
	SellerMap      map[string]Seller
	DatacenterMap  map[uint64]Datacenter
	DatacenterMaps map[uint64]map[uint64]DatacenterMap
	//                 ^ Buyer.ID   ^ DatacenterMap map index
}

func CreateEmptyDatabaseBinWrapper() *DatabaseBinWrapper {
	wrapper := &DatabaseBinWrapper{
		CreationTime:   "",
		Creator:        "",
		Relays:         []Relay{},
		RelayMap:       make(map[uint64]Relay),
		BuyerMap:       make(map[uint64]Buyer),
		SellerMap:      make(map[string]Seller),
		DatacenterMap:  make(map[uint64]Datacenter),
		DatacenterMaps: make(map[uint64]map[uint64]DatacenterMap),
	}

	return wrapper
}

func (wrapper DatabaseBinWrapper) IsEmpty() bool {
	if len(wrapper.RelayMap) != 0 {
		return false
	} else if len(wrapper.BuyerMap) != 0 {
		return false
	} else if len(wrapper.SellerMap) != 0 {
		return false
	} else if len(wrapper.DatacenterMap) != 0 {
		return false
	} else if len(wrapper.DatacenterMaps) != 0 {
		return false
	} else if wrapper.CreationTime == "" {
		return false
	} else if wrapper.Creator == "" {
		return false
	} else if len(wrapper.Relays) != 0 {
		return false
	}

	return true
}

const (
	DatabaseBinWrapperReferenceVersion = 1
)

type DatabaseBinWrapperReference struct {
	Version        uint32
	Relays         []RelayReference
	RelayMap       map[uint64]RelayReference
	Buyers         []uint64
	Sellers        []string
	Datacenters    []string
	DatacenterMaps map[uint64][]uint64
}

type RelayReference struct {
	PublicIP    net.UDPAddr
	DisplayName string
}

func (ref *DatabaseBinWrapperReference) Serialize(stream encoding.Stream) error {
	stream.SerializeUint32(&ref.Version)

	fmt.Println("Serializing relays")

	numRelays := uint32(len(ref.Relays))
	stream.SerializeUint32(&numRelays)

	for i := uint32(0); i < numRelays; i++ {
		stream.SerializeString(&ref.Relays[i].DisplayName, MaxRelayNameLength)
		stream.SerializeAddress(&ref.Relays[i].PublicIP)
	}

	fmt.Println("Serializing buyers")

	numBuyers := uint32(len(ref.Buyers))
	stream.SerializeUint32(&numBuyers)

	for i := uint32(0); i < numBuyers; i++ {
		stream.SerializeUint64(&ref.Buyers[i])
	}

	fmt.Println("Serializing sellers")

	numSellers := uint32(len(ref.Sellers))
	stream.SerializeUint32(&numSellers)

	for i := uint32(0); i < numSellers; i++ {
		stream.SerializeString(&ref.Sellers[i], 32) // TODO: Make const
	}

	fmt.Println("Serializing DCs")

	numDatacenters := uint32(len(ref.Datacenters))
	stream.SerializeUint32(&numDatacenters)

	for i := uint32(0); i < numDatacenters; i++ {
		stream.SerializeString(&ref.Datacenters[i], MaxRelayNameLength) // TODO: Make into its own const
	}

	fmt.Println("Serializing relay map")

	numRelayKeys := uint32(len(ref.RelayMap))
	stream.SerializeUint32(&numRelayKeys)

	for relayID, relayRef := range ref.RelayMap {
		stream.SerializeUint64(&relayID)
		stream.SerializeString(&relayRef.DisplayName, MaxRelayNameLength)
		stream.SerializeAddress(&relayRef.PublicIP)
	}

	fmt.Println("Serializing DC map")

	numDCMapKeys := uint32(len(ref.DatacenterMaps))
	stream.SerializeUint32(&numDCMapKeys)

	for buyerID, maps := range ref.DatacenterMaps {
		stream.SerializeUint64(&buyerID)

		numMaps := uint32(len(maps))
		stream.SerializeUint32(&numMaps)

		for i := uint32(0); i < numMaps; i++ {
			stream.SerializeUint64(&maps[i])
		}
	}

	return stream.Error()
}
