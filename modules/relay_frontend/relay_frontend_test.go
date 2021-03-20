package relay_frontend

import (
	"bytes"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/networknext/backend/modules/encoding"

	"github.com/networknext/backend/modules/common/helpers"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/routing"

	"github.com/networknext/backend/modules/storage"
	"github.com/stretchr/testify/assert"
)

func TestNew(t *testing.T) {
	t.Parallel()
	store := storage.MatrixStoreMock{}
	cfg := &Config{MasterTimeVariance: timeVariance(15)}
	svc, err := New(&store, cfg)
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
		Id:        "12345",
		Address:   "1.1.1.1",
		InitAt:    currTime.Add(-10 * time.Second),
		UpdatedAt: currTime,
	}

	rb2 := storage.RelayBackendLiveData{
		Id:        "54321",
		Address:   "2.1.1.1",
		InitAt:    currTime.Add(-15 * time.Second),
		UpdatedAt: currTime,
	}

	rb3 := storage.RelayBackendLiveData{
		Id:        "67890",
		Address:   "3.1.1.1",
		InitAt:    currTime.Add(-2 * time.Second),
		UpdatedAt: currTime,
	}

	store := storage.MatrixStoreMock{
		GetRelayBackendLiveDataFunc: func(address []string) ([]storage.RelayBackendLiveData, error) {
			return []storage.RelayBackendLiveData{rb1, rb2, rb3}, nil
		},
	}

	cfg := &Config{MasterTimeVariance: time.Second}
	svc, err := New(&store, cfg)
	assert.Nil(t, err)

	//rb2 should be master
	err = svc.UpdateRelayBackendMaster()
	assert.Nil(t, err)
	assert.Equal(t, "2.1.1.1", svc.currentMasterBackendAddress)

	//change to rb1 as master
	rb2.UpdatedAt = rb2.UpdatedAt.Add(-6 * time.Second)
	err = svc.UpdateRelayBackendMaster()
	assert.Nil(t, err)
	assert.Equal(t, "1.1.1.1", svc.currentMasterBackendAddress)
}

func TestRelayFrontendSvc_UpdateRelayBackendMasterCurrent(t *testing.T) {
	currTime := time.Now()
	rb1 := storage.RelayBackendLiveData{
		Id:        "12345",
		Address:   "1.1.1.1",
		InitAt:    currTime.Add(-10 * time.Second),
		UpdatedAt: currTime,
	}

	rb2 := storage.RelayBackendLiveData{
		Id:        "54321",
		Address:   "2.1.1.1",
		InitAt:    currTime.Add(-15 * time.Second),
		UpdatedAt: currTime,
	}

	store := storage.MatrixStoreMock{
		GetRelayBackendLiveDataFunc: func(address []string) ([]storage.RelayBackendLiveData, error) {
			return []storage.RelayBackendLiveData{rb1, rb2}, nil
		},
	}

	cfg := &Config{MasterTimeVariance: time.Second}
	svc, err := New(&store, cfg)
	assert.Nil(t, err)

	svc.currentMasterBackendAddress = rb2.Address
	//rb2 should be master
	err = svc.UpdateRelayBackendMaster()
	assert.Nil(t, err)
	assert.Equal(t, "2.1.1.1", svc.currentMasterBackendAddress)
}

func TestRelayFrontendSvc_ChooseRelayBackendMaster(t *testing.T) {
	currTime := time.Now()
	rbArr := []storage.RelayBackendLiveData{
		{
			Id:        "12345",
			Address:   "1.1.1.1",
			InitAt:    currTime.Add(-10 * time.Second),
			UpdatedAt: currTime,
		},
		{
			Id:        "54321",
			Address:   "2.1.1.1",
			InitAt:    currTime.Add(-15 * time.Second),
			UpdatedAt: currTime,
		},
		{
			Id:        "67890",
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

	address, err = svc.GetMatrixAddress(MatrixTypeValve)
	assert.Nil(t, err)
	assert.Equal(t, "http://1.1.1.1/route_matrix_valve", address)

	address, err = svc.GetMatrixAddress("dog")
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

func TestRelayFrontendSvc_CacheMatrixValve(t *testing.T) {
	testMatrix := testMatrix(t)
	bin := testMatrix.GetResponseData()
	assert.NotEqual(t, 0, len(bin))

	svc := new(RelayFrontendSvc)
	svc.routeMatrixValve = new(helpers.MatrixData)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/octet-stream")
		buffer := bytes.NewBuffer(bin)
		_, err := buffer.WriteTo(w)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
		}
	}))

	err := svc.cacheMatrixInternal(ts.URL, MatrixTypeValve)
	assert.NoError(t, err)
	assert.Equal(t, bin, svc.routeMatrixValve.GetMatrix())
}

func TestRelayFrontendSvc_GetCostMatrix(t *testing.T) {

	svc := &RelayFrontendSvc{}
	svc.costMatrix = new(helpers.MatrixData)
	testMatrix := testMatrix(t)
	svc.costMatrix.SetMatrix(testMatrix.GetResponseData())

	ts := httptest.NewServer(http.HandlerFunc(svc.GetCostMatrix()))

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

func TestRelayFrontendSvc_GetRouteMatrix(t *testing.T) {

	svc := &RelayFrontendSvc{}
	svc.routeMatrix = new(helpers.MatrixData)
	testMatrix := testMatrix(t)

	svc.routeMatrix.SetMatrix(testMatrix.GetResponseData())

	ts := httptest.NewServer(http.HandlerFunc(svc.GetRouteMatrix()))

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

func TestRelayFrontendSvc_GetRouteMatrixValve(t *testing.T) {

	svc := &RelayFrontendSvc{}
	svc.routeMatrixValve = new(helpers.MatrixData)

	testMatrix := testMatrix(t)
	svc.routeMatrixValve.SetMatrix(testMatrix.GetResponseData())

	ts := httptest.NewServer(http.HandlerFunc(svc.GetRouteMatrixValve()))

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
