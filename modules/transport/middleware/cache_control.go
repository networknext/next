package middleware

import (
	"net/http"
)

func CacheControl(cacheSetting string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodGet {
			w.Header().Set("Cache-Control", cacheSetting)
		}

		next.ServeHTTP(w, r)
	})
}
