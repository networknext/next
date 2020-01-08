package transport_test

import (
	"bytes"
	"encoding/binary"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/networknext/backend/transport"
	"github.com/stretchr/testify/assert"
)

const sizeOfUpdateRequestVersion = 4
const sizeOfRelayToken = 32
const sizeOfNumberOfRelays = 4

func putUpdateRequestVersion(buff []byte) {
	const gUpdateRequestVersion = 0
	binary.LittleEndian.PutUint32(buff, gUpdateRequestVersion)
}

func relayUpdateAssertions(t *testing.T, body []byte, expectedCode int) http.ResponseWriter {
	backend := transport.NewBackend()
	writer := httptest.NewRecorder()
	request, _ := http.NewRequest("POST", "/relay_update", bytes.NewBuffer(body))

	handler := transport.RelayInitHandlerFunc(backend)

	handler(writer, request)

	assert.Equal(t, writer.Code, expectedCode)

	return writer
}

func TestRelayUpdateHandler_IncorrectUpdateRequestVersion(t *testing.T) {
	buff := make([]byte, 0)
	relayUpdateAssertions(t, buff, http.StatusBadRequest)
}

func TestRelayUpdateHandler_MissingRelayAddress(t *testing.T) {
	buff := make([]byte, sizeOfUpdateRequestVersion)
	relayUpdateAssertions(t, buff, http.StatusBadRequest)
}

func TestRelayUpdateHandler_MissingRelayToken(t *testing.T) {
	addr := "127.0.0.1"
	buff := make([]byte, sizeOfUpdateRequestVersion+4+len(addr))
	putUpdateRequestVersion(buff)
	relayUpdateAssertions(t, buff, http.StatusBadRequest)
}

func TestRelayUpdateHandler_RelayNotFound(t *testing.T) {
	addr := "127.0.0.1"
	buff := make([]byte, sizeOfUpdateRequestVersion+4+len(addr)+sizeOfRelayToken)
	relayUpdateAssertions(t, buff, http.StatusNotFound)
}

func TestRelayUpdateHandler_NumberOfRelaysNotFound(t *testing.T) {
	addr := "127.0.0.1"
	buff := make([]byte, sizeOfUpdateRequestVersion+4+len(addr)+sizeOfRelayToken)
	relayUpdateAssertions(t, buff, http.StatusBadRequest)
}

func TestRelayUpdateHandler_NumberOfRelaysExceedsMax(t *testing.T) {
	addr := "127.0.0.1"
	buff := make([]byte, sizeOfUpdateRequestVersion+4+len(addr)+sizeOfRelayToken+sizeOfNumberOfRelays)
	relayUpdateAssertions(t, buff, http.StatusBadRequest)
}
