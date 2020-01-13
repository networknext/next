package transport_test

import (
	"bytes"
	"encoding/binary"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/go-redis/redis/v7"
	"github.com/networknext/backend/core"
	"github.com/networknext/backend/transport"
	"github.com/stretchr/testify/assert"
)

const (
	sizeOfInitRequestMagic   = 4
	sizeOfInitRequestVersion = 4
	sizeOfNonceBytes         = 24
	sizeOfEncryptedToken     = 32 + 16 // global + value of MACBYTES
)

// Returns the writer as a means to read the data that the writer contains
func relayInitAssertions(t *testing.T, body []byte, expectedCode int, redisClient *redis.Client) http.ResponseWriter {
	if redisClient == nil {
		_, redisClient = NewTestRedis()
	}

	writer := httptest.NewRecorder()
	request, _ := http.NewRequest("POST", "/relay_init", bytes.NewBuffer(body))

	handler := transport.RelayInitHandlerFunc(redisClient)

	handler(writer, request)

	assert.Equal(t, expectedCode, writer.Code)

	return writer
}

func putInitRequestMagic(buff []byte) {
	const gInitRequestMagic = uint32(0x9083708f)
	binary.LittleEndian.PutUint32(buff, gInitRequestMagic)
}

func putInitRelayAddress(buff []byte, address string) {
	offset := sizeOfInitRequestMagic + sizeOfInitRequestVersion + sizeOfNonceBytes
	binary.LittleEndian.PutUint32(buff[offset:], uint32(len(address)))
	copy(buff[offset+4:], address)
}

func putInitRequestVersion(buff []byte) {
	const gInitRequestVersion = 0
	binary.LittleEndian.PutUint32(buff[4:], gInitRequestVersion)
}

func TestRelayInitHandler(t *testing.T) {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	t.Run("missing magic number", func(t *testing.T) {
		buff := make([]byte, 0)
		relayInitAssertions(t, buff, http.StatusBadRequest, nil)
	})

	t.Run("missing request version", func(t *testing.T) {
		buff := make([]byte, sizeOfInitRequestMagic)
		putInitRequestMagic(buff)
		relayInitAssertions(t, buff, http.StatusBadRequest, nil)
	})

	t.Run("missing nonce bytes", func(t *testing.T) {
		buff := make([]byte, sizeOfInitRequestMagic+sizeOfInitRequestVersion)
		putInitRequestMagic(buff)
		putInitRequestVersion(buff)
		relayInitAssertions(t, buff, http.StatusBadRequest, nil)
	})

	t.Run("missing relay address", func(t *testing.T) {
		t.Run("byte array is not proper length", func(t *testing.T) {
			buff := make([]byte, sizeOfInitRequestMagic+sizeOfInitRequestVersion+sizeOfNonceBytes)
			putInitRequestMagic(buff)
			putInitRequestVersion(buff)
			relayInitAssertions(t, buff, http.StatusBadRequest, nil)
		})

		t.Run("byte array is proper length but there is a blank string", func(t *testing.T) {
			addr := ""
			buff := make([]byte, sizeOfInitRequestMagic+sizeOfInitRequestVersion+sizeOfNonceBytes+4+len(addr))
			putInitRequestMagic(buff)
			putInitRequestVersion(buff)
			putInitRelayAddress(buff, addr)
			relayInitAssertions(t, buff, http.StatusBadRequest, nil)
		})
	})

	t.Run("missing encryption token", func(t *testing.T) {
		addr := "127.0.0.1"
		buff := make([]byte, sizeOfInitRequestMagic+sizeOfInitRequestVersion+sizeOfNonceBytes+4+len(addr))
		putInitRequestMagic(buff)
		putInitRequestVersion(buff)
		putInitRelayAddress(buff, addr)
		relayInitAssertions(t, buff, http.StatusBadRequest, nil)
	})

	t.Run("encryption token is 0'ed", func(t *testing.T) {
		addr := "127.0.0.1"
		buff := make([]byte, sizeOfInitRequestMagic+sizeOfInitRequestVersion+sizeOfNonceBytes+4+len(addr)+sizeOfEncryptedToken)
		putInitRequestMagic(buff)
		putInitRequestVersion(buff)
		putInitRelayAddress(buff, addr)
		relayInitAssertions(t, buff, http.StatusOK, nil) // should it return ok if it is 0'ed out?
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
		buff := make([]byte, sizeOfInitRequestMagic+sizeOfInitRequestVersion+sizeOfNonceBytes+4+len(addr)+sizeOfEncryptedToken)
		putInitRequestMagic(buff)
		putInitRequestVersion(buff)
		putInitRelayAddress(buff, addr)
		writer := relayInitAssertions(t, buff, http.StatusOK, redisClient)
		header := writer.Header()
		contentType, _ := header["Content-Type"]
		assert.Equal(t, "application/octet-stream", contentType[0])
		expected := transport.RelayData{
			ID:        core.GetRelayID(addr),
			Name:      addr,
			Address:   addr,
			PublicKey: core.RandomBytes(transport.LengthOfRelayToken),
		}

		resp := redisClient.HGet(transport.RedisHashName, transport.IDToKey(core.GetRelayID(addr)))
		assert.Truef(t, resp.Err() == nil || resp.Err() == redis.Nil, "test response had error: %v", resp.Err())
		var actual transport.RelayData
		bin, _ := resp.Bytes()
		actual.UnmarshalBinary(bin)
		assert.Equal(t, expected.ID, actual.ID)
		assert.Equal(t, expected.Name, actual.Name)
		assert.Equal(t, expected.Address, actual.Address)
		assert.NotZero(t, actual.LastUpdateTime)
		assert.Len(t, actual.PublicKey, len(expected.PublicKey))
	})
}
