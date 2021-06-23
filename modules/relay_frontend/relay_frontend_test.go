package relay_frontend

import (
	"bytes"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/networknext/backend/modules/analytics"
	"github.com/networknext/backend/modules/common/helpers"
	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/encoding"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/storage"

	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	store := storage.MatrixStoreMock{}
	cfg := &RelayFrontendConfig{MasterTimeVariance: timeVariance(15)}
	svc, err := NewRelayFrontend(&store, cfg)
	assert.Nil(t, err)
	assert.NotNil(t, svc)
	assert.Equal(t, timeVariance(15), svc.cfg.MasterTimeVariance)
	assert.Equal(t, &store, svc.store)
	assert.NotEqual(t, 0, svc.id)
}

func timeVariance(value int) time.Duration {
	return time.Duration(value) * time.Millisecond
}

func TestRelayFrontendSvc_UpdateRelayBackendMasterSetAndUpdate(t *testing.T) {
	currTime := time.Now()
	rb1 := storage.RelayBackendLiveData{
		ID:        "12345",
		Address:   "1.1.1.1",
		InitAt:    currTime.Add(-10 * time.Second),
		UpdatedAt: currTime,
	}

	rb2 := storage.RelayBackendLiveData{
		ID:        "54321",
		Address:   "2.1.1.1",
		InitAt:    currTime.Add(-15 * time.Second),
		UpdatedAt: currTime,
	}

	rb3 := storage.RelayBackendLiveData{
		ID:        "67890",
		Address:   "3.1.1.1",
		InitAt:    currTime.Add(-2 * time.Second),
		UpdatedAt: currTime,
	}

	store := storage.MatrixStoreMock{
		GetRelayBackendLiveDataFunc: func() ([]storage.RelayBackendLiveData, error) {
			return []storage.RelayBackendLiveData{rb1, rb2, rb3}, nil
		},
	}

	cfg := &RelayFrontendConfig{MasterTimeVariance: time.Second}
	svc, err := NewRelayFrontend(&store, cfg)
	assert.Nil(t, err)

	// rb2 should be master
	err = svc.UpdateRelayBackendMaster()
	assert.Nil(t, err)
	assert.Equal(t, "2.1.1.1", svc.currentMasterBackendAddress)

	// change to rb1 as master
	rb2.UpdatedAt = rb2.UpdatedAt.Add(-6 * time.Second)
	err = svc.UpdateRelayBackendMaster()
	assert.Nil(t, err)
	assert.Equal(t, "1.1.1.1", svc.currentMasterBackendAddress)
}

func TestRelayFrontendSvc_UpdateRelayBackendMasterCurrent(t *testing.T) {
	currTime := time.Now()
	rb1 := storage.RelayBackendLiveData{
		ID:        "12345",
		Address:   "1.1.1.1",
		InitAt:    currTime.Add(-10 * time.Second),
		UpdatedAt: currTime,
	}

	rb2 := storage.RelayBackendLiveData{
		ID:        "54321",
		Address:   "2.1.1.1",
		InitAt:    currTime.Add(-15 * time.Second),
		UpdatedAt: currTime,
	}

	store := storage.MatrixStoreMock{
		GetRelayBackendLiveDataFunc: func() ([]storage.RelayBackendLiveData, error) {
			return []storage.RelayBackendLiveData{rb1, rb2}, nil
		},
	}

	cfg := &RelayFrontendConfig{MasterTimeVariance: time.Second}
	svc, err := NewRelayFrontend(&store, cfg)
	assert.Nil(t, err)

	svc.currentMasterBackendAddress = rb2.Address
	// rb2 should be master
	err = svc.UpdateRelayBackendMaster()
	assert.Nil(t, err)
	assert.Equal(t, "2.1.1.1", svc.currentMasterBackendAddress)
}

func TestRelayFrontendSvc_ChooseRelayBackendMaster(t *testing.T) {
	currTime := time.Now()
	rbArr := []storage.RelayBackendLiveData{
		{
			ID:        "12345",
			Address:   "1.1.1.1",
			InitAt:    currTime.Add(-10 * time.Second),
			UpdatedAt: currTime,
		},
		{
			ID:        "54321",
			Address:   "2.1.1.1",
			InitAt:    currTime.Add(-15 * time.Second),
			UpdatedAt: currTime,
		},
		{
			ID:        "67890",
			Address:   "3.1.1.1",
			InitAt:    currTime.Add(-5 * time.Second),
			UpdatedAt: currTime,
		},
	}

	rbAddr, err := chooseRelayBackendMaster(rbArr, time.Second)
	assert.Nil(t, err)
	assert.Equal(t, rbArr[1].Address, rbAddr)

	rbArr[1].UpdatedAt = currTime.Add(-5 * time.Second)
	rbAddr, err = chooseRelayBackendMaster(rbArr, time.Second)
	assert.Nil(t, err)
	assert.Equal(t, rbArr[0].Address, rbAddr)

	rbArr[0].UpdatedAt = currTime.Add(-5 * time.Second)
	rbAddr, err = chooseRelayBackendMaster(rbArr, time.Second)
	assert.Nil(t, err)
	assert.Equal(t, rbArr[2].Address, rbAddr)

	rbArr[2].UpdatedAt = currTime.Add(-5 * time.Second)
	rbAddr, err = chooseRelayBackendMaster(rbArr, time.Second)
	assert.NotNil(t, err)
	assert.Equal(t, "", rbAddr)
}

func TestRelayFrontendSvc_GetMatrixAddress(t *testing.T) {
	svc := new(RelayFrontendSvc)
	svc.currentMasterBackendAddress = "1.1.1.1"

	address, err := svc.GetMatrixAddress(MatrixTypeCost)
	assert.Nil(t, err)
	assert.Equal(t, "http://1.1.1.1/cost_matrix", address)

	address, err = svc.GetMatrixAddress(MatrixTypeNormal)
	assert.Nil(t, err)
	assert.Equal(t, "http://1.1.1.1/route_matrix", address)

	address, err = svc.GetMatrixAddress("not_an_address")
	assert.NotNil(t, err)
	assert.Equal(t, "", address)
}

func TestRelayFrontendSvc_GetHttpMatrix(t *testing.T) {

	testMatrix := testMatrix(t)
	bin := testMatrix.GetResponseData()
	assert.NotEqual(t, 0, len(bin))

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		buffer := bytes.NewBuffer(bin)
		_, err := buffer.WriteTo(w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))

	matrix, err := getHttpMatrix(ts.URL)
	assert.NoError(t, err)
	assert.Equal(t, bin, matrix)
}

func TestRelayFrontendSvc_CacheMatrixCost(t *testing.T) {
	testMatrix := testMatrix(t)
	bin := testMatrix.GetResponseData()
	assert.NotEqual(t, 0, len(bin))

	svc := new(RelayFrontendSvc)
	svc.costMatrix = new(helpers.MatrixData)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		buffer := bytes.NewBuffer(bin)
		_, err := buffer.WriteTo(w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))

	err := svc.cacheMatrixInternal(ts.URL, MatrixTypeCost)
	assert.NoError(t, err)
	assert.Equal(t, bin, svc.costMatrix.GetMatrix())
}

func TestRelayFrontendSvc_CacheMatrixNormal(t *testing.T) {
	testMatrix := testMatrix(t)
	bin := testMatrix.GetResponseData()
	assert.NotEqual(t, 0, len(bin))

	svc := new(RelayFrontendSvc)
	svc.routeMatrix = new(helpers.MatrixData)

	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		buffer := bytes.NewBuffer(bin)
		_, err := buffer.WriteTo(w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))

	err := svc.cacheMatrixInternal(ts.URL, MatrixTypeNormal)
	assert.NoError(t, err)
	assert.Equal(t, bin, svc.routeMatrix.GetMatrix())
}

func TestRelayFrontendSvc_GetCostMatrix(t *testing.T) {
	svc := &RelayFrontendSvc{}
	svc.costMatrix = new(helpers.MatrixData)
	testMatrix := testMatrix(t)
	svc.costMatrix.SetMatrix(testMatrix.GetResponseData())

	ts := httptest.NewServer(http.HandlerFunc(svc.GetCostMatrixHandlerFunc()))

	resp, err := http.Get(ts.URL)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Body)

	buffer, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.NotEmpty(t, buffer)

	var newRouteMatrix routing.RouteMatrix
	rs := encoding.CreateReadStream(buffer)
	err = newRouteMatrix.Serialize(rs)
	assert.NoError(t, err)

	newRouteMatrix.WriteResponseData(5000)
	assert.Equal(t, testMatrix, newRouteMatrix)
}

func TestRelayFrontendSvc_GetCostMatrixNotFound(t *testing.T) {
	svc := &RelayFrontendSvc{}
	svc.costMatrix = new(helpers.MatrixData)
	ts := httptest.NewServer(http.HandlerFunc(svc.GetCostMatrixHandlerFunc()))

	resp, err := http.Get(ts.URL)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, http.StatusNotFound)
}

func TestRelayFrontendSvc_ResetCostMatrix(t *testing.T) {
	svc := &RelayFrontendSvc{}
	svc.costMatrix = new(helpers.MatrixData)
	testMatrix := testMatrix(t)
	svc.costMatrix.SetMatrix(testMatrix.GetResponseData())

	err := svc.ResetCachedMatrix(MatrixTypeCost)
	assert.NoError(t, err)

	expectedEmptyCostMatrix := routing.CostMatrix{
		RelayIDs:           []uint64{},
		RelayAddresses:     []net.UDPAddr{},
		RelayNames:         []string{},
		RelayLatitudes:     []float32{},
		RelayLongitudes:    []float32{},
		RelayDatacenterIDs: []uint64{},
		Costs:              []int32{},
		Version:            routing.CostMatrixSerializeVersion,
		DestRelays:         []bool{},
	}

	receivedCostMatrixBin := svc.costMatrix.GetMatrix()
	var receivedCostMatrix routing.CostMatrix

	readStream := encoding.CreateReadStream(receivedCostMatrixBin)
	err = receivedCostMatrix.Serialize(readStream)
	assert.NoError(t, err)

	assert.Equal(t, expectedEmptyCostMatrix, receivedCostMatrix)
}

func TestRelayFrontendSvc_GetRouteMatrix(t *testing.T) {
	svc := &RelayFrontendSvc{}
	svc.routeMatrix = new(helpers.MatrixData)
	testMatrix := testMatrix(t)
	svc.routeMatrix.SetMatrix(testMatrix.GetResponseData())
	ts := httptest.NewServer(http.HandlerFunc(svc.GetRouteMatrixHandlerFunc()))

	resp, err := http.Get(ts.URL)
	assert.NoError(t, err)
	assert.NotEmpty(t, resp.Body)

	buffer, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.NotEmpty(t, buffer)

	var newRouteMatrix routing.RouteMatrix
	rs := encoding.CreateReadStream(buffer)
	err = newRouteMatrix.Serialize(rs)
	assert.NoError(t, err)

	newRouteMatrix.WriteResponseData(5000)
	assert.Equal(t, testMatrix, newRouteMatrix)
}

func TestRelayFrontendSvc_GetRouteMatrixNotFound(t *testing.T) {
	svc := &RelayFrontendSvc{}
	svc.routeMatrix = new(helpers.MatrixData)
	ts := httptest.NewServer(http.HandlerFunc(svc.GetRouteMatrixHandlerFunc()))

	resp, err := http.Get(ts.URL)
	assert.NoError(t, err)
	assert.Equal(t, resp.StatusCode, http.StatusNotFound)
}

func TestRelayFrontendSvc_ResetRouteMatrix(t *testing.T) {
	svc := &RelayFrontendSvc{}
	svc.routeMatrix = new(helpers.MatrixData)
	testMatrix := testMatrix(t)
	testMatrix.Version = routing.RouteMatrixSerializeVersion
	svc.routeMatrix.SetMatrix(testMatrix.GetResponseData())

	err := svc.ResetCachedMatrix(MatrixTypeNormal)
	assert.NoError(t, err)

	expectedEmptyRouteMatrix := routing.RouteMatrix{
		RelayIDsToIndices:   make(map[uint64]int32),
		RelayIDs:            []uint64{},
		RelayAddresses:      []net.UDPAddr{},
		RelayNames:          []string{},
		RelayLatitudes:      []float32{},
		RelayLongitudes:     []float32{},
		RelayDatacenterIDs:  []uint64{},
		RouteEntries:        []core.RouteEntry{},
		BinFileBytes:        0,
		CreatedAt:           0,
		Version:             routing.RouteMatrixSerializeVersion,
		DestRelays:          []bool{},
		PingStats:           []analytics.PingStatsEntry{},
		RelayStats:          []analytics.RelayStatsEntry{},
		FullRelayIDs:        []uint64{},
		FullRelayIndicesSet: make(map[int32]bool),
	}

	receivedRouteMatrixBin := svc.routeMatrix.GetMatrix()
	var receivedRouteMatrix routing.RouteMatrix

	readStream := encoding.CreateReadStream(receivedRouteMatrixBin)
	err = receivedRouteMatrix.Serialize(readStream)
	assert.NoError(t, err)

	assert.Equal(t, expectedEmptyRouteMatrix, receivedRouteMatrix)
}

func TestRelayFrontendSvc_GetRelayBackendHandler(t *testing.T) {
	svc := &RelayFrontendSvc{}

	backendHandler := func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	}

	bSvr := httptest.NewServer(http.HandlerFunc(backendHandler))
	svc.currentMasterBackendAddress = strings.TrimLeft(bSvr.URL, "http://")
	assert.NotEqual(t, svc.currentMasterBackendAddress, bSvr.URL)

	fSvr := httptest.NewServer(http.HandlerFunc(svc.GetRelayBackendHandlerFunc("/database_version")))
	resp, err := http.Get(fSvr.URL)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.NotEmpty(t, body)
	assert.Equal(t, []byte("test"), body)
}

func TestRelayFrontendSvc_GetRelayDashboardHandler(t *testing.T) {
	svc := &RelayFrontendSvc{}

	backendHandler := func(w http.ResponseWriter, r *http.Request) {
		u, p, _ := r.BasicAuth()
		if u != "testUsername" || p != "testPassword" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("test"))
	}

	bSvr := httptest.NewServer(http.HandlerFunc(backendHandler))
	svc.currentMasterBackendAddress = strings.TrimLeft(bSvr.URL, "http://")
	assert.NotEqual(t, svc.currentMasterBackendAddress, bSvr.URL)

	fSvr := httptest.NewServer(http.HandlerFunc(svc.GetRelayDashboardHandlerFunc("testUsername", "testPassword")))
	client := &http.Client{}

	req, err := http.NewRequest("GET", fSvr.URL, nil)
	assert.NoError(t, err)
	req.SetBasicAuth("testUsername", "testPassword")

	resp, err := client.Do(req)
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, resp.StatusCode)

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	assert.NoError(t, err)
	assert.NotEmpty(t, body)
	assert.Equal(t, []byte("test"), body)
}

func testMatrix(t *testing.T) routing.RouteMatrix {
	relayAddr1, err := net.ResolveUDPAddr("udp", "127.0.0.1:10000")
	assert.NoError(t, err)
	relayAddr2, err := net.ResolveUDPAddr("udp", "127.0.0.1:10001")
	assert.NoError(t, err)

	expected := routing.RouteMatrix{
		RelayIDsToIndices:  map[uint64]int32{1: 0, 2: 1},
		RelayIDs:           []uint64{1, 2},
		RelayAddresses:     []net.UDPAddr{*relayAddr1, *relayAddr2},
		RelayNames:         []string{"test.relay.1", "test.relay.2"},
		RelayLatitudes:     []float32{90, 89},
		RelayLongitudes:    []float32{180, 179},
		RelayDatacenterIDs: []uint64{10, 10},
		RouteEntries: []core.RouteEntry{
			{
				DirectCost:     65,
				NumRoutes:      int32(core.TriMatrixLength(2)),
				RouteCost:      [core.MaxRoutesPerEntry]int32{35},
				RouteNumRelays: [core.MaxRoutesPerEntry]int32{2},
				RouteRelays: [core.MaxRoutesPerEntry][core.MaxRelaysPerRoute]int32{
					{
						0, 1,
					},
				},
				RouteHash: [core.MaxRoutesPerEntry]uint32{core.RouteHash(0, 1)},
			},
		},
	}
	err = expected.WriteResponseData(5000)
	assert.NoError(t, err)
	return expected
}
