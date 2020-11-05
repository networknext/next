package main

import (
	"fmt"

	"github.com/networknext/backend/modules/envvar"
	"time"
)

type Config struct{
	maxJitter float32
	maxPacketLoss float32
	matrixBufferSize  int
	relayCacheUpdate time.Duration
	relayStoreAddress string
	RelayStoreReadTimeout int
	RelayStoreWriteTimeout int
	RelayStoreRelayTimeout int
}

func GetConfig() (*Config, error){
	var err error
	// Get the max jitter and max packet loss env vars
	if !envvar.Exists("RELAY_ROUTER_MAX_JITTER") {
		return nil, fmt.Errorf("RELAY_ROUTER_MAX_JITTER not set")
	}

	maxJitter, err := envvar.GetFloat("RELAY_ROUTER_MAX_JITTER", 0)
	if err != nil {
		return nil, err
	}

	if !envvar.Exists("RELAY_ROUTER_MAX_PACKET_LOSS") {
		return nil, fmt.Errorf("RELAY_ROUTER_MAX_PACKET_LOSS not set")
	}

	maxPacketLoss, err := envvar.GetFloat("RELAY_ROUTER_MAX_PACKET_LOSS", 0)
	if err != nil {
		return nil, err
	}

	matrixBufferSize, err := envvar.GetInt("MATRIX_BUFFER_SIZE", 100000)
	if err != nil {
		return nil, err
	}

	relayCacheUpdate, err := envvar.GetDuration("RELAY_CACHE_UPDATE", 1 *time.Second)
	if err != nil {
		return nil, err
	}

	if !envvar.Exists("RELAY_STORE_ADDRESS") {
		return nil, fmt.Errorf("RELAY_STORE_ADDRESS not set")
	}
	relayStoreAddress := envvar.Get("RELAY_STORE_ADDRESS", "")


	return &Config{
		maxJitter: float32(maxJitter),
		maxPacketLoss: float32(maxPacketLoss),
		matrixBufferSize: matrixBufferSize,
		relayCacheUpdate: relayCacheUpdate,
		relayStoreAddress: relayStoreAddress,
	}, nil
}