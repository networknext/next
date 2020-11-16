package relay_gateway

import (
	"fmt"
	"github.com/networknext/backend/modules/envvar"
	"time"
)

type Config struct{
	PublisherSendBuffer int
	PublishToHosts []string
	RouterPrivateKey []byte
	RelayCacheUpdate time.Duration
	RelayStoreAddress string
	RelayStoreReadTimeout time.Duration
	RelayStoreWriteTimeout time.Duration
	RelayStoreRelayTimeout time.Duration
}

func NewConfig() (*Config, error){
	cfg := new(Config)

	routerPrivateKey, err := envvar.GetBase64("RELAY_ROUTER_PRIVATE_KEY", nil)
	if err != nil {
		return nil, fmt.Errorf("RELAY_ROUTER_PRIVATE_KEY not set")
	}
	cfg.RouterPrivateKey = routerPrivateKey

	relayCacheUpdate, err := envvar.GetDuration("RELAY_CACHE_UPDATE", 1 *time.Second)
	if err != nil {
		return nil, err
	}
	cfg.RelayCacheUpdate = relayCacheUpdate

	relayStoreAddress := envvar.Get("RELAY_STORE_ADDRESS", "127.0.0.1:6379")
	cfg.RelayStoreAddress = relayStoreAddress


	relayStoreReadTimeout, err := envvar.GetDuration( "RELAY_STORE_READ_TIMEOUT", 500 * time.Millisecond)
	if err != nil {
		return nil, err
	}
	cfg.RelayStoreReadTimeout = relayStoreReadTimeout

	relayStoreWriteTimeout, err := envvar.GetDuration( "RELAY_STORE_WRITE_TIMEOUT", 500 * time.Millisecond)
	if err != nil {
		return nil, err
	}
	cfg.RelayStoreWriteTimeout = relayStoreWriteTimeout

	relayStoreRelayTimeout, err := envvar.GetDuration( "RELAY_STORE_RELAY_TIMEOUT", 5 * time.Second)
	if err != nil {
		return nil, err
	}
	cfg.RelayStoreRelayTimeout = relayStoreRelayTimeout

	publishToHosts := envvar.GetList("PUBLISH_TO_HOSTS", []string{"tcp://127.0.0.1:5555"})
	cfg.PublishToHosts = publishToHosts

	publisherSendBuffer, err := envvar.GetInt( "PUBLISHER_SEND_BUFFER", 100000)
	if err != nil {
		return nil, err
	}
	cfg.PublisherSendBuffer = publisherSendBuffer

	return cfg, nil
}