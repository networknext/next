package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/networknext/backend/transport/middleware"
	"gopkg.in/go-playground/assert.v1"
)

func TestCacheControl(t *testing.T) {
	t.Parallel()

	noopHandler := func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}

	t.Run("no-cache", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		cacheControlMiddleware := middleware.CacheControl("no-cache", http.HandlerFunc(noopHandler))
		cacheControlMiddleware.ServeHTTP(w, r)
		assert.Equal(t, "no-cache", w.Header().Get("Cache-Control"))
	})

	t.Run("max-age=100", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodGet, "/", nil)
		w := httptest.NewRecorder()

		cacheControlMiddleware := middleware.CacheControl("max-age=100", http.HandlerFunc(noopHandler))
		cacheControlMiddleware.ServeHTTP(w, r)
		assert.Equal(t, "max-age=100", w.Header().Get("Cache-Control"))
	})

	t.Run("not set on non-GET request", func(t *testing.T) {
		r := httptest.NewRequest(http.MethodPost, "/", nil)
		w := httptest.NewRecorder()

		cacheControlMiddleware := middleware.CacheControl("max-age=100", http.HandlerFunc(noopHandler))
		cacheControlMiddleware.ServeHTTP(w, r)
		assert.Equal(t, "", w.Header().Get("Cache-Control"))
	})
}
