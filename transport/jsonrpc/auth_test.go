package jsonrpc_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/networknext/backend/transport/jsonrpc"
	"github.com/stretchr/testify/assert"
)

func TestAuthMiddleware(t *testing.T) {
	t.Run("skip auth", func(t *testing.T) {
		noopHandler := func(w http.ResponseWriter, _ *http.Request) {
			w.WriteHeader(http.StatusOK)
		}

		authMiddleware := jsonrpc.AuthMiddleware("", http.HandlerFunc(noopHandler))

		req := httptest.NewRequest(http.MethodGet, "/", nil)
		res := httptest.NewRecorder()

		authMiddleware.ServeHTTP(res, req)
		assert.Equal(t, res.Code, http.StatusOK)
	})
}
