package transport_test

import (
	"bytes"
	"encoding/binary"
	"math"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/networknext/backend/transport"
	"github.com/stretchr/testify/assert"
)

const sizeOfUpdateRequestVersion = 4
const sizeOfRelayToken = 32
const sizeOfNumberOfRelays = 4
const sizeOfRelayPingStat = 20

func putUpdateRequestVersion(buff []byte) {
	const gUpdateRequestVersion = 0
	binary.LittleEndian.PutUint32(buff, gUpdateRequestVersion)
}

func putUpdateRelayAddress(buff []byte, address string) {
	offset := sizeOfUpdateRequestVersion
	binary.LittleEndian.PutUint32(buff[offset:], uint32(len(address)))
	copy(buff[offset+4:], address)
}

func putPingStats(buff []byte, count uint64, addressLength int) {
	offset := sizeOfUpdateRequestVersion + 4 + addressLength + sizeOfRelayToken

	binary.LittleEndian.PutUint64(buff[offset:], count)
	offset += 4

	for i := 0; i < int(count); i++ {
		id := uint64(i)
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

func relayUpdateAssertions(t *testing.T, body []byte, expectedCode int, backend *transport.Backend) http.ResponseWriter {
	if backend == nil {
		backend = transport.NewBackend()
	}

	writer := httptest.NewRecorder()
	request, _ := http.NewRequest("POST", "/relay_update", bytes.NewBuffer(body))

	handler := transport.RelayUpdateHandlerFunc(backend)

	handler(writer, request)

	assert.Equal(t, expectedCode, writer.Code)

	return writer
}

func TestRelayUpdateHandler(t *testing.T) {
	t.Run("missing request version", func(t *testing.T) {
		buff := make([]byte, 0)
		relayUpdateAssertions(t, buff, http.StatusBadRequest, nil)
	})

	t.Run("missing relay address", func(t *testing.T) {
		t.Run("byte array is not proper length", func(t *testing.T) {
			buff := make([]byte, sizeOfUpdateRequestVersion)
			putUpdateRequestVersion(buff)
			relayUpdateAssertions(t, buff, http.StatusBadRequest, nil)
		})

		t.Run("byte array is proper length but value is empty string", func(t *testing.T) {
			addr := ""
			buff := make([]byte, sizeOfUpdateRequestVersion+4+len(addr))
			putUpdateRequestVersion(buff)
			putUpdateRelayAddress(buff, addr)
			relayUpdateAssertions(t, buff, http.StatusBadRequest, nil)
		})
	})

	t.Run("missing relay token", func(t *testing.T) {
		addr := "127.0.0.1"
		buff := make([]byte, sizeOfUpdateRequestVersion+4+len(addr))
		putUpdateRequestVersion(buff)
		putUpdateRelayAddress(buff, addr)
		relayUpdateAssertions(t, buff, http.StatusBadRequest, nil)
	})

	t.Run("missing number of relays", func(t *testing.T) {
		addr := "127.0.0.1"
		buff := make([]byte, sizeOfUpdateRequestVersion+4+len(addr)+sizeOfRelayToken)
		putUpdateRequestVersion(buff)
		putUpdateRelayAddress(buff, addr)
		relayUpdateAssertions(t, buff, http.StatusBadRequest, nil)
	})

	t.Run("number of relays exceeds max", func(t *testing.T) {
		numRelays := 1025
		addr := "127.0.0.1"
		buff := make([]byte, sizeOfUpdateRequestVersion+4+len(addr)+sizeOfRelayToken+sizeOfNumberOfRelays+numRelays*sizeOfRelayPingStat)
		putUpdateRequestVersion(buff)
		putUpdateRelayAddress(buff, addr)
		putPingStats(buff, uint64(numRelays), len(addr))
		relayUpdateAssertions(t, buff, http.StatusBadRequest, nil)
	})

	t.Run("relay not found", func(t *testing.T) {
		numRelays := 3
		addr := "127.0.0.1"
		buff := make([]byte, sizeOfUpdateRequestVersion+4+len(addr)+sizeOfRelayToken+sizeOfNumberOfRelays+numRelays*sizeOfRelayPingStat)
		putUpdateRequestVersion(buff)
		putUpdateRelayAddress(buff, addr)
		putPingStats(buff, uint64(numRelays), len(addr))
		relayUpdateAssertions(t, buff, http.StatusNotFound, nil)
	})
}
