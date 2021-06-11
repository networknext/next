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
	relayBackendLiveData = "RelayBackendLiveData"
)

type RedisMatrixStore struct {
	pool          *redis.Pool
	matrixTimeout time.Duration
}

func NewRedisMatrixStore(hostname string, password string, maxIdleConnections int, maxActiveConnections int, readTimeout time.Duration, writeTimeout time.Duration, matrixExpire time.Duration) (*RedisMatrixStore, error) {
	// Get a standard redis pool
	pool := NewRedisPool(hostname, password, maxIdleConnections, maxActiveConnections)
	// Update the dial func with the read and write timeout
	pool.Dial = func() (redis.Conn, error) {
		return redis.Dial("tcp", hostname,
			redis.DialPassword(password),
			redis.DialReadTimeout(readTimeout),
			redis.DialWriteTimeout(writeTimeout))
	}

	r := &RedisMatrixStore{pool: pool, matrixTimeout: matrixExpire}
	r.cleanupHook()

	// Ensure the pool is valid
	if err := ValidateRedisPool(pool); err != nil {
		return nil, fmt.Errorf("could not validate redis pool: %v", err)
	}

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

func (r *RedisMatrixStore) SetRelayBackendLiveData(data RelayBackendLiveData) error {
	bin, err := RelayBackendLiveDataToJSON(data)
	if err != nil {
		return err
	}

	conn := r.pool.Get()
	defer conn.Close()

	key := fmt.Sprintf("%s-%s", relayBackendLiveData, data.Address)
	reply, err := conn.Do("SET", key, bin, "PX", r.matrixTimeout.Milliseconds())
	if reply != "OK" {
		return fmt.Errorf("SetRelayBackendLiveData(): reply is not OK, instead got %s: %v", reply, err)
	}
	return err
}

func (r *RedisMatrixStore) GetRelayBackendLiveData() ([]RelayBackendLiveData, error) {
	conn := r.pool.Get()
	defer conn.Close()

	keys, err := redis.Strings(conn.Do("KEYS", relayBackendLiveData+"*"))
	if err == redis.ErrNil || err != nil {
		return []RelayBackendLiveData{}, fmt.Errorf("issue with db: %v", err)
	}

	if len(keys) == 0 {
		return []RelayBackendLiveData{}, fmt.Errorf("keys not found")
	}

	rbArr := make([]RelayBackendLiveData, len(keys))
	for i, key := range keys {
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

		rbArr[i] = relayData
	}

	return rbArr, nil
}
