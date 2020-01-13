package transport_test

import (
	"bytes"
	"encoding/binary"
	"log"
	"math"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v7"
	"github.com/networknext/backend/core"
	"github.com/networknext/backend/encoding"
	"github.com/networknext/backend/transport"
	"github.com/stretchr/testify/assert"
)

const (
	sizeOfUpdateRequestVersion = 4
	sizeOfRelayToken           = 32
	sizeOfNumberOfRelays       = 4
	sizeOfRelayPingStat        = 20
)

func putUpdateRequestVersion(buff []byte) {
	const gUpdateRequestVersion = 0
	binary.LittleEndian.PutUint32(buff, gUpdateRequestVersion)
}

func putUpdateRelayAddress(buff []byte, address string) {
	offset := sizeOfUpdateRequestVersion
	binary.LittleEndian.PutUint32(buff[offset:], uint32(len(address)))
	copy(buff[offset+4:], address)
}

// doesn't actually insert anything
func putStubbedPingStats(buff []byte, addressLength int, count uint64) {
	offset := sizeOfUpdateRequestVersion + 4 + addressLength + sizeOfRelayToken
	binary.LittleEndian.PutUint64(buff[offset:], count)
}

func putPingStats(buff []byte, addressLength int, addrs ...string) {
	offset := sizeOfUpdateRequestVersion + 4 + addressLength + sizeOfRelayToken

	binary.LittleEndian.PutUint64(buff[offset:], uint64(len(addrs)))
	offset += 4

	for _, addr := range addrs {
		id := core.GetRelayID(addr)
		rtt := rand.Float32()
		jitter := rand.Float32()
		packetLoss := rand.Float32()

		binary.LittleEndian.PutUint64(buff[offset:], id)
		offset += 8
		binary.LittleEndian.PutUint32(buff[offset:], math.Float32bits(rtt))
		offset += 4
		binary.LittleEndian.PutUint32(buff[offset:], math.Float32bits(jitter))
		offset += 4
		binary.LittleEndian.PutUint32(buff[offset:], math.Float32bits(packetLoss))
		offset += 4
	}
}

func seedRedis(redisServer *miniredis.Miniredis, addressesToAdd []string) {
	addEntry := func(addr string) {
		relay := transport.NewRelayData(addr)
		bin, _ := relay.MarshalBinary()
		redisServer.HSet(transport.RedisHashName, transport.IDToRedisKey(relay.ID), string(bin))
	}

	for _, addr := range addressesToAdd {
		addEntry(addr)
	}

}

func relayUpdateAssertions(t *testing.T, body []byte, expectedCode int, redisClient *redis.Client, statsdb *core.StatsDatabase) *httptest.ResponseRecorder {
	if redisClient == nil {
		_, redisClient = NewTestRedis()
	}

	if statsdb == nil {
		statsdb = core.NewStatsDatabase()
	}

	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("POST", "/relay_update", bytes.NewBuffer(body))

	handler := transport.RelayUpdateHandlerFunc(redisClient, statsdb)

	handler(recorder, request)

	assert.Equal(t, expectedCode, recorder.Code)

	return recorder
}

func TestRelayUpdateHandler(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	t.Run("missing request version", func(t *testing.T) {
		buff := make([]byte, 0)
		relayUpdateAssertions(t, buff, http.StatusBadRequest, nil, nil)
	})

	t.Run("missing relay address", func(t *testing.T) {
		t.Run("byte array is not proper length", func(t *testing.T) {
			buff := make([]byte, sizeOfUpdateRequestVersion)
			putUpdateRequestVersion(buff)
			relayUpdateAssertions(t, buff, http.StatusBadRequest, nil, nil)
		})

		t.Run("byte array is proper length but value is empty string", func(t *testing.T) {
			addr := ""
			buff := make([]byte, sizeOfUpdateRequestVersion+4+len(addr))
			putUpdateRequestVersion(buff)
			putUpdateRelayAddress(buff, addr)
			relayUpdateAssertions(t, buff, http.StatusBadRequest, nil, nil)
		})
	})

	t.Run("missing relay token", func(t *testing.T) {
		addr := "127.0.0.1"
		buff := make([]byte, sizeOfUpdateRequestVersion+4+len(addr))
		putUpdateRequestVersion(buff)
		putUpdateRelayAddress(buff, addr)
		relayUpdateAssertions(t, buff, http.StatusBadRequest, nil, nil)
	})

	t.Run("missing number of relays", func(t *testing.T) {
		addr := "127.0.0.1"
		buff := make([]byte, sizeOfUpdateRequestVersion+4+len(addr)+sizeOfRelayToken)
		putUpdateRequestVersion(buff)
		putUpdateRelayAddress(buff, addr)
		relayUpdateAssertions(t, buff, http.StatusBadRequest, nil, nil)
	})

	t.Run("number of relays exceeds max", func(t *testing.T) {
		numRelays := 1025
		addr := "127.0.0.1"
		buff := make([]byte, sizeOfUpdateRequestVersion+4+len(addr)+sizeOfRelayToken+sizeOfNumberOfRelays+numRelays*sizeOfRelayPingStat)
		putUpdateRequestVersion(buff)
		putUpdateRelayAddress(buff, addr)
		putStubbedPingStats(buff, len(addr), uint64(numRelays))
		relayUpdateAssertions(t, buff, http.StatusBadRequest, nil, nil)
	})

	t.Run("relay not found", func(t *testing.T) {
		numRelays := 3
		addr := "127.0.0.1"
		buff := make([]byte, sizeOfUpdateRequestVersion+4+len(addr)+sizeOfRelayToken+sizeOfNumberOfRelays+numRelays*sizeOfRelayPingStat)
		putUpdateRequestVersion(buff)
		putUpdateRelayAddress(buff, addr)
		putStubbedPingStats(buff, len(addr), uint64(numRelays))
		relayUpdateAssertions(t, buff, http.StatusNotFound, nil, nil)
	})

	t.Run("valid", func(t *testing.T) {
		redisServer, redisClient := NewTestRedis()
		numRelays := 4
		addr := "127.0.0.1"
		buff := make([]byte, sizeOfUpdateRequestVersion+4+len(addr)+sizeOfRelayToken+sizeOfNumberOfRelays+numRelays*sizeOfRelayPingStat)
		putUpdateRequestVersion(buff)
		putUpdateRelayAddress(buff, addr)

		testAddrs := []string{"127.0.0.2", "127.0.0.3", "127.0.0.4", "127.0.0.5"}
		putPingStats(buff, len(addr), testAddrs...)

		seedRedis(redisServer, testAddrs)

		entry := transport.RelayData{
			ID:             core.GetRelayID(addr),
			Name:           addr,
			Address:        addr,
			Datacenter:     1,
			DatacenterName: "some name",
			PublicKey:      make([]byte, transport.LengthOfRelayToken),
			LastUpdateTime: uint64(time.Now().Unix() - 1),
		}

		raw, _ := entry.MarshalBinary()
		redisServer.HSet(transport.RedisHashName, transport.IDToRedisKey(core.GetRelayID(addr)), string(raw))

		recorder := relayUpdateAssertions(t, buff, http.StatusOK, redisClient, nil)

		res := redisClient.HGet(transport.RedisHashName, transport.IDToRedisKey(core.GetRelayID(addr)))
		var actual transport.RelayData
		raw, _ = res.Bytes()
		actual.UnmarshalBinary(raw)

		assert.Equal(t, entry.ID, actual.ID)
		assert.Equal(t, entry.Name, actual.Name)
		assert.Equal(t, entry.Address, actual.Address)
		assert.Equal(t, entry.Datacenter, actual.Datacenter)
		assert.Equal(t, entry.DatacenterName, actual.DatacenterName)
		assert.Equal(t, entry.PublicKey, actual.PublicKey)
		assert.NotEqual(t, entry.LastUpdateTime, actual.LastUpdateTime)

		// response assertions
		header := recorder.Header()
		contentType, _ := header["Content-Type"]
		assert.Equal(t, "application/octet-stream", contentType[0])

		indx := 0
		body := recorder.Body.Bytes()

		var version uint32
		encoding.ReadUint32(body, &indx, &version)

		var numRelaysToPing uint32
		encoding.ReadUint32(body, &indx, &numRelaysToPing)

		assert.Equal(t, uint32(len(testAddrs)), numRelaysToPing)

		relaysToPingIDs := make([]uint64, 0)
		relaysToPingAddrs := make([]string, 0)

		for i := 0; uint32(i) < numRelaysToPing; i++ {
			var id uint64
			var addr string
			encoding.ReadUint64(body, &indx, &id)
			encoding.ReadString(body, &indx, &addr, transport.MaxRelayAddressLength)
			relaysToPingIDs = append(relaysToPingIDs, id)
			relaysToPingAddrs = append(relaysToPingAddrs, addr)
		}

		for _, addr := range testAddrs {
			assert.Contains(t, relaysToPingIDs, core.GetRelayID(addr))
			assert.Contains(t, relaysToPingAddrs, addr)
		}

		assert.NotContains(t, relaysToPingIDs, core.GetRelayID(addr))
		assert.NotContains(t, relaysToPingAddrs, addr)
	})
}
