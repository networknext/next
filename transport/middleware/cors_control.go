package middleware

import (
	"net/http"
	"os"
	"strings"

	"github.com/rs/cors"
)

func CORSControlHandler(allowCORS bool, next http.Handler) http.Handler {
	if !allowCORS {
		allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
		return cors.New(cors.Options{
			AllowedOrigins:   strings.Split(allowedOrigins, ","),
			AllowCredentials: true,
			AllowedHeaders:   []string{"Authorization", "Content-Type"},
			AllowedMethods:   []string{"POST", "GET", "OPTION"},
		}).Handler(next)
	}
	return next
}

func CORSControlHandlerFunc(allowCORS bool, w http.ResponseWriter, r *http.Request) {
	if !allowCORS {
		allowedOrigins := os.Getenv("ALLOWED_ORIGINS")
		cors.New(cors.Options{
			AllowedOrigins:   strings.Split(allowedOrigins, ","),
			AllowCredentials: true,
			AllowedHeaders:   []string{"Authorization", "Content-Type"},
			AllowedMethods:   []string{"POST", "GET", "OPTION"},
		}).HandlerFunc(w, r)
	}
}
