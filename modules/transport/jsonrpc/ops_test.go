package jsonrpc_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/go-kit/kit/log"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/storage"
	"github.com/networknext/backend/modules/transport/jsonrpc"
	"github.com/networknext/backend/modules/transport/middleware"
	"github.com/networknext/backend/modules/transport/notifications"
	"github.com/stretchr/testify/assert"
)

/*func TestBuyers(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}
	storer.AddBuyer(context.Background(), routing.Buyer{ID: 1, Name: "local.local.1"})

	logger := log.NewNopLogger()
	svc := jsonrpc.OpsService{
		Storage: &storer,
		Logger:  logger,
	}

	t.Run("list", func(t *testing.T) {
		var reply jsonrpc.BuyersReply
		err := svc.Buyers(nil, &jsonrpc.BuyersArgs{}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, reply.Buyers[0].ID, "1")
		assert.Equal(t, reply.Buyers[0].Name, "local.local.1")
	})
}

// 1 customer with a buyer and a seller ID
func TestCustomersSingle(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}
	storer.AddBuyer(context.Background(), routing.Buyer{ID: 1, Name: "Fred Scuttle"})
	storer.AddSeller(context.Background(), routing.Seller{ID: "some seller", Name: "Fred Scuttle"})

	logger := log.NewNopLogger()
	svc := jsonrpc.OpsService{
		Storage: &storer,
		Logger:  logger,
	}

	t.Run("single customer", func(t *testing.T) {
		var reply jsonrpc.CustomersReply
		err := svc.Customers(nil, &jsonrpc.CustomersArgs{}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, "1", reply.Customers[0].BuyerID)
		assert.Equal(t, "some seller", reply.Customers[0].SellerID)
		assert.Equal(t, "Fred Scuttle", reply.Customers[0].Name)
	})
}

// Multiple customers with different names (2 records)
func TestCustomersMultiple(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}
	storer.AddBuyer(context.Background(), routing.Buyer{ID: 1, Name: "Fred Scuttle"})
	storer.AddSeller(context.Background(), routing.Seller{ID: "some seller", Name: "Bull Winkle"})

	logger := log.NewNopLogger()
	svc := jsonrpc.OpsService{
		Storage: &storer,
		Logger:  logger,
	}

	t.Run("multiple customers", func(t *testing.T) {
		var reply jsonrpc.CustomersReply
		err := svc.Customers(nil, &jsonrpc.CustomersArgs{}, &reply)
		assert.NoError(t, err)

		// sorted alphabetically by name
		assert.Equal(t, "", reply.Customers[0].BuyerID)
		assert.Equal(t, "some seller", reply.Customers[0].SellerID)
		assert.Equal(t, "Bull Winkle", reply.Customers[0].Name)

		assert.Equal(t, "1", reply.Customers[1].BuyerID)
		assert.Equal(t, "", reply.Customers[1].SellerID)
		assert.Equal(t, "Fred Scuttle", reply.Customers[1].Name)
	})
}

func TestAddBuyer(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	logger := log.NewNopLogger()
	svc := jsonrpc.OpsService{
		Storage: &storer,
		Logger:  logger,
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
		assert.Equal(t, buyersReply.Buyers[0].ID, fmt.Sprintf("%x", expected.ID))
		assert.Equal(t, buyersReply.Buyers[0].Name, expected.Name)
	})

	t.Run("exists", func(t *testing.T) {
		var reply jsonrpc.AddBuyerReply

		err = svc.AddBuyer(nil, &jsonrpc.AddBuyerArgs{Buyer: expected}, &reply)
		assert.EqualError(t, err, "buyer with reference 1 already exists")
	})
}

func TestRemoveBuyer(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	logger := log.NewNopLogger()
	svc := jsonrpc.OpsService{
		Storage: &storer,
		Logger:  logger,
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

	t.Run("doesn't exist", func(t *testing.T) {
		var reply jsonrpc.RemoveBuyerReply

		err = svc.RemoveBuyer(nil, &jsonrpc.RemoveBuyerArgs{ID: fmt.Sprintf("%x", expected.ID)}, &reply)
		assert.EqualError(t, err, "buyer with reference 1 not found")
	})

	t.Run("remove", func(t *testing.T) {
		var addReply jsonrpc.AddBuyerReply
		err := svc.AddBuyer(nil, &jsonrpc.AddBuyerArgs{Buyer: expected}, &addReply)
		assert.NoError(t, err)

		var reply jsonrpc.RemoveBuyerReply
		err = svc.RemoveBuyer(nil, &jsonrpc.RemoveBuyerArgs{ID: fmt.Sprintf("%x", expected.ID)}, &reply)
		assert.NoError(t, err)

		var buyersReply jsonrpc.BuyersReply
		err = svc.Buyers(nil, &jsonrpc.BuyersArgs{}, &buyersReply)
		assert.NoError(t, err)

		assert.Len(t, buyersReply.Buyers, 0)
	})
}

func TestRoutingRulesSettings(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	logger := log.NewNopLogger()
	svc := jsonrpc.OpsService{
		Storage: &storer,
		Logger:  logger,
	}

	t.Run("doesn't exist", func(t *testing.T) {
		var reply jsonrpc.RoutingRulesSettingsReply

		err := svc.RoutingRulesSettings(nil, &jsonrpc.RoutingRulesSettingsArgs{BuyerID: "0"}, &reply)
		assert.EqualError(t, err, "buyer with reference 0 not found")
	})

	t.Run("list", func(t *testing.T) {
		storer.AddBuyer(context.Background(), routing.Buyer{ID: 0, Name: "local.local.1", RoutingRulesSettings: routing.DefaultRoutingRulesSettings})

		var reply jsonrpc.RoutingRulesSettingsReply
		err := svc.RoutingRulesSettings(nil, &jsonrpc.RoutingRulesSettingsArgs{BuyerID: "0"}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, reply.RoutingRuleSettings[0].EnvelopeKbpsUp, routing.DefaultRoutingRulesSettings.EnvelopeKbpsUp)
		assert.Equal(t, reply.RoutingRuleSettings[0].EnvelopeKbpsDown, routing.DefaultRoutingRulesSettings.EnvelopeKbpsDown)
		assert.Equal(t, reply.RoutingRuleSettings[0].Mode, routing.DefaultRoutingRulesSettings.Mode)
		assert.Equal(t, reply.RoutingRuleSettings[0].MaxCentsPerGB, routing.DefaultRoutingRulesSettings.MaxCentsPerGB)
		assert.Equal(t, reply.RoutingRuleSettings[0].RTTEpsilon, routing.DefaultRoutingRulesSettings.RTTEpsilon)
		assert.Equal(t, reply.RoutingRuleSettings[0].RTTThreshold, routing.DefaultRoutingRulesSettings.RTTThreshold)
		assert.Equal(t, reply.RoutingRuleSettings[0].RTTHysteresis, routing.DefaultRoutingRulesSettings.RTTHysteresis)
		assert.Equal(t, reply.RoutingRuleSettings[0].RTTVeto, routing.DefaultRoutingRulesSettings.RTTVeto)
		assert.Equal(t, reply.RoutingRuleSettings[0].EnableYouOnlyLiveOnce, routing.DefaultRoutingRulesSettings.EnableYouOnlyLiveOnce)
		assert.Equal(t, reply.RoutingRuleSettings[0].EnablePacketLossSafety, routing.DefaultRoutingRulesSettings.EnablePacketLossSafety)
		assert.Equal(t, reply.RoutingRuleSettings[0].EnableMultipathForPacketLoss, routing.DefaultRoutingRulesSettings.EnableMultipathForPacketLoss)
		assert.Equal(t, reply.RoutingRuleSettings[0].EnableMultipathForJitter, routing.DefaultRoutingRulesSettings.EnableMultipathForJitter)
		assert.Equal(t, reply.RoutingRuleSettings[0].EnableMultipathForRTT, routing.DefaultRoutingRulesSettings.EnableMultipathForRTT)
		assert.Equal(t, reply.RoutingRuleSettings[0].EnableABTest, routing.DefaultRoutingRulesSettings.EnableABTest)
		assert.Equal(t, reply.RoutingRuleSettings[0].EnableTryBeforeYouBuy, routing.DefaultRoutingRulesSettings.EnableTryBeforeYouBuy)
		assert.Equal(t, reply.RoutingRuleSettings[0].TryBeforeYouBuyMaxSlices, routing.DefaultRoutingRulesSettings.TryBeforeYouBuyMaxSlices)
	})
}

func TestSetRoutingRulesSettings(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	logger := log.NewNopLogger()
	svc := jsonrpc.OpsService{
		Storage: &storer,
		Logger:  logger,
	}

	t.Run("doesn't exist", func(t *testing.T) {
		var reply jsonrpc.SetRoutingRulesSettingsReply

		err := svc.SetRoutingRulesSettings(nil, &jsonrpc.SetRoutingRulesSettingsArgs{BuyerID: "0", RoutingRulesSettings: routing.LocalRoutingRulesSettings}, &reply)
		assert.EqualError(t, err, "SetRoutingRulesSettings() Storage.Buyer error: buyer with reference 0 not found")
	})

	t.Run("set", func(t *testing.T) {
		storer.AddBuyer(context.Background(), routing.Buyer{ID: 1, Name: "local.local.1", RoutingRulesSettings: routing.DefaultRoutingRulesSettings})

		var reply jsonrpc.SetRoutingRulesSettingsReply
		err := svc.SetRoutingRulesSettings(nil, &jsonrpc.SetRoutingRulesSettingsArgs{BuyerID: "1", RoutingRulesSettings: routing.LocalRoutingRulesSettings}, &reply)
		assert.NoError(t, err)

		var rrsReply jsonrpc.RoutingRulesSettingsReply
		err = svc.RoutingRulesSettings(nil, &jsonrpc.RoutingRulesSettingsArgs{BuyerID: "1"}, &rrsReply)
		assert.NoError(t, err)

		assert.Equal(t, rrsReply.RoutingRuleSettings[0].EnvelopeKbpsUp, routing.LocalRoutingRulesSettings.EnvelopeKbpsUp)
		assert.Equal(t, rrsReply.RoutingRuleSettings[0].EnvelopeKbpsDown, routing.LocalRoutingRulesSettings.EnvelopeKbpsDown)
		assert.Equal(t, rrsReply.RoutingRuleSettings[0].Mode, routing.LocalRoutingRulesSettings.Mode)
		assert.Equal(t, rrsReply.RoutingRuleSettings[0].MaxCentsPerGB, routing.LocalRoutingRulesSettings.MaxCentsPerGB)
		assert.Equal(t, rrsReply.RoutingRuleSettings[0].RTTEpsilon, routing.LocalRoutingRulesSettings.RTTEpsilon)
		assert.Equal(t, rrsReply.RoutingRuleSettings[0].RTTThreshold, routing.LocalRoutingRulesSettings.RTTThreshold)
		assert.Equal(t, rrsReply.RoutingRuleSettings[0].RTTHysteresis, routing.LocalRoutingRulesSettings.RTTHysteresis)
		assert.Equal(t, rrsReply.RoutingRuleSettings[0].RTTVeto, routing.LocalRoutingRulesSettings.RTTVeto)
		assert.Equal(t, rrsReply.RoutingRuleSettings[0].EnableYouOnlyLiveOnce, routing.LocalRoutingRulesSettings.EnableYouOnlyLiveOnce)
		assert.Equal(t, rrsReply.RoutingRuleSettings[0].EnablePacketLossSafety, routing.LocalRoutingRulesSettings.EnablePacketLossSafety)
		assert.Equal(t, rrsReply.RoutingRuleSettings[0].EnableMultipathForPacketLoss, routing.LocalRoutingRulesSettings.EnableMultipathForPacketLoss)
		assert.Equal(t, rrsReply.RoutingRuleSettings[0].EnableMultipathForJitter, routing.LocalRoutingRulesSettings.EnableMultipathForJitter)
		assert.Equal(t, rrsReply.RoutingRuleSettings[0].EnableMultipathForRTT, routing.LocalRoutingRulesSettings.EnableMultipathForRTT)
		assert.Equal(t, rrsReply.RoutingRuleSettings[0].EnableABTest, routing.LocalRoutingRulesSettings.EnableABTest)
		assert.Equal(t, rrsReply.RoutingRuleSettings[0].EnableTryBeforeYouBuy, routing.DefaultRoutingRulesSettings.EnableTryBeforeYouBuy)
	})
}

func TestSellers(t *testing.T) {
	t.Parallel()

	expected := routing.Seller{
		ID:                "1",
		Name:              "local.local.1",
		IngressPriceCents: 10,
		EgressPriceCents:  20,
	}

	storer := storage.InMemory{}
	storer.AddSeller(context.Background(), expected)

	logger := log.NewNopLogger()
	svc := jsonrpc.OpsService{
		Storage: &storer,
		Logger:  logger,
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
	t.Parallel()

	storer := storage.InMemory{}

	logger := log.NewNopLogger()
	svc := jsonrpc.OpsService{
		Storage: &storer,
		Logger:  logger,
	}

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

		err := svc.AddSeller(nil, &jsonrpc.AddSellerArgs{Seller: expected}, &reply)
		assert.EqualError(t, err, "AddSeller() error: seller with reference id already exists")
	})
}

func TestRemoveSeller(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	logger := log.NewNopLogger()
	svc := jsonrpc.OpsService{
		Storage: &storer,
		Logger:  logger,
	}

	expected := routing.Seller{
		ID:                "1",
		Name:              "local seller",
		IngressPriceCents: 10,
		EgressPriceCents:  20,
	}

	t.Run("doesn't exist", func(t *testing.T) {
		var reply jsonrpc.RemoveSellerReply

		err := svc.RemoveSeller(nil, &jsonrpc.RemoveSellerArgs{ID: expected.ID}, &reply)
		assert.EqualError(t, err, "RemoveSeller() error: seller with reference 1 not found")
	})

	t.Run("remove", func(t *testing.T) {
		var addReply jsonrpc.AddSellerReply
		err := svc.AddSeller(nil, &jsonrpc.AddSellerArgs{Seller: expected}, &addReply)
		assert.NoError(t, err)

		var reply jsonrpc.RemoveSellerReply
		err = svc.RemoveSeller(nil, &jsonrpc.RemoveSellerArgs{ID: expected.ID}, &reply)
		assert.NoError(t, err)

		var sellersReply jsonrpc.SellersReply
		err = svc.Sellers(nil, &jsonrpc.SellersArgs{}, &sellersReply)
		assert.NoError(t, err)

		assert.Len(t, sellersReply.Sellers, 0)
	})
}

func TestRelays(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	seller := routing.Seller{
		ID:   "sellerID",
		Name: "seller name",
	}

	datacenter := routing.Datacenter{
		ID:   crypto.HashID("datacenter name"),
		Name: "datacenter name",
	}

	relay1 := routing.Relay{
		ID:         1,
		Name:       "local.local.1",
		Seller:     seller,
		Datacenter: datacenter,
	}

	relay2 := routing.Relay{
		ID:         2,
		Name:       "local.local.2",
		Seller:     seller,
		Datacenter: datacenter,
	}

	relay3 := routing.Relay{
		ID:         3,
		Name:       "local.local.23",
		Seller:     seller,
		Datacenter: datacenter,
	}

	err := storer.AddSeller(context.Background(), seller)
	assert.NoError(t, err)
	err = storer.AddDatacenter(context.Background(), datacenter)
	assert.NoError(t, err)
	err = storer.AddRelay(context.Background(), relay1)
	assert.NoError(t, err)
	err = storer.AddRelay(context.Background(), relay2)
	assert.NoError(t, err)
	err = storer.AddRelay(context.Background(), relay3)
	assert.NoError(t, err)

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	logger := log.NewNopLogger()
	svc := jsonrpc.OpsService{
		Storage:     &storer,
		RedisClient: redisClient,
		Logger:      logger,
	}

	t.Run("list", func(t *testing.T) {
		var reply jsonrpc.RelaysReply
		err := svc.Relays(nil, &jsonrpc.RelaysArgs{}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, reply.Relays[0].ID, uint64(1))
		assert.Equal(t, reply.Relays[0].Name, "local.local.1")
		assert.Equal(t, reply.Relays[1].ID, uint64(2))
		assert.Equal(t, reply.Relays[1].Name, "local.local.2")
		assert.Equal(t, reply.Relays[2].ID, uint64(3))
		assert.Equal(t, reply.Relays[2].Name, "local.local.23")
	})

	t.Run("exact match", func(t *testing.T) {
		var reply jsonrpc.RelaysReply
		err := svc.Relays(nil, &jsonrpc.RelaysArgs{Regex: "local.local.2"}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, len(reply.Relays), 1)
		assert.Equal(t, reply.Relays[0].ID, uint64(2))
		assert.Equal(t, reply.Relays[0].Name, "local.local.2")

		var empty jsonrpc.RelaysReply
		err = svc.Relays(nil, &jsonrpc.RelaysArgs{Regex: "not.found"}, &empty)
		assert.NoError(t, err)

		assert.Equal(t, len(empty.Relays), 0)
	})

	t.Run("filter", func(t *testing.T) {
		var reply jsonrpc.RelaysReply
		err := svc.Relays(nil, &jsonrpc.RelaysArgs{Regex: "local.1"}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, len(reply.Relays), 1)
		assert.Equal(t, reply.Relays[0].ID, uint64(1))
		assert.Equal(t, reply.Relays[0].Name, "local.local.1")

		var empty jsonrpc.RelaysReply
		err = svc.Relays(nil, &jsonrpc.RelaysArgs{Regex: "not.found"}, &empty)
		assert.NoError(t, err)

		assert.Equal(t, len(empty.Relays), 0)
	})

	t.Run("filter by seller", func(t *testing.T) {
		var reply jsonrpc.RelaysReply
		err := svc.Relays(nil, &jsonrpc.RelaysArgs{Regex: "seller name"}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, len(reply.Relays), 3)
		assert.Equal(t, reply.Relays[0].ID, uint64(1))
		assert.Equal(t, reply.Relays[0].Name, "local.local.1")
		assert.Equal(t, reply.Relays[1].ID, uint64(2))
		assert.Equal(t, reply.Relays[1].Name, "local.local.2")
		assert.Equal(t, reply.Relays[2].ID, uint64(3))
		assert.Equal(t, reply.Relays[2].Name, "local.local.23")

		var empty jsonrpc.RelaysReply
		err = svc.Relays(nil, &jsonrpc.RelaysArgs{Regex: "not.found"}, &empty)
		assert.NoError(t, err)

		assert.Equal(t, len(empty.Relays), 0)
	})
}

func TestAddRelay(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	logger := log.NewNopLogger()
	svc := jsonrpc.OpsService{
		Storage:     &storer,
		RedisClient: redisClient,
		Logger:      logger,
	}

	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
	assert.NoError(t, err)

	expected := routing.Relay{
		ID:   crypto.HashID(addr.String()),
		Name: "local relay",
		Addr: *addr,
	}

	t.Run("seller doesn't exist", func(t *testing.T) {
		var reply jsonrpc.AddRelayReply
		err := svc.AddRelay(nil, &jsonrpc.AddRelayArgs{Relay: expected}, &reply)
		assert.EqualError(t, err, "AddRelay() error: seller with reference  not found")
	})

	t.Run("datacenter doesn't exist", func(t *testing.T) {
		expected.Seller = routing.Seller{
			ID:                "sellerID",
			Name:              "seller name",
			IngressPriceCents: 10,
			EgressPriceCents:  20,
		}

		var sellerReply jsonrpc.AddSellerReply
		err := svc.AddSeller(nil, &jsonrpc.AddSellerArgs{Seller: expected.Seller}, &sellerReply)
		assert.NoError(t, err)

		var reply jsonrpc.AddRelayReply
		err = svc.AddRelay(nil, &jsonrpc.AddRelayArgs{Relay: expected}, &reply)
		assert.EqualError(t, err, "AddRelay() error: datacenter with reference 0 not found")
	})

	t.Run("add", func(t *testing.T) {
		expected.Datacenter = routing.Datacenter{
			ID:       crypto.HashID("datacenter name"),
			Name:     "datacenter name",
			Enabled:  true,
			Location: routing.LocationNullIsland,
		}

		var datacenterReply jsonrpc.AddDatacenterReply
		err := svc.AddDatacenter(nil, &jsonrpc.AddDatacenterArgs{Datacenter: expected.Datacenter}, &datacenterReply)
		assert.NoError(t, err)

		var reply jsonrpc.AddRelayReply
		err = svc.AddRelay(nil, &jsonrpc.AddRelayArgs{Relay: expected}, &reply)
		assert.NoError(t, err)

		var relaysReply jsonrpc.RelaysReply
		err = svc.Relays(nil, &jsonrpc.RelaysArgs{}, &relaysReply)
		assert.NoError(t, err)

		assert.Len(t, relaysReply.Relays, 1)
		assert.Equal(t, relaysReply.Relays[0].ID, expected.ID)
		assert.Equal(t, relaysReply.Relays[0].Name, expected.Name)
		assert.Equal(t, relaysReply.Relays[0].Addr, expected.Addr.String())
	})

	t.Run("exists", func(t *testing.T) {
		var reply jsonrpc.AddRelayReply

		err = svc.AddRelay(nil, &jsonrpc.AddRelayArgs{Relay: expected}, &reply)
		assert.EqualError(t, err, fmt.Sprintf("AddRelay() error: relay with reference %d already exists", expected.ID))
	})
}

func TestRemoveRelay(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	redisServer, err := miniredis.Run()
	assert.NoError(t, err)
	redisClient := redis.NewClient(&redis.Options{Addr: redisServer.Addr()})

	logger := log.NewNopLogger()
	svc := jsonrpc.OpsService{
		Storage:     &storer,
		RedisClient: redisClient,
		Logger:      logger,
	}

	addr, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
	assert.NoError(t, err)

	seller := routing.Seller{
		ID:   "sellerID",
		Name: "seller name",
	}

	datacenter := routing.Datacenter{
		ID:   crypto.HashID("datacenter name"),
		Name: "datacenter name",
	}

	expected := routing.Relay{
		ID:         crypto.HashID(addr.String()),
		Name:       "local relay",
		Addr:       *addr,
		Seller:     seller,
		Datacenter: datacenter,
	}

	svc.AddSeller(nil, &jsonrpc.AddSellerArgs{Seller: seller}, &jsonrpc.AddSellerReply{})
	svc.AddDatacenter(nil, &jsonrpc.AddDatacenterArgs{Datacenter: datacenter}, &jsonrpc.AddDatacenterReply{})

	t.Run("doesn't exist", func(t *testing.T) {
		var reply jsonrpc.RemoveRelayReply

		err = svc.RemoveRelay(nil, &jsonrpc.RemoveRelayArgs{RelayID: expected.ID}, &reply)
		assert.EqualError(t, err, fmt.Sprintf("RemoveRelay() Storage.Relay error: relay with reference %d not found", expected.ID))
	})

	t.Run("remove", func(t *testing.T) {
		var addReply jsonrpc.AddRelayReply
		err := svc.AddRelay(nil, &jsonrpc.AddRelayArgs{Relay: expected}, &addReply)
		assert.NoError(t, err)

		var reply jsonrpc.RemoveRelayReply
		err = svc.RemoveRelay(nil, &jsonrpc.RemoveRelayArgs{RelayID: expected.ID}, &reply)
		assert.NoError(t, err)

		var relaysReply jsonrpc.RelaysReply
		err = svc.Relays(nil, &jsonrpc.RelaysArgs{}, &relaysReply)
		assert.NoError(t, err)

		// Remove shouldn't actually remove it anymore, just set the state to decommissioned
		assert.Len(t, relaysReply.Relays, 1)
		assert.Equal(t, relaysReply.Relays[0].ID, expected.ID)
		assert.Equal(t, relaysReply.Relays[0].State, routing.RelayStateDecommissioned.String())
	})
}

func TestRelayStateUpdate(t *testing.T) {
	t.Parallel()

	logger := log.NewNopLogger()
	makeSvc := func() *jsonrpc.OpsService {
		var storer storage.InMemory

		seller := routing.Seller{
			ID:   "sellerID",
			Name: "seller name",
		}

		datacenter := routing.Datacenter{
			ID:   crypto.HashID("datacenter name"),
			Name: "datacenter name",
		}

		storer.AddSeller(context.Background(), seller)
		storer.AddDatacenter(context.Background(), datacenter)

		err := storer.AddRelay(context.Background(), routing.Relay{
			ID:         1,
			State:      0,
			Seller:     seller,
			Datacenter: datacenter,
		})
		assert.NoError(t, err)
		err = storer.AddRelay(context.Background(), routing.Relay{
			ID:         2,
			State:      123456,
			Seller:     seller,
			Datacenter: datacenter,
		})
		assert.NoError(t, err)

		return &jsonrpc.OpsService{
			Storage: &storer,
			Logger:  logger,
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
	t.Parallel()

	logger := log.NewNopLogger()

	makeSvc := func() *jsonrpc.OpsService {
		var storer storage.InMemory

		seller := routing.Seller{
			ID:   "sellerID",
			Name: "seller name",
		}

		datacenter := routing.Datacenter{
			ID:   crypto.HashID("datacenter name"),
			Name: "datacenter name",
		}

		storer.AddSeller(context.Background(), seller)
		storer.AddDatacenter(context.Background(), datacenter)

		err := storer.AddRelay(context.Background(), routing.Relay{
			ID:         1,
			PublicKey:  []byte("oldpublickey"),
			Seller:     seller,
			Datacenter: datacenter,
		})
		assert.NoError(t, err)
		err = storer.AddRelay(context.Background(), routing.Relay{
			ID:         2,
			PublicKey:  []byte("oldpublickey"),
			Seller:     seller,
			Datacenter: datacenter,
		})
		assert.NoError(t, err)

		return &jsonrpc.OpsService{
			Storage: &storer,
			Logger:  logger,
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

func TestRelayNICSpeedUpdate(t *testing.T) {
	t.Parallel()

	logger := log.NewNopLogger()

	makeSvc := func() *jsonrpc.OpsService {
		var storer storage.InMemory

		seller := routing.Seller{
			ID:   "sellerID",
			Name: "seller name",
		}

		datacenter := routing.Datacenter{
			ID:   crypto.HashID("datacenter name"),
			Name: "datacenter name",
		}

		storer.AddSeller(context.Background(), seller)
		storer.AddDatacenter(context.Background(), datacenter)

		err := storer.AddRelay(context.Background(), routing.Relay{
			ID:           1,
			NICSpeedMbps: 1000,
			Seller:       seller,
			Datacenter:   datacenter,
		})
		assert.NoError(t, err)
		err = storer.AddRelay(context.Background(), routing.Relay{
			ID:           2,
			NICSpeedMbps: 2000,
			Seller:       seller,
			Datacenter:   datacenter,
		})
		assert.NoError(t, err)

		return &jsonrpc.OpsService{
			Storage: &storer,
			Logger:  logger,
		}
	}

	t.Run("found", func(t *testing.T) {
		svc := makeSvc()
		err := svc.RelayNICSpeedUpdate(nil, &jsonrpc.RelayNICSpeedUpdateArgs{
			RelayID:       1,
			RelayNICSpeed: 10000,
		}, &jsonrpc.RelayNICSpeedUpdateReply{})
		assert.NoError(t, err)

		relay, err := svc.Storage.Relay(1)
		assert.NoError(t, err)
		assert.Equal(t, uint64(10000), relay.NICSpeedMbps)

		relay, err = svc.Storage.Relay(2)
		assert.NoError(t, err)
		assert.Equal(t, uint64(2000), relay.NICSpeedMbps)
	})

	t.Run("not found", func(t *testing.T) {
		svc := makeSvc()
		err := svc.RelayNICSpeedUpdate(nil, &jsonrpc.RelayNICSpeedUpdateArgs{
			RelayID:       987654321,
			RelayNICSpeed: 10000,
		}, &jsonrpc.RelayNICSpeedUpdateReply{})
		assert.Error(t, err)

		relay, err := svc.Storage.Relay(1)
		assert.NoError(t, err)
		assert.Equal(t, uint64(1000), relay.NICSpeedMbps)

		relay, err = svc.Storage.Relay(2)
		assert.NoError(t, err)
		assert.Equal(t, uint64(2000), relay.NICSpeedMbps)
	})
}

func TestDatacenters(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}
	storer.AddDatacenter(context.Background(), routing.Datacenter{ID: 1, Name: "local.local.1"})
	storer.AddDatacenter(context.Background(), routing.Datacenter{ID: 2, Name: "local.local.2"})

	logger := log.NewNopLogger()
	svc := jsonrpc.OpsService{
		Storage: &storer,
		Logger:  logger,
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
	t.Parallel()

	storer := storage.InMemory{}

	logger := log.NewNopLogger()
	svc := jsonrpc.OpsService{
		Storage: &storer,
		Logger:  logger,
	}

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

		err := svc.AddDatacenter(nil, &jsonrpc.AddDatacenterArgs{Datacenter: expected}, &reply)
		assert.EqualError(t, err, "AddDatacenter() error: datacenter with reference 1 already exists")
	})
}

func TestRemoveDatacenter(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	logger := log.NewNopLogger()
	svc := jsonrpc.OpsService{
		Storage: &storer,
		Logger:  logger,
	}

	expected := routing.Datacenter{
		ID:      crypto.HashID("local datacenter"),
		Name:    "local datacenter",
		Enabled: false,
		Location: routing.Location{
			Latitude:  70.5,
			Longitude: 120.5,
		},
	}

	t.Run("doesn't exist", func(t *testing.T) {
		var reply jsonrpc.RemoveDatacenterReply

		err := svc.RemoveDatacenter(nil, &jsonrpc.RemoveDatacenterArgs{Name: expected.Name}, &reply)
		assert.EqualError(t, err, fmt.Sprintf("RemoveDatacenter() error: datacenter with reference %d not found", expected.ID))
	})

	t.Run("remove", func(t *testing.T) {
		var addReply jsonrpc.AddDatacenterReply
		err := svc.AddDatacenter(nil, &jsonrpc.AddDatacenterArgs{Datacenter: expected}, &addReply)
		assert.NoError(t, err)

		var reply jsonrpc.RemoveDatacenterReply
		err = svc.RemoveDatacenter(nil, &jsonrpc.RemoveDatacenterArgs{Name: expected.Name}, &reply)
		assert.NoError(t, err)

		var datacentersReply jsonrpc.DatacentersReply
		err = svc.Datacenters(nil, &jsonrpc.DatacentersArgs{}, &datacentersReply)
		assert.NoError(t, err)

		assert.Len(t, datacentersReply.Datacenters, 0)
	})
}
*/

func TestNotifications(t *testing.T) {
	logger := log.NewNopLogger()

	var storer storage.Storer

	storer, err := storage.NewSQLite3(context.Background(), logger)
	assert.NoError(t, err)

	err = storer.AddCustomer(context.Background(), routing.Customer{Code: "local", Name: "Local"})
	assert.NoError(t, err)

	pubkey := make([]byte, 4)
	err = storer.AddBuyer(context.Background(), routing.Buyer{ID: 1, CompanyCode: "local", PublicKey: pubkey})
	assert.NoError(t, err)

	svc := jsonrpc.OpsService{
		Storage: storer,
		Logger:  logger,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("insufficient privileges - anonymous", func(t *testing.T) {
		var reply jsonrpc.NotificationsReply
		err := svc.Notifications(req, &jsonrpc.NotificationsArgs{}, &reply)
		assert.Error(t, err)
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	t.Run("success - no notifications", func(t *testing.T) {
		var reply jsonrpc.NotificationsReply
		err := svc.Notifications(req, &jsonrpc.NotificationsArgs{}, &reply)
		assert.NoError(t, err)

		assert.Len(t, reply.Notifications, 0)
	})

	err = storer.AddNotificationType(notifications.NotificationType{
		Name: "system",
	})
	assert.NoError(t, err)

	err = storer.AddNotificationType(notifications.NotificationType{
		Name: "analytics",
	})
	assert.NoError(t, err)

	err = storer.AddNotificationType(notifications.NotificationType{
		Name: "invoice",
	})
	assert.NoError(t, err)

	systemType, err := storer.NotificationTypeByName("system")
	assert.NoError(t, err)

	analyticsType, err := storer.NotificationTypeByName("analytics")
	assert.NoError(t, err)

	invoiceType, err := storer.NotificationTypeByName("invoice")
	assert.NoError(t, err)

	err = storer.AddNotificationPriority(notifications.NotificationPriority{
		Name:  "default",
		Color: jsonrpc.DEFAULT_COLOR,
	})
	assert.NoError(t, err)

	priority, err := storer.NotificationPriorityByName("default")
	assert.NoError(t, err)

	defaultSystemNotification := notifications.Notification{
		Timestamp:    time.Now(),
		Author:       "me",
		Title:        "Test system notification",
		Message:      "This is a test system notification",
		Type:         systemType,
		CustomerCode: "local",
		Priority:     priority,
		Public:       false,
		Paid:         false,
		Data:         "",
	}

	defaultAnalyticsNotification := defaultSystemNotification
	defaultAnalyticsNotification.Type = analyticsType
	defaultAnalyticsNotification.Title = "Test analytics notification"
	defaultAnalyticsNotification.Message = "This is a  test analytics notification"
	defaultAnalyticsNotification.CustomerCode = "test"

	defaultInvoiceNotification := defaultSystemNotification
	defaultInvoiceNotification.Type = invoiceType
	defaultInvoiceNotification.Title = "Test invoice notification"
	defaultInvoiceNotification.Message = "This is a  test invoice notification"

	err = storer.AddNotification(defaultSystemNotification)
	assert.NoError(t, err)

	err = storer.AddNotification(defaultAnalyticsNotification)
	assert.NoError(t, err)

	err = storer.AddNotification(defaultInvoiceNotification)
	assert.NoError(t, err)

	t.Run("success - all", func(t *testing.T) {
		var reply jsonrpc.NotificationsReply
		err := svc.Notifications(req, &jsonrpc.NotificationsArgs{}, &reply)
		assert.NoError(t, err)

		assert.Len(t, reply.Notifications, 3)
	})

	t.Run("success - sorted", func(t *testing.T) {
		var reply jsonrpc.NotificationsReply
		err := svc.Notifications(req, &jsonrpc.NotificationsArgs{CustomerCode: "test"}, &reply)
		assert.NoError(t, err)

		assert.Len(t, reply.Notifications, 1)
		assert.Equal(t, defaultAnalyticsNotification.Type.ID, reply.Notifications[0].TypeID)
		assert.Equal(t, defaultAnalyticsNotification.Title, reply.Notifications[0].Title)
		assert.Equal(t, defaultAnalyticsNotification.Message, reply.Notifications[0].Message)
	})
}

func TestAddNotification(t *testing.T) {
	logger := log.NewNopLogger()

	var storer storage.Storer

	storer, err := storage.NewSQLite3(context.Background(), logger)
	assert.NoError(t, err)

	err = storer.AddCustomer(context.Background(), routing.Customer{Code: "local", Name: "Local"})
	assert.NoError(t, err)

	pubkey := make([]byte, 4)
	err = storer.AddBuyer(context.Background(), routing.Buyer{ID: 1, CompanyCode: "local", PublicKey: pubkey})
	assert.NoError(t, err)

	svc := jsonrpc.OpsService{
		Storage: storer,
		Logger:  logger,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.AddNotificationReply
		err := svc.AddNotification(req, &jsonrpc.AddNotificationArgs{}, &reply)
		assert.Error(t, err)
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Admin",
	})
	reqContext = context.WithValue(reqContext, middleware.Keys.UserKey, &jwt.Token{
		Claims: jwt.MapClaims{
			"sub": "123456",
		},
	})
	req = req.WithContext(reqContext)

	err = storer.AddNotificationType(notifications.NotificationType{
		Name: "system",
	})
	assert.NoError(t, err)

	systemType, err := storer.NotificationTypeByName("system")
	assert.NoError(t, err)

	err = storer.AddNotificationPriority(notifications.NotificationPriority{
		Name:  "default",
		Color: jsonrpc.DEFAULT_COLOR,
	})
	assert.NoError(t, err)

	priority, err := storer.NotificationPriorityByName("default")
	assert.NoError(t, err)

	t.Run("success - no customer codes", func(t *testing.T) {
		var reply jsonrpc.AddNotificationReply
		args := jsonrpc.AddNotificationArgs{
			Title:         "Test system notification",
			Message:       "This is a test system notification",
			TypeID:        fmt.Sprintf("%016x", systemType.ID),
			CustomerCodes: []string{},
			PriorityID:    fmt.Sprintf("%016x", priority.ID),
			Public:        false,
			Paid:          false,
			Data:          "",
		}
		err := svc.AddNotification(req, &args, &reply)
		assert.NoError(t, err)

		allNotifications := storer.Notifications()

		assert.Len(t, allNotifications, 1)
		assert.Equal(t, args.Title, allNotifications[0].Title)
		assert.Equal(t, args.Message, allNotifications[0].Message)
		assert.Equal(t, args.TypeID, allNotifications[0].Type)
		assert.Equal(t, "", allNotifications[0].CustomerCode)
		assert.Equal(t, args.PriorityID, allNotifications[0].Priority)
		assert.Equal(t, args.Public, allNotifications[0].Public)
		assert.Equal(t, args.Paid, allNotifications[0].Paid)
		assert.Equal(t, args.Data, allNotifications[0].Data)
	})

	t.Run("success - 1 customer code", func(t *testing.T) {
		var reply jsonrpc.AddNotificationReply
		args := jsonrpc.AddNotificationArgs{
			Title:         "Test system notification",
			Message:       "This is a test system notification",
			TypeID:        fmt.Sprintf("%016x", systemType.ID),
			CustomerCodes: []string{"test"},
			PriorityID:    fmt.Sprintf("%016x", priority.ID),
			Public:        false,
			Paid:          false,
			Data:          "",
		}
		err := svc.AddNotification(req, &args, &reply)
		assert.NoError(t, err)

		allNotifications := storer.Notifications()

		assert.Len(t, allNotifications, 1)
		assert.Equal(t, args.Title, allNotifications[0].Title)
		assert.Equal(t, args.Message, allNotifications[0].Message)
		assert.Equal(t, args.TypeID, allNotifications[0].Type)
		assert.Equal(t, "test", allNotifications[0].CustomerCode)
		assert.Equal(t, args.PriorityID, allNotifications[0].Priority)
		assert.Equal(t, args.Public, allNotifications[0].Public)
		assert.Equal(t, args.Paid, allNotifications[0].Paid)
		assert.Equal(t, args.Data, allNotifications[0].Data)
	})

	t.Run("success - multiple customer codes", func(t *testing.T) {
		var reply jsonrpc.AddNotificationReply
		args := jsonrpc.AddNotificationArgs{
			Title:         "Test system notification",
			Message:       "This is a test system notification",
			TypeID:        fmt.Sprintf("%016x", systemType.ID),
			CustomerCodes: []string{"test", "testing", "tested"},
			PriorityID:    fmt.Sprintf("%016x", priority.ID),
			Public:        false,
			Paid:          false,
			Data:          "",
		}
		err := svc.AddNotification(req, &args, &reply)
		assert.NoError(t, err)

		allNotifications := storer.Notifications()

		assert.Len(t, allNotifications, 3)

		found := false
		for _, dbNotification := range allNotifications {
			if dbNotification.CustomerCode == args.CustomerCodes[0] || dbNotification.CustomerCode == args.CustomerCodes[1] || dbNotification.CustomerCode == args.CustomerCodes[2] {
				found = true
			} else {
				found = false
			}
			assert.True(t, found)
			found = false
		}
	})
}

func TestUpdateNotification(t *testing.T) {
	logger := log.NewNopLogger()

	var storer storage.Storer

	storer, err := storage.NewSQLite3(context.Background(), logger)
	assert.NoError(t, err)

	err = storer.AddCustomer(context.Background(), routing.Customer{Code: "local", Name: "Local"})
	assert.NoError(t, err)

	pubkey := make([]byte, 4)
	err = storer.AddBuyer(context.Background(), routing.Buyer{ID: 1, CompanyCode: "local", PublicKey: pubkey})
	assert.NoError(t, err)

	svc := jsonrpc.OpsService{
		Storage: storer,
		Logger:  logger,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	err = storer.AddNotificationType(notifications.NotificationType{
		Name: "system",
	})
	assert.NoError(t, err)

	systemType, err := storer.NotificationTypeByName("system")
	assert.NoError(t, err)

	err = storer.AddNotificationPriority(notifications.NotificationPriority{
		Name:  "default",
		Color: jsonrpc.DEFAULT_COLOR,
	})
	assert.NoError(t, err)

	defaultPriority, err := storer.NotificationPriorityByName("default")
	assert.NoError(t, err)

	analyticsType, err := storer.NotificationTypeByName("analytics")
	assert.NoError(t, err)

	err = storer.AddNotificationPriority(notifications.NotificationPriority{
		Name:  "urgent",
		Color: jsonrpc.DEFAULT_COLOR,
	})
	assert.NoError(t, err)

	urgentPriority, err := storer.NotificationPriorityByName("urgent")
	assert.NoError(t, err)

	oldNotification := notifications.Notification{
		Timestamp:    time.Now(),
		Title:        "Test notification",
		Message:      "Test notification message",
		Author:       "me",
		Type:         systemType,
		CustomerCode: "",
		Priority:     defaultPriority,
		Public:       false,
		Paid:         false,
		Data:         "",
	}

	err = storer.AddNotification(oldNotification)
	assert.NoError(t, err)

	oldNotificationID := storer.Notifications()[0].ID

	t.Run("insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.UpdateNotificationReply
		err := svc.UpdateNotification(req, &jsonrpc.UpdateNotificationArgs{ID: ""}, &reply)
		assert.Error(t, err)
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	t.Run("no id", func(t *testing.T) {
		var reply jsonrpc.UpdateNotificationReply

		err := svc.UpdateNotification(req, &jsonrpc.UpdateNotificationArgs{ID: ""}, &reply)
		assert.Error(t, err)
	})

	t.Run("success - no customer codes", func(t *testing.T) {
		var reply jsonrpc.UpdateNotificationReply
		args := jsonrpc.UpdateNotificationArgs{
			ID:    fmt.Sprintf("%016x", oldNotificationID),
			Field: "Title",
			Value: "Updated Test Notification",
		}
		err := svc.UpdateNotification(req, &args, &reply)
		assert.NoError(t, err)

		updatedNotification, err := storer.NotificationByID(oldNotificationID)
		assert.NoError(t, err)

		assert.NotEqual(t, oldNotification.Title, updatedNotification.Title)
		assert.Equal(t, "Updated Test Notification", updatedNotification.Title)

		args = jsonrpc.UpdateNotificationArgs{
			ID:    fmt.Sprintf("%016x", oldNotificationID),
			Field: "Message",
			Value: "Updated test notification message",
		}
		err = svc.UpdateNotification(req, &args, &reply)
		assert.NoError(t, err)

		updatedNotification, err = storer.NotificationByID(oldNotificationID)
		assert.NoError(t, err)

		assert.NotEqual(t, oldNotification.Message, updatedNotification.Message)
		assert.Equal(t, "Updated test notification message", updatedNotification.Message)

		args = jsonrpc.UpdateNotificationArgs{
			ID:    fmt.Sprintf("%016x", oldNotificationID),
			Field: "Author",
			Value: "you",
		}
		err = svc.UpdateNotification(req, &args, &reply)
		assert.NoError(t, err)

		updatedNotification, err = storer.NotificationByID(oldNotificationID)
		assert.NoError(t, err)

		assert.NotEqual(t, oldNotification.Author, updatedNotification.Author)
		assert.Equal(t, "you", updatedNotification.Author)

		args = jsonrpc.UpdateNotificationArgs{
			ID:    fmt.Sprintf("%016x", oldNotificationID),
			Field: "CustomerCode",
			Value: "test",
		}
		err = svc.UpdateNotification(req, &args, &reply)
		assert.NoError(t, err)

		updatedNotification, err = storer.NotificationByID(oldNotificationID)
		assert.NoError(t, err)

		assert.NotEqual(t, oldNotification.CustomerCode, updatedNotification.CustomerCode)
		assert.Equal(t, "test", updatedNotification.CustomerCode)

		args = jsonrpc.UpdateNotificationArgs{
			ID:    fmt.Sprintf("%016x", oldNotificationID),
			Field: "Paid",
			Value: "true",
		}
		err = svc.UpdateNotification(req, &args, &reply)
		assert.NoError(t, err)

		updatedNotification, err = storer.NotificationByID(oldNotificationID)
		assert.NoError(t, err)

		assert.NotEqual(t, oldNotification.Paid, updatedNotification.Paid)
		assert.True(t, updatedNotification.Paid)

		args = jsonrpc.UpdateNotificationArgs{
			ID:    fmt.Sprintf("%016x", oldNotificationID),
			Field: "Public",
			Value: "true",
		}
		err = svc.UpdateNotification(req, &args, &reply)
		assert.NoError(t, err)

		updatedNotification, err = storer.NotificationByID(oldNotificationID)
		assert.NoError(t, err)

		assert.NotEqual(t, oldNotification.Public, updatedNotification.Public)
		assert.True(t, updatedNotification.Public)

		args = jsonrpc.UpdateNotificationArgs{
			ID:    fmt.Sprintf("%016x", oldNotificationID),
			Field: "Type",
			Value: fmt.Sprintf("%016x", analyticsType.ID),
		}
		err = svc.UpdateNotification(req, &args, &reply)
		assert.NoError(t, err)

		updatedNotification, err = storer.NotificationByID(oldNotificationID)
		assert.NoError(t, err)

		assert.NotEqual(t, oldNotification.Type.ID, updatedNotification.Type.ID)
		assert.Equal(t, analyticsType.ID, updatedNotification.Type.ID)
		assert.Equal(t, analyticsType.Name, updatedNotification.Type.Name)

		args = jsonrpc.UpdateNotificationArgs{
			ID:    fmt.Sprintf("%016x", oldNotificationID),
			Field: "Priority",
			Value: fmt.Sprintf("%016x", urgentPriority.ID),
		}
		err = svc.UpdateNotification(req, &args, &reply)
		assert.NoError(t, err)

		updatedNotification, err = storer.NotificationByID(oldNotificationID)
		assert.NoError(t, err)

		assert.NotEqual(t, oldNotification.Priority.ID, updatedNotification.Priority.ID)
		assert.Equal(t, urgentPriority.ID, updatedNotification.Priority.ID)
		assert.Equal(t, urgentPriority.Name, updatedNotification.Priority.Name)
		assert.Equal(t, urgentPriority.Color, updatedNotification.Priority.Color)
	})
}

func TestRemoveNotification(t *testing.T) {
	logger := log.NewNopLogger()

	var storer storage.Storer

	storer, err := storage.NewSQLite3(context.Background(), logger)
	assert.NoError(t, err)

	err = storer.AddCustomer(context.Background(), routing.Customer{Code: "local", Name: "Local"})
	assert.NoError(t, err)

	pubkey := make([]byte, 4)
	err = storer.AddBuyer(context.Background(), routing.Buyer{ID: 1, CompanyCode: "local", PublicKey: pubkey})
	assert.NoError(t, err)

	svc := jsonrpc.OpsService{
		Storage: storer,
		Logger:  logger,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	err = storer.AddNotificationType(notifications.NotificationType{
		Name: "system",
	})
	assert.NoError(t, err)

	systemType, err := storer.NotificationTypeByName("system")
	assert.NoError(t, err)

	err = storer.AddNotificationPriority(notifications.NotificationPriority{
		Name:  "default",
		Color: jsonrpc.DEFAULT_COLOR,
	})
	assert.NoError(t, err)

	defaultPriority, err := storer.NotificationPriorityByName("default")
	assert.NoError(t, err)

	oldNotification := notifications.Notification{
		Timestamp:    time.Now(),
		Title:        "Test notification",
		Message:      "Test notification message",
		Author:       "me",
		Type:         systemType,
		CustomerCode: "",
		Priority:     defaultPriority,
		Public:       false,
		Paid:         false,
		Data:         "",
	}

	err = storer.AddNotification(oldNotification)
	assert.NoError(t, err)

	oldNotificationID := storer.Notifications()[0].ID

	t.Run("insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.RemoveNotificationReply
		err := svc.RemoveNotification(req, &jsonrpc.RemoveNotificationArgs{ID: ""}, &reply)
		assert.Error(t, err)
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	t.Run("no id", func(t *testing.T) {
		var reply jsonrpc.RemoveNotificationReply
		err := svc.RemoveNotification(req, &jsonrpc.RemoveNotificationArgs{ID: ""}, &reply)
		assert.NoError(t, err)
	})

	t.Run("success", func(t *testing.T) {
		allNotifications := storer.Notifications()
		assert.Len(t, allNotifications, 1)

		var reply jsonrpc.RemoveNotificationReply
		err := svc.RemoveNotification(req, &jsonrpc.RemoveNotificationArgs{ID: fmt.Sprintf("%016x", oldNotificationID)}, &reply)
		assert.NoError(t, err)

		allNotifications = storer.Notifications()
		assert.Len(t, allNotifications, 0)

		_, err = storer.NotificationByID(oldNotificationID)
		assert.Error(t, err)
	})
}

func TestNotificationTypes(t *testing.T) {
	logger := log.NewNopLogger()

	var storer storage.Storer

	storer, err := storage.NewSQLite3(context.Background(), logger)
	assert.NoError(t, err)

	svc := jsonrpc.OpsService{
		Storage: storer,
		Logger:  logger,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("insufficient privileges - anonymous", func(t *testing.T) {
		var reply jsonrpc.NotificationTypesReply
		err := svc.NotificationTypes(req, &jsonrpc.NotificationTypesArgs{}, &reply)
		assert.Error(t, err)
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	t.Run("success - no types", func(t *testing.T) {
		var reply jsonrpc.NotificationTypesReply
		err := svc.NotificationTypes(req, &jsonrpc.NotificationTypesArgs{}, &reply)
		assert.NoError(t, err)

		assert.Len(t, reply.NotificationTypes, 0)
	})

	err = storer.AddNotificationType(notifications.NotificationType{
		Name: "system",
	})
	assert.NoError(t, err)

	err = storer.AddNotificationType(notifications.NotificationType{
		Name: "analytics",
	})
	assert.NoError(t, err)

	err = storer.AddNotificationType(notifications.NotificationType{
		Name: "invoice",
	})

	t.Run("success - all", func(t *testing.T) {
		var reply jsonrpc.NotificationTypesReply
		err := svc.NotificationTypes(req, &jsonrpc.NotificationTypesArgs{}, &reply)
		assert.NoError(t, err)

		assert.Len(t, reply.NotificationTypes, 3)
	})
}

func TestAddNotificationType(t *testing.T) {
	logger := log.NewNopLogger()

	var storer storage.Storer

	storer, err := storage.NewSQLite3(context.Background(), logger)
	assert.NoError(t, err)

	svc := jsonrpc.OpsService{
		Storage: storer,
		Logger:  logger,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.AddNotificationTypeReply
		err := svc.AddNotificationType(req, &jsonrpc.AddNotificationTypeArgs{}, &reply)
		assert.Error(t, err)
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	t.Run("no name", func(t *testing.T) {
		var reply jsonrpc.AddNotificationTypeReply
		err := svc.AddNotificationType(req, &jsonrpc.AddNotificationTypeArgs{}, &reply)
		assert.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		var reply jsonrpc.AddNotificationTypeReply
		args := jsonrpc.AddNotificationTypeArgs{
			Name: "test",
		}
		err := svc.AddNotificationType(req, &args, &reply)
		assert.NoError(t, err)

		allNotificationTypes := storer.NotificationTypes()
		testType, err := storer.NotificationTypeByName("test")
		assert.NoError(t, err)

		assert.Len(t, allNotificationTypes, 1)
		assert.Equal(t, testType.ID, allNotificationTypes[0].ID)
		assert.Equal(t, testType.Name, allNotificationTypes[0].Name)
		assert.Equal(t, args.Name, allNotificationTypes[0].Name)
	})
}

func TestUpdateNotificationType(t *testing.T) {
	logger := log.NewNopLogger()

	var storer storage.Storer

	storer, err := storage.NewSQLite3(context.Background(), logger)
	assert.NoError(t, err)

	svc := jsonrpc.OpsService{
		Storage: storer,
		Logger:  logger,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	err = storer.AddNotificationType(notifications.NotificationType{
		Name: "system",
	})
	assert.NoError(t, err)

	systemType, err := storer.NotificationTypeByName("system")
	assert.NoError(t, err)

	t.Run("insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.UpdateNotificationTypeReply
		err := svc.UpdateNotificationType(req, &jsonrpc.UpdateNotificationTypeArgs{ID: ""}, &reply)
		assert.Error(t, err)
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	t.Run("no id", func(t *testing.T) {
		var reply jsonrpc.UpdateNotificationTypeReply

		err := svc.UpdateNotificationType(req, &jsonrpc.UpdateNotificationTypeArgs{ID: ""}, &reply)
		assert.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		var reply jsonrpc.UpdateNotificationTypeReply
		args := jsonrpc.UpdateNotificationTypeArgs{
			ID:    fmt.Sprintf("%016x", systemType.ID),
			Field: "Name",
			Value: "updated system",
		}
		err := svc.UpdateNotificationType(req, &args, &reply)
		assert.NoError(t, err)

		updatedNotificationType, err := storer.NotificationTypeByID(systemType.ID)
		assert.NoError(t, err)

		assert.NotEqual(t, systemType.Name, updatedNotificationType.Name)
		assert.Equal(t, "updated system", updatedNotificationType.Name)
	})
}

func TestRemoveNotificationType(t *testing.T) {
	logger := log.NewNopLogger()

	var storer storage.Storer

	storer, err := storage.NewSQLite3(context.Background(), logger)
	assert.NoError(t, err)

	svc := jsonrpc.OpsService{
		Storage: storer,
		Logger:  logger,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	err = storer.AddNotificationType(notifications.NotificationType{
		Name: "system",
	})
	assert.NoError(t, err)

	systemType, err := storer.NotificationTypeByName("system")
	assert.NoError(t, err)

	t.Run("insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.RemoveNotificationTypeReply
		err := svc.RemoveNotificationType(req, &jsonrpc.RemoveNotificationTypeArgs{ID: ""}, &reply)
		assert.Error(t, err)
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	t.Run("no id", func(t *testing.T) {
		var reply jsonrpc.RemoveNotificationTypeReply
		err := svc.RemoveNotificationType(req, &jsonrpc.RemoveNotificationTypeArgs{ID: ""}, &reply)
		assert.NoError(t, err)
	})

	t.Run("success", func(t *testing.T) {
		allNotifications := storer.NotificationTypes()
		assert.Len(t, allNotifications, 1)

		var reply jsonrpc.RemoveNotificationTypeReply
		err := svc.RemoveNotificationType(req, &jsonrpc.RemoveNotificationTypeArgs{ID: fmt.Sprintf("%016x", systemType.ID)}, &reply)
		assert.NoError(t, err)

		allNotificationTypes := storer.NotificationTypes()
		assert.Len(t, allNotificationTypes, 0)

		_, err = storer.NotificationByID(systemType.ID)
		assert.Error(t, err)
	})
}

func TestNotificationPriorities(t *testing.T) {
	logger := log.NewNopLogger()

	var storer storage.Storer

	storer, err := storage.NewSQLite3(context.Background(), logger)
	assert.NoError(t, err)

	svc := jsonrpc.OpsService{
		Storage: storer,
		Logger:  logger,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("insufficient privileges - anonymous", func(t *testing.T) {
		var reply jsonrpc.NotificationPrioritiesReply
		err := svc.NotificationPriorities(req, &jsonrpc.NotificationPrioritiesArgs{}, &reply)
		assert.Error(t, err)
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	t.Run("success - no priorities", func(t *testing.T) {
		var reply jsonrpc.NotificationPrioritiesReply
		err := svc.NotificationPriorities(req, &jsonrpc.NotificationPrioritiesArgs{}, &reply)
		assert.NoError(t, err)

		assert.Len(t, reply.NotificationPriorities, 0)
	})

	err = storer.AddNotificationPriority(notifications.NotificationPriority{
		Name: "default",
	})
	assert.NoError(t, err)

	err = storer.AddNotificationPriority(notifications.NotificationPriority{
		Name: "urgent",
	})
	assert.NoError(t, err)

	err = storer.AddNotificationPriority(notifications.NotificationPriority{
		Name: "warning",
	})

	t.Run("success - all", func(t *testing.T) {
		var reply jsonrpc.NotificationPrioritiesReply
		err := svc.NotificationPriorities(req, &jsonrpc.NotificationPrioritiesArgs{}, &reply)
		assert.NoError(t, err)

		assert.Len(t, reply.NotificationPriorities, 3)
	})
}

func TestAddNotificationPriority(t *testing.T) {
	logger := log.NewNopLogger()

	var storer storage.Storer

	storer, err := storage.NewSQLite3(context.Background(), logger)
	assert.NoError(t, err)

	svc := jsonrpc.OpsService{
		Storage: storer,
		Logger:  logger,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	t.Run("insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.AddNotificationPriorityReply
		err := svc.AddNotificationPriority(req, &jsonrpc.AddNotificationPriorityArgs{}, &reply)
		assert.Error(t, err)
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	t.Run("no name", func(t *testing.T) {
		var reply jsonrpc.AddNotificationPriorityReply
		err := svc.AddNotificationPriority(req, &jsonrpc.AddNotificationPriorityArgs{}, &reply)
		assert.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		var reply jsonrpc.AddNotificationPriorityReply
		args := jsonrpc.AddNotificationPriorityArgs{
			Name:  "test",
			Color: jsonrpc.DEFAULT_COLOR,
		}
		err := svc.AddNotificationPriority(req, &args, &reply)
		assert.NoError(t, err)

		allNotificationPriorities := storer.NotificationPriorities()
		testPriority, err := storer.NotificationPriorityByName("test")
		assert.NoError(t, err)

		assert.Len(t, allNotificationPriorities, 1)
		assert.Equal(t, testPriority.ID, allNotificationPriorities[0].ID)
		assert.Equal(t, testPriority.Name, allNotificationPriorities[0].Name)
		assert.Equal(t, testPriority.Color, allNotificationPriorities[0].Color)
		assert.Equal(t, args.Name, allNotificationPriorities[0].Name)
		assert.Equal(t, args.Color, allNotificationPriorities[0].Color)
	})
}

func TestUpdateNotificationPriority(t *testing.T) {
	logger := log.NewNopLogger()

	var storer storage.Storer

	storer, err := storage.NewSQLite3(context.Background(), logger)
	assert.NoError(t, err)

	svc := jsonrpc.OpsService{
		Storage: storer,
		Logger:  logger,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	err = storer.AddNotificationPriority(notifications.NotificationPriority{
		Name: "default",
	})
	assert.NoError(t, err)

	defaultPriority, err := storer.NotificationPriorityByName("default")
	assert.NoError(t, err)

	t.Run("insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.UpdateNotificationPriorityReply
		err := svc.UpdateNotificationPriority(req, &jsonrpc.UpdateNotificationPriorityArgs{ID: ""}, &reply)
		assert.Error(t, err)
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	t.Run("no id", func(t *testing.T) {
		var reply jsonrpc.UpdateNotificationPriorityReply

		err := svc.UpdateNotificationPriority(req, &jsonrpc.UpdateNotificationPriorityArgs{ID: ""}, &reply)
		assert.Error(t, err)
	})

	t.Run("success", func(t *testing.T) {
		var reply jsonrpc.UpdateNotificationPriorityReply
		args := jsonrpc.UpdateNotificationPriorityArgs{
			ID:    fmt.Sprintf("%016x", defaultPriority.ID),
			Field: "Name",
			Value: "updated system",
		}
		err := svc.UpdateNotificationPriority(req, &args, &reply)
		assert.NoError(t, err)

		updatedNotificationPriority, err := storer.NotificationPriorityByID(defaultPriority.ID)
		assert.NoError(t, err)

		assert.NotEqual(t, defaultPriority.Name, updatedNotificationPriority.Name)
		assert.Equal(t, "updated system", updatedNotificationPriority.Name)
		assert.Equal(t, defaultPriority.Color, updatedNotificationPriority.Color)

		args = jsonrpc.UpdateNotificationPriorityArgs{
			ID:    fmt.Sprintf("%016x", defaultPriority.ID),
			Field: "Color",
			Value: "12354642346",
		}
		err = svc.UpdateNotificationPriority(req, &args, &reply)
		assert.NoError(t, err)

		updatedNotificationPriority, err = storer.NotificationPriorityByID(defaultPriority.ID)
		assert.NoError(t, err)

		assert.NotEqual(t, defaultPriority.Color, updatedNotificationPriority.Color)
		assert.Equal(t, "updated system", updatedNotificationPriority.Name)
		assert.Equal(t, int64(12354642346), updatedNotificationPriority.Color)
	})
}

func TestRemoveNotificationPriority(t *testing.T) {
	logger := log.NewNopLogger()

	var storer storage.Storer

	storer, err := storage.NewSQLite3(context.Background(), logger)
	assert.NoError(t, err)

	svc := jsonrpc.OpsService{
		Storage: storer,
		Logger:  logger,
	}

	req := httptest.NewRequest(http.MethodGet, "/", nil)

	err = storer.AddNotificationPriority(notifications.NotificationPriority{
		Name: "default",
	})
	assert.NoError(t, err)

	defaultPriority, err := storer.NotificationPriorityByName("default")
	assert.NoError(t, err)

	t.Run("insufficient privileges", func(t *testing.T) {
		var reply jsonrpc.RemoveNotificationPriorityReply
		err := svc.RemoveNotificationPriority(req, &jsonrpc.RemoveNotificationPriorityArgs{ID: ""}, &reply)
		assert.Error(t, err)
	})

	reqContext := req.Context()
	reqContext = context.WithValue(reqContext, middleware.Keys.RolesKey, []string{
		"Admin",
	})
	req = req.WithContext(reqContext)

	t.Run("no id", func(t *testing.T) {
		var reply jsonrpc.RemoveNotificationPriorityReply
		err := svc.RemoveNotificationPriority(req, &jsonrpc.RemoveNotificationPriorityArgs{ID: ""}, &reply)
		assert.NoError(t, err)
	})

	t.Run("success", func(t *testing.T) {
		allNotifications := storer.NotificationPriorities()
		assert.Len(t, allNotifications, 1)

		var reply jsonrpc.RemoveNotificationPriorityReply
		err := svc.RemoveNotificationPriority(req, &jsonrpc.RemoveNotificationPriorityArgs{ID: fmt.Sprintf("%016x", defaultPriority.ID)}, &reply)
		assert.NoError(t, err)

		allNotificationPriorities := storer.NotificationPriorities()
		assert.Len(t, allNotificationPriorities, 0)

		_, err = storer.NotificationByID(defaultPriority.ID)
		assert.Error(t, err)
	})
}
