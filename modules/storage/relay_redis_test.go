package storage

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
)

func relayRedisTestHelperRedisStore(t *testing.T) (*RedisRelayStore, *miniredis.Miniredis) {
	rSvr, err := miniredis.Run()
	assert.Nil(t, err)
	store, err := NewRedisRelayStore(rSvr.Addr(), 200*time.Millisecond, 200*time.Millisecond, 5*time.Second)
	assert.Nil(t, err)
	return store, rSvr
}

func TestRedisRelayStore(t *testing.T) {
	rs, ms := relayRedisTestHelperRedisStore(t)
	assert.NotNil(t, rs.pool)
	assert.Equal(t, 5*time.Second, rs.relayTimeout)
	rs.pool.Close()
	ms.Close()
}

func TestNewRedisRelayStoreClose(t *testing.T) {
	rs, ms := relayRedisTestHelperRedisStore(t)
	assert.NotNil(t, rs.pool)

	err := rs.Close()
	assert.Nil(t, err)
	ms.Close()
}

func TestRedisRelayStoreSuite(t *testing.T) {
	rs, ms := relayRedisTestHelperRedisStore(t)
	assert.NotNil(t, rs.pool)

	ts := &RelayStoreTestSuite{}
	ts.RunAll(t, rs)

	err := rs.Close()
	assert.Nil(t, err)
	ms.Close()
}
