package relay_forwarder

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/gorilla/mux"
	"github.com/networknext/backend/modules/envvar"
)

func main() {

	lbAddr := envvar.Get("GATEWAY_LB", "")
	_, err := url.Parse(lbAddr)
	if err != nil {
		log.Fatal(err)
	}

	healthURI := fmt.Sprintf("http://%s/health", lbAddr)
	versionURI := fmt.Sprintf("http://%s/version", lbAddr)
	initURI := fmt.Sprintf("http://%s/relay_init", lbAddr)
	updateURI := fmt.Sprintf("http://%s/relay_update", lbAddr)
	costURI := fmt.Sprintf("http://%s/cost_matrix", lbAddr)
	matrixURI := fmt.Sprintf("http://%s/route_matrix", lbAddr)
	valveURI := fmt.Sprintf("http://%s/route_matrix_valve", lbAddr)
	//debugURI := fmt.Sprintf("http://%s/debug/vars", lbAddr)
	//dashURI := fmt.Sprintf("http://%s/relay_dashboard", lbAddr)
	//statsURI := fmt.Sprintf("http://%s/relay_stats", lbAddr)

	fmt.Printf("starting http server\n")

	router := mux.NewRouter()
	router.HandleFunc("/health", forwardGet(healthURI, false))
	router.HandleFunc("/version", forwardGet(versionURI, false))
	router.HandleFunc("/relay_init", forwardPost(initURI)).Methods("POST")
	router.HandleFunc("/relay_update", forwardPost(updateURI)).Methods("POST")
	router.HandleFunc("/cost_matrix", forwardGet(costURI, true)).Methods("GET")
	router.HandleFunc("/route_matrix", forwardGet(matrixURI, true)).Methods("GET")
	router.HandleFunc("/route_matrix_valve", forwardGet(valveURI, true)).Methods("GET")
	//router.Handle("/debug/vars", expvar.Handler())
	//router.HandleFunc("/relay_dashboard", forwardGet(dashURI))
	//router.HandleFunc("/relay_stats", forwardGet(statsURI))

	go func() {
		port := envvar.Get("PORT", "30000")

		err := http.ListenAndServe(":"+port, router)
		if err != nil {
			log.Fatal(err)
			os.Exit(1) // todo: don't os.Exit() here, but find a way to exit
		}
	}()
}

func forwardGet(address string, octet bool) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		resp, err := http.Get(address)
		if err != nil {
			log.Printf("error forwarding get: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Printf("error reading response body: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()
		w.WriteHeader(resp.StatusCode)
		if octet {
			w.Header().Set("Content-Type", "application/octet-stream")
		}
		w.Write(body)
	}
}

func forwardPost(address string) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {

		fmt.Println("post started")
		resp, err := http.Post(address, r.Header.Get("Content-Type"), r.Body)
		if err != nil {
			fmt.Printf("error forwarding get: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			fmt.Printf("error reading response body: %s", err.Error())
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		defer resp.Body.Close()
		w.WriteHeader(resp.StatusCode)
		w.Header().Set("Content-Type", r.Header.Get("Content-Type"))
		w.Write(body)

		fmt.Println("post finished")
	}
}
