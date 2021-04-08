package transport

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/networknext/backend/modules/metrics"
	"github.com/networknext/backend/modules/routing"
	"github.com/networknext/backend/modules/storage"
	"github.com/networknext/backend/modules/transport/pubsub"
)

type GatewayConfig struct {
	UseHTTP               bool
	RelayBackendAddresses []string
	HTTPTimeout           time.Duration

	PublishToHosts        []string
	PublisherSendBuffer   int
	PublisherRefreshTimer time.Duration
}

// GatewayRelayUpdateHandlerFunc receives relay update requests and puts them in requestChan
func GatewayRelayUpdateHandlerFunc(logger log.Logger, handlerMetrics *metrics.RelayUpdateMetrics, requestChan chan []byte) func(writer http.ResponseWriter, request *http.Request) {
	handlerLogger := log.With(logger, "handler", "update")

	return func(writer http.ResponseWriter, request *http.Request) {
		durationStart := time.Now()
		defer func() {
			durationSince := time.Since(durationStart)
			handlerMetrics.UpdateMetrics.DurationGauge.Set(float64(durationSince.Milliseconds()))
			handlerMetrics.UpdateMetrics.Invocations.Add(1)
		}()

		body, err := ioutil.ReadAll(request.Body)
		if err != nil {
			level.Error(logger).Log("msg", "could not read packet", "err", err)
			writer.WriteHeader(http.StatusInternalServerError) // 500
			return
		}
		defer request.Body.Close()

		if request.Header.Get("Content-Type") != "application/octet-stream" {
			level.Error(logger).Log("err", fmt.Sprintf("%s - error: relay update unsupported content type", request.RemoteAddr))
			writer.WriteHeader(http.StatusBadRequest) // 400
			return
		}

		// Insert the body into the channel
		requestChan <- body

		level.Debug(logger).Log("msg", fmt.Sprintf("inserted relay update %s body into channel", request.RemoteAddr))
	}
}
