package route_matrix

import (
	"fmt"
	"testing"
	"time"

	"github.com/networknext/backend/storage"
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
		{2, time.Now().Add(time.Duration(-20) * time.Second), time.Now().Add(time.Duration(-1) * time.Second), storage.MatrixTypeNormal,[]byte("optimizer2")},
		{2, time.Now().Add(time.Duration(-25) * time.Second), time.Now().Add(time.Duration(-1) * time.Second), storage.MatrixTypeValve, []byte("optimizer2Valve")},
		{3, time.Now().Add(time.Duration(-40) * time.Second), time.Now().Add(time.Duration(-3) * time.Second), storage.MatrixTypeNormal,[]byte("optimizer3")},
		{3, time.Now().Add(time.Duration(-45) * time.Second), time.Now().Add(time.Duration(-3) * time.Second), storage.MatrixTypeValve, []byte("optimizer3Valve")},
	}
}

func TestNew(t *testing.T) {
	t.Parallel()
	store := storage.MatrixStoreMock{}
	svc, err := New(&store, 10, 15)
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
	createdTime := time.Now().Add(-5*time.Second)
	store := storage.MatrixStoreMock{UpdateMatrixSvcFunc: func(matrixSvcData storage.MatrixSvcData) error {
		if matrixSvcData.ID != 5{
			return fmt.Errorf("not the right service id")
		}
		if matrixSvcData.CreatedAt != createdTime{
			return fmt.Errorf("not correct created at time")
		}
		return nil
	},}
	svc, err := New(&store,5 ,5)
	assert.Nil(t, err)
	svc.id = 5
	svc.createdAt = createdTime

	err = svc.UpdateSvcDB()
	assert.Nil(t, err)
}

func TestRouteMatrixSvc_AmMaster(t *testing.T) {
	t.Parallel()
	store := storage.MatrixStoreMock{}
	svc, _ := New(&store, 10, 15)
	assert.False(t, svc.AmMaster())
	svc.currentlyMaster = true
	assert.True(t, svc.AmMaster())
}

func TestRouteMatrixSvc_DetermineMaster_NotMaster(t *testing.T) {
	t.Parallel()
	store := storage.MatrixStoreMock{
		GetMatrixSvcsFunc: func() (data []storage.MatrixSvcData, e error){
			return testMatrixSvcData(), nil
		},
		GetMatrixSvcMasterFunc: func ()(uint64,error){
			return 3, nil
		},
		UpdateMatrixSvcMasterFunc: func(uint64) error{
			return fmt.Errorf("should not be called")
		},
	}
	svc, err := New(&store, 4000, 15)
	assert.Nil(t, err)
	svc.id = 1

	err = svc.DetermineMaster()
	assert.Nil(t,err)
	assert.False(t,svc.currentlyMaster)
}

func TestRouteMatrixSvc_DetermineMaster_ChosenMasterNotCurrent(t *testing.T) {
	t.Parallel()
	store := storage.MatrixStoreMock{
		GetMatrixSvcsFunc: func() (data []storage.MatrixSvcData, e error){
			return testMatrixSvcData(), nil
		},
		GetMatrixSvcMasterFunc: func ()(uint64,error){
			return 3, nil
		},
		UpdateMatrixSvcMasterFunc: func(uint64) error{
			return nil
		},
	}
	svc, err := New(&store, 2000, 15)
	assert.Nil(t, err)
	svc.id = 2
	assert.False(t,svc.currentlyMaster)
	err = svc.DetermineMaster()
	assert.Nil(t,err)
	assert.True(t,svc.currentlyMaster)
}

func TestRouteMatrixSvc_DetermineMaster_IsCurrentMaster(t *testing.T) {
	t.Parallel()
	store := storage.MatrixStoreMock{
		GetMatrixSvcsFunc: func() (data []storage.MatrixSvcData, e error){
			return testMatrixSvcData(), nil
		},
		GetMatrixSvcMasterFunc: func ()(uint64,error){
			return 3, nil
		},
		UpdateMatrixSvcMasterFunc: func(uint64) error{
			return fmt.Errorf("should not be called")
		},
	}
	svc, err := New(&store, 4000, 15)
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
		GetOptimizerMasterFunc: func ()(uint64, error){
			return 3, nil
		},
		UpdateOptimizerMasterFunc: func(id uint64) error {
			return fmt.Errorf("should not be called")
		},
		UpdateLiveMatrixFunc: func(matrixData []byte, matrixType string) error {
			if string(matrixData) != "optimizer3"{
				return fmt.Errorf("not the correct matrix: %s", string(matrixData))
			}
			return nil
		},
	}

	svc, err:= New(&store, 10, 4000)
	assert.Nil(t, err)
	svc.currentMasterOptimizer = 3

	err = svc.UpdateLiveRouteMatrix()
	assert.Nil(t, err)
}

func TestRouteMatrixSvc_UpdateLiveRouteMatrix_ChooseOptimizerMaster(t *testing.T) {
	t.Parallel()
	store := storage.MatrixStoreMock{
		GetOptimizerMatricesFunc: func() (matrices []storage.Matrix, e error) {
			return testOptimizerMatrices(), nil
		},
		GetOptimizerMasterFunc: func ()(uint64,error){
			return 3, nil
		},
		UpdateOptimizerMasterFunc: func(id uint64) error {
			if id != 2{
				return fmt.Errorf("wrong optimizer: %v", id)
			}
			return nil
		},
		UpdateLiveMatrixFunc: func(matrixData []byte, matrixType string) error {
			if string(matrixData) != "optimizer2"{
				return fmt.Errorf("not the correct matrix: %s", string(matrixData))
			}
			return nil
		},
	}

	svc, err:= New(&store, 10, 2000)
	assert.Nil(t, err)
	svc.currentMasterOptimizer = 3

	err = svc.UpdateLiveRouteMatrix()
	assert.Nil(t, err)
	assert.Equal(t, uint64(2), svc.currentMasterOptimizer)
}

func timeVariance(value int)time.Duration{
	return time.Duration(value)*time.Millisecond
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
		DeleteMatrixSvcFunc: func(id uint64) (e error){
			if id != 1{
				return fmt.Errorf("should not have been called for matrix svc id %v", id)
			}
			return nil
		},
		DeleteOptimizerMatrixFunc: func(id uint64, matrixType string) (e error){
			if id == 2{
				return fmt.Errorf("should not have been called for optimizer id %v", id)
			}
			return nil
		},
	}

	svc, err:= New(&store, 4000, 2000)
	assert.Nil(t, err)
	
	err = svc.CleanUpDB()
	assert.Nil(t, err)
}