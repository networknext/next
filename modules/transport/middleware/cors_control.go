package middleware

import (
	"net/http"

	"github.com/rs/cors"
)

func CORSControlHandler(allowedOrigins []string, next http.Handler) http.Handler {
	return cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowCredentials: true,
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowedMethods:   []string{"POST", "GET", "OPTION"},
	}).Handler(next)
}

func CORSControlHandlerFunc(allowedOrigins []string, w http.ResponseWriter, r *http.Request) {
	cors.New(cors.Options{
		AllowedOrigins:   allowedOrigins,
		AllowCredentials: true,
		AllowedHeaders:   []string{"Authorization", "Content-Type"},
		AllowedMethods:   []string{"POST", "GET", "OPTION"},
	}).HandlerFunc(w, r)
}
