package jsonrpc_test

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"testing"

	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport/jsonrpc"
	"github.com/stretchr/testify/assert"
)

func TestBuyers(t *testing.T) {
	storer := storage.InMemory{}
	storer.AddBuyer(context.Background(), routing.Buyer{ID: 1, Name: "local.local.1"})

	svc := jsonrpc.OpsService{
		Storage: &storer,
	}

	t.Run("list", func(t *testing.T) {
		var reply jsonrpc.BuyersReply
		err := svc.Buyers(nil, &jsonrpc.BuyersArgs{}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, reply.Buyers[0].ID, uint64(1))
		assert.Equal(t, reply.Buyers[0].Name, "local.local.1")
	})
}

func TestAddBuyer(t *testing.T) {
	storer := storage.InMemory{}

	svc := jsonrpc.OpsService{
		Storage: &storer,
	}

	publicKey := make([]byte, crypto.KeySize)
	_, err := rand.Read(publicKey)
	assert.NoError(t, err)

	expected := routing.Buyer{
		ID:                   1,
		Name:                 "local buyer",
		Active:               true,
		Live:                 false,
		PublicKey:            publicKey,
		RoutingRulesSettings: routing.DefaultRoutingRulesSettings,
	}

	t.Run("add", func(t *testing.T) {
		var reply jsonrpc.AddBuyerReply
		err := svc.AddBuyer(nil, &jsonrpc.AddBuyerArgs{Buyer: expected}, &reply)
		assert.NoError(t, err)

		var buyersReply jsonrpc.BuyersReply
		err = svc.Buyers(nil, &jsonrpc.BuyersArgs{}, &buyersReply)
		assert.NoError(t, err)

		assert.Len(t, buyersReply.Buyers, 1)
		assert.Equal(t, buyersReply.Buyers[0].ID, expected.ID)
		assert.Equal(t, buyersReply.Buyers[0].Name, expected.Name)
	})

	t.Run("exists", func(t *testing.T) {
		var reply jsonrpc.AddBuyerReply

		err = svc.AddBuyer(nil, &jsonrpc.AddBuyerArgs{Buyer: expected}, &reply)
		assert.EqualError(t, err, "buyer with id 1 already exists in memory storage")
	})
}

func TestSellers(t *testing.T) {
	expected := routing.Seller{
		ID:                "1",
		Name:              "local.local.1",
		IngressPriceCents: 10,
		EgressPriceCents:  20,
	}

	storer := storage.InMemory{}
	storer.AddSeller(context.Background(), expected)

	svc := jsonrpc.OpsService{
		Storage: &storer,
	}

	t.Run("list", func(t *testing.T) {
		var reply jsonrpc.SellersReply
		err := svc.Sellers(nil, &jsonrpc.SellersArgs{}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, reply.Sellers[0].ID, expected.ID)
		assert.Equal(t, reply.Sellers[0].Name, expected.Name)
		assert.Equal(t, reply.Sellers[0].IngressPriceCents, expected.IngressPriceCents)
		assert.Equal(t, reply.Sellers[0].EgressPriceCents, expected.EgressPriceCents)
	})
}

func TestAddSeller(t *testing.T) {
	storer := storage.InMemory{}

	svc := jsonrpc.OpsService{
		Storage: &storer,
	}

	publicKey := make([]byte, crypto.KeySize)
	_, err := rand.Read(publicKey)
	assert.NoError(t, err)

	expected := routing.Seller{
		ID:                "id",
		Name:              "local seller",
		IngressPriceCents: 10,
		EgressPriceCents:  20,
	}

	t.Run("add", func(t *testing.T) {
		var reply jsonrpc.AddSellerReply
		err := svc.AddSeller(nil, &jsonrpc.AddSellerArgs{Seller: expected}, &reply)
		assert.NoError(t, err)

		var sellersReply jsonrpc.SellersReply
		err = svc.Sellers(nil, &jsonrpc.SellersArgs{}, &sellersReply)
		assert.NoError(t, err)

		assert.Len(t, sellersReply.Sellers, 1)
		assert.Equal(t, sellersReply.Sellers[0].ID, expected.ID)
		assert.Equal(t, sellersReply.Sellers[0].Name, expected.Name)
		assert.Equal(t, sellersReply.Sellers[0].IngressPriceCents, expected.IngressPriceCents)
		assert.Equal(t, sellersReply.Sellers[0].EgressPriceCents, expected.EgressPriceCents)
	})

	t.Run("exists", func(t *testing.T) {
		var reply jsonrpc.AddSellerReply

		err = svc.AddSeller(nil, &jsonrpc.AddSellerArgs{Seller: expected}, &reply)
		assert.EqualError(t, err, "seller with id id already exists in memory storage")
	})
}

func TestRelays(t *testing.T) {
	storer := storage.InMemory{}
	storer.AddRelay(context.Background(), routing.Relay{ID: 1, Name: "local.local.1"})
	storer.AddRelay(context.Background(), routing.Relay{ID: 2, Name: "local.local.2"})

	svc := jsonrpc.OpsService{
		Storage: &storer,
	}

	t.Run("list", func(t *testing.T) {
		var reply jsonrpc.RelaysReply
		err := svc.Relays(nil, &jsonrpc.RelaysArgs{}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, reply.Relays[0].ID, uint64(1))
		assert.Equal(t, reply.Relays[0].Name, "local.local.1")
		assert.Equal(t, reply.Relays[1].ID, uint64(2))
		assert.Equal(t, reply.Relays[1].Name, "local.local.2")
	})

	t.Run("filter", func(t *testing.T) {
		var reply jsonrpc.RelaysReply
		err := svc.Relays(nil, &jsonrpc.RelaysArgs{Name: "local.1"}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, len(reply.Relays), 1)
		assert.Equal(t, reply.Relays[0].ID, uint64(1))
		assert.Equal(t, reply.Relays[0].Name, "local.local.1")

		var empty jsonrpc.RelaysReply
		err = svc.Relays(nil, &jsonrpc.RelaysArgs{Name: "not.found"}, &empty)
		assert.NoError(t, err)

		assert.Equal(t, len(empty.Relays), 0)
	})
}

func TestRelayStateUpdate(t *testing.T) {
	makeSvc := func() *jsonrpc.OpsService {
		var storer storage.InMemory
		storer.AddRelay(context.Background(), routing.Relay{
			ID:    1,
			State: 0,
		})
		storer.AddRelay(context.Background(), routing.Relay{
			ID:    2,
			State: 123456,
		})

		return &jsonrpc.OpsService{
			Storage: &storer,
		}
	}

	t.Run("found", func(t *testing.T) {
		svc := makeSvc()
		err := svc.RelayStateUpdate(nil, &jsonrpc.RelayStateUpdateArgs{
			RelayID:    1,
			RelayState: routing.RelayStateDisabled,
		}, &jsonrpc.RelayStateUpdateReply{})
		assert.NoError(t, err)

		relay, err := svc.Storage.Relay(1)
		assert.NoError(t, err)
		assert.Equal(t, routing.RelayStateDisabled, relay.State)

		relay, err = svc.Storage.Relay(2)
		assert.NoError(t, err)
		assert.Equal(t, routing.RelayState(123456), relay.State)
	})

	t.Run("not found", func(t *testing.T) {
		svc := makeSvc()
		err := svc.RelayStateUpdate(nil, &jsonrpc.RelayStateUpdateArgs{
			RelayID:    987654321,
			RelayState: routing.RelayStateDisabled,
		}, &jsonrpc.RelayStateUpdateReply{})
		assert.Error(t, err)

		relay, err := svc.Storage.Relay(1)
		assert.NoError(t, err)
		assert.Equal(t, routing.RelayState(0), relay.State)

		relay, err = svc.Storage.Relay(2)
		assert.NoError(t, err)
		assert.Equal(t, routing.RelayState(123456), relay.State)
	})
}

func TestRelayPublicKeyUpdate(t *testing.T) {
	makeSvc := func() *jsonrpc.OpsService {
		var storer storage.InMemory
		storer.AddRelay(context.Background(), routing.Relay{
			ID:        1,
			PublicKey: []byte("oldpublickey"),
		})
		storer.AddRelay(context.Background(), routing.Relay{
			ID:        2,
			PublicKey: []byte("oldpublickey"),
		})

		return &jsonrpc.OpsService{
			Storage: &storer,
		}
	}

	t.Run("found", func(t *testing.T) {
		svc := makeSvc()
		err := svc.RelayPublicKeyUpdate(nil, &jsonrpc.RelayPublicKeyUpdateArgs{
			RelayID:        1,
			RelayPublicKey: "newpublickey",
		}, &jsonrpc.RelayPublicKeyUpdateReply{})
		assert.NoError(t, err)

		relay, err := svc.Storage.Relay(1)
		assert.NoError(t, err)
		assert.Equal(t, "newpublickey", base64.StdEncoding.EncodeToString(relay.PublicKey))

		relay, err = svc.Storage.Relay(2)
		assert.NoError(t, err)
		assert.Equal(t, []byte("oldpublickey"), relay.PublicKey)
	})

	t.Run("not found", func(t *testing.T) {
		svc := makeSvc()
		err := svc.RelayPublicKeyUpdate(nil, &jsonrpc.RelayPublicKeyUpdateArgs{
			RelayID:        987654321,
			RelayPublicKey: "newpublickey",
		}, &jsonrpc.RelayPublicKeyUpdateReply{})
		assert.Error(t, err)

		relay, err := svc.Storage.Relay(1)
		assert.NoError(t, err)
		assert.Equal(t, []byte("oldpublickey"), relay.PublicKey)

		relay, err = svc.Storage.Relay(2)
		assert.NoError(t, err)
		assert.Equal(t, []byte("oldpublickey"), relay.PublicKey)
	})
}

func TestDatacenters(t *testing.T) {
	storer := storage.InMemory{}
	storer.AddDatacenter(context.Background(), routing.Datacenter{ID: 1, Name: "local.local.1"})
	storer.AddDatacenter(context.Background(), routing.Datacenter{ID: 2, Name: "local.local.2"})

	svc := jsonrpc.OpsService{
		Storage: &storer,
	}

	t.Run("list", func(t *testing.T) {
		var reply jsonrpc.DatacentersReply
		err := svc.Datacenters(nil, &jsonrpc.DatacentersArgs{}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, reply.Datacenters[0].Name, "local.local.1")
		assert.Equal(t, reply.Datacenters[1].Name, "local.local.2")
	})

	t.Run("filter", func(t *testing.T) {
		var reply jsonrpc.DatacentersReply
		err := svc.Datacenters(nil, &jsonrpc.DatacentersArgs{Name: "local.1"}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, len(reply.Datacenters), 1)
		assert.Equal(t, reply.Datacenters[0].Name, "local.local.1")

		var empty jsonrpc.DatacentersReply
		err = svc.Datacenters(nil, &jsonrpc.DatacentersArgs{Name: "not.found"}, &empty)
		assert.NoError(t, err)

		assert.Equal(t, len(empty.Datacenters), 0)
	})
}

func TestAddDatacenter(t *testing.T) {
	storer := storage.InMemory{}

	svc := jsonrpc.OpsService{
		Storage: &storer,
	}

	publicKey := make([]byte, crypto.KeySize)
	_, err := rand.Read(publicKey)
	assert.NoError(t, err)

	expected := routing.Datacenter{
		ID:      1,
		Name:    "local datacenter",
		Enabled: false,
		Location: routing.Location{
			Latitude:  70.5,
			Longitude: 120.5,
		},
	}

	t.Run("add", func(t *testing.T) {
		var reply jsonrpc.AddDatacenterReply
		err := svc.AddDatacenter(nil, &jsonrpc.AddDatacenterArgs{Datacenter: expected}, &reply)
		assert.NoError(t, err)

		var datacentersReply jsonrpc.DatacentersReply
		err = svc.Datacenters(nil, &jsonrpc.DatacentersArgs{}, &datacentersReply)
		assert.NoError(t, err)

		assert.Len(t, datacentersReply.Datacenters, 1)
		assert.Equal(t, datacentersReply.Datacenters[0].Name, expected.Name)
		assert.Equal(t, datacentersReply.Datacenters[0].Latitude, expected.Location.Latitude)
		assert.Equal(t, datacentersReply.Datacenters[0].Longitude, expected.Location.Longitude)
		assert.Equal(t, datacentersReply.Datacenters[0].Enabled, expected.Enabled)
	})

	t.Run("exists", func(t *testing.T) {
		var reply jsonrpc.AddDatacenterReply

		err = svc.AddDatacenter(nil, &jsonrpc.AddDatacenterArgs{Datacenter: expected}, &reply)
		assert.EqualError(t, err, "datacenter with id 1 already exists in memory storage")
	})
}
