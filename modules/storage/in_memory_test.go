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

func TestInMemoryGetCustomer(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("customer not found by code", func(t *testing.T) {
		inMemory := storage.InMemory{}

		actual, err := inMemory.Customer(ctx, "not found")
		assert.Empty(t, actual)
		assert.EqualError(t, err, "customer with reference not found not found")
	})

	t.Run("success by code", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Customer{
			Code: "found",
		}

		err := inMemory.AddCustomer(ctx, expected)
		assert.NoError(t, err)

		actual, err := inMemory.Customer(ctx, expected.Code)
		assert.NoError(t, err)

		assert.Equal(t, expected, actual)
	})

	t.Run("customer not found by id", func(t *testing.T) {
		inMemory := storage.InMemory{}

		actual, err := inMemory.CustomerByID(ctx, 0)
		assert.Empty(t, actual)
		assert.EqualError(t, err, "customer with reference 0 not found")
	})

	t.Run("success by id", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Customer{
			DatabaseID: 0,
		}

		err := inMemory.AddCustomer(ctx, expected)
		assert.NoError(t, err)

		actual, err := inMemory.Customer(ctx, expected.Code)
		assert.NoError(t, err)

		assert.Equal(t, expected, actual)
	})
}

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

func TestInMemoryUpdateInternalConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	int32Fields := []string{"RouteSelectThreshold", "RouteSwitchThreshold", "MaxLatencyTradeOff",
		"RTTVeto_Default", "RTTVeto_PacketLoss", "RTTVeto_Multipath",
		"MultipathOverloadThreshold", "MaxRTT", "RouteDiversity", "MultipathThreshold",
		"ReducePacketLossMinSliceNumber"}

	boolFields := []string{"TryBeforeYouBuy", "ForceNext", "LargeCustomer", "Uncommitted",
		"HighFrequencyPings", "EnableVanityMetrics"}

	t.Run("buyer does not exist", func(t *testing.T) {
		inMemory := storage.InMemory{}

		err := inMemory.UpdateInternalConfig(ctx, 0, "", "")
		assert.EqualError(t, err, "buyer with reference 0 not found")
	})

	t.Run("failed int32 fields", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Buyer{
			ID:             1,
			InternalConfig: core.NewInternalConfig(),
		}

		err := inMemory.AddBuyer(ctx, expected)
		assert.NoError(t, err)

		for _, field := range int32Fields {
			err := inMemory.UpdateInternalConfig(ctx, expected.ID, field, float64(-1))
			assert.EqualError(t, fmt.Errorf("%s: %v is not a valid int32 type (%T)", field, float64(-1), float64(-1)), err.Error())
		}
	})

	t.Run("failed bool fields", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Buyer{
			ID:             1,
			InternalConfig: core.NewInternalConfig(),
		}

		err := inMemory.AddBuyer(ctx, expected)
		assert.NoError(t, err)

		for _, field := range boolFields {
			err := inMemory.UpdateInternalConfig(ctx, expected.ID, field, float64(-1))
			assert.EqualError(t, fmt.Errorf("%s: %v is not a valid boolean type (%T)", field, float64(-1), float64(-1)), err.Error())
		}
	})

	t.Run("unknown field", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Buyer{
			ID:             1,
			InternalConfig: core.NewInternalConfig(),
		}

		err := inMemory.AddBuyer(ctx, expected)
		assert.NoError(t, err)

		err = inMemory.UpdateInternalConfig(ctx, expected.ID, "unknown", float64(-1))
		assert.EqualError(t, fmt.Errorf("Field '%v' does not exist on the InternalConfig type", "unknown"), err.Error())
	})

	t.Run("success int32 fields", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Buyer{
			ID:             1,
			InternalConfig: core.NewInternalConfig(),
		}

		err := inMemory.AddBuyer(ctx, expected)
		assert.NoError(t, err)

		for _, field := range int32Fields {
			err := inMemory.UpdateInternalConfig(ctx, expected.ID, field, int32(1))
			assert.NoError(t, err)
		}
	})

	t.Run("success bool fields", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Buyer{
			ID:             1,
			InternalConfig: core.NewInternalConfig(),
		}

		err := inMemory.AddBuyer(ctx, expected)
		assert.NoError(t, err)

		for _, field := range boolFields {
			err := inMemory.UpdateInternalConfig(ctx, expected.ID, field, true)
			assert.NoError(t, err)
		}
	})
}

func TestInMemoryRemoveInternalConfig(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("buyer does not exist", func(t *testing.T) {
		inMemory := storage.InMemory{}

		err := inMemory.RemoveInternalConfig(ctx, 0)
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

		ic, err := inMemory.InternalConfig(ctx, expected.ID)
		assert.NoError(t, err)
		assert.Equal(t, core.NewInternalConfig(), ic)

		err = inMemory.RemoveInternalConfig(ctx, expected.ID)
		assert.NoError(t, err)

		ic, err = inMemory.InternalConfig(ctx, expected.ID)
		assert.Error(t, err)
		assert.Equal(t, core.InternalConfig{}, ic)
	})
}

func TestInMemoryRouteShader(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("buyer does not exist", func(t *testing.T) {
		inMemory := storage.InMemory{}

		actual, err := inMemory.RouteShader(ctx, 0)
		assert.Equal(t, core.RouteShader{}, actual)
		assert.EqualError(t, err, "buyer with reference 0 not found")
	})

	t.Run("buyer does not have route shader", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Buyer{
			ID: 1,
		}

		err := inMemory.AddBuyer(ctx, expected)
		assert.NoError(t, err)

		actual, err := inMemory.RouteShader(ctx, expected.ID)
		assert.Equal(t, core.RouteShader{}, actual)
		assert.EqualError(t, fmt.Errorf("RouteShader with reference %016x not found", expected.ID), err.Error())
	})

	t.Run("success", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Buyer{
			ID:          1,
			RouteShader: core.NewRouteShader(),
		}

		err := inMemory.AddBuyer(ctx, expected)
		assert.NoError(t, err)

		actual, err := inMemory.RouteShader(ctx, expected.ID)
		assert.NoError(t, err)
		assert.Equal(t, core.NewRouteShader(), actual)
	})
}

func TestInMemoryAddRouteShader(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("buyer does not exist", func(t *testing.T) {
		inMemory := storage.InMemory{}

		err := inMemory.AddRouteShader(ctx, core.NewRouteShader(), 0)
		assert.EqualError(t, err, "buyer with reference 0 not found")
	})

	t.Run("success", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Buyer{
			ID:          1,
			RouteShader: core.NewRouteShader(),
		}

		err := inMemory.AddBuyer(ctx, expected)
		assert.NoError(t, err)

		newShader := core.NewRouteShader()
		newShader.SelectionPercent = 50

		err = inMemory.AddRouteShader(ctx, newShader, expected.ID)
		assert.NoError(t, err)

		actual, err := inMemory.RouteShader(ctx, expected.ID)
		assert.NoError(t, err)
		assert.Equal(t, newShader, actual)
	})
}

func TestInMemoryUpdateRouteShader(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	intFields := []string{"SelectionPercent"}

	int32Fields := []string{"AcceptableLatency", "LatencyThreshold", "BandwidthEnvelopeUpKbps",
		"BandwidthEnvelopeDownKbps"}

	boolFields := []string{"DisableNetworkNext", "ABTest", "ProMode", "ReduceLatency",
		"ReduceJitter", "ReducePacketLoss", "Multipath"}

	float32Fields := []string{"AcceptablePacketLoss", "PacketLossSustained"}

	t.Run("buyer does not exist", func(t *testing.T) {
		inMemory := storage.InMemory{}

		err := inMemory.UpdateRouteShader(ctx, 0, "", "")
		assert.EqualError(t, err, "buyer with reference 0 not found")
	})

	t.Run("failed int fields", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Buyer{
			ID:          1,
			RouteShader: core.NewRouteShader(),
		}

		err := inMemory.AddBuyer(ctx, expected)
		assert.NoError(t, err)

		for _, field := range intFields {
			err := inMemory.UpdateRouteShader(ctx, expected.ID, field, float64(-1))
			assert.EqualError(t, fmt.Errorf("%s: %v is not a valid int type (%T)", field, float64(-1), float64(-1)), err.Error())
		}
	})

	t.Run("failed int32 fields", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Buyer{
			ID:          1,
			RouteShader: core.NewRouteShader(),
		}

		err := inMemory.AddBuyer(ctx, expected)
		assert.NoError(t, err)

		for _, field := range int32Fields {
			err := inMemory.UpdateRouteShader(ctx, expected.ID, field, float64(-1))
			assert.EqualError(t, fmt.Errorf("%s: %v is not a valid int32 type (%T)", field, float64(-1), float64(-1)), err.Error())
		}
	})

	t.Run("failed bool fields", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Buyer{
			ID:          1,
			RouteShader: core.NewRouteShader(),
		}

		err := inMemory.AddBuyer(ctx, expected)
		assert.NoError(t, err)

		for _, field := range boolFields {
			err := inMemory.UpdateRouteShader(ctx, expected.ID, field, float64(-1))
			assert.EqualError(t, fmt.Errorf("%s: %v is not a valid boolean type (%T)", field, float64(-1), float64(-1)), err.Error())
		}
	})

	t.Run("failed float32 fields", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Buyer{
			ID:          1,
			RouteShader: core.NewRouteShader(),
		}

		err := inMemory.AddBuyer(ctx, expected)
		assert.NoError(t, err)

		for _, field := range float32Fields {
			err := inMemory.UpdateRouteShader(ctx, expected.ID, field, int(-1))
			assert.EqualError(t, fmt.Errorf("%s: %v is not a valid float32 type (%T)", field, int(-1), int(-1)), err.Error())
		}
	})

	t.Run("unknown field", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Buyer{
			ID:          1,
			RouteShader: core.NewRouteShader(),
		}

		err := inMemory.AddBuyer(ctx, expected)
		assert.NoError(t, err)

		err = inMemory.UpdateRouteShader(ctx, expected.ID, "unknown", float64(-1))
		assert.EqualError(t, fmt.Errorf("Field '%v' does not exist on the RouteShader type", "unknown"), err.Error())
	})

	t.Run("success int fields", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Buyer{
			ID:          1,
			RouteShader: core.NewRouteShader(),
		}

		err := inMemory.AddBuyer(ctx, expected)
		assert.NoError(t, err)

		for _, field := range intFields {
			err := inMemory.UpdateRouteShader(ctx, expected.ID, field, int(1))
			assert.NoError(t, err)
		}
	})

	t.Run("success int32 fields", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Buyer{
			ID:          1,
			RouteShader: core.NewRouteShader(),
		}

		err := inMemory.AddBuyer(ctx, expected)
		assert.NoError(t, err)

		for _, field := range int32Fields {
			err := inMemory.UpdateRouteShader(ctx, expected.ID, field, int32(1))
			assert.NoError(t, err)
		}
	})

	t.Run("success bool fields", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Buyer{
			ID:          1,
			RouteShader: core.NewRouteShader(),
		}

		err := inMemory.AddBuyer(ctx, expected)
		assert.NoError(t, err)

		for _, field := range boolFields {
			err := inMemory.UpdateRouteShader(ctx, expected.ID, field, true)
			assert.NoError(t, err)
		}
	})

	t.Run("success float32 fields", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Buyer{
			ID:          1,
			RouteShader: core.NewRouteShader(),
		}

		err := inMemory.AddBuyer(ctx, expected)
		assert.NoError(t, err)

		for _, field := range float32Fields {
			err := inMemory.UpdateRouteShader(ctx, expected.ID, field, float32(1))
			assert.NoError(t, err)
		}
	})
}

func TestInMemoryRemoveRouteShader(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("buyer does not exist", func(t *testing.T) {
		inMemory := storage.InMemory{}

		err := inMemory.RemoveRouteShader(ctx, 0)
		assert.EqualError(t, err, "buyer with reference 0 not found")
	})

	t.Run("success", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Buyer{
			ID:          1,
			RouteShader: core.NewRouteShader(),
		}

		err := inMemory.AddBuyer(ctx, expected)
		assert.NoError(t, err)

		rs, err := inMemory.RouteShader(ctx, expected.ID)
		assert.NoError(t, err)
		assert.Equal(t, core.NewRouteShader(), rs)

		err = inMemory.RemoveRouteShader(ctx, expected.ID)
		assert.NoError(t, err)

		rs, err = inMemory.RouteShader(ctx, expected.ID)
		assert.Error(t, err)
		assert.Equal(t, core.RouteShader{}, rs)
	})
}

func TestInMemoryUpdateRelay(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("relay does not exist", func(t *testing.T) {
		inMemory := storage.InMemory{}

		err := inMemory.UpdateRelay(ctx, 0, "", "")
		assert.EqualError(t, err, "relay with reference 0 not found")
	})

	float64Fields := []string{"NICSpeedMbps", "IncludedBandwidthGB", "MaxBandwidthMbps", "ContractTerm", "SSHPort", "MaxSessions"}
	stringFields := []string{"ManagementAddr", "SSHUser", "Version"}
	timeFields := []string{"StartDate", "EndDate"}
	addressFields := []string{"Addr", "InternalAddr"}
	nibblinFields := []string{"EgressPriceOverride", "MRC", "Overage"}

	// special cases: PublicKey, State, BWRule, Type, BillingSupplier

	relay := routing.Relay{
		ID:   0,
		Name: "relay name",
		Seller: routing.Seller{
			ID:         "seller ID",
			Name:       "seller name",
			DatabaseID: 1,
		},
		Datacenter: routing.Datacenter{
			ID:   crypto.HashID("datacenter name"),
			Name: "datadcenter name",
		},
	}

	inMemory := storage.InMemory{}

	err := inMemory.AddSeller(ctx, relay.Seller)
	assert.NoError(t, err)

	err = inMemory.AddDatacenter(ctx, relay.Datacenter)
	assert.NoError(t, err)

	err = inMemory.AddRelay(ctx, relay)
	assert.NoError(t, err)

	t.Run("unknown field", func(t *testing.T) {
		err := inMemory.UpdateRelay(ctx, 0, "unknown", "1")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "does not exist on the routing.Relay type")
	})

	t.Run("invalid float64 fields", func(t *testing.T) {
		for _, field := range float64Fields {
			err := inMemory.UpdateRelay(ctx, 0, field, "a")
			assert.Error(t, err)
			assert.EqualError(t, err, fmt.Sprintf("%s is not a valid float64 type", "a"))
		}
	})

	t.Run("invalid string fields", func(t *testing.T) {
		for _, field := range stringFields {
			err := inMemory.UpdateRelay(ctx, 0, field, float64(-1))
			assert.Error(t, err)
			assert.EqualError(t, err, fmt.Sprintf("%v is not a valid string value", float64(-1)))
		}
	})

	t.Run("invalid time fields", func(t *testing.T) {
		for _, field := range timeFields {
			err := inMemory.UpdateRelay(ctx, 0, field, float64(-1))
			assert.Error(t, err)
			assert.EqualError(t, err, fmt.Sprintf("%v is not a valid string value", float64(-1)))
		}

		for _, field := range timeFields {
			err := inMemory.UpdateRelay(ctx, 0, field, "2021/11/17")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "must be of the form 'January 2, 2006'")
		}
	})

	t.Run("invalid address fields", func(t *testing.T) {
		for _, field := range addressFields {
			err := inMemory.UpdateRelay(ctx, 0, field, float64(-1))
			assert.Error(t, err)
			assert.EqualError(t, err, fmt.Sprintf("%v is not a valid string value", float64(-1)))
		}

		for _, field := range addressFields {
			err := inMemory.UpdateRelay(ctx, 0, field, "127.0.0.1.1")
			assert.Error(t, err)
			assert.Contains(t, err.Error(), "unable to parse address")
		}
	})

	t.Run("invalid nibblin fields", func(t *testing.T) {
		for _, field := range nibblinFields {
			err := inMemory.UpdateRelay(ctx, 0, field, "a")
			assert.Error(t, err)
			assert.EqualError(t, err, fmt.Sprintf("%s is not a valid float64 type", "a"))
		}
	})

	t.Run("invalid public key", func(t *testing.T) {

		err := inMemory.UpdateRelay(ctx, 0, "PublicKey", float64(-1))
		assert.Error(t, err)
		assert.EqualError(t, err, fmt.Sprintf("%v is not a valid string type", float64(-1)))

		err = inMemory.UpdateRelay(ctx, 0, "PublicKey", "a")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "PublicKey: failed to encode string public key")
	})

	t.Run("invalid relay state", func(t *testing.T) {

		err := inMemory.UpdateRelay(ctx, 0, "State", "a")
		assert.Error(t, err)
		assert.EqualError(t, err, fmt.Sprintf("%s is not a valid float64 type", "a"))

		err = inMemory.UpdateRelay(ctx, 0, "State", float64(-1))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "-1 is not a valid RelayState value")

		err = inMemory.UpdateRelay(ctx, 0, "State", float64(6))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "6 is not a valid RelayState value")
	})

	t.Run("invalid bw rule", func(t *testing.T) {

		err := inMemory.UpdateRelay(ctx, 0, "BWRule", "a")
		assert.Error(t, err)
		assert.EqualError(t, err, fmt.Sprintf("%s is not a valid float64 type", "a"))

		err = inMemory.UpdateRelay(ctx, 0, "BWRule", float64(-1))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "-1 is not a valid BandWidthRule value")

		err = inMemory.UpdateRelay(ctx, 0, "BWRule", float64(5))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "5 is not a valid BandWidthRule value")
	})

	t.Run("invalid machine type", func(t *testing.T) {

		err := inMemory.UpdateRelay(ctx, 0, "Type", "a")
		assert.Error(t, err)
		assert.EqualError(t, err, fmt.Sprintf("%s is not a valid float64 type", "a"))

		err = inMemory.UpdateRelay(ctx, 0, "Type", float64(-1))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "-1 is not a valid MachineType value")

		err = inMemory.UpdateRelay(ctx, 0, "Type", float64(3))
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "3 is not a valid MachineType value")
	})

	t.Run("invalid billing supplier", func(t *testing.T) {

		err := inMemory.UpdateRelay(ctx, 0, "BillingSupplier", float64(-1))
		assert.Error(t, err)
		assert.EqualError(t, err, fmt.Sprintf("%v is not a valid string value", float64(-1)))

		err = inMemory.UpdateRelay(ctx, 0, "BillingSupplier", "unknown seller")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), fmt.Sprintf("%s is not a valid seller ID", "unknown seller"))
	})

	t.Run("success float64 fields", func(t *testing.T) {
		for _, field := range float64Fields {
			err := inMemory.UpdateRelay(ctx, 0, field, float64(1))
			assert.NoError(t, err)
		}
	})

	t.Run("success string fields", func(t *testing.T) {
		for _, field := range stringFields {
			err := inMemory.UpdateRelay(ctx, 0, field, "a")
			assert.NoError(t, err)
		}
	})

	t.Run("success time fields", func(t *testing.T) {
		for _, field := range timeFields {
			err := inMemory.UpdateRelay(ctx, 0, field, "November 17, 2021")
			assert.NoError(t, err)
		}
	})

	t.Run("success address fields", func(t *testing.T) {
		for _, field := range addressFields {
			err := inMemory.UpdateRelay(ctx, 0, field, "127.0.0.1:40000")
			assert.NoError(t, err)
		}
	})

	t.Run("success nibblin fields", func(t *testing.T) {
		for _, field := range nibblinFields {
			err := inMemory.UpdateRelay(ctx, 0, field, float64(100))
			assert.NoError(t, err)
		}
	})

	t.Run("success public key", func(t *testing.T) {
		err := inMemory.UpdateRelay(ctx, 0, "PublicKey", "YFWQjOJfHfOqsCMM/1pd+c5haMhsrE2Gm05bVUQhCnG7YlPUrI/d1g==")
		assert.NoError(t, err)
	})

	t.Run("success relay state", func(t *testing.T) {
		err := inMemory.UpdateRelay(ctx, 0, "State", float64(1))
		assert.NoError(t, err)
	})

	t.Run("success bw rule", func(t *testing.T) {

		err := inMemory.UpdateRelay(ctx, 0, "BWRule", float64(1))
		assert.NoError(t, err)
	})

	t.Run("success machine type", func(t *testing.T) {

		err := inMemory.UpdateRelay(ctx, 0, "Type", float64(1))
		assert.NoError(t, err)
	})

	t.Run("success billing supplier", func(t *testing.T) {

		err := inMemory.UpdateRelay(ctx, 0, "BillingSupplier", "seller ID")
		assert.NoError(t, err)
	})
}

func TestInMemoryAddBannedUser(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("buyer does not exist", func(t *testing.T) {
		inMemory := storage.InMemory{}

		err := inMemory.AddBannedUser(ctx, 0, 0)
		assert.EqualError(t, err, "buyer with reference 0 not found")
	})

	t.Run("buyer does not have route shader", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Buyer{
			ID: 1,
		}

		err := inMemory.AddBuyer(ctx, expected)
		assert.NoError(t, err)

		err = inMemory.AddBannedUser(ctx, 1, 0)
		assert.EqualError(t, err, fmt.Sprintf("%s with reference %016x not found", "RouteShader", 1))
	})

	t.Run("success", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Buyer{
			ID:          1,
			RouteShader: core.NewRouteShader(),
		}

		err := inMemory.AddBuyer(ctx, expected)
		assert.NoError(t, err)

		err = inMemory.AddBannedUser(ctx, 1, 0)
		assert.NoError(t, err)
	})
}

func TestInMemoryRemoveBannedUser(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("buyer does not exist", func(t *testing.T) {
		inMemory := storage.InMemory{}

		err := inMemory.RemoveBannedUser(ctx, 0, 0)
		assert.EqualError(t, err, "buyer with reference 0 not found")
	})

	t.Run("buyer does not have route shader", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Buyer{
			ID: 1,
		}

		err := inMemory.AddBuyer(ctx, expected)
		assert.NoError(t, err)

		err = inMemory.RemoveBannedUser(ctx, 1, 0)
		assert.EqualError(t, err, fmt.Sprintf("%s with reference %016x not found", "RouteShader", 1))
	})

	t.Run("success if user does not exist", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Buyer{
			ID:          1,
			RouteShader: core.NewRouteShader(),
		}

		err := inMemory.AddBuyer(ctx, expected)
		assert.NoError(t, err)

		routeShader, err := inMemory.RouteShader(ctx, expected.ID)
		assert.NoError(t, err)
		assert.Zero(t, len(routeShader.BannedUsers))

		err = inMemory.RemoveBannedUser(ctx, 1, 0)
		assert.NoError(t, err)
	})

	t.Run("success", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Buyer{
			ID:          1,
			RouteShader: core.NewRouteShader(),
		}

		err := inMemory.AddBuyer(ctx, expected)
		assert.NoError(t, err)

		err = inMemory.AddBannedUser(ctx, 1, 0)
		assert.NoError(t, err)

		routeShader, err := inMemory.RouteShader(ctx, expected.ID)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(routeShader.BannedUsers))

		err = inMemory.RemoveBannedUser(ctx, 1, 0)
		assert.NoError(t, err)

		routeShader, err = inMemory.RouteShader(ctx, expected.ID)
		assert.NoError(t, err)
		assert.Zero(t, len(routeShader.BannedUsers))
	})
}

func TestInMemoryBannedUsers(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	t.Run("buyer does not exist", func(t *testing.T) {
		inMemory := storage.InMemory{}

		_, err := inMemory.BannedUsers(ctx, 0)
		assert.EqualError(t, err, "buyer with reference 0 not found")
	})

	t.Run("buyer does not have route shader", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Buyer{
			ID: 1,
		}

		err := inMemory.AddBuyer(ctx, expected)
		assert.NoError(t, err)

		_, err = inMemory.BannedUsers(ctx, 1)
		assert.EqualError(t, err, fmt.Sprintf("%s with reference %016x not found", "RouteShader", 1))
	})

	t.Run("success", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Buyer{
			ID:          1,
			RouteShader: core.NewRouteShader(),
		}

		err := inMemory.AddBuyer(ctx, expected)
		assert.NoError(t, err)

		bannedUsers, err := inMemory.BannedUsers(ctx, 1)
		assert.NoError(t, err)
		assert.Zero(t, len(bannedUsers))

		err = inMemory.AddBannedUser(ctx, 1, 0)
		assert.NoError(t, err)

		bannedUsers, err = inMemory.BannedUsers(ctx, 1)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(bannedUsers))

		err = inMemory.RemoveBannedUser(ctx, 1, 0)
		assert.NoError(t, err)

		bannedUsers, err = inMemory.BannedUsers(ctx, 1)
		assert.NoError(t, err)
		assert.Zero(t, len(bannedUsers))
	})
}

func TestInMemoryUpdateBuyer(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	stringFields := []string{"ShortName", "PublicKey"}

	boolFields := []string{"Live", "Debug", "Analytics", "Billing", "Trial"}

	float64Fields := []string{"ExoticLocationFee", "StandardLocationFee"}

	int64Fields := []string{"LookerSeats"}

	t.Run("buyer does not exist", func(t *testing.T) {
		inMemory := storage.InMemory{}

		err := inMemory.UpdateBuyer(ctx, 0, "", "")
		assert.EqualError(t, err, "buyer with reference 0 not found")
	})

	t.Run("failed string fields", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Buyer{
			ID: 1,
		}

		err := inMemory.AddBuyer(ctx, expected)
		assert.NoError(t, err)

		for _, field := range stringFields {
			err := inMemory.UpdateBuyer(ctx, expected.ID, field, float64(-1))
			assert.EqualError(t, fmt.Errorf("%s: %v is not a valid string type (%T)", field, float64(-1), float64(-1)), err.Error())
		}
	})

	t.Run("failed bool fields", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Buyer{
			ID: 1,
		}

		err := inMemory.AddBuyer(ctx, expected)
		assert.NoError(t, err)

		for _, field := range boolFields {
			err := inMemory.UpdateBuyer(ctx, expected.ID, field, float64(-1))
			assert.EqualError(t, fmt.Errorf("%s: %v is not a valid boolean type (%T)", field, float64(-1), float64(-1)), err.Error())
		}
	})

	t.Run("failed float64 fields", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Buyer{
			ID: 1,
		}

		err := inMemory.AddBuyer(ctx, expected)
		assert.NoError(t, err)

		for _, field := range float64Fields {
			err := inMemory.UpdateBuyer(ctx, expected.ID, field, "a")
			assert.EqualError(t, fmt.Errorf("%s: %v is not a valid float64 type (%T)", field, "a", "a"), err.Error())
		}
	})

	t.Run("failed int64 fields", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Buyer{
			ID: 1,
		}

		err := inMemory.AddBuyer(ctx, expected)
		assert.NoError(t, err)

		for _, field := range int64Fields {
			err := inMemory.UpdateBuyer(ctx, expected.ID, field, "a")
			assert.EqualError(t, fmt.Errorf("%s: %v is not a valid int64 type (%T)", field, "a", "a"), err.Error())
		}
	})

	t.Run("bad public key", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Buyer{
			ID: 1,
		}

		err := inMemory.AddBuyer(ctx, expected)
		assert.NoError(t, err)

		err = inMemory.UpdateBuyer(ctx, expected.ID, "PublicKey", "a")
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "PublicKey: failed to decode string public key")
	})

	t.Run("success bool fields", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Buyer{
			ID: 1,
		}

		err := inMemory.AddBuyer(ctx, expected)
		assert.NoError(t, err)

		for _, field := range boolFields {
			err := inMemory.UpdateBuyer(ctx, expected.ID, field, false)
			assert.NoError(t, err)
		}
	})

	t.Run("success float64 fields", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Buyer{
			ID: 1,
		}

		err := inMemory.AddBuyer(ctx, expected)
		assert.NoError(t, err)

		for _, field := range float64Fields {
			err := inMemory.UpdateBuyer(ctx, expected.ID, field, float64(1))
			assert.NoError(t, err)
		}
	})

	t.Run("success int64 fields", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Buyer{
			ID: 1,
		}

		err := inMemory.AddBuyer(ctx, expected)
		assert.NoError(t, err)

		for _, field := range int64Fields {
			err := inMemory.UpdateBuyer(ctx, expected.ID, field, int64(1))
			assert.NoError(t, err)
		}
	})

	t.Run("success string fields - short name", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Buyer{
			ID: 1,
		}

		err := inMemory.AddBuyer(ctx, expected)
		assert.NoError(t, err)

		err = inMemory.UpdateBuyer(ctx, expected.ID, "ShortName", "a")
		assert.NoError(t, err)
	})

	t.Run("success string fields - public key", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Buyer{
			ID: 1,
		}

		err := inMemory.AddBuyer(ctx, expected)
		assert.NoError(t, err)

		publicKey := "YFWQjOJfHfOqsCMM/1pd+c5haMhsrE2Gm05bVUQhCnG7YlPUrI/d1g=="
		err = inMemory.UpdateBuyer(ctx, expected.ID, "PublicKey", publicKey)
		assert.NoError(t, err)
	})
}

func TestInMemoryUpdateSeller(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	stringFields := []string{"ShortName"}

	boolFields := []string{"Secret"}

	float64Fields := []string{"EgressPriceNibblinsPerGB"}

	immutableFields := []string{"ID", "Name", "CompanyCode", "DatabaseID", "CustomerID"}

	t.Run("seller does not exist", func(t *testing.T) {
		inMemory := storage.InMemory{}

		err := inMemory.UpdateSeller(ctx, "0", "", "")
		assert.EqualError(t, err, "seller with reference 0 not found")
	})

	t.Run("unknown field", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Seller{
			ID: "sellerID",
		}

		err := inMemory.AddSeller(ctx, expected)
		assert.NoError(t, err)

		err = inMemory.UpdateSeller(ctx, expected.ID, "unknown field", "")
		assert.EqualError(t, fmt.Errorf("Field '%v' does not exist (or is not editable) on the routing.Seller type", "unknown field"), err.Error())
	})

	t.Run("immutable fields", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Seller{
			ID: "sellerID",
		}

		err := inMemory.AddSeller(ctx, expected)
		assert.NoError(t, err)

		for _, field := range immutableFields {
			err := inMemory.UpdateSeller(ctx, expected.ID, field, "")
			assert.EqualError(t, fmt.Errorf("Field '%v' does not exist (or is not editable) on the routing.Seller type", field), err.Error())
		}
	})

	t.Run("failed string fields", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Seller{
			ID: "sellerID",
		}

		err := inMemory.AddSeller(ctx, expected)
		assert.NoError(t, err)

		for _, field := range stringFields {
			err := inMemory.UpdateSeller(ctx, expected.ID, field, float64(-1))
			assert.EqualError(t, fmt.Errorf("%v is not a valid string value", float64(-1)), err.Error())
		}
	})

	t.Run("failed bool fields", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Seller{
			ID: "sellerID",
		}

		err := inMemory.AddSeller(ctx, expected)
		assert.NoError(t, err)

		for _, field := range boolFields {
			err := inMemory.UpdateSeller(ctx, expected.ID, field, float64(-1))
			assert.EqualError(t, fmt.Errorf("%v is not a valid boolean type", float64(-1)), err.Error())
		}
	})

	t.Run("failed float64 fields", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Seller{
			ID: "sellerID",
		}

		err := inMemory.AddSeller(ctx, expected)
		assert.NoError(t, err)

		for _, field := range float64Fields {
			err := inMemory.UpdateSeller(ctx, expected.ID, field, "a")
			assert.EqualError(t, fmt.Errorf("%v is not a valid float64 type", "a"), err.Error())
		}
	})

	t.Run("success string fields", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Seller{
			ID: "sellerID",
		}

		err := inMemory.AddSeller(ctx, expected)
		assert.NoError(t, err)

		for _, field := range stringFields {
			err := inMemory.UpdateSeller(ctx, expected.ID, field, "newString")
			assert.NoError(t, err)
		}
	})

	t.Run("success bool fields", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Seller{
			ID: "sellerID",
		}

		err := inMemory.AddSeller(ctx, expected)
		assert.NoError(t, err)

		for _, field := range boolFields {
			err := inMemory.UpdateSeller(ctx, expected.ID, field, true)
			assert.NoError(t, err)
		}
	})

	t.Run("success float64 fields", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Seller{
			ID: "sellerID",
		}

		err := inMemory.AddSeller(ctx, expected)
		assert.NoError(t, err)

		for _, field := range float64Fields {
			err := inMemory.UpdateSeller(ctx, expected.ID, field, float64(100))
			assert.NoError(t, err)
		}
	})
}

func TestInMemoryUpdateCustomer(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	stringFields := []string{"AutomaticSigninDomains", "Name"}

	immutableFields := []string{"Code", "BuyerRef", "SellerRef", "DatabaseID"}

	t.Run("customer does not exist", func(t *testing.T) {
		inMemory := storage.InMemory{}

		err := inMemory.UpdateCustomer(ctx, "0", "", "")
		assert.EqualError(t, err, "customer with reference 0 not found")
	})

	t.Run("unknown field", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Customer{
			Code: "customerID",
		}

		err := inMemory.AddCustomer(ctx, expected)
		assert.NoError(t, err)

		err = inMemory.UpdateCustomer(ctx, expected.Code, "unknown field", "")
		assert.EqualError(t, fmt.Errorf("Field '%v' does not exist (or is not editable) on the routing.Customer type", "unknown field"), err.Error())
	})

	t.Run("immutable fields", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Customer{
			Code: "customerID",
		}

		err := inMemory.AddCustomer(ctx, expected)
		assert.NoError(t, err)

		for _, field := range immutableFields {
			err := inMemory.UpdateCustomer(ctx, expected.Code, field, "")
			assert.EqualError(t, fmt.Errorf("Field '%v' does not exist (or is not editable) on the routing.Customer type", field), err.Error())
		}
	})

	t.Run("failed string fields", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Customer{
			Code: "customerID",
		}

		err := inMemory.AddCustomer(ctx, expected)
		assert.NoError(t, err)

		for _, field := range stringFields {
			err := inMemory.UpdateCustomer(ctx, expected.Code, field, float64(-1))
			assert.EqualError(t, fmt.Errorf("%v is not a valid string value", float64(-1)), err.Error())
		}
	})

	t.Run("success string fields", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Customer{
			Code: "sellerID",
		}

		err := inMemory.AddCustomer(ctx, expected)
		assert.NoError(t, err)

		for _, field := range stringFields {
			err := inMemory.UpdateCustomer(ctx, expected.Code, field, "newString")
			assert.NoError(t, err)
		}
	})
}

func TestInMemoryUpdateDatacenter(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	float32Fields := []string{"Latitude", "Longitude"}

	immutableFields := []string{"ID", "Name", "AliasName", "SellerID", "DatabaseID"}

	t.Run("datacenter does not exist", func(t *testing.T) {
		inMemory := storage.InMemory{}

		err := inMemory.UpdateDatacenter(ctx, 0, "", "")
		assert.EqualError(t, err, "datacenter with reference 0 not found")
	})

	t.Run("unknown field", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Datacenter{
			ID:   crypto.HashID("datacenter name"),
			Name: "datadcenter name",
		}

		err := inMemory.AddDatacenter(ctx, expected)
		assert.NoError(t, err)

		err = inMemory.UpdateDatacenter(ctx, expected.ID, "unknown field", "")
		assert.EqualError(t, fmt.Errorf("Field '%v' does not exist (or is not editable) on the routing.Datacenter type", "unknown field"), err.Error())
	})

	t.Run("immutable fields", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Datacenter{
			ID:   crypto.HashID("datacenter name"),
			Name: "datadcenter name",
		}

		err := inMemory.AddDatacenter(ctx, expected)
		assert.NoError(t, err)

		for _, field := range immutableFields {
			err := inMemory.UpdateDatacenter(ctx, expected.ID, field, "")
			assert.EqualError(t, fmt.Errorf("Field '%v' does not exist (or is not editable) on the routing.Datacenter type", field), err.Error())
		}
	})

	t.Run("failed float32 fields", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Datacenter{
			ID:   crypto.HashID("datacenter name"),
			Name: "datadcenter name",
		}

		err := inMemory.AddDatacenter(ctx, expected)
		assert.NoError(t, err)

		for _, field := range float32Fields {
			err := inMemory.UpdateDatacenter(ctx, expected.ID, field, "a")
			assert.EqualError(t, fmt.Errorf("%v is not a valid float32 value", "a"), err.Error())
		}
	})

	t.Run("success float32 fields", func(t *testing.T) {
		inMemory := storage.InMemory{}

		expected := routing.Datacenter{
			ID:   crypto.HashID("datacenter name"),
			Name: "datadcenter name",
		}

		err := inMemory.AddDatacenter(ctx, expected)
		assert.NoError(t, err)

		for _, field := range float32Fields {
			err := inMemory.UpdateDatacenter(ctx, expected.ID, field, float32(23.32))
			assert.NoError(t, err)
		}
	})
}
