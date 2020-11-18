package storage

import (
	"github.com/stretchr/testify/assert"
	"net"
	"testing"
)

type RelayStoreTestSuite struct{}

func (r *RelayStoreTestSuite)RunAll(t *testing.T, store RelayStore){
	r.testSetGetDelete(t, store)
	r.testGetAll(t, store)
	r.testExpire(t, store)
}

func (r *RelayStoreTestSuite)testSetGetDelete(t *testing.T, store RelayStore){
	relays := r.relayTestData()
	
	err:=store.Set(*relays[0])
	assert.Nil(t, err)
	
	relay, err := store.Get(relays[0].ID)
	assert.Nil(t, err)
	assert.Equal(t,relays[0],relay)
	
	relay, err = store.Get(12345)
	assert.NotNil(t, err)
	assert.Equal(t, "unable to find relay data", err.Error())
	assert.Nil(t, relay)
	
	err = store.Delete(relays[0].ID)
	assert.Nil(t, err)

	err = store.Delete(relays[0].ID)
	assert.Nil(t, err)
	
	relay, err = store.Get(relays[0].ID)
	assert.NotNil(t, err)
	assert.Equal(t, "unable to find relay data", err.Error())
	assert.Nil(t, relay)
}

func (r *RelayStoreTestSuite) testGetAll(t *testing.T, store RelayStore){
	relays := r.relayTestData()
	
	for _, relay := range relays{
		err := store.Set(*relay)
		assert.Nil(t, err)
	}
	
	relayArr, err:= store.GetAll()
	assert.Nil(t, err)
	
	for _, relay := range relays{
		found := false
		for _, sr := range relayArr{
			if sr.ID == relay.ID {
				found = true
			}
		}
		assert.True(t,found)
	}
}

func (r *RelayStoreTestSuite)testExpire(t *testing.T, store RelayStore){
	relays := r.relayTestData()

	err := store.Set(*relays[1])
	assert.Nil(t,err)

	err = store.ExpireReset(relays[1].ID)
	assert.Nil(t, err)

	err = store.ExpireReset(12345)
	assert.NotNil(t, err)
}

func (r *RelayStoreTestSuite)relayTestData() []*RelayStoreData{
	testData := make([]*RelayStoreData,3)
	testData[0] = NewRelayStoreData(0, "v0", net.UDPAddr{})
	testData[1] = NewRelayStoreData(1, "v1", net.UDPAddr{})
	testData[2] = NewRelayStoreData(2, "v2", net.UDPAddr{})
	return testData
}