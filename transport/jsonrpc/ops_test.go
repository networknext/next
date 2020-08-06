package jsonrpc_test

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
		RouteShader: routing.DefaultRouteShader,
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
		RouteShader: routing.DefaultRouteShader,
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

func TestRouteShader(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	logger := log.NewNopLogger()
	svc := jsonrpc.OpsService{
		Storage: &storer,
		Logger:  logger,
	}

	t.Run("doesn't exist", func(t *testing.T) {
		var reply jsonrpc.RouteShaderReply

		err := svc.RouteShader(nil, &jsonrpc.RouteShaderArgs{BuyerID: "0"}, &reply)
		assert.EqualError(t, err, "buyer with reference 0 not found")
	})

	t.Run("list", func(t *testing.T) {
		storer.AddBuyer(context.Background(), routing.Buyer{ID: 0, Name: "local.local.1", RouteShader: routing.DefaultRouteShader})

		var reply jsonrpc.RouteShaderReply
		err := svc.RouteShader(nil, &jsonrpc.RouteShaderArgs{BuyerID: "0"}, &reply)
		assert.NoError(t, err)

		assert.Equal(t, reply.RoutingRuleSettings[0].EnvelopeKbpsUp, routing.DefaultRouteShader.EnvelopeKbpsUp)
		assert.Equal(t, reply.RoutingRuleSettings[0].EnvelopeKbpsDown, routing.DefaultRouteShader.EnvelopeKbpsDown)
		assert.Equal(t, reply.RoutingRuleSettings[0].Mode, routing.DefaultRouteShader.Mode)
		assert.Equal(t, reply.RoutingRuleSettings[0].MaxCentsPerGB, routing.DefaultRouteShader.MaxCentsPerGB)
		assert.Equal(t, reply.RoutingRuleSettings[0].RTTEpsilon, routing.DefaultRouteShader.RTTEpsilon)
		assert.Equal(t, reply.RoutingRuleSettings[0].RTTThreshold, routing.DefaultRouteShader.RTTThreshold)
		assert.Equal(t, reply.RoutingRuleSettings[0].RTTHysteresis, routing.DefaultRouteShader.RTTHysteresis)
		assert.Equal(t, reply.RoutingRuleSettings[0].RTTVeto, routing.DefaultRouteShader.RTTVeto)
		assert.Equal(t, reply.RoutingRuleSettings[0].EnableYouOnlyLiveOnce, routing.DefaultRouteShader.EnableYouOnlyLiveOnce)
		assert.Equal(t, reply.RoutingRuleSettings[0].EnablePacketLossSafety, routing.DefaultRouteShader.EnablePacketLossSafety)
		assert.Equal(t, reply.RoutingRuleSettings[0].EnableMultipathForPacketLoss, routing.DefaultRouteShader.EnableMultipathForPacketLoss)
		assert.Equal(t, reply.RoutingRuleSettings[0].EnableMultipathForJitter, routing.DefaultRouteShader.EnableMultipathForJitter)
		assert.Equal(t, reply.RoutingRuleSettings[0].EnableMultipathForRTT, routing.DefaultRouteShader.EnableMultipathForRTT)
		assert.Equal(t, reply.RoutingRuleSettings[0].EnableABTest, routing.DefaultRouteShader.EnableABTest)
		assert.Equal(t, reply.RoutingRuleSettings[0].EnableTryBeforeYouBuy, routing.DefaultRouteShader.EnableTryBeforeYouBuy)
		assert.Equal(t, reply.RoutingRuleSettings[0].TryBeforeYouBuyMaxSlices, routing.DefaultRouteShader.TryBeforeYouBuyMaxSlices)
	})
}

func TestSetRouteShader(t *testing.T) {
	t.Parallel()

	storer := storage.InMemory{}

	logger := log.NewNopLogger()
	svc := jsonrpc.OpsService{
		Storage: &storer,
		Logger:  logger,
	}

	t.Run("doesn't exist", func(t *testing.T) {
		var reply jsonrpc.SetRouteShaderReply

		err := svc.SetRouteShader(nil, &jsonrpc.SetRouteShaderArgs{BuyerID: "0", RouteShader: routing.LocalRouteShader}, &reply)
		assert.EqualError(t, err, "SetRouteShader() Storage.Buyer error: buyer with reference 0 not found")
	})

	t.Run("set", func(t *testing.T) {
		storer.AddBuyer(context.Background(), routing.Buyer{ID: 1, Name: "local.local.1", RouteShader: routing.DefaultRouteShader})

		var reply jsonrpc.SetRouteShaderReply
		err := svc.SetRouteShader(nil, &jsonrpc.SetRouteShaderArgs{BuyerID: "1", RouteShader: routing.LocalRouteShader}, &reply)
		assert.NoError(t, err)

		var rrsReply jsonrpc.RouteShaderReply
		err = svc.RouteShader(nil, &jsonrpc.RouteShaderArgs{BuyerID: "1"}, &rrsReply)
		assert.NoError(t, err)

		assert.Equal(t, rrsReply.RoutingRuleSettings[0].EnvelopeKbpsUp, routing.LocalRouteShader.EnvelopeKbpsUp)
		assert.Equal(t, rrsReply.RoutingRuleSettings[0].EnvelopeKbpsDown, routing.LocalRouteShader.EnvelopeKbpsDown)
		assert.Equal(t, rrsReply.RoutingRuleSettings[0].Mode, routing.LocalRouteShader.Mode)
		assert.Equal(t, rrsReply.RoutingRuleSettings[0].MaxCentsPerGB, routing.LocalRouteShader.MaxCentsPerGB)
		assert.Equal(t, rrsReply.RoutingRuleSettings[0].RTTEpsilon, routing.LocalRouteShader.RTTEpsilon)
		assert.Equal(t, rrsReply.RoutingRuleSettings[0].RTTThreshold, routing.LocalRouteShader.RTTThreshold)
		assert.Equal(t, rrsReply.RoutingRuleSettings[0].RTTHysteresis, routing.LocalRouteShader.RTTHysteresis)
		assert.Equal(t, rrsReply.RoutingRuleSettings[0].RTTVeto, routing.LocalRouteShader.RTTVeto)
		assert.Equal(t, rrsReply.RoutingRuleSettings[0].EnableYouOnlyLiveOnce, routing.LocalRouteShader.EnableYouOnlyLiveOnce)
		assert.Equal(t, rrsReply.RoutingRuleSettings[0].EnablePacketLossSafety, routing.LocalRouteShader.EnablePacketLossSafety)
		assert.Equal(t, rrsReply.RoutingRuleSettings[0].EnableMultipathForPacketLoss, routing.LocalRouteShader.EnableMultipathForPacketLoss)
		assert.Equal(t, rrsReply.RoutingRuleSettings[0].EnableMultipathForJitter, routing.LocalRouteShader.EnableMultipathForJitter)
		assert.Equal(t, rrsReply.RoutingRuleSettings[0].EnableMultipathForRTT, routing.LocalRouteShader.EnableMultipathForRTT)
		assert.Equal(t, rrsReply.RoutingRuleSettings[0].EnableABTest, routing.LocalRouteShader.EnableABTest)
		assert.Equal(t, rrsReply.RoutingRuleSettings[0].EnableTryBeforeYouBuy, routing.DefaultRouteShader.EnableTryBeforeYouBuy)
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
