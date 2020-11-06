package main

import (
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-kit/kit/util/conn"
	"github.com/networknext/backend/modules/common/helpers"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport"
	"net/http"
	"time"
)

type Gateway struct{
	cfg *Config
	logger log.Logger
	relayLogger log.Logger
	metrics *Metrics
	store *storage.Storer
	RelayStore storage.RelayStore
	RelayCache *storage.RelayCache
	shutdown bool
}

func (g *Gateway) Shutdown(){
	g.shutdown = true
	time.Sleep(10 *time.Second)
}

func (g *Gateway) RelayInitHandlerFunc() func(writer http.ResponseWriter, request *http.Request) {

	cfg := &transport.GatewayHandlerConfig{
		RelayStore: 		g.RelayStore,
		RelayCache:	    	*g.RelayCache,
		Storer:           	*g.store,
		InitMetrics:      	g.metrics.RelayInitMetrics,
		UpdateMetrics:	 	g.metrics.RelayUpdateMetrics,
		RouterPrivateKey:	g.cfg.RouterPrivateKey,
	}

	return transport.GatewayRelayInitHandlerFunc(g.logger, cfg)
}

func (g *Gateway) RelayUpdateHandlerFunc() func(writer http.ResponseWriter, request *http.Request) {

	cfg := &transport.GatewayHandlerConfig{
		RelayStore:       g.RelayStore,
		RelayCache:       *g.RelayCache,
		Storer:           *g.store,
		InitMetrics:      g.metrics.RelayInitMetrics,
		UpdateMetrics:    g.metrics.RelayUpdateMetrics,
		RouterPrivateKey: g.cfg.RouterPrivateKey,
	}

	return transport.GatewayRelayUpdateHandlerFunc(g.logger,g.relayLogger, cfg)
}

func (g *Gateway) RelayCacheRunner() error{

	errCount := 0
	syncTimer := helpers.NewSyncTimer(g.cfg.RelayCacheUpdate)
	for !g.shutdown{
		syncTimer.Run()

		if errCount > 10 {
			return fmt.Errorf("relay cached errored %v in a row", conn.ErrConnectionUnavailable)
		}

		relayArr, err := g.RelayStore.GetAll()
		if err != nil {
			level.Error(g.logger).Log("msg", "unable to get relays from Relay Store", "err", err.Error())
			errCount++
			continue
		}

		err = g.RelayCache.SetAll(relayArr)
		if err != nil {
			level.Error(g.logger).Log("msg", "unable to get relays from Relay Store", "err", err.Error())
			errCount++
			continue
		}

		errCount = 0
	}

	return nil
}