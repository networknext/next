package storage

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

type matrixTestSuite struct{}

func (ts *matrixTestSuite) RunAll(t *testing.T, store MatrixStore) {
	ts.TestRelayBackendLiveData(t, store)
}

func (ts *matrixTestSuite) TestRelayBackendLiveData(t *testing.T, store MatrixStore) {
	currTime := time.Now()
	ld := NewRelayBackendLiveData("12345", "1.1.1.1", currTime.Add(-10*time.Minute), currTime)
	ld2 := NewRelayBackendLiveData("54321", "2.2.2.2", ld.UpdatedAt.Add(-5*time.Minute), currTime)

	rbArr, err := store.GetRelayBackendLiveData()
	assert.NotNil(t, err)

	err = store.SetRelayBackendLiveData(ld)
	assert.Nil(t, err)

	err = store.SetRelayBackendLiveData(ld2)
	assert.Nil(t, err)

	rbArr, err = store.GetRelayBackendLiveData()
	assert.Nil(t, err)
	assert.Equal(t, ld.Id, rbArr[0].Id)
	assert.Equal(t, ld2.Id, rbArr[1].Id)
}
