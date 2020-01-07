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

func putUpdateRequestVersion(buff []byte) {
	const gUpdateRequestVersion = 0
	binary.LittleEndian.PutUint32(buff, gUpdateRequestVersion)
}

func relayUpdateAssertions(t *testing.T, body []byte, expectedCode int) http.ResponseWriter {
	writer := httptest.NewRecorder()
	request, _ := http.NewRequest("POST", "/relay_update", bytes.NewBuffer(body))

	handler := transport.RelayInitHandlerFunc(backend)

	handler(writer, request)

	assert.Equal(t, writer.Code, expectedCode)

	return writer
}
