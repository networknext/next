package transport

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/networknext/backend/modules/transport/middleware"
)

func HealthHandlerFunc() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		_, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer r.Body.Close()
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(http.StatusText(http.StatusOK)))
	}
}

func VersionHandlerFunc(buildTime string, commitMessage string, commitHash string, allowedOrigins []string) func(w http.ResponseWriter, r *http.Request) {
	version := map[string]string{
		"build_time":     buildTime,
		"commit_message": commitMessage,
		"commit_hash":    commitHash,
	}

	return func(w http.ResponseWriter, r *http.Request) {
		middleware.CORSControlHandlerFunc(allowedOrigins, w, r)
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(version); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
