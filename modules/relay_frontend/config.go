package relay_frontend

import (
	"time"

	"github.com/networknext/backend/modules/envvar"
)

type Config struct {
	ENV                string
	MasterTimeVariance time.Duration
	MatrixStoreAddress string
	MSReadTimeout      time.Duration
	MSWriteTimeout     time.Duration
	MSMatrixTimeout    time.Duration
	ValveMatrix        bool
}

func GetConfig() (*Config, error) {
	cfg := new(Config)
	var err error

	cfg.ENV = envvar.Get("ENV", "local")

	cfg.MasterTimeVariance, err = envvar.GetDuration("MASTER_TIME_VARIANCE", 5000*time.Millisecond)
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

	cfg.ValveMatrix, err = envvar.GetBool("FEATURE_VALVE_MARTRIX", false)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}
