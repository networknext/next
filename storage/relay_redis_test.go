package storage

import(
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
)

func relayRedisTestHelperRedisStore(t *testing.T) (*RedisRelayStore, *miniredis.Miniredis){
	rSvr, err := miniredis.Run()
	assert.Nil(t, err)
	store, err := NewRedisRelayStore(rSvr.Addr(),10,10,5)
	assert.Nil(t, err)
	return store, rSvr
}

func TestRedisRelayStore(t *testing.T) {
	rs, ms := relayRedisTestHelperRedisStore(t)
	assert.NotNil(t, rs.conn)
	assert.Equal(t,int64(5), rs.relayTimeout)

	rs.conn.Close()
	ms.Close()
}

func TestNewRedisRelayStoreClose(t *testing.T) {
	rs, ms := relayRedisTestHelperRedisStore(t)
	assert.NotNil(t, rs.conn)

	err :=rs.Close()
	assert.Nil(t, err)
	ms.Close()
}

func TestRedisRelayStoreSuite(t *testing.T) {
	rs, ms := relayRedisTestHelperRedisStore(t)
	assert.NotNil(t, rs.conn)

	ts :=&RelayStoreTestSuite{}
	ts.RunAll(t,rs)

	err :=rs.Close()
	assert.Nil(t, err)
	ms.Close()
}