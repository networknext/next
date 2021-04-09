package transport

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/networknext/backend/modules/metrics"
)

// GatewayRelayUpdateHandlerFunc receives relay update requests and puts them in requestChan
func GatewayRelayUpdateHandlerFunc(logger log.Logger, gatewayMetrics *metrics.RelayGatewayMetrics, requestChan chan []byte) func(writer http.ResponseWriter, request *http.Request) {
	handlerLogger := log.With(logger, "handler", "update")

	return func(writer http.ResponseWriter, request *http.Request) {
		durationStart := time.Now()
		defer func() {
			durationSince := time.Since(durationStart)
			gatewayMetrics.HandlerMetrics.Duration.Set(float64(durationSince.Milliseconds()))
			gatewayMetrics.HandlerMetrics.Invocations.Add(1)
			if durationSince.Milliseconds() > 100 {
				gatewayMetrics.HandlerMetrics.LongDuration.Add(1)
			}
		}()

		// Read the body from the request
		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			gatewayMetrics.ErrorMetrics.ReadPacketFailure.Add(1)
			level.Error(logger).Log("msg", "could not read packet", "err", err)
			writer.WriteHeader(http.StatusInternalServerError) // 500
			return
		}
		defer request.Body.Close()

		// Ensure the content type is valid
		if request.Header.Get("Content-Type") != "application/octet-stream" {
			gatewayMetrics.ErrorMetrics.ContentTypeFailure.Add(1)
			level.Error(logger).Log("err", fmt.Sprintf("%s - error: relay update unsupported content type", request.RemoteAddr))
			writer.WriteHeader(http.StatusBadRequest) // 400
			return
		}

		// Insert the body into the channel
		requestChan <- body

		level.Debug(logger).Log("msg", fmt.Sprintf("inserted relay update %s body into channel", request.RemoteAddr))
	}
}
