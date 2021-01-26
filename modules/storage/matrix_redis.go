package storage

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gomodule/redigo/redis"
)

const (
	hSet  = "HSET"
	hGet  = "HGET"
	hVals = "HVALS"
	hDel  = "HDEL"
	set   = "SET"
	setex = "SETEX"
	get   = "GET"
	mGet  = "MGET"

	optimizer            = "OptimizerHash"
	masters              = "MatrixMastersHash"
	matrixSvc            = "MatrixServiceHash"
	svcMaster            = "ServiceMaster"
	matrixMaster         = "LiveMatrix"
	optimizerMaster      = "OptimizerMaster"
	relayBackendLiveData = "RelayBackendLiveData"
	RelayBackendMaster   = "RelayBackendMaster"
)

type RedisMatrixStore struct {
	pool          *redis.Pool
	matrixTimeout time.Duration
}

func NewRedisMatrixStore(addr string, readTimeout, writeTimeout, matrixExpire time.Duration) (*RedisMatrixStore, error) {
	r := new(RedisMatrixStore)
	pool := &redis.Pool{
		MaxIdle:     5,
		IdleTimeout: 60 * time.Second,
		Dial: func() (redis.Conn, error) {
			return redis.Dial("tcp", addr,
				redis.DialReadTimeout(readTimeout),
				redis.DialWriteTimeout(writeTimeout))
		},
	}
	r.pool = pool
	r.cleanupHook()
	r.matrixTimeout = matrixExpire

	return r, nil
}

func (r *RedisMatrixStore) cleanupHook() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	signal.Notify(c, syscall.SIGTERM)
	signal.Notify(c, syscall.SIGKILL)
	go func() {
		<-c
		r.pool.Close()
	}()
}

func (r *RedisMatrixStore) Close() error {
	return r.pool.Close()
}

func (r *RedisMatrixStore) GetLiveMatrix(matrixType string) ([]byte, error) {
	conn := r.pool.Get()
	data, err := redis.Bytes(conn.Do(get, matrixMaster+matrixType))
	if err == redis.ErrNil {
		return []byte{}, fmt.Errorf("matrix not found")
	}
	if err != nil {
		return []byte{}, err
	}

	return data, nil
}

func (r *RedisMatrixStore) UpdateLiveMatrix(matrixData []byte, matrixType string) error {
	conn := r.pool.Get()
	_, err := conn.Do("SET", matrixMaster+matrixType, matrixData, "PX", r.matrixTimeout.Milliseconds())
	return err
}

func (r *RedisMatrixStore) GetOptimizerMatrices() ([]Matrix, error) {
	conn := r.pool.Get()
	dataArr, err := redis.ByteSlices(conn.Do(hVals, optimizer))
	if err == redis.ErrNil || len(dataArr) == 0 {
		return []Matrix{}, fmt.Errorf("optimizer matrices not found")
	}
	if err != nil {
		return []Matrix{}, err
	}

	matrices := make([]Matrix, len(dataArr))
	for i, v := range dataArr {
		matrix, err := MatrixFromJSON(v)
		if err != nil {
			return []Matrix{}, err
		}
		matrices[i] = matrix
	}

	return matrices, nil
}

func (r *RedisMatrixStore) UpdateOptimizerMatrix(matrix Matrix) error {
	jsonMatrix, err := MatrixToJSON(matrix)
	if err != nil {
		return err
	}

	conn := r.pool.Get()
	_, err = conn.Do(hSet, optimizer, fmt.Sprint(matrix.OptimizerID)+matrix.Type, jsonMatrix)
	return err
}

func (r *RedisMatrixStore) DeleteOptimizerMatrix(id uint64, matrixType string) error {
	conn := r.pool.Get()
	_, err := conn.Do(hDel, optimizer, fmt.Sprint(id)+matrixType)
	return err
}

func (r *RedisMatrixStore) GetMatrixSvcs() ([]MatrixSvcData, error) {
	conn := r.pool.Get()
	dataArr, err := redis.ByteSlices(conn.Do(hVals, matrixSvc))
	if err == redis.ErrNil || len(dataArr) == 0 {
		return []MatrixSvcData{}, fmt.Errorf("matrix svc data not found")
	}
	if err != nil {
		return []MatrixSvcData{}, err
	}

	matrices := make([]MatrixSvcData, len(dataArr))
	for i, v := range dataArr {
		matrix, err := MatrixSvcFromJSON(v)
		if err != nil {
			return []MatrixSvcData{}, err
		}
		matrices[i] = matrix
	}

	return matrices, nil
}

func (r *RedisMatrixStore) UpdateMatrixSvc(matrixSvcData MatrixSvcData) error {
	jsonMatrixSvcData, err := MatrixSvcToJSON(matrixSvcData)
	if err != nil {
		return err
	}

	conn := r.pool.Get()
	_, err = conn.Do(hSet, matrixSvc, matrixSvcData.ID, jsonMatrixSvcData)
	return err
}

func (r *RedisMatrixStore) DeleteMatrixSvc(id uint64) error {
	conn := r.pool.Get()
	_, err := conn.Do(hDel, matrixSvc, id)
	return err
}

func (r *RedisMatrixStore) UpdateMatrixSvcMaster(id uint64) error {
	conn := r.pool.Get()
	_, err := conn.Do(hSet, masters, svcMaster, id)
	return err
}

func (r *RedisMatrixStore) UpdateOptimizerMaster(id uint64) error {
	conn := r.pool.Get()
	_, err := conn.Do(hSet, masters, optimizerMaster, id)
	return err
}

func (r *RedisMatrixStore) GetMatrixSvcMaster() (uint64, error) {
	conn := r.pool.Get()
	value, err := redis.Uint64(conn.Do(hGet, masters, svcMaster))
	if err == redis.ErrNil {
		return 0, fmt.Errorf("matrix svc master not found")
	}
	return value, err
}

func (r *RedisMatrixStore) GetOptimizerMaster() (uint64, error) {
	conn := r.pool.Get()
	value, err := redis.Uint64(conn.Do(hGet, masters, optimizerMaster))
	if err == redis.ErrNil {
		return value, fmt.Errorf("optimizer master not found")
	}

	return value, err
}

func (r *RedisMatrixStore) SetRelayBackendLiveData(data RelayBackendLiveData) error {
	bin, err := RelayBackendLiveDataToJSON(data)
	if err != nil {
		return err
	}

	conn := r.pool.Get()
	key := fmt.Sprintf("%s-%s", relayBackendLiveData, data.Address)
	_, err = conn.Do("SET", key, bin, "EX", r.matrixTimeout.Milliseconds())
	return err
}

func (r *RedisMatrixStore) GetRelayBackendLiveData(addresses []string) ([]RelayBackendLiveData, error) {
	conn := r.pool.Get()
	var rbArr []RelayBackendLiveData
	for _, addr := range addresses {
		key := fmt.Sprintf("%s-%s", relayBackendLiveData, addr)
		bin, err := redis.Bytes(conn.Do("GET", key))
		if err == redis.ErrNil {
			continue
		}
		if err != nil {
			return []RelayBackendLiveData{}, fmt.Errorf("issue with db: %s", err.Error())
		}

		relayData, err := RelayBackendLiveDataFromJson(bin)
		if err != nil {
			return []RelayBackendLiveData{}, fmt.Errorf("unable to unmarshal relay data: %s", err.Error())
		}

		rbArr = append(rbArr, relayData)
	}
	return rbArr, nil
}

func (r *RedisMatrixStore) SetRelayBackendMaster(data RelayBackendLiveData) error {
	conn := r.pool.Get()
	bin, err := RelayBackendLiveDataToJSON(data)
	if err != nil {
		return err
	}
	_, err = conn.Do(hSet, masters, RelayBackendMaster, bin)
	return err
}

func (r *RedisMatrixStore) GetRelayBackendMaster() (RelayBackendLiveData, error) {
	conn := r.pool.Get()
	bin, err := redis.Bytes(conn.Do(hGet, masters, RelayBackendMaster))
	if err == redis.ErrNil {
		return RelayBackendLiveData{}, fmt.Errorf("relay backend master not found")
	}
	if err != nil {
		return RelayBackendLiveData{}, fmt.Errorf("issue with db: %s", err.Error())
	}

	return RelayBackendLiveDataFromJson(bin)
}
