package storage

import(
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
)

func matrixRedisTestHelperRedisStore(t *testing.T) *RedisMatrixStore{
	rSvr, err := miniredis.Run()
	assert.Nil(t, err)
	store, err := NewRedisMatrixStore(rSvr.Addr(),200*time.Millisecond,200*time.Millisecond,5*time.Second)
	assert.Nil(t, err)
	return store
}

func TestNewRedisMatrix_New(t *testing.T) {
	store := matrixRedisTestHelperRedisStore(t)
	assert.NotNil(t, store)
	assert.Equal(t,5*time.Second, store.matrixTimeout)
}

func TestRedisMatrixStore_Close(t *testing.T) {
	store := matrixRedisTestHelperRedisStore(t)

	err := store.Close()
	assert.Nil(t, err)
}

func TestRedisMatrixStore_MatrixTestSuite(t *testing.T) {
	store := matrixRedisTestHelperRedisStore(t)

	ts := new(matrixTestSuite)
	ts.RunAll(t, store)

	err := store.Close()
	assert.Nil(t, err)
}