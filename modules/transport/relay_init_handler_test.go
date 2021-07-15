package transport_test

// func testRelayInitHandlerConfig(RelayMap *routing.RelayMap, storer storage.Storer, metric *metrics.RelayInitMetrics, routerPrivateKey []byte) *transport.RelayInitHandlerConfig {

// 	if RelayMap == nil {
// 		RelayMap = routing.NewRelayMap(func(relay *routing.RelayData) error {
// 			return nil
// 		})
// 	}

// 	if metric == nil {
// 		metric = &metrics.EmptyRelayInitMetrics
// 	}

// 	return &transport.RelayInitHandlerConfig{
// 		RelayMap:         RelayMap,
// 		Storer:           storer,
// 		Metrics:          metric,
// 		RouterPrivateKey: routerPrivateKey,
// 	}
// }

// func pingRelayBackendInit(t *testing.T, contentType string, body []byte, handlerConfig *transport.RelayInitHandlerConfig) *httptest.ResponseRecorder {
// 	customerPublicKey := make([]byte, crypto.KeySize)
// 	rand.Read(customerPublicKey)

// 	recorder := httptest.NewRecorder()
// 	request, err := http.NewRequest("POST", "/relay_init", bytes.NewBuffer(body))
// 	assert.NoError(t, err)
// 	request.Header.Add("Content-Type", contentType)

// 	handler := transport.RelayInitHandlerFunc(log.NewNopLogger(), handlerConfig)

// 	handler(recorder, request)
// 	return recorder
// }

// func testMetric(t *testing.T) metrics.Counter {
// 	localMetrics := metrics.LocalHandler{}
// 	metric, err := localMetrics.NewCounter(context.Background(), &metrics.Descriptor{ID: "test metric"})
// 	assert.NoError(t, err)
// 	return metric
// }

// func testAddRelayToStore(t *testing.T, storer *storage.InMemory, relay routing.Relay) {
// 	err := storer.AddSeller(context.Background(), relay.Seller)
// 	assert.NoError(t, err)
// 	err = storer.AddDatacenter(context.Background(), relay.Datacenter)
// 	assert.NoError(t, err)
// 	err = storer.AddRelay(context.Background(), relay)
// 	assert.NoError(t, err)
// }

// func relayInitErrorAssertions(t *testing.T, recorder *httptest.ResponseRecorder, expectedCode int, errMetric metrics.Counter) {
// 	assert.Equal(t, expectedCode, recorder.Code)
// 	assert.Equal(t, 1.0, errMetric.ValueReset())
// }

// func relayInitSuccessAssertions(t *testing.T, recorder *httptest.ResponseRecorder, expectedContentType string, handlerConfig *transport.RelayInitHandlerConfig, addr string, before uint64, expected routing.RelayData) {
// 	assert.Equal(t, http.StatusOK, recorder.Code)

// 	actual := handlerConfig.RelayMap.GetRelayData(addr)
// 	assert.NotNil(t, actual)
// 	relay, err := handlerConfig.Storer.Relay(actual.ID)
// 	assert.NoError(t, err)

// 	header := recorder.Header()
// 	contentType, ok := header["Content-Type"]
// 	assert.True(t, ok)

// 	body := recorder.Body.Bytes()
// 	var response transport.RelayInitResponse
// 	switch expectedContentType {
// 	case "application/octet-stream":
// 		err = response.UnmarshalBinary(body)
// 		assert.NoError(t, err)
// 	default:
// 		assert.FailNow(t, "Invalid expected content type")
// 	}

// 	assert.Equal(t, expectedContentType, contentType[0])
// 	assert.Equal(t, transport.VersionNumberInitResponse, int(response.Version))
// 	assert.LessOrEqual(t, before, response.Timestamp)
// 	assert.GreaterOrEqual(t, uint64(time.Now().Unix()*1000), response.Timestamp)
// 	assert.Equal(t, actual.PublicKey, response.PublicKey) // entry gets a public key assigned at init which is returned in the response

// 	assert.Equal(t, expected.ID, actual.ID)
// 	assert.Equal(t, expected.Name, actual.Name)
// 	assert.Equal(t, expected.Addr, actual.Addr)
// 	assert.NotZero(t, actual.LastUpdateTime)
// 	assert.Len(t, actual.PublicKey, 32)
// 	assert.Equal(t, routing.RelayStateEnabled, relay.State)

// 	errMetricsStruct := reflect.ValueOf(handlerConfig.Metrics.ErrorMetrics)
// 	for i := 0; i < errMetricsStruct.NumField(); i++ {
// 		if errMetricsStruct.Field(i).CanInterface() {
// 			assert.Equal(t, 0.0, errMetricsStruct.Field(i).Interface().(metrics.Counter).ValueReset())
// 		}
// 	}
// }

// func TestRelayInitUnmarshalFailure(t *testing.T) {
// 	handlerConfig := testRelayInitHandlerConfig(nil, &storage.StorerMock{}, nil, []byte{})
// 	metric := testMetric(t)
// 	handlerConfig.Metrics.ErrorMetrics.UnmarshalFailure = metric
// 	buff := []byte("bad packet")
// 	recorder := pingRelayBackendInit(t, "application/octet-stream", buff, handlerConfig)
// 	relayInitErrorAssertions(t, recorder, http.StatusBadRequest, metric)

// }

// func TestRelayInitInvalidMagic(t *testing.T) {
// 	packet := transport.RelayInitRequest{
// 		Magic:          0xFFFFFFFF,
// 		Nonce:          make([]byte, crypto.NonceSize),
// 		EncryptedToken: make([]byte, routing.EncryptedRelayTokenSize),
// 	}

// 	handlerConfig := testRelayInitHandlerConfig(nil, &storage.StorerMock{}, nil, []byte{})
// 	metric := testMetric(t)
// 	handlerConfig.Metrics.ErrorMetrics.InvalidMagic = metric

// 	buff, err := packet.MarshalBinary()
// 	assert.NoError(t, err)
// 	recorder := pingRelayBackendInit(t, "application/octet-stream", buff, handlerConfig)
// 	relayInitErrorAssertions(t, recorder, http.StatusBadRequest, metric)

// }

// func TestRelayInitInvalidAddress(t *testing.T) {
// 	relayPublicKey, _, err := box.GenerateKey(rand.Reader)
// 	assert.NoError(t, err)
// 	_, routerPrivateKey, err := box.GenerateKey(rand.Reader)
// 	assert.NoError(t, err)

// 	addr := "127.0.0.1:40000"
// 	udp, err := net.ResolveUDPAddr("udp", addr)
// 	assert.NoError(t, err)

// 	relay := routing.Relay{
// 		ID:        crypto.HashID(addr),
// 		PublicKey: relayPublicKey[:],
// 		Seller: routing.Seller{
// 			ID:   "sellerID",
// 			Name: "seller name",
// 		},
// 		Datacenter: routing.Datacenter{
// 			ID:   crypto.HashID("some datacenter"),
// 			Name: "some datacenter",
// 		},
// 	}
// 	inMemory := &storage.InMemory{}
// 	testAddRelayToStore(t, inMemory, relay)

// 	packet := transport.RelayInitRequest{
// 		Magic:          transport.InitRequestMagic,
// 		Version:        0,
// 		Nonce:          make([]byte, crypto.NonceSize),
// 		Address:        *udp,
// 		EncryptedToken: make([]byte, routing.EncryptedRelayTokenSize),
// 	}

// 	handlerConfig := testRelayInitHandlerConfig(nil, inMemory, nil, routerPrivateKey[:])
// 	metric := testMetric(t)
// 	handlerConfig.Metrics.ErrorMetrics.UnmarshalFailure = metric

// 	buff, err := packet.MarshalBinary()
// 	assert.NoError(t, err)
// 	badAddr := "invalid address"        // "invalid address" is luckily the same number of characters as "127.0.0.1:40000"
// 	for i := 0; i < len(badAddr); i++ { // Replace the address with the bad address character by character
// 		buff[4+4+crypto.NonceSize+4+i] = badAddr[i]
// 	}
// 	recorder := pingRelayBackendInit(t, "application/octet-stream", buff, handlerConfig)
// 	relayInitErrorAssertions(t, recorder, http.StatusBadRequest, metric)

// }

// func TestRelayInitRelayNotFound(t *testing.T) {
// 	addr := "127.0.0.1:40000"
// 	udpAddr, err := net.ResolveUDPAddr("udp", addr)
// 	assert.NoError(t, err)

// 	inMemory := &storage.InMemory{} // Have empty storage to fail lookup

// 	packet := transport.RelayInitRequest{
// 		Magic:          transport.InitRequestMagic,
// 		Nonce:          make([]byte, crypto.NonceSize),
// 		Address:        *udpAddr,
// 		EncryptedToken: make([]byte, routing.EncryptedRelayTokenSize),
// 	}

// 	handlerConfig := testRelayInitHandlerConfig(nil, inMemory, nil, []byte{})
// 	metric := testMetric(t)
// 	handlerConfig.Metrics.ErrorMetrics.RelayNotFound = metric

// 	buff, err := packet.MarshalBinary()
// 	assert.NoError(t, err)
// 	recorder := pingRelayBackendInit(t, "application/octet-stream", buff, handlerConfig)
// 	relayInitErrorAssertions(t, recorder, http.StatusNotFound, metric)
// }

// func TestRelayInitQuarantinedRelay(t *testing.T) {

// 	relayPublicKey, relayPrivateKey, err := box.GenerateKey(rand.Reader)
// 	assert.NoError(t, err)
// 	routerPublicKey, routerPrivateKey, err := box.GenerateKey(rand.Reader)
// 	assert.NoError(t, err)

// 	// generate nonce
// 	nonce := make([]byte, crypto.NonceSize)
// 	rand.Read(nonce)

// 	// generate token
// 	token := make([]byte, crypto.KeySize)
// 	rand.Read(token)

// 	// encrypt token
// 	encryptedToken := crypto.Seal(token, nonce, routerPublicKey[:], relayPrivateKey[:])

// 	addr := "127.0.0.1:40000"
// 	udpAddr, err := net.ResolveUDPAddr("udp", addr)
// 	assert.NoError(t, err)

// 	relay := routing.Relay{
// 		ID: crypto.HashID(addr),
// 		Seller: routing.Seller{
// 			ID:   "sellerID",
// 			Name: "seller name",
// 		},
// 		Datacenter: routing.Datacenter{
// 			ID:   crypto.HashID("some datacenter"),
// 			Name: "some datacenter",
// 			Location: routing.Location{
// 				Latitude:  13,
// 				Longitude: 13,
// 			},
// 		},
// 		PublicKey: relayPublicKey[:],
// 		State:     routing.RelayStateQuarantine,
// 	}

// 	packet := transport.RelayInitRequest{
// 		Magic:          transport.InitRequestMagic,
// 		Version:        0,
// 		Nonce:          nonce,
// 		Address:        *udpAddr,
// 		EncryptedToken: encryptedToken,
// 	}

// 	inMemory := &storage.InMemory{}
// 	customerPublicKey := make([]byte, crypto.KeySize)
// 	rand.Read(customerPublicKey)
// 	err = inMemory.AddBuyer(context.Background(), routing.Buyer{
// 		PublicKey: customerPublicKey,
// 	})
// 	assert.NoError(t, err)
// 	testAddRelayToStore(t, inMemory, relay)

// 	handlerConfig := testRelayInitHandlerConfig(nil, inMemory, nil, routerPrivateKey[:])
// 	metric := testMetric(t)
// 	handlerConfig.Metrics.ErrorMetrics.RelayQuarantined = metric

// 	buff, err := packet.MarshalBinary()
// 	assert.NoError(t, err)
// 	recorder := pingRelayBackendInit(t, "application/octet-stream", buff, handlerConfig)
// 	relayInitErrorAssertions(t, recorder, http.StatusUnauthorized, metric)
// }

// func TestRelayInitInvalidToken(t *testing.T) {
// 	_, routerPrivateKey, err := box.GenerateKey(rand.Reader)
// 	assert.NoError(t, err)

// 	// generate nonce
// 	nonce := make([]byte, crypto.NonceSize)
// 	rand.Read(nonce)

// 	// generate token but leave it as 0's
// 	token := make([]byte, routing.EncryptedRelayTokenSize)

// 	addr := "127.0.0.1:40000"
// 	udp, err := net.ResolveUDPAddr("udp", addr)
// 	assert.NoError(t, err)
// 	relay := routing.Relay{
// 		ID: crypto.HashID(addr),
// 		Seller: routing.Seller{
// 			ID:   "sellerID",
// 			Name: "seller name",
// 		},
// 		Datacenter: routing.Datacenter{
// 			ID:   crypto.HashID("some datacenter"),
// 			Name: "some datacenter",
// 		},
// 		PublicKey: []byte("fake"),
// 	}
// 	inMemory := &storage.InMemory{}
// 	testAddRelayToStore(t, inMemory, relay)

// 	packet := transport.RelayInitRequest{
// 		Magic:          transport.InitRequestMagic,
// 		Version:        0,
// 		Nonce:          nonce,
// 		Address:        *udp,
// 		EncryptedToken: token,
// 	}

// 	handlerConfig := testRelayInitHandlerConfig(nil, inMemory, nil, routerPrivateKey[:])
// 	metric := testMetric(t)
// 	handlerConfig.Metrics.ErrorMetrics.DecryptionFailure = metric

// 	buff, err := packet.MarshalBinary()
// 	assert.NoError(t, err)
// 	recorder := pingRelayBackendInit(t, "application/octet-stream", buff, handlerConfig)
// 	relayInitErrorAssertions(t, recorder, http.StatusUnauthorized, metric)
// }

// func TestRelayInitInvalidNonce(t *testing.T) {
// 	relayPublicKey, relayPrivateKey, err := box.GenerateKey(rand.Reader)
// 	assert.NoError(t, err)
// 	routerPublicKey, routerPrivateKey, err := box.GenerateKey(rand.Reader)
// 	assert.NoError(t, err)

// 	// generate nonce
// 	nonce := make([]byte, crypto.NonceSize)
// 	rand.Read(nonce)

// 	// generate random token
// 	token := make([]byte, crypto.KeySize)
// 	rand.Read(token)

// 	// seal it with the bad nonce
// 	encryptedToken := crypto.Seal(token, nonce, routerPublicKey[:], relayPrivateKey[:])

// 	addr := "127.0.0.1:40000"
// 	udp, err := net.ResolveUDPAddr("udp", addr)
// 	assert.NoError(t, err)
// 	relay := routing.Relay{
// 		ID:        crypto.HashID(addr),
// 		PublicKey: relayPublicKey[:],
// 		Seller: routing.Seller{
// 			ID:   "sellerID",
// 			Name: "seller name",
// 		},
// 		Datacenter: routing.Datacenter{
// 			ID:   crypto.HashID("some datacenter"),
// 			Name: "some datacenter",
// 		},
// 	}
// 	packet := transport.RelayInitRequest{
// 		Magic:          transport.InitRequestMagic,
// 		Version:        0,
// 		Nonce:          make([]byte, crypto.NonceSize), // Send a different nonce than the one used
// 		Address:        *udp,
// 		EncryptedToken: encryptedToken,
// 	}

// 	inMemory := &storage.InMemory{}
// 	testAddRelayToStore(t, inMemory, relay)
// 	handlerConfig := testRelayInitHandlerConfig(nil, inMemory, nil, routerPrivateKey[:])
// 	metric := testMetric(t)
// 	handlerConfig.Metrics.ErrorMetrics.DecryptionFailure = metric

// 	buff, err := packet.MarshalBinary()
// 	assert.NoError(t, err)
// 	recorder := pingRelayBackendInit(t, "application/octet-stream", buff, handlerConfig)
// 	relayInitErrorAssertions(t, recorder, http.StatusUnauthorized, metric)
// }

// func TestRelayInitRelayExists(t *testing.T) {
// 	relayPublicKey, relayPrivateKey, err := box.GenerateKey(rand.Reader)
// 	assert.NoError(t, err)
// 	routerPublicKey, routerPrivateKey, err := box.GenerateKey(rand.Reader)
// 	assert.NoError(t, err)

// 	// generate nonce
// 	nonce := make([]byte, crypto.NonceSize)
// 	rand.Read(nonce)

// 	// generate token
// 	token := make([]byte, crypto.KeySize)
// 	rand.Read(token)

// 	// encrypt token
// 	encryptedToken := crypto.Seal(token, nonce, routerPublicKey[:], relayPrivateKey[:])

// 	name := "some name"
// 	addr := "127.0.0.1:40000"
// 	udpAddr, err := net.ResolveUDPAddr("udp", addr)
// 	assert.NoError(t, err)
// 	dcname := "another name"

// 	entry := &routing.RelayData{
// 		ID:        crypto.HashID(addr),
// 		Name:      name,
// 		Addr:      *udpAddr,
// 		PublicKey: token,
// 		Datacenter: routing.Datacenter{
// 			ID:   crypto.HashID(dcname),
// 			Name: dcname,
// 		},
// 		LastUpdateTime: time.Now(),
// 	}

// 	relay := routing.Relay{
// 		ID: crypto.HashID(addr),
// 		Datacenter: routing.Datacenter{
// 			Name: "some datacenter",
// 		},
// 		PublicKey: relayPublicKey[:],
// 	}

// 	packet := transport.RelayInitRequest{
// 		Magic:          transport.InitRequestMagic,
// 		Version:        0,
// 		Nonce:          nonce,
// 		Address:        *udpAddr,
// 		EncryptedToken: encryptedToken,
// 	}

// 	inMemory := &storage.InMemory{}
// 	testAddRelayToStore(t, inMemory, relay)
// 	handlerConfig := testRelayInitHandlerConfig(nil, inMemory, nil, routerPrivateKey[:])
// 	metric := testMetric(t)
// 	handlerConfig.RelayMap.AddRelayDataEntry(addr, entry)
// 	handlerConfig.Metrics.ErrorMetrics.RelayAlreadyExists = metric

// 	buff, err := packet.MarshalBinary()
// 	assert.NoError(t, err)
// 	recorder := pingRelayBackendInit(t, "application/octet-stream", buff, handlerConfig)
// 	relayInitErrorAssertions(t, recorder, http.StatusConflict, metric)

// }

// func TestRelayInitSuccess(t *testing.T) {
// 	relayPublicKey, relayPrivateKey, err := box.GenerateKey(rand.Reader)
// 	assert.NoError(t, err)
// 	routerPublicKey, routerPrivateKey, err := box.GenerateKey(rand.Reader)
// 	assert.NoError(t, err)

// 	location := routing.Location{
// 		Latitude:  float32(math.Round(mrand.Float64()*1000) / 1000),
// 		Longitude: float32(math.Round(mrand.Float64()*1000) / 1000),
// 	}

// 	nonce := make([]byte, crypto.NonceSize)
// 	rand.Read(nonce)

// 	addr := "127.0.0.1:40000"
// 	udpAddr, err := net.ResolveUDPAddr("udp", addr)
// 	assert.NoError(t, err)

// 	token := make([]byte, crypto.KeySize)
// 	rand.Read(token)

// 	encryptedToken := crypto.Seal(token, nonce, routerPublicKey[:], relayPrivateKey[:])

// 	before := uint64(time.Now().Unix())

// 	relay := routing.Relay{
// 		ID:        crypto.HashID(addr),
// 		Addr:      *udpAddr,
// 		PublicKey: relayPublicKey[:],
// 		Seller: routing.Seller{
// 			ID:   "sellerID",
// 			Name: "seller name",
// 		},
// 		Datacenter: routing.Datacenter{
// 			ID:       crypto.HashID("some datacenter"),
// 			Name:     "some datacenter",
// 			Location: location,
// 		},
// 		State: routing.RelayStateOffline,
// 	}

// 	packet := transport.RelayInitRequest{
// 		Magic:          transport.InitRequestMagic,
// 		Nonce:          nonce,
// 		Address:        *udpAddr,
// 		EncryptedToken: encryptedToken,
// 	}

// 	expected := routing.RelayData{
// 		ID:   crypto.HashID(addr),
// 		Addr: *udpAddr,
// 		Datacenter: routing.Datacenter{
// 			ID:   crypto.HashID("some datacenter"),
// 			Name: "some datacenter",
// 		},
// 	}

// 	inMemory := &storage.InMemory{}
// 	customerPublicKey := make([]byte, crypto.KeySize)
// 	rand.Read(customerPublicKey)

// 	err = inMemory.AddBuyer(context.Background(), routing.Buyer{
// 		PublicKey: customerPublicKey,
// 	})
// 	testAddRelayToStore(t, inMemory, relay)
// 	handlerConfig := testRelayInitHandlerConfig(nil, inMemory, nil, routerPrivateKey[:])
// 	metric := testMetric(t)

// 	initMetrics := metrics.RelayInitMetrics{
// 		Invocations:   &metrics.EmptyCounter{},
// 		DurationGauge: &metrics.EmptyGauge{},
// 	}
// 	v := reflect.ValueOf(&initMetrics.ErrorMetrics).Elem()
// 	for i := 0; i < v.NumField(); i++ {
// 		if v.Field(i).CanSet() {
// 			v.Field(i).Set(reflect.ValueOf(metric))
// 		}
// 	}

// 	handlerConfig.Metrics = &initMetrics

// 	buff, err := packet.MarshalBinary()
// 	assert.NoError(t, err)

// 	recorder := pingRelayBackendInit(t, "application/octet-stream", buff, handlerConfig)
// 	relayInitSuccessAssertions(t, recorder, "application/octet-stream", handlerConfig, udpAddr.String(), before, expected)
// }
