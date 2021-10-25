package transport

import (
	"fmt"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"

	"github.com/networknext/backend/modules/core"
	"github.com/networknext/backend/modules/metrics"
)

type RelayForwarderParams struct {
	GatewayAddr string
	Metrics     *metrics.RelayForwarderMetrics
}

/*
	ForwardPostHandlerFunc() utilizes HTTP reverse proxy to allow Riot relays to communicate
	with our Relay Backend by forwarding the request and returning the corresponding response
	to the Relay Gateway.
*/
func ForwardPostHandlerFunc(params *RelayForwarderParams) func(w http.ResponseWriter, r *http.Request) {

	return func(w http.ResponseWriter, r *http.Request) {
		// Record duration time and invocations
		durationStart := time.Now()
		defer func() {
			milliseconds := float64(time.Since(durationStart).Milliseconds())
			params.Metrics.HandlerMetrics.Duration.Set(float64(milliseconds))
			if milliseconds > 100 {
				params.Metrics.HandlerMetrics.LongDuration.Add(1)
			}
			params.Metrics.HandlerMetrics.Invocations.Add(1)
		}()

		// Parse the remote address to get the origin URL
		origin, err := url.Parse(fmt.Sprintf("//%s", r.RemoteAddr))
		if err != nil {
			core.Error("error parsing request remote addr as URL (%s): %v", r.RemoteAddr, err)
			w.WriteHeader(http.StatusInternalServerError)
			params.Metrics.ErrorMetrics.ParseURLError.Add(1)
			return
		}

		// Get the requested path (i.e. /relay_update)
		requestedPath := r.RequestURI

		// Create a reverse proxy
		reverseProxy := httputil.NewSingleHostReverseProxy(origin)

		// Modify the director to forward the request to the relay gateway
		reverseProxy.Director = func(req *http.Request) {
			req.Header.Add("X-Forwarded-Host", req.Host)
			req.Header.Add("X-Origin-Host", origin.Host)
			req.URL.Scheme = "http"
			req.URL.Host = params.GatewayAddr
			req.URL.Path = requestedPath
		}

		// Add an error handler to use our logger
		reverseProxy.ErrorHandler = func(writer http.ResponseWriter, req *http.Request, err error) {
			if err != nil {
				core.Error("error reaching relay gateway: %v", err)
				writer.WriteHeader(http.StatusInternalServerError)
				writer.Write([]byte(err.Error()))
				params.Metrics.ErrorMetrics.ForwardPostError.Add(1)
			}
		}

		// Serve the request
		reverseProxy.ServeHTTP(w, r)
	}
}
