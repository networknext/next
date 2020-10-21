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
		{1, time.Now().Add(time.Duration(-50) * time.Second), time.Now().Add(time.Duration(-5) * time.Second), []byte("optimizer1")},
		{2, time.Now().Add(time.Duration(-20) * time.Second), time.Now().Add(time.Duration(-1) * time.Second), []byte("optimizer2")},
		{3, time.Now().Add(time.Duration(-40) * time.Second), time.Now().Add(time.Duration(-3) * time.Second), []byte("optimizer3")},
	}
}

func TestNew(t *testing.T) {
	store := storage.MatrixStoreMock{}
	svc, err := New(&store, 10, 15)
	assert.Nil(t, err)
	assert.NotNil(t, svc)
	assert.Equal(t, int64(10), svc.matrixSvcTimeVariance)
	assert.Equal(t, int64(15), svc.optimizerTimeVariance)
	assert.Equal(t, &store, svc.store)
	assert.False(t, svc.currentlyMaster)
	assert.NotEqual(t, 0, svc.id)
}

func TestRouteMatrixSvc_UpdateSvcDB(t *testing.T) {
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
}

func TestRouteMatrixSvc_AmMaster(t *testing.T) {
	store := storage.MatrixStoreMock{}
	svc, _ := New(&store, 10, 15)
	assert.False(t, svc.AmMaster())
	svc.currentlyMaster = true
	assert.True(t, svc.AmMaster())
}

func TestRouteMatrixSvc_DetermineMaster_NotMaster(t *testing.T) {
	store := storage.MatrixStoreMock{
		GetMatrixSvcsFunc: func() (data []storage.MatrixSvcData, u uint64, e error){
			return testMatrixSvcData(), 3,nil
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
	store := storage.MatrixStoreMock{
		GetMatrixSvcsFunc: func() (data []storage.MatrixSvcData, u uint64, e error){
			return testMatrixSvcData(), 3,nil
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
	store := storage.MatrixStoreMock{
		GetMatrixSvcsFunc: func() (data []storage.MatrixSvcData, u uint64, e error){
			return testMatrixSvcData(), 3,nil
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
	store := storage.MatrixStoreMock{
		GetMatricesFunc: func() (matrices []storage.Matrix, u uint64, e error) {
			return testOptimizerMatrices(), 3, nil
		},
		UpdateOptimizerMasterFunc: func(id uint64) error {
			return fmt.Errorf("should not be called")
		},
		UpdateLiveMatrixFunc: func(matrixData []byte) error {
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
	store := storage.MatrixStoreMock{
		GetMatricesFunc: func() (matrices []storage.Matrix, u uint64, e error) {
			return testOptimizerMatrices(), 3, nil
		},
		UpdateOptimizerMasterFunc: func(id uint64) error {
			if id != 2{
				return fmt.Errorf("wrong optimizer: %v", id)
			}
			return nil
		},
		UpdateLiveMatrixFunc: func(matrixData []byte) error {
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


func TestRouteMatrixSvc_isMasterMatrixSvcValid(t *testing.T) {
	matrices := testMatrixSvcData()

	assert.True(t, isMasterMatrixSvcValid(matrices, 2, 2000))
	assert.False(t, isMasterMatrixSvcValid(matrices, 1, 2000))
	assert.False(t, isMasterMatrixSvcValid(matrices, 50, 2000))
}

func TestRouteMatrixSvc_isMasterOptimizerValid(t *testing.T) {
	matrices := testOptimizerMatrices()

	assert.True(t, isMasterOptimizerValid(matrices, 2, 2000))
	assert.False(t, isMasterOptimizerValid(matrices, 1, 2000))
	assert.False(t, isMasterOptimizerValid(matrices, 50, 2000))
}

func TestRouteMatrixSvc_chooseMatrixSvcMaster(t *testing.T) {
	matrices := testMatrixSvcData()

	assert.Equal(t, uint64(3), chooseMatrixSvcMaster(matrices, 4000))
	assert.Equal(t, uint64(2), chooseMatrixSvcMaster(matrices, 2000))
	assert.Equal(t, uint64(1), chooseMatrixSvcMaster(matrices, 6000))
	assert.Equal(t, uint64(0), chooseMatrixSvcMaster(matrices, 500))
}

func TestRouteMatrixSvc_chooseOptimizerMaster(t *testing.T) {
	matrices := testOptimizerMatrices()

	assert.Equal(t, uint64(3), chooseOptimizerMaster(matrices, 4000))
	assert.Equal(t, uint64(2), chooseOptimizerMaster(matrices, 2000))
	assert.Equal(t, uint64(1), chooseOptimizerMaster(matrices, 6000))
	assert.Equal(t, uint64(0), chooseOptimizerMaster(matrices, 500))
}
