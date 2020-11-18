package route_matrix_selector

import (
	"github.com/networknext/backend/modules/envvar"
	"time"
)

type Config struct{
	MatrixSvcTimeVariance time.Duration
	OptimizerTimeVariance time.Duration
	MatrixStoreAddress string
	MSReadTimeout time.Duration
	MSWriteTimeout time.Duration
	MSMatrixTimeout time.Duration

}

func GetConfig() (*Config, error){
	cfg := new(Config)
	var err error

	cfg.MatrixSvcTimeVariance, err = envvar.GetDuration("MATRIX_SVC_TIME_VARIANCE", 2000 *time.Millisecond)
	if err != nil {
		return nil, err
	}

	cfg.OptimizerTimeVariance, err = envvar.GetDuration( "OPTIMIZER_TIME_VARIANCE",5000 * time.Millisecond)
	if err != nil {
		return nil, err
	}

	cfg.MatrixStoreAddress = envvar.Get("MATRIX_STORE_ADDRESS", "127.0.0.1:6379")

	cfg.MSReadTimeout, err = envvar.GetDuration( "MATRIX_STORE_READ_TIMEOUT", 250 * time.Millisecond)
	if err != nil {
		return nil, err
	}

	cfg.MSWriteTimeout, err = envvar.GetDuration( "MATRIX_STORE_WRITE_TIMEOUT", 250 * time.Millisecond)
	if err != nil {
		return nil, err
	}

	cfg.MSMatrixTimeout, err = envvar.GetDuration( "MATRIX_STORE_MATRIX_TIMEOUT", 5 * time.Second)
	if err != nil {
		return nil, err
	}

	return cfg, nil
}