package relay_gateway

import (
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-kit/kit/util/conn"
	"github.com/networknext/backend/modules/common/helpers"
	"github.com/networknext/backend/storage"
	"github.com/networknext/backend/transport"
	"github.com/networknext/backend/transport/pubsub"
	"net/http"
	"time"
)

type Gateway struct{
	Cfg *Config
	Logger log.Logger
	RelayLogger log.Logger
	Metrics *Metrics
	Publishers []pubsub.Publisher
	Store *storage.Storer
	RelayStore storage.RelayStore
	RelayCache *storage.RelayCache
	ShutdownSvc bool
}


func (g *Gateway) Shutdown(){
	g.ShutdownSvc = true
	time.Sleep(10 *time.Second)
}

func (g *Gateway) RelayInitHandlerFunc() func(writer http.ResponseWriter, request *http.Request) {

	fmt.Println("init request recieved")
	Cfg := &transport.GatewayHandlerConfig{
		RelayStore: 		g.RelayStore,
		RelayCache:	    	*g.RelayCache,
		Storer:           	*g.Store,
		InitMetrics:      	g.Metrics.RelayInitMetrics,
		UpdateMetrics:	 	g.Metrics.RelayUpdateMetrics,
		RouterPrivateKey:	g.Cfg.RouterPrivateKey,
	}

	return transport.GatewayRelayInitHandlerFunc(g.Logger, Cfg)
}

func (g *Gateway) RelayUpdateHandlerFunc() func(writer http.ResponseWriter, request *http.Request) {
	fmt.Println("update request recieved")
	Cfg := &transport.GatewayHandlerConfig{
		RelayStore:       	g.RelayStore,
		RelayCache:       	*g.RelayCache,
		Storer:           	*g.Store,
		UpdateMetrics:    	g.Metrics.RelayUpdateMetrics,
		RouterPrivateKey:	g.Cfg.RouterPrivateKey,
		Publishers: 		g.Publishers,
	}

	return transport.GatewayRelayUpdateHandlerFunc(g.Logger,g.RelayLogger, Cfg)
}

func (g *Gateway) RelayCacheRunner() error{

	errCount := 0
	syncTimer := helpers.NewSyncTimer(g.Cfg.RelayCacheUpdate)
	for !g.ShutdownSvc{
		syncTimer.Run()

		if errCount > 10 {
			return fmt.Errorf("relay cached errored %v in a row", conn.ErrConnectionUnavailable)
		}

		relayArr, err := g.RelayStore.GetAll()
		if err != nil {
			level.Error(g.Logger).Log("msg", "unable to get relays from Relay Store", "err", err.Error())
			errCount++
			continue
		}

		err = g.RelayCache.SetAll(relayArr)
		if err != nil {
			level.Error(g.Logger).Log("msg", "unable to get relays from Relay Store", "err", err.Error())
			errCount++
			continue
		}

		errCount = 0
	}

	return nil
}