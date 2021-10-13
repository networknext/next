package transport

// func testGatewayHandlerConfig(storer storage.Storer) *GatewayHandlerConfig {

// 	return &GatewayHandlerConfig{
// 		InitMetrics:           &metrics.EmptyRelayInitMetrics,
// 		UpdateMetrics:         &metrics.EmptyRelayUpdateMetrics,
// 		Storer:                storer,
// 		RouterPrivateKey:      []byte{},
// 		Publishers:            []pubsub.Publisher{},
// 		RelayBackendAddresses: []string{},
// 		NRBNoInit:             false,
// 		NRBHTTP:               false,
// 	}
// }

// func testMetric(t *testing.T) metrics.Counter {
// 	localMetrics := metrics.LocalHandler{}
// 	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
// 	assert.NoError(t, err)
// 	return metric
// }

// func pingRelayGatewayUpdate(t *testing.T, contentType string, body []byte, handlerConfig *GatewayHandlerConfig) *httptest.ResponseRecorder {
// 	recorder := httptest.NewRecorder()
// 	request, err := http.NewRequest("POST", "/relay_update", bytes.NewBuffer(body))
// 	assert.NoError(t, err)
// 	request.Header.Add("Content-Type", contentType)

// 	updateChan := make(chan []byte, 1)

// 	handler := GatewayRelayUpdateHandlerFunc(log.NewNopLogger(), log.NewNopLogger(), handlerConfig, updateChan)
// 	handler(recorder, request)
// 	return recorder
// }

// func relayGatewayErrorAssertions(t *testing.T, recorder *httptest.ResponseRecorder, expectedCode int, errMetric metrics.Counter) {
// 	assert.Equal(t, expectedCode, recorder.Code)
// 	assert.Equal(t, 1.0, errMetric.ValueReset())
// }

// func relayGatewayUpdateShutdownAssertions(t *testing.T, recorder *httptest.ResponseRecorder, handlerConfig *GatewayHandlerConfig, addr string) {
// 	if recorder.Code != 200 {
// 		body, err := ioutil.ReadAll(recorder.Body)
// 		assert.Nil(t, err)
// 		fmt.Println(string(body))
// 	}

// 	relay, err := handlerConfig.Storer.Relay(crypto.HashID(addr))
// 	assert.NoError(t, err)
// 	assert.Equal(t, routing.RelayStateMaintenance, relay.State)

// 	errMetricsStruct := reflect.ValueOf(handlerConfig.UpdateMetrics.ErrorMetrics)
// 	for i := 0; i < errMetricsStruct.NumField(); i++ {
// 		if errMetricsStruct.Field(i).CanInterface() {
// 			assert.Equal(t, 0.0, errMetricsStruct.Field(i).Interface().(metrics.Counter).ValueReset())
// 		}
// 	}
// }

// func relayGatewayUpdateSuccessAssertions(t *testing.T, recorder *httptest.ResponseRecorder, expectedContentType string, handlerConfig *GatewayHandlerConfig, statIps []string, addr string) {
// 	assert.Equal(t, http.StatusOK, recorder.Code)

// 	// response assertions
// 	header := recorder.Header()
// 	contentType, ok := header["Content-Type"]
// 	assert.True(t, ok)
// 	if assert.NotNil(t, contentType) && assert.Len(t, contentType, 1) {
// 		assert.Equal(t, expectedContentType, contentType[0])
// 	}

// 	body := recorder.Body.Bytes()

// 	var response RelayUpdateResponse
// 	switch expectedContentType {
// 	case "application/octet-stream":
// 		err := response.UnmarshalBinary(body)
// 		assert.NoError(t, err)
// 	default:
// 		assert.FailNow(t, "Invalid expected content type")
// 	}

// 	assert.Equal(t, len(statIps), len(response.RelaysToPing))

// 	relaysToPingIDs := make([]uint64, 0)
// 	relaysToPingAddrs := make([]string, 0)

// 	for _, data := range response.RelaysToPing {
// 		relaysToPingIDs = append(relaysToPingIDs, data.ID)
// 		relaysToPingAddrs = append(relaysToPingAddrs, data.Address)
// 	}

// 	for _, addr := range statIps {
// 		id := crypto.HashID(addr)
// 		assert.Contains(t, relaysToPingIDs, id)
// 		assert.Contains(t, relaysToPingAddrs, addr)
// 	}

// 	assert.NotContains(t, relaysToPingIDs, crypto.HashID(addr))
// 	assert.NotContains(t, relaysToPingAddrs, addr)

// 	relay, err := handlerConfig.Storer.Relay(crypto.HashID(addr))
// 	assert.NoError(t, err)

// 	assert.Equal(t, routing.RelayStateEnabled, relay.State)

// 	errMetricsStruct := reflect.ValueOf(handlerConfig.UpdateMetrics.ErrorMetrics)
// 	for i := 0; i < errMetricsStruct.NumField(); i++ {
// 		if errMetricsStruct.Field(i).CanInterface() {
// 			assert.Equal(t, 0.0, errMetricsStruct.Field(i).Interface().(metrics.Counter).ValueReset())
// 		}
// 	}
// }

// func TestRelayGatewayUpdateUnmarshalFailure(t *testing.T) {
// 	handlerConfig := testGatewayHandlerConfig(&storage.StorerMock{})
// 	metric := testMetric(t)
// 	handlerConfig.UpdateMetrics.ErrorMetrics.UnmarshalFailure = metric

// 	buff := make([]byte, 10) // invalid relay packet size
// 	recorder := pingRelayGatewayUpdate(t, "application/octet-stream", buff, handlerConfig)
// 	relayGatewayErrorAssertions(t, recorder, http.StatusBadRequest, metric)
// }

// func TestRelayGatewayUpdateInvalidAddress(t *testing.T) {
// 	udp, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
// 	assert.NoError(t, err)
// 	packet := RelayUpdateRequest{
// 		Address: *udp,
// 		Token:   make([]byte, crypto.KeySize),
// 	}

// 	handlerConfig := testGatewayHandlerConfig(&storage.StorerMock{})
// 	metric := testMetric(t)
// 	handlerConfig.UpdateMetrics.ErrorMetrics.UnmarshalFailure = metric

// 	buff, err := packet.MarshalBinary()
// 	assert.NoError(t, err)
// 	badAddr := "invalid address"        // "invalid address" is luckily the same number of characters as "127.0.0.1:40000"
// 	for i := 0; i < len(badAddr); i++ { // Replace the address with the bad address character by character
// 		buff[8+i] = badAddr[i]
// 	}

// 	recorder := pingRelayGatewayUpdate(t, "application/octet-stream", buff, handlerConfig)
// 	relayGatewayErrorAssertions(t, recorder, http.StatusBadRequest, metric)
// }

// func TestRelayGatewayUpdateExceedMaxRelays(t *testing.T) {
// 	udp, err := net.ResolveUDPAddr("udp", "127.0.0.1:40000")
// 	assert.NoError(t, err)
// 	packet := RelayUpdateRequest{
// 		Address:   *udp,
// 		Token:     make([]byte, crypto.KeySize),
// 		PingStats: make([]routing.RelayStatsPing, 1025),
// 	}

// 	handlerConfig := testGatewayHandlerConfig(&storage.StorerMock{})
// 	metric := testMetric(t)
// 	handlerConfig.UpdateMetrics.ErrorMetrics.ExceedMaxRelays = metric

// 	buff, err := packet.MarshalBinary()
// 	assert.NoError(t, err)
// 	recorder := pingRelayGatewayUpdate(t, "application/octet-stream", buff, handlerConfig)
// 	relayGatewayErrorAssertions(t, recorder, http.StatusBadRequest, metric)

// }

// func TestRelayGatewayUpdateGhostRelayIgnore(t *testing.T) {
// 	addr := "127.0.0.1:40000"
// 	udp, err := net.ResolveUDPAddr("udp", addr)
// 	assert.NoError(t, err)

// 	packet := RelayUpdateRequest{
// 		Address:   *udp,
// 		Token:     make([]byte, crypto.KeySize),
// 		PingStats: make([]routing.RelayStatsPing, 0),
// 	}

// 	handlerConfig := testGatewayHandlerConfig(&storage.InMemory{})
// 	metric := testMetric(t)
// 	handlerConfig.UpdateMetrics.ErrorMetrics.RelayNotFound = metric

// 	buff, err := packet.MarshalBinary()
// 	assert.NoError(t, err)
// 	recorder := pingRelayGatewayUpdate(t, "application/octet-stream", buff, handlerConfig)
// 	relayGatewayErrorAssertions(t, recorder, http.StatusNotFound, metric)
// }
// func TestRelayGatewayShuttingDown(t *testing.T) {
// 	addr := "127.0.0.1:40000"
// 	udp, err := net.ResolveUDPAddr("udp", addr)
// 	assert.NoError(t, err)

// 	packet := RelayUpdateRequest{
// 		Address:      *udp,
// 		Token:        make([]byte, crypto.KeySize),
// 		ShuttingDown: true,
// 	}

// 	relay := routing.Relay{
// 		ID:   crypto.HashID(addr),
// 		Addr: *udp,
// 		Datacenter: routing.Datacenter{
// 			ID:   1,
// 			Name: "some name",
// 		},
// 		PublicKey: make([]byte, crypto.KeySize),
// 		Seller: routing.Seller{
// 			ID:   "sellerID",
// 			Name: "seller name",
// 		},
// 		State: routing.RelayStateEnabled,
// 	}

// 	inMemory := &storage.InMemory{}
// 	testAddRelayToStore(t, inMemory, relay)

// 	localMetrics := metrics.LocalHandler{}
// 	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
// 	assert.NoError(t, err)

// 	updateMetrics := metrics.EmptyRelayUpdateMetrics
// 	v := reflect.ValueOf(&updateMetrics.ErrorMetrics).Elem()
// 	for i := 0; i < v.NumField(); i++ {
// 		if v.Field(i).CanSet() {
// 			v.Field(i).Set(reflect.ValueOf(metric))
// 		}
// 	}

// 	handlerConfig := testGatewayHandlerConfig(inMemory)
// 	handlerConfig.UpdateMetrics = &updateMetrics
// 	buff, err := packet.MarshalBinary()
// 	assert.NoError(t, err)
// 	recorder := pingRelayGatewayUpdate(t, "application/octet-stream", buff, handlerConfig)
// 	relayGatewayUpdateShutdownAssertions(t, recorder, handlerConfig, addr)
// }

// func TestRelayGatewayUpdateRelayUnmarshalFailure(t *testing.T) {
// 	addr := "127.0.0.1:40000"
// 	udp, err := net.ResolveUDPAddr("udp", addr)
// 	assert.NoError(t, err)

// 	packet := RelayUpdateRequest{
// 		Address:   *udp,
// 		Token:     make([]byte, crypto.KeySize),
// 		PingStats: make([]routing.RelayStatsPing, 0),
// 	}

// 	storedToken := make([]byte, crypto.KeySize)
// 	entry := &routing.RelayData{
// 		ID:   crypto.HashID(addr),
// 		Addr: *udp,
// 		Datacenter: routing.Datacenter{
// 			ID:   1,
// 			Name: "some name",
// 		},
// 		PublicKey:      storedToken,
// 		LastUpdateTime: time.Now().Add(-time.Second),
// 	}

// 	// add a relay to storage to pass the ghost checks in RelayUpdateHandlerFunc
// 	relay := routing.Relay{
// 		ID:   entry.ID,
// 		Addr: entry.Addr,
// 		Seller: routing.Seller{
// 			ID:   "sellerID",
// 			Name: "seller name",
// 		},
// 		Datacenter: entry.Datacenter,
// 		PublicKey:  entry.PublicKey,
// 		State:      routing.RelayStateEnabled,
// 	}

// 	inMemory := &storage.InMemory{}
// 	testAddRelayToStore(t, inMemory, relay)

// 	handlerConfig := testGatewayHandlerConfig(inMemory)
// 	metric := testMetric(t)
// 	handlerConfig.UpdateMetrics.ErrorMetrics.UnmarshalFailure = metric
// 	buff, err := packet.MarshalBinary()
// 	buff[3] = 'a'
// 	assert.NoError(t, err)

// 	recorder := pingRelayGatewayUpdate(t, "application/octet-stream", buff, handlerConfig)
// 	relayGatewayErrorAssertions(t, recorder, http.StatusBadRequest, metric)
// }

// func TestRelayGatewayUpdateInvalidToken(t *testing.T) {
// 	addr := "127.0.0.1:40000"
// 	udp, err := net.ResolveUDPAddr("udp", addr)
// 	assert.NoError(t, err)

// 	incomingToken := make([]byte, crypto.KeySize)
// 	rand.Read(incomingToken)
// 	storedToken := make([]byte, crypto.KeySize)
// 	rand.Read(storedToken)
// 	packet := RelayUpdateRequest{
// 		Address:   *udp,
// 		Token:     incomingToken,
// 		PingStats: make([]routing.RelayStatsPing, 0),
// 	}

// 	// add a relay to storage to pass the ghost checks in RelayUpdateHandlerFunc
// 	relay := routing.Relay{
// 		ID:   crypto.HashID(addr),
// 		Addr: *udp,
// 		Seller: routing.Seller{
// 			ID:   "sellerID",
// 			Name: "seller name",
// 		},
// 		Datacenter: routing.Datacenter{
// 			ID:   1,
// 			Name: "some name",
// 		},
// 		PublicKey: storedToken,
// 		State:     routing.RelayStateEnabled,
// 	}

// 	inMemory := &storage.InMemory{}
// 	testAddRelayToStore(t, inMemory, relay)

// 	handlerConfig := testGatewayHandlerConfig(inMemory)
// 	metric := testMetric(t)
// 	handlerConfig.UpdateMetrics.ErrorMetrics.InvalidToken = metric

// 	buff, err := packet.MarshalBinary()
// 	assert.NoError(t, err)
// 	recorder := pingRelayGatewayUpdate(t, "application/octet-stream", buff, handlerConfig)
// 	relayGatewayErrorAssertions(t, recorder, http.StatusBadRequest, metric)

// }

// func TestRelayGatewayUpdateInvalidState(t *testing.T) {
// 	addr := "127.0.0.1:40000"
// 	udp, err := net.ResolveUDPAddr("udp", addr)
// 	assert.NoError(t, err)
// 	packet := RelayUpdateRequest{
// 		Address: *udp,
// 		Token:   make([]byte, crypto.KeySize),
// 	}

// 	// add a relay to storage to pass the ghost checks in RelayUpdateHandlerFunc
// 	relay := routing.Relay{
// 		ID:   crypto.HashID(addr),
// 		Addr: *udp,
// 		Seller: routing.Seller{
// 			ID:   "sellerID",
// 			Name: "seller name",
// 		},
// 		Datacenter: routing.Datacenter{
// 			ID:   1,
// 			Name: "some name",
// 		},
// 		PublicKey: make([]byte, crypto.KeySize),
// 		State:     routing.RelayStateQuarantine,
// 	}

// 	inMemory := &storage.InMemory{}
// 	testAddRelayToStore(t, inMemory, relay)

// 	handlerConfig := testGatewayHandlerConfig(inMemory)
// 	metric := testMetric(t)
// 	handlerConfig.UpdateMetrics.ErrorMetrics.RelayNotEnabled = metric
// 	buff, err := packet.MarshalBinary()
// 	assert.NoError(t, err)
// 	recorder := pingRelayGatewayUpdate(t, "application/octet-stream", buff, handlerConfig)
// 	relayGatewayErrorAssertions(t, recorder, http.StatusUnauthorized, metric)
// }

// func TestRelayGatewayUpdateSuccess(t *testing.T) {
// 	addr := "127.0.0.1:40000"
// 	udp, err := net.ResolveUDPAddr("udp", addr)
// 	assert.NoError(t, err)

// 	statIps := []string{"127.0.0.2:40000", "127.0.0.3:40000", "127.0.0.4:40000", "127.0.0.5:40000"}
// 	packet := RelayUpdateRequest{
// 		Address: *udp,
// 		Token:   make([]byte, crypto.KeySize),
// 	}

// 	packet.PingStats = make([]routing.RelayStatsPing, len(statIps))
// 	for i, addr := range statIps {
// 		stats := &packet.PingStats[i]
// 		stats.RelayID = crypto.HashID(addr)
// 		stats.RTT = rand.Float32()
// 		stats.Jitter = rand.Float32()
// 		stats.PacketLoss = rand.Float32()
// 	}

// 	relay := routing.Relay{
// 		ID:   crypto.HashID(addr),
// 		Addr: *udp,
// 		Seller: routing.Seller{
// 			ID:   "sellerID",
// 			Name: "seller name",
// 		},
// 		Datacenter: routing.Datacenter{
// 			ID:   1,
// 			Name: "some name",
// 		},
// 		PublicKey: make([]byte, crypto.KeySize),
// 		State:     routing.RelayStateEnabled,
// 	}

// 	inMemory := &storage.InMemory{}
// 	testAddRelayToStore(t, inMemory, relay)
// 	seedStorage(t, inMemory, statIps)

// 	localMetrics := metrics.LocalHandler{}
// 	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
// 	assert.NoError(t, err)

// 	updateMetrics := metrics.RelayUpdateMetrics{
// 		Invocations:   &metrics.EmptyCounter{},
// 		DurationGauge: &metrics.EmptyGauge{},
// 	}
// 	v := reflect.ValueOf(&updateMetrics.ErrorMetrics).Elem()
// 	for i := 0; i < v.NumField(); i++ {
// 		if v.Field(i).CanSet() {
// 			v.Field(i).Set(reflect.ValueOf(metric))
// 		}
// 	}

// 	handlerConfig := testGatewayHandlerConfig(inMemory)

// 	buff, err := packet.MarshalBinary()
// 	assert.NoError(t, err)
// 	recorder := pingRelayGatewayUpdate(t, "application/octet-stream", buff, handlerConfig)
// 	relayGatewayUpdateSuccessAssertions(t, recorder, "application/octet-stream", handlerConfig, statIps, addr)
// }

// func seedStorage(t *testing.T, inMemory *storage.InMemory, addressesToAdd []string) {
// 	for i, addrString := range addressesToAdd {
// 		addr, err := net.ResolveUDPAddr("udp", addrString)
// 		assert.NoError(t, err)

// 		relay := routing.Relay{
// 			ID:   crypto.HashID(addrString),
// 			Name: fmt.Sprintf("Relay %d", i),
// 			Addr: *addr,
// 			Seller: routing.Seller{
// 				ID:   fmt.Sprintf("%d", i),
// 				Name: fmt.Sprintf("Seller %d", i),
// 			},
// 			Datacenter: routing.Datacenter{
// 				ID:   crypto.HashID(fmt.Sprintf("Datacenter %d", i)),
// 				Name: fmt.Sprintf("Datacenter %d", i),
// 			},
// 			State: routing.RelayStateEnabled,
// 		}

// 		err = inMemory.AddSeller(context.Background(), relay.Seller)
// 		assert.NoError(t, err)

// 		err = inMemory.AddDatacenter(context.Background(), relay.Datacenter)
// 		assert.NoError(t, err)

// 		err = inMemory.AddRelay(context.Background(), relay)
// 		assert.NoError(t, err)
// 	}
// }
