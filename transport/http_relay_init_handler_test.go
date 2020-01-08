package transport_test

import (
	"bytes"
	"encoding/binary"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/networknext/backend/transport"
	"github.com/stretchr/testify/assert"
)

const sizeOfInitRequestMagic = 4
const sizeOfInitRequestVersion = 4
const sizeOfNonceBytes = 24
const sizeOfEncryptedToken = 32 + 16 // global + value of MACBYTES

// Returns the writer as a means to read the data that the writer contains
func relayInitAssertions(t *testing.T, body []byte, expectedCode int, backend *transport.Backend) http.ResponseWriter {
	if backend == nil {
		backend = transport.NewBackend()
	}

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
	copy(buff[offset+4:], address)
}

func putInitRequestVersion(buff []byte) {
	const gInitRequestVersion = 0
	binary.LittleEndian.PutUint32(buff[4:], gInitRequestVersion)
}

func TestRelayInitHandler(t *testing.T) {
	log.SetOutput(ioutil.Discard)
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
			putRelayAddress(buff, addr)
			relayInitAssertions(t, buff, http.StatusBadRequest, nil)
		})
	})

	t.Run("missing encryption token", func(t *testing.T) {
		addr := "127.0.0.1"
		buff := make([]byte, sizeOfInitRequestMagic+sizeOfInitRequestVersion+sizeOfNonceBytes+4+len(addr))
		putInitRequestMagic(buff)
		putInitRequestVersion(buff)
		putRelayAddress(buff, addr)
		relayInitAssertions(t, buff, http.StatusBadRequest, nil)
	})

	t.Run("encryption token is 0'ed", func(t *testing.T) {
		addr := "127.0.0.1"
		buff := make([]byte, sizeOfInitRequestMagic+sizeOfInitRequestVersion+sizeOfNonceBytes+4+len(addr)+sizeOfEncryptedToken)
		putInitRequestMagic(buff)
		putInitRequestVersion(buff)
		putRelayAddress(buff, addr)
		relayInitAssertions(t, buff, http.StatusOK, nil) // should it return ok if it is 0'ed out?
	})

	t.Run("relay already exists", func(t *testing.T) {
		backend := transport.NewBackend()
		addr := "127.0.0.1"
		backend.RelayDatabase[addr] = transport.RelayEntry{}
		buff := make([]byte, sizeOfInitRequestMagic+sizeOfInitRequestVersion+sizeOfNonceBytes+4+len(addr)+sizeOfEncryptedToken)
		putInitRequestMagic(buff)
		putInitRequestVersion(buff)
		putRelayAddress(buff, addr)
		relayInitAssertions(t, buff, http.StatusNotFound, backend)
	})

	t.Run("valid", func(t *testing.T) {
		addr := "127.0.0.1"
		buff := make([]byte, sizeOfInitRequestMagic+sizeOfInitRequestVersion+sizeOfNonceBytes+4+len(addr)+sizeOfEncryptedToken)
		putInitRequestMagic(buff)
		putInitRequestVersion(buff)
		putRelayAddress(buff, addr)
		writer := relayInitAssertions(t, buff, http.StatusOK, nil)
		header := writer.Header()
		contentType, _ := header["Content-Type"]
		assert.Equal(t, "application/octet-stream", contentType[0])
	})
}
