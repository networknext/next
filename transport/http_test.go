package transport_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/networknext/backend/transport"
	"github.com/stretchr/testify/assert"
)

type Backend struct {
}

var backend Backend

func TestRelayInitHandler(t *testing.T) {
	writer := httptest.NewRecorder()
	request, _ := http.NewRequest("POST", "/relay_init", nil)

	handler := transport.RelayInitHandlerFunc(backend)

	handler(writer, request)

	assert.Equal(t, writer.Code, http.StatusOK)
}
