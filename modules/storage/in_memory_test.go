package storage_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/crypto"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/storage"
	"github.com/stretchr/testify/assert"
)

func TestInMemoryGetBuyer(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("buyer not found", func(t *testing.T) {
		inMemory := storage.InMemory{}

		actual, err := inMemory.Buyer(ctx, 0)
		assert.Empty(t, actual)
		assert.EqualError(t, err, "buyer with reference 0 not found")
	})

	t.Run("success", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Buyer{
			ID: 1,
		}

		err := inMemory.AddBuyer(ctx, expected)
		assert.NoError(t, err)

		actual, err := inMemory.Buyer(ctx, expected.ID)
		assert.NoError(t, err)

		assert.Equal(t, expected, actual)
	})
}

func TestInMemoryGetBuyers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("no buyers", func(t *testing.T) {
		inMemory := storage.InMemory{}

		buyers := inMemory.Buyers(ctx)
		assert.NotNil(t, buyers)
		assert.Len(t, buyers, 0)
	})

	t.Run("success", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Buyer{
			ID: 1,
		}

		err := inMemory.AddBuyer(ctx, expected)
		assert.NoError(t, err)

		actual := inMemory.Buyers(ctx)
		assert.Equal(t, []routing.Buyer{expected}, actual)
	})
}

func TestInMemoryAddBuyer(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("buyer already exists", func(t *testing.T) {
		inMemory := storage.InMemory{}

		buyer := routing.Buyer{
			ID: 0,
		}

		err := inMemory.AddBuyer(ctx, buyer)
		assert.NoError(t, err)

		err = inMemory.AddBuyer(ctx, buyer)
		assert.EqualError(t, err, "buyer with reference 0 already exists")
	})

	t.Run("success", func(t *testing.T) {
		inMemory := storage.InMemory{}

		buyer := routing.Buyer{
			ID: 1,
		}

		err := inMemory.AddBuyer(ctx, buyer)
		assert.NoError(t, err)

		buyer, err = inMemory.Buyer(ctx, buyer.ID)
		assert.NotEmpty(t, buyer)
		assert.NoError(t, err)
	})
}

func TestInMemoryRemoveBuyer(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("buyer doesn't exist", func(t *testing.T) {
		inMemory := storage.InMemory{}

		err := inMemory.RemoveBuyer(ctx, 0)
		assert.EqualError(t, err, "buyer with reference 0 not found")
	})

	t.Run("success removing last element", func(t *testing.T) {
		inMemory := storage.InMemory{}

		buyers := []routing.Buyer{
			{
				ID: 1,
			},
			{
				ID: 2,
			},
		}

		for i := 0; i < len(buyers); i++ {
			err := inMemory.AddBuyer(ctx, buyers[i])
			assert.NoError(t, err)
		}

		err := inMemory.RemoveBuyer(ctx, 2)
		assert.NoError(t, err)

		expected := []routing.Buyer{buyers[0]}
		actual := inMemory.Buyers(ctx)
		assert.Equal(t, expected, actual)
	})

	t.Run("success removing not last element", func(t *testing.T) {
		inMemory := storage.InMemory{}

		buyers := []routing.Buyer{
			{
				ID: 1,
			},
			{
				ID: 2,
			},
		}

		for i := 0; i < len(buyers); i++ {
			err := inMemory.AddBuyer(ctx, buyers[i])
			assert.NoError(t, err)
		}

		err := inMemory.RemoveBuyer(ctx, 1)
		assert.NoError(t, err)

		expected := []routing.Buyer{buyers[1]}
		actual := inMemory.Buyers(ctx)
		assert.Equal(t, expected, actual)
	})
}

func TestInMemorySetBuyer(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("buyer doesn't exist", func(t *testing.T) {
		inMemory := storage.InMemory{}

		buyer := routing.Buyer{
			ID: 0,
		}

		err := inMemory.SetBuyer(ctx, buyer)
		assert.EqualError(t, err, "buyer with reference 0 not found")
	})

	t.Run("success", func(t *testing.T) {
		inMemory := storage.InMemory{}

		buyer := routing.Buyer{
			ID: 1,
		}

		err := inMemory.AddBuyer(ctx, buyer)
		assert.NoError(t, err)

		err = inMemory.SetBuyer(ctx, buyer)
		assert.NoError(t, err)

		buyerInStorage, err := inMemory.Buyer(ctx, buyer.ID)
		assert.NoError(t, err)
		assert.Equal(t, buyer, buyerInStorage)
	})
}

func TestInMemoryGetSeller(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("seller not found", func(t *testing.T) {
		inMemory := storage.InMemory{}

		actual, err := inMemory.Seller(ctx, "id")
		assert.Empty(t, actual)
		assert.EqualError(t, err, "seller with reference id not found")
	})

	t.Run("success", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Seller{
			ID:   "id",
			Name: "seller name",
		}

		err := inMemory.AddSeller(ctx, expected)
		assert.NoError(t, err)

		actual, err := inMemory.Seller(ctx, expected.ID)
		assert.NoError(t, err)

		assert.Equal(t, expected, actual)
	})
}

func TestInMemoryGetSellers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("no sellers", func(t *testing.T) {
		inMemory := storage.InMemory{}

		sellers := inMemory.Sellers(ctx)
		assert.NotNil(t, sellers)
		assert.Len(t, sellers, 0)
	})

	t.Run("success", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Seller{
			ID:   "id",
			Name: "seller name",
		}

		err := inMemory.AddSeller(ctx, expected)
		assert.NoError(t, err)

		actual := inMemory.Sellers(ctx)
		assert.Equal(t, []routing.Seller{expected}, actual)
	})
}

func TestInMemoryAddSeller(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("seller already exists", func(t *testing.T) {
		inMemory := storage.InMemory{}

		seller := routing.Seller{
			ID:   "id",
			Name: "seller name",
		}

		err := inMemory.AddSeller(ctx, seller)
		assert.NoError(t, err)

		err = inMemory.AddSeller(ctx, seller)
		assert.EqualError(t, err, "seller with reference id already exists")
	})

	t.Run("success", func(t *testing.T) {
		inMemory := storage.InMemory{}

		seller := routing.Seller{
			ID:   "id",
			Name: "seller name",
		}

		err := inMemory.AddSeller(ctx, seller)
		assert.NoError(t, err)

		seller, err = inMemory.Seller(ctx, seller.ID)
		assert.NotEmpty(t, seller)
		assert.NoError(t, err)
	})
}

func TestInMemoryRemoveSeller(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("seller doesn't exist", func(t *testing.T) {
		inMemory := storage.InMemory{}

		err := inMemory.RemoveSeller(ctx, "id")
		assert.EqualError(t, err, "seller with reference id not found")
	})

	t.Run("success removing last element", func(t *testing.T) {
		inMemory := storage.InMemory{}

		sellers := []routing.Seller{
			{
				ID:   "id1",
				Name: "seller name",
			},
			{
				ID:   "id2",
				Name: "seller name",
			},
		}

		for i := 0; i < len(sellers); i++ {
			err := inMemory.AddSeller(ctx, sellers[i])
			assert.NoError(t, err)
		}

		err := inMemory.RemoveSeller(ctx, "id2")
		assert.NoError(t, err)

		expected := []routing.Seller{sellers[0]}
		actual := inMemory.Sellers(ctx)
		assert.Equal(t, expected, actual)
	})

	t.Run("success removing not last element", func(t *testing.T) {
		inMemory := storage.InMemory{}

		sellers := []routing.Seller{
			{
				ID:   "id1",
				Name: "seller name",
			},
			{
				ID:   "id2",
				Name: "seller name",
			},
		}

		for i := 0; i < len(sellers); i++ {
			err := inMemory.AddSeller(ctx, sellers[i])
			assert.NoError(t, err)
		}

		err := inMemory.RemoveSeller(ctx, "id1")
		assert.NoError(t, err)

		expected := []routing.Seller{sellers[1]}
		actual := inMemory.Sellers(ctx)
		assert.Equal(t, expected, actual)
	})
}

func TestInMemorySetSeller(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("seller doesn't exist", func(t *testing.T) {
		inMemory := storage.InMemory{}

		seller := routing.Seller{
			ID:   "id",
			Name: "seller name",
		}

		err := inMemory.SetSeller(ctx, seller)
		assert.EqualError(t, err, "seller with reference id not found")
	})

	t.Run("success", func(t *testing.T) {
		inMemory := storage.InMemory{}

		seller := routing.Seller{
			ID:   "id",
			Name: "seller name",
		}

		err := inMemory.AddSeller(ctx, seller)
		assert.NoError(t, err)

		seller.Name = "new seller name"

		err = inMemory.SetSeller(ctx, seller)
		assert.NoError(t, err)

		sellerInStorage, err := inMemory.Seller(ctx, seller.ID)
		assert.NoError(t, err)
		assert.Equal(t, seller, sellerInStorage)
	})
}

func TestInMemoryGetRelay(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("relay not found", func(t *testing.T) {
		inMemory := storage.InMemory{}

		actual, err := inMemory.Relay(ctx, 0)
		assert.Empty(t, actual)
		assert.EqualError(t, err, "relay with reference 0 not found")
	})

	t.Run("success", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Relay{
			ID:   1,
			Name: "relay name",
			Seller: routing.Seller{
				ID:   "seller ID",
				Name: "seller name",
			},
			Datacenter: routing.Datacenter{
				ID:   crypto.HashID("datacenter name"),
				Name: "datadcenter name",
			},
		}

		err := inMemory.AddSeller(ctx, expected.Seller)
		assert.NoError(t, err)

		err = inMemory.AddDatacenter(ctx, expected.Datacenter)
		assert.NoError(t, err)

		err = inMemory.AddRelay(ctx, expected)
		assert.NoError(t, err)

		actual, err := inMemory.Relay(ctx, expected.ID)
		assert.NoError(t, err)

		assert.Equal(t, expected, actual)
	})
}

func TestInMemoryGetRelays(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("no relays", func(t *testing.T) {
		inMemory := storage.InMemory{}

		relays := inMemory.Relays(ctx)
		assert.NotNil(t, relays)
		assert.Len(t, relays, 0)
	})

	t.Run("success", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Relay{
			ID:   1,
			Name: "relay name",
			Seller: routing.Seller{
				ID:   "seller ID",
				Name: "seller name",
			},
			Datacenter: routing.Datacenter{
				ID:   crypto.HashID("datacenter name"),
				Name: "datadcenter name",
			},
		}

		err := inMemory.AddSeller(ctx, expected.Seller)
		assert.NoError(t, err)

		err = inMemory.AddDatacenter(ctx, expected.Datacenter)
		assert.NoError(t, err)

		err = inMemory.AddRelay(ctx, expected)
		assert.NoError(t, err)

		actual := inMemory.Relays(ctx)
		assert.Equal(t, []routing.Relay{expected}, actual)
	})
}

func TestInMemoryAddRelay(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("relay already exists", func(t *testing.T) {
		inMemory := storage.InMemory{}

		relay := routing.Relay{
			ID:   0,
			Name: "relay name",
			Seller: routing.Seller{
				ID:   "seller ID",
				Name: "seller name",
			},
			Datacenter: routing.Datacenter{
				ID:   crypto.HashID("datacenter name"),
				Name: "datadcenter name",
			},
		}

		err := inMemory.AddSeller(ctx, relay.Seller)
		assert.NoError(t, err)

		err = inMemory.AddDatacenter(ctx, relay.Datacenter)
		assert.NoError(t, err)

		err = inMemory.AddRelay(ctx, relay)
		assert.NoError(t, err)

		err = inMemory.AddRelay(ctx, relay)
		assert.EqualError(t, err, "relay with reference 0 already exists")
	})

	t.Run("no seller", func(t *testing.T) {
		inMemory := storage.InMemory{}

		relay := routing.Relay{
			ID:   1,
			Name: "relay name",
		}

		err := inMemory.AddRelay(ctx, relay)
		assert.EqualError(t, err, "seller with reference  not found")
	})

	t.Run("no datacenter", func(t *testing.T) {
		inMemory := storage.InMemory{}

		relay := routing.Relay{
			ID:   1,
			Name: "relay name",
			Seller: routing.Seller{
				ID:   "seller ID",
				Name: "seller name",
			},
		}

		err := inMemory.AddSeller(ctx, relay.Seller)
		assert.NoError(t, err)

		err = inMemory.AddRelay(ctx, relay)
		assert.EqualError(t, err, "datacenter with reference 0 not found")
	})

	t.Run("success", func(t *testing.T) {
		inMemory := storage.InMemory{}

		relay := routing.Relay{
			ID:   1,
			Name: "relay name",
			Seller: routing.Seller{
				ID:   "seller ID",
				Name: "seller name",
			},
			Datacenter: routing.Datacenter{
				ID:   crypto.HashID("datacenter name"),
				Name: "datadcenter name",
			},
		}

		err := inMemory.AddSeller(ctx, relay.Seller)
		assert.NoError(t, err)

		err = inMemory.AddDatacenter(ctx, relay.Datacenter)
		assert.NoError(t, err)

		err = inMemory.AddRelay(ctx, relay)
		assert.NoError(t, err)

		relay, err = inMemory.Relay(ctx, relay.ID)
		assert.NotEmpty(t, relay)
		assert.NoError(t, err)
	})
}

func TestInMemoryRemoveRelay(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("relay doesn't exist", func(t *testing.T) {
		inMemory := storage.InMemory{}

		err := inMemory.RemoveRelay(ctx, 0)
		assert.EqualError(t, err, "relay with reference 0 not found")
	})

	t.Run("success removing last element", func(t *testing.T) {
		inMemory := storage.InMemory{}

		relays := []routing.Relay{
			{
				ID:   1,
				Name: "relay name",
				Seller: routing.Seller{
					ID:   "seller ID",
					Name: "seller name",
				},
				Datacenter: routing.Datacenter{
					ID:   crypto.HashID("datacenter name"),
					Name: "datadcenter name",
				},
			},
			{
				ID:   2,
				Name: "relay name",
				Seller: routing.Seller{
					ID:   "seller ID",
					Name: "seller name",
				},
				Datacenter: routing.Datacenter{
					ID:   crypto.HashID("datacenter name"),
					Name: "datadcenter name",
				},
			},
		}

		err := inMemory.AddSeller(ctx, relays[0].Seller)
		assert.NoError(t, err)

		err = inMemory.AddDatacenter(ctx, relays[0].Datacenter)
		assert.NoError(t, err)

		for i := 0; i < len(relays); i++ {
			err := inMemory.AddRelay(ctx, relays[i])
			assert.NoError(t, err)
		}

		err = inMemory.RemoveRelay(ctx, 2)
		assert.NoError(t, err)

		expected := []routing.Relay{relays[0]}
		actual := inMemory.Relays(ctx)
		assert.Equal(t, expected, actual)
	})

	t.Run("success removing not last element", func(t *testing.T) {
		inMemory := storage.InMemory{}

		relays := []routing.Relay{
			{
				ID:   1,
				Name: "relay name",
				Seller: routing.Seller{
					ID:   "seller ID",
					Name: "seller name",
				},
				Datacenter: routing.Datacenter{
					ID:   crypto.HashID("datacenter name"),
					Name: "datadcenter name",
				},
			},
			{
				ID:   2,
				Name: "relay name",
				Seller: routing.Seller{
					ID:   "seller ID",
					Name: "seller name",
				},
				Datacenter: routing.Datacenter{
					ID:   crypto.HashID("datacenter name"),
					Name: "datadcenter name",
				},
			},
		}

		err := inMemory.AddSeller(ctx, relays[0].Seller)
		assert.NoError(t, err)

		err = inMemory.AddDatacenter(ctx, relays[0].Datacenter)
		assert.NoError(t, err)

		for i := 0; i < len(relays); i++ {
			err := inMemory.AddRelay(ctx, relays[i])
			assert.NoError(t, err)
		}

		err = inMemory.RemoveRelay(ctx, 1)
		assert.NoError(t, err)

		expected := []routing.Relay{relays[1]}
		actual := inMemory.Relays(ctx)
		assert.Equal(t, expected, actual)
	})
}

func TestInMemorySetRelay(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("relay doesn't exist", func(t *testing.T) {
		inMemory := storage.InMemory{}

		relay := routing.Relay{
			ID:   0,
			Name: "relay name",
		}

		err := inMemory.SetRelay(ctx, relay)
		assert.EqualError(t, err, "relay with reference 0 not found")
	})

	t.Run("success", func(t *testing.T) {
		inMemory := storage.InMemory{}

		relay := routing.Relay{
			ID:   1,
			Name: "relay name",
			Seller: routing.Seller{
				ID:   "seller ID",
				Name: "seller name",
			},
			Datacenter: routing.Datacenter{
				ID:   crypto.HashID("datacenter name"),
				Name: "datadcenter name",
			},
		}

		err := inMemory.AddSeller(ctx, relay.Seller)
		assert.NoError(t, err)

		err = inMemory.AddDatacenter(ctx, relay.Datacenter)
		assert.NoError(t, err)

		err = inMemory.AddRelay(ctx, relay)
		assert.NoError(t, err)

		relay.Name = "new relay name"

		err = inMemory.SetRelay(ctx, relay)
		assert.NoError(t, err)

		relayInStorage, err := inMemory.Relay(ctx, relay.ID)
		assert.NoError(t, err)
		assert.Equal(t, relay, relayInStorage)
	})
}

func TestInMemoryGetDatacenter(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("datacenter not found", func(t *testing.T) {
		inMemory := storage.InMemory{}

		actual, err := inMemory.Datacenter(ctx, 0)
		assert.Empty(t, actual)
		assert.EqualError(t, err, "datacenter with reference 0 not found")
	})

	t.Run("success", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Datacenter{
			ID:   1,
			Name: "datacenter name",
		}

		err := inMemory.AddDatacenter(ctx, expected)
		assert.NoError(t, err)

		actual, err := inMemory.Datacenter(ctx, expected.ID)
		assert.NoError(t, err)

		assert.Equal(t, expected, actual)
	})
}

func TestInMemoryGetDatacenters(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("no datacenters", func(t *testing.T) {
		inMemory := storage.InMemory{}

		datacenters := inMemory.Datacenters(ctx)
		assert.NotNil(t, datacenters)
		assert.Len(t, datacenters, 0)
	})

	t.Run("success", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Datacenter{
			ID:   1,
			Name: "datacenter name",
		}

		err := inMemory.AddDatacenter(ctx, expected)
		assert.NoError(t, err)

		actual := inMemory.Datacenters(ctx)
		assert.Equal(t, []routing.Datacenter{expected}, actual)
	})
}

func TestInMemoryAddDatacenter(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("datacenter already exists", func(t *testing.T) {
		inMemory := storage.InMemory{}

		datacenter := routing.Datacenter{
			ID:   0,
			Name: "datacenter name",
		}

		err := inMemory.AddDatacenter(ctx, datacenter)
		assert.NoError(t, err)

		err = inMemory.AddDatacenter(ctx, datacenter)
		assert.EqualError(t, err, "datacenter with reference 0 already exists")
	})

	t.Run("success", func(t *testing.T) {
		inMemory := storage.InMemory{}

		datacenter := routing.Datacenter{
			ID:   1,
			Name: "datacenter name",
		}

		err := inMemory.AddDatacenter(ctx, datacenter)
		assert.NoError(t, err)

		datacenter, err = inMemory.Datacenter(ctx, datacenter.ID)
		assert.NotEmpty(t, datacenter)
		assert.NoError(t, err)
	})
}

func TestInMemoryRemoveDatacenter(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("datacenter doesn't exist", func(t *testing.T) {
		inMemory := storage.InMemory{}

		err := inMemory.RemoveDatacenter(ctx, 0)
		assert.EqualError(t, err, "datacenter with reference 0 not found")
	})

	t.Run("success removing last element", func(t *testing.T) {
		inMemory := storage.InMemory{}

		datacenters := []routing.Datacenter{
			{
				ID:   1,
				Name: "datacenter name",
			},
			{
				ID:   2,
				Name: "datacenter name",
			},
		}

		for i := 0; i < len(datacenters); i++ {
			err := inMemory.AddDatacenter(ctx, datacenters[i])
			assert.NoError(t, err)
		}

		err := inMemory.RemoveDatacenter(ctx, 2)
		assert.NoError(t, err)

		expected := []routing.Datacenter{datacenters[0]}
		actual := inMemory.Datacenters(ctx)
		assert.Equal(t, expected, actual)
	})

	t.Run("success removing not last element", func(t *testing.T) {
		inMemory := storage.InMemory{}

		datacenters := []routing.Datacenter{
			{
				ID:   1,
				Name: "datacenter name",
			},
			{
				ID:   2,
				Name: "datacenter name",
			},
		}

		for i := 0; i < len(datacenters); i++ {
			err := inMemory.AddDatacenter(ctx, datacenters[i])
			assert.NoError(t, err)
		}

		err := inMemory.RemoveDatacenter(ctx, 1)
		assert.NoError(t, err)

		expected := []routing.Datacenter{datacenters[1]}
		actual := inMemory.Datacenters(ctx)
		assert.Equal(t, expected, actual)
	})
}

func TestInMemorySetDatacenter(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("datacenter doesn't exist", func(t *testing.T) {
		inMemory := storage.InMemory{}

		datacenter := routing.Datacenter{
			ID:   0,
			Name: "datacenter name",
		}

		err := inMemory.SetDatacenter(ctx, datacenter)
		assert.EqualError(t, err, "datacenter with reference 0 not found")
	})

	t.Run("success", func(t *testing.T) {
		inMemory := storage.InMemory{}

		datacenter := routing.Datacenter{
			ID:   1,
			Name: "datacenter name",
		}

		err := inMemory.AddDatacenter(ctx, datacenter)
		assert.NoError(t, err)

		datacenter.Name = "new datacenter name"

		err = inMemory.SetDatacenter(ctx, datacenter)
		assert.NoError(t, err)

		datacenterInStorage, err := inMemory.Datacenter(ctx, datacenter.ID)
		assert.NoError(t, err)
		assert.Equal(t, datacenter, datacenterInStorage)
	})
}

func TestInMemoryInternalConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("buyer does not exist", func(t *testing.T) {
		inMemory := storage.InMemory{}

		actual, err := inMemory.InternalConfig(ctx, 0)
		assert.Equal(t, core.InternalConfig{}, actual)
		assert.EqualError(t, err, "buyer with reference 0 not found")
	})

	t.Run("buyer does not have internal config", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Buyer{
			ID: 1,
		}

		err := inMemory.AddBuyer(ctx, expected)
		assert.NoError(t, err)

		actual, err := inMemory.InternalConfig(ctx, expected.ID)
		assert.Equal(t, core.InternalConfig{}, actual)
		assert.EqualError(t, fmt.Errorf("InternalConfig with reference %016x not found", expected.ID), err.Error())
	})

	t.Run("success", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Buyer{
			ID:             1,
			InternalConfig: core.NewInternalConfig(),
		}

		err := inMemory.AddBuyer(ctx, expected)
		assert.NoError(t, err)

		actual, err := inMemory.InternalConfig(ctx, expected.ID)
		assert.NoError(t, err)
		assert.Equal(t, core.NewInternalConfig(), actual)
	})
}

func TestInMemoryAddInternalConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("buyer does not exist", func(t *testing.T) {
		inMemory := storage.InMemory{}

		err := inMemory.AddInternalConfig(ctx, core.NewInternalConfig(), 0)
		assert.EqualError(t, err, "buyer with reference 0 not found")
	})

	t.Run("success", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Buyer{
			ID:             1,
			InternalConfig: core.NewInternalConfig(),
		}

		err := inMemory.AddBuyer(ctx, expected)
		assert.NoError(t, err)

		newConfig := core.NewInternalConfig()
		newConfig.RouteDiversity = 3

		err = inMemory.AddInternalConfig(ctx, newConfig, expected.ID)
		assert.NoError(t, err)

		actual, err := inMemory.InternalConfig(ctx, expected.ID)
		assert.NoError(t, err)
		assert.Equal(t, newConfig, actual)
	})
}
