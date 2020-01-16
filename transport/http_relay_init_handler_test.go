package transport_test

import (
	"bytes"
	"encoding/binary"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/networknext/backend/core"
	"github.com/networknext/backend/crypto"
	"github.com/networknext/backend/routing"
	"github.com/networknext/backend/transport"
	"github.com/stretchr/testify/assert"
)

const (
	sizeOfInitRequestMagic   = 4
	sizeOfInitRequestVersion = 4
)

// Returns the writer as a means to read the data that the writer contains
func relayInitAssertions(t *testing.T, body []byte, expectedCode int, relaydb *core.RelayDatabase) http.ResponseWriter {
	if relaydb == nil {
		relaydb = core.NewRelayDatabase()
	}

	writer := httptest.NewRecorder()
	request, _ := http.NewRequest("POST", "/relay_init", bytes.NewBuffer(body))

	handler := transport.RelayInitHandlerFunc(relaydb)

	handler(writer, request)

	assert.Equal(t, expectedCode, writer.Code)

	return writer
}

func writeRelayAddress(buff []byte, address string) {
	offset := sizeOfInitRequestMagic + sizeOfInitRequestVersion + crypto.NonceSize
	binary.LittleEndian.PutUint32(buff[offset:], uint32(len(address)))
	copy(buff[offset+4:], address)
}

func TestRelayInitHandler(t *testing.T) {
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
		t.Skip("missing dependancy on config store to pull relay's public key to pass decryption")

		relaydb := core.NewRelayDatabase()
		addr := "127.0.0.1"
		relaydb.Relays[core.GetRelayID(addr)] = core.RelayData{}
		buff := make([]byte, sizeOfInitRequestMagic+sizeOfInitRequestVersion+crypto.NonceSize+4+len(addr)+routing.TokenSize)
		binary.LittleEndian.PutUint32(buff, transport.InitRequestMagic)
		binary.LittleEndian.PutUint32(buff[4:], 0)
		writeRelayAddress(buff, addr)
		relayInitAssertions(t, buff, http.StatusNotFound, relaydb)
	})

	t.Run("valid", func(t *testing.T) {
		t.Skip("missing dependancy on config store to pull relay's public key to pass decryption")

		addr := "127.0.0.1"
		buff := make([]byte, sizeOfInitRequestMagic+sizeOfInitRequestVersion+crypto.NonceSize+4+len(addr)+routing.TokenSize)
		binary.LittleEndian.PutUint32(buff, transport.InitRequestMagic)
		binary.LittleEndian.PutUint32(buff[4:], 0)
		writeRelayAddress(buff, addr)

		writer := relayInitAssertions(t, buff, http.StatusOK, nil)
		contentType := writer.Header().Get("Content-Type")
		assert.Equal(t, "application/octet-stream", contentType)
	})
}
