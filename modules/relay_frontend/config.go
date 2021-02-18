package relay_frontend

import (
	"fmt"
	"time"

	"github.com/networknext/backend/modules/envvar"
)

type Config struct {
	ENV                   string
	MatrixSvcTimeVariance time.Duration
	OptimizerTimeVariance time.Duration
	MatrixStoreAddress    string
	MSReadTimeout         time.Duration
	MSWriteTimeout        time.Duration
	MSMatrixTimeout       time.Duration
	RB15Enabled           bool
	RB20Enabled           bool
	RelayBackendAddresses []string
}

func GetConfig() (*Config, error) {
	cfg := new(Config)
	var err error

	cfg.ENV = envvar.Get("ENV", "local")

	cfg.MatrixSvcTimeVariance, err = envvar.GetDuration("MATRIX_SVC_TIME_VARIANCE", 2000*time.Millisecond)
	if err != nil {
		return nil, err
	}

	cfg.OptimizerTimeVariance, err = envvar.GetDuration("OPTIMIZER_TIME_VARIANCE", 5000*time.Millisecond)
	if err != nil {
		return nil, err
	}

	cfg.MatrixStoreAddress = envvar.Get("MATRIX_STORE_ADDRESS", "127.0.0.1:6379")

	cfg.MSReadTimeout, err = envvar.GetDuration("MATRIX_STORE_READ_TIMEOUT", 250*time.Millisecond)
	if err != nil {
		return nil, err
	}

	cfg.MSWriteTimeout, err = envvar.GetDuration("MATRIX_STORE_WRITE_TIMEOUT", 250*time.Millisecond)
	if err != nil {
		return nil, err
	}

	cfg.MSMatrixTimeout, err = envvar.GetDuration("MATRIX_STORE_MATRIX_TIMEOUT", 5*time.Second)
	if err != nil {
		return nil, err
	}

	cfg.RB15Enabled, err = envvar.GetBool("FEATURE_RB15_ENABLED", false)
	if err != nil {
		return nil, err
	}

	cfg.RB20Enabled, err = envvar.GetBool("FEATURE_RB20_ENABLED", false)
	if err != nil {
		return nil, err
	}

	if cfg.RB15Enabled == cfg.RB20Enabled {
		return nil, fmt.Errorf("rb15 and rb20 cannot be the same")
	}

	cfg.RelayBackendAddresses = envvar.GetList("FEATURE_RB15_RELAY_ADDRESSES", []string{})
	if cfg.ENV == "local" {
		cfg.RelayBackendAddresses = []string{"127.0.0.1:30002"}
	}

	return cfg, nil
}
