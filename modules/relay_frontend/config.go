package relay_frontend

import (
	"fmt"
	"time"

	"github.com/networknext/backend/modules/envvar"
)

type RelayFrontendConfig struct {
	Env                    string
	MasterTimeVariance     time.Duration
	MatrixStoreAddress     string
	MSMaxIdleConnections   int
	MSMaxActiveConnections int
	MSReadTimeout          time.Duration
	MSWriteTimeout         time.Duration
	MSMatrixExpireTimeout  time.Duration
}

func GetConfig() (*RelayFrontendConfig, error) {
	cfg := new(RelayFrontendConfig)
	var err error

	cfg.Env = envvar.Get("ENV", "local")

	cfg.MasterTimeVariance, err = envvar.GetDuration("MASTER_TIME_VARIANCE", 5000*time.Millisecond)
	if err != nil {
		return nil, err
	}

	cfg.MatrixStoreAddress = envvar.Get("MATRIX_STORE_ADDRESS", "")
	if cfg.MatrixStoreAddress == "" {
		return nil, fmt.Errorf("MATRIX_STORE_ADDRESS not set")
	}

	maxIdleConnections, err := envvar.GetInt("MATRIX_STORE_MAX_IDLE_CONNS", 5)
	if err != nil {
		return nil, err
	}
	cfg.MSMaxIdleConnections = maxIdleConnections

	maxActiveConnections, err := envvar.GetInt("MATRIX_STORE_MAX_ACTIVE_CONNS", 5)
	if err != nil {
		return nil, err
	}
	cfg.MSMaxActiveConnections = maxActiveConnections

	cfg.MSReadTimeout, err = envvar.GetDuration("MATRIX_STORE_READ_TIMEOUT", 250*time.Millisecond)
	if err != nil {
		return nil, err
	}

	cfg.MSWriteTimeout, err = envvar.GetDuration("MATRIX_STORE_WRITE_TIMEOUT", 250*time.Millisecond)
	if err != nil {
		return nil, err
	}

	cfg.MSMatrixExpireTimeout, err = envvar.GetDuration("MATRIX_STORE_EXPIRE_TIMEOUT", 5*time.Second)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
