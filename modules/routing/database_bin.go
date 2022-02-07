package routing

import (
	"encoding/binary"
	"fmt"
	"hash/fnv"
	"net"
	"sort"

	"github.com/networknext/backend/modules/encoding"
)

const (
	MaxDatabaseBinWrapperSize          = 100000000
	DatabaseBinWrapperReferenceVersion = 1
)

// DatabaseBinWrapper contains all the data from the database for
// static use by the relay_gateway, relay_backend, and server_backend
type DatabaseBinWrapper struct {
	SHA            string
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

func (wrapper DatabaseBinWrapper) WrapperToReference() DatabaseBinWrapperReference {
	dbReference := DatabaseBinWrapperReference{
		Version: DatabaseBinWrapperReferenceVersion,
	}

	dbReference.Buyers = make([]uint64, len(wrapper.BuyerMap))
	dbReference.Sellers = make([]string, len(wrapper.SellerMap))
	dbReference.Datacenters = make([]string, len(wrapper.DatacenterMap))
	dbReference.DatacenterMaps = make(map[uint64][]uint64, len(wrapper.DatacenterMaps))
	dbReference.Relays = make([]RelayReference, len(wrapper.Relays))
	dbReference.RelayMap = make(map[uint64]RelayReference, len(wrapper.RelayMap))

	index := 0
	for buyerID := range wrapper.BuyerMap {
		dbReference.Buyers[index] = buyerID

		// TODO: Clean this up or just switch over to using the datacenter maps from the wrapper - don't really want to do the latter due to excess data
		dcIndex := 0
		for dcID := range wrapper.DatacenterMaps[buyerID] {
			if _, ok := dbReference.DatacenterMaps[buyerID]; !ok {
				dbReference.DatacenterMaps[buyerID] = make([]uint64, len(wrapper.DatacenterMaps[buyerID]))
			}
			dbReference.DatacenterMaps[buyerID][dcIndex] = dcID
			dcIndex++
		}

		sort.Slice(dbReference.DatacenterMaps[buyerID], func(i, j int) bool {
			return dbReference.DatacenterMaps[buyerID][i] < dbReference.DatacenterMaps[buyerID][j]
		})
		index++
	}

	index = 0
	for sellerName := range wrapper.SellerMap {
		dbReference.Sellers[index] = sellerName
		index++
	}

	index = 0
	for _, datacenter := range wrapper.DatacenterMap {
		dbReference.Datacenters[index] = datacenter.Name
		index++
	}

	index = 0
	for _, relay := range wrapper.Relays {
		dbReference.Relays[index] = RelayReference{
			PublicIP:    relay.Addr,
			DisplayName: relay.Name,
		}
		index++
	}

	for relayID, relay := range wrapper.RelayMap {
		dbReference.RelayMap[relayID] = RelayReference{
			DisplayName: relay.Name,
			PublicIP:    relay.Addr,
		}
	}

	sort.Slice(dbReference.Buyers, func(i, j int) bool { return dbReference.Buyers[i] < dbReference.Buyers[j] })
	sort.Slice(dbReference.Sellers, func(i, j int) bool { return dbReference.Sellers[i] < dbReference.Sellers[j] })
	sort.Slice(dbReference.Datacenters, func(i, j int) bool { return dbReference.Datacenters[i] < dbReference.Datacenters[j] })
	sort.Slice(dbReference.Relays, func(i, j int) bool { return dbReference.Relays[i].DisplayName < dbReference.Relays[j].DisplayName })

	return dbReference
}

func (wrapper DatabaseBinWrapper) Hash() (uint64, error) {
	dbReference := wrapper.WrapperToReference()

	buffer := make([]byte, MaxDatabaseBinWrapperSize) // TODO: This is probably way to big
	ws, err := encoding.CreateWriteStream(buffer)
	if err != nil {
		return 0, err
	}

	if err := dbReference.Serialize(ws); err != nil {
		err := fmt.Errorf("DatabaseBinWrapper.Hash(): Something went wrong serializing the database reference: %v", err)
		return 0, err
	}

	ws.Flush()

	hasher := fnv.New64a()
	_, err = hasher.Write(buffer)
	if err != nil {
		err := fmt.Errorf("DatabaseBinWrapper.Hash(): Something went wrong hashing the serialized database reference: %v", err)
		return 0, err
	}

	return binary.LittleEndian.Uint64(buffer), nil
}

func (ref *DatabaseBinWrapperReference) Hash() (uint64, error) {
	buffer := make([]byte, MaxDatabaseBinWrapperSize) // TODO: This is probably way to big
	ws, err := encoding.CreateWriteStream(buffer)
	if err != nil {
		return 0, err
	}

	if err := ref.Serialize(ws); err != nil {
		err := fmt.Errorf("DatabaseBinWrapperReference.Hash(): Something went wrong serializing the database reference: %v", err)
		return 0, err
	}

	ws.Flush()

	hasher := fnv.New64a()
	_, err = hasher.Write(buffer)
	if err != nil {
		err := fmt.Errorf("DatabaseBinWrapperReference.Hash(): Something went wrong hashing the serialized database reference: %v", err)
		return 0, err
	}

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
		buyerID := ref.Buyers[i]
		stream.SerializeUint64(&buyerID)
	}

	numDCMapKeys := uint32(len(ref.DatacenterMaps))
	stream.SerializeUint32(&numDCMapKeys)

	dcMapKeys := make([]uint64, numDCMapKeys)

	index := 0
	for buyerID := range ref.DatacenterMaps {
		dcMapKeys[index] = buyerID
		index++
	}

	for i := uint32(0); i < numDCMapKeys; i++ {
		buyerID := dcMapKeys[i]
		stream.SerializeUint64(&buyerID)

		for _, dcID := range ref.DatacenterMaps[buyerID] {
			stream.SerializeUint64(&dcID)
		}
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

	relayKeys := make([]uint64, numRelayKeys)

	index = 0
	for relayID := range ref.RelayMap {
		relayKeys[index] = relayID
		index++
	}

	for _, relayID := range relayKeys {
		ip := ref.RelayMap[relayID].PublicIP
		name := ref.RelayMap[relayID].DisplayName
		stream.SerializeUint64(&relayID)
		stream.SerializeString(&name, MaxRelayNameLength)
		stream.SerializeAddress(&ip)
	}

	return stream.Error()
}
