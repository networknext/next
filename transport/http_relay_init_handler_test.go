package transport_test

import (
	"bytes"
	"encoding/binary"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/go-redis/redis/v7"
	"github.com/networknext/backend/core"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/encoding"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/transport"
	"github.com/stretchr/testify/assert"
)

const (
	sizeOfInitRequestMagic   = 4
	sizeOfInitRequestVersion = 4
)

func relayInitAssertions(t *testing.T, body []byte, expectedCode int, redisClient *redis.Client) *httptest.ResponseRecorder {
	if redisClient == nil {
		_, redisClient = NewTestRedis()
	}

	recorder := httptest.NewRecorder()
	request, _ := http.NewRequest("POST", "/relay_init", bytes.NewBuffer(body))

	handler := transport.RelayInitHandlerFunc(redisClient)

	handler(recorder, request)

	assert.Equal(t, expectedCode, recorder.Code)

	return recorder
}

func writeRelayAddress(buff []byte, address string) {
	offset := sizeOfInitRequestMagic + sizeOfInitRequestVersion + crypto.NonceSize
	binary.LittleEndian.PutUint32(buff[offset:], uint32(len(address)))
	copy(buff[offset+4:], address)
}

func TestRelayInitHandler(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	t.Run("missing magic number", func(t *testing.T) {
		buff := make([]byte, 0)
		relayInitAssertions(t, buff, http.StatusBadRequest, nil)
	})

	t.Run("missing request version", func(t *testing.T) {
		buff := make([]byte, sizeOfInitRequestMagic)
		binary.LittleEndian.PutUint32(buff, transport.InitRequestMagic)
		relayInitAssertions(t, buff, http.StatusBadRequest, nil)
	})

	t.Run("missing nonce bytes", func(t *testing.T) {
		buff := make([]byte, sizeOfInitRequestMagic+sizeOfInitRequestVersion)
		binary.LittleEndian.PutUint32(buff, transport.InitRequestMagic)
		binary.LittleEndian.PutUint32(buff[4:], 0)
		relayInitAssertions(t, buff, http.StatusBadRequest, nil)
	})

	t.Run("missing relay address", func(t *testing.T) {
		t.Run("byte array is not proper length", func(t *testing.T) {
			buff := make([]byte, sizeOfInitRequestMagic+sizeOfInitRequestVersion+crypto.NonceSize)
			binary.LittleEndian.PutUint32(buff, transport.InitRequestMagic)
			binary.LittleEndian.PutUint32(buff[4:], 0)
			relayInitAssertions(t, buff, http.StatusBadRequest, nil)
		})

		t.Run("byte array is proper length but there is a blank string", func(t *testing.T) {
			addr := ""
			buff := make([]byte, sizeOfInitRequestMagic+sizeOfInitRequestVersion+crypto.NonceSize+4+len(addr))
			binary.LittleEndian.PutUint32(buff, transport.InitRequestMagic)
			binary.LittleEndian.PutUint32(buff[4:], 0)
			writeRelayAddress(buff, addr)
			relayInitAssertions(t, buff, http.StatusBadRequest, nil)
		})
	})

	t.Run("missing encryption token", func(t *testing.T) {
		addr := "127.0.0.1"
		buff := make([]byte, sizeOfInitRequestMagic+sizeOfInitRequestVersion+crypto.NonceSize+4+len(addr))
		binary.LittleEndian.PutUint32(buff, transport.InitRequestMagic)
		binary.LittleEndian.PutUint32(buff[4:], 0)
		writeRelayAddress(buff, addr)
		relayInitAssertions(t, buff, http.StatusBadRequest, nil)
	})

	t.Run("encryption token is 0'ed", func(t *testing.T) {
		addr := "127.0.0.1"
		buff := make([]byte, sizeOfInitRequestMagic+sizeOfInitRequestVersion+crypto.NonceSize+4+len(addr)+routing.TokenSize)
		binary.LittleEndian.PutUint32(buff, transport.InitRequestMagic)
		binary.LittleEndian.PutUint32(buff[4:], 0)
		writeRelayAddress(buff, addr)
		relayInitAssertions(t, buff, http.StatusBadRequest, nil)
	})

	t.Run("relay already exists", func(t *testing.T) {
		name := "some name"
		addr := "127.0.0.1"
		dcname := "another name"
		pubkey := make([]byte, 32)
		entry := transport.RelayData{
			ID:             core.GetRelayID(addr),
			Name:           name,
			Address:        addr,
			Datacenter:     32,
			DatacenterName: dcname,
			PublicKey:      pubkey,
			LastUpdateTime: 1234,
		}

		data, _ := entry.MarshalBinary()
		redisServer, redisClient := NewTestRedis()
		redisServer.HSet(transport.RedisHashName, transport.RedisHashKeyStart+strconv.FormatUint(entry.ID, 10), string(data))
		buff := make([]byte, sizeOfInitRequestMagic+sizeOfInitRequestVersion+sizeOfNonceBytes+4+len(addr)+sizeOfEncryptedToken)
		putInitRequestMagic(buff)
		putInitRequestVersion(buff)
		putInitRelayAddress(buff, addr)
		relayInitAssertions(t, buff, http.StatusNotFound, redisClient)
	})

	t.Run("valid", func(t *testing.T) {
		_, redisClient := NewTestRedis()
		addr := "127.0.0.1"
		before := uint64(time.Now().Unix())
		buff := make([]byte, sizeOfInitRequestMagic+sizeOfInitRequestVersion+sizeOfNonceBytes+4+len(addr)+sizeOfEncryptedToken)
		putInitRequestMagic(buff)
		putInitRequestVersion(buff)
		putInitRelayAddress(buff, addr)
		recorder := relayInitAssertions(t, buff, http.StatusOK, redisClient)

		header := recorder.Header()
		contentType, _ := header["Content-Type"]
		expected := transport.RelayData{
			ID:      core.GetRelayID(addr),
			Name:    addr,
			Address: addr,
		}

		resp := redisClient.HGet(transport.RedisHashName, transport.IDToRedisKey(core.GetRelayID(addr)))

		var actual transport.RelayData
		bin, _ := resp.Bytes()
		actual.UnmarshalBinary(bin)

		indx := 0
		body := recorder.Body.Bytes()

		var version uint32
		encoding.ReadUint32(body, &indx, &version)

		var timestamp uint64
		encoding.ReadUint64(body, &indx, &timestamp)

		var publicKey []byte
		encoding.ReadBytes(body, &indx, &publicKey, transport.LengthOfRelayToken)

		assert.Equal(t, "application/octet-stream", contentType[0])
		assert.Equal(t, transport.VersionNumberInitResponse, int(version))
		assert.LessOrEqual(t, before, timestamp)
		assert.GreaterOrEqual(t, uint64(time.Now().Unix()), timestamp)
		assert.Equal(t, actual.PublicKey, publicKey) // entry gets a public key assigned at init which is returned in the response

		assert.Equal(t, expected.ID, actual.ID)
		assert.Equal(t, expected.Name, actual.Name)
		assert.Equal(t, expected.Address, actual.Address)
		assert.NotZero(t, actual.LastUpdateTime)
		assert.Len(t, actual.PublicKey, 32)
	})
}
