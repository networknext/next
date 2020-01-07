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

const sizeOfInitRequestMagic = 4
const sizeOfInitRequestVersion = 4
const sizeOfNonceBytes = 24
const sizeOfRelayAddressLength = 256
const sizeOfEncryptedToken = 32 + 16 // global + value of MACBYTES

// Returns the writer as a means to read the data that the writer contains
func relayInitAssertions(t *testing.T, body []byte, expectedCode int) http.ResponseWriter {
	backend := transport.NewBackend()
	writer := httptest.NewRecorder()
	request, _ := http.NewRequest("POST", "/relay_init", bytes.NewBuffer(body))

	handler := transport.RelayInitHandlerFunc(backend)

	handler(writer, request)

	assert.Equal(t, expectedCode, writer.Code)

	return writer
}

func putInitRequestMagic(buff []byte) {
	const gInitRequestMagic = uint32(0x9083708f)
	binary.LittleEndian.PutUint32(buff, gInitRequestMagic)
}

func putRelayAddress(buff []byte, address string) {
	offset := sizeOfInitRequestMagic + sizeOfInitRequestVersion + sizeOfNonceBytes
	binary.LittleEndian.PutUint32(buff[offset:], uint32(len(address)))
	offset += 4
	copy(buff[offset:], address)
}

func putInitRequestVersion(buff []byte) {
	const gInitRequestVersion = 0
	binary.LittleEndian.PutUint32(buff[4:], gInitRequestVersion)
}

func TestRelayInitHandler_MissingMagicNumber(t *testing.T) {
	buff := make([]byte, 0)
	relayInitAssertions(t, buff, http.StatusBadRequest)
}

func TestRelayInitHandler_MissingRequestVersion(t *testing.T) {
	buff := make([]byte, sizeOfInitRequestMagic)
	putInitRequestMagic(buff)
	relayInitAssertions(t, buff, http.StatusBadRequest)
}

func TestRelayInitHandler_MissingNonceBytes(t *testing.T) {
	buff := make([]byte, sizeOfInitRequestMagic+sizeOfInitRequestVersion)
	putInitRequestMagic(buff)
	putInitRequestVersion(buff)
	relayInitAssertions(t, buff, http.StatusBadRequest)
}

func TestRelayInitHandler_MissingRelayAddress1(t *testing.T) {
	buff := make([]byte, sizeOfInitRequestMagic+sizeOfInitRequestVersion+sizeOfNonceBytes)
	putInitRequestMagic(buff)
	putInitRequestVersion(buff)
	// ? can nonce bytes be 0'ed
	relayInitAssertions(t, buff, http.StatusBadRequest)
}

func TestRelayInitHandler_MissingRelayAddress2(t *testing.T) {
	buff := make([]byte, sizeOfInitRequestMagic+sizeOfInitRequestVersion+sizeOfNonceBytes+sizeOfRelayAddressLength)
	putInitRequestMagic(buff)
	putInitRequestVersion(buff)
	putRelayAddress(buff, "")
	// ? can nonce bytes be 0'ed
	relayInitAssertions(t, buff, http.StatusBadRequest)
}

func TestRelayInitHandler_MissingEncryptedToken(t *testing.T) {
	buff := make([]byte, sizeOfInitRequestMagic+sizeOfInitRequestVersion+sizeOfNonceBytes+sizeOfRelayAddressLength)
	putInitRequestMagic(buff)
	putInitRequestVersion(buff)
	putRelayAddress(buff, "127.0.0.1")
	// ? can relay address also be 0'ed
	relayInitAssertions(t, buff, http.StatusBadRequest)
}

func TestRelayInitHandler_CryptoCheckFails(t *testing.T) {
	buff := make([]byte, sizeOfInitRequestMagic+sizeOfInitRequestVersion+sizeOfNonceBytes+sizeOfRelayAddressLength+sizeOfEncryptedToken)
	putInitRequestMagic(buff)
	putInitRequestVersion(buff)
	putRelayAddress(buff, "127.0.0.1")
	// ? if encrypted token is 0'ed will that cause a fail
	relayInitAssertions(t, buff, http.StatusBadRequest)
}

func TestRelayInitHandler_RelayAlreadyExists(t *testing.T) {
	buff := make([]byte, sizeOfInitRequestMagic+sizeOfInitRequestVersion+sizeOfNonceBytes+sizeOfRelayAddressLength+sizeOfEncryptedToken)
	putInitRequestMagic(buff)
	putInitRequestVersion(buff)
	putRelayAddress(buff, "127.0.0.1")
	// put address into backend.relayDatabase here
	relayInitAssertions(t, buff, http.StatusNotFound)
}

func TestRelayInitHandler_Valid(t *testing.T) {
	buff := make([]byte, sizeOfInitRequestMagic+sizeOfInitRequestVersion+sizeOfNonceBytes+sizeOfRelayAddressLength+sizeOfEncryptedToken)
	putInitRequestMagic(buff)
	putInitRequestVersion(buff)
	putRelayAddress(buff, "127.0.0.1")
	// stub stuff out here
	writer := relayInitAssertions(t, buff, http.StatusOK)
	assert.Equal(t, writer.Header()["Content-Type"], "application/octet-stream")
	// TODO assert writer data, unsure how to access that, found MultiWriter but unsure if that's the right thing to use
}
