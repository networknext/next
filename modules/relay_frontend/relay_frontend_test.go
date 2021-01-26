package relay_frontend

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/networknext/backend/modules/encoding"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/routing"

	"github.com/networknext/backend/modules/storage"
	"github.com/stretchr/testify/assert"
)

func testMatrixSvcData() []storage.MatrixSvcData {
	return []storage.MatrixSvcData{
		{1, time.Now().Add(-50 * time.Second), time.Now().Add(-5 * time.Second)},
		{2, time.Now().Add(-2 * time.Second), time.Now().Add(-1 * time.Second)},
		{3, time.Now().Add(-40 * time.Second), time.Now().Add(-3 * time.Second)},
	}
}

func testOptimizerMatrices() []storage.Matrix {
	return []storage.Matrix{
		{1, time.Now().Add(time.Duration(-50) * time.Second), time.Now().Add(time.Duration(-5) * time.Second), storage.MatrixTypeNormal, []byte("optimizer1")},
		{1, time.Now().Add(time.Duration(-49) * time.Second), time.Now().Add(time.Duration(-5) * time.Second), storage.MatrixTypeValve, []byte("optimizer1Valve")},
		{2, time.Now().Add(time.Duration(-20) * time.Second), time.Now().Add(time.Duration(-1) * time.Second), storage.MatrixTypeNormal, []byte("optimizer2")},
		{2, time.Now().Add(time.Duration(-25) * time.Second), time.Now().Add(time.Duration(-1) * time.Second), storage.MatrixTypeValve, []byte("optimizer2Valve")},
		{3, time.Now().Add(time.Duration(-40) * time.Second), time.Now().Add(time.Duration(-3) * time.Second), storage.MatrixTypeNormal, []byte("optimizer3")},
		{3, time.Now().Add(time.Duration(-45) * time.Second), time.Now().Add(time.Duration(-3) * time.Second), storage.MatrixTypeValve, []byte("optimizer3Valve")},
	}
}

func TestNew(t *testing.T) {
	t.Parallel()
	store := storage.MatrixStoreMock{}
	svc, err := New(&store, timeVariance(10), timeVariance(15))
	assert.Nil(t, err)
	assert.NotNil(t, svc)
	assert.Equal(t, timeVariance(10), svc.matrixSvcTimeVariance)
	assert.Equal(t, timeVariance(15), svc.optimizerTimeVariance)
	assert.Equal(t, &store, svc.store)
	assert.False(t, svc.currentlyMaster)
	assert.NotEqual(t, 0, svc.id)
}

func TestRouteMatrixSvc_UpdateSvcDB(t *testing.T) {
	t.Parallel()
	createdTime := time.Now().Add(-5 * time.Second)
	store := storage.MatrixStoreMock{UpdateMatrixSvcFunc: func(matrixSvcData storage.MatrixSvcData) error {
		if matrixSvcData.ID != 5 {
			return fmt.Errorf("not the right service id")
		}
		if matrixSvcData.CreatedAt != createdTime {
			return fmt.Errorf("not correct created at time")
		}
		return nil
	}}
	svc, err := New(&store, 5, 5)
	assert.Nil(t, err)
	svc.id = 5
	svc.createdAt = createdTime

	err = svc.UpdateSvcDB()
	assert.Nil(t, err)
}

func TestRouteMatrixSvc_AmMaster(t *testing.T) {
	t.Parallel()
	store := storage.MatrixStoreMock{}
	svc, _ := New(&store, timeVariance(10), timeVariance(15))
	assert.False(t, svc.AmMaster())
	svc.currentlyMaster = true
	assert.True(t, svc.AmMaster())
}

func TestRouteMatrixSvc_DetermineMaster_NotMaster(t *testing.T) {
	t.Parallel()
	store := storage.MatrixStoreMock{
		GetMatrixSvcsFunc: func() (data []storage.MatrixSvcData, e error) {
			return testMatrixSvcData(), nil
		},
		GetMatrixSvcMasterFunc: func() (uint64, error) {
			return 3, nil
		},
		UpdateMatrixSvcMasterFunc: func(uint64) error {
			return fmt.Errorf("should not be called")
		},
	}
	svc, err := New(&store, timeVariance(4000), timeVariance(15))
	assert.Nil(t, err)
	svc.id = 1

	err = svc.DetermineMaster()
	assert.Nil(t, err)
	assert.False(t, svc.currentlyMaster)
}

func TestRouteMatrixSvc_DetermineMaster_ChosenMasterNotCurrent(t *testing.T) {
	t.Parallel()
	store := storage.MatrixStoreMock{
		GetMatrixSvcsFunc: func() (data []storage.MatrixSvcData, e error) {
			return testMatrixSvcData(), nil
		},
		GetMatrixSvcMasterFunc: func() (uint64, error) {
			return 3, nil
		},
		UpdateMatrixSvcMasterFunc: func(uint64) error {
			return nil
		},
	}
	svc, err := New(&store, timeVariance(2000), timeVariance(15))
	assert.Nil(t, err)
	svc.id = 2
	assert.False(t, svc.currentlyMaster)
	err = svc.DetermineMaster()
	assert.Nil(t, err)
	assert.True(t, svc.currentlyMaster)
}

func TestRouteMatrixSvc_DetermineMaster_IsCurrentMaster(t *testing.T) {
	t.Parallel()
	store := storage.MatrixStoreMock{
		GetMatrixSvcsFunc: func() (data []storage.MatrixSvcData, e error) {
			return testMatrixSvcData(), nil
		},
		GetMatrixSvcMasterFunc: func() (uint64, error) {
			return 3, nil
		},
		UpdateMatrixSvcMasterFunc: func(uint64) error {
			return fmt.Errorf("should not be called")
		},
	}
	svc, err := New(&store, timeVariance(4000), timeVariance(15))
	assert.Nil(t, err)
	svc.id = 3
	svc.currentlyMaster = true
	err = svc.DetermineMaster()
	assert.Nil(t, err)
	assert.True(t, svc.currentlyMaster)
}

func TestRouteMatrixSvc_UpdateLiveRouteMatrix_OptimizerMasterCurrent(t *testing.T) {
	t.Parallel()
	store := storage.MatrixStoreMock{
		GetOptimizerMatricesFunc: func() (matrices []storage.Matrix, e error) {
			return testOptimizerMatrices(), nil
		},
		GetOptimizerMasterFunc: func() (uint64, error) {
			return 3, nil
		},
		UpdateOptimizerMasterFunc: func(id uint64) error {
			return fmt.Errorf("should not be called")
		},
		UpdateLiveMatrixFunc: func(matrixData []byte, matrixType string) error {
			if string(matrixData) == "optimizer3" || string(matrixData) == "optimizer3Valve" {
				return nil
			}
			return fmt.Errorf("not the correct matrix: %s", string(matrixData))
		},
	}

	svc, err := New(&store, timeVariance(10), timeVariance(4000))
	assert.Nil(t, err)
	svc.currentMasterOptimizer = 3

	err = svc.UpdateLiveRouteMatrixOptimizer()
	assert.Nil(t, err)
}

func TestRouteMatrixSvc_UpdateLiveRouteMatrix_ChooseOptimizerMaster(t *testing.T) {
	t.Parallel()
	store := storage.MatrixStoreMock{
		GetOptimizerMatricesFunc: func() (matrices []storage.Matrix, e error) {
			return testOptimizerMatrices(), nil
		},
		GetOptimizerMasterFunc: func() (uint64, error) {
			return 3, nil
		},
		UpdateOptimizerMasterFunc: func(id uint64) error {
			if id != 2 {
				return fmt.Errorf("wrong optimizer: %v", id)
			}
			return nil
		},
		UpdateLiveMatrixFunc: func(matrixData []byte, matrixType string) error {
			if string(matrixData) == "optimizer2" || string(matrixData) == "optimizer2Valve" {
				return nil
			}
			return fmt.Errorf("not the correct matrix: %s", string(matrixData))
		},
	}

	svc, err := New(&store, timeVariance(10), timeVariance(2000))
	assert.Nil(t, err)
	svc.currentMasterOptimizer = 3

	err = svc.UpdateLiveRouteMatrixOptimizer()
	assert.Nil(t, err)
	assert.Equal(t, uint64(2), svc.currentMasterOptimizer)
}

func timeVariance(value int) time.Duration {
	return time.Duration(value) * time.Millisecond
}

func TestRouteMatrixSvc_isMasterMatrixSvcValid(t *testing.T) {
	t.Parallel()
	matrices := testMatrixSvcData()

	assert.True(t, isMasterMatrixSvcValid(matrices, 2, timeVariance(2000)))
	assert.False(t, isMasterMatrixSvcValid(matrices, 1, timeVariance(2000)))
	assert.False(t, isMasterMatrixSvcValid(matrices, 50, timeVariance(2000)))
}

func TestRouteMatrixSvc_isMasterOptimizerValid(t *testing.T) {
	t.Parallel()
	matrices := testOptimizerMatrices()

	assert.True(t, isMasterOptimizerValid(matrices, 2, timeVariance(2000)))
	assert.False(t, isMasterOptimizerValid(matrices, 1, timeVariance(2000)))
	assert.False(t, isMasterOptimizerValid(matrices, 50, timeVariance(2000)))
}

func TestRouteMatrixSvc_chooseMatrixSvcMaster(t *testing.T) {
	t.Parallel()
	matrices := testMatrixSvcData()

	assert.Equal(t, uint64(3), chooseMatrixSvcMaster(matrices, timeVariance(4000)))
	assert.Equal(t, uint64(2), chooseMatrixSvcMaster(matrices, timeVariance(2000)))
	assert.Equal(t, uint64(1), chooseMatrixSvcMaster(matrices, timeVariance(6000)))
	assert.Equal(t, uint64(0), chooseMatrixSvcMaster(matrices, timeVariance(500)))
}

func TestRouteMatrixSvc_chooseOptimizerMaster(t *testing.T) {
	t.Parallel()
	matrices := testOptimizerMatrices()

	assert.Equal(t, uint64(3), chooseOptimizerMaster(matrices, timeVariance(4000)))
	assert.Equal(t, uint64(2), chooseOptimizerMaster(matrices, timeVariance(2000)))
	assert.Equal(t, uint64(1), chooseOptimizerMaster(matrices, timeVariance(6000)))
	assert.Equal(t, uint64(0), chooseOptimizerMaster(matrices, timeVariance(500)))
}

func TestRouteMatrixSvc_CleanUpDB(t *testing.T) {
	store := storage.MatrixStoreMock{
		GetMatrixSvcsFunc: func() (data []storage.MatrixSvcData, e error) {
			return testMatrixSvcData(), nil
		},
		GetOptimizerMatricesFunc: func() (matrices []storage.Matrix, e error) {
			return testOptimizerMatrices(), nil
		},
		DeleteMatrixSvcFunc: func(id uint64) (e error) {
			if id != 1 {
				return fmt.Errorf("should not have been called for matrix svc id %v", id)
			}
			return nil
		},
		DeleteOptimizerMatrixFunc: func(id uint64, matrixType string) (e error) {
			if id == 2 {
				return fmt.Errorf("should not have been called for optimizer id %v", id)
			}
			return nil
		},
	}

	svc, err := New(&store, timeVariance(4000), timeVariance(2000))
	assert.Nil(t, err)

	err = svc.CleanUpDB()
	assert.Nil(t, err)
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

	var masterRB *storage.RelayBackendLiveData
	store := storage.MatrixStoreMock{
		GetRelayBackendLiveDataFunc: func(address []string) ([]storage.RelayBackendLiveData, error) {
			return []storage.RelayBackendLiveData{rb1, rb2, rb3}, nil
		},
		GetRelayBackendMasterFunc: func() (storage.RelayBackendLiveData, error) {
			if masterRB == nil {
				return storage.RelayBackendLiveData{}, fmt.Errorf("relay backend master not found")
			}
			return *masterRB, nil
		},
		SetRelayBackendMasterFunc: func(relay storage.RelayBackendLiveData) error {
			masterRB = &relay
			return nil
		},
	}

	svc, err := New(&store, time.Second, time.Second)
	assert.Nil(t, err)

	//rb2 should be master
	err = svc.UpdateRelayBackendMaster()
	assert.Nil(t, err)
	assert.Equal(t, "2.1.1.1", svc.currentMasterBackendAddress)
	assert.Equal(t, rb2, *masterRB)

	//change to rb1 as master
	rb2.UpdatedAt = rb2.UpdatedAt.Add(-6 * time.Second)
	err = svc.UpdateRelayBackendMaster()
	assert.Nil(t, err)
	assert.Equal(t, "1.1.1.1", svc.currentMasterBackendAddress)
	assert.Equal(t, rb1, *masterRB)
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

	masterRB := rb2
	store := storage.MatrixStoreMock{
		GetRelayBackendLiveDataFunc: func(address []string) ([]storage.RelayBackendLiveData, error) {
			return []storage.RelayBackendLiveData{rb1, rb2}, nil
		},
		GetRelayBackendMasterFunc: func() (storage.RelayBackendLiveData, error) {
			return masterRB, nil
		},
		SetRelayBackendMasterFunc: func(relay storage.RelayBackendLiveData) error {
			return fmt.Errorf("should not be called")
		},
	}

	svc, err := New(&store, time.Second, time.Second)
	assert.Nil(t, err)

	svc.currentMasterBackendAddress = rb2.Address
	//rb2 should be master
	err = svc.UpdateRelayBackendMaster()
	assert.Nil(t, err)
	assert.Equal(t, "2.1.1.1", svc.currentMasterBackendAddress)
	assert.Equal(t, rb2, masterRB)
}

func TestRelayFrontendSvc_IsValidRelayBackendMaster(t *testing.T) {
	currTime := time.Now()
	rb1 := storage.RelayBackendLiveData{
		Id:        "12345",
		Address:   "1.1.1.1",
		InitAt:    currTime.Add(-10 * time.Second),
		UpdatedAt: currTime.Add(-2 * time.Second),
	}

	rb2 := storage.RelayBackendLiveData{
		Id:        "54321",
		Address:   "2.1.1.1",
		InitAt:    currTime.Add(-15 * time.Second),
		UpdatedAt: currTime,
	}

	rbArr := []storage.RelayBackendLiveData{rb1, rb2}

	valid := isMasterRelayBackendValid(rbArr, rb2.Address, time.Second)
	assert.True(t, valid)

	valid = isMasterRelayBackendValid(rbArr, rb1.Address, time.Second)
	assert.False(t, valid)

	valid = isMasterRelayBackendValid(rbArr, "fake", time.Second)
	assert.False(t, valid)
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

	address, err := svc.GetMatrixAddress(storage.MatrixTypeNormal)
	assert.Nil(t, err)
	assert.Equal(t, "http:/1.1.1.1/route_matrix", address)

	address, err = svc.GetMatrixAddress(storage.MatrixTypeValve)
	assert.Nil(t, err)
	assert.Equal(t, "http:/1.1.1.1/route_matrix_valve", address)

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

func TestRelayFrontendSvc_UpdateLiveRouteMatrixBackend(t *testing.T) {

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

	svc := &RelayFrontendSvc{}
	store := storage.MatrixStoreMock{
		UpdateLiveMatrixFunc: func(matrixData []byte, matrixType string) error {
			if string(matrixData) != string(bin) {
				return fmt.Errorf("not the matrix")
			}
			return nil
		},
	}
	svc.store = &store

	err := svc.UpdateLiveRouteMatrixBackend(ts.URL, storage.MatrixTypeNormal)
	assert.NoError(t, err)
}

func TestRelayFrontendSvc_CacheMatrixNormal(t *testing.T) {
	testMatrix := testMatrix(t)
	bin := testMatrix.GetResponseData()
	assert.NotEqual(t, 0, len(bin))

	svc := &RelayFrontendSvc{}
	store := storage.MatrixStoreMock{
		GetLiveMatrixFunc: func(matrixType string) ([]byte, error) {
			return bin, nil
		},
	}
	svc.store = &store

	err := svc.CacheMatrix(storage.MatrixTypeNormal)
	assert.NoError(t, err)
	assert.Equal(t, bin, svc.routeMatrix)
}

func TestRelayFrontendSvc_CacheMatrixValve(t *testing.T) {
	testMatrix := testMatrix(t)
	bin := testMatrix.GetResponseData()
	assert.NotEqual(t, 0, len(bin))

	svc := &RelayFrontendSvc{}
	store := storage.MatrixStoreMock{
		GetLiveMatrixFunc: func(matrixType string) ([]byte, error) {
			return bin, nil
		},
	}
	svc.store = &store

	err := svc.CacheMatrix(storage.MatrixTypeValve)
	assert.NoError(t, err)
	assert.Equal(t, bin, svc.routeMatrixValve)
}

func TestRelayFrontendSvc_GetMatrix(t *testing.T) {

	svc := &RelayFrontendSvc{}
	testMatrix := testMatrix(t)
	svc.routeMatrix = testMatrix.GetResponseData()

	ts := httptest.NewServer(http.HandlerFunc(svc.GetMatrix()))

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

func TestRelayFrontendSvc_GetMatrixValve(t *testing.T) {

	svc := &RelayFrontendSvc{}
	testMatrix := testMatrix(t)
	svc.routeMatrixValve = testMatrix.GetResponseData()

	ts := httptest.NewServer(http.HandlerFunc(svc.GetMatrixValve()))

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
