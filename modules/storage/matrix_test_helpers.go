package storage

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type matrixTestSuite struct{}

func (ts *matrixTestSuite) RunAll(t *testing.T, store MatrixStore) {
	ts.TestLiveMatrix(t, store)
	ts.TestOptimizerMatrices(t, store)
	ts.TestMatrixSvcData(t, store)
	ts.UpdateAndGetSvcMaster(t, store)
	ts.UpdateAndGetOptimizerMaster(t, store)
	ts.TestRelayBackendLiveData(t, store)
	ts.TestRelayBackendMaster(t, store)
}

func (ts *matrixTestSuite) TestLiveMatrix(t *testing.T, store MatrixStore) {
	testLiveMatrix := []byte("test live matrix")
	_, err := store.GetLiveMatrix(MatrixTypeNormal)
	assert.NotNil(t, err)
	assert.Equal(t, "matrix not found", err.Error())

	err = store.UpdateLiveMatrix(testLiveMatrix, MatrixTypeNormal)
	assert.Nil(t, err)

	testValveMatrix := []byte("test valve matrix")
	_, err = store.GetLiveMatrix(MatrixTypeValve)
	assert.NotNil(t, err)
	assert.Equal(t, "matrix not found", err.Error())

	err = store.UpdateLiveMatrix(testValveMatrix, MatrixTypeValve)
	assert.Nil(t, err)

	matrix, err := store.GetLiveMatrix(MatrixTypeNormal)
	assert.Nil(t, err)
	assert.Equal(t, string(testLiveMatrix), string(matrix))

	matrix, err = store.GetLiveMatrix(MatrixTypeValve)
	assert.Nil(t, err)
	assert.Equal(t, string(testValveMatrix), string(matrix))
}

func (ts *matrixTestSuite) TestOptimizerMatrices(t *testing.T, store MatrixStore) {
	matrices := ts.testOptimizerMatricesData()

	_, err := store.GetOptimizerMatrices()
	assert.NotNil(t, err)
	assert.Equal(t, "optimizer matrices not found", err.Error())

	for _, m := range matrices {
		err = store.UpdateOptimizerMatrix(m)
		assert.Nil(t, err)
	}

	storeMatrices, err := store.GetOptimizerMatrices()
	for _, m := range matrices {
		found := false
		for _, sm := range storeMatrices {
			if m.OptimizerID == sm.OptimizerID {
				found = true
			}
		}
		assert.True(t, found)
	}

	err = store.DeleteOptimizerMatrix(matrices[0].OptimizerID, MatrixTypeNormal)
	assert.Nil(t, err)

	storeMatrices, err = store.GetOptimizerMatrices()
	for _, sm := range storeMatrices {
		if matrices[0].OptimizerID == sm.OptimizerID {
			assert.Fail(t, "should not have been found")
		}
	}
}

func (ts *matrixTestSuite) TestMatrixSvcData(t *testing.T, store MatrixStore) {
	matrices := ts.testMatrixSvcData()

	_, err := store.GetMatrixSvcs()
	assert.NotNil(t, err)
	assert.Equal(t, "matrix svc data not found", err.Error())

	for _, m := range matrices {
		err = store.UpdateMatrixSvc(m)
		assert.Nil(t, err)
	}

	storeMatrices, err := store.GetMatrixSvcs()
	for _, m := range matrices {
		found := false
		for _, sm := range storeMatrices {
			if m.ID == sm.ID {
				found = true
			}
		}
		assert.True(t, found)
	}

	err = store.DeleteMatrixSvc(matrices[0].ID)
	assert.Nil(t, err)

	storeMatrices, err = store.GetMatrixSvcs()
	for _, sm := range storeMatrices {
		if matrices[0].ID == sm.ID {
			assert.Fail(t, "should not have been found")
		}
	}
}

func (ts *matrixTestSuite) UpdateAndGetSvcMaster(t *testing.T, store MatrixStore) {
	masterID, err := store.GetMatrixSvcMaster()
	assert.NotNil(t, err)
	assert.Equal(t, "matrix svc master not found", err.Error())
	assert.Equal(t, uint64(0), masterID)

	err = store.UpdateMatrixSvcMaster(10)
	assert.Nil(t, err)

	masterID, err = store.GetMatrixSvcMaster()
	assert.Nil(t, err)
	assert.Equal(t, uint64(10), masterID)
}

func (ts *matrixTestSuite) UpdateAndGetOptimizerMaster(t *testing.T, store MatrixStore) {
	masterID, err := store.GetOptimizerMaster()
	assert.NotNil(t, err)
	assert.Equal(t, "optimizer master not found", err.Error())
	assert.Equal(t, uint64(0), masterID)

	err = store.UpdateOptimizerMaster(25)
	assert.Nil(t, err)

	masterID, err = store.GetOptimizerMaster()
	assert.Nil(t, err)
	assert.Equal(t, uint64(25), masterID)

}

func (ts *matrixTestSuite) testMatrixSvcData() []MatrixSvcData {
	return []MatrixSvcData{
		{1, time.Now().Add(-50 * time.Second), time.Now().Add(-2 * time.Second)},
		{2, time.Now().Add(-20 * time.Second), time.Now().Add(-1 * time.Second)},
		{3, time.Now().Add(-40 * time.Second), time.Now().Add(-3 * time.Second)},
	}
}

func (ts *matrixTestSuite) testOptimizerMatricesData() []Matrix {
	return []Matrix{
		{1, time.Now().Add(-50 * time.Second), time.Now().Add(-5 * time.Second), MatrixTypeNormal, []byte("optimizer1")},
		{2, time.Now().Add(-20 * time.Second), time.Now().Add(-1 * time.Second), MatrixTypeNormal, []byte("optimizer2")},
		{3, time.Now().Add(-40 * time.Second), time.Now().Add(-3 * time.Second), MatrixTypeNormal, []byte("optimizer3")},
	}
}

func (ts *matrixTestSuite) TestRelayBackendLiveData(t *testing.T, store MatrixStore) {
	currTime := time.Now()
	ld := NewRelayBackendLiveData("12345", "1.1.1.1", currTime.Add(-10*time.Minute), currTime)
	ld2 := NewRelayBackendLiveData("54321", "2.2.2.2", ld.UpdatedAt.Add(-5*time.Minute), currTime)

	err := store.SetRelayBackendLiveData(ld)
	assert.Nil(t, err)

	err = store.SetRelayBackendLiveData(ld2)
	assert.Nil(t, err)

	rbArr, err := store.GetRelayBackendLiveData([]string{ld.Address, ld2.Address})
	assert.Nil(t, err)
	assert.Equal(t, ld.Id, rbArr[0].Id)
	assert.Equal(t, ld2.Id, rbArr[1].Id)

	rbArr, err = store.GetRelayBackendLiveData([]string{"fake", ld2.Address})
	assert.Nil(t, err)
	assert.Equal(t, ld2.Id, rbArr[0].Id)

}

func (ts *matrixTestSuite) TestRelayBackendMaster(t *testing.T, store MatrixStore) {
	currTime := time.Now()
	ld := NewRelayBackendLiveData("12345", "1.1.1.1", currTime.Add(-10*time.Minute), currTime)

	err := store.SetRelayBackendMaster(ld)
	assert.Nil(t, err)

	master, err := store.GetRelayBackendMaster()
	assert.Nil(t, err)
	assert.Equal(t, ld.Id, master.Id)
}
