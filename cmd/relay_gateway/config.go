package main

import (
	"fmt"
	"github.com/networknext/backend/modules/envvar"
	"time"
)

type Config struct{
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
	if err != nil {
		return nil, err
	}
	cfg.RelayStoreAddress = relayStoreAddress


	relayStoreReadTimeout, err := envvar.GetDuration( "RELAY_STORE_READ_TIMEOUT", 250 * time.Millisecond)
	if err != nil {
		return nil, err
	}
	cfg.RelayStoreReadTimeout = relayStoreReadTimeout

	relayStoreWriteTimeout, err := envvar.GetDuration( "RELAY_STORE_WRITE_TIMEOUT", 250 * time.Millisecond)
	if err != nil {
		return nil, err
	}
	cfg.RelayStoreWriteTimeout = relayStoreWriteTimeout

	relayStoreRelayTimeout, err := envvar.GetDuration( "RELAY_STORE_RELAY_TIMEOUT", 5 * time.Second)
	if err != nil {
		return nil, err
	}
	cfg.RelayStoreRelayTimeout = relayStoreRelayTimeout

	return cfg, nil
}