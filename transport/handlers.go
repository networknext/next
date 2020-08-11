package transport

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/networknext/backend/transport/middleware"
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

func VersionHandlerFunc(buildtime string, sha string, tag string, commitMessage string, allowCORS bool) func(w http.ResponseWriter, r *http.Request) {
	version := map[string]string{
		"build_timestamp": buildtime,
		"commit_message":  commitMessage,
		"sha":             sha,
		"tag":             tag,
	}

	return func(w http.ResponseWriter, r *http.Request) {
		middleware.CORSControlHandlerFunc(allowCORS, w, r)
		w.WriteHeader(http.StatusOK)
		if err := json.NewEncoder(w).Encode(version); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
