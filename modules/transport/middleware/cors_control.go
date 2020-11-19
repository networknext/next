package middleware

import (
	"net/http"

	"github.com/rs/cors"
)

func CORSControlHandler(allowCORS bool, allowedOrigins []string, next http.Handler) http.Handler {
	if !allowCORS {
		return cors.New(cors.Options{
			AllowedOrigins:   allowedOrigins,
			AllowCredentials: true,
			AllowedHeaders:   []string{"Authorization", "Content-Type"},
			AllowedMethods:   []string{"POST", "GET", "OPTION"},
		}).Handler(next)
	}
	return next
}

func CORSControlHandlerFunc(allowCORS bool, allowedOrigins []string, w http.ResponseWriter, r *http.Request) {
	if !allowCORS {
		cors.New(cors.Options{
			AllowedOrigins:   allowedOrigins,
			AllowCredentials: true,
			AllowedHeaders:   []string{"Authorization", "Content-Type"},
			AllowedMethods:   []string{"POST", "GET", "OPTION"},
		}).HandlerFunc(w, r)
	}
}
