package optimizer

import (
	"fmt"

	"github.com/networknext/backend/modules/envvar"
	"time"
)

type Config struct{
	MaxJitter float32
	MaxPacketLoss float32
	MatrixBufferSize  int
	RelayCacheUpdate time.Duration
	RelayStoreAddress string
	RelayStoreReadTimeout time.Duration
	RelayStoreWriteTimeout time.Duration
	RelayStoreRelayTimeout time.Duration
	subscriberPort 		string
	subscriberRecieveBufferSize int
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

	relayStoreReadTimeout, err := envvar.GetDuration( "RELAY_STORE_READ_TIMEOUT", 250 * time.Millisecond)
	if err != nil {
		return nil, err
	}

	relayStoreWriteTimeout, err := envvar.GetDuration( "RELAY_STORE_WRITE_TIMEOUT", 250 * time.Millisecond)
	if err != nil {
		return nil, err
	}

	relayStoreRelayTimeout, err := envvar.GetDuration( "RELAY_STORE_RELAY_TIMEOUT", 5 * time.Second)
	if err != nil {
		return nil, err
	}

	subscriberPort := envvar.Get("SUBSCRIBER_PORT", "5555")

	subscriberRecieveBufferSize, err := envvar.GetInt("SUBSCRIBER_RECIEVE_BUFFER_SIZE",100000)
	if err != nil {
		return nil, err
	}

	return &Config{
		MaxJitter: float32(maxJitter),
		MaxPacketLoss: float32(maxPacketLoss),
		MatrixBufferSize: matrixBufferSize,
		RelayCacheUpdate: relayCacheUpdate,
		RelayStoreAddress: relayStoreAddress,
		RelayStoreReadTimeout: relayStoreReadTimeout,
		RelayStoreWriteTimeout: relayStoreWriteTimeout,
		RelayStoreRelayTimeout: relayStoreRelayTimeout,
		subscriberPort: subscriberPort,
		subscriberRecieveBufferSize: subscriberRecieveBufferSize,
	}, nil
}