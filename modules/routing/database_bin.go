package routing

import (
	"encoding/binary"
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

func (wrapper DatabaseBinWrapper) Hash() (uint64, error) {
	dbReference := DatabaseBinWrapperReference{
		Version: DatabaseBinWrapperReferenceVersion,
	}

	dbReference.Buyers = make([]uint64, len(wrapper.BuyerMap))
	dbReference.Sellers = make([]string, len(wrapper.SellerMap))
	dbReference.Relays = make([]RelayReference, len(wrapper.Relays))
	dbReference.RelayMap = make(map[uint64]RelayReference)
	dbReference.Datacenters = make([]string, len(wrapper.DatacenterMap))
	dbReference.DatacenterMaps = make(map[uint64][]uint64, len(wrapper.DatacenterMaps))
	//                                     ^ DC ID   ^ Buyer IDs

	// TODO: Not sure if this is the best route here -> might be better just to serialize the whole dc map rather than creating a new one
	index := 0
	for buyerID := range wrapper.BuyerMap {
		dbReference.Buyers[index] = buyerID

		for dcID, dcMaps := range wrapper.DatacenterMaps {
			if _, ok := dcMaps[buyerID]; !ok {
				continue
			}

			if _, ok := dbReference.DatacenterMaps[dcID]; !ok {
				dbReference.DatacenterMaps[dcID] = make([]uint64, 0)
			}

			dbReference.DatacenterMaps[dcID] = append(dbReference.DatacenterMaps[dcID], buyerID)
		}

		index++
	}

	index = 0
	for shortName := range wrapper.SellerMap {
		dbReference.Sellers[index] = shortName
		index++
	}

	for index, relay := range wrapper.Relays {
		dbReference.Relays[index] = RelayReference{
			PublicIP:    relay.Addr,
			DisplayName: relay.Name,
		}
	}

	// TODO: This may be able to be combined with the above for loop
	for relayID, relay := range wrapper.RelayMap {
		dbReference.RelayMap[relayID] = RelayReference{
			PublicIP:    relay.Addr,
			DisplayName: relay.Name,
		}
	}

	index = 0
	for _, datacenter := range wrapper.DatacenterMap {
		dbReference.Datacenters[index] = datacenter.AliasName
		index++
	}

	buffer := make([]byte, MaxDatabaseBinWrapperSize) // TODO: This is probably way to big
	ws, err := encoding.CreateWriteStream(buffer)
	if err != nil {
		return 0, err
	}

	if err := dbReference.Serialize(ws); err != nil {
		fmt.Println("Something went wrong serializing the db reference")
		return 0, err
	}

	// TODO: Not sure if this is really "hashing" or not - other methods (sha1 and fnv) returned different values each time (expected based off cpu clock etc?)
	return binary.LittleEndian.Uint64(buffer), nil
}

const (
	DatabaseBinWrapperReferenceVersion = 1
)

// TODO: add more fields to !Relays
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

func (ref *DatabaseBinWrapperReference) Hash() (uint64, error) {
	buffer := make([]byte, MaxDatabaseBinWrapperSize) // TODO: This is probably way to big
	ws, err := encoding.CreateWriteStream(buffer)
	if err != nil {
		return 0, err
	}

	if err := ref.Serialize(ws); err != nil {
		fmt.Println("Something went wrong serializing the db reference")
		return 0, err
	}

	// TODO: Not sure if this is really "hashing" or not - other methods (sha1 and fnv) returned different values each time (expected based off cpu clock etc?)
	return binary.LittleEndian.Uint64(buffer), nil
}

func (ref *DatabaseBinWrapperReference) Serialize(stream encoding.Stream) error {
	stream.SerializeUint32(&ref.Version)

	numRelays := uint32(len(ref.Relays))
	stream.SerializeUint32(&numRelays)

	for i := uint32(0); i < numRelays; i++ {
		stream.SerializeString(&ref.Relays[i].DisplayName, MaxRelayNameLength)
		stream.SerializeAddress(&ref.Relays[i].PublicIP)
	}

	numBuyers := uint32(len(ref.Buyers))
	stream.SerializeUint32(&numBuyers)

	for i := uint32(0); i < numBuyers; i++ {
		stream.SerializeUint64(&ref.Buyers[i])
	}

	numSellers := uint32(len(ref.Sellers))
	stream.SerializeUint32(&numSellers)

	for i := uint32(0); i < numSellers; i++ {
		stream.SerializeString(&ref.Sellers[i], 32) // TODO: Make const
	}

	numDatacenters := uint32(len(ref.Datacenters))
	stream.SerializeUint32(&numDatacenters)

	for i := uint32(0); i < numDatacenters; i++ {
		stream.SerializeString(&ref.Datacenters[i], MaxRelayNameLength) // TODO: Make into its own const
	}

	numRelayKeys := uint32(len(ref.RelayMap))
	stream.SerializeUint32(&numRelayKeys)

	for relayID, relayRef := range ref.RelayMap {
		stream.SerializeUint64(&relayID)
		stream.SerializeString(&relayRef.DisplayName, MaxRelayNameLength)
		stream.SerializeAddress(&relayRef.PublicIP)
	}

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
