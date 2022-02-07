package routing_test

import (
	"context"
	"encoding/binary"
	"fmt"
	"net"
	"testing"

	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/storage"
	"github.com/stretchr/testify/assert"
)

func SetupReferenceWrapper(t *testing.T, numBuyers int, numSellers int, numRelays int, numDatacenters int) routing.DatabaseBinWrapper {
	wrapper := routing.CreateEmptyDatabaseBinWrapper()

	wrapper.BuyerMap = make(map[uint64]routing.Buyer)

	for i := 0; i < numBuyers; i++ {
		buyerID := uint64(i + 1)
		wrapper.BuyerMap[buyerID] = routing.Buyer{
			ID: buyerID,
		}
	}

	wrapper.SellerMap = make(map[string]routing.Seller)

	for i := 0; i < numSellers; i++ {
		sellerName := fmt.Sprintf("seller%d", i+1)
		wrapper.SellerMap[sellerName] = routing.Seller{
			ID:        sellerName,
			Name:      sellerName,
			ShortName: sellerName,
		}
	}

	wrapper.DatacenterMap = make(map[uint64]routing.Datacenter)
	wrapper.DatacenterMaps = make(map[uint64]map[uint64]routing.DatacenterMap)

	for i := 0; i < numDatacenters; i++ {
		dcName := fmt.Sprintf("datacenter%d", i+1)
		dcID := crypto.HashID(dcName)

		wrapper.DatacenterMap[dcID] = routing.Datacenter{
			ID:   dcID,
			Name: dcName,
		}
	}

	for buyer := range wrapper.BuyerMap {
		wrapper.DatacenterMaps[buyer] = make(map[uint64]routing.DatacenterMap)
		dcName := fmt.Sprintf("datacenter%d", buyer)
		wrapper.DatacenterMaps[buyer][crypto.HashID(dcName)] = routing.DatacenterMap{
			BuyerID:      buyer,
			DatacenterID: crypto.HashID(dcName),
		}
	}

	wrapper.RelayMap = make(map[uint64]routing.Relay)
	wrapper.Relays = make([]routing.Relay, numRelays)

	for i := 0; i < numRelays; i++ {
		relayID := uint64(i + 1)

		relayAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.%d:10000", i+1))
		assert.NoError(t, err)

		wrapper.RelayMap[relayID] = routing.Relay{
			ID:         relayID,
			Name:       fmt.Sprintf("relay%d", i+1),
			Addr:       *relayAddr,
			Seller:     wrapper.SellerMap[fmt.Sprintf("seller%d", i+1)],
			Datacenter: wrapper.DatacenterMap[crypto.HashID(fmt.Sprintf("datacenter%d", i+1))],
		}

		wrapper.Relays[i] = wrapper.RelayMap[relayID]
	}

	assert.False(t, wrapper.IsEmpty())

	return *wrapper
}

func SetupStorageWrapper(t *testing.T, numBuyers int, numSellers int, numRelays int, numDatacenters int) routing.DatabaseBinWrapper {
	var storer = storage.InMemory{}
	ctx := context.Background()

	for i := 0; i < numBuyers; i++ {
		err := storer.AddBuyer(ctx, routing.Buyer{
			ID: uint64(i + 1),
		})
		assert.NoError(t, err)
	}

	for i := 0; i < numSellers; i++ {
		sellerName := fmt.Sprintf("seller%d", i+1)
		err := storer.AddSeller(ctx, routing.Seller{
			ID:        sellerName,
			Name:      sellerName,
			ShortName: sellerName,
		})
		assert.NoError(t, err)
	}

	for i := 0; i < numDatacenters; i++ {
		dcName := fmt.Sprintf("datacenter%d", i+1)
		err := storer.AddDatacenter(ctx, routing.Datacenter{
			ID:   crypto.HashID(dcName),
			Name: dcName,
		})
		assert.NoError(t, err)
	}

	for i := 0; i < numRelays; i++ {
		relayAddr, err := net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.%d:10000", i+1))
		assert.NoError(t, err)

		sellerName := fmt.Sprintf("seller%d", i+1)
		dcName := fmt.Sprintf("datacenter%d", i+1)

		err = storer.AddRelay(ctx, routing.Relay{
			ID:   uint64(i + 1),
			Name: fmt.Sprintf("relay%d", i+1),
			Addr: *relayAddr,
			Seller: routing.Seller{
				ID:        sellerName,
				Name:      sellerName,
				ShortName: sellerName,
			},
			Datacenter: routing.Datacenter{
				ID:   crypto.HashID(dcName),
				Name: dcName,
			},
		})
		assert.NoError(t, err)
	}

	for i := 0; i < numDatacenters; i++ {
		dcName := fmt.Sprintf("datacenter%d", i+1)
		err := storer.AddDatacenterMap(ctx, routing.DatacenterMap{
			BuyerID:      uint64(i + 1),
			DatacenterID: crypto.HashID(dcName),
		})
		assert.NoError(t, err)
	}

	wrapper := routing.CreateEmptyDatabaseBinWrapper()

	wrapper.BuyerMap = make(map[uint64]routing.Buyer)
	wrapper.DatacenterMaps = make(map[uint64]map[uint64]routing.DatacenterMap)

	allBuyers := storer.Buyers(ctx)
	assert.Equal(t, numBuyers, len(allBuyers))

	for _, buyer := range allBuyers {
		wrapper.BuyerMap[buyer.ID] = buyer
		wrapper.DatacenterMaps[buyer.ID] = storer.GetDatacenterMapsForBuyer(ctx, buyer.ID)
	}

	wrapper.SellerMap = make(map[string]routing.Seller)

	allSellers := storer.Sellers(ctx)
	assert.Equal(t, numSellers, len(allSellers))

	for _, seller := range allSellers {
		wrapper.SellerMap[seller.ShortName] = seller
	}

	wrapper.RelayMap = make(map[uint64]routing.Relay)
	wrapper.Relays = make([]routing.Relay, numRelays)

	allRelays := storer.Relays(ctx)
	assert.Equal(t, numRelays, len(allRelays))

	for i, relay := range allRelays {
		wrapper.RelayMap[relay.ID] = relay
		wrapper.Relays[i] = relay
	}

	wrapper.DatacenterMap = make(map[uint64]routing.Datacenter)

	allDatacenters := storer.Datacenters(ctx)
	assert.Equal(t, numDatacenters, len(allDatacenters))

	for _, datacenter := range allDatacenters {
		wrapper.DatacenterMap[datacenter.ID] = datacenter
	}

	assert.False(t, wrapper.IsEmpty())

	return *wrapper
}

func TestDatabaseBinToReference(t *testing.T) {
	numBuyers := 3
	numSellers := 3
	numRelays := 3
	numDatacenters := 3

	wrapperActual := SetupReferenceWrapper(t, numBuyers, numSellers, numRelays, numDatacenters)

	wrapperExpected := SetupStorageWrapper(t, numBuyers, numSellers, numRelays, numDatacenters)

	assert.Equal(t, wrapperExpected.BuyerMap, wrapperActual.BuyerMap)
	assert.Equal(t, wrapperExpected.SellerMap, wrapperActual.SellerMap)
	assert.Equal(t, wrapperExpected.DatacenterMap, wrapperActual.DatacenterMap)
	assert.Equal(t, wrapperExpected.DatacenterMaps, wrapperActual.DatacenterMaps)
	assert.Equal(t, wrapperExpected.RelayMap, wrapperActual.RelayMap)
	assert.Equal(t, wrapperExpected.Relays, wrapperActual.Relays)

	dbReferenceActual := wrapperActual.WrapperToReference()

	for i := 0; i < numBuyers; i++ {
		buyerID := uint64(i + 1)
		assert.Equal(t, buyerID, dbReferenceActual.Buyers[i])
		assert.Equal(t, crypto.HashID(fmt.Sprintf("datacenter%d", buyerID)), dbReferenceActual.DatacenterMaps[buyerID][0])
	}

	for i := 0; i < numSellers; i++ {
		assert.Equal(t, fmt.Sprintf("seller%d", i+1), dbReferenceActual.Sellers[i])
	}

	for i := 0; i < numRelays; i++ {
		assert.Equal(t, fmt.Sprintf("relay%d", i+1), dbReferenceActual.Relays[i].DisplayName)
		assert.Equal(t, fmt.Sprintf("127.0.0.%d", i+1), dbReferenceActual.Relays[i].PublicIP.IP.String())
	}

	for i := 0; i < numDatacenters; i++ {
		assert.Equal(t, fmt.Sprintf("datacenter%d", i+1), dbReferenceActual.Datacenters[i])
	}
}

func TestReferenceSerialization(t *testing.T) {
	numBuyers := 3
	numSellers := 3
	numRelays := 3
	numDatacenters := 3

	expectedWrapper := SetupStorageWrapper(t, numBuyers, numSellers, numRelays, numDatacenters)
	dbReferenceExpected := expectedWrapper.WrapperToReference()
	expectedBuffer := make([]byte, routing.MaxDatabaseBinWrapperSize)
	expectedWriteStream, err := encoding.CreateWriteStream(expectedBuffer)
	assert.NoError(t, err)

	err = dbReferenceExpected.Serialize(expectedWriteStream)
	assert.NoError(t, err)

	expectedWriteStream.Flush()

	actualWrapper := SetupReferenceWrapper(t, numBuyers, numSellers, numRelays, numDatacenters)
	dbReferenceActual := actualWrapper.WrapperToReference()
	actualBuffer := make([]byte, routing.MaxDatabaseBinWrapperSize)
	actualWriteStream, err := encoding.CreateWriteStream(actualBuffer)
	assert.NoError(t, err)

	err = dbReferenceActual.Serialize(actualWriteStream)
	assert.NoError(t, err)

	actualWriteStream.Flush()

	assert.Equal(t, binary.LittleEndian.Uint64(expectedBuffer), binary.LittleEndian.Uint64(actualBuffer))
}

func TestReferenceHashing(t *testing.T) {
	numBuyers := 3
	numSellers := 3
	numRelays := 3
	numDatacenters := 3

	expectedWrapper := SetupStorageWrapper(t, numBuyers, numSellers, numRelays, numDatacenters)
	dbReferenceExpected := expectedWrapper.WrapperToReference()

	expectedHash, err := dbReferenceExpected.Hash()
	assert.NoError(t, err)

	actualWrapper := SetupReferenceWrapper(t, numBuyers, numSellers, numRelays, numDatacenters)
	dbReferenceActual := actualWrapper.WrapperToReference()

	actualHash, err := dbReferenceActual.Hash()
	assert.NoError(t, err)

	assert.Equal(t, expectedHash, actualHash)
}
