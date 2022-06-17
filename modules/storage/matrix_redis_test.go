package storage

// todo: convert to functional tests

/*
import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
)

func matrixRedisTestHelperRedisStore(t *testing.T) *RedisMatrixStore {
	rSvr, err := miniredis.Run()
	assert.NoError(t, err)
	store, err := NewRedisMatrixStore(rSvr.Addr(), "", 5, 5, 500*time.Millisecond, 500*time.Millisecond, 5*time.Second)
	assert.NoError(t, err)
	return store
}

func TestNewRedisMatrix_New(t *testing.T) {
	t.Parallel()

	store := matrixRedisTestHelperRedisStore(t)
	assert.NotNil(t, store)
	assert.Equal(t, 5*time.Second, store.matrixTimeout)
}

func TestRedisMatrixStore_Close(t *testing.T) {
	t.Parallel()

	store := matrixRedisTestHelperRedisStore(t)

	err := store.Close()
	assert.NoError(t, err)
}

func TestRedisMatrixStore_MatrixTestSuite_RelayBackendLiveData(t *testing.T) {
	t.Parallel()

	store := matrixRedisTestHelperRedisStore(t)
	defer func() {
		err := store.Close()
		assert.NoError(t, err)
	}()

	currTime := time.Now()
	ld := NewRelayBackendLiveData("12345", "1.1.1.1", currTime.Add(-10*time.Minute), currTime)
	ld2 := NewRelayBackendLiveData("54321", "2.2.2.2", ld.UpdatedAt.Add(-5*time.Minute), currTime)

	rbArr, err := store.GetRelayBackendLiveData()
	assert.Error(t, err)

	err = store.SetRelayBackendLiveData(ld)
	assert.NoError(t, err)

	err = store.SetRelayBackendLiveData(ld2)
	assert.NoError(t, err)

	rbArr, err = store.GetRelayBackendLiveData()
	assert.NoError(t, err)
	assert.Equal(t, ld.ID, rbArr[0].ID)
	assert.Equal(t, ld2.ID, rbArr[1].ID)
}
*/